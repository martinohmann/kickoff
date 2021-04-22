package kickoff

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// FileRef contains paths and other information about a skeleton file. May also
// reference a directory with templated path segments.
type FileRef struct {
	// RelPath is the file path relative to root directory of the skeleton.
	// This is used to construct the path for the file relative to the the
	// target project directory.
	RelPath string `json:"relPath"`
	// AbsPath is the absolute path to the file on disk.
	AbsPath string `json:"absPath"`
	// FileMode is the os.FileMode for the file. This provides information
	// about the type of file, e.g. whether it is a directory or not.
	FileMode os.FileMode `json:"mode"`
}

// Path implements the File interface.
func (r *FileRef) Path() string { return r.RelPath }

// Mode implements the File interface.
func (r *FileRef) Mode() os.FileMode { return r.FileMode }

// Reader implements the File interface.
func (r *FileRef) Reader() (io.Reader, error) { return os.Open(r.AbsPath) }

// IsTemplate implements the File interface.
func (r *FileRef) IsTemplate() bool { return isTemplateFile(r) }

// BufferedFile is a file that is already buffered in memory.
type BufferedFile struct {
	// RelPath is the file path relative to root directory of the skeleton.
	// This is used to construct the path for the file relative to the the
	// target project directory.
	RelPath string `json:"relPath"`
	// Content holds the file contents.
	Content []byte `json:"content"`
	// FileMode is the os.FileMode for the file. This provides information
	// about the type of file, e.g. whether it is a directory or not.
	FileMode os.FileMode `json:"mode"`
}

// NewBufferedFile creates a new *BufferedFile.
func NewBufferedFile(relPath string, content []byte, mode os.FileMode) *BufferedFile {
	return &BufferedFile{RelPath: relPath, Content: content, FileMode: mode}
}

// Path implements the File interface.
func (f *BufferedFile) Path() string { return f.RelPath }

// Mode implements the File interface.
func (f *BufferedFile) Mode() os.FileMode { return f.FileMode }

// Reader implements the File interface.
func (f *BufferedFile) Reader() (io.Reader, error) { return bytes.NewBuffer(f.Content), nil }

// IsTemplate implements the File interface.
func (f *BufferedFile) IsTemplate() bool { return isTemplateFile(f) }

func isTemplateFile(f File) bool {
	return !f.Mode().IsDir() && filepath.Ext(f.Path()) == SkeletonTemplateExtension
}

func mergeFiles(lhs, rhs []File) []File {
	fileMap := make(map[string]File)

	for _, f := range lhs {
		fileMap[f.Path()] = f
	}

	for _, f := range rhs {
		fileMap[f.Path()] = f
	}

	filePaths := make([]string, 0, len(fileMap))
	for path := range fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]File, len(filePaths))
	for i, path := range filePaths {
		files[i] = fileMap[path]
	}

	return files
}
