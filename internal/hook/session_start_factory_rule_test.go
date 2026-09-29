package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// session_start_factory_rule_test.go covers the factory lane SessionStart
// rule (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-019, AC-SD-019): the rule a
// lane session receives on source startup (under every clear policy) or
// clear, chosen by the backend variable and rendered in the session's
// conversation language.

// sdScrubRuleEnv neutralizes every factory key the rule gate reads, so a
// lane variable inherited from the outer session cannot leak into a case
// (§B: the fixture environment is the test's own).
func sdScrubRuleEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		config.EnvMoaiFactoryWorker, config.EnvMoaiFactoryWorkers,
		config.EnvMoaiKanbanBackend, config.EnvMoaiKanbanCard,
		config.EnvFactoryClearPolicy,
	} {
		t.Setenv(k, "")
	}
}

// sdRuleMCPTools are the six REQ-SD-014 MCP tools the Claude-lane rule must
// name; sdRuleCLITools are their CLI equivalents.
var sdRuleMCPTools = [...]string{
	"todo_add", "todo_list", "factory_next", "factory_stage", "factory_complete", "factory_decide",
}

var sdRuleCLITools = [...]string{
	"moai todo add", "moai todo", "moai factory next",
	"moai factory stage", "moai factory complete", "moai factory decide",
}

// sdRuleAuthPin pins the operator-authorization statement per locale — the
// phrase each locale's rule carries for "lane queue promotion through
// `moai factory next` is operator-authorized for the self-dispatch lane
// mode".
var sdRuleAuthPin = map[string]string{
	"en": "operator-authorized",
	"ko": "운영자가 승인",
	"ja": "オペレーターが承認",
	"zh": "已获运营者授权",
}

// AC-SD-019 — the rule matrix: backend claude under every clear policy on
// source startup in each of en, ko, ja, zh carries the next-card rule
// naming the six MCP tools with their CLI equivalents and the
// operator-authorization statement; source clear carries it; resume and
// compact do not; a gpt lane receives the owned-card rule naming its card
// and the two CLI verbs with none of the MCP tool names; leader and
// keyless environments receive no rule for either source.
func TestSD_AC019_NextCardRuleInjection(t *testing.T) {
	t.Run("claude lane: every policy, source startup, four locales", func(t *testing.T) {
		for _, lang := range []string{"en", "ko", "ja", "zh"} {
			for _, policy := range []string{
				config.FactoryClearPolicyEach, config.FactoryClearPolicyWhenFull, config.FactoryClearPolicyRelaunch,
			} {
				sdScrubRuleEnv(t)
				t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
				t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
				t.Setenv(config.EnvFactoryClearPolicy, policy)
				rule := factoryLaneRuleForSource("startup", lang)
				if rule == "" {
					t.Fatalf("locale %s policy %s: no rule injected on startup", lang, policy)
				}
				for _, tool := range sdRuleMCPTools {
					if !strings.Contains(rule, tool) {
						t.Errorf("locale %s policy %s: rule does not name MCP tool %s", lang, policy, tool)
					}
				}
				for _, cliTool := range sdRuleCLITools {
					if !strings.Contains(rule, cliTool) {
						t.Errorf("locale %s policy %s: rule does not name CLI equivalent %q", lang, policy, cliTool)
					}
				}
				if !strings.Contains(rule, sdRuleAuthPin[lang]) {
					t.Errorf("locale %s policy %s: rule does not state the operator-authorized queue promotion (want pin %q)", lang, policy, sdRuleAuthPin[lang])
				}
			}
		}
	})

	t.Run("glm lane: source startup carries the rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-3")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGLM)
		rule := factoryLaneRuleForSource("startup", "en")
		if rule == "" {
			t.Fatal("backend glm: no rule injected on startup")
		}
		for _, tool := range sdRuleMCPTools {
			if !strings.Contains(rule, tool) {
				t.Errorf("backend glm: rule does not name MCP tool %s", tool)
			}
		}
	})

	t.Run("source clear carries the rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
		if rule := factoryLaneRuleForSource("clear", "en"); rule == "" {
			t.Fatal("source clear: no rule injected")
		}
	})

	t.Run("resume and compact carry no rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
		for _, source := range []string{"resume", "compact"} {
			if rule := factoryLaneRuleForSource(source, "en"); rule != "" {
				t.Errorf("source %s: rule injected, want none", source)
			}
		}
	})

	t.Run("gpt lane: owned-card rule names the card and the two CLI verbs only", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
		t.Setenv(config.EnvMoaiKanbanCard, "t9")
		rule := factoryLaneRuleForSource("startup", "en")
		if rule == "" {
			t.Fatal("backend gpt: no owned-card rule injected on startup")
		}
		for _, want := range []string{"t9", "moai factory stage", "moai factory complete"} {
			if !strings.Contains(rule, want) {
				t.Errorf("backend gpt: rule does not name %q", want)
			}
		}
		// The owned-card rule never names `moai factory next` or any MCP
		// tool of REQ-SD-014 — an underscore tool name of any kind is the
		// discriminator (REQ-SD-019).
		for _, forbidden := range []string{
			"factory next", "factory_next", "todo_add", "todo_list",
			"factory_stage", "factory_complete", "factory_decide",
		} {
			if strings.Contains(rule, forbidden) {
				t.Errorf("backend gpt: owned-card rule carries the forbidden token %q", forbidden)
			}
		}
	})

	t.Run("gpt lane without a card id: no rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
		if rule := factoryLaneRuleForSource("startup", "en"); rule != "" {
			t.Errorf("backend gpt without a card id: rule injected, want none (got %q)", rule)
		}
	})

	t.Run("leader environment: no rule for either source", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorkers, "3")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
		for _, source := range []string{"startup", "clear"} {
			if rule := factoryLaneRuleForSource(source, "en"); rule != "" {
				t.Errorf("leader env source %s: rule injected, want none", source)
			}
		}
	})

	t.Run("no factory keys: no rule for either source", func(t *testing.T) {
		sdScrubRuleEnv(t)
		for _, source := range []string{"startup", "clear"} {
			if rule := factoryLaneRuleForSource(source, "en"); rule != "" {
				t.Errorf("keyless env source %s: rule injected, want none", source)
			}
		}
	})

	t.Run("legacy lane label: no rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "worker-2")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
		for _, source := range []string{"startup", "clear"} {
			if rule := factoryLaneRuleForSource(source, "en"); rule != "" {
				t.Errorf("legacy label source %s: rule injected, want none", source)
			}
		}
	})

	t.Run("unknown backend: no rule", func(t *testing.T) {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvMoaiKanbanBackend, "shell")
		if rule := factoryLaneRuleForSource("startup", "en"); rule != "" {
			t.Errorf("unknown backend: rule injected, want none (got %q)", rule)
		}
	})
}
