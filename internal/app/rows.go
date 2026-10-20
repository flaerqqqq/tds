package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tds/internal/session"
	"tds/internal/tabulate"
	"tds/internal/tmux"
)

var projectHeader = []string{"", "PROJECT", "SESSION", "PROFILE"}
var sessionHeader = []string{"", "PROJECT", "SESSION", "PROFILE", "ACTIVE"}

// tilde shortens paths under $HOME to ~/..., leaving $HOME itself absolute.
func tilde(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	if strings.HasPrefix(p, home+"/") {
		return "~/" + strings.TrimPrefix(p, home+"/")
	}
	return p
}

// projectRows returns [dir, label, session, profile] rows. With short, projects get their
// basename and other directories their path with ~.
func (a *App) projectRows(dirs []string, short bool, managed map[string]string) [][]string {
	rows := make([][]string, 0, len(dirs))
	for _, dir := range dirs {
		label := dir
		if short {
			if a.cfg.IsProjectDir(dir) {
				label = filepath.Base(dir)
			} else {
				label = tilde(dir)
			}
		}
		name := session.Name(dir)
		row := []string{dir, label, "-", "-"}
		if prof, ok := managed[name]; ok {
			prof = strings.TrimSpace(prof)
			if prof == "" {
				prof = "-"
			}
			row[2], row[3] = name, prof
		}
		rows = append(rows, row)
	}
	return rows
}

func managedProfiles() map[string]string {
	out := map[string]string{}
	for _, s := range tmux.ListManaged() {
		out[s.Name] = s.Profile
	}
	return out
}

// sessionRows returns managed sessions with the current one first, then most recently active.
func (a *App) sessionRows() [][]string {
	sessions := tmux.ListManaged()
	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].Activity > sessions[j].Activity
	})
	current := ""
	if tmux.InTmux() {
		current = tmux.CurrentSession()
	}
	ordered := make([]tmux.Session, 0, len(sessions))
	for _, s := range sessions {
		if s.Name == current {
			ordered = append([]tmux.Session{s}, ordered...)
		} else {
			ordered = append(ordered, s)
		}
	}
	now := time.Now().Unix()
	rows := make([][]string, 0, len(ordered))
	for _, s := range ordered {
		label := tilde(s.Path)
		if a.cfg.IsProjectDir(s.Path) {
			label = filepath.Base(s.Path)
		}
		rows = append(rows, []string{s.Name, label, s.Name, s.Profile, age(now, s.Activity)})
	}
	return rows
}

func age(now, then int64) string {
	d := now - then
	switch {
	case d < 60:
		return "just now"
	case d < 3600:
		return fmt.Sprintf("%dm ago", d/60)
	case d < 86400:
		return fmt.Sprintf("%dh ago", d/3600)
	default:
		return fmt.Sprintf("%dd ago", d/86400)
	}
}

// withHeader prepends the header row and returns the padded picker lines.
func withHeader(header []string, rows [][]string) string {
	all := append([][]string{header}, rows...)
	return strings.Join(tabulate.Lines(all), "\n") + "\n"
}
