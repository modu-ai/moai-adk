# SPEC-HOOK-DIAG-SINK-001 — 구현 계획

> 절 순서는 **되돌리기 어려움(decision reversibility)** 기준이다. 바뀔 가능성이 큰 결정
> (목적지 형태 · 싱크 이름 · 레벨 경계 · 가드 설계)을 앞에 두고, 기계적 편집을 뒤로 미룬다.

## §A Tier 판정

**Tier M.** 산출물: `spec.md` + `plan.md` + `acceptance.md` (+ `progress.md`).

근거:

- 편집 대상 프로덕션 파일 4~5개: `internal/cli/logging.go`(목적지 결정),
  새 싱크 파일 1개, `internal/hook/prune_logs.go`(후보 집합 + `PruneStats`),
  `internal/config/defaults.go`(보존 상수), 필요 시 `internal/hook/session_end.go`(호출부).
- **[HARD] 기존 자산 갱신 2건** — 역방향 검사(`acceptance.md` 판정 원칙 6)가 적발했다:
  - `internal/cli/logging_test.go` — `TestLoggingHandlerSelection` 의 `hook_discards` /
    `hook_behind_a_flag_discards` / `hook_ignores_log_level_env` 세 건 기대값 갱신
    (**이 SPEC 이 의도적으로 뒤집는 단언이다**). AC-HDS-015.
  - `internal/hook/` 주석 3곳 — `instructions_loaded.go` · `config_change.go` ·
    `factory_messages.go` 가 현재의 `io.Discard` 동작을 **사실로** 서술하므로 M1 이후
    거짓이 된다. `factory_messages.go` 의 "Closing that is card t1144" 는 이 SPEC 이 그
    카드다. AC-HDS-016.
- 테스트 7개(행동 · 계약 · fail-open · 정리 편입 · 동시성 · 레벨 경계 · 지연 개방).
- 새 타입/함수 시그니처가 1~2개 늘어나므로 Tier S(단일 관심사 사소 변경)가 아니다.
- 사용자 표면(CLI 플래그 · 훅 출력 · 설정 스키마) 변경이 없고 마일스톤이 5개 이하이므로
  Tier L 도 아니다. `design.md` / `research.md` 는 만들지 않는다.

REQ 15건 / AC 15건 — Tier M 상한(각 16)의 범위 안이다. AC 는 v0.2.0 에서 14 → 15 로
늘었다(AC-HDS-015, 기존 테스트 갱신 판정).

## §B 설계 결정 (되돌리기 어려운 순)

### B-1 목적지의 형태 — `loggingDecision.dest` 를 파일로 교체 (핵심 결정)

현재:

```go
func resolveLoggingDecision(args []string) loggingDecision {
    if isHookCommand(args) {
        return loggingDecision{dest: io.Discard, level: defaultLogLevel}
    }
    return loggingDecision{dest: os.Stderr, level: resolveLogLevel()}
}
```

도입 방향: 훅 분기가 `io.Discard` 대신 **지연 개방(lazy-open) append writer** 를 반환한다.

- `dest` 는 이미 `io.Writer` 이므로 **타입 표면이 바뀌지 않는다.** 설치 지점
  (`configureLogging`)도 그대로다. 이것이 이 형태를 고른 주된 이유다.
- 지연 개방: 첫 `Write` 가 올 때 비로소 `MkdirAll` + `OpenFile(O_APPEND|O_CREATE|O_WRONLY)`
  를 수행한다. 레코드가 한 건도 없으면 파일을 만들지 않는다(REQ-HDS-007).
- 개방 실패 시 내부 상태를 "폐기"로 고정하고 이후 모든 쓰기를 삼킨다 — 오류를 위로
  전파하지 않는다(REQ-HDS-003). `internal/config/log.go` 가 이미 같은 자세를 취한다.
- 프로세스 종료 시 파일 닫기: 훅은 일회성 프로세스이고 각 레코드가 한 번의 쓰기로
  끝나므로 배경 고루틴도 플러시 장벽도 필요 없다(REQ-HDS-008). `Close` 를 둘지는
  run-phase 재량이되, **경계 없는 대기를 도입하지 않는다**는 제약은 불변이다.

**기각한 대안**

- **(a) `slog.Handler` 를 훅 전용으로 교체** — 레벨·포맷을 더 정교하게 다룰 수 있으나,
  `configureLogging` 의 단일 설치 지점 계약을 흔들고 비-훅 경로와 포맷이 갈린다.
  이 SPEC 이 바꾸려는 것은 목적지 하나다.
- **(b) `internal/hook` 안에서 직접 파일에 쓰는 별도 진단 API 신설** — 373곳을 전부
  고쳐 불러야 한다. 호출 지점을 건드리지 않고 목적지만 바꾸는 (B-1)이 압도적으로 작다.
- **(c) 훅 경로에서 stderr 를 다시 연다** — §4 에서 기각 확정. 런타임 판독을 깨뜨린다.

### B-2 프로젝트 루트 해소 — 기존 폴백 사슬을 재사용

`configureLogging` 은 `args` 만 받는다. 싱크 경로를 만들려면 프로젝트 루트가 필요하다.

- 해소 순서: `CLAUDE_PROJECT_DIR`(`internal/config/envkeys.go` `EnvClaudeProjectDir`) →
  프로세스 cwd. 이 사슬은 `internal/cli` 안에 이미 여러 벌 존재하므로 새로 만들지 않고
  그중 하나를 재사용한다(run-phase 가 어느 헬퍼를 쓸지 고른다).
- 어느 쪽으로도 해소되지 않으면 **폐기로 강등**한다(REQ-HDS-003).
- 주의: 훅 프로세스의 cwd 는 워크트리일 수도 primary 체크아웃일 수도 있다. 이 SPEC 은
  "해소된 루트 하위 `.moai/logs/`" 이상을 약속하지 않는다 — 어느 트리에 남는지는
  런타임이 훅에 준 환경이 정한다.

### B-3 싱크 이름 — `.moai/logs/hook-runtime.log` (명명 충돌 주의)

- **"diagnostic" 이라는 낱말은 이 저장소에서 이미 임자가 있다**: `lsphook.Diagnostic`,
  `internal/hook/wire_format_freeze_test.go` 의 `canonicalHookDiagnostics` 는 전부 LSP
  진단을 가리킨다. 싱크를 `hook-diagnostics.*` 로 부르면 그 낱말이 두 뜻을 갖는다.
  따라서 **`hook-runtime.log`** 를 쓴다(측정 시점 `60017eb83` 트리에서 `hook-runtime` /
  `hookRuntimeLog` 식별자 검색 결과 0건 — 충돌 없음).
- 형식은 `slog.NewTextHandler` 의 기본 줄 단위 출력 그대로다. JSONL 로 바꾸지 않는다 —
  비-훅 경로와 같은 포맷이어야 같은 레코드를 같은 눈으로 읽는다.
- 확장자 `.log`: `PruneObservationLogs` 가 `trace-*.jsonl` 을 이름으로 골라내므로,
  `.jsonl` 을 쓰면 기존 후보 판별식과 섞일 위험이 있다.

**세션 분할 여부** — 단일 파일로 간다.

- `trace-<session>.jsonl` 처럼 세션별로 쪼개면 동시 append 경합이 줄지만, 파일 수가
  세션 수만큼 늘고 정리 부담이 그만큼 커진다(`trace-*.jsonl` 수백 개가 바로 그 선례다).
- 동시 append 는 `O_APPEND` + **레코드당 한 번의 쓰기**로 감당한다(REQ-HDS-006).
  POSIX 에서 `O_APPEND` 쓰기는 오프셋 갱신과 원자적이다. 한 레코드를 여러 번에 나눠
  쓰면 그 보장이 깨지므로, 레코드를 조립한 뒤 **한 번에** 내보낸다.
- 잔여 위험: 아주 긴 레코드는 원자성 보장 범위를 넘을 수 있고, Windows 의 `O_APPEND`
  는 에뮬레이션이다. 이 SPEC 은 "다른 프로세스의 레코드를 훼손하지 않는다"까지만
  약속하며, 행 순서는 약속하지 않는다.

### B-4 레벨 경계 — warn 기본, `MOAI_LOG_LEVEL` 로만 낮춤

- 훅 분기의 `level` 을 `resolveLogLevel()` 로 바꾼다. 변수가 없으면 `defaultLogLevel`
  (warn)이므로 기본 동작은 warn 이상 201곳(base `60017eb83`)에 한정된다.
- 이 한 줄이 REQ-HDS-004 와 REQ-HDS-005 를 동시에 만족한다. 목적지는 B-1 이 이미
  고정했으므로 변수는 레벨만 지배한다(REQ-HDS-002 불변).

### B-5 증가 억제 — 기존 정리 후보 집합에 한 행 추가

- `PruneObservationLogs` 의 후보 집합은 `trace-*.jsonl` + `task-metrics.jsonl` +
  `agent-model-audit.jsonl` 로 닫혀 있고, 각각 `PruneStats` 필드를 갖는다.
  `hook-runtime.log` 를 **같은 방식으로** 한 행 추가한다 — 새 기제가 아니다.
- 보존 임계값은 `internal/config/defaults.go` 의 명명 상수. `DefaultTraceRetentionDays`
  를 재사용할지 전용 상수를 둘지는 run-phase 판단이되, **인라인 리터럴은 금지**한다.
- 자기참조 주의: 이 정리 코드 자신의 실패 보고 `slog.Warn` 은 SessionEnd 훅 안에서만
  실행되므로 지금까지 한 줄도 남지 않았다. 이 SPEC 이 닫히면 비로소 남는다.

### B-6 가드 설계 — 효과를 보고, 반증 가능하게

| 가드 | 테스트 이름 | 패키지 | 판정 AC |
|---|---|---|---|
| 1 (행동) | `TestHookPathWarnRecordReachesSink` | `internal/cli` | AC-HDS-001 |
| 2 (계약) | `TestHookPathHookDestIsNeitherStdStream` | `internal/cli` | AC-HDS-004 · 009 |
| 3 (fail-open) | `TestHookPathSinkFailureIsFailOpen` | `internal/cli` | AC-HDS-005 · 006 |
| 4 (정리 편입) | `TestHookSinkIsPrunedAtSessionEnd` | `internal/hook` | AC-HDS-010 |
| 5 (동시성) | `TestHookSinkConcurrentAppendIsIntact` | `internal/cli` | AC-HDS-003 |
| 6 (레벨 경계) | `TestHookSinkLevelBoundary` | `internal/cli` | AC-HDS-007 · 008 |
| 7 (지연 개방) | `TestHookSinkIsNotCreatedWithoutRecords` | `internal/cli` | AC-HDS-012 |

- 일곱 이름 모두 이 트리에 존재하지 않는다 — `go test -list … | grep -cE '^Test'` 존재
  확인이 일곱 개 전부 **0** 으로 갈렸고, 두 패키지 각각에서 양성 대조가 **1** 로 발화했다
  (`acceptance.md` §기준선 (b)(c), HEAD `9b973bc04`). 이 SPEC 이 소유하는 이름이다.
- 가드 5·6·7 은 v0.3.0 신설이다 — 그 전에는 AC-HDS-003 · 007 · 008 · 012 가 가드 1·3 에
  접혀 있어 접힌 단언이 작성됐는지조차 가르지 못했다. 특히 지연 개방(REQ-HDS-007)은
  §B-1 이 그 설계를 고른 유일한 근거이므로 자기 가드를 갖는다.
- 가드 1·3·4 는 **파일 내용**을 단언한다. "라우팅 함수가 호출됐다"가 아니라 "레코드가 파일에
  있다"를 본다.
- 가드 2 는 **결정값의 포인터 동일성**을 단언한다 — 런타임 stdout/stderr 바이트 대조가 이
  저장소에서 판정 불가이기 때문이다(`acceptance.md` 판정 원칙 7). 이름이 초판의
  `TestHookPathLeavesStdoutAndStderrUntouched` 에서 바뀐 이유가 그것이다.
- **가드 2 는 정적 축의 판독 범위도 진다.** 기존 `TestLoggingNeverTargetsStdout` 은
  `os.ReadFile("logging.go")` + `os.Stdout` 토큰 하나라 **새 싱크 파일이 범위 밖**이다.
  가드 2 는 훅 경로 목적지를 구성하는 **파일 집합**을 읽고 파일별로 다른 토큰을 요구한다 —
  `logging.go` 는 `os.Stdout` 만(비-훅 분기가 `os.Stderr` 를 정당하게 쓴다), 싱크 writer
  선언 파일은 **둘 다**. 집합은 싱크 경로 상수(AC-HDS-011)의 **선언 파일**을 반드시
  포함해야 하고, 열거한 파일을 읽을 수 없으면 건너뛰지 말고 실패한다 — 그래야 파일이
  늘어날 때 조용히 빠지지 않는다(AC-HDS-004).
- **[HARD] 판정은 두 단이다** — `-list … | grep -cE '^Test'` = 1 이 먼저이고, 0 이면 AC 는
  미충족이다. `-run` 단독은 가드가 없을 때도 exit 0 을 낸다(같은 절의 실측).
- 반증 절차는 `acceptance.md` AC-HDS-002 가 명시한다.
- 일곱 가드 외에 **기존 자산 2건을 갱신**한다 — `TestLoggingHandlerSelection`(AC-HDS-015)과
  `internal/hook` 주석 3곳(AC-HDS-016). 둘 다 M1.

## §C 마일스톤

되돌리기 어려운 순서이자 의존 순서다.

- **M1 — 목적지 교체**: B-1 의 지연 개방 append writer 도입 + `resolveLoggingDecision`
  훅 분기 교체 + B-2 루트 해소. 가드 1(AC-HDS-001) RED→GREEN.
  **동시에 `internal/cli/logging_test.go` `TestLoggingHandlerSelection` 의 세 서브테스트
  (`hook_discards` / `hook_behind_a_flag_discards` / `hook_ignores_log_level_env`)의 기대값을
  새 목적지로 갱신한다**(AC-HDS-015). 목적지를 바꾸는 순간 그 3건이 빨개지는데, 이것은
  회귀가 아니라 의도된 계약 변경이다 — 빨개진 것을 보고 구현을 되돌리는 것이 이 자리에서
  가능한 가장 나쁜 대응이다. `hook_ignores_log_level_env` 는 기대값만 바꾸고 **"환경변수가
  목적지를 바꾸지 못한다"는 단언은 유지**한다(REQ-HDS-002). 같은 마일스톤에서
  `internal/hook` 주석 3곳도 새 사실로 갱신한다(AC-HDS-016).
- **M2 — 계약·강등 고정**: 가드 2(AC-HDS-004)와 가드 3(AC-HDS-006). stdout/stderr
  불변과 fail-open 강등을 기계적으로 묶는다.
- **M3 — 레벨 경계**: B-4. 기본 warn(AC-HDS-007), `MOAI_LOG_LEVEL` 하향 시 info 기록
  (AC-HDS-008), 목적지 불변(AC-HDS-009).
- **M4 — 증가 억제**: B-5. 정리 후보 편입 + `PruneStats` 필드 + 보존 상수. 가드 4.
- **M5 — 기계적 마감**: 상수 정리(AC-HDS-011), 크로스 빌드(AC-HDS-013),
  `go vet` / `golangci-lint`, 변경 패키지 테스트.

우선순위: M1 High · M2 High · M3 Medium · M4 Medium · M5 Low.

## §D 제약

- stdout / stderr carve-out 보전(REQ-HDS-002). 이 제약을 깨는 설계는 전부 기각이다.
- `hook.Registry` 인터페이스와 훅 디스패처 구조는 건드리지 않는다.
- `internal/template/templates/` 하위 미수정.
- 훅 예산 5초 안에서 경계 없는 대기 금지.
- 하드코딩 금지 — 경로·보존일수·환경변수명은 전부 명명 상수 경유.
- **기존 `TestLoggingNeverTargetsStdout` 은 계속 통과해야 한다** — 그 정적 가드를 약화시키거나
  **판독 범위를 좁히거나** 삭제해서 AC 를 통과시키는 것은 미충족이다. 범위를 넓히는 일은
  그 가드가 아니라 가드 2 가 진다(§B-6).
- **싱크 writer 선언 파일은 `os.Stdout` 도 `os.Stderr` 도 참조하지 않는다.** 훅 경로 전용
  파일이므로 두 스트림 어느 쪽도 정당한 목적지가 아니다. `logging.go` 에는 같은 규칙을
  적용하지 않는다 — 비-훅 분기의 `os.Stderr` 는 정당하다(파일별 비대칭, AC-HDS-004).
- **`TestLoggingHandlerSelection` 의 훅 케이스 3건을 삭제하지 않는다** — 기대값만 갱신한다.
  AC-HDS-015 의 grep 3 유지 판정이 이 축을 기계적으로 막는다.

## §E 위험

- **R1 — 소음 역전.** warn 201곳이 전부 "진짜 이상"은 아니다. 기록이 시작되면 정상
  분기의 잡음이 함께 쌓일 수 있다. 완화: 이 SPEC 은 기록만 연다. 호출 지점 내용 심사는
  범위 밖(§4)이며, 실제 분포가 보인 **뒤에** 후속 카드로 다룬다.
- **R2 — 증가.** 훅은 세션당 수십~수백 회 발화한다. 완화: 지연 개방(REQ-HDS-007) +
  warn 기본(REQ-HDS-004) + 기존 정리 편입(REQ-HDS-009).
- **R3 — 동시 append.** B-3 의 잔여 위험(긴 레코드·Windows 에뮬레이션). 완화: 레코드당
  단일 쓰기, 행 순서 무보장을 명시.
- **R4 — 루트 오해소.** 워크트리에서 발화한 훅의 기록이 어느 트리에 남는지는 런타임이
  준 환경이 정한다. 완화: 약속 범위를 "해소된 루트 하위"로 한정(B-2).
- **R5 — 측정 귀속.** 이 계획의 모든 수치는 `60017eb83` 에 묶여 있다. 병합 트리가
  달라지면 재측정해야 하며, 인용 시 base SHA 를 함께 옮긴다.

## §F 교차참조

- `.moai/reports/t1144/verdict.md` — 측정 출처(base `60017eb83`)
- `spec.md` §1.3 — 코드 표면 앵커 전량
- `acceptance.md` — AC 14건과 각 판정 명령
