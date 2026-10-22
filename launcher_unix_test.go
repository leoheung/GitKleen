//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLauncherScript(t *testing.T) {
	dir := t.TempDir()
	// A stand-in for GitKleen's executable, in a folder with a space, saying
	// what it was given.
	app := filepath.Join(dir, "My Git", "GitKleen")
	os.MkdirAll(filepath.Dir(app), 0o755)
	got := filepath.Join(dir, "got")
	os.WriteFile(app, []byte("#!/bin/sh\nprintf '%s' \"$1\" > '"+got+"'\n"), 0o755)
	cmd := filepath.Join(dir, "gitkleen")
	os.WriteFile(cmd, []byte(launcherScript(app)), 0o755)

	repo := filepath.Join(dir, "repo")
	os.MkdirAll(filepath.Join(repo, "sub"), 0o755)
	for arg, want := range map[string]string{"": repo, "sub": filepath.Join(repo, "sub"), ".": repo} {
		os.Remove(got)
		c := exec.Command(cmd)
		if arg != "" {
			c.Args = append(c.Args, arg)
		}
		c.Dir = repo
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("gitkleen %s: %v %s", arg, err, out)
		}
		var b []byte
		for i := 0; i < 50 && len(b) == 0; i++ { // the app is started in the background
			time.Sleep(20 * time.Millisecond)
			b, _ = os.ReadFile(got)
		}
		want, _ = filepath.EvalSymlinks(want)
		if g, _ := filepath.EvalSymlinks(string(b)); g != want {
			t.Errorf("gitkleen %q opened %q, want %q", arg, b, want)
		}
	}
	if !ours(cmd) {
		t.Error("the script isn't seen as GitKleen's")
	}
}

func TestInstallAndUninstall(t *testing.T) {
	home, _ := os.UserHomeDir()
	own, err := os.MkdirTemp(home, ".gitkleen-test-bin-")
	if err != nil {
		t.Skip("no folder to make in the home folder:", err)
	}
	defer os.RemoveAll(own)
	other := t.TempDir()
	old := shellPATH
	shellPATH = func() []string { return []string{other, own} }
	defer func() { shellPATH = old }()

	// The user's own folder before the system's.
	path, _, err := writeLauncher("/Applications/GitKleen.app/Contents/MacOS/GitKleen")
	if err != nil || path != filepath.Join(own, "gitkleen") {
		t.Fatalf("installed at %q: %v", path, err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "'/Applications/GitKleen.app/Contents/MacOS/GitKleen' \"$dir\"") {
		t.Errorf("script:\n%s", b)
	}
	// Another program's gitkleen stays, and is never overwritten.
	foreign := filepath.Join(other, "gitkleen")
	os.WriteFile(foreign, []byte("#!/bin/sh\necho another gitkleen\n"), 0o755)
	shellPATH = func() []string { return []string{own, other} }
	if _, _, err := writeLauncher("/x/GitKleen"); err != nil {
		t.Fatalf("reinstalling over our own: %v", err)
	}
	shellPATH = func() []string { return []string{other} }
	if _, _, err := writeLauncher("/x/GitKleen"); err == nil || !strings.Contains(err.Error(), foreign) {
		t.Errorf("overwrote another program's gitkleen: %v", err)
	}

	shellPATH = func() []string { return []string{own, other} }
	removed, err := removeLaunchers()
	if err != nil || len(removed) != 1 || removed[0] != path {
		t.Fatalf("removed %v: %v", removed, err)
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Error("removed another program's gitkleen")
	}
	if removed, _ := removeLaunchers(); len(removed) != 0 {
		t.Errorf("removed again: %v", removed)
	}
}

func TestLauncherDirPrefersTheUsual(t *testing.T) {
	home, _ := os.UserHomeDir()
	local := filepath.Join(home, ".local", "bin")
	if !writable(local) {
		t.Skip("no ~/.local/bin to write to")
	}
	tool, err := os.MkdirTemp(home, ".gitkleen-test-tool-")
	if err != nil {
		t.Skip(err)
	}
	defer os.RemoveAll(tool)
	if d, _ := launcherDir([]string{tool, local}); d != local {
		t.Errorf("chose %q over %q", d, local)
	}
}

func TestOursKnowsTheOldScript(t *testing.T) {
	p := filepath.Join(t.TempDir(), "gitkleen")
	os.WriteFile(p, []byte("#!/bin/sh\ndir=$(cd \"${1:-.}\" && pwd) || exit 1\n'/Applications/GitKleen.app/Contents/MacOS/GitKleen' \"$dir\" >/dev/null 2>&1 &\n"), 0o755)
	if !ours(p) {
		t.Error("the script of before the mark isn't seen as GitKleen's")
	}
}
