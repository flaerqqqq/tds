package app

import (
	"reflect"
	"testing"
)

func renderTree(lines []treeLine) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, l.Prefix+l.Dir)
	}
	return out
}

func TestTreeSingleBookmarkIsRoot(t *testing.T) {
	got := renderTree(buildTree(nil, []string{"/x/tool"}, nil))
	if want := []string{"/x/tool"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTreeProjectWithBookmarkBelowIsShownAsRoot(t *testing.T) {
	lines := buildTree([]string{"/d/tds", "/d/other"}, []string{"/d/tds/docs"}, nil)
	got := renderTree(lines)
	want := []string{"/d/tds", "└─ /d/tds/docs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	if lines[0].Virtual || lines[0].Bookmark {
		t.Errorf("project root should be neither virtual nor bookmark: %+v", lines[0])
	}
}

func TestTreeProjectWithoutBookmarksIsHidden(t *testing.T) {
	got := renderTree(buildTree([]string{"/d/empty"}, []string{"/x/tool"}, nil))
	if want := []string{"/x/tool"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTreeSiblingsGetVirtualRootAndNestedGap(t *testing.T) {
	marks := []string{"/w/p/module1", "/w/p/something/module2"}
	lines := buildTree(nil, marks, nil)
	got := renderTree(lines)
	want := []string{
		"/w/p",
		"├─ /w/p/module1",
		"└─ /w/p/something",
		"   └─ /w/p/something/module2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
	if !lines[0].Virtual || !lines[0].Root {
		t.Errorf("common parent should be a virtual root: %+v", lines[0])
	}
	if !lines[2].Virtual {
		t.Errorf("intermediate directory should be virtual: %+v", lines[2])
	}
	if !lines[1].Bookmark || !lines[3].Bookmark {
		t.Errorf("bookmarks not marked: %+v %+v", lines[1], lines[3])
	}
}

func TestTreeSkippedCommonAncestorIsNotShown(t *testing.T) {
	marks := []string{"/d/p/a/m1", "/d/p/b/m2"}
	got := renderTree(buildTree(nil, marks, []string{"/d/p"}))
	want := []string{"/d/p/a/m1", "/d/p/b/m2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestTreeSharedGapIsShownOnce(t *testing.T) {
	marks := []string{"/w/p/s/m1", "/w/p/s/m2"}
	got := renderTree(buildTree(nil, marks, nil))
	want := []string{
		"/w/p/s",
		"├─ /w/p/s/m1",
		"└─ /w/p/s/m2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCommonAncestor(t *testing.T) {
	cases := map[[2]string]string{
		{"/a/b/c", "/a/b/d"}: "/a/b",
		{"/a/b", "/a/b/c"}:   "/a/b",
		{"/a", "/b"}:         "/",
	}
	for in, want := range cases {
		if got := commonAncestor(in[0], in[1]); got != want {
			t.Errorf("commonAncestor(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

func TestSubtreeReturnsOnlyTheContainingRoot(t *testing.T) {
	lines := buildTree(nil, []string{"/w/p/module1", "/w/p/something/module2", "/x/tool"}, []string{"/"})
	got, found := subtree(lines, "/w/p/something")
	if !found {
		t.Fatal("scope inside a tree not found")
	}
	if want := []string{"/w/p", "├─ /w/p/module1", "└─ /w/p/something", "   └─ /w/p/something/module2"}; !reflect.DeepEqual(renderTree(got), want) {
		t.Fatalf("got %q, want %q", renderTree(got), want)
	}
}

func TestSubtreeOutsideAnyTreeReturnsAll(t *testing.T) {
	lines := buildTree(nil, []string{"/w/p/a", "/x/tool"}, []string{"/"})
	got, found := subtree(lines, "/nowhere")
	if found || len(got) != 0 {
		t.Fatalf("found=%v len=%d", found, len(got))
	}
}

func TestNearestLineFindsDeepestContainingNode(t *testing.T) {
	lines := buildTree(nil, []string{"/w/p/module1", "/w/p/something/module2"}, nil)
	if got := nearestLine(lines, "/w/p/something/module2/src"); got != "/w/p/something/module2" {
		t.Errorf("nearestLine = %q", got)
	}
	if got := nearestLine(lines, "/w/p/other"); got != "/w/p" {
		t.Errorf("nearestLine = %q", got)
	}
}

func TestSubtreeIncludesRootsInsideScope(t *testing.T) {
	lines := buildTree(nil, []string{"/w/p/module1", "/x/tool"}, []string{"/"})
	got, found := subtree(lines, "/w/p")
	if !found || !reflect.DeepEqual(renderTree(got), []string{"/w/p/module1"}) {
		t.Fatalf("found=%v got %q", found, renderTree(got))
	}
}

func TestSubtreeEmptyAfterScopedBookmarksAreGone(t *testing.T) {
	lines := buildTree(nil, []string{"/x/tool"}, []string{"/"})
	if got, found := subtree(lines, "/w/p"); found || len(got) != 0 {
		t.Fatalf("found=%v got %q", found, renderTree(got))
	}
}
