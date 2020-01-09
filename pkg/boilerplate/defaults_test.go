package boilerplate

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/martinohmann/kickoff/pkg/kickoff"
	"github.com/martinohmann/kickoff/pkg/skeleton"
)

func TestDefaultConfigBytes_EnsureValidYAML(t *testing.T) {
	var config kickoff.Config

	err := yaml.Unmarshal(DefaultConfigBytes(), &config)
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}
}

func TestDefaultSkeletonConfigBytes_EnsureValidYAML(t *testing.T) {
	var config skeleton.Config

	err := yaml.Unmarshal(DefaultSkeletonConfigBytes(), &config)
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}
}
