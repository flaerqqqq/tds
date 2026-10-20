package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"tds/internal/picker"
	"tds/internal/session"
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

// listProjects returns directories directly under the search paths. The current directory
// comes first; when it is not a project it is marked with "*".
func (a *App) listProjects() [][]string {
	var dirs []string
	for _, sp := range a.cfg.SearchPaths {
		entries, err := os.ReadDir(sp)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				dirs = append(dirs, filepath.Join(sp, e.Name()))
			}
		}
	}
	sort.Strings(dirs)

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
	if len(rows) > 0 {
		rows[0][1] = mark + rows[0][1]
	}
	return rows
}

func (a *App) pickProject() string {
	requireCmd("fzf")
	if len(a.cfg.SearchPaths) == 0 {
		die("No search paths defined in %s", a.cfg.ConfigFile)
	}

	label := "Projects & Submodules"
	if a.layout.label != "basic" {
		label = "Projects (Profile: " + a.layout.label + ")"
	}

	var key, selected string
	for {
		input := withHeader(projectHeader, a.listProjects())
		var err error
		key, selected, err = picker.Run(input, true,
			"--expect=ctrl-r,ctrl-p,ctrl-x",
			"--no-sort",
			"--delimiter=\t",
			"--with-nth=2..",
			"--header-lines=1",
			"--border-label= 󰅍 "+label+" ",
			"--header=enter open · ctrl-r rebuild · ctrl-p rebuild with profile · ctrl-x kill session",
			"--prompt=󰝰 Project > ")
		if err != nil {
			die("%v", err)
		}
		target := picker.Field(selected)
		if key == "ctrl-x" && target != "" {
			name := session.Name(target)
			if tmux.IsManaged(name) {
				_ = tmux.Kill("=" + name)
			}
			continue
		}
		break
	}

	target := picker.Field(selected)
	if key == "ctrl-r" || key == "ctrl-p" {
		a.opts.Recreate = true
	}
	if key == "ctrl-p" && target != "" {
		if !a.pickProfile(session.Name(target)) {
			cancel()
		}
	}
	if target == "" {
		cancel()
	}
	return target
}
