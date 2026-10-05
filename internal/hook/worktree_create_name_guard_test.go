package hook

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWorktreeCreateHandler_RejectsTraversalName is the regression test for
// the t600 defect: the name "../.." folded the joined path back onto the
// repository root, and the reuse branch then returned the PRIMARY CHECKOUT to
// an agent that had asked for an isolated worktree. The observed pre-fix
// result was err=<nil> with the returned path equal to the repository root.
func TestWorktreeCreateHandler_RejectsTraversalName(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	got, err := NewWorktreeCreateHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-t600",
		AgentName:     "probe",
		WorktreeName:  "../..",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	})
	if err == nil {
		t.Fatalf("Handle(name=%q) = %+v, want error", "../..", got)
	}
	if got != nil && got.WorktreePath == repo {
		t.Fatalf("Handle returned the primary checkout %q", repo)
	}
}

// TestValidateWorktreeName covers one case per rejection axis plus the
// positive control: a legitimate multi-segment name stays accepted, because
// Claude Code's name contract declares "/"-separated segments valid.
func TestValidateWorktreeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		axis    string
	}{
		{name: "parent traversal", input: "../..", wantErr: true, axis: "relative segment"},
		{name: "bare parent", input: "..", wantErr: true, axis: "relative segment"},
		{name: "dot segment", input: "feat/./x", wantErr: true, axis: "relative segment"},
		{name: "leading separator", input: "/abs", wantErr: true, axis: "empty segment"},
		{name: "doubled separator", input: "feat//x", wantErr: true, axis: "empty segment"},
		{name: "trailing separator", input: "feat/", wantErr: true, axis: "empty segment"},
		{name: "backslash separator", input: `feat\x`, wantErr: true, axis: "alphabet"},
		{name: "space", input: "feat x", wantErr: true, axis: "alphabet"},
		{name: "shell metacharacter", input: "feat;rm", wantErr: true, axis: "alphabet"},
		{name: "single segment", input: "backend-impl", wantErr: false, axis: "positive control"},
		{name: "multi segment", input: "feat/x", wantErr: false, axis: "positive control"},
		{name: "dotted segment", input: "v1.2_x-y", wantErr: false, axis: "positive control"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateWorktreeName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("validateWorktreeName(%q) = nil, want error (%s)", tt.input, tt.axis)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateWorktreeName(%q) = %v, want nil (%s)", tt.input, err, tt.axis)
			}
		})
	}
}

// TestEnsureWithinWorktreeParent measures the containment backstop directly —
// the axis that must hold even if the name check is later loosened.
func TestEnsureWithinWorktreeParent(t *testing.T) {
	t.Parallel()

	parent := filepath.Join("/repo", ".claude", "worktrees")
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "inside", path: filepath.Join(parent, "a"), wantErr: false},
		{name: "nested inside", path: filepath.Join(parent, "a", "b"), wantErr: false},
		{name: "the parent itself", path: parent, wantErr: true},
		{name: "repository root", path: "/repo", wantErr: true},
		{name: "sibling escape", path: "/repo/.claude/other", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ensureWithinWorktreeParent(parent, tt.path)
			if tt.wantErr && err == nil {
				t.Fatalf("ensureWithinWorktreeParent(%q, %q) = nil, want error", parent, tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ensureWithinWorktreeParent(%q, %q) = %v, want nil", parent, tt.path, err)
			}
		})
	}
}

// TestWorktreeCreateHandler_MultiSegmentNameStillCreates is the end-to-end
// positive control: the guard must not refuse a name the contract admits.
func TestWorktreeCreateHandler_MultiSegmentNameStillCreates(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	got, err := NewWorktreeCreateHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-t600-ok",
		AgentName:     "probe",
		WorktreeName:  "feat/x",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	})
	if err != nil {
		t.Fatalf("Handle(name=%q): %v", "feat/x", err)
	}

	want := filepath.Join(repo, ".claude", "worktrees", "feat", "x")
	if got.WorktreePath != want {
		t.Fatalf("WorktreePath = %q, want %q", got.WorktreePath, want)
	}
	if info, statErr := os.Stat(filepath.Join(want, ".git")); statErr != nil || info.IsDir() {
		t.Fatalf("%s is not a git worktree (stat err %v)", want, statErr)
	}
}

// TestWorktreeCreateHandler_RejectsUnregisteredDirectoryReuse pins the second
// half of the damage path: the reuse branch keyed on "a directory exists
// here", so any plain directory sitting at the target path was handed back as
// though it were an isolated worktree.
func TestWorktreeCreateHandler_RejectsUnregisteredDirectoryReuse(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	plain := filepath.Join(repo, ".claude", "worktrees", "squatter")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := NewWorktreeCreateHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-t600-squat",
		AgentName:     "probe",
		WorktreeName:  "squatter",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	})
	if err == nil {
		t.Fatal("Handle reused a plain directory as a worktree, want error")
	}
	if !strings.Contains(err.Error(), "not a registered worktree") {
		t.Fatalf("error = %v, want a not-registered-worktree error", err)
	}
}

// TestWorktreeCreateHandler_RejectsSymlinkReuse covers the external-link axis:
// a symlink at the target path resolves elsewhere, so reusing it hands the
// agent a tree outside the worktree parent.
func TestWorktreeCreateHandler_RejectsSymlinkReuse(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	outside := t.TempDir()
	parent := filepath.Join(repo, ".claude", "worktrees")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(parent, "linked")); err != nil {
		t.Skipf("symlink unsupported here: %v", err)
	}

	_, err := NewWorktreeCreateHandler().Handle(context.Background(), &HookInput{
		SessionID:     "sess-t600-link",
		AgentName:     "probe",
		WorktreeName:  "linked",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	})
	if err == nil {
		t.Fatal("Handle reused a symlink as a worktree, want error")
	}
	// Pinned to the symlink guard specifically. Without this the assertion
	// passes on the containment backstop firing downstream, so removing the
	// symlink branch would leave the test green and the axis unasserted.
	if !strings.Contains(err.Error(), "is a symlink") {
		t.Fatalf("error = %v, want the symlink guard to reject it", err)
	}
}

// TestWorktreeCreateHandler_ReusesRegisteredWorktree is the reuse-path
// positive control: idempotent reuse must survive the new registry check.
func TestWorktreeCreateHandler_ReusesRegisteredWorktree(t *testing.T) {
	repo := initWorktreeTestRepo(t)

	h := NewWorktreeCreateHandler()
	input := &HookInput{
		SessionID:     "sess-t600-reuse",
		AgentName:     "probe",
		WorktreeName:  "reused",
		CWD:           repo,
		HookEventName: "WorktreeCreate",
	}

	first, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("first Handle: %v", err)
	}
	second, err := h.Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("second Handle (reuse): %v", err)
	}
	if first.WorktreePath != second.WorktreePath {
		t.Fatalf("reuse returned %q, want %q", second.WorktreePath, first.WorktreePath)
	}
}
