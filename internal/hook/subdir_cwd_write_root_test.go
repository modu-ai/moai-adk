package hook

// subdir_cwd_write_root_test.go — card t1165. Each test drives one hook write
// site with HookInput.CWD naming a subdirectory of a project root that already
// carries .moai/, and asserts the write does not grow a stray <subdir>/.moai
// tree. Sibling of card t1160 (instructions_loaded_audit_test.go), which fixed
// the same defect class at the InstructionsLoaded audit write.
//
// None of these tests is parallel: each pins CLAUDE_PROJECT_DIR with t.Setenv
// so an ambient value cannot redirect the write.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// newMoaiProjectRoot returns a temp dir that already carries .moai/, the
// write-side guard's precondition for a MoAI project root. Fixtures for the
// write sites this card moved onto that guard use it; a bare t.TempDir() would
// make a positive assertion fail and a negative one pass on the guard alone
// instead of on the reason the test names. Safe for parallel tests: TestMain
// clears CLAUDE_PROJECT_DIR for the binary.
func newMoaiProjectRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	markMoaiProjectRoot(t, root)
	return root
}

// markMoaiProjectRoot creates <dir>/.moai/ so an existing fixture directory
// (e.g. a freshly initialized git repo) qualifies as a MoAI project root.
func markMoaiProjectRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o750); err != nil {
		t.Fatal(err)
	}
}

// newSubdirProject returns a project root carrying .moai/ and a subdirectory
// of it carrying none, with CLAUDE_PROJECT_DIR cleared.
func newSubdirProject(t *testing.T) (root, sub string) {
	t.Helper()
	root = newMoaiProjectRoot(t)
	sub = filepath.Join(root, "sub")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvClaudeProjectDir, "")
	return root, sub
}

// assertNoStrayMoai fails when <sub>/.moai exists.
func assertNoStrayMoai(t *testing.T, sub string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(sub, ".moai")); !os.IsNotExist(err) {
		t.Errorf("stray %s/.moai created (stat err = %v)", sub, err)
	}
}

// Site 1 — ConfigChange audit log.
func TestConfigChangeAudit_NoStrayTreeInSubdirCWD(t *testing.T) {
	_, sub := newSubdirProject(t)

	appendConfigChangeAudit(&HookInput{
		SessionID:     "sess-sub",
		HookEventName: "ConfigChange",
		CWD:           sub,
	}, configChangeResultReloaded, "probe")

	assertNoStrayMoai(t, sub)
}

// Site 2 — PreCompact session memo.
func TestCompact_NoStrayTreeInSubdirCWD(t *testing.T) {
	_, sub := newSubdirProject(t)

	if _, err := NewCompactHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-sub",
		HookEventName: "PreCompact",
		CWD:           sub,
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	assertNoStrayMoai(t, sub)
}

// Site 3 — FileChanged MX sidecar. The changed file lives inside the
// subdirectory so the containment guards (FilePath within the resolved root)
// pass and the sidecar write is actually reached.
func TestRunMXScan_NoStrayTreeInSubdirCWD(t *testing.T) {
	_, sub := newSubdirProject(t)
	file := filepath.Join(sub, "tagged.go")
	if err := os.WriteFile(file, []byte("// @MX:NOTE: probe\npackage sub\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewFileChangedHandler().(*fileChangedHandler)
	h.runMXScan(context.Background(), &HookInput{
		SessionID:     "sess-sub",
		FilePath:      file,
		ChangeType:    "modified",
		HookEventName: "FileChanged",
		CWD:           sub,
	})

	assertNoStrayMoai(t, sub)
}

// Site 4 — WorktreeCreate registry entry. registerEntry is the step Handle
// calls after the worktree exists; driving it directly avoids provisioning a
// real git worktree, which does not bear on where the registry is written.
func TestWorktreeCreateRegister_NoStrayTreeInSubdirCWD(t *testing.T) {
	root, sub := newSubdirProject(t)

	h := NewWorktreeCreateHandler().(*worktreeCreateHandler)
	h.registerEntry(&HookInput{
		SessionID:     "sess-sub",
		HookEventName: "WorktreeCreate",
		CWD:           sub,
		AgentName:     "probe",
	}, filepath.Join(root, ".claude", "worktrees", "probe"), "worktree-probe")

	assertNoStrayMoai(t, sub)
}

// Site 5 — WorktreeRemove registry rewrite.
func TestWorktreeRemove_NoStrayTreeInSubdirCWD(t *testing.T) {
	root, sub := newSubdirProject(t)

	if _, err := NewWorktreeRemoveHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-sub",
		HookEventName: "WorktreeRemove",
		CWD:           sub,
		WorktreePath:  filepath.Join(root, ".claude", "worktrees", "probe"),
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	assertNoStrayMoai(t, sub)
}

// TestSubdirCWDWritesLandInProjectDir is the positive half for sites 1-5: with
// CLAUDE_PROJECT_DIR naming the root, a subdirectory cwd still records each
// write under <root>/.moai/ and nothing under the subdirectory. Without it the
// no-stray tests above would pass on a site that simply stopped writing.
func TestSubdirCWDWritesLandInProjectDir(t *testing.T) {
	root, sub := newSubdirProject(t)
	t.Setenv(config.EnvClaudeProjectDir, root)
	ctx := context.Background()

	appendConfigChangeAudit(&HookInput{SessionID: "s", CWD: sub}, configChangeResultReloaded, "probe")

	if _, err := NewCompactHandler().Handle(ctx, &HookInput{SessionID: "s", HookEventName: "PreCompact", CWD: sub}); err != nil {
		t.Fatalf("compact: %v", err)
	}

	file := filepath.Join(root, "tagged.go")
	if err := os.WriteFile(file, []byte("// @MX:NOTE: probe\npackage root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	NewFileChangedHandler().(*fileChangedHandler).runMXScan(ctx, &HookInput{
		SessionID: "s", FilePath: file, ChangeType: "modified", HookEventName: "FileChanged", CWD: sub,
	})

	wtPath := filepath.Join(root, ".claude", "worktrees", "probe")
	NewWorktreeCreateHandler().(*worktreeCreateHandler).registerEntry(&HookInput{SessionID: "s", CWD: sub}, wtPath, "b")
	if got := loadWorktreeEntries(worktreeStateFile(root)); len(got) != 1 || got[0].Path != wtPath {
		t.Errorf("registry after create = %+v, want one entry for %s", got, wtPath)
	}
	if _, err := NewWorktreeRemoveHandler().Handle(ctx, &HookInput{SessionID: "s", CWD: sub, WorktreePath: wtPath}); err != nil {
		t.Fatalf("worktree remove: %v", err)
	}
	if got := loadWorktreeEntries(worktreeStateFile(root)); len(got) != 0 {
		t.Errorf("registry after remove = %+v, want empty", got)
	}

	for _, rel := range []string{configChangeAuditRelPath, ".moai/state/session-memo.md"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Errorf("%s missing under root: %v", rel, err)
		}
	}
	if entries, err := os.ReadDir(filepath.Join(root, ".moai", "state")); err != nil || len(entries) < 3 {
		t.Errorf("root .moai/state entries = %d (err %v), want memo + mx sidecar + worktrees.json", len(entries), err)
	}
	assertNoStrayMoai(t, sub)
}

// TestCompactMemoRoundTripFromSubdirCWD pins the PreCompact writer and the
// PostCompact reader to the same root. Moving only the writer to the project
// root would leave the reader looking under the subdirectory and silently lose
// the memo in every subdirectory session.
func TestCompactMemoRoundTripFromSubdirCWD(t *testing.T) {
	root, sub := newSubdirProject(t)
	t.Setenv(config.EnvClaudeProjectDir, root)
	ctx := context.Background()

	if _, err := NewCompactHandler().Handle(ctx, &HookInput{SessionID: "sess-rt", HookEventName: "PreCompact", CWD: sub}); err != nil {
		t.Fatalf("pre-compact: %v", err)
	}
	out, err := NewPostCompactHandler().Handle(ctx, &HookInput{SessionID: "sess-rt", HookEventName: "PostCompact", CWD: sub})
	if err != nil {
		t.Fatalf("post-compact: %v", err)
	}
	if !strings.Contains(out.SystemMessage, "sess-rt") {
		t.Errorf("post-compact SystemMessage = %q, want the restored memo naming sess-rt", out.SystemMessage)
	}
}

// Site 6 — chain lineage banner, called as session_start.go calls it, after
// the protocol normalization the CLI applies to every event (validateInput
// fills ProjectDir from CWD when CLAUDE_PROJECT_DIR is unset). With ProjectDir
// left empty the banner returns before any write, so the normalized input is
// the only route from a subdirectory cwd to the store's MkdirAll.
func TestChainBanner_NoStrayTreeFromNormalizedSubdirCWD(t *testing.T) {
	_, sub := newSubdirProject(t)
	t.Setenv(config.EnvChainNodeID, "")

	in := &HookInput{SessionID: "sess-sub", HookEventName: "SessionStart", CWD: sub}
	if err := validateInput(in); err != nil {
		t.Fatal(err)
	}
	_ = chainLineageBanner(in.ProjectDir, in.CWD, in.SessionID)

	assertNoStrayMoai(t, sub)
}

// Site 7 — kanban session record, ProjectDir empty so CWD stands in as the
// root. Precondition: a kanban role env (MOAI_KANBAN set, factory/label vars
// cleared) and Source "startup"; without a role the write is never reached.
func TestKanbanSessionRecord_NoStrayTreeInSubdirCWD(t *testing.T) {
	_, sub := newSubdirProject(t)
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "")
	t.Setenv(config.EnvMoaiKanbanLabel, "")
	t.Setenv(config.EnvMoaiKanban, "1")

	writeKanbanSessionRecord(&HookInput{SessionID: "sess-sub", Source: "startup", CWD: sub})

	assertNoStrayMoai(t, sub)
}
