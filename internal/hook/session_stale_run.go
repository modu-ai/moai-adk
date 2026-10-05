// session_stale_run.go — the stale-run notice (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-022 hook clause, REQ-RNC-025).
//
// A session whose LAUNCH LABEL or whose existing SESSION RECORD carries a
// legacy role value (`lead`, `worker`, `agent` — pre-rename vocabulary) must
// never be silently adopted into the new vocabulary: SessionStart emits the
// stale-run notice instead of a leader or lane notice, and the session-record
// writer leaves the record absent or byte-identical. The remedy is a relaunch,
// printed as an executable `moai factory relaunch` command line for the
// operator (SPEC-FACTORY-STALE-RUN-HEAL-001).
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
)

// legacyLaunchLabelValue returns the launch-label value carrying a legacy
// role, or "" when every label in the environment is current-vocabulary (or
// absent). The leader label (config.EnvFactoryLeadName — chain-session and factory
// leaders alike) and the lane label (config.EnvMoaiFactoryWorker) are the two
// role-bearing launch labels.
func legacyLaunchLabelValue() string {
	if label := strings.TrimSpace(os.Getenv(config.EnvFactoryLeadName)); factory.IsLegacyLeaderSpelling(label) {
		return label
	}
	if label := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker)); factory.IsLegacyFactoryRoleValue(label) {
		return label
	}
	return ""
}

// isLegacyRecordRole reports whether a session record's role value is legacy
// vocabulary — detection only, never a role mapping (REQ-RNC-009).
func isLegacyRecordRole(role string) bool {
	return role == "lead" || role == "worker" || role == "agent"
}

// staleRunNoticeFor returns the stale-run notice for this session when its
// launch label OR its existing session record carries a legacy role value,
// and "" when the session is current-vocabulary (or not a run member). The
// factory branch is run-state gated (staleRunPrescriptionGate,
// SPEC-STALE-RUN-LABEL-001); the retired-mode relaunch branch is not a retire
// prescription and stays ungated.
func staleRunNoticeFor(root, sessionID, lang string) string {
	if label := legacyLaunchLabelValue(); label != "" {
		return gatedStaleRunAnswer(root, sessionID, label, lang)
	}
	if sessionID != "" && root != "" {
		if rec, err := factory.Read(root, sessionID); err == nil && isLegacyRecordRole(rec.Role) {
			return gatedStaleRunAnswer(root, sessionID, rec.Role, lang)
		}
	}
	return ""
}

// staleRunMessages is the operator-facing prose of one locale for the
// stale-run notice (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-022, REQ-RNC-025).
//
// The same two invariants as the sibling i18n tables govern it: only the
// prose lives here — the legacy value, the run id, and the
// `moai factory relaunch` command lines are protocol tokens interpolated
// verbatim, byte-identical in every locale (SPEC-FACTORY-STALE-RUN-HEAL-001
// REQ-SRH-002) — and no field carries a leading or trailing newline. The one
// interior newline puts a command line on its own line.
//
// The "stale run:" prefix is kept verbatim in every locale: it names the
// condition the M1 tests grep for and the operator scans for.
//
// Format-argument order is pinned with explicit %[n] indices where the
// locales' natural word orders differ; the legacy value is always %[1] and
// the run id always %[2].
type staleRunMessages struct {
	roleValueRetire       string // legacy role value %[1]q, run id %[2]s, command line %[3]s — the factory branch
	roleValueRelaunch     string // legacy role value %[1]q — the retired-mode branch
	laneLabelRetire       string // legacy lane label %[1]q, run id %[2]s, command line %[3]s — the factory message hook
	laneLabelUnbind       string // orphan label %[1]q, run %[2]s, measured state %[3]s — the one-time unbind notice (REQ-SRL-005)
	laneLabelUnbindRebind string // header of the re-bind command line, no format args — one active run (REQ-SRH-001)
	laneLabelUnbindMany   string // header of the re-bind command lines, no format args — several active runs (REQ-SRH-001)
	laneLabelUnbindMore   string // candidate runs left out by the line cap, count %[1]d (DP10)
}

var staleRunLocales = map[string]staleRunMessages{
	langEnglish: {
		roleValueRetire: "stale run: this session carries legacy role value %[1]q from a binary before the leader/lane rename — " +
			"end this session; the operator relaunches it from a terminal with the command below (a command for the operator, not an instruction for the agent). " +
			"It retires run %[2]s only when its owner is dead, then joins a live run:\n%[3]s",
		roleValueRelaunch: "stale run: this session carries legacy role value %[1]q from a binary before the leader/lane rename — " +
			"end this session and relaunch it under the current vocabulary",
		laneLabelRetire: "stale run: lane label %[1]q is legacy vocabulary from a binary before the leader/lane rename — " +
			"end this session; the operator relaunches it from a terminal with the command below (a command for the operator, not an instruction for the agent). " +
			"It retires run %[2]s only when its owner is dead, then joins a live run:\n%[3]s",
		laneLabelUnbind: "stale run: lane label %[1]q is legacy vocabulary from a binary before the leader/lane rename, and the factory run %[2]s measures %[3]s (not active) — " +
			"this session is unbound from factory messaging; no further stale-run notices will be emitted for this session",
		laneLabelUnbindRebind: "an active factory run exists in this project — to rejoin it, end this session; " +
			"the operator runs the command below from a terminal (a command for the operator, not an instruction for the agent):",
		laneLabelUnbindMany: "several active factory runs exist in this project — to rejoin one, end this session; " +
			"the operator runs one of the commands below from a terminal (commands for the operator, not instructions for the agent):",
		laneLabelUnbindMore: "(+%[1]d more active factory runs — list them with 'moai factory runs')",
	},
	"ko": {
		roleValueRetire: "stale run: 이 세션은 리더/레인 개칭 이전 바이너리의 레거시 역할 값 %[1]q 을(를) 담고 있습니다 — " +
			"이 세션을 끝내세요. 아래 명령은 운영자가 터미널에서 직접 실행하는 명령이며 에이전트에게 내리는 지시가 아닙니다. " +
			"run %[2]s 은(는) 소유자가 종료된 경우에만 은퇴시키고, 이어서 살아 있는 run 에 합류합니다.\n%[3]s",
		roleValueRelaunch: "stale run: 이 세션은 리더/레인 개칭 이전 바이너리의 레거시 역할 값 %[1]q 을(를) 담고 있습니다 — " +
			"이 세션을 끝내고 현행 어휘로 다시 띄우세요.",
		laneLabelRetire: "stale run: 레인 라벨 %[1]q 은(는) 리더/레인 개칭 이전 바이너리의 레거시 어휘입니다 — " +
			"이 세션을 끝내세요. 아래 명령은 운영자가 터미널에서 직접 실행하는 명령이며 에이전트에게 내리는 지시가 아닙니다. " +
			"run %[2]s 은(는) 소유자가 종료된 경우에만 은퇴시키고, 이어서 살아 있는 run 에 합류합니다.\n%[3]s",
		laneLabelUnbind: "stale run: 레인 라벨 %[1]q 은(는) 리더/레인 개칭 이전 바이너리의 레거시 어휘이고, 팩토리 run %[2]s 의 측정 상태는 %[3]s (not active) 입니다 — " +
			"이 세션은 팩토리 메시징에서 언바운드되었으며 이 세션에 대해 스테일 런 안내를 더 내지 않습니다.",
		laneLabelUnbindRebind: "이 프로젝트에 활성 팩토리 run 이 있습니다 — 다시 합류하려면 이 세션을 끝내세요. " +
			"운영자가 터미널에서 아래 명령을 실행합니다(에이전트에게 내리는 지시가 아닙니다):",
		laneLabelUnbindMany: "이 프로젝트에 활성 팩토리 run 이 여러 개 있습니다 — 하나에 다시 합류하려면 이 세션을 끝내세요. " +
			"운영자가 터미널에서 아래 명령 중 하나를 실행합니다(에이전트에게 내리는 지시가 아닙니다):",
		laneLabelUnbindMore: "(활성 팩토리 run 이 %[1]d개 더 있습니다 — 'moai factory runs' 로 목록을 확인하세요)",
	},
	"ja": {
		roleValueRetire: "stale run: このセッションはリーダー/レーン改名前のバイナリのレガシー役割値 %[1]q を持っています — " +
			"このセッションを終了してください。下のコマンドはオペレーターがターミナルで実行するもので、エージェントへの指示ではありません。" +
			"run %[2]s はオーナーが終了している場合に限って退役させ、続けて稼働中の run に参加します。\n%[3]s",
		roleValueRelaunch: "stale run: このセッションはリーダー/レーン改名前のバイナリのレガシー役割値 %[1]q を持っています — " +
			"このセッションを終了し、現行の語彙で起動し直してください。",
		laneLabelRetire: "stale run: レーンラベル %[1]q はリーダー/レーン改名前のバイナリのレガシー語彙です — " +
			"このセッションを終了してください。下のコマンドはオペレーターがターミナルで実行するもので、エージェントへの指示ではありません。" +
			"run %[2]s はオーナーが終了している場合に限って退役させ、続けて稼働中の run に参加します。\n%[3]s",
		laneLabelUnbind: "stale run: レーンラベル %[1]q はリーダー/レーン改名前のバイナリのレガシー語彙で、ファクトリ run %[2]s の測定状態は %[3]s (not active) です — " +
			"このセッションはファクトリメッセージングからアンバインドされました。このセッションに対してこれ以上 stale-run の案内は発行されません。",
		laneLabelUnbindRebind: "このプロジェクトにはアクティブなファクトリ run があります — 再参加するには、このセッションを終了してください。" +
			"オペレーターがターミナルで次のコマンドを実行します（エージェントへの指示ではありません）:",
		laneLabelUnbindMany: "このプロジェクトにはアクティブなファクトリ run が複数あります — いずれかに再参加するには、このセッションを終了してください。" +
			"オペレーターがターミナルで次のコマンドのいずれかを実行します（エージェントへの指示ではありません）:",
		laneLabelUnbindMore: "（アクティブなファクトリ run があと%[1]d件あります — 'moai factory runs' で一覧できます）",
	},
	"zh": {
		roleValueRetire: "stale run：本会话携带主导/泳道改名前二进制文件的遗留角色值 %[1]q —— " +
			"请结束本会话。下面的命令由操作员在终端中自行运行，不是给代理的指令。" +
			"仅当 run %[2]s 的所有者已终止时才会将其退役，随后加入存活的 run。\n%[3]s",
		roleValueRelaunch: "stale run：本会话携带主导/泳道改名前二进制文件的遗留角色值 %[1]q —— " +
			"请结束本会话，并按现行词汇重新启动。",
		laneLabelRetire: "stale run：泳道标签 %[1]q 是主导/泳道改名前二进制文件的遗留词汇 —— " +
			"请结束本会话。下面的命令由操作员在终端中自行运行，不是给代理的指令。" +
			"仅当 run %[2]s 的所有者已终止时才会将其退役，随后加入存活的 run。\n%[3]s",
		laneLabelUnbind: "stale run：泳道标签 %[1]q 是主导/泳道改名前二进制文件的遗留词汇，且工厂 run %[2]s 的实测状态为 %[3]s（not active）—— " +
			"本会话已从工厂消息通道解绑，此后不再为本会话发出 stale-run 提示。",
		laneLabelUnbindRebind: "本项目存在活跃的工厂 run —— 如需重新加入，请先结束本会话，" +
			"再由操作员在终端中运行下面的命令（不是给代理的指令）：",
		laneLabelUnbindMany: "本项目存在多个活跃的工厂 run —— 如需重新加入其中之一，请先结束本会话，" +
			"再由操作员在终端中运行下面的命令之一（不是给代理的指令）：",
		laneLabelUnbindMore: "（另有 %[1]d 个活跃的工厂 run —— 可用 'moai factory runs' 列出）",
	},
}

// staleRunMessagesFor resolves a locale to its stale-run prose, falling back
// to English for anything the table does not carry — the same contract as
// factoryMessagesFor.
func staleRunMessagesFor(lang string) staleRunMessages {
	if m, ok := staleRunLocales[lang]; ok {
		return m
	}
	return staleRunLocales[langEnglish]
}

// staleRunNotice renders the stale-run message. The relaunch command is named
// only for factory sessions (the run id + MOAI_FACTORY_WORKERS discriminator):
// a retired-mode run has no factory run to relaunch into — its remedy is ending the
// session and relaunching.
func staleRunNotice(value, lang string) string {
	runID := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID))
	m := staleRunMessagesFor(lang)
	if runID != "" && os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return fmt.Sprintf(m.roleValueRetire, value, runID, legacyRelaunchLine(runID))
	}
	return fmt.Sprintf(m.roleValueRelaunch, value)
}

// legacyRelaunchLine is row R6 of the notice-line table (spec.md §D.7): the
// stale-run prescription's command for a legacy session whose run measures
// active. The provider token comes from the session's own launch backend
// (REQ-SRH-016).
func legacyRelaunchLine(runID string) string {
	return strings.Join(factory.RelaunchNoticeFor(factory.RelaunchNoticeState{
		Provider:  factory.RelaunchProviderForBackend(os.Getenv(config.EnvFactoryBackend)),
		Legacy:    true,
		Run:       runID,
		RunActive: true,
	}).Lines, "\n")
}

// legacyFactoryHookNotice is the factory message hook's stale-run answer: a
// lane whose label is legacy never registers a broker peer — it reports the
// stale-run condition instead. The SessionStart surface passes the operator's
// locale; the broker hook surface (agent-facing additionalContext) passes
// langEnglish.
func legacyFactoryHookNotice(label, runID, lang string) string {
	if !factory.IsLegacyFactoryRoleValue(strings.TrimSpace(label)) {
		return ""
	}
	if runID != "" {
		return fmt.Sprintf(staleRunMessagesFor(lang).laneLabelRetire, label, runID, legacyRelaunchLine(runID))
	}
	return staleRunNotice(label, lang)
}
