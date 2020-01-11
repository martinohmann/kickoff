package cli

import (
	"bytes"
	"io"
	"os"
)

// DefaultIOStreams provides the default streams for os.Stdin, os.Stdout and
// os.Stderr.
var DefaultIOStreams = IOStreams{
	In:     os.Stdin,
	Out:    os.Stdout,
	ErrOut: os.Stderr,
}

// IOStreams is a holder for input and output streams. Commands should use this
// instead of directly relying on os.Stdin, os.Stdout and os.Stderr to make it
// possible to replace the streams in tests.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// NewTestIOStreams provides IOStreams that use a *bytes.Buffer. This can be
// used in tests to make assertions on command output as well as control the
// input stream.
func NewTestIOStreams() IOStreams {
	return IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
}
