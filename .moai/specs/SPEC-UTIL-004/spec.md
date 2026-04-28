---
id: SPEC-UTIL-004
title: "ast-grep ScanMultiple Goroutine Spawn Pattern Alignment"
version: "0.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-24
author: MoAI v2.15 Backlog Writer
priority: P2 Medium
phase: "v2.15 — Utility Performance Backlog"
module: "internal/hook/security/"
dependencies: []
related_problem: [IMP-V3U-022]
related_pattern: []
related_principle: []
related_decision: []
related_theme: "v2.15 Utility Performance"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "astgrep, goroutine, semaphore, performance"
---

# SPEC-UTIL-004: ast-grep ScanMultiple Goroutine Spawn Pattern Alignment

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-24 | MoAI v2.15 Backlog Writer | Initial draft. v2.14 `manager-quality` 멀티관점 코드 리뷰에서 발견된 Warning 등급 성능 이슈(PERFORMANCE, `internal/hook/security/ast_grep.go:207-215`). SPEC-UTIL-002 머지 이후 착수 가능. Non-breaking, public API surface unchanged. |

---

## 1. Goal (목적)

`ast_grep.go` `ScanMultiple`의 goroutine 스폰 패턴을 MX validator(`internal/hook/mx/validator.go:513`) 패턴에 정렬하여, 대용량 파일셋(수천 개)에서 N개 goroutine이 즉시 생성되던 기존 동작을 동시 실행 상한(`runtime.NumCPU()*2`)과 동일한 수의 goroutine 수로 제한한다. 세마포어 획득 위치를 `go func()` 내부에서 외부로 이동하여 goroutine object 메모리 압박을 제거하며, 기존 세마포어 실행 상한 보증은 그대로 유지된다.

## 2. Scope (범위)

### 2.1 In Scope

- `internal/hook/security/ast_grep.go`의 `ScanMultiple` 스폰 루프 변경 (라인 207-215 범위)
- 세마포어 `sem <- struct{}{}` 획득을 `go func()` 호출 **이전**으로 이동
- `defer <-sem`은 goroutine 내부 유지 (기존 패턴 그대로)
- 1000-file synthetic fixture 기반 peak goroutine count 측정 테스트 추가

### 2.2 Out of Scope

- Public API 표면 변경 금지 (`ScanMultiple` 시그니처, 반환 타입, 에러 계약 불변)
- `ast_grep.go`의 다른 함수(`Scan`, `scanFile` 등) 수정 금지
- 세마포어 용량(`runtime.NumCPU()*2`) 변경 금지 — 실행 상한 정책은 유지
- 워커 풀 재설계(채널 기반 worker pool 도입 등) 금지 — v3.0 breaking 이슈

## 3. Environment (환경)

### 3.1 현재 코드 상태

`ast_grep.go:207-215` (pre-fix):

```
for _, file := range files {
    wg.Add(1)
    go func(f string) {
        sem <- struct{}{}       // ← 세마포어 획득이 goroutine 내부
        defer func() { <-sem }()
        defer wg.Done()
        // ... Scan 호출 ...
    }(file)
}
```

- N개 파일 → N개 goroutine 즉시 생성
- 세마포어는 `NumCPU()*2`만 동시 `Scan` 실행 허용
- 나머지 goroutine은 세마포어 채널에서 블록 대기 (goroutine object 메모리 할당됨)
- 10k 파일 × ~2KB goroutine stack → 약 20MB 피크 메모리 압박

### 3.2 MX validator 대비 (reference pattern)

`internal/hook/mx/validator.go:513` (post-SPEC-UTIL-001, 현재 main):

```
for _, file := range files {
    sem <- struct{}{}           // ← 세마포어 획득이 go 호출 이전
    wg.Add(1)
    go func(f string) {
        defer func() { <-sem }()
        defer wg.Done()
        // ... analyzeFile 호출 ...
    }(file)
}
```

- 동시 goroutine 수 = 세마포어 용량 = `NumCPU()*2`
- 생산자(for 루프)가 세마포어에서 블록되어 goroutine 생성 속도 조절

### 3.3 v2.14 AC-UTIL-002-07 검증 gap

SPEC-UTIL-002의 `AC-UTIL-002-07` (TestScanMultiple_SemaphoreBound)는 **활성 `Scan` 호출 수**만 검증하며, 실제 **생성된 goroutine 수**는 검증하지 않는다. 따라서 pre-fix 패턴도 기존 테스트를 통과한다.

## 4. Requirements (요구사항, EARS)

- **REQ-UTIL-004-001** (Ubiquitous): `ScanMultiple`은 세마포어 획득(`sem <- struct{}{}`)을 `go func(...)` 호출 **이전**에 실행해야 한다.
- **REQ-UTIL-004-002** (Ubiquitous): `defer <-sem`은 goroutine 내부에 유지해야 한다. 해제 타이밍은 기존 계약(`Scan` 호출 완료 후)과 동일하게 보전된다.
- **REQ-UTIL-004-003** (Event-Driven): 1000-file synthetic fixture를 `ScanMultiple`에 주입했을 때, atomic counter shim이 관찰한 peak goroutine count는 `runtime.NumCPU()*2` 이하여야 한다 (예: `NumCPU=4` 환경에서 peak ≤ 8).
- **REQ-UTIL-004-004** (Ubiquitous): 변경은 non-breaking이다. `ScanMultiple`의 시그니처, 반환 타입, 에러 계약, 공개 식별자 surface는 byte-identical이어야 한다.

## 5. Acceptance Criteria (수락 기준)

- **AC-UTIL-004-01**: 1000-file synthetic fixture + atomic counter shim 조합으로 `TestScanMultiple_PeakGoroutineBound` 추가. `NumCPU=4` 환경에서 peak goroutine count ≤ 8 관찰.
- **AC-UTIL-004-02**: 기존 `TestScanMultiple_SemaphoreBound` (v2.14 AC-UTIL-002-07) 통과 유지. 활성 `Scan` 호출 수 상한 검증은 회귀 없음.
- **AC-UTIL-004-03**: apigen snapshot 도구로 `internal/hook/security/` 공개 식별자 surface가 pre-fix와 byte-identical임을 확인.

## 6. Constraints (제약)

- Non-breaking: API 시그니처, 에러 타입, JSON 출력 스키마 전부 불변
- stdlib only: `sync`, `runtime`, `sync/atomic` 사용. 새로운 외부 의존성 금지
- SPEC-UTIL-002 머지 완료 후 착수 (현재 `release/v2.14.0` 브랜치에서 머지 대기)
- Windows/macOS/Linux 3개 플랫폼 CI 모두 통과해야 함

## 7. Risks (위험)

- **R1 (저위험)**: 세마포어 획득 위치 변경으로 `ScanMultiple` 진입 지연 가능 (생산자가 세마포어 대기). Mitigation: 기존 MX validator에서 이미 검증된 패턴이며 동일 코드에서 사용 중. 성능 회귀 없음을 `BenchmarkScanMultiple_1kFiles`로 확인.

## 8. Dependencies (의존성)

- **Blocked by**: 없음 (SPEC-UTIL-002는 `release/v2.14.0`에서 머지 대기, v2.15 착수 시점에는 main 반영됨)
- **Blocks**: 없음
- **Related**: SPEC-UTIL-002 (ast-grep 통합 원본), SPEC-UTIL-001 §4.2 (MX validator 세마포어 패턴 reference)
