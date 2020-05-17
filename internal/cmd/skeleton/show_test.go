package skeleton

import (
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd_Execute_NonexistantRepository(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, _, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"myskeleton",
		"--config", configFile.Name(),
		"--repository", "nonexistent",
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
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"minimal",
		"--config", configFile.Name(),
	})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	assert.Regexp(t, `Name\s+minimal`, output)
	assert.Regexp(t, `Values\s+foo: bar`, output)
}

func TestShowCmd_Execute_YAMLOutput(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"minimal",
		"--config", configFile.Name(),
		"--output", "yaml",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	expected := `values:
  foo: bar`

	assert.Contains(t, output, expected)
}

func TestShowCmd_Execute_JSONOutput(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{
		"minimal",
		"--config", configFile.Name(),
		"--output", "json",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	expected := `"values": {
    "foo": "bar"
  }`

	assert.Contains(t, output, expected)
}
