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

// backgroundStartInput builds the SubagentStart payload shape of a background
// Agent() spawn (card t1544): the auditor type and session id are carried,
// the agent id is not.
func backgroundStartInput(root, agentType, sessionID string) *HookInput {
	return &HookInput{
		CWD:           root,
		AgentType:     agentType,
		SessionID:     sessionID,
		HookEventName: string(EventSubagentStart),
	}
}

// Card t1544: a background Agent() spawn delivers SubagentStart without
// agent_id, so the auditor start marker must be recorded under the identity
// the payload does carry (session_id + agent_type). Keyed by agent id alone
// the write was a structural no-op — the measured failure: five spawns, zero
// markers — and every later auditor PASS permanently unprovable.
func TestSubagentStart_BackgroundSpawnWithoutAgentIDRecordsMarker(t *testing.T) {
	for _, agentType := range []string{auditreceipt.AgentPlanAuditor, auditreceipt.AgentSyncAuditor} {
		root := newGateTree(t, "required")
		if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, agentType, "sess-bg-1")); err != nil {
			t.Fatalf("%s: SubagentStart Handle: %v", agentType, err)
		}
		key := auditreceipt.StartMarkerKey("", "sess-bg-1", agentType)
		if key == "" {
			t.Fatalf("%s: StartMarkerKey returned empty for a background spawn", agentType)
		}
		m, err := auditreceipt.ReadStartMarker(root, key)
		if err != nil {
			t.Fatalf("%s: no start marker under the background key %q: %v", agentType, key, err)
		}
		if m.TreeRoot != root || m.StartedAt.IsZero() {
			t.Errorf("%s: marker = %+v, want tree %q and a non-zero start time", agentType, m, root)
		}
	}
}

// Card t1544: the background-spawn audit ceremony must close. An auditor
// spawned without an agent id mints a receipt during its run, cites it in its
// verdict line, and the stop accepts the PASS instead of refusing it with
// "start marker missing" — the refusal that, once persisted, denied every
// phase-entry spawn with no path to resolution (t1509 deadlock).
func TestSubagentStop_BackgroundSpawnAuditorPassIsProvable(t *testing.T) {
	root := newGateTree(t, "required")
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentPlanAuditor, "sess-bg-2")); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
	// A receipt minted after the marker's start time (Now at write), so the
	// citation must qualify once the marker is findable.
	id := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(2 * time.Second),
	})
	stop := &HookInput{
		CWD:                  root,
		AgentType:            auditreceipt.AgentPlanAuditor,
		SessionID:            "sess-bg-2",
		LastAssistantMessage: "AUDIT-VERDICT: PASS spec=SPEC-BG-001 receipts=" + id,
		HookEventName:        string(EventSubagentStop),
	}
	out := runStop(t, stop)
	if out.Decision != "" {
		t.Fatalf("decision = %q, want none — a proven background PASS must not block (reason %q)", out.Decision, out.Reason)
	}
	// The accepted PASS keeps the derived marker: it is a session-era anchor
	// a concurrent same-role instance may still need (card t1544 card-review
	// P2) — only an agent-id-keyed marker is consumed at stop.
	if key := auditreceipt.StartMarkerKey("", "sess-bg-2", auditreceipt.AgentPlanAuditor); key != "" {
		if _, err := auditreceipt.ReadStartMarker(root, key); err != nil {
			t.Errorf("derived start marker %q was consumed by an accepted PASS: %v", key, err)
		}
	}
}

// Card t1544 card-review P2: two concurrent same-role background auditors
// share the derived marker. The first accepted PASS must not destroy the
// second instance's provability, and the second start must not move the
// anchor's start time past receipts the first instance will cite.
func TestSubagentStop_ConcurrentBackgroundAuditorsShareEraAnchor(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-bg-5"
	start := backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session)
	if _, err := NewSubagentStartHandler().Handle(context.Background(), start); err != nil {
		t.Fatalf("first SubagentStart Handle: %v", err)
	}
	key := auditreceipt.StartMarkerKey("", session, auditreceipt.AgentPlanAuditor)
	first, err := auditreceipt.ReadStartMarker(root, key)
	if err != nil {
		t.Fatalf("first start wrote no marker: %v", err)
	}
	// A second same-role spawn of the same session: the anchor keeps the
	// earliest start.
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session)); err != nil {
		t.Fatalf("second SubagentStart Handle: %v", err)
	}
	second, err := auditreceipt.ReadStartMarker(root, key)
	if err != nil {
		t.Fatalf("marker vanished after the second start: %v", err)
	}
	if second.StartedAt != first.StartedAt {
		t.Errorf("anchor StartedAt moved from %v to %v, want keep-earliest", first.StartedAt, second.StartedAt)
	}
	// Each instance proves a receipt minted after the shared anchor began.
	r1 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: second.StartedAt.Add(time.Second)})
	r2 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: second.StartedAt.Add(2 * time.Second)})
	for i, id := range []string{r1, r2} {
		stop := &HookInput{
			CWD:                  root,
			AgentType:            auditreceipt.AgentPlanAuditor,
			SessionID:            session,
			LastAssistantMessage: "AUDIT-VERDICT: PASS spec=SPEC-BG-004 receipts=" + id,
			HookEventName:        string(EventSubagentStop),
		}
		if out := runStop(t, stop); out.Decision != "" {
			t.Fatalf("auditor %d decision = %q, want none (reason %q)", i+1, out.Decision, out.Reason)
		}
	}
}

// Card t1544 card-review P2: an agent-id-carrying auditor's FAIL must not
// delete a concurrent background auditor's derived marker.
func TestSubagentStop_AgentIDFailKeepsBackgroundMarker(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-bg-6"
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentSyncAuditor, session)); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
	key := auditreceipt.StartMarkerKey("", session, auditreceipt.AgentSyncAuditor)
	if _, err := auditreceipt.ReadStartMarker(root, key); err != nil {
		t.Fatalf("background marker missing before the FAIL: %v", err)
	}
	stop := &HookInput{
		CWD:                  root,
		AgentID:              "id-other-1",
		AgentType:            auditreceipt.AgentSyncAuditor,
		SessionID:            session,
		LastAssistantMessage: "AUDIT-VERDICT: FAIL spec=SPEC-BG-005 receipts=none",
		HookEventName:        string(EventSubagentStop),
	}
	if out := runStop(t, stop); out.Decision != "" || out.SystemMessage != "" {
		t.Fatalf("FAIL output = %+v, want silence", out)
	}
	if _, err := auditreceipt.ReadStartMarker(root, key); err != nil {
		t.Errorf("an agent-id FAIL deleted the background marker %q: %v", key, err)
	}
}

// Card t1544: a stop payload that carries an agent id the start payload did
// not must still find the marker the start recorded under the derived key.
func TestSubagentStop_BackgroundStartWithAgentIDStopFindsMarker(t *testing.T) {
	root := newGateTree(t, "required")
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentSyncAuditor, "sess-bg-3")); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
	id := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolAuditMulti,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(2 * time.Second),
	})
	stop := &HookInput{
		CWD:                  root,
		AgentID:              "id-late-1",
		AgentType:            auditreceipt.AgentSyncAuditor,
		SessionID:            "sess-bg-3",
		LastAssistantMessage: "AUDIT-VERDICT: PASS spec=SPEC-BG-002 receipts=" + id,
		HookEventName:        string(EventSubagentStop),
	}
	if out := runStop(t, stop); out.Decision != "" {
		t.Fatalf("decision = %q, want none (reason %q)", out.Decision, out.Reason)
	}
}

// Card t1544: the derived key must not weaken the anti-recycling invariant —
// a receipt minted BEFORE the background auditor began is still refused under
// the session_id+agent_type key.
func TestSubagentStop_BackgroundSpawnRecycledReceiptStillRefused(t *testing.T) {
	root := newGateTree(t, "required")
	recycled := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(-time.Minute),
	})
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentPlanAuditor, "sess-bg-4")); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
	stop := &HookInput{
		CWD:                  root,
		AgentType:            auditreceipt.AgentPlanAuditor,
		SessionID:            "sess-bg-4",
		LastAssistantMessage: "AUDIT-VERDICT: PASS spec=SPEC-BG-003 receipts=" + recycled,
		HookEventName:        string(EventSubagentStop),
	}
	out := runStop(t, stop)
	if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseReceiptBeforeStart) {
		t.Fatalf("output = %+v, want a block naming %q", out, auditreceipt.CauseReceiptBeforeStart)
	}
}

// Card t1544: a payload with neither agent id nor session id still keys no
// marker — no invented identity, no store write.
func TestSubagentStart_NoIdentityNoMarker(t *testing.T) {
	root := newGateTree(t, "required")
	if _, err := NewSubagentStartHandler().Handle(context.Background(), backgroundStartInput(root, auditreceipt.AgentPlanAuditor, "")); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
	if _, err := os.Stat(filepath.Join(auditreceipt.StateDir(root), "starts")); !os.IsNotExist(err) {
		t.Errorf("a marker was written for a payload with no agent id and no session id")
	}
}

// runStart feeds a SubagentStart input to its handler, failing the test on a
// handler error.
func runStart(t *testing.T, input *HookInput) {
	t.Helper()
	if _, err := NewSubagentStartHandler().Handle(context.Background(), input); err != nil {
		t.Fatalf("SubagentStart Handle: %v", err)
	}
}

// bgStopInput builds the SubagentStop payload shape of a background Agent()
// spawn (SPEC-RECEIPT-REUSE-001): session id carried, agent id absent, so the
// marker lookup resolves the derived session-era key.
func bgStopInput(root, agentType, sessionID, message string, reentry bool) *HookInput {
	return &HookInput{
		CWD:                  root,
		AgentType:            agentType,
		SessionID:            sessionID,
		LastAssistantMessage: message,
		StopHookActive:       reentry,
		HookEventName:        string(EventSubagentStop),
	}
}

// freezeClock replaces the store clock with one the test advances explicitly,
// so the relative order of receipts, markers, and end records is deterministic.
// Restores the previous clock on test end (same pattern as
// wsr_audit_receipt_tree_test.go). Not for use with t.Parallel tests.
func freezeClock(t *testing.T) *time.Time {
	t.Helper()
	prev := auditreceipt.Now
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	auditreceipt.Now = func() time.Time { return now }
	t.Cleanup(func() { auditreceipt.Now = prev })
	return &now
}

// advanceClock moves a freezeClock clock forward.
func advanceClock(now *time.Time, d time.Duration) {
	*now = now.Add(d)
}

// endAcceptedPass ends the live instance with an accepted PASS citing r1
// (SPEC-RECEIPT-REUSE-001 AC-RR-001 arm a).
func endAcceptedPass(t *testing.T, now *time.Time, root, session, r1 string) {
	t.Helper()
	advanceClock(now, time.Second)
	msg := "AUDIT-VERDICT: PASS spec=SPEC-RR-000 receipts=" + r1
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, msg, false)); out.Decision != "" {
		t.Fatalf("setup: predecessor accepted-PASS end blocked: %q (%s)", out.Decision, out.Reason)
	}
}

// endReentryRefusal ends the live instance with a first-stop block followed by
// a re-entry refusal (arm b).
func endReentryRefusal(t *testing.T, now *time.Time, root, session, _ string) {
	t.Helper()
	advanceClock(now, time.Second)
	unproven := "AUDIT-VERDICT: PASS spec=SPEC-RR-000 receipts=none"
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, unproven, false)); out.Decision != "block" {
		t.Fatalf("setup: first stop decision = %q, want block", out.Decision)
	}
	advanceClock(now, time.Second)
	out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, unproven, true))
	if out.Decision != "" || !strings.Contains(out.SystemMessage, "not accepted") {
		t.Fatalf("setup: re-entry refusal end output = %+v, want a non-blocking not-accepted message", out)
	}
}

// endFailVerdict ends the live instance with a FAIL verdict (arm c).
func endFailVerdict(t *testing.T, now *time.Time, root, session, _ string) {
	t.Helper()
	advanceClock(now, time.Second)
	msg := "AUDIT-VERDICT: FAIL spec=SPEC-RR-000 receipts=none"
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, msg, false)); out.Decision != "" || out.SystemMessage != "" {
		t.Fatalf("setup: FAIL end output = %+v, want silence", out)
	}
}

// endNoVerdictReentry ends the live instance with a re-entry stop that carries
// no parseable verdict line (arm d).
func endNoVerdictReentry(t *testing.T, now *time.Time, root, session, _ string) {
	t.Helper()
	advanceClock(now, time.Second)
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "   \n ", false)); out.Decision != "block" {
		t.Fatalf("setup: no-verdict first stop = %+v, want block", out)
	}
	advanceClock(now, time.Second)
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "   \n ", true)); out.Decision != "" {
		t.Fatalf("setup: no-verdict re-entry end = %+v, want non-block", out)
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-001/AC-RR-002: the derived background start
// marker is a session-era anchor, so without an instance boundary a SECOND
// sequential auditor of the same session and role proves its PASS with a
// receipt minted during a PREDECESSOR instance's lifetime, with zero audit-tool
// calls. All four predecessor end shapes must seal the era they leave behind,
// and the resulting refusal must carry the reuse-dedicated cause, deny the
// phase-entry spawns, and be clearable by a PASS citing a qualifying receipt.
func TestSubagentStop_SequentialAuditorReceiptReuseIsRefused(t *testing.T) {
	for _, arm := range []struct {
		name string
		end  func(t *testing.T, now *time.Time, root, session, r1 string)
	}{
		{"accepted-pass-end", endAcceptedPass},
		{"reentry-refusal-end", endReentryRefusal},
		{"fail-end", endFailVerdict},
		{"no-verdict-reentry-end", endNoVerdictReentry},
	} {
		t.Run(arm.name, func(t *testing.T) {
			root := newGateTree(t, "required")
			session := "sess-rr-seq"
			now := freezeClock(t)

			// Instance 1 begins and mints r1 during its lifetime.
			runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
			advanceClock(now, time.Second)
			r1 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
			arm.end(t, now, root, session, r1)

			// Instance 2: same session and role, starts after instance 1 has
			// terminally ended, cites r1 without calling the audit tool.
			runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
			advanceClock(now, time.Second)
			msg := "AUDIT-VERDICT: PASS spec=SPEC-RR-001 receipts=" + r1
			out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, msg, false))
			if out.Decision != "block" {
				t.Fatalf("decision = %q, want block — a PASS resting on a predecessor-era receipt must be refused (reason %q)", out.Decision, out.Reason)
			}
			if !strings.Contains(out.Reason, auditreceipt.CauseReceiptReused) {
				t.Errorf("reason = %q, want it to name %q", out.Reason, auditreceipt.CauseReceiptReused)
			}

			// AC-RR-002: the persisted refusal carries the reuse-dedicated
			// cause and denies every phase-entry spawn...
			rj, err := auditreceipt.ReadRejection(root, auditreceipt.AgentPlanAuditor, "SPEC-RR-001")
			if err != nil {
				t.Fatalf("rejection record missing: %v", err)
			}
			if rj.Cause != auditreceipt.CauseReceiptReused {
				t.Errorf("rejection cause = %q, want %q", rj.Cause, auditreceipt.CauseReceiptReused)
			}
			for _, agent := range []string{"manager-develop", "manager-docs", "manager-git"} {
				isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, agent, "Agent")))
				if !isDenied {
					t.Errorf("%s spawn was not denied while the reuse refusal is outstanding", agent)
					continue
				}
				if !strings.HasPrefix(reason, "AUDIT_RECEIPT_VIOLATION") {
					t.Errorf("%s denial reason = %q, want the AUDIT_RECEIPT_VIOLATION prefix", agent, reason)
				}
			}

			// ...and a PASS citing a receipt minted after the boundary clears
			// the refusal and re-opens the spawns.
			advanceClock(now, time.Second)
			r2 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
			out = runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-001 receipts="+r2, false))
			if out.Decision != "" {
				t.Fatalf("clearing PASS decision = %q, want none (reason %q)", out.Decision, out.Reason)
			}
			if isDenied, reason := denied(runSpawn(t, receiptSpawnInput(root, "manager-develop", "Agent"))); isDenied && strings.HasPrefix(reason, "AUDIT_RECEIPT_VIOLATION") {
				t.Errorf("manager-develop spawn still denied after a qualifying PASS: %q", reason)
			}
		})
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-009: the end-event boundary semantics, pinned as
// two named sequences. Counts include the ender. (i) A terminal end arriving
// while exactly one instance is outstanding advances the boundary, so a later
// instance citing a receipt minted before that end is refused. (ii) A terminal
// end arriving while MORE THAN ONE instance is outstanding is ambiguous and
// does NOT advance the boundary — a later instance citing a receipt minted in
// the overlap is still judged against the existing anchor and accepted, the
// concurrency semantics card t1544 pinned.
func TestSubagentStop_ReceiptBoundaryAmbiguitySemantics(t *testing.T) {
	t.Run("single-live-end-advances-boundary", func(t *testing.T) {
		root := newGateTree(t, "required")
		session := "sess-rr-bnd-1"
		now := freezeClock(t)

		// X starts t0, mints r t1, terminally ends t2 as the only instance.
		runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
		advanceClock(now, time.Second)
		r := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
		advanceClock(now, time.Second)
		if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: FAIL spec=SPEC-RR-009 receipts=none", false)); out.Decision != "" {
			t.Fatalf("setup: X's FAIL end output = %+v, want silence", out)
		}

		// Y starts t3 and cites r (t1 < t2): refused, the boundary advanced.
		runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
		advanceClock(now, time.Second)
		out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-009 receipts="+r, false))
		if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseReceiptReused) {
			t.Fatalf("output = %+v, want a block naming %q", out, auditreceipt.CauseReceiptReused)
		}
	})

	t.Run("ambiguous-end-freezes-boundary", func(t *testing.T) {
		root := newGateTree(t, "required")
		session := "sess-rr-bnd-2"
		now := freezeClock(t)

		// A starts t0, B starts t1, B mints r t2, B terminally ends t3 while
		// two instances are outstanding (ender + surviving sibling): ambiguous.
		runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
		advanceClock(now, time.Second)
		runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
		advanceClock(now, time.Second)
		r := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
		advanceClock(now, time.Second)
		if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: FAIL spec=SPEC-RR-009 receipts=none", false)); out.Decision != "" {
			t.Fatalf("setup: B's FAIL end output = %+v, want silence", out)
		}

		// C starts t4 and cites r (t2 > t0): the boundary stayed frozen at the
		// anchor, so the citation qualifies and the PASS is accepted.
		runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
		advanceClock(now, time.Second)
		out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-009 receipts="+r, false))
		if out.Decision != "" {
			t.Fatalf("decision = %q, want none — an ambiguous end must not advance the boundary (reason %q)", out.Decision, out.Reason)
		}
	})
}

// SPEC-RECEIPT-REUSE-001 AC-RR-004 (REQ-RR-005): the first-stop block keeps the
// start marker on purpose — the instance continues, mints a receipt, and the
// receipt it cites must still prove its PASS.
func TestSubagentStop_FirstStopBlockThenOwnReceiptProvable(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-rr-firststop"
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))

	unproven := "AUDIT-VERDICT: PASS spec=SPEC-RR-004 receipts=none"
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, unproven, false)); out.Decision != "block" {
		t.Fatalf("first stop decision = %q, want block", out.Decision)
	}
	// The instance continues after the block and mints a receipt.
	id := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(2 * time.Second),
	})
	out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-004 receipts="+id, false))
	if out.Decision != "" {
		t.Fatalf("decision = %q, want none — the post-block receipt must prove the PASS (reason %q)", out.Decision, out.Reason)
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-006 (REQ-RR-002): a trailing instance that cites
// only receipts it minted during its OWN lifetime — after a predecessor has
// terminally ended — must remain provable. The repair fences predecessor-era
// receipts, never the instance's own.
func TestSubagentStop_SequentialOwnReceiptStillProvable(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-rr-own"
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	id1 := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(2 * time.Second),
	})
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-006 receipts="+id1, false)); out.Decision != "" {
		t.Fatalf("setup: predecessor accepted end blocked: %q (%s)", out.Decision, out.Reason)
	}

	// Instance 2 starts after the predecessor ended and cites only its own
	// receipt.
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	id2 := seedReceipt(t, root, auditreceipt.Receipt{
		Tool:      auditreceipt.ToolCodexAudit,
		TreeRoot:  root,
		CreatedAt: time.Now().UTC().Add(4 * time.Second),
	})
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-006 receipts="+id2, false)); out.Decision != "" {
		t.Fatalf("decision = %q, want none — the instance's own receipt must stay provable (reason %q)", out.Decision, out.Reason)
	}
}

// SPEC-RECEIPT-REUSE-001 (--security --deep review repair, card t1562): a
// ledger that exists but cannot be read makes the end-event boundary
// unknowable — the PASS is REFUSED (fail closed), not waved through with zero
// vision, and the refusal names the condition the operator fixes.
func TestSubagentStop_UnreadableInstanceLedgerFailsClosed(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-rr-ledger"
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	// Corrupt the ledger file directly: the boundary is now unreadable.
	ledgerFile := filepath.Join(auditreceipt.StateDir(root), "ledgers",
		auditreceipt.StartMarkerKey("", session, auditreceipt.AgentPlanAuditor)+".json")
	if err := os.WriteFile(ledgerFile, []byte("{broken"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	msg := "AUDIT-VERDICT: PASS spec=SPEC-RR-002 receipts=rcpt-ffffffffffffffffffff"
	out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, msg, false))
	if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseInstanceLedgerUnreadable) {
		t.Fatalf("output = %+v, want a block naming %q", out, auditreceipt.CauseInstanceLedgerUnreadable)
	}
	rj, err := auditreceipt.ReadRejection(root, auditreceipt.AgentPlanAuditor, "SPEC-RR-002")
	if err != nil {
		t.Fatalf("rejection record missing: %v", err)
	}
	if rj.Cause != auditreceipt.CauseInstanceLedgerUnreadable {
		t.Errorf("rejection cause = %q, want %q", rj.Cause, auditreceipt.CauseInstanceLedgerUnreadable)
	}
}

// Post-sync repair (P2-2, card t1562 mid-flight review) — the gate's
// reproduction end to end: B's start-record write is discarded (the ledger
// lock is held past the start-record wait budget), so the outstanding count is
// short when A terminally ends. Without the repair the boundary advances on
// the incomplete count and B's OWN freshly-minted receipt is refused as
// predecessor-era — a false positive AC-RR-002 forbids. With it, the dropped
// start marks the key uncertain, A's end freezes the boundary, and B's own
// receipt stays provable.
func TestSubagentStop_DroppedStartFreezesBoundary(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-rr-drop"
	now := freezeClock(t)
	key := auditreceipt.StartMarkerKey("", session, auditreceipt.AgentPlanAuditor)
	lockFile := filepath.Join(auditreceipt.StateDir(root), "ledgers", key+".json.lock")

	// A begins and stays live.
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))

	// Hold the ledger lock past the start-record wait budget, so B's
	// start-record write is discarded.
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("hold the ledger lock: %v", err)
	}
	_ = lf.Close()
	// B begins: its marker write proceeds, its ledger count does not.
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	// The dropped start must leave its durable uncertainty mark.
	uncertainFile := filepath.Join(auditreceipt.StateDir(root), "ledgers", key+".json.uncertain")
	if _, err := os.Stat(uncertainFile); err != nil {
		t.Errorf("the dropped start left no uncertainty mark: %v", err)
	}
	// Release the fake hold so later ledger writes succeed.
	if err := os.Remove(lockFile); err != nil {
		t.Fatalf("release the fake lock hold: %v", err)
	}

	// B mints its own receipt during its lifetime; A mints and cites its own.
	advanceClock(now, time.Second)
	rb := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
	advanceClock(now, time.Second)
	rA := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
	advanceClock(now, time.Second)
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-003 receipts="+rA, false)); out.Decision != "" {
		t.Fatalf("setup: A's accepted end blocked: %q (%s)", out.Decision, out.Reason)
	}

	// B ends citing its own receipt: accepted — the uncertain count froze the
	// boundary.
	out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-003 receipts="+rb, false))
	if out.Decision != "" {
		t.Fatalf("decision = %q, want none — B's own receipt must not be refused as predecessor-era on a count known to be incomplete (reason %q)", out.Decision, out.Reason)
	}
}

// Post-sync repair (P2-2 mirror, card t1562 gate round 19): the END-record
// side of the wait budget. A's terminal end is real but its ledger write is
// discarded (the lock is held past the budget) — without the repair the
// boundary never seals, and the NEXT auditor of the session accepts A's
// receipt with zero audit calls. With it, the lost end is marked pending and
// the next ledger operation replays it under the lock, so the boundary seals
// at the end time and the successor's reuse is REFUSED.
func TestSubagentStop_DroppedEndStillSealsBoundary(t *testing.T) {
	root := newGateTree(t, "required")
	session := "sess-rr-dropend"
	now := freezeClock(t)
	key := auditreceipt.StartMarkerKey("", session, auditreceipt.AgentPlanAuditor)
	lockFile := filepath.Join(auditreceipt.StateDir(root), "ledgers", key+".json.lock")

	// A begins, mints rA, and terminally ends — while the ledger lock is held
	// past the end-record budget, so the end-record write is discarded.
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	advanceClock(now, time.Second)
	rA := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: auditreceipt.Now()})
	advanceClock(now, time.Second)
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("hold the ledger lock: %v", err)
	}
	_ = lf.Close()
	if out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-019 receipts="+rA, false)); out.Decision != "" {
		t.Fatalf("setup: A's accepted end blocked: %q (%s)", out.Decision, out.Reason)
	}
	// The dropped end must leave its durable pending mark (one file per
	// instance, named by an id the test cannot predict).
	pendingMarks, _ := filepath.Glob(filepath.Join(auditreceipt.StateDir(root), "ledgers", key+".json.end-pending-*"))
	if len(pendingMarks) == 0 {
		t.Errorf("the dropped end left no pending mark")
	}
	if err := os.Remove(lockFile); err != nil {
		t.Fatalf("release the fake lock hold: %v", err)
	}

	// B begins after A's end and cites A's receipt with zero audit calls.
	runStart(t, backgroundStartInput(root, auditreceipt.AgentPlanAuditor, session))
	advanceClock(now, time.Second)
	out := runStop(t, bgStopInput(root, auditreceipt.AgentPlanAuditor, session, "AUDIT-VERDICT: PASS spec=SPEC-RR-019 receipts="+rA, false))
	if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseReceiptReused) {
		t.Fatalf("output = %+v, want a block naming %q — the replayed end must seal the boundary against the successor's reuse", out, auditreceipt.CauseReceiptReused)
	}
}

// SPEC-RECEIPT-REUSE-001 AC-RR-007 (REQ-RR-002): the foreground path —
// agent-id-keyed markers, consumed at the instance's own stop — is unchanged:
// a foreground successor citing a predecessor foreground instance's receipt is
// refused with the before-start cause, and citing its own receipt is accepted.
func TestSubagentStop_ForegroundSequentialUnchanged(t *testing.T) {
	root := newGateTree(t, "required")
	start := time.Now().UTC()
	seedStart(t, root, "fg-1", auditreceipt.AgentPlanAuditor, start)
	r1 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(time.Second)})
	if out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "fg-1", "AUDIT-VERDICT: PASS spec=SPEC-RR-007 receipts="+r1, false)); out.Decision != "" {
		t.Fatalf("setup: foreground predecessor accepted end blocked: %q (%s)", out.Decision, out.Reason)
	}

	// Foreground instance 2 owns its own marker: instance 1's receipt predates
	// it and is refused with the existing before-start cause.
	seedStart(t, root, "fg-2", auditreceipt.AgentPlanAuditor, start.Add(2*time.Second))
	out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "fg-2", "AUDIT-VERDICT: PASS spec=SPEC-RR-007 receipts="+r1, false))
	if out.Decision != "block" || !strings.Contains(out.Reason, auditreceipt.CauseReceiptBeforeStart) {
		t.Fatalf("output = %+v, want a block naming %q", out, auditreceipt.CauseReceiptBeforeStart)
	}

	// Its own receipt is accepted.
	r2 := seedReceipt(t, root, auditreceipt.Receipt{Tool: auditreceipt.ToolCodexAudit, TreeRoot: root, CreatedAt: start.Add(3 * time.Second)})
	if out := runStop(t, stopInput(root, auditreceipt.AgentPlanAuditor, "fg-2", "AUDIT-VERDICT: PASS spec=SPEC-RR-007 receipts="+r2, false)); out.Decision != "" {
		t.Fatalf("decision = %q, want none (reason %q)", out.Decision, out.Reason)
	}
}
