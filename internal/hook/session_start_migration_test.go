package hook_test

import (
	"testing"
)

// TestSessionStart_InvokesMigrationRunner는 session-start hook이 runner를 호출함을 검증합니다.
// REQ-V3R2-RT-007-020: session-start hook은 MigrationRunner.Apply를 호출합니다.
func TestSessionStart_InvokesMigrationRunner(t *testing.T) {
	// RED: session_start handler가 MigrationRunner를 호출하지 않음
	t.Skip("waiting for session-start migration integration")
}

// TestSessionStart_MigrationFailure_SurfacesViaSystemMessage는 실패 시 SystemMessage 전파를 검증합니다.
// REQ-V3R2-RT-007-021: migration 실패는 SystemMessage로 전파되지만 세션을 차단하지 않습니다.
func TestSessionStart_MigrationFailure_SurfacesViaSystemMessage(t *testing.T) {
	// RED: SystemMessage가 아직 구현되지 않음 (RT-001 머지 필요)
	t.Skip("waiting for RT-001 HookResponse SystemMessage merge")
}

// TestSessionStart_DisabledViaSystemYaml는 system.yaml disabled 처리를 검증합니다.
// REQ-V3R2-RT-007-032: migrations.disabled: true 시 runner를 skip합니다.
func TestSessionStart_DisabledViaSystemYaml(t *testing.T) {
	// RED: config.Migrations 필드가 아직 존재하지 않음
	t.Skip("waiting for system.yaml migrations.disabled implementation")
}

// TestSessionStart_EnabledByDefault는 기본적으로 migration이 활성화됨을 검증합니다.
// REQ-V3R2-RT-007-032: migrations.disabled가 없거나 false이면 runner를 호출합니다.
func TestSessionStart_EnabledByDefault(t *testing.T) {
	// RED: config.Migrations 필드가 아직 존재하지 않음
	t.Skip("waiting for system.yaml migrations.disabled implementation")
}

