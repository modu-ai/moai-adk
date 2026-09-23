---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: plan
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Plan — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## Assumptions and resolved decisions

| Assumption | Confidence | Evidence basis | Risk if wrong | Validation |
|---|---|---|---|---|
| t1074의 stable lane/current endpoint seam이 handoff 주소로 재사용 가능하다. | High | `factorymsg.Peer`, `ResolveLane`, generation-bound receipt가 현재 tree에 존재한다. | 별도 주소체계가 생겨 split-brain이 된다. | AC-FLH-015에서 기존 broker/roster regression과 단일 store를 확인한다. |
| interactive `/cd`는 사용자 동작이며 자동화 API가 아니다. | High | 공식 app-server에는 slash-command 실행 API가 없고 로컬 관측만 UUID 회전을 보였다. | hook/MCP 권한을 과장하고 TUI를 비공개 방식으로 조작한다. | AC-FLH-003/014에서 private-control 호출 0을 검증한다. |
| headless relocation은 app-server의 idle fork/start가 반환한 공식 thread ID와 controller의 target cwd/branch/HEAD readback으로 닫을 수 있다. | High | 공식 `thread/fork`, `thread/start`, `thread/started` 계약과 initialize+`thread/start`-only 로컬 관측. | SessionStart를 기다리거나 빈 turn을 만들어 비용과 상태를 왜곡한다. | AC-FLH-004와 headless LIVE AC에서 lineage/cwd/direct BOUND를 확인한다. |
| interactive `/cd` 뒤 결합 증거는 사용자의 다음 정상 turn에서 온 SessionStart여야 한다. | High | initialize+`thread/start`-only 관측에서는 5초 내 SessionStart sidecar/factory DB가 없었다. | pre-turn SessionStart를 보장으로 오인하거나 빈 turn을 만들게 된다. | AC-FLH-003과 interactive LIVE AC에서 pending 유지와 empty-turn 0을 확인한다. |
| BOUND/rebind/tombstone/dispatch release는 같은 broker DB transaction이어야 한다. | High | 다른 DB에 나누면 crash window에서 current endpoint와 release 상태가 갈라진다. | 이중 endpoint 또는 body 조기 전달이 생긴다. | AC-FLH-005/006/009의 fault injection으로 검증한다. |

미해결 질문은 없다. Plan→run 전에는 별도 plan-auditor PASS와 사용자 Implementation Kickoff Approval이 필요하다.

## Constraints

- 구현은 이 plan-phase에서 시작하지 않는다.
- primary checkout branch를 바꾸지 않는다.
- L1 worktree 생성은 기존 MoAI launcher materializer만 호출한다.
- `develop`은 local ref를 pin하며, 생성 직후 exact HEAD equality를 확인한다.
- dirty source, active turn, permission wait, interrupt, untrusted cwd, branch/path collision, base drift는 모두 NACK다.
- interactive와 headless transition은 별도 함수/오류/테스트로 유지한다.
- 새 MCP tool은 기본 설계에 포함하지 않는다. 기존 factory message surfaces와 launcher/app-server seams를 조합한다.

## Milestones

### M1 — Durable reservation and worktree provenance

- `internal/factorymsg`의 기존 broker schema에 lane handoff row/event/tombstone을 추가한다. 별도 DB는 만들지 않는다.
- reservation은 project/run/lane/card/SPEC/source endpoint/generation/nonce/local develop HEAD/target path/target branch를 CAS로 고정한다.
- (REQ-FLH-016) reservation admission은 source 행 판독, launch-pending 판정, `RESERVED` 삽입을 하나의 broker write transaction(`BEGIN IMMEDIATE`) 안에서 수행한다. source endpoint가 t1074 launch-pending이면 `ENDPOINT_LAUNCH_PENDING` NACK를 반환하고 reservation·tombstone·worktree·app-server 요청·endpoint 변경·dispatch release를 하나도 쓰지 않는다. 판독 위치는 AC-FLH-019 순서 (vii)이 판별하며, 그 순서가 run phase의 첫 RED다.
- canonical project identity는 `internal/homestate.CanonicalProjectRoot`로, canonical run은 기존 factory run resolver `internal/cli/factory.go:222` `enterSelectedFactoryRun`(→ `factorymsg.ResolveActiveRun`)으로 구한다. Run resolver는 homestate에 있지 않다.
- 기존 MoAI L1 worktree materializer를 호출하고 exact target path, `HEAD == pinned develop`, clean target, `WT-<slug>` branch, branch uniqueness를 읽어 `WT_READY`로 만든다.
- creation-base drift를 `BASE_DRIFT`로 기록하고 BOUND 없이 보존/복구 대상으로 남긴다.

### M2 — Interactive and headless relocation adapters

- interactive adapter는 `IDLE`이 확인된 lane에 `/cd <absolute-target>` 안내와 nonce만 발행하고 `SWITCH_PENDING_INTERACTIVE`로 전이한다.
- headless adapter는 active turn/permission wait/interrupt를 먼저 거부한다. history가 있으면 target `cwd`를 지정한 `thread/fork`, 없으면 `thread/start(cwd)`를 사용하고, 공식 응답의 새 thread ID, `thread/started`, lineage를 기록한다.
- controller는 target cwd/branch/HEAD를 직접 readback하고 headless relocation evidence(새 thread ID, `thread/started`, lineage)를 기록한다. BOUND는 M2에서 쓰지 않는다 — BOUND는 source-tuple CAS·tombstone·peer swap·BOUND 상태·BOUND receipt·dispatch release marker를 한 transaction으로 묶는 M3 atomic rebind의 산출물이다(REQ-FLH-008, design.md §6). SessionStart, 빈 `turn/start`, `turn/steer`는 relocation evidence로 사용하지 않는다.
- M2는 두 adapter를 각각 `SWITCH_PENDING_INTERACTIVE` / `SWITCH_PENDING_HEADLESS`까지 전이시키고, 기존 app-server client에 `thread/fork`를 추가한다. AC-FLH-014는 M2에서 닫는다. M2의 adapter 단위 테스트는 AC-FLH-003/004의 named test와 다른 이름을 쓴다.
- 기존 `mcp_codex.go` JSON-RPC transport, request/notification parser, bounded process cleanup을 재사용한다.

### M3 — Mode-evidence atomic rebind and dispatch release

- interactive는 사용자의 다음 정상 turn에서 온 SessionStart를, headless는 공식 fork/start 결과와 controller readback을 받아 canonical cwd, worktree root, HEAD, branch, nonce, new UUID, PID/process-start를 검증한다.
- 같은 broker transaction에서 old peer tombstone, new peer generation, handoff `BOUND`, BOUND receipt, dispatch release marker를 기록한다.
- headless는 M2가 기록한 relocation evidence와 controller readback을 받아 이 transaction으로 원자적으로 BOUND한다. AC-FLH-003의 `TestFactoryLaneHandoffInteractiveStateMachine`과 AC-FLH-004의 `TestFactoryLaneHandoffHeadlessAppServerStateMachine`은 BOUND 관측을 요구하므로 rebind와 함께 M3에서 작성하고 PASS시킨다.
- BOUND 전 body claim/read/ACK와 code-write authorization을 거부한다.
- stale send/ACK는 현재 endpoint/generation metadata를 포함한 NACK로 응답한다.
- duplicate dispatch는 broker의 idempotency key로 판정하되 같은 recipient generation 안에서만 멱등 처리한다. t1082는 idempotency 기준 자체를 바꾸지 않으며, 기준은 t1100(SPEC-DUAL-HARNESS-RECOVERY-001)이 소유한다(design.md §8). BOUND 뒤 rebound endpoint로의 재전송은 기준과 무관하게 새 key를 쓴다(같은 key를 다른 recipient로 다시 쓴 요청은 기존 봉투로 합쳐지지 않는다). handoff generation은 key에 넣지 않으며, 이전 generation으로 향한 redispatch는 stale NACK다. AC-FLH-008 test는 key 기준에 기대는 단언을 두지 않는다.
- (REQ-FLH-017) 착지된 t1074 `Store.RegisterPeer`(`internal/factorymsg/store.go:315`; `RegisterLaunchPending` :391이 이 함수로 쓴다)의 write transaction 안에 같은 lane의 handoff 상태 판독을 추가한다. 비종결 handoff가 있는 lane에서 launcher provisional 등록이 commit되면 같은 transaction에서 handoff를 `NACK`/`STALE_GENERATION`으로 종결하고 tombstone·BOUND receipt·dispatch release는 쓰지 않는다. t1074 live-owner 규칙으로 거부된 등록은 handoff를 바꾸지 않는다. rebind는 같은 SQLite transaction 영역에서 등록과 직렬화하고, 등록 뒤에 실행되면 `STALE_GENERATION`을 반환하며 아무것도 쓰지 않는다. 검증: AC-FLH-018, AC-FLH-019 (v)·(vii).
- (REQ-FLH-018) 같은 `RegisterPeer` transaction 안에서, 비종결 handoff가 있는 lane의 UserPromptSubmit 등록(`registerFactoryUserPromptPeer`, `internal/hook/factory_messages.go:49` → `registerFactoryHookPeer` :53)을 `ENDPOINT_HANDOFF_PENDING`으로 거부하고 endpoint 행·generation·tombstone·receipt·release marker를 바꾸지 않는다. rebind가 먼저 commit된 뒤 tombstone된 source session UUID로 오는 등록은 `STALE_ENDPOINT`로 거부한다. 검증: AC-FLH-019 (i)–(iv).

### M4 — Recovery, safety, and compatibility

- crash points: reserve 후, create 전/후, rename 전/후, SWITCH_PENDING 후, rebind transaction 전/후, receipt 전/후를 fault injection한다.
- restart reconciler는 DB와 filesystem/git/app-server facts를 읽고 resume/finalize/ABANDONED만 선택한다.
- dirty/unmerged/owner-unknown worktree는 자동 삭제하지 않는다.
- REQ-FLH-010 launcher resume 경로 차단(리드 결정 D1=(b)): `BindLaunchPending`(`internal/factorymsg/store.go:446`)의 write transaction 시작 직후, 행 판독·변경 전에 `sessionTombstoned`(`internal/factorymsg/handoff_bind.go:500`)로 session UUID만의 tombstone 판독을 넣고, tombstone이 있으면 `STALE_ENDPOINT`(현재 endpoint redirect 포함)로 거부하며 무쓰기로 끝낸다. `staleOrUnregistered`의 `(session_uuid, generation)` 쌍 매칭은 쓰지 않는다. SessionStart hook(`internal/hook/factory_messages.go:53` `registerFactoryHookPeer`)은 이 거부를 `factoryHandoffRegistrationNotice`의 endpoint-replaced 안내문으로 싣는다. AC-FLH-007의 launcher resume 다리와 양성 다리를 named test `TestFactoryLaneHandoffStaleEndpointRejected`에 먼저 써서 RED(tombstone된 UUID가 결합되어 현재 endpoint가 됨)를 확인한 뒤 고친다. 이 변경은 닫힌 카드 t1074가 소유한 `internal/factorymsg/store.go`를 건드리며, 그 비용은 리드 결정으로 받아들였다. `BindLaunchPending`을 부르는 기존 테스트(`internal/factorymsg/handoff_test.go`, `launch_pending_rollback_test.go`, `internal/hook/factory_handoff_race_test.go`)를 함께 재실행한다.
- operator 종결 경로 `moai factory handoff abandon-lane --slot <slot>`을 기존 `moai factory handoff` 명령 그룹(`internal/cli/factory_handoff_recover.go`)에 붙이고 종결 transaction은 `internal/factorymsg`에 둔다(REQ-FLH-011). source owner가 current이거나 판정 불가면 `SOURCE_OWNER_LIVE` 거부, 아니면 한 transaction에서 `ABANDONED`/`OPERATOR_ABANDONED`를 쓰고 worktree 보존·BOUND/tombstone/receipt/release 0·종결 뒤 t1074 UserPromptSubmit 복귀를 AC-FLH-020 named test로 먼저 RED 확인한다. source owner 판정 seam은 종결 store 함수가 인자로 받는 `func(int) (string, homestate.ProcessIdentityState)` probe 하나다(선례 `internal/cli/factory_handoff_recover.go:30`의 `RecoverLegacyResume(..., homestate.ProbeProcessIdentity, ...)`). cli 명령은 이를 패키지 변수 하나로 연결하고, AC-FLH-020 테스트는 그 변수만 바꿔 Live/Dead/Indeterminate를 만든다. 3상태를 bool로 접는 `ownerCurrent`는 쓰지 않는다. 시간 기준 자동 종결은 만들지 않는다. sync phase 문서는 이 명령을 막힌 lane의 수동 복구 명령으로 적는다.
- t1074 factory broker/roster/receipt/SessionStart tests와 MCP catalog invariant를 재실행한다.

### M5 — Real mixed-factory verification

- built-tree `moai`와 실제 별도 model contexts로 Codex↔Codex 및 Claude lead↔Codex를 실행한다.
- interactive 행은 실제 `/cd`와 다음 정상 turn SessionStart 및 empty-turn count 0을, headless 행은 실제 `thread/fork(cwd)` 반환 ID와 SessionStart 대기 0을 강제한다.
- 각 행에서 pre-BOUND write count 0, target statusline/cwd/branch, old/new endpoint, generation, nonce round-trip, explicit receipt, message count를 수집한다.
- 각 LIVE test는 `.moai/reports/t1082/ac12-evidence.json` 또는 `ac13-evidence.json`을 원자적으로 기록한다. JSON에는 real/separate production contexts의 argv/PID/process-start/binary hash, mode별 transition/RPC 사실, reserved/observed cwd·branch·HEAD, receipt/stale rejection, zero-write/message counters, distinct nonces, cleanup, no-bypass fields를 typed values로 담는다.
- 공통 gate-quality selector는 필수 필드 누락, fixture/mock/direct-registration, child fail, child skip, 그리고 stored-history headless evidence를 `thread/start`/`forked_from_id=null`로 바꾼 `wrong_method_thread_start` mutant를 동일 production predicate에 넣어 전부 거부되는지 확인한다. No-history `thread/start(cwd)` 허용은 AC-FLH-004 unit 경로에만 남기며 AC-FLH-013 LIVE predicate에는 허용하지 않는다. LIVE stdout의 자유 형식 문자열은 판정 자료로 쓰지 않는다.
- process cleanup을 외부 timeout과 test cleanup으로 보장한다.
- 어느 package/child/subtest든 `fail` 또는 `skip`이 있거나 `NOT_RUN`, missing/malformed evidence가 한 행이라도 있으면 LIVE 전체를 FAIL로 남긴다.

## Intended file ownership for run phase

정확한 파일 목록은 RED test 작성 뒤 최소화한다. 예상 책임 경계는 다음과 같다.

| Area | Reuse/extension intent |
|---|---|
| `internal/factorymsg/` | 기존 broker transaction, peer generation, receipt에 handoff state/rebind/tombstone 추가; 같은 transaction 안에서 source 행을 읽는 reservation admission과 `ENDPOINT_LAUNCH_PENDING` NACK(REQ-FLH-016); 착지된 t1074 `Store.RegisterPeer` transaction 안의 handoff 상태 판독·launcher 등록 시 종결(REQ-FLH-017)·UserPromptSubmit 등록 거부(REQ-FLH-018) |
| `internal/homestate/` | canonical project/process identity helper 재사용; 새 transport 금지 (run selection은 `internal/cli/factory.go:222` `enterSelectedFactoryRun`) |
| `internal/cli/worktree/`, launcher seams | existing L1 materializer 호출 및 develop pin/branch trace 검증 |
| `internal/cli/mcp_codex.go` 주변 | existing app-server client에 fork/start/cwd lifecycle 최소 확장 |
| `internal/hook/` | interactive의 다음 정상 turn SessionStart evidence를 shared atomic rebind에 전달; UserPromptSubmit 등록 경로가 `RegisterPeer`의 `ENDPOINT_HANDOFF_PENDING`/`STALE_ENDPOINT` 거부를 받는 지점(REQ-FLH-018)과 AC-FLH-018/019의 package `hook` 경합 test |
| `internal/cli/*_live_test.go` | cleanup-guaranteed real mixed-factory LIVE harness |

## Verification strategy

1. 각 AC의 exact named test를 먼저 작성해 RED를 확인한다.
2. 단위 테스트는 Go JSON event와 `jq -se`로 정확히 한 named PASS, 전체 child/subtest/package fail 0, 전체 skip 0, `NOT_RUN` 0을 요구한다.
3. LIVE는 provider credential을 command 안에서 scrub하되 실제 설치 인증 context를 사용하는 기존 t1074 harness 규칙을 따르고, 별도 `jq -e`로 card-scoped evidence의 exact typed schema와 cross-field equality를 검증한다.
4. 두 LIVE gate는 `TestFactoryLaneHandoffLiveEvidenceGateRejectsMutants` PASS도 요구한다. 이 test는 acceptance의 production predicate를 직접 호출해 missing/fixture/mock/direct-registration/child-fail/child-skip 및 stored-history `wrong_method_thread_start` mutant가 모두 거부됨을 증명한다.
5. 변경 범위 unit/race/vet 후 t1074 regression과 MCP catalog 36/14/22 invariant를 실행한다.
6. 전체 suite는 로컬 loaded-machine proof로 대체하지 않고 push 후 CI에 맡긴다.

## Rollback and failure handling

- `RESERVED`/`WT_READY`/`SWITCH_PENDING`에서 실패하면 current endpoint는 바꾸지 않는다.
- BOUND transaction이 commit되지 않았으면 old endpoint가 current다. commit됐다면 durable receipt/outbox로 재발행만 허용하고 rebind를 반복하지 않는다.
- ABANDONED worktree는 경로, branch, dirty/unmerged, owner 상태와 사유를 남기며 자동 제거하지 않는다.
- rollback은 schema backward read와 feature disable을 제공하되 BOUND tombstone을 지워 stale endpoint를 부활시키지 않는다.

## Anti-patterns prohibited by this plan

- hook/MCP가 `/cd`를 실행한다고 가정하기.
- tmux `send-keys`, private socket, model prompt로 cwd를 바꾸기.
- Desktop Handoff UI를 headless MoAI capability로 간주하기.
- raw `git worktree add` 또는 primary branch switch 경로 추가하기.
- rebind와 dispatch release를 서로 다른 DB transaction에 나누기.
- 새 broker/daemon을 만들어 t1074 상태와 이중화하기.
- test 부재, skip, `NOT_RUN`, fixture-only result를 LIVE PASS로 처리하기.
- LIVE stdout의 자기보고 문자열만 grep해서 structured evidence의 필수 사실을 대체하기.
