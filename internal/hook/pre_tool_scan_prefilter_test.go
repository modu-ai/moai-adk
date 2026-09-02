package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// SPEC-SEC-SCAN-SURFACE-001 M2 — the end-to-end consequences of the literal
// pre-filter, counted on the same instrument M1 uses: ScanFile is the only
// route from this path to an `sg` spawn, so a ScanFile count of 0 proves a
// spawn count of 0.

// jsCleanPayload carries none of the tokens derived for javascript from the
// shipped ruleset — {AIza, AKIA, child_process.exec, cp.exec, ghp_, sk-, xox}.
const jsCleanPayload = "function add(a, b) {\n  return a + b;\n}\n\nmodule.exports = { add };\n"

// TestScanWriteContentPrefilterSkip closes AC-SSS-011: a payload no
// error-severity rule can match is not scanned at all.
// Pre-implementation measurement: 1 ScanFile call.
//
// The language is javascript rather than go deliberately: spec.md §C.3 measured
// the Go skip rate at 0.9% against javascript's 96.3%, so a Go-based version of
// this criterion would assert the thing that measurement says will not happen.
func TestScanWriteContentPrefilterSkip(t *testing.T) {
	root := projectRootWithShippedRuleset(t)

	t.Run("payload with no error-rule token is skipped", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: root}

		decision, reason := h.scanWriteContent(context.Background(), writePayload(t, "add.js", jsCleanPayload))

		if fake.calls != 0 {
			t.Errorf("expected 0 ScanFile calls for a payload no error rule can match, got %d", fake.calls)
		}
		if decision != "" {
			t.Errorf("expected allow, got decision=%q reason=%q", decision, reason)
		}
	})

	// Control: the same language, same ruleset, a payload that DOES carry a
	// token still scans — so the zero above is the pre-filter working rather
	// than javascript never being scanned.
	t.Run("control: a payload carrying a token still scans", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: root}

		content := "const cp = require('child_process');\n\nfunction run(x) {\n  cp.exec(x);\n}\n"
		h.scanWriteContent(context.Background(), writePayload(t, "run.js", content))

		if fake.calls != 1 {
			t.Errorf("expected 1 ScanFile call for a payload carrying cp.exec, got %d", fake.calls)
		}
	})
}

// TestScanWriteContentUnderivableEscalates closes the end-to-end half of
// AC-SSS-010: a ruleset that makes a covered language underivable escalates to
// a scan even for a payload containing no dangerous construct.
// Pre-implementation measurement: 1 ScanFile call — this is the fail-open
// behaviour M2 must not break, and the control below is what makes the arm
// meaningful.
func TestScanWriteContentUnderivableEscalates(t *testing.T) {
	// Both roots carry a javascript-only ruleset. They differ in one rule
	// shape: the underivable arm's second rule is a `kind:`-only rule, which
	// the extraction table cannot read.
	derivableRule := "id: js-exec-probe\nlanguage: javascript\nseverity: error\nmessage: probe\nrule:\n  pattern: cp.exec($CMD)\n"
	underivableRule := "id: js-kind-probe\nlanguage: javascript\nseverity: error\nmessage: probe\nrule:\n  kind: string\n"

	newRoot := func(t *testing.T, rules string) string {
		t.Helper()
		root := t.TempDir()
		dir := filepath.Join(root, ".moai", "config", "astgrep-rules", "js")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "probe.yml"), []byte(rules), 0o644); err != nil {
			t.Fatalf("write rule: %v", err)
		}
		cfg := filepath.Join(root, ".moai", "config", "astgrep-rules", "sgconfig.yml")
		if err := os.WriteFile(cfg, []byte("ruleDirs:\n  - js\n"), 0o644); err != nil {
			t.Fatalf("write sgconfig: %v", err)
		}
		return root
	}

	t.Run("underivable language escalates", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: newRoot(t, derivableRule+"---\n"+underivableRule)}

		h.scanWriteContent(context.Background(), writePayload(t, "add.js", jsCleanPayload))

		if fake.calls != 1 {
			t.Errorf("expected 1 ScanFile call (fail-open escalation), got %d", fake.calls)
		}
	})

	t.Run("control: the same payload with only the derivable rule is skipped", func(t *testing.T) {
		fake := &countingScanFacade{}
		h := &preToolHandler{scanner: fake, projectDir: newRoot(t, derivableRule)}

		h.scanWriteContent(context.Background(), writePayload(t, "add.js", jsCleanPayload))

		if fake.calls != 0 {
			t.Errorf("expected 0 ScanFile calls once every rule for the language is derivable, got %d", fake.calls)
		}
	})
}
