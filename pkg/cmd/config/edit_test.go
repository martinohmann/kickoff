package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEditCmdArgs(t *testing.T) {
	var tests = []struct {
		name     string
		path     string
		editor   string
		shell    string
		expected []string
	}{
		{
			name:     "default editor and shell",
			path:     "/tmp/foo.yaml",
			expected: []string{"sh", "-c", "vi /tmp/foo.yaml"},
		},
		{
			name:     "editor from env",
			editor:   "nvim",
			path:     "/tmp/foo.yaml",
			expected: []string{"sh", "-c", "nvim /tmp/foo.yaml"},
		},
		{
			name:     "shell from env",
			shell:    "/bin/zsh",
			path:     "/tmp/foo.yaml",
			expected: []string{"/bin/zsh", "-c", "vi /tmp/foo.yaml"},
		},
		{
			name:     "editor and shell from env",
			editor:   "nano",
			shell:    "/bin/ash",
			path:     "/tmp/foo.yaml",
			expected: []string{"/bin/ash", "-c", "nano /tmp/foo.yaml"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oldEditor := os.Getenv("EDITOR")
			oldShell := os.Getenv("SHELL")

			defer func() {
				os.Setenv("EDITOR", oldEditor)
				os.Setenv("SHELL", oldShell)
			}()

			os.Setenv("EDITOR", test.editor)
			os.Setenv("SHELL", test.shell)

			actual := getEditCmdArgs(test.path)

			assert.Equal(t, test.expected, actual)
		})
	}
}
