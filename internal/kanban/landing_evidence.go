// landing_evidence.go — the landing-evidence record and its encoding
// (SPEC-TODO-LANDING-EVIDENCE-001 REQ-TLE-005, REQ-TLE-006, REQ-TLE-012,
// REQ-TLE-013, M2).
//
// This file exists to keep ONE distinction structural rather than
// conventional: an operator's assertion about which commit delivered a card,
// and a machine's observation of where a ref stood. Both are SHAs, they look
// identical in a rendered row, and the failure mode of confusing them is
// silent — a record that names the wrong delivering commit is still a
// well-formed record and reads as authoritative forever after.
//
// `SPEC-KANBAN-QUEUE-PR-SYNC-001` REQ-1.10 rules that the landed resolver
// names no delivering commit, because the grep predicate finds commits that
// MENTION a card and a mention is not a delivery. This SPEC's REQ-TLE-012
// restates the prohibition at the storage layer, and REQ-TLE-013 requires the
// ref position be carried AS a ref position. The three keys below —
// `ref_head`, `sha`, `sha_source` — are how that survives a round trip: the
// observation and the assertion never share a key, and the assertion cannot
// be stored without its provenance, because the encoder refuses the pairing
// that would let one masquerade as the other.
//
// The record is a pure value. It reads no configuration, spawns no process,
// and knows nothing about a card: the verb (M3) supplies every field, and
// this file's only job is to refuse the shapes the requirements forbid.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// The record's wire keys, as named constants rather than as bare tags.
//
// A test asking "are the ref position and the delivering SHA under different
// keys?" must ask it of the implementation's own names; a test that
// transcribes the strings would still pass after the two were collapsed in
// the source. These are the symbols AC-TLE-013 keys on.
const (
	// LandingKeyRef is the ref the observation was made against.
	LandingKeyRef = "ref"
	// LandingKeyRefHead is the OBSERVED REF POSITION — where that ref stood
	// at the observation instant. It is never a delivery claim (REQ-TLE-013).
	LandingKeyRefHead = "ref_head"
	// LandingKeyObservedAt is the observation instant.
	LandingKeyObservedAt = "observed_at"
	// LandingKeyDeliveringSHA is the OPERATOR-ASSERTED delivering commit. It
	// is present only when the operator supplied one (REQ-TLE-012), and it is
	// deliberately a different key from LandingKeyRefHead.
	LandingKeyDeliveringSHA = "sha"
	// LandingKeySHASource is the provenance of LandingKeyDeliveringSHA, and
	// the field a machine keys on to tell an assertion from an observation.
	LandingKeySHASource = "sha_source"
	// LandingKeySpecStatus is the SPEC frontmatter status read at record time.
	LandingKeySpecStatus = "spec_status"
)

const (
	// LandingSHASourceOperator is the ONLY provenance this SPEC defines. The
	// machine has no other lawful source for a delivering SHA — deriving one
	// from the card-token grep predicate is what REQ-TLE-012 forbids — so the
	// value set is closed at one rather than left open for a future
	// "inferred" that would reintroduce the hazard by the back door.
	LandingSHASourceOperator = "operator"

	// LandingMarkerRefHead labels a record that carries only an observed ref
	// position. It is the counterpart of LandingSHASourceOperator, so a
	// reader always sees WHICH of the two a rendered SHA is.
	LandingMarkerRefHead = "ref-head"

	// LandingSpecStatusUnknown is the explicit marker for a SPEC status that
	// could not be read. It is a stored VALUE, never an omitted key: an
	// omitted key means the question was never asked, and REQ-TLE-010 needs
	// those two facts to stay apart.
	LandingSpecStatusUnknown = "unknown"
)

// ErrLandingEvidenceAbsent is returned when a stored value carries no record.
// Absence is SQL NULL (REQ-TLE-006), so this is what a caller sees when it
// asks a column that was never written — distinct from a decode failure on a
// value that IS present and malformed.
var ErrLandingEvidenceAbsent = errors.New("kanban: no landing evidence recorded")

// LandingEvidence is one operator-recorded landing observation (REQ-TLE-005).
//
// Six facts, and the split between the last three is the point:
//
//	Ref, RefHead, ObservedAt — what the MACHINE observed, and when.
//	SHA, SHASource           — what the OPERATOR asserted, and that they did.
//	SpecStatus               — what the machine read at record time, or the
//	                           explicit unknown marker.
//
// RefHead and SHA are both commit SHAs and are NOT interchangeable. RefHead
// says "this is where the ref stood"; SHA says "the operator says this commit
// delivered the card". Nothing in this package can produce the second from
// the first, and nothing derives either from the landed grep predicate.
type LandingEvidence struct {
	// Ref is the ref the observation was made against, as resolved by
	// LandedRefFor. The answer is meaningless without it.
	Ref string `json:"ref"`
	// RefHead is that ref's head SHA at ObservedAt. A REF POSITION, so a
	// later reader can re-derive the history the observer saw — never a claim
	// about which commit delivered anything (REQ-TLE-013).
	RefHead string `json:"ref_head"`
	// ObservedAt is the observation instant, RFC 3339 UTC — the same shape
	// the queue already uses for added_at. A reader judges staleness from it.
	ObservedAt string `json:"observed_at"`
	// SHA is the delivering commit ON THE OPERATOR'S AUTHORITY, present only
	// when they supplied one. Empty is the normal case.
	SHA string `json:"sha,omitempty"`
	// SHASource is the provenance of SHA, present if and only if SHA is.
	// LandingSHASourceOperator is its only lawful value.
	SHASource string `json:"sha_source,omitempty"`
	// SpecStatus is the SPEC frontmatter status read at record time, or
	// LandingSpecStatusUnknown when it could not be read. Empty means the
	// card carries no spec_id and the question was never asked.
	SpecStatus string `json:"spec_status,omitempty"`
}

// Marker reports which kind of SHA a reader is looking at:
// LandingSHASourceOperator when the record carries an operator assertion,
// LandingMarkerRefHead when it carries only the observed ref position.
//
// It exists so the render surface labels the two rather than re-deriving the
// distinction from field emptiness — a reading that breaks the moment a
// caller forgets which field it is holding.
func (e LandingEvidence) Marker() string {
	if strings.TrimSpace(e.SHA) != "" {
		return LandingSHASourceOperator
	}
	return LandingMarkerRefHead
}

// Validate reports why a record may not be stored, or nil.
//
// The three observed facts are mandatory: a record missing its ref answers a
// question nobody can identify, one missing its head cannot be re-derived,
// and one missing its instant cannot be judged for staleness.
//
// The provenance pairing is enforced in BOTH directions and is the structural
// half of REQ-TLE-012: a SHA with no provenance would be stored with its
// origin lost, and a provenance with no SHA labels nothing. A SHA whose
// provenance is anything other than LandingSHASourceOperator is refused
// outright — that is precisely the shape a derived-from-grep value would take.
func (e LandingEvidence) Validate() error {
	if strings.TrimSpace(e.Ref) == "" {
		return fmt.Errorf("kanban: landing evidence has no %s", LandingKeyRef)
	}
	if strings.TrimSpace(e.RefHead) == "" {
		return fmt.Errorf("kanban: landing evidence has no %s", LandingKeyRefHead)
	}
	if strings.TrimSpace(e.ObservedAt) == "" {
		return fmt.Errorf("kanban: landing evidence has no %s", LandingKeyObservedAt)
	}
	hasSHA := strings.TrimSpace(e.SHA) != ""
	hasSource := strings.TrimSpace(e.SHASource) != ""
	switch {
	case hasSHA && !hasSource:
		return fmt.Errorf("kanban: landing evidence carries %s without %s; a stored delivering commit is operator-asserted or absent",
			LandingKeyDeliveringSHA, LandingKeySHASource)
	case hasSource && !hasSHA:
		return fmt.Errorf("kanban: landing evidence carries %s without %s; a provenance labels nothing on its own",
			LandingKeySHASource, LandingKeyDeliveringSHA)
	case hasSHA && e.SHASource != LandingSHASourceOperator:
		return fmt.Errorf("kanban: landing evidence %s = %q, want %q; the machine has no other lawful source for a delivering commit",
			LandingKeySHASource, e.SHASource, LandingSHASourceOperator)
	}
	return nil
}

// EncodeLandingEvidence renders a record for storage, refusing any shape the
// requirements forbid rather than storing it and leaving the refusal to a
// reader who may never come.
func EncodeLandingEvidence(e LandingEvidence) (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("kanban: encode landing evidence: %w", err)
	}
	return string(encoded), nil
}

// DecodeLandingEvidence reads a stored value back.
//
// An empty or unparseable value is an ERROR, never a zero-valued record: a
// record that decoded to blanks would render as an observation that found
// nothing, which is a different fact from a value that could not be read at
// all. The same reasoning the landed querier applies to its three-valued
// answer applies here.
func DecodeLandingEvidence(stored string) (LandingEvidence, error) {
	var e LandingEvidence
	if strings.TrimSpace(stored) == "" {
		return e, ErrLandingEvidenceAbsent
	}
	if err := json.Unmarshal([]byte(stored), &e); err != nil {
		return LandingEvidence{}, fmt.Errorf("kanban: decode landing evidence: %w", err)
	}
	if err := e.Validate(); err != nil {
		return LandingEvidence{}, fmt.Errorf("kanban: stored landing evidence is not a record: %w", err)
	}
	return e, nil
}

// LandingEvidenceValue converts an optional record into the value the driver
// stores: nil (SQL NULL) for no record, the encoded string for one.
//
// It is the single seam through which the column is written, so REQ-TLE-006's
// "absence is NULL" is a property of the type rather than a discipline each
// call site must remember. A caller holding no record cannot reach `{}` or
// `""` from here — those are the two shapes the requirement names, and both
// would render as a present-but-empty observation.
func LandingEvidenceValue(e *LandingEvidence) (any, error) {
	if e == nil {
		return nil, nil
	}
	encoded, err := EncodeLandingEvidence(*e)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}
