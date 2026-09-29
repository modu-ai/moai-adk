// session_start_factory_autodispatch_test.go — SPEC-TODO-CLASSIFY-DISPATCH-001
// AC-TCD-012's payload snapshot: the bootstrap payload a default -f lane
// launch carries is the self-dispatch rule; the payload an opted-out launch
// carries is a DIFFERENT manual-mode rule. Absence reads as the default
// (fail-open, REQ-TCD-011 — the default is recorded in code).
package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestLaneBootstrapCarriesDispatchMode(t *testing.T) {
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, "claude")

	t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchAuto)
	autoPayload := factoryLaneRuleForSource("startup", "en")
	if !strings.Contains(autoPayload, "factory next") {
		t.Errorf("auto-dispatch payload carries no self-dispatch instruction:\n%s", autoPayload)
	}

	t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchManual)
	manualPayload := factoryLaneRuleForSource("startup", "en")
	if strings.Contains(manualPayload, "run `moai factory next`") {
		t.Errorf("manual payload still instructs self-dispatch:\n%s", manualPayload)
	}
	if !strings.Contains(manualPayload, "manual") {
		t.Errorf("manual payload does not name the manual mode:\n%s", manualPayload)
	}
	if autoPayload == manualPayload {
		t.Errorf("the two launches carry the SAME bootstrap payload; the opt-out must be visible in the payload")
	}

	// Fail-open: absence reads as the default (auto), never as manual.
	t.Setenv(config.EnvFactoryAutoDispatch, "")
	if factoryLaneRuleForSource("startup", "en") != autoPayload {
		t.Errorf("absent selection did not read as the auto default")
	}

	// Every conversation locale carries the manual rule too — a locale the
	// table misses must not fall back INTO a dispatch instruction.
	t.Setenv(config.EnvFactoryAutoDispatch, config.FactoryDispatchManual)
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		payload := factoryLaneRuleForSource("startup", lang)
		if payload == "" {
			t.Errorf("locale %s: manual payload is empty", lang)
			continue
		}
		if strings.Contains(payload, "run `moai factory next`") {
			t.Errorf("locale %s: manual payload still instructs self-dispatch:\n%s", lang, payload)
		}
	}
}
