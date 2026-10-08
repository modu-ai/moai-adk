// m0_collision_red_test.go — SPEC-USERASSET-DEPLOY-GUARD-001 M0, collision
// family: AC-010 (FIFO non-blocking precheck), AC-011 (confined write pinned
// to the validated parent), AC-025 (exec-bit preservation).
//
// M0 discipline: observation only — no production change.
package userassets

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"testing/fstest"
	"time"
)

// TestCollisionPrecheckSkipsFifoWithoutBlock — AC-010 (ledger 8c,
// REQ-COL-001). Given a FIFO at an install target, the collision
// determination must classify without blocking. RED-now reason: the precheck
// reads the target with os.ReadFile (install.go:211), which blocks forever on
// a FIFO with no writer. The command per acceptance.md carries
// `-timeout 30s`; this in-test watchdog makes the same observation
// deterministically without relying on the framework kill.
func TestCollisionPrecheckSkipsFifoWithoutBlock(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("FIFO semantics are unix — the windows axis of this family is the GOOS=windows build gate (B1)")
	}
	f := newFixture(t)
	fifoPath := filepath.Join(f.home, ".claude", "skills", "moai-alpha", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(fifoPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(fifoPath, 0o644); err != nil {
		t.Fatalf("mkfifo: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := f.installer(t).Install(nil)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("install returned an error on a FIFO target: %v", err)
		}
		// Completed without blocking — the contract held (GREEN shape).
	case <-time.After(10 * time.Second):
		t.Fatalf("RED (intended): the collision precheck blocked for over 10s on a FIFO target — os.ReadFile at install.go:211 reads the FIFO with no writer and never returns")
	}
}

// TestConfinedWritePinnedToValidatedParent — AC-011 (ledger 12,
// REQ-COL-002). Given the parent directory chain swapped to an outside-
// pointing symlink AFTER validation, the write must not follow the swapped
// parent: it lands pinned to the validated parent (or is refused). The AC
// names this an external-record probe against the current path-based code.
//
// Probe mechanics (the plan's determinization note): a watcher polls the
// validated parent for the confinedWrite temp file (.ua-write-*); the moment
// one appears — i.e. validation passed and the window :725→:750 is open — it
// swaps the mid-chain directory for a symlink into an external sentinel dir.
// The rename at :750 then resolves dest through the swapped parent. Up to
// 300 fresh attempts; a caught external write is the RED evidence.
func TestConfinedWritePinnedToValidatedParent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink-swap probe uses unix rename/symlink semantics — the windows axis is the GOOS=windows build gate (B1)")
	}
	const attempts = 60
	externalWrite := ""
	var writeErrs []string

	for i := 0; i < attempts && externalWrite == ""; i++ {
		ext, werr := attemptParentSwapProbe(t)
		if ext != "" {
			externalWrite = ext
			break
		}
		if werr != nil {
			writeErrs = append(writeErrs, werr.Error())
		}
	}
	if externalWrite == "" {
		summary := "no write errors observed"
		if len(writeErrs) > 0 {
			summary = "write errors: " + writeErrs[0]
			if len(writeErrs) > 1 {
				summary += fmt.Sprintf(" (+%d more)", len(writeErrs)-1)
			}
		}
		t.Fatalf("probe inconclusive after %d attempts (tool failure, not RED): the validation→rename window was never caught by the watcher; %s", attempts, summary)
	}
	t.Fatalf("RED (intended): the write escaped the validated parent after the post-validation swap — external sentinel received %s (rename at install.go:750 followed the swapped parent)", externalWrite)
}

// attemptParentSwapProbe runs one probe attempt: a fresh root, a
// root/sub/file.txt confinedWrite, and a watcher that swaps root/sub to a
// sentinel symlink the moment the temp file appears inside it. The payload
// is large on purpose — the Write duration IS the :725→:750 window the
// watcher aims at, and a multi-megabyte write turns it into milliseconds.
// Returns (external path that received the write, confinedWrite's error), or
// ("", nil) when the attempt ended without an escape and without an error.
func attemptParentSwapProbe(t *testing.T) (string, error) {
	t.Helper()
	home := t.TempDir()
	sentinel := t.TempDir()

	rootDir := filepath.Join(home, ".claude", "skills")
	resolvedRoot, err := resolveRoot(home, Root{Slug: RootClaudeSkills, Dir: rootDir})
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	subDir := filepath.Join(resolvedRoot.dir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(sentinel, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	inst := &Installer{Home: home, MoaiVersion: "test"}
	payload := make([]byte, 32*1024*1024) // 32 MB: the Write loop is the wide window

	// Watcher: the temp file's appearance in the VALIDATED parent is the
	// window signal. The swap keeps tmpName VALID so the write proceeds to
	// its rename instead of aborting: move the temp file itself into the
	// sentinel (same volume — the writer's open FD follows the inode), then
	// repoint the mid-chain dir at the sentinel. confinedWrite's Chmod and
	// Rename then resolve tmpName THROUGH the new symlink — straight into
	// the sentinel.
	stop := make(chan struct{})
	watcherDone := make(chan struct{})
	go func() {
		defer close(watcherDone)
		for {
			select {
			case <-stop:
				return
			default:
			}
			entries, err := os.ReadDir(subDir)
			if err == nil {
				for _, e := range entries {
					if len(e.Name()) >= 10 && e.Name()[:10] == ".ua-write-" {
						tmpPath := filepath.Join(subDir, e.Name())
						outPath := filepath.Join(sentinel, "sub", e.Name())
						if moveErr := os.Rename(tmpPath, outPath); moveErr != nil {
							return
						}
						stash := subDir + ".stash"
						_ = os.Remove(stash)
						if renameErr := os.Rename(subDir, stash); renameErr != nil {
							return
						}
						if linkErr := os.Symlink(filepath.Join(sentinel, "sub"), subDir); linkErr != nil {
							return
						}
						return
					}
				}
			}
			time.Sleep(20 * time.Microsecond)
		}
	}()

	writeErr := inst.confinedWrite(resolvedRoot, "sub/file.txt", payload, true)
	close(stop)
	<-watcherDone

	// Restore the stashed dir so t.TempDir cleanup stays trivial; the probe
	// reads the sentinel either way.
	stash := subDir + ".stash"
	if fi, statErr := os.Lstat(subDir); statErr == nil && fi.Mode()&os.ModeSymlink != 0 {
		_ = os.Remove(subDir)
	}
	if _, statErr := os.Lstat(stash); statErr == nil {
		if _, subErr := os.Lstat(subDir); subErr != nil {
			_ = os.Rename(stash, subDir)
		} else {
			_ = os.RemoveAll(stash)
		}
	}

	// The external record: any file inside the sentinel.
	sentEntries, err := os.ReadDir(filepath.Join(sentinel, "sub"))
	if err == nil && len(sentEntries) > 0 {
		return filepath.Join(sentinel, "sub", sentEntries[0].Name()), writeErr
	}
	return "", writeErr // a refused or failed write without an escape is a missed attempt
}

// TestConfinedWritePreservesExecBit — AC-025 (ledger 6b, REQ-COL-003).
// Given an install set carrying an executable script asset
// (navigator-audit.sh shape), the installed .sh must record mode 0755.
// RED-now reason: confinedWrite hardcodes 0o644 (install.go:743), so every
// installed script drops its exec bit.
func TestConfinedWritePreservesExecBit(t *testing.T) {
	f := newFixture(t)
	script := []byte("#!/bin/sh\necho navigator-audit probe\n")
	f.src[".claude/skills/moai-alpha/tools/navigator-audit.sh"] = &fstest.MapFile{Data: script}
	f.cat.Catalog.Core.Skills[0].Path = "templates/.claude/skills/moai-alpha/"

	if _, err := f.installer(t).Install(nil); err != nil {
		t.Fatalf("Install: %v", err)
	}
	installed := filepath.Join(f.home, ".claude", "skills", "moai-alpha", "tools", "navigator-audit.sh")
	info, err := os.Stat(installed)
	if err != nil {
		t.Fatalf("installed script missing: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o755 {
		t.Fatalf("RED (intended): installed .sh mode = %o, want 755 — confinedWrite hardcodes 0o644 (install.go:743) and drops the exec bit", got)
	}
}
