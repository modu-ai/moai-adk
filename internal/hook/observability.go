package hook

import (
	"context"
	"log/slog"
)

// observabilityOptIn checks whether a RETIRE-OBS-ONLY event should emit logs.
// Returns true only if the event is in system.yaml hook.observability_events whitelist.
// This implements REQ-V3R2-RT-006-040: retired events emit structured logs only when opted in.
//
// Pattern A from plan.md M2-T07: silent return if !observabilityOptIn().
func observabilityOptIn(ctx context.Context, eventName string) bool {
	// Get config from context
	cfg := configFromContext(ctx)
	if cfg == nil {
		// No config available → silent (no-op)
		return false
	}

	// Check if event is in observability whitelist
	for _, optInEvent := range cfg.System.Hook.ObservabilityEvents {
		if optInEvent == eventName {
			return true
		}
	}

	// Event not in whitelist → silent
	slog.Debug("retired event not opted in", "event", eventName, "observability_events", cfg.System.Hook.ObservabilityEvents)
	return false
}

// configFromContext extracts Config from context, if available.
// This is a helper to avoid direct coupling to config package in handler code.
func configFromContext(ctx context.Context) *Config {
	// TODO: Wire with config.FromContext(ctx) once RT-005 context propagation is available
	// For now, return nil to make handlers silent (defensive default)
	return nil
}
