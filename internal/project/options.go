package project

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
)

// Option is a func that configures a *Project.
type Option func(*Project) error

// Options is a collection of project configuration options.
type Options []Option

// Add adds options to o and returns a copy if it.
func (o *Options) Add(options ...Option) Options {
	*o = append(*o, options...)
	return *o
}

// WithOverwrite configures the project to overwrite files present in the target
// directory. Without this option existing files will not be altered.
func WithOverwrite(p *Project) error {
	p.overwrite = true
	return nil
}

// WithOverwriteFiles allows for selectively overwriting only a subset of
// existing file paths. The provided paths must be relative to the project
// root.
func WithOverwriteFiles(paths ...string) Option {
	return func(p *Project) error {
		return addCleanRelPathsToFileMap(p.overwriteMap, paths)
	}
}

// WithSkipFiles allows for selectively excluding files from the new project.
// The provided paths must be relative to the project root.
func WithSkipFiles(paths ...string) Option {
	return func(p *Project) error {
		return addCleanRelPathsToFileMap(p.skipMap, paths)
	}
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

// WithFilesystem sets the filesystem the project is created on. For example,
// in tests or during dry run this can be used to perform all operations
// against an in-memory filesystem instead.
func WithFilesystem(fs afero.Fs) Option {
	return func(p *Project) error {
		p.fs = fs
		return nil
	}
}

// WithExtraFile adds an extra file to the project that is not included in the
// skeleton the project is created from. The provided io.Reader is used to read
// the file contents. Path must be relative to the new project root.
func WithExtraFile(r io.Reader, path string, mode os.FileMode) Option {
	return func(p *Project) error {
		p.extraFiles = append(p.extraFiles, NewSource(r, path, mode))
		return nil
	}
}

// WithExtraValues adds additional template values to the project.
func WithExtraValues(values template.Values) Option {
	return func(p *Project) error {
		vals, err := template.MergeValues(p.extraValues, values)
		if err != nil {
			return err
		}

		p.extraValues = vals
		return nil
	}
}

// WithGitignore adds a .gitignore file to the project with given text.
func WithGitignore(text string) Option {
	return func(p *Project) error {
		return WithExtraFile(bytes.NewBufferString(text), ".gitignore", 0644)(p)
	}
}

// WithLicense adds a LICENSE file to the project which is populated from given
// info. Placeholders for project name and owner are replaced automatically
// replaced with values from the project config prior to writing the file. The
// license info is also made available to templates.
func WithLicense(info *license.Info) Option {
	return func(p *Project) error {
		p.license = info

		text := license.ResolvePlaceholders(info.Body, license.FieldMap{
			"project": p.config.Name,
			"author":  p.config.Owner,
			"year":    strconv.Itoa(time.Now().Year()),
		})

		return WithExtraFile(bytes.NewBufferString(text), "LICENSE", 0644)(p)
	}
}

// WithLogger configures the logger that is used to output actions performed
// while creating the project.
func WithLogger(logger Logger) Option {
	return func(p *Project) error {
		p.logger = logger
		return nil
	}
}
