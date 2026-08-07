package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// SPEC-AUTONOMY-TIERS-001 M8 — moai web console autonomy-tier toggle (AC-002).
// The toggle LOGIC (config.TierToggleOptions) is done in M5; this surface wires
// the HTTP handler + console UI fragment that surfaces the 3-tier toggle in the
// browser, calling TierToggleOptions for gating. fully-autonomous MUST be
// rendered disabled when no sandbox proof is present AND under the kill-switch.

func serveAutonomyTiers(t *testing.T, h http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/autonomy/tiers", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandleAutonomyTiers_ThreeTiersPresent(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	a := newTestApp(t)
	rec := serveAutonomyTiers(t, a.routes())

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	for _, tier := range []string{
		config.AutonomyTierSemiAuto,
		config.AutonomyTierAutomatic,
		config.AutonomyTierFullyAutonomous,
	} {
		if !strings.Contains(string(body), tier) {
			t.Errorf("toggle body must name tier %q; body: %s", tier, body)
		}
	}
}

func TestHandleAutonomyTiers_FullyAutonomousDisabledWithoutProof(t *testing.T) {
	// AC-002: no sandbox proof -> fully-autonomous disabled in the toggle.
	t.Setenv("MOAI_SANDBOX_PROOF", "")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	a := newTestApp(t)
	rec := serveAutonomyTiers(t, a.routes())

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	// The fully-autonomous control MUST carry a disabled marker; the lower
	// two tiers MUST NOT. We assert the HTML emits a `disabled` attribute on
	// the fully-autonomous control and not on the others.
	if !hasDisabledOn(body, config.AutonomyTierFullyAutonomous) {
		t.Errorf("fully-autonomous control should be disabled without proof; body: %s", body)
	}
	if hasDisabledOn(body, config.AutonomyTierSemiAuto) {
		t.Errorf("semi-auto control should NOT be disabled; body: %s", body)
	}
	if hasDisabledOn(body, config.AutonomyTierAutomatic) {
		t.Errorf("automatic control should NOT be disabled; body: %s", body)
	}
}

func TestHandleAutonomyTiers_KillSwitchDisablesFullyAutonomousEvenWithProof(t *testing.T) {
	// AC-005 trumps AC-002: kill-switch disables fully-autonomous even WITH proof.
	t.Setenv("MOAI_SANDBOX_PROOF", "docker")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "true")
	a := newTestApp(t)
	rec := serveAutonomyTiers(t, a.routes())

	body := rec.Body.String()
	if !hasDisabledOn(body, config.AutonomyTierFullyAutonomous) {
		t.Errorf("kill-switch should disable fully-autonomous even with proof; body: %s", body)
	}
}

func TestHandleAutonomyTiers_FullyAutonomousEnabledWithProofAndNoKillSwitch(t *testing.T) {
	t.Setenv("MOAI_SANDBOX_PROOF", "docker")
	t.Setenv("MOAI_DISABLE_BYPASS_PERMISSIONS_MODE", "")
	a := newTestApp(t)
	rec := serveAutonomyTiers(t, a.routes())

	body := rec.Body.String()
	if hasDisabledOn(body, config.AutonomyTierFullyAutonomous) {
		t.Errorf("fully-autonomous should be enabled with proof + no kill-switch; body: %s", body)
	}
}

// hasDisabledOn reports whether the tier's control in the rendered HTML carries
// a `disabled` attribute. The handler renders one control per tier; we locate
// the control whose value attribute is the tier and check for `disabled`.
func hasDisabledOn(body, tier string) bool {
	// Each control is emitted on its own line containing value="<tier>".
	for _, line := range strings.Split(body, "\n") {
		if !strings.Contains(line, `value="`+tier+`"`) {
			continue
		}
		return strings.Contains(line, "disabled")
	}
	return false
}
