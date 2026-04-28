package spec

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseResult은 파싱 결과를 포함합니다.
// Errors는 치명적 오류, Warnings는 경고입니다.
type ParseResult struct {
	Criteria []Acceptance
	Errors   []error
	Warnings []error
}

// ParseAcceptanceCriteria는 SPEC 마크다운에서 Acceptance Criteria 섹션을 파싱합니다.
//
// 지원되는 기능:
//   - 들여쓰기 기반 계층 구조 (2스페이스 = 1레벨)
//   - 자동 Given 상속
//   - 중복 ID 검출
//   - 최대 깊이 검증 (MaxDepth)
//   - Flat 형식 자동 래핑 (최상위 자식 없는 AC를 1자식 부모로 변환)
//   - REQ 맵핑 추출 ((maps REQ-XXX) 패턴)
func ParseAcceptanceCriteria(markdown string, isFlatFormat bool) ([]Acceptance, []error) {
	result := parseAcceptanceCriteriaInternal(markdown, isFlatFormat)
	return result.Criteria, append(result.Errors, result.Warnings...)
}

// ParseAcceptanceCriteriaTyped는 파싱 결과를 Errors와 Warnings로 분리하여 반환합니다.
func ParseAcceptanceCriteriaTyped(markdown string, isFlatFormat bool) *ParseResult {
	return parseAcceptanceCriteriaInternal(markdown, isFlatFormat)
}

func parseAcceptanceCriteriaInternal(markdown string, isFlatFormat bool) *ParseResult {
	result := &ParseResult{}

	lines := strings.Split(markdown, "\n")

	// Acceptance Criteria 섹션 찾기
	startIdx := findACSectionStart(lines)
	if startIdx < 0 {
		result.Errors = append(result.Errors, fmt.Errorf("acceptance criteria section not found"))
		return result
	}

	// AC 라인 추출
	acLines := extractACLines(lines, startIdx, isFlatFormat)

	if len(acLines) == 0 {
		return result
	}

	// 스택 기반 트리 구성
	result.Criteria = buildTree(acLines, isFlatFormat, result)

	// 깊이 검증
	for i := range result.Criteria {
		if err := result.Criteria[i].ValidateDepth(); err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	return result
}

// acParsedLine은 파싱된 AC 라인 정보를 나타냅니다.
type acParsedLine struct {
	id      string
	given   string
	when    string
	then    string
	reqIDs  []string
	indent  int // 들여쓰기 레벨 (0, 1, 2, ...)
}

// findACSectionStart는 마크다운에서 Acceptance Criteria 섹션 시작 인덱스를 찾습니다.
func findACSectionStart(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "##") && strings.Contains(strings.ToLower(trimmed), "acceptance") {
			return i + 1
		}
	}
	return -1
}

// extractACLines는 AC 섹션에서 파싱된 라인 목록을 추출합니다.
func extractACLines(lines []string, startIdx int, isFlatFormat bool) []acParsedLine {
	var acLines []acParsedLine

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// 빈 라인 무시
		if trimmed == "" {
			continue
		}

		// 다른 섹션 시작하면 중단
		if strings.HasPrefix(trimmed, "##") {
			break
		}

		// AC 라인 파싱
		parsed := parseSingleACLine(trimmed)
		if parsed == nil {
			continue
		}

		// 들여쓰기 계산
		indent := calculateIndentationDepth(line)

		// Flat 형식이면 모든 라인을 indent 0으로 처리
		if isFlatFormat {
			indent = 0
		}

		acLines = append(acLines, acParsedLine{
			id:     parsed.id,
			given:  parsed.given,
			when:   parsed.when,
			then:   parsed.then,
			reqIDs: parsed.reqIDs,
			indent: indent,
		})
	}

	return acLines
}

// buildTree는 파싱된 라인 목록을 계층 구조로 변환합니다.
func buildTree(acLines []acParsedLine, _ bool, result *ParseResult) []Acceptance {
	// 스택: [root_level_node, level_1_node, ...]
	type stackEntry struct {
		node     *Acceptance
		indent   int
		lineIdx  int
	}

	var roots []Acceptance
	var stack []stackEntry
	seenIDs := make(map[string]bool)

	for i, acLine := range acLines {
		node := Acceptance{
			ID:             acLine.id,
			Given:          acLine.given,
			When:           acLine.when,
			Then:           acLine.then,
			RequirementIDs: acLine.reqIDs,
		}

		// 중복 ID 검출
		if seenIDs[acLine.id] {
			result.Errors = append(result.Errors, &DuplicateAcceptanceID{
				ID:    acLine.id,
				Depth: acLine.indent,
			})
			continue
		}
		seenIDs[acLine.id] = true

		// 스택에서 현재 indent보다 깊은 항목 제거
		for len(stack) > 0 && stack[len(stack)-1].indent >= acLine.indent {
			stack = stack[:len(stack)-1]
		}

		if acLine.indent == 0 || len(stack) == 0 {
			// 최상위 노드
			roots = append(roots, node)
			stack = append(stack, stackEntry{
				node:    &roots[len(roots)-1],
				indent:  acLine.indent,
				lineIdx: i,
			})
		} else {
			// 부모의 자식으로 추가
			parent := stack[len(stack)-1].node
			parent.Children = append(parent.Children, node)

			// Given 상속: 자식의 Given이 비어있으면 부모의 Given 상속
			child := &parent.Children[len(parent.Children)-1]
			if child.Given == "" && parent.Given != "" {
				child.Given = parent.Given
			}

			stack = append(stack, stackEntry{
				node:    child,
				indent:  acLine.indent,
				lineIdx: i,
			})
		}
	}

	// 자동 래핑: 최상위 노드 중 자식이 없고 ID에 접미사가 없는 것만
	for i := range roots {
		if len(roots[i].Children) == 0 && !hasIDSuffix(roots[i].ID) {
			roots[i] = autoWrapSingle(roots[i])
		}
	}

	return roots
}

// hasIDSuffix는 AC ID에 접미사(.a, .i 등)가 있는지 확인합니다.
func hasIDSuffix(id string) bool {
	return strings.Contains(id, ".")
}

// autoWrapSingle는 자식이 없는 최상위 AC를 1자식 부모로 래핑합니다.
func autoWrapSingle(ac Acceptance) Acceptance {
	childID := ac.ID + ".a"
	child := Acceptance{
		ID:             childID,
		Given:          ac.Given,
		When:           ac.When,
		Then:           ac.Then,
		RequirementIDs: ac.RequirementIDs,
	}
	wrapped := Acceptance{
		ID:       ac.ID,
		Children: []Acceptance{child},
	}
	return wrapped
}

// parseSingleACLine는 단일 AC 라인을 파싱합니다.
func parseSingleACLine(line string) *struct {
	id     string
	given  string
	when   string
	then   string
	reqIDs []string
} {
	trimmed := strings.TrimSpace(line)

	// 리스트 마커 제거 (- , * )
	trimmed = strings.TrimLeft(trimmed, "- *")
	trimmed = strings.TrimSpace(trimmed)

	// AC ID 패턴: AC-XXX-NNN-NN 또는 AC-XXX-NNN-NN.a 또는 AC-XXX-NNN-NN.a.i
	acIDPattern := regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)
	idMatch := acIDPattern.FindStringSubmatch(trimmed)

	if len(idMatch) < 2 {
		return nil
	}

	id := idMatch[1]
	content := strings.TrimSpace(trimmed[len(idMatch[0]):])

	// REQ 맵핑 추출
	reqIDs := ExtractRequirementMappings(content)

	// REQ 맵핑 부분 제거
	reqRemover := regexp.MustCompile(`\(?\s*(?:maps|MAPS)\s+REQ-[A-Z0-9-]+\s*\)?`)
	cleanContent := strings.TrimSpace(reqRemover.ReplaceAllString(content, ""))

	// EARS 패턴 파싱: Given ... When ... Then ...
	var given, when, then string

	// Given 추출
	givenRe := regexp.MustCompile(`(?i)^Given\s+(.+?)(?:,\s*(?:When|then)|$)`)
	if match := givenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		given = "Given " + strings.TrimSpace(match[1])
		cleanContent = strings.TrimSpace(cleanContent[len(match[0]):])
	}

	// When 추출
	whenRe := regexp.MustCompile(`(?i)^When\s+(.+?)(?:,\s*(?:Then|then)|$)`)
	if match := whenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		when = "When " + strings.TrimSpace(match[1])
		cleanContent = strings.TrimSpace(cleanContent[len(match[0]):])
	}

	// Then 추출
	thenRe := regexp.MustCompile(`(?i)^Then\s+(.+)`)
	if match := thenRe.FindStringSubmatch(cleanContent); len(match) > 1 {
		then = "Then " + strings.TrimSpace(match[1])
	}

	// 패턴이 매칭되지 않으면 전체를 Then으로 처리
	if when == "" && then == "" && given == "" {
		then = cleanContent
	}

	return &struct {
		id     string
		given  string
		when   string
		then   string
		reqIDs []string
	}{
		id:     id,
		given:  given,
		when:   when,
		then:   then,
		reqIDs: reqIDs,
	}
}

// calculateIndentationDepth는 들여쓰기 기반으로 깊이를 계산합니다.
// 2스페이스 = 1레벨
func calculateIndentationDepth(line string) int {
	spaces := 0
	for _, r := range line {
		switch r {
		case ' ':
			spaces++
		case '\t':
			spaces += 2
		default:
			// goto out would be cleaner but switch-break is fine
		}
		if r != ' ' && r != '\t' {
			break
		}
	}
	return spaces / 2
}

// CheckDanglingReferences는 존재하지 않는 REQ 참조를 검사합니다.
func CheckDanglingReferences(criteria []Acceptance, existingREQs map[string]bool) []error {
	var errors []error

	var check func(ac *Acceptance)
	check = func(ac *Acceptance) {
		for _, reqID := range ac.RequirementIDs {
			if !existingREQs[reqID] {
				errors = append(errors, &DanglingRequirementReference{
					ACID:     ac.ID,
					ReqID:    reqID,
					Location: ac.ID,
				})
			}
		}
		for i := range ac.Children {
			check(&ac.Children[i])
		}
	}

	for i := range criteria {
		check(&criteria[i])
	}

	return errors
}
