---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
document: research
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1082
module: "internal/factorymsg"
---

# Research — SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## 1. Measured repository facts

- `internal/factorymsg/store.go`는 stable `slot`과 replaceable `SessionUUID`/`Generation`을 가진 `Peer`, `RegisterPeer`, `ResolveLane`, generation-bound message/receipt, roster를 이미 제공한다.
- `RegisterPeer`는 살아 있는 logical lane owner를 임의로 덮어쓰지 않는다. Handoff 구현은 이 invariant를 우회하지 않고 reservation nonce와 atomic rebind라는 명시 경로를 추가해야 한다.
- `internal/homestate/paths.go`는 linked worktree를 primary checkout으로 canonicalize하는 `CanonicalProjectRoot`와 shared home-state path를 제공한다.
- `internal/cli/worktree/new.go`는 `moai worktree new <name>`을 existing shared session-worktree materializer에 위임한다. 새 raw Git creation path는 필요 없다.
- `internal/cli/worktree_branch_flag.go`는 existing branch checkout과 branch collision 검증을 제공하지만, 카드 계약은 local develop pin에서 새 `WT-<slug>` branch를 만드는 경로이므로 existing-branch 기능과 혼동하면 안 된다.
- `internal/cli/mcp_codex.go`는 app-server process, initialize, `thread/start`/`thread/resume`, async NDJSON, turn notification, bounded close를 이미 구현한다. Fork/cwd handoff는 이 client seam의 최소 확장 대상이다.
- `internal/hook/session_start_record.go`는 SessionStart가 실제 session ID를 처음 소유하는 actor임을 명시하고, cwd에서 card ID를 계산하는 기존 패턴을 가진다.
- `internal/hook/cwd_changed_relocate.go`는 일반 session registry의 cwd를 fail-open으로 갱신하지만 factory lane peer generation과 BOUND receipt는 다루지 않는다. 이를 BOUND 증거로 재해석하면 안 된다.
- t1074는 factory MCP catalog를 36개로 확장했고 현재 분류는 write 14/read 22다(plan 시점 관측이다. 현재 총수는 `.claude/rules/moai/core/moai-mcp-tools.md`가, 쓰기/읽기 구성은 `internal/mcp/catalog.go`의 `WriteCapable` 표시가 정하며, 카드 t1143이 39개(쓰기 15, 읽기 24)로 올렸다).

## 2. Measured branch and dependency baseline

이번 plan 작성 중 다음을 직접 관측했다.

```text
git -C .claude/worktrees/t1082 status --short --branch
## WT-factory-lane-worktree-handoff

git -C .claude/worktrees/t1082 rev-parse --short HEAD
bf39a539d

git -C .claude/worktrees/t1082 reflog --date=iso --format='%h %gd %gs' -12
bf39a539d ... commit (merge): merge(t1082): absorb t1074 factory dependency
3f3ffbb57 ... merge develop: Fast-forward
2213871af ... Branch: renamed refs/heads/t1082 to refs/heads/WT-factory-lane-worktree-handoff
```

`git merge-base --is-ancestor 3f3ffbb57 bf39a539d`와 `git merge-base --is-ancestor 8c5d9be99 bf39a539d`는 모두 exit 0이었다. Merge commit `bf39a539d`의 parents는 `3f3ffbb57`와 t1074 dependency `8c5d9be99`다.

이 reflog는 실제 creation-base drift를 보여준다. 새 worktree가 primary `main@2213871af`에서 시작한 뒤 local `develop@3f3ffbb57`로 fast-forward 보정됐다. 향후 handoff는 사후 보정이 아니라 create 전 pin과 create 후 exact equality로 실패를 막아야 한다.

## 3. Measured dependency checks

환경 변수를 한 command invocation 안에서 scrub하고 다음 narrow baseline을 실행했다.

```text
MOAI_HOME=/tmp/t1082-plan-baseline-home GOCACHE=/tmp/t1082-plan-baseline-cache \
go test ./internal/mcp -run '^(TestMoaiMCPTools_CatalogSize|TestMoaiMCPTools_FourteenWriteCapable|TestMoaiMCPTools_NoDuplicateNames|TestMoaiMCPToolNames_MatchesCatalog)$' -count=1 -timeout=90s
ok github.com/modu-ai/moai-adk/internal/mcp 0.238s

MOAI_HOME=/tmp/t1082-plan-baseline-home GOCACHE=/tmp/t1082-plan-baseline-cache \
go test ./internal/cli -run '^(TestMoaiMCPServer_RegistrationMatchesCatalog|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s
ok github.com/modu-ai/moai-adk/internal/cli 2.716s
```

이 결과는 t1074 catalog/status seam의 현재 narrow baseline일 뿐이며 t1082 handoff 동작이 존재하거나 통과한다는 뜻이 아니다.

## 4. Official Codex app-server contract

공식 근거: [Codex app-server](https://developers.openai.com/codex/app-server/)

- `thread/start`는 새 conversation을 만들고 `thread/started`를 발행한다.
- `thread/resume`은 저장된 thread를 같은 ID로 다시 연다.
- `thread/fork`는 저장 history를 새 thread ID로 복사하고, 응답에 새 `thread.id` 및 가능한 경우 `forkedFromId`를 담으며 새 thread의 `thread/started`를 발행한다.
- In-progress `lastTurnId`를 지정한 fork는 거부된다. Source가 mid-turn인데 `lastTurnId`를 생략하면 partial turn을 그대로 보존하는 대신 interruption marker가 기록된다. 이 SPEC은 active-turn handoff를 허용하지 않으므로 그 fallback에 의존하지 않는다.
- `turn/start`의 `cwd`는 turn-level override이며 지정하면 같은 thread의 이후 turn 기본값이 된다.
- `turn/steer`는 현재 active turn에 input을 추가하는 API다. `expectedTurnId`가 active ID와 일치해야 하고 active turn이 없으면 실패하며, `cwd`를 포함한 turn-level override를 받지 않는다.

결론: active-turn relocation은 공식 경로가 아니다. Idle `thread/fork`/`thread/start`가 반환한 공식 thread ID와 controller의 target provenance readback이 headless 결합 증거이며, `turn/steer`는 handoff에 사용할 수 없다. Handoff 결합만을 위해 빈 `turn/start`를 만들지 않는다.

같은 날 설치된 Codex 0.155.1의 공식 protocol generator를 직접 실행했다.

```text
codex app-server generate-json-schema --out <tmp>
exit 0
ClientRequest.json definitions.ThreadForkParams.properties.cwd: ["string", "null"]
ClientRequest.json definitions.ThreadForkParams.required: ["threadId"]
```

따라서 이 설치 버전의 `thread/fork`는 optional `cwd` override를 지원한다. LIVE gate는 문서 예시만 추론하지 않고 실제 fork request에 reserved target `cwd`를 넣고 반환 thread ID와 controller readback을 함께 증명해야 한다.

## 5. Official Worktrees boundary

공식 근거: [ChatGPT Worktrees](https://learn.chatgpt.com/docs/environments/git-worktrees)

- 문서의 `Handoff`는 ChatGPT desktop app에서 Local과 app-managed Worktree 사이로 chat과 code를 옮기는 UI flow다.
- 해당 Worktree는 ChatGPT desktop app이 local checkout에서 만들며 기본적으로 detached HEAD일 수 있다.
- Git은 branch 하나를 동시에 두 worktree에 checkout하지 못하므로 branch collision을 명시적으로 다뤄야 한다.

결론: Desktop Handoff는 headless MoAI CLI의 암묵적 capability가 아니다. MoAI card worktree는 launcher가 만든 `.claude/worktrees/<card-id>`와 named `WT-*` branch를 사용한다. 두 종류의 worktree를 같은 것으로 취급하지 않는다.

## 6. Local `/cd` observation boundary

t1074 plan evidence에는 installed Codex CLI `0.155.1`에서 `/cd` 뒤 visible conversation이 이어지고 statusline cwd/branch와 session UUID가 바뀐 로컬 관측이 있다. 이는 구현 방향을 뒷받침하지만 공식 API 보장은 아니다.

따라서 interactive PASS는 “`/cd`가 언제나 UUID를 fork한다”가 아니라 다음 readback을 요구한다.

- 사용자 동작 뒤 다음 정상 turn에서 온 새 SessionStart evidence
- actual absolute cwd == reserved worktree root
- actual branch == reserved `WT-*`
- actual HEAD == pinned develop base
- old endpoint와 구분되는 current endpoint/generation
- binding을 유도한 empty model turn 0

## 7. Pre-turn SessionStart observation boundary

2026-09-22에 trust가 완료된 임시 fixture에서 실제 Codex 0.155.1 app-server를 initialize한 뒤 `thread/start`만 호출한 bounded probe는 공식 thread ID를 반환했다. `turn/start`는 호출하지 않았고, 관측한 5초 안에는 SessionStart sidecar, factory DB, current-session-id가 나타나지 않았다.

이 한 번의 버전·fixture·관측 창은 “Codex는 첫 turn 전에 SessionStart를 절대 실행하지 않는다”는 일반 법칙을 증명하지 않는다. 다만 pre-turn SessionStart를 t1082의 필수 headless 전제로 둘 근거도 제공하지 않는다. 따라서 설계는 다음처럼 capability truth에 맞춘다.

- headless: 공식 fork/start 응답의 thread ID와 controller provenance readback으로 직접 BOUND한다.
- interactive: `/cd` 뒤 `SWITCH_PENDING_INTERACTIVE`를 유지하고 사용자의 다음 정상 turn SessionStart를 기다린다.
- 어느 mode도 결합 증거를 만들기 위한 empty model turn을 발행하지 않는다.

## 8. Root-cause analysis

| Layer | Finding |
|---|---|
| Surface | 배차된 lane이 primary/이전 cwd에 남아 잘못된 tree에서 구현할 수 있다. |
| Why 1 | Card dispatch와 cwd relocation/rebind가 하나의 상태 계약으로 묶이지 않았다. |
| Why 2 | Stable lane 주소와 physical thread UUID를 같은 identity처럼 다루면 UUID 회전이 message loss 또는 stale ACK가 된다. |
| Why 3 | Worktree creation base, branch traceability, SessionStart evidence가 서로 다른 subsystems에 흩어져 있다. |
| Root cause | “dispatch 가능”과 “검증된 card work-root에 current endpoint가 bound됨”을 구분하는 durable gate가 없다. |

## 9. Research gaps and run-phase gates

- Real interactive Codex `/cd`의 현재 installed-version behavior는 LIVE에서 다시 측정해야 한다.
- Real app-server `thread/fork(cwd)` 반환 ID의 controller direct BOUND는 아직 NOT_RUN이다.
- Real interactive `/cd` 뒤 다음 정상 turn SessionStart 결합은 아직 NOT_RUN이다.
- Claude lead↔Codex lane과 Codex↔Codex nonce/receipt proof는 아직 NOT_RUN이다.
- Branch collision, dirty source, crash points, restart, abandoned worktree는 아직 test가 없다.
- Codex statusline은 corroborating evidence다. BOUND authority는 broker transaction/readback이다.
- Same-UID local process identity는 hostile-user security boundary가 아니다. 이 설계는 stale/mistaken ownership 방지용이다.
- LIVE 판정은 자유 형식 stdout marker로 닫지 않는다. AC-FLH-012/013의 card-scoped structured evidence schema와 exact `jq` predicate가 process identity, mode별 공식 evidence, provenance equality, receipts, counters, cleanup, no-bypass를 검증하며, 공통 mutant selector가 missing/fixture/mock/direct-registration/child-fail/child-skip 및 stored-history `wrong_method_thread_start` 거부를 증명해야 한다. AC-FLH-013은 actual `thread/fork(cwd)`, nonempty returned thread ID, nonempty `forkedFromId`, `thread/started=true`를 정확히 요구하고, no-history `thread/start(cwd)`는 AC-FLH-004 unit 경로에만 허용한다.
