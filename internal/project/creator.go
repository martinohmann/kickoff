package project

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	git "github.com/go-git/go-git/v5"
	"github.com/martinohmann/kickoff/internal/config"
	"github.com/martinohmann/kickoff/internal/file"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/spf13/afero"
)

// CreateOptions provide optional configuration for the project creator. If
// DryRun is set to true, actions will only be logged, but nothing will be
// written to the project output dir.
type CreateOptions struct {
	DryRun     bool
	Config     config.Project
	Values     template.Values
	Gitignore  string
	InitGit    bool
	Overwrite  bool
	AllowEmpty bool
	License    *license.Info
}

// CreateProject creates a new project with given options from skeleton s in
// targetDir.
func Create(s *skeleton.Skeleton, targetDir string, options *CreateOptions) error {
	creator := NewCreator(options)

	return creator.CreateProject(s, targetDir)
}

// Creator can create new projects from project skeletons.
type Creator struct {
	Filesystem afero.Fs
	Options    *CreateOptions
}

// NewCreator creates a new project creator with given options.
func NewCreator(options *CreateOptions) *Creator {
	if options == nil {
		options = &CreateOptions{}
	}

	return &Creator{
		Filesystem: createFilesystem(options),
		Options:    options,
	}
}

// CreateProject creates a new project from skeleton s in targetDir.
func (c *Creator) CreateProject(s *skeleton.Skeleton, targetDir string) error {
	values, err := c.getTemplateValues(s.Values)
	if err != nil {
		return err
	}

	fw := NewSkeletonFileWriter(c.Filesystem, c.Options.Overwrite, c.Options.AllowEmpty)

	log.Infof("creating project in %s", targetDir)

	err = fw.WriteFiles(s.Files, targetDir, values)
	if err != nil {
		return err
	}

	if c.Options.License != nil {
		body := strings.ReplaceAll(c.Options.License.Body, "[fullname]", c.Options.Config.Owner)
		body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

		err = c.writeFile(targetDir, "LICENSE", []byte(body), 0644)
		if err != nil {
			return err
		}
	}

	if len(c.Options.Gitignore) > 0 {
		err = c.writeFile(targetDir, ".gitignore", []byte(c.Options.Gitignore), 0644)
		if err != nil {
			return err
		}
	}

	if c.Options.InitGit && !c.Options.DryRun {
		err = initGitRepository(targetDir)
		if err != nil {
			return err
		}
	}

	log.Info("project creation complete")

	return nil
}

func (c *Creator) getTemplateValues(values template.Values) (template.Values, error) {
	values, err := template.MergeValues(values, c.Options.Values)
	if err != nil {
		return nil, err
	}

	return template.Values{
		"Project": &c.Options.Config,
		"Values":  values,
		"License": c.Options.License,
	}, nil
}

func (c *Creator) writeFile(targetDir, path string, content []byte, mode os.FileMode) error {
	targetPath := filepath.Join(targetDir, path)

	if file.Exists(targetPath) && !c.Options.Overwrite {
		log.WithField("path", path).Warn("target exists, skipping")
		return nil
	}

	log.WithField("path", path).Info("writing file")

	return afero.WriteFile(c.Filesystem, targetPath, content, mode)
}

func initGitRepository(path string) error {
	log.Info("initializing git repository")

	_, err := git.PlainInit(path, false)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return err
	}

	return nil
}

func createFilesystem(options *CreateOptions) afero.Fs {
	if options.DryRun {
		return afero.NewMemMapFs()
	}

	return afero.NewOsFs()
}
