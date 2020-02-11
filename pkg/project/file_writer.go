package project

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/martinohmann/kickoff/pkg/skeleton"
	"github.com/martinohmann/kickoff/pkg/template"
	"github.com/spf13/afero"
)

// SkeletonFileWriter writes skeleton files into a target directory after
// rendering all templates contained in them.
type SkeletonFileWriter struct {
	fs     afero.Fs
	dirMap map[string]string
}

// NewSkeletonFileWriter creates a new *SkeletonFileWriter which uses fs as the
// target filesystem for all files and directories that may be created.
func NewSkeletonFileWriter(fs afero.Fs) *SkeletonFileWriter {
	return &SkeletonFileWriter{
		fs:     fs,
		dirMap: make(map[string]string),
	}
}

// WriteFiles writes files to targetDir. The provided template values are
// passed to template files and template filenames before rendering them.
func (fw *SkeletonFileWriter) WriteFiles(files []*skeleton.File, targetDir string, values template.Values) error {
	for _, f := range files {
		err := fw.writeFile(f, targetDir, values)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fw *SkeletonFileWriter) writeFile(f *skeleton.File, targetDir string, values template.Values) error {
	if f.Inherited {
		log.WithField("path", f.AbsPath).Debug("processing inherited resource")
	}

	targetRelPath, err := fw.buildTargetRelPath(f, values)
	if err != nil {
		return err
	}

	targetAbsPath := filepath.Join(targetDir, targetRelPath)

	logger := log.WithField("path", f.RelPath)

	if f.RelPath != targetRelPath {
		logger = log.WithFields(log.Fields{
			"path.src":    f.RelPath,
			"path.target": targetRelPath,
		})
	}

	switch {
	case f.Info.IsDir():
		logger.Info("creating directory")

		// Track potentially templated directory names that need to be
		// rewritten for all files contained in them.
		if f.RelPath != targetRelPath {
			fw.dirMap[f.RelPath] = targetRelPath
		}

		return fw.fs.MkdirAll(targetAbsPath, f.Info.Mode())
	case filepath.Ext(f.RelPath) == ".skel":
		logger.Info("rendering template")

		contents, err := template.RenderFile(f.AbsPath, values)
		if err != nil {
			return err
		}

		return afero.WriteFile(fw.fs, targetAbsPath, []byte(contents), f.Info.Mode())
	default:
		logger.Info("copying file")

		targetFile, err := fw.fs.Create(targetAbsPath)
		if err != nil {
			return err
		}

		srcFile, err := os.Open(f.AbsPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(targetFile, srcFile)
		if err != nil {
			return err
		}

		return fw.fs.Chmod(targetAbsPath, f.Info.Mode())
	}
}

func (fw *SkeletonFileWriter) buildTargetRelPath(f *skeleton.File, values template.Values) (string, error) {
	srcFilename := filepath.Base(f.RelPath)
	srcRelDir := filepath.Dir(f.RelPath)

	targetFilename, err := template.RenderText(srcFilename, values)
	if err != nil {
		return "", err
	}

	if len(targetFilename) == 0 {
		return "", fmt.Errorf("templated filename %q resolved to an empty string", srcFilename)
	}

	targetRelDir := srcRelDir

	// If the src dir's name was templated, lookup the resolved name and
	// use that as destination.
	if dir, ok := fw.dirMap[srcRelDir]; ok {
		targetRelDir = dir
	}

	targetRelPath := filepath.Join(targetRelDir, targetFilename)

	// Sanity check to guard against malicious injection of directory
	// traveral (e.g. "../../" in the template string).
	if filepath.Dir(targetRelPath) != targetRelDir {
		return "", fmt.Errorf("templated filename %q injected illegal directory traversal: %s", srcFilename, targetFilename)
	}

	// Trim .skel extension.
	if ext := filepath.Ext(targetRelPath); ext == ".skel" {
		targetRelPath = targetRelPath[:len(targetRelPath)-len(ext)]
	}

	return targetRelPath, nil
}
