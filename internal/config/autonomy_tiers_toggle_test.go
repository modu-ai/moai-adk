package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPEC-AUTONOMY-TIERS-001 M5 — web toggle availability (AC-002) + downgrade advisory (AC-005).

func TestTierToggleOptions_Default(t *testing.T) {
	// No proof, no kill-switch: semi-auto + automatic selectable, fully-autonomous
	// DISABLED (AC-002: fully-autonomous disabled without sandbox proof).
	opts := TierToggleOptions(false, false)
	if len(opts) != 3 {
		t.Fatalf("expected 3 toggle options, got %d", len(opts))
	}
	enabled := map[string]bool{}
	for _, o := range opts {
		enabled[o.Tier] = o.Enabled
	}
	if !enabled[AutonomyTierSemiAuto] || !enabled[AutonomyTierAutomatic] {
		t.Errorf("semi-auto/automatic must be enabled by default; got %+v", enabled)
	}
	if enabled[AutonomyTierFullyAutonomous] {
		t.Errorf("fully-autonomous MUST be disabled without sandbox proof (AC-002)")
	}
}

func TestTierToggleOptions_ProofEnablesFullyAutonomous(t *testing.T) {
	opts := TierToggleOptions(true, false)
	for _, o := range opts {
		if o.Tier == AutonomyTierFullyAutonomous && !o.Enabled {
			t.Errorf("fully-autonomous MUST be enabled when sandbox proof present (AC-002)")
		}
	}
}

func TestTierToggleOptions_KillSwitchDisablesFullyAutonomousAlways(t *testing.T) {
	// AC-005 trumps AC-002: kill-switch engaged → fully-autonomous disabled
	// EVEN WHEN sandbox proof is present.
	opts := TierToggleOptions(true, true)
	for _, o := range opts {
		if o.Tier == AutonomyTierFullyAutonomous && o.Enabled {
			t.Errorf("fully-autonomous MUST be disabled under kill-switch even with proof (AC-005 trumps AC-002)")
		}
	}
	// Lower tiers stay enabled.
	for _, o := range opts {
		if (o.Tier == AutonomyTierSemiAuto || o.Tier == AutonomyTierAutomatic) && !o.Enabled {
			t.Errorf("%s must stay enabled under kill-switch (REQ-005)", o.Tier)
		}
	}
}

func TestAppendDowngradeAdvisory(t *testing.T) {
	// AC-005: a downgrade records an advisory to .moai/logs/autonomy-downgrade.log.
	dir := t.TempDir()
	logPath := filepath.Join(dir, ".moai", "logs", "autonomy-downgrade.log")
	if err := AppendDowngradeAdvisory(logPath, AutonomyTierFullyAutonomous, AutonomyTierAutomatic, "no sandbox proof"); err != nil {
		t.Fatalf("AppendDowngradeAdvisory: %v", err)
	}
	body, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read advisory log: %v", err)
	}
	s := string(body)
	for _, want := range []string{"autonomy-downgrade", AutonomyTierFullyAutonomous, AutonomyTierAutomatic, "no sandbox proof"} {
		if !strings.Contains(s, want) {
			t.Errorf("advisory log missing %q\nlog: %s", want, s)
		}
	}
	// Append (not overwrite) — a second advisory is additive.
	if err := AppendDowngradeAdvisory(logPath, AutonomyTierFullyAutonomous, AutonomyTierAutomatic, "kill-switch"); err != nil {
		t.Fatal(err)
	}
	body2, _ := os.ReadFile(logPath)
	if !strings.Contains(string(body2), "kill-switch") {
		t.Errorf("second advisory not appended")
	}
	// The grep AC-005 verification: grep autonomy-downgrade must match.
	if !strings.Contains(string(body2), "autonomy-downgrade") {
		t.Errorf("advisory must contain the 'autonomy-downgrade' marker for the AC-005 grep")
	}
}

func TestAppendDowngradeAdvisory_CreatesParentDirs(t *testing.T) {
	// The advisory path is under .moai/logs/ which may not exist yet.
	dir := t.TempDir()
	logPath := filepath.Join(dir, "deep", "nested", "autonomy-downgrade.log")
	if err := AppendDowngradeAdvisory(logPath, AutonomyTierFullyAutonomous, AutonomyTierAutomatic, "test"); err != nil {
		t.Fatalf("AppendDowngradeAdvisory should create parent dirs: %v", err)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("advisory log not created: %v", err)
	}
}
