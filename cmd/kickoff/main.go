package main

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmd"
)

func main() {
	log.SetHandler(cli.Default)

	rootCmd := cmd.NewRootCmd()

	rootCmd.AddCommand(cmd.NewCreateCmd())
	rootCmd.AddCommand(cmd.NewLicenseCmd())
	rootCmd.AddCommand(cmd.NewLicensesCmd())
	rootCmd.AddCommand(cmd.NewVersionCmd())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
