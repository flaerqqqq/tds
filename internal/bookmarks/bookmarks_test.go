package bookmarks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddDedupesAndKeepsOrder(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bookmarks")
	for _, d := range []string{"/a", "/b", "/a"} {
		if _, err := Add(file, d); err != nil {
			t.Fatal(err)
		}
	}
	got := Read(file)
	if len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
		t.Fatalf("entries = %v", got)
	}
}

func TestAddReportsDuplicate(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bookmarks")
	added, _ := Add(file, "/a")
	again, _ := Add(file, "/a")
	if !added || again {
		t.Fatalf("added=%v again=%v", added, again)
	}
}

func TestRemove(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bookmarks")
	Add(file, "/a")
	Add(file, "/b")
	removed, err := Remove(file, "/a")
	if err != nil || !removed {
		t.Fatalf("removed=%v err=%v", removed, err)
	}
	if got := Read(file); len(got) != 1 || got[0] != "/b" {
		t.Fatalf("entries = %v", got)
	}
	if removed, _ := Remove(file, "/zzz"); removed {
		t.Fatal("removing an unknown path reported success")
	}
}

func TestRemoveLastLeavesEmptyFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "bookmarks")
	Add(file, "/a")
	Remove(file, "/a")
	raw, err := os.ReadFile(file)
	if err != nil || len(raw) != 0 {
		t.Fatalf("raw=%q err=%v", raw, err)
	}
}

func TestNormalize(t *testing.T) {
	got, err := Normalize("/home/u/x/../y/")
	if err != nil || got != "/home/u/y" {
		t.Fatalf("Normalize = %q, %v", got, err)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if !Exists(dir) {
		t.Fatal("existing dir reported missing")
	}
	if Exists(filepath.Join(dir, "gone")) {
		t.Fatal("missing dir reported existing")
	}
}

func TestReadMissingFile(t *testing.T) {
	if got := Read(filepath.Join(t.TempDir(), "nope")); got != nil {
		t.Fatalf("entries = %v", got)
	}
}
