package hook

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// writeSessionRecord is a test helper that writes a single UsageRecord as a
// JSONL line to <projectRoot>/.moai/evolution/telemetry/usage-<today>.jsonl,
// so that telemetry.LoadBySession picks it up. This exercises the gate's real
// LoadBySession read path (not a mock).
func writeSessionRecord(t *testing.T, projectRoot string, r telemetry.UsageRecord) {
	t.Helper()
	r.Timestamp = time.Now().UTC()
	if err := telemetry.RecordSkillUsage(projectRoot, r); err != nil {
		t.Fatalf("writeSessionRecord: %v", err)
	}
}

// newMoaiProject creates a temp dir with a .moai/ marker so resolveProjectRoot
// / RecordSkillUsage guards pass.
func newMoaiProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestRunEvidenceGate_ReadOnlyNoBlock verifies the gate entry function exists,
// is void (cannot block — it returns nothing for Handle to act on), and runs
// without panic for both a finding-present and a finding-absent session
// (REQ-SEG-005 fail-open at the function level). This forces the RED until
// runEvidenceGate is implemented.
func TestRunEvidenceGate_ReadOnlyNoBlock(t *testing.T) {
	t.Parallel()

	dir := newMoaiProject(t)
	sid := "sess-gate-direct"
	writeSessionRecord(t, dir, telemetry.UsageRecord{
		SessionID: sid, Phase: "run", AgentType: "manager-develop",
		Outcome: telemetry.OutcomeSuccess, PathKind: telemetry.PathKindCodeChange,
		IsTestFail: true,
	})

	// runEvidenceGate is void: it surfaces advisories via slog/stderr only and
	// returns nothing the caller could use to block. Calling it must not panic.
	runEvidenceGate(dir, sid)

	// Empty session + nonexistent project both fail-open (no panic, early return).
	runEvidenceGate(dir, "sess-nonexistent")
	runEvidenceGate("", "")
}

// TestStopEvidenceGate_FailOpen verifies AC-SEG-005: Handle() returns the empty
// allow output (&HookOutput{}) with nil error on every path — finding present,
// finding absent, and StopHookActive=true. The gate NEVER blocks the stop event.
func TestStopEvidenceGate_FailOpen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) *HookInput
	}{
		{
			name: "finding_present_code_change_unbacked_success",
			setup: func(t *testing.T) *HookInput {
				dir := newMoaiProject(t)
				sid := "sess-finding"
				// code-change session with success claim + observed fail + no pass → finding.
				writeSessionRecord(t, dir, telemetry.UsageRecord{
					SessionID: sid, Phase: "run", AgentType: "manager-develop",
					Outcome: telemetry.OutcomeSuccess, PathKind: telemetry.PathKindCodeChange,
					IsTestFail: true,
				})
				return &HookInput{SessionID: sid, ProjectDir: dir, HookEventName: "Stop"}
			},
		},
		{
			name: "finding_absent_clean_session",
			setup: func(t *testing.T) *HookInput {
				dir := newMoaiProject(t)
				sid := "sess-clean"
				writeSessionRecord(t, dir, telemetry.UsageRecord{
					SessionID: sid, Phase: "run", AgentType: "manager-develop",
					Outcome: telemetry.OutcomeSuccess, PathKind: telemetry.PathKindCodeChange,
					IsTestPass: true,
				})
				return &HookInput{SessionID: sid, ProjectDir: dir, HookEventName: "Stop"}
			},
		},
		{
			name: "stop_hook_active_short_circuits",
			setup: func(t *testing.T) *HookInput {
				return &HookInput{SessionID: "sess-active", CWD: "/tmp", HookEventName: "Stop", StopHookActive: true}
			},
		},
		{
			name: "empty_session_no_records",
			setup: func(t *testing.T) *HookInput {
				dir := newMoaiProject(t)
				return &HookInput{SessionID: "sess-empty", ProjectDir: dir, HookEventName: "Stop"}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := tt.setup(t)
			h := NewStopHandler()
			got, err := h.Handle(context.Background(), input)
			if err != nil {
				t.Fatalf("Handle returned non-nil error: %v", err)
			}
			if got == nil {
				t.Fatal("Handle returned nil output")
			}
			// fail-open: must be the empty allow output — no stop-decision fields.
			if got.Decision != "" {
				t.Errorf("Decision = %q, want empty (gate must never block stop)", got.Decision)
			}
			if got.Reason != "" {
				t.Errorf("Reason = %q, want empty", got.Reason)
			}
			if got.HookSpecificOutput != nil {
				t.Errorf("HookSpecificOutput = %+v, want nil (Stop returns empty {})", got.HookSpecificOutput)
			}
		})
	}
}

// TestStopEvidenceGate_StdoutContractUnchanged verifies AC-SEG-006: when a
// finding fires, the returned HookOutput carries no stop-decision fields
// (Decision/Reason). The advisory goes to stderr/slog only; the stdout
// HookOutput JSON contract is unchanged.
func TestStopEvidenceGate_StdoutContractUnchanged(t *testing.T) {
	t.Parallel()

	dir := newMoaiProject(t)
	sid := "sess-stdout-contract"
	writeSessionRecord(t, dir, telemetry.UsageRecord{
		SessionID: sid, Phase: "run", AgentType: "manager-develop",
		Outcome: telemetry.OutcomeSuccess, PathKind: telemetry.PathKindCodeChange,
		IsTestFail: true, // observed fail, no pass → finding fires
	})

	h := NewStopHandler()
	got, err := h.Handle(context.Background(), &HookInput{SessionID: sid, ProjectDir: dir, HookEventName: "Stop"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The HookOutput must be the empty allow object — the gate adds nothing to stdout.
	if got.Decision != "" || got.Reason != "" || got.SystemMessage != "" {
		t.Errorf("gate polluted stdout HookOutput: %+v", got)
	}
}

// TestStopEvidenceGate_StopHookActiveSkipsGate verifies AC-SEG-009: when
// StopHookActive=true, the early-return guard (stop.go L44-47) short-circuits
// BEFORE the gate is reached. Even a finding-producing session must emit NO
// advisory to stderr, proving the gate is positioned after the guard. This test
// is non-parallel because it temporarily redirects os.Stderr.
func TestStopEvidenceGate_StopHookActiveSkipsGate(t *testing.T) {
	dir := newMoaiProject(t)
	sid := "sess-active-skips"
	// A session that WOULD produce a finding if the gate ran.
	writeSessionRecord(t, dir, telemetry.UsageRecord{
		SessionID: sid, Phase: "run", AgentType: "manager-develop",
		Outcome: telemetry.OutcomeSuccess, PathKind: telemetry.PathKindCodeChange,
		IsTestFail: true,
	})

	// Capture stderr around the Handle call.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	h := NewStopHandler()
	out, herr := h.Handle(context.Background(), &HookInput{
		SessionID: sid, ProjectDir: dir, HookEventName: "Stop", StopHookActive: true,
	})

	_ = w.Close()
	os.Stderr = origStderr
	captured, _ := io.ReadAll(r)

	if herr != nil {
		t.Fatalf("unexpected error: %v", herr)
	}
	if out == nil || out.Decision != "" {
		t.Errorf("StopHookActive must return empty allow, got %+v", out)
	}
	// The gate must NOT have run: no evidence-gate advisory on stderr.
	if strings.Contains(string(captured), "evidence-gate") {
		t.Errorf("gate fired under StopHookActive=true (guard not short-circuiting before gate); stderr: %s", string(captured))
	}
}

// TestStopEvidenceGate_LegacyFixtureNotFalseFlagged verifies AC-SEG-008(b) /
// AC-SEG-010 at runtime: a legacy JSONL record (success claim, NO new binary
// fields) written to a real telemetry file does NOT produce a finding. We
// assert via fail-open behavior (Handle returns allow) AND via the evaluator
// directly through buildSessionLedger over LoadBySession output.
func TestStopEvidenceGate_LegacyFixtureNotFalseFlagged(t *testing.T) {
	t.Parallel()

	dir := newMoaiProject(t)
	sid := "sess-legacy-runtime"
	// Legacy: code-bearing phase + success, but no binary fields set (zero-value).
	writeSessionRecord(t, dir, telemetry.UsageRecord{
		SessionID: sid, Phase: "run", AgentType: "manager-develop",
		Outcome: telemetry.OutcomeSuccess,
		// IsTestPass / IsTestFail / PathKind all zero-value (legacy record).
	})

	records, err := telemetry.LoadBySession(dir, sid)
	if err != nil {
		t.Fatalf("LoadBySession: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	ledger := buildSessionLedger(records)
	if finding := evaluateEvidence(ledger); finding != nil {
		t.Errorf("legacy record (no binary fields) must NOT be flagged, got finding: %+v", finding)
	}

	// And Handle stays fail-open.
	h := NewStopHandler()
	got, herr := h.Handle(context.Background(), &HookInput{SessionID: sid, ProjectDir: dir, HookEventName: "Stop"})
	if herr != nil || got == nil || got.Decision != "" {
		t.Errorf("Handle must stay allow for legacy session, got=%+v err=%v", got, herr)
	}
}
