package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/imdario/mergo"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/file"
	"github.com/martinohmann/kickoff/pkg/license"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

func NewCreateCmd() *cobra.Command {
	o := NewCreateOptions()

	cmd := &cobra.Command{
		Use:   "create <output-dir>",
		Short: "Create project skeletons",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.AddFlags(cmd)

	return cmd
}

type createStats struct {
	filesCopied       int64
	dirsCreated       int64
	templatesRendered int64
}

type CreateOptions struct {
	InputDir  string
	OutputDir string
	DryRun    bool
	Force     bool

	ConfigPath string
	Config     *config.Config

	LicenseInfo *license.Info

	stats createStats
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		Config: config.NewDefaultConfig(),
	}
}

func (o *CreateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "Only print what would be done")
	cmd.Flags().BoolVar(&o.Force, "force", o.Force, "Forces overwrite of existing output directory")
	cmd.Flags().StringVar(&o.ConfigPath, "config", o.ConfigPath, fmt.Sprintf("Path to config file (defaults to %q if the file exists)", config.DefaultConfigPath))

	o.Config.AddFlags(cmd)
}

func (o *CreateOptions) Complete(args []string) (err error) {
	if args[0] != "" {
		o.OutputDir, err = filepath.Abs(args[0])
		if err != nil {
			return err
		}
	}

	if o.ConfigPath == "" && file.Exists(config.DefaultConfigPath) {
		o.ConfigPath = config.DefaultConfigPath
	}

	if o.ConfigPath != "" {
		log.WithField("path", o.ConfigPath).Debugf("loading config file")

		config, err := config.Load(o.ConfigPath)
		if err != nil {
			return err
		}

		err = mergo.Merge(o.Config, config)
		if err != nil {
			return err
		}
	}

	err = o.Config.Complete(o.OutputDir)
	if err != nil {
		return err
	}

	o.InputDir = o.Config.SkeletonDir()

	skeletonConfigPath := o.Config.SkeletonConfigPath()

	if file.Exists(skeletonConfigPath) {
		log.WithField("skeleton", o.InputDir).Debugf("found %s, merging config values", config.SkeletonConfigFile)

		config, err := config.Load(skeletonConfigPath)
		if err != nil {
			return err
		}

		err = mergo.Merge(o.Config, config)
		if err != nil {
			return err
		}
	}

	if o.Config.License != "" && o.Config.License != "none" {
		log.WithField("license", o.Config.License).Debugf("fetching license info from GitHub")

		o.LicenseInfo, err = o.fetchLicenseInfo(o.Config.License)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *CreateOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output-dir %s already exists, add --force to overwrite", o.OutputDir)
	}

	ok, err := file.IsDirectory(o.InputDir)
	if err != nil {
		return fmt.Errorf("failed to stat input directory: %v", err)
	}

	if !ok {
		return fmt.Errorf("invalid skeleton: %s is not a directory", o.InputDir)
	}

	if o.OutputDir == "" {
		return errors.New("output-dir must not be an empty string")
	}

	return o.Config.Validate()
}

func (o *CreateOptions) Run() error {
	if o.DryRun {
		log.Warn("dry run: no changes will be made")
	}

	log.WithFields(log.Fields{
		"skeleton": o.InputDir,
		"target":   o.OutputDir,
	}).Info("creating project from skeleton")

	log.WithField("config", fmt.Sprintf("%#v", o.Config)).Debug("using config")

	err := o.processFiles(o.InputDir, o.OutputDir)
	if err != nil {
		return err
	}

	err = o.writeLicenseFile(o.OutputDir)
	if err != nil {
		return err
	}

	err = o.initGitRepository(o.OutputDir)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"files.copied":       o.stats.filesCopied,
		"dirs.created":       o.stats.dirsCreated,
		"templates.rendered": o.stats.templatesRendered,
	}).Infof("project created")

	return nil
}

func (o *CreateOptions) fetchLicenseInfo(name string) (*license.Info, error) {
	info, err := license.Get(name)
	if err == license.ErrLicenseNotFound {
		return nil, fmt.Errorf("license %q not found, use the `licenses` subcommand to get a list of available licenses", name)
	} else if err != nil {
		return nil, err
	}

	return info, nil
}

func (o *CreateOptions) processFiles(srcPath, dstPath string) error {
	templateData := map[string]interface{}{
		"Author":      o.Config.Author,
		"Custom":      o.Config.Custom,
		"License":     o.LicenseInfo,
		"ProjectName": o.Config.ProjectName,
		"Repository":  o.Config.Repository,
	}

	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		if relPath == config.SkeletonConfigFile {
			// ignore skeleton config file
			return nil
		}

		outputPath := filepath.Join(dstPath, relPath)

		if info.IsDir() {
			log.WithField("path", outputPath).Info("creating directory")

			return o.makeDirectory(outputPath, info.Mode())
		}

		if ext := filepath.Ext(path); ext == ".skel" {
			outputPath = outputPath[:len(outputPath)-5]

			log.WithField("template", relPath).Info("rendering template")

			return o.writeTemplate(path, outputPath, info.Mode(), templateData)
		}

		log.WithField("file", relPath).Info("copying file")

		return o.copyFile(path, outputPath)
	})
}

func (o *CreateOptions) makeDirectory(path string, mode os.FileMode) error {
	o.stats.dirsCreated++

	if o.DryRun {
		return nil
	}

	return os.MkdirAll(path, mode)
}

func (o *CreateOptions) copyFile(src, dst string) error {
	o.stats.filesCopied++

	if o.DryRun {
		return nil
	}

	return file.Copy(src, dst)
}

func (o *CreateOptions) writeTemplate(src, dst string, mode os.FileMode, data interface{}) error {
	buf, err := template.Render(src, data)
	if err != nil {
		return err
	}

	o.stats.templatesRendered++

	if o.DryRun {
		return nil
	}

	return ioutil.WriteFile(dst, buf, mode)
}

func (o *CreateOptions) writeLicenseFile(outputPath string) error {
	if o.LicenseInfo == nil {
		return nil
	}

	outputPath = filepath.Join(outputPath, "LICENSE")

	log.WithField("path", "LICENSE").Infof("writing %s", o.LicenseInfo.Name)

	if o.DryRun {
		return nil
	}

	body := strings.ReplaceAll(o.LicenseInfo.Body, "[fullname]", o.Config.Author.String())
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}

func (o *CreateOptions) initGitRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err == nil || err != git.ErrRepositoryNotExists {
		return err
	}

	log.WithField("path", path).Info("initializing git repository")

	if o.DryRun {
		return nil
	}

	_, err = git.PlainInit(path, false)

	return err
}
