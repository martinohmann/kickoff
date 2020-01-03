package cmd

import (
	"fmt"

	"github.com/martinohmann/skeleton-go/pkg/license"
	"github.com/spf13/cobra"
)

func NewLicensesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "licenses",
		Short: "List available licenses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			licenses, err := license.List()
			if err != nil {
				return err
			}

			fmt.Printf("%-20s NAME\n", "KEY")
			for _, license := range licenses {
				fmt.Printf("%-20s %s\n", license.Key, license.Name)
			}

			return nil
		},
	}

	return cmd
}
