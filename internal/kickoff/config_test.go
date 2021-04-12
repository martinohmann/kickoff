package kickoff

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitconfig "github.com/tcnksm/go-gitconfig"
)

func TestConfig_ApplyDefaults(t *testing.T) {
	config := Config{
		Project: ProjectConfig{
			Owner:     "johndoe",
			License:   NoLicense,
			Gitignore: NoGitignore,
		},
	}

	config.ApplyDefaults()

	expected := Config{
		Project: ProjectConfig{
			Host:  DefaultProjectHost,
			Owner: "johndoe",
		},
		Repositories: map[string]string{
			DefaultRepositoryName: DefaultRepositoryURL,
		},
		Values: template.Values{},
	}

	assert.Equal(t, expected, config)
}

func TestConfig_MergeFromFile(t *testing.T) {
	config := Config{
		Project: ProjectConfig{
			Host: DefaultProjectHost,
		},
	}

	err := config.MergeFromFile("../testdata/config/config.yaml")
	require.NoError(t, err)

	expected := Config{
		Project: ProjectConfig{
			Host:  DefaultProjectHost,
			Owner: "johndoe",
		},
		Repositories: map[string]string{
			"local":  "/some/local/path",
			"remote": "https://git.john.doe/johndoe/remote-repo",
		},
		Values: template.Values{
			"foo": "bar",
		},
	}

	assert.Equal(t, expected, config)
}

type validatorTestCase struct {
	name string
	v    Validator
	err  error
}

func runValidatorTests(t *testing.T, testCases []validatorTestCase) {
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.v.Validate()
			if tc.err == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []validatorTestCase{
		{
			name: "config with defaults is valid",
			v: func() *Config {
				c := Config{}
				c.ApplyDefaults()
				return &c
			}(),
		},
		{
			name: "config with invalid defaults",
			v:    &Config{Project: ProjectConfig{Host: "inval\\:"}},
			err:  newProjectConfigError(`invalid Host: parse "inval\\:": first path segment in URL cannot contain colon`),
		},
		{
			name: "config with empty repository name",
			v: &Config{
				Repositories: map[string]string{"": "/tmp/foo"},
			},
			err: newRepositoryRefError(`repository name must not be empty`),
		},
		{
			name: "config with empty repository URL",
			v: &Config{
				Repositories: map[string]string{"foo": ""},
			},
			err: newRepositoryRefError(`repository URL must not be empty`),
		},
		{
			name: "config with invalid repository url",
			v: &Config{
				Repositories: map[string]string{"default": "inval\\:"},
			},
			err: newRepositoryRefError(`invalid URL: parse "inval\\:": first path segment in URL cannot contain colon`),
		},
		{
			name: "config with repository ref",
			v: &Config{
				Repositories: map[string]string{"default": "/tmp/foo"},
			},
		},
	}

	runValidatorTests(t, testCases)
}

func TestDetectDefaultProjectOwner(t *testing.T) {
	restore := mockGitconfig(nil)
	defer restore()

	assert.Empty(t, detectDefaultProjectOwner())

	restore = mockGitconfig(map[string]string{
		"user.name": "johndoe",
	})
	defer restore()

	assert.Equal(t, "johndoe", detectDefaultProjectOwner())
}

func mockGitconfig(config map[string]string) (restore func()) {
	gitconfigFn = func(key string) (string, error) {
		v, ok := config[key]
		if !ok {
			return "", errors.New("not found")
		}

		return v, nil
	}

	return func() { gitconfigFn = gitconfig.Global }
}

type loadConfigTestCase struct {
	name     string
	path     string
	expected interface{}
	err      error
}

func runLoadConfigTests(t *testing.T, testCases []loadConfigTestCase, fn func(string) (interface{}, error)) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := fn(tc.path)
			if tc.err == nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, config)
			} else {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	restore := mockGitconfig(map[string]string{
		"user.name": "johndoe",
	})
	defer restore()

	testCases := []loadConfigTestCase{
		{
			name: "empty",
			path: "../testdata/config/empty-config.yaml",
			expected: &Config{
				Project: ProjectConfig{
					Host:  "github.com",
					Owner: "johndoe",
				},
				Repositories: map[string]string{
					"default": "https://github.com/martinohmann/kickoff-skeletons",
				},
				Values: template.Values{},
			},
		},
		{
			name: "simple",
			path: "../testdata/config/config.yaml",
			expected: &Config{
				Project: ProjectConfig{
					Host:  "github.com",
					Owner: "johndoe",
				},
				Repositories: map[string]string{
					"local":  "/some/local/path",
					"remote": "https://git.john.doe/johndoe/remote-repo",
				},
				Values: template.Values{
					"foo": "bar",
				},
			},
		},
	}

	runLoadConfigTests(t, testCases, func(path string) (interface{}, error) {
		return LoadConfig(path)
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves config", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		require.NoError(t, SaveConfig(path, &Config{}))
		require.FileExists(t, path)
	})

	t.Run("creates nonexistent dirs", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nonexistent-dir", "config.yaml")
		require.NoError(t, SaveConfig(path, &Config{}))
		require.FileExists(t, path)
	})

	t.Run("returns error if creating parent dir fails", func(t *testing.T) {
		tmpdir := t.TempDir()
		f, err := os.Create(filepath.Join(tmpdir, "actually-a-file"))
		require.NoError(t, err)
		require.NoError(t, f.Close())

		path := filepath.Join(tmpdir, "actually-a-file", "config.yaml")
		require.Error(t, SaveConfig(path, &Config{}))
		require.NoFileExists(t, path)
	})

	t.Run("does not save invalid config", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.yaml")
		require.Error(t, SaveConfig(path, &Config{
			Project: ProjectConfig{Host: "invalid\\:"},
		}))
		require.NoFileExists(t, path)
	})
}
