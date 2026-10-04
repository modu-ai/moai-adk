# Real-Time Performance Monitoring

> Module: Real-time performance monitoring system with alerting
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: the host language's process-introspection API, background-task support

## Core Implementation

### RealTimeMonitor Class

Process metrics (cpu_percent, memory, open files, threads, context switches) are read via the host language's process/OS API (Go `runtime`/`gopsutil`, Python `psutil`, Node `process`/`os`, Rust `sysinfo`, Java `ManagementFactory`).

```text
record PerformanceSnapshot:
    timestamp:       timestamp
    cpu_percent:     float
    memory_mb:       float
    memory_percent:  float
    open_files:      int
    threads:         int
    context_switches:int
    custom_metrics:  Map<text, float>   # default {}

class RealTimeMonitor(sampling_interval = 1.0):
    sampling_interval
    is_monitoring = false
    snapshots = bounded_deque(maxlen=1000)   # keep last 1000 snapshots
    callbacks  = []
    alerts     = []

    start_monitoring():
        if is_monitoring: return
        is_monitoring = true
        spawn_background_task(monitor_loop())   # daemon thread / goroutine / async task

    stop_monitoring():
        is_monitoring = false
        join the monitor task (timeout ~2s)

    monitor_loop():
        while is_monitoring:
            try:
                snapshot = PerformanceSnapshot(
                    timestamp=now(),
                    cpu_percent=process.cpu_percent(),
                    memory_mb=process.rss() / MB,
                    memory_percent=process.memory_percent(),
                    open_files=len(process.open_files()),
                    threads=process.num_threads(),
                    context_switches=process.context_switches())
                for callback in callbacks:
                    try: snapshot.custom_metrics.update(callback())
                    except e: log("Custom metric callback error: " + e)
                snapshots.append(snapshot)
                check_alerts(snapshot)
                sleep(sampling_interval)
            except e:
                log("Monitoring error: " + e)
                sleep(sampling_interval)

    add_callback(callback):
        callbacks.append(callback)

    check_alerts(snapshot):
        alerts = []
        if snapshot.cpu_percent > 90:
            alerts.append({ type:"high_cpu", message:"High CPU usage: " + snapshot.cpu_percent })
        if snapshot.memory_percent > 85:
            alerts.append({ type:"high_memory", message:"High memory usage: " + snapshot.memory_percent })
        if snapshot.open_files > 1000:
            alerts.append({ type:"file_handle_leak", message:"High open files: " + snapshot.open_files })
        alerts.extend(alerts)

    get_recent_snapshots(count = 100):
        return last `count` of snapshots

    get_average_metrics(duration_minutes = 5):
        cutoff = now() - duration_minutes*60
        recent = [s for s in snapshots if s.timestamp >= cutoff]
        if recent is empty: return {}
        return {
            avg_cpu_percent:    mean(s.cpu_percent    for s in recent),
            avg_memory_mb:      mean(s.memory_mb      for s in recent),
            avg_memory_percent: mean(s.memory_percent for s in recent),
            avg_open_files:     mean(s.open_files     for s in recent),
            avg_threads:        mean(s.threads        for s in recent)
        }
```

## Usage Examples

```text
# Real-time monitoring example
monitor = RealTimeMonitor(sampling_interval=0.5)
monitor.start_monitoring()

# Add a custom metrics callback
custom_metrics():
    return { custom_counter: some_global_counter, queue_size: len(some_queue) }
monitor.add_callback(custom_metrics)

# Run the application while monitoring
# ... your application code ...

# Stop monitoring and read results
monitor.stop_monitoring()
recent_snapshots = monitor.get_recent_snapshots(10)
avg_metrics = monitor.get_average_metrics(5)
print("Average CPU: " + avg_metrics.avg_cpu_percent + "%")
print("Average Memory: " + avg_metrics.avg_memory_mb + "MB")
```

## Best Practices

1. **Sampling Interval**: Choose appropriate intervals (0.5-2.0 seconds) to balance overhead and granularity
2. **Snapshot Limit**: Use a bounded deque to prevent memory growth
3. **Background Task**: Run monitoring in a separate daemon task/goroutine
4. **Custom Metrics**: Add domain-specific metrics via callbacks
5. **Alert Thresholds**: Configure thresholds based on application requirements

---

Related: [Profiler Core](./profiler-core.md) | [Bottleneck Detection](./bottleneck-detection.md)
