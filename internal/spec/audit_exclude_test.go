package spec

import (
	"os"
	"path/filepath"
	"testing"
)

// ExcludeSpecs skips the named SPEC directories: not counted, not audited
// (SPEC-WORKTREE-STATE-ROOT-001 REQ-WSR-012 — a SPEC shadowed by another
// catalogue is audited once, from the copy that wins).
func TestAudit_ExcludeSpecsSkipsNamedSpecs(t *testing.T) {
	base := t.TempDir()
	for _, id := range []string{"SPEC-EXA-001", "SPEC-EXB-001"} {
		dir := filepath.Join(base, ".moai", "specs", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "---\nid: " + id + "\ntitle: \"x\"\nversion: \"0.1.0\"\nstatus: draft\ncreated: 2026-01-01\n" +
			"updated: 2026-01-01\nauthor: t\npriority: P3\nphase: \"v3.0.0\"\nmodule: \"x\"\nlifecycle: spec-anchored\ntags: \"x\"\n---\n"
		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	all, err := Audit(AuditOptions{BaseDir: base})
	if err != nil || all.TotalSpecs != 2 {
		t.Fatalf("premise: both SPECs audited, got %+v err=%v", all, err)
	}
	got, err := Audit(AuditOptions{BaseDir: base, ExcludeSpecs: map[string]bool{"SPEC-EXA-001": true}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalSpecs != 1 {
		t.Errorf("TotalSpecs = %d, want 1", got.TotalSpecs)
	}
	for _, f := range got.DriftFindings {
		if f.SpecID == "SPEC-EXA-001" {
			t.Errorf("excluded SPEC produced a finding: %+v", f)
		}
	}
}
