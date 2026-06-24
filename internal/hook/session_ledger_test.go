package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/telemetry"
)

// rec is a small constructor helper for building UsageRecord fixtures in tests.
func rec(phase, agentType, outcome, pathKind string, pass, fail bool) telemetry.UsageRecord {
	return telemetry.UsageRecord{
		SessionID:  "sess-test",
		Phase:      phase,
		AgentType:  agentType,
		Outcome:    outcome,
		PathKind:   pathKind,
		IsTestPass: pass,
		IsTestFail: fail,
	}
}

// TestBuildSessionLedger verifies that buildSessionLedger consumes a
// []telemetry.UsageRecord and aggregates the session-level signals
// (AC-SEG-001 signature + aggregation correctness).
func TestBuildSessionLedger(t *testing.T) {
	t.Parallel()

	records := []telemetry.UsageRecord{
		rec("run", "manager-develop", telemetry.OutcomeSuccess, "", false, true),
		rec("run", "manager-develop", telemetry.OutcomeUnknown, "", false, false),
	}

	ledger := buildSessionLedger(records)

	if ledger.SessionID != "sess-test" {
		t.Errorf("SessionID = %q, want %q", ledger.SessionID, "sess-test")
	}
	if ledger.SuccessClaims != 1 {
		t.Errorf("SuccessClaims = %d, want 1", ledger.SuccessClaims)
	}
	if ledger.BinaryPass {
		t.Errorf("BinaryPass = true, want false")
	}
	if !ledger.BinaryFail {
		t.Errorf("BinaryFail = false, want true")
	}
	if ledger.PathKind != telemetry.PathKindCodeChange {
		t.Errorf("PathKind = %q, want %q", ledger.PathKind, telemetry.PathKindCodeChange)
	}
	if len(ledger.Records) != 2 {
		t.Errorf("Records len = %d, want 2", len(ledger.Records))
	}
}

// TestInferPathKind covers the path-kind classification taxonomy (AC-SEG-011 +
// AC-SEG-003): explicit wins, phase inference, ambiguous fallback to unknown.
func TestInferPathKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		records []telemetry.UsageRecord
		want    string
	}{
		{
			name:    "explicit_pathkind_wins",
			records: []telemetry.UsageRecord{rec("sync", "manager-docs", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, false, false)},
			want:    telemetry.PathKindCodeChange,
		},
		{
			name:    "phase_sync",
			records: []telemetry.UsageRecord{rec("sync", "", telemetry.OutcomeSuccess, "", false, false)},
			want:    telemetry.PathKindDocsOnly,
		},
		{
			name:    "agent_manager_docs",
			records: []telemetry.UsageRecord{rec("", "manager-docs", telemetry.OutcomeSuccess, "", false, false)},
			want:    telemetry.PathKindDocsOnly,
		},
		{
			name:    "phase_run",
			records: []telemetry.UsageRecord{rec("run", "", telemetry.OutcomeSuccess, "", false, false)},
			want:    telemetry.PathKindCodeChange,
		},
		{
			name:    "phase_plan",
			records: []telemetry.UsageRecord{rec("plan", "", telemetry.OutcomeSuccess, "", false, false)},
			want:    telemetry.PathKindCodeChange,
		},
		{
			name:    "agent_manager_develop",
			records: []telemetry.UsageRecord{rec("", "manager-develop", telemetry.OutcomeSuccess, "", false, false)},
			want:    telemetry.PathKindCodeChange,
		},
		{
			name:    "ambiguous",
			records: []telemetry.UsageRecord{rec("none", "", telemetry.OutcomeUnknown, "", false, false)},
			want:    telemetry.PathKindUnknown,
		},
		{
			name: "mixed_code_wins_over_docs",
			// 한 세션에 docs + code 레코드 혼재 시 code-change 우선 (보수적, edge case #4).
			records: []telemetry.UsageRecord{
				rec("sync", "manager-docs", telemetry.OutcomeSuccess, "", false, false),
				rec("run", "manager-develop", telemetry.OutcomeSuccess, "", false, false),
			},
			want: telemetry.PathKindCodeChange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := inferPathKind(tt.records); got != tt.want {
				t.Errorf("inferPathKind = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEvaluateEvidence covers the core decision logic (design §0.4) across all
// REQ-mapped cases:
//   - AC-SEG-002: binary-evidence-first (flip IsTestPass false→true flips verdict)
//   - AC-SEG-003: docs-only exempt
//   - AC-SEG-004: code-change + success + observed fail + no pass → finding
//   - AC-SEG-008/010: legacy (no binary signal) → no finding
//   - AC-SEG-011: unknown → no finding
func TestEvaluateEvidence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		records     []telemetry.UsageRecord
		wantFinding bool
	}{
		{
			// AC-SEG-002 (1) + AC-SEG-004: success Outcome, a binary signal observed
			// (IsTestFail=true) but NO IsTestPass → finding.
			name:        "code_change_success_no_pass_binary_present",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, false, true)},
			wantFinding: true,
		},
		{
			// AC-SEG-002 (2): SAME success Outcome but IsTestPass=true → NO finding.
			// The only delta from the case above is the binary signal → proves the
			// decision pivots on IsTestPass, not on Outcome.
			name:        "success_test_pass_observed",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, true, true)},
			wantFinding: false,
		},
		{
			// AC-SEG-003: docs-only session with success claim + no test-pass → no finding.
			name:        "docs_only_success_no_test_pass",
			records:     []telemetry.UsageRecord{rec("sync", "manager-docs", telemetry.OutcomeSuccess, telemetry.PathKindDocsOnly, false, false)},
			wantFinding: false,
		},
		{
			// AC-SEG-008 / AC-SEG-010: legacy record — success claim but NO binary
			// signal observable (both false, zero-value) → no finding (absent ≠ failed).
			name:        "legacy_success_no_binary",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeSuccess, "", false, false)},
			wantFinding: false,
		},
		{
			// AC-SEG-011: unknown path-kind never flagged.
			name:        "unknown_pathkind",
			records:     []telemetry.UsageRecord{rec("none", "", telemetry.OutcomeSuccess, "", false, true)},
			wantFinding: false,
		},
		{
			// design §0.4: success + observed fail + no pass (binaryObservable via fail) → finding.
			name:        "code_change_success_only_fail_observed",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, false, true)},
			wantFinding: true,
		},
		{
			// edge case #5: no success claim (all unknown/error) → no finding.
			name:        "no_success_claim",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeError, telemetry.PathKindCodeChange, false, true)},
			wantFinding: false,
		},
		{
			// edge case #6: IsTestPass=true AND IsTestFail=true (mixed) → backed, no finding.
			name:        "mixed_pass_and_fail",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, true, true)},
			wantFinding: false,
		},
		{
			// partial outcome counts as a success claim (Outcome in {success, partial}).
			name:        "code_change_partial_no_pass_binary_present",
			records:     []telemetry.UsageRecord{rec("run", "manager-develop", telemetry.OutcomePartial, telemetry.PathKindCodeChange, false, true)},
			wantFinding: true,
		},
		{
			// empty records → no finding (edge case #1).
			name:        "empty_records",
			records:     []telemetry.UsageRecord{},
			wantFinding: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ledger := buildSessionLedger(tt.records)
			finding := evaluateEvidence(ledger)
			if tt.wantFinding && finding == nil {
				t.Errorf("expected a non-nil Finding, got nil")
			}
			if !tt.wantFinding && finding != nil {
				t.Errorf("expected nil Finding, got %+v", finding)
			}
		})
	}
}

// TestFindingHumanReadable verifies the advisory finding's human-readable output
// names the path-kind and the success-claim count (AC-SEG-004 + design §0.5).
func TestFindingHumanReadable(t *testing.T) {
	t.Parallel()

	records := []telemetry.UsageRecord{
		rec("run", "manager-develop", telemetry.OutcomeSuccess, telemetry.PathKindCodeChange, false, true),
	}
	ledger := buildSessionLedger(records)
	finding := evaluateEvidence(ledger)
	if finding == nil {
		t.Fatal("expected a non-nil Finding")
	}

	hr := finding.HumanReadable()
	if !strings.Contains(hr, "path-kind") {
		t.Errorf("HumanReadable must name the path-kind, got: %s", hr)
	}
	if !strings.Contains(hr, telemetry.PathKindCodeChange) {
		t.Errorf("HumanReadable must include %q, got: %s", telemetry.PathKindCodeChange, hr)
	}
	// success-claim count must appear (the count is 1 here).
	if !strings.Contains(hr, "success") {
		t.Errorf("HumanReadable must reference the success claim, got: %s", hr)
	}

	// slogArgs must be non-empty and even-length (key/value pairs).
	args := finding.slogArgs()
	if len(args) == 0 || len(args)%2 != 0 {
		t.Errorf("slogArgs must be a non-empty even-length key/value slice, got %d items", len(args))
	}
}
