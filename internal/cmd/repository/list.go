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
func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{
		IOStreams:  streams,
		OutputFlag: cmdutil.NewOutputFlag("name", "table", "wide"),
	}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List configured skeleton repositories",
		Long: cmdutil.LongDesc(`
			Lists all configured skeleton repositories.`),
		Example: cmdutil.Examples(`
			# List repositories for different config
			kickoff repository list --config custom-config.yaml`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.ConfigFlags.AddFlags(cmd)
	o.OutputFlag.AddFlag(cmd)

	return cmd
}

// ListOptions holds the options for the list command.
type ListOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.OutputFlag
}

// Run lists all configured skeleton repositories.
func (o *ListOptions) Run() error {
	repoNames := make([]string, 0, len(o.Repositories))
	for name := range o.Repositories {
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
			ref, err := kickoff.ParseRepoRef(o.Repositories[name])
			if err != nil {
				return err
			}

			typ, url, revision := makeListTableFields(ref)
			localPath := homedir.MustCollapse(ref.LocalPath())

			tw.Append(name, typ, url, revision, localPath)
		}

		tw.Render()
	default:
		tw := cli.NewTableWriter(o.Out)
		tw.SetHeader("Name", "Type", "URL", "Revision")

		for _, name := range repoNames {
			ref, err := kickoff.ParseRepoRef(o.Repositories[name])
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

	return "local", ref.Path, "-"
}
