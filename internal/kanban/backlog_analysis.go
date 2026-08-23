// backlog_analysis.go — SPEC-TODO-ANALYSIS-001 M3: the mechanical text
// analyser behind `moai todo add` and `moai todo analyze`.
//
// Everything here is a pure function over text: no store, no file, no clock.
// That is deliberate. The analyser decides whether a card is admitted, and a
// decision that can only be exercised through the CLI is a decision that
// gets tested through the CLI — slowly, and only along the paths the CLI
// happens to take. Kept pure, the classification boundaries are table-tested
// directly.
//
// The layer reads TEXT and nothing else. Two cards that mean the same thing
// in different words are `none` here, and that is not a gap to close later:
// a false positive at this layer refuses a legitimate card, and the operator
// loses work the queue was supposed to hold. Judging meaning is the agent
// layer's job (`todo relate`), where the outcome is a record rather than a
// refusal.
package kanban

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// BacklogNearDuplicateThreshold is the token-set Jaccard score at or above
// which two card texts are recorded as near duplicates.
//
// The error is asymmetric, so the value errs high. Set too low, findings
// outnumber the cards they annotate, the operator stops reading the listing,
// and the feature is worse than absent. Set too high, a near duplicate goes
// unrecorded — which is exactly the state the queue was in before this
// feature, so the cost of missing is the status quo rather than a regression.
const BacklogNearDuplicateThreshold = 0.80

// BacklogMatchKind classifies a candidate text against the existing queue.
type BacklogMatchKind string

const (
	// BacklogMatchNone marks a candidate resembling no card closely enough
	// to record.
	BacklogMatchNone BacklogMatchKind = "none"
	// BacklogMatchNear marks a measured similarity in [threshold, 1.0).
	BacklogMatchNear BacklogMatchKind = "near"
	// BacklogMatchExact marks a candidate whose normalized text equals a
	// card's normalized text. This is the only classification that refuses
	// an append.
	BacklogMatchExact BacklogMatchKind = "exact"
)

// BacklogMatch is one classification result: what was found, which card it
// was found against, and the measured score behind it.
type BacklogMatch struct {
	Kind  BacklogMatchKind
	ID    string
	Score float64
}

// NormalizeCardText applies the four-step comparison pipeline: Unicode NFC,
// trim, internal-whitespace collapse, case-fold.
//
// The pipeline's LENGTH is part of the contract. Every additional step —
// stemming, stop-word removal, punctuation stripping — raises measured
// similarity across the whole queue and walks distinct cards over the
// near-duplicate threshold from below. Punctuation is therefore kept: "ship
// it" and "ship it!" are different texts here, and the cost of treating them
// as different (a missed finding) is smaller than the cost of the general
// loosening that would merge them.
//
// The result is a comparison key only. No card's stored text is ever
// replaced by it — the queue holds what the operator typed.
func NormalizeCardText(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(norm.NFC.String(s)), " "))
}

// TokenSetJaccard measures similarity as |A ∩ B| / |A ∪ B| over the SETS of
// whitespace-separated tokens in the two normalized texts.
//
// Sets, not multisets: a card repeating a word is not thereby less similar
// to one that says it once. Unordered: a card is not a different card for
// having its words rearranged. Both properties are what make the measure
// usable on the short, telegraphic texts a backlog actually holds.
//
// An empty text scores 0 against everything, including another empty text —
// the store already refuses empty cards, so the case exists only to keep the
// function total.
func TokenSetJaccard(a, b string) float64 {
	setA, setB := tokenSet(a), tokenSet(b)
	if len(setA) == 0 || len(setB) == 0 {
		return 0
	}
	intersection := 0
	for token := range setA {
		if _, ok := setB[token]; ok {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// tokenSet returns the distinct tokens of the normalized text.
func tokenSet(s string) map[string]struct{} {
	fields := strings.Fields(NormalizeCardText(s))
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	return set
}

// ClassifyCardText classifies candidate against every non-dropped card,
// returning the strongest match found.
//
// An exact match wins outright and returns immediately: it is the only
// classification with a consequence (a refused append), so a weaker match
// found later must not displace it. Among near matches the highest score
// wins, with the first card in queue order breaking a tie — the queue's
// order is the operator's, and the earliest card is the one they are most
// likely to recognise as the original.
//
// A dropped card is not a comparison subject. The operator discarded it; a
// candidate resembling it is not a duplicate of anything the queue holds,
// and refusing on that basis would make `drop` a way to lock text out of the
// queue forever.
//
// `near` is the half-open range [threshold, 1.0). A score of exactly 1.0
// with unequal normalized texts means a token permutation — the same words
// in a different order. That is not an exact duplicate and it is not "nearly
// the same wording" either, so it falls outside both classes rather than
// being quietly folded into one; widening the range later is then a
// deliberate edit rather than a silent drift.
func ClassifyCardText(candidate string, items []BacklogItem) BacklogMatch {
	normalized := NormalizeCardText(candidate)
	best := BacklogMatch{Kind: BacklogMatchNone}
	for _, item := range items {
		if item.State == BacklogStateDropped {
			continue
		}
		other := NormalizeCardText(item.Text)
		if normalized != "" && normalized == other {
			return BacklogMatch{Kind: BacklogMatchExact, ID: item.ID, Score: 1}
		}
		score := TokenSetJaccard(normalized, other)
		if score < BacklogNearDuplicateThreshold || score >= 1.0 {
			continue
		}
		if best.Kind == BacklogMatchNone || score > best.Score {
			best = BacklogMatch{Kind: BacklogMatchNear, ID: item.ID, Score: score}
		}
	}
	return best
}
