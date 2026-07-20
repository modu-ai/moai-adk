// SPEC-HNS-PREFIX-RENAME-001 M2 — RED-phase specification tests for the v4
// lifecycle dual-pattern artifact matcher (REQ-HPR-010, REQ-HPR-011).
//
// The lifecycle verbs (list/edit/remove) MUST resolve artifacts carrying either
// the canonical `hns-<name>` or the legacy `harness-<name>` prefix — including
// mixed-generation harnesses whose artifacts straddle both prefixes.
//
// @MX:NOTE: [AUTO] AC-HPR-006 specification — dual-pattern artifact matching +
// mixed-generation removal + longest-name-first shadowing on the hns- branch.
// @MX:SPEC: SPEC-HNS-PREFIX-RENAME-001 acceptance.md AC-HPR-006
package harness

import (
	"os"
	"path/filepath"
	"testing"
)

// hnsManifestJSON returns a schema-valid v4 manifest whose runner_workflow
// carries the canonical hns- prefix.
func hnsManifestJSON(name string) string {
	return `{
  "name": "` + name + `",
  "domain": "hns test domain",
  "source_request": "build a harness for ` + name + `",
  "patterns": ["Pipeline"],
  "specialists": [
    {"role": "core", "primitive": "sub-agent", "isolation": "none", "effort": "low", "model": "haiku"}
  ],
  "sprint_contract": {
    "dimensions": ["correctness"],
    "thresholds": {"correctness": 0.9}
  },
  "entry_command": "/harness:` + name + `",
  "runner_workflow": "hns-` + name + `-run.js"
}`
}

// TestHarnessArtifactBelongsTo_HNSDualPattern pins the dual-pattern matcher:
// both hns-<name>* and harness-<name>* entry names belong to harness <name>,
// and the longest-name-first shadowing discipline extends to the hns- branch
// (edge case §E.4: hns-release-update-* is never claimed by "release").
func TestHarnessArtifactBelongsTo_HNSDualPattern(t *testing.T) {
	t.Parallel()

	allNames := []string{"release", "release-update", "acme"}

	tests := []struct {
		name      string
		entryName string
		harness   string
		want      bool
	}{
		// Canonical hns- generation
		{"hns exact", "hns-acme", "acme", true},
		{"hns role suffix", "hns-acme-core-specialist.md", "acme", true},
		{"hns verify skill", "hns-acme-verify", "acme", true},

		// Legacy harness- generation (no regression)
		{"legacy exact", "harness-acme", "acme", true},
		{"legacy role suffix", "harness-acme-core-specialist.md", "acme", true},

		// Longest-name-first shadowing on the hns- branch (§E.4)
		{"hns release-update not claimed by release", "hns-release-update-auditor-specialist.md", "release", false},
		{"hns release-update claimed by release-update", "hns-release-update-auditor-specialist.md", "release-update", true},
		{"hns release claimed by release", "hns-release-auditor-specialist.md", "release", true},

		// Cross-generation shadowing: legacy artifact name vs hns query name
		{"legacy release-update not claimed by release", "harness-release-update-auditor-specialist.md", "release", false},

		// Non-matching prefixes
		{"moai-harness template skill", "moai-harness-learner", "acme", false},
		{"hnsx not hns- prefixed", "hnsx-acme-specialist.md", "acme", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := harnessArtifactBelongsTo(tt.entryName, tt.harness, allNames)
			if got != tt.want {
				t.Errorf("harnessArtifactBelongsTo(%q, %q, %v) = %v, want %v",
					tt.entryName, tt.harness, allNames, got, tt.want)
			}
		})
	}
}

// TestListHarnesses_HNSManifest verifies list enumerates a harness whose
// manifest declares the canonical hns-<name>-run.js Runner (AC-HPR-006 list leg).
func TestListHarnesses_HNSManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	commandsDir := filepath.Join(root, ".claude", "commands", "harness")
	writeFile(t, filepath.Join(commandsDir, "acme.md"),
		"---\ndescription: harness acme\n---\nRun harness acme\n")
	writeFile(t, filepath.Join(commandsDir, "acme", "manifest.json"), hnsManifestJSON("acme"))

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	if entries[0].Name != "acme" {
		t.Errorf("Name = %q, want %q", entries[0].Name, "acme")
	}
	if entries[0].RunnerWorkflow != "hns-acme-run.js" {
		t.Errorf("RunnerWorkflow = %q, want %q", entries[0].RunnerWorkflow, "hns-acme-run.js")
	}
}

// TestEditHarness_HNSArtifacts verifies edit collects hns- prefixed specialist
// files and companion skill directories.
func TestEditHarness_HNSArtifacts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	commandsDir := filepath.Join(root, ".claude", "commands", "harness")
	writeFile(t, filepath.Join(commandsDir, "acme.md"), "# thin\n")
	writeFile(t, filepath.Join(commandsDir, "acme", "manifest.json"), hnsManifestJSON("acme"))
	writeFile(t, filepath.Join(root, ".claude", "agents", "harness", "hns-acme-core-specialist.md"),
		"---\nname: hns-acme-core-specialist\n---\nspecialist\n")
	writeFile(t, filepath.Join(root, ".claude", "skills", "hns-acme-verify", "SKILL.md"),
		"# hns-acme-verify skill\n")

	paths, err := EditHarness(root, "acme")
	if err != nil {
		t.Fatalf("EditHarness: %v", err)
	}
	if len(paths.SpecialistPaths) != 1 {
		t.Errorf("SpecialistPaths = %v, want exactly the hns- specialist", paths.SpecialistPaths)
	}
	if len(paths.SkillPaths) != 1 {
		t.Errorf("SkillPaths = %v, want exactly the hns- companion skill dir", paths.SkillPaths)
	}
}

// TestRemoveHarness_MixedGeneration is the AC-HPR-006 remove leg + §D.1
// Scenario 3: a harness whose artifacts straddle generations (hns- Runner via
// manifest + legacy harness- specialist + hns- skill) is removed atomically —
// artifacts of BOTH prefixes belonging to the harness are removed, and
// artifacts of OTHER harnesses (either prefix) are untouched.
func TestRemoveHarness_MixedGeneration(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	commandsDir := filepath.Join(root, ".claude", "commands", "harness")

	// Target harness "acme" — mixed generations.
	writeFile(t, filepath.Join(commandsDir, "acme.md"), "# thin\n")
	writeFile(t, filepath.Join(commandsDir, "acme", "manifest.json"), hnsManifestJSON("acme"))
	runnerPath := filepath.Join(root, ".claude", "workflows", "hns-acme-run.js")
	writeFile(t, runnerPath, "// hns runner\n")
	legacySpecialist := filepath.Join(root, ".claude", "agents", "harness", "harness-acme-core-specialist.md")
	writeFile(t, legacySpecialist, "# legacy specialist\n")
	hnsSkillDir := filepath.Join(root, ".claude", "skills", "hns-acme-verify")
	writeFile(t, filepath.Join(hnsSkillDir, "SKILL.md"), "# hns skill\n")

	// Bystander harness "other" (legacy generation) — must survive.
	writeFile(t, filepath.Join(commandsDir, "other.md"), "# thin\n")
	otherSpecialist := filepath.Join(root, ".claude", "agents", "harness", "harness-other-core-specialist.md")
	writeFile(t, otherSpecialist, "# other specialist\n")
	otherSkillDir := filepath.Join(root, ".claude", "skills", "hns-other-verify")
	writeFile(t, filepath.Join(otherSkillDir, "SKILL.md"), "# other skill\n")

	if err := RemoveHarness(root, "acme"); err != nil {
		t.Fatalf("RemoveHarness(acme): %v", err)
	}

	// All acme artifacts (both generations) removed.
	for _, p := range []string{
		filepath.Join(commandsDir, "acme.md"),
		filepath.Join(commandsDir, "acme"),
		runnerPath,
		legacySpecialist,
		hnsSkillDir,
	} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still exists after RemoveHarness (want removed)", p)
		}
	}

	// Bystander artifacts untouched.
	for _, p := range []string{
		filepath.Join(commandsDir, "other.md"),
		otherSpecialist,
		filepath.Join(otherSkillDir, "SKILL.md"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("bystander %s missing after RemoveHarness(acme): %v", p, err)
		}
	}
}
