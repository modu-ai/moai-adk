// Package harness — JSONL 로그 스키마 타입 정의.
// REQ-HL-001: 이벤트 스키마 v1 정의.
// REQ-HL-002: Pattern, Tier, Promotion 타입 정의.
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

// ─────────────────────────────────────────────
// Phase 2: 패턴 분류 타입 (REQ-HL-002)
// ─────────────────────────────────────────────

// Tier는 패턴의 성숙도 단계를 나타내는 열거형이다.
// REQ-HL-002: 누적 관찰 횟수 {1,3,5,10}에 따라 분류된다.
//
// @MX:ANCHOR: [AUTO] ClassifyTier 반환값으로 다수 호출자에서 사용.
// @MX:REASON: [AUTO] fan_in >= 3: learner.go, learner_test.go, applier.go
type Tier int

const (
	// TierObservation은 1회 이상 관찰된 패턴 (가장 낮은 tier).
	TierObservation Tier = iota + 1

	// TierHeuristic은 3회 이상 관찰된 패턴; description enrichment 대상.
	TierHeuristic

	// TierRule은 5회 이상 관찰된 패턴; trigger injection 대상.
	TierRule

	// TierAutoUpdate는 10회 이상 관찰된 패턴; 자동 업데이트 후보.
	TierAutoUpdate
)

// String은 Tier를 사람이 읽을 수 있는 문자열로 반환한다.
func (t Tier) String() string {
	switch t {
	case TierObservation:
		return "observation"
	case TierHeuristic:
		return "heuristic"
	case TierRule:
		return "rule"
	case TierAutoUpdate:
		return "auto_update"
	default:
		return "unknown"
	}
}

// Pattern은 (event_type, subject, context_hash) 조합으로 식별되는 단일 패턴이다.
// REQ-HL-002: learner가 JSONL 로그를 읽어 패턴으로 집계한다.
type Pattern struct {
	// Key는 "event_type:subject:context_hash" 형식의 고유 식별자이다.
	Key string

	// EventType은 패턴의 이벤트 종류이다.
	EventType EventType

	// Subject는 패턴의 주체 (예: "/moai plan", "expert-backend").
	Subject string

	// ContextHash는 컨텍스트 식별자이다.
	ContextHash string

	// Count는 이 패턴이 관찰된 총 횟수이다.
	Count int

	// Confidence는 패턴 신뢰도 (0.0 ~ 1.0).
	// 0.70 미만이면 Count에 관계없이 TierObservation으로 분류된다.
	Confidence float64

	// Tier는 현재 분류된 tier이다 (ClassifyTier 호출 후 설정).
	Tier Tier
}

// Promotion은 tier-promotions.jsonl에 기록되는 tier 승격 이벤트이다.
// plan.md §4.2 스키마 기준.
type Promotion struct {
	// Ts는 승격 시각 (UTC RFC3339).
	Ts time.Time `json:"ts"`

	// PatternKey는 "event_type:subject" 형식 (plan.md §4.2 기준).
	PatternKey string `json:"pattern_key"`

	// FromTier는 승격 이전 tier 문자열.
	FromTier string `json:"from_tier"`

	// ToTier는 승격된 tier 문자열.
	ToTier string `json:"to_tier"`

	// ObservationCount는 승격 시점의 관찰 횟수.
	ObservationCount int `json:"observation_count"`

	// Confidence는 패턴 신뢰도.
	Confidence float64 `json:"confidence"`
}
