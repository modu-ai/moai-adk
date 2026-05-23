// Package harness — EventType extension and Event optional-field tests.
// REQ-HRN-OBS-001..005: 4 hook event types enum tests.
// REQ-HRN-OBS-009: preserve the existing 4-field schema via omitempty.
package harness

import (
	"encoding/json"
	"testing"
	"time"
)

// ─────────────────────────────────────────────
// T-A1: EventType constant extension tests
// ─────────────────────────────────────────────

// TestEventType_Extension verifies that the 3 new EventType constants exist with the
// exact expected string values.
// REQ-HRN-OBS-015: use SEMANTIC values (not the runtime hook names).
func TestEventType_Extension(t *testing.T) {
	t.Parallel()

	cases := []struct {
		constant EventType
		want     string
	}{
		{EventTypeSessionStop, "session_stop"},
		{EventTypeSubagentStop, "subagent_stop"},
		{EventTypeUserPrompt, "user_prompt"},
	}

	for _, tc := range cases {
		if string(tc.constant) != tc.want {
			t.Errorf("EventType constant value: got=%q, want=%q", string(tc.constant), tc.want)
		}
	}
}

// ─────────────────────────────────────────────
// T-A2: Event optional-field omitempty tests
// ─────────────────────────────────────────────

// TestEvent_OptionalFieldsOmitEmpty verifies that when all 12 optional fields hold
// zero values they are not included in the JSONL serialization output.
// REQ-HRN-OBS-009: preserve the existing 4-field schema — new fields are additive omitempty only.
func TestEvent_OptionalFieldsOmitEmpty(t *testing.T) {
	t.Parallel()

	evt := Event{
		Timestamp:     time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		EventType:     EventTypeAgentInvocation,
		Subject:       "Edit",
		ContextHash:   "",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Existing fields must always be present (REQ-HRN-FND-010 preservation)
	for _, field := range []string{"timestamp", "event_type", "subject", "tier_increment", "schema_version"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("existing field %q is missing — schema additivity violation", field)
		}
	}

	// Optional fields must be omitted when they hold zero values
	optionalFields := []string{
		"session_id",
		"last_assistant_message_hash",
		"last_assistant_message_len",
		"agent_name",
		"agent_type",
		"agent_id",
		"parent_session_id",
		"prompt_hash",
		"prompt_len",
		"prompt_lang",
		"prompt_preview",
		"prompt_full",
	}
	for _, field := range optionalFields {
		if _, ok := raw[field]; ok {
			t.Errorf("optional field %q was serialized despite holding the zero value — omitempty violation", field)
		}
	}
}

// TestEvent_OptionalFieldsSerializedWhenSet verifies that, when set, optional fields
// are included correctly in the serialization output.
func TestEvent_OptionalFieldsSerializedWhenSet(t *testing.T) {
	t.Parallel()

	evt := Event{
		Timestamp:     time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		EventType:     EventTypeSessionStop,
		Subject:       "SPEC-V3R4-HARNESS-002",
		ContextHash:   "",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
		// Stop-event optional fields
		SessionID:                "sess-abc123",
		LastAssistantMessageHash: "sha256-xyz",
		LastAssistantMessageLen:  4200,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Set optional fields must be present
	for _, field := range []string{"session_id", "last_assistant_message_hash", "last_assistant_message_len"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("set optional field %q is missing from serialization", field)
		}
	}

	// Unset optional fields must be absent
	for _, field := range []string{"agent_name", "agent_type", "prompt_hash"} {
		if _, ok := raw[field]; ok {
			t.Errorf("unset optional field %q appears in serialization — omitempty violation", field)
		}
	}
}
