package cmd

import (
	"fmt"

	"github.com/martinohmann/skeleton-go/pkg/license"
	"github.com/spf13/cobra"
)

func NewLicenseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license <key>",
		Short: "Fetch a license text",
		Long:  "Fetches a license text via the GitHub Licenses API (https://developer.github.com/v3/licenses/#get-an-individual-license).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			license, err := license.Get(args[0])
			if err != nil {
				return err
			}

			fmt.Println(license.Body)

			return nil
		},
	}

	return cmd
}
