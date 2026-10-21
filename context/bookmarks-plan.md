# Bookmarks

Let `tds` manage directories outside `search_paths` as first-class projects, stored in the cache.

## Storage

- File: `$XDG_CACHE_HOME/tds/bookmarks`, one absolute path per line, in insertion order.
- `config.Paths` gets `BookmarksFile`.
- `tds clean` clears only history. It never touches bookmarks.
- Paths are stored as `filepath.Abs` + `filepath.Clean`. Symlinks are not resolved, so the
  session name (md5 of the path) matches what `open` already produces.
- New package `internal/bookmarks`:
  - `Read(file) []string`
  - `Add(file, dir) error` writes a new line, with no duplicates.
  - `Remove(file, dir) (bool, error)` matches lexically, so a deleted directory can still be removed.
  - `Exists(dir) bool` checks with `os.Stat`.

## CLI

| Command | Behavior |
| --- | --- |
| `tds bookmark` | Open the bookmark picker. |
| `tds bookmark add [DIR]` | Add DIR (default: cwd). Must be an existing directory. Duplicates are a no-op with a message. A directory that is a direct child of a `search_paths` entry is refused with a message, because `find` already lists it. |
| `tds bookmark list` | Print path and status (`ok` or `missing`). |
| `tds bookmark remove <DIR>` | Remove by path. Accepts `.`. |

Dispatch: add `bookmark` to the command list in `main.go`. Flags `-p`, `-P`, `--recreate`, `--no-profile`,
`-c`, `-i`, `-e` apply to the picker, like `find`.

## Bookmark picker

Rows use the same columns as `find` (project, session, profile). Only bookmarks are listed.

| Key | Action |
| --- | --- |
| `enter` | Open the session. On a missing bookmark, show a notice and reopen the picker. |
| `ctrl-r` | Rebuild the session. |
| `ctrl-p` | Pick a profile, then rebuild. |
| `ctrl-x` | Kill the managed session. Keep the bookmark. |
| `ctrl-d` | Kill the managed session if one exists, remove the bookmark, reopen the picker. |

- Missing bookmarks are shown in gray, labeled as they are.
- Notices (missing path, error) go into the picker header for the next iteration only.
- Empty state: `No bookmarks. Use 'tds bookmark add' to add one.`
- Only sessions tagged `@tds_managed` are killed.

## `find` changes

- `listProjects` merges sorted `search_paths` children with bookmarks, in stored order,
  with duplicates removed. A bookmark that is a child of a search path is listed once.
- Missing bookmarks are gray. Enter on one shows a notice and reopens the picker, the same way the
  bookmark picker does.
- `find` keys are unchanged. `ctrl-d` is not added to `find`.
- `*` marks cwd when it is not in the merged set.
- `pickProject` no longer dies when `search_paths` is empty. It dies only when there are no
  search paths and no bookmarks, and the message points to `tds bookmark add`.

## Implementation notes

- Picker loop: extract the shared loop from `pickProject` into one function parameterized by rows,
  header, and key handlers. `find` and the bookmark picker both use it.
- fzf needs `--ansi` for the gray labels. Pad with `tabulate.Pad` first, then wrap the padded
  label in the color codes. Otherwise escape codes break column widths.
- Pure functions, so they can be tested without fzf or tmux:
  - `mergeProjects(children, bookmarks []string, here string) (ordered []string, cwdIsKnown bool)`
  - bookmark row builder that reports missing status.
- `doctor`: warn for each bookmark whose path is missing.
- `help.go` and `README.md`: document the command, the keys, and the storage path.

## Tests

- `internal/bookmarks`: add (new, duplicate, missing dir, under search path), remove (by `.`,
  by path of a deleted dir), read order, exists.
- `mergeProjects`: dedupe, cwd first, `*` marker when cwd is unknown, child-of-search-path dedupe.
- Row builder: missing flag and label.

## Out of scope

- Ordering or renaming bookmarks.
- Bookmarks for `history` or the yazi flow.
- A confirmation prompt for `ctrl-d`.

## Order of work

1. Config path + `internal/bookmarks` with tests.
2. `tds bookmark add|list|remove` with tests of the storage rules.
3. Shared picker loop, bookmark picker with all keys.
4. `find` merge, gray missing rows, `--ansi`, empty-state change.
5. `doctor`, help, README.
6. `go build`, `go vet`, `go test ./...`. Manual fzf and tmux checks are still needed, since they can't be run in tests.
