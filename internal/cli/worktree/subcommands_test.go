package worktree

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk-go/internal/core/git"
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
