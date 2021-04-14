package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/spf13/cobra"
)

// NewShowFileCmd creates a command for inspecting project skeleton files.
func NewShowFileCmd(streams cli.IOStreams) *cobra.Command {
	o := &ShowFileOptions{
		IOStreams:   streams,
		TimeoutFlag: cmdutil.NewDefaultTimeoutFlag(),
	}

	cmd := &cobra.Command{
		Use:     "show-file <skeleton> <filename>",
		Aliases: []string{"file", "sf"},
		Short:   "Show skeleton file content",
		Long: cmdutil.LongDesc(`
			Show the content of a skeleton file.`),
		Example: cmdutil.Examples(`
			# Show the content of a skeleton file in a specific repository
			kickoff skeleton show-file myrepo:myskeleton relpath/to/file`),
		Args: cmdutil.ExactNonEmptyArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(args); err != nil {
				return err
			}

			return o.Run()
		},
	}

	o.ConfigFlags.AddFlags(cmd)
	o.TimeoutFlag.AddFlag(cmd)

	return cmd
}

// ShowFileOptions holds options for the show command.
type ShowFileOptions struct {
	cli.IOStreams
	cmdutil.ConfigFlags
	cmdutil.TimeoutFlag

	SkeletonName string
	FilePath     string
}

// Complete completes the show options.
func (o *ShowFileOptions) Complete(args []string) error {
	o.SkeletonName = args[0]
	o.FilePath = filepath.Clean(args[1])

	return o.ConfigFlags.Complete()
}

// Run prints information about a project skeleton in the output format
// specified by the user.
func (o *ShowFileOptions) Run() error {
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

	file, err := findFile(skeleton.Files, o.FilePath)
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

	fmt.Fprintln(o.Out, string(buf))

	return nil
}

func findFile(files []kickoff.File, path string) (kickoff.File, error) {
	for _, file := range files {
		if file.Path() == path {
			return file, nil
		}
	}

	return nil, os.ErrNotExist
}
