package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// parseDOMBody returns the concatenated text content of the parsed HTML body.
func parseDOMBody(t *testing.T, raw []byte) string {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return b.String()
}

// fixtureSpecDir writes a minimal SPEC directory under <root>/.moai/specs/<id>/
// containing spec.md, plan.md, acceptance.md. Returns the spec dir path.
func fixtureSpecDir(t *testing.T, root, specID, goalClause string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "specs", specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := fmt.Sprintf(`---
id: %s
title: "Fixture SPEC"
version: "0.1.0"
tier: M
---

## §A. Context

Wire the renderer so that %s.

## §B. Goals

- REQ-WIRE-001 — wire the path
- REQ-WIRE-002 — write the file
`, specID, goalClause)
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}
	planMD := `# plan.md

## §F. Milestones

### Milestone M1

First.

### Milestone M2

Second.
`
	if err := os.WriteFile(filepath.Join(dir, "plan.md"), []byte(planMD), 0o644); err != nil {
		t.Fatal(err)
	}
	acceptanceMD := `# acceptance.md

## §D. AC Matrix

| AC ID | REQ | Subject |
|-------|-----|---------|
| AC-X-001 | REQ-WIRE-001 | first AC |
| AC-X-002 | REQ-WIRE-002 | second AC |

### AC-X-001
**Then** the file exists.
`
	if err := os.WriteFile(filepath.Join(dir, "acceptance.md"), []byte(acceptanceMD), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// fixtureReviewFile writes a plan-audit review markdown carrying the markers
// the planhtml parser extracts (Verdict / Overall Score / must-pass / defect).
func fixtureReviewFile(t *testing.T, root, specID string) string {
	t.Helper()
	dir := filepath.Join(root, ".moai", "reports", "plan-audit")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, specID+"-review-1.md")
	md := `Verdict: PASS
Overall Score: 0.85

- [PASS] MP-1 must-pass row A
- [FAIL] MP-2 must-pass row B

D1. defect line one
D2. defect line two
`
	if err := os.WriteFile(path, []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// drivePlanRenderHTML invokes the `moai plan render-html <specID>` subcommand
// against a fresh goal-agnostic root with CLAUDE_PROJECT_DIR pinned to root.
func drivePlanRenderHTML(t *testing.T, root, specID string, jsonOutput bool) (string, *bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()
	t.Setenv("CLAUDE_PROJECT_DIR", root)
	cmd := newPlanCmd()
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errBuf)
	args := []string{"render-html", specID}
	if jsonOutput {
		args = append(args, "--json")
	}
	cmd.SetArgs(args)
	// newPlanCmd's root Execution needs the parent flags bound; set --json via
	// PersistentFlags lookup so it lands correctly.
	if jsonOutput {
		cmd.PersistentFlags().Set("json", "true")
	}
	return specID, out, errBuf, cmd.Execute()
}

// TestPlanCmd_RegisteredOnRoot verifies AC-WIRE-003: `moai plan` is a real
// Cobra subcommand findable from the root command tree, and `render-html` is a
// registered child.
func TestPlanCmd_RegisteredOnRoot(t *testing.T) {
	// rootCmd is the package-level cmd; locate `plan` and `plan render-html`.
	plan, _, _ := rootCmd.Find([]string{"plan"})
	if plan == nil || !strings.HasPrefix(plan.Use, "plan") {
		t.Fatalf("root command tree does not resolve 'plan'; got: %+v", plan)
	}
	renderHTML, _, _ := rootCmd.Find([]string{"plan", "render-html"})
	if renderHTML == nil || !strings.HasPrefix(renderHTML.Use, "render-html") {
		t.Fatalf("root command tree does not resolve 'plan render-html'; got: %+v", renderHTML)
	}
}

// TestPlanRenderHTML_HelpExitsZero verifies AC-WIRE-003 help-text contract:
// `moai plan --help` and `moai plan render-html --help` both exit 0, and the
// child help names the <SPEC-ID> positional argument.
func TestPlanRenderHTML_HelpExitsZero(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"render-html", "--help"}} {
		buf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd := newPlanCmd()
		cmd.SetOut(buf)
		cmd.SetErr(errBuf)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Errorf("plan %v --help returned error: %v; stderr=%s", args, err, errBuf.String())
		}
	}
	// Child help names the SPEC-ID positional.
	childBuf := &bytes.Buffer{}
	child := newPlanCmd()
	child.SetOut(childBuf)
	child.SetErr(childBuf)
	child.SetArgs([]string{"render-html", "--help"})
	_ = child.Execute()
	if !strings.Contains(childBuf.String(), "SPEC-ID") {
		t.Errorf("render-html --help does not name SPEC-ID; help:\n%s", childBuf.String())
	}
}

// TestPlanRenderHTML_WritesFileAndDOM verifies AC-WIRE-004 end-to-end: the CLI
// invoked against a fixture SPEC dir + review file writes
// .moai/reports/plan-html/<SPEC-ID>-plan.html whose DOM shows the goal, the
// 8-field contract, the verdict score, must-pass rows, and milestones.
func TestPlanRenderHTML_WritesFileAndDOM(t *testing.T) {
	root := t.TempDir()
	specID := "SPEC-FIXTURE-001"
	goalClause := "the plan-html report is produced on disk"
	fixtureSpecDir(t, root, specID, goalClause)
	fixtureReviewFile(t, root, specID)

	_, _, errBuf, execErr := drivePlanRenderHTML(t, root, specID, false)
	if execErr != nil {
		t.Fatalf("plan render-html failed: %v; stderr=%s", execErr, errBuf.String())
	}

	htmlPath := filepath.Join(root, ".moai", "reports", "plan-html", specID+"-plan.html")
	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read rendered plan html: %v", err)
	}
	body := parseDOMBody(t, raw)

	// (a) goal text from §A
	if !strings.Contains(body, goalClause) {
		t.Errorf("DOM body missing goal clause %q;\nbody:\n%s", goalClause, body)
	}
	// (b) 8-field contract labels present
	for _, label := range []string{"goal", "scope", "non-goals", "stopping-condition", "evidence"} {
		if !strings.Contains(body, label) {
			t.Errorf("DOM body missing contract label %q;\nbody:\n%s", label, body)
		}
	}
	// (c) verdict score
	if !strings.Contains(body, "0.85") {
		t.Errorf("DOM body missing verdict score 0.85;\nbody:\n%s", body)
	}
	// (d) must-pass row text
	if !strings.Contains(body, "must-pass row A") {
		t.Errorf("DOM body missing must-pass row A;\nbody:\n%s", body)
	}
	// (e) milestone list from plan.md §F
	if !strings.Contains(body, "M1") || !strings.Contains(body, "M2") {
		t.Errorf("DOM body missing milestones M1/M2;\nbody:\n%s", body)
	}
}

// TestPlanRenderHTML_FailOpenOnMissingReview verifies AC-WIRE-005(a): with no
// review file in .moai/reports/plan-audit/, the report is still written with
// the "audit verdict unavailable" placeholder AND exit 0.
func TestPlanRenderHTML_FailOpenOnMissingReview(t *testing.T) {
	root := t.TempDir()
	specID := "SPEC-FIXTURE-NO REVIEW"
	// Write the SPEC dir with an unusual title to avoid accidental matching.
	fixtureSpecDir(t, root, specID, "no review file present")
	// No review file.

	_, _, errBuf, execErr := drivePlanRenderHTML(t, root, specID, false)
	if execErr != nil {
		t.Fatalf("plan render-html with no review should exit 0; got err=%v; stderr=%s",
			execErr, errBuf.String())
	}
	htmlPath := filepath.Join(root, ".moai", "reports", "plan-html", specID+"-plan.html")
	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("expected report written even without review; read err: %v", err)
	}
	body := parseDOMBody(t, raw)
	if !strings.Contains(body, "audit verdict unavailable") {
		t.Errorf("DOM missing fail-open placeholder; body:\n%s", body)
	}
}

// TestPlanRenderHTML_MissingSpecDirExitsNonZero verifies AC-WIRE-005(b): with
// a SPEC-ID whose .moai/specs/<SPEC-ID>/ directory does not exist, the command
// exits non-zero, emits a diagnostic naming the SPEC-ID, and writes no html.
func TestPlanRenderHTML_MissingSpecDirExitsNonZero(t *testing.T) {
	root := t.TempDir()
	missing := "SPEC-NO-SUCH-999"

	_, _, errBuf, execErr := drivePlanRenderHTML(t, root, missing, false)
	if execErr == nil {
		t.Fatal("plan render-html on missing SPEC dir should exit non-zero")
	}
	if !strings.Contains(errBuf.String(), missing) {
		t.Errorf("stderr does not name the missing SPEC-ID %q; stderr=%s", missing, errBuf.String())
	}
	htmlPath := filepath.Join(root, ".moai", "reports", "plan-html", missing+"-plan.html")
	if _, err := os.Stat(htmlPath); err == nil {
		t.Errorf("html file should NOT have been written for missing SPEC dir; path=%s", htmlPath)
	}
}
