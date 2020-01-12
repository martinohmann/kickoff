package skeleton

import (
	"bytes"
	"testing"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd_Execute_NonexistantRepository(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"myskeleton", "--repositories", "default=nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute_InvalidOutput(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"myskeleton", "--output", "enterprise-xml"})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"a-skeleton", "--repositories", "default=../../skeleton/testdata/local-dir"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()

	expected := `values:
  foo: bar`

	assert.Contains(t, output, expected)
}
