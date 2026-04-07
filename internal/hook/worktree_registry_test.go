package hook

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterWorktree(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		branch    string
		agentName string
	}{
		{
			name:      "register single entry",
			path:      "/worktrees/abc123",
			branch:    "feat/auth",
			agentName: "expert-backend",
		},
		{
			name:      "register with empty agent name",
			path:      "/worktrees/def456",
			branch:    "main",
			agentName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectDir := t.TempDir()
			registerWorktree(projectDir, tt.path, tt.branch, tt.agentName)

			stateFile := worktreeStateFile(projectDir)
			entries := loadWorktreeEntries(stateFile)

			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}
			if entries[0].Path != tt.path {
				t.Errorf("Path = %v, want %v", entries[0].Path, tt.path)
			}
			if entries[0].Branch != tt.branch {
				t.Errorf("Branch = %v, want %v", entries[0].Branch, tt.branch)
			}
			if entries[0].AgentName != tt.agentName {
				t.Errorf("AgentName = %v, want %v", entries[0].AgentName, tt.agentName)
			}
			if entries[0].CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
		})
	}
}

func TestRegisterMultipleWorktrees(t *testing.T) {
	projectDir := t.TempDir()

	registerWorktree(projectDir, "/wt/a", "branch-a", "agent-a")
	registerWorktree(projectDir, "/wt/b", "branch-b", "agent-b")
	registerWorktree(projectDir, "/wt/c", "branch-c", "agent-c")

	entries := loadWorktreeEntries(worktreeStateFile(projectDir))
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestUnregisterWorktree(t *testing.T) {
	projectDir := t.TempDir()

	registerWorktree(projectDir, "/wt/keep", "main", "agent-1")
	registerWorktree(projectDir, "/wt/remove", "feat", "agent-2")
	registerWorktree(projectDir, "/wt/also-keep", "develop", "agent-3")

	unregisterWorktree(projectDir, "/wt/remove")

	entries := loadWorktreeEntries(worktreeStateFile(projectDir))
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after unregister, got %d", len(entries))
	}

	for _, e := range entries {
		if e.Path == "/wt/remove" {
			t.Error("removed entry should not be present")
		}
	}
}

func TestUnregisterNonExistentWorktree(t *testing.T) {
	projectDir := t.TempDir()

	registerWorktree(projectDir, "/wt/existing", "main", "agent-1")

	// Unregistering a path that does not exist should be a no-op.
	unregisterWorktree(projectDir, "/wt/nonexistent")

	entries := loadWorktreeEntries(worktreeStateFile(projectDir))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestLoadWorktreeEntries_MissingFile(t *testing.T) {
	projectDir := t.TempDir()
	stateFile := worktreeStateFile(projectDir)

	entries := loadWorktreeEntries(stateFile)
	if len(entries) != 0 {
		t.Errorf("expected empty slice for missing file, got %d entries", len(entries))
	}
}

func TestLoadWorktreeEntries_CorruptFile(t *testing.T) {
	projectDir := t.TempDir()
	stateFile := worktreeStateFile(projectDir)

	// Create the directory and write corrupt JSON.
	if err := os.MkdirAll(filepath.Dir(stateFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stateFile, []byte("not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := loadWorktreeEntries(stateFile)
	if len(entries) != 0 {
		t.Errorf("expected empty slice for corrupt file, got %d entries", len(entries))
	}
}

func TestWorktreeStateFile(t *testing.T) {
	got := worktreeStateFile("/project")
	want := "/project/.moai/state/worktrees.json"
	if got != want {
		t.Errorf("worktreeStateFile() = %v, want %v", got, want)
	}
}
