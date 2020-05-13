package cmd

import (
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd/license"
	"github.com/spf13/cobra"
)

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
