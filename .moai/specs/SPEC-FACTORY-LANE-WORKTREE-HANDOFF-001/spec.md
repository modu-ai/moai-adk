---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
title: "Factory lane card worktree handoff"
version: "0.5.9"
status: in-progress
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/factorymsg"
lifecycle: spec-anchored
tags: "factory,lane,worktree,codex,app-server,rebind"
tier: L
depends_on:
  - SPEC-FACTORY-MIXED-HOOK-001
card: t1082
---

# SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.5.9 | 2026-09-23 | 리드 결정: v058 minor M1·M2 를 재감사 없이 M4 에 흡수 — M1 순서 정정(의미 불변), M2 hook 다리 단언 추가. 근거는 `.moai/reports/t1082/plan-audit-delta-v058.md`의 M1·M2다. M1: acceptance.md AC-FLH-007에서 BOUND 행 다리를 launcher 다리 앞으로 옮겨 서술 순서를 상태 진행 순서와 맞췄고, 문장·명령·jq 게이트·named test는 바꾸지 않았다. M2: AC-FLH-003 hook 안내문 다리에, 현재 endpoint가 launch-pending 행일 때 tombstone된 session의 SessionStart 안내문이 같은 상태의 UserPromptSubmit 안내문(빈 session 렌더링)과 바이트 동일하고 행을 바꾸지 않는다는 단언을 더했으며, 요약행과 RED 원장(HEAD `ec33efa6a`, exit 1)·앵커·양성 대조·행동 탐침 관측을 기록했다. |
| 0.5.8 | 2026-09-23 | 델타 plan-audit(`.moai/reports/t1082/plan-audit-delta-v057.md`) FAIL 0.80의 N1–N5를 한 개정으로 모두 고쳤다. 리드 결정: Tier L이고 사용자에게 보이는 동작이 AC 없이 run에 들어가면 안 되므로 N1–N5를 함께 고친다. Jev 판정(0.47/0.45)은 결정적이지 않아 근거로 쓰지 않았고 독트린으로 정했다. N1: REQ-FLH-017 종착 문장에서 「결합된 현재 endpoint 정확히 하나」와 「고아 launch-pending 없음」 두 조항이 같은 예외를 받도록 문장을 다시 짰다. REQ-FLH-010으로 bind를 거부당한 살아 있는 launcher owner가 쥔 launch-pending 행은, 그 owner가 current인 동안 결합되지 않은 채 lane의 유일한 현재 endpoint로 남으며 고아가 아니다. design §2.1 정합 항목도 맞췄다. N2: AC-FLH-007 launcher 다리의 bind 입력 `Generation`을 production SessionStart hook의 상수 1로 고정하고 tombstone generation과 다름을 테스트가 먼저 단언하게 했으며, 송신 다리 두 개의 identity를 「거부된 bind가 만들었을 identity」(tombstone된 UUID, launch-pending generation + 1, launcher PID·process-start)로 고정했다. N3: AC-FLH-007에 이미 BOUND인 행에서 tombstone된 UUID의 `BindLaunchPending`이 `STALE_ENDPOINT`로 거부되고 아무것도 쓰지 않는 다리를, AC-FLH-003(hook 패키지 named test)에 BOUND 뒤 이전 session의 SessionStart가 기존 endpoint-replaced 안내문을 그대로 내는 다리를 더했다. REQ-FLH-010 추적에 AC-FLH-003을 더하고, 두 다리의 RED 원장(HEAD `03390136a`)과 앵커·양성 대조를 기록했다. N4: design의 운영자 신호 서술을 코드에 맞췄다. roster 값은 `launch_pending`이고, 기존 안내문은 조치를 지시하지 않으며, launch-pending 행이 현재 endpoint일 때 session 자리는 빈 값으로 렌더된다. 빈 session 렌더링은 새 요구가 아니라 잔여 위험으로 적었다. N5: plan M4 재실행 목록을 `rg -l 'BindLaunchPending' internal --glob '*_test.go'`의 실제 출력 8개 파일과 그 세 패키지로 바꿨다. D3·D5·D6은 건드리지 않았다. |
| 0.5.7 | 2026-09-23 | 델타 plan-audit(`.moai/reports/t1082/plan-audit-delta-v056.md`) FAIL 0.80의 D1·D2·D4를 리드 결정대로 고쳤다. D1은 선택지 (b)로 정했다(Jev noul 0.82). REQ-FLH-010의 「resume까지 포함한 영구 거부」가 launcher resume 경로에도 걸리도록, SessionStart launcher bind(`BindLaunchPending`)가 자기 write transaction 안에서 tombstone 집합을 session UUID만으로(generation 무관) 읽고 `STALE_ENDPOINT`와 현재 endpoint redirect로 거부하며 아무것도 쓰지 않게 했다. 거부 뒤 launch-pending 행은 launcher 등록이 commit한 그대로 두고, 그 owner가 t1074 규칙으로 current가 아니게 된 뒤 새 launcher 등록이 대체한다. 행 삭제(rollback)는 generation을 되돌리므로 택하지 않았다. REQ-FLH-017의 dead-owner 재등록 문장과 「고아 launch-pending 없음」 불변식에 이 경우를 명시했다. 운영자 신호는 SessionStart hook의 기존 `STALE_ENDPOINT` 안내문과 lane roster의 `launch-pending` 결합 상태다. AC-FLH-007에 launcher 다리와 양성 다리를 더하고, 사실과 어긋났던 launcher 경로 제외문과 그 근거(D4: 뒤이은 SessionStart bind는 이전 UUID를 쓴다)를 지웠으며, 요약행을 맞추고 RED 원장에 launcher 다리 측정(HEAD `113daeb83`, exit 1)과 앵커·양성 대조를 더했다. 구현은 M4에서 하며, 닫힌 카드 t1074가 소유한 `internal/factorymsg/store.go`를 건드리는 비용은 리드 결정으로 받아들였다. D2는 AC-FLH-020에 허용 seam을 이름으로 적었다. 종결 store 함수가 `func(int) (string, homestate.ProcessIdentityState)` probe를 인자로 받고(선례 `internal/cli/factory_handoff_recover.go:30` `RecoverLegacyResume`), cli 명령이 그것을 패키지 변수로 연결하며, 테스트는 그 변수만 바꿔 Live/Dead/Indeterminate를 만든다. 이 다리가 `ownerCurrent` bool 재사용 변이를 잡는다는 점도 적었다. D3·D5·D6은 건드리지 않았다. |
| 0.5.6 | 2026-09-23 | 리드 결정: 0.5.5에서 REQ-FLH-010에 더한 「tombstone된 session UUID는 나중 resume까지 포함해 영구히 거부된다」를 단언하는 AC가 없었으므로, 새 AC를 만들지 않고 REQ-FLH-010이 이미 매핑된 AC-FLH-007에 다리 하나를 더했다. BOUND rebind가 tombstone한 이전 session UUID로 t1074 UserPromptSubmit 등록 경로(`RegisterPeer`, launch-pending이 아닌 session)를 통해 나중 resume·재기동·broker 핸들 재개방 뒤 재등록하면 각각 `STALE_ENDPOINT`로 거부되고, `StaleEndpoint` redirect가 현재 endpoint session·generation을 담으며, endpoint 행·tombstone·BOUND receipt가 바뀌지 않는다. launcher 경로는 launcher bind가 tombstone을 읽지 않으므로 단언 대상에서 명시적으로 뺐다. named test·명령·jq 게이트·요약행·추적표는 바꾸지 않았고, acceptance.md RED 원장에 이 다리가 아직 named test에 없다는 측정(HEAD `e2c2d33b1`, exit 1)과 앵커·양성 대조를 더했다. |
| 0.5.5 | 2026-09-23 | 리드 결정 두 건을 한 개정으로 반영했다. (1) acceptance.md AC-FLH-008 수신 측 문구만 명시적으로 고쳤다(문구 변경, 판정식 불변). 「같은 recipient generation 안의 K1 봉투 반복 전달」을 「receipt 전에 lease가 만료되어 같은 봉투가 다시 claim되는 경우」로, 「one accepted receipt」를 「broker가 수락한 receipt row가 정확히 1건」으로 바꿨다. 근거: 현재 기준에서 같은 봉투가 다시 도착하는 경로는 at-least-once lease 만료 재전달뿐이며, named test `TestFactoryLaneHandoffDuplicateAndSameLaneRedispatch`가 이미 그렇게 모델링한다. named test·명령·jq 게이트·요약행은 바꾸지 않았다. (2) 막힌 handoff의 operator 종결 경로를 추가했다. 비종결 handoff 동안 REQ-FLH-018이 UserPromptSubmit 등록을 모두 거부하므로, `/cd`를 하지 않고 launcher 밖에서 재기동한 interactive lane은 영구히 거부되고 시간 초과나 포기 계기도 없었다. REQ-FLH-011에 `moai factory handoff abandon-lane --slot <slot>`을 규정해 source owner가 current가 아님을 t1074 PID·process-start 규칙으로 확인한 뒤(아니면 `SOURCE_OWNER_LIVE` 거부) 한 transaction에서 `ABANDONED`/`OPERATOR_ABANDONED`로 종결하고, worktree 보존·BOUND/tombstone/receipt/release 무쓰기·종결 뒤 t1074 복귀를 요구했다. AC-FLH-020을 추가했고, 시간 기준 자동 종결은 명시적으로 범위 밖에 두었으며, sync 문서가 이 명령을 적도록 했다. design.md §3 `WT_READY` 전이·§2.1 F②-3·§9 결정표와 plan.md M4를 맞췄다. 또 REQ-FLH-010에 tombstone된 session UUID는 이전 session의 나중 resume까지 포함해 영구히 거부된다는 문장과 그 이유를 더했다. |
| 0.5.4 | 2026-09-23 | idempotency 기준에 대해 중립으로 고쳤다(리드 조율: t1100이 AC-020 근거에 따라 기준을 송신 session에서 송신 lane slot 범위로 옮긴다. 근거 기록 `.moai/reports/t1082/ac008-idempotency-check.md`). 기준이나 스키마가 「바뀌지 않는다」고 단언하던 문장을 REQ-FLH-009, design.md §2.1 잔여 위험·§8, plan.md M3, acceptance.md AC-FLH-008·요약행에서 걷어내고, 「t1082는 idempotency 기준 자체를 바꾸지 않으며 기준은 t1100(SPEC-DUAL-HARNESS-RECOVERY-001)이 소유한다」로 바꿨다. AC-FLH-008은 어느 기준에서도 성립하도록 다시 적었다: 다른 recipient로 K1을 재사용한 요청은 기존 K1 봉투로 합쳐지지 않는다는 것만 단언하고 거부인지 별도 봉투인지는 단언하지 않으며, fixture가 key 기준에 기대지 않아야 한다는 조건을 더했다. 같은 generation 안 body 1회 실행, 수신 측 `DispositionDuplicate`, 이전 generation stale NACK, BOUND 뒤 새 key 요구는 그대로다. |
| 0.5.3 | 2026-09-23 | plan.md만 바꿨다. AC-FLH-003/004의 named test가 BOUND 관측을 요구하는데 BOUND는 M3 atomic rebind 한 transaction의 산출물이므로(REQ-FLH-008, design.md §6), M2에서 headless 원자적 BOUND 문구를 빼고 두 adapter를 `SWITCH_PENDING_*`까지로 한정했으며 headless BOUND와 AC-FLH-003/004 named test를 M3로 옮겼다(lane 결정 option A). AC 본문은 바꾸지 않았다. |
| 0.5.2 | 2026-09-23 | AC-FLH-008을 좁혔다(리드 결정 (a), 근거 `.moai/reports/t1082/ac008-idempotency-check.md`). `Store.Send`는 같은 `(sender_session, idem_key)`라도 recipient가 다르면 거부하므로, BOUND 전후 같은 key로 보낸 재전송은 duplicate가 될 수 없다. 그래서 BOUND 뒤 rebound endpoint로의 재전송은 새 key를 쓰게 하고, 같은 key duplicate 처리는 같은 recipient generation 안에서만 요구했다. stale-generation NACK와 같은 generation duplicate의 body 1회 실행은 유지했다. 송신 측 `Send` 판정과 수신 측 receipt 처분 `DispositionDuplicate`를 층별로 나눠 적었다. 스키마와 idempotency 기준은 바꾸지 않았다(t1100 소관). REQ-FLH-009, design.md §2.1 잔여 위험·§8, plan.md M3, acceptance.md 요약행을 같은 기준으로 맞췄다. |
| 0.5.1 | 2026-09-23 | run 진입 채무 정리(plan-audit iter-6 N7–N11): reservation이 source 행을 자기 write transaction 안에서 읽는지 판별하는 AC-FLH-019 순서 (vii)를 추가하고 REQ-FLH-016을 AC-FLH-019에 추적 연결했으며, REQ-FLH-017의 「나중 launcher 등록은 결합 endpoint를 바꾸지 않는다」를 결합 owner가 current인 동안으로 한정했고, fresh reservation 재진입의 dirty target 사유 `TARGET_DIRTY`를 정의했다. AC-FLH-018/019 fixture 문구와 plan 마일스톤도 맞췄다. |
| 0.5.0 | 2026-09-23 | 미종결 handoff 중 launcher 가등록이 같은 transaction에서 handoff를 `NACK`/`STALE_GENERATION`으로 종결하도록 규정해 재기동 lane의 t1074 결합 경로를 복원했고(REQ-FLH-017, 결정 ②), t1074 결합자 서술을 UserPromptSubmit 필수·SessionStart best-effort로 바로잡았으며(REQ-FLH-015), 같은 transaction 판독 판별 순서·NACK 뒤 fresh reservation 진입·idempotency key 기준을 명시했다. |
| 0.4.0 | 2026-09-23 | t1074 착지분의 세 번째 slot 쓰기 경로인 UserPromptSubmit `RegisterPeer`를 모델링해, 미완료 handoff 동안 그 경로가 같은 broker transaction 안에서 거부되도록 규정했다(REQ-FLH-018). |
| 0.3.0 | 2026-09-23 | t1074 착지(develop `861510fb6`) 후 전제 재검증 결과를 반영해 launch-pending source endpoint의 handoff 거부(REQ-FLH-016)와 launcher provisional bind 대 handoff rebind의 직렬화·순서(REQ-FLH-017)를 추가했다. |
| 0.2.0 | 2026-09-22 | 관측된 Codex 0.155.1 경계를 반영해 interactive는 다음 정상 turn의 SessionStart, headless는 공식 app-server 반환 thread ID를 mode별 결합 증거로 분리했다. |
| 0.1.0 | 2026-09-22 | 카드 t1082의 안정 lane → 전용 card worktree 전환 및 generation-safe endpoint rebind 계약을 최초 정의했다. |

## WHY

혼합 factory의 안정 논리 lane은 카드를 배차받은 뒤에도 기존 checkout과 물리 session/thread endpoint에 머물 수 있다. 그 상태에서 곧바로 dispatch body를 처리하면 잘못된 cwd 수정, primary checkout의 branch 전환, stale generation ACK, worktree base drift가 섞인다. t1074가 만든 안정 lane 주소와 current endpoint seam을 유지하면서, 카드 전용 L1 worktree와 새 물리 endpoint를 검증된 순서로 결합해야 한다.

## WHAT

이 SPEC은 카드 배차 뒤 다음 순서를 하나의 내구성 있는 handoff 계약으로 정의한다.

`reserve → local develop HEAD pin → MoAI launcher L1 worktree create → WT-<slug> + card traceability → mode별 SWITCH_PENDING → interactive의 다음 정상 turn SessionStart 또는 headless의 공식 RPC 결과 검증 → atomic rebind → BOUND receipt → dispatch body`

논리 lane ID는 재배차와 재시작에도 유지되는 주소다. Codex thread/session UUID와 generation은 handoff마다 바뀔 수 있는 물리 endpoint다. interactive 경로와 headless 경로는 trigger와 실패 의미를 분리하지만, 같은 reservation nonce, provenance 검증, atomic rebind, stale-endpoint 차단 규칙으로 합류한다.

## HOW

- t1074의 `internal/factorymsg` broker/peer generation/receipt와 `internal/homestate` canonical project/run identity를 확장한다.
- worktree는 기존 MoAI launcher의 L1 materializer만 사용한다. raw `git worktree add` 경로를 새로 만들지 않는다.
- headless Codex는 기존 app-server JSON-RPC client를 확장하여 공식 `thread/fork` 또는 `thread/start`의 target `cwd` 요청과 반환 thread ID를 사용한다. 결합을 위해 SessionStart나 빈 `turn/start`를 만들지 않는다.
- interactive Codex는 idle 상태에서 사용자가 직접 `/cd`를 실행하도록 안내한다. hook, MCP, tmux `send-keys`, private socket, 모델에게 보내는 “cd 해라” 메시지는 relocation 수단이 아니다.
- rebind, old endpoint tombstone, durable `BOUND` receipt, pending dispatch release를 기존 factory broker의 한 transaction에 둔다. 새 broker나 daemon은 추가하지 않는다.

## User story

factory lead 운영자로서 안정 lane에 카드를 배차한 뒤 그 lane의 대화 연속성을 잃지 않고 카드 전용 worktree로 옮기고 싶다. 새 endpoint가 올바른 cwd와 branch에 실제로 결합되기 전에는 어떤 구현도 시작되지 않아야 하고, 이전 endpoint나 중복 dispatch가 뒤늦게 나타나도 현재 endpoint를 덮어쓰거나 ACK하지 못해야 한다.

## Scope and measured baseline

- 카드: `t1082`, Tier L.
- 대상 worktree/branch: `.claude/worktrees/t1082`, `WT-factory-lane-worktree-handoff`.
- 현재 plan 기준 HEAD: `bf39a539d97f49edf3b11517ee7c982239c60df3`.
- 이 worktree는 `moai worktree new t1082` 실행 시 primary `main@2213871af`에서 잘못 시작했고, 이후 `git merge --ff-only develop`로 `develop@3f3ffbb57`에 맞춰졌다. 이 관측된 creation-base drift는 REQ-FLH-004와 AC-FLH-016의 필수 실패 사례다.
- t1074 의존 merge는 `bf39a539d`, 의존 commit `8c5d9be99`는 현재 HEAD의 ancestor다.
- 현재 MCP catalog는 단위 검증에서 36개(쓰기 14, 읽기 22)로 관측됐다. 이 SPEC은 새 MCP tool을 기본 해법으로 추가하지 않는다.

## Requirements (GEARS)

### REQ-FLH-001 — Stable lane, replaceable endpoint

While a factory handoff is active, the handoff SHALL keep the stable logical lane ID as the address, treat the thread/session UUID and generation as a replaceable physical endpoint, and scope every handoff state, receipt, and dispatch to that lane within one canonical project key and run ID.

### REQ-FLH-002 — Separate interactive and headless state machines

Where the lane is interactive Codex, the handoff SHALL allow only a user-driven `/cd` relocation followed by the user's next normal turn. Where the lane is headless Codex, the handoff SHALL allow only the official app-server `thread/fork` or `thread/start` with the reserved target `cwd`. The two paths SHALL use distinct transitions, evidence, and error codes and converge only at the validated atomic rebind after their mode-specific `SWITCH_PENDING` state.

### REQ-FLH-003 — Fail-closed admission

When the source lane has an active turn, a permission wait, an interrupt in progress, a dirty source checkout, a stale reservation, an untrusted cwd, a branch owned by another worktree, a conflicting target path, or another unfinished handoff for the same lane, the handoff SHALL return a reason-specific NACK without creating a worktree, relocating the endpoint, or releasing a dispatch. When a fresh reservation follows a `NACK`ed handoff for the same card while the lane's cwd is already that card's target worktree, the handoff SHALL treat that cwd and target as neither untrusted nor conflicting and SHALL admit the reservation into `WT_READY` reconstruction without a second materializer call only if the target is clean, on the reserved `WT-*` branch, and at the fresh develop pin; otherwise it SHALL NACK with `TARGET_DIRTY` when the target worktree has uncommitted or untracked changes, `BRANCH_COLLISION` when the target is not on the reserved `WT-*` branch, or `BASE_DRIFT` when the target HEAD differs from the fresh develop pin, checked in that order.

### REQ-FLH-004 — Develop pin and creation-base integrity

When a lane reserves a handoff, the handoff SHALL record the canonical primary checkout's local `develop` HEAD as an immutable base pin before using the existing MoAI launcher L1 materializer to create `<primary>/.claude/worktrees/<card-id>`. When the created worktree HEAD differs from the pin or the base moves during creation, the handoff SHALL NACK with `BASE_DRIFT` and emit neither `BOUND` nor a dispatch body.

### REQ-FLH-005 — Card, branch, commit, and report traceability

While a handoff is in progress, the handoff SHALL require the card ID in the worktree directory, a collision-free `WT-<descriptive-slug>` branch, the card and SPEC IDs in the dispatch envelope, the card ID in future commit subjects, and `.moai/reports/<card-id>/verdict.md` as the evidence path. It SHALL NOT enter `SWITCH_PENDING` before branch rename and collision validation complete.

### REQ-FLH-006 — Interactive relocation boundary

While an interactive lane is idle and has no permission request, the handoff SHALL emit only a user-executed `/cd <target>` instruction containing the absolute target and reservation nonce and remain `SWITCH_PENDING_INTERACTIVE` until SessionStart evidence arrives from the user's next normal turn. It SHALL NOT assume a hook or MCP tool can execute a slash command, use tmux `send-keys`, a private TUI socket, or a model prompt to change cwd, or create an empty model turn to obtain evidence.

### REQ-FLH-007 — Headless official relocation

While a headless source thread is idle, the handoff SHALL prefer a history-preserving `thread/fork` with the reserved target `cwd` through the existing app-server client and verify the returned new `thread.id`, `forkedFromId`, and `thread/started`. It SHALL use `thread/start(cwd)` only when no source thread exists, SHALL NACK relocation during an active turn, and SHALL NOT wait for SessionStart, issue an empty `turn/start`, or use `turn/steer` to manufacture binding evidence.

### REQ-FLH-008 — Atomic validated rebind and BOUND receipt

When mode-specific endpoint evidence arrives, the handoff SHALL validate the reservation nonce, card and SPEC, canonical target cwd, exact worktree root, pinned HEAD, current branch, new session/thread UUID, new generation, PID, and process-start identity. Interactive evidence SHALL be SessionStart from the user's next normal turn after `/cd`; headless evidence SHALL be the official `thread/fork` or `thread/start` result returned to the controller together with controller readback of target cwd/branch/HEAD. Only after every validation succeeds SHALL one broker transaction replace the current endpoint, tombstone the old endpoint, mark the handoff `BOUND`, and record a durable `BOUND` receipt.

### REQ-FLH-009 — Dispatch ordering and idempotency

While a handoff is not `BOUND`, the factory broker SHALL preserve dispatch metadata for the lane but SHALL deny body claim/read and implementation side effects. When the matching durable `BOUND` receipt is observed, the broker SHALL release the dispatch body exactly once to the current generation. The broker SHALL handle a duplicate dispatch within the same recipient generation without duplicate execution, detecting it by the broker's idempotency key; same-key duplicate handling is required only within one recipient generation. This handoff does not change the idempotency basis itself; the basis is owned by t1100 (SPEC-DUAL-HARNESS-RECOVERY-001). When the sender resends a dispatch after `BOUND` to the rebound endpoint, the sender SHALL use a new idempotency key, whichever basis is in force. The handoff generation SHALL NOT be part of the idempotency key, and a redispatch addressed to a stale generation SHALL be NACKed.

### REQ-FLH-010 — Tombstone and stale traffic rejection

When an old endpoint, stale generation, stale reservation token, or pre-handoff claim token attempts send, read, receipt, or ACK, the broker SHALL reject it with `STALE_ENDPOINT` or `STALE_GENERATION` and return non-secret redirect metadata naming the current lane endpoint and generation. The tombstone SHALL survive restart. A tombstoned session UUID SHALL be refused permanently, including a later resume of the old session, because an endpoint that has once been relocated must never be revived as a second writer on the same lane. This refusal SHALL bind the launcher resume path as well: when the t1074 SessionStart bind (`BindLaunchPending`) presents a session UUID that has any tombstone, the bind SHALL read the tombstone set inside its own write transaction before it changes any row, SHALL match on the session UUID alone regardless of generation, and SHALL refuse with `STALE_ENDPOINT` carrying the same non-secret current-endpoint redirect while writing nothing. The launch-pending row that the preceding launcher registration committed SHALL then stay exactly as committed (private token, generation, PID, process-start) until its owner is not current under the t1074 live-owner rule, after which a fresh launcher registration replaces it; a turn registration of the same session SHALL also be refused with `STALE_ENDPOINT`, so the outcome does not depend on whether SessionStart or UserPromptSubmit fires first. The SessionStart hook SHALL surface this refusal as its existing endpoint-replaced (`STALE_ENDPOINT`) notice in the session's additional context, not as a degraded-messaging line.

### REQ-FLH-011 — Crash, restart, and abandoned worktree recovery

When a crash occurs before or after creation, branch rename, `SWITCH_PENDING`, the rebind transaction, or receipt delivery, recovery SHALL reread the stored nonce and filesystem, Git, and broker facts and select exactly one of resume, idempotent finalize, or `ABANDONED`. It SHALL preserve dirty or unmerged worktrees without automatic deletion and SHALL NOT guess past a live foreign owner or uncertain endpoint. When an operator runs `moai factory handoff abandon-lane --slot <slot>` (a verb on the existing `moai factory handoff` command group, using the existing factory broker store and adding no daemon, broker, or store) for a lane whose handoff is non-final, the command SHALL first verify with the t1074 PID and process-start liveness rule that the handoff's recorded source owner is not current, and SHALL refuse with `SOURCE_OWNER_LIVE` and write nothing when that owner is current or its liveness cannot be established. When that verification passes, the command SHALL write the handoff `ABANDONED` with reason `OPERATOR_ABANDONED` in one broker write transaction that rereads the handoff state and the source owner, SHALL leave the worktree, its branch, and its commits in place, and SHALL write no `BOUND` state, no tombstone, no BOUND receipt, and no dispatch release; after that commit REQ-FLH-018 no longer applies to the lane and t1074 UserPromptSubmit registration semantics return. When the lane has no non-final handoff, the command SHALL write nothing and report `HANDOFF_NOT_PENDING`. The sync-phase user documentation SHALL name `moai factory handoff abandon-lane --slot <slot>` as the manual recovery command for a lane stuck behind a non-final handoff.

### REQ-FLH-012 — Authority and product-boundary truth

While implementation handles a handoff, it SHALL NOT assume hook or MCP authority to execute slash commands, and SHALL NOT describe ChatGPT Desktop Handoff or a Codex-managed detached worktree as a headless MoAI CLI capability. The card worktree SHALL be a MoAI-launcher-managed worktree on a named `WT-*` branch, distinct from a Desktop-managed worktree.

### REQ-FLH-013 — Zero-write safety before BOUND

While a handoff is not `BOUND`, the lane SHALL perform no code edit, commit, card completion, or task ACK. Acceptance evidence SHALL separately count and prove zero pre-BOUND code edits, zero wrong-cwd edits, zero primary-checkout branch switches, and zero messages lost during endpoint replacement.

### REQ-FLH-014 — Live cross-harness proof

When LIVE verification runs, it SHALL exercise an actual interactive `/cd` path and an actual stored-history headless app-server `thread/fork(cwd)` path across Codex lead↔Codex lane and Claude lead↔Codex lane in real separate model/CLI contexts and observe post-rebind bidirectional unique nonces, generation-bound explicit receipts, statusline, cwd, branch, and old-endpoint rejection. The interactive proof SHALL use the user's next normal turn and record zero empty model turns; the headless proof SHALL require method `thread/fork`, a nonempty returned thread ID, a nonempty `forkedFromId`, and `thread/started=true`, then bind from that official result without waiting for SessionStart. Each proof SHALL write versioned card-scoped structured evidence whose process identity, mode transition, target provenance, receipts, message counters, cleanup, and bypass-denial fields are validated by exact typed predicates rather than self-reported stdout strings. A gate-quality test SHALL prove that missing-field, mock, fixture, direct-registration, child-failure, child-skip, and stored-history `wrong_method_thread_start` mutants are rejected. No-history `thread/start(cwd)` remains valid only for the unit contract in REQ-FLH-007/AC-FLH-004 and SHALL NOT satisfy the stored-history LIVE proof. A child/subtest/package failure, skip, mock, fixture-only path, direct registration, manual DB seed, private control, malformed or missing evidence, or `NOT_RUN` SHALL fail the affected LIVE criterion.

### REQ-FLH-015 — Reuse and compatibility

While implementation handles a handoff, it SHALL reuse t1074 canonical run selection, stable lane/current endpoint generation, launcher provisional endpoint bound by the first legitimate non-empty UserPromptSubmit (mandatory) or an earlier order-independent SessionStart (best-effort) per SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-001, the factory message broker, peer roster, explicit receipts, existing homestate, the MoAI worktree launcher, and the app-server client. Headless controller binding SHALL coexist with that interactive first-turn contract without adding a broker, daemon, private transport, or parallel message store or changing legacy `sessionmsg` semantics.

### REQ-FLH-016 — Launch-pending source endpoint admission

When a handoff is requested for a lane whose current endpoint is still the t1074 launcher provisional (launch-pending) endpoint, the handoff SHALL return NACK with reason `ENDPOINT_LAUNCH_PENDING`, SHALL write no reservation, tombstone, worktree, app-server request, endpoint mutation, or dispatch release, and SHALL leave the provisional endpoint neither bound nor rolled back. The handoff SHALL NOT wait for, poll, retry against, or manufacture the lane's first normal turn; a later handoff SHALL require a fresh reservation after the lane's endpoint is bound.

### REQ-FLH-017 — Serialized ordering of launcher bind and handoff rebind

The handoff SHALL order the t1074 launcher provisional-endpoint bind before handoff admission, and SHALL serialize its atomic rebind against every launcher provisional registration or bind on the same lane as write transactions within the existing factory broker's SQLite transaction domain on that lane's single endpoint row. While a handoff is non-final (`RESERVED`, `WT_READY`, `SWITCH_PENDING_INTERACTIVE`, or `SWITCH_PENDING_HEADLESS`), the lane endpoint row SHALL equal the reserved source endpoint (session/thread UUID, generation, PID, process-start), and any committed change to that row other than the handoff rebind SHALL finalize the handoff in the same transaction. When a t1074 launcher provisional registration commits on a lane whose handoff is non-final, that same registration transaction SHALL move the handoff to `NACK` with reason `STALE_GENERATION` and SHALL write no tombstone, BOUND receipt, or dispatch release; a registration rejected by the t1074 live-owner rule SHALL leave the handoff unchanged. When the rebind runs after such a registration, it SHALL return `STALE_GENERATION`, write nothing, and leave the launcher-owned endpoint current, which the t1074 binder then binds. While the handoff-bound owner is current under the t1074 live-owner rule, when the rebind has committed first, a later launcher registration or bind SHALL leave the handoff-bound endpoint unchanged; once that owner is not current, the t1074 dead-owner restart re-registers the slot unchanged by this SPEC, except that a SessionStart bind carrying a tombstoned session UUID is refused under REQ-FLH-010. In every interleaving the lane SHALL end, once the t1074 binder has run, with exactly one current endpoint, which is bound, a generation that never decreases, no orphan launch-pending endpoint, and no non-final handoff whose reserved source tuple differs from the row. One exception applies to both the bound clause and the orphan clause alike: a launch-pending row whose live launcher owner was refused a bind under REQ-FLH-010 remains the lane's single current endpoint, unbound and held by that owner rather than orphaned, until that owner is not current under the t1074 live-owner rule, after which a fresh launcher registration replaces it. The handoff SHALL NOT create its new endpoint through the launcher provisional registration path.

### REQ-FLH-018 — UserPromptSubmit peer registration during an unfinished handoff

While a lane has a handoff in a non-final state (`RESERVED`, `WT_READY`, `SWITCH_PENDING_INTERACTIVE`, or `SWITCH_PENDING_HEADLESS`), the t1074 UserPromptSubmit peer registration for that lane's slot SHALL be rejected with `ENDPOINT_HANDOFF_PENDING` inside the same broker write transaction that reads the handoff state, SHALL leave the endpoint row, generation, tombstones, receipts, and dispatch release markers unchanged, and SHALL NOT move the endpoint even when the caller's PID and process-start identity equal the current owner's and only the session UUID differs. When the handoff rebind has committed first, a later UserPromptSubmit registration carrying the handoff-bound endpoint identity SHALL leave that endpoint and its generation unchanged, and one carrying the tombstoned source session UUID SHALL be rejected with `STALE_ENDPOINT`. When the handoff reaches `NACK` or `ABANDONED`, UserPromptSubmit registration SHALL resume t1074 semantics without releasing any dispatch body. This rejection SHALL NOT strand a lane restarted through the launcher: that launcher's provisional registration finalizes the handoff first (REQ-FLH-017), after which the handoff is final, this requirement no longer applies, and the t1074 UserPromptSubmit binding proceeds. The rejection SHALL be enforced within the existing factory broker and hook path without adding a broker, daemon, or store.

## Requirement-to-acceptance traceability

| Requirement anchor | Acceptance criteria |
|---|---|
| § REQ-FLH-001 | AC-FLH-001, AC-FLH-007 |
| § REQ-FLH-002 | AC-FLH-003, AC-FLH-004 |
| § REQ-FLH-003 | AC-FLH-001 |
| § REQ-FLH-004 | AC-FLH-002, AC-FLH-016 |
| § REQ-FLH-005 | AC-FLH-002 |
| § REQ-FLH-006 | AC-FLH-003, AC-FLH-014 |
| § REQ-FLH-007 | AC-FLH-004, AC-FLH-014 |
| § REQ-FLH-008 | AC-FLH-005, AC-FLH-011 |
| § REQ-FLH-009 | AC-FLH-006, AC-FLH-008, AC-FLH-011 |
| § REQ-FLH-010 | AC-FLH-003, AC-FLH-007 |
| § REQ-FLH-011 | AC-FLH-009, AC-FLH-010, AC-FLH-020 |
| § REQ-FLH-012 | AC-FLH-014 |
| § REQ-FLH-013 | AC-FLH-011, AC-FLH-012, AC-FLH-013 |
| § REQ-FLH-014 | AC-FLH-012, AC-FLH-013 |
| § REQ-FLH-015 | AC-FLH-015, AC-FLH-018 |
| § REQ-FLH-016 | AC-FLH-017, AC-FLH-019 |
| § REQ-FLH-017 | AC-FLH-018, AC-FLH-019 |
| § REQ-FLH-018 | AC-FLH-019 |

### Out of Scope — Idle wake and autonomous TUI control

- BOUND 이후 idle session을 깨우는 기능은 t1075 범위다.
- tmux 키 주입, private socket, TUI 내부 API를 통한 무인 조작은 구현하지 않는다.

### Out of Scope — Desktop Handoff emulation

- ChatGPT Desktop의 Handoff UI와 Git 이동 동작을 MoAI CLI에서 복제하지 않는다.
- Codex-managed detached worktree를 card worktree로 채택하지 않는다.

### Out of Scope — Queue ownership and completion

- 이 handoff는 card 선택, queue 완료, PR/merge, deploy 권한을 부여하지 않는다.
- `BOUND` receipt는 transport/work-root 결합 증거이며 구현 완료 증거가 아니다.

### Out of Scope — Timeout-based automatic handoff finalization

- 비종결 handoff를 시간 경과만으로 자동 종결하지 않는다. 시간 기준 종결에는 주기적으로 도는 행위자가 필요한데 이 SPEC은 새 daemon·polling service를 두지 않으며, interactive `SWITCH_PENDING`은 사용자의 `/cd`와 다음 정상 turn을 기다리는 사람 속도의 대기라 어떤 상한을 정해도 정상적으로 느린 lane을 끊을 수 있다. 막힌 lane은 REQ-FLH-011의 operator 명령 `moai factory handoff abandon-lane --slot <slot>`로 종결한다.

### Out of Scope — New transport infrastructure

- 새 message broker, daemon, polling service, private IPC를 만들지 않는다.
- t1074 broker와 legacy `sessionmsg` 사이를 자동 이관하지 않는다.
