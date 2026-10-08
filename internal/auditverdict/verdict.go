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
	// RequiredBackendFails collects every required_backend_fail line's named
	// backend; any entry makes the file inadmissible regardless of its label
	// (SPEC-AUDIT-CEILING-002 REQ-ACR-005). Repeats both collect.
	RequiredBackendFails []string
	// MalformedDebts counts debt lines that did not carry an id, a valid
	// dispose_in, and a description.
	MalformedDebts int
	// Receipt carries the convergence receipt lines (SPEC-AUDIT-CEILING-001
	// REQ-ACE-008): the convergence overall verdict and one line per audited
	// required backend.
	Receipt Receipt
	// MalformedReceipts counts receipt lines that did not carry a value in
	// the receipt vocabulary (pass|fail|inconclusive); any entry makes the
	// file inadmissible.
	MalformedReceipts int
}

// Receipt is the convergence receipt a required-backend tree's verdict file
// records. ConvergenceOverall is pass or fail; Backends maps each backend the
// audit covered to pass, fail, or inconclusive. Present reports whether any
// receipt line appeared.
type Receipt struct {
	Present            bool
	ConvergenceOverall string
	Backends           map[string]string
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
	seen := map[string]string{}
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
		case "verdict", "overall_score", "must_pass_failed", "blocking_count", "plan_artifact_hash", "convergence_overall":
			norm, ok := normalizeDecisionValue(canonical, value)
			if prev, dup := seen[canonical]; dup && (!ok || prev != norm) {
				f.DuplicateKeys = append(f.DuplicateKeys, canonical)
			}
			if _, dup := seen[canonical]; !dup {
				seen[canonical] = norm
			}
		case "required_backend_fail":
			// Every occurrence collects — a backend recorded fail is a refusal
			// signal, never a dedupe candidate (REQ-ACR-005).
			if value != "" {
				f.RequiredBackendFails = append(f.RequiredBackendFails, value)
			}
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
		case "convergence_overall":
			if v := strings.ToLower(value); v == "pass" || v == "fail" {
				f.Receipt.Present = true
				f.Receipt.ConvergenceOverall = v
			} else {
				f.Receipt.Present = true
				f.MalformedReceipts++
			}
		case "required_backend":
			// Repeatable by design — one line per audited backend (D18). A
			// second line for a recorded backend is a duplicate only when the
			// recorded verdict differs; a value outside the receipt vocabulary
			// stays malformed.
			f.Receipt.Present = true
			name, verdict, _ := strings.Cut(value, " ")
			verdict = strings.TrimSpace(verdict)
			name = strings.TrimSpace(name)
			if name == "" || (verdict != "pass" && verdict != "fail" && verdict != "inconclusive") {
				f.MalformedReceipts++
				break
			}
			if prev, dup := f.Receipt.Backends[name]; dup {
				if prev != verdict {
					f.DuplicateKeys = append(f.DuplicateKeys, "required_backend("+name+")")
				}
				break
			}
			if f.Receipt.Backends == nil {
				f.Receipt.Backends = map[string]string{}
			}
			f.Receipt.Backends[name] = verdict
		}
	}
	return f
}

// normalizeDecisionValue canonicalizes a decision value so equal repeats
// (a report header plus its machine line) compare equal; ok is false when
// the value does not parse, which makes any repeat of that key a conflict.
func normalizeDecisionValue(key, value string) (string, bool) {
	switch key {
	case "verdict":
		return strings.ToUpper(value), value != ""
	case "overall_score":
		s, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(s) || math.IsInf(s, 0) {
			return value, false
		}
		return strconv.FormatFloat(s, 'f', -1, 64), true
	case "must_pass_failed", "blocking_count":
		n, err := strconv.Atoi(value)
		if err != nil {
			return value, false
		}
		return strconv.Itoa(n), true
	case "convergence_overall":
		v := strings.ToLower(value)
		return v, v == "pass" || v == "fail"
	}
	return value, value != ""
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
// hash binds the current plan artifacts; requiredBackends is the tree's
// resolved required-backend set (REQ-ACE-009/010 — a fail or inconclusive
// recorded line refuses regardless of the own label, a configured backend
// with no receipt line refuses, and an empty set admits a receipt-less
// verdict). All but requiredBackends are ignored in the sync phase.
// A refusal always carries a reason.
//
// @MX:ANCHOR: [AUTO] the single verdict admission predicate shared by contract rules, kickoff decide, the card-transition guard, and the plan-audit ceiling evaluation
// @MX:REASON: a second copy of this rule is exactly the drift that let PASS-WITH-DEBT pass three code sites while doctrine blocked it
func Admit(f Fields, phase Phase, threshold float64, hashOK bool) (bool, string) {
	return admitBase(f, phase, threshold, hashOK)
}

// AdmitWithRequired is Admit plus the required-backend receipt checks
// (SPEC-AUDIT-CEILING-001 REQ-ACE-009/010): a configured required backend
// with a fail/inconclusive receipt line, an omitted receipt, or a malformed
// receipt line refuses fail-closed. The plan-audit ceiling evaluation and
// the admission seams that must persist per-backend refusal records call
// this form; callers without a required-backend configuration use Admit.
func AdmitWithRequired(f Fields, phase Phase, threshold float64, hashOK bool, requiredBackends []string) (bool, string) {
	if ok, reason := admitBase(f, phase, threshold, hashOK); !ok {
		return false, reason
	}
	if reason, refused := admitReceipt(f, requiredBackends); refused {
		return false, reason
	}
	return true, ""
}

func admitBase(f Fields, phase Phase, threshold float64, hashOK bool) (bool, string) {
	if len(f.DuplicateKeys) > 0 {
		return false, "duplicated decision key(s): " + strings.Join(f.DuplicateKeys, ", ")
	}
	// Unconditional (the sync phase included): a required backend the exporting
	// auditor recorded as fail refuses the verdict regardless of its own label
	// (SPEC-AUDIT-CEILING-002 REQ-ACR-005). The absence of lines refuses
	// nothing — the label and field checks below stay the primary gates.
	if len(f.RequiredBackendFails) > 0 {
		return false, "required backend(s) recorded fail: " + strings.Join(dedupeStrings(f.RequiredBackendFails), ", ")
	}
	// The receipt's own overall verdict is the exporting auditor's bottom
	// line: a recorded "fail" refuses the verdict regardless of its label or
	// backend lines (card t1571 — the same unconditional posture as the
	// required_backend_fail refusal above). An absent convergence_overall is
	// the legacy shape and stays on the label and field checks below.
	if f.Receipt.ConvergenceOverall == "fail" {
		return false, "convergence receipt records overall fail"
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

// ReceiptRefusal reports whether the verdict carries a required-backend
// receipt refusal (REQ-ACE-009/010) independent of Admit's other checks —
// the admission seams use it to persist the refusal record REQ-ACE-007/012
// requires for every required-backend refusal, including the below-ceiling
// ones the ceiling ladder never sees (card-review F7).
func ReceiptRefusal(f Fields, required []string) (string, bool) {
	return admitReceipt(f, required)
}

// admitReceipt applies the required-backend receipt checks of
// SPEC-AUDIT-CEILING-001 (design.md §4): a required backend recorded fail or
// inconclusive refuses (REQ-ACE-009); a configured required backend with no
// receipt at all, or a receipt omitting its line, refuses fail-closed
// (REQ-ACE-010); a malformed receipt line refuses (edges 2-3). No required
// backend configured and no malformed line admits (C4).
func admitReceipt(f Fields, required []string) (string, bool) {
	if len(required) > 0 {
		for _, b := range required {
			v, ok := f.Receipt.Backends[b]
			if !ok {
				if !f.Receipt.Present {
					return fmt.Sprintf("required backend %s configured and the verdict carries no convergence receipt", b), true
				}
				return fmt.Sprintf("receipt omits the configured required backend %s", b), true
			}
			if v != "pass" {
				return fmt.Sprintf("required backend %s recorded %s", b, v), true
			}
		}
	}
	if f.MalformedReceipts > 0 {
		return fmt.Sprintf("%d malformed receipt line(s)", f.MalformedReceipts), true
	}
	return "", false
}

// dedupeStrings returns the values in order, duplicates removed — refusal
// reasons name each backend once.
func dedupeStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := values[:0:0]
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
