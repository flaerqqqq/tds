package history

import (
	"path/filepath"
	"testing"
)

func TestAddPrependsDedupesAndCaps(t *testing.T) {
	file := filepath.Join(t.TempDir(), "history")
	for _, d := range []string{"/a", "/b", "/c", "/a"} {
		if err := Add(file, d, 2); err != nil {
			t.Fatal(err)
		}
	}
	got := Read(file)
	if len(got) != 2 || got[0] != "/a" || got[1] != "/c" {
		t.Fatalf("history = %v, want [/a /c]", got)
	}
}
