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
//
// The leadLoop field (t85) embeds its protocol tokens — the queue path,
// `moai todo list`, SendMessage, ANTHROPIC_DEFAULT_*_MODEL, and the
// cache-aware-execution directive citation — verbatim in every locale,
// following the glmSubstitute precedent: those strings are addresses the
// operator and the lead must resolve identically, not prose to translate.
type factoryMessages struct {
	leadHeader      string // run id
	leadIdentity    string // lead label
	leadManual      string // worker count ×2
	glmSubstitute   string // worker count ×2
	leaderSocket    string // socket path
	leadLoop        string // lead-loop discipline: poll / stagger / no model override
	leadQueueSummary string // queued card count
	leadFreeSlots   string // free-slot label list
	leadSlotsNone   string // rendered when every slot is claimed
	settingsAuto    string
	settingsVerify  string
	workerJoin      string // worker label, worker count
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
		glmSubstitute: "Substitute 'moai glm -f %d --name ...' for 'moai cc -f %d --name ...' on any worker to run it on the GLM backend.",
		leaderSocket:  "Leader socket: %s",
		leadLoop: "Lead loop: poll the backlog queue (.moai/state/kanban/backlog.json, or `moai todo list`); " +
			"when a queued card exists AND a worker slot is free, pick the card and dispatch it to that worker " +
			"by cross-session SendMessage to the worker name; repeat.\n" +
			"Stagger — MUST: never activate every worker at once. Activate the first worker, wait until it has " +
			"started producing output (first job or visible progress), then activate the remaining free-slot " +
			"workers. Concurrent requests cannot read a cache entry still being written " +
			"(cache-aware-execution directive 2).\n" +
			"No model override — MUST: dispatches carry NO model override. The GLM tier mapping rides the " +
			"ANTHROPIC_DEFAULT_*_MODEL slot env; a per-spawn override splits the caches and can bypass the " +
			"slot-to-GLM mapping.",
		leadQueueSummary: "Factory backlog: %d waiting — run `moai todo list` to view the queue.",
		leadFreeSlots:    "Free worker slots right now: %s.",
		leadSlotsNone:    "none — every slot is held by a live session",
		settingsAuto:     "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify:   "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		workerJoin:       "Factory Mode: joined a %d-worker run as %s.",
	},
	"ko": {
		leadHeader:   "팩토리 모드: run %s, 리더 세션.",
		leadIdentity: "이 세션의 이름은 %s 여야 합니다. 이름이 다르면 아래 실행 명령은 다른 run 의 것이니, 복사하지 말고 해당 이름으로 다시 띄우세요.",
		leadManual: "이 세션이 세션 간 메시지로 카드를 워커 %d명에게 배분합니다.\n" +
			"아래 워커는 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.\n" +
			"워커 이름은 worker-1..worker-%d 이며, 생존 세션이 이미 쓰고 있는 번호는 다음 빈 번호로 늘어납니다.",
		glmSubstitute: "워커를 GLM 백엔드로 돌리려면 'moai cc -f %d --name ...' 대신 'moai glm -f %d --name ...' 을 사용하세요.",
		leaderSocket:  "리더 소켓: %s",
		leadLoop: "리더 루프: 백로그 큐(.moai/state/kanban/backlog.json, 또는 `moai todo list`)를 폴링하세요. " +
			"대기 중인 카드가 있고 워커 슬롯이 비어 있을 때 카드를 골라 해당 워커에게 세션 간 SendMessage 로 배분하고, 이를 반복합니다.\n" +
			"순차 기동 — 필수: 절대 모든 워커를 동시에 활성화하지 마세요. 첫 번째 워커를 먼저 활성화하고, " +
			"실제 출력을 내기 시작했다는 증거(첫 작업 수행 또는 진행 흔적)를 확인한 뒤 나머지 빈 슬롯의 워커를 활성화하세요. " +
			"동시 요청은 아직 기록 중인 캐시 항목을 읽을 수 없습니다(cache-aware-execution directive 2).\n" +
			"모델 오버라이드 금지 — 필수: 배분 메시지에 모델 오버라이드를 실어 보내지 마세요. " +
			"GLM 티어 매핑은 ANTHROPIC_DEFAULT_*_MODEL 슬롯 환경변수를 타고 들어가며, 스폰마다 오버라이드를 붙이면 " +
			"캐시가 쪼개지고 슬롯→GLM 매핑을 우회할 수 있습니다.",
		leadQueueSummary: "팩토리 백로그: %d장 대기 중 — `moai todo list` 를 실행하면 큐를 볼 수 있습니다.",
		leadFreeSlots:    "현재 빈 워커 슬롯: %s.",
		leadSlotsNone:    "없음 — 모든 슬롯을 생존 세션이 사용 중입니다",
		settingsAuto:     "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify:   "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		workerJoin:       "팩토리 모드: 워커 %d명 런에 %s 로 합류했습니다.",
	},
	"ja": {
		leadHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leadIdentity: "このセッションの名前は %s である必要があります。名前が異なる場合、以下の起動コマンドは別の run のものなので、コピーせずにその名前で起動し直してください。",
		leadManual: "このセッションが、セッション間メッセージでカードをワーカー %d 台に割り振ります。\n" +
			"以下のワーカーは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。\n" +
			"ワーカー名は worker-1..worker-%d で、生存セッションが保持する番号は次の空き番号へ繰り上がります。",
		glmSubstitute: "ワーカーを GLM バックエンドで動かす場合は、'moai cc -f %d --name ...' の代わりに 'moai glm -f %d --name ...' を使用してください。",
		leaderSocket:  "リーダーソケット: %s",
		leadLoop: "リーダーループ: バックログキュー(.moai/state/kanban/backlog.json、または `moai todo list`)をポーリングしてください。 " +
			"待機中のカードがあり、かつワーカースロットが空いているときにカードを選び、そのワーカーへセッション間 SendMessage で割り振ります。これを繰り返します。\n" +
			"段階的起動 — 必須: すべてのワーカーを同時に起動してはいけません。最初のワーカーを起動し、実際に出力を生成し始めた証拠 " +
			"(最初のジョブまたは進行の形跡)を確認してから、残りの空きスロットのワーカーを起動してください。 " +
			"同時リクエストは書き込み中のキャッシュエントリを読めません(cache-aware-execution directive 2)。\n" +
			"モデルオーバーライド禁止 — 必須: 割り振りにモデルオーバーライドを含めないでください。 " +
			"GLM のティアマッピングは ANTHROPIC_DEFAULT_*_MODEL スロット環境変数経由で渡り、スポーンごとのオーバーライドは " +
			"キャッシュを分割し、スロット→GLM マッピングを迂回するおそれがあります。",
		leadQueueSummary: "ファクトリーバックログ: %d件が待機中 — `moai todo list` を実行するとキューを確認できます。",
		leadFreeSlots:    "現在の空きワーカースロット: %s。",
		leadSlotsNone:    "なし — すべてのスロットを生存セッションが保持しています",
		settingsAuto:     "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify:   "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		workerJoin:       "ファクトリーモード: ワーカー %d 台の run に %s として参加しました。",
	},
	"zh": {
		leadHeader:   "工厂模式：run %s，主导会话。",
		leadIdentity: "本会话的名称应为 %s。若名称不同，下面的启动命令属于另一个 run，请不要复制，改用该名称重新启动。",
		leadManual: "本会话通过跨会话消息把卡片分发给 %d 个工作会话。\n" +
			"下面的工作会话需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。\n" +
			"工作会话命名为 worker-1..worker-%d；已被存活会话占用的编号会顺延到下一个空位。",
		glmSubstitute: "如需让某个工作会话运行在 GLM 后端，请将 'moai cc -f %d --name ...' 换成 'moai glm -f %d --name ...'。",
		leaderSocket:  "主导会话套接字：%s",
		leadLoop: "主导循环：轮询待办队列(.moai/state/kanban/backlog.json，或 `moai todo list`)。 " +
			"当有等待中的卡片且存在空闲工作槽时，选取该卡片，通过跨会话 SendMessage 派发给对应工作会话；如此反复。\n" +
			"分批启动 — 必须：绝不要同时激活所有工作会话。先激活第一个工作会话，等到它确实开始产出 " +
			"(首个任务或可见进展)之后，再激活其余空闲槽位的工作会话。 并发请求无法读取仍在写入的缓存条目 " +
			"(cache-aware-execution directive 2)。\n" +
			"禁止模型覆盖 — 必须：派发内容不携带模型覆盖。 GLM 档位映射依赖 ANTHROPIC_DEFAULT_*_MODEL 槽位环境变量； " +
			"逐次覆盖会拆分缓存，并可能绕过槽位→GLM 映射。",
		leadQueueSummary: "工厂待办队列：%d 张卡片在等待 — 运行 `moai todo list` 可查看队列。",
		leadFreeSlots:    "当前空闲工作槽：%s。",
		leadSlotsNone:    "无 — 所有槽位均被存活会话占用",
		settingsAuto:     "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify:   "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		workerJoin:       "工厂模式：已以 %s 身份加入 %d 个工作会话的 run。",
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
