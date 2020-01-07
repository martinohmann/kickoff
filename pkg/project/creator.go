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
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/git"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	gogit "gopkg.in/src-d/go-git.v4"
)

type CreateOptions struct {
	DryRun  bool
	License *license.Info
	Git     git.Config
	Values  template.Values
}

type Creator struct {
	config      Config
	createStats *createStats
}

func NewCreator(config Config) *Creator {
	return &Creator{
		config:      config,
		createStats: &createStats{},
	}
}

// Create creates the project with given options.
func (c *Creator) Create(skeleton *skeleton.Info, outputDir string, options *CreateOptions) error {
	if options == nil {
		options = &CreateOptions{}
	}

	if options.DryRun {
		log.Warn("dry run: no changes will be made")
	}

	config, err := skeleton.Config()
	if err != nil {
		return err
	}

	err = config.Values.Merge(options.Values)
	if err != nil {
		return err
	}

	log.WithField("values", fmt.Sprintf("%#v", config.Values)).Debug("merged values")

	log.WithFields(log.Fields{
		"skeleton": skeleton.Path,
		"target":   outputDir,
	}).Info("creating project from skeleton")

	err = c.processFiles(skeleton, outputDir, config.Values, options)
	if err != nil {
		return err
	}

	if options.License != nil {
		err = c.writeLicenseFile(outputDir, options)
		if err != nil {
			return err
		}
	}

	err = c.initializeRepository(outputDir, options.DryRun)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"files.copied":       c.createStats.filesCopied,
		"dirs.created":       c.createStats.dirsCreated,
		"templates.rendered": c.createStats.templatesRendered,
	}).Infof("project created")

	return nil
}

func (c *Creator) processFiles(skeleton *skeleton.Info, dstPath string, values map[string]interface{}, options *CreateOptions) error {
	templateData := map[string]interface{}{
		"ProjectName": c.config.Name, // left here for backwards compat
		"Project":     &c.config,
		"Values":      values,
		"License":     options.License,
		"Git":         &options.Git,
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

		dstFilename, err := renderDestinationFilename(srcFilename, templateData)
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

			return c.makeDirectory(outputPath, info.Mode(), options.DryRun)
		}

		ext := filepath.Ext(path)
		if ext != ".skel" {
			log.Info("copying file")

			return c.copyFile(path, outputPath, options.DryRun)
		}

		// strip .skel extension
		dstRelPath = dstRelPath[:len(dstRelPath)-len(ext)]
		outputPath = filepath.Join(dstPath, dstRelPath)

		log.WithField("path.target", dstRelPath).Info("rendering template")

		return c.writeTemplate(path, outputPath, info.Mode(), templateData, options.DryRun)
	})
}

func (c *Creator) makeDirectory(path string, mode os.FileMode, dryRun bool) error {
	c.createStats.dirsCreated++

	if dryRun {
		return nil
	}

	return os.MkdirAll(path, mode)
}

func (c *Creator) copyFile(src, dst string, dryRun bool) error {
	c.createStats.filesCopied++

	if dryRun {
		return nil
	}

	return file.Copy(src, dst)
}

func (c *Creator) writeTemplate(src, dst string, mode os.FileMode, data interface{}, dryRun bool) error {
	rendered, err := template.RenderFile(src, data)
	if err != nil {
		return err
	}

	c.createStats.templatesRendered++

	if dryRun {
		return nil
	}

	return ioutil.WriteFile(dst, []byte(rendered), mode)
}

func (c *Creator) writeLicenseFile(outputPath string, options *CreateOptions) error {
	outputPath = filepath.Join(outputPath, "LICENSE")

	log.WithField("path", "LICENSE").Infof("writing %s", options.License.Name)

	if options.DryRun {
		return nil
	}

	body := strings.ReplaceAll(options.License.Body, "[fullname]", c.config.AuthorString())
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}

func (c *Creator) initializeRepository(path string, dryRun bool) error {
	_, err := gogit.PlainOpen(path)
	if err == nil || err != gogit.ErrRepositoryNotExists {
		return err
	}

	log.WithField("path", path).Info("initializing git repository")

	if dryRun {
		return nil
	}

	_, err = gogit.PlainInit(path, false)

	return err
}

func renderDestinationFilename(srcFilename string, data interface{}) (string, error) {
	dstFilename, err := template.RenderText(srcFilename, data)
	if err != nil {
		return "", err
	}

	if len(dstFilename) == 0 {
		return "", fmt.Errorf("templated filename %q resolved to an empty string", srcFilename)
	}

	return dstFilename, nil
}

type createStats struct {
	filesCopied       int64
	dirsCreated       int64
	templatesRendered int64
}
