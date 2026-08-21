package deploy

// Symlink-awareness tests for CleanMoaiManagedPaths (SPEC-CLI-CLEAN-SYMLINK-001).
// The contract under test: every entry in the clean set is classified with
// os.Lstat semantics — a symlink (live or dangling) enters a link-dedicated
// branch BEFORE the IsDir classification and never falls into the file or
// directory branch (REQ-CSL-001). The P0 case is the shipped Run D defect: a
// dangling symlink at a non-glob managed root made clean report "Skipped (not
// found)" while the link survived, and deployment's MkdirAll then died with
// EEXIST — leaving update permanently broken until the user removed the link
// by hand.
//
// The deploy side of the boundary is represented by the deployer's exact
// per-file parent-creation call (os.MkdirAll(destDir, 0o755),
// internal/template/deployer.go:189) rather than by importing the template
// package — deploy is a leaf package and the Go-test boundary is the card's
// verification surface (dossier §5 gap 6).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// makeSymlink creates link -> target, skipping the test when the host
// platform or privileges refuse symlink creation (REQ-CSL-012: skip, never
// assume creation succeeded).
func makeSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink %s -> %s on this host: %v (REQ-CSL-012)", link, target, err)
	}
}

// symlinkProgressLine reports whether output carries at least one line that
// names both the given path and the "symlink" token — the stable progress
// line contract (plan.md §D-7) that keeps link dispositions greppable
// (REQ-CSL-005).
func symlinkProgressLine(output, path string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, path) && strings.Contains(line, "symlink") {
			return true
		}
	}
	return false
}

// deploySimMkdirAll mirrors the deployer's per-file parent-directory
// creation (internal/template/deployer.go:189) — the exact call the shipped
// Run D defect broke update with (EEXIST on a lingering dangling link).
func deploySimMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Errorf("deploy-side MkdirAll(%s) failed: %v", dir, err)
	}
}

// TestCleanMoaiManagedPaths_DanglingSymlinkAtNonGlobRoot is AC-CSL-001 — the
// Run D repro converted to a Go test. Axes combined (REQ-CSL-011): link
// existence, progress output, deploy-side termination, redeployment, and the
// immediate-rerun loop closure.
func TestCleanMoaiManagedPaths_DanglingSymlinkAtNonGlobRoot(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, defs.ClaudeDir, defs.AgentsMoaiSubdir)
	// Given: the managed root exists as a real directory (post-init shape),
	// is removed, and a dangling symlink pointing at an absent path replaces
	// it — the link itself exists, its target does not.
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(agentsDir); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, filepath.Join(root, "t173-gone-target"), agentsDir)

	tmplFS := preCleanTestFS(".claude/agents/moai/manager.md")

	// When: the clean step processes the root.
	var out bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out, tmplFS); err != nil {
		t.Errorf("clean failed on dangling-symlink root: %v", err)
	}

	// Then 1: the link itself is gone (Lstat ENOENT — not "link still here").
	if _, lerr := os.Lstat(agentsDir); !os.IsNotExist(lerr) {
		t.Errorf("dangling symlink still present after clean (Lstat err = %v, want IsNotExist)", lerr)
	}

	// Then 2: the progress output names the path AND the dangling-symlink
	// form on one line.
	if !symlinkProgressLine(out.String(), filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir)) {
		t.Errorf("progress output has no line naming %s as a symlink:\n%s",
			filepath.Join(defs.ClaudeDir, defs.AgentsMoaiSubdir), out.String())
	}

	// Then 3: deployment's MkdirAll no longer hits EEXIST on the root.
	deploySimMkdirAll(t, agentsDir)

	// Then 4: the root is redeployable as a real directory carrying template
	// content (simulated at the deploy boundary: MkdirAll + one file write).
	tmplFile := filepath.Join(agentsDir, "manager.md")
	if err := os.WriteFile(tmplFile, []byte("template:manager"), 0o644); err != nil {
		t.Errorf("redeploy write failed: %v", err)
	}
	if fi, serr := os.Lstat(agentsDir); serr != nil || !fi.Mode().IsDir() {
		t.Errorf("root is not a real directory after redeploy: stat err = %v", serr)
	}

	// Then 5: an immediate rerun of clean + deploy succeeds — the rerun loop
	// the shipped defect left permanently open.
	var out2 bytes.Buffer
	if err := CleanMoaiManagedPaths(root, &out2, tmplFS); err != nil {
		t.Errorf("rerun clean failed: %v", err)
	}
	deploySimMkdirAll(t, agentsDir)
	if err := os.WriteFile(tmplFile, []byte("template:manager"), 0o644); err != nil {
		t.Errorf("rerun redeploy write failed: %v", err)
	}
}

// TestMakeSymlink_SkipsWhenCreationFails is AC-CSL-011 (REQ-CSL-012): when
// os.Symlink fails — here a deterministic EEXIST collision — the helper must
// t.Skip the test, never let it fail or assume creation succeeded. The
// reached-flag stays false exactly when the helper skipped (a skip unwinds
// the subtest before the assignment runs).
func TestMakeSymlink_SkipsWhenCreationFails(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "t")
	link := filepath.Join(root, "l")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("host cannot create symlinks at all, skip-path precondition unavailable: %v", err)
	}
	reached := false
	t.Run("collision", func(t *testing.T) {
		makeSymlink(t, target, link) // link path occupied: os.Symlink fails
		reached = true
	})
	if reached {
		t.Error("makeSymlink did not skip on os.Symlink failure (REQ-CSL-012: skip, not fail)")
	}
}

// TestCleanMoaiManagedPaths_SelfReferentialLinkSpotCheck is the acceptance
// §D.3 loop-link spot check: a self-pointing link must not send the clean
// step into an unbounded traversal (removal-only dispositions never walk
// through a link). The form classification Stat cannot resolve the loop
// (ELOOP), so the run fails loudly with the symlink-target attribution —
// bounded termination, no hang, no partial removal.
func TestCleanMoaiManagedPaths_SelfReferentialLinkSpotCheck(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, defs.ClaudeDir, defs.AgentsMoaiSubdir)
	if err := os.MkdirAll(filepath.Dir(agentsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	makeSymlink(t, agentsDir, agentsDir) // link points at itself

	var out bytes.Buffer
	err := CleanMoaiManagedPaths(root, &out, preCleanTestFS(".claude/agents/moai/manager.md"))
	if err == nil {
		t.Fatal("expected an attributed error for the unresolvable self-referential link, got nil")
	}
	if !strings.Contains(err.Error(), "stat symlink target") {
		t.Errorf("error does not attribute the symlink-target stat: %v", err)
	}
	// The pathological link itself is untouched — the run aborted before
	// removal, which is the loud-failure disposition for unclassifiable
	// forms (dangling links, by contrast, are removed; REQ-CSL-002).
	if fi, lerr := os.Lstat(agentsDir); lerr != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Errorf("self-referential link not left in place after abort: %v", lerr)
	}
}
