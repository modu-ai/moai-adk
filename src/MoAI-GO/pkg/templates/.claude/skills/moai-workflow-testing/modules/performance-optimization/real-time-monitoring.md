# Real-Time Performance Monitoring

> Module: Real-time performance monitoring system with alerting
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: psutil, threading, asyncio

## Core Implementation

### RealTimeMonitor Class

```python
import threading
import time
from typing import Dict, List, Callable
from dataclasses import dataclass, field
from collections import deque
import psutil

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

class RealTimeMonitor:
    """Real-time performance monitoring system."""

    def __init__(self, sampling_interval: float = 1.0):
        self.sampling_interval = sampling_interval
        self.is_monitoring = False
        self.monitor_thread = None
        self.snapshots = deque(maxlen=1000)  # Keep last 1000 snapshots
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
```

## Usage Examples

```python
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
```

## Best Practices

1. **Sampling Interval**: Choose appropriate intervals (0.5-2.0 seconds) to balance overhead and granularity
2. **Snapshot Limit**: Use deque with maxlen to prevent memory growth
3. **Thread Safety**: Monitoring runs in separate daemon thread
4. **Custom Metrics**: Add domain-specific metrics via callbacks
5. **Alert Thresholds**: Configure thresholds based on application requirements

---

Related: [Profiler Core](./profiler-core.md) | [Bottleneck Detection](./bottleneck-detection.md)
