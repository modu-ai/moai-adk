package config

// cache.yaml was retired (SPEC-CONFIG-DEAD-SWEEP-001): the cache_control
// injector, doctor metric, and PostToolUse telemetry it once fed were all
// already removed as unreachable — prompt caching is performed by Claude Code
// itself, and the live cache-hit signal is the statusline's segment
// (internal/statusline.renderCacheHit), which reads context_window.current_usage
// straight from Claude Code. With the loader gone, cache.yaml round-trips
// through nothing.
//
// The single live consumer of the closed-set of accepted session_ttl values
// is the moai web settings seam (internal/settings/schema_sections.go), which
// builds the session_ttl select options from ValidSessionTTLs(). That accessor
// and its backing slice survive here; everything else — LoadCacheConfig,
// CacheConfig, Validate, DefaultCacheConfig, cacheFileWrapper — was dead and is
// removed.

// sessionTTLValues is the ordered closed set of accepted session_ttl enum
// values — the single source for ValidSessionTTLs, so the web settings select
// cannot drift from what the cache layer once accepted.
var sessionTTLValues = []string{"1h", "5m", "off"}

// ValidSessionTTLs returns the ordered closed set of accepted cacheStrategy
// session_ttl values ("1h" | "5m" | "off"). The moai web console cache section
// consumes this as the single source of the session_ttl select options
// (SPEC-WEB-CONSOLE-013 REQ-WC13-013) — no divergent re-declaration is allowed.
func ValidSessionTTLs() []string {
	return append([]string(nil), sessionTTLValues...)
}
