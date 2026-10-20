package app

import (
	"os"

	"tds/internal/history"
	"tds/internal/picker"
)

func cmdHistory(a *App, args []string) {
	a.parseFlags(args)
	if len(a.opts.Positional) > 0 {
		die("history takes no directory argument.")
	}
	requireCmd("fzf")
	a.resolveLayout()

	if _, err := os.Stat(a.cfg.HistoryFile); err != nil {
		die("No history cache found.")
	}
	rows := a.projectRows(history.Read(a.cfg.HistoryFile), false, managedProfiles())
	_, selected, err := picker.Run(withHeader(projectHeader, rows), false,
		"--delimiter=\t",
		"--with-nth=2..",
		"--header-lines=1",
		"--header=enter open",
		"--prompt=󰉋 Recent Dev Projects > ")
	if err != nil {
		die("%v", err)
	}
	target := picker.Field(selected)
	if target == "" {
		cancel()
	}
	a.openDirectory(target)
}
