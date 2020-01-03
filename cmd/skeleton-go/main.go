package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "skeleton-go",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
}

func main() {
	logHandler := cli.New(os.Stdout)
	logHandler.Padding = 0

	log.SetHandler(logHandler)

	rootCmd := newRootCommand()

	rootCmd.AddCommand(newCreateCommand())

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
