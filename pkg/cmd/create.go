package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/apex/log"
	"github.com/imdario/mergo"
	"github.com/martinohmann/skeleton-go/pkg/config"
	"github.com/martinohmann/skeleton-go/pkg/file"
	"github.com/martinohmann/skeleton-go/pkg/license"
	"github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

func NewCreateCmd() *cobra.Command {
	o := &CreateOptions{}

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

type CreateOptions struct {
	InputDir  string
	OutputDir string
	DryRun    bool
	Force     bool

	ConfigPath string
	Config     config.Config

	LicenseInfo *license.Info
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
		fileConfig, err := config.Load(o.ConfigPath)
		if err != nil {
			return err
		}

		err = mergo.Merge(&o.Config, fileConfig)
		if err != nil {
			return err
		}
	}

	err = o.Config.Complete(o.OutputDir)
	if err != nil {
		return err
	}

	o.InputDir = o.Config.SkeletonDir()

	if o.Config.License != "" {
		o.LicenseInfo, err = o.fetchLicenseInfo(o.Config.License)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *CreateOptions) Validate() error {
	if file.Exists(o.OutputDir) && !o.Force {
		return fmt.Errorf("output-dir %q already exists, add --force to overwrite", o.OutputDir)
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
		log.Warn("DRY RUN: no changes will be made")
	}

	err := o.processFiles(o.InputDir, o.OutputDir)
	if err != nil {
		return err
	}

	err = o.writeLicenseFile(o.OutputDir)
	if err != nil {
		return err
	}

	return o.initGitRepository(o.OutputDir)
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
	log.WithField("skeleton", o.Config.Skeleton).Info("creating project from skeleton")

	log.Debugf("using config: %#v", o.Config)

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

		outputPath := filepath.Join(dstPath, relPath)

		if info.IsDir() {
			return o.ensureDirectory(outputPath, info.Mode())
		}

		if ext := filepath.Ext(path); ext == ".skel" {
			outputPath = outputPath[:len(outputPath)-5]

			return o.writeTemplate(path, outputPath, info.Mode(), templateData)
		}

		return o.copyFile(path, outputPath)
	})
}

func (o *CreateOptions) ensureDirectory(path string, mode os.FileMode) error {
	log.WithFields(log.Fields{"path": path}).Info("creating directory")

	if o.DryRun {
		return nil
	}

	if file.Exists(path) {
		log.WithFields(log.Fields{"path": path}).Warn("directory already exists")
		return nil
	}

	return os.MkdirAll(path, mode)
}

func (o *CreateOptions) copyFile(src, dst string) error {
	log.WithFields(log.Fields{"src": src, "dst": dst}).Info("copying file")

	if o.DryRun {
		return nil
	}

	if file.Exists(dst) {
		log.WithFields(log.Fields{"dst": dst}).Warn("file already exists")
	}

	return file.Copy(src, dst)
}

func (o *CreateOptions) writeTemplate(src, dst string, mode os.FileMode, data interface{}) error {
	name := filepath.Base(src)

	tpl, err := template.New(name).ParseFiles(src)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	err = tpl.Execute(&buf, data)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"src": src, "dst": dst}).Info("writing template")

	if o.DryRun {
		return nil
	}

	if file.Exists(dst) {
		log.WithFields(log.Fields{"dst": dst}).Warn("file already exists")
	}

	return ioutil.WriteFile(dst, buf.Bytes(), mode)
}

func (o *CreateOptions) writeLicenseFile(outputPath string) error {
	if o.LicenseInfo == nil {
		return nil
	}

	outputPath = filepath.Join(outputPath, "LICENSE")

	log.WithField("path", outputPath).Info("writing LICENSE file")

	if o.DryRun {
		return nil
	}

	body := strings.ReplaceAll(o.LicenseInfo.Body, "[fullname]", o.Config.Author.String())
	body = strings.ReplaceAll(body, "[year]", strconv.Itoa(time.Now().Year()))

	return ioutil.WriteFile(outputPath, []byte(body), 0644)
}

func (o *CreateOptions) initGitRepository(path string) error {
	_, err := git.PlainOpen(path)
	if err != nil && err != git.ErrRepositoryNotExists {
		return err
	}

	log.WithField("path", path).Info("initializing git repository")

	if o.DryRun {
		return nil
	}

	_, err = git.PlainInit(path, false)

	return err
}
