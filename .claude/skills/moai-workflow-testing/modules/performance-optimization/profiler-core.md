# Performance Profiler Core

> Module: CPU, memory, and line profiling with analysis
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: the host language's CPU/memory/line profilers

## Core Implementation

The profiler wraps the host language's profilers (e.g. Go `pprof`/`runtime`, Python `cProfile`/`tracemalloc`/`line_profiler`/`memory_profiler`, Rust `flamegraph`/`heaptrack`, Node `clinic.js`). The data structures and orchestration below are language-neutral; the `_analyze_*` parsers consume each profiler's native output format.

### PerformanceProfiler Class

```text
enum PerformanceMetric:
    CPU_TIME, WALL_TIME, MEMORY_USAGE, MEMORY_PEAK,
    FUNCTION_CALLS, EXECUTION_COUNT, AVERAGE_TIME, MAX_TIME, MIN_TIME

record FunctionProfile:
    name:              text
    file_path:         text
    line_number:       int
    total_time:        float
    cumulative_time:   float
    call_count:        int
    per_call_time:     float
    cumulative_per_call_time: float
    memory_before:     float
    memory_after:      float
    memory_delta:      float
    optimization_suggestions: List<text>   # default []

class PerformanceProfiler:
    docs
    cpu_profiler = none
    memory_profiler = none
    line_profiler = none
    realtime_monitor = none
    profiles = {}
    bottlenecks = []

    start_profiling(profile_types = ["cpu", "memory"]):
        if "cpu"    in profile_types: cpu_profiler    = start_cpu_profiler()
        if "memory" in profile_types: memory_profiler = start_memory_profiler()
        if "line"   in profile_types: line_profiler   = start_line_profiler()

    stop_profiling():
        results = {}
        if cpu_profiler:
            cpu_profiler.stop()
            results.cpu_profile = analyze_cpu_profile()
        if memory_profiler_is_tracing():
            current, peak = traced_memory()
            stop_memory_profiler()
            results.memory_profile = { current_mb: current/MB, peak_mb: peak/MB }
        if memory_profiler:
            memory_profiler.stop()
            results.memory_line_profile = analyze_memory_profile()
        if line_profiler:
            line_profiler.stop()
            results.line_profile = analyze_line_profile()
        return results

    analyze_cpu_profile():
        # Read the CPU profiler's stats (cumulative-time sorted), parse each
        # function row into a FunctionProfile: name, file, line, total/cumulative
        # time, call count, per-call times. ncalls may be "primary/total" form —
        # use the primary count.
        function_profiles = []
        for row in cpu_stats_rows():
            profile = FunctionProfile(
                name=row.func_name, file_path=row.filename, line_number=row.line,
                total_time=row.tottime, cumulative_time=row.cumtime,
                call_count=row.ncalls,
                per_call_time=row.tottime / max(row.ncalls, 1),
                cumulative_per_call_time=row.cumtime / max(row.ncalls, 1),
                memory_before=0.0, memory_after=0.0, memory_delta=0.0)
            function_profiles.append(profile)
        return function_profiles

    analyze_memory_profile():
        stats = memory_profiler.get_stats()
        return {
            total_samples: len(stats),
            max_memory_usage: max(s.usage for s in stats) default 0,
            memory_by_function: group_memory_by_function(stats)
        }

    group_memory_by_function(stats):
        by_fn = {}    # "file:line" -> total bytes
        for s in stats: by_fn[s.filename + ":" + s.line] += s.usage
        return by_fn

    analyze_line_profile():
        stats = line_profiler.get_stats()
        return { timings: stats.timings, unit: stats.unit }
```

## Usage Examples

```text
# Initialize the performance profiler
profiler = PerformanceProfiler(docs_client=docs)

# Example function to profile
expensive_function(n):
    result = []
    for i in range(n):
        temp = []
        for j in range(i): temp.append(j * j)
        result.extend(temp)
    return result

# Start profiling
profiler.start_profiling(["cpu", "memory", "line"])
# Attach the line profiler to a specific function, if line profiling is active
if profiler.line_profiler: profiler.line_profiler.add_function(expensive_function)

# Run the code to be profiled
result = expensive_function(1000)

# Stop profiling and read results
profile_results = profiler.stop_profiling()
print("CPU Profile: " + len(profile_results.cpu_profile default []) + " functions")
print("Memory Peak: " + (profile_results.memory_profile default {}).peak_mb default 0 + " MB")
```

## Best Practices

1. **Profile Types**: Start with CPU and memory profiling, add line profiling for specific functions
2. **Baseline Measurement**: Always profile before optimization
3. **Realistic Workloads**: Profile with production-like data and patterns
4. **Multiple Runs**: Profile multiple times to account for variability
5. **Statistical Significance**: Ensure sufficient execution time for accurate measurements

---

Related: [Real-Time Monitoring](./real-time-monitoring.md) | [Bottleneck Detection](./bottleneck-detection.md)
