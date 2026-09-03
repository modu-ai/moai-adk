package template

// glm_slot_test.go — t360: the alias→slot mapping and its two per-slot readers.
//
// The mapping is the ONE alias/slot pairing in the tree (it mirrors setGLMEnv's
// ANTHROPIC_DEFAULT_*_MODEL assignments). Both readers funnel through it so the
// effort half and the model half cannot drift apart — the drift this card
// repaired keyed the model half to the high slot while the effort half was
// slot-keyed, silently discarding a stored effort.

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

func TestGLMSlotForModel(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"opus", GLMSlotHigh},
		{"sonnet", GLMSlotMedium},
		{"haiku", GLMSlotLow},
		{"fable", GLMSlotFable},
		{"opus[1m]", GLMSlotHigh},               // 1M suffix split before lookup
		{"fable[1m]", GLMSlotFable},             //
		{ModelIDOpus5, GLMSlotHigh},             // canonical id reverse-mapped
		{"claude-sonnet-5", GLMSlotMedium},      //
		{"claude-haiku-4-5", GLMSlotLow},        //
		{"", ""},                                // no model pinned
		{config.DefaultGLM53, ""},               // a raw GLM id is not an alias
		{ModelInherit, ""},                      // the inherit sentinel owns no slot
		{ModelAliasCanonicalID("opusplan"), ""}, // routing alias owns no tier slot
	}
	for _, tc := range cases {
		if got := GLMSlotForModel(tc.model); got != tc.want {
			t.Errorf("GLMSlotForModel(%q) = %q, want %q", tc.model, got, tc.want)
		}
	}
}

func TestGLMSlotEffortForModel(t *testing.T) {
	effort := config.GLMTierEffort{High: "e-high", Medium: "e-medium", Low: "e-low", Fable: "e-fable"}
	cases := []struct{ model, want string }{
		{"opus", "e-high"},
		{"sonnet", "e-medium"},
		{"haiku", "e-low"},
		{"fable", "e-fable"},
		{"", ""},                  // no slot claim → caller keeps its own fallback
		{config.DefaultGLM53, ""}, //
	}
	for _, tc := range cases {
		if got := GLMSlotEffortForModel(effort, tc.model); got != tc.want {
			t.Errorf("GLMSlotEffortForModel(%q) = %q, want %q", tc.model, got, tc.want)
		}
	}
}

// TestGLMSlotModelOrHigh pins the model half — including its high-slot fallback,
// which is what keeps a session claiming no slot byte-identical to the
// pre-repair behaviour.
func TestGLMSlotModelOrHigh(t *testing.T) {
	models := config.GLMModels{
		High:   "m-high",
		Medium: "m-medium",
		Low:    "m-low",
		Fable:  "m-fable",
	}
	cases := []struct{ model, want string }{
		{"opus", "m-high"},
		{"sonnet", "m-medium"},
		{"haiku", "m-low"},
		{"fable", "m-fable"},
		{"fable[1m]", "m-fable"},
		{"", "m-high"},                  // no slot claim → high-slot fallback
		{config.DefaultGLM53, "m-high"}, //
		{ModelInherit, "m-high"},        //
	}
	for _, tc := range cases {
		if got := GLMSlotModelOrHigh(models, tc.model); got != tc.want {
			t.Errorf("GLMSlotModelOrHigh(%q) = %q, want %q", tc.model, got, tc.want)
		}
	}
}

// TestGLMSlotHalvesAgree is the anti-drift guard, and it is the reason the two
// readers share the mapping by CALL rather than by a comment claiming they
// match: for every alias, the effort half and the model half MUST land on the
// same slot, and every unknown alias must fall off the mapping in both. A future
// edit that keys either half to a fixed slot — the shape this card repaired —
// fails here instead of silently discarding a stored effort at runtime.
//
// The per-slot sentinels are the slot names themselves, so a cross-slot read
// names the slot it wrongly reached.
func TestGLMSlotHalvesAgree(t *testing.T) {
	effort := config.GLMTierEffort{High: GLMSlotHigh, Medium: GLMSlotMedium, Low: GLMSlotLow, Fable: GLMSlotFable}
	models := config.GLMModels{High: GLMSlotHigh, Medium: GLMSlotMedium, Low: GLMSlotLow, Fable: GLMSlotFable}

	// The four aliases: both halves read the slot the mapping names.
	for alias, want := range map[string]string{
		"opus":   GLMSlotHigh,
		"sonnet": GLMSlotMedium,
		"haiku":  GLMSlotLow,
		"fable":  GLMSlotFable,
	} {
		if got := GLMSlotForModel(alias); got != want {
			t.Errorf("%s: mapping returns slot %q, want %q", alias, got, want)
		}
		if got := GLMSlotEffortForModel(effort, alias); got != want {
			t.Errorf("%s: effort half reads slot %q, want %q", alias, got, want)
		}
		if got := GLMSlotModelOrHigh(models, alias); got != want {
			t.Errorf("%s: model half reads slot %q, want %q", alias, got, want)
		}
	}

	// Unknown aliases: the mapping claims no slot, and both halves say so in
	// their own vocabulary — the effort half with the empty string (the caller
	// keeps its prefs chain), the model half with the high-slot fallback (the
	// pre-repair behaviour, preserved deliberately). A half that resolved an
	// unknown alias to some other slot would mean it stopped consulting the
	// mapping.
	for _, unknown := range []string{"", ModelInherit, config.DefaultGLM53, "gpt-9", ModelAliasCanonicalID("opusplan")} {
		if got := GLMSlotForModel(unknown); got != "" {
			t.Errorf("%q: mapping returns slot %q, want no slot", unknown, got)
		}
		if got := GLMSlotEffortForModel(effort, unknown); got != "" {
			t.Errorf("%q: effort half reads slot %q, want \"\" (no slot claim)", unknown, got)
		}
		if got := GLMSlotModelOrHigh(models, unknown); got != GLMSlotHigh {
			t.Errorf("%q: model half reads slot %q, want the %q fallback", unknown, got, GLMSlotHigh)
		}
	}
}
