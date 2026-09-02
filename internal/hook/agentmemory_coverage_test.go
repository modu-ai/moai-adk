// agentmemory_coverage_test.go — branch coverage for the drain/mirror core
// (SPEC-AGENT-MEMORY-DRAIN-001): the collision-refresh path, no-op paths,
// helper branches, and the mirror wrapper's input-tolerance cases.
package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestDrainTreeCollisionRefreshKeepsPlainIntact pins the tree-qualified slot
// semantics: a second sync from the same tree with NEW content refreshes its
// own `.wt-` slot and never touches the plain file, and never fans out into
// further names.
func TestDrainTreeCollisionRefreshKeepsPlainIntact(t *testing.T) {
	t.Parallel()
	primary := t.TempDir()
	tree := t.TempDir()
	seedAgentStore(t, primary, map[string]string{
		"manager-spec/feedback_dup.md": "primary content",
	})
	seedAgentStore(t, tree, map[string]string{
		"manager-spec/feedback_dup.md": "v1",
	})
	if _, err := DrainTree(primary, tree, true); err != nil {
		t.Fatalf("DrainTree v1: %v", err)
	}

	// The tree's content evolves (the write-time mirror's repeated-edit case).
	seedAgentStore(t, tree, map[string]string{
		"manager-spec/feedback_dup.md": "v2",
	})
	rec, err := DrainTree(primary, tree, true)
	if err != nil {
		t.Fatalf("DrainTree v2: %v", err)
	}
	suffixedPath := filepath.Join(primary, ".claude", "agent-memory", "manager-spec",
		"feedback_dup.wt-"+filepath.Base(tree)+".md")
	if got := readFileOrFatal(t, suffixedPath); got != "v2" {
		t.Errorf("tree-qualified slot not refreshed: %q", got)
	}
	if got := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_dup.md")); got != "primary content" {
		t.Errorf("plain primary file was overwritten: %q", got)
	}
	if rec.Collided != 1 || rec.Copied != 0 {
		t.Errorf("refresh misclassified: %+v", rec)
	}
	// No name fan-out: exactly one suffixed copy exists.
	matches, err := filepath.Glob(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_dup*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("expected plain + one suffixed copy, got %v", matches)
	}
}

// TestMirrorAgentMemoryNoOps walks every documented no-op input: the index
// itself, a relative-but-resolvable path is still mirrored (path forms), and
// a missing file is an error (fail-open at the wrapper).
//
// NOT parallel: mutates the agentMemoryPrimaryRootFn package seam.
func TestMirrorAgentMemoryNoOps(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	seedWorktreeX(t, tree)
	swapMirrorPrimaryRoot(t, func(string) (string, bool, error) {
		return primary, true, nil
	})

	// The index itself never mirrors.
	idx := filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "MEMORY.md")
	mirrored, err := MirrorAgentMemoryFile(idx)
	if err != nil || mirrored {
		t.Errorf("index mirrored (mirrored=%v err=%v)", mirrored, err)
	}

	// A missing topic file is an error (the wrapper turns it into a notice).
	missing := filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_gone.md")
	if _, err := MirrorAgentMemoryFile(missing); err == nil {
		t.Error("missing file did not error")
	}
}

// TestSplitAgentMemoryPathLeadingForm covers the bare relative form
// `.claude/agent-memory/<agent>/<file>` with no tree-root prefix.
func TestSplitAgentMemoryPathLeadingForm(t *testing.T) {
	t.Parallel()
	root, agent, rel, ok := SplitAgentMemoryPath(".claude/agent-memory/manager-spec/f.md")
	if !ok {
		t.Fatal("leading form rejected")
	}
	if root != "" || agent != "manager-spec" || rel != "f.md" {
		t.Errorf("got (%q,%q,%q), want (\"\", manager-spec, f.md)", root, agent, rel)
	}
	if _, _, _, ok := SplitAgentMemoryPath(".claude/agent-memory/onlyagent"); ok {
		t.Error("agent without rel accepted")
	}
}

// TestAgentMemoryHelpers exercises the small helpers' direct branches.
func TestAgentMemoryHelpers(t *testing.T) {
	t.Parallel()
	if got := treeQualifiedCopyName("feedback_x.md", "t223"); got != "feedback_x.wt-t223.md" {
		t.Errorf("treeQualifiedCopyName(.md) = %q", got)
	}
	if got := treeQualifiedCopyName("notes", "t223"); got != "notes.wt-t223" {
		t.Errorf("treeQualifiedCopyName(bare) = %q", got)
	}
	if !indexLinksTarget("- [a](feedback_x.md) — x", "feedback_x.md") {
		t.Error("direct link not detected")
	}
	if indexLinksTarget("- [a](sub/feedback_y.md) — y", "feedback_x.md") {
		t.Error("wrong target matched")
	}
	if got := absPathAgainst("/base", "rel/dir"); got != "/base/rel/dir" {
		t.Errorf("absPathAgainst relative = %q", got)
	}
	if got := absPathAgainst("/base", "/abs/dir"); got != "/abs/dir" {
		t.Errorf("absPathAgainst absolute = %q", got)
	}
}

// TestMirrorAgentMemoryWrapperInputTolerance covers the wrapper's early
// returns: unparseable tool input and a missing/empty file_path are silent
// no-ops (no stderr, no panic).
func TestMirrorAgentMemoryWrapperInputTolerance(t *testing.T) {
	t.Parallel()
	mirrorAgentMemory(&HookInput{ToolName: "Write", ToolInput: json.RawMessage(`{not-json`)})
	mirrorAgentMemory(&HookInput{ToolName: "Write", ToolInput: json.RawMessage(`{"file_path":""}`)})
	mirrorAgentMemory(&HookInput{ToolName: "Write", ToolInput: json.RawMessage(`{}`)})

	// A non-agent-memory path is a silent no-op through the wrapper too.
	toolInput, err := json.Marshal(map[string]any{"file_path": "/tmp/docs/agent-memory/x.md"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mirrorAgentMemory(&HookInput{ToolName: "Write", ToolInput: toolInput})
}

// TestMirrorAgentMemoryRelativePath covers the relative file_path form: the
// mirror resolves it against the process cwd before splitting. Uses a real
// chdir (guarded, non-parallel).
func TestMirrorAgentMemoryRelativePath(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	seedWorktreeX(t, tree)
	swapMirrorPrimaryRoot(t, func(string) (string, bool, error) {
		return primary, true, nil
	})

	saved, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tree); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(saved) })

	mirrored, err := MirrorAgentMemoryFile(".claude/agent-memory/manager-spec/feedback_x.md")
	if err != nil {
		t.Fatalf("MirrorAgentMemoryFile relative: %v", err)
	}
	if !mirrored {
		t.Error("relative worktree path not mirrored")
	}
	if _, err := os.Stat(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md")); err != nil {
		t.Errorf("primary copy missing: %v", err)
	}
}
