package hook_test

import (
	"testing"
)

// TestSessionStart_InvokesMigrationRunner verifies that the session-start hook calls the runner.
// REQ-V3R2-RT-007-020: the session-start hook calls MigrationRunner.Apply.
func TestSessionStart_InvokesMigrationRunner(t *testing.T) {
	// RED: the session_start handler does not call MigrationRunner
	t.Skip("waiting for session-start migration integration")
}

// TestSessionStart_MigrationFailure_SurfacesViaSystemMessage verifies that failures are propagated via SystemMessage.
// REQ-V3R2-RT-007-021: migration failures propagate via SystemMessage but do not block the session.
func TestSessionStart_MigrationFailure_SurfacesViaSystemMessage(t *testing.T) {
	// RED: SystemMessage is not yet implemented (RT-001 merge required)
	t.Skip("waiting for RT-001 HookResponse SystemMessage merge")
}

// TestSessionStart_DisabledViaSystemYaml verifies the disabled handling via system.yaml.
// REQ-V3R2-RT-007-032: when migrations.disabled: true, the runner is skipped.
func TestSessionStart_DisabledViaSystemYaml(t *testing.T) {
	// RED: the config.Migrations field does not yet exist
	t.Skip("waiting for system.yaml migrations.disabled implementation")
}

// TestSessionStart_EnabledByDefault verifies that migration is enabled by default.
// REQ-V3R2-RT-007-032: when migrations.disabled is missing or false, the runner is called.
func TestSessionStart_EnabledByDefault(t *testing.T) {
	// RED: the config.Migrations field does not yet exist
	t.Skip("waiting for system.yaml migrations.disabled implementation")
}
