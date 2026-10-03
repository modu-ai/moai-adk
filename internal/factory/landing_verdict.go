// landing_verdict.go — the done-time landing verdict record
// (SPEC-TODO-TRANSITION-STAMPS-001 REQ-TST-008/010/013, M3).
//
// This is a NEW record shape sitting ALONGSIDE — not inside — the
// operator-authored LandingEvidence machinery (the lead's design ruling on
// plan.md M3): the existing machinery's validation is untouched, and the two
// coexist on one archived row because they are different facts. LandingEvidence
// is what an OPERATOR asserted about a delivering commit; LandingVerdict is
// what the done-time landing query ANSWERED, together with the ref that
// answered and the instant it answered at.
//
// The record deliberately carries NO SHA. A query-derived SHA is outside the
// landing evidence store's write authority — LandingEvidence.Validate accepts
// operator-recorded SHAs only (LandingSHASourceOperator is its only lawful
// provenance) — so a verdict record that stored one would smuggle a derived
// SHA into the queue under a second schema. The struct has no SHA field at
// all: the delivering SHA is re-derived at re-adjudication time by re-running
// the axis-F attribution predicate against the recorded ref (REQ-TST-013) —
// storage records the predicate's answer and its coordinates; it never
// re-derives attribution.
//
// The record is a pure value: it reads no configuration, spawns no process,
// and knows nothing about a card. The done verb (internal/cli) supplies every
// field, and this file's only job is to refuse the shapes the requirements
// forbid — above all REQ-TST-010's: a verdict without its answering ref is
// not storable, because a snapshot of what no named ref answered is an
// attribution claim wearing a timestamp.
package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// The record's wire keys, as named constants rather than bare tags — the
// same discipline LandingEvidence's keys follow, so a test asking "does the
// record carry a SHA key?" asks it of the implementation's own names.
const (
	// LandingVerdictKeyVerdict is the query's three-valued answer.
	LandingVerdictKeyVerdict = "verdict"
	// LandingVerdictKeyRef is the ref the query was asked against — the
	// coordinate every re-adjudication re-runs the predicate against.
	LandingVerdictKeyRef = "ref"
	// LandingVerdictKeyAt is the instant the query answered, RFC 3339 UTC.
	LandingVerdictKeyAt = "at"
)

// ErrLandingVerdictAbsent is returned when a stored value carries no record.
// Absence is SQL NULL — never {} and never "" (REQ-TST-010) — so this is
// what a caller sees when it asks a column that was never written.
var ErrLandingVerdictAbsent = errors.New("factory: no landing verdict recorded")

// LandingVerdict is one done-time landing query answer, persisted on the
// archived row when `done --require-landed` ran and the archive accepted the
// card (REQ-TST-008).
//
// Verdict carries the query's LandingAnswer verbatim — landed, not-landed
// (which reaches storage only if a future caller persists a refusal-path
// answer; the done verb's refusal path archives nothing), or unknown when
// the query ran and could not answer ("unknown-as-answered" is a fact about
// the query, distinct from no query having run — REQ-TST-009 keeps that
// distinction structural: without the flag nothing is persisted at all).
//
// Ref is the answering ref, exactly as the printed `ref=<ref>` suffix named
// it. At is the verdict instant. There is no SHA field, by design — see the
// file comment.
type LandingVerdict struct {
	// Verdict is the landing query's answer. LandingAnswer is reused so the
	// stored vocabulary cannot drift from the query's own three values.
	Verdict LandingAnswer `json:"verdict"`
	// Ref is the ref the query answered against. Mandatory: a verdict with
	// no answering ref is refused at every write (REQ-TST-010).
	Ref string `json:"ref"`
	// At is the verdict instant, RFC 3339 UTC — the same shape the queue's
	// other timestamps use.
	At string `json:"at"`
}

// Validate reports why a record may not be stored, or nil.
//
// All three fields are mandatory. The verdict must be one of the query's
// three values — a fourth spelling is not storable, because the record
// claims to be the query's answer and nothing else. The ref rule is
// REQ-TST-010: every write path funnels through the encoders below, so an
// unverifiable shape cannot reach the column.
func (v LandingVerdict) Validate() error {
	switch strings.TrimSpace(string(v.Verdict)) {
	case string(LandingLanded), string(LandingNotLanded), string(LandingUnknown):
	default:
		return fmt.Errorf("factory: landing verdict %q is not one of the landing query's answers (%s, %s, %s)",
			v.Verdict, LandingLanded, LandingNotLanded, LandingUnknown)
	}
	if strings.TrimSpace(v.Ref) == "" {
		return fmt.Errorf("factory: landing verdict has no %s; a verdict without its answering ref is not storable", LandingVerdictKeyRef)
	}
	if strings.TrimSpace(v.At) == "" {
		return fmt.Errorf("factory: landing verdict has no %s", LandingVerdictKeyAt)
	}
	return nil
}

// EncodeLandingVerdict renders a record for storage, refusing any shape the
// requirements forbid rather than storing it for a reader to discover.
func EncodeLandingVerdict(v LandingVerdict) (string, error) {
	if err := v.Validate(); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("factory: encode landing verdict: %w", err)
	}
	return string(encoded), nil
}

// DecodeLandingVerdict reads a stored value back. An empty value is
// ErrLandingVerdictAbsent — a different fact from a value that will not
// parse or fails its own validation, both of which surface as errors, never
// as a zero-valued record that would read as an answer nobody gave.
func DecodeLandingVerdict(stored string) (LandingVerdict, error) {
	var v LandingVerdict
	if strings.TrimSpace(stored) == "" {
		return v, ErrLandingVerdictAbsent
	}
	if err := json.Unmarshal([]byte(stored), &v); err != nil {
		return LandingVerdict{}, fmt.Errorf("factory: decode landing verdict: %w", err)
	}
	if err := v.Validate(); err != nil {
		return LandingVerdict{}, fmt.Errorf("factory: stored landing verdict is not a record: %w", err)
	}
	return v, nil
}

// LandingVerdictValue converts an optional record into the value the driver
// stores: nil (SQL NULL) for no record, the encoded string for one.
//
// It is the single seam through which the column is written, so REQ-TST-010's
// "no verdict without its ref" is a property of the type rather than a
// discipline each call site must remember: a caller holding a record without
// a ref cannot reach the column from here.
func LandingVerdictValue(v *LandingVerdict) (any, error) {
	if v == nil {
		return nil, nil
	}
	encoded, err := EncodeLandingVerdict(*v)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}
