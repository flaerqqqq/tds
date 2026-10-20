package app

import "testing"

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
