package cli

import (
	"log/slog"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

func TestInitDependencies(t *testing.T) {
	// Save and restore original deps
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil
	InitDependencies()

	if deps == nil {
		t.Fatal("InitDependencies should set deps")
	}

	if deps.Config == nil {
		t.Error("deps.Config should not be nil")
	}
	if deps.HookProtocol == nil {
		t.Error("deps.HookProtocol should not be nil")
	}
	if deps.HookRegistry == nil {
		t.Error("deps.HookRegistry should not be nil")
	}
	if deps.Logger == nil {
		t.Error("deps.Logger should not be nil")
	}
}

func TestGetDeps_ReturnsNilBeforeInit(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	deps = nil
	if GetDeps() != nil {
		t.Error("GetDeps should return nil before InitDependencies")
	}
}

func TestGetDeps_ReturnsAfterInit(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	InitDependencies()
	if GetDeps() == nil {
		t.Error("GetDeps should return non-nil after InitDependencies")
	}
}

func TestSetDeps(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()

	custom := &Dependencies{}
	SetDeps(custom)

	if GetDeps() != custom {
		t.Error("SetDeps should replace the global deps")
	}
}

func TestEnsureGit_NotGitRepo(t *testing.T) {
	d := &Dependencies{}
	err := d.EnsureGit(t.TempDir())
	if err == nil {
		t.Error("EnsureGit should error for non-git directory")
	}
}

func TestEnsureGit_AlreadyInitialized(t *testing.T) {
	d := &Dependencies{}
	// Set Git to a non-nil value to simulate already initialized
	d.Git = &mockGitRepository{}

	err := d.EnsureGit("/some/path")
	if err != nil {
		t.Errorf("EnsureGit should return nil when Git is already set: %v", err)
	}
}

func TestEnsureUpdate_Success(t *testing.T) {
	d := &Dependencies{}
	err := d.EnsureUpdate()
	if err != nil {
		t.Errorf("EnsureUpdate should succeed: %v", err)
	}
	if d.UpdateChecker == nil {
		t.Error("EnsureUpdate should set UpdateChecker")
	}
	if d.UpdateOrch == nil {
		t.Error("EnsureUpdate should set UpdateOrch")
	}
}

func TestEnsureUpdate_AlreadyInitialized(t *testing.T) {
	d := &Dependencies{}
	d.UpdateChecker = &mockUpdateChecker{}

	err := d.EnsureUpdate()
	if err != nil {
		t.Errorf("EnsureUpdate should return nil when UpdateChecker is already set: %v", err)
	}
}

// TestInitDependencies_SetsDefaultSlogToDiscard는 InitDependencies 호출 후
// slog.Default() 핸들러가 discard로 교체되어 있음을 검증한다.
// 이 테스트는 이슈 #565의 재현 테스트로, slog.SetDefault 호출이 누락된 경우 실패한다.
// Go 기본 slog 핸들러는 *slog.defaultHandler 타입이며, discard로 설정하면
// *slog.TextHandler로 변경된다.
func TestInitDependencies_SetsDefaultSlogToDiscard(t *testing.T) {
	origDeps := deps
	// slog.Default()도 원상 복구한다
	origDefaultLogger := slog.Default()
	defer func() {
		deps = origDeps
		slog.SetDefault(origDefaultLogger)
	}()

	// InitDependencies 호출 전: Go 기본 핸들러는 *slog.defaultHandler
	deps = nil
	defaultHandlerBefore := slog.Default().Handler()

	InitDependencies()

	// InitDependencies 이후: slog.SetDefault가 호출되었으므로 핸들러가 변경되어야 한다.
	// 변경되지 않았다면 slog.SetDefault 호출이 누락된 것으로, 이슈 #565가 재현된 상황이다.
	defaultHandlerAfter := slog.Default().Handler()
	if defaultHandlerBefore == defaultHandlerAfter {
		t.Error("InitDependencies 이후 slog.Default() 핸들러가 변경되지 않음: slog.SetDefault(logger) 호출이 누락되었을 수 있음 (이슈 #565)")
	}

	// deps.Logger 핸들러와 slog.Default() 핸들러가 동일해야 한다.
	if deps.Logger.Handler() != slog.Default().Handler() {
		t.Error("deps.Logger 핸들러와 slog.Default() 핸들러가 일치하지 않음")
	}
}

// mockGitRepository implements git.Repository for testing EnsureGit
type mockGitRepository struct{}

func (m *mockGitRepository) CurrentBranch() (string, error)   { return "main", nil }
func (m *mockGitRepository) Status() (*git.GitStatus, error)  { return nil, nil }
func (m *mockGitRepository) Log(_ int) ([]git.Commit, error)  { return nil, nil }
func (m *mockGitRepository) Diff(_, _ string) (string, error) { return "", nil }
func (m *mockGitRepository) IsClean() (bool, error)           { return true, nil }
func (m *mockGitRepository) Root() string                     { return "/mock/root" }
