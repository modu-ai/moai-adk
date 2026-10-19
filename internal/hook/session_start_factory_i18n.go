package hook

// session_start_factory_i18n.go holds the operator-facing prose of the
// Factory Mode bootstrap notice, one message set per locale.
//
// Two invariants govern it. Only the prose lives here: the launch
// commands, the run id, the socket path, and the lane labels are protocol
// tokens the operator copies verbatim, so they stay in the builder and never
// enter this table. And no field carries a leading or trailing newline: the
// builder joins lines within a block and blank-separates the blocks, so the
// notice is laid out identically in every locale by construction.
//
// The settingsAuto / settingsVerify / leaderSocket fields describe the
// injected-settings mechanism and the socket surface; they live in this table
// so the notice is self-contained and free to drift when the mechanism does.

// factoryMessages is the operator-facing prose of one locale.
//
// Fields carrying %s / %d are format strings. leaderManual and entryGuide are
// plain text: the leader notice states how to start a lane and carries no
// number (SPEC-LAUNCHER-ENTRY-FLAGS-001 REQ-009). laneJoin pins its argument
// order with explicit %[1]s / %[2]d indices because the locales' natural word
// orders differ (en/ja/ko say the count first, zh the label first) — see
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
	leaderManual      string // the one lane-start sentence: the three -l entries
	entryGuide        string // the -f leader entries
	agentFanout       string // per-lane concurrent agent cap
	leaderSocket      string // socket path
	leaderClasses     string // whole-card routing: one lane runs the serial 3-stage path in-session
	leaderStagger     string // fan-out-only staggered activation
	gateSummary       string // one sentence: name token + pointer
	operationalStatus string // explicit run-bound read-only operational query
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

	// laneRecheckRule is the stall-recovery rule a Claude-harness lane
	// carries after its next-card or manual-dispatch rule (card t1451): the
	// standing recheck cron that wakes a lane stopped on an API error or on a
	// delegate's report that never comes, and the disk-evidence-first reading
	// of any wake. The tool names and the watchdog skill name stay verbatim
	// in every locale; the cron expression is the %[1]s argument, the arm
	// expression rendered from the effective renewal period
	// (workflow.factory_lease.renew_interval, SPEC-FACTORY-LEASE-RENEW-001 REQ-FLR-012). The Codex-harness owned-card
	// rule does not carry it — that harness has no session cron tool.
	laneRecheckRule string

	// laneSpawnAuthority is the standing spawn authority every lane join
	// notice carries (card t224). It moved here from the English-only const
	// in lane_spawn_authority.go (card t1603): the three design decisions
	// the sentence encodes — the delegation to the Status Transition
	// Ownership Matrix pointer, the depth-1 leaf seal, and the standing
	// bootstrap placement no peer message can grant or revoke — are
	// preserved semantically in every locale, and the matrix pointer path
	// stays verbatim as a protocol token. The en entry is the canonical
	// sentence, unchanged (SPEC-SESSION-START-GUIDE-I18N-001 REQ-002).
	laneSpawnAuthority string

	// autoModeGuideLeader / autoModeGuideLane explain the
	// `/moai todo --auto-leader` / `--auto-lane` two-mode split (guidance card
	// t1603; the flags shipped with card t1814). REQ-005 truthfulness: every
	// variant carries the locale-independent literal "shipped — t1814"
	// verbatim in the same block; it is a protocol token under the
	// no-translation rule, and a bare card number without the full literal
	// fails the AC-005 test. The
	// leader variant leads with the leader mode, the lane variant with the
	// lane mode; each names the other mode in passing (REQ-004).
	autoModeGuideLeader string
	autoModeGuideLane   string

	// docsPointer is the online-documentation sentence (REQ-003):
	// https://adk.mo.ai.kr is the one whitelisted domain and stays verbatim
	// in every locale; only the surrounding prose localizes. Both the
	// leader and the lane notice carry it.
	docsPointer string
}

// factoryLocales is the conversation-language table; its four entries are the
// complete set of conversation languages, falling back to English for
// anything the table does not carry (factoryMessagesFor holds the contract).
var factoryLocales = map[string]factoryMessages{
	langEnglish: {
		leaderHeader:   "Factory Mode: run %s, leader session.",
		leaderIdentity: "This session is named %s. It carries no run id: a second leader launched while this one is live takes the next free number instead, and peers address whichever name the session actually launched under. The session list shows that same name — the title is registered on your first prompt, and a later /rename still wins.",
		leaderManual: "This session dispatches cards to lanes over cross-session messages.\n" +
			"A session cannot launch another session, so to start a lane, enter `moai cc -l` (or `moai glm -l`, `moai codex -l`) in a new terminal.",
		entryGuide: "Entry points: `moai cc -f` starts a Claude factory leader and `moai glm -f` a GLM leader — " +
			"the launcher picks the backend, `-f` the factory.",
		agentFanout:  "Every lane can run up to 10 agents concurrently in parallel.",
		leaderSocket: "Leader socket: %s",
		leaderClasses: "Card routing: every card is routed WHOLE to one lane, and that lane carries it through " +
			"the serial 3-stage path (plan -> run -> sync, one stage completing before the next begins) in-session — " +
			"a card is never split across lanes.\n" +
			"The queue is polled and cards are picked by the operator or the factory foreman loop (bare `/loop`); " +
			"this factory routes PICKED cards to free lanes.",
		leaderStagger: "Stagger — MUST: never activate every lane at once. Activate the first lane, wait until " +
			"it has started producing output (first job or visible progress), then activate the remaining " +
			"free-slot lanes. Concurrent requests cannot read a cache entry still being written " +
			"(cache-aware-execution directive 2). This rule governs FACTORY fan-out only — the workflow " +
			"runtime staggers itself (CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS) and is not governed here.",
		gateSummary:       "When two or more plan→run Kickoff decisions wait for the operator, present them as one batch gate summary instead of asking card by card — the format and its limits live in `.claude/rules/moai/workflow/auto-semantics.md` §9.2.",
		operationalStatus: "Inspect operational lanes with factory_msg_status({\"run_id\":%q}); endpoint liveness does not establish task activity, which remains unknown without evidence.",
		settingsAuto:      "Cross-session messages are auto-accepted via the injected --settings.",
		settingsVerify:    "Verify \"crossSessionInbound\": \"accept\" is present in your --settings file so cross-session messages are accepted.",
		laneJoin:          "Factory Mode: joined the leader's %[2]d-lane run as %[1]s.",
		laneJoinNoCount:   "Factory Mode: joined the leader's factory run as %[1]s.",
		laneNextCardRule:  "Factory lane: self-dispatch only. Lease via operator-authorized `moai factory next` from primary; use `--card <id>` for your assignment with siblings pending (t1542). Never remove a card worktree. Plan→run→sync in that tree; integrate, verify locally, run `codex_review` scope=card, then `moai factory complete` (push, PR, observe merge). Trusted native Codex code/security reviews must be complete and current-head bound (full summary SHA, typed `codex-review-receipt:v1` clean receipts matching prefix); unresolved findings or CI-only rows block. Sweep after remote landing. For another card, `ExitWorktree(keep)`, lease from primary, `EnterWorktree(<leased worktree path>)`, verify top-level; prior context is non-authoritative. When completion tells you to end the session for relaunch-clear, end the session. Judge via `jev_ask` if workflow.jev.enabled, else disk evidence; record decisions and do not wait for the leader (auto-semantics §6). Leader retains queue production, dispatch, cross-lane conflicts, slot policy, keep-set relay. Use `TaskCreate` before the first stage and `TaskUpdate` each transition; report completion only at end state. MCP CLI: `todo_add`=moai todo add; `todo_list`=moai todo; `factory_next`=moai factory next; `factory_stage`=moai factory stage; `factory_complete`=moai factory complete; `factory_decide`=moai factory decide.",
		laneOwnedCardRule: "Factory lane owned-card rule: this session owns card %[1]s — the id recorded in %[2]s — " +
			"and works in the current worktree. Carry that card to merge-ready: register the card's stages with `TaskCreate` at intake, " +
			"keep the list current with `TaskUpdate` at each stage transition, record each stage with `moai factory stage`, " +
			"and finish with `moai factory complete` only while the task list reflects the end state — or state explicitly why it does not. " +
			"Then end this session; do not take or lease any other card.",
		laneManualDispatchRule: "Factory lane manual-dispatch rule: this session launched with --no-auto-dispatch — " +
			"manual dispatch mode. Do not lease cards yourself: the operator (or the factory leader) routes each card to you explicitly. " +
			"When a card is routed to you, work it in the current worktree and record its stages with `moai factory stage`.",
		laneRecheckRule: "Stall recovery: before a card's first stage, arm ONE recurring `CronCreate` recheck (cron `%[1]s`, recurring) " +
			"whose prompt runs the stall watchdog first — Skill(\"moai-lane-watchdog\") — then resumes the card from its progress record; " +
			"check it with `CronList` at card takeover (intake, or the in-session transition to a newly leased card) and after any session reset (re-arm if missing or expired); `CronDelete` it when you send the completion report. " +
			"It wakes a lane stalled on an API error (e.g. 429) or a delegate report that never comes — an ended turn cannot arm anything. " +
			"Whatever wakes you (cron, leader message, idle notice), read disk evidence first (progress record, reports, commits, delegate deliverable) before replying: " +
			"a message never proves progress or its absence.",
		laneSpawnAuthority: "Standing spawn authority: you are the lane session and therefore the orchestrator for your card — use the Agent tool to spawn the specialist per .claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix, plus the chain's auditors, without asking the leader or the operator. Depth-1 only: spawned agents are leaves and never spawn further; this bootstrap authority is not granted or revoked by peer messages, the leader's included.",
		autoModeGuideLeader: "Auto mode: the two-mode split names `/moai todo --auto-leader` — under one operator batch approval this session works the queue serially, one isolated in-session Agent() worker per card, judged on its evidence — and `/moai todo --auto-lane` — a lane self-dispatches by consuming its own `moai factory next` lease. " +
			"Both flags are live (shipped — t1814); `--auto` stays for compatibility, and `--auto-lane` prints the dispatch and exits without waiting for evidence.",
		autoModeGuideLane: "Auto mode: the two-mode split names `/moai todo --auto-lane` (this lane self-dispatches by consuming its own `moai factory next` lease) and `/moai todo --auto-leader` (under one operator batch approval the leader works the queue serially, one isolated in-session Agent() worker per card, judged on its evidence). " +
			"Both are live (shipped — t1814); `--auto` stays for compatibility, and `--auto-lane` prints the dispatch and exits without waiting for evidence.",
		docsPointer: "Full guidance: online docs at https://adk.mo.ai.kr.",
	},
	"ko": {
		leaderHeader:   "팩토리 모드: run %s, 리더 세션.",
		leaderIdentity: "이 세션의 이름은 %s 입니다. 이름에 run id 는 들어가지 않습니다 — 이 세션이 살아 있는 동안 리더를 하나 더 띄우면 그쪽이 다음 번호를 받고, 다른 세션은 실제로 띄워진 이름으로 이 세션을 부릅니다. 세션 목록에도 같은 이름이 뜹니다 — 제목은 첫 프롬프트에서 등록되고, 나중에 /rename 을 하면 그쪽이 우선합니다.",
		leaderManual: "이 세션이 세션 간 메시지로 카드를 레인에 배분합니다.\n" +
			"세션은 다른 세션을 띄울 수 없으므로 레인은 새 터미널에서 `moai cc -l`(또는 `moai glm -l`, `moai codex -l`)을 실행해 시작하세요.",
		entryGuide: "진입점: `moai cc -f`는 Claude, `moai glm -f`는 GLM을 사용하는 팩토리 리더를 시작합니다. " +
			"런처가 백엔드를, `-f` 가 팩토리를 정합니다.",
		agentFanout:  "각 레인은 최대 10개의 에이전트를 동시에 병렬로 실행할 수 있습니다.",
		leaderSocket: "리더 소켓: %s",
		leaderClasses: "카드 라우팅: 모든 카드는 한 레인에 통째로 배정되고, 그 레인이 세션 안에서 직렬 3단계 경로" +
			"(plan -> run -> sync, 한 단계가 끝나야 다음 단계)를 끝까지 수행합니다 — 카드를 여러 레인에 쪼개 배분하지 않습니다.\n" +
			"큐 폴링과 카드 선택은 운영자 또는 팩토리 포어맨 루프(단독 `/loop`)가 담당합니다. 이 팩토리는 선택된(picked) 카드를 빈 레인으로 라우팅합니다.",
		leaderStagger: "순차 기동 — 필수: 절대 모든 레인을 동시에 활성화하지 마세요. 첫 번째 레인을 먼저 활성화하고, " +
			"실제 출력을 내기 시작했다는 증거(첫 작업 수행 또는 진행 흔적)를 확인한 뒤 나머지 빈 슬롯의 레인을 활성화하세요. " +
			"동시 요청은 아직 기록 중인 캐시 항목을 읽을 수 없습니다(cache-aware-execution directive 2). " +
			"이 규칙은 팩토리 팬아웃에만 적용됩니다 — 워크플로 런타임은 스스로 스태거합니다(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS).",
		gateSummary:       "plan→run Kickoff 결정이 둘 이상 운영자의 답을 기다리면 카드마다 따로 묻지 말고 batch gate summary 하나로 제시하세요 — 서식과 한계는 `.claude/rules/moai/workflow/auto-semantics.md` §9.2 에 있습니다.",
		operationalStatus: "factory_msg_status({\"run_id\":%q})로 운영 레인 상태를 조회하세요. 프로세스 생존 여부만으로 작업 중이라고 판단하지 않으며, 작업 관측이 없으면 unknown입니다.",
		settingsAuto:      "세션 간 메시지는 주입된 --settings 로 자동 수락됩니다.",
		settingsVerify:    "--settings 파일에 \"crossSessionInbound\": \"accept\" 가 있는지 확인하세요. 세션 간 메시지 수락에 필요합니다.",
		laneJoin:          "팩토리 모드: 리더의 레인 %[2]d개 런에 %[1]s 로 합류했습니다.",
		laneJoinNoCount:   "팩토리 모드: 리더의 팩토리 run 에 %[1]s 로 합류했습니다.",
		laneNextCardRule: "팩토리 레인 다음 카드 규칙: 이 세션은 셀프 디스패치 레인입니다. " +
			"kept 로 남은 워크트리는 그대로 둡니다 — 카드 워크트리를 절대 제거하지 않습니다. " +
			"부모 체크아웃에서 다음 카드를 가져옵니다. `moai factory next`(MCP 도구 `factory_next`)를 실행하세요. " +
			"셀프 디스패치 레인 모드에서 `moai factory next` 를 통한 레인 큐 승격은 운영자가 승인했고, " +
			"자기에게 배정된 카드는 형제 배정이 남아 있어도 `moai factory next --card <id>` 로 임대합니다(카드 t1542). " +
			"카드의 워크트리에 진입해 plan, run, sync 로 카드를 끝까지 수행하고, 통합한 뒤에도 워크트리는 kept 로 남기고, " +
			"`moai factory complete`(MCP 도구 `factory_complete`)로 완료를 기록합니다. " +
			"그다음에는 세션 안에서 카드를 전환합니다: `ExitWorktree(keep)` 로 워크트리를 유지한 채 나오고, 부모 체크아웃에서 `moai factory next` 로 다음 카드를 받고, `EnterWorktree(<임대된 워크트리 경로>)` 로 들어가 toplevel 을 확인하며, 이전 카드의 대화 맥락은 근거로 삼지 않습니다. 단 완료 출력이 세션을 종료하라고 알려 주면(relaunch 정책) 전환하지 말고 세션을 종료합니다. " +
			"판단은 레인의 몫입니다(카드 t1542) — workflow.jev.enabled 가 참이면 `jev_ask` 를 직접 부르고, 아니면 디스크 증거로 스스로 판단해 " +
			"카드 진행 기록에 결정 기록을 남깁니다. 리더에게 묻고 기다리지 않습니다. 사다리는 .claude/rules/moai/workflow/auto-semantics.md §6 에 있습니다. " +
			"착지도 레인이 실행합니다(auto-semantics.md §13, mechanics § The lane's standard landing): " +
			"레인 로컬 검증 → `codex_review` scope=card 카드 리뷰 → `moai factory complete`(push·PR·머지 관측) → " +
			"native Codex 리뷰 증거를 확인합니다. 최신 상태 요약과 `codex-review-receipt:v1`이 현재 head를 가리켜야 합니다(요약은 full SHA, receipt는 일치하는 SHA prefix). 코드·보안 리뷰가 모두 완료되고 미해결 finding이 없어야 하며 CI 상태 행만으로 통과 처리하지 않습니다. 원격 착지 확인 후 `moai worktree sweep`을 실행합니다. " +
			"merged-pr 가 자기 큐 카드를 닫습니다. 리더는 큐 생산·배차·레인 간 충돌·슬롯 정책·keep-set 게이트 전달만 남습니다. " +
			"태스크 규율: 첫 단계를 시작하기 전에 카드의 실행 단계를 `TaskCreate` 로 등록하고, 단계가 바뀔 때마다 `TaskUpdate` 로 목록을 현재 상태로 유지하며, " +
			"목록이 종결 상태를 반영할 때만 완료를 기록합니다 — 반영하지 않는다면 그 이유를 명시합니다. " +
			"도구와 CLI 등가물: `todo_add`(`moai todo add`), `todo_list`(`moai todo`), `factory_next`(`moai factory next`), " +
			"`factory_stage`(`moai factory stage`), `factory_complete`(`moai factory complete`), `factory_decide`(`moai factory decide`).",
		laneOwnedCardRule: "팩토리 레인 담당 카드 규칙: 이 세션은 카드 %[1]s — %[2]s 에 기록된 id — 를 담당하며 현재 워크트리에서 작업합니다. " +
			"그 카드를 merge-ready 까지 진행하세요. 카드 인계 시 실행 단계를 `TaskCreate` 로 등록하고, 단계가 바뀔 때마다 `TaskUpdate` 로 목록을 현재 상태로 유지하며, " +
			"각 단계를 `moai factory stage` 로 기록하고, 태스크 목록이 종결 상태를 반영할 때만 `moai factory complete` 로 마칩니다 — 반영하지 않는다면 그 이유를 명시합니다. " +
			"그런 다음 이 세션을 종료합니다. 다른 카드를 가져오거나 임대하지 않습니다.",
		laneManualDispatchRule: "팩토리 레인 수동 디스패치 규칙: 이 세션은 --no-auto-dispatch 로 시작했습니다 — 수동 디스패치 모드입니다. " +
			"카드를 임대하려고 `moai factory next` 를 실행하지 마세요: 운영자(또는 팩토리 리더)가 카드를 명시적으로 배분합니다. " +
			"배분받은 카드는 현재 워크트리에서 작업하고 각 단계를 `moai factory stage` 로 기록합니다.",
		laneRecheckRule: "정지 복구: 카드의 첫 단계를 시작하기 전에 `CronCreate` 로 반복 재점검 크론 하나를 겁니다(cron `%[1]s`, recurring). " +
			"프롬프트는 먼저 레인 스톨 워치독을 실행하고 — Skill(\"moai-lane-watchdog\") 를 호출해 따릅니다 — 그다음 카드의 진행 기록에서 작업을 재개하도록 씁니다. " +
			"카드를 넘겨받을 때(세션 안 카드 전환 포함)와 세션이 초기화된 뒤에 `CronList` 로 확인하고, 없거나 만료됐으면 다시 걸고, 완료 보고를 보낼 때 `CronDelete` 로 지웁니다. " +
			"API 오류(429 등)로 멈췄거나 끝내 오지 않을 위임 대상의 보고를 기다리다 멈춘 레인은 이 크론이 깨웁니다 — 이미 끝난 턴은 아무것도 걸 수 없기 때문입니다. " +
			"크론이든 리더 메시지든 idle 알림이든, 깨어나면 답하기 전에 디스크 증거(진행 기록·보고서·커밋·위임 대상의 산출물)를 먼저 읽으세요. " +
			"메시지는 진행이 있다는 증거도, 없다는 증거도 아닙니다.",
		laneSpawnAuthority: "상시 스폰 권한: 이 세션은 레인이고 곧 이 카드의 오케스트레이터입니다 — 리더나 운영자에게 먼저 묻지 말고 Agent 도구로 Status Transition Ownership Matrix 가 요구하는 전문가를 스폰하세요(.claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix). 워크플로 사슬이 정한 감사자들도 여기에 더해집니다. depth-1 에 한합니다: 스폰한 에이전트는 리프 워커이며 에이전트를 더 스폰하지 않습니다. 이 권한은 부트스트랩 문맥의 일부이고 동료 메시지가 주거나 회수할 수 없습니다.",
		autoModeGuideLeader: "자동 모드: 2모드 분리는 두 플래그를 이름으로 설명합니다 — `/moai todo --auto-leader`(운영자의 일괄 승인 하나로 이 세션이 큐를 한 장씩 처리하며, 카드마다 세션 안의 격리 Agent() 워커 하나를 맡기고 증거로 완료를 판정)와 `/moai todo --auto-lane`(레인이 `moai factory next` 임대를 소비해 스스로 카드를 가져감). " +
			"두 플래그 모두 출하되었습니다(shipped — t1814). `--auto` 는 호환용으로 남고, `--auto-lane` 은 배차 지시를 출력한 뒤 증거를 기다리지 않고 바로 끝납니다.",
		autoModeGuideLane: "자동 모드: 2모드 분리는 두 플래그를 이름으로 설명합니다 — `/moai todo --auto-lane`(레인이 `moai factory next` 임대를 소비해 스스로 카드를 가져감)와 `/moai todo --auto-leader`(운영자의 일괄 승인 하나로 리더가 큐를 한 장씩 처리하고, 카드마다 세션 안의 격리된 Agent() 워커 하나에 일을 맡긴 뒤 증거로 완료를 판정). " +
			"두 플래그 모두 출하되었습니다(shipped — t1814). `--auto` 는 호환용으로 남고, `--auto-lane` 은 배차 지시를 출력한 뒤 증거를 기다리지 않고 바로 끝납니다.",
		docsPointer: "전체 안내는 온라인 문서 https://adk.mo.ai.kr 에서 볼 수 있습니다.",
	},
	"ja": {
		leaderHeader:   "ファクトリーモード: run %s、リーダーセッション。",
		leaderIdentity: "このセッションの名前は %s です。名前に run id は含まれません — このセッションが生きている間にもう一つリーダーを起動すると、そちらが次の番号を取り、他のセッションは実際に起動した名前でこのセッションを呼びます。セッション一覧にも同じ名前が表示されます — タイトルは最初のプロンプトで登録され、後から /rename すればそちらが優先されます。",
		leaderManual: "このセッションが、セッション間メッセージでカードをレーンに割り振ります。\n" +
			"セッションから別のセッションは起動できないため、レーンは新しいターミナルで `moai cc -l`(または `moai glm -l`、`moai codex -l`)を実行して始めてください。",
		entryGuide: "起動コマンド: `moai cc -f` は Claude、`moai glm -f` は GLM のファクトリーリーダーを起動します。" +
			"ランチャーがバックエンドを、`-f` がファクトリーを決めます。",
		agentFanout:  "各レーンは最大 10 個のエージェントを同時に並列実行できます。",
		leaderSocket: "リーダーソケット: %s",
		leaderClasses: "カードルーティング: すべてのカードは1つのレーンへ丸ごと割り当てられ、そのレーンがセッション内で直列3段階パス" +
			"(plan -> run -> sync、各段階の完了後に次へ)を最後まで実行します — カードを複数レーンに分割することはありません。\n" +
			"キューのポーリングとカードの選択はオペレーターまたはファクトリーフォアマンループ(単独の `/loop`)が担います。このファクトリーは選択済み(picked)カードを空きレーンへ振り分けます。",
		leaderStagger: "段階的起動 — 必須: すべてのレーンを同時に起動してはいけません。最初のレーンを起動し、実際に出力を生成し始めた証拠 " +
			"(最初のジョブまたは進行の形跡)を確認してから、残りの空きスロットのレーンを起動してください。 " +
			"同時リクエストは書き込み中のキャッシュエントリを読めません(cache-aware-execution directive 2)。 " +
			"このルールはファクトリーファンアウトにのみ適用されます — ワークフローランタイムは自身でスタガーします(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		gateSummary:       "plan→run Kickoff の判断が 2 件以上オペレーターの回答を待っているときは、カードごとに個別に尋ねず batch gate summary 1 つにまとめて提示してください — 書式と制約は `.claude/rules/moai/workflow/auto-semantics.md` §9.2 にあります。",
		operationalStatus: "factory_msg_status({\"run_id\":%q}) でレーンの稼働状態を確認してください。プロセスの生存は作業中である証拠ではなく、作業の観測がなければ unknown です。",
		settingsAuto:      "セッション間メッセージは、注入された --settings により自動的に受理されます。",
		settingsVerify:    "--settings ファイルに \"crossSessionInbound\": \"accept\" があることを確認してください。セッション間メッセージの受理に必要です。",
		laneJoin:          "ファクトリーモード: リーダーのレーン %[2]d 本の run に %[1]s として参加しました。",
		laneJoinNoCount:   "ファクトリーモード: リーダーのファクトリー run に %[1]s として参加しました。",
		laneNextCardRule: "ファクトリーレーン次カード規則: このセッションはセルフディスパッチレーンです。 " +
			"kept のワークツリーはそのまま残します — カードのワークツリーを削除してはいけません。 " +
			"親チェックアウトから次のカードを取得します。`moai factory next`(MCP ツール `factory_next`)を実行してください。 " +
			"セルフディスパッチレーンモードでは、`moai factory next` によるレーンキューの昇格はオペレーターが承認済みで、 " +
			"自分に割り当てられたカードは兄弟の割り当てが残っていても `moai factory next --card <id>` でリースします(カード t1542)。 " +
			"カードのワークツリーに入り、plan・run・sync を通してカードを完遂し、統合したうえでワークツリーを kept のまま残し、 " +
			"`moai factory complete`(MCP ツール `factory_complete`)で完了を記録します。 " +
			"その後はセッション内でカードを切り替えます：`ExitWorktree(keep)` でワークツリーを残したまま出て、親チェックアウトから `moai factory next` で次のカードを取得し、`EnterWorktree(<リースされたワークツリーのパス>)` で入って toplevel を確認し、前のカードの会話コンテキストは根拠にしません。 ただし完了出力がセッションを終了するよう示した場合（relaunch ポリシー）は、切り替えずにセッションを終了します。 " +
			"判断はレーンのものです(カード t1542) — workflow.jev.enabled が true なら `jev_ask` を直接呼び、そうでなければディスクの証拠から自ら判断し、 " +
			"カードの進捗記録に決定記録を書きます。リーダーに尋ねて待つことはしません。 ラダーは .claude/rules/moai/workflow/auto-semantics.md §6 にあります。 " +
			"ランディングもレーンが実行します(auto-semantics.md §13、mechanics § The lane's standard landing): " +
			"レーンローカル検証 → `codex_review` scope=card のカードレビュー → `moai factory complete`(push・PR・マージ観測) → " +
			"native Codex レビューの証拠を確認します。最新の状態サマリーと `codex-review-receipt:v1` が現在の head を示す必要があります（サマリーは full SHA、receipt は一致する SHA prefix）。コードレビューとセキュリティレビューの両方が完了し、未解決の指摘がないことを確認します。CI のステータス行だけではレビュー通過と判断しません。リモートへの反映を確認した後に `moai worktree sweep` を実行します。 " +
			"merged-pr が自分のキューカードを閉じます。 リーダーはキュー生産・ディスパッチ・レーン間衝突・スロットポリシー・keep-set ゲート中継のみを保持します。 " +
			"タスク規律: 最初の段階を始める前にカードの実行段階を `TaskCreate` で登録し、段階が移るたびに `TaskUpdate` で一覧を現在の状態に保ち、 " +
			"一覧が終了状態を反映しているときにだけ完了を記録します — 反映していない場合はその理由を明示します。 " +
			"ツールと CLI の対応: `todo_add`(`moai todo add`)、`todo_list`(`moai todo`)、`factory_next`(`moai factory next`)、 " +
			"`factory_stage`(`moai factory stage`)、`factory_complete`(`moai factory complete`)、`factory_decide`(`moai factory decide`)。",
		laneOwnedCardRule: "ファクトリーレーン担当カード規則: このセッションはカード %[1]s — %[2]s に記録された id — を担当し、現在のワークツリーで作業します。 " +
			"そのカードを merge-ready まで進めます。カードの引き受け時に実行段階を `TaskCreate` で登録し、段階が移るたびに `TaskUpdate` で一覧を現在の状態に保ち、 " +
			"各段階を `moai factory stage` で記録し、タスク一覧が終了状態を反映しているときにだけ `moai factory complete` で締めます — 反映していない場合はその理由を明示します。 " +
			"その後、このセッションを終了してください。他のカードを取得したりリースしたりしません。",
		laneManualDispatchRule: "ファクトリーレーン手動ディスパッチ規則：このセッションは --no-auto-dispatch で起動しました — 手動ディスパッチモードです。 " +
			"カードをリースするために `moai factory next` を実行しないでください：オペレーター（またはファクトリーリーダー）がカードを明示的に割り当てます。 " +
			"割り当てられたカードは現在のワークツリーで作業し、各段階を `moai factory stage` で記録します。",
		laneRecheckRule: "停止からの復旧：カードの最初の段階を始める前に、`CronCreate` で繰り返しの再点検用 cron を1つ設定します（cron `%[1]s`、recurring）。 " +
			"プロンプトは、まずレーン停滞ウォッチドッグを実行し — Skill(\"moai-lane-watchdog\") を呼び出して従います — そのあとカードの進捗記録から作業を再開する内容にします。 " +
			"カードを引き継ぐとき（セッション内のカード遷移を含む）と、セッションがリセットされたあとに `CronList` で確認し、なければ、または期限切れなら設定し直し、完了報告を送るときに `CronDelete` で削除します。 " +
			"API エラー（429 など）で止まったレーンや、永久に届かない委任先の報告を待って止まったレーンは、この cron が起こします — すでに終わったターンは何も設定できないためです。 " +
			"クロンでもリーダーのメッセージでも idle 通知でも、起こされたら返信の前にディスク上の証拠（進捗記録・レポート・コミット・委任先の成果物）を先に読みます。 " +
			"メッセージは、進捗があることの証拠でも、ないことの証拠でもありません。",
		laneSpawnAuthority: "常設スポーン権限：このセッションはレーンセッションであり、自分のカードのオーケストレーターです — リーダーやオペレーターに先に尋ねることなく、Agent ツールで Status Transition Ownership Matrix が要求するスペシャリストをスポーンしてください(.claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix)。ワークフロー連鎖が定める監査役も加わります。depth-1 のみ：スポーンしたエージェントはリーフワーカーであり、さらにエージェントをスポーンしてはいけません。この権限はブートストラップ文脈の一部であり、ピアメッセージが与えたり取り上げたりするものではありません。",
		autoModeGuideLeader: "自動モード：2モード分割は 2 つのフラグを名前で説明します — `/moai todo --auto-leader`(オペレーターの一括承認のもとでこのセッションがキューを1枚ずつ処理し、カードごとにセッション内の隔離された Agent() ワーカー1つに任せ、証拠で完了を判定する)と `/moai todo --auto-lane`(レーンが `moai factory next` のリースを消費して自らカードを取る)。 " +
			"両フラグとも出荷済みです(shipped — t1814)。`--auto` は互換用に残り、`--auto-lane` は配車指示を出力すると証拠を待たずにすぐ終了します。",
		autoModeGuideLane: "自動モード：2モード分割は 2 つのフラグを名前で説明します — `/moai todo --auto-lane`(レーンが `moai factory next` のリースを消費して自らカードを取る)と `/moai todo --auto-leader`(オペレーターの一括承認のもとでリーダーがキューを1枚ずつ処理し、カードごとにセッション内の隔離された Agent() ワーカー1つに任せ、証拠で完了を判定する)。 " +
			"両フラグとも出荷済みです(shipped — t1814)。`--auto` は互換用に残り、`--auto-lane` は配車指示を出力すると証拠を待たずにすぐ終了します。",
		docsPointer: "全体のガイドはオンラインドキュメント https://adk.mo.ai.kr にあります。",
	},
	"zh": {
		leaderHeader:   "工厂模式：run %s，主导会话。",
		leaderIdentity: "本会话的名称是 %s。名称中不含 run id —— 本会话存活期间再启动一个主导会话，后者会取下一个编号；其他会话按实际启动时的名称来称呼本会话。会话列表中也显示同一名称 —— 标题在首次提示时注册，之后 /rename 优先。",
		leaderManual: "本会话通过跨会话消息把卡片分发给各条泳道。\n" +
			"会话无法启动另一个会话，需要在新终端中运行 `moai cc -l`（或 `moai glm -l`、`moai codex -l`）来启动泳道。",
		entryGuide: "启动命令：`moai cc -f` 使用 Claude，`moai glm -f` 使用 GLM，分别启动工厂主导会话。" +
			"启动器决定后端，`-f` 决定工厂。",
		agentFanout:  "每条泳道最多可同时并行运行 10 个代理。",
		leaderSocket: "主导会话套接字：%s",
		leaderClasses: "卡片路由：每张卡片整体分发给一条泳道，由该泳道在会话内走完串行三阶段路径" +
			"(plan -> run -> sync，上一阶段完成后才进入下一阶段)—— 绝不把卡片拆分到多条泳道。\n" +
			"队列轮询与卡片挑选由操作者或工厂工头循环(单独的 `/loop`)负责。本工厂把已挑选(picked)的卡片路由到空闲泳道。",
		leaderStagger: "分批启动 — 必须: 绝不要同时激活所有泳道。先激活第一条泳道，等到它确实开始产出 " +
			"(首个任务或可见进展)之后，再激活其余空闲槽位的泳道。 并发请求无法读取仍在写入的缓存条目 " +
			"(cache-aware-execution directive 2)。本规则仅约束工厂分发 — 工作流运行时会自行错峰(CLAUDE_CODE_WORKFLOW_PREFIX_STAGGER_MS)。",
		gateSummary:       "当有两个及以上 plan→run Kickoff 决定在等待操作者答复时，不要逐张卡片提问，请合并为一份 batch gate summary 提交 —— 格式与限制见 `.claude/rules/moai/workflow/auto-semantics.md` §9.2。",
		operationalStatus: "使用 factory_msg_status({\"run_id\":%q}) 查询运行通道状态。进程存活不代表正在执行任务；没有任务观测依据时，状态为 unknown。",
		settingsAuto:      "跨会话消息通过注入的 --settings 自动接受。",
		settingsVerify:    "请确认 --settings 文件中包含 \"crossSessionInbound\": \"accept\"，跨会话消息的接受依赖该配置。",
		laneJoin:          "工厂模式：已以 %[1]s 身份加入主导会话的 %[2]d 条泳道的 run。",
		laneJoinNoCount:   "工厂模式：已以 %[1]s 身份加入主导会话的工厂 run。",
		laneNextCardRule: "工厂泳道下一张卡规则：本会话是自调度泳道。 " +
			"保持 kept 状态的工作树原样保留 — 绝不删除卡片的工作树。 " +
			"从父检出获取下一张卡：运行 `moai factory next`（MCP 工具 `factory_next`）。 " +
			"在自调度泳道模式下，通过 `moai factory next` 进行的泳道队列提升已获运营者授权； " +
			"分配给自己的卡即使兄弟分配仍在，也用 `moai factory next --card <id>` 租用（卡片 t1542）。 " +
			"进入卡片的工作树，带着卡片走完 plan、run、sync，完成集成后工作树保持 kept， " +
			"并用 `moai factory complete`（MCP 工具 `factory_complete`）记录完成。 " +
			"之后在会话内切换卡片：用 `ExitWorktree(keep)` 保留工作树并退出，在父检出运行 `moai factory next` 获取下一张卡，用 `EnterWorktree(<被租用的工作树路径>)` 进入并核对 toplevel，不再以此前卡片的对话上下文为依据；但若完成输出指示结束会话（relaunch 策略），则改为结束会话。 " +
			"判断属于泳道（卡片 t1542） — workflow.jev.enabled 为 true 时直接调用 `jev_ask`，否则依据磁盘证据自行判断， " +
			"并把决定记录写入卡片进度记录；绝不询问主导会话并等待。 阶梯见 .claude/rules/moai/workflow/auto-semantics.md §6。 " +
			"落地也由泳道执行（auto-semantics.md §13、mechanics § The lane's standard landing）： " +
			"泳道本地验证 → `codex_review` scope=card 卡片评审 → `moai factory complete`（push、PR、合并观测） → " +
			"请核对 native Codex 评审证据：最新状态摘要和 `codex-review-receipt:v1` 都必须对应当前 head（摘要须有完整 SHA，receipt 的 SHA 前缀须匹配）；代码评审与安全评审都已完成且没有未解决发现。仅凭 CI 状态行不能判定评审通过。确认远程落地后运行 `moai worktree sweep`。 " +
			"merged-pr 会关闭自己的队列卡。 主导会话只保留队列生产、分发、泳道间冲突、槽位策略与 keep-set 门转发。 " +
			"任务纪律：在开始第一阶段前用 `TaskCreate` 登记卡片的执行阶段，每次阶段切换时用 `TaskUpdate` 保持列表反映当前状态； " +
			"仅当列表已反映结束状态时才记录完成 — 否则明确说明原因。 " +
			"工具及其 CLI 等价物：`todo_add`（`moai todo add`）、`todo_list`（`moai todo`）、`factory_next`（`moai factory next`）、 " +
			"`factory_stage`（`moai factory stage`）、`factory_complete`（`moai factory complete`）、`factory_decide`（`moai factory decide`）。",
		laneOwnedCardRule: "工厂泳道负责卡规则：本会话负责卡片 %[1]s — %[2]s 中记录的 id — 并在当前工作树中工作。 " +
			"把该卡推进到 merge-ready：接手卡片时用 `TaskCreate` 登记执行阶段，每次阶段切换时用 `TaskUpdate` 保持列表反映当前状态， " +
			"用 `moai factory stage` 记录各阶段，仅当任务列表已反映结束状态时才用 `moai factory complete` 收尾 — 否则明确说明原因。 " +
			"然后结束本会话；不要领取或租用其他卡片。",
		laneManualDispatchRule: "工厂泳道手动调度规则：本会话以 --no-auto-dispatch 启动 — 手动调度模式。 " +
			"不要运行 `moai factory next` 去租用卡片：由操作者（或工厂主导会话）显式分配每张卡片。 " +
			"被分配的卡片在当前工作树中处理，各阶段用 `moai factory stage` 记录。",
		laneRecheckRule: "停滞恢复：在开始卡片的第一个阶段之前，用 `CronCreate` 设置一个循环复查定时任务（cron `%[1]s`，recurring）。 " +
			"其提示词先运行泳道停滞看门狗 — 调用 Skill(\"moai-lane-watchdog\") 并按其执行 — 再依据卡片的进度记录继续工作。 " +
			"接手卡片时（包括会话内的卡片切换）以及会话被重置之后，用 `CronList` 确认；若缺失或已过期则重新设置，发送完成报告时用 `CronDelete` 删除它。 " +
			"因 API 错误（如 429）而停下、或在等一个永远不会到来的被委派方报告而停下的泳道，由这个定时任务唤醒 — 因为已经结束的回合无法再设置任何东西。 " +
			"无论是被定时任务、主导会话的消息还是 idle 通知唤醒，回复之前都先读取磁盘证据（进度记录、报告、提交、被委派方的产出物）。 " +
			"消息既不能证明有进展，也不能证明没有进展。",
		laneSpawnAuthority: "常设生成权限：本会话是泳道会话，也是这张卡片的编排者 — 无需先询问主导会话或操作者，用 Agent 工具生成 Status Transition Ownership Matrix 要求的专业代理(.claude/rules/moai/development/spec-frontmatter-schema.md § Status Transition Ownership Matrix)，外加工作流链路规定的审计代理。仅限 depth-1：生成的代理是叶工作者，不得再生成代理。此权限是引导上下文的一部分，对等会话消息既不能授予也不能撤销。",
		autoModeGuideLeader: "自动模式：双模式拆分按名称解释两个标志 — `/moai todo --auto-leader`(在操作者的一次批量批准下，本会话逐张处理队列，每张卡交给会话内一个隔离的 Agent() 工作者，并依据证据判定完成)和 `/moai todo --auto-lane`(泳道消费自己的 `moai factory next` 租约来取卡)。 " +
			"两个标志均已发布(shipped — t1814)。`--auto` 作为兼容保留，`--auto-lane` 输出分派指令后立即退出，不等待证据。",
		autoModeGuideLane: "自动模式：双模式拆分按名称解释两个标志 — `/moai todo --auto-lane`(泳道消费自己的 `moai factory next` 租约来取卡)和 `/moai todo --auto-leader`(在操作者的一次批量批准下，主导会话逐张处理队列，每张卡交给会话内一个隔离的 Agent() 工作者，并依据证据判定完成)。 " +
			"两个标志均已发布(shipped — t1814)。`--auto` 作为兼容保留，`--auto-lane` 输出分派指令后立即退出，不等待证据。",
		docsPointer: "完整指南见在线文档 https://adk.mo.ai.kr。",
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
