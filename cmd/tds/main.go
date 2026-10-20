package main

import (
	"fmt"
	"os"
	"os/exec"

	"tds/internal/app"
)

func main() {
	if _, err := exec.LookPath("tmux"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: Required tool 'tmux' is not installed.")
		os.Exit(1)
	}
	app.Main(os.Args[1:])
}
