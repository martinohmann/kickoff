package cmd

import (
	"github.com/martinohmann/kickoff/pkg/cmd/license"
	"github.com/spf13/cobra"
)

func NewLicenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "license",
		Aliases: []string{"lic", "licenses"},
		Short:   "Manage licenses",
	}

	cmd.AddCommand(license.NewListCmd())
	cmd.AddCommand(license.NewShowCmd())

	return cmd
}
