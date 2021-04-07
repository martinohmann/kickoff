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
			Owner: "johndoe",
		},
	}

	config.ApplyDefaults()

	expected := Config{
		Project: ProjectConfig{
			Host:      DefaultProjectHost,
			Owner:     "johndoe",
			License:   NoLicense,
			Gitignore: NoGitignore,
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
