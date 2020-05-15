package template

import (
	"testing"
)

func TestRenderer(t *testing.T) {
	r := NewRenderer(Values{
		"path": "github.com/foo/Bar-baz_v1\n",
	})

	templateText := "package {{.path|goPackageName}}"

	rendered, err := r.Render(templateText)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	expected := `package barbazv1`

	if rendered != expected {
		t.Fatalf("expected %q, got %q", expected, rendered)
	}
}
