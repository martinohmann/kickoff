package repository

import (
	"sort"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/spf13/cobra"
)

// NewListCmd creates a command for listing all configured skeleton
// repositories.
func NewListCmd(streams cli.IOStreams) *cobra.Command {
	o := &ListOptions{IOStreams: streams}

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

// Run lists all configured skeleton repositories.
func (o *ListOptions) Run() error {
	repoNames := make([]string, 0, len(o.Repositories))
	for name := range o.Repositories {
		repoNames = append(repoNames, name)
	}

	sort.Strings(repoNames)

	tw := cli.NewTableWriter(o.Out)
	tw.SetHeader("Name", "Type", "Path", "URL", "Revision")

	for _, name := range repoNames {
		repoURL := o.Repositories[name]

		info, err := skeleton.ParseRepositoryURL(repoURL)
		if err != nil {
			return err
		}

		url := "-"
		revision := "-"
		typ := "local"

		if !info.Local {
			url = info.String()
			typ = "remote"

			if info.Revision != "" {
				revision = info.Revision
			}
		}

		path, err := homedir.Collapse(info.LocalPath())
		if err != nil {
			return err
		}

		tw.Append(name, typ, path, url, revision)
	}

	tw.Render()

	return nil
}
