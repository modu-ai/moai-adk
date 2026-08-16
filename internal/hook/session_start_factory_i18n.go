package hook

// session_start_factory_i18n.go holds the operator-facing prose of the
// Factory Mode bootstrap notice, one message set per locale — the factory
// sibling of session_start_kanban_i18n.go.
//
// The same two invariants govern it. Only the prose lives here: the launch
// commands, the run id, the socket path, and the worker labels are protocol
// tokens the operator copies verbatim, so they stay in the builder and never
// enter this table. And no field carries a leading or trailing newline: the
// builder joins lines within a block and blank-separates the blocks, so the
// notice is laid out identically in every locale by construction.
//
// The settingsAuto / settingsVerify / leaderSocket fields carry the same
// wording as their kanban counterparts — the sentences describe the same
// injected-settings mechanism and the same socket surface, and keeping them
// in this table (rather than referencing kanbanLocales) keeps each notice
// self-contained and free to drift when one mode's mechanism does.

// factoryMessages is the operator-facing prose of one locale.
//
// Fields carrying %s / %d are format strings. leadManual and glmSubstitute
// each carry TWO %d (the worker count twice — the sentence names it in two
// places); the per-locale word order is why the count is a format argument
// rather than pre-rendered text.
type factoryMessages struct {
	leadHeader     string // run id
	leadIdentity   string // lead label
	leadManual     string // worker count ×2
	glmSubstitute  string // worker count ×2
	leaderSocket   string // socket path
	settingsAuto   string
	settingsVerify string
	workerJoin     string // worker label, worker count
}

// factoryLocales is the conversation-language table; its four entries are the
// complete set of conversation languages, falling back to English for
// anything the table does not carry (kanbanMessagesFor holds the same
// contract).
var factoryLocales = map[string]factoryMessages{
	langEnglish: {
		leadHeader:   "Factory Mode: run %s, lead session.",
		leadIdentity: "This session should be named %s. If its name differs, the launch commands below belong to another run — relaunch with that name rather than copying them.",
		leadManual: "This session dispatches cards to %d workers over cross-session messages.\n" +
			"The workers below are launched by hand, one per new terminal, because a session cannot launch another session.\n" +
			"Workers are named worker-1..worker-%d; a number whose label is held by a live session is bumped to the next free number.",
		glmSubstitute:  "Substitute 'moai glm -f %d --name ...' for 'moai cc -f %d --name ...' on any worker to run it on the GLM backend.",
		leaderSocket:   "Leader socket: %s",
		settingsAuto:   "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify: "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		workerJoin:     "Factory Mode: joined a %d-worker run as %s.",
	},
	"ko": {
		leadHeader:   "팩토리 모드: run %s, 리더 세션.",
		leadIdentity: "이 세션의 이름은 %s 여야 합니다. 이름이 다르면 아래 실행 명령은 다른 run 의 것이니, 복사하지 말고 해당 이름으로 다시 띄우세요.",
		leadManual: "이 세션이 세션 간 메시지로 카드를 워커 %d명에게 배분합니다.\n" +
			"아래 워커는 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.\n" +
			"워커 이름은 worker-1..worker-%d 이며, 생존 세션이 이미 쓰고 있는 번호는 다음 빈 번호로 늘어납니다.",
		glmSubstitute:  "워커를 GLM 백엔드로 돌리려면 'moai cc -f %d --name ...' 대신 'moai glm -f %d --name ...' 을 사용하세요.",
		leaderSocket:   "리더 소켓: %s",
		settingsAuto:   "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify: "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		workerJoin:     "팩토리 모드: 워커 %d명 런에 %s 로 합류했습니다.",
	},
	"ja": {
		leadHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leadIdentity: "このセッションの名前は %s である必要があります。名前が異なる場合、以下の起動コマンドは別の run のものなので、コピーせずにその名前で起動し直してください。",
		leadManual: "このセッションが、セッション間メッセージでカードをワーカー %d 台に割り振ります。\n" +
			"以下のワーカーは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。\n" +
			"ワーカー名は worker-1..worker-%d で、生存セッションが保持する番号は次の空き番号へ繰り上がります。",
		glmSubstitute:  "ワーカーを GLM バックエンドで動かす場合は、'moai cc -f %d --name ...' の代わりに 'moai glm -f %d --name ...' を使用してください。",
		leaderSocket:   "リーダーソケット: %s",
		settingsAuto:   "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify: "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		workerJoin:     "ファクトリーモード: ワーカー %d 台の run に %s として参加しました。",
	},
	"zh": {
		leadHeader:   "工厂模式：run %s，主导会话。",
		leadIdentity: "本会话的名称应为 %s。若名称不同，下面的启动命令属于另一个 run，请不要复制，改用该名称重新启动。",
		leadManual: "本会话通过跨会话消息把卡片分发给 %d 个工作会话。\n" +
			"下面的工作会话需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。\n" +
			"工作会话命名为 worker-1..worker-%d；已被存活会话占用的编号会顺延到下一个空位。",
		glmSubstitute:  "如需让某个工作会话运行在 GLM 后端，请将 'moai cc -f %d --name ...' 换成 'moai glm -f %d --name ...'。",
		leaderSocket:   "主导会话套接字：%s",
		settingsAuto:   "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify: "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		workerJoin:     "工厂模式：已以 %s 身份加入 %d 个工作会话的 run。",
	},
}

// factoryMessagesFor resolves a locale to its message set, falling back to
// English for anything the table does not carry — including the empty string,
// which is what an unconfigured or unreadable language.yaml yields.
func factoryMessagesFor(lang string) factoryMessages {
	if m, ok := factoryLocales[lang]; ok {
		return m
	}
	return factoryLocales[langEnglish]
}
