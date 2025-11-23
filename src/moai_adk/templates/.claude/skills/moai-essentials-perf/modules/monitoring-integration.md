# Enterprise Monitoring Integration

Setting up comprehensive performance monitoring for production systems.

## Real-Time Performance Monitoring

### Prometheus + Grafana Setup

```python
from prometheus_client import Counter, Histogram, start_http_server
import time

# Define metrics
request_count = Counter('requests_total', 'Total requests')
request_duration = Histogram('request_duration_seconds', 'Request duration')

# Start metrics server
start_http_server(8000)

# Use metrics
@request_duration.time()
def handle_request():
    request_count.inc()
    # Process request
    pass
```

### Performance Dashboard

```python
class PerformanceDashboard:
    """Real-time performance monitoring dashboard."""

    def __init__(self):
        self.metrics = {
            'cpu': [],
            'memory': [],
            'latency': [],
            'throughput': []
        }

    def record_metric(self, metric_type, value):
        """Record performance metric."""
        self.metrics[metric_type].append({
            'value': value,
            'timestamp': time.time()
        })

    def get_current_status(self):
        """Get current performance status."""
        return {
            'cpu_avg': self.average_recent('cpu', window=60),
            'memory_peak': self.max_recent('memory', window=300),
            'latency_p95': self.percentile_recent('latency', 0.95),
            'throughput': self.throughput_recent()
        }

    def average_recent(self, metric, window=60):
        """Average metric over time window."""
        recent = [m for m in self.metrics[metric]
                  if time.time() - m['timestamp'] < window]
        if not recent:
            return 0
        return sum(m['value'] for m in recent) / len(recent)
```

## Anomaly Detection

### Real-Time Anomaly Detection

```python
class AnomalyDetector:
    """Detect performance anomalies automatically."""

    def __init__(self, baseline_window=300):
        self.baseline_window = baseline_window
        self.metrics_history = []

    def detect_anomaly(self, current_metrics):
        """Detect if current metrics deviate from baseline."""

        baseline = self.calculate_baseline()

        if not baseline:
            return False, None

        # Check for significant deviations
        deviations = {}
        for metric, value in current_metrics.items():
            baseline_value = baseline[metric]
            deviation = abs(value - baseline_value) / baseline_value

            if deviation > 0.5:  # 50% threshold
                deviations[metric] = deviation

        is_anomaly = len(deviations) > 0

        return is_anomaly, deviations

    def calculate_baseline(self):
        """Calculate baseline from recent history."""

        if not self.metrics_history:
            return None

        baseline = {}
        for metric in self.metrics_history[0].keys():
            values = [m[metric] for m in self.metrics_history
                      if metric in m]
            baseline[metric] = sum(values) / len(values)

        return baseline
```

## CI/CD Performance Integration

### Performance Testing in Pipeline

```yaml
# .github/workflows/performance.yml
name: Performance Testing

on: [push, pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Python
        uses: actions/setup-python@v2
        with:
          python-version: '3.9'

      - name: Install dependencies
        run: pip install scalene pytest pytest-benchmark

      - name: Run performance tests
        run: |
          scalene --cpu --memory --html tests/performance_tests.py
          pytest tests/performance_tests.py --benchmark-only

      - name: Upload results
        uses: actions/upload-artifact@v2
        with:
          name: performance-report
          path: |
            scalene_profile.html
            .benchmarks/
```

## Alerting Configuration

### Performance-Based Alerts

```python
class PerformanceAlerter:
    """Alert on performance degradation."""

    def __init__(self):
        self.alert_thresholds = {
            'cpu_percent': 80,
            'memory_percent': 85,
            'latency_ms': 1000,
            'error_rate': 0.01
        }

    def check_and_alert(self, metrics):
        """Check metrics against thresholds."""

        alerts = []

        if metrics['cpu_percent'] > self.alert_thresholds['cpu_percent']:
            alerts.append({
                'level': 'warning',
                'message': f"CPU at {metrics['cpu_percent']}%"
            })

        if metrics['memory_percent'] > self.alert_thresholds['memory_percent']:
            alerts.append({
                'level': 'critical',
                'message': f"Memory at {metrics['memory_percent']}%"
            })

        if metrics['latency_ms'] > self.alert_thresholds['latency_ms']:
            alerts.append({
                'level': 'warning',
                'message': f"Latency: {metrics['latency_ms']}ms"
            })

        return alerts

    def send_alerts(self, alerts):
        """Send alerts to monitoring system."""
        for alert in alerts:
            if alert['level'] == 'critical':
                self.send_critical_alert(alert)
            else:
                self.send_warning_alert(alert)
```

## Performance Benchmarking

### Automated Benchmarking

```python
import pytest

@pytest.mark.benchmark
class TestPerformance:
    """Performance benchmarks."""

    def test_database_query_performance(self, benchmark):
        """Benchmark database query."""

        def run_query():
            return database.query("SELECT * FROM large_table")

        result = benchmark(run_query)

        # Assert performance targets
        assert result.stats.mean < 0.1  # < 100ms

    def test_api_response_time(self, benchmark):
        """Benchmark API endpoint."""

        def call_api():
            return requests.get('http://localhost:8000/api/data')

        result = benchmark(call_api)

        assert result.stats.median < 0.05  # < 50ms
```

## Continuous Monitoring

### Long-Running Performance Monitoring

```python
class ContinuousMonitor:
    """Monitor performance continuously."""

    def start_monitoring(self, interval=60):
        """Start monitoring loop."""

        import threading

        monitor_thread = threading.Thread(
            target=self._monitor_loop,
            args=(interval,),
            daemon=True
        )
        monitor_thread.start()

    def _monitor_loop(self, interval):
        """Monitoring loop."""

        while True:
            metrics = self.collect_metrics()
            self.store_metrics(metrics)

            # Check for anomalies
            is_anomaly, deviations = self.detect_anomalies(metrics)

            if is_anomaly:
                self.send_alert(deviations)

            time.sleep(interval)

    def collect_metrics(self):
        """Collect current performance metrics."""
        return {
            'cpu': psutil.cpu_percent(),
            'memory': psutil.virtual_memory().percent,
            'disk': psutil.disk_usage('/').percent,
            'network': psutil.net_io_counters()
        }
```

---

**Related Tools**: Prometheus, Grafana, Datadog, New Relic
