package planhtml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parsePlanHTML(t *testing.T, raw []byte) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}
	return doc
}

func bodyText(doc *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return b.String()
}

// writeFixture writes a minimal SPEC artifact set + plan-auditor review to a
// temp specDir and returns the specDir + review path.
func writeFixture(t *testing.T) (specDir, reviewPath string) {
	t.Helper()
	specDir = t.TempDir()
	specMD := `---
id: SPEC-TEST-001
title: "Fixture SPEC"
tier: M
phase: "v3.1 target"
---
# SPEC-TEST-001
## §A. User Story
As a maintainer I want X so that the goal-clause-end-state is reached.
## §B. Scope
- REQ-T-001 — first requirement summary
- REQ-T-002 — second requirement summary
### Out of Scope — Deferred thing
- deferred sub-item
## §D. Constraints
1. stdlib only
`
	planMD := `# plan.md
## §E. Self-Verification
- Go unit test covers X
## §F. Milestones
### Milestone M1 — substrate
### Milestone M2 — verb
### Milestone M3 — report
`
	acceptanceMD := `# acceptance.md
## §D. AC Matrix
| AC ID | REQ | Severity |
|-------|-----|----------|
| AC-T-001 | REQ-T-001 | MUST |
| AC-T-002 | REQ-T-002 | MUST |
`
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "plan.md"), []byte(planMD), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatal(err)
	}
	// settings.json with permissions.allow (project root = specDir for the test).
	settingsJSON := `{"permissions":{"allow":["Bash(go test:*)","Read(*)"]}}`
	if err := os.WriteFile(filepath.Join(specDir, "settings.json"), []byte(settingsJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	reviewMD := `# SPEC Review Report: SPEC-TEST-001
Iteration: 1/3
Verdict: PASS
Overall Score: 0.85

## Must-Pass Results
- [PASS] MP-1 REQ number consistency: spec.md L1
- [PASS] MP-2 EARS format compliance: spec.md L5

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.85 | 0.75 band | spec.md L1 |

## Defects Found (structured defect-list)
D1. FIND-001 — spec.md:L1 — minor issue — Severity: minor — Class: optional — Required fix: tighten wording
`
	reviewPath = filepath.Join(specDir, "review-1.md")
	if err := os.WriteFile(reviewPath, []byte(reviewMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return specDir, reviewPath
}

// TestRenderPlanHTML_DOMShowsAllFields verifies AC-GHF-005: a DOM parse of the
// rendered plan HTML shows the goal, all 8 contract fields (non-empty), the
// verdict score, and the milestone list.
func TestRenderPlanHTML_DOMShowsAllFields(t *testing.T) {
	specDir, reviewPath := writeFixture(t)

	raw, err := RenderPlanHTML(specDir, reviewPath)
	if err != nil {
		t.Fatalf("RenderPlanHTML: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("RenderPlanHTML returned empty bytes")
	}
	doc := parsePlanHTML(t, raw)
	bt := bodyText(doc)

	// goal text (from §A "so that" clause)
	if !strings.Contains(bt, "goal-clause-end-state") {
		t.Errorf("missing goal text in body")
	}
	// verdict score 0.85
	if !strings.Contains(bt, "0.85") {
		t.Errorf("missing verdict score 0.85")
	}
	// verdict PASS
	if !strings.Contains(bt, "PASS") {
		t.Errorf("missing verdict PASS")
	}
	// milestone list (M1/M2/M3)
	for _, m := range []string{"M1", "M2", "M3"} {
		if !strings.Contains(bt, m) {
			t.Errorf("missing milestone %s", m)
		}
	}
	// all 8 contract field labels present
	for _, label := range []string{
		"goal", "scope", "non-goals", "tools-permissions",
		"stopping-condition", "evidence", "escalation", "budget",
	} {
		if !strings.Contains(strings.ToLower(bt), label) {
			t.Errorf("missing contract field label %q", label)
		}
	}
}

// TestRenderPlanHTML_Determinism verifies AC-GHF-009: two runs over the same
// fixture produce byte-identical output.
func TestRenderPlanHTML_Determinism(t *testing.T) {
	specDir, reviewPath := writeFixture(t)
	first, err := RenderPlanHTML(specDir, reviewPath)
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := RenderPlanHTML(specDir, reviewPath)
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("AC-GHF-009 determinism violated: two runs differ (%d vs %d bytes)", len(first), len(second))
	}
}

// TestRenderPlanHTML_FailOpenOnMissingReview verifies REQ-GHF-007 fail-open:
// when the review file is absent, the renderer emits the report with an "audit
// verdict unavailable" placeholder rather than failing.
func TestRenderPlanHTML_FailOpenOnMissingReview(t *testing.T) {
	specDir, _ := writeFixture(t)
	missingReview := filepath.Join(specDir, "nonexistent-review.md")

	raw, err := RenderPlanHTML(specDir, missingReview)
	if err != nil {
		t.Fatalf("fail-open: RenderPlanHTML must not error on missing review: %v", err)
	}
	bt := bodyText(parsePlanHTML(t, raw))
	if !strings.Contains(strings.ToLower(bt), "unavailable") {
		t.Errorf("fail-open placeholder missing; got:\n%s", bt)
	}
}

// TestRenderPlanHTML_SettingsJSONAbsent verifies the 8-field derivation rule
// for tools-permissions degrades to "undetermined" when settings.json is absent.
func TestRenderPlanHTML_SettingsJSONAbsent(t *testing.T) {
	specDir, reviewPath := writeFixture(t)
	// Remove settings.json so the tools-permissions field cannot be derived.
	if err := os.Remove(filepath.Join(specDir, "settings.json")); err != nil {
		t.Fatal(err)
	}
	raw, err := RenderPlanHTML(specDir, reviewPath)
	if err != nil {
		t.Fatalf("RenderPlanHTML: %v", err)
	}
	bt := bodyText(parsePlanHTML(t, raw))
	if !strings.Contains(strings.ToLower(bt), "undetermined") {
		t.Errorf("tools-permissions should degrade to 'undetermined' when settings.json absent")
	}
}
