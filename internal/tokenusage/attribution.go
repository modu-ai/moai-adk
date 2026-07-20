package tokenusage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 귀속 방법 (REQ-TA-005, REQ-TA-006).
const (
	// AttributionSessionSet — lineage가 가용하여 session-set 합산으로 귀속 (REQ-TA-005).
	AttributionSessionSet = "session-set"
	// AttributionCurrentSession — lineage 부재로 활성 세션 단독 측정 폴백 (REQ-TA-006).
	AttributionCurrentSession = "current-session"
)

// 신뢰도 한정자 (REQ-TA-007).
const (
	// ConfidenceHigh — 모든 기여 세션이 SPEC-전용 (다른 SPEC이 동일 UUID를 참조하지 않음).
	ConfidenceHigh = "high"
	// ConfidenceLow — 어느 기여 세션이라도 공유되거나 lineage가 부재.
	ConfidenceLow = "low"
)

// uuidPattern은 표준 8-4-4-4-12 hex UUID를 매칭한다.
var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// Attribution은 per-SPEC 토큰 귀속 결과이다. 기여 세션 집합 전체를 합산한 Usage에
// 귀속 방법, 신뢰도 한정자, 세션 수, 감사 추적용 기여 세션 UUID 목록을 포함한다
// (REQ-TA-005, REQ-TA-006, REQ-TA-007). JSON 태그는 progress.md §I 필드 스키마에
// 정합한다 (tokens_spent, token_attribution, token_attribution_confidence,
// token_session_count).
type Attribution struct {
	Usage                         // 포함된 합산 사용량 (필드 승격)
	AttributionMethod    string   `json:"token_attribution"`
	Confidence           string   `json:"token_attribution_confidence"`
	SessionCount         int      `json:"token_session_count"`
	ContributingSessions []string `json:"token_contributing_sessions,omitempty"`
}

// Attribute는 session-set 합산 방식으로 per-SPEC 토큰 귀속을 계산한다 (REQ-TA-005).
// progressPath에서 source_session_id UUID들을 파싱하고, 각 UUID를
// <transcriptRootDir>/<uuid>.jsonl로 해석하여 SumSession으로 합산한 뒤 결과를
// 집계한다.
//
// REQ-TA-006 (lineage 부재 폴백): source_session_id에서 실제 UUID를 찾지 못하면
// (environment-fallback "not-available" 토큰이거나 UUID가 0개) 활성 세션 단독 측정으로
// 폴백하고 token_attribution_confidence를 "low"로 기록한다.
//
// REQ-TA-007 (신뢰도 한정자): 모든 기여 세션이 SPEC-전용이면 "high", 어느 하나라도
// 다른 SPEC lineage에 공유되면 "low"로 기록하여 session-set 방식이 제공하는 정밀도
// 이상을 주장하지 않는다 (verification-claim-integrity §1.1).
//
// REQ-TA-013 (견고성): 집합 내 UUID의 transcript 파일이 부재하면 해당 세션은 0
// 기여로 skip-and-continue하며 panic이나 주변 audit 실행 중단 없이 진행한다.
//
// transcriptRootDir와 specsDir를 파라미터로 주입하는 것은 테스트가 t.TempDir()로
// 제어하기 위함이며, 실제 ~/.claude/projects/**를 건드리지 않는다.
//
// @MX:ANCHOR: [AUTO] per-SPEC 토큰 귀속 공개 진입점 — session-set 합산 + 신뢰도 판정 계약
// @MX:REASON: [AUTO] 공개 API 경계 + M3(§I writer)/M4(audit 표면) 다운스트림 통합 지점. 기여 세션 집합 합산, lineage-부재 폴백, 공유-UUID 강등 규칙이 이 함수의 불변계약이며 어기면 모든 per-SPEC 토큰 수치와 신뢰도 한정자가 오염됨
// @MX:SPEC: SPEC-TOKEN-ACCOUNTING-001
func Attribute(progressPath, transcriptRootDir, specsDir, activeSessionUUID string) (Attribution, error) {
	content, err := os.ReadFile(progressPath)
	if err != nil {
		return Attribution{}, fmt.Errorf("tokenusage: read progress %q: %w", progressPath, err)
	}

	lineageUUIDs := extractSessionUUIDs(string(content))
	lineageAvailable := len(lineageUUIDs) > 0

	// 세션 집합 구성
	var sessionSet []string
	var method string
	if lineageAvailable {
		method = AttributionSessionSet
		sessionSet = dedupeAppend(lineageUUIDs, activeSessionUUID)
	} else {
		// REQ-TA-006 폴백: lineage 부재 → 활성 세션 단독 측정
		method = AttributionCurrentSession
		if activeSessionUUID != "" {
			sessionSet = []string{activeSessionUUID}
		}
	}

	confidence := determineConfidence(sessionSet, lineageAvailable, specsDir, progressPath)

	// Transcript 합산 — M1 SumSession 재사용 (중복 금지)
	var agg Usage
	for _, uuid := range sessionSet {
		transcriptPath := filepath.Join(transcriptRootDir, uuid+".jsonl")
		u, err := SumSession(transcriptPath)
		if err != nil {
			// REQ-TA-013: 부재 파일 / 읽기 오류 → 0 기여, skip-and-continue
			continue
		}
		agg.TokensInput += u.TokensInput
		agg.TokensOutput += u.TokensOutput
		agg.TokensCacheCreation += u.TokensCacheCreation
		agg.TokensCacheRead += u.TokensCacheRead
	}
	agg.finalize() // M1 산술 재사용 — TokensSpent + CacheHitRatio

	return Attribution{
		Usage:                agg,
		AttributionMethod:    method,
		Confidence:           confidence,
		SessionCount:         len(sessionSet),
		ContributingSessions: sessionSet,
	}, nil
}

// determineConfidence는 REQ-TA-007에 따라 신뢰도 한정자를 판정한다.
//
// @MX:NOTE: [AUTO] 신뢰도 휴리스틱 — lineage 부재 시 environment-fallback이므로 측정 정밀도를 보장할 수 없어 강제 "low". 공유 UUID는 세션이 여러 SPEC에 인터리브되었음을 의미하여 과다계상 가능성 → "low" 강등. "high"는 모든 기여 세션이 SPEC-전용일 때만 부여
func determineConfidence(sessionSet []string, lineageAvailable bool, specsDir, currentProgressPath string) string {
	if !lineageAvailable {
		return ConfidenceLow
	}
	if anySharedSession(sessionSet, specsDir, currentProgressPath) {
		return ConfidenceLow
	}
	return ConfidenceHigh
}

// extractSessionUUIDs는 content에서 "source_session_id"가 포함된 라인의 UUID들을
// 추출한다. "not-available"이 포함된 라인(environment-fallback)은 건너뛴다.
// 중복 UUID는 한 번만 반환한다. 순서는 첫 등장 순서를 보존한다.
func extractSessionUUIDs(content string) []string {
	var uuids []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(content))
	// progress.md 라인은 보통 짧지만, 대형 SPEC의 source_session_id 라인이
	// 길어질 수 있으므로 스캐너 버퍼를 확장한다 (M1의 bufio.Reader 접근과 동일 동기).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "source_session_id") {
			continue
		}
		if strings.Contains(line, "not-available") {
			continue // environment-fallback 라인 — 실제 UUID 아님
		}
		for _, match := range uuidPattern.FindAllString(line, -1) {
			if !seen[match] {
				seen[match] = true
				uuids = append(uuids, match)
			}
		}
	}
	return uuids
}

// anySharedSession은 sessionSet의 UUID 중 하나라도 specsDir 하위의 다른 SPEC
// progress.md에 등장하면 true를 반환한다 (REQ-TA-007 "shared across SPECs").
// currentProgressPath와 동일한 파일은 자기 자신이므로 제외한다. specsDir를 읽을 수
// 없거나 다른 progress.md를 읽을 수 없으면 공유를 증명할 수 없으므로 false를
// 반환한다.
func anySharedSession(sessionSet []string, specsDir, currentProgressPath string) bool {
	if len(sessionSet) == 0 {
		return false
	}
	uuidSet := make(map[string]bool, len(sessionSet))
	for _, u := range sessionSet {
		uuidSet[u] = true
	}
	absCurrent, err := filepath.Abs(currentProgressPath)
	if err != nil {
		absCurrent = filepath.Clean(currentProgressPath)
	}
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return false // specsDir 읽기 불가 → 공유 증명 불가 → not shared
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		otherProgress := filepath.Join(specsDir, entry.Name(), "progress.md")
		absOther, err := filepath.Abs(otherProgress)
		if err != nil {
			absOther = filepath.Clean(otherProgress)
		}
		if absOther == absCurrent {
			continue // 자기 자신 제외
		}
		content, err := os.ReadFile(otherProgress)
		if err != nil {
			continue // 읽기 불가 → skip
		}
		for _, u := range extractSessionUUIDs(string(content)) {
			if uuidSet[u] {
				return true
			}
		}
	}
	return false
}

// dedupeAppend는 base에 items 중 중복되지 않고 비어있지 않은 항목만 추가한다.
// base의 기존 순서를 보존하고 items의 새 항목을 그 뒤에 append한다.
func dedupeAppend(base []string, items ...string) []string {
	seen := make(map[string]bool, len(base))
	for _, b := range base {
		seen[b] = true
	}
	for _, item := range items {
		if item != "" && !seen[item] {
			seen[item] = true
			base = append(base, item)
		}
	}
	return base
}
