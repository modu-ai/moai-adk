// memory_mirror_test.go — write-time agent-memory mirror in the PostToolUse
// hook (SPEC-AGENT-MEMORY-DRAIN-001 M2).
//
// Fixture HookInputs + a seam over primary resolution; no real tree is
// touched.
package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// swapMirrorPrimaryRoot points the mirror's primary resolution at a fixture.
// NOT parallel-safe (mutates the package-level agentMemoryPrimaryRootFn seam)
// — callers must not t.Parallel(). Same contract the handoffRenameFunc seam
// tests already hold, and the one captureStderr states for os.Stderr.
func swapMirrorPrimaryRoot(t *testing.T, fn func(string) (string, bool, error)) {
	t.Helper()
	prev := agentMemoryPrimaryRootFn
	agentMemoryPrimaryRootFn = fn
	t.Cleanup(func() {
		agentMemoryPrimaryRootFn = prev
	})
}

// mirrorInput builds a Write-tool HookInput whose file_path is filePath.
func mirrorInput(t *testing.T, filePath string) *HookInput {
	t.Helper()
	toolInput, err := json.Marshal(map[string]any{"file_path": filePath})
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	return &HookInput{ToolName: "Write", ToolInput: toolInput}
}

// TestMirrorAgentMemoryCopiesToPrimary covers AC-AM-005: a Write targeting a
// worktree agent-memory file is copied to the same agent-relative path in
// the resolved primary store, and the primary index gains the line.
//
// NOT parallel: mutates the agentMemoryPrimaryRootFn package seam.
func TestMirrorAgentMemoryCopiesToPrimary(t *testing.T) {
	primary := t.TempDir()
	tree := t.TempDir()
	seedWorktreeX(t, tree)
	swapMirrorPrimaryRoot(t, func(string) (string, bool, error) {
		return primary, true, nil
	})

	src := filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_x.md")
	mirrored, err := MirrorAgentMemoryFile(src)
	if err != nil {
		t.Fatalf("MirrorAgentMemoryFile: %v", err)
	}
	if !mirrored {
		t.Fatal("mirror reported a no-op for a worktree agent-memory write")
	}
	if got := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md")); got != feedbackXBody {
		t.Errorf("primary copy differs:\nwant %q\ngot  %q", feedbackXBody, got)
	}
	idx := readFileOrFatal(t, filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "MEMORY.md"))
	if !strings.Contains(idx, "feedback_x.md") {
		t.Errorf("primary index missing the mirrored topic:\n%s", idx)
	}
}

// TestMirrorFailsOpenOnUnresolvablePrimary covers AC-AM-006: when the
// primary root cannot be resolved, the hook path emits a stderr notice and
// never blocks — no error propagates out of the wrapper.
//
// NOT parallel: mutates the agentMemoryPrimaryRootFn package seam, and
// captures os.Stderr.
func TestMirrorFailsOpenOnUnresolvablePrimary(t *testing.T) {
	tree := t.TempDir()
	seedWorktreeX(t, tree)
	swapMirrorPrimaryRoot(t, func(string) (string, bool, error) {
		return "", false, os.ErrInvalid
	})

	src := filepath.Join(tree, ".claude", "agent-memory", "manager-spec", "feedback_x.md")
	stderr := captureStderr(t, func() {
		mirrorAgentMemory(mirrorInput(t, src))
	})
	if !strings.Contains(stderr, "memory-mirror") {
		t.Errorf("no fail-open notice on stderr:\n%q", stderr)
	}
}

// TestMirrorNoOpInPrimarySession covers AC-AM-008: when the session's
// project root IS the primary (common dir == own .git), no copy occurs and
// no `.wt-` self-duplicates appear.
//
// NOT parallel: mutates the agentMemoryPrimaryRootFn package seam.
func TestMirrorNoOpInPrimarySession(t *testing.T) {
	primary := t.TempDir()
	seedWorktreeX(t, primary) // the "worktree" is the primary itself
	swapMirrorPrimaryRoot(t, func(dir string) (string, bool, error) {
		return dir, false, nil
	})

	src := filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "feedback_x.md")
	mirrored, err := MirrorAgentMemoryFile(src)
	if err != nil {
		t.Fatalf("MirrorAgentMemoryFile: %v", err)
	}
	if mirrored {
		t.Error("primary-session write was mirrored")
	}
	matches, _ := filepath.Glob(filepath.Join(primary, ".claude", "agent-memory", "manager-spec", "*.wt-*.md"))
	if len(matches) != 0 {
		t.Errorf("primary session produced tree-qualified self-copies: %v", matches)
	}
}

// TestMirrorIgnoresUnanchoredPath pins the D7 negative at the mirror entry:
// a docs/agent-memory path must not mirror even though the audit's looser
// predicate would scan it.
//
// NOT parallel: mutates the agentMemoryPrimaryRootFn package seam.
func TestMirrorIgnoresUnanchoredPath(t *testing.T) {
	called := false
	swapMirrorPrimaryRoot(t, func(string) (string, bool, error) {
		called = true
		return "", false, nil
	})

	mirrored, err := MirrorAgentMemoryFile("/repo/docs/agent-memory/x.md")
	if err != nil || mirrored {
		t.Errorf("unanchored path mirrored (mirrored=%v err=%v)", mirrored, err)
	}
	if called {
		t.Error("primary resolution ran for an unanchored path")
	}
}

// TestMirrorAgentMemoryDispatchedInWriteEditBranch is the wiring pin: the
// PostToolUse Write/Edit branch must invoke the mirror (the mechanism that
// makes the write-time guarantee; a mirror that exists but is never called
// drains nothing).
func TestMirrorAgentMemoryDispatchedInWriteEditBranch(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("post_tool.go")
	if err != nil {
		t.Fatalf("read post_tool.go: %v", err)
	}
	body := string(data)
	writeBranch := strings.Index(body, "runMemoryAudit(input)")
	if writeBranch < 0 {
		t.Fatal("runMemoryAudit call site not found in post_tool.go")
	}
	tail := body[writeBranch:]
	if !strings.Contains(tail, "mirrorAgentMemory(input)") {
		t.Error("the Write/Edit branch of post_tool.go does not call mirrorAgentMemory(input)")
	}
}
