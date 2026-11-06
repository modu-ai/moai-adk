package hook

import (
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// SPEC-AUDIT-ENGINE-CONVERGENCE-001 M4 RED, sub-step S2b — what the auditor Stop
// step does with an INCONCLUSIVE last line (AC-AEC-002b; REQ-AEC-007; card t2017).
//
// The line is a parsed verdict line: it records no CauseVerdictLineMissing and no
// rejection, it is not a PASS so no receipt is demanded, it clears no refusal that
// is outstanding (a parsed FAIL clears a missing-verdict-line refusal; an audit
// that did not run answers nothing), and the hook's message states the verdict
// and, when a cited receipt records one, the error class.

const inconclusiveLine = "AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts=none"

// A first INCONCLUSIVE stop is a parsed line: no block, no refusal recorded, the
// start marker consumed, and no phase-entry spawn is denied on its account.
func TestSubagentStop_InconclusiveLastLine_IsAParsedLine_NotAMissingOne(t *testing.T) {
	root := newGateTree(t, "required")
	seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, time.Now().UTC())

	out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a1", "no audit ran\n"+inconclusiveLine, false))
	if out.Decision != "" {
		t.Errorf("decision = %q, want none: INCONCLUSIVE is a legitimate way to end an audit", out.Decision)
	}
	if got, err := auditreceipt.ListRejections(root); err != nil || len(got) != 0 {
		t.Errorf("INCONCLUSIVE recorded refusals %+v (err %v); it must record none, least of all %q", got, err, auditreceipt.CauseVerdictLineMissing)
	}
	if _, err := auditreceipt.ReadStartMarker(root, "a1"); err == nil {
		t.Errorf("the start marker must be consumed by a parsed non-PASS line, as after FAIL")
	}
	if isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent"))); isDenied {
		t.Errorf("an INCONCLUSIVE stop with nothing outstanding must not deny a phase-entry spawn: %s", reason)
	}
}

// An INCONCLUSIVE stop clears nothing: a refusal that is outstanding stays, and the
// phase-entry spawn stays denied. A parsed FAIL, by contrast, clears the
// missing-verdict-line refusal (card t1855) — the control that makes the
// difference visible.
func TestSubagentStop_InconclusiveLastLine_ClearsNoOutstandingRefusal(t *testing.T) {
	seed := func(t *testing.T, root string) {
		t.Helper()
		for _, r := range []auditreceipt.Rejection{
			{AgentType: auditreceipt.AgentPlanAuditor, SpecID: auditreceipt.UnknownSpec, Cause: auditreceipt.CauseVerdictLineMissing},
			{AgentType: auditreceipt.AgentPlanAuditor, SpecID: "SPEC-Z-001", Cause: auditreceipt.CauseNoReceiptCited},
		} {
			r := r
			if err := auditreceipt.WriteRejection(root, &r); err != nil {
				t.Fatalf("WriteRejection: %v", err)
			}
		}
	}

	t.Run("control: a parsed FAIL clears the missing-verdict-line refusal", func(t *testing.T) {
		root := newGateTree(t, "required")
		seed(t, root)
		runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "f1", "AUDIT-VERDICT: FAIL spec=SPEC-X-001 receipts=none", false))
		got, err := auditreceipt.ListRejections(root)
		if err != nil || len(got) != 1 || got[0].SpecID != "SPEC-Z-001" {
			t.Errorf("after FAIL: %+v (err %v), want only the SPEC-Z-001 refusal", got, err)
		}
	})

	t.Run("INCONCLUSIVE leaves both refusals and the spawn stays denied", func(t *testing.T) {
		root := newGateTree(t, "required")
		seed(t, root)
		out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "i1", inconclusiveLine, false))
		if out.Decision != "" {
			t.Errorf("decision = %q, want none", out.Decision)
		}
		got, err := auditreceipt.ListRejections(root)
		if err != nil || len(got) != 2 {
			t.Errorf("after INCONCLUSIVE: %+v (err %v), want both refusals untouched", got, err)
		}
		for _, r := range got {
			if r.SpecID == auditreceipt.UnknownSpec && r.Cause != auditreceipt.CauseVerdictLineMissing {
				t.Errorf("the unknown-spec refusal was rewritten: %+v", r)
			}
		}
		isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent")))
		if !isDenied || !strings.Contains(reason, "SPEC-Z-001") {
			t.Errorf("the spawn must stay denied for the outstanding refusal; denied=%v reason=%q", isDenied, reason)
		}
	})
}

// The hook's message carries the error class the cited receipt records, so the
// reader learns why the audit did not run; with no receipt it states the verdict
// and claims no class.
func TestSubagentStop_InconclusiveLastLine_MessageCarriesTheCitedReceiptsClass(t *testing.T) {
	t.Run("a cited receipt with a class", func(t *testing.T) {
		root := newGateTree(t, "required")
		rid := seedReceipt(t, root, auditreceipt.Receipt{
			Tool: auditreceipt.ToolAuditMulti, TreeRoot: root, CodexVerdict: "inconclusive",
			ErrorClass: "quota", CreatedAt: time.Now().UTC(),
		})
		out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "c1",
			"AUDIT-VERDICT: INCONCLUSIVE spec=SPEC-X-001 receipts="+rid, false))
		if out.Decision != "" {
			t.Errorf("decision = %q, want none", out.Decision)
		}
		for _, want := range []string{"INCONCLUSIVE", "quota"} {
			if !strings.Contains(out.SystemMessage, want) {
				t.Errorf("the message does not carry %q: %q", want, out.SystemMessage)
			}
		}
	})
	t.Run("no receipt cited", func(t *testing.T) {
		root := newGateTree(t, "required")
		out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "c2", inconclusiveLine, false))
		if !strings.Contains(out.SystemMessage, "INCONCLUSIVE") {
			t.Errorf("the message must state the verdict: %q", out.SystemMessage)
		}
		for _, class := range []string{"quota", "timeout", "auth", "unavailable"} {
			if strings.Contains(out.SystemMessage, class) {
				t.Errorf("no class is known, yet the message names %q: %q", class, out.SystemMessage)
			}
		}
	})
}
