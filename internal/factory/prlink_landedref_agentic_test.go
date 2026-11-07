package factory

import (
	"os"
	"path/filepath"
	"testing"
)

// agenticTrunkProject writes a manual af-flow git-strategy (the workflow value
// the loader maps to AgenticTrunkWorkflow, integration target main).
func agenticTrunkProject(t *testing.T, baseKey string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "git_strategy:\n    mode: manual\n"
	if baseKey != "" {
		body += "    worktree_base_branch: " + baseKey + "\n"
	}
	body += "    manual:\n        workflow: agentic-trunk\n"
	if err := os.WriteFile(filepath.Join(dir, "git-strategy.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
	return root
}

// Under af-flow the landed ref is the LOCAL integration target: nothing is
// pushed per card, so an origin/ ref would never contain a landed card. The
// answer wins over a configured worktree_base_branch and never consults the
// origin/HEAD symref.
func TestLandedRefForWithLevel_AgenticTrunkAnswersLocalTarget(t *testing.T) {
	for _, base := range []string{"", "main"} {
		root := agenticTrunkProject(t, base)
		calls := refRunRecorder(t)
		ref, level := LandedRefForWithLevel(root)
		if ref != "main" || level != LandedRefAgenticTrunk {
			t.Errorf("base=%q: ref=%q level=%d, want main at level %d", base, ref, level, LandedRefAgenticTrunk)
		}
		if n := len(*calls); n != 0 {
			t.Errorf("base=%q: the symref was consulted %d time(s): %v", base, n, *calls)
		}
	}
	if got := LandedRefLevelSource(LandedRefAgenticTrunk); got == "" || got == LandedRefLevelSource(LandedRefDefault) {
		t.Errorf("LandedRefLevelSource(af-flow) = %q, want its own disclosure", got)
	}
}

// A non-af-flow project keeps the origin/ answer exactly.
func TestLandedRefForWithLevel_NonAgenticTrunkUnchanged(t *testing.T) {
	root := configuredProject(t, "develop")
	if ref, level := LandedRefForWithLevel(root); ref != "origin/develop" || level != LandedRefConfigured {
		t.Errorf("ref=%q level=%d, want origin/develop at level %d", ref, level, LandedRefConfigured)
	}
}
