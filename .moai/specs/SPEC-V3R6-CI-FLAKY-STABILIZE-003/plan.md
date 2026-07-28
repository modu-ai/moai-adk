# Implementation Plan — SPEC-V3R6-CI-FLAKY-STABILIZE-003

> Tier M (standard). Owner: manager-develop (run-phase). Approach ordered by decision-reversibility — the most change-likely decisions first.

---

## §A. Context

`internal/session` 의 `Registry.withLock`(`registry.go:361`) 는 1000회 in-process 경합 시 macOS CI `-race` 에서 NB-flock 의 공정성 부재로 인해 `ErrLockTimeout`(60s 예산 초과) 을 발생시킨다(`§A`/`§B` 참조). 전작 SPEC-CI-FLAKY-STABILIZE-001 의 jitter 백오프는 확률적 완화일 뿐이었고, 한계가 관측되었다.

본 SPEC 은 **in-process mutex + cross-process flock double-locking** 패턴으로 전환해 동일 프로세스 경합을 결정론적으로 직렬화한다. 레지스트리의 본래 목적인 cross-process 조정(REQ-COORD-008/022) 은 flock + 지터 백오프로 보존된다.

**선순위 결정(결정-가역성 높은 순):**
1. 뮤텍스 범위 선택 (path-keyed vs per-Registry) — §B 결정 안건.
2. `Registry` 구조체 필드 형태(`*sync.Mutex` 포인터 필수 — copylocks 회피).
3. 반증 가능 구조적 테스트 probe 설계.
4. 60s per-acquisition override 축소/제거 평가.
5. 기계적 수정(헬퍼 추출, MX 태깅, 주석).

---

## §B. Known Issues (이슈 / 트레이드오프)

### B.1 뮤텍스 범위 — [RESOLVED: ADOPTED A — path-keyed (orchestrator at Implementation Kickoff Approval)]

> **설계 결정 RECORD (OQ-1 RESOLVED).** Implementation Kickoff Approval 게이트에서 **Option A (path-keyed)** 가 채택되었다. 아래 A/B 표는 rationale 보존용이며, B 안은 채택되지 않았다(rejected).

**채택 이유 (auditor-verified):** `defaultRegistry()` 가 호출마다 **새 `Registry` 인스턴스**를 반환하므로, per-Registry (B) 안은 패키지 헬퍼(`RegisterSession` 등) 가 거치는 경로를 직렬화하지 못한다. 동일 in-process 경합이라도 별도 `Registry` 인스턴스면 뮤텍스가 미공유되어 결함이 잔존한다. path-keyed (A) 만이 모든 in-process 경합 형태(인스턴스 무관)를 구조적으로 차단한다. `Registry` 복사(`WithLockTimeout`) 시 `*sync.Mutex` 를 공유해야 하는 요구(B 안도 동일)가 이미 존재하므로 A 가 추가로 요구하는 설계 비용이 작다.

| 옵션 | 설명 | 장점 | 단점 |
|------|------|------|------|
| **A. path-keyed (ADOPTED)** | 패키지 레벨 `sync.Map[absLockPath]*sync.Mutex`; `Registry.withLock` 가 LoadOrStore 후 Lock. | (1) 동일 경로를 가리키는 별도 `Registry` 인스턴스(예: 패키지 헬퍼 `RegisterSession` 가 호출마다 생성하는 `defaultRegistry()`) 도 직렬화. (2) AC-CFS3-002 의 반증 가능 신호가 경로-수준이라 모든 in-process 경합 형태를 커버. | 맵이 프로세스 수명 동안 단조 성장(테스트 임시 경로). 실영향 미미. |
| **B. per-Registry (REJECTED)** | `Registry.mu *sync.Mutex` 필드; `NewRegistry` 에서 초기화. | 단순. 맵 관리 부재. | 패키지 헬퍼(`defaultRegistry()` 매 호출 새 인스턴스) 경로를 직렬화하지 못함 — 동일 in-process 라도 별도 `Registry` 면 뮤텍스 미공유. `defaultRegistry()` 가 새 인스턴스를 반환하므로 본 SPEC 의 결함 경로 자체가 B 안에서는 수정되지 않는다. |

[RESOLVED: ADOPTED A — path-keyed (orchestrator at Implementation Kickoff Approval)]

### B.2 Windows parity 관측 공백

`registry_lock_windows.go` 의 in-process 직렬화는 코드 수준에서 동일 추상화로 보장된다(뮤텍스 자체는 OS-독립적 `sync.Mutex`). 단, 본 리포지토리의 Windows CI 잡에서의 결정론적 green 실검은 sync-phase 관측에 의존한다. run-phase는 `GOOS=windows go vet ./internal/session/...` 와 크로스 컴파일로 빌드 호환성만 보장하고, Windows 런타임 실검은 sync-phase 의 CI 증빙으로 연기한다(§G gap 4).

### B.3 반증 가능 테스트 probe 설계 난이도

AC-CFS3-002 의 구조적 테스트는 "뮤텍스 제거 시 실패" 여야 하지만, 동시에 (a) 우발적 flaky 가 아님, (b) 타이밍 의존 아님, (c) 뮤텍스 존재 시 결정론적 green 여야 한다.

후보 probe 설계(결정론적):
- `atomic.Int64` inCriticalSection 카운터를 뮤텍스 보호 구간 진입 시 increment, 이탈 시 decrement.
- 테스트는 카운터가 1 을 초과한 적이 있는지(=임계 구역 동시 진입) 를 관측. 뮤텍스 존재 시 항상 0 또는 1, 제거 시 ≥ 2 관측 가능.
- N goroutine × M iteration 으로 stress. 단, "관측 가능성" 은 스케줄러에 따라 달라지므로 **충분한 iteration** 으로 확률을 1 에 수렴시킨다(결정론에 가까운 관측).

이 테스트 자체가 flaky 가 되는 것을 막기 위해 (a) iteration 수를 충분히 크게(예: 총 10000 회), (b) probe 의 increment/decrement 는 반드시 atomic, (c) `_test.go` 안에서만 probe 를 노출(exported 표면 아님).

**★ Probe placement (load-bearing, AC-CFS3-002):** atomic counter probe 의 inc/dec 는 **`Registry.withLock` 본문 전체**를 bracket 해야 한다 — in-process mutex 획득 직후(동일 프로세스 lock 획득 후 `withLock` 본문 최상단)에 inc, 함수 exit 에서 dec. 이 배치가 probe 로 하여금 "현재 `withLock` 본문 내부에 있는 goroutine 수" 를 측정하게 하고, "현재 flock 을 보유한 goroutine 수" 가 아니게 만든다. OS-level flock 은 post-acquire 임계 구역을 단일 보유자로 직렬화하므로, probe 가 flock 보유 구간만 bracket 하면 뮤텍스 존재 여부와 무관하게 max ≤ 1 만 관측한다 — 양방향 반증(적용 시 ≤ 1 / 제거 시 ≥ 2) 이 불가능한 vacuous AC 가 되어 `verification-claim-integrity.md` §1.1 surface 3 에 위반된다.

### B.4 60s per-acquisition override 결정

`registry_test.go:291` 의 `r = r.WithLockTimeout(60 * time.Second)` 는 프로덕션 기본값(2s) 이 테스트의 in-process 경합을 감당 못했기 때문이다. 구조적 뮤텍스 적용 후에는 in-process 경합이 직렬화되므로 2s 로도 충분할 것으로 예상. 단, run-phase 가 실측 후 (i) 2s 유지 green → 제거, (ii) 약간의 여유 필요 → 5~10s 축소, (iii) 유지(운영 여유) — 중 합리적 결정을 택한다. 결정과 근거를 progress.md §E.2 에 기록.

### B.5 `*sync.Mutex` 포인터 — copylocks 회피 (설계 제약)

`Registry.WithLockTimeout`(`registry.go:141`) 이 `clone := *r` 얕은 복사를 한다. 따라서:
- `Registry` 에 `sync.Mutex` **값 필드**를 두면 안 된다 → `go vet` copylocks 경고.
- `*sync.Mutex` **포인터 필드**를 둔다 → clone 이 동일 뮤텍스를 가리킨다 → 올바른 의미론(동일 파일 레지스트리 복제본은 동일 in-process 직렬화를 공유).
- path-keyed 설계(A)에서는 `Registry.mu *sync.Mutex` 대신 `Registry.withLock` 내에서 `acquireInProcessMutex(absLockPath)` 헬퍼 호출 — 포인터 필드 불필요, 맵에서 조회.

---

## §C. Pre-flight (run-phase 착수 전 확인)

- [ ] 현재 브랜치가 `feat/SPEC-V3R6-CI-FLAKY-STABILIZE-003` (또는 동등 worktree), base `origin/main`.
- [ ] `git fetch origin main` + divergence 확인(orchestrator pre-spawn sync check).
- [ ] `internal/session/registry.go`, `registry_lock_unix.go`, `registry_lock_windows.go`, `registry_starvation_test.go`, `registry_test.go` 가 HEAD 와 일치.
- [ ] §B.1 [RESOLVED: ADOPTED A — path-keyed] 가 확정되어 있음(orchestrator 가 Implementation Kickoff Approval 게이트에서 채택). run-phase M2 는 A 를 전제로 진행한다.

---

## §D. Constraints (코딩/테스트 표준)

- **Go test isolation:** `t.TempDir()`; 병렬 테스트에서 OTEL env 금지(`CLAUDE.local.md §WARN`).
- **Hardcoding prevention:** 신규 상수(예: 맵 초기 용량 — 실질적으로 필요하지 않으면 추가하지 않음) 는 `internal/session` 패키지 `const` 또는 `internal/config/defaults.go` 단일 원천.
- **Copylocks:** `Registry` 얕은 복사 경로(`WithLockTimeout`) 호환성 — `*sync.Mutex` 포인터 또는 path-keyed 헬퍼.
- **Lock ordering:** `Registry.in-process mutex` → `registryLock.mu`(fd 가드) → flock. 순환 없음. 리뷰 확인.
- **말투:** 영문 코드·주석, 한국어 산출물.
- **MX tags:** `registry.go:15` ANCHOR + withLock NOTE 보존. 신규 내보낸 헬퍼(`acquireInProcessMutex` 등) 에 `@MX:NOTE`.

---

## §E. Self-Verification (run-phase E1-E7 매트릭스)

run-phase 가 `progress.md` §E.2 에 채우는 항목들:
- E1: AC PASS/FAIL 매트릭스(AC-CFS3-001..015).
- E2: `go test -race ./internal/session/...` 베리스 출력.
- E3: `go test -cover ./internal/session/...` 커버리지(패키지 85%+ 유지).
- E4: Archived-agent / AskUserQuestion boundary grep — 해당 없음(코드 전용).
- E5: `golangci-lint run` + `go vet ./...` green.
- E6: commit/push 상태.
- E7: 동기화 게이트 단계에서 sync-auditor 4-dim 스코어.

---

## §F. Technical Approach — Milestones (decision-reversibility 역순)

> Milestone 은 결정-가역성 높은 순으로 배치. P0 가 가장 변경 가능성 높은 의사결정.

### P0 — 설계 확정 (뮤텍스 범위) — [RESOLVED: ADOPTED A]

§B.1 OQ-1 은 orchestrator Implementation Kickoff Approval 게이트에서 **A (path-keyed)** 로 확정되었다. run-phase 는 A 가 적용되어 있음을 확인하고, 아래 milestone 들(M1~M6) 은 A 를 전제로 파생한다. (다시 결정을 여는 것이 아님 — ADOPTED A 를 그대로 실행.)

### M1 — RED: 반증 가능 구조적 테스트 먼저 작성 (TDD)

`registry_starvation_test.go` 또는 신규 `registry_inprocess_mutex_test.go` 에:

1. **TestRegistryWithLockInProcessSerialization** (AC-CFS3-002): N goroutine × M iteration 경합시 probe 가 임계 구역 동시 진입을 관측하지 않는다(카운터 max ≤ 1). 결정론적 관측을 위해 N=10, M=1000(총 10000 회) 정도.
2. **Falsifiability check**: 동일 테스트의 서브-케이스로, 의도적으로 뮤텍스를 끄는 옵션(`Registry.noMutex` 테스트 전용 필드, 또는 별도 bypass 함수) 을 두고 그 상태에서 카운터 max ≥ 2 가 관측됨을 확인. 이 서브-케이스는 일반 CI 에서는 skip(`t.Run("falsifiability", ...)` + env flag) 하고 로컬 검증 용도로만 실행.
3. **TestRegistryWithLockCrossProcessTimeoutPreserved** (AC-CFS3-004 보강): 기존 `TestWithLockTimeoutContract` 의 변주 — blocker 가 `registryLock.acquire` 직접 경로(뮤텍스 미경유) 로 flock 보유 시, `Registry.withLock` 경로가 여전히 `ErrLockTimeout` 을 반환함. (이미 `TestWithLockTimeoutContract` 가 커버하므로, 신규 테스트가 필요할지는 run-phase 가 중복 회피 후 결정.)

관측: M1 테스트는 뮤텍스 구현 전에는 실패해야 한다(RED). 단, "결정론적 실패" 가 아니라 "구조적 실패" 여야 하므로, M1 은 우선 probe 카운터 관측치를 기록하는 characterization 용도로 작성하고, M2 적용 후 green 전환을 관측한다.

### M2 — GREEN: in-process mutex 적용

선택된 설계(A 권장) 에 따라:

**A (path-keyed):**
```go
// registry.go (or new registry_inprocess.go)
var inProcessMutexes sync.Map // map[string]*sync.Mutex

func acquireInProcessMutex(absLockPath string) *sync.Mutex {
    v, _ := inProcessMutexes.LoadOrStore(absLockPath, &sync.Mutex{})
    return v.(*sync.Mutex)
}

// in withLock, at the top:
absLockPath, _ := filepath.Abs(lockPath)
ipm := acquireInProcessMutex(absLockPath)
ipm.Lock()
defer ipm.Unlock()
// ... 기존 flock 획득 + 지터 백오프 루프 + 임계 구역 ...
```

**B (per-Registry):** `Registry` 에 `mu *sync.Mutex` 필드 추가; `NewRegistry` 에서 초기화; `WithLockTimeout` clone 은 동일 포인터 공유(자동); `withLock` 상단에서 `r.mu.Lock()` / `defer r.mu.Unlock()`.

양안 모두 flock 루프와 `ErrLockTimeout` 반환 경로는 **변경 없음**.

### M3 — 기존 계약 테스트 전수 green 확인

- `TestRegisterSessionConcurrent` (registry_test.go:283) — green.
- `TestRegisterStarvationCharacterization`, `TestWithLockTimeoutContract`, `TestLockRetryDelayBoundsAndJitter` — 모두 green.
- 전체 `go test -race ./internal/session/...` green.

### M4 — 60s override 평가 (REQ-CFS3-010)

`registry_test.go:291` 의 60s override 를:
- (i) 제거(`WithLockTimeout` 호출 삭제) 후 `go test -race -count=20 -run TestRegisterSessionConcurrent ./internal/session/` green 관측 → 제거 확정.
- (ii) 제거 시 간헐 실패 → 적정 값(예: 10s) 로 축소.
- (iii) 유지(운영 여유) — 주석으로 "구조적 뮤텍스가 정확성 메커니즘이고, 이 60s 는 운영 여유일 뿐" 명시.

결정과 근거를 `progress.md` §E.2 에 기록.

### M5 — 크로스 플랫폼 빌드 보장

- `GOOS=windows go vet ./internal/session/...` green.
- `GOOS=windows go build ./internal/session/...` green.
- `registry_lock_windows.go` 가 in-process mutex 의미론 동일하게 적용받는지 코드 리뷰(뮤텍스는 `Registry.withLock` 공통 경로에 있으므로 자동 적용 — but 명시적으로 리뷰에서 확인).

### M6 — MX 태깅 + 주석 + 커밋

- 신규 헬퍼(`acquireInProcessMutex` 또는 `Registry.mu`) 에 `@MX:NOTE` ("in-process 직렬화; cross-process flock 과 이중 락킹").
- `registry.go:326` 의 기존 자체 기술 주석(확률적 완화 한계) 갱신 — 본 SPEC 이 구조적으로 해결했음을 명시, 단 cross-process 경합은 여전히 동일 전략 유지.
- Conventional Commit: `fix(session): structural in-process mutex for registry flock starvation (SPEC-V3R6-CI-FLAKY-STABILIZE-003)`.

---

## §G. Anti-Patterns (회피 패턴)

- **AP-CFS3-001** — 차단 flock(`LOCK_EX` 단독) 전환. REQ-CFS3-007 위반. cross-process 데드 홀더 시 무한 블록 재도입.
- **AP-CFS3-002** — 지터 백오프 제거. REQ-CFS3-008 위반. cross-process 공정성 퇴행.
- **AP-CFS3-003** — 뮤텍스 값 필드(`sync.Mutex`) 를 `Registry` 에 두어 `WithLockTimeout` clone 시 copylocks 경고. §B.5 위반.
- **AP-CFS3-004** — 뮤텍스를 `registryLock.acquire` 내부에 넣기. blocker 직접 경로(`TestWithLockTimeoutContract`) 가 뮤텍스를 공유하게 되어 ErrLockTimeout 계약이 깨짐. REQ-CFS3-004 위반.
- **AP-CFS3-005** — 신규 구조적 테스트를 타이밍 의존 `time.Since` 비교로 작성. 우발적 flaky 재발.
- **AP-CFS3-006** — `TestRegisterSessionConcurrent` 의 60s override 를 "정확성 게이트" 로 착각. REQ-CFS3-010 위반 — 뮤텍스가 정확성 메커니즘이며 60s 는 운영 여유일 뿐.
- **AP-CFS3-007** — `internal/session/lock.go`(`flockLock`, SPEC-V3R2-RT-004 CHECKPOINT 용) 를 본 SPEC 범위로 혼동해 수정. §C.2 위반.
- **AP-CFS3-008** — WES-001 / worktree / launcher 파일에 손대기. §C.2 위반 (STANDALONE SPEC).

---

## §H. Cross-References

- spec.md: 본 SPEC 의 요구사항/AC 매트릭스.
- acceptance.md: Given-When-Then 시나리오 + 반증 가능 AC 상세.
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX 태깅(M6).
- `CLAUDE.local.md §6`(테스트 격리), §14(하드코딩 방지), §22.8(worktree auto-toggle 기본 off — 본 SPEC 무관하지만 reference).
- 선행 SPEC: `SPEC-V3R6-MULTI-SESSION-COORD-001`(REQ-COORD-008/022), `SPEC-CI-FLAKY-STABILIZE-001`(REQ-CFS-010/011/014/015/016).
