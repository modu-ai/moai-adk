// board_store.go — the guarded write path: sole writer, board-wide lock,
// atomic replacement (SPEC-KANBAN-BOARD-001 REQ-KB-017/018/009, M1).
//
// Exactly ONE guarded write entry point exists for the board state file
// (AC-KB-017's static half — enforced here, not documented): every mutation
// resolves the caller's DECLARED role and refuses anything but `lead`, takes
// the board-wide lock across the entire read-modify-write, and lands the
// write through the same-directory temp + atomic rename. No call site writes
// the state file directly; the write helper below is unexported and reachable
// only from this entry point and the bounded recovery's sole-writer path.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/cli/specid"
)

// ErrNotSoleWriter is returned when a session whose declared role is not
// `lead` — including a session with no readable declaration at all — attempts
// a board write. The guard fails CLOSED: a refusal with nothing to read does
// not refuse, it admits every write (spec.md §A.8).
var ErrNotSoleWriter = errors.New("kanban board write refused: caller is not the lead")

// IsNotSoleWriter reports whether err is the sole-writer refusal.
func IsNotSoleWriter(err error) bool {
	return errors.Is(err, ErrNotSoleWriter)
}

// ErrWipLimitExceeded is the named refusal for a transition into `run` that
// would make the third concurrent card (REQ-KB-009). The board is left
// unchanged.
var ErrWipLimitExceeded = errors.New("kanban board WIP limit exceeded for the run column")

// IsWipLimitExceeded reports whether err is the WIP refusal.
func IsWipLimitExceeded(err error) bool {
	return errors.Is(err, ErrWipLimitExceeded)
}

// runWIPLimitDefault is the board's `run` concurrency bound (REQ-KB-009):
// at most two cards occupy `run` at once. It is a property of the BOARD and
// is deliberately independent of the deployed coder-session count
// (REQ-KB-010) — nothing here reads a session count.
//
// The bound is enforceable only beneath the board-wide lock: two concurrent
// transitions of two different cards each holding only their own card's lock
// would each observe the bound satisfied and each write, landing at WIP 3.
const runWIPLimitDefault = 2

// BoardOptions carries the board's OWN inputs (REQ-KB-010): the WIP bound is
// one of them, and the deployed coder-session count is deliberately NOT —
// the admission path reads no session-count value from configuration,
// topology, or any observable. A zero RunWIPLimit selects the default.
type BoardOptions struct {
	RunWIPLimit int
}

// runLimit resolves the effective `run` WIP bound.
func (o BoardOptions) runLimit() int {
	if o.RunWIPLimit > 0 {
		return o.RunWIPLimit
	}
	return runWIPLimitDefault
}

// Board-lock acquisition is non-blocking at the substrate, so a mutation
// racing the current holder retries for a bounded window rather than failing
// the mutation: two concurrent transitions must BOTH reach the admission
// decision in turn (exactly one succeeding on the WIP bound), not have one
// bounce off the lock. The window is bounded so a genuinely stuck holder
// surfaces as an error instead of a hang.
//
// The queue lock-wait policy below is the SHARED budget-and-backoff both
// queue-lock acquisition paths consume — this one and (*BacklogStore).acquireLock
// (backlog_store.go). A change here applies to both without a second edit
// (SPEC-BACKLOG-LOCK-BUDGET-001 REQ-BLB-006).
//
// What the derivation is NOT. The substrate acquire is
// flock(LOCK_EX|LOCK_NB) (board_lock_unix.go) and the callers poll, so the
// lock provides no queue and no fairness: a contender that loses a race gains
// no priority for the next round, and flock guarantees no ordering between
// waiters. A contender's maximum wait is therefore the tail of a run of lost
// races, not a queue position, and it is unbounded in principle — no
// closed-form worst case exists. The product below is a SIZING HEURISTIC WITH
// STATED HEADROOM, NOT a worst-case bound, and must not be read as one. It is
// sized to survive a bounded run of lost races at the supported contender
// count on a CI-class machine; it is not sized for FIFO queue depth, which
// does not describe this lock.
const (
	// boardLockSupportedWriters is the concurrent lane count the product
	// supports against one queue: Factory mode runs up to ten lanes, the
	// figure of record in backlog_concurrency_test.go's header comment.
	boardLockSupportedWriters = 10

	// boardLockCIMutationCost is the per-mutation cost observed on a
	// CI-class machine under -race: 1.57s across 48 serialized mutations,
	// or ~33ms each. The isolated local figure (~14ms) is deliberately NOT
	// the input — sizing the budget to the faster machine is what left the
	// retired 1.025s window 87% consumed on a machine where the guard
	// passed.
	boardLockCIMutationCost = 33 * time.Millisecond

	// boardLockHeadroom is the stated headroom factor over the product
	// above. Five means: survive roughly five consecutive rounds of losing
	// to every peer before a stuck holder is declared.
	//
	// The product boardLockSupportedWriters * boardLockHeadroom = 50 is the
	// serialized-mutation count this policy budgets for, and it sits barely
	// above the 8 x 6 = 48 mutations TestConcurrencyStress serializes through
	// one flock. That near-coincidence used to be accidental;
	// TestBoardLockWaitBudgetCoversSerializedMutations (board_lock_wait_test.go,
	// SPEC-STRESS-INVARIANT-VERDICT-001) now pins it, so lowering either
	// constant here fails that guard. It is a constant-coherence relation and
	// asserts nothing about the wait a real machine needs.
	boardLockHeadroom = 5

	// boardLockWaitBudget is the derived elapsed window a contender polls
	// before giving up. Bounding by elapsed time rather than by an attempt
	// count is what makes the window a duration the derivation can state.
	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom

	// boardLockWaitMin and boardLockWaitMax bound every per-attempt wait the
	// policy produces. They are the policy's declared contract; tests assert
	// against these rather than against a sampled value.
	boardLockWaitMin = 5 * time.Millisecond
	boardLockWaitMax = 50 * time.Millisecond

	// boardLockWaitStep grows the jitter band per attempt, so early retries
	// poll tightly and a long wait backs off toward the ceiling.
	boardLockWaitStep = 10 * time.Millisecond
)

// boardLockRetryWait returns the wait before retry number attempt (0-based).
//
// It is jitter applied over a modest linear backoff, and both halves are
// load-bearing. Backoff alone would NOT break the lockstep: contenders
// released together grow their delays identically and keep arriving at the
// same bad moment relative to the holder's release. Jitter is what makes two
// contenders at the same attempt index diverge, so no contender is
// systematically beaten by the same peers (REQ-BLB-003/004).
//
// The randomness comes from math/rand/v2's per-call top-level source: no
// package-level seeding, and no global generator a test would have to
// control.
func boardLockRetryWait(attempt int) time.Duration {
	ceil := boardLockWaitMin + time.Duration(attempt+1)*boardLockWaitStep
	if ceil > boardLockWaitMax {
		ceil = boardLockWaitMax
	}
	span := ceil - boardLockWaitMin
	if span <= 0 {
		return boardLockWaitMin
	}
	return boardLockWaitMin + time.Duration(rand.Int64N(int64(span)+1))
}

// acquireBoardLockSerialized acquires the board-wide lock, retrying contention
// until the shared wait budget elapses. Returns ErrBoardLockHeld if it does.
func acquireBoardLockSerialized(root string) (*BoardLock, error) {
	var lastErr error
	deadline := time.Now().Add(boardLockWaitBudget)
	for attempt := 0; ; attempt++ {
		lock, err := AcquireBoardLock(root)
		if err == nil {
			return lock, nil
		}
		if !IsBoardLockHeld(err) {
			return nil, err
		}
		lastErr = err
		if !time.Now().Before(deadline) {
			return nil, lastErr
		}
		time.Sleep(boardLockRetryWait(attempt))
	}
}

// requireLeadRole resolves the caller's declared role and refuses anything
// but `lead` (REQ-KB-017's runtime half — the role is READ from the
// declaration carrier, never derived from a session identifier or a launch
// label). It is the guard BOTH guarded entry points run before touching the
// board: the mutation path and the bounded recovery, which is a board write
// too.
func requireLeadRole(root, sessionID string) error {
	role, err := ResolveDeclaredRole(root, sessionID)
	if err != nil {
		return fmt.Errorf("%w: session %q has no readable role declaration: %v", ErrNotSoleWriter, sessionID, err)
	}
	if role != RoleLead {
		return fmt.Errorf("%w: session %q declared role %q", ErrNotSoleWriter, sessionID, role)
	}
	return nil
}

// @MX:ANCHOR: [AUTO] WriteBoardState — the sole guarded board-write entry point (REQ-KB-017)
// @MX:REASON: expected fan_in >= 3 (TransitionIntoRun, M2 admissions/transitions, any future board mutator); a second write path around it is the AP-17 defect
//
// WriteBoardState is THE guarded entry point for board state mutations: it
// enforces the sole-writer role guard, then holds the board-wide lock across
// the entire read-modify-write (REQ-KB-019): the read of the board, mutate's
// admission or transition decision, and the atomic write.
//
// mutate operates on the loaded state in place; returning an error from
// mutate aborts the mutation with the board unchanged. An unknown board
// (unreadable state file) propagates as ErrBoardUnknown — the write path
// repairs nothing (REQ-KB-013; recovery is the explicit operation in
// board_recover.go).
//
// The write itself creates the temp file in the TARGET'S OWN directory and
// renames it over the state file (REQ-KB-018): rename(2) is atomic only
// within one filesystem, so a temp file in the system temp directory could
// cross a device boundary and degrade into the partially-observable write
// this exists to prevent. On the absent-file path the sole writer creates
// the state directory (REQ-KB-021 — it is gitignored and cannot ship).
func WriteBoardState(root, sessionID string, mutate func(*BoardState) error) (err error) {
	if err := requireLeadRole(root, sessionID); err != nil {
		return err
	}

	lock, err := acquireBoardLockSerialized(root)
	if err != nil {
		return fmt.Errorf("write board state: %w", err)
	}
	// Release errors are JOINED into the result rather than discarded (see
	// joinBoardReleaseErr): on Windows release removes the artifact, so a
	// failed removal would silently block every later writer (sync-audit F5).
	// The join runs in BOTH directions — a release failure arriving alongside
	// a mutation failure must survive too, or the wedged lock hides behind an
	// unrelated refusal message.
	defer func() {
		err = joinBoardReleaseErr(err, lock.Release(), "write board state")
	}()

	st, err := LoadBoard(root)
	if err != nil {
		// Unknown board: propagate, repair nothing.
		return fmt.Errorf("write board state: %w", err)
	}

	if err := mutate(st); err != nil {
		return fmt.Errorf("write board state: mutation refused: %w", err)
	}

	// Post-mutate SpecID sweep — re-review ITEM 1: this is the exported
	// mutation anchor (every board mutator funnels through here), so the
	// identity guard must hold HERE, not only at one caller's argument. A
	// traversal-shaped SpecID reaching board.json would be an identity the
	// board can never subsequently read (the read boundary refuses it), and
	// since ReadCardStatus propagates errors, one such card can fail a
	// board-wide read. Refuse before the atomic write.
	//
	// The sweep is SHAPE-CONDITIONAL and only shape-conditional (rereview2,
	// operator decision — minimal fix): an earlier form also refused an EMPTY
	// spec id here, an addition beyond AC-KB-022's conjunct ("conditional on
	// shape, never unconditional") with no traversal shape at all — and it
	// WEDGED a readable board, because a state document carrying one empty id
	// (reachable from this branch's own earlier history) then refused every
	// mutation, including one that would repair the bad card, while
	// RecoverBoard correctly reported the board readable. An empty id stays
	// useless (ReadCardStatus refuses one); it no longer freezes the board.
	for i, card := range st.Cards {
		if err := specid.ValidateSpecID(card.SpecID); err != nil {
			return fmt.Errorf("write board state: card[%d] spec id: %w", i, err)
		}
	}

	return writeBoardAtomic(root, st)
}

// joinBoardReleaseErr folds a board-lock release failure into the mutation's
// own result. It JOINS rather than overwrites, and joins in both directions: a
// release failure that arrives alongside a mutation failure is the dangerous
// case, not the harmless one. Reporting only the mutation error there hides a
// wedged lock behind an unrelated message — on Windows release removes the
// artifact, so the survivor blocks every later writer while the operator
// retries against what looks like a transient mutation fault.
//
// op names the calling entry point ("write board state" / "recover board") so
// the joined release error reads as that site's own.
//
// Returns nil only when both are nil, so the caller's `err` stays clean on the
// success path.
func joinBoardReleaseErr(mutErr, relErr error, op string) error {
	if relErr == nil {
		return mutErr
	}
	return errors.Join(mutErr, fmt.Errorf("%s: lock release failed: %w", op, relErr))
}

// writeBoardAtomic persists st through the same-directory temp + atomic
// rename. Unexported and reachable only from WriteBoardState and the bounded
// recovery — the sole-writer rule is enforced at those entry points, so no
// other call site can reach the file.
func writeBoardAtomic(root string, st *BoardState) error {
	dir := BoardDir(root)
	// The sole writer creates the state directory on the write path: it is
	// gitignored and therefore cannot be present on arrival (REQ-KB-021).
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("write board state: creating board dir: %w", err)
	}

	encoded, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("write board state: encoding: %w", err)
	}
	encoded = append(encoded, '\n')

	// Temp file in the TARGET'S OWN directory — never the system temp dir
	// (REQ-KB-018, AP-20): the rename must stay inside one filesystem to be
	// atomic.
	tmp, err := os.CreateTemp(dir, ".board-*.tmp")
	if err != nil {
		return fmt.Errorf("write board state: creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(encoded); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write board state: writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write board state: closing temp file: %w", err)
	}

	if err := atomicfile.Replace(tmpName, BoardPath(root)); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("write board state: replacing state file: %w", err)
	}
	return nil
}

// TransitionIntoRun moves the card for specID into the `run` column beneath
// the board-wide lock with the board's default options, refusing with
// ErrWipLimitExceeded — the board unchanged — when the transition would
// exceed the bound (REQ-KB-009).
func TransitionIntoRun(root, sessionID, specID string) error {
	return TransitionIntoRunOpts(root, sessionID, specID, BoardOptions{})
}

// TransitionIntoRunOpts is TransitionIntoRun with explicit board options —
// the knob AC-KB-012 varies while holding every other input fixed. The
// admission is never gated on session availability (REQ-KB-011): a card
// admitted with no session free is recorded UNHELD and that is a legal
// steady state, not an error or a stall.
func TransitionIntoRunOpts(root, sessionID, specID string, opts BoardOptions) error {
	// A traversal-shaped specID is not persistable: it would land in the
	// board state file as the card's identity and later reach a status-read
	// path join. Refuse at the write boundary, matching ReadCardStatus. An
	// empty id is its own named refusal (ValidateSpecID passes "").
	if specID == "" {
		return fmt.Errorf("transition into run: empty spec id")
	}
	if err := specid.ValidateSpecID(specID); err != nil {
		return fmt.Errorf("transition into run: %w", err)
	}
	return WriteBoardState(root, sessionID, func(st *BoardState) error {
		limit := opts.runLimit()
		inRun := 0
		existing := -1
		for i := range st.Cards {
			if st.Cards[i].Column == ColumnRun {
				inRun++
			}
			if st.Cards[i].SpecID == specID {
				existing = i
			}
		}
		// A card already in run re-enters idempotently; a card elsewhere
		// moving in occupies one more run slot, so the bound is checked
		// before any move.
		if existing >= 0 && st.Cards[existing].Column == ColumnRun {
			return nil
		}
		if inRun >= limit {
			return fmt.Errorf("%w: %d cards already occupy run (bound %d)", ErrWipLimitExceeded, inRun, limit)
		}
		moved := time.Now().UTC().Format(time.RFC3339)
		if existing >= 0 {
			st.Cards[existing].Column = ColumnRun
			st.Cards[existing].LastMovedAt = moved
			return nil
		}
		st.Cards = append(st.Cards, Card{SpecID: specID, Column: ColumnRun, LastMovedAt: moved})
		return nil
	})
}
