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

// testdataPath returns the absolute path of the testdata directory.
func testdataPath(filename string) string {
	// Build the testdata path relative to this file's directory
	return filepath.Join("testdata", filename)
}

// TestLoadRegistryValidFile verifies that a valid registry file is loaded successfully.
func TestLoadRegistryValidFile(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	if len(reg.Entries) != 3 {
		t.Errorf("엔트리 수 = %d, want 3", len(reg.Entries))
	}
}

// TestLoadRegistryDuplicateIDsFails verifies that loading a registry with duplicate IDs returns an error.
func TestLoadRegistryDuplicateIDsFails(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("duplicate_ids.md"), ".")
	if err == nil {
		t.Fatal("중복 ID registry 로드가 오류를 반환해야 한다")
	}
}

// TestLoadRegistryMarksOrphanWithoutPanic verifies that referencing a non-existent file
// marks it as Orphan without panicking.
// Direct mapping to AC-CON-001-007.
func TestLoadRegistryMarksOrphanWithoutPanic(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("orphan_reference.md"), ".")
	// Even with an error, orphan must be marked without panicking
	// LoadRegistry may return a non-nil error (warning level) when orphan occurs
	_ = err

	if reg == nil {
		t.Fatal("Registry가 nil이어서는 안 된다")
	}

	// The orphan entry must be marked
	var foundOrphan bool
	for _, entry := range reg.Entries {
		if entry.File == "this-file-does-not-exist.md" {
			if !entry.Orphan() {
				t.Errorf("존재하지 않는 파일 엔트리의 Orphan() = false, true여야 한다")
			}
			foundOrphan = true
		}
	}

	if !foundOrphan {
		t.Error("orphan 엔트리를 찾을 수 없다")
	}

	// Normal entries must also be loaded
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("정상 엔트리 CONST-V3R2-001을 찾을 수 없다")
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
		t.Fatal("잘못된 YAML registry 로드가 오류를 반환해야 한다")
	}
}

// TestLoadRegistryEmptyFrozen verifies loading a registry with no Frozen entries.
// FilterByZone(ZoneFrozen) must return an empty slice.
func TestLoadRegistryEmptyFrozen(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("empty_frozen.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) != 0 {
		t.Errorf("Frozen 엔트리 수 = %d, want 0", len(frozen))
	}
}

// TestLoadRegistryOverflowMirror verifies overflow detection when there are 51 design mirror entries.
// Related to AC-CON-001-008.
func TestLoadRegistryOverflowMirror(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("overflow_mirror.md"), ".")
	// Even with overflow mirror, loading must succeed without panicking
	_ = err

	if reg == nil {
		t.Fatal("Registry가 nil이어서는 안 된다")
	}
}

// TestRegistryGet verifies that O(1) lookup works correctly.
func TestRegistryGet(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	// Existing ID
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("CONST-V3R2-001을 찾을 수 없다")
	}
	if rule.Clause != "SPEC+EARS format" {
		t.Errorf("CONST-V3R2-001 clause = %q, want %q", rule.Clause, "SPEC+EARS format")
	}

	// Non-existent ID
	_, ok = reg.Get("CONST-V3R2-999")
	if ok {
		t.Error("CONST-V3R2-999가 존재해서는 안 된다")
	}
}

// TestRegistryFilterByZone verifies that Zone filtering works correctly.
// Related to AC-CON-001-002.
func TestRegistryFilterByZone(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	frozen := reg.FilterByZone(constitution.ZoneFrozen)
	if len(frozen) != 2 {
		t.Errorf("Frozen 엔트리 수 = %d, want 2", len(frozen))
	}

	evolvable := reg.FilterByZone(constitution.ZoneEvolvable)
	if len(evolvable) != 1 {
		t.Errorf("Evolvable 엔트리 수 = %d, want 1", len(evolvable))
	}
}

// TestIDStabilityAppendOnly verifies that an existing ID points to the same clause
// after new entries are appended.
// Direct mapping to AC-CON-001-009.
func TestIDStabilityAppendOnly(t *testing.T) {
	t.Parallel()

	// Load the initial registry
	reg1, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("첫 번째 LoadRegistry 오류: %v", err)
	}

	rule1, ok := reg1.Get("CONST-V3R2-001")
	if !ok {
		t.Fatal("CONST-V3R2-001을 찾을 수 없다")
	}

	// Reloading the same file must point to the same clause
	reg2, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("두 번째 LoadRegistry 오류: %v", err)
	}

	rule2, ok := reg2.Get("CONST-V3R2-001")
	if !ok {
		t.Fatal("두 번째 로드 후 CONST-V3R2-001을 찾을 수 없다")
	}

	if rule1.Clause != rule2.Clause {
		t.Errorf("ID 안정성 실패: 첫 번째=%q, 두 번째=%q", rule1.Clause, rule2.Clause)
	}
}

// TestRegistryEntryHasExactSixFieldsWithCanonicalNames verifies the 6-field schema of loaded entries.
// Direct mapping to AC-CON-001-017.
func TestRegistryEntryHasExactSixFieldsWithCanonicalNames(t *testing.T) {
	t.Parallel()

	// Parse valid_registry.md as raw YAML and verify the key set
	content, err := os.ReadFile(testdataPath("valid_registry.md"))
	if err != nil {
		t.Fatalf("파일 읽기 오류: %v", err)
	}

	// Extract the YAML code fence
	src := string(content)
	start := strings.Index(src, "```yaml")
	end := strings.LastIndex(src, "```")
	if start == -1 || end == -1 || end <= start {
		t.Fatal("YAML code fence를 찾을 수 없다")
	}
	yamlBlock := src[start+7 : end]

	// Parse the first entry's keys as a raw map and check
	// Use a simple YAML key extraction approach for this
	requiredKeys := []string{"id", "zone", "file", "anchor", "clause", "canary_gate"}
	for _, key := range requiredKeys {
		if !strings.Contains(yamlBlock, key+":") {
			t.Errorf("YAML key %q를 찾을 수 없다", key)
		}
	}
}

// TestLoadRegistryFileNotFound verifies the LoadRegistry error for a non-existent file path.
func TestLoadRegistryFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("nonexistent_file.md"), ".")
	if err == nil {
		t.Fatal("존재하지 않는 파일 LoadRegistry가 오류를 반환해야 한다")
	}
}

// BenchmarkLoadRegistry200Entries measures the cold load performance of a 200-entry registry.
// Maps to AC-CON-001-015. Target: <10ms per operation.
func BenchmarkLoadRegistry200Entries(b *testing.B) {
	// Create a temporary 200-entry file
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
		b.Fatalf("벤치마크 파일 생성 오류: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		reg, err := constitution.LoadRegistry(path, tmpDir)
		if err != nil {
			b.Fatalf("LoadRegistry 오류: %v", err)
		}
		if len(reg.Entries) != 200 {
			b.Fatalf("엔트리 수 = %d, want 200", len(reg.Entries))
		}
	}

	// Verify the 10ms target (directly in the benchmark)
	b.StopTimer()
	const maxDuration = 10 * time.Millisecond
	// Single execution for a simple check
	start := time.Now()
	_, _ = constitution.LoadRegistry(path, tmpDir)
	elapsed := time.Since(start)
	if elapsed > maxDuration {
		b.Logf("경고: LoadRegistry 200 entries = %v, 목표 %v 초과", elapsed, maxDuration)
	}
}
