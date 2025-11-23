# Profiling Techniques & Real-World Examples

Context7-enhanced profiling techniques with real-world performance analysis scenarios.

## Profiling Tools Overview

### Scalene Profiler
- **Purpose**: AI-powered CPU/GPU/memory profiling
- **Best For**: Python applications, ML workloads
- **Integration**: Context7 latest patterns

### cProfile
- **Purpose**: Standard library profiler
- **Best For**: Function-level profiling
- **Integration**: Python standard library

### Py-Spy
- **Purpose**: Sampling profiler without code changes
- **Best For**: Production environments
- **Integration**: Minimal overhead profiling

## Real-World Performance Analysis Examples

### Example 1: Web API Performance Analysis

```python
async def analyze_web_api_performance():
    """Comprehensive web API performance analysis."""

    # Phase 1: Baseline
    analyzer = BaselineAnalyzer()
    baseline = await analyzer.measure_baseline("https://api.example.com")

    print(f"Baseline P95 Response Time: {baseline.response_time.p95}ms")

    # Phase 2: Identify bottlenecks
    identifier = BottleneckIdentifier()
    bottlenecks = await identifier.identify_bottlenecks("/path/to/api/code")

    print(f"Found {len(bottlenecks.prioritized_list)} bottlenecks")

    # Phase 3: Deep profiling
    profiler = ProfilingAnalyzer()
    profile = await profiler.comprehensive_profile("/path/to/api/code")

    print(f"Top CPU hotspot: {profile.combined_insights.hotspots[0]}")

    # Phase 4: Create strategy
    strategist = OptimizationStrategist()
    strategy = await strategist.create_strategy(profile)

    print(f"Quick wins: {len(strategy.quick_wins)}")
    print(f"Estimated total improvement: {strategy.estimated_total_improvement}")

    # Phase 5: Setup monitoring
    monitor = PerformanceMonitor()
    monitoring = await monitor.setup_monitoring("api")

    print(f"Monitoring dashboard: {monitoring.dashboard}")
```

### Example 2: Database Query Optimization

```python
async def optimize_database_queries():
    """Analyze and optimize database query performance."""

    # Measure slow queries
    slow_queries = await identify_slow_queries(database_connection)

    for query in slow_queries:
        # Analyze execution plan
        execution_plan = await analyze_execution_plan(query)

        # Identify missing indexes
        missing_indexes = identify_missing_indexes(execution_plan)

        # Suggest optimizations
        optimizations = suggest_query_optimizations(query, execution_plan)

        print(f"Query: {query.sql[:50]}...")
        print(f"  Current time: {query.execution_time}ms")
        print(f"  Missing indexes: {len(missing_indexes)}")
        print(f"  Optimization suggestions: {len(optimizations)}")

        # Apply optimizations
        for optimization in optimizations:
            if optimization.impact > 0.5:  # >50% improvement
                await apply_optimization(query, optimization)
```

### Example 3: Memory Leak Detection

```python
async def detect_memory_leaks():
    """Detect and analyze memory leaks."""

    profiler = MemoryProfiler()

    # Monitor memory over time
    memory_samples = []
    for i in range(100):
        snapshot = await profiler.take_snapshot()
        memory_samples.append(snapshot)

        await asyncio.sleep(1)  # Sample every second

    # Analyze memory growth
    growth_rate = calculate_growth_rate(memory_samples)

    if growth_rate > 0.01:  # 1% growth per minute
        print("⚠️ Potential memory leak detected")

        # Find leaked objects
        leaked_objects = await profiler.find_leaked_objects(memory_samples)

        for obj in leaked_objects:
            print(f"Leaked object: {obj.type}")
            print(f"  Count: {obj.count}")
            print(f"  Size: {obj.size / 1024 / 1024:.2f} MB")
            print(f"  Allocation site: {obj.allocation_site}")
```

### Example 4: Load Testing Analysis

```python
async def analyze_load_test_results():
    """Analyze load testing results."""

    load_tester = LoadTester()

    # Run load test
    test_config = LoadTestConfig(
        target_url="https://api.example.com",
        concurrent_users=1000,
        duration_seconds=300,
        ramp_up_seconds=60
    )

    results = await load_tester.run_test(test_config)

    # Analyze results
    print(f"Total requests: {results.total_requests}")
    print(f"Success rate: {results.success_rate * 100:.2f}%")
    print(f"Average response time: {results.average_response_time}ms")
    print(f"P95 response time: {results.p95_response_time}ms")
    print(f"P99 response time: {results.p99_response_time}ms")
    print(f"Peak throughput: {results.peak_throughput} req/s")

    # Identify performance degradation
    if results.p95_response_time > 500:
        print("⚠️ P95 response time exceeds 500ms threshold")

        # Analyze causes
        degradation_analysis = await analyze_degradation(results)

        for cause in degradation_analysis.root_causes:
            print(f"Root cause: {cause.type}")
            print(f"  Impact: {cause.impact}%")
            print(f"  Recommendation: {cause.recommendation}")
```

### Example 5: Real-Time Performance Dashboard

```python
class RealTimePerformanceDashboard:
    """Real-time performance monitoring dashboard."""

    def __init__(self):
        self.metrics_collector = MetricsCollector()
        self.alert_manager = AlertManager()

    async def start_monitoring(self):
        """Start real-time monitoring."""

        while True:
            # Collect current metrics
            metrics = await self.metrics_collector.collect_all()

            # Update dashboard
            await self.update_dashboard(metrics)

            # Check for anomalies
            anomalies = self.detect_anomalies(metrics)

            if anomalies:
                for anomaly in anomalies:
                    await self.alert_manager.send_alert(anomaly)

            await asyncio.sleep(5)  # Update every 5 seconds

    async def update_dashboard(self, metrics: MetricsSnapshot):
        """Update dashboard with latest metrics."""

        dashboard_data = {
            "timestamp": metrics.timestamp.isoformat(),
            "cpu": {
                "current": metrics.cpu_usage,
                "trend": self.calculate_trend(metrics.cpu_history),
                "status": self.get_status(metrics.cpu_usage, threshold=70)
            },
            "memory": {
                "current": metrics.memory_usage,
                "trend": self.calculate_trend(metrics.memory_history),
                "status": self.get_status(metrics.memory_usage, threshold=80)
            },
            "response_time": {
                "p95": metrics.response_time_p95,
                "trend": self.calculate_trend(metrics.response_time_history),
                "status": self.get_status(metrics.response_time_p95, threshold=200)
            }
        }

        await self.dashboard.update(dashboard_data)
```

---

**Last Updated**: 2025-11-23
**Status**: Production Ready
**Lines**: 270
**Code Examples**: 5 comprehensive examples
