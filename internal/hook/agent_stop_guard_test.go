package hook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// stopGuardTestHelpers — fixture builders for the agent-stop guard (M1).
//
// The fixtures are synthetic per acceptance.md §A.4: unit ACs run on
// synthetic HookInput fixtures with t.TempDir() registries; the live wire
// format is verified separately by the E-P1 recipe.

// taskStopInput builds a PostToolUse HookInput for a TaskStop completion.
func taskStopInput(sessionID, payload string) *HookInput {
	return &HookInput{
		HookEventName: "PostToolUse",
		ToolName:      "TaskStop",
		SessionID:     sessionID,
		ToolInput:     json.RawMessage(payload),
	}
}

// sendMessageInput builds a PreToolUse HookInput for a SendMessage issuance.
func sendMessageInput(sessionID, to string) *HookInput {
	payload := map[string]any{"to": to, "message": "coordination note"}
	if to == "" {
		payload = map[string]any{"message": "no recipient"}
	}
	raw, _ := json.Marshal(payload)
	return &HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "SendMessage",
		SessionID:     sessionID,
		ToolInput:     raw,
	}
}

// newStopGuardTestHandler builds a preToolHandler with the project root pinned
// to root (the M1 observer consults no gate — observation is unconditional).
func newStopGuardTestHandler(t *testing.T, root string) *preToolHandler {
	t.Helper()
	cfg := config.NewDefaultConfig()
	return &preToolHandler{
		cfg:        &auditConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: root,
	}
}

// readAgentStopRegistryFile loads the per-session stop registry from root.
func readAgentStopRegistryFile(t *testing.T, root, sessionID string) *AgentStopRegistry {
	t.Helper()
	data, err := os.ReadFile(agentStopRegistryPath(root, sessionID))
	if err != nil {
		t.Fatalf("read stop registry: %v", err)
	}
	var reg AgentStopRegistry
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatalf("unmarshal stop registry: %v", err)
	}
	return &reg
}

// readAgentStopAuditRows loads every audit row appended under root.
func readAgentStopAuditRows(t *testing.T, root string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "logs", agentStopAuditFileName))
	if err != nil {
		t.Fatalf("read stop-guard audit log: %v", err)
	}
	var rows []map[string]any
	for _, ln := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if ln == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(ln), &rec); err != nil {
			t.Fatalf("audit line is not JSON: %v (%s)", err, ln)
		}
		rows = append(rows, rec)
	}
	return rows
}

// TestRecordAgentStopRecordsRegistryAndAudit pins AC-TRG-001: a TaskStop
// completion against a named teammate persists a stop record AND appends
// exactly one stop_recorded audit row.
func TestRecordAgentStopRecordsRegistryAndAudit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))

	reg := readAgentStopRegistryFile(t, root, "sess-1")
	if reg.SessionID != "sess-1" {
		t.Errorf("registry session_id: got %q, want sess-1", reg.SessionID)
	}
	if len(reg.Entries) != 1 {
		t.Fatalf("registry entries: got %d, want 1", len(reg.Entries))
	}
	e := reg.Entries[0]
	if e.Name != "worker-a" {
		t.Errorf("entry name: got %q, want worker-a", e.Name)
	}
	if e.AgentID != "a1" {
		t.Errorf("entry agent_id: got %q, want a1", e.AgentID)
	}
	if e.StoppedAt == "" {
		t.Errorf("entry stopped_at: got empty, want a UTC RFC3339 timestamp")
	}

	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 1 {
		t.Fatalf("audit rows: got %d, want exactly 1", len(rows))
	}
	if rows[0]["kind"] != AgentStopKindStopRecorded {
		t.Errorf("audit kind: got %v, want %q", rows[0]["kind"], AgentStopKindStopRecorded)
	}
	if rows[0]["name"] != "worker-a" {
		t.Errorf("audit name: got %v, want worker-a", rows[0]["name"])
	}
}

// TestRecordAgentStopIdempotentUpsert pins AC-TRG-001/AC-TRG-008: repeated
// stops of the same name refresh stopped_at without duplicating entries, and
// each observed completion appends its own audit row.
func TestRecordAgentStopIdempotentUpsert(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))

	reg := readAgentStopRegistryFile(t, root, "sess-1")
	if len(reg.Entries) != 1 {
		t.Fatalf("registry entries after repeat stop: got %d, want 1 (idempotent upsert)", len(reg.Entries))
	}
	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 2 {
		t.Fatalf("audit rows: got %d, want 2 (one per observed completion)", len(rows))
	}
}

// TestRecordAgentStopSeparateSessions pins the per-session scoping design
// anchor: two sessions produce two registry files, never a shared list.
func TestRecordAgentStopSeparateSessions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
	RecordAgentStop(root, taskStopInput("sess-2", `{"name":"worker-a"}`))

	for _, s := range []string{"sess-1", "sess-2"} {
		reg := readAgentStopRegistryFile(t, root, s)
		if len(reg.Entries) != 1 || reg.Entries[0].Name != "worker-a" {
			t.Errorf("session %s: registry = %+v, want exactly one worker-a entry", s, reg.Entries)
		}
	}
}

// TestRecordAgentStopFailOpen pins the recorder's uncertainty paths: an
// unparseable payload, an absent target, or an empty session id produce no
// artifacts and never panic.
func TestRecordAgentStopFailOpen(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	t.Run("unparseable tool_input", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop(root, taskStopInput("sess-1", `not json`))
	})
	t.Run("absent target fields", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop(root, taskStopInput("sess-1", `{"reason":"cleanup"}`))
	})
	t.Run("empty tool_input", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop(root, taskStopInput("sess-1", ``))
	})
	t.Run("empty session id", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop(root, taskStopInput("", `{"name":"worker-a"}`))
	})
	t.Run("unresolved project root", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop("", taskStopInput("sess-1", `{"name":"worker-a"}`))
	})
	t.Run("nil input", func(t *testing.T) {
		t.Parallel()
		RecordAgentStop(root, nil)
	})
}

// TestRecordAgentStopUnwritableRoot pins the fail-open write path: a blocked
// state directory never panics.
func TestRecordAgentStopUnwritableRoot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Make <root>/.moai a regular file so the agent-stops MkdirAll fails.
	if err := os.WriteFile(filepath.Join(root, ".moai"), []byte("x"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
}

// TestExtractTaskStopTarget pins the defensive TaskStop payload parser across
// the plausible wire shapes (snake/camel, name/id-only variants).
func TestExtractTaskStopTarget(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		payload  string
		wantName string
		wantID   string
		wantOK   bool
	}{
		{"name and agent_id", `{"name":"worker-a","agent_id":"a1"}`, "worker-a", "a1", true},
		{"name only", `{"name":"worker-a"}`, "worker-a", "", true},
		{"agent_id only", `{"agent_id":"a1"}`, "", "a1", true},
		{"camelCase agentId", `{"agentId":"a1"}`, "", "a1", true},
		{"agent_name variant", `{"agent_name":"worker-a"}`, "worker-a", "", true},
		{"task_id variant", `{"task_id":"t9"}`, "", "t9", true},
		{"camelCase taskId", `{"taskId":"t9"}`, "", "t9", true},
		{"no target fields", `{"reason":"x"}`, "", "", false},
		{"unparseable", `not json`, "", "", false},
		{"empty", ``, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			name, id, ok := extractTaskStopTarget(json.RawMessage(tc.payload))
			if ok != tc.wantOK {
				t.Errorf("ok: got %v, want %v", ok, tc.wantOK)
			}
			if name != tc.wantName || id != tc.wantID {
				t.Errorf("target: got (%q,%q), want (%q,%q)", name, id, tc.wantName, tc.wantID)
			}
		})
	}
}

// TestParseSendRecipient pins the SendMessage recipient parser including the
// sanctioned "name [ref]" disambiguated addressing form (parse-and-strip
// before any registry comparison).
func TestParseSendRecipient(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload string
		want    string
		wantOK  bool
	}{
		{"bare name", `{"to":"worker-a"}`, "worker-a", true},
		{"name [ref] stripped", `{"to":"worker-a [3fa9c1]"}`, "worker-a", true},
		{"main", `{"to":"main"}`, "main", true},
		{"missing to", `{"message":"hi"}`, "", false},
		{"empty to", `{"to":""}`, "", false},
		{"unparseable", `not json`, "", false},
		{"empty payload", ``, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseSendRecipient(json.RawMessage(tc.payload))
			if ok != tc.wantOK {
				t.Errorf("ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("recipient: got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestLookupAgentStopEntry pins the registry lookup: recipient matches by
// name or agent id, per-session only, and miss/absent-registry fail open.
func TestLookupAgentStopEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))

	cases := []struct {
		name      string
		session   string
		recipient string
		wantFound bool
	}{
		{"match by name", "sess-1", "worker-a", true},
		{"match by agent id", "sess-1", "a1", true},
		{"live teammate name misses", "sess-1", "worker-b", false},
		{"other session misses", "sess-2", "worker-a", false},
		{"absent registry misses", "sess-9", "worker-a", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entry, found := lookupAgentStopEntry(root, tc.session, tc.recipient)
			if found != tc.wantFound {
				t.Fatalf("found: got %v, want %v", found, tc.wantFound)
			}
			if found && entry.Name != "worker-a" {
				t.Errorf("matched entry name: got %q, want worker-a", entry.Name)
			}
		})
	}
}

// TestCheckStopGuardObserveOnly pins the M1 observer: every SendMessage gets
// exactly one send_observed row; a registry hit surfaces a non-blocking
// advisory naming the stopped teammate; no deny decision exists in M1.
func TestCheckStopGuardObserveOnly(t *testing.T) {
	t.Parallel()

	t.Run("send to stopped name: advisory, no deny", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
		h := newStopGuardTestHandler(t, root)

		decision, reason, advisory := h.checkStopGuard(sendMessageInput("sess-1", "worker-a"))
		if decision == DecisionDeny {
			t.Errorf("decision: got deny in M1 observe-only observer, want allow (reason=%s)", reason)
		}
		if advisory == "" {
			t.Fatalf("advisory: got empty, want a non-blocking advisory naming the stopped teammate")
		}
		if !strings.Contains(advisory, "worker-a") {
			t.Errorf("advisory omits the stopped teammate name: %s", advisory)
		}
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 2 {
			t.Fatalf("audit rows: got %d, want 2 (stop_recorded + send_observed)", len(rows))
		}
		if rows[1]["kind"] != AgentStopKindSendObserved {
			t.Errorf("audit kind: got %v, want %q", rows[1]["kind"], AgentStopKindSendObserved)
		}
		if rows[1]["name"] != "worker-a" {
			t.Errorf("audit name: got %v, want worker-a", rows[1]["name"])
		}
	})

	t.Run("send to live teammate: observed, no advisory", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
		h := newStopGuardTestHandler(t, root)

		decision, _, advisory := h.checkStopGuard(sendMessageInput("sess-1", "worker-b"))
		if decision == DecisionDeny {
			t.Errorf("decision: got deny for a live teammate, want allow")
		}
		if advisory != "" {
			t.Errorf("advisory: got %q for a live teammate, want empty", advisory)
		}
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 2 || rows[1]["kind"] != AgentStopKindSendObserved || rows[1]["name"] != "worker-b" {
			t.Errorf("audit rows = %+v, want a send_observed row for worker-b", rows)
		}
	})

	t.Run("malformed payload: observed with empty name, no panic", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		h := newStopGuardTestHandler(t, root)

		in := &HookInput{HookEventName: "PreToolUse", ToolName: "SendMessage", SessionID: "sess-1", ToolInput: json.RawMessage(`not json`)}
		decision, _, _ := h.checkStopGuard(in)
		if decision == DecisionDeny {
			t.Errorf("decision: got deny for a malformed payload, want allow (fail-open)")
		}
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 1 || rows[0]["kind"] != AgentStopKindSendObserved {
			t.Errorf("audit rows = %+v, want exactly one send_observed row for the uncertain event", rows)
		}
	})

	t.Run("unresolved project root: no panic", func(t *testing.T) {
		t.Parallel()
		h := newStopGuardTestHandler(t, "")
		RecordAgentStop(t.TempDir(), taskStopInput("sess-1", `{"name":"worker-a"}`))
		h.checkStopGuard(sendMessageInput("sess-1", "worker-a"))
	})
}

// TestAgentStopAuditRowShape pins AC-TRG-008 for the M1 event kinds: every
// row carries all six required fields with a UTC RFC3339 timestamp.
func TestAgentStopAuditRowShape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	h := newStopGuardTestHandler(t, root)
	h.checkStopGuard(sendMessageInput("sess-1", "worker-b"))

	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 2 {
		t.Fatalf("audit rows: got %d, want 2", len(rows))
	}
	for _, rec := range rows {
		for _, field := range []string{"timestamp", "session_id", "kind", "name", "agent_id", "decision"} {
			if _, ok := rec[field]; !ok {
				t.Errorf("audit row missing required field %q: %v", field, rec)
			}
		}
		if ts, _ := rec["timestamp"].(string); ts == "" {
			t.Errorf("audit row timestamp empty: %v", rec)
		} else if !strings.Contains(ts, "T") || !strings.HasSuffix(ts, "Z") {
			t.Errorf("audit timestamp %q is not UTC RFC3339", ts)
		}
	}
}

// TestPreToolSendMessageObserveOnlyViaHandle pins the Handle() wiring: a
// SendMessage PreToolUse input reaches the stop-guard observer, is never
// denied in M1, and the registry-hit advisory rides SystemMessage.
func TestPreToolSendMessageObserveOnlyViaHandle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	h := newStopGuardTestHandler(t, root)

	out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertNotDeny(t, out)
	if !strings.Contains(out.SystemMessage, "worker-a") {
		t.Errorf("SystemMessage: got %q, want the stop-guard advisory naming worker-a", out.SystemMessage)
	}
	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 2 || rows[1]["kind"] != AgentStopKindSendObserved {
		t.Errorf("audit rows = %+v, want stop_recorded + send_observed", rows)
	}
}

// TestPreToolSendMessageLiveTeammateUnblockedViaHandle pins the mirror-image
// direction at the wiring level: a send to a live teammate carries no
// advisory and no deny.
func TestPreToolSendMessageLiveTeammateUnblockedViaHandle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
	h := newStopGuardTestHandler(t, root)

	out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-b"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertNotDeny(t, out)
	if out.SystemMessage != "" {
		t.Errorf("SystemMessage: got %q, want empty for a live teammate", out.SystemMessage)
	}
}

// ---------------------------------------------------------------------------
// M2 — enforcement layer
// ---------------------------------------------------------------------------

// newStopGuardGatedHandler builds a preToolHandler with the enforcement gate
// set to the requested state (workflow.agent_stop_guard.enabled).
func newStopGuardGatedHandler(t *testing.T, root string, gateEnabled bool) *preToolHandler {
	t.Helper()
	cfg := config.NewDefaultConfig()
	cfg.Workflow.AgentStopGuard.Enabled = gateEnabled
	return &preToolHandler{
		cfg:        &auditConfigProvider{cfg: cfg},
		policy:     DefaultSecurityPolicy(),
		projectDir: root,
	}
}

// spawnInput builds a PreToolUse HookInput for an Agent/Task spawn carrying a
// teammate name.
func spawnInput(sessionID, name string) *HookInput {
	raw, _ := json.Marshal(map[string]any{
		"subagent_type": "general-purpose",
		"description":   "revival spawn",
		"prompt":        "p",
		"name":          name,
	})
	return &HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "Agent",
		SessionID:     sessionID,
		ToolInput:     raw,
	}
}

// sendMessageInputWithExtras builds a SendMessage input carrying extra fields
// (notify_when_idle) alongside the recipient.
func sendMessageInputWithExtras(t *testing.T, sessionID, to string, extra map[string]any) *HookInput {
	t.Helper()
	payload := map[string]any{"to": to, "message": "coordination note"}
	for k, v := range extra {
		payload[k] = v
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return &HookInput{
		HookEventName: "PreToolUse",
		ToolName:      "SendMessage",
		SessionID:     sessionID,
		ToolInput:     raw,
	}
}

// TestAgentStopGuardDenySentinelPath pins AC-TRG-002: gate enabled + recipient
// matching a live registry entry denies with the sentinel prefix, names the
// teammate, names the orchestrator route, and appends a send_denied row.
func TestAgentStopGuardDenySentinelPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	h := newStopGuardGatedHandler(t, root, true)

	out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertDeny(t, out)
	reason := reasonOf(out)
	if !strings.HasPrefix(reason, SentinelAgentStopViolation+":") {
		t.Errorf("deny reason lacks the %q sentinel prefix: %s", SentinelAgentStopViolation, reason)
	}
	if !strings.Contains(reason, "worker-a") {
		t.Errorf("deny reason omits the stopped teammate name: %s", reason)
	}
	if !strings.Contains(reason, "orchestrator") {
		t.Errorf("deny reason omits the orchestrator route: %s", reason)
	}

	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 2 {
		t.Fatalf("audit rows: got %d, want 2", len(rows))
	}
	if rows[1]["kind"] != AgentStopKindSendDenied {
		t.Errorf("audit kind: got %v, want %q", rows[1]["kind"], AgentStopKindSendDenied)
	}
	if rows[1]["decision"] != "denied" {
		t.Errorf("audit decision: got %v, want denied", rows[1]["decision"])
	}
}

// TestAgentStopGuardGateOffNeverDenies pins AC-TRG-007: with the resolved
// shipped default (enabled false) a registry hit surfaces an advisory, never a
// deny, and the audit row records the would-have-denied state via the
// advisory flag.
func TestAgentStopGuardGateOffNeverDenies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
	h := newStopGuardGatedHandler(t, root, false)

	out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertNotDeny(t, out)
	if !strings.Contains(out.SystemMessage, "worker-a") {
		t.Errorf("SystemMessage: got %q, want the advisory naming worker-a", out.SystemMessage)
	}

	rows := readAgentStopAuditRows(t, root)
	if len(rows) != 2 {
		t.Fatalf("audit rows: got %d, want 2", len(rows))
	}
	if rows[1]["kind"] != AgentStopKindSendObserved {
		t.Errorf("audit kind: got %v, want %q (gate off observes)", rows[1]["kind"], AgentStopKindSendObserved)
	}
	if adv, _ := rows[1]["advisory"].(bool); !adv {
		t.Errorf("audit advisory flag: got %v, want true (would-have-denied state)", rows[1]["advisory"])
	}
}

// TestAgentStopGuardUncertainNeverDenies pins AC-TRG-003: malformed JSON,
// missing recipient, and an unreadable registry file all allow under the
// enabled gate, each with an observe-only audit row.
func TestAgentStopGuardUncertainNeverDenies(t *testing.T) {
	t.Parallel()

	t.Run("malformed JSON tool_input", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
		h := newStopGuardGatedHandler(t, root, true)
		in := &HookInput{HookEventName: "PreToolUse", ToolName: "SendMessage", SessionID: "sess-1", ToolInput: json.RawMessage(`not json`)}
		out, err := h.Handle(context.Background(), in)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("missing to field", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
		h := newStopGuardGatedHandler(t, root, true)
		in := &HookInput{HookEventName: "PreToolUse", ToolName: "SendMessage", SessionID: "sess-1", ToolInput: json.RawMessage(`{"message":"hi"}`)}
		out, err := h.Handle(context.Background(), in)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})

	t.Run("unreadable registry file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".moai", "state", "agent-stops"), 0o750); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(agentStopRegistryPath(root, "sess-1"), []byte(`{corrupt`), 0o600); err != nil {
			t.Fatalf("seed corrupt registry: %v", err)
		}
		h := newStopGuardGatedHandler(t, root, true)
		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 1 || rows[0]["kind"] != AgentStopKindSendObserved {
			t.Errorf("audit rows = %+v, want exactly one observe-only row", rows)
		}
	})
}

// TestAgentStopGuardLiveTeammatesUnaffected pins AC-TRG-004 (the mirror
// mutant): sends to a live name Y, to main, and a notify_when_idle send to Y
// all allow under the enabled gate with send_observed rows.
func TestAgentStopGuardLiveTeammatesUnaffected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))

	cases := []struct {
		name  string
		input func(t *testing.T) *HookInput
	}{
		{"live teammate Y", func(t *testing.T) *HookInput {
			return sendMessageInput("sess-1", "worker-b")
		}},
		{"main", func(t *testing.T) *HookInput {
			return sendMessageInput("sess-1", "main")
		}},
		{"notify_when_idle to Y", func(t *testing.T) *HookInput {
			return sendMessageInputWithExtras(t, "sess-1", "worker-b", map[string]any{"notify_when_idle": true})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newStopGuardGatedHandler(t, root, true)
			out, err := h.Handle(context.Background(), tc.input(t))
			if err != nil {
				t.Fatalf("Handle: %v", err)
			}
			assertNotDeny(t, out)
		})
	}
}

// TestAgentStopGuardNameRefBothDirections pins AC-TRG-011: the sanctioned
// "name [ref]" form is parsed and stripped before registry comparison — a
// stopped "X [ref]" DENIES, a live "Y [ref]" ALLOWS with a send_observed row.
func TestAgentStopGuardNameRefBothDirections(t *testing.T) {
	t.Parallel()

	t.Run("stopped X [ref] denied", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
		h := newStopGuardGatedHandler(t, root, true)

		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a [3fa9c1]"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertDeny(t, out)
		if !strings.HasPrefix(reasonOf(out), SentinelAgentStopViolation+":") {
			t.Errorf("deny reason lacks sentinel prefix: %s", reasonOf(out))
		}
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 2 || rows[1]["kind"] != AgentStopKindSendDenied {
			t.Errorf("audit rows = %+v, want stop_recorded + send_denied", rows)
		}
	})

	t.Run("live Y [ref] allowed", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
		h := newStopGuardGatedHandler(t, root, true)

		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-b [9c2d44]"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
		rows := readAgentStopAuditRows(t, root)
		if len(rows) != 2 || rows[1]["kind"] != AgentStopKindSendObserved || rows[1]["name"] != "worker-b" {
			t.Errorf("audit rows = %+v, want a send_observed row for worker-b", rows)
		}
	})
}

// TestAgentStopGuardAgentIdRecipientDenied pins the agent-id addressing form:
// a recipient equal to a registry entry's agent_id denies under the gate.
func TestAgentStopGuardAgentIdRecipientDenied(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	h := newStopGuardGatedHandler(t, root, true)

	out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "a1"))
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	assertDeny(t, out)
}

// TestSpawnNameClearsStoppedEntry pins AC-TRG-005: a fresh Agent spawn
// carrying the stopped name clears the registry entry BEFORE the spawn
// proceeds (respawn_cleared row), and a subsequent send to the name allows.
func TestSpawnNameClearsStoppedEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a","agent_id":"a1"}`))
	h := newStopGuardGatedHandler(t, root, true)

	out, err := h.Handle(context.Background(), spawnInput("sess-1", "worker-a"))
	if err != nil {
		t.Fatalf("Handle spawn: %v", err)
	}
	assertNotDeny(t, out)

	// Registry entry gone (deliberate revival escape hatch).
	reg, err := os.ReadFile(agentStopRegistryPath(root, "sess-1"))
	if err == nil && strings.Contains(string(reg), "worker-a") {
		t.Errorf("registry still holds worker-a after same-name spawn: %s", reg)
	}

	rows := readAgentStopAuditRows(t, root)
	foundCleared := false
	for _, r := range rows {
		if r["kind"] == AgentStopKindRespawnCleared && r["name"] == "worker-a" {
			foundCleared = true
		}
	}
	if !foundCleared {
		t.Errorf("audit rows = %+v, want a respawn_cleared row for worker-a", rows)
	}

	// Subsequent send to the respawned name allows.
	out2, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
	if err != nil {
		t.Fatalf("Handle send-after-respawn: %v", err)
	}
	assertNotDeny(t, out2)
}

// TestSpawnOtherNameKeepsEntry pins the complement: a spawn whose name does
// not match a stopped entry clears nothing.
func TestSpawnOtherNameKeepsEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
	h := newStopGuardGatedHandler(t, root, true)

	if _, err := h.Handle(context.Background(), spawnInput("sess-1", "worker-z")); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	reg := readAgentStopRegistryFile(t, root, "sess-1")
	if len(reg.Entries) != 1 || reg.Entries[0].Name != "worker-a" {
		t.Errorf("registry = %+v, want worker-a intact after unrelated spawn", reg.Entries)
	}
}

// TestExtractAgentSpawnName pins the spawn-payload Name read used by the
// respawn-clear path.
func TestExtractAgentSpawnName(t *testing.T) {
	t.Parallel()

	sp, ok := extractAgentSpawn(json.RawMessage(`{"subagent_type":"general-purpose","name":"worker-a"}`))
	if !ok || sp.Name != "worker-a" {
		t.Errorf("got (%+v, ok=%v), want Name worker-a ok=true", sp, ok)
	}
	sp, ok = extractAgentSpawn(json.RawMessage(`{"subagent_type":"general-purpose"}`))
	if !ok || sp.Name != "" {
		t.Errorf("got (%+v, ok=%v), want empty Name ok=true", sp, ok)
	}
}

// TestSessionEndClearsStops pins AC-TRG-006: the session-end path removes the
// session's registry file and appends a session_cleared audit row.
func TestSessionEndClearsStops(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))
	RecordAgentStop(root, taskStopInput("sess-2", `{"name":"worker-b"}`))

	h := NewSessionEndHandler()
	out, err := h.Handle(context.Background(), &HookInput{
		HookEventName: "SessionEnd",
		SessionID:     "sess-1",
		CWD:           root,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out == nil {
		t.Fatalf("nil output")
	}

	if _, err := os.Stat(agentStopRegistryPath(root, "sess-1")); !os.IsNotExist(err) {
		t.Errorf("sess-1 registry file still exists after session end")
	}
	// Other sessions unaffected (per-session scoping).
	if _, err := os.Stat(agentStopRegistryPath(root, "sess-2")); err != nil {
		t.Errorf("sess-2 registry must survive sess-1's end: %v", err)
	}

	rows := readAgentStopAuditRows(t, root)
	found := false
	for _, r := range rows {
		if r["kind"] == AgentStopKindSessionCleared && r["session_id"] == "sess-1" {
			found = true
		}
	}
	if !found {
		t.Errorf("audit rows = %+v, want a session_cleared row for sess-1", rows)
	}
}

// TestAgentStopGuardNilConfigNeverDenies pins the fail-open gate read: a nil
// ConfigProvider or nil Config can never reach the deny path.
func TestAgentStopGuardNilConfigNeverDenies(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	RecordAgentStop(root, taskStopInput("sess-1", `{"name":"worker-a"}`))

	t.Run("nil ConfigProvider", func(t *testing.T) {
		t.Parallel()
		h := &preToolHandler{cfg: nil, policy: DefaultSecurityPolicy(), projectDir: root}
		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})
	t.Run("nil Config", func(t *testing.T) {
		t.Parallel()
		h := &preToolHandler{cfg: &auditConfigProvider{cfg: nil}, policy: DefaultSecurityPolicy(), projectDir: root}
		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "worker-a"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertNotDeny(t, out)
	})
}

// TestExtractTaskStopTargetLiveShape pins the LIVE-measured TaskStop payload
// shape (E-P1 step 1-2, fresh headless session, orchestrator-run probe): the
// display-name field arrives EMPTY and the spawn name travels in the field
// the parser maps to agent_id. Registry entry verbatim:
// {"name":"","agent_id":"<spawn-name>"}. This fixture is the ground truth the
// parser and the both-fields matcher must not regress against; the defensive
// snake/camel variants above supplement, not replace, it.
func TestExtractTaskStopTargetLiveShape(t *testing.T) {
	t.Parallel()

	// Verbatim live shape: empty display name + spawn name in the id field.
	name, agentID, ok := extractTaskStopTarget(json.RawMessage(`{"name":"","agent_id":"t267-ep1-probe"}`))
	if !ok || name != "" || agentID != "t267-ep1-probe" {
		t.Errorf("live shape: got (name=%q, agent_id=%q, ok=%v), want (name=\"\", agent_id=\"t267-ep1-probe\", ok=true)", name, agentID, ok)
	}
}

// TestAgentStopGuardLiveShapeLifecycle pins the end-to-end consequences of
// the live-measured registry shape under the enabled gate: a bare-name send
// to the spawn name DENIES (agent_id match), the [ref]-stripped base DENIES,
// and a same-name spawn CLEARS the entry — the t232 vector was name-addressed
// and the live shape puts that name in agent_id. Each subtest seeds its OWN
// root: the clear subtest mutates the registry, so a shared root would race
// the two deny subtests' lookups.
func TestAgentStopGuardLiveShapeLifecycle(t *testing.T) {
	t.Parallel()

	recordLiveShape := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		RecordAgentStop(root, taskStopInput("sess-1", `{"name":"","agent_id":"t267-ep1-probe"}`))
		return root
	}

	t.Run("bare spawn-name send denied via agent_id match", func(t *testing.T) {
		t.Parallel()
		h := newStopGuardGatedHandler(t, recordLiveShape(t), true)
		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "t267-ep1-probe"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertDeny(t, out)
		if !strings.HasPrefix(reasonOf(out), SentinelAgentStopViolation+":") {
			t.Errorf("deny reason lacks sentinel prefix: %s", reasonOf(out))
		}
	})

	t.Run("[ref]-stripped spawn-name send denied", func(t *testing.T) {
		t.Parallel()
		h := newStopGuardGatedHandler(t, recordLiveShape(t), true)
		out, err := h.Handle(context.Background(), sendMessageInput("sess-1", "t267-ep1-probe [3fa9c1]"))
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		assertDeny(t, out)
	})

	t.Run("same-name spawn clears the agent_id-keyed entry", func(t *testing.T) {
		t.Parallel()
		root := recordLiveShape(t)
		h := newStopGuardGatedHandler(t, root, true)
		out, err := h.Handle(context.Background(), spawnInput("sess-1", "t267-ep1-probe"))
		if err != nil {
			t.Fatalf("Handle spawn: %v", err)
		}
		assertNotDeny(t, out)
		if _, err := os.Stat(agentStopRegistryPath(root, "sess-1")); !os.IsNotExist(err) {
			t.Errorf("live-shape entry survived the same-name spawn (registry file still exists)")
		}
	})
}
