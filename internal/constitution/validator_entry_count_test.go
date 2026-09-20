package constitution_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// The validate result must carry how many entries were actually checked, so the
// CLI can report a measured number instead of a literal. Card t988.
//
// The empty-registry case is the discriminator: a constant reported count is
// indistinguishable from a real count until one input makes the two disagree.
// A registry with entries must report a non-zero count, and an empty registry
// must report zero — only a real count satisfies both at once.

// TestValidateCountsCheckedEntries reports one checked entry per live registry entry.
func TestValidateCountsCheckedEntries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", "# Rules\n\nfirst clause here.\nsecond clause here.\n")

	regContent := `- id: CONST-V3R2-001
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "first clause here."
  canary_gate: false

- id: CONST-V3R2-002
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "second clause here."
  canary_gate: false
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if result.Status != constitution.ValidateStatusOK {
		t.Fatalf("Status = %q, want %q; entries: %v", result.Status, constitution.ValidateStatusOK, result.Entries)
	}
	if result.CheckedCount != 2 {
		t.Errorf("CheckedCount = %d, want 2", result.CheckedCount)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

// TestValidateCountsZeroOnEmptyRegistry is the control that separates a real
// count from a hardcoded one: here zero is the correct answer.
func TestValidateCountsZeroOnEmptyRegistry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	regPath := writeRegistryInDir(t, dir, "")

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if result.CheckedCount != 0 {
		t.Errorf("CheckedCount = %d, want 0 on an empty registry", result.CheckedCount)
	}
	if result.TotalCount != 0 {
		t.Errorf("TotalCount = %d, want 0 on an empty registry", result.TotalCount)
	}
}

// TestValidateExcludesRetiredFromCheckedCount keeps the two counts distinct: a
// retired entry is listed but never checked, so it belongs to TotalCount only.
func TestValidateExcludesRetiredFromCheckedCount(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", "# Rules\n\nlive clause here.\n")

	regContent := `- id: CONST-V3R2-001
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "live clause here."
  canary_gate: false

- id: CONST-V3R2-002
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "[SUPERSEDED by CONST-V3R2-001] withdrawn clause."
  canary_gate: false
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if result.CheckedCount != 1 {
		t.Errorf("CheckedCount = %d, want 1 (the retired entry is not checked)", result.CheckedCount)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2 (the retired entry is still listed)", result.TotalCount)
	}
	if result.RetiredCount != 1 {
		t.Errorf("RetiredCount = %d, want 1", result.RetiredCount)
	}
	if got, want := result.CheckedCount+result.RetiredCount, result.TotalCount; got != want {
		t.Errorf("CheckedCount+RetiredCount = %d, want TotalCount %d — the two counts must reconcile", got, want)
	}
}
