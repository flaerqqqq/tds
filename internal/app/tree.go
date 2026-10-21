package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tds/internal/bookmarks"
)

// treeLine is one row of the bookmark tree. Prefix holds the drawing characters before the name.
type treeLine struct {
	Dir      string
	Prefix   string
	Name     string
	Root     bool
	Bookmark bool
	Virtual  bool
}

type treeNode struct {
	dir      string
	bookmark bool
	virtual  bool
	parent   *treeNode
	kids     []*treeNode
}

// buildTree lays out bookmarks and the projects that contain them. Common ancestors of shown
// directories become virtual nodes unless they are in skip. Directories between a node and its
// nearest shown ancestor become virtual nodes too.
func buildTree(projects, marks, skip []string) []treeLine {
	nodes := map[string]*treeNode{}
	for _, m := range marks {
		nodes[m] = &treeNode{dir: m, bookmark: true}
	}
	for _, p := range projects {
		if _, ok := nodes[p]; !ok && hasBookmarkBelow(p, marks) {
			nodes[p] = &treeNode{dir: p}
		}
	}

	skipped := map[string]bool{}
	for _, s := range skip {
		skipped[s] = true
	}
	known := make([]string, 0, len(nodes))
	for d := range nodes {
		known = append(known, d)
	}
	for i := range known {
		for j := i + 1; j < len(known); j++ {
			lca := commonAncestor(known[i], known[j])
			if _, ok := nodes[lca]; !ok && !skipped[lca] {
				nodes[lca] = &treeNode{dir: lca, virtual: true}
			}
		}
	}

	dirs := make([]string, 0, len(nodes))
	for d := range nodes {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	for _, d := range dirs {
		var gaps []string
		var anc *treeNode
		for cur := d; ; {
			up := filepath.Dir(cur)
			if up == cur {
				break
			}
			if a, ok := nodes[up]; ok {
				anc = a
				break
			}
			gaps = append(gaps, up)
			cur = up
		}
		if anc == nil {
			continue
		}
		parent := anc
		for i := len(gaps) - 1; i >= 0; i-- {
			g, ok := nodes[gaps[i]]
			if !ok {
				g = &treeNode{dir: gaps[i], virtual: true}
				nodes[gaps[i]] = g
			}
			link(g, parent)
			parent = g
		}
		link(nodes[d], parent)
	}

	var roots []*treeNode
	for _, d := range dirs {
		n := nodes[d]
		if n.parent == nil {
			roots = append(roots, n)
		}
		sort.Slice(n.kids, func(i, j int) bool { return n.kids[i].dir < n.kids[j].dir })
	}

	var out []treeLine
	for _, r := range roots {
		emitTree(r, "", "", true, &out)
	}
	return out
}

func link(child, parent *treeNode) {
	if child.parent != nil {
		return
	}
	child.parent = parent
	parent.kids = append(parent.kids, child)
}

func emitTree(n *treeNode, head, cont string, root bool, out *[]treeLine) {
	*out = append(*out, treeLine{
		Dir:      n.dir,
		Prefix:   head,
		Name:     filepath.Base(n.dir),
		Root:     root,
		Bookmark: n.bookmark,
		Virtual:  n.virtual,
	})
	for i, k := range n.kids {
		if i == len(n.kids)-1 {
			emitTree(k, cont+"└─ ", cont+"   ", false, out)
		} else {
			emitTree(k, cont+"├─ ", cont+"│  ", false, out)
		}
	}
}

func hasBookmarkBelow(dir string, marks []string) bool {
	for _, m := range marks {
		if strings.HasPrefix(m, dir+"/") {
			return true
		}
	}
	return false
}

// commonAncestor returns the deepest directory containing both absolute paths.
func commonAncestor(a, b string) string {
	pa, pb := strings.Split(a, "/"), strings.Split(b, "/")
	var common []string
	for i := 0; i < len(pa) && i < len(pb) && pa[i] == pb[i]; i++ {
		common = append(common, pa[i])
	}
	if s := strings.Join(common, "/"); s != "" {
		return s
	}
	return "/"
}

func (a *App) treeSkip() []string {
	skip := append([]string{"/"}, a.cfg.SearchPaths...)
	if home, err := os.UserHomeDir(); err == nil {
		skip = append(skip, home)
	}
	return skip
}

// treeLines returns the whole tree, or with scope set, only the tree that contains scope. When
// scope is not inside any tree, it returns just scope as a single root, and the second result is false.
func (a *App) treeLines(scope string) ([]treeLine, bool) {
	marks := bookmarks.Read(a.cfg.BookmarksFile)
	lines := buildTree(a.searchChildren(), marks, a.treeSkip())
	if scope == "" {
		return lines, true
	}
	sub, ok := subtree(lines, scope)
	if !ok {
		return []treeLine{{Dir: scope, Name: filepath.Base(scope), Root: true}}, false
	}
	return sub, true
}

// subtree returns the root blocks related to scope: roots that contain it, or that sit inside it.
func subtree(lines []treeLine, scope string) ([]treeLine, bool) {
	var out []treeLine
	selected := false
	for i := 0; i < len(lines); i++ {
		if !lines[i].Root {
			continue
		}
		if !within(scope, lines[i].Dir) && !within(lines[i].Dir, scope) {
			continue
		}
		selected = true
		end := i + 1
		for end < len(lines) && !lines[end].Root {
			end++
		}
		out = append(out, lines[i:end]...)
		i = end - 1
	}
	return out, selected
}

// within reports whether dir is path or inside it.
func within(dir, path string) bool {
	return dir == path || strings.HasPrefix(dir, path+"/")
}

// nearestLine returns the deepest line whose directory contains scope, or -1.
func nearestLine(lines []treeLine, scope string) string {
	best := ""
	for _, l := range lines {
		if within(scope, l.Dir) && len(l.Dir) > len(best) {
			best = l.Dir
		}
	}
	return best
}

func (a *App) treeRows(lines []treeLine) ([][]string, []bool) {
	dirs := make([]string, len(lines))
	for i, l := range lines {
		dirs[i] = l.Dir
	}
	rows := a.projectRows(dirs, false, managedProfiles())
	dim := make([]bool, len(lines))
	for i, l := range lines {
		name := fitLabel(l.Name)
		if l.Root {
			name = fitLabel(tilde(l.Dir))
		}
		if l.Bookmark {
			name = bookmarkIcon + " " + name
		}
		rows[i][1] = l.Prefix + name
		dim[i] = l.Virtual || !bookmarks.Exists(l.Dir)
	}
	return rows, dim
}

func cmdBookmarkTree(a *App, args []string) {
	sessionScope := false
	var rest []string
	for _, arg := range args {
		if arg == "--session-scope" {
			sessionScope = true
		} else {
			rest = append(rest, arg)
		}
	}
	a.parseFlags(rest)
	if len(a.opts.Positional) > 1 {
		die("bm tree takes at most one directory.")
	}
	if sessionScope && len(a.opts.Positional) == 1 {
		die("--session-scope cannot be combined with a directory.")
	}
	scope := ""
	if len(a.opts.Positional) == 1 {
		abs, err := bookmarks.Normalize(a.opts.Positional[0])
		if err != nil {
			die("%v", err)
		}
		scope = abs
	} else if sessionScope {
		scope = currentDir()
	}
	if len(bookmarks.Read(a.cfg.BookmarksFile)) == 0 {
		printNoBookmarks()
		return
	}
	a.resolveLayout()

	lines, found := a.treeLines(scope)
	focus := ""
	if found && scope != "" {
		focus = nearestLine(lines, scope)
	}

	rows := func() ([][]string, []bool) {
		if scope == "" {
			all, _ := a.treeLines("")
			return a.treeRows(all)
		}
		scopedLines, ok := a.treeLines(scope)
		if found && !ok {
			return nil, nil
		}
		return a.treeRows(scopedLines)
	}

	target := a.pickFrom(pickerSpec{
		border:   " Bookmark tree ",
		prompt:   "Tree > ",
		help:     "enter open · ^r rebuild · ^p profile · ^x kill · ^d delete bookmark",
		rows:     rows,
		onDelete: a.removeBookmark,
		topDown:  true,
		focus:    focus,
	})
	a.openDirectory(target)
}
