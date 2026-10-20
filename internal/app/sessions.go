package app

import (
	"tds/internal/picker"
	"tds/internal/tmux"
)

func cmdSessions(a *App, args []string) {
	a.parseFlags(args)
	if len(a.opts.Positional) > 0 {
		die("sessions takes no directory argument.")
	}
	requireCmd("fzf")
	a.resolveLayout()

	var key, selected string
	for {
		var err error
		key, selected, err = picker.Run(withHeader(sessionHeader, a.sessionRows()), true,
			"--expect=ctrl-r,ctrl-p,ctrl-x",
			"--delimiter=\t",
			"--with-nth=2..",
			"--header-lines=1",
			"--border-label= 󰖲 Sessions ",
			"--header=enter switch · ctrl-r recreate · ctrl-p recreate with profile · ctrl-x kill",
			"--prompt=󰖯 Session > ")
		if err != nil {
			die("%v", err)
		}
		pick := picker.Field(selected)
		if key == "ctrl-x" && pick != "" {
			_ = tmux.Kill("=" + pick)
			continue
		}
		break
	}

	pick := picker.Field(selected)
	if pick == "" {
		cancel()
	}
	if key == "ctrl-r" || key == "ctrl-p" {
		a.opts.Recreate = true
		if key == "ctrl-p" && !a.pickProfile(pick) {
			cancel()
		}
		a.openDirectory(tmux.SessionPath(pick))
		return
	}
	if tmux.InTmux() {
		if err := tmux.SwitchClient("=" + pick); err != nil {
			die("%v", err)
		}
		return
	}
	a.attach(pick)
}
