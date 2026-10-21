package bookmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Normalize returns the absolute, cleaned form stored in the file.
func Normalize(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// Exists reports whether dir is currently an existing directory.
func Exists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// Read returns the bookmarks in insertion order. A missing file gives none.
func Read(file string) []string {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// Add appends dir, which must already be normalized. It reports false when dir is already present.
func Add(file, dir string) (bool, error) {
	entries := Read(file)
	for _, e := range entries {
		if e == dir {
			return false, nil
		}
	}
	entries = append(entries, dir)
	return true, write(file, entries)
}

// Remove drops dir, which must already be normalized. It reports false when dir was not present.
func Remove(file, dir string) (bool, error) {
	entries := Read(file)
	kept := make([]string, 0, len(entries))
	for _, e := range entries {
		if e != dir {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(entries) {
		return false, nil
	}
	return true, write(file, kept)
}

func write(file string, entries []string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", filepath.Dir(file), err)
	}
	content := ""
	if len(entries) > 0 {
		content = strings.Join(entries, "\n") + "\n"
	}
	return os.WriteFile(file, []byte(content), 0o644)
}
