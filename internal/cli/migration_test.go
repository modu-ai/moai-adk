package cli_test

import (
	"testing"
)

// TestMigrationStatus_Human은 human-readable status 출력을 검증합니다.
// REQ-V3R2-RT-007-015: moai migration status는 현재 버전, pending, last applied를 출력합니다.
func TestMigrationStatus_Human(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationStatus_JSON은 JSON format 출력을 검증합니다.
// REQ-V3R2-RT-007-041: --json 플래그 시 machine-readable JSON을 출력합니다.
func TestMigrationStatus_JSON(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRun_AppliesPending은 pending migration 적용을 검증합니다.
// REQ-V3R2-RT-007-040: moai migration run은 pending 마이그레이션을 적용합니다.
func TestMigrationRun_AppliesPending(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRun_NoPending은 pending 없을 때 동작을 검증합니다.
func TestMigrationRun_NoPending(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_NoRollbackable은 rollback 불가능한 경우를 검증합니다.
// REQ-V3R2-RT-007-024: Rollback이 nil이면 MigrationNotRollbackable 에러를 반환합니다.
func TestMigrationRollback_NoRollbackable(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_Succeeds는 rollback 성공 경로를 검증합니다.
// REQ-V3R2-RT-007-042: Rollback이 선언된 마이그레이션은 rollback을 지원합니다.
func TestMigrationRollback_Succeeds(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}

// TestMigrationRollback_M001_Rejected는 m001 rollback 거부를 검증합니다.
// REQ-V3R2-RT-007-024: m001은 CRITICAL bug-fix이므로 rollback 불가입니다.
func TestMigrationRollback_M001_Rejected(t *testing.T) {
	// RED: migration CLI가 아직 존재하지 않음
	t.Skip("waiting for migration CLI implementation")
}
