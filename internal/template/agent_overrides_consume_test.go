package template

// agent_overrides_consume_test.go — REQ-AFR-016 (SPEC-WEB-AGENTFM-RESTORE-001
// v0.3.0 M7, card t1421): the consumption resolver behind the
// llm.agent_overrides_consume opt-in gate.
//
// The four scenarios the acceptance AC-AFR-015 pins, including the
// discriminating one: with the gate ON, an agent that has a profile-matrix
// cell but NO override entry resolves to plain session inheritance — the
// profile cell is console-default machinery and never reaches the spawn path
// (an earlier draft that fed the resolver through ResolveAgentModelEffort's
// cell fallback was rejected on exactly this assertion).
import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// consumeFixture builds an LLMConfig with the gate state and one override pin.
func consumeFixture(on bool, pins map[string]config.ModelEffort) config.LLMConfig {
	cfg := config.NewDefaultLLMConfig()
	cfg.AgentOverridesConsume = on
	for agent, me := range pins {
		if cfg.AgentOverrides == nil {
			cfg.AgentOverrides = map[string]config.ModelEffort{}
		}
		cfg.AgentOverrides[agent] = me
	}
	return cfg
}

func TestAgentOverridesConsumption(t *testing.T) {
	t.Parallel()

	t.Run("override wins when the gate is on", func(t *testing.T) {
		t.Parallel()
		got := ResolveAgentOverrideConsumption(consumeFixture(true, map[string]config.ModelEffort{
			"manager-develop": {Model: "opus", Effort: "xhigh"},
		}), "manager-develop")
		if !got.ConsumeEnabled {
			t.Error("the gate is on — ConsumeEnabled must be true")
		}
		if !got.Pinned || got.Model != "opus" {
			t.Errorf("the override must win: Pinned=%v Model=%q, want Pinned=true Model=opus", got.Pinned, got.Model)
		}
		if got.Effort != "xhigh" {
			t.Errorf("the pinned effort must be carried for the doctrine-level delivery (REQ-AFR-017), got %q", got.Effort)
		}
		if got.InheritNoOp {
			t.Error("a concrete pin is not an inherit no-op")
		}
	})

	t.Run("absent entry resolves plain inheritance even for a cell-mapped agent", func(t *testing.T) {
		t.Parallel()
		// manager-develop HAS a profile-matrix cell (the medium column maps it
		// to opus/medium) — the consumption contract must still resolve plain
		// inheritance: absent entry means the session model/effort, never the
		// cell (REQ-AFR-016, REQ-AFR-002).
		got := ResolveAgentOverrideConsumption(consumeFixture(true, nil), "manager-develop")
		if !got.ConsumeEnabled {
			t.Error("the gate is on — ConsumeEnabled must be true")
		}
		if got.Pinned || got.Model != "" {
			t.Errorf("an absent entry must resolve no pin: Pinned=%v Model=%q", got.Pinned, got.Model)
		}
		if got.Effort != "" {
			t.Errorf("an absent entry must carry no effort, got %q", got.Effort)
		}
	})

	t.Run("inherit pin is an explicit inheritance no-op", func(t *testing.T) {
		t.Parallel()
		got := ResolveAgentOverrideConsumption(consumeFixture(true, map[string]config.ModelEffort{
			"manager-develop": {Model: "inherit"},
		}), "manager-develop")
		if !got.ConsumeEnabled {
			t.Error("the gate is on — ConsumeEnabled must be true")
		}
		if !got.InheritNoOp {
			t.Error("model: inherit is the explicit inheritance no-op")
		}
		if got.Pinned || got.Model != "" {
			t.Errorf("the no-op pins nothing: Pinned=%v Model=%q", got.Pinned, got.Model)
		}
	})

	t.Run("gate off stays storage-only even with pins stored", func(t *testing.T) {
		t.Parallel()
		got := ResolveAgentOverrideConsumption(consumeFixture(false, map[string]config.ModelEffort{
			"manager-develop": {Model: "opus", Effort: "xhigh"},
		}), "manager-develop")
		if got.ConsumeEnabled {
			t.Error("the gate is off — ConsumeEnabled must be false")
		}
		if got.Pinned || got.Model != "" || got.Effort != "" || got.InheritNoOp {
			t.Errorf("with the gate off the resolver consumes nothing: %+v", got)
		}
	})
}
