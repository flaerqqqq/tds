package app

import (
	"fmt"
	"os"
	"path/filepath"

	"tds/internal/session"
	"tds/internal/tabulate"
	"tds/internal/tmux"
)

func cmdList(args []string) {
	sessions := tmux.ListManaged()
	if len(sessions) == 0 {
		fmt.Println("No tds sessions.")
		return
	}
	rows := [][]string{{"SESSION", "WINDOWS", "STATE", "PATH"}}
	for _, s := range sessions {
		state := "detached"
		if s.Attached {
			state = "attached"
		}
		rows = append(rows, []string{s.Name, fmt.Sprint(s.Windows), state, s.Path})
	}
	for _, line := range tabulate.Table(rows) {
		fmt.Println(line)
	}
}

func cmdKill(args []string) {
	if len(args) == 0 || args[0] == "" {
		die("kill requires a session name or directory.")
	}
	name := args[0]
	if info, err := os.Stat(name); err == nil && info.IsDir() {
		abs, _ := filepath.Abs(name)
		name = session.Name(abs)
	}
	if !tmux.IsManaged(name) {
		die("'%s' is not a tds session.", name)
	}
	if err := tmux.Kill("=" + name); err != nil {
		die("%v", err)
	}
	fmt.Printf("Killed session '%s'.\n", name)
}

func cmdClean(a *App, args []string) {
	if _, err := os.Stat(a.cfg.HistoryFile); err != nil {
		fmt.Println("No history cache found.")
		return
	}
	if err := os.Remove(a.cfg.HistoryFile); err != nil {
		die("%v", err)
	}
	fmt.Println("Project history cache cleared successfully.")
}
