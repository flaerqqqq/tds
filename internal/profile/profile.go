package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Window struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type Profile struct {
	Name    string
	Windows []Window
}

// Load reads and validates profiles/<name>.json. A profile must have a "windows" array.
func Load(dir, name string) (Profile, error) {
	path := filepath.Join(dir, name+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, fmt.Errorf("profile '%s' not found at %s", name, path)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return Profile{}, fmt.Errorf("profile '%s' is invalid JSON: %w", name, err)
	}
	w, ok := top["windows"]
	if !ok || !strings.HasPrefix(strings.TrimSpace(string(w)), "[") {
		return Profile{}, fmt.Errorf("profile '%s' is invalid (expected a 'windows' array)", name)
	}
	var p Profile
	if err := json.Unmarshal(w, &p.Windows); err != nil {
		return Profile{}, fmt.Errorf("profile '%s' has invalid windows: %w", name, err)
	}
	p.Name = name
	return p, nil
}

// Valid reports whether the profile file exists and has a "windows" array.
func Valid(dir, name string) bool {
	_, err := Load(dir, name)
	return err == nil
}
