package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLongDesc(t *testing.T) {
	assert.Equal(t, "foo\nbar", LongDesc("\n  foo\n  bar"))
	assert.Equal(t, "foo\nbar", LongDesc("  foo\n  bar"))
	assert.Equal(t, "foo\nbar", LongDesc("foo\nbar"))
}

func TestExamples(t *testing.T) {
	assert.Equal(t, "  foo", Examples("foo"))
	assert.Equal(t, "  foo", Examples("    foo"))
	assert.Equal(t, "  foo\n  bar", Examples("foo\nbar"))
	assert.Equal(t, "  foo\n  bar", Examples("foo\n  bar"))
}
