package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// precommit_e2e_fixtures_test.go — cross-platform plumbing shared by the
// pre-commit e2e tests (precommit_gate_scope_e2e_test.go,
// precommit_staged_read_e2e_test.go).
//
// The Windows CI legs (release-pr-multi-os run 42147372729, windows cli-3)
// failed every pre-commit e2e test with
//
//	error: cannot spawn .git/hooks/pre-commit: No such file or directory
//
// because the fixtures replaced the commit subprocess PATH with the
// colon-separated POSIX string "/usr/bin:/bin:/usr/sbin:/sbin" — on Windows
// that is one nonexistent directory entry, so git.exe could not resolve its
// bundled sh to spawn the hook at all. The helpers here build the PATH from
// real directories with os.PathListSeparator, and compile the test doubles
// into real binaries (moai.exe / gofmt.exe on Windows) instead of installing
// #!/bin/sh scripts as extension-less executables, which Windows cannot
// look up or exec.

// hookE2EPath builds the PATH for the commit subprocess: extraDir entries
// first (the fixture binary dirs), then the directories of the tools the
// REAL hook shells out to (git, sh, grep, sed, sort, tr, dirname), resolved
// from this process's PATH, then the classic POSIX directories where they
// exist. Deliberately NOT the ambient PATH: the tests must control whether
// moai (and the real gofmt) is reachable, so only named tool directories
// enter the list.
func hookE2EPath(t *testing.T, extraDirs ...string) string {
	t.Helper()
	var dirs []string
	add := func(d string) {
		if d == "" {
			return
		}
		for _, x := range dirs {
			if x == d {
				return
			}
		}
		dirs = append(dirs, d)
	}
	for _, d := range extraDirs {
		add(d)
	}
	for _, tool := range []string{"git", "sh", "grep", "sed", "sort", "tr", "dirname"} {
		if p, err := exec.LookPath(tool); err == nil {
			add(filepath.Dir(p))
		}
	}
	// POSIX fallbacks for macOS/Linux — included only where they exist;
	// the hardcoded list does not exist on Windows and previously left the
	// hook unspawnable there.
	for _, d := range []string{"/usr/bin", "/bin", "/usr/sbin", "/sbin"} {
		if fi, err := os.Stat(d); err == nil && fi.IsDir() {
			add(d)
		}
	}
	return strings.Join(dirs, string(os.PathListSeparator))
}

// buildE2EBinary compiles src (a package main) into a real binary named
// name inside binDir and returns binDir's path. The cross-platform
// replacement for writing a #!/bin/sh script as an extension-less
// executable: on Windows the binary gets the .exe suffix Windows exec
// lookups require, on other platforms it keeps the bare name.
func buildE2EBinary(t *testing.T, src, binDir, name string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not on PATH; the e2e binary fixture cannot be built")
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module e2efixture\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(binDir, name), ".")
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(), "GO111MODULE=on", "GOPROXY=off", "GOTOOLCHAIN=local", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build fixture %s: %v\n%s", name, err, out)
	}
	return binDir
}
