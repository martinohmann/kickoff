package config

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/martinohmann/kickoff/pkg/cli"
	"github.com/stretchr/testify/assert"
)

func TestShowCmd_Execute_NonexistantConfig(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", "nonexistent"})

	expectedErr := errors.New(`file "nonexistent" does not exist`)

	err := cmd.Execute()
	if !reflect.DeepEqual(expectedErr, err) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestShowCmd_Execute_InvalidOutput(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--output", "enterprise-xml"})

	err := cmd.Execute()
	if err != ErrInvalidOutputFormat {
		t.Fatalf("expected error %v, got %v", ErrInvalidOutputFormat, err)
	}
}

func TestShowCmd_Execute(t *testing.T) {
	streams := cli.NewTestIOStreams()
	cmd := NewShowCmd(streams)
	cmd.SetArgs([]string{"--config", "testdata/config.yaml"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := streams.Out.(*bytes.Buffer).String()

	expected := `values:
  foo: bar`

	assert.Contains(t, output, expected)
}
