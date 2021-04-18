package skeleton

import (
	"fmt"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command for listing available project skeletons.
func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	o := &ListOptions{
		IOStreams:  f.IOStreams,
		Repository: f.Repository,
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
			return o.Run()
		},
	}

	cmdutil.AddOutputFlag(cmd, &o.Output, "table", "wide", "name")
	cmdutil.AddRepositoryFlag(cmd, f, &o.RepoNames)

	return cmd
}

// ListOptions holds the options for the list command.
type ListOptions struct {
	cli.IOStreams

	Repository func(...string) (kickoff.Repository, error)

	Output    string
	RepoNames []string
}

// Run lists all project skeletons available in the configured skeleton
// repositories.
func (o *ListOptions) Run() error {
	repo, err := o.Repository(o.RepoNames...)
	if err != nil {
		return err
	}

	skeletons, err := repo.ListSkeletons()
	if err != nil {
		return err
	}

	switch o.Output {
	case "name":
		for _, skeleton := range skeletons {
			fmt.Fprintln(o.Out, skeleton.String())
		}
	case "wide":
		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Repository", "Name", "Path")

		for _, skeleton := range skeletons {
			path := homedir.MustCollapse(skeleton.Path)

			tw.Append(skeleton.Repo.Name, skeleton.Name, path)
		}

		tw.Render()
	default:
		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Repository", "Name")

		for _, skeleton := range skeletons {
			tw.Append(skeleton.Repo.Name, skeleton.Name)
		}

		tw.Render()
	}

	return nil
}
