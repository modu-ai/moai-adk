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
const (
	boardLockRetryDelay = 25 * time.Millisecond
	boardLockRetries    = 40
)

// acquireBoardLockSerialized acquires the board-wide lock, retrying contention
// for a bounded window. Returns ErrBoardLockHeld if the window elapses.
func acquireBoardLockSerialized(root string) (*BoardLock, error) {
	var lastErr error
	for attempt := 0; attempt <= boardLockRetries; attempt++ {
		lock, err := AcquireBoardLock(root)
		if err == nil {
			return lock, nil
		}
		if !IsBoardLockHeld(err) {
			return nil, err
		}
		lastErr = err
		time.Sleep(boardLockRetryDelay)
	}
	return nil, lastErr
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
func WriteBoardState(root, sessionID string, mutate func(*BoardState) error) error {
	if err := requireLeadRole(root, sessionID); err != nil {
		return err
	}

	lock, err := acquireBoardLockSerialized(root)
	if err != nil {
		return fmt.Errorf("write board state: %w", err)
	}
	defer func() { _ = lock.Release() }()

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
	// board-wide read. Refuse before the atomic write; the empty id is its
	// own named refusal (ValidateSpecID passes "").
	for i, card := range st.Cards {
		if card.SpecID == "" {
			return fmt.Errorf("write board state: card[%d] carries an empty spec id", i)
		}
		if err := specid.ValidateSpecID(card.SpecID); err != nil {
			return fmt.Errorf("write board state: card[%d] spec id: %w", i, err)
		}
	}

	return writeBoardAtomic(root, st)
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
