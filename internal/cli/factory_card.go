package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// factoryCardNow is the clock the factory card commands read; tests replace it
// to drive lease expiry without sleeping.
var factoryCardNow = time.Now

// factoryCardRoot is the project root the factory record and the queue share.
func factoryCardRoot() string { return resolveTodoQueueRoot() }

// Lane predicates (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-015), deliberately
// asymmetric. Admission — the right to run `moai factory next`, `stage`, and
// `complete` — holds only when the factory role marker equals the role-value
// constant. Refusal — the duty NOT to mutate the queue or record decisions —
// additionally fires on a non-empty lane label and on the Codex backend
// value, because both of those ride the frozen Codex MCP env_vars allowlist:
// a Codex lane session has no role marker there, yet must stay refused.
// Widening only the deny direction can grant nothing (plan B3).

// factoryLaneAdmission reports lane admission: the role marker equals the
// role-value constant, read through the internal/config constants only
// (REQ-SD-017 forbids string literals at stamp/compare sites).
func factoryLaneAdmission() bool {
	return os.Getenv(config.EnvFactoryRole) == config.FactoryRoleLane
}

// factoryLaneRefusal reports lane refusal: admission, or a non-empty
// lane-label variable, or the backend variable naming the Codex harness.
func factoryLaneRefusal() bool {
	return factoryLaneAdmission() ||
		os.Getenv(config.EnvMoaiKanbanLabel) != "" ||
		os.Getenv(config.EnvMoaiKanbanBackend) == kanban.BackendGPT
}

// The refusal sentinels. One wording source per refusal kind, so the queue
// guard, the decide guard, and the lane verbs cannot drift apart.
const (
	// factoryLaneBoundarySentinel names the lane permission boundary in the
	// queue-mutation and decide refusals (REQ-SD-015/-016).
	factoryLaneBoundarySentinel = "lane boundary"
	// factoryNotALaneSentinel names the admission refusal of the lane verbs
	// (REQ-SD-015): next/stage/complete belong to a lane session only.
	factoryNotALaneSentinel = "not a lane session"
)

// factoryNotALaneError refuses a lane verb invoked outside a lane session.
func factoryNotALaneError(verb string) error {
	return fmt.Errorf("factory %s: refused — %s: set %s=%s in a lane session (the launcher stamps it)",
		verb, factoryNotALaneSentinel, config.EnvFactoryRole, config.FactoryRoleLane)
}

// factoryCodexMergeSentinel names the REQ-SD-025 refusal: while the backend
// variable identifies the Codex harness, the merge-ready → merging edge is
// refused on every path. One wording source, so `complete`, `stage`, and
// their MCP tools (M3) cannot drift apart.
const factoryCodexMergeSentinel = "the Codex harness cannot take the merge-ready → merging edge"

// factoryRefuseCodexMergeEdge is the REQ-SD-025 reusable check: it refuses
// the merge-ready → merging edge when the lane's backend variable identifies
// the Codex harness (gpt), and returns nil for every other backend. The CLI
// verbs call it before touching any record; the MCP factory tools (M3) call
// the same function.
func factoryRefuseCodexMergeEdge(verb string) error {
	if os.Getenv(config.EnvMoaiKanbanBackend) != kanban.BackendGPT {
		return nil
	}
	return fmt.Errorf("factory %s: refused — %s: a Codex lane stops at merge-ready; integration is the Claude lane's or the leader's (F3)",
		verb, factoryCodexMergeSentinel)
}

// factoryAssertParentCheckout refuses when dir is not the repository's
// parent (primary) checkout, naming the parent path (REQ-SD-010). The CLI
// path evaluates it against the command's project root; the MCP factory
// tools (M3) evaluate the same function against the caller's project_root
// argument, so both surfaces share one rule.
func factoryAssertParentCheckout(dir string) error {
	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}
	primary, _, err := identifyPrimaryCheckout(dir)
	if err != nil {
		return fmt.Errorf("factory next: cannot identify the parent checkout of %s: %w", dir, err)
	}
	if primary != dir {
		return fmt.Errorf("factory next: refused — this verb runs from the parent checkout %s, not from %s", primary, dir)
	}
	return nil
}

// The `next` wait contract (REQ-SD-009, plan B10): a fixed check interval,
// a fixed default bound, and one flag to change the bound.
const (
	factoryNextWaitInterval     = 5 * time.Second
	factoryNextWaitBoundDefault = 15 * time.Minute
)

// factoryNextWaitSleep is the seam the --wait loop sleeps through; tests
// replace it to drive the interval without sleeping.
var factoryNextWaitSleep = time.Sleep

// factoryNextSelectionAttempts bounds the retry loop around a selection that
// lost a lease race: each attempt re-reads the record and the queue, so a
// bounded number of attempts is enough to converge without spinning.
const factoryNextSelectionAttempts = 5

// factoryNextNoCardExit is the status `next` reports when no card qualifies
// (REQ-SD-008): 3, so a supervising launcher can distinguish it from failure.
const factoryNextNoCardExit = 3

// factoryNextLeaseOnce selects and leases one card for lane through the F1
// transition API (REQ-SD-008): a card assigned to this lane, then an
// operator-picked card assigned to no lane, then the oldest queued card
// (promoted to picked in the same operation). It never selects a card
// assigned to another lane. A race with another lane is retried inside; the
// returned bool reports whether a card was leased.
func factoryNextLeaseOnce(ctx context.Context, root, runID, lane string) (homestate.Card, bool, error) {
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return homestate.Card{}, false, fmt.Errorf("open factory record: %w", err)
	}
	defer func() { _ = db.Close() }()
	skip := factoryNextSkipForBackend()
	for attempt := 0; attempt < factoryNextSelectionAttempts; attempt++ {
		card, leased, raced, err := factoryNextSelectAndLease(ctx, db, root, runID, lane, skip)
		if err != nil {
			return homestate.Card{}, false, err
		}
		if leased {
			return card, true, nil
		}
		if !raced {
			return homestate.Card{}, false, nil
		}
	}
	return homestate.Card{}, false, nil
}

// factoryNextSelectAndLease runs one selection pass. raced reports that
// another lane moved the candidate first and the caller should re-select.
func factoryNextSelectAndLease(ctx context.Context, db *homestate.FactoryDB, root, runID, lane string, skip func(homestate.Card) bool) (homestate.Card, bool, bool, error) {
	cards, err := db.ListCards(ctx, runID)
	if err != nil {
		return homestate.Card{}, false, false, err
	}
	recorded := make(map[string]bool, len(cards))
	for _, c := range cards {
		recorded[c.CardID] = true
	}
	// (a) a card assigned to this lane — the lease edge alone (T3).
	for _, c := range cards {
		if c.State != homestate.CardAssigned || c.OwnerLabel != lane {
			continue
		}
		if skip(c) {
			continue
		}
		return factoryNextClaim(ctx, db, runID, c, lane)
	}
	// (b) an operator-picked card assigned to no lane. The record row sits at
	// `picked` with no owner; the queue item must still be picked, so an
	// unpicked queue item disqualifies the row rather than failing the verb.
	for _, c := range cards {
		if c.State != homestate.CardPicked || strings.TrimSpace(c.OwnerLabel) != "" {
			continue
		}
		if skip(c) {
			continue
		}
		state, inQueue, err := queueItemState(c.CardID)
		if err != nil {
			return homestate.Card{}, false, false, err
		}
		if !inQueue || state != kanban.BacklogStatePicked {
			continue
		}
		return factoryNextClaim(ctx, db, runID, c, lane)
	}
	// (b2) a queue-picked card with no record row yet: record it, then claim.
	rec, err := newTodoReadStore().LoadPure()
	if err != nil {
		return homestate.Card{}, false, false, err
	}
	for _, it := range rec.Items {
		if it.State != kanban.BacklogStatePicked || recorded[it.ID] {
			continue
		}
		return factoryNextRecordAndClaim(ctx, db, runID, it.ID, lane)
	}
	// (c) the oldest queued card: promote it to picked in the queue FIRST,
	// then record — a record-write failure leaves it a plain unowned picked
	// card the next `next` takes at arm (b) (design.md §3).
	var promoted string
	if err := newTodoStore().Mutate(func(r *kanban.BacklogRecord) error {
		for i := range r.Items {
			if r.Items[i].State == kanban.BacklogStateQueued {
				r.Items[i].State = kanban.BacklogStatePicked
				promoted = r.Items[i].ID
				return nil
			}
		}
		return nil
	}); err != nil {
		return homestate.Card{}, false, false, err
	}
	if promoted == "" {
		// Another lane promoted the oldest card between the read and the
		// write; re-select against the new state.
		return homestate.Card{}, false, true, nil
	}
	return factoryNextRecordAndClaim(ctx, db, runID, promoted, lane)
}

// factoryNextRecordAndClaim records a queue-picked card (T1) and claims it.
func factoryNextRecordAndClaim(ctx context.Context, db *homestate.FactoryDB, runID, cardID, lane string) (homestate.Card, bool, bool, error) {
	fresh, err := db.RecordPicked(ctx, runID, cardID, homestate.CardFields{}, "factory-next", factoryCardNow())
	if err != nil {
		return factoryNextClaimRefused(err)
	}
	return factoryNextClaim(ctx, db, runID, fresh, lane)
}

// factoryNextClaim takes a card from `picked` or `assigned` to `leased` for
// lane through the version-checked F1 edges (T2 then T3), with the lane's
// label as the lease holder.
func factoryNextClaim(ctx context.Context, db *homestate.FactoryDB, runID string, cur homestate.Card, lane string) (homestate.Card, bool, bool, error) {
	c := cur
	if c.State == homestate.CardPicked {
		next, err := db.Transition(ctx, homestate.TransitionRequest{
			RunID: runID, CardID: c.CardID, To: homestate.CardAssigned,
			ExpectedVersion: c.Version, Actor: "factory-next", Owner: lane, Now: factoryCardNow(),
		})
		if err != nil {
			return factoryNextClaimRefused(err)
		}
		c = next
	}
	leased, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: c.CardID, To: homestate.CardLeased,
		ExpectedVersion: c.Version, Actor: lane, Now: factoryCardNow(),
	})
	if err != nil {
		return factoryNextClaimRefused(err)
	}
	return leased, true, false, nil
}

// factoryNextClaimRefused maps a claim refusal: a stale version or a holder
// mismatch is a race to retry; anything else is a real error.
func factoryNextClaimRefused(err error) (homestate.Card, bool, bool, error) {
	if errors.Is(err, homestate.ErrStaleVersion) || errors.Is(err, homestate.ErrLeaseHolder) {
		return homestate.Card{}, false, true, nil
	}
	return homestate.Card{}, false, false, err
}

// factoryNextSkipForBackend is the REQ-SD-025 selection half: a Codex lane
// never selects a card whose recorded stage or state it cannot advance —
// `merge-ready` or later, including a card returned to `assigned` by lease
// expiry with its stage kept. Every other backend selects freely.
func factoryNextSkipForBackend() func(homestate.Card) bool {
	if os.Getenv(config.EnvMoaiKanbanBackend) != kanban.BackendGPT {
		return func(homestate.Card) bool { return false }
	}
	return func(c homestate.Card) bool {
		return cardStageAtOrAfterMergeReady(c.Stage) || cardStageAtOrAfterMergeReady(c.State)
	}
}

// cardStageAtOrAfterMergeReady reports whether s is a card stage or state at
// or past merge-ready in the F1 pipeline order.
func cardStageAtOrAfterMergeReady(s string) bool {
	switch s {
	case homestate.CardMergeReady, homestate.CardMerging, homestate.CardMergedLocal,
		homestate.CardPushed, homestate.CardCIGreen, homestate.CardDone:
		return true
	}
	return false
}

// newFactoryNextCommand — `moai factory next [--wait] [--wait-bound <d>]`
// (REQ-SD-008/-009/-010): a lane session's self-dispatch verb, run from the
// parent checkout.
func newFactoryNextCommand() *cobra.Command {
	var wait bool
	var waitBound time.Duration
	var run string
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Lease the lane's next card through the factory record (lane session, parent checkout)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("next")
			}
			lane := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLabel))
			if lane == "" {
				return fmt.Errorf("factory next: %s is empty — a lane session carries its lane label there", config.EnvMoaiKanbanLabel)
			}
			if err := factoryAssertParentCheckout(resolveProjectDir()); err != nil {
				return err
			}
			root := factoryCardRoot()
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			runID, err := resolveFactoryCardRun(ctx, root, run)
			if err != nil {
				return fmt.Errorf("factory next: %w", err)
			}
			deadline := factoryCardNow().Add(waitBound)
			for {
				card, leased, err := factoryNextLeaseOnce(ctx, root, runID, lane)
				if err != nil {
					return fmt.Errorf("factory next: %w", err)
				}
				if leased {
					return factoryNextPrint(cmd, card)
				}
				if !wait || !factoryCardNow().Before(deadline) {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "no card is available")
					return &exitCodeError{code: factoryNextNoCardExit, msg: "factory next: no card is available"}
				}
				factoryNextWaitSleep(factoryNextWaitInterval)
			}
		},
	}
	cmd.Flags().BoolVar(&wait, "wait", false,
		"Keep re-checking at a fixed interval until a card is leased or the wait bound elapses")
	cmd.Flags().DurationVar(&waitBound, "wait-bound", factoryNextWaitBoundDefault,
		"How long --wait re-checks before reporting no card")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// factoryNextPrint reports the leased card (REQ-SD-008, plan B9): its id,
// its stage, its worktree name, and its pull-request and landed state
// exactly as `moai todo pr` prints them — the pre-dispatch cross-check the
// leader performs, reported by the lane itself for a self-dispatched card.
func factoryNextPrint(cmd *cobra.Command, card homestate.Card) error {
	out := cmd.OutOrStdout()
	worktree := dash("")
	if p := strings.TrimSpace(card.WorktreePath); p != "" {
		worktree = filepath.Base(p)
	}
	_, _ = fmt.Fprintf(out, "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), worktree)
	rec, err := newTodoReadStore().LoadPure()
	if err != nil {
		return fmt.Errorf("factory next: read the queue for the pr line: %w", err)
	}
	rows := computeTodoPRRows(cmd.ErrOrStderr(), rec, card.CardID)
	writeTodoPRRows(out, rec, rows)
	return nil
}

// newFactoryStageCommand — `moai factory stage <card> <state>` (REQ-SD-012).
// M1 registers the admission refusal; the transition behavior lands with
// milestone M3 of this SPEC.
func newFactoryStageCommand() *cobra.Command {
	var run string
	cmd := &cobra.Command{
		Use:   "stage <card> <state>",
		Short: "Apply a card's next stage transition with its evidence (lane session)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("stage")
			}
			// REQ-SD-025: the merge-ready → merging edge is refused on every
			// path while the backend identifies the Codex harness. The rest of
			// the transition behavior lands with milestone M3.
			if len(args) >= 2 && args[1] == homestate.CardMerging {
				if err := factoryRefuseCodexMergeEdge("stage"); err != nil {
					return err
				}
			}
			return errors.New("factory stage: not yet implemented (SPEC-FACTORY-SELF-DISPATCH-001 M3)")
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// newFactoryCompleteCommand — `moai factory complete <card> [remeasure]`
// (REQ-SD-013/-023): a Claude-harness lane takes a merge-ready card through
// `merging` to `merged-local` by the F1 merge gate, using as integration
// branch the branch the integration window records. The optional positional
// names the lane's re-measure evidence file; when omitted, complete writes
// the merge record itself under .moai/reports/<card>/ (it names the merge
// commit — it records the merge identity, never a test-run claim). The
// window is NOT released here: the lane releases it as its next step.
func newFactoryCompleteCommand() *cobra.Command {
	var run string
	cmd := &cobra.Command{
		Use:   "complete <card> [remeasure]",
		Short: "Take a merge-ready card through merging to merged-local (lane session)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("complete")
			}
			if err := factoryRefuseCodexMergeEdge("complete"); err != nil {
				return err
			}
			lane := strings.TrimSpace(os.Getenv(config.EnvMoaiKanbanLabel))
			if lane == "" {
				return fmt.Errorf("factory complete: %s is empty — a lane session carries its lane label there", config.EnvMoaiKanbanLabel)
			}
			remeasure := ""
			if len(args) > 1 {
				remeasure = args[1]
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return factoryCompleteCard(ctx, cmd, args[0], remeasure, run, lane)
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// factoryCompleteCard is the complete body after the lane checks: hold the
// integration window (the same record acquire writes, never a shell-out),
// apply the REQ-SD-023 refusals, merge --no-ff inside the worktree that has
// the integration branch checked out, and record the F1 edges T14 then T16.
func factoryCompleteCard(ctx context.Context, cmd *cobra.Command, cardID, remeasure, run, lane string) error {
	root := factoryCardRoot()
	lockRoot := integrationLockRoot()
	// The same session identity acquire resolves: a window whose holder is
	// unresolvable can be neither taken nor re-taken, so an empty id is a
	// blocker to report, never a value to invent.
	sessionID := integrationSessionID("")
	if sessionID == "" {
		return fmt.Errorf("factory complete: cannot resolve this session's id; the integration window needs a holder address (acquire resolves it from the session environment)")
	}
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}

	// REQ-SD-023: the window phase. A window this session already holds
	// keeps ITS recorded branch — the branch the window records is the
	// integration branch — so a lane that pre-acquired with --branch is not
	// re-resolved underneath its own choice. A window held by another live
	// session refuses naming the holder; a free or stale window is resolved
	// exactly as acquire resolves it and taken over.
	lock, err := kanban.ReadIntegrationLock(lockRoot)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	heldByUs := lock.Held() && lock.SessionID == sessionID && lock.Branch != ""
	var branch, source string
	if heldByUs {
		branch, source = lock.Branch, lock.BranchSource
	} else {
		if lock.Held() && lock.SessionID != sessionID && !lock.Stale() {
			return fmt.Errorf("factory complete: refused — the integration window is held by %s (pid %d) since %s on %s; complete after the holder releases (moai integration status reads it)",
				factoryHolderLabel(lock), lock.PID, lock.AcquiredAt, lock.Branch)
		}
		branch, source = factoryResolveIntegrationBranch(root, card)
	}

	// REQ-SD-023 refusals — each fires before any record changes.
	// (1) The caller-source window: acquire fell back to the caller's own
	// tree, which for a lane is its card worktree — never an integration
	// branch. The remedy is acquire's --branch.
	if source == kanban.BranchSourceCaller {
		return fmt.Errorf("factory complete: refused — the integration window's branch %q is the caller's own tree (source %s); re-acquire with --branch <integration-target> (a card's own tree is not its integration branch)", branch, kanban.BranchSourceCaller)
	}
	// (2) A card's own branch never serves as its integration branch.
	cardBranch := factoryBranchOfWorktree(card.WorktreePath)
	windowTree := ""
	if heldByUs {
		windowTree = lock.Worktree
	}
	integTree := factoryWorktreeForBranchIn(factoryRepoDir(root, card), branch)
	if (cardBranch != "" && branch == cardBranch) || factorySameTree(windowTree, card.WorktreePath) || factorySameTree(integTree, card.WorktreePath) {
		return fmt.Errorf("factory complete: refused — the integration branch %q is the card's own branch (worktree %s): a card's own branch never serves as its integration branch", branch, card.WorktreePath)
	}
	// (3) The integration worktree must be provisioned: the only tree holding
	// the integration branch may not be the parent checkout (which never
	// changes branch), and no tree at all is the same refusal.
	primary, _, err := identifyPrimaryCheckout(root)
	if err != nil {
		return fmt.Errorf("factory complete: cannot identify the parent checkout of %s: %w", root, err)
	}
	if integTree == "" || factorySameTree(integTree, primary) {
		return fmt.Errorf("factory complete: refused — the integration worktree for %q is not provisioned: no tree holds it, or only the parent checkout %s does (the parent never changes branch; the leader provisions the integration worktree)", branch, primary)
	}

	// Hold the window as the lane: the same record acquire writes, resolved
	// the same way (the owner pid, never this process's). A window already
	// ours is not re-written — the recorded branch choice stands.
	if !heldByUs {
		ownerPID, _ := session.ResolveOwnerPID()
		replaced, err := kanban.AcquireIntegrationLock(lockRoot, kanban.IntegrationLock{
			SessionID:    sessionID,
			SessionName:  lane,
			PID:          ownerPID,
			PIDSource:    kanban.PIDSourceSessionOwner,
			Branch:       branch,
			BranchSource: source,
			Worktree:     integTree,
			Card:         card.CardID,
		}, false)
		if err != nil {
			return fmt.Errorf("factory complete: %w", err)
		}
		if replaced != nil {
			// Never silent, exactly like acquire: the next lane must be able
			// to say what was cleared.
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  displaced stale window of %s (pid %d), held since %s\n", factoryHolderLabel(replaced), replaced.PID, replaced.AcquiredAt)
		}
	}

	// T14 — merge-ready → merging: the lease holder's edge, so a lane that
	// does not hold this card's lease is refused by F1 verbatim.
	merging, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: card.CardID, To: homestate.CardMerging,
		ExpectedVersion: card.Version, Actor: lane, Now: factoryCardNow(),
	})
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	mergeSHA, err := factoryMergeNoFF(integTree, cardBranch, card.CardID, branch)
	if err != nil {
		// The card stays in `merging` — the honest state for a merge in
		// progress that failed; the lane resolves the tree (T15) or the lease
		// expiry moves it to blocked. The window stays held by this lane.
		return fmt.Errorf("factory complete: card %s is in merging; the merge failed: %w", card.CardID, err)
	}
	path := remeasure
	if path == "" {
		if path, err = factoryWriteMergeRecord(root, card.CardID, mergeSHA, branch, integTree); err != nil {
			return fmt.Errorf("factory complete: card %s is in merging; recording the merge evidence failed: %w", card.CardID, err)
		}
	}
	done, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: card.CardID, To: homestate.CardMergedLocal,
		ExpectedVersion: merging.Version, Actor: lane,
		MergeSHA: mergeSHA, RemeasurePath: path, IntegrationBranch: branch, Now: factoryCardNow(),
	})
	if err != nil {
		return fmt.Errorf("factory complete: card %s is in merging; the F1 merge gate refused: %w", card.CardID, err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s merge=%s branch=%s worktree=%s\n",
		done.CardID, done.State, done.MergeSHA, branch, integTree)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  the integration window is still held by this session — run moai integration release next")
	return nil
}

// factoryResolveIntegrationBranch mirrors acquire's branch resolution
// (resolveIntegrationTarget) without its $PWD legs: the configured git-flow
// develop branch decides; with none configured the caller's own tree decided
// the window, which complete refuses, so the branch is the caller's — taken
// from the card worktree, the lane's own tree, never from the process cwd.
func factoryResolveIntegrationBranch(root string, card homestate.Card) (string, string) {
	if branch := strings.TrimSpace(config.LoadGitFlowIntegrationConfig(root).DevelopBranch); branch != "" {
		return branch, kanban.BranchSourceConfig
	}
	return factoryBranchOfWorktree(card.WorktreePath), kanban.BranchSourceCaller
}

// factoryRepoDir names the repository directory the worktree lookup runs
// from: the card's own worktree when the record carries one, else the
// project root — both answer for the same repository the integration branch
// lives in.
func factoryRepoDir(root string, card homestate.Card) string {
	if p := strings.TrimSpace(card.WorktreePath); p != "" {
		return p
	}
	return root
}

// factoryBranchOfWorktree reports the branch checked out in dir, or "" when
// dir is empty or git cannot answer (best-effort, like currentBranch).
func factoryBranchOfWorktree(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// factoryWorktreeForBranchIn resolves the worktree holding branch, anchored
// at repoDir — never at the process cwd (plan B7: the recorded window names
// the branch; the tree holding it is looked up in the card's repository).
// It reuses worktreeForBranchFromList, the same parser acquire's resolution
// reads.
func factoryWorktreeForBranchIn(repoDir, branch string) string {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return worktreeForBranchFromList(string(out), branch)
}

// factorySameTree reports whether a and b name the same directory — by
// identity when both stat, else by resolved string equality. Empty never
// matches (an unset worktree is not every worktree).
func factorySameTree(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	sa, ea := os.Stat(a)
	sb, eb := os.Stat(b)
	if ea == nil && eb == nil {
		return os.SameFile(sa, sb)
	}
	if fa, err := filepath.EvalSymlinks(a); err == nil {
		a = fa
	}
	if fb, err := filepath.EvalSymlinks(b); err == nil {
		b = fb
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

// factoryHolderLabel mirrors kanban's holder label: the human-facing name a
// lane recognizes its queue position by, else the session id.
func factoryHolderLabel(lock *kanban.IntegrationLock) string {
	if lock == nil {
		return "unknown"
	}
	if lock.SessionName != "" {
		return lock.SessionName
	}
	if lock.SessionID != "" {
		return lock.SessionID
	}
	return "unknown"
}

// factoryMergeNoFF performs `git merge --no-ff` of the card branch inside
// the worktree holding the integration branch, and returns the resulting
// HEAD. A branch already merged answers "Already up to date" and leaves HEAD
// at the existing merge commit — the AC-SD-013 shape where the lane merged
// before running complete.
func factoryMergeNoFF(integTree, cardBranch, cardID, branch string) (string, error) {
	if cardBranch == "" {
		return "", fmt.Errorf("the card records no worktree, so its branch cannot be resolved")
	}
	merge := exec.Command("git", "merge", "--no-ff", "-m",
		fmt.Sprintf("Merge %s into %s (card %s, factory complete)", cardBranch, branch, cardID), cardBranch)
	merge.Dir = integTree
	if out, err := merge.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git merge in %s: %v: %s", integTree, err, strings.TrimSpace(string(out)))
	}
	rev := exec.Command("git", "rev-parse", "HEAD")
	rev.Dir = integTree
	out, err := rev.Output()
	if err != nil {
		return "", fmt.Errorf("read HEAD of %s: %v", integTree, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// factoryWriteMergeRecord writes the merge record complete records when the
// lane passed no re-measure file: it names the merge commit and the tree
// identity the F1 merge gate verifies. It records the merge identity only —
// a re-measure the lane ran is the lane's own file, passed as the positional.
func factoryWriteMergeRecord(root, cardID, mergeSHA, branch, integTree string) (string, error) {
	dir := filepath.Join(root, ".moai", "reports", cardID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "merge-record.txt")
	body := fmt.Sprintf("merge %s\nbranch %s\nintegration worktree %s\nrecorded by moai factory complete (card %s)\n",
		mergeSHA, branch, integTree, cardID)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

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
	// UnavailableSkipped counts unparseable unavailable-log lines the reader
	// skipped — reported as a warning, never a read failure (F2).
	UnavailableSkipped int `json:"unavailable_skipped,omitempty"`
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
			entries, skipped, err := homestate.ReadRecordUnavailable(root, report.Run)
			if err != nil {
				return fmt.Errorf("factory status: %w", err)
			}
			report.Unavailable = append([]homestate.RecordUnavailableEntry{}, entries...)
			report.UnavailableSkipped = skipped
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
	if r.UnavailableSkipped > 0 {
		_, _ = fmt.Fprintf(w, "%s warning: skipped %d unparseable line(s)\n", factoryRecordUnavailableTag, r.UnavailableSkipped)
	}
}

func newFactoryDecideCommand() *cobra.Command {
	var gate, choice, decider, run string
	cmd := &cobra.Command{
		Use:   "decide <card>...",
		Short: "Record an operator decision: --gate kickoff --choice approve|reject, --gate push, or --choice resume|block|unblock|abandon",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// REQ-SD-016: a session for which lane refusal holds — a lane by
			// marker, label, or Codex backend — cannot record decisions. The
			// guard runs before any card row is read or written.
			if factoryLaneRefusal() {
				return fmt.Errorf("factory decide: refused — %s: decide records the operator's decisions; a lane session cannot (on the Codex MCP path the same refusal covers factory_decide)", factoryLaneBoundarySentinel)
			}
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
