package repository

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewAddCmd creates a new command for added a skeleton repository to the
// kickoff config.
func NewAddCmd(f *cmdutil.Factory) *cobra.Command {
	o := &AddOptions{
		IOStreams:  f.IOStreams,
		ConfigPath: f.ConfigPath,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a skeleton repository to the config",
		Long: cmdutil.LongDesc(`
			Adds a skeleton repository to the config. If a config for the same repository name already exists it will be overridden.`),
		Example: cmdutil.Examples(`
			# Add a new skeleton repository
			kickoff repository add myskeletons /path/to/skeleton/repo

			# Add a remote skeleton repository in a specific revision
			kickoff repository add myskeletons https://github.com/martinohmann/kickoff-skeletons --revision v1.0.0`),
		Args: cmdutil.ExactNonEmptyArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveFilterDirs
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.RepoName = args[0]
			o.RepoURL = args[1]

			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.Revision, "revision", o.Revision, "Revision to checkout. Can be a branch name, tag or commit SHA.")

	return cmd
}

// AddOptions holds the options for the add command.
type AddOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	ConfigPath string
	RepoName   string
	RepoURL    string
	Revision   string
}

// Run adds a skeleton repository to the kickoff config.
func (o *AddOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	if _, ok := config.Repositories[o.RepoName]; ok {
		return cmdutil.RepositoryAlreadyExistsError(o.RepoName)
	}

	ref, err := kickoff.ParseRepoRef(o.RepoURL)
	if err != nil {
		return err
	}

	if ref.IsLocal() {
		// ensure local path is absolute
		o.RepoURL = ref.LocalPath()
	} else if o.Revision != "" {
		ref.Revision = o.Revision

		o.RepoURL = ref.String()
	}

	_, err = repository.OpenRef(context.Background(), *ref, nil)
	if err != nil {
		removeCacheDir(ref)
		return err
	}

	config.Repositories[o.RepoName] = o.RepoURL

	err = kickoff.SaveConfig(o.ConfigPath, config)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, color.GreenString("✓"), "Repository added")
	fmt.Fprintf(o.Out, "\nYou can inspect it by running: %s\n", bold.Sprintf("kickoff skeleton list -r %s", o.RepoName))

	return nil
}
