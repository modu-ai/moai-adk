# Monitoring Examples

Practical implementations for metrics, logging, and tracing.

---

## Prometheus Metrics

### Python/FastAPI

```python
from prometheus_client import Counter, Histogram, Gauge, generate_latest
from fastapi import FastAPI, Response

app = FastAPI()

# Metrics
http_requests_total = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)

http_request_duration_seconds = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration',
    ['method', 'endpoint']
)

active_requests = Gauge(
    'http_requests_active',
    'Active HTTP requests'
)

# Middleware
@app.middleware("http")
async def monitor_requests(request, call_next):
   active_requests.inc()

    with http_request_duration_seconds.labels(
        method=request.method,
        endpoint=request.url.path
    ).time():
        response = await call_next(request)

    http_requests_total.labels(
        method=request.method,
        endpoint=request.url.path,
        status=response.status_code
    ).inc()

    active_requests.dec()
    return response

# Metrics endpoint
@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type="text/plain")
```

### Node.js/Express

```javascript
const client = require("prom-client");
const express = require("express");

const app = express();

// Create metrics
const httpRequestsTotal = new client.Counter({
  name: "http_requests_total",
  help: "Total HTTP requests",
  labelNames: ["method", "endpoint", "status"],
});

const httpRequestDuration = new client.Histogram({
  name: "http_request_duration_seconds",
  help: "HTTP request duration",
  labelNames: ["method", "endpoint"],
});

// Middleware
app.use((req, res, next) => {
  const end = httpRequestDuration.startTimer({
    method: req.method,
    endpoint: req.path,
  });

  res.on("finish", () => {
    end();
    httpRequestsTotal.inc({
      method: req.method,
      endpoint: req.path,
      status: res.statusCode,
    });
  });

  next();
});

// Metrics endpoint
app.get("/metrics", async (req, res) => {
  res.set("Content-Type", client.register.contentType);
  res.end(await client.register.metrics());
});
```

---

## Structured Logging

### Python

```python
import logging
import json
from datetime import datetime

class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_data = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
        }

        # Add extra fields
        if hasattr(record, 'user_id'):
            log_data['user_id'] = record.user_id

        # Add exception info
        if record.exc_info:
            log_data['exception'] = self.formatException(record.exc_info)

        return json.dumps(log_data)

# Setup
logger = logging.getLogger(__name__)
handler = logging.StreamHandler()
handler.setFormatter(JSONFormatter())
logger.addHandler(handler)
logger.setLevel(logging.INFO)

# Usage
logger.info("User logged in", extra={"user_id": 123})
logger.error("Payment failed", extra={"order_id": 456, "amount": 99.99})
```

### Node.js (Winston)

```javascript
const winston = require("winston");

const logger = winston.createLogger({
  level: "info",
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  transports: [
    new winston.transports.Console(),
    new winston.transports.File({ filename: "app.log" }),
  ],
});

// Usage
logger.info("User logged in", { user_id: 123 });
logger.error("Payment failed", { order_id: 456, amount: 99.99 });
```

---

## Distributed Tracing

### OpenTelemetry (Python)

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger.thrift import JaegerExporter

# Setup
trace.set_tracer_provider(TracerProvider())
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

tracer = trace.get_tracer(__name__)

# Usage
@app.post("/orders")
async def create_order(order_data):
    with tracer.start_as_current_span("create-order") as span:
        span.set_attribute("order.amount", order_data.amount)

        # Child span
        with tracer.start_as_current_span("validate-payment"):
            payment_valid = await validate_payment(order_data.payment)

        with tracer.start_as_current_span("save-to-db"):
            order = await save_order(order_data)

        return {"order_id": order.id}
```

---

## Alerting

### Prometheus Alertmanager

```yaml
# alert.rules.yml
groups:
  - name: api_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "High error rate on {{ $labels.endpoint }}"
          description: "Error rate is {{ $value }} req/s"

      - alert: HighLatency
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency on {{ $labels.endpoint }}"
          description: "p95 latency is {{ $value }}s"
```

### Alertmanager Config

```yaml
# alertmanager.yml
route:
  receiver: "slack-notifications"
  group_by: ["alertname", "severity"]
  group_wait: 10s
  group_interval: 5m
  repeat_interval: 4h

receivers:
  - name: "slack-notifications"
    slack_configs:
      - api_url: "https://hooks.slack.com/services/..."
        channel: "#alerts"
        title: "{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"
```

---

## Grafana Dashboards

### Dashboard JSON (Simplified)

```json
{
  "dashboard": {
    "title": "API Monitoring",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Latency (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration_seconds)"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

---

**See also**: [reference.md](./reference.md) for PromQL query reference.
