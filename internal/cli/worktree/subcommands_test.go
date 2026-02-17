package worktree

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// mockWorktreeManager implements git.WorktreeManager for testing.
type mockWorktreeManager struct {
	addFunc    func(path, branch string) error
	listFunc   func() ([]git.Worktree, error)
	removeFunc func(path string, force bool) error
	pruneFunc  func() error
	repairFunc func() error
	rootPath   string
}

func (m *mockWorktreeManager) Add(path, branch string) error {
	if m.addFunc != nil {
		return m.addFunc(path, branch)
	}
	return nil
}

func (m *mockWorktreeManager) List() ([]git.Worktree, error) {
	if m.listFunc != nil {
		return m.listFunc()
	}
	return nil, nil
}

func (m *mockWorktreeManager) Remove(path string, force bool) error {
	if m.removeFunc != nil {
		return m.removeFunc(path, force)
	}
	return nil
}

func (m *mockWorktreeManager) Prune() error {
	if m.pruneFunc != nil {
		return m.pruneFunc()
	}
	return nil
}

func (m *mockWorktreeManager) Repair() error {
	if m.repairFunc != nil {
		return m.repairFunc()
	}
	return nil
}

func (m *mockWorktreeManager) Root() string {
	if m.rootPath != "" {
		return m.rootPath
	}
	return "/repo"
}

// --- Tests for runNew ---

func TestRunNew_Success(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	var capturedPath, capturedBranch string
	WorktreeProvider = &mockWorktreeManager{
		addFunc: func(path, branch string) error {
			capturedPath = path
			capturedBranch = branch
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"feature-x"})
			if err != nil {
				t.Fatalf("runNew error: %v", err)
			}

			if capturedBranch != "feature-x" {
				t.Errorf("branch = %q, want %q", capturedBranch, "feature-x")
			}
			if capturedPath == "" {
				t.Error("path should not be empty")
			}
			if !strings.Contains(buf.String(), "Created worktree") {
				t.Errorf("output should contain 'Created worktree', got %q", buf.String())
			}
			if !strings.Contains(buf.String(), "feature-x") {
				t.Errorf("output should contain branch name, got %q", buf.String())
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

func TestRunNew_AddError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		addFunc: func(_, _ string) error {
			return errors.New("path already exists")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			err := cmd.RunE(cmd, []string{"feature-x"})
			if err == nil {
				t.Error("runNew should error when Add fails")
			}
			if !strings.Contains(err.Error(), "create worktree") {
				t.Errorf("error should mention create worktree, got %v", err)
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

// --- Tests for runList ---

func TestRunList_WithWorktrees(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc12345def67890"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "789abcde00012345"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runList error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Active Worktrees") {
				t.Errorf("output should contain 'Active Worktrees', got %q", output)
			}
			if !strings.Contains(output, "main") {
				t.Errorf("output should contain 'main', got %q", output)
			}
			if !strings.Contains(output, "feature") {
				t.Errorf("output should contain 'feature', got %q", output)
			}
			if !strings.Contains(output, "/repo") {
				t.Errorf("output should contain path, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunList_Empty(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runList error: %v", err)
			}

			if !strings.Contains(buf.String(), "No worktrees found") {
				t.Errorf("output should say no worktrees, got %q", buf.String())
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunList_EmptySlice(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runList error: %v", err)
			}

			if !strings.Contains(buf.String(), "No worktrees found") {
				t.Errorf("output should say no worktrees, got %q", buf.String())
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunList_Error(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, errors.New("git error")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("runList should error when List fails")
			}
			if !strings.Contains(err.Error(), "list worktrees") {
				t.Errorf("error should mention list worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestRunList_ShortHEAD(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runList error: %v", err)
			}

			if !strings.Contains(buf.String(), "abc") {
				t.Errorf("output should contain short HEAD, got %q", buf.String())
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

// --- Tests for runSwitch ---

func TestRunSwitch_Found(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feat", Branch: "feat", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"feat"})
			if err != nil {
				t.Fatalf("runSwitch error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "/repo-feat") {
				t.Errorf("output should contain path, got %q", output)
			}
			if !strings.Contains(output, "feat") {
				t.Errorf("output should contain branch name, got %q", output)
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

func TestRunSwitch_NotFound(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			err := cmd.RunE(cmd, []string{"nonexistent"})
			if err == nil {
				t.Error("runSwitch should error for unknown branch")
			}
			if !strings.Contains(err.Error(), "no worktree found") {
				t.Errorf("error should mention no worktree, got %v", err)
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

func TestRunSwitch_ListError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, errors.New("git error")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			err := cmd.RunE(cmd, []string{"feat"})
			if err == nil {
				t.Error("runSwitch should error on list failure")
			}
			if !strings.Contains(err.Error(), "list worktrees") {
				t.Errorf("error should mention list worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

func TestRunSwitch_EmptyList(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			err := cmd.RunE(cmd, []string{"feat"})
			if err == nil {
				t.Error("runSwitch should error when no worktrees exist")
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

// --- Tests for runSync ---

func TestRunSync_WithProvider(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runSync error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Syncing worktree") {
				t.Errorf("output should contain 'Syncing worktree', got %q", output)
			}
			if !strings.Contains(output, "Sync complete") {
				t.Errorf("output should contain 'Sync complete', got %q", output)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

// --- Tests for runRemove ---

func TestRunRemove_Success(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	var capturedPath string
	var capturedForce bool
	WorktreeProvider = &mockWorktreeManager{
		removeFunc: func(path string, force bool) error {
			capturedPath = path
			capturedForce = force
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "remove" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"/tmp/test-wt"})
			if err != nil {
				t.Fatalf("runRemove error: %v", err)
			}

			if capturedPath != "/tmp/test-wt" {
				t.Errorf("path = %q, want %q", capturedPath, "/tmp/test-wt")
			}
			if capturedForce {
				t.Errorf("force = %v, want false", capturedForce)
			}
			if !strings.Contains(buf.String(), "Removed worktree") {
				t.Errorf("output should contain 'Removed worktree', got %q", buf.String())
			}
			return
		}
	}
	t.Error("remove subcommand not found")
}

func TestRunRemove_Error(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		removeFunc: func(_ string, _ bool) error {
			return errors.New("dirty worktree")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "remove" {
			err := cmd.RunE(cmd, []string{"/tmp/test-wt"})
			if err == nil {
				t.Error("runRemove should error when Remove fails")
			}
			if !strings.Contains(err.Error(), "remove worktree") {
				t.Errorf("error should mention remove worktree, got %v", err)
			}
			return
		}
	}
	t.Error("remove subcommand not found")
}

// --- Tests for runClean ---

func TestRunClean_Success(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runClean error: %v", err)
			}

			if !strings.Contains(buf.String(), "Cleaned stale") {
				t.Errorf("output should contain 'Cleaned stale', got %q", buf.String())
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

func TestRunClean_PruneError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		pruneFunc: func() error {
			return errors.New("prune failed")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("runClean should error when Prune fails")
			}
			if !strings.Contains(err.Error(), "prune worktrees") {
				t.Errorf("error should mention prune worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

// --- Tests for subcommand descriptions ---

func TestWorktreeCmd_SubcommandsHaveLongDesc(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Long == "" {
			t.Errorf("worktree subcommand %q should have a Long description", cmd.Name())
		}
	}
}

func TestWorktreeCmd_SubcommandsHaveRunE(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.RunE == nil {
			t.Errorf("worktree subcommand %q should have RunE set", cmd.Name())
		}
	}
}

func TestWorktreeCmd_SwitchRequiresArg(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Error("worktree switch should require an argument")
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

func TestWorktreeCmd_NewHasPathFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			f := cmd.Flags().Lookup("path")
			if f == nil {
				t.Error("worktree new should have --path flag")
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

func TestWorktreeCmd_RemoveHasForceFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "remove" {
			f := cmd.Flags().Lookup("force")
			if f == nil {
				t.Error("worktree remove should have --force flag")
			}
			return
		}
	}
	t.Error("remove subcommand not found")
}

// --- Tests for runRecover ---

func TestRunRecover_Success(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	repairCalled := false
	pruneCalled := false
	WorktreeProvider = &mockWorktreeManager{
		repairFunc: func() error {
			repairCalled = true
			return nil
		},
		pruneFunc: func() error {
			pruneCalled = true
			return nil
		},
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc12345"},
			}, nil
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "recover" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runRecover error: %v", err)
			}

			if !repairCalled {
				t.Error("Repair should have been called")
			}
			if !pruneCalled {
				t.Error("Prune should have been called")
			}
			if !strings.Contains(buf.String(), "Recovered") {
				t.Errorf("output should contain 'Recovered', got %q", buf.String())
			}
			return
		}
	}
	t.Error("recover subcommand not found")
}

func TestRunRecover_RepairError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		repairFunc: func() error {
			return errors.New("repair failed")
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "recover" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("runRecover should error when Repair fails")
			}
			if !strings.Contains(err.Error(), "repair worktrees") {
				t.Errorf("error should mention repair worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("recover subcommand not found")
}

// --- Tests for runDone ---

func TestRunDone_Success(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	removeCalled := false
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "def"},
			}, nil
		},
		removeFunc: func(path string, force bool) error {
			removeCalled = true
			if path != "/repo-feature" {
				t.Errorf("path = %q, want %q", path, "/repo-feature")
			}
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"feature"})
			if err != nil {
				t.Fatalf("runDone error: %v", err)
			}

			if !removeCalled {
				t.Error("Remove should have been called")
			}
			if !strings.Contains(buf.String(), "Done") {
				t.Errorf("output should contain 'Done', got %q", buf.String())
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

func TestRunDone_BranchNotFound(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			err := cmd.RunE(cmd, []string{"nonexistent"})
			if err == nil {
				t.Error("runDone should error for unknown branch")
			}
			if !strings.Contains(err.Error(), "no worktree found") {
				t.Errorf("error should mention no worktree, got %v", err)
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

func TestRunDone_HasForceFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			f := cmd.Flags().Lookup("force")
			if f == nil {
				t.Error("worktree done should have --force flag")
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

func TestRunDone_HasDeleteBranchFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			f := cmd.Flags().Lookup("delete-branch")
			if f == nil {
				t.Error("worktree done should have --delete-branch flag")
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

// --- Tests for runConfig ---

func TestRunConfig_ShowAll(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/test/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "config" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runConfig error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Worktree Configuration") {
				t.Errorf("output should contain 'Worktree Configuration', got %q", output)
			}
			if !strings.Contains(output, "/test/repo") {
				t.Errorf("output should contain root path, got %q", output)
			}
			return
		}
	}
	t.Error("config subcommand not found")
}

func TestRunConfig_ShowRoot(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/test/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "config" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"root"})
			if err != nil {
				t.Fatalf("runConfig error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Worktree root") {
				t.Errorf("output should contain 'Worktree root', got %q", output)
			}
			if !strings.Contains(output, "/test/repo") {
				t.Errorf("output should contain root path, got %q", output)
			}
			return
		}
	}
	t.Error("config subcommand not found")
}

func TestRunConfig_UnknownKey(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/test/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "config" {
			err := cmd.RunE(cmd, []string{"unknown"})
			if err == nil {
				t.Error("runConfig should error for unknown key")
			}
			if !strings.Contains(err.Error(), "unknown config key") {
				t.Errorf("error should mention unknown config key, got %v", err)
			}
			return
		}
	}
	t.Error("config subcommand not found")
}

// --- Tests for runStatus ---

func TestRunStatus_WithWorktrees(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	pruneCalled := false
	WorktreeProvider = &mockWorktreeManager{
		pruneFunc: func() error {
			pruneCalled = true
			return nil
		},
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc12345def67890"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "789abcde00012345"},
			}, nil
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runStatus error: %v", err)
			}

			if !pruneCalled {
				t.Error("Prune should have been called")
			}

			output := buf.String()
			if !strings.Contains(output, "Total worktrees: 2") {
				t.Errorf("output should contain 'Total worktrees: 2', got %q", output)
			}
			if !strings.Contains(output, "main") {
				t.Errorf("output should contain 'main', got %q", output)
			}
			if !strings.Contains(output, "feature") {
				t.Errorf("output should contain 'feature', got %q", output)
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRunStatus_Empty(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, nil
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runStatus error: %v", err)
			}

			if !strings.Contains(buf.String(), "No worktrees found") {
				t.Errorf("output should say no worktrees, got %q", buf.String())
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRunStatus_PruneError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		pruneFunc: func() error {
			return errors.New("prune failed")
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("runStatus should error when Prune fails")
			}
			if !strings.Contains(err.Error(), "prune worktrees") {
				t.Errorf("error should mention prune worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestRunStatus_DetachedHEAD(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "", HEAD: "abc12345"},
			}, nil
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runStatus error: %v", err)
			}

			if !strings.Contains(buf.String(), "(detached)") {
				t.Errorf("output should contain '(detached)', got %q", buf.String())
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

// --- Tests for go subcommand ---

func TestRunGo_Found(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feat", Branch: "feat", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"feat"})
			if err != nil {
				t.Fatalf("runGo error: %v", err)
			}

			output := buf.String()
			if output != "/repo-feat\n" {
				t.Errorf("output = %q, want %q", output, "/repo-feat\n")
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

func TestRunGo_NotFound(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			err := cmd.RunE(cmd, []string{"nonexistent"})
			if err == nil {
				t.Error("runGo should error for unknown branch")
			}
			if !strings.Contains(err.Error(), "no worktree found") {
				t.Errorf("error should mention no worktree found, got %v", err)
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

func TestRunGo_ListError(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, errors.New("git error")
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			err := cmd.RunE(cmd, []string{"feat"})
			if err == nil {
				t.Error("runGo should error on list failure")
			}
			if !strings.Contains(err.Error(), "list worktrees") {
				t.Errorf("error should mention list worktrees, got %v", err)
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

// --- Tests for go with SPEC-ID ---

func TestRunGo_SpecID(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-auth", Branch: "feature/SPEC-AUTH-001", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Pass raw SPEC ID; should resolve to feature/SPEC-AUTH-001
			err := cmd.RunE(cmd, []string{"SPEC-AUTH-001"})
			if err != nil {
				t.Fatalf("runGo error: %v", err)
			}

			output := buf.String()
			if output != "/repo-auth\n" {
				t.Errorf("output = %q, want %q", output, "/repo-auth\n")
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

// --- Tests for switch with SPEC-ID ---

func TestRunSwitch_SpecID(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-auth", Branch: "feature/SPEC-AUTH-001", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Pass raw SPEC ID; should resolve to feature/SPEC-AUTH-001
			err := cmd.RunE(cmd, []string{"SPEC-AUTH-001"})
			if err != nil {
				t.Fatalf("runSwitch error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "/repo-auth") {
				t.Errorf("output should contain path, got %q", output)
			}
			if !strings.Contains(output, "feature/SPEC-AUTH-001") {
				t.Errorf("output should contain resolved branch name, got %q", output)
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

// --- Tests for SPEC-ID resolution ---

func TestResolveSpecBranch(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"regular branch", "feature-x", "feature-x"},
		{"spec id", "SPEC-AUTH-001", "feature/SPEC-AUTH-001"},
		{"spec id with long name", "SPEC-UI-042", "feature/SPEC-UI-042"},
		{"not spec just prefix", "SPEC-", "SPEC-"},
		{"not spec one part", "SPEC-AUTH", "SPEC-AUTH"},
		{"not spec trailing dash", "SPEC-X-", "SPEC-X-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSpecBranch(tt.input)
			if got != tt.want {
				t.Errorf("resolveSpecBranch(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSpecID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid spec", "SPEC-AUTH-001", true},
		{"valid spec 2", "SPEC-UI-042", true},
		{"not spec no prefix", "feature-x", false},
		{"not spec too few parts", "SPEC-AUTH", false},
		{"not spec just prefix", "SPEC-", false},
		{"not spec trailing dash", "SPEC-X-", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSpecID(tt.input)
			if got != tt.want {
				t.Errorf("isSpecID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunNew_SpecID(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	var capturedPath, capturedBranch string
	WorktreeProvider = &mockWorktreeManager{
		addFunc: func(path, branch string) error {
			capturedPath = path
			capturedBranch = branch
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Reset --path flag to ensure default path is used.
			_ = cmd.Flags().Set("path", "")

			err := cmd.RunE(cmd, []string{"SPEC-AUTH-001"})
			if err != nil {
				t.Fatalf("runNew error: %v", err)
			}

			if capturedBranch != "feature/SPEC-AUTH-001" {
				t.Errorf("branch = %q, want %q", capturedBranch, "feature/SPEC-AUTH-001")
			}
			wantPath := filepath.Join(".moai", "worktrees", "SPEC-AUTH-001")
			if capturedPath != wantPath {
				t.Errorf("path = %q, want %q", capturedPath, wantPath)
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

func TestRunNew_DefaultPath(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	tests := []struct {
		name     string
		input    string
		wantPath string
	}{
		{
			name:     "SPEC ID uses .moai/worktrees",
			input:    "SPEC-AUTH-001",
			wantPath: filepath.Join(".moai", "worktrees", "SPEC-AUTH-001"),
		},
		{
			name:     "regular branch uses sibling directory",
			input:    "feature-x",
			wantPath: filepath.Join("..", "feature-x"),
		},
		{
			name:     "SPEC with different category",
			input:    "SPEC-UI-042",
			wantPath: filepath.Join(".moai", "worktrees", "SPEC-UI-042"),
		},
		{
			name:     "branch with slash passes through",
			input:    "fix/something",
			wantPath: filepath.Join("..", "fix/something"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPath string
			WorktreeProvider = &mockWorktreeManager{
				addFunc: func(path, _ string) error {
					capturedPath = path
					return nil
				},
			}

			for _, cmd := range WorktreeCmd.Commands() {
				if cmd.Name() == "new" {
					buf := new(bytes.Buffer)
					cmd.SetOut(buf)
					cmd.SetErr(buf)

					// Reset --path flag to empty to test default behavior.
					_ = cmd.Flags().Set("path", "")

					err := cmd.RunE(cmd, []string{tt.input})
					if err != nil {
						t.Fatalf("runNew error: %v", err)
					}

					if capturedPath != tt.wantPath {
						t.Errorf("path = %q, want %q", capturedPath, tt.wantPath)
					}
					return
				}
			}
			t.Error("new subcommand not found")
		})
	}
}

// --- Tests for new --base flag ---

func TestWorktreeCmd_NewHasBaseFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			f := cmd.Flags().Lookup("base")
			if f == nil {
				t.Error("worktree new should have --base flag")
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

// --- Tests for enhanced sync ---

func TestRunSync_WithBranch(t *testing.T) {
	origProvider := WorktreeProvider
	origSync := mockSyncFunc
	defer func() {
		WorktreeProvider = origProvider
		mockSyncFunc = origSync
	}()

	var syncPath, syncBase, syncStrategy string
	mockSyncFunc = func(wtPath, baseBranch, strategy string) error {
		syncPath = wtPath
		syncBase = baseBranch
		syncStrategy = strategy
		return nil
	}

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feat", Branch: "feat", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.RunE(cmd, []string{"feat"})
			if err != nil {
				t.Fatalf("runSync error: %v", err)
			}

			if syncPath != "/repo-feat" {
				t.Errorf("sync path = %q, want %q", syncPath, "/repo-feat")
			}
			if syncBase != "main" {
				t.Errorf("sync base = %q, want %q", syncBase, "main")
			}
			if syncStrategy != "merge" {
				t.Errorf("sync strategy = %q, want %q", syncStrategy, "merge")
			}

			output := buf.String()
			if !strings.Contains(output, "Syncing worktree") {
				t.Errorf("output should contain 'Syncing worktree', got %q", output)
			}
			if !strings.Contains(output, "Sync complete") {
				t.Errorf("output should contain 'Sync complete', got %q", output)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestRunSync_BranchNotFound(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			err := cmd.RunE(cmd, []string{"nonexistent"})
			if err == nil {
				t.Error("runSync should error for unknown branch")
			}
			if !strings.Contains(err.Error(), "no worktree found") {
				t.Errorf("error should mention no worktree found, got %v", err)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestRunSync_SyncError(t *testing.T) {
	origProvider := WorktreeProvider
	origSync := mockSyncFunc
	defer func() {
		WorktreeProvider = origProvider
		mockSyncFunc = origSync
	}()

	mockSyncFunc = func(_, _, _ string) error {
		return errors.New("sync failed")
	}

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo-feat", Branch: "feat", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			err := cmd.RunE(cmd, []string{"feat"})
			if err == nil {
				t.Error("runSync should error when Sync fails")
			}
			if !strings.Contains(err.Error(), "sync worktree") {
				t.Errorf("error should mention sync worktree, got %v", err)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestWorktreeCmd_SyncHasBaseFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			f := cmd.Flags().Lookup("base")
			if f == nil {
				t.Error("worktree sync should have --base flag")
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestWorktreeCmd_SyncHasStrategyFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			f := cmd.Flags().Lookup("strategy")
			if f == nil {
				t.Error("worktree sync should have --strategy flag")
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

// --- Tests for list --verbose ---

func TestRunList_Verbose(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc12345def67890"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "789abcde00012345"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("verbose", "true"); err != nil {
				t.Fatalf("set verbose flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("verbose", "false") }()

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runList error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Branch:") {
				t.Errorf("verbose output should contain 'Branch:', got %q", output)
			}
			if !strings.Contains(output, "Path:") {
				t.Errorf("verbose output should contain 'Path:', got %q", output)
			}
			if !strings.Contains(output, "HEAD:") {
				t.Errorf("verbose output should contain 'HEAD:', got %q", output)
			}
			// Verbose shows full HEAD hash, not truncated.
			if !strings.Contains(output, "abc12345def67890") {
				t.Errorf("verbose output should contain full HEAD hash, got %q", output)
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestWorktreeCmd_ListHasVerboseFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			f := cmd.Flags().Lookup("verbose")
			if f == nil {
				t.Error("worktree list should have --verbose flag")
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

// --- Tests for clean --merged-only ---

func TestRunClean_MergedOnly(t *testing.T) {
	origProvider := WorktreeProvider
	origMerged := mockIsBranchMergedFunc
	defer func() {
		WorktreeProvider = origProvider
		mockIsBranchMergedFunc = origMerged
	}()

	mockIsBranchMergedFunc = func(branch, base string) (bool, error) {
		return branch == "feature", nil
	}

	var removedPath string
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "def"},
			}, nil
		},
		removeFunc: func(path string, force bool) error {
			removedPath = path
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("merged-only", "true"); err != nil {
				t.Fatalf("set merged-only flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("merged-only", "false") }()

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runClean error: %v", err)
			}

			if removedPath != "/repo-feature" {
				t.Errorf("removedPath = %q, want %q", removedPath, "/repo-feature")
			}

			output := buf.String()
			if !strings.Contains(output, "Removing merged worktree") {
				t.Errorf("output should contain 'Removing merged worktree', got %q", output)
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

func TestRunClean_MergedOnlyNone(t *testing.T) {
	origProvider := WorktreeProvider
	origMerged := mockIsBranchMergedFunc
	defer func() {
		WorktreeProvider = origProvider
		mockIsBranchMergedFunc = origMerged
	}()

	mockIsBranchMergedFunc = func(_, _ string) (bool, error) {
		return false, nil
	}

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("merged-only", "true"); err != nil {
				t.Fatalf("set merged-only flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("merged-only", "false") }()

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runClean error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "No merged worktrees to clean") {
				t.Errorf("output should contain 'No merged worktrees to clean', got %q", output)
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

func TestWorktreeCmd_CleanHasMergedOnlyFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			f := cmd.Flags().Lookup("merged-only")
			if f == nil {
				t.Error("worktree clean should have --merged-only flag")
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

// --- Tests for enhanced done --delete-branch ---

func TestRunDone_WithDeleteBranch(t *testing.T) {
	origProvider := WorktreeProvider
	origDeleteBranch := mockDeleteBranchFunc
	defer func() {
		WorktreeProvider = origProvider
		mockDeleteBranchFunc = origDeleteBranch
	}()

	var deletedBranch string
	mockDeleteBranchFunc = func(name string) error {
		deletedBranch = name
		return nil
	}

	removeCalled := false
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "def"},
			}, nil
		},
		removeFunc: func(_ string, _ bool) error {
			removeCalled = true
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("delete-branch", "true"); err != nil {
				t.Fatalf("set delete-branch flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("delete-branch", "false") }()

			err := cmd.RunE(cmd, []string{"feature"})
			if err != nil {
				t.Fatalf("runDone error: %v", err)
			}

			if !removeCalled {
				t.Error("Remove should have been called")
			}
			if deletedBranch != "feature" {
				t.Errorf("deletedBranch = %q, want %q", deletedBranch, "feature")
			}

			output := buf.String()
			if !strings.Contains(output, "Branch feature deleted") {
				t.Errorf("output should contain 'Branch feature deleted', got %q", output)
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

func TestRunDone_DeleteBranchError(t *testing.T) {
	origProvider := WorktreeProvider
	origDeleteBranch := mockDeleteBranchFunc
	defer func() {
		WorktreeProvider = origProvider
		mockDeleteBranchFunc = origDeleteBranch
	}()

	mockDeleteBranchFunc = func(_ string) error {
		return errors.New("branch in use")
	}

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-feature", Branch: "feature", HEAD: "def"},
			}, nil
		},
		removeFunc: func(_ string, _ bool) error {
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("delete-branch", "true"); err != nil {
				t.Fatalf("set delete-branch flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("delete-branch", "false") }()

			err := cmd.RunE(cmd, []string{"feature"})
			// Should NOT error (graceful degradation).
			if err != nil {
				t.Fatalf("runDone should not error on branch delete failure, got: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "Warning: could not delete branch") {
				t.Errorf("output should contain warning about branch deletion, got %q", output)
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

// --- Tests for status --all ---

func TestRunStatus_ShowAll(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc12345def67890"},
			}, nil
		},
		rootPath: "/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			if err := cmd.Flags().Set("all", "true"); err != nil {
				t.Fatalf("set all flag: %v", err)
			}
			defer func() { _ = cmd.Flags().Set("all", "false") }()

			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("runStatus error: %v", err)
			}

			output := buf.String()
			// With --all, full hash should be shown.
			if !strings.Contains(output, "abc12345def67890") {
				t.Errorf("output should contain full hash with --all, got %q", output)
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

func TestWorktreeCmd_StatusHasAllFlag(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "status" {
			f := cmd.Flags().Lookup("all")
			if f == nil {
				t.Error("worktree status should have --all flag")
			}
			return
		}
	}
	t.Error("status subcommand not found")
}

// --- Tests for done with SPEC-ID ---

func TestRunDone_SpecID(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	removeCalled := false
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-auth", Branch: "feature/SPEC-AUTH-001", HEAD: "def"},
			}, nil
		},
		removeFunc: func(path string, force bool) error {
			removeCalled = true
			if path != "/repo-auth" {
				t.Errorf("path = %q, want %q", path, "/repo-auth")
			}
			return nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "done" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Reset flags.
			_ = cmd.Flags().Set("force", "false")
			_ = cmd.Flags().Set("delete-branch", "false")

			// Pass raw SPEC ID; should resolve to feature/SPEC-AUTH-001
			err := cmd.RunE(cmd, []string{"SPEC-AUTH-001"})
			if err != nil {
				t.Fatalf("runDone error: %v", err)
			}

			if !removeCalled {
				t.Error("Remove should have been called")
			}
			if !strings.Contains(buf.String(), "Done") {
				t.Errorf("output should contain 'Done', got %q", buf.String())
			}
			if !strings.Contains(buf.String(), "feature/SPEC-AUTH-001") {
				t.Errorf("output should contain resolved branch name, got %q", buf.String())
			}
			return
		}
	}
	t.Error("done subcommand not found")
}

// --- Tests for sync with SPEC-ID ---

func TestRunSync_SpecID(t *testing.T) {
	origProvider := WorktreeProvider
	origSync := mockSyncFunc
	defer func() {
		WorktreeProvider = origProvider
		mockSyncFunc = origSync
	}()

	var syncPath string
	mockSyncFunc = func(wtPath, _, _ string) error {
		syncPath = wtPath
		return nil
	}

	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{
				{Path: "/repo", Branch: "main", HEAD: "abc"},
				{Path: "/repo-auth", Branch: "feature/SPEC-AUTH-001", HEAD: "def"},
			}, nil
		},
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Pass raw SPEC ID; should resolve to feature/SPEC-AUTH-001
			err := cmd.RunE(cmd, []string{"SPEC-AUTH-001"})
			if err != nil {
				t.Fatalf("runSync error: %v", err)
			}

			if syncPath != "/repo-auth" {
				t.Errorf("sync path = %q, want %q", syncPath, "/repo-auth")
			}

			output := buf.String()
			if !strings.Contains(output, "Sync complete") {
				t.Errorf("output should contain 'Sync complete', got %q", output)
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

// --- Tests for config with 2 args ---

func TestRunConfig_SetNotSupported(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = &mockWorktreeManager{
		rootPath: "/test/repo",
	}

	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "config" {
			err := cmd.RunE(cmd, []string{"root", "some-value"})
			if err == nil {
				t.Error("runConfig should error for set operation")
			}
			if !strings.Contains(err.Error(), "config set is not yet supported") {
				t.Errorf("error should mention config set not supported, got %v", err)
			}
			return
		}
	}
	t.Error("config subcommand not found")
}
