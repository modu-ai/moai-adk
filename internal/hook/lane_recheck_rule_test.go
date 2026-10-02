package hook

// lane_recheck_rule_test.go — card t1451 regression: a lane that stops (an API
// 429 ends its turn, or it waits on a delegate's report that never comes) must
// be woken by something it armed BEFORE it stopped, because a turn that has
// ended cannot arm anything. These tests EXECUTE the rule builder and assert
// its OUTPUT: the defect was a lane bootstrap that never mentioned a
// scheduler, so the guard is over the rendered text the lane reads.

import (
	"strings"
	"testing"
	"unicode"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// recheckRuleTokens are locale-independent: the tool names, the watchdog skill
// the recheck prompt must invoke, and the off-minute recurring cadence stay
// verbatim in every locale the way the MCP tool names do.
var recheckRuleTokens = [...]string{"CronCreate", "CronList", "CronDelete", "moai-lane-watchdog", "7,27,47 * * * *"}

// recheckLocaleScript is the script the appended recheck paragraph must be
// written in per locale; a locale whose paragraph is still the English sentence
// passes the token sweep alone, so the script count is its own check.
var recheckLocaleScript = map[string]*unicode.RangeTable{
	"ko": unicode.Hangul,
	"ja": unicode.Hiragana,
	"zh": unicode.Han,
}

// recheckParagraph returns the paragraph the lane rule appends after the
// next-card or manual-dispatch rule.
func recheckParagraph(rule string) string {
	return rule[strings.LastIndex(rule, "\n\n")+2:]
}

func scriptRunes(s string, table *unicode.RangeTable) int {
	n := 0
	for _, r := range s {
		if unicode.Is(table, r) {
			n++
		}
	}
	return n
}

func TestLaneRuleCarriesStandingRecheckCron(t *testing.T) {
	for _, backend := range []string{kanban.BackendClaude, kanban.BackendGLM} {
		for _, dispatch := range []string{config.FactoryDispatchAuto, config.FactoryDispatchManual} {
			for _, source := range []string{"startup", "clear"} {
				for _, lang := range []string{"en", "ko", "ja", "zh"} {
					t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
					t.Setenv(config.EnvMoaiKanbanBackend, backend)
					t.Setenv(config.EnvFactoryAutoDispatch, dispatch)
					rule := factoryLaneRuleForSource(source, lang)
					if rule == "" {
						t.Fatalf("backend %s dispatch %s source %s locale %s: no lane rule", backend, dispatch, source, lang)
					}
					for _, token := range recheckRuleTokens {
						if !strings.Contains(rule, token) {
							t.Errorf("backend %s dispatch %s source %s locale %s: lane rule lacks the recheck token %q:\n%s", backend, dispatch, source, lang, token, rule)
						}
					}
					if table, ok := recheckLocaleScript[lang]; ok {
						if n := scriptRunes(recheckParagraph(rule), table); n < 20 {
							t.Errorf("locale %s: the recheck paragraph carries only %d runes of its own script, want at least 20 — it reads as the English text:\n%s", lang, n, recheckParagraph(rule))
						}
					}
				}
			}
		}
	}
}

// A Codex lane has no session cron tool, so its owned-card rule must not teach
// one — an instruction naming a tool the harness lacks would be a dead letter.
func TestLaneRuleOmitsCronOnCodexLane(t *testing.T) {
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGPT)
	t.Setenv(config.EnvMoaiKanbanCard, "t9")
	rule := factoryLaneRuleForSource("startup", "en")
	if rule == "" {
		t.Fatal("backend gpt: no owned-card rule injected on startup")
	}
	if strings.Contains(rule, "CronCreate") {
		t.Errorf("owned-card rule teaches a cron tool the Codex harness lacks:\n%s", rule)
	}
}
