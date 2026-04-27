// Package harness — JSONL 로그 스키마 타입 정의.
// REQ-HL-001: 이벤트 스키마 v1 정의.
// REQ-HL-002: Pattern, Tier, Promotion 타입 정의.
// REQ-HL-005~008: Phase 3 Safety Architecture 타입 정의.
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

// ─────────────────────────────────────────────
// Phase 3: 5-Layer Safety Architecture 타입 (REQ-HL-005~008)
// ─────────────────────────────────────────────

// Proposal은 자동 업데이트 제안 페이로드이다.
// REQ-HL-005: Tier 4 도달 시 system이 생성하여 safety pipeline을 통과시킨다.
//
// @MX:ANCHOR: [AUTO] Proposal은 safety pipeline 전체의 입력 타입이다.
// @MX:REASON: [AUTO] fan_in >= 3: safety/pipeline.go, safety/frozen_guard.go, safety/canary.go
type Proposal struct {
	// ID는 제안 고유 식별자이다.
	ID string `json:"id"`

	// TargetPath는 수정 대상 파일 경로이다 (예: ".claude/skills/my-harness-x/SKILL.md").
	TargetPath string `json:"target_path"`

	// FieldKey는 수정 대상 frontmatter 필드 (예: "description", "triggers").
	FieldKey string `json:"field_key"`

	// NewValue는 새로 적용할 값이다.
	NewValue string `json:"new_value"`

	// PatternKey는 이 제안을 생성한 패턴 식별자이다.
	PatternKey string `json:"pattern_key"`

	// Tier는 이 제안을 생성한 패턴의 tier이다.
	Tier Tier `json:"tier"`

	// ObservationCount는 패턴 관찰 횟수이다.
	ObservationCount int `json:"observation_count"`

	// CreatedAt은 제안 생성 시각이다.
	CreatedAt time.Time `json:"created_at"`
}

// DecisionKind는 safety pipeline의 최종 판정 종류이다.
type DecisionKind string

const (
	// DecisionApproved는 모든 safety layer를 통과한 경우이다.
	DecisionApproved DecisionKind = "approved"

	// DecisionRejected는 하나 이상의 safety layer에서 거부된 경우이다.
	DecisionRejected DecisionKind = "rejected"

	// DecisionPendingApproval은 L5 Human Oversight 승인 대기 중인 경우이다.
	DecisionPendingApproval DecisionKind = "pending_approval"
)

// Decision은 safety pipeline 실행 결과이다.
// REQ-HL-005~008: L1→L2→L3→L4→L5 평가 결과를 종합한다.
type Decision struct {
	// Kind는 판정 결과 종류이다.
	Kind DecisionKind `json:"kind"`

	// RejectedBy는 거부를 발생시킨 layer 번호이다 (1~5, 통과 시 0).
	RejectedBy int `json:"rejected_by,omitempty"`

	// Reason은 거부 또는 대기 이유이다.
	Reason string `json:"reason,omitempty"`

	// OversightProposal은 L5 승인 대기 시 orchestrator에게 반환할 페이로드이다.
	OversightProposal *OversightProposal `json:"oversight_proposal,omitempty"`
}

// Session은 effectiveness 계산을 위한 단일 세션 메트릭이다.
// REQ-HL-007: canary check에서 baseline vs proposal 비교에 사용된다.
type Session struct {
	// ID는 세션 식별자이다.
	ID string `json:"id"`

	// SubcommandSuccessRate는 /moai 서브커맨드 성공률 (0.0~1.0).
	SubcommandSuccessRate float64 `json:"subcommand_success_rate"`

	// AgentInvocationSuccess는 에이전트 호출 성공률 (0.0~1.0).
	AgentInvocationSuccess float64 `json:"agent_invocation_success"`

	// CompletionRate는 SPEC 완료율 (0.0~1.0).
	CompletionRate float64 `json:"completion_rate"`

	// Timestamp는 세션 시작 시각이다.
	Timestamp time.Time `json:"timestamp"`
}

// CanaryResult는 canary check 실행 결과이다.
// REQ-HL-007: baseline 대비 effectiveness 하락이 0.10 초과이면 rejected=true.
type CanaryResult struct {
	// BaselineScore는 제안 적용 전 baseline effectiveness 점수이다.
	BaselineScore float64 `json:"baseline_score"`

	// ProjectedScore는 제안 적용 후 예측 effectiveness 점수이다.
	ProjectedScore float64 `json:"projected_score"`

	// Delta는 ProjectedScore - BaselineScore이다 (음수이면 하락).
	Delta float64 `json:"delta"`

	// Rejected는 delta가 임계값(-0.10)을 초과하여 거부된 경우 true이다.
	Rejected bool `json:"rejected"`

	// Reason은 거부 이유 (Rejected=true일 때만 설정).
	Reason string `json:"reason,omitempty"`
}

// ContradictionType은 탐지된 모순의 종류이다.
type ContradictionType string

const (
	// ContradictionOverlappingTriggers는 여러 skill의 trigger keyword가 겹치는 경우이다.
	ContradictionOverlappingTriggers ContradictionType = "overlapping_triggers"

	// ContradictionChainRules는 chaining-rules.yaml의 모순(같은 phase, 서로 다른 규칙)이다.
	ContradictionChainRules ContradictionType = "contradictory_chain_rules"
)

// ContradictionItem은 단일 모순 항목이다.
type ContradictionItem struct {
	// Type은 모순 종류이다.
	Type ContradictionType `json:"type"`

	// Description은 모순 상세 설명이다.
	Description string `json:"description"`

	// ConflictingPaths는 모순이 발생한 파일 경로 목록이다.
	ConflictingPaths []string `json:"conflicting_paths"`

	// ConflictingValues는 모순을 일으킨 값 쌍이다.
	ConflictingValues []string `json:"conflicting_values"`
}

// ContradictionReport는 contradiction detector의 탐지 결과이다.
// REQ-HL-008: 빈 Items이면 모순 없음, Items가 있으면 orchestrator에게 반환한다.
type ContradictionReport struct {
	// Items는 탐지된 모순 목록이다.
	Items []ContradictionItem `json:"items"`
}

// HasContradiction은 모순이 존재하면 true를 반환한다.
func (r ContradictionReport) HasContradiction() bool {
	return len(r.Items) > 0
}

// OversightOption은 AskUserQuestion 단일 선택지이다.
// REQ-HL-005: max 4 options, 첫 번째 옵션은 recommended.
type OversightOption struct {
	// Label은 선택지 레이블이다 (예: "승인 (권장)", "거부").
	Label string `json:"label"`

	// Description은 선택지 상세 설명이다.
	Description string `json:"description"`

	// Value는 프로그래밍 처리용 값이다 (예: "approve", "reject").
	Value string `json:"value"`

	// Recommended는 권장 선택지이면 true이다 (첫 번째 옵션에만 설정).
	Recommended bool `json:"recommended"`
}

// OversightProposal은 L5 Human Oversight가 orchestrator에게 반환하는 페이로드이다.
// REQ-HL-005: subagent는 AskUserQuestion을 직접 호출하지 않고 이 구조체를 반환한다.
// orchestrator(MoAI)가 이 페이로드를 받아 AskUserQuestion으로 사용자에게 표시한다.
//
// @MX:ANCHOR: [AUTO] OversightProposal은 subagent→orchestrator 인터페이스 경계이다.
// @MX:REASON: [AUTO] fan_in >= 3: safety/oversight.go, safety/pipeline.go, Phase 4 coordinator
type OversightProposal struct {
	// Question은 사용자에게 표시할 질문 텍스트이다.
	Question string `json:"question"`

	// Options는 선택지 목록이다 (max 4, 첫 번째가 recommended).
	Options []OversightOption `json:"options"`

	// ProposalID는 관련 Proposal의 ID이다.
	ProposalID string `json:"proposal_id"`

	// Context는 사용자가 결정하기 위한 부가 컨텍스트이다.
	Context string `json:"context"`
}
