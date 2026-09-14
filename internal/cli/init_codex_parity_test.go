package cli

// SPEC-INIT-HARNESS-001 M5 — synthesis (REQ-IH-012, REQ-IH-014,
// AC-IH-011/013/014): the non-interactive codex deployment equals the
// interactive codex selection's file set, and the preserved contracts stay
// pinned.

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sort"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// snapshotDeployedFileSet walks projectDir and returns the sorted relative
// path list, skipping the runtime state directory (codex wiring sidecar and
// shell-env writes land there and are identical across both paths).
func snapshotDeployedFileSet(t *testing.T, projectDir string) []string {
	t.Helper()
	var paths []string
	err := filepath.WalkDir(projectDir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(projectDir, p)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		// .moai/state carries the codex wiring trust sidecar written by both
		// paths identically; .moai/db/<name>-<path-hash> embeds the project's
		// absolute path, so the hash differs per temp dir by construction.
		// Both are runtime artifacts, not deployment — keep the comparison on
		// deployment.
		if rel == ".moai/state" || strings.HasPrefix(rel, ".moai/state/") ||
			rel == ".moai/db" || strings.HasPrefix(rel, ".moai/db/") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk deployed project: %v", err)
	}
	sort.Strings(paths)
	return paths
}

// TestInitCodexNonInteractiveParity verifies REQ-IH-012 (AC-IH-011): a
// non-interactive `--llm codex` init deploys the SAME file set as an
// interactive wizard codex selection. The non-interactive MCP asymmetry
// (REQ-IQW-006) is unobservable here BY CONSTRUCTION: codex writes no
// .mcp.json on either path.
func TestInitCodexNonInteractiveParity(t *testing.T) {
	homeInteractive := t.TempDir()
	t.Setenv("HOME", homeInteractive)
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	wiz := &wizard.WizardResult{AgentWiring: "codex"}
	interactiveDir := runInitForAutonomyAtHome(t, homeInteractive, wiz, nil)

	homeNonInteractive := t.TempDir()
	t.Setenv("HOME", homeNonInteractive)
	nonInteractiveDir := runInitForAutonomyAtHome(t, homeNonInteractive, nil, map[string]string{"llm": "codex"})

	got := snapshotDeployedFileSet(t, interactiveDir)
	want := snapshotDeployedFileSet(t, nonInteractiveDir)

	if len(got) != len(want) {
		t.Fatalf("file-set size mismatch: interactive=%d non-interactive=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("file-set diff at [%d]: interactive=%q non-interactive=%q", i, got[i], want[i])
		}
	}
	// Sanity: the parity target actually deployed codex surfaces.
	if _, err := os.Stat(filepath.Join(interactiveDir, "AGENTS.md")); err != nil {
		t.Fatalf("interactive codex init deployed no AGENTS.md — fixture broken: %v", err)
	}
	if _, err := os.Stat(filepath.Join(interactiveDir, ".claude")); err == nil {
		t.Fatal("interactive codex init deployed .claude/ — fixture broken")
	}
}
