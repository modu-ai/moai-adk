package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCheckFileAccessPosixBackslashSymlinkEscape reproduces the P1 finding of
// the card-t1533 codex review gate round 1 (evidence:
// .moai/reports/t1556/codex-review-gate-1.md; SPEC-HOOK-BACKSLASH-SYMLINK-001):
// on POSIX a directory literally named `innocent\dir` and symlinked outside the
// project must NOT let a Write through — the guard must resolve the literal
// component as the symlink it is and deny.
//
// The plan-phase RED expectation (SPEC-HOOK-BACKSLASH-SYMLINK-001 AC-HBS-001):
// pre-fix, the unconditional backslash→slash conversion in resolvePhysicalWalk
// (pre_tool.go:1397) splits `innocent\dir` into the non-existent `innocent`
// component, the walk rejoins an in-project fictional path, the boundary check
// allows, and the simulated tool write lands OUTSIDE the project.
func TestCheckFileAccessPosixBackslashSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-only fixture: on Windows a backslash IS a path separator, so the literal component name is not creatable")
	}

	projectDir := t.TempDir()
	outsideDir := t.TempDir()

	linkPath := filepath.Join(projectDir, `innocent\dir`)
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("cannot create directory symlink: %v", err)
	}

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     &SecurityPolicy{},
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, `innocent\dir`, "escaped.txt"),
	})
	if err != nil {
		t.Fatalf("failed to marshal tool input: %v", err)
	}

	decision, reason := handler.checkFileAccess(toolInput, "Write")
	if decision != DecisionDeny {
		// Simulate exactly what the Write tool would do when the guard allows:
		// write through the literal path the tool received, following the
		// symlink the way the OS walks it.
		simulated := filepath.Join(projectDir, `innocent\dir`, "escaped.txt")
		if werr := os.WriteFile(simulated, []byte("escaped"), 0o644); werr != nil {
			t.Logf("simulated tool write failed: %v", werr)
		}
		t.Fatalf("guard allowed the backslash-symlink escape: decision=%q reason=%q (external write landed at %s)",
			decision, reason, filepath.Join(outsideDir, "escaped.txt"))
	}

	// Guard denied: no byte may have reached the external destination.
	if _, err := os.Stat(filepath.Join(outsideDir, "escaped.txt")); err == nil {
		t.Fatalf("external file %s exists although the guard denied the write", filepath.Join(outsideDir, "escaped.txt"))
	}
}

// TestResolveThroughExistingParentPosixBackslashSymlinkDivergence pins the
// resolution half of the same finding (SPEC-HOOK-BACKSLASH-SYMLINK-001
// AC-HBS-003): on POSIX the walk must follow the literal `innocent\dir`
// component as the symlink it is, so the resolved path lands OUTSIDE the
// project. Pre-fix, the backslash→slash conversion splits the component into a
// non-existent `innocent`, and the walk rejoins a fictional IN-PROJECT path
// with ok=true — validating a path the OS would never walk.
func TestResolveThroughExistingParentPosixBackslashSymlinkDivergence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-only fixture: on Windows a backslash IS a path separator, so the literal component name is not creatable")
	}

	projectDir := t.TempDir()
	outsideDir := t.TempDir()

	linkPath := filepath.Join(projectDir, `innocent\dir`)
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("cannot create directory symlink: %v", err)
	}

	abs, err := absoluteUncleaned(filepath.Join(projectDir, `innocent\dir`, "escaped.txt"))
	if err != nil {
		t.Fatalf("absoluteUncleaned failed: %v", err)
	}

	resolved, ok := resolveThroughExistingParent(abs)
	t.Logf("resolved=%q ok=%v projectDir=%q abs=%q", resolved, ok, projectDir, abs)
	if !ok {
		t.Fatalf("walk refused the literal path entirely (resolved=%q); it must resolve through the literal symlinked component instead", resolved)
	}
	// macOS /var is a symlink to /private/var: compare against the resolved
	// project prefix or Rel reports an error and the assert silently passes.
	realProject, perr := filepath.EvalSymlinks(projectDir)
	if perr != nil {
		t.Fatalf("EvalSymlinks(projectDir) failed: %v", perr)
	}
	if rel, relErr := filepath.Rel(realProject, resolved); relErr == nil && !strings.HasPrefix(rel, "..") {
		t.Fatalf("walk validated a fictional IN-PROJECT path %q while the literal component `innocent\\dir` is a symlink to %q — validated path diverges from the path the OS walks", resolved, outsideDir)
	}
}
