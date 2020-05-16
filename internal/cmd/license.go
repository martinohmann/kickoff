package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/license"
	"github.com/spf13/cobra"
)

// NewLicenseCmd creates a new command which provides subcommands for
// inspecting open source licenses provided by the GitHub Licenses API.
func NewLicenseCmd(streams cli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "license",
		Aliases: []string{"lic", "licenses"},
		Short:   "Inspect open source licenses",
	}

	cmd.AddCommand(license.NewListCmd(streams))
	cmd.AddCommand(license.NewShowCmd(streams))

	return cmd
}
