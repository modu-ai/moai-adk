// Package auditverdict is the one admission predicate for audit verdict
// files. The contract rule, the kickoff evaluator, and the card-transition
// guard all decide admission here, so a verdict cannot pass one site and fail
// another.
package auditverdict

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Phase selects the admission rule set.
type Phase int

const (
	// PhasePlan admits a plan-audit verdict (T7 and the Kickoff sites).
	PhasePlan Phase = iota
	// PhaseSync admits a sync-audit verdict (T13): the label-only check,
	// unchanged.
	PhaseSync
)

// Labels.
const (
	LabelPass         = "PASS"
	LabelPassWithDebt = "PASS-WITH-DEBT"
)

// Debt is one enumerated debt item of a PASS-WITH-DEBT verdict.
type Debt struct {
	ID          string
	DisposeIn   string
	Description string
}

// Fields are the machine fields read from a verdict file.
type Fields struct {
	Label            string
	Score            float64
	ScoreOK          bool
	MustPassFailed   int
	MustPassKnown    bool
	BlockingCount    int
	BlockingKnown    bool
	PlanArtifactHash string
	Debts            []Debt
	// DuplicateKeys names decision keys that appeared more than once; any
	// entry makes the file inadmissible.
	DuplicateKeys []string
	// MalformedDebts counts debt lines that did not carry an id, a valid
	// dispose_in, and a description.
	MalformedDebts int
}

var (
	debtLine     = regexp.MustCompile(`^-?\s*debt:\s*(.*)$`)
	debtWellForm = regexp.MustCompile(`^(\S+)\s+dispose_in=(run|sync)\s+(\S.*)$`)
)

// Parse reads the machine fields. Keys are matched case-insensitively with
// spaces folded to underscores, so "Overall Score: 0.9" and
// "overall_score: 0.9" both set the score. A debt is one line of the form
// "- debt: <id> dispose_in=<run|sync> <description>".
func Parse(raw []byte) Fields {
	var f Fields
	seen := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if m := debtLine.FindStringSubmatch(line); m != nil {
			if d := debtWellForm.FindStringSubmatch(strings.TrimSpace(m[1])); d != nil {
				f.Debts = append(f.Debts, Debt{ID: d[1], DisposeIn: d[2], Description: strings.TrimSpace(d[3])})
			} else {
				f.MalformedDebts++
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(key), " ", "_"))
		value = strings.TrimSpace(strings.Trim(strings.TrimSpace(value), "`"))
		// A decision key seen twice makes the file inadmissible (never
		// last-wins): an appended line must not overturn the first.
		canonical := key
		switch key {
		case "score":
			canonical = "overall_score"
		case "plan-artifact-hash":
			canonical = "plan_artifact_hash"
		}
		switch canonical {
		case "verdict", "overall_score", "must_pass_failed", "blocking_count", "plan_artifact_hash":
			if seen[canonical] {
				f.DuplicateKeys = append(f.DuplicateKeys, canonical)
			}
			seen[canonical] = true
		}
		switch key {
		case "verdict":
			f.Label = value
		case "overall_score", "score":
			// A score is a finite number in [0, 1]; NaN, ±Inf, and out-of-range
			// values stay unparsed so the threshold comparison never sees them.
			if s, err := strconv.ParseFloat(value, 64); err == nil && !math.IsNaN(s) && !math.IsInf(s, 0) && s >= 0 && s <= 1 {
				f.Score, f.ScoreOK = s, true
			}
		case "must_pass_failed":
			if n, err := strconv.Atoi(value); err == nil {
				f.MustPassFailed, f.MustPassKnown = n, true
			}
		case "blocking_count":
			if n, err := strconv.Atoi(value); err == nil {
				f.BlockingCount, f.BlockingKnown = n, true
			}
		case "plan_artifact_hash", "plan-artifact-hash":
			f.PlanArtifactHash = value
		}
	}
	return f
}

// AdmitLabel reports whether a verdict label is PASS-family. It is the whole
// rule where a site holds only a label (the sync phase, the contract's
// recorded plan-audit label).
func AdmitLabel(label string) bool {
	return label == LabelPass || label == LabelPassWithDebt
}

// planThresholds are the tier plan-audit PASS thresholds (spec-workflow.md
// § SPEC Complexity Tier).
var planThresholds = map[string]float64{"S": 0.75, "M": 0.80, "L": 0.85}

// SpecTier reads spec.md's tier in specDir; an absent or unknown tier is L,
// the backward-compatible default.
func SpecTier(specDir string) string {
	data, err := os.ReadFile(filepath.Join(specDir, "spec.md"))
	if err != nil {
		return "L"
	}
	for _, l := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(l), "tier:"); ok {
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if _, known := planThresholds[v]; known {
				return v
			}
		}
	}
	return "L"
}

// PlanThreshold is the plan PASS threshold of the SPEC in specDir.
func PlanThreshold(specDir string) float64 {
	return planThresholds[SpecTier(specDir)]
}

// Admit applies the phase's admission predicate. threshold is the tier's
// plan PASS threshold; hashOK reports whether the verdict's plan-artifact
// hash binds the current plan artifacts. Both are ignored in the sync phase.
// A refusal always carries a reason.
//
// @MX:ANCHOR: [AUTO] the single verdict admission predicate shared by contract rules, kickoff decide, and the card-transition guard
// @MX:REASON: a second copy of this rule is exactly the drift that let PASS-WITH-DEBT pass three code sites while doctrine blocked it
func Admit(f Fields, phase Phase, threshold float64, hashOK bool) (bool, string) {
	if len(f.DuplicateKeys) > 0 {
		return false, "duplicated decision key(s): " + strings.Join(f.DuplicateKeys, ", ")
	}
	if !AdmitLabel(f.Label) {
		if f.Label == "" {
			return false, "no verdict"
		}
		return false, "verdict " + f.Label
	}
	if phase == PhaseSync {
		return true, ""
	}
	switch {
	case !f.ScoreOK:
		return false, "unparseable or missing score"
	case f.Score < threshold:
		return false, fmt.Sprintf("score %.3f below the tier threshold %.2f", f.Score, threshold)
	case !f.MustPassKnown:
		return false, "verdict carries no must_pass_failed field"
	case f.MustPassFailed != 0:
		return false, fmt.Sprintf("must_pass_failed %d", f.MustPassFailed)
	case !f.BlockingKnown:
		return false, "verdict carries no blocking_count field"
	case f.BlockingCount != 0:
		return false, fmt.Sprintf("blocking_count %d", f.BlockingCount)
	case !hashOK:
		return false, "plan_artifact_hash does not bind the current plan artifacts"
	}
	if f.Label == LabelPassWithDebt {
		if f.MalformedDebts > 0 {
			return false, fmt.Sprintf("%d malformed debt line(s)", f.MalformedDebts)
		}
		if len(f.Debts) == 0 {
			return false, "PASS-WITH-DEBT enumerates no debts"
		}
	}
	return true, ""
}
