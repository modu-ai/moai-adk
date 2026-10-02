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
// Fields carrying %s / %d are format strings. leaderManual carries the lane
// count twice; entryGuide uses %[1]d for the count and %[2]s for the fixed
// launcher entry token. laneJoin pins its argument order with
// explicit %[1]s / %[2]d indices because the locales' natural word orders
// differ (en/ja/ko say the count first, zh the label first) — see
// factoryLaneNotice.
//
// The discipline fields (leaderClasses / leaderStagger) embed
// their protocol tokens — `/loop`, the stage names, and
// CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS, plus the cache-aware-execution
// directive citation — verbatim in every locale, following the entryGuide
// precedent: those strings are addresses the operator and the leader must
// resolve identically, not prose to translate.
type factoryMessages struct {
	leaderHeader      string // run id
	leaderIdentity    string // leader label
	leaderManual      string // lane count ×2
	entryGuide        string // cc/glm/gpt entry points; count %[1]d, entry %[2]s
	agentFanout       string // per-lane concurrent agent cap
	leaderSocket      string // socket path
	leaderClasses     string // whole-card routing: one lane runs the serial 3-stage path in-session
	leaderStagger     string // fan-out-only staggered activation
	gateSummary       string // one sentence: name token + pointer
	leaderFreeSlots   string // free-slot label list
	operationalStatus string // explicit run-bound read-only operational query
	leaderSlotsNone   string // rendered when every slot is claimed
	settingsAuto      string
	settingsVerify    string
	laneJoin          string // lane label %[1]s, lane count %[2]d
	laneJoinNoCount   string // lane label only — the incremental form

	// laneNextCardRule is the Claude-harness lane's next-card rule
	// (REQ-SD-019): the rule names the six REQ-SD-014 MCP tools and their
	// CLI equivalents, so those tokens stay verbatim in every locale the
	// same way the launch commands do. laneOwnedCardRule is the
	// Codex-harness lane's owned-card rule; it names only the two CLI verbs
	// and interpolates the card id (%[1]s) — it never carries an MCP tool
	// name of any kind (REQ-SD-019's negative constraint).
	//
	// laneManualDispatchRule is the manual-mode rule a --no-auto-dispatch
	// launch carries (SPEC-TODO-CLASSIFY-DISPATCH-001 REQ-TCD-011): it
	// states the manual mode and carries NO self-dispatch instruction — it
	// names no lease verb, so a manual lane cannot drift into the loop.
	laneNextCardRule       string
	laneOwnedCardRule      string // card id %[1]s
	laneManualDispatchRule string
}

// factoryLocales is the conversation-language table; its four entries are the
// complete set of conversation languages, falling back to English for
// anything the table does not carry (kanbanMessagesFor holds the same
// contract).
var factoryLocales = map[string]factoryMessages{
	langEnglish: {
		leaderHeader:   "Factory Mode: run %s, leader session.",
		leaderIdentity: "This session is named %s. It carries no run id: a second leader launched while this one is live takes the next free number instead, and peers address whichever name the session actually launched under. The session list shows that same name — the title is registered on your first prompt, and a later /rename still wins.",
		leaderManual: "This session dispatches cards to %d lanes over cross-session messages.\n" +
			"The lanes below are launched by hand, one per new terminal, because a session cannot launch another session.\n" +
			"Lanes are named lane-1..lane-%d, and a lane that joins with `-f lane` takes the next free lane-<n>; a number whose label is held by a live session is bumped to the next free number.",
		entryGuide: "Entry points: `moai cc -f` starts a Claude factory leader and `moai glm -f` a GLM leader — " +
			"the launcher picks the backend, `-f` the factory. `-f` with no count starts the one-lane default; " +
			"to add a lane using the same provider as this leader, run `moai %[2]s -f lane`, which takes the next free lane-<n>, " +
			"or pin a number with `moai %[2]s -f lane-<n>`.",
		agentFanout:  "Every lane can run up to 10 agents concurrently in parallel.",
		leaderSocket: "Leader socket: %s",
		leaderClasses: "Card routing: every card is routed WHOLE to one lane, and that lane carries it through " +
			"the serial 3-stage path (plan -> run -> sync, one stage completing before the next begins) in-session — " +
			"a card is never split across lanes.\n" +
			"The queue is polled and cards are picked by the operator or the kanban foreman loop (bare `/loop`); " +
			"this factory routes PICKED cards to free lanes.",
		leaderStagger: "Stagger — MUST: never activate every lane at once. Activate the first lane, wait until " +
			"it has started producing output (first job or visible progress), then activate the remaining " +
			"free-slot lanes. Concurrent requests cannot read a cache entry still being written " +
			"(cache-aware-execution directive 2). This rule governs FACTORY fan-out only — the workflow " +
			"runtime staggers itself (CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS) and is not governed here.",
		gateSummary:       "When two or more plan→run Kickoff decisions wait for the operator, present them as one batch gate summary instead of asking card by card — the format and its limits live in `.claude/rules/moai/workflow/auto-semantics.md` §9.2.",
		leaderFreeSlots:   "Free lane slots right now: %s.",
		operationalStatus: "Inspect operational lanes with factory_msg_status({\"run_id\":%q}); endpoint liveness does not establish task activity, which remains unknown without evidence.",
		leaderSlotsNone:   "none — every slot is held by a live session",
		settingsAuto:      "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify:    "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		laneJoin:          "Factory Mode: joined the leader's %[2]d-lane run as %[1]s.",
		laneJoinNoCount:   "Factory Mode: joined the leader's factory run as %[1]s.",
		laneNextCardRule: "Factory lane next-card rule: this session is a self-dispatch lane. " +
			"Leave any kept worktree as it is — never remove a card worktree. " +
			"Take the next card from the parent checkout: run `moai factory next` (MCP tool `factory_next`); " +
			"lane queue promotion through `moai factory next` is operator-authorized for the self-dispatch lane mode. " +
			"Enter the card's worktree, carry the card through plan, run, and sync, integrate it, leave its worktree kept, " +
			"and record completion with `moai factory complete` (MCP tool `factory_complete`). " +
			"Then follow the clear policy the completion output prints. " +
			"Task discipline: register the card's execution stages with `TaskCreate` before starting the first stage, " +
			"keep that list current with `TaskUpdate` at every stage transition, and record completion only while " +
			"the list reflects the end state — or state explicitly why it does not. " +
			"Tools and their CLI equivalents: `todo_add` (`moai todo add`), `todo_list` (`moai todo`), " +
			"`factory_next` (`moai factory next`), `factory_stage` (`moai factory stage`), " +
			"`factory_complete` (`moai factory complete`), `factory_decide` (`moai factory decide`).",
		laneOwnedCardRule: "Factory lane owned-card rule: this session owns card %[1]s — the id recorded in MOAI_KANBAN_CARD — " +
			"and works in the current worktree. Carry that card to merge-ready: register the card's stages with `TaskCreate` at intake, " +
			"keep the list current with `TaskUpdate` at each stage transition, record each stage with `moai factory stage`, " +
			"and finish with `moai factory complete` only while the task list reflects the end state — or state explicitly why it does not. " +
			"Then end this session; do not take or lease any other card.",
		laneManualDispatchRule: "Factory lane manual-dispatch rule: this session launched with --no-auto-dispatch — " +
			"manual dispatch mode. Do not lease cards yourself: the operator (or the factory leader) routes each card to you explicitly. " +
			"When a card is routed to you, work it in the current worktree and record its stages with `moai factory stage`.",
	},
	"ko": {
		leaderHeader:   "팩토리 모드: run %s, 리더 세션.",
		leaderIdentity: "이 세션의 이름은 %s 입니다. 이름에 run id 는 들어가지 않습니다 — 이 세션이 살아 있는 동안 리더를 하나 더 띄우면 그쪽이 다음 번호를 받고, 다른 세션은 실제로 띄워진 이름으로 이 세션을 부릅니다. 세션 목록에도 같은 이름이 뜹니다 — 제목은 첫 프롬프트에서 등록되고, 나중에 /rename 을 하면 그쪽이 우선합니다.",
		leaderManual: "이 세션이 세션 간 메시지로 카드를 레인 %d개에 배분합니다.\n" +
			"아래 레인은 터미널을 하나씩 새로 열어 직접 실행하세요 — 세션은 다른 세션을 띄울 수 없습니다.\n" +
			"레인 이름은 lane-1..lane-%d 이고, `-f lane` 로 합류한 레인은 비어 있는 다음 lane-<n> 이름을 받습니다. 생존 세션이 이미 쓰고 있는 번호는 다음 빈 번호로 늘어납니다.",
		entryGuide: "진입점: `moai cc -f`는 Claude, `moai glm -f`는 GLM을 사용하는 팩토리 리더를 시작합니다. " +
			"런처가 백엔드를, `-f` 가 팩토리를 정합니다. `-f` 에 개수를 붙이지 않으면 레인 1개 기본으로 시작하고, " +
			"현재 리더와 같은 공급자로 레인을 추가할 때는 `moai %[2]s -f lane`를 쓰면 비어 있는 다음 lane-<n> 번호를 알아서 잡습니다. " +
			"번호를 직접 정하려면 `moai %[2]s -f lane-<n>`을 쓰세요.",
		agentFanout:  "각 레인은 최대 10개의 에이전트를 동시에 병렬로 실행할 수 있습니다.",
		leaderSocket: "리더 소켓: %s",
		leaderClasses: "카드 라우팅: 모든 카드는 한 레인에 통째로 배정되고, 그 레인이 세션 안에서 직렬 3단계 경로" +
			"(plan -> run -> sync, 한 단계가 끝나야 다음 단계)를 끝까지 수행합니다 — 카드를 여러 레인에 쪼개 배분하지 않습니다.\n" +
			"큐 폴링과 카드 선택은 운영자 또는 칸반 포어맨 루프(단독 `/loop`)가 담당합니다. 이 팩토리는 선택된(picked) 카드를 빈 레인으로 라우팅합니다.",
		leaderStagger: "순차 기동 — 필수: 절대 모든 레인을 동시에 활성화하지 마세요. 첫 번째 레인을 먼저 활성화하고, " +
			"실제 출력을 내기 시작했다는 증거(첫 작업 수행 또는 진행 흔적)를 확인한 뒤 나머지 빈 슬롯의 레인을 활성화하세요. " +
			"동시 요청은 아직 기록 중인 캐시 항목을 읽을 수 없습니다(cache-aware-execution directive 2). " +
			"이 규칙은 팩토리 팬아웃에만 적용됩니다 — 워크플로 런타임은 스스로 스태거합니다(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS).",
		gateSummary:       "plan→run Kickoff 결정이 둘 이상 운영자의 답을 기다리면 카드마다 따로 묻지 말고 batch gate summary 하나로 제시하세요 — 서식과 한계는 `.claude/rules/moai/workflow/auto-semantics.md` §9.2 에 있습니다.",
		leaderFreeSlots:   "현재 빈 레인 슬롯: %s.",
		operationalStatus: "factory_msg_status({\"run_id\":%q})로 운영 레인 상태를 조회하세요. 프로세스 생존 여부만으로 작업 중이라고 판단하지 않으며, 작업 관측이 없으면 unknown입니다.",
		leaderSlotsNone:   "없음 — 모든 슬롯을 생존 세션이 사용 중입니다",
		settingsAuto:      "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify:    "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		laneJoin:          "팩토리 모드: 리더의 레인 %[2]d개 런에 %[1]s 로 합류했습니다.",
		laneJoinNoCount:   "팩토리 모드: 리더의 팩토리 run 에 %[1]s 로 합류했습니다.",
		laneNextCardRule: "팩토리 레인 다음 카드 규칙: 이 세션은 셀프 디스패치 레인입니다. " +
			"kept 로 남은 워크트리는 그대로 둡니다 — 카드 워크트리를 절대 제거하지 않습니다. " +
			"부모 체크아웃에서 다음 카드를 가져옵니다. `moai factory next`(MCP 도구 `factory_next`)를 실행하세요. " +
			"셀프 디스패치 레인 모드에서 `moai factory next` 를 통한 레인 큐 승격은 운영자가 승인했습니다. " +
			"카드의 워크트리에 진입해 plan, run, sync 로 카드를 끝까지 수행하고, 통합한 뒤에도 워크트리는 kept 로 남기고, " +
			"`moai factory complete`(MCP 도구 `factory_complete`)로 완료를 기록합니다. " +
			"그다음에는 완료 출력이 알려 주는 clear 정책을 따릅니다. " +
			"태스크 규율: 첫 단계를 시작하기 전에 카드의 실행 단계를 `TaskCreate` 로 등록하고, 단계가 바뀔 때마다 `TaskUpdate` 로 목록을 현재 상태로 유지하며, " +
			"목록이 종결 상태를 반영할 때만 완료를 기록합니다 — 반영하지 않는다면 그 이유를 명시합니다. " +
			"도구와 CLI 등가물: `todo_add`(`moai todo add`), `todo_list`(`moai todo`), `factory_next`(`moai factory next`), " +
			"`factory_stage`(`moai factory stage`), `factory_complete`(`moai factory complete`), `factory_decide`(`moai factory decide`).",
		laneOwnedCardRule: "팩토리 레인 담당 카드 규칙: 이 세션은 카드 %[1]s — MOAI_KANBAN_CARD 에 기록된 id — 를 담당하며 현재 워크트리에서 작업합니다. " +
			"그 카드를 merge-ready 까지 진행하세요. 카드 인계 시 실행 단계를 `TaskCreate` 로 등록하고, 단계가 바뀔 때마다 `TaskUpdate` 로 목록을 현재 상태로 유지하며, " +
			"각 단계를 `moai factory stage` 로 기록하고, 태스크 목록이 종결 상태를 반영할 때만 `moai factory complete` 로 마칩니다 — 반영하지 않는다면 그 이유를 명시합니다. " +
			"그런 다음 이 세션을 종료합니다. 다른 카드를 가져오거나 임대하지 않습니다.",
		laneManualDispatchRule: "팩토리 레인 수동 디스패치 규칙: 이 세션은 --no-auto-dispatch 로 시작했습니다 — 수동 디스패치 모드입니다. " +
			"카드를 임대하려고 `moai factory next` 를 실행하지 마세요: 운영자(또는 팩토리 리더)가 카드를 명시적으로 배분합니다. " +
			"배분받은 카드는 현재 워크트리에서 작업하고 각 단계를 `moai factory stage` 로 기록합니다.",
	},
	"ja": {
		leaderHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leaderIdentity: "このセッションの名前は %s です。名前に run id は含まれません — このセッションが生きている間にもう一つリーダーを起動すると、そちらが次の番号を取り、他のセッションは実際に起動した名前でこのセッションを呼びます。セッション一覧にも同じ名前が表示されます — タイトルは最初のプロンプトで登録され、後から /rename すればそちらが優先されます。",
		leaderManual: "このセッションが、セッション間メッセージでカードをレーン %d 本に割り振ります。\n" +
			"以下のレーンは、ターミナルを 1 つずつ新規に開いて手動で起動してください — セッションが別のセッションを起動することはできません。\n" +
			"レーン名は lane-1..lane-%d で、`-f lane` で参加したレーンには空いている次の lane-<n> が割り当てられます。生存セッションが保持する番号は次の空き番号へ繰り上がります。",
		entryGuide: "起動コマンド: `moai cc -f` は Claude、`moai glm -f` は GLM のファクトリーリーダーを起動します。" +
			"ランチャーがバックエンドを、`-f` がファクトリーを決めます。`-f` に個数を付けなければレーン1本のデフォルトで開始し、 " +
			"現在のリーダーと同じプロバイダーでレーンを追加するには `moai %[2]s -f lane` を使ってください。空いている次の lane-<n> が自動で割り当てられます。" +
			"番号を指定したい場合は `moai %[2]s -f lane-<n>` を使います。",
		agentFanout:  "各レーンは最大 10 個のエージェントを同時に並列実行できます。",
		leaderSocket: "リーダーソケット: %s",
		leaderClasses: "カードルーティング: すべてのカードは1つのレーンへ丸ごと割り当てられ、そのレーンがセッション内で直列3段階パス" +
			"(plan -> run -> sync、各段階の完了後に次へ)を最後まで実行します — カードを複数レーンに分割することはありません。\n" +
			"キューのポーリングとカードの選択はオペレーターまたはかんばんフォアマンループ(単独の `/loop`)が担います。このファクトリーは選択済み(picked)カードを空きレーンへ振り分けます。",
		leaderStagger: "段階的起動 — 必須: すべてのレーンを同時に起動してはいけません。最初のレーンを起動し、実際に出力を生成し始めた証拠 " +
			"(最初のジョブまたは進行の形跡)を確認してから、残りの空きスロットのレーンを起動してください。 " +
			"同時リクエストは書き込み中のキャッシュエントリを読めません(cache-aware-execution directive 2)。 " +
			"このルールはファクトリーファンアウトにのみ適用されます — ワークフローランタイムは自身でスタガーします(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		gateSummary:       "plan→run Kickoff の判断が 2 件以上オペレーターの回答を待っているときは、カードごとに個別に尋ねず batch gate summary 1 つにまとめて提示してください — 書式と制約は `.claude/rules/moai/workflow/auto-semantics.md` §9.2 にあります。",
		leaderFreeSlots:   "現在の空きレーンスロット: %s。",
		operationalStatus: "factory_msg_status({\"run_id\":%q}) でレーンの稼働状態を確認してください。プロセスの生存は作業中である証拠ではなく、作業の観測がなければ unknown です。",
		leaderSlotsNone:   "なし — すべてのスロットを生存セッションが保持しています",
		settingsAuto:      "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify:    "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		laneJoin:          "ファクトリーモード: リーダーのレーン %[2]d 本の run に %[1]s として参加しました。",
		laneJoinNoCount:   "ファクトリーモード: リーダーのファクトリー run に %[1]s として参加しました。",
		laneNextCardRule: "ファクトリーレーン次カード規則: このセッションはセルフディスパッチレーンです。 " +
			"kept のワークツリーはそのまま残します — カードのワークツリーを削除してはいけません。 " +
			"親チェックアウトから次のカードを取得します。`moai factory next`(MCP ツール `factory_next`)を実行してください。 " +
			"セルフディスパッチレーンモードでは、`moai factory next` によるレーンキューの昇格はオペレーターが承認済みです。 " +
			"カードのワークツリーに入り、plan・run・sync を通してカードを完遂し、統合したうえでワークツリーを kept のまま残し、 " +
			"`moai factory complete`(MCP ツール `factory_complete`)で完了を記録します。 " +
			"その後は、完了出力が示す clear ポリシーに従います。 " +
			"タスク規律: 最初の段階を始める前にカードの実行段階を `TaskCreate` で登録し、段階が移るたびに `TaskUpdate` で一覧を現在の状態に保ち、 " +
			"一覧が終了状態を反映しているときにだけ完了を記録します — 反映していない場合はその理由を明示します。 " +
			"ツールと CLI の対応: `todo_add`(`moai todo add`)、`todo_list`(`moai todo`)、`factory_next`(`moai factory next`)、 " +
			"`factory_stage`(`moai factory stage`)、`factory_complete`(`moai factory complete`)、`factory_decide`(`moai factory decide`)。",
		laneOwnedCardRule: "ファクトリーレーン担当カード規則: このセッションはカード %[1]s — MOAI_KANBAN_CARD に記録された id — を担当し、現在のワークツリーで作業します。 " +
			"そのカードを merge-ready まで進めます。カードの引き受け時に実行段階を `TaskCreate` で登録し、段階が移るたびに `TaskUpdate` で一覧を現在の状態に保ち、 " +
			"各段階を `moai factory stage` で記録し、タスク一覧が終了状態を反映しているときにだけ `moai factory complete` で締めます — 反映していない場合はその理由を明示します。 " +
			"その後、このセッションを終了してください。他のカードを取得したりリースしたりしません。",
		laneManualDispatchRule: "ファクトリーレーン手動ディスパッチ規則：このセッションは --no-auto-dispatch で起動しました — 手動ディスパッチモードです。 " +
			"カードをリースするために `moai factory next` を実行しないでください：オペレーター（またはファクトリーリーダー）がカードを明示的に割り当てます。 " +
			"割り当てられたカードは現在のワークツリーで作業し、各段階を `moai factory stage` で記録します。",
	},
	"zh": {
		leaderHeader:   "工厂模式：run %s，主导会话。",
		leaderIdentity: "本会话的名称是 %s。名称中不含 run id —— 本会话存活期间再启动一个主导会话，后者会取下一个编号；其他会话按实际启动时的名称来称呼本会话。会话列表中也显示同一名称 —— 标题在首次提示时注册，之后 /rename 优先。",
		leaderManual: "本会话通过跨会话消息把卡片分发给 %d 条泳道。\n" +
			"下面的泳道需要各自新开一个终端手动启动 —— 会话无法启动另一个会话。\n" +
			"泳道命名为 lane-1..lane-%d，通过 `-f lane` 加入的泳道会占用下一个空闲的 lane-<n>；已被存活会话占用的编号会顺延到下一个空位。",
		entryGuide: "启动命令：`moai cc -f` 使用 Claude，`moai glm -f` 使用 GLM，分别启动工厂主导会话。" +
			"启动器决定后端，`-f` 决定工厂。`-f` 不带数量时以一条泳道的默认配置启动；" +
			"如需使用与当前主导会话相同的提供商添加泳道，请运行 `moai %[2]s -f lane`，它会自动占用下一个空闲的 lane-<n>；" +
			"如需指定编号，请运行 `moai %[2]s -f lane-<n>`。",
		agentFanout:  "每条泳道最多可同时并行运行 10 个代理。",
		leaderSocket: "主导会话套接字：%s",
		leaderClasses: "卡片路由：每张卡片整体分发给一条泳道，由该泳道在会话内走完串行三阶段路径" +
			"(plan -> run -> sync，上一阶段完成后才进入下一阶段)—— 绝不把卡片拆分到多条泳道。\n" +
			"队列轮询与卡片挑选由操作者或看板工头循环(单独的 `/loop`)负责。本工厂把已挑选(picked)的卡片路由到空闲泳道。",
		leaderStagger: "分批启动 — 必须: 绝不要同时激活所有泳道。先激活第一条泳道，等到它确实开始产出 " +
			"(首个任务或可见进展)之后，再激活其余空闲槽位的泳道。 并发请求无法读取仍在写入的缓存条目 " +
			"(cache-aware-execution directive 2)。本规则仅约束工厂分发 — 工作流运行时会自行错峰(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		gateSummary:       "当有两个及以上 plan→run Kickoff 决定在等待操作者答复时，不要逐张卡片提问，请合并为一份 batch gate summary 提交 —— 格式与限制见 `.claude/rules/moai/workflow/auto-semantics.md` §9.2。",
		leaderFreeSlots:   "当前空闲泳道：%s。",
		operationalStatus: "使用 factory_msg_status({\"run_id\":%q}) 查询运行通道状态。进程存活不代表正在执行任务；没有任务观测依据时，状态为 unknown。",
		leaderSlotsNone:   "无 — 所有槽位均被存活会话占用",
		settingsAuto:      "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify:    "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		laneJoin:          "工厂模式：已以 %[1]s 身份加入主导会话的 %[2]d 条泳道的 run。",
		laneJoinNoCount:   "工厂模式：已以 %[1]s 身份加入主导会话的工厂 run。",
		laneNextCardRule: "工厂泳道下一张卡规则：本会话是自调度泳道。 " +
			"保持 kept 状态的工作树原样保留 — 绝不删除卡片的工作树。 " +
			"从父检出获取下一张卡：运行 `moai factory next`（MCP 工具 `factory_next`）。 " +
			"在自调度泳道模式下，通过 `moai factory next` 进行的泳道队列提升已获运营者授权。 " +
			"进入卡片的工作树，带着卡片走完 plan、run、sync，完成集成后工作树保持 kept， " +
			"并用 `moai factory complete`（MCP 工具 `factory_complete`）记录完成。 " +
			"之后遵循完成输出给出的 clear 策略。 " +
			"任务纪律：在开始第一阶段前用 `TaskCreate` 登记卡片的执行阶段，每次阶段切换时用 `TaskUpdate` 保持列表反映当前状态； " +
			"仅当列表已反映结束状态时才记录完成 — 否则明确说明原因。 " +
			"工具及其 CLI 等价物：`todo_add`（`moai todo add`）、`todo_list`（`moai todo`）、`factory_next`（`moai factory next`）、 " +
			"`factory_stage`（`moai factory stage`）、`factory_complete`（`moai factory complete`）、`factory_decide`（`moai factory decide`）。",
		laneOwnedCardRule: "工厂泳道负责卡规则：本会话负责卡片 %[1]s — MOAI_KANBAN_CARD 中记录的 id — 并在当前工作树中工作。 " +
			"把该卡推进到 merge-ready：接手卡片时用 `TaskCreate` 登记执行阶段，每次阶段切换时用 `TaskUpdate` 保持列表反映当前状态， " +
			"用 `moai factory stage` 记录各阶段，仅当任务列表已反映结束状态时才用 `moai factory complete` 收尾 — 否则明确说明原因。 " +
			"然后结束本会话；不要领取或租用其他卡片。",
		laneManualDispatchRule: "工厂泳道手动调度规则：本会话以 --no-auto-dispatch 启动 — 手动调度模式。 " +
			"不要运行 `moai factory next` 去租用卡片：由操作者（或工厂主导会话）显式分配每张卡片。 " +
			"被分配的卡片在当前工作树中处理，各阶段用 `moai factory stage` 记录。",
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
