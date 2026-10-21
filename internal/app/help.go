package app

import (
	"fmt"

	"tds/internal/config"
)

func printHelp() {
	paths := config.DefaultPaths()
	fmt.Printf(`Usage: tds [command] [options]

Commands:
  open [DIR]        Open DIR (default: current directory). Used when no command is given.
  find              Pick a project from the search paths with fzf.
  history           Pick a project from recent history with fzf.
  bm                Pick a bookmarked directory with fzf.
  bm add [DIR]      Bookmark DIR (default: current directory).
  bm list           List bookmarks and whether they exist.
  bm remove <DIR>   Remove a bookmark.
  sessions          Pick a managed tmux session with fzf.
  list              List managed tmux sessions.
  kill <name|dir>   Kill a managed session by name, or the session for a directory.
  doctor            Check tmux, fzf, config, profiles and managed sessions.
  clean             Clear the recent history.
  help              Show this message.

Open options (open, find, history, sessions):
  -p, --profile <name>  Use a window layout profile. Overrides default_profile.
  -P, --select-profile  Pick the profile with fzf. Rebuilds the session if it exists.
      --no-profile      Ignore profiles and use the single-window layout.
      --recreate        Kill and rebuild the session so a changed profile takes effect.
  -c, --code            Open the project in VS Code.
  -i, --idea            Open the project in IntelliJ IDEA.
  -e, --editor <cmd>    Open the project with any command.

find options:
  -y, --yazi            Pick the directory in yazi instead of fzf.

Configuration:
  Config:   %s
  Profiles: %s

Examples:
  tds                       Open the current directory
  tds ./my-app -p backend   Open a directory with the backend profile
  tds find -p dev           Pick a project and use the dev profile
  tds find -y               Pick a directory with yazi
  tds history               Pick from recent projects
  tds bm add                Bookmark the current directory
  tds bm                    Pick a bookmark
  tds sessions              Pick a running session
  tds kill .                Kill the session for the current directory
  tds find -p dev --recreate
`, paths.ConfigFile, paths.ProfilesDir)
}
