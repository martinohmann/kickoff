package skeleton

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func (o *ShowOptions) showSkeletonFile(skeleton *kickoff.Skeleton, path string) error {
	file, err := findFile(skeleton.Files, path)
	if err != nil {
		return err
	}

	if file.Mode().IsDir() {
		return fmt.Errorf("%q is a directory", file.Path())
	}

	r, err := file.Reader()
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f := kickoff.NewBufferedFile(file.Path(), buf, file.Mode())

	switch o.Output {
	case "json":
		return cmdutil.RenderJSON(o.Out, f)
	case "yaml":
		return cmdutil.RenderYAML(o.Out, f)
	default:
		fmt.Fprintln(o.Out, string(buf))
		return nil
	}
}

func findFile(files []kickoff.File, path string) (kickoff.File, error) {
	for _, file := range files {
		if file.Path() == path {
			return file, nil
		}
	}

	return nil, os.ErrNotExist
}
