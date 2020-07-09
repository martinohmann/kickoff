package template

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderReader(t *testing.T) {
	testCases := []struct {
		name        string
		r           io.Reader
		values      Values
		expected    string
		expectedErr error
	}{
		{
			name: "go package name",
			r:    strings.NewReader("package {{.path|goPackageName}}"),
			values: Values{
				"path": "github.com/foo/Bar-baz_v1\n",
			},
			expected: `package barbazv1`,
		},
		{
			name:        "invalid template",
			r:           strings.NewReader("package {{invalid"),
			expectedErr: errors.New(`failed to prepare template: template: :1: function "invalid" not defined`),
		},
		{
			name:        "errors on missing template values",
			r:           strings.NewReader("{{.missing}}"),
			expectedErr: errors.New(`failed to render template: template: :1:2: executing "" at <.missing>: map has no entry for key "missing"`),
		},
		{
			name:        "forwards reader errors",
			r:           badReader(0),
			expectedErr: errors.New(`bad reader`),
		},
		{
			name: "has template funcs",
			r:    strings.NewReader("{{ toYAML . }}"),
			values: Values{
				"key1": "one",
				"key2": 2,
			},
			expected: "key1: one\nkey2: 2\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rendered, err := RenderReader(tc.r, tc.values)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, rendered)
			}
		})
	}
}

type badReader int

func (badReader) Read(_ []byte) (int, error) {
	return 0, errors.New("bad reader")
}
