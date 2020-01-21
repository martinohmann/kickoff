package skeleton

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmdutil"
	"github.com/martinohmann/kickoff/pkg/skeleton"
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

	o.OutputFlags.AddFlags(cmd)
	o.ConfigFlags.AddFlags(cmd)

	return cmd
}

type ShowOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.OutputFlags

	Skeleton string
}

func (o *ShowOptions) Complete(args []string) error {
	o.Skeleton = args[0]

	return o.ConfigFlags.Complete("")
}

func (o *ShowOptions) Run() error {
	repo, err := skeleton.NewMultiRepo(o.Repositories)
	if err != nil {
		return err
	}

	skeleton, err := repo.SkeletonInfo(o.Skeleton)
	if err != nil {
		return err
	}

	config, err := skeleton.Config()
	if err != nil {
		return err
	}

	var buf []byte

	switch o.Output {
	case "json":
		buf, err = json.MarshalIndent(config, "", "  ")
	default:
		buf, err = yaml.Marshal(config)
	}

	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, strings.TrimSpace(string(buf)))

	return nil
}
