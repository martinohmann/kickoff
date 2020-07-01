// Package project contains a builder for creating new projects from various
// options.
package project

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
)

// Builder provides a fluent interface for setting options for project creation
// before building it.
type Builder struct {
	config config.Project

	overwriteAll bool
	overwriteMap map[string]bool
	skipMap      map[string]bool
	dirMap       map[string]string

	fs      afero.Fs
	files   []File
	values  template.Values
	license *license.Info

	templateRenderer *template.Renderer

	err   error
	stats Stats
}

// NewBuilder creates a new *Builder for the given project configuration.
func NewBuilder(config config.Project) *Builder {
	return &Builder{
		config:       config,
		dirMap:       make(map[string]string),
		overwriteMap: make(map[string]bool),
		skipMap:      make(map[string]bool),
		values:       template.Values{},
	}
}

// WithFilesystem sets the filesystem the builder should use. For dry run mode
// this can be an afero.MemMapFs. If not set afero.OsFs will be used.
func (b *Builder) WithFilesystem(fs afero.Fs) *Builder {
	b.fs = fs
	return b
}

// AddFile adds a File that should be created in the project.
func (b *Builder) AddFile(file File) *Builder {
	b.files = append(b.files, file)
	return b
}

// OverwriteAll allows overwriting any existing file in the target directory
// that has the same path as one of the files added to the builder. The default
// is to not overwrite existing files.
func (b *Builder) OverwriteAll(overwrite bool) *Builder {
	b.overwriteAll = overwrite
	return b
}

// OverwriteFiles adds a list of targetPaths to the builder that are allowed to
// be overwritting if the already exist. This allows for selectively
// overwriting only specific files, keeping others unchanged. Paths are assumed
// to be relative to the new project dir. Passing absolute paths will cause
// successive calls to Build to return an error.
func (b *Builder) OverwriteFiles(paths []string) *Builder {
	if b.err != nil {
		return b
	}

	for _, path := range paths {
		if filepath.IsAbs(path) {
			b.err = fmt.Errorf("found absolute path in overwrites: %s", path)
			return b
		}

		relPath := filepath.Clean(path)

		b.overwriteMap[relPath] = true
	}
	return b
}

// @TODO add doc
func (b *Builder) SkipFiles(paths []string) *Builder {
	if b.err != nil {
		return b
	}

	for _, path := range paths {
		if filepath.IsAbs(path) {
			b.err = fmt.Errorf("found absolute path in files to skip: %s", path)
			return b
		}

		relPath := filepath.Clean(path)

		b.skipMap[relPath] = true
	}
	return b
}

// AddValues adds values that should be injected into templates. Values will be
// merged on to of already added values. Within templates these values are
// accessible via the `.Values` map.
func (b *Builder) AddValues(values template.Values) *Builder {
	if b.err == nil {
		b.values, b.err = template.MergeValues(b.values, values)
	}
	return b
}

// WithGitignore adds a .gitignore file with given content to the project. The
// default is to not include a .gitignore file.
func (b *Builder) WithGitignore(content string) *Builder {
	return b.AddFile(&fileInfo{
		relPath: ".gitignore",
		content: []byte(content),
		mode:    0644,
	})
}

// WithLicense adds a LICENSE file to the project. The license info is made
// available to templates via the `.License` value. In license texts,
// [fullname] placeholders are replaced with the owner value from the project
// config and [year] placeholders are replaced with the current year. The
// default is to not include a LICENSE file.
func (b *Builder) WithLicense(info *license.Info) *Builder {
	b.license = info

	text := license.ResolvePlaceholders(info.Body, license.FieldMap{
		"project": b.config.Name,
		"author":  b.config.Owner,
		"year":    strconv.Itoa(time.Now().Year()),
	})

	return b.AddFile(&fileInfo{
		relPath: "LICENSE",
		content: []byte(text),
		mode:    0644,
	})
}

// Build builds the project at targetDir using the options provided to the
// builder. Returns stats about created and skipped files.
func (b *Builder) Build(targetDir string) (Stats, error) {
	if b.err != nil {
		return Stats{}, b.err
	}

	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return Stats{}, err
	}

	if b.fs == nil {
		b.fs = afero.NewOsFs()
	}

	err = b.fs.MkdirAll(targetDir, 0755)
	if err != nil {
		return Stats{}, err
	}

	b.templateRenderer = template.NewRenderer(template.Values{
		"Project": &b.config,
		"Values":  b.values,
		"License": b.license,
	})

	// We sort files by path so we can ensure that parent directories get
	// created before we attempt to write the files contained in them. This
	// works because we are going to process files sequentially. In the future
	// we might take a different approach and do things concurrently.
	sort.SliceStable(b.files, func(i, j int) bool {
		return b.files[i].Path() < b.files[j].Path()
	})

	for _, f := range b.files {
		err := b.processFile(f, targetDir)
		if err != nil {
			return b.stats, err
		}
	}

	return b.stats, nil
}

// shouldOverwrite returns true if path is allowed to be overwritten.
func (b *Builder) shouldOverwrite(path string) bool {
	return b.overwriteAll || b.overwriteMap[path]
}

func (b *Builder) shouldSkip(path string) bool {
	for {
		if b.skipMap[path] {
			return true
		}

		if path = filepath.Dir(path); path == "." || path == "/" {
			break
		}
	}

	return false
}

// processFile processes f and writes the result to the targetDir.
func (b *Builder) processFile(f File, targetDir string) error {
	targetRelPath, err := b.buildTargetRelPath(f)
	if err != nil {
		return err
	}

	targetAbsPath := filepath.Join(targetDir, targetRelPath)
	relPath := f.Path()

	// Track potentially templated directory names that need to be
	// rewritten for all files contained in them.
	if f.Mode().IsDir() && relPath != targetRelPath {
		b.dirMap[relPath] = targetRelPath
	}

	logger := log.WithField("path", relPath)

	if relPath != targetRelPath {
		logger = log.WithFields(log.Fields{
			"path.src":    relPath,
			"path.target": targetRelPath,
		})
	}

	if b.shouldSkip(targetRelPath) {
		logger.Warnf("skipping exluded %s", fileType(f))
		b.stats.Skipped++
		return nil
	}

	if !b.shouldOverwrite(targetRelPath) && file.Exists(targetAbsPath) {
		logger.Warnf("skipping existing %s", fileType(f))
		b.stats.Skipped++
		return nil
	}

	return b.writeFile(logger, f, targetAbsPath)
}

// writeFile writes f to targetAbsPath.
func (b *Builder) writeFile(logger *log.Entry, f File, targetAbsPath string) (err error) {
	switch {
	case f.Mode().IsDir():
		logger.Info("creating directory")

		err = b.fs.MkdirAll(targetAbsPath, f.Mode())
	case filepath.Ext(f.Path()) == ".skel":
		logger.Info("rendering template")

		var rendered string

		rendered, err = b.renderTemplateFile(f)
		if err != nil {
			return err
		}

		err = afero.WriteFile(b.fs, targetAbsPath, []byte(rendered), f.Mode())
	default:
		logger.Info("copying file")

		err = b.copyFile(f, targetAbsPath)
	}

	if err != nil {
		return err
	}

	b.stats.increment(f)

	return nil
}

// renderTemplateFile renders the templated content of f and returns it.
func (b *Builder) renderTemplateFile(f File) (string, error) {
	r, err := f.Reader()
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return b.templateRenderer.Render(string(buf))
}

// copyFile copies f to dst and sets the correct file mode.
func (b *Builder) copyFile(f File, dst string) error {
	srcReader, err := f.Reader()
	if err != nil {
		return err
	}

	err = afero.WriteReader(b.fs, dst, srcReader)
	if err != nil {
		return err
	}

	return b.fs.Chmod(dst, f.Mode())
}

// buildTargetRelPath builds the target path of f relative to the project
// directory and returns it. Resolves templated filenames and returns potential
// errors.
func (b *Builder) buildTargetRelPath(f File) (string, error) {
	relPath := f.Path()
	srcFilename := filepath.Base(relPath)
	srcRelDir := filepath.Dir(relPath)

	targetFilename, err := b.templateRenderer.Render(srcFilename)
	if err != nil {
		return "", fmt.Errorf("failed to resolve templated filename %q: %v", srcFilename, err)
	}

	if len(targetFilename) == 0 {
		return "", fmt.Errorf("templated filename %q resolved to an empty string", srcFilename)
	}

	targetRelDir := srcRelDir

	// If the src dir's name was templated, lookup the resolved name and
	// use that as destination.
	if dir, ok := b.dirMap[srcRelDir]; ok {
		targetRelDir = dir
	}

	targetRelPath := filepath.Join(targetRelDir, targetFilename)

	// Sanity check to guard against malicious injection of directory
	// traveral (e.g. "../../" in the template string).
	if filepath.Dir(targetRelPath) != targetRelDir {
		return "", fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, targetFilename)
	}

	// Trim .skel extension.
	if ext := filepath.Ext(targetRelPath); ext == ".skel" {
		targetRelPath = targetRelPath[:len(targetRelPath)-len(ext)]
	}

	return targetRelPath, nil
}
