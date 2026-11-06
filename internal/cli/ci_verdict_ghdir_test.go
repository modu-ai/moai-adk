package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeGhMain is the compiled stub's source: it prints the process working
// directory, which is what the F1 pin compares against cmd.Dir.
const fakeGhMain = `package main

import (
	"fmt"
	"os"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(wd)
}
`

// writeFakeGhDir builds the fake gh as a compiled executable (sync-audit F6)
// into a fresh temp directory and returns that directory; the caller prepends
// it to PATH, first, so no real gh shadows it. A binary is selectable on
// every release-gate OS (sync-audit F5): exec.LookPath on Windows only tries
// PATHEXT extensions, so the helper builds as gh.exe there and as an
// extension-less gh elsewhere, and it prints the process cwd identically on
// all three hosts instead of leaning on shell semantics.
func writeFakeGhDir(t *testing.T) string {
	t.Helper()
	fake := t.TempDir()
	src := filepath.Join(fake, "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir fake gh src: %v", err)
	}
	main := filepath.Join(src, "main.go")
	if err := os.WriteFile(main, []byte(fakeGhMain), 0o644); err != nil {
		t.Fatalf("write fake gh main.go: %v", err)
	}
	name := "gh"
	if runtime.GOOS == "windows" {
		name = "gh.exe"
	}
	build := exec.Command("go", "build", "-o", filepath.Join(fake, name), main)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build fake gh: %v: %s", err, out)
	}
	return fake
}

// fakeGhBinaryName names the fake gh executable the stub builder produces on
// this host.
func fakeGhBinaryName() string {
	if runtime.GOOS == "windows" {
		return "gh.exe"
	}
	return "gh"
}

// TestFakeGhHelperIsCompiledBinary pins the fake gh stub's format (sync-audit
// F6): the stub must be a compiled executable — PE, ELF, or Mach-O — not a
// #!/bin/sh script or a .bat body, so the F1 cwd pin does not depend on host
// shell semantics.
func TestFakeGhHelperIsCompiledBinary(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(writeFakeGhDir(t), fakeGhBinaryName()))
	if err != nil {
		t.Fatalf("read fake gh: %v", err)
	}
	head := body
	if len(head) > 4 {
		head = head[:4]
	}
	if bytes.HasPrefix(body, []byte("#!")) {
		t.Fatalf("fake gh stub is a shell script (leading bytes %q); want a compiled binary (sync-audit F6)", head)
	}
	if bytes.HasPrefix(body, []byte("@echo")) {
		t.Fatalf("fake gh stub is a .bat body (leading bytes %q); want a compiled binary (sync-audit F6)", head)
	}
	magics := [][]byte{
		[]byte("MZ"),             // PE (windows)
		{0x7f, 'E', 'L', 'F'},    // ELF (linux)
		{0xFE, 0xED, 0xFA, 0xCE}, // Mach-O 32, big-endian
		{0xCE, 0xFA, 0xED, 0xFE}, // Mach-O 32, little-endian
		{0xFE, 0xED, 0xFA, 0xCF}, // Mach-O 64, big-endian
		{0xCF, 0xFA, 0xED, 0xFE}, // Mach-O 64, little-endian
	}
	for _, magic := range magics {
		if bytes.HasPrefix(body, magic) {
			return
		}
	}
	t.Fatalf("fake gh stub has no known binary magic (leading bytes % x); want PE MZ, ELF, or Mach-O (sync-audit F6)", head)
}

// t1710 sync-audit F1 — defaultGhRunner scrubbed the environment but never
// pinned the working directory, so a cross-tree invocation (this factory's
// normal posture: a leader running ci-verdict --project-root <worktree> from
// the primary checkout) let gh resolve the repository from the PROCESS cwd and
// read another repository's checks as the landing gate's verdict. Reproduced
// by the auditor with a fake gh logging its cwd. The runner must run gh IN the
// resolved project root.
func TestDefaultGhRunnerRunsGhInTheProjectRoot(t *testing.T) {
	t.Setenv("PATH", writeFakeGhDir(t)+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := t.TempDir()
	// pwd reports the symlink-resolved path (/private prefix on darwin);
	// compare against the same normalization.
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	out, err := defaultGhRunner(root, "run", "list")
	if err != nil {
		t.Fatalf("defaultGhRunner: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != want && !strings.EqualFold(got, want) {
		t.Errorf("gh ran in %q, want the resolved project root %q (the positional vector: gh resolved the repo from the process cwd)", got, want)
	}
}
