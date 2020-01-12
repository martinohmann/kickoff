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
		Long:  "Show the config of a single skeleton",
		Args:  cobra.ExactArgs(1),
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
	cmdutil.AddRepositoryURLFlag(cmd, &o.Skeletons.RepositoryURL)

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
	repo, err := skeleton.OpenRepository(o.Skeletons.RepositoryURL)
	if err != nil {
		return err
	}

	skeleton, err := repo.Skeleton(o.Skeleton)
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
