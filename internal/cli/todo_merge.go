// todo_merge.go — SPEC-TODO-CARD-ISSUANCE-001 M4 (REQ-TCI-019): the
// OPERATOR-invoked merge verb. Folding one card into another is the act the
// queue doctrine reserves to the operator by name — the analyser, `analyze`,
// `relate`, and lane sessions keep their "never folds" shape, and this file
// is the one place the shape gains its exception. The call-graph test pins
// that boundary mechanically (TestTodoMergeNeverInvokedByAnalysis): the
// merge entry function is called from no other production file.
//
// One locked write, or none: every refusal returns from inside Mutate's
// callback, so the queue file stays byte-identical on every refusal.
package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// newTodoMergeCmd — `moai todo merge <into> <from>`: fold `<from>`'s text
// into `<into>` as an explicit section, drop `<from>` with the reason
// `merged into <into>`, and record the merged-into relation. The relation
// kind is this verb's only writer — `relate` refuses it (the projection
// kinds stay out of the writable set).
func newTodoMergeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "merge <into> <from>",
		Short: "Fold one queued card's text into another and drop the source (operator only)",
		Long: `Append ` + "`[merged from <from>] <from text>`" + ` to the end of <into>'s text,
drop <from> with the reason ` + "`merged into <into>`" + ` (the dropped marker and
the drop-reason attribute, exactly as ` + "`drop`" + ` writes them), and record the
merged-into relation between the two — one locked write.

Refusals: either card picked, <from> already merged, <into> closed or
already merged, a merged-into cycle, or an id not in the queue. A lane
session cannot call the verb — the lane queue guard refuses it ahead of
RunE, like every other mutation.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			into, from := normalizeTodoRef(args[0]), normalizeTodoRef(args[1])
			if err := runTodoMerge(cmd, into, from); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return err
			}
			return nil
		},
	}
}

// runTodoMerge is the merge body — the single entry function the call-graph
// test pins to this file.
func runTodoMerge(cmd *cobra.Command, into, from string) error {
	if into == from {
		return fmt.Errorf("todo merge: a card cannot be merged into itself (%s)", into)
	}
	err := newTodoStore().Mutate(func(rec *factory.BacklogRecord) error {
		var intoItem, fromItem *factory.BacklogItem
		for i := range rec.Items {
			switch rec.Items[i].ID {
			case into:
				intoItem = &rec.Items[i]
			case from:
				fromItem = &rec.Items[i]
			}
		}
		if intoItem == nil {
			return fmt.Errorf("todo merge: no card %s in the queue", into)
		}
		if fromItem == nil {
			return fmt.Errorf("todo merge: no card %s in the queue", from)
		}
		// Either side picked: the lease belongs to a lane, and folding a
		// leased card's text out from under it is exactly the interference
		// the refusal exists for (REQ-TCI-019).
		for _, item := range []*factory.BacklogItem{intoItem, fromItem} {
			if item.State == factory.BacklogStatePicked {
				return fmt.Errorf("todo merge: %s is picked — unpick it first; a picked card's text belongs to its lane", item.ID)
			}
		}
		mergedAlready := func(id string) bool {
			for _, f := range rec.Findings {
				if f.Relation == string(factory.CardRelationMergedInto) && f.SubjectID == id {
					return true
				}
			}
			return false
		}
		// <into> open means queued or hold; everything else is closed.
		switch intoItem.State {
		case factory.BacklogStateQueued, factory.BacklogStateHold:
			// the open states
		default:
			return fmt.Errorf("todo merge: %s is %s — merge into an open card", into, intoItem.State)
		}
		if mergedAlready(into) {
			return fmt.Errorf("todo merge: %s is already merged into another card — a second merged-into target is refused", into)
		}
		// <from> must be queued: the verb destroys its state (dropped), and
		// only a queued card's state is the verb's to take. POSITIVE
		// enumeration (REQ-THS-012): the accepted state is named; every
		// other state — a state added later included — falls through.
		switch fromItem.State {
		case factory.BacklogStateQueued:
			// the only state the verb takes <from> from
		default:
			return fmt.Errorf("todo merge: %s is %s, not queued", from, fromItem.State)
		}
		if mergedAlready(from) {
			return fmt.Errorf("todo merge: %s is already merged into another card", from)
		}
		// Cycle guard over the merged-into edges: adding <from>→<into> must
		// not close a directed cycle (the state refusals above already make
		// the two-card case unreachable; this closes the longer chains).
		if rec.RelationKindClosesCycle(from, into, string(factory.CardRelationMergedInto)) {
			return fmt.Errorf("todo merge: %s → %s would close a merged-into cycle", from, into)
		}
		// Fold the text: the section carries <from>'s CURRENT text, before
		// the dropped marker is written onto it.
		fromText := fromItem.Text
		intoItem.Text = strings.TrimRight(intoItem.Text, "\n") + "\n\n[merged from " + from + "] " + fromText
		// Drop <from> exactly the way `drop` writes a drop: marker, stamp,
		// state, and the machine-readable reason attribute.
		reason := "merged into " + into
		fromItem.Text = todoDropMarkerOpen + reason + todoDropMarkerClose + fromText
		fromItem.State = factory.BacklogStateDropped
		fromItem.DroppedAt = todoStampNow()
		if fromItem.Issuance == nil {
			fromItem.Issuance = &factory.BacklogIssuance{}
		}
		fromItem.Issuance.DropReason = reason
		// The merged-into finding, agent-sourced — the same source set
		// `relate` writes, so the sources enumeration stays closed.
		rec.Findings = append(rec.Findings, factory.BacklogFinding{
			SubjectID: from,
			RelatedID: into,
			Relation:  string(factory.CardRelationMergedInto),
			Source:    factory.BacklogSourceAgent,
			At:        time.Now().UTC().Format(time.RFC3339),
		})
		return nil
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "merged %s into %s\n", from, into)
	return nil
}
