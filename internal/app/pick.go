package app

import (
	"fmt"
	"os"

	"tds/internal/picker"
	"tds/internal/session"
	"tds/internal/tmux"
)

// pickerSpec describes one project picker. rows returns the rows and, aligned with them,
// which directories are missing. onDelete removes a bookmark and reports whether it existed.
type pickerSpec struct {
	border   string
	prompt   string
	help     string
	notice   string
	rows     func() ([][]string, []bool)
	onDelete func(dir string) bool
	topDown  bool
	focus    string
}

// pickFrom runs a project picker and returns the chosen directory. ctrl-r and ctrl-p set the
// recreate and profile options. ctrl-x kills the managed session. ctrl-d also deletes the
// bookmark, if the directory is one. The list stays open after both.
func (a *App) pickFrom(spec pickerSpec) string {
	requireCmd("fzf")

	notice := spec.notice
	focus := spec.focus
	var key, selected string
	for {
		rows, dim := spec.rows()
		if len(rows) == 0 {
			os.Exit(0)
		}
		layout := "--layout=default"
		if spec.topDown {
			layout = "--layout=reverse"
		}
		extra := []string{}
		if focus != "" {
			for i, r := range rows {
				if r[0] == focus {
					extra = append(extra, fmt.Sprintf("--bind=load:pos(%d)", i+1))
					break
				}
			}
			focus = ""
		}
		args := []string{
			"--expect=ctrl-r,ctrl-p,ctrl-x,ctrl-d",
			"--ansi",
			"--no-sort",
			layout,
			"--delimiter=\t",
			"--with-nth=2..",
			"--header-lines=1",
			"--border-label=" + spec.border,
			"--footer=" + spec.help,
			"--prompt=" + spec.prompt,
		}
		if notice != "" {
			args = append(args, "--header="+notice)
		}
		args = append(args, extra...)
		var err error
		key, selected, err = picker.Run(pickerInput(projectHeader, rows, dim), true, args...)
		if err != nil {
			die("%v", err)
		}
		target := picker.Field(selected)
		notice = ""
		if target == "" {
			break
		}
		if key == "ctrl-d" && !spec.onDelete(target) {
			notice = "Not a bookmark: " + tilde(target)
			continue
		}
		if key == "ctrl-x" || key == "ctrl-d" {
			name := session.Name(target)
			if tmux.IsManaged(name) {
				_ = tmux.Kill("=" + name)
			}
			continue
		}
		if _, err := os.Stat(target); err != nil {
			notice = "Missing directory: " + tilde(target)
			continue
		}
		break
	}

	target := picker.Field(selected)
	if target == "" {
		cancel()
	}
	if key == "ctrl-r" || key == "ctrl-p" {
		a.opts.Recreate = true
	}
	if key == "ctrl-p" && !a.pickProfile(session.Name(target)) {
		cancel()
	}
	return target
}
