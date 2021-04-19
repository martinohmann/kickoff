package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddOutputFlag(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var output string

		cmd := &cobra.Command{}
		AddOutputFlag(cmd, &output, "json", "yaml")
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "json", output)
	})

	t.Run("valid output", func(t *testing.T) {
		var output string

		cmd := &cobra.Command{}
		AddOutputFlag(cmd, &output, "json", "yaml")
		cmd.SetArgs([]string{"--output", "yaml"})
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "yaml", output)
	})

	t.Run("invalid output", func(t *testing.T) {
		var output string

		cmd := &cobra.Command{}
		AddOutputFlag(cmd, &output, "json", "yaml")
		cmd.SetArgs([]string{"--output", "xml"})
		require.Error(t, cmd.Execute())

		assert.Equal(t, "json", output)
	})

	t.Run("panic on empty list of allowed values", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic but got none")
			}
		}()

		var output string

		cmd := &cobra.Command{}
		AddOutputFlag(cmd, &output)
	})
}
