package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadValidProfile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "dev", `{"windows":[{"name":"Code","command":"nvim ."},{"name":"Terminal","command":""}]}`)
	p, err := Load(dir, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Windows) != 2 || p.Windows[0].Command != "nvim ." {
		t.Fatalf("unexpected windows: %+v", p.Windows)
	}
}

func TestLoadRejectsMissingOrNonArrayWindows(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "nowin", `{"foo":1}`)
	write(t, dir, "strwin", `{"windows":"x"}`)
	for _, name := range []string{"nowin", "strwin", "absent"} {
		if Valid(dir, name) {
			t.Errorf("%s should be invalid", name)
		}
	}
}

func TestEmptyWindowsArrayIsValid(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "empty", `{"windows":[]}`)
	if !Valid(dir, "empty") {
		t.Fatal("empty windows array should be valid, as in bash")
	}
}
