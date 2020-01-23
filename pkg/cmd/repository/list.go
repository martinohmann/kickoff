package repository

import (
	"sort"

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

type ListOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
}

func (o *ListOptions) Run() error {
	repoNames := make([]string, 0, len(o.Repositories))
	for name := range o.Repositories {
		repoNames = append(repoNames, name)
	}

	sort.Strings(repoNames)

	tw := cli.NewTableWriter(o.Out)
	tw.SetHeader("Name", "Type", "Branch", "URL", "LocalPath")

	for _, name := range repoNames {
		url := o.Repositories[name]

		info, err := skeleton.ParseRepositoryURL(url)
		if err != nil {
			return err
		}

		typ := "remote"
		if info.Local {
			typ = "local"
		}

		branch := "-"
		if info.Branch != "" {
			branch = info.Branch
		}

		tw.Append(name, typ, branch, info.String(), info.LocalPath())
	}

	tw.Render()

	return nil
}
