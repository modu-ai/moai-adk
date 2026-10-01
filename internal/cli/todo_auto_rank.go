package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// SPEC-TODO-AUTO-PRIORITY-001 (card t1400) — the ranking stage of
// `moai todo --auto`.
//
// The stage orders the QUEUED pickup candidates the cycle is about to accept.
// It is a pure function of the loaded record plus two injected inputs (the
// landed lookup and, from M2, the Jev answer) and it writes nothing: no queue
// verb, no card field, no finding. Only the order in which the cycle chooses
// changes — the cycle's own pick, unpick and done transitions stay the sole
// queue writes. This file is the only place the `selection:` record literals
// are written; todo_auto.go calls the renderer and holds none of them.

// The ranking sources and the readiness signals the record names.
const (
	autoRankSourceFallback = "fallback"
	autoRankSourceJev      = "jev"

	autoRankSignalHoldMarker    = "hold-marker"
	autoRankSignalLanded        = "landed"
	autoRankSignalNearDuplicate = "near-duplicate"
)

// autoRankHoldMarker is the text prefix an operator writes to park a card in
// prose. Only a card whose trimmed text OPENS with it is demoted; a mention
// further into the text, or a hold stated without the marker, is not (the
// structural hold is the `hold` queue state, which the pickup already skips).
const autoRankHoldMarker = "[보류"

// The closed vocabulary of fallback reasons, one per unavailable Jev
// condition. The `jev-` names that mirror a jev.Availability value are
// produced by autoFallbackReason; the incomplete-answer reason is separate
// because the capability itself answered `Available` — it is the stage that
// found the answer set defective (M2).
const (
	autoRankReasonDisabled         = "jev-disabled"
	autoRankReasonNoCredential     = "jev-no-credential"
	autoRankReasonUnauthorized     = "jev-unauthorized"
	autoRankReasonRateLimited      = "jev-rate-limited"
	autoRankReasonOverloaded       = "jev-overloaded"
	autoRankReasonUnreachable      = "jev-unreachable"
	autoRankReasonOversize         = "jev-oversize"
	autoRankReasonSecretDetected   = "jev-secret-detected"
	autoRankReasonMalformed        = "jev-malformed"
	autoRankReasonIncompleteAnswer = "jev-incomplete-answer"
)

// Record line vocabulary. Every printed selection line is built from these.
const (
	autoRankLinePrefix   = "selection: "
	autoRankLineSource   = autoRankLinePrefix + "source="
	autoRankLineReason   = " reason="
	autoRankLineRanked   = autoRankLinePrefix + "ranked "
	autoRankLineFlagged  = autoRankLinePrefix + "flagged "
	autoRankLineExcluded = autoRankLinePrefix + "excluded "
	autoRankLineNote     = autoRankLinePrefix + "note "
	autoRankBlockedLabel = "blocked"
)

// autoLandedLookup is the landed-state seam: one call over the whole record
// answering, per card id, the pull-request outcome kind `moai todo pr`
// reports. The stage reads only the `landed` kind; `unknown`, an absent id and
// a returned error are all UNMEASURED, never read as poor.
type autoLandedLookup func(rec *kanban.BacklogRecord) (map[string]kanban.PRLinkKind, error)

// autoRankCandidate is one eligible queued card with its ranking inputs.
type autoRankCandidate struct {
	Item         kanban.BacklogItem
	PriorityRank int
	Signals      []string
}

// poor reports whether the candidate carries at least one readiness-poor
// signal.
func (c autoRankCandidate) poor() bool { return len(c.Signals) > 0 }

// autoRankFlag is one readiness-poor card and the signals that name it.
type autoRankFlag struct {
	ID      string
	Signals []string
}

// autoRankResult is the stage's whole output: the ranked targets in
// processing order, plus the facts the printed record states.
type autoRankResult struct {
	Source   string
	Reason   string
	Ranked   []kanban.BacklogItem
	Flagged  []autoRankFlag
	Excluded []string
	Notes    []string
}

// autoRankHoldMarked reports whether the card text, with surrounding
// whitespace trimmed, begins with the hold marker.
func autoRankHoldMarked(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), autoRankHoldMarker)
}

// autoRankNearDuplicate reports whether a live near-duplicate finding names
// the card on either side, whatever recorded it.
func autoRankNearDuplicate(rec *kanban.BacklogRecord, id string) bool {
	findings, _ := rec.FindingsNaming(id)
	for _, f := range findings {
		if f.Relation == kanban.BacklogRelationNearDuplicate {
			return true
		}
	}
	return false
}

// autoRankPartition splits the queued candidates into the eligible ones and
// the ids excluded as blocked, both in queue order. Blocked is an eligibility
// fact applied on every ranking source; it changes no card state.
func autoRankPartition(queued []kanban.BacklogItem) (eligible []kanban.BacklogItem, excluded []string) {
	for _, it := range queued {
		if kanban.EffectiveCardClassification(it).Blocked {
			excluded = append(excluded, it.ID)
			continue
		}
		eligible = append(eligible, it)
	}
	return eligible, excluded
}

// autoRankCandidates measures the readiness signals for the eligible cards and
// returns one candidate per card in queue order. The landed lookup runs once
// per call. A signal that cannot be measured yields no demotion and one note
// naming it.
func autoRankCandidates(rec *kanban.BacklogRecord, eligible []kanban.BacklogItem, landed autoLandedLookup) (cands []autoRankCandidate, notes []string) {
	if len(eligible) == 0 {
		return nil, nil
	}

	// lookedUp is true only when a lookup ran and returned without error; the
	// per-card unknown note below is for that case alone, because the other two
	// cases already name the whole signal as unmeasured.
	var kinds map[string]kanban.PRLinkKind
	lookedUp := false
	if landed == nil {
		notes = append(notes, fmt.Sprintf("%s signal unmeasured (no lookup wired)", autoRankSignalLanded))
	} else if got, err := landed(rec); err != nil {
		notes = append(notes, fmt.Sprintf("%s signal unmeasured (lookup failed: %v)", autoRankSignalLanded, err))
	} else {
		kinds, lookedUp = got, true
	}

	var unknown []string
	for _, it := range eligible {
		c := autoRankCandidate{
			Item:         it,
			PriorityRank: kanban.PriorityRank(kanban.EffectiveCardClassification(it).Priority),
		}
		if autoRankHoldMarked(it.Text) {
			c.Signals = append(c.Signals, autoRankSignalHoldMarker)
		}
		if lookedUp {
			kind, answered := kinds[it.ID]
			switch {
			case answered && kind == kanban.PRLinkLanded:
				c.Signals = append(c.Signals, autoRankSignalLanded)
			case !answered || kind == kanban.PRLinkUnknown:
				unknown = append(unknown, it.ID)
			}
		}
		if autoRankNearDuplicate(rec, it.ID) {
			c.Signals = append(c.Signals, autoRankSignalNearDuplicate)
		}
		cands = append(cands, c)
	}
	if len(unknown) > 0 {
		notes = append(notes, fmt.Sprintf("%s signal unmeasured for %s (no answer or unknown)",
			autoRankSignalLanded, strings.Join(unknown, " ")))
	}
	return cands, notes
}

// autoRankOrderFallback orders candidates by the fallback keys: every clean
// card before every poor card, then recorded priority (highest rank first),
// then queue order (the stable sort keeps the input order within one key).
func autoRankOrderFallback(cands []autoRankCandidate) []autoRankCandidate {
	ordered := make([]autoRankCandidate, len(cands))
	copy(ordered, cands)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].poor() != ordered[j].poor() {
			return !ordered[i].poor()
		}
		return ordered[i].PriorityRank > ordered[j].PriorityRank
	})
	return ordered
}

// autoRankFlags lists the readiness-poor candidates in the given order.
func autoRankFlags(cands []autoRankCandidate) []autoRankFlag {
	var flags []autoRankFlag
	for _, c := range cands {
		if c.poor() {
			flags = append(flags, autoRankFlag{ID: c.Item.ID, Signals: c.Signals})
		}
	}
	return flags
}

// autoRankFallback is the stage on the fallback source: blocked cards are
// excluded, the rest are ordered by priority with readiness-poor cards moved
// behind the clean ones — demoted, never dropped. queued is the queued suffix
// of the pickup targets, in queue order. The caller sets Reason.
func autoRankFallback(rec *kanban.BacklogRecord, queued []kanban.BacklogItem, landed autoLandedLookup) autoRankResult {
	eligible, excluded := autoRankPartition(queued)
	cands, notes := autoRankCandidates(rec, eligible, landed)
	ordered := autoRankOrderFallback(cands)

	res := autoRankResult{
		Source:   autoRankSourceFallback,
		Excluded: excluded,
		Notes:    notes,
		Flagged:  autoRankFlags(ordered),
	}
	for _, c := range ordered {
		res.Ranked = append(res.Ranked, c.Item)
	}
	return res
}

// renderAutoSelectionRecord renders the selection decision record, one line
// per element: the source line (with its reason on the fallback), the ranked
// ids in processing order (omitted when none survived), one flagged line per
// readiness-poor card, one excluded line per blocked card, and the notes.
func renderAutoSelectionRecord(res autoRankResult) []string {
	source := autoRankLineSource + res.Source
	if res.Reason != "" {
		source += autoRankLineReason + res.Reason
	}
	lines := []string{source}
	if len(res.Ranked) > 0 {
		ids := make([]string, 0, len(res.Ranked))
		for _, it := range res.Ranked {
			ids = append(ids, it.ID)
		}
		lines = append(lines, autoRankLineRanked+strings.Join(ids, " "))
	}
	for _, f := range res.Flagged {
		lines = append(lines, fmt.Sprintf("%s%s (%s)", autoRankLineFlagged, f.ID, strings.Join(f.Signals, ", ")))
	}
	for _, id := range res.Excluded {
		lines = append(lines, fmt.Sprintf("%s%s (%s)", autoRankLineExcluded, id, autoRankBlockedLabel))
	}
	for _, n := range res.Notes {
		lines = append(lines, autoRankLineNote+n)
	}
	return lines
}

// autoFallbackReason maps an unavailable Jev availability onto the closed
// fallback-reason vocabulary. An available result has no fallback reason and
// returns empty; a value outside the known set maps to the conservative
// unreachable reason, so a printed fallback always carries exactly one reason.
func autoFallbackReason(a jev.Availability) string {
	switch a {
	case jev.Available:
		return ""
	case jev.Disabled:
		return autoRankReasonDisabled
	case jev.NoCredential:
		return autoRankReasonNoCredential
	case jev.Unauthorized:
		return autoRankReasonUnauthorized
	case jev.RateLimited:
		return autoRankReasonRateLimited
	case jev.Overloaded:
		return autoRankReasonOverloaded
	case jev.Unreachable:
		return autoRankReasonUnreachable
	case jev.Oversize:
		return autoRankReasonOversize
	case jev.SecretDetected:
		return autoRankReasonSecretDetected
	case jev.Malformed:
		return autoRankReasonMalformed
	default:
		return autoRankReasonUnreachable
	}
}
