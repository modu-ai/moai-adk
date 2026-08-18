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
// Fields carrying %s / %d are format strings. leadManual and entryGuide each
// carry TWO %d (the worker count twice — the sentence names it in two
// places); the per-locale word order is why the count is a format argument
// rather than pre-rendered text. workerJoin pins its argument order with
// explicit %[1]s / %[2]d indices because the locales' natural word orders
// differ (en/ja/ko say the count first, zh the label first) — see
// factoryWorkerNotice.
//
// The discipline fields (leadClasses / leadStagger / leadNoOverride) embed
// their protocol tokens — `/loop`, the stage names, ANTHROPIC_DEFAULT_*_MODEL,
// CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS, and the cache-aware-execution
// directive citation — verbatim in every locale, following the entryGuide
// precedent: those strings are addresses the operator and the lead must
// resolve identically, not prose to translate.
type factoryMessages struct {
	leadHeader        string // run id
	leadIdentity      string // lead label
	leadManual        string // worker count ×2
	entryGuide        string // cc/glm entry points, -f forms; worker count ×2
	agentFanout       string // per-lane concurrent agent cap
	leaderSocket      string // socket path
	leadClasses       string // card-class routing: A/B wholesale, C serial 3-stage
	leadStagger       string // fan-out-only staggered activation
	leadNoOverride    string // no model override on dispatches
	leadFreeSlots     string // free-slot label list
	leadSlotsNone     string // rendered when every slot is claimed
	settingsAuto      string
	settingsVerify    string
	workerJoin        string // worker label %[1]s, worker count %[2]d
	workerJoinNoCount string // worker label only — the incremental form
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
		entryGuide: "Entry points: `moai cc -f %d` is the Claude-backend factory lead, `moai glm -f %d` the GLM-backend one — " +
			"the launcher picks the backend, `-f` the factory. `-f` with no count starts the one-worker default; " +
			"add one lane at a time as the queue demands with `moai cc -f worker-<n>` (or the glm form).",
		agentFanout:  "Every worker can run up to 10 agents concurrently in parallel.",
		leaderSocket: "Leader socket: %s",
		leadClasses: "Card routing (kanban-dispatch classes): A-class and B-class cards are fanned out WHOLESALE — " +
			"the whole card goes to one worker. C-class cards (design changes) run the serial 3-stage path " +
			"(plan -> run -> sync, one stage completing before the next begins) and are never fanned out.\n" +
			"The queue is polled and cards are picked by the operator or the kanban foreman loop (bare `/loop`); " +
			"this factory routes PICKED cards to worker lanes.",
		leadStagger: "Stagger — MUST: never activate every worker at once. Activate the first worker, wait until " +
			"it has started producing output (first job or visible progress), then activate the remaining " +
			"free-slot workers. Concurrent requests cannot read a cache entry still being written " +
			"(cache-aware-execution directive 2). This rule governs FACTORY fan-out only — the workflow " +
			"runtime staggers itself (CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS) and is not governed here.",
		leadNoOverride: "No model override — MUST: dispatches carry NO model override. The GLM tier mapping rides " +
			"the ANTHROPIC_DEFAULT_*_MODEL slot env; a per-spawn override splits the caches and can bypass " +
			"the slot-to-GLM mapping.",
		leadFreeSlots:     "Free worker slots right now: %s.",
		leadSlotsNone:     "none — every slot is held by a live session",
		settingsAuto:      "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify:    "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		workerJoin:        "Factory Mode: joined a %[2]d-worker run as %[1]s.",
		workerJoinNoCount: "Factory Mode: joined the factory run as %[1]s.",
	},
	"ko": {
		leadHeader:   "팩토리 모드: run %s, 리더 세션.",
		leadIdentity: "이 세션의 이름은 %s 여야 합니다. 이름이 다르면 아래 실행 명령은 다른 run 의 것이니, 복사하지 말고 해당 이름으로 다시 띄우세요.",
		leadManual: "이 세션이 세션 간 메시지로 카드를 워커 %d명에게 배분합니다.\n" +
			"아래 워커는 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.\n" +
			"워커 이름은 worker-1..worker-%d 이며, 생존 세션이 이미 쓰고 있는 번호는 다음 빈 번호로 늘어납니다.",
		entryGuide: "진입점: `moai cc -f %d` 는 Claude 백엔드 팩토리 리더, `moai glm -f %d` 는 GLM 백엔드 팩토리 리더 — " +
			"런처가 백엔드를, `-f` 가 팩토리를 정합니다. `-f` 에 개수를 붙이지 않으면 워커 1개 기본으로 시작하고, " +
			"필요에 따라 `moai cc -f worker-<n>` (또는 glm 형태)로 레인을 한 명씩 추가하세요.",
		agentFanout:  "각 워커는 최대 10개의 에이전트를 동시에 병렬로 실행할 수 있습니다.",
		leaderSocket: "리더 소켓: %s",
		leadClasses: "카드 라우팅(칸반 디스패치 분류): A/B등급 카드는 통째로 팬아웃합니다 — 카드 전체가 워커 하나에 갑니다. " +
			"C등급 카드(설계 변경)는 직렬 3단계 경로(plan -> run -> sync, 한 단계가 끝나야 다음 단계)로 진행하며 절대 통째로 팬아웃하지 않습니다.\n" +
			"큐 폴링과 카드 선택은 운영자 또는 칸반 포어맨 루프(단독 `/loop`)가 담당합니다. 이 팩토리는 선택된(picked) 카드를 워커 레인으로 라우팅합니다.",
		leadStagger: "순차 기동 — 필수: 절대 모든 워커를 동시에 활성화하지 마세요. 첫 번째 워커를 먼저 활성화하고, " +
			"실제 출력을 내기 시작했다는 증거(첫 작업 수행 또는 진행 흔적)를 확인한 뒤 나머지 빈 슬롯의 워커를 활성화하세요. " +
			"동시 요청은 아직 기록 중인 캐시 항목을 읽을 수 없습니다(cache-aware-execution directive 2). " +
			"이 규칙은 팩토리 팬아웃에만 적용됩니다 — 워크플로 런타임은 스스로 스태거합니다(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS).",
		leadNoOverride: "모델 오버라이드 금지 — 필수: 배분 메시지에 모델 오버라이드를 실어 보내지 마세요. " +
			"GLM 티어 매핑은 ANTHROPIC_DEFAULT_*_MODEL 슬롯 환경변수를 타고 들어가며, 스폰마다 오버라이드를 붙이면 " +
			"캐시가 쪼개지고 슬롯→GLM 매핑을 우회할 수 있습니다.",
		leadFreeSlots:     "현재 빈 워커 슬롯: %s.",
		leadSlotsNone:     "없음 — 모든 슬롯을 생존 세션이 사용 중입니다",
		settingsAuto:      "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify:    "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		workerJoin:        "팩토리 모드: 워커 %[2]d명 런에 %[1]s 로 합류했습니다.",
		workerJoinNoCount: "팩토리 모드: 팩토리 run 에 %[1]s 로 합류했습니다.",
	},
	"ja": {
		leadHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leadIdentity: "このセッションの名前は %s である必要があります。名前が異なる場合、以下の起動コマンドは別の run のものなので、コピーせずにその名前で起動し直してください。",
		leadManual: "このセッションが、セッション間メッセージでカードをワーカー %d 台に割り振ります。\n" +
			"以下のワーカーは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。\n" +
			"ワーカー名は worker-1..worker-%d で、生存セッションが保持する番号は次の空き番号へ繰り上がります。",
		entryGuide: "入口: `moai cc -f %d` は Claude バックエンドのファクトリーリーダー、`moai glm -f %d` は GLM バックエンドのファクトリーリーダー — " +
			"ランチャーがバックエンドを、`-f` がファクトリーを決めます。`-f` に個数を付けなければワーカー1台のデフォルトで開始し、 " +
			"必要に応じて `moai cc -f worker-<n>`（または glm 形式）でレーンを1台ずつ追加してください。",
		agentFanout:  "各ワーカーは最大 10 個のエージェントを同時に並列実行できます。",
		leaderSocket: "リーダーソケット: %s",
		leadClasses: "カードルーティング(かんばんディスパッチ分類): A/Bクラスのカードは丸ごとファンアウトします — カード全体を1つのワーカーへ渡します。 " +
			"Cクラスのカード(設計変更)は直列3段階パス(plan -> run -> sync、各段階の完了後に次へ)で進め、丸ごとのファンアウトはしません。\n" +
			"キューのポーリングとカードの選択はオペレーターまたはかんばんフォアマンループ(単独の `/loop`)が担います。このファクトリーは選択済み(picked)カードをワーカーレーンへ振り分けます。",
		leadStagger: "段階的起動 — 必須: すべてのワーカーを同時に起動してはいけません。最初のワーカーを起動し、実際に出力を生成し始めた証拠 " +
			"(最初のジョブまたは進行の形跡)を確認してから、残りの空きスロットのワーカーを起動してください。 " +
			"同時リクエストは書き込み中のキャッシュエントリを読めません(cache-aware-execution directive 2)。 " +
			"このルールはファクトリーファンアウトにのみ適用されます — ワークフローランタイムは自身でスタガーします(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		leadNoOverride: "モデルオーバーライド禁止 — 必須: 割り振りにモデルオーバーライドを含めないでください。 " +
			"GLM のティアマッピングは ANTHROPIC_DEFAULT_*_MODEL スロット環境変数経由で渡り、スポーンごとのオーバーライドは " +
			"キャッシュを分割し、スロット→GLM マッピングを迂回するおそれがあります。",
		leadFreeSlots:     "現在の空きワーカースロット: %s。",
		leadSlotsNone:     "なし — すべてのスロットを生存セッションが保持しています",
		settingsAuto:      "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify:    "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		workerJoin:        "ファクトリーモード: ワーカー %[2]d 台の run に %[1]s として参加しました。",
		workerJoinNoCount: "ファクトリーモード: ファクトリー run に %[1]s として参加しました。",
	},
	"zh": {
		leadHeader:   "工厂模式：run %s，主导会话。",
		leadIdentity: "本会话的名称应为 %s。若名称不同，下面的启动命令属于另一个 run，请不要复制，改用该名称重新启动。",
		leadManual: "本会话通过跨会话消息把卡片分发给 %d 个工作会话。\n" +
			"下面的工作会话需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。\n" +
			"工作会话命名为 worker-1..worker-%d；已被存活会话占用的编号会顺延到下一个空位。",
		entryGuide: "入口：`moai cc -f %d` 是 Claude 后端的工厂主导会话，`moai glm -f %d` 是 GLM 后端的工厂主导会话 — " +
			"启动器决定后端，`-f` 决定工厂。`-f` 不带数量时以一个工作会话的默认配置启动；" +
			"需要扩容时用 `moai cc -f worker-<n>`（或 glm 形式）逐个添加工作槽。",
		agentFanout:  "每个工作会话最多可同时并行运行 10 个代理。",
		leaderSocket: "主导会话套接字：%s",
		leadClasses: "卡片路由(看板调度分级): A/B 级卡片整体分发 — 整张卡片交给一个工作会话。 " +
			"C 级卡片(设计变更)走串行三阶段路径(plan -> run -> sync，上一阶段完成后才进入下一阶段)，绝不整体分发。\n" +
			"队列轮询与卡片挑选由操作者或看板工头循环(单独的 `/loop`)负责。本工厂只把已挑选(picked)的卡片路由到工作槽。",
		leadStagger: "分批启动 — 必须: 绝不要同时激活所有工作会话。先激活第一个工作会话，等到它确实开始产出 " +
			"(首个任务或可见进展)之后，再激活其余空闲槽位的工作会话。 并发请求无法读取仍在写入的缓存条目 " +
			"(cache-aware-execution directive 2)。本规则仅约束工厂分发 — 工作流运行时会自行错峰(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		leadNoOverride: "禁止模型覆盖 — 必须: 派发内容不携带模型覆盖。 GLM 档位映射依赖 ANTHROPIC_DEFAULT_*_MODEL 槽位环境变量； " +
			"逐次覆盖会拆分缓存，并可能绕过槽位→GLM 映射。",
		leadFreeSlots:     "当前空闲工作槽：%s。",
		leadSlotsNone:     "无 — 所有槽位均被存活会话占用",
		settingsAuto:      "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify:    "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		workerJoin:        "工厂模式：已以 %[1]s 身份加入 %[2]d 个工作会话的 run。",
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
