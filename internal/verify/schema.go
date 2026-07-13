// Package verify implements the shared diagnostic snapshot contract: a
// structured, working-tree-keyed record of executed quality checks (tests,
// lint, coverage, ...) that sibling verification layers can reuse instead of
// re-executing, under a strict freshness rule (key equality AND wall-clock
// TTL). The package owns the snapshot schema (schema.go), the 3-input key
// computation (key.go), the freshness predicate (freshness.go), the atomic
// store (store.go), and the goal-evaluator lookup adapter (source.go).
//
// A stale snapshot is never reusable evidence: reuse is valid attribution only
// while the stored key equals the current working-tree key and the recorded
// capture timestamp is within the TTL (verification-claim-integrity.md §2 —
// the snapshot IS the observed evidence; the freshness rule is what keeps the
// attribution valid for the current tree).
package verify

import "time"

// Conditions is the parsed-result block of one check, read-compatible with
// the loop-verdict `conditions` shape (zero_errors / error_count / tests_pass /
// coverage_threshold / coverage_actual / zero_warnings). Fields are pointers so
// each is populated only where applicable to the check class.
type Conditions struct {
	ZeroErrors        *bool    `json:"zero_errors,omitempty"`
	ErrorCount        *int     `json:"error_count,omitempty"`
	TestsPass         *bool    `json:"tests_pass,omitempty"`
	CoverageThreshold *float64 `json:"coverage_threshold,omitempty"`
	CoverageActual    *float64 `json:"coverage_actual,omitempty"`
	ZeroWarnings      *bool    `json:"zero_warnings,omitempty"`
}

// CheckEntry records one executed check: identifier, the exact byte-string
// command executed, its exit code, capture timestamp, execution duration, and
// the parsed result counts where applicable.
type CheckEntry struct {
	CheckID    string      `json:"check_id"`
	Command    string      `json:"command"`
	ExitCode   int         `json:"exit_code"`
	RecordedAt time.Time   `json:"recorded_at"`
	DurationMS int64       `json:"duration_ms"`
	Conditions *Conditions `json:"conditions,omitempty"`
}

// Snapshot is the per-key diagnostic snapshot document persisted as one JSON
// file per key under .moai/state/verify/snapshots/. RecordedAt tracks the
// latest write; each CheckEntry carries its own capture timestamp.
type Snapshot struct {
	Key        string       `json:"key"`
	RecordedAt time.Time    `json:"recorded_at"`
	Checks     []CheckEntry `json:"checks"`
}

// FindCommand returns the check entry whose recorded command equals cmd by
// exact byte-string comparison — no normalization: a variant differing by even
// one byte never matches. Returns nil when no entry matches.
func (s *Snapshot) FindCommand(cmd string) *CheckEntry {
	if s == nil {
		return nil
	}
	for i := range s.Checks {
		if s.Checks[i].Command == cmd {
			return &s.Checks[i]
		}
	}
	return nil
}
