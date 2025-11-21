---
name: moai-domain-monitoring
description: Production Monitoring & Observability with Prometheus, Grafana, OpenTelemetry, and Alerting
---

## Quick Reference (30 seconds)

# Production Monitoring & Observability

**Primary Focus**: Metrics, logs, traces, alerting, observability
**Best For**: System monitoring, performance tracking, incident response, debugging production
**Key Tools**: Prometheus 3.0, Grafana 11, OpenTelemetry 1.28, ELK Stack
**Auto-triggers**: Monitoring setup, alerting, observability, performance analysis

| Tool | Version | Features |
|------|---------|----------|
| Prometheus | 3.0 | Metrics collection |
| Grafana | 11 | Visualization |
| OpenTelemetry | 1.28 | Distributed tracing |
| AlertManager | 0.27 | Alert routing |

---

## What It Does

Production monitoring and observability with metrics, logs, and traces. Real-time alerting, dashboards, performance tracking, and incident response.

**Key capabilities**:
- ✅ Metrics collection with Prometheus
- ✅ Visualization with Grafana dashboards
- ✅ Distributed tracing with OpenTelemetry
- ✅ Log aggregation and analysis
- ✅ Alerting and incident management
- ✅ Performance profiling
- ✅ Health checks and SLO monitoring

---

## When to Use

**Automatic triggers**:
- Production deployments
- Performance optimization
- Incident investigation
- System health monitoring
- Alerting and SLO setup

**Manual invocation**:
- Design monitoring strategy
- Create dashboards
- Configure alerts
- Incident post-mortems
- Performance analysis

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core monitoring concepts:
- **Metrics**: Counters, gauges, histograms, summaries
- **Dashboards**: Key metrics, visualization
- **Alerting**: Thresholds, notification channels
- **Logs**: Structured logging, centralization
- **Health Checks**: Readiness and liveness probes

### Level 2: Advanced Patterns (See modules/tracing-strategy.md)

Production-ready observability:
- **Distributed Tracing**: OpenTelemetry, spans, context
- **Correlation IDs**: Request tracing across services
- **Sampling Strategies**: Adaptive sampling, decision-based
- **Error Tracking**: Exception handling, error context
- **Performance Profiling**: CPU, memory, allocations

### Level 3: Operations & SLOs (See modules/slo-management.md)

Production operations:
- **SLO Definition**: Availability, latency, error rate
- **Alert Tuning**: Reducing false positives
- **Incident Management**: On-call, escalation
- **Capacity Planning**: Growth forecasting
- **Cost Optimization**: Resource monitoring

---

## Best Practices

✅ **DO**:
- Instrument business metrics
- Use structured logging
- Implement distributed tracing
- Define clear SLOs
- Set up meaningful alerts
- Document runbooks
- Review metrics regularly
- Test alert escalation

❌ **DON'T**:
- Alert on everything
- Ignore cardinality explosion
- Skip log retention policies
- Forget correlation IDs
- Ignore tail latency
- Create unmaintainable alerts
- Skip post-mortems
- Ignore cost of monitoring

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **Prometheus** | 3.0 | Metrics |
| **Grafana** | 11 | Dashboards |
| **OpenTelemetry** | 1.28 | Tracing |
| **ELK Stack** | 8.x | Logging |
| **AlertManager** | 0.27 | Alerting |

---

## Works Well With

- `moai-domain-backend` (Backend instrumentation)
- `moai-domain-devops` (Infrastructure monitoring)
- `moai-essentials-perf` (Performance analysis)

---

## Learn More

- **Examples**: See `examples.md` for real-world patterns
- **Tracing**: See `modules/tracing-strategy.md` for distributed tracing
- **SLOs**: See `modules/slo-management.md` for SLO setup
- **Prometheus Docs**: https://prometheus.io/docs/
- **Grafana Docs**: https://grafana.com/docs/
- **OpenTelemetry**: https://opentelemetry.io/

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized with tracing and SLO modules
- **v3.0.0** (2025-11-13): OpenTelemetry 1.28, Prometheus 3.0
- **v2.0.0** (2025-10-01): Advanced alerting, SLO management
- **v1.0.0** (2025-03-01): Initial release

---

**Skills**: Skill("moai-domain-backend"), Skill("moai-domain-devops"), Skill("moai-essentials-perf")
**Auto-loads**: Monitoring and observability projects

