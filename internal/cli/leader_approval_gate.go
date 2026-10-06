// leader_approval_gate.go — the backlog-side plumbing for the leader
// approval receipt (SPEC-FACTORY-COMPLETION-RECOVERY-001 M1, REQ-FCR-001/002/005):
//
//   - the close-time gate every backlog completion surface (manual `todo
//     done`, the `todo --auto` cycle, the auto-done scan) wraps its archive
//     in. The gate holds a factory write transaction across verification AND
//     the queue write's persistence, so the archive-moment recheck is
//     serialized with concurrent factory transitions (REQ-FCR-002a/004) and
//     binds the row as it stands at the close, never as it stood at scan
//     time.
//
//   - `moai factory approve` — the leader path's receipt mint. Issuance is
//     reachable only here: lane sessions are refused at the boundary
//     (factoryLaneRefusal), and nothing else in the tree calls
//     IssueLeaderApproval. A card with no factory row is not factory-linked;
//     there is nothing to approve (REQ-FCR-002's scope sentence).
//
// SUBAGENT BOUNDARY (C-HRA-008): nothing here prompts; refusals are errors
// the command layer prints to stderr.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// backlogApprovalGate couples an open homestate.ApprovalGate with its factory
// DB handle so a close can verify, persist under the factory write lock, and
// then release — in that order, every time. All methods are safe on the nil
// receiver: a nil gate means the project carries no factory database at all,
// so no factory-linked card exists and the gate does not apply (REQ-FCR-002's
// scope sentence — a non-factory card keeps its receipt-less completion).
type backlogApprovalGate struct {
	db   *homestate.FactoryDB
	gate *homestate.ApprovalGate
}

// holdDoneApprovalGate opens the gate. A MISSING factory database (the path
// does not exist) means the project carries no factory state — the gate
// returns nil without creating the database as a side effect. Every other
// stat failure (permission denied, a symlink loop, ...) is a database we
// cannot verify against: fail closed and refuse the close rather than
// silently completing without the receipt check.
//
// The gate deliberately runs on context.Background(): the queue write it
// guards persists on a background context (LockedBacklog.Mutate), so a
// request cancellation that arrived after verification must not roll the
// factory transaction back while the save continues — the lock's lifetime
// covers the whole persistence window (review round-4 P2-1).
func holdDoneApprovalGate(_ context.Context, root string) (*backlogApprovalGate, error) {
	ctx := context.Background()
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("leader approval gate: %w", statErr)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	gate, err := db.HoldApprovalGate(ctx)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	return &backlogApprovalGate{db: db, gate: gate}, nil
}

// verifyForClose loads the archive-moment factory row inside the gate's
// transaction — resolved through the runs table's dispatch binding — and
// verifies the receipt bound to the card's backlog uuid. A card with no
// factory row passes — not factory-linked.
func (g *backlogApprovalGate) verifyForClose(ctx context.Context, cardID, cardUUID string) error {
	if g == nil {
		return nil
	}
	card, linked, err := g.gate.Row(ctx, cardID)
	if err != nil {
		return fmt.Errorf("leader approval gate: %w", err)
	}
	if !linked {
		return nil
	}
	return g.gate.Verify(ctx, card, cardUUID)
}

// release settles the gate. The gate's transaction never writes, so
// rollback is the settle path on every exit — after a refusal, after the
// guarded mutation failed, and after the guarded mutation has PERSISTED (the
// caller releases only once the queue write is done, keeping the factory
// write lock across the whole persistence window).
func (g *backlogApprovalGate) release() {
	if g == nil {
		return
	}
	_ = g.gate.Rollback()
	_ = g.db.Close()
}

// todoCardUUID extracts the projected identity of a backlog item; an absent
// identity is not a failure here — the receipt lookup will simply find
// nothing and refuse, which is the fail-closed answer for a factory-linked
// card whose identity was never minted.
func todoCardUUID(item *factory.BacklogItem) string {
	if item == nil || item.CardUUID == nil {
		return ""
	}
	return strings.TrimSpace(*item.CardUUID)
}

// scanApprovalStates reads, once per scan, the leader-approval receipt state
// of every live queued/picked candidate (REQ-FCR-003). A card with no
// factory row is not factory-linked and assembles ReceiptGateNone; a card
// whose receipt verifies against its recorded-run factory row assembles
// ReceiptGateVerified; an absent or non-binding receipt is
// ReceiptGateUnverified; an unreadable factory state is ReceiptGateUnknown —
// an unanswerable question, never a close.
//
// The scan opens the factory database STRICTLY READ-ONLY (review P2-2): no
// DDL, no migration, nothing created — a dry-run against an older-schema
// store reads it as it stands, and a missing leader_approvals table reads as
// "no receipts" (unverified), never as a reason to migrate.
func scanApprovalStates(ctx context.Context, root string, snapshot *factory.BacklogRecord) map[string]factory.ReceiptGateState {
	states := make(map[string]factory.ReceiptGateState)
	ids := make([]string, 0, len(snapshot.Items))
	for i := range snapshot.Items {
		switch snapshot.Items[i].State {
		case factory.BacklogStateQueued, factory.BacklogStatePicked:
			ids = append(ids, snapshot.Items[i].ID)
		}
	}
	if len(ids) == 0 {
		return states
	}
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		// The factory state cannot even be located: an unanswerable
		// question for every candidate — never a silent pass.
		for _, id := range ids {
			states[id] = factory.ReceiptGateUnknown
		}
		return states
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			// No factory database: no factory-linked cards exist; every
			// state stays the zero value (ReceiptGateNone).
			return states
		}
		for _, id := range ids {
			states[id] = factory.ReceiptGateUnknown
		}
		return states
	}
	db, err := homestate.OpenFactoryReadonly(path)
	if err != nil {
		for _, id := range ids {
			states[id] = factory.ReceiptGateUnknown
		}
		return states
	}
	defer func() { _ = db.Close() }()
	hasApprovals, err := db.FactoryTablePresent(ctx, "leader_approvals")
	if err != nil {
		for _, id := range ids {
			states[id] = factory.ReceiptGateUnknown
		}
		return states
	}
	for _, id := range ids {
		var cardUUID string
		for i := range snapshot.Items {
			if snapshot.Items[i].ID == id {
				cardUUID = todoCardUUID(&snapshot.Items[i])
				break
			}
		}
		if !hasApprovals {
			// An older-schema store: no receipts exist anywhere. A
			// factory-linked card reads as unverified — never as a
			// migration trigger.
			if _, linked, rerr := db.RecordedCardRowReadonly(ctx, id); rerr != nil {
				states[id] = factory.ReceiptGateUnknown
			} else if !linked {
				continue // not factory-linked: the axis does not apply
			} else {
				states[id] = factory.ReceiptGateUnverified
			}
			continue
		}
		switch err := db.VerifyApprovalReadonly(ctx, id, cardUUID); {
		case err == nil:
			states[id] = factory.ReceiptGateVerified
		case isApprovalRefusal(err):
			states[id] = factory.ReceiptGateUnverified
		case errors.Is(err, homestate.ErrCardNotFound):
			// Not factory-linked: the axis does not apply (ReceiptGateNone).
			continue
		default:
			// Includes ErrApprovalRunUnresolvable: the dispatch cannot be
			// determined — an unanswerable question, never a close.
			states[id] = factory.ReceiptGateUnknown
		}
	}
	return states
}

// isApprovalRefusal reports whether err is one of the receipt-gate refusal
// sentinels — an answer about the receipt, not about the database.
func isApprovalRefusal(err error) bool {
	return errors.Is(err, homestate.ErrApprovalMissing) ||
		errors.Is(err, homestate.ErrApprovalIssuer) ||
		errors.Is(err, homestate.ErrApprovalCardMismatch) ||
		errors.Is(err, homestate.ErrApprovalRunMismatch) ||
		errors.Is(err, homestate.ErrApprovalStale) ||
		errors.Is(err, homestate.ErrApprovalHashMismatch)
}

// newFactoryApproveCommand is `moai factory approve` — the leader path's
// issuance surface for leader approval receipts (REQ-FCR-001/014). The
// receipt binds what the card's factory row IS at issuance time (current
// version and evidence hash); the caller supplies only identities and the
// run, never a verdict.
//
// @MX:ANCHOR: [AUTO] the only CLI surface that mints a leader approval receipt
// @MX:REASON: receipt issuance is the leader path's exclusive authority (REQ-FCR-014); any second minting path would break the lane boundary this verb enforces
func newFactoryApproveCommand() *cobra.Command {
	var runID, issuer string
	cmd := &cobra.Command{
		Use:   "approve <card-id> --run <run-id>",
		Short: "Record the leader's evidence-review receipt for a factory card (leader path only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if factoryLaneRefusal() {
				return fmt.Errorf("factory approve: refused — %s: receipt issuance is the leader path's act (%s=%s marks a lane)",
					factoryLaneBoundarySentinel, config.EnvFactoryRole, config.FactoryRoleLane)
			}
			cardID := strings.TrimSpace(args[0])
			runID = strings.TrimSpace(runID)
			if cardID == "" || runID == "" {
				return fmt.Errorf("factory approve: a card id and --run are required")
			}
			root := factoryCardRoot()
			ctx := cmd.Context()

			// The uuid binding is the backlog identity: read it from the
			// queue record, never from a flag.
			rec, err := newTodoStore().LoadPure()
			if err != nil {
				return fmt.Errorf("factory approve: the queue could not be read: %w", err)
			}
			var cardUUID string
			for i := range rec.Items {
				if rec.Items[i].ID == cardID {
					cardUUID = todoCardUUID(&rec.Items[i])
					break
				}
			}
			if cardUUID == "" {
				for i := range rec.Archived {
					if rec.Archived[i].Item.ID == cardID {
						cardUUID = todoCardUUID(&rec.Archived[i].Item)
						break
					}
				}
			}
			if cardUUID == "" {
				return fmt.Errorf("factory approve: card %s carries no recorded identity — cannot bind a receipt", cardID)
			}

			db, err := homestate.OpenFactory(root)
			if err != nil {
				return err
			}
			defer func() { _ = db.Close() }()
			// The receipt binds THE NAMED RUN's row — the version and
			// evidence the leader actually reviewed in that run, never
			// whichever row happens to be the most recently modified
			// (review P2-1): approving --run run-old must read run-old's
			// row or refuse.
			row, err := db.LoadCard(ctx, runID, cardID)
			if errors.Is(err, homestate.ErrCardNotFound) {
				return fmt.Errorf("factory approve: card %s has no factory card row in run %s — nothing to approve", cardID, runID)
			}
			if err != nil {
				return err
			}
			if strings.TrimSpace(issuer) == "" {
				issuer = strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker))
			}
			if issuer == "" {
				issuer = factory.RoleLeader
			}
			approval, err := db.IssueLeaderApproval(ctx, homestate.LeaderApproval{
				CardUUID:       cardUUID,
				RunID:          runID,
				CardID:         cardID,
				FactoryVersion: row.Version,
				EvidenceHash:   row.EvidenceSHA,
				Issuer:         issuer,
				IssuerRole:     homestate.ApprovalIssuerLeader,
			})
			if err != nil {
				return err
			}
			evidence := approval.EvidenceHash
			if len(evidence) > 12 {
				evidence = evidence[:12]
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"approved %s run=%s version=%d evidence=%s issuer=%s\n",
				approval.CardID, approval.RunID, approval.FactoryVersion, evidence, approval.Issuer)
			return nil
		},
	}
	cmd.Flags().StringVar(&runID, "run", "", "factory run id the receipt binds")
	cmd.Flags().StringVar(&issuer, "issuer", "", "leader label recorded on the receipt (default: MOAI_FACTORY_WORKER, else \"leader\")")
	return cmd
}
