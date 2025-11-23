# Scalene Profiling Guide

Complete guide to using Scalene for CPU, GPU, and memory profiling with Context7 integration.

## Scalene Setup & Installation

```bash
# Install Scalene
pip install scalene

# Or with GPU support
pip install scalene[gpu]
```

## Decorator-Based Profiling

```python
from scalene import scalene_profiler

@scalene_profiler.profile
def cpu_intensive_function():
    """Function profiled with @profile decorator."""
    total = 0
    for i in range(1000000):
        total += i * i
    return total

# Run with: scalene --cpu --memory script.py
```

## Programmatic Profiling Control

```python
from scalene import scalene_profiler

# Start profiling
scalene_profiler.start()

# Code to profile
result = process_data()

# Stop profiling
scalene_profiler.stop()

# Get results
profile_data = scalene_profiler.get_profile_data()
```

## Command-Line Profiling

### Basic Profiling
```bash
# CPU profiling
scalene app.py

# CPU + GPU profiling
scalene --cpu --gpu app.py

# CPU + Memory profiling
scalene --cpu --memory app.py

# Comprehensive profiling
scalene --cpu --gpu --memory --html app.py
```

### Advanced Options

```bash
# Profile specific functions only
scalene --profile-only function_name app.py

# Reduced profiling overhead
scalene --reduced-profile app.py

# Set CPU threshold
scalene --cpu-percent-threshold=1.0 app.py

# Generate HTML report
scalene --html output.html app.py
```

## Performance Profiling Patterns

### Pattern 1: Function-Level Profiling

```python
class PerformanceProfiler:
    def profile_function(self, func, *args, **kwargs):
        """Profile individual function execution."""

        scalene_profiler.start()
        result = func(*args, **kwargs)
        scalene_profiler.stop()

        return result, scalene_profiler.get_profile_data()
```

### Pattern 2: Context-Based Profiling

```python
from contextlib import contextmanager

@contextmanager
def profile_context(label):
    """Profile code block with context manager."""

    scalene_profiler.start()
    try:
        yield
    finally:
        scalene_profiler.stop()
        print(f"Profile data for {label}:")
        print(scalene_profiler.get_profile_data())

# Usage
with profile_context("data processing"):
    process_large_dataset()
```

### Pattern 3: Loop-Based Profiling

```python
def profile_iterations(func, iterations=100):
    """Profile function across multiple iterations."""

    results = []
    for i in range(iterations):
        scalene_profiler.start()
        result = func()
        scalene_profiler.stop()

        profile = scalene_profiler.get_profile_data()
        results.append({
            'iteration': i,
            'result': result,
            'profile': profile
        })

    return results
```

## Interpreting Profiling Results

### CPU Profiling Output

- **% of CPU time**: Percentage of CPU consumed
- **num calls**: Number of times function called
- **Time/call**: Average time per call
- **Tottime**: Total time in function
- **Cumtime**: Cumulative time (including callees)

### Memory Profiling Output

- **Peak memory**: Maximum memory during execution
- **Memory growth**: Change in memory
- **Allocations**: Number of memory allocations
- **Memory per call**: Average memory per call

### GPU Profiling Output

- **GPU utilization**: % of GPU utilization
- **GPU memory**: GPU memory used
- **Data transfer**: Host-to-device transfer time
- **Kernel execution**: GPU kernel execution time

## Context7 Scalene Patterns

### Advanced Profiling with AI

```python
class Context7ScaleneProfiler:
    """Scalene profiler with Context7 optimization patterns."""

    async def profile_with_context7(self, target_file):
        """Profile with Context7 patterns."""

        # Get Context7 patterns
        context7_patterns = await self.context7.get_library_docs(
            context7_library_id="/plasma-umass/scalene",
            topic="AI-powered profiling optimization"
        )

        # Apply Context7 profiling command
        profile_command = self.build_context7_command(
            target_file, context7_patterns
        )

        # Execute profiling
        result = self.execute_command(profile_command)

        # Analyze with AI
        ai_analysis = self.ai_analyzer.analyze_profile(result)

        return {
            'profile': result,
            'context7_patterns': context7_patterns,
            'ai_optimizations': ai_analysis
        }
```

## Best Practices

### ✅ DO
- Profile before optimizing (identify real bottlenecks)
- Use comprehensive profiling (CPU + GPU + Memory)
- Profile with realistic data
- Profile under production-like load
- Generate HTML reports for analysis
- Compare before/after profiles

### ❌ DON'T
- Optimize without profiling
- Profile with tiny datasets
- Use profiling in tight loops (overhead)
- Ignore Context7 patterns
- Skip memory profiling for large applications
- Optimize without measuring improvements

---

**Scalene Documentation**: [GitHub](https://github.com/plasma-umass/scalene)
