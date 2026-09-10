// Package cli — tests for the `moai hook codex-review-gate` CLI wiring
// (SPEC-CODEX-COVER-RESIDUAL-001, REQ-CCR-001..004 / AC-CCR-001..005).
//
// codex_review_gate_test.go pins the PURE gate logic (HandleCodexReviewGate).
// This file pins the surrounding WIRING that makes the gate runnable:
//   - the cobra RunE fail-OPEN contract on a stdin the parser rejects,
//   - the full stdin → project-dir resolve → config read → handler → stdout
//     chain on the well-formed default path,
//   - the fail-OPEN contract on a handler error, with the reason on stderr,
//   - BLOCK propagation through the RunE.
//
// This file is the codex-side counterpart of multi_review_gate_wiring_test.go,
// whose header claims a one-for-one symmetry that had never been built: the
// multi gate got the wiring tests, the codex gate did not, and runCodexReviewGate
// sat at 0.0% coverage. Subcommand registration is NOT re-tested here —
// TestCodexReviewGate_SubcommandRegistered (codex_review_gate_test.go) already
// pins it.
//
// None of these tests may call t.Parallel(): every one mutates package-level
// seam variables (reviewGateChangeDetector, codexRunner, codexLookPath,
// codexSession), so parallel execution would race on shared state.
//
// @MX:SPEC: SPEC-CODEX-COVER-RESIDUAL-001
package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newCodexGateCmd builds a throwaway cobra command wired to the
// codex-review-gate RunE with in-memory stdin/stdout/stderr, so the fail-open
// paths are testable without touching the process streams. It is deliberately
// NOT newGateCmd (multi_review_gate_wiring_test.go), which is hard-wired to
// runMultiReviewGate — both live at package scope, so the names must differ.
func newCodexGateCmd(stdin string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	var out, errOut bytes.Buffer
	cmd := &cobra.Command{Use: "codex-review-gate", RunE: runCodexReviewGate, SilenceUsage: true}
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	return cmd, &out, &errOut
}

// codexGatePayload marshals a well-formed Stop-hook payload naming projectDir.
func codexGatePayload(t *testing.T, sessionID, projectDir string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"session_id":  sessionID,
		"project_dir": projectDir,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return string(payload)
}

// --- AC-CCR-001: malformed stdin fails open ---

// TestRunCodexReviewGate_InvalidStdinFailsOpen proves a malformed stdin payload
// emits an empty ALLOW ({}) and exits 0. The Stop pipeline must NEVER be trapped
// by a parse failure; the diagnostic goes to stderr instead.
func TestRunCodexReviewGate_InvalidStdinFailsOpen(t *testing.T) {
	cmd, out, errOut := newCodexGateCmd("{not json")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("invalid stdin must not error (fail-open); got %v", err)
	}
	assertAllowJSON(t, out.String())
	if !strings.Contains(errOut.String(), "codex-review-gate:") {
		t.Errorf("stderr must carry a codex-review-gate diagnostic, got %q", errOut.String())
	}
}

// --- AC-CCR-002: empty stdin fails open ---

// TestRunCodexReviewGate_EmptyStdinFailsOpen proves an empty stdin (no hook
// payload at all) also fail-opens rather than propagating a parse error.
func TestRunCodexReviewGate_EmptyStdinFailsOpen(t *testing.T) {
	cmd, out, _ := newCodexGateCmd("")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("empty stdin must not error (fail-open); got %v", err)
	}
	assertAllowJSON(t, out.String())
}

// --- AC-CCR-003: happy path allows without consulting codex ---

// TestRunCodexReviewGate_HappyPathAllow proves the well-formed default path:
// the gate is opt-in and the temp project's workflow.yaml turns it off, so the
// reader reads false and the handler ALLOWs at step 1 without ever resolving
// the codex binary. Exercises the full stdin → project-dir resolve → config
// read → handler → stdout chain.
//
// withChangeDetector(t, true) is load-bearing, not decoration. It is inert on
// this green path (the disabled gate returns at step 1, long before the
// detector is consulted at step 3), but it is what makes mutant M2 — forcing
// `enabled := true` — able to reach the codexLookPath guard below. Without it
// the mutated handler would ALLOW at step 3 instead, because the production
// detector runs `git status` against a t.TempDir() that is not a git
// repository and therefore returns false.
func TestRunCodexReviewGate_HappyPathAllow(t *testing.T) {
	dir := writeWorkflowYAML(t, "workflow:\n  codex:\n    review_gate:\n      enabled: false\n")
	withChangeDetector(t, true) // even with changes present, a disabled gate ALLOWs
	withCodexLookPath(t, func(string) (string, error) {
		t.Fatal("codex must not be consulted when the gate is disabled")
		return "", nil
	})

	cmd, out, _ := newCodexGateCmd(codexGatePayload(t, "sess-happy", dir))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("happy path must not error; got %v", err)
	}
	assertAllowJSON(t, out.String())
}

// --- AC-CCR-004: handler error fails open, with the reason on stderr ---

// TestRunCodexReviewGate_HandlerErrorFailsOpen proves a handler error degrades
// to an empty ALLOW with a diagnostic on stderr, rather than propagating out of
// Execute() and trapping the Stop pipeline.
//
// The stderr assertion is the only discriminator available: the error branch
// and the success branch write byte-identical {} to stdout, so a stdout-only
// test would assert nothing about this arm.
//
// All three codex seams are swapped together under one t.Cleanup, copied from
// TestReviewGate_FailOpenOnCodexError (codex_review_gate_test.go). The
// codexLookPath line is mandatory, not one of three interchangeable seams: the
// production default is exec.LookPath and the handler consults it at step 4,
// BEFORE the session starts, so an unswapped fixture performs a real PATH
// lookup and the verdict then depends on whether the host happens to carry a
// codex binary.
func TestRunCodexReviewGate_HandlerErrorFailsOpen(t *testing.T) {
	dir := writeWorkflowYAML(t, "workflow:\n  codex:\n    review_gate:\n      enabled: true\n")
	withChangeDetector(t, true)
	prevRunner, prevLook, prevSess := codexRunner, codexLookPath, codexSession
	codexRunner = stubCodexRunner{}
	codexLookPath = func(string) (string, error) { return "/fake/codex", nil }
	codexSession = &fakeCodexSession{startErr: errFakeCodexCrash}
	t.Cleanup(func() { codexRunner, codexLookPath, codexSession = prevRunner, prevLook, prevSess })

	cmd, out, errOut := newCodexGateCmd(codexGatePayload(t, "sess-err", dir))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("handler error must not error out of Execute (fail-open); got %v", err)
	}
	assertAllowJSON(t, out.String())
	if !strings.Contains(errOut.String(), "codex-review-gate: error:") {
		t.Errorf("stderr must carry the handler-error diagnostic, got %q", errOut.String())
	}
}

// --- AC-CCR-005: BLOCK propagates through the RunE ---

// TestRunCodexReviewGate_BlockVerdictPropagates proves the RunE forwards the
// handler's BLOCK to stdout rather than substituting an ALLOW: with the gate
// enabled, a reviewable change present, and a canned session whose review text
// carries severity-tagged finding bullets, stdout decodes to
// {decision: "block", reason: ...}.
func TestRunCodexReviewGate_BlockVerdictPropagates(t *testing.T) {
	dir := writeWorkflowYAML(t, "workflow:\n  codex:\n    review_gate:\n      enabled: true\n")
	withChangeDetector(t, true)
	withCodexSession(t, codexSessionScript("- [P1] found issues\n- [P2] more issues"))

	cmd, out, _ := newCodexGateCmd(codexGatePayload(t, "sess-block", dir))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("BLOCK path must not error; got %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &decoded); err != nil {
		t.Fatalf("stdout must be valid JSON, got %q (%v)", out.String(), err)
	}
	if decoded["decision"] != "block" {
		t.Errorf("expected BLOCK to propagate through the RunE, got %q", out.String())
	}
	if reason, _ := decoded["reason"].(string); reason == "" {
		t.Errorf("BLOCK must carry a non-empty reason, got %q", out.String())
	}
}
