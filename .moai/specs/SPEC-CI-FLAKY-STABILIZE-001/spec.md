---
id: SPEC-CI-FLAKY-STABILIZE-001
title: "CI Flaky 안정화 — cobra lazy-sort 데이터 레이스 + session registry flock 기아"
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

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | 최초 작성 — CI 로그로 확인된 2개 flaky 결함(cobra lazy-sort 레이스 / flock 기아)의 근본 원인 수정 SPEC |
| 0.2.0 | 2026-07-25 | manager-spec | plan-auditor FAIL(0.69) 반영. 결함 1 전면 개정 — 결함 클래스 재분류(구문 후보 11 중 전역 수신자 7), REQ-CFS-002 재귀 warm-up 의무화(`githubCmd` 누락 수정), REQ-CFS-004 관측 기전을 정렬 상태→도달 가능성 신호로 교체, REQ-CFS-006 검증 가능화, §G에 `-count` 의미론 공백 추가. 결함 2는 변경 없음 |
| 0.3.0 | 2026-07-25 | manager-spec | plan-auditor PASS(0.844) 잔여 지적 반영. REQ-CFS-004에 신호 대입 위치 규칙 추가(warm-up 함수 본문 내부 강제 — `TestMain` 별도 문장 금지), §G에 M1 재현의 확률적 비관측 공백 추가(7→8항). `plan.md` §B.4에 도달 집합 정적 도출 절차, M1에 `-run` 앵커 실행 격리 추가. `acceptance.md`에 AC-CFS-004b 신설. 결함 2는 변경 없음 |

---

## §A. 배경 (Context)

moai-adk-go CI에서 두 종류의 간헐적(flaky) 실패가 재발하고 있다. 두 결함 모두 **GitHub Actions 실행 로그에서 verbatim으로 확인**되었으며, 추정이 아니다. 브랜치 보호가 `strict: true`이므로 flaky 실패 1건이 머지 반복 지연으로 직결된다.

| 결함 | 패키지 | 실패 형태 | 근거 |
|------|--------|-----------|------|
| 결함 1 | `internal/cli/preference` | `WARNING: DATA RACE` → 해당 패키지 내 테스트 11개가 연쇄 FAIL | GH Actions run `30135630413` attempt 1 (macos-latest) |
| 결함 2 | `internal/session` | `TestRegisterSessionConcurrent` 65.75s 후 `ErrLockTimeout` | GH Actions run `30136807508` attempt 1 (ubuntu-latest) |

> **용어 구분 (혼동 주의 — 두 개의 "11")**: 위 표의 "테스트 11개 연쇄 FAIL"은 **CI 1회 실행에서 `internal/cli/preference` 패키지 내부에 파급된 실패 건수**다. §C.1의 "구문 후보 11건"은 **두 패키지에 걸친 AWK 구문 필터 결과 집합**이며, 그중 실제 결함 클래스는 7건이다. 두 숫자가 우연히 같을 뿐 서로 무관한 집합이다. 이하 본문에서는 전자를 **"CI 파급 11건"**, 후자를 **"구문 후보 11건"**으로 구분해 부른다.

두 결함 모두 **로컬에서 재현되지 않는다**(§G 참조). 따라서 본 SPEC은 "로컬 red→green 재현"이 아니라 **메커니즘 정확성 + CI 미재발**을 검증 축으로 삼는다. 이 한계는 은폐하지 않고 §G에 명시한다.

---

## §B. 문제 정의 (Problem Statement)

### B.1 결함 1 — cobra `Commands()` 지연 정렬(lazy sort) 데이터 레이스

cobra v1.10.2 `command.go:1332-1339`:

```go
// Commands returns a sorted slice of child commands.
func (c *Command) Commands() []*Command {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}
```

`EnableCommandSorting`은 기본 `true`다(`cobra.go:59`). 최초 호출 시 `c.commands`를 **접근자 내부에서 제자리(in place) 정렬**하고 `c.commandsAreSorted = true`를 **기록(write)** 한다(`command.go:1336`). 이후 호출은 해당 플래그를 **읽기(read)** 만 한다(`command.go:1334`).

`t.Parallel()`을 선언한 두 테스트가 **동일한 패키지 전역 `*cobra.Command` 변수**에 대해 최초 `Commands()` 호출을 동시에 수행하면, 쓰기와 읽기가 경합하여 race detector가 발화한다. 플래그가 `true`가 된 이후에는 창이 닫히므로 실패가 **간헐적**이다.

**결함 클래스는 전역 수신자를 공유하는 7건이다.** 구문 후보 11건 중 4건은 테스트마다 새 트리를 생성하는 로컬 수신자를 쓰므로 공유 상태가 없고 레이스가 성립하지 않는다(§C.1). 실제 CI에서 실패한 것은 그중 2건이지만, 나머지 5건은 동일 메커니즘을 그대로 갖고 있다.

### B.2 결함 2 — session registry flock 기아(starvation)

`internal/session/registry_lock_unix.go`는 `unix.LOCK_EX|unix.LOCK_NB`(비차단)로 락을 획득한다. 비차단이므로 경합 시 커널 대기열에 줄서지 않고 즉시 `EWOULDBLOCK`을 반환한다. `internal/session/registry.go`의 `withLock`은 이를 **고정 20ms 슬립 + 무지터(no jitter)** 루프로 재시도한다.

커널 큐 부재 + 균일한 깨어남 간격 = **공정성 없음**. 경합자들이 동일 주기로 재시도하여, 불운한 goroutine이 반복적으로 획득 경쟁에서 패배할 수 있다.

이것이 "전반적 느림"이 아니라 "기아"인 근거: `deadline`은 `withLock` **호출마다 재계산**되므로 60초 예산은 획득 1회당 예산이다. 해당 테스트는 10 goroutine × 100 등록 = 1000회 획득이며 로컬 전체 소요는 22.7초(획득당 약 23ms)다. 단일 `Register`가 60초를 초과한다는 것은 집계 지연으로 설명될 수 없고, 한 goroutine이 반복적으로 경쟁에서 밀렸음을 뜻한다.

---

## §C. 범위 (Scope)

### C.1 결함 1 대상 — 구문 후보 11건, 결함 클래스 7건

`t.Parallel()`과 `.Commands()`를 자기 함수 본문에 동시에 포함하는 테스트 전수(구문 후보)와, 각 `.Commands()` 호출의 **수신자 종류**:

| 파일 | 테스트 함수 | 수신자 | 종류 | 결함 클래스 | CI 실패 이력 |
|------|-------------|--------|------|-------------|--------------|
| `internal/cli/preference/cmd_test.go` | `TestPreferenceCmd_HasDecayScanChild` | `PreferenceCmd` | 패키지 전역 | 예 | 실패함 |
| `internal/cli/preference/cmd_test.go` | `TestPreferenceCmd_HasToggleChild` | `PreferenceCmd` | 패키지 전역 | 예 | 실패함 |
| `internal/cli/harness_retirement_test.go` | `TestHarnessV3R5VerbSurface` | `rootCmd` | 패키지 전역 | 예 | — |
| `internal/cli/handoff_test.go` | `TestHandoffCmdRegistered` | `rootCmd` | 패키지 전역 | 예 | — |
| `internal/cli/hook_e2e_test.go` | `TestHookSubcommands_AllNewEventsRegistered` | `hookCmd` | 패키지 전역 | 예 | — |
| `internal/cli/hook_e2e_test.go` | `TestHookValidEventTypes_AllHaveSubcommands` | `hookCmd` | 패키지 전역 | 예 | — |
| `internal/cli/github_integration_test.go` | `TestGithubCLI_ErrorPropagation` | `githubCmd` | 패키지 전역 | 예 | — |
| `internal/cli/harness_route_test.go` | `TestHarnessRouterCmd` | `newHarnessRouterCmd()` | 테스트 로컬 | **아니오** | — |
| `internal/cli/harness_clusters_test.go` | `TestHarnessClustersRegisteredInLiveTree` | `newHarnessRouterCmd()` | 테스트 로컬 | **아니오** | — |
| `internal/cli/harness_execute_test.go` | `TestExecuteCmd_RegisteredInRouter` | `newHarnessRouterCmd()` | 테스트 로컬 | **아니오** | — |
| `internal/cli/tool_policy_test.go` | `TestToolPolicyCmd_Registered` | `newToolPolicyCmd()` | 테스트 로컬 | **아니오** | — |

수신자 판정 근거:

- 전역: `PreferenceCmd` (`internal/cli/preference/cmd.go:38`), `githubCmd` (`internal/cli/github.go:89`; `rootCmd.AddCommand(githubCmd)`는 `github.go:100`, 테스트 호출부는 `github_integration_test.go:56` 및 `:182`), `rootCmd` / `hookCmd` (`internal/cli` 패키지 전역).
- 로컬: `newHarnessRouterCmd()` (`internal/cli/harness_route.go:59` — 호출마다 새 `&cobra.Command{}` 반환), `newToolPolicyCmd()` (`internal/cli/tool_policy.go:33` — 동일). 각 호출이 자기 `commandsAreSorted` 필드를 소유하므로 공유 상태·레이스·warm-up 효과가 모두 성립하지 않는다.

> **초판 정정**: 초판은 구문 후보 11건 전체를 결함 클래스로 서술했다. AWK 인벤토리는 **구문 필터**(`t.Parallel()` + `.Commands()` 동일 함수 본문 공존)이지 **결함 클래스 필터**(공유 전역 수신자)가 아니다. 사용자가 확정한 "11건 범위" 결정은 유지된다 — 재분류는 "무엇을 고치는가"를 정정할 뿐 "무엇을 약화하지 않는가"의 범위를 줄이지 않는다.

**구문 후보 11건 전체가 `t.Parallel()` 보존 대상이다**(REQ-CFS-005).

`internal/cli`와 `internal/cli/preference` 어느 쪽에도 현재 `func TestMain`이 존재하지 않는다(grep 확인). 두 패키지 모두 신규 생성 대상이다.

### C.2 결함 2 대상

- `internal/session/registry.go`의 `withLock` 재시도 루프 1개소.

### C.3 제외 항목 (Exclusions)

본 SPEC은 아래 항목을 명시적으로 out of scope로 둔다.

### Out of Scope — 락 아키텍처 변경

- 비차단 `LOCK_NB` → 차단 flock 전환. 커널 대기열 기반 공정성 확보는 구조적 해법이지만 CLI/hook 컨텍스트의 무한 블록 위험(`AP-MSC-005`)을 재도입하므로 본 SPEC에서 다루지 않는다.
- 파일 락 → 다른 동기화 기전(예: 락 서버, DB 트랜잭션) 교체.
- `Query` 경로의 락 비우회(현재 eventually-consistent read 설계) 변경.

### Out of Scope — 프로덕션 동작 변경

- `LockTimeout = 2 * time.Second` 기본값 변경.
- cobra 커맨드 트리 구성(부모/자식 등록 순서, 그룹 지정) 변경.
- `ErrLockTimeout` 에러 계약(반환 조건·래핑 형태) 변경.
- `cobra.EnableCommandSorting` 전역 토글 조작 — 정렬을 끄면 레이스는 사라지지만 CLI 도움말 출력 순서가 바뀌는 사용자 가시 변경이다.

### Out of Scope — 테스트 약화

- 구문 후보 11건에서 `t.Parallel()` 제거(직렬화로 레이스를 회피하는 방식).
- 어떤 테스트든 삭제·skip·조건부 비활성화.
- `-race` 플래그를 CI에서 제거하거나 대상 패키지만 예외 처리.

### Out of Scope — 인접 리팩터링

- 두 결함과 무관한 `internal/cli` / `internal/session` 코드 정리.
- 다른 패키지의 flaky 이력 조사·수정.
- CI 워크플로 재시도 정책(`retry` 횟수) 조정 — 증상 은폐이므로 근본 수정 대상이 아니다.
- 테스트 로컬 수신자 4건의 구성 변경 — 결함 클래스가 아니므로 손대지 않는다.

---

## §D. 요구사항 (GEARS)

### D.1 결함 1 — cobra lazy-sort 레이스

**REQ-CFS-001** (Ubiquitous)
`internal/cli` 패키지와 `internal/cli/preference` 패키지는 각각 `func TestMain(m *testing.M)`를 정의해야 한다(shall).

**REQ-CFS-002** (Event-driven)
**When** 테스트 바이너리가 기동되면, 각 패키지의 `TestMain`은 `m.Run()` 호출 **이전에** 해당 패키지의 루트 전역 커맨드(`internal/cli`는 `rootCmd`, `internal/cli/preference`는 `PreferenceCmd`)를 기점으로 **깊이 우선 재귀 순회**하며, 도달하는 모든 노드에 대해 `Commands()`를 1회씩 호출하여 지연 정렬을 완료시켜야 한다(shall).

재귀가 요구되는 이유: 대상 전역은 루트 하나가 아니다. `hookCmd`·`githubCmd`는 `rootCmd`의 자식으로 등록된 별개의 패키지 전역이며(`github.go:100` `rootCmd.AddCommand(githubCmd)`), 각각이 자기 `commandsAreSorted` 필드를 갖는다. 개별 커맨드를 열거하는 방식은 새 전역이 추가될 때마다 누락되므로 금지한다.

**REQ-CFS-003** (Ubiquitous)
warm-up 호출부는 cobra lazy-sort 메커니즘과 삭제 시 재발하는 결함을 설명하는 영문 주석을 동반해야 한다(shall). 주석은 `commandsAreSorted` 쓰기/읽기 경합과 `t.Parallel()` 상호작용을 명시해야 한다.

**REQ-CFS-004** (Ubiquitous)
각 패키지는 warm-up 제거를 **결정론적으로** 탐지하는 가드 테스트를 보유해야 한다(shall). 가드는 **도달 가능성 신호**(reachability signal)를 관측해야 한다 — 즉 warm-up 코드 경로가 실제로 실행되었음을 기록하는 테스트 스코프 변수를 확인해야 한다.

신호 변수의 대입은 **warm-up 함수 본문 안**에 위치해야 하며(shall), `TestMain`에서 warm-up 호출과 분리된 별도 문장으로 대입해서는 안 된다(shall not). 판별 기준: **warm-up 호출 한 줄을 제거했을 때 신호 대입도 함께 도달 불가가 되어야 한다.** 대입이 호출부 바깥에 있으면 호출만 제거해도 신호가 설정된 채 남아 가드가 통과하므로, 요구된 결정론적 탐지가 성립하지 않는다.

가드는 커맨드 슬라이스의 **정렬 상태를 관측해서는 안 된다**(shall not). 근거: `Commands()`는 접근자 내부에서 제자리 정렬을 수행하므로(§B.1 인용 코드), 정렬 상태는 warm-up 실행 여부와 무관하게 **어떤 호출 이후에도 항상 참**이다. 정렬 기반 가드는 자신이 탐지하려는 변이를 스스로 일으키며, 어떤 트리 형태에서도 실패할 수 없다(구조적 반증 불가).

**REQ-CFS-005** (Unwanted)
본 SPEC의 변경은 §C.1의 구문 후보 11건 어느 것에서도 `t.Parallel()` 선언을 제거해서는 안 된다(shall not). 이 보존 범위는 결함 클래스 7건이 아니라 후보 11건 전체에 적용된다.

**REQ-CFS-006** (Event-driven)
**When** 테스트가 실행되면, 인벤토리 가드 테스트는 대상 패키지의 `_test.go` 파일에서 `t.Parallel()`과 `.Commands()`를 함께 포함하는 함수를 재도출하고 각 `.Commands()` 수신자를 해석하여, **warm-up 재귀가 도달하지 않는 패키지 전역**을 수신자로 갖는 항목이 발견되면 실패를 보고해야 한다(shall).

이 요구사항은 신규 병렬 테스트가 warm-up 범위 밖의 전역을 도입하는 재발 경로를 막는 유일한 지속 방어선이다.

### D.2 결함 2 — flock 기아

**REQ-CFS-010** (Ubiquitous)
`internal/session`의 `withLock` 재시도 루프는 고정 간격 대신 **무작위 지터(jitter)를 포함한 지수 백오프(exponential backoff)** 로 대기해야 한다(shall).

**REQ-CFS-011** (Ubiquitous)
백오프 대기 시간은 상한(cap)을 가져야 하며, 상한은 유효 `lockTimeout`을 초과해서는 안 된다(shall).

**REQ-CFS-012** (Unwanted)
본 SPEC의 변경은 `LockTimeout` 상수의 기본값 `2 * time.Second`를 변경해서는 안 된다(shall not).

**REQ-CFS-013** (Unwanted)
본 SPEC의 변경은 `unix.LOCK_NB`(비차단) 획득 방식을 차단 flock으로 전환해서는 안 된다(shall not).

**REQ-CFS-014** (Event-driven)
**When** 재시도 루프가 유효 `lockTimeout` 데드라인을 초과하면, `withLock`은 기존과 동일하게 `ErrLockTimeout`으로 래핑된 오류를 반환해야 한다(shall). 반환 오류의 래핑 형태와 `errors.Is(err, ErrLockTimeout)` 판정 결과는 변경 전후 동일해야 한다.

**REQ-CFS-015** (Ubiquitous)
`internal/session`은 경합 하 **획득 1회당 대기 시간 분포 또는 최대 단일 획득 대기 시간**을 관측하는 기아 특성화(starvation characterization) 테스트를 보유해야 한다(shall).

**REQ-CFS-016** (Where — capability gate)
**Where** 테스트가 `-short` 모드로 실행되는 경우, 기아 특성화 테스트는 skip되어야 한다(shall) — 기존 `TestRegisterSessionConcurrent`의 `-short` 처리 관례와 일치시킨다.

### D.3 공통 — 절차 및 품질

**REQ-CFS-020** (Event-driven)
**When** 각 결함의 수정이 착수되면, 재현/특성화 테스트가 수정 코드보다 **먼저** 작성되고 그 실패 또는 관측 결과가 기록된 후에 수정이 적용되어야 한다(shall). (CLAUDE.md §7 Rule 4 재현 우선)

**REQ-CFS-021** (Unwanted)
본 SPEC의 변경은 §C.1·§C.2에 열거된 파일 외의 프로덕션 코드를 수정해서는 안 된다(shall not).

**REQ-CFS-022** (Ubiquitous)
변경 후 전체 테스트 스위트(`go test ./...`), `go vet ./...`, `golangci-lint run`이 모두 통과해야 한다(shall).

**REQ-CFS-023** (Ubiquitous)
모든 변경은 PR을 경유해야 한다(shall) — 브랜치 보호 `enforce_admins: true`로 `main` 직접 푸시가 차단되어 있다.

---

## §E. 성공 기준 (Success Criteria)

| # | 기준 | 측정 방법 |
|---|------|-----------|
| 1 | 결함 1 재현 테스트가 수정 전 `-race`에서 실패 | 실제 전역(`PreferenceCmd` / `rootCmd`) 대상 테스트의 수정 전 실행 출력 |
| 2 | 결함 1 수정 후 두 패키지가 **신규 프로세스 50회 반복**에서 전부 통과 | 프로세스별 `-count=1` 루프의 종료 코드 0 |
| 3 | 구문 후보 11건 모두 `t.Parallel()` 유지 | 소스 인벤토리 재도출 = 11 |
| 4 | 결함 2 특성화 테스트가 최대 단일 획득 대기 시간을 수치로 보고 | 테스트 로그 출력 |
| 5 | jitter+backoff 적용 후 최대 단일 획득 대기 시간 개선 | 적용 전후 수치 대조 |
| 6 | `LockTimeout` 기본값 불변 | 소스 grep |
| 7 | 전체 스위트 · vet · lint green | 각 명령 종료 코드 0 |

상세 검증 명령과 기대 출력은 `acceptance.md` §D AC 매트릭스에 있다.

---

## §F. 의존성 및 제약

- 개발 모드: `tdd` (`.moai/config/sections/quality.yaml` `constitution.development_mode`)
- Go 1.26.4, cobra v1.10.2
- 대상 브랜치: `feat/SPEC-CI-FLAKY-STABILIZE-001` (base `origin/main` @ `03c47e3bb`)
- 브랜치 보호: `enforce_admins: true`, `strict: true` — 머지 전 base 최신화 필요
- 커밋 규약: Conventional Commits, 영문
- 코드·주석: 영문 / SPEC 산출물 산문: 한국어

---

## §G. 검증 공백 (Gaps — 미검증 사항)

`.claude/rules/moai/core/verification-claim-integrity.md` §3.4에 따라, 관측되지 **않은** 것을 명시한다.

1. **두 결함 모두 로컬에서 재현되지 않았다.**
   - 결함 1: `go test -race -count=50 ./internal/cli/preference/` → `ok ... 39.994s`, 종료 코드 0.
     증거: `.moai/state/verify/b5fc437f/pref-race-50.log`
   - 결함 2: `go test -race -count=20 -run TestRegisterSessionConcurrent ./internal/session/` → `ok ... 492.390s`, 종료 코드 0.
     증거: `.moai/state/verify/b5fc437f/session-flock-20.log`

2. **결함 1의 위 로컬 실행은 50회 독립 시도가 아니다.** `go test -count=N`은 **단일 프로세스 안에서** N회 반복하며 패키지 수준 상태가 반복 간에 유지된다. `commandsAreSorted`는 전역 `*cobra.Command`의 1회성 플래그이므로, 반복 1이 레이스 창을 닫고 반복 2~50은 창이 이미 닫힌 상태로 실행된다. 따라서 이 로그는 **유효 시도 1회 + 무효 반복 49회**이며, "50회 시도해도 재현 안 됨"으로 읽어서는 안 된다. 결함 1의 로컬 미재현 증거는 사실상 **1회 시도**에 불과하다.
   (결함 2의 `-count=20`에는 이 문제가 없다 — 매 반복이 새 임시 디렉터리를 만들고 프로세스 전역 1회성 플래그에 의존하지 않으므로 20회가 실질적으로 독립이다.)

3. 따라서 본 SPEC의 수정은 **원래 CI 실패의 로컬 red→green 재현으로 검증되지 않는다.** 검증 축은 (a) 메커니즘 정확성 — CI 로그가 지목한 정확한 코드 경로를 수정했는가, (b) CI 미재발 — 병합 후 동일 실패가 관측되지 않는가, 두 가지다.

4. 결함 2의 특성화 테스트는 기아를 **결정론적으로 재현하지 않는다.** 경합 하 대기 분포를 관측할 뿐이며, 로컬에서 기아가 발생하지 않으면 "개선 전후 대조"는 꼬리 지연(tail latency) 감소로만 나타날 수 있고 기아 소멸을 직접 보이지 못한다.

5. **결함 1의 미실패 5건**(전역 수신자이면서 CI 실패 이력이 없는 항목)은 수정 효과를 관측할 수 없다. 실패한 적이 없으므로 수정 후에도 "여전히 실패하지 않음"만 확인된다.

6. CI 미재발은 **유한 횟수 관측**이다. 간헐적 결함에 대해 N회 성공은 부재의 증거로서 약하다.

7. **결함 1의 재현(M1)은 확률적이며 비관측이 가능하다.** M1 테스트는 실제 전역(`PreferenceCmd`/`rootCmd`)을 대상으로 하지만, 레이스 창은 최초 `Commands()` 호출 순간에만 열리므로 goroutine 스케줄링에 따라 20회 반복에서도 관측되지 않을 수 있다. 비관측은 "결함 없음"이 아니라 **blocker**로 라우팅되며(AC-CFS-009), 이 경우 결함 1은 RED 증거 없이 GREEN만 갖게 된다. 이는 §H 잔여 위험 6(신규 프로세스 **검증 루프**의 확률성)과 다른 항목이다 — 여기서는 **재현 자체**의 확률성을 말한다.

8. **증거 로그 2건이 자기기술적(self-describing)이지 않다.** `pref-race-50.log` / `session-flock-20.log`는 `ok … <duration>` + `exit=0`만 담고 있고 호출 명령줄을 포함하지 않는다. `verification-claim-integrity.md` §3.2는 Evidence를 "명령 + verbatim 출력"으로 규정하므로, 위 1·2항의 명령 문자열은 이 SPEC 본문이 보증하는 것이지 로그 파일 자체가 보증하는 것이 아니다. 본 계획은 이 로그를 재생성하지 않는다(재생성 시 원래 관측 시점과 달라짐).

---

## §H. 잔여 위험 (Residual Risk)

1. **jitter+backoff는 확률적 완화이지 기아의 구조적 제거가 아니다.** 비차단 flock에는 커널 공정성 큐가 없으므로, 충분히 긴 경합 하에서 한 경합자가 반복 패배할 확률은 0이 아니다. 사용자는 이 잔여 위험을 명시적으로 수용했다 — 실제 프로덕션 경합은 세션 수 개 수준이며 테스트의 10개 hot goroutine과 다르다는 근거다.

2. **TestMain warm-up은 미래 기여자가 삭제 가능하다.** REQ-CFS-003(의도 주석) + REQ-CFS-004(도달 가능성 가드) + REQ-CFS-006(인벤토리 가드)로 완화한다. REQ-CFS-004 가드는 warm-up 삭제 시 **결정론적으로** 실패하므로 초판의 정렬 기반 가드(구조적으로 항상 통과)보다 실질 방어력이 있다. 다만 가드 테스트 자체가 함께 삭제되면 방어가 무력화되는 근본 한계는 남는다.

3. **REQ-CFS-006 인벤토리 가드는 수신자 해석 범위 안에서만 동작한다.** 단순 식별자 수신자(`PreferenceCmd.Commands()`)는 정적 해석되지만, 지역 변수를 경유하거나 함수 반환값을 즉시 호출하는 형태는 해석이 어려워 누락될 수 있다. 가드는 완전 검증이 아니라 흔한 재발 경로의 차단이다.

4. **CI 러너 특성 의존.** 결함 1은 macos-latest, 결함 2는 ubuntu-latest에서 각각 관측되었다. 러너 코어 수·스케줄러 특성이 바뀌면 재발 확률이 변동한다.

5. **백오프 상한 선택이 지연을 유발할 수 있다.** 상한이 과도하면 저경합 상황에서 불필요한 대기가 추가되어 `moai session register` 등 CLI 경로의 응답 지연으로 나타날 수 있다.

6. **신규 프로세스 반복(§E 기준 2)도 확률적이다.** 50개 독립 프로세스 각각이 진짜 레이스 창을 1회씩 갖지만, 창이 좁아 warm-up 부재 상태에서도 50회 전부 통과할 수 있다. 이 반복은 회귀 탐지 보조 수단이며 결함 부재의 증명이 아니다 — 결정론적 방어는 REQ-CFS-004 가드가 담당한다.
