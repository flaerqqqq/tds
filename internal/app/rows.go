package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

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
		row := []string{dir, fitLabel(label), "-", "-"}
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

// labelWidth caps the project column so one long path cannot push the other columns off screen.
const labelWidth = 40

// fitLabel shortens a path to labelWidth runes. Middle directories go first, as "…", keeping the
// start and the last segments. A single segment is elided in its middle.
func fitLabel(label string) string {
	if utf8.RuneCountInString(label) <= labelWidth {
		return label
	}
	parts := strings.Split(label, "/")
	head, rest := parts[0], parts[1:]
	if head == "" && len(rest) > 0 {
		head, rest = "/"+rest[0], rest[1:]
	}
	if len(rest) >= 2 {
		best := ""
		for i := len(rest) - 1; i >= 1; i-- {
			cand := head + "/…/" + strings.Join(rest[i:], "/")
			if utf8.RuneCountInString(cand) > labelWidth {
				break
			}
			best = cand
		}
		if best != "" {
			return best
		}
	}
	return elideMiddle(label, labelWidth)
}

func elideMiddle(s string, width int) string {
	r := []rune(s)
	keep := width - 1
	start := keep / 2
	end := keep - start
	return string(r[:start]) + "…" + string(r[len(r)-end:])
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
		rows = append(rows, []string{s.Name, fitLabel(label), s.Name, s.Profile, age(now, s.Activity)})
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

// pickerInput is withHeader for pickers that run with --ansi. The label of each row with
// dim[i] set is gray. Coloring happens after padding so escape codes don't skew the columns.
func pickerInput(header []string, rows [][]string, dim []bool) string {
	all := append([][]string{header}, rows...)
	padded := tabulate.Pad(all)
	lines := make([]string, 0, len(padded))
	for i, row := range padded {
		if i > 0 && dim[i-1] {
			row[1] = grayStart + row[1] + grayEnd
		}
		lines = append(lines, strings.Join(row, "\t"))
	}
	return strings.Join(lines, "\n") + "\n"
}

const (
	grayStart = "\x1b[90m"
	grayEnd   = "\x1b[0m"
)
