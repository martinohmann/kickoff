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

	c := Values{
		"nested": map[string]interface{}{
			"bar": "foo",
		},
	}

	expected := Values{
		"foo":      "bar",
		"somebool": false,
		"nested": map[string]interface{}{
			"bar": "foo",
		},
	}

	merged, err := MergeValues(a, b, c)
	require.NoError(t, err)
	assert.Equal(t, expected, merged)

	// immutability
	assert.Equal(t, true, a["somebool"])
}

func TestLoadValues(t *testing.T) {
	values, err := LoadValues("../testdata/values/values.yaml")
	require.NoError(t, err)

	expected := Values{
		"foo": "bar",
		"baz": "qux",
	}

	assert.Equal(t, expected, values)

	_, err = LoadValues("../testdata/values/invalid.yaml")
	require.Error(t, err)

	_, err = LoadValues("../testdata/values/nonexistent.yaml")
	require.Error(t, err)
}
