package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/template"
)

// Config holds the configuration for a new project.
type Config struct {
	// Name is made available to templates.
	Name string
	// Host is the project host, e.g. github.com. Available in templates.
	Host string
	// Owner is the project owner, e.g. SCM username. Available in templates.
	Owner string
	// ProjectDir is the directory where project files should be written.
	ProjectDir string
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
	// Skeleton provides the files and values for the new project.
	Skeleton *kickoff.Skeleton
	// Values are user defined values that are merged on top of values from the
	// project skeleton.
	Values template.Values
}

// OpType defines the type of operation that should be performed for a given
// project file, template or directory.
type OpType uint8

const (
	OpCreate OpType = iota
	OpSkipExisting
	OpSkipUser
	OpOverwrite
)

// Destination describes the destination a project file should be written to.
type Destination struct {
	// Base is the base dir of the project.
	Base string
	// Path is the path relative to the base dir.
	Path string
}

// RelPath returns the path relative to the project root.
func (d Destination) RelPath() string {
	return d.Path
}

// AbsPath returns the absolute file path of the destination.
func (d Destination) AbsPath() string {
	return filepath.Join(d.Base, d.Path)
}

// Exists returns true if the destination already exists.
func (d Destination) Exists() bool {
	_, err := os.Stat(d.AbsPath())
	return err == nil
}

// Operation defines an operation involving a source and destination file. This
// is produced by a plan to review it before actually performing it.
type Operation struct {
	Type   OpType
	Source *kickoff.BufferedFile
	Dest   *Destination
}

// Plan holds the operations to create a new project. A plan is created from a
// project configuration and can be inspected/printed before being applied.
type Plan struct {
	OpCounts   map[OpType]int
	Operations []*Operation

	values        template.Values
	dirRewriteMap map[string]string
	skipMap       map[string]bool
	overwriteMap  map[string]bool
}

// MakePlan creates a creation plan for the given config.
func MakePlan(config *Config) (*Plan, error) {
	p := &Plan{
		OpCounts:      make(map[OpType]int),
		dirRewriteMap: make(map[string]string),
		skipMap:       make(map[string]bool),
		overwriteMap:  make(map[string]bool),
	}

	err := addCleanRelPathsToFileMap(p.skipMap, config.SkipFiles)
	if err != nil {
		return nil, err
	}

	err = addCleanRelPathsToFileMap(p.overwriteMap, config.OverwriteFiles)
	if err != nil {
		return nil, err
	}

	err = p.makeTemplateValues(config, config.Skeleton)
	if err != nil {
		return nil, err
	}

	err = p.makeOperations(config)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// SkipsExisting returns true if the plan skips some existing files.
func (p *Plan) SkipsExisting() bool {
	return p.OpCounts[OpSkipExisting] > 0
}

// IsNoOp returns true if the plan does not contain any operations or if there
// are solely skip operations.
func (p *Plan) IsNoOp() bool {
	return p.OpCounts[OpCreate] == 0 && p.OpCounts[OpOverwrite] == 0
}

// Apply applies the plan. It will write all necessary project files to the
// target directory.
func (p *Plan) Apply() error {
	for _, op := range p.Operations {
		if err := p.executeOperation(op); err != nil {
			return err
		}
	}

	return nil
}

// Create is a convenience wrapper to make a plan and immediately apply it.
func Create(config *Config) error {
	plan, err := MakePlan(config)
	if err != nil {
		return err
	}

	return plan.Apply()
}

func (p *Plan) executeOperation(op *Operation) error {
	if op.Type == OpSkipUser || op.Type == OpSkipExisting {
		return nil
	}

	source := op.Source
	dest := op.Dest

	if source.Mode.IsDir() {
		return os.MkdirAll(dest.AbsPath(), source.Mode)
	}

	if err := os.MkdirAll(filepath.Dir(dest.AbsPath()), 0755); err != nil {
		return err
	}

	content := source.Content

	if filepath.Ext(source.RelPath) == kickoff.SkeletonTemplateExtension {
		rendered, err := template.Render(string(content), p.values)
		if err != nil {
			return err
		}

		content = []byte(rendered)
	}

	return ioutil.WriteFile(dest.AbsPath(), content, source.Mode)
}

func (p *Plan) makeTemplateValues(config *Config, skeleton *kickoff.Skeleton) error {
	values, err := template.MergeValues(skeleton.Values, config.Values)
	if err != nil {
		return err
	}

	var (
		licenseName    string
		gitignoreQuery string
	)

	if config.License != nil {
		licenseName = config.License.Name
	}

	if config.Gitignore != nil {
		gitignoreQuery = config.Gitignore.Query
	}

	p.values = template.Values{
		"Project": map[string]string{
			"Name":          config.Name,
			"Host":          config.Host,
			"Owner":         config.Owner,
			"License":       licenseName,
			"Gitignore":     gitignoreQuery,
			"URL":           fmt.Sprintf("https://%s/%s/%s", config.Host, config.Owner, config.Name),
			"GoPackagePath": fmt.Sprintf("%s/%s/%s", config.Host, config.Owner, config.Name),
		},
		"Values":  values,
		"License": config.License,
	}

	return nil
}

func makeSources(config *Config) []*kickoff.BufferedFile {
	var extraFiles []*kickoff.BufferedFile

	if config.License != nil {
		text := license.ResolvePlaceholders(config.License.Body, license.FieldMap{
			"project": config.Name,
			"author":  config.Owner,
			"year":    strconv.Itoa(time.Now().Year()),
		})

		extraFiles = append(extraFiles, &kickoff.BufferedFile{
			RelPath: "LICENSE",
			Content: []byte(text),
			Mode:    0644,
		})
	}

	if config.Gitignore != nil {
		extraFiles = append(extraFiles, &kickoff.BufferedFile{
			RelPath: ".gitignore",
			Content: config.Gitignore.Content,
			Mode:    0644,
		})
	}

	sources := kickoff.MergeFiles(config.Skeleton.Files, extraFiles)

	// We sort files by path so we can ensure that parent directories get
	// created before we attempt to write the files contained in them. This
	// works because we are going to process files sequentially. In the future
	// we might take a different approach and do things concurrently.
	sort.SliceStable(sources, func(i, j int) bool {
		return sources[i].RelPath < sources[j].RelPath
	})

	return sources
}

func (p *Plan) makeOperations(config *Config) error {
	p.Operations = make([]*Operation, 0)

	sources := makeSources(config)

	for _, source := range sources {
		dest, err := p.makeDestination(config.ProjectDir, source)
		if err != nil {
			return err
		}

		var opType OpType
		switch {
		case dest.Exists() && !config.Overwrite && !matchPathPrefix(p.overwriteMap, dest.RelPath()):
			opType = OpSkipExisting
		case matchPathPrefix(p.skipMap, dest.RelPath()):
			opType = OpSkipUser
		case dest.Exists():
			opType = OpOverwrite
		}

		p.Operations = append(p.Operations, &Operation{
			Type:   opType,
			Source: source,
			Dest:   dest,
		})
		p.OpCounts[opType]++
	}

	return nil
}

func (p *Plan) makeDestination(targetDir string, f *kickoff.BufferedFile) (*Destination, error) {
	relPath := f.RelPath
	srcFilename := filepath.Base(relPath)
	srcRelDir := filepath.Dir(relPath)

	targetFilename, err := template.Render(srcFilename, p.values)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve templated filename %q: %w", srcFilename, err)
	}

	if len(targetFilename) == 0 {
		return nil, fmt.Errorf("templated filename %q resolved to an empty string", srcFilename)
	}

	targetRelDir := p.resolveTargetDir(srcRelDir)

	targetRelPath := filepath.Join(targetRelDir, targetFilename)

	// Sanity check to guard against malicious injection of directory
	// traveral (e.g. "../../" in the template string).
	if filepath.Dir(targetRelPath) != targetRelDir {
		return nil, fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, targetFilename)
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

	return &Destination{Base: targetDir, Path: targetRelPath}, nil
}

func (p *Plan) resolveTargetDir(dir string) string {
	// If the src dir's name was templated, lookup the resolved name and
	// use that as destination.
	if rewrittenDir, ok := p.dirRewriteMap[dir]; ok {
		return rewrittenDir
	}

	return dir
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
