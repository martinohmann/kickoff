package kickoff

import (
	"path/filepath"

	"github.com/kirsle/configdir"
)

const (
	// DefaultConfigFileName is the default name kickoff's user config file.
	DefaultConfigFileName = "config.yaml"
	// SkeletonConfigFileName is the name of the file that is searched to
	// file skeletons and their config.
	SkeletonConfigFileName = ".kickoff.yaml"
)

const (
	// DefaultProjectHost denotes the default git host that is passed to
	// templates so that project related urls can be rendered in files like
	// READMEs.
	DefaultProjectHost = "github.com"

	// DefaultRepositoryName is the name of the default skeleton repository.
	DefaultRepositoryName = "default"

	// DefaultSkeletonName is the name of the default skeleton in a repository.
	DefaultSkeletonName = "default"
)

const (
	// NoLicense means that no license file will be generated for a new
	// project.
	NoLicense = "none"

	// NoGitignore means that no .gitignore file will be generated for a new
	// project.
	NoGitignore = "none"
)

var (
	// LocalConfigDir points to the user's local configuration dir which is
	// platform specific.
	LocalConfigDir = configdir.LocalConfig("kickoff")

	// DefaultConfigPath holds the default kickof config path in the user's
	// local config directory.
	DefaultConfigPath = filepath.Join(LocalConfigDir, DefaultConfigFileName)

	// DefaultRepositoryURL is the url of the default skeleton repository if
	// the user did not configure anything else.
	DefaultRepositoryURL = "https://github.com/martinohmann/kickoff-skeletons"
)
