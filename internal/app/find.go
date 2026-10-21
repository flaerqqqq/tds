package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"tds/internal/bookmarks"
	"tds/internal/tmux"
)

func cmdFind(a *App, args []string) {
	a.parseFlags(args)
	if len(a.opts.Positional) > 0 {
		die("find takes no directory argument.")
	}
	a.resolveLayout()
	var target string
	if a.opts.Yazi {
		target = pickWithYazi()
	} else {
		target = a.pickProject()
	}
	a.openDirectory(target)
}

func pickWithYazi() string {
	requireCmd("yazi")
	tmp, err := os.CreateTemp("", "tds.")
	if err != nil {
		die("%v", err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	cmd := exec.Command("yazi", "--cwd-file="+tmp.Name())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	_ = cmd.Run()
	raw, _ := os.ReadFile(tmp.Name())
	target := strings.TrimRight(string(raw), "\n")
	if target == "" {
		cancel()
	}
	return target
}

// mergeProjects joins search-path children and bookmarks, keeping the first occurrence of each.
func mergeProjects(children, marks []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, d := range append(append([]string{}, children...), marks...) {
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// listProjects returns the search-path children followed by bookmarks. The current directory
// comes first; when it is not a project it is marked with "*". The second result marks rows
// whose directory no longer exists.
func (a *App) listProjects() ([][]string, []bool) {
	var children []string
	for _, sp := range a.cfg.SearchPaths {
		entries, err := os.ReadDir(sp)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				children = append(children, filepath.Join(sp, e.Name()))
			}
		}
	}
	sort.Strings(children)
	marks := bookmarks.Read(a.cfg.BookmarksFile)
	dirs := mergeProjects(children, marks)
	bookmarked := make(map[string]bool, len(marks))
	for _, d := range marks {
		bookmarked[d] = true
	}

	here, _ := os.Getwd()
	if tmux.InTmux() && os.Getenv("TMUX_PANE") == "" {
		if managed, path := tmux.ActiveSessionInfo(); managed && path != "" {
			here = path
		}
	}

	mark := ""
	ordered := []string{here}
	found := false
	for _, d := range dirs {
		if d == here {
			found = true
			continue
		}
		ordered = append(ordered, d)
	}
	if !found {
		mark = "* "
	}

	rows := a.projectRows(ordered, true, managedProfiles())
	for i := range rows {
		prefix := "  "
		switch {
		case i == 0 && mark != "":
			prefix = mark
		case bookmarked[ordered[i]]:
			prefix = bookmarkIcon + " "
		}
		rows[i][1] = prefix + rows[i][1]
	}
	dim := make([]bool, len(ordered))
	for i, d := range ordered {
		dim[i] = !bookmarks.Exists(d)
	}
	return rows, dim
}

func (a *App) pickProject() string {
	if len(a.cfg.SearchPaths) == 0 && len(bookmarks.Read(a.cfg.BookmarksFile)) == 0 {
		die("No search paths or bookmarks. Add a search path to %s or run 'tds bm add DIR'.", a.cfg.ConfigFile)
	}

	label := "Projects & Submodules"
	if a.layout.label != "basic" {
		label = "Projects (Profile: " + a.layout.label + ")"
	}

	return a.pickFrom(pickerSpec{
		border:   " 󰅍 " + label + " ",
		prompt:   "󰝰 Project > ",
		help:     "enter open · ctrl-r rebuild · ctrl-p rebuild with profile · ctrl-x kill session · ctrl-d delete bookmark",
		rows:     a.listProjects,
		onDelete: a.removeBookmark,
	})
}
