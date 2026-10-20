package session

import "testing"

// Expected values come from the bash implementation on the same machine.
func TestNameMatchesBash(t *testing.T) {
	cases := map[string]string{
		"/home/vitalii_v/dotfiles": "dotfiles-9a37",
		"/home/vitalii_v/demo":     "demo-30df",
	}
	for path, want := range cases {
		if got := Name(path); got != want {
			t.Errorf("Name(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestNameReplacesDots(t *testing.T) {
	if got := Name("/tmp/my.app"); got[:6] != "my_app" {
		t.Errorf("Name = %q, want prefix my_app", got)
	}
}
