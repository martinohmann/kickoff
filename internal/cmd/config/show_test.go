package config

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd_Execute_NonexistentConfig(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", "nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute_InvalidOutput(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--output", "enterprise-xml"})

	err := cmd.Execute()
	if err != cmdutil.ErrInvalidOutputFormat {
		t.Fatalf("expected error %v, got %v", cmdutil.ErrInvalidOutputFormat, err)
	}
}

func TestShowCmd_Execute(t *testing.T) {
	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", "../../testdata/config/values-config.yaml"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()

	expected := `values:
  foo: bar`

	assert.Contains(t, output, expected)
}
