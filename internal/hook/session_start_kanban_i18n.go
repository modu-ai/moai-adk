package hook

// session_start_kanban_i18n.go holds the operator-facing prose of the Kanban
// Mode bootstrap notice, one message set per locale.
//
// Only the prose lives here. The launch commands, the run id, the socket path,
// and the SPEC identifier are protocol tokens the operator copies verbatim into
// a terminal — translating them would break the paste, so they stay in the
// builder (session_start_kanban.go) and never enter this table.
//
// Which locale applies is decided by the channel, not by preference. The notice
// has two audiences reading two different surfaces, and .moai/config/sections/
// language.yaml already draws that line: `agent_prompt_language` (English) for
// what an agent reads, `conversation_language` for what the operator reads. The
// caller renders the additionalContext copy in English and the systemMessage
// copy in the operator's language.

// langEnglish is the fallback locale and the language of every agent-facing
// copy. A locale absent from kanbanLocales resolves here rather than yielding an
// empty notice — an English instruction the operator can still act on beats no
// instruction at all.
const langEnglish = "en"

// kanbanMessages is the operator-facing prose of one locale. Fields carrying a
// %s are format strings; the verb is applied by the builder.
type kanbanMessages struct {
	leadHeader     string // run id
	leadManual     string
	glmSubstitute  string
	leaderSocket   string // socket path
	settingsAuto   string
	settingsVerify string
	specLine       string // SPEC identifier
	epicPointer    string
	companionJoin  string // run id
}

// kanbanLocales is the conversation-language table. Its four entries are the
// COMPLETE set of conversation languages (en / ko / ja / zh) — the same set the
// docs-site and the README locale families carry. It is unrelated to the 16
// supported PROGRAMMING languages, which govern template neutrality and never
// touch prose. The langEnglish fallback therefore exists for unset or malformed
// configuration, not for some population of unsupported locales.
var kanbanLocales = map[string]kanbanMessages{
	langEnglish: {
		leadHeader: "Kanban Mode: run %s, lead session.\n",
		leadManual: "This session drives the chain; the four companions below are launched by hand, " +
			"one per new terminal, because a session cannot launch another session.\n",
		glmSubstitute:  "Substitute 'moai glm -k --name ...' for 'moai cc -k --name ...' on any companion to run it on the GLM backend.\n",
		leaderSocket:   "Leader socket: %s\n",
		settingsAuto:   "Cross-session messages are auto-accepted via the injected --settings.\n",
		settingsVerify: "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.\n",
		specLine:       "SPEC: %s",
		epicPointer:    "Epic context: run `moai epic status <prefix>` for a disk-grounded milestone map.\n",
		companionJoin:  "Kanban Mode: joined run %s.",
	},
	"ko": {
		leadHeader: "칸반 모드: run %s, 리더 세션.\n",
		leadManual: "이 세션이 체인을 주도합니다. 아래 네 개의 동반 세션은 터미널을 하나씩 새로 열어 직접 실행하세요 — " +
			"세션은 다른 세션을 띄울 수 없습니다.\n",
		glmSubstitute:  "동반 세션을 GLM 백엔드로 돌리려면 'moai cc -k --name ...' 대신 'moai glm -k --name ...' 을 사용하세요.\n",
		leaderSocket:   "리더 소켓: %s\n",
		settingsAuto:   "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.\n",
		settingsVerify: "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.\n",
		specLine:       "SPEC: %s",
		epicPointer:    "에픽 맥락: `moai epic status <prefix>` 를 실행하면 디스크 기반 마일스톤 지도를 볼 수 있습니다.\n",
		companionJoin:  "칸반 모드: run %s 에 합류했습니다.",
	},
	"ja": {
		leadHeader: "かんばんモード: run %s、リーダーセッション。\n",
		leadManual: "このセッションがチェーンを進行します。以下の 4 つの併走セッションは、ターミナルを 1 つずつ新規に開いて手動で起動してください — " +
			"セッションが別のセッションを起動することはできません。\n",
		glmSubstitute:  "併走セッションを GLM バックエンドで動かす場合は、'moai cc -k --name ...' の代わりに 'moai glm -k --name ...' を使用してください。\n",
		leaderSocket:   "リーダーソケット: %s\n",
		settingsAuto:   "セッション間メッセージは、注入された --settings により自動的に受理されます。\n",
		settingsVerify: "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。\n",
		specLine:       "SPEC: %s",
		epicPointer:    "Epic の文脈: `moai epic status <prefix>` を実行すると、ディスク上の情報に基づくマイルストーン一覧が得られます。\n",
		companionJoin:  "かんばんモード: run %s に参加しました。",
	},
	"zh": {
		leadHeader: "看板模式：run %s，主导会话。\n",
		leadManual: "本会话负责推进整条链路。下面四个协同会话需要各自新开一个终端手动启动 —— " +
			"会话无法启动另一个会话。\n",
		glmSubstitute:  "如需让某个协同会话运行在 GLM 后端，请将 'moai cc -k --name ...' 换成 'moai glm -k --name ...'。\n",
		leaderSocket:   "主导会话套接字：%s\n",
		settingsAuto:   "跨会话消息通过注入的 --settings 自动接受。\n",
		settingsVerify: "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。\n",
		specLine:       "SPEC: %s",
		epicPointer:    "Epic 上下文：运行 `moai epic status <prefix>` 可获取基于磁盘的里程碑视图。\n",
		companionJoin:  "看板模式：已加入 run %s。",
	},
}

// kanbanMessagesFor resolves a locale to its message set, falling back to
// English for anything the table does not carry — including the empty string,
// which is what an unconfigured or unreadable language.yaml yields.
func kanbanMessagesFor(lang string) kanbanMessages {
	if m, ok := kanbanLocales[lang]; ok {
		return m
	}
	return kanbanLocales[langEnglish]
}

// operatorLang reports the locale the operator reads, or langEnglish when the
// configuration is unavailable.
//
// Fail-open like the rest of the hook: a nil provider, a nil config, or an empty
// conversation_language degrades to English rather than failing the session
// start. kanbanMessagesFor absorbs an unknown value, so no validation happens
// here — the locale is passed through as configured.
func operatorLang(cfg ConfigProvider) string {
	if cfg == nil {
		return langEnglish
	}
	c := cfg.Get()
	if c == nil || c.Language.ConversationLanguage == "" {
		return langEnglish
	}
	return c.Language.ConversationLanguage
}
