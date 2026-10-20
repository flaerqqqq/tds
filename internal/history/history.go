package history

import (
	"os"
	"path/filepath"
	"strings"
)

// Add moves dir to the top of the history file and keeps at most limit entries.
func Add(file, dir string, limit int) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	entries := []string{dir}
	for _, line := range Read(file) {
		if line != dir {
			entries = append(entries, line)
		}
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return os.WriteFile(file, []byte(strings.Join(entries, "\n")+"\n"), 0o644)
}

// Read returns the entries, newest first. A missing file gives no entries.
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
