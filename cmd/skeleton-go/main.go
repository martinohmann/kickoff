package main

import (
	"github.com/apex/log"
	"github.com/martinohmann/skeleton-go/pkg/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	rootCmd.AddCommand(cmd.NewCreateCmd())
	rootCmd.AddCommand(cmd.NewLicensesCmd())
	rootCmd.AddCommand(cmd.NewVersionCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
