package guardliveness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// storeSubdir is the persistence directory under the ~/.moai state tree.
const storeSubdir = "guard-liveness"

// Per-root filename suffixes. Two things are persisted about a root and they
// answer different questions: what the mechanism last MEASURED, and what the
// reader was last SHOWN.
const (
	resultSuffix   = ".json"
	renderedSuffix = ".rendered.json"
)

// ErrNoPersistedResult reports that no evaluation result has been persisted for
// a root yet. It is deliberately distinct from an all-clear: before the first
// refresh completes there is nothing to report ABOUT, which is not the same as
// having looked and found nothing (spec.md §A.0).
var ErrNoPersistedResult = errors.New("guard liveness: no evaluation result has been persisted for this root")

// Snapshot is one persisted evaluation, carrying the moment it was taken.
//
// The timestamp is the measurement's own, recorded when the refresh completed.
// The advisory derives the age it reports from THIS field and never from the
// render moment, which is what lets a stale advisory declare its own staleness
// rather than reading as a current all-clear (REQ-GDL-006).
type Snapshot struct {
	TakenAt time.Time `json:"taken_at"`
	Result  Result    `json:"result"`
}

// Store persists evaluation results between activations.
//
// The render and the refresh are separate acts because the host surface is
// latency-bounded and the refresh issues one query per subject (spec.md §B.3).
// This is what carries a result from one to the other: the refresh writes here
// when it completes, and a LATER activation's render reads it.
//
// The directory lives outside every evaluated working tree (REQ-GDL-008). A
// cache written into the tree would show up as drift for the next reader, and
// the advisory path must leave the tree byte-identical.
type Store struct {
	dir string
}

// NewStore returns a store rooted at dir.
func NewStore(dir string) *Store { return &Store{dir: dir} }

// DefaultStore returns the store under the user's ~/.moai state tree.
func DefaultStore() (*Store, error) {
	state, err := paths.StateDir()
	if err != nil {
		return nil, fmt.Errorf("guard liveness: resolve the state directory: %w", err)
	}
	return NewStore(filepath.Join(state, storeSubdir)), nil
}

// Save records a result together with the moment it was taken.
//
// The write is atomic — a temporary file renamed into place — so a render
// racing a refresh reads either the previous result or the new one, never a
// half-written file that would parse as a malformed result and be reported as
// a contract violation that never happened.
func (s *Store) Save(root string, r Result, takenAt time.Time) error {
	return s.write(s.pathFor(root, resultSuffix), Snapshot{TakenAt: takenAt, Result: r})
}

// SaveRendered records what this render announced, for the NEXT render to diff
// against (REQ-GDL-007). It is a separate file from the persisted result: one
// is what the mechanism measured, the other is what the reader was shown, and a
// render that overwrote the first with the second would hand the next
// activation a measurement that never happened.
//
// This is a write, and it deliberately lands in the same directory as the
// result — outside every evaluated working tree (REQ-GDL-008). The render
// leaves the tree byte-identical; its memory lives elsewhere.
func (s *Store) SaveRendered(root string, rec RenderRecord) error {
	return s.write(s.pathFor(root, renderedSuffix), rec)
}

// LoadRendered returns what the previous render announced for a root. Nothing
// written yet is an empty record rather than an error: before any render every
// non-clean entry is a change, which is what a reader who has been shown
// nothing needs.
func (s *Store) LoadRendered(root string) (RenderRecord, error) {
	payload, err := os.ReadFile(s.pathFor(root, renderedSuffix))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RenderRecord{}, nil
		}
		return RenderRecord{}, fmt.Errorf("guard liveness: read the render record: %w", err)
	}

	var rec RenderRecord
	if err := json.Unmarshal(payload, &rec); err != nil {
		return RenderRecord{}, fmt.Errorf("guard liveness: decode the render record: %w", err)
	}
	return rec, nil
}

// write places a value at path atomically.
func (s *Store) write(final string, v any) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("guard liveness: create the persistence directory: %w", err)
	}

	payload, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("guard liveness: encode the result: %w", err)
	}

	tmp, err := os.CreateTemp(s.dir, "snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("guard liveness: open a temporary file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("guard liveness: write the result: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("guard liveness: close the temporary file: %w", err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("guard liveness: place the result: %w", err)
	}
	return nil
}

// Load returns the most recently persisted snapshot for a root, or
// ErrNoPersistedResult when none has been written.
func (s *Store) Load(root string) (Snapshot, error) {
	payload, err := os.ReadFile(s.pathFor(root, resultSuffix))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Snapshot{}, ErrNoPersistedResult
		}
		return Snapshot{}, fmt.Errorf("guard liveness: read the persisted result: %w", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(payload, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("guard liveness: decode the persisted result: %w", err)
	}
	return snap, nil
}

// pathFor keys a root by a digest of its path, with a suffix separating the two
// things persisted per root. One store serves every tree on the machine, and a
// digest keeps two trees apart without embedding a filesystem path — which
// would not survive as a filename — in the key.
func (s *Store) pathFor(root, suffix string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(root)))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])[:32]+suffix)
}
