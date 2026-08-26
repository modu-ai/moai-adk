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
