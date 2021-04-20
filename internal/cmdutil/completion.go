package cmdutil

import (
	"context"
	"sort"

	"github.com/martinohmann/kickoff/internal/gitignore"
	"github.com/martinohmann/kickoff/internal/license"
	log "github.com/sirupsen/logrus"
)

// RepositoryNames compiles a sorted list of repository names from the config.
// Used for completion.
func RepositoryNames(f *Factory) []string {
	config, err := f.Config()
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(config.Repositories))
	for name := range config.Repositories {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// SkeletonNames compiles a sorted list of skeleton names from the configured
// repositories. Used for completion.
func SkeletonNames(f *Factory, repoNames ...string) []string {
	repo, err := f.Repository(repoNames...)
	if err != nil {
		return nil
	}

	refs, err := repo.ListSkeletons()
	if err != nil {
		return nil
	}

	names := make([]string, len(refs))
	for i, ref := range refs {
		names[i] = ref.String()
	}

	sort.Strings(names)
	return names
}

// SkeletonFilenames compiles a sorted list of skeleton file names from the
// configured repositories. Used for completion.
func SkeletonFilenames(f *Factory, skeletonName string, repoNames ...string) []string {
	repo, err := f.Repository(repoNames...)
	if err != nil {
		return nil
	}

	skeleton, err := repo.LoadSkeleton(skeletonName)
	if err != nil {
		return nil
	}

	paths := make([]string, 0, len(skeleton.Files))
	for _, file := range skeleton.Files {
		if file.Mode().IsRegular() {
			paths = append(paths, file.Path())
		}
	}

	sort.Strings(paths)
	return paths
}

// GitignoreNames compiles a sorted list of gitignore template names. Used for
// completion.
func GitignoreNames(f *Factory) []string {
	client := gitignore.NewClient(f.HTTPClient())

	names, err := client.ListTemplates(context.Background())
	if err != nil {
		return nil
	}

	sort.Strings(names)
	return names
}

// LicenseNames compiles a sorted list of license names. Used for
// completion.
func LicenseNames(f *Factory) []string {
	client := license.NewClient(f.HTTPClient())

	licenses, err := client.ListLicenses(context.Background())
	if err != nil {
		return nil
	}

	names := make([]string, len(licenses))
	for i, license := range licenses {
		names[i] = license.Key
	}

	sort.Strings(names)
	return names
}

// LogLevelNames compiles a list of log level names. Used for completion.
func LogLevelNames() []string {
	names := make([]string, len(log.AllLevels))
	for i, lvl := range log.AllLevels {
		names[i] = lvl.String()
	}
	return names
}
