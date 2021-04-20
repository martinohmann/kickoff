package testutil

import (
	"path/filepath"
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/martinohmann/kickoff/internal/template"
	"github.com/stretchr/testify/require"
)

// ConfigFileBuilder is a utility to build kickoff config files in tests.
type ConfigFileBuilder struct {
	*testing.T
	kickoff.Config
}

// NewConfigFileBuilder creates a new *ConfigFileBuilder.
func NewConfigFileBuilder(t *testing.T) *ConfigFileBuilder {
	return &ConfigFileBuilder{T: t}
}

// WithProjectOwner sets the project.owner config field.
func (b *ConfigFileBuilder) WithProjectOwner(owner string) *ConfigFileBuilder {
	b.Project.Owner = owner
	return b
}

// WithRepository adds a repository with name and url to the config.
func (b *ConfigFileBuilder) WithRepository(name, url string) *ConfigFileBuilder {
	if b.Repositories == nil {
		b.Repositories = make(map[string]string)
	}

	b.Repositories[name] = url

	return b
}

// WithValues sets the values in the config.
func (b *ConfigFileBuilder) WithValues(values template.Values) *ConfigFileBuilder {
	b.Values = values
	return b
}

// Create creates the config file in a temp dir that gets cleaned up after the
// current test.
func (b *ConfigFileBuilder) Create() string {
	path := filepath.Join(b.TempDir(), "config.yaml")
	require.NoError(b, kickoff.SaveConfig(path, &b.Config))
	return path
}
