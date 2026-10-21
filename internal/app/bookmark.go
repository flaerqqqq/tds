package app

import (
	"fmt"
	"strings"

	"tds/internal/bookmarks"
	"tds/internal/tabulate"
)

const bookmarkIcon = "󰃀"

func printNoBookmarks() {
	fmt.Println("No bookmarks. Use 'tds bm add' to add one.")
}

func cmdBookmark(a *App, args []string) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		cmdBookmarkPicker(a, args)
		return
	}
	switch args[0] {
	case "tree":
		cmdBookmarkTree(a, args[1:])
	case "add":
		bookmarkAdd(a, args[1:])
	case "list":
		bookmarkList(a)
	case "remove":
		bookmarkRemove(a, args[1:])
	default:
		die("Unknown bm command: %s (use add, list or remove).", args[0])
	}
}

func cmdBookmarkPicker(a *App, args []string) {
	a.parseFlags(args)
	if len(a.opts.Positional) > 0 {
		die("bm takes no directory argument; use 'tds bm add DIR'.")
	}
	if len(bookmarks.Read(a.cfg.BookmarksFile)) == 0 {
		printNoBookmarks()
		return
	}
	a.resolveLayout()

	target := a.pickFrom(pickerSpec{
		border:   " " + bookmarkIcon + " Bookmarks ",
		prompt:   bookmarkIcon + " Bookmark > ",
		help:     "enter open · ^r rebuild · ^p profile · ^x kill · ^d delete bookmark",
		rows:     a.bookmarkRows,
		onDelete: a.removeBookmark,
	})
	a.openDirectory(target)
}

func (a *App) removeBookmark(dir string) bool {
	removed, err := bookmarks.Remove(a.cfg.BookmarksFile, dir)
	if err != nil {
		die("%v", err)
	}
	return removed
}

// bookmarkRows returns the bookmarks in stored order, with a flag for each missing directory.
func (a *App) bookmarkRows() ([][]string, []bool) {
	marks := bookmarks.Read(a.cfg.BookmarksFile)
	dim := make([]bool, len(marks))
	for i, d := range marks {
		dim[i] = !bookmarks.Exists(d)
	}
	return a.projectRows(marks, true, managedProfiles()), dim
}

func bookmarkAdd(a *App, args []string) {
	if len(args) > 1 {
		die("bm add takes at most one directory.")
	}
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}
	if !bookmarks.Exists(dir) {
		die("Directory '%s' not found.", dir)
	}
	abs, err := bookmarks.Normalize(dir)
	if err != nil {
		die("%v", err)
	}
	if a.cfg.IsProjectDir(abs) {
		fmt.Printf("%s is already listed under search_paths.\n", abs)
		return
	}
	added, err := bookmarks.Add(a.cfg.BookmarksFile, abs)
	if err != nil {
		die("%v", err)
	}
	if !added {
		fmt.Printf("Already bookmarked: %s\n", abs)
		return
	}
	fmt.Printf("Bookmarked %s\n", abs)
}

func bookmarkList(a *App) {
	marks := bookmarks.Read(a.cfg.BookmarksFile)
	if len(marks) == 0 {
		fmt.Println("No bookmarks.")
		return
	}
	rows := [][]string{{"PATH", "STATUS"}}
	for _, d := range marks {
		status := "ok"
		if !bookmarks.Exists(d) {
			status = "missing"
		}
		rows = append(rows, []string{d, status})
	}
	for _, line := range tabulate.Table(rows) {
		fmt.Println(line)
	}
}

func bookmarkRemove(a *App, args []string) {
	if len(args) != 1 || args[0] == "" {
		die("bookmark remove requires a directory.")
	}
	abs, err := bookmarks.Normalize(args[0])
	if err != nil {
		die("%v", err)
	}
	removed, err := bookmarks.Remove(a.cfg.BookmarksFile, abs)
	if err != nil {
		die("%v", err)
	}
	if !removed {
		die("'%s' is not a bookmark.", abs)
	}
	fmt.Printf("Removed bookmark %s\n", abs)
}
