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
// The semantic relations are the judgements a text analyser cannot reach.
// contains / absorbs / replaces / conflicts came first; blocks / depends
// joined them (card t1309) so card sequencing stops living in prose alone —
// `A blocks B` reads "A must land before B proceeds", `A depends B` reads
// "A waits on B". Since SPEC-RELATION-PICKUP-FILTER-001 the sequencing pair
// is no longer purely observational: the todo --auto pickup selection reads
// it (a relation-blocked card is skipped with a labelled non-finding), and
// the write path below refuses a relation that would close a waits-on
// cycle. The other four relations stay record-only — the operator decides,
// exactly as before.
//
// SUBAGENT BOUNDARY (REQ-TA-015): nothing here prompts.
package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// newTodoRelateCmd — `moai todo relate <a> <b> --relation <r> [--note <text>]`
// (REQ-TA-008): record one agent-sourced finding between two existing cards.
// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-011: `--disposition <value>` turns the
// verb into the disposition recorder — it sets the disposition of the
// matching finding pair (either order) and records no new relation.
func newTodoRelateCmd() *cobra.Command {
	var relation, note, disposition string
	cmd := &cobra.Command{
		Use:   "relate <a> <b> --relation <contains|absorbs|replaces|conflicts|blocks|depends>",
		Short: "Record a relation between two cards (records only — changes no card)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			subject, related := normalizeTodoRef(args[0]), normalizeTodoRef(args[1])
			if disposition != "" {
				if relation != "" {
					err := fmt.Errorf("todo relate: --disposition and --relation are mutually exclusive")
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
					return err
				}
				if err := factory.RecordFindingDisposition(newTodoStore(), subject, related, disposition); err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "disposition %s recorded on the %s/%s finding\n", disposition, subject, related)
				return nil
			}
			if err := runTodoRelate(cmd, subject, related, relation, note); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&relation, "relation", "",
		"One of: "+strings.Join(factory.BacklogWriteableRelations, ", "))
	cmd.Flags().StringVar(&note, "note", "",
		"Free text recorded with the finding")
	cmd.Flags().StringVar(&disposition, "disposition", "",
		"One of: "+strings.Join(factory.IssuanceDispositionValues, ", ")+" — set the finding's disposition instead of recording a relation")
	return cmd
}

// runTodoRelate validates the relation and both ids, then appends exactly
// one agent finding under the lock.
func runTodoRelate(cmd *cobra.Command, subject, related, relation, note string) error {
	if !isWriteableRelation(relation) {
		return fmt.Errorf("todo relate: --relation must be one of %s (got %q)",
			strings.Join(factory.BacklogWriteableRelations, ", "), relation)
	}
	if subject == related {
		return fmt.Errorf("todo relate: a card cannot be related to itself (%s)", subject)
	}
	// SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013: symmetric kinds normalize the
	// pair (smaller id first) BEFORE the dedup check, so an opposite-order
	// re-record maps onto the first finding instead of creating a second.
	if factory.BacklogRelationIsSymmetricForDedup(relation) {
		subject, related, _ = factory.NormalizeRelationPair(subject, related)
	}
	var index int
	err := newTodoStore().Mutate(func(rec *factory.BacklogRecord) error {
		for _, id := range []string{subject, related} {
			if !todoCardExists(rec, id) {
				return fmt.Errorf("todo relate: no card %s in the queue", id)
			}
		}
		// SPEC-RELATION-PICKUP-FILTER-001 (REQ-RPF-005): a candidate
		// sequencing relation is refused BEFORE the write when it would
		// close a directed cycle in the waits-on graph formed by the
		// recorded findings — the record stays unchanged, and the error
		// names both endpoints (waiter and target are exactly the two
		// argument ids, whichever spelling the caller used).
		if waiter, target, ok := factory.WaitsOnOf(factory.BacklogFinding{
			SubjectID: subject,
			RelatedID: related,
			Relation:  relation,
		}); ok {
			if rec.WaitsOnClosesCycle(waiter, target) {
				return fmt.Errorf("todo relate: %s %s %s would close a dependency cycle (%s already waits on %s through recorded relations)",
					subject, relation, related, waiter, target)
			}
		}
		// REQ-TCI-013: supersedes cycles are refused the same way — the
		// mapped walk sees legacy replaces rows as supersedes edges.
		if relation == "supersedes" {
			if rec.RelationKindClosesCycle(subject, related, "supersedes", "replaces") {
				return fmt.Errorf("todo relate: %s supersedes %s would close a supersedes cycle", subject, related)
			}
		}
		finding := factory.BacklogFinding{
			SubjectID: subject,
			RelatedID: related,
			Relation:  relation,
			Source:    factory.BacklogSourceAgent,
			Note:      note,
			At:        time.Now().UTC().Format(time.RFC3339),
		}
		// REQ-TCI-013 (card t1454 card-review r2 finding 15): the stored rows
		// are normalized in the dedup too — a pair an older writer recorded
		// in the opposite order still maps onto the first record.
		if rec.HasNormalizedFindingTuple(finding) {
			return fmt.Errorf("todo relate: %s %s %s is already recorded", subject, relation, related)
		}
		rec.Findings = append(rec.Findings, finding)
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
			var removed factory.BacklogFinding
			mutErr := newTodoStore().Mutate(func(rec *factory.BacklogRecord) error {
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

// isWriteableRelation — SPEC-TODO-CARD-ISSUANCE-001 REQ-TCI-013: relate's
// writable set widens to the three new kinds; the projection kinds stay
// refused (they are add-time attribute projections or todo merge's output).
func isWriteableRelation(r string) bool {
	for _, allowed := range factory.BacklogWriteableRelations {
		if r == allowed {
			return true
		}
	}
	return false
}

// todoCardExists reports whether the queue holds a card with id.
func todoCardExists(rec *factory.BacklogRecord, id string) bool {
	for _, it := range rec.Items {
		if it.ID == id {
			return true
		}
	}
	return false
}
