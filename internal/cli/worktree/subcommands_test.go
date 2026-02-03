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
	removeFunc func(path string) error
	pruneFunc  func() error
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

func (m *mockWorktreeManager) Remove(path string) error {
	if m.removeFunc != nil {
		return m.removeFunc(path)
	}
	return nil
}

func (m *mockWorktreeManager) Prune() error {
	if m.pruneFunc != nil {
		return m.pruneFunc()
	}
	return nil
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
	WorktreeProvider = &mockWorktreeManager{
		removeFunc: func(path string) error {
			capturedPath = path
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
		removeFunc: func(_ string) error {
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
