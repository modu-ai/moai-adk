// SPEC-V3R3-HARNESS-001 / T-M3-01
// RED phase: classifySkill and runSkillsCheck do not exist yet — tests fail.
// GREEN phase: implement doctor_skills.go → tests pass.

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/template"
)

// materializeEmbeddedSkills creates a .claude/skills/<name>/ directory for every
// embedded moai-* skill (reproducing a template-fresh install) under root.
func materializeEmbeddedSkills(t *testing.T, root string) []string {
	t.Helper()
	names, err := template.EmbeddedMoaiSkillNames()
	if err != nil {
		t.Fatalf("EmbeddedMoaiSkillNames: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("expected a non-empty embedded moai-* skill set")
	}
	for _, n := range names {
		if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", n), 0o755); err != nil {
			t.Fatalf("mkdir skill %s: %v", n, err)
		}
	}
	return names
}

// TestCheckSkillsAllowlist_TemplateFreshZeroUnknown is AC-DFS-004 (RED before
// fix): a project whose .claude/skills/moai-* directories were installed by the
// running binary's embedded templates must report ZERO unknown moai- skills.
// On the pre-fix static-22 allowlist this FAILs (10 embedded skills classify as
// unknown → CheckWarn).
func TestCheckSkillsAllowlist_TemplateFreshZeroUnknown(t *testing.T) {
	root := t.TempDir()
	materializeEmbeddedSkills(t, root)

	check := checkSkillsAllowlist(root, false)

	if check.Status != uikit.CheckOK {
		t.Errorf("template-fresh skills must be OK (0 unknown), got %v: %s", check.Status, check.Message)
	}
	if strings.Contains(check.Message, "unknown moai- skill") {
		t.Errorf("no unknown moai- skills expected on a template-fresh project: %q", check.Message)
	}
}

// TestClassifySkill_EmbeddedAntiDrift is AC-DFS-005 (RED before fix + anti-drift
// invariant): every embedded moai-* skill MUST classify non-WARN. This re-fails
// automatically if a future skill is added to templates without the derivation
// picking it up — the guard against re-drift.
func TestClassifySkill_EmbeddedAntiDrift(t *testing.T) {
	names, err := template.EmbeddedMoaiSkillNames()
	if err != nil {
		t.Fatalf("EmbeddedMoaiSkillNames: %v", err)
	}
	for _, n := range names {
		if got := classifySkill(n); got == "WARN" {
			t.Errorf("embedded skill %q classified as WARN — allowlist drifted from the manifest", n)
		}
	}
}

// TestCheckSkillsAllowlist_GenuineUnknownWarns is AC-DFS-006 (guard): a bogus
// moai-* directory absent from the embedded manifest still warns (count==1),
// proving the manifest derivation did NOT disable the check.
func TestCheckSkillsAllowlist_GenuineUnknownWarns(t *testing.T) {
	root := t.TempDir()
	materializeEmbeddedSkills(t, root)
	// One genuinely-unknown moai-* skill.
	if err := os.MkdirAll(filepath.Join(root, ".claude", "skills", "moai-nonexistent-xyz"), 0o755); err != nil {
		t.Fatalf("mkdir bogus skill: %v", err)
	}

	check := checkSkillsAllowlist(root, false)

	if check.Status != uikit.CheckWarn {
		t.Errorf("a genuine unknown moai- skill must WARN, got %v: %s", check.Status, check.Message)
	}
	if classifySkill("moai-nonexistent-xyz") != "WARN" {
		t.Errorf("moai-nonexistent-xyz must classify WARN")
	}
	// Exactly one unknown (the bogus dir) — every embedded skill is known.
	if !strings.Contains(check.Message, "1 unknown moai- skill") {
		t.Errorf("expected exactly 1 unknown, got message %q", check.Message)
	}
}

// TestClassifySkill verifies the classification logic for skill names.
// Table-driven, covering all 4 classification branches from REQ-HARNESS-003.
func TestClassifySkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		skillName string
		wantClass string
	}{
		{
			name:      "valid static core skill returns PASS",
			skillName: "moai-foundation-cc",
			wantClass: "PASS",
		},
		{
			name:      "unknown moai- prefixed skill returns WARN",
			skillName: "moai-custom-foo",
			wantClass: "WARN",
		},
		{
			name:      "user customization harness prefix returns INFO",
			skillName: "harness-test",
			wantClass: "INFO",
		},
		{
			name:      "empty string returns INFO (non-moai, no enforcement)",
			skillName: "",
			wantClass: "INFO",
		},
		{
			name:      "moai- prefix only (no name part) returns WARN",
			skillName: "moai-",
			wantClass: "WARN",
		},
		{
			name:      "valid static core skill moai-meta-harness returns PASS",
			skillName: "moai-meta-harness",
			wantClass: "PASS",
		}, {
			name:      "third-party skill without moai- prefix returns INFO",
			skillName: "my-custom-thing",
			wantClass: "INFO",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifySkill(tt.skillName)
			if got != tt.wantClass {
				t.Errorf("classifySkill(%q) = %q, want %q", tt.skillName, got, tt.wantClass)
			}
		})
	}
}
