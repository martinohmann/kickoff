package boilerplate

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/config"
)

func TestDefaultConfigBytes_EnsureValidYAML(t *testing.T) {
	var cfg config.Config

	err := yaml.Unmarshal(DefaultConfigBytes(), &cfg)
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}
}

func TestDefaultSkeletonConfigBytes_EnsureValidYAML(t *testing.T) {
	// @FIXME: probably get rid of the whole package as it causes dependecy
	// cycles.
	var cfg map[string]interface{}

	err := yaml.Unmarshal(DefaultSkeletonConfigBytes(), &cfg)
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}
}
