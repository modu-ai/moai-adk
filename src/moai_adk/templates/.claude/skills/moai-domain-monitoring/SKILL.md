---
name: moai-domain-monitoring
description: Enterprise-grade monitoring and observability platform with Prometheus 2.55.x, Grafana 11.x Scenes, OpenTelemetry 1.33.x, Loki 3.0+ logs, Jaeger distributed tracing, Elasticsearch 8.x, Sentry error tracking, and APM platforms; activates for metrics monitoring, log aggregation, distributed tracing, error tracking, SLO/SLI implementation, and production observability architecture.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# 🔍 Enterprise Monitoring & Observability Platform — v4.0

## 🎯 Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | On-demand for observability, monitoring, and tracing requests |
| **Trigger cues** | Prometheus, Grafana, metrics, logs, tracing, OpenTelemetry, Jaeger, Loki, Elasticsearch, Sentry, APM, SLO, error tracking, distributed tracing, alerting |

---

## Enterprise Observability Stack v4.0 — 2025 November Stable Versions

### Technology Domains

**Metrics & Alerting**:
- Prometheus 2.55.x (legacy) / 3.7.x (latest)
- Grafana 11.x with Scenes-powered modular dashboards
- Alertmanager 0.27.x
- VictorOps incident management
- Thanos 0.36.x for HA Prometheus

**Distributed Tracing**:
- OpenTelemetry 1.33.x (production-ready SDKs)
- Jaeger 1.62.x with multi-tenancy support
- Context propagation with W3C Trace Context

**Log Aggregation**:
- Loki 3.x with Bloom filters (70-90% query acceleration)
- ELK Stack: Elasticsearch 8.x, Logstash 8.x, Kibana 8.x
- Native OTLP ingestion (no intermediate exporters)

**Application Performance Monitoring**:
- Sentry 24.x (error tracking and performance monitoring)
- New Relic (AI-powered anomaly detection)
- Datadog (infrastructure-first observability)
- Elastic APM 8.x (code-level diagnostics)

**Infrastructure Monitoring**:
- Prometheus Node Exporter 1.x
- cAdvisor (container metrics)
- Telegraf 1.x (metrics collection)
- netdata (real-time system monitoring)

---

## 1. Metrics Collection & Visualization

### Prometheus Scrape Configuration

**Service Discovery**:
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'production'
    environment: 'prod'

scrape_configs:
  # Kubernetes service discovery
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - monitoring
            - production
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: 'true'
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: '(.+)'
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: '([^:]+)(?::\d+)?;(\d+)'
        replacement: '$1:$2'
        target_label: __address__

  # Consul service discovery
  - job_name: 'consul-services'
    consul_sd_configs:
      - server: 'localhost:8500'
    relabel_configs:
      - source_labels: [__meta_consul_service]
        target_label: service
      - source_labels: [__meta_consul_dc]
        target_label: datacenter
```

**Custom Metric Cardinality Management**:
```yaml
metric_relabeling:
  # Drop high-cardinality labels
  - source_labels: [__name__]
    regex: 'request_duration_seconds'
    target_label: __tmp_high_cardinality
  - source_labels: [user_id]
    action: drop
    regex: '.*'
  
  # Relabel to reduce cardinality
  - source_labels: [http_path]
    action: replace
    regex: '/api/users/[0-9]+'
    replacement: '/api/users/{id}'
    target_label: http_path
```

---

### Grafana 11.x Scenes-Powered Dashboards

**Modular Dashboard with Scenes API**:
```json
{
  "title": "Production Observability v4.0",
  "uid": "prod-observability-v4",
  "version": 1,
  "panels": [
    {
      "type": "stat",
      "title": "Request Rate (5m avg)",
      "targets": [
        {
          "expr": "rate(http_requests_total[5m])",
          "legendFormat": "{{job}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "color": {
            "mode": "gradient-gauge"
          },
          "thresholds": {
            "mode": "percentage",
            "steps": [
              {"color": "green", "value": null},
              {"color": "yellow", "value": 70},
              {"color": "red", "value": 90}
            ]
          }
        }
      }
    },
    {
      "type": "timeseries",
      "title": "Response Time p99",
      "targets": [
        {
          "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))"
        }
      ]
    }
  ]
}
```

**Scenes-Based Dynamic Dashboard (TypeScript)**:
```typescript
import { SceneApp, SceneFlexItem, SceneFlexLayout } from '@grafana/scenes';
import { PanelBuilders } from '@grafana/scenes';

export function createObservabilityDashboard() {
  return new SceneApp({
    title: 'Production Observability',
    body: new SceneFlexLayout({
      children: [
        new SceneFlexItem({
          width: '50%',
          body: PanelBuilders.stat()
            .setTitle('Request Rate (5m avg)')
            .setUnit('reqps')
            .build()
        }),
        new SceneFlexItem({
          width: '50%',
          body: PanelBuilders.timeseries()
            .setTitle('Response Time p99')
            .setUnit('ms')
            .build()
        })
      ]
    })
  });
}
```

**Key Grafana 11.x Features**:
- **Scenes Library**: Modular, reusable dashboard components
- **Edit Mode**: Simplified dashboard discovery and editing
- **Fixed Time Picker**: Persistent time range while scrolling
- **Nested Subfolders**: Better dashboard organization
- **PDF Export**: 200-panel dashboard in 11 seconds (vs 7+ minutes)

---

## 2. Distributed Tracing with OpenTelemetry & Jaeger

### OpenTelemetry Node.js Instrumentation

**Setup and Configuration**:
```javascript
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { JaegerExporter } = require('@opentelemetry/exporter-trace-jaeger');
const { BatchSpanProcessor } = require('@opentelemetry/sdk-trace-node');

// Initialize OpenTelemetry SDK
const sdk = new NodeSDK({
  traceExporter: new JaegerExporter({
    endpoint: 'http://jaeger:14268/api/traces',
  }),
  instrumentations: [getNodeAutoInstrumentations()],
  serviceName: 'my-app',
  serviceVersion: '4.0.0',
});

sdk.start();
console.log('OpenTelemetry SDK started');
```

**Instrumentation with Context Propagation**:
```javascript
const { trace, context } = require('@opentelemetry/api');
const { W3CTraceContextPropagator } = require('@opentelemetry/core');

const tracer = trace.getTracer('my-service', '1.0.0');

// Express middleware for automatic span creation
app.use((req, res, next) => {
  const span = tracer.startSpan('http.request', {
    attributes: {
      'http.method': req.method,
      'http.url': req.url,
      'http.target': req.path,
      'http.host': req.hostname,
      'http.scheme': req.protocol,
      'http.status_code': res.statusCode,
    },
  });

  context.with(trace.setSpan(context.active(), span), () => {
    res.on('finish', () => {
      span.setStatus({ code: res.statusCode < 400 ? 0 : 2 });
      span.end();
    });
    next();
  });
});

// Custom span creation
app.post('/api/users', (req, res) => {
  const span = tracer.startSpan('user.create');
  span.addEvent('user_validation_started');
  
  try {
    // Validate user
    span.addEvent('user_validation_passed');
    
    // Create user
    span.addEvent('database_write_started');
    const userId = db.createUser(req.body);
    span.addEvent('database_write_completed');
    
    span.setStatus({ code: 0 });
    res.json({ userId });
  } catch (error) {
    span.recordException(error);
    span.setStatus({ code: 2, message: error.message });
    res.status(500).json({ error: error.message });
  } finally {
    span.end();
  }
});
```

### Jaeger Distributed Tracing

**Jaeger Configuration (Docker Compose)**:
```yaml
version: '3'
services:
  jaeger:
    image: jaegertracing/all-in-one:1.62
    container_name: jaeger
    environment:
      COLLECTOR_OTLP_ENABLED: 'true'
      MEMORY_MAX_TRACES: '10000'
    ports:
      - "6831:6831/udp"  # Jaeger agent compact
      - "14268:14268"     # Jaeger collector
      - "16686:16686"     # Jaeger UI
    volumes:
      - ./jaeger-config.yml:/etc/jaeger/config.yml
```

**Jaeger Query Patterns**:
```
# Find error traces
status:error AND duration:>100ms

# Trace specific service
service.name:my-app AND http.status_code:500

# Trace database operations
span.kind:CLIENT AND db.operation:SELECT
```

**Multi-Tenancy in Jaeger v2**:
```yaml
# jaeger-config.yml
collectors:
  otlp:
    enabled: true
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

storage:
  type: elasticsearch
  elasticsearch:
    server_urls: http://elasticsearch:9200
    index_prefix: jaeger

multitenancy:
  enabled: true
  tenants:
    - production
    - staging
    - development
```

---

## 3. Log Aggregation with Loki & ELK Stack

### Loki 3.x with Bloom Filters

**LogQL Queries with Bloom Optimization**:
```logql
# Basic query (benefits from Bloom filters)
{job="my-app", level="error"} | json

# Error rate with aggregation
sum(rate({job="my-app", level="error"} [5m])) by (service)

# Complex query with Bloom acceleration
{job="my-app"} 
  | json response_time=response_time_ms 
  | response_time > 1000 
  | stats avg(response_time) by (endpoint)

# Pattern matching (Bloom optimized)
{job="api-server"} |= "ERROR" |= "database"
```

**Loki Configuration with OTLP Support**:
```yaml
# loki-config.yml
auth_enabled: false

ingester:
  chunk_idle_period: 3m
  max_chunk_age: 1h
  max_streams_per_user: 10000
  max_global_streams_per_user: 10000

schema_config:
  configs:
    - from: 2025-01-01
      store: tsdb
      object_store: s3
      schema: v13

storage_config:
  s3:
    s3: s3://bucket-name
    endpoint: s3.amazonaws.com

limits_config:
  # Cost optimization
  ingestion_rate_mb: 100
  ingestion_burst_size_mb: 200
  max_cache_freshness_per_query: 10m
  reject_old_samples: true
  reject_old_samples_max_age: 168h

# Native OTLP support (no intermediate exporter needed)
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
```

**Bloom Filters Benefits**:
- 70-90% reduction in chunks processed
- Faster error trace retrieval
- Lower query latency on specific text searches
- Cost reduction through efficient filtering

### ELK Stack Integration

**Elasticsearch Index Strategy**:
```bash
# Create time-based indices for cost optimization
PUT /logs-app-prod-2025.11.12
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "index.lifecycle.name": "logs-policy",
    "index.lifecycle.rollover_alias": "logs-app-prod"
  },
  "mappings": {
    "properties": {
      "@timestamp": { "type": "date" },
      "message": { "type": "text" },
      "level": { "type": "keyword" },
      "service": { "type": "keyword" },
      "trace_id": { "type": "keyword" },
      "span_id": { "type": "keyword" },
      "duration_ms": { "type": "integer" }
    }
  }
}
```

**Kibana Discovery & Alerting**:
```json
{
  "query": {
    "bool": {
      "must": [
        { "range": { "@timestamp": { "gte": "now-1h" } } },
        { "match": { "level": "ERROR" } },
        { "match": { "service": "payment-api" } }
      ],
      "filter": [
        { "term": { "environment": "production" } }
      ]
    }
  },
  "aggs": {
    "errors_by_endpoint": {
      "terms": {
        "field": "endpoint.keyword",
        "size": 20
      },
      "aggs": {
        "error_count": { "value_count": { "field": "trace_id" } }
      }
    }
  }
}
```

---

## 4. Application Performance Monitoring

### Sentry Error Tracking & Performance Monitoring

**SDK Integration with Sampling**:
```javascript
const Sentry = require("@sentry/node");
const { ProfilingIntegration } = require("@sentry/profiling-node");

Sentry.init({
  dsn: process.env.SENTRY_DSN,
  environment: process.env.NODE_ENV,
  release: "my-app@4.0.0",
  tracesSampleRate: 0.1,  // 10% transaction sample rate
  profilesSampleRate: 0.1, // 10% profiling sample rate
  integrations: [
    new Sentry.Integrations.OnUncaughtException(),
    new Sentry.Integrations.OnUnhandledRejection(),
    new ProfilingIntegration(),
  ],
  beforeSend(event, hint) {
    // Filter 404 errors
    if (event.exception && event.exception[0].value === 'NotFound') {
      return null;
    }
    return event;
  }
});

// Express middleware
app.use(Sentry.Handlers.requestHandler());
app.use(Sentry.Handlers.errorHandler());
```

**Custom Error Context**:
```javascript
Sentry.captureException(error, {
  level: 'error',
  tags: {
    'payment.processor': 'stripe',
    'user.tier': 'premium'
  },
  contexts: {
    payment: {
      amount: 99.99,
      currency: 'USD',
      processor: 'stripe'
    }
  }
});
```

### New Relic & Datadog APM

**New Relic Integration (Node.js)**:
```javascript
const newrelic = require('newrelic');

// Automatic instrumentation for HTTP, databases, caches
app.post('/api/checkout', (req, res) => {
  const transaction = newrelic.startSegment('checkout', true, () => {
    // Business logic
    return processCheckout(req.body);
  });
});

// Custom metrics
newrelic.recordCustomEvent('CheckoutCompleted', {
  amount: 99.99,
  currency: 'USD',
  duration: 1234
});
```

**Datadog APM (Python)**:
```python
from ddtrace import tracer, patch_all

patch_all()

@tracer.wrap()
def process_payment(payment_data):
    with tracer.trace("payment.processing"):
        # Process payment
        pass
    
    # Custom span
    with tracer.trace("payment.verification", tags={
        'payment.id': payment_data['id'],
        'payment.amount': payment_data['amount']
    }):
        verify_payment(payment_data)
```

---

## 5. Alert Rules & SLO/SLI Implementation

### Prometheus Alert Rules

**Threshold-Based Alerts**:
```yaml
groups:
  - name: application_alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"
      
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 1.0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High p99 latency"
```

**Composite Alert Rules**:
```yaml
- alert: ServiceDegradation
  expr: |
    (rate(http_requests_total{status=~"5.."}[5m]) > 0.01) 
    AND 
    (histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5)
  for: 5m
  annotations:
    summary: "Service experiencing degradation"
```

### SLO/SLI with Error Budgets

**SLO Definition**:
```yaml
# SLO: 99.9% availability (0.1% error budget)
groups:
  - name: slo_availability
    interval: 30s
    rules:
      - record: slo:http_availability:ratio
        expr: |
          sum(rate(http_requests_total{status!~"5.."}[30d]))
          /
          sum(rate(http_requests_total[30d]))
      
      - alert: SLOErrorBudgetBurn
        expr: |
          (1 - slo:http_availability:ratio) 
          * 30 
          > 0.001  # 30-day error budget
        for: 1h
        annotations:
          summary: "Error budget burn rate is high"
```

**SLI Metrics**:
- Availability SLI: (successful_requests / total_requests)
- Latency SLI: (requests_under_threshold / total_requests)
- Durability SLI: (non-corrupted_data / total_data)

---

## 6. Cost Optimization & Cardinality Management

### High-Cardinality Label Prevention

**Cardinality Explosion Prevention**:
```yaml
# Bad: Creates M * N * O cardinality
http_request_duration_seconds_bucket{
  job="api",
  method="GET",
  path="/api/users/123",     # DO NOT use user IDs
  status="200",
  instance="server1"
}

# Good: Use relabeling to reduce cardinality
http_request_duration_seconds_bucket{
  job="api",
  method="GET",
  path="/api/users/{id}",    # Use placeholders
  status="200",
  instance="server1"
}
```

**Metrics Retention Policy**:
```yaml
# prometheus.yml
global:
  external_labels:
    cluster: prod

remote_write:
  - url: http://thanos-receive:19291/api/v1/receive
    write_relabel_configs:
      # Keep only critical metrics for long-term storage
      - source_labels: [__name__]
        regex: '(up|http_requests_total|http_request_duration_seconds)'
        action: keep
      # Reduce resolution for old data
      - source_labels: [__name__]
        regex: '.*'
        action: keep
        if_false: true
```

### Storage Cost Optimization

**Loki Cost Optimization**:
```yaml
# Reduce ingestion costs
ingestion_rate_mb: 100          # Limit ingestion rate
max_streams_per_user: 5000      # Limit streams per tenant

# TTL management
retention_period: 720h          # 30-day retention

# Index cost reduction
schema_config:
  configs:
    - from: 2025-11-01
      store: tsdb
      # Use more compact storage
      index:
        prefix: loki_index_
```

**Elasticsearch Cost Optimization**:
```bash
# Use ILM (Index Lifecycle Management)
PUT _ilm/policy/logs-policy
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0d",
        "actions": {
          "rollover": { "max_size": "50GB", "max_age": "1d" }
        }
      },
      "warm": {
        "min_age": "3d",
        "actions": {
          "set_priority": { "priority": 50 },
          "forcemerge": { "max_num_segments": 1 }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": { "delete": {} }
      }
    }
  }
}
```

---

## 7. Infrastructure Monitoring

### Prometheus Node Exporter Metrics

**Essential Node Metrics**:
```
# CPU utilization
node_cpu_seconds_total{mode="user"}
rate(node_cpu_seconds_total{mode!="idle"}[5m]) * 100

# Memory usage
node_memory_MemAvailable_bytes
100 * (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))

# Disk I/O
rate(node_disk_io_time_seconds_total[5m])
rate(node_disk_read_bytes_total[5m]) / 1024 / 1024  # MB/s

# Network
rate(node_network_receive_bytes_total[5m]) / 1024 / 1024  # MB/s
```

### Container Monitoring with cAdvisor

**cAdvisor Prometheus Metrics**:
```yaml
# container_cpu_usage_seconds_total
# container_memory_usage_bytes
# container_network_receive_bytes_total
# container_fs_usage_bytes
```

---

## 8. Production Best Practices

**High Availability Architecture**:
1. **Multi-region Prometheus**: With Thanos for long-term storage
2. **Jaeger Distributed Deployment**: With Elasticsearch backend
3. **Loki HA Setup**: With Consul for peer discovery
4. **Load-balanced Grafana**: With shared datasources

**Monitoring the Monitors**:
```yaml
- alert: PrometheusHighMemoryUsage
  expr: process_resident_memory_bytes / 1024 / 1024 > 500
  
- alert: JaegerBackendDown
  expr: up{job="jaeger"} == 0
  
- alert: LokiIndexingSlowdown
  expr: loki_chunk_store_index_lookups_total > threshold
```

**Secure Observability Stack**:
- Enable TLS/mTLS between components
- Implement RBAC in Jaeger multi-tenancy
- Encrypt data at rest (Elasticsearch)
- Use private networks for backend communication

---

## Key Enhancements in v4.0

- **Grafana 11.x Scenes**: Modular, maintainable dashboard code
- **Loki Bloom Filters**: 70-90% query acceleration
- **OpenTelemetry 1.33**: Production-ready distributed tracing
- **Jaeger Multi-Tenancy**: Enterprise-scale deployments
- **Native OTLP Ingestion**: Simplified observability stack
- **Cost Optimization**: Cardinality management and retention policies

---

## Skill Activation Pattern

This skill activates automatically when you:
- Need Prometheus, Grafana, or Alertmanager configuration
- Implement distributed tracing with OpenTelemetry/Jaeger
- Set up log aggregation with Loki or ELK Stack
- Configure Sentry, New Relic, or Datadog APM
- Implement SLO/SLI with error budgets
- Optimize monitoring costs and cardinality
- Build production observability platforms

Invoke explicitly: `Skill("moai-domain-monitoring")`

