package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// todo_flag_independence_test.go pins REQ-3 and the REQ-2 D2 clause: what
// workflow.todo.enabled turns off is the GUIDANCE, never the feature. The
// command stays registered and every verb keeps working, so an operator who
// still sees `/moai todo` in the skill listing (which this SPEC deliberately
// does not suppress — spec.md §E.3) gets a working command when they call it
// by name rather than a silent no-op.

// todoDisabledFixture prepares a project whose workflow.yaml turns the todo
// guidance off, and isolates BOTH the project root and the home directory.
//
// t.TempDir() alone is NOT sufficient isolation here. resolveTodoQueueRoot
// falls back to ~/.moai/todo/<project-key>/ for any launch context git cannot
// resolve to a primary checkout, so a test that only pins the project root
// writes into the developer's real home. userHomeDirFn is the package's
// existing injection seam for exactly this (precedent:
// todo_queue_root_test.go TestResolveTodoQueueRoot_FallbackNoGit).
func todoDisabledFixture(t *testing.T) (root, home string) {
	t.Helper()

	root = t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	initGitRepo(t, root)

	sections := filepath.Join(root, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	body := "workflow:\n    todo:\n        enabled: false\n"
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write workflow.yaml: %v", err)
	}

	home = t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	return root, home
}

// assertRealHomeUntouched checks that nothing landed under the injected home's
// fallback path — the seam is only useful if the test also proves it held.
func assertRealHomeUntouched(t *testing.T, home string) {
	t.Helper()
	fallback := filepath.Join(home, ".moai", "todo")
	if _, err := os.Stat(fallback); err == nil {
		t.Fatalf("queue landed in the home fallback %q — the project root was not resolved as a checkout", fallback)
	}
}

// TestTodoCommandRegisteredRegardlessOfFlag is AC-T-005 first case (REQ-3):
// the command is registered on rootCmd and runnable with the flag off.
//
// Hiding the command would leave the foreman skill's allowed-tools entry and
// the operator's existing queue file in place with only the entry point gone —
// a state that is harder to diagnose than a visible command, which is why
// REQ-3 makes keeping it a decision rather than an omission.
func TestTodoCommandRegisteredRegardlessOfFlag(t *testing.T) {
	_, home := todoDisabledFixture(t)

	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Name() == "todo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("`todo` is not registered on rootCmd")
	}

	if _, _, err := runTodo(t, "--help"); err != nil {
		t.Fatalf("todo --help with the flag off: %v", err)
	}

	assertRealHomeUntouched(t, home)
}

// TestTodoVerbsUnaffectedByFlag is AC-T-005 second case, and the behavioural
// judgment of the REQ-2 D2 clause. The skill's qualifying sentence ("an
// explicit /moai todo invocation still works") is prose and cannot be verified
// by searching for it; what CAN be verified is the behaviour it promises. If
// this round trip fails, that sentence is false no matter how it is worded.
func TestTodoVerbsUnaffectedByFlag(t *testing.T) {
	_, home := todoDisabledFixture(t)

	if _, _, err := runTodo(t, "add", "a card added while todo guidance is off"); err != nil {
		t.Fatalf("add with the flag off: %v", err)
	}

	out, _, err := runTodo(t)
	if err != nil {
		t.Fatalf("bare list with the flag off: %v", err)
	}
	if !strings.Contains(out, "a card added while todo guidance is off") {
		t.Fatalf("added card missing from the listing with the flag off:\n%s", out)
	}

	assertRealHomeUntouched(t, home)
}
