package skeleton

import (
	"fmt"
	"io"

	"github.com/martinohmann/kickoff/pkg/repo"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	o := &ListOptions{}

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

	o.Out = cmd.OutOrStdout()

	cmd.Flags().StringVar(&o.URL, "repository-url", o.URL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", repo.DefaultRepositoryURL))

	return cmd
}

type ListOptions struct {
	repo.Config
	Out io.Writer
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
