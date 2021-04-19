package cmd

import (
	"github.com/martinohmann/kickoff/internal/cmd/license"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/spf13/cobra"
)

// NewLicenseCmd creates a new command which provides subcommands for
// inspecting open source licenses provided by the GitHub Licenses API.
func NewLicenseCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "license",
		Aliases: []string{"lic", "licenses"},
		Short:   "Inspect open source licenses",
	}

	cmd.AddCommand(license.NewListCmd(f))
	cmd.AddCommand(license.NewShowCmd(f))

	return cmd
}
