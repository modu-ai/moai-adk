// Package harness — 5-Layer 안전 아키텍처 보존 아키텍처 테스트 (T-C5).
//
// 이 파일은 SPEC-V3R4-HARNESS-002 Wave C의 architectural assertion 테스트다.
// 두 테스트는 코드가 아닌 '아키텍처 불변 조건'을 검증한다:
//  1. constitution.md §5 Safety Architecture에 정확히 5개 레이어 이름이 존재.
//  2. frozenPrefixes 슬라이스가 정확히 4개의 canonical 항목을 포함.
//
// REQ-HRN-FND-006: FROZEN 경로 보호 목록 불변성 보장.
// REQ-HRN-OBS-002: 5-Layer Safety Architecture 불변성 보장.
package harness

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// findProjectRoot은 테스트 파일의 위치에서 프로젝트 루트를 탐지한다.
// Go 테스트는 패키지 디렉터리에서 실행되므로, 루트는 N단계 위에 있다.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	// 테스트 파일의 런타임 경로로 프로젝트 루트 탐지
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller 실패")
	}
	// internal/harness/safety_preservation_test.go → ../../ = 프로젝트 루트
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("filepath.Abs 실패: %v", err)
	}
	return abs
}

// TestSafetyArchitecture_LayerCount는 constitution.md §5가 정확히 5개의
// canonical Safety Architecture 레이어 이름을 포함하는지 검증한다.
//
// 검증 대상 레이어 (Layer N: Name 패턴):
//   - Layer 1: Frozen Guard
//   - Layer 2: Canary Check
//   - Layer 3: Contradiction Detector
//   - Layer 4: Rate Limiter
//   - Layer 5: Human Oversight
//
// 이 테스트는 문자열 검색으로 constitution.md가 수정되지 않았음을 검증한다.
func TestSafetyArchitecture_LayerCount(t *testing.T) {
	t.Parallel()

	projectRoot := findProjectRoot(t)
	constitutionPath := filepath.Join(projectRoot, ".claude", "rules", "moai", "design", "constitution.md")

	data, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("constitution.md 읽기 실패 (%s): %v", constitutionPath, err)
	}
	body := string(data)

	// canonical 레이어 이름 목록 (constitution.md §5 Safety Architecture 기준)
	wantLayers := []string{
		"Frozen Guard",
		"Canary Check",
		"Contradiction Detector",
		"Rate Limiter",
		"Human Oversight",
	}

	for _, layerName := range wantLayers {
		if !strings.Contains(body, layerName) {
			t.Errorf("constitution.md §5에 레이어 이름 %q 없음 — 5-Layer 구조 훼손 가능성", layerName)
		}
	}

	// 총 레이어 수: "### Layer N:" 패턴 카운트
	layerCount := strings.Count(body, "### Layer ")
	if layerCount != 5 {
		t.Errorf("constitution.md §5 '### Layer ' 수: got=%d, want=5", layerCount)
	}
}

// TestSafetyArchitecture_FrozenZoneUnchanged는 frozenPrefixes 슬라이스가
// 정확히 4개의 canonical 항목을 포함하는지 검증한다.
// REQ-HRN-FND-006: FROZEN 경로 보호 목록 불변성.
//
// 주의: tasks.md T-C5는 `.moai/project/brand/`를 4번째 항목으로 명시하나,
// 현재 frozen_guard.go 구현은 `.claude/skills/moai/`를 포함한다.
// 이 테스트는 실제 코드 상태를 검증하며, 구현과 tasks.md 간 차이는
// SPEC-V3R4-HARNESS-002 구현 노트 (discrepancy)로 기록한다.
// 미래 변경 시 이 테스트도 함께 업데이트해야 한다.
func TestSafetyArchitecture_FrozenZoneUnchanged(t *testing.T) {
	t.Parallel()

	// 현재 frozen_guard.go 코드의 실제 항목 (순서 포함)
	wantPrefixes := []string{
		".claude/agents/moai/",
		".claude/skills/moai-",
		".claude/skills/moai/",
		".claude/rules/moai/",
	}

	// 항목 수 검증
	if len(frozenPrefixes) != len(wantPrefixes) {
		t.Errorf("frozenPrefixes 항목 수: got=%d, want=%d", len(frozenPrefixes), len(wantPrefixes))
		t.Logf("실제 항목: %v", frozenPrefixes)
		return
	}

	// 각 항목 순서 및 값 검증
	for i, want := range wantPrefixes {
		if frozenPrefixes[i] != want {
			t.Errorf("frozenPrefixes[%d]: got=%q, want=%q", i, frozenPrefixes[i], want)
		}
	}

	// FROZEN 경로가 실제로 차단되는지 통합 검증
	// constitution.md 경로는 .claude/rules/moai/ 접두사로 차단되어야 함
	constitutionPath := ".claude/rules/moai/design/constitution.md"
	_, err := IsAllowedPath(constitutionPath)
	if err == nil {
		t.Errorf("IsAllowedPath(%q)는 FrozenViolationError를 반환해야 함", constitutionPath)
	} else {
		var frozenErr *FrozenViolationError
		if !isFrozenViolationError(err, &frozenErr) {
			t.Errorf("IsAllowedPath(%q) 에러 타입: got=%T, want=*FrozenViolationError", constitutionPath, err)
		}
	}
}

// isFrozenViolationError는 에러가 *FrozenViolationError 타입인지 확인한다.
func isFrozenViolationError(err error, out **FrozenViolationError) bool {
	if fve, ok := err.(*FrozenViolationError); ok {
		if out != nil {
			*out = fve
		}
		return true
	}
	return false
}
