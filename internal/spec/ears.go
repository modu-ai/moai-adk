package spec

import (
	"fmt"
	"regexp"
	"strings"
)

// Acceptance는 EARS 형식의 Acceptance Criteria를 나타냅니다.
// 계층 구조를 지원하며 최대 3레벨 깊이까지 허용됩니다.
type Acceptance struct {
	ID             string        // AC-<DOMAIN>-<NNN>-<NN> 형식 (예: AC-SPC-001-05)
	Given          string        // 전제 조건
	When           string        // 트리거 이벤트
	Then           string        // 예상 결과
	RequirementIDs []string      // 참조하는 REQ ID 목록 (예: ["SPC-001", "SPC-002"])
	Children       []Acceptance  // 하위 Acceptance Criteria (계층 구조)
}

// MaxDepth는 Acceptance 트리의 최대 허용 깊이입니다.
const MaxDepth = 3

// topLevelIDPattern은 최상위 AC ID 형식을 검증합니다.
// 예: AC-SPC-001-05
var topLevelIDPattern = regexp.MustCompile(`^AC-[A-Z0-9]+-[0-9]+-[0-9]+$`)

// ValidateID는 Acceptance ID가 유효한 형식인지 검증합니다.
func (a *Acceptance) ValidateID() error {
	if !topLevelIDPattern.MatchString(a.ID) {
		return fmt.Errorf("invalid acceptance criteria ID format: %s (expected AC-<DOMAIN>-<NNN>-<NN>)", a.ID)
	}
	return nil
}

// GenerateChildID는 현재 ID에서 자식 ID를 생성합니다.
// 레벨 1 자식: .a, .b, .c (소문자 알파벳)
// 레벨 2 자식: .a.i, .a.ii (소문자 로마 숫자)
func (a *Acceptance) GenerateChildID(depth int, index int) (string, error) {
	if depth < 1 || depth > 2 {
		return "", fmt.Errorf("invalid depth for child generation: %d (must be 1 or 2)", depth)
	}

	if depth == 1 {
		// 레벨 1: .a, .b, ..., .z
		if index < 0 || index > 25 {
			return "", fmt.Errorf("invalid index for level 1 child: %d (must be 0-25)", index)
		}
		suffix := string(rune('a' + index))
		return fmt.Sprintf("%s.%s", a.ID, suffix), nil
	}

	// 레벨 2: .a.i, .a.ii, ..., .a.xxvi
	romanNumerals := []string{
		"i", "ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
		"xi", "xii", "xiii", "xiv", "xv", "xvi", "xvii", "xviii", "xix", "xx",
		"xxi", "xxii", "xxiii", "xxiv", "xxv", "xxvi",
	}
	if index < 0 || index >= len(romanNumerals) {
		return "", fmt.Errorf("invalid index for level 2 child: %d (must be 0-25)", index)
	}

	// 레벨 2 자식은 부모 ID에 로마 숫자를 직접 추가
	return fmt.Sprintf("%s.%s", a.ID, romanNumerals[index]), nil
}

// Depth는 현재 노드의 깊이를 반환합니다.
// 최상위 레벨: 0, 레벨 1 자식: 1, 레벨 2 자식: 2
func (a *Acceptance) Depth() int {
	parts := strings.Split(a.ID, ".")
	// 최상위: AC-XXX-NNN-NN (parts 길이 1)
	// 레벨 1: AC-XXX-NNN-NN.a (parts 길이 2)
	// 레벨 2: AC-XXX-NNN-NN.a.i (parts 길이 3)
	return len(parts) - 1
}

// InheritGiven은 자식 노드가 Given이 비어있을 때 부모의 Given을 상속받습니다.
func (a *Acceptance) InheritGiven(parent *Acceptance) {
	if a.Given == "" && parent != nil && parent.Given != "" {
		a.Given = parent.Given
	}
}

// IsLeaf는 이 노드가 리프(자식이 없는) 노드인지 확인합니다.
func (a *Acceptance) IsLeaf() bool {
	return len(a.Children) == 0
}

// CountLeaves는 이 트리의 총 리프 노드 수를 반환합니다.
func (a *Acceptance) CountLeaves() int {
	if a.IsLeaf() {
		return 1
	}

	count := 0
	for i := range a.Children {
		count += a.Children[i].CountLeaves()
	}
	return count
}

// ValidateDepth는 트리의 깊이가 MaxDepth를 초과하지 않는지 검증합니다.
// 허용 깊이: 0 (최상위), 1 (.a), 2 (.a.i). depth >= MaxDepth(3) 이면 오류.
func (a *Acceptance) ValidateDepth() error {
	if a.Depth() >= MaxDepth {
		return &MaxDepthExceeded{
			ID:    a.ID,
			Depth: a.Depth(),
			Max:   MaxDepth - 1,
		}
	}

	for i := range a.Children {
		if err := a.Children[i].ValidateDepth(); err != nil {
			return err
		}
	}

	return nil
}

// ExtractRequirementMappings은 텍스트에서 (maps REQ-...) 패턴을 추출합니다.
func ExtractRequirementMappings(text string) []string {
	// 소문자 "maps"와 대문자 "MAPS" 모두 지원, 공백 유연하게 처리
	pattern := regexp.MustCompile(`\(?\s*(?:maps|MAPS)\s+REQ-([A-Z0-9-]+)\s*\)?`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	var reqIDs []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			reqID := match[1]
			if !seen[reqID] {
				seen[reqID] = true
				reqIDs = append(reqIDs, reqID)
			}
		}
	}

	return reqIDs
}

// ValidateRequirementMappings는 모든 리프 노드가 REQ 맵핑을 가지고 있는지 검증합니다.
// 리프가 아닌 부모 노드는 모든 자식이 맵핑을 가지고 있다면 생략할 수 있습니다.
func (a *Acceptance) ValidateRequirementMappings() []error {
	var errors []error

	a.validateRequirementMappingsRecursive(&errors)
	return errors
}

func (a *Acceptance) validateRequirementMappingsRecursive(errors *[]error) {
	if a.IsLeaf() {
		// 리프 노드는 반드시 REQ 맵핑이 있어야 함
		if len(a.RequirementIDs) == 0 {
			*errors = append(*errors, &MissingRequirementMapping{ACID: a.ID})
		}
	} else {
		// 부모 노드는 REQ 맵핑이 있어도 되고 없어도 됨
		// 하지만 자식들의 유효성은 검증해야 함
		for i := range a.Children {
			a.Children[i].validateRequirementMappingsRecursive(errors)
		}
	}
}
