package skeleton

import (
	"bytes"
	"strings"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/filetree"
	"github.com/martinohmann/kickoff/internal/homedir"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/spf13/cobra"
)

var bold = color.New(color.Bold)

// NewShowCmd creates a command for inspecting project skeletons.
func NewShowCmd(f *cmdutil.Factory) *cobra.Command {
	o := &ShowOptions{
		IOStreams:  f.IOStreams,
		Repository: f.Repository,
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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return cmdutil.SkeletonNames(f, o.RepoNames...), cobra.ShellCompDirectiveDefault
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SkeletonName = args[0]

			return o.Run()
		},
	}

	cmdutil.AddOutputFlag(cmd, &o.Output, "full", "json", "yaml")
	cmdutil.AddRepositoryFlag(cmd, f, &o.RepoNames)

	return cmd
}

// ShowOptions holds options for the show command.
type ShowOptions struct {
	cli.IOStreams

	Repository func(...string) (kickoff.Repository, error)

	Output       string
	RepoNames    []string
	SkeletonName string
}

// Run prints information about a project skeleton in the output format
// specified by the user.
func (o *ShowOptions) Run() error {
	repo, err := o.Repository(o.RepoNames...)
	if err != nil {
		return err
	}

	skeleton, err := repo.LoadSkeleton(o.SkeletonName)
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

		tw.Append(bold.Sprint("Name"), skeleton.Ref.String())
		tw.Append(bold.Sprint("Path"), homedir.MustCollapse(skeleton.Ref.Path))

		description := strings.TrimSpace(skeleton.Description)

		if description != "" {
			tw.Append(bold.Sprint("Description"), description)
		}

		tree := filetree.Build(skeleton)

		tw.Append(bold.Sprint("Files"), tree.Print())

		if len(skeleton.Values) > 0 {
			var buf bytes.Buffer

			err := cmdutil.RenderYAML(&buf, skeleton.Values)
			if err != nil {
				return err
			}

			tw.Append(bold.Sprint("Values"), color.BlueString(strings.TrimRight(buf.String(), "\n")))
		}

		tw.Render()

		return nil
	}
}
