// session_stale_run.go — the stale-run notice (SPEC-ROLE-NAMING-CODE-001
// REQ-RNC-022 hook clause, REQ-RNC-025).
//
// A session whose LAUNCH LABEL or whose existing SESSION RECORD carries a
// legacy role value (`lead`, `worker`, `agent` — pre-rename vocabulary) must
// never be silently adopted into the new vocabulary: SessionStart emits the
// stale-run notice instead of a leader or lane notice, and the session-record
// writer leaves the record absent or byte-identical. The remedy is a relaunch
// (factory runs first via `moai factory runs --retire`).
package hook

import (
	"fmt"
	"os"
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// legacyLaunchLabelValue returns the launch-label value carrying a legacy
// role, or "" when every label in the environment is current-vocabulary (or
// absent). The leader label (MOAI_KANBAN_LEAD_NAME — kanban and factory
// leaders alike) and the lane label (MOAI_FACTORY_WORKER) are the two
// role-bearing launch labels.
func legacyLaunchLabelValue() string {
	if label := strings.TrimSpace(os.Getenv(config.EnvFactoryLeadName)); kanban.IsLegacyLeaderSpelling(label) {
		return label
	}
	if label := strings.TrimSpace(os.Getenv(config.EnvMoaiFactoryWorker)); kanban.IsLegacyFactoryRoleValue(label) {
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
// SPEC-STALE-RUN-LABEL-001); the kanban relaunch branch is not a retire
// prescription and stays ungated.
func staleRunNoticeFor(root, sessionID, lang string) string {
	if label := legacyLaunchLabelValue(); label != "" {
		return gatedStaleRunAnswer(root, sessionID, label, lang)
	}
	if sessionID != "" && root != "" {
		if rec, err := kanban.Read(root, sessionID); err == nil && isLegacyRecordRole(rec.Role) {
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
// `moai factory runs --retire` step are protocol tokens interpolated
// verbatim — and no field carries a leading or trailing newline.
//
// The "stale run:" prefix is kept verbatim in every locale: it names the
// condition the M1 tests grep for and the operator scans for.
//
// Format-argument order is pinned with explicit %[n] indices where the
// locales' natural word orders differ; the legacy value is always %[1] and
// the run id always %[2].
type staleRunMessages struct {
	roleValueRetire       string // legacy role value %[1]q, run id %[2]s — the factory branch
	roleValueRelaunch     string // legacy role value %[1]q — the kanban branch
	laneLabelRetire       string // legacy lane label %[1]q, run id %[2]s — the factory message hook
	laneLabelUnbind       string // orphan label %[1]q, run %[2]s, measured state %[3]s — the one-time unbind notice (REQ-SRL-005)
	laneLabelUnbindRebind string // the re-bind entry, no format args — appended only while an active run exists (REQ-SRL-006)
}

var staleRunLocales = map[string]staleRunMessages{
	langEnglish: {
		roleValueRetire: "stale run: this session carries legacy role value %[1]q from a binary before the leader/lane rename — " +
			"end this session, retire the run with 'moai factory runs --retire %[2]s', then relaunch",
		roleValueRelaunch: "stale run: this session carries legacy role value %[1]q from a binary before the leader/lane rename — " +
			"end this session and relaunch it under the current vocabulary",
		laneLabelRetire: "stale run: lane label %[1]q is legacy vocabulary from a binary before the leader/lane rename — " +
			"end this session, retire the run with 'moai factory runs --retire %[2]s', then relaunch",
		laneLabelUnbind: "stale run: lane label %[1]q is legacy vocabulary from a binary before the leader/lane rename, and the factory run %[2]s measures %[3]s (not active) — " +
			"this session is unbound from factory messaging; no further stale-run notices will be emitted for this session",
		laneLabelUnbindRebind: "an active factory run exists in this project — to rejoin its slot set, end this session and relaunch with 'moai cc -l', which takes the next free slot",
	},
	"ko": {
		roleValueRetire: "stale run: 이 세션은 리더/레인 개칭 이전 바이너리의 레거시 역할 값 %[1]q 을(를) 담고 있습니다 — " +
			"이 세션을 끝내고 'moai factory runs --retire %[2]s' 로 run 을 은퇴시킨 뒤 다시 띄우세요.",
		roleValueRelaunch: "stale run: 이 세션은 리더/레인 개칭 이전 바이너리의 레거시 역할 값 %[1]q 을(를) 담고 있습니다 — " +
			"이 세션을 끝내고 현행 어휘로 다시 띄우세요.",
		laneLabelRetire: "stale run: 레인 라벨 %[1]q 은(는) 리더/레인 개칭 이전 바이너리의 레거시 어휘입니다 — " +
			"이 세션을 끝내고 'moai factory runs --retire %[2]s' 로 run 을 은퇴시킨 뒤 다시 띄우세요.",
		laneLabelUnbind: "stale run: 레인 라벨 %[1]q 은(는) 리더/레인 개칭 이전 바이너리의 레거시 어휘이고, 팩토리 run %[2]s 의 측정 상태는 %[3]s (not active) 입니다 — " +
			"이 세션은 팩토리 메시징에서 언바운드되었으며 이 세션에 대해 스테일 런 안내를 더 내지 않습니다.",
		laneLabelUnbindRebind: "이 프로젝트에 활성 팩토리 run 이 있습니다 — 슬롯에 다시 합류하려면 이 세션을 끝내고 'moai cc -l' 로 다시 띄우세요. 비어 있는 다음 슬롯을 받습니다.",
	},
	"ja": {
		roleValueRetire: "stale run: このセッションはリーダー/レーン改名前のバイナリのレガシー役割値 %[1]q を持っています — " +
			"このセッションを終了し、'moai factory runs --retire %[2]s' で run を退役させてから起動し直してください。",
		roleValueRelaunch: "stale run: このセッションはリーダー/レーン改名前のバイナリのレガシー役割値 %[1]q を持っています — " +
			"このセッションを終了し、現行の語彙で起動し直してください。",
		laneLabelRetire: "stale run: レーンラベル %[1]q はリーダー/レーン改名前のバイナリのレガシー語彙です — " +
			"このセッションを終了し、'moai factory runs --retire %[2]s' で run を退役させてから起動し直してください。",
		laneLabelUnbind: "stale run: レーンラベル %[1]q はリーダー/レーン改名前のバイナリのレガシー語彙で、ファクトリ run %[2]s の測定状態は %[3]s (not active) です — " +
			"このセッションはファクトリメッセージングからアンバインドされました。このセッションに対してこれ以上 stale-run の案内は発行されません。",
		laneLabelUnbindRebind: "このプロジェクトにはアクティブなファクトリ run があります — スロットに再参加するには、このセッションを終了して 'moai cc -l' で起動し直してください。空いている次のスロットが割り当てられます。",
	},
	"zh": {
		roleValueRetire: "stale run：本会话携带主导/泳道改名前二进制文件的遗留角色值 %[1]q —— " +
			"请结束本会话，用 'moai factory runs --retire %[2]s' 退役该 run 后重新启动。",
		roleValueRelaunch: "stale run：本会话携带主导/泳道改名前二进制文件的遗留角色值 %[1]q —— " +
			"请结束本会话，并按现行词汇重新启动。",
		laneLabelRetire: "stale run：泳道标签 %[1]q 是主导/泳道改名前二进制文件的遗留词汇 —— " +
			"请结束本会话，用 'moai factory runs --retire %[2]s' 退役该 run 后重新启动。",
		laneLabelUnbind: "stale run：泳道标签 %[1]q 是主导/泳道改名前二进制文件的遗留词汇，且工厂 run %[2]s 的实测状态为 %[3]s（not active）—— " +
			"本会话已从工厂消息通道解绑，此后不再为本会话发出 stale-run 提示。",
		laneLabelUnbindRebind: "本项目存在活跃的工厂 run —— 若要重新加入其槽位，请结束本会话并用 'moai cc -l' 重新启动，它会占用下一个空闲槽位。",
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

// staleRunNotice renders the stale-run message. The factory retire step is
// named only for factory sessions (the run id + MOAI_FACTORY_WORKERS
// discriminator): a kanban run has no factory run to retire — its remedy is
// ending the session and relaunching.
func staleRunNotice(value, lang string) string {
	runID := strings.TrimSpace(os.Getenv(config.EnvFactoryRunID))
	m := staleRunMessagesFor(lang)
	if runID != "" && os.Getenv(config.EnvMoaiFactoryWorkers) != "" {
		return fmt.Sprintf(m.roleValueRetire, value, runID)
	}
	return fmt.Sprintf(m.roleValueRelaunch, value)
}

// legacyFactoryHookNotice is the factory message hook's stale-run answer: a
// lane whose label is legacy never registers a broker peer — it reports the
// stale-run condition instead. The SessionStart surface passes the operator's
// locale; the broker hook surface (agent-facing additionalContext) passes
// langEnglish.
func legacyFactoryHookNotice(label, runID, lang string) string {
	if !kanban.IsLegacyFactoryRoleValue(strings.TrimSpace(label)) {
		return ""
	}
	if runID != "" {
		return fmt.Sprintf(staleRunMessagesFor(lang).laneLabelRetire, label, runID)
	}
	return staleRunNotice(label, lang)
}
