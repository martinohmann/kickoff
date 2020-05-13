package skeleton

import (
	"strings"

	"github.com/disiqueira/gotree"
	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/spf13/cobra"
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{IOStreams: streams}

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show the config of a skeleton",
		Long: cmdutil.LongDesc(`
			Show the config of a single skeleton.`),
		Example: cmdutil.Examples(`
			# Show skeleton config
			kickoff skeleton show myskeleton

			# Show skeleton config in a specific repository
			kickoff skeleton show myrepo:myskeleton

			# Show skeleton config using different output
			kickoff skeleton show myskeleton --output json`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.OutputFlag.AddFlag(cmd)
	o.ConfigFlags.AddFlags(cmd)

	return cmd
}

type ShowOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.OutputFlag

	Skeleton string
}

func (o *ShowOptions) Complete(args []string) error {
	o.Skeleton = args[0]

	return o.ConfigFlags.Complete()
}

func (o *ShowOptions) Run() error {
	loader, err := skeleton.NewRepositoryAggregateLoader(o.Repositories)
	if err != nil {
		return err
	}

	skeleton, err := loader.LoadSkeleton(o.Skeleton)
	if err != nil {
		return err
	}

	var buf []byte

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, skeleton)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, skeleton)
	default:
		tw := cli.NewTableWriter(o.Out)

		description := strings.TrimSpace(skeleton.Description)

		if len(description) == 0 {
			description = "-"
		}

		tree := gotree.New(skeleton.Info.Name)

		for _, file := range skeleton.Files {
			if file.RelPath == "." {
				continue
			}

			tree.Add(file.RelPath)
		}

		parent := "-"
		if skeleton.Parent != nil {
			parent, err = homedir.Collapse(skeleton.Parent.Info.Path)
			if err != nil {
				return err
			}
		}

		values := "-"
		if len(skeleton.Values) > 0 {
			buf, err = yaml.Marshal(skeleton.Values)
			if err != nil {
				return err
			}

			values = strings.TrimSpace(string(buf))
		}

		path, err := homedir.Collapse(skeleton.Info.Path)
		if err != nil {
			return err
		}

		tw.Append("Name", skeleton.Info.Name)
		tw.Append("Path", path)
		tw.Append("Description", description)
		tw.Append("Files", strings.TrimSpace(tree.Print()))
		tw.Append("Parent", parent)
		tw.Append("Values", values)

		tw.Render()

		return nil
	}
}
