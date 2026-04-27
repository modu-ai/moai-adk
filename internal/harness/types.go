// Package harness — JSONL 로그 스키마 타입 정의.
// REQ-HL-001: 이벤트 스키마 v1 정의.
package harness

import "time"

// LogSchemaVersion은 JSONL 로그 파일의 스키마 버전이다.
// 향후 스키마 변경 시 이 버전을 올려 하위 호환성을 추적한다.
const LogSchemaVersion = "v1"

// EventType은 관찰 가능한 이벤트 종류를 나타내는 열거형이다.
// REQ-HL-001: /moai 서브커맨드, 에이전트 호출, SPEC-ID 참조, /moai feedback 이벤트를 기록한다.
type EventType string

const (
	// EventTypeMoaiSubcommand는 /moai 서브커맨드 호출을 나타낸다.
	EventTypeMoaiSubcommand EventType = "moai_subcommand"

	// EventTypeAgentInvocation는 에이전트(subagent) 호출을 나타낸다.
	EventTypeAgentInvocation EventType = "agent_invocation"

	// EventTypeSpecReference는 SPEC-ID 참조(언급)를 나타낸다.
	EventTypeSpecReference EventType = "spec_reference"

	// EventTypeFeedback는 /moai feedback 이벤트를 나타낸다.
	EventTypeFeedback EventType = "feedback"
)

// Event는 usage-log.jsonl 파일의 단일 JSONL 라인 스키마이다.
// REQ-HL-001: timestamp, event_type, subject, context_hash, tier_increment 필드 포함.
//
// @MX:TODO: [AUTO] Phase 4에서 learning.enabled 설정 키로 gate 추가 예정.
// @MX:SPEC: SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-001
type Event struct {
	// Timestamp는 이벤트 발생 시각 (UTC).
	Timestamp time.Time `json:"timestamp"`

	// EventType는 이벤트 종류 (EventType 열거형).
	EventType EventType `json:"event_type"`

	// Subject는 이벤트의 주체 (예: "/moai plan", "expert-backend", "SPEC-001").
	Subject string `json:"subject"`

	// ContextHash는 세션 컨텍스트 식별자 (충돌 방지용 해시).
	ContextHash string `json:"context_hash"`

	// TierIncrement는 이 이벤트로 인한 tier 카운트 증분 (기본 0).
	TierIncrement int `json:"tier_increment"`

	// SchemaVersion은 로그 스키마 버전 (LogSchemaVersion).
	SchemaVersion string `json:"schema_version"`
}
