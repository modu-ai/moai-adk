---

name: moai-alfred-performance-optimizer
description: Performance analysis and optimization suggestions with profiling, bottleneck detection, and language-specific optimizations. Use when planning performance improvements or regressions checks.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Performance Optimizer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run performance lane |
| Trigger cues | Performance regressions surfaced in Alfred workflows, profiling or tuning requests. |

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- Activates when Alfred must diagnose or remediate performance bottlenecks.
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

## Best Practices
- 사용자에게 보여주는 문구는 TUI/보고서용 표현으로 작성합니다.
- 도구 실행 시 명령과 결과 요약을 함께 기록합니다.

## Examples
```markdown
- /alfred 커맨드 내부에서 이 스킬을 호출해 보고서를 생성합니다.
- Completion Report에 요약을 추가합니다.
```

## Inputs
- MoAI-ADK 프로젝트 맥락 (`.moai/project/`, `.claude/` 템플릿 등).
- 사용자 명령 또는 상위 커맨드에서 전달한 파라미터.

## Outputs
- Alfred 워크플로우에 필요한 보고서, 체크리스트 또는 추천 항목.
- 후속 서브 에이전트 호출을 위한 구조화된 데이터.

## Failure Modes
- 필수 입력 문서가 없거나 권한이 제한된 경우.
- 사용자 승인 없이 파괴적인 변경이 요구될 때.

## Dependencies
- cc-manager, project-manager 등 상위 에이전트와 협력이 필요합니다.

## References
- Google SRE. "The Four Golden Signals." https://sre.google/sre-book/monitoring-distributed-systems/ (accessed 2025-03-29).
- Dynatrace. "Application Performance Monitoring Best Practices." https://www.dynatrace.com/resources/ebooks/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-code-reviewer
- alfred-debugger-pro
