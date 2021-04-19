package cmdutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer

	v := map[string]string{"foo": "bar"}

	require.NoError(t, RenderJSON(&buf, v))
	assert.Equal(t, "{\n  \"foo\": \"bar\"\n}", buf.String())
}

func TestRenderYAML(t *testing.T) {
	var buf bytes.Buffer

	v := map[string]string{"foo": "bar"}

	require.NoError(t, RenderYAML(&buf, v))
	assert.Equal(t, "foo: bar\n", buf.String())
}
