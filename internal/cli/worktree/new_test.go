package worktree

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withWorktreeCreator(t *testing.T, fn func(string, io.Writer) (string, error)) {
	t.Helper()
	old := WorktreeCreator
	WorktreeCreator = func(name string, out io.Writer) (string, error) {
		return fn(name, out)
	}
	t.Cleanup(func() { WorktreeCreator = old })
}

func TestNew_CreatesThroughInjectedCurrentPlumbing(t *testing.T) {
	var gotName string
	withWorktreeCreator(t, func(name string, _ io.Writer) (string, error) {
		gotName = name
		return "/repo/.claude/worktrees/WT-card", nil
	})

	cmd := newNewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"WT-card"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute new: %v", err)
	}
	if gotName != "WT-card" {
		t.Fatalf("creator name = %q, want WT-card", gotName)
	}
	if got := out.String(); !strings.Contains(got, "WT-card") || !strings.Contains(got, "/repo/.claude/worktrees/WT-card") {
		t.Fatalf("stdout = %q, want branch and path", got)
	}
}

func TestNew_RejectsMissingNameBeforeCreation(t *testing.T) {
	called := false
	withWorktreeCreator(t, func(string, io.Writer) (string, error) {
		called = true
		return "", nil
	})
	cmd := newNewCmd()
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("missing name accepted")
	}
	if !strings.Contains(err.Error(), "<name>") {
		t.Fatalf("missing-name error = %q, want expected argument <name>", err)
	}
	if called {
		t.Fatal("creator called for missing name")
	}
}

func TestNew_RejectsPathEscapeBeforeCreation(t *testing.T) {
	for _, name := range []string{"..", "../escape", "nested/name", `nested\\name`, "/absolute", "WT-..-escape", "."} {
		t.Run(strings.ReplaceAll(name, "/", "_"), func(t *testing.T) {
			called := false
			withWorktreeCreator(t, func(string, io.Writer) (string, error) {
				called = true
				return "", nil
			})
			cmd := newNewCmd()
			cmd.SetArgs([]string{name})
			if err := cmd.Execute(); err == nil {
				t.Fatalf("unsafe name %q accepted", name)
			}
			if called {
				t.Fatalf("creator called for unsafe name %q", name)
			}
		})
	}
}

func TestNew_ReportsCollisionWithoutRetry(t *testing.T) {
	calls := 0
	withWorktreeCreator(t, func(string, io.Writer) (string, error) {
		calls++
		return "", errors.New("destination already exists")
	})
	cmd := newNewCmd()
	cmd.SetArgs([]string{"WT-existing"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "WT-existing") || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("collision error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("creator calls = %d, want 1", calls)
	}
}

func TestNew_HasNoRetiredFlags(t *testing.T) {
	cmd := newNewCmd()
	for _, name := range []string{"base", "from-current", "path", "tmux", "team"} {
		if cmd.Flags().Lookup(name) != nil {
			t.Fatalf("retired flag --%s was revived", name)
		}
	}
}

func TestNew_NoAskUserQuestion(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("new.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(src), "AskUserQuestion") || strings.Contains(string(src), "mcp__askuser__") {
		t.Fatal("worktree new must remain non-interactive")
	}
}
