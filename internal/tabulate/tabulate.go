package tabulate

import (
	"strings"
	"unicode/utf8"
)

// Pad pads every column except the last to the widest cell in that column.
// Row element 0 is a key and is never padded.
func Pad(rows [][]string) [][]string {
	widths := map[int]int{}
	for _, row := range rows {
		for i := 1; i < len(row); i++ {
			if w := utf8.RuneCountInString(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		padded := []string{row[0]}
		for i := 1; i < len(row); i++ {
			cell := row[i]
			if i < len(row)-1 {
				cell += strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell))
			}
			padded = append(padded, cell)
		}
		out = append(out, padded)
	}
	return out
}

// Lines renders padded rows as tab-separated lines, keeping the key as the first field.
func Lines(rows [][]string) []string {
	padded := Pad(rows)
	lines := make([]string, 0, len(padded))
	for _, row := range padded {
		lines = append(lines, strings.Join(row, "\t"))
	}
	return lines
}

// Table renders rows (key excluded) as aligned text lines for plain output.
func Table(rows [][]string) []string {
	trimmed := make([][]string, 0, len(rows))
	for _, row := range rows {
		trimmed = append(trimmed, append([]string{""}, row...))
	}
	lines := make([]string, 0, len(trimmed))
	for _, row := range Pad(trimmed) {
		lines = append(lines, strings.TrimRight(strings.Join(row[1:], "  "), " "))
	}
	return lines
}
