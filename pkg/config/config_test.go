package config

import (
	"testing"

	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ApplyDefaults(t *testing.T) {
	config := Config{
		Project: Project{
			Owner: "johndoe",
		},
	}

	config.ApplyDefaults()

	expected := Config{
		Project: Project{
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
		Project: Project{
			Host: DefaultProjectHost,
		},
	}

	err := config.MergeFromFile("../testdata/config/config.yaml")
	require.NoError(t, err)

	expected := Config{
		Project: Project{
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

func TestProject_GoPackagePath(t *testing.T) {
	p := Project{Owner: "foo", Name: "bar", Host: "github.com"}
	assert.Equal(t, "github.com/foo/bar", p.GoPackagePath())
}

func TestProject_URL(t *testing.T) {
	p := Project{Owner: "foo", Name: "bar", Host: "github.com"}
	assert.Equal(t, "https://github.com/foo/bar", p.URL())
}
