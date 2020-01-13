// Package project contains the core logic to create a project from a skeleton.
package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/gitignore"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	git "gopkg.in/src-d/go-git.v4"
)

// CreateOptions provide optional configuration for the project creator. If
// DryRun is set to true, actions will only be logged, but nothing will be
// written to the project output dir.
type CreateOptions struct {
	DryRun bool
	Config config.Config
}

// Create creates a new project in outputDir using the provided skeleton.
// Options provide additional configuration for the project creation behaviour.
// Returns an error if project creation fails.
func Create(skeleton *skeleton.Info, outputDir string, options *CreateOptions) error {
	if options == nil {
		options = &CreateOptions{}
	}

	c := &creator{
		dryRun: options.DryRun,
		config: options.Config,
		stats:  &createStats{},
	}

	var err error

	if options.Config.HasLicense() {
		c.license, err = license.Get(options.Config.License)
		if err == license.ErrNotFound {
			return fmt.Errorf("license %q not found, run `kickoff licenses list` to get a list of available licenses", options.Config.License)
		} else if err != nil {
			return err
		}
	}

	if options.Config.HasGitignore() {
		c.gitignore, err = gitignore.Get(options.Config.Gitignore)
		if err == gitignore.ErrNotFound {
			return fmt.Errorf("gitignore template %q not found, run `kickoff gitignore list` to get a list of available templates", options.Config.Gitignore)
		} else if err != nil {
			return err
		}
	}

	return c.create(skeleton, outputDir)
}

type createStats struct {
	dirsCreated       int
	filesCopied       int
	templatesRendered int
}

type creator struct {
	dryRun    bool
	gitignore string
	license   *license.Info
	config    config.Config
	stats     *createStats
}

func (c *creator) create(skeleton *skeleton.Info, outputDir string) error {
	if c.dryRun {
		log.Warn("dry run: no changes will be made")
	}

	config, err := skeleton.Config()
	if err != nil {
		return err
	}

	log.WithField("values", fmt.Sprintf("%#v", config.Values)).Debug("skeleton values")

	err = config.Values.Merge(c.config.Values)
	if err != nil {
		return err
	}

	log.WithField("values", fmt.Sprintf("%#v", config.Values)).Debug("merged values")

	log.WithFields(log.Fields{
		"skeleton": skeleton.Path,
		"target":   outputDir,
	}).Info("creating project from skeleton")

	err = c.processFiles(skeleton, outputDir, config.Values)
	if err != nil {
		return err
	}

	err = c.writeLicenseFile(outputDir)
	if err != nil {
		return err
	}

	err = c.writeGitignoreFile(outputDir)
	if err != nil {
		return err
	}

	err = c.initializeRepository(outputDir)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"files.copied":       c.stats.filesCopied,
		"dirs.created":       c.stats.dirsCreated,
		"templates.rendered": c.stats.templatesRendered,
	}).Infof("project created")

	return nil
}

func (c *creator) processFiles(skeleton *skeleton.Info, dstPath string, values map[string]interface{}) error {
	templateData := map[string]interface{}{
		"ProjectName": c.config.Project.Name, // left here for backwards compat
		"Project":     &c.config.Project,
		"Values":      values,
		"License":     c.license,
		"Git":         &c.config.Git,
	}

	dirMap := make(map[string]string)

	srcPath := skeleton.Path

	return skeleton.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		srcRelPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		srcFilename := filepath.Base(srcRelPath)
		srcRelDir := filepath.Dir(srcRelPath)

		dstFilename, err := renderFilename(srcFilename, templateData)
		if err != nil {
			return err
		}

		dstRelDir := srcRelDir

		// If the src dir's name was templated, lookup the resolved name and
		// use that as destination.
		if dir, ok := dirMap[srcRelDir]; ok {
			dstRelDir = dir
		}

		dstRelPath := filepath.Join(dstRelDir, dstFilename)

		// Sanity check to guard against malicious injection of directory
		// traveral (e.g. "../../" in the template string).
		if filepath.Dir(dstRelPath) != dstRelDir {
			return fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, dstFilename)
		}

		outputPath := filepath.Join(dstPath, dstRelPath)

		log := log.WithField("path", srcRelPath)
		if srcRelPath != dstRelPath {
			log = log.WithField("path.target", dstRelPath)
		}

		if info.IsDir() {
			log.Info("creating directory")

			// Track potentially templated directory names that need to be
			// rewritten for all files contained in them.
			if srcRelPath != dstRelPath {
				dirMap[srcRelPath] = dstRelPath
			}

			return c.makeDirectory(outputPath, info.Mode())
		}

		ext := filepath.Ext(path)
		if ext != ".skel" {
			log.Info("copying file")

			return c.copyFile(path, outputPath)
		}

		// strip .skel extension
		dstRelPath = dstRelPath[:len(dstRelPath)-len(ext)]
		outputPath = filepath.Join(dstPath, dstRelPath)

		log.WithField("path.target", dstRelPath).Info("rendering template")

		return c.writeTemplate(path, outputPath, info.Mode(), templateData)
	})
}

func (c *creator) makeDirectory(path string, mode os.FileMode) error {
	c.stats.dirsCreated++

	if c.dryRun {
		return nil
	}

	return os.MkdirAll(path, mode)
}

func (c *creator) copyFile(src, dst string) error {
	c.stats.filesCopied++

	if c.dryRun {
		return nil
	}

	return file.Copy(src, dst)
}

func (c *creator) writeTemplate(src, dst string, mode os.FileMode, data interface{}) error {
	rendered, err := template.RenderFile(src, data)
	if err != nil {
		return err
	}

	c.stats.templatesRendered++

	if c.dryRun {
		return nil
	}

	return ioutil.WriteFile(dst, []byte(rendered), mode)
}

func (c *creator) writeLicenseFile(outputPath string) error {
	if c.license == nil {
		return nil
	}

	outputPath = filepath.Join(outputPath, "LICENSE")

	log.WithField("path", "LICENSE").Infof("writing %s", c.license.Name)

	if c.dryRun {
		return nil
	}

	body := strings.ReplaceAll(c.license.Body, "[fullname]", c.config.Project.AuthorString())
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}

func (c *creator) writeGitignoreFile(outputPath string) error {
	if c.gitignore == "" {
		return nil
	}

	outputPath = filepath.Join(outputPath, ".gitignore")

	log.Info("writing .gitignore")

	if c.dryRun {
		return nil
	}

	return ioutil.WriteFile(outputPath, []byte(c.gitignore), 0644)
}

func (c *creator) initializeRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err == nil || err != git.ErrRepositoryNotExists {
		return err
	}

	log.WithField("path", path).Info("initializing git repository")

	if c.dryRun {
		return nil
	}

	_, err = git.PlainInit(path, false)

	return err
}

func renderFilename(filenameTemplate string, data interface{}) (string, error) {
	filename, err := template.RenderText(filenameTemplate, data)
	if err != nil {
		return "", err
	}

	if len(filename) == 0 {
		return "", fmt.Errorf("templated filename %q resolved to an empty string", filenameTemplate)
	}

	return filename, nil
}
