// Package cli — `moai spec ceiling` (SPEC-AUDIT-CEILING-002): the one
// user- and agent-invocable recording path for the plan-audit ceiling policy
// outcome. The R1 CLI-level test drives the compiled verb with --record
// against a temp project and asserts the written JSON — a CLI ignoring
// --record passes the help gate and the runtime tests, so only driving the
// real verb closes it.
package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/runtime"
)

// ceilingVerbProject builds a temp project whose SPEC (Tier M) already
// carries three recorded plan-audit rounds — above the Tier M ceiling of 2 —
// with a FAIL verdict on the latest iteration, and returns its root.
func ceilingVerbProject(t *testing.T) string {
	t.Helper()
	project := t.TempDir()

	specDir := filepath.Join(project, ".moai", "specs", "SPEC-CEILFIX-001")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	specMD := "---\nid: SPEC-CEILFIX-001\ntier: M\n---\n# fixture\n"
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte(specMD), 0o644); err != nil {
		t.Fatal(err)
	}

	reports := filepath.Join(project, ".moai", "reports", "SPEC-CEILFIX-001")
	if err := os.MkdirAll(reports, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"plan-audit.md":       "verdict: FAIL\noverall_score: 0.72\nmust_pass_failed: 1\nblocking_count: 1\n",
		"plan-audit-iter1.md": "verdict: FAIL\noverall_score: 0.74\nmust_pass_failed: 1\nblocking_count: 1\n",
		"plan-audit-iter2.md": "verdict: FAIL\noverall_score: 0.75\nmust_pass_failed: 1\nblocking_count: 1\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(reports, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return project
}

// TestSpecCeilingRecordWritesJSON — AC-ACR-004's R1 arm: the verb driven with
// --record against a temp project writes the JSON record the one recording
// path defines.
func TestSpecCeilingRecordWritesJSON(t *testing.T) {
	project := ceilingVerbProject(t)
	// The verb resolves the project root from the environment (CLAUDE_PROJECT_DIR)
	// or the CWD; this test anchors both to the temp project.
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(project)

	cmd := newSpecCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("spec ceiling --record: %v\noutput: %s", err, out.String())
	}

	raw, err := os.ReadFile(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json"))
	if err != nil {
		t.Fatalf("the verb wrote no record: %v\noutput: %s", err, out.String())
	}
	var got runtime.CeilingOutcome
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("record is not the CeilingOutcome JSON: %v\nraw: %s", err, raw)
	}
	if got.Disposition != "hold" {
		t.Errorf("disposition = %q, want hold (3 rounds ≥ Tier M ceiling 2, FAIL verdict, shipped hold-and-split)", got.Disposition)
	}
	if got.Count != 3 || got.Ceiling != 2 {
		t.Errorf("count/ceiling = %d/%d, want 3/2", got.Count, got.Ceiling)
	}
	if got.VerdictLabel != "FAIL" {
		t.Errorf("verdict label = %q, want FAIL", got.VerdictLabel)
	}
	if got.SplitProposalRef == "" {
		t.Error("hold-and-split record carries no split-proposal reference")
	}
	if got.SpecID != "SPEC-CEILFIX-001" {
		t.Errorf("spec id = %q, want SPEC-CEILFIX-001", got.SpecID)
	}
	if len(got.EvidencePaths) == 0 {
		t.Error("record carries no evidence paths")
	}

	// A read-only run (no --record) after removing the record writes nothing.
	if err := os.Remove(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")); err != nil {
		t.Fatal(err)
	}
	cmd = newSpecCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("spec ceiling (read-only): %v\noutput: %s", err, out.String())
	}
	if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")); !os.IsNotExist(statErr) {
		t.Errorf("a read-only run wrote a record (stat err = %v)", statErr)
	}
}

// TestSpecCeilingRecordLandsUnderProjectNotCwd — CR-P2-1 (card-review r1):
// --record writes the outcome record under the JUDGED project root, never the
// calling cwd, when the two differ (CLAUDE_PROJECT_DIR ≠ cwd).
func TestSpecCeilingRecordLandsUnderProjectNotCwd(t *testing.T) {
	project := ceilingVerbProject(t)
	cwd := t.TempDir() // the calling directory — deliberately NOT the project
	t.Setenv("CLAUDE_PROJECT_DIR", project)
	t.Chdir(cwd)

	cmd := newSpecCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ceiling", "SPEC-CEILFIX-001", "--record"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("spec ceiling --record (cwd ≠ project): %v\noutput: %s", err, out.String())
	}

	projectRecord := filepath.Join(project, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")
	if _, err := os.Stat(projectRecord); err != nil {
		t.Fatalf("record missing under the judged project %s: %v\noutput: %s", project, err, out.String())
	}
	cwdRecord := filepath.Join(cwd, ".moai", "state", "audit-ceiling", "SPEC-CEILFIX-001.json")
	if _, err := os.Stat(cwdRecord); !os.IsNotExist(err) {
		t.Errorf("record leaked into the calling cwd %s (stat err = %v)", cwd, err)
	}
}

// TestSpecCeilingBelowCeilingNoRecord — the read path's below-ceiling arm: no
// ceiling outcome applies, and --record writes nothing.
func TestSpecCeilingBelowCeilingNoRecord(t *testing.T) {
	project := t.TempDir()
	specDir := filepath.Join(project, ".moai", "specs", "SPEC-CEILLOW-002")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("---\nid: SPEC-CEILLOW-002\ntier: M\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// No evidence directory at all: 0 rounds, below every ceiling.

	t.Setenv("CLAUDE_PROJECT_DIR", "")
	t.Chdir(project)

	cmd := newSpecCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ceiling", "SPEC-CEILLOW-002", "--record"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("spec ceiling --record (below ceiling): %v\noutput: %s", err, out.String())
	}
	if _, statErr := os.Stat(filepath.Join(project, ".moai", "state", "audit-ceiling")); !os.IsNotExist(statErr) {
		t.Errorf("a below-ceiling run created the record directory (stat err = %v)", statErr)
	}
}
