package config

import "testing"

// cache.yaml was retired (SPEC-CONFIG-DEAD-SWEEP-001): LoadCacheConfig,
// CacheConfig, Validate, and DefaultCacheConfig were removed as dead code.
// The single live consumer of the closed-set of accepted session_ttl values is
// the moai web settings seam (internal/settings/schema_sections.go), which
// builds the session_ttl select options from ValidSessionTTLs(). That accessor
// survives; these tests guard it.

// TestValidSessionTTLs verifies the exported ordered accessor consumed by the
// moai web console cache section (SPEC-WEB-CONSOLE-013 REQ-WC13-013). It returns
// exactly {1h, 5m, off} in order, and the returned slice is a defensive copy
// (mutating it does not corrupt the SSOT).
func TestValidSessionTTLs(t *testing.T) {
	got := ValidSessionTTLs()
	want := []string{"1h", "5m", "off"}
	if len(got) != len(want) {
		t.Fatalf("ValidSessionTTLs() len = %d, want %d", len(got), len(want))
	}
	for i, v := range want {
		if got[i] != v {
			t.Errorf("ValidSessionTTLs()[%d] = %q, want %q (ordered closed set)", i, got[i], v)
		}
	}
	// Defensive copy: mutating the returned slice must not affect a later call.
	got[0] = "MUTATED"
	if again := ValidSessionTTLs(); again[0] != "1h" {
		t.Errorf("ValidSessionTTLs() returned a shared slice — mutation leaked (got[0]=%q)", again[0])
	}
}
