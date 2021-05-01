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
	// SkeletonTemplateExtension is the file extension for template files
	// within kickoff skeletons. Although kickoff template files are
	// gotemplates, we must not use .tmpl as user may want to include their own
	// gotemplate files in skeletons which must not be evaluated by kickoff,
	// hence we use .skel to avoid issues here.
	SkeletonTemplateExtension = ".skel"
	// SkeletonsDir is the subdirectory of a repository where skeletons
	// can be found.
	SkeletonsDir = "skeletons"
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
	// EnvKeyLogLevel configures custom path to the kickoff config file.
	EnvKeyConfig = "KICKOFF_CONFIG"
	// EnvKeyEditor configures the command used to edit config files.
	EnvKeyEditor = "KICKOFF_EDITOR"
	// EnvKeyLogLevel configures the log level.
	EnvKeyLogLevel = "KICKOFF_LOG_LEVEL"
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
	// LocalCacheDir points to the user's local cache dir which is
	// platform specific.
	LocalCacheDir = configdir.LocalCache("kickoff")

	LocalRepositoryCacheDir = filepath.Join(LocalCacheDir, "repositories")
)
