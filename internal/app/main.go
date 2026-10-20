package app

import (
	"tds/internal/config"
)

// Main runs the CLI and returns nothing; commands exit the process on errors.
func Main(args []string) {
	cmd := "open"
	if len(args) > 0 {
		switch args[0] {
		case "open", "find", "history", "sessions", "list", "kill", "doctor", "clean", "help":
			cmd, args = args[0], args[1:]
		case "-h", "--help":
			cmd, args = "help", args[1:]
		}
	}
	if cmd == "help" {
		printHelp()
		return
	}

	cfg, err := config.Load(config.DefaultPaths())
	if err != nil {
		die("%v", err)
	}
	a := &App{cfg: cfg}

	switch cmd {
	case "open":
		a.parseFlags(args)
		if a.opts.Yazi {
			die("use 'tds find --yazi' to pick a directory with yazi.")
		}
		if len(a.opts.Positional) > 1 {
			die("open takes at most one directory argument.")
		}
		a.resolveLayout()
		target := ""
		if len(a.opts.Positional) == 1 {
			target = a.opts.Positional[0]
		}
		a.openDirectory(target)
	case "find":
		cmdFind(a, args)
	case "history":
		cmdHistory(a, args)
	case "sessions":
		cmdSessions(a, args)
	case "list":
		cmdList(args)
	case "kill":
		cmdKill(args)
	case "doctor":
		cmdDoctor(a, args)
	case "clean":
		cmdClean(a, args)
	}
}
