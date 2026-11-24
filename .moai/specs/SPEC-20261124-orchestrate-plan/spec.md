---
id: SPEC-20261124-orchestrate-plan
title: "`moai card orchestrate plan` reads the ledger and the lane registry and prints free lanes, dispatches, clears, audit requests, the landing queue and user items"
status: approved
revision: 3
created: 2026-11-24
parent: SPEC-20261122-autonomous-orchestration
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8109 (2026-11-24T02:21:07Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 8108 에 답함): 판 3 승인(커밋 5dc75f4966)과 D-13 기본값(integration_branch, tip_of, contained 세 함수를 head 에 바로 씀); 노트 판정 reviewer 한계는 Result 에 한 줄; 레인은 그 노트를 인용만 해요"
  digest: "sha256:74bb0236acb94f8fa62c15aca7a876c853f32eab0bcac46d6aa60d8642054556"
---

# `moai card orchestrate plan`: the coordination window's round as one read-only plan

## Intent

사용자 요청(2026-11-22)과 사용자 선택(2026-11-22T23:49:36Z, 카드 노트 7144)이 정한 모양은 이래요. 조정은 코드와 원장이 하고, 상주 리더 세션은 없어요. `/moai leader` 로 등록된 창(C-329 가 착지한 경로 글)은 상주 역할이 아니라 그 코드를 한 번 돌리는 보통 창이에요. 11-22 의 `SPEC-20261122-autonomous-orchestration`(초안, 승인 없음)은 전체 모양을 적었고, 그 뒤 세 가지가 바뀌었어요. 등록된 조정 창이 생겼고(C-144, C-329, C-373), 메시지를 보내는 `moai host-wake` 와 타입 있는 판정 `moai card verdict` 가 착지했고, 공식 문서가 입력 소켓과 `claude agents --json` 필드를 적었어요.

지금 리더가 손이나 Orca 화면 읽기로 하는 일은 다섯 가지예요. (1) 빈 레인 찾기, (2) 빈 레인에 다음 카드 배차, (3) 끝난 레인에 `/clear`, (4) 검토 중인 머리에 감사 요청, (5) 감사를 통과한 머리의 착지 순서 정하기. 이 SPEC 은 그 다섯 가지를 "읽기만 하는 계획 하나"로 코드에 옮기는 첫 조각을 정해요. 계획은 원장과 레인 등록부(와 선택으로 호스트 상태 파일)만 읽고 행동 목록을 찍어요. 메시지, `/clear`, 감사 실행, 착지는 이 조각에서 하지 않아요. 계획이 먼저인 까닭은 세 가지예요. 실행은 계획의 같은 행동 키를 읽으면 되고, 계획은 호스트 프로세스 없이 임시 원장만으로 시험돼요, 그리고 틀린 계획은 아무것도 망가뜨리지 않아요.

옮기는 일과 그대로 두는 일 (Decisions D-3):

| 일 | 코드(이 SPEC 의 계획) | 실행 길 | 이 조각 |
| --- | --- | --- | --- |
| 빈 레인 찾기 | 등록부의 레인과 원장의 예약을 맞춰 `free`/`busy`/`unknown` | 없음(읽기) | 한다 |
| 배차 | 빈 레인마다 `card next` 가 고를 카드 하나 | `moai host-wake`(기본), Orca 어댑터는 `lane-send` | 계획만 |
| `/clear` | 끝난 카드와 열린 일 없음을 원장으로 확인 | Orca 터미널만. 메시지는 `/clear` 를 실행하지 못해요 | 계획만 |
| 감사 요청 | 판정 없는 머리와 저자 세션 | 비저자 `claude -p` 실행기(헤드리스 계약) | 계획만 |
| 착지 대기열 | 독립 PASS 가 있는 머리의 순서와 정확한 `card land` 줄 | `moai card land --by leader`, 등록된 창만 | 계획만 |
| 사용자 몫 | 결정 요청, 허가 프롬프트 | 사용자 | 목록만 |

### Read before writing (2026-11-23T21:58Z–22:02Z UTC, 설치된 Claude Code 2.1.331)

카드 본문과 노트 6058·6066·6081·6083·6116·6243·6279·7144, 완독 보고서 `skill-command-structure-readthrough-2026-11-23.md`(A–I 절), 적용 범위 점검 보고서 `coverage-check-2026-11-22.md` 의 A23·A25·D25 행, 측정 보고서 `bg-session-research-2026-11-22/report.md`, 이웃 SPEC 의 전문(`autonomous-orchestration`, `moai-leader-route`, `lane-completion-handshake`, `leader-registration-lease`)과 요구사항·결정 줄(`session-messages`, `headless-run-contract`, `nonstop-lane`), 그리고 `moai card --help`, `card next|inbox|wait|verdict|land|adopt --help`, `host-wake --help`, `session-register --help`, `leader-round --help`.

공식 문서는 `code.claude.com/docs/en/<page>.md` 를 `curl` 로 받아 읽기 전용 조사 에이전트 셋이 읽었고, 아래 표는 받은 사본의 체크섬이에요. 줄 번호는 그 사본의 줄이에요. 사본은 세션의 임시 폴더에 있어서 오래 가지 않으니, 이 SPEC 이 쓰는 문장은 아래 Facts 에 직접 인용해요. 읽은 범위가 일부인 쪽은 표에 적었어요.

| 쪽 | 받은 시각(UTC, 2026-11-23) | sha256 | 읽은 범위 |
| --- | --- | --- | --- |
| agent-view | 21:58:25 | 102fda743c24…5759c46 | 전문 |
| worktrees | 21:58:25 | 3912fa949021…5d109680 | 전문 |
| goal | 21:58:25 | a967e8db7460…41bb5d5 | 전문 |
| cli-reference | 21:58:25 | ec5680e89969…83297ce | 전문 |
| cross-session-messaging | 21:58:25 | d68cba02a859…b37f0 | 전문 |
| scheduled-tasks | 21:58:25 | 3bbc729ac8e3…f18ce | 전문 |
| sub-agents | 21:58:25 | 604914d74612…96639 | 전문 |
| agent-teams | 21:58:26 | d5ebbeeb8925…770bb9 | 전문 |
| remote-control | 21:58:26 | 51a15df1e0a3…4d3999 | 전문 |
| headless | 21:58:25 | b9bb6c1392ed…d116f5 | 전문 |
| statusline | 21:58:25 | 00d24928b789…d3954cd | 사용 가능한 데이터 절 전문 |
| hooks | 21:58:25 | b885b253fac8…e0f26 | 일부(1–230, 399–432, 741–1375, 2161–2235, 2567–3035, 3680–3725 줄) |
| tools-reference | 21:58:25 | acfb2f5d6090…56e4 | 일부(표와 Monitor 절) |
| settings-reference | 21:58:58–21:59:30 | 5f20b1359c41…ce6298 | 일부(sandbox, permissions.defaultMode) |
| sandboxing, permissions, hooks-guide | 21:58:25 | — | 읽지 않음(체크섬 안 적어요) |

읽지 않은 쪽과 일부만 읽은 쪽의 사실은 이 SPEC 의 REQ 가 기대지 않아요. 기대는 곳은 아래 Facts 가 인용한 것뿐이에요.

### Facts that decide the shape

각 줄은 쪽, 인용, 그 사본의 줄이에요. 모두 2026-11-23T21:58Z 에 받은 사본이고 버전 하한은 설치본 2.1.331 아래에 있어요.

- 상태 읽기: `claude agents --json` 은 "the supported way to read session state from outside Claude Code, for example from a status bar, a scheduler, or another Claude session that supervises background work" (agent-view L819). 필드는 `state`(`working`, `blocked`, `done`, `failed`, `stopped`), `status`(`busy`, `waiting`, `idle`), `waitingFor`(`permission prompt`, `input needed`, `sandbox request`, `worker request`, `dialog open`) 이에요 (L808–815). `~/.claude/jobs/<id>/` 아래 파일은 "not a stable interface" (L830). `blocked` 는 "always means the session needs something from you before it can continue" (L828). 이 기능은 "in research preview" (L21).
- 허가 프롬프트: "A permission prompt or another dialog … replying doesn't answer it. To answer the dialog, attach" (agent-view L228). 프로그램이 `--bg` 세션의 프롬프트에 답하는 길은 쪽에 없어요. 메시지는 "can't approve anything … a message from another session never counts as your consent" (cross-session-messaging L165).
- 메시지로 `/clear` 는 안 돼요: "A command like `/compact` in the text arrives as plain text. Claude Code never executes it" (cross-session-messaging L166–168). 예약 실행도 "Built-in commands … reach Claude as plain text instead of executing" (scheduled-tasks L52–55). 그래서 네이티브 길에서 레인을 비우는 방법은 메시지가 아니에요.
- 메시지 길: 도구 `SendMessage`/`ListAgents` 는 모델이 불러요 (L28). 세션마다 입력 소켓이 있고 `CLAUDE_CODE_MESSAGING_SOCKET`, `CLAUDE_CODE_MESSAGING_TOKEN` 으로 내보내요 (L243–255). 문서가 든 예는 "A script posting to its own session's socket" 뿐이고, 다른 세션의 소켓에 보내는 프로그램은 풀이가 없어요 (L255, L262–267). 11-22 SPEC 의 "소켓은 문서에 없다"는 이제 틀리지만 "다른 세션에 쏘는 프로그램의 배달은 보장된다"도 문서에 없어요. 그 동작은 `moai host-wake` 의 실측(SPEC-20261122-session-messages)만이 근거예요.
- 대기 알림: `notify_when_idle` 은 "Only the Claude in your main conversation can subscribe, and only to your sessions on this machine" (L107), 12시간 뒤 만료 (L100).
- 권한 모드 차이: 받는 세션이 우회 모드이고 보내는 세션이 아니면 "Claude Code holds each message for your approval" (L212). `-p` 일꾼은 `--settings` 의 `crossSessionInbound: accept` 로 시작해야 "unattended" 로 받아요 (L237). 5분 안에 안 열리면 버려요 (L216–219, `dialogExpiry`).
- 유휴 정지: 끝나고 한 시간쯤 붙지 않은 백그라운드 세션은 감독자가 프로세스를 멈추고 "the next time you attach or reply, the session resumes" (agent-view L847). `claude respawn <id>` 와 `claude --resume <full id> --bg "<prompt>"` 가 프로그램용으로 이어 가는 길이에요 (L531–539, L791). 실행 중인 백그라운드 세션에 `claude --resume <session> "prompt"` 는 터미널을 여는 길이라서 프로그램용이 아니에요 (cli-reference L128, sessions L29–39).
- 시작: `claude --bg "<prompt>"` 는 신뢰하지 않은 폴더에서 `Workspace not trusted` 로 끝나요 (agent-view L519). 성공 종료 코드와 `--bg --worktree <name>` 의 이름·브랜치는 쪽에 없어요(측정 전). 워크트리 기본 위치는 `.claude/worktrees/<name>`, 브랜치 `worktree-<name>`, 기존 브랜치에서 `baseRef` 를 고르는 길은 없어요 (worktrees L24, L177). 그래서 레인 워크트리는 계속 moai 가 git 으로 만들어요.
- 주의 신호: `Notification` 훅의 `agent_needs_input` 과 `agent_completed` 는 "while agent view is open in a terminal" 에서만 와요 (hooks L2573–2588). 훅은 알림을 막거나 바꾸지 못해요 (L2658). 그래서 사용자 몫 목록은 훅이 아니라 `claude agents --json` 의 `waitingFor` 와 원장 노트에서 읽어요.
- 한도: 상태줄 입력의 `rate_limits.five_hour|seven_day.{used_percentage,resets_at}` 는 "appears only for claude.ai Pro, Max, and Ultra subscribers, or behind a Claude apps gateway that sets a spend limit", 그리고 "only after the first API response in the session" (statusline L205–208, L356). 없으면 알 수 없음이고 0 이 아니에요. 헤드리스 일꾼의 핀과 기록은 `SPEC-20261122-headless-run-contract` 의 몫이에요.
- 계속 일하는 길: 자기 속도 `/loop` 는 `ScheduleWakeup`(1분–1시간), 7일 뒤 만료, `--resume` 으로 복원되지 않아요 (scheduled-tasks L132, L213, L243). `Monitor` 한 번의 감시 시간은 5분(기본)–30분이에요 (tools-reference L394–396). `/clear` 에서 루프가 어떻게 되는지는 쪽에 없어요(측정 전).
- 샌드박스: `sandbox.filesystem.allowWrite` 는 작업 폴더 밖 쓰기 경로를 더하고, `allowUnsandboxedCommands: false` 는 엄격 모드예요 (settings-reference L1415, L1518–1540). `sandboxing` 쪽은 읽지 않았어요.
- 지금 코드에 이미 있는 것: `moai card next|inbox|wait|verdict|land|adopt`, `session-register`(`--mode leader`, `--show`, `--continue`), `leader-round`(Orca 화면 관찰 파일을 읽어야 함), `host-wake`, `lane-send`(Orca 전용), `src/lane_registry.rs` 의 `session_entries`, `src/headless_run.rs`. `moai card orchestrate` 는 없어요 (`error: unrecognized subcommand 'orchestrate'`, 2026-11-23T22:00Z 에 쟀어요).

### Terms (revision 3)

- Judgement of a head: for one card and one head, the newest `card.verdict` event of that head; when the head has no event, the newest note of the same card (origin `local`) whose first line `gate::parse_verdict` reads for that head as `AUDIT … head <40 hex> base <40 hex> — verdict: <word>`, where the word `PASS` reads `pass` and every other word reads `fail` (conditions make a verdict FAIL, `skills/moai/references/audit.md`). An event always wins over a note, and the reviewer of a note judgement is the note's author. A line that is not the first line of a note, a note of another card, and a note that names another head judge nothing.
- Landed head: a head that `landed::contained` reports inside the tip of `landed::integration_branch`, the two functions and `landed::tip_of` that `landed::find` itself is built from. When git cannot tell, the head is not landed.

## Interview

- Q1: Who runs the pass? -> A: Any session may run `plan`, because it reads only. The acting steps (`run`, later slices) belong to the one window that is registered with `moai session-register --mode leader` (user pick 2026-11-22T23:49:36Z, note 7144), a registered coordination window that holds no standing role; `resident leader` stays a forbidden word.
- Q2: Which command name? -> A: `moai card orchestrate` with the action words `plan` now and `run`, `done` later (leader decision of 2026-11-22, note 6083; the flow skeleton allows no new top-level word).
- Q3: Why is the first slice read-only? -> A: The plan is the part every later step needs, it is testable on a temporary ledger with no host process, and a wrong plan breaks nothing (Intent).
- Q4: How does a lane get `/clear` on the native host? -> A: Not by message (facts above). The plan names the action and the channel it needs; with no terminal channel the plan says `channel: none` with the reason and the lane is cleared by its person or by the Orca adapter; one card per session through a new background session is a later measurement, not a promise here (D-5).
- Q5: What happens to the 11-22 SPEC? -> A: It stays `draft` and is not edited here. Its twelve requirements remain the long list; this SPEC fixes the order and the first slice, and every later slice gets its own revision and its own approval (D-1).

## Non-goals

- Starting, waking, continuing, clearing or messaging any session; dispatching a card; running an audit; landing. The plan only prints them. Those are later slices (Plan).
- Answering, approving or choosing for the user: decision requests and permission prompts are listed, never answered; a message never counts as consent (facts above).
- A resident leader, a daemon, a timer of moai or a loop prompt: repetition belongs to the caller (`internal/loop`, `lane-run`).
- The text of the LEADER route (`skills/moai/workflows/leader.md`) and the completion message: C-329 and C-373 own them. A later text slice there may call `moai card orchestrate plan`; this SPEC edits no skill text.
- The Orca adapter, the lane cap and usage-limit budget, Remote Control, other machines and Windows: later slices or other SPECs (D-6, D-7). Nothing here is measured on Windows.
- Reading Orca screens: the plan never runs `orca` and never reads a terminal screen.
- Changing `moai card land`, `card verdict`, `card next` or `leader-round`; the plan calls their read paths and leaves the commands as they are.

## Requirements

- REQ-1: When `moai card orchestrate plan` runs, the command shall read the ledger and the lane registry of the project and print one JSON answer with the parts `window`, `lanes`, `actions`, `land_queue`, `attention` and `host` · failure: none — a project without a ledger prints every part empty and `ledger: "absent"`, exit 0.
- REQ-2: The command shall write nothing and answer `writes` false · failure: none — the ledger file, the registry folder and the state folder stay byte for byte as they were.
- REQ-3: The `lanes` part shall list every registry entry of mode agent whose lease runs past now with its name, its host session id, and the state `busy` when a held reservation names that session or one of its previous sessions, `free` when none does and no held card is left behind, and `unknown` with the reason when the registry entry cannot be matched to a session · failure: none — an unreadable entry is `unknown`, never `free`.
- REQ-4: While a lane is `free` and a ready card could be taken, the `actions` part shall hold at most one `dispatch` action for that lane naming the card that `moai card next` ranks first among the cards no earlier action of the same plan named · failure: none — a card with no write scope, a held card and a card whose SPEC is not approved are listed as `blocked` with the reason and never dispatched, and a card whose newest hand-back head is landed, or is judged `pass` and waits for its landing, is never dispatched, while one whose newest judgement is `fail` is dispatched again for repair.
- REQ-5: When a lane's last card is `done` or released and the lane holds no open card, the `actions` part shall hold one `clear` action naming the lane and the channel `orca_terminal` or `none` with the reason · failure: none — a lane that holds a card or whose newest completion note the ledger does not back gets no `clear`.
- REQ-6: When a card that is not `done` has a newest hand-back note (`LANE-READY-INTEGRATION`, `REPAIRED` or `READY`, read with the marker table of `ready_receipt::marker_of_line`) that names a head which is neither landed nor judged, the `actions` part shall hold one `audit` action that names the card, the head, the base, every author session and the rule that the runner session differs from every author session, and no `dispatch` action for that card · failure: none — a card whose reservation was given back is `ready` or `active`, not `review`, and is audited all the same; an action that would name an author session as the runner is never printed; a landed head gets no `audit`; when git cannot read the integration branch no head counts as landed and the `attention` part lists the integration branch as `unread`.
- REQ-7: When the judgement of a head is `pass`, the head is not landed and its reviewer differs from every author session of the card, the `land_queue` part shall list that head with the exact command `moai card land <card> --head <40 hex> --by leader` and the `source` of the judgement, `event` or `note`, ordered by card priority and then by card number · failure: none — a head with a `fail` judgement, no judgement, a landed head, or a reviewer equal to an author is never listed.
- REQ-8: The `attention` part shall list each unanswered `LANE-BLOCKED` and `DECISION-QUESTION` note of a card that is not `done`, and each session of an agent-state file that waits (`waitingFor` set), with where it is answered and with no answer written for it · failure: none — an item whose source cannot be read is listed as `unread`.
- REQ-9: When the same ledger is planned twice, the command shall print byte for byte equal answers whose every action carries a stable `key` of the form `<kind>:<card or lane>:<anchor>` · failure: none — the anchor is the newest note id, the head or the session id the action answers, and the same key in a later plan is the same action.
- REQ-10: When `--agents-json` names a file, the command shall read that file in the shape of `claude agents --json` and put the `state`, `status` and `waitingFor` of each matching session into `lanes` and `attention` · failure: none — without the flag the `host` part reads `not_read`, and the command never starts `claude` or `orca`.
- REQ-11: The `window` part shall report `registered` with the lease end when the leader entry of the registry is held by the running session, `other` when another session holds it and `none` when it is free, using the read path of `session-register --show` · failure: none — the part never registers or renews.
- REQ-12: The help text and every printed field of the command shall hold no key-shaped text and no path outside the project folder and its own state folders · failure: none — a path outside is printed as the project-relative form or left out.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-11): on a temporary ledger with a registered leader entry, one free lane and one ready card, `plan` prints the six parts, `window` reads `registered`, `writes` is false, and the ledger file, the registry folder and the state folder are byte for byte as before; on a project with no ledger every part is empty with `ledger: "absent"` and exit 0 · verify: `cargo test --locked --offline --test orchestrate_cli plan_`
- AC-2 (REQ-3, REQ-4): a lane whose session holds a reservation is `busy`, one without is `free`, an unmatchable entry is `unknown`, a free lane gets one `dispatch` naming the card `card next` ranks first and a second free lane gets the next card, and a card with an empty scope or an unapproved SPEC is `blocked` with its reason · verify: `cargo test --locked --offline --test orchestrate_cli lanes_`
- AC-3 (REQ-5): a lane whose card closed and which holds nothing gets a `clear` with channel `none` and the reason that a message never runs `/clear`, a lane that holds a card gets none, and a completion note the ledger does not back gets none · verify: `cargo test --locked --offline --test orchestrate_cli clear_`
- AC-4 (REQ-4, REQ-6): a card with a `LANE-READY-INTEGRATION`, `REPAIRED` or `READY` head and no verdict gets one `audit` action that names every author session and the rule, whether it is in `review`, or `ready` or `active` after its reservation was given back, and it gets no `dispatch` while a free lane waits; a head that has a verdict gets no `audit`, a head whose newest verdict is `pass` and is not landed gets no `dispatch` and a head whose newest verdict is `fail` is dispatched again (also a head with a `pass` and then a `fail` of one reviewer, because the newest wins), a note that only mentions a marker (`MERGE-READY`, `READYING`, a marker inside a line) names no head, and the printed runner rule never names an author · verify: `cargo test --locked --offline --test orchestrate_cli audit_`
- AC-5 (REQ-7): two heads with `pass` verdicts by a non-author reviewer are listed in priority then number order with the exact `card land` line, a `fail` head, an unjudged head and a head whose reviewer is the author are not listed, and the plan lands nothing · verify: `cargo test --locked --offline --test orchestrate_cli land_`
- AC-6 (REQ-8, REQ-10): a blocked decision note and a stand-in agent file with `waitingFor: permission prompt` are listed as attention items with where they are answered, the answer text of the plan holds no approval or key press, an unreadable agent file is `unread`, and without the file `host` reads `not_read` · verify: `cargo test --locked --offline --test orchestrate_cli attention_`
- AC-7 (REQ-9): two plans of the same ledger are byte for byte equal, every action has a key of the stated form, and a new note changes only the keys it answers · verify: `cargo test --locked --offline --test orchestrate_cli keys_`
- AC-8 (REQ-12): the help text and the plan of every test hold no key-shaped text and no path outside the project · verify: `cargo test --locked --offline --test orchestrate_cli text_`
- AC-10 (REQ-4, REQ-6, REQ-7): on a ledger with no `card.verdict` event, a head whose first-line AUDIT note reads `PASS` gets no `audit`, no `dispatch`, and one `land_queue` entry with `source: note`; a `FAIL` note dispatches the card again; an event on the same head wins over a note in both directions (note `PASS` with event `fail` dispatches, note `FAIL` with event `pass` lists the head with `source: event`); an AUDIT line below the first line, a note of another card, a note that names another head and an AUDIT line that `gate::parse_verdict` cannot read judge nothing; a note whose reviewer (its author) is an author session of the card is never listed · verify: `cargo test --locked --offline --test orchestrate_cli judged_`
- AC-11 (REQ-4, REQ-6, REQ-7): in a temporary repository whose integration branch holds a head, a card whose newest hand-back head is inside the tip gets no `audit`, no `dispatch` and no `land_queue` entry even with a `pass` judgement, a head outside the tip is read as before, a project whose integration branch git cannot read treats no head as landed and lists the integration branch in `attention` as `unread`, and the plan changes no ref · verify: `cargo test --locked --offline --test orchestrate_cli landed_`
- AC-9: the neighbour checks pass: the help fixtures, the card command tests and the style checks · verify: `cargo fmt --check`, `cargo clippy --locked --offline --all-targets -- -D warnings` and `moai check affected <base>..<head>` at the READY head

## Scope

- this slice, write: `.moai/specs/SPEC-20261124-orchestrate-plan/spec.md`
- code slice, write (after the leader's approval; the code card's take names the exact paths): `src/card/orchestrate.rs` (new, the pure plan over rows the ledger and the registry already read), `src/card/mod.rs`, `src/card/command.rs` (the action word), `src/ops/card.rs` (the one call), `src/help.rs` with `tests/fixtures/help/**`, `tests/orchestrate_cli.rs` (new), and the user-facing `## Docs` section in the first code commit
- read-only: `src/gate.rs` (`verdict_line`, `parse_verdict`), `src/card/landed.rs` (`integration_branch`, `tip_of`, `contained`), `src/card/inbox.rs`, `src/card/verdict.rs`, `src/card/dispatch.rs`, `src/card/land.rs`, `src/lane_registry.rs`, `src/leader_round/registration.rs`, `src/host_wake.rs`, `src/headless_run.rs`
- forbidden: a real `claude` or `orca` process, any model call, any network, the real HOME, any write of a card, a note, a reservation or a verdict by the plan, every skill and workflow text, the integration branch

## Plan

1. This SPEC, read by the leader. The leader's approval is a decision note that names the digest of this revision; the lane writes no approval.
2. Code slice 1 (the only slice with requirements here), tests first: `tests/orchestrate_cli.rs` with the prefixes `plan_`, `lanes_`, `clear_`, `audit_`, `land_`, `attention_`, `keys_`, `text_` on a temporary home and a temporary project; RED on the missing command; then `src/card/orchestrate.rs`; one guard mutant per new guard test; one `check affected` at the READY head; the READY note.
3. Later slices, each with its own SPEC revision and approval, in this order: (a) `run` for `dispatch` through `host-wake` and the `done` record of a session action (the old SPEC's REQ-4, REQ-5, REQ-10); (b) the audit dispatch through the headless run contract (REQ-7 of the old SPEC), with the cap of three audits at once (D-7 of the old SPEC; this slice prints every audit it finds); (c) lane start and continue through the native adapter once the background-session measurement names keep (REQ-3, REQ-6) and the lane cap with the usage record (REQ-9); (d) the Orca adapter with `/clear` and the worktree path (D-6) and the retirement of the screen-reading `leader-round`, where `clear` also reads the host's prompt state (a lane at a prompt gets no `clear`); (e) a text slice in `workflows/leader.md` that calls `plan`, owned by the card that owns that file.

Revision 3 (repair after the audit of head `86e72f4c03`, finding R1 and the non-blocking R2; leader decision 2026-11-24T01:53:22Z): tests first on the prefixes `judged_` and `landed_` plus the R2 test under `audit_`; RED on the current code; then the change in `src/card/orchestrate.rs` only (the judgement read from events and notes, the landed read through `landed.rs`); one guard mutant per new guard test; one `check affected` at the REPAIRED head; the REPAIRED note. R3 of the audit (the head is read from the first 2000 characters instead of the first-line shape) stays the approved follow-up after landing, when `ready_receipt::marker_of_line` is on the base.

## Risks

- A plan from a stale agent file lists a session as waiting that is not: the file's read time is printed and the plan never acts on it; only a person reads it.
- A lane that was cleared has a new session id: the registry keeps `previous_sessions` from `session-register --continue`, and a lane that registered without it reads `unknown` (REQ-3), so a wrong `free` cannot send a card to a busy lane.
- A READY head that has moved: the plan lists it and `card land` answers `stale_base` or `tip_moved` itself; the plan does not guess the tip.
- A plan is not a lock: two windows may print the same plan. Only the registered window may act on it in the later slices, and the keys make a second pass find the first one's events.
- The native features are a research preview whose keys may change: the plan reads one file in the documented shape, so a change breaks one parser and nothing else.
- A sandboxed lane cannot land (the isolation report and coverage row A23): the plan prints the landing line for the registered window only, which stays outside the sandbox (note 6279, item 3). A window that is sandboxed needs `sandbox.filesystem.allowWrite` for `~/.moai`, which is a user pick; the plan does not decide it.
- The clear of a native lane has no documented program path: the plan says `channel: none` and does not hide it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | This SPEC fixes the order and the first slice; the 11-22 SPEC stays a draft and is not edited; each later slice gets its own revision and approval. | The 11-22 SPEC was never approved (card show of C-348, 2026-11-23T22:00Z); the facts it rests on changed (`host-wake`, `card verdict`, documented socket). | Fold this into a revision of the 11-22 SPEC. |
| D-2 | The command is `moai card orchestrate`, the first action `plan`. | Leader decision, note 6083 (2026-11-22). | Another command word. |
| D-3 | The five jobs move to code as a plan first; delivery stays with `host-wake`, `lane-send`, the headless runner and `card land` (table in the Intent). | `moai host-wake --help`, `moai lane-send --help`, SPEC-20261122-headless-run-contract, `moai card land --help` (read 2026-11-23). | Move delivery into the same command. |
| D-4 | The host state is one file in the shape of `claude agents --json`, passed as `--agents-json`; the plan starts no process. | agent-view L819 ("the supported way to read session state from outside Claude Code"); tests need no host. | Let the plan run `claude agents --json` itself. |
| D-5 | The plan names `clear` with `channel: none` on the native host and `orca_terminal` only when the profile selects Orca; it never plans a message that carries `/clear`. | cross-session-messaging L166–168, scheduled-tasks L52–55. | Plan a message `/clear` after a measurement says it runs. |
| D-6 | Orca is an optional adapter selected by the profile and later; it starts lanes in worktrees under `.claude/worktrees/<name>` made by moai and sets no Orca configuration. | Note 6243 (A25), SPEC-20261122-autonomous-orchestration D-5, user pick 2026-11-22 (worktrees under the project). | Write Orca's own project path from moai. |
| D-7 | The lane cap and the usage record are read in slice (c), not in the plan; an absent `rate_limits` field is unknown and never zero. | statusline L356; old SPEC D-9. | Read them in the plan. |
| D-8 | The plan prints the landing line for the registered window only and never lands; landing from a sandboxed session needs `allowWrite` for `~/.moai` (a user pick) or a landing from the window outside the sandbox. | Note 6243 (A23), note 6279 item 3, `moai card land --help` ("Only `--by leader` lands"). | Let a lane land. |
| D-9 | Every lane runs under the config root of the session the user talks to; the plan reads one registry. | User decision recorded in note 6066 and SPEC-20261122-session-messages REQ-5. The messaging page states no such requirement (searched 2026-11-23), so the decision rests on the measurement of 2026-11-22, not on the page. | Read several roots. |
| D-10 | Revision 2: an audit is owed by every card that is not `done` and whose newest hand-back note names a head with no verdict, and that card is not dispatched; only `review` was read in revision 1. | Audit of head `3858807ed5` (blocking finding B1, leader decision 2026-11-24T00:05:15Z): a lane that hands a card back gives its reservation back, so the card is `ready` or `active` (`skills/moai/references/flows/next-flow.md`, Lane); a probe on a copy of the real ledger printed `dispatch probe-1 C-433` and no audit. | Read only `review` and let `card next` rank handed-back cards. |
| D-11 | A card whose newest hand-back head has a `pass` verdict and is not yet landed is not dispatched either (REQ-4). | Leader decision, card note after 8062 (2026-11-24T00:51:54Z): the card waits for its landing and is no work for a new lane; the real ledger holds cards of that shape. | Let `card next` rank such a card again. |
| D-12 | Revision 3: a head is judged by its newest `card.verdict` event, and with no event by the newest first-line AUDIT note of the same card; an event wins. | Audit of head `86e72f4c03`, blocking finding R1: a probe on a copy of the real ledger found 0 `card.verdict` events and 52 requested audits, all verdicts living as AUDIT notes; leader decision 2026-11-24T01:53:22Z (a). The leader session once held C-348, so `moai card verdict` refuses it there (leader record 2026-11-24T01:53:29Z) and only a note carries that verdict. | Read events only (the plan then stays blind to every note verdict). |
| D-13 | Revision 3: a head inside the integration tip is landed and is neither audited, dispatched nor queued; the check is `landed::contained` on the head, with `landed::integration_branch` and `landed::tip_of`, not `landed::find` itself. | Leader decision 2026-11-24T01:53:22Z (b) names `landed::find` as the evidence to reuse; `find` proves a card by its newest recorded landed head or a merged pull request, which may differ from the hand-back head the plan reads, while the three functions are what `find` calls for its git proof (`src/card/landed.rs`, read 2026-11-24). The same audit's probe found 27 of the 52 requested heads already inside NeXT. | Call `landed::find` for the cards as well. |

## Result

Status: approved (leader note 7945); first code slice done: `moai card orchestrate plan`, read only.

Measured 2026-11-24 (UTC) on macOS in the worktree `.claude/worktrees/orchestrate-plan`, branch `spec/orchestrate-plan-20261124`, base `b175f6968b`. Code: `src/card/orchestrate.rs`, wired in `src/card/mod.rs` and `src/ops/card.rs`; the command row is in `docs/design-review/appendix/operations/card.md`. RED first: 18 of 18 tests of `tests/orchestrate_cli.rs` failed on `unrecognized subcommand 'orchestrate'` (commit `5e120d71ea`); after the code 18 of 18 passed, and the file now holds 19 tests.

- AC-1 met (REQ-1, REQ-2, REQ-11): `plan_prints_the_six_parts_and_writes_nothing` pins that the temporary HOME (ledger file, registry folder, state folder) is byte for byte equal before and after; `plan_of_a_project_without_a_ledger_is_empty_and_exits_zero` pins `ledger: "absent"` with empty parts and exit 0; `plan_window_reads_registered_other_and_none` pins the three window states and that the entry file is not rewritten.
- AC-2 met (REQ-3, REQ-4): `lanes_are_busy_free_or_unknown_by_their_reservations` (a previous session id counts, a leader entry is no lane), `lanes_a_free_lane_gets_the_first_ranked_card_and_the_next_lane_the_next`, `lanes_lane_order_is_numeric_not_alphabetical`, `lanes_cards_that_cannot_be_dispatched_are_blocked_with_a_reason`.
- AC-3 met (REQ-5): `clear_a_lane_whose_card_closed_gets_a_clear_with_no_channel`, `clear_a_lane_that_holds_a_card_gets_none_even_with_a_done_note`, `clear_a_reclaimed_card_that_is_still_open_gets_none`. The channel is always `none` in this slice; the Orca channel is a later slice (D-6).
- AC-4 met (REQ-6): `audit_a_ready_head_without_a_verdict_gets_one_audit_that_names_the_author`, `audit_a_head_with_a_verdict_gets_none_and_a_new_head_gets_one_again`; the action holds no runner field and the rule text never names the author.
- AC-5 met (REQ-7): `land_lists_independent_passes_by_priority_then_number` (a fail head, an unjudged head and an author-reviewed head are absent; the ledger is equal before and after), `land_a_later_fail_removes_the_head_from_the_queue`.
- AC-6 met (REQ-8, REQ-10): `attention_lists_decisions_and_waiting_sessions_and_answers_none` and `attention_an_unreadable_agent_file_is_unread_and_no_file_is_not_read`; the host file is a fixture file and no `claude` or `orca` process is started by the command or the tests.
- AC-7 met (REQ-9): `keys_two_plans_of_one_ledger_are_equal_and_every_action_has_a_key` (stdout equal byte for byte, keys of the form `<kind>:<x>:<anchor>`, unique), `keys_a_new_note_changes_only_the_keys_it_answers`.
- AC-8 met (REQ-12): `text_help_and_plan_hold_no_key_shaped_text_and_no_outside_path`.
- AC-9 not_run here: the neighbour checks (help fixtures, card command tests, style checks) are the `check affected` receipt of the READY note, run once at the READY head.
- Run: `moai heavy-slot-acquire --name orchestrate-plan --max 3 -- cargo nextest run --offline --locked --no-fail-fast --test orchestrate_cli` (with `RUSTC_WRAPPER=sccache`): 19 tests, 19 passed.
- Guard mutants (applied one at a time to `458bd5cf82` plus the REQ-7 tidy-up, checked as applied, run with nextest on the named tests, restored): A land author guard removed, B land `pass` guard removed, C audit `judged` skip removed, D clear condition `done || released` made `true`, E busy lane made free, F numeric lane order made alphabetical, G window `mine` made `true`, H lane-held `held` skip removed, I `spec_not_approved` check removed, J a file write added to the plan, K an unreadable host file read as an empty list. All eleven exited 100 (killed); the unmutated run exited 0.
- Deviation kept small: the first version of the landing queue also refused a head when any other reviewer's newest verdict was `fail`; REQ-7 asks only for the newest verdict, so that guard was removed (`7ae011e8a6`).
- Repairs after the first `check affected` (base `b175f6968b`, head `35327ae4ca`, `test=100`; 4261 tests, 4 failed): `orca_optional_3005` (the word `orca` in two source comments, removed), `state_folder_literal_scan` (a state folder name in the test file's header comment, reworded), `test_area_layout` (the word `orchestrate` added to the `card` area of `tests/areas.toml` through `card line-lease`, leader note after 7996). The fourth, `heavy_slot_cli with_one_slot_the_second_caller_waits_its_turn`, touches no file of this change; it passed when run alone on the repaired tree (heavy-slot, 58 tests of four binaries, exit 0), so it counts as a load-timing flake, not as a pass of the receipt.
- Not done and not claimed: the Orca `/clear` channel, any session started or messaged by code, the `leader.md` text that calls `plan`; they are the later slices of Plan item 3.

Revision 2 (repair after the audit of head `3858807ed5`, approved by the leader note 8063; digest of the revision `sha256:a2154e5f9aa5939ad5a65e6ea2489ea9579fac73993801169627ce6e3c988e4e`; the approval header was written in `e154d54682`). RED first: of 25 tests, the three new B1/N6 tests failed before the code change (22 passed, 3 failed; commit `513f91b74d`); after `c20166f6ef` 26 of 26 passed (`tests/orchestrate_cli.rs` now holds 26 tests).

- AC-4 met (REQ-4, REQ-6): `audit_a_handed_back_card_is_audited_and_never_dispatched` (a card in `ready` and one in `active` after the reservation was given back get an audit and no dispatch), `audit_a_failed_head_is_dispatched_again_and_the_repaired_head_is_audited` (a `fail` head is dispatched for repair, a `REPAIRED` head is audited), `audit_a_passed_head_that_waits_for_landing_is_not_dispatched` (D-11), `audit_the_hand_back_marker_table_is_the_one_of_the_ready_receipt_gate` (`REPAIRED`, `READY (`, a leading blank on the first line count; `MERGE-READY`, `READYING`, a marker inside a line and a first blank line do not).
- Guard tests of the audit findings that need no SPEC change: `attention_a_host_file_of_exactly_the_limit_is_read_and_one_byte_more_is_unread` (B2, 4 MiB both sides), `land_every_holder_of_an_earlier_generation_is_an_author_too` (B3, the landing queue and the `author_sessions` of an audit), `clear_names_only_the_card_of_the_lanes_newest_reservation` (N1).
- N4 met: the `card orchestrate plan` row of `docs/design-review/appendix/operations/card.md` (written under `card line-lease`, the file belongs to another card) now reads the audit rule of revision 2 and the landing rule of the code. N2 and N3 are written into Plan item 3 (slices (d) and (b)); no new card.
- N6 partly met: the marker table sits in `src/card/orchestrate.rs` with the same three markers as `ready_receipt::marker_of_line`, because that function exists only on the integration tip and not on the base `b175f6968b` of this branch. The swap for a call of that function is one line after the tip is merged at landing (open follow-up, not claimed).
- Guard mutants of revision 2 (each applied once to `c20166f6ef`, checked as applied, run with nextest on the one named test, restored; all exited 100): the retain of waiting cards removed; the `pass` verdict made not waiting; the audit read made `review` only; `REPAIRED` made a word that never matches; `READY ` made `READY`; the blank skip before the marker removed; the newest reservation made the oldest; the author set of the landing queue made the latest holder; the author set of an audit made the latest holder; the host limit made `>=`; the host limit made `+ 1`.
- Not claimed: N5, N7, N8 and N9 of the audit are non-blocking and left as they were; `claude agents --json` was never run, the host shape rests on the documented shape and fixture files.

Revision 3 (repair after the audit of head `86e72f4c03`, findings R1 and R2; approved by the leader note 8109, digest `sha256:74bb0236acb94f8fa62c15aca7a876c853f32eab0bcac46d6aa60d8642054556`, approval header in `8e8de433aa`). RED first (`b02e30ffd3`): of 34 tests 7 failed and 27 passed; after `3479210c92` all 34 passed; `cf7e30e168` adds the guard tests that the mutants below asked for, and `tests/orchestrate_cli.rs` now holds 35 tests. The change is in `src/card/orchestrate.rs` only (plus the `card.md` row, written under `card line-lease`).

- AC-4 met (R2): `audit_a_pass_then_a_fail_on_one_head_is_dispatched_again` pins that the newest verdict of a head wins (a `pass` then a `fail` is dispatched again and not queued).
- AC-10 met (REQ-4, REQ-6, REQ-7): `judged_a_pass_note_judges_a_head_that_has_no_event` (no audit, no dispatch, one queue entry with `source: note`, the ledger equal before and after), `judged_a_fail_note_or_any_other_word_dispatches_the_card_again` (`FAIL` and `CONDITIONAL`), `judged_an_event_wins_over_a_note_in_both_directions`, `judged_the_newest_note_of_a_head_wins`, `judged_lines_of_another_shape_judge_nothing` (an AUDIT line on the second line, a note of another card, a note naming another head, `head zz`, a quoted line in a hand-back note), `judged_a_note_whose_reviewer_is_an_author_session_is_never_listed`.
- AC-11 met (REQ-4, REQ-6, REQ-7): `landed_a_head_inside_the_tip_is_neither_audited_nor_dispatched_nor_queued` (a real temporary repository with the branch `integ`; a head inside the tip with or without a pass, a head outside it, a head that names no commit, no ref moved) and `landed_an_unreadable_integration_branch_counts_no_head_as_landed` (the audit stays, `attention` lists `integration` as `unread`, and nothing is listed when no head was handed back).
- AC-9 not_run here: the neighbour checks are the `check affected` receipt of the REPAIRED note.
- Run: `moai heavy-slot-acquire --name orchestrate-plan --max 3 -- cargo nextest run --offline --locked --no-fail-fast --test orchestrate_cli` (with `RUSTC_WRAPPER=sccache`): 35 tests, 35 passed.
- Guard mutants of revision 3 (each applied once to `cf7e30e168`, checked as applied, run with nextest on the one named test, restored; all exited 100, the unmutated run of the `judged_` and `landed_` tests exited 0): the event made not win over a note; the newest note made the oldest; the word `PASS` made "not FAIL"; the first line made the last line; the author guard of the queue removed; the landed check of the audit loop removed; the landed check of the queue removed; the landed card left out of the waiting set; the `unread` attention removed; a head that names no commit read as landed; the `source` label of a note made `event`; the newest event of a head made the oldest (R2).
- Known limit: the reviewer of a judgement that comes from a note is the note's author, so a name that is not a session id (such as `leader`) cannot be matched against the author sessions of the card; the plan only prints the line and `card land` makes the real check (leader note 8109).
- Not claimed: R3 of the audit (the head is read from the first 2000 characters instead of the first-line shape) is the approved follow-up after landing; the marker table swap of N6 is the same follow-up. No real ledger was planned in this run.

- Read as listed under "Read before writing": the pages marked "전문" were read whole by three read-only research agents; hooks, tools-reference and settings-reference were read in part and `sandboxing`, `permissions` and `hooks-guide` were not read. REQ-1 to REQ-12 rest on none of the unread or partly read text.
- Not run: nothing is built or tested here; the SPEC check is the only command of this slice and its result is in the card note that asks for the approval.
- Not measured and written as such: the exit code of a successful `claude --bg`, the name and branch of `claude --bg --worktree <name>`, the behaviour of a loop after `/clear`, whether a non-child program's message to another session's socket is delivered, and whether `/goal` runs as a `--bg` first prompt.
