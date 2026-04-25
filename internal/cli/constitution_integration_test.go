//go:build integration

package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestConstitutionList_RealRegistry는 실제 zone-registry.md를 사용한 통합 테스트다.
// 실제 registry 파일이 존재하는 환경에서만 실행된다.
// AC-CON-001-001, AC-CON-001-002 관련.
func TestConstitutionList_RealRegistry(t *testing.T) {
	// 실제 zone-registry.md 경로 탐색
	registryPath := os.Getenv("MOAI_CONSTITUTION_REGISTRY")
	if registryPath == "" {
		// 프로젝트 루트 기준으로 탐색
		cwd, err := os.Getwd()
		if err != nil {
			t.Skipf("cwd 확인 불가: %v", err)
		}
		// internal/cli에서 프로젝트 루트까지 상위 이동
		projectRoot := filepath.Join(cwd, "..", "..")
		registryPath = filepath.Join(projectRoot, ".claude", "rules", "moai", "core", "zone-registry.md")
	}

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Skipf("zone-registry.md not found at %q - skipping integration test", registryPath)
	}

	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(registryPath)))))

	t.Run("list all entries", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, nil, "", "table")
		if err != nil {
			t.Fatalf("runConstitutionList 오류: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "CONST-V3R2-001") {
			t.Errorf("실제 registry에 CONST-V3R2-001이 포함되어야 한다\n출력: %s", output)
		}
	})

	t.Run("filter frozen zone", func(t *testing.T) {
		frozen := constitution.ZoneFrozen
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, &frozen, "", "table")
		if err != nil {
			t.Fatalf("--zone frozen 오류: %v", err)
		}
		output := buf.String()
		if strings.Contains(output, "Evolvable") {
			t.Errorf("Frozen 필터 결과에 Evolvable이 포함되어서는 안 된다")
		}
	})

	t.Run("json format valid", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, nil, "", "json")
		if err != nil {
			t.Fatalf("--format json 오류: %v", err)
		}
		var result struct {
			Entries []map[string]any `json:"entries"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON 파싱 오류: %v\n출력: %s", err, buf.String())
		}
		if len(result.Entries) == 0 {
			t.Error("실제 registry JSON에 entries가 있어야 한다")
		}
		t.Logf("실제 registry: %d entries", len(result.Entries))
	})

	t.Run("minimum frozen entries", func(t *testing.T) {
		frozen := constitution.ZoneFrozen
		var buf bytes.Buffer
		err := runConstitutionList(&buf, io.Discard, projectDir, registryPath, &frozen, "", "json")
		if err != nil {
			t.Fatalf("--zone frozen --format json 오류: %v", err)
		}
		var result struct {
			Entries []map[string]any `json:"entries"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON 파싱 오류: %v", err)
		}
		// AC-CON-001-006: 최소 7개의 Frozen 불변 조항이 있어야 한다
		const minFrozen = 7
		if len(result.Entries) < minFrozen {
			t.Errorf("Frozen entries = %d, want >= %d (7 canonical invariants)", len(result.Entries), minFrozen)
		}
	})
}

// TestConstitutionGuard_RealRegistry는 실제 registry를 사용한 guard 통합 테스트다.
func TestConstitutionGuard_RealRegistry(t *testing.T) {
	registryPath := os.Getenv("MOAI_CONSTITUTION_REGISTRY")
	if registryPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			t.Skipf("cwd 확인 불가: %v", err)
		}
		projectRoot := filepath.Join(cwd, "..", "..")
		registryPath = filepath.Join(projectRoot, ".claude", "rules", "moai", "core", "zone-registry.md")
	}

	if _, err := os.Stat(registryPath); os.IsNotExist(err) {
		t.Skipf("zone-registry.md not found at %q - skipping integration test", registryPath)
	}

	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(registryPath)))))

	t.Run("no violations is OK", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionGuard(&buf, io.Discard, projectDir, registryPath, []string{})
		if err != nil {
			t.Errorf("위반 없음 시 nil 반환 기대: %v", err)
		}
	})

	t.Run("CONST-V3R2-001 is Frozen violation", func(t *testing.T) {
		var buf bytes.Buffer
		err := runConstitutionGuard(&buf, io.Discard, projectDir, registryPath, []string{"CONST-V3R2-001"})
		if err == nil {
			t.Error("CONST-V3R2-001 변경은 Frozen zone 위반이므로 에러를 반환해야 한다")
		}
	})
}
