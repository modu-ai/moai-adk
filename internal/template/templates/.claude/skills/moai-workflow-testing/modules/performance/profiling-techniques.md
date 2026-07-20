# Profiling Techniques

> Sub-module: Detailed profiling methods and analysis techniques
> Parent: [Performance Optimization](../performance-optimization.md)
> Complexity: Advanced
> Time: 20+ minutes

## Overview

Comprehensive guide to profiling techniques, covering CPU, memory, line-by-line, and real-time monitoring approaches. The tool names vary by language — the metrics and interpretation are universal. Map each section to your language's profiler (Go pprof, Python cProfile/tracemalloc, Rust flamegraph, Java async-profiler, Node clinic.js, .NET dotnet-trace).

## CPU Profiling

### Profiler Integration

```text
class PerformanceProfiler:
    start_profiling(profile_types = ["cpu", "memory"]):
        if "cpu" in profile_types:
            # Start the language's CPU profiler (e.g. Go pprof, Python cProfile,
            # Rust perf/flamegraph, Java async-profiler).
            cpu_profiler.start()

    analyze_cpu_profile():
        if no cpu_profiler: return []
        stats = cpu_profiler.snapshot_sorted_by(cumulative_time)
        return parse_stats(stats)
```

### CPU Profile Analysis

Key Metrics:
- **cumulative time**: Total time spent in function including called functions
- **total time**: Time spent in function excluding called functions
- **call count**: Number of times function was called
- **percall time**: Average time per call

Interpretation Guide:
- High cumulative time + low total time: Function calling slow sub-functions
- High total time: Function itself has expensive operations
- High call count + high percall: Consider optimization or caching

## Memory Profiling

### Memory Profiler Integration

```text
class PerformanceProfiler:
    start_profiling(profile_types):
        if "memory" in profile_types:
            # Start the language's memory tracer (e.g. Go runtime.MemProfile,
            # Python tracemalloc/memory_profiler, Rust heaptrack, Java JFR).
            memory_tracer.start()

    stop_profiling():
        results = {}
        if memory_tracer.is_tracing():
            current, peak = memory_tracer.traced_memory()
            memory_tracer.stop()
            results.memory_profile = { current_mb: current/MB, peak_mb: peak/MB }
        return results
```

### Memory Analysis Techniques

Memory Profile Metrics:
- **current memory**: Current memory allocation
- **peak memory**: Maximum memory during profiling
- **memory by function**: Per-function memory usage

Memory Leak Detection:
- Compare snapshots before and after operations
- Look for continuously growing allocations
- Check for circular references preventing GC

## Line-by-Line Profiling

### Line Profiler Usage

```text
class PerformanceProfiler:
    start_profiling(profile_types):
        if "line" in profile_types:
            # Use a line-level profiler (e.g. Python line_profiler, Rust cargo-llvm-lines,
            # Java JFR method profiler, JS CPU profiler with line info).
            line_profiler.start()

# Add specific functions to profile
line_profiler.add_function(expensive_function)
```

### Line Profile Interpretation

Line Profile Metrics:
- **Hits**: Number of times line executed
- **Time**: Total time spent on line
- **Per Hit**: Average time per execution

Optimization Targets:
- Lines with high time and low hits: Expensive operations
- Lines in loops with high hits: Consider moving outside loop
- Lines with high per-hit time: Algorithmic improvements

## Real-time Monitoring

### RealTimeMonitor Class

```text
class RealTimeMonitor(sampling_interval = 1.0):
    sampling_interval
    is_monitoring = false
    snapshots = bounded_deque(maxlen=1000)
    callbacks = []
    alerts = []

    start_monitoring():
        if is_monitoring: return
        is_monitoring = true
        spawn_background_task(monitor_loop())   # daemon thread / goroutine / async task

    monitor_loop():
        while is_monitoring:
            try:
                snapshot = PerformanceSnapshot(
                    timestamp=now(),
                    cpu_percent=process.cpu_percent(),        # via OS/process API
                    memory_mb=process.rss() / MB,
                    memory_percent=process.memory_percent(),
                    open_files=len(process.open_files()),
                    threads=process.num_threads(),
                    context_switches=process.context_switches())

                for callback in callbacks:
                    try: snapshot.custom_metrics.update(callback())
                    except e: log("Custom metric callback error: " + e)

                snapshots.append(snapshot)
                check_alerts(snapshot)
                sleep(sampling_interval)
            except e:
                log("Monitoring error: " + e)
                sleep(sampling_interval)
```

### Performance Alerting

```text
check_alerts(snapshot):
    if snapshot.cpu_percent > 90:
        alerts.append({ type: "high_cpu", message: "High CPU usage: " + snapshot.cpu_percent })
    if snapshot.memory_percent > 85:
        alerts.append({ type: "high_memory", message: "High memory usage: " + snapshot.memory_percent })
    if snapshot.open_files > 1000:
        alerts.append({ type: "file_handle_leak", message: "High open files: " + snapshot.open_files })
```

### Custom Metrics Integration

```text
custom_metrics():
    return {
        custom_counter: some_global_counter,
        queue_size:     len(some_queue),
        cache_hit_rate: cache.hits / cache.requests if cache.requests > 0 else 0
    }

monitor.add_callback(custom_metrics)

# Get recent metrics
recent_snapshots = monitor.get_recent_snapshots(10)
avg_metrics      = monitor.get_average_metrics(5)
```

## Bottleneck Detection

### CPU Bottleneck Detection

```text
detect_cpu_bottlenecks(cpu_profiles, docs_patterns = none):
    bottlenecks = []
    total_cpu_time = sum(p.cumulative_time for p in cpu_profiles)

    for profile in cpu_profiles:
        if profile.cumulative_time < 0.01: continue
        impact = profile.cumulative_time / max(total_cpu_time, 0.001)

        severity = critical if impact > 0.5
                   else high if impact > 0.2
                   else medium if impact > 0.1
                   else low

        (opt_type, suggestions, est_improvement) =
            generate_cpu_optimization_suggestions(profile, docs_patterns)

        bottlenecks.append(PerformanceBottleneck(
            function_name=profile.name,
            file_path=profile.file_path, line_number=profile.line_number,
            bottleneck_type="cpu", severity=severity, impact_score=impact,
            description="Function '" + profile.name + "' consumes " + pct(impact) + " of CPU",
            optimization_type=opt_type, suggested_fixes=suggestions,
            estimated_improvement=est_improvement))
    return bottlenecks
```

### Memory Bottleneck Detection

```text
detect_memory_bottlenecks(memory_profile, docs_patterns = none):
    bottlenecks = []
    if memory_profile has memory_by_function:
        by_fn = memory_profile.memory_by_function
        max_mem = max(by_fn.values())
        for (func_key, usage) in by_fn:
            if usage < 1*MB: continue
            impact = usage / max(max_mem, 1)
            severity = critical if impact > 0.7
                       else high if impact > 0.4
                       else medium if impact > 0.2
                       else low
            (opt_type, suggestions, est) = generate_memory_optimization_suggestions(usage, docs_patterns)
            bottlenecks.append(PerformanceBottleneck(
                function_name="Function at " + func_key,
                bottleneck_type="memory", severity=severity, impact_score=impact,
                description="High memory: " + (usage/MB) + "MB",
                optimization_type=opt_type, suggested_fixes=suggestions,
                estimated_improvement=est))
    return bottlenecks
```

## Profiling Best Practices

### Selecting Profiling Type

CPU Profiling Use Cases:
- Identifying slow functions
- Analyzing call frequency
- Finding algorithmic complexity issues
- Optimizing hot paths

Memory Profiling Use Cases:
- Detecting memory leaks
- Reducing memory footprint
- Optimizing data structures
- Analyzing allocation patterns

Line Profiling Use Cases:
- Detailed function analysis
- Identifying slow lines within functions
- Optimizing loops and iterations
- Understanding time distribution

Real-time Monitoring Use Cases:
- Production performance tracking
- Long-running application monitoring
- Performance regression detection
- Resource usage alerting

### Profiling Overhead

Expected Overhead by Type:
- CPU profiling: 5-15% performance impact
- Memory profiling: 10-30% performance impact
- Line profiling: 20-50% performance impact
- Real-time monitoring: <1% at 1-second intervals

Mitigation Strategies:
- Profile for representative time periods
- Use sampling for long-running processes
- Profile in production-like environments
- Consider asynchronous profiling for production

### Data Collection

Optimal Sampling:
- CPU profiling: 30-60 seconds minimum
- Memory profiling: Full operation cycles
- Line profiling: Specific function executions
- Real-time monitoring: Continuous with configurable intervals

Data Management:
- Limit snapshot history (bounded deque)
- Aggregate metrics over time windows
- Store only critical bottlenecks
- Implement data retention policies

## Analysis Techniques

### Performance Trend Analysis

```text
get_average_metrics(duration_minutes = 5):
    cutoff = now() - duration_minutes*60
    recent = [s for s in snapshots if s.timestamp >= cutoff]
    if recent is empty: return {}
    return {
        avg_cpu_percent:    mean(s.cpu_percent    for s in recent),
        avg_memory_mb:      mean(s.memory_mb      for s in recent),
        avg_memory_percent: mean(s.memory_percent for s in recent),
        avg_open_files:     mean(s.open_files     for s in recent),
        avg_threads:        mean(s.threads        for s in recent)
    }
```

### Performance Regression Detection

Monitor key metrics over time:
- Track average response times
- Compare against baseline performance
- Alert on significant degradation
- Maintain performance history

## Integration with Documentation

### AI-Powered Analysis

```text
detect_bottlenecks(profile_results, docs_patterns = none):
    if docs_patterns is none:
        return rule_based_detection(profile_results)

    optimization_patterns = docs.get_library_docs(
        topic="advanced performance optimization patterns",
        tokens=5000)
    return ai_enhanced_detection(profile_results, optimization_patterns)
```

## Advanced Techniques

### Comparative Profiling

Before/After Analysis:
- Profile before optimization
- Apply optimization
- Profile after optimization
- Compare metrics quantitatively

### Statistical Analysis

Confidence Intervals:
- Run multiple profiling sessions
- Calculate mean and standard deviation
- Establish confidence intervals
- Validate statistical significance

### Flame Graphs

Visualization:
- Generate flame graphs from profiler data
- Identify call stack hot paths
- Understand performance hierarchy
- Share visual insights

---

Sub-module: `modules/performance/profiling-techniques.md`
Parent: [Performance Optimization](../performance-optimization.md)
Version: 2.0.0
Last Updated: 2025-12-07
