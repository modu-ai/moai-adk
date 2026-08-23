package sessionmsg

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestEnvelopeA2AAlignment verifies AC-CSM-001: the envelope JSON serialization
// carries the six A2A Message core keys (messageId, contextId, taskId, role,
// parts, metadata) as camelCase JSON names plus the Part kind discriminator,
// and that envelope validation enforces the A2A-aligned constraints
// (REQ-CSM-002): only text|data parts, non-empty messageId, valid role,
// bounded parts count, and bounded text size.
func TestEnvelopeA2AAlignment(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)

	t.Run("core keys are camelCase A2A names", func(t *testing.T) {
		env := Envelope{
			Message: Message{
				MessageID: "msg-0001",
				ContextID: "ctx-9",
				TaskID:    "task-3",
				Role:      RoleAgent,
				Parts: []Part{
					{Kind: PartKindText, Text: "hello"},
				},
				Metadata: map[string]any{"origin": "test"},
			},
			Delivery: Delivery{
				SenderID:   "claude-abcd1234",
				SenderKind: KindClaude,
				SentAt:     now,
				ExpiresAt:  now.Add(24 * time.Hour),
			},
		}
		data, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("marshal envelope: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		msg, ok := raw["message"].(map[string]any)
		if !ok {
			t.Fatalf("envelope JSON has no message object: %s", data)
		}
		for _, key := range []string{"messageId", "contextId", "taskId", "role", "parts", "metadata"} {
			if _, present := msg[key]; !present {
				t.Errorf("message JSON missing A2A core key %q (camelCase): %s", key, data)
			}
		}
		// Snake_case drift must NOT appear (naming-drift guard, design.md §3.1 R2).
		for _, snake := range []string{"message_id", "context_id", "task_id"} {
			if _, present := msg[snake]; present {
				t.Errorf("message JSON carries snake_case key %q — A2A alignment requires camelCase", snake)
			}
		}
		parts, ok := msg["parts"].([]any)
		if !ok || len(parts) != 1 {
			t.Fatalf("message JSON parts is not a 1-element array: %s", data)
		}
		part, ok := parts[0].(map[string]any)
		if !ok {
			t.Fatalf("part[0] is not an object: %s", data)
		}
		if part["kind"] != PartKindText {
			t.Errorf("part carries no kind discriminator: got %#v", part["kind"])
		}
		if part["text"] != "hello" {
			t.Errorf("part text mismatch: got %#v", part["text"])
		}
	})

	t.Run("delivery keys are camelCase broker names", func(t *testing.T) {
		env := Envelope{
			Message: Message{
				MessageID: "msg-0002",
				Role:      RoleAgent,
				Parts:     []Part{{Kind: PartKindText, Text: "x"}},
			},
			Delivery: Delivery{
				SenderID:   "codex-abcd1234",
				SenderKind: KindCodex,
				SentAt:     now,
				ExpiresAt:  now.Add(time.Hour),
			},
		}
		data, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("marshal envelope: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		delivery, ok := raw["delivery"].(map[string]any)
		if !ok {
			t.Fatalf("envelope JSON has no delivery object: %s", data)
		}
		for _, key := range []string{"senderId", "senderKind", "sentAt", "expiresAt"} {
			if _, present := delivery[key]; !present {
				t.Errorf("delivery JSON missing key %q: %s", key, data)
			}
		}
	})

	t.Run("round trip preserves message identity", func(t *testing.T) {
		env := Envelope{
			Message: Message{
				MessageID: "msg-0003",
				ContextID: "ctx-1",
				TaskID:    "task-1",
				Role:      RoleUser,
				Parts: []Part{
					{Kind: PartKindText, Text: "body"},
					{Kind: PartKindData, Data: json.RawMessage(`{"n":1}`)},
				},
			},
			Delivery: Delivery{
				SenderID:   "claude-abcd1234",
				SenderKind: KindClaude,
				SentAt:     now,
				ExpiresAt:  now.Add(time.Hour),
				ClaimedAt:  &now,
			},
		}
		data, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var back Envelope
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if back.Message.MessageID != env.Message.MessageID {
			t.Errorf("messageId round-trip mismatch: %q != %q", back.Message.MessageID, env.Message.MessageID)
		}
		if back.Message.Role != RoleUser {
			t.Errorf("role round-trip mismatch: %q", back.Message.Role)
		}
		if len(back.Message.Parts) != 2 || back.Message.Parts[1].Kind != PartKindData {
			t.Errorf("parts round-trip mismatch: %+v", back.Message.Parts)
		}
		if back.Delivery.ClaimedAt == nil || !back.Delivery.ClaimedAt.Equal(now) {
			t.Errorf("claimedAt round-trip mismatch: %+v", back.Delivery.ClaimedAt)
		}
	})

	t.Run("validation rejects invalid messages", func(t *testing.T) {
		valid := Message{
			MessageID: "msg-ok",
			Role:      RoleAgent,
			Parts:     []Part{{Kind: PartKindText, Text: "fine"}},
		}
		if err := valid.Validate(); err != nil {
			t.Errorf("valid message rejected: %v", err)
		}

		cases := []struct {
			name    string
			mutate  func(m *Message)
			wantSub string // expected error substring
		}{
			{
				name:    "empty messageId",
				mutate:  func(m *Message) { m.MessageID = "" },
				wantSub: "messageId",
			},
			{
				name:    "empty role",
				mutate:  func(m *Message) { m.Role = "" },
				wantSub: "role",
			},
			{
				name:    "unknown role",
				mutate:  func(m *Message) { m.Role = Role("system") },
				wantSub: "role",
			},
			{
				name:    "no parts",
				mutate:  func(m *Message) { m.Parts = nil },
				wantSub: "parts",
			},
			{
				name:    "empty parts slice",
				mutate:  func(m *Message) { m.Parts = []Part{} },
				wantSub: "parts",
			},
			{
				name: "raw part kind excluded",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: "raw", Text: "x"}}
				},
				wantSub: "kind",
			},
			{
				name: "url part kind excluded",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: "url", Text: "x"}}
				},
				wantSub: "kind",
			},
			{
				name: "empty part kind",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: "", Text: "x"}}
				},
				wantSub: "kind",
			},
			{
				name: "text part with empty text",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: PartKindText, Text: ""}}
				},
				wantSub: "text",
			},
			{
				name: "data part with empty data",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: PartKindData}}
				},
				wantSub: "data",
			},
			{
				name: "text part carrying data payload",
				mutate: func(m *Message) {
					m.Parts = []Part{{Kind: PartKindText, Text: "x", Data: json.RawMessage(`{"a":1}`)}}
				},
				wantSub: "text",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				m := valid
				tc.mutate(&m)
				err := m.Validate()
				if err == nil {
					t.Fatalf("expected validation error containing %q, got nil", tc.wantSub)
				}
				if !strings.Contains(err.Error(), tc.wantSub) {
					t.Errorf("error %q does not mention %q", err.Error(), tc.wantSub)
				}
			})
		}
	})

	t.Run("validation bounds text size and part count", func(t *testing.T) {
		oversized := Message{
			MessageID: "msg-big",
			Role:      RoleAgent,
			Parts:     []Part{{Kind: PartKindText, Text: strings.Repeat("a", config.DefaultSessionMsgMaxTextBytes+1)}},
		}
		if err := oversized.Validate(); err == nil {
			t.Errorf("expected oversize text rejection, got nil")
		}

		tooMany := Message{
			MessageID: "msg-many",
			Role:      RoleAgent,
			Parts:     make([]Part, config.DefaultSessionMsgMaxParts+1),
		}
		for i := range tooMany.Parts {
			tooMany.Parts[i] = Part{Kind: PartKindText, Text: "p"}
		}
		if err := tooMany.Validate(); err == nil {
			t.Errorf("expected part-count rejection, got nil")
		}
	})

	t.Run("delivery validation", func(t *testing.T) {
		base := Envelope{
			Message: Message{
				MessageID: "msg-d",
				Role:      RoleAgent,
				Parts:     []Part{{Kind: PartKindText, Text: "x"}},
			},
			Delivery: Delivery{
				SenderID:   "claude-abcd1234",
				SenderKind: KindClaude,
				SentAt:     now,
				ExpiresAt:  now.Add(time.Hour),
			},
		}
		if err := base.Validate(); err != nil {
			t.Errorf("valid envelope rejected: %v", err)
		}

		cases := []struct {
			name   string
			mutate func(e *Envelope)
		}{
			{"empty senderId", func(e *Envelope) { e.Delivery.SenderID = "" }},
			{"unknown senderKind", func(e *Envelope) { e.Delivery.SenderKind = "gemini" }},
			{"zero sentAt", func(e *Envelope) { e.Delivery.SentAt = time.Time{} }},
			{"expiresAt before sentAt", func(e *Envelope) { e.Delivery.ExpiresAt = e.Delivery.SentAt.Add(-time.Minute) }},
			{"invalid message", func(e *Envelope) { e.Message.MessageID = "" }},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				e := base
				tc.mutate(&e)
				if err := e.Validate(); err == nil {
					t.Errorf("expected envelope validation error, got nil")
				}
			})
		}
	})
}
