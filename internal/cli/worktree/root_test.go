package worktree

import (
	"slices"
	"testing"
)

func TestWorktreeCmd_Exists(t *testing.T) {
	if WorktreeCmd == nil {
		t.Fatal("WorktreeCmd should not be nil")
	}
}

func TestWorktreeCmd_Use(t *testing.T) {
	if WorktreeCmd.Use != "worktree" {
		t.Errorf("WorktreeCmd.Use = %q, want %q", WorktreeCmd.Use, "worktree")
	}
}

func TestWorktreeCmd_Alias(t *testing.T) {
	if len(WorktreeCmd.Aliases) == 0 {
		t.Fatal("WorktreeCmd should have aliases")
	}
	found := slices.Contains(WorktreeCmd.Aliases, "wt")
	if !found {
		t.Error("WorktreeCmd should have 'wt' alias")
	}
}

func TestWorktreeCmd_Short(t *testing.T) {
	if WorktreeCmd.Short == "" {
		t.Error("WorktreeCmd.Short should not be empty")
	}
}

func TestWorktreeCmd_HasSubcommands(t *testing.T) {
	expected := []string{"new", "list", "switch", "go", "sync", "remove", "clean", "recover", "done", "config", "status"}
	for _, name := range expected {
		found := false
		for _, cmd := range WorktreeCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("worktree should have %q subcommand", name)
		}
	}
}

func TestWorktreeCmd_SubcommandCount(t *testing.T) {
	count := len(WorktreeCmd.Commands())
	if count != 11 {
		t.Errorf("worktree should have 11 subcommands, got %d", count)
	}
}

func TestWorktreeCmd_SubcommandsHaveShortDesc(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Short == "" {
			t.Errorf("worktree subcommand %q should have a short description", cmd.Name())
		}
	}
}

func TestWorktreeCmd_NewRequiresArg(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			// "new" command requires exactly 1 arg
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Error("worktree new should require an argument")
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

func TestWorktreeCmd_RemoveRequiresArg(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "remove" {
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Error("worktree remove should require an argument")
			}
			return
		}
	}
	t.Error("remove subcommand not found")
}

func TestWorktreeCmd_ListNoArgs(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "list" {
			// "list" has no WorktreeProvider, should error
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("worktree list should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("list subcommand not found")
}

func TestWorktreeCmd_CleanNoProvider(t *testing.T) {
	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "clean" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("worktree clean should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("clean subcommand not found")
}

func TestWorktreeCmd_SyncNoProvider(t *testing.T) {
	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "sync" {
			err := cmd.RunE(cmd, []string{})
			if err == nil {
				t.Error("worktree sync should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("sync subcommand not found")
}

func TestWorktreeCmd_SwitchNoProvider(t *testing.T) {
	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "switch" {
			err := cmd.RunE(cmd, []string{"test-branch"})
			if err == nil {
				t.Error("worktree switch should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("switch subcommand not found")
}

func TestWorktreeCmd_NewNoProvider(t *testing.T) {
	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "new" {
			err := cmd.RunE(cmd, []string{"test-branch"})
			if err == nil {
				t.Error("worktree new should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("new subcommand not found")
}

func TestWorktreeCmd_RemoveNoProvider(t *testing.T) {
	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "remove" {
			err := cmd.RunE(cmd, []string{"/tmp/test"})
			if err == nil {
				t.Error("worktree remove should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("remove subcommand not found")
}

func TestWorktreeCmd_GoRequiresArg(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			err := cmd.Args(cmd, []string{})
			if err == nil {
				t.Error("worktree go should require an argument")
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

func TestWorktreeCmd_GoNoProvider(t *testing.T) {
	origProvider := WorktreeProvider
	defer func() { WorktreeProvider = origProvider }()

	WorktreeProvider = nil
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Name() == "go" {
			err := cmd.RunE(cmd, []string{"test-branch"})
			if err == nil {
				t.Error("worktree go should error without WorktreeProvider")
			}
			return
		}
	}
	t.Error("go subcommand not found")
}

func TestMinLen(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{0, 0, 0},
		{8, 8, 8},
	}
	for _, tt := range tests {
		got := minLen(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("minLen(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
