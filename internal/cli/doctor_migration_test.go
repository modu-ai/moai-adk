package cli_test

import (
	"testing"
)

// TestDoctorMigration_PrintsCurrentVersion은 현재 버전 출력을 검증합니다.
// REQ-V3R2-RT-007-015: moai doctor --check migration는 현재 버전을 표시합니다.
func TestDoctorMigration_PrintsCurrentVersion(t *testing.T) {
	// RED: doctor migration extension이 아직 존재하지 않음
	t.Skip("waiting for doctor migration implementation")
}

// TestDoctorMigration_PrintsPendingCount는 pending 마이그레이션 수 출력을 검증합니다.
// REQ-V3R2-RT-007-015: moai doctor --check migration는 pending 목록을 표시합니다.
func TestDoctorMigration_PrintsPendingCount(t *testing.T) {
	// RED: doctor migration extension이 아직 존재하지 않음
	t.Skip("waiting for doctor migration implementation")
}
