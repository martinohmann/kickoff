package kickoff

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
