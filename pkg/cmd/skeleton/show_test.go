package skeleton

import (
	"testing"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd_Execute_NonexistantRepository(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"myskeleton",
		"--config", "../../testdata/config/empty-config.yaml",
		"--repositories", "default=nonexistent",
	})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute_InvalidOutput(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"myskeleton",
		"--config", "../../testdata/config/empty-config.yaml",
		"--output", "enterprise-xml",
	})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute(t *testing.T) {
	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"minimal",
		"--config", "../../testdata/config/empty-config.yaml",
		"--repositories", "default=../../testdata/repos/repo1",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	expected := `values:
  foo: bar`

	assert.Contains(t, output, expected)
}
