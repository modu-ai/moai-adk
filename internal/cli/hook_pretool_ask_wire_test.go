// End-to-end wire-format regression test for the PreToolUse "ask" verdict.
//
// "deny" already had CLI-level coverage (hook_protocol_fix_test.go); "ask" had
// none, which is why the registry merge could drop it and the suite stayed
// green for the whole life of the defect. This test pins "ask" the same way
// "deny" is pinned: real protocol, real registry, real handler, assertion on
// the JSON that actually reaches stdout.
package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runPreToolHookCLI feeds payload to the `moai hook pre-tool` subcommand with
// the real wiring and returns the JSON object written to stdout.
func runPreToolHookCLI(t *testing.T, payload string) map[string]any {
	t.Helper()

	origDeps := deps
	t.Cleanup(func() { deps = origDeps })
	InitDependencies()

	swapStdinString(t, payload)

	origStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = wOut

	var runErr error
	found := false
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "pre-tool" {
			found = true
			cmd.SetContext(context.Background())
			runErr = cmd.RunE(cmd, []string{})
			break
		}
	}
	_ = wOut.Close()
	os.Stdout = origStdout

	if !found {
		t.Fatal("pre-tool subcommand not found")
	}
	if runErr != nil {
		t.Fatalf("RunE returned error: %v", runErr)
	}

	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}

	var decoded map[string]any
	if err := json.NewDecoder(strings.NewReader(strings.TrimSpace(string(out)))).Decode(&decoded); err != nil {
		t.Fatalf("decode stdout %q: %v", string(out), err)
	}
	return decoded
}

// hookSpecific pulls hookSpecificOutput out of a decoded hook response.
func hookSpecific(t *testing.T, decoded map[string]any) map[string]any {
	t.Helper()

	hso, ok := decoded["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput missing or not an object in %v", decoded)
	}
	return hso
}

// TestHookPreToolCLI_AskReachesTheWire is the end-to-end counterpart of the
// registry-level tests: an AskPatterns-matching file edit must emit
// permissionDecision "ask" with a reason on stdout. Before the merge fix this
// emitted "allow", so every one of the 27 file patterns was silently
// unguarded.
func TestHookPreToolCLI_AskReachesTheWire(t *testing.T) {
	// Not parallel: swaps global deps, os.Stdin, os.Stdout, and cwd.
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Chdir(dir)

	payload, err := json.Marshal(map[string]any{
		"hook_event_name": "PreToolUse",
		"session_id":      "wire-ask-file",
		"cwd":             dir,
		"tool_name":       "Edit",
		"tool_input": map[string]string{
			"file_path":  filepath.Join(dir, "package.json"),
			"old_string": "a",
			"new_string": "b",
		},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	hso := hookSpecific(t, runPreToolHookCLI(t, string(payload)))

	if got := hso["permissionDecision"]; got != "ask" {
		t.Errorf("permissionDecision = %v, want \"ask\" (the ask verdict never reached the wire)", got)
	}
	if reason, _ := hso["permissionDecisionReason"].(string); reason == "" {
		t.Error("permissionDecisionReason is empty on the wire; the dialog would explain nothing")
	}
}

// TestHookPreToolCLI_AskBashReachesTheWire pins the Bash half of the same
// contract: `git reset --hard` is an AskBashPatterns entry and was equally
// downgraded to "allow" on the wire.
func TestHookPreToolCLI_AskBashReachesTheWire(t *testing.T) {
	// Not parallel: swaps global deps, os.Stdin, os.Stdout, and cwd.
	dir := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Chdir(dir)

	payload, err := json.Marshal(map[string]any{
		"hook_event_name": "PreToolUse",
		"session_id":      "wire-ask-bash",
		"cwd":             dir,
		"tool_name":       "Bash",
		"tool_input":      map[string]string{"command": "git reset --hard HEAD~1"},
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	hso := hookSpecific(t, runPreToolHookCLI(t, string(payload)))

	if got := hso["permissionDecision"]; got != "ask" {
		t.Errorf("permissionDecision = %v, want \"ask\" for `git reset --hard`", got)
	}
	if reason, _ := hso["permissionDecisionReason"].(string); reason == "" {
		t.Error("permissionDecisionReason is empty on the wire for the ask-bash path")
	}
}
