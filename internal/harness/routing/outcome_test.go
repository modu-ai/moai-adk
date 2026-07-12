package routing

import "testing"

// TestDeriveOutcome asserts the fixed precedence of REQ-HEV-013 (AC-HEV-010):
// abort-kind > non-zero gate_exit > terminal passing signal > non-terminal.
// There is no code path that accepts a free-text outcome override (AP-1).
func TestDeriveOutcome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		refs        []EvidenceRef
		wantOutcome Outcome
		wantTerm    bool
	}{
		{
			name:     "empty evidence stays pending",
			refs:     nil,
			wantTerm: false,
		},
		{
			name: "abort marker wins over everything",
			refs: []EvidenceRef{
				{Kind: KindGateExit, Value: "0", Terminal: true},
				{Kind: KindAbort, Value: "interrupt"},
			},
			wantOutcome: OutcomeAbort,
			wantTerm:    true,
		},
		{
			name: "abort marker outranks a failing gate",
			refs: []EvidenceRef{
				{Kind: KindGateExit, Value: "1", Terminal: true},
				{Kind: KindAbort, Value: "killed"},
			},
			wantOutcome: OutcomeAbort,
			wantTerm:    true,
		},
		{
			name:        "non-zero gate_exit -> fail (even non-terminal)",
			refs:        []EvidenceRef{{Kind: KindGateExit, Value: "2"}},
			wantOutcome: OutcomeFail,
			wantTerm:    true,
		},
		{
			name:        "non-zero gate_exit outranks a passing signal",
			refs: []EvidenceRef{
				{Kind: KindVerifyPath, Ref: "x.log", Terminal: true},
				{Kind: KindGateExit, Value: "1"},
			},
			wantOutcome: OutcomeFail,
			wantTerm:    true,
		},
		{
			name:        "terminal gate_exit 0 -> success",
			refs:        []EvidenceRef{{Kind: KindGateExit, Value: "0", Terminal: true}},
			wantOutcome: OutcomeSuccess,
			wantTerm:    true,
		},
		{
			name:        "terminal audit_score -> success",
			refs:        []EvidenceRef{{Kind: KindAuditScore, Value: "0.91", Terminal: true}},
			wantOutcome: OutcomeSuccess,
			wantTerm:    true,
		},
		{
			name:        "terminal verify_path -> success",
			refs:        []EvidenceRef{{Kind: KindVerifyPath, Ref: "1-go-test.log", Terminal: true}},
			wantOutcome: OutcomeSuccess,
			wantTerm:    true,
		},
		{
			name:     "gate_exit 0 but NOT terminal stays pending",
			refs:     []EvidenceRef{{Kind: KindGateExit, Value: "0", Terminal: false}},
			wantTerm: false,
		},
		{
			name:     "audit_score without terminal stays pending",
			refs:     []EvidenceRef{{Kind: KindAuditScore, Value: "0.9", Terminal: false}},
			wantTerm: false,
		},
		{
			name:     "empty gate_exit value is not a fail signal",
			refs:     []EvidenceRef{{Kind: KindGateExit, Value: ""}},
			wantTerm: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, term := DeriveOutcome(tt.refs)
			if term != tt.wantTerm {
				t.Fatalf("terminal = %v, want %v (refs=%+v)", term, tt.wantTerm, tt.refs)
			}
			if term && got != tt.wantOutcome {
				t.Fatalf("outcome = %q, want %q", got, tt.wantOutcome)
			}
		})
	}
}
