package skeleton

import (
	"os"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCmd_Execute(t *testing.T) {
	configFile := testutil.NewConfigFileBuilder(t).
		WithRepository("default", "../../testdata/repos/repo1").
		Create()
	defer os.Remove(configFile.Name())

	streams, _, out, _ := cli.NewTestIOStreams()
	cmd := NewListCmd(streams)
	cmd.SetArgs([]string{"--config", configFile.Name()})

	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), `minimal`)
}
