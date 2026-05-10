package migration_test

import (
	"testing"
)

// TestRegistry_DuplicateVersion_Panics는 중복 Version 선언 시 panic을 검증합니다.
// REQ-V3R2-RT-007-053: 동일 Version을 가진 두 마이그레이션이 등록되면 panic이 발생해야 합니다.
func TestRegistry_DuplicateVersion_Panics(t *testing.T) {
	// RED: registry 패키지가 아직 존재하지 않음
	t.Skip("waiting for migration package implementation")
}

// TestRegistry_Pending는 현재 버전 기준 pending 마이그레이션 목록을 검증합니다.
// REQ-V3R2-RT-007-012: Pending(current)은 Version > current인 마이그레이션 목록을 반환합니다.
func TestRegistry_Pending(t *testing.T) {
	// RED: registry 패키지가 아직 존재하지 않음
	t.Skip("waiting for migration package implementation")
}

// TestRegistry_Highest는 registry의 최대 버전을 검증합니다.
// REQ-V3R2-RT-007-016: registry는 compile-time static이며 최대 버전을 조회할 수 있습니다.
func TestRegistry_Highest(t *testing.T) {
	// RED: registry 패키지가 아직 존재하지 않음
	t.Skip("waiting for migration package implementation")
}
