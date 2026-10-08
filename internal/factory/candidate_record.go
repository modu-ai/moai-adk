// candidate_record.go — the candidate-CI record and its store (card t1478,
// SPEC-CANDIDATE-CI-001 REQ-CCI-005).
//
// A candidate record is the per-card state one card's candidate push leaves
// behind: keyed by (card id, pinned SHA), it carries what the landing check
// (REQ-CCI-011) and the acquire precondition (REQ-CCI-012) later judge —
// the candidate commit, the integration branch and tip it was built
// against, the candidate branch, the CI run identity once one is observed,
// the verdict, and the push/observation timestamps.
//
// The store lives under the PRIMARY checkout's .moai/state, the same
// visibility contract the integration lock record has: a verdict recorded
// in one worktree must be readable from the tree that later merges. The
// red-candidate hold (REQ-CCI-012, design.md D10) is exactly this record —
// per-card by construction — and never the shared integration window
// policy, which stays card-blind on purpose.
package factory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Candidate verdict values (REQ-CCI-005/010). Pending is the push-time
// state; only an explicit observation of a completed CI run moves it — the
// landing check treats pending as not-green, fail-closed.
const (
	CandidateVerdictPending = "pending"
	CandidateVerdictGreen   = "green"
	CandidateVerdictRed     = "red"
)

// ErrCandidateRecordAbsent marks an absent key. Absence is a DIFFERENT
// state from a store error: the landing check refuses on absence with its
// own cause, while a read failure is an error to report.
var ErrCandidateRecordAbsent = errors.New("no candidate record")

// CandidateRecord is one card's candidate state, keyed by (CardID,
// PinnedSHA). A re-candidate writes a NEW record under the new pinned SHA —
// prior records stay readable but superseded (REQ-CCI-003's replace-in-place
// is the BRANCH's, never the history's).
type CandidateRecord struct {
	CardID string `json:"card_id"`
	// PinnedSHA is the key half: the card branch commit the candidate was
	// built from — the same SHA the merge step pins (REQ-CCI-004).
	PinnedSHA string `json:"pinned_sha"`
	// CandidateSHA is the two-parent candidate commit the push carried.
	CandidateSHA string `json:"candidate_sha"`
	// IntegrationBranch and IntegrationTip name what the candidate was
	// built against; a tip that has since moved is what makes the record
	// stale for a merge (the landing check's ancestry re-verification).
	IntegrationBranch string `json:"integration_branch"`
	IntegrationTip    string `json:"integration_tip"`
	// CandidateBranch is the remote ref the run judged (ci/<card>).
	CandidateBranch string `json:"candidate_branch"`
	// RunID is the CI run identity, recorded by the verdict observation —
	// empty until one runs (REQ-CCI-010).
	RunID string `json:"run_id,omitempty"`
	// Verdict is pending until an explicit observation records green or red.
	Verdict string `json:"verdict"`
	// PushedAt is the push timestamp; ObservedAt the verdict observation's.
	PushedAt   string `json:"pushed_at"`
	ObservedAt string `json:"observed_at,omitempty"`
}

// candidateDir resolves the store under the project's state directory.
func candidateDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".moai", "state", "candidate")
}

// candidateRecordPath resolves one record's file. The card id and SHA are
// path-segment guards, not trust: a caller that bypasses the verb's card-id
// validation must not be able to write outside the store.
func candidateRecordPath(projectRoot, cardID, pinnedSHA string) (string, error) {
	if err := validCandidateKeyPart(cardID); err != nil {
		return "", fmt.Errorf("candidate record: card id: %w", err)
	}
	if err := validCandidateKeyPart(pinnedSHA); err != nil {
		return "", fmt.Errorf("candidate record: pinned sha: %w", err)
	}
	return filepath.Join(candidateDir(projectRoot), cardID, pinnedSHA+".json"), nil
}

// validCandidateKeyPart refuses anything that is not a single safe path
// segment.
func validCandidateKeyPart(part string) error {
	if strings.TrimSpace(part) == "" {
		return errors.New("empty")
	}
	if part != filepath.Base(part) || part == "." || part == ".." || strings.ContainsAny(part, `/\`) {
		return fmt.Errorf("%q is not a single path segment", part)
	}
	return nil
}

// WriteCandidateRecord stores the record keyed by its (card, pinned SHA).
// The write is atomic (temp + rename), so a reader never sees a partial
// record and two racing writers leave a whole file.
func WriteCandidateRecord(projectRoot string, rec CandidateRecord) error {
	path, err := candidateRecordPath(projectRoot, rec.CardID, rec.PinnedSHA)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("candidate store: %w", err)
	}
	if err := atomicWriteFile(path, rec); err != nil {
		return fmt.Errorf("candidate store: %w", err)
	}
	return nil
}

// ReadCandidateRecord returns the record keyed by (cardID, pinnedSHA).
// Absence reads as ErrCandidateRecordAbsent; an unreadable record is a
// plain error — the two states must not collapse, because "no candidate
// was ever pushed" and "the record is corrupt" refuse with different
// causes.
func ReadCandidateRecord(projectRoot, cardID, pinnedSHA string) (*CandidateRecord, error) {
	path, err := candidateRecordPath(projectRoot, cardID, pinnedSHA)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w for card %s at pinned %s", ErrCandidateRecordAbsent, cardID, pinnedSHA)
	}
	if err != nil {
		return nil, fmt.Errorf("read candidate record: %w", err)
	}
	var rec CandidateRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("candidate record for card %s at pinned %s is unreadable: %w", cardID, pinnedSHA, err)
	}
	// The key fields read AS WRITTEN, never forced to the lookup key: the
	// landing check compares the record's own identity claims, and forcing
	// them here would turn those comparisons into dead code.
	return &rec, nil
}
