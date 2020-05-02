package main

import (
	"github.com/apex/log"
	"kickoff.run/pkg/cli"
	"kickoff.run/pkg/cmd"
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
