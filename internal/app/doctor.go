package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"tds/internal/bookmarks"
	"tds/internal/profile"
	"tds/internal/tmux"
)

func cmdDoctor(a *App, args []string) {
	status := 0
	ok := func(format string, args ...any) { fmt.Printf("OK    "+format+"\n", args...) }
	warnLine := func(format string, args ...any) { fmt.Printf("WARN  "+format+"\n", args...) }
	errLine := func(format string, args ...any) {
		fmt.Printf("ERR   "+format+"\n", args...)
		status = 1
	}

	for _, tool := range []string{"tmux", "fzf"} {
		if _, err := exec.LookPath(tool); err == nil {
			ok("%s found", tool)
		} else {
			errLine("%s is not installed", tool)
		}
	}

	if _, err := os.Stat(a.cfg.ConfigFile); err == nil {
		ok("config parses: %s", a.cfg.ConfigFile)
	} else {
		errLine("config is missing: %s", a.cfg.ConfigFile)
	}
	if len(a.cfg.SearchPaths) == 0 {
		warnLine("no search_paths configured")
	}
	for _, sp := range a.cfg.SearchPaths {
		if info, err := os.Stat(sp); err == nil && info.IsDir() {
			ok("search path exists: %s", sp)
		} else {
			warnLine("search path missing: %s", sp)
		}
	}

	for _, d := range bookmarks.Read(a.cfg.BookmarksFile) {
		if !bookmarks.Exists(d) {
			warnLine("bookmark directory is missing: %s", d)
		}
	}

	files, _ := filepath.Glob(filepath.Join(a.cfg.ProfilesDir, "*.json"))
	if len(files) == 0 {
		warnLine("no profiles in %s", a.cfg.ProfilesDir)
	}
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json")
		if _, err := profile.Load(a.cfg.ProfilesDir, name); err == nil {
			ok("profile valid: %s", name)
		} else {
			errLine("profile invalid (no windows array): %s", f)
		}
	}

	if def := a.cfg.DefaultProfile; def != "" {
		if _, err := profile.Load(a.cfg.ProfilesDir, def); err == nil {
			ok("default_profile '%s' is valid", def)
		} else {
			errLine("default_profile '%s' is missing or invalid", def)
		}
	}

	for _, s := range tmux.ListManaged() {
		if info, err := os.Stat(s.Path); err == nil && info.IsDir() {
			ok("session %s -> %s", s.Name, s.Path)
		} else {
			warnLine("session %s: directory is missing: %s", s.Name, s.Path)
		}
		if s.Profile != "" && s.Profile != "basic" {
			if _, err := profile.Load(a.cfg.ProfilesDir, s.Profile); err != nil {
				warnLine("session %s: profile '%s' no longer exists", s.Name, s.Profile)
			}
		}
	}

	os.Exit(status)
}
