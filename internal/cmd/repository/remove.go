package repository

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRemoveCmd creates a command for removing skeleton repositories from the
// config.
func NewRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	o := &RemoveOptions{
		IOStreams:  f.IOStreams,
		Config:     f.Config,
		ConfigPath: f.ConfigPath,
	}

	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a skeleton repository from the config",
		Long: cmdutil.LongDesc(`
			Removes a skeleton repository from the config. Does not remove local repositories from disk, but cleans up the local cache dir for remote repositories.`),
		Example: cmdutil.Examples(`
			# Remove a skeleton repository
			kickoff repository remove myrepo`),
		Args: cmdutil.ExactNonEmptyArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cmdutil.RepositoryNames(f), cobra.ShellCompDirectiveDefault
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.RepoName = args[0]

			return o.Run()
		},
	}

	return cmd
}

// RemoveOptions holds the options for the remove command.
type RemoveOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	ConfigPath string
	RepoName   string
}

// Run removes a skeleton repository from the config.
func (o *RemoveOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	if _, ok := config.Repositories[o.RepoName]; !ok {
		return cmdutil.RepositoryNotConfiguredError(o.RepoName)
	}

	repoRef, err := kickoff.ParseRepoRef(config.Repositories[o.RepoName])
	if err != nil {
		return err
	}

	if repoRef.IsRemote() {
		removeCacheDir(repoRef)
	}

	delete(config.Repositories, o.RepoName)

	err = kickoff.SaveConfig(o.ConfigPath, config)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, color.GreenString("✓"), "Repository removed")

	return nil
}

func removeCacheDir(ref *kickoff.RepoRef) {
	if ref.IsLocal() {
		return
	}

	localPath := ref.LocalPath()

	// Prevent removal of anything outside of the local user cache dir.
	// Remote repos should never ever reside outside of the user cache dir.
	// If they do this is a programmer error.
	if !strings.HasPrefix(localPath, kickoff.LocalRepositoryCacheDir) {
		log.WithField("path", localPath).
			Panic("unexpected repository location: found remote repository cache outside of user cache dir, refusing to delete")
	}

	log.WithField("path", localPath).Debug("deleting repository cache dir")

	if err := os.RemoveAll(localPath); err != nil {
		log.WithError(err).
			WithField("path", localPath).
			Error("failed to delete repository cache dir")
	}
}
