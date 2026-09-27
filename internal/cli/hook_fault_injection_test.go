package cli

// SPEC-DUAL-HARNESS-HOOK-PARITY-001 M2c — AC-HPR-008 unit leg (REQ-HPR-009).
// Faults on a decision-bearing event under `--harness codex` must never reach
// Codex as an allow: a timeout, a handler error (the in-process analogue of
// exit 1), unparseable output, and an explicit exit 2. The live leg — what the
// Codex host itself does when the process is killed at its hooks.json timeout
// — is TestLiveHookFaultOutcome, which this SPEC does not run (operator Q5).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/hook"
)

// faultRegistry serves a canned dispatch result: an output, or an error.
type faultRegistry struct {
	output *hook.HookOutput
	err    error
}

func (r *faultRegistry) Register(_ hook.Handler)                  {}
func (r *faultRegistry) Handlers(_ hook.EventType) []hook.Handler { return nil }
func (r *faultRegistry) Dispatch(_ context.Context, _ hook.EventType, _ *hook.HookInput) (*hook.HookOutput, error) {
	return r.output, r.err
}

// corruptProtocol serializes every output as bytes no JSON parser accepts.
type corruptProtocol struct{ input *hook.HookInput }

func (p *corruptProtocol) ReadInput(_ io.Reader) (*hook.HookInput, error) { return p.input, nil }
func (p *corruptProtocol) WriteOutput(w io.Writer, _ *hook.HookOutput) error {
	_, err := io.WriteString(w, `{"hookSpecificOutput":{"permissionDecision":`)
	return err
}

// runCodexHookWithDeps runs one hook subcommand under --harness codex with the
// given deps, in an isolated project root. It returns stdout, the project
// root, and the RunE error.
func runCodexHookWithDeps(t *testing.T, subcommand string, d *Dependencies) (string, string, error) {
	t.Helper()

	root := t.TempDir()
	t.Setenv("CLAUDE_PROJECT_DIR", root)

	origDeps := deps
	deps = d
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
	if err := sub.ParseFlags([]string{"--harness", "codex"}); err != nil {
		t.Fatalf("parse --harness codex: %v", err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	sub.SetContext(context.Background())

	var runErr error
	stdout := captureStdoutDuring(t, func() { runErr = sub.RunE(sub, []string{}) })
	return stdout, root, runErr
}

// assertCodexDeny fails unless stdout is a Codex deny for ev with a non-empty
// reason. The empty object and any allow are the failures this AC exists for.
func assertCodexDeny(t *testing.T, ev hook.EventType, stdout string) {
	t.Helper()
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" || trimmed == "{}" {
		t.Fatalf("%s: stdout = %q, want a fail-closed deny (Codex resolves an empty or {} output as allow)", ev, trimmed)
	}
	if strings.Contains(trimmed, `"allow"`) {
		t.Fatalf("%s: stdout carries an allow: %s", ev, trimmed)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(trimmed), &v); err != nil {
		t.Fatalf("%s: stdout is not JSON: %q", ev, trimmed)
	}
	var reason string
	switch ev {
	case hook.EventPreToolUse:
		hso, _ := v["hookSpecificOutput"].(map[string]any)
		if hso["permissionDecision"] != "deny" {
			t.Fatalf("%s: permissionDecision = %v, want deny: %s", ev, hso["permissionDecision"], trimmed)
		}
		reason, _ = hso["permissionDecisionReason"].(string)
	case hook.EventPermissionRequest:
		hso, _ := v["hookSpecificOutput"].(map[string]any)
		dec, _ := hso["decision"].(map[string]any)
		if dec["behavior"] != "deny" {
			t.Fatalf("%s: decision.behavior = %v, want deny: %s", ev, dec["behavior"], trimmed)
		}
		reason, _ = dec["message"].(string)
	case hook.EventStop:
		if v["decision"] != "block" {
			t.Fatalf("%s: decision = %v, want block: %s", ev, v["decision"], trimmed)
		}
		reason, _ = v["reason"].(string)
	}
	if strings.TrimSpace(reason) == "" {
		t.Fatalf("%s: deny carries no reason; Codex rejects a blank-reason deny: %s", ev, trimmed)
	}
}

// assertFaultRecorded fails unless the adapter sink under root holds a record
// of the fault, so a fail-closed deny is visible rather than silent.
func assertFaultRecorded(t *testing.T, ev hook.EventType, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, codexadapter.DiagnosticSinkRel))
	if err != nil {
		t.Fatalf("%s: no diagnostic sink record for the fault: %v", ev, err)
	}
	if !strings.Contains(string(data), string(ev)) {
		t.Fatalf("%s: sink does not record this event's fault:\n%s", ev, data)
	}
}

// TestHookFaultInjection is the AC-HPR-008 unit leg.
func TestHookFaultInjection(t *testing.T) {
	type event struct {
		ev  hook.EventType
		sub string
	}
	dispatched := []event{{hook.EventPreToolUse, "pre-tool"}, {hook.EventStop, "stop"}}

	faults := []struct {
		name string
		deps func(input *hook.HookInput) *Dependencies
	}{
		{"timeout", func(in *hook.HookInput) *Dependencies {
			return &Dependencies{
				HookRegistry: &faultRegistry{err: fmt.Errorf("%w: %v", hook.ErrHookTimeout, context.DeadlineExceeded)},
				HookProtocol: &codexHarnessProtocol{input: in},
			}
		}},
		{"exit 1 (handler error)", func(in *hook.HookInput) *Dependencies {
			return &Dependencies{
				HookRegistry: &faultRegistry{err: errors.New("handler 0: boom")},
				HookProtocol: &codexHarnessProtocol{input: in},
			}
		}},
		{"unparseable output", func(in *hook.HookInput) *Dependencies {
			return &Dependencies{
				HookRegistry: &faultRegistry{output: &hook.HookOutput{}},
				HookProtocol: &corruptProtocol{input: in},
			}
		}},
	}

	for _, e := range dispatched {
		for _, f := range faults {
			t.Run(string(e.ev)+"/"+f.name, func(t *testing.T) {
				in := &hook.HookInput{HookEventName: string(e.ev), SessionID: "s1"}
				stdout, root, err := runCodexHookWithDeps(t, e.sub, f.deps(in))
				if err != nil {
					t.Fatalf("RunE = %v; a fault must be answered with a fail-closed deny on exit 0, which Codex reads", err)
				}
				assertCodexDeny(t, e.ev, stdout)
				assertFaultRecorded(t, e.ev, root)
			})
		}

		// exit 2: the exit code is the block on Codex and passes through
		// unchanged (SPEC-CODEX-HOOK-ADAPTER-001 REQ-4). Adversarial output: the
		// handler also carries an allow, which must not reach stdout.
		t.Run(string(e.ev)+"/exit 2", func(t *testing.T) {
			in := &hook.HookInput{HookEventName: string(e.ev), SessionID: "s1"}
			out := &hook.HookOutput{ExitCode: 2}
			if e.ev == hook.EventPreToolUse {
				out = hook.NewAllowOutput()
				out.ExitCode = 2
			}
			stdout, _, err := runCodexHookWithDeps(t, e.sub, &Dependencies{
				HookRegistry: &faultRegistry{output: out},
				HookProtocol: &codexHarnessProtocol{input: in},
			})
			var coder interface{ ExitCode() int }
			if !errors.As(err, &coder) || coder.ExitCode() != 2 {
				t.Fatalf("RunE = %v, want exit code 2 to pass through", err)
			}
			if strings.Contains(stdout, `"allow"`) {
				t.Fatalf("exit-2 stdout carries an allow: %s", stdout)
			}
		})
	}

	// PermissionRequest is not an adapted Codex row until M2e (events.go), so
	// the dispatcher refuses it before any handler runs. The fail-closed writer
	// the dispatcher calls is exercised directly for it.
	for _, f := range []struct {
		name  string
		cause error
	}{
		{"timeout", fmt.Errorf("%w: %v", hook.ErrHookTimeout, context.DeadlineExceeded)},
		{"exit 1 (handler error)", errors.New("handler 0: boom")},
		{"unparseable output", fmt.Errorf("map hook output for codex: %w", &json.SyntaxError{})},
	} {
		t.Run(string(hook.EventPermissionRequest)+"/"+f.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("CLAUDE_PROJECT_DIR", root)
			var err error
			stdout := captureStdoutDuring(t, func() { err = writeCodexFailClosed(hook.EventPermissionRequest, f.cause) })
			if err != nil {
				t.Fatalf("writeCodexFailClosed = %v", err)
			}
			assertCodexDeny(t, hook.EventPermissionRequest, stdout)
			assertFaultRecorded(t, hook.EventPermissionRequest, root)
		})
	}
}
