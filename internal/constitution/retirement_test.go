package constitution_test

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// retiredRegistryEntry is a registry entry whose clause carries the
// [SUPERSEDED …] retirement marker and whose text no longer exists in source.
const retiredRegistryEntry = `- id: CONST-V3R2-021
  zone: Evolvable
  zone_class: evolvable-experimental
  file: CLAUDE.md
  anchor: "#14-parallel-execution-safeguards"
  clause: "[SUPERSEDED by worktree-opt-in policy] Implementation teammates MUST use isolation: worktree when spawned via Agent()"
  canary_gate: false
`

// liveSource is a source file that does NOT contain the retired clause text.
const liveSource = "# Rules\n\n[ZONE:Evolvable] [HARD] Worktree isolation is opt-in.\n"

// TestValidateSkipsDriftForRetiredEntry verifies that a [SUPERSEDED …]-prefixed
// clause is not reported as DRIFT, so an entry can be retired without deleting it.
// Issue #1595.
func TestValidateSkipsDriftForRetiredEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", liveSource)
	regPath := writeRegistryInDir(t, dir, retiredRegistryEntry)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if result.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0; entries: %v", result.DriftCount, result.Entries)
	}
	if result.Status != constitution.ValidateStatusOK {
		t.Errorf("Status = %q, want %q; entries: %v", result.Status, constitution.ValidateStatusOK, result.Entries)
	}
	if result.RetiredCount != 1 {
		t.Errorf("RetiredCount = %d, want 1", result.RetiredCount)
	}
}

// TestValidateRetiredFrozenEntrySkipsCanaryCheck verifies that a retired Frozen
// entry may carry canary_gate:false without a FROZEN_WITHOUT_CANARY error —
// a retired clause is never amended, so shadow evaluation cannot apply.
// Issue #1595.
func TestValidateRetiredFrozenEntrySkipsCanaryCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", liveSource)
	regPath := writeRegistryInDir(t, dir, `- id: CONST-V3R2-030
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "[SUPERSEDED by CONST-V3R6-001] Some retired frozen clause."
  canary_gate: false
`)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if len(result.Entries) != 0 {
		t.Errorf("Entries = %v, want none", result.Entries)
	}
}

// TestValidateRetiredEntryToleratesMissingSource verifies that a retired entry
// whose source file no longer exists does not raise the fatal
// SOURCE_FILE_MISSING error — the entry survives purely as an audit record.
// Issue #1595.
func TestValidateRetiredEntryToleratesMissingSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	regPath := writeRegistryInDir(t, dir, `- id: CONST-V3R2-031
  zone: Evolvable
  zone_class: evolvable-tuning
  file: deleted-rule.md
  anchor: "#gone"
  clause: "[SUPERSEDED by nothing] A clause whose source file was deleted."
  canary_gate: false
`)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if result.MissingCount != 0 {
		t.Errorf("MissingCount = %d, want 0; entries: %v", result.MissingCount, result.Entries)
	}
}

// TestValidateStrictEnforcesRetiredEntry verifies that --strict ignores the
// retirement marker and checks the clause verbatim, so a maintainer can audit
// what the retired entries would report.
// Issue #1595.
func TestValidateStrictEnforcesRetiredEntry(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", liveSource)
	regPath := writeRegistryInDir(t, dir, retiredRegistryEntry)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
		Strict:       true,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if result.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1 under --strict; entries: %v", result.DriftCount, result.Entries)
	}
}

// TestValidateNonRetiredEntryStillDrifts is the control: an ordinary entry whose
// clause is absent from source is still reported, so the retirement skip does
// not widen into a blanket bypass.
func TestValidateNonRetiredEntryStillDrifts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceInDir(t, dir, "CLAUDE.md", liveSource)
	regPath := writeRegistryInDir(t, dir, `- id: CONST-V3R2-032
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "A live clause that is absent from the source file."
  canary_gate: false
`)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
	if result.DriftCount != 1 {
		t.Errorf("DriftCount = %d, want 1; entries: %v", result.DriftCount, result.Entries)
	}
	if result.RetiredCount != 0 {
		t.Errorf("RetiredCount = %d, want 0", result.RetiredCount)
	}
}

// TestIsRetiredClause pins the marker-detection boundary.
func TestIsRetiredClause(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		clause string
		want   bool
	}{
		{"bare marker", "[SUPERSEDED] text", true},
		{"marker with reason", "[SUPERSEDED by X — see Y] text", true},
		{"leading whitespace", "  [SUPERSEDED] text", true},
		{"lowercase is not a marker", "[superseded] text", false},
		{"mid-clause mention is not a marker", "mark superseded entries with [SUPERSEDED by <file>] prefix", false},
		{"unrelated bracket", "[HARD] text", false},
		{"longer word starting with the marker is not a marker", "[SUPERSEDEDLY] text", false},
		{"unterminated marker", "[SUPERSEDED text", false},
		{"a later bracket does not close the marker", "[SUPERSEDED live [HARD]", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := constitution.IsRetiredClause(tc.clause); got != tc.want {
				t.Errorf("IsRetiredClause(%q) = %v, want %v", tc.clause, got, tc.want)
			}
		})
	}
}
