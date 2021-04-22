package filetree

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	s := &kickoff.Skeleton{
		Ref: &kickoff.SkeletonRef{Name: "my/skeleton"},
		Files: []*kickoff.BufferedFile{
			{RelPath: ".kickoff.yaml"},
			{RelPath: "README.md.skel"},
			{RelPath: "foo/bar"},
			{RelPath: "foo/sometemplate.skel"},
			{RelPath: "foo/{{.Values.filename}}/qux"},
		},
	}

	tree := Build(s)

	expected := `my/skeleton
├── .kickoff.yaml
├── README.md.skel
└── foo/
    ├── bar
    ├── sometemplate.skel
    └── {{.Values.filename}}/
        └── qux`

	assert.Equal(t, expected, tree.Print())
}
