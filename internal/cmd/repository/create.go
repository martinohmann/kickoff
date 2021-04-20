package repository

import (
	"fmt"
	"os"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a command for creating a local skeleton repository.
func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	o := &CreateOptions{
		IOStreams:    f.IOStreams,
		Config:       f.Config,
		ConfigPath:   f.ConfigPath,
		SkeletonName: kickoff.DefaultSkeletonName,
	}

	cmd := &cobra.Command{
		Use:   "create <name> <dir>",
		Short: "Create a new skeleton repository",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton repository with a default skeleton to get you started.`),
		Example: cmdutil.Examples(`
			# Create a new repository
			kickoff repository create myrepo /repository/output/path`),
		Args: cmdutil.ExactNonEmptyArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveFilterDirs
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			o.RepoName = args[0]
			o.RepoDir = args[1]

			return o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.SkeletonName, "skeleton-name", "s", o.SkeletonName,
		"Name of the default skeleton that will be created in the new repository.")

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	ConfigPath   string
	RepoName     string
	RepoDir      string
	SkeletonName string
}

// Run creates a new skeleton repository in the provided output directory and
// seeds it with a default skeleton.
func (o *CreateOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	if _, ok := config.Repositories[o.RepoName]; ok {
		return cmdutil.RepositoryAlreadyExistsError(o.RepoName)
	}

	repo, err := repository.Create(o.RepoDir)
	if err != nil {
		return err
	}

	if _, err := repo.CreateSkeleton(o.SkeletonName); err != nil {
		if err := os.RemoveAll(o.RepoDir); err != nil {
			log.WithError(err).
				WithField("path", o.RepoDir).
				Error("failed to remove newly created repository directory")
		}

		return err
	}

	config.Repositories[o.RepoName] = o.RepoDir

	err = kickoff.SaveConfig(o.ConfigPath, config)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Created new skeleton repository in %s.\n\n", o.RepoDir)
	fmt.Fprintf(o.Out, "You can inspect it by running `kickoff skeleton list -r %s`.\n", o.RepoName)

	return nil
}
