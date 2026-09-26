package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// factoryCardNow is the clock the factory card commands read; tests replace it
// to drive lease expiry without sleeping.
var factoryCardNow = time.Now

// factoryCardRoot is the project root the factory record and the queue share.
func factoryCardRoot() string { return resolveTodoQueueRoot() }

// resolveFactoryCardRun returns the explicit --run value, or the single active
// factory run.
func resolveFactoryCardRun(ctx context.Context, root, explicit string) (string, error) {
	if run := strings.TrimSpace(explicit); run != "" {
		return run, nil
	}
	run, err := factorymsg.ResolveActiveRun(ctx, root, "")
	if err != nil {
		return "", fmt.Errorf("no --run given and no single active factory run (%w)", err)
	}
	return run, nil
}

// queueItemState reads the queue state of cardID; ok is false when the id is
// not in the queue at all. The factory record never writes the queue.
func queueItemState(cardID string) (kanban.BacklogState, bool, error) {
	record, err := newTodoReadStore().LoadPure()
	if err != nil {
		return "", false, err
	}
	for _, item := range record.Items {
		if item.ID == cardID {
			return item.State, true, nil
		}
	}
	for _, entry := range record.Archived {
		if entry.Item.ID == cardID {
			return entry.Item.State, true, nil
		}
	}
	return "", false, nil
}

// requireQueuePicked is the REQ-FR-022 precondition: only a card whose queue
// item is `picked` is admitted to the factory record.
func requireQueuePicked(cardID string) error {
	state, ok, err := queueItemState(cardID)
	if err != nil {
		return fmt.Errorf("read queue: %w", err)
	}
	if !ok {
		return fmt.Errorf("queue item %s is not in the queue", cardID)
	}
	if state != kanban.BacklogStatePicked {
		return fmt.Errorf("queue item %s is %s, not picked", cardID, state)
	}
	return nil
}

func newFactoryAssignCommand() *cobra.Command {
	var to, prefer, after, spec, worktree, contractRef, run string
	cmd := &cobra.Command{
		Use:   "assign <card>",
		Short: "Record a queue-picked card in the factory record, optionally assigning it to a lane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cardID := args[0]
			if !homestate.ValidCardID(cardID) {
				return fmt.Errorf("factory assign: %q is not a card id", cardID)
			}
			fields := homestate.CardFields{}
			flag := func(name string, v *string) *string {
				if cmd.Flags().Changed(name) {
					return v
				}
				return nil
			}
			fields.HintPrefer, fields.HintAfter, fields.SpecID = flag("prefer", &prefer), flag("after", &after), flag("spec", &spec)
			if cmd.Flags().Changed("worktree") {
				abs := worktree
				if abs != "" {
					var err error
					if abs, err = filepath.Abs(worktree); err != nil {
						return fmt.Errorf("factory assign: %w", err)
					}
				}
				fields.WorktreePath = &abs
			}
			if cmd.Flags().Changed("contract-ref") {
				ref, err := homestate.ParseContractRef(contractRef)
				if err != nil {
					return fmt.Errorf("factory assign: %w", err)
				}
				fields.Contract = &ref
			}
			if err := requireQueuePicked(cardID); err != nil {
				return fmt.Errorf("factory assign: %w", err)
			}
			root := factoryCardRoot()
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			runID, err := resolveFactoryCardRun(ctx, root, run)
			if err != nil {
				return fmt.Errorf("factory assign: %w", err)
			}
			db, err := homestate.OpenFactory(root)
			if err != nil {
				return fmt.Errorf("factory assign: %w", err)
			}
			defer func() { _ = db.Close() }()
			now := factoryCardNow()
			card, err := db.RecordPicked(ctx, runID, cardID, fields, "assign", now)
			if err != nil {
				return fmt.Errorf("factory assign: %w", err)
			}
			if strings.TrimSpace(to) != "" {
				card, err = db.Transition(ctx, homestate.TransitionRequest{RunID: runID, CardID: cardID, To: homestate.CardAssigned, ExpectedVersion: card.Version, Actor: "assign", Owner: to, Now: now})
				if err != nil {
					return fmt.Errorf("factory assign: %w", err)
				}
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s v%d owner=%s\n", card.CardID, card.State, card.Version, dash(card.OwnerLabel))
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "assign the card to this lane label (picked → assigned)")
	cmd.Flags().StringVar(&prefer, "prefer", "", "assignment preference hint, key=value (reported, never enforced)")
	cmd.Flags().StringVar(&after, "after", "", "predecessor card that must reach merged-local first (\"\" clears)")
	cmd.Flags().StringVar(&spec, "spec", "", "SPEC identifier for the card")
	cmd.Flags().StringVar(&worktree, "worktree", "", "card worktree path")
	cmd.Flags().StringVar(&contractRef, "contract-ref", "", "contract pointer <spec-id>,<sha256>,<signed-at>[,<event>]")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// factoryCardView is one card as `status` reports it.
type factoryCardView struct {
	RunID          string                 `json:"run_id"`
	CardID         string                 `json:"card_id"`
	State          string                 `json:"state"`
	Legacy         bool                   `json:"legacy"`
	Stage          string                 `json:"stage"`
	Version        int64                  `json:"version"`
	Owner          string                 `json:"owner"`
	LeaseHolder    string                 `json:"lease_holder"`
	LeaseExpiresAt string                 `json:"lease_expires_at"`
	LeaseExpired   bool                   `json:"lease_expired"`
	DecisionGate   string                 `json:"decision_gate"`
	Question       string                 `json:"decision_question"`
	Resume         string                 `json:"decision_resume"`
	Prefer         string                 `json:"prefer"`
	After          string                 `json:"after"`
	SpecID         string                 `json:"spec_id"`
	FailureReason  string                 `json:"failure_reason"`
	Contract       *homestate.ContractRef `json:"contract"`
}

type factoryStatusReport struct {
	Run   string            `json:"run"`
	Cards []factoryCardView `json:"cards"`
	// Unavailable lists the dispatch mirror writes that failed and have not
	// been reconciled by a later successful write (REQ-FR-025).
	Unavailable []homestate.RecordUnavailableEntry `json:"unavailable"`
}

func factoryCardViewOf(c homestate.Card, now time.Time) factoryCardView {
	v := factoryCardView{
		RunID: c.RunID, CardID: c.CardID, State: c.State, Legacy: c.Legacy(), Stage: c.Stage, Version: c.Version,
		Owner: c.OwnerLabel, LeaseHolder: c.LeaseHolder, LeaseExpiresAt: c.LeaseExpiresAt, LeaseExpired: c.LeaseExpired(now),
		DecisionGate: c.DecisionGate, Question: c.DecisionQuestion, Resume: c.DecisionResume,
		Prefer: c.HintPrefer, After: c.HintAfter, SpecID: c.SpecID, FailureReason: c.FailureReason,
	}
	if c.ContractSpecID != "" {
		v.Contract = &homestate.ContractRef{SpecID: c.ContractSpecID, SHA256: c.ContractSHA256, SignedAt: c.ContractSignedAt, Event: c.ContractEvent}
	}
	return v
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func newFactoryStatusCommand() *cobra.Command {
	var run string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report factory card records (read-only; an expired lease is shown, never returned)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := factoryCardRoot()
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			report := factoryStatusReport{Run: strings.TrimSpace(run), Cards: []factoryCardView{}}
			path, err := homestate.FactoryDBPath(root)
			if err != nil {
				return fmt.Errorf("factory status: %w", err)
			}
			// Never create a database just to report that it is empty.
			if _, statErr := os.Stat(path); statErr == nil {
				db, err := homestate.OpenFactory(root)
				if err != nil {
					return fmt.Errorf("factory status: %w", err)
				}
				cards, err := db.ListCards(ctx, report.Run)
				_ = db.Close()
				if err != nil {
					return fmt.Errorf("factory status: %w", err)
				}
				now := factoryCardNow()
				for _, c := range cards {
					report.Cards = append(report.Cards, factoryCardViewOf(c, now))
				}
			} else if !errors.Is(statErr, os.ErrNotExist) {
				return fmt.Errorf("factory status: %w", statErr)
			}
			entries, err := homestate.ReadRecordUnavailable(root, report.Run)
			if err != nil {
				return fmt.Errorf("factory status: %w", err)
			}
			report.Unavailable = append([]homestate.RecordUnavailableEntry{}, entries...)
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			writeFactoryStatusText(cmd.OutOrStdout(), report)
			return nil
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: every run)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "print JSON")
	return cmd
}

func writeFactoryStatusText(w io.Writer, r factoryStatusReport) {
	if len(r.Cards) == 0 {
		_, _ = fmt.Fprintln(w, "no factory card records")
	}
	for _, c := range r.Cards {
		lease := "none"
		if c.LeaseHolder != "" {
			lease = c.LeaseHolder + " until " + c.LeaseExpiresAt
			if c.LeaseExpired {
				lease = c.LeaseHolder + " expired " + c.LeaseExpiresAt
			}
		}
		state := c.State
		if c.Legacy {
			state += " (legacy)"
		}
		contract := "none"
		if c.Contract != nil {
			contract = strings.Join([]string{c.Contract.SpecID, c.Contract.SHA256, c.Contract.SignedAt, dash(c.Contract.Event)}, ",")
		}
		_, _ = fmt.Fprintf(w, "%s run=%s state=%s stage=%s version=%d owner=%s lease=%s gate=%s prefer=%s after=%s contract=%s\n",
			c.CardID, c.RunID, state, dash(c.Stage), c.Version, dash(c.Owner), lease, dash(c.DecisionGate), dash(c.Prefer), dash(c.After), contract)
		if c.Question != "" {
			_, _ = fmt.Fprintf(w, "  question: %s (resumes to %s)\n", c.Question, dash(c.Resume))
		}
	}
	for _, e := range r.Unavailable {
		_, _ = fmt.Fprintf(w, "%s run=%s card=%s lane=%s at=%s error=%s\n", factoryRecordUnavailableTag, e.RunID, e.CardID, e.Lane, e.At, e.Error)
	}
}

func newFactoryDecideCommand() *cobra.Command {
	var gate, choice, decider, run string
	cmd := &cobra.Command{
		Use:   "decide <card>...",
		Short: "Record an operator decision: --gate kickoff --choice approve|reject, --gate push, or --choice resume|block|unblock|abandon",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if decider != homestate.DeciderHuman {
				return fmt.Errorf("factory decide: decider %q is not accepted; F1 records only %q decisions", decider, homestate.DeciderHuman)
			}
			switch {
			case gate == "kickoff" && (choice == "approve" || choice == "reject"):
			case gate == "push" && choice == "":
			case gate == "" && (choice == "resume" || choice == "block" || choice == "unblock" || choice == "abandon"):
			default:
				return fmt.Errorf("factory decide: want --gate kickoff --choice approve|reject, --gate push, or --choice resume|block|unblock|abandon")
			}
			root := factoryCardRoot()
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			runID, err := resolveFactoryCardRun(ctx, root, run)
			if err != nil {
				return fmt.Errorf("factory decide: %w", err)
			}
			db, err := homestate.OpenFactory(root)
			if err != nil {
				return fmt.Errorf("factory decide: %w", err)
			}
			defer func() { _ = db.Close() }()
			integration := config.LoadGitFlowIntegrationConfig(root).IntegrationTarget
			refused := 0
			for _, cardID := range args {
				card, err := decideOne(ctx, db, runID, cardID, gate, choice, integration)
				if err != nil {
					refused++
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: refused: %v\n", cardID, err)
					continue
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s: %s (v%d)\n", cardID, card.State, card.Version)
			}
			if refused > 0 {
				return fmt.Errorf("factory decide: %d of %d cards refused", refused, len(args))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&gate, "gate", "", "decision gate: kickoff or push")
	cmd.Flags().StringVar(&choice, "choice", "", "approve|reject (kickoff), or resume|block|unblock|abandon")
	cmd.Flags().StringVar(&decider, "decider", homestate.DeciderHuman, "who decided (F1 accepts only human)")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// decideOne applies one card's decision as its own version-checked
// transition; a refusal for one card does not affect the others.
func decideOne(ctx context.Context, db *homestate.FactoryDB, runID, cardID, gate, choice, integration string) (homestate.Card, error) {
	cur, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return homestate.Card{}, err
	}
	want := func(state string) error {
		if cur.State != state {
			return fmt.Errorf("card is %s, not %s", cur.State, state)
		}
		return nil
	}
	var to string
	switch {
	case gate == "kickoff":
		if err := want(homestate.CardKickoff); err != nil {
			return cur, err
		}
		to = homestate.CardAssigned
		if choice == "reject" {
			to = homestate.CardBlocked
		}
	case gate == "push":
		if err := want(homestate.CardMergedLocal); err != nil {
			return cur, err
		}
		if to, err = homestate.PushGateTarget(ctx, cur); err != nil {
			return cur, err
		}
		if to == homestate.CardPushed && integration == "" {
			return cur, errors.New("no integration branch is configured (git_strategy develop_branch)")
		}
	case choice == "resume":
		if err := want(homestate.CardNeedsDecision); err != nil {
			return cur, err
		}
		if to = homestate.ResumeTarget(cur.DecisionResume); to == "" {
			return cur, fmt.Errorf("card records no resumable state (%q)", cur.DecisionResume)
		}
	case choice == "block":
		if err := want(homestate.CardNeedsDecision); err != nil {
			return cur, err
		}
		to = homestate.CardBlocked
	case choice == "unblock":
		if err := want(homestate.CardBlocked); err != nil {
			return cur, err
		}
		to = homestate.CardAssigned
	case choice == "abandon":
		to = homestate.CardAbandoned
	}
	return db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: cardID, To: to, ExpectedVersion: cur.Version,
		Actor: "operator", Decider: homestate.DeciderHuman, IntegrationBranch: integration, Now: factoryCardNow(),
	})
}
