package picker

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// Run pipes input to fzf. With expectKeys, fzf prints the pressed key on the first line
// and the selection on the second. Cancelling fzf gives empty results, not an error.
func Run(input string, expectKeys bool, args ...string) (key, selected string, err error) {
	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return "", "", err
		}
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if expectKeys {
		if len(lines) > 1 {
			return lines[0], lines[1], nil
		}
		return lines[0], "", nil
	}
	return "", lines[0], nil
}

// Field returns the first tab-separated field of a picker line.
func Field(line string) string {
	return strings.SplitN(line, "\t", 2)[0]
}
