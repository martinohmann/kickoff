package cmd

import (
	"github.com/spf13/cobra"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmd/license"
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
