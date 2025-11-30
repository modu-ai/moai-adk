# Performance Optimization with Real-time Profiling

> Module: Real-time performance profiling, bottleneck detection, and optimization strategies
> Complexity: Advanced
> Time: 30+ minutes
> Dependencies: Python 3.8+, cProfile, memory_profiler, psutil, Context7 MCP, asyncio

## Core Implementation

### PerformanceProfiler Class

```python
import cProfile
import pstats
import io
import time
import threading
import asyncio
from typing import Dict, List, Optional, Any, Union, Callable
from dataclasses import dataclass, field
from enum import Enum
import psutil
import memory_profiler
import line_profiler
from collections import defaultdict, deque
import json
import sys
import tracemalloc
import resource

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

class OptimizationType(Enum):
 """Types of performance optimizations."""
 ALGORITHM_IMPROVEMENT = "algorithm_improvement"
 CACHING = "caching"
 CONCURRENCY = "concurrency"
 MEMORY_OPTIMIZATION = "memory_optimization"
 I/O_OPTIMIZATION = "io_optimization"
 DATABASE_OPTIMIZATION = "database_optimization"
 DATA_STRUCTURE_CHANGE = "data_structure_change"

@dataclass
class PerformanceSnapshot:
 """Snapshot of performance metrics at a point in time."""
 timestamp: float
 cpu_percent: float
 memory_mb: float
 memory_percent: float
 open_files: int
 threads: int
 context_switches: int
 custom_metrics: Dict[str, float] = field(default_factory=dict)

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

@dataclass
class PerformanceBottleneck:
 """Identified performance bottleneck with analysis."""
 function_name: str
 file_path: str
 line_number: int
 bottleneck_type: str # "cpu", "memory", "io", "algorithm"
 severity: str # "low", "medium", "high", "critical"
 impact_score: float # 0.0 to 1.0
 description: str
 metrics: Dict[str, float]
 optimization_type: OptimizationType
 suggested_fixes: List[str]
 estimated_improvement: str
 code_snippet: str

@dataclass
class OptimizationPlan:
 """Comprehensive optimization plan with prioritized actions."""
 bottlenecks: List[PerformanceBottleneck]
 execution_order: List[int]
 estimated_total_improvement: str
 implementation_complexity: str
 risk_level: str
 prerequisites: List[str]
 validation_strategy: str

class RealTimeMonitor:
 """Real-time performance monitoring system."""

 def __init__(self, sampling_interval: float = 1.0):
 self.sampling_interval = sampling_interval
 self.is_monitoring = False
 self.monitor_thread = None
 self.snapshots = deque(maxlen=1000) # Keep last 1000 snapshots
 self.callbacks = []
 self.alerts = []

 def start_monitoring(self):
 """Start real-time performance monitoring."""
 if self.is_monitoring:
 return

 self.is_monitoring = True
 self.monitor_thread = threading.Thread(target=self._monitor_loop, daemon=True)
 self.monitor_thread.start()

 def stop_monitoring(self):
 """Stop real-time performance monitoring."""
 self.is_monitoring = False
 if self.monitor_thread:
 self.monitor_thread.join(timeout=2.0)

 def _monitor_loop(self):
 """Main monitoring loop."""
 process = psutil.Process()

 while self.is_monitoring:
 try:
 # Collect system metrics
 snapshot = PerformanceSnapshot(
 timestamp=time.time(),
 cpu_percent=process.cpu_percent(),
 memory_mb=process.memory_info().rss / 1024 / 1024,
 memory_percent=process.memory_percent(),
 open_files=len(process.open_files()),
 threads=process.num_threads(),
 context_switches=process.num_ctx_switches().voluntary + process.num_ctx_switches().involuntary
 )

 # Check for custom metrics callbacks
 for callback in self.callbacks:
 try:
 custom_metrics = callback()
 snapshot.custom_metrics.update(custom_metrics)
 except Exception as e:
 print(f"Custom metric callback error: {e}")

 self.snapshots.append(snapshot)

 # Check for alerts
 self._check_alerts(snapshot)

 time.sleep(self.sampling_interval)

 except Exception as e:
 print(f"Monitoring error: {e}")
 time.sleep(self.sampling_interval)

 def add_callback(self, callback: Callable[[], Dict[str, float]]):
 """Add custom metric collection callback."""
 self.callbacks.append(callback)

 def _check_alerts(self, snapshot: PerformanceSnapshot):
 """Check for performance alerts."""
 alerts = []

 # CPU usage alert
 if snapshot.cpu_percent > 90:
 alerts.append({
 'type': 'high_cpu',
 'message': f"High CPU usage: {snapshot.cpu_percent:.1f}%",
 'timestamp': snapshot.timestamp
 })

 # Memory usage alert
 if snapshot.memory_percent > 85:
 alerts.append({
 'type': 'high_memory',
 'message': f"High memory usage: {snapshot.memory_percent:.1f}%",
 'timestamp': snapshot.timestamp
 })

 # File handle alert
 if snapshot.open_files > 1000:
 alerts.append({
 'type': 'file_handle_leak',
 'message': f"High number of open files: {snapshot.open_files}",
 'timestamp': snapshot.timestamp
 })

 self.alerts.extend(alerts)

 def get_recent_snapshots(self, count: int = 100) -> List[PerformanceSnapshot]:
 """Get recent performance snapshots."""
 return list(self.snapshots)[-count:]

 def get_average_metrics(self, duration_minutes: int = 5) -> Dict[str, float]:
 """Get average metrics over specified duration."""
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

class PerformanceProfiler:
 """Advanced performance profiler with bottleneck detection."""

 def __init__(self, context7_client=None):
 self.context7 = context7_client
 self.profiler = None
 self.memory_profiler = None
 self.line_profiler = None
 self.realtime_monitor = RealTimeMonitor()
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

 # Start real-time monitoring
 self.realtime_monitor.start_monitoring()

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

 # Stop real-time monitoring
 self.realtime_monitor.stop_monitoring()
 results['realtime_metrics'] = self.realtime_monitor.get_average_metrics()

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
 memory_before=0.0, # Will be filled by memory profiler
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
 self, cpu_profiles: List[FunctionProfile],
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
 if memory_usage < 1024 * 1024: # 1MB
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
 self, profile: FunctionProfile,
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

 if memory_usage > 100 * 1024 * 1024: # 100MB
 suggestions.extend([
 "Implement streaming processing for large datasets",
 "Use generators instead of creating large lists",
 "Process data in chunks to reduce memory footprint"
 ])
 estimated_improvement = "60-90% memory reduction"

 elif memory_usage > 10 * 1024 * 1024: # 10MB
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

 async def create_optimization_plan(
 self, bottlenecks: List[PerformanceBottleneck],
 context7_patterns: Dict[str, Any] = None
 ) -> OptimizationPlan:
 """Create comprehensive optimization plan."""

 # Prioritize bottlenecks by impact and severity
 prioritized_bottlenecks = self._prioritize_bottlenecks(bottlenecks)

 # Create execution order
 execution_order = self._create_optimization_execution_order(prioritized_bottlenecks)

 # Estimate total improvement
 total_improvement = self._estimate_total_improvement(prioritized_bottlenecks)

 # Assess implementation complexity
 complexity = self._assess_implementation_complexity(prioritized_bottlenecks)

 # Assess risk level
 risk_level = self._assess_optimization_risk(prioritized_bottlenecks)

 # Identify prerequisites
 prerequisites = self._identify_optimization_prerequisites(prioritized_bottlenecks)

 # Create validation strategy
 validation_strategy = self._create_validation_strategy(prioritized_bottlenecks)

 return OptimizationPlan(
 bottlenecks=prioritized_bottlenecks,
 execution_order=execution_order,
 estimated_total_improvement=total_improvement,
 implementation_complexity=complexity,
 risk_level=risk_level,
 prerequisites=prerequisites,
 validation_strategy=validation_strategy
 )

 def _prioritize_bottlenecks(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> List[PerformanceBottleneck]:
 """Prioritize bottlenecks by impact and implementation complexity."""

 # Sort by severity, impact score, and optimization type
 severity_order = {'critical': 4, 'high': 3, 'medium': 2, 'low': 1}

 return sorted(
 bottlenecks,
 key=lambda x: (
 severity_order.get(x.severity, 0),
 x.impact_score,
 self._get_optimization_priority(x.optimization_type)
 ),
 reverse=True
 )

 def _get_optimization_priority(self, opt_type: OptimizationType) -> int:
 """Get priority weight for optimization type."""
 priorities = {
 OptimizationType.ALGORITHM_IMPROVEMENT: 4,
 OptimizationType.CACHING: 3,
 OptimizationType.CONCURRENCY: 3,
 OptimizationType.MEMORY_OPTIMIZATION: 2,
 OptimizationType.DATA_STRUCTURE_CHANGE: 2,
 OptimizationType.I/O_OPTIMIZATION: 2,
 OptimizationType.DATABASE_OPTIMIZATION: 1
 }
 return priorities.get(opt_type, 1)

 def _create_optimization_execution_order(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> List[int]:
 """Create optimal execution order for optimizations."""

 # Group by optimization type
 type_groups = defaultdict(list)
 for i, bottleneck in enumerate(bottlenecks):
 type_groups[bottleneck.optimization_type].append(i)

 # Define execution order by type
 execution_order = []
 type_order = [
 OptimizationType.ALGORITHM_IMPROVEMENT,
 OptimizationType.DATA_STRUCTURE_CHANGE,
 OptimizationType.CACHING,
 OptimizationType.MEMORY_OPTIMIZATION,
 OptimizationType.CONCURRENCY,
 OptimizationType.I_IO_OPTIMIZATION,
 OptimizationType.DATABASE_OPTIMIZATION
 ]

 for opt_type in type_order:
 if opt_type in type_groups:
 execution_order.extend(type_groups[opt_type])

 return execution_order

 def _estimate_total_improvement(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> str:
 """Estimate total performance improvement."""

 if not bottlenecks:
 return "No significant improvement expected"

 # Calculate weighted improvement
 total_weighted_improvement = 0
 total_weight = 0

 for bottleneck in bottlenecks:
 # Extract improvement percentage from description
 improvement_range = self._parse_improvement_estimate(bottleneck.estimated_improvement)
 if improvement_range:
 avg_improvement = (improvement_range[0] + improvement_range[1]) / 2
 weight = bottleneck.impact_score
 total_weighted_improvement += avg_improvement * weight
 total_weight += weight

 if total_weight > 0:
 avg_improvement = total_weighted_improvement / total_weight
 return f"{avg_improvement:.0f}% average performance improvement"

 return "Performance improvement depends on implementation"

 def _parse_improvement_estimate(self, estimate: str) -> tuple:
 """Parse improvement percentage from estimate string."""
 import re

 # Look for percentage ranges like "20-50%" or "30%"
 match = re.search(r'(\d+)-?(\d+)?%', estimate)
 if match:
 start = int(match.group(1))
 end = int(match.group(2)) if match.group(2) else start
 return (start, end)

 return None

 def _assess_implementation_complexity(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> str:
 """Assess overall implementation complexity."""

 complexity_scores = {
 OptimizationType.ALGORITHM_IMPROVEMENT: 3,
 OptimizationType.DATA_STRUCTURE_CHANGE: 3,
 OptimizationType.CONCURRENCY: 4,
 OptimizationType.DATABASE_OPTIMIZATION: 3,
 OptimizationType.CACHING: 2,
 OptimizationType.MEMORY_OPTIMIZATION: 2,
 OptimizationType.I_O_OPTIMIZATION: 2
 }

 if not bottlenecks:
 return "low"

 avg_complexity = sum(
 complexity_scores.get(b.optimization_type, 2) * b.impact_score
 for b in bottlenecks
 ) / sum(b.impact_score for b in bottlenecks)

 if avg_complexity > 3.5:
 return "high"
 elif avg_complexity > 2.5:
 return "medium"
 else:
 return "low"

 def _assess_optimization_risk(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> str:
 """Assess risk level of optimizations."""

 high_risk_types = {
 OptimizationType.ALGORITHM_IMPROVEMENT,
 OptimizationType.DATA_STRUCTURE_CHANGE,
 OptimizationType.CONCURRENCY
 }

 high_risk_count = sum(
 1 for b in bottlenecks
 if b.optimization_type in high_risk_types and b.impact_score > 0.3
 )

 if high_risk_count > 3:
 return "high"
 elif high_risk_count > 1:
 return "medium"
 else:
 return "low"

 def _identify_optimization_prerequisites(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> List[str]:
 """Identify prerequisites for safe optimization."""

 prerequisites = [
 "Create comprehensive performance benchmarks",
 "Ensure version control with current implementation",
 "Set up performance testing environment"
 ]

 # Add specific prerequisites based on bottleneck types
 optimization_types = set(b.optimization_type for b in bottlenecks)

 if OptimizationType.CONCURRENCY in optimization_types:
 prerequisites.extend([
 "Review thread safety and shared resource access",
 "Implement proper synchronization mechanisms"
 ])

 if OptimizationType.DATABASE_OPTIMIZATION in optimization_types:
 prerequisites.extend([
 "Create database backup before optimization",
 "Set up database performance monitoring"
 ])

 if OptimizationType.ALGORITHM_IMPROVEMENT in optimization_types:
 prerequisites.extend([
 "Verify algorithm correctness with test suite",
 "Compare against known reference implementations"
 ])

 return prerequisites

 def _create_validation_strategy(
 self, bottlenecks: List[PerformanceBottleneck]
 ) -> str:
 """Create validation strategy for optimizations."""

 strategy = """
 Validation Strategy:
 1. Baseline Performance Measurement
 - Record current performance metrics
 - Establish performance regression thresholds

 2. Incremental Testing
 - Apply optimizations one at a time
 - Measure performance impact after each change

 3. Automated Performance Testing
 - Implement performance regression tests
 - Set up continuous performance monitoring

 4. Functional Validation
 - Run complete test suite after each optimization
 - Verify no functional regressions introduced

 5. Production Monitoring
 - Monitor performance in staging environment
 - Gradual rollout with performance validation
 """

 return strategy

# Usage Examples
"""
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

# Detect bottlenecks
bottlenecks = await profiler.detect_bottlenecks(profile_results)

print(f"Found {len(bottlenecks)} performance bottlenecks:")
for bottleneck in bottlenecks[:5]: # Show top 5
 print(f"\nBottleneck: {bottleneck.function_name}")
 print(f" Type: {bottleneck.bottleneck_type}")
 print(f" Severity: {bottleneck.severity}")
 print(f" Impact: {bottleneck.impact_score:.2f}")
 print(f" Description: {bottleneck.description}")
 print(f" Optimization type: {bottleneck.optimization_type.value}")
 print(f" Suggested fixes:")
 for fix in bottleneck.suggested_fixes:
 print(f" - {fix}")

# Create optimization plan
optimization_plan = await profiler.create_optimization_plan(bottlenecks)

print(f"\nOptimization Plan:")
print(f" Estimated improvement: {optimization_plan.estimated_total_improvement}")
print(f" Implementation complexity: {optimization_plan.implementation_complexity}")
print(f" Risk level: {optimization_plan.risk_level}")
print(f" Prerequisites: {len(optimization_plan.prerequisites)} items")

# Real-time monitoring example
monitor = RealTimeMonitor(sampling_interval=0.5)
monitor.start_monitoring()

# Add custom metrics callback
def custom_metrics():
 return {
 'custom_counter': some_global_counter,
 'queue_size': len(some_queue)
 }

monitor.add_callback(custom_metrics)

# Run application while monitoring
# ... your application code ...

# Stop monitoring and get results
monitor.stop_monitoring()
recent_snapshots = monitor.get_recent_snapshots(10)
avg_metrics = monitor.get_average_metrics(5)

print(f"Average CPU: {avg_metrics.get('avg_cpu_percent', 0):.1f}%")
print(f"Average Memory: {avg_metrics.get('avg_memory_mb', 0):.1f}MB")
"""
```

## Advanced Features

### Intelligent Performance Optimization

AI-Powered Optimization Suggestions:
```python
class IntelligentOptimizer(PerformanceProfiler):
 """Optimizer that uses AI to suggest the best optimizations."""

 def __init__(self, context7_client=None):
 super().__init__(context7_client)
 self.optimization_history = []
 self.performance_models = {}

 async def get_ai_optimization_suggestions(
 self, bottlenecks: List[PerformanceBottleneck],
 codebase_context: Dict[str, Any]
 ) -> Dict[str, Any]:
 """Get AI-powered optimization suggestions using Context7."""

 if not self.context7:
 return self._get_rule_based_suggestions(bottlenecks)

 # Get latest performance optimization patterns
 try:
 optimization_patterns = await self.context7.get_library_docs(
 context7_library_id="/performance/python-profiling",
 topic="advanced performance optimization patterns 2025",
 tokens=5000
 )

 # Get algorithm complexity patterns
 algorithm_patterns = await self.context7.get_library_docs(
 context7_library_id="/algorithms/python",
 topic="algorithm optimization big-O complexity reduction",
 tokens=3000
 )

 # Generate AI suggestions
 ai_suggestions = await self._generate_ai_suggestions(
 bottlenecks, optimization_patterns, algorithm_patterns, codebase_context
 )

 return ai_suggestions

 except Exception as e:
 print(f"AI optimization failed: {e}")
 return self._get_rule_based_suggestions(bottlenecks)

 async def _generate_ai_suggestions(
 self, bottlenecks: List[PerformanceBottleneck],
 opt_patterns: Dict, algo_patterns: Dict, context: Dict
 ) -> Dict[str, Any]:
 """Generate AI-powered optimization suggestions."""

 suggestions = {
 'algorithm_improvements': [],
 'data_structure_optimizations': [],
 'concurrency_improvements': [],
 'caching_strategies': [],
 'io_optimizations': []
 }

 for bottleneck in bottlenecks:
 # Analyze bottleneck characteristics
 if bottleneck.bottleneck_type == "cpu":
 # Check for algorithmic improvements
 if "O(" in bottleneck.description or any(
 keyword in bottleneck.description.lower()
 for keyword in ["loop", "iteration", "search", "sort"]
 ):
 improvement = self._suggest_algorithm_improvement(
 bottleneck, algo_patterns
 )
 suggestions['algorithm_improvements'].append(improvement)

 # Check for concurrency opportunities
 if bottleneck.call_count > 1000:
 concurrency = self._suggest_concurrency_improvement(bottleneck)
 suggestions['concurrency_improvements'].append(concurrency)

 elif bottleneck.bottleneck_type == "memory":
 # Suggest data structure optimizations
 data_structure = self._suggest_data_structure_improvement(
 bottleneck, opt_patterns
 )
 suggestions['data_structure_optimizations'].append(data_structure)

 return suggestions

 def _suggest_algorithm_improvement(
 self, bottleneck: PerformanceBottleneck, algo_patterns: Dict
 ) -> Dict[str, Any]:
 """Suggest algorithmic improvements based on Context7 patterns."""

 # Analyze function name and code to identify algorithm type
 function_name = bottleneck.function_name.lower()

 suggestions = []

 if any(keyword in function_name for keyword in ["search", "find"]):
 suggestions.extend([
 "Consider using binary search for sorted data",
 "Implement hash-based lookup for O(1) average case",
 "Use trie structures for prefix searches"
 ])

 elif any(keyword in function_name for keyword in ["sort", "order"]):
 suggestions.extend([
 "Consider using Timsort (Python's built-in sort)",
 "Use radix sort for uniform integer data",
 "Implement bucket sort for uniformly distributed data"
 ])

 elif "nested" in function_name or bottleneck.metrics.get('per_call_time', 0) > 0.1:
 suggestions.extend([
 "Look for O(n²) nested loops to optimize",
 "Consider dynamic programming for overlapping subproblems",
 "Use memoization to avoid repeated calculations"
 ])

 return {
 'bottleneck': bottleneck.function_name,
 'suggestions': suggestions,
 'estimated_improvement': "30-90% depending on algorithm",
 'implementation_complexity': "medium to high"
 }

 def _suggest_concurrency_improvement(
 self, bottleneck: PerformanceBottleneck
 ) -> Dict[str, Any]:
 """Suggest concurrency improvements."""

 return {
 'bottleneck': bottleneck.function_name,
 'suggestions': [
 "Implement multiprocessing for CPU-bound tasks",
 "Use threading for I/O-bound operations",
 "Consider asyncio for concurrent I/O operations",
 "Use concurrent.futures for thread/process pool execution"
 ],
 'estimated_improvement': "2-8x speedup on multi-core systems",
 'implementation_complexity': "medium"
 }

 def _suggest_data_structure_improvement(
 self, bottleneck: PerformanceBottleneck, opt_patterns: Dict
 ) -> Dict[str, Any]:
 """Suggest data structure optimizations."""

 return {
 'bottleneck': bottleneck.function_name,
 'suggestions': [
 "Use generators instead of lists for large datasets",
 "Implement lazy loading for expensive data structures",
 "Use memoryviews or numpy arrays for numerical data",
 "Consider using collections.deque for queue operations",
 "Use set/dict for O(1) lookups instead of list searches"
 ],
 'estimated_improvement': "30-80% memory reduction",
 'implementation_complexity': "low to medium"
 }
```

## Best Practices

1. Baseline Measurement: Always establish performance baseline before optimization
2. Incremental Changes: Apply one optimization at a time to measure impact
3. Comprehensive Testing: Ensure functionality is preserved during optimization
4. Real-world Workloads: Profile with realistic data and usage patterns
5. Continuous Monitoring: Implement ongoing performance monitoring in production

---

Module: `modules/performance-optimization.md`
Related: [AI Debugging](./ai-debugging.md) | [Smart Refactoring](./smart-refactoring.md)
