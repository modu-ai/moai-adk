package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// sampleRegistryContent is a minimal registry markdown for testing.
const sampleRegistryContent = `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  file: CLAUDE.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true

- id: CONST-V3R2-003
  zone: Evolvable
  file: CLAUDE.md
  anchor: "#time-estimation"
  clause: "Never use time predictions in plans or reports."
  canary_gate: false
` + "```" + `
`

// TestConstitutionListAllEntries verifies full entry rendering of the registry.
// Related to AC-CON-001-001.
func TestConstitutionListAllEntries(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "table")
	if err != nil {
		t.Fatalf("runConstitutionList error: %v", err)
	}

	output := buf.String()
	// All 3 entries must be present in the output
	for _, id := range []string{"CONST-V3R2-001", "CONST-V3R2-002", "CONST-V3R2-003"} {
		if !strings.Contains(output, id) {
			t.Errorf("%q not found in output\noutput: %s", id, output)
		}
	}
}

// TestConstitutionListFilterFrozen verifies the --zone frozen filter.
// Direct mapping to AC-CON-001-002.
func TestConstitutionListFilterFrozen(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	frozenZone := constitution.ZoneFrozen
	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, &frozenZone, "", "table")
	if err != nil {
		t.Fatalf("runConstitutionList --zone frozen error: %v", err)
	}

	output := buf.String()

	// Only Frozen entries should be included
	if !strings.Contains(output, "CONST-V3R2-001") {
		t.Errorf("CONST-V3R2-001 not found in output")
	}
	if !strings.Contains(output, "CONST-V3R2-002") {
		t.Errorf("CONST-V3R2-002 not found in output")
	}

	// Evolvable entries should be excluded
	if strings.Contains(output, "CONST-V3R2-003") {
		t.Errorf("Evolvable entry CONST-V3R2-003 must not be in output")
	}
}

// TestConstitutionListFilterByFile verifies the --file filter.
func TestConstitutionListFilterByFile(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")

	content := `# Test Registry

## Entries

` + "```yaml" + `
- id: CONST-V3R2-001
  zone: Frozen
  file: CLAUDE.md
  anchor: "#1-core-identity"
  clause: "SPEC+EARS format"
  canary_gate: true

- id: CONST-V3R2-002
  zone: Frozen
  file: .claude/rules/moai/core/moai-constitution.md
  anchor: "#quality-gates"
  clause: "TRUST 5"
  canary_gate: true
` + "```" + `
`
	if err := os.WriteFile(registryPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "CLAUDE.md", "table")
	if err != nil {
		t.Fatalf("runConstitutionList --file error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "CONST-V3R2-001") {
		t.Errorf("CONST-V3R2-001 not found in output")
	}
	if strings.Contains(output, "CONST-V3R2-002") {
		t.Errorf("moai-constitution.md entry must not be in output")
	}
}

// TestConstitutionListJSON verifies JSON format output.
func TestConstitutionListJSON(t *testing.T) {
	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "json")
	if err != nil {
		t.Fatalf("runConstitutionList --format json error: %v", err)
	}

	output := buf.String()

	// Must be valid JSON
	var result struct {
		Entries []map[string]any `json:"entries"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON parse error: %v\noutput: %s", err, output)
	}

	if len(result.Entries) != 3 {
		t.Errorf("JSON entries count = %d, want 3", len(result.Entries))
	}
}

// TestConstitutionListRegistryMissing_FileNotFound verifies that an error is returned when the registry file is absent.
// Direct mapping to AC-CON-001-013 (file-missing subtest).
func TestConstitutionListRegistryMissing_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	nonExistentPath := filepath.Join(dir, "nonexistent-registry.md")

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, nonExistentPath, nil, "", "table")
	if err == nil {
		t.Fatal("must return an error for a non-existent registry path")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, nonExistentPath) {
		t.Errorf("error message must contain path %q: %v", nonExistentPath, err)
	}
}

// TestConstitutionListRegistryMissing_PermissionDenied verifies an error
// is returned when the registry file is not readable.
// AC-CON-001-013 (permission-denied subtest).
//
// Skipped on Windows: os.Chmod(path, 0o000) does not remove read access on
// Windows because the Windows ACL model does not honor POSIX permission bits
// in the same way. The test relies on POSIX permission semantics.
func TestConstitutionListRegistryMissing_PermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics required; skipped on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("permission-denied test skipped when running as root")
	}

	dir := t.TempDir()
	registryPath := filepath.Join(dir, "zone-registry.md")
	if err := os.WriteFile(registryPath, []byte(sampleRegistryContent), 0o600); err != nil {
		t.Fatalf("failed to create registry file: %v", err)
	}

	// Remove permissions
	if err := os.Chmod(registryPath, 0o000); err != nil {
		t.Fatalf("chmod error: %v", err)
	}
	defer func() {
		_ = os.Chmod(registryPath, 0o600)
	}()

	var buf bytes.Buffer
	err := runConstitutionList(&buf, io.Discard, dir, registryPath, nil, "", "table")
	if err == nil {
		t.Fatal("must return an error for a registry file with no permissions")
	}
}
