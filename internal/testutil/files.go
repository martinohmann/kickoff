package testutil

import (
	"github.com/martinohmann/kickoff/internal/kickoff"
)

// MockRepositoryCacheDir changes the global repository cache dir to a
// temporary directory and returns a func to restore it.
//
// Usage example in tests:
//
//	defer testutil.MockRepositoryCacheDir(t.TempDir())()
func MockRepositoryCacheDir(dir string) func() {
	oldCacheDir := kickoff.LocalRepositoryCacheDir
	kickoff.LocalRepositoryCacheDir = dir
	return func() { kickoff.LocalRepositoryCacheDir = oldCacheDir }
}

// MockDefaultConfigPath changes the global default config path to a
// user-defined location and returns a func to restore it.
//
// Usage example in tests:
//
//	defer testutil.MockDefaultConfigPath(configPath)()
func MockDefaultConfigPath(path string) func() {
	oldConfigPath := kickoff.DefaultConfigPath
	kickoff.DefaultConfigPath = path
	return func() { kickoff.DefaultConfigPath = oldConfigPath }
}
