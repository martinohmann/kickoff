package skeleton

import (
	"fmt"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available skeletons",
		Long:    "Lists all skeletons available in the repository",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.ApplyDefaults()

			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.RepositoryURL, "repository-url", o.RepositoryURL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", config.DefaultSkeletonRepositoryURL))

	return cmd
}

type ListOptions struct {
	cli.IOStreams
	config.Skeletons
}

func (o *ListOptions) Run() error {
	repo, err := skeleton.OpenRepository(o.RepositoryURL)
	if err != nil {
		return err
	}

	skeletons, err := repo.Skeletons()
	if err != nil {
		return err
	}

	tw := cli.NewTableWriter(o.Out)
	tw.SetHeader("Name", "Path")

	for _, skeleton := range skeletons {
		tw.Append(skeleton.Name, skeleton.Path)
	}

	tw.Render()

	return nil
}
