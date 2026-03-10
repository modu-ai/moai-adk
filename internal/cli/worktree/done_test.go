package worktree

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// TestRunDoneWithAutoMode_AfterMerge tests auto-cleanup after PR merge.
func TestRunDoneWithAutoMode_AfterMerge(t *testing.T) {
	tests := []struct {
		name         string
		worktrees    []git.Worktree
		specID       string // Use specID which will be resolved to feature/SPEC-XXX
		wantSuccess  bool
		shouldRemove bool
	}{
		{
			name: "PR merged - worktree removed",
			worktrees: []git.Worktree{
				{Branch: "feature/SPEC-TEST-001", Path: "/worktree/SPEC-TEST-001"},
			},
			specID:       "SPEC-TEST-001",
			wantSuccess:  true,
			shouldRemove: true,
		},
		{
			name: "No worktree found - returns success (graceful)",
			worktrees: []git.Worktree{
				{Branch: "feature/SPEC-OTHER", Path: "/worktree/OTHER"},
			},
			specID:       "SPEC-TEST-001",
			wantSuccess:  true,
			shouldRemove: false,
		},
		{
			name:        "No worktrees at all",
			worktrees:   []git.Worktree{},
			specID:      "SPEC-TEST-001",
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock provider
			origProvider := WorktreeProvider
			mock := &mockWorktreeProvider{
				worktrees: make([]git.Worktree, len(tt.worktrees)),
			}
			copy(mock.worktrees, tt.worktrees)
			WorktreeProvider = mock
			defer func() { WorktreeProvider = origProvider }()

			// Act - runDoneWithAutoMode takes a branch name, not specID
			// We need to resolve the specID to branch name first
			branchName := resolveSpecBranch(tt.specID)

			success, err := runDoneWithAutoMode(branchName, false, false, true)

			// Assert
			if err != nil {
				t.Errorf("runDoneWithAutoMode() unexpected error = %v", err)
			}
			if success != tt.wantSuccess {
				t.Errorf("runDoneWithAutoMode() success = %v, want %v", success, tt.wantSuccess)
			}
			if tt.shouldRemove && !mock.removeCalled {
				t.Error("Remove() was not called")
			}
		})
	}
}

// TestRunDoneWithAutoMode_NoProvider tests graceful degradation when provider is nil.
func TestRunDoneWithAutoMode_NoProvider(t *testing.T) {
	// Setup: nil provider
	origProvider := WorktreeProvider
	WorktreeProvider = nil
	defer func() { WorktreeProvider = origProvider }()

	// Act
	success, err := runDoneWithAutoMode("SPEC-TEST-001", false, false, true)

	// Assert: auto mode should return gracefully with false (not success)
	if err != nil {
		t.Errorf("runDoneWithAutoMode() should not error in auto mode, got = %v", err)
	}
	// In auto mode with nil provider, we return (false, nil) for graceful degradation
	if success {
		t.Errorf("runDoneWithAutoMode() should return false when provider is nil, got true")
	}
}

// TestRunDoneWithAutoMode_DeleteBranch tests branch deletion in auto mode.
func TestRunDoneWithAutoMode_DeleteBranch(t *testing.T) {
	// Setup mock provider with worktree
	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{
		worktrees: []git.Worktree{
			{Branch: "feature/SPEC-TEST-001", Path: "/worktree/SPEC-TEST-001"},
		},
	}
	WorktreeProvider = mock
	defer func() { WorktreeProvider = origProvider }()

	// Act with deleteBranch=true - use resolved branch name
	branchName := resolveSpecBranch("SPEC-TEST-001")
	_, err := runDoneWithAutoMode(branchName, false, true, true)

	// Assert
	if err != nil {
		t.Errorf("runDoneWithAutoMode() unexpected error = %v", err)
	}
	if !mock.deleteBranchCalled {
		t.Error("DeleteBranch() was not called")
	}
}

// mockWorktreeProvider is a minimal mock for auto-cleanup tests.
// Uses a different name to avoid conflict with new_test.go's MockWorktreeProvider.
type mockWorktreeProvider struct {
	removeCalled       bool
	deleteBranchCalled bool
	worktrees          []git.Worktree
}

func (m *mockWorktreeProvider) Add(path, branch string) error {
	return nil
}

func (m *mockWorktreeProvider) Remove(path string, force bool) error {
	m.removeCalled = true
	// Remove from slice
	for i, wt := range m.worktrees {
		if wt.Path == path {
			m.worktrees = append(m.worktrees[:i], m.worktrees[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockWorktreeProvider) List() ([]git.Worktree, error) {
	return m.worktrees, nil
}

func (m *mockWorktreeProvider) DeleteBranch(branch string) error {
	m.deleteBranchCalled = true
	return nil
}

func (m *mockWorktreeProvider) Prune() error  { return nil }
func (m *mockWorktreeProvider) Repair() error { return nil }
func (m *mockWorktreeProvider) Root() string  { return "/test/repo" }
func (m *mockWorktreeProvider) Sync(wtPath, baseBranch, strategy string) error {
	return nil
}
func (m *mockWorktreeProvider) IsBranchMerged(branch, base string) (bool, error) {
	return false, nil
}
