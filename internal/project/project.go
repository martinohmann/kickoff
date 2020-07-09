package project

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
)

// Project holds the configuration for a new project that can be created from a
// *skeleton.Skeleton.
type Project struct {
	config    config.Project
	targetDir string

	dirRewriteMap map[string]string
	skipMap       map[string]bool
	overwriteMap  map[string]bool
	overwrite     bool

	fs          afero.Fs
	logger      Logger
	extraFiles  []Source
	extraValues template.Values
	license     *license.Info
}

// New creates a new *Project with given config and targetDir.
func New(config config.Project, targetDir string, options ...Option) (*Project, error) {
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, err
	}

	p := &Project{
		config:        config,
		targetDir:     targetDir,
		dirRewriteMap: make(map[string]string),
		overwriteMap:  make(map[string]bool),
		skipMap:       make(map[string]bool),
		extraValues:   template.Values{},
	}

	for _, option := range options {
		if err := option(p); err != nil {
			return nil, err
		}
	}

	if p.fs == nil {
		p.fs = afero.NewOsFs()
	}

	if p.logger == nil {
		p.logger = NewLogger(ioutil.Discard)
	}

	return p, nil
}

// CreateFromSkeleton creates the project from given skeleton.
func (p *Project) CreateFromSkeleton(skeleton *skeleton.Skeleton) error {
	values, err := p.buildTemplateValues(skeleton)
	if err != nil {
		return err
	}

	defer p.logger.Flush()

	sources := p.collectSources(skeleton)

	return p.processSources(sources, values)
}

func (p *Project) buildTemplateValues(skeleton *skeleton.Skeleton) (template.Values, error) {
	values, err := template.MergeValues(skeleton.Values, p.extraValues)
	if err != nil {
		return nil, err
	}

	vals := template.Values{
		"Project": &p.config,
		"Values":  values,
		"License": p.license,
	}

	return vals, nil
}

func (p *Project) collectSources(skeleton *skeleton.Skeleton) []Source {
	sources := make([]Source, 0, len(skeleton.Files)+len(p.extraFiles))

	for _, file := range skeleton.Files {
		sources = append(sources, file)
	}

	sources = append(sources, p.extraFiles...)

	// We sort files by path so we can ensure that parent directories get
	// created before we attempt to write the files contained in them. This
	// works because we are going to process files sequentially. In the future
	// we might take a different approach and do things concurrently.
	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].Path() < sources[j].Path()
	})

	return sources
}

func (p *Project) processSources(sources []Source, values template.Values) error {
	for _, source := range sources {
		dest, err := p.buildDestination(source, values)
		if err != nil {
			return err
		}

		var actionType ActionType
		switch {
		case matchPathPrefix(p.skipMap, dest.RelPath()):
			actionType = ActionTypeSkipUser
		case dest.Exists() && !p.overwrite && !matchPathPrefix(p.overwriteMap, dest.RelPath()):
			actionType = ActionTypeSkipExisting
		case dest.Exists():
			actionType = ActionTypeOverwrite
		}

		action := Action{
			Type:        actionType,
			Source:      source,
			Destination: dest,
		}

		err = p.executeAction(action, values)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) buildDestination(f Source, values template.Values) (Destination, error) {
	relPath := f.Path()
	srcFilename := filepath.Base(relPath)
	srcRelDir := filepath.Dir(relPath)

	targetFilename, err := template.Render(srcFilename, values)
	if err != nil {
		return Destination{}, fmt.Errorf("failed to resolve templated filename %q: %v", srcFilename, err)
	}

	if len(targetFilename) == 0 {
		return Destination{}, fmt.Errorf("templated filename %q resolved to an empty string", srcFilename)
	}

	targetRelDir := p.resolveTargetDir(srcRelDir)

	targetRelPath := filepath.Join(targetRelDir, targetFilename)

	// Sanity check to guard against malicious injection of directory
	// traveral (e.g. "../../" in the template string).
	if filepath.Dir(targetRelPath) != targetRelDir {
		return Destination{}, fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, targetFilename)
	}

	// Trim .skel extension.
	if ext := filepath.Ext(targetRelPath); ext == ".skel" {
		targetRelPath = targetRelPath[:len(targetRelPath)-len(ext)]
	}

	// Track potentially templated directory names that need to be
	// rewritten for all files contained in them.
	if f.Mode().IsDir() && relPath != targetRelPath {
		p.dirRewriteMap[relPath] = targetRelPath
	}

	dest := Destination{
		Base: p.targetDir,
		Path: targetRelPath,
	}

	return dest, nil
}

func (p *Project) resolveTargetDir(dir string) string {
	// If the src dir's name was templated, lookup the resolved name and
	// use that as destination.
	if rewrittenDir, ok := p.dirRewriteMap[dir]; ok {
		return rewrittenDir
	}

	return dir
}

func (p *Project) executeAction(action Action, values template.Values) error {
	p.logger.Log(action)

	source := action.Source
	dest := action.Destination

	switch action.Type {
	case ActionTypeSkipUser, ActionTypeSkipExisting:
		return nil
	default:
		switch {
		case source.Mode().IsDir():
			return p.fs.MkdirAll(dest.AbsPath(), source.Mode())
		default:
			r, err := p.sourceReader(source, values)
			if err != nil {
				return err
			}

			return p.writeReader(dest.AbsPath(), r, source.Mode())
		}
	}
}

func (p *Project) sourceReader(source Source, values template.Values) (io.Reader, error) {
	r, err := source.Reader()
	if err != nil {
		return nil, err
	}

	if source.IsTemplate() {
		rendered, err := template.RenderReader(r, values)
		if err != nil {
			return nil, err
		}

		r = bytes.NewBufferString(rendered)
	}

	return r, nil
}

func (p *Project) writeReader(path string, r io.Reader, mode os.FileMode) error {
	err := afero.WriteReader(p.fs, path, r)
	if err != nil {
		return err
	}

	return p.fs.Chmod(path, mode)
}

// matchPathPrefix returns true if path or any parent dir of path is set in
// pathMap.
// E.g. if path is `pkg/foo/bar` and `pkg` or `pkg/foo` (or `pkg/foo/bar`) is
// present in pathMap, this returns true, otherwise false.
func matchPathPrefix(pathMap map[string]bool, path string) bool {
	for {
		if pathMap[path] {
			return true
		}

		if path = filepath.Dir(path); path == "." || path == "/" {
			break
		}
	}

	return false
}
