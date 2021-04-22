package kickoff

import (
	"os"
	"sort"
)

// BufferedFile is a file that is already buffered in memory.
type BufferedFile struct {
	// RelPath is the file path relative to root directory of the skeleton.
	// This is used to construct the path for the file relative to the the
	// target project directory.
	RelPath string `json:"relPath"`
	// Content holds the file contents.
	Content []byte `json:"content"`
	// Mode is the os.Mode for the file. This provides information about the
	// type of file, e.g. whether it is a directory or not.
	Mode os.FileMode `json:"mode"`
}

// MergeFiles merges two lists of files. Files in the rhs list take precedence
// over files in the lhs list with the same name.
func MergeFiles(lhs, rhs []*BufferedFile) []*BufferedFile {
	fileMap := make(map[string]*BufferedFile)

	for _, f := range lhs {
		fileMap[f.RelPath] = f
	}

	for _, f := range rhs {
		fileMap[f.RelPath] = f
	}

	filePaths := make([]string, 0, len(fileMap))
	for path := range fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]*BufferedFile, len(filePaths))
	for i, path := range filePaths {
		files[i] = fileMap[path]
	}

	return files
}
