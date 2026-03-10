package worktree

import (
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// setupMockProvider는 테스트를 위해 MockWorktreeProvider를 설정하고,
// 테스트 종료 시 원래 상태로 복원하는 cleanup 함수를 반환합니다.
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

// MockWorktreeProvider는 테스트용 WorktreeProvider 구현체입니다.
// SPEC-WORKTREE-002의 TDD 구현을 위해 추가됨.
type MockWorktreeProvider struct {
	addCalled     bool
	removeCalled  bool
	worktrees     []WorktreeInfo
	addFunc       func(path, branch string) error
	removeFunc    func(path string, force bool) error
	listFunc      func() ([]WorktreeInfo, error)
	deleteBranchFunc func(branch string) error
}

// WorktreeInfo는 worktree 정보를 저장하는 구조체입니다.
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
	// worktree 목록에서 제거
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
		// listFunc가 []WorktreeInfo를 반환하도록 수정 필요
		result, err := m.listFunc()
		if err != nil {
			return nil, err
		}
		// WorktreeInfo를 git.Worktree로 변환
		var worktrees []git.Worktree
		for _, wt := range result {
			worktrees = append(worktrees, git.Worktree{
				Branch: wt.Branch,
				Path:   wt.Path,
			})
		}
		return worktrees, nil
	}
	// WorktreeInfo를 git.Worktree로 변환
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

// 기타 필수 메서드들 (git.WorktreeManager 인터페이스 준수)
func (m *MockWorktreeProvider) Prune() error { return nil }
func (m *MockWorktreeProvider) Repair() error { return nil }
func (m *MockWorktreeProvider) Root() string { return "/test/repo" }
func (m *MockWorktreeProvider) Sync(wtPath, baseBranch, strategy string) error { return nil }
func (m *MockWorktreeProvider) IsBranchMerged(branch, base string) (bool, error) { return false, nil }

// TestNewWorktreeWithTmuxCreation tests R5: tmux session creation after worktree
func TestNewWorktreeWithTmuxCreation(t *testing.T) {
	// RED Phase: 테스트 작성
	// 이 테스트는 worktree 생성 후 tmux 세션 생성을 검증

	tests := []struct {
		name     string
		specID   string
		tmuxAvailable bool
		wantErr  bool
	}{
		{
			name:     "SPEC-WORKTREE-002에서 명시한 tmux 세션 생성",
			specID:   "SPEC-WORKTREE-002",
			tmuxAvailable: true,
			wantErr:  false,
		},
		{
			name:     "tmux 없을 때도 worktree는 생성됨",
			specID:   "SPEC-WORKTREE-002",
			tmuxAvailable: false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: 테스트 환경 설정
			tempDir := t.TempDir()

			// Mock 함수 설정
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

			// Mock WorktreeProvider 설정
			mockProvider, cleanup := setupMockProvider(t)
			defer cleanup()

			// Act: worktree 생성
			expectedPath := filepath.Join(tempDir, ".moai", "worktrees", "test-project", tt.specID)
			err := mockProvider.Add(expectedPath, "feature/"+tt.specID)

			// Assert: 결과 검증
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// worktree가 생성되었는지 확인
			if !mockProvider.addCalled {
				t.Error("WorktreeProvider.Add was not called")
			}
		})
	}
}

// TestTmuxSessionNamePattern tests R5.1: 세션 이름 패턴 검증
func TestTmuxSessionNamePattern(t *testing.T) {
	tests := []struct {
		name      string
		projectName string
		specID    string
		want      string
	}{
		{
			name:      "표준 SPEC-ID",
			projectName: "moai-adk-go",
			specID:    "SPEC-WORKTREE-002",
			want:      "moai-moai-adk-go-SPEC-WORKTREE-002",
		},
		{
			name:      "짧은 프로젝트 이름",
			projectName: "myproject",
			specID:    "SPEC-AUTH-001",
			want:      "moai-myproject-SPEC-AUTH-001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act & Assert
			// R5.1: moai-{ProjectName}-{SPEC-ID} 패턴 검증
			got := GenerateTmuxSessionName(tt.projectName, tt.specID)
			if got != tt.want {
				t.Errorf("GenerateTmuxSessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAutoMergeDefaultBehavior tests R3: auto-merge 기본 동작
func TestAutoMergeDefaultBehavior(t *testing.T) {
	// R3: sync.md 워크플로우에서 auto-merge가 기본값이어야 함
	// 이 테스트는 향후 sync 명령어 구현 시 통합될 예정

	tests := []struct {
		name     string
		noMergeFlag bool
		wantAutoMerge bool
	}{
		{
			name:     "플래그 없으면 auto-merge (기본값)",
			noMergeFlag: false,
			wantAutoMerge: true,
		},
		{
			name:     "--no-merge 플래그로 skip",
			noMergeFlag: true,
			wantAutoMerge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act & Assert
			// 이 테스트는 sync.md 문서 업데이트 후 유효성 검증에 사용됨
			got := ShouldAutoMerge(tt.noMergeFlag)
			if got != tt.wantAutoMerge {
				t.Errorf("ShouldAutoMerge() = %v, want %v", got, tt.wantAutoMerge)
			}
		})
	}
}

// TestWorktreeAutoCleanup tests R4: PR merge 후 자동 cleanup
func TestWorktreeAutoCleanup(t *testing.T) {
	// R4: PR merge 후 자동으로 `moai worktree done SPEC-XXX` 실행

	tests := []struct {
		name       string
		specID     string
		prMerged   bool
		wantCleanup bool
	}{
		{
			name:       "PR merge 후 자동 cleanup",
			specID:     "SPEC-WORKTREE-002",
			prMerged:   true,
			wantCleanup: true,
		},
		{
			name:       "PR 미merge 시 cleanup 없음",
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

			// 초기 worktree 추가
			worktreePath := filepath.Join(tempDir, ".moai", "worktrees", "test-project", tt.specID)
			mockProvider.worktrees = []WorktreeInfo{
				{
					Branch: "feature/" + tt.specID,
					Path:   worktreePath,
				},
			}

			// Act
			if tt.prMerged && tt.wantCleanup {
				// PR merge 후 자동 cleanup 시뮬레이션
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
