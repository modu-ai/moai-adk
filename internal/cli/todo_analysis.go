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
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// appendAnalyzedCard classifies text against the queue and appends it,
// refusing an exact duplicate unless force is set. It runs INSIDE a Mutate
// callback: returning an error there aborts the whole write, which is what
// leaves the queue file byte-identical on a refusal — the same contract
// `edit`, `drop`, and `next --expect` already stand on, rather than a new
// invariant invented for this path.
//
// The returned position counts queued cards, matching the store's own Add.
func appendAnalyzedCard(rec *factory.BacklogRecord, text string, state factory.BacklogState, force bool) (factory.BacklogItem, int, error) {
	match := factory.ClassifyCardText(text, rec.Items)
	if match.Kind == factory.BacklogMatchExact && !force {
		return factory.BacklogItem{}, 0, fmt.Errorf(
			"todo add: %s already holds this card (%q) — pass --force to add it anyway",
			match.ID, todoTextPrefix(todoCardText(rec, match.ID)))
	}

	rec.LastSeq++
	item := factory.BacklogItem{
		ID:      fmt.Sprintf("t%d", rec.LastSeq),
		Text:    text,
		AddedAt: time.Now().UTC().Format(time.RFC3339),
		State:   state,
	}
	// REQ-TST-004: entering the picked state stamps picked_at — add --pick
	// is one of the two pick transitions, so its one locked write carries
	// the stamp the live pick would.
	if state == factory.BacklogStatePicked {
		item.PickedAt = todoStampNow()
	}
	rec.Items = append(rec.Items, item)

	// The finding names the NEW card as the subject: it is the card whose
	// admission the finding explains. A finding is a record and nothing
	// more — neither branch below touches a field of the card it names.
	switch match.Kind {
	case factory.BacklogMatchExact:
		rec.AppendFindingOnce(factory.BacklogFinding{
			SubjectID: item.ID,
			RelatedID: match.ID,
			Relation:  factory.BacklogRelationDuplicateForced,
			Source:    factory.BacklogSourceMechanical,
			Score:     match.Score,
			At:        item.AddedAt,
		})
	case factory.BacklogMatchNear:
		rec.AppendFindingOnce(factory.BacklogFinding{
			SubjectID: item.ID,
			RelatedID: match.ID,
			Relation:  factory.BacklogRelationNearDuplicate,
			Source:    factory.BacklogSourceMechanical,
			Score:     match.Score,
			At:        item.AddedAt,
		})
	case factory.BacklogMatchNone:
	}

	// Consumer C (SPEC-JEV-CONSUMERS-001, REQ-JEVN-001): the model's
	// near-duplicate judgment is recorded HERE and only here. The `analyze`
	// re-sweep below is a distinct entry point and deliberately does not call
	// it — the re-sweep walks pairs that already carry findings, so a consumer
	// there changes the finding population on a path the SPEC never measured.
	//
	// It runs AFTER the mechanical branches above so that a pair the analyser
	// just recorded suppresses the arriving Jev finding (REQ-JEVN-006 half
	// (a)) rather than the other way round. The return value is deliberately
	// discarded: nothing may act on a model signal.
	appendJevNearDuplicateFinding(rec, item)

	pos := 0
	for _, it := range rec.Items {
		if it.State == factory.BacklogStateQueued {
			pos++
		}
	}
	return item, pos, nil
}

// todoCardText returns the text of the card with id, or the id itself when
// no such card is present — a message is never worth a panic.
func todoCardText(rec *factory.BacklogRecord, id string) string {
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
			err := newTodoStore().Mutate(func(rec *factory.BacklogRecord) error {
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
func analyzeQueue(rec *factory.BacklogRecord) (pairs, recorded int) {
	now := time.Now().UTC().Format(time.RFC3339)
	// Prepare once per card while preserving the exact historical comparator:
	// NFC/case/whitespace normalization and sets of normalized words.
	normalized := make([]string, len(rec.Items))
	tokens := make([]map[string]struct{}, len(rec.Items))
	for i, item := range rec.Items {
		if item.State == factory.BacklogStateDropped {
			continue
		}
		normalized[i] = factory.NormalizeCardText(item.Text)
		tokens[i] = make(map[string]struct{})
		for _, token := range strings.Fields(normalized[i]) {
			tokens[i][token] = struct{}{}
		}
	}
	for j, subject := range rec.Items {
		if subject.State == factory.BacklogStateDropped {
			continue
		}
		for i, related := range rec.Items[:j] {
			if related.State == factory.BacklogStateDropped {
				continue
			}
			pairs++
			relation := ""
			score := todoTokenSetScore(tokens[j], tokens[i])
			switch {
			case normalized[j] != "" && normalized[j] == normalized[i]:
				relation, score = factory.BacklogRelationDuplicateForced, 1
			case score >= factory.BacklogNearDuplicateThreshold && score < 1.0:
				relation = factory.BacklogRelationNearDuplicate
			default:
				continue
			}
			if rec.AppendFindingOnce(factory.BacklogFinding{
				SubjectID: subject.ID,
				RelatedID: related.ID,
				Relation:  relation,
				Source:    factory.BacklogSourceMechanical,
				Score:     score,
				At:        now,
			}) {
				recorded++
			}
		}
	}
	return pairs, recorded
}

func todoTokenSetScore(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for token := range a {
		if _, ok := b[token]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(a)+len(b)-intersection)
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
func todoFindingLine(rec *factory.BacklogRecord, cardID string, f factory.BacklogFinding) string {
	counterpart := f.RelatedID
	if f.SubjectID != cardID {
		counterpart = f.SubjectID
	}
	mark := ""
	if f.Source == factory.BacklogSourceMechanical && !rec.HasAgentFindingForPair(f) {
		mark = ", machine-only"
	}
	// The score is a measurement, so it is printed only where one was
	// taken. An agent judgement carries no score, and rendering it as
	// "0.00" would read as a measured dissimilarity rather than as the
	// absence of a measurement.
	//
	// A Jev finding is a THIRD thing (REQ-JEVN-005): a calibrated model
	// confidence, which is neither a measured similarity nor the absence of a
	// measurement. Both existing branches are wrong for it — inheriting the
	// mechanical one would print the probability as `score N.NN` and read as a
	// measurement, inheriting the agent one would drop the probability
	// silently — so it renders with its own label. The label is the
	// package constant rather than a literal, so the marker a reader learns to
	// recognise here cannot drift from the one internal/jev emits elsewhere.
	// The Jev fragment is built in todo_jev_finding.go so that the
	// internal/jev dependency stays confined to the one file that owns the
	// consumer.
	score := ""
	switch f.Source {
	case factory.BacklogSourceMechanical:
		score = fmt.Sprintf(", score %.2f", f.Score)
	case factory.BacklogSourceJev:
		score = jevFindingSignalFragment(f)
	}
	note := ""
	if f.Note != "" {
		note = fmt.Sprintf(" — %s", todoPRCell(f.Note))
	}
	// A near-duplicate finding explains a card's ADMISSION, so its drop/edit
	// suggestion names the finding's SUBJECT — the newer card — on both rows;
	// filling it with the row it sits under told the operator, beneath the
	// original card, to drop the original (card t1470, GitHub #1732). Every
	// other relation is a judgement about the row's own card, so its
	// suggestion keeps naming that row (card t1484).
	target := cardID
	if f.Relation == kanban.BacklogRelationNearDuplicate {
		target = f.SubjectID
	}
	return fmt.Sprintf("\t↳ %s %s (%s%s%s)%s — moai todo drop %s | moai todo edit %s \"<text>\"",
		f.Relation, counterpart, f.Source, score, mark, note, target, target)
}
