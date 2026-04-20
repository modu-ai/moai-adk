# Bottleneck Detection

> Module: Performance bottleneck detection and analysis
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: asyncio, Context7 MCP

## Core Implementation

### Bottleneck Detection System

```python
from typing import Dict, List, Any, Optional
from dataclasses import dataclass
from enum import Enum
import asyncio

class OptimizationType(Enum):
    """Types of performance optimizations."""
    ALGORITHM_IMPROVEMENT = "algorithm_improvement"
    CACHING = "caching"
    CONCURRENCY = "concurrency"
    MEMORY_OPTIMIZATION = "memory_optimization"
    I_O_OPTIMIZATION = "io_optimization"
    DATABASE_OPTIMIZATION = "database_optimization"
    DATA_STRUCTURE_CHANGE = "data_structure_change"

@dataclass
class PerformanceBottleneck:
    """Identified performance bottleneck with analysis."""
    function_name: str
    file_path: str
    line_number: int
    bottleneck_type: str  # "cpu", "memory", "io", "algorithm"
    severity: str  # "low", "medium", "high", "critical"
    impact_score: float  # 0.0 to 1.0
    description: str
    metrics: Dict[str, float]
    optimization_type: OptimizationType
    suggested_fixes: List[str]
    estimated_improvement: str
    code_snippet: str

class BottleneckDetector:
    """Detect and analyze performance bottlenecks."""

    def __init__(self, profiler):
        self.profiler = profiler

    async def detect_bottlenecks(
        self, profile_results: Dict[str, Any],
        context7_patterns: Dict[str, Any] = None
    ) -> List[PerformanceBottleneck]:
        """Detect performance bottlenecks from profiling results."""

        bottlenecks = []

        # Analyze CPU bottlenecks
        if 'cpu_profile' in profile_results:
            cpu_bottlenecks = await self._detect_cpu_bottlenecks(
                profile_results['cpu_profile'], context7_patterns
            )
            bottlenecks.extend(cpu_bottlenecks)

        # Analyze memory bottlenecks
        if 'memory_profile' in profile_results:
            memory_bottlenecks = await self._detect_memory_bottlenecks(
                profile_results['memory_profile'], context7_patterns
            )
            bottlenecks.extend(memory_bottlenecks)

        # Analyze real-time metrics
        if 'realtime_metrics' in profile_results:
            realtime_bottlenecks = await self._detect_realtime_bottlenecks(
                profile_results['realtime_metrics'], context7_patterns
            )
            bottlenecks.extend(realtime_bottlenecks)

        # Sort by impact score
        bottlenecks.sort(key=lambda x: x.impact_score, reverse=True)
        return bottlenecks

    async def _detect_cpu_bottlenecks(
        self, cpu_profiles: List,
        context7_patterns: Dict[str, Any] = None
    ) -> List[PerformanceBottleneck]:
        """Detect CPU-related bottlenecks."""

        bottlenecks = []
        total_cpu_time = sum(p.cumulative_time for p in cpu_profiles)

        for profile in cpu_profiles:
            # Skip functions with very low total time
            if profile.cumulative_time < 0.01:
                continue

            # Calculate impact score
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

            # Get code snippet
            code_snippet = self._get_code_snippet(profile.file_path, profile.line_number)

            # Generate optimization suggestions
            optimization_type, suggestions, estimated_improvement = await self._generate_cpu_optimization_suggestions(
                profile, context7_patterns
            )

            bottleneck = PerformanceBottleneck(
                function_name=profile.name,
                file_path=profile.file_path,
                line_number=profile.line_number,
                bottleneck_type="cpu",
                severity=severity,
                impact_score=impact_score,
                description=f"Function '{profile.name}' consumes {impact_score:.1%} of total CPU time",
                metrics={
                    'cumulative_time': profile.cumulative_time,
                    'total_time': profile.total_time,
                    'call_count': profile.call_count,
                    'per_call_time': profile.per_call_time
                },
                optimization_type=optimization_type,
                suggested_fixes=suggestions,
                estimated_improvement=estimated_improvement,
                code_snippet=code_snippet
            )
            bottlenecks.append(bottleneck)

        return bottlenecks

    async def _detect_memory_bottlenecks(
        self, memory_profile: Dict[str, Any],
        context7_patterns: Dict[str, Any] = None
    ) -> List[PerformanceBottleneck]:
        """Detect memory-related bottlenecks."""

        bottlenecks = []

        if 'memory_line_profile' in memory_profile:
            memory_by_function = memory_profile['memory_line_profile'].get('memory_by_function', {})

            if memory_by_function:
                max_memory = max(memory_by_function.values())

                for func_key, memory_usage in memory_by_function.items():
                    # Skip very small memory usage
                    if memory_usage < 1024 * 1024:  # 1MB
                        continue

                    # Calculate impact score
                    impact_score = memory_usage / max(max_memory, 1)

                    # Determine severity
                    if impact_score > 0.7:
                        severity = "critical"
                    elif impact_score > 0.4:
                        severity = "high"
                    elif impact_score > 0.2:
                        severity = "medium"
                    else:
                        severity = "low"

                    # Extract file path and line number
                    if ':' in func_key:
                        file_path, line_num = func_key.split(':', 1)
                        line_number = int(line_num)
                    else:
                        continue

                    # Get code snippet
                    code_snippet = self._get_code_snippet(file_path, line_number)

                    # Generate optimization suggestions
                    optimization_type, suggestions, estimated_improvement = await self._generate_memory_optimization_suggestions(
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
                        metrics={
                            'memory_usage_mb': memory_usage / 1024 / 1024,
                            'impact_score': impact_score
                        },
                        optimization_type=optimization_type,
                        suggested_fixes=suggestions,
                        estimated_improvement=estimated_improvement,
                        code_snippet=code_snippet
                    )
                    bottlenecks.append(bottleneck)

        return bottlenecks

    async def _detect_realtime_bottlenecks(
        self, realtime_metrics: Dict[str, Any],
        context7_patterns: Dict[str, Any] = None
    ) -> List[PerformanceBottleneck]:
        """Detect bottlenecks from real-time monitoring."""

        bottlenecks = []

        # Check CPU usage
        avg_cpu = realtime_metrics.get('avg_cpu_percent', 0)
        if avg_cpu > 80:
            bottleneck = PerformanceBottleneck(
                function_name="System CPU Usage",
                file_path="system",
                line_number=0,
                bottleneck_type="cpu",
                severity="high" if avg_cpu > 90 else "medium",
                impact_score=avg_cpu / 100.0,
                description=f"High average CPU usage: {avg_cpu:.1f}%",
                metrics={'avg_cpu_percent': avg_cpu},
                optimization_type=OptimizationType.CONCURRENCY,
                suggested_fixes=[
                    "Implement parallel processing",
                    "Optimize algorithms",
                    "Add caching for expensive operations"
                ],
                estimated_improvement="20-50% reduction in CPU usage",
                code_snippet="# System-wide optimization required"
            )
            bottlenecks.append(bottleneck)

        # Check memory usage
        avg_memory = realtime_metrics.get('avg_memory_percent', 0)
        if avg_memory > 75:
            bottleneck = PerformanceBottleneck(
                function_name="System Memory Usage",
                file_path="system",
                line_number=0,
                bottleneck_type="memory",
                severity="high" if avg_memory > 85 else "medium",
                impact_score=avg_memory / 100.0,
                description=f"High average memory usage: {avg_memory:.1f}%",
                metrics={'avg_memory_percent': avg_memory},
                optimization_type=OptimizationType.MEMORY_OPTIMIZATION,
                suggested_fixes=[
                    "Implement memory pooling",
                    "Use generators instead of lists",
                    "Optimize data structures",
                    "Implement object caching with size limits"
                ],
                estimated_improvement="30-60% reduction in memory usage",
                code_snippet="# System-wide memory optimization required"
            )
            bottlenecks.append(bottleneck)

        return bottlenecks

    async def _generate_cpu_optimization_suggestions(
        self, profile,
        context7_patterns: Dict[str, Any] = None
    ) -> tuple:
        """Generate CPU optimization suggestions for a function."""

        suggestions = []
        optimization_type = OptimizationType.ALGORITHM_IMPROVEMENT

        # Analyze function characteristics
        if profile.call_count > 10000 and profile.per_call_time > 0.001:
            optimization_type = OptimizationType.CACHING
            suggestions.extend([
                "Implement memoization for expensive function calls",
                "Add LRU cache for frequently called functions",
                "Consider using functools.lru_cache"
            ])
            estimated_improvement = "50-90% for repeated calls"

        elif profile.cumulative_time > 1.0 and profile.call_count > 100:
            suggestions.extend([
                "Analyze algorithm complexity",
                "Look for O(n²) or worse operations",
                "Consider using more efficient data structures"
            ])
            estimated_improvement = "20-80% depending on algorithm"

        elif profile.call_count < 10 and profile.cumulative_time > 0.5:
            suggestions.extend([
                "Consider parallel processing for long-running operations",
                "Implement asynchronous processing",
                "Use multiprocessing for CPU-bound tasks"
            ])
            optimization_type = OptimizationType.CONCURRENCY
            estimated_improvement = "30-70% with proper concurrency"

        else:
            suggestions.extend([
                "Profile line-by-line to identify slow operations",
                "Check for unnecessary loops or computations",
                "Optimize string operations and regular expressions"
            ])
            estimated_improvement = "10-40% with micro-optimizations"

        return optimization_type, suggestions, estimated_improvement

    async def _generate_memory_optimization_suggestions(
        self, memory_usage: int,
        context7_patterns: Dict[str, Any] = None
    ) -> tuple:
        """Generate memory optimization suggestions."""

        suggestions = []
        optimization_type = OptimizationType.MEMORY_OPTIMIZATION

        if memory_usage > 100 * 1024 * 1024:  # 100MB
            suggestions.extend([
                "Implement streaming processing for large datasets",
                "Use generators instead of creating large lists",
                "Process data in chunks to reduce memory footprint"
            ])
            estimated_improvement = "60-90% memory reduction"

        elif memory_usage > 10 * 1024 * 1024:  # 10MB
            suggestions.extend([
                "Use memory-efficient data structures",
                "Implement object pooling for frequently allocated objects",
                "Consider using numpy arrays for numerical data"
            ])
            estimated_improvement = "30-60% memory reduction"

        else:
            suggestions.extend([
                "Release unused objects explicitly",
                "Use weak references for caching",
                "Avoid circular references"
            ])
            estimated_improvement = "10-30% memory reduction"

        return optimization_type, suggestions, estimated_improvement

    def _get_code_snippet(self, file_path: str, line_number: int, context_lines: int = 5) -> str:
        """Get code snippet around the specified line."""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            start_line = max(0, line_number - context_lines - 1)
            end_line = min(len(lines), line_number + context_lines)

            snippet_lines = []
            for i in range(start_line, end_line):
                marker = ">>> " if i == line_number - 1 else " "
                snippet_lines.append(f"{marker}{i+1:4d}: {lines[i].rstrip()}")

            return '\n'.join(snippet_lines)

        except Exception:
            return f"// Code not available for {file_path}:{line_number}"
```

## Usage Examples

```python
# Detect bottlenecks from profiling results
detector = BottleneckDetector(profiler)
bottlenecks = await detector.detect_bottlenecks(profile_results)

print(f"Found {len(bottlenecks)} performance bottlenecks:")
for bottleneck in bottlenecks[:5]:  # Show top 5
    print(f"\nBottleneck: {bottleneck.function_name}")
    print(f"  Type: {bottleneck.bottleneck_type}")
    print(f"  Severity: {bottleneck.severity}")
    print(f"  Impact: {bottleneck.impact_score:.2f}")
    print(f"  Description: {bottleneck.description}")
    print(f"  Optimization type: {bottleneck.optimization_type.value}")
    print(f"  Suggested fixes:")
    for fix in bottleneck.suggested_fixes:
        print(f"    - {fix}")
```

## Best Practices

1. **Severity Prioritization**: Focus on critical and high severity bottlenecks first
2. **Impact Score**: Use impact scores to quantify optimization potential
3. **Context-Aware**: Consider codebase context when suggesting optimizations
4. **Incremental**: Adddess one bottleneck at a time to measure impact
5. **Validation**: Always validate optimizations with performance tests

---

Related: [Profiler Core](./profiler-core.md) | [Optimization Plan](./optimization-plan.md)
