package statusline

import (
	"context"
	"errors"
	"testing"
	"time"

	gitpkg "github.com/modu-ai/moai-adk/internal/core/git"
)

// mockGitRepo implements git.Repository for testing.
type mockGitRepo struct {
	branch    string
	branchErr error
	status    *gitpkg.GitStatus
	statusErr error
}

func (m *mockGitRepo) CurrentBranch() (string, error) {
	return m.branch, m.branchErr
}

func (m *mockGitRepo) Status() (*gitpkg.GitStatus, error) {
	return m.status, m.statusErr
}

func (m *mockGitRepo) Log(_ int) ([]gitpkg.Commit, error) { return nil, nil }
func (m *mockGitRepo) Diff(_, _ string) (string, error)   { return "", nil }
func (m *mockGitRepo) IsClean() (bool, error)             { return true, nil }
func (m *mockGitRepo) Root() string                       { return "/mock" }

func TestGitCollector_CollectGitStatus(t *testing.T) {
	tests := []struct {
		name          string
		repo          gitpkg.Repository
		wantBranch    string
		wantModified  int
		wantStaged    int
		wantUntracked int
		wantAhead     int
		wantBehind    int
		wantAvail     bool
	}{
		{
			name: "valid git status",
			repo: &mockGitRepo{
				branch: "feature/auth",
				status: &gitpkg.GitStatus{
					Modified:  []string{"file1.go", "file2.go", "file3.go"},
					Staged:    []string{"staged1.go", "staged2.go"},
					Untracked: []string{"new.go"},
					Ahead:     1,
					Behind:    0,
				},
			},
			wantBranch:    "feature/auth",
			wantModified:  3,
			wantStaged:    2,
			wantUntracked: 1,
			wantAhead:     1,
			wantBehind:    0,
			wantAvail:     true,
		},
		{
			name:      "nil repo",
			repo:      nil,
			wantAvail: false,
		},
		{
			name: "branch error - graceful degradation",
			repo: &mockGitRepo{
				branchErr: errors.New("detached HEAD"),
				status: &gitpkg.GitStatus{
					Modified: []string{"file.go"},
				},
			},
			wantBranch:   "",
			wantModified: 1,
			wantAvail:    true,
		},
		{
			name: "status error - returns partial data with branch",
			repo: &mockGitRepo{
				branch:    "main",
				statusErr: errors.New("git status failed"),
			},
			wantBranch: "main",
			wantAvail:  true,
		},
		{
			name: "clean repo with no changes",
			repo: &mockGitRepo{
				branch: "main",
				status: &gitpkg.GitStatus{},
			},
			wantBranch: "main",
			wantAvail:  true,
		},
		{
			name: "repo with ahead and behind",
			repo: &mockGitRepo{
				branch: "develop",
				status: &gitpkg.GitStatus{
					Ahead:  3,
					Behind: 2,
				},
			},
			wantBranch: "develop",
			wantAhead:  3,
			wantBehind: 2,
			wantAvail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewGitCollector(tt.repo)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			got, err := collector.CollectGitStatus(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Branch != tt.wantBranch {
				t.Errorf("Branch = %q, want %q", got.Branch, tt.wantBranch)
			}
			if got.Modified != tt.wantModified {
				t.Errorf("Modified = %d, want %d", got.Modified, tt.wantModified)
			}
			if got.Staged != tt.wantStaged {
				t.Errorf("Staged = %d, want %d", got.Staged, tt.wantStaged)
			}
			if got.Untracked != tt.wantUntracked {
				t.Errorf("Untracked = %d, want %d", got.Untracked, tt.wantUntracked)
			}
			if got.Ahead != tt.wantAhead {
				t.Errorf("Ahead = %d, want %d", got.Ahead, tt.wantAhead)
			}
			if got.Behind != tt.wantBehind {
				t.Errorf("Behind = %d, want %d", got.Behind, tt.wantBehind)
			}
			if got.Available != tt.wantAvail {
				t.Errorf("Available = %v, want %v", got.Available, tt.wantAvail)
			}
		})
	}
}
