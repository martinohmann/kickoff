// Package filetree provides a tree type which can be used to build and print
// file trees of skeletons.
package filetree

import (
	"strings"

	gotree "github.com/disiqueira/gotree/v3"
	"github.com/fatih/color"
	"github.com/martinohmann/kickoff/internal/skeleton"
)

var (
	bold  = color.New(color.Bold)
	green = color.New(color.FgGreen)
)

type tree struct {
	gotree.Tree
}

// Build builds a printable file tree for s.
func Build(s *skeleton.Skeleton) gotree.Tree {
	root := New(s.Info.Name)

	for _, f := range s.Files {
		parts := strings.Split(f.RelPath, "/")
		if parts[0] == "." {
			// @TODO(mohmann): the skeleton dir should be removed from the file
			// list at one point. We'll ignore it for now.
			continue
		}
		for tree := root; len(parts) > 0; parts = parts[1:] {
			tree = tree.Add(parts[0])
		}
	}

	return root
}

// New creates a new tree node with text.
func New(text string) gotree.Tree {
	return &tree{
		Tree: gotree.New(text),
	}
}

// Text implements gotree.Tree.
//
// Returns formatted node text. This is not necessarily the text that the node
// was created with.
func (t *tree) Text() string {
	text := t.Tree.Text()
	if len(t.Items()) > 0 {
		return bold.Sprint(text + "/")
	}

	if strings.HasSuffix(text, ".skel") {
		return green.Sprint(text)
	}

	return text
}

// AddTree implements gotree.Tree.
//
// It adds the other as a child of t.
func (t *tree) AddTree(other gotree.Tree) {
	if o, ok := other.(*tree); !ok {
		other = &tree{o}
	}
	t.Tree.AddTree(other)
}

// Add implements gotree.Tree.
//
// It adds a new tree node with text if not present yet and returns it. If
// present, it just returns the existing node.
func (t *tree) Add(text string) gotree.Tree {
	if item := t.find(text); item != nil {
		return item
	}

	n := New(text)
	t.AddTree(n)
	return n
}

// Print implements gotree.Tree.
//
// Prints the tree with all leading and trailing whitespace trimmed.
func (t *tree) Print() string {
	return strings.TrimSpace(t.Tree.Print())
}

func (t *tree) find(text string) gotree.Tree {
	for _, item := range t.Items() {
		itemText := item.Text()
		if ft, ok := item.(*tree); ok {
			// use the raw text of the underlying node for comparison.
			itemText = ft.Tree.Text()
		}

		if itemText == text {
			return item
		}
	}

	return nil
}
