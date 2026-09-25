# SPEC-HOOK-DIAG-SINK-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- 카드: t1144 · 워크트리 `.claude/worktrees/t1144` · 브랜치 `WT-hook-diag-channel` · base `60017eb83`
- 1단계(도달성 측정) 완료 — 증거: `.moai/reports/t1144/verdict.md`, `slog-sites.tsv`, `reachability.log`
- 2단계(plan) 산출물: `spec.md` + `plan.md` + `acceptance.md` (Tier M) + 본 파일
- SPEC ID 정규식 자가검사: `SPEC-HOOK-DIAG-SINK-001` → `PASS` (실행 출력)
- 프론트매터 12필드 스키마 대조 완료 · 기존 `.moai/specs/` 내 ID 중복 없음
- 방향 (2)(세션 컨텍스트 노출)는 `spec.md` §4 에 운영자 결정 사항으로 이연 — 요구사항·AC 없음

## §E.2 Run-phase Evidence

### M1 — 목적지 교체 (cycle_type=tdd)

범위: `plan.md` §C M1 뿐이다. M2~M5(계약·강등 고정 / 레벨 경계 / 증가 억제 / 기계적 마감)는
착수하지 않았다. 따라서 훅 분기의 `level` 은 아직 `defaultLogLevel` 고정이며 `resolveLogLevel()`
로 바뀌지 않았다 — 그 교체는 M3(§B-4)의 것이다.

편집 대상:

- `internal/cli/hook_sink.go` (신규) — 지연 개방 append writer + 경로 상수
- `internal/cli/logging.go` — `resolveLoggingDecision` 훅 분기 + 그 함수의 주석
- `internal/cli/hook_sink_test.go` (신규) — 가드 1
- `internal/cli/logging_test.go` — `TestLoggingHandlerSelection` 훅 케이스 3건 기대값 (AC-HDS-015)
- `internal/hook/{config_change,instructions_loaded,factory_messages,pre_tool}.go` — 주석 4곳 (AC-HDS-016)

호출 지점 373곳은 **한 곳도 건드리지 않았다.** `loggingDecision.dest` 가 이미 `io.Writer` 이므로
타입 표면이 바뀌지 않는다는 §B-1 의 근거가 그대로 유지된다.

루트 해소는 새 사슬을 만들지 않고 기존 `resolveHookProjectRoot()`
(`internal/cli/hook.go`)를 재사용했다 — `CLAUDE_PROJECT_DIR` → cwd, 실패 시 `""`. §B-2 가
"run-phase 가 어느 헬퍼를 쓸지 고른다"고 남긴 선택지에서 훅 전용 사슬을 고른 것이다.

#### AC PASS/FAIL 매트릭스 (M1 소관분)

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-HDS-001 | PASS | `go test ./internal/cli/ -list '^TestHookPathWarnRecordReachesSink$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookPathWarnRecordReachesSink$' -count=1` | 1단 `1` · 2단 `ok github.com/modu-ai/moai-adk/internal/cli 0.848s` (EXIT=0) |
| AC-HDS-015 | PASS | `go test ./internal/cli/ -run '^TestLoggingHandlerSelection$' -count=1` + `grep -c 'hook_discards\|hook_behind_a_flag_discards\|hook_ignores_log_level_env' internal/cli/logging_test.go` | 2단 `ok … 0.854s` (7/7 서브테스트 PASS) · grep `3` 유지 |
| AC-HDS-016 | PASS (기계적 축) | 아래 § AC-HDS-016 전수 판정 기록 | 그물 22 / 기록 22 / 지정 4 / 유보 1 |

M1 밖의 AC(002~014)는 **미판정**이다. 여기에 PASS 로 적지 않는다.

#### RED 증거 (§E E8 — GREEN 이전 실패 출력 전문)

RED 는 두 단으로 관측했다. 1단은 심볼 부재(가드가 참조하는 상수가 없음), 2단은 **행동 실패**
(상수·writer 는 있으나 `resolveLoggingDecision` 에 배선하지 않은 상태)다. 2단이 이 가드가
호출이 아니라 효과를 겨눴다는 증거다.

RED 1 — `go test ./internal/cli/ -run '^TestHookPathWarnRecordReachesSink$' -count=1` (EXIT=1):

```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/hook_sink_test.go:39:30: undefined: hookRuntimeLogRelPath
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

RED 2 — 같은 명령, `hook_sink.go` 추가 후 배선 전 (EXIT=1):

```
--- FAIL: TestHookPathWarnRecordReachesSink (0.00s)
    hook_sink_test.go:42: read hook sink /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestHookPathWarnRecordReachesSink3992298639/001/.moai/logs/hook-runtime.log: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestHookPathWarnRecordReachesSink3992298639/001/.moai/logs/hook-runtime.log: no such file or directory
        a warn record emitted on the hook path must reach the sink file (AC-HDS-001)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.916s
FAIL
```

GREEN — 같은 명령, 배선 후 (EXIT=0):

```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.848s
```

### M3 — 레벨 경계 (cycle_type=tdd)

범위: `plan.md` §C M3 뿐이다. M4(증가 억제)·M5(기계적 마감)는 착수하지 않았다.

편집 대상:

- `internal/cli/hook_sink_level_test.go` (신규) — 가드 6 `TestHookSinkLevelBoundary`
- `internal/cli/logging.go` — 훅 분기의 `level` 을 `defaultLogLevel` 고정에서 `resolveLogLevel()`
  로 교체 + `defaultLogLevel` / `resolveLoggingDecision` 주석 2곳

M1 이 §E.2 에 "훅 분기의 `level` 은 아직 `defaultLogLevel` 고정이며 … 그 교체는 M3(§B-4)의
것이다"라고 남긴 그 교체다. **그 고정은 REQ-HDS-005 가 미구현 상태였다는 뜻이고, RED 가
그것을 이름으로 잡았다** — 아래 RED 증거의 첫 출력이 `resolved level WARN, want INFO` 다.

주석 2곳을 함께 고친 이유: `resolveLoggingDecision` 의 "그 carve-out 은 무조건적이다 —
`MOAI_LOG_LEVEL` 이 다시 열지 않는다"는 문장은 **목적지에 대해서는 참으로 남지만**, 이
교체 이후에는 "변수가 이 경로에서 아무 효과도 없다"로 읽힐 수 있다. `defaultLogLevel` 의
"비-훅 경로에서" 라는 한정도 이제 거짓이다. 둘 다 테스트를 깨뜨리지 않으므로 AC-HDS-016 이
막으려는 것과 같은 계열의 조용한 거짓이다.

#### AC PASS/FAIL 매트릭스 (M3 소관분)

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-HDS-007 | PASS | `go test ./internal/cli/ -list '^TestHookSinkLevelBoundary$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookSinkLevelBoundary$' -count=1` | 1단 `1` · 2단 `ok github.com/modu-ai/moai-adk/internal/cli 0.990s` (EXIT=0), 서브테스트 `default_config_admits_warn_only` PASS |
| AC-HDS-008 | PASS | 같은 두 단 명령 | 1단 `1` · 2단 EXIT=0, 서브테스트 `lowered_level_admits_info` + `the_same_record_differs_across_the_two_settings` PASS |

AC-HDS-009 는 **M3 소관이 아니다.** `acceptance.md` § 가드 이름 표가 그것을 가드 2
(`TestHookPathHookDestIsNeitherStdStream`, M2 착지)에 배정했고, 그 가드가 이미 네 레벨 값 ×
두 인자 모양을 순회하며 `dest` 가 두 표준 스트림 어느 쪽도 아님을 단언한다. 가드 6 에서 다시
단언하면 그 표가 바뀔 때 고칠 자리가 둘이 되고 얻는 커버리지는 0 이다 — 두 가드가 변수의 두
축(목적지 / 레벨)을 하나씩 나눠 진다. 이 판단을 가드 6 의 doc comment 에도 적어 두었다.

M3 밖의 AC(002~006 · 010~014)는 여전히 **미판정**이다.

#### RED 증거 (§E E8 — GREEN 이전 실패 출력 전문)

RED — `go test ./internal/cli/ -run '^TestHookSinkLevelBoundary$' -count=1` (EXIT=1):

```
--- FAIL: TestHookSinkLevelBoundary (0.00s)
    --- FAIL: TestHookSinkLevelBoundary/lowered_level_admits_info (0.00s)
        hook_sink_level_test.go:68: resolveLoggingDecision with MOAI_LOG_LEVEL="INFO" resolved level WARN, want INFO; the variable governs the hook sink's minimum level (REQ-HDS-005)
    --- FAIL: TestHookSinkLevelBoundary/the_same_record_differs_across_the_two_settings (0.00s)
        hook_sink_level_test.go:83: resolveLoggingDecision with MOAI_LOG_LEVEL="INFO" resolved level WARN, want INFO; the variable governs the hook sink's minimum level (REQ-HDS-005)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.706s
FAIL
```

GREEN — 같은 명령, 교체 후 (EXIT=0, `-v`):

```
--- PASS: TestHookSinkLevelBoundary (0.00s)
    --- PASS: TestHookSinkLevelBoundary/default_config_admits_warn_only (0.00s)
    --- PASS: TestHookSinkLevelBoundary/lowered_level_admits_info (0.00s)
    --- PASS: TestHookSinkLevelBoundary/the_same_record_differs_across_the_two_settings (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.990s
```

#### 반증 (변이 2건 — 두 축 각각)

RED 는 **관대한 방향**(경계가 너무 낮게 잡히는 실패)을 재지 않는다. 변이를 두 개 돌려 두 축을
각각 빨갛게 만들었다. 둘 다 커밋하지 않았고 원본을 백업에서 복원했다.

변이 A — 훅 분기를 `level: slog.LevelDebug` 로 고정 (결정값 축):

```
--- FAIL: TestHookSinkLevelBoundary/default_config_admits_warn_only
        resolveLoggingDecision with MOAI_LOG_LEVEL="" resolved level DEBUG, want WARN
--- FAIL: TestHookSinkLevelBoundary/lowered_level_admits_info
        resolveLoggingDecision with MOAI_LOG_LEVEL="INFO" resolved level DEBUG, want INFO
--- FAIL: TestHookSinkLevelBoundary/the_same_record_differs_across_the_two_settings
```

변이 A 는 **결정값 단언에서 먼저 멈추므로 싱크 행 수 축을 빨갛게 만들지 못한다.** 그래서
변이 B 를 따로 돌렸다 — `defaultLogLevel = slog.LevelDebug`. 기댓값이 상수와 함께 움직이므로
결정값 단언을 통과하고 파일 내용 단언에 도달한다:

```
--- FAIL: TestHookSinkLevelBoundary/default_config_admits_warn_only
        sink holds 1 line(s) carrying "hook-sink-level-probe-info", want 0: info sits below defaultLogLevel …
        sink holds 1 line(s) carrying "hook-sink-level-probe-debug", want 0: debug sits below defaultLogLevel …
--- FAIL: TestHookSinkLevelBoundary/the_same_record_differs_across_the_two_settings
        the info probe appears 1 time(s) at the default AND 1 time(s) with MOAI_LOG_LEVEL=INFO; …
```

변이 B 의 세 번째 줄이 **공허 방지 팔이 실제로 문다**는 증거다: 두 환경이 같은 답을 내면
빨개진다. AC-HDS-008 은 "낮추면 info 가 싱크에 있다"만 요구하므로, `t.Setenv` 를 아예
호출하지 않아도 — 애초에 기본값이 아닌 싱크를 읽으면 — 충족될 수 있다. 차분 팔이 그 경로를
막는다.

### M4 — 증가 억제 (cycle_type=tdd)

범위: `plan.md` §C M4 뿐이다. M5(기계적 마감)는 착수하지 않았다.

편집 대상:

- `internal/hook/hook_sink_prune_test.go` (신규) — 가드 4 `TestHookSinkIsPrunedAtSessionEnd`
- `internal/hook/prune_logs.go` — `hookRuntimeLogFileName` 상수 · `PruneStats.HookRuntimeLogAged`
  필드 · 후보 분기 1개 · 요약 로그 행 1개 · doc comment
- `internal/config/defaults.go` — `DefaultHookRuntimeLogRetentionDays`

새 정리 기제를 만들지 않았다(REQ-HDS-009). 싱크는 `agentModelAuditFileName` 후보와 **글자
그대로 같은 모양**으로 편입됐다 — 이름 일치 → 임계값 비교 → `PruneStats` 필드 보고.

#### 판단 1건 — 싱크의 임계값은 호출자의 `retentionDays` 가 아니다

REQ-HDS-010 이 보존일수를 `internal/config/defaults.go` 의 **명명 상수**로 요구한다. 그 요구를
그대로 따르면 상수가 실제로 쓰여야 하므로, 싱크는 `DefaultHookRuntimeLogRetentionDays` 로
자기 컷오프를 계산하고 호출자가 넘긴 `retentionDays`(trace · task-metrics ·
agent-model-audit 를 지배)를 쓰지 않는다. 오늘 두 값은 같은 30 이고, 호출자는 하나뿐이며
그 하나가 `config.DefaultTraceRetentionDays` 를 넘긴다.

**대가를 적어 둔다**: 호출자가 `retentionDays` 를 낮춰도 싱크는 따라 내려가지 않는다. 이것은
SPEC 이 명시적으로 정하지 않은 자리이고, 상수를 만들되 쓰지 않는 쪽(죽은 상수)과 상수를 안
만드는 쪽(REQ-HDS-010 위반) 사이에서 고른 값이다. 되돌리려면 `sinkCutoff` 를 `cutoff` 로
바꾸는 한 줄이다.

#### AC PASS/FAIL 매트릭스 (M4 소관분)

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-HDS-010 | PASS | `go test ./internal/hook/... -list '^TestHookSinkIsPrunedAtSessionEnd$' \| grep -cE '^Test'` / `go test ./internal/hook/... -run '^TestHookSinkIsPrunedAtSessionEnd$' -count=1` | 1단 `1` · 2단 `ok github.com/modu-ai/moai-adk/internal/hook 0.388s` (EXIT=0), 서브테스트 `aged_sink_is_removed_and_reported` + `sink_inside_the_threshold_is_kept` 둘 다 PASS |

AC-HDS-011 은 **M4 소관이 아니다**(M5). 다만 M4 가 그것을 깨지 않았음을 확인했다 —
`grep -rn '"\.moai/logs/hook-runtime\.log"' --include='*.go' internal/ | grep -v _test.go`
는 여전히 1행이고 그 행이 `internal/cli/hook_sink.go:19` 의 `const` 선언이다. M4 가
`internal/hook` 에 더한 상수는 **기본 이름**(`"hook-runtime.log"`)이라 그 그물에 걸리지 않는다
— `internal/cli` 가 `internal/hook` 을 import 하지 (그 반대가 아니라) 때문에 전체 경로 상수를
공유할 수 없고, 정리는 `logsDir` 를 엔트리 이름으로 훑으므로 기본 이름만 있으면 된다.

M4 밖의 AC(002~006 · 011~014)는 여전히 **미판정**이다.

#### RED 증거 (§E E8 — GREEN 이전 실패 출력 전문)

RED — `go test ./internal/hook/... -run '^TestHookSinkIsPrunedAtSessionEnd$' -count=1` (EXIT=1):

```
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/hook_sink_prune_test.go:27:33: undefined: hookRuntimeLogFileName
internal/hook/hook_sink_prune_test.go:52:12: stats.HookRuntimeLogAged undefined (type PruneStats has no field or method HookRuntimeLogAged)
internal/hook/hook_sink_prune_test.go:53:54: stats.HookRuntimeLogAged undefined (type PruneStats has no field or method HookRuntimeLogAged)
internal/hook/hook_sink_prune_test.go:69:12: stats.HookRuntimeLogAged undefined (type PruneStats has no field or method HookRuntimeLogAged)
internal/hook/hook_sink_prune_test.go:71:11: stats.HookRuntimeLogAged undefined (type PruneStats has no field or method HookRuntimeLogAged)
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
```

GREEN — 같은 명령, 구현 후 (EXIT=0, `-v`):

```
=== RUN   TestHookSinkIsPrunedAtSessionEnd
=== RUN   TestHookSinkIsPrunedAtSessionEnd/aged_sink_is_removed_and_reported
=== RUN   TestHookSinkIsPrunedAtSessionEnd/sink_inside_the_threshold_is_kept
--- PASS: TestHookSinkIsPrunedAtSessionEnd (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.388s
```

#### 반증 (변이 2건 — 두 팔 각각)

두 팔이 서로 다른 변이에 문다. 한 변이로 둘 다 빨개지지 않는다는 것이 두 팔이 서로 다른
것을 재고 있다는 증거다. 둘 다 커밋하지 않았고 원본을 백업에서 복원했다.

변이 A — 싱크를 후보에서 제외(`if name == hookRuntimeLogFileName && false`):

```
--- FAIL: TestHookSinkIsPrunedAtSessionEnd/aged_sink_is_removed_and_reported (0.00s)
        hook_sink_prune_test.go:53: HookRuntimeLogAged = 0, want 1
        hook_sink_prune_test.go:56: aged sink still present: stat err = <nil>, want IsNotExist
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.547s
```

변이 B — 임계값 무시하고 전부 제거(`if true || info.ModTime().Before(sinkCutoff)`):

```
--- FAIL: TestHookSinkIsPrunedAtSessionEnd/sink_inside_the_threshold_is_kept (0.00s)
        hook_sink_prune_test.go:70: HookRuntimeLogAged = 1, want 0 (file is inside the threshold)
        hook_sink_prune_test.go:74: sink inside the threshold was removed: … no such file or directory
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.555s
```

변이 A 만 돌렸다면 "전부 지우는" 구현이 가드를 통과한다. 변이 B 가 그 경로를 막는다 —
보존 팔은 장식이 아니라 제거 팔을 해석 가능하게 만드는 것이다.

변이 B 의 출력에는 부수적으로 요약 로그 행도 찍혔다(`hook_runtime_log_aged=1`) — 새 필드가
`slog.Info` 요약까지 실제로 도달한다는 관측이다.

#### 영향 패키지 실행 (M4 착지 시점)

범위는 `./internal/cli/... ./internal/hook/... ./internal/config/...` 이고, 판정은 **두 번의
실행**으로 나뉘었다. 나눈 이유를 적는다 — 한 번에 담지 못한 것이지 범위를 줄인 것이 아니다.

1차(합본, `-timeout 20m`) — **EXIT=1**. `internal/cli` 가 20분 벽에 걸려
`panic: test timed out after 20m0s` 를 냈고(패키지 벽시계 1201.206s, 실행 중이던 테스트
`TestRunTemplateSync_EmbeddedTemplatesError`), `internal/hook` 은 `--- FAIL` 1건만 냈다:

```
--- FAIL: TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift (0.30s)
    session_start_drift_fill_test.go:583: Handle cost 52.265833ms more on a cache miss than on a hit, want < 50ms (hit=123.754083ms miss=176.019916ms)
```

둘 다 **부하 의존**이다(측정 시 `load averages: 7.60 8.55 8.66`). 둘 중 어느 것도 이 카드의
변경이 닿는 경로가 아니다 — M4 의 diff 는 `internal/hook/prune_logs.go` ·
`internal/hook/hook_sink_prune_test.go`(신규) · `internal/config/defaults.go` 셋뿐이고
`internal/cli` 에는 한 줄도 없다.

2차(분리 재측정) — 둘 다 **EXIT=0**:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	1171.767s          # -timeout 40m, 단독
ok  	github.com/modu-ai/moai-adk/internal/hook	245.745s          # ./internal/hook/... ./internal/config/... 단독
ok  	github.com/modu-ai/moai-adk/internal/config	8.217s
```

`internal/cli` 의 `ok` 는 이 SPEC 의 나머지 가드(1 · 2 · 3 · 5 · 6 · 7)가 M4 착지 후에도
함께 통과한다는 관측이기도 하다 — 여섯 모두 그 패키지에 있다.

**관측하지 않은 것**: 합본 1회 실행으로 전 범위가 동시에 초록인 상태는 보지 못했다.
`internal/cli` 가 이 머신에서 20분을 넘기기 때문이며, 그 용량 문제 자체는 이 카드가 고치지
않았다.

### M5 — 기계적 마감 (cycle_type=tdd)

범위: `plan.md` §C M5 뿐이다. sync 는 착수하지 않았다.

편집 대상:

- `internal/cli/hook_sink_concurrent_test.go` (신규) — 가드 5 `TestHookSinkConcurrentAppendIsIntact`
- `internal/cli/hook_sink_lazy_test.go` (신규) — 가드 7 `TestHookSinkIsNotCreatedWithoutRecords`

**`internal/` 의 프로덕션 코드는 한 줄도 바뀌지 않았다.** M5 가 더한 것은 가드 둘뿐이며, 둘 다
M1~M4 가 이미 착지시킨 행동을 판정한다. 이 비대칭이 아래 § RED 증거의 형태를 정한다.

#### 두 가드가 왜 지금 생기는가

`acceptance.md` § 가드 이름 표의 주석이 근거다. v0.1.x 에서 AC-HDS-003 과 AC-HDS-012 는
가드 1·3 에 **접혀** 있었고, 접힌 단언은 작성됐는지조차 가르지 못했다. 특히 AC-HDS-012(지연
개방)는 가드 3(개방 **실패** 경로)에 얹혀 있었는데 그 경로는 개방을 **시도한 뒤**의 이야기라
지연 개방을 보지 않고도 통과한다 — 지연 개방은 `plan.md` §B-1 이 이 설계를 고른 **유일한
근거**(REQ-HDS-007)이므로 자기 가드를 갖는다.

#### 가드 5 가 단언하지 않는 것 (의도적)

**행 순서를 판정하지 않는다.** `plan.md` §B-3 이 약속하는 것은 "다른 프로세스의 레코드를
훼손하지 않는다"까지이고, 그 약속의 기제는 `O_APPEND` + **레코드당 한 번의 쓰기**다 — 어느
writer 가 먼저 닿는지는 그 기제가 정하지 않는다. 순서를 요구하는 가드는 설계가 애초에 막겠다고
한 적 없는 이유로 빨개지고, 그 빨강은 싱크의 결함이 아니라 가드의 결함으로 읽혀야 한다.

같은 이유로 잔여 위험 둘도 범위 밖이다(`plan.md` §B-3 · §E R3): 플랫폼의 원자적 쓰기 한계를
넘는 **아주 긴 레코드**와, `O_APPEND` 가 에뮬레이션인 **Windows**. 가드의 filler 는 256바이트로
그 한계에 **닿지 않게** 잡았다 — 한계를 탐침하는 것이 아니라 피하는 값이다.

단언하는 것: 총 행 수 = N×M(8×25=200) · 각 행이 마커를 **정확히 1회** 담음 · 각 행이 온전한
한 레코드의 **전체 일치**(양끝 앵커 정규식) · (writer, seq) 200쌍이 **각각 정확히 1회** 출현.

#### RED 증거 (§E E8 — 이 마일스톤의 형태)

M5 의 두 가드는 **이미 착지한 행동**을 고정하므로, 구현 부재로 인한 RED 가 존재하지 않는다.
그것을 그대로 적는다 — 없는 RED 를 있다고 적는 쪽이 더 나쁘다. 대신 **각 가드가 실제로 빨개질
수 있음**을 변이로 관측했고, 그 출력이 이 자리의 증거다. 두 변이 모두 커밋하지 않았고 백업에서
원본을 복원했다(복원 후 `git status --short` 무출력).

변이 A — `newHookSink` 이 **생성 시점에** 개방하도록 바꿈(지연 개방 제거):

```
=== RUN   TestHookSinkIsNotCreatedWithoutRecords
    hook_sink_lazy_test.go:71: no record emitted: sink …/001/.moai/logs/hook-runtime.log exists (stat err <nil>), want absent — the file must open on the first admitted record, not before (AC-HDS-012)
    hook_sink_lazy_test.go:71: no record emitted: directory …/001/.moai/logs exists (stat err <nil>), want absent — MkdirAll runs inside the lazy open, so the directory must not be created either (AC-HDS-012)
    hook_sink_lazy_test.go:75: record emitted below the level bar: sink …/001/.moai/logs/hook-runtime.log exists (stat err <nil>), want absent — the file must open on the first admitted record, not before (AC-HDS-012)
    hook_sink_lazy_test.go:75: record emitted below the level bar: directory …/001/.moai/logs exists (stat err <nil>), want absent — MkdirAll runs inside the lazy open, so the directory must not be created either (AC-HDS-012)
--- FAIL: TestHookSinkIsNotCreatedWithoutRecords (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.639s
```

변이 A 가 **파일과 디렉터리 두 축 모두** 빨개진 것이 이 가드가 `os.Stat(file)` 한 축보다 넓게
겨눴다는 관측이다 — 개방은 `MkdirAll` 을 먼저 돌리므로, 디렉터리를 만들고 개방에 실패하는
구현은 파일만 보는 검사를 통과하면서 지연이 막으려던 비용을 이미 치른다.

변이 B — `hookSink.Write` 가 한 레코드를 **두 번에 나눠** 씀(레코드당 한 번의 쓰기 제거):

```
--- FAIL: TestHookSinkConcurrentAppendIsIntact (0.01s)
    hook_sink_concurrent_test.go:113: line 0 carries the probe marker 4 time(s), want exactly 1 … (한 행에 네 레코드의 조각)
    hook_sink_concurrent_test.go:113: line 1 carries the probe marker 0 time(s), want exactly 1 … (찢어진 레코드의 꼬리)
    hook_sink_concurrent_test.go:113: line 6 carries the probe marker 2 time(s), want exactly 1 …
    (이하 동형 다수)
```

변이 B 가 **0회 행과 2회 이상 행을 동시에** 냈다 — 찢김은 한 행의 결손과 다른 행의 중첩으로
동시에 나타난다. 이 출력이 가드의 실패 메시지를 고치게 했다: 초안은 어느 쪽이든 "한 행이 여러
레코드의 조각을 담았다"로 보고했는데, 0회 행은 그 서술이 거짓이다(그 행은 **꼬리**다). 두 모양을
갈라 보고하도록 고친 뒤 재측정했다.

#### 양성 대조 — 가드 7 의 3단계

가드 7 은 부재를 두 번 단언한 뒤 **admitted warn 레코드로 파일이 실제로 생기는지**를 마지막에
확인한다. 이 3단계가 없으면 앞의 두 부재는 "지연 개방이 동작한다"와 "가드가 엉뚱한 경로를 보고
있어 무엇도 영영 안 생긴다"를 가르지 못한다 — 부재 팔만 재는 것은 아무것도 출력하지 않는 명령과
구별되지 않는다(`acceptance.md` 판정 원칙 5).

#### AC PASS/FAIL 매트릭스 (M5 소관분)

| AC | 판정 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-HDS-003 | PASS | `go test ./internal/cli/... -list '^TestHookSinkConcurrentAppendIsIntact$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookSinkConcurrentAppendIsIntact$' -race -count=1` | 1단 `1` · 2단 `--- PASS: TestHookSinkConcurrentAppendIsIntact (0.01s)` · `ok github.com/modu-ai/moai-adk/internal/cli 1.967s` (EXIT=0, **`-race`**) |
| AC-HDS-012 | PASS | `go test ./internal/cli/... -list '^TestHookSinkIsNotCreatedWithoutRecords$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookSinkIsNotCreatedWithoutRecords$' -count=1` | 1단 `1` · 2단 `--- PASS: TestHookSinkIsNotCreatedWithoutRecords (0.00s)` · `ok … 0.600s` (EXIT=0) |
| AC-HDS-011 | PASS | `grep -rn '"\.moai/logs/hook-runtime\.log"' --include='*.go' internal/ \| grep -v _test.go` | 정확히 1행: `internal/cli/hook_sink.go:19:const hookRuntimeLogRelPath = ".moai/logs/hook-runtime.log"` — 그 행이 `const` 선언이다 |
| AC-HDS-013 | PASS | `GOOS=linux GOARCH=amd64 go build ./... && GOOS=darwin GOARCH=arm64 go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | 무출력, 마지막 EXIT=0 (세 대상 연쇄이므로 마지막 0 은 셋 모두 0) |
| AC-HDS-002 | PASS | 훅 분기를 `io.Discard` 로 되돌린 뒤 가드 1 의 두 단 / 복원 후 재실행 | 아래 § AC-HDS-002 판정 |

AC-HDS-014 는 아래 § 영향 패키지 실행이 판정한다.

#### AC-HDS-002 판정 (M1~M4 에서 미판정으로 남아 있던 건)

`acceptance.md` AC-HDS-002 는 가드 1 이 **반증 가능**함을 요구한다 — 훅 분기를 `io.Discard` 로
되돌리면 1단은 1 을 유지한 채 2단만 FAIL 해야 한다. M1 의 §E.2 는 **배선 전 RED**(RED 2)를
인용했는데, 그것은 "아직 배선되지 않은 상태"이지 "배선된 구현을 되돌린 상태"가 아니다. 두 상태가
같은 출력을 낼 것이라는 기대는 합리적이지만 **관측은 아니다.** M5 에서 실제로 측정했다.

변이(`dest: io.Discard` 복원) — 1단 `1` **유지**, 2단 EXIT=1:

```
--- FAIL: TestHookPathWarnRecordReachesSink (0.00s)
    hook_sink_test.go:42: read hook sink …/001/.moai/logs/hook-runtime.log: open …: no such file or directory
        a warn record emitted on the hook path must reach the sink file (AC-HDS-001)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.612s
```

복원 후 같은 명령 — EXIT=0:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.588s
```

복원은 `git checkout --` 이 아니라 백업 파일 복사로 했고(워크트리의 git 상태를 건드리지 않기
위해서다), 복원 뒤 `git status --short internal/cli/logging.go` 는 무출력이다 — 커밋본과
바이트 동일하다.

#### M2 소관 AC 4건의 판정 자리 메움 (§E.2 에 M2 절이 없다)

**먼저 결함을 적는다**: 커밋 `3b889723f`("M2 pin the hook-destination contract and fail-open
degradation")는 가드 2·3 을 착지시켰으나 **§E.2 에 M2 절이 없다.** M1 · M3 · M4 는 각각 절을
갖는데 M2 만 비어 있고, 따라서 AC-HDS-004 · 005 · 006 · 009 는 인용된 판정 출력이 없다.

M5 가 **M2 의 기록을 대신 쓰지는 않는다.** M2 의 RED 를 관측한 것은 M2 이지 M5 가 아니고,
관측하지 않은 것을 기록으로 만들면 그것이 바로 이 SPEC 이 막으려는 관측 없는 주장이다. 대신
M5 가 **지금 이 트리에서 직접 잰** 두 단 판정만 아래에 적고, 출처를 M5 로 귀속한다.

| AC | 판정 | 검증 명령 (M5 가 실행) | 실제 출력 |
|---|---|---|---|
| AC-HDS-004 · 009 | PASS | `go test ./internal/cli/... -list '^TestHookPathHookDestIsNeitherStdStream$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookPathHookDestIsNeitherStdStream$' -count=1 -v` | 1단 `1` · 2단 `ok … 0.604s` (EXIT=0). 서브테스트에 네 레벨 × 두 인자 모양 순회가 보이고(`behind_aflag_level_{INFO,WARN,ERROR}` 등), `source_tokens_absent_per_file` PASS |
| AC-HDS-005 · 006 | PASS | `go test ./internal/cli/... -list '^TestHookPathSinkFailureIsFailOpen$' \| grep -cE '^Test'` / `go test ./internal/cli/ -run '^TestHookPathSinkFailureIsFailOpen$' -count=1 -v` | 1단 `1` · 2단 `ok … 0.598s` (EXIT=0). 서브테스트 `unresolvable_root_degrades_to_discard` · `empty_root_degrades_to_discard` · `sink_path_preempted_by_directory` 셋 다 PASS |

**관측하지 않은 것**: 이 네 AC 의 **RED**(가드 2·3 이 GREEN 이전에 실패하는 출력). 그것은 M2
시점에만 관측 가능했고 기록되지 않았다 — 지금 이 트리에서는 구현이 이미 있으므로 재현할 수 없다.
두 가드 자체가 변이로 빨개지는지도 M5 는 재지 않았다(가드 5·7 에 대해서만 쟀다). 처분은 감사 몫이다.

#### 일곱 가드 전수 존재 확인 (M5 착지 시점)

`acceptance.md` 의 **[HARD] 일곱 가드 전부 존재 확인을 갖는다** 를 한 번에 재측정했다.

`internal/cli` 여섯 — `go test ./internal/cli/... -list '<여섯 이름의 교대>' | grep -E '^Test' | sort`:

```
TestHookPathHookDestIsNeitherStdStream
TestHookPathSinkFailureIsFailOpen
TestHookPathWarnRecordReachesSink
TestHookSinkConcurrentAppendIsIntact
TestHookSinkIsNotCreatedWithoutRecords
TestHookSinkLevelBoundary
```

`internal/hook` 하나 — `go test ./internal/hook/... -list '^TestHookSinkIsPrunedAtSessionEnd$' | grep -cE '^Test'` → `1`.

일곱 전부 존재한다. M4 착지 시점에는 다섯이었다(가드 5·7 이 M5 에서 생겼다).

> **M4 §E.2 의 한 문장을 정정한다 (그 기록은 고치지 않는다).** M4 의 § 영향 패키지 실행은
> `internal/cli` 의 `ok` 를 두고 "이 SPEC 의 나머지 가드(1 · 2 · 3 · **5** · 6 · **7**)가 M4
> 착지 후에도 함께 통과한다는 관측"이라고 적었다. 가드 5 와 7 은 그 시점에 **존재하지
> 않았으므로** 그 둘에 대해서는 관측이 아니라 미관측이다(관측 없는 주장). 관측 기록은 나중
> 사실에 맞춰 고쳐 쓰지 않으므로 M4 의 문장은 그대로 두고, 정정을 여기에 앞으로 적는다 — 그
> 문장이 덮는 범위는 가드 1 · 2 · 3 · 6 넷이다.

#### 도구 체인

| 검사 | 명령 | 출력 |
|---|---|---|
| 포맷 | `gofmt -l internal/cli/ internal/hook/` | 무출력 (해당 파일 0건) |
| vet | `go vet ./internal/cli/... ./internal/hook/...` | 무출력, EXIT=0 |
| lint | `golangci-lint run ./internal/cli/... ./internal/hook/...` | `0 issues.` |

#### 영향 패키지 실행 (M5 착지 시점) — AC-HDS-014

M4 는 이 판정을 **두 번으로 나눠** 냈다(`internal/cli` 가 20분 벽에 걸렸고, `internal/hook` 이
부하 의존 타이밍 테스트 1건으로 빨개졌다). M5 는 **한 번의 합본 실행**으로 냈다 — 나눈 적 없다.

명령(실제로 돌린 그대로, 환경 스크럽과 같은 compound 호출):

```
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
      MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER \
      MOAI_FACTORY_WORKERS && \
go test ./internal/cli/... ./internal/hook/... ./internal/config/... -count=1 -timeout 40m
```

**EXIT=0**, 32 패키지 전부 `ok`, `FAIL` 0건, 타임아웃 0건. 출력 전문:
`.moai/state/verify/t1144-m5/affected-packages.log`. 주요 행:

```
ok  	github.com/modu-ai/moai-adk/internal/cli	1253.516s
ok  	github.com/modu-ai/moai-adk/internal/hook	288.200s
ok  	github.com/modu-ai/moai-adk/internal/config	10.751s
```

| AC | 판정 | 근거 |
|---|---|---|
| AC-HDS-014 | PASS | 위 실행이 `-timeout 40m` 안에서 EXIT=0 — `panic: test timed out` 없음 |

부하 조건도 적어 둔다(판정의 해석 범위를 좁히기 위해서다): 실행 시작 시
`load averages: 25.10 17.85 13.94`, 종료 시 `11.75 11.10 11.88`. M4 를 빨갛게 만든
`TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift` 는 이번 실행에서 발화하지 않았다 — 그것이
그 테스트가 고쳐졌다는 뜻은 아니다. 부하 의존 플레이크는 부하가 낮을 때 통과하며, 이 실행은
그 조건에 해당했다.

무거운 실행이므로 `moai slot acquire --resource internal-cli-suite --max-duration 45m` 으로
자원 임대를 먼저 잡았고(`CLAUDE.local.md` §8), 종료 후 `moai slot release` 했다.

**M4 가 관측하지 못했다고 적은 것을 M5 가 관측했다**: "합본 1회 실행으로 전 범위가 동시에
초록인 상태". 이번 실행이 그 상태다.

### AC-HDS-016 전수 판정 기록

모집단은 기준선 SHA `bbc855f45` 에 고정된 그물이다:

```
git grep -niE 'discard' bbc855f45 -- internal/hook | grep '\.go:' | grep -v _test.go
```

실측 **22행 / 13파일** — `acceptance.md` 이 적은 값과 일치한다(직접 재측정했다).
행 번호는 그물이 낸 기준선 값 그대로이며, M1 의 편집으로 현재 트리에서 이동한 행이 있어도
바꾸지 않는다(그물이 움직이면 행 수 일치가 공허해진다).

그물은 **의도적으로 과다 포착**하므로 아래 22행 중 대다수는 이 SPEC 과 무관하다. 그물은 주석만
잡는 것이 아니라 **문자열 리터럴도 잡는다**(`handoff_inject.go:92`).

| 그물 행 | 분류 | 근거 |
|---|---|---|
| internal/hook/branch_guard.go:172 | 무관 | `quotedArgumentPlaceholder` 가 따옴표 구간의 **내용**을 버린다는 서술. 로깅 목적지와 무관 |
| internal/hook/branch_guard.go:193 | 무관 | 셸 래퍼 안의 git 호출이 **치환된 구간**에 들어가 매치를 놓친다는 잔여 위험 서술 |
| internal/hook/branch_guard.go:282 | 무관 | `)` `}` 뒤의 `#` 을 셸이 주석으로 **버리는** 텍스트까지 가드가 훑는다는 과다 매치 설명 |
| internal/hook/branch_guard.go:333 | 무관 | 바로 위와 같은 축(셸이 버리는 구간을 가드가 훑음). 로깅과 무관 |
| internal/hook/config_change.go:192 | 갱신 대상 | `io.Discard` 로 "EVERY record" 라우팅을 **사실로** 서술. 갱신함 — 두 표준 스트림 미도달은 참으로 남기고, 폐기 서술을 파일 싱크 사실로 바꾸되 **별도 감사 행이 필요한 이유**(레벨 게이트·보존 정리·혼합 스트림)를 명시. `MOAI_LOG_LEVEL` 절은 REQ-HDS-002 로 참이므로 보존 |
| internal/hook/factory_messages.go:144 | 갱신 대상 | "a hook process discards slog records too … still not reported anywhere. **Closing that is card t1144**" — 이 SPEC 이 그 카드다. 갱신함: 호출부가 상태 문자열을 버린다는 참인 절은 보존하고, "어디에도 보고되지 않는다"를 "warn 레벨로 보고 가능하나 이 경로는 아직 방출하지 않는다"로 교체 |
| internal/hook/handoff_inject.go:92 | 무관 | **주석이 아니라 사용자 표시 문자열**(`"… to discard it.\n"`). 대기 중인 handoff 레코드를 버리라는 안내문 |
| internal/hook/handoff_inject.go:213 | 무관 | `saved_at` 부재 시 레코드를 **버리지 않는다**는 보수적 stale 판정 서술 |
| internal/hook/instructions_loaded.go:48 | 갱신 대상 | "routes every `moai hook` invocation to io.Discard, unconditionally" — 갱신함. 이 레코드는 **Info** 라 M1 이후에도 싱크의 warn 게이트 아래에 있어 여전히 어디에도 안 남는다는 사실을 새로 서술(감사 행이 필요한 이유가 그대로 유지되는 근거) |
| internal/hook/pre_tool.go:442 | 갱신 대상 | "the `moai hook` path installs a discarding handler, so a log record here would be silent by construction" — 폐기를 **설계 근거**로 삼은 주석. 갱신함: `gateNotice` 가 구조화 출력을 타는 이유를 「폐기되므로」가 아니라 「싱크는 stdout/stderr 에 닿지 않아 응답과 함께 도착하지 못하고, 통과한 게이트는 warn 미만」으로 다시 세움 |
| internal/hook/pre_tool.go:623 | 무관 | `checkForeignSessionAdvisory` 의 **반환값**을 의도적으로 버린다는 서술(자문 전용, Deny 없음) |
| internal/hook/quality/gate.go:903 | 무관 | `executeStep` 의 `$PATH` LookPath 스킵이 **러너**를 버린다는 서술. 도달성 측정에서 `internal/hook/quality` 는 OTHER-ONLY(훅 경로 비도달, `spec.md` §1.1) |
| internal/hook/registry.go:114 | 무관 | 종전 `ctx.Err()`-우선 순서가 핸들러의 **유효 출력**을 버렸다는 과거 결함 서술 |
| internal/hook/registry.go:321 | 무관 | 권한 사다리 없이는 핸들러의 `"ask"` 가 버려지고 pre-seed `"allow"` 가 나간다는 서술 |
| internal/hook/session_start.go:333 | 무관 | 종전 cold 세션이 bounded join 을 다 치르고 **계산 결과**를 버렸다는 latency 서술 |
| internal/hook/session_start.go:420 | 판정 유보 | `acceptance.md` 의 지정 유보 행. **과거형**이고 폐기 주체를 `resolveLoggingDecision` 이 아니라 **셸 래퍼의 stderr 처리**로 돌린다. 이 SPEC 은 stderr 를 열지 않으므로(REQ-HDS-002) M1 이후 이 문장이 거짓이 되는지 단정할 수 없다 — 갱신하지 않았고, 처분은 감사 몫이다 |
| internal/hook/session_start.go:454 | 무관 | `writeKanbanSessionRecord` 가 **모든 실패**를 버려 세션 시작이 그것에 gate 하지 않는다는 fail-open 서술 |
| internal/hook/session_start_drift_fill.go:12 | 무관 | 캐시 writer 부재로 cold 세션이 join 결과를 버렸다는 과거 결함 서술 |
| internal/hook/session_start_kanban.go:202 | 무관 | dropped 카드를 **운영자가** 버렸다는 큐 상태 정의 |
| internal/hook/session_start_record.go:20 | 무관 | `WriteBestEffort` 가 모든 실패를 버린다는 fail-open 서술 |
| internal/hook/session_start_record.go:51 | 무관 | 재진입 시 재기록하면 오케스트레이터가 쓴 세 필드를 버리게 된다는 서술 |
| internal/hook/user_prompt_submit.go:76 | 무관 | 2 rune 미만 파생 제목을 버린다는 `titleMinRunes` 상수 주석 |

지정 4행은 전부 `갱신 대상` 으로 분류되고 실제로 갱신됐으며, `session_start.go:420` 은
`판정 유보` 로 기록했다. **나머지 18행의 분류 타당성은 감사 몫이다**(`acceptance.md` §Gaps).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-25T08:40+09:00
run_commit_sha: bedc731d6        # M5 커밋 SHA. 이 블록이 그 커밋에 함께 들어가므로 자기 참조가
                                 # 불가능했다 — sync 가 backfill 했다(§E.4 backfill 항 참조)
run_status: PASS
ac_pass_count: 16
ac_fail_count: 0
ac_total: 16
ac_attribution:                  # 각 AC 를 어느 마일스톤이 관측했는가 (§E.2 의 인용 출처)
  M1: [AC-HDS-001, AC-HDS-015, AC-HDS-016]
  M2: []                         # M2 는 §E.2 절이 없다 — 아래 caveat 참조
  M3: [AC-HDS-007, AC-HDS-008]
  M4: [AC-HDS-010]
  M5: [AC-HDS-002, AC-HDS-003, AC-HDS-004, AC-HDS-005, AC-HDS-006,
       AC-HDS-009, AC-HDS-011, AC-HDS-012, AC-HDS-013, AC-HDS-014]
ac_caveats:
  - "AC-HDS-004 · 005 · 006 · 009 는 M2 가 구현을 착지시켰으나 §E.2 에 M2 절이 없다. M5 가
     현재 트리에서 두 단 판정을 직접 재고 M5 귀속으로 기록했다. 네 건의 RED 는 관측되지
     않았고 재현 불가다 (§E.2 § M2 소관 AC 4건)."
  - "AC-HDS-016 은 지정 4행 갱신 + 유보 1행. 나머지 18행의 분류 타당성은 감사 몫."
guard_inventory:                 # acceptance.md [HARD] 일곱 가드 전수 존재 확인, M5 재측정
  present: 7
  expected: 7
preserve_list_post_run_count: 0  # 미커밋 잔여물 없음. M5 의 변이 3건(가드 5 · 7 · AC-002)은
                                 # 모두 백업 복원했고 복원 후 git status 무출력
l44_pre_commit_fetch: "126 7"    # git rev-list --count --left-right origin/develop...HEAD
                                 # 좌 126 = 카드 분기 이후 develop 이 나아간 양(통합 창이 흡수).
                                 # 우 7 = 이 카드의 커밋. 발산이 아니라 카드 브랜치의 정상 형태
l44_post_push_fetch: NOT_MEASURED  # 레인은 push 하지 않는다 (CLAUDE.local.md §4.1 — develop
                                   # push 는 리드 일괄). 원격 착지 검증은 리드 몫
new_warnings_or_lints_introduced: 0
toolchain:
  gofmt: "0 files (internal/cli/ internal/hook/)"
  go_vet: "EXIT=0, no output"
  golangci_lint: "0 issues."
cross_platform_build:
  linux_amd64: PASS
  darwin_arm64: PASS
  windows_amd64: PASS
  command: "GOOS=linux GOARCH=amd64 go build ./... && GOOS=darwin GOARCH=arm64 go build ./... && GOOS=windows GOARCH=amd64 go build ./..."
  observed: "no output, final EXIT=0"
affected_package_run:
  scope: "./internal/cli/... ./internal/hook/... ./internal/config/..."
  form: "단일 합본 실행 (M4 의 2분할과 달리 나누지 않았다)"
  exit: 0
  packages_ok: 32
  packages_failed: 0
  timeouts: 0
  evidence: .moai/state/verify/t1144-m5/affected-packages.log
  load_at_start: "25.10 17.85 13.94"
  load_at_end: "11.75 11.10 11.88"
total_run_phase_files: 19        # 카드 기여 전체 (merge-base 60017eb83..HEAD 17 + M5 신규 2)
  run_phase_source_files: 15     # internal/ 아래
  spec_artifact_files: 4         # .moai/specs/SPEC-HOOK-DIAG-SINK-001/
m1_to_mN_commit_strategy: "마일스톤당 커밋 1건 이상, amend·force-push 없음. M1 f69be5b77 ·
  M2 3b889723f · M3 5f539e53d · M4 900d45f3f(+증거 23278982f) · M5 (이 커밋).
  M5 는 internal/ 프로덕션 코드를 한 줄도 바꾸지 않았다 — 가드 2개 신규 + progress.md 뿐이다."
residual_risk:
  - "동시 append 의 두 잔여 위험(원자 쓰기 한계를 넘는 긴 레코드 · Windows 의 O_APPEND
     에뮬레이션)은 plan.md §B-3 · §E R3 이 범위 밖으로 선언했고 가드 5 도 겨누지 않는다.
     가드의 filler 256바이트는 한계를 피한 값이지 탐침한 값이 아니다."
  - "AC-HDS-014 의 PASS 는 load 11~25 구간의 이 머신 1회 실행이다. 부하 의존 플레이크
     TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift 는 이번에 발화하지 않았으나 고쳐진
     것이 아니다 — 높은 부하에서 재발할 수 있다."
  - "darwin/windows 의 실행 판정은 없다. 크로스 빌드는 컴파일까지이고, 실제 실행 판정은
     원격 CI 매트릭스 몫이다(develop push 후, 리드 판독)."
  - "가드 2·3 의 변이 반증은 M5 가 재지 않았다. 가드 5·7 에 대해서만 쟀다."
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-25T10:05+09:00
sync_commit_sha: b51cc0d57               # sync 커밋. 그 커밋 안에서는 자기 해시를 인용할 수 없어
                                         # placeholder(`pending-backfill-sync`)로 두고 직후 커밋이
                                         # 채웠다(spec-frontmatter-schema.md
                                         # § SHA placeholder backfill exemption)
sync_status: PASS
b12_self_test_a:                         # 중복 방출 방지 — 방출 전 grep
  command: "git show HEAD:CHANGELOG.md | grep -c 'SPEC-HOOK-DIAG-SINK-001'"
  observed: "0"
  verdict: PASS                          # 0 이므로 방출 진행. 방출 후 working tree 재측정 = 1
  post_emission: "grep -c 'SPEC-HOOK-DIAG-SINK-001' CHANGELOG.md → 1"
b12_self_test_b:                         # AC 수 일치
  command: "grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' .moai/specs/SPEC-HOOK-DIAG-SINK-001/acceptance.md | sort -u | wc -l"
  observed: "17"
  disambiguation: "17 은 고유 토큰 수이고 AC 는 16 건이다. 17번째 토큰은 산문 안의 약칭
    `AC-016`(3곳: acceptance.md:580 · 640 · 642)이며 AC-HDS-016 을 가리키는 별칭이지 17번째
    기준이 아니다. `grep -oE 'AC-HDS-[0-9]+' … | sort -u | wc -l` → 16."
  ac_total: 16
  changelog_states: 16
  verdict: PASS                          # 0 이 아니므로 공허 비교가 아니다
b12_self_test_c:                         # CHANGELOG 가 주장하는 경로 전수 존재 확인
  command: "ls internal/cli/logging.go internal/cli/hook_sink.go internal/hook/prune_logs.go internal/config/defaults.go"
  observed: "4 경로 전부 존재 (exit 0)"
  verdict: PASS
changelog_entry_position: "[Unreleased] → ### Fixed 최상단 (CHANGELOG.md:73). 이 저장소의
  관례대로 최신 항목이 절 머리에 온다. Added 에는 중복 방출하지 않았다 — 결함(훅 진단 소실)
  해소가 이 SPEC 의 성격이고, 새 파일 서술은 같은 항목의 하위 항목이 담는다."
frontmatter_status_transitions:
  spec_md: "in-progress → implemented → completed (이 sync 커밋 1건에 병합). updated: 2026-09-24 → 2026-09-25"
  plan_md: "해당 없음 — 이 파일에 frontmatter 블록이 없다(관측: 1행이 `# SPEC-… 구현 계획`)"
  acceptance_md: "해당 없음 — 동일"
  progress_md: "해당 없음 — 상태는 본문 §E 절이 나른다(스키마 doctrine 대로)"
  note: "「4 산출물 원자 전이」의 실제 적용 범위는 spec.md 하나다. 나머지 3개는
    spec-frontmatter-schema.md § Artifact Statelessness 가 status 축에서 무상태로 규정한
    파일이고, 이 트리에서는 frontmatter 블록 자체가 없다. 없는 필드를 전이했다고 적지 않는다."
canary_compliance_check:
  applicable: false
  reason: "이 SPEC 은 자기 sync 가 시험할 전방향 정책을 정의하지 않는다 — 산출물은 런타임
    동작 1건(목적지 교체)과 가드 7건이고, 다른 카드에 의무를 부과하는 조항이 없다."
run_sha_backfill:
  field: "§E.3 run_commit_sha"
  was: "PLACEHOLDER-M5"
  now: "bedc731d6"
  authority: "§E.3 블록 자신이 `sync 가 backfill 한다`고 지정했다. D3 SHA placeholder backfill
    exemption 과 같은 계열의 기계적 채움이며, §E.3 의 다른 필드는 건드리지 않았다."
docs_touched:
  - "CHANGELOG.md — [Unreleased] ### Fixed 신규 항목 1건(+하위 3건)"
  - ".moai/project/codemaps/entry-points.md — 문장 1건 정정 + 헤더 「문장 정정」 스탬프 1행"
  - ".moai/specs/SPEC-HOOK-DIAG-SINK-001/spec.md — frontmatter status/updated"
  - ".moai/specs/SPEC-HOOK-DIAG-SINK-001/progress.md — §E.3 SHA backfill + 본 §E.4"
docs_not_touched:
  - "internal/template/templates/** — 변경 불필요. 실측: `grep -rn 'hook-runtime'
     internal/template/templates/` 무출력(exit 1), 그리고 `.moai/logs/.gitkeep` 이 이미
     디렉터리를 스캐폴드한다. 싱크 경로는 Go 상수이므로 템플릿에 나타날 자리가 없다."
  - "docs-site/** (ko/en/ja/zh) — 손대지 않았다. 이 사이트에는 `.moai/logs/` 인벤토리를
     나열하는 참조 페이지가 없고(실측: `.moai/logs` 언급은 agent-model-audit.jsonl ·
     navigator-sync.log · ci-autofix/ · task-metrics.jsonl 처럼 각 기능 페이지 안에서만
     나타난다), hook-runtime.log 는 자기 기능 페이지를 갖지 않는 내부 진단 파일이다.
     4-locale 동시 갱신 의무를 지는 변경을 귀속할 페이지가 없어 CHANGELOG 를 사용자 통지
     경로로 삼았다. 이 판단은 감사가 뒤집을 수 있는 자리다."
  - ".moai/docs/hook-development.md — 실측상 로깅 목적지를 서술하지 않는다(`grep -n logs`
     무출력). 정정할 문장이 없다."
  - "internal/hook/ 주석 4곳 — M1 이 이미 정정했다. 그중 instructions_loaded.go ·
     config_change.go 는 리드 지시로 이 sync 가 건드리지 않는다(develop 충돌 예정)."
merge_window_carry_forward:                # 이 sync 의 수리 대상이 아니다 — 병합 창이 읽을 기록
  - id: MW-1
    what: "origin/develop 이 M1 이 편집한 internal/hook 파일 4개 중 2개를 이미 바꿨다.
      텍스트 충돌이 예상된다."
    measured: |
      git fetch origin develop && git log --oneline HEAD..origin/develop -- <경로>
      (HEAD = 9d6bc91fe 직전 상태, 이 sync 트리)
        internal/hook/instructions_loaded.go → 3fd0cc5ee (card t1160)
        internal/hook/config_change.go       → 50de6b866 (card t1165)
      양성 대조: 같은 형태를 internal/hook 전체에 걸어 4행 — 0 이 아니므로 위 두 줄의
      1행씩은 부재-미측정이 아니라 실제 접촉이다.
      리드 배차문은 config_change.go 의 소유 카드를 명명하지 않았다. t1165 는 이 sync 의
      독립 측정이다.
    beyond_text: "t1160 이 더한 주석은 이 SPEC 이 제거하는 `io.Discard` 전제에서 추론한다.
      충돌을 기계적으로 봉합하면 틀린 서술이 develop 에 남는다 — 병합자가 문장 자체를
      읽어야 한다."
    owner: "병합 창(리드). 이 sync 는 두 파일을 열지 않았다."
  - id: MW-2
    what: "AC-HDS-016 의 판정 그물이 `bbc855f45` 에 핀돼 있어, 22행 기록은 그 이후 develop 이
      더한 주석 행을 담을 수 없다."
    required: "흡수 후 그물을 다시 떠서 누락 행을 대조해야 한다. 재측정 없이 22행을 그대로
      인용하는 것은 사라진 트리에 대한 수치 인용이다."
    owner: "병합 창(리드)."
known_items:                               # 기록만 한다 — 이 sync 가 고치지 않는다
  - "`hook_discards` / `hook_behind_a_flag_discards` 두 서브테스트 이름이 낡았다 — 지금 그
     둘은 싱크를 단언한다. 운영자가 세 차례 제안을 받고 세 차례 평범한 마일스톤을 택했으므로
     개명하지 않았다."
  - "AC-HDS-004 · 005 · 006 · 009 의 RED 는 관측되지 않았고 재현 불가다(§E.3 ac_caveats).
     sync 가 그 상태를 바꾸지 않았다."
gaps:                                      # 이 sync 가 관측하지 않은 것
  - "Go 테스트를 실행하지 않았다. 이 sync 는 Go 파일을 한 줄도 바꾸지 않았다
     (변경: CHANGELOG.md · entry-points.md · spec.md frontmatter · progress.md).
     따라서 run-phase 의 affected-package 판정을 그대로 물려받고 다시 재지 않았다."
  - "acceptance.md 를 바꾸지 않았으므로 `./internal/spec/...` 재측정 범위도 열지 않았다."
  - "push 하지 않았고 PR 도 열지 않았다. origin/develop 은 126 커밋 앞서 있으며 흡수는
     병합 창의 일이다. 원격 CI 판정은 이 기록에 없다."
  - "docs-site 를 손대지 않기로 한 판단(위 docs_not_touched)은 이 sync 의 판단이고,
     사용자 통지가 CHANGELOG 로 충분한지는 감사·운영자가 뒤집을 수 있다."
```
