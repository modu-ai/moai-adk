---
id: SPEC-CI-FLAKY-STABILIZE-001
title: "CI Flaky 안정화 — 구현 계획"
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

## §A. 컨텍스트

`spec.md` §A/§B가 SSOT다. 본 문서는 조사 결과(별도 `research.md` 없이 여기 통합)와 실행 계획을 담는다.

개발 모드: **tdd** (RED → GREEN → REFACTOR). 두 결함 모두 "테스트 먼저" 순서가 밀레스톤에 강제되어 있다(§F).

버전 0.2.0 개정 요약: 결함 1의 가드 관측 기전(§B.2), warm-up 범위(§B.3), 재현 테스트 대상(§F M1), 반복 검증 명령(§E)이 전면 교체되었다. 결함 2(백오프 파라미터·데드라인 클램프·오류 계약)는 변경 없다.

버전 0.3.0 개정 요약: §B.2에 **대입 위치 규칙**(신호 대입은 warm-up 함수 본문 내부 — `TestMain` 별도 문장 금지) 신설, §B.3 코드 개형과 §F M2 단계를 그 규칙에 맞춰 정정, §B.4에 도달 집합 **정적 도출 절차(0단계)** 추가, §F M1에 `-run` 앵커 기반 **실행 격리** 추가. 결함 2는 변경 없다.

---

## §B. 결정 표면 (Decision Surface — 변경 가능성 높은 순)

리뷰 시 이 절을 먼저 볼 것. 아래로 갈수록 기계적이고 되돌리기 쉽다.

### B.1 [최상위 결정] 백오프 파라미터 (결함 2)

가장 바뀔 가능성이 큰 결정이며, 코드 한 곳(`withLock` 루프)에 국소화된다.

제안 기본값:

| 파라미터 | 제안값 | 근거 |
|----------|--------|------|
| base | `5ms` | 현행 고정 20ms보다 짧게 시작해 저경합 시 응답성 개선 |
| 배수(multiplier) | `2.0` | 표준 지수 백오프 |
| 상한(cap) | `50ms` | 획득당 로컬 평균 23ms의 약 2배. 상한이 과도하면 CLI 지연(`spec.md` §H 잔여 위험 5) |
| 지터 형태 | **full jitter** — `sleep = rand(0, min(cap, base*2^n))` | 균일 주기 동기화를 깨는 것이 목적이므로 equal jitter보다 분산이 큰 full jitter가 적합 |
| 난수원 | `math/rand/v2` (crypto 불필요) | 보안 목적 아님, 전역 시드 오염 회피 |

되돌리기 비용: 낮음(상수 조정). 다만 **cap을 lockTimeout보다 크게 잡으면 REQ-CFS-011 위반**이므로 `min(cap, remaining-to-deadline)` 클램프가 필요하다.

### B.2 [상위 결정] 가드 테스트의 관측 기전 (결함 1) — 초판에서 교체됨

`REQ-CFS-004`는 warm-up의 **효과**를 관측하라고 요구한다. 초판이 채택했던 정렬 상태 관측은 **구조적으로 반증 불가**하므로 폐기한다.

| 후보 | 방식 | 판정 |
|------|------|------|
| (a) 소스 grep 가드 | `TestMain` 소스에 warm-up 호출 문자열이 있는지 확인 | **부적합** — 도달 가능성 미증명. 호출을 주석 처리해도 통과 |
| (b) 정렬 상태 관측 (초판 채택 → **폐기**) | `Commands()` 결과가 `Name()` 오름차순인지 확인 | **부적합 — 반증 불가** (아래 근거) |
| (c) 도달 가능성 신호 (채택) | `main_test.go`의 테스트 스코프 변수 `warmUpDone`를 warm-up 내부에서 `true`로 설정하고, 가드가 이를 확인 | **채택** |

**(b) 폐기 근거 (cobra v1.10.2 `command.go:1332-1339` 실측)**: `Commands()`는 접근자 **내부에서** `sort.Sort(...)` 후 `c.commandsAreSorted = true`를 설정한다. `EnableCommandSorting`은 기본 `true`(`cobra.go:59`). 즉 warm-up 실행 여부와 무관하게, 가드가 `Commands()`를 호출하는 그 순간 정렬이 수행되므로 반환 슬라이스는 **항상** 정렬되어 있다. 가드는 자신이 탐지하려는 변이를 스스로 일으킨다. 초판이 "등록 순서가 우연히 사전순인 경우 통과할 수 있다"고 서술한 것은 과소평가였다 — 실제로는 **모든 트리 형태에서 항상 통과**하며, warm-up을 주석 처리해도 절대 FAIL하지 않는다. 초판의 왕복 검증(AC-CFS-005)은 성립할 수 없었다.

초판이 (c)를 기각하며 든 사유("프로덕션 코드에 테스트 전용 상태 추가")는 **본 케이스에 적용되지 않는다**. `TestMain`은 이미 `main_test.go`(테스트 전용 파일)에 존재하며, `warmUpDone` 변수도 같은 파일의 패키지 스코프에 둔다. 프로덕션 바이너리에는 어떤 심볼도 추가되지 않는다(`_test.go`는 프로덕션 빌드에 포함되지 않음). 따라서 기각 사유가 사실과 다르며, (c)를 채택한다.

(c) 구현 개형(패키지당 1개, `main_test.go` 내):

```go
// warmUpDone records that the cobra lazy-sort warm-up below actually ran.
// The guard test asserts this; commenting out warmUpCommandTree makes it fail
// deterministically. Do NOT assert on slice sortedness instead — Commands()
// sorts in place inside the accessor, so sortedness is always true after any
// call and cannot detect a removed warm-up.
var warmUpDone bool
```

가드 테스트는 `t.Parallel()` 없이(직렬) 실행하며 `warmUpDone == true`만 확인한다. 이 신호는 순환 논리가 아니다 — warm-up **코드 경로의 실행 여부**를 기록할 뿐, warm-up의 *효과*(정렬 완료)를 재확인하지 않는다. 효과 자체는 cobra 소스가 보증한다(호출되면 반드시 정렬된다).

#### 대입 위치 규칙 (HARD)

`warmUpDone = true`는 **`warmUpCommandTree` 함수 본문 첫 줄**에 두어야 한다. `TestMain` 안에서 `warmUpCommandTree(rootCmd)` **호출 다음 줄의 별도 문장으로 대입하는 것은 금지한다.**

금지 사유: AC-CFS-005·GWT-2의 왕복 검증은 `warmUpCommandTree(...)` **호출 한 줄만** 주석 처리한다. 대입이 `TestMain`에 별도 문장으로 있으면 그 줄은 여전히 실행되어 `warmUpDone`이 `true`로 남고 가드가 **통과**한다 — 즉 "결정론적 FAIL"이 성립하지 않는다. 이는 iter1 M1(정렬 상태 가드)과 동일한 실패 계열이 좁은 자리에서 재발한 형태이므로 명시적으로 봉쇄한다.

판별 기준 한 줄: **"warm-up 호출 한 줄을 주석 처리했을 때 대입도 함께 도달 불가가 되는가?"** — 되면 적법, 안 되면 위반.

되돌리기 비용: 낮음(테스트 파일 국소).

### B.3 [중간 결정] warm-up 범위 — 재귀 확정 (초판의 "옵션"에서 승격)

초판은 재귀 warm-up을 "되돌리기 쉬운 권장 옵션"으로 제시했고, REQ-CFS-002는 `rootCmd`/`hookCmd` 2개만 열거했다. 이 불일치가 `githubCmd`(`internal/cli/github.go:89` 패키지 전역, `github_integration_test.go:56`/`:182`에서 `.Commands()` 호출)를 누락시켰다.

개정: REQ-CFS-002가 **재귀를 의무화**한다. 계획은 요구사항을 따르며 더 이상 선택지가 아니다.

구현 개형:

```go
func warmUpCommandTree(c *cobra.Command) {
	warmUpDone = true                  // MUST be inside this function — never in TestMain
	for _, sub := range c.Commands() { // triggers the lazy sort on c
		warmUpCommandTree(sub)
	}
}
```

플래그 대입은 **반드시 이 함수 본문 안**에 있어야 한다(§B.2 대입 위치 규칙). `TestMain`에서 별도 문장으로 대입하면 호출부만 주석 처리했을 때 플래그가 그대로 설정되어 가드가 통과한다.

`rootCmd`를 기점으로 호출하면 `rootCmd.AddCommand(githubCmd)`(`github.go:100`)를 통해 `githubCmd`에, 동일 경로로 `hookCmd`에 도달한다. `internal/cli/preference`는 `PreferenceCmd`를 기점으로 호출한다.

### B.4 [중간 결정] REQ-CFS-006 인벤토리 가드 — 신규 범위 (비용 명시)

REQ-CFS-006은 초판에서 검증 불가 상태였다(AC가 "기존 가드가 통과한다"만 확인). 개정판은 실제 AC를 부여하며, 이는 **신규 작업 범위 추가**다. 오케스트레이터가 구현 착수 승인 게이트에서 사용자에게 제시할 수 있도록 비용을 명시한다.

- **산출물**: 테스트 파일 1개(`internal/cli/inventory_guard_test.go`). `internal/cli/preference`는 후보가 2건뿐이고 수신자가 `PreferenceCmd` 단일이므로 별도 가드 없이 `internal/cli` 가드가 두 패키지 디렉터리를 함께 스캔한다.
- **규모 추정**: 약 100-150 LOC. `go/parser` + `go/ast` 기반 4단계:
  - **(0) warm-up 도달 집합의 정적 도출** — 하드코딩 리터럴 목록 금지. 비-테스트 파일에서 `<parent>.AddCommand(<ident>)` 형태의 호출을 전부 수집해 부모→자식 인접 관계를 만들고, warm-up 루트(`rootCmd`, `PreferenceCmd`)에서 **전이적 폐포**(transitive closure)를 계산한다.
  - (1) `_test.go`에서 함수 본문에 `t.Parallel()` 호출과 `.Commands()` 선택자가 공존하는 함수를 수집.
  - (2) 각 `.Commands()` 수신자 표현식이 **단순 식별자**인 경우, 그 식별자가 비-테스트 파일의 패키지 레벨 `var` 선언에 해당하는지 확인.
  - (3) 해당하면서 (0)의 도달 집합에 없으면 FAIL.

- **(0)이 정적 도출이어야 하는 이유**: 손으로 관리하는 리터럴 목록은 **새 전역이 등록되는 바로 그 순간** 낡는다 — REQ-CFS-006이 막으려는 재발 시나리오와 정확히 같은 시점이다. `AddCommand` 스캔 방식은 등록과 동시에 자동 갱신되며, 신규 전역이 병렬 테스트에서 쓰이지만 warm-up 루트 아래에 **등록되지 않은** 경우를 정확히 FAIL로 잡는다.
- **런타임 순회가 대안이 될 수 없는 이유**: Go는 패키지 레벨 변수의 *이름*에 대한 리플렉션을 제공하지 않는다. `rootCmd.Commands()`를 런타임에 순회해도 AST 식별자(`githubCmd`)를 런타임 `*cobra.Command` 값에 대응시키려면 결국 손으로 관리하는 이름↔값 브리지가 필요하므로, 회피하려던 문제가 그대로 돌아온다. 따라서 (0)은 **순수 정적** 도출이다.
- **의존성**: 표준 라이브러리만(`go/ast`, `go/parser`, `go/token`). 신규 모듈 의존 없음.
- **한계**: 수신자가 지역 변수·함수 반환값 즉시 호출인 경우 정적 해석이 어려워 "로컬 수신자"로 분류하고 통과시킨다(보수적 false-negative). 이 한계는 `spec.md` §H 잔여 위험 3에 기록되어 있다.
- **경계 유지**: 단일 집중 테스트이며 프레임워크가 아니다. 범용 lint 규칙 확장은 out of scope.

되돌리기 비용: 중간(파일 1개 삭제로 원복 가능하나, 삭제하면 REQ-CFS-006이 검증 불가로 되돌아감).

### B.5 [하위 결정] 특성화 테스트의 관측 지표 (결함 2)

`REQ-CFS-015`는 "획득 1회당 대기 분포 **또는** 최대 단일 획득 대기 시간"을 요구한다. 최대값 단독은 노이즈에 취약하므로 **p50 / p95 / max 3종**을 함께 기록할 것을 권장한다. 테스트는 `t.Log`로 수치를 출력하고, **어서션은 최대값 상한 하나만** 건다(과도한 어서션은 그 자체가 flaky 원인이 된다).

### B.6 [기계적] TestMain 배치·주석 문안

파일 위치는 각 패키지의 신규 `main_test.go`. 문안은 REQ-CFS-003 요구사항을 만족하면 자유.

---

## §C. 조사 결과 (research 통합 — 이미 검증됨, 재조사 금지)

### C.1 결함 1 근거

증거 파일: `.moai/state/verify/b5fc437f/ci-30135630413-a1.log` 17-65행 (GH Actions run `30135630413`, attempt 1, macos-latest).

```
WARNING: DATA RACE
Read at 0x0001026067c0 by goroutine 15:
  github.com/spf13/cobra.(*Command).Commands()
      cobra@v1.10.2/command.go:1334
  .../internal/cli/preference.TestPreferenceCmd_HasDecayScanChild()
      internal/cli/preference/cmd_test.go:228
Previous write at 0x0001026067c0 by goroutine 16:
  github.com/spf13/cobra.(*Command).Commands()
      cobra@v1.10.2/command.go:1336
  .../internal/cli/preference.TestPreferenceCmd_HasToggleChild()
      internal/cli/preference/cmd_test.go:244
```

동일 실행에서 패키지 내 11개 테스트가 `--- FAIL ... (0.00s)` + `testing.go:1712: race detected during execution of test`로 연쇄 실패했고, `FAIL github.com/modu-ai/moai-adk/internal/cli/preference 0.402s`로 종료했다(= "CI 파급 11건", `spec.md` §A 용어 구분 참조).

메커니즘 및 cobra 소스 인용은 `spec.md` §B.1에 있다(SSOT). 요점: `Commands()`가 접근자 내부에서 제자리 정렬 + 플래그 기록을 수행하므로, 전역 커맨드에 대한 최초 동시 호출이 write/read 경합을 만든다.

인벤토리 재도출 명령(**구문 필터** — 결함 클래스 필터가 아님):

```bash
awk '
/^func Test/ { name=$2; sub(/\(.*/,"",name); body=""; inf=1 }
inf { body = body $0 }
/^}$/ { if (inf && body ~ /t\.Parallel\(\)/ && body ~ /\.Commands\(\)/) print FILENAME": "name; inf=0 }
' internal/cli/*_test.go internal/cli/preference/*_test.go | sort
# → 11행 (구문 후보). 이 중 전역 수신자 7건만이 결함 클래스 — spec.md §C.1 표 참조.
```

수신자 판정 실측:

```
internal/cli/github.go:89        var githubCmd = &cobra.Command{
internal/cli/github.go:100       rootCmd.AddCommand(githubCmd)
internal/cli/github_integration_test.go:56   for _, cmd := range githubCmd.Commands() {
internal/cli/github_integration_test.go:182          for _, cmd := range githubCmd.Commands() {
internal/cli/harness_route.go:59 func newHarnessRouterCmd() *cobra.Command {   → cmd := &cobra.Command{...} (호출마다 신규)
internal/cli/tool_policy.go:33   func newToolPolicyCmd() *cobra.Command {      → cmd := &cobra.Command{...} (호출마다 신규)
```

`func TestMain` 부재 확인: `grep -rn "func TestMain" internal/cli/` → 출력 없음.

### C.2 결함 2 근거

증거 파일: `.moai/state/verify/b5fc437f/ci-30136807508-a1.log` 90-91행 (GH Actions run `30136807508`, attempt 1, ubuntu-latest).

```
--- FAIL: TestRegisterSessionConcurrent (65.75s)
    registry_test.go:316: concurrent Register: session registry: lock acquisition timed out: registry lock flock /tmp/TestRegisterSessionConcurrent1231906729/001/active-sessions.json.lock: resource temporarily unavailable
FAIL	github.com/modu-ai/moai-adk/internal/session	67.033s
```

`registry_lock_unix.go:40`:

```go
if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
```

`registry.go:317-340` `withLock` 현행 루프:

```go
deadline := time.Now().Add(timeout)
for {
    err := lock.acquire(lockPath)
    if err == nil { break }
    if time.Now().After(deadline) {
        return fmt.Errorf("%w: %v", ErrLockTimeout, err)
    }
    time.Sleep(20 * time.Millisecond)
}
```

기아 판정 근거: `deadline`이 `withLock` 호출마다 재계산되므로 60초는 **획득 1회당** 예산이다. 테스트(`registry_test.go:283-331`)는 `r.WithLockTimeout(60 * time.Second)` 후 10 goroutine × 100 등록 = 1000회 획득을 수행하며, 로컬 단일 실행 전체가 22.7초(획득당 약 23ms)다. 단일 `Register`가 60초를 넘는 것은 집계 지연으로 설명 불가하다.

프로덕션 기본값: `registry.go:68` `LockTimeout = 2 * time.Second`.

### C.3 로컬 미재현 (GAP — 은폐 금지)

| 결함 | 명령 | 결과 | 유효 시도 횟수 | 증거 |
|------|------|------|----------------|------|
| 1 | `go test -race -count=50 ./internal/cli/preference/` | `ok ... 39.994s`, exit 0 | **1회** (아래 참조) | `.moai/state/verify/b5fc437f/pref-race-50.log` |
| 2 | `go test -race -count=20 -run TestRegisterSessionConcurrent ./internal/session/` | `ok ... 492.390s`, exit 0 | 20회 (실질 독립) | `.moai/state/verify/b5fc437f/session-flock-20.log` |

**결함 1의 `-count=50`은 50회 독립 시도가 아니다.** `go test -count=N`은 단일 프로세스 안에서 N회 반복하며 패키지 수준 상태가 반복 간 유지된다. `commandsAreSorted`는 전역 `*cobra.Command`의 1회성 플래그이므로 반복 1이 레이스 창을 닫고 반복 2~50은 창이 닫힌 상태로 실행된다 → **유효 1회 + 무효 49회**. 이 로그를 "50회 시도했으나 재현 안 됨"으로 읽어서는 안 된다.

결함 2의 `-count=20`에는 이 문제가 없다(매 반복이 새 임시 디렉터리 생성, 프로세스 전역 1회성 플래그 미의존).

`spec.md` §G가 이 공백의 정식 기록이다. 본 계획은 로컬 재현을 전제하지 않는다.

---

## §D. 제약

- 프로덕션 동작 보존: `LockTimeout` 2s 기본값, cobra 트리 구성, `ErrLockTimeout` 계약 불변.
- 테스트 약화 금지: 구문 후보 11건의 `t.Parallel()` 제거·테스트 삭제·skip 추가 금지(REQ-CFS-016 대상 1건 제외).
- 변경 파일 한정:
  - `internal/cli/main_test.go` (신규 — TestMain + `warmUpDone` + 재귀 warm-up + 가드 테스트)
  - `internal/cli/preference/main_test.go` (신규 — 동일 구조)
  - `internal/cli/inventory_guard_test.go` (신규 — REQ-CFS-006, §B.4)
  - `internal/session/registry.go` (수정 — `withLock` 루프)
  - `internal/session/registry_starvation_test.go` (신규 — 특성화 테스트)
- 브랜치 보호: `enforce_admins: true` → `main` 직접 푸시 불가, 모든 변경 PR 경유. `strict: true` → 머지 전 base 최신화 필요.
- 커밋: Conventional Commits, 영문. 코드·주석 영문.
- 후속 단계: run → sync → PR(auto-merge squash).

---

## §E. 자체 검증 (실행 시 수행)

| # | 항목 | 명령 |
|---|------|------|
| E1 | AC 매트릭스 PASS/FAIL | `acceptance.md` §D 전 항목 실행 |
| E2 | 결함 1 — **신규 프로세스** 50회 반복 (`-count=50` 금지) | `for i in $(seq 1 50); do go test -race -count=1 ./internal/cli/preference/ \|\| exit 1; done` 및 동일 루프의 `./internal/cli/` |
| E3 | 결함 2 반복 | `go test -race -count=10 ./internal/session/...` (여기서는 `-count` 사용 정당 — §C.3) |
| E4 | 전체 스위트 | `go test ./...` |
| E5 | 정적 분석 | `go vet ./...` && `golangci-lint run` |
| E6 | 범위 규율 | `git diff --stat origin/main...HEAD` — §D 열거 파일 외 변경 0 |
| E7 | 푸시 상태 | `git log origin/feat/SPEC-CI-FLAKY-STABILIZE-001..HEAD --oneline` |

**E2가 `-count=50`이 아닌 이유**: `-count=N`은 단일 프로세스 반복이며 `commandsAreSorted`는 1회성 전역 플래그다. `-count=50`은 진짜 레이스 창을 1회만 제공하므로 warm-up을 삭제해도 반복 2회차부터 통과한다 — 즉 판별력이 `-count=1`과 동일하다. 신규 프로세스 루프는 매 회 새 창을 만든다. `-count=1`은 테스트 캐시도 무력화한다.

---

## §F. 밀레스톤 (재현 우선 순서 — CLAUDE.md §7 Rule 4)

밀레스톤은 **실행 순서**이며, 되돌리기 어려운 설계 결정은 §B에 이미 앞세워 두었다.

### M1 — 결함 1 RED: 실제 전역 대상 레이스 재현 테스트

- **실제 패키지 전역**(`internal/cli/preference`는 `PreferenceCmd`, `internal/cli`는 `rootCmd`)에 대해 2개 이상 goroutine이 동시에 `Commands()`를 최초 호출하는 테스트를 작성한다. 테스트 로컬 트리를 새로 만들어서는 **안 된다** — 로컬 트리는 warm-up이 도달할 수 없어 M2에서 GREEN 확인이 불가능해진다.
- **실행 격리 (필수)**: RED 실행은 반드시 아래 형태로 한다.

  ```bash
  for i in $(seq 1 20); do go test -race -count=1 -run '^<M1TestName>$' ./internal/cli/preference/; done
  ```

  `-run '^<M1TestName>$'` 앵커와 **신규 프로세스 반복**이 "최초 `Commands()` 호출" 전제를 성립시킨다. 이 전제는 같은 바이너리 안에서 먼저 실행된 테스트가 해당 전역에 `Commands()`를 이미 호출하지 않았을 것을 요구하는데, `internal/cli`에는 그런 호출 지점이 9곳, `internal/cli/preference`에는 2곳 있다. 패키지 전체 실행(`go test ./internal/cli/...`)은 M1이 실행되기 전에 레이스 창을 닫아버릴 수 있고, 그러면 **메커니즘과 무관한 이유로** 비관측이 발생해 불필요하게 blocker로 라우팅된다. `-count=1`은 테스트 캐시도 무력화한다.

- **`TestMain` warm-up이 없는 상태에서 위 명령 실행 → 실패 확인 및 출력 기록.** 20회 반복에서도 관측되지 않으면 goroutine 수를 늘려 재시도하고, 그래도 실패하면 blocker 보고(추측 진행 금지).
- **증거 보존 의무**: 실패 출력 verbatim **및 테스트의 전체 소스**를 `progress.md` §E.2에 기록한다. 이후 M2에서 파일을 제거하더라도 감사자가 재구성·재실행할 수 있어야 한다(§G AP-3 참조).
- 산출: 실패 출력 + 테스트 전체 소스.

### M2 — 결함 1 GREEN: TestMain 재귀 warm-up 도입

- `internal/cli/main_test.go` 신규:
  - 패키지 스코프 `var warmUpDone bool`
  - `func warmUpCommandTree(c *cobra.Command)` — 함수 본문 첫 줄에서 `warmUpDone = true` 대입 + 깊이 우선 재귀(§B.3)
  - `func TestMain(m *testing.M)` — `warmUpCommandTree(rootCmd)` **호출만** 수행한 뒤 `os.Exit(m.Run())`. **`TestMain` 안에서 `warmUpDone`을 대입하지 않는다** — 플래그는 함수 내부에서만 설정된다(§B.2 대입 위치 규칙)
  - REQ-CFS-003 의도 주석(영문, cobra lazy-sort 메커니즘 + 삭제 시 재발 + 정렬 상태 관측 금지 사유 명시)
- `internal/cli/preference/main_test.go` 신규: 동일 구조, `PreferenceCmd` 기점.
- M1 재현 테스트를 warm-up 적용 상태에서 재실행 → 통과 확인. 통과 확인 후 테스트 파일 제거(소스는 progress.md에 보존됨).
- `githubCmd`·`hookCmd` 도달 확인: 재귀가 `rootCmd` 자식들을 순회하므로 도달한다. 필요 시 warm-up 직후 `githubCmd`/`hookCmd`가 방문되었음을 임시 로그로 1회 확인한다(최종 커밋에는 남기지 않음).

### M3 — 결함 1 가드 (REQ-CFS-004): 도달 가능성 신호

- 각 패키지에 가드 테스트 추가(직렬, `t.Parallel()` 없음): `warmUpDone == true` 어서션.
- **가드 자체 검증(왕복)**: `warmUpCommandTree(rootCmd)` 호출부를 일시 주석 처리 → 가드 FAIL 확인 → 원복 → 가드 PASS 확인. 두 출력 모두 기록한다. 초판의 정렬 기반 가드와 달리 이 왕복은 실제로 FAIL을 생성한다.

### M4 — 결함 1 인벤토리 가드 (REQ-CFS-006)

- §B.4 설계대로 `internal/cli/inventory_guard_test.go` 작성.
- 자체 검증: 임시로 warm-up 도달 집합에서 하나를 제외하거나, 전역 수신자를 쓰는 더미 병렬 테스트를 추가 → 가드 FAIL 확인 → 원복 → PASS 확인.

### M5 — 결함 1 검증

- E2의 신규 프로세스 50회 루프를 두 패키지에 대해 실행, 전부 통과.
- 구문 후보 11건의 `t.Parallel()` 잔존 확인(인벤토리 재도출 = 11).

### M6 — 결함 2 RED: 기아 특성화 테스트

- 경합 하 **획득당 대기 시간**을 계측하는 테스트 작성(p50/p95/max를 `t.Log` 출력, 어서션은 max 상한 1개만).
- `-short`에서 skip(REQ-CFS-016).
- 현행 고정 20ms 코드 상태에서 **기준선 수치를 측정하고 기록**한다.
- 정직성 요구: 이 테스트는 CI 실패를 결정론적으로 재현하지 않는다. `progress.md` 기록 시 "특성화이지 재현이 아님"을 명시한다.

### M7 — 결함 2 GREEN: jitter + 지수 백오프 적용

- `registry.go` `withLock` 루프의 `time.Sleep(20 * time.Millisecond)`를 §B.1 파라미터의 full jitter 지수 백오프로 교체.
- 데드라인 클램프: 남은 시간보다 긴 슬립 금지.
- `LockTimeout` 상수·`ErrLockTimeout` 반환 계약·`LOCK_NB` 획득 방식 불변 확인.

### M8 — 결함 2 검증

- M6 특성화 테스트 재실행 → 기준선 대비 수치 대조 기록.
- `go test -race -count=10 ./internal/session/...` 통과.
- `errors.Is(err, ErrLockTimeout)` 회귀 테스트 통과(타임아웃 경로 계약 불변).

### M9 — 품질 게이트 및 PR 준비

- `go test ./...`, `go vet ./...`, `golangci-lint run` 전부 green.
- 범위 규율 확인(§E6).
- Conventional Commits로 커밋 정리, base 최신화 후 PR 생성은 sync 단계에서 `manager-git`이 담당.

---

## §G. 안티패턴 (금지)

- **AP-1 — 직렬화로 레이스 회피**: `t.Parallel()` 제거는 증상 은폐이며 REQ-CFS-005 위반이다.
- **AP-2 — 소스 grep 가드**: warm-up 호출 문자열 존재만 확인하는 가드는 도달 가능성을 증명하지 못한다(§B.2 (a)).
- **AP-2b — 정렬 상태 가드**: `Commands()`가 접근자 내부에서 정렬하므로 정렬 상태는 항상 참이며, 이 가드는 구조적으로 FAIL할 수 없다(§B.2 (b)). 초판의 오류이므로 재도입 금지.
- **AP-3 — 재현 테스트 증거 소실**: M1 테스트를 최종 트리에서 제거하는 것 자체는 유지하되(상시 레이스 유발 테스트는 새 flaky 원인이 됨), **전체 소스를 `progress.md` §E.2에 보존하지 않고 제거하는 것**은 금지한다. 소스 없이 실패 출력만 남기면 AC-CFS-009는 산문 신뢰에만 의존하게 된다.
  - 참고(정직한 기록): M3 개정으로 M1 테스트가 **실제 전역**을 대상으로 하게 되면서, warm-up 적용 후에는 이 테스트가 레이스를 일으킬 수 없다. 따라서 "상시 유지 시 새 flaky 유발"이라는 AP-3 원래 근거는 약해졌다. 그럼에도 제거를 유지하는 이유는 (i) 항상 통과하는 테스트는 판별력이 없고, (ii) warm-up 삭제 시의 결정론적 방어는 M3 가드가 담당하기 때문이다.
- **AP-4 — 타임아웃 상향으로 무마**: `WithLockTimeout(60s)`를 더 늘리는 것은 기아를 감출 뿐 해결하지 않는다.
- **AP-5 — 백오프 cap을 lockTimeout 이상으로 설정**: 단일 슬립이 데드라인을 넘겨 재시도 기회를 잃는다(REQ-CFS-011).
- **AP-6 — 미재현을 재현으로 보고**: 로컬 통과를 "결함 없음"으로 서술하는 것은 금지. 특히 결함 1의 `-count=50`을 "50회 시도"로 서술하는 것은 사실과 다르다(§C.3). `spec.md` §G가 정본이다.
- **AP-7 — 인접 리팩터링 편승**: 두 결함과 무관한 정리 작업 금지(REQ-CFS-021).
- **AP-8 — `-count=N`으로 결함 1 회귀 검증**: 단일 프로세스 반복은 1회성 전역 플래그 때문에 판별력이 `-count=1`과 같다. 신규 프로세스 루프를 쓸 것(§E2).

---

## §H. 교차 참조

- `spec.md` — 요구사항 SSOT (§B.1 cobra 소스 인용, §C.1 수신자 분류표, §D GEARS, §G Gaps, §H Residual Risk)
- `acceptance.md` — AC 매트릭스 및 검증 명령
- `progress.md` — 단계별 증거 기록 (M1 테스트 전체 소스 포함)
- `CLAUDE.md` §7 Rule 4 — 재현 우선 버그 수정
- `.claude/rules/moai/core/verification-claim-integrity.md` §3.2(Evidence = 명령 + verbatim 출력) / §3.4(Gaps 명시 의무)
- 증거 디렉터리: `.moai/state/verify/b5fc437f/`
- cobra v1.10.2 `command.go:1332-1339`, `cobra.go:59` — lazy-sort 접근자 및 `EnableCommandSorting` 기본값
