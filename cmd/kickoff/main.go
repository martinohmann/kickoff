package main

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmd"
)

func main() {
	log.SetHandler(cli.Default)

	rootCmd := cmd.NewRootCmd()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
