package skeleton

import (
	"bytes"
	"strings"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/filetree"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewShowCmd creates a command for inspecting project skeletons.
func NewShowCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowOptions{
		IOStreams:   streams,
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

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
		Args: cmdutil.ExactNonEmptyArgs(1),
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
	o.TimeoutFlag.AddFlag(cmd)

	return cmd
}

// ShowOptions holds options for the show command.
type ShowOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.OutputFlag
	cmdutil.TimeoutFlag

	SkeletonName string
}

// Complete completes the show options.
func (o *ShowOptions) Complete(args []string) error {
	o.SkeletonName = args[0]

	return o.ConfigFlags.Complete()
}

// Run prints information about a project skeleton in the output format
// specified by the user.
func (o *ShowOptions) Run() error {
	ctx, cancel := o.TimeoutFlag.Context()
	defer cancel()

	repo, err := repository.NewFromMap(o.Repositories)
	if err != nil {
		return err
	}

	skeleton, err := repository.LoadSkeleton(ctx, repo, o.SkeletonName)
	if err != nil {
		return err
	}

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, skeleton)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, skeleton)
	default:
		tw := cli.NewTableWriter(o.Out)

		path := homedir.MustCollapse(skeleton.Ref.Path)

		repoInfo := skeleton.Ref.Repo

		tw.Append("Name", skeleton.Ref.Name)
		tw.Append("Repository", repoInfo.Name)

		if repoInfo.IsRemote() {
			tw.Append("URL", repoInfo.URL)
			if repoInfo.Revision != "" {
				tw.Append("Revision", repoInfo.Revision)
			}
			tw.Append("Local Path", path)
		} else {
			tw.Append("Path", path)
		}

		description := strings.TrimSpace(skeleton.Description)

		if description != "" {
			tw.Append("Description", description)
		}

		if skeleton.Parent != nil {
			parentPath := homedir.MustCollapse(skeleton.Parent.Ref.Path)

			tw.Append("Parent", parentPath)
		}

		if len(skeleton.Values) > 0 {
			var buf bytes.Buffer

			err := cmdutil.RenderYAML(&buf, skeleton.Values)
			if err != nil {
				return err
			}

			tw.Append("Values", color.BlueString(buf.String()))
		}

		tree := filetree.Build(skeleton)

		tw.Append("Files", tree.Print())

		tw.Render()

		return nil
	}
}
