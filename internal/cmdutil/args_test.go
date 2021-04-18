package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExactNonEmptyArgs(t *testing.T) {
	fn := ExactNonEmptyArgs(2)

	assert.Error(t, fn(nil, nil))
	assert.Error(t, fn(nil, []string{"one"}))
	assert.NoError(t, fn(nil, []string{"one", "two"}))
	assert.Error(t, fn(nil, []string{"", "two"}))
	assert.Error(t, fn(nil, []string{"one", ""}))
	assert.Error(t, fn(nil, []string{"one", "two", "three"}))
}
