package tmux

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ManagedFilter selects sessions tagged by tds.
const ManagedFilter = "#{==:#{@tds_managed},1}"

type Session struct {
	Name     string
	Profile  string
	Path     string
	Activity int64
	Windows  int
	Attached bool
}

func run(args ...string) (string, error) {
	var out bytes.Buffer
	cmd := exec.Command("tmux", args...)
	cmd.Stdout = &out
	err := cmd.Run()
	return strings.TrimRight(out.String(), "\n"), err
}

func interactive(args ...string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InTmux() bool { return os.Getenv("TMUX") != "" }

// ServerRunning mirrors the bash check `pgrep tmux`.
func ServerRunning() bool { return exec.Command("pgrep", "tmux").Run() == nil }

func ListManaged() []Session {
	out, err := run("list-sessions", "-f", ManagedFilter, "-F",
		"#{session_name}\t#{@tds_profile}\t#{session_path}\t#{session_activity}\t#{session_windows}\t#{session_attached}")
	if err != nil {
		return nil
	}
	var list []Session
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\t")
		if len(f) < 6 || f[0] == "" {
			continue
		}
		activity, _ := strconv.ParseInt(f[3], 10, 64)
		windows, _ := strconv.Atoi(f[4])
		list = append(list, Session{
			Name:     f[0],
			Profile:  f[1],
			Path:     f[2],
			Activity: activity,
			Windows:  windows,
			Attached: f[5] != "" && f[5] != "0",
		})
	}
	return list
}

func IsManaged(name string) bool {
	for _, s := range ListManaged() {
		if s.Name == name {
			return true
		}
	}
	return false
}

func HasSession(name string) bool {
	_, err := run("has-session", "-t", "="+name)
	return err == nil
}

func CurrentSession() string {
	out, _ := run("display-message", "-p", "#S")
	return out
}

// ActiveSessionInfo describes the session of the client that is currently active.
func ActiveSessionInfo() (managed bool, path string) {
	out, _ := run("display-message", "-p", "#{@tds_managed}\t#{session_path}")
	f := strings.SplitN(out, "\t", 2)
	if len(f) != 2 {
		return false, ""
	}
	return f[0] == "1", f[1]
}

func Option(name, key string) string {
	out, _ := run("display-message", "-p", "-t", "="+name+":", "#{"+key+"}")
	return out
}

func SessionPath(name string) string {
	out, _ := run("display-message", "-p", "-t", "="+name+":", "#{session_path}")
	return out
}

func SetOption(name, key, value string) error {
	_, err := run("set-option", "-t", name+":", key, value)
	return err
}

func NewSession(name, dir, window, command string) error {
	args := []string{"new-session", "-d", "-s", name, "-c", dir, "-n", window}
	if command != "" {
		args = append(args, command)
	}
	_, err := run(args...)
	return err
}

func NewWindow(name string, index int, dir, window, command string) error {
	args := []string{"new-window", "-t", name + ":" + strconv.Itoa(index), "-n", window, "-c", dir}
	if command != "" {
		args = append(args, command)
	}
	_, err := run(args...)
	return err
}

func SelectWindow(name string, index int) error {
	_, err := run("select-window", "-t", name+":"+strconv.Itoa(index))
	return err
}

func Kill(target string) error {
	_, err := run("kill-session", "-t", target)
	return err
}

func Rename(target, newName string) error {
	_, err := run("rename-session", "-t", target, newName)
	return err
}

func SwitchClient(target string) error {
	_, err := run("switch-client", "-t", target)
	return err
}

func Attach(target string) error {
	return interactive("attach-session", "-t", target)
}
