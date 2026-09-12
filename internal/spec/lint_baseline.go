package spec

// lint_baseline.go implements the per-rule baseline ratchet
// (SPEC-SPECLINT-GATE-SIGNAL-001 M2).
//
// The problem it solves: `--strict` escalates EVERY non-advisory warning to an
// error, so a standing inventory of one or more such warnings makes the gate
// permanently red. A permanently-red gate carries no signal — the person who
// broke something and the person who changed nothing see the same red.
//
// The mechanism: record the per-rule non-advisory warning counts in a
// checked-in file; gate on the DELTA. Standing inventory at or below its
// recorded count passes (and stays visible in the output); any rule above its
// recorded count is red and is named with its delta.

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// BaselineSchemaVersion is the on-disk schema version of the baseline file.
// It is bumped only on an incompatible shape change.
const BaselineSchemaVersion = 1

// Baseline is the checked-in per-rule ratchet record.
//
// The gating dimension is deliberately per-rule NON-ADVISORY warning counts.
// Advisory findings are excluded on purpose: the advisory boundary itself is
// owned by a separate in-flight change (SPEC-SPEC-LINT-BLIND-AXES-001), and
// that change moves the advisory population wholesale. Recording advisory
// counts here would guarantee an immediate rebaseline the moment it lands,
// which turns the audited rebaseline procedure (REQ-SLGS-008) into a routine.
//
// Error-severity findings are also excluded: they gate unconditionally
// (REQ-SLGS-009) and must never be absorbable by a baseline.
type Baseline struct {
	// Version is the on-disk schema version (see BaselineSchemaVersion).
	Version int `json:"version"`
	// UpdatedAt is the ISO date (YYYY-MM-DD) the record was captured.
	UpdatedAt string `json:"updated_at"`
	// TreeSHA is the git commit the counts were measured against. A count
	// with no tree attribution is not a measurement.
	TreeSHA string `json:"tree_sha"`
	// Reason is the mandatory non-empty rebaseline rationale (REQ-SLGS-008).
	Reason string `json:"reason"`
	// Rules maps a finding code to its recorded non-advisory warning count.
	Rules map[string]int `json:"rules"`
}

// RuleDelta is one rule's recorded-vs-current pair.
type RuleDelta struct {
	Code     string `json:"code"`
	Recorded int    `json:"recorded"`
	Current  int    `json:"current"`
}

// Delta is the signed difference (positive on an increase).
func (d RuleDelta) Delta() int { return d.Current - d.Recorded }

// BaselineComparison is the verdict of one report against one baseline.
type BaselineComparison struct {
	// ErrorCount is the number of error-severity findings. Non-zero fails
	// the gate regardless of everything else (REQ-SLGS-009).
	ErrorCount int
	// TotalWarnings counts ALL warnings, advisory included, so the standing
	// inventory stays observable while the gate is green (REQ-SLGS-004).
	TotalWarnings int
	// NonAdvisoryWarnings counts only the gated dimension.
	NonAdvisoryWarnings int
	// Increases are the rules above their recorded count, sorted by code.
	Increases []RuleDelta
	// Decreases are the rules below their recorded count, sorted by code.
	Decreases []RuleDelta
}

// Failed reports whether the gate is red.
func (c BaselineComparison) Failed() bool {
	return c.ErrorCount > 0 || len(c.Increases) > 0
}

// ErrorGated reports whether the failure cause is error severity.
//
// This distinction is load-bearing, not cosmetic: it is what makes the
// error-BEFORE-baseline ordering (REQ-SLGS-009, plan.md §D.5) observable. Both
// causes exit 1, so an implementation that checked the baseline first would be
// indistinguishable by exit code alone; it is distinguishable by which cause
// the output names.
func (c BaselineComparison) ErrorGated() bool { return c.ErrorCount > 0 }

// ComputeBaselineRules counts non-advisory warnings per finding code.
// Codes with a zero count are absent from the map rather than recorded as 0.
func ComputeBaselineRules(r *Report) map[string]int {
	rules := map[string]int{}
	if r == nil {
		return rules
	}
	for _, f := range r.Findings {
		if f.Severity != SeverityWarning || f.Advisory {
			continue
		}
		rules[f.Code]++
	}
	return rules
}

// NewBaselineFromReport builds a baseline record from a fresh measurement.
func NewBaselineFromReport(r *Report, treeSHA, updatedAt, reason string) *Baseline {
	return &Baseline{
		Version:   BaselineSchemaVersion,
		UpdatedAt: updatedAt,
		TreeSHA:   treeSHA,
		Reason:    reason,
		Rules:     ComputeBaselineRules(r),
	}
}

// MarshalBaseline serializes deterministically: fixed field order, sorted rule
// keys, two-space indent, trailing newline. Byte-identical on every platform —
// the file is reviewed as a git diff, so a non-deterministic encoding would
// make the audit trail unreadable (plan.md §D.6).
//
// encoding/json sorts map keys and emits no OS-dependent separators, and the
// record holds no filesystem paths, so nothing here varies by platform.
func MarshalBaseline(b *Baseline) ([]byte, error) {
	if b == nil {
		return nil, fmt.Errorf("marshal baseline: nil record")
	}
	toWrite := *b
	if toWrite.Rules == nil {
		toWrite.Rules = map[string]int{}
	}
	data, err := json.MarshalIndent(&toWrite, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal baseline: %w", err)
	}
	return append(data, '\n'), nil
}

// WriteBaseline serializes and writes the record.
func WriteBaseline(path string, b *Baseline) error {
	data, err := MarshalBaseline(b)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write baseline %q: %w", path, err)
	}
	return nil
}

// LoadBaseline reads and decodes a baseline file.
//
// A missing file is returned as an error satisfying os.IsNotExist so the caller
// can name the path rather than falling back silently — a silent fallback to
// today's behaviour is exactly what the boundary-case contract forbids.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path) //nolint:gosec // caller-supplied gate path
	if err != nil {
		return nil, err
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse baseline %q: %w", path, err)
	}
	if b.Rules == nil {
		b.Rules = map[string]int{}
	}
	return &b, nil
}

// CompareBaseline judges a report against a baseline.
//
// A rule absent from the baseline has a recorded count of 0, so its first
// occurrence is an increase — that is the "new rule fires for the first time"
// case, and it is a genuine new signal.
func CompareBaseline(r *Report, b *Baseline) BaselineComparison {
	var cmp BaselineComparison
	if r == nil {
		return cmp
	}

	current := map[string]int{}
	for _, f := range r.Findings {
		switch f.Severity {
		case SeverityError:
			cmp.ErrorCount++
		case SeverityWarning:
			cmp.TotalWarnings++
			if !f.Advisory {
				cmp.NonAdvisoryWarnings++
				current[f.Code]++
			}
		case SeverityInfo:
			// info findings never gate and never enter the baseline
		}
	}

	recorded := map[string]int{}
	if b != nil && b.Rules != nil {
		recorded = b.Rules
	}

	// Union of both key sets, so a rule that stopped firing is reported as a
	// decrease rather than silently vanishing.
	codes := make([]string, 0, len(current)+len(recorded))
	seen := map[string]bool{}
	for c := range current {
		if !seen[c] {
			seen[c] = true
			codes = append(codes, c)
		}
	}
	for c := range recorded {
		if !seen[c] {
			seen[c] = true
			codes = append(codes, c)
		}
	}
	sort.Strings(codes)

	for _, code := range codes {
		d := RuleDelta{Code: code, Recorded: recorded[code], Current: current[code]}
		switch {
		case d.Current > d.Recorded:
			cmp.Increases = append(cmp.Increases, d)
		case d.Current < d.Recorded:
			cmp.Decreases = append(cmp.Decreases, d)
		}
	}

	return cmp
}
