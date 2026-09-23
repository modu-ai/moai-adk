# SPEC 감사 보고서: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

Iteration: 4 (DELTA — t1074 착지 후 전제 변경에 따른 재감사, iter-3 PASS 이후)
Verdict: **FAIL**
Overall Score: 0.81 (Tier L 임계 0.85 미달)

작성자 추론 맥락은 M1 Context Isolation 에 따라 배제했다. 판단 근거는 SPEC 산출물, 착지 코드, 그리고 이번 감사에서 직접 실행한 명령의 출력뿐이다.

> 반복 상한 주의: 하네스 상한은 3회다. 이번 4회차는 iter-3 PASS 이후 전제(t1074 착지)가 바뀌어 열린 델타 감사이므로, 이 FAIL 이후의 진행 방식(수정 후 재감사 / PASS-with-debt / 범위 축소)은 오케스트레이터가 사용자에게 확인해야 한다. iter-3 은 수치 점수를 남기지 않아 점수 회귀(STOP) 판정은 불가하다.

## Claim

개정 커밋 `d1d268f89` 는 M1 과 M3 는 건전하게 닫았으나 M2 는 닫지 못했다. design.md §2.1 과 REQ-FLH-017 은 같은 slot 에 쓰는 경로를 「launcher provisional bind」와 「handoff CAS rebind」 두 개로 고정했다고 서술하지만, 같은 t1074 착지분이 추가한 UserPromptSubmit 훅의 `RegisterPeer` 가 세 번째 비-launcher 쓰기 경로이며, 이번 감사의 프로브에서 동일 소유자·새 session UUID(= interactive `/cd` 뒤 모양)로 slot 을 generation+1 로 교체함을 확인했다. 이 경로는 REQ-FLH-017 의 직렬화 대상에도, AC-FLH-018 의 경합 재현에도 없다.

## Must-Pass Results

- [PASS] MP-1 REQ 번호 일관성: `grep '^### REQ-FLH\|^## ' spec.md` → REQ-FLH-001(L65) … REQ-FLH-015(L121), REQ-FLH-016(L125), REQ-FLH-017(L129). 공백·중복 없음. AC 는 AC-FLH-001..018 연속(acceptance.md L76..L252).
- [PASS] MP-2 GEARS 준수 (요구사항 층 판정): REQ-FLH-016 은 Event-driven「When a handoff is requested for a lane whose current endpoint is still … the handoff SHALL return NACK …」, REQ-FLH-017 은 Ubiquitous 주문장「The handoff SHALL order …」에 While/When 절이 붙은 compound 형. AC-FLH-017/018 의 Given-When-Then 은 검증 층 형식이며 MP-2 대상이 아니다.
- [PASS] MP-3 frontmatter (델타 범위): `version: "0.3.0"`, `updated: 2026-09-23` 만 변경, 나머지 12 필드 + `tier: L` 그대로(명령 출력 아래 Evidence E7).
- [N/A] MP-4 언어 중립성: Go 내부 패키지 전용 SPEC, 템플릿 대상 아님.
- [PASS] MP-5 D7: 개정 diff 안의 SPEC ID 는 자기 자신 하나(`SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001`). BLOCKING 없음.
- [PASS] MP-6 D8: `grep -c syscall spec.md` → 0. auto-PASS.
- [PASS] MP-7 clarification gate: `grep -rn 'NEEDS CLARIFICATION' plan.md research.md` → 출력 없음, exit 1.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | 0.75 | design.md:87「같은 slot에 쓰는 경로가 … 둘로 고정」이 착지 코드와 불일치(D1). REQ-FLH-016 문언은 명확. |
| Completeness | 0.75 | 0.75 | t1074 slot 쓰기 경로 4개 중 UserPromptSubmit `RegisterPeer`(factory_messages.go:93) 미모델링(D1). |
| Testability | 0.75 | 0.75 | AC-FLH-017 은 이진 판정·no-op 에 FAIL. AC-FLH-018 은 「one broker」가 단일 Store 핸들이면 주장된 BEGIN IMMEDIATE 경계를 행사하지 않고도 통과 가능(D2). |
| Traceability | 1.0 | 1.0 | spec.md:152-153 추적표, acceptance.md:43-44 요약표, :69-70 RED ledger 모두 1:1. |

## Defects Found

D1. M2-THIRD-WRITER — design.md:87, spec.md:131 (REQ-FLH-017), acceptance.md:254 (AC-FLH-018) — t1074 착지분의 UserPromptSubmit 훅(`internal/hook/user_prompt_submit.go:112` → `internal/hook/factory_messages.go:93` `s.RegisterPeer(ctx, want)`)이 같은 slot 행의 세 번째 쓰기 경로인데, SPEC 은 쓰기 경로가 둘이라고 단정하고 REQ-FLH-017 의 직렬화 대상을 「launcher provisional registration or bind」로 한정했다. 프로브 P1 에서 살아 있는 동일 PID/process-start 소유자가 새 session UUID 로 `RegisterPeer` 를 호출하자 거절 없이 slot 이 `post-cd-uuid` gen 2 로 교체됐다 — interactive `/cd` 직후의 모양 그대로다. 결과: handoff rebind 가 검증 실패·훅 기한 초과 등으로 커밋되지 못한 턴에서 UserPromptSubmit 이 tombstone·BOUND receipt 없이 endpoint 를 옮길 수 있고, 이는 REQ-FLH-008(검증 성공 후에만 한 transaction 이 endpoint 교체)·REQ-FLH-010(old endpoint 의 STALE redirect)·REQ-FLH-013(BOUND 전 쓰기 0)과 충돌한다. 기준점 `8c5d9be99` 에는 이 훅 쓰기 경로가 0건이고 `08113ff0f` 에서 생겼으므로, 이번 개정이 모델링해야 했던 t1074 착지분 그 자체다. — Severity: critical — Class: blocking — Required fix: (a) design.md §2.1 의 쓰기 경로 목록을 착지 코드 전수(launcher `RegisterLaunchPending`, launcher rollback `RollbackLaunchPending`, SessionStart `BindLaunchPending`, UserPromptSubmit `RegisterPeer`)로 바꾸고 「둘로 고정」 문장을 삭제한다. (b) REQ-FLH-017 에 handoff 가 `SWITCH_PENDING_*` 인 동안 UserPromptSubmit `RegisterPeer` 의 동작을 규정한다 — 예: 미완료 handoff 가 있는 slot 의 session UUID 교체를 같은 transaction 안에서 거부하거나, 교체가 먼저 커밋되면 handoff 는 `STALE_GENERATION` NACK 이되 대체된 endpoint 에 tombstone 을 남긴다 중 하나를 선택하고 근거를 적는다. (c) AC-FLH-018(또는 새 AC)에 post-`/cd` UserPromptSubmit 이 rebind 전/후에 커밋되는 두 순서와, handoff 검증 실패 후 같은 턴 UserPromptSubmit 이 오는 경우를 강제 interleaving 으로 추가한다.

D2. AC18-SINGLE-HANDLE — acceptance.md:254 — AC-FLH-018 은 「one broker」라고만 적었다. `Open` 은 `SetMaxOpenConns(1)`(store.go:219)이라, A·B 가 같은 `*Store` 를 공유하면 직렬화는 Go 커넥션 풀에서 일어나고 design.md:89 가 주장하는 SQLite `BEGIN IMMEDIATE` 교차 커넥션 경계는 한 번도 행사되지 않은 채 테스트가 통과한다. 실제 경합은 launcher 프로세스와 훅/controller 프로세스가 각자 연 핸들 사이에서 일어난다. — Severity: major — Class: blocking — Required fix: AC-FLH-018 Given 에 「A 와 B 는 같은 broker 경로를 각자 `Open` 한 별도 `Store` 핸들을 쓴다」를 명시하고, 단일 핸들 공유를 FAIL 조건으로 적는다.

D3. AC18-OWNER-FIXTURE — acceptance.md:254 — B 선행 분기의 「A's registration returned the t1074 live-owner rejection」은 handoff 가 바인딩한 새 endpoint 의 소유자 identity 가 source 와 다르고 current 일 때만 도달한다(프로브 P2: `factory logical lane has a live owner`). 같은 identity 이거나 current 가 아니면 launcher 가 행을 덮어 launch-pending 으로 만든다(프로브 P3: `registerErr=<nil> … laneErr=factory endpoint is launch-pending`). design.md:143 의 「new endpoint PID/process-start identity가 current」와 결합하면 추론 가능하지만 AC 에는 적혀 있지 않다. — Severity: minor — Class: optional — Required fix: Given 에 「handoff 새 endpoint 소유자는 source 와 다른 identity 이며 current 로 주입」을 명시한다.

D4. M3-SAME-CLASS (판정: 결함 아님) — design.md:43 「`internal/homestate`는 canonical path/run/process identity를 제공」, spec.md:44 동류. `factorymsg.ResolveActiveRun`(store.go:240)은 `homestate.OpenFactory` 의 `runs` 테이블을 읽으므로 run identity 의 저장소는 실제로 homestate 다. 선택 로직만 factorymsg/cli 에 있고, plan.md:41·:84 가 이를 정확히 적었다. — Severity: minor — Class: optional — Required fix: 없음(원하면 design.md:43 을 「run 행 저장, 선택은 `factorymsg.ResolveActiveRun`」으로 다듬는다).

D5. PROGRESS-COUNT — progress.md:42 「15 REQ/16 AC」 — 이번 개정 후 17/18 이다. 날짜가 붙은 과거 기록이면 무해하나 현재 상태 서술로 읽힐 수 있다. — Severity: minor — Class: optional — Required fix: progress.md 에 0.3.0 개정 항목을 추가하거나 해당 줄에 측정 시점을 붙인다.

## 범위별 판정

1. **M1 — REQ-FLH-016 / AC-FLH-017: 닫힘.** 즉시 `ENDPOINT_LAUNCH_PENDING` NACK·부작용 0은 REQ-FLH-003 의 reason-specific NACK·side effect 0(spec.md:75), REQ-FLH-006 의 빈 model turn 금지(spec.md:87), Out of Scope 의 polling service 금지(「새 message broker, daemon, polling service, private IPC를 만들지 않는다」)와 일관된다. AC 는 두 mode 모두 NACK reason·row 수 0·spy 0·provisional 행 바이트 동일을 요구하고, 대조군(bind 뒤 fresh reserve 허용)이 「항상 NACK」 구현을 잡는다. no-op 구현은 NACK reason 단언에서 FAIL, 테스트 부재는 pass 1건 조건에서 FAIL 한다. `ResolveLane` 이 launch-pending 에서 `ErrEndpointLaunchPending` 을 돌려주는 것을 store.go:505-507 에서 확인.
2. **M2 — REQ-FLH-017 / AC-FLH-018: 미해결(D1, D2).** 두 launcher 경로에 대한 순서 주장 자체는 코드와 맞다 — `_txlock=immediate`(store.go:169, 214)로 `BeginTx` 가 `BEGIN IMMEDIATE`, `BindLaunchPending` 은 행이 pending 이 아니면 `return current, false, nil`(store.go:419-421) 로 no-op, CAS 는 `WHERE … session_uuid=? AND generation=? AND pid=? AND process_start=?`(store.go:434-437). 패자 거부 두 갈래는 모두 도달 가능하다(P2 live-owner 거절, A 선행 시 CAS 불일치 → `STALE_GENERATION` 은 설계상 자명). 세 강제 interleaving 과 「관측되지 않으면 FAIL」 조항은 공허 통과를 막는다. 그러나 쓰기 경로 누락(D1)과 단일 핸들 공유 허용(D2) 때문에 「먼저 커밋한 writer 가 이긴다」는 모델이 착지 코드 전체에 대해 성립한다고 볼 수 없다.
3. **M3 — plan.md 좌표: 닫힘.** `internal/cli/factory.go:222 func enterSelectedFactoryRun(…)`, 같은 파일 :226 `factorymsg.ResolveActiveRun(…)` 확인. design.md:43 은 D4 판정대로 결함 아님.
4. **Traceability: 통과.** spec.md:152-153, acceptance.md:43-44(요약표), :69-70(RED ledger), 정책 문구 :15 `AC-FLH-001..018`. REQ heading 형식 `### REQ-FLH-NNN — <title>` 기존과 동일.
5. **Lint 사각지대: 수집 확인.** `internal/spec/lint_req_heading.go:96 reqHeadingWidePattern` 이 `### REQ-…-NNN — …` 를 잡고, `lint_req_widen.go:124` 에서 병합된다. 임시 프로브로 spec.md 를 직접 파싱해 17건 수집, REQ-FLH-016(line 125)·017(line 129) 포함을 확인했다. 다만 heading 출처 항목은 `Widened: true` 로 「reports without gating」(lint_req_heading.go:141-142)이므로, `[]` 는 「finding 없음」이지 게이트 판정이 아니다.

## Evidence

E1. 트리 확인
```
$ git rev-parse --show-toplevel; git branch --show-current; git rev-parse --short HEAD
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1082
WT-factory-lane-worktree-handoff
d1d268f89
```

E2. 슬롯 쓰기 호출부 전수(비-테스트)
```
$ grep -rn 'RegisterPeer\|BindLaunchPending\|RegisterLaunchPending\|RollbackLaunchPending' internal --include='*.go' | grep -v _test.go
internal/cli/factory_launch_pending.go:58:	return s.RegisterLaunchPending(ctx, factorymsg.Peer{
internal/cli/factory_launch_pending.go:73:	_, err = s.RollbackLaunchPending(ctx, pending)
internal/hook/factory_messages.go:84:		p, bound, bindErr := s.BindLaunchPending(ctx, want)
internal/hook/factory_messages.go:93:	p, err := s.RegisterPeer(ctx, want)
(store.go 정의 줄 생략)
$ grep -n 'registerFactoryUserPromptPeer\|registerFactorySessionStartPeer' internal/hook/*.go | grep -v _test
internal/hook/session_start.go:458:	if factoryNotice := registerFactorySessionStartPeer(ctx, input); factoryNotice != "" {
internal/hook/user_prompt_submit.go:112:		bindNotice := registerFactoryUserPromptPeer(bindCtx, input)
```

E3. 기준점 대비 도입 시점
```
$ git grep -c 'registerFactoryUserPromptPeer\|RegisterPeer(ctx, want)' 8c5d9be99 -- internal/hook/
(출력 없음)
$ git grep -c 'registerFactoryUserPromptPeer\|RegisterPeer(ctx, want)' 08113ff0f -- internal/hook/
08113ff0f:internal/hook/factory_messages.go:2
08113ff0f:internal/hook/factory_messages_test.go:2
08113ff0f:internal/hook/user_prompt_submit.go:1
```

E4. 임시 프로브(`internal/factorymsg/zz_auditprobe_t1082_test.go`, 실행 후 삭제; `openTestStore` 사용)
```
$ MOAI_HOME=<scratchpad>/mh go test ./internal/factorymsg -run '^TestZZAuditProbeT1082$' -count=1 -v
P1 err=<nil> got.session=post-cd-uuid got.gen=2 lane.session=post-cd-uuid lane.gen=2 ownerCurrent(src)=true
P2 registerErr=factory logical lane has a live owner lane.session=handoff-uuid lane.gen=1 laneErr=<nil>
P3 registerErr=<nil> pending.gen=2 laneErr=factory endpoint is launch-pending
--- PASS: TestZZAuditProbeT1082 (0.14s)
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	0.481s
```

E5. Lint heading 수집 프로브(`internal/spec/zz_auditprobe_t1082_test.go`, 실행 후 삭제) 및 lint
```
$ go test ./internal/spec -run '^TestZZAuditProbeT1082Heading$' -count=1 -v
heading REQ-FLH-001 line=65 text="While a factory handoff is active, the handoff SHALL keep th"
heading REQ-FLH-016 line=125 text="When a handoff is requested for a lane whose current endpoin"
heading REQ-FLH-017 line=129 text="The handoff SHALL order the t1074 launcher provisional-endpo"
heading-collected=17
ok  	github.com/modu-ai/moai-adk/internal/spec	0.276s
$ moai spec lint SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001 --strict --json
[]
exit=0
```

E6. 프로브 제거 후 트리
```
$ git status --short
(출력 없음)
$ git rev-parse --short HEAD
d1d268f89
```

E7. frontmatter·MP-5/6/7
```
$ grep -rn 'NEEDS CLARIFICATION' plan.md research.md ; echo nc_exit=$?
nc_exit=1
$ git show d1d268f89 -- .moai/specs/ | grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' | sort -u
SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
$ grep -c syscall spec.md
0
version: "0.3.0" / status: draft / created: 2026-09-22 / updated: 2026-09-23 / priority: P1 / lifecycle: spec-anchored / tier: L (그 외 필드 존재)
```

E8. RED ledger 재측정 및 AC-018 jq 판정식 우선순위
```
$ rg -n -F 'func TestFactoryLaneHandoffLaunchPendingSourceNack(' internal --glob '*_test.go'; echo exit=$?
exit=1
$ rg -n -F 'func TestFactoryLaneHandoffRebindVsLaunchBindRace(' internal --glob '*_test.go'; echo exit=$?
exit=1
# DATA RACE 출력 + pass 1건 → false/exit 1, pass 1건만 → true/exit 0
```

E9. M3 좌표
```
$ grep -n '^func enterSelectedFactoryRun\|ResolveActiveRun' internal/cli/factory.go
222:func enterSelectedFactoryRun(root, explicit string, requireActive bool) (func(), error) {
226:	runID, err := factorymsg.ResolveActiveRun(context.Background(), root, explicit)
```

## Baseline-attribution

- 트리: `.claude/worktrees/t1082`, 브랜치 `WT-factory-lane-worktree-handoff`, HEAD `d1d268f89`(감사 전후 동일, E1·E6).
- 코드 좌표는 이 HEAD 의 `internal/factorymsg/store.go`, `internal/hook/factory_messages.go`, `internal/hook/user_prompt_submit.go`, `internal/cli/factory.go` 에서 직접 읽었다.
- 프로브 두 개는 이 트리에 임시로 두었다가 실행 직후 삭제했고, 삭제 후 `git status --short` 가 비어 있음을 확인했다(E6).

## Gaps

- `audit_multi` / `codex_audit` / `glm_audit` 교차 모델 의견은 호출하지 않았다(Claude 단독 델타 감사).
- D1 의 실제 위해 시나리오(handoff rebind 실패 턴의 UserPromptSubmit 교체)는 handoff 구현이 없어 종단 재현하지 않았다. 확인한 것은 t1074 `RegisterPeer` 가 동일 소유자·새 UUID 로 slot 을 교체한다는 저장소 수준 사실(P1)과, 그 경로가 SPEC 의 쓰기 경로 목록에 없다는 문서 사실이다.
- Codex 가 `/cd` 뒤 SessionStart 와 UserPromptSubmit 을 어떤 순서로 발화하는지, `/cd` 슬래시 자체가 UserPromptSubmit 을 일으키는지는 관측하지 않았다.
- 교차 프로세스 `BEGIN IMMEDIATE` 직렬화는 코드 읽기(`_txlock=immediate`)로만 확인했고 두 핸들 경합은 돌리지 않았다.
- 개정이 건드리지 않은 절(REQ-FLH-001..015, AC-FLH-001..016 본문)은 재감사하지 않았다.

## Residual-risk

- D1 을 「미완료 handoff 가 있으면 훅 RegisterPeer 거부」로 고치면 interactive 경로에서 handoff 검증이 계속 실패할 때 lane 이 post-`/cd` session 으로 영영 바인딩되지 못하는 교착이 생길 수 있다. 선택한 해법의 복구 경로(ABANDONED 또는 NACK 후 fresh reservation)를 design.md §9 결정표에 함께 넣어야 한다.
- 프로브 P1 은 `openTestStore` 의 `ownerCurrent` 주입 아래에서 측정했다. 실제 훅은 `homestate.ProbeProcessIdentity` 를 쓰지만 동일 소유자 분기는 `ownerCurrent` 를 호출하지 않으므로(store.go 의 `(oldPID != p.PID || oldStart != p.ProcessStart) &&` 단락) 결론은 주입과 무관하다.

## Recommendation

1. D1 을 먼저 고친다 — design.md §2.1 의 쓰기 경로 목록을 착지 코드 전수로 바꾸고, REQ-FLH-017 에 UserPromptSubmit `RegisterPeer` 와 `SWITCH_PENDING_*` 의 관계를 규정하며, AC 에 해당 interleaving 을 추가한다.
2. D2 — AC-FLH-018 에 별도 `Store` 핸들 사용을 명시한다.
3. D3·D5 는 선택 사항이다.
4. 재감사는 D1·D2 델타로 한정할 수 있으나, 반복 상한(3회)을 이미 넘었으므로 진행 여부는 오케스트레이터가 사용자에게 확인한다.
