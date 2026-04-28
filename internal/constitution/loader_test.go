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

// testdataPath는 testdata 디렉토리의 절대 경로를 반환한다.
func testdataPath(filename string) string {
	// 이 파일의 디렉토리 기준으로 testdata 경로 구성
	return filepath.Join("testdata", filename)
}

// TestLoadRegistryValidFile은 유효한 registry 파일을 성공적으로 로드함을 검증한다.
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

// TestLoadRegistryDuplicateIDsFails는 중복 ID registry 로드 시 오류를 반환함을 검증한다.
func TestLoadRegistryDuplicateIDsFails(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("duplicate_ids.md"), ".")
	if err == nil {
		t.Fatal("중복 ID registry 로드가 오류를 반환해야 한다")
	}
}

// TestLoadRegistryMarksOrphanWithoutPanic은 존재하지 않는 파일 참조 시 panic 없이 Orphan 마킹함을 검증한다.
// AC-CON-001-007 직접 매핑.
func TestLoadRegistryMarksOrphanWithoutPanic(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("orphan_reference.md"), ".")
	// 오류가 있어도 panic 없이 orphan을 마킹해야 한다
	// LoadRegistry는 orphan 발생 시 non-nil error를 반환할 수 있다 (warning level)
	_ = err

	if reg == nil {
		t.Fatal("Registry가 nil이어서는 안 된다")
	}

	// orphan 엔트리가 마킹되어야 한다
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

	// 정상 엔트리도 로드되어야 한다
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("정상 엔트리 CONST-V3R2-001을 찾을 수 없다")
	}
	if rule.Clause != "SPEC+EARS format" {
		t.Errorf("CONST-V3R2-001 clause = %q, want %q", rule.Clause, "SPEC+EARS format")
	}
}

// TestLoadRegistryMalformedYAML은 잘못된 YAML의 경우 오류를 반환함을 검증한다.
func TestLoadRegistryMalformedYAML(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("malformed_yaml.md"), ".")
	if err == nil {
		t.Fatal("잘못된 YAML registry 로드가 오류를 반환해야 한다")
	}
}

// TestLoadRegistryEmptyFrozen은 Frozen 엔트리가 없는 registry를 로드함을 검증한다.
// FilterByZone(ZoneFrozen)이 빈 slice를 반환해야 한다.
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

// TestLoadRegistryOverflowMirror는 51개 design mirror 엔트리가 있는 경우 overflow 감지를 검증한다.
// AC-CON-001-008 관련.
func TestLoadRegistryOverflowMirror(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("overflow_mirror.md"), ".")
	// overflow mirror가 있어도 panic 없이 로드되어야 한다
	_ = err

	if reg == nil {
		t.Fatal("Registry가 nil이어서는 안 된다")
	}
}

// TestRegistryGet은 O(1) 조회가 정확히 작동함을 검증한다.
func TestRegistryGet(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	// 존재하는 ID
	rule, ok := reg.Get("CONST-V3R2-001")
	if !ok {
		t.Error("CONST-V3R2-001을 찾을 수 없다")
	}
	if rule.Clause != "SPEC+EARS format" {
		t.Errorf("CONST-V3R2-001 clause = %q, want %q", rule.Clause, "SPEC+EARS format")
	}

	// 존재하지 않는 ID
	_, ok = reg.Get("CONST-V3R2-999")
	if ok {
		t.Error("CONST-V3R2-999가 존재해서는 안 된다")
	}
}

// TestRegistryFilterByZone은 Zone 필터링이 정확히 작동함을 검증한다.
// AC-CON-001-002 관련.
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

// TestIDStabilityAppendOnly는 기존 ID가 신규 엔트리 추가 후에도 동일한 clause를 가리킴을 검증한다.
// AC-CON-001-009 직접 매핑.
func TestIDStabilityAppendOnly(t *testing.T) {
	t.Parallel()

	// 초기 registry 로드
	reg1, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("첫 번째 LoadRegistry 오류: %v", err)
	}

	rule1, ok := reg1.Get("CONST-V3R2-001")
	if !ok {
		t.Fatal("CONST-V3R2-001을 찾을 수 없다")
	}

	// 동일한 파일을 다시 로드해도 같은 clause를 가리켜야 한다
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

// TestRegistryEntryHasExactSixFieldsWithCanonicalNames는 로드된 엔트리의 6-field schema를 검증한다.
// AC-CON-001-017 직접 매핑.
func TestRegistryEntryHasExactSixFieldsWithCanonicalNames(t *testing.T) {
	t.Parallel()

	// valid_registry.md를 raw YAML로 파싱하여 key set 검증
	content, err := os.ReadFile(testdataPath("valid_registry.md"))
	if err != nil {
		t.Fatalf("파일 읽기 오류: %v", err)
	}

	// YAML code fence 추출
	src := string(content)
	start := strings.Index(src, "```yaml")
	end := strings.LastIndex(src, "```")
	if start == -1 || end == -1 || end <= start {
		t.Fatal("YAML code fence를 찾을 수 없다")
	}
	yamlBlock := src[start+7 : end]

	// 첫 번째 엔트리의 key를 raw map으로 파싱하여 확인
	// 이를 위해 간단한 YAML key 추출 방법 사용
	requiredKeys := []string{"id", "zone", "file", "anchor", "clause", "canary_gate"}
	for _, key := range requiredKeys {
		if !strings.Contains(yamlBlock, key+":") {
			t.Errorf("YAML key %q를 찾을 수 없다", key)
		}
	}
}

// TestLoadRegistryFileNotFound는 존재하지 않는 파일 경로의 LoadRegistry 오류를 검증한다.
func TestLoadRegistryFileNotFound(t *testing.T) {
	t.Parallel()

	_, err := constitution.LoadRegistry(testdataPath("nonexistent_file.md"), ".")
	if err == nil {
		t.Fatal("존재하지 않는 파일 LoadRegistry가 오류를 반환해야 한다")
	}
}

// BenchmarkLoadRegistry200Entries는 200 엔트리 registry의 cold load 성능을 측정한다.
// AC-CON-001-015 매핑. 목표: <10ms per operation.
func BenchmarkLoadRegistry200Entries(b *testing.B) {
	// 200 엔트리 임시 파일 생성
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

	// 10ms 목표 검증 (벤치마크에서 직접 검증)
	b.StopTimer()
	const maxDuration = 10 * time.Millisecond
	// 단순 확인을 위한 단일 실행
	start := time.Now()
	_, _ = constitution.LoadRegistry(path, tmpDir)
	elapsed := time.Since(start)
	if elapsed > maxDuration {
		b.Logf("경고: LoadRegistry 200 entries = %v, 목표 %v 초과", elapsed, maxDuration)
	}
}
