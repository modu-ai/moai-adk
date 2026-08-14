// board.go — single-origin kanban board state store
// (SPEC-KANBAN-BOARD-001 REQ-KB-005/006/013/021, M1).
//
// The board state has ONE origin: the primary checkout. Every session —
// whichever worktree it runs in — resolves that one path rather than its own
// tree's copy, because a worktree's .moai/state/ is gitignored and therefore
// private to that tree; a board read out of each session's own tree is six
// boards.
package kanban

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gitcore "github.com/modu-ai/moai-adk/internal/core/git"
)

// boardDirSegments is the board's OWN home beneath the primary root. It is
// deliberately NOT the session record's stateDirSegments (record.go): that
// constant resolves per-tree beneath each session's own root, which is
// correct for the session record and wrong for the board. Reusing it would
// land one board per worktree — AP-24.
var boardDirSegments = []string{".moai", "state", "kanban-board"}

// boardStateFileName names the board state document inside the board
// directory. The role-declaration carrier lives in a sibling subdirectory, so
// recovery's extent (the state file alone) never touches declarations.
const boardStateFileName = "board.json"

// ErrBoardUnknown reports that the board state file exists but cannot be read
// or parsed (REQ-KB-013's detection half). The board is reported as unknown
// and every dispatch is refused — an empty board is NEVER presented as an
// accurate one, because that would admit cards past a WIP limit whose
// contents are unknown.
var ErrBoardUnknown = errors.New("kanban board state unknown")

// IsBoardUnknown reports whether err is the unknown-board sentinel.
func IsBoardUnknown(err error) bool {
	return errors.Is(err, ErrBoardUnknown)
}

// Card is one board card's persisted record. The column is a RECORDED value
// (REQ-KB-006): the board reads it and computes nothing — no helper recovers
// or derives a column from a status, a progress marker, or any other
// observable. M2 completes the model (closed column enumeration, consistency
// table); M1 carries the recorded field with no derivation path anywhere.
//
// Holder is empty for an unheld card rather than a synthesized value; the
// last-transition instant is RFC3339.
type Card struct {
	SpecID      string `json:"spec_id"`
	Column      string `json:"column"`
	Holder      string `json:"holder,omitempty"`
	LastMovedAt string `json:"last_transition_at"`
}

// @MX:ANCHOR: [AUTO] Board/Card state schema — the persisted board contract every session reads and the sole writer writes
// @MX:REASON: the JSON keys bind the board store, the sole-writer guard, the recovery sidecar expectations, and M2's card model; a renamed key breaks readers this package cannot see
//
// BoardState is the JSON document persisted at the board path. An empty Cards
// slice is a valid, accurate board only when it was read from an absent state
// file (REQ-KB-021) or written as such by the sole writer — never when the
// prior file was unreadable.
type BoardState struct {
	Cards []Card `json:"cards"`
}

// BoardRoot resolves the primary checkout's root from anywhere in the
// repository: the parent of the ABSOLUTE common git directory obtained
// through the extracted resolver (REQ-KB-005). The bare --git-common-dir
// form is never used alone — it returns a relative .git in the primary
// checkout, and "the parent of it" would not be a path.
func BoardRoot(startDir string) (string, error) {
	dirs, err := gitcore.ResolveGitDirs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve board root: %w", err)
	}
	return filepath.Dir(dirs.CommonDir), nil
}

// BoardDir returns the board state directory beneath root.
func BoardDir(root string) string {
	return filepath.Join(append([]string{root}, boardDirSegments...)...)
}

// BoardPath returns the board state file path beneath root.
func BoardPath(root string) string {
	return filepath.Join(BoardDir(root), boardStateFileName)
}

// LoadBoard reads the board state beneath root.
//
// An ABSENT state file is a legitimately empty board (REQ-KB-021): dispatch
// is permitted and no recovery is required. An UNREADABLE or unparseable one
// is the unknown board (REQ-KB-013): ErrBoardUnknown, no state returned. The
// read path performs no repair — recovery is an explicit operator-visible
// act, never a fallback a read takes (AP-21).
func LoadBoard(root string) (*BoardState, error) {
	raw, err := os.ReadFile(BoardPath(root))
	if err != nil {
		if os.IsNotExist(err) {
			// Absent: the operating system's own answer that the file does
			// not exist — distinguishable without judgment from unreadable,
			// which is what makes assigning them different behavior safe.
			return &BoardState{Cards: []Card{}}, nil
		}
		return nil, fmt.Errorf("%w: reading board state: %v", ErrBoardUnknown, err)
	}

	var st BoardState
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, fmt.Errorf("%w: parsing board state: %v", ErrBoardUnknown, err)
	}
	if st.Cards == nil {
		st.Cards = []Card{}
	}
	return &st, nil
}
