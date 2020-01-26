package skeleton

import (
	"errors"
	"fmt"
	"sort"

	"github.com/martinohmann/kickoff/pkg/template"
)

// Merge merges multiple skeletons together and returns a new *Skeleton.
// The skeletons are merged left to right with template values, skeleton files
// and skeleton info of the rightmost skeleton taking preference over already
// existing values. Template values are recursively merged and may cause errors
// on type mismatch. The resulting *Skeleton will have the second-to-last
// skeleton set as its parent, the second-to-last will have the third-to-last
// as parent and so forth. The original skeletons are not altered. If only one
// skeleton is passed it will be returned as is without modification. Passing a
// slice with length of zero will return in an error.
func Merge(skeletons ...*Skeleton) (*Skeleton, error) {
	if len(skeletons) == 0 {
		return nil, errors.New("cannot merge empty list of skeletons")
	}

	if len(skeletons) == 1 {
		return skeletons[0], nil
	}

	lhs := skeletons[0]

	var err error

	for _, rhs := range skeletons[1:] {
		lhs, err = merge(lhs, rhs)
		if err != nil {
			return nil, err
		}
	}

	return lhs, nil
}

func merge(lhs, rhs *Skeleton) (*Skeleton, error) {
	values, err := template.MergeValues(lhs.Values, rhs.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to merge skeleton %s and %s: %v", lhs.Info, rhs.Info, err)
	}

	s := &Skeleton{
		Values:      values,
		Files:       mergeFiles(lhs.Files, rhs.Files),
		Description: rhs.Description,
		Parent:      lhs,
		Info:        rhs.Info,
	}

	return s, nil
}

func mergeFiles(lhs, rhs []*File) []*File {
	fileMap := make(map[string]*File)

	for _, f := range lhs {
		fileMap[f.RelPath] = &File{
			AbsPath:   f.AbsPath,
			RelPath:   f.RelPath,
			Info:      f.Info,
			Inherited: true,
		}
	}

	for _, f := range rhs {
		fileMap[f.RelPath] = f
	}

	filePaths := make([]string, 0, len(fileMap))
	for path := range fileMap {
		filePaths = append(filePaths, path)
	}

	sort.Strings(filePaths)

	files := make([]*File, len(filePaths))
	for i, path := range filePaths {
		files[i] = fileMap[path]
	}

	return files
}
