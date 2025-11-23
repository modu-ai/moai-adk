# Performance Optimization Reference Guide

**Version**: 2.0.0
**Last Updated**: 2025-11-24
**Status**: Production Ready

---

## Quick Command Reference

### Scalene Profiler

```bash
# Basic profiling (CPU only)
scalene script.py

# Complete profiling (CPU + GPU + Memory)
scalene --cpu --gpu --memory script.py

# HTML report generation
scalene --cpu --gpu --memory --html --outfile report.html script.py

# Profile specific function
scalene --profile-only my_module.my_function script.py

# Reduced output (top 10 functions)
scalene --reduced-profile script.py

# Profile with specific Python
python -m scalene script.py

# JSON output for automation
scalene --json --outfile profile.json script.py
```

### cProfile (Standard Library)

```bash
# Basic profiling
python -m cProfile script.py

# Save profile data
python -m cProfile -o profile.stats script.py

# Sort by cumulative time
python -m cProfile -s cumulative script.py

# Analyze with pstats
python -c "import pstats; p = pstats.Stats('profile.stats'); p.sort_stats('cumulative').print_stats(20)"
```

### Py-Spy (Production Profiling)

```bash
# Profile running process
py-spy record -o profile.svg --pid <PID>

# Top functions (live monitoring)
py-spy top --pid <PID>

# Flame graph generation
py-spy record -o flamegraph.svg --format flamegraph -- python script.py

# Sample rate control
py-spy record --rate 100 --pid <PID>
```

### Memray (Memory Profiling)

```bash
# Memory profiling
python -m memray run script.py

# Flamegraph from memory profile
python -m memray flamegraph memray-output.bin

# Memory leak detection
python -m memray run --live script.py

# Track allocations
python -m memray tree memray-output.bin
```

---

## Tool Versions (2025-11-24)

### Primary Tools

| Tool | Latest Version | Release Date | Compatibility | Context7 ID |
|------|----------------|--------------|---------------|-------------|
| **Scalene** | 1.5.42 | 2024-11-15 | Python 3.8-3.12 | `/plasma-umass/scalene` |
| **cProfile** | Built-in | Python 3.12 | Python 2.7+ | `/python/cpython` |
| **Py-Spy** | 0.3.14 | 2024-09-20 | Python 2.7-3.12 | `/benfred/py-spy` |
| **Memray** | 1.14.0 | 2024-10-18 | Python 3.7-3.12 | `/bloomberg/memray` |

### Monitoring Tools

| Tool | Latest Version | Release Date | Context7 ID |
|------|----------------|--------------|-------------|
| **Prometheus** | 2.48.0 | 2024-11-10 | `/prometheus/prometheus` |
| **Grafana** | 10.2.2 | 2024-11-05 | `/grafana/grafana` |
| **Node Exporter** | 1.7.0 | 2024-10-25 | `/prometheus/node_exporter` |

### GPU Tools

| Tool | Latest Version | Release Date | Context7 ID |
|------|----------------|--------------|-------------|
| **CUDA Toolkit** | 12.3 | 2024-11-01 | `/nvidia/cuda-samples` |
| **cuDNN** | 9.0.0 | 2024-10-20 | `/nvidia/cudnn` |
| **NVIDIA Nsight** | 2024.3 | 2024-11-12 | `/nvidia/nsight-systems` |

---

## Context7 Library Mapping

### Performance Profiling

```python
# Scalene profiler patterns
scalene_docs = await context7.get_library_docs(
    context7_library_id="/plasma-umass/scalene",
    topic="profiling optimization patterns CPU GPU memory 2025",
    tokens=5000
)

# Python standard library (cProfile)
cpython_docs = await context7.get_library_docs(
    context7_library_id="/python/cpython",
    topic="cProfile profiling pstats performance analysis",
    tokens=3000
)

# Py-Spy production profiling
pyspy_docs = await context7.get_library_docs(
    context7_library_id="/benfred/py-spy",
    topic="production profiling sampling flame graphs",
    tokens=2000
)

# Memray memory profiling
memray_docs = await context7.get_library_docs(
    context7_library_id="/bloomberg/memray",
    topic="memory profiling leak detection allocation tracking",
    tokens=3000
)
```

### Monitoring & Alerting

```python
# Prometheus metrics
prometheus_docs = await context7.get_library_docs(
    context7_library_id="/prometheus/prometheus",
    topic="metrics collection PromQL alerting rules 2025",
    tokens=4000
)

# Grafana dashboards
grafana_docs = await context7.get_library_docs(
    context7_library_id="/grafana/grafana",
    topic="dashboard creation visualization alerting",
    tokens=3000
)
```

### GPU Acceleration

```python
# CUDA optimization
cuda_docs = await context7.get_library_docs(
    context7_library_id="/nvidia/cuda-samples",
    topic="CUDA optimization kernel tuning memory coalescing 2025",
    tokens=5000
)

# cuDNN deep learning
cudnn_docs = await context7.get_library_docs(
    context7_library_id="/nvidia/cudnn",
    topic="cuDNN convolution optimization batch normalization",
    tokens=4000
)

# NVIDIA profiling tools
nsight_docs = await context7.get_library_docs(
    context7_library_id="/nvidia/nsight-systems",
    topic="GPU profiling performance analysis",
    tokens=3000
)
```

---

## API Reference

### Scalene Integration

```python
from scalene import scalene_profiler

class ScaleneIntegration:
    """Enterprise Scalene profiling integration."""

    # Start profiling
    def start_profiling(self):
        """Start Scalene profiler."""
        scalene_profiler.start()

    # Stop profiling
    def stop_profiling(self):
        """Stop profiler and generate report."""
        scalene_profiler.stop()

    # Decorator usage
    @scalene_profiler.profile
    def function_to_profile(self):
        """Function will be profiled automatically."""
        pass

    # Context manager usage
    def profile_block(self):
        """Profile specific code block."""
        with scalene_profiler.profile():
            # Code to profile
            expensive_operation()
```

### Performance Monitoring

```python
class PerformanceMonitor:
    """Real-time performance monitoring."""

    async def collect_metrics(self) -> MetricsSnapshot:
        """Collect current performance metrics."""

        return MetricsSnapshot(
            timestamp=datetime.now(),
            cpu_usage=await self.measure_cpu(),
            memory_usage=await self.measure_memory(),
            response_time_p95=await self.measure_p95(),
            throughput=await self.measure_throughput()
        )

    async def setup_alerts(self, thresholds: dict):
        """Configure performance alerts."""

        alert_rules = []
        for metric, threshold in thresholds.items():
            rule = AlertRule(
                metric=metric,
                threshold=threshold,
                action=self.send_alert
            )
            alert_rules.append(rule)

        return alert_rules
```

### AI Bottleneck Detection

```python
class AIBottleneckDetector:
    """AI-powered bottleneck detection."""

    async def detect_bottlenecks(self, perf_data: dict) -> BottleneckReport:
        """Detect performance bottlenecks with AI."""

        # Get Context7 patterns
        patterns = await self.context7.get_library_docs(
            context7_library_id="/plasma-umass/scalene",
            topic="bottleneck detection optimization patterns",
            tokens=5000
        )

        # AI analysis
        bottlenecks = self.ai_analyzer.detect(perf_data)

        # Match patterns
        recommendations = self.match_patterns(bottlenecks, patterns)

        return BottleneckReport(
            bottlenecks=bottlenecks,
            recommendations=recommendations,
            priority=self.prioritize(bottlenecks)
        )
```

---

## Configuration Examples

### Scalene Configuration

```python
# scalene.config.json
{
    "profiling": {
        "cpu": true,
        "gpu": true,
        "memory": true
    },
    "output": {
        "format": "html",
        "filename": "profile-report.html",
        "reduced": false
    },
    "sampling": {
        "rate": 0.01,
        "interval": 0.1
    },
    "filters": {
        "profile_only": ["my_module.critical_function"],
        "exclude": ["tests/*", "vendor/*"]
    }
}
```

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'application'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']

rule_files:
  - 'alert_rules.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

### Alert Rules Configuration

```yaml
# alert_rules.yml
groups:
  - name: performance_alerts
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"

      - alert: MemoryLeak
        expr: memory_growth_rate > 0.01
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Potential memory leak"

      - alert: SlowResponseTime
        expr: http_request_duration_p95 > 200
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "P95 response time exceeds threshold"
```

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "Performance Monitoring",
    "panels": [
      {
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(cpu_usage[5m])",
            "legendFormat": "CPU Usage"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "memory_usage_bytes",
            "legendFormat": "Memory Usage"
          }
        ]
      },
      {
        "title": "Response Time (P95)",
        "type": "graph",
        "targets": [
          {
            "expr": "http_request_duration_p95",
            "legendFormat": "P95 Response Time"
          }
        ]
      }
    ]
  }
}
```

---

## Performance Metrics Reference

### Latency Targets

| Metric | Excellent | Good | Acceptable | Critical |
|--------|-----------|------|------------|----------|
| **P50** | <20ms | <50ms | <100ms | >100ms |
| **P95** | <50ms | <200ms | <500ms | >500ms |
| **P99** | <100ms | <500ms | <1000ms | >1000ms |
| **P99.9** | <200ms | <1000ms | <2000ms | >2000ms |

### Resource Utilization

| Resource | Healthy | Warning | Critical | Action Required |
|----------|---------|---------|----------|----------------|
| **CPU** | <60% | 60-80% | 80-90% | >90% |
| **Memory** | <75% | 75-85% | 85-95% | >95% |
| **Disk I/O** | <70% | 70-85% | 85-95% | >95% |
| **Network** | <60% | 60-80% | 80-90% | >90% |

### Throughput Benchmarks

| Application Type | Low | Medium | High | Excellent |
|------------------|-----|--------|------|-----------|
| **Web API** | <100 req/s | 100-1000 req/s | 1000-5000 req/s | >5000 req/s |
| **Database** | <100 qps | 100-1000 qps | 1000-10000 qps | >10000 qps |
| **Message Queue** | <100 msg/s | 100-1000 msg/s | 1000-10000 msg/s | >10000 msg/s |

---

## Troubleshooting Guide

### Common Issues

#### Issue 1: Scalene Not Profiling GPU

**Symptoms**: GPU column shows 0% utilization

**Diagnosis**:
```bash
# Check CUDA installation
nvidia-smi

# Verify Scalene GPU support
python -c "import scalene; print(scalene.has_cuda_support())"
```

**Solution**:
```bash
# Reinstall Scalene with GPU support
pip uninstall scalene
pip install scalene[cuda]
```

#### Issue 2: cProfile Output Too Large

**Symptoms**: Profile file >100MB, difficult to analyze

**Solution**:
```python
import pstats

# Load and filter profile
p = pstats.Stats('profile.stats')

# Show only top 20 functions by cumulative time
p.sort_stats('cumulative').print_stats(20)

# Filter by module
p.print_stats('my_module')
```

#### Issue 3: Memory Profiling Overhead

**Symptoms**: Application slows down significantly with Memray

**Solution**:
```bash
# Use sampling mode (lower overhead)
python -m memray run --sample-rate 10 script.py

# Profile specific sections only
# Use context manager in code:
from memray import Tracker

with Tracker("output.bin"):
    # Only profile this block
    critical_section()
```

#### Issue 4: Prometheus High Cardinality

**Symptoms**: Prometheus storage usage increasing rapidly

**Diagnosis**:
```promql
# Check metric cardinality
topk(10, count by (__name__)({__name__=~".+"}))
```

**Solution**:
- Reduce label count
- Use histogram_quantile instead of raw quantiles
- Set appropriate retention period

---

## Official Documentation Links

### Profiling Tools
- [Scalene Documentation](https://github.com/plasma-umass/scalene)
- [Python cProfile](https://docs.python.org/3/library/profile.html)
- [Py-Spy GitHub](https://github.com/benfred/py-spy)
- [Memray Documentation](https://bloomberg.github.io/memray/)

### Monitoring
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)

### GPU
- [CUDA Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/)
- [cuDNN Documentation](https://docs.nvidia.com/deeplearning/cudnn/)
- [NVIDIA Nsight Systems](https://developer.nvidia.com/nsight-systems)

---

## Context7 Pattern Examples (2025)

### Latest Optimization Patterns

```python
# Fetch latest Scalene patterns
scalene_patterns = await context7.get_library_docs(
    context7_library_id="/plasma-umass/scalene",
    topic="2025 optimization patterns AI-powered profiling",
    tokens=5000
)

# Extract recommendations
for pattern in scalene_patterns['patterns']:
    print(f"Pattern: {pattern['name']}")
    print(f"Use Case: {pattern['use_case']}")
    print(f"Expected Improvement: {pattern['expected_improvement']}")
```

### Validation Example

```python
# Validate optimization against Context7 patterns
async def validate_optimization(optimization: dict) -> ValidationReport:
    """Validate optimization with Context7 best practices."""

    # Get relevant patterns from monitoring and profiling tools
    patterns = await context7.get_library_docs(
        context7_library_id="/prometheus/prometheus",
        topic=f"performance monitoring validation patterns {optimization['type']}",
        tokens=3000
    )

    # Check compliance
    compliance = check_pattern_compliance(optimization, patterns)

    return ValidationReport(
        optimization=optimization,
        patterns=patterns,
        compliance=compliance,
        recommendations=generate_recommendations(compliance)
    )
```

---

**Last Updated**: 2025-11-24
**Version**: 2.0.0
**Maintained by**: MoAI-ADK Team
**Status**: Production Ready

**For detailed implementations**: See modules/ directory
**For working examples**: See examples.md
