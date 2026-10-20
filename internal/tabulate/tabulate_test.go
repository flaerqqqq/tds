package tabulate

import "testing"

func TestTablePadsColumnsExceptLast(t *testing.T) {
	rows := [][]string{
		{"SESSION", "WINDOWS", "PATH"},
		{"a", "5", "/x"},
		{"longer-name", "12", "/y"},
	}
	got := Table(rows)
	want := []string{
		"SESSION      WINDOWS  PATH",
		"a            5        /x",
		"longer-name  12       /y",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPadCountsRunes(t *testing.T) {
	rows := [][]string{{"", "é", "x"}, {"", "ab", "y"}}
	got := Pad(rows)
	if got[0][1] != "é " || got[1][1] != "ab" {
		t.Errorf("unexpected padding: %q %q", got[0][1], got[1][1])
	}
}
