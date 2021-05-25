package kickoff

import (
	"errors"
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
)

func TestLoadSkeletonConfig(t *testing.T) {
	testCases := []loadConfigTestCase{
		{
			name: "minimal config with values",
			path: "../testdata/repos/repo2/skeletons/minimal/.kickoff.yaml",
			expected: &SkeletonConfig{
				Values: template.Values{"foo": "bar"},
			},
		},
		{
			name:     "empty config",
			path:     "../testdata/repos/repo3/skeletons/simple/.kickoff.yaml",
			expected: &SkeletonConfig{},
		},
		{
			name: "validates config on load",
			path: "../testdata/config/skeleton-with-invalid-params.yaml",
			err:  errors.New(`parameter "foo": invalid parameter spec: invalid parameter type "invalid", allowed values: string, number, bool, list<string>, list<number>`),
		},
	}

	runLoadConfigTests(t, testCases, func(path string) (interface{}, error) {
		return LoadSkeletonConfig(path)
	})
}
