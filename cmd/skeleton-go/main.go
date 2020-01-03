package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/martinohmann/skeleton-go/pkg/cmd"
)

func main() {
	logHandler := cli.New(os.Stdout)
	logHandler.Padding = 0

	log.SetHandler(logHandler)

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
