package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/spf13/cobra"
)

// newFactoryPickupCommand is the classified self-service pickup consumption
// surface of SPEC-FACTORY-LANE-AUTONOMY-001 fragment 2 (design.md D2). The
// consumer only: the classification schema stays t1332's producer concern
// (spec.md §F), so until t1332 lands every card evaluates through the
// fallback classification — the REQ-FLA-007 single-dispatch default. A
// decision is a report, not a failure: even a denied sequential pickup
// exits 0 with the holder named, like the probe's unavailable verdict.
func newFactoryPickupCommand() *cobra.Command {
	pickup := &cobra.Command{
		Use:   "pickup",
		Short: "Classified self-service pickup rules (fallback default until t1332 lands)",
	}
	pickup.AddCommand(newFactoryPickupPlanCommand())
	return pickup
}

// newFactoryPickupPlanCommand evaluates one candidate card against the
// REQ-FLA-006/007/008 rules: sequential work is pickable only while no other
// lane holds the same sequential group, parallel work is pickable
// concurrently, and absent or unrecognized metadata follows the fallback
// classification with the tolerated-unknown condition logged.
func newFactoryPickupPlanCommand() *cobra.Command {
	var card, axis, group string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Evaluate whether this lane may pick a card under the classified-pickup rules",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			lane, err := factoryLaneLabelFromEnv("pickup plan")
			if err != nil {
				return err
			}
			if strings.TrimSpace(card) == "" {
				return fmt.Errorf("factory pickup plan: --card is required")
			}
			holds, err := factoryPickupHolds()
			if err != nil {
				return err
			}
			// The M2 consumption wiring (design.md D2): the candidate's
			// classification is what the flags carry — absent flags mean
			// absent metadata. When t1332 lands, its reader implements
			// factorylane.Classifier and replaces this adapter; the
			// tolerance point (NormalizeClassification) does not move.
			decision, err := factorylane.PlanPickup(lane, card, holds, factoryPickupFlagReader{card: card, axis: axis, group: group})
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.Marshal(decision)
				if err != nil {
					return fmt.Errorf("factory pickup plan: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}
			cls := "fallback (absent or unrecognized metadata — tolerated, REQ-FLA-008)"
			if decision.Classification.Known {
				cls = string(decision.Classification.Axis)
				if decision.Classification.Group != "" {
					cls += " group " + decision.Classification.Group
				}
			}
			verdict := fmt.Sprintf("decision: allowed — %s", decision.Reason)
			if !decision.Allowed {
				verdict = fmt.Sprintf("decision: denied — waiting on %s (%s)", decision.WaitOn, decision.Reason)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "card: %s\nlane: %s\nclassification: %s\n%s\nmulti-pick: %t\nholds considered: %d\n",
				decision.Card, decision.Lane, cls, verdict, decision.MultiPick, len(holds))
			return nil
		},
	}
	cmd.Flags().StringVar(&card, "card", "", "Candidate card id")
	_ = cmd.MarkFlagRequired("card")
	cmd.Flags().StringVar(&axis, "axis", "", "Execution axis read from the card metadata (sequential | parallel); absent until t1332 lands")
	cmd.Flags().StringVar(&group, "group", "", "Sequential group name read from the card metadata")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	return cmd
}

// factoryPickupFlagReader serves one card's flag-carried classification —
// the declared minimal consumption wiring of M2. The candidate reads what
// the caller supplied, normalized through factorylane.NormalizeClassification;
// any other card id (a held card) reads unknown — with no producer, no card
// carries metadata, so no hold ever scopes a sequential group yet.
type factoryPickupFlagReader struct {
	card  string
	axis  string
	group string
}

// Classify implements factorylane.Classifier over the flag input.
func (r factoryPickupFlagReader) Classify(cardID string) (factorylane.Classification, error) {
	if cardID != r.card {
		return factorylane.Classification{Axis: factorylane.AxisFallback}, nil
	}
	return factorylane.NormalizeClassification(r.axis, r.group), nil
}

// factoryPickupHolds reads the F1 lease model's active holds — card records
// in a lease-holding state with a holder label. The exclusivity decision
// rides these records; no second serialization mechanism is introduced
// (REQ-FLA-006, design.md D3). A missing database reads as no holds:
// nothing has been leased yet.
func factoryPickupHolds() ([]factorylane.Hold, error) {
	root := factoryCardRoot()
	path, err := homestate.FactoryDBPath(root)
	if err != nil {
		return nil, fmt.Errorf("factory pickup: %w", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("factory pickup: %w", statErr)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return nil, fmt.Errorf("factory pickup: %w", err)
	}
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	cards, err := db.ListCards(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("factory pickup: %w", err)
	}
	var holds []factorylane.Hold
	for _, c := range cards {
		if !homestate.IsLeaseHoldingState(c.State) {
			continue
		}
		holder := c.LeaseHolder
		if holder == "" {
			holder = c.OwnerLabel
		}
		if holder == "" {
			continue
		}
		holds = append(holds, factorylane.Hold{Lane: holder, Card: c.CardID})
	}
	return holds, nil
}
