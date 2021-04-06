package cmd

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/repository"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for kickoff.
func NewRootCmd(streams cli.IOStreams) *cobra.Command {
	logLevel := log.WarnLevel.String()

	cmd := &cobra.Command{
		Use:           "kickoff",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			err := configureLogger(streams.ErrOut, logLevel)
			if err != nil {
				return err
			}

			// We silence usage output here instead of doing so while
			// initializing the struct above because we want to print the usage
			// if the user actually misused the CLI (e.g. missing arguments,
			// wrong flags), but we do not want to show the usage on errors
			// that occurred when the CLI arguments where actually valid (e.g.
			// user queried for a skeleton that does not exist). Since
			// PersistentPreRun is called after argument parsing happened, we
			// silence the usage here for all subsequent errors.
			//
			// Also see the following issue:
			// https://github.com/spf13/cobra/issues/340#issuecomment-378726225
			cmd.SilenceUsage = true

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", logLevel, "Level for stderr log output")

	cmd.AddCommand(NewCacheCmd(streams))
	cmd.AddCommand(NewCompletionCmd(streams))
	cmd.AddCommand(NewConfigCmd(streams))
	cmd.AddCommand(NewGitignoreCmd(streams))
	cmd.AddCommand(NewInitCmd(streams))
	cmd.AddCommand(NewLicenseCmd(streams))
	cmd.AddCommand(NewProjectCmd(streams))
	cmd.AddCommand(NewRepositoryCmd(streams))
	cmd.AddCommand(NewSkeletonCmd(streams))
	cmd.AddCommand(NewVersionCmd(streams))

	return cmd
}

func Execute() {
	streams := cli.DefaultIOStreams

	cmd := NewRootCmd(streams)

	if err := cmd.Execute(); err != nil {
		handleError(streams.ErrOut, err)
		os.Exit(1)
	}
}

func handleError(w io.Writer, err error) {
	fmt.Fprintln(w, color.RedString("error:"), err)

	var (
		contextMsg           string
		netErr               net.Error
		skeletonNotFoundErr  repository.SkeletonNotFoundError
		repoNotConfiguredErr cmdutil.RepositoryNotConfiguredError
	)

	switch {
	case errors.Is(err, gitignore.ErrNotFound):
		contextMsg = "Run `kickoff gitignore list` to get a list of available templates."
	case errors.Is(err, license.ErrNotFound):
		contextMsg = "Run `kickoff licenses list` to get a list of available licenses."
	case errors.As(err, &skeletonNotFoundErr):
		contextMsg = "Run `kickoff skeleton list` to get a list of available skeletons."
	case errors.As(err, &repoNotConfiguredErr):
		contextMsg = "Run `kickoff repository list` to get a list of available repositories."
	case errors.As(err, &netErr):
		if netErr.Temporary() {
			contextMsg = "Temporary network error. Check your internet connection."
		}
	}

	if contextMsg != "" {
		fmt.Fprintf(w, "\n%s\n", contextMsg)
	}
}

func configureLogger(out io.Writer, level string) error {
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("%w: available levels: %v", err, log.AllLevels)
	}

	formatter := &log.TextFormatter{
		DisableTimestamp:       true,
		PadLevelText:           true,
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			const pkgBase = "github.com/martinohmann/kickoff/"

			function := strings.TrimPrefix(f.Function, pkgBase)
			file := fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			return function, file
		},
	}

	if lvl >= log.DebugLevel {
		formatter.DisableTimestamp = false
		log.SetReportCaller(true)
	}

	log.SetLevel(lvl)
	log.SetFormatter(formatter)
	log.SetOutput(out)

	return nil
}
