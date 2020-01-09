package skeleton

import (
	"fmt"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/repo"
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

	cmd.Flags().StringVar(&o.URL, "repository-url", o.URL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", repo.DefaultRepositoryURL))

	return cmd
}

type ListOptions struct {
	cli.IOStreams
	repo.Config
}

func (o *ListOptions) Run() error {
	repo, err := repo.Open(o.URL)
	if err != nil {
		return err
	}

	skeletons, err := repo.Skeletons()
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Skeletons available in %s:\n\n", o.URL)

	for _, skeleton := range skeletons {
		fmt.Fprintf(o.Out, "%s => %s\n", skeleton.Name, skeleton.Path)
	}

	return nil
}
