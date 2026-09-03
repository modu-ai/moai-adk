package cli

// glm_slot_model_wire_test.go — t360 regression guard: the reasoning-effort
// collapse is MODEL-keyed, so it must read the model of the SAME slot the
// effort was read from.
//
// The effort is slot-keyed (glmSlotEffortForModel: fable → effort.fable) but the
// collapse was keyed to the high slot unconditionally. With every slot on
// glm-5.3-flash the two agreed by coincidence; splitting the fable slot onto
// glm-5.3 broke the coincidence, and a stored effort.fable=low was silently
// discarded — flash pins every effort to max, and the high slot is flash.

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/template"
)

// reasoningEffortFromEnv extracts ANTHROPIC_REASONING_EFFORT from an env slice.
func reasoningEffortFromEnv(env []string) string {
	prefix := config.EnvAnthropicReasoningEffort + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

// TestBuildEnvForGLMLaunch_KeysCollapseToTheSessionSlot pins the wire value to
// the slot the session model actually occupies — the defect this card repairs.
func TestBuildEnvForGLMLaunch_KeysCollapseToTheSessionSlot(t *testing.T) {
	// The mismatch: high slot on flash, fable slot on glm-5.3.
	models := config.GLMModels{
		High:   config.DefaultGLM53Flash,
		Medium: config.DefaultGLM53Flash,
		Low:    config.DefaultGLM53Flash,
		Fable:  config.DefaultGLM53,
	}
	tier := config.GLMTierEffort{Fable: template.GLMStateLow}
	base := []string{"PATH=/usr/bin"}

	t.Run("fable session on a non-flash slot keeps its stored low", func(t *testing.T) {
		// The launch composition, verbatim from launchClaude.
		effort := resolveGLMMainSessionEffort("fable", tier, "high")
		if effort != template.GLMStateLow {
			t.Fatalf("stored effort did not reach the launch: %q, want %q", effort, template.GLMStateLow)
		}
		got := reasoningEffortFromEnv(buildEnvForGLMLaunch(models, "fable", effort, base))
		if got != template.GLMReasoningEffortLow {
			t.Errorf("%s = %q, want %q — the fable slot runs %s (not flash), so a stored low must survive; keying the collapse to the high slot (%s) discards it",
				config.EnvAnthropicReasoningEffort, got, template.GLMReasoningEffortLow, models.Fable, models.High)
		}
	})

	t.Run("opus session on the flash high slot still pins to max", func(t *testing.T) {
		got := reasoningEffortFromEnv(buildEnvForGLMLaunch(models, "opus", template.GLMStateLow, base))
		if got != template.GLMReasoningEffortMax {
			t.Errorf("%s = %q, want %q — the high slot is flash, which accepts max only",
				config.EnvAnthropicReasoningEffort, got, template.GLMReasoningEffortMax)
		}
	})

	t.Run("a model claiming no slot falls back to the high slot", func(t *testing.T) {
		// Byte-identical to the pre-repair behaviour for an unresolved session
		// model: the high slot (flash) governs, so the wire is max.
		for _, m := range []string{"", config.DefaultGLM53, template.ModelInherit} {
			got := reasoningEffortFromEnv(buildEnvForGLMLaunch(models, m, template.GLMStateLow, base))
			if got != template.GLMReasoningEffortMax {
				t.Errorf("session model %q: %s = %q, want %q (high-slot fallback)",
					m, config.EnvAnthropicReasoningEffort, got, template.GLMReasoningEffortMax)
			}
		}
	})
}
