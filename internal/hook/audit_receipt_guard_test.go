package hook

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/auditreceipt"
)

// newGateTree creates a git-backed project tree whose codex audit gate carries
// the given value ("" writes no config at all).
func newGateTree(t *testing.T, gate string) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", root).CombinedOutput(); err != nil {
		t.Skipf("git unavailable: %v (%s)", err, out)
	}
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if gate != "" {
		dir := filepath.Join(root, ".moai", "config", "sections")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		body := "workflow:\n  audit:\n    gates:\n      codex: " + gate + "\n"
		if err := os.WriteFile(filepath.Join(dir, "workflow.yaml"), []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return auditreceipt.Canonical(root)
}

func seedReceipt(t *testing.T, storeRoot string, r auditreceipt.Receipt) string {
	t.Helper()
	id, err := auditreceipt.WriteReceipt(storeRoot, &r)
	if err != nil {
		t.Fatalf("WriteReceipt: %v", err)
	}
	return id
}

func seedStart(t *testing.T, root, agentID, agentType string, at time.Time) {
	t.Helper()
	m := auditreceipt.StartMarker{AgentID: agentID, AgentType: agentType, TreeRoot: root, StartedAt: at}
	if err := auditreceipt.WriteStartMarker(root, &m); err != nil {
		t.Fatalf("WriteStartMarker: %v", err)
	}
}

func stopInput(root, agentType, agentID, message string, reentry bool) *HookInput {
	return &HookInput{
		CWD:                  root,
		AgentID:              agentID,
		AgentType:            agentType,
		LastAssistantMessage: message,
		StopHookActive:       reentry,
		HookEventName:        string(EventSubagentStop),
	}
}

func runStop(t *testing.T, input *HookInput) *HookOutput {
	t.Helper()
	out, err := NewSubagentStopHandler().Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("SubagentStop Handle: %v", err)
	}
	return out
}

func receiptSpawnInput(root, subagentType, toolName string) *HookInput {
	raw, _ := json.Marshal(map[string]string{"subagent_type": subagentType, "prompt": "go"})
	return &HookInput{CWD: root, ToolName: toolName, ToolInput: raw, HookEventName: "PreToolUse"}
}

func runSpawn(t *testing.T, input *HookInput) *HookOutput {
	t.Helper()
	// The spawn path also appends the agent-model audit log, which resolves its
	// directory from CLAUDE_PROJECT_DIR and would otherwise land inside the
	// package directory. Point it at the tree under test.
	t.Setenv("CLAUDE_PROJECT_DIR", input.CWD)
	out, err := NewPreToolHandler(nil, DefaultSecurityPolicy()).Handle(context.Background(), input)
	if err != nil {
		t.Fatalf("PreToolUse Handle: %v", err)
	}
	return out
}

func denied(out *HookOutput) (bool, string) {
	if out == nil || out.HookSpecificOutput == nil {
		return false, ""
	}
	h := out.HookSpecificOutput
	return h.PermissionDecision == DecisionDeny, h.PermissionDecisionReason
}

// AC-CAG-009: SubagentStart writes a start marker for the two auditors only.
func TestSubagentStart_WritesAuditorStartMarker(t *testing.T) {
	root := newGateTree(t, "required")
	sub := filepath.Join(root, "nested", "dir")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	h := NewSubagentStartHandler()
	for _, tc := range []struct {
		agentType string
		wantFile  bool
	}{
		{auditreceipt.AgentPlanAuditor, true},
		{auditreceipt.AgentSyncAuditor, true},
		{"manager-develop", false},
	} {
		input := &HookInput{CWD: sub, AgentID: "id-" + tc.agentType, AgentType: tc.agentType, SessionID: "s1"}
		if _, err := h.Handle(context.Background(), input); err != nil {
			t.Fatalf("SubagentStart Handle: %v", err)
		}
		got, err := auditreceipt.ReadStartMarker(root, "id-"+tc.agentType)
		switch {
		case tc.wantFile && err != nil:
			t.Fatalf("no start marker for %s: %v", tc.agentType, err)
		case tc.wantFile:
			if got.TreeRoot != root {
				t.Errorf("marker tree_root = %q, want the git toplevel %q", got.TreeRoot, root)
			}
			if got.StartedAt.IsZero() || got.AgentType != tc.agentType {
				t.Errorf("marker = %+v", got)
			}
		case !tc.wantFile && err == nil:
			t.Errorf("a start marker was written for %s, which is not an auditor", tc.agentType)
		}
	}
}

// AC-CAG-010: the first stop of an auditor PASS that cannot show a receipt is
// blocked, with a reason naming the failed condition, and a rejection record is
// persisted.
func TestSubagentStop_BlocksUnprovenPassAndPersistsRejection(t *testing.T) {
	other := newGateTree(t, "required")
	for _, agentType := range []string{auditreceipt.AgentPlanAuditor, auditreceipt.AgentSyncAuditor} {
		root := newGateTree(t, "required")
		start := time.Now().UTC()
		seedStart(t, root, "a1", agentType, start)
		r1 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(time.Second)})
		r2 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: other, CreatedAt: start.Add(time.Second)})
		r3 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(-time.Minute)})

		for name, tc := range map[string]struct {
			agentID   string
			receipts  string
			wantCause string
		}{
			"no-citation":   {"a1", "none", auditreceipt.CauseNoReceiptCited},
			"unknown":       {"a1", "rcpt-ffffffffffffffffffff", auditreceipt.CauseReceiptUnknown},
			"other-tree":    {"a1", r2, auditreceipt.CauseReceiptOtherTree},
			"before-start":  {"a1", r3, auditreceipt.CauseReceiptBeforeStart},
			"start-missing": {"no-marker", r1, auditreceipt.CauseStartMarkerMissing},
		} {
			t.Run(agentType+"/"+name, func(t *testing.T) {
				msg := "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=" + tc.receipts
				out := runStop(t, stopInput(root, agentType, tc.agentID, msg, false))
				if out.Decision != "block" {
					t.Fatalf("decision = %q, want block", out.Decision)
				}
				if !strings.Contains(out.Reason, tc.wantCause) {
					t.Errorf("reason = %q, want it to name %q", out.Reason, tc.wantCause)
				}
				rj, err := auditreceipt.ReadRejection(root, agentType, "SPEC-X-001")
				if err != nil {
					t.Fatalf("rejection record missing: %v", err)
				}
				if rj.Cause != tc.wantCause {
					t.Errorf("rejection cause = %q, want %q", rj.Cause, tc.wantCause)
				}
			})
		}

		// A PASS whose citation checks out is accepted.
		t.Run(agentType+"/valid-pass", func(t *testing.T) {
			msg := "AUDIT-VERDICT: PASS spec=SPEC-OK-001 receipts=" + r1
			out := runStop(t, stopInput(root, agentType, "a1", msg, false))
			if out.Decision != "" {
				t.Errorf("decision = %q, want none for a proven PASS", out.Decision)
			}
		})
	}
}

// AC-CAG-011: reentry warns without blocking, a valid PASS clears every
// rejection of that role, and a FAIL changes nothing.
func TestSubagentStop_ReentryWarnsAcceptanceClearsRoleFailIsInert(t *testing.T) {
	root := newGateTree(t, "required")
	start := time.Now().UTC()
	seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, start)
	r1 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(time.Second)})

	unproven := "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none"
	if out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a1", unproven, false)); out.Decision != "block" {
		t.Fatalf("setup: first stop decision = %q, want block", out.Decision)
	}
	for _, seed := range []auditreceipt.Rejection{
		{AgentType: auditreceipt.AgentPlanAuditor, SpecID: "SPEC-Z-001", Cause: auditreceipt.CauseNoReceiptCited},
		{AgentType: auditreceipt.AgentPlanAuditor, SpecID: auditreceipt.UnknownSpec, Cause: auditreceipt.CauseVerdictLineMissing},
		{AgentType: auditreceipt.AgentSyncAuditor, SpecID: "SPEC-X-001", Cause: auditreceipt.CauseNoReceiptCited},
	} {
		seed := seed
		if err := auditreceipt.WriteRejection(root, &seed); err != nil {
			t.Fatalf("WriteRejection: %v", err)
		}
	}

	// (i) reentry: no block, record kept and marked warned.
	out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a1", unproven, true))
	if out.Decision != "" {
		t.Errorf("reentry decision = %q, want none", out.Decision)
	}
	if !strings.Contains(out.SystemMessage, "not accepted") {
		t.Errorf("reentry systemMessage = %q, want it to state the PASS is not accepted", out.SystemMessage)
	}
	rj, err := auditreceipt.ReadRejection(root, auditreceipt.AgentPlanAuditor, "SPEC-X-001")
	if err != nil || !rj.ReentryWarned {
		t.Fatalf("rejection after reentry = %+v (err %v), want it kept with reentry_warned true", rj, err)
	}

	// (ii) a proven PASS clears every plan-auditor record, leaving the sync-auditor one.
	seedStart(t, root, "a2", auditreceipt.AgentPlanAuditor, start)
	out = runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a2", "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts="+r1, false))
	if out.Decision != "" {
		t.Fatalf("proven PASS decision = %q, want none", out.Decision)
	}
	remaining, err := auditreceipt.ListRejections(root)
	if err != nil {
		t.Fatalf("ListRejections: %v", err)
	}
	if len(remaining) != 1 || remaining[0].AgentType != auditreceipt.AgentSyncAuditor {
		t.Errorf("remaining rejections = %+v, want only the sync-auditor record", remaining)
	}

	// (iii) a FAIL neither blocks nor records.
	fresh := newGateTree(t, "required")
	out = runStop(t, stopInput(fresh, auditreceipt.AgentPlanAuditor, "b1", "AUDIT-VERDICT: FAIL spec=SPEC-Y-001 receipts=none", false))
	if out.Decision != "" || out.SystemMessage != "" {
		t.Errorf("FAIL output = %+v, want no decision and no message", out)
	}
	if got, err := auditreceipt.ListRejections(fresh); err != nil || len(got) != 0 {
		t.Errorf("FAIL left rejections %+v (err %v)", got, err)
	}
}

// AC-CAG-012: the PreToolUse consumer denies phase-entry spawns while a
// rejection is outstanding, and only those.
func TestPreToolUse_DeniesPhaseEntrySpawnsWhileRejectionOutstanding(t *testing.T) {
	blocked := newGateTree(t, "required")
	rj := auditreceipt.Rejection{AgentType: auditreceipt.AgentPlanAuditor, SpecID: "SPEC-X-001", Cause: auditreceipt.CauseNoReceiptCited}
	if err := auditreceipt.WriteRejection(blocked, &rj); err != nil {
		t.Fatalf("WriteRejection: %v", err)
	}
	clean := newGateTree(t, "required")
	advisory := newGateTree(t, "advisory")
	rj2 := rj
	if err := auditreceipt.WriteRejection(advisory, &rj2); err != nil {
		t.Fatalf("WriteRejection: %v", err)
	}

	for _, agent := range []string{"manager-develop", "manager-docs", "manager-git"} {
		for _, tool := range []string{"Agent", "Task"} {
			isDenied, reason := denied(runSpawn(t, receiptSpawnInput(blocked, agent, tool)))
			if !isDenied {
				t.Errorf("%s spawn via %s was not denied while a rejection is outstanding", agent, tool)
				continue
			}
			if !strings.HasPrefix(reason, "AUDIT_RECEIPT_VIOLATION") {
				t.Errorf("reason = %q, want the AUDIT_RECEIPT_VIOLATION sentinel prefix", reason)
			}
			for _, want := range []string{auditreceipt.AgentPlanAuditor, "SPEC-X-001", auditreceipt.CauseNoReceiptCited} {
				if !strings.Contains(reason, want) {
					t.Errorf("reason = %q, want it to name %q", reason, want)
				}
			}
		}
	}

	for _, tc := range []struct{ root, agent string }{
		{blocked, "Explore"},
		{clean, "manager-develop"},
		{advisory, "manager-develop"},
	} {
		if isDenied, reason := denied(runSpawn(t, receiptSpawnInput(tc.root, tc.agent, "Agent"))); isDenied && strings.HasPrefix(reason, "AUDIT_RECEIPT_VIOLATION") {
			t.Errorf("%s spawn was denied by the receipt guard, want allowed (%s)", tc.agent, reason)
		}
	}
}

// AC-CAG-013: a tree that did not declare the gate is untouched by all three
// additions.
func TestAuditReceiptGuard_NonRequiredTreesAreInert(t *testing.T) {
	for _, gate := range []string{"off", "advisory", ""} {
		root := newGateTree(t, gate)
		if _, err := NewSubagentStartHandler().Handle(context.Background(), &HookInput{CWD: root, AgentID: "a1", AgentType: auditreceipt.AgentPlanAuditor}); err != nil {
			t.Fatalf("SubagentStart Handle: %v", err)
		}
		out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a1", "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none", false))
		if out.Decision != "" || out.SystemMessage != "" {
			t.Errorf("gate %q: stop output = %+v, want silence", gate, out)
		}
		if _, err := os.Stat(auditreceipt.StateDir(root)); !os.IsNotExist(err) {
			t.Errorf("gate %q: the receipt store was created in a tree that never declared the gate", gate)
		}
	}
}

// AC-CAG-014: unreadable evidence never reads as an accepted PASS or an
// allowed spawn.
func TestAuditReceiptGuard_UnreadableEvidence(t *testing.T) {
	root := newGateTree(t, "required")
	start := time.Now().UTC()
	seedStart(t, root, "a1", auditreceipt.AgentPlanAuditor, start)
	corrupt := "rcpt-" + "00112233445566778899"
	path := filepath.Join(auditreceipt.StateDir(root), "receipts", corrupt+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// (i) corrupt receipt.
	out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "a1", "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts="+corrupt, false))
	if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseReceiptUnreadable) {
		t.Errorf("corrupt receipt: output = %+v, want a block naming an unreadable receipt", out)
	}

	// (ii) blank final message, (iii) malformed verdict line: both unknown-spec.
	for name, msg := range map[string]string{"blank": "   \n ", "malformed": "AUDIT-VERDICT: PASS spec=nope receipts=none"} {
		tree := newGateTree(t, "required")
		seedStart(t, tree, "a1", auditreceipt.AgentPlanAuditor, start)
		out := runStop(t, stopInput(tree, auditreceipt.AgentPlanAuditor, "a1", msg, false))
		if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseVerdictLineMissing) {
			t.Fatalf("%s: output = %+v, want a block naming a missing verdict line", name, out)
		}
		if _, err := auditreceipt.ReadRejection(tree, auditreceipt.AgentPlanAuditor, auditreceipt.UnknownSpec); err != nil {
			t.Fatalf("%s: unknown-spec rejection missing: %v", name, err)
		}
		isDenied, reason := denied(runSpawn(t, receiptSpawnInput(tree, "manager-develop", "Agent")))
		if !isDenied || !strings.Contains(reason, auditreceipt.UnknownSpec) {
			t.Errorf("%s: spawn denial = (%v, %q), want a deny naming unknown-spec", name, isDenied, reason)
		}
		// Re-entry keeps the record and does not block.
		if out := runStop(t, stopInput(tree, auditreceipt.AgentPlanAuditor, "a1", msg, true)); out.Decision != "" {
			t.Errorf("%s: reentry decision = %q, want none", name, out.Decision)
		}
		if _, err := auditreceipt.ReadRejection(tree, auditreceipt.AgentPlanAuditor, auditreceipt.UnknownSpec); err != nil {
			t.Errorf("%s: reentry dropped the rejection record: %v", name, err)
		}
	}

	// (iv) corrupt rejection record denies and names the file; (v) no rejections dir allows.
	bad := newGateTree(t, "required")
	badPath := filepath.Join(auditreceipt.StateDir(bad), "rejections", "plan-auditor--SPEC-X-001.json")
	if err := os.MkdirAll(filepath.Dir(badPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(badPath, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	isDenied, reason := denied(runSpawn(t, receiptSpawnInput(bad, "manager-develop", "Agent")))
	if !isDenied || !strings.Contains(reason, badPath) {
		t.Errorf("corrupt rejection: (%v, %q), want a deny naming %s", isDenied, reason, badPath)
	}
	empty := newGateTree(t, "required")
	if isDenied, reason := denied(runSpawn(t, receiptSpawnInput(empty, "manager-develop", "Agent"))); isDenied && strings.HasPrefix(reason, "AUDIT_RECEIPT_VIOLATION") {
		t.Errorf("a tree with no rejections directory denied the spawn: %q", reason)
	}

	// (vi) a valid PASS whose verdict line ends in spaces and a CR is accepted.
	ok := newGateTree(t, "required")
	seedStart(t, ok, "a1", auditreceipt.AgentPlanAuditor, start)
	id := seedReceipt(t, ok, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: ok, CreatedAt: start.Add(time.Second)})
	if out := runStop(t, stopInput(ok, auditreceipt.AgentPlanAuditor, "a1", "AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts="+id+" \r", false)); out.Decision != "" {
		t.Errorf("trailing whitespace PASS: decision = %q, want none", out.Decision)
	}
}
