package kickoff

import (
	"reflect"
	"testing"
)

func TestFindSkeletons(t *testing.T) {
	skeletons, err := FindSkeletons("testdata/skeletons")
	if err != nil {
		t.Fatalf("expected nil error but got: %v", err)
	}

	expected := []string{
		"bar",
		"foo/bar",
		"nested",
		"nested/skeletons",
	}

	if !reflect.DeepEqual(expected, skeletons) {
		t.Fatalf("expected %#v, got %#v", expected, skeletons)
	}
}

func TestFindSkeletons_Error(t *testing.T) {
	_, err := FindSkeletons("testdata/nonexistent")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}
