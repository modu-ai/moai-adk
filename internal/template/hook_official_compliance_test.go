// hook_official_compliance_test.go: SPEC-HOOK-OFFICIAL-COMPLIANCE-001 M1
// characterization tests for the blocking-gate JSON contract, PreToolUse
// exit-code passthrough, and SessionStart probe hookEventName.
//
// These tests read the TEMPLATE hook files (the single source of truth under
// internal/template/templates/) and assert the official-doc-compliant output
// shape. They are the durable RED→GREEN driver for AC-HOC-001..005.
//
// Sentinel on failure: HOOK_OFFICIAL_COMPLIANCE_M1
// Origin: SPEC-HOOK-OFFICIAL-COMPLIANCE-001 M1 (HIGH severity).
//
// @MX:ANCHOR: [AUTO] M1 hook-contract characterization guard — fan_in=5 ACs route through here.
// @MX:REASON: Without this guard, a regression to the non-compliant decision:block / hardcoded
// exit 0 / missing hookEventName shape would silently disable the blocking quality gates
// under a runtime that tightens JSON-schema validation.
package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hocProjectRoot walks up to the directory containing go.mod (same pattern as
// findProjectRootForMirrorTest in rule_template_mirror_test.go).
func hocProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found; cannot determine project root")
		}
		dir = parent
	}
}

// hocReadHook reads a template hook file relative to the repo root and returns
// its content. Skips the test if the file is absent (forward-compatible).
func hocReadHook(t *testing.T, rel string) string {
	t.Helper()
	root := hocProjectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "internal", "template", "templates", ".claude", "hooks", "moai", rel))
	if err != nil {
		t.Skipf("hook template not found: %s: %v", rel, err)
	}
	return string(data)
}

// AC-HOC-001 (REQ-HOC-001, GAP-GD-01/STT-02): team-ac-verify TaskCompleted
// reject path MUST emit {"continue":false,"stopReason":...} — NOT
// {"decision":"block",...}. The "decision" field is not documented for the
// TaskCompleted event.
func TestHookOfficialCompliance_AC001_TeamACVerifyRejectShape(t *testing.T) {
	src := hocReadHook(t, "team-ac-verify.sh")

	// Locate the --reject branch printf (the line carrying BOTH printf and ledger_note).
	var rejectPrintf string
	for ln := range strings.SplitSeq(src, "\n") {
		if strings.Contains(ln, "ledger_note") && strings.Contains(ln, "printf") {
			rejectPrintf = ln
			break
		}
	}
	if rejectPrintf == "" {
		t.Fatal("AC-001: no --reject printf carrying ledger_note found in team-ac-verify.sh")
	}

	if !strings.Contains(rejectPrintf, `"continue":false`) && !strings.Contains(rejectPrintf, `"continue": false`) {
		t.Errorf("AC-001 FAIL: reject path does not emit \"continue\":false.\n  printf: %s", strings.TrimSpace(rejectPrintf))
	}
	if !strings.Contains(rejectPrintf, "stopReason") {
		t.Errorf("AC-001 FAIL: reject path does not emit stopReason.\n  printf: %s", strings.TrimSpace(rejectPrintf))
	}
	if strings.Contains(rejectPrintf, `"decision":"block"`) || strings.Contains(rejectPrintf, `"decision": "block"`) {
		t.Errorf("AC-001 FAIL: reject path still emits decision:block (not valid for TaskCompleted).\n  printf: %s", strings.TrimSpace(rejectPrintf))
	}
}

// AC-HOC-002 (REQ-HOC-002, GAP-GD-02): sync-phase-quality-gate Stop block path
// MUST nest decision/reason inside a hookSpecificOutput object that also carries
// hookEventName:"Stop". The bare top-level {"decision":"block",...} form is
// non-compliant.
func TestHookOfficialCompliance_AC002_SyncGateStopHookSpecificOutput(t *testing.T) {
	src := hocReadHook(t, "sync-phase-quality-gate.sh")

	// The blocking printf is the one emitting decision:block on the Stop path.
	var blockPrintf string
	for ln := range strings.SplitSeq(src, "\n") {
		if strings.Contains(ln, "printf") && strings.Contains(ln, "decision") && strings.Contains(ln, "block") {
			blockPrintf = ln
			break
		}
	}
	if blockPrintf == "" {
		t.Fatal("AC-002: no blocking decision printf found in sync-phase-quality-gate.sh")
	}

	if !strings.Contains(blockPrintf, "hookSpecificOutput") {
		t.Errorf("AC-002 FAIL: block printf does not wrap decision in hookSpecificOutput.\n  printf: %s", strings.TrimSpace(blockPrintf))
	}
	if !strings.Contains(blockPrintf, `hookEventName`) || !strings.Contains(blockPrintf, "Stop") {
		t.Errorf("AC-002 FAIL: block printf missing hookEventName:\"Stop\".\n  printf: %s", strings.TrimSpace(blockPrintf))
	}
}

// AC-HOC-003 (REQ-HOC-003, GAP-TOOL-01): handle-pre-tool wrapper MUST propagate
// the moai binary's exit code on every resolution branch (replacing hardcoded
// `exit 0`), so an exit-2 reject path would reach Claude Code.
func TestHookOfficialCompliance_AC003_PreToolExitCodePassthrough(t *testing.T) {
	src := hocReadHook(t, "handle-pre-tool.sh.tmpl")
	lines := strings.Split(src, "\n")

	branches := 0
	for i, ln := range lines {
		if !strings.Contains(ln, "hook pre-tool") {
			continue
		}
		branches++
		// find the next non-blank, non-comment line
		var exitLine string
		for j := i + 1; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			exitLine = trimmed
			break
		}
		if exitLine == "" {
			t.Errorf("AC-003 FAIL: no exit line found after pre-tool invocation at line %d.\n  invocation: %s", i+1, strings.TrimSpace(ln))
			continue
		}
		if exitLine == "exit 0" {
			t.Errorf("AC-003 FAIL: hardcoded `exit 0` after pre-tool invocation at line %d (must propagate exit code).\n  invocation: %s\n  exit: %s", i+1, strings.TrimSpace(ln), exitLine)
		}
	}
	if branches == 0 {
		t.Fatal("AC-003: no `hook pre-tool` invocation found in handle-pre-tool.sh.tmpl")
	}
}

// AC-HOC-004 (REQ-HOC-003, GAP-TOOL-01): handle-pre-tool wrapper MUST preserve
// stderr as the Claude-feedback channel — the `2>>"$MOAI_HOOK_STDERR_LOG"`
// redirect MUST be removed (or made event-conditional) on the PreToolUse
// invocation lines so the reject reason reaches the model.
func TestHookOfficialCompliance_AC004_PreToolStderrPreserved(t *testing.T) {
	src := hocReadHook(t, "handle-pre-tool.sh.tmpl")

	for i, ln := range strings.Split(src, "\n") {
		if !strings.Contains(ln, "hook pre-tool") {
			continue
		}
		if strings.Contains(ln, "2>>") && strings.Contains(ln, "MOAI_HOOK_STDERR_LOG") {
			t.Errorf("AC-004 FAIL: pre-tool invocation at line %d redirects stderr to the log (swallows reject reason).\n  line: %s", i+1, strings.TrimSpace(ln))
		}
	}
}

// AC-HOC-005 (REQ-HOC-004, GAP-LC2-01): SessionStart probe MUST emit
// {"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":...}}
// — the hookEventName field is mandatory.
func TestHookOfficialCompliance_AC005_SessionStartProbeHookEventName(t *testing.T) {
	src := hocReadHook(t, "handle-session-start.sh.tmpl")

	// The probe printf is the line emitting hookSpecificOutput.additionalContext
	// via printf (exclude the descriptive comment line which mentions both tokens).
	var probePrintf string
	for ln := range strings.SplitSeq(src, "\n") {
		if strings.Contains(ln, "printf") && strings.Contains(ln, "hookSpecificOutput") && strings.Contains(ln, "additionalContext") {
			probePrintf = ln
			break
		}
	}
	if probePrintf == "" {
		t.Fatal("AC-005: no hookSpecificOutput.additionalContext printf found in handle-session-start.sh.tmpl")
	}
	if !strings.Contains(probePrintf, "hookEventName") || !strings.Contains(probePrintf, "SessionStart") {
		t.Errorf("AC-005 FAIL: probe printf missing hookEventName:\"SessionStart\".\n  printf: %s", strings.TrimSpace(probePrintf))
	}
}
