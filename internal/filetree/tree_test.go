package filetree

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/kickoff"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	s := &kickoff.Skeleton{
		Ref: &kickoff.SkeletonRef{Name: "my/skeleton"},
		Files: []kickoff.File{
			&kickoff.FileRef{RelPath: ".kickoff.yaml"},
			&kickoff.FileRef{RelPath: "README.md.skel"},
			&kickoff.FileRef{RelPath: "foo/bar"},
			&kickoff.FileRef{RelPath: "foo/sometemplate.skel"},
			&kickoff.FileRef{RelPath: "foo/{{.Values.filename}}/qux"},
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
