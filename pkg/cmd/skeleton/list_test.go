package skeleton

import (
	"bytes"
	"testing"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd_Execute(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewListCmd(streams)
	cmd.SetArgs([]string{"--config", "../../config/testdata/empty-config.yaml", "--repositories", "default=../../skeleton/testdata/local-dir"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := streams.Out.(*bytes.Buffer).String()

	expected := `a-skeleton`

	assert.Contains(t, output, expected)
}
