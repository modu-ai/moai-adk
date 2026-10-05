// leader_approval_gate.go — the backlog-side plumbing for the leader
// approval receipt (SPEC-FACTORY-COMPLETION-RECOVERY-001 M1, REQ-FCR-001/002/005):
//
//   - the close-time gate every backlog completion surface (manual `todo
//     done`, the `todo --auto` cycle) wraps its archive in. The gate holds a
//     factory write transaction across verification and archive, so the
//     archive-moment recheck is serialized with concurrent factory
//     transitions (REQ-FCR-002a/004) and binds the row as it stands at the
//     close, never as it stood at scan time.
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

// backlogApprovalGate couples an open ApprovalGate with its factory DB
// handle so the caller can verify, then settle (commit after the guarded
// archive, rollback on any refusal or failure), then release the
// connection — in that order, every time. All methods are safe on the nil
// receiver and on an unlinked card: the gate simply does not apply
// (REQ-FCR-002's scope sentence — a non-factory card keeps its receipt-less
// completion).
type backlogApprovalGate struct {
	db   *homestate.FactoryDB
	gate *homestate.ApprovalGate
}

// holdDoneApprovalGate opens the gate for a queue card's close. A missing
// factory database means the project carries no factory state at all — no
// factory-linked cards exist, so it returns a nil gate without creating the
// database as a side effect.
func holdDoneApprovalGate(ctx context.Context, root, cardID string) (*backlogApprovalGate, error) {
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		return nil, nil
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	gate, err := db.HoldApprovalGate(ctx, cardID)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("leader approval gate: %w", err)
	}
	return &backlogApprovalGate{db: db, gate: gate}, nil
}

// verify checks the receipt bound to the card's backlog uuid against the
// archive-moment factory row.
func (g *backlogApprovalGate) verify(ctx context.Context, cardUUID string) error {
	if g == nil {
		return nil
	}
	return g.gate.Verify(ctx, cardUUID)
}

// commit settles the gate after the archive it guarded has landed inside the
// callback. A commit failure rolls the transaction back: the caller's
// Mutate callback then returns the error and the backlog write is discarded
// too, leaving both stores untouched.
func (g *backlogApprovalGate) commit() error {
	if g == nil {
		return nil
	}
	err := g.gate.Commit()
	_ = g.db.Close()
	if err != nil {
		return fmt.Errorf("leader approval gate: %w", err)
	}
	return nil
}

// refuse settles the gate on any refusal or archive failure — the
// transaction wrote nothing either way.
func (g *backlogApprovalGate) refuse() {
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
			row, err := db.LatestCardByID(ctx, cardID)
			if errors.Is(err, homestate.ErrCardNotFound) {
				return fmt.Errorf("factory approve: card %s has no factory card row — not factory-linked, nothing to approve", cardID)
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
