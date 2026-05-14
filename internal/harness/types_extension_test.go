// Package harness — EventType 확장 및 Event 옵션 필드 테스트.
// REQ-HRN-OBS-001..005: 4개 후크 이벤트 유형 열거형 테스트.
// REQ-HRN-OBS-009: omitempty로 인한 기존 4-필드 스키마 보존 테스트.
package harness

import (
	"encoding/json"
	"testing"
	"time"
)

// ─────────────────────────────────────────────
// T-A1: EventType 상수 확장 테스트
// ─────────────────────────────────────────────

// TestEventType_Extension은 3개의 신규 EventType 상수가 정확한 문자열 값으로
// 존재하는지 검증한다.
// REQ-HRN-OBS-015: SEMANTIC 값(런타임 훅 이름 아님) 사용.
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
			t.Errorf("EventType 상수 값: got=%q, want=%q", string(tc.constant), tc.want)
		}
	}
}

// ─────────────────────────────────────────────
// T-A2: Event 옵션 필드 omitempty 테스트
// ─────────────────────────────────────────────

// TestEvent_OptionalFieldsOmitEmpty는 12개 옵션 필드가 전부 제로 값일 때
// JSONL 직렬화 결과에 포함되지 않음을 검증한다.
// REQ-HRN-OBS-009: 기존 4-필드 스키마 보존 — 새 필드는 additive omitempty 전용.
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
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
	}

	// 기존 필드는 항상 존재해야 함 (REQ-HRN-FND-010 보존)
	for _, field := range []string{"timestamp", "event_type", "subject", "tier_increment", "schema_version"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("기존 필드 %q가 누락됨 — schema additivity 위반", field)
		}
	}

	// 옵션 필드는 제로 값일 때 누락되어야 함
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
			t.Errorf("옵션 필드 %q가 제로 값임에도 직렬화됨 — omitempty 위반", field)
		}
	}
}

// TestEvent_OptionalFieldsSerializedWhenSet은 옵션 필드가 설정됐을 때
// 직렬화 결과에 올바르게 포함되는지 검증한다.
func TestEvent_OptionalFieldsSerializedWhenSet(t *testing.T) {
	t.Parallel()

	evt := Event{
		Timestamp:     time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC),
		EventType:     EventTypeSessionStop,
		Subject:       "SPEC-V3R4-HARNESS-002",
		ContextHash:   "",
		TierIncrement: 0,
		SchemaVersion: LogSchemaVersion,
		// Stop 이벤트 옵션 필드
		SessionID:                "sess-abc123",
		LastAssistantMessageHash: "sha256-xyz",
		LastAssistantMessageLen:  4200,
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("json.Marshal 실패: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal 실패: %v", err)
	}

	// 설정된 옵션 필드는 존재해야 함
	for _, field := range []string{"session_id", "last_assistant_message_hash", "last_assistant_message_len"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("설정된 옵션 필드 %q가 직렬화에서 누락됨", field)
		}
	}

	// 설정 안 된 옵션 필드는 없어야 함
	for _, field := range []string{"agent_name", "agent_type", "prompt_hash"} {
		if _, ok := raw[field]; ok {
			t.Errorf("미설정 옵션 필드 %q가 직렬화에 포함됨 — omitempty 위반", field)
		}
	}
}
