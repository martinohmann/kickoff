// Package homedir provides functionality to expand `~` to the absolute home
// directory of a user and vice-versa.
package homedir

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// Collapse replaces the homedir in absolute paths with `~`.
//
// E.g. `/home/user/foo` will be rewritten to `~/foo`. Relative paths are
// returned as is. If GOOS is windows this func does nothing. Panics if
// discovery of the home directory fails.
func Collapse(path string) string {
	if runtime.GOOS == "windows" || !filepath.IsAbs(path) {
		return path
	}

	home, err := homedir.Dir()
	if err != nil {
		log.WithError(err).Panic("failed to discover homedir")
	}

	unprefixed := strings.TrimPrefix(path, strings.TrimRight(home, "/"))
	if unprefixed == path {
		return path
	}

	if len(unprefixed) == 0 {
		return "~"
	}

	if unprefixed[0] != '/' {
		return path
	}

	return fmt.Sprintf("~%s", unprefixed)
}

// expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is. Panics if discovery of the home directory fails.
func Expand(path string) string {
	path, err := homedir.Expand(path)
	if err != nil {
		log.WithError(err).Panic("failed to expand homedir")
	}

	return path
}
