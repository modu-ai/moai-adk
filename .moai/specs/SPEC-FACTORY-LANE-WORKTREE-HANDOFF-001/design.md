---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: design
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Design — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## 1. Component boundaries

```text
factory lead dispatch
        │ reserve(card, stable lane)
        ▼
existing factorymsg broker ────────────────┐
  handoff row + peer generation + messages │ one durable authority
        │                                   │
        ├─ pin local develop HEAD           │
        ├─ call existing MoAI L1 materializer
        └─ verify path/HEAD/WT-branch       │
                    │                       │
              SWITCH_PENDING               │
               ┌────┴────┐                  │
               │         │                  │
      interactive      headless             │
      user `/cd`       app-server            │
               │         │                  │
               └────┬────┘                  │
        mode-specific endpoint evidence     │
                    │                       │
         validate cwd/branch/HEAD/nonce     │
                    ▼                       │
        atomic peer rebind + tombstone + BOUND receipt
                    │
                    ▼
           release dispatch body
```

새 broker나 daemon은 없다. Handoff row와 tombstone은 기존 factory message broker의 peer/message와 같은 SQLite transaction domain에 둔다. `internal/homestate`는 canonical path/run/process identity를 제공하고, worktree creation은 기존 launcher materializer가 담당한다.

## 2. Identity model

```text
Stable address:
  {project_key, run_id, lane_id}

Handoff generation:
  {card_id, spec_id, reservation_nonce, handoff_generation}

Physical endpoint:
  {backend, thread_or_session_uuid, peer_generation, pid, process_start}

Work-root provenance:
  {canonical_primary, develop_pin, worktree_path, branch, head_sha}
```

`lane_id`는 lead가 계속 메시지를 보내는 주소다. `thread_or_session_uuid`는 endpoint일 뿐이며 `/cd`, fork, restart 때 바뀔 수 있다. Old endpoint tombstone은 `{lane_id, old_uuid, old_generation, replaced_by_uuid, replaced_by_generation, bound_at}`를 보존한다.

### 2.1 Endpoint binding order with t1074 launch-pending (REQ-FLH-016, REQ-FLH-017, REQ-FLH-018)

t1074 착지본의 broker는 slot당 endpoint row 하나(`peers.slot` PRIMARY KEY)를 두고, launcher가 먼저 private `launch-pending:` 키로 provisional row를 등록(`RegisterLaunchPending`, `internal/factorymsg/store.go:381`)한 뒤 첫 정상 turn의 SessionStart가 exact owner identity로 그 row를 bind한다(`BindLaunchPending`, `store.go:392`; 호출부 `internal/hook/factory_messages.go:84`). Provisional row가 current인 동안 `ResolveLane`·`Peer`·`PeerByOwner`·송신 검증은 모두 `ErrEndpointLaunchPending`을 돌려준다(`store.go:485,506,521,564`).

착지 코드(HEAD `d28ed9aa4`)에서 같은 slot 행에 쓰는 비-테스트 경로는 네 개이고, 이 SPEC의 handoff CAS rebind가 다섯 번째로 더해진다.

| # | Writer | Store entry point | 호출부 | 행에 대한 효과 |
|---|---|---|---|---|
| 1 | launcher provisional registration | `RegisterLaunchPending` (`store.go:381`, 내부적으로 `RegisterPeer`) | `internal/cli/factory_launch_pending.go:58` | `launch-pending:` 키로 row 생성 또는 교체 |
| 2 | launcher rollback | `RollbackLaunchPending` (`store.go:455`) | `internal/cli/factory_launch_pending.go:73` | exact provisional identity일 때만 row 삭제 |
| 3 | SessionStart bind | `BindLaunchPending` (`store.go:392`) | `internal/hook/factory_messages.go:84` | pending row만 CAS로 bound 전환, 그 외 no-op |
| 4 | UserPromptSubmit registration | `RegisterPeer` (`store.go:305`) | `internal/hook/factory_messages.go:93` ← `registerFactoryUserPromptPeer` (`factory_messages.go:38`) ← `internal/hook/user_prompt_submit.go:112` | 살아 있는 같은 owner(PID·process-start 동일)가 새 session UUID를 내면 거절 없이 generation+1로 endpoint 교체 |
| 5 | handoff atomic rebind (이 SPEC) | §6 transaction | handoff controller | reserved source tuple CAS 후 교체 + tombstone + BOUND receipt |

4번 경로의 효과는 interactive `/cd` 직후 turn과 같은 모양이다. 새 session UUID, 같은 Codex process, 그리고 tombstone·receipt 없는 endpoint 교체다.

```text
lane endpoint row lifecycle (one row per slot)

launcher start ──► LAUNCH_PENDING(g) ──first normal-turn SessionStart──► BOUND(g+1)
                   │  handoff reserve → NACK ENDPOINT_LAUNCH_PENDING      │
                   │  (no reservation/tombstone/worktree; row untouched)  │
                                                                          ▼
                                             handoff reserve (source = BOUND(g+1) tuple)
                                                                          │
                                                                  SWITCH_PENDING_*
                                                                          │
              ┌───────────── one BEGIN IMMEDIATE writer at a time ────────┴──────────┐
              ▼                                                                     ▼
   handoff rebind commits first                              launcher re-register/bind commits first
   CAS(source tuple) matches → BOUND(g+2)                    row → LAUNCH_PENDING(g+2) → BOUND(g+3)
   + tombstone + BOUND receipt + release                     handoff rebind CAS mismatch →
   later RegisterLaunchPending → t1074 live-owner reject     NACK STALE_GENERATION, no tombstone/
   later BindLaunchPending → no-op (row not pending)         receipt/release; launcher owner current
```

순서 규칙은 세 가지다. 첫째, launcher bind가 handoff admission보다 먼저다 — launch-pending인 lane에는 reservation 자체가 생기지 않는다. 둘째, reservation 이후의 경합은 "먼저 commit한 writer가 이긴다"이며, handoff rebind는 reserved source tuple(session/thread UUID, generation, PID, process-start)에 대한 CAS이므로 launcher가 row를 먼저 바꿨다면 반드시 `STALE_GENERATION`으로 진다. 셋째, handoff는 자기 새 endpoint를 launcher provisional 경로로 만들지 않는다. 넷째, handoff가 비종결 상태(`RESERVED`, `WT_READY`, `SWITCH_PENDING_*`)인 동안 위 표 4번 UserPromptSubmit registration은 같은 broker write transaction 안에서 handoff 상태를 읽고 `ENDPOINT_HANDOFF_PENDING`으로 거부되며 행을 바꾸지 않는다(REQ-FLH-018). 따라서 비종결 handoff가 있는 slot에서 endpoint를 옮길 수 있는 writer는 launcher 경로(1-3번, REQ-FLH-017이 CAS로 순서를 정함)와 handoff rebind뿐이다.

**UserPromptSubmit 경로의 처리 선택 (REQ-FLH-018).** "먼저 commit하면 tombstone을 남기고 handoff는 `STALE_GENERATION`으로 진다"가 아니라 "비종결 handoff 동안 같은 transaction 안에서 거부"를 택했다. 먼저 commit을 허용하면 검증되지 않은 transaction이 endpoint를 교체하게 되어 REQ-FLH-008의 "모든 검증이 성공한 뒤에만 한 transaction이 교체한다"가 깨지고, tombstone은 REQ-FLH-010이 전제하는 교체의 일부가 아니라 사후 보정이 되며, 그 사이 lane은 BOUND 없이 새 cwd에 결합되어 REQ-FLH-013의 pre-BOUND 무결성을 흐린다. 또한 Codex가 `/cd` 뒤 turn에서 SessionStart와 UserPromptSubmit을 어떤 순서로 발화하는지는 관측되지 않았는데, 거부를 택하면 순서와 무관하게 결과가 같다(UserPromptSubmit이 먼저면 거부되고 SessionStart evidence가 rebind하며, rebind가 먼저면 이후 UserPromptSubmit은 같은 identity라 no-op이다).

교착 방지는 §9 결정표가 맡는다. Handoff가 `NACK` 또는 `ABANDONED`라는 종결 상태에 이르면 UserPromptSubmit registration은 t1074 의미로 복귀하되, BOUND가 아니므로 dispatch body는 계속 막힌다.

직렬화 경계는 새 장치가 아니라 기존 broker의 SQLite transaction domain이다. Broker는 `_txlock=immediate`로 열리므로(`store.go:169,214`) 모든 쓰기 transaction이 `BEGIN IMMEDIATE`로 시작해 프로세스 사이에서도 한 writer만 RESERVED lock을 잡는다. 단 각 `Store` 핸들은 `SetMaxOpenConns(1)`(`store.go:174,219`)이라, 같은 핸들을 공유하는 두 goroutine은 Go connection pool에서 먼저 직렬화되고 `BEGIN IMMEDIATE` 경계에 도달하지 않는다. 실제 경합은 launcher 프로세스, hook 프로세스, handoff controller가 각자 연 핸들 사이에서 일어나므로 경합 재현은 racer마다 별도 `Store` 핸들을 요구한다. Handoff rebind의 CAS, tombstone, BOUND receipt, release marker와 UserPromptSubmit 거부 판정은 각각 한 transaction 안에 있어야 이 경계가 성립한다. 이 경합은 코드 읽기와 저장소 수준 프로브로 세운 가설이며 종단 재현된 실패가 아니므로, AC-FLH-018(launcher 경로)과 AC-FLH-019(UserPromptSubmit 경로)가 별도 핸들 위의 강제 interleaving과 비강제 동시 반복으로 재현 경로를 제공한다.

**Launch-pending 요청의 처리 선택 (REQ-FLH-016).** 대기 후 진행이나 bounded retry가 아니라 즉시 `ENDPOINT_LAUNCH_PENDING` NACK를 택했다. Provisional row를 bound로 바꾸는 유일한 증거는 사용자의 첫 정상 turn이며, 이 SPEC은 빈 model turn 생성(REQ-FLH-006)과 새 polling service(Out of Scope)를 모두 금지하므로 handoff 쪽에서 기다리거나 재시도할 수단이 없다. 즉시 NACK는 REQ-FLH-003의 fail-closed admission과 같은 terminal 의미(`NACK` → fresh reservation only)를 그대로 쓰고, 부작용이 0이라 provisional row의 이후 bind를 방해하지 않는다.

## 3. Common state machine

| State | Entry condition | Allowed next | Body access |
|---|---|---|---|
| `UNBOUND` | handoff 없음 | `RESERVED` | denied |
| `RESERVED` | admission + CAS reservation | `WT_READY`, `NACK`, `ABANDONED` | denied |
| `WT_READY` | exact path/HEAD/branch verified | interactive/headless `SWITCH_PENDING` | denied |
| `SWITCH_PENDING_INTERACTIVE` | user `/cd` instruction issued while idle | next normal-turn SessionStart → `BOUND`, `NACK`, `ABANDONED` | denied |
| `SWITCH_PENDING_HEADLESS` | fork/start request issued while idle | official returned thread ID + controller readback → `BOUND`, `NACK`, `ABANDONED` | denied |
| `BOUND` | atomic rebind transaction committed | dispatch release | current endpoint only |
| `NACK` | fail-closed terminal attempt | fresh reservation only | denied |
| `ABANDONED` | safe auto-recovery 불가 | operator recovery | denied |

Interactive와 headless는 서로의 pending state로 전환하지 않는다. 실패 뒤 다른 mode를 시도하려면 새 nonce/generation reservation이 필요하다.

## 4. Interactive state machine

1. Source lane이 idle이고 permission wait/interrupt가 아님을 확인한다.
2. Target absolute path와 nonce를 operator-visible 안내로 출력한다.
3. 사용자가 Codex TUI에서 `/cd <target>`를 실행한다.
4. Pending 상태에서는 body access를 계속 막고 빈 model turn을 만들지 않는다.
5. 사용자의 다음 정상 turn에서 온 SessionStart evidence로 actual cwd/session UUID를 받는다.
6. Git readback으로 exact worktree root, `WT-*` branch, pinned HEAD를 확인한다.
7. atomic rebind를 수행하고 BOUND receipt를 내보낸다.

설치 CLI에서 `/cd` 후 history가 이어지고 UUID가 바뀐 것은 로컬 실측이다. 공식 API 보장은 아니므로 성공 판정은 새 SessionStart/cwd/branch readback으로만 한다.

## 5. Headless state machine

공식 [Codex app-server 문서](https://developers.openai.com/codex/app-server/)가 정의하는 계약을 따른다.

- `thread/start`: 새 thread를 만들고 `thread/started`를 발행한다.
- `thread/resume`: 기존 thread를 다시 연다.
- `thread/fork`: 저장 history를 복사해 새 `thread.id`를 만들고 `forkedFromId` 및 `thread/started`를 제공한다. In-progress `lastTurnId`는 거부된다.
- `turn/start`: `cwd` override를 받을 수 있고, 그 설정은 같은 thread의 이후 turn 기본값이 된다.
- `turn/steer`: active turn에만 적용되고 `expectedTurnId`가 필요하며 `cwd` override를 받지 않는다.

따라서 자동 headless 경로는 source가 idle일 때 history가 있으면 target `cwd`를 지정한 `thread/fork`, 없으면 `thread/start(cwd=target)`를 수행한다. Controller는 공식 응답의 새 thread ID와 lineage를 기록하고 target cwd/branch/HEAD를 직접 readback한 뒤 SessionStart나 model turn 없이 atomic BOUND를 수행한다. Active turn relocation은 interrupt를 자동 실행하지 않고 NACK한다. Interrupt는 사용자가 별도 의사결정을 한 뒤 새 reservation으로 다시 시작한다.

## 6. Atomic rebind transaction

한 transaction의 precondition과 write set은 다음과 같다.

### Preconditions

- handoff state가 mode별 `SWITCH_PENDING`이다.
- nonce/card/SPEC/lane/source generation이 reservation과 일치한다.
- actual cwd가 canonical target exact root다. Subdirectory, symlink escape, look-alike prefix는 거부한다.
- actual branch가 reserved `WT-<slug>`이고 다른 worktree가 점유하지 않는다.
- `HEAD == develop_pin`이며 target worktree는 dispatch body release 전 clean이다.
- new endpoint UUID가 old UUID와 다르고 PID/process-start identity가 current다.
- interactive이면 evidence가 `/cd` 뒤 사용자의 다음 정상 turn에서 온 SessionStart다.
- headless이면 evidence가 controller가 직접 받은 공식 fork/start 응답과 target provenance readback이며 SessionStart나 빈 turn에 의존하지 않는다.

### Writes

1. Old peer를 tombstone으로 기록한다.
2. Lane current peer를 new UUID/new generation으로 교체한다.
3. Handoff state를 `BOUND`로 전이한다.
4. Durable BOUND receipt/outbox row를 기록한다.
5. Matching pending dispatch의 release generation을 new generation으로 고정한다.

어느 write라도 실패하면 전부 rollback한다. Commit ACK가 유실된 경우 retry는 nonce/generation readback으로 이미 BOUND인지 확인하고 같은 receipt를 재전달한다.

## 7. Worktree provenance and base drift

기존 [ChatGPT Worktrees 문서](https://learn.chatgpt.com/docs/environments/git-worktrees)는 Desktop Handoff가 ChatGPT desktop app이 만든 worktree와 Local 사이에서 chat/code를 옮기는 UI 기능이며, Codex-managed worktree가 기본적으로 detached HEAD일 수 있음을 설명한다. 이 SPEC은 그 기능을 사용하지 않는다.

MoAI card worktree contract는 다음과 같다.

- canonical primary의 local `develop` SHA를 reservation에 먼저 기록한다.
- 기존 MoAI L1 materializer가 `.claude/worktrees/<card-id>`를 만든다.
- 생성 후 target `HEAD`, branch, worktree list를 다시 읽는다.
- target HEAD가 pin과 다르면 `BASE_DRIFT`; fast-forward나 merge로 자동 보정하지 않는다.
- branch는 `WT-<descriptive-slug>`로 만들거나 안전하게 rename하고 card traceability는 directory/dispatch/commit/report로 보강한다.

t1082 자체가 `main@2213871af`에서 생성된 뒤 `develop@3f3ffbb57`로 수동 보정된 사건은 이 gate가 필요한 실제 baseline이다.

## 8. Dispatch and stale endpoint behavior

- Lead는 stable lane으로 보내므로 SWITCH_PENDING 중 message metadata는 유실되지 않는다.
- Body lookup과 claim은 handoff generation이 BOUND일 때만 current endpoint에 허용된다.
- Old endpoint의 direct send/read/receipt/ACK는 tombstone lookup 후 NACK한다.
- NACK의 redirect metadata는 current UUID/generation만 포함하며 body, token, secret은 포함하지 않는다.
- Duplicate dispatch는 t1074 idempotency key에 handoff generation을 결합한다.
- Same-lane redispatch는 current BOUND generation이면 duplicate disposition, 이전 generation이면 stale NACK다.

## 9. Crash recovery decision table

| Observed facts after restart | Decision |
|---|---|
| reservation만 있고 target 없음 | create 재시도 또는 explicit ABANDONED |
| target 존재, exact pin/branch/clean | `WT_READY` 재구성 |
| target 존재, wrong HEAD/branch/collision | `BASE_DRIFT`/`BRANCH_COLLISION` NACK |
| interactive SWITCH_PENDING, old endpoint still current, 다음 정상-turn SessionStart 없음 | pending 유지; 빈 turn 및 자동 dispatch 금지 |
| 비종결 handoff 동안 UserPromptSubmit registration이 `ENDPOINT_HANDOFF_PENDING`으로 거부됨 | endpoint 불변; evidence가 오면 rebind, 검증 실패면 `NACK`, 판단 불가면 `ABANDONED`. 종결 뒤 UserPromptSubmit은 t1074 의미로 복귀하고 dispatch body는 계속 거부; 재시도는 fresh reservation만 |
| headless SWITCH_PENDING, 공식 RPC result 또는 provenance readback 불완전 | NACK 또는 safe retry; SessionStart 대기 및 자동 dispatch 금지 |
| BOUND row와 new peer/tombstone/receipt 모두 존재 | idempotent finalize/readback |
| new peer만 보이고 BOUND/receipt가 없음 | transaction 불가능 상태이므로 corruption NACK; 추측 복구 금지 |
| dirty/unmerged worktree | ABANDONED로 보존, 자동 삭제 금지 |
| old/new owner liveness 불명 | ABANDONED, operator recovery 필요 |

## 10. Security and authority boundaries

- Cwd is untrusted input: clean, absolute, symlink-resolved, canonical-root-contained exact target only.
- Card ID/slug/branch는 allowlist validation을 거친다.
- Shell command string 조립 대신 existing worktree/client typed arguments를 사용한다.
- Hook/MCP는 `/cd` 실행 authority가 없다.
- tmux `send-keys`, private socket, TUI 내부 프로토콜은 금지한다.
- Model에게 “cwd를 바꿔라”라고 보내는 메시지는 evidence가 아니다.
- BOUND는 구현 권한의 최소 전제일 뿐, queue completion 권한이 아니다.

## 11. LIVE evidence boundary

LIVE 판정은 Go test stdout의 자기보고 문자열과 분리한다. 각 LIVE test는 test가 소유한 임시 파일을 완성한 뒤 card-scoped destination으로 atomic rename하여 `.moai/reports/t1082/ac12-evidence.json` 또는 `ac13-evidence.json` 한 개를 남긴다. 판정기는 다음 두 입력을 모두 요구한다.

1. Go JSON event stream: expected LIVE parent와 공통 gate-quality parent가 각각 정확히 한 번 PASS하고, 전체 package/child/subtest에 `fail`, `skip`, `NOT_RUN`이 없다.
2. Structured evidence JSON: schema version과 card/SPEC/mode가 정확하고, process identity, target provenance, mode-specific evidence, receipt, stale rejection, zero-write counters, message conservation, distinct nonces, cleanup, no-bypass가 typed predicate를 만족한다.

Interactive evidence는 실제 `/cd`, `SWITCH_PENDING_INTERACTIVE`, 사용자의 다음 정상 turn SessionStart, reserved/observed cwd·branch·HEAD equality를 요구한다. AC-FLH-013 stored-history LIVE headless evidence는 반드시 실제 `thread/fork(cwd)` request, nonempty returned thread ID, nonempty `forkedFromId`, `thread/started=true`, controller readback equality, SessionStart wait 0을 요구한다. No-history `thread/start(cwd)` 허용은 AC-FLH-004 unit 경로에만 적용되며 AC-FLH-013 evidence predicate에는 포함되지 않는다. 두 mode 모두 real/separate production CLI/model context의 argv, PID, process-start, binary hash를 기록한다.

공통 gate-quality test는 production predicate 자체에 required-field omission, fixture/mock/direct-registration, child failure, child skip, stored-history `wrong_method_thread_start` mutant를 주입한다. 마지막 mutant는 method를 `thread/start`로 바꾸고 `forked_from_id`를 null로 만들어 AC-FLH-013 predicate가 반드시 거부함을 증명한다. 별도로 복제한 느슨한 predicate를 테스트해서는 안 된다. Evidence JSON에는 payload body나 credential을 기록하지 않는다.
