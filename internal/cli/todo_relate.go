// todo_relate.go — SPEC-TODO-ANALYSIS-001 M4: the agent-layer relation verbs.
//
// `relate` writes one finding. `unrelate` removes one finding. Neither
// writes a card field, and that is enforced by the SHAPE of the code rather
// than by a comment: the Mutate callbacks below read rec.Items only to
// confirm the two ids exist, and touch rec.Findings alone. Making `absorbs`
// actually absorb would take new code that does not exist here — so it
// cannot happen by accident, which is the property the doctrine's
// prohibition needs in order to be more than a promise.
//
// The four semantic relations (contains / absorbs / replaces / conflicts)
// are the judgements a text analyser cannot reach. Recording one causes
// nothing: the operator reads it and decides. That asymmetry is deliberate —
// a wrong mechanical refusal costs a card, a wrong record costs a line of
// output.
//
// SUBAGENT BOUNDARY (REQ-TA-015): nothing here prompts.
package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// newTodoRelateCmd — `moai todo relate <a> <b> --relation <r> [--note <text>]`
// (REQ-TA-008): record one agent-sourced finding between two existing cards.
func newTodoRelateCmd() *cobra.Command {
	var relation, note string
	cmd := &cobra.Command{
		Use:   "relate <a> <b> --relation <contains|absorbs|replaces|conflicts>",
		Short: "Record a relation between two cards (records only — changes no card)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			subject, related := normalizeTodoRef(args[0]), normalizeTodoRef(args[1])
			if err := runTodoRelate(cmd, subject, related, relation, note); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&relation, "relation", "",
		"One of: "+strings.Join(kanban.BacklogSemanticRelations, ", "))
	cmd.Flags().StringVar(&note, "note", "",
		"Free text recorded with the finding")
	return cmd
}

// runTodoRelate validates the relation and both ids, then appends exactly
// one agent finding under the lock.
func runTodoRelate(cmd *cobra.Command, subject, related, relation, note string) error {
	if !isSemanticRelation(relation) {
		return fmt.Errorf("todo relate: --relation must be one of %s (got %q)",
			strings.Join(kanban.BacklogSemanticRelations, ", "), relation)
	}
	if subject == related {
		return fmt.Errorf("todo relate: a card cannot be related to itself (%s)", subject)
	}
	var index int
	err := newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {
		for _, id := range []string{subject, related} {
			if !todoCardExists(rec, id) {
				return fmt.Errorf("todo relate: no card %s in the queue", id)
			}
		}
		finding := kanban.BacklogFinding{
			SubjectID: subject,
			RelatedID: related,
			Relation:  relation,
			Source:    kanban.BacklogSourceAgent,
			Note:      note,
			At:        time.Now().UTC().Format(time.RFC3339),
		}
		if !rec.AppendFindingOnce(finding) {
			return fmt.Errorf("todo relate: %s %s %s is already recorded", subject, relation, related)
		}
		index = len(rec.Findings)
		return nil
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "recorded %d %s %s %s\n", index, subject, relation, related)
	return nil
}

// newTodoUnrelateCmd — `moai todo unrelate <index>` (REQ-TA-008): remove the
// addressed finding, changing no card.
//
// The address is the 1-based index `todo why` prints, not a pair: a pair can
// carry several findings, and removing "the one about t1 and t2" would be
// ambiguous exactly when it matters.
func newTodoUnrelateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unrelate <index>",
		Short: "Remove one recorded finding by its index (changes no card)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			index, err := strconv.Atoi(strings.TrimSpace(args[0]))
			if err != nil || index < 1 {
				err = fmt.Errorf("todo unrelate: index must be a positive integer (got %q)", args[0])
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			var removed kanban.BacklogFinding
			mutErr := newTodoStore().Mutate(func(rec *kanban.BacklogRecord) error {
				if index > len(rec.Findings) {
					return fmt.Errorf("todo unrelate: no finding %d (the queue has %d)",
						index, len(rec.Findings))
				}
				removed = rec.Findings[index-1]
				rec.Findings = append(rec.Findings[:index-1], rec.Findings[index:]...)
				return nil
			})
			if mutErr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", mutErr)
				return mutErr
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "removed %d %s %s %s\n",
				index, removed.SubjectID, removed.Relation, removed.RelatedID)
			return nil
		},
	}
}

// isSemanticRelation reports whether r is one of the four relations `relate`
// accepts. The mechanical relations are deliberately excluded: a hand-written
// `near-duplicate` would claim a measurement nobody measured.
func isSemanticRelation(r string) bool {
	for _, allowed := range kanban.BacklogSemanticRelations {
		if r == allowed {
			return true
		}
	}
	return false
}

// todoCardExists reports whether the queue holds a card with id.
func todoCardExists(rec *kanban.BacklogRecord, id string) bool {
	for _, it := range rec.Items {
		if it.ID == id {
			return true
		}
	}
	return false
}
