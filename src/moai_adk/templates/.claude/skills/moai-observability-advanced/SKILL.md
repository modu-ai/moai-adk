---
name: moai-observability-advanced
version: 3.0.0
updated: "2025-11-19"
status: stable
description: Advanced observability patterns with OpenTelemetry, distributed tracing, eBPF monitoring, SLO/SLI implementation, and production strategies. Use when implementing observability, monitoring, or distributed tracing.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Advanced Observability & Monitoring

Production-grade observability with OpenTelemetry, distributed tracing, and SLO/SLI implementation.

## Quick Start

**OpenTelemetry in 5 Minutes**:

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger import JaegerExporter

# Setup tracing
trace.set_tracer_provider(TracerProvider())
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

# Use tracer
tracer = trace.get_tracer(__name__)

with tracer.start_as_current_span("process_request"):
    # Your code here
    result = do_work()
```

**Auto-triggers**: observability, monitoring, tracing, OpenTelemetry, metrics, logs, SLO, SLI

---

## Three Pillars of Observability

### 1. Metrics (What's happening)

- **USE Method**: Utilization, Saturation, Errors
- **RED Method**: Rate, Errors, Duration
- **Golden Signals**: Latency, Traffic, Errors, Saturation

### 2. Logs (Event records)

- Structured logging (JSON)
- Correlation IDs
- Log levels and sampling

### 3. Traces (Request flow)

- Distributed tracing
- Span relationships
- Performance bottlenecks

---

## OpenTelemetry Core

### Instrumentation

**Automatic** (recommended):

```python
from opentelemetry.instrumentation.flask import FlaskInstrument or
from opentelemetry.instrumentation.requests import RequestsInstrumentor

# Auto-instrument Flask
FlaskInstrumentor().instrument_app(app)

# Auto-instrument HTTP requests
RequestsInstrumentor().instrument()
```

**Manual** (fine-grained control):

```python
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

def process_order(order_id):
    with tracer.start_as_current_span("process_order") as span:
        span.set_attribute("order.id", order_id)
        span.set_attribute("order.value", calculate_value())

        # Nested span
        with tracer.start_as_current_span("validate_payment"):
            validate(order_id)
```

### Context Propagation

```python
from opentelemetry.propagate import inject, extract

# Inject context into headers
headers = {}
inject(headers)
requests.post(url, headers=headers)

# Extract context from headers
ctx = extract(request.headers)
with trace.use_context(ctx):
    process_request()
```

---

## SLO/SLI Implementation

### Service Level Indicators (SLIs)

Common SLIs:

- **Availability**: % of successful requests
- **Latency**: % of requests < threshold
- **Throughput**: Requests per second

### Service Level Objectives (SLOs)

**Example SLOs**:

```yaml
api_service:
  availability:
    target: 99.9% # 3 nines
    window: 30 days

  latency:
    target: 95% of requests < 200ms
    window: 30 days

  error_budget:
    allowed_failures: 0.1% # 43.2 minutes/month
```

### Error Budget Calculation

```python
def calculate_error_budget(slo_target, time_window_seconds):
    """Calculate remaining error budget"""
    allowed_downtime = time_window_seconds * (1 - slo_target)
    return allowed_downtime

# Example: 99.9% SLO for 30 days
budget = calculate_error_budget(0.999, 30 * 24 * 60 * 60)
print(f"Allowed downtime: {budget / 60:.2f} minutes")  # 43.2 minutes
```

---

## Distributed Tracing

### Trace Structure

```
Trace ID: abc123
├─ Span: API Gateway (200ms)
│  ├─ Span: Auth Service (50ms)
│  └─ Span: Order Service (150ms)
│     ├─ Span: Database Query (80ms)
│     └─ Span: Payment Gateway (70ms)
```

### Trace Analysis

**Critical Path**: Longest span chain determines total latency
**Parallel Work**: Concurrent spans reduce total time
**Bottlenecks**: Spans with high duration or error rates

### Jaeger Query Example

```python
from jaeger_client import Config

config = Config(
    config={
        'sampler': {'type': 'const', 'param': 1},
        'local_agent': {'reporting_host': 'localhost'},
    },
    service_name='my-service',
)
tracer = config.initialize_tracer()
```

---

## Metrics Collection

### Prometheus Integration

```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# Define metrics
request_count = Counter('http_requests_total', 'Total requests', ['method', 'endpoint'])
request_duration = Histogram('http_request_duration_seconds', 'Request duration')
active_users = Gauge('active_users', 'Currently active users')

# Use metrics
@request_duration.time()
def handle_request():
    request_count.labels(method='GET', endpoint='/api/users').inc()
    # Process request
    active_users.set(get_active_user_count())

# Expose metrics
start_http_server(8000)  # Metrics at :8000/metrics
```

### Grafana Dashboards

**Key Panels**:

- Request rate (requests/sec)
- Error rate (%)
- P50, P95, P99 latencies
- Resource utilization (CPU, memory)
- SLI/SLO tracking

---

## eBPF Monitoring

**Use cases**:

- Low-overhead kernel-level monitoring
- Network traffic analysis
- Security monitoring
- Performance profiling

**Tools**:

- **Pixie**: Kubernetes observability
- **Cilium**: Network monitoring
- **BCC**: eBPF toolkit

**Example (network monitoring)**:

```bash
# Install Pixie
px deploy

# Monitor HTTP traffic
px live http_data
```

---

## Log Aggregation

### Structured Logging

```python
import structlog

logger = structlog.get_logger()

logger.info(
    "order_processed",
    order_id="12345",
    user_id="user_789",
    amount=99.99,
    duration_ms=150
)
# Output: {"event": "order_processed", "order_id": "12345", ...}
```

### ELK Stack Integration

```python
from python_logging_elk import ElkHandler

handler = ElkHandler(
    host='elasticsearch.example.com',
    port=9200,
    index='application-logs'
)
logger.addHandler(handler)
```

---

## Alerting Best Practices

### Alert Rule Design

**Good Alert**:

- Actionable (clear next steps)
- Urgent (requires immediate attention)
- Specific (narrow scope)

**Example**:

```yaml
- alert: HighErrorRate
  expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
  for: 5m
  annotations:
    summary: "Error rate above 5% for 5 minutes"
    description: "{{ $value }}% of requests are failing"
  labels:
    severity: critical
```

---

## Performance Analysis

### RED Method (Requests)

- **Rate**: Requests per second
- **Errors**: Failed requests percentage
- **Duration**: Request latency (p50, p95, p99)

### USE Method (Resources)

- **Utilization**: % time resource busy
- **Saturation**: Queue length or wait time
- **Errors**: Error count

---

## Production Checklist

✅ **DO**:

- Implement all three pillars (metrics, logs, traces)
- Define SLOs based on user impact
- Use structured logging (JSON)
- Sample traces (1-10% in production)
- Set up alerts for SLO violations
- Create runbooks for alerts
- Monitor error budgets

❌ **DON'T**:

- Log sensitive data (PII, passwords)
- Trace 100% in production (sample!)
- Create alerts without runbooks
- Ignore tail latencies (p99)
- Skip correlation IDs

---

## Common Patterns

For detailed implementation:

- **[examples.md](examples.md)**: Complete setups, custom metrics, alert rules
- **[reference.md](reference.md)**: OpenTelemetry API, exporters, best practices

**Related Skills**:

- `moai-essentials-perf`: Performance optimization
- `moai-cloud-aws-advanced`: AWS CloudWatch integration
- `moai-cloud-gcp-advanced`: Google Cloud Monitoring

---

**Tools**: OpenTelemetry, Prometheus, Grafana, Jaeger, ELK, Pixie

**Version**: 3.0.0  
**Last Updated**: 2025-11-19  
**Status**: Production Ready
