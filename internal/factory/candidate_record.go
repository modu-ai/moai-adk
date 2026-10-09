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
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	// Seq is the card's push sequence — allocated inside the per-card
	// push critical section (NextCandidateSequence), it orders same-second
	// pushes deterministically where PushedAt (second granularity) ties
	// (card t1478 M4 observation-path repair).
	Seq int64 `json:"seq,omitempty"`
	// RunID is the CI run identity, recorded by the verdict observation —
	// empty until one runs (REQ-CCI-010).
	RunID string `json:"run_id,omitempty"`
	// RunAttempt is the attempt of RunID the observation recorded (card t1478
	// Finding 2): a re-run keeps its run id and carries a higher attempt. An
	// absent attempt — a record written before attempts were recorded — reads
	// as attempt 1 of its recorded run.
	RunAttempt int `json:"run_attempt,omitempty"`
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
//
// The store is CONFINED (card t1478 M2 repair): every path component from
// the state directory down to the card directory is Lstat-verified not a
// symlink before anything is created — a `.moai/state/candidate/<card>`
// swapped for a link to an external directory would otherwise be followed
// by MkdirAll (which succeeds on an existing link) and overwrite an
// external <pinnedSHA>.json.
func WriteCandidateRecord(projectRoot string, rec CandidateRecord) error {
	path, err := candidateRecordPath(projectRoot, rec.CardID, rec.PinnedSHA)
	if err != nil {
		return err
	}
	if err := candidateStoreRealPath(projectRoot, rec.CardID); err != nil {
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

// candidateStoreRealPath refuses a symlink anywhere along the store's
// directory chain (.moai → state → candidate → <card>). Missing components
// are fine — MkdirAll creates them as real directories; an existing
// component that is a link would redirect every write beneath it. The
// archiveThenRemove ensureNoSymlinkPath precedent walks the same way.
func candidateStoreRealPath(projectRoot, cardID string) error {
	parts := []string{".moai", "state", "candidate", cardID}
	cur := projectRoot
	for _, part := range parts {
		cur = filepath.Join(cur, part)
		info, err := os.Lstat(cur)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("candidate store: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("candidate store: %s traverses a symlink — the store is confined to %s", cur, candidateDir(projectRoot))
		}
	}
	return nil
}

// ReadCandidateRecord returns the record keyed by (cardID, pinnedSHA).
// Absence reads as ErrCandidateRecordAbsent; an unreadable record is a
// plain error — the two states must not collapse, because "no candidate
// was ever pushed" and "the record is corrupt" refuse with different
// causes.
//
// Non-regular files refuse WITHOUT being read (card t1478 M2 repair): a
// FIFO swapped in at the record path parked the former plain os.ReadFile
// inside the candidate mutation lock, wedging every later candidate call
// of the card past every deadline. openCandidateRecordFile opens with
// non-blocking, no-follow flags and verifies the OPENED descriptor is a
// regular file — the Lstat precheck alone left a TOCTOU window where a
// regular file swapped for a FIFO after the check still parked the read.
func ReadCandidateRecord(projectRoot, cardID, pinnedSHA string) (*CandidateRecord, error) {
	path, err := candidateRecordPath(projectRoot, cardID, pinnedSHA)
	if err != nil {
		return nil, err
	}
	f, err := openCandidateRecordFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w for card %s at pinned %s", ErrCandidateRecordAbsent, cardID, pinnedSHA)
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
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

// LatestCandidateRecord returns the card's newest record — by push
// sequence first (allocated in the push critical section, so same-second
// pushes order deterministically), PushedAt as the tiebreak for records
// written before sequences existed. What an observation (REQ-CCI-010) and
// the acquire precondition (REQ-CCI-012) read when no pinned SHA is at
// hand. Absence reads as ErrCandidateRecordAbsent.
//
// Each entry opens through the non-blocking, regular-verified open — a
// `.json`-named FIFO in the scan parked --observe and the acquire
// precondition inside the mutation lock (card t1478 M4 repair).
func LatestCandidateRecord(projectRoot, cardID string) (*CandidateRecord, error) {
	if err := validCandidateKeyPart(cardID); err != nil {
		return nil, fmt.Errorf("candidate store: card id: %w", err)
	}
	dir := filepath.Join(candidateDir(projectRoot), cardID)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w for card %s", ErrCandidateRecordAbsent, cardID)
	}
	if err != nil {
		return nil, fmt.Errorf("read candidate store: %w", err)
	}
	var latest *CandidateRecord
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		f, err := openCandidateRecordFile(path)
		if err != nil {
			continue
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			continue
		}
		var rec CandidateRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		rec.PinnedSHA = strings.TrimSuffix(entry.Name(), ".json")
		if latest == nil || candidateNewer(&rec, latest) {
			latest = &rec
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("%w for card %s", ErrCandidateRecordAbsent, cardID)
	}
	return latest, nil
}

// candidateNewer orders two records of one card: higher sequence wins;
// without sequences (or on a tie) the later PushedAt wins.
func candidateNewer(a, b *CandidateRecord) bool {
	if a.Seq != b.Seq {
		return a.Seq > b.Seq
	}
	return a.PushedAt > b.PushedAt
}

// NextCandidateSequence returns the card's next push sequence — max
// existing + 1. It MUST be called inside the card's mutation critical
// section (the push path already holds it), so two racing pushes cannot
// allocate the same number.
func NextCandidateSequence(projectRoot, cardID string) (int64, error) {
	latest, err := LatestCandidateRecord(projectRoot, cardID)
	if errors.Is(err, ErrCandidateRecordAbsent) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	return latest.Seq + 1, nil
}

// integrationCandidateMutationHook is the observation write's test seam —
// called inside the per-card critical section before the write (the
// integrationLockMutationTestHook precedent), so a test can present the
// re-candidate-replaced record the re-read must catch.
var integrationCandidateMutationHook func()

// CandidateRunState is what one observed CI run reports — the scripted
// double the tests feed and the gh read maps into.
type CandidateRunState struct {
	RunID      string
	HeadSHA    string
	Ref        string
	Status     string // queued | in_progress | completed
	Conclusion string // success | failure | ... (meaningful only on completed)
	// Attempt is the run attempt (card t1478 Finding 2); 0 reads as attempt 1.
	Attempt int
}

// ObserveCandidateVerdict applies one observed run to the record keyed
// (cardID, pinnedSHA) and returns the (possibly) updated record and
// whether anything was written (REQ-CCI-010).
//
// THE VERDICT-SHA BINDING: a run's verdict may be recorded for a candidate
// only when the run's head SHA equals the record's candidate commit SHA
// AND the run's ref equals the record's candidate branch. A completed run
// matching neither — a late-arriving green for a superseded candidate on
// the same ci/<card> ref — is DISCARDED: nothing is written and the
// record's verdict stands exactly as it was. A run still in progress is
// not a verdict either. The mutation-lock the write shares with the push
// path keeps an observation from interleaving with a re-candidate.
//
// EXECUTION ORDER (card t1478 Finding 2): under the lock, a verdict also
// requires the run to be at least as new as the run the record carries — a
// higher run id, or the same id at an equal or higher attempt. A stale
// observation of an older run never overwrites a newer verdict.
func ObserveCandidateVerdict(projectRoot, cardID, pinnedSHA string, run CandidateRunState, now time.Time) (CandidateRecord, bool, error) {
	rec, err := ReadCandidateRecord(projectRoot, cardID, pinnedSHA)
	if err != nil {
		return CandidateRecord{}, false, err
	}
	if run.HeadSHA != rec.CandidateSHA || run.Ref != rec.CandidateBranch {
		// The binding refuses silently-but-truly: discard is the CONTRACT
		// (wrote=false), not an error — the caller asked about a run that
		// is not this candidate's.
		return *rec, false, nil
	}
	if run.Status != "completed" {
		return *rec, false, nil
	}
	var verdict string
	switch run.Conclusion {
	case "success":
		verdict = CandidateVerdictGreen
	case "failure":
		verdict = CandidateVerdictRed
	default:
		// cancelled / timed out / skipped: no verdict either way — the
		// CI run did not judge the tree.
		return *rec, false, nil
	}
	observedAt := now.UTC().Format(time.RFC3339)
	var written CandidateRecord
	writeErr := WithCandidateMutation(projectRoot, cardID, func() error {
		if integrationCandidateMutationHook != nil {
			integrationCandidateMutationHook()
		}
		// Re-read under the lock: the outer read and the gh runs list were
		// taken OUTSIDE the section — a re-candidate replacing the record
		// in between is caught here, and only a record whose candidate SHA
		// STILL matches the run writes.
		current, err := ReadCandidateRecord(projectRoot, cardID, pinnedSHA)
		if err != nil {
			return err
		}
		if current.CandidateSHA != run.HeadSHA || current.CandidateBranch != run.Ref {
			return nil
		}
		// Execution order (card t1478 Finding 2): the run is judged against
		// the run the RE-READ record already carries — an older run, or an
		// older attempt of the same run, is refused without a write.
		admitted, orderErr := candidateRunAdmitted(current, run)
		if orderErr != nil {
			return orderErr
		}
		if !admitted {
			return nil
		}
		// The verdict fields refresh onto the RE-READ record — a re-push
		// of the same candidate SHA that landed in between keeps its push
		// info (PushedAt, sequence); only verdict, run id and attempt, and
		// observation time are the observation's to write (card t1478 M4
		// repair).
		current.Verdict = verdict
		current.RunID = run.RunID
		current.RunAttempt = candidateAttempt(run.Attempt)
		current.ObservedAt = observedAt
		if err := WriteCandidateRecord(projectRoot, *current); err != nil {
			return err
		}
		written = *current
		return nil
	})
	if writeErr != nil {
		return CandidateRecord{}, false, fmt.Errorf("observe candidate verdict: %w", writeErr)
	}
	if written.Verdict == "" {
		return *rec, false, nil
	}
	return written, true, nil
}

// candidateRunAdmitted decides, under the candidate lock, whether the observed
// run is at least as new as the run the record already carries (card t1478
// Finding 2). GitHub run ids increase with execution and a re-run keeps its id
// with a higher attempt, so the order is the numeric id, then the attempt. A
// record with no recorded run admits the first observation. An empty or
// non-numeric run id has no order: it is refused with an error, so it can never
// be recorded green (fail closed), and a recorded run that cannot be ordered
// refuses the observation the same way.
func candidateRunAdmitted(current *CandidateRecord, run CandidateRunState) (bool, error) {
	observed, err := parseCandidateRunID(run.RunID)
	if err != nil {
		return false, fmt.Errorf("observe candidate verdict: refused — %w; an unorderable run never reads green", err)
	}
	if current.RunID == "" {
		return true, nil
	}
	recorded, err := parseCandidateRunID(current.RunID)
	if err != nil {
		return false, fmt.Errorf("observe candidate verdict: refused — the recorded run cannot be ordered against this observation: %w", err)
	}
	switch {
	case observed > recorded:
		return true, nil
	case observed < recorded:
		return false, nil
	default:
		return candidateAttempt(run.Attempt) >= candidateAttempt(current.RunAttempt), nil
	}
}

// parseCandidateRunID reads a run identity as its numeric execution order.
// Empty, signed, padded, and alphabetic ids have no order.
func parseCandidateRunID(runID string) (uint64, error) {
	id, err := strconv.ParseUint(runID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("run id %q is not numeric", runID)
	}
	return id, nil
}

// candidateAttempt normalizes a run attempt: anything below 1 — an absent
// field, or a record written before attempts were recorded — reads as attempt
// 1 of its run.
func candidateAttempt(attempt int) int {
	if attempt < 1 {
		return 1
	}
	return attempt
}
