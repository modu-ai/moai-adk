package cli

// Card t1216: hand-edited git-strategy.yaml values (git-flow workflow keys and
// worktree_base_branch) must survive `moai update` in a project whose merge
// BASE snapshot was written by a binary from before card t1139. Those binaries
// took the snapshot AFTER the restore step, so it holds the user's values; the
// update then reads every customization as "unchanged from BASE" and replaces
// it with the template default.
//
// Not parallel: the init and update helpers set env vars and chdir.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// plantLegacyPostRestoreSnapshot replaces the project's snapshot with a byte
// copy of the current section files and nothing else — what a pre-t1139
// update left on disk right after its restore step.
func plantLegacyPostRestoreSnapshot(t *testing.T, root string) {
	t.Helper()
	snapDir := backup.SnapshotDir(root)
	if err := os.RemoveAll(snapDir); err != nil {
		t.Fatalf("clear snapshot: %v", err)
	}
	src := filepath.Join(root, ".moai", "config", "sections")
	dst := filepath.Join(snapDir, "sections")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir snapshot: %v", err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read sections: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			t.Fatalf("write snapshot %s: %v", e.Name(), err)
		}
	}
}

// editGitStrategyToGitFlow applies the edits this repository re-applies by
// hand after every update: worktree_base_branch, the manual block's workflow,
// and one git-flow key the template does not ship.
func editGitStrategyToGitFlow(t *testing.T, root string) {
	t.Helper()
	p := filepath.Join(root, ".moai", "config", "sections", "git-strategy.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read git-strategy.yaml: %v", err)
	}
	s := string(data)
	for old, repl := range map[string]string{
		`worktree_base_branch: ""`: `worktree_base_branch: develop`,
		// The first workflow line belongs to the manual block.
		"workflow: github-flow": "workflow: git-flow\n    develop_branch: develop",
	} {
		if !strings.Contains(s, old) {
			t.Fatalf("premise: git-strategy.yaml no longer carries %q", old)
		}
		s = strings.Replace(s, old, repl, 1)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatalf("write git-strategy.yaml: %v", err)
	}
}

func assertGitFlowKeysKept(t *testing.T, root, when string) {
	t.Helper()
	got := flattenSections(t, root)
	for key, want := range map[string]string{
		"git-strategy.yaml:git_strategy.worktree_base_branch":  "develop",
		"git-strategy.yaml:git_strategy.manual.workflow":       "git-flow",
		"git-strategy.yaml:git_strategy.manual.develop_branch": "develop",
	} {
		if got[key] != want {
			t.Errorf("%s: %s = %q, want %q", when, key, got[key], want)
		}
	}
}

// TestUpdate_LegacySnapshot_KeepsGitStrategyCustomizations is the t1216
// reproduction: the first update after the fix must not spend the user's
// customizations on the legacy snapshot, and the second must keep them too.
func TestUpdate_LegacySnapshot_KeepsGitStrategyCustomizations(t *testing.T) {
	root := initUserOwnedKeysProject(t)
	editGitStrategyToGitFlow(t, root)
	plantLegacyPostRestoreSnapshot(t, root)

	for i := 1; i <= 2; i++ {
		if out, err := runForcedTemplateSyncResult(t, root); err != nil {
			t.Fatalf("update #%d: %v\n%s", i, err, out)
		}
		assertGitFlowKeysKept(t, root, fmt.Sprintf("after update #%d", i))
	}
}

// TestUpdate_CurrentSnapshot_KeepsGitStrategyCustomizations is the positive
// control for the harness above: with the snapshot this build writes, the
// same edits already survive, so a failure of the legacy case is attributable
// to the snapshot and not to the fixture or the assertion.
func TestUpdate_CurrentSnapshot_KeepsGitStrategyCustomizations(t *testing.T) {
	root := initUserOwnedKeysProject(t)
	editGitStrategyToGitFlow(t, root)

	for i := 1; i <= 2; i++ {
		if out, err := runForcedTemplateSyncResult(t, root); err != nil {
			t.Fatalf("update #%d: %v\n%s", i, err, out)
		}
		assertGitFlowKeysKept(t, root, fmt.Sprintf("after update #%d", i))
	}
}
