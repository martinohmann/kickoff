package template

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestRenderFile(t *testing.T) {
	content := []byte("foo: {{.foo}}")
	f, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("unexpected error while creating temp file: %v", err)
	}

	filename := f.Name()

	defer os.Remove(filename)

	_, err = f.Write(content)
	if err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	err = f.Close()
	if err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	data := map[string]interface{}{"foo": "bar"}

	rendered, err := RenderFile(filename, data)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	expected := `foo: bar`

	if rendered != expected {
		t.Fatalf("expected %q, got %q", expected, rendered)
	}
}
