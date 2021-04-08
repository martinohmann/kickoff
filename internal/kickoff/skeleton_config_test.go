package kickoff

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/template"
)

func TestSkeletonConfig_Validate(t *testing.T) {
	testCases := []validatorTestCase{
		{
			name: "config with empty non-nil Parent is invalid",
			v:    &SkeletonConfig{Parent: &ParentRef{}},
			err:  newParentRefError("SkeletonName must not be empty"),
		},
		{
			name: "parent with SkeletonName is valid",
			v: &SkeletonConfig{
				Parent: &ParentRef{SkeletonName: "foo"},
			},
		},
		{
			name: "parent with invalid repository URL is invalid",
			v: &SkeletonConfig{
				Parent: &ParentRef{
					SkeletonName:  "foo",
					RepositoryURL: "inval\\:",
				},
			},
			err: newRepositoryRefError("invalid RepositoryURL: parse \"inval\\\\:\": first path segment in URL cannot contain colon"),
		},
	}

	runValidatorTests(t, testCases)
}

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
