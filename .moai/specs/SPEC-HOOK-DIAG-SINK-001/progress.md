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

_(pending run-phase)_

## §E.4 Sync-phase Audit-Ready Signal

_(pending sync-phase)_
