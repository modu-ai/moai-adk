# Performance Optimization with Real-time Profiling

> Module: Real-time performance profiling, bottleneck detection, and optimization strategies
> Complexity: Advanced
> Time: 30+ minutes
> Dependencies: Python 3.8+, cProfile, memory_profiler, psutil, Context7 MCP, asyncio

## Overview

Performance optimization module providing comprehensive profiling, bottleneck detection, and AI-powered optimization suggestions for Python applications.

### Core Capabilities

- **Real-time Monitoring**: Continuous performance tracking with alerting
- **Multi-dimensional Profiling**: CPU, memory, and line-level profiling
- **Intelligent Detection**: Automatic bottleneck identification with severity scoring
- **AI-Powered Suggestions**: Context7-based optimization recommendations
- **Comprehensive Planning**: Structured optimization plans with risk assessment

## Module Structure

### 1. Real-Time Monitoring
[real-time-monitoring.md](./performance-optimization/real-time-monitoring.md)

Continuous performance monitoring system with snapshot collection and alerting.

Key Features:
- Configurable sampling intervals (0.5-2.0 seconds)
- CPU, memory, file handle, and thread monitoring
- Custom metric callbacks for domain-specific tracking
- Automatic alert generation for threshold violations
- Rolling snapshot buffer with configurable size

Usage:
```python
monitor = RealTimeMonitor(sampling_interval=0.5)
monitor.start_monitoring()
monitor.add_callback(custom_metrics)
# ... application code ...
monitor.stop_monitoring()
metrics = monitor.get_average_metrics(5)
```

### 2. Profiler Core
[profiler-core.md](./performance-optimization/profiler-core.md)

CPU, memory, and line profiling infrastructure with statistical analysis.

Key Features:
- cProfile integration for CPU profiling
- tracemalloc and memory_profiler for memory tracking
- line_profiler for function-level timing
- Comprehensive profile analysis and parsing
- Function-level performance statistics

Usage:
```python
profiler = PerformanceProfiler(context7_client=context7)
profiler.start_profiling(['cpu', 'memory', 'line'])
# ... code to profile ...
results = profiler.stop_profiling()
```

### 3. Bottleneck Detection
[bottleneck-detection.md](./performance-optimization/bottleneck-detection.md)

Automatic detection and classification of performance bottlenecks.

Key Features:
- CPU bottleneck detection with impact scoring
- Memory leak identification and tracking
- Real-time metric analysis
- Severity classification (critical, high, medium, low)
- Context-aware optimization suggestions

Usage:
```python
detector = BottleneckDetector(profiler)
bottlenecks = await detector.detect_bottlenecks(profile_results)
for bottleneck in bottlenecks[:5]:
    print(f"{bottleneck.function_name}: {bottleneck.severity}")
    print(f"  Impact: {bottleneck.impact_score:.2f}")
    print(f"  Fixes: {bottleneck.suggested_fixes}")
```

### 4. Optimization Planning
[optimization-plan.md](./performance-optimization/optimization-plan.md)

Comprehensive optimization plan generation with execution ordering.

Key Features:
- Bottleneck prioritization by impact and complexity
- Execution order optimization for minimal risk
- Implementation complexity assessment
- Risk level evaluation (low, medium, high)
- Prerequisite identification
- Validation strategy generation

Usage:
```python
planner = OptimizationPlanner(detector)
plan = await planner.create_optimization_plan(bottlenecks)
print(f"Estimated improvement: {plan.estimated_total_improvement}")
print(f"Complexity: {plan.implementation_complexity}")
print(f"Risk: {plan.risk_level}")
```

### 5. AI-Powered Optimization
[ai-optimization.md](./performance-optimization/ai-optimization.md)

Intelligent optimization suggestions using Context7 documentation integration.

Key Features:
- Context7 integration for latest performance patterns
- Algorithm complexity analysis and recommendations
- Data structure optimization suggestions
- Concurrency improvement strategies
- Hybrid AI and rule-based approach

Usage:
```python
optimizer = IntelligentOptimizer(context7_client=context7)
suggestions = await optimizer.get_ai_optimization_suggestions(
    bottlenecks,
    codebase_context={'project_type': 'web_api'}
)
```

## Quick Start Example

Complete performance optimization workflow:

```python
# Initialize profiler
profiler = PerformanceProfiler(context7_client=context7)

# Start profiling
profiler.start_profiling(['cpu', 'memory', 'line'])

# Run code to profile
result = expensive_function(1000)

# Stop profiling
profile_results = profiler.stop_profiling()

# Detect bottlenecks
detector = BottleneckDetector(profiler)
bottlenecks = await detector.detect_bottlenecks(profile_results)

# Create optimization plan
planner = OptimizationPlanner(detector)
plan = await planner.create_optimization_plan(bottlenecks)

# Get AI suggestions
optimizer = IntelligentOptimizer(context7_client)
ai_suggestions = await optimizer.get_ai_optimization_suggestions(
    bottlenecks,
    codebase_context={'project_type': 'web_api'}
)

# Report results
print(f"Found {len(bottlenecks)} bottlenecks")
print(f"Estimated improvement: {plan.estimated_total_improvement}")
print(f"Risk level: {plan.risk_level}")
```

## Performance Metrics Types

### CPU Metrics
- Total time (function execution time excluding subcalls)
- Cumulative time (total time including subcalls)
- Call count (number of invocations)
- Per-call time (average execution time)

### Memory Metrics
- Current memory usage (bytes)
- Peak memory usage (bytes)
- Memory by function (line-level allocation)
- Memory delta (before/after comparison)

### Real-Time Metrics
- CPU percentage (process-wide)
- Memory usage (MB and percentage)
- Open file handles
- Thread count
- Context switches

## Optimization Categories

### Algorithm Improvements
- Big-O complexity reduction
- Dynamic programming for overlapping subproblems
- Memoization for repeated calculations
- Efficient search and sort algorithms

### Caching Strategies
- LRU cache implementation
- Memoization decorators
- Query result caching
- Object pooling

### Concurrency Improvements
- Multiprocessing for CPU-bound tasks
- Threading for I/O-bound operations
- Asyncio for concurrent I/O
- Thread/process pool execution

### Memory Optimizations
- Generator usage for large datasets
- Lazy loading patterns
- Memory-efficient data structures
- Object lifecycle management

### Data Structure Changes
- Set/dict for O(1) lookups
- Deque for queue operations
- Numpy arrays for numerical data
- Trie structures for prefix searches

## Best Practices

### Profiling
1. **Baseline Measurement**: Always establish performance baseline before optimization
2. **Realistic Workloads**: Profile with production-like data and usage patterns
3. **Multiple Runs**: Profile multiple times to account for variability
4. **Statistical Significance**: Ensure sufficient execution time for accuracy
5. **Incremental Changes**: Profile after each optimization to measure impact

### Bottleneck Detection
1. **Severity Prioritization**: Focus on critical and high severity bottlenecks first
2. **Impact Scores**: Use quantitative scores to prioritize optimization efforts
3. **Root Cause Analysis**: Understand underlying causes before applying fixes
4. **Context Awareness**: Consider codebase context when suggesting optimizations

### Optimization Execution
1. **One at a Time**: Apply optimizations individually to measure impact
2. **Version Control**: Maintain git branches for each optimization
3. **Comprehensive Testing**: Ensure functionality is preserved during optimization
4. **Performance Regression Tests**: Automate performance validation
5. **Production Monitoring**: Monitor performance in staging before production rollout

### Risk Management
1. **Low-Risk First**: Start with low-complexity, high-impact optimizations
2. **Backup Strategy**: Always maintain ability to rollback changes
3. **Testing Strategy**: Validate with comprehensive test suite
4. **Gradual Rollout**: Use feature flags and gradual deployment
5. **Monitoring**: Set up performance monitoring and alerting

## Integration with Context7

The AI-powered optimization module integrates with Context7 to provide:

- Latest performance optimization patterns from official documentation
- Algorithm complexity analysis and best practices
- Framework-specific optimization techniques
- Real-time documentation updates for 2025+ patterns

Context7 Queries:
```python
# Performance optimization patterns
await context7.get_library_docs(
    context7_library_id="/performance/python-profiling",
    topic="advanced performance optimization patterns 2025",
    tokens=5000
)

# Algorithm optimization
await context7.get_library_docs(
    context7_library_id="/algorithms/python",
    topic="algorithm optimization big-O complexity reduction",
    tokens=3000
)
```

## Common Use Cases

### Web Application Optimization
Profile request handlers, database queries, and template rendering to identify bottlenecks in web applications.

### Data Pipeline Optimization
Optimize ETL processes, data transformation, and batch processing jobs for better throughput.

### API Performance
Identify slow endpoints, optimize database queries, and implement caching strategies.

### Memory Leak Detection
Track memory usage over time to identify and fix memory leaks in long-running processes.

### Algorithm Optimization
Analyze algorithmic complexity and implement more efficient data structures and algorithms.

## Dependencies

Required:
- Python 3.8+
- cProfile (standard library)
- psutil (system monitoring)
- memory_profiler (memory tracking)
- line_profiler (line-level profiling)
- tracemalloc (standard library)

Optional:
- Context7 MCP (AI-powered suggestions)
- asyncio (async profiling support)

## Module Versions

- real-time-monitoring.md: v1.0.0
- profiler-core.md: v1.0.0
- bottleneck-detection.md: v1.0.0
- optimization-plan.md: v1.0.0
- ai-optimization.md: v1.0.0

## Performance Impact

Profiling overhead varies by configuration:
- CPU profiling: 5-15% overhead
- Memory profiling: 10-20% overhead
- Line profiling: 20-30% overhead
- Real-time monitoring: 2-5% overhead

Recommendation: Use selective profiling to minimize overhead in production environments.

---

Module: `modules/performance-optimization.md`
Related: [AI Debugging](./ai-debugging.md) | [Smart Refactoring](./smart-refactoring.md)
