package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// pathOnce keeps the login shell's PATH, asked for once.
var (
	pathOnce sync.Once
	pathDirs []string
)

// ShellPath is the PATH of the user's terminals: an app started from the
// Finder or the Dock has the system's alone, so the login shell is asked,
// as VS Code does. The app's own comes after it.
func ShellPath() []string {
	pathOnce.Do(func() {
		dirs := []string{}
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
			if runtime.GOOS == "darwin" {
				shell = "/bin/zsh"
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Between markers, as a shell's startup files may print too.
		out, err := exec.CommandContext(ctx, shell, "-ilc", `printf '\n@@%s@@\n' "$PATH"`).Output()
		if err == nil {
			if _, rest, ok := strings.Cut(string(out), "\n@@"); ok {
				if p, _, ok := strings.Cut(rest, "@@\n"); ok {
					dirs = filepath.SplitList(p)
				}
			}
		}
		for _, d := range filepath.SplitList(os.Getenv("PATH")) {
			if !Contains(dirs, d) {
				dirs = append(dirs, d)
			}
		}
		pathDirs = dirs
	})
	return pathDirs
}

// Path is the PATH of the user's terminals, as one string, for the
// environment of a command.
func Path() string {
	return strings.Join(ShellPath(), string(filepath.ListSeparator))
}

// Contains tells whether the list holds the string.
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
