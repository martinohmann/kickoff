// Package file provides utility functionality for working with files and
// directories.
package file

import (
	"fmt"
	"io"
	"os"
)

// Copy copies srcPath to dstPath. It internally uses io.Copy. The destination
// file is created with the same file mode as the source file. Returns any
// error that occurred prior to or during the copy operation.
func Copy(srcPath, dstPath string) error {
	fileInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", srcPath)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// Exists returns true if path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

// IsDirectory returns true if path is a directory. Returns an error as the
// second return value if calling stat on path failed.
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
