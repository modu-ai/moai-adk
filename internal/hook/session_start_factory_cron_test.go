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
