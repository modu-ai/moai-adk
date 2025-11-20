---
name: moai-domain-monitoring
version: 4.0.0
updated: 2025-11-20
status: stable
description: Observability with Prometheus, Grafana, logging, and tracing
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Monitoring & Observability Expert

**Production Monitoring Stack**

> **Focus**: Metrics, Logs, Traces (Three Pillars of Observability)  
> **Stack**: Prometheus, Grafana, Loki, OpenTelemetry, Jaeger

---

## Overview

Complete observability for production systems.

### Three Pillars

1.  **Metrics**: Time-series data (Prometheus)
2.  **Logs**: Event records (Loki, ELK)
3.  **Traces**: Distributed request tracking (Jaeger, Tempo)

---

## Quick Start

### 1. Metrics (Prometheus)

Counter, Gauge, Histogram for application metrics.

**Key Metrics**:

- Request rate (requests/sec)
- Error rate (5xx/total)
- Latency (p50, p95, p99)

See: [examples.md](./examples.md#prometheus-metrics)

### 2. Logging (Structured)

JSON-formatted logs for easy parsing.

**Fields**: timestamp, level, message, context

See: [examples.md](./examples.md#structured-logging)

### 3. Tracing (OpenTelemetry)

Track requests across microservices.

**Concepts**: Span, Trace ID, Parent-Child relationships

See: [examples.md](./examples.md#distributed-tracing)

### 4. Alerting

Automated alerts on threshold breaches.

**Examples**: Error rate >5%, CPU >80%

See: [examples.md](./examples.md#alerting)

---

## Monitoring Methodologies

### RED Method (for Services)

- **R**ate: Requests per second
- **E**rrors: Error percentage
- **D**uration: Latency distribution

### USE Method (for Resources)

- **U**tilization: % busy
- **S**aturation: Queue depth
- **E**rrors: Error count

---

## Best Practices

1.  **Cardinality**: Avoid high-cardinality labels (e.g., user_id)
2.  **Retention**: 15 days (metrics), 7 days (logs)
3.  **Sampling**: 1% trace sampling for high-traffic services
4.  **Dashboards**: One dashboard per service

---

## Validation Checklist

- [ ] **Metrics**: Prometheus scraping configured?
- [ ] **Logs**: Structured (JSON) format?
- [ ] **Traces**: OpenTelemetry instrumented?
- [ ] **Alerts**: Critical alerts defined?
- [ ] **Dashboards**: Grafana dashboards created?

---

## Related Skills

- `moai-essentials-perf`: Performance profiling
- `moai-devops-docker`: Container monitoring
- `moai-cloud-aws-advanced`: CloudWatch

---

## Additional Resources

- [examples.md](./examples.md): Implementation code
- [reference.md](./reference.md): Prometheus query language (PromQL)

---

**Last Updated**: 2025-11-20
