---
name: "moai-domain-monitoring"
version: "4.0.0"
description: Enterprise-grade monitoring and observability platform with Prometheus 3.7.x, Grafana 11.x Scenes, OpenTelemetry 1.33.x, Loki 3.x+ logs, Jaeger 1.62.x distributed tracing, Elasticsearch 8.17 with ILM, Sentry error tracking, and multi-cluster HA architectures; activates for metrics collection, visualization, log aggregation, distributed tracing, error tracking, SLO/SLI implementation, cost optimization, and production observability infrastructure.
allowed-tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Enterprise Monitoring & Observability Platform — v4.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Updated** | 2025-11-12 |
| **Stable Stack** | Prometheus 3.7.3, Grafana 11.3+, OpenTelemetry 1.33.x, Jaeger 1.62.x, Loki 3.x, Elasticsearch 8.17 |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | On-demand for metrics, logs, traces, SLO/SLI, APM, observability requests |
| **Trigger cues** | Prometheus, Grafana, metrics, logs, tracing, OpenTelemetry, Jaeger, Loki, Elasticsearch, Sentry, APM, SLO, error tracking, distributed tracing, alerting, Bloom filters |

---

## Stable 2025 Monitoring Stack

### Core Technologies (November 2025 Stable)

**Metrics & Visualization**:
- Prometheus 3.7.3 (latest, released Oct 29, 2025) with Remote-Write 2.0
- Prometheus 2.55.x (legacy support, rollback compatible)
- Grafana 11.3+ with Scenes-powered dashboards (GA)
- Alertmanager 0.27.x
- Thanos 0.36.x for multi-cluster HA

**Distributed Tracing**:
- OpenTelemetry 1.33.x (production-stable, May 2025)
- Jaeger 1.62.x with native OTLP support (v1.35+)
- W3C Trace Context propagation

**Log Aggregation**:
- Loki 3.x with Bloom filters (experimental, production-ready)
- Elasticsearch 8.17 with Index Lifecycle Management
- Logstash 8.x for advanced data processing
- Native OTLP ingestion

**Application Performance Monitoring**:
- Sentry 24.x (error tracking + performance)
- New Relic (code-level diagnostics)
- Datadog (infrastructure-first)
- Elastic APM 8.17

**Infrastructure**:
- Prometheus Node Exporter 1.x
- cAdvisor (container metrics)
- Telegraf 1.x (metrics collection)
- netdata (real-time system monitoring)

---

## 1. Prometheus 3.7.x: Metrics Collection & Remote Write 2.0

### Remote-Write 2.0 for High-Availability

**Benefits over Legacy Remote Write**:
- Native metadata, exemplars, created timestamps
- Native histogram support (cardinality optimization)
- String interning for 30-40% payload reduction
- Reduced CPU usage and bandwidth
- Full backward compatibility with v2.55

**Configuration with Remote-Write 2.0**:
```yaml
# prometheus.yml (3.7.x)
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'production'
    environment: 'prod'
    region: 'us-east-1'

# Remote Write 2.0 with metadata
remote_write:
  - url: http://thanos-receive:19291/api/v1/receive
    # Enable Remote Write 2.0 features
    resource_to_telemetry_conversion:
      enabled: true
    metadata_config:
      send: true
      send_interval: 60s
    write_relabel_configs:
      # Keep critical metrics
      - source_labels: [__name__]
        regex: '(up|http_requests_total|http_request_duration_seconds|process_.*)'
        action: keep

scrape_configs:
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
```

### Agent Mode (Stable in 3.0+)

**Lightweight Metrics Collection for Edge/IoT**:
```bash
# Run Prometheus in agent mode (no local storage)
prometheus --storage.agent.path=/tmp/agent \
           --storage.agent.wal-segment-size=104857600 \
           --config.file=prometheus.yml
```

---

## 2. Grafana 11.3+: Scenes-Powered Dashboards

### Scenes API for Type-Safe Dashboards

**TypeScript Dashboard with Scenes (v11.3+)**:
```typescript
import {
  SceneApp,
  SceneFlexItem,
  SceneFlexLayout,
  SceneTimeRangeState,
  PanelBuilders,
  QueryVariable,
  VariableValueSelectors,
} from '@grafana/scenes';

export function createEnterpriseObservabilityDashboard() {
  // Environment selector variable
  const envVar = new QueryVariable({
    name: 'env',
    label: 'Environment',
    value: 'production',
    options: {
      query: 'label_values(up, environment)',
    },
  });

  return new SceneApp({
    title: 'Enterprise Observability — Scenes v1',
    timeRange: new SceneTimeRangeState({
      from: 'now-6h',
      to: 'now',
    }),
    body: new SceneFlexLayout({
      direction: 'column',
      children: [
        // Variables row
        new SceneFlexItem({
          minHeight: 50,
          body: new VariableValueSelectors({
            variables: [envVar],
          }),
        }),

        // SLO Status Row
        new SceneFlexItem({
          minHeight: 200,
          body: new SceneFlexLayout({
            children: [
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Availability SLO')
                  .setUnit('percentunit')
                  .setTargets([
                    {
                      expr: 'slo:http_requests:ratio_30d{env="${env}"}',
                      refId: 'A',
                    }
                  ])
                  .setFieldConfig({
                    defaults: {
                      color: { mode: 'gradient-gauge' },
                      thresholds: {
                        mode: 'absolute',
                        steps: [
                          { color: 'red', value: 0.99 },
                          { color: 'yellow', value: 0.995 },
                          { color: 'green', value: 0.9995 },
                        ],
                      },
                    },
                  })
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Error Budget (30d)')
                  .setUnit('percentunit')
                  .setTargets([
                    {
                      expr: '(1 - slo:http_requests:ratio_30d{env="${env}"}) * 100',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Request Rate')
                  .setUnit('reqps')
                  .setTargets([
                    {
                      expr: 'sum(rate(http_requests_total{environment="${env}"}[5m]))',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Error Rate')
                  .setUnit('percent')
                  .setTargets([
                    {
                      expr: 'sum(rate(http_requests_total{environment="${env}", status=~"5.."}[5m])) / sum(rate(http_requests_total{environment="${env}"}[5m])) * 100',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
            ],
          })
        }),

        // Time-series analytics
        new SceneFlexItem({
          minHeight: 300,
          body: PanelBuilders.timeseries()
            .setTitle('Request Rate & Error Rate (5m)')
            .setUnit('short')
            .setTargets([
              {
                expr: 'rate(http_requests_total{environment="${env}"}[5m])',
                legendFormat: 'Requests {{status}}',
                refId: 'A',
              },
              {
                expr: 'rate(http_requests_total{environment="${env}", status=~"5.."}[5m])',
                legendFormat: 'Errors {{status}}',
                refId: 'B',
              }
            ])
            .build()
        }),

        // Trace distribution
        new SceneFlexItem({
          minHeight: 300,
          body: PanelBuilders.heatmap()
            .setTitle('Response Time Distribution (ms)')
            .setUnit('ms')
            .setTargets([
              {
                expr: 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{environment="${env}"}[5m])) * 1000',
                refId: 'A',
              }
            ])
            .build()
        }),
      ],
    }),
  });
}
```

### Key Grafana 11.3+ Features

- **Scenes-Powered Dashboards**: Version control friendly (TypeScript vs JSON)
- **Edit Mode**: Simplified dashboard discovery and editing
- **Nested Subfolders**: Better organization for 1000+ dashboards
- **PDF Export**: 200-panel dashboard in 11 seconds (7x faster than v10)
- **Dynamic Variables**: Type-safe variable management
- **Reusable Components**: Build library of dashboard patterns

---

## 3. OpenTelemetry 1.33.x & Jaeger 1.62.x: Distributed Tracing

### OpenTelemetry Node.js with OTLP (v1.33)

**Native OTLP Protocol (Recommended)**:
```javascript
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-metrics-otlp-http');
const { BatchSpanProcessor } = require('@opentelemetry/sdk-trace-node');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics-node');

// OTLP HTTP exporter (standard 2025 approach)
const traceExporter = new OTLPTraceExporter({
  url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://jaeger:4318/v1/traces',
  headers: {
    // Add tenant ID for multi-tenancy
    'X-Tenant-ID': process.env.TENANT_ID || 'default',
  },
});

const metricExporter = new OTLPMetricExporter({
  url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://prometheus:4317/v1/metrics',
});

const sdk = new NodeSDK({
  traceExporter,
  metricReader: new PeriodicExportingMetricReader({
    exporter: metricExporter,
    intervalMillis: 10000,
  }),
  instrumentations: [getNodeAutoInstrumentations()],
  serviceName: 'order-service',
  serviceVersion: '1.0.0',
  resource: {
    attributes: {
      'deployment.environment': process.env.NODE_ENV,
      'service.namespace': 'payment',
      'telemetry.sdk.language': 'nodejs',
    },
  },
});

sdk.start();
```

### Jaeger v1.62 Configuration with OTLP (v1.35+)

**Docker Compose with OTLP Native Support**:
```yaml
version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one:1.62
    environment:
      # Enable OTLP protocol (native since v1.35)
      COLLECTOR_OTLP_ENABLED: 'true'
      # Storage backend
      SPAN_STORAGE_TYPE: 'elasticsearch'
      ES_SERVER_URLS: 'http://elasticsearch:9200'
      # Multi-tenancy support
      MULTI_TENANCY_ENABLED: 'true'
      # Sampling configuration
      SAMPLING_TYPE: 'remote'
      SAMPLING_PARAM: '0.1'
    ports:
      # OTLP gRPC (recommended)
      - "4317:4317"
      # OTLP HTTP
      - "4318:4318"
      # Jaeger UI
      - "16686:16686"
      # Jaeger collector (legacy, for backward compat)
      - "14268:14268"
    depends_on:
      - elasticsearch

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.17.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms1g -Xmx1g"
    ports:
      - "9200:9200"
```

### Multi-Tenancy in Jaeger v1.62

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

# Multi-tenancy via headers
multitenancy:
  enabled: true
  header_name: "X-Tenant-ID"
  tenants:
    - production
    - staging
    - development

storage:
  type: elasticsearch
  elasticsearch:
    server_urls: 'http://elasticsearch:9200'
    index_prefix: 'jaeger'
    # ILM for cost optimization
    use_ilm: true
    ilm_policy_name: 'jaeger-policy'
```

---

## 4. Loki 3.x: Bloom Filters & Log Aggregation

### Bloom Filters for Query Acceleration (70-90% faster)

**Loki 3.x Configuration with Bloom Filters**:
```yaml
# loki-config.yml (3.x)
auth_enabled: false

# Bloom filter configuration
common:
  storage:
    filesystem:
      directory: /loki/boltdb-shipper-shared
  ring:
    kvstore:
      store: inmemory

# Bloom filters (experimental → production-ready in 3.1+)
blooms_enabled: true

storage_config:
  filesystem:
    directory: /loki/chunks

# Cost optimization via retention
limits_config:
  ingestion_rate_mb: 100
  ingestion_burst_size_mb: 200
  max_streams_per_user: 10000
  max_cache_freshness_per_query: 10m
  # TTL for cost management
  retention_period: 720h  # 30 days

# OTLP receiver (native, no exporter needed)
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
```

### LogQL with Bloom Filter Optimization

```logql
# Basic query (Bloom optimized for pattern matching)
{job="api-server", level="ERROR"} | json

# Error rate by service (Bloom filters on first filter)
sum(rate({level="ERROR"}[5m])) by (service)

# Complex pattern matching (highly optimized with Bloom)
{job="payment-service", env="prod"}
  |= "ERROR"
  |= "database"
  |= "timeout"
  | json response_time=response_time_ms
  | response_time > 1000
  | stats avg(response_time), count() by (endpoint)

# Trace correlation with Bloom acceleration
{job="microservice"} 
  | json trace_id=trace_id, span_id=span_id
  | trace_id="abc123def456"
```

---

## 5. Elasticsearch 8.17: ILM for Cost Optimization

### Index Lifecycle Management (ILM) Best Practice

```bash
# Create production logs ILM policy
PUT _ilm/policy/logs-policy
{
  "policy": {
    "phases": {
      "hot": {
        "min_age": "0d",
        "actions": {
          "rollover": {
            "max_size": "50GB",
            "max_age": "1d",
            "max_primary_shard_size": "50GB"
          }
        }
      },
      "warm": {
        "min_age": "3d",
        "actions": {
          "set_priority": { "priority": 50 },
          "forcemerge": { "max_num_segments": 1 }
        }
      },
      "cold": {
        "min_age": "14d",
        "actions": {
          "set_priority": { "priority": 0 },
          "searchable_snapshot": {
            "snapshot_repository": "s3-backup"
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}

# Create index with ILM template
PUT _index_template/logs-template
{
  "index_patterns": ["logs-*"],
  "template": {
    "settings": {
      "number_of_shards": 3,
      "number_of_replicas": 1,
      "index.lifecycle.name": "logs-policy",
      "index.lifecycle.rollover_alias": "logs"
    },
    "mappings": {
      "properties": {
        "@timestamp": { "type": "date" },
        "level": { "type": "keyword" },
        "service": { "type": "keyword" },
        "trace_id": { "type": "keyword" },
        "duration_ms": { "type": "integer" }
      }
    }
  }
}
```

---

## 6. SLO/SLI with Error Budgets (2025 Best Practice)

### Error Budget Framework

```yaml
groups:
  - name: slo_definitions
    interval: 1m
    rules:
      # Define SLI: successful requests (4xx/5xx excluded)
      - record: slo:http_requests:success_rate_5m
        expr: |
          sum(rate(http_requests_total{status!~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))

      # Calculate 30-day SLI (for error budget)
      - record: slo:http_requests:success_rate_30d
        expr: |
          sum(increase(http_requests_total{status!~"5.."}[30d]))
          /
          sum(increase(http_requests_total[30d]))

      # Error budget burn rate (5m window)
      - record: slo:error_budget_burn_rate_5m
        expr: |
          (1 - slo:http_requests:success_rate_5m) 
          / 
          (1 - 0.9995)  # 99.95% SLO → 0.05% error budget

      # Alert: Error budget burn too fast
      - alert: SLOErrorBudgetBurnTooFast
        expr: |
          slo:error_budget_burn_rate_5m > 1
        for: 15m
        labels:
          severity: critical
          team: platform
        annotations:
          summary: "Error budget consuming {{ $value | humanize }}x normal rate"
          description: "SLO {{ $labels.slo }} is on track to exhaust error budget in {{ ($value | humanize) }} days"

      # Alert: Error budget exhausted
      - alert: SLOErrorBudgetExhausted
        expr: |
          slo:error_budget_burn_rate_5m > 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Error budget EXHAUSTED for {{ $labels.slo }}"
```

### Error Budget Policies

| Budget Status | Action | Duration |
|-----------|--------|----------|
| **0-25%** | Monitor, document incidents | Ongoing |
| **25-75%** | Pause non-critical features, focus on stability | Until recovery |
| **75-100%** | Freeze deployments, incident response mode | Until recovery |
| **>100%** | SLO breach, post-incident review mandatory | Post-incident |

---

## 7. Multi-Cluster Monitoring with Thanos

### Thanos Architecture (HA Prometheus Aggregation)

```yaml
# Thanos Sidecar (runs alongside each Prometheus)
sidecar:
  prometheus:
    url: http://prometheus:9090
  objstore_config:
    type: s3
    config:
      bucket: prometheus-blocks
      endpoint: s3.amazonaws.com

# Thanos Query (global query interface)
query:
  stores:
    - 'dns+sd://thanos-sidecars:9091'
    - 'dns+sd://thanos-sidecars-eu:9091'
    - 'thanos-store-gateway:10901'
  query_timeout: 5m
  dedup_enabled: true

# Thanos Compact (downsampling for cost)
compact:
  data_dir: /var/thanos/compact
  consistency_delay: 30m
  downsample:
    enabled: true
    downsample_intervals:
      - 30d:5m
      - 1y:1h
```

### Multi-Region Query Pattern

```bash
# Query across all regions with Thanos
curl 'http://thanos-query:9090/api/v1/query' \
  --data-urlencode 'query=rate(http_requests_total[5m])'
```

---

## 8. Production Best Practices (2025)

### Cardinality Management

**Pattern (Optimized)**:
```yaml
# USE: relabeling to reduce cardinality
http_request_duration_seconds{
  service="user-api",
  path="/api/users/{id}",
  method="GET",
  status="200"
}
# Result: ~100 series instead of 10M
```

### Monitoring the Monitors

```yaml
groups:
  - name: prometheus_health
    rules:
      - alert: PrometheusHighMemory
        expr: process_resident_memory_bytes / 1024 / 1024 > 2000
        annotations:
          summary: "Prometheus using {{ $value }}MB memory"

      - alert: JaegerBackendDown
        expr: up{job="jaeger"} == 0
        for: 5m

      - alert: LokiHighChunkBacklog
        expr: loki_chunk_store_index_lookups_total > 100000
        for: 10m
```

### Security & Compliance

1. TLS/mTLS for inter-component communication
2. RBAC in Jaeger, Loki, Grafana
3. Encryption at rest in Elasticsearch, S3
4. Audit logging for configuration changes
5. Network policies for traffic restriction
6. Data retention compliance (GDPR/locality)

---

## Key Enhancements in v4.0 (2025)

- **Prometheus 3.7.3**: Remote-Write 2.0, agent mode, native histograms
- **Grafana 11.3+**: Scenes GA, type-safe dashboards, 7x faster PDF export
- **OpenTelemetry 1.33.x**: Production-stable, full OTEL protocol support
- **Jaeger 1.62.x**: OTLP native (v1.35+), multi-tenancy, ES backend
- **Loki 3.x**: Bloom filters (70-90% acceleration), native OTLP
- **Elasticsearch 8.17**: ILM production-ready, searchable snapshots
- **Error Budgets**: SLO/SLI framework with burn-rate alerting
- **Multi-Cluster**: Thanos with downsampling and global queries
- **Cost Optimization**: Cardinality management, TTL policies, ILM tiers

---

## Skill Activation Pattern

This skill activates automatically when you:
- Configure Prometheus 3.x with Remote-Write 2.0 or agent mode
- Build Grafana dashboards using Scenes API
- Implement distributed tracing with OpenTelemetry 1.33+ and Jaeger
- Set up log aggregation with Loki 3.x Bloom filters
- Configure Elasticsearch 8.17 with ILM policies
- Implement SLO/SLI frameworks with error budgets
- Monitor multi-cluster infrastructure with Thanos
- Optimize observability costs and cardinality

Invoke explicitly: `Skill("moai-domain-monitoring")`
