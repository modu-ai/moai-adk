---
id: SPEC-UTIL-006
title: "analyzeFile Tree-sitter Parse Cache"
version: "0.1.0"
status: draft
created: 2026-04-24
updated: 2026-04-24
author: MoAI v2.15 Backlog Writer
priority: P2 Medium
phase: "v2.15 — Utility Performance Backlog"
module: "internal/hook/mx/, internal/hook/mx/complexity/"
dependencies: []
related_problem: [IMP-V3U-024]
related_pattern: []
related_principle: []
related_decision: []
related_theme: "v2.15 Utility Performance"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "mx, tree-sitter, cache, performance"
---

# SPEC-UTIL-006: analyzeFile Tree-sitter Parse Cache

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-24 | MoAI v2.15 Backlog Writer | Initial draft. v2.14 `manager-quality` 멀티관점 리뷰 Warning 등급(QUALITY + 성능, `internal/hook/mx/validator.go:196-223`). `analyzeFile`이 파일 내 각 exported 함수마다 `complexity.Measure`를 호출, 내부에서 tree-sitter 파싱을 매번 반복하는 구조. SPEC-UTIL-001 머지 이후 착수 가능. Non-breaking. |

---

## 1. Goal (목적)

`analyzeFile`이 한 파일 내 N개 exported 함수에 대해 tree-sitter 파싱을 N번 반복하는 현 구조를 **파일당 단 한 번 파싱** 후 함수별 쿼리만 반복하는 구조로 리팩토링한다. SPEC-UTIL-001에서 도입된 `complexity.Measure` API는 그대로 유지하면서(non-breaking), 내부적으로 `ParseFile` + `MeasureWithTree` 이원화를 추가하여 tree 객체를 재사용한다. 20 함수/파일 기준 tree-sitter 파싱 비용을 1/N로 감축하고, 누적 hot-path(`validator.go:196-223`)의 CPU 소비를 80% 이상 줄인다.

## 2. Scope (범위)

### 2.1 In Scope

- `internal/hook/mx/complexity/` 패키지에 `ParseFile(lang string, content []byte) (*Tree, error)` 공개 API 추가
- `MeasureWithTree(lang string, tree *Tree, funcName string, startLine int) (Result, error)` 공개 API 추가
- 기존 `Measure(lang, content, funcName, startLine)` 시그니처 유지 — 내부적으로 `ParseFile` + `MeasureWithTree` 위임으로 재구현
- `internal/hook/mx/validator.go`의 `analyzeFile` 개조: 파일당 `ParseFile` 1회 → exported 함수마다 `MeasureWithTree` 호출
- `Tree` 객체 lifecycle 관리 (runtime.SetFinalizer 또는 명시적 `Close` 메서드)
- 벤치마크 추가: `BenchmarkAnalyzeFile_BeforeCache` (baseline) / `BenchmarkAnalyzeFile_WithCache`

### 2.2 Out of Scope

- Cross-file parse cache 금지 — 파일 간 재사용은 staleness 위험, v3.0 범위
- 기존 `Measure` API signature 변경 금지 — backward-compat 필수
- tree-sitter 바인딩 교체 금지 — `github.com/smacker/go-tree-sitter` 유지 (SPEC-UTIL-001 §4.1)
- 5+11 언어 지원 범위 변경 금지 — SPEC-UTIL-001에서 확정된 범위 유지
- `measure_nocgo.go` stub 동작 변경 금지 (cgo-off 빌드는 `Supported: false` 반환)

## 3. Environment (환경)

### 3.1 현재 코드 상태

`validator.go:196-223` (simplified):

```
func analyzeFile(path string, ...) ([]Violation, error) {
    content, _ := os.ReadFile(path)
    funcs := extractFunctions(content)
    for _, fn := range funcs {
        res, err := complexity.Measure(lang, content, fn.Name, fn.StartLine)
        // res.Cyclomatic, res.Supported 사용
        if res.Cyclomatic >= 15 {
            // WARN 위반 생성
        }
    }
    // ...
}
```

- 20 exported 함수 × tree-sitter 파싱 = 20회 반복 파싱
- `complexity.Measure` 내부(`measure_cgo.go`)에서 매번 `sitter.NewParser()` + `parser.Parse(nil, content)` 수행
- tree-sitter 파싱 비용: 파일 크기 O(n), 1KB당 ~0.3-0.5ms (Go 기준)
- 10KB Go 파일 × 20 함수 = 파일당 ~60-100ms (vs 이상적 ~3-5ms)

### 3.2 cgo 의존 표면

- `measure_cgo.go`: `github.com/smacker/go-tree-sitter` + 언어별 서브패키지 import, 실제 파싱 수행
- `measure_nocgo.go`: `Result{Supported: false}` stub 반환 — cgo-off 빌드 대응 (SPEC-UTIL-001 §4.2)

이 SPEC은 `measure_cgo.go`를 주 대상으로 하며, `measure_nocgo.go`는 `ParseFile`/`MeasureWithTree`도 stub으로 대응 추가한다.

### 3.3 1 MiB content cap (SPEC-UTIL-001 §4.2에서 확정)

현재 `Measure`는 1 MiB 초과 content를 거부한다. `ParseFile`은 이 검증을 1회 수행하고, `MeasureWithTree`는 pre-validated Tree를 전제로 중복 검증을 생략한다.

## 4. Requirements (요구사항, EARS)

- **REQ-UTIL-006-001** (Ubiquitous): `complexity` 패키지는 `ParseFile(lang string, content []byte) (*Tree, error)` 공개 함수를 제공해야 한다. `Tree`는 tree-sitter AST를 wrapping하는 opaque 타입이다.
- **REQ-UTIL-006-002** (Ubiquitous): `complexity` 패키지는 `MeasureWithTree(lang string, tree *Tree, funcName string, startLine int) (Result, error)` 공개 함수를 제공해야 한다.
- **REQ-UTIL-006-003** (Ubiquitous): 기존 `Measure(lang, content, funcName, startLine)` API 시그니처는 유지되어야 하며, 내부적으로 `ParseFile` + `MeasureWithTree`로 위임 구현되어야 한다.
- **REQ-UTIL-006-004** (Ubiquitous): `analyzeFile`은 파일당 `ParseFile`을 단 1회 호출하고, exported 함수마다 `MeasureWithTree`를 호출해야 한다.
- **REQ-UTIL-006-005** (Ubiquitous): `Tree` 객체는 GC 시 tree-sitter native 메모리가 회수되도록 `runtime.SetFinalizer` 또는 명시적 `Close` 메서드로 관리되어야 한다. use-after-free 금지.
- **REQ-UTIL-006-006** (Event-Driven): 20 함수 × 10KB Go 파일 fixture 기준, `BenchmarkAnalyzeFile_WithCache`는 `BenchmarkAnalyzeFile_BeforeCache` 대비 소요시간 80% 이상 감소해야 한다.

## 5. Acceptance Criteria (수락 기준)

- **AC-UTIL-006-01**: `BenchmarkAnalyzeFile_BeforeCache` baseline 벤치마크 기록 — 20 함수 × 10KB Go fixture, pre-cache 구조 측정.
- **AC-UTIL-006-02**: `BenchmarkAnalyzeFile_WithCache` — baseline 대비 소요시간 80% 이상 감소 (예: baseline 200ms → target ≤ 40ms).
- **AC-UTIL-006-03**: `TestMeasure_BackwardCompat` — SPEC-UTIL-001에서 도입된 `complexity_test.go`의 18+ 테스트가 모두 PASS. `Measure` API signature/동작 byte-identical.
- **AC-UTIL-006-04**: `TestParseFile_MemoryCleanup` — `ParseFile`로 생성된 Tree 객체가 scope exit + `runtime.GC()` 호출 시 finalizer 실행되어 tree-sitter native 메모리가 해제됨을 확인 (테스트 가능한 hook 또는 memory profile delta 측정).

## 6. Constraints (제약)

- Non-breaking: 기존 `complexity.Measure` API 유지, `analyzeFile` 외부 동작 불변
- 1 MiB content cap 유지 — `ParseFile`에서 1회 검증, `MeasureWithTree`는 pre-validated Tree 전제
- cgo 의존: `measure_cgo.go` 대상. `measure_nocgo.go`는 `ParseFile`/`MeasureWithTree` stub도 `Supported: false` 반환 유지
- Race detector (`go test -race`) + `go vet` 통과 필수 — tree 객체 lifecycle 검증
- SPEC-UTIL-001 머지 완료 후 착수 (`release/v2.14.0` 머지 대기)

## 7. Risks (위험)

- **R1 (중위험)**: tree-sitter `Tree` 객체 lifetime 관리 오류 → use-after-free, 잠재 crash. **Mitigation**: `runtime.SetFinalizer`로 GC 시 native 해제 보장, 테스트에서 race detector + `go vet` 엄격 적용. 추가로 Tree에 대한 `Close()` 명시적 메서드도 제공하여 deterministic cleanup 경로 확보.
- **R2 (저위험)**: 1 MiB cap 검증이 `ParseFile`과 `MeasureWithTree` 양쪽에서 중복 수행될 경우 성능 회귀. **Mitigation**: `ParseFile`에서 1회 검증 후 Tree에 flag 저장, `MeasureWithTree`는 Tree 전제로 재검증 skip.
- **R3 (저위험)**: `MeasureWithTree`에 잘못된 `lang` 인자 전달 시 쿼리 패턴 mismatch 가능. **Mitigation**: Tree 생성 시 lang을 Tree 구조체에 저장, `MeasureWithTree`에서 `tree.lang != lang` 검증 후 error 반환.

## 8. Dependencies (의존성)

- **Blocked by**: SPEC-UTIL-001 Phase 3.1 구현 완료 (`release/v2.14.0` 머지 대기). v2.15 착수 시점에는 main 반영됨
- **Blocks**: 없음
- **Related**: SPEC-UTIL-001 §4.2 (complexity 패키지 원본 구조), SPEC-UTIL-004 (동일 `internal/hook/` 트리 성능 backlog)
