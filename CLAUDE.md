# tds

tmux dev session launcher written in Go. See README.md for user-facing behavior.

## Build and install

The installed `tds` binary is built from this checkout. After any code change, rebuild it before testing:

```sh
./install.sh            # builds with go and installs to ~/.local/bin/tds
```

`go test ./...` does not rebuild the installed binary. Run `./install.sh` again after every change you want to try from the shell.

## Checks

```sh
go vet ./...
go test ./...
gofmt -l internal cmd   # should print nothing
```

fzf pickers and tmux behavior are not covered by tests. Check them by hand, for example inside a detached tmux pane (`tmux -L <socket> new-session -d ...` and `capture-pane`).

## Layout

- `cmd/tds`: entry point.
- `internal/app`: commands, pickers, and the `find`/`history`/`sessions`/`bm` flows.
- `internal/config`: paths, `config.json`, search paths.
- `internal/bookmarks`: bookmark file (cache dir), add/remove/read.
- `internal/history`, `internal/profile`, `internal/session`, `internal/tmux`, `internal/picker`, `internal/tabulate`: supporting packages.
- `context/`: design plans for features. Check there before starting a feature.

## Conventions

- Commands exit through `die`/`cancel` in `internal/app/app.go`. Don't call `os.Exit` directly unless it matches the existing picker code.
- Session names come from `session.Name(path)`. It must stay compatible with the older bash implementation.
- Keep state files in the cache dir (`$XDG_CACHE_HOME/tds`). Config stays in `$XDG_CONFIG_HOME/tds`.
