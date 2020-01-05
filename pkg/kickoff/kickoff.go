package kickoff

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
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/template"
	git "gopkg.in/src-d/go-git.v4"
)

// Kickoff implements the core functionality for bootstrapping projects from
// skeletons.
type Kickoff struct {
	config      *config.Config
	licenseInfo *license.Info
	dryRun      bool
	stats       *stats
}

// New creates a new *Kickoff value with given config. If dryRun is set to
// true, all actions that would be carried out will just be printed but not
// actually executed.
func New(config *config.Config, dryRun bool) *Kickoff {
	return &Kickoff{
		config: config,
		dryRun: dryRun,
		stats:  &stats{},
	}
}

// Create creates a new project in outputDir based on the config passed to New.
// Returns any error that occurs along the way.
func (k *Kickoff) Create(outputDir string) (err error) {
	if k.dryRun {
		log.Warn("dry run: no changes will be made")
	}

	if k.config.Skeleton.License != "" {
		log.WithField("license", k.config.Skeleton.License).Debugf("fetching license info from GitHub")

		k.licenseInfo, err = fetchLicenseInfo(k.config.Skeleton.License)
		if err != nil {
			return err
		}
	}

	inputDir := k.config.SkeletonDir()

	log.WithFields(log.Fields{
		"skeleton": inputDir,
		"target":   outputDir,
	}).Info("creating project from skeleton")

	log.WithField("config", fmt.Sprintf("%#v", k.config)).Debug("using config")

	err = k.processFiles(inputDir, outputDir)
	if err != nil {
		return err
	}

	err = k.writeLicenseFile(outputDir)
	if err != nil {
		return err
	}

	err = k.initializeRepository(outputDir)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"files.copied":       k.stats.filesCopied,
		"dirs.created":       k.stats.dirsCreated,
		"templates.rendered": k.stats.templatesRendered,
	}).Infof("project created")

	return nil
}

func (k *Kickoff) processFiles(srcPath, dstPath string) error {
	templateData := map[string]interface{}{
		"Author":      k.config.Author,
		"Values":      k.config.Skeleton.Values,
		"License":     k.licenseInfo,
		"ProjectName": k.config.ProjectName,
		"Repository":  k.config.Repository,
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		srcRelPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		if srcRelPath == config.SkeletonConfigFile {
			// ignore skeleton config file
			return nil
		}

		srcFilename := filepath.Base(srcRelPath)
		srcRelDir := filepath.Dir(srcRelPath)

		dstFilename, err := renderDestinationFilename(srcFilename, templateData)
		if err != nil {
			return err
		}

		dstRelPath := filepath.Join(srcRelDir, dstFilename)

		// sanity check
		if filepath.Dir(dstRelPath) != srcRelDir {
			return fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, dstFilename)
		}

		outputPath := filepath.Join(dstPath, dstRelPath)

		log := log.WithField("path", srcRelPath)
		if srcRelPath != dstRelPath {
			log = log.WithField("path.target", dstRelPath)
		}

		if info.IsDir() {
			log.Info("creating directory")

			return k.makeDirectory(outputPath, info.Mode())
		}

		ext := filepath.Ext(path)
		if ext != ".skel" {
			log.Info("copying file")

			return k.copyFile(path, outputPath)
		}

		// strip .skel extension
		dstRelPath = dstRelPath[:len(dstRelPath)-len(ext)]
		outputPath = filepath.Join(dstPath, dstRelPath)

		log.WithField("path.target", dstRelPath).Info("rendering template")

		return k.writeTemplate(path, outputPath, info.Mode(), templateData)
	})
}

func (k *Kickoff) makeDirectory(path string, mode os.FileMode) error {
	k.stats.dirsCreated++

	if k.dryRun {
		return nil
	}

	return os.MkdirAll(path, mode)
}

func (k *Kickoff) copyFile(src, dst string) error {
	k.stats.filesCopied++

	if k.dryRun {
		return nil
	}

	return file.Copy(src, dst)
}

func (k *Kickoff) writeTemplate(src, dst string, mode os.FileMode, data interface{}) error {
	rendered, err := template.RenderFile(src, data)
	if err != nil {
		return err
	}

	k.stats.templatesRendered++

	if k.dryRun {
		return nil
	}

	return ioutil.WriteFile(dst, []byte(rendered), mode)
}

func (k *Kickoff) writeLicenseFile(outputPath string) error {
	if k.licenseInfo == nil {
		return nil
	}

	outputPath = filepath.Join(outputPath, "LICENSE")

	log.WithField("path", "LICENSE").Infof("writing %s", k.licenseInfo.Name)

	if k.dryRun {
		return nil
	}

	body := strings.ReplaceAll(k.licenseInfo.Body, "[fullname]", k.config.Author.String())
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}

func (k *Kickoff) initializeRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err == nil || err != git.ErrRepositoryNotExists {
		return err
	}

	log.WithField("path", path).Info("initializing git repository")

	if k.dryRun {
		return nil
	}

	_, err = git.PlainInit(path, false)

	return err
}

func fetchLicenseInfo(name string) (*license.Info, error) {
	info, err := license.Get(name)
	if err == license.ErrLicenseNotFound {
		return nil, fmt.Errorf("license %q not found, use the `licenses` subcommand to get a list of available licenses", name)
	} else if err != nil {
		return nil, err
	}

	return info, nil
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

type stats struct {
	filesCopied       int64
	dirsCreated       int64
	templatesRendered int64
}
