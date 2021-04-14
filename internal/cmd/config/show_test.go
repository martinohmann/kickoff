package config

import (
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd_Execute_NonexistentConfig(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", "nonexistent"})

	assert.Error(t, cmd.Execute())
}

func TestShowCmd_Execute_InvalidOutput(t *testing.T) {
	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--output", "enterprise-xml"})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestShowCmd_Execute(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithValues(template.Values{"foo": "bar"}).
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", configFile.Name()})

	require.NoError(t, cmd.Execute())

	expected := `values:
  foo: bar`

	assert.Contains(t, out.String(), expected)
}
