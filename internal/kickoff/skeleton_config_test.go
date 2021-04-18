package kickoff

import (
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
	}

	runLoadConfigTests(t, testCases, func(path string) (interface{}, error) {
		return LoadSkeletonConfig(path)
	})
}
