package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ghodss/yaml"
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
		Use:   "show <name> [<filepath>]",
		Short: "Show config and files for a skeleton",
		Long: cmdutil.LongDesc(`
			Show config and files for a single skeleton.`),
		Example: cmdutil.Examples(`
			# Show skeleton config
			kickoff skeleton show myskeleton

			# Show skeleton config in a specific repository
			kickoff skeleton show myrepo:myskeleton

			# Show the contents of a skeleton file
			kickoff skeleton show myrepo:myskeleton relpath/to/file

			# Show skeleton config using different output
			kickoff skeleton show myskeleton --output json`),
		Args: cobra.RangeArgs(1, 2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return cmdutil.SkeletonNames(f, o.RepoNames...), cobra.ShellCompDirectiveDefault
			case 1:
				return cmdutil.SkeletonFilenames(f, args[0], o.RepoNames...), cobra.ShellCompDirectiveDefault
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			o.SkeletonName = args[0]

			if len(args) > 1 {
				o.FilePath = filepath.Clean(args[1])
			}

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

	FilePath     string
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

	if o.FilePath != "" {
		return o.showSkeletonFile(skeleton, o.FilePath)
	}

	return o.showSkeleton(skeleton)
}

func (o *ShowOptions) showSkeleton(skeleton *kickoff.Skeleton) error {
	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, skeleton)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, skeleton)
	default:
		tw := cli.NewTableWriter(o.Out)
		tw.Append(bold.Sprint("Repository"), skeleton.Ref.Repo.Name)
		tw.Append(bold.Sprint("Name"), skeleton.Ref.Name)
		tw.Append(bold.Sprint("Path"), homedir.Collapse(skeleton.Ref.Path))
		tw.Render()

		fmt.Fprintln(o.Out)

		description := strings.TrimSpace(skeleton.Description)

		if description != "" {
			fmt.Fprintln(o.Out, bold.Sprint("Description"))
			fmt.Fprintln(o.Out, description)
			fmt.Fprintln(o.Out)
		}

		tree := filetree.Build(skeleton)

		buf, err := yaml.Marshal(skeleton.Values)
		if err != nil {
			return err
		}

		tw = cli.NewTableWriter(o.Out)
		tw.SetHeader("Files", "Values")
		tw.Append(tree.Print(), string(buf))
		tw.Render()

		return nil
	}
}

func (o *ShowOptions) showSkeletonFile(skeleton *kickoff.Skeleton, path string) error {
	file, err := findFile(skeleton.Files, path)
	if err != nil {
		return err
	}

	if file.Mode.IsDir() {
		return fmt.Errorf("%s is a directory", file.RelPath)
	}

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, file)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, file)
	default:
		_, err = o.Out.Write(file.Content)
		return err
	}
}

func findFile(files []*kickoff.BufferedFile, path string) (*kickoff.BufferedFile, error) {
	for _, file := range files {
		if file.RelPath == path {
			return file, nil
		}
	}

	return nil, os.ErrNotExist
}
