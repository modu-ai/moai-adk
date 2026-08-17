// reconcile.go — the (column, status) compatibility table and the
// reconciled card view (SPEC-KANBAN-BOARD-001 REQ-KB-008, M2).
//
// Because the board and the frontmatter answer different questions over the
// same card, some disagreement is normal: backlog/plan both correspond to
// draft, run/review both to in-progress. A pair INSIDE the table is
// consistent; a pair OUTSIDE it is an inconsistency the board resolves in
// the safe direction — the card is marked inconsistent, reported not
// dispatchable, and both values are surfaced. It repairs NEITHER: rewriting
// the status would write a lifecycle transition the board does not own
// (REQ-KB-007), and rewriting the column would overwrite the record of an
// actor that observed something the board did not. A board that quietly
// reconciles a disagreement destroys the only evidence that something
// upstream is wrong (AP-4).
//
// Out-of-lifecycle terminals (archived / superseded / rejected) appear in no
// row: no card is created for them at all — decided BEFORE reconciliation,
// by ShouldCreateCard — so the table is never consulted for them, and card
// absence stays distinguishable from an inconsistency report (AC-KB-021).
package kanban

import "fmt"

// CardView is one card's reconciled state — the board's read-side answer
// for a consumer deciding whether to dispatch.
type CardView struct {
	Card Card
	// Status is the source-resolved status: a canonical enum value, ""
	// (no spec.md), or StatusUnresolved.
	Status          string
	StatusSource    string
	SpecFilePresent bool

	// Dispatchable is the dispatch verdict.
	Dispatchable bool

	// Inconsistent is true when the (column, status) pairing falls outside
	// the compatibility table (or the recorded column is outside the closed
	// set). Never repaired; both values surfaced in Details.
	Inconsistent bool

	// ColumnAmbiguous is true when the pairing is legal AND the status does
	// not by itself decide between the two columns it legally pairs with
	// (draft/planned → backlog vs plan; in-progress → run vs review). The
	// recorded column stands; the board resolves nothing (AC-KB-010).
	ColumnAmbiguous bool

	// Unresolved is true when the status SOURCE resolved to no single tree
	// (REQ-KB-024). A distinct outcome from Inconsistent — both refuse
	// dispatch, but an unresolved source is an operator-visible resolution
	// failure, not a pairing violation, and no enum member is substituted.
	Unresolved bool

	// Candidates surfaces every candidate the resolution found, when
	// unresolved.
	Candidates []string

	// WorktreePath names the observed live worktree, when one exists.
	WorktreePath string

	// Details surfaces the values behind a refusal, for the operator.
	Details string
}

// ShouldCreateCard decides whether a SPEC becomes a board card at all
// (AC-KB-021): out-of-lifecycle terminals are not work in flight — done
// means this card was worked and finished, and a rejected SPEC was never
// worked — so no card is created for them, in any column. This is card
// ABSENCE, not an inconsistency: REQ-KB-008 is never reached.
func ShouldCreateCard(cs *CardStatus) bool {
	if cs == nil {
		return false
	}
	switch cs.Status {
	case StatusArchived, StatusSuperseded, StatusRejected:
		return false
	default:
		return true
	}
}

// @MX:ANCHOR: [AUTO] ReconcileCard — the read-side dispatch verdict every board consumer decides on
// @MX:REASON: expected fan_in >= 3 (the dispatch path, the review/sync column transitions, any board rendering); an inconsistent card slipping through here dispatches work the SPEC refused
//
// ReconcileCard applies the compatibility table of spec.md §A.4 (as revised
// at v0.2.0: `planned` admitted in backlog and plan and nowhere else;
// `completed` admitted in sync; terminals producing no card upstream) to one
// recorded card and its source-resolved status.
func ReconcileCard(card Card, cs *CardStatus) CardView {
	view := CardView{Card: card}
	if cs != nil {
		view.Status = cs.Status
		view.StatusSource = cs.Source
		view.SpecFilePresent = cs.SpecFilePresent
		view.Candidates = cs.Candidates
		view.WorktreePath = cs.WorktreePath
	}

	// REQ-KB-024 first: an unresolved source is not a pairing question — no
	// status was read, so the table is not reached, no enum member is
	// substituted, and the card is NOT reported inconsistent. It keeps its
	// recorded column, is not dispatchable, and every candidate surfaces.
	//
	// The verdict keys on the SOURCE discriminator, never on the status
	// STRING: a frontmatter literally carrying `status: unresolved` is not a
	// member of the canonical enum, but parseFrontmatterStatus returns it
	// verbatim — keying on cs.Status would misfile such a card as a
	// resolution failure when its source resolved cleanly, and would hide
	// the pairing violation behind an unrelated detail line. Keying on
	// cs.Source routes the literal to the table, where it is the
	// outside-every-row inconsistency it actually is.
	if cs != nil && cs.Source == StatusSourceUnresolved {
		view.Unresolved = true
		view.Details = fmt.Sprintf("status source unresolved (observed worktree %q reported no branch); recorded column %q stands; candidates: %v",
			cs.WorktreePath, card.Column, cs.Candidates)
		return view
	}

	// A recorded column outside the closed set is an inconsistency the
	// reconciliation surfaces — never a silent reinterpretation.
	if !card.Column.Valid() {
		view.Inconsistent = true
		view.Details = fmt.Sprintf("recorded column %q is outside the closed six-column set", card.Column)
		return view
	}

	consistent := pairingConsistent(card.Column, view.Status, view.SpecFilePresent)
	if !consistent {
		view.Inconsistent = true
		view.Details = fmt.Sprintf("pairing (column %q, status %q) is outside the compatibility table; neither value repaired",
			card.Column, describeStatus(view.Status, view.SpecFilePresent))
		return view
	}

	view.ColumnAmbiguous = statusDecidesMultipleColumns(view.Status)
	view.Dispatchable = card.Column.HasOwningSession()
	if !view.Dispatchable {
		view.Details = fmt.Sprintf("column %q has no owning session", card.Column)
	}
	return view
}

// pairingConsistent decides the compatibility table of spec.md §A.4:
//
//	backlog | plan : no spec.md, draft, planned
//	run     : in-progress
//	sync    : in-progress, implemented, completed
//	done    : completed
//
// Terminals appear in no row (no card exists for them by the time the table
// runs); a pairing naming them is therefore outside every row.
func pairingConsistent(col Column, status string, specFilePresent bool) bool {
	if !specFilePresent {
		// A card admitted before planning has no frontmatter at all.
		return col == ColumnBacklog || col == ColumnPlan
	}
	switch status {
	case StatusDraft, StatusPlanned:
		return col == ColumnBacklog || col == ColumnPlan
	case StatusInProgress:
		return col == ColumnRun || col == ColumnSync
	case StatusImplemented:
		return col == ColumnSync
	case StatusCompleted:
		return col == ColumnSync || col == ColumnDone
	default:
		// Terminals and anything unrecognized: outside every row.
		return false
	}
}

// statusDecidesMultipleColumns reports the collisions AC-KB-010 names:
// draft does not by itself decide between backlog and plan, in-progress does
// not decide between run and sync. The board reports the ambiguity of its
// status input; it never resolves a column from it (REQ-KB-006).
func statusDecidesMultipleColumns(status string) bool {
	switch status {
	case StatusDraft, StatusPlanned, StatusInProgress:
		return true
	default:
		return false
	}
}

// describeStatus renders a status for an operator-facing detail line.
func describeStatus(status string, specFilePresent bool) string {
	if !specFilePresent {
		return "no spec.md"
	}
	return status
}
