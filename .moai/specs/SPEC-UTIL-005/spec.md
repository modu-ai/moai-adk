---
id: SPEC-UTIL-005
title: "walkSourceFiles Incremental Scan"
version: "0.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-24
author: MoAI v2.15 Backlog Writer
priority: P2 Medium
phase: "v2.15 — Utility Performance Backlog"
module: "internal/hook/quality/"
dependencies: []
related_problem: [IMP-V3U-023]
related_pattern: []
related_principle: []
related_decision: []
related_theme: "v2.15 Utility Performance"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "hook, quality, walkSourceFiles, incremental, performance, monorepo"
---

# SPEC-UTIL-005: walkSourceFiles Incremental Scan

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-24 | MoAI v2.15 Backlog Writer | Initial draft. v2.14 `manager-quality` 멀티관점 리뷰 Warning 등급(PERFORMANCE, `internal/hook/quality/astgrep_gate.go:29-41`). Post-tool-use 훅 트리거마다 전체 트리 재귀 탐색 문제. SPEC-UTIL-002 머지 이후 착수 가능. Non-breaking (changedFiles optional). |

---

## 1. Goal (목적)

`walkSourceFiles`의 full-tree 재귀 탐색 비용을 incremental scan으로 전환한다. `RunAstGrepGateV2`가 post-tool-use 훅으로 파일 저장마다 트리거될 때, 전체 프로젝트 디렉토리(`filepath.WalkDir(projectDir)`)를 매번 순회하는 대신 **hook payload에 이미 포함된 `changedFiles`** 리스트를 primary source로 활용한다. changedFiles가 비어있거나 제공되지 않을 경우에만 기존 full scan으로 폴백하여 backward compatibility를 보장한다.

## 2. Scope (범위)

### 2.1 In Scope

- `internal/hook/quality/astgrep_gate.go`의 `walkSourceFiles` 함수 시그니처 확장 (optional changedFiles 파라미터)
- `RunAstGrepGateV2` 함수가 hook payload의 changedFiles를 `walkSourceFiles`로 전달하도록 연결
- Backward-compatible: changedFiles가 nil 또는 empty이면 기존 full scan 동작
- 10k-file monorepo fixture 기반 benchmark 추가 및 회귀 가드

### 2.2 Out of Scope

- Hook payload 스키마 변경 금지 — `changedFiles` 필드는 이미 존재, 추가/제거 없음
- Git 통합 자동화(`git diff --name-only`) 금지 — hook payload가 이미 신뢰 가능한 source
- 파일 mtime 기반 결과 캐싱 금지 — staleness 위험
- Symlink / submodule 동작 변경 금지 — SPEC-UTIL-002의 기존 exclusion 규칙 재사용
- `walkSourceFiles`의 suppression pairing 로직 자체는 수정 없음 (입력 범위만 축소)

## 3. Environment (환경)

### 3.1 현재 코드 상태

`astgrep_gate.go:29-41` (simplified):

```
func walkSourceFiles(projectDir string) ([]string, error) {
    var files []string
    err := filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
        // ... exclusion 규칙 (vendor, node_modules, .git 등) ...
        if isSourceFile(path) {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}

func RunAstGrepGateV2(ctx context.Context, projectDir string, ...) error {
    files, err := walkSourceFiles(projectDir)
    // ... lint ...
}
```

### 3.2 비용 분석

- `filepath.WalkDir`: O(전체 파일 수), SSD에서도 10k 파일 기준 ~30-80ms
- Post-tool-use 훅은 파일 저장 이벤트마다 트리거 → 활발한 개발 세션에서 분당 수십 회 호출 가능
- 대형 모노레포(10k+ 파일): 누적 CPU 소비 무시 불가
- 그러나 **실제 변경된 파일은 1-5개**가 전형적 → 99% 파일은 매번 무의미하게 스캔됨

### 3.3 Hook payload의 changedFiles 필드

post-tool-use 훅 payload 스키마(`internal/hook/payload.go`)에는 이미 `changedFiles []string` 필드가 존재하지만, `RunAstGrepGateV2`는 이를 사용하지 않고 full scan으로 덮어쓴다.

## 4. Requirements (요구사항, EARS)

- **REQ-UTIL-005-001** (Ubiquitous): `walkSourceFiles`는 optional `changedFiles []string` 파라미터를 수용해야 한다 (variadic 또는 options struct 중 결정, 구현 단계).
- **REQ-UTIL-005-002** (State-Driven): `changedFiles`가 nil 또는 empty이면 `walkSourceFiles`는 기존 full scan 동작(`filepath.WalkDir`)을 수행해야 한다.
- **REQ-UTIL-005-003** (State-Driven): `changedFiles`가 non-empty이면 `walkSourceFiles`는 해당 경로만 source-file 필터링 + suppression pairing 검사를 수행해야 하며, 전체 트리 순회를 하지 않아야 한다.
- **REQ-UTIL-005-004** (Ubiquitous): `RunAstGrepGateV2` 시그니처에 changedFiles를 전달할 수 있는 수단(variadic 또는 options struct)을 추가해야 한다. 기존 caller(변경 안 한 경우) 동작 유지.
- **REQ-UTIL-005-005** (Event-Driven): 10k-file synthetic monorepo fixture에 대해, full scan baseline 대비 changed=5 files incremental scan의 소요시간이 90% 이상 감소해야 한다.
- **REQ-UTIL-005-006** (Ubiquitous): 변경은 non-breaking이다. 기존 `walkSourceFiles(projectDir)` / `RunAstGrepGateV2(ctx, projectDir, ...)` 호출은 동일한 결과를 반환해야 한다.

## 5. Acceptance Criteria (수락 기준)

- **AC-UTIL-005-01**: `BenchmarkWalkSourceFiles_10kFullScan` baseline 벤치마크 추가, CI metric으로 등록 (회귀 탐지용).
- **AC-UTIL-005-02**: `BenchmarkWalkSourceFiles_5ChangedFiles_IncrementalScan` 추가, 소요시간 < 100ms (10k-file fixture 환경).
- **AC-UTIL-005-03**: `TestWalkSourceFiles_BackwardCompat_NilChangedFiles` 추가. changedFiles nil일 때 full scan 결과가 pre-fix와 동일함을 확인.
- **AC-UTIL-005-04**: `TestRunAstGrepGateV2_IncrementalPath` 추가. 10k-file fixture + changedFiles=["a.go", "b.go"] → lint는 해당 2개 파일에 대해서만 실행되고, 나머지 9998개 파일은 참조되지 않음 (FS access audit hook으로 검증).

## 6. Constraints (제약)

- Non-breaking: 기존 caller, hook payload 스키마, JSON 출력 포맷 전부 불변
- Hook payload의 `changedFiles`는 이미 존재 → 신규 필드 추가 없음
- SPEC-UTIL-002의 exclusion 규칙(`vendor/`, `node_modules/`, `.git/`, symlink 무시 등) 재사용
- Windows/macOS/Linux 3개 플랫폼에서 동일하게 동작 (경로 구분자 정규화 필수)

## 7. Risks (위험)

- **R1 (중위험)**: Hook payload의 changedFiles가 신뢰할 수 없는 경우 (uncommitted/stash 상태) → suppression pairing 누락 가능. **Mitigation**: hook payload의 changedFiles를 primary source로 신뢰하되, payload 자체가 nil/empty일 때만 full scan 폴백. git 외부 상태 처리는 v3.0 이슈로 defer.
- **R2 (저위험)**: Symlink / submodule edge case에서 changedFiles 경로가 symlink 타깃을 가리킬 수 있음. **Mitigation**: SPEC-UTIL-002의 exclusion 규칙을 incremental path 내부에서도 동일 적용.
- **R3 (저위험)**: changedFiles 경로가 projectDir 밖을 가리키는 경우 보안 이슈. **Mitigation**: `filepath.Rel(projectDir, changedFile)` + `strings.HasPrefix(rel, "..")` 검사로 escape 방지.

## 8. Dependencies (의존성)

- **Blocked by**: SPEC-UTIL-002 Phase 3.2 구현 완료 (현재 `release/v2.14.0`에서 머지 대기). v2.15 착수 시점에는 main 반영됨
- **Blocks**: 없음
- **Related**: SPEC-UTIL-002 §3 (exclusion 규칙 원본), SPEC-UTIL-004 (동일 `internal/hook/` 트리 성능 backlog)
