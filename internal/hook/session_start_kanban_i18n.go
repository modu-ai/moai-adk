package hook

// session_start_kanban_i18n.go holds the operator-facing prose of the Kanban
// Mode bootstrap notice, one message set per locale.
//
// Only the prose lives here. The launch commands, the run id, the socket path,
// and the SPEC identifier are protocol tokens the operator copies verbatim into
// a terminal — translating them would break the paste, so they stay in the
// builder (session_start_kanban.go) and never enter this table.
//
// Layout is not here either. No field carries a leading or trailing newline: the
// builder joins lines within a block and blank-separates the blocks, so the
// notice is laid out identically in every locale by construction. A locale
// cannot drift into its own spacing, and a field cannot silently weld itself to
// the next line by forgetting a terminator.
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

// kanbanMessages is the operator-facing prose of one locale.
//
// Fields carrying a %s or %d are format strings. No field carries a trailing
// newline; leadManual carries one internal newline because it is two sentences
// the operator reads as separate lines, and backendRecommend carries them
// because it is a header line, one indented row per role, and a trailer —
// a block the operator scans as a table, not a wrapped sentence.
type kanbanMessages struct {
	leadHeader       string // run id
	leadIdentity     string // lead label
	leadManual       string
	glmSubstitute    string
	backendRecommend string // recommended backend per role; internal newlines
	agentFanout      string
	nameChoices      string
	leaderSocket     string // socket path
	settingsAuto     string
	settingsVerify   string
	specLine         string // SPEC identifier
	backlogSummary   string // queued card count
	companionJoin    string // launch label
}

// kanbanLocales is the conversation-language table. Its four entries are the
// COMPLETE set of conversation languages (en / ko / ja / zh) — the same set the
// docs-site and the README locale families carry. It is unrelated to the 16
// supported PROGRAMMING languages, which govern template neutrality and never
// touch prose. The langEnglish fallback therefore exists for unset or malformed
// configuration, not for some population of unsupported locales.
var kanbanLocales = map[string]kanbanMessages{
	langEnglish: {
		leadHeader:   "Kanban Mode: run %s, lead session.",
		leadIdentity: "This session should be named %s. If its name differs, the launch commands below belong to another run — relaunch with that name rather than copying them.",
		leadManual: "This session drives the kanban chain.\n" +
			"The three companions below are launched by hand, one per new terminal, " +
			"because a session cannot launch another session.",
		glmSubstitute: "Entry points: `moai cc -k` is the Claude-backend foreman, `moai glm -k` the GLM-backend " +
			"foreman — the launcher picks the backend, `-k` the kanban role. Substitute 'moai glm -k --name ...' " +
			"for 'moai cc -k --name ...' on any companion to run it on the GLM backend.",
		backendRecommend: "Recommended mix (token availability first — one GLM account, one Claude account):\n" +
			"  lead      → GLM\n" +
			"  plan      → Claude (Opus)\n" +
			"  run       → GLM\n" +
			"  sync      → Claude (Opus)\n" +
			"Any other mix, or one backend everywhere, works just as well.",
		agentFanout:    "Every companion can run up to 10 agents concurrently in parallel.",
		nameChoices:    "Name options: the default Agent worker / `judge` (Claude verdicts — the GLM foreman's only Claude path) / `lane-N` (numbered lanes).",
		leaderSocket:   "Leader socket: %s",
		settingsAuto:   "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify: "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		specLine:       "SPEC: %s",
		backlogSummary: "Kanban backlog: %d waiting — run `moai todo` to view the queue.",
		companionJoin:  "Kanban Mode: joined the kanban run as %s.",
	},
	"ko": {
		leadHeader:   "칸반 모드: run %s, 리더 세션.",
		leadIdentity: "이 세션의 이름은 %s 여야 합니다. 이름이 다르면 아래 실행 명령은 다른 run 의 것이니, 복사하지 말고 해당 이름으로 다시 띄우세요.",
		leadManual: "이 세션이 칸반 체인을 주도합니다.\n" +
			"아래 세 개의 동반 세션은 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.",
		glmSubstitute: "진입점: `moai cc -k` 는 Claude 백엔드 공장장, `moai glm -k` 는 GLM 백엔드 공장장 — 런처가 백엔드를, `-k` 가 칸반 역할을 정합니다. " +
			"동반 세션을 GLM 백엔드로 돌리려면 'moai cc -k --name ...' 대신 'moai glm -k --name ...' 을 사용하세요.",
		backendRecommend: "추천 조합 (토큰 가용성 우선 — GLM·Claude 계정은 각각 1개씩 사용 가능):\n" +
			"  lead      → GLM\n" +
			"  plan      → Claude (Opus)\n" +
			"  run       → GLM\n" +
			"  sync      → Claude (Opus)\n" +
			"다른 조합이나 전 세션 한 백엔드 통일도 무방합니다.",
		agentFanout:    "각 동반 세션은 최대 10개의 에이전트를 동시에 병렬로 실행할 수 있습니다.",
		nameChoices:    "이름 선택지: 기본 Agent 작업자 / `judge` (Claude 판정 — GLM 공장장이 Claude 를 쓰는 유일한 경로) / `lane-N` (번호 레인).",
		leaderSocket:   "리더 소켓: %s",
		settingsAuto:   "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify: "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		specLine:       "SPEC: %s",
		backlogSummary: "칸반 백로그: %d장 대기 중 — `moai todo` 를 실행하면 큐를 볼 수 있습니다.",
		companionJoin:  "칸반 모드: 칸반 run 에 %s 로 합류했습니다.",
	},
	"ja": {
		leadHeader:   "かんばんモード: run %s、リーダーセッション。",
		leadIdentity: "このセッションの名前は %s である必要があります。名前が異なる場合、以下の起動コマンドは別の run のものなので、コピーせずにその名前で起動し直してください。",
		leadManual: "このセッションがかんばんチェーンを進行します。\n" +
			"以下の 3 つの併走セッションは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。",
		glmSubstitute: "入口: `moai cc -k` は Claude バックエンドの親方、`moai glm -k` は GLM バックエンドの親方 — ランチャーがバックエンドを、`-k` がかんばんの役割を決めます。 " +
			"併走セッションを GLM バックエンドで動かす場合は、'moai cc -k --name ...' の代わりに 'moai glm -k --name ...' を使用してください。",
		backendRecommend: "推奨組み合わせ（トークン余裕優先 — GLM・Claude アカウントは各1つずつ使用可能）:\n" +
			"  lead      → GLM\n" +
			"  plan      → Claude (Opus)\n" +
			"  run       → GLM\n" +
			"  sync      → Claude (Opus)\n" +
			"他の組み合わせや、全セッションを同一バックエンドに統一しても問題ありません。",
		agentFanout:    "各併走セッションは最大 10 個のエージェントを同時に並列実行できます。",
		nameChoices:    "名前の選択肢: デフォルトの Agent 作業者 / `judge`（Claude による判定 — GLM 親方が Claude を使う唯一の経路） / `lane-N`（番号付きレーン）。",
		leaderSocket:   "リーダーソケット: %s",
		settingsAuto:   "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify: "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		specLine:       "SPEC: %s",
		backlogSummary: "かんばんバックログ: %d件が待機中 — `moai todo` を実行するとキューを確認できます。",
		companionJoin:  "かんばんモード: かんばん run に %s として参加しました。",
	},
	"zh": {
		leadHeader:   "看板模式：run %s，主导会话。",
		leadIdentity: "本会话的名称应为 %s。若名称不同，下面的启动命令属于另一个 run，请不要复制，改用该名称重新启动。",
		leadManual: "本会话负责推进整条看板链路。\n" +
			"下面三个协同会话需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。",
		glmSubstitute: "入口：`moai cc -k` 是 Claude 后端的工头，`moai glm -k` 是 GLM 后端的工头 —— 启动器决定后端，`-k` 决定看板角色。 " +
			"如需让某个协同会话运行在 GLM 后端，请将 'moai cc -k --name ...' 换成 'moai glm -k --name ...'。",
		backendRecommend: "推荐组合（优先考虑令牌余量 —— GLM·Claude 账号各有一个可用）:\n" +
			"  lead      → GLM\n" +
			"  plan      → Claude (Opus)\n" +
			"  run       → GLM\n" +
			"  sync      → Claude (Opus)\n" +
			"其他组合、或将全会话统一到同一后端同样可行。",
		agentFanout:    "每个协同会话最多可同时并行运行 10 个代理。",
		nameChoices:    "命名选项：默认 Agent 作业者 / `judge`（Claude 判定 —— GLM 工头使用 Claude 的唯一途径） / `lane-N`（编号泳道）。",
		leaderSocket:   "主导会话套接字：%s",
		settingsAuto:   "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify: "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		specLine:       "SPEC: %s",
		backlogSummary: "看板待办队列：%d 张卡片在等待 — 运行 `moai todo` 可查看队列。",
		companionJoin:  "看板模式：已以 %s 身份加入看板 run。",
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
