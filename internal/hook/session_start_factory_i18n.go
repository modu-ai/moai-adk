package hook

// session_start_factory_i18n.go holds the operator-facing prose of the
// Factory Mode bootstrap notice, one message set per locale — the factory
// sibling of session_start_kanban_i18n.go.
//
// The same two invariants govern it. Only the prose lives here: the launch
// commands, the run id, the socket path, and the lane labels are protocol
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
// Fields carrying %s / %d are format strings. leadManual and entryGuide each
// carry TWO %d (the lane count twice — the sentence names it in two
// places); the per-locale word order is why the count is a format argument
// rather than pre-rendered text. workerJoin pins its argument order with
// explicit %[1]s / %[2]d indices because the locales' natural word orders
// differ (en/ja/ko say the count first, zh the label first) — see
// factoryWorkerNotice.
//
// The discipline fields (leadClasses / leadStagger) embed
// their protocol tokens — `/loop`, the stage names, and
// CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS, plus the cache-aware-execution
// directive citation — verbatim in every locale, following the entryGuide
// precedent: those strings are addresses the operator and the lead must
// resolve identically, not prose to translate.
type factoryMessages struct {
	leadHeader        string // run id
	leadIdentity      string // lead label
	leadManual        string // lane count ×2
	entryGuide        string // cc/glm entry points, -f forms; lane count ×2
	agentFanout       string // per-lane concurrent agent cap
	leaderSocket      string // socket path
	leadClasses       string // whole-card routing: one lane runs the serial 3-stage path in-session
	leadStagger       string // fan-out-only staggered activation
	leadFreeSlots     string // free-slot label list
	leadSlotsNone     string // rendered when every slot is claimed
	settingsAuto      string
	settingsVerify    string
	workerJoin        string // lane label %[1]s, lane count %[2]d
	workerJoinNoCount string // lane label only — the incremental form
}

// factoryLocales is the conversation-language table; its four entries are the
// complete set of conversation languages, falling back to English for
// anything the table does not carry (kanbanMessagesFor holds the same
// contract).
var factoryLocales = map[string]factoryMessages{
	langEnglish: {
		leadHeader:   "Factory Mode: run %s, lead session.",
		leadIdentity: "This session is named %s. It carries no run id: a second lead launched while this one is live takes the next free number instead, and peers address whichever name the session actually launched under. The session list shows that same name — the title is registered on your first prompt, and a later /rename still wins.",
		leadManual: "This session dispatches cards to %d lanes over cross-session messages.\n" +
			"The lanes below are launched by hand, one per new terminal, because a session cannot launch another session.\n" +
			"Lanes are named lane-1..lane-%d; a number whose label is held by a live session is bumped to the next free number.",
		entryGuide: "Entry points: `moai cc -f %d` is the Claude-backend factory lead, `moai glm -f %d` the GLM-backend one — " +
			"the launcher picks the backend, `-f` the factory. `-f` with no count starts the one-lane default; " +
			"add one lane at a time as the queue demands with `moai cc -f lane-<n>` (or the glm form).",
		agentFanout:  "Every lane can run up to 10 agents concurrently in parallel.",
		leaderSocket: "Leader socket: %s",
		leadClasses: "Card routing: every card is routed WHOLE to one lane, and that lane carries it through " +
			"the serial 3-stage path (plan -> run -> sync, one stage completing before the next begins) in-session — " +
			"a card is never split across lanes.\n" +
			"The queue is polled and cards are picked by the operator or the kanban foreman loop (bare `/loop`); " +
			"this factory routes PICKED cards to free lanes.",
		leadStagger: "Stagger — MUST: never activate every lane at once. Activate the first lane, wait until " +
			"it has started producing output (first job or visible progress), then activate the remaining " +
			"free-slot lanes. Concurrent requests cannot read a cache entry still being written " +
			"(cache-aware-execution directive 2). This rule governs FACTORY fan-out only — the workflow " +
			"runtime staggers itself (CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS) and is not governed here.",
		leadFreeSlots:     "Free lane slots right now: %s.",
		leadSlotsNone:     "none — every slot is held by a live session",
		settingsAuto:      "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify:    "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		workerJoin:        "Factory Mode: joined a %[2]d-lane run as %[1]s.",
		workerJoinNoCount: "Factory Mode: joined the factory run as %[1]s.",
	},
	"ko": {
		leadHeader:   "팩토리 모드: run %s, 리더 세션.",
		leadIdentity: "이 세션의 이름은 %s 입니다. 이름에 run id 는 들어가지 않습니다 — 이 세션이 살아 있는 동안 리드를 하나 더 띄우면 그쪽이 다음 번호를 받고, 다른 세션은 실제로 띄워진 이름으로 이 세션을 부릅니다. 세션 목록에도 같은 이름이 뜹니다 — 제목은 첫 프롬프트에서 등록되고, 나중에 /rename 을 하면 그쪽이 우선합니다.",
		leadManual: "이 세션이 세션 간 메시지로 카드를 레인 %d개에 배분합니다.\n" +
			"아래 레인은 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.\n" +
			"레인 이름은 lane-1..lane-%d 이며, 생존 세션이 이미 쓰고 있는 번호는 다음 빈 번호로 늘어납니다.",
		entryGuide: "진입점: `moai cc -f %d` 는 Claude 백엔드 팩토리 리더, `moai glm -f %d` 는 GLM 백엔드 팩토리 리더 — " +
			"런처가 백엔드를, `-f` 가 팩토리를 정합니다. `-f` 에 개수를 붙이지 않으면 레인 1개 기본으로 시작하고, " +
			"필요에 따라 `moai cc -f lane-<n>` (또는 glm 형태)로 레인을 한 개씩 추가하세요.",
		agentFanout:  "각 레인은 최대 10개의 에이전트를 동시에 병렬로 실행할 수 있습니다.",
		leaderSocket: "리더 소켓: %s",
		leadClasses: "카드 라우팅: 모든 카드는 한 레인에 통째로 배정되고, 그 레인이 세션 안에서 직렬 3단계 경로" +
			"(plan -> run -> sync, 한 단계가 끝나야 다음 단계)를 끝까지 수행합니다 — 카드를 여러 레인에 쪼개 배분하지 않습니다.\n" +
			"큐 폴링과 카드 선택은 운영자 또는 칸반 포어맨 루프(단독 `/loop`)가 담당합니다. 이 팩토리는 선택된(picked) 카드를 빈 레인으로 라우팅합니다.",
		leadStagger: "순차 기동 — 필수: 절대 모든 레인을 동시에 활성화하지 마세요. 첫 번째 레인을 먼저 활성화하고, " +
			"실제 출력을 내기 시작했다는 증거(첫 작업 수행 또는 진행 흔적)를 확인한 뒤 나머지 빈 슬롯의 레인을 활성화하세요. " +
			"동시 요청은 아직 기록 중인 캐시 항목을 읽을 수 없습니다(cache-aware-execution directive 2). " +
			"이 규칙은 팩토리 팬아웃에만 적용됩니다 — 워크플로 런타임은 스스로 스태거합니다(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS).",
		leadFreeSlots:     "현재 빈 레인 슬롯: %s.",
		leadSlotsNone:     "없음 — 모든 슬롯을 생존 세션이 사용 중입니다",
		settingsAuto:      "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify:    "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		workerJoin:        "팩토리 모드: 레인 %[2]d개 런에 %[1]s 로 합류했습니다.",
		workerJoinNoCount: "팩토리 모드: 팩토리 run 에 %[1]s 로 합류했습니다.",
	},
	"ja": {
		leadHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leadIdentity: "このセッションの名前は %s です。名前に run id は含まれません — このセッションが生きている間にもう一つリーダーを起動すると、そちらが次の番号を取り、他のセッションは実際に起動した名前でこのセッションを呼びます。セッション一覧にも同じ名前が表示されます — タイトルは最初のプロンプトで登録され、後から /rename すればそちらが優先されます。",
		leadManual: "このセッションが、セッション間メッセージでカードをレーン %d 本に割り振ります。\n" +
			"以下のレーンは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。\n" +
			"レーン名は lane-1..lane-%d で、生存セッションが保持する番号は次の空き番号へ繰り上がります。",
		entryGuide: "入口: `moai cc -f %d` は Claude バックエンドのファクトリーリーダー、`moai glm -f %d` は GLM バックエンドのファクトリーリーダー — " +
			"ランチャーがバックエンドを、`-f` がファクトリーを決めます。`-f` に個数を付けなければレーン1本のデフォルトで開始し、 " +
			"必要に応じて `moai cc -f lane-<n>`（または glm 形式）でレーンを1本ずつ追加してください。",
		agentFanout:  "各レーンは最大 10 個のエージェントを同時に並列実行できます。",
		leaderSocket: "リーダーソケット: %s",
		leadClasses: "カードルーティング: すべてのカードは1つのレーンへ丸ごと割り当てられ、そのレーンがセッション内で直列3段階パス" +
			"(plan -> run -> sync、各段階の完了後に次へ)を最後まで実行します — カードを複数レーンに分割することはありません。\n" +
			"キューのポーリングとカードの選択はオペレーターまたはかんばんフォアマンループ(単独の `/loop`)が担います。このファクトリーは選択済み(picked)カードを空きレーンへ振り分けます。",
		leadStagger: "段階的起動 — 必須: すべてのレーンを同時に起動してはいけません。最初のレーンを起動し、実際に出力を生成し始めた証拠 " +
			"(最初のジョブまたは進行の形跡)を確認してから、残りの空きスロットのレーンを起動してください。 " +
			"同時リクエストは書き込み中のキャッシュエントリを読めません(cache-aware-execution directive 2)。 " +
			"このルールはファクトリーファンアウトにのみ適用されます — ワークフローランタイムは自身でスタガーします(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		leadFreeSlots:     "現在の空きレーンスロット: %s。",
		leadSlotsNone:     "なし — すべてのスロットを生存セッションが保持しています",
		settingsAuto:      "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify:    "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		workerJoin:        "ファクトリーモード: レーン %[2]d 本の run に %[1]s として参加しました。",
		workerJoinNoCount: "ファクトリーモード: ファクトリー run に %[1]s として参加しました。",
	},
	"zh": {
		leadHeader:   "工厂模式：run %s，主导会话。",
		leadIdentity: "本会话的名称是 %s。名称中不含 run id —— 本会话存活期间再启动一个主导会话，后者会取下一个编号；其他会话按实际启动时的名称来称呼本会话。会话列表中也显示同一名称 —— 标题在首次提示时注册，之后 /rename 优先。",
		leadManual: "本会话通过跨会话消息把卡片分发给 %d 条泳道。\n" +
			"下面的泳道需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。\n" +
			"泳道命名为 lane-1..lane-%d；已被存活会话占用的编号会顺延到下一个空位。",
		entryGuide: "入口：`moai cc -f %d` 是 Claude 后端的工厂主导会话，`moai glm -f %d` 是 GLM 后端的工厂主导会话 — " +
			"启动器决定后端，`-f` 决定工厂。`-f` 不带数量时以一条泳道的默认配置启动；" +
			"需要扩容时用 `moai cc -f lane-<n>`（或 glm 形式）逐条添加泳道。",
		agentFanout:  "每条泳道最多可同时并行运行 10 个代理。",
		leaderSocket: "主导会话套接字：%s",
		leadClasses: "卡片路由：每张卡片整体分发给一条泳道，由该泳道在会话内走完串行三阶段路径" +
			"(plan -> run -> sync，上一阶段完成后才进入下一阶段)—— 绝不把卡片拆分到多条泳道。\n" +
			"队列轮询与卡片挑选由操作者或看板工头循环(单独的 `/loop`)负责。本工厂把已挑选(picked)的卡片路由到空闲泳道。",
		leadStagger: "分批启动 — 必须: 绝不要同时激活所有泳道。先激活第一条泳道，等到它确实开始产出 " +
			"(首个任务或可见进展)之后，再激活其余空闲槽位的泳道。 并发请求无法读取仍在写入的缓存条目 " +
			"(cache-aware-execution directive 2)。本规则仅约束工厂分发 — 工作流运行时会自行错峰(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		leadFreeSlots:     "当前空闲泳道：%s。",
		leadSlotsNone:     "无 — 所有槽位均被存活会话占用",
		settingsAuto:      "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify:    "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		workerJoin:        "工厂模式：已以 %[1]s 身份加入 %[2]d 条泳道的 run。",
		workerJoinNoCount: "工厂模式：已以 %[1]s 身份加入工厂 run。",
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
