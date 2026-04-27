// Package harness — frontmatter 수정 applier.
// REQ-HL-003: description enrichment (Tier 2 heuristic).
// REQ-HL-004: trigger injection (Tier 3 rule, feature-gated).
package harness

import (
	"fmt"
	"os"
	"strings"
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
