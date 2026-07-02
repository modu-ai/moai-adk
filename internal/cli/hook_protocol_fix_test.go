// Regression tests for the official-spec hook CLI alignment:
//
//   - G7: malformed/truncated stdin JSON must NOT produce a cobra error
//     (usage noise + exit 1); the hook emits a stderr warning, the event's
//     default output on stdout, and returns nil (exit 0).
//   - G10: readStdinLines reads os.Stdin via io.ReadAll — the previous
//     os.ReadFile("/dev/stdin") does not exist on Windows.
package cli

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/hook"
)

// swapStdinString replaces os.Stdin with a pipe carrying s for the test's
// duration (pattern shared with hook_harness_observe_test.go).
func swapStdinString(t *testing.T, s string) {
	t.Helper()
	orig := os.Stdin
	t.Cleanup(func() { os.Stdin = orig })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if _, err := w.WriteString(s); err != nil {
		t.Fatalf("write pipe: %v", err)
	}
	_ = w.Close()
	os.Stdin = r
}

// TestReadStdinLines_PipedInput is the G10 regression test: readStdinLines
// must read from os.Stdin (io.ReadAll) so it works on every platform —
// os.ReadFile("/dev/stdin") silently failed on Windows and skipped validation.
func TestReadStdinLines_PipedInput(t *testing.T) {
	swapStdinString(t, "feat: first commit\n\n  fix: second commit  \n")

	lines, err := readStdinLines()
	if err != nil {
		t.Fatalf("readStdinLines error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("readStdinLines returned %d lines, want 2: %v", len(lines), lines)
	}
	if lines[0] != "feat: first commit" || lines[1] != "fix: second commit" {
		t.Errorf("lines = %v, want trimmed commit subjects", lines)
	}
}

// TestRunHookEvent_MalformedStdinGraceful is the G7 regression test: a
// malformed stdin payload must yield exit 0 (nil error), a single valid JSON
// object on stdout, and NO dispatch — never a cobra error with usage noise.
func TestRunHookEvent_MalformedStdinGraceful(t *testing.T) {
	// Not parallel: swaps global deps + os.Stdin/os.Stdout.
	origDeps := deps
	defer func() { deps = origDeps }()

	spy := &spyRegistry{}
	deps = &Dependencies{
		HookRegistry: spy,
		HookProtocol: hook.NewProtocol(), // REAL protocol so parse failure occurs
	}

	swapStdinString(t, `{"broken`)

	// Capture stdout.
	origStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = wOut
	defer func() { os.Stdout = origStdout }()

	// Locate the pre-tool subcommand and execute.
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
		t.Fatalf("RunE returned error %v, want nil (graceful default output + exit 0)", runErr)
	}
	if len(spy.dispatched) != 0 {
		t.Errorf("dispatch called %d times on malformed input, want 0", len(spy.dispatched))
	}

	out, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	trimmed := strings.TrimSpace(string(out))
	if !json.Valid([]byte(trimmed)) {
		t.Fatalf("stdout is not a single valid JSON object: %q", trimmed)
	}
	// Exactly one JSON value (no {JSON}{JSON} double-emit).
	dec := json.NewDecoder(strings.NewReader(trimmed))
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if dec.More() {
		t.Errorf("stdout carries more than one JSON value: %q", trimmed)
	}
}
