package skeleton

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"kickoff.run/pkg/cli"
)

func TestListCmd_Execute(t *testing.T) {
	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewListCmd(streams)
	cmd.SetArgs([]string{
		"--config", "../../testdata/config/empty-config.yaml",
		"--repositories", "default=../../testdata/repos/repo1",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	expected := `minimal`

	assert.Contains(t, output, expected)
}
