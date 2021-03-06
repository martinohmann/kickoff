package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/license"
	"github.com/martinohmann/kickoff/internal/repository"
	"github.com/martinohmann/kickoff/internal/update"
	"github.com/martinohmann/kickoff/internal/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for kickoff.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	logLevel := os.Getenv(kickoff.EnvKeyLogLevel)
	if logLevel == "" {
		logLevel = log.WarnLevel.String()
	}

	cmd := &cobra.Command{
		Use:           "kickoff",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			err := configureLogger(f.IOStreams.ErrOut, logLevel)
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
	cmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return cmdutil.LogLevelNames(), cobra.ShellCompDirectiveDefault
	})

	cmd.AddCommand(NewCacheCmd(f.IOStreams))
	cmd.AddCommand(NewCompletionCmd(f.IOStreams))
	cmd.AddCommand(NewConfigCmd(f))
	cmd.AddCommand(NewGitignoreCmd(f))
	cmd.AddCommand(NewInitCmd(f))
	cmd.AddCommand(NewLicenseCmd(f))
	cmd.AddCommand(NewProjectCmd(f))
	cmd.AddCommand(NewRepositoryCmd(f))
	cmd.AddCommand(NewSkeletonCmd(f))
	cmd.AddCommand(NewVersionCmd(f.IOStreams))

	return cmd
}

func Execute() {
	streams := cli.DefaultIOStreams
	f := cmdutil.NewFactory(streams)

	updateCh := make(chan *update.Info)
	errCh := make(chan error)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		statePath := filepath.Join(kickoff.LocalCacheDir, "update-state.json")
		current := version.Get()

		info, err := update.Check(ctx, statePath, current.GitVersion, 24*time.Hour)
		if err != nil {
			errCh <- err
		} else {
			updateCh <- info
		}
	}()

	cmd := NewRootCmd(f)

	if err := cmd.Execute(); err != nil {
		handleError(streams.ErrOut, err)
		os.Exit(1)
	}

	select {
	case info := <-updateCh:
		if info != nil && info.IsUpdate {
			fmt.Fprintf(
				streams.ErrOut,
				"\nA new release of kickoff is available: %s → %s\n\n%s\n",
				color.CyanString(info.CurrentVersion),
				color.CyanString(info.LatestVersion),
				info.LatestReleaseURL,
			)
		}
	case err := <-errCh:
		log.WithError(err).Debug("update check failed")
	}
}

func handleError(w io.Writer, err error) {
	fmt.Fprintln(w, color.RedString("error:"), err)

	var (
		errorContext         string
		netErr               net.Error
		gitignoreNotFoundErr gitignore.NotFoundError
		licenseNotFoundErr   license.NotFoundError
		skeletonNotFoundErr  repository.SkeletonNotFoundError
		repoNotConfiguredErr cmdutil.RepositoryNotConfiguredError
		revisionNotFoundErr  repository.RevisionNotFoundError
		invalidRepoErr       repository.InvalidSkeletonRepositoryError
	)

	switch {
	case errors.Is(err, repository.ErrNoRepositories):
		errorContext = "For kickoff to be functional you need to configure at least one skeleton repository.\n\nHere are two ways how to do that:\n\n"
		errorContext += fmt.Sprintf("❯ Interactively: %s\n", bold.Sprint("kickoff init"))
		errorContext += fmt.Sprintf("❯ Manually: %s", bold.Sprint("kickoff repository add <name> <repo-url>"))
	case errors.As(err, &gitignoreNotFoundErr):
		errorContext = fmt.Sprintf("To get a list of available gitignore templates run: %s", bold.Sprint("kickoff gitignore list"))
	case errors.As(err, &licenseNotFoundErr):
		errorContext = fmt.Sprintf("To get a list of available licenses run: %s", bold.Sprint("kickoff license list"))
	case errors.As(err, &skeletonNotFoundErr):
		errorContext = fmt.Sprintf("To get a list of available skeletons run: %s", bold.Sprint("kickoff skeleton list"))
	case errors.As(err, &repoNotConfiguredErr):
		errorContext = fmt.Sprintf("To get a list of available repositories run: %s", bold.Sprint("kickoff repository list"))
	case errors.As(err, &revisionNotFoundErr):
		ref := revisionNotFoundErr.RepoRef

		if ref.Name != "" {
			errorContext = fmt.Sprintf("You may want to re-add the repository with an existing revision:\n\n%s\n%s",
				bold.Sprintf("  kickoff repository remove %s", ref.Name),
				bold.Sprintf("  kickoff repository add %s %s --revision <existing-revision>", ref.Name, ref.URL))
		}
	case errors.As(err, &invalidRepoErr):
		ref := invalidRepoErr.RepoRef

		errorContext = fmt.Sprintf("Ensure that the repository contains a %s subdirectory.", bold.Sprint("skeletons/"))

		if ref.Name != "" {
			errorContext += fmt.Sprintf("\n\nTo remove it run: %s", bold.Sprintf("kickoff repository remove %s", ref.Name))
		}
	case errors.As(err, &netErr):
		if netErr.Temporary() {
			errorContext = "Temporary network error. Check your internet connection."
		}
	}

	if errorContext == "" {
		return
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, errorContext)
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
