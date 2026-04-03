package worktree

import (
	"errors"
	"testing"
)

// TestWorktreeError_Error tests that WorktreeError implements error interface correctly
func TestWorktreeError_Error(t *testing.T) {
	tests := []struct {
		name      string
		err       *WorktreeError
		wantSubstr string // substring that must be in error message
	}{
		{
			name: "error message includes SPEC-ID",
			err: &WorktreeError{
				prefix:  "Worktree creation failed",
				SpecID:  "SPEC-TEST-001",
				Err:     errors.New("git error"),
				Recovery: "moai worktree new SPEC-TEST-001",
			},
			wantSubstr: "SPEC-TEST-001",
		},
		{
			name: "error message includes recovery command",
			err: &WorktreeError{
				prefix:  "Worktree creation failed",
				SpecID:  "SPEC-TEST-001",
				Err:     errors.New("git error"),
				Recovery: "moai worktree new SPEC-TEST-001",
			},
			wantSubstr: "moai worktree new SPEC-TEST-001",
		},
		{
			name: "error message includes prefix",
			err: &WorktreeError{
				prefix:  "Worktree creation failed",
				SpecID:  "SPEC-TEST-001",
				Err:     errors.New("git error"),
				Recovery: "moai worktree new SPEC-TEST-001",
			},
			wantSubstr: "Worktree creation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !contains(got, tt.wantSubstr) {
				t.Errorf("Error() = %v, want substring %v", got, tt.wantSubstr)
			}
		})
	}
}

// TestWorktreeError_Unwrap tests that Unwrap returns the underlying error
func TestWorktreeError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("git error")
	err := &WorktreeError{
		prefix:   "Worktree creation failed",
		SpecID:   "SPEC-TEST-001",
		Err:      underlyingErr,
		Recovery: "moai worktree new SPEC-TEST-001",
	}

	if got := err.Unwrap(); got != underlyingErr {
		t.Errorf("Unwrap() = %v, want %v", got, underlyingErr)
	}
}

// TestNewWorktreeCreateError tests the worktree creation error constructor
func TestNewWorktreeCreateError(t *testing.T) {
	underlyingErr := errors.New("git worktree add failed")
	specID := "SPEC-TEST-001"

	err := NewWorktreeCreateError(specID, underlyingErr)

	// Verify fields
	if err.SpecID != specID {
		t.Errorf("SpecID = %v, want %v", err.SpecID, specID)
	}
	if err.Err != underlyingErr {
		t.Errorf("Err = %v, want %v", err.Err, underlyingErr)
	}

	// Verify error message contains required elements
	msg := err.Error()
	if !contains(msg, "Worktree creation failed") {
		t.Errorf("Error message should contain 'Worktree creation failed', got: %v", msg)
	}
	if !contains(msg, specID) {
		t.Errorf("Error message should contain SPEC-ID %v, got: %v", specID, msg)
	}
	if !contains(msg, "moai worktree new") {
		t.Errorf("Error message should contain recovery command, got: %v", msg)
	}
}

// TestNewTmuxNotAvailableError tests the tmux not available error constructor
func TestNewTmuxNotAvailableError(t *testing.T) {
	specID := "SPEC-TEST-001"
	worktreePath := "/home/user/.moai/worktrees/project/SPEC-TEST-001"

	err := NewTmuxNotAvailableError(specID, worktreePath)

	// Verify fields
	if err.SpecID != specID {
		t.Errorf("SpecID = %v, want %v", err.SpecID, specID)
	}

	// Verify error message contains manual cd instructions
	msg := err.Error()
	if !contains(msg, "tmux not available") {
		t.Errorf("Error message should contain 'tmux not available', got: %v", msg)
	}
	if !contains(msg, "cd") {
		t.Errorf("Error message should contain 'cd' instruction, got: %v", msg)
	}
	if !contains(msg, worktreePath) {
		t.Errorf("Error message should contain worktree path, got: %v", msg)
	}
	if !contains(msg, "/moai run") {
		t.Errorf("Error message should contain /moai run command, got: %v", msg)
	}
}

// TestNewAutoMergeBlockedError tests the auto-merge blocked error constructor
func TestNewAutoMergeBlockedError(t *testing.T) {
	tests := []struct {
		name           string
		specID         string
		reason         string
		wantSubstrings []string
	}{
		{
			name:   "CI checks failed",
			specID: "SPEC-TEST-001",
			reason: "CI checks failed",
			wantSubstrings: []string{
				"Auto-merge blocked",
				"CI checks failed",
				"/moai sync",
			},
		},
		{
			name:   "Merge conflicts",
			specID: "SPEC-TEST-001",
			reason: "Merge conflicts detected",
			wantSubstrings: []string{
				"Auto-merge blocked",
				"Merge conflicts",
				"manually in PR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAutoMergeBlockedError(tt.specID, tt.reason)

			// Verify fields
			if err.SpecID != tt.specID {
				t.Errorf("SpecID = %v, want %v", err.SpecID, tt.specID)
			}

			// Verify error message contains required elements
			msg := err.Error()
			for _, substr := range tt.wantSubstrings {
				if !contains(msg, substr) {
					t.Errorf("Error message should contain %v, got: %v", substr, msg)
				}
			}
		})
	}
}

// TestNewCleanupFailedError tests the cleanup failed error constructor
func TestNewCleanupFailedError(t *testing.T) {
	underlyingErr := errors.New("permission denied")
	specID := "SPEC-TEST-001"

	err := NewCleanupFailedError(specID, underlyingErr)

	// Verify fields
	if err.SpecID != specID {
		t.Errorf("SpecID = %v, want %v", err.SpecID, specID)
	}
	if err.Err != underlyingErr {
		t.Errorf("Err = %v, want %v", err.Err, underlyingErr)
	}

	// Verify error message contains required elements
	msg := err.Error()
	if !contains(msg, "Worktree cleanup failed") {
		t.Errorf("Error message should contain 'Worktree cleanup failed', got: %v", msg)
	}
	if !contains(msg, specID) {
		t.Errorf("Error message should contain SPEC-ID %v, got: %v", specID, msg)
	}
	if !contains(msg, "moai worktree done") {
		t.Errorf("Error message should contain recovery command, got: %v", msg)
	}
}

// TestErrorInterfaceCompliance tests that all error constructors return types that implement error interface
func TestErrorInterfaceCompliance(t *testing.T) {
	var _ error = &WorktreeError{}

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "NewWorktreeCreateError",
			err:  NewWorktreeCreateError("SPEC-TEST-001", errors.New("test")),
		},
		{
			name: "NewTmuxNotAvailableError",
			err:  NewTmuxNotAvailableError("SPEC-TEST-001", "/path/to/worktree"),
		},
		{
			name: "NewAutoMergeBlockedError",
			err:  NewAutoMergeBlockedError("SPEC-TEST-001", "test reason"),
		},
		{
			name: "NewCleanupFailedError",
			err:  NewCleanupFailedError("SPEC-TEST-001", errors.New("test")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Constructor returned nil")
			}
			// Just calling Error() to verify it implements the interface
			_ = tt.err.Error()
		})
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
