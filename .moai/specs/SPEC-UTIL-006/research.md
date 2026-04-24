# SPEC-UTIL-006 Research Notes

## §1. 배경 (v2.14 Review Finding)

v2.14.0 Phase 3.1 완료 후 `manager-quality` 멀티관점 코드 리뷰에서 다음 항목이 **Warning — QUALITY + PERFORMANCE** 등급으로 제기되었다.

- **Location**: `internal/hook/mx/validator.go:196-223` (analyzeFile 함수)
- **Category**: 성능 — 중복 파싱 + 아키텍처 관심사 (단일 파일 내 AST를 매번 재생성)
- **Severity**: Warning (v2.14 blocking은 아니며 v2.15 backlog로 defer)
- **Reviewer Note**: "`analyzeFile`이 한 파일 내 모든 exported 함수에 대해 `complexity.Measure`를 순차 호출. 각 호출이 tree-sitter 파싱 O(파일크기)을 반복 수행 → 20 함수/파일이면 동일 파일을 20번 파싱."
- **Approach Hint**: "파일당 단 한 번 파싱 후 함수별 쿼리만 반복."

## §2. Baseline (호출 경로 & 비용)

### 2.1 현재 호출 경로

```
ValidateFiles (validator.go:500+)
  → worker goroutine (세마포어 bound)
    → analyzeFile (validator.go:196)
      → extractFunctions  (regex 기반, 빠름)
      → for each exported func:
          complexity.Measure(lang, content, fn.Name, fn.StartLine)
            → measure_cgo.go:
                sitter.NewParser()     # 매번 생성
                parser.SetLanguage(...)
                parser.Parse(nil, content)   # 파일 전체 파싱 O(n)
                // ... 쿼리 매칭 ...
                tree.Close()
```

### 2.2 측정 기반 비용 추정 (SPEC-UTIL-001 `complexity_test.go` 기준)

| 파일 크기 | 함수 수 | 파싱 1회 | 현재 (N파싱) | 목표 (1파싱+N쿼리) |
|-----------|---------|----------|--------------|---------------------|
| 5 KB | 10 | ~2ms | ~20ms | ~4-5ms |
| 10 KB | 20 | ~4ms | ~80ms | ~8-10ms |
| 50 KB | 50 | ~20ms | ~1000ms | ~30-40ms |

- 파싱은 파일 크기 O(n) 지배적, 쿼리 매칭은 tree walk O(nodes)로 상대적으로 저렴
- 큰 파일(50KB+)에서 비선형 악화, 리포지토리 전체 MX 스캔 시 심각

### 2.3 1 MiB content cap (SPEC-UTIL-001 §4.2 확정)

- `Measure`는 `len(content) > 1<<20` 시 `Result{Supported: false}` + sentinel error 반환
- `ParseFile`은 이 검증을 1회 수행, `MeasureWithTree`는 pre-validated 전제

## §3. Approach

### 3.1 API 설계

```
// complexity/parse.go (new)
type Tree struct {
    lang    string
    native  *sitter.Tree   // cgo opaque handle
    content []byte         // for span extraction
}

func ParseFile(lang string, content []byte) (*Tree, error) {
    if len(content) > 1<<20 {
        return nil, ErrContentTooLarge
    }
    // ... sitter.NewParser + Parse ...
    t := &Tree{lang: lang, native: nativeTree, content: content}
    runtime.SetFinalizer(t, func(t *Tree) { t.native.Close() })
    return t, nil
}

func (t *Tree) Close() { ... }  // deterministic cleanup

// complexity/measure.go (new public API)
func MeasureWithTree(lang string, tree *Tree, funcName string, startLine int) (Result, error) {
    if tree.lang != lang {
        return Result{}, ErrLangMismatch
    }
    // ... 기존 쿼리 매칭 로직 재사용, parser setup 생략 ...
}

// complexity/measure.go (existing, now delegating)
func Measure(lang string, content []byte, funcName string, startLine int) (Result, error) {
    t, err := ParseFile(lang, content)
    if err != nil {
        return Result{}, err
    }
    defer t.Close()
    return MeasureWithTree(lang, t, funcName, startLine)
}
```

### 3.2 analyzeFile 리팩토링

```
func analyzeFile(path string, ...) ([]Violation, error) {
    content, _ := os.ReadFile(path)
    funcs := extractFunctions(content)

    lang := detectLang(path)
    tree, err := complexity.ParseFile(lang, content)
    if err == complexity.ErrContentTooLarge {
        // skip complexity, keep fan-in checks
    } else if err != nil {
        return nil, err
    } else {
        defer tree.Close()
    }

    for _, fn := range funcs {
        if tree != nil {
            res, _ := complexity.MeasureWithTree(lang, tree, fn.Name, fn.StartLine)
            // ...
        }
    }
}
```

### 3.3 Tree 객체 lifecycle

- **Primary**: `defer tree.Close()` in `analyzeFile` — deterministic cleanup
- **Safety net**: `runtime.SetFinalizer` — defer 누락이나 panic 시에도 native 메모리 회수
- **Re-Close idempotent**: `Close()`는 double-call 안전해야 함 (`atomic.Bool` flag 또는 nil 체크)

## §4. Measurement Targets

- **Baseline**: 20 함수 × 10KB Go fixture, pre-cache 구조 기준 ~200ms (벤치마크 결과는 런타임 측정 필요)
- **Target**: 파일당 1회 파싱으로 ~40ms 이하 (80% 감소)
- **Validation**: `BenchmarkAnalyzeFile_BeforeCache` vs `BenchmarkAnalyzeFile_WithCache` 동일 fixture, `go test -bench=AnalyzeFile -benchmem` 결과 비교

## §5. References

- v2.14 `manager-quality` 멀티관점 리뷰 보고서 (Warning — QUALITY + PERFORMANCE 섹션)
- SPEC-UTIL-001 §4.2 (tree-sitter complexity 패키지 원본 + 1 MiB cap 규약)
- SPEC-UTIL-001 §4.5 (5+11 언어 지원 범위 — 본 SPEC에서도 동일 유지)
- `measure_cgo.go` 현재 구조 (SPEC-UTIL-001 구현, `release/v2.14.0` 머지 대기)
- `github.com/smacker/go-tree-sitter` Tree/Parser lifecycle 문서
