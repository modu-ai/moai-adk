// todo_classify_cmd.go — SPEC-FACTORY-DEFAULT-MODE-001 M2: the official
// reclassify verb `moai todo classify <id> --mode <serial|parallelizable>
// --reason <text>`.
//
// The verb records ONE human judgment in exactly one locked queue write —
// one Mutate acquisition (REQ-FDM-004) — and refuses by early return inside
// the callback, so every refusal leaves the queue file byte-identical (C5,
// the hold-verb contract). The queue is the only store carrying the mode:
// the factory record carries no mode field and every factory-side reader
// derives from the queue at read time, so this single write is also the
// synchronization point for selection, nominated validation, and the status
// view (REQ-FDM-005, plan D3) — there is no second write to make.
//
// The judgment is HUMAN (decider fixed at human, C6): the verb accepts no
// decider flag, and it records the given mode and reason verbatim, leaving
// the card's priority and blocked fields, its state and text, and every
// other queue item untouched. The decider-failure fail-safe is untouched by
// this verb: the add path's fallback stays serial (REQ-FDM-003).
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// classifyRecoveryAdvice names the path that returns a card to a classifiable
// state, so a refusal tells the operator the two-step path instead of only
// the fact of the refusal (the hold-verb advice pattern).
func classifyRecoveryAdvice(state factory.BacklogState) string {
	switch state {
	case factory.BacklogStateHold:
		return "unhold it first"
	case factory.BacklogStateDropped:
		return "undrop it first"
	default:
		return "no recovery verb is defined for that state"
	}
}

// todoClassifySerialSlotConflict reports whether cardID is in flight — its
// homestate factory card row is slot-held (the selection gate's own
// factorySerialSlotHeld read) — while another DISTINCT serial card holds the
// serial slot (factorySerialInFlightExcluding over the record's cards:
// slot-held ∧ mode serial), so the verb adds no second serial-overlap policy
// (plan D4, decision-index Q6). The gate keys on the factory record's row,
// NEVER on the queue row's state: `todo unpick` rewrites only the queue row,
// so a `queued` queue row can still be in flight with a live lease (card
// t1737 F1) — a queue-state gate would let a second serial card be
// reclassified beside the un-picked holder. A target with no factory row, or
// whose rows have all released the slot, is not in flight: making it serial
// cannot leave a second in-flight card in serial mode, and the selection gate
// still orders the lease itself. The queue snapshot is the verb's own
// in-Mutate record, so the classification read and the refused-or-not write
// share one lock acquisition. A root with no factory record has no in-flight
// cards: the database is never created just to answer that there is nothing
// in flight (the status surface's no-create rule), and ONLY a not-exist stat
// says the record is absent — any other stat error leaves the record
// unreadable, not absent, and is returned so the verb refuses with nothing
// written (fail-closed; card t1737 F2).
//
// The returned holder id names the conflicting card for the refusal message;
// it is message enrichment only — the verdict is the shared predicate's.
func todoClassifySerialSlotConflict(root string, rec *factory.BacklogRecord, cardID string, now time.Time) (conflict bool, holder string, err error) {
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return false, "", fmt.Errorf("todo classify: resolve the factory record: %w", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if !os.IsNotExist(statErr) {
			return false, "", fmt.Errorf("todo classify: stat the factory record: %w", statErr)
		}
		// No factory record: no in-flight cards to conflict with.
		return false, "", nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return false, "", fmt.Errorf("todo classify: read the factory record: %w", err)
	}
	defer func() { _ = db.Close() }()
	cards, err := db.ListCards(context.Background(), "")
	if err != nil {
		return false, "", fmt.Errorf("todo classify: read the factory cards: %w", err)
	}
	targetInFlight := false
	for _, c := range cards {
		if c.CardID == cardID && factorySerialSlotHeld(c, now) {
			targetInFlight = true
			break
		}
	}
	if !targetInFlight {
		return false, "", nil
	}
	classOf := func(cid string) factory.CardClassification {
		return factoryQueueClassification(rec, cid)
	}
	if !factorySerialInFlightExcluding(cards, classOf, cardID, now, false) {
		return false, "", nil
	}
	for _, c := range cards {
		if c.CardID == cardID {
			continue
		}
		if c.State == homestate.CardPicked && (c.BundleID != "" || c.HintAfter != "") {
			continue
		}
		if factorySerialSlotHeld(c, now) && classOf(c.CardID).Mode == factory.ClassModeSerial {
			return true, c.CardID, nil
		}
	}
	return true, "", nil
}

// newTodoClassifyCmd — `moai todo classify <id> --mode <serial|parallelizable>
// --reason <text>`: record the operator's (or the leader speaking for one)
// execution-mode judgment on a queued or picked card as one locked write.
func newTodoClassifyCmd() *cobra.Command {
	var (
		mode   string
		reason string
	)
	cmd := &cobra.Command{
		Use:   "classify <id> --mode <serial|parallelizable> --reason <text>",
		Short: "Record a human classification judgment on a queued or picked card (operator or leader only)",
		Long: `Record one execution-mode judgment on the addressed queued or picked card
as ONE locked queue write: decider human, the given mode and reason
verbatim, a fresh classified-at stamp — the card's priority and blocked
fields, its state and text, and every other item untouched. The verb takes
no decider flag: the judgment is human, unconditionally.

An in-flight card (a live factory-record lease, whatever its queue row
reads) requested serial while ANOTHER distinct in-flight card holds the
serial slot is refused with nothing written — a reclassification never
leaves two serial cards in flight at once (the selection gate's own
exclusivity read). Every refusal leaves the queue file byte-identical.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := normalizeTodoRef(args[0])
			if mode != factory.ClassModeSerial && mode != factory.ClassModeParallelizable {
				return &exitCodeError{code: 2, msg: fmt.Sprintf("todo classify: --mode %q is not one of serial, parallelizable (nothing written)", mode)}
			}
			if strings.TrimSpace(reason) == "" {
				return &exitCodeError{code: 2, msg: "todo classify: --reason is required — name the file set or conflict the judgment rests on (nothing written)"}
			}
			root := factoryCardRoot()
			store := todoStoreAt(root)
			now := factoryCardNow()
			if err := store.Mutate(func(rec *factory.BacklogRecord) error {
				for i := range rec.Items {
					if rec.Items[i].ID != id {
						continue
					}
					it := &rec.Items[i]
					switch it.State {
					case factory.BacklogStateQueued, factory.BacklogStatePicked:
						// the classifiable states
					default:
						return &exitCodeError{code: 2, msg: fmt.Sprintf("todo classify: backlog item %s is %s, not classifiable — %s (nothing written)",
							id, it.State, classifyRecoveryAdvice(it.State))}
					}
					if mode == factory.ClassModeSerial {
						// The slot gate reads the factory record's own
						// in-flight state — never the queue row's state,
						// which `todo unpick` rewrites alone (card t1737).
						conflict, holder, err := todoClassifySerialSlotConflict(root, rec, id, now)
						if err != nil {
							return err
						}
						if conflict {
							name := holder
							if name == "" {
								name = "another in-flight serial card"
							}
							return &exitCodeError{code: 2, msg: fmt.Sprintf("todo classify: %s already holds the serial slot — refusing to make in-flight %s serial: wait for %s to land, reclassify the holder, or classify %s parallelizable instead (nothing written)",
								name, id, name, id)}
						}
					}
					// The seam hands a present judgment through and derives
					// absence: priority and blocked are the card's own, never
					// the verb's to change.
					cls := factory.EffectiveCardClassification(*it)
					cls.Mode = mode
					cls.Decider = factory.DeciderIdentityHuman
					cls.ClassifiedAt = now.UTC().Format(time.RFC3339)
					cls.Reason = reason
					if err := factory.ValidateCardClassification(cls); err != nil {
						return &exitCodeError{code: 2, msg: fmt.Sprintf("todo classify: %v (nothing written)", err)}
					}
					c := cls
					it.Classification = &c
					return nil
				}
				return &exitCodeError{code: 2, msg: fmt.Sprintf("todo classify: no backlog item %s (nothing written)", id)}
			}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "classified %s mode=%s\n", id, mode)
			return nil
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "",
		"The execution mode to record: serial or parallelizable")
	cmd.Flags().StringVar(&reason, "reason", "",
		"The judgment's reason, recorded verbatim — name the file set or the conflict it rests on")
	_ = cmd.MarkFlagRequired("mode")
	_ = cmd.MarkFlagRequired("reason")
	return cmd
}
