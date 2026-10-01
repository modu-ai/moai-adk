package cli

import (
	"context"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/jev"
	"github.com/modu-ai/moai-adk/internal/jevcred"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// SPEC-TODO-AUTO-PRIORITY-001 (card t1400) — the ranking stage of
// `moai todo --auto`.
//
// The stage orders the QUEUED pickup candidates the cycle is about to accept.
// It is a pure function of the loaded record plus two injected inputs (the
// landed lookup and the Jev ranker) and it writes nothing: no queue verb, no
// card field, no finding. Only the order in which the cycle chooses changes —
// the cycle's own pick, unpick and done transitions stay the sole queue
// writes. This file is the only place the `selection:` record literals are
// written; todo_auto.go calls the renderer and holds none of them.
//
// Jev is a selection-order signal ONLY. A Jev answer is read once, validated as
// a whole, and used to order the candidates the cycle is about to accept; it is
// never a completion verdict, a merge approval, an operator gate or a queue
// mutation, and nothing here claims the ordering is accurate. The consumer is
// built behind the default-off `workflow.jev.enabled` gate.

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

// The Jev ordering request. The bounds are named constants, not inline
// literals: the candidate bound mirrors the near-duplicate consumer's, the call
// bound keeps a degraded endpoint from stalling the cycle, and the score scale
// is stated once — autoRankJevLevelCount named levels, zero-based, so a valid
// score is a finite number in [autoRankJevScoreMin, autoRankJevScoreMax].
const (
	autoRankJevCandidateLimit = 40
	autoRankJevCallTimeout    = 5 * time.Second
	autoRankJevLevelCount     = 5
	autoRankJevScoreMin       = 0.0
	autoRankJevScoreMax       = float64(autoRankJevLevelCount - 1)
)

// autoRankJevLevels are the score levels, lowest first. Level 0 is the honest
// no-match answer: a card nobody could usefully start now. A higher level means
// the card is more ready to be picked now.
var autoRankJevLevels = [autoRankJevLevelCount]string{
	"not ready: no worker could usefully start this card now (parked, already delivered, duplicated, or waiting on something else)",
	"barely ready: a worker could start only after settling a significant open question",
	"partly ready: workable, with some open questions",
	"ready: a worker can start now with the card text as written",
	"most ready: clear, unblocked, and the best card to pick now",
}

// Record line vocabulary. Every printed selection line is built from these.
const (
	autoRankLinePrefix = "selection: "
	// The source line is spelled out whole: acceptance AC-TAP-001's ledger row
	// finds the record's writer by this contiguous literal, and a literal built
	// from the prefix plus a fragment is invisible to that search.
	autoRankLineSource   = "selection: source="
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

// autoJevRanker is the Jev ordering seam: one bounded request in, the
// capability's own typed Result out. It never returns an error — every absence
// is a typed unavailability the stage maps onto a fallback reason. A nil seam
// is Jev unavailable (`jev-disabled`): the cycle makes no call.
type autoJevRanker func(req jev.Request) jev.Result

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

// autoRankResultOf assembles the stage output for one source from the cards in
// their final processing order.
func autoRankResultOf(source string, excluded, notes []string, ordered []autoRankCandidate) autoRankResult {
	res := autoRankResult{
		Source:   source,
		Excluded: excluded,
		Notes:    notes,
		Flagged:  autoRankFlags(ordered),
	}
	for _, c := range ordered {
		res.Ranked = append(res.Ranked, c.Item)
	}
	return res
}

// autoRankFallback is the stage on the fallback source: blocked cards are
// excluded, the rest are ordered by priority with readiness-poor cards moved
// behind the clean ones — demoted, never dropped. queued is the queued suffix
// of the pickup targets, in queue order. The caller sets Reason.
func autoRankFallback(rec *kanban.BacklogRecord, queued []kanban.BacklogItem, landed autoLandedLookup) autoRankResult {
	eligible, excluded := autoRankPartition(queued)
	cands, notes := autoRankCandidates(rec, eligible, landed)
	return autoRankResultOf(autoRankSourceFallback, excluded, notes, autoRankOrderFallback(cands))
}

// autoRank is the whole stage. Blocked cards are excluded on both sources; the
// readiness signals are measured once. When the Jev signal is available (one
// bounded request, a complete and valid answer set) it orders the candidates;
// otherwise the fallback order stands and the record names the one reason.
// queued is the queued suffix of the pickup targets, in queue order.
//
// Demotion is a fallback-source rule: on the Jev source a readiness-poor card
// is flagged in the record but keeps the place Jev gave it.
func autoRank(rec *kanban.BacklogRecord, queued []kanban.BacklogItem, landed autoLandedLookup, ask autoJevRanker) autoRankResult {
	eligible, excluded := autoRankPartition(queued)
	cands, notes := autoRankCandidates(rec, eligible, landed)
	fallback := autoRankOrderFallback(cands)

	if len(fallback) == 0 {
		// Nothing is left to rank, so no request is sent: the stage never
		// sends an empty question list. The closed vocabulary has no "not
		// consulted" reason; this names the conservative one and says why.
		res := autoRankResultOf(autoRankSourceFallback, excluded, notes, nil)
		res.Reason = autoRankReasonDisabled
		res.Notes = append(res.Notes, "no eligible candidate remained, so jev was not consulted")
		return res
	}

	ordered, reason, jevNotes := autoRankJevOrder(fallback, ask)
	notes = append(notes, jevNotes...)
	if reason != "" {
		res := autoRankResultOf(autoRankSourceFallback, excluded, notes, fallback)
		res.Reason = reason
		return res
	}
	return autoRankResultOf(autoRankSourceJev, excluded, notes, ordered)
}

// autoRankTargets applies the stage to the queued suffix of the pickup targets
// and prints the selection record. The dead-owner rescue targets keep their
// place at the head, untouched and unsent; only the queued suffix is ranked.
// The record is skipped when no queued candidate survived the relation filter.
// It writes nothing to the queue.
func autoRankTargets(out io.Writer, rec *kanban.BacklogRecord, targets []kanban.BacklogItem, opts autoOptions) []kanban.BacklogItem {
	var rescue, queued []kanban.BacklogItem
	for _, it := range targets {
		if it.State == kanban.BacklogStateQueued {
			queued = append(queued, it)
		} else {
			rescue = append(rescue, it)
		}
	}
	if len(queued) == 0 {
		return targets
	}
	res := autoRank(rec, queued, opts.landed, opts.jevRank)
	for _, line := range renderAutoSelectionRecord(res) {
		_, _ = fmt.Fprintln(out, line)
	}
	return append(rescue, res.Ranked...)
}

// autoRankJevOrder asks Jev to order the candidates, which arrive in fallback
// order. It returns the full processing order, or — when the signal is not
// available — an empty order and the one fallback reason. At most
// autoRankJevCandidateLimit candidates are sent (the first in fallback order);
// the surplus follows the Jev-ordered cards in fallback order and a note says
// so. No lock is held: the cycle ranks over an in-memory record.
func autoRankJevOrder(fallback []autoRankCandidate, ask autoJevRanker) (ordered []autoRankCandidate, reason string, notes []string) {
	if ask == nil {
		return nil, autoRankReasonDisabled, nil
	}
	sent, surplus := fallback, []autoRankCandidate(nil)
	if len(fallback) > autoRankJevCandidateLimit {
		sent, surplus = fallback[:autoRankJevCandidateLimit], fallback[autoRankJevCandidateLimit:]
	}

	res := ask(autoRankJevRequest(sent))
	if !res.OK() {
		return nil, autoFallbackReason(res.Availability), nil
	}
	answers, defect := autoRankJevAnswers(sent, res.Answers)
	if defect != "" {
		return nil, autoRankReasonIncompleteAnswer, []string{"jev answer set rejected as a whole: " + defect}
	}

	ordered = make([]autoRankCandidate, len(sent))
	copy(ordered, sent)
	// Higher score first, then higher confidence; the stable sort keeps the
	// fallback order for a full tie.
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := answers[ordered[i].Item.ID], answers[ordered[j].Item.ID]
		if a.score != b.score {
			return a.score > b.score
		}
		return a.confidence > b.confidence
	})
	if len(surplus) > 0 {
		ids := make([]string, 0, len(surplus))
		for _, c := range surplus {
			ids = append(ids, c.Item.ID)
		}
		notes = append(notes, fmt.Sprintf("jev request bound %d reached: %d further card(s) follow the jev order in fallback order (%s)",
			autoRankJevCandidateLimit, len(surplus), strings.Join(ids, " ")))
	}
	return append(ordered, surplus...), "", notes
}

// autoRankJevRequest builds the one request: one state listing the candidates
// as computed facts, and one score question per candidate keyed by its card id.
func autoRankJevRequest(cands []autoRankCandidate) jev.Request {
	var state strings.Builder
	state.WriteString("A work queue holds cards waiting to be picked, one at a time, by a worker. ")
	state.WriteString("Each card is listed with its recorded priority and the readiness flags a program measured for it.\n\nCARDS\n")
	questions := make([]jev.Question, 0, len(cands))
	for _, c := range cands {
		flags := "none"
		if c.poor() {
			flags = strings.Join(c.Signals, ", ")
		}
		fmt.Fprintf(&state, "%s | priority=%s | readiness-flags=%s | %s\n",
			c.Item.ID, kanban.EffectiveCardClassification(c.Item).Priority, flags, todoTextPrefix(c.Item.Text))
		questions = append(questions, jev.Question{
			ID: c.Item.ID,
			Text: fmt.Sprintf("How ready is card %s to be picked and worked on right now? "+
				"Level 0 means no worker could usefully start it; the highest level means it is the most ready card to pick now.", c.Item.ID),
			Kind:   jev.KindScore,
			Levels: autoRankJevLevels[:],
		})
	}
	return jev.Request{State: state.String(), Questions: questions}
}

// autoRankJevAnswer is one validated score answer.
type autoRankJevAnswer struct {
	score      float64
	confidence float64
}

// autoRankJevAnswers validates the answer set as a whole against the cards that
// were sent. Any defect — an answer naming a card that was not sent, a card
// answered twice or not at all, an answer that is not a score, a score or
// confidence that is not a finite number, a score outside the declared range —
// rejects the whole set: nothing is applied partially.
func autoRankJevAnswers(sent []autoRankCandidate, answers []jev.Answer) (map[string]autoRankJevAnswer, string) {
	wanted := make(map[string]bool, len(sent))
	for _, c := range sent {
		wanted[c.Item.ID] = true
	}
	got := make(map[string]autoRankJevAnswer, len(sent))
	for _, a := range answers {
		switch {
		case !wanted[a.QuestionID]:
			return nil, fmt.Sprintf("answer names card %q, which was not sent", a.QuestionID)
		case a.Kind != jev.KindScore:
			return nil, fmt.Sprintf("answer for card %q is not a score", a.QuestionID)
		case math.IsNaN(a.Score) || math.IsInf(a.Score, 0):
			return nil, fmt.Sprintf("score %v for card %q is not a finite number", a.Score, a.QuestionID)
		case a.Score < autoRankJevScoreMin || a.Score > autoRankJevScoreMax:
			return nil, fmt.Sprintf("score %v for card %q is outside %v..%v", a.Score, a.QuestionID, autoRankJevScoreMin, autoRankJevScoreMax)
		case math.IsNaN(a.Probability) || math.IsInf(a.Probability, 0):
			return nil, fmt.Sprintf("confidence %v for card %q is not a finite number", a.Probability, a.QuestionID)
		}
		if _, dup := got[a.QuestionID]; dup {
			return nil, fmt.Sprintf("card %q answered more than once", a.QuestionID)
		}
		got[a.QuestionID] = autoRankJevAnswer{score: a.Score, confidence: a.Probability}
	}
	for _, c := range sent {
		if _, ok := got[c.Item.ID]; !ok {
			return nil, fmt.Sprintf("card %q has no answer", c.Item.ID)
		}
	}
	return got, ""
}

// autoRankJevDoer is the transport the live ranker hands the Jev client. Nil
// means the client's own default; tests install a fake, so no test contacts
// the endpoint.
var autoRankJevDoer jev.Doer

// @MX:NOTE: [AUTO] gate-unrun ordering consumer (SPEC-TODO-AUTO-PRIORITY-001) —
// unreachable at the shipped default (workflow.jev.enabled is false), selection
// order only, and it claims no ordering accuracy; enabling it is the operator's
// act under SPEC-JEV-OPTIN-MEASURE-001. Not an error path.
//
// liveAutoJevRanker is the production seam body: resolve the gate, then make
// one bounded call through the capability's own client. A gate that is off, or
// a config that cannot be read, is Jev disabled and constructs no request.
func liveAutoJevRanker(req jev.Request) jev.Result {
	enabled, err := jevEnabled(resolveProjectDir())
	if err != nil || !enabled {
		return jev.Result{
			Availability: jev.Disabled,
			Condition:    "workflow.jev.enabled is false or unreadable; no request constructed",
		}
	}
	client := jev.New(true)
	client.LoadCredential = jevcred.Load
	if autoRankJevDoer != nil {
		client.HTTP = autoRankJevDoer
	}
	ctx, cancel := context.WithTimeout(context.Background(), autoRankJevCallTimeout)
	defer cancel()
	return client.Ask(ctx, req)
}

// liveAutoLandedLookup is the production landed seam: one pass over the same
// computation `moai todo pr` renders — one `gh` query, a local git check per
// card, fail-open to `unknown`. Its degradation notes are discarded here; the
// selection record names the unmeasured signal instead.
func liveAutoLandedLookup(rec *kanban.BacklogRecord) (map[string]kanban.PRLinkKind, error) {
	rows := computeTodoPRRows(io.Discard, rec, "")
	kinds := make(map[string]kanban.PRLinkKind, len(rows))
	for _, r := range rows {
		kinds[r.CardID] = r.Kind
	}
	return kinds, nil
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
