// backlog_issuance.go — SPEC-TODO-CARD-ISSUANCE-001 M1: the read-only
// issuance presentation lookup. Deliberately separate from the duplicate
// classifier (backlog_analysis.go): the classifier skips dropped cards
// because a dropped card is not a refusal subject, while the presentation
// WANTS dropped and archived cards as notice targets. Only the measures
// (NormalizeCardText, TokenSetJaccard) are shared — ClassifyCardText is
// never called here (REQ-TCI-006; AC-TCI-007 pins the classifier file
// byte-unchanged).
//
// M1 RED stub: the functions exist so the acceptance tests compile, and
// every one returns its zero value so the tests fail on assertions rather
// than on a missing symbol. The GREEN commit fills them in.
package factory

// IssuanceNeighborLimit — baseline threshold table, similar-card cap
// (card-fixed value 3).
const IssuanceNeighborLimit = 3

// IssuanceDisplayFloor — baseline threshold table, token-set Jaccard
// display floor (M0 remeasurement keeps the 0.30 working default).
const IssuanceDisplayFloor = 0.30

// IssuanceComponentDepth — baseline threshold table: a component key is the
// first two segments of a repo path (depth-3 keys shared 0 pairs, SB07).
const IssuanceComponentDepth = 2

type IssuanceNeighbor struct {
	ID     string
	State  string
	Score  float64
	Text   string
	Reason string
	Exact  bool
}

func IssuanceNeighbors(candidate string, items []BacklogItem, archived []BacklogArchiveEntry) []IssuanceNeighbor {
	return nil
}

type IssuanceComponent struct {
	ID    string
	State string
	Key   string
}

func IssuanceSameComponent(candidate string, items []BacklogItem) []IssuanceComponent {
	return nil
}

type IssuanceCompletedSpec struct {
	ID     string
	Status string
	Module string
	Title  string
	Tags   string
}

type IssuanceSpecMatch struct {
	ID     string
	Status string
}

func IssuanceCompletedSpecCoverage(candidate string, specs []IssuanceCompletedSpec) []IssuanceSpecMatch {
	return nil
}

type IssuanceInFlightCard struct {
	ID    string
	Lane  string
	Files []string
}

type IssuanceOverlapItem struct {
	CardID string
	Lane   string
	Path   string
}

type IssuanceOverlap struct {
	Items      []IssuanceOverlapItem
	None       bool
	Unmeasured string
}

func IssuanceInFlightOverlap(candidateFiles []string, inFlight []IssuanceInFlightCard) IssuanceOverlap {
	return IssuanceOverlap{}
}
