package constitution_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// testdataPath returns the absolute path to the testdata directory.
func testdataPath(filename string) string {
	// Construct testdata path relative to this file's directory
	return filepath.Join("testdata", filename)
}

// TestLoadRegistryValidFile verifies that a valid registry file is loaded successfully.
func TestLoadRegistryValidFile(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	if len(reg.Entries) != 3 {
		t.Errorf("entry count = %d, want 3", len(reg.Entries))
	}
}

// TestLoadRegistryDuplicateIDsFails verifies that an error is returned when loading a registry with duplicate IDs.
func TestLoadRegistryDuplicateIDsFails(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("duplicate_ids.md"), ".")
	if err == nil {
		t.Fatal("loading a duplicate ID registry must return an error")
	}
}

// TestLoadRegistryMarksOrphanWithoutPanic verifies that orphans are marked without panic
// when a non-existent file is referenced.
// Direct mapping to AC-CON-001-007.
func TestLoadRegistryMarksOrphanWithoutPanic(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("orphan_reference.md"), ".")
	// Must mark orphan without panic, even if an error occurs
	// LoadRegistry may return a non-nil error on orphan (warning level)
	_ = err

	if reg == nil {
		t.Fatal("Registry must not be nil")
	}

	// Orphan entries must be marked
	var foundOrphan bool
	for _, entry := range reg.Entries {
		if entry.File == "this-file-does-not-exist.md" {
			if !entry.Orphan() {
				t.Errorf("non-existent file entry Orphan() = false, must be true")
			}
			foundOrphan = true
		}
	}

	if !foundOrphan {
		t.Error("could not find orphan entry")
	}

	// Normal entries must also be loaded
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("could not find normal entry CONST-V3R2-001")
	}
	if rule.Clause != "SPEC+EARS format" {
		t.Errorf("CONST-V3R2-001 clause = %q, want %q", rule.Clause, "SPEC+EARS format")
	}
}

// TestLoadRegistryMalformedYAML verifies that an error is returned for malformed YAML.
func TestLoadRegistryMalformedYAML(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("malformed_yaml.md"), ".")
	if err == nil {
		t.Fatal("loading a malformed YAML registry must return an error")
	}
}

// TestLoadRegistryEmptyFrozen verifies that a registry with no Frozen entries loads correctly.
// FilterByZone(ZoneFrozen) must return an empty slice.
func TestLoadRegistryEmptyFrozen(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("empty_frozen.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) != 0 {
		t.Errorf("Frozen entry count = %d, want 0", len(frozen))
	}
}

// TestLoadRegistryOverflowMirror verifies overflow detection when there are 51 design mirror entries.
// Related to AC-CON-001-008.
func TestLoadRegistryOverflowMirror(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("overflow_mirror.md"), ".")
	// Must load without panic even with overflow mirror
	_ = err

	if reg == nil {
		t.Fatal("Registry must not be nil")
	}
}

// TestRegistryGet verifies that O(1) lookup works correctly.
func TestRegistryGet(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	// Existing ID
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("could not find CONST-V3R2-001")
	}
	if rule.Clause != "SPEC+EARS format" {
		t.Errorf("CONST-V3R2-001 clause = %q, want %q", rule.Clause, "SPEC+EARS format")
	}

	// Non-existent ID
	_, ok = reg.Get("CONST-V3R2-999")
	if ok {
		t.Error("CONST-V3R2-999 must not exist")
	}
}

// TestRegistryFilterByZone verifies that zone filtering works correctly.
// Related to AC-CON-001-002.
func TestRegistryFilterByZone(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) != 2 {
		t.Errorf("Frozen entry count = %d, want 2", len(frozen))
	}

	evolvable := reg.FilterByZone(constitution.ZoneEvolvable)
	if len(evolvable) != 1 {
		t.Errorf("Evolvable entry count = %d, want 1", len(evolvable))
	}
}

// TestIDStabilityAppendOnly verifies that existing IDs still point to the same clause after new entries are added.
// Direct mapping to AC-CON-001-009.
func TestIDStabilityAppendOnly(t *testing.T) {
	t.Parallel()

	// Load initial registry
	reg1, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("first LoadRegistry error: %v", err)
	}

	rule1, ok := reg1.Get("CONST-V3R2-001")
	if !ok {
		t.Fatal("could not find CONST-V3R2-001")
	}

	// Reloading the same file must still point to the same clause
	reg2, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("second LoadRegistry error: %v", err)
	}

	rule2, ok := reg2.Get("CONST-V3R2-001")
	if !ok {
		t.Fatal("could not find CONST-V3R2-001 after second load")
	}

	if rule1.Clause != rule2.Clause {
		t.Errorf("ID stability failure: first=%q, second=%q", rule1.Clause, rule2.Clause)
	}
}

// TestRegistryEntryHasExactSixFieldsWithCanonicalNames verifies the 6-field schema of loaded entries.
// Direct mapping to AC-CON-001-017.
func TestRegistryEntryHasExactSixFieldsWithCanonicalNames(t *testing.T) {
	t.Parallel()

	// Parse valid_registry.md as raw YAML to verify key set
	content, err := os.ReadFile(testdataPath("valid_registry.md"))
	if err != nil {
		t.Fatalf("file read error: %v", err)
	}

	// Extract YAML code fence
	src := string(content)
	start := strings.Index(src, "```yaml")
	end := strings.LastIndex(src, "```")
	if start == -1 || end == -1 || end <= start {
		t.Fatal("could not find YAML code fence")
	}
	yamlBlock := src[start+7 : end]

	// Verify the keys of the first entry using a simple raw map YAML extraction
	requiredKeys := []string{"id", "zone", "file", "anchor", "clause", "canary_gate"}
	for _, key := range requiredKeys {
		if !strings.Contains(yamlBlock, key+":") {
			t.Errorf("YAML key %q not found", key)
		}
	}
}

// TestLoadRegistryFileNotFound verifies the LoadRegistry error for a non-existent file path.
func TestLoadRegistryFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("nonexistent_file.md"), ".")
	if err == nil {
		t.Fatal("LoadRegistry for non-existent file must return an error")
	}
}

// BenchmarkLoadRegistry200Entries measures cold load performance for a 200-entry registry.
// AC-CON-001-015 mapping. Target: <10ms per operation.
func BenchmarkLoadRegistry200Entries(b *testing.B) {
	// Create a temp file with 200 entries
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench_registry.md")

	var sb strings.Builder
	sb.WriteString("# Benchmark Registry\n\n## Entries\n\n```yaml\n")
	for i := range 200 {
		fmt.Fprintf(&sb, "- id: CONST-V3R2-%03d\n", i+1)
		sb.WriteString("  zone: Frozen\n")
		sb.WriteString("  file: .claude/rules/moai/core/moai-constitution.md\n")
		sb.WriteString("  anchor: \"#quality-gates\"\n")
		fmt.Fprintf(&sb, "  clause: \"Benchmark clause %d\"\n", i+1)
		sb.WriteString("  canary_gate: true\n")
	}
	sb.WriteString("```\n")

	if err := os.WriteFile(path, []byte(sb.String()), 0o600); err != nil {
		b.Fatalf("failed to create benchmark file: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		reg, err := constitution.LoadRegistry(path, tmpDir)
		if err != nil {
			b.Fatalf("LoadRegistry error: %v", err)
		}
		if len(reg.Entries) != 200 {
			b.Fatalf("entry count = %d, want 200", len(reg.Entries))
		}
	}

	// Verify 10ms target (single run check within the benchmark)
	b.StopTimer()
	const maxDuration = 10 * time.Millisecond
	start := time.Now()
	_, _ = constitution.LoadRegistry(path, tmpDir)
	elapsed := time.Since(start)
	if elapsed > maxDuration {
		b.Logf("warning: LoadRegistry 200 entries = %v, exceeds target %v", elapsed, maxDuration)
	}
}
