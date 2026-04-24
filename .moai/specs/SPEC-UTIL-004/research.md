# SPEC-UTIL-004 Research Notes

## §1. 배경 (v2.14 Review Finding)

v2.14.0 릴리스 플랜 Phase 3.2 완료 후 `manager-quality` 멀티관점 코드 리뷰에서 다음 항목이 **Warning — PERFORMANCE** 등급으로 제기되었다.

- **Location**: `internal/hook/security/ast_grep.go:207-215`
- **Category**: 성능 — 메모리 압박
- **Severity**: Warning (v2.14 blocking은 아니며 v2.15 backlog로 defer)
- **Reviewer Note**: "세마포어 획득이 goroutine 내부에서 일어남 → N개 파일이면 N개 goroutine 즉시 생성, 세마포어는 `NumCPU()*2`만 동시 실행. 대용량 파일셋(수천 개)에서 goroutine object 메모리 압박."
- **Reference Pattern**: MX validator(`internal/hook/mx/validator.go:513`)는 `sem <- struct{}{}` 획득이 `go func()` 호출 **이전**에 실행되어 동시 goroutine 수가 세마포어 용량과 동일.

## §2. Baseline (현재 동작)

### 2.1 현재 ScanMultiple 패턴 (simplified)

```
for _, file := range files {
    wg.Add(1)
    go func(f string) {
        sem <- struct{}{}
        defer func() { <-sem }()
        defer wg.Done()
        res, err := Scan(ctx, f, rules)
        // ...
    }(file)
}
wg.Wait()
```

### 2.2 MX validator 차이 (post-SPEC-UTIL-001)

```
for _, file := range files {
    sem <- struct{}{}  // outside go func
    wg.Add(1)
    go func(f string) {
        defer func() { <-sem }()
        defer wg.Done()
        analyzeFile(f, ...)
    }(file)
}
wg.Wait()
```

### 2.3 메모리 영향 추정

- Go goroutine 초기 stack: ~2KB (runtime/runtime2.go `_StackMin`)
- 10,000 파일 × 2KB = ~20MB 추가 heap pressure (GC 스캔 대상)
- 세마포어 블록 시 goroutine은 gopark → M 반환하지만 goroutine object는 heap에 유지
- Windows post-tool-use 훅에서 10k+ 파일 프로젝트의 저장 빈도가 높을 경우 누적 부담

### 2.4 AC-UTIL-002-07 검증 gap

SPEC-UTIL-002의 `TestScanMultiple_SemaphoreBound` 테스트는 `Scan` 내부에서 atomic counter를 increment/decrement하여 **동시 실행 중인 `Scan` 호출 수**의 상한을 확인한다. 그러나 실제 **생성된 goroutine 수**(세마포어에서 블록 중인 것 포함)는 측정하지 않으므로, pre-fix 패턴도 이 테스트를 통과한다.

## §3. Approach

세마포어 획득 위치를 `go func()` 호출 **외부로** 이동한다. 이렇게 하면 for 루프(생산자)가 세마포어 슬롯 확보 시까지 블록되어, 동시 생성 goroutine 수가 세마포어 용량(`NumCPU()*2`)으로 자동 제한된다. `defer <-sem`은 goroutine 내부에 유지되어 `Scan` 완료 후 해제 타이밍은 변경되지 않는다.

### 3.1 변경 Diff 개요

```
 for _, file := range files {
+    sem <- struct{}{}
     wg.Add(1)
     go func(f string) {
-        sem <- struct{}{}
         defer func() { <-sem }()
         defer wg.Done()
         // ...
     }(file)
 }
```

### 3.2 검증 전략

- **Peak goroutine counter shim**: `ScanMultiple` 진입 전 `runtime.NumGoroutine()` baseline 기록, atomic counter로 스폰/종료 카운트, 1000-file fixture로 peak 관찰
- **Backward-compat smoke**: 기존 `TestScanMultiple_SemaphoreBound` 회귀 없음 확인
- **Benchmark regression guard**: `BenchmarkScanMultiple_1kFiles` 처리량 ±5% 이내 유지

## §4. References

- v2.14 `manager-quality` 멀티관점 리뷰 보고서 (Warning — PERFORMANCE 섹션)
- SPEC-UTIL-002 §3 (ast-grep 통합 스코프)
- SPEC-UTIL-001 §4.2 (MX validator worker pool 세마포어 계약, `validator.go:513`)
- Go runtime `sync.WaitGroup` + buffered channel semaphore idiom (stdlib test 예시 `runtime/pprof/pprof_test.go`)
