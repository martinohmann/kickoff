package skeleton

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// NewCreateCmd creates a command for creating project skeletons.
func NewCreateCmd(f *cmdutil.Factory) *cobra.Command {
	o := &CreateOptions{
		IOStreams:  f.IOStreams,
		Repository: f.Repository,
	}

	cmd := &cobra.Command{
		Use:   "create <repo-name> <skeleton-name>",
		Short: "Create a new skeleton in a local repository",
		Long: cmdutil.LongDesc(`
			Creates a new skeleton directory in a local repository with some boilerplate to get started.`),
		Example: cmdutil.Examples(`
			# Create a new skeleton in myrepo
			kickoff skeleton create myrepo myskeleton`),
		Args: cmdutil.ExactNonEmptyArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cmdutil.RepositoryNames(f), cobra.ShellCompDirectiveDefault
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.RepoName = args[0]
			o.SkeletonName = args[1]

			return o.Run()
		},
	}

	return cmd
}

// CreateOptions holds the options for the create command.
type CreateOptions struct {
	cli.IOStreams

	Repository func(...string) (kickoff.Repository, error)

	RepoName     string
	SkeletonName string
}

// Run creates a new project skeleton in the provided output directory.
func (o *CreateOptions) Run() error {
	repo, err := o.Repository(o.RepoName)
	if err != nil {
		return err
	}

	ref, err := repo.CreateSkeleton(o.SkeletonName)
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "%s Created new skeleton %s in repository %s\n\n", color.GreenString("✓"), bold.Sprint(ref.Name), bold.Sprint(o.RepoName))
	fmt.Fprintln(o.Out, "You can inspect it by running:", bold.Sprintf("kickoff skeleton show %s:%s", o.RepoName, ref.Name))

	return nil
}
