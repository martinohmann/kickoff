package skeleton

import (
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available skeletons",
		Long: cmdutil.LongDesc(`
			Lists all skeletons available in the configured repositories.`),
		Example: cmdutil.Examples(`
			# List skeletons from custom repositories
			kickoff skeleton list --repositories my-repo=https://github.com/martinohmann/kickoff-skeletons`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(""); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.ConfigFlags.AddFlags(cmd)

	return cmd
}

type ListOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
}

func (o *ListOptions) Run() error {
	repo, err := skeleton.NewMultiRepo(o.Repositories)
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
		tw.Append(skeleton.Repo.Name, skeleton.Name, skeleton.Path)
	}

	tw.Render()

	return nil
}
