package kickoff

import "testing"

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
