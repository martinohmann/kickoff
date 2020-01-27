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
			Email: "john@example.com",
		},
	}

	config.ApplyDefaults()

	expected := Config{
		Project: Project{
			Host:      DefaultProjectHost,
			Owner:     "johndoe",
			Email:     "john@example.com",
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
			Email: "johndoe@example.com",
		},
	}

	err := config.MergeFromFile("../testdata/config/config.yaml")
	require.NoError(t, err)

	expected := Config{
		Project: Project{
			Owner: "johndoe",
			Email: "johndoe@example.com",
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

func TestProject_Author(t *testing.T) {
	p1 := Project{Owner: "johndoe", Email: "john@example.com"}
	assert.Equal(t, "johndoe <john@example.com>", p1.Author())

	p2 := Project{Owner: "janedoe"}
	assert.Equal(t, "janedoe", p2.Author())
}
