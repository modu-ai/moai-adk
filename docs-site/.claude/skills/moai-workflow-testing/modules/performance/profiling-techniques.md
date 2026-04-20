# Profiling Techniques

> Sub-module: Detailed profiling methods and analysis techniques
> Parent: [Performance Optimization](../performance-optimization.md)
> Complexity: Advanced
> Time: 20+ minutes

## Overview

Comprehensive guide to profiling techniques for Python applications, covering CPU, memory, line-by-line, and real-time monitoring approaches.

## CPU Profiling

### cProfile Integration

```python
import cProfile
import pstats
import io

class PerformanceProfiler:
    def start_profiling(self, profile_types: List[str] = None):
        """Start performance profiling with specified types."""
        if profile_types is None:
            profile_types = ['cpu', 'memory']

        if 'cpu' in profile_types:
            self.profiler = cProfile.Profile()
            self.profiler.enable()

    def _analyze_cpu_profile(self) -> List[FunctionProfile]:
        """Analyze CPU profiling results."""
        if not self.profiler:
            return []

        s = io.StringIO()
        ps = pstats.Stats(self.profiler, stream=s)
        ps.sort_stats('cumulative')
        ps.print_stats()

        return self._parse_stats_output(s.getvalue())
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

### Tracemalloc Integration

```python
import tracemalloc
import memory_profiler

class PerformanceProfiler:
    def start_profiling(self, profile_types: List[str] = None):
        if 'memory' in profile_types:
            tracemalloc.start()
            self.memory_profiler = memory_profiler.Profile()

    def stop_profiling(self) -> Dict[str, Any]:
        results = {}
        
        if tracemalloc.is_tracing():
            current, peak = tracemalloc.get_traced_memory()
            tracemalloc.stop()
            results['memory_profile'] = {
                'current_mb': current / 1024 / 1024,
                'peak_mb': peak / 1024 / 1024
            }

        if self.memory_profiler:
            self.memory_profiler.disable()
            results['memory_line_profile'] = self._analyze_memory_profile()

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

```python
from line_profiler import LineProfiler

class PerformanceProfiler:
    def start_profiling(self, profile_types: List[str] = None):
        if 'line' in profile_types:
            self.line_profiler = LineProfiler()
            self.line_profiler.enable_by_count()

# Add specific functions to profile
profiler.line_profiler.add_function(expensive_function)
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

```python
import psutil
import threading
import time
from collections import deque

class RealTimeMonitor:
    def __init__(self, sampling_interval: float = 1.0):
        self.sampling_interval = sampling_interval
        self.is_monitoring = False
        self.monitor_thread = None
        self.snapshots = deque(maxlen=1000)
        self.callbacks = []
        self.alerts = []

    def start_monitoring(self):
        """Start real-time performance monitoring."""
        if self.is_monitoring:
            return

        self.is_monitoring = True
        self.monitor_thread = threading.Thread(
            target=self._monitor_loop, daemon=True
        )
        self.monitor_thread.start()

    def _monitor_loop(self):
        """Main monitoring loop."""
        process = psutil.Process()

        while self.is_monitoring:
            try:
                snapshot = PerformanceSnapshot(
                    timestamp=time.time(),
                    cpu_percent=process.cpu_percent(),
                    memory_mb=process.memory_info().rss / 1024 / 1024,
                    memory_percent=process.memory_percent(),
                    open_files=len(process.open_files()),
                    threads=process.num_threads(),
                    context_switches=process.num_ctx_switches().voluntary + 
                                   process.num_ctx_switches().involuntary
                )

                for callback in self.callbacks:
                    try:
                        custom_metrics = callback()
                        snapshot.custom_metrics.update(custom_metrics)
                    except Exception as e:
                        print(f"Custom metric callback error: {e}")

                self.snapshots.append(snapshot)
                self._check_alerts(snapshot)
                time.sleep(self.sampling_interval)

            except Exception as e:
                print(f"Monitoring error: {e}")
                time.sleep(self.sampling_interval)
```

### Performance Alerting

```python
def _check_alerts(self, snapshot: PerformanceSnapshot):
    """Check for performance alerts."""
    alerts = []

    if snapshot.cpu_percent > 90:
        alerts.append({
            'type': 'high_cpu',
            'message': f"High CPU usage: {snapshot.cpu_percent:.1f}%",
            'timestamp': snapshot.timestamp
        })

    if snapshot.memory_percent > 85:
        alerts.append({
            'type': 'high_memory',
            'message': f"High memory usage: {snapshot.memory_percent:.1f}%",
            'timestamp': snapshot.timestamp
        })

    if snapshot.open_files > 1000:
        alerts.append({
            'type': 'file_handle_leak',
            'message': f"High number of open files: {snapshot.open_files}",
            'timestamp': snapshot.timestamp
        })

    self.alerts.extend(alerts)
```

### Custom Metrics Integration

```python
def custom_metrics():
    return {
        'custom_counter': some_global_counter,
        'queue_size': len(some_queue),
        'cache_hit_rate': cache.hits / cache.requests if cache.requests > 0 else 0
    }

monitor.add_callback(custom_metrics)

# Get recent metrics
recent_snapshots = monitor.get_recent_snapshots(10)
avg_metrics = monitor.get_average_metrics(5)

print(f"Average CPU: {avg_metrics.get('avg_cpu_percent', 0):.1f}%")
print(f"Average Memory: {avg_metrics.get('avg_memory_mb', 0):.1f}MB")
```

## Bottleneck Detection

### CPU Bottleneck Detection

```python
async def _detect_cpu_bottlenecks(
    self, cpu_profiles: List[FunctionProfile],
    context7_patterns: Dict[str, Any] = None
) -> List[PerformanceBottleneck]:
    bottlenecks = []
    total_cpu_time = sum(p.cumulative_time for p in cpu_profiles)

    for profile in cpu_profiles:
        if profile.cumulative_time < 0.01:
            continue

        impact_score = profile.cumulative_time / max(total_cpu_time, 0.001)

        # Determine severity
        if impact_score > 0.5:
            severity = "critical"
        elif impact_score > 0.2:
            severity = "high"
        elif impact_score > 0.1:
            severity = "medium"
        else:
            severity = "low"

        optimization_type, suggestions, estimated_improvement = \
            await self._generate_cpu_optimization_suggestions(profile, context7_patterns)

        bottleneck = PerformanceBottleneck(
            function_name=profile.name,
            file_path=profile.file_path,
            line_number=profile.line_number,
            bottleneck_type="cpu",
            severity=severity,
            impact_score=impact_score,
            description=f"Function '{profile.name}' consumes {impact_score:.1%} of total CPU time",
            optimization_type=optimization_type,
            suggested_fixes=suggestions,
            estimated_improvement=estimated_improvement
        )
        bottlenecks.append(bottleneck)

    return bottlenecks
```

### Memory Bottleneck Detection

```python
async def _detect_memory_bottlenecks(
    self, memory_profile: Dict[str, Any],
    context7_patterns: Dict[str, Any] = None
) -> List[PerformanceBottleneck]:
    bottlenecks = []

    if 'memory_line_profile' in memory_profile:
        memory_by_function = memory_profile['memory_line_profile'].get('memory_by_function', {})

        if memory_by_function:
            max_memory = max(memory_by_function.values())

            for func_key, memory_usage in memory_by_function.items():
                if memory_usage < 1024 * 1024:  # 1MB
                    continue

                impact_score = memory_usage / max(max_memory, 1)

                if impact_score > 0.7:
                    severity = "critical"
                elif impact_score > 0.4:
                    severity = "high"
                elif impact_score > 0.2:
                    severity = "medium"
                else:
                    severity = "low"

                optimization_type, suggestions, estimated_improvement = \
                    await self._generate_memory_optimization_suggestions(
                        memory_usage, context7_patterns
                    )

                bottleneck = PerformanceBottleneck(
                    function_name=f"Function at {func_key}",
                    file_path=file_path,
                    line_number=line_number,
                    bottleneck_type="memory",
                    severity=severity,
                    impact_score=impact_score,
                    description=f"High memory usage: {memory_usage / 1024 / 1024:.1f}MB",
                    optimization_type=optimization_type,
                    suggested_fixes=suggestions,
                    estimated_improvement=estimated_improvement
                )
                bottlenecks.append(bottleneck)

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
- Limit snapshot history (deque with maxlen)
- Aggregate metrics over time windows
- Store only critical bottlenecks
- Implement data retention policies

## Analysis Techniques

### Performance Trend Analysis

```python
def get_average_metrics(self, duration_minutes: int = 5) -> Dict[str, float]:
    cutoff_time = time.time() - (duration_minutes * 60)
    recent_snapshots = [s for s in self.snapshots if s.timestamp >= cutoff_time]

    if not recent_snapshots:
        return {}

    return {
        'avg_cpu_percent': sum(s.cpu_percent for s in recent_snapshots) / len(recent_snapshots),
        'avg_memory_mb': sum(s.memory_mb for s in recent_snapshots) / len(recent_snapshots),
        'avg_memory_percent': sum(s.memory_percent for s in recent_snapshots) / len(recent_snapshots),
        'avg_open_files': sum(s.open_files for s in recent_snapshots) / len(recent_snapshots),
        'avg_threads': sum(s.threads for s in recent_snapshots) / len(recent_snapshots),
    }
```

### Performance Regression Detection

Monitor key metrics over time:
- Track average response times
- Compare against baseline performance
- Alert on significant degradation
- Maintain performance history

## Integration with Context7

### AI-Powered Analysis

```python
async def detect_bottlenecks(
    self, profile_results: Dict[str, Any],
    context7_patterns: Dict[str, Any] = None
) -> List[PerformanceBottleneck]:
    if not context7_patterns:
        return await self._rule_based_detection(profile_results)
    
    optimization_patterns = await self.context7.get_library_docs(
        context7_library_id="/performance/python-profiling",
        topic="advanced performance optimization patterns 2025",
        tokens=5000
    )
    
    return await self._ai_enhanced_detection(profile_results, optimization_patterns)
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
- Generate flame graphs from cProfile data
- Identify call stack hot paths
- Understand performance hierarchy
- Share visual insights

---

Sub-module: `modules/performance/profiling-techniques.md`
Parent: [Performance Optimization](../performance-optimization.md)
Version: 2.0.0
Last Updated: 2025-12-07
