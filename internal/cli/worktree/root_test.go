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
	expected := []string{
		"sync", "remove", "clean", "recover", "done",
		"snapshot", "verify", "restore", // worktree state guard
	}
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

// TestWorktreeCmd_RetiredSubcommands pins the retirement of the creation,
// navigation, and inspection subcommands. Creating and entering a worktree is
// `moai cc -w <name>`; git's own `worktree list` covers inspection. A
// resurrected subcommand here would put a second, diverging entry path back in
// front of users.
func TestWorktreeCmd_RetiredSubcommands(t *testing.T) {
	retired := []string{"new", "list", "switch", "go", "config", "status"}
	for _, name := range retired {
		for _, cmd := range WorktreeCmd.Commands() {
			if cmd.Name() == name {
				t.Errorf("worktree subcommand %q is retired and must not be registered", name)
			}
		}
	}
}

func TestWorktreeCmd_SubcommandCount(t *testing.T) {
	count := len(WorktreeCmd.Commands())
	const expected = 8 // sync, remove, clean, recover, done + guard snapshot/verify/restore
	if count != expected {
		t.Errorf("worktree should have %d subcommands, got %d", expected, count)
	}
}

func TestWorktreeCmd_SubcommandsHaveShortDesc(t *testing.T) {
	for _, cmd := range WorktreeCmd.Commands() {
		if cmd.Short == "" {
			t.Errorf("worktree subcommand %q should have a short description", cmd.Name())
		}
	}
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
