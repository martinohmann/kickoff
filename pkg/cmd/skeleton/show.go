package skeleton

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/config"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/spf13/cobra"
)

var (
	// ErrInvalidOutputFormat is returned if the output format flag contains an
	// invalid value.
	ErrInvalidOutputFormat = errors.New("--output must be 'yaml' or 'json'")
)

func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{IOStreams: streams, Output: "yaml"}

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show the config of a skeleton",
		Long:  "Show the config of a single skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Skeleton = args[0]

			o.ApplyDefaults()

			err := o.Validate()
			if err != nil {
				return err
			}

			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.Output, "output", o.Output, "Output format")
	cmd.Flags().StringVar(&o.RepositoryURL, "repository-url", o.RepositoryURL, fmt.Sprintf("URL of the skeleton repository. Can be a local path or remote git repository. (defaults to %q if the directory exists)", config.DefaultSkeletonRepositoryURL))

	return cmd
}

type ShowOptions struct {
	cli.IOStreams
	config.Skeletons
	Skeleton string
	Output   string
}

func (o *ShowOptions) Validate() error {
	if o.Output != "yaml" && o.Output != "json" {
		return ErrInvalidOutputFormat
	}

	return nil
}

func (o *ShowOptions) Run() error {
	repo, err := skeleton.OpenRepository(o.RepositoryURL)
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
		buf, err = json.Marshal(config)
	default:
		buf, err = yaml.Marshal(config)
	}

	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, string(buf))

	return nil
}
