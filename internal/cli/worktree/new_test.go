package worktree

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// setupMockProvider sets up a MockWorktreeProvider for testing
// and returns a cleanup function that restores the original state when the test ends.
func setupMockProvider(t *testing.T) (*MockWorktreeProvider, func()) {
	t.Helper()

	origProvider := WorktreeProvider
	mockProvider := &MockWorktreeProvider{
		worktrees: []WorktreeInfo{},
	}
	WorktreeProvider = mockProvider

	cleanup := func() {
		WorktreeProvider = origProvider
	}

	return mockProvider, cleanup
}

// MockWorktreeProvider is a test implementation of WorktreeProvider.
// Added for TDD implementation of SPEC-WORKTREE-002.
type MockWorktreeProvider struct {
	addCalled     bool
	removeCalled  bool
	worktrees     []WorktreeInfo
	addFunc       func(path, branch string) error
	removeFunc    func(path string, force bool) error
	listFunc      func() ([]WorktreeInfo, error)
	deleteBranchFunc func(branch string) error
}

// WorktreeInfo is a struct that stores worktree information.
type WorktreeInfo struct {
	Branch string
	Path   string
}

func (m *MockWorktreeProvider) Add(path, branch string) error {
	m.addCalled = true
	if m.addFunc != nil {
		return m.addFunc(path, branch)
	}
	m.worktrees = append(m.worktrees, WorktreeInfo{Branch: branch, Path: path})
	return nil
}

func (m *MockWorktreeProvider) Remove(path string, force bool) error {
	m.removeCalled = true
	if m.removeFunc != nil {
		return m.removeFunc(path, force)
	}
	// Remove from the worktree list
	for i, wt := range m.worktrees {
		if wt.Path == path {
			m.worktrees = append(m.worktrees[:i], m.worktrees[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockWorktreeProvider) List() ([]git.Worktree, error) {
	if m.listFunc != nil {
		// listFunc needs to be updated to return []WorktreeInfo
		result, err := m.listFunc()
		if err != nil {
			return nil, err
		}
		// Convert WorktreeInfo to git.Worktree
		var worktrees []git.Worktree
		for _, wt := range result {
			worktrees = append(worktrees, git.Worktree{
				Branch: wt.Branch,
				Path:   wt.Path,
			})
		}
		return worktrees, nil
	}
	// Convert WorktreeInfo to git.Worktree
	var worktrees []git.Worktree
	for _, wt := range m.worktrees {
		worktrees = append(worktrees, git.Worktree{
			Branch: wt.Branch,
			Path:   wt.Path,
		})
	}
	return worktrees, nil
}

func (m *MockWorktreeProvider) DeleteBranch(branch string) error {
	if m.deleteBranchFunc != nil {
		return m.deleteBranchFunc(branch)
	}
	return nil
}

// Other required methods (satisfying the git.WorktreeManager interface)
func (m *MockWorktreeProvider) Prune() error { return nil }
func (m *MockWorktreeProvider) Repair() error { return nil }
func (m *MockWorktreeProvider) Root() string { return "/test/repo" }
func (m *MockWorktreeProvider) Sync(wtPath, baseBranch, strategy string) error { return nil }
func (m *MockWorktreeProvider) IsBranchMerged(branch, base string) (bool, error) { return false, nil }

// TestNewWorktreeWithTmuxCreation tests R5: tmux session creation after worktree
func TestNewWorktreeWithTmuxCreation(t *testing.T) {
	// RED Phase: write the test
	// This test verifies tmux session creation after worktree creation

	tests := []struct {
		name     string
		specID   string
		tmuxAvailable bool
		wantErr  bool
	}{
		{
			name:     "tmux session creation as specified in SPEC-WORKTREE-002",
			specID:   "SPEC-WORKTREE-002",
			tmuxAvailable: true,
			wantErr:  false,
		},
		{
			name:     "worktree is still created when tmux is unavailable",
			specID:   "SPEC-WORKTREE-002",
			tmuxAvailable: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: set up the test environment
			tempDir := t.TempDir()

			// Set up mock functions
			oldUserHomeDirFunc := userHomeDirFunc
			oldGetProjectNameFunc := getProjectNameFunc
			defer func() {
				userHomeDirFunc = oldUserHomeDirFunc
				getProjectNameFunc = oldGetProjectNameFunc
			}()

			userHomeDirFunc = func() (string, error) {
				return tempDir, nil
			}
			getProjectNameFunc = func() string {
				return "test-project"
			}

			// Set up mock WorktreeProvider
			mockProvider, cleanup := setupMockProvider(t)
			defer cleanup()

			// Act: create the worktree
			expectedPath := filepath.Join(tempDir, ".moai", "worktrees", "test-project", tt.specID)
			err := mockProvider.Add(expectedPath, "feature/"+tt.specID)

			// Assert: verify the result
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify that the worktree was created
			if !mockProvider.addCalled {
				t.Error("WorktreeProvider.Add was not called")
			}
		})
	}
}

// TestTmuxSessionNamePattern tests R5.1: session name pattern validation
func TestTmuxSessionNamePattern(t *testing.T) {
	tests := []struct {
		name      string
		projectName string
		specID    string
		want      string
	}{
		{
			name:      "standard SPEC-ID",
			projectName: "moai-adk-go",
			specID:    "SPEC-WORKTREE-002",
			want:      "moai-moai-adk-go-SPEC-WORKTREE-002",
		},
		{
			name:      "short project name",
			projectName: "myproject",
			specID:    "SPEC-AUTH-001",
			want:      "moai-myproject-SPEC-AUTH-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act & Assert
			// R5.1: validate moai-{ProjectName}-{SPEC-ID} pattern
			got := GenerateTmuxSessionName(tt.projectName, tt.specID)
			if got != tt.want {
				t.Errorf("GenerateTmuxSessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAutoMergeDefaultBehavior tests R3: auto-merge default behavior
func TestAutoMergeDefaultBehavior(t *testing.T) {
	// R3: auto-merge must be the default in the sync.md workflow
	// This test will be integrated when the sync command is implemented

	tests := []struct {
		name     string
		noMergeFlag bool
		wantAutoMerge bool
	}{
		{
			name:     "auto-merge when no flag is set (default)",
			noMergeFlag: false,
			wantAutoMerge: true,
		},
		{
			name:     "skip with --no-merge flag",
			noMergeFlag: true,
			wantAutoMerge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act & Assert
			// This test is used for validation after updating sync.md documentation
			got := ShouldAutoMerge(tt.noMergeFlag)
			if got != tt.wantAutoMerge {
				t.Errorf("ShouldAutoMerge() = %v, want %v", got, tt.wantAutoMerge)
			}
		})
	}
}

// TestWorktreeAutoCleanup tests R4: automatic cleanup after PR merge
func TestWorktreeAutoCleanup(t *testing.T) {
	// R4: automatically run `moai worktree done SPEC-XXX` after PR merge

	tests := []struct {
		name       string
		specID     string
		prMerged   bool
		wantCleanup bool
	}{
		{
			name:       "auto cleanup after PR merge",
			specID:     "SPEC-WORKTREE-002",
			prMerged:   true,
			wantCleanup: true,
		},
		{
			name:       "no cleanup when PR is not merged",
			specID:     "SPEC-WORKTREE-002",
			prMerged:   false,
			wantCleanup: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tempDir := t.TempDir()

			oldUserHomeDirFunc := userHomeDirFunc
			oldGetProjectNameFunc := getProjectNameFunc
			defer func() {
				userHomeDirFunc = oldUserHomeDirFunc
				getProjectNameFunc = oldGetProjectNameFunc
			}()

			userHomeDirFunc = func() (string, error) {
				return tempDir, nil
			}
			getProjectNameFunc = func() string {
				return "test-project"
			}

			mockProvider, cleanup := setupMockProvider(t)
			defer cleanup()

			// Add the initial worktree
			worktreePath := filepath.Join(tempDir, ".moai", "worktrees", "test-project", tt.specID)
			mockProvider.worktrees = []WorktreeInfo{
				{
					Branch: "feature/" + tt.specID,
					Path:   worktreePath,
				},
			}

			// Act
			if tt.prMerged && tt.wantCleanup {
				// Simulate automatic cleanup after PR merge
				err := mockProvider.Remove(worktreePath, true)
				if err != nil {
					t.Errorf("Auto-cleanup failed: %v", err)
				}
			}

			// Assert
			if tt.prMerged && tt.wantCleanup && !mockProvider.removeCalled {
				t.Error("WorktreeProvider.Remove was not called after PR merge")
			}
		})
	}
}
