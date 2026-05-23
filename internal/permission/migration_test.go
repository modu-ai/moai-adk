package permission

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestMigrateLegacyBypassRules_HappyPath verifies migration of legacy bypassPermissions action -> acceptEdits.
// Related to T-RT002-30 and AC-11.
func TestMigrateLegacyBypassRules_HappyPath(t *testing.T) {
	t.Parallel()

	rules := []PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  Decision("bypassPermissions"), // legacy v2 action.
			Source:  config.SrcProject,
			Origin:  ".claude/settings.json",
		},
	}

	migrated, warnings := MigrateLegacyBypassRules(rules)
	if len(warnings) == 0 {
		t.Error("MigrateLegacyBypassRules() should return deprecation warnings")
	}
	if len(migrated) != 1 {
		t.Fatalf("MigrateLegacyBypassRules() should return 1 migrated rule, got %d", len(migrated))
	}
	if migrated[0].Action != DecisionAllow {
		t.Errorf("MigrateLegacyBypassRules() migrated action = %v, want DecisionAllow", migrated[0].Action)
	}
	// Verify the warning includes the origin file path.
	if !containsMiddle(warnings[0], ".claude/settings.json") {
		t.Errorf("warning should mention origin file, got: %s", warnings[0])
	}
}

// TestMigrateLegacyBypassRules_NoLegacy verifies no warnings when no legacy action is present.
// Related to T-RT002-30.
func TestMigrateLegacyBypassRules_NoLegacy(t *testing.T) {
	t.Parallel()

	rules := []PermissionRule{
		{
			Pattern: "Bash(go test:*)",
			Action:  DecisionAllow, // already the correct action.
			Source:  config.SrcProject,
			Origin:  ".claude/settings.json",
		},
	}

	migrated, warnings := MigrateLegacyBypassRules(rules)
	if len(warnings) != 0 {
		t.Errorf("MigrateLegacyBypassRules() should return no warnings for non-legacy rules, got: %v", warnings)
	}
	if len(migrated) != 1 {
		t.Fatalf("MigrateLegacyBypassRules() should return 1 rule, got %d", len(migrated))
	}
	if migrated[0].Action != DecisionAllow {
		t.Errorf("MigrateLegacyBypassRules() should not change non-legacy rule action")
	}
}

// TestMigrateLegacyBypassRules_MultipleOrigins verifies that each of multiple legacy rules names its origin.
// Related to T-RT002-30.
func TestMigrateLegacyBypassRules_MultipleOrigins(t *testing.T) {
	t.Parallel()

	rules := []PermissionRule{
		{
			Pattern: "Bash(curl:*)",
			Action:  Decision("bypassPermissions"),
			Source:  config.SrcProject,
			Origin:  "file-a.json",
		},
		{
			Pattern: "Write(*)",
			Action:  Decision("bypassPermissions"),
			Source:  config.SrcLocal,
			Origin:  "file-b.json",
		},
	}

	migrated, warnings := MigrateLegacyBypassRules(rules)
	if len(warnings) != 2 {
		t.Errorf("MigrateLegacyBypassRules() should return 2 warnings for 2 legacy rules, got %d", len(warnings))
	}
	if len(migrated) != 2 {
		t.Fatalf("MigrateLegacyBypassRules() should return 2 migrated rules, got %d", len(migrated))
	}
	for i, r := range migrated {
		if r.Action != DecisionAllow {
			t.Errorf("migrated[%d].Action = %v, want DecisionAllow", i, r.Action)
		}
	}
	// Each warning includes the corresponding file origin.
	if !containsMiddle(warnings[0], "file-a.json") && !containsMiddle(warnings[1], "file-a.json") {
		t.Error("one of the warnings should mention file-a.json")
	}
	if !containsMiddle(warnings[0], "file-b.json") && !containsMiddle(warnings[1], "file-b.json") {
		t.Error("one of the warnings should mention file-b.json")
	}
}
