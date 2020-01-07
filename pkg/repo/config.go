package repo

import (
	"path/filepath"

	"github.com/kirsle/configdir"
)

var (
	localConfigDir = configdir.LocalConfig("kickoff")

	DefaultRepositoryURL = filepath.Join(localConfigDir, "repository")
)

type Config struct {
	URL string `json:"url"`
}

func (c *Config) ApplyDefaults() {
	if c.URL == "" {
		c.URL = DefaultRepositoryURL
	}
}
