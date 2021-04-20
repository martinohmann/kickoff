package cmdutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/cli"
	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		configPath := testutil.NewConfigFileBuilder(t).
			WithRepository("default", "../testdata/repos/repo1").
			WithRepository("other", "../testdata/repos/repo2").
			Create()

		streams, _, _, _ := cli.NewTestIOStreams()

		f := NewFactoryWithConfigPath(streams, configPath)

		assert.Equal(t, configPath, f.ConfigPath)

		config, err := f.Config()
		require.NoError(t, err)

		assert.Equal(t, "../testdata/repos/repo1", config.Repositories["default"])
		assert.Equal(t, "../testdata/repos/repo2", config.Repositories["other"])

		repo1, err := f.Repository()
		require.NoError(t, err)

		refs, err := repo1.ListSkeletons()
		require.NoError(t, err)
		assert.Len(t, refs, 3)

		repo2, err := f.Repository("default")
		require.NoError(t, err)

		refs, err = repo2.ListSkeletons()
		require.NoError(t, err)
		assert.Len(t, refs, 2)

		_, err = f.Repository("not-configured")
		require.EqualError(t, err, `repository "not-configured" not configured`)
	})

	t.Run("with non-existent config", func(t *testing.T) {
		nonexistent := filepath.Join(t.TempDir(), "nonexistent")

		streams, _, _, _ := cli.NewTestIOStreams()

		f := NewFactoryWithConfigPath(streams, nonexistent)

		_, err := f.Config()
		require.Error(t, err)

		_, err = f.Repository()
		require.Error(t, err)
	})

	t.Run("missing default config is allowed", func(t *testing.T) {
		nonexistent := filepath.Join(t.TempDir(), "nonexistent")

		defer testutil.MockDefaultConfigPath(nonexistent)()

		streams, _, _, _ := cli.NewTestIOStreams()

		f := NewFactory(streams)

		_, err := f.Config()
		require.NoError(t, err)
	})
}

func TestGetConfigPath(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		oldEnv, ok := os.LookupEnv("KICKOFF_CONFIG")
		if ok {
			defer func() { os.Setenv("KICKOFF_CONFIG", oldEnv) }()
		}

		os.Setenv("KICKOFF_CONFIG", "/config/from/env/config.yaml")

		assert.Equal(t, "/config/from/env/config.yaml", getConfigPath())
	})

	t.Run("default config if env empty or unset", func(t *testing.T) {
		oldEnv, ok := os.LookupEnv("KICKOFF_CONFIG")
		if ok {
			defer func() { os.Setenv("KICKOFF_CONFIG", oldEnv) }()
			os.Unsetenv("KICKOFF_CONFIG")
		}

		assert.Equal(t, kickoff.DefaultConfigPath, getConfigPath())
	})
}
