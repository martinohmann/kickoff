package skeleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// Info holds information about the location of a skeleton.
type Info struct {
	Name string    `json:"name"`
	Path string    `json:"path"`
	Repo *RepoInfo `json:"repo"`
}

// String implements fmt.Stringer.
func (i *Info) String() string {
	if i.Repo == nil || i.Repo.Name == "" {
		return i.Name
	}

	return fmt.Sprintf("%s:%s", i.Repo.Name, i.Name)
}

// LoadConfig loads the skeleton config for the info.
func (i *Info) LoadConfig() (Config, error) {
	configPath := filepath.Join(i.Path, ConfigFileName)

	return LoadConfig(configPath)
}

// RepoInfo holds information about a skeleton repository.
type RepoInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	URL      string `json:"url,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// IsRemote returns true if the repo info describes a remote repository.
func (i *RepoInfo) IsRemote() bool {
	return i.URL != ""
}

// FindSkeletons recursively finds all skeletons in dir and attaches i to the
// results. Returns any error that may occur while traversing dir.
func (i *RepoInfo) FindSkeletons() ([]*Info, error) {
	return findSkeletons(i, filepath.Join(i.Path, "skeletons"))
}

func findSkeletons(repoInfo *RepoInfo, dir string) ([]*Info, error) {
	skeletons := make([]*Info, 0)

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range fileInfos {
		if !info.IsDir() {
			continue
		}

		path := filepath.Join(dir, info.Name())

		ok, err := IsSkeletonDir(path)
		if os.IsPermission(err) {
			log.Warnf("permission error, skipping dir: %v", err)
			continue
		}

		if err != nil {
			return nil, err
		}

		if ok {
			abspath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}

			skeletons = append(skeletons, &Info{
				Name: info.Name(),
				Path: abspath,
				Repo: repoInfo,
			})
			continue
		}

		skels, err := findSkeletons(repoInfo, path)
		if err != nil {
			return nil, err
		}

		for _, s := range skels {
			skeletons = append(skeletons, &Info{
				Name: filepath.Join(info.Name(), s.Name),
				Path: s.Path,
				Repo: repoInfo,
			})
		}
	}

	return skeletons, nil
}
