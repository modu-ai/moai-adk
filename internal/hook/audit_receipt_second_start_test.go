package hook

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// A second SubagentStart carrying the same real agent id must keep the first
// start's StartedAt. Overwriting it moves the fence forward past receipts the
// instance minted between the two starts, and the later stop then refuses them
// as "created before the auditor started" (card t1951).
func TestSubagentStart_SecondStartSameAgentIDKeepsFirstStartedAt(t *testing.T) {
	root := newGateTree(t, "required")
	t0 := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)
	t2 := t0.Add(2 * time.Minute)

	orig := auditreceipt.Now
	t.Cleanup(func() { auditreceipt.Now = orig })

	start := &HookInput{CWD: root, AgentID: "agent-same-id", AgentType: auditreceipt.AgentPlanAuditor, SessionID: "sess-t1951"}
	h := NewSubagentStartHandler()

	auditreceipt.Now = func() time.Time { return t0 }
	if _, err := h.Handle(context.Background(), start); err != nil {
		t.Fatalf("first SubagentStart: %v", err)
	}
	receiptID := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: t1,
	})
	auditreceipt.Now = func() time.Time { return t2 }
	if _, err := h.Handle(context.Background(), start); err != nil {
		t.Fatalf("second SubagentStart: %v", err)
	}

	m, err := auditreceipt.ReadStartMarker(root, "agent-same-id")
	if err != nil {
		t.Fatalf("ReadStartMarker: %v", err)
	}
	if !m.StartedAt.Equal(t0) {
		t.Fatalf("StartedAt = %s after the second start, want the first start %s", m.StartedAt, t0)
	}

	stop := &HookInput{
		CWD:                  root,
		AgentID:              "agent-same-id",
		AgentType:            auditreceipt.AgentPlanAuditor,
		SessionID:            "sess-t1951",
		LastAssistantMessage: "AUDIT-VERDICT: PASS spec=SPEC-T1951-001 receipts=" + receiptID,
		HookEventName:        string(EventSubagentStop),
	}
	out := runStop(t, stop)
	if out != nil && strings.Contains(out.Reason, auditreceipt.CauseReceiptBeforeStart) {
		t.Fatalf("receipt minted between the two starts was refused as before-start: %+v", out)
	}
}
