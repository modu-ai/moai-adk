# Bottleneck Detection

> Module: Performance bottleneck detection and analysis
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: WebSearch/WebFetch (optional)

## Core Implementation

### Bottleneck Detection System

```text
enum OptimizationType:
    ALGORITHM_IMPROVEMENT, CACHING, CONCURRENCY, MEMORY_OPTIMIZATION,
    IO_OPTIMIZATION, DATABASE_OPTIMIZATION, DATA_STRUCTURE_CHANGE

record PerformanceBottleneck:
    function_name:        text
    file_path:            text
    line_number:          int
    bottleneck_type:      text     # "cpu", "memory", "io", "algorithm"
    severity:             text     # "low", "medium", "high", "critical"
    impact_score:         float    # 0.0 to 1.0
    description:          text
    metrics:              Map<text, float>
    optimization_type:    OptimizationType
    suggested_fixes:      List<text>
    estimated_improvement:text
    code_snippet:         text

class BottleneckDetector(profiler):
    detect_bottlenecks(profile_results, docs_patterns = none):
        bottlenecks = []
        if "cpu_profile"      in profile_results: bottlenecks.extend(detect_cpu_bottlenecks(profile_results.cpu_profile, docs_patterns))
        if "memory_profile"   in profile_results: bottlenecks.extend(detect_memory_bottlenecks(profile_results.memory_profile, docs_patterns))
        if "realtime_metrics" in profile_results: bottlenecks.extend(detect_realtime_bottlenecks(profile_results.realtime_metrics, docs_patterns))
        sort bottlenecks descending by impact_score
        return bottlenecks

    detect_cpu_bottlenecks(cpu_profiles, docs_patterns = none):
        bottlenecks = []
        total_cpu_time = sum(p.cumulative_time for p in cpu_profiles)
        for profile in cpu_profiles:
            if profile.cumulative_time < 0.01: continue
            impact = profile.cumulative_time / max(total_cpu_time, 0.001)
            severity = critical if impact > 0.5 else high if impact > 0.2 else medium if impact > 0.1 else low
            (opt_type, suggestions, est) = generate_cpu_optimization_suggestions(profile, docs_patterns)
            bottlenecks.append(PerformanceBottleneck(
                function_name=profile.name, file_path=profile.file_path, line_number=profile.line_number,
                bottleneck_type="cpu", severity=severity, impact_score=impact,
                description="Function '" + profile.name + "' consumes " + pct(impact) + " of CPU time",
                metrics={ cumulative_time, total_time, call_count, per_call_time },
                optimization_type=opt_type, suggested_fixes=suggestions,
                estimated_improvement=est, code_snippet=get_code_snippet(profile.file_path, profile.line_number)))
        return bottlenecks

    detect_memory_bottlenecks(memory_profile, docs_patterns = none):
        bottlenecks = []
        if memory_profile has memory_by_function:
            by_fn = memory_profile.memory_by_function
            max_mem = max(by_fn.values())
            for (func_key, usage) in by_fn:
                if usage < 1*MB: continue
                impact = usage / max(max_mem, 1)
                severity = critical if impact > 0.7 else high if impact > 0.4 else medium if impact > 0.2 else low
                (file_path, line_number) = split_location(func_key)
                (opt_type, suggestions, est) = generate_memory_optimization_suggestions(usage, docs_patterns)
                bottlenecks.append(PerformanceBottleneck(
                    function_name="Function at " + func_key, file_path, line_number,
                    bottleneck_type="memory", severity, impact_score=impact,
                    description="High memory usage: " + (usage/MB) + "MB",
                    metrics={ memory_usage_mb: usage/MB, impact_score: impact },
                    optimization_type=opt_type, suggested_fixes=suggestions,
                    estimated_improvement=est, code_snippet=get_code_snippet(file_path, line_number)))
        return bottlenecks

    detect_realtime_bottlenecks(realtime_metrics, docs_patterns = none):
        bottlenecks = []
        avg_cpu = realtime_metrics.avg_cpu_percent default 0
        if avg_cpu > 80:
            bottlenecks.append(PerformanceBottleneck(
                function_name="System CPU Usage", file_path="system", line_number=0,
                bottleneck_type="cpu", severity=high if avg_cpu > 90 else medium,
                impact_score=avg_cpu/100, description="High average CPU usage: " + avg_cpu + "%",
                metrics={ avg_cpu_percent: avg_cpu }, optimization_type=CONCURRENCY,
                suggested_fixes=["Implement parallel processing","Optimize algorithms","Add caching"],
                estimated_improvement="20-50% CPU reduction", code_snippet="# system-wide optimization"))
        avg_memory = realtime_metrics.avg_memory_percent default 0
        if avg_memory > 75:
            bottlenecks.append(PerformanceBottleneck(
                function_name="System Memory Usage", file_path="system", line_number=0,
                bottleneck_type="memory", severity=high if avg_memory > 85 else medium,
                impact_score=avg_memory/100, description="High average memory usage: " + avg_memory + "%",
                metrics={ avg_memory_percent: avg_memory }, optimization_type=MEMORY_OPTIMIZATION,
                suggested_fixes=["Implement memory pooling","Use streaming/generators","Optimize data structures"],
                estimated_improvement="30-60% memory reduction", code_snippet="# system-wide memory optimization"))
        return bottlenecks

    generate_cpu_optimization_suggestions(profile, docs_patterns = none):
        if profile.call_count > 10000 and profile.per_call_time > 0.001:
            return (CACHING, ["Memoize expensive calls","Add an LRU cache for hot functions"], "50-90% for repeated calls")
        if profile.cumulative_time > 1.0 and profile.call_count > 100:
            return (ALGORITHM_IMPROVEMENT, ["Analyze algorithm complexity","Look for O(n^2) or worse","Use more efficient data structures"], "20-80% depending on algorithm")
        if profile.call_count < 10 and profile.cumulative_time > 0.5:
            return (CONCURRENCY, ["Parallelize long-running operations","Use async processing","Use a worker pool for CPU-bound tasks"], "30-70% with proper concurrency")
        return (ALGORITHM_IMPROVEMENT, ["Profile line-by-line","Check for unnecessary loops","Optimize string/regex operations"], "10-40% with micro-optimizations")

    generate_memory_optimization_suggestions(memory_usage, docs_patterns = none):
        if memory_usage > 100*MB:
            return (MEMORY_OPTIMIZATION, ["Stream large datasets","Use generators","Process in chunks"], "60-90% memory reduction")
        if memory_usage > 10*MB:
            return (MEMORY_OPTIMIZATION, ["Use memory-efficient structures","Pool frequently-allocated objects","Use contiguous arrays for numerical data"], "30-60% memory reduction")
        return (MEMORY_OPTIMIZATION, ["Release unused objects","Use weak references for caches","Avoid circular references"], "10-30% memory reduction")

    get_code_snippet(file_path, line_number, context_lines = 5):
        try:
            lines = read_lines(file_path)
            start = max(0, line_number - context_lines - 1)
            end   = min(len(lines), line_number + context_lines)
            out = []
            for i in start..end:
                marker = ">>> " if i == line_number - 1 else "    "
                out.append(marker + (i+1) + ": " + lines[i])
            return join(out, "\n")
        except:
            return "// Code not available for " + file_path + ":" + line_number
```

## Usage Examples

```text
# Detect bottlenecks from profiling results
detector = BottleneckDetector(profiler)
bottlenecks = detector.detect_bottlenecks(profile_results)
print("Found " + len(bottlenecks) + " performance bottlenecks:")
for bottleneck in bottlenecks[:5]:
    print("Bottleneck: " + bottleneck.function_name)
    print("  Type: " + bottleneck.bottleneck_type + "  Severity: " + bottleneck.severity)
    print("  Impact: " + bottleneck.impact_score + "  " + bottleneck.description)
    print("  Optimization: " + bottleneck.optimization_type)
    for fix in bottleneck.suggested_fixes: print("    - " + fix)
```

## Best Practices

1. **Severity Prioritization**: Focus on critical and high severity bottlenecks first
2. **Impact Score**: Use impact scores to quantify optimization potential
3. **Context-Aware**: Consider codebase context when suggesting optimizations
4. **Incremental**: Address one bottleneck at a time to measure impact
5. **Validation**: Always validate optimizations with performance tests

---

Related: [Profiler Core](./profiler-core.md) | [Optimization Plan](./optimization-plan.md)
