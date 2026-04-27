//go:build integration
// +build integration

// Package harness_integration — T-P5-06: 스냅샷 롤백 byte-identical 복원 검증 (REQ-HL-009).
package harness_integration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/harness"
	"github.com/modu-ai/moai-adk/internal/harness/safety"
)

// TestIT06_Rollback은 스냅샷 생성 → 파일 수정 → 롤백 후
// 원본 파일이 byte-identical하게 복원되는지 검증한다.
// REQ-HL-009: rollback <date> verb에서 RestoreSnapshot()을 사용한다.
func TestIT06_Rollback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// ── 원본 SKILL.md 생성 ────────────────────────────────────────────────
	// 다양한 UTF-8 문자와 공백을 포함하여 byte-identical 복원을 엄격히 검증
	originalContent := `---
name: my-harness-rollback
description: 롤백 테스트용 스킬
triggers:
  - keyword: "rollback"
  - keyword: "restore"
---

# 롤백 테스트 스킬

이 스킬은 byte-identical 복원 검증을 위해 생성되었다.

## 특수 문자 포함

- UTF-8: 한글, emoji(🔄), 특수기호(←→)
- 줄바꿈: LF 전용
- 들여쓰기: 스페이스 2칸
`
	skillPath := setupSkillFile(t, dir, "my-harness-rollback", originalContent)

	// 원본 바이트 저장 (rollback 후 비교용)
	originalBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("원본 파일 읽기 실패: %v", err)
	}

	// ── Apply() 호출: 스냅샷 생성 + 파일 수정 ────────────────────────────
	snapshotBase := filepath.Join(dir, "snapshots")
	violationLogPath := filepath.Join(dir, "frozen-guard-violations.jsonl")
	rateLimitPath := filepath.Join(dir, "rate-limit-state.json")

	pipeline := safety.NewPipeline(safety.PipelineConfig{
		ViolationLogPath: violationLogPath,
		RateLimitPath:    rateLimitPath,
		AutoApply:        true,
	})

	proposal := harness.Proposal{
		ID:               "prop-rollback-001",
		TargetPath:       skillPath,
		FieldKey:         "description",
		NewValue:         "수정된 description (롤백 전)",
		PatternKey:       "moai_subcommand:/moai run",
		Tier:             harness.TierRule,
		ObservationCount: 5,
		CreatedAt:        time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	applier := harness.NewApplier()
	if err := applier.Apply(proposal, pipeline, snapshotBase, nil); err != nil {
		t.Fatalf("Apply 실패: %v", err)
	}

	// ── 파일이 실제로 수정되었는지 확인 ──────────────────────────────────
	modifiedBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("수정된 파일 읽기 실패: %v", err)
	}
	if bytes.Equal(originalBytes, modifiedBytes) {
		t.Fatal("Apply() 후 파일이 수정되지 않음")
	}

	// ── 스냅샷 디렉토리 탐색 ─────────────────────────────────────────────
	snapshotEntries, err := os.ReadDir(snapshotBase)
	if err != nil {
		t.Fatalf("스냅샷 베이스 디렉토리 읽기 실패: %v", err)
	}
	if len(snapshotEntries) == 0 {
		t.Fatal("스냅샷 디렉토리가 비어 있음")
	}

	// 첫 번째 (유일한) 스냅샷 디렉토리
	snapshotDir := filepath.Join(snapshotBase, snapshotEntries[0].Name())

	// ── RestoreSnapshot 호출 ──────────────────────────────────────────────
	if err := harness.RestoreSnapshot(snapshotDir); err != nil {
		t.Fatalf("RestoreSnapshot 실패: %v", err)
	}

	// ── byte-identical 복원 검증 ──────────────────────────────────────────
	restoredBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("복원된 파일 읽기 실패: %v", err)
	}

	if !bytes.Equal(originalBytes, restoredBytes) {
		t.Errorf("복원된 파일이 원본과 byte-identical하지 않음\n원본 길이: %d\n복원 길이: %d",
			len(originalBytes), len(restoredBytes))

		// 차이점 출력 (디버깅용)
		for i := range min(len(originalBytes), len(restoredBytes)) {
			if originalBytes[i] != restoredBytes[i] {
				t.Errorf("첫 번째 불일치 위치: offset=%d, original=%02x, restored=%02x",
					i, originalBytes[i], restoredBytes[i])
				break
			}
		}
	}
}

// TestIT06_RollbackInvalidSnapshot은 manifest.json이 없는 디렉토리에서
// RestoreSnapshot이 오류를 반환하는지 검증한다.
func TestIT06_RollbackInvalidSnapshot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	emptySnapshotDir := filepath.Join(dir, "empty-snapshot")
	if err := os.MkdirAll(emptySnapshotDir, 0o755); err != nil {
		t.Fatalf("빈 스냅샷 디렉토리 생성 실패: %v", err)
	}

	err := harness.RestoreSnapshot(emptySnapshotDir)
	if err == nil {
		t.Error("manifest.json 없는 디렉토리에서 오류 기대, nil 반환됨")
	}
}

// min은 두 정수 중 작은 값을 반환한다 (Go 1.21+ builtin 사용).
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
