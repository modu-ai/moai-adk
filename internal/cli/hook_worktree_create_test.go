// End-to-end wire-format regression test for the WorktreeCreate contract
// (issue #1570).
//
// Claude Code's WorktreeCreate hook is an active creator: the payload carries
// the suggested slug in the official `name` field, and the hook must create
// the worktree itself and echo its absolute path to stdout as plain text —
// empty stdout with exit 0 ABORTS the agent spawn. The passthrough
// implementation that preceded this test always printed nothing (it waited
// for a `worktree_path` input field this event never sends), so every
// isolation: worktree spawn died. This test pins the full dispatcher path:
// stdin JSON → handler → created worktree → path on stdout.
package cli

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// TestHookWorktreeCreate_EchoesCreatedPath runs the real `moai hook
// worktree-create` subcommand wiring against a throwaway git repository and
// asserts the contract: exit 0 with exactly one line on stdout — the absolute
// path of a directory that now exists.
func TestHookWorktreeCreate_EchoesCreatedPath(t *testing.T) {
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve symlinks: %v", err)
	}
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, repo, "add", ".")
	runTestGit(t, repo, "commit", "-m", "Initial commit")

	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	InitDependencies()

	payload := `{"hook_event_name":"WorktreeCreate","cwd":` + jsonQuote(repo) + `,"session_id":"sess-e2e","name":"e2e-probe"}`
	swapStdinString(t, payload)

	origStdout := os.Stdout
	rOut, wOut, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("os.Pipe: %v", pipeErr)
	}
	os.Stdout = wOut

	var runErr error
	found := false
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "worktree-create" {
			found = true
			cmd.SetContext(context.Background())
			runErr = cmd.RunE(cmd, []string{})
			break
		}
	}
	_ = wOut.Close()
	os.Stdout = origStdout

	if !found {
		t.Fatal("worktree-create subcommand not found")
	}
	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}

	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	wantPath := filepath.Join(repo, ".claude", "worktrees", "e2e-probe")
	if got := strings.TrimSpace(string(out)); got != wantPath {
		t.Errorf("stdout = %q, want %q (empty stdout aborts the agent spawn)", got, wantPath)
	}
	if info, statErr := os.Stat(wantPath); statErr != nil {
		t.Errorf("worktree not created at %s: %v", wantPath, statErr)
	} else if !info.IsDir() {
		t.Errorf("%s is not a directory", wantPath)
	}
}

// TestWriteHookOutput_WorktreeCreatePrefersOutputPath pins the dispatcher's
// echo rule directly: WorktreeCreate echoes the path the handler created
// (output.WorktreePath) and falls back to input.WorktreePath when the
// handler set none; WorktreeRemove keeps echoing the input path.
//
// Not parallel: swaps the process-global os.Stdout.
func TestWriteHookOutput_WorktreeCreatePrefersOutputPath(t *testing.T) {
	tests := []struct {
		name       string
		event      string
		outputPath string
		inputPath  string
		wantStdout string
	}{
		{
			name:       "create prefers handler-created path",
			event:      "WorktreeCreate",
			outputPath: "/repo/.claude/worktrees/agent-a",
			inputPath:  "/ignored",
			wantStdout: "/repo/.claude/worktrees/agent-a",
		},
		{
			name:       "create falls back to input path",
			event:      "WorktreeCreate",
			outputPath: "",
			inputPath:  "/repo/.claude/worktrees/agent-b",
			wantStdout: "/repo/.claude/worktrees/agent-b",
		},
		{
			name:       "remove echoes input path",
			event:      "WorktreeRemove",
			outputPath: "/ignored",
			inputPath:  "/repo/.claude/worktrees/agent-c",
			wantStdout: "/repo/.claude/worktrees/agent-c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureWriteHookOutput(t, hook.EventType(tt.event),
				&hook.HookInput{WorktreePath: tt.inputPath},
				&hook.HookOutput{WorktreePath: tt.outputPath})
			if out != tt.wantStdout {
				t.Errorf("stdout = %q, want %q", out, tt.wantStdout)
			}
		})
	}
}

// captureWriteHookOutput runs writeHookOutput with os.Stdout swapped for a
// pipe and returns the captured text (trimmed). The worktree events never
// reach the JSON protocol branch, so the global deps graph is not needed.
func captureWriteHookOutput(t *testing.T, event hook.EventType, input *hook.HookInput, output *hook.HookOutput) string {
	t.Helper()

	origStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = wOut
	defer func() { os.Stdout = origStdout }()

	if err := writeHookOutput(event, input, output); err != nil {
		t.Fatalf("writeHookOutput: %v", err)
	}

	_ = wOut.Close()
	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// runTestGit executes git in dir and fails the test on error.
func runTestGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %s: %v", args, dir, string(out), err)
	}
}

// jsonQuote renders s as a JSON string literal (the payloads built in this
// file only need the escaping of backslashes and quotes).
func jsonQuote(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
