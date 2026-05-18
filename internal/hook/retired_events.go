// Resolution: KEEP — shared canonical list of retired event names (SPEC-V3R2-RT-006 + SPEC-V3R2-MIG-002).
// Package hook provides the authoritative retired-event name list for both the audit
// test suite and the migration step (internal/migrate.CleanupUserSettings).
package hook

// RetiredEventNames is the canonical set of Claude Code hook event names that were
// removed from the active settings.json registry (RETIRE-OBS-ONLY resolution per
// SPEC-V3R2-RT-006). Go handlers for these events remain in the binary as observability
// taps, enabling re-activation via system.yaml hook.observability_events opt-in.
//
// This exported symbol is consumed by:
//   - internal/hook/audit_test.go (retiredEventNames alias points here)
//   - internal/migrate.CleanupUserSettings (removes stale entries from user settings.json)
//
// @MX:ANCHOR: [AUTO] RetiredEventNames is the single source of truth for retired hook events
// @MX:REASON: fan_in=3, consumed by audit_test + migrate.CleanupUserSettings + future observability tooling
var RetiredEventNames = []string{
	"Notification",
	"Elicitation",
	"ElicitationResult",
	"TaskCreated",
}
