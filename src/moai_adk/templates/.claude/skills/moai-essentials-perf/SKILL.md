---

name: moai-essentials-perf
description: Performance optimization with profiling, bottleneck detection, and tuning strategies. Use when performing baseline performance reviews.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Performance Optimizer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (performance triage) |
| Trigger cues | Latency complaints, profiling requests, throughput bottleneck analysis. |

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- Loads when diagnosing performance regressions or planning optimization work.
- “Improve performance”, “Find slow parts”, “How to optimize?”
- “Profiling”, “Bottleneck”, “Memory leak”

## How it works

**Profiling Tools**:
- **Python**: cProfile, memory_profiler
- **TypeScript**: Chrome DevTools, clinic.js
- **Java**: JProfiler, VisualVM
- **Go**: pprof
- **Rust**: flamegraph, criterion

**Common Performance Issues**:
- **N+1 Query Problem**: Use eager loading/joins
- **Inefficient Loop**: O(n²) → O(n) with Set/Map
- **Memory Leak**: Remove event listeners, close connections

**Optimization Checklist**:
- [ ] Current performance benchmark
- [ ] Bottleneck identification
- [ ] Profiling data collected
- [ ] Algorithm complexity improved (O(n²) → O(n))
- [ ] Unnecessary operations removed
- [ ] Caching applied
- [ ] Async processing introduced
- [ ] Post-optimization benchmark
- [ ] Side effects checked

**Language-specific Optimizations**:
- **Python**: List comprehension, generators, @lru_cache
- **TypeScript**: Memoization, lazy loading, code splitting
- **Java**: Stream API, parallel processing
- **Go**: Goroutines, buffered channels
- **Rust**: Zero-cost abstractions, borrowing

**Performance Targets**:
- API response time: <200ms (P95)
- Page load time: <2s
- Memory usage: <512MB
- CPU usage: <70%

## Examples
```markdown
- 현재 diff를 점검하고 즉시 수정 가능한 항목을 나열합니다.
- 후속 작업은 TodoWrite로 예약합니다.
```

## Inputs
- 현재 작업 중인 코드/테스트/문서 스냅샷.
- 진행 중인 에이전트 상태 정보.

## Outputs
- 즉시 실행 가능한 체크리스트 또는 개선 제안.
- 다음 단계 실행 여부에 대한 권장 사항.

## Failure Modes
- 필요한 파일이나 테스트 결과를 찾지 못한 경우.
- 작업 범위가 과도하게 넓어 간단한 지원만으로 해결할 수 없을 때.

## Dependencies
- 주로 `tdd-implementer`, `quality-gate` 등과 연계해 사용합니다.

## References
- Google SRE. "The Four Golden Signals." https://sre.google/sre-book/monitoring-distributed-systems/ (accessed 2025-03-29).
- Dynatrace. "Application Performance Monitoring Best Practices." https://www.dynatrace.com/resources/ebooks/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Essentials 스킬의 입력/출력 정의를 정비했습니다.

## Works well with

- moai-essentials-refactor

## Best Practices
- 간단한 개선이라도 결과를 기록해 추적 가능성을 높입니다.
- 사람 검토가 필요한 항목을 명확히 표시하여 자동화와 구분합니다.
