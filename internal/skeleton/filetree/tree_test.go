package filetree

import (
	"testing"

	"github.com/martinohmann/kickoff/internal/skeleton"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	s := &skeleton.Skeleton{
		Info: &skeleton.Info{Name: "my/skeleton"},
		Files: []*skeleton.File{
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
