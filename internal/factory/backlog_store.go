// backlog_store.go — the lock-guarded backlog queue store
// (SPEC-KANBAN-TODO-CLI-001 REQ-TODO-006..013, M1).
//
// The backlog is the OPERATOR's queue, not the board's: any session may
// append, pick, or complete a card, so this store deliberately applies NO
// sole-writer role guard (the explicit contrast with the retired board store — the
// board has exactly one writer, the leader; the backlog has every writer).
// What it does share with the board is the concurrency substrate: mutations
// serialize on a sibling advisory lock (backlog.lock, the same
// path-parameterized flock/atomic-create split the board lock uses) across
// the entire read-modify-write, and land through the same-directory temp +
// atomic rename. Reads are lock-free.
//
// Ids are issued INSIDE the locked mutation from the persisted high-water
// mark `last_seq`. The mark — never max-present-id — decides the next id,
// because `done` removes rows and a derived mark would reuse the removed
// card's id (the t4/t5/t6 collision this SPEC exists to kill).
package factory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// backlogLockFileName names the lock artifact sibling to the backlog file.
const backlogLockFileName = "backlog.lock"

const backlogRetiredFileName = "backlog.relocated"

var ErrBacklogRelocated = errors.New("backlog relocated; resolve the project queue again")

// legacyBacklogLockFileName is the lock artifact name an earlier revision of
// this store used. No code reads it anymore, but an install that lived
// through the rename keeps a stale zero-byte artifact beside the live lock
// (measured on the primary checkout: 0 B, three days older than backlog.lock).
// The adopting open sweeps it best-effort so the directory settles on one lock
// name instead of carrying two.
const legacyBacklogLockFileName = "backlog.json.lock"

// backlogVersion is the record schema version. The schema is ADDITIVE within
// version 1: `last_seq` was appended as a top-level field with no version
// bump, and no per-item field may ever be added (spec.md §E out-of-scope).
// `findings` follows the same precedent — a second top-level field, no
// version bump, and the five-field per-item contract untouched.
const backlogVersion = 1

// BacklogState is a queue item's lifecycle state.
type BacklogState string

const (
	// BacklogStateQueued marks a card waiting in the queue.
	BacklogStateQueued BacklogState = "queued"
	// BacklogStatePicked marks a card chosen for a SPEC (spec_id attached).
	BacklogStatePicked BacklogState = "picked"
	// BacklogStateDropped marks a card the operator discarded.
	BacklogStateDropped BacklogState = "dropped"
	// BacklogStateHold marks a card an OPERATOR parked out of the queue
	// (SPEC-TODO-HOLD-STATE-001). It is the state, not a text marker: every
	// actionable surface enumerates accepted states positively, so a held
	// card is invisible to every machine selector by construction, and no
	// lease path gains a verb that sets or clears it.
	BacklogStateHold BacklogState = "hold"
)

// BacklogItem is one queued card. The five original fields are the frozen
// per-item contract (REQ-TODO-013): SpecID is a pointer so an absent spec id
// round-trips as JSON null, not as an omitted key.
//
// Landing is ADDITIVE (SPEC-TODO-LANDING-EVIDENCE-001, design.md §5): the
// five keep their names, types, and JSON tags, and the new one is `omitempty`
// so a card with no evidence marshals byte-identically to before. Absence is
// a nil pointer here and SQL NULL in the column — never `{}` and never `""`
// (REQ-TLE-006), which is why the field is a pointer rather than a value.
type BacklogItem struct {
	ID      string       `json:"id"`
	Text    string       `json:"text"`
	AddedAt string       `json:"added_at"`
	SpecID  *string      `json:"spec_id"`
	State   BacklogState `json:"state"`
	// Landing is the operator-recorded landing evidence, or nil when none
	// was recorded. Written ONLY through LandingEvidenceValue, so the
	// encoder's refusals (a SHA without provenance, a provenance that is not
	// `operator`) hold on every write rather than at each call site.
	Landing  *LandingEvidence `json:"landing,omitempty"`
	CardUUID *string          `json:"card_uuid"`
	// PickedAt / DroppedAt are ADDITIVE (SPEC-TODO-TRANSITION-STAMPS-001
	// REQ-TST-001): nullable transition stamps in the added_at TEXT format,
	// `omitempty` after the Landing precedent so a card carrying none
	// marshals byte-identically to before. PickedAt answers "when did the
	// CURRENT picked episode begin" — it is overwritten on every re-pick and
	// cleared on unpick; DroppedAt is stamped on drop and cleared on undrop.
	// Absence is a nil pointer here and SQL NULL in the column — never {}
	// and never "" (the REQ-TLE-006 discipline this SPEC follows).
	PickedAt  *string `json:"picked_at,omitempty"`
	DroppedAt *string `json:"dropped_at,omitempty"`
	// Classification is ADDITIVE (SPEC-TODO-CLASSIFY-DISPATCH-001
	// REQ-TCD-002): the creation-time judgment recorded through the add
	// path's decider seam — priority, blocked, mode, decider identity,
	// classified-at stamp, one-line reason — in ONE nullable field after the
	// Landing precedent, so a card with no classification marshals
	// byte-identically to the pre-SPEC record. Absence is a nil pointer here
	// and SQL NULL in the classification column — never `{}` and never `""`.
	// Every reader derives the defaults through EffectiveCardClassification
	// (REQ-TCD-014); no consumer tests the pointer itself.
	Classification *CardClassification `json:"classification,omitempty"`
	// PickedBy / LeaseExpiresAt are ADDITIVE (SPEC-TODO-CLAIM-LEASE-001
	// REQ-TCL-001): the claim lease's holder label and its RFC 3339 expiry,
	// `omitempty` after the stamp precedent so a card carrying neither
	// marshals byte-identically to before — the `todo list --json` golden
	// gate (REQ-TCL-014) rests on this. Absence is a nil pointer here and
	// SQL NULL in the column — never {} and never "" (REQ-TLE-006). A card
	// in any state other than picked carries neither field; the claim-family
	// operations are their only writers.
	PickedBy       *string `json:"picked_by,omitempty"`
	LeaseExpiresAt *string `json:"lease_expires_at,omitempty"`
	// Issuance is ADDITIVE (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-007): the
	// creation-time issuance attributes — spawn parent, closed-set origin,
	// size estimate, expected files, drop reason — in ONE nullable JSON
	// field after the lease precedent, so a card carrying none marshals
	// byte-identically to before. Absence is a nil pointer here and SQL NULL
	// in the column — never {} and never "" (REQ-TLE-006). Legacy cards are
	// never retrofitted.
	Issuance *BacklogIssuance `json:"issuance,omitempty"`
}

// BacklogIssuance is the issuance attribute record one card carries. Every
// key is optional; a card with no keys stores SQL NULL and serializes as
// absent (omitempty). Origin draws from the closed set the add path
// validates (IssuanceOrigins); Files are the explicitly recorded expected
// files — the hub chain (REQ-TCI-020) and the in-flight overlap read THEM,
// never a body-derived inference (D13).
type BacklogIssuance struct {
	SpawnedBy  string   `json:"spawned_by,omitempty"`
	Origin     string   `json:"origin,omitempty"`
	SizeLines  *int     `json:"size_lines,omitempty"`
	Files      []string `json:"files,omitempty"`
	DropReason string   `json:"drop_reason,omitempty"`
}

// IssuanceOrigins is the closed origin set (design §4.1). The add path
// refuses an --origin outside it; the list is the single home of the set.
var IssuanceOrigins = []string{
	"operator", "leader", "audit-finding", "follow-up",
	"ci-repair", "standing", "external", "split", "debt",
}

// IssuanceOriginValid reports whether origin is inside the closed set.
func IssuanceOriginValid(origin string) bool {
	for _, o := range IssuanceOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

// Relation values a finding may carry. The first two are MECHANICAL — the
// analyser measures them from card text alone; the last four are SEMANTIC —
// only a reader who understands what the cards mean can judge them, so they
// arrive through `todo relate`.
const (
	// BacklogRelationDuplicateForced records a card admitted despite an exact
	// text collision (`add --force`), so a forced duplicate leaves a trace.
	BacklogRelationDuplicateForced = "duplicate-forced"
	// BacklogRelationNearDuplicate records a measured text similarity at or
	// above the near-duplicate threshold.
	BacklogRelationNearDuplicate = "near-duplicate"
	// BacklogRelationContains records that one card's scope covers another's.
	BacklogRelationContains = "contains"
	// BacklogRelationAbsorbs records that one card could subsume another.
	// Recording it is ALL it does: folding one card into another is the act
	// the queue doctrine forbids by name, so it stays the operator's.
	BacklogRelationAbsorbs = "absorbs"
	// BacklogRelationReplaces records that one card supersedes another.
	BacklogRelationReplaces = "replaces"
	// BacklogRelationConflicts records that two cards pull against each
	// other. No resolution is proposed — both may be legitimate.
	BacklogRelationConflicts = "conflicts"
	// BacklogRelationBlocks records that the subject must land before the
	// related card can proceed (card t1309). Since
	// SPEC-RELATION-PICKUP-FILTER-001 the sequencing pair is no longer
	// purely observational: the todo --auto pickup selection
	// (internal/cli autoPickTargets) excludes the blocked-side card while
	// the finding is live, and todo relate refuses a write that would close
	// a waits-on cycle (BacklogRecord.WaitsOnClosesCycle). Factory-lease
	// consumption stays the separate adjudication named in the code (t1240).
	BacklogRelationBlocks = "blocks"
	// BacklogRelationDepends is the inverse spelling of blocks: the subject
	// waits on the related card. Consumed by the same two paths as blocks —
	// WaitsOnOf is the single direction normalization.
	BacklogRelationDepends = "depends"
)

// Source values a finding may carry.
const (
	// BacklogSourceMechanical marks a finding the text analyser produced.
	BacklogSourceMechanical = "mechanical"
	// BacklogSourceAgent marks a finding written through `todo relate`.
	BacklogSourceAgent = "agent"
	// BacklogSourceJev marks a finding a TypeSafe System One answer produced
	// (SPEC-JEV-CONSUMERS-001, REQ-JEVN-002).
	//
	// It is a THIRD constant rather than a reuse of either existing one
	// because the two partition the finding space by WHO OBSERVED the
	// relation, and a model answer is neither. Under `mechanical` a model
	// probability would render as a `score` indistinguishable from a measured
	// similarity; under `agent` it would clear the `machine-only` mark from
	// pairs nobody reviewed, making the queue assert a review that never
	// happened (REQ-JEVN-003). The second direction is the dangerous one,
	// because a wrongly-cleared mark is invisible by construction.
	BacklogSourceJev = "jev"
)

// BacklogSemanticRelations lists the six relations `todo relate` accepts.
// The two mechanical relations are deliberately absent: a caller must not be
// able to record a measurement it did not take. The sequencing pair
// (blocks / depends) is a judgement the operator or a dispatching agent
// makes, recorded so the sequencing stops living in card prose alone —
// recording one changes nothing (card t1309).
var BacklogSemanticRelations = []string{
	BacklogRelationContains,
	BacklogRelationAbsorbs,
	BacklogRelationReplaces,
	BacklogRelationConflicts,
	BacklogRelationBlocks,
	BacklogRelationDepends,
}

// BacklogFinding is one recorded relation between two cards.
//
// A finding is a RECORD and nothing else: no code path writes a card field as
// a consequence of one. That is a structural property rather than a
// convention — folding, reordering, dropping, and editing a card in response
// to a finding would each require code that does not exist, so none of them
// can happen by accident.
type BacklogFinding struct {
	SubjectID string  `json:"subject_id"`
	RelatedID string  `json:"related_id"`
	Relation  string  `json:"relation"`
	Source    string  `json:"source"`
	Score     float64 `json:"score"`
	Note      string  `json:"note"`
	At        string  `json:"at"`
	// Disposition is ADDITIVE (SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-010):
	// the operator's recorded disposition of a finding — accept/merge/reject.
	// Recording-only (D11): it never changes the card, the order, or the
	// pickup filter. Absence is a nil pointer here and SQL NULL in the
	// disposition column — never {} and never "" (REQ-TLE-006).
	Disposition *string `json:"disposition,omitempty"`
}

// Names reports whether the finding refers to id in either position.
func (f BacklogFinding) Names(id string) bool {
	return f.SubjectID == id || f.RelatedID == id
}

// WaitsOnOf normalizes a sequencing finding to one directed waits-on edge:
// the WAITER (the blocked side) and the card it waits on (its predecessor).
// `depends {S,R}` means S waits on R; `blocks {S,R}` means R waits on S
// (SPEC-RELATION-PICKUP-FILTER-001 spec.md B.3). Non-sequencing relations
// return ok=false — the pickup filter and the cycle guard consume exactly
// blocks and depends (REQ-RPF-006).
//
// @MX:ANCHOR: WaitsOnOf — the direction-normalization SSOT for the sequencing pair
// @MX:REASON: fan_in 4 (todo_auto pickup filter, todo_relate cycle guard, FindingsBlocking, WaitsOnClosesCycle); a second direction spelling would let the two consumers disagree about which side waits
// @MX:NOTE: single normalization point for both sequencing consumers —
// the pickup filter and the todo-relate cycle guard read direction ONLY
// through this function.
func WaitsOnOf(f BacklogFinding) (waiter, target string, ok bool) {
	switch f.Relation {
	case BacklogRelationDepends:
		return f.SubjectID, f.RelatedID, true
	case BacklogRelationBlocks:
		return f.RelatedID, f.SubjectID, true
	}
	return "", "", false
}

// FindingsBlocking returns the live findings whose blocked side (the
// waits-on waiter) names id. The finding's EXISTENCE in the live record is
// the whole unresolved predicate (REQ-RPF-001): a predecessor's done moves
// the finding into the archive through ArchiveCard, so resolution needs no
// card-state scan — and a dropped or held predecessor keeps the finding, so
// the successor conservatively keeps waiting (spec.md B.1).
func (r *BacklogRecord) FindingsBlocking(id string) []BacklogFinding {
	var out []BacklogFinding
	for _, f := range r.Findings {
		if waiter, _, ok := WaitsOnOf(f); ok && waiter == id {
			out = append(out, f)
		}
	}
	return out
}

// WaitsOnClosesCycle reports whether the record's live sequencing findings
// already connect target back to waiter — adding the waiter→target edge
// would then close a directed cycle (SPEC-RELATION-PICKUP-FILTER-001
// REQ-RPF-005). Same-pair opposite spellings normalize to the SAME edge
// through WaitsOnOf and never form a cycle (spec.md B.3).
func (r *BacklogRecord) WaitsOnClosesCycle(waiter, target string) bool {
	edges := map[string][]string{}
	for _, f := range r.Findings {
		if w, t, ok := WaitsOnOf(f); ok {
			edges[w] = append(edges[w], t)
		}
	}
	return waitsOnReaches(edges, target, waiter)
}

// waitsOnReaches reports whether to is reachable from from over the directed
// waits-on edges.
func waitsOnReaches(edges map[string][]string, from, to string) bool {
	seen := map[string]bool{from: true}
	queue := []string{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == to {
			return true
		}
		for _, next := range edges[cur] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

// SamePairAs reports whether two findings refer to the same UNORDERED pair.
// The comparison is unordered because a relation between two cards is a
// property of the pair, not of the direction it happened to be written in —
// an agent recording {b, a} answers a mechanical finding about {a, b}.
func (f BacklogFinding) SamePairAs(other BacklogFinding) bool {
	return (f.SubjectID == other.SubjectID && f.RelatedID == other.RelatedID) ||
		(f.SubjectID == other.RelatedID && f.RelatedID == other.SubjectID)
}

// BacklogArchivedFinding is one archived finding plus the index it occupied
// in Findings when its card was archived. The position is what makes a
// restore return the record to the bytes it had rather than to an
// equivalent-looking record with a different finding order.
type BacklogArchivedFinding struct {
	Finding  BacklogFinding `json:"finding"`
	Position int            `json:"position"`
}

// BacklogArchiveEntry is one archived card: the row itself, the findings that
// named it, and the position each occupied.
//
// The positions are load-bearing (REQ-TDG-002). `done` splices a row out of
// the middle of the queue, so a restore that appended would return the same
// SET of cards in a different ORDER — a record that reads correctly and is
// not the record the operator had. Carrying the index costs one integer and
// removes the whole class.
//
// The card's findings travel WITH it rather than in a parallel array: a
// finding that outlives its subject points at nothing (the reason `done`
// removed them in the first place), and keeping them inside the entry makes
// "restore the card and everything recorded about it" one move.
type BacklogArchiveEntry struct {
	Item     BacklogItem              `json:"item"`
	Position int                      `json:"position"`
	Findings []BacklogArchivedFinding `json:"findings"`
	// ArchivedAt is ADDITIVE (SPEC-TODO-TRANSITION-STAMPS-001 REQ-TST-002):
	// the archive-time stamp, `omitempty` after the Landing precedent. The
	// item's own stamps ride inside Item — ArchiveCard copies the row
	// wholesale — so the archived row is the card's final, readable home
	// (REQ-TST-007).
	ArchivedAt *string `json:"archived_at,omitempty"`
	// LandingVerdict is ADDITIVE (SPEC-TODO-TRANSITION-STAMPS-001
	// REQ-TST-008): the done-time landing query's answer — verdict, answering
	// ref, verdict time — persisted alongside, never instead of, the
	// operator-authored evidence in Item.Landing (REQ-TST-009). It carries no
	// SHA by construction; see landing_verdict.go.
	LandingVerdict *LandingVerdict `json:"landing_verdict,omitempty"`
}

// BacklogRecord is the backlog file's document shape. LastSeq is the
// persisted id high-water mark — additive top-level, absent in files that
// predate the field (derived from max present id on load). Findings is the
// second additive top-level field, following the same precedent: absent in
// older files, and always rendered as an array — never null, never an
// omitted key — so a reader never has to tell "no findings" apart from "no
// such feature".
// Archived is the third additive top-level field, following the same
// precedent again: absent in older files, always rendered as an array, and
// invisible to every live-queue reader by CONSTRUCTION rather than by filter
// — a reader that iterates Items cannot see an archived row, so no state
// enum, no listing filter, and no analysis input had to change for it
// (spec.md §B.1).
type BacklogRecord struct {
	ProjectUUID *string               `json:"project_uuid"`
	Version     int                   `json:"version"`
	LastSeq     int                   `json:"last_seq"`
	Items       []BacklogItem         `json:"items"`
	Findings    []BacklogFinding      `json:"findings"`
	Archived    []BacklogArchiveEntry `json:"archived"`
	Runtime     TodoRuntime           `json:"runtime"`
}

// ArchivedIndex returns the index of id in the archive, or -1.
func (r *BacklogRecord) ArchivedIndex(id string) int {
	for i, entry := range r.Archived {
		if entry.Item.ID == id {
			return i
		}
	}
	return -1
}

// itemIndex returns the index of id among the live items, or -1.
func (r *BacklogRecord) itemIndex(id string) int {
	for i, it := range r.Items {
		if it.ID == id {
			return i
		}
	}
	return -1
}

// ArchiveCard moves the addressed card and every finding naming it out of the
// live queue and into the archive, recording the position each occupied.
//
// This is what `done` performs instead of discarding. The findings move with
// the card for the reason they were previously deleted with it — a finding
// that outlives its subject points at a card the operator can no longer see —
// but moving rather than deleting is what makes the act reversible, and the
// relations are operator judgment that a discard loses silently (spec.md §A.2).
func (r *BacklogRecord) ArchiveCard(id string) error {
	if r.ArchivedIndex(id) >= 0 {
		return fmt.Errorf("backlog item %s is already archived", id)
	}
	at := r.itemIndex(id)
	if at < 0 {
		return fmt.Errorf("no backlog item %s", id)
	}

	entry := BacklogArchiveEntry{
		Item:     r.Items[at],
		Position: at,
		Findings: []BacklogArchivedFinding{},
	}
	// REQ-TST-007: the archive is the row's final home and carries its own
	// archive-time stamp. The item's picked_at / dropped_at stamps ride
	// inside the copied Item, as they stood at archive time.
	now := time.Now().UTC().Format(time.RFC3339)
	entry.ArchivedAt = &now
	kept := make([]BacklogFinding, 0, len(r.Findings))
	for i, f := range r.Findings {
		if f.Names(id) {
			entry.Findings = append(entry.Findings, BacklogArchivedFinding{Finding: f, Position: i})
			continue
		}
		kept = append(kept, f)
	}
	r.Findings = kept
	r.Items = append(r.Items[:at:at], r.Items[at+1:]...)
	r.Archived = append(r.Archived, entry)
	return nil
}

// RestoreCard moves an archived card and its findings back into the live
// queue at the positions they held, and REMOVES the archive entry so each row
// has exactly one home (REQ-TDG-016).
//
// It refuses when the id has since been reissued to a different live card
// (REQ-TDG-013): overwriting the live row would destroy a card nobody asked
// to lose, and splitting one id across two cards is worse than the mistake
// the restore is undoing.
func (r *BacklogRecord) RestoreCard(id string) error {
	at := r.ArchivedIndex(id)
	if at < 0 {
		return fmt.Errorf("no archived backlog item %s", id)
	}
	if live := r.itemIndex(id); live >= 0 {
		return fmt.Errorf("cannot restore %s: the id has been reissued to a live card %q — refusing rather than overwriting it",
			id, r.Items[live].Text)
	}

	entry := r.Archived[at]
	r.Items = insertBacklogItem(r.Items, entry.Item, entry.Position)
	// Ascending position order: each insertion shifts the ones after it, so
	// restoring low indexes first reproduces the original layout.
	sorted := append([]BacklogArchivedFinding(nil), entry.Findings...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Position < sorted[j].Position })
	for _, af := range sorted {
		deferred := false
		if r.itemIndex(af.Finding.SubjectID) < 0 || r.itemIndex(af.Finding.RelatedID) < 0 {
			// Keep a suspended relation with its still-archived endpoint; it
			// becomes live only after both cards are restored.
			for i := range r.Archived {
				if i != at && af.Finding.Names(r.Archived[i].Item.ID) {
					r.Archived[i].Findings = append(r.Archived[i].Findings, af)
					deferred = true
					break
				}
			}
		}
		if deferred {
			continue
		}
		// Legacy records may already name a deleted endpoint. Preserve their
		// evidence rather than silently deleting it during restoration.
		r.Findings = insertBacklogFinding(r.Findings, af.Finding, af.Position)
	}
	r.Archived = append(r.Archived[:at:at], r.Archived[at+1:]...)
	return nil
}

// insertBacklogItem inserts it at pos, clamping into range. Clamping rather
// than failing is deliberate: the round trip this SPEC guarantees is
// `done` immediately followed by `undone`, where the position is exact. A
// restore taken after other cards moved has no exact answer, and refusing
// there would make the archive unrecoverable for the operator who waited.
func insertBacklogItem(items []BacklogItem, it BacklogItem, pos int) []BacklogItem {
	if pos < 0 {
		pos = 0
	}
	if pos > len(items) {
		pos = len(items)
	}
	out := make([]BacklogItem, 0, len(items)+1)
	out = append(out, items[:pos]...)
	out = append(out, it)
	return append(out, items[pos:]...)
}

// insertBacklogFinding inserts f at pos, clamping into range (see above).
func insertBacklogFinding(findings []BacklogFinding, f BacklogFinding, pos int) []BacklogFinding {
	if pos < 0 {
		pos = 0
	}
	if pos > len(findings) {
		pos = len(findings)
	}
	out := make([]BacklogFinding, 0, len(findings)+1)
	out = append(out, findings[:pos]...)
	out = append(out, f)
	return append(out, findings[pos:]...)
}

// HasFindingTuple reports whether a finding carrying the same
// {subject, related, relation, source} tuple is already recorded. The
// timestamp and the note are deliberately outside the key: re-running the
// analyser must not stack a second copy of a relation it already recorded,
// or the operator's listing fills with duplicates of one measurement and
// stops being read.
func (r *BacklogRecord) HasFindingTuple(f BacklogFinding) bool {
	for _, existing := range r.Findings {
		if existing.SubjectID == f.SubjectID && existing.RelatedID == f.RelatedID &&
			existing.Relation == f.Relation && existing.Source == f.Source {
			return true
		}
	}
	return false
}

// AppendFindingOnce records f unless its tuple is already present, reporting
// whether it was appended.
func (r *BacklogRecord) AppendFindingOnce(f BacklogFinding) bool {
	if r.HasFindingTuple(f) {
		return false
	}
	r.Findings = append(r.Findings, f)
	return true
}

// FindingsNaming returns the findings referring to id, each paired with its
// 1-based index in the record — the index `todo unrelate` addresses.
func (r *BacklogRecord) FindingsNaming(id string) (findings []BacklogFinding, indexes []int) {
	for i, f := range r.Findings {
		if f.Names(id) {
			findings = append(findings, f)
			indexes = append(indexes, i+1)
		}
	}
	return findings, indexes
}

// RemoveFindingsNaming drops every finding referring to id and returns how
// many were dropped. Called when a card leaves the queue: a finding that
// outlives its subject points at nothing, and the listing would render a
// relation to a card the operator can no longer see.
func (r *BacklogRecord) RemoveFindingsNaming(id string) int {
	kept := make([]BacklogFinding, 0, len(r.Findings))
	for _, f := range r.Findings {
		if !f.Names(id) {
			kept = append(kept, f)
		}
	}
	removed := len(r.Findings) - len(kept)
	r.Findings = kept
	return removed
}

// HasAgentFindingForPair reports whether any agent-sourced finding names the
// same unordered pair as f. It is the predicate behind the `machine-only`
// mark: the mark records the ABSENCE of an agent-sourced record for the
// pair, never that any review took place.
func (r *BacklogRecord) HasAgentFindingForPair(f BacklogFinding) bool {
	for _, existing := range r.Findings {
		if existing.Source == BacklogSourceAgent && existing.SamePairAs(f) {
			return true
		}
	}
	return false
}

// HasFindingForPairAnySource reports whether ANY finding — of any source —
// names the same unordered pair as f with the same relation
// (SPEC-JEV-CONSUMERS-001, REQ-JEVN-006 half (a)).
//
// It exists because neither existing path can express the settled precedence
// rule, and the reason is the same property that makes the rule's other half
// free. HasFindingTuple's key is {subject, related, relation, SOURCE} —
// ordered AND source-inclusive — so a Jev finding never matches a mechanical
// or agent one and AppendFindingOnce would never suppress it; and
// AppendFindingOnce never calls SamePairAs, so the unordered comparison is
// not on the append path at all.
//
// The asymmetry this predicate enforces is deliberate: a model signal must
// never suppress a measurement or a judgement (that half is
// AppendFindingOnce's unchanged default), and must never be silently
// suppressed by one either — an un-suppressed append is indistinguishable
// from a correct one at every surface a reader looks at, so the suppression
// has to be expressed rather than inherited.
//
// Source is deliberately outside the key. Relation is deliberately inside it:
// a `contains` recorded about a pair does not answer a `near-duplicate`
// question about the same pair.
func (r *BacklogRecord) HasFindingForPairAnySource(f BacklogFinding) bool {
	for _, existing := range r.Findings {
		if existing.Relation == f.Relation && existing.SamePairAs(f) {
			return true
		}
	}
	return false
}

// BacklogStore guards one backlog file. Load is lock-free; every mutation
// goes through Mutate, which holds the sibling lock across the whole
// read-modify-write.
type BacklogStore struct {
	path string
}

// NewBacklogStore constructs a handle without touching operator files.
// Legacy lock cleanup belongs to the adopting path, never a pure read.
func NewBacklogStore(path string) *BacklogStore {
	return &BacklogStore{path: path}
}

// QueuedCount returns the number of items in state "queued", failing open to
// 0 (the store already reads a missing file as an empty queue). It is the
// shared count shape both the queue notice and the factory leader loop render
// from, so the notice and the queue command cannot disagree about what
// "waiting" means: state queued and nothing else — a picked card is in
// flight on another lane, a dropped card was discarded, and a finished card
// is removed outright.
func (s *BacklogStore) QueuedCount() int {
	// PURE read (REQ-TOSQ-009): this is the statusline's per-render path and
	// the SessionStart notice's, both of which run on surfaces that must never
	// move operator bytes. It therefore reads whichever layout it finds and
	// NEVER migrates — migration is the adopting path's act (Load / Mutate),
	// which is where the queue lock is already in play.
	layout := inspectBacklogLayout(s.path)
	if layout.dbExists {
		eng, err := openBacklogReader(backlogSQLitePath(s.path))
		if err != nil {
			return 0
		}
		defer func() { _ = eng.close() }()
		n, err := eng.countQueued(context.Background())
		if err != nil {
			return 0
		}
		return n
	}
	if !layout.jsonExists {
		return 0
	}
	rec, err := loadLegacyBacklogJSON(s.path)
	if err != nil {
		return 0
	}
	n := 0
	for _, it := range rec.Items {
		if it.State == BacklogStateQueued {
			n++
		}
	}
	return n
}

// BacklogStateCounts is the queue reduced to what a glance needs: how much
// work is in flight and how much is waiting. Dropped items are deliberately
// not counted — they are history, and a number that only ever grows is noise.
type BacklogStateCounts struct {
	Picked    int
	Queued    int
	Available bool
}

// @MX:ANCHOR: [AUTO] BacklogCountsForRoot — the statusline's per-render read
// @MX:REASON: expected fan_in >= 3 (statusline render, queue notice, factory loop); it runs once per status render, so a non-constant-cost implementation here puts every render on a path that grows with the queue
//
// BacklogCountsForRoot counts items by state under a project root.
//
// PURE and fail-open, both load-bearing. Pure because this runs on a
// per-render process that must never perform the one-time storage cutover or
// directory relocation; fail-open because an unreadable queue must render
// nothing rather than an authoritative-looking zero — "no cards" and "could
// not read the cards" are different claims, and Available is what separates
// them.
//
// Constant cost on the database layout: three aggregates over an indexed
// column, not a whole-document parse (REQ-TOSQ-009, C-2).
func BacklogCountsForRoot(root string) BacklogStateCounts {
	if root == "" {
		return BacklogStateCounts{}
	}
	path := BacklogPathForRoot(root)
	layout := inspectBacklogLayout(path)
	if layout.dbExists {
		eng, err := openBacklogReader(backlogSQLitePath(path))
		if err != nil {
			return BacklogStateCounts{}
		}
		defer func() { _ = eng.close() }()
		picked, queued, err := eng.countByState(context.Background())
		if err != nil {
			return BacklogStateCounts{}
		}
		return BacklogStateCounts{Picked: picked, Queued: queued, Available: true}
	}
	if !layout.jsonExists {
		return BacklogStateCounts{}
	}
	rec, err := loadLegacyBacklogJSON(path)
	if err != nil {
		return BacklogStateCounts{}
	}
	counts := BacklogStateCounts{Available: true}
	for _, it := range rec.Items {
		switch it.State {
		case BacklogStatePicked:
			counts.Picked++
		case BacklogStateQueued:
			counts.Queued++
		}
	}
	return counts
}

// QueuedBacklogCountForRoot is the one-call form of QueuedCount under a
// project root. An empty root reads as 0 (no project, no queue) — the
// fail-open posture the notice builders already hold.
func QueuedBacklogCountForRoot(root string) int {
	if root == "" {
		return 0
	}
	return NewBacklogStore(BacklogPathForRoot(root)).QueuedCount()
}

// Path returns the backlog file's path.
func (s *BacklogStore) Path() string { return s.path }

// EnginePath returns the storage engine's artifact path — the queue file's
// sibling database (diagnostics and tests). It is the physical carrier tests
// stat when they mean "the queue's own file"; Path() still names the legacy
// document the downgrade route regenerates.
func (s *BacklogStore) EnginePath() string { return backlogSQLitePath(s.path) }

// LockPath returns the sibling lock artifact's path (diagnostics and tests).
func (s *BacklogStore) LockPath() string {
	return filepath.Join(filepath.Dir(s.path), backlogLockFileName)
}

// @MX:ANCHOR: [AUTO] Load — the lock-free read every backlog verb renders from
// @MX:REASON: expected fan_in >= 3 (M2 list/next/done verbs + tests); the sole load path on the file
//
// Load reads the backlog file without taking the lock. A missing file is an
// empty queue, never an error (REQ-TODO-012). A malformed file surfaces as a
// parse error with the file untouched — there is no repair-on-load path; the
// operator's queued intent is the one thing that cannot be regenerated.
func (s *BacklogStore) Load() (*BacklogRecord, error) {
	rec, err := s.load()
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// @MX:ANCHOR: [AUTO] LoadPure — the read that never moves bytes
// @MX:REASON: expected fan_in >= 3 (web console read seam, statusline, future read-only surfaces); it is the pure half of the pure-vs-adopting split, and a read-only surface calling Load instead would migrate the operator queue from a page render
//
// LoadPure reads the queue WITHOUT adopting: it serves whichever layout it
// finds and performs no migration, no quarantine, and no relocation on any
// branch. It is the read for surfaces that must never mutate operator state —
// the web console's page render and the statusline — mirroring the existing
// ResolveTodoQueueRoot / ResolveTodoQueueRootAdopting split one layer down
// (REQ-TOSQ-010, REQ-WTQ-001/004).
//
// Load, by contrast, adopts: the `moai todo` verbs are where the one-time
// cutover belongs, because that is where the queue lock is already in play.
func (s *BacklogStore) LoadPure() (*BacklogRecord, error) {
	// @MX:NOTE: [TID:PURE] A pure read projects nullable identities but never issues, stamps, or backfills them.
	layout := inspectBacklogLayout(s.path)
	if !layout.dbExists {
		if !layout.jsonExists {
			return &BacklogRecord{
				Version:  backlogVersion,
				Items:    []BacklogItem{},
				Findings: []BacklogFinding{},
			}, nil
		}
		return loadLegacyBacklogJSON(s.path)
	}
	eng, err := openBacklogReader(backlogSQLitePath(s.path))
	if err != nil {
		return nil, err
	}
	defer func() { _ = eng.close() }()
	return eng.readRecord(context.Background())
}

// load is Load without the record copy — the shared read both the lock-free
// verbs and the locked mutation path call. A missing queue in either layout is
// an empty record, never an error.
func (s *BacklogStore) load() (*BacklogRecord, error) {
	eng, err := s.openEngine(false)
	if err != nil {
		return nil, err
	}
	defer func() { _ = eng.close() }()
	return eng.readRecord(context.Background())
}

// @MX:ANCHOR: [AUTO] openEngine — the layout resolver every store operation enters through
// @MX:REASON: expected fan_in >= 3 (load, Mutate, export); it is the single place the migration state machine fires, so a second entry point would let a queue be migrated outside the lock
//
// openEngine resolves the storage layout (design.md §4), performs the one-time
// lazy migration when the legacy JSON is the only thing present, and returns
// an open engine. lockHeld says whether the caller already holds the queue
// lock; when it does not and a migration is required, this acquires the lock
// for the migration's duration so concurrent factory lanes serialize on the
// same artifact they always have (REQ-TOSQ-008).
func (s *BacklogStore) openEngine(lockHeld bool) (*backlogEngine, error) {
	_ = os.Remove(filepath.Join(filepath.Dir(s.path), legacyBacklogLockFileName))
	switch layout := inspectBacklogLayout(s.path); {
	case layout.dbExists && layout.jsonExists:
		// State D — a database beside a backlog.json. The database is
		// authoritative on every path either way, but the json's identity
		// decides what happens to it: pre-cutover legacy left by a crash
		// between the migration commit and the quarantine is finished
		// best-effort (REQ-TOSQ-013), while an export written for a downgrade
		// is left exactly where the operator put it. The in-flight marker
		// inside the database is what tells them apart.
		completeInterruptedQuarantine(s.path)
	case !layout.dbExists && layout.jsonExists:
		// State B — migrate under the lock.
		if err := s.migrateUnderLock(lockHeld); err != nil {
			return nil, err
		}
	}
	return openBacklogEngine(backlogSQLitePath(s.path))
}

// migrateUnderLock runs the migration with the queue lock held, acquiring it
// first when the caller does not already hold it.
//
// This is the one place in the family whose errors name Path() rather than
// EnginePath(), and it is deliberate: migration reaches here only in State B
// (`!dbExists && jsonExists`), where the legacy document is the file being
// read and is present. Rewriting this to EnginePath() for consistency with
// the Mutate family would make the message wrong in the other direction —
// naming a database that does not exist yet (card t910).
func (s *BacklogStore) migrateUnderLock(lockHeld bool) error {
	if lockHeld {
		return migrateLegacyBacklog(s.path)
	}
	lock, err := s.acquireLock()
	if err != nil {
		return err
	}
	migErr := migrateLegacyBacklog(s.path)
	return joinBacklogReleaseErr(migErr, lock.Release(), s.path)
}

// @MX:WARN: [AUTO] Mutate — cross-process lock held across the entire read-modify-write
// @MX:REASON: a mutation that loads or writes outside the lock reintroduces the lost-update race this SPEC kills (REQ-TODO-006/008)
//
// Mutate is THE guarded mutation entry point: it acquires the sibling lock,
// loads the record, applies mutate in place, and atomically writes the
// result. Returning an error from mutate aborts with the file unchanged.
// Ids must be issued from rec.LastSeq inside the callback — the high-water
// mark is normalized (max of persisted and max-present) before mutate runs,
// so a version-1 file predating last_seq or a hand-edited low value both
// resolve before issuance. Release errors are JOINED into the result rather
// than discarded: on Windows release removes the artifact, so a silent
// release failure would block every later writer.
//
// Every error this family raises names EnginePath(), not Path(). Path() is
// the legacy `backlog.json` document the downgrade route regenerates, and the
// engine never writes it — in the steady state it is simply absent, so naming
// it hands the operator a file that is not there. Card t899 lost half an
// investigation to that path before the real cause was found (card t910).
func (s *BacklogStore) Mutate(mutate func(*BacklogRecord) error) error {
	return s.WithLock(func(l *LockedBacklog) error {
		return l.Mutate(mutate)
	})
}

// @MX:ANCHOR: [AUTO] WithLock — the one place the queue's cross-process lock is taken for a mutation
// @MX:REASON: expected fan_in >= 3 (Mutate, the factory lease section, tests); a second acquisition path would let two sections order differently against the same lock
// @MX:WARN: [AUTO] WithLock — the lock is a non-reentrant flock on its own descriptor, so calling the public Mutate inside fn waits out the whole budget and errors
// @MX:REASON: inside fn mutate only through the handle (LockedBacklog.Mutate); the handle must not outlive fn, since the lock is released when fn returns
//
// WithLock holds the queue's lock across fn (SPEC-FACTORY-ATOMIC-LEASE-001
// plan D1): it acquires through acquireLock — the same wait policy and the
// same timeout error as Mutate — refuses a relocated queue exactly as Mutate
// does, runs fn with a LockedBacklog, and releases on every exit, a panic
// included. A release failure is JOINED into the result, as in Mutate. fn's own
// error is returned unwrapped, so callers can match it with errors.Is.
func (s *BacklogStore) WithLock(fn func(*LockedBacklog) error) (err error) {
	lock, err := s.acquireLock()
	if err != nil {
		return err
	}
	defer func() {
		err = joinBacklogReleaseErr(err, lock.Release(), s.EnginePath())
	}()
	if target, readErr := os.ReadFile(filepath.Join(filepath.Dir(s.path), backlogRetiredFileName)); readErr == nil {
		return fmt.Errorf("%w: %s", ErrBacklogRelocated, target)
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	return fn(&LockedBacklog{s: s})
}

// LockedBacklog is the handle WithLock lends its callback while the queue lock
// is held. It offers the non-adopting read and the guarded mutation, and no
// adopting read: BacklogStore.Load adopts (migrates, relocates), and the lease
// path it serves reads with LoadPure, which never does (plan D1, mutant MU8).
type LockedBacklog struct {
	s *BacklogStore
}

// LoadPure reads the queue without adopting it, a fresh read each call.
func (l *LockedBacklog) LoadPure() (*BacklogRecord, error) {
	return l.s.LoadPure()
}

// Claim is the lock-held form of BacklogStore.Claim: the CAS+lease body is
// identical, but the caller already holds the queue lock and needs the
// claim and its follow-up writes (the dispatch binding) to land under that
// one held lock — a claim whose binding update can interleave with a
// concurrent completion would let the old run's approval close new work
// (SPEC-FACTORY-COMPLETION-RECOVERY-001 review round-16 P1-1).
func (l *LockedBacklog) Claim(holder string) (*BacklogClaim, error) {
	if strings.TrimSpace(holder) == "" {
		holder = BacklogOperatorHolder
	}
	var result *BacklogClaim
	err := l.Mutate(func(rec *BacklogRecord) error {
		res, err := claimBacklogMutation(rec, holder, time.Now().UTC())
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Mutate is today's Mutate body without the lock acquisition: it loads the
// record, applies mutate in place, and atomically writes the result. Returning
// an error from mutate aborts with the file unchanged.
func (l *LockedBacklog) Mutate(mutate func(*BacklogRecord) error) error {
	return l.s.mutateLocked(mutate)
}

// mutateLocked is the read-modify-write body of Mutate; the caller holds the
// queue lock.
func (s *BacklogStore) mutateLocked(mutate func(*BacklogRecord) error) error {
	// @MX:WARN: [TID:TX] Identity schema, UUID backfill, and the card record must commit in one writer transaction.
	// @MX:REASON: [TID:TX] MaxOpenConns(1) forbids e.db re-entry while that transaction is active; every identity query uses its *sql.Tx.
	eng, err := s.openEngine(true)
	if err != nil {
		return err
	}
	defer func() { _ = eng.close() }()

	ctx := context.Background()
	rec, err := eng.readRecord(ctx)
	if err != nil {
		return err
	}
	archiveBefore, err := json.Marshal(rec.Archived)
	if err != nil {
		return err
	}
	if err := mutate(rec); err != nil {
		return fmt.Errorf("mutate backlog %s: mutation refused: %w", s.EnginePath(), err)
	}
	// Re-normalize post-mutate: a callback may append or rewrite items, and
	// the written high-water mark must clear every present id.
	normalizeBacklogRecord(rec)
	archiveAfter, err := json.Marshal(rec.Archived)
	if err != nil {
		return err
	}
	return eng.writeRecordArchive(ctx, rec, !bytes.Equal(archiveBefore, archiveAfter))
}

// joinBacklogReleaseErr folds a lock-release failure into the mutation's own
// result. It JOINS rather than overwrites, and joins in both directions: a
// release failure that arrives alongside a mutation failure is the dangerous
// case, not the harmless one. Reporting only the mutation error there hides a
// wedged lock behind an unrelated message — on Windows release removes the
// artifact, so the survivor blocks every later writer while the operator
// retries against what looks like a transient mutation fault.
//
// Returns nil only when both are nil, so the caller's `err` stays clean on the
// success path.
func joinBacklogReleaseErr(mutErr, relErr error, path string) error {
	if relErr == nil {
		return mutErr
	}
	return errors.Join(mutErr, fmt.Errorf("mutate backlog %s: lock release failed: %w", path, relErr))
}

// LockWaitBudget returns the queue lock's wait budget — the elapsed window a
// writer polls for the lock before giving up. It is the accessor a bound over
// the lock derives from (SPEC-FACTORY-ATOMIC-LEASE-001 plan D1/D2), so the
// bound never copies the number.
func LockWaitBudget() time.Duration { return stateLockWaitBudget }

// @MX:ANCHOR: [AUTO] Add — the id-issuing append every add-path verb calls
// @MX:REASON: expected fan_in >= 3 (M2 add verb, tests, future importers); the only id issuer, and issuance outside the lock would mint duplicates
//
// Add appends text as a new queued card and returns the issued item with its
// 1-based position among the queued cards. The id comes from the persisted
// high-water mark inside the locked mutation, so a removed card's id is
// never reused (REQ-TODO-008).
func (s *BacklogStore) Add(text string) (*BacklogItem, int, error) {
	return s.addWithCardUUID(text, nil)
}

// addWithCardUUID is the identity-aware form used by GTD publication. A
// caller-supplied UUID makes a queue commit discoverable after a crash that
// occurs before the GTD link is recorded. Ordinary todo callers retain the
// existing identity issuer by passing nil through Add.
func (s *BacklogStore) addWithCardUUID(text string, cardUUID *string) (*BacklogItem, int, error) {
	// @MX:NOTE: [TID:RETURN] Add reads back the committed row so its card_uuid is exactly the persisted identity.
	var item BacklogItem
	var pos int
	err := s.Mutate(func(rec *BacklogRecord) error {
		rec.LastSeq++
		item = BacklogItem{
			ID:       fmt.Sprintf("t%d", rec.LastSeq),
			Text:     text,
			AddedAt:  time.Now().UTC().Format(time.RFC3339),
			State:    BacklogStateQueued,
			CardUUID: cloneIdentity(cardUUID),
		}
		rec.Items = append(rec.Items, item)
		// REQ-TCD-005: the sort is re-established inside the same locked
		// write as every add that may change it — an append to a sorted queue
		// leaves it sorted (an unclassified append ranks normal/non-blocked
		// and may move ahead of existing low-priority cards). The printed
		// position is the SORTED queued position (REQ-TCD-006), which an
		// append to the end no longer implies.
		rec.SortByClassification()
		pos = rec.QueuedPosition(item.ID)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	record, err := s.LoadPure()
	if err != nil {
		return nil, 0, err
	}
	for i := range record.Items {
		if record.Items[i].ID == item.ID {
			persisted := record.Items[i]
			return &persisted, pos, nil
		}
	}
	return nil, 0, fmt.Errorf("added backlog item %s missing from committed record", item.ID)
}

// Claim-family sentinels (SPEC-TODO-CLAIM-LEASE-001). Each is a DISTINCT
// indication so the CLI can map them onto separate exit surfaces
// (REQ-TCL-006's raced indication, REQ-TCL-012's no-card code, the renew
// refusals REQ-TCL-008 states).
var (
	// ErrClaimRaced reports that the claim lost a race or a live lease
	// guards every candidate: the queue holds picked cards and none of them
	// is claimable. The record is byte-identical — Mutate wrote nothing.
	ErrClaimRaced = errors.New("todo queue claim raced: the queue's cards are already picked (live lease or concurrent claim)")
	// ErrClaimNoCard reports that no eligible card exists at all — no queued
	// card, and no picked card either (queue empty, or only held/dropped
	// cards). Non-eligible states refuse by POSITIVE enumeration
	// (REQ-TCL-011): only state=='queued' is ever selected, so a state added
	// later is refused by the same fall-through.
	ErrClaimNoCard = errors.New("todo queue claim: no eligible card (nothing queued)")
	// ErrLeaseExpired reports a renew against a lapsed lease. The
	// expiry-first return has COMMITTED (the homestate committedRefusal
	// shape); the card is queued again and the renew did not extend.
	ErrLeaseExpired = errors.New("todo queue lease expired")
	// ErrLeaseHolder reports a renew by a label that does not hold the
	// lease — refused with no change (REQ-TCL-008).
	ErrLeaseHolder = errors.New("todo queue lease holder refused")
)

// BacklogOperatorHolder is the holder label a bare (non---lane) claim
// carries — the operator's own act, in the same actor vocabulary the
// factory record's operator decisions use.
const BacklogOperatorHolder = "operator"

// BacklogReclamation is one expired lease the expiry-first pass returned to
// queued. It is the C5 audit surface: id plus the previous holder, rendered
// by the CLI/MCP claim output — no events table (C5).
type BacklogReclamation struct {
	ItemID     string
	PrevHolder string
	Text       string
}

// BacklogClaim is the outcome of a successful claim-family operation: the
// claimed card and the reclamation lines this operation's expiry-first pass
// produced (usually empty; a claim that reclaims takes the reclaimed card
// itself, since it is the oldest after the pass).
type BacklogClaim struct {
	Item      BacklogItem
	Reclaimed []BacklogReclamation
}

// backlogLeaseExpired reports whether the item's lease has lapsed at now.
// Only a picked card can hold a lease; C4 pins an unparseable expiry as
// EXPIRED — a corrupt value cannot be trusted as a live lease (the
// card_record.go LeaseExpired default this SPEC adopts).
func backlogLeaseExpired(it *BacklogItem, now time.Time) bool {
	if it.State != BacklogStatePicked || it.LeaseExpiresAt == nil {
		return false
	}
	at, err := time.Parse(time.RFC3339, *it.LeaseExpiresAt)
	if err != nil {
		return true
	}
	return !now.Before(at)
}

// clearBacklogLease applies reclamation's cleared-field set: the lease
// fields (picked_by, lease_expires_at) plus the pick-time picked_at and
// spec_id — the generalized unpick (REQ-TCL-009; the operator's manual
// unpick keeps its existing semantics).
func clearBacklogLease(it *BacklogItem) {
	it.PickedBy = nil
	it.LeaseExpiresAt = nil
	it.PickedAt = nil
	it.SpecID = nil
}

// reclaimExpiredBacklogLeases runs the expiry-first pass over the record:
// every picked card whose lease has lapsed (C4: unparseable counts) returns
// to queued with the cleared-field set, and each return is surfaced as a
// BacklogReclamation for the human audit line (C5). Only LAPSED cards are
// mutated — a live lease is never crossed (REQ-TCL-007).
func reclaimExpiredBacklogLeases(rec *BacklogRecord, now time.Time) []BacklogReclamation {
	var out []BacklogReclamation
	for i := range rec.Items {
		it := &rec.Items[i]
		if !backlogLeaseExpired(it, now) {
			continue
		}
		prev := ""
		if it.PickedBy != nil {
			prev = *it.PickedBy
		}
		clearBacklogLease(it)
		it.State = BacklogStateQueued
		out = append(out, BacklogReclamation{ItemID: it.ID, PrevHolder: prev, Text: it.Text})
	}
	return out
}

// @MX:ANCHOR: [AUTO] Claim — the atomic compare-and-set card claim every claim surface calls
// @MX:REASON: expected fan_in >= 3 (CLI claim verb, MCP todo_claim mirror, reclamation paths, tests); the whole operation rides ONE Mutate so the flock — not a version column (C2) — supplies the CAS atomicity, and a claim routed around it would race the read-modify-write this method serializes
//
// @MX:WARN: [TID:TX] claim-family engine access lives ONLY inside the Mutate callback
// @MX:REASON: [TID:TX] a row-level update issued outside the callback's whole-record write races writeRecordArchive's delete-and-rewrite in BOTH directions (silent lost update, REQ-TCL-004); every claim statement must run inside this Mutate or not at all
//
// Claim takes the oldest queued card (stored order — the seq ORDER the
// reader returns) as one compare-and-set inside a single Mutate: the
// callback's freshly-loaded record IS the post-race state, so the state
// predicate on it is the CAS and the flock is the atomicity (C2). The pass
// runs expiry-first (REQ-TCL-009): lapsed leases return to queued before
// selection, so a claim immediately re-takes a card whose holder lapsed.
// The winner's card is stamped picked_by=holder, lease_expires_at=
// now+DefaultFactoryLeaseDuration, picked_at=now. Failures write nothing:
// a live-leased queue refuses with ErrClaimRaced (REQ-TCL-006), an empty
// one with ErrClaimNoCard. An empty holder reads as the operator's own act.
func (s *BacklogStore) Claim(holder string) (*BacklogClaim, error) {
	if strings.TrimSpace(holder) == "" {
		holder = BacklogOperatorHolder
	}
	var result *BacklogClaim
	err := s.Mutate(func(rec *BacklogRecord) error {
		res, err := claimBacklogMutation(rec, holder, time.Now().UTC())
		if err != nil {
			return err
		}
		result = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// claimBacklogMutation is the CAS+lease claim body shared by the
// store-level Claim (self-locking) and the lock-held LockedBacklog.Claim.
func claimBacklogMutation(rec *BacklogRecord, holder string, now time.Time) (*BacklogClaim, error) {
	expiry := now.Add(config.DefaultFactoryLeaseDuration).Format(time.RFC3339)
	stamp := now.Format(time.RFC3339)
	reclaimed := reclaimExpiredBacklogLeases(rec, now)
	for i := range rec.Items {
		// POSITIVE enumeration (REQ-TCL-011): only state=='queued' is
		// claimable; every other state — a future one included — falls
		// through the selection and refuses below.
		if rec.Items[i].State != BacklogStateQueued {
			continue
		}
		it := &rec.Items[i]
		it.State = BacklogStatePicked
		it.PickedBy = &holder
		it.LeaseExpiresAt = &expiry
		it.PickedAt = &stamp
		return &BacklogClaim{Item: *it, Reclaimed: reclaimed}, nil
	}
	if reclaimed != nil {
		// Unreachable: a reclaim leaves a queued card, which the loop
		// above would have claimed. Kept as the race-free guarantee that
		// a claim over its own reclamation never reports failure.
		return &BacklogClaim{Item: rec.Items[0], Reclaimed: reclaimed}, nil
	}
	for _, it := range rec.Items {
		if it.State == BacklogStatePicked {
			// A picked card survived the expiry-first pass: its lease is
			// live — this claim lost (REQ-TCL-006).
			return nil, fmt.Errorf("%w: %s is held by %s until %s", ErrClaimRaced,
				it.ID, derefOr(it.PickedBy, "unknown"), derefOr(it.LeaseExpiresAt, "unknown"))
		}
	}
	return nil, ErrClaimNoCard
}

// @MX:ANCHOR: [AUTO] RenewLease — the holder-checked lease extension
// @MX:REASON: expected fan_in >= 3 (CLI claim --renew, MCP mirror, factory stage integration, tests); a renewal that skipped the holder predicate would let any session extend another holder's lease (REQ-TCL-007/008)
//
// @MX:WARN: [TID:TX] claim-family engine access lives ONLY inside the Mutate callback
// @MX:REASON: [TID:TX] same chokepoint discipline as Claim — the renewal and its expiry-first pass commit as one whole-record write or not at all
//
// RenewLease extends the holder's lease by the lease duration and changes
// no other field (REQ-TCL-008). The pass runs expiry-first: a lapsed lease
// is returned to queued (committed — the homestate committedRefusal shape)
// and the renew refuses with ErrLeaseExpired WITHOUT extending
// (AC-TCL-006 arm 3). A foreign holder is refused with ErrLeaseHolder and
// nothing is written. Renewal does not move picked_at, picked_by, state,
// or spec_id.
func (s *BacklogStore) RenewLease(id, holder string) (*BacklogClaim, error) {
	now := time.Now().UTC()
	var result *BacklogClaim
	var committedRefusal error
	err := s.Mutate(func(rec *BacklogRecord) error {
		reclaimed := reclaimExpiredBacklogLeases(rec, now)
		for i := range rec.Items {
			if rec.Items[i].ID != id {
				continue
			}
			it := &rec.Items[i]
			for _, r := range reclaimed {
				if r.ItemID == id {
					// The addressed card's own lease had lapsed: the return
					// commits (callback returns nil) and the caller still
					// sees the refusal — renew-after-expiry never extends.
					// The result carries the reclamation so the surface can
					// render its C5 audit line even on this refusal.
					committedRefusal = fmt.Errorf("%w: %s lapsed at %s and was returned to queued",
						ErrLeaseExpired, id, derefOr(it.LeaseExpiresAt, "unknown"))
					result = &BacklogClaim{Item: *it, Reclaimed: reclaimed}
					return nil
				}
			}
			if it.State != BacklogStatePicked {
				return fmt.Errorf("backlog item %s is %s, not picked — nothing to renew", id, it.State)
			}
			if it.PickedBy == nil || strings.TrimSpace(holder) != *it.PickedBy {
				return fmt.Errorf("%w: %q does not hold the lease on %s (holder %s)",
					ErrLeaseHolder, holder, id, derefOr(it.PickedBy, "none"))
			}
			extended := now.Add(config.DefaultFactoryLeaseDuration).Format(time.RFC3339)
			it.LeaseExpiresAt = &extended
			result = &BacklogClaim{Item: *it, Reclaimed: reclaimed}
			return nil
		}
		return fmt.Errorf("no backlog item %s", id)
	})
	if err != nil {
		return nil, err
	}
	if committedRefusal != nil {
		// A non-nil result alongside the error is the committed expiry-first
		// return: the caller renders the reclamation line, then the refusal.
		return result, committedRefusal
	}
	return result, nil
}

// @MX:ANCHOR: [AUTO] ReclaimExpired — the queue-wide expiry-first return
// @MX:REASON: expected fan_in >= 3 (claim/renew run the same pass internally; surfaces call it for the standalone reclaim audit sweep); an out-of-Mutate variant would race the whole-record write (REQ-TCL-004)
//
// @MX:WARN: [TID:TX] claim-family engine access lives ONLY inside the Mutate callback
// @MX:REASON: [TID:TX] the pass mutates only lapsed cards, but it must still commit through the single whole-record write the flock serializes
//
// ReclaimExpired returns every lapsed lease to queued and surfaces each as
// a reclamation (id + previous holder, C5). Live leases are never crossed
// (REQ-TCL-007): only a card whose own expiry has passed — C4: unparseable
// counts as passed — is mutated.
func (s *BacklogStore) ReclaimExpired() ([]BacklogReclamation, error) {
	now := time.Now().UTC()
	var out []BacklogReclamation
	err := s.Mutate(func(rec *BacklogRecord) error {
		out = reclaimExpiredBacklogLeases(rec, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// derefOr renders a nullable stamp for a refusal message; absence reads as
// the named fallback rather than an empty cell.
func derefOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// acquireBacklogLockSerialized acquires the backlog's sibling lock, retrying
// contention against the SAME shared wait policy as the board lock
// (stateLockWaitBudget and stateLockRetryWait, state_lock_wait.go — REQ-BLB-006:
// one policy, both call sites, so a change to either the budget or the
// backoff applies here without a second edit): a mutation racing a
// short-lived holder serializes behind it instead of failing, while a
// genuinely stuck holder surfaces as an error rather than a hang. The timeout
// error names the lock artifact so the operator can act on the right file.
func (s *BacklogStore) acquireLock() (*StateLock, error) {
	// The ancestor check runs before the directory is created: MkdirAll follows
	// a symlinked `.moai` and would create the queue directory outside the
	// project before the opener refused (card t1458).
	if err := ensureStateLockDir(s.LockPath()); err != nil {
		if errors.Is(err, ErrStateLockUnsafePath) {
			return nil, fmt.Errorf("mutate backlog %s: lock %s: %w", s.EnginePath(), s.LockPath(), err)
		}
		return nil, fmt.Errorf("mutate backlog %s: creating dir: %w", s.EnginePath(), err)
	}
	var lastErr error
	deadline := time.Now().Add(stateLockWaitBudget)
	for attempt := 0; ; attempt++ {
		impl, err := acquireStateLockImpl(s.LockPath())
		if err == nil {
			return &StateLock{path: s.LockPath(), impl: impl}, nil
		}
		if !IsStateLockHeld(err) {
			return nil, fmt.Errorf("mutate backlog %s: lock %s: %w", s.EnginePath(), s.LockPath(), err)
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			return nil, fmt.Errorf("mutate backlog %s: lock %s: %w", s.EnginePath(), s.LockPath(), lastErr)
		}
		time.Sleep(stateLockRetryWait(attempt))
	}
}

// normalizeBacklogRecord establishes the invariants every in-memory record
// holds: version 1, a non-nil item slice, and a high-water mark that clears
// every present id (max of persisted and max-present — REQ-TODO-009's
// derive-on-absent and the hand-edited-low-value guard in one rule).
func normalizeBacklogRecord(rec *BacklogRecord) {
	if rec.Version == 0 {
		rec.Version = backlogVersion
	}
	if rec.Items == nil {
		rec.Items = []BacklogItem{}
	}
	if rec.Findings == nil {
		rec.Findings = []BacklogFinding{}
	}
	if rec.Archived == nil {
		rec.Archived = []BacklogArchiveEntry{}
	}
	for i := range rec.Archived {
		if rec.Archived[i].Findings == nil {
			rec.Archived[i].Findings = []BacklogArchivedFinding{}
		}
	}
	// The mark must clear every id the record HOLDS, archived rows included:
	// a hand-edited low last_seq beside an archived t7 would otherwise reissue
	// t7 to a new card and make that card's restore unreachable (REQ-TDG-013).
	if max := maxPresentBacklogSeq(rec.Items); max > rec.LastSeq {
		rec.LastSeq = max
	}
	for _, entry := range rec.Archived {
		if n, ok := parseBacklogSeq(entry.Item.ID); ok && n > rec.LastSeq {
			rec.LastSeq = n
		}
	}
}

// maxPresentBacklogSeq returns the highest numeric suffix among t<N> ids.
func maxPresentBacklogSeq(items []BacklogItem) int {
	max := 0
	for _, it := range items {
		if n, ok := parseBacklogSeq(it.ID); ok && n > max {
			max = n
		}
	}
	return max
}

// parseBacklogSeq extracts N from an id of the form t<N> (N > 0).
func parseBacklogSeq(id string) (int, bool) {
	if !strings.HasPrefix(id, "t") {
		return 0, false
	}
	n, err := strconv.Atoi(id[1:])
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// ParseBacklogSeq is the exported form of parseBacklogSeq, for read
// surfaces outside this package that must interpret a card id against the
// id-space accounting (SPEC-TODO-ARCHIVE-QUERY-001's history verb
// qualifies an absent answer against the issued-id mark). A second parser
// would be a second chance to drift from the id form the store issues.
func ParseBacklogSeq(id string) (int, bool) {
	return parseBacklogSeq(id)
}
