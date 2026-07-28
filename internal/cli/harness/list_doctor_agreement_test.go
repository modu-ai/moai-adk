// Package harness — AC-HLR-013: list and doctor agreement on a command-only
// thin harness (SPEC-HARNESS-LOOP-REPAIR-001 M6).
//
// A command-only thin harness (a .claude/commands/harness/<name>.md file with
// no co-located manifest.json / Runner) is an EXPECTED state, not a defect:
// doctor classifies it as SeverityInfo ("command-only thin harness ... —
// Runner/agent axes not applicable"). The `list` verb MUST describe the same
// state in non-defect-suggesting language so the two commands agree. Before M6,
// `list` printed "(manifest missing)" for the domain cell — defect-suggesting
// language that contradicts doctor's INFO classification.
package harness

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestHarnessListAndDoctor_AgreeOnCommandOnlyThinHarness verifies AC-HLR-013:
// for a command-only thin harness, `list` does NOT describe the state in
// defect-suggesting terms while `doctor` classifies the SAME state as expected
// (INFO). Both commands agree on the "command-only / expected" framing.
//
// Falsification (acceptance.md §H.3): revert the list domain-cell change so it
// prints "(manifest missing)" again → the defect-suggesting-word assertion
// fails. Confirmed RED before the alignment change landed.
func TestHarnessListAndDoctor_AgreeOnCommandOnlyThinHarness(t *testing.T) {
	t.Parallel()

	// Fixture: a command-only thin harness (command file, NO manifest, NO Runner).
	root := t.TempDir()
	writeFile(t, root+"/.claude/commands/harness/thinh.md",
		"---\ndescription: thin harness\n---\nRun thin\n")

	// --- doctor half: classifies the thin harness as INFO (expected), NOT ERROR ---
	// Uses --json so the assertion is on the structured report (error_count),
	// not on the word "error" appearing in a plaintext summary line.
	docCmd := NewHarnessDoctorCmd()
	var docOut bytes.Buffer
	docCmd.SetOut(&docOut)
	docCmd.SetErr(&bytes.Buffer{})
	docCmd.SetArgs([]string{"--json", "--project-root", root})
	if err := docCmd.Execute(); err != nil {
		t.Fatalf("doctor on thin harness returned error (should exit 0 for INFO-only): %v", err)
	}
	var report DoctorReport
	if err := json.Unmarshal(docOut.Bytes(), &report); err != nil {
		t.Fatalf("doctor --json output is not valid JSON: %v\n%s", err, docOut.String())
	}
	if report.ErrorCount != 0 {
		t.Errorf("doctor error_count = %d, want 0 (thin harness must not be a defect); findings=%+v", report.ErrorCount, report.Findings)
	}
	if report.InfoCount < 1 {
		t.Errorf("doctor info_count = %d, want >= 1 (thin harness INFO classification)", report.InfoCount)
	}
	// The INFO finding MUST frame the state as "command-only" (the expected-state
	// vocabulary), confirming doctor treats the thin harness as non-defect.
	infoFramesCommandOnly := false
	for _, f := range report.Findings {
		if f.Severity == SeverityInfo && strings.Contains(strings.ToLower(f.Message), "command-only") {
			infoFramesCommandOnly = true
		}
	}
	if !infoFramesCommandOnly {
		t.Errorf("doctor INFO finding does not frame the thin harness as command-only; findings=%+v", report.Findings)
	}

	// --- list half: must NOT describe the thin harness in defect-suggesting terms ---
	// `list` inherits --project-root from the parent harness command's persistent
	// flag (NewHarnessV4ListCmd defines only --json locally). Attach it to a
	// minimal parent so the inherited flag is available — mirrors how
	// newHarnessRouterCmd() mounts the verb in production.
	parent := &cobra.Command{Use: "harness"}
	parent.PersistentFlags().String("project-root", "", "project root path")
	listCmd := NewHarnessV4ListCmd()
	parent.AddCommand(listCmd)
	var listOut bytes.Buffer
	parent.SetOut(&listOut)
	parent.SetErr(&bytes.Buffer{})
	parent.SetArgs([]string{"list", "--project-root", root})
	if err := parent.Execute(); err != nil {
		t.Fatalf("list on thin harness returned error: %v", err)
	}
	listStr := listOut.String()
	// "missing" is the defect-suggesting word the pre-M6 list used ("(manifest
	// missing)"). After alignment with doctor's INFO framing, list MUST NOT use it.
	if strings.Contains(strings.ToLower(listStr), "missing") {
		t.Errorf("AC-HLR-013: list describes the thin harness in defect-suggesting terms (contains \"missing\");\n"+
			"doctor classifies the SAME state as expected (INFO). The two must agree.\nlist output: %s", listStr)
	}
	// Both commands agree: each frames the state as command-only / expected.
	if !strings.Contains(strings.ToLower(listStr), "command-only") {
		t.Errorf("AC-HLR-013: list does not frame the thin harness as command-only (doctor does);\nlist output: %s", listStr)
	}
}
