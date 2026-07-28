# Acceptance Criteria — SPEC-V3R6-CI-FLAKY-STABILIZE-003

> Given-When-Then 시나리오 + 반증 가능 구조적 AC. 본 SPEC 의 결함은 로컬에서 재현되지 않으므로(`registry_starvation_test.go:8`), AC 는 **구조적 불변(invariant)** 기반이다. CI-only 재현 가능성에 의존하지 않는다.

---

## §D. AC Matrix

### AC-CFS3-001 — in-process mutex 가 임계 구역을 직렬화한다 (REQ-CFS3-001)

**Given** 하나의 `Registry` 인스턴스(또는 동일 경로를 가리키는 여러 `Registry` 인스턴스)가 `t.TempDir()` 의 단일 레지스트리 파일에 바인딩되어 있고,
**When** 10개의 goroutine 이 각각 100회(`workers=10 × perWorker=100`, 총 1000회) `Registry.Register` 를 동시에 호출하면,
**Then** 모든 1000회 호출이 오류 없이 완료되고 최종 엔트리 수는 정확히 1000이며, 이것은 in-process mutex 가 각 호출의 임계 구역을 상호 배타적으로 직렬화했기 때문이다(경합에 의한 `ErrLockTimeout` 발생 0건).

**Evidence:** `go test -race -run TestRegisterSessionConcurrent ./internal/session/` 의 종료 코드 0 + 테스트 본문의 엔트리 수 비교 통과.

### AC-CFS3-002 — 반증 가능 구조적 AC: 뮤텍스 제거 시 실패 (REQ-CFS3-002) ★ 핵심

**Given** 동일 `Registry` 의 `withLock` 임계 구역 진입/이탈을 추적하는 `atomic.Int64` probe(임계 구역 진입 시 increment, 이탈 시 decrement, "현재 임계 구역 동시 진입 goroutine 수" 를 관측)가 테스트 전용으로 노출되어 있고, **probe 의 inc/dec 는 `Registry.withLock` 본문 전체를 bracket 해야 한다** (in-process mutex 획득 직후 — 즉 동일 프로세스 lock 을 잡은 직후 `withLock` 본문 최상단 — 에 inc, 함수 exit 에서 dec). 이 배치가 probe 로 하여금 "현재 `withLock` 본문 내부에 있는 goroutine 수" 를 측정하게 한다. **주의(load-bearing):** probe 가 flock 보유 구간만 bracket 하면, OS-level flock 이 post-acquire 임계 구역을 단일 보유자로 직렬화하므로 뮤텍스 존재 여부와 무관하게 max ≤ 1 만 관측한다 — 이는 양방향 반증(적용 시 ≤ 1 / 제거 시 ≥ 2) 이 불가능한 vacuous AC 가 되어 `verification-claim-integrity.md` §1.1 surface 3 에 위반된다.
**When** 10 goroutine × 1000 iteration(총 10000회) 경합 하에서 probe 의 "최대 동시 진입 수" 를 관측하면,
**Then** in-process mutex 가 적용된 상태에서는 최대 동시 진입 수 ≤ 1 이고, **뮤텍스 획득문을 제거/무력화한 상태에서는 충분한 iteration(총 10000회 권장) 하에서 최대 동시 진입 수 ≥ 2 가 관측된다** (확률적 수렴; 관측되지 않을 경우 iteration 수를 늘려 재실행). 이 조건이 충족되지 않으면 AC 는 반증 불가능한 dead prose 이다.

**Evidence:**
- 적용 상태: 신규 테스트 `TestRegistryWithLockInProcessSerialization` 종료 코드 0 (probe max ≤ 1).
- 반증 상태(로컬 검증): 동일 테스트의 falsifiability 서브-케이스(M1) 실행 시 충분한 iteration 하에서 probe max ≥ 2 가 관측된다(확률적 수렴; 미관측 시 iteration 증가 재실행). 이 서브-케이스는 CI 기본 skip, env flag 로만 실행(`MOAI_CFS3_FALSIFIABILITY=1`).
- run-phase 는 양방향(적용/제거) 관측 출력을 `progress.md` §E.2 에 인용.

**★ falsifiability 요구:** 이 AC 의 유효성 조건은 "뮤텍스 제거 시 실패" 이다. run-phase 가 이 조건을 관측하지 못한 채 green 만 보고하면 AC 는 미충족이다(`verification-claim-integrity.md` §1.1 surface 3). 반증 관측은 충분한 iteration 하의 확률적 수렴이므로 미관측 시 iteration 수를 늘려 재실행한다.

### AC-CFS3-003 — cross-process timeout 경로 보존 (REQ-CFS3-003, REQ-CFS3-004)

**Given** 별도의 `registryLock` 인스턴스(`newRegistryLock()`, `Registry.withLock` 의 in-process mutex 를 경유하지 않음)가 레지스트리 파일의 OS flock 을 획득하여 보유하고 있고,
**When** `Registry.WithLockTimeout(80ms).Register(...)` 가 이와 경합하면,
**Then** 기존과 동일하게 `ErrLockTimeout` 으로 래핑된 오류가 반환되며 `errors.Is(err, ErrLockTimeout) == true` 이다. in-process mutex 가 cross-process timeout 경로를 단축하지 않는다.

**Evidence:** `TestWithLockTimeoutContract`(`registry_starvation_test.go:107`) 가 수정·삭제 없이 green. 테스트 종료 코드 0.

### AC-CFS3-004 — 지터 백오프 계약 보존 (REQ-CFS3-005)

**Given** `lockRetryDelay` 가 cross-process 경합 하에서 호출되면,
**When** attempt ∈ [0, 20] 범위로 호출 sampled,
**Then** 각 지연 값은 (0, `lockBackoffCap`] 범위에 속하고 remaining budget 을 초과하지 않는다. full jitter 가 실제로 작동한다.

**Evidence:** `TestLockRetryDelayBoundsAndJitter`(`registry_starvation_test.go:137`) green 유지.

### AC-CFS3-005 — 기아 특성화 테스트 보존 (REQ-CFS3-009)

**Given** 8 goroutine × 25 registration = 200회 경합 시,
**When** in-process mutex 가 적용된 상태에서 실행하면,
**Then** 최대 단일 획득 대기 시간이 `maxWaitCeiling = 15s` 이하이다. (구조적 뮤텍스 적용 후 이 값은 사실상 스케줄링 오버헤드 수준으로 수렴해야 한다.)

**Evidence:** `TestRegisterStarvationCharacterization`(`registry_starvation_test.go:32`) green 유지 + `t.Logf` 의 p50/p95/max 수치가 이전 관측 대비 개선 또는 동등.

### AC-CFS3-006a — Windows run-phase build parity (REQ-CFS3-006)

**Given** `registry_lock_windows.go` 가 존재하고,
**When** in-process mutex 설계가 적용되면,
**Then** Windows 빌드에서도 동일한 직렬화 의미론이 코드 수준에서 적용된다(in-process mutex 는 `Registry.withLock` 공통 경로에 있으므로 platform-conditional 코드가 아니다). run-phase 증빙은 **빌드 수준**이다.

**Evidence:**
- `GOOS=windows go vet ./internal/session/...` green.
- `GOOS=windows go build ./internal/session/...` green.

**Note:** Windows 런타임 결정론성(green `go test` 실행)은 run-phase 에서는 관측 불가하며 sync-phase CI 증빙으로 연기된다 — see AC-CFS3-006b.

### AC-CFS3-006b — Windows sync-phase CI runtime green (REQ-CFS3-006)

**Given** AC-CFS3-006a 의 빌드 호환성이 확보되어 있고,
**When** 본 SPEC 변경이 머지되면,
**Then** Windows CI 잡의 `go test ./internal/session/...` 가 green 이다. Windows 런타임 결정론성은 sync-phase CI 증빙으로 관측된다(run-phase 에서는 Windows 런타임 실검이 불가하므로 SHOULD).

**Evidence:** sync-phase Windows CI 잡 green 증빙.

### AC-CFS3-007 — 비차단 flock 유지 (REQ-CFS3-007, REQ-CFS3-008)

**Given** 본 SPEC 의 변경이 적용되어도,
**When** `registry_lock_unix.go` 의 `acquire` 호출을 정적 검사하면,
**Then** `unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)` 호출이 그대로 존재하며(차단 flock 전환 금지), `lockRetryDelay` 지터 백오프 루프가 `withLock` 에 그대로 존재한다(제거 금지).

**Evidence:** `grep -n "LOCK_EX|unix.LOCK_NB" internal/session/registry_lock_unix.go` 일치; `grep -n "lockRetryDelay" internal/session/registry.go` 일치.

### AC-CFS3-008 — LockTimeout 기본값 불변 (REQ-CFS3-011, 관련 전작 REQ-CFS-012)

**Given** 본 SPEC 변경 전후,
**When** `grep -n "^const LockTimeout" internal/session/registry.go` 실행,
**Then** `LockTimeout = 2 * time.Second` 가 불변.

**Evidence:** grep 출력 일치.

### AC-CFS3-009 — 60s override 평가 (REQ-CFS3-010)

**Given** 구조적 뮤텍스가 적용되어 `TestRegisterSessionConcurrent` 가 결정론적으로 green 이고,
**When** run-phase 가 `registry_test.go:291` 의 60s override 를 제거/축소/유지 중 하나로 결정하면,
**Then** 그 결정과 근거가 `progress.md` §E.2 에 기록된다. 제거를 택한 경우 프로덕션 기본값(2s)에서 `-count=20 -race` green 이 관측되어야 한다.

**Evidence:** progress.md §E.2 의 결정 기록 + (제거 시) 2s 기본값에서의 green 테스트 출력.

### AC-CFS3-010 — 하드코딩 방지 (REQ-CFS3-011, REQ-CFS3-012)

**Given** 본 SPEC 이 신규 상수를 도입하는 경우,
**When** 소스 정적 검사,
**Then** 해당 상수가 `internal/session` 패키지 레벨 `const` 또는 `internal/config/defaults.go` 에 단일 원천으로 존재하며 hot path 인라인 리터럴이 아니다. 본 SPEC 은 프로덕션 코드를 `internal/session` 외에 수정하지 않는다(defaults.go 신규 상수 추가는 예외).

**Evidence:** grep + 코드 리뷰.

### AC-CFS3-011 — 전체 스위트 + vet + lint + race green (REQ-CFS3-013)

**Given** 본 SPEC 변경이 적용되면,
**When** 다음 명령을 실행:
- `go test ./...`
- `go test -race ./internal/session/...`
- `go vet ./...`
- `golangci-lint run`

**Then** 네 명령 모두 종료 코드 0.

**Evidence:** 각 명령의 종료 코드 + verbatim tail 출력.

### AC-CFS3-012 — PR 경유 (REQ-CFS3-014)

**Given** 브랜치 보호 `enforce_admins: true`,
**When** 본 SPEC 의 변경을 머지하면,
**Then** 모든 변경이 PR(squash merge) 을 경유하며 main 직접 push 는 없다.

**Evidence:** `gh pr view <PR> --json mergeCommit,mergedAt` + commit history.

### AC-CFS3-013 — MX 태그 보존/추가 (REQ-CFS3-015)

**Given** 패키지가 기존 `@MX:ANCHOR`(`registry.go:15`, withLock NOTE ~358) 를 보유하고,
**When** 본 SPEC 이 신규 헬퍼/필드를 도입하면,
**Then** 기존 ANCHOR 는 보존되고 신규 표면(`acquireInProcessMutex` 헬퍼 또는 `Registry.mu` 필드)에는 `@MX:NOTE` (또는 fan_in ≥ 3 예상 시 `@MX:ANCHOR`) 가 태깅된다.

**Evidence:** `grep -n "@MX" internal/session/registry*.go` + 리뷰.

---

## §D.1 Severity Classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-CFS3-001 | MUST | 결함 재발 방지 핵심. |
| AC-CFS3-002 | MUST ★ | 반증 가능성이 본 SPEC 의 유효성 조건. 관측 못하면 SPEC 무효. 단, 충분한 iteration(총 10000회 권장) 하의 확률적 수렴이며, 미관측 시 iteration 증가 재실행. |
| AC-CFS3-003 | MUST | ErrLockTimeout 계약 — 전작 SPEC 의 핵심 계약 보존. |
| AC-CFS3-004 | MUST | 지터 백오프 계약 보존 — cross-process 전략 회귀 방지. |
| AC-CFS3-005 | SHOULD | 기아 특성화 유지 — 수치 개선은 관측 부수물. |
| AC-CFS3-006a | MUST | Windows run-phase 빌드 패리티(REQ-COORD-022 연장). run-phase 관측 가능. |
| AC-CFS3-006b | SHOULD | Windows CI 런타임 green. run-phase 불가 → sync-phase CI 증빙. |
| AC-CFS3-007 | MUST | 비차단 + 지터 유지(회귀 방지). |
| AC-CFS3-008 | MUST | 프로덕션 기본값 불변. |
| AC-CFS3-009 | SHOULD | 60s override 결정 자체가 요구; 세부 결정은 run-phase 재량. |
| AC-CFS3-010 | MUST | 하드코딩 방지(CLAUDE.local.md §14). |
| AC-CFS3-011 | MUST | 전체 게이트. |
| AC-CFS3-012 | MUST | PR 의무. |
| AC-CFS3-013 | SHOULD | MX 태깅(M6). |

---

## §D.2 Traceability (REQ ↔ AC)

| REQ | AC |
|-----|----|
| REQ-CFS3-001 | AC-CFS3-001 |
| REQ-CFS3-002 | AC-CFS3-002 ★ |
| REQ-CFS3-003 | AC-CFS3-003 |
| REQ-CFS3-004 | AC-CFS3-003 |
| REQ-CFS3-005 | AC-CFS3-004 |
| REQ-CFS3-006 | AC-CFS3-006a, AC-CFS3-006b |
| REQ-CFS3-007 | AC-CFS3-007 |
| REQ-CFS3-008 | AC-CFS3-007 |
| REQ-CFS3-009 | AC-CFS3-005, AC-CFS3-004, AC-CFS3-003 |
| REQ-CFS3-010 | AC-CFS3-009 |
| REQ-CFS3-011 | AC-CFS3-010 |
| REQ-CFS3-012 | AC-CFS3-010 |
| REQ-CFS3-013 | AC-CFS3-011 |
| REQ-CFS3-014 | AC-CFS3-012 |
| REQ-CFS3-015 | AC-CFS3-013 |

---

## §D.3 Indirect Verification (직접 관측이 어려운 항목)

- **결함의 로컬 미재현성(`§G` gap 1):** AC-CFS3-001/002 는 이 한계를 보상하기 위해 (a) probe 카운터(결정론적 신호), (b) 반증 가능성(뮤텍스 제거 시 실패 관측) 의 두 메커니즘으로 검증한다. CI-only 재현 가능성에 의존하지 않는다.
- **cross-process 공정성(§H 잔여 위험 1):** 본 SPEC 은 이 영역을 건드리지 않으므로 AC 가 없다. 대신 AC-CFS3-007 로 cross-process 전략(지터 백오프)의 회귀만 방지한다.
- **Windows 런타임 결정론성(§G gap 4):** AC-CFS3-006a(run-phase 빌드/vet 수준 MUST) 와 AC-CFS3-006b(sync-phase Windows CI runtime green SHOULD) 로 분리 관측.

---

## §D.4 Closure Gates (Definition of Done)

run-phase 종료 조건:
1. 모든 MUST AC 가 PASS(observability 충족).
2. AC-CFS3-002 의 반증 관측(뮤텍스 제거 시 probe max ≥ 2)이 `progress.md` §E.2 에 인용됨.
3. AC-CFS3-009(60s override 결정)가 `progress.md` §E.2 에 기록됨.
4. `progress.md` §E.2 E1-E7 매트릭스가 관측된 명령/출력과 함께 채워짐.

sync-phase 종료 조건(추가):
5. Windows CI 잡 green 증빙(AC-CFS3-006b — Windows runtime).
6. CI 미재발 관측: 본 SPEC 머지 후 최소 3회 연속 macOS CI green 관측(유한 관측이지만 회귀 탐지 보조).

---

## §D.5 Forward-Looking Checks (미래 회귀 방지)

- **뮤텍스 제거 리팩터 탐지:** AC-CFS3-002 의 falsifiability 서브-케이스는 CI 기본 skip 이므로 future 리팩터를 자동 탐지하지 못한다. 권장: 일종의 "mutation test" 를 periodic job 으로 두거나, falsifiability 검증을 로컬 `make` 타겟으로 등록(선택). 본 SPEC 은 이를 강제하지 않는다.
- **패키지 헬퍼 경로 회귀:** path-keyed 설계(A) 채택 시 `defaultRegistry()` 매 호출 새 인스턴스 경로도 동일 뮤텍스 공유. per-Registry(B) 채택 시 이 회귀 경로가 열리므로, B 안을 택할 경우 AC-CFS3-002 의 probe 를 패키지 헬퍼 경로까지 확장 적용할 것을 권장.
- **cross-process timeout 계약 회귀:** `TestWithLockTimeoutContract` 가 본 SPEC 의 회귀 신호. 테스트 유지.
