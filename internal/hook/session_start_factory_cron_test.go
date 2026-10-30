package hook

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// The lane wake is self-paced (the user's choice of 2026-10-30, recorded by the leader and read back from the
// leader session's AskUserQuestion answer): each wake ends by scheduling at most ONE next wake at a delay
// relative to now, and none when no work remains. No repeating job is registered and no absolute-clock
// one-shot is used. The tool, command and ladder tokens stay verbatim in every locale.

var ladderTokens = []string{"jev_ask", "codex_task", "claude_audit"}

// selfPaced are the verbatim tokens of the new wake contract.
var selfPaced = []string{"ScheduleWakeup", "/loop"}

// noRepeatingJob is, per locale, the sentence that forbids a repeating job and an absolute one-shot.
var noRepeatingJob = map[string]string{
	"en": "Never register a fixed-interval repeating job (no recurring `CronCreate`) and never an absolute-clock one-shot.",
	"ko": "고정 간격 반복 작업(반복 `CronCreate`)도, 절대 시각 one-shot도 등록하지 않습니다.",
	"ja": "固定間隔の繰り返しジョブ（繰り返しの `CronCreate`）も、絶対時刻の one-shot も登録しません。",
	"zh": "不要注册固定间隔的重复任务（不用循环的 `CronCreate`），也不要注册绝对时刻的 one-shot。",
}

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

// repeatingJobInstructions are phrases that would TEACH arming a repeating job, in any locale.
var repeatingJobInstructions = []string{
	"CronList", "CronDelete", "(cron `", "cron `*/", "arm ONE recurring", "recurring `CronCreate` recheck",
	"반복 재점검 크론 하나를 겁니다", "繰り返しの再点検用 cron を1つ設定", "设置一个 `CronCreate`",
}

func laneRuleFor(t *testing.T, lang string) string {
	t.Helper()
	sdScrubRuleEnv(t)
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
	t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
	return factoryLaneRuleForSource("startup", lang)
}

func TestLaneWake_SelfPacedWithTime_NoFixedIntervalJob(t *testing.T) {
	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		lane := laneRuleFor(t, lang)
		leader := factoryLeaderNotice("run-1", 2, lang)
		for name, text := range map[string]string{"lane": lane, "leader": leader} {
			for _, tok := range selfPaced {
				if !strings.Contains(text, tok) {
					t.Errorf("%s %s: missing the self-paced token %q", lang, name, tok)
				}
			}
			if !strings.Contains(text, noRepeatingJob[lang]) {
				t.Errorf("%s %s: no prohibition of a repeating job / absolute one-shot (want %q)", lang, name, noRepeatingJob[lang])
			}
			if !strings.Contains(text, idleExit[lang]) {
				t.Errorf("%s %s: no idle exit (want %q)", lang, name, idleExit[lang])
			}
			if !strings.Contains(text, advisoryCondition[lang]) {
				t.Errorf("%s %s: the advisory model is not conditional (want %q)", lang, name, advisoryCondition[lang])
			}
			for _, tok := range ladderTokens {
				if !strings.Contains(text, tok) {
					t.Errorf("%s %s: missing the judgment-ladder token %s", lang, name, tok)
				}
			}
			for _, bad := range repeatingJobInstructions {
				if strings.Contains(text, bad) {
					t.Errorf("%s %s: still teaches a repeating job (%q)", lang, name, bad)
				}
			}
			if !strings.Contains(text, laneWakeBound()) {
				t.Errorf("%s %s: no wake bound %q (the lease renewal period)", lang, name, laneWakeBound())
			}
		}
		if !strings.Contains(lane, "/moai goal") {
			t.Errorf("%s lane: no /moai goal continuation", lang)
		}
	}
}

// Which session-start sources carry the lane rule (leader decision): startup, clear and compact, plus an
// unset source. After a compact only a conversation summary remains, so the rule text is gone and must be
// injected again - the same three sources the factory role-injection rules already use. Resume does not
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
