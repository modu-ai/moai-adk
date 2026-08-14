// board_recover.go — the bounded recovery out of the unknown-board state
// (SPEC-KANBAN-BOARD-001 REQ-KB-013/022, M1).
//
// The unknown state must be escapable rather than terminal, and the exit is
// an EXPLICIT operator-visible act performed by the sole writer while holding
// the board-wide lock — never a fallback the read path takes. Until recovery
// runs, every dispatch is refused and no empty board is presented.
package kanban

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Recovery verdicts — one definite verdict per invocation (REQ-KB-022).
const (
	// RecoveryVerdictRecovered: the prior state file was readable (or
	// absent); nothing was replaced and no card moved.
	RecoveryVerdictRecovered = "recovered"
	// RecoveryVerdictReplaced: the prior state file was unreadable; a fresh
	// empty board was written and what could not be recovered was recorded
	// durably in a sidecar artifact.
	RecoveryVerdictReplaced = "replaced"
	// RecoveryVerdictNotRecoverable: the loss could not be recorded, so
	// nothing was replaced — presenting a replacement without a record of
	// what it lost would be the empty-board harm wearing the word
	// "explicit" (AP-26).
	RecoveryVerdictNotRecoverable = "not-recoverable"
)

// recoverySidecarPrefix names the loss-record artifacts beneath the board
// directory (one per replaced recovery, timestamped).
const recoverySidecarPrefix = "recovery-"

// RecoveryResult reports what one recovery invocation observed and did.
type RecoveryResult struct {
	// Verdict is exactly one of the RecoveryVerdict* values.
	Verdict string
	// LostRecorded is true when the prior content could not be read and was
	// preserved in a durable sidecar artifact.
	LostRecorded bool
	// SidecarPath names the loss record when LostRecorded is true.
	SidecarPath string
	// Summary states, for the operator, what happened.
	Summary string
}

// RecoverBoard is the explicit, bounded recovery operation: the sole writer
// (the caller's declared role must be `lead` — recovery is a board write),
// holding the board-wide lock, reconstructs or replaces the state file so
// the board leaves the unknown state.
//
// Bounded in EXTENT (REQ-KB-022): it modifies the board state file alone —
// no SPEC frontmatter is written, no worktree removed, and no card it can
// still read moves. Bounded in EFFECT: one invocation, one definite verdict
// (recovered / replaced / not-recoverable), no re-entry, no retry. Where the
// prior content cannot be reconstructed, what was lost is recorded durably
// in a sidecar beneath the board directory and surfaced in the result —
// never presented as the board that was lost.
func RecoverBoard(root, sessionID string) (result *RecoveryResult, err error) {
	// Recovery IS a board write, so the sole-writer guard binds it too.
	if err := requireLeadRole(root, sessionID); err != nil {
		return nil, err
	}

	lock, err := acquireBoardLockSerialized(root)
	if err != nil {
		return nil, fmt.Errorf("recover board: %w", err)
	}
	// See WriteBoardState: release errors are joined, not discarded (F5).
	defer func() {
		if relErr := lock.Release(); relErr != nil && err == nil {
			err = fmt.Errorf("recover board: verdict delivered but lock release failed: %w", relErr)
		}
	}()

	raw, err := os.ReadFile(BoardPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			// Absent is not damage: a legitimately empty board needing no
			// recovery (REQ-KB-021). The sole writer creates the directory
			// so the next write has a home.
			if merr := os.MkdirAll(BoardDir(root), 0o755); merr != nil {
				return nil, fmt.Errorf("recover board: creating board dir: %w", merr)
			}
			return &RecoveryResult{
				Verdict: RecoveryVerdictRecovered,
				Summary: "no state file present — a legitimately empty board, nothing recovered",
			}, nil
		}
		return nil, fmt.Errorf("recover board: reading state file: %w", err)
	}

	var st BoardState
	if jerr := json.Unmarshal(raw, &st); jerr == nil {
		// Readable: the safe exit. Move no card it can still read.
		return &RecoveryResult{
			Verdict: RecoveryVerdictRecovered,
			Summary: "state file readable — recovered in place, no card moved",
		}, nil
	}

	// Unreadable: record what was lost durably, THEN replace. If the loss
	// cannot be recorded, do not replace — an unrecorded replacement is the
	// empty-board harm through the door labelled "explicit".
	sidecarPath, serr := writeRecoverySidecar(root, raw)
	if serr != nil {
		return nil, fmt.Errorf("recover board: could not record the lost content, nothing replaced: %w", serr)
	}

	if werr := writeBoardAtomic(root, &BoardState{Cards: []Card{}}); werr != nil {
		return nil, fmt.Errorf("recover board: writing replacement state (loss already recorded at %s): %w", sidecarPath, werr)
	}
	return &RecoveryResult{
		Verdict:      RecoveryVerdictReplaced,
		LostRecorded: true,
		SidecarPath:  sidecarPath,
		Summary:      fmt.Sprintf("state file was unreadable; replaced with an empty board and preserved the lost content at %s", sidecarPath),
	}, nil
}

// recoveryLossRecord is the durable sidecar documenting what a replacement
// could not recover: the raw bytes of the unreadable prior state file,
// preserved verbatim.
type recoveryLossRecord struct {
	RecordedAt   string `json:"recorded_at"`
	Outcome      string `json:"outcome"`
	Reason       string `json:"reason"`
	RawPreserved string `json:"raw_preserved"`
}

// writeRecoverySidecar preserves the unreadable prior content beneath the
// board directory and returns the artifact's path.
func writeRecoverySidecar(root string, raw []byte) (string, error) {
	dir := BoardDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating board dir: %w", err)
	}
	rec := recoveryLossRecord{
		RecordedAt:   time.Now().UTC().Format(time.RFC3339),
		Outcome:      RecoveryVerdictReplaced,
		Reason:       "prior board state file existed but could not be parsed; its cards could not be reconstructed",
		RawPreserved: string(raw),
	}
	encoded, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encoding loss record: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102T150405.000000000")
	path := filepath.Join(dir, recoverySidecarPrefix+stamp+".json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("writing loss record: %w", err)
	}
	return path, nil
}
