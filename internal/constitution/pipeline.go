package constitution

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Pipeline은 5-layer safety gate를 실행하는 constitutional amendment pipeline이다.
// SPEC-V3R2-CON-002 REQ-CON-002-002 구현.
type Pipeline struct {
	// FrozenGuard는 Layer 1 gate이다.
	FrozenGuard FrozenGuard
	// Canary는 Layer 2 gate이다.
	Canary Canary
	// ContradictionDetector는 Layer 3 gate이다.
	ContradictionDetector ContradictionDetector
	// RateLimiter는 Layer 4 gate이다.
	RateLimiter RateLimiter
	// HumanOversight는 Layer 5 gate이다.
	HumanOversight HumanOversight

	// LockFilePath는 single-writer lock 파일 경로이다.
	LockFilePath string
}

// NewPipeline은 기본 구현으로 Pipeline을 생성한다.
func NewPipeline() *Pipeline {
	return &Pipeline{
		FrozenGuard:          NewFrozenGuard(),
		Canary:               NewCanary(),
		ContradictionDetector: NewContradictionDetector(),
		RateLimiter:          NewRateLimiter(),
		HumanOversight:       NewHumanOversight(),
	}
}

// Execute는 proposal에 대해 5-layer safety gate를 실행하고 amendment를 적용한다.
// Dry-run mode: 모든 layer를 실행하지만 파일을 수정하지 않는다.
// SPEC-V3R2-CON-002 AC-CON-002-001 구현.
//
// Layer 실행 순서 (FROZEN):
// 1. FrozenGuard: Frozen zone check
// 2. Canary: Shadow evaluation
// 3. ContradictionDetector: Conflict scan
// 4. RateLimiter: Frequency check
// 5. HumanOversight: User approval
//
// 성공 시:
// - Registry 파일 업데이트 (source rule file 수정)
// - Zone registry 업데이트 (.claude/rules/moai/core/zone-registry.md)
// - Evolution-log.md에 기록
//
// 실패 시: 해당 layer 에러 반환.
func (p *Pipeline) Execute(proposal *AmendmentProposal, projectDir string, dryRun bool) (*AmendmentLog, error) {
	// 0. Single-writer lock 획득 시도
	if err := p.acquireLock(dryRun); err != nil {
		return nil, err
	}
	if !dryRun {
		defer p.releaseLock()
	}

	// Registry 로드
	registryPath := filepath.Join(projectDir, ".claude", "rules", "moai", "core", "zone-registry.md")
	registry, err := LoadRegistry(registryPath, projectDir)
	if err != nil {
		return nil, fmt.Errorf("registry 로드 오류: %w", err)
	}

	// 현재 rule 조회
	currentRule, exists := registry.Get(proposal.RuleID)
	if !exists {
		return nil, fmt.Errorf("rule %q을(를) 찾을 수 없음", proposal.RuleID)
	}

	// canary_gate가 false인 rule은 Canary skip
	skipCanary := !currentRule.CanaryGate

	// ===== Layer 1: FrozenGuard =====
	if err := p.FrozenGuard.Check(proposal, currentRule.Zone); err != nil {
		return nil, fmt.Errorf("layer 1 (FrozenGuard) 실패: %w", err)
	}

	// ===== Layer 2: Canary =====
	if !skipCanary {
		canaryResult, err := p.Canary.Evaluate(proposal, projectDir)
		proposal.CanaryResult = canaryResult
		if err != nil {
			// CanaryUnavailable은 치명적이 아님 (skip과 유사)
			if _, unavailable := err.(*ErrCanaryUnavailable); !unavailable {
				return nil, fmt.Errorf("layer 2 (Canary) 실패: %w", err)
			}
			// CanaryUnavailable은 계속 진행
		} else if !canaryResult.Passed {
			return nil, fmt.Errorf("layer 2 (Canary) 실패: score drop %.2f > threshold %.2f",
				canaryResult.MaxDrop, canaryScoreDropThreshold)
		}
	} else {
		proposal.CanaryResult = &CanaryResult{
			Available: false,
			Reason:    fmt.Sprintf("Rule %q의 canary_gate=false", proposal.RuleID),
		}
	}

	// ===== Layer 3: ContradictionDetector =====
	contradictionResult, err := p.ContradictionDetector.Scan(proposal, registry)
	proposal.Contradicts = contradictionResult
	if err != nil {
		return nil, fmt.Errorf("layer 3 (ContradictionDetector) 실패: %w", err)
	}

	// ===== Layer 4: RateLimiter =====
	evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")
	if err := p.RateLimiter.Admit(proposal, evolutionLogPath); err != nil {
		return nil, fmt.Errorf("layer 4 (RateLimiter) 실패: %w", err)
	}

	// ===== Layer 5: HumanOversight =====
	approved, err := p.HumanOversight.Approve(proposal, dryRun)
	if err != nil {
		return nil, fmt.Errorf("layer 5 (HumanOversight) 실패: %w", err)
	}
	if !approved {
		return nil, fmt.Errorf("사용자가 amendment를 거부했습니다")
	}
	proposal.Approved = true
	proposal.ApprovedBy = "human"
	proposal.ApprovedAt = time.Now()

	// ===== Amendment 적용 =====
	if dryRun {
		// Dry-run: log 생성만 반환
		log := p.createLogEntry(proposal, currentRule.Zone)
		return log, nil
	}

	// 실제 적용: source file, registry, evolution-log 업데이트
	if err := p.applyAmendment(proposal, currentRule, projectDir, registryPath); err != nil {
		return nil, fmt.Errorf("amendment 적용 오류: %w", err)
	}

	log := p.createLogEntry(proposal, currentRule.Zone)
	return log, nil
}

// createLogEntry는 proposal에서 AmendmentLog를 생성한다.
func (p *Pipeline) createLogEntry(proposal *AmendmentProposal, originalZone Zone) *AmendmentLog {
	// Canary verdict
	canaryVerdict := "skipped"
	if proposal.CanaryResult != nil {
		if proposal.CanaryResult.Available {
			if proposal.CanaryResult.Passed {
				canaryVerdict = "passed"
			} else {
				canaryVerdict = "rejected"
			}
		} else {
			canaryVerdict = "unavailable"
		}
	}

	// Contradictions
	var contradictions []string
	if proposal.Contradicts != nil {
		for _, c := range proposal.Contradicts.Conflicts {
			contradictions = append(contradictions,
				fmt.Sprintf("%s: %s", c.ConflictingRuleID, c.Description))
		}
	}

	return &AmendmentLog{
		ID:            "", // Execute 후반에서 생성
		RuleID:        proposal.RuleID,
		ZoneBefore:    originalZone,
		ZoneAfter:     originalZone, // Zone 변경은 demotion evidence로만 허용
		ClauseBefore:  proposal.Before,
		ClauseAfter:   proposal.After,
		CanaryVerdict: canaryVerdict,
		Contradictions: contradictions,
		ApprovedBy:    proposal.ApprovedBy,
		ApprovedAt:    proposal.ApprovedAt,
		RolledBack:    false,
	}
}

// applyAmendment는 amendment를 파일 시스템에 적용한다.
// 3개 파일 수정: source rule file, zone registry, evolution-log.
func (p *Pipeline) applyAmendment(proposal *AmendmentProposal, rule Rule, projectDir, registryPath string) error {
	// 1. Source rule file 업데이트
	sourceFilePath := rule.File
	if !filepath.IsAbs(sourceFilePath) {
		sourceFilePath = filepath.Join(projectDir, sourceFilePath)
	}
	if err := updateSourceFile(sourceFilePath, rule.Anchor, proposal.After); err != nil {
		return fmt.Errorf("source file 업데이트 오류: %w", err)
	}

	// 2. Zone registry 업데이트 (Clause만)
	if err := updateRegistryClause(registryPath, proposal.RuleID, proposal.After); err != nil {
		return fmt.Errorf("registry 업데이트 오류: %w", err)
	}

	// 3. Evolution-log에 기록
	evolutionLogPath := filepath.Join(projectDir, ".moai", "research", "evolution-log.md")
	logs, _ := LoadEvolutionLogs(evolutionLogPath)
	log := p.createLogEntry(proposal, rule.Zone)
	log.ID = GenerateLogID(time.Now(), logs)

	if err := AppendEvolutionLog(evolutionLogPath, log); err != nil {
		return fmt.Errorf("evolution-log 기록 오류: %w", err)
	}

	return nil
}

// acquireLock은 single-writer lock을 획득한다.
func (p *Pipeline) acquireLock(dryRun bool) error {
	if dryRun {
		return nil // Dry-run은 lock 불필요
	}

	lockPath := p.LockFilePath
	if lockPath == "" {
		// 기본 경로 사용
		lockPath = ".moai/research/.amendment.lock"
	}

	if _, err := os.Stat(lockPath); err == nil {
		return &ErrAmendmentInProgress{LockFilePath: lockPath}
	}

	// Lock 파일 생성
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return fmt.Errorf("lock 디렉토리 생성 오류: %w", err)
	}
	if err := os.WriteFile(lockPath, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		return fmt.Errorf("lock 파일 생성 오류: %w", err)
	}

	p.LockFilePath = lockPath
	return nil
}

// releaseLock은 single-writer lock을 해제한다.
func (p *Pipeline) releaseLock() {
	if p.LockFilePath != "" {
		_ = os.Remove(p.LockFilePath)
		p.LockFilePath = ""
	}
}

// updateSourceFile은 source rule file에서 해당 anchor의 clause를 업데이트한다.
// TODO: 구현 필요 - anchor 검색 + 치환 로직.
func updateSourceFile(filePath, anchor, newClause string) error {
	// 간단 구현: 파일 전체 읽고 anchor 다음 줄을 치환
	// 실제 구현에서는 markdown 섹션 검색이 필요
	return fmt.Errorf("updateSourceFile: 아직 구현되지 않음 (path=%s, anchor=%s)", filePath, anchor)
}

// updateRegistryClause는 zone registry에서 해당 rule의 clause를 업데이트한다.
// TODO: 구현 필요 - YAML 파싱 + 치환 로직.
func updateRegistryClause(registryPath, ruleID, newClause string) error {
	// 간단 구현: YAML 파싱 후 rule ID로 찾아 clause 치환
	return fmt.Errorf("updateRegistryClause: 아직 구현되지 않음 (path=%s, rule=%s)", registryPath, ruleID)
}
