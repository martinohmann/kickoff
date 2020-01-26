package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeValues(t *testing.T) {
	a := Values{
		"foo":      "bar",
		"somebool": true,
	}

	b := Values{
		"somebool": false,
		"nested": map[string]interface{}{
			"bar": "baz",
		},
	}

	expected := Values{
		"foo":      "bar",
		"somebool": false,
		"nested": map[string]interface{}{
			"bar": "baz",
		},
	}

	merged, err := MergeValues(a, b)
	require.NoError(t, err)
	assert.Equal(t, expected, merged)

	// immutability
	assert.Equal(t, true, a["somebool"])
}
