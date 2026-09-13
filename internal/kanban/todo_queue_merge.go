// todo_queue_merge.go — the pure two-store merge core
// (SPEC-TODO-QUEUE-HOME-MERGE-001 M1, plan.md §F).
//
// The card t657 backlog queue diverged into two stores (spec.md §1): the
// project-local legacy store and the canonical home store, each minting ids
// the other cannot see. This file is the merge's DECISION layer: a pure
// function over two BacklogRecord values that classifies every project card
// into the SPEC's collision taxonomy and produces the merged record plus the
// audit rows (mapping, duplicates, reconciliation) the evidence artifacts are
// rendered from.
//
// The function is PURE by contract: neither input is mutated, and no store,
// lock, or file is touched. The one-off merge shell (cmd/t657-merge) owns the
// IO half — LoadPure ×2, this function, ONE Mutate on the home store.
//
// @MX:ANCHOR: [AUTO] MergeBacklogRecords — the single merge decision point
// @MX:REASON: the one-off merge shell, its rehearsal tests, and the M4 window procedure all consume this function; a second merge implementation would let rehearsal and execution disagree about the taxonomy
package kanban

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// MergeReconcileIdentityCollision is the reconciliation row Kind recorded when
// a project card's identity UUID is already held by the home store
// (acceptance.md §D.5): the merge issues a fresh UUIDv7 and records the
// re-issue — never an error, never a silent reuse.
const MergeReconcileIdentityCollision = "identity-uuid-collision"

// MergeMappingRow is one old→new renumbering decision (REQ-TQM-005/008).
type MergeMappingRow struct {
	OldID string
	NewID string
}

// MergeDuplicateRow records one number-shared, content-identical pair the
// merge resolved by keeping the home copy (REQ-TQM-006).
type MergeDuplicateRow struct {
	ID string
}

// MergeReconciliationRow is one audit row about a decision that is not a
// plain migration (REQ-TQM-011's reconciliation evidence, identity branch).
type MergeReconciliationRow struct {
	CardID string // the card's PRE-merge (project) id
	Kind   string
	Detail string
}

// MergeReport is the decision log one merge produced. Mapping is the
// id-mapping.tsv content; Reconciliation feeds reconciliation.tsv.
type MergeReport struct {
	HighWater      int
	Migrated       []string
	Renumbered     []MergeMappingRow
	Duplicates     []MergeDuplicateRow
	Reconciliation []MergeReconciliationRow
}

// Mapping returns the renumbering rows in decision order — the exact content
// of the mapping-table artifact.
func (r *MergeReport) Mapping() []MergeMappingRow { return r.Renumbered }

// MergeOptions configures the merge. NewUUID issues a fresh identity UUID for
// collision resolution; nil means uuid.NewV7. Tests inject a deterministic
// issuer.
type MergeOptions struct {
	NewUUID func() (string, error)
}

// mergeCardTokenPattern matches a whole card token inside free text. The boundary
// anchors are the t642/t6420 hazard: a rewrite keyed on the bare token would
// corrupt every longer id sharing the prefix (plan.md §B).
var mergeCardTokenPattern = regexp.MustCompile(`\bt[0-9]+\b`)

// rewriteCardTokens rewrites whole card tokens in s through mapping. Tokens
// with no mapping pass through unchanged.
func rewriteCardTokens(s string, mapping map[string]string) string {
	if len(mapping) == 0 {
		return s
	}
	return mergeCardTokenPattern.ReplaceAllStringFunc(s, func(tok string) string {
		if next, ok := mapping[tok]; ok {
			return next
		}
		return tok
	})
}

// mergeCardContentEqual reports whether a number-shared pair is the SAME card
// observed twice (REQ-TQM-006's content-identical case): same text, state,
// and spec pointer. AddedAt and landing evidence are metadata, deliberately
// outside the comparison — the home copy wins either way.
func mergeCardContentEqual(a, b BacklogItem) bool {
	if a.Text != b.Text || a.State != b.State {
		return false
	}
	if (a.SpecID == nil) != (b.SpecID == nil) {
		return false
	}
	return a.SpecID == nil || *a.SpecID == *b.SpecID
}

// mergedHighWater is the id ceiling the merge issues renumbers from:
// max(last_seq_home, last_seq_project, max id home, max id project), archived
// rows included (acceptance.md §D.5 — the merge must not itself mint a
// colliding id).
func mergedHighWater(home, project *BacklogRecord) int {
	hw := home.LastSeq
	if project.LastSeq > hw {
		hw = project.LastSeq
	}
	for _, it := range home.Items {
		if n, ok := parseBacklogSeq(it.ID); ok && n > hw {
			hw = n
		}
	}
	for _, e := range home.Archived {
		if n, ok := parseBacklogSeq(e.Item.ID); ok && n > hw {
			hw = n
		}
	}
	for _, it := range project.Items {
		if n, ok := parseBacklogSeq(it.ID); ok && n > hw {
			hw = n
		}
	}
	for _, e := range project.Archived {
		if n, ok := parseBacklogSeq(e.Item.ID); ok && n > hw {
			hw = n
		}
	}
	return hw
}

// MergeBacklogRecords merges the project (legacy) record INTO the home
// (canonical) record and returns the merged result plus the decision log.
//
// Taxonomy per shared number (spec.md §3.2): content-identical → the home
// copy stands, one duplicate row; different content → the project card is
// renumbered above the merged high-water and one mapping row records
// old→new; project-only ids migrate unchanged. References to renumbered ids
// are rewritten in structured finding fields, runtime assignment card ids,
// and free-text finding notes (token-boundary aware). An identity UUID the
// home store already holds is replaced by a fresh UUIDv7 plus a
// reconciliation row.
//
// Neither input is mutated. An error aborts the merge with nothing decided.
func MergeBacklogRecords(home, project *BacklogRecord, opts MergeOptions) (*BacklogRecord, *MergeReport, error) {
	if home == nil || project == nil {
		return nil, nil, fmt.Errorf("merge backlog records: both stores are required")
	}

	report := &MergeReport{HighWater: mergedHighWater(home, project)}
	mapping := map[string]string{}
	for _, row := range report.Renumbered {
		mapping[row.OldID] = row.NewID
	}

	// Deep-copy home so neither input is ever written through.
	merged := cloneBacklogRecord(home)
	merged.ProjectUUID = cloneIdentity(home.ProjectUUID)

	// Home occupancy and identity claims.
	homeLive := map[string]*BacklogItem{}
	for i := range home.Items {
		homeLive[home.Items[i].ID] = &home.Items[i]
	}
	homeArchived := map[string]*BacklogItem{}
	for i := range home.Archived {
		homeArchived[home.Archived[i].Item.ID] = &home.Archived[i].Item
	}
	claimedUUID := map[string]bool{}
	if home.ProjectUUID != nil {
		claimedUUID[*home.ProjectUUID] = true
	}
	for _, it := range home.Items {
		if it.CardUUID != nil {
			claimedUUID[*it.CardUUID] = true
		}
	}
	for _, e := range home.Archived {
		if e.Item.CardUUID != nil {
			claimedUUID[*e.Item.CardUUID] = true
		}
	}

	newUUID := opts.NewUUID
	if newUUID == nil {
		newUUID = func() (string, error) {
			issued, err := uuid.NewV7()
			if err != nil {
				return "", fmt.Errorf("merge backlog records: issue UUIDv7: %w", err)
			}
			return issued.String(), nil
		}
	}

	// resolveIdentity returns the identity a migrated card carries: its own
	// when the home store does not hold it, a freshly issued one plus a
	// reconciliation row when it does (never an error, never a silent reuse).
	resolveIdentity := func(cardID string, preferred *string) (*string, error) {
		if preferred == nil || !claimedUUID[*preferred] {
			if preferred != nil {
				claimedUUID[*preferred] = true
			}
			return cloneIdentity(preferred), nil
		}
		var fresh string
		for attempt := 0; attempt < 8; attempt++ {
			issued, err := newUUID()
			if err != nil {
				return nil, fmt.Errorf("merge backlog records: card %s: %w", cardID, err)
			}
			if !claimedUUID[issued] {
				fresh = issued
				break
			}
		}
		if fresh == "" {
			return nil, fmt.Errorf("merge backlog records: card %s: uuid issuer produced only contested values", cardID)
		}
		claimedUUID[fresh] = true
		report.Reconciliation = append(report.Reconciliation, MergeReconciliationRow{
			CardID: cardID,
			Kind:   MergeReconcileIdentityCollision,
			Detail: fmt.Sprintf("preferred identity %s already held by the home store; fresh identity %s issued", *preferred, fresh),
		})
		return &fresh, nil
	}

	// classify decides one project card's fate against the home id-space.
	// Returns the id the card carries into the merged record and whether it
	// migrates at all.
	classify := func(card BacklogItem) (target *BacklogItem, migrate bool, err error) {
		homeItem, live := homeLive[card.ID]
		if !live {
			homeItem, live = homeArchived[card.ID]
		}
		if !live {
			migrated, idErr := resolveIdentity(card.ID, card.CardUUID)
			if idErr != nil {
				return nil, false, idErr
			}
			card.CardUUID = migrated
			report.Migrated = append(report.Migrated, card.ID)
			return &card, true, nil
		}
		if mergeCardContentEqual(card, *homeItem) {
			report.Duplicates = append(report.Duplicates, MergeDuplicateRow{ID: card.ID})
			return nil, false, nil
		}
		// Different content: renumber from the high-water (REQ-TQM-005).
		report.HighWater++
		newID := fmt.Sprintf("t%d", report.HighWater)
		mapping[card.ID] = newID
		report.Renumbered = append(report.Renumbered, MergeMappingRow{OldID: card.ID, NewID: newID})
		identity, idErr := resolveIdentity(card.ID, card.CardUUID)
		if idErr != nil {
			return nil, false, idErr
		}
		card.ID = newID
		card.CardUUID = identity
		return &card, true, nil
	}

	for _, card := range project.Items {
		target, migrate, err := classify(card)
		if err != nil {
			return nil, nil, err
		}
		if migrate {
			merged.Items = append(merged.Items, *target)
		}
	}
	for _, entry := range project.Archived {
		target, migrate, err := classify(entry.Item)
		if err != nil {
			return nil, nil, err
		}
		if !migrate {
			continue
		}
		merged.Archived = append(merged.Archived, BacklogArchiveEntry{
			Item:     *target,
			Position: entry.Position,
			Findings: rewriteArchivedFindings(entry.Findings, mapping),
		})
	}

	// Project findings join the merged record with references rewritten and
	// exact-tuple duplicates of home findings dropped (the same
	// re-run-must-not-stack rule HasFindingTuple enforces on the analyser).
	for _, f := range project.Findings {
		f.SubjectID = rewriteCardTokens(f.SubjectID, mapping)
		f.RelatedID = rewriteCardTokens(f.RelatedID, mapping)
		f.Note = rewriteCardTokens(f.Note, mapping)
		if merged.HasFindingTuple(f) {
			continue
		}
		merged.Findings = append(merged.Findings, f)
	}

	// Runtime: the merged record is a report, so project runs join when their
	// id is new and assignments join under rewritten card ids, home winning
	// any (run, card) collision — the store's runtime tables are keyed
	// exactly this way.
	knownRuns := map[string]bool{}
	for _, run := range merged.Runtime.Runs {
		knownRuns[run.RunID] = true
	}
	for _, run := range project.Runtime.Runs {
		if knownRuns[run.RunID] {
			continue
		}
		knownRuns[run.RunID] = true
		merged.Runtime.Runs = append(merged.Runtime.Runs, run)
	}
	knownAssign := map[string]bool{}
	for _, a := range merged.Runtime.Assignments {
		knownAssign[a.RunID+"\x00"+a.CardID] = true
	}
	for _, a := range project.Runtime.Assignments {
		a.CardID = rewriteCardTokens(a.CardID, mapping)
		key := a.RunID + "\x00" + a.CardID
		if knownAssign[key] {
			continue
		}
		knownAssign[key] = true
		merged.Runtime.Assignments = append(merged.Runtime.Assignments, a)
	}

	merged.LastSeq = report.HighWater
	return merged, report, nil
}

// rewriteArchivedFindings rewrites the finding references riding inside a
// migrated archive entry.
func rewriteArchivedFindings(findings []BacklogArchivedFinding, mapping map[string]string) []BacklogArchivedFinding {
	if len(findings) == 0 {
		return findings
	}
	out := make([]BacklogArchivedFinding, 0, len(findings))
	for _, af := range findings {
		af.Finding.SubjectID = rewriteCardTokens(af.Finding.SubjectID, mapping)
		af.Finding.RelatedID = rewriteCardTokens(af.Finding.RelatedID, mapping)
		af.Finding.Note = rewriteCardTokens(af.Finding.Note, mapping)
		out = append(out, af)
	}
	return out
}

// cloneBacklogRecord deep-copies a record's slices and identity pointers so
// the merge never writes through an input.
func cloneBacklogRecord(rec *BacklogRecord) *BacklogRecord {
	out := &BacklogRecord{
		ProjectUUID: cloneIdentity(rec.ProjectUUID),
		Version:     rec.Version,
		LastSeq:     rec.LastSeq,
		Items:       append([]BacklogItem(nil), rec.Items...),
		Findings:    append([]BacklogFinding(nil), rec.Findings...),
		Archived:    append([]BacklogArchiveEntry(nil), rec.Archived...),
		Runtime: TodoRuntime{
			Runs:        append([]TodoRuntimeRun(nil), rec.Runtime.Runs...),
			Assignments: append([]TodoRuntimeAssignment(nil), rec.Runtime.Assignments...),
		},
	}
	for i := range out.Items {
		out.Items[i].SpecID = cloneIdentity(out.Items[i].SpecID)
		out.Items[i].CardUUID = cloneIdentity(out.Items[i].CardUUID)
		out.Items[i].Landing = rec.Items[i].Landing
	}
	for i := range out.Archived {
		out.Archived[i].Item.SpecID = cloneIdentity(out.Archived[i].Item.SpecID)
		out.Archived[i].Item.CardUUID = cloneIdentity(out.Archived[i].Item.CardUUID)
		out.Archived[i].Findings = append([]BacklogArchivedFinding(nil), out.Archived[i].Findings...)
	}
	return out
}
