---
id: SPEC-HOOK-TRACE-FLUSH-001
title: 훅 트레이스 비동기 기록의 종료 시 유실 복구 및 @MX 주석 정정
version: 0.2.0
status: completed
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: High
phase: v3.0.2
module: hook
lifecycle: spec-anchored
tags: "hook, observability, trace, flush, goroutine, mx-tag"
tier: M
---

## HISTORY

- 2026-08-02 — plan-phase 최초 저작 (v0.1.0). 통제 실험(실험 A/B)으로 유실을 실측한 뒤
  작성했다. 수정 메커니즘(경계 대기 Close + defer teardown), 회귀 가드 2종(행동 가드 +
  호출자 가드), 범위(플러시 복구 + @MX 정정)는 저작 이전에 사용자 승인을 받은 결정이다.
- 2026-08-02 — 독립 감사 FAIL(0.66, Testability 0.45) 반영 개정 (v0.2.0). 요구사항 계층
  변경분만 기록한다: REQ-HTF-001의 무조건 "전량 기록" 주장이 REQ-HTF-003의 예산 포기와
  모순되어 잔여 유실이 침묵하는 결함(D4)을 해소했고 — REQ-001을 예산 범위로 한정하고
  REQ-HTF-013(미배수 건수를 담은 관측 가능 신호)을 신설했다 —, REQ-HTF-003의 GEARS
  태그를 `Event-driven`으로 정정했으며(D9), REQ-HTF-011이 AC-HTF-010의 강등+리포트
  판정보다 약해 AC가 REQ를 앞지르던 불일치를 해소했다(D12). 프론트매터
  `priority`/`phase`를 스키마 열거값과 릴리스 타깃으로 교정했다(D10/D11).
  요구사항 총수 12건 → 13건. 범위 제외는 불변이다.

## 1. 배경 (Context)

`internal/hook/trace.TraceWriter`는 비동기 기록기다. `Write()`는 버퍼 채널
(`channelCapacity = 100`)에 항목을 넣고 즉시 반환하며, 실제 디스크 기록은 배경
고루틴 `run()`이 수행한다. 이 설계에서 **플러시 장벽은 `Close()` 하나뿐이다**
(`internal/hook/trace/writer.go`, `func (w *TraceWriter) Close() error`).

그런데 `Close()`에는 **프로덕션 호출자가 하나도 없다**. `moai hook <event>`는
표준입력을 읽고 → 디스패치하고 → 표준출력을 쓰고 → 종료하는 일회성 프로세스이므로,
배경 고루틴이 기록을 마치기 전에 프로세스가 사라진다. 결과적으로 관측 데이터가
소리 없이 유실된다.

### 1.1 실측 근거 (통제 실험)

설치 바이너리 `moai` v3.0.1(`/Users/goos/go/bin/moai`), `.moai/config/sections/observability.yaml`과
`.moai/logs/`만 둔 스크래치 프로젝트에서 측정했다.

- **실험 A — 양성 대조군**: `moai hook pre-tool`(PreToolUse, 등록 핸들러 1개)을 동일
  `session_id`로 121회 호출. `.moai/logs/trace-<sid>.jsonl`이 **생성되었으나 0바이트**.
  파일이 생성되었다는 사실은 고루틴이 `writeEntry`의 `os.OpenFile`까지 도달했음을
  증명한다 — 즉 파이프라인은 살아 있고, 최종 기록만이 프로세스 종료와 경쟁해서 진다.
  증거원 자체가 비어 있는(관측이 꺼진) 경우가 아님을 이 대조군이 배제한다.
- **실험 B — 손실률**: `moai hook session-start`(SessionStart, 등록 핸들러 3개)를 10회
  호출. 기대 30건 대비 **관측 8건(73% 유실)**. 8건 전부가
  `handler = "*hook.sessionStartHandler"`(1번 핸들러)였고, 2번(`autoUpdateHandler`)과
  3번(`handoffInjectHandler`)은 기대 20건 중 **0건**이었다.
- **유실의 구조적 편향**: 앞선 핸들러의 항목은 뒤 핸들러가 실행되는 동안 플러시될
  시간을 얻지만, **마지막 핸들러의 항목은 항상 프로세스 종료와 경쟁**한다. 판정에
  가장 중요한 항목(차단 결정을 내리는 마지막 핸들러)이 가장 잘 사라진다.
- **프로덕션 정황 증거**: `.moai/logs/` 트레이스 파일 460개, 총 3,098줄. 이벤트 분포는
  `SessionStart` 919건(핸들러 3개), `PreToolUse` 455건(핸들러 1개)이며, 460개 세션에
  대해 세션당 약 1건이다. PreToolUse가 세션당 수십~수백 회 발화함에도 그렇다 —
  위 편향과 일치한다.

### 1.2 부수 결함 — 거짓 @MX 주석

`internal/hook/trace/writer.go`의 `Close` 직전 주석은 다음과 같이 주장한다.

```
// @MX:ANCHOR: [AUTO] Cleanup entry point for all trace consumers; must be called to drain background goroutine
// @MX:REASON: fan_in=24, all hook handler teardown paths call Close; omitting this call causes the run() goroutine to leak
```

`fan_in=24`는 거짓이다. 실측 프로덕션 fan_in은 0이다.

```
grep -rnE '(traceWriter|tw|w)\.Close\(\)' --include='*.go' internal/ cmd/ pkg/ | grep -v _test.go
```

는 2건만 반환하며, 둘 다 `internal/harness/retention.go`의 gzip writer로 타입이 다르다.
이 주석은 "정리 경로가 이미 존재한다"는 잘못된 안심을 주어 결함을 은폐해 왔다.

### 1.3 현재 코드 표면 (앵커)

- `internal/hook/trace/writer.go` — `TraceWriter`(ch/done/once/closed),
  `NewTraceWriter`(`go w.run()` 기동), `Write`(비차단 enqueue, 가득 차면 드롭),
  `Close`(`once.Do(close(ch))` 후 **무한 대기** `<-w.done`), `run`, `writeEntry`, `rotate`.
- `internal/hook/registry.go` — `registry`가 `traceWriter *trace.TraceWriter`와 `logDir`를
  보유. `Dispatch`가 `ensureTraceWriter(resolveTraceSessionID(input))`를 호출하고 핸들러마다
  `writeTrace(...)`를 호출한다. `SetTraceWriter` / `EnableObservability` / `ensureTraceWriter` /
  `writeTrace`가 존재하나 **`registry`에도 `hook.Registry` 인터페이스에도 teardown 메서드가 없다**.
- `internal/cli/hook.go` — 프로덕션 `Dispatch` 호출 지점은 `runHookEvent`와 `runAgentHook`
  둘뿐이며, 둘 다 이미 `context.WithTimeout(cmd.Context(), config.DefaultHookDispatcherTimeout)` +
  `defer cancel()` 구조를 갖는다.
- `internal/cli/deps.go` — `enableObservabilityIfConfigured`가 익명 `observabilityEnabler`
  인터페이스를 `hook.Registry`에 타입 단언한다. **인터페이스를 넓히지 않고 선택적 능력을
  추가하는 기존 선례**다.
- `internal/hook/main_test.go` — `goleak.IgnoreTopFunction(".../trace.(*TraceWriter).run")`이
  본 결함이 유발하는 고루틴 누수를 현재 억제하고 있다.

## 2. 목적 (Purpose)

관측 데이터를 신뢰할 수 있게 만든다. 구체적으로:

1. 정상 경로에서 훅 프로세스가 종료하기 전에 적재된 트레이스 항목이 **전부** 디스크에 도달한다.
2. 느린 디스크가 훅 핫 경로를 무한히 붙잡지 못한다 — 대기는 경계가 있다.
3. 이 결함이 다시 도입되면 **행동 수준에서** 실패하는 가드가 남는다.
4. 코드가 스스로에 대해 참인 것만 말한다(@MX 정정).

본 SPEC이 닫히기 전에는 "훅 N개가 발화하지 않았다"는 어떤 주장도 성립하지 않는다.
트레이스 카운트는 현재 발동 증거가 아니기 때문이다.

## 3. 요구사항 (GEARS)

### 3.1 플러시 보장

- **REQ-HTF-001** (Ubiquitous)
  훅 디스패치 프로세스는 종료 전에 **teardown 시점까지 적재된 트레이스 항목을 플러시
  예산 범위 내에서** 디스크에 기록해야 한다.

  > 한정의 근거: 무조건적인 "전량 기록"은 REQ-HTF-003의 예산 포기와 직접 모순된다.
  > 예산이 소진되어 배수를 포기하면 아직 채널에 남은 항목은 유실되며, 그 유실이
  > 침묵하면 본 SPEC이 없애려는 결함 부류가 그대로 재현된다. 따라서 REQ-001은 예산
  > 범위로 한정하고, 예산 밖 잔여분은 REQ-HTF-013이 **측정 가능**하게 만든다.

- **REQ-HTF-013** (Event-driven)
  **When** 플러시 예산이 소진되어 배수가 중단되면, teardown 경로는 **미배수 항목 수를
  담은 관측 가능한 신호**(구조화 로그 필드)를 방출해야 한다.

  > 잔여 유실을 0으로 만들 수는 없지만(비동기 설계의 대가), 유실이 **침묵**해서는 안 된다.
  > 미배수 건수가 로그에 남으면 예산값이 부족한지 여부를 사후에 판정할 수 있고,
  > §plan.md §B-3의 잠정 예산값을 실측으로 교정할 근거가 된다.

- **REQ-HTF-002** (Event-driven)
  **When** 훅 디스패치가 완료되면, 훅 CLI 진입점은 레지스트리의 teardown 경로를 호출하여
  트레이스 기록기의 플러시 장벽을 통과시켜야 한다.

- **REQ-HTF-003** (Event-driven)
  **When** 플러시 예산이 배경 고루틴의 배수 완료 전에 소진되면, 트레이스 기록기는 대기를
  포기하고 구별 가능한 타임아웃 신호를 반환해야 하며, 호출자를 예산 이상으로 붙잡아서는 안 된다.

- **REQ-HTF-004** (Unwanted)
  트레이스 기록기는 플러시 장벽에서 무한 대기해서는 안 된다.

- **REQ-HTF-005** (Where — 능력 게이트)
  **Where** 관측성이 비활성이어서 트레이스 기록기가 존재하지 않는 경우, teardown 경로는
  무해한 무동작이어야 하며 오류를 발생시켜서는 안 된다.

### 3.2 구성 규율

- **REQ-HTF-006** (Ubiquitous)
  플러시 예산은 `internal/config/defaults.go`의 명명된 상수여야 하며, 호출 지점에 인라인
  리터럴로 나타나서는 안 된다(CLAUDE.local.md §14 하드코딩 금지).

- **REQ-HTF-007** (Unwanted)
  본 변경은 `hook.Registry` 인터페이스를 넓혀서는 안 되며, `TraceWriter.Write`의 비차단
  계약을 바꿔서도 안 된다.

### 3.3 회귀 가드

- **REQ-HTF-008** (Ubiquitous — 행동 가드, 핵심)
  테스트 스위트는 관측성을 켠 레지스트리를 통해 복수 핸들러를 디스패치하고 teardown을
  실행한 뒤, 트레이스 파일에서 **마지막 핸들러의 항목이 실제로 존재함**을 단언해야 한다.

- **REQ-HTF-009** (Ubiquitous — 반증 가능성)
  행동 가드는 반증 가능해야 한다 — teardown 호출을 제거하면 가드가 실패해야 하며, 그
  반증 절차가 acceptance.md에 명시되어야 한다.

- **REQ-HTF-010** (Ubiquitous — 호출자 가드, 보조)
  테스트 스위트는 플러시 장벽에 대해 `_test.go`가 아닌 프로덕션 호출자가 1개 이상
  존재함을 단언해야 한다.

### 3.4 주석 정직성

- **REQ-HTF-011** (Ubiquitous)
  `internal/hook/trace/writer.go`의 `Close` @MX 주석은 실측된 실제 fan_in을 반영해야 하며,
  거짓 수치(`fan_in=24`)와 거짓 서술("all hook handler teardown paths call Close")을
  포함해서는 안 된다. 실측 fan_in이 3 미만이므로 해당 주석은 `@MX:ANCHOR`에서
  `@MX:NOTE`로 **강등**되어야 하고, 그 강등 사실이 run-phase @MX 리포트에 **기재**되어야
  한다(`.claude/rules/moai/workflow/mx-tag-protocol.md` — "demote to NOTE via report",
  "Demoted to NOTE when fan_in drops below 3 (requires report)").

### 3.5 이식성

- **REQ-HTF-012** (Ubiquitous)
  변경된 코드는 linux / darwin / windows 세 대상에 대해 빌드되어야 한다.

## 4. 범위 제외 (Exclusions)

### Out of Scope — observability.enabled 키 미독해 결함

- `internal/cli/deps.go`의 `enableObservabilityIfConfigured`는
  `.moai/config/sections/observability.yaml`의 **존재 여부**만 검사하고
  `observability.enabled` 키를 읽지 않는다. 따라서 `enabled: false`로 설정해도 트레이싱이
  꺼지지 않는다.
- 본 SPEC은 이 결함을 수정하지 않는다. 유실 복구와 독립적인 별개 결함이며, 함께 묶으면
  회귀 가드의 판정 범위가 흐려진다. 후속 SPEC 소관.
- **결합 우려 해소 (가드 1에 한정)**: 이 이연이 **가드 1**(AC-HTF-007,
  `TestRegistryShutdownFlushesLastHandlerEntry`)을 무력화하지 않는다. 가드 1은
  `registry.EnableObservability(logDir)`를 직접 호출하여 관측성을 켜므로,
  `enableObservabilityIfConfigured`(설정 파일 **존재**만 검사하는 버그 있는 경로)를
  아예 경유하지 않는다.
- **가드 1E는 반대로 그 경로를 의도적으로 통과한다**: AC-HTF-007E는 프로세스 경계를
  넘는 E2E이므로, 설정 **파일을 배치**하는 방식으로 관측성을 켠다. 이는
  `internal/cli/deps.go`의 `enableObservabilityIfConfigured`를 그대로 경유하며, 그
  함수는 `os.Stat(cfgPath)` 실패 시 조기 반환할 뿐 `enabled` 키를 읽지 않는다. 현재는
  파일 존재만으로 켜지므로 가드 1E가 성립한다.
  → **후속 SPEC 주의**: `observability.enabled`를 실제로 존중하도록 고치는 후속 작업은
  가드 1E의 픽스처가 `enabled: true`를 담도록 함께 갱신해야 한다. 그러지 않으면 가드 1E는
  관측성을 켜지 못한 채 **조용히 무의미해진다**. 실패 양상은 둘이며 위험도가 다르다 —
  트레이스가 아예 기록되지 않아 "마지막 항목 존재" 단언이 결함과 무관하게 실패하는 경우는
  시끄럽게 드러나지만, 픽스처가 느슨해 **공허 통과**하는 경우는 침묵하므로 **이쪽이 더
  위험하다**.

### Out of Scope — 미검증 훅 4개 계열 재조사

- `security-*`, sync-gate, workflow 훅 5종, ci-watch 계열의 실제 발동 여부는 여전히
  미검증이다.
- 본 SPEC은 이를 조사하지 않는다. 관측성이 신뢰 가능해진 **이후에만** 그 조사가 의미를
  가지므로 순서상 본 SPEC 이후여야 한다.

### Out of Scope — 트레이스 기록 아키텍처의 동기화 전환

- `Write`를 완전 동기로 바꾸는 재설계는 채택하지 않는다(§ plan.md §B 기각 대안 (a)).
- 채널 용량, 회전 정책, JSONL 스키마, 요약 로직(`summary.go`)은 변경 대상이 아니다.

### Out of Scope — Write의 @MX:ANCHOR 주석(`fan_in=20`)

- `writer.go`의 `Write` 앞 `@MX:ANCHOR ... fan_in=20` 주석 역시 미검증 수치이나, 사용자가
  승인한 범위는 `Close`의 주석(파일 내 `Cleanup entry point for all trace consumers` 문구를
  담은 절)뿐이다.
- 별도 후속 항목으로 기록한다.

### Out of Scope — 배포 템플릿 트리

- 본 변경은 Go 런타임 코드이며 배포 자산이 아니다. `internal/template/templates/` 하위는
  일절 수정하지 않는다.

## 5. 참조

- `internal/hook/trace/writer.go` — `TraceWriter`, `Close`, `run`, `writeEntry`
- `internal/hook/registry.go` — `Dispatch`, `ensureTraceWriter`, `writeTrace`
- `internal/cli/hook.go` — `runHookEvent`, `runAgentHook`
- `internal/cli/deps.go` — `enableObservabilityIfConfigured`(선택적 능력 타입 단언 선례)
- `internal/hook/main_test.go` — `goleak.IgnoreTopFunction` 억제 항목
- `internal/config/defaults.go` — `DefaultHookDispatcherTimeout`(상수 명명 선례)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — ANCHOR 자동 삭제 금지 / NOTE 강등 규칙
- `internal/hook/CLAUDE.md` — 훅 5초 예산 규율
- `.claude/rules/moai/core/verification-claim-integrity.md` — 미관측 주장 금지
