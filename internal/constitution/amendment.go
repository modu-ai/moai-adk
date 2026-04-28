// Package constitution은 MoAI-ADK 규칙 트리의 FROZEN/EVOLVABLE zone 모델을 구현한다.
// amendment는 SPEC-V3R2-CON-002의 5-layer safety gate를 통한 constitutional amendment protocol을 구현한다.
package constitution

import (
	"fmt"
	"time"
)

// ErrFrozenAmendment는 Frozen zone rule에 대한 amendment 시도를 나타내는 에러이다.
// Frozen→Evolvable demotion 증거가 없을 때 반환된다.
type ErrFrozenAmendment struct {
	RuleID string
	Reason string
}

func (e *ErrFrozenAmendment) Error() string {
	return fmt.Sprintf("Frozen zone amendment 거부: %s - %s", e.RuleID, e.Reason)
}

// ErrCanaryUnavailable는 Canary evaluation을 실행할 SPEC이 부족할 때 반환된다.
type ErrCanaryUnavailable struct {
	RequiredCount int
	ActualCount   int
}

func (e *ErrCanaryUnavailable) Error() string {
	return fmt.Sprintf("Canary unavailable: 최소 %d개 SPEC 필요 (현재: %d)", e.RequiredCount, e.ActualCount)
}

// ErrCanaryRejected는 Canary evaluation에서 점수 하락이 탐지되었을 때 반환된다.
type ErrCanaryRejected struct {
	RuleID         string
	ScoreDrop      float64
	Threshold      float64
	AffectedSpecs  []string
}

func (e *ErrCanaryRejected) Error() string {
	return fmt.Sprintf("Canary rejected: %s 점수 하락 %.2f > 임계값 %.2f (영향받은 SPEC: %v)",
		e.RuleID, e.ScoreDrop, e.Threshold, e.AffectedSpecs)
}

// ErrContradictionDetected는 amendment가 기존 rule과 모순될 때 반환된다.
type ErrContradictionDetected struct {
	NewRuleID      string
	ConflictingIDs []string
	Conflicts      []string // 모순 설명
}

func (e *ErrContradictionDetected) Error() string {
	return fmt.Sprintf("Contradiction detected: %s가 다음 rule과 모순됨: %v\n상세: %v",
		e.NewRuleID, e.ConflictingIDs, e.Conflicts)
}

// ErrRateLimitExceeded는 amendment 빈도 제한을 초과했을 때 반환된다.
type ErrRateLimitExceeded struct {
	MaxPerWeek    int
	CooldownHours int
	NextAllowedAt time.Time
}

func (e *ErrRateLimitExceeded) Error() string {
	return fmt.Sprintf("Rate limit exceeded: 최대 %d회/주, %d시간 쿨다운 (다음 가능: %s)",
		e.MaxPerWeek, e.CooldownHours, e.NextAllowedAt.Format(time.RFC3339))
}

// ErrAmendmentInProgress는 다른 amendment가 이미 진행 중일 때 반환된다.
// Single-writer lock 위반.
type ErrAmendmentInProgress struct {
	LockFilePath string
}

func (e *ErrAmendmentInProgress) Error() string {
	return fmt.Sprintf("Amendment already in progress: lock file %s 존재", e.LockFilePath)
}

// ErrRolledBack은 이미 rollback된 rule을 재-amendment하려 할 때 반환된다.
type ErrRolledBack struct {
	RuleID       string
	RolledBackAt time.Time
	CooldownDays int
}

func (e *ErrRolledBack) Error() string {
	return fmt.Sprintf("Rolled back rule: %s는 %s에 rollback됨. %d일 쿨다운 중",
		e.RuleID, e.RolledBackAt.Format(time.RFC3339), e.CooldownDays)
}

// AmendmentProposal은 constitutional amendment proposal을 나타낸다.
// SPEC-V3R2-CON-002 REQ-CON-002-001 직접 구현.
type AmendmentProposal struct {
	// RuleID는 수정하려는 rule의 ID이다 (CONST-V3R2-NNN).
	RuleID string
	// Before는 현재 clause 텍스트이다.
	Before string
	// After는 새로운 clause 텍스트이다.
	After string
	// Evidence는 amendment justification이다 (Frozen zone 필수).
	Evidence string
	// CanaryResult는 canary evaluation 결과이다.
	CanaryResult *CanaryResult
	// Contradicts는 모순 탐지 결과이다.
	Contradicts *ContradictionResult
	// Approved는 승인 여부이다 (HumanOversight layer).
	Approved bool
	// ApprovedBy는 승인자 식별자이다 ("human" 또는 시스템 ID).
	ApprovedBy string
	// ApprovedAt은 승인 시각이다.
	ApprovedAt time.Time
}

// CanaryResult는 canary evaluation 결과를 나타낸다.
type CanaryResult struct {
	// Available은 canary 실행 가능 여부이다.
	Available bool
	// EvaluatedSpecs는 평가한 SPEC ID 목록이다 (최대 3개).
	EvaluatedSpecs []string
	// ScoreBefore는 amendment 전 평균 점수이다.
	ScoreBefore float64
	// ScoreAfter는 amendment 후 평균 점수이다 (shadow evaluation).
	ScoreAfter float64
	// MaxDrop은 최대 점수 하락이다.
	MaxDrop float64
	// Passed는 canary 통과 여부이다 (ScoreDrop <= 0.10).
	Passed bool
	// Reason은 canary 불가 또는 실패 사유이다.
	Reason string
}

// ContradictionResult는 모순 탐지 결과를 나타낸다.
type ContradictionResult struct {
	// Conflicts는 충돌하는 rule ID와 설명이다.
	Conflicts []ConflictDetail
	// HasBlockingContradiction은 차단 모순 존재 여부이다.
	HasBlockingContradiction bool
}

// ConflictDetail은 단일 모순 상세를 나타낸다.
type ConflictDetail struct {
	ConflictingRuleID string
	Description       string
	IsBlocking        bool // true면 해결 없이 진행 불가
}

// FrozenGuard는 Layer 1 safety gate이다.
// Frozen zone rule 수정을 차단한다.
type FrozenGuard interface {
	// Check는 proposal이 Frozen zone rule을 수정하는지 확인한다.
	// Frozen→Evolvable demotion은 Evidence로 허용.
	Check(proposal *AmendmentProposal, currentZone Zone) error
}

// Canary는 Layer 2 safety gate이다.
// Shadow evaluation으로 점수 하락을 탐지한다.
type Canary interface {
	// Evaluate는 proposal의 영향을 last 3 completed SPECs에 대해 평가한다.
	// < 3 SPECs → CanaryUnavailable error.
	// ScoreDrop > 0.10 → CanaryRejected error.
	// canary_gate: false인 rule → CanaryUnavailable (skip).
	Evaluate(proposal *AmendmentProposal, projectDir string) (*CanaryResult, error)
}

// ContradictionDetector는 Layer 3 safety gate이다.
// 기존 rule과의 모순을 탐지한다.
type ContradictionDetector interface {
	// Scan은 proposal이 registry의 다른 rule과 모순하는지 확인한다.
	// 모순을 찾으면 ConflictDetail 목록을 반환한다.
	Scan(proposal *AmendmentProposal, registry *Registry) (*ContradictionResult, error)
}

// RateLimiter는 Layer 4 safety gate이다.
// Amendment 빈도를 제한한다.
type RateLimiter interface {
	// Admit는 proposal이 rate limit 내에 있는지 확인한다.
	// Max 3 amendments/7-day window, 24h cooldown, 50 active learnings cap.
	// Rolled-back rules는 30-day cooldown.
	Admit(proposal *AmendmentProposal, evolutionLogPath string) error
}

// HumanOversight는 Layer 5 safety gate이다.
// CLI에서 terminal Y/N prompt로 승인을 받는다.
// (AskUserQuestion은 orchestrator-only이므로 CLI에서 직접 구현).
type HumanOversight interface {
	// Approve는 사용자에게 proposal diff를 보여고 승인을 요청한다.
	// Dry-run mode에서는 항상 true 반환 (실제 승인 없음).
	Approve(proposal *AmendmentProposal, dryRun bool) (bool, error)
}

// AmendmentLog는 evolution-log.md 엔트리를 나타낸다.
// SPEC-V3R2-CON-002 REQ-CON-002-003 직접 구현.
type AmendmentLog struct {
	// ID는 LEARN-YYYYMMDD-NNN 형식의 고유 식별자이다.
	ID string
	// RuleID는 수정된 rule ID이다 (CONST-V3R2-NNN).
	RuleID string
	// ZoneBefore는 수정 전 zone이다.
	ZoneBefore Zone
	// ZoneAfter는 수정 후 zone이다.
	ZoneAfter Zone
	// ClauseBefore는 수정 전 clause 텍스트이다.
	ClauseBefore string
	// ClauseAfter는 수정 후 clause 텍스트이다.
	ClauseAfter string
	// CanaryVerdict는 canary evaluation 결과이다.
	CanaryVerdict string // "passed", "skipped", "rejected", "unavailable"
	// Contradictions는 모순 탐지 결과이다.
	Contradictions []string
	// ApprovedBy는 승인자이다 ("human" 또는 시스템 ID).
	ApprovedBy string
	// ApprovedAt은 승인 시각이다.
	ApprovedAt time.Time
	// RolledBack은 rollback 여부이다.
	RolledBack bool
	// RollbackReason은 rollback 사유이다 (선택).
	RollbackReason string
	// RollbackAt은 rollback 시각이다 (선택).
	RollbackAt *time.Time
}

// Validate는 AmendmentLog의 필수 필드를 검증한다.
func (l *AmendmentLog) Validate() error {
	if l.ID == "" {
		return fmt.Errorf("amendment log ID가 비어 있다")
	}
	if l.RuleID == "" {
		return fmt.Errorf("rule ID가 비어 있다")
	}
	if l.ClauseBefore == "" || l.ClauseAfter == "" {
		return fmt.Errorf("clause가 비어 있다")
	}
	if l.ApprovedBy == "" {
		return fmt.Errorf("approved by가 비어 있다")
	}
	if l.ApprovedAt.IsZero() {
		return fmt.Errorf("approved at이 비어 있다")
	}
	return nil
}
