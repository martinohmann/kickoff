package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmd"
)

func main() {
	streams := cli.DefaultIOStreams

	cmd := cmd.NewRootCmd(streams)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(streams.ErrOut, color.RedString("error:"), err)
		os.Exit(1)
	}
}
