package app

import (
	"reflect"
	"strings"
	"testing"
)

func TestMergeProjectsDedupesAndKeepsOrder(t *testing.T) {
	children := []string{"/p/a", "/p/b"}
	marks := []string{"/x/z", "/p/b", "/x/y"}
	got := mergeProjects(children, marks)
	want := []string{"/p/a", "/p/b", "/x/z", "/x/y"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeProjects = %v, want %v", got, want)
	}
}

func TestMergeProjectsDoesNotAliasInput(t *testing.T) {
	children := make([]string, 1, 4)
	children[0] = "/p/a"
	mergeProjects(children, []string{"/x"})
	if len(children) != 1 {
		t.Fatalf("input slice changed: %v", children)
	}
}

func TestPickerInputGraysOnlyFlaggedLabels(t *testing.T) {
	header := []string{"", "PROJECT", "SESSION", "PROFILE"}
	rows := [][]string{
		{"/a", "alpha", "-", "-"},
		{"/b", "beta", "-", "-"},
	}
	out := pickerInput(header, rows, []bool{false, true})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if strings.Contains(lines[1], grayStart) {
		t.Errorf("present row is grayed: %q", lines[1])
	}
	if !strings.Contains(lines[2], grayStart+"beta") || !strings.Contains(lines[2], grayEnd) {
		t.Errorf("missing row not grayed: %q", lines[2])
	}
	if strings.Count(lines[0], "\t") != strings.Count(lines[1], "\t") {
		t.Errorf("header and row have different field counts")
	}
}
