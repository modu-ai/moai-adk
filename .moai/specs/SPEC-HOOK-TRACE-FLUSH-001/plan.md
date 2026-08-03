# SPEC-HOOK-TRACE-FLUSH-001 — 구현 계획

> 절 순서는 **되돌리기 어려움(decision reversibility)** 기준이다. 바뀔 가능성이 큰 결정
> (타입 인터페이스 · 예산값 · 가드 설계)을 앞에 두고, 기계적 편집을 뒤로 미룬다.

## §A Tier 판정

**Tier M.**

근거:

- 편집 대상 프로덕션 파일 4개(`internal/hook/trace/writer.go`, `internal/hook/registry.go`,
  `internal/cli/hook.go`, `internal/config/defaults.go`) + 테스트 2~3개.
- 새 타입 시그니처 2개(`TraceWriter.CloseWithTimeout`, `registry.Shutdown`)를 도입하므로
  단일 관심사 사소 변경(Tier S)이 아니다.
- 사용자 표면(CLI 플래그 · 출력 · 설정 스키마) 변경이 없고 마일스톤이 5개 이하이므로
  Tier L도 아니다. `design.md` / `research.md`는 만들지 않는다.

산출물: `spec.md` + `plan.md` + `acceptance.md` + `progress.md`.

## §B 설계 결정 (되돌리기 어려운 순)

### B-1 플러시 장벽의 시그니처 — 경계 대기 도입 (승인됨)

현재:

```go
func (w *TraceWriter) Close() error {
    w.once.Do(func() { w.closed.Store(true); close(w.ch) })
    <-w.done            // 무한 대기
    return nil
}
```

도입:

```go
// CloseWithTimeout flushes pending entries, waiting at most d for the
// background goroutine to drain. Returns a distinguishable timeout signal
// when the budget elapses first.
func (w *TraceWriter) CloseWithTimeout(d time.Duration) error
```

- `Close()`는 보존한다(기존 테스트 호출자 존재, 후방 호환). `Close()`는
  `CloseWithTimeout`의 무한 대기 변형으로 남기거나 내부 공통 경로를 공유한다 —
  택일은 run-phase 재량이나, **`Close()`의 기존 의미(완전 배수까지 대기)는 바꾸지 않는다.**
- 타임아웃 신호는 구별 가능해야 한다(예: 패키지 수준 센티널 오류). 호출자는 이를
  치명 오류로 승격하지 않고 `slog.Warn`으로 낮춰 기록한다 — 훅은 관측 실패로 사용자
  세션을 깨뜨리지 않는다.

**기각한 대안**

- **(a) `Write`를 완전 동기화** — 가장 단순하고 무손실이지만, 디스크 지연이 훅 핫 경로에
  그대로 노출된다. 훅 예산은 5초이며(`internal/hook/CLAUDE.md`), 느린 파일시스템에서
  사용자 세션이 직접 느려진다.
- **(b) 기존 무한 대기 `Close()`를 그대로 defer로 호출** — 디프가 가장 작지만, 멈춘 디스크
  하나가 5초 예산을 초과시키고 세션을 정지시킨다. 비동기 설계의 이점을 통째로 버린다.

### B-2 teardown 도달 경로 — 인터페이스를 넓히지 않는다

`registry`에 `Shutdown()`을 추가하되 **`hook.Registry` 인터페이스에는 넣지 않는다**
(REQ-HTF-007). 호출 지점은 `internal/cli/deps.go`의
`enableObservabilityIfConfigured`가 이미 쓰는 **익명 인터페이스 타입 단언 선례**를 따른다:

```go
type registryShutdowner interface{ Shutdown() }
if rs, ok := deps.HookRegistry.(registryShutdowner); ok {
    defer rs.Shutdown()
}
```

- `Shutdown()`은 `traceWriter == nil`일 때 무동작이어야 한다(REQ-HTF-005).
- `Shutdown()`은 뮤텍스 아래에서 writer 참조를 꺼낸 뒤 잠금을 놓고 대기해야 한다 —
  `writeTrace`가 같은 뮤텍스를 쓰므로 잠금을 쥔 채 배수를 기다리면 교착 위험이 있다.
- `Shutdown()`은 멱등해야 한다(`TraceWriter.once`가 이미 보장하나, 이중 호출 시
  두 번째가 즉시 반환되는지 확인한다).

### B-3 플러시 예산값 — 200ms는 **추론 기반 잠정 앵커** (실측 아님)

- `internal/config/defaults.go`에 명명 상수를 둔다(REQ-HTF-006). 명명 선례는 같은 파일의
  `DefaultHookDispatcherTimeout`이다.
- **200ms는 측정값이 아니라 추론으로 정한 잠정 앵커다.** `writeEntry`의 실제 소요 시간
  분포(p50/p99)를 측정한 적이 없으므로, "정상 경로 비용이 0에 수렴한다"는 서술은 근거가
  없다 — 이전 판(v0.1.0)의 해당 문장은 실행한 명령도 관측한 출력도 없는 미귀속 주장이었고
  (`verification-claim-integrity.md` §2), 본 개정에서 철회한다.
- **어느 예산의 4%인가 — 두 예산은 다른 것이다** (이전 판이 혼동했다):
  - **정책 예산 5초** — MoAI가 `settings.json` 훅 항목에 두는 기본 timeout 값이며
    (`CLAUDE.local.md` §7), Claude Code가 훅 **프로세스**를 강제 종료하는 상한이다.
    템플릿의 실측 분포는 `{5초 × 24, 10초 × 2, 30초 × 2, 60초 × 3, 120초 × 1}`으로
    5초가 압도적 다수이자 **가장 빡빡한 값**이다. 따라서 최악 케이스를 5초로 잡는
    아래 계산은 유효하다.
  - **Go 디스패처 상수 30초** — `internal/config/defaults.go`의
    `DefaultHookDispatcherTimeout = 30 * time.Second`. `Dispatch` **호출**에 걸리는 상한이며,
    프로세스 수명 상한이 아니다.
  - 플러시는 `Dispatch` **이후** 프로세스 종료 **직전**에 일어나므로 실효 상한은
    (설정된 훅 timeout − 디스패치 소요)이다. 가장 빡빡한 항목(5초)을 기준으로 삼으면
    200ms는 약 4%다. 30초 디스패처 상수를 기준으로 계산한 값이 아니다.
- **run-phase에서 측정으로 교정한다**: AC-HTF-014가 실제 배수 소요를 측정하고, REQ-HTF-013의
  미배수 건수 신호가 운영 중 예산 부족을 드러낸다. 측정 결과가 200ms를 반박하면 상수를
  바꾸는 것이 정상 경로이며, 이는 SPEC 개정 사유가 아니라 상수의 의도된 조정이다.
- **의존 방향 주의**: `internal/hook/trace`는 현재 어떤 내부 패키지도 import하지 않는다.
  이 순수성을 유지하기 위해 상수를 `trace` 패키지가 직접 읽지 않고 **호출자가 인자로
  전달**한다. `internal/hook`은 이미 `internal/config`를 import하므로(`config_change.go` 등)
  `registry.Shutdown()`이 상수를 읽어 `CloseWithTimeout(d)`에 넘기는 형태가 순환을
  만들지 않는다(`internal/config`는 `internal/hook`을 import하지 않음 — 실측 확인).

### B-4 회귀 가드 설계 (승인됨 — 2종)

**가드 이름 고정 (SPEC 소유 · 사전 존재 불가)**

감사에서 드러난 근본 원인은 가드를 **패턴**으로 지목한 것이었다. `Trace.*Flush`는
기존 `internal/hook/trace/writer_test.go`의 `TestTraceWriter_Close_FlushPending`
(구식 무한 대기 `Close()`를 검증하는, 다른 패키지의 기존 테스트)와 충돌해 미수정 트리에서도
"통과"했다. 따라서 본 SPEC의 가드는 **정확한 이름**으로 고정한다.

| 가드 | 테스트 이름 (고정) | 패키지 |
|---|---|---|
| 가드 1 · 레지스트리 행동 가드 | `TestRegistryShutdownFlushesLastHandlerEntry` | `internal/hook` |
| 가드 1E · CLI E2E 행동 가드 | `TestHookCommandFlushesLastHandlerEntry` | `internal/cli` |
| 가드 2 · 정적 호출자 가드 | `TestFlushBarrierHasProductionCaller` | `internal/cli` |

판정은 `-run '^<정확한이름>$'`으로 앵커하며, 접두 일치 패턴을 쓰지 않는다.

**가드 1 · 레지스트리 행동 가드 (핵심, REQ-HTF-008/009)**

- 임시 디렉터리(`t.TempDir()`)를 로그 디렉터리로 삼아 `registry.EnableObservability`를
  **직접 호출**해 관측성을 켜고, **핸들러를 2개 이상** 등록한 뒤 1회 디스패치하고
  teardown을 호출한다.
- 그 후 트레이스 파일을 읽어 **마지막 핸들러의 항목이 존재함**을 단언한다. 마지막 항목이
  현재 사라지는 바로 그 항목이므로, 이 단언이 결함의 직접 반증이다.
- **반증 메커니즘은 단일하다**: `registry.Shutdown()` 본문의 플러시 호출을 무력화한다.
  이 가드는 `internal/hook` 패키지에 살고, **`internal/hook`은 `internal/cli`에 의존하지
  않으므로** `internal/cli/hook.go`의 `defer`를 지우는 것으로는 이 가드를 반증할 수 없다.
  (실측: `go list -deps ./internal/hook | grep -cx '…/internal/cli'` → `0`.
  일치하는 유일한 항목은 하위 패키지 `internal/cli/preference`다.)
  이전 판은 두 반증 경로를 택일로 제시했는데, 그중 CLI 경로를 고른 run-phase 에이전트는
  가드가 여전히 통과하는 것을 보고 **정상 가드를 결함으로 오판**하게 된다.

**가드 1E · CLI E2E 행동 가드 (REQ-HTF-002의 유일한 행동 판정)**

- 가드 1은 `internal/cli`의 배선을 전혀 검증하지 못한다. AC-HTF-006의 grep은 텍스트
  존재만 본다. 그 사이에 **REQ-HTF-002가 행동 판정 없이 남는 공백**이 있었다.
- 이를 메우기 위해 바이너리를 빌드해(`go build -o <tmp>/moai ./cmd/moai`) `t.TempDir()`
  프로젝트(`.moai/config/sections/observability.yaml` + `.moai/logs/`)에 대해
  `hook session-start`를 실행하고, 기록된 트레이스에서 **마지막 핸들러의 항목**을 단언한다.
  이것이 실험 B를 자동화한 형태다.
- 이 가드는 프로세스 경계를 넘으므로 `internal/cli/hook.go`의 `defer` 제거로 반증된다.

**가드 2 · 정적 호출자 가드 (보조, REQ-HTF-010)**

- 플러시 장벽에 대해 `_test.go`가 아닌 프로덕션 호출자가 1개 이상임을 단언한다.
- **선택자는 넓히지 않고 좁힌다.** 이전 판은 "값 복사 소비자를 놓치지 않도록 넓게"라고
  적었으나, 그 결과 선택자가 변별력을 잃고 `internal/web/handlers.go`의 **한국어 주석 한 줄**
  (`httpSrv.Shutdown() 이 … 교착(dead lock) 한다`)에 걸려 미수정 트리에서 통과했다.
  범위를 `internal/hook/` + `internal/cli/`로 한정하고 주석 행을 제외한다.
- **공허 통과 위험을 인정한다** — 이 가드는 "호출이 소스에 존재한다"만 증명하며 효과를
  증명하지 못한다. 그래서 보조이며, 가드 1/1E를 대체하지 않는다.

### B-5 @MX 주석 정정 — ANCHOR → NOTE 강등 (규칙 인용)

`.claude/rules/moai/workflow/mx-tag-protocol.md`:

- "ANCHOR: NEVER auto-delete; demote to NOTE via report"
- 수명주기: "Demoted to NOTE when fan_in drops below 3 (requires report)"

실측 fan_in은 0이고, 수정 후에도 `Close`/`CloseWithTimeout`의 프로덕션 직접 호출자는
`registry.Shutdown` 1개다. **3 미만이므로 ANCHOR 유지 조건을 충족하지 못한다.**
따라서 **삭제가 아니라 `@MX:NOTE`로 강등**하고, 강등 사실을 run-phase @MX 리포트에
기록한다(규칙이 요구하는 "requires report").

- 강등된 NOTE는 실측 사실만 말한다: 이 함수가 유일한 플러시 장벽이라는 것, 그리고
  프로덕션 도달 경로가 `registry.Shutdown` 하나라는 것.
- 거짓 수치 `fan_in=24`와 거짓 서술 `all hook handler teardown paths call Close`는 제거한다.
- `NewTraceWriter` 위의 기존 `@MX:WARN`(고루틴 수명 위험)은 **보존**한다. 위험 자체가
  사라지지 않았고, 규칙상 WARN은 위험이 제거될 때만 삭제한다.

### B-6 goleak 억제 항목의 처리

`internal/hook/main_test.go`의
`goleak.IgnoreTopFunction(".../trace.(*TraceWriter).run")`은 본 결함이 만든 누수를 가리고 있다.

- run-phase는 이 억제를 **제거 시도**하고 전체 패키지 테스트가 통과하는지 확인한다.
- 통과하면 제거하고 주석에서 근거를 갱신한다.
- 통과하지 못하면(= writer를 닫지 않는 테스트가 남아 있으면) **억제를 유지하되, 유지
  사유를 그 자리 주석에 실측 기반으로 다시 적는다.** "SPEC-V3R6 범위 밖"이라는 현재
  사유는 본 SPEC 이후 더 이상 정확하지 않다.
- 어느 쪽이든 결과를 acceptance.md 판정으로 남긴다.

## §C 사전 점검 (Pre-flight)

run-phase 진입 시 1턴 병렬 배치로 확인한다.

```bash
go build ./... && echo BUILD_OK
go test ./internal/hook/... -count=1 2>&1 | tail -20
grep -rn "internal/hook" internal/config/*.go | grep -v "^.*://" | head   # 순환 부재 재확인
grep -rnE '(traceWriter|tw|w)\.Close\(\)' --include='*.go' internal/ cmd/ pkg/ | grep -v _test.go
go test ./internal/hook/trace/... -count=1 -v 2>&1 | head -30
```

기준선: 마지막 grep이 `internal/harness/retention.go` 2건만 반환해야 한다(= 프로덕션
fan_in 0의 재확인). 다른 결과가 나오면 전제가 바뀐 것이므로 계획을 재도출한다.

## §D 제약

- `hook.Registry` 인터페이스 불변(REQ-HTF-007).
- `TraceWriter.Write`의 비차단 계약 불변(REQ-HTF-007).
- 훅 이벤트당 MoAI 5초 예산 준수 — teardown 추가 지연은 예산 대비 무시 가능해야 한다.
- 하드코딩 금지: 예산값은 `internal/config/defaults.go` 상수(CLAUDE.local.md §14).
- `internal/template/templates/` 무수정 — 런타임 Go 코드이지 배포 자산이 아니다.
- 테스트 격리: 로그 디렉터리는 `t.TempDir()`만 사용(CLAUDE.local.md §6).
- 병렬 테스트에서 OTEL 환경변수 `t.Setenv` 금지(CLAUDE.local.md §2).
- 크로스 컴파일 3종(linux/darwin/windows) 통과.
- 코드 주석·식별자는 영어(`code_comments: en`).

## §E 자기검증

run-phase 완료 보고는 다음을 실행 출력과 함께 제시한다.

- E1 — AC 표(PASS/FAIL), 각 행에 실행한 명령과 관측 출력.
- E2 — 3-OS 빌드 결과(`GOOS=linux|darwin|windows go build ./...`).
- E3 — `go test ./internal/hook/... ./internal/cli/... -count=1` 통과.
- E4 — 행동 가드 반증 왕복 증거(teardown 제거 시 FAIL, 복원 시 PASS).
- E5 — `golangci-lint run` 무경고.
- E6 — @MX 리포트(ANCHOR → NOTE 강등 1건).
- E7 — 실제 배수 소요 측정값(AC-HTF-014). 200ms 잠정 앵커의 타당성 판단 근거이며,
  측정 결과가 앵커를 반박하면 상수를 조정한 뒤 조정 사유를 함께 보고한다.

## §F 마일스톤

우선순위 순. 앞선 마일스톤일수록 되돌리기 어려운 결정을 포함한다.

### M1 — 플러시 장벽 도입 (`trace` 패키지)

- `CloseWithTimeout(d time.Duration) error` 추가, 구별 가능한 타임아웃 신호 정의.
- `Close()` 기존 의미 보존.
- 예산 소진 시 **미배수 항목 수**를 구조화 로그 필드로 방출(REQ-HTF-013). 채널에 남은
  건수는 `len(w.ch)`로 얻을 수 있다.
- 단위 테스트: (1) 정상 배수 시 전 항목 기록, (2) 예산 소진 시 예산 내 반환 + 타임아웃
  신호 + 미배수 건수 방출, (3) 이중 호출 멱등.
- 판정: REQ-HTF-003 / REQ-HTF-004 / REQ-HTF-013.

### M2 — 예산 상수화

- `internal/config/defaults.go`에 명명 상수 추가(기준선 200ms), godoc에 근거 기재.
- 인라인 리터럴 부재 확인.
- 판정: REQ-HTF-006.

### M3 — 레지스트리 teardown + CLI 배선

- `registry.Shutdown()` 추가(무동작 안전 + 멱등 + 잠금 비보유 대기).
- `internal/cli/hook.go`의 `runHookEvent` / `runAgentHook` 두 지점에 익명 인터페이스
  타입 단언 + `defer` 배선.
- 판정: REQ-HTF-001 / REQ-HTF-002 / REQ-HTF-005 / REQ-HTF-007.

### M4 — 회귀 가드 3종

- `TestRegistryShutdownFlushesLastHandlerEntry` (가드 1, `internal/hook`).
- `TestHookCommandFlushesLastHandlerEntry` (가드 1E, `internal/cli`) — 바이너리를 빌드해
  프로세스 경계를 넘는 E2E. REQ-HTF-002의 유일한 행동 판정이다.
- `TestFlushBarrierHasProductionCaller` (가드 2, `internal/cli`).
- 세 이름 모두 SPEC 소유이며 미수정 트리에 존재하지 않음을 M4 착수 시 재확인한다
  (`-list '^<이름>$' | grep -cE '^Test'` → `0`).
- 반증 왕복 수행 및 양방향 출력 기록.
- 판정: REQ-HTF-002 / REQ-HTF-008 / REQ-HTF-009 / REQ-HTF-010.

### M5 — @MX 정정 및 goleak 억제 재평가

- `writer.go` `Close` 앞 ANCHOR → NOTE 강등 + 거짓 수치·서술 제거 + 리포트.
- `main_test.go` 억제 제거 시도, 결과에 따라 제거 또는 사유 갱신.
- 판정: REQ-HTF-011.

### M6 — 이식성 및 품질 게이트

- 3-OS 빌드, 전체 테스트, `golangci-lint run`.
- 판정: REQ-HTF-012.

## §G 반패턴

- **AP-1 — 호출 여부만 증명하는 가드**: "Shutdown이 호출되었다"를 세는 가드는 효과를
  증명하지 않는다. 파일에 마지막 항목이 있는지 읽어야 한다.
- **AP-2 — 공허한 테스트 선택자**: `go test -run <패턴>`이 0개를 선택해도 종료 코드는 0이다.
  모든 판정 명령은 선택 집합이 비어 있지 않음을 함께 보여야 한다.
- **AP-2a — 공허 방지 가드 자체가 공허한 경우**: `go test -list` 출력에는 항상 후행
  `ok <pkg>` 줄이 붙으므로 `| grep -c .`는 무조건 1 이상이다. 반드시
  `| grep -cE '^Test'`로 테스트 이름 줄만 센다. (실측: `-list 'ZZZNoSuchTest'`에 대해
  `grep -c .` → `1`, `grep -cE '^Test'` → `0`.)
- **AP-2b — `-list`와 `-run`의 패턴 불일치**: 공허 검사에 `-list '.*'`를 쓰고 판정에
  `-run '<좁은패턴>'`을 쓰면, 검사가 실제 판정 선택자를 구속하지 못한다. 두 곳에
  **동일한 패턴 문자열**을 쓴다.
- **AP-2c — 접두 일치 패턴으로 가드를 지목하는 것**: `Trace.*Flush`는 기존
  `TestTraceWriter_Close_FlushPending`을 잡아 미수정 트리에서도 통과했다. 가드는
  `-run '^<정확한이름>$'`으로 앵커한다.
- **AP-2d — 패키지 의존 방향을 무시한 반증 절차**: `internal/hook` 테스트는
  `internal/cli`의 코드를 실행하지 않으므로, CLI 측 편집으로 반증되지 않는다. 반증
  메커니즘은 가드가 사는 패키지 안에 있어야 한다.
- **AP-2e — 변별력을 잃을 만큼 넓힌 선택자**: 범위를 `internal/`+`cmd/`+`pkg/` 전체로
  넓힌 grep은 무관한 파일의 **주석 한 줄**에 걸려 통과했다. 넓히는 대신 관련 패키지로
  좁히고 주석 행을 제외한다.
- **AP-2f — 사전 존재 부채를 델타 없이 판정하는 것**: "0건"을 요구하는 AC는 기존 코드에
  무관한 매치가 있으면 영영 통과할 수 없고, 범위 밖 편집을 강요한다. 측정된 기준선을
  명시하고 **델타**를 판정한다.
- **AP-3 — 라인 번호 앵커**: `writer.go:84-85` 같은 절대 앵커는 편집 즉시 낡는다. 판정은
  파일 내 문자열(예: `Cleanup entry point for all trace consumers`, `fan_in=24`) 기준으로 건다.
- **AP-4 — 범위 확장**: `observability.enabled` 미독해 결함과 4개 훅 계열 재조사를 이번에
  끌어들이지 않는다(spec.md §4).
- **AP-5 — 예산값 인라인**: `200*time.Millisecond`를 호출 지점에 직접 쓰는 것.
- **AP-6 — ANCHOR 삭제**: 규칙상 자동 삭제 금지. 강등 + 리포트가 정해진 처리다.
- **AP-7 — 유실을 "테스트 통과"로 덮기**: 억제(`goleak.IgnoreTopFunction`)를 유지하는 선택은
  허용되나, 유지 사유를 실측 기반으로 다시 적지 않고 남겨두는 것은 허용되지 않는다.

## §H 교차 참조

- `spec.md` — REQ-HTF-001 … REQ-HTF-012
- `acceptance.md` — AC-HTF-001 … (판정 명령 및 반증 절차)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — ANCHOR 강등 규칙(B-5 인용)
- `.claude/rules/moai/core/verification-claim-integrity.md` — 미관측 주장 금지
- `internal/hook/CLAUDE.md` — 훅 타임아웃 규율, 서브에이전트 경계
- CLAUDE.local.md §6(테스트 격리) · §14(하드코딩 금지)
