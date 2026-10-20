package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tds/internal/config"
	"tds/internal/history"
	"tds/internal/picker"
	"tds/internal/profile"
	"tds/internal/session"
	"tds/internal/tmux"
)

const historyLimit = 50

type Options struct {
	ProfileName     string
	ProfileExplicit bool
	ProfilePick     bool
	NoProfile       bool
	Recreate        bool
	Yazi            bool
	Editor          string
	Positional      []string
}

// layout is the window set a new session is built from. label is stored on the session.
type layout struct {
	label   string
	windows []profile.Window
}

type App struct {
	cfg    config.Config
	opts   Options
	layout layout
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
}

func cancel() {
	fmt.Println("Cancelled.")
	os.Exit(0)
}

func requireCmd(name string) {
	if _, err := exec.LookPath(name); err != nil {
		die("%s is required for this command.", name)
	}
}

func basicLayout() layout { return layout{label: "basic"} }

func fromProfile(p profile.Profile) layout {
	return layout{label: p.Name, windows: p.Windows}
}

func (a *App) parseFlags(args []string) {
	a.opts = Options{}
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "-p", "--profile":
			if i+1 >= len(args) || args[i+1] == "" {
				die("--profile requires a profile name.")
			}
			i++
			a.opts.ProfileName = args[i]
			a.opts.ProfileExplicit = true
		case "-P", "--select-profile":
			a.opts.ProfilePick = true
		case "--no-profile":
			a.opts.NoProfile = true
		case "--recreate":
			a.opts.Recreate = true
		case "-y", "--yazi":
			a.opts.Yazi = true
		case "-c", "--code":
			a.opts.Editor = "code"
		case "-i", "--idea":
			a.opts.Editor = "idea"
		case "-e", "--editor":
			if i+1 >= len(args) || args[i+1] == "" {
				die("--editor requires a command.")
			}
			i++
			a.opts.Editor = args[i]
		case "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				die("Unknown option: %s (run 'tds help' for usage).", arg)
			}
			a.opts.Positional = append(a.opts.Positional, arg)
		}
	}
}

// resolveLayout picks the layout from -p, then default_profile, then the single window.
func (a *App) resolveLayout() {
	a.layout = basicLayout()
	if a.opts.NoProfile {
		return
	}
	if a.opts.ProfileExplicit {
		p, err := profile.Load(a.cfg.ProfilesDir, a.opts.ProfileName)
		if err != nil {
			die("%v", err)
		}
		a.layout = fromProfile(p)
		return
	}
	if a.cfg.DefaultProfile == "" {
		return
	}
	p, err := profile.Load(a.cfg.ProfilesDir, a.cfg.DefaultProfile)
	if err != nil {
		warn("default_profile '%s' is missing or invalid; using single-window layout.", a.cfg.DefaultProfile)
		return
	}
	a.layout = fromProfile(p)
}

func (a *App) profileNames() []string {
	seen := map[string]bool{"basic": true}
	files, _ := filepath.Glob(filepath.Join(a.cfg.ProfilesDir, "*.json"))
	for _, f := range files {
		seen[strings.TrimSuffix(filepath.Base(f), ".json")] = true
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// pickProfile asks which profile to use for the session name. It returns false when cancelled.
func (a *App) pickProfile(name string) bool {
	def := a.cfg.DefaultProfile
	if def == "" {
		def = "basic"
	}
	pref, marker := def, "Default"
	if tmux.IsManaged(name) {
		pref = tmux.Option(name, "@tds_profile")
		if pref == "" {
			pref = "basic"
		}
		marker = "Current"
	}

	var rows []string
	if marker == "Current" {
		rows = append(rows, fmt.Sprintf("%s\t[Current: %s]", pref, pref))
	}
	rows = append(rows, fmt.Sprintf("%s\t[Default: %s]", def, def))
	for _, p := range a.profileNames() {
		if p != pref && p != def {
			rows = append(rows, p+"\t"+p)
		}
	}

	_, selected, err := picker.Run(strings.Join(rows, "\n"), false,
		"--no-sort", "--delimiter=\t", "--with-nth=2..",
		"--prompt=󰒓 Profile > ", "--border-label= 󰒓 Profiles ")
	if err != nil {
		die("%v", err)
	}
	chosen := pickedField(selected)
	if chosen == "" {
		return false
	}
	a.opts.ProfileExplicit = true
	if chosen == "basic" {
		a.layout = basicLayout()
		return true
	}
	p, err := profile.Load(a.cfg.ProfilesDir, chosen)
	if err != nil {
		a.layout = basicLayout()
		return true
	}
	a.layout = fromProfile(p)
	return true
}

// openDirectory records the directory in history and opens its session.
// An empty target means the current directory.
func (a *App) openDirectory(target string) {
	selected, err := os.Getwd()
	if err != nil {
		die("%v", err)
	}
	if target != "" {
		info, err := os.Stat(target)
		if err != nil || !info.IsDir() {
			die("Directory '%s' not found.", target)
		}
		selected, _ = filepath.Abs(target)
	}
	if err := history.Add(a.cfg.HistoryFile, selected, historyLimit); err != nil {
		warn("could not update history: %v", err)
	}
	a.openSelected(selected)
}

func (a *App) openSelected(selected string) {
	name := session.Name(selected)

	if a.opts.ProfilePick && !a.opts.ProfileExplicit {
		a.opts.Recreate = true
		if !a.pickProfile(name) {
			cancel()
		}
	}

	// Not in tmux and no server running: start one with this session.
	if !tmux.InTmux() && !tmux.ServerRunning() {
		a.mustBuild(name, selected)
		a.attach(name)
		return
	}

	if a.opts.Recreate && tmux.HasSession(name) {
		if !a.opts.NoProfile && !a.opts.ProfileExplicit {
			a.layoutFromSession(name)
		}
		// Rebuilding the session we are in: build a copy, switch, then swap names.
		if tmux.InTmux() && tmux.CurrentSession() == name {
			a.rebuildInPlace(name, selected)
			return
		}
		_ = tmux.Kill("=" + name)
	}

	if !tmux.HasSession(name) {
		a.mustBuild(name, selected)
	}
	if tmux.InTmux() {
		_ = tmux.SwitchClient("=" + name)
	} else {
		a.attach(name)
	}
}

func (a *App) layoutFromSession(name string) {
	current := tmux.Option(name, "@tds_profile")
	switch {
	case current == "basic":
		a.layout = basicLayout()
	case current != "":
		p, err := profile.Load(a.cfg.ProfilesDir, current)
		if err != nil {
			warn("session profile '%s' is missing or invalid; using the default profile.", current)
			return
		}
		a.layout = fromProfile(p)
	}
}

func (a *App) rebuildInPlace(name, selected string) {
	rebuild := name + "-rebuild"
	old := name + "-old"
	a.mustBuild(rebuild, selected)
	_ = tmux.SwitchClient("=" + rebuild)
	_ = tmux.Rename("="+name, old)
	_ = tmux.Rename("="+rebuild, name)
	_ = tmux.Kill("=" + old)
}

func (a *App) mustBuild(name, dir string) {
	if err := a.build(name, dir); err != nil {
		die("could not create session '%s': %v", name, err)
	}
}

func (a *App) build(name, dir string) error {
	if len(a.layout.windows) == 0 {
		if err := tmux.NewSession(name, dir, "󰞷 Terminal", ""); err != nil {
			return err
		}
	}
	for i, w := range a.layout.windows {
		var err error
		if i == 0 {
			err = tmux.NewSession(name, dir, w.Name, w.Command)
		} else {
			err = tmux.NewWindow(name, i+1, dir, w.Name, w.Command)
		}
		if err != nil {
			return err
		}
	}
	if err := tmux.SetOption(name, "@tds_profile", a.layout.label); err != nil {
		return err
	}
	if err := tmux.SetOption(name, "@tds_managed", "1"); err != nil {
		return err
	}
	_ = tmux.SelectWindow(name, 1)
	if a.opts.Editor != "" {
		launchEditor(a.opts.Editor, dir)
	}
	return nil
}

// launchEditor starts the editor detached, so it outlives tds.
func launchEditor(command, dir string) {
	cmd := exec.Command("setsid", command, dir)
	if err := cmd.Start(); err == nil {
		time.Sleep(200 * time.Millisecond)
	}
}

func (a *App) attach(name string) {
	if err := tmux.Attach("=" + name); err != nil {
		die("could not attach to '%s': %v", name, err)
	}
}

func pickedField(line string) string {
	return strings.SplitN(line, "\t", 2)[0]
}
