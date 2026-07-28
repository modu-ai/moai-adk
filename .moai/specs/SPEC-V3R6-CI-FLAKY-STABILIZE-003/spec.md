---
id: SPEC-V3R6-CI-FLAKY-STABILIZE-003
title: "CI Flaky Test Stabilization — session registry flock starvation structural fix (in-process mutex + cross-process flock double-locking)"
version: "0.2.0"
status: draft
created: 2026-07-29
updated: 2026-07-29
author: manager-spec
priority: P1
phase: "v3.0.2"
module: "internal/session"
lifecycle: spec-anchored
tags: "ci, flaky-test, flock, starvation, concurrency, mutex, double-locking, cross-platform, structural"
tier: M
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-29 | manager-spec | Initial plan-phase draft. Structural fix for the macOS CI flaky `TestRegisterSessionConcurrent` (`internal/session`). Supersedes the jitter-backoff mitigation in SPEC-CI-FLAKY-STABILIZE-001 defect 2 (REQ-CFS-010/011/014/015/016), which self-documented as a probabilistic, non-structural mitigation and has now failed on macOS CI (`-race`, run `30374096068`, 2026-07-28). Introduces the standard in-process mutex + cross-process flock double-locking pattern. |
| 0.2.0 | 2026-07-29 | manager-spec | iter-1 audit defect fixes D1–D5 applied (OQ-1 resolved A path-keyed); technical approach unchanged (auditor-verified-sound). D1: OQ-1 ADOPTED A (path-keyed) — `defaultRegistry()` returns new Registry per call so B would not serialize package-helper path. D2: AC-CFS3-002 probe placement brackets entire `withLock` body. D3: AC-CFS3-002 Then-clause softened to probabilistic convergence. D4: REQ-CFS3-001 scope extended to entire `withLock` body incl. NB-flock retry loop. D5: AC-CFS3-006 split into 006a (build parity MUST) / 006b (Windows CI runtime SHOULD). |

---

## §A. 배경 (Context)

`internal/session` 의 다중 세션 레지스트리(`active-sessions.json`)는 `SPEC-V3R6-MULTI-SESSION-COORD-001` 이 도입한 cross-process 조정 프리미티브다. 모든 mutation 경로(`Register` / `Heartbeat` / `Deregister` / `Purge`)는 `Registry.withLock` (`registry.go:361`) 을 경유해 비차단 `flock(LOCK_EX|LOCK_NB)` 으로 보호된 임계 구역을 직렬화한다.

`TestRegisterSessionConcurrent` (`registry_test.go:283`) 는 10 goroutine × 100 registration = 1000회의 **동일 프로세스 내(in-process)** 경합으로 이 락을 스트레스한다. 2026-07-28 macOS CI (`-race`, run `30374096068`) 에서 이 테스트가 102.01s 만에 다음과 같이 실패했다:

```
--- FAIL: TestRegisterSessionConcurrent (102.01s)
    registry_test.go:316: concurrent Register: session registry: lock acquisition timed out:
    registry lock flock .../active-sessions.json.lock: resource temporarily unavailable
```

이 실패는 SPEC-CI-FLAKY-STABILIZE-001 defect 2 의 재발이다. 전작 SPEC은 결함 2 에 대해 **jittered exponential backoff** (`lockRetryDelay`, `registry.go:330`) 와 **기아 특성화 테스트** (`registry_starvation_test.go`) 를 도입했으나, `registry.go:326` 에 스스로 "이것은 확률적 완화이지 구조적 공정성이 아니다 — 지속적 경합 하에서 반복 패배자는 여전히 가능하다" 고 명시했고, macOS CI `-race` 조건에서 정확히 그 시나리오가 관측되었다.

### A.1 로컬 미재현

`registry_starvation_test.go:8` 에 이미 문서화된 대로, 이 결함은 로컬에서 재현되지 않는다. 본 SPEC 의 검증 축은 (a) 메커니즘 정확성 — 식별된 코드 경로를 구조적으로 수정했는가, (b) 반증 가능한 구조적 AC — 뮤텍스를 제거하면 AC 가 실패하는가, (c) CI 미재발 — 의 세 축이다. 로컬 red→green 재현은 검증 축이 아니다(`§G` 참조).

---

## §B. 문제 정의 (Problem Statement)

### B.1 근본 원인 — 비차단 flock + in-process 직렬화 부재

`registryLock.acquire` (`registry_lock_unix.go:31`) 는 `unix.FLOCK_EX | unix.LOCK_NB`(비차단) 를 사용한다. 구조체의 `mu sync.Mutex` 필드는 `fd` 필드 보호용일 뿐, **파일 임계 구역 자체를 직렬화하지 않는다**. `Registry.withLock` 은 NB-flock → EAGAIN → `lockRetryDelay` 지터 백오프 sleep → 재시도 루프를 `lockTimeout` 데드라인까지 반복한다.

NB flock 은 커널 대기열에 경합자를 큐잉하지 않는다. 즉 **동일 프로세스의 goroutine 들 사이에 커널 공정성 보장이 없다**. 10 goroutine 이 동일 프로세스에서 동일 락 파일에 대해 반복적으로 NB 획득을 시도할 때, 불운한 goroutine 이 반복적으로 경쟁에서 패배해 `lockTimeout` 데드라인을 넘길 수 있다 — 이것이 관측된 `ErrLockTimeout` 이다.

### B.2 왜 "전반적 느림"이 아니라 "기아"인가

`withLock` 의 `deadline` 은 **호출마다** 재계산된다(`registry.go:372`). 테스트의 60s 예산은 **획득 1회당** 예산이다. 1000회 획득 중 단일 `Register` 가 60s 를 넘겼다는 것은 집계 지연으로 설명될 수 없고, 한 goroutine 이 반복적으로 경쟁에서 밀렸음을 뜻한다. 이 판정은 전작 SPEC-CI-FLAKY-STABILIZE-001 §B.2 에서 이미 확립된 것이다.

### B.3 전작의 확률적 완화가 충분하지 않은 이유

`lockRetryDelay` 의 full-jitter 백오프는 경합자들의 깨어남 주기를 무작위화하여 동시성 동기화(synchronization) 를 끊는다. 그러나 NB flock 자체가 커널 공정성 큐를 제공하지 않기 때문에, 지속적 경합 하에서 **반복 패배자가 확률적으로 여전히 존재**한다. 이것은 전작 SPEC 의 §H 잔여 위험 1 과 `registry.go:326` 의 자체 기술 주석이 이미 인정한 한계다. macOS CI `-race` 조건(스케줄러 특성 + race detector 오버헤드)이 이 확률적 창을 실제로 넘어뜨렸다.

---

## §C. 범위 (Scope)

### C.1 대상 파일

- `internal/session/registry.go` — `Registry.withLock` 임계 구역; `Registry` 구조체(또는 패키지 레벨 path-keyed 뮤텍스 맵).
- `internal/session/registry_lock_unix.go` — `registryLock` 의 `mu` 역할 재확정(또는 축소).
- `internal/session/registry_lock_windows.go` — Windows 패리티.
- `internal/session/registry_starvation_test.go` — 기존 특성화/계약 테스트 **보존** + 신규 반증 가능 구조적 AC 추가.
- `internal/session/registry_test.go:283` — `TestRegisterSessionConcurrent`. 결정론화 이후 60s override 축소/제거 평가.

### C.2 비대상 (명시적 분리)

- `internal/session/lock.go` (`flockLock` / `fileLock`) — `SPEC-V3R2-RT-004` CHECKPOINT 서브시스템 전용이며 **다른 락**이다. 레지스트리 락과 동일한 인프라를 공유하지 않는 한 이 SPEC 에서 절대 건드리지 않는다.
- WES-001 / worktree / launcher / `moai cc -w` 관련 파일 — STANDALONE SPEC, 일절 미접촉.

### C.3 제외 항목 (Exclusions)

본 SPEC 은 아래 항목을 명시적으로 out of scope 로 둔다.

### Out of Scope — 락 아키텍처 전환

- 비차단 `LOCK_NB` → 차단 flock (`LOCK_EX` 단독) 전환. 커널 대기열 기반 공정성 확보는 구조적 해법이지만, CLI/hook 컨텍스트에서 데드 홀더에 의한 무한 블록 위험(`AP-MSC-005`) 을 재도입하므로 본 SPEC 에서 다루지 않는다.
- 파일 락 → 다른 동기화 기전(락 서버, DB 트랜잭션, 단일 프로세스 싱글톤 registry) 교체.

### Out of Scope — 프로덕션 동작 변경

- `LockTimeout = 2 * time.Second` 기본값 변경.
- `ErrLockTimeout` 에러 계약(반환 조건, 래핑 형태, `errors.Is(err, ErrLockTimeout)` 판정) 변경 — `TestWithLockTimeoutContract` 가 green 이어야 함.
- `Query` 경로의 eventually-consistent read 설계 변경.
- `Entry` 스키마(`SPEC-V3R6-MULTI-SESSION-COORD-001` REQ-COORD-024 frozen) 변경.
- 레지스트리 파일 포맷(`active-sessions.json`) 변경.

### Out of Scope — 테스트 약화

- `TestRegisterSessionConcurrent` 삭제 / skip / 조건부 비활성화.
- `TestRegisterStarvationCharacterization`, `TestWithLockTimeoutContract`, `TestLockRetryDelayBoundsAndJitter` 삭제 또는 약화 — 전작 SPEC 의 계약을 그대로 보존한다.
- `-race` 플래그를 CI 에서 제거하거나 대상 패키지만 예외 처리.

### Out of Scope — 인접 리팩터링

- 본 결함과 무관한 `internal/session` 코드 정리.
- 다른 flaky 테스트 이력 조사·수정.
- CI 워크플로 재시도 정책(`retry` 횟수) 조정 — 증상 은폐.

### Out of Scope — 전작 SPEC 결함 1 (cobra lazy-sort race)

- `SPEC-CI-FLAKY-STABILIZE-001` 결함 1 (`internal/cli` / `internal/cli/preference` TestMain warm-up) 은 무관하며, 본 SPEC 은 그것을 건드리지 않는다. 본 SPEC 은 동일 전작의 **결함 2 (레지스트리 flock 기아) 만** 구조적으로 대체한다.

---

## §D. 요구사항 (GEARS)

### D.1 핵심 — in-process 직렬화

**REQ-CFS3-001** (Ubiquitous)
`Registry.withLock` 의 본문 전체(NB-flock 재시도 루프 + open + flock + read + mutate + write + close 포함)은, 동일 프로세스 내에서 동일 락 경로를 공유하는 모든 경합자를 직렬하는 **in-process mutex** 로 보호되어야 한다(shall). 이 뮤텍스는 post-acquire 임계 구역만이 아니라 **NB-flock 재시도 루프 전체**를 감싸야 한다(shall) — 기아는 바로 이 재시도 루프 수준에서 발생한다(NB flock 가 커널 공정성 큐를 제공하지 않기 때문에, 불운한 goroutine 이 반복적으로 경쟁에서 밀려 `lockTimeout` 을 넘긴다). 따라서 post-acquire 임계 구역만 감쌀 경우 결함이 잔존하므로, 뮤텍스는 `withLock` 본문 전체를 감싸야 결함이 구조적으로 제거된다. 이 뮤텍스는 동일 프로세스 내 경합자가 다른 동일 프로세스 경합자에 의해 flock 수준에서 굶주리는 것을 구조적으로 차단한다.

**REQ-CFS3-002** (Event-driven / 구조적 반증)
**When** N ≥ 2 goroutine 이 동일한 레지스트리 경로(동일 `Registry` 인스턴스 또는 동일 경로를 가리키는 별도 `Registry` 인스턴스) 에 대해 동시에 `Registry.withLock` 을 진입하면, in-process mutex 는 그들의 임계 구역을 **상호 배타적으로 직렬**해야 한다(shall). 이 요구사항은 **반증 가능한 구조적 테스트**로 검증된다 — 뮤텍스 획득문을 제거하면 테스트가 실패해야 한다(shall). (`acceptance.md` §D AC-CFS3-002)

**REQ-CFS3-003** (Capability gate / 설계 규칙)
**Where** `Registry.withLock` 가 in-process mutex 를 획득한 상태에서 cross-process(또는 `Registry.withLock` 를 경유하지 않는 별도 `registryLock`) 경합자가 이미 OS flock 을 보유하고 있으면, `Registry.withLock` 는 기존의 지터 백오프 재시도 루프에 따라 NB-flock 재시도를 계속해야 하며(shall), 데드라인 도달 시 `ErrLockTimeout` 으로 래핑된 오류를 반환해야 한다(shall). in-process mutex 는 cross-process timeout 경로를 단축하지 않는다.

### D.2 계약 보존 — ErrLockTimeout + jitter

**REQ-CFS3-004** (Event-driven)
**When** `TestWithLockTimeoutContract` (`registry_starvation_test.go:107`) 가 별도의 `registryLock` 인스턴스로 OS flock 을 보유한 상태에서 `Registry.withLock` 경로가 이와 경합하면, 기존과 동일하게 `ErrLockTimeout` 으로 래핑된 오류가 반환되어야 하며(shall), `errors.Is(err, ErrLockTimeout)` 판정 결과는 변경 전후 동일해야 한다. 해당 테스트는 본 SPEC 변경 후에도 수정·삭제 없이 green 이어야 한다(shall).

**REQ-CFS3-005** (Ubiquitous)
`lockRetryDelay` (`registry.go:330`) 의 지터 백오프 계약(`lockBackoffBase`, `lockBackoffCap`, deadline clamp) 은 **보존**되어야 한다(shall). in-process mutex 가 동일 프로세스 경합을 제거한 후에도, cross-process 경합(레지스트리의 본래 목적) 은 여전히 NB flock 이므로 지터 백오프는 cross-process 재시도 전략으로 그대로 쓰인다. `TestLockRetryDelayBoundsAndJitter` 는 green 이어야 한다(shall).

### D.3 크로스 플랫폼 패리티

**REQ-CFS3-006** (Ubiquitous)
in-process mutex guard 는 `registry_lock_windows.go` 경로에 동일하게 적용되어야 한다(shall). in-process 직렬화는 OS-독립적이므로 Windows 빌드에서도 동일한 직렬화 의미론을 가져야 한다. Windows parity 증빙은 두 티어로 분리된다: (a) **run-phase 빌드 호환성(MUST)** — `GOOS=windows go vet ./internal/session/...` + `GOOS=windows go build ./internal/session/...` green 으로 관측(`acceptance.md` AC-CFS3-006a); (b) **sync-phase Windows CI 런타임 green(SHOULD)** — Windows CI 잡의 `go test ./internal/session/...` green 증빙(run-phase 에서는 Windows 런타임 실검이 불가하므로 SHOULD, `acceptance.md` AC-CFS3-006b). Windows parity 는 `REQ-COORD-022` (크로스 플랫폼 호환) 의 연장선이다.

### D.4 금지 — 비차단 flock 제거 / 차단 flock 전환

**REQ-CFS3-007** (Unwanted)
본 SPEC 의 변경은 `unix.LOCK_NB`(비차단) 획득 방식을 차단 flock(`LOCK_EX` 단독) 으로 전환해서는 안 된다(shall not). 이유: CLI/hook 컨텍스트에서 데드 홀더에 의한 무한 블록 위험(`AP-MSC-005`) 재도입. in-process mutex 는 동일 프로세스 경합을 제거하지만, cross-process 데드 홀더 시나리오는 여전히 가능하므로 비차단 + timeout 계약이 여전히 필요하다.

**REQ-CFS3-008** (Unwanted)
본 SPEC 의 변경은 기존의 `lockRetryDelay` 지터 백오프 루프를 제거해서는 안 된다(shall not). cross-process 경합 하에서 지터 백오프는 여전히 유효한 재시도 전략이다. 제거할 경우 cross-process 데드라인 경합의 공정성이 퇴행한다.

### D.5 기존 테스트 보존 + 60s override 평가

**REQ-CFS3-009** (Unwanted)
본 SPEC 의 변경은 `TestRegisterSessionConcurrent`, `TestRegisterStarvationCharacterization`, `TestWithLockTimeoutContract`, `TestLockRetryDelayBoundsAndJitter` 중 어느 것도 삭제·skip·조건부 비활성화해서는 안 된다(shall not). 이 테스트들은 전작 SPEC-CI-FLAKY-STABILIZE-001 의 계약을 구성하며, 본 SPEC 은 그것을 회귀시키지 않는다.

**REQ-CFS3-010** (Event-driven / 평가)
**When** 본 SPEC 의 구조적 수정이 적용되어 `TestRegisterSessionConcurrent` 가 결정론적으로 green 이 되면, run-phase 는 `registry_test.go:291` 의 60s per-acquisition timeout override 가 여전히 필요한지 **평가**해야 한다(shall). 결정(축소/제거/유지)과 그 근거는 `progress.md` §E.2 에 기록되어야 하며(shall), 제거 시 프로덕션 기본값(2s)에서도 테스트가 green 이어야 한다(shall). (평가 후 유지를 결정할 수 있으며, 그 경우에도 구조적 뮤텍스가 정확성 메커니즘이고 60s override는 운영 여유일 뿐임을 주석으로 명시해야 한다.)

### D.6 품질 / 절차

**REQ-CFS3-011** (Ubiquitous / 코딩 표준)
본 SPEC 이 새로 도입하는 임계값/상수(예: path-keyed mutex 맵 초기 용량, 신규 패키지 상수) 가 있으면 `internal/session` 패키지 레벨 `const` 또는 `internal/config/defaults.go` 에 단일 원천으로 정의되어야 한다(shall). 인라인 리터럴이 hot path 에 산재해서는 안 된다(shall not). (`CLAUDE.local.md §14` hardcoding-prevention)

**REQ-CFS3-012** (Ubiquitous)
본 SPEC 의 변경은 `internal/session` 패키지 외의 프로덕션 코드를 수정해서는 안 된다(shall not). 단, 신규 상수가 `defaults.go` 에 추가되는 경우는 예외로 한다.

**REQ-CFS3-013** (Ubiquitous)
변경 후 `go test ./...`, `go test -race ./internal/session/...`, `go vet ./...`, `golangci-lint run` 이 모두 통과해야 한다(shall).

**REQ-CFS3-014** (Ubiquitous)
모든 변경은 PR을 경유해야 한다(shall) — 브랜치 보호 `enforce_admins: true`.

**REQ-CFS3-015** (Ubiquitous / MX)
패키지가 보유한 기존 `@MX:ANCHOR` (`registry.go:15` withLock NOTE ~358)는 보존되어야 하며(shall), 신규 내보낸 표면(예: path-keyed mutex 맵, 신규 헬퍼)에는 `@MX:NOTE` 또는 `@MX:ANCHOR` 가 태깅되어야 한다(shall) — `.claude/rules/moai/workflow/mx-tag-protocol.md`.

---

## §E. 성공 기준 (Success Criteria)

| # | 기준 | 측정 방법 |
|---|------|-----------|
| 1 | `TestRegisterSessionConcurrent` 가 구조적 뮤텍스 적용 후 결정론적으로 green | `go test -race -run TestRegisterSessionConcurrent ./internal/session/` 종료 코드 0 |
| 2 | 반증 가능 구조적 AC(AC-CFS3-002)가 뮤텍스 제거 시 FAIL, 적용 시 PASS | run-phase 에서 뮤텍스 라인을 제거/무력화한 상태의 `go test` 실패 출력 |
| 3 | `TestWithLockTimeoutContract` green 유지 | 테스트 종료 코드 0 (ErrLockTimeout 계약 보존 근거) |
| 4 | `TestLockRetryDelayBoundsAndJitter`, `TestRegisterStarvationCharacterization` green 유지 | 각 테스트 종료 코드 0 |
| 5 | 전체 스위트 + `-race ./internal/session/` + vet + lint green | 각 명령 종료 코드 0 |
| 6 | Windows parity — `registry_lock_windows.go` 경로 동일 직렬화 의미론 | Windows 빌드 크로스 컴파일 + `go vet` (CI Windows 잡에서 실검 증빙은 sync-phase) |
| 7 | `LockTimeout = 2 * time.Second` 기본값 불변 | 소스 grep |

상세 검증 명령과 기대 출력은 `acceptance.md` §D AC 매트릭스에 있다.

---

## §F. 의존성 및 제약

- 선행 SPEC (확립된 계약): `SPEC-V3R6-MULTI-SESSION-COORD-001` (REQ-COORD-008 cross-process 조정, REQ-COORD-022 크로스 플랫폼 패리티), `SPEC-CI-FLAKY-STABILIZE-001` (REQ-CFS-010/011/014/015/016 — 본 SPEC이 structurally 대체).
- 개발 모드: `tdd` (`.moai/config/sections/quality.yaml` `constitution.development_mode`).
- Go 1.26.4.
- 대상 브랜치: `feat/SPEC-V3R6-CI-FLAKY-STABILIZE-003` (base `origin/main`).
- 브랜치 보호: `enforce_admins: true`, `strict: true` — 머지 전 base 최신화 필요.
- 커밋 규약: Conventional Commits, 영문.
- 코드·주석: 영문 / SPEC 산출물 산문: 한국어.
- 하드코딩 방지(`CLAUDE.local.md §14`): 신규 상수는 `internal/session` 패키지 `const` 또는 `internal/config/defaults.go`.
- 테스트 격리(`CLAUDE.local.md §6`): `t.TempDir()`; 병렬 테스트에서 OTEL env 금지.

---

## §G. 검증 공백 (Gaps — 미검증 사항)

`.claude/rules/moai/core/verification-claim-integrity.md` §3.4 에 따라 관측되지 **않은** 것을 명시한다.

1. **결함이 로컬에서 재현되지 않는다.** `registry_starvation_test.go:8` 에 이미 문서화된 전작의 한계가 그대로 적용된다. 본 SPEC 의 수정은 원래 CI 실패의 로컬 red→green 재현으로 검증되지 않는다.
2. **CI 미재발은 유한 횟수 관측이다.** 간헐적 결함에 대해 N 회 성공은 부재의 증거로서 약하다. 본 SPEC 은 (a) 메커니즘 정확성(동일 프로세스 경합을 뮤텍스로 직렬화하는 것이 공정성 부재를 구조적으로 제거하는가) + (b) 반증 가능 AC(뮤텍스를 제거하면 테스트가 실패하는가) 로 검증 축을 삼는다.
3. **반증 가능 구조적 테스트의 설계 정확성.** 뮤텍스 제거 시 테스트가 실패하는 것은 올바른 신호지만, "뮤텍스를 제거한 상태에서의 경합 관측" 자체가 확률적일 수 있다. AC-CFS3-002 는 이 한계를 인정하며, 가능한 한 결정론적인 probe(임계 구역 진입 카운터 / 교차 검증) 로 관측한다 — run-phase 설계 상세.
4. **Windows CI 실검.** 본 SPEC 의 Windows parity 는 코드 수준에서 보장되나(동일한 in-process mutex 추상화), 실제 Windows CI 잡에서의 결정론적 green 증빙은 sync-phase 관측에 의존한다. plan.md §B 에 명시.
5. **60s override 제거 결정.** REQ-CFS3-010 의 평가는 run-phase 관측에 기반하며, 유지·축소·제거 중 어느 쪽이든 합리적 근거가 있으면 수용 가능하다. 본 SPEC 은 결정을 강제하지 않는다.

---

## §H. 잔여 위험 (Residual Risk)

1. **cross-process 경합의 공정성은 본 SPEC 이 다루지 않는다.** in-process mutex 는 동일 프로세스 경합자 간의 공정성을 구조적으로 보장하지만, 서로 다른 moai 프로세스(또는 서로 다른 Claude Code 세션) 간의 cross-process 경합은 여전히 NB flock + 지터 백오프에 의존한다. 이것은 레지스트리의 본래 설계 의도(REQ-COORD-008) 이며, 전작 SPEC §H 잔여 위험 1 이 인정한 대로 프로덕션 cross-process 경합은 세션 수 개 수준이라 실제 기아 확률이 낮다. 본 SPEC 은 이 영역을 건드리지 않는다.

2. **path-keyed mutex 맵의 무한 성장(채택 설계인 경우).** 패키지 레벨 `sync.Map[path]*sync.Mutex` 설계를 택할 경우, 테스트가 무한히 다른 임시 경로를 만들면 맵이 성장한다. 프로세스 수명 자원이므로 실제 영향은 미미하지만, run-phase 가 명시적 청소 정책을 둘지 결정해야 한다. plan.md §F 대안 B 참조.

3. **`Registry` 복사 시 뮤텍스 의미론.** `Registry.WithLockTimeout` 이 `clone := *r` 로 얕은 복사를 한다. 뮤텍스를 값 필드로 두면 `go vet` copylocks 경고가 발생한다. 따라서 `*sync.Mutex` 포인터 필드(path-keyed 맵 설계에서는 동일 뮤텍스를 가리키는 포인터)를 사용해야 하며, clone 도 동일 뮤텍스를 공유해야 한다. 이것은 설계 제약이지 위험은 아니지만, run-phase 가 이를 간과할 위험을 명시한다.

4. **신규 구조적 테스트 자체가 우발적 flaky 가 될 가능성.** 반증 가능 AC(AC-CFS3-002) 의 probe 가 타이밍 의존적이면 신규 테스트 자체가 flaky 를 유발할 수 있다. run-phase 는 probe 를 카운터 기반 결정론적 신호로 설계해야 한다(동시성 안전한 카운터, 절대 타이밍 비교 금지).

5. **뮤텍스 vs flock 순서로 인한 데드락 가능성(현재 설계에서는 없음).** 현재 설계는 `Registry.mu` → `registryLock.mu`(fd 가드) → flock 순서이며 순환 없음. run-phase 가 이 순서를 깨거나 추가 락을 끼워넣지 않는 한 데드락은 발생하지 않는다. 코드 리뷰에서 순서 불변을 확인해야 한다.

---

## §I. Cross-References

- `SPEC-V3R6-MULTI-SESSION-COORD-001` REQ-COORD-008 (cross-process lockfile atomic write), REQ-COORD-022 (cross-platform parity) — 본 SPEC이 보존하는 계약.
- `SPEC-CI-FLAKY-STABILIZE-001` defect 2 (REQ-CFS-010/011/014/015/016) — 본 SPEC이 structurally 대체하는 확률적 완화.
- `SPEC-V3R6-CI-FLAKY-STABILIZE-001`, `-002` — 동일 패밀리의 다른 flaky 결함 (본 SPEC과 무관).
- `.claude/rules/moai/core/verification-claim-integrity.md` §3.4 — Gaps 섹션 의무.
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX 태깅 의무(REQ-CFS3-015).
- `CLAUDE.local.md §6`(테스트 격리), §14(하드코딩 방지) — 코딩 표준.
