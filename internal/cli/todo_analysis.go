// todo_analysis.go — SPEC-TODO-ANALYSIS-001 M3/M4: the analyser's CLI seam.
//
// Three things live here: the add-path append that runs the analysis inside
// the SAME locked write that would append the card, the `analyze` verb that
// re-reads the whole queue and records without appending, and the finding
// line the operator sees under a card in `todo list`.
//
// Why the analysis runs inside the lock rather than before it: an analysis
// performed outside Mutate is a read whose answer can be stale by the time
// the append happens. Two sessions adding the same card would both measure
// against a queue lacking it and both be admitted — the read-modify-write
// race SPEC-KANBAN-TODO-CLI-001 closed for the queue file, reopened one
// layer up.
//
// SUBAGENT BOUNDARY (REQ-TA-015): nothing here prompts. A refusal is an
// error on stderr and a non-zero exit; the caller decides what to do about
// it, and `--force` is how they say "add it anyway" in advance.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// appendAnalyzedCard classifies text against the queue and appends it,
// refusing an exact duplicate unless force is set. It runs INSIDE a Mutate
// callback: returning an error there aborts the whole write, which is what
// leaves the queue file byte-identical on a refusal — the same contract
// `edit`, `drop`, and `next --expect` already stand on, rather than a new
// invariant invented for this path.
//
// The returned position counts queued cards, matching the store's own Add.
func appendAnalyzedCard(rec *kanban.BacklogRecord, text string, state kanban.BacklogState, force bool) (kanban.BacklogItem, int, error) {
	match := kanban.ClassifyCardText(text, rec.Items)
	if match.Kind == kanban.BacklogMatchExact && !force {
		return kanban.BacklogItem{}, 0, fmt.Errorf(
			"todo add: %s already holds this card (%q) — pass --force to add it anyway",
			match.ID, todoTextPrefix(todoCardText(rec, match.ID)))
	}

	rec.LastSeq++
	item := kanban.BacklogItem{
		ID:      fmt.Sprintf("t%d", rec.LastSeq),
		Text:    text,
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		State:   state,
	}
	rec.Items = append(rec.Items, item)

	// The finding names the NEW card as the subject: it is the card whose
	// admission the finding explains. A finding is a record and nothing
	// more — neither branch below touches a field of the card it names.
	switch match.Kind {
	case kanban.BacklogMatchExact:
		rec.AppendFindingOnce(kanban.BacklogFinding{
			SubjectID: item.ID,
			RelatedID: match.ID,
			Relation:  kanban.BacklogRelationDuplicateForced,
			Source:    kanban.BacklogSourceMechanical,
			Score:     match.Score,
			At:        item.AddedAt,
		})
	case kanban.BacklogMatchNear:
		rec.AppendFindingOnce(kanban.BacklogFinding{
			SubjectID: item.ID,
			RelatedID: match.ID,
			Relation:  kanban.BacklogRelationNearDuplicate,
			Source:    kanban.BacklogSourceMechanical,
			Score:     match.Score,
			At:        item.AddedAt,
		})
	case kanban.BacklogMatchNone:
	}

	pos := 0
	for _, it := range rec.Items {
		if it.State == kanban.BacklogStateQueued {
			pos++
		}
	}
	return item, pos, nil
}

// todoCardText returns the text of the card with id, or the id itself when
// no such card is present — a message is never worth a panic.
func todoCardText(rec *kanban.BacklogRecord, id string) string {
	for _, it := range rec.Items {
		if it.ID == id {
			return it.Text
		}
	}
	return id
}

// newTodoAnalyzeCmd — `moai todo analyze` (REQ-TA-002): re-read the whole
// queue and record what the analyser finds, appending, removing,
// reordering, and editing nothing.
//
// Re-running is idempotent by construction: every write goes through
// AppendFindingOnce, whose key is {subject, related, relation, source} with
// the timestamp deliberately outside it. Without that, a second run would
// stack a second copy of every measurement, the listing would fill with
// duplicates of one finding, and the operator would stop reading findings —
// which costs more than never having recorded them.
func newTodoAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "Re-analyse the whole queue and record findings (records only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var pairs, recorded int
			err := newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {
				pairs, recorded = analyzeQueue(rec)
				return nil
			})
			if err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "analyzed %d pairs, recorded %d findings\n",
				pairs, recorded)
			return nil
		},
	}
}

// analyzeQueue records a finding for every non-dropped pair the analyser
// classifies, returning how many pairs were compared and how many findings
// were newly recorded.
//
// The later card is the subject and the earlier one the related card, which
// is the same orientation the add path uses — so a finding `analyze` would
// write for a pair `add` already recorded carries an identical tuple and is
// deduplicated rather than doubled.
func analyzeQueue(rec *kanban.BacklogRecord) (pairs, recorded int) {
	now := time.Now().UTC().Format(time.RFC3339)
	for j, subject := range rec.Items {
		if subject.State == kanban.BacklogStateDropped {
			continue
		}
		for _, related := range rec.Items[:j] {
			if related.State == kanban.BacklogStateDropped {
				continue
			}
			pairs++
			relation := ""
			score := kanban.TokenSetJaccard(subject.Text, related.Text)
			switch {
			case kanban.NormalizeCardText(subject.Text) != "" &&
				kanban.NormalizeCardText(subject.Text) == kanban.NormalizeCardText(related.Text):
				relation, score = kanban.BacklogRelationDuplicateForced, 1
			case score >= kanban.BacklogNearDuplicateThreshold && score < 1.0:
				relation = kanban.BacklogRelationNearDuplicate
			default:
				continue
			}
			if rec.AppendFindingOnce(kanban.BacklogFinding{
				SubjectID: subject.ID,
				RelatedID: related.ID,
				Relation:  relation,
				Source:    kanban.BacklogSourceMechanical,
				Score:     score,
				At:        now,
			}) {
				recorded++
			}
		}
	}
	return pairs, recorded
}

// todoFindingLine renders one finding beneath the card it names in
// `todo list` (REQ-TA-011).
//
// The line carries the counterpart id relative to THIS card, so the same
// finding reads correctly under either end of the pair, and it names the
// literal commands the operator would run — a finding the operator cannot
// act on from the screen it appears on is a finding they scroll past.
//
// `machine-only` (REQ-TA-013) marks the absence of an agent-sourced finding
// for the same UNORDERED pair. It says nothing about whether anyone
// reviewed the pair: the CLI cannot know who called it, so the only
// honest claim available is "nothing agent-sourced was recorded here".
func todoFindingLine(rec *kanban.BacklogRecord, cardID string, f kanban.BacklogFinding) string {
	counterpart := f.RelatedID
	if f.SubjectID != cardID {
		counterpart = f.SubjectID
	}
	mark := ""
	if f.Source == kanban.BacklogSourceMechanical && !rec.HasAgentFindingForPair(f) {
		mark = ", machine-only"
	}
	// The score is a measurement, so it is printed only where one was
	// taken. An agent judgement carries no score, and rendering it as
	// "0.00" would read as a measured dissimilarity rather than as the
	// absence of a measurement.
	score := ""
	if f.Source == kanban.BacklogSourceMechanical {
		score = fmt.Sprintf(", score %.2f", f.Score)
	}
	note := ""
	if f.Note != "" {
		note = fmt.Sprintf(" — %s", f.Note)
	}
	return fmt.Sprintf("\t↳ %s %s (%s%s%s)%s — moai todo drop %s | moai todo edit %s \"<text>\"",
		f.Relation, counterpart, f.Source, score, mark, note, cardID, cardID)
}
