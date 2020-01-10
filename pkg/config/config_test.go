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
			Author: "John Doe",
			Email:  "john@example.com",
		},
		Git: Git{
			User: "johndoe",
		},
	}

	config.ApplyDefaults("myproject")

	expected := Config{
		License: DefaultLicense,
		Project: Project{
			Name:   "myproject",
			Author: "John Doe",
			Email:  "john@example.com",
		},
		Git: Git{
			Host:     DefaultGitHost,
			User:     "johndoe",
			RepoName: "myproject",
		},
		Skeletons: Skeletons{
			RepositoryURL: DefaultSkeletonRepositoryURL,
		},
	}

	assert.Equal(t, expected, config)
}

func TestConfig_MergeFromFile(t *testing.T) {
	config := Config{
		Project: Project{
			Author: "John Doe",
		},
	}

	err := config.MergeFromFile("testdata/config.yaml")
	require.NoError(t, err)

	expected := Config{
		Project: Project{
			Author: "John Doe",
		},
		Git: Git{
			User: "johndoe",
		},
		Skeletons: Skeletons{
			RepositoryURL: "https://git.john.doe/johndoe/kickoff-skeletons",
		},
		Values: template.Values{
			"foo": "bar",
		},
	}

	assert.Equal(t, expected, config)
}

func TestLoadSkeleton(t *testing.T) {
	config, err := LoadSkeleton("testdata/.kickoff.yaml")
	require.NoError(t, err)

	expected := Skeleton{
		Values: template.Values{
			"bar": "baz",
		},
	}

	assert.Equal(t, expected, config)
}

func TestGit_GoPackagePath(t *testing.T) {
	g := Git{User: "foo", RepoName: "bar", Host: "github.com"}
	assert.Equal(t, "github.com/foo/bar", g.GoPackagePath())
}

func TestGit_URL(t *testing.T) {
	g := Git{User: "foo", RepoName: "bar", Host: "github.com"}
	assert.Equal(t, "https://github.com/foo/bar", g.URL())
}

func TestProject_AuthorString(t *testing.T) {
	p1 := Project{Author: "John Doe", Email: "john@example.com"}
	assert.Equal(t, "John Doe <john@example.com>", p1.AuthorString())

	p2 := Project{Author: "Jane Doe"}
	assert.Equal(t, "Jane Doe", p2.AuthorString())
}
