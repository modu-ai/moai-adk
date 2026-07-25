---
id: SPEC-CI-FLAKY-STABILIZE-001
title: "CI Flaky 안정화 — 인수 기준"
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

## §A. 검증 원칙

모든 AC는 **실행 가능한 명령 + 그 출력이 보여야 하는 것**으로 기술한다. `grep -c`만으로 충족되는 AC(문자열 존재만 확인, 도달 가능성 미증명)는 이 SPEC에서 유효하지 않다. 문자열 검사를 쓰는 AC는 반드시 동작 관측 AC와 짝지어 둔다.

추가 원칙(0.2.0): AC는 **반증 가능**해야 한다. 어떤 구현 상태에서도 반드시 통과하는 AC는 무효다. 초판의 정렬 상태 기반 AC가 이 원칙을 위반해 교체되었다(§D.3).

추가 원칙(0.3.0): 왕복 검증(round-trip)을 요구하는 AC는 **FAIL 측 경로가 실제로 도달 가능한지** 함께 검증해야 한다. "제거하면 FAIL한다"는 서술만으로는 부족하며, 제거 대상과 관측 신호가 **같은 코드 경로**에 있어야 한다. AC-CFS-004b가 이 선행 조건을 기계적으로 확인한다.

모든 명령은 워크트리 루트에서 실행한다:
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/ci-flaky/`

---

## §B. Given-When-Then 시나리오

### GWT-1 — cobra 지연 정렬이 병렬 테스트 시작 전에 완료된다

- **Given** `internal/cli/preference` 패키지의 두 테스트가 `t.Parallel()` 상태로 남아 있고, 둘 다 `PreferenceCmd.Commands()`를 호출한다.
- **When** `go test -race ./internal/cli/preference/`를 실행한다.
- **Then** `TestMain`이 `m.Run()` 이전에 재귀 warm-up을 수행했으므로 `commandsAreSorted`는 이미 `true`이며, 테스트 중에는 읽기만 발생하여 race detector가 발화하지 않는다. 종료 코드 0, 출력에 `WARNING: DATA RACE` 부재.

### GWT-2 — warm-up이 제거되면 가드 테스트가 결정론적으로 실패한다

- **Given** 각 패키지의 `main_test.go`에 `warmUpDone` 신호와 이를 확인하는 가드 테스트가 존재하며, `warmUpDone = true` 대입이 **`warmUpCommandTree` 함수 본문 안**에 있다(`TestMain`의 별도 문장이 아님).
- **When** `TestMain`의 `warmUpCommandTree(...)` 호출 **한 줄만** 주석 처리하고 `go test -run '<가드테스트명>' ./internal/cli/preference/`를 실행한다.
- **Then** 대입문이 함수 본문 안에 있어 호출 제거와 함께 도달 불가가 되므로 `warmUpDone`이 `false`로 남고, 가드가 **반드시** FAIL한다(확률적이지 않음). 주석을 원복하면 PASS로 돌아온다.
- **전제 조건 주의**: 대입이 `TestMain` 안에 별도 문장으로 있으면 그 줄이 계속 실행되어 가드가 통과하고 이 시나리오는 성립하지 않는다 — `plan.md` §B.2 대입 위치 규칙이 이를 금지한다.

### GWT-3 — warm-up 재귀가 `githubCmd`에 도달한다

- **Given** `githubCmd`는 `internal/cli/github.go:89`의 패키지 전역이며 `rootCmd.AddCommand(githubCmd)`(`github.go:100`)로 트리에 연결되어 있고, `TestGithubCLI_ErrorPropagation`이 `t.Parallel()` + `githubCmd.Commands()`를 수행한다.
- **When** `internal/cli`의 `TestMain`이 `warmUpCommandTree(rootCmd)`를 호출한다.
- **Then** 재귀가 `rootCmd`의 자식을 순회하며 `githubCmd`를 방문하고 그 `Commands()`를 1회 호출하므로, `githubCmd`의 지연 정렬이 테스트 시작 전에 완료된다.

### GWT-4 — 경합 하 최대 단일 획득 대기 시간이 개선된다

- **Given** `withLock` 재시도가 고정 20ms 슬립인 상태(기준선).
- **When** 기아 특성화 테스트를 실행하여 p50/p95/max 획득 대기 시간을 기록하고, jitter+backoff를 적용한 뒤 동일 테스트를 재실행한다.
- **Then** 두 수치 집합이 `progress.md`에 나란히 기록되며, max 값이 기준선 이하다.

### GWT-5 — 타임아웃 오류 계약이 불변이다

- **Given** 획득 불가능한 락(외부에서 점유)과 짧은 `lockTimeout`.
- **When** `withLock`을 호출한다.
- **Then** 반환 오류에 대해 `errors.Is(err, ErrLockTimeout)`이 `true`다 — 변경 전후 동일.

### GWT-6 — 프로덕션 기본 타임아웃이 변경되지 않는다

- **Given** 본 SPEC의 모든 변경이 적용된 트리.
- **When** `LockTimeout` 상수 선언을 확인한다.
- **Then** `2 * time.Second`이며, `go test ./internal/session/...`의 기존 타임아웃 관련 테스트가 모두 통과한다.

---

## §C. 엣지 케이스

| # | 케이스 | 기대 동작 |
|---|--------|-----------|
| EC-1 | 백오프 슬립이 남은 데드라인보다 김 | 남은 시간으로 클램프. 데드라인 초과 시 `ErrLockTimeout` 반환 (무한 대기 금지) |
| EC-2 | `lockTimeout`이 0 이하 | 기존 동작 유지 — `LockTimeout` 기본값(2s)으로 대체 |
| EC-3 | 첫 시도에 락 획득 성공 | 슬립 0회. 백오프 도입으로 인한 추가 지연 없음 |
| EC-4 | `-short` 모드 | 기아 특성화 테스트 skip, `TestRegisterSessionConcurrent` skip (기존 관례 유지) |
| EC-5 | Windows 빌드 | `registry_lock_unix.go`는 unix 빌드 태그. `withLock`(플랫폼 공통) 변경이 Windows 빌드를 깨지 않아야 함 |
| EC-6 | 커맨드 트리에 자식이 0개인 노드 | 재귀 warm-up이 패닉 없이 종료(빈 슬라이스 순회) |
| EC-7 | 커맨드 트리에 순환 참조 | cobra는 자기 자신을 자식으로 추가하면 패닉하므로(`AddCommand`) 순환은 구조적으로 불가. 재귀 깊이 방어 불필요 |
| EC-8 | 인벤토리 가드가 수신자를 정적 해석할 수 없음 | 보수적으로 "로컬 수신자"로 분류하고 통과(false-negative 허용). 한계는 `spec.md` §H 잔여 위험 3에 기록됨 |

---

## §D. AC 매트릭스

### D.1 결함 1 — cobra lazy-sort 레이스 (AC-CFS-001 ~ AC-CFS-010)

| AC ID | 요구사항 | 검증 명령 | 기대 관측 | 심각도 |
|-------|----------|-----------|-----------|--------|
| AC-CFS-001 | REQ-CFS-001 | `grep -c "func TestMain" internal/cli/main_test.go` | `1`. 단독으로는 불충분하며 AC-CFS-003과 함께 판정 | MUST |
| AC-CFS-002 | REQ-CFS-001 | `grep -c "func TestMain" internal/cli/preference/main_test.go` | `1` | MUST |
| AC-CFS-003 | REQ-CFS-002 | `go test -run '<가드테스트명>' -v ./internal/cli/ ./internal/cli/preference/` | 두 패키지 모두 `--- PASS`. 가드는 `warmUpDone == true`(도달 가능성 신호)를 어서션 — **정렬 상태를 어서션해서는 안 됨** | MUST |
| AC-CFS-004 | REQ-CFS-003 | `grep -B2 -A8 "warmUpDone" internal/cli/main_test.go internal/cli/preference/main_test.go` | 출력에 `commandsAreSorted`, `t.Parallel`, `sorts in place` 취지의 영문 설명 포함. 삭제 시 재발 + 정렬 상태 관측 금지 사유 명시 | MUST |
| AC-CFS-004b | REQ-CFS-004 | `sed -n '/func warmUpCommandTree/,/^}/p' internal/cli/main_test.go internal/cli/preference/main_test.go` 및 `sed -n '/func TestMain/,/^}/p' internal/cli/main_test.go internal/cli/preference/main_test.go` | 전자 출력에 `warmUpDone = true` **존재**, 후자 출력에 `warmUpDone` 대입 **부재**. 대입이 `TestMain`에 있으면 AC-CFS-005의 왕복이 FAIL을 만들 수 없으므로 즉시 blocker | MUST |
| AC-CFS-005 | REQ-CFS-004 | (a) `warmUpCommandTree(...)` 호출 **한 줄만** 주석 처리 → `go test -run '<가드테스트명>' ./internal/cli/preference/`; (b) 원복 후 재실행 | (a) `FAIL` — 대입이 함수 본문 안에 있어 호출과 함께 도달 불가가 되므로 `warmUpDone`이 `false`, **결정론적** 실패; (b) `ok`. 두 출력 모두 `progress.md`에 기록. (a)에서 FAIL이 나오지 않으면 가드가 무효이므로 blocker 보고 — 가장 흔한 원인은 대입이 `TestMain`에 있는 것이며 AC-CFS-004b가 이를 선행 차단한다 | MUST |
| AC-CFS-006 | REQ-CFS-005 | §D.5의 AWK 인벤토리 명령 재실행 | 정확히 11행 출력, `spec.md` §C.1 표의 구문 후보 집합과 동일 | MUST |
| AC-CFS-007 | REQ-CFS-002 | `for i in $(seq 1 50); do go test -race -count=1 ./internal/cli/preference/ \|\| exit 1; done; echo "exit=$?"` | `exit=0`, 어떤 반복에서도 `WARNING: DATA RACE` 부재. **`-count=50` 단일 프로세스 실행으로 대체 금지**(§F 6항) | MUST |
| AC-CFS-008 | REQ-CFS-002 | `for i in $(seq 1 50); do go test -race -count=1 ./internal/cli/ \|\| exit 1; done; echo "exit=$?"` | `exit=0`, `WARNING: DATA RACE` 부재 | MUST |
| AC-CFS-009 | REQ-CFS-020 | M1 재현 테스트(실제 전역 `PreferenceCmd`/`rootCmd` 대상)를 warm-up **미적용** 상태에서 실행: `for i in $(seq 1 20); do go test -race -count=1 -run '^<M1TestName>$' ./internal/cli/preference/; done` | `WARNING: DATA RACE` 관측. 해당 출력 verbatim **및 테스트 전체 소스**가 `progress.md` §E.2에 존재 — 소스 보존은 AC-CFS-010의 파일 제거로 증거가 소실되지 않게 하는 필수 조건. **`-run` 앵커 + 신규 프로세스 반복이 필수** — 패키지 전체 실행은 선행 테스트가 레이스 창을 먼저 닫아 메커니즘과 무관한 비관측을 만든다(`internal/cli` 9곳, `internal/cli/preference` 2곳의 선행 `Commands()` 호출 지점). 관측 실패 시 blocker 보고(추측 금지) | MUST |
| AC-CFS-010 | plan §G AP-3 | `git diff --stat origin/main...HEAD` | M1 재현 테스트 파일이 최종 diff에 **부재**. 단 AC-CFS-009의 소스 보존이 선행 충족되어 있어야 함 | SHOULD |

### D.2 결함 2 — flock 기아 (AC-CFS-011 ~ AC-CFS-019)

결함 2 관련 AC는 0.1.0에서 변경되지 않았다.

| AC ID | 요구사항 | 검증 명령 | 기대 관측 | 심각도 |
|-------|----------|-----------|-----------|--------|
| AC-CFS-011 | REQ-CFS-010 | `grep -n "time.Sleep(20 \* time.Millisecond)" internal/session/registry.go` 및 `sed -n '/func (r \*Registry) withLock/,/^}/p' internal/session/registry.go` | 전자 출력 없음(고정 슬립 소멸). 후자 본문에 난수 지터 + 지수 증가 로직 존재 | MUST |
| AC-CFS-012 | REQ-CFS-012 | `grep -n "LockTimeout = " internal/session/registry.go` | `LockTimeout = 2 * time.Second` — 변경 없음 | MUST |
| AC-CFS-013 | REQ-CFS-013 | `grep -n "unix.LOCK_EX" internal/session/registry_lock_unix.go` | `unix.LOCK_EX\|unix.LOCK_NB` 유지 | MUST |
| AC-CFS-014 | REQ-CFS-014 | `go test -run 'LockTimeout\|Timeout' -v ./internal/session/` | 타임아웃 경로 테스트 PASS. 테스트가 `errors.Is(err, ErrLockTimeout)`을 어서션 | MUST |
| AC-CFS-015 | REQ-CFS-015 | `go test -run '<특성화테스트명>' -v ./internal/session/` | PASS + 로그에 p50/p95/max 수치 출력 | MUST |
| AC-CFS-016 | REQ-CFS-016 | `go test -short -run '<특성화테스트명>' -v ./internal/session/` | `--- SKIP` | MUST |
| AC-CFS-017 | REQ-CFS-020 | 수정 **전** 상태에서 특성화 테스트 실행 | 기준선 p50/p95/max 수치가 `progress.md` §E.2에 기록 | MUST |
| AC-CFS-018 | REQ-CFS-011 | 코드 리뷰 + `go test -run '<타임아웃테스트명>' ./internal/session/` | 슬립이 남은 데드라인으로 클램프됨을 소스에서 확인, 타임아웃 테스트가 예산 내에 종료 | MUST |
| AC-CFS-019 | REQ-CFS-010 | `go test -race -count=10 ./internal/session/` | 종료 코드 0, `ErrLockTimeout` 미발생. (여기서 `-count=10`은 유효 — 매 반복이 새 임시 디렉터리를 만들고 프로세스 1회성 전역 플래그에 의존하지 않음) | MUST |

### D.3 초판 정렬 상태 AC 폐기 기록 (0.2.0)

0.1.0의 AC-CFS-003 / AC-CFS-005는 `Commands()` 반환 슬라이스의 정렬 상태를 관측하도록 규정했다. 이는 **구조적으로 반증 불가**하여 무효였다:

cobra v1.10.2 `command.go:1332-1339`는 `Commands()` 접근자 **내부에서** `sort.Sort(...)`를 수행하고 `commandsAreSorted = true`를 설정하며, `EnableCommandSorting`은 기본 `true`(`cobra.go:59`)다. 따라서 가드가 `Commands()`를 호출하는 순간 정렬이 발생하므로, warm-up 실행 여부와 무관하게 반환 슬라이스는 **항상 정렬 상태**다. 초판이 기록한 한계("등록 순서가 우연히 사전순인 경우 통과")는 과소평가였다 — 실제로는 모든 트리 형태에서 항상 통과하며, 초판 AC-CFS-005의 왕복 검증(주석 처리 → FAIL 기대)은 **결코 FAIL을 생성할 수 없었다**.

0.2.0은 관측 대상을 **도달 가능성 신호**(`warmUpDone` 테스트 스코프 변수)로 교체했다. 이 신호는 warm-up 코드 경로의 실행 여부를 직접 기록하므로 warm-up 제거 시 결정론적으로 FAIL한다. 초판이 대안으로 언급했던 "등록 순서 셔플 어서션"도 동일 이유로 무효이므로 채택하지 않는다(어떤 순서로 등록하든 `Commands()` 호출 후에는 정렬된다).

`warmUpDone`은 `main_test.go`(테스트 전용 파일)의 패키지 스코프 변수이므로 프로덕션 코드에 테스트 전용 상태를 추가하지 않는다 — 초판이 이 방식을 기각한 사유는 본 케이스에 적용되지 않았다.

### D.4 REQ-CFS-006 인벤토리 가드 (AC-CFS-026)

| AC ID | 요구사항 | 검증 명령 | 기대 관측 | 심각도 |
|-------|----------|-----------|-----------|--------|
| AC-CFS-026 | REQ-CFS-006 | (a) `go test -run '<인벤토리가드명>' -v ./internal/cli/` → PASS 확인; (b) **warm-up 도달 집합 계산에서 항목 하나를 일시 제외**(예: 전이적 폐포 계산에서 `githubCmd` 간선을 건너뛰도록 임시 수정) → 가드 FAIL 확인; (c) 원복 → PASS 확인 | (a) `--- PASS`; (b) `--- FAIL` + 위반 항목의 파일·함수·수신자 이름이 실패 메시지에 표시; (c) `--- PASS`. (b)에서 FAIL이 나오지 않으면 가드가 무효이므로 blocker 보고 | MUST |

가드는 `internal/cli` 및 `internal/cli/preference` 디렉터리의 `_test.go`를 `go/ast`로 파싱하고, `t.Parallel()`과 `.Commands()`가 같은 함수 본문에 공존하는 항목을 수집한 뒤 각 `.Commands()` 수신자를 해석한다. 수신자가 비-테스트 파일의 패키지 레벨 `var`이면서 warm-up 도달 집합에 없으면 실패한다. 도달 집합은 비-테스트 파일의 `<parent>.AddCommand(<ident>)` 스캔 + 루트로부터의 전이적 폐포로 **정적 도출**되며 리터럴 목록이 아니다. 설계·비용은 `plan.md` §B.4 참조.

**왕복 검증 경로 선택 근거**: 초판이 대안으로 제시했던 "warm-up 도달 집합 밖의 전역 수신자를 쓰는 임시 병렬 테스트 추가"는 **구성 불가능**하다. `internal/cli`의 패키지 레벨 cobra 전역은 모두 `rootCmd.AddCommand(...)`로 등록되어 있어(`githubCmd`, `hookCmd` 포함) 도달 집합 밖 전역이 존재하지 않으며, 그런 전역을 새로 만들려면 프로덕션 파일을 건드려야 해 REQ-CFS-021(범위 규율)에 저촉된다. 따라서 도달 집합 계산 측을 일시 수정하는 경로만 채택한다 — 반증 가능성은 동일하게 확보된다.

### D.5 공통 품질 게이트 (AC-CFS-020 ~ AC-CFS-025)

| AC ID | 요구사항 | 검증 명령 | 기대 관측 | 심각도 |
|-------|----------|-----------|-----------|--------|
| AC-CFS-020 | REQ-CFS-022 | `go test ./...` | 종료 코드 0 | MUST |
| AC-CFS-021 | REQ-CFS-022 | `go vet ./...` | 종료 코드 0, 출력 없음 | MUST |
| AC-CFS-022 | REQ-CFS-022 | `golangci-lint run` | 종료 코드 0 | MUST |
| AC-CFS-023 | REQ-CFS-021 | `git diff --name-only origin/main...HEAD` | `internal/cli/main_test.go`, `internal/cli/preference/main_test.go`, `internal/cli/inventory_guard_test.go`, `internal/session/registry.go`, session 특성화 테스트 파일, `.moai/specs/SPEC-CI-FLAKY-STABILIZE-001/*` 외 **0건** | MUST |
| AC-CFS-024 | REQ-CFS-005 | `git diff origin/main...HEAD -- '*_test.go' \| grep -E '^-.*(func Test\|t\.Parallel\(\)\|t\.Skip)'` | 출력 없음(테스트 함수·병렬 선언 삭제 0건). `t.Skip` 신규 **추가**는 REQ-CFS-016 대상 1건만 허용 | MUST |
| AC-CFS-025 | REQ-CFS-023 | `git log origin/main..HEAD --format='%s'` | 전 커밋이 Conventional Commits 형식, 영문 | SHOULD |

인벤토리 재도출 명령(AC-CFS-006용 — **구문 필터**이며 결함 클래스 필터가 아님):

```bash
awk '
/^func Test/ { name=$2; sub(/\(.*/,"",name); body=""; inf=1 }
inf { body = body $0 }
/^}$/ { if (inf && body ~ /t\.Parallel\(\)/ && body ~ /\.Commands\(\)/) print FILENAME": "name; inf=0 }
' internal/cli/*_test.go internal/cli/preference/*_test.go | sort | wc -l
# 기대: 11 (구문 후보). 이 중 전역 수신자 7건만 결함 클래스 — spec.md §C.1 참조.
```

---

## §E. 추적성 매트릭스

| 요구사항 | AC |
|----------|-----|
| REQ-CFS-001 | AC-CFS-001, AC-CFS-002 |
| REQ-CFS-002 | AC-CFS-003, AC-CFS-007, AC-CFS-008 |
| REQ-CFS-003 | AC-CFS-004 |
| REQ-CFS-004 | AC-CFS-004b, AC-CFS-005 |
| REQ-CFS-005 | AC-CFS-006, AC-CFS-024 |
| REQ-CFS-006 | AC-CFS-026 |
| REQ-CFS-010 | AC-CFS-011, AC-CFS-019 |
| REQ-CFS-011 | AC-CFS-018 |
| REQ-CFS-012 | AC-CFS-012 |
| REQ-CFS-013 | AC-CFS-013 |
| REQ-CFS-014 | AC-CFS-014 |
| REQ-CFS-015 | AC-CFS-015 |
| REQ-CFS-016 | AC-CFS-016 |
| REQ-CFS-020 | AC-CFS-009, AC-CFS-017 |
| REQ-CFS-021 | AC-CFS-023 |
| REQ-CFS-022 | AC-CFS-020, AC-CFS-021, AC-CFS-022 |
| REQ-CFS-023 | AC-CFS-025 |

미매핑 요구사항: 없음. 미매핑 AC: AC-CFS-010(계획 §G AP-3 근거, 요구사항이 아닌 안티패턴 방어).

AC ID 배번 주의: `AC-CFS-026`은 §D.4에 있으며 `AC-CFS-020`~`025`(§D.5)보다 뒤 번호다. 초판(0.1.0)에 없던 신규 AC를 기존 번호를 밀지 않고 뒤에 붙였기 때문이며, 의도적이다.

---

## §F. 검증되지 않는 것 (Gaps — AC로 커버 불가)

다음 항목은 **어떤 AC로도 증명되지 않는다.** `spec.md` §G의 요약이며, PASS 판정 시 함께 읽어야 한다.

1. **원래 CI 실패의 로컬 재현.** 결함 1은 로컬 통과, 결함 2는 20/20 통과로 재현되지 않았다. 모든 AC가 PASS해도 "원래 실패가 고쳐졌다"는 직접 증거는 존재하지 않는다.
2. **AC-CFS-009가 검증하는 것.** 0.2.0에서 대상이 실제 전역(`PreferenceCmd`/`rootCmd`)으로 교체되어 초판의 "인위적 트리 유사물" 공백은 해소되었다. 남은 한계는 재현이 **확률적**이라는 점 — goroutine 수·러너 특성에 따라 관측되지 않을 수 있으며, 그 경우 blocker로 보고된다.
3. **AC-CFS-015/017이 검증하는 것.** 대기 시간 분포의 특성화이지 기아의 결정론적 재현이 아니다. 로컬에서 기아가 발생하지 않으면 개선은 꼬리 지연 감소로만 나타난다.
4. **결함 1의 미실패 5건**(전역 수신자이면서 CI 실패 이력 없음). 수정 후에도 "여전히 실패하지 않음"만 관측되며, 수정 효과인지 원래 운인지 구분 불가하다.
5. **CI 미재발.** 유한 횟수 관측이며 간헐 결함의 부재 증거로 약하다.
6. **결함 1의 로컬 미재현 로그는 유효 시도 1회다.** `.moai/state/verify/b5fc437f/pref-race-50.log`의 `-count=50`은 **단일 프로세스** 50회 반복이며, `commandsAreSorted`가 1회성 전역 플래그이므로 반복 1이 레이스 창을 닫고 반복 2~50은 창이 닫힌 상태로 실행된다. 즉 유효 시도 1회 + 무효 반복 49회이며, "50회 시도해도 재현 안 됨"으로 읽을 수 없다. 이 사실은 AC-CFS-007/008이 신규 프로세스 루프를 요구하는 근거이기도 하다. (결함 2의 `-count=20`은 이 문제가 없어 20회가 실질 독립이다.)
7. **증거 로그가 자기기술적이지 않다.** `pref-race-50.log` / `session-flock-20.log`는 `ok … <duration>` + `exit=0`만 담고 호출 명령줄을 포함하지 않는다(`verification-claim-integrity.md` §3.2는 Evidence를 명령 + verbatim 출력으로 규정). 명령 문자열은 SPEC 본문이 보증하는 것이지 로그 자체가 보증하지 않는다. 본 계획은 이 로그를 재생성하지 않는다.
8. **AC-CFS-026 인벤토리 가드의 정적 해석 한계.** 수신자가 지역 변수·함수 반환값 즉시 호출인 경우 해석 불가로 통과시킨다(EC-8). 흔한 재발 경로만 차단하며 완전 검증이 아니다.

---

## §G. 종료 게이트 (Definition of Done)

전부 충족해야 sync 단계로 진입한다.

- [ ] §D의 MUST AC 25건 전부 PASS, 각 항목에 실제 명령 출력 인용
- [ ] SHOULD AC 2건(AC-CFS-010, AC-CFS-025) PASS 또는 미충족 사유 기록
- [ ] `progress.md` §E.2에 M1 실패 출력 **및 M1 테스트 전체 소스**, M6 기준선 수치, M8 개선 후 수치가 verbatim으로 기록
- [ ] AC-CFS-004b 대입 위치 확인 통과 — `warmUpDone` 대입이 `warmUpCommandTree` 본문에 존재하고 `TestMain`에 부재
- [ ] AC-CFS-005 가드 왕복 검증 결과(FAIL → PASS) 기록 — FAIL이 실제로 관측되었을 것
- [ ] AC-CFS-026 인벤토리 가드 왕복 검증 결과(PASS → FAIL → PASS) 기록
- [ ] §F Gaps 8항목이 완료 보고에 그대로 포함(축약·삭제 금지)
- [ ] `spec.md` §G 검증 공백 8항목 + §H 잔여 위험 6항목이 완료 보고에 포함
- [ ] 브랜치가 `origin/main` 대비 최신화(`strict: true` 대응)
- [ ] PR 생성(자기 머지 허용, `enforce_admins: true`로 직접 푸시 불가)

---

## §H. 향후 점검 (Forward-Looking Checks)

머지 이후 별도 트랙에서 관찰할 항목(본 SPEC의 종료 조건 아님):

- 결함 1: 이후 CI 실행에서 `internal/cli` 계열 `DATA RACE` 재발 여부. 재발 시 AC-CFS-026 가드가 놓친 수신자 형태(지역 변수 경유 등)인지 확인.
- 결함 2: `ErrLockTimeout` 재발 여부. 재발 시 §C.3의 out-of-scope였던 차단 flock 전환을 후속 SPEC으로 승격 검토.
- 테스트 로컬 수신자 4건이 향후 전역 수신자로 리팩터링되면 결함 클래스에 편입된다 — AC-CFS-026 가드가 이를 자동 탐지해야 하며, 탐지하지 못하면 가드의 해석 범위를 확장한다.
