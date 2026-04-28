// Package harness — frontmatter 수정 applier.
// REQ-HL-003: description enrichment (Tier 2 heuristic).
// REQ-HL-004: trigger injection (Tier 3 rule, feature-gated).
// REQ-HL-005: Apply() — snapshot 우선 생성 후 파일 수정 (Phase 4).
package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// enableTriggerInjectionWrites는 InjectTrigger의 실제 파일 쓰기를 활성화하는 feature flag이다.
// Phase 2에서는 기본 OFF — dedup 로직만 검증하고 실제 write는 수행하지 않는다.
//
// @MX:TODO: [AUTO] Phase 4: wire learning.auto_apply config to enable writes
// @MX:SPEC: SPEC-V3R3-HARNESS-LEARNING-001 REQ-HL-004 (T-P2-05)
var enableTriggerInjectionWrites = false

// Applier는 SKILL.md 파일의 frontmatter를 수정하는 컴포넌트이다.
// 모든 수정은 description 또는 triggers 필드만 대상으로 하며,
// 다른 frontmatter 필드와 body는 byte-identical하게 보존된다.
//
// @MX:ANCHOR: [AUTO] EnrichDescription, InjectTrigger는 학습 파이프라인 write 경로.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, safety.go(Phase 3), CLI apply(Phase 4)
type Applier struct {
	// allowWrites는 InjectTrigger의 실제 파일 쓰기를 허용하는 인스턴스 레벨 flag이다.
	// 기본값은 enableTriggerInjectionWrites (패키지 레벨 flag).
	// 테스트에서 newApplierWithWritesEnabled()로 true 설정 가능.
	allowWrites bool
}

// NewApplier는 기본 Applier를 생성한다.
// InjectTrigger의 실제 파일 쓰기는 패키지 레벨 flag(enableTriggerInjectionWrites)에 따른다.
func NewApplier() *Applier {
	return &Applier{allowWrites: enableTriggerInjectionWrites}
}

// newApplierWithWritesEnabled는 InjectTrigger 실제 쓰기가 활성화된 Applier를 생성한다.
// 테스트 전용 함수이다.
func newApplierWithWritesEnabled() *Applier {
	return &Applier{allowWrites: true}
}

// EnrichDescription은 SKILL.md의 description 필드에 heuristicNote를 추가한다.
// REQ-HL-003: description 필드만 수정하며 다른 frontmatter와 body는 보존된다.
// 이미 동일 노트가 있으면 idempotent하게 처리한다 (중복 추가 없음).
//
// heuristicNote는 "# heuristic: <note>" 형식으로 description에 추가된다.
func (a *Applier) EnrichDescription(skillPath, heuristicNote string) error {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("applier: 파일 읽기 실패 %s: %w", skillPath, err)
	}

	fm, body, err := splitFrontmatterBody(string(content))
	if err != nil {
		return fmt.Errorf("applier: frontmatter 파싱 실패 %s: %w", skillPath, err)
	}

	// description 필드 찾기 및 수정
	newFM, changed := enrichDescriptionInFrontmatter(fm, heuristicNote)
	if !changed {
		// 변경 없음 (이미 해당 노트 포함) — idempotent
		return nil
	}

	// 재결합
	newContent := "---\n" + newFM + "---\n" + body

	if err := os.WriteFile(skillPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("applier: 파일 쓰기 실패 %s: %w", skillPath, err)
	}
	return nil
}

// InjectTrigger는 SKILL.md의 triggers 목록에 keyword를 추가한다.
// REQ-HL-004: 중복 키워드는 추가하지 않는다 (dedup).
//
// @MX:WARN: [AUTO] enableTriggerInjectionWrites가 OFF(Phase 2)면 실제 파일 write 생략.
// @MX:REASON: [AUTO] Phase 4 이전에는 파일 변경 없이 dedup 로직만 검증한다.
func (a *Applier) InjectTrigger(skillPath, keyword string) error {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return fmt.Errorf("applier: 파일 읽기 실패 %s: %w", skillPath, err)
	}

	fm, body, err := splitFrontmatterBody(string(content))
	if err != nil {
		return fmt.Errorf("applier: frontmatter 파싱 실패 %s: %w", skillPath, err)
	}

	// dedup: 이미 존재하는 키워드인지 확인
	newFM, changed := injectTriggerInFrontmatter(fm, keyword)
	if !changed {
		// 이미 존재하거나 변경 없음
		return nil
	}

	// feature flag 확인 — OFF이면 실제 write 생략 (Phase 2 gate)
	if !a.allowWrites {
		return nil
	}

	// 실제 파일 쓰기 (Phase 4에서 config로 활성화)
	newContent := "---\n" + newFM + "---\n" + body
	if err := os.WriteFile(skillPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("applier: 파일 쓰기 실패 %s: %w", skillPath, err)
	}
	return nil
}

// ─────────────────────────────────────────────
// Phase 4: Apply() — snapshot + safety pipeline 통합
// ─────────────────────────────────────────────

// SafetyEvaluator는 safety pipeline의 Evaluate 메서드 인터페이스이다.
// 순환 임포트 방지: harness → safety 직접 임포트 불가.
// safety.Pipeline이 이 인터페이스를 구현한다.
type SafetyEvaluator interface {
	Evaluate(proposal Proposal, sessions []Session) (Decision, error)
}

// ApplyPendingError는 safety pipeline이 pending_approval을 반환할 때 발생하는 오류이다.
// orchestrator(moai-harness-learner skill)가 이 오류를 받아 OversightProposal을
// AskUserQuestion으로 사용자에게 제시한다.
//
// @MX:ANCHOR: [AUTO] ApplyPendingError는 subagent→orchestrator 경계 타입이다.
// @MX:REASON: [AUTO] fan_in >= 3: applier.go, applier_test.go, harness CLI apply, moai-harness-learner skill
type ApplyPendingError struct {
	// OversightPayload는 orchestrator가 AskUserQuestion에 사용할 페이로드이다.
	OversightPayload *OversightProposal
}

func (e *ApplyPendingError) Error() string {
	if e.OversightPayload != nil {
		return fmt.Sprintf("apply: 사용자 승인 대기 중 (proposal_id=%s)", e.OversightPayload.ProposalID)
	}
	return "apply: 사용자 승인 대기 중"
}

// snapshotManifest는 snapshot 디렉토리의 manifest.json 스키마이다.
type snapshotManifest struct {
	// ProposalID는 이 스냅샷을 생성한 제안 ID이다.
	ProposalID string `json:"proposal_id"`

	// CreatedAt은 스냅샷 생성 시각 (UTC).
	CreatedAt time.Time `json:"created_at"`

	// Files는 백업된 파일 목록이다.
	Files []snapshotFile `json:"files"`
}

// snapshotFile은 단일 백업 파일 정보이다.
type snapshotFile struct {
	// OriginalPath는 원본 파일 경로이다.
	OriginalPath string `json:"original_path"`

	// BackupName은 스냅샷 디렉토리 내 백업 파일명이다.
	BackupName string `json:"backup_name"`
}

// Apply는 Proposal을 safety pipeline 평가 후 안전하게 적용한다.
// [HARD] 반드시 evaluator.Evaluate()를 먼저 호출하고, 거부 시 즉시 반환한다.
// [HARD] 스냅샷은 파일 write보다 먼저 생성되어야 한다. 스냅샷 실패 시 write 중단.
//
// evaluator는 SafetyEvaluator 인터페이스(safety.Pipeline이 구현)이다.
// snapshotBase는 ".moai/harness/learning-history/snapshots/" 형식의 기본 경로이다.
// sessions는 L2 canary check에 사용되는 최근 세션 목록이다.
//
// @MX:ANCHOR: [AUTO] Apply는 Phase 4 학습 적용 파이프라인의 단일 진입점이다.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, harness CLI apply, moai-harness-learner skill
func (a *Applier) Apply(proposal Proposal, evaluator SafetyEvaluator, snapshotBase string, sessions []Session) error {
	// ── Step 1: Safety Pipeline 평가 ─────────────────────────────────────────
	// [HARD] Frozen Guard를 포함한 5-Layer를 반드시 통과해야 한다.
	decision, err := evaluator.Evaluate(proposal, sessions)
	if err != nil {
		return fmt.Errorf("applier: safety pipeline 평가 오류: %w", err)
	}

	switch decision.Kind {
	case DecisionRejected:
		return fmt.Errorf("applier: 제안 거부됨 (L%d, rejected)", decision.RejectedBy)

	case DecisionPendingApproval:
		// [HARD] subagent는 AskUserQuestion을 직접 호출하지 않는다.
		// orchestrator에게 payload를 반환하여 사용자 승인을 위임한다.
		return &ApplyPendingError{OversightPayload: decision.OversightProposal}

	case DecisionApproved:
		// approved — 계속 진행
	}

	// ── Step 2: Snapshot 생성 (write보다 먼저) ───────────────────────────────
	// [HARD] snapshot 실패 시 write를 중단한다.
	if err := a.createSnapshot(proposal, snapshotBase); err != nil {
		return fmt.Errorf("applier: snapshot 생성 실패 — write 중단: %w", err)
	}

	// ── Step 3: 실제 파일 수정 ───────────────────────────────────────────────
	switch proposal.FieldKey {
	case "description":
		return a.EnrichDescription(proposal.TargetPath, proposal.NewValue)
	case "triggers":
		// 쓰기 활성화된 Applier로 InjectTrigger 수행
		w := newApplierWithWritesEnabled()
		return w.InjectTrigger(proposal.TargetPath, proposal.NewValue)
	default:
		return fmt.Errorf("applier: 지원하지 않는 fieldKey %q", proposal.FieldKey)
	}
}

// createSnapshot은 proposal.TargetPath의 현재 내용을 snapshotBase/<ISO-DATE>/ 에 백업한다.
// manifest.json을 생성한 후 파일 복사를 수행한다.
func (a *Applier) createSnapshot(proposal Proposal, snapshotBase string) error {
	// ISO-DATE 형식 디렉토리명 생성 (날짜 + nano 충돌 방지)
	now := time.Now().UTC()
	dirName := now.Format("2006-01-02T15-04-05.000000000Z")
	snapshotDir := filepath.Join(snapshotBase, dirName)

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return fmt.Errorf("createSnapshot: 디렉토리 생성 실패 %s: %w", snapshotDir, err)
	}

	// 원본 파일 읽기
	originalData, err := os.ReadFile(proposal.TargetPath)
	if err != nil {
		return fmt.Errorf("createSnapshot: 원본 파일 읽기 실패 %s: %w", proposal.TargetPath, err)
	}

	// 백업 파일명: 원본 파일명 그대로 사용
	backupName := filepath.Base(proposal.TargetPath)
	backupPath := filepath.Join(snapshotDir, backupName)

	if err := os.WriteFile(backupPath, originalData, 0o644); err != nil {
		return fmt.Errorf("createSnapshot: 백업 파일 쓰기 실패 %s: %w", backupPath, err)
	}

	// manifest.json 생성
	manifest := snapshotManifest{
		ProposalID: proposal.ID,
		CreatedAt:  now,
		Files: []snapshotFile{
			{
				OriginalPath: proposal.TargetPath,
				BackupName:   backupName,
			},
		},
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("createSnapshot: manifest 직렬화 실패: %w", err)
	}

	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		return fmt.Errorf("createSnapshot: manifest 쓰기 실패: %w", err)
	}

	return nil
}

// RestoreSnapshot은 snapshotDir의 manifest.json을 읽어 원본 파일을 복원한다.
// REQ-HL-009: rollback <date> verb에서 사용된다.
//
// @MX:ANCHOR: [AUTO] RestoreSnapshot은 rollback 기능의 핵심 함수이다.
// @MX:REASON: [AUTO] fan_in >= 3: applier_test.go, harness CLI rollback, Phase 5 IT
func RestoreSnapshot(snapshotDir string) error {
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("RestoreSnapshot: manifest.json 읽기 실패 %s: %w", manifestPath, err)
	}

	var manifest snapshotManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("RestoreSnapshot: manifest 파싱 실패: %w", err)
	}

	for _, f := range manifest.Files {
		backupPath := filepath.Join(snapshotDir, f.BackupName)
		backupData, err := os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("RestoreSnapshot: 백업 파일 읽기 실패 %s: %w", backupPath, err)
		}

		// 원본 경로에 복원
		if err := os.WriteFile(f.OriginalPath, backupData, 0o644); err != nil {
			return fmt.Errorf("RestoreSnapshot: 원본 파일 복원 실패 %s: %w", f.OriginalPath, err)
		}
	}

	return nil
}

// ─────────────────────────────────────────────
// 내부 헬퍼: frontmatter 파싱 및 수정
// ─────────────────────────────────────────────

// splitFrontmatterBody는 SKILL.md 내용을 frontmatter와 body로 분리한다.
// frontmatter는 --- 구분자 안의 내용이고, body는 두 번째 --- 이후이다.
// frontmatter가 없으면 오류를 반환한다.
func splitFrontmatterBody(content string) (fm, body string, err error) {
	// 개행 통일 (CRLF → LF)
	content = strings.ReplaceAll(content, "\r\n", "\n")

	const sep = "---"

	// 첫 번째 줄이 ---로 시작해야 함
	if !strings.HasPrefix(content, sep+"\n") && content != sep {
		return "", "", fmt.Errorf("frontmatter 시작 구분자 없음")
	}

	// 첫 번째 --- 제거 후 두 번째 ---를 찾는다
	rest := content[len(sep)+1:] // "---\n" 이후

	idx := strings.Index(rest, "\n"+sep+"\n")
	if idx == -1 {
		// 끝에 ---만 있는 경우 ("---\n" 없이 파일 끝)
		idx = strings.Index(rest, "\n"+sep)
		if idx == -1 {
			return "", "", fmt.Errorf("frontmatter 종료 구분자 없음")
		}
		fm = rest[:idx+1]  // '\n' 포함
		body = ""
		return fm, body, nil
	}

	fm = rest[:idx+1]               // '\n' 포함
	body = rest[idx+1+len(sep)+1:]  // "---\n" 이후 body
	return fm, body, nil
}

// enrichDescriptionInFrontmatter는 frontmatter YAML 텍스트에서 description 필드에
// "# heuristic: <note>"를 추가한다. 이미 존재하면 changed=false를 반환한다.
// 줄 기반 파싱을 사용하여 다른 필드를 보존한다.
func enrichDescriptionInFrontmatter(fm, heuristicNote string) (newFM string, changed bool) {
	targetLine := "# heuristic: " + heuristicNote
	lines := strings.Split(fm, "\n")

	// 이미 존재하는지 확인 (idempotent)
	for _, line := range lines {
		if strings.Contains(line, targetLine) {
			return fm, false
		}
	}

	// description 필드를 찾아 수정
	var result []string
	inDescription := false
	descModified := false

	for i, line := range lines {
		// description: 로 시작하는 줄 탐지
		if !descModified && strings.HasPrefix(strings.TrimLeft(line, " \t"), "description:") {
			trimmed := strings.TrimLeft(line, " \t")
			indent := line[:len(line)-len(trimmed)]

			// description: value (단일 라인)
			after := strings.TrimPrefix(trimmed, "description:")
			after = strings.TrimLeft(after, " ")

			if after == "" || after == "|" || after == "|-" || after == "|+" {
				// 블록 스칼라 — 이 케이스는 단순 처리 불가, 줄 뒤에 추가
				result = append(result, line)
				inDescription = true
			} else {
				// 인라인 값: description: original value
				result = append(result, line)
				// 다음 줄에 heuristic note 삽입 (동일 indent, 줄 연속)
				// 단순하게 description 값에 "\n# heuristic: ..." 를 append
				// 그러나 단일 라인 YAML이므로 multi-line 블록으로 변환 필요
				// 더 단순한 방법: description 뒤 바로 다음 줄에 삽입
				_ = i
				result = append(result, indent+"# heuristic: "+heuristicNote)
				descModified = true
			}
			continue
		}

		if inDescription {
			// description 블록 내부
			if line == "" || (!strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")) {
				// 블록 종료 — heuristic note 삽입 후 현재 줄 추가
				result = append(result, "# heuristic: "+heuristicNote)
				result = append(result, line)
				inDescription = false
				descModified = true
				continue
			}
		}
		result = append(result, line)
	}

	if !descModified {
		return fm, false
	}

	return strings.Join(result, "\n"), true
}

// injectTriggerInFrontmatter는 frontmatter의 triggers 목록에 keyword를 추가한다.
// 이미 존재하면 changed=false를 반환한다.
// triggers 필드가 없으면 추가하지 않고 changed=false를 반환한다.
func injectTriggerInFrontmatter(fm, keyword string) (newFM string, changed bool) {
	targetEntry := `keyword: "` + keyword + `"`

	// 이미 존재하는지 확인
	if strings.Contains(fm, targetEntry) {
		return fm, false
	}

	lines := strings.Split(fm, "\n")
	var result []string
	triggersFound := false
	lastTriggerIdx := -1

	// triggers: 섹션과 마지막 trigger 항목 위치를 찾는다
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "triggers:") {
			triggersFound = true
		}
		if triggersFound && strings.Contains(line, `keyword:`) {
			lastTriggerIdx = i
		}
	}

	if !triggersFound || lastTriggerIdx == -1 {
		// triggers 섹션 없음 — 변경 없이 반환
		return fm, false
	}

	// lastTriggerIdx 줄의 들여쓰기 수준을 참고하여 새 항목 삽입
	lastLine := lines[lastTriggerIdx]
	lastTrimmed := strings.TrimLeft(lastLine, " \t")
	indent := lastLine[:len(lastLine)-len(lastTrimmed)]

	// 마지막 trigger 항목 바로 다음에 새 항목 삽입
	for i, line := range lines {
		result = append(result, line)
		if i == lastTriggerIdx {
			result = append(result, indent+`keyword: "`+keyword+`"`)
		}
	}

	return strings.Join(result, "\n"), true
}
