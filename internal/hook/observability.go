// Resolution: KEEP — observability gate helper for RETIRE-OBS-ONLY handlers.
// Implements SPEC-V3R2-RT-006 REQ-040: observability opt-in via system.yaml hook.observability_events.
package hook

import (
	"strings"
)

// observabilityOptIn reports whether the named retired event is enabled
// as an observability tap for the current configuration.
//
// It reads hook.observability_events from SystemHookConfig via the ConfigProvider.
// Returns false when:
//   - cfg is nil
//   - observability_events list is empty (default)
//   - event name is not in the list
//
// Pattern A (SPEC-V3R2-RT-006 §5.2): callers MUST silently return HookOutput{}
// when observabilityOptIn returns false. NEVER emit SystemMessage or Continue:false.
//
// @MX:ANCHOR: [AUTO] observabilityOptIn guards all RETIRE-OBS-ONLY handler entry paths
// @MX:REASON: fan_in=4, called by notification/elicitation/elicitationResult/taskCreated handlers
func observabilityOptIn(cfg ConfigProvider, eventName string) bool {
	if cfg == nil {
		return false
	}
	underlying := cfg.Get()
	if underlying == nil {
		return false
	}

	events := underlying.System.Hook.ObservabilityEvents
	if len(events) == 0 {
		return false
	}

	needle := strings.ToLower(eventName)
	for _, e := range events {
		if strings.ToLower(e) == needle {
			return true
		}
	}
	return false
}
