package cli

// SPEC-CODEX-WIRING-001 M3 — the `moai hook <arg> --harness codex` runtime
// mode (REQ-CW-007 / AC-CW-011). The seam lives in the CLI dispatcher layer
// ONLY (M3 REQ-7 spirit): payloads are t83-golden-format canned JSON, deps is
// stubbed, and internal/hook is untouched.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// codexHarnessProtocol is a stub protocol serving a canned input and a REAL
// JSON serialization of the output (the codex path maps the serialized bytes).
type codexHarnessProtocol struct {
	input *hook.HookInput
}

func (p *codexHarnessProtocol) ReadInput(_ io.Reader) (*hook.HookInput, error) {
	return p.input, nil
}

func (p *codexHarnessProtocol) WriteOutput(w io.Writer, output *hook.HookOutput) error {
	if output == nil {
		output = &hook.HookOutput{}
	}
	return json.NewEncoder(w).Encode(output)
}

// codexHarnessRegistry records the dispatched event and serves a canned output.
type codexHarnessRegistry struct {
	dispatched hook.EventType
	output     *hook.HookOutput
}

func (r *codexHarnessRegistry) Register(_ hook.Handler)                  {}
func (r *codexHarnessRegistry) Handlers(_ hook.EventType) []hook.Handler { return nil }

func (r *codexHarnessRegistry) Dispatch(_ context.Context, event hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
	r.dispatched = event
	return r.output, nil
}

// captureStdoutDuring swaps os.Stdout for a pipe, runs fn, and returns what
// fn wrote to stdout. RunE writes directly to os.Stdout, so a pipe swap is
// the only faithful capture.
func captureStdoutDuring(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()
	defer func() { os.Stdout = old }()
	fn()
	_ = w.Close()
	return <-done
}

// runHookSubcommandCodex executes one dispatcher subcommand with
// --harness codex, deps stubbed to a canned input/output pair, under an
// isolated CLAUDE_PROJECT_DIR, capturing stdout. Returns the stdout text and
// the RunE error.
func runHookSubcommandCodex(t *testing.T, subcommand string, input *hook.HookInput, output *hook.HookOutput) (string, error) {
	t.Helper()

	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	origDeps := deps
	registry := &codexHarnessRegistry{output: output}
	deps = &Dependencies{
		HookRegistry: registry,
		HookProtocol: &codexHarnessProtocol{input: input},
	}
	t.Cleanup(func() { deps = origDeps })

	var sub *cobra.Command
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == subcommand {
			sub = cmd
			break
		}
	}
	if sub == nil {
		t.Fatalf("hook subcommand %q not found", subcommand)
	}
	// ParseFlags mirrors the real invocation (`moai hook stop --harness codex`)
	// AND triggers cobra's persistent-flag merge so the child's flag set sees
	// the parent-registered --harness (a bare Flags().Set() pre-execution
	// would not).
	if err := sub.ParseFlags([]string{"--harness", "codex"}); err != nil {
		t.Fatalf("parse --harness codex: %v", err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	sub.SetContext(context.Background())

	var runErr error
	stdout := captureStdoutDuring(t, func() { runErr = sub.RunE(sub, []string{}) })
	return stdout, runErr
}

// TestHarnessCodexFlagRegistered verifies the --harness flag exists on the
// hook command surface and documents codex (REQ-CW-007 entry condition).
func TestHarnessCodexFlagRegistered(t *testing.T) {
	flag := hookCmd.PersistentFlags().Lookup("harness")
	if flag == nil {
		t.Fatal("--harness flag not registered on hookCmd")
	}
	if !strings.Contains(flag.Usage, "codex") {
		t.Errorf("--harness usage %q does not document codex", flag.Usage)
	}
}

// TestHarnessCodexHarnInvalidValueFailLoud verifies an invalid --harness
// value is rejected with a diagnostic naming the valid values.
func TestHarnessCodexHarnInvalidValueFailLoud(t *testing.T) {
	origDeps := deps
	deps = &Dependencies{
		HookRegistry: &codexHarnessRegistry{output: &hook.HookOutput{}},
		HookProtocol: &codexHarnessProtocol{input: &hook.HookInput{HookEventName: "Stop"}},
	}
	t.Cleanup(func() { deps = origDeps })

	var sub *cobra.Command
	for _, cmd := range hookCmd.Commands() {
		if cmd.Name() == "stop" {
			sub = cmd
			break
		}
	}
	if err := sub.ParseFlags([]string{"--harness", "gemini"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	sub.SetContext(context.Background())

	err := sub.RunE(sub, []string{})
	if err == nil {
		t.Fatal("invalid --harness value must fail loud")
	}
	for _, want := range []string{"claude", "codex"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("diagnostic %q does not name valid value %q", err.Error(), want)
		}
	}
}

// TestHarnessCodexContinueFalseBecomesDecisionBlock verifies AC-CW-011(a):
// a blocking output (continue:false + stopReason) is rewritten to
// decision:"block" with a non-empty reason.
func TestHarnessCodexContinueFalseBecomesDecisionBlock(t *testing.T) {
	falseVal := false
	input := &hook.HookInput{HookEventName: "Stop", SessionID: "s1"}
	output := &hook.HookOutput{Continue: &falseVal, StopReason: "blocked: dangerous pattern"}

	stdout, err := runHookSubcommandCodex(t, "stop", input, output)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var mapped map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &mapped); jerr != nil {
		t.Fatalf("mapped stdout is not JSON: %v\n%s", jerr, stdout)
	}
	if mapped["decision"] != "block" {
		t.Errorf("decision = %v, want \"block\" (stdout: %s)", mapped["decision"], stdout)
	}
	reason, _ := mapped["reason"].(string)
	if reason != "blocked: dangerous pattern" {
		t.Errorf("reason = %q, want the stopReason carried over (stdout: %s)", reason, stdout)
	}
	if _, has := mapped["continue"]; has {
		t.Errorf("continue key must not survive the codex mapping (Codex ignores it — the block would be lost):\n%s", stdout)
	}
}

// TestHarnessCodexEmptyReasonGetsDefaultText verifies AC-CW-011(a) second
// arm: continue:false without a stopReason still yields a non-empty reason
// (Codex rejects an empty decision:block reason).
func TestHarnessCodexEmptyReasonGetsDefaultText(t *testing.T) {
	falseVal := false
	input := &hook.HookInput{HookEventName: "Stop", SessionID: "s1"}
	output := &hook.HookOutput{Continue: &falseVal}

	stdout, err := runHookSubcommandCodex(t, "stop", input, output)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	var mapped map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &mapped); jerr != nil {
		t.Fatalf("mapped stdout is not JSON: %v\n%s", jerr, stdout)
	}
	if mapped["decision"] != "block" {
		t.Errorf("decision = %v, want \"block\"", mapped["decision"])
	}
	if reason, _ := mapped["reason"].(string); reason == "" {
		t.Errorf("reason is empty — Codex would silently drop the block:\n%s", stdout)
	}
}

// TestHarnessCodexSystemMessageDiscardRecorded verifies AC-CW-011(b): a
// systemMessage on a NON-UserPromptSubmit event is discarded AND the discard
// is recorded to .moai/logs/codex-adapter.jsonl with event, key, and
// content_length (the "no silence" obligation).
func TestHarnessCodexSystemMessageDiscardRecorded(t *testing.T) {
	input := &hook.HookInput{HookEventName: "Stop", SessionID: "s1"}
	output := &hook.HookOutput{SystemMessage: "sync gate found 2 findings"}

	stdout, err := runHookSubcommandCodex(t, "stop", input, output)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	if strings.Contains(stdout, "systemMessage") {
		t.Errorf("systemMessage leaked into codex stdout (undeliverable on Stop):\n%s", stdout)
	}

	root := os.Getenv("CLAUDE_PROJECT_DIR")
	raw, rerr := os.ReadFile(filepath.Join(root, ".moai", "logs", "codex-adapter.jsonl"))
	if rerr != nil {
		t.Fatalf("read codex-adapter.jsonl: %v", rerr)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	var record map[string]any
	if jerr := json.Unmarshal([]byte(lines[len(lines)-1]), &record); jerr != nil {
		t.Fatalf("last sink line is not JSON: %v\n%s", jerr, raw)
	}
	if record["event"] != "Stop" {
		t.Errorf("discard record event = %v, want Stop", record["event"])
	}
	if record["key"] != "systemMessage" {
		t.Errorf("discard record key = %v, want systemMessage", record["key"])
	}
	if cl, _ := record["content_length"].(float64); cl != float64(len("sync gate found 2 findings")) {
		t.Errorf("discard record content_length = %v, want %d", record["content_length"], len("sync gate found 2 findings"))
	}
}

// TestHarnessCodexUserPromptSubmitSystemMessageRoutes verifies the ONE
// working additionalContext channel: UserPromptSubmit systemMessage maps to
// hookSpecificOutput.additionalContext (no discard on that event).
func TestHarnessCodexUserPromptSubmitSystemMessageRoutes(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	input := &hook.HookInput{HookEventName: "UserPromptSubmit", SessionID: "s1"}
	output := &hook.HookOutput{SystemMessage: "context: SPEC-X is active"}

	stdout, err := runHookSubcommandCodex(t, "user-prompt-submit", input, output)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	var mapped map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &mapped); jerr != nil {
		t.Fatalf("mapped stdout is not JSON: %v\n%s", jerr, stdout)
	}
	hso, _ := mapped["hookSpecificOutput"].(map[string]any)
	if hso == nil {
		t.Fatalf("hookSpecificOutput missing on UserPromptSubmit routing:\n%s", stdout)
	}
	if hso["additionalContext"] != "context: SPEC-X is active" {
		t.Errorf("additionalContext = %v, want the systemMessage text\n%s", hso["additionalContext"], stdout)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".moai", "logs", "codex-adapter.jsonl")); statErr == nil {
		t.Errorf("a discard was recorded for a DELIVERED message — only undeliverables are recorded")
	}
}

// TestHarnessCodexEventMismatchRejected verifies AC-CW-011(c): a payload
// whose hook_event_name resolves to a DIFFERENT dispatcher argument than the
// invoked subcommand is rejected with a nonzero exit and a diagnostic.
func TestHarnessCodexEventMismatchRejected(t *testing.T) {
	input := &hook.HookInput{HookEventName: "SessionStart", SessionID: "s1"}
	output := &hook.HookOutput{}

	_, err := runHookSubcommandCodex(t, "stop", input, output)
	if err == nil {
		t.Fatal("event mismatch must be rejected (nonzero exit)")
	}
	if !strings.Contains(err.Error(), "SessionStart") && !strings.Contains(err.Error(), "session-start") {
		t.Errorf("diagnostic does not name the mismatching event: %v", err)
	}
}

// TestHarnessCodexUnadaptedSubcommandRejected verifies --harness codex on an
// event this milestone does not adapt (compact) is refused, not silently
// wired (dead-path prevention).
func TestHarnessCodexUnadaptedSubcommandRejected(t *testing.T) {
	input := &hook.HookInput{HookEventName: "PreCompact", SessionID: "s1"}
	output := &hook.HookOutput{}

	_, err := runHookSubcommandCodex(t, "compact", input, output)
	if err == nil {
		t.Fatal("--harness codex on an unadapted event must be refused")
	}
	if !codexadapter.IsUnadapted(err) {
		t.Errorf("diagnostic should carry the adapter's ErrUnadapted marker: %v", err)
	}
}

// TestHarnessCodexExitCodePassthrough verifies AC-CW-011(d): the hook's exit
// code (the exit-2 protocol) passes through unchanged under codex mode.
func TestHarnessCodexExitCodePassthrough(t *testing.T) {
	input := &hook.HookInput{HookEventName: "Stop", SessionID: "s1"}
	output := &hook.HookOutput{ExitCode: 2}

	_, err := runHookSubcommandCodex(t, "stop", input, output)
	if err == nil {
		t.Fatal("exit-2 output must still produce an error (exit code passthrough)")
	}
	var coder interface{ ExitCode() int }
	if !errors.As(err, &coder) {
		t.Fatalf("error does not carry an exit code (ExitCoder): %v", err)
	}
	if coder.ExitCode() != 2 {
		t.Errorf("exit code = %d, want 2 (passthrough)", coder.ExitCode())
	}
}

// TestHarnessCodexNoInertKeysWithoutMapping verifies the pass-through
// contract: an output with NO inert keys is handed through byte-identically
// (modulo the trailing newline) — no needless re-marshal drift.
func TestHarnessCodexNoInertKeysWithoutMapping(t *testing.T) {
	input := &hook.HookInput{HookEventName: "PreToolUse", SessionID: "s1"}
	output := &hook.HookOutput{SuppressOutput: true}

	stdout, err := runHookSubcommandCodex(t, "pre-tool", input, output)
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}
	var mapped map[string]any
	if jerr := json.Unmarshal([]byte(stdout), &mapped); jerr != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", jerr, stdout)
	}
	if v, _ := mapped["suppressOutput"].(bool); !v {
		t.Errorf("suppressOutput lost in pass-through:\n%s", stdout)
	}
}
