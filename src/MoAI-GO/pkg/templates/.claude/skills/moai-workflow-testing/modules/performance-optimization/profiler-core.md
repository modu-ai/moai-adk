# Performance Profiler Core

> Module: CPU, memory, and line profiling with analysis
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: cProfile, memory_profiler, line_profiler, tracemalloc

## Core Implementation

### PerformanceProfiler Class

```python
import cProfile
import pstats
import io
from typing import Dict, List, Any, Optional
from dataclasses import dataclass, field
from enum import Enum
import memory_profiler
import line_profiler
import tracemalloc
from collections import defaultdict

class PerformanceMetric(Enum):
    """Types of performance metrics to track."""
    CPU_TIME = "cpu_time"
    WALL_TIME = "wall_time"
    MEMORY_USAGE = "memory_usage"
    MEMORY_PEAK = "memory_peak"
    FUNCTION_CALLS = "function_calls"
    EXECUTION_COUNT = "execution_count"
    AVERAGE_TIME = "average_time"
    MAX_TIME = "max_time"
    MIN_TIME = "min_time"

@dataclass
class FunctionProfile:
    """Detailed profile information for a function."""
    name: str
    file_path: str
    line_number: int
    total_time: float
    cumulative_time: float
    call_count: int
    per_call_time: float
    cummulative_per_call_time: float
    memory_before: float
    memory_after: float
    memory_delta: float
    optimization_suggestions: List[str] = field(default_factory=list)

class PerformanceProfiler:
    """Advanced performance profiler with bottleneck detection."""

    def __init__(self, context7_client=None):
        self.context7 = context7_client
        self.profiler = None
        self.memory_profiler = None
        self.line_profiler = None
        self.realtime_monitor = None
        self.profiles = {}
        self.bottlenecks = []

    def start_profiling(self, profile_types: List[str] = None):
        """Start performance profiling with specified types."""
        if profile_types is None:
            profile_types = ['cpu', 'memory']

        # Start CPU profiling
        if 'cpu' in profile_types:
            self.profiler = cProfile.Profile()
            self.profiler.enable()

        # Start memory profiling
        if 'memory' in profile_types:
            tracemalloc.start()
            self.memory_profiler = memory_profiler.Profile()

        # Start line profiling for specific functions
        if 'line' in profile_types:
            self.line_profiler = line_profiler.LineProfiler()
            self.line_profiler.enable_by_count()

    def stop_profiling(self) -> Dict[str, Any]:
        """Stop profiling and collect results."""
        results = {}

        # Stop CPU profiling
        if self.profiler:
            self.profiler.disable()
            results['cpu_profile'] = self._analyze_cpu_profile()

        # Stop memory profiling
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

        # Stop line profiling
        if self.line_profiler:
            self.line_profiler.disable()
            results['line_profile'] = self._analyze_line_profile()

        return results

    def _analyze_cpu_profile(self) -> List[FunctionProfile]:
        """Analyze CPU profiling results."""
        if not self.profiler:
            return []

        # Create stats object
        s = io.StringIO()
        ps = pstats.Stats(self.profiler, stream=s)
        ps.sort_stats('cumulative')
        ps.print_stats()

        # Parse the stats
        function_profiles = []
        lines = s.getvalue().split('\n')

        # Skip header lines
        for line in lines[6:]:
            if line.strip() and not line.startswith('ncalls'):
                try:
                    parts = line.split()
                    if len(parts) >= 6:
                        ncalls = parts[0]
                        tottime = float(parts[1])
                        cumtime = float(parts[3])

                        # Extract function name and location
                        filename_func = ' '.join(parts[5:])
                        if '{' in filename_func:
                            filename, line_num, func_name = self._parse_function_line(filename_func)
                        else:
                            continue

                        # Convert ncalls to integer (handle format like "1000/1000")
                        if '/' in ncalls:
                            ncalls = int(ncalls.split('/')[0])
                        else:
                            ncalls = int(ncalls)

                        profile = FunctionProfile(
                            name=func_name,
                            file_path=filename,
                            line_number=int(line_num),
                            total_time=tottime,
                            cumulative_time=cumtime,
                            call_count=ncalls,
                            per_call_time=tottime / max(ncalls, 1),
                            cummulative_per_call_time=cumtime / max(ncalls, 1),
                            memory_before=0.0,  # Will be filled by memory profiler
                            memory_after=0.0,
                            memory_delta=0.0
                        )
                        function_profiles.append(profile)

                except (ValueError, IndexError) as e:
                    continue

        return function_profiles

    def _parse_function_line(self, line: str) -> tuple:
        """Parse function line from pstats output."""
        # Format: "filename(line_number)function_name"
        try:
            paren_idx = line.rfind('(')
            if paren_idx == -1:
                return line, "0", "unknown"

            filename = line[:paren_idx]
            rest = line[paren_idx:]

            closing_idx = rest.find(')')
            if closing_idx == -1:
                return filename, "0", "unknown"

            line_num = rest[1:closing_idx]
            func_name = rest[closing_idx + 1:]

            return filename, line_num, func_name
        except Exception:
            return line, "0", "unknown"

    def _analyze_memory_profile(self) -> Dict[str, Any]:
        """Analyze memory profiling results."""
        if not self.memory_profiler:
            return {}

        # Get memory profile statistics
        stats = self.memory_profiler.get_stats()

        return {
            'total_samples': len(stats),
            'max_memory_usage': max((stat[2] for stat in stats), default=0),
            'memory_by_function': self._group_memory_by_function(stats)
        }

    def _group_memory_by_function(self, stats: List) -> Dict[str, float]:
        """Group memory usage by function."""
        memory_by_function = defaultdict(float)

        for stat in stats:
            filename, line_no, mem_usage = stat
            # Extract function name from filename and line
            func_key = f"{filename}:{line_no}"
            memory_by_function[func_key] += mem_usage

        return dict(memory_by_function)

    def _analyze_line_profile(self) -> Dict[str, Any]:
        """Analyze line profiling results."""
        if not self.line_profiler:
            return {}

        # Get line profiler statistics
        stats = self.line_profiler.get_stats()

        return {
            'timings': stats.timings,
            'unit': stats.unit
        }
```

## Usage Examples

```python
# Initialize performance profiler
profiler = PerformanceProfiler(context7_client=context7)

# Example function to profile
def expensive_function(n):
    result = []
    for i in range(n):
        # Simulate expensive computation
        temp = []
        for j in range(i):
            temp.append(j * j)
        result.extend(temp)
    return result

# Start profiling
profiler.start_profiling(['cpu', 'memory', 'line'])

# Add line profiler for specific function
if profiler.line_profiler:
    profiler.line_profiler.add_function(expensive_function)

# Run the code to be profiled
result = expensive_function(1000)

# Stop profiling and get results
profile_results = profiler.stop_profiling()

print(f"CPU Profile: {len(profile_results.get('cpu_profile', []))} functions")
print(f"Memory Peak: {profile_results.get('memory_profile', {}).get('peak_mb', 0):.2f} MB")
```

## Best Practices

1. **Profile Types**: Start with CPU and memory profiling, add line profiling for specific functions
2. **Baseline Measurement**: Always profile before optimization
3. **Realistic Workloads**: Profile with production-like data and patterns
4. **Multiple Runs**: Profile multiple times to account for variability
5. **Statistical Significance**: Ensure sufficient execution time for accurate measurements

---

Related: [Real-Time Monitoring](./real-time-monitoring.md) | [Bottleneck Detection](./bottleneck-detection.md)
