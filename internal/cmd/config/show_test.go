package config

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/cmdutil"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCmd(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		configPath := testutil.NewConfigFileBuilder(t).
			WithValues(template.Values{"foo": "bar"}).
			Create()

		streams, _, out, _ := cli.NewTestIOStreams()
		f := cmdutil.NewFactoryWithConfigPath(streams, configPath)

		cmd := NewShowCmd(f)
		cmd.SetOut(io.Discard)

		require.NoError(t, cmd.Execute())

		expected := `values:
  foo: bar`

		assert.Contains(t, out.String(), expected)
	})

	t.Run("nonexistent config", func(t *testing.T) {
		nonexistent := filepath.Join(t.TempDir(), "nonexistent")

		streams, _, _, _ := cli.NewTestIOStreams()
		f := cmdutil.NewFactoryWithConfigPath(streams, nonexistent)

		cmd := NewShowCmd(f)
		cmd.SetOut(io.Discard)

		require.Error(t, cmd.Execute())
	})
}
