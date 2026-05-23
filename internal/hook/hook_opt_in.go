// Resolution: KEEP — HOI master toggle gate for 3 observability hook series.
// Implements SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-002 / REQ-HOI-003.
package hook

// hookOptInEnabled reports whether the 3 observability-only hook series
// (TaskCreated, Notification, handle-harness-observe-*) are enabled via the
// system.yaml hook.opt_in.enabled master toggle.
//
// Returns false when:
//   - cfg is nil
//   - cfg.Get() returns nil
//   - hook.opt_in sub-block is absent (Go zero-value default, R3 mitigation)
//   - hook.opt_in.enabled is explicitly false
//
// INDEPENDENT of observabilityOptIn() (observability.go) which reads
// hook.observability_events (SPEC-V3R2-RT-006 REQ-040 per-event RETIRE-OBS-ONLY
// whitelist). Different semantics — do NOT unify without a fresh SPEC.
// See SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 §A.3 cohabitation contract.
//
// @MX:ANCHOR: [AUTO] hookOptInEnabled is the master gate for the 3 HOI hook series
// @MX:REASON: fan_in=4, called by TaskCreated/Notification dispatch + 3 harness-observe wrappers
func hookOptInEnabled(cfg ConfigProvider) bool {
	if cfg == nil {
		return false
	}
	underlying := cfg.Get()
	if underlying == nil {
		return false
	}
	return underlying.System.Hook.OptIn.Enabled
}
