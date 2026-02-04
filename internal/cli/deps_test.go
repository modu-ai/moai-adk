package cli

import (
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
	if deps.RankCredStore == nil {
		t.Error("deps.RankCredStore should not be nil")
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

func TestEnsureRank_NilCredStore(t *testing.T) {
	d := &Dependencies{}
	err := d.EnsureRank()
	if err == nil {
		t.Error("EnsureRank should error when RankCredStore is nil")
	}
}

func TestEnsureRank_NoAPIKey(t *testing.T) {
	d := &Dependencies{
		RankCredStore: &mockCredentialStore{
			getKeyFunc: func() (string, error) { return "", nil },
		},
	}
	err := d.EnsureRank()
	if err == nil {
		t.Error("EnsureRank should error when no API key is configured")
	}
}

func TestEnsureRank_Success(t *testing.T) {
	d := &Dependencies{
		RankCredStore: &mockCredentialStore{
			getKeyFunc: func() (string, error) { return "test-key", nil },
		},
	}
	err := d.EnsureRank()
	if err != nil {
		t.Errorf("EnsureRank should succeed: %v", err)
	}
	if d.RankClient == nil {
		t.Error("EnsureRank should set RankClient")
	}
}

func TestEnsureRank_AlreadyInitialized(t *testing.T) {
	d := &Dependencies{}
	d.RankClient = &mockRankClient{}

	err := d.EnsureRank()
	if err != nil {
		t.Errorf("EnsureRank should return nil when RankClient is already set: %v", err)
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
