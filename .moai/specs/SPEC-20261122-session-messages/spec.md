---
id: SPEC-20261122-session-messages
title: "Session messages ride Claude Code cross-session messaging; Orca mail only for Orca bots and Orca workers; the ledger stays the record and a message only wakes"
status: draft
revision: 2
created: 2026-11-22
parent: null
---

# Session messages: cross-session messaging between flow sessions, Orca mail only for Orca

## Intent

흐름 재설계의 host adapter 단계(작업 순서 5)가 쓸 메시지 길을 정해요. 사용자 결정(2026-11-22, 재설계 마스터 카드의 결정 기록): 세션 사이 메시지는 Claude Code 의 세션 간 메시지(cross-session messaging)로 보내 사용자가 Claude Code 안에서 직접 보게 하고, Orca 우편은 Orca 봇과 Orca 작업자 사이에만 쓰며, 원장(ledger)이 흐름의 기록으로 남고 메시지는 깨우는 역할만 해요. 같은 날 leader 세션에서 잰 agent 목록이, 레인 여덟이 Orca 터미널에서 돌고 있는데도 비어 있어서, 이 조각은 먼저 원인을 재고 공식 문서를 통째로 읽은 뒤 이 SPEC 을 써요. 코드는 없어요.

측정(2026-11-22T01:23Z–01:35Z, 전부 읽기 전용):

- **원인(실측)**: 세션 탐색 레지스트리는 `CLAUDE_CONFIG_DIR`(Orca 프로필 폴더) 루트마다 따로예요. 레인 8개(leader 포함)가 있는 루트의 `sessions/`에는 json 항목이 8개였고, 다른 프로필 루트 넷과 기본 `~/.claude`에는 1개(그 루트 자신의 leader, 11-16 시작)·0개·0개·0개·0개였어요. 다른 루트의 세션은 그 여덟을 전혀 보지 못해요 — 관찰된 빈 목록이 이 모양을 정확히 재현해요. GLM 엔드포인트는 원인이 아니에요(아래).
- **GLM 측정 넷**: (a) GLM 세션(agent-24, Claude Code 2.1.330, 엔드포인트 호스트 api.z.ai)의 ListAgents 가 peer 7개(leader, agent-3/15/16/17/22/23, 버전 2.1.326–2.1.330)를 보았고(01:23:30Z), 같은 세션이 leader 에게 SendMessage 로 보낸 글에 수신 상자 적재가 확인됐어요(01:23:55Z, 미보류·미거부). 역방향도 잡혔어요: 그 수신 세션이 보낸 답장이 cross-session-message 프레임(from-name leader, from-mode bypass, uds 소켓 주소)으로 이 GLM 세션에 도착했어요 — 송신 약 02:13Z(송신 세션이 적은 11:13 KST), 수신은 02:14:38Z 이전(턴 기록이 그 시각에 이미 인용). 왕복 약 49분은 수신 세션이 읽기까지의 지연이지 채널 지연이 아니에요. 같은 머신에서 GLM 라인↔다른 Claude Code 세션 양방향이 기록으로 성립해요. (b) `crossSessionInbound`·`isolatePeerMachines`·`dialogExpiry` 는 어느 설정 소스(프로젝트 settings.json·settings.local.json·사용자 settings.json·프로필 루트·관리 설정 없음)에도 없어서, 수신 기본값인 권한-클래스 두 분류 규칙이 GLM 세션에도 그대로 적용돼요. 공식 문서의 해당 문장: "The receiving session bypasses permission prompts: Claude Code holds each message for your approval. It delivers one only when the sending session identifies itself as also bypassing." — 즉 보류 여부는 수신자의 클래스가 먼저 정해요. 흐름 세션들은 bypassPermissions 로 서로 bypassing 이라 전달되고(실측과 일치), bypassing 발신자가 prompting 수신자에게 보내면 보류돼요(보고서 7-1의 hold 는 이 쪽 방향이에요). (c) 상태줄 stdin: 공식 문서는 `rate_limits` 가 claude.ai 구독(Pro/Max/Ultra)이나 게이트웨이 spend limit 세션에만 온다고 적어요. 서드파티 엔드포인트 세션에서는 absent 가 문서상 예상이고, 모델 카탈로그 캐시에 GLM 항목이 없는 것도 확인했어요(살아 있는 stdin 표본은 읽기 전용 범위 밖 — 잡지 않았어요). (d) 이 GLM 세션에는 Artifact·ArtifactComments·ArtifactData 도구가 제공되고(보기만 했고 아무것도 게시하지 않았어요), Remote Control 도구·연결은 없어요(레지스트리 항목의 peerFeatures 는 notify_idle, reply_across_default_dirs, artifact_yield).
- **채널 게이트(문서+실측)**: 기능 플래그 슬롯 `cachedGrowthBookFeatures.tengu_harbor_kite` 는 머신 전역, 마지막 기록자가 이기고, 일방(first-party) 세션만 써요. 지금 두 `.claude.json`(HOME 와 프로필 루트) 모두 `true`(01:33Z)라 GLM 세션의 채널이 열려 있어요. 값이 `false` 로 뒤집히면 서드파티 백엔드 세션 전체가 한꺼번에, 조용히, 채널을 잃어요(선행 연구의 반복 측정: 슬롯 true 두 번=inbox 소켓 있음, false 두 번=없음).

읽은 문서(전문): code.claude.com/docs/en/cross-session-messaging.md(2026-11-22T01:2xZ), 같은 statusline.md(스키마 전체, 01:29Z), 그리고 선행 연구 규칙 세 파일(cross-session-messaging.md·-core.md·-detail.md, 버전 1.3.0). 문서가 적는 채널의 한계: 같은 머신은 세션마다 소켓(macOS/Linux 유닉스 도메인, `/tmp/cc-socks/`, OS 사용자별 격리 — 실측: 내 세션 `/tmp/cc-socks/28912.sock` 바인딩, 디렉터리 44개 소켓 중 오늘 생성 약 8개), 머신 밖은 Anthropic 서버 경유이고 claude.ai 로그인/Remote Control 이 필요해 API 키·서드파티 엔드포인트로는 불가; bare mode 세션은 수신함이 없어 목록에 안 나옴; 메시지 약 100만 자 상한, 같은 세션에 순간 집중 전송은 송신 측에서 선거부, 수신 큐 50, 보류 상한 100, 기본 보류 대화시간 `dialogExpiry` 5분(명시적 `hold` 는 무기한); 컨테이너와 호스트는 서로 불가.

**누가 누구에 닿는가(실측+문서)**: 같은 루트의 8세션은 서로 닿아요(GLM↔first-party 양방향: GLM→first-party 송신 실측, first-party→GLM 수신은 이 세션이 받은 배차로 확인). 다른 루트끼리는 목록에서 안 보여요(실측). 클라우드·다른 머신은 이 세션에서 제공되지 않아요(문서: API 키·서드파티 엔드포인트 세션은 머신 밖 탐색 불가). Codex 세션은 Claude Code 런타임이 없어 구조적으로 불가해요.

### What exists (tip `62e457ceaf`)

- 깨움의 지금 길은 Orca 터미널이에요: `src/lane_send.rs:1-14` 는 `orca terminal send` 앞의 가드(화면을 읽어 상태를 분류한 뒤에만 typed), `src/leader_round.rs:2-11` 는 각 레인 터미널의 화면(`orca terminal read --screen --json`)을 읽어 상태를 적는 루프예요. 결정은 이 화면 길을 세션 간 메시지로 바꿔요.
- 흐름의 기록은 이미 원장이에요: 카드 노트·사건(event) 행, 첫 줄로 노트의 족보를 읽는 `family_of`(`src/card/inbox.rs:50`, `card inbox`), 타임라인 커서. 메시지가 깨우기만 하면 원장이 사실의 자리를 지켜요.
- `src/lane_registry.rs:139` 의 `session_live` 는 레인 이름으로 살아 있음을 묻고, `sessions/` 레지스트리(`registry_dir`)는 프로젝트 상태 아래에 있어요. Claude Code 자체의 탐색 레지스트리는 별개로 `CLAUDE_CONFIG_DIR/sessions/` 에 있고(실측), 두 레지스트리가 만나는 자리가 host adapter 예요.
- 세션 간 메시지 도구(`ListAgents`·`SendMessage`, `notify_when_idle` 요청)는 흐름 세션에 이미 들어 있어요(이 세션 실측). moai 가 부르는 쪽 코드는 없어요 — 이 SPEC 이 그 자리를 정해요.

## Non-goals

- 이 조각의 코드·시험 변경: 조사와 SPEC 만 써요.
- leader round·lane registry·화면 읽기 지우기: 재설계 순서의 host adapter 코드 조각이 그 자신의 카드로 해요. 이 SPEC 은 그 조각이 쓸 메시지 길의 요구만 정해요.
- 머신 밖·클라우드 세션에 메시지 보내기: Remote Control 과 claude.ai 로그인이 필요하고 GLM 라우팅 세션은 쓸 수 없어서 범위 밖이에요.
- 상태줄 다시 설계(진행 순서 표시 카드)와 `rate_limits` 표시: 이 SPEC 은 absent 가능성만 기록해요.
- Orca 봇·작업자 우편의 양식: Orca 쪽 설계가 따로 있어요.

## Requirements

- REQ-1: When the flow wakes another flow session, the wake shall be one Claude Code cross-session message addressed to the session's name whose text names only the ledger change to read, carrying no state the ledger does not already hold.
- REQ-2 (REQ-1): IF the addressee does not appear in the listing of reachable sessions, THEN the wake command shall report the constraint once with its reason (a different config root, no channel from the shared flag slot or the version, or a session without an inbox) and neither re-send nor write into the addressee's terminal.
- REQ-3: The wake command shall read the send result and report its three shapes (delivered to the inbox, held with a delivery notice, refused) without a second send · failure: none — a held or refused wake changes nothing in the ledger.
- REQ-4: moai shall address Orca mail only to Orca bots and from one Orca worker to another, never between two Claude Code sessions through Orca mail or a terminal send.
- REQ-5: The lane launcher shall start every flow session with the same Claude Code config root as the session the user talks to, because session discovery reads one config root (measured 2026-11-22: one root held the eight live flow sessions and every other root saw none of them).
- REQ-6: A wake to a GLM-routed session shall use the same message path as a wake to a first-party session, requiring neither a claude.ai sign-in nor rate-limit data (measured: a GLM session listed its seven peers and its message to a first-party session was delivered).
- REQ-7: The reachability of a started lane shall be checked from the session the user talks to before work is dispatched to it, by the lane appearing in that session's listing of reachable sessions.
- REQ-8: When a card lands, the card land command shall notify the sibling flow sessions of the landing as one cross-session message that names the landed card and the new integration head · failure: none — a landing whose notification cannot be sent still lands and the missed notice is read from the ledger, because a message only wakes.
- REQ-9: When an author waits for a verdict or a landing, the wait shall attach the one-shot idle notice to the message instead of polling the peer · failure: none — a wait that cannot subscribe falls back to reading the ledger cursor, never to polling the peer.
- REQ-10: When moai starts an unattended headless worker, the launcher shall set that worker's inbound control to accept in the settings the worker reads, so a default-held message cannot expire unseen · failure: none — a worker whose settings cannot be written starts anyway and its reachability is reported at start.

## Acceptance

- AC-1 (REQ-1): a wake is one SendMessage to a session name, the ledger row exists before the message leaves, and the text names the change and where to read it and nothing else · verify: `cargo test --locked --offline --test host_wake_cli ac1_`
- AC-2 (REQ-2, REQ-3): an addressee missing from the listing refuses once with the reason and sends nothing; a delivered, a held-with-notice and a refused send each report their shape with no second send · verify: `cargo test --locked --offline --test host_wake_cli ac2_`
- AC-3 (REQ-4): between two Claude Code sessions no Orca call is made, and the Orca transport appears only with an Orca bot or worker address · verify: `cargo test --locked --offline --test host_wake_cli ac3_`
- AC-4 (REQ-5, REQ-7): lanes the launcher starts under a temporary shared config root appear in the listing of a session in that root and receive a message, and a lane in another root is reported unreachable — the listing comes from a fake session registry the test builds, no real session is started · verify: `cargo test --locked --offline --test host_wake_cli ac4_`
- AC-5 (REQ-6): the wake path is the same call for a first-party and a GLM-routed addressee and reads no rate-limit field · verify: `cargo test --locked --offline --test host_wake_cli ac5_`
- AC-6 (REQ-8, REQ-9, REQ-10): a landing sends one notification naming the card and the new head to the sibling sessions, a wait attaches the idle notice once instead of polling, and a headless worker starts with its inbound control set to accept in the settings file it reads · verify: `cargo test --locked --offline --test host_wake_cli ac6_`

## Scope

- this slice, write: `.moai/specs/SPEC-20261122-session-messages/**`
- later slices, write: the wake command and its refusal reasons (`src/` module of the host adapter), the launcher's config-root sharing, the removal of the leader round, the lane registry and the screen reading (`src/leader_round.rs`, `src/lane_registry.rs`, `src/lane_screen.rs`, `src/lane_send.rs`, `src/ops/lane_send.rs`, `src/ops/leader_round.rs`), the tests that pin them
- forbidden: writing the real HOME or any config root of this machine; printing key material or the messaging token; a bulk edit of the real ledger

## Plan

1. (done) Research slice: documents read whole, the four GLM measurements, the cause with the registry-root evidence — this SPEC.
2. Wake command: one SendMessage to a session name, ledger row first, refusal reasons of REQ-2, result shapes of REQ-3; a fake session registry in tests stands for the listing.
3. Launcher: flow sessions share one config root with the session the user talks to; a lane start reports its reachability (REQ-7). The operator escape hatch `CLAUDE_CODE_HARBOR_KITE` stays undressed by moai (upstream internal flag; D-3).
4. Host adapter code slice (its own card): leader round, lane registry and screen reading leave, riding the redesign order; the Orca transport keeps only bot and worker addresses.

## Risks

- The shared flag slot is machine-global and last-writer-wins: a first-party session can flip it to false mid-run and every third-party-backend session loses the channel quietly. REQ-2's listing check is the detector; the ledger (REQ-1) is the fallback, so a lost wake costs nothing.
- Which config root an Orca terminal gets is Orca's choice today (five profile folders exist, two hold sessions). REQ-5 may need the launcher to export the root itself; how Orca names profiles is unverified.
- The peer protocol of 2.1.330 lists `reply_across_default_dirs`; whether a reply truly crosses config roots is unmeasured — the flow must not rely on it.
- A message from a bypassing sender to a prompting receiver is held by the default inbound rule; flow sessions run bypass and deliver, a user session in default mode holds. The inbox stays the fallback either way.
- `rate_limits` is absent in a GLM session's status line input (documented, subscription-only); anything that reads the status line must tolerate absence — the status line redesign card owns the display.
- A live sample of the GLM status line stdin was not taken: no read-only capture path exists (the configured status line renderer logs nothing). One operator-approved temporary capture can close this if the number itself is ever needed.

## Decisions

| D-8 | The CLI wake sends the minimum user line the leader's research measured (`{"type":"user","message":{"role":"user","content":"…"}}` plus one newline, one direction, close after) with no auth token read and no from-mode faked; a bypass lane would hold such a message, so moai sets `crossSessionInbound` to accept by default in the user settings of every config folder it installs into (a default item of the install/update preview, with one line on what it allows and how to turn it off — the user decision replacing the earlier pick), and the held-or-delivered status notice goes to a sender socket the CLI does not have — its answer is the queued shape. The model-session path stays the native SendMessage tool. | The leader's decision and the measured format (2026-11-22, session-socket-format report); the settings write rides the check of the settings-write card. | Send only after official docs name a format; make accept a pick again. |

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The flow messages same-machine sessions only; cross-machine and cloud stay out. | Beyond this machine needs Remote Control and a claude.ai sign-in, which a GLM-routed session cannot have (docs; measured absence of the tool). | Add a Remote Control slice later. |
| D-2 | Reachability is fixed by one shared config root, not by the `CLAUDE_CODE_HARBOR_KITE` escape hatch. | The hatch is an upstream internal flag; the root is the measured cause and a documented interface. | Export the hatch per session. |
| D-3 | moai never writes the shared flag slot and never injects the hatch; it only detects and reports. | Only first-party sessions write the slot (prior measurement); writing it is upstream's job. | Write the slot from the launcher. |
| D-4 | A wake message is a pointer into the ledger, never the record. | The queue-on-disk doctrine: a held or lost message must cost nothing. | Allow stateful messages. |
| D-5 | When messaging is absent the wake surfaces the constraint; it does not fall back to the terminal send. | The 2026-11-19 screen-send accident is why the guard exists; silence with a reason beats typing blind. | Restore the screen fallback. |
| D-6 | A wake may attach the one-shot idle notice instead of polling. | `notify_when_idle` replaces a polling loop (docs; step 4 of the redesign order: event wake instead of polling). | Poll the ledger. |

### Open questions (the user's pick before approval)

- REQ-5 — 흐름 세션 전체가 사용자가 말하는 세션과 하나의 config root 를 공유하는 것(Orca 프로필 구조와 얽힘).
- D-1 — 머신 밖·클라우드로 메시지를 내보내지 않는 것(Remote Control 도입 여부).
- D-2 — 탈출구 환경 변수를 쓰지 않고 루트 공유로 해결하는 것.
셋 다 사용자가 고르기 전까지 이 SPEC 은 draft 로 있어요.

## Result

Research-and-SPEC slice, done 2026-11-22. No code, no runtime acceptance criterion run (AC-1…AC-6 belong to the host-adapter code slice). Measured, with times: same-root peer listing 01:23:30Z (7 peers); GLM→peer send delivered to the inbox 01:23:55Z (not held, not refused); reverse peer→GLM reply framed and received — sent ≈02:13Z (sender-stated 11:13 KST), received before 02:14:38Z (first reference in this session's turn record); cause of the empty listing measured as per-config-root discovery registries (8/1/0/0/0/0 entries across the roots); GLM endpoint cleared; inbound defaults unset everywhere, so the two-class default rule applies (doc quoted in Intent); status-line rate_limits documented as subscription-only, absent expected on a third-party endpoint (no live sample taken, read-only); Artifact tools offered, Remote Control not. Both card criteria met with recorded evidence (research note and round-trip addendum). The three open questions above stay with the user; the SPEC stays draft until then.
