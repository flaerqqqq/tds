# bm tree

A tree view of bookmarks, in the same picker as `find`, for working inside a multi-module project.

## Command

`tds bm tree` opens the tree picker. It accepts the same profile flags as `bm`.
`tds bm` (flat list) and `tds find` are unchanged.

## What is shown

- A bookmark is always shown.
- A `search_paths` project is shown only when at least one bookmark is below it.
- Common ancestors of two shown directories are shown as virtual roots, so siblings get a common parent.
  The following are never used as virtual roots: `/`, the home directory, and the `search_paths` directories.
- Directories between a node and its nearest shown ancestor are shown as virtual nodes, so the path
  to a nested bookmark is visible.
- Virtual nodes are gray. Missing bookmarks are gray, as in `find`.
- Unbookmarked subdirectories are not shown.

## Keys

The same as `find`, with `ctrl-d` deleting the bookmark when the selected row is one.
On a virtual or project row, `ctrl-d` shows the "Not a bookmark" notice.
`enter` on a virtual row opens that directory like any other project.

## Implementation

- `buildTree(projects, marks, skip)` is a pure function that returns `treeLine` values with drawing
  prefixes (`├─ `, `└─ `, `│  `). It is tested without fzf.
- `treeRows` converts the lines into picker rows. `pickFrom` runs the picker, with the same keys as `find`.
- `searchChildren` is extracted from `listProjects` and shared.

## Tests

- Single bookmark: one root.
- Project with a bookmark below: project is a root; the bookmark nests under it.
- Two bookmarks under a common directory: the directory is a virtual root.
- A bookmark two levels down: the intermediate directory is virtual.
- Skipped common ancestors do not appear.
