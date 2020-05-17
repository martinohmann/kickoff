package skeleton

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command for listing available project skeletons.
func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available skeletons",
		Long: cmdutil.LongDesc(`
			Lists all skeletons available in the configured repositories.`),
		Example: cmdutil.Examples(`
			# List skeletons only from the "myrepo" repository
			kickoff skeleton list --repository myrepo`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.ConfigFlags.AddFlags(cmd)

	return cmd
}

// ListOptions holds the options for the list command.
type ListOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
}

// Run lists all project skeletons available in the configured skeleton
// repositories.
func (o *ListOptions) Run() error {
	repo, err := skeleton.NewRepositoryAggregate(o.Repositories)
	if err != nil {
		return err
	}

	skeletons, err := repo.SkeletonInfos()
	if err != nil {
		return err
	}

	tw := cli.NewTableWriter(o.Out)
	tw.SetHeader("RepoName", "Name", "Path")

	for _, skeleton := range skeletons {
		path, err := homedir.Collapse(skeleton.Path)
		if err != nil {
			return err
		}

		tw.Append(skeleton.Repo.Name, skeleton.Name, path)
	}

	tw.Render()

	return nil
}
