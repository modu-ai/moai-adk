package constitution_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// ── RED phase: tests written before implementation ──────────────────────────
// All tests below should FAIL until validator.go is implemented (GREEN phase).
// AC-CDL-003, AC-CDL-004, AC-CDL-006, AC-CDL-008, AC-CDL-009, AC-CDL-010.

// writeRegistryInDir writes registry.md inside dir and returns its path.
func writeRegistryInDir(t *testing.T, dir, yamlContent string) string {
	t.Helper()
	content := "# Test Registry\n\n## Entries\n\n```yaml\n" + yamlContent + "\n```\n"
	path := filepath.Join(dir, "registry.md")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeRegistryInDir: %v", err)
	}
	return path
}

// writeSourceInDir writes a source file at dir/relPath.
func writeSourceInDir(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatalf("writeSourceInDir: %v", err)
	}
}

// TestValidateHappyPath returns status=ok when the registry and source file are fully in sync.
// AC-CDL-003.
func TestValidateHappyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Test Rules\n\n[ZONE:Frozen] [HARD] All user-facing questions MUST go through AskUserQuestion.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#test-rules"
  clause: "All user-facing questions MUST go through AskUserQuestion."
  canary_gate: true
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
		t.Errorf("Status = %q, want %q; entries: %v", result.Status, constitution.ValidateStatusOK, result.Entries)
	}
	if result.DriftCount != 0 {
		t.Errorf("DriftCount = %d, want 0", result.DriftCount)
	}
}

// TestValidateSourceFileMissing returns a SOURCE_FILE_MISSING error when the registered source file does not exist.
// AC-CDL-004 / REQ-CDL-013.
func TestValidateSourceFileMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: nonexistent.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	// SOURCE_FILE_MISSING exits with code 2 — error is returned
	if err == nil {
		t.Fatalf("Validate() should return error for missing source file")
	}

	var validErr *constitution.ValidationError
	if !constitution.AsValidationError(err, &validErr) {
		t.Fatalf("error type = %T, want *ValidationError", err)
	}
	if !hasEntryWithStatus(result, constitution.SentinelSourceFileMissing) {
		t.Errorf("entries do not contain SOURCE_FILE_MISSING; got: %v", result.Entries)
	}
}

// TestValidateDrift returns a DRIFT error when the clause text does not match the source.
// AC-CDL-004 / REQ-CDL-012.
func TestValidateDrift(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] All user-facing questions MUST go through AskUserQuestion.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	// Registry clause deliberately different from source
	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "All user-directed inquiries SHOULD route through AskUserQuestion."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if result.Status == constitution.ValidateStatusOK {
		t.Error("Status = ok, want drift")
	}
	if !hasEntryWithStatus(result, constitution.SentinelDrift) {
		t.Errorf("entries do not contain DRIFT; got: %v", result.Entries)
	}
}

// TestValidateFrozenWithoutCanary returns an error when a Frozen entry has canary_gate:false.
// AC-CDL-004 / REQ-CDL-015.
func TestValidateFrozenWithoutCanary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] AskUserQuestion is used ONLY by MoAI orchestrator.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	// Frozen entry with canary_gate: false — violation
	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "AskUserQuestion is used ONLY by MoAI orchestrator."
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

	if !hasEntryWithStatus(result, constitution.SentinelFrozenWithoutCanary) {
		t.Errorf("entries do not contain FROZEN_WITHOUT_CANARY; got: %v", result.Entries)
	}
}

// TestValidateInvalidZoneClass returns an error when zone_class is outside the 4-value enum.
// AC-CDL-006.
func TestValidateInvalidZoneClass(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Some rule.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: invalid-class
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if !hasEntryWithStatus(result, constitution.SentinelInvalidZoneClass) {
		t.Errorf("entries do not contain INVALID_ZONE_CLASS; got: %v", result.Entries)
	}
}

// TestValidateDuplicateZoneMarkerWarning returns a warning when two ZONE markers appear on the same line.
// EC-CDL-007.
func TestValidateDuplicateZoneMarkerWarning(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Line has two ZONE markers
	sourceContent := "# Rules\n\n[ZONE:Frozen] [ZONE:Evolvable] [HARD] Some rule.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	// Check that DUPLICATE_ZONE_MARKER warning is present
	hasWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, constitution.SentinelDuplicateZoneMarker) {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Errorf("DUPLICATE_ZONE_MARKER warning not found in warnings: %v", result.Warnings)
	}
}

// TestValidateReadOnly verifies that the validator does not write files.
// AC-CDL-010.
func TestValidateReadOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Some rule text.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule text."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	// Make the source file read-only (but not the dir, to allow the registry read)
	sourcePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.Chmod(sourcePath, 0o444); err != nil {
		t.Skipf("cannot make source read-only: %v", err)
	}
	if err := os.Chmod(regPath, 0o444); err != nil {
		t.Skipf("cannot make registry read-only: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(sourcePath, 0o644)
		_ = os.Chmod(regPath, 0o644)
	})

	// Validator should work fine even with read-only files — it only reads
	_, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	// Any permission-denied error means the validator tried to write — fail
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		t.Errorf("Validate() caused permission-denied error — validator must be read-only: %v", err)
	}
}

// TestValidateSkipOverride verifies that validation is bypassed when MOAI_CONSTITUTION_SKIP_VALIDATE=1.
// AC-CDL-009.
func TestValidateSkipOverride(t *testing.T) {
	// Non-parallel: uses t.Setenv
	dir := t.TempDir()

	// Source file with drift (clause mismatch)
	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Actual rule text in source.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Different text that would cause drift."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	// Without override: should detect drift
	resultWithDrift, errNoOverride := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if errNoOverride != nil {
		t.Fatalf("Validate() without override unexpected error: %v", errNoOverride)
	}
	if resultWithDrift.Status == constitution.ValidateStatusOK {
		t.Error("Without override: expected drift, got ok")
	}

	// With override: should bypass and return ok
	t.Setenv("MOAI_CONSTITUTION_SKIP_VALIDATE", "1")
	resultSkipped, errOverride := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if errOverride != nil {
		t.Fatalf("Validate() with override unexpected error: %v", errOverride)
	}
	if resultSkipped.Status != constitution.ValidateStatusSkipped {
		t.Errorf("With override: Status = %q, want %q", resultSkipped.Status, constitution.ValidateStatusSkipped)
	}
	if !resultSkipped.Skipped {
		t.Error("With override: Skipped should be true")
	}
}

// TestValidateWhitespaceNormalization verifies that multiple whitespace characters in a clause are normalized.
// EC-CDL-002.
func TestValidateWhitespaceNormalization(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Source has double spaces
	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD]  All  user-facing  questions  MUST.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	// Registry has normalized single spaces
	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "All user-facing questions MUST."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	// Should NOT detect drift after whitespace normalization
	if hasEntryWithStatus(result, constitution.SentinelDrift) {
		t.Errorf("DRIFT detected due to whitespace differences — normalization should prevent this; entries: %v", result.Entries)
	}
}

// TestValidateReflectsUpdatesWithoutRestart reflects new entries on re-invocation after the registry is modified.
// AC-CDL-008 / REQ-CDL-009.
func TestValidateReflectsUpdatesWithoutRestart(t *testing.T) {
	// Non-parallel: modifies file in place
	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Evolvable] [HARD] Some rule.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: false
`
	regPath := writeRegistryInDir(t, dir, regContent)

	// First call
	result1, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("First Validate(): %v", err)
	}
	if result1.Status != constitution.ValidateStatusOK {
		t.Errorf("First call: Status = %q, want ok", result1.Status)
	}

	// Add a new valid entry to the registry
	extra := `
- id: CONST-V3R5-001
  zone: Evolvable
  zone_class: evolvable-tuning
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: false
`
	// Rewrite the file with an extra entry
	newContent := "# Test Registry\n\n## Entries\n\n```yaml\n" + regContent + extra + "\n```\n"
	if err := os.WriteFile(regPath, []byte(newContent), 0o600); err != nil {
		t.Fatalf("rewrite registry: %v", err)
	}

	// Second call without restart — should load fresh from disk
	result2, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Second Validate(): %v", err)
	}
	// Verify no caching issue: second call should also succeed without error
	if result2.Status != constitution.ValidateStatusOK {
		t.Errorf("Second call: Status = %q, want ok; entries: %v", result2.Status, result2.Entries)
	}
}

// TestValidateCodeFenceExclusion verifies that [HARD] inside a markdown code fence is excluded from the count.
// EC-CDL-005.
func TestValidateCodeFenceExclusion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Source has [HARD] in a code block (should be excluded) and one real rule
	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Real rule.\n\n```\n[HARD] Example in code block — should not count.\n```\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Real rule."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	// Should NOT report ZONE_UNREGISTERED for the code-fence [HARD]
	for _, e := range result.Entries {
		if e.SentinelKey == constitution.SentinelZoneUnregistered {
			t.Errorf("ZONE_UNREGISTERED reported — code-fence [HARD] was incorrectly counted: %v", e)
		}
	}
}

// TestValidateStrictModeWithDrift returns status=drift when drift is detected.
func TestValidateStrictModeWithDrift(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Actual clause.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Different clause that won't match."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
		Strict:       true,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	if result.Status == constitution.ValidateStatusOK {
		t.Error("Status = ok with drift, want drift")
	}
}

// TestValidateJSONFormat verifies that Validate returns a JSON-serializable result.
func TestValidateJSONFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	sourceContent := "# Rules\n\n[ZONE:Frozen] [HARD] Some rule.\n"
	writeSourceInDir(t, dir, "CLAUDE.md", sourceContent)

	regContent := `- id: CONST-V3R2-001
  zone: Frozen
  zone_class: frozen-canonical
  file: CLAUDE.md
  anchor: "#rules"
  clause: "Some rule."
  canary_gate: true
`
	regPath := writeRegistryInDir(t, dir, regContent)

	result, err := constitution.Validate(constitution.ValidateOptions{
		RegistryPath: regPath,
		ProjectDir:   dir,
	})
	if err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}

	// Verify struct has required fields for JSON serialization per AC-CDL-003
	if result.Status == "" {
		t.Error("result.Status is empty")
	}
	_ = result.DriftCount
	_ = result.MissingCount
	_ = result.UnregisteredCount
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// hasEntryWithStatus reports whether ValidationResult.Entries contains the specified sentinel key.
func hasEntryWithStatus(result constitution.ValidationResult, key string) bool {
	for _, e := range result.Entries {
		if e.SentinelKey == key {
			return true
		}
	}
	return false
}
