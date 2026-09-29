// classification.go — SPEC-TODO-CLASSIFY-DISPATCH-001 M1: the
// creation-time card-classification data model.
//
// A card carries ONE additive nullable classification field on BacklogItem —
// priority, blocked, execution mode, decider identity, classified-at stamp,
// one-line reason — following the Landing/PickedAt additive discipline: a
// pointer, `omitempty`, absence marshaling byte-identically to the pre-SPEC
// record, and the SQLite mirror added by the pragma_table_info-gated
// ALTER TABLE ... ADD COLUMN path (backlog_sqlite.go).
//
// Value sets are CLOSED and validated here, once: no other file re-spells
// them. Absence is never a hidden member of either axis — a card recorded
// before this SPEC reads through EffectiveCardClassification, which derives
// the defaults positively on read (REQ-TCD-014): priority normal, blocked
// false, mode serial, decider default. The serial read default is the
// conservative direction (leader ruling 2026-09-29, OD-3 extension): absent
// = unclassified = treated as serial, so an unclassified card participates
// in the serial-vs-serial mutual exclusivity of the dispatch layer.
package kanban

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// The classification value sets. The decider identities reuse the contract
// vocabulary's product members (llm, human) plus the deterministic fallback
// (default); the jev identities are REFUSED on every product classification
// path (REQ-TCD-004 — "jev" names local-only tooling no product path may
// depend on).
const (
	ClassPriorityHigh   = "high"
	ClassPriorityNormal = "normal"
	ClassPriorityLow    = "low"

	ClassModeSerial         = "serial"
	ClassModeParallelizable = "parallelizable"

	DeciderIdentityDefault = "default"
	DeciderIdentityHuman   = "human"
	DeciderIdentityLLM     = "llm"

	// DeciderIdentityJev and its compound form are named only to refuse them:
	// the validator's refusal message quotes the rejected value, so a caller
	// supplying either sees the identity it named.
	DeciderIdentityJev    = "jev"
	DeciderIdentityLLMJev = "llm+jev"
)

// classPriorities and classModes are the closed sets, ranked where the sort
// key needs a rank. classDeciders is the accepted identity set.
var (
	classPriorities = map[string]int{
		ClassPriorityHigh:   2,
		ClassPriorityNormal: 1,
		ClassPriorityLow:    0,
	}
	classModes = map[string]bool{
		ClassModeSerial:         true,
		ClassModeParallelizable: true,
	}
	classDeciders = map[string]bool{
		DeciderIdentityDefault: true,
		DeciderIdentityHuman:   true,
		DeciderIdentityLLM:     true,
	}
)

// CardClassification is one card's recorded creation-time judgment. It is
// written only through the add path's decider seam, inside the locked write
// that appends the card.
type CardClassification struct {
	Priority     string `json:"priority"`
	Blocked      bool   `json:"blocked"`
	Mode         string `json:"mode"`
	Decider      string `json:"decider"`
	ClassifiedAt string `json:"classified_at"`
	Reason       string `json:"reason"`
}

// DefaultCardClassification is the fail-safe default (REQ-TCD-003): the
// values an unavailable or failed decider promotes, and the values an absent
// field derives on read (REQ-TCD-014). One function, both call sites — the
// two defaults are the same defaults by ruling, not by coincidence.
func DefaultCardClassification() CardClassification {
	return CardClassification{
		Priority: ClassPriorityNormal,
		Blocked:  false,
		Mode:     ClassModeSerial,
		Decider:  DeciderIdentityDefault,
	}
}

// ValidateCardClassification refuses anything outside the closed sets, the
// jev identities included. The error names the offending field and value.
func ValidateCardClassification(c CardClassification) error {
	if _, ok := classPriorities[c.Priority]; !ok {
		return fmt.Errorf("classification priority %q is not one of high, normal, low", c.Priority)
	}
	if !classModes[c.Mode] {
		return fmt.Errorf("classification mode %q is not one of serial, parallelizable", c.Mode)
	}
	if !classDeciders[c.Decider] {
		if c.Decider == DeciderIdentityJev || c.Decider == DeciderIdentityLLMJev {
			return fmt.Errorf("classification decider %q is refused: jev is never a product classification decider", c.Decider)
		}
		return fmt.Errorf("classification decider %q is not one of llm, human, default", c.Decider)
	}
	return nil
}

// ParseCardClassificationJSON decodes and validates one supplied
// classification (the judgement-file input of REQ-TCD-004). A malformed
// document and an out-of-set value are both refused — the caller turns the
// error into a usage refusal with nothing written.
func ParseCardClassificationJSON(data []byte) (CardClassification, error) {
	var c CardClassification
	if err := json.Unmarshal(data, &c); err != nil {
		return CardClassification{}, fmt.Errorf("classification json: %w", err)
	}
	if err := ValidateCardClassification(c); err != nil {
		return CardClassification{}, err
	}
	return c, nil
}

// EffectiveCardClassification derives the card's classification at read:
// the recorded judgment when present, the positive default derivation when
// absent (REQ-TCD-014). Every machine consumer reads through this seam —
// none tests the pointer for nil itself.
//
// @MX:ANCHOR: [AUTO] EffectiveCardClassification — the sole absent-field default derivation seam (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-014)
// @MX:REASON: expected fan_in >= 3 (factory selection, add-path rendering, list/web consumers); a second derivation site would let one reader treat absence differently from the rest, which is exactly the silent divergence the SPEC exists to close
// @MX:SPEC: SPEC-TODO-CLASSIFY-DISPATCH-001
func EffectiveCardClassification(it BacklogItem) CardClassification {
	if it.Classification != nil {
		return *it.Classification
	}
	return DefaultCardClassification()
}

// CardClassificationValue renders a classification for its SQLite TEXT
// column: nil for absence (SQL NULL — the REQ-TLE-006 discipline), the
// canonical JSON encoding for a present record. The classified_at stamp and
// reason may be empty; the value-set fields may not — a write of an invalid
// record is refused here, aborting the caller's transaction, rather than
// landing an undecodable column value.
func CardClassificationValue(c *CardClassification) (any, error) {
	if c == nil {
		return nil, nil
	}
	if err := ValidateCardClassification(*c); err != nil {
		return nil, err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("encode classification: %w", err)
	}
	return string(data), nil
}

// decodeCardClassification decodes a column value back into the record. The
// column's only writer is CardClassificationValue, which validates on write,
// so an undecodable value is external corruption — surfaced, never dropped
// (the same contract as the landing evidence read).
func decodeCardClassification(s string) (CardClassification, error) {
	var c CardClassification
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		return CardClassification{}, fmt.Errorf("decode classification %q: %w", truncateClassification(s), err)
	}
	if err := ValidateCardClassification(c); err != nil {
		return CardClassification{}, fmt.Errorf("decode classification: %w", err)
	}
	return c, nil
}

// truncateClassification bounds a corrupt value quoted in an error.
func truncateClassification(s string) string {
	if len(s) <= 64 {
		return s
	}
	return strings.Clone(s[:64]) + "..."
}

// SortByClassification re-establishes the queue's classification order
// (REQ-TCD-005): non-blocked before blocked, then priority high > normal >
// low, then insertion order stable within a rank. It runs inside the same
// locked write as the add that may change the order — never on read (plan
// G5: two orderings that can disagree is the defect this SPEC exists to
// close). Every state takes its rank position; selection filters states,
// sorting only positions them (constraint C3).
//
// @MX:ANCHOR: [AUTO] SortByClassification — the sole queue-order restorer (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-005)
// @MX:REASON: expected fan_in >= 3 (CLI append add, CLI add --pick, store.Add/GTD publication, tests); a write path that appends without re-sorting here leaves the queue unsorted and every position-reading consumer lies
// @MX:SPEC: SPEC-TODO-CLASSIFY-DISPATCH-001
func (r *BacklogRecord) SortByClassification() {
	items := r.Items
	sort.SliceStable(items, func(i, j int) bool {
		a := EffectiveCardClassification(items[i])
		b := EffectiveCardClassification(items[j])
		if a.Blocked != b.Blocked {
			return !a.Blocked
		}
		ra, rb := classPriorities[a.Priority], classPriorities[b.Priority]
		return ra > rb
	})
}

// QueuedPosition returns the id's 1-based position among the queued items in
// the record's (sorted) order, or 0 when the id is absent or not queued. The
// position add prints is THIS position — the count of queued cards that
// precede it in sorted order plus one — never the append index.
func (r *BacklogRecord) QueuedPosition(id string) int {
	pos := 0
	for _, it := range r.Items {
		if it.State != BacklogStateQueued {
			continue
		}
		pos++
		if it.ID == id {
			return pos
		}
	}
	return 0
}

// CardDecider is the classification decider seam (REQ-TCD-001/-012): the
// add path resolves every admitted card's judgment through Classify INSIDE
// the locked write that appends the card. An error means the decider is
// unavailable or failed — the add path then promotes the fail-safe default
// (REQ-TCD-003) and never blocks admission on classification.
type CardDecider interface {
	Classify(text string) (CardClassification, error)
}

// DefaultCardDecider is the deterministic shipped implementation
// (REQ-TCD-012): with no judgment supplied, it answers with the fail-safe
// defaults under the default decider identity. It is a judgment, not a
// failure — it never errors and prints no fallback notice.
type DefaultCardDecider struct{}

// Classify returns the deterministic default classification.
func (DefaultCardDecider) Classify(string) (CardClassification, error) {
	c := DefaultCardClassification()
	c.Reason = "deterministic default: no classification judgment supplied"
	return c, nil
}

// StaticCardDecider carries one pre-validated supplied judgment — the
// --classification-file transport (REQ-TCD-004). The judgment was validated
// and its decider identity recorded before it reached the seam; Classify
// hands it through verbatim.
type StaticCardDecider struct {
	Class CardClassification
}

// Classify returns the carried judgment.
func (d StaticCardDecider) Classify(string) (CardClassification, error) {
	return d.Class, nil
}

// UnavailableCardDecider is the transport-unavailable shape (REQ-TCD-003):
// every Classify call fails with the wrapped cause, so the add path's
// single fallback branch promotes the defaults and prints the one-line
// notice regardless of which transport failed.
type UnavailableCardDecider struct {
	Cause error
}

// Classify always fails with the carried cause.
func (d UnavailableCardDecider) Classify(string) (CardClassification, error) {
	return CardClassification{}, d.Cause
}
