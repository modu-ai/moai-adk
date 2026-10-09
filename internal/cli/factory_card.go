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
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/worktree"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/session"
	"github.com/modu-ai/moai-adk/internal/statusline"
)

// factoryCardNow is the clock the factory card commands read; tests replace it
// to drive lease expiry without sleeping.
var factoryCardNow = time.Now

// factoryNominateBeforeRecord is the test seam of the nominated lease
// (SPEC-TODO-AUTO-PICK-001 plan N1/N2): the nomination path calls it after it
// has promoted the nominee to `picked` and before the first record write
// (RecordPicked), so a test can inject a claim failure or a competing lease at
// the one point the compensation can undo. The default is inert.
var factoryNominateBeforeRecord = func(cardID string) error { return nil }

// factoryLeaseBeforeClaim is the test seam of the bare selection arms
// (SPEC-FACTORY-ATOMIC-LEASE-001 plan WM1): every arm calls it immediately
// before its claim — arm names "a", "b", "b2" and "c", with the card the arm is
// about to claim — so a test can hold a lane between the arm's decision and its
// claim, or start an operator write there. An error it returns ends the pass
// like any other claim failure. The default is inert.
var factoryLeaseBeforeClaim = func(arm, cardID string) error { return nil }

// factoryLeaseAtPassEntry is the test seam at the entry of the lease section —
// a selection pass or a nominated lease — before it reads the clock, the record
// or the queue. The default is inert.
var factoryLeaseAtPassEntry = func() {}

// The lease path's claim bound and the worktree step's lock wait
// (SPEC-FACTORY-ATOMIC-LEASE-001 plan D2, D3). The claim wait cap is the sum of
// the claim's context deadline and the busy timeout its record connection
// carries; it must stay within one third of the queue lock's wait budget
// (factory.LockWaitBudget). The worktree step takes its lock for at most
// factoryWorktreeStepWait (factoryCreateAndRenameWorktree).
const (
	factoryLeaseClaimDeadline    = 800 * time.Millisecond
	factoryLeaseClaimBusyTimeout = 200 * time.Millisecond
	factoryLeaseClaimWaitCap     = factoryLeaseClaimDeadline + factoryLeaseClaimBusyTimeout

	// factoryWorktreeStepWaitDefault is the wait for the worktree-step lock:
	// lanes (10) x the worst observed step (2.9 s) x headroom (2), taken as 60 s.
	factoryWorktreeStepWaitDefault = 60 * time.Second
)

// factoryWorktreeStepWait is the wait factoryEnsureCardWorktree gives the
// worktree-step lock; a test shortens it and restores it at cleanup.
var factoryWorktreeStepWait = factoryWorktreeStepWaitDefault

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
//
// @MX:ANCHOR: [AUTO] Lane admission predicate — every lane-gated surface funnels through it.
// @MX:REASON: 7 call sites across factory_card.go and mcp_factory_card.go; a queue-writing verb that skips this predicate breaks the lane permission boundary (REQ-SD-015/-016/-017).
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func factoryLaneAdmission() bool {
	return os.Getenv(config.EnvFactoryRole) == config.FactoryRoleLane
}

// factoryLaneRefusal reports lane refusal: admission, or a non-empty
// lane-label variable, or the backend variable naming the Codex harness.
//
// @MX:ANCHOR: [AUTO] Lane refusal predicate — one wording source for the queue guard, the decide guard, and the todo guard.
// @MX:REASON: 4 call sites (factory_card.go, mcp_factory_card.go, mcp_todo.go, todo.go); widening one path alone would let a lane queue mutation slip past its siblings (REQ-SD-015).
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func factoryLaneRefusal() bool {
	return factoryLaneAdmission() ||
		os.Getenv(config.EnvMoaiFactoryWorker) != "" ||
		os.Getenv(config.EnvFactoryBackend) == factory.BackendGPT
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

// factoryLaneLabelFromEnv reads the lane-identity variable for a lane verb
// and refuses legacy spellings before any work is done (AC-SD-021): the
// refusal mirrors the REQ-RNC-003/-005/-007 producer shapes, naming the
// canonical lane label. Every lane verb reads its identity through this
// helper, so a hand-set legacy label cannot reach the factory record.
func factoryLaneLabelFromEnv(verb string) (string, error) {
	lane := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker))
	if lane == "" {
		return "", fmt.Errorf("factory %s: %s is empty — a lane session carries its lane label there", verb, config.EnvMoaiFactoryWorker)
	}
	lowered := strings.ToLower(lane)
	if n, ok := factory.SplitFactoryLegacyLabel(lowered); ok {
		return "", fmt.Errorf("factory %s: %q is the legacy lane label; use %q (rejoin with -l)", verb, lane, factory.FactoryLaneLabel(n))
	}
	if factory.IsLegacyFactoryRoleValue(lowered) {
		return "", fmt.Errorf("factory %s: %q is the legacy role token; lane sessions carry lane-<n> labels (rejoin with -l)", verb, lane)
	}
	if factory.IsLegacyLeaderSpelling(lowered) {
		return "", fmt.Errorf("factory %s: %q is the legacy leader spelling; lane sessions carry lane-<n> labels (the leader launches with -f)", verb, lane)
	}
	return lane, nil
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
	if os.Getenv(config.EnvFactoryBackend) != factory.BackendGPT {
		return nil
	}
	return fmt.Errorf("factory %s: refused — %s: a Codex lane stops at merge-ready; integration is the Claude lane's or the leader's (F3)",
		verb, factoryCodexMergeSentinel)
}

// sameDirPath reports whether two path spellings name the same directory.
// Beyond the literal form it accepts case-only differences on the platforms
// whose filesystems resolve case-insensitively (t1293): a lane launched from
// a lowercase logical PWD occupies the same physical checkout git names with
// its on-disk spelling, so spelling equality must not gate lane admission
// there. On case-sensitive platforms only the literal form passes — there a
// case-distinct path IS a different directory.
func sameDirPath(a, b string) bool {
	if a == b {
		return true
	}
	switch runtime.GOOS {
	case "darwin", "windows":
		return strings.EqualFold(a, b)
	default:
		return false
	}
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
	if !sameDirPath(primary, dir) {
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

// factorySerialSlotFree positively enumerates the card states that re-admit
// serial selection (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-008). The slot
// protects the ORDERING OF IMPLEMENTATION WORK: it is held from the moment a
// serial card is recorded until its implementation pipeline ends — every
// state at or after merge-ready (cardStageAtOrAfterMergeReady: the card's
// implementation is finished and what remains is the integration pipeline,
// which the integration window serializes on its own) plus failed and
// abandoned, the terminal exits no transition ever leaves (homestate
// IsTerminalCardState). A state added later keeps the slot held — the
// enumeration names the RELEASING states and never the holding ones (plan
// G2): a negative check (`state != done && ...`) would silently release the
// slot for every state added after it was written. The state is not the whole
// read: factorySerialSlotHeld also frees a lease-holding card whose lease has
// expired.
//
// merge-ready and later release the slot because that is the recorded
// behavior the absorbed self-dispatch suite pins: a Codex lane's relaunch
// loop leases its next card while the previous one sits at merge-ready
// awaiting integration (factory_m5_test.go) — a merge-ready card is not
// being worked, and blocking the fleet on it would cost the factory its
// throughput without protecting any ordering the integration window does not
// already protect.
func factorySerialSlotFree(state string) bool {
	switch state {
	case homestate.CardMergeReady, homestate.CardMerging, homestate.CardMergedLocal,
		homestate.CardPushed, homestate.CardCIGreen, homestate.CardDone, homestate.CardFailed,
		homestate.CardAbandoned:
		return true
	case homestate.CardPROpen, homestate.CardMergedPR: // github-flow delivery: the PR edge ended the lane's work on the card
		return true
	default:
		return false
	}
}

// factorySerialSlotHeld reports whether a recorded card holds the serial slot
// at now. Three conditions hold together: the state is not one of the
// releasing states; — for a card in a lease-holding state — its lease has not
// expired; and a driver is recorded. An expired lease is only collected
// lazily, by the next transition on that same card, so the row keeps its
// lease-holding state after the lane that held it is gone; reading the state
// alone would hold the slot for that lane indefinitely (card t1407).
//
// The driver condition (card t1513): a row whose OwnerLabel is empty holds
// nothing. A picked row recorded without a lane (`factory assign` with no
// --lane) names a nomination nobody is driving, and the lease-expiry net
// above never applies to it because it carries no lease — the slot would
// hold for as long as the row exists. Measured 2026-10-05: run tmf011's
// t1453 sat picked, ownerless, and lease-less for a day while every lane
// lease in the run was refused `serial-slot`. A later `factory assign
// --lane` or the nominate arm sets the owner and the row holds again, so
// only genuinely driverless rows release.
func factorySerialSlotHeld(c homestate.Card, now time.Time) bool {
	return !factorySerialSlotFree(c.State) && !c.LeaseExpired(now) && c.OwnerLabel != ""
}

// factorySerialInFlightExcluding reports whether a serial card OTHER than
// cardID sits in the record in a state that still holds the serial slot at
// now. The candidate's own row is excluded by identity: the exclusivity holds
// against DISTINCT cards (REQ-TCD-008). The unnominated arms and the nominated
// lease read the slot through this one function.
//
// ignoreAssigned is arm (a)'s read (operator ruling 2026-10-03, card t1407):
// a lane leasing the card assigned TO ITSELF does not count sibling cards
// that are merely `assigned`, which would otherwise wedge every
// leader-assigned serial card against the others with nothing in flight. The
// unnominated arm and — since card t1542 — the nominated lease of the lane's
// own assigned card pass true; every path that takes a NEW card passes
// false.
//
// A `picked` row carrying a bundle identity is excluded the same way (card
// t1454 card-review r2 P1-2): it is a chain member waiting on its head, not
// an independently picked serial card — its bundle orders it, and selection
// skips it until the predecessor merges. The sibling class (card t1533,
// card-review r2f finding 2): a `picked` row carrying an AFTER hint is
// chain-ordered work the same way — it waits on its predecessor's merge and
// nobody is implementing it — so it is excluded too, while a standalone
// driven picked row (no bundle identity, no hint) keeps holding the slot
// exactly as the t1407 ruling's tests pin.
//
// @MX:NOTE: [AUTO] The one serial-slot read of selection — fan-in 3 (both selection closures and the nominated validation); an ANCHOR is owed but factory_card.go is at its 3-anchor limit, and a queue-mutation path bypassing this read would break REQ-TCD-008.
func factorySerialInFlightExcluding(cards []homestate.Card, classOf func(string) factory.CardClassification, cardID string, now time.Time, ignoreAssigned bool) bool {
	for _, c := range cards {
		if c.CardID == cardID {
			continue
		}
		if ignoreAssigned && c.State == homestate.CardAssigned {
			continue
		}
		if c.State == homestate.CardPicked && (c.BundleID != "" || c.HintAfter != "") {
			continue
		}
		if factorySerialSlotHeld(c, now) && classOf(c.CardID).Mode == factory.ClassModeSerial {
			return true
		}
	}
	return false
}

// factoryQueueClassification reads one card's classification from a queue
// record snapshot. A card absent from the queue reads as the absent-field
// default derivation (REQ-TCD-014): serial, normal, non-blocked.
func factoryQueueClassification(rec *factory.BacklogRecord, cardID string) factory.CardClassification {
	if rec == nil {
		return factory.DefaultCardClassification()
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			return factory.EffectiveCardClassification(it)
		}
	}
	for _, entry := range rec.Archived {
		if entry.Item.ID == cardID {
			return factory.EffectiveCardClassification(entry.Item)
		}
	}
	return factory.DefaultCardClassification()
}

// factoryNextLeaseOnce selects and leases one card for lane through the F1
// transition API (REQ-SD-008): a card assigned to this lane, then an
// operator-picked card assigned to no lane, then the oldest queued card
// (promoted to picked in the same operation). It never selects a card
// assigned to another lane. A race with another lane is retried inside; the
// returned bool reports whether a card was leased.
func factoryNextLeaseOnce(ctx context.Context, root, runID, lane string) (homestate.Card, bool, error) {
	return factoryNextLeaseOnceGated(ctx, root, runID, lane, false)
}

// factoryNextLeaseOnceGated is factoryNextLeaseOnce with the quota gate's
// verdict handed in as one boolean (SPEC-QUOTA-AWARE-SCHEDULING-001 REQ-QAS-009):
// noNewCards leaves only the card already assigned to this lane leasable. The
// `next` verb and the factory_next MCP handler call this form; the relaunch
// loop and the Codex loop keep calling the ungated function above.
//
// @MX:NOTE: [AUTO] The gated lease entry point — the noNewCards boolean is the quota gate's whole effect on card selection; the ungated factoryNextLeaseOnce must stay for the relaunch and Codex loops. Fan-in 3 (the next verb, the factory_next MCP handler, the ungated wrapper) — kept a NOTE because factory_card.go is at its 3-anchor limit.
// @MX:SPEC: SPEC-QUOTA-AWARE-SCHEDULING-001
func factoryNextLeaseOnceGated(ctx context.Context, root, runID, lane string, noNewCards bool) (homestate.Card, bool, error) {
	// The record connection carries the claim's busy timeout and is opened
	// before the section, so a slow open never lengthens the lock's hold.
	db, err := homestate.OpenFactoryBounded(root, factoryLeaseClaimBusyTimeout)
	if err != nil {
		return homestate.Card{}, false, fmt.Errorf("open factory record: %w", err)
	}
	defer func() { _ = db.Close() }()
	store := todoStoreAt(root)
	skip := factoryNextSkipForBackend()
	for attempt := 0; attempt < factoryNextSelectionAttempts; attempt++ {
		var (
			card          homestate.Card
			leased, raced bool
			passErr       error
		)
		ran, lockErr := factoryLeaseSection(store, func(l *factory.LockedBacklog) {
			card, leased, raced, passErr = factoryNextSelectAndLease(ctx, l, db, root, runID, lane, skip, noNewCards)
		})
		if !ran {
			// A queue lock that could not be taken ends the pass at once with an
			// error: re-selecting would keep the verb waiting one more budget per
			// attempt and then report an empty queue it never read.
			return homestate.Card{}, false, factoryQueueLockError(lockErr)
		}
		if passErr != nil {
			return homestate.Card{}, false, passErr
		}
		if leased {
			// A failed lock release after a committed lease does not undo the lease;
			// reporting it would orphan the card, and the next queue write meets the
			// held lock loudly.
			return card, true, nil
		}
		if lockErr != nil {
			return homestate.Card{}, false, lockErr
		}
		if !raced {
			return homestate.Card{}, false, nil
		}
	}
	return homestate.Card{}, false, nil
}

// factoryLeaseSection runs fn inside the lease section: the queue's own
// cross-process lock held across the whole selection pass or nominated lease
// (SPEC-FACTORY-ATOMIC-LEASE-001), so the decision and the claim share one
// exclusion with every queue writer. ran reports whether fn was entered; when
// it is false err is the lock's acquisition failure, and when it is true err is
// at most the lock's release failure.
//
// @MX:WARN: [AUTO] the lease section holds the queue lock across a factory-record write, so its worst hold is the claim wait cap plus one promotion and one restore
// @MX:REASON: inside fn the queue is read and written only through the handle (a public Mutate waits out the whole budget on the section's own descriptor), and nothing here may start a git process or create a worktree (REQ-FAL-007)
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func factoryLeaseSection(store *factory.BacklogStore, fn func(*factory.LockedBacklog)) (ran bool, err error) {
	err = store.WithLock(func(l *factory.LockedBacklog) error {
		ran = true
		fn(l)
		return nil
	})
	return ran, err
}

// factoryQueueLockError names a queue lock the lease could not take.
func factoryQueueLockError(err error) error {
	if factory.IsStateLockHeld(err) {
		return fmt.Errorf("factory next: the queue lock stayed held for the whole %s wait budget, so nothing was leased: %w", factory.LockWaitBudget(), err)
	}
	return fmt.Errorf("factory next: take the queue lock: %w", err)
}

// errFactoryRecordBusy marks a claim that gave up because the factory record
// stayed busy for the claim's whole wait cap.
var errFactoryRecordBusy = errors.New("the factory record stayed busy")

// factoryClaimContext bounds one claim — RecordPicked and the two transitions
// taken together — by factoryLeaseClaimDeadline, and marks it so that none of
// its record writes waits for the drift log's lock (plan D2). The record
// connection's own busy timeout bounds how far past the deadline one write can
// run; a child of an already-bounded claim context cannot extend it.
func factoryClaimContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(homestate.WithBoundedReconcile(ctx), factoryLeaseClaimDeadline)
}

// factoryClaimFailure marks a claim write's error as the busy-record outcome
// when the claim's deadline has passed; a stale version or a holder mismatch is
// the lost race the caller already retries, and any other error stays as it is.
func factoryClaimFailure(claimCtx context.Context, err error) error {
	if errors.Is(err, homestate.ErrStaleVersion) || errors.Is(err, homestate.ErrLeaseHolder) {
		return err
	}
	if errors.Is(claimCtx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w for %s: %v", errFactoryRecordBusy, factoryLeaseClaimWaitCap, err)
	}
	return err
}

// factoryCardWorktreeDir is the card's landing directory under the shared
// materializer root — the leaf the REQ-SD-011 refusal and the carry-over both
// read.
func factoryCardWorktreeDir(root, cardID string) string {
	return filepath.Join(root, sessionWorktreeSubdir, cardID)
}

// factoryRefuseForeignWorktree is the REQ-SD-011 refusal: the card's landing
// directory already exists and no card record names it (the card itself
// records no worktree), so `next` refuses before any claim edge and the card
// row stays unchanged. The leaf name belongs to the card id, so an existing
// directory there can only be a foreign tree — never one the materializer
// created for this card. The carry-over (factoryCarryPreviousWorktree) runs
// before this refusal at the lease boundary: a directory a previous run of
// the SAME card recorded is that card's own tree, re-recorded rather than
// refused (card t1521); every other shape lands here.
func factoryRefuseForeignWorktree(root, cardID string) error {
	dir := factoryCardWorktreeDir(root, cardID)
	if info, err := os.Lstat(dir); err != nil || !info.IsDir() {
		return nil
	}
	return fmt.Errorf("factory next: refused — the worktree directory %s already exists and no card record names it; a card never adopts another card's tree (remove or rename the directory first)", dir)
}

// factoryCarryPreviousWorktree resolves the binding a run replacement may
// carry into runID for cardID (card t1521): the card's landing directory
// exists, a previous run's record of the SAME card names that same directory,
// and the tree it names still exists on disk. Only that exact shape carries —
// the new run re-records what its predecessor already bound instead of
// refusing the card's own surviving tree as foreign. It returns "" when the
// carry does not apply, leaving the REQ-SD-011 refusal (and the fresh-tree
// materialization when no directory stands in the way) to the caller.
func factoryCarryPreviousWorktree(ctx context.Context, db *homestate.FactoryDB, root, runID, cardID string) (string, error) {
	dir := factoryCardWorktreeDir(root, cardID)
	if info, err := os.Lstat(dir); err != nil || !info.IsDir() {
		return "", nil
	}
	prev, ok, err := db.PreviousCardWorktree(ctx, cardID, runID)
	if err != nil || !ok {
		return "", err
	}
	if info, err := os.Stat(prev); err != nil || !info.IsDir() {
		return "", nil
	}
	if !sameDirPath(prev, dir) {
		return "", nil
	}
	return prev, nil
}

// factoryWorktreeSlug derives the card worktree's WT- branch slug from the
// card's queue title (the Factory Dispatch Protocol branch-naming rule): lowercase
// [a-z0-9-], at most three tokens, at most 24 characters, and never
// containing the card id — a token carrying the id is dropped. A title with
// no usable token falls back to "card".
func factoryWorktreeSlug(cardID, title string) string {
	id := strings.ToLower(strings.TrimSpace(cardID))
	var tokens []string
	cur := &strings.Builder{}
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		tok := cur.String()
		cur.Reset()
		if id != "" && strings.Contains(tok, id) {
			return
		}
		tokens = append(tokens, tok)
	}
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	if len(tokens) > 3 {
		tokens = tokens[:3]
	}
	slug := strings.Join(tokens, "-")
	if len(slug) > 24 {
		slug = slug[:24]
	}
	if slug == "" {
		return "card"
	}
	return slug
}

// factoryCardQueueTitle reads the card's queue title (the text `todo add`
// recorded), anchored at root like every other queue read of the verb; a
// card absent from the queue derives the slug from no title.
func factoryCardQueueTitle(root, cardID string) string {
	rec, err := todoReadStoreAt(root).LoadPure()
	if err != nil {
		return ""
	}
	for _, it := range rec.Items {
		if it.ID == cardID {
			return it.Text
		}
	}
	for _, entry := range rec.Archived {
		if entry.Item.ID == cardID {
			return entry.Item.Text
		}
	}
	return ""
}

// factoryEnsureCardWorktree is the REQ-SD-011 record step of a successful
// lease: a card with a recorded worktree reuses that card's own tree; a card
// without one gains a worktree created through the shared worktree
// materializer (worktree.WorktreeCreator, wired to the session-worktree
// materializer — never a bare `git worktree add`), whose directory leaf is
// the card id and whose branch is renamed in place to WT-<slug> from the
// card's queue title, and records the created path on the card. It returns
// the card's worktree path and whether a new tree was created.
//
// @MX:ANCHOR: [AUTO] Per-card worktree materialization — the only path that creates a card worktree.
// @MX:REASON: 4 callers (factory_card, mcp_factory_card, factory_lane_relaunch, codex_launcher); a caller bypassing it would create a bare `git worktree add` outside the shared materializer (REQ-SD-011).
// @MX:SPEC: SPEC-FACTORY-SELF-DISPATCH-001
func factoryEnsureCardWorktree(ctx context.Context, root, runID string, card homestate.Card, lane string, out io.Writer) (string, bool, error) {
	if p := strings.TrimSpace(card.WorktreePath); p != "" {
		return p, false, nil
	}
	if worktree.WorktreeCreator == nil {
		return "", false, errors.New("worktree creator is not initialized")
	}
	slug := factoryWorktreeSlug(card.CardID, factoryCardQueueTitle(root, card.CardID))
	wt, err := factoryCreateAndRenameWorktree(root, card.CardID, slug, out)
	if err != nil {
		return "", false, err
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return "", false, fmt.Errorf("record the card worktree: %w", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.RecordCardWorktree(ctx, runID, card.CardID, wt, lane, factoryCardNow()); err != nil {
		return "", false, fmt.Errorf("record the card worktree: %w", err)
	}
	return wt, true, nil
}

// factoryCreateAndRenameWorktree runs the creator and the in-place branch
// rename under the worktree-step lock (SPEC-FACTORY-ATOMIC-LEASE-001 plan D3):
// two lanes creating a worktree at the same instant collided in git's own ref
// and worktree bookkeeping, so the step is serialized across processes. The
// lock is released before the caller's record write, and a wait that runs out
// fails the step with nothing created.
//
// @MX:NOTE: [AUTO] The lock covers the creator and `git branch -m` only; the record write stays outside it (D3, lock order).
// @MX:SPEC: SPEC-FACTORY-ATOMIC-LEASE-001
func factoryCreateAndRenameWorktree(root, cardID, slug string, out io.Writer) (string, error) {
	release, err := factory.AcquireFactoryStepLock(root, factoryWorktreeStepWait)
	if err != nil {
		return "", fmt.Errorf("serialize the card worktree step: %w", err)
	}
	defer func() { _ = release() }()
	wt, err := worktree.WorktreeCreator(cardID, out)
	if err != nil {
		return "", fmt.Errorf("create the card worktree: %w", err)
	}
	rename := exec.Command("git", "-C", wt, "branch", "-m", SessionWorktreeBranchPrefix+slug)
	if outb, err := rename.CombinedOutput(); err != nil {
		return "", fmt.Errorf("rename the card worktree branch: %v: %s", err, strings.TrimSpace(string(outb)))
	}
	return wt, nil
}

// factoryNextSelectAndLease runs one selection pass. raced reports that
// another lane moved the candidate first and the caller should re-select.
//
// Classification eligibility (SPEC-TODO-CLASSIFY-DISPATCH-001): one pure
// queue read anchors every mode lookup; a serial card recorded in a
// non-terminal state holds the serial slot — unless its lease has expired
// (factorySerialSlotHeld) — so no lane leases another serial
// card through ANY arm, except that arm (a), a lane leasing the card assigned
// to itself, does not count sibling cards that are merely assigned — while
// parallelizable candidates stay leasable
// throughout (REQ-TCD-008). The auto-promotion arm additionally never
// selects a blocked card (REQ-TCD-007).
//
// noNewCards is the quota gate's verdict (SPEC-QUOTA-AWARE-SCHEDULING-001
// REQ-QAS-009), evaluated once per `next` invocation outside the arms: it ends
// the pass after arm (a), so a card already assigned to this lane still leases
// while arms (b), (b2), and (c) — every arm that takes a NEW card — are
// bypassed without a second predicate.
//
// The pass runs inside the lease section (factoryLeaseSection): l is the handle
// of the queue lock the caller holds, every queue read and write goes through
// it, and the claim is made before the section ends.
func factoryNextSelectAndLease(ctx context.Context, l *factory.LockedBacklog, db *homestate.FactoryDB, root, runID, lane string, skip func(homestate.Card) bool, noNewCards bool) (homestate.Card, bool, bool, error) {
	factoryLeaseAtPassEntry()
	// The clock is read BEFORE the record snapshot: a lease whose expiry the
	// snapshot shows as already past was expired when the clock was read too,
	// so a renewal landing after the snapshot cannot be read as an expiry. Read
	// afterwards, a renewal slipping between the two would free the slot for a
	// lease that is in fact live and let two serial cards run at once.
	now := factoryCardNow()
	cards, err := db.ListCards(ctx, runID)
	if err != nil {
		return homestate.Card{}, false, false, err
	}
	recorded := make(map[string]bool, len(cards))
	for _, c := range cards {
		recorded[c.CardID] = true
	}
	// One pure queue read for the classification snapshot. A failed read is
	// not silently swallowed: selection refuses rather than guessing modes
	// (a nil record would read every card as its default — silent, wrong).
	queueRec, err := l.LoadPure()
	if err != nil {
		return homestate.Card{}, false, false, fmt.Errorf("read the queue for classification: %w", err)
	}
	classOf := func(cardID string) factory.CardClassification {
		return factoryQueueClassification(queueRec, cardID)
	}
	// serialInFlightExcluding reports whether a serial card OTHER than
	// cardID sits in the record in a non-terminal state. The candidate's own
	// row is excluded by identity: a lane re-leasing ITS OWN assigned/picked
	// serial card is not a second serial card in flight — the exclusivity
	// holds against DISTINCT cards (REQ-TCD-008), and a self-blocked
	// candidate would wedge every lease of a legacy serial row.
	//
	// ignoreAssigned is arm (a)'s read (operator ruling 2026-10-03, card
	// t1407); see factorySerialInFlightExcluding. Every arm that takes a NEW
	// card passes false.
	serialInFlightExcluding := func(cardID string, ignoreAssigned bool) bool {
		return factorySerialInFlightExcluding(cards, classOf, cardID, now, ignoreAssigned)
	}
	modeEligible := func(cardID string) bool {
		if classOf(cardID).Mode != factory.ClassModeSerial {
			return true
		}
		return !serialInFlightExcluding(cardID, false)
	}
	// Bundle attributes of the recorded rows (REQ-TCI-018): the bundle's
	// lane is the owner recorded on any of its members (the loader assigns
	// the head), and a member whose after predecessor has not reached the
	// local merge is SKIPPED by selection rather than handed to the claim to
	// fail — the skip-not-error hook of design §7.2. The same skip covers a
	// hub-chained card (REQ-TCI-020): its hint waits the same way, so no
	// lane wedges on a lease its predecessor has not earned yet.
	//
	// mergedLocal reads every run (card t1454 card-review r2 finding 5):
	// predecessorMerged — the T2 guard the hint claims against — reads every
	// run, so the selection's skip must read the same set; a predecessor
	// merged under a previous run id would otherwise wedge its successor
	// forever.
	bundleLane := make(map[string]string, len(cards))
	rowByID := make(map[string]homestate.Card, len(cards))
	for _, c := range cards {
		rowByID[c.CardID] = c
		if c.BundleID != "" && strings.TrimSpace(c.OwnerLabel) != "" {
			if _, seen := bundleLane[c.BundleID]; !seen {
				bundleLane[c.BundleID] = strings.TrimSpace(c.OwnerLabel)
			}
		}
	}
	allRuns, err := db.ListCards(ctx, "")
	if err != nil {
		return homestate.Card{}, false, false, err
	}
	mergedLocal := factoryMergedCards(allRuns)
	// selectionSkips reports why a recorded candidate must not lease on this
	// pass: it is another lane's bundle member, its predecessor is
	// unmerged, or an open sharer of one of its hub paths is unmerged —
	// the stored hint names one predecessor, a multi-hub candidate has one
	// per hub path (card t1533, card-review r2f finding 4). The direct
	// nominated path keeps the T2 error — the skip is the un-nominated
	// selection's shape alone.
	hubWaitUnmerged := func(cardID string) (string, bool) {
		return factoryHubWaitUnmerged(queueRec, cards, mergedLocal, cardID)
	}
	selectionSkips := func(c homestate.Card) bool {
		if c.BundleID != "" {
			if owner, ok := bundleLane[c.BundleID]; ok && owner != lane {
				return true
			}
		}
		if c.HintAfter != "" && !mergedLocal[c.HintAfter] {
			return true
		}
		_, wait := hubWaitUnmerged(c.CardID)
		return wait
	}
	// hubFields computes the hub-chain hint a record CREATION carries for
	// cardID, from the same one queue read every arm sees (REQ-TCI-020).
	hubFields := func(cardID string) homestate.CardFields {
		return factoryHubChainFields(queueRec, cards, cardID, nil)
	}
	// (a) a card assigned to this lane — the lease edge alone (T3).
	for _, c := range cards {
		if c.State != homestate.CardAssigned || c.OwnerLabel != lane {
			continue
		}
		if skip(c) {
			continue
		}
		if selectionSkips(c) {
			continue
		}
		// The queue item's CURRENT state gates the row's lease edge through
		// the same keep-set predicate the nominated path applies (card t1577
		// card-review; the t1516 gate this replaces covered hold and queued
		// only): an operator exclusion — dropped, held, hold-marked, blocked,
		// or back at queued (card t1516) — skips the assigned row. The row
		// re-enters a lease only once the item is picked again — an operator
		// pick, or arm (c)'s promotion on this same pass, which flips the
		// state before the next attempt re-reads it. A card in no queue row is
		// not gated: the record row is all that remains of it. The serial-slot
		// clause is arm (a)'s own check below (a self-excluding read the
		// predicate's shared form does not carry).
		if it, inQueue := queueItemIn(queueRec, c.CardID); inQueue {
			if r := factoryKeepSetRefusal(it, &c, lane, false); r != nil {
				continue
			}
		}
		if classOf(c.CardID).Mode == factory.ClassModeSerial && serialInFlightExcluding(c.CardID, true) {
			continue
		}
		if err := factoryLeaseBeforeClaim("a", c.CardID); err != nil {
			return homestate.Card{}, false, false, err
		}
		return factoryNextClaim(ctx, db, root, runID, c, lane)
	}
	if noNewCards {
		return homestate.Card{}, false, false, nil
	}
	// (b-priority) this lane's own bundle's next member outranks unowned
	// picked cards (REQ-TCI-018): the bundle lane serves its chain before
	// anything else an unowned picked row could offer. The conditions are
	// arm (b)'s, tightened by the bundle ownership.
	for _, c := range cards {
		if c.State != homestate.CardPicked || strings.TrimSpace(c.OwnerLabel) != "" || c.BundleID == "" {
			continue
		}
		if owner, ok := bundleLane[c.BundleID]; !ok || owner != lane {
			continue
		}
		if skip(c) || !modeEligible(c.CardID) || selectionSkips(c) {
			continue
		}
		state, inQueue := queueItemStateIn(queueRec, c.CardID)
		if !inQueue || state != factory.BacklogStatePicked {
			continue
		}
		if err := factoryLeaseBeforeClaim("b-priority", c.CardID); err != nil {
			return homestate.Card{}, false, false, err
		}
		return factoryNextClaim(ctx, db, root, runID, c, lane)
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
		if !modeEligible(c.CardID) {
			continue
		}
		if selectionSkips(c) {
			continue
		}
		state, inQueue := queueItemStateIn(queueRec, c.CardID)
		if !inQueue || state != factory.BacklogStatePicked {
			continue
		}
		if err := factoryLeaseBeforeClaim("b", c.CardID); err != nil {
			return homestate.Card{}, false, false, err
		}
		return factoryNextClaim(ctx, db, root, runID, c, lane)
	}
	// (b2) a queue-picked card with no record row yet: record it, then claim.
	// The queue cannot have changed since the pass's one read: the section holds
	// its lock, so arm (b2) reads the same record.
	for _, it := range queueRec.Items {
		// Positive enumeration (SPEC-TODO-HOLD-STATE-001 REQ-THS-012): the
		// state this arm claims is named; every other state falls through.
		if it.State == factory.BacklogStatePicked && !recorded[it.ID] {
			if !modeEligible(it.ID) {
				continue
			}
			// The no-record arm waits like every other path (card t1533,
			// review-gate r6/r7): recording first and refusing at the claim
			// errored the whole next call while unrelated ready cards waited
			// behind it. Skipped here, the card is neither recorded nor
			// claimed, and selection reaches the ready cards.
			if _, wait := hubWaitUnmerged(it.ID); wait {
				continue
			}
			if err := factoryLeaseBeforeClaim("b2", it.ID); err != nil {
				return homestate.Card{}, false, false, err
			}
			return factoryNextRecordAndClaim(ctx, db, root, runID, it.ID, lane, hubFields(it.ID))
		}
	}
	// (c) the highest-ranked eligible queued card (REQ-TCD-007/-008). The
	// queue is kept sorted by classification inside the add's locked write
	// (REQ-TCD-005), so stored order IS priority order — selection reads it,
	// it never re-sorts (plan G5). A blocked card is never auto-selected (an
	// operator pick or unblock is the only path that dispatches one); a
	// serial card is skipped while another serial card is in flight.
	var promoted string
	sawQueued := 0
	if err := l.Mutate(func(r *factory.BacklogRecord) error {
		for i := range r.Items {
			it := &r.Items[i]
			// Positive enumeration (REQ-THS-012): the state this arm promotes
			// is named; every other state — a state added later included —
			// falls through.
			if it.State == factory.BacklogStateQueued {
				sawQueued++
				// The hold marker keeps a card out of the lease path, and the
				// skipped card still counts as seen, so a queue of marker cards
				// alone ends on the no-card answer (SPEC-TODO-AUTO-PICK-001
				// REQ-TAU-007; the nominated lease applies the same predicate).
				if factoryQueuedHoldMarked(*it) {
					continue
				}
				cls := factory.EffectiveCardClassification(*it)
				if cls.Blocked {
					continue
				}
				// Another lane's bundle member, a candidate whose after
				// predecessor is unmerged, or a candidate with an unmerged
				// sharer on ANY of its hub paths is never auto-promoted
				// (REQ-TCI-018/-020): the record row, when one exists, says
				// which. A rowless candidate's hub hint is computed at
				// promotion, so its predecessor condition is checked HERE —
				// the same predicate selectionSkips applies to rows
				// (card t1454 card-review r2 finding 4, multi-hub sweep card
				// t1533). Promoting a candidate whose hint names an unmerged
				// predecessor failed the claim and errored the whole verb.
				if row, ok := rowByID[it.ID]; ok {
					if selectionSkips(row) {
						continue
					}
				} else if hf := hubFields(it.ID); hf.HintAfter != nil && !mergedLocal[*hf.HintAfter] {
					continue
				} else if _, wait := hubWaitUnmerged(it.ID); wait {
					continue
				}
				if cls.Mode == factory.ClassModeSerial && serialInFlightExcluding(it.ID, false) {
					continue
				}
				it.State = factory.BacklogStatePicked
				promoted = it.ID
				return nil
			}
		}
		return nil
	}); err != nil {
		return homestate.Card{}, false, false, err
	}
	switch {
	case promoted != "":
		if err := factoryLeaseBeforeClaim("c", promoted); err != nil {
			return homestate.Card{}, false, false, err
		}
		// Generated hub hints are creation inputs; preserve every existing row's hint.
		var promotedRow *homestate.Card
		if row, exists := rowByID[promoted]; exists {
			promotedRow = &row
		}
		fields := factoryGeneratedHubFields(hubFields(promoted), promotedRow)
		if promotedRow != nil && promotedRow.State == homestate.CardAssigned {
			// Reuse the assignment unchanged, then re-select through arm (a).
			// Hub hints are creation inputs, not edits to an assigned row.
			fields = homestate.CardFields{}
		}
		return factoryNextRecordAndClaim(ctx, db, root, runID, promoted, lane, fields)
	case sawQueued > 0:
		// Queued cards existed but none was eligible (blocked, or serial with
		// the slot held). That is the no-card answer, not a race: re-selecting
		// would spin on the same ineligible candidates.
		return homestate.Card{}, false, false, nil
	default:
		// No queued card at all — another lane promoted the one the read saw
		// between the read and the write; re-select against the new state.
		return homestate.Card{}, false, true, nil
	}
}

// factoryNextRecordAndClaim records a queue-picked card (T1) and claims it.
// fields carries the record-creation attributes the caller computed — the
// hub-chain after hint (REQ-TCI-020); the unnominated arms pass the hub
// read's answer, the nominated path the one computed inside its section.
func factoryNextRecordAndClaim(ctx context.Context, db *homestate.FactoryDB, root, runID, cardID, lane string, fields homestate.CardFields) (homestate.Card, bool, bool, error) {
	ctx, cancel := factoryClaimContext(ctx)
	defer cancel()
	fresh, err := db.RecordPicked(ctx, runID, cardID, fields, "factory-next", factoryCardNow())
	if err != nil {
		return factoryNextClaimRefused(factoryClaimFailure(ctx, err))
	}
	// RecordPicked returns the existing row unchanged when fields are empty —
	// whatever state it is in (card_picked.go). Both callers are queue-picked
	// arms, so a non-picked row means another lane recorded and advanced the
	// card between the ListCards snapshot and the queue read: a race to
	// re-select, not a fresh picked row to claim.
	if fresh.State != homestate.CardPicked {
		return homestate.Card{}, false, true, nil
	}
	return factoryNextClaim(ctx, db, root, runID, fresh, lane)
}

// factoryNextClaim takes a card from `picked` or `assigned` to `leased` for
// lane through the version-checked F1 edges (T2 then T3), with the lane's
// label as the lease holder. The REQ-SD-011 worktree refusal fires before
// the edges: a card with no recorded worktree whose landing directory is
// already taken by a foreign tree is refused, and no row changes. Before the
// refusal, the run-replacement carry-over re-records a landing directory a
// previous run of the SAME card bound (card t1521) — the card's own tree, so
// the lease proceeds onto it instead of refusing it.
func factoryNextClaim(ctx context.Context, db *homestate.FactoryDB, root, runID string, cur homestate.Card, lane string) (homestate.Card, bool, bool, error) {
	ctx, cancel := factoryClaimContext(ctx)
	defer cancel()
	c := cur
	if strings.TrimSpace(c.WorktreePath) == "" {
		carried, err := factoryCarryPreviousWorktree(ctx, db, root, runID, c.CardID)
		if err != nil {
			return factoryNextClaimRefused(factoryClaimFailure(ctx, err))
		}
		if carried != "" {
			next, rerr := db.RecordCardWorktree(ctx, runID, c.CardID, carried, "factory-next", factoryCardNow())
			if rerr != nil {
				return factoryNextClaimRefused(factoryClaimFailure(ctx, rerr))
			}
			c = next
		} else if err := factoryRefuseForeignWorktree(root, c.CardID); err != nil {
			return homestate.Card{}, false, false, err
		}
	}
	// T2 and T3 each re-point the factory binding onto this run in their own
	// transaction, so the queue's current-dispatch record is written inside
	// each, right before its commit (turn-end gate relay #4, after review
	// round-24 V2; the in-transaction form after the turn-end gate that
	// followed): the claim runs inside the lease section, which holds the queue
	// lock on this same store. A record that cannot be written rolls that
	// transition back — no lease, no binding move — and a refused transition
	// never moves the record. Recording after the transition would leave a
	// committed lease and binding ahead of a stale record, where a retried
	// older dispatch reads the record as current and drags the binding back.
	// The owner is the row the transition writes — the lane once T2 names it.
	followRecord := func(next homestate.Card) error {
		return todoStoreAt(root).RefreshDispatchCurrentLockHeld(next.CardID, runID, next.OwnerLabel)
	}
	if c.State == homestate.CardPicked {
		next, err := db.Transition(ctx, homestate.TransitionRequest{
			RunID: runID, CardID: c.CardID, To: homestate.CardAssigned,
			ExpectedVersion: c.Version, Actor: "factory-next", Owner: lane, Now: factoryCardNow(),
			BeforeCommit: followRecord,
		})
		if err != nil {
			return factoryNextClaimRefused(factoryClaimFailure(ctx, err))
		}
		c = next
	}
	leased, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: c.CardID, To: homestate.CardLeased,
		ExpectedVersion: c.Version, Actor: lane, Now: factoryCardNow(),
		BeforeCommit: followRecord,
	})
	if err != nil {
		return factoryNextClaimRefused(factoryClaimFailure(ctx, err))
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

// factoryNextRefusedExit is the status `next --card` reports when the nominee
// is refused (SPEC-TODO-AUTO-PICK-001 REQ-TAU-005): 4, distinct from the
// no-card status (3) and from an ordinary failure (1), so a session that
// nominated can tell "re-select" from "nothing to take" from "something broke".
const factoryNextRefusedExit = 4

// The closed set of refusal tokens of the nominated lease (spec § C.2). Every
// refusal of the nominee carries exactly one; an infrastructure failure
// carries none.
const (
	factoryRefuseUnknownCard    = "unknown-card"
	factoryRefuseDropped        = "dropped"
	factoryRefuseHeld           = "held"
	factoryRefuseOwned          = "owned"
	factoryRefuseRecorded       = "recorded"
	factoryTokenForeignWorktree = "foreign-worktree"
	factoryRefuseHoldMarker     = "hold-marker"
	factoryRefuseBlocked        = "blocked"
	factoryRefuseSerialSlot     = "serial-slot"
	factoryRefuseQuotaHold      = "quota-hold"
	factoryRefuseBackendSkip    = "backend-skip"
	factoryRefuseRaced          = "raced"
)

// factoryNominateRefusal is the refusal of a nominated card: a token from the
// closed set plus a one-line detail. It carries the refusal status itself, so
// both surfaces (the cobra verb and the MCP tool) report it identically.
type factoryNominateRefusal struct {
	Token  string
	Detail string
}

// Error renders the one-line form both surfaces print.
func (r *factoryNominateRefusal) Error() string {
	return "factory next: refused " + r.Token + ": " + r.Detail
}

// ExitCode satisfies ExitCoder: a refusal exits with factoryNextRefusedExit.
func (r *factoryNominateRefusal) ExitCode() int { return factoryNextRefusedExit }

// waitable reports whether `--wait` keeps waiting through the refusal: the
// three states a bare `--wait` also waits through (another lane's claim, a
// held serial slot, a quota hold). Any other refusal is permanent.
func (r *factoryNominateRefusal) waitable() bool {
	switch r.Token {
	case factoryRefuseRaced, factoryRefuseSerialSlot, factoryRefuseQuotaHold:
		return true
	}
	return false
}

func factoryRefusal(token, format string, args ...any) *factoryNominateRefusal {
	return &factoryNominateRefusal{Token: token, Detail: fmt.Sprintf(format, args...)}
}

// factoryQueuedHoldMarked reports whether a QUEUED card's text, trimmed, opens
// with the hold marker — the one predicate the nominated lease and the
// unnominated promotion arm share (the serial cycle demotes the same cards
// through autoRankHoldMarked). A card in any other queue state is not a
// marker card: the marker parks a card the lease would otherwise promote.
func factoryQueuedHoldMarked(it factory.BacklogItem) bool {
	return it.State == factory.BacklogStateQueued && autoRankHoldMarked(it.Text)
}

// factoryRecordRefusal maps a factory-record row to the refusal it earns for
// a lane that nominates the card (spec § C.2), nil when the row is leasable by
// that lane: a row at `picked` with no owner (an operator pick) or a row
// `assigned` to this lane (arm (a)). Every in-flight state is `owned` for any
// holder, this lane included; the terminal and parked states — and any value
// outside the nineteen — are `recorded`.
func factoryRecordRefusal(row homestate.Card, lane string) *factoryNominateRefusal {
	switch row.State {
	case homestate.CardPicked:
		if strings.TrimSpace(row.OwnerLabel) == "" {
			return nil
		}
		return factoryRefusal(factoryRefuseOwned, "the factory record holds the card at picked for %s", row.OwnerLabel)
	case homestate.CardAssigned:
		if row.OwnerLabel == lane {
			return nil
		}
		return factoryRefusal(factoryRefuseOwned, "the factory record assigns the card to %s", dash(row.OwnerLabel))
	case homestate.CardLeased, homestate.CardPlan, homestate.CardPlanAudit, homestate.CardKickoff, homestate.CardRun,
		homestate.CardSync, homestate.CardSyncAudit, homestate.CardMergeReady, homestate.CardMerging,
		homestate.CardMergedLocal, homestate.CardPushed, homestate.CardCIGreen:
		return factoryRefusal(factoryRefuseOwned, "the card is in flight at %s (holder %s)", row.State, dash(factoryRowHolder(row)))
	default:
		return factoryRefusal(factoryRefuseRecorded, "the factory record has the card at %s; only an operator unblock or re-pick moves it", row.State)
	}
}

// factoryRowHolder names who holds a record row: the lease holder, else the owner.
func factoryRowHolder(row homestate.Card) string {
	if h := strings.TrimSpace(row.LeaseHolder); h != "" {
		return h
	}
	return strings.TrimSpace(row.OwnerLabel)
}

// factoryKeepSetRefusal is the one keep-set predicate of the nominated lease
// (SPEC-TODO-AUTO-PICK-001 REQ-TAU-009): it returns the § C.2 refusal for a
// card the lease must not take, nil for a card that passes. It reads only the
// queue item (state, text marker, classification), the factory-record row,
// and the serial slot — never a relation store (REQ-TAU-013). The same
// function runs at validation and again inside the queue lock at promotion.
// The unnominated promotion arm applies only its marker clause, through
// factoryQueuedHoldMarked.
//
// @MX:NOTE: [AUTO] The shared keep-set predicate of the nominated lease; positive enumeration of the leasable queue states, so a state added later is refused, never leased. Fan-in 2 (validation and the in-lock re-validation) — a NOTE because factory_card.go is at its 3-anchor limit.
// @MX:SPEC: SPEC-TODO-AUTO-PICK-001
func factoryKeepSetRefusal(it factory.BacklogItem, row *homestate.Card, lane string, serialHeld bool) *factoryNominateRefusal {
	switch it.State {
	case factory.BacklogStateQueued, factory.BacklogStatePicked:
	case factory.BacklogStateDropped:
		return factoryRefusal(factoryRefuseDropped, "the card was dropped from the queue")
	case factory.BacklogStateHold:
		return factoryRefusal(factoryRefuseHeld, "the card is held (moai gtd hold); only the operator releases it")
	default:
		return factoryRefusal(factoryRefuseHeld, "the card's queue state is %s, which no lease path takes", it.State)
	}
	if row != nil {
		if r := factoryRecordRefusal(*row, lane); r != nil {
			return r
		}
	}
	// An assigned row whose queue item is back at queued is a shape no lease
	// path takes (card t1516): the item is promotable pool material again, so
	// the nomination does not re-arm the row — an operator pick re-arms it, and
	// the bare path reaches the card only through arm (c)'s promotion, which
	// flips the state to picked first. `hold` is refused above; this clause
	// names `queued` alone.
	if row != nil && row.State == homestate.CardAssigned && it.State == factory.BacklogStateQueued {
		return factoryRefusal(factoryRefuseRecorded, "the record row sits assigned while the queue item is queued; an operator pick re-arms the lease")
	}
	if factoryQueuedHoldMarked(it) {
		return factoryRefusal(factoryRefuseHoldMarker, "the card's text opens with the hold marker %s; the operator parked it", autoRankHoldMarker)
	}
	if factory.EffectiveCardClassification(it).Blocked {
		return factoryRefusal(factoryRefuseBlocked, "the card's classification is blocked; the nominated lease refuses it, the operator decides it")
	}
	if serialHeld {
		return factoryRefusal(factoryRefuseSerialSlot, "another serial card is in flight and holds the serial slot")
	}
	return nil
}

// factoryNominee is what the read-only validation read about the nominee.
type factoryNominee struct {
	item       factory.BacklogItem
	row        *homestate.Card // nil when the factory record has no row for the card
	serialHeld bool
}

// factoryNextValidate is step 1 of the nominated lease: one pure queue read
// through the section's locked handle, one factory-record read, and the
// read-only foreign-worktree precheck decide every § C.2 token that can be
// decided before a write. It writes nothing.
func factoryNextValidate(ctx context.Context, l *factory.LockedBacklog, db *homestate.FactoryDB, root, runID, lane, cardID, quotaHold string) (factoryNominee, *factoryNominateRefusal, error) {
	var nom factoryNominee
	queueRec, err := l.LoadPure()
	if err != nil {
		return nom, nil, fmt.Errorf("read the queue: %w", err)
	}
	found := false
	for _, it := range queueRec.Items {
		if it.ID == cardID {
			nom.item, found = it, true
			break
		}
	}
	if !found {
		return nom, factoryRefusal(factoryRefuseUnknownCard, "%s is in no live queue row", cardID), nil
	}
	// The clock is read before the record snapshot (see factoryNextSelectAndLease).
	now := factoryCardNow()
	cards, err := db.ListCards(ctx, runID)
	if err != nil {
		return nom, nil, err
	}
	for i := range cards {
		if cards[i].CardID == cardID {
			row := cards[i]
			nom.row = &row
			break
		}
	}
	classOf := func(id string) factory.CardClassification { return factoryQueueClassification(queueRec, id) }
	// Arm (a) through the nominated door (card t1542): a lane nominating the
	// card assigned TO ITSELF takes the same read the unnominated arm takes —
	// sibling rows that are merely `assigned` hold nothing against it, or the
	// leader's own batch pre-assignment would wedge every assigned card out
	// of its lane (measured 2026-10-07: nine assigned rows in run tmhxo0
	// refused the one lease a lane was dispatched to run). A nomination of a
	// card NOT assigned to this lane is a NEW card and still passes false.
	assignedHere := nom.row != nil && nom.row.State == homestate.CardAssigned && nom.row.OwnerLabel == lane
	nom.serialHeld = classOf(cardID).Mode == factory.ClassModeSerial && factorySerialInFlightExcluding(cards, classOf, cardID, now, assignedHere)
	if r := factoryKeepSetRefusal(nom.item, nom.row, lane, nom.serialHeld); r != nil {
		return nom, r, nil
	}
	// Another lane's bundle member is never a nominee (card t1454
	// card-review r2c finding C2). The unnominated selection skips it; the
	// nominated path refused nothing once the head had merged — the member's
	// own after guard passed and lane-2 leased lane-1's member outright. The
	// bundle's lane is the owner recorded on any of its members, the same
	// read the selection's bundleLane map makes.
	if nom.row != nil && nom.row.BundleID != "" {
		for i := range cards {
			c := cards[i]
			owner := strings.TrimSpace(c.OwnerLabel)
			if c.BundleID != nom.row.BundleID || owner == "" {
				continue
			}
			if owner != lane {
				return nom, factoryRefusal(factoryRefuseOwned, "the card is %s's bundle member", owner), nil
			}
			break
		}
	}
	// The multi-hub wait the un-nominated selection applies runs here too
	// (card t1533, review-gate r5): the stored hint names ONE predecessor,
	// and a candidate sharing other hub paths leased straight past their
	// still-in-flight sharers through the direct path. The wait is decided
	// before the promotion, so the refusal writes nothing, and it carries the
	// T2 guard's own sentinel so both paths read as the same predecessor
	// error.
	allRuns, err := db.ListCards(ctx, "")
	if err != nil {
		return nom, nil, err
	}
	if blocker, wait := factoryHubWaitUnmerged(queueRec, cards, factoryMergedCards(allRuns), cardID); wait {
		return nom, nil, fmt.Errorf("factory next: %w: %s has not reached %s (git-flow) or %s (github-flow)", homestate.ErrPredecessorUnmerged, blocker, homestate.CardMergedLocal, homestate.CardMergedPR)
	}
	// The claim would refuse a foreign tree only after the promotion; deciding
	// it here keeps the refusal write-free. The carry-over read runs first: a
	// landing directory the card's own previous run recorded is that card's
	// tree, and the claim at step 3 re-records the binding instead of
	// refusing (card t1521) — so this precheck refuses only the shapes the
	// carry-over does not explain.
	if nom.row == nil || strings.TrimSpace(nom.row.WorktreePath) == "" {
		carried, err := factoryCarryPreviousWorktree(ctx, db, root, runID, cardID)
		if err != nil {
			return nom, nil, err
		}
		if carried == "" {
			if err := factoryRefuseForeignWorktree(root, cardID); err != nil {
				return nom, factoryRefusal(factoryTokenForeignWorktree, "%s", strings.TrimPrefix(err.Error(), "factory next: refused — ")), nil
			}
		}
	}
	// The quota hold leaves only a card already assigned to this lane leasable
	// (arm (a) is not a new lease).
	if quotaHold != "" && !assignedHere {
		return nom, factoryRefusal(factoryRefuseQuotaHold, "%s", quotaHold), nil
	}
	if nom.row != nil && factoryNextSkipForBackend()(*nom.row) {
		return nom, factoryRefusal(factoryRefuseBackendSkip, "a Codex lane cannot advance a card recorded at %s", dash(nom.row.Stage)), nil
	}
	return nom, nil, nil
}

// factoryNextNominate is the nominated lease (SPEC-TODO-AUTO-PICK-001
// REQ-TAU-004/-005/-006): four steps in order.
//
//  1. Validate, read-only — every readable refusal is decided before the first
//     write (factoryNextValidate).
//  2. Promote a `queued` nominee to `picked` inside the queue lock, re-validating
//     there; a nominee that is no longer `queued` was taken by another lane
//     first (`raced`, nothing written). An operator-picked nominee skips this.
//  3. Call the nomination seam, then claim through the version-checked
//     record edges the unnominated arms use (factoryNextRecordAndClaim /
//     factoryNextClaim) — no second lease route.
//  4. Compensate: when the claim fails, undo only the promotion this
//     invocation made, and only while the record shows no row, or a row at
//     `picked` with no owner, and the queue item is still `picked`.
//
// The factory record has no delete: a failure AFTER RecordPicked leaves its
// `picked`, unowned row (spec § B.8); the queue item is still restored. The
// returned error is a *factoryNominateRefusal for a refusal of the card and a
// plain error for anything else (a compensation that itself failed included).
//
// @MX:NOTE: [AUTO] The nominated lease entry point — shared by the `next --card` verb and the factory_next MCP handler (one implementation per verb, design.md §3); the single seam factoryNominateBeforeRecord sits between the promotion and RecordPicked. Fan-in 2 — a NOTE because factory_card.go is at its 3-anchor limit.
// @MX:SPEC: SPEC-TODO-AUTO-PICK-001
func factoryNextNominate(ctx context.Context, root, runID, lane, cardID, quotaHold string) (homestate.Card, error) {
	// Opened before the section with the claim's busy timeout in its DSN (see
	// factoryNextLeaseOnceGated).
	db, err := homestate.OpenFactoryBounded(root, factoryLeaseClaimBusyTimeout)
	if err != nil {
		return homestate.Card{}, fmt.Errorf("open factory record: %w", err)
	}
	defer func() { _ = db.Close() }()

	var (
		card    homestate.Card
		outcome error
	)
	ran, lockErr := factoryLeaseSection(todoStoreAt(root), func(l *factory.LockedBacklog) {
		card, outcome = factoryNominateInSection(ctx, l, db, root, runID, lane, cardID, quotaHold)
	})
	switch {
	case !ran:
		if factory.IsStateLockHeld(lockErr) {
			return homestate.Card{}, factoryRefusal(factoryRefuseRaced, "the queue lock stayed held for the whole %s wait budget, so %s was not leased; retry", factory.LockWaitBudget(), cardID)
		}
		return homestate.Card{}, fmt.Errorf("take the queue lock: %w", lockErr)
	case outcome != nil:
		return homestate.Card{}, outcome
	}
	// A failed lock release after a committed lease does not undo the lease;
	// reporting it would orphan the card, and the next queue write meets the
	// held lock loudly.
	return card, nil
}

// factoryNominateInSection runs steps 1 to 4 of the nominated lease inside the
// lease section: validation reads the queue and the record under the lock, so
// the decision and the claim see one queue state with no writer between them.
func factoryNominateInSection(ctx context.Context, l *factory.LockedBacklog, db *homestate.FactoryDB, root, runID, lane, cardID, quotaHold string) (homestate.Card, error) {
	factoryLeaseAtPassEntry()
	nom, refusal, err := factoryNextValidate(ctx, l, db, root, runID, lane, cardID, quotaHold)
	if err != nil {
		return homestate.Card{}, err
	}
	if refusal != nil {
		return homestate.Card{}, refusal
	}

	promoted := false
	if nom.item.State == factory.BacklogStateQueued {
		var lost *factoryNominateRefusal
		if err := l.Mutate(func(r *factory.BacklogRecord) error {
			for i := range r.Items {
				it := &r.Items[i]
				if it.ID != cardID {
					continue
				}
				if it.State == factory.BacklogStateQueued {
					it.State = factory.BacklogStatePicked
					promoted = true
					return nil
				}
				lost = factoryRefusal(factoryRefuseRaced, "another lane moved %s out of queued first", cardID)
				return nil
			}
			lost = factoryRefusal(factoryRefuseRaced, "%s left the queue first", cardID)
			return nil
		}); err != nil {
			return homestate.Card{}, fmt.Errorf("promote the nominee: %w", err)
		}
		if lost != nil {
			return homestate.Card{}, lost
		}
	}

	// Hub-chain hint (REQ-TCI-020): computed inside the section from the
	// same locked queue the validation read — a read failing under a held
	// lock is infrastructural and fails the lease loudly, never silently
	// hint-less.
	queueRec, err := l.LoadPure()
	if err != nil {
		return homestate.Card{}, fmt.Errorf("read the queue for the hub chain: %w", err)
	}
	hubRows, err := db.ListCards(ctx, runID)
	if err != nil {
		return homestate.Card{}, fmt.Errorf("read the records for the hub chain: %w", err)
	}
	// A row that already carries a hint keeps it (card t1533): the generated
	// hint is a record-creation input, never an overwrite — the recomputed
	// tail drifts as the chain moves, and overwriting with it re-ordered a
	// bundle member against its own stored chain.
	hubHint := factoryGeneratedHubFields(factoryHubChainFields(queueRec, hubRows, cardID, nil), nom.row)

	// A claim that neither leased nor errored lost a race (the same signal the
	// unnominated arms re-select on); an error is a failure of the claim.
	card, leased, _, claimErr := factoryNextNominatedClaim(ctx, db, root, runID, lane, nom, hubHint)
	if claimErr == nil && leased {
		return card, nil
	}
	otherHolder, cerr := factoryNominateCompensate(ctx, l, db, runID, lane, cardID, promoted)
	switch {
	case cerr != nil:
		return homestate.Card{}, cerr
	case errors.Is(claimErr, errFactoryRecordBusy):
		return homestate.Card{}, factoryRefusal(factoryRefuseRaced, "the factory record was busy for the whole %s claim wait cap, so %s was not leased; retry", factoryLeaseClaimWaitCap, cardID)
	case otherHolder || claimErr == nil:
		return homestate.Card{}, factoryRefusal(factoryRefuseRaced, "another lane leased or claimed %s first", cardID)
	default:
		return homestate.Card{}, claimErr
	}
}

// factoryNextNominatedClaim is step 3: the seam, then the claim. A row
// already `assigned` to this lane is arm (a)'s lease edge; every other leasable
// shape (no row, or an unowned `picked` row) records the card and claims it,
// which re-reads the row fresh.
func factoryNextNominatedClaim(ctx context.Context, db *homestate.FactoryDB, root, runID, lane string, nom factoryNominee, fields homestate.CardFields) (homestate.Card, bool, bool, error) {
	if err := factoryNominateBeforeRecord(nom.item.ID); err != nil {
		return homestate.Card{}, false, false, err
	}
	if nom.row != nil && nom.row.State == homestate.CardAssigned {
		return factoryNextClaim(ctx, db, root, runID, *nom.row, lane)
	}
	return factoryNextRecordAndClaim(ctx, db, root, runID, nom.item.ID, lane, fields)
}

// factoryNominateCompensate is step 4. It acts only on a promotion this
// invocation made (promoted), and restores the queue item from `picked` to
// `queued` in one write that does nothing when the item is no longer `picked`,
// only while the record shows no row, or a row at `picked` with no owner. It
// reports whether another holder owns the card (their queue state is left
// alone). It runs inside the lease section, through the section's locked
// handle l: the promotion and its undo share one lock hold, so no operator
// write can land between them and be overwritten by the restore. The record row
// is read under that lock; a lease landing after the read, before the write, is
// the two-store window the spec accepts. A failure of the restoring write — or
// of the record read — is returned as a non-token error naming the card; the
// card then stays `picked` and unowned.
func factoryNominateCompensate(ctx context.Context, l *factory.LockedBacklog, db *homestate.FactoryDB, runID, lane, cardID string, promoted bool) (bool, error) {
	if !promoted {
		return false, nil
	}
	otherHolder := false
	if err := l.Mutate(func(r *factory.BacklogRecord) error {
		var row *homestate.Card
		cur, err := db.LoadCard(ctx, runID, cardID)
		switch {
		case errors.Is(err, homestate.ErrCardNotFound):
		case err != nil:
			return fmt.Errorf("read the factory record: %w", err)
		default:
			row = &cur
		}
		if row != nil && (row.State != homestate.CardPicked || strings.TrimSpace(row.OwnerLabel) != "") {
			// Another holder (or this lane's own assigned row): the queue state is
			// theirs and is left alone.
			otherHolder = row.OwnerLabel != lane && row.LeaseHolder != lane
			return nil
		}
		for i := range r.Items {
			if r.Items[i].ID == cardID && r.Items[i].State == factory.BacklogStatePicked {
				r.Items[i].State = factory.BacklogStateQueued
			}
		}
		return nil
	}); err != nil {
		return false, fmt.Errorf("compensation failed: card %s stays picked and unowned: %w", cardID, err)
	}
	return otherHolder, nil
}

// factoryNextSkipForBackend is the REQ-SD-025 selection half: a Codex lane
// never selects a card whose recorded stage or state it cannot advance —
// `merge-ready` or later, including a card returned to `assigned` by lease
// expiry with its stage kept. Every other backend selects freely.
func factoryNextSkipForBackend() func(homestate.Card) bool {
	if os.Getenv(config.EnvFactoryBackend) != factory.BackendGPT {
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
	case homestate.CardPROpen, homestate.CardMergedPR: // github-flow delivery states sit past merge-ready
		return true
	}
	return false
}

// newFactoryNextCommand — `moai factory next [--card <id>] [--wait]
// [--wait-bound <d>]` (REQ-SD-008/-009/-010): a lane session's self-dispatch
// verb, run from the parent checkout. Bare, it takes the CLI's priority-order
// choice; with --card it leases exactly the nominated card or refuses it
// (SPEC-TODO-AUTO-PICK-001: exit status 4, one line
// `factory next: refused <token>: <detail>` on the error stream).
func newFactoryNextCommand() *cobra.Command {
	var wait bool
	var waitBound time.Duration
	var run, nominee string
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Lease the lane's next card through the factory record (lane session, parent checkout)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("next")
			}
			lane, err := factoryLaneLabelFromEnv("next")
			if err != nil {
				return err
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
			// An explicit --card with a blank value is an error, never a silent
			// fall-back to the priority-order choice the session did not make.
			nominee = strings.TrimSpace(nominee)
			if cmd.Flags().Changed("card") && nominee == "" {
				return errors.New("factory next: --card needs a card id")
			}
			deadline := factoryCardNow().Add(waitBound)
			// The quota gate (SPEC-QUOTA-AWARE-SCHEDULING-001): evaluated once per
			// pass, outside the selection arms, and carried across the --wait
			// re-checks by the latch.
			quotaLatch := &factoryQuotaLatch{}
			for {
				held, holdLine := quotaLatch.evaluate(root)
				var card homestate.Card
				leased := false
				if nominee != "" {
					quotaHold := ""
					if held {
						quotaHold = holdLine
					}
					card, err = factoryNextNominate(ctx, root, runID, lane, nominee, quotaHold)
					var refusal *factoryNominateRefusal
					switch {
					case err == nil:
						leased = true
					case errors.As(err, &refusal):
						// --wait keeps waiting through the refusals a bare --wait also
						// waits through; a permanent refusal ends it at once.
						if wait && refusal.waitable() && factoryCardNow().Before(deadline) {
							factoryNextWaitSleep(factoryNextWaitInterval)
							continue
						}
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), refusal.Error())
						return refusal
					default:
						return fmt.Errorf("factory next: %w", err)
					}
				} else {
					card, leased, err = factoryNextLeaseOnceGated(ctx, root, runID, lane, held)
					if err != nil {
						return fmt.Errorf("factory next: %w", err)
					}
				}
				if leased {
					wt, _, err := factoryEnsureCardWorktree(ctx, root, runID, card, lane, cmd.OutOrStdout())
					if err != nil {
						return fmt.Errorf("factory next: %w", err)
					}
					card.WorktreePath = wt
					return factoryNextPrint(cmd, card)
				}
				if !wait || !factoryCardNow().Before(deadline) {
					if held {
						// A hold: one line on the error stream, nothing on standard
						// output, and the no-card status (REQ-QAS-009/-010); the hold
						// line is what tells it from an empty queue.
						_, _ = fmt.Fprintln(cmd.ErrOrStderr(), holdLine)
						return &exitCodeError{code: factoryNextNoCardExit, msg: "factory next: " + holdLine}
					}
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
	cmd.Flags().StringVar(&nominee, "card", "",
		"Lease exactly this card (the session's own judged pick) or refuse it: exit 4 with one line `factory next: refused <token>: <detail>`; the quota hold and the Codex skip apply as to any new card")
	return cmd
}

// factoryNextPrint reports the leased card (REQ-SD-008, plan B9) on the
// command's streams; the writer form is what the MCP factory_next handler
// shares (design.md §3 — one implementation per verb).
func factoryNextPrint(cmd *cobra.Command, card homestate.Card) error {
	return factoryNextWriteOutput(cmd.OutOrStdout(), cmd.ErrOrStderr(), factoryCardRoot(), card)
}

// factoryNextWriteOutput reports the leased card: its id, its stage, its
// worktree name, and its pull-request and landed state exactly as
// `moai todo pr` prints them — the pre-dispatch cross-check the leader
// performs, reported by the lane itself for a self-dispatched card. The
// queue read is anchored at root, the tree the verb resolved.
func factoryNextWriteOutput(out, errOut io.Writer, root string, card homestate.Card) error {
	wtName := dash("")
	if p := strings.TrimSpace(card.WorktreePath); p != "" {
		wtName = filepath.Base(p)
	}
	_, _ = fmt.Fprintf(out, "%s stage=%s worktree=%s\n", card.CardID, dash(card.Stage), wtName)
	rec, err := todoReadStoreAt(root).LoadPure()
	if err != nil {
		return fmt.Errorf("factory next: read the queue for the pr line: %w", err)
	}
	rows := computeTodoPRRows(errOut, rec, card.CardID)
	writeTodoPRRows(out, rec, rows)
	return nil
}

// newFactoryStageCommand — `moai factory stage <card> <state> [evidence]`
// (REQ-SD-012): the lane applies the card's next F1 stage transition, the
// evidence positional feeding the guards that need it.
func newFactoryStageCommand() *cobra.Command {
	var run string
	cmd := &cobra.Command{
		Use:   "stage <card> <state> [evidence]",
		Short: "Apply a card's next stage transition with its evidence (lane session)",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("stage")
			}
			lane, err := factoryLaneLabelFromEnv("stage")
			if err != nil {
				return err
			}
			evidence := ""
			if len(args) > 2 {
				evidence = args[2]
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			card, err := factoryStageCard(ctx, factoryCardRoot(), args[0], args[1], evidence, run, lane)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s v%d lease renewed\n", card.CardID, card.State, card.Version)
			return nil
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// factoryStageCard applies one stage transition as the lane (REQ-SD-012):
// the F1 edge through the Transition API with the lane's label as actor, the
// evidence positional (`<sha>` for the commit-only guards,
// `<sha>:<repo-relative-artifact>` where the guard also names the artifact)
// feeding the guards that read evidence themselves, and a lease renewal on
// success. It is the one implementation both the cobra RunE and the MCP
// factory_stage handler call (design.md §3); the REQ-SD-025 Codex
// merge-edge refusal rides inside, so every surface refuses it identically.
// The returned error carries the "factory stage:" prefix with the F1
// refusal verbatim inside.
func factoryStageCard(ctx context.Context, root, cardID, state, evidence, run, lane string) (homestate.Card, error) {
	refuse := func(err error) (homestate.Card, error) {
		return homestate.Card{}, fmt.Errorf("factory stage: %w", err)
	}
	if state == homestate.CardMerging {
		if err := factoryRefuseCodexMergeEdge("stage"); err != nil {
			return refuse(err)
		}
	}
	runID, err := resolveFactoryCardRun(ctx, root, run)
	if err != nil {
		return refuse(err)
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		return refuse(err)
	}
	defer func() { _ = db.Close() }()
	card, err := db.LoadCard(ctx, runID, cardID)
	if err != nil {
		return refuse(err)
	}
	req := homestate.TransitionRequest{
		RunID: runID, CardID: cardID, To: state,
		ExpectedVersion: card.Version, Actor: lane, Now: factoryCardNow(),
	}
	if sha, artifact, ok := strings.Cut(evidence, ":"); ok {
		req.SHA, req.ArtifactPath = strings.TrimSpace(sha), strings.TrimSpace(artifact)
	} else if strings.TrimSpace(evidence) != "" {
		req.SHA = strings.TrimSpace(evidence)
	}
	next, err := db.Transition(ctx, req)
	if err != nil {
		return refuse(err)
	}
	if _, err := db.RenewLease(ctx, runID, cardID, lane, factoryCardNow()); err != nil {
		return refuse(err)
	}
	return next, nil
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
		Short: "Take a merge-ready card through merging to merged-local; under github-flow, to pr-open then merged-pr (lane session)",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !factoryLaneAdmission() {
				return factoryNotALaneError("complete")
			}
			lane, err := factoryLaneLabelFromEnv("complete")
			if err != nil {
				return err
			}
			remeasure := ""
			if len(args) > 1 {
				remeasure = args[1]
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return factoryCompleteCard(ctx, cmd.OutOrStdout(), factoryCardRoot(), integrationLockRoot(), args[0], remeasure, run, lane)
		},
	}
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// factoryCompleteCard is the complete body after the lane checks: hold the
// integration window (the same record acquire writes, never a shell-out),
// apply the REQ-SD-023 refusals, merge --no-ff inside the worktree that has
// the integration branch checked out, and record the F1 edges T14 then T16.
// root and lockRoot name the tree the record and the window live in — the
// resolved project root on the MCP path (design.md §3, one implementation).
func factoryCompleteCard(ctx context.Context, out io.Writer, root, lockRoot, cardID, remeasure, run, lane string) error {
	// REQ-SD-025: the merge-ready → merging edge is refused on every path
	// while the backend identifies the Codex harness — inside the shared
	// body, so the CLI verb and the MCP tool refuse identically.
	if err := factoryRefuseCodexMergeEdge("complete"); err != nil {
		return err
	}
	// github-flow delivers by pull request, takes no integration window and
	// never merges locally (factory_card_pr.go, REQ-GFD-004/005/006).
	if factoryGitHubFlow(root) {
		return factoryCompleteGitHubFlow(ctx, out, root, cardID, remeasure, run, lane)
	}
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

	// REQ-MWQ-019 step 1 — the card gates run READ-ONLY, before the window
	// is taken or the branch moves: merge-ready, the caller's own unexpired
	// lease, version as read. O4 pins this to ONE read predicate shared
	// with the merge verb's gate — the state the gate sees is the state the
	// merge step sees, and the version rides through untouched.
	cardState, err := integrationReadMergeCardForRun(ctx, root, runID, cardID, lane)
	if err != nil {
		return fmt.Errorf("factory complete: %w", err)
	}
	if cardState.Stage != "merge-ready" {
		return fmt.Errorf("factory complete: refused — card %s is %q, not merge-ready; the integration branch and the card state are unchanged (REQ-MWQ-019 step 1)", cardID, cardState.Stage)
	}
	if !cardState.LeaseUnexpired {
		return fmt.Errorf("factory complete: refused — the caller does not hold card %s's unexpired lease; the integration branch and the card state are unchanged (REQ-MWQ-019 step 1)", cardID)
	}

	// REQ-SD-023: the window phase. A window this session already holds
	// keeps ITS recorded branch — the branch the window records is the
	// integration branch — so a lane that pre-acquired with --branch is not
	// re-resolved underneath its own choice. A window held by another live
	// session refuses naming the holder; a free or stale window is resolved
	// exactly as acquire resolves it and taken over. The refusal releases
	// NOTHING: the window is the other session's (O1).
	//
	// F7 (card-review r3): the merge step reads WindowLeaseDuration when its
	// seam is unset, so complete initializes the override here — the same
	// initialization every other window verb performs at entry — or a
	// configured lease_minutes: 0 would read as the shipped default in this
	// process alone.
	initWindowLeaseOverride(lockRoot)
	lock, err := factory.ReadIntegrationLock(lockRoot)
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
	if source == factory.BranchSourceCaller {
		fix := "re-acquire with --branch <integration-target>"
		if g := config.LoadGitFlowIntegrationConfig(root).EmptyTargetGuidance(root, fix); g != "" {
			fix = g
		}
		return fmt.Errorf("factory complete: refused — the integration window's branch %q is the caller's own tree (source %s); %s (a card's own tree is not its integration branch)", branch, factory.BranchSourceCaller, fix)
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
	acquiredHere := false
	if !heldByUs {
		ownerPID, _ := session.ResolveOwnerPID()
		// The per-card red hold (SPEC-CANDIDATE-CI-001 REQ-CCI-012, card t1478
		// Finding 4): complete's own acquisition is a window grant to this card,
		// so a red candidate refuses it before the window record is touched. The
		// refusal carries the landing refusal's exit code — the same red candidate,
		// reached one step earlier.
		if candidateCIEnabled(lockRoot) {
			if holdErr := factory.CandidateHoldRefusal(lockRoot, card.CardID); holdErr != nil {
				return &exitCodeError{code: factory.MergeExitLandingRefused, msg: fmt.Sprintf("factory complete: refused — %v", holdErr)}
			}
		}
		if factoryCompleteAcquireHook != nil {
			factoryCompleteAcquireHook()
		}
		replaced, err := factory.AcquireIntegrationWindow(lockRoot, factory.IntegrationLock{
			SessionID:    sessionID,
			SessionName:  lane,
			PID:          ownerPID,
			PIDSource:    factory.PIDSourceSessionOwner,
			Branch:       branch,
			BranchSource: source,
			Worktree:     integTree,
			Card:         card.CardID,
		}, false, &factory.AcquireWindowOptions{LeaseDuration: integrationLeaseDuration(lockRoot)})
		if err != nil {
			return fmt.Errorf("factory complete: %w", err)
		}
		acquiredHere = true
		if replaced != nil {
			// Never silent, exactly like acquire: the next lane must be able
			// to say what was cleared.
			_, _ = fmt.Fprintf(out, "  displaced stale window of %s (pid %d), held since %s\n", factoryHolderLabel(replaced), replaced.PID, replaced.AcquiredAt)
		}
	}

	// REQ-MWQ-019 step 2 — adoption. The integration branch's tip already
	// carries a merge commit whose second parent is the card branch's
	// CURRENT tip and whose tree matches a VALID re-measure record: the lane
	// merged through the merge verb first, and complete records merged-local
	// from that commit without calling the merge step and without a fresh
	// re-measure. A card branch that gained commits after that merge is NOT
	// adopted and falls through to step 3.
	tip := factoryBranchTip(integTree, branch)
	cardTip := factoryBranchTip(card.WorktreePath, cardBranch)
	adopted := ""
	if parents := factoryCommitParents(integTree, tip); len(parents) == 3 && parents[2] == cardTip {
		if tipTree := factoryTreeOf(integTree, branch); tipTree != "" {
			if record, recErr := factory.ReadRemeasureRecord(lockRoot, tipTree); recErr == nil && factory.ValidateRemeasureRecord(record) == nil {
				adopted = tip
			}
		}
	}
	if adopted != "" {
		done, err := completeTransitions(ctx, db, out, lockRoot, runID, card, cardID, lane, adopted, remeasure, branch, integTree)
		if err != nil {
			// O1: the adoption's transition failure is the post-merge class
			// (the commit already sits on the integration branch) — but the
			// window held by ANOTHER session is never released or altered;
			// the caller releases only its own hold.
			return completePostMergeConflict(out, lockRoot, sessionID, cardID, adopted, err)
		}
		_ = done
		// P2-6 (card-review r1): the adoption releases the window after its
		// transitions too — step 4's deferred release is not a reason for
		// step 2 to hold the window forever; the next live ticket is what
		// proves the release.
		if _, err := factory.ReleaseIntegrationLock(lockRoot, sessionID, 0, false); err != nil {
			_, _ = fmt.Fprintf(out, "  releasing the window failed (%v) — moai integration release by hand\n", err)
		}
		factoryPrintClearPolicyLine(out, root)
		return nil
	}

	// REQ-MWQ-019 step 3 — no valid re-measure record for the CURRENT
	// candidate tree refuses with the integration branch and the card state
	// unchanged, under the re-measure-and-re-acquire code.
	//
	// t1576 review round 2: the refusals also release the window complete
	// acquired in THIS invocation — returning with it held parked the next
	// lane until the lease lapsed. A hold the lane brought (heldByUs) is
	// its deliberate place and stays.
	releaseOwnAcquisition := func() {
		if !acquiredHere {
			return
		}
		if _, err := factory.ReleaseIntegrationLock(lockRoot, sessionID, 0, false); err != nil {
			_, _ = fmt.Fprintf(out, "  releasing the window failed (%v) — moai integration release by hand\n", err)
		}
	}
	candidateTree := factoryTreeOf(card.WorktreePath, cardBranch)
	if candidateTree == "" {
		releaseOwnAcquisition()
		return fmt.Errorf("factory complete: card %s's candidate tree cannot be read from %s — everything is unchanged (REQ-MWQ-019 step 3)", cardID, card.WorktreePath)
	}
	if _, err := factory.ReadRemeasureRecord(lockRoot, candidateTree); err != nil {
		releaseOwnAcquisition()
		return &exitCodeError{code: factory.MergeExitBaseMoved, msg: fmt.Sprintf("factory complete: refused — no valid re-measure record for the candidate tree %s (%v); run moai integration remeasure, then re-acquire --wait — the re-measure-and-re-acquire code", candidateTree[:12], err)}
	}

	// REQ-MWQ-019 step 4 — the merge runs ONLY by calling the REQ-MWQ-017
	// step (its own merge is gone), with the step's release DEFERRED until
	// complete's state transitions are done. The step's pre-merge causes
	// release inside it; a step failure leaves the card state unchanged
	// (REQ-MWQ-019: the card state shall not change).
	mergeSeams := factory.MergeStepSeams{
		ReadCard: func(id string) (factory.MergeCardState, error) {
			return integrationReadMergeCardForRun(ctx, root, runID, id, lane)
		},
	}
	// AC-CCI-011-1's both-call-sites clause: the self-issued merge path
	// wires the SAME shared landing check the merge verb does — its seams
	// formerly carried only ReadCard, and the step gates on
	// seams.LandingCheck != nil, so an unwired complete merged with no
	// landing gate at all. The step's own fail-closed guard refuses an
	// unwired seam when the key is enabled, so this wiring is mandatory
	// exactly while the gate is on.
	if candidateCIEnabled(lockRoot) {
		mergeSeams.LandingCheck = func(id, sha string) error {
			return factory.CandidateLandingCheck(factory.LandingCheckInput{
				Root:                lockRoot,
				CardID:              id,
				PinnedSHA:           sha,
				TargetBranch:        branch,
				IntegrationWorktree: integTree,
			})
		}
	}
	mergeSHA, err := factory.RunMergeStep(factory.MergeStepInput{
		Root:                lockRoot,
		IntegrationWorktree: integTree,
		IntegrationBranch:   branch,
		CardID:              cardID,
		CallerSessionID:     sessionID,
		DeferRelease:        true,
	}, mergeSeams)
	if err != nil {
		if code, ok := factory.MergeExitCode(err); ok {
			return &exitCodeError{code: code, msg: fmt.Sprintf("factory complete: the merge step refused: %v", err)}
		}
		return fmt.Errorf("factory complete: the merge step failed: %w", err)
	}

	// The transitions ride the merge (T14 then T16), and a transition
	// failing AFTER the merge commit exists is the post-merge-transition-
	// conflict outcome: the commit stays, the hold names it, the window
	// releases only after the hold is written, and the exit code is
	// complete's own (REQ-MWQ-019; distinct from the thirteen).
	if factoryCompleteTransitionHook != nil {
		factoryCompleteTransitionHook()
	}
	if _, err := completeTransitions(ctx, db, out, lockRoot, runID, card, cardID, lane, mergeSHA, remeasure, branch, integTree); err != nil {
		return completePostMergeConflict(out, lockRoot, sessionID, cardID, mergeSHA, err)
	}

	// The transitions finished: release the window (the step's deferred
	// release — the next live ticket is promoted).
	if _, err := factory.ReleaseIntegrationLock(lockRoot, sessionID, 0, false); err != nil {
		_, _ = fmt.Fprintf(out, "  releasing the window failed (%v) — moai integration release by hand\n", err)
	}
	factoryPrintClearPolicyLine(out, root)
	return nil
}

// completeTransitions runs T14 then T16 for the merge the step (or the
// adoption) produced: merge-ready → merging at the version step 1 read,
// then merging → merged-local through the F1 merge gate. remeasure stays
// what the lane passed — the GATE no longer reads a lane file for the
// re-measure (REQ-MWQ-021: the record is the gate), so a stand-in file can
// never satisfy it (REQ-MWQ-020).
func completeTransitions(ctx context.Context, db *homestate.FactoryDB, out io.Writer, lockRoot, runID string, card homestate.Card, cardID, lane, mergeSHA, remeasure, branch, integTree string) (homestate.Card, error) {
	merging, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: cardID, To: homestate.CardMerging,
		ExpectedVersion: card.Version, Actor: lane, Now: factoryCardNow(),
	})
	if err != nil {
		return homestate.Card{}, fmt.Errorf("the merge-ready → merging transition failed: %w", err)
	}
	done, err := db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: cardID, To: homestate.CardMergedLocal,
		ExpectedVersion: merging.Version, Actor: lane,
		MergeSHA: mergeSHA, RemeasurePath: remeasure, IntegrationBranch: branch, Now: factoryCardNow(),
		// REQ-MWQ-021: the re-measure the gate accepts is the RECORD keyed
		// to the merge commit's tree — a file naming the merge SHA as text
		// is no longer sufficient, so a stand-in can never satisfy the gate
		// (REQ-MWQ-020).
		VerifyRemeasure: func(treeSHA, mergeSHA string) error {
			rec, err := factory.ReadRemeasureRecord(lockRoot, treeSHA)
			if err != nil {
				return err
			}
			return factory.ValidateRemeasureRecord(rec)
		},
	})
	if err != nil {
		return homestate.Card{}, fmt.Errorf("the merging → merged-local transition failed: %w", err)
	}
	_, _ = fmt.Fprintf(out, "%s %s merge=%s branch=%s worktree=%s\n",
		done.CardID, done.State, done.MergeSHA, branch, integTree)
	return done, nil
}

// completePostMergeConflict is REQ-MWQ-019's conflict outcome: the merge
// commit stays on the integration branch, the policy holds with cause
// post-merge-transition-conflict naming the merge SHA (no queued ticket is
// promoted onto the conflict), the window is released ONLY after the hold
// is written — and only the caller's own hold; O1 pins that a window held
// by another session is never touched — and the exit code is complete's
// own, distinct from the thirteen.
func completePostMergeConflict(out io.Writer, lockRoot, sessionID, cardID, mergeSHA string, transitionErr error) error {
	if holdErr := factory.CompletePostMergeHold(lockRoot, cardID, mergeSHA); holdErr != nil {
		_, _ = fmt.Fprintf(out, "  writing the hold also failed (%v) — the merge commit %s stays on the integration branch; moai integration policy hold by hand\n", holdErr, mergeSHA[:12])
	}
	if _, err := factory.ReleaseIntegrationLock(lockRoot, sessionID, 0, false); err != nil {
		_, _ = fmt.Fprintf(out, "  releasing the window failed (%v) — release it by hand after reading the hold\n", err)
	}
	return &exitCodeError{code: completePostMergeTransitionConflictExit, msg: fmt.Sprintf("factory complete: %v — the merge commit %s stays on the integration branch; the window policy holds with cause post-merge-transition-conflict", transitionErr, mergeSHA[:12])}
}

// completePostMergeTransitionConflictExit is complete's own exit code
// (REQ-MWQ-019), distinct from the thirteen merge-step codes and the
// holder refusals.
const completePostMergeTransitionConflictExit = 20

// factoryCompleteTransitionHook is a TEST-ONLY interleaving point invoked
// between the merge step's success and complete's state transitions — the
// exact window a concurrent version bump occupies in AC-MWQ-019 scenario
// 8. Unexported and package-level, so only `package cli` assigns it, and
// no non-test file does (the closure-gate convention
// integrationLockMutationTestHook established). Every production path
// leaves it nil, and the call site is nil-guarded.
var factoryCompleteTransitionHook func()

// factoryCompleteAcquireHook is the acquisition's test seam (card t1478
// Finding 4): called just before complete takes the integration window for
// its card, after the card's own refusals. A red candidate must never reach it.
var factoryCompleteAcquireHook func()

// cliGitIn runs one git command in dir — the small adoption probes' helper.
func cliGitIn(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// factoryBranchTip reads dir's tip for branch (empty dir falls back to the
// process repository — the caller's resolution already ran).
func factoryBranchTip(dir, branch string) string {
	if strings.TrimSpace(dir) == "" {
		return ""
	}
	out, err := cliGitIn(dir, "rev-parse", "refs/heads/"+branch)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// factoryTreeOf reads dir's tip tree for branch.
func factoryTreeOf(dir, branch string) string {
	tip := factoryBranchTip(dir, branch)
	if tip == "" {
		return ""
	}
	out, err := cliGitIn(dir, "rev-parse", tip+"^{tree}")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// factoryCommitParents reads the tip's parent list (a two-parent merge
// commit yields three fields).
func factoryCommitParents(dir, tip string) []string {
	if tip == "" {
		return nil
	}
	out, err := cliGitIn(dir, "rev-list", "--parents", "-n", "1", tip)
	if err != nil {
		return nil
	}
	return strings.Fields(out)
}

// factoryRelaunchSupersededNote is the one line the relaunch policy prints
// before degrading to the one-shot lane session (card t1554): the
// supervising lease loop is removed — card consumption moved to the unified
// `moai todo --auto` engine — so the launcher starts one lane session, which
// consumes the queue itself.
const factoryRelaunchSupersededNote = "moai: --clear-policy relaunch is superseded (card t1554): card consumption moved to the unified `moai todo --auto` engine; starting one lane session"

// factoryClearPolicySelected reads the lane's clear policy from the carrier
// constant. Absence — and any value that is not one of the three policies —
// reads as the default clear-each (REQ-SD-020), matching the hook's
// fail-open reading of its environment.
func factoryClearPolicySelected() string {
	switch p := os.Getenv(config.EnvFactoryClearPolicy); p {
	case config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch:
		return p
	default:
		return config.FactoryClearPolicyEach
	}
}

// factoryContextAtHandoffThreshold reports whether the session's
// context-usage record shows usage at or above the model-specific handoff
// threshold (context-window-management rule): a large window
// (>= config.HandoffLargeWindowCutoff) hands off at
// config.HandoffSoftLargePct, a standard one at
// config.HandoffSoftStandardPct. A session with no resolvable id, a refused
// key, or a missing/unparseable record reads as BELOW threshold — the
// conservative answer that keeps the lane working instead of clearing.
func factoryContextAtHandoffThreshold(root string) bool {
	sessionID := strings.TrimSpace(os.Getenv(config.EnvClaudeCodeSessionID))
	if sessionID == "" {
		return false
	}
	path := statusline.SessionTelemetryPath(filepath.Join(root, ".moai", "state"), sessionID)
	if path == "" {
		return false
	}
	rec, err := statusline.ReadSessionTelemetry(path)
	if err != nil || rec == nil {
		return false
	}
	threshold := float64(config.HandoffSoftStandardPct)
	if rec.ContextWindowSize >= config.HandoffLargeWindowCutoff {
		threshold = float64(config.HandoffSoftLargePct)
	}
	return rec.RawPct >= threshold
}

// factoryPrintClearPolicyLine prints the one line the lane follows after a
// completion (REQ-SD-020): clear-each — the default, absence included —
// asks the operator to /clear; clear-when-full asks only once the session's
// context-usage record is at or above the handoff threshold and prints
// nothing below it; relaunch asks the operator to end the session so the
// supervising launcher starts the next card's session.
func factoryPrintClearPolicyLine(out io.Writer, root string) {
	policy := factoryClearPolicySelected()
	if policy == config.FactoryClearPolicyRelaunch {
		_, _ = fmt.Fprintln(out, "  clear policy relaunch: end this session; the supervising launcher starts the next card's session")
		return
	}
	if policy == config.FactoryClearPolicyWhenFull && !factoryContextAtHandoffThreshold(root) {
		return // below the handoff threshold: continue with the next card, no line
	}
	_, _ = fmt.Fprintf(out, "  clear policy %s: ask the operator to /clear, then take the next card on the fresh session\n", policy)
}

// factoryResolveIntegrationBranch mirrors acquire's branch resolution
// (resolveIntegrationTarget) without its $PWD legs: the configured integration
// target decides (git-flow: the develop branch; github-flow: main); with none
// configured the caller's own tree decided the window, which complete refuses,
// so the branch is the caller's — taken from the card worktree, the lane's own
// tree, never from the process cwd.
func factoryResolveIntegrationBranch(root string, card homestate.Card) (string, string) {
	if branch := strings.TrimSpace(config.LoadGitFlowIntegrationConfig(root).IntegrationTarget); branch != "" {
		return branch, factory.BranchSourceConfig
	}
	return factoryBranchOfWorktree(card.WorktreePath), factory.BranchSourceCaller
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

// factoryHolderLabel renders an integration-lock holder: the human-facing
// name a lane recognizes its queue position by, else the session id.
func factoryHolderLabel(lock *factory.IntegrationLock) string {
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

// (factoryMergeNoFF and factoryWriteMergeRecord were retired with card
// t1479, REQ-MWQ-019 step 4 and REQ-MWQ-020: complete merges ONLY by
// calling the REQ-MWQ-017 step, and no record complete writes ever stands
// in for the re-measure — the record store is the gate.)

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

// queueItemState reads the queue state of cardID, anchored at root; ok is
// false when the id is not in the queue at all. The factory record never
// writes the queue.
func queueItemState(root, cardID string) (factory.BacklogState, bool, error) {
	record, err := todoReadStoreAt(root).LoadPure()
	if err != nil {
		return "", false, err
	}
	state, ok := queueItemStateIn(record, cardID)
	return state, ok, nil
}

// queueItemIn is the item-returning search over a record already in hand —
// the live items, then the archive. The lease section reads the queue once
// through its locked handle and searches that record.
func queueItemIn(record *factory.BacklogRecord, cardID string) (factory.BacklogItem, bool) {
	for _, item := range record.Items {
		if item.ID == cardID {
			return item, true
		}
	}
	for _, entry := range record.Archived {
		if entry.Item.ID == cardID {
			return entry.Item, true
		}
	}
	return factory.BacklogItem{}, false
}

// queueItemStateIn is queueItemState's search over a record already in hand —
// the live items, then the archive. The lease section reads the queue once
// through its locked handle and searches that record.
func queueItemStateIn(record *factory.BacklogRecord, cardID string) (factory.BacklogState, bool) {
	if item, ok := queueItemIn(record, cardID); ok {
		return item.State, true
	}
	return "", false
}

// requireQueuePicked is the REQ-FR-022 precondition: only a card whose queue
// item is `picked` is admitted to the factory record. The gate enumerates
// POSITIVELY (SPEC-TODO-HOLD-STATE-001 REQ-THS-012) — the state it admits is
// named, every other state (a state added later included) refuses.
func requireQueuePicked(root, cardID string) error {
	state, ok, err := queueItemState(root, cardID)
	if err != nil {
		return fmt.Errorf("read queue: %w", err)
	}
	if !ok {
		return fmt.Errorf("queue item %s is not in the queue", cardID)
	}
	switch state {
	case factory.BacklogStatePicked:
		// the only admissible state for the factory record
	default:
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
			if err := requireQueuePicked(factoryCardRoot(), cardID); err != nil {
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
			// The record write runs under the queue lock, so a first
			// dispatch can never interleave with a completion that already
			// decided the factory DB was absent (review round-7 P1-REPEAT);
			// the two paths serialize on the same lock.
			queueStore := newTodoStore()
			// Hub-chain hint (REQ-TCI-020): filled ONLY on record creation,
			// and ONLY when the operator gave no --after of their own — an
			// explicit input outranks the computed one, and a card that
			// already has a record keeps whatever hint it carries.
			if fields.HintAfter == nil {
				if _, cerr := db.LoadCard(ctx, runID, cardID); errors.Is(cerr, homestate.ErrCardNotFound) {
					if rec, qerr := newTodoStore().LoadPure(); qerr == nil {
						if rows, rerr := db.ListCards(ctx, runID); rerr == nil {
							// Merge ONLY the computed after hint — the operator's
							// own fields (prefer, spec, worktree, contract) were
							// set above and must survive the fill (AC-TCI-020's
							// explicit-input-outranks clause cuts the other way
							// for --after alone, never for the whole struct).
							if hub := factoryHubChainFields(rec, rows, cardID, nil); hub.HintAfter != nil {
								fields.HintAfter = hub.HintAfter
							}
						} else {
							return fmt.Errorf("factory assign: read the records for the hub chain: %w", rerr)
						}
					} else {
						return fmt.Errorf("factory assign: read the queue for the hub chain: %w", qerr)
					}
				} else if cerr != nil {
					return fmt.Errorf("factory assign: %w", cerr)
				}
			}
			now := factoryCardNow()
			var card homestate.Card
			err = queueStore.WithLock(func(l *factory.LockedBacklog) error {
				recordCard, err := db.RecordPicked(ctx, runID, cardID, fields, "assign", now)
				if err != nil {
					return err
				}
				card = recordCard
				toTrim := strings.TrimSpace(to)
				if card.State != homestate.CardPicked {
					// State-preserving re-bind (review rounds 6-7): ANY
					// non-picked, non-legacy row recovers or refreshes its
					// dispatch binding without a state change, at every
					// post-assigned state (assigned, merge-ready,
					// merged-local, pushed, ci-green, done). The lane, when
					// given, must match the recorded owner; the row's
					// state, version, and owner are untouched.
					if !card.Legacy() && (toTrim == "" || toTrim == card.OwnerLabel) {
						// Binding recovery is leader/operator territory
						// (REQ-FCR-014): a lane session can never rotate the
						// dispatch binding — least of all onto a past row whose
						// old approval would then revive (review round-13 P1).
						if factoryLaneRefusal() {
							return fmt.Errorf("factory assign: refused — %s: dispatch-binding recovery is the leader path's act (%s=%s marks a lane)",
								factoryLaneBoundarySentinel, config.EnvFactoryRole, config.FactoryRoleLane)
						}
						// The queue's record is written FIRST (turn-end gate, card
						// t1538): this branch's factory write has no guard that can
						// refuse after it, so a record that cannot be written stops
						// the re-bind before the binding moves.
						if err := l.RefreshDispatchCurrent(cardID, runID, card.OwnerLabel); err != nil {
							return err
						}
						return db.RecordDispatchBinding(ctx, cardID, runID, now)
					}
					return fmt.Errorf("factory assign: card %s is %s (owner %q); only a picked card can move to assigned%s", cardID, card.State, dash(card.OwnerLabel), func() string {
						if !card.Legacy() && toTrim != "" && toTrim != card.OwnerLabel {
							return " a different lane"
						}
						return ""
					}())
				}
				if toTrim != "" {
					// T2 re-points the binding in its own transaction, so the
					// queue's current-dispatch identity (review round-24 P1-3)
					// is written inside it, right before its commit (turn-end
					// gate, card t1538): a record that cannot be written rolls
					// the assignment back — the binding stays where it was — and
					// a refused T2 never moves the record.
					card, err = db.Transition(ctx, homestate.TransitionRequest{
						RunID: runID, CardID: cardID, To: homestate.CardAssigned, ExpectedVersion: card.Version, Actor: "assign", Owner: to, Now: now,
						BeforeCommit: func(homestate.Card) error { return l.RefreshDispatchCurrent(cardID, runID, toTrim) },
					})
					return err
				}
				// A --to-less successful assign records THIS run as the
				// current dispatch at any version (review round-7 P1-1); the
				// queue's record is written first, as in the re-bind above.
				if err := l.RefreshDispatchCurrent(cardID, runID, card.OwnerLabel); err != nil {
					return err
				}
				return db.RecordDispatchBinding(ctx, cardID, runID, now)
			})
			if err != nil {
				return fmt.Errorf("factory assign: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %s v%d owner=%s\n", card.CardID, card.State, card.Version, dash(card.OwnerLabel))
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "assign the card to this lane label (picked → assigned)")
	cmd.Flags().StringVar(&prefer, "prefer", "", "assignment preference hint, key=value (reported, never enforced)")
	cmd.Flags().StringVar(&after, "after", "", "predecessor card that must reach merged-local (git-flow) or merged-pr (github-flow) first (\"\" clears)")
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
	// Mode and Priority are the card's recorded classification
	// (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-010), read from the queue so a
	// status reader sees the mode-aware dispatch state without opening the
	// queue. An absent classification reads as the derived default
	// (serial / normal — REQ-TCD-014), never as empty cells.
	Mode     string `json:"mode"`
	Priority string `json:"priority"`
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
	// Quota is the read-only quota block (SPEC-QUOTA-AWARE-SCHEDULING-001
	// REQ-QAS-013): present only while the quota gate is enabled and some window
	// has data, and last, so every pre-existing key keeps its place.
	Quota *factoryQuotaBlock `json:"quota,omitempty"`
}

func factoryCardViewOf(c homestate.Card, now time.Time, cls factory.CardClassification) factoryCardView {
	v := factoryCardView{
		RunID: c.RunID, CardID: c.CardID, State: c.State, Legacy: c.Legacy(), Stage: c.Stage, Version: c.Version,
		Owner: c.OwnerLabel, LeaseHolder: c.LeaseHolder, LeaseExpiresAt: c.LeaseExpiresAt, LeaseExpired: c.LeaseExpired(now),
		DecisionGate: c.DecisionGate, Question: c.DecisionQuestion, Resume: c.DecisionResume,
		Prefer: c.HintPrefer, After: c.HintAfter, SpecID: c.SpecID, FailureReason: c.FailureReason,
		Mode: cls.Mode, Priority: cls.Priority,
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
			// The classification snapshot for the mode/priority cells
			// (REQ-TCD-010): one pure queue read; a failed read degrades the
			// cells to the derived defaults rather than failing the whole
			// status surface (the lease rows survive a queue read fault).
			queueRec, queueErr := todoReadStoreAt(root).LoadPure()
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
					cls := factory.DefaultCardClassification()
					if queueErr == nil {
						cls = factoryQueueClassification(queueRec, c.CardID)
					}
					report.Cards = append(report.Cards, factoryCardViewOf(c, now, cls))
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
			report.Quota = factoryQuotaStatusBlock(root)
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
		_, _ = fmt.Fprintf(w, "%s run=%s state=%s stage=%s version=%d owner=%s lease=%s gate=%s prefer=%s after=%s contract=%s mode=%s priority=%s\n",
			c.CardID, c.RunID, state, dash(c.Stage), c.Version, dash(c.Owner), lease, dash(c.DecisionGate), dash(c.Prefer), dash(c.After), contract, dash(c.Mode), dash(c.Priority))
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
	if r.Quota != nil {
		writeFactoryQuotaText(w, r.Quota)
	}
}

// factoryDecideLaneRefusal is the REQ-SD-016 refusal — one wording source
// for the CLI decide guard and the MCP factory_decide handler.
func factoryDecideLaneRefusal() error {
	return fmt.Errorf("factory decide: refused — %s: decide records the operator's decisions; a lane session cannot (on the Codex MCP path the same refusal covers factory_decide)", factoryLaneBoundarySentinel)
}

// factoryDecideCards applies one gate/choice decision to each named card —
// the operator loop both the cobra RunE and the MCP factory_decide handler
// run (design.md §3, one implementation per verb). The queue-independent
// parts (the lane guard, the decider and gate/choice shape) stay with the
// callers; this is the record-writing body, anchored at root.
func factoryDecideCards(ctx context.Context, root string, out io.Writer, cards []string, gate, choice, run string) error {
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
	for _, cardID := range cards {
		card, err := decideOne(ctx, db, root, runID, cardID, gate, choice, integration)
		if err != nil {
			refused++
			_, _ = fmt.Fprintf(out, "%s: refused: %v\n", cardID, err)
			continue
		}
		_, _ = fmt.Fprintf(out, "%s: %s (v%d)\n", cardID, card.State, card.Version)
	}
	if refused > 0 {
		return fmt.Errorf("factory decide: %d of %d cards refused", refused, len(cards))
	}
	return nil
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
			// The one lane exception: a kickoff approval by the audit decider,
			// admitted per card only when the lane is the card's record owner
			// (checked in decideOne).
			auditApprove := decider == homestate.DeciderAudit && gate == "kickoff" && choice == "approve"
			if factoryLaneRefusal() && !auditApprove {
				return factoryDecideLaneRefusal()
			}
			if decider != homestate.DeciderHuman && !auditApprove {
				return fmt.Errorf("factory decide: decider %q is not accepted here; %q decides every gate and %q only --gate kickoff --choice approve", decider, homestate.DeciderHuman, homestate.DeciderAudit)
			}
			switch {
			case gate == "kickoff" && (choice == "approve" || choice == "reject"):
			case gate == "push" && choice == "":
			case gate == "" && (choice == "resume" || choice == "block" || choice == "unblock" || choice == "abandon"):
			default:
				return fmt.Errorf("factory decide: want --gate kickoff --choice approve|reject, --gate push, or --choice resume|block|unblock|abandon")
			}
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if auditApprove {
				return factoryAuditDecideCards(ctx, factoryCardRoot(), cmd.OutOrStdout(), args, run)
			}
			return factoryDecideCards(ctx, factoryCardRoot(), cmd.OutOrStdout(), args, gate, choice, run)
		},
	}
	cmd.Flags().StringVar(&gate, "gate", "", "decision gate: kickoff or push")
	cmd.Flags().StringVar(&choice, "choice", "", "approve|reject (kickoff), or resume|block|unblock|abandon")
	cmd.Flags().StringVar(&decider, "decider", homestate.DeciderHuman, "who decided: human, or audit (kickoff approve only, on verdict-file evidence)")
	cmd.Flags().StringVar(&run, "run", "", "factory run id (default: the single active run)")
	return cmd
}

// decideOne applies one card's decision as its own version-checked
// transition; a refusal for one card does not affect the others. projectRoot
// is the same root the factory DB was opened at: the queue record the
// approval uuid is read from MUST come from this root, not the server's cwd
// (review round-6 P2).
func decideOne(ctx context.Context, db *homestate.FactoryDB, projectRoot, runID, cardID, gate, choice, integration string) (homestate.Card, error) {
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
	case gate == "push" && cur.State == homestate.CardMergedPR:
		to = homestate.CardDone // github-flow: the merge is already on the remote — the gate closes the card, nothing is pushed
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
	// Receipt-relevant edges bind the backlog identity: the done edges for
	// the close check (review round-6 P1-2) and T17 for the push chain's
	// re-stamp (review round-8 P2-2). Other decisions — abandon, block,
	// unblock, resume, kickoff — are factory-only recovery paths and never
	// touch the queue record, so a corrupt or unreadable queue cannot block
	// them (review round-10 P2-2). The uuid is identity knowledge read from
	// the queue record at the SAME project root the factory DB was opened
	// at (review round-6 P2): a server-cwd queue is never substituted for
	// the target project.
	needsUUID := to == homestate.CardDone || to == homestate.CardPushed
	approvalUUID := ""
	if needsUUID {
		queueStore := factory.NewBacklogStore(todoBacklogPath(projectRoot))
		rec, err := queueStore.LoadPure()
		if err != nil {
			return cur, fmt.Errorf("factory decide: the queue could not be read for the approval check: %w", err)
		}
		for i := range rec.Items {
			if rec.Items[i].ID == cardID {
				approvalUUID = todoCardUUID(&rec.Items[i])
				break
			}
		}
		if approvalUUID == "" {
			for i := range rec.Archived {
				if rec.Archived[i].Item.ID == cardID {
					approvalUUID = todoCardUUID(&rec.Archived[i].Item)
					break
				}
			}
		}
		if to == homestate.CardDone && approvalUUID == "" {
			return cur, fmt.Errorf("factory decide: card %s has no backlog identity for the approval check", cardID)
		}
	}
	return db.Transition(ctx, homestate.TransitionRequest{
		RunID: runID, CardID: cardID, To: to, ExpectedVersion: cur.Version,
		Actor: "operator", Decider: homestate.DeciderHuman, IntegrationBranch: integration, Now: factoryCardNow(),
		ApprovalUUID: approvalUUID,
	})
}
