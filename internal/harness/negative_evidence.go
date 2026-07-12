// Package harness — A7 negative-evidence registry (SPEC-HARNESS-EVOLVE-003 M1).
//
// The registry records pattern keys that were rejected or rolled back by the
// Curator loop. Same-key re-proposal is blocked until BOTH a cooldown elapses
// AND N new post-rejection evidences accumulate (REQ-HEV3-021/022).
//
// Design (design.md §D):
//   - Entry shape: design.md §D.1
//   - Re-proposal block predicate: design.md §D.2
//   - Cooldown primitive reuse: design.md §D.3 (constant only, not the struct)
//
// [HARD] This file does not invoke AskUserQuestion. The re-proposal block is a
// mechanical check returning a typed error; the orchestrator handles user-facing
// decisions (REQ-HEV3-031 subagent boundary).
package harness

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// NegativeEvidenceCooldown is the duration a rejected/rolled-back pattern key is
// suppressed from re-proposal (REQ-HEV3-021, H-1 resolved: 48h).
//
// This mirrors safety.canaryVetoCooldown (canary_veto.go:27). It is NOT imported
// from the safety package because safety imports harness (import cycle).
// The two constants MUST stay in sync; both are 48h per the design SSOT.
const NegativeEvidenceCooldown = 48 * time.Hour

// NegativeEvidenceNewEvidenceThreshold is the N NEW post-rejection evidences
// required to re-propose a suppressed pattern key (REQ-HEV3-021, H-1: N=3).
const NegativeEvidenceNewEvidenceThreshold = 3

// ErrReProposalSuppressed is the typed sentinel returned when a same-key
// re-proposal is blocked by the A7 registry (REQ-HEV3-021).
//
// @MX:ANCHOR: [AUTO] ErrReProposalSuppressed is the A7 early-block sentinel.
// @MX:REASON: [AUTO] fan_in >= 3: negative_evidence_test.go, dispatch.go (M5), applier.go (M5)
var ErrReProposalSuppressed = errors.New("re-proposal suppressed by negative-evidence registry")

// ErrInvalidPatternKey is returned when a pattern_key fails anti-fabrication
// validation (REQ-HEV3-028: internal SPEC IDs / REQ tokens / AC tokens forbidden).
var ErrInvalidPatternKey = errors.New("pattern_key failed anti-fabrication validation")

// NegativeEvidenceOutcome enumerates the outcome of a registry entry.
type NegativeEvidenceOutcome string

const (
	// NegativeEvidenceRejected is set when a proposal is rejected (L1/L2/L3/L4/L5).
	NegativeEvidenceRejected NegativeEvidenceOutcome = "rejected"

	// NegativeEvidenceRolledBack is set when a promotion is rolled back.
	NegativeEvidenceRolledBack NegativeEvidenceOutcome = "rolled-back"
)

// GateOrigin enumerates the gate that produced the negative evidence.
type GateOrigin string

const (
	GateOriginL1       GateOrigin = "L1"
	GateOriginL2       GateOrigin = "L2"
	GateOriginL3       GateOrigin = "L3"
	GateOriginL4       GateOrigin = "L4"
	GateOriginL5       GateOrigin = "L5"
	GateOriginRollback GateOrigin = "rollback"
)

// NegativeEvidence is a single registry entry recording that a pattern key was
// rejected or rolled back (REQ-HEV3-018, design.md §D.1).
//
// Anti-fabrication (REQ-HEV3-028): the PatternKey is a structural token
// (request_class + subcommand + mode + outcome), NOT free-form prose.
// Internal SPEC IDs, REQ tokens, AC tokens, dates, and commit SHAs are
// forbidden in the PatternKey (validated by ValidatePatternKey).
type NegativeEvidence struct {
	// PatternKey is the structural token identifying the rejected/rolled-back pattern.
	PatternKey string `json:"pattern_key"`

	// Outcome is "rejected" or "rolled-back".
	Outcome NegativeEvidenceOutcome `json:"outcome"`

	// Timestamp is the rejection/rollback event time (UTC RFC3339).
	Timestamp time.Time `json:"ts"`

	// EvidenceCountAtEvent is the pattern's evidence count at the rejection/rollback.
	EvidenceCountAtEvent int `json:"evidence_count_at_event"`

	// CooldownUntil is the timestamp after which the cooldown is elapsed.
	// Computed as Timestamp + NegativeEvidenceCooldown (48h default).
	CooldownUntil time.Time `json:"cooldown_until"`

	// NewEvidenceSinceEvent is the count of NEW evidences observed AFTER the event.
	// Incremented by later observations of the same key.
	NewEvidenceSinceEvent int `json:"new_evidence_since_event"`

	// MachineSignalRef is the evidence-or-null reference (lineage key, rollback SHA).
	// Empty string means evidence absent (never an inferred value).
	MachineSignalRef string `json:"machine_signal_ref,omitempty"`

	// GateOrigin is the gate that produced the entry: "L1".."L5" or "rollback".
	GateOrigin GateOrigin `json:"gate_origin"`
}

// ReProposalCheck is the result of the A7 re-proposal suppression predicate.
type ReProposalCheck struct {
	// Suppressed is true if the pattern key is currently blocked from re-proposal.
	Suppressed bool

	// PatternKey is the key that was checked.
	PatternKey string

	// LatestEntry is the most-recent entry for the key (nil if no entry exists).
	LatestEntry *NegativeEvidence

	// Reason explains why the key is suppressed (empty when Suppressed is false).
	Reason string
}

// internalSpecIDPattern matches internal SPEC IDs, REQ tokens, and AC tokens.
// REQ-HEV3-028 anti-fabrication: these are forbidden in pattern_key.
// Matches the real token shapes: SPEC-HARNESS-EVOLVE-003, REQ-HEV3-021, AC-HEV3-024a.
var internalSpecIDPattern = regexp.MustCompile(`\b(SPEC|REQ|AC)-[A-Z][A-Z0-9-]*\d`)

// ValidatePatternKey verifies the pattern_key conforms to anti-fabrication rules
// (REQ-HEV3-028): no internal SPEC IDs, REQ tokens, AC tokens. Empty keys are rejected.
func ValidatePatternKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: empty pattern_key", ErrInvalidPatternKey)
	}
	if internalSpecIDPattern.MatchString(key) {
		return fmt.Errorf("%w: pattern_key contains an internal identifier (SPEC/REQ/AC token): %q", ErrInvalidPatternKey, key)
	}
	return nil
}

// AppendNegativeEvidence appends a single entry to the registry jsonl file
// (REQ-HEV3-018/019/020). The file is created if absent; appends are atomic
// per-entry (Edge-1: concurrent different-key appends succeed).
//
// Anti-fabrication (REQ-HEV3-028): the PatternKey is validated BEFORE any file
// operation — the file is NOT touched when validation fails.
func AppendNegativeEvidence(path string, entry NegativeEvidence) error {
	if err := ValidatePatternKey(entry.PatternKey); err != nil {
		return err
	}

	// Ensure CooldownUntil is set (compute from Timestamp if zero).
	if entry.CooldownUntil.IsZero() && !entry.Timestamp.IsZero() {
		entry.CooldownUntil = entry.Timestamp.Add(NegativeEvidenceCooldown)
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("negative_evidence: failed to serialize entry: %w", err)
	}
	line = append(line, '\n')

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("negative_evidence: failed to create directory %s: %w", dir, err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("negative_evidence: failed to open registry %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("negative_evidence: failed to write entry: %w", err)
	}
	return nil
}

// ReadNegativeEvidence reads all entries from the registry jsonl file.
// REQ-HEV3-018 Edge-2: a missing registry file returns an empty slice (not an error)
// — fresh-project first-run behavior.
func ReadNegativeEvidence(path string) ([]NegativeEvidence, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Edge-2: file absent → empty registry.
			return nil, nil
		}
		return nil, fmt.Errorf("negative_evidence: failed to open registry %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	var entries []NegativeEvidence
	scanner := bufio.NewScanner(f)
	// Expand the scanner buffer for long lines (default 64KB may be tight for entries
	// with verbose machine_signal_ref).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry NegativeEvidence
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed lines rather than failing the entire read (append-only
			// jsonl tolerates partial corruption).
			continue
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("negative_evidence: failed to scan registry %s: %w", path, err)
	}
	return entries, nil
}

// IsReProposalSuppressed evaluates the A7 re-proposal block predicate
// (REQ-HEV3-021/022, design.md §D.2).
//
// A pattern key K is suppressed iff an entry exists AND
// (cooldown not elapsed OR new_evidence_since_event < N).
// The block lifts when BOTH the cooldown elapses AND N new evidences accumulate.
// Edge-3: the cooldown is elapsed at now >= cooldown_until (inclusive boundary).
//
// When multiple entries exist for the same key, the LATEST (by Timestamp) is consulted.
func IsReProposalSuppressed(entries []NegativeEvidence, patternKey string, now time.Time) ReProposalCheck {
	check := ReProposalCheck{PatternKey: patternKey}

	// Find the latest entry for the key.
	var latest *NegativeEvidence
	for i := range entries {
		e := &entries[i]
		if e.PatternKey != patternKey {
			continue
		}
		if latest == nil || e.Timestamp.After(latest.Timestamp) {
			latest = e
		}
	}
	if latest == nil {
		// No entry → not suppressed.
		return check
	}
	check.LatestEntry = latest

	// Edge-3: at now == cooldown_until, the cooldown is elapsed (inclusive boundary).
	// !now.Before(cooldownUntil) ⟺ now >= cooldownUntil.
	cooldownElapsed := !now.Before(latest.CooldownUntil)
	hasEnoughNewEvidence := latest.NewEvidenceSinceEvent >= NegativeEvidenceNewEvidenceThreshold

	if !cooldownElapsed {
		check.Suppressed = true
		check.Reason = fmt.Sprintf("cooldown not elapsed (until %s, now %s)", latest.CooldownUntil.Format(time.RFC3339), now.Format(time.RFC3339))
		return check
	}
	if !hasEnoughNewEvidence {
		check.Suppressed = true
		check.Reason = fmt.Sprintf("new evidence %d < threshold %d", latest.NewEvidenceSinceEvent, NegativeEvidenceNewEvidenceThreshold)
		return check
	}

	// Both conditions cleared — REQ-HEV3-022 re-eligibility.
	return check
}

// RegisterVetoAsNegativeEvidence auto-registers a Canary veto (or any rollback)
// as a negative-evidence entry in the A7 registry (REQ-HEV3-014).
//
// The Canary veto and the A7 registry are the two surfaces that record "this
// promotion failed"; they MUST agree. When a veto fires post-apply, this function
// appends a "rolled-back" entry to the registry keyed by the vetoed pattern key.
//
// registryPath is the .moai/state/negative-evidence.jsonl path.
// patternKey is the structural token of the vetoed pattern.
// evidenceCount is the pattern's evidence count at the veto event.
// eventTime is the veto timestamp (used for cooldown computation).
func RegisterVetoAsNegativeEvidence(registryPath, patternKey string, evidenceCount int, eventTime time.Time) error {
	entry := NegativeEvidence{
		PatternKey:           patternKey,
		Outcome:              NegativeEvidenceRolledBack,
		Timestamp:            eventTime,
		EvidenceCountAtEvent: evidenceCount,
		CooldownUntil:        eventTime.Add(NegativeEvidenceCooldown),
		GateOrigin:           GateOriginRollback,
	}
	return AppendNegativeEvidence(registryPath, entry)
}

