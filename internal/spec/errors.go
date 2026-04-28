package spec

import "fmt"

// DuplicateAcceptanceID는 동일한 깊이에서 중복된 ID가 발견되었을 때 반환됩니다.
type DuplicateAcceptanceID struct {
	ID    string
	Depth int
}

func (e *DuplicateAcceptanceID) Error() string {
	return fmt.Sprintf("duplicate acceptance criteria ID: %s at depth %d", e.ID, e.Depth)
}

// MaxDepthExceeded는 최대 허용 깊이(3)를 초과했을 때 반환됩니다.
type MaxDepthExceeded struct {
	ID    string
	Depth int
	Max   int
}

func (e *MaxDepthExceeded) Error() string {
	return fmt.Sprintf("max depth exceeded: %s has depth %d, maximum allowed is %d", e.ID, e.Depth, e.Max)
}

// DanglingRequirementReference는 존재하지 않는 REQ를 참조할 때 반환됩니다.
// 이는 경고이지 치명적 오류가 아닙니다.
type DanglingRequirementReference struct {
	ACID     string
	ReqID    string
	Location string
}

func (e *DanglingRequirementReference) Error() string {
	return fmt.Sprintf("dangling requirement reference: %s references non-existent REQ-%s in %s", e.ACID, e.ReqID, e.Location)
}

// MissingRequirementMapping은 리프 노드에 REQ 맵핑이 누락되었을 때 반환됩니다.
type MissingRequirementMapping struct {
	ACID string
}

func (e *MissingRequirementMapping) Error() string {
	return fmt.Sprintf("missing requirement mapping: leaf node %s has no (maps REQ-...) tail", e.ACID)
}
