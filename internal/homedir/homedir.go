// Package homedir provides functionality to expand `~` to the absolute home
// directory of a user and vice-versa.
package homedir

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// Collapse replaces the homedir in absolute paths with `~`. E.g.
// `/home/user/foo` will be rewritten to `~/foo`. Relative paths are returned
// as is. Returns an error of the home dir of the current user cannot be
// determined.
func Collapse(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return path, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	unprefixed := strings.TrimPrefix(path, home)
	if unprefixed == path {
		return path, nil
	}

	if len(unprefixed) == 0 {
		return "~", nil
	}

	return fmt.Sprintf("~/%s", strings.TrimLeft(unprefixed, "/")), nil
}

// MustCollapse is the same as Collapse but will panic on errors while
// attempting to collapse the home directory.
func MustCollapse(path string) string {
	path, err := Collapse(path)
	if err != nil {
		log.WithError(err).Panic("failed to collapse homedir")
	}

	return path
}

// Expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func Expand(path string) (string, error) {
	return homedir.Expand(path)
}

// MustExpand is the same as Expand but will panic on errors while attempting
// to expand the home directory.
func MustExpand(path string) string {
	path, err := Expand(path)
	if err != nil {
		log.WithError(err).Panic("failed to expand homedir")
	}

	return path
}
