package astgrep

// M3 severity + anchor contract, asserted against the SHIPPED ruleset.
//
// These tests were authored red: at authoring time no security rule carried a
// metadata.anchor entry, so TestSecurityRulesCarryCweAndAnchor failed for all
// fourteen. The point of asserting against the shipped tree rather than a
// fixture is that a future rule cannot be added without meeting the contract —
// a fixture-only test would leave the real tree unguarded.

import (
	"os"
	"strings"
	"testing"
)

const ruleTestRoot = "testdata/rule-tests"

func loadShipped(t *testing.T) []ShippedRule {
	t.Helper()
	rules, err := LoadShippedRules(shippedRulesDir)
	if err != nil {
		t.Fatalf("load shipped rules: %v", err)
	}
	if len(rules) != 26 {
		t.Fatalf("loaded %d rule documents, want the 26 shipped rules", len(rules))
	}
	return rules
}

// REQ-A16-011 — every security rule carries a CWE label AND a citation-or-probe
// anchor for its matched head symbol. The CWE alone is reviewer-checkable; the
// anchor is the mechanical half, held to the same standard an EXEMPT matrix
// cell owes for an absence claim.
func TestSecurityRulesCarryCweAndAnchor(t *testing.T) {
	var security int
	for _, r := range loadShipped(t) {
		if !r.IsSecurity() {
			continue
		}
		security++
		if strings.TrimSpace(r.Metadata.CWE) == "" {
			t.Errorf("%s (%s): security rule carries no metadata.cwe", r.ID, r.Language)
		}
		if !r.HasAnchor() {
			t.Errorf("%s (%s): metadata.anchor is %q; want a %v prefixed entry naming the "+
				"reference documenting its matched head symbol, or a recorded probe",
				r.ID, r.Language, r.Metadata.Anchor, AnchorPrefixes)
		}
	}
	if security != 14 {
		t.Fatalf("counted %d security rules, want 14", security)
	}
}

// REQ-A16-012 clause 1 — a rule outside a security family may not carry `error`,
// and clause 2 — every `error` rule owns a valid case whose benign construct
// shares the invalid case's shape. `sg test` proves the zero-findings half;
// this asserts the case actually exists, which `sg test` passes vacuously
// without.
func TestErrorSeverityFollowsTheTwoClausePredicate(t *testing.T) {
	cases, err := LoadRuleTestCases(ruleTestRoot)
	if err != nil {
		t.Fatalf("load rule test cases: %v", err)
	}
	for _, r := range loadShipped(t) {
		switch r.Severity {
		case SeverityError:
			if !r.IsSecurity() {
				t.Errorf("%s (%s): severity error outside a security family violates clause 1",
					r.ID, r.Language)
			}
			tc, ok := cases[r.ID]
			if !ok {
				t.Errorf("%s: severity error with no rule-test document; clause 2 is unprovable", r.ID)
				continue
			}
			if len(tc.Valid) == 0 {
				t.Errorf("%s (%s): severity error with no valid case; clause 2 requires a benign "+
					"same-shape construct producing zero findings", r.ID, r.Language)
			}
			if len(tc.Invalid) == 0 {
				t.Errorf("%s (%s): severity error with no invalid case; nothing establishes the shape",
					r.ID, r.Language)
			}
		case SeverityWarning:
			// Permitted for any rule; REQ-A16-013 makes it the default.
		default:
			t.Errorf("%s (%s): severity %q is neither error nor warning", r.ID, r.Language, r.Severity)
		}
	}
}

// REQ-A16-014 / AC-A16-017 — a shape matcher stays at `warning`, and the
// demotion is visible in the matrix rather than inferred from a severity field.
func TestShapeMatchersStayWarningWithRecordedLimitation(t *testing.T) {
	// Each entry names a security rule whose pattern matches a construct SHAPE
	// rather than a specific dangerous construct, so clause 2 has no satisfiable
	// benign same-shape counterpart.
	shapeMatchers := map[string]string{
		"sec-csrf-no-token-check":       "F6",
		"sec-log-injection-unsanitized": "F7",
	}

	seen := map[string]bool{}
	for _, r := range loadShipped(t) {
		if _, isShape := shapeMatchers[r.ID]; !isShape {
			continue
		}
		seen[r.ID] = true
		if r.Severity != SeverityWarning {
			t.Errorf("%s: severity %q; a shape matcher must stay warning (REQ-A16-014)",
				r.ID, r.Severity)
		}
	}
	for id := range shapeMatchers {
		if !seen[id] {
			t.Errorf("shape matcher %s absent from the shipped ruleset", id)
		}
	}

	data, err := os.ReadFile(matrixDocPath)
	if err != nil {
		t.Fatalf("read matrix document: %v", err)
	}
	cells, err := ParseCoverageMatrix(string(data))
	if err != nil {
		t.Fatalf("parse matrix document: %v", err)
	}
	byKey := map[string]MatrixCell{}
	for _, c := range cells {
		byKey[c.Key()] = c
	}
	for id, family := range shapeMatchers {
		cell, ok := byKey[family+"/go"]
		if !ok {
			t.Fatalf("matrix has no %s/go cell for %s", family, id)
		}
		if !strings.Contains(strings.ToLower(cell.Evidence), "precision limitation") {
			t.Errorf("%s cell (%s) evidence = %q; REQ-A16-015 requires a recorded "+
				"precision-limitation annotation naming why clause 2 cannot be satisfied",
				id, cell.Key(), cell.Evidence)
		}
	}
}

// plan.md M3 item 5 — the handed measurement record listed this rule as
// `warning`; the shipped tree measures `error`. Pinning the measured value
// stops the stale record being re-derived from memory.
func TestTemplateInjectionRuleMeasuresError(t *testing.T) {
	for _, r := range loadShipped(t) {
		if r.ID == "sec-template-injection-html" {
			if r.Severity != SeverityError {
				t.Fatalf("sec-template-injection-html severity = %q, want %q",
					r.Severity, SeverityError)
			}
			return
		}
	}
	t.Fatal("sec-template-injection-html absent from the shipped ruleset")
}

// AC-A16-018 — the severity split across all 26 is measured, not asserted.
// A drifting count is a signal to re-apply the predicate, not to edit this
// number.
func TestSeveritySplitAcrossAllTwentySix(t *testing.T) {
	var errCount, warnCount, securityWarn int
	for _, r := range loadShipped(t) {
		switch r.Severity {
		case SeverityError:
			errCount++
		case SeverityWarning:
			warnCount++
			if r.IsSecurity() {
				securityWarn++
			}
		}
	}
	if errCount != 12 || warnCount != 14 {
		t.Errorf("severity split = %d error / %d warning, want 12 / 14", errCount, warnCount)
	}
	if securityWarn != 2 {
		t.Errorf("%d security rules sit at warning, want 2 (both shape matchers)", securityWarn)
	}
}
