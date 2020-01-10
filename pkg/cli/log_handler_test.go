package cli

import (
	"bytes"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogHandler(t *testing.T) {
	var buf bytes.Buffer

	h := NewLogHandler(&buf)

	e := &log.Entry{
		Level: log.DebugLevel,
		Fields: log.Fields{
			"foo": "bar",
		},
		Message: "hello world",
	}

	err := h.HandleLog(e)
	require.NoError(t, err)

	expected := "• hello world                       foo=bar\n"
	assert.Equal(t, expected, buf.String())
}
