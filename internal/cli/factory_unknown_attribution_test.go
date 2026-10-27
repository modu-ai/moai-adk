// factory_unknown_attribution_test.go — GH #1924 (card t1988): the
// unknown-attribution refusal must state how the operator mark is actually
// obtained. No code path stamps MOAI_FACTORY_ROLE=operator — the only
// Setenv writer stamps the lane role (factory.go) — so the refusal may not
// claim the launcher stamps it; it names the manual env prefix instead.
package cli

import (
	"strings"
	"testing"
)

func TestFactoryUnknownAttributionErrorStatesManualStamp(t *testing.T) {
	err := factoryUnknownAttributionError("factory decide")
	if err == nil {
		t.Fatal("factoryUnknownAttributionError returned nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "MOAI_FACTORY_ROLE=operator") {
		t.Errorf("refusal must name the exact manual prefix, got: %s", msg)
	}
	if strings.Contains(msg, "the launcher or the operator environment stamps it") {
		t.Errorf("refusal still claims the launcher stamps the operator mark (no writer exists — GH #1924): %s", msg)
	}
	if !strings.Contains(msg, "manual") {
		t.Errorf("refusal must state stamping is manual, got: %s", msg)
	}
}
