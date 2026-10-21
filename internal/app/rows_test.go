package app

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestAge(t *testing.T) {
	now := int64(100000)
	cases := map[int64]string{
		now - 10:    "just now",
		now - 300:   "5m ago",
		now - 7200:  "2h ago",
		now - 90000: "1d ago",
	}
	for then, want := range cases {
		if got := age(now, then); got != want {
			t.Errorf("age(%d) = %q, want %q", now-then, got, want)
		}
	}
}

func TestTilde(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	if got := tilde("/home/u/demo"); got != "~/demo" {
		t.Errorf("tilde(demo) = %q", got)
	}
	if got := tilde("/home/u"); got != "/home/u" {
		t.Errorf("tilde(home) = %q, want absolute home", got)
	}
	if got := tilde("/srv/x"); got != "/srv/x" {
		t.Errorf("tilde(srv) = %q", got)
	}
}

func TestFitLabelKeepsShortLabels(t *testing.T) {
	if got := fitLabel("~/dev/app"); got != "~/dev/app" {
		t.Errorf("fitLabel = %q", got)
	}
}

func TestFitLabelDropsMiddleDirectories(t *testing.T) {
	in := "~/work/clients/acme/services/billing/very-long-project-name"
	got := fitLabel(in)
	if utf8.RuneCountInString(got) > labelWidth {
		t.Fatalf("fitLabel(%q) = %q, longer than %d", in, got, labelWidth)
	}
	if !strings.HasPrefix(got, "~/") || !strings.Contains(got, "/…/") || !strings.HasSuffix(got, "very-long-project-name") {
		t.Errorf("fitLabel = %q, want start, …, and end", got)
	}
}

func TestFitLabelElidesSingleLongSegment(t *testing.T) {
	in := strings.Repeat("a", 30) + "END" + strings.Repeat("b", 30)
	got := fitLabel(in)
	if utf8.RuneCountInString(got) != labelWidth || !strings.Contains(got, "…") {
		t.Errorf("fitLabel = %q", got)
	}
	if !strings.HasPrefix(got, "aaa") || !strings.HasSuffix(got, "bbb") {
		t.Errorf("fitLabel = %q, want both ends kept", got)
	}
}

func TestFitLabelKeepsFirstDirectoryOfAbsolutePath(t *testing.T) {
	got := fitLabel("/tmp/bml/clients/acme/services/billing/handlers/proj")
	if !strings.HasPrefix(got, "/tmp/…/") || !strings.HasSuffix(got, "/proj") {
		t.Errorf("fitLabel = %q", got)
	}
}
