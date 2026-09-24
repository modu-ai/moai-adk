---
id: SPEC-HOOK-DIAG-SINK-001
title: "훅 경로 진단 레코드의 파일 싱크 기록 — 무신호 해소"
version: "0.3.1"
status: draft
created: 2026-09-24
updated: 2026-09-24
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hook, logging, observability, diagnostics, slog, fail-open"
tier: M
---

## HISTORY

- 2026-09-24 — plan-phase 최초 저작 (v0.1.0). 카드 t1144 의 1단계(도달성 측정)가 선행됐고,
  이 SPEC 의 모든 수치는 그 측정에 귀속된다(§1.1). 방향은 카드가 지정한 (1) 파일 기록
  하나뿐이며, 방향 (2)(세션 컨텍스트 노출)는 운영자 결정 사항으로 §4 에 이연한다.
- 2026-09-24 — 측정 수치 정정 (v0.1.1). 측정 도구가 `log/slog` 의 패키지 초기화 간선
  7건을 호출 지점으로 집계하고 있었고, 도구를 고쳐 원자료를 재생성했다. 정정 대상은
  세 값뿐이다: 전체 호출 지점 396 → **389**, BOTH 96 → **89**, 훅 경로 합 380 → **373**.
  **HOOK-ONLY 284 · OTHER-ONLY 5 · DEAD 11 · 레벨 분해(warn 이상 201 = 157 + 44,
  info 74, debug 97)는 불변**이므로 REQ-HDS-004 의 범위 경계도, 어떤 요구사항이나 AC 의
  내용도 바뀌지 않는다 — 인용 숫자만 고쳤다. base SHA `60017eb83` 도 불변이다(측정 트리는
  그대로이고 도구만 고쳤다). 아울러 §1.1 레벨 분해표에 동적 레벨 1건(`slog.Log`)과 총계
  행을 넣어 표가 스스로 합을 보이게 했다 — 초판에서 **201 + 171 = 372 ≠ 380** 으로 합이
  닫히지 않았던 것이 이 결함의 단서였다.
- 2026-09-24 — 인수 기준 판정 계층 개정 (v0.2.0). 감사 D2 가 `acceptance.md` 의 테스트 기반
  판정식이 **미수정 트리에서 통과**함을 실측했다(`.moai/reports/t1144/ac-baseline.md`,
  HEAD `2f63958a9`). `go test -run <정규식>` 은 매칭되는 테스트가 없으면 실패가 아니라
  성공하므로, "가드가 아직 없는 상태"와 "가드가 있고 통과하는 상태"가 구별되지 않았다.
  모든 테스트 기반 AC 를 **존재 확인 + 실행** 두 단으로 바꾸고, 유보 문구였던 미수정 트리
  기준선을 실측값으로 교체했다. 아울러 기존 `internal/cli/logging_test.go` 를 읽고 두 가지를
  반영했다 — (a) 훅 목적지를 `io.Discard` 로 단언하는 기존 케이스 3건이 M1 에서 의도적으로
  빨개진다는 사실을 §1.3 과 acceptance.md 에 명시하고 AC-HDS-015 를 신설했으며, (b) 런타임
  stdout/stderr 바이트 대조가 이 저장소 테스트 바이너리에서 판정 불가라는 그 파일의 기록에
  따라 AC-HDS-004 / AC-HDS-009 를 **결정값 포인터 동일성 + 소스 토큰 부재** 판정으로 바꿨다.
  요구사항 15건은 문언·내용 모두 불변이며, AC 는 14건 → 15건이 되었다.
- 2026-09-24 — 판정식 적용 범위 확산 + 역방향 검사 승격 (v0.3.0). 감사가 앞선 진단을 정정
  했다: 공허했던 것은 테스트 기반 AC 전부가 아니라 **존재 확인을 갖지 못한 것들**
  (AC-HDS-004 · 006 · 010 과 접힌 하위 케이스)이며, AC-HDS-001 은 이미 옳은 형식을 갖고
  있었다. 따라서 v0.2.0 이 새로 만든 형식을 물리고 **AC-HDS-001 의 형식
  (`-list … | grep -cE '^Test'` = 1)을 나머지 전부에 확산**했다. 접혀 있던 AC 네 건은
  전용 가드 3개(동시성 · 레벨 경계 · 지연 개방)를 받아 자기 이름으로 앵커한다 — 가드 수
  4 → 7. 아울러 **역방향 검사**(바꾸려는 심볼의 기존 단언을 먼저 세는 검사)를 판정 원칙 6
  으로 승격했고, 그 검사가 `internal/hook` 주석 3곳이 M1 이후 거짓이 됨을 새로 적발해
  AC-HDS-016 을 신설했다. 요구사항 15건은 여전히 불변, AC 는 15건 → 16건.
- 2026-09-24 — 정적 축 판독 범위 확대 (v0.3.1). 감사 D5: v0.2.0 이 런타임 바이트 대조를
  결정값 단언으로 바꾸면서 **소스 토큰 가드가 정적 축의 유일한 담지자**가 됐는데, 기존
  `TestLoggingNeverTargetsStdout` 의 판독 범위가 `logging.go` 한 파일·`os.Stdout` 한
  토큰이라 **이 SPEC 이 새로 쓰는 싱크 파일이 범위 밖**이었다. AC-HDS-004 를 훅 경로
  목적지의 **파일 집합**으로 넓히고 **파일별 금지 토큰**(logging.go → `os.Stdout` 만,
  싱크 파일 → 둘 다)과 **집합 멤버십·읽기 실패 시 실패** 조건을 넣었다. 비대칭의 이유
  (비-훅 분기의 `os.Stderr` 는 정당하다)는 판정 원칙 8 에 기록했다. 아울러 남은 잔여
  위험 — 결정값을 받은 뒤 tee 하는 writer 는 현행 판정 전부를 통과하며 효과 층은
  run-phase 실 바이너리 관측 몫 — 을 §4.1 과 `acceptance.md` §Gaps 에 적었다.
  요구사항·AC 수는 불변(15 / 16). `internal/cli/logging.go` 의 `defaultLogLevel` 행 번호
  인용을 14 → 15 로 정정했다(나머지 8건의 행 번호 인용은 재측정 결과 정확).

## 1. 배경 (Context)

`moai hook` 로 실행되는 모든 호출은 slog 레코드를 **전량 버린다**.
`internal/cli/logging.go` `resolveLoggingDecision` 는 `isHookCommand(args)` 가 참이면
`io.Discard` 를 무조건 반환하며, 같은 파일의 주석이 밝히듯 `MOAI_LOG_LEVEL` 로도 열리지
않는다(`internal/cli/logging.go:50-62`). 결과적으로 `internal/hook` 이 훅 실행 중에
방출하는 진단은 **어디에도 도달하지 않는다.**

이 carve-out 자체는 정당하다. 훅의 stdout 은 구조화 JSON 계약을 나르고 stderr 는
Claude Code 런타임이 읽으므로, 떠도는 레코드 한 줄이 교환을 깨뜨린다. 따라서 이 SPEC 이
건드리는 것은 **목적지**이지 carve-out 이 아니다.

### 1.1 실측 근거 (base SHA `60017eb83`)

아래 수치는 전부 `60017eb83` 트리에서 RTA 호출그래프 간선 BFS 로 측정한 값이며,
인용 시 이 base SHA 를 함께 옮긴다. 원자료:

- `.moai/reports/t1144/verdict.md` — 5절 증거 기록 전문
- `.moai/reports/t1144/slog-sites.tsv` — 호출 지점 389행 전수(분류·위치·둘러싼 함수·호출 대상)
- `.moai/reports/t1144/reachability.log` — 도구 요약 출력

| 분류 | 건수 | 뜻 |
|---|---|---|
| 전체 slog 호출 지점 (`internal/hook`, 비테스트 SSA 명령) | 389 | — |
| HOOK-ONLY | 284 | 훅 디스패치에서만 도달 → **어떤 호출로도 관측 불가** |
| BOTH | 89 | 비-훅 커맨드에서도 도달 → 훅 실행 중에만 소실 |
| OTHER-ONLY | 5 | 훅 경로에 오지 않음(`internal/hook/quality`, `/moai gate` 계열) |
| DEAD | 11 | 어느 cobra 루트에서도 간선이 닿지 않음 |

**훅 한 번 실행 중 소실되는 지점은 373곳**(284 + 89)이다.

레벨 분해 — 이 SPEC 의 기본 범위를 정하는 수치다. 비-훅 실행의 기본 최소 레벨은 warn
이므로(`internal/cli/logging.go:15` `defaultLogLevel`), **warn 이상만이 "원래대로면
사람이 보았을 레코드"** 다.

| 레벨 | HOOK-ONLY | BOTH | 훅 경로 합 |
|---|---|---|---|
| Warn | 151 | 40 | 191 |
| Error | 6 | 4 | 10 |
| **warn 이상 소계** | **157** | **44** | **201** |
| Info | 56 | 18 | 74 |
| Debug | 71 | 26 | 97 |
| **info/debug 소계** | **127** | **44** | **171** |
| 동적 레벨 (`slog.Log`) | 0 | 1 | 1 |
| **훅 경로 총계** | **284** | **89** | **373** |

표는 스스로 합이 맞는다: **201 + 171 + 1 = 373**. 마지막 1건은
`internal/hook/path_resolve.go:87` (`resolveProjectRootFromEnvAt`)의 `slog.Log` 로,
레벨이 실행 시점에 정해지므로 정적으로는 warn/info/debug 어느 칸에도 들어가지 않는다.

> 초판(v0.1.0)은 이 합이 닫히지 않았다(201 + 171 = 372 ≠ 380). 그 불일치가
> 측정 도구 결함의 단서였다 — §HISTORY v0.1.1 참조.

info/debug 171곳은 훅이 아니어도 기본 설정에서 침묵한다. 이들까지 무조건 파일로
쏟아내는 것은 결함 해소가 아니라 **새 비용**이다(§3.2).

### 1.2 확인된 형제 사례

- `closeFactoryHookStore` (`internal/hook/factory_messages.go:41`, t1097 발) — 증인 경로
  탐색에서 비-훅 루트로부터 `NOT reachable`, 즉 HOOK-ONLY 확정.
- `stopHandler.Handle` (`internal/hook/stop.go:36,45,62`) 과
  `userPromptSubmitHandler.Handle` (`internal/hook/user_prompt_submit.go:130`) —
  t1133 이 부분적으로 다룬 `factoryHookBatch` 열화 인박스 경로의 보고 지점이며 모두
  HOOK-ONLY.
- t1133 D3(단계 라벨 재구성)는 이 SPEC 이 흡수하는 후속 항목이다.

파일별 상위 분포(발췌): `session_end.go` 48(HOOK-ONLY), `pre_tool.go` 19,
`post_tool.go` 14, `reflective_write.go` 12, `handoff/persist.go` 12,
`session_start.go` 31(BOTH), `file_changed.go` 9(BOTH).

> 이 분포가 드러내는 자기참조: `.moai/logs/` 의 보존 정리를 수행하는
> `PruneObservationLogs`(`internal/hook/prune_logs.go`)의 실패 보고 `slog.Warn` 역시
> SessionEnd 훅 안에서만 실행되므로 **지금까지 한 줄도 남지 않았다.**

### 1.3 현재 코드 표면 (앵커)

- `internal/cli/logging.go` — `defaultLogLevel`(warn), `loggingDecision{dest, level}`,
  `resolveLogLevel`(`MOAI_LOG_LEVEL`), `resolveLoggingDecision`(훅이면 `io.Discard`),
  `configureLogging`(프로세스 전역 기본 핸들러 설치 1지점), `isHookCommand`.
- `internal/hook/branch_guard.go:47` — `branchGuardAuditRelPath = ".moai/logs/branch-guard-audit.log"`
  (fail-open advisory append 선례).
- `internal/config/log.go` — `.moai/logs/config.log` 에 append 하는 best-effort 기록기,
  실패 시 `slog.Warn` 으로 격하하고 절대 panic 하지 않는 선례.
- `internal/codexadapter/diagnostics.go:16` — `DiagnosticSinkRel = ".moai/logs/codex-adapter.jsonl"`,
  **stderr 를 의도적으로 쓰지 않는** JSONL 싱크 선례.
- `internal/hook/prune_logs.go` — `PruneObservationLogs`(SessionEnd 보존 정리),
  후보 집합은 `trace-*.jsonl` + `task-metrics.jsonl` + `agent-model-audit.jsonl` 로 닫혀
  있고 `PruneStats` 가 건수를 보고한다.
- `internal/config/defaults.go:262` — `DefaultTraceRetentionDays = 30`(보존 상수 선례).
- `internal/config/envkeys.go:485` — `EnvClaudeProjectDir = "CLAUDE_PROJECT_DIR"`.
- `internal/hook/` 주석 3곳 — **현재의 폐기 동작을 사실로 서술하므로 M1 이후 거짓이 된다.**
  `instructions_loaded.go`(감사 행을 따로 쓰는 이유), `config_change.go`(`runReload` 기록
  근거), `factory_messages.go`(열화 반환 직전 — **"Closing that is card t1144"** 라고 적혀
  있고 이 SPEC 이 바로 그 카드다). 지우지 않고 새 사실로 갱신한다(AC-HDS-016).
- `internal/cli/logging_test.go` — **기존 통과 테스트 2종. 둘 다 이 SPEC 과 직접 맞물린다.**
  - `TestLoggingHandlerSelection` — `resolveLoggingDecision` 의 **결정값**(`dest`)을 단언하는
    표 구동 테스트. 훅 케이스 3건(`hook_discards`, `hook_behind_a_flag_discards`,
    `hook_ignores_log_level_env`)이 `io.Discard` 를 기대하므로, REQ-HDS-001 이 목적지를
    바꾸면 **의도적으로 빨개진다**(AC-HDS-015 가 갱신을 판정한다). 회귀가 아니다.
  - `TestLoggingNeverTargetsStdout` — 정적 가드. **판독 범위가 `logging.go` 한 파일이고
    토큰도 `os.Stdout` 하나다**(`os.ReadFile("logging.go")` + `bytes.Contains(src,
    []byte("os.Stdout"))`). 따라서 이 SPEC 이 **새로 쓰는 싱크 파일은 그 범위 밖**이며,
    정적 축을 그 파일까지 넓히는 것은 AC-HDS-004 의 몫이다. 이 가드 자체는 그대로 통과해야
    한다(범위를 좁히거나 삭제해서 통과시키는 것은 미충족 — `plan.md` §D).
    그 주석이 기록하듯 런타임 동일성 검사는 이 저장소에서 판정 불가다:
    `go test ./...` 아래 테스트 바이너리가 두 스트림에 같은 `*os.File` 을 받아
    `dest == os.Stdout == os.Stderr` 가 같은 포인터로 측정됐다. REQ-HDS-002 의 판정이
    바이트 대조가 아니라 **결정값 + 소스 토큰**인 이유다(acceptance.md 판정 원칙 7).

## 2. 목적 (Purpose)

1. 훅 실행 중 방출되는 **warn 이상 진단이 사람이 사후에 읽을 수 있는 자리**에 남는다.
2. stdout / stderr 계약은 **바이트 단위로 불변**이다 — 목적지만 늘어난다.
3. 훅이 기록에 실패해도 훅은 실패하지 않는다(fail-open).
4. 기록이 `.moai/logs/` 를 무한히 키우지 않는다 — 기존 보존 정리 기제에 편입된다.
5. 이 결함이 다시 도입되면 **행동 수준에서** 실패하는 가드가 남는다.

이 SPEC 이 닫히기 전에는 "훅에서 아무 이상도 보고되지 않았다"는 어떤 주장도 성립하지
않는다. 훅 경로의 무신호는 정상의 증거가 아니기 때문이다.

## 3. 요구사항 (GEARS)

### 3.1 목적지

- **REQ-HDS-001** (Ubiquitous)
  `moai hook` 실행 중 방출되어 최소 레벨을 통과한 slog 레코드는 프로젝트 루트 하위
  `.moai/logs/` 의 append 모드 파일 싱크에 기록되어야 한다.

- **REQ-HDS-002** (Unwanted — 계약 보전)
  훅 경로의 로깅 목적지는 stdout 또는 stderr 로 어떤 레코드도 내보내서는 안 된다.
  `MOAI_LOG_LEVEL` 을 포함한 어떤 입력으로도 두 스트림이 열려서는 안 된다.

  > 근거: stdout 은 훅 JSON 계약, stderr 는 Claude Code 런타임 판독 대상이다. 이
  > carve-out 은 이 SPEC 의 수정 대상이 아니라 보전 대상이다.

- **REQ-HDS-003** (Where — 능력 게이트)
  **Where** 프로젝트 루트를 해소할 수 없거나 싱크 파일을 열 수 없는 경우, 훅은 기존
  폐기 동작(`io.Discard`)으로 강등되어 계속 실행되어야 하며, 오류를 반환하거나 종료
  코드를 바꾸거나 훅 출력을 변형해서는 안 된다.

### 3.2 레벨 경계

- **REQ-HDS-004** (Ubiquitous)
  훅 싱크의 기본 최소 레벨은 비-훅 경로와 동일한 `defaultLogLevel`(warn)이어야 한다.
  info / debug 레코드는 기본 구성에서 싱크에 기록되어서는 안 된다.

  > 근거: warn 이상 201곳이 "원래대로면 보였을 레코드"이고, info/debug 171곳은 훅이
  > 아니어도 기본값에서 침묵한다(§1.1, base `60017eb83`). 후자를 기본으로 기록하면
  > 훅 실행마다 `.moai/logs/` 가 자라며, 이는 결함 해소가 아니라 새 비용이다.

- **REQ-HDS-005** (Event-driven)
  **When** `MOAI_LOG_LEVEL` 이 최소 레벨을 낮추면, 훅 싱크는 그 레벨 이상을 기록해야
  한다. 이 변수는 **레벨만** 지배하며 목적지를 바꿔서는 안 된다(REQ-HDS-002 불변).

### 3.3 동시성과 비용

- **REQ-HDS-006** (Ubiquitous)
  훅 프로세스는 일회성이고 여러 개가 동시에 실행될 수 있으므로, 싱크는 동시 append 를
  견뎌야 한다. 레코드는 한 줄 단위이고 한 레코드는 한 번의 쓰기로 방출되어야 하며,
  동시 기록이 다른 프로세스의 레코드를 훼손해서는 안 된다.

- **REQ-HDS-007** (Unwanted — 무비용 침묵)
  기록할 레코드가 한 건도 없는 훅 호출은 싱크 파일을 생성하거나 열어서는 안 된다.

  > 근거: 훅은 세션당 수십~수백 회 발화한다. 매 호출이 파일을 건드리면 침묵하는
  > 호출조차 비용을 낸다.

- **REQ-HDS-008** (Unwanted)
  싱크 기록은 훅의 5초 예산 안에서 경계 없는 대기를 도입해서는 안 된다.

### 3.4 증가 억제

- **REQ-HDS-009** (Ubiquitous)
  싱크 파일은 기존 SessionEnd 보존 정리(`PruneObservationLogs`,
  `internal/hook/prune_logs.go`)의 후보 집합에 편입되어야 하며, 새 정리 기제를
  도입해서는 안 된다. 정리 결과는 기존 `PruneStats` 표면으로 보고되어야 한다.

- **REQ-HDS-010** (Ubiquitous)
  보존 임계값과 싱크 경로는 명명된 상수여야 한다(경로는 `internal/hook`/`internal/cli`
  의 패키지 상수, 보존일수는 `internal/config/defaults.go`). 호출 지점에 인라인
  리터럴로 나타나서는 안 된다(CLAUDE.local.md §14).

### 3.5 회귀 가드

- **REQ-HDS-011** (Ubiquitous — 행동 가드, 핵심)
  테스트 스위트는 훅 경로에서 warn 레코드를 방출한 뒤 싱크 파일에 **그 레코드가 실제로
  존재함**을 단언해야 한다. 호출 여부가 아니라 효과를 본다.

- **REQ-HDS-012** (Ubiquitous — 반증 가능성)
  행동 가드는 반증 가능해야 한다 — 훅 경로의 싱크 라우팅을 제거하면 가드가 실패해야
  하며, 그 반증 절차가 `acceptance.md` 에 명시되어야 한다.

- **REQ-HDS-013** (Ubiquitous — 계약 가드)
  테스트 스위트는 훅 경로에서 레코드가 방출되어도 stdout 과 stderr 가 변하지 않음을
  단언해야 한다(REQ-HDS-002 의 기계적 판정).

- **REQ-HDS-014** (Ubiquitous — fail-open 가드)
  테스트 스위트는 싱크를 열 수 없는 조건에서 훅이 정상 종료하고 출력이 불변임을
  단언해야 한다(REQ-HDS-003 의 기계적 판정).

### 3.6 이식성

- **REQ-HDS-015** (Ubiquitous)
  변경된 코드는 linux / darwin / windows 세 대상에 대해 빌드되어야 한다.

## 4. 범위 제외 (Exclusions)

### Out of Scope — 방향 (2): 세션 컨텍스트 노출 (운영자 결정, 이연)

- 열화(degraded) 상태를 **세션 컨텍스트로 알리는** 안이다. `bind` 는 이미 이 방식으로
  알린다(`internal/hook/user_prompt_submit.go:153-160`).
- 이 SPEC 은 이 방향의 요구사항도 인수 기준도 두지 않는다. 레인 운영자가 보는 표면이
  바뀌므로 **운영자 결정 사항**이며, 카드가 명시적으로 이 SPEC 의 범위 밖에 두었다.
- 트레이드오프: 파일 싱크는 사후 판독용이라 **사람이 찾아가야** 보이고, 세션 컨텍스트
  알림은 즉시 보이지만 **레인의 주의 예산을 소비**한다. 둘은 배타적이지 않다.
- **미측정 사실**: 세션 컨텍스트 알림 후보는 위 373곳 전부가 아니라 **핸들러가 degraded
  로 빠지는 분기**에 한정되며, 그 수는 t1144 1단계에서 세지 않았다. 이 방향을 택한다면
  **그 수를 재는 것이 첫 일이다.**

### Out of Scope — stdout / stderr carve-out 의 재개방

- 훅 경로에서 stdout 또는 stderr 로 레코드를 내보내는 설계는 **기각한다**(REQ-HDS-002).
  stdout 은 JSON 계약을, stderr 는 런타임 판독을 깨뜨린다.

### Out of Scope — info / debug 의 기본 기록

- 훅 경로 info 74곳 + debug 97곳(base `60017eb83`)을 기본으로 파일에 기록하지 않는다.
  명시적 레벨 opt-in(`MOAI_LOG_LEVEL`) 뒤에 둔다(REQ-HDS-004 / REQ-HDS-005).

### Out of Scope — 훅 디스패처 · 레지스트리 · 로깅 결정의 argv 기반 형태

- `internal/hook` 의 디스패처와 레지스트리 구조, `isHookCommand` 의 argv 기반 판별
  방식은 변경 대상이 아니다. 이 SPEC 은 `resolveLoggingDecision` 이 훅 경로에 대해
  **어떤 목적지를 고르는가**만 바꾼다.

### Out of Scope — 호출 지점의 내용 심사

- warn/error 157곳이 각각 어떤 상황을 보고하는지 읽어 분류하는 작업은 하지 않는다.
  개중에는 실제 이상 신호도 정상 분기의 잡음도 있을 것이며, 그 심사는 기록이 실제로
  남기 **시작한 이후에만** 의미를 가진다.

### Out of Scope — DEAD 11곳의 처분

- `contract.go` `Validate` 6곳, `generic_handler.go` 2곳, `subagent_start.go` 3곳은
  어느 cobra 루트에서도 도달하지 않았으나, 이는 "사문"이 아니라 "이 방법으로 경로를
  찾지 못했다"는 기록이다. 삭제도 보존 판정도 이 SPEC 의 소관이 아니다.

### Out of Scope — 배포 템플릿 트리

- 본 변경은 Go 런타임 코드이며 배포 자산이 아니다. `internal/template/templates/`
  하위는 일절 수정하지 않는다.

## 4.1 이 SPEC 의 판정 계층이 지키지 않는 것 (Gap)

범위 제외(§4)가 "무엇을 만들지 않는가"라면, 이 절은 **"무엇을 판정하지 못하는가"** 다.

- **[HARD] `-list` 존재 확인은 이름을 지킬 뿐 내용을 지키지 않는다.** 모든 테스트 기반 AC 는
  `go test -list` 가 가드 이름을 실제로 출력하는지를 1단으로 삼는다(acceptance.md 판정 원칙 3).
  이 단은 "가드가 사라지거나 개명되는" 실패를 막지만, **본문이 빈 테스트도 1단과 2단을 모두
  통과한다.** 즉 판정 계층은 가드의 **존재와 통과**를 보장할 뿐 **적합성**을 보장하지 않는다.
- **적합성은 run-phase 감사 몫이다.** 각 가드가 실제로 자기 AC 의 Given-When-Then 을 단언하는지는
  사람이 코드를 읽어 판정한다. AC-HDS-002 의 반증 절차가 이 Gap 을 **부분적으로만** 메운다 —
  가드 1 하나에 대해, 그것도 행동을 겨눴다는 사실까지만 보인다.
- **AC-HDS-016(주석 갱신)의 판정도 대부분 사람이 읽어야 한다.** 기계적 축은
  `card t1144` 문구가 미완료 서술로 남아 있지 않을 것 하나뿐이고, 나머지 두 주석이 *새
  사실을 정확히* 서술하는지는 감사 판단이다.
- **[HARD] REQ-HDS-002 의 판정은 효과가 아니라 결정값 + 소스 토큰이다.** 요구사항은 "훅
  경로가 stdout / stderr 로 **어떤 레코드도 내보내지 않는다**"는 효과를 구속하지만, 판정
  (AC-HDS-004)은 그 한 층 아래 — `dest` 포인터가 두 스트림 어느 쪽도 아님 + 파일 집합의
  토큰 부재 — 를 본다. 결정값을 받은 뒤 두 스트림으로 **tee 하는 writer 는 현행 판정
  전부를 통과한다.** 요구사항의 구속력은 그대로이고 판정만 좁다. 효과 층은 **run-phase 의
  실 바이너리 관측 몫**이며, 그 관측 없이 AC-HDS-004 PASS 를 "아무것도 나가지 않음"의
  증거로 읽어서는 안 된다. 상세: `acceptance.md` §Gaps.
- 이 Gap 들은 알려진 채로 수용한다. 내용 적합성을 기계적으로 판정하려면 가드의 가드가
  필요하고, 효과 층을 판정하려면 실 바이너리 관측이 필요하다 — 둘 다 이 SPEC 의 범위를 넘는다.

## 5. 참조

- `.moai/reports/t1144/verdict.md` · `slog-sites.tsv` · `reachability.log` — 측정 출처(base `60017eb83`)
- `internal/cli/logging.go` — `resolveLoggingDecision` / `defaultLogLevel` / `configureLogging` / `isHookCommand`
- `internal/hook/branch_guard.go` — `branchGuardAuditRelPath`(fail-open advisory 선례)
- `internal/config/log.go` — `.moai/logs/config.log` best-effort append 선례
- `internal/codexadapter/diagnostics.go` — `DiagnosticSinkRel`(stderr 비사용 JSONL 싱크 선례)
- `internal/hook/prune_logs.go` — `PruneObservationLogs` / `PruneStats`(보존 정리 선례)
- `internal/config/defaults.go` — `DefaultTraceRetentionDays`(보존 상수 선례)
- `internal/config/envkeys.go` — `EnvLogLevel` / `EnvClaudeProjectDir`
- `.claude/rules/moai/core/verification-claim-integrity.md` — 미관측 주장 금지
