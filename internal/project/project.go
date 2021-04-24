package project

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
)

// Config holds the configuration for a new project that can be created from a
// *kickoff.Skeleton.
type Config struct {
	// ProjectName is made available to templates. If empty the basename of the
	// target directory is used.
	ProjectName string
	// Host is the project host, e.g. github.com. Available in templates.
	Host string
	// Owner is the project owner, e.g. SCM username. Available in templates.
	Owner string
	// Gitignore template to use for creating .gitignore. If nil, no .gitignore
	// is created.
	Gitignore *gitignore.Template
	// License info for the open source license to use. If nil, no LICENSE file
	// is created.
	License *license.Info
	// If Overwrite is true, existing file in the target directory that matches
	// the name of one of the skeleton files is overwritten.
	Overwrite bool
	// OverwriteFiles can be used to selectively overwrite existing files. File
	// paths must be relative to the target directory.
	OverwriteFiles []string
	// SkipFiles can be used to selectively skip creation of files or
	// directories. File paths must be relative to the target directory.
	SkipFiles []string
	// Filesystem to use for creating the project. Can be set to
	// *afero.MemMapFs in tests or for dry running project creation. If nil
	// an *afero.OsFs is used.
	Filesystem afero.Fs
	// Values are user defined values that are merged on top of values from the
	// project skeleton.
	Values template.Values
	// Output configures the io.Writer where the project creation summary is
	// written to. If nil output is discarded.
	Output io.Writer
}

// Project is the type responsible for project creation.
type Project struct {
	targetDir string
	name      string
	host      string
	owner     string

	dirRewriteMap map[string]string
	skipMap       map[string]bool
	overwriteMap  map[string]bool
	overwrite     bool

	fs        afero.Fs
	output    io.Writer
	license   *license.Info
	gitignore *gitignore.Template
	values    template.Values

	result *Result
}

// New creates a new *Project with given targetDir and config.
func New(targetDir string, config *Config) (*Project, error) {
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, err
	}

	p := &Project{
		targetDir:     targetDir,
		name:          config.ProjectName,
		owner:         config.Owner,
		host:          config.Host,
		fs:            config.Filesystem,
		output:        config.Output,
		dirRewriteMap: make(map[string]string),
		skipMap:       make(map[string]bool),
		overwriteMap:  make(map[string]bool),
		overwrite:     config.Overwrite,
		values:        config.Values,
		license:       config.License,
		gitignore:     config.Gitignore,
		result:        new(Result),
	}

	err = addCleanRelPathsToFileMap(p.skipMap, config.SkipFiles)
	if err != nil {
		return nil, err
	}

	err = addCleanRelPathsToFileMap(p.overwriteMap, config.OverwriteFiles)
	if err != nil {
		return nil, err
	}

	p.applyDefaults()

	return p, nil
}

func (p *Project) applyDefaults() {
	if p.name == "" {
		p.name = filepath.Base(p.targetDir)
	}

	if p.fs == nil {
		p.fs = afero.NewOsFs()
	}

	if p.output == nil {
		p.output = ioutil.Discard
	}
}

// Create creates a project in targetDir from given skeleton with the provided
// config. The returned result contains information about all actions that were
// performed.
func Create(s *kickoff.Skeleton, targetDir string, config *Config) (*Result, error) {
	p, err := New(targetDir, config)
	if err != nil {
		return nil, err
	}

	return p.Create(s)
}

// Create creates the project from given skeleton. The returned result contains
// information about all actions that were performed.
func (p *Project) Create(s *kickoff.Skeleton) (*Result, error) {
	values, err := p.makeTemplateValues(s)
	if err != nil {
		return nil, err
	}

	err = p.create(s, values)
	if err != nil {
		return nil, err
	}

	p.writeSummary(p.output)

	return p.result, nil
}

func (p *Project) makeTemplateValues(skeleton *kickoff.Skeleton) (template.Values, error) {
	values, err := template.MergeValues(skeleton.Values, p.values)
	if err != nil {
		return nil, err
	}

	var (
		licenseName    string
		gitignoreQuery string
	)

	if p.license != nil {
		licenseName = p.license.Name
	}

	if p.gitignore != nil {
		gitignoreQuery = p.gitignore.Query
	}

	vals := template.Values{
		"Project": map[string]string{
			"Name":          p.name,
			"Host":          p.host,
			"Owner":         p.owner,
			"License":       licenseName,
			"Gitignore":     gitignoreQuery,
			"URL":           fmt.Sprintf("https://%s/%s/%s", p.host, p.owner, p.name),
			"GoPackagePath": fmt.Sprintf("%s/%s/%s", p.host, p.owner, p.name),
		},
		"Values":  values,
		"License": p.license,
	}

	return vals, nil
}

func (p *Project) makeSources(skeleton *kickoff.Skeleton) []*kickoff.BufferedFile {
	var extraFiles []*kickoff.BufferedFile

	if p.license != nil {
		text := license.ResolvePlaceholders(p.license.Body, license.FieldMap{
			"project": p.name,
			"author":  p.owner,
			"year":    strconv.Itoa(time.Now().Year()),
		})

		extraFiles = append(extraFiles, &kickoff.BufferedFile{
			RelPath: "LICENSE",
			Content: []byte(text),
			Mode:    0644,
		})
	}

	if p.gitignore != nil {
		extraFiles = append(extraFiles, &kickoff.BufferedFile{
			RelPath: ".gitignore",
			Content: p.gitignore.Content,
			Mode:    0644,
		})
	}

	sources := kickoff.MergeFiles(skeleton.Files, extraFiles)

	// We sort files by path so we can ensure that parent directories get
	// created before we attempt to write the files contained in them. This
	// works because we are going to process files sequentially. In the future
	// we might take a different approach and do things concurrently.
	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].RelPath < sources[j].RelPath
	})

	return sources
}

func (p *Project) create(s *kickoff.Skeleton, values template.Values) error {
	sources := p.makeSources(s)

	for _, source := range sources {
		dest, err := p.makeDestination(source, values)
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

func (p *Project) makeDestination(f *kickoff.BufferedFile, values template.Values) (Destination, error) {
	relPath := f.RelPath
	srcFilename := filepath.Base(relPath)
	srcRelDir := filepath.Dir(relPath)

	targetFilename, err := template.Render(srcFilename, values)
	if err != nil {
		return Destination{}, fmt.Errorf("failed to resolve templated filename %q: %w", srcFilename, err)
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
	if ext := filepath.Ext(targetRelPath); ext == kickoff.SkeletonTemplateExtension {
		targetRelPath = targetRelPath[:len(targetRelPath)-len(ext)]
	}

	// Track potentially templated directory names that need to be
	// rewritten for all files contained in them.
	if f.Mode.IsDir() && relPath != targetRelPath {
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

func (p *Project) trackAction(action Action) {
	if p.result.Stats == nil {
		p.result.Stats = make(Stats)
	}

	p.result.Actions = append(p.result.Actions, action)
	p.result.Stats[action.Type]++
}

func (p *Project) executeAction(action Action, values template.Values) error {
	p.trackAction(action)

	source := action.Source
	dest := action.Destination

	switch action.Type {
	case ActionTypeSkipUser, ActionTypeSkipExisting:
		return nil
	default:
		if source.Mode.IsDir() {
			return p.fs.MkdirAll(dest.AbsPath(), source.Mode)
		}

		if err := p.fs.MkdirAll(filepath.Dir(dest.AbsPath()), 0755); err != nil {
			return err
		}

		content := source.Content
		if filepath.Ext(source.RelPath) == kickoff.SkeletonTemplateExtension {
			rendered, err := template.Render(string(content), values)
			if err != nil {
				return err
			}

			content = []byte(rendered)
		}

		return afero.WriteFile(p.fs, dest.AbsPath(), content, source.Mode)
	}
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

func addCleanRelPathsToFileMap(fileMap map[string]bool, paths []string) error {
	for _, path := range paths {
		if filepath.IsAbs(path) {
			return fmt.Errorf("found illegal absolute path: %s", path)
		}

		relPath := filepath.Clean(path)

		fileMap[relPath] = true
	}

	return nil
}
