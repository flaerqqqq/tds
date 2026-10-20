package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultConfig(t *testing.T) {
	root := t.TempDir()
	p := Paths{
		ConfigDir:   root,
		ConfigFile:  filepath.Join(root, "config.json"),
		ProfilesDir: filepath.Join(root, "profiles"),
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.ConfigFile); err != nil {
		t.Fatalf("config not created: %v", err)
	}
	if len(cfg.SearchPaths) != 1 || filepath.Base(cfg.SearchPaths[0]) != "projects" {
		t.Fatalf("unexpected search paths: %v", cfg.SearchPaths)
	}
}

func TestLoadExpandsHomeAndSkipsEmpty(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.json"),
		[]byte(`{"default_profile":"dev","search_paths":["~/dev/projects","","/abs/x"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Paths{ConfigFile: filepath.Join(root, "config.json"), ProfilesDir: filepath.Join(root, "profiles")}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProfile != "dev" {
		t.Fatalf("default profile = %q", cfg.DefaultProfile)
	}
	want := []string{filepath.Join(home(), "dev/projects"), "/abs/x"}
	if len(cfg.SearchPaths) != 2 || cfg.SearchPaths[0] != want[0] || cfg.SearchPaths[1] != want[1] {
		t.Fatalf("search paths = %v, want %v", cfg.SearchPaths, want)
	}
}

func TestLoadRejectsBadJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "config.json"), []byte(`{`), 0o644); err != nil {
		t.Fatal(err)
	}
	p := Paths{ConfigFile: filepath.Join(root, "config.json"), ProfilesDir: filepath.Join(root, "profiles")}
	if _, err := Load(p); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestIsProjectDir(t *testing.T) {
	cfg := Config{SearchPaths: []string{"/home/u/dev/projects"}}
	if !cfg.IsProjectDir("/home/u/dev/projects/demo") {
		t.Error("demo should be a project")
	}
	if cfg.IsProjectDir("/home/u/demo") {
		t.Error("~/demo should not be a project")
	}
	if cfg.IsProjectDir("/home/u/dev/projects") {
		t.Error("the search root itself should not be a project")
	}
}
