# tds

tmux dev session launcher. It opens a project directory as a tmux session named after the
project, builds a window layout from a JSON profile, and lets you pick, rebuild, switch and kill
sessions with fzf. Sessions that `tds` creates are tagged, and `tds` only manages those.

## Requirements

- Go 1.22+ (to build)
- tmux 3.2+ (`list-sessions -f`, `display-popup`)
- fzf
- Optional: yazi (`find --yazi`), VS Code (`--code`), IntelliJ IDEA (`--idea`), or any `--editor` command

## Install

```sh
git clone <repo> ~/dev/projects/tds   # or wherever you keep it
~/dev/projects/tds/install.sh          # builds tds and installs it to ~/.local/bin/tds
```

The result is a single binary with no runtime dependency on this checkout. Copy the binary or
run `install.sh` on each machine. Make sure `~/.local/bin` is on your `PATH`.

To remove it:

```sh
~/dev/projects/tds/install.sh --uninstall
```

This deletes `~/.local/bin/tds` only if it was installed from this checkout. The repo and your
config stay in place.

## Usage

```
tds [command] [options]
```

| Command | Purpose |
| --- | --- |
| `open [DIR]` | Open DIR (default: current directory). This is used when no command is given. |
| `find` | Pick a project from `search_paths` with fzf. `find --yazi` picks a directory in yazi. |
| `history` | Pick from recently opened directories. |
| `bm` | Pick a bookmarked directory with fzf. |
| `bm add [DIR]` | Bookmark DIR (default: current directory). |
| `bm list` | List bookmarks with `ok` or `missing` status. |
| `bm tree` | Pick from bookmarks shown as a tree under their parent directories. |
| `bm remove <DIR>` | Remove a bookmark. |
| `sessions` | Pick a managed tmux session. Enter switches to it. |
| `list` | List managed sessions. |
| `kill <name\|dir>` | Kill a managed session by exact name, or the session for a directory. |
| `doctor` | Check dependencies, config, profiles and managed sessions. Exits 1 on errors. |
| `clean` | Clear the recent-history file. |
| `help` | Show the built-in help. |

### Options for `open`, `find`, `history` and `sessions`

| Flag | Behavior |
| --- | --- |
| `-p, --profile <name>` | Use `profiles/<name>.json`. Overrides `default_profile`. |
| `-P, --select-profile` | Pick the profile with fzf. Rebuilds the session if it exists. |
| `--no-profile` | Use the single-window layout, ignoring profiles. |
| `--recreate` | Kill and rebuild the session so a changed profile takes effect. Without `-p`, it reuses the profile recorded on the session. |
| `-c`, `-i`, `-e <cmd>` | Launch VS Code, IntelliJ IDEA, or any command on the directory, in the background. Only when the session is created. |

Flags go after the command: `tds find -p dev`. Commands that pick a directory themselves
(`find`, `history`, `sessions`) reject a directory argument; use `tds open DIR` for that.

### Examples

```sh
tds                          # session for the current directory
tds ./my-app -p backend      # open a directory with the backend profile
tds find -p dev              # pick a project, use the dev profile
tds find --yazi              # pick a directory in yazi
tds sessions                 # pick a running session
tds kill .                   # kill the session for the current directory
tds find -p dev --recreate   # rebuild the picked project's session with dev
```

### fzf pickers

`find` and `history` show projects. The current directory is always the first row, next to the
prompt and preselected. If it isn't under `search_paths`, it is marked with `*` and shown with its path.
`find` also lists bookmarks after the `search_paths` projects. A bookmark whose directory is gone
is shown in gray; `enter` on it shows a notice and keeps the picker open.

| Key | Action |
| --- | --- |
| `enter` | Open the project. |
| `ctrl-r` | Rebuild the session (creates it if missing). |
| `ctrl-p` | Pick a profile, then rebuild. |
| `ctrl-x` | Kill the project's session and reopen the list. |
| `ctrl-d` | On a bookmark: kill its managed session if any, delete the bookmark, and reopen the list. On any other project it shows a notice. |

`bm` shows only bookmarks, with the same keys.

`sessions` shows managed sessions, the current session first, then most recently active. Each row
shows the project, session name, profile and last activity. `enter` switches to the session instead
of opening a project. `ctrl-r`, `ctrl-p` and `ctrl-x` work as above.

Profile pickers list `[Current: …]` (the session's profile, when it exists), `[Default: …]`, then every
profile in the profiles directory, and `basic`.

### Bookmarks

A bookmark is a directory outside `search_paths` that should appear in `find`. Bookmarks are
stored in `$XDG_CACHE_HOME/tds/bookmarks` (default `~/.cache/tds/bookmarks`), one absolute path per
line. `tds clean` does not remove them.

```sh
cd ~/scratch/one-off
tds bm add         # bookmark the current directory
tds bm add ~/work/client-x
tds bm list
tds bm remove ~/work/client-x
```

A directory that is already directly under a `search_paths` entry is not bookmarked, because
`find` already lists it.

`tds bm tree DIR` shows only the tree that contains DIR and puts the cursor on its nearest node. If DIR
is not in any tree, only DIR is shown. `tds bm tree --session-scope` does the same for the current tmux
session's directory. Without either flag it shows the whole tree. A `search_paths` project appears only when it has a
bookmark below it. Directories between a root and a bookmark, and a common parent of two bookmarks,
appear in gray so the tree stays connected. The keys are the same as `find`.

## Configuration

Config is read from `$XDG_CONFIG_HOME/tds` (default `~/.config/tds`). Created on first run.

`config.json`:

```json
{
  "default_profile": "dev",
  "search_paths": ["~/dev/projects"]
}
```

- `search_paths`: directories whose immediate children are projects. `~` and environment
  variables are expanded.
- `default_profile`: profile used when `-p` is not given. If it's missing or invalid, `tds`
  warns and uses the single-window layout.

### Profiles

Profiles live in `profiles/<name>.json` next to `config.json`. A profile is a list of windows,
created in order. The first window creates the session, and `command` (optional) runs in it.

```json
{
  "windows": [
    { "name": "󰈚 Code", "command": "nvim ." },
    { "name": "󰞷 Terminal", "command": "" },
    { "name": "󰚩 AI", "command": "" }
  ]
}
```

A profile must have a `windows` array. Profiles only apply when a session is created or
rebuilt, so editing one has no effect on a running session until you use `--recreate`, `ctrl-r`
or `ctrl-p`.

Sessions created without a profile are tagged `basic` and use a single window named `Terminal`.

## How sessions work

- **Name:** `<basename with dots replaced by _>-<first 4 hex chars of md5(absolute path)>`. Two
  directories with the same name get different sessions.
- **Tags:** each session `tds` builds gets `@tds_managed=1` and `@tds_profile=<name|basic>`.
  Every `tds` lookup (pickers, `list`, `kill`, labels, `doctor`) only looks at tagged
  sessions. Sessions made outside `tds`, such as your own `tmux new-session`, are ignored.
- **Adopting an existing session:** `tmux set-option -t <name>: @tds_managed 1`.
- **Inside tmux:** `tds` switches the client to the session. **Outside tmux:** it attaches.
  If no tmux server is running, it starts one.
- **Rebuilding:** `--recreate`, `ctrl-r` and `ctrl-p` kill the session and build it again. This
  stops everything running in it. If you rebuild the session you are in, `tds` builds a copy
  named `<name>-rebuild`, switches to it, and renames it back.
- **Popups:** the `prefix + f` and `prefix + F` tmux bindings open `tds find` and `tds sessions`
  in a popup. `tds` reads the session's `session_path` for the current directory when it runs in a
  popup.

## Project layout

```
cmd/tds/main.go                 entry point
internal/app/                  commands: open, find, history, sessions, list, kill, doctor, clean, help
internal/config/               config.json, search paths, default profile
internal/profile/              profile loading and validation
internal/session/              session naming (must match the bash naming)
internal/tmux/                 thin wrapper around the tmux CLI
internal/picker/               fzf runner
internal/tabulate/             column alignment
internal/history/              recent-directory history
legacy/                        the previous bash implementation, kept for reference
```

Files in `$XDG_CACHE_HOME/tds`: `history` (newest first, max 50 entries).

## Troubleshooting

Run `tds doctor`. It checks that tmux and fzf are installed, that `config.json` and every
profile parse, that `search_paths` and `default_profile` exist, and that each managed session
still has its directory and profile. It prints one `OK`, `WARN` or `ERR` line per check and
exits 1 on any `ERR`.

## Development

```sh
go test ./...
go vet ./...
```

Anything that creates sessions should run on a separate tmux server, so your real server is
never touched:

```sh
env -u TMUX TMUX_TMPDIR=<tmpdir> tmux new-session -d -s iso
```

Pass `TMUX` to `tds` as the socket path of that server (`#{socket_path},#{pid},0`) when testing
the in-tmux paths. Use a temporary `XDG_CONFIG_HOME` and `XDG_CACHE_HOME` to keep your real
config and history untouched.
