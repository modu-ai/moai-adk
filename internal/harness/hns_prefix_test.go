// SPEC-HNS-PREFIX-RENAME-001 M2 — RED-phase specification tests for the
// harness protection surfaces (REQ-HPR-009):
//   - frozen_guard.go allowedPrefixes gains .claude/skills/hns-
//   - prefix_conflict.go recognizes hns- alongside both legacy prefixes
//
// @MX:NOTE: [AUTO] AC-HPR-007 specification — hns- recognition on the
// frozen-guard allow list and the prefix-conflict detector.
// @MX:SPEC: SPEC-HNS-PREFIX-RENAME-001 acceptance.md AC-HPR-007
package harness

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFrozenGuard_HNSAllowed pins the hns- skill prefix as an allowed write
// target for the meta-harness write guard (canonical generation), while the
// FROZEN moai-managed prefixes remain rejected.
func TestFrozenGuard_HNSAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    bool
		wantErr bool
	}{
		// Canonical hns- generation (NEW)
		{"hns skill allowed", ".claude/skills/hns-acme-verify/SKILL.md", true, false},

		// Legacy generations (no regression)
		{"legacy harness skill allowed", ".claude/skills/harness-acme/SKILL.md", true, false},
		{"legacy my-harness skill allowed", ".claude/skills/my-harness-acme/SKILL.md", true, false},
		{"harness agents dir allowed", ".claude/agents/harness/hns-acme-core-specialist.md", true, false},

		// FROZEN moai-managed prefixes still rejected
		{"moai- skill frozen", ".claude/skills/moai-foundation-core/SKILL.md", false, true},

		// Neutral: hnsx- is NOT an allowed prefix (byte-exact hns- HasPrefix)
		{"hnsx skill neutral", ".claude/skills/hnsx-foo/SKILL.md", false, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := IsAllowedPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("IsAllowedPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("IsAllowedPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestDetectPrefixConflicts_HNSPrefix pins the hns- generation on the
// prefix-conflict detector: an hns-<suffix> skill colliding with a
// moai-<suffix> skill MUST be flagged, with the hns- prefix trimmed correctly.
func TestDetectPrefixConflicts_HNSPrefix(t *testing.T) {
	t.Parallel()

	skillsDir := t.TempDir()
	for _, dir := range []string{"hns-foundation-core", "moai-foundation-core", "hns-unrelated-thing"} {
		if err := os.MkdirAll(filepath.Join(skillsDir, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	conflicts, err := DetectPrefixConflicts(skillsDir)
	if err != nil {
		t.Fatalf("DetectPrefixConflicts: %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("len(conflicts) = %d, want 1; conflicts=%+v", len(conflicts), conflicts)
	}
	if conflicts[0].MyHarnessSkill != "hns-foundation-core" {
		t.Errorf("MyHarnessSkill = %q, want %q", conflicts[0].MyHarnessSkill, "hns-foundation-core")
	}
	if conflicts[0].MoaiSkill != "moai-foundation-core" {
		t.Errorf("MoaiSkill = %q, want %q", conflicts[0].MoaiSkill, "moai-foundation-core")
	}
}
