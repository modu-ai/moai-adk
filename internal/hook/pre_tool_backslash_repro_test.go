package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
		// symlink the way the OS walks it. The write is load-bearing evidence:
		// its failure is fatal, and the file is read back so the recorded
		// landing claims observed content, not an assumed one.
		simulated := filepath.Join(projectDir, `innocent\dir`, "escaped.txt")
		if werr := os.WriteFile(simulated, []byte("escaped"), 0o644); werr != nil {
			t.Fatalf("guard allowed the escape AND the simulated write failed: decision=%q writeErr=%v", decision, werr)
		}
		data, rerr := os.ReadFile(filepath.Join(outsideDir, "escaped.txt"))
		if rerr != nil {
			t.Fatalf("guard allowed the escape AND the written external file is unreadable: decision=%q readErr=%v", decision, rerr)
		}
		if string(data) != "escaped" {
			t.Fatalf("external file content %q, want %q", string(data), "escaped")
		}
		t.Fatalf("guard allowed the backslash-symlink escape: decision=%q reason=%q (external write VERIFIED at %s, content %q)",
			decision, reason, filepath.Join(outsideDir, "escaped.txt"), string(data))
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

// TestCheckFileAccessPosixBackslashLegitNameAllowed pins the false-positive
// face of the repair (SPEC-HOOK-BACKSLASH-SYMLINK-001 AC-HBS-004 /
// REQ-HBS-005): a legitimate NEW file whose name merely contains a `\` must
// not be denied by the boundary check. Platform-independent by design — the
// correct semantics differ per platform (on POSIX `weird\name.txt` is ONE
// missing component and rejoins as the unresolved tail; on Windows the same
// spelling splits into two missing components) — but the decision outcome is
// the same everywhere: NOT a boundary deny. This guards the repair against a
// character-blacklist regression.
func TestCheckFileAccessPosixBackslashLegitNameAllowed(t *testing.T) {
	projectDir := t.TempDir()

	handler := &preToolHandler{
		cfg:        &mockConfigProvider{cfg: newTestConfig()},
		policy:     &SecurityPolicy{},
		projectDir: projectDir,
	}

	toolInput, err := json.Marshal(map[string]string{
		"file_path": filepath.Join(projectDir, `weird\name.txt`),
	})
	if err != nil {
		t.Fatalf("failed to marshal tool input: %v", err)
	}

	decision, reason := handler.checkFileAccess(toolInput, "Write")
	if decision == DecisionDeny {
		t.Fatalf("legitimate new file `weird\\name.txt` inside the project was denied: reason=%q", reason)
	}
}

// TestPathSegmentsPlatformSeparatorSemantics pins the segmentation unit
// behind the walk at the string level, no filesystem required
// (SPEC-HOOK-BACKSLASH-SYMLINK-001 AC-HBS-005): the Windows branch treats
// `\` as a separator; the POSIX branch preserves every `\` inside a
// component verbatim. Runs on every CI runner.
func TestPathSegmentsPlatformSeparatorSemantics(t *testing.T) {
	t.Run("windows splits on both separators", func(t *testing.T) {
		got := pathSegments(`C:\proj\linked\..\x`, true)
		want := []string{"C:", "proj", "linked", "..", "x"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("pathSegments(windows) = %q, want %q", got, want)
		}
	})
	t.Run("posix preserves literal backslash inside a component", func(t *testing.T) {
		got := pathSegments(`/project/innocent\dir/escaped.txt`, false)
		want := []string{"", "project", `innocent\dir`, "escaped.txt"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("pathSegments(posix) = %q, want %q (the backslash must survive as an ordinary filename character)", got, want)
		}
	})
}

// TestResolvePhysicalWalkBranchPreservation characterizes the two walk
// branches the repair must NOT change (SPEC-HOOK-BACKSLASH-SYMLINK-001
// AC-HBS-007 / REQ-HBS-006), with fixtures whose shapes actually ENGAGE the
// branches (plan-audit iteration-2 D3-r respec):
//
//  1. depth bound (pre_tool.go `depth >= zoneSymlinkDepthBound`): a chain of
//     zoneSymlinkDepthBound+1 directory symlinks with a NON-EXISTENT
//     terminal target. The missing destination makes every component's
//     EvalSymlinks fail while Lstat/Readlink succeed, forcing the hop-by-hop
//     recursion that increments depth per link — an existing-terminal chain
//     would resolve wholesale at the first EvalSymlinks and never reach the
//     bound. Expect ok=false (fail-closed), and the decision-level fallback
//     keeps the existing unresolved in-project path (NOT a deny).
//  2. physical `..` pop (the `..` branch): `<project>/linked/../leaf` with
//     `linked` an EXISTING outside-pointing directory symlink. The component
//     resolves, then `..` pops against the RESOLVED outside prefix — the
//     physical-pop semantics that make the escape visible. Expect ok=true
//     with the resolution OUTSIDE the project, and checkFileAccess denying.
func TestResolvePhysicalWalkBranchPreservation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-only fixtures: unprivileged directory symlinks are not creatable on Windows")
	}

	t.Run("depth bound with non-existent terminal target fails closed", func(t *testing.T) {
		projectDir := t.TempDir()

		// l1 -> l2 -> ... -> lN -> missing-terminal (N = bound + 1 links).
		const links = zoneSymlinkDepthBound + 1
		prev := "missing-terminal-that-never-exists"
		for i := links; i >= 1; i-- {
			link := filepath.Join(projectDir, fmt.Sprintf("l%d", i))
			if err := os.Symlink(prev, link); err != nil {
				t.Skipf("cannot create directory symlink: %v", err)
			}
			prev = fmt.Sprintf("l%d", i)
		}

		walkPath, err := absoluteUncleaned(filepath.Join(projectDir, "l1", "leaf"))
		if err != nil {
			t.Fatalf("absoluteUncleaned failed: %v", err)
		}
		resolved, ok := resolveThroughExistingParent(walkPath)
		if ok {
			t.Fatalf("chain of %d symlinks past the depth bound resolved to %q with ok=true; want fail-closed ok=false", links, resolved)
		}

		// Companion decision-level assert: the existing fallback semantics
		// apply — ok=false keeps the unresolved (in-project) Abs path and
		// does NOT deny (REQ-HBS-006 behavior preservation).
		handler := &preToolHandler{
			cfg:        &mockConfigProvider{cfg: newTestConfig()},
			policy:     &SecurityPolicy{},
			projectDir: projectDir,
		}
		toolInput, err := json.Marshal(map[string]string{
			"file_path": filepath.Join(projectDir, "l1", "leaf"),
		})
		if err != nil {
			t.Fatalf("failed to marshal tool input: %v", err)
		}
		decision, reason := handler.checkFileAccess(toolInput, "Write")
		if decision == DecisionDeny {
			t.Fatalf("depth-bound fallback changed the existing semantics: in-project unresolved path was denied: reason=%q", reason)
		}
	})

	t.Run("dot dot pops against resolved outside symlink prefix", func(t *testing.T) {
		projectDir := t.TempDir()
		outsideDir := t.TempDir()

		linkPath := filepath.Join(projectDir, "linked")
		if err := os.Symlink(outsideDir, linkPath); err != nil {
			t.Skipf("cannot create directory symlink: %v", err)
		}

		// Built by CONCATENATION, never Join: filepath.Join would Clean the
		// `linked/..` segment against the lexical parent before the walk can
		// resolve it physically (plan.md §G anti-pattern).
		walkPath, err := absoluteUncleaned(projectDir + string(os.PathSeparator) + "linked/../leaf")
		if err != nil {
			t.Fatalf("absoluteUncleaned failed: %v", err)
		}
		resolved, ok := resolveThroughExistingParent(walkPath)
		if !ok {
			t.Fatalf("physical `..` pop refused the path entirely (resolved=%q); the pop against the resolved outside prefix must resolve with ok=true", resolved)
		}
		realProject, perr := filepath.EvalSymlinks(projectDir)
		if perr != nil {
			t.Fatalf("EvalSymlinks(projectDir) failed: %v", perr)
		}
		if rel, relErr := filepath.Rel(realProject, resolved); relErr == nil && !strings.HasPrefix(rel, "..") {
			t.Fatalf("`linked/../leaf` resolved to the IN-PROJECT path %q; the `..` must pop against the RESOLVED outside prefix %q (physical pop, not lexical)", resolved, outsideDir)
		}

		// Companion decision-level assert: the escape the pop made visible
		// is denied at the boundary check, unchanged.
		handler := &preToolHandler{
			cfg:        &mockConfigProvider{cfg: newTestConfig()},
			policy:     &SecurityPolicy{},
			projectDir: projectDir,
		}
		toolInput, err := json.Marshal(map[string]string{
			// Raw absolute spelling, never Join — see the walkPath note above.
			"file_path": projectDir + string(os.PathSeparator) + "linked/../leaf",
		})
		if err != nil {
			t.Fatalf("failed to marshal tool input: %v", err)
		}
		decision, reason := handler.checkFileAccess(toolInput, "Write")
		if decision != DecisionDeny {
			t.Fatalf("`linked/../leaf` escape through an outside symlink was not denied: decision=%q reason=%q", decision, reason)
		}
	})
}
