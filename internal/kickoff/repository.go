package kickoff

// RepoRef holds information about a skeleton repository's location.
type RepoRef struct {
	// Name holds the optional local name for the repository.
	Name string `json:"name,omitempty"`
	// URL to the repository, e.g. 'https://github.com/foobar/baz' or
	// `/some/local/path`.
	URL string `json:"url,omitempty"`
	// Path is the path to a local repository.
	Path string `json:"path,omitempty"`
	// Revision holds the revision of the remote git repository to checkout.
	// This can be a branch, tag or commit SHA.
	Revision string `json:"revision,omitempty"`
}

// IsRemote returns true if the repo info describes a remote repository.
func (r *RepoRef) IsRemote() bool {
	return r.URL != ""
}
