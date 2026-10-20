package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Paths struct {
	ConfigDir   string
	ConfigFile  string
	ProfilesDir string
	CacheDir    string
	HistoryFile string
}

type Config struct {
	Paths
	SearchPaths    []string
	DefaultProfile string
}

type file struct {
	SearchPaths    []string `json:"search_paths"`
	DefaultProfile string   `json:"default_profile"`
}

func DefaultPaths() Paths {
	configRoot := os.Getenv("XDG_CONFIG_HOME")
	if configRoot == "" {
		configRoot = filepath.Join(home(), ".config")
	}
	cacheRoot := os.Getenv("XDG_CACHE_HOME")
	if cacheRoot == "" {
		cacheRoot = filepath.Join(home(), ".cache")
	}
	configDir := filepath.Join(configRoot, "tds")
	cacheDir := filepath.Join(cacheRoot, "tds")
	return Paths{
		ConfigDir:   configDir,
		ConfigFile:  filepath.Join(configDir, "config.json"),
		ProfilesDir: filepath.Join(configDir, "profiles"),
		CacheDir:    cacheDir,
		HistoryFile: filepath.Join(cacheDir, "history"),
	}
}

// Load reads config.json, creating it with a default search path on first run.
func Load(p Paths) (Config, error) {
	if _, err := os.Stat(p.ConfigFile); os.IsNotExist(err) {
		if err := os.MkdirAll(p.ProfilesDir, 0o755); err != nil {
			return Config{}, err
		}
		def := "{\n  \"search_paths\": [\n    \"~/projects\"\n  ]\n}\n"
		if err := os.WriteFile(p.ConfigFile, []byte(def), 0o644); err != nil {
			return Config{}, err
		}
	}
	if err := os.MkdirAll(p.ProfilesDir, 0o755); err != nil {
		return Config{}, err
	}

	raw, err := os.ReadFile(p.ConfigFile)
	if err != nil {
		return Config{}, err
	}
	var f file
	if err := json.Unmarshal(raw, &f); err != nil {
		return Config{}, fmt.Errorf("config is not valid JSON: %s: %w", p.ConfigFile, err)
	}

	cfg := Config{Paths: p, DefaultProfile: f.DefaultProfile}
	for _, sp := range f.SearchPaths {
		if sp == "" {
			continue
		}
		cfg.SearchPaths = append(cfg.SearchPaths, expand(sp))
	}
	return cfg, nil
}

// IsProjectDir reports whether dir sits directly under one of the search paths.
func (c Config) IsProjectDir(dir string) bool {
	parent := filepath.Dir(dir)
	for _, sp := range c.SearchPaths {
		if parent == strings.TrimSuffix(sp, "/") {
			return true
		}
	}
	return false
}

func expand(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		p = filepath.Join(home(), p[1:])
	}
	return os.ExpandEnv(p)
}

func home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return h
}
