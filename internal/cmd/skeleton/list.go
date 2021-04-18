package skeleton

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command for listing available project skeletons.
func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{
		IOStreams:   streams,
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

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
	o.TimeoutFlag.AddFlag(cmd)

	return cmd
}

// ListOptions holds the options for the list command.
type ListOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag
}

// Run lists all project skeletons available in the configured skeleton
// repositories.
func (o *ListOptions) Run() error {
	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	repo, err := repository.OpenMap(ctx, o.Repositories, nil)
	if err != nil {
		return err
	}

	skeletons, err := repo.ListSkeletons()
	if err != nil {
		return err
	}

	tw := cli.NewTableWriter(o.Out)
	tw.SetHeader("RepoName", "Name", "Path")

	for _, skeleton := range skeletons {
		path := homedir.MustCollapse(skeleton.Path)

		tw.Append(skeleton.Repo.Name, skeleton.Name, path)
	}

	tw.Render()

	return nil
}
