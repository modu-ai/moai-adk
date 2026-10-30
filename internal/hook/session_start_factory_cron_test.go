package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// The standing recheck belongs to the SESSION, not to a card; the leader carries its own watch cron; and
// both state the judgment ladder that replaces ask-and-wait. Tool names and the ladder tokens stay verbatim
// in every locale.

var ladderTokens = []string{"jev_ask", "codex_task", "claude_audit"}

// sessionBound is a phrase per locale that says the cron is armed once for the session.
var sessionBound = map[string]string{
	"en": "ONE recurring `CronCreate` recheck for this session",
	"ko": "이 세션에 `CronCreate` 반복 재점검 크론 하나",
	"ja": "このセッションに `CronCreate` で繰り返しの再点検用 cron を1つ",
	"zh": "为本会话设置一个 `CronCreate` 循环复查定时任务",
}

func TestLaneRecheckIsBoundToTheSessionNotTheCard(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
		rule := factoryLaneRuleForSource("startup", lang)
		if !strings.Contains(rule, sessionBound[lang]) {
			t.Errorf("locale %s: the recheck is not bound to the session (want %q)", lang, sessionBound[lang])
		}
		for _, perCard := range []string{"before a card's first stage", "카드의 첫 단계를 시작하기 전에", "カードの最初の段階を始める前に", "在开始卡片的第一个阶段之前"} {
			if strings.Contains(rule, perCard) {
				t.Errorf("locale %s: the rule still arms the cron per card (%q)", lang, perCard)
			}
		}
		for _, perCard := range []string{"`CronDelete` it when you send the completion report", "완료 보고를 보낼 때 `CronDelete`", "完了報告を送るときに `CronDelete`", "发送完成报告时用 `CronDelete`"} {
			if strings.Contains(rule, perCard) {
				t.Errorf("locale %s: the rule still deletes the cron at the card's completion report (%q)", lang, perCard)
			}
		}
		for _, tok := range ladderTokens {
			if !strings.Contains(rule, tok) {
				t.Errorf("locale %s: the lane rule does not carry the judgment-ladder token %s", lang, tok)
			}
		}
	}
}

func TestLeaderNoticeCarriesTheWatchCronAndTheLadder(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		sdScrubRuleEnv(t)
		n := factoryLeaderNotice("run-1", 2, lang)
		for _, tok := range append([]string{"CronCreate", "CronList", "moai-factory-foreman", laneArmExpression()}, ladderTokens...) {
			if !strings.Contains(n, tok) {
				t.Errorf("locale %s: the leader notice does not carry %q", lang, tok)
			}
		}
	}
}

// An idle exit per locale: with no card and no backlog the cron records once, deletes itself and stops, so
// it can never keep spending turns on nothing. The advisory model is named only "if configured and
// available". Neither text may promise a CronDelete at a moment the session cannot act (its end).
var idleExit = map[string]string{
	"en": "`idle: backlog empty <time>`",
	"ko": "`idle: backlog empty <시각>`",
	"ja": "`idle: backlog empty <時刻>`",
	"zh": "`idle: backlog empty <时间>`",
}

var advisoryCondition = map[string]string{
	"en": "only if configured and available",
	"ko": "설정돼 있고 쓸 수 있을 때만",
	"ja": "設定済みで利用できる場合に限り",
	"zh": "仅在已配置且可用时",
}

var endOfSessionDelete = []string{"only when the session ends", "only when the run ends", "세션이 끝날 때만", "실행이 끝날 때만", "セッションが終わるときだけ", "実行が終わるときだけ", "只在会话结束时", "只在本次运行结束时"}

func TestTheCronsStopWhenThereIsNothingLeftAndTheLadderNamesTheAdvisoryConditionally(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
		lane := factoryLaneRuleForSource("startup", lang)
		leader := factoryLeaderNotice("run-1", 2, lang)
		for name, text := range map[string]string{"lane": lane, "leader": leader} {
			if !strings.Contains(text, idleExit[lang]) || !strings.Contains(text, "CronDelete") {
				t.Errorf("%s %s: no idle exit (want %q and CronDelete)", lang, name, idleExit[lang])
			}
			if !strings.Contains(text, advisoryCondition[lang]) {
				t.Errorf("%s %s: the advisory model is not conditional (want %q)", lang, name, advisoryCondition[lang])
			}
			for _, bad := range endOfSessionDelete {
				if strings.Contains(text, bad) {
					t.Errorf("%s %s: still promises a delete at session end (%q)", lang, name, bad)
				}
			}
		}
	}
}

// Which session-start sources carry the lane rule (leader decision): startup, clear and compact, plus an
// unset source. After a compact only a conversation summary remains, so the rule text is gone and must be
// injected again — the same three sources the factory role-injection rules already use. Resume does not
// carry it: the transcript still holds the earlier injection verbatim.
func TestTheLaneRuleIsInjectedOnStartupClearAndCompactButNotResume(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{{"startup", true}, {"clear", true}, {"compact", true}, {"", true}, {"resume", false}} {
		sdScrubRuleEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
		got := factoryLaneRuleForSource(tc.source, "en") != ""
		if got != tc.want {
			t.Errorf("source %q: rule present = %v, want %v", tc.source, got, tc.want)
		}
	}
}
