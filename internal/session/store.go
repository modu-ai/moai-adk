package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// ErrCheckpointStale is returned when a checkpoint is older than the stale TTL.
	ErrCheckpointStale = errors.New("checkpoint is stale")
	// ErrBlockerOutstanding is returned when attempting to advance phase with unresolved blocker.
	ErrBlockerOutstanding = errors.New("blocker is outstanding, cannot advance phase")
	// ErrCheckpointInvalid is returned when checkpoint validation fails.
	ErrCheckpointInvalid = errors.New("checkpoint validation failed")
	// ErrCheckpointConcurrent is returned when concurrent checkpoint writes are detected.
	ErrCheckpointConcurrent = errors.New("concurrent checkpoint write conflict")
	// ErrArtifactEncodingMismatch is returned when a text artifact contains invalid UTF-8.
	// SPEC-V3R2-RT-004 REQ-043: WriteRunArtifact UTF-8 validation.
	ErrArtifactEncodingMismatch = errors.New("artifact encoding mismatch: text artifact contains invalid UTF-8")
)

// HydrateOpts controls optional behavior of HydrateWithOpts.
// SPEC-V3R2-RT-004 REQ-033: options tied to the --resume flag.
type HydrateOpts struct {
	// Setting SkipStaleCheck to true skips the staleness check.
	// WARN: loading a stale checkpoint can cause data inconsistency.
	// Use this only with the --resume flag.
	SkipStaleCheck bool
}

// SessionStore manages phase state persistence to .moai/state/.
type SessionStore interface {
	// Checkpoint persists the phase state to disk.
	Checkpoint(state PhaseState) error
	// Hydrate loads the phase state from disk.
	Hydrate(phase Phase, specID string) (*PhaseState, error)
	// HydrateWithOpts loads phase state from disk with optional behavior control.
	// SPEC-V3R2-RT-004 REQ-033: integrates with the --resume flag via SkipStaleCheck.
	HydrateWithOpts(phase Phase, specID string, opts HydrateOpts) (*PhaseState, error)
	// AppendTaskLedger adds an entry to the task ledger.
	AppendTaskLedger(entry TaskLedgerEntry) error
	// WriteRunArtifact writes an artifact file for a run iteration.
	WriteRunArtifact(iterID, name string, body []byte) error
	// RecordBlocker persists a blocker report.
	RecordBlocker(report BlockerReport) error
	// ResolveBlocker marks a blocker as resolved.
	ResolveBlocker(phase Phase, specID string, resolution string) error
	// DetectInFlightTransition scans state dir and returns the most recent phase
	// transition that has started but not completed for the given specID.
	// SPEC-V3R2-RT-004 REQ-050: in-flight transition detection.
	DetectInFlightTransition(specID string) (fromPhase Phase, toPhase Phase, found bool, err error)
	// MergeTeamCheckpoints reads per-agent checkpoint files and returns a merged PhaseState.
	// SPEC-V3R2-RT-004 REQ-021, REQ-051: team-mode per-agent checkpoint merge.
	MergeTeamCheckpoints(specID string, phase Phase, agentNames []string) (*PhaseState, error)
}

// FileSessionStore implements SessionStore using .moai/state/ directory.
type FileSessionStore struct {
	stateDir string        // .moai/state/
	staleTTL time.Duration // staleness threshold
}

// NewFileSessionStore creates a new FileSessionStore.
func NewFileSessionStore(stateDir string, staleTTL time.Duration) *FileSessionStore {
	return &FileSessionStore{
		stateDir: stateDir,
		staleTTL: staleTTL,
	}
}

// Checkpoint persists the phase state to disk with atomic write.
// SPEC-V3R2-RT-004 REQ-040: prevents concurrent writes via advisory lock (3-retry / 10ms-backoff).
// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-004 REQ-002/004/010/020/040 enforcer — every phase boundary writes through here
// @MX:REASON: Validator + lock + blocker-file scan order is contract; touching this affects all 9 phases
func (fs *FileSessionStore) Checkpoint(state PhaseState) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Inline BlockerRpt check
	if state.BlockerRpt != nil && !state.BlockerRpt.Resolved {
		return ErrBlockerOutstanding
	}

	// SPEC-V3R2-RT-004 AC-04: scan blocker files on disk (keyed by phase+specID)
	if err := fs.checkBlockerFiles(state.Phase, state.SPECID); err != nil {
		return err
	}

	// SPEC-V3R2-RT-004 REQ-004: Validate checkpoint before write
	if state.Checkpoint != nil {
		if err := state.Checkpoint.Validate(); err != nil {
			return fmt.Errorf("%w: %v", ErrCheckpointInvalid, err)
		}
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	filename := fs.checkpointPath(state.Phase, state.SPECID)
	tmpFile := filename + ".tmp"

	// SPEC-V3R2-RT-004 REQ-040: Acquire advisory lock (3-retry / 10ms-backoff)
	lock := newFileLock()
	if err := acquireWithRetry(lock, filename, 3, 10*time.Millisecond); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer func() {
		_ = lock.release() // Ignore lock release failures (checkpoint already written)
	}()

	// Write to temporary file
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, filename); err != nil {
		_ = os.Remove(tmpFile) // Clean up on failure
		return fmt.Errorf("atomic rename: %w", err)
	}

	return nil
}

// Hydrate loads the phase state from disk and checks for staleness.
// @MX:WARN: [AUTO] STALE_SECONDS gate — bypassing staleness check defeats crash-resume safety
// @MX:REASON: Use HydrateWithOpts(SkipStaleCheck: true) for --resume, which logs a stderr warning
func (fs *FileSessionStore) Hydrate(phase Phase, specID string) (*PhaseState, error) {
	return fs.HydrateWithOpts(phase, specID, HydrateOpts{})
}

// HydrateWithOpts loads the phase state from disk with optional behavior control.
// SPEC-V3R2-RT-004 REQ-033: integrates with the --resume flag via SkipStaleCheck.
func (fs *FileSessionStore) HydrateWithOpts(phase Phase, specID string, opts HydrateOpts) (*PhaseState, error) {
	filename := fs.checkpointPath(phase, specID)

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no checkpoint
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var state PhaseState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	// SPEC-V3R2-RT-004 REQ-004: validate after read (AC-09 corrupted checkpoint)
	if state.Checkpoint != nil {
		if err := state.Checkpoint.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCheckpointInvalid, err)
		}
	}

	// Staleness check (bypassed when SkipStaleCheck=true; emits a WARN on stderr)
	if !opts.SkipStaleCheck {
		if time.Since(state.UpdatedAt) > fs.staleTTL {
			return nil, ErrCheckpointStale
		}
	} else if time.Since(state.UpdatedAt) > fs.staleTTL {
		// Emit a stale warning when --resume is used
		fmt.Fprintf(os.Stderr, "WARN: checkpoint for %s/%s is stale (age=%v); loading anyway (--resume)\n",
			phase, specID, time.Since(state.UpdatedAt).Round(time.Second))
	}

	return &state, nil
}

// DetectInFlightTransition scans .moai/state/ for a SPECID where a phase checkpoint
// exists but the next phase checkpoint does not — indicating a suspended transition.
// SPEC-V3R2-RT-004 REQ-050: in-flight transition detection.
func (fs *FileSessionStore) DetectInFlightTransition(specID string) (Phase, Phase, bool, error) {
	// Phase order: plan → run → sync
	phaseOrder := []Phase{PhasePlan, PhaseRun, PhaseSync}

	// Check whether each phase has a checkpoint
	present := make(map[Phase]bool)
	for _, ph := range phaseOrder {
		path := fs.checkpointPath(ph, specID)
		if _, err := os.Stat(path); err == nil {
			present[ph] = true
		}
	}

	if len(present) == 0 {
		return "", "", false, nil
	}

	// Find the last completed phase; if the next phase is missing, the transition is in-flight
	for i, ph := range phaseOrder {
		if present[ph] && i+1 < len(phaseOrder) {
			nextPh := phaseOrder[i+1]
			if !present[nextPh] {
				return ph, nextPh, true, nil
			}
		}
	}

	return "", "", false, nil
}

// MergeTeamCheckpoints reads per-agent checkpoint files and merges them into one PhaseState.
// File path: checkpoint-{phase}-{specID}-{agentName}.json
// SPEC-V3R2-RT-004 REQ-021, REQ-051: team-mode merge + blocker bubble-mode.
func (fs *FileSessionStore) MergeTeamCheckpoints(specID string, phase Phase, agentNames []string) (*PhaseState, error) {
	if len(agentNames) == 0 {
		return nil, fmt.Errorf("agentNames must not be empty")
	}

	// Sort agent names alphabetically (deterministic merge)
	sorted := make([]string, len(agentNames))
	copy(sorted, agentNames)
	sort.Strings(sorted)

	var states []PhaseState
	var auditPaths []string

	for _, agent := range sorted {
		agentPath := filepath.Join(fs.stateDir, fmt.Sprintf("checkpoint-%s-%s-%s.json", phase, specID, agent))
		auditPaths = append(auditPaths, agentPath)

		data, err := os.ReadFile(agentPath)
		if err != nil {
			return nil, fmt.Errorf("read agent checkpoint %s: %w", agentPath, err)
		}

		var state PhaseState
		if err := json.Unmarshal(data, &state); err != nil {
			return nil, fmt.Errorf("unmarshal agent checkpoint %s: %w", agentPath, err)
		}

		// REQ-051: blocker bubble-mode — return an error immediately if there is an unresolved blocker
		if state.BlockerRpt != nil && !state.BlockerRpt.Resolved {
			return nil, ErrBlockerOutstanding
		}

		states = append(states, state)
	}

	// Phase-specific union merge
	merged, err := mergePhaseStates(phase, specID, states)
	if err != nil {
		return nil, fmt.Errorf("merge phase states: %w", err)
	}

	// Set provenance
	merged.Provenance = ProvenanceTag{
		Source: "session",
		Origin: strings.Join(auditPaths, ","),
		Loaded: time.Now(),
	}

	return merged, nil
}

// mergePhaseStates applies the per-phase union-merge rules.
func mergePhaseStates(phase Phase, specID string, states []PhaseState) (*PhaseState, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states to merge")
	}

	merged := &PhaseState{
		Phase:     phase,
		SPECID:    specID,
		UpdatedAt: time.Now(),
	}

	switch phase {
	case PhaseRun:
		// RunCheckpoint: sum TestsTotal/TestsPassed/FilesModified
		var total, passed, modified int
		var status, harness string
		for _, s := range states {
			if rc, ok := s.Checkpoint.(*RunCheckpoint); ok {
				total += rc.TestsTotal
				passed += rc.TestsPassed
				modified += rc.FilesModified
				if status == "" {
					status = rc.Status
				}
				if harness == "" {
					harness = rc.Harness
				}
			}
		}
		merged.Checkpoint = &RunCheckpoint{
			SPECID:        specID,
			Status:        status,
			Harness:       harness,
			TestsTotal:    total,
			TestsPassed:   passed,
			FilesModified: modified,
		}
	case PhasePlan:
		// PlanCheckpoint: use the last state (assumes a single plan agent)
		if len(states) > 0 {
			merged.Checkpoint = states[len(states)-1].Checkpoint
		}
	case PhaseSync:
		// SyncCheckpoint: use the last state
		if len(states) > 0 {
			merged.Checkpoint = states[len(states)-1].Checkpoint
		}
	default:
		// Unknown phase: use the last state
		if len(states) > 0 {
			merged.Checkpoint = states[len(states)-1].Checkpoint
		}
	}

	return merged, nil
}

// checkBlockerFiles scans for unresolved blocker files matching the given phase+specID.
// SPEC-V3R2-RT-004 AC-04: disk-file-based blocker check (not inline ref).
func (fs *FileSessionStore) checkBlockerFiles(phase Phase, specID string) error {
	// Scan the blocker-{phase}-{specID}-*.json pattern
	pattern := filepath.Join(fs.stateDir, fmt.Sprintf("blocker-%s-%s-*.json", phase, specID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob blocker files: %w", err)
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue // ignore read failure
		}

		var blocker BlockerReport
		if err := json.Unmarshal(data, &blocker); err != nil {
			continue // ignore parse failure
		}

		if !blocker.Resolved {
			return ErrBlockerOutstanding
		}
	}

	return nil
}

// AppendTaskLedger adds an entry to the task ledger markdown file.
func (fs *FileSessionStore) AppendTaskLedger(entry TaskLedgerEntry) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	ledgerPath := filepath.Join(fs.stateDir, "task-ledger.md")

	// Open file in append mode, create if not exists
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(entry.ToMarkdown()); err != nil {
		return fmt.Errorf("write entry: %w", err)
	}

	return nil
}

// textExtensions is the list of text file extensions that require UTF-8 validation.
// SPEC-V3R2-RT-004 REQ-043: text-declared artifacts must be UTF-8.
// @MX:WARN: [AUTO] SPEC-V3R2-RT-004 REQ-043 ArtifactEncodingMismatch
// @MX:REASON: Text-declared artifacts (.md|.txt|.json|.yaml) MUST be UTF-8.
// Bypassing this corrupts moai state dump output and breaks Ralph iteration parsing.
var textExtensions = map[string]bool{
	".md":   true,
	".txt":  true,
	".json": true,
	".yaml": true,
	".yml":  true,
}

// WriteRunArtifact writes an artifact file for a run iteration.
// SPEC-V3R2-RT-004 REQ-043: text-extension files are written only after UTF-8 validation.
func (fs *FileSessionStore) WriteRunArtifact(iterID, name string, body []byte) error {
	artifactDir := filepath.Join(fs.stateDir, "runs", iterID)
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("create artifact dir: %w", err)
	}

	// SPEC-V3R2-RT-004 REQ-043: validate UTF-8 for text file extensions
	ext := strings.ToLower(filepath.Ext(name))
	if textExtensions[ext] {
		if !utf8.Valid(body) {
			return ErrArtifactEncodingMismatch
		}
	}

	artifactPath := filepath.Join(artifactDir, name)

	tmpFile := artifactPath + ".tmp"
	if err := os.WriteFile(tmpFile, body, 0644); err != nil {
		return fmt.Errorf("write artifact: %w", err)
	}

	if err := os.Rename(tmpFile, artifactPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename artifact: %w", err)
	}

	return nil
}

// RecordBlocker persists a blocker report to disk.
// SPEC-V3R2-RT-004 REQ-012: filename format blocker-{phase}-{specID}-{timestamp}.json.
func (fs *FileSessionStore) RecordBlocker(report BlockerReport) error {
	if err := fs.ensureStateDir(); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	timestamp := report.Timestamp.Format("20060102-150405")
	// REQ-012: blocker-{phase}-{specID}-{timestamp}.json format
	phaseStr := string(report.Phase)
	if phaseStr == "" {
		phaseStr = "unknown"
	}
	specIDStr := report.SPECID
	if specIDStr == "" {
		specIDStr = "unknown"
	}
	filename := filepath.Join(fs.stateDir, fmt.Sprintf("blocker-%s-%s-%s.json", phaseStr, specIDStr, timestamp))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal blocker: %w", err)
	}

	tmpFile := filename + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write tmp blocker: %w", err)
	}

	if err := os.Rename(tmpFile, filename); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename blocker: %w", err)
	}

	return nil
}

// ResolveBlocker finds and updates the most recent unresolved blocker for the phase/spec.
func (fs *FileSessionStore) ResolveBlocker(phase Phase, specID string, resolution string) error {
	pattern := filepath.Join(fs.stateDir, "blocker-*.json")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob blockers: %w", err)
	}

	// Find the most recent unresolved blocker
	var latestBlocker *BlockerReport
	var latestPath string
	var latestTime time.Time

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		var blocker BlockerReport
		if err := json.Unmarshal(data, &blocker); err != nil {
			continue
		}

		if !blocker.Resolved && blocker.Timestamp.After(latestTime) {
			latestBlocker = &blocker
			latestPath = match
			latestTime = blocker.Timestamp
		}
	}

	if latestBlocker == nil {
		return fmt.Errorf("no outstanding blocker found")
	}

	// Update the blocker
	latestBlocker.Resolve(resolution)
	data, err := json.MarshalIndent(latestBlocker, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal resolved blocker: %w", err)
	}

	tmpFile := latestPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write resolved blocker: %w", err)
	}

	if err := os.Rename(tmpFile, latestPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("atomic rename resolved blocker: %w", err)
	}

	return nil
}

// ensureStateDir creates the state directory if it doesn't exist.
func (fs *FileSessionStore) ensureStateDir() error {
	return os.MkdirAll(fs.stateDir, 0755)
}

// checkpointPath returns the file path for a phase checkpoint.
func (fs *FileSessionStore) checkpointPath(phase Phase, specID string) string {
	return filepath.Join(fs.stateDir, fmt.Sprintf("checkpoint-%s-%s.json", phase, specID))
}
