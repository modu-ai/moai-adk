# SLO Management & Incident Response

_Last updated: 2025-11-22_

## SLO Definition

### Common SLO Types

#### Availability SLO
```
99.9% availability = maximum 43 minutes downtime/month

Metric: successful_requests / total_requests >= 99.9%
Alert: availability < 99.5% (warning) or < 99% (critical)
```

#### Latency SLO
```
p95 latency < 500ms for API endpoints

Metric: histogram_quantile(0.95, latency) < 500ms
Alert: p95 > 600ms (warning) or > 1000ms (critical)
```

#### Error Rate SLO
```
Error rate < 0.1% across all services

Metric: failed_requests / total_requests < 0.001
Alert: error_rate > 0.05% (warning) or > 0.1% (critical)
```

## Alert Configuration

### Prometheus Alert Rules

```yaml
groups:
  - name: api_slos
    rules:
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total[5m]))
          ) > 0.001
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

      - alert: SlowLatency
        expr: |
          histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
          > 0.5
        for: 10m
        labels:
          severity: warning
```

## Incident Management

### On-Call Rotation
- Primary and secondary on-call
- Escalation after 15 minutes
- Page only for critical issues
- Runbook for each alert

### Incident Response Template
1. Acknowledge alert (2 min)
2. Investigate root cause (5 min)
3. Identify impact (2 min)
4. Communicate status (1 min)
5. Mitigate issue (varies)
6. Post-mortem within 24 hours

### Postmortem Checklist
- [ ] Root cause identified
- [ ] Timeline documented
- [ ] Impact quantified
- [ ] Remediation items listed
- [ ] Owner assigned to each item
- [ ] Follow-up date set

---

**Last Updated**: 2025-11-22

