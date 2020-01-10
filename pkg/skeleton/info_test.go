package skeleton

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRepositoryInfo_LocalPath(t *testing.T) {
	var tests = []struct {
		name     string
		given    *RepositoryInfo
		expected string
	}{
		{
			name: "local",
			given: &RepositoryInfo{
				Local: true,
				Path:  "/tmp/myrepo",
			},
			expected: "/tmp/myrepo",
		},
		{
			name: "remote",
			given: &RepositoryInfo{
				Scheme: "https",
				Host:   "github.com",
				Path:   "user/repo",
			},
			expected: filepath.Join(LocalCache, "github.com/user/repo"),
		},
		{
			name: "remote with user",
			given: &RepositoryInfo{
				Scheme: "ssh",
				User:   "git",
				Host:   "github.com",
				Path:   "user/repo",
				Branch: "develop",
			},
			expected: filepath.Join(LocalCache, "github.com/user/repo"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.given.LocalPath()
			if actual != test.expected {
				t.Fatalf("expected %q but got %q", test.expected, actual)
			}
		})
	}
}

func TestRepositoryInfo_String(t *testing.T) {
	var tests = []struct {
		name     string
		given    *RepositoryInfo
		expected string
	}{
		{
			name: "local",
			given: &RepositoryInfo{
				Local: true,
				Path:  "/tmp/myrepo",
			},
			expected: "/tmp/myrepo",
		},
		{
			name: "remote",
			given: &RepositoryInfo{
				Scheme: "https",
				Host:   "github.com",
				Path:   "user/repo",
			},
			expected: "https://github.com/user/repo",
		},
		{
			name: "remote with user",
			given: &RepositoryInfo{
				Scheme: "ssh",
				User:   "git",
				Host:   "github.com",
				Path:   "user/repo",
			},
			expected: "ssh://git@github.com/user/repo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.given.String()
			if actual != test.expected {
				t.Fatalf("expected %q but got %q", test.expected, actual)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	pwd, _ := os.Getwd()

	var tests = []struct {
		name        string
		given       string
		expected    *RepositoryInfo
		expectedErr error
	}{
		{
			name:  "local",
			given: "my/repo",
			expected: &RepositoryInfo{
				Local: true,
				Path:  pwd + "/my/repo",
			},
		},
		{
			name:  "local with branch",
			given: "my/repo?branch=develop",
			expected: &RepositoryInfo{
				Local:  true,
				Path:   pwd + "/my/repo",
				Branch: "develop",
			},
		},
		{
			name:  "remote https",
			given: "https://github.com/martinohmann/kickoff-skeletons",
			expected: &RepositoryInfo{
				Local:  false,
				Scheme: "https",
				Host:   "github.com",
				Path:   "martinohmann/kickoff-skeletons",
			},
		},
		{
			name:  "remote ssh with branch",
			given: "ssh://git@github.com:22/martinohmann/kickoff-skeletons?branch=v1.1.1",
			expected: &RepositoryInfo{
				Local:  false,
				Scheme: "ssh",
				User:   "git",
				Host:   "github.com:22",
				Path:   "martinohmann/kickoff-skeletons",
				Branch: "v1.1.1",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseRepositoryURL(test.given)
			switch {
			case test.expectedErr != nil && err == nil:
				t.Fatalf("expected error %#v but got nil", test.expectedErr)
			case test.expectedErr != nil && err != nil:
				if !reflect.DeepEqual(test.expectedErr, err) {
					t.Fatalf("expected error %#v but got %v", test.expectedErr, err)
				}
			case test.expectedErr == nil && err != nil:
				t.Fatalf("expected nil error but got %#v", err)
			case test.expectedErr == nil && err == nil:
				if !reflect.DeepEqual(test.expected, actual) {
					t.Fatalf("expected %#v but got %#v", test.expected, actual)
				}
			}
		})
	}
}
