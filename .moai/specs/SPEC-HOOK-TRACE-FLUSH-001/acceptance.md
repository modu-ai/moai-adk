# SPEC-HOOK-TRACE-FLUSH-001 — 인수 기준 (v0.2.0)

> **개정 사유**: v0.1.0의 판정 계층이 독립 감사에서 Testability 0.45로 FAIL했다. 근본
> 원인은 하나다 — **여러 AC가 미수정 트리에서 이미 통과했다.** 그 상태로 run-phase에
> 진입하면 아무것도 구현하지 않고 전원 GREEN을 보고할 수 있었다.
>
> 본 개정의 모든 명령은 **미수정 트리에서 실제로 실행**했고, 관측 출력을 각 AC에
> 기준선(baseline)으로 인용했다. 인용된 기준선은 "오늘의 값"이며, 각 AC는 그로부터의
> **델타**를 판정한다.

## 판정 원칙 (v0.2.0에서 강화)

1. **내용 앵커만 사용한다.** 절대 라인 번호와 커밋 SHA는 판정에 쓰지 않는다.
2. **공허 방지 가드도 공허할 수 있다.** `go test -list`는 항상 후행 `ok <pkg>` 줄을
   출력하므로 `| grep -c .`는 무조건 ≥ 1이다. 반드시 `| grep -cE '^Test'`를 쓴다.

   ```
   $ go test ./internal/hook/trace/... -list 'ZZZNoSuchTest' | grep -c .
   1                                   # 공허 — 존재하지 않는 테스트에도 1
   $ go test ./internal/hook/trace/... -list 'ZZZNoSuchTest' | grep -cE '^Test'
   0                                   # 변별력 있음
   ```

3. **`-list`와 `-run`은 동일한 패턴 문자열을 쓴다.** 넓은 패턴(`-list '.*'`)으로 존재를
   확인하고 좁은 패턴으로 판정하면, 확인이 실제 판정 선택자를 구속하지 못한다.
4. **가드는 정확한 이름으로 앵커한다** — `-run '^<이름>$'`. 접두 일치 패턴은 기존
   테스트와 충돌한다(v0.1.0의 실제 실패 원인, AC-HTF-007 참조).
5. **호출이 아니라 효과를 본다.**
6. **사전 존재 부채는 델타로 판정한다.** 무관한 기존 매치를 "0건"으로 요구하는 AC는
   통과 불가이며 범위 밖 편집을 강요한다(AC-HTF-004 참조).

## 가드 이름 (SPEC 소유 · 미수정 트리에 존재하지 않음)

| 가드 | 테스트 이름 | 패키지 | 판정 AC |
|---|---|---|---|
| 1 | `TestRegistryShutdownFlushesLastHandlerEntry` | `internal/hook` | AC-HTF-007 |
| 1E | `TestHookCommandFlushesLastHandlerEntry` | `internal/cli` | AC-HTF-007E |
| 2 | `TestFlushBarrierHasProductionCaller` | `internal/cli` | AC-HTF-009 |

**사전 존재 부재 실측(미수정 트리)**:

```
$ go test ./internal/hook/... -list '^TestRegistryShutdownFlushesLastHandlerEntry$' | grep -cE '^Test'
0
```

## REQ ↔ AC 커버리지

| REQ | AC |
|---|---|
| REQ-HTF-001 (예산 내 전량 기록) | AC-HTF-001, AC-HTF-007 |
| REQ-HTF-002 (CLI teardown 호출) | AC-HTF-006, **AC-HTF-007E** |
| REQ-HTF-003 (경계 초과 시 타임아웃 신호) | AC-HTF-002 |
| REQ-HTF-004 (무한 대기 금지) | AC-HTF-002 |
| REQ-HTF-005 (writer 부재 시 무동작) | AC-HTF-005 |
| REQ-HTF-006 (예산 상수화) | AC-HTF-004 |
| REQ-HTF-007 (인터페이스·Write 계약 불변) | AC-HTF-003, AC-HTF-006b |
| REQ-HTF-008 (행동 가드) | AC-HTF-007, AC-HTF-007E |
| REQ-HTF-009 (반증 가능성) | AC-HTF-008 |
| REQ-HTF-010 (호출자 가드) | AC-HTF-009 |
| REQ-HTF-011 (@MX 강등 + 리포트) | AC-HTF-010 |
| REQ-HTF-012 (3-OS 빌드) | AC-HTF-012 |
| REQ-HTF-013 (미배수 건수 신호) | **AC-HTF-014** |
| (plan.md §B-6 파생 — goleak 억제) | AC-HTF-011 |
| (전체 품질 게이트 — 회귀 방지) | **AC-HTF-013** |

미커버 REQ 0건. 고아 AC 0건.

---

## AC-HTF-001 — 정상 배수 시 적재 항목이 전부 기록된다

**Given** 임시 로그 디렉터리를 가진 `TraceWriter`에 N개(N ≥ 10) 항목이 `Write`로 적재된 상태에서
**When** `CloseWithTimeout`을 넉넉한 예산으로 호출하면
**Then** 트레이스 파일의 줄 수가 정확히 N이어야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestTraceWriterCloseWithTimeoutFlushesAll$'
go test ./internal/hook/trace/... -list "$P" | grep -cE '^Test'          # 기대: 1
go test ./internal/hook/trace/... -run  "$P" -count=1 -v 2>&1 | tail -20
```

**미수정 트리 기준선**: 첫 명령 → `0`. 이 AC는 오늘 미충족이므로 공허 통과하지 않는다.
기대(수정 후): `1` + `--- PASS: TestTraceWriterCloseWithTimeoutFlushesAll`.

---

## AC-HTF-002 — 예산이 소진되면 예산 내에서 타임아웃 신호와 함께 반환한다

**Given** 배경 배수가 예산보다 오래 걸리도록 구성된 `TraceWriter`가 있는 상태에서
**When** 짧은 예산으로 `CloseWithTimeout`을 호출하면
**Then** 호출은 예산의 소수 배 이내에 반환하고, 반환값이 구별 가능한 타임아웃 신호여야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestTraceWriterCloseWithTimeoutAbandonsOnBudget$'
go test ./internal/hook/trace/... -list "$P" | grep -cE '^Test'          # 기대: 1
go test ./internal/hook/trace/... -run  "$P" -count=1 -v 2>&1 | tail -20
```

**미수정 트리 기준선**: 첫 명령 → `0`.
무한 대기(REQ-HTF-004 위반)로 회귀하면 이 테스트는 패키지 타임아웃으로 실패한다 —
따라서 반증 가능하다.

---

## AC-HTF-003 — `Write`의 비차단 계약과 `Close`의 기존 의미가 보존된다

**Given** 기존 `internal/hook/trace/writer_test.go`의 계약 테스트가 있는 상태에서
**When** 전체 `trace` 패키지 테스트를 실행하면
**Then** 기존 테스트 함수가 삭제·의미 변경 없이 통과해야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
git diff -- internal/hook/trace/writer_test.go | grep -cE '^-func Test'   # 기대: 0
grep -c 'func TestTraceWriter_Close_FlushPending' internal/hook/trace/writer_test.go  # 기준선 1 → 기대 1
go test ./internal/hook/trace/... -count=1 2>&1 | tail -5
```

기대: 1번 `0`(기존 테스트 함수 삭제 없음 — 추가는 허용), 2번 `1`(무한 대기 `Close()`
계약 테스트 존치), 패키지 `ok`. 기존 테스트를 고쳐서 통과시킨 흔적이 있으면 FAIL.

---

## AC-HTF-004 — 플러시 예산이 명명 상수이며 배선 지점에 인라인 리터럴이 없다

**Given** 예산값이 필요한 상태에서
**When** **배선된 함수 본문**을 검사하면
**Then** 상수 정의가 존재하고, 배선 지점에 시간 리터럴이 없어야 한다.

> **사전 존재 기준선(실측)**: `internal/cli/hook.go` **전체**에는 시간 리터럴이 1건 있다 —
> `runSpecStatus` 내부의 `5*time.Second`이며 훅 디스패치와 무관하다.
>
> ```
> $ grep -nE '[0-9]+ *\* *time\.(Millisecond|Second)' internal/cli/hook.go
> internal/cli/hook.go:411:	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
> ```
>
> v0.1.0의 "파일 전체 0건" 기대는 이 무관한 부채 때문에 **영영 통과할 수 없었고**,
> 통과시키려면 범위 밖 편집을 강요했다(plan.md §D 범위 규율 위반). 판정을 배선 함수
> 본문으로 좁힌다.

```bash
cd /Users/goos/MoAI/moai-adk-go
grep -cE 'TraceFlush|FlushTimeout' internal/config/defaults.go                                              # 기준선 0 → 기대 >= 1
awk '/^func runHookEvent/,/^}/' internal/cli/hook.go | grep -cE '[0-9]+ *\* *time\.(Millisecond|Second)'    # 기준선 0 → 기대 0
awk '/^func runAgentHook/,/^}/' internal/cli/hook.go | grep -cE '[0-9]+ *\* *time\.(Millisecond|Second)'    # 기준선 0 → 기대 0
grep -cE '[0-9]+ *\* *time\.(Millisecond|Second)' internal/hook/registry.go                                 # 기준선 0 → 기대 0
```

**미수정 트리 기준선**: `0` / `0` / `0` / `0`.
1번이 `0`인 한 이 AC는 통과할 수 없으므로 공허하지 않다. 2·3·4번은 기준선이 이미 0이므로
**회귀 방지 방향**으로 유효하다(인라인 리터럴을 새로 들이지 않는다는 판정).

> `awk` 범위 지정의 타당성 실측: `awk '/^func runHookEvent/,/^}/' … | grep -c 'HookRegistry.Dispatch'` → `1`.
> 즉 이 범위가 해당 함수의 디스패치 지점을 정확히 포함한다(범위가 빈 집합이 아니다).

---

## AC-HTF-005 — 관측성 비활성 시 teardown은 무해한 무동작이다

**Given** `EnableObservability`를 호출하지 않은 레지스트리에서
**When** teardown을 호출하면
**Then** 패닉·오류 없이 즉시 반환해야 하며, 이중 호출도 안전해야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestRegistryShutdownNoopWithoutObservability$|^TestRegistryShutdownIsIdempotent$'
go test ./internal/hook/... -list "$P" | grep -cE '^Test'                # 기대: 2
go test ./internal/hook/... -run  "$P" -count=1 -v 2>&1 | grep -E '^(---|ok|FAIL)' | tail -10
```

**미수정 트리 기준선**: 첫 명령 → `0`.

---

## AC-HTF-006 — 두 CLI 디스패치 지점이 teardown을 defer로 호출한다 (텍스트 존재 판정)

**Given** 프로덕션 `Dispatch` 호출 지점이 `runHookEvent`와 `runAgentHook` 둘뿐인 상태에서
**When** 소스를 검사하면
**Then** 두 함수 모두에 teardown defer 배선이 있어야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
grep -c 'HookRegistry.Dispatch' internal/cli/hook.go      # 기준선 2 (호출 지점 총수)
grep -cE 'defer .*Shutdown\(\)' internal/cli/hook.go      # 기준선 0 → 기대 2
```

**미수정 트리 기준선**: 1번 → `2`, 2번 → `0`.
1번이 3 이상으로 변하면 새 디스패치 지점이 생긴 것이므로 배선 누락 가능성을 재검토한다.

**한계 명시**: 이 AC는 텍스트 존재만 본다. REQ-HTF-002의 **행동** 판정은 AC-HTF-007E다.

## AC-HTF-006b — `hook.Registry` 인터페이스가 넓어지지 않았다

```bash
cd /Users/goos/MoAI/moai-adk-go
sed -n '/type Registry interface/,/^}/p' internal/hook/types.go | grep -cE '^	[A-Z][A-Za-z]*\('   # 기준선 3
sed -n '/type Registry interface/,/^}/p' internal/hook/types.go | grep -ciE 'shutdown|close'        # 기준선 0
```

**미수정 트리 기준선**: `3`(`Register` / `Dispatch` / `Handlers`) / `0`.
수정 후에도 동일해야 한다. 값이 달라지면 REQ-HTF-007 위반.

---

## AC-HTF-007 — 레지스트리 teardown 후 마지막 핸들러 항목이 파일에 존재한다 (핵심 판정)

**Given** `t.TempDir()` 로그 디렉터리로 `registry.EnableObservability`를 **직접 호출**해
관측성을 켜고 핸들러 2개 이상이 등록된 상태에서
**When** 1회 디스패치 후 teardown을 호출하고 트레이스 파일을 읽으면
**Then** **마지막으로 등록된 핸들러의 항목이 존재**해야 하고, 항목 총수가 등록 핸들러 수와
일치해야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestRegistryShutdownFlushesLastHandlerEntry$'
go test ./internal/hook/... -list "$P" | grep -cE '^Test'                # 기대: 1
go test ./internal/hook/... -run  "$P" -count=1 -v 2>&1 | grep -E '^(---|ok|FAIL)' | tail -10
```

**미수정 트리 기준선 (v0.1.0 실패의 핵심)**:

```
$ go test ./internal/hook/... -list '^TestRegistryShutdownFlushesLastHandlerEntry$' | grep -cE '^Test'
0
$ go test ./internal/hook/... -run '^TestRegistryShutdownFlushesLastHandlerEntry$' -count=1 2>&1 | tail -2
ok  github.com/modu-ai/moai-adk/internal/hook/trace   1.899s [no tests to run]
```

> **v0.1.0에서 고친 것 — 선택자 충돌**: 이전 판의 선택자는 `Trace.*Flush|LastHandler`였고,
> 이는 기존 `internal/hook/trace/writer_test.go`의 `TestTraceWriter_Close_FlushPending`
> (구식 **무한 대기 `Close()`**를 검증하는, **다른 패키지의 기존 테스트**)와 충돌했다.
>
> ```
> $ go test ./internal/hook/... -list '.*' | grep -iE 'trace.*flush|flush.*trace|LastHandler'
> TestTraceWriter_Close_FlushPending
> $ go test ./internal/hook/... -run 'Trace.*Flush|LastHandler' -count=1
> ok  github.com/modu-ai/moai-adk/internal/hook   1.329s [no tests to run]      # exit 0
> ```
>
> 정확한 이름 앵커(`-run '^…$'`)로 충돌을 제거했다.

**단언 내용 규율**: "파일이 생성되었다"만 단언하면 FAIL — 실험 A에서 **0바이트 파일이
생성**되었으므로 파일 존재는 아무것도 증명하지 않는다.

---

## AC-HTF-007E — 훅 바이너리 실행 후 마지막 핸들러 항목이 디스크에 남는다 (E2E, REQ-HTF-002)

**Given** `t.TempDir()`에 `.moai/config/sections/observability.yaml`과 `.moai/logs/`만 둔
프로젝트가 있는 상태에서
**When** 빌드한 `moai hook session-start`를 그 프로젝트에서 실행하면
**Then** 기록된 트레이스에 **마지막 핸들러의 항목**이 존재해야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestHookCommandFlushesLastHandlerEntry$'
go test ./internal/cli/... -list "$P" | grep -cE '^Test'                 # 기대: 1
go test ./internal/cli/... -run  "$P" -count=3 -v 2>&1 | grep -E '^(---|ok|FAIL)' | tail -10
```

**미수정 트리 기준선**: 첫 명령 → `0`.

> **`-count=3`인 이유 — 반증이 프로세스 종료 경쟁이기 때문**: 이 가드가 탐지하는 결함은
> 배경 고루틴과 프로세스 종료의 **경쟁**이며, 확률적이다. §1.1 실험 B는 결정적으로
> 관측되었지만(핸들러 2·3번이 기대 20건 중 0건), 빠른 파일시스템에서 단 1회 실행은
> 경쟁에서 이겨 결함을 가릴 수 있다. 반복 실행으로 그 확률을 낮춘다. 테스트 내부에서
> 핸들러 수를 늘리거나 루프를 도는 방식도 동등하게 허용한다 — 요지는 **단발 실행에
> 의존하지 않는 것**이다.

> **이 AC가 신설된 이유 — REQ-HTF-002의 행동 판정 공백**: AC-HTF-007의 가드는
> `internal/hook`에 살고, **`internal/hook`은 `internal/cli`에 의존하지 않는다**.
> 따라서 그 가드는 `internal/cli/hook.go`의 배선을 전혀 실행하지 못한다. AC-HTF-006은
> 텍스트 존재만 본다. 이 AC가 없으면 REQ-HTF-002는 **행동 판정 없이** 남는다.
>
> **의존 방향 실측**:
>
> ```
> $ go list -deps ./internal/hook | grep -cx 'github.com/modu-ai/moai-adk/internal/cli'
> 0
> $ go list -deps ./internal/hook | grep 'moai-adk/internal/cli'
> github.com/modu-ai/moai-adk/internal/cli/preference      # 하위 패키지일 뿐
> ```

테스트는 `go build -o <t.TempDir()>/moai ./cmd/moai`로 바이너리를 만들고 프로세스 경계를
넘어 실행한다. 실험 B(핸들러 3개 중 2·3번이 0건)를 자동화한 형태다.

---

## AC-HTF-008 — 행동 가드가 반증 가능하다 (왕복 판정, 가드별 단일 메커니즘)

**Given** AC-HTF-007과 AC-HTF-007E의 가드가 통과하는 상태에서
**When** **각 가드에 대응하는 단일 반증 지점**을 무력화하고 재실행하면
**Then** 해당 가드가 **실패**해야 하고, 원복 후 다시 통과해야 한다.

| 가드 | 반증 지점 (단일 · 택일 아님) |
|---|---|
| AC-HTF-007 (`internal/hook`) | `registry.Shutdown()` 본문의 플러시 호출 무력화 |
| AC-HTF-007E (`internal/cli`) | `internal/cli/hook.go`의 `defer …Shutdown()` 제거 |

```bash
cd /Users/goos/MoAI/moai-adk-go

# 0) 전제 확인: 두 경로가 깨끗해야 한다. 출력이 비어 있지 않으면 중단한다.
#    (미커밋 변경이 있는 상태로 반증을 시작하면 3)의 원복이 그 변경까지 날린다.)
git status --porcelain -- internal/hook/registry.go internal/cli/hook.go

# 1) 반증: 위 표대로 각 가드의 단일 반증 지점을 무력화한다 (Edit 도구로 편집).

# 2) 두 가드 실행 — FAIL 이 기대값
go test ./internal/hook/... -run '^TestRegistryShutdownFlushesLastHandlerEntry$' -count=1 2>&1 | tail -5
go test ./internal/cli/...  -run '^TestHookCommandFlushesLastHandlerEntry$'      -count=1 2>&1 | tail -5

# 3) 원복: Edit 도구로 원문을 되돌리거나, 아래 경로 한정 restore 를 쓴다.
#    stash 를 만들지 않고, 지정한 두 경로 외에는 아무것도 건드리지 않는다.
git restore --source=HEAD --worktree -- internal/hook/registry.go internal/cli/hook.go

# 4) 재실행 — PASS 가 기대값
go test ./internal/hook/... -run '^TestRegistryShutdownFlushesLastHandlerEntry$' -count=1 2>&1 | tail -3
go test ./internal/cli/...  -run '^TestHookCommandFlushesLastHandlerEntry$'      -count=1 2>&1 | tail -3
```

기대: 0)이 빈 출력, 2)에서 두 명령 모두 `FAIL`, 4)에서 두 명령 모두 `ok`.
**왕복 4개 테스트 출력이 모두** 보고에 인용되어야 성립한다.

> **v0.1.0에서 고친 것 ① — 반증 경로의 교차 오배정**: 이전 판은 AC-HTF-007의 반증을
> "`internal/cli/hook.go`의 defer 제거 **또는** `Shutdown` 본문 무력화"의 **택일**로
> 적었다. 의존 방향상 전자로는 `internal/hook` 가드를 반증할 수 없으므로, 전자를 고른
> run-phase 에이전트는 가드가 여전히 통과하는 것을 보고 **정상 가드를 결함으로 오판**하게
> 된다. 가드별 반증 지점을 1:1로 고정했다.
>
> **v0.2.0에서 고친 것 ② — 원복 프리미티브**: 이 저장소는 병렬 세션이 공유하는
> 체크아웃이므로, 무자격 `git checkout -- <경로>`는 같은 파일에 대한 **다른 세션의 편집을
> 조용히 파괴**한다. 그 금지는 유지한다.
>
> 다만 v0.2.0이 대체재로 지정했던 `git stash push -- <경로>` / `git stash pop`은 **더
> 나쁜 프리미티브였고**, 본 개정에서 철회한다. 구현이 커밋된 상태(= 반증 시점의 정상
> 상태)에서 두 경로는 깨끗하므로 `stash push`는 `No local changes to save`를 출력하고
> **스택을 만들지 않는다**. 뒤이은 `stash pop`은 **스택 맨 위에 이미 있던 남의 작업을
> 꺼내 놓는다** — 금지 사유로 든 바로 그 사고를, 금지의 대체재가 일으킨다.
>
> 격리 저장소 실측(다른 세션 작업을 stash 에 올려둔 상태 재현):
>
> ```
> $ git stash push -- a.go b.go
> No local changes to save
> $ git stash list | wc -l
> 1                      # 푸시 전과 동일 — 스택이 생기지 않았다
> $ git stash pop        # (a.go 를 사보타주한 뒤)
> $ cat other.go
> other
> OTHER SESSION WORK     # 남의 작업이 워킹트리로 튀어나왔다
> $ cat a.go
> A SABOTAGED            # 정작 내 사보타주는 그대로 남았다
> ```
>
> 대칭 실패도 있다. 구현이 **미커밋** 상태라면 0단계의 `stash push`가 구현 자체를
> 치워버리므로, 가드는 "teardown 이 제거되어서"가 아니라 "구현이 없어서" 실패한다 —
> 공허한 반증이다.
>
> 대체재는 **경로 한정 restore**다. 스택을 만들지 않고, 지정한 경로만 HEAD 로 되돌리며,
> 다른 파일의 미커밋 변경을 보존한다(같은 격리 저장소 실측: 사보타주한 `a.go`/`b.go`만
> 원복되고 `other.go`의 동시 미커밋 편집은 그대로, stash 깊이 불변).

---

## AC-HTF-009 — 플러시 장벽에 프로덕션 호출자가 1개 이상 존재한다 (보조)

**Given** 수정이 적용된 상태에서
**When** **훅·CLI 패키지의 비주석 프로덕션 코드**를 검색하면
**Then** 플러시 장벽 호출이 1건 이상이어야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
grep -rnE 'CloseWithTimeout\(|\.Shutdown\(\)' --include='*.go' internal/hook/ internal/cli/ \
  | grep -v _test.go | grep -vE ':[[:space:]]*//' | grep -c .
```

**미수정 트리 기준선**: `0` → 기대 `>= 1`. 오늘 FAIL이므로 델타를 판정한다.

> **v0.1.0에서 고친 것 — 변별력을 잃은 선택자**: 이전 판은 범위를 `internal/ cmd/ pkg/`
> 전체로 넓혔고("값 복사 소비자를 놓치지 않도록"), 그 결과 무관한 **한국어 주석 한 줄**에
> 걸려 미수정 트리에서 통과했다.
>
> ```
> $ grep -rnE 'CloseWithTimeout\(|\.Shutdown\(\)' --include='*.go' internal/ cmd/ pkg/ | grep -v _test.go
> internal/web/handlers.go:604:// httpSrv.Shutdown() 이 이 핸들러의 반환을 대기하며 교착(dead lock) 한다. 고루틴이
> ```
>
> 범위를 관련 패키지로 좁히고 주석 행을 제외해 기준선을 0으로 만들었다. 넓히는 것이
> 아니라 좁히는 것이 변별력을 만든다.

**한계 명시**: 이 AC는 "소스에 호출이 있다"만 증명하고 효과를 증명하지 않는다.
효과 판정은 AC-HTF-007 / 007E / 008이 담당하며, 이 AC가 그것을 대체하지 않는다.

---

## AC-HTF-010 — `Close` 앞 @MX 주석이 실측된 사실만 말하고, 강등이 리포트된다

**Given** 실측 프로덕션 fan_in이 3 미만인 상태에서
**When** `internal/hook/trace/writer.go`를 검사하면
**Then** 거짓 수치·서술이 제거되고 `@MX:NOTE`로 강등되어 있어야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
grep -c 'fan_in=24' internal/hook/trace/writer.go                                    # 기준선 1 → 기대 0
grep -c 'all hook handler teardown paths call Close' internal/hook/trace/writer.go   # 기준선 1 → 기대 0
grep -c '@MX:NOTE' internal/hook/trace/writer.go                                     # 기준선 0 → 기대 >= 1
grep -c '@MX:WARN' internal/hook/trace/writer.go                                     # 기준선 1 → 기대 >= 1 (보존)
```

**미수정 트리 기준선**: `1` / `1` / `0` / `1`.
기대(수정 후): `0` / `0` / `>=1` / `>=1`.

4번(`NewTraceWriter` 위의 고루틴 수명 WARN)은 **보존**해야 한다 — 위험이 사라지지
않았으므로 삭제하면 FAIL. 강등 사실은 run-phase @MX 리포트에 기재되어야 한다
(REQ-HTF-011, `mx-tag-protocol.md` "demote to NOTE via report").

---

## AC-HTF-011 — goleak 억제 항목이 해소되었거나 사유가 갱신되었다

**Given** `internal/hook/main_test.go`가 `trace.(*TraceWriter).run` 억제를 갖고 있던 상태에서
**When** 억제 제거를 시도하고 패키지 테스트를 실행하면
**Then** 억제가 제거되었거나, 유지된 경우 사유가 실측 기반으로 갱신되어 있어야 한다.

> **`IgnoreTopFunction`은 2회 등장하고 성격이 다르다**(실측): L13은 **문서 주석**
> (`// IgnoreTopFunction entries below cover …`), L35는 **실제 호출**
> (`goleak.IgnoreTopFunction("…trace.(*TraceWriter).run")`). 따라서 "총 등장 0건"을
> 요구하면 호출만 지워도 주석 때문에 1이 남아 **엉뚱하게 실패**하고, 의도치 않은 주석
> 편집을 강요한다. 판정은 **호출 라인**만 센다.

```bash
cd /Users/goos/MoAI/moai-adk-go
grep -c 'goleak.IgnoreTopFunction' internal/hook/main_test.go                                  # 기준선 1 (호출만)
grep -cE 'goleak\.IgnoreTopFunction\(".*trace\.\(\*TraceWriter\)\.run"\)' internal/hook/main_test.go   # 기준선 1 (호출 라인 한정)
grep -c 'SPEC-HOOK-TRACE-FLUSH-001' internal/hook/main_test.go                                 # 기준선 0
go test ./internal/hook/ -count=1 2>&1 | tail -3
```

**미수정 트리 기준선**: `1` / `1` / `0`.

두 참고값 모두 **주석 충돌 때문에 쓰지 않는다**: `grep -c 'IgnoreTopFunction'`은 L13
문서 주석을 포함해 `2`이고, `grep -c 'trace.(\*TraceWriter).run'` 역시 L17 문서 주석을
포함해 `2`다. 후자는 위 rationale이 경고한 것과 **같은 부류의 충돌**이므로, 2번 선택자는
`goleak.IgnoreTopFunction(...)` **호출 라인**에 앵커한다.

기대: 다음 **둘 중 하나**이며, 어느 쪽이든 패키지가 `ok`여야 한다.

- **(a) 제거**: 2번이 `0`. `trace.(*TraceWriter).run` 억제 **호출**이 사라졌다는 뜻이며,
  이것이 plan.md §B-6이 지시하는 **제거 시도의 성공** 결과다. (L17 문서 주석은 판정 대상이
  아니므로 남아 있어도 무방하다 — 범위 밖 편집을 강요하지 않는다.)
- **(b) 유지 + 재정당화**: 2번이 `1`이고 3번이 `>= 1`. 억제 호출을 남기되 본 SPEC 기준의
  실측 사유로 갱신했다는 뜻이다.

**(b)를 선택한 경우 추가 의무**: 갱신된 주석은 **제거를 시도했고 실패했다는 사실과 그
실패 이유**(어떤 테스트가 writer 를 닫지 않는지)를 명시해야 한다. 단순히 SPEC ID만
덧붙이면 3번은 통과하지만 §B-6이 요구한 **평가**가 아니라 문서 편집에 그친다 — 이 경우
run-phase 보고에 제거 시도의 실행 출력을 함께 인용해야 (b)가 성립한다.

2번이 `1`이고 3번이 `0`인 채로 남으면 FAIL — 기존 사유("SPEC-V3R6-HOOK-ASYNC-EXPAND-001
범위 밖")는 본 SPEC 이후 더 이상 정확하지 않다.

---

## AC-HTF-012 — 3-OS 빌드가 통과한다 (회귀 방지)

```bash
cd /Users/goos/MoAI/moai-adk-go
GOOS=linux   GOARCH=amd64 go build ./... && echo LINUX_OK
GOOS=darwin  GOARCH=arm64 go build ./... && echo DARWIN_OK
GOOS=windows GOARCH=amd64 go build ./... && echo WINDOWS_OK
```

기대: 세 줄 모두 출력.
**성격 명시**: 기준선도 통과하므로 이 AC는 **회귀 방지 전용**이며, 단독으로는 구현 여부를
판정하지 않는다(공허 통과가 아니라 의도된 회귀 가드).

---

## AC-HTF-013 — 전체 테스트와 린트가 통과한다 (회귀 방지)

```bash
cd /Users/goos/MoAI/moai-adk-go
go test ./internal/hook/... ./internal/cli/... ./internal/config/... -count=1 2>&1 | grep -c '^FAIL'   # 기대: 0
golangci-lint run 2>&1 | tail -20
go vet ./internal/hook/... ./internal/cli/... 2>&1 | tail -10
```

기대: FAIL 0건, 린트 무경고, vet 무출력.
**성격 명시**: AC-HTF-012와 같은 회귀 방지 AC다.

---

## AC-HTF-014 — 예산 소진 시 미배수 건수가 관측 가능하게 방출된다 (REQ-HTF-013)

**Given** 배수가 예산 내에 끝나지 않도록 구성된 `TraceWriter`가 있는 상태에서
**When** 짧은 예산으로 teardown을 호출하면
**Then** 방출된 로그 레코드에 **미배수 항목 수**가 담겨 있어야 하고, 그 값이 실제 잔여
건수와 일치해야 한다.

```bash
cd /Users/goos/MoAI/moai-adk-go
P='^TestCloseWithTimeoutReportsUndrainedCount$'
go test ./internal/hook/trace/... -list "$P" | grep -cE '^Test'          # 기대: 1
go test ./internal/hook/trace/... -run  "$P" -count=1 -v 2>&1 | grep -E '^(---|ok|FAIL)' | tail -5
```

**미수정 트리 기준선**: 첫 명령 → `0`.

이 AC는 REQ-HTF-001(예산 내 전량 기록)과 REQ-HTF-003(예산 소진 시 포기)의 경계에 남는
**잔여 유실을 침묵시키지 않기 위한** 판정이다. 유실을 0으로 만들 수는 없지만, 측정
가능해야 한다.

**부수 산출물(plan.md §E7 — 판정 아님, 보고 의무)**: 이 테스트를 작성하면서 정상 경로의
실제 배수 소요를 함께 측정하고 그 값을 run-phase 보고에 기록한다. 200ms는 plan.md §B-3이
명시한 **추론 기반 잠정 앵커**이므로, 측정 결과가 앵커를 반박하면 상수를 조정하고 조정
사유를 보고한다.

---

## 판정 규율

- 각 AC는 **실행된 명령과 그 출력**을 함께 인용해야 한다. 요약("전부 통과")은 증거가 아니다.
- AC-HTF-008(반증 왕복)은 왕복 4개 출력이 모두 인용되어야 성립한다.
- 어떤 AC도 절대 라인 번호나 커밋 SHA에 앵커해서는 안 된다.
- 테스트 기반 AC는 **`-list`와 `-run`에 동일한 패턴**을 쓰고,
  `-list … | grep -cE '^Test'`가 기대값과 일치함을 먼저 보여야 한다.
- 되돌리기에는 **Edit 도구로 원문 복원** 또는
  `git restore --source=HEAD --worktree -- <경로>`를 쓴다. 둘 다 경로 한정이며 stash 를
  만들지 않는다. 무자격 `git checkout -- <경로>`는 병렬 세션의 작업을 파괴하므로 금지하고,
  `git stash push` / `git stash pop` 쌍도 금지한다 — 깨끗한 경로에서는 스택을 만들지 않은
  채 `pop`이 **남의 stash 를 꺼내므로** 같은 사고를 일으킨다(AC-HTF-008 주석의 실측 참조).
- AC-HTF-012 / AC-HTF-013은 기준선도 통과하는 **회귀 방지 AC**다. 이 둘의 PASS를
  구현 완료의 근거로 인용해서는 안 된다.

## Definition of Done

- AC-HTF-001 … AC-HTF-014 전부 PASS, 각각 실행 출력 인용.
- 미커버 REQ 0건, 고아 AC 0건.
- `internal/template/templates/` 무수정 (`git status --porcelain internal/template/` 출력 없음).
- @MX 리포트에 ANCHOR → NOTE 강등 1건 기재.
- AC-HTF-011을 **(b) 유지 + 재정당화**로 종결한 경우, run-phase 보고가 goleak 억제
  **제거 시도의 명령과 그 출력**을 인용함(세 grep으로는 기계적으로 확인되지 않는
  항목이므로 보고 검토에서 확인한다).
- 실제 배수 소요 측정값이 보고에 포함되고, 200ms 앵커의 유지/조정 판단이 함께 기재됨.
- spec.md §4 범위 제외 2건(observability.enabled 미독해, 훅 4계열 재조사)이 후속 항목으로
  기록되었고 본 SPEC에서 구현되지 않았음.
