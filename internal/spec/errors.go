package spec

import "fmt"

type DuplicateAcceptanceID struct {
	ID    string
	Depth int
}

func (e *DuplicateAcceptanceID) Error() string {
	return fmt.Sprintf("duplicate acceptance criteria ID: %s at depth %d", e.ID, e.Depth)
}

type MaxDepthExceeded struct {
	ID    string
	Depth int
	Max   int
}

func (e *MaxDepthExceeded) Error() string {
	return fmt.Sprintf("max depth exceeded: %s has depth %d, maximum allowed is %d", e.ID, e.Depth, e.Max)
}

// DanglingRequirementReference is returned when referencing a non-existent REQ
type DanglingRequirementReference struct {
	ACID     string
	ReqID    string
	Location string
}

func (e *DanglingRequirementReference) Error() string {
	return fmt.Sprintf("dangling requirement reference: %s references non-existent REQ-%s in %s", e.ACID, e.ReqID, e.Location)
}

// MissingRequirementMapping is returned when leaf node is missing REQ mapping
type MissingRequirementMapping struct {
	ACID string
}

func (e *MissingRequirementMapping) Error() string {
	return fmt.Sprintf("missing requirement mapping: leaf node %s has no (maps REQ-...) tail", e.ACID)
}
