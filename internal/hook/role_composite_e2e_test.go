package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestCompositeBudgetRealAssembly measures the assembled composite through
// the REAL producer/assembly path (attribution + GLM reminder +
// factoryBootstrapNoticeForSource + factoryLaneRuleForSource + bound line +
// roleRuleInjectionFor -> assembleInjectionComposite -> roleRuleSizeGate)
// for leader/lane roles x ko/en locales, against the 10,000 delivery cap
// (operator decisions 6c5e / 1670 / leader ruling on the instrument).
// The variable advisory pad is the lane-ko measured value 1,470 UTF-16
// (binary-latency + guard banners + chain banner + degraded notice; source:
// lane-15 live capture, delta1 audit 4,797 fixed - 3,327 decisive), applied
// uniformly per the Opus fit document's conservative assumption.
func TestCompositeBudgetRealAssembly(t *testing.T) {
	clearFactoryEnv(t)
	root := t.TempDir()
	treeRoot := "../.."
	for _, rel := range roleRuleFiles {
		data, err := os.ReadFile(filepath.Join(treeRoot, filepath.FromSlash(rel.Rel)))
		if err != nil {
			t.Fatalf("read tree rule %s: %v", rel.Rel, err)
		}
		writeRoleRuleFixture(t, root, rel, string(data))
	}
	const varPad = 1470 // UTF-16 units; see doc comment

	cases := []struct {
		name    string
		worker  bool
		leader  bool
		lang    string
		padWant int
	}{
		{"lane_ko_glm", true, false, "ko", varPad},
		{"lane_en_glm", true, false, "en", varPad},
		{"leader_ko", false, true, "ko", varPad},
		{"leader_en", false, true, "en", varPad},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.worker {
				t.Setenv(config.EnvMoaiFactoryWorker, "lane-15")
			}
			if tc.leader {
				t.Setenv(config.EnvMoaiFactoryWorkers, "1")
			}
			t.Setenv(config.EnvFactoryBackend, factoryBackendGLMForTest)

			attribution := "moai session attribution: source_session_id=measure-e2e\n"
			glm := glmGuardrailReminderText
			bootstrap := factoryBootstrapNoticeForSource("startup", root, "measure-e2e", tc.lang)
			laneRule := factoryLaneRuleForSource("startup", tc.lang)

			pad := ""
			if tc.padWant > 0 {
				pad = strings.Repeat("p", tc.padWant)
			}
			existing := strings.Join([]string{attribution, glm, bootstrap, laneRule, pad}, "\n\n")

			inj := roleRuleInjectionFor(root, "startup", tc.lang)
			if inj.Context == "" {
				t.Fatal("role injection produced an empty core")
			}
			composite := assembleInjectionComposite(existing, inj)
			composite, _ = roleRuleSizeGate(composite, inj.Context, root, "factory-lane", tc.lang)

			n := utf16Len(composite)
			t.Logf("RESULT %s: composite utf16=%d (cap 10000, margin=%d)", tc.name, n, 10000-n)
			if n > 10000 {
				t.Errorf("%s: composite %d exceeds the 10000 cap", tc.name, n)
			}
		})
	}
}

// factoryBackendGLMForTest avoids importing the factory package for one
// constant; the value is the documented backend token (session_start_factory
// switch: factory.BackendGLM).
const factoryBackendGLMForTest = "glm"
