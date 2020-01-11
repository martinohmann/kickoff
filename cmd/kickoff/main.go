package main

import (
	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/martinohmann/kickoff/pkg/cmd"
)

func main() {
	streams := cli.DefaultIOStreams
	handler := cli.NewLogHandler(streams.ErrOut)

	log.SetHandler(handler)

	rootCmd := cmd.NewRootCmd(streams)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err.Error())
	}
}
