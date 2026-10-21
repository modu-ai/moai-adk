package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestP1844BadAlarmValueKeepsHealthySettings — PR #1844 review P1
// (PRRT_kwDOPw9rL86t5NP8): a single mis-typed alarm value
// (card_age_alarm.gate_warn_after: [6h]) must NOT reset the rest of the
// workflow config (execution_mode: team, branch_guard.enabled: true must
// survive Loader.Load).
func TestP1844BadAlarmValueKeepsHealthySettings(t *testing.T) {
	dir := t.TempDir()
	sections := filepath.Join(dir, ".moai", "config", "sections")
	if err := os.MkdirAll(sections, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := `workflow:
  execution_mode: team
  branch_guard:
    enabled: true
  card_age_alarm:
    enabled: true
    gate_warn_after: [6h]
`
	if err := os.WriteFile(filepath.Join(sections, "workflow.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := (&Loader{}).Load(filepath.Join(dir, ".moai"))
	if err != nil {
		t.Fatalf("Load with a mis-typed alarm value failed (P1 repro): %v", err)
	}
	if got := loaded.Workflow.ExecutionMode; got != "team" {
		t.Errorf("execution_mode = %q, want team — the bad alarm value reset healthy settings (P1 REPRODUCED)", got)
	}
	if !loaded.Workflow.BranchGuard.Enabled {
		t.Error("branch_guard.enabled = false, want true — the bad alarm value reset healthy settings (P1 REPRODUCED)")
	}
	if loaded.Workflow.CardAgeAlarm.GateWarnAfter != DefaultCardAgeAlarmGateWarnAfter.String() {
		t.Errorf("card_age_alarm.gate_warn_after = %q, want the compiled default %q (the alarm block must fall back alone)",
			loaded.Workflow.CardAgeAlarm.GateWarnAfter, DefaultCardAgeAlarmGateWarnAfter.String())
	}
}
