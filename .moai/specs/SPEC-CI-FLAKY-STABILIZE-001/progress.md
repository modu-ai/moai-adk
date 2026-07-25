---
id: SPEC-CI-FLAKY-STABILIZE-001
title: "CI Flaky 안정화 — 진행 기록"
version: "0.3.0"
status: in-progress
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.0.1"
module: "internal/cli, internal/cli/preference, internal/session"
lifecycle: spec-anchored
tags: "ci, flaky-test, data-race, cobra, flock, starvation, jitter, backoff, tdd"
tier: M
---

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** — 산출물 4종(spec.md / plan.md / acceptance.md / progress.md). `research.md`·`design.md`는 조사 완료분을 `plan.md` §C에 통합하여 생략.
- SPEC ID 정규식 사전 검증: `SPEC-CI-FLAKY-STABILIZE-001` → **PASS** (Bash 실행 확인).
- ID 유일성: 기존 `SPEC-V3R6-CI-FLAKY-STABILIZE-001/002`(모두 `completed`, 서로 다른 결함 대상)와 ID 충돌 없음.
- 요구사항: GEARS 표기, `REQ-CFS-001`~`REQ-CFS-023` (17건: 001-006 / 010-016 / 020-023).
- 인수 기준: `AC-CFS-001`~`AC-CFS-026` + `AC-CFS-004b` (27건, MUST 25 / SHOULD 2). `AC-CFS-026`은 0.2.0 신규(REQ-CFS-006 인벤토리 가드), `AC-CFS-004b`는 0.3.0 신규(신호 대입 위치 확인)이며 둘 다 기존 번호를 밀지 않고 배번했다.
- 개발 모드: `tdd` (`.moai/config/sections/quality.yaml`).
- clarification 마커: **0건** — 두 결함의 수정 방향·범위·기법이 모두 사용자 결정으로 확정됨.
- Gaps 명시: `spec.md` §G (8항목) / `acceptance.md` §F (8항목) — 두 결함 모두 로컬 미재현이며 은폐하지 않음.
- 잔여 위험 명시: `spec.md` §H (6항목).

### 0.2.0 개정 (plan-auditor FAIL 0.69 대응)

결함 1 관련 5개 MUST-FIX + 2개 SHOULD-FIX + 3개 NICE-TO-HAVE 반영. 결함 2는 진단·파라미터·오류 계약 모두 감사 통과로 무변경.

| 항목 | 변경 |
|------|------|
| 가드 관측 기전 | 정렬 상태(구조적 반증 불가) → 도달 가능성 신호 `warmUpDone`. cobra `Commands()`가 접근자 내부에서 제자리 정렬하므로 정렬 상태는 항상 참 |
| warm-up 범위 | `rootCmd`/`hookCmd` 2개 열거 → `rootCmd`(및 `PreferenceCmd`) 기점 깊이 우선 재귀 의무화. `githubCmd` 누락 해소 |
| M1 재현 테스트 대상 | 테스트 로컬 트리 → 실제 패키지 전역. warm-up 도달 불가로 인한 M1↔M2 모순 해소 |
| 반복 검증 명령 | `-race -count=50` → 신규 프로세스 50회 루프(`-count=1`). 단일 프로세스 `-count=N`은 1회성 전역 플래그 때문에 판별력이 `-count=1`과 동일 |
| 결함 클래스 | 구문 후보 11건 전체 → 전역 수신자 7건. 4건(`newHarnessRouterCmd()`×3, `newToolPolicyCmd()`×1)은 테스트 로컬 수신자. `t.Parallel()` 보존 범위는 11건 전체로 유지 |
| REQ-CFS-006 | 검증 불가(기존 가드 통과만 확인) → `AC-CFS-026` 신규 부여(go/ast 인벤토리 가드). **신규 작업 범위 추가** — 비용은 `plan.md` §B.4 |
| AC-CFS-009 증거 | 실패 출력만 → 실패 출력 + 테스트 전체 소스를 §E.2에 보존(AC-CFS-010 파일 제거로 증거 소실 방지) |
| Gaps | `-count` 의미론 공백 추가(`spec.md` §G 2항, `acceptance.md` §F 6항), 증거 로그 비자기기술성 추가 |
| 라벨 | REQ-CFS-006 `(Event-detected)` → `(Event-driven)` (GEARS 정본 표기) |

### 0.3.0 개정 (plan-auditor PASS 0.844 잔여 지적)

MUST-FIX 1건 + SHOULD-FIX 3건 + NICE-TO-HAVE 2건 반영. 결함 2는 이번에도 무변경.

| 항목 | 변경 |
|------|------|
| 신호 대입 위치 (MUST) | `warmUpDone = true`를 `warmUpCommandTree` 본문 첫 줄로 고정. `TestMain` 별도 문장 대입 금지 — 호출부만 주석 처리했을 때 대입이 살아남아 가드가 통과하던 구멍을 봉쇄. `spec.md` REQ-CFS-004 + `plan.md` §B.2 대입 위치 규칙 + `acceptance.md` AC-CFS-004b 신설 |
| 도달 집합 도출 | `plan.md` §B.4에 0단계 추가 — `<parent>.AddCommand(<ident>)` 정적 스캔 + 루트 전이적 폐포. 하드코딩 리터럴 목록 금지(등록 시점에 낡음). Go 리플렉션 대안 불가 사유 명시 |
| M1 실행 격리 | `-run '^<M1TestName>$'` 앵커 + 신규 프로세스 20회 반복. 패키지 전체 실행은 선행 `Commands()` 호출(`internal/cli` 9곳 / `preference` 2곳)이 창을 먼저 닫아 무관한 비관측 유발 |
| Gaps 정합 | `spec.md` §G에 M1 재현의 확률적 비관측 항목 추가(7→8항). `acceptance.md` §F ⊆ `spec.md` §G 관계 복원 — §G가 정본이라는 선언이 이제 사실과 일치 |
| 마커 리터럴 | `progress.md`의 clarification 마커 언급에서 대괄호 리터럴 제거(MP-7 grep 오탐 방지) |
| AC-CFS-026 왕복 경로 | 구성 불가능한 첫 대안(도달 집합 밖 전역 테스트 추가 — `internal/cli` 전역이 모두 `rootCmd` 등록이라 프로덕션 수정 없이는 불가, REQ-CFS-021 저촉) 제거. 도달 집합 계산 일시 수정 경로만 채택 |

_run-phase 대기 중._

### Plan-phase completion signal

```yaml
plan_complete_at: 2026-07-25T02:32:28Z
plan_status: audit-ready
```

- plan-auditor iteration 이력: iter1 FAIL 0.69 → iter2 PASS 0.844 → iter3 잔여 지적 반영(본 0.3.0 개정).
- 다음 게이트: 구현 착수 승인(Implementation Kickoff Approval) — plan-auditor 점수와 무관한 별도 사용자 결정.
- 승인 시 제시할 신규 범위: REQ-CFS-006 인벤토리 가드(`plan.md` §B.4, 약 100-150 LOC, stdlib 전용).

## §E.2 Run-phase Evidence

### M1 — 결함 1 RED: 실제 전역 대상 레이스 재현

재현 **성공**. 두 패키지 모두 신규 프로세스 20회 반복에서 **20/20 실패**했다. 이는 `spec.md` §G 7항(재현의 확률적 비관측 가능성)이 실현되지 않았음을 뜻한다 — blocker 라우팅 불필요.

`plan.md` §F M1의 `-run` 앵커 + 신규 프로세스 반복 격리가 결정적이었다. 패키지 전체 실행에서는 선행 테스트가 레이스 창을 먼저 닫으므로 관측되지 않는다.

#### M1-a 재현 테스트 전체 소스 (`internal/cli/preference/race_repro_test.go`, M2에서 제거됨 — AP-3 보존 의무)

```go
package preference

// M1 reproduction test (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-020).
//
// TEMPORARY: this file is removed in M2 once the TestMain warm-up lands. Its
// full source and the observed failure output are preserved verbatim in
// .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/progress.md §E.2 (plan.md §G AP-3).

import (
	"sync"
	"testing"
)

// TestRaceRepro_PreferenceCmdLazySort reproduces the cobra lazy-sort data race
// on the real package-level PreferenceCmd global — the exact defect observed in
// GitHub Actions run 30135630413 attempt 1 (macos-latest).
//
// cobra v1.10.2 command.go:1332-1339 sorts c.commands in place INSIDE the
// Commands() accessor and writes c.commandsAreSorted there. Goroutines making
// the FIRST Commands() call concurrently race that write against each other's
// read, so the window exists only until the flag flips to true.
//
// MUST run in isolation so this test owns the first Commands() call in the
// process (a package-wide run lets an earlier test close the window first):
//
//	for i in $(seq 1 20); do \
//	  go test -race -count=1 -run '^TestRaceRepro_PreferenceCmdLazySort$' ./internal/cli/preference/; \
//	done
func TestRaceRepro_PreferenceCmdLazySort(t *testing.T) {
	const goroutines = 16

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start // release all goroutines at once to widen the race window
			_ = PreferenceCmd.Commands()
		}()
	}
	close(start)
	wg.Wait()
}
```

#### M1-b 재현 테스트 전체 소스 (`internal/cli/race_repro_test.go`, M2에서 제거됨)

```go
package cli

// M1 reproduction test (SPEC-CI-FLAKY-STABILIZE-001, REQ-CFS-020).
//
// TEMPORARY: this file is removed in M2 once the TestMain warm-up lands. Its
// full source and the observed output are preserved verbatim in
// .moai/specs/SPEC-CI-FLAKY-STABILIZE-001/progress.md §E.2 (plan.md §G AP-3).

import (
	"sync"
	"testing"
)

// TestRaceRepro_RootCmdLazySort is the internal/cli counterpart of the
// preference-package reproduction: concurrent first Commands() calls on the
// real rootCmd global, which is the receiver used by the parallel tests in
// harness_retirement_test.go / handoff_test.go.
//
// MUST run in isolation (see the preference-package test for the rationale):
//
//	for i in $(seq 1 20); do \
//	  go test -race -count=1 -run '^TestRaceRepro_RootCmdLazySort$' ./internal/cli/; \
//	done
func TestRaceRepro_RootCmdLazySort(t *testing.T) {
	const goroutines = 16

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			<-start
			_ = rootCmd.Commands()
		}()
	}
	close(start)
	wg.Wait()
}
```

#### M1-c 실행 명령 및 관측 (preference)

```
$ for i in $(seq 1 20); do go test -race -count=1 -run '^TestRaceRepro_PreferenceCmdLazySort$' ./internal/cli/preference/; done
preference: failures=20 / 20
```

iteration 1 verbatim 출력:

```
==================
WARNING: DATA RACE
Read at 0x0001049066c0 by goroutine 18:
  github.com/spf13/cobra.(*Command).Commands()
      /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go:1334 +0xbc
  github.com/modu-ai/moai-adk/internal/cli/preference.TestRaceRepro_PreferenceCmdLazySort.func1()
      .../internal/cli/preference/race_repro_test.go:39 +0x74

Previous write at 0x0001049066c0 by goroutine 10:
  github.com/spf13/cobra.(*Command).Commands()
      /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go:1336 +0x108
  github.com/modu-ai/moai-adk/internal/cli/preference.TestRaceRepro_PreferenceCmdLazySort.func1()
      .../internal/cli/preference/race_repro_test.go:39 +0x74
==================
--- FAIL: TestRaceRepro_PreferenceCmdLazySort (0.00s)
    testing.go:1712: race detected during execution of test
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/preference	0.449s
FAIL
```

**메커니즘 일치 확인**: `command.go:1334`(읽기) vs `command.go:1336`(쓰기) — `spec.md` §B.1이 인용한 CI 로그의 프레임과 **동일한 두 줄**이다. 즉 재현된 것은 원인 메커니즘 그 자체이며 유사물이 아니다.

#### M1-d 실행 명령 및 관측 (internal/cli)

```
$ for i in $(seq 1 20); do go test -race -count=1 -run '^TestRaceRepro_RootCmdLazySort$' ./internal/cli/; done
cli: failures=20 / 20
```

iteration 1 verbatim 발췌 (rootCmd는 자식이 많아 슬라이스 원소 스왑 경합까지 다수 발생):

```
==================
WARNING: DATA RACE
Write at 0x00c0003f6890 by goroutine 36:
  github.com/spf13/cobra.commandSorterByName.Swap()
      /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go:1328 +0xa0
  sort.partition()
  sort.pdqsort()
  sort.Sort()
  github.com/spf13/cobra.(*Command).Commands()
      /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go:1335 +0x100
  github.com/modu-ai/moai-adk/internal/cli.TestRaceRepro_RootCmdLazySort.func1()
      .../internal/cli/race_repro_test.go:34 +0x74

Previous read at 0x00c0003f6890 by goroutine 37:
  github.com/spf13/cobra.commandSorterByName.Less()
      /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go:1329 +0x48
...
--- FAIL: TestRaceRepro_RootCmdLazySort (0.02s)
    testing.go:1712: race detected during execution of test
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.722s
FAIL
```

#### M1-e 조사 부수 발견 (계획 외 관측, 코드 변경 없음)

`internal/cli/help_order.go:54` `reorderRootHelpCommands`가 `cobra.EnableCommandSorting = false`를 설정한다. 이 함수는 `root.go:71`(Execute 경로)과 `help_order_test.go` 5곳에서 호출된다. 즉 `internal/cli` 테스트 바이너리에서는 **비병렬 테스트가 먼저 실행되며 그 과정에서 정렬이 꺼지므로**, 이후 재개되는 병렬 테스트에는 레이스 창이 이미 닫혀 있을 수 있다 — CI에서 `internal/cli`가 아니라 `internal/cli/preference`만 실패한 것과 정합한다. 이는 **우연한 방어**이며 테스트 실행 순서에 의존한다. M2의 `TestMain` warm-up은 이 우연성에 의존하지 않는 명시적 보장을 제공한다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | M |
| scope (file count) | ~13 (10 test files edited, 3 new: 2x `main_test.go` + `inventory_guard_test.go`, 1 source: `registry.go`) |
| domain count | 2 (`internal/cli` + `internal/cli/preference`; `internal/session`) |
| file language mix | 100% Go |
| concurrency benefit | LOW — coding-heavy, with ordered RED→GREEN dependencies inside each defect track |
| Implementation Kickoff Approval | PASSED (user selected "run 진입 — 계획 그대로", inventory guard included) |

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 `trivial` | no | 13 files with semantic behavior change; not a typo/single-line edit |
| 2 `background` | no | Work is write-heavy (Write/Edit on source + tests), not read-only |
| 3 `agent-team` | no | RETIRED tombstone — never selected by the decision tree |
| 4 `parallel` | no | 2 domains (< 3 threshold) AND coding-heavy — Anthropic's coding-task parallelism caveat routes this to sequential |
| 5 `sub-agent` | **yes** | Default fallback; coding-heavy work with milestone ordering (M1 RED must precede M2 GREEN) |
| 6 `workflow` | no | ~13 files (« ~30 threshold) and the transformation is not a single uniform mechanical rule — two distinct defect mechanisms with reproduction-first ordering |

### Decision

Decision: `sub-agent`

### Justification

The run-phase is coding-heavy and carries hard intra-track ordering: M1 (RED, reproduce the race) must be observed failing before M2 (GREEN, TestMain warm-up), and M5 must precede M6 for the flock track. Parallel fan-out cannot preserve that ordering and would violate the reproduction-first requirement (REQ-CFS-020). Anthropic's coding-task parallelism caveat — most coding tasks involve fewer truly parallelizable tasks than research — makes Mode 5 the correct default here. The two defect tracks are independent of each other, but each is internally sequential, so a single `manager-develop` sequencing M1→M9 is simpler than coordinating two parallel agents for no wall-clock gain at this scope.

### Boundary case

Domain count is 2 against a 3-domain Mode 4 threshold — below threshold, no tie-breaker needed. File count ~13 sits above the 10-file Mode 4 signal but the coding-heavy tie-breaker rule (§B.2 "Coding-heavy + multi-domain: prefer Mode 5 over Mode 4") resolves it to Mode 5 regardless.
