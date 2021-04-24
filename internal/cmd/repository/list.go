package repository

import (
	"fmt"
	"sort"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command for listing all configured skeleton
// repositories.
func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	o := &ListOptions{
		IOStreams: f.IOStreams,
		Config:    f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured skeleton repositories",
		Long: cmdutil.LongDesc(`
			Lists all configured skeleton repositories.`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmdutil.AddOutputFlag(cmd, &o.Output, "table", "wide", "name")

	return cmd
}

// ListOptions holds the options for the list command.
type ListOptions struct {
	cli.IOStreams

	Config func() (*kickoff.Config, error)

	Output string
}

// Run lists all configured skeleton repositories.
func (o *ListOptions) Run() error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	repos := config.Repositories

	repoNames := make([]string, 0, len(repos))
	for name := range repos {
		repoNames = append(repoNames, name)
	}

	sort.Strings(repoNames)

	switch o.Output {
	case "name":
		for _, name := range repoNames {
			fmt.Fprintln(o.Out, name)
		}
	case "wide":
		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Name", "Type", "URL", "Revision", "Local Path")

		for _, name := range repoNames {
			ref, err := kickoff.ParseRepoRef(repos[name])
			if err != nil {
				return err
			}

			typ, url, revision := makeListTableFields(ref)
			localPath := homedir.Collapse(ref.LocalPath())

			tw.Append(name, typ, url, revision, localPath)
		}

		tw.Render()
	default:
		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Name", "Type", "URL", "Revision")

		for _, name := range repoNames {
			ref, err := kickoff.ParseRepoRef(repos[name])
			if err != nil {
				return err
			}

			typ, url, revision := makeListTableFields(ref)

			tw.Append(name, typ, url, revision)
		}

		tw.Render()
	}

	return nil
}

func makeListTableFields(ref *kickoff.RepoRef) (typ string, url string, rev string) {
	if ref.IsRemote() {
		revision := "<default-branch>"

		if ref.Revision != "" {
			revision = ref.Revision
		}

		return "remote", ref.URL, revision
	}

	return "local", homedir.Collapse(ref.Path), "-"
}
