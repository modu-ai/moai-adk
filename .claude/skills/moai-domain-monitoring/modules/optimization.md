# Monitoring Optimization and SLO Management

> **Version**: 4.0.0
> **Last Updated**: 2025-11-22
> **Focus**: SLI/SLO definition, alert optimization, performance baselines, monitoring strategy

---

## SLI and SLO Definition

### Service Level Indicators and Objectives

```python
from dataclasses import dataclass
from enum import Enum

class SLOTarget(Enum):
    """Standard SLO targets"""
    # Availability targets
    AVAILABILITY_99_9 = 0.999   # 43.2 minutes downtime/month
    AVAILABILITY_99_99 = 0.9999  # 4.32 minutes downtime/month
    AVAILABILITY_99_999 = 0.99999  # 25.9 seconds downtime/month

    # Performance targets
    LATENCY_P95_100MS = 0.95
    LATENCY_P99_500MS = 0.99

@dataclass
class ServiceLevelIndicator:
    """Define SLI for a service"""

    name: str
    description: str
    measurement_type: str  # availability, latency, error_rate
    threshold: float
    window: int  # seconds

    # Prometheus query to measure SLI
    measurement_query: str

class SLOManager:
    """Manage SLOs and track error budgets"""

    @staticmethod
    def define_service_slo(service_name: str) -> dict:
        """Define comprehensive SLO for a service"""

        slos = {
            "availability": ServiceLevelIndicator(
                name="service_availability",
                description="Percentage of successful requests",
                measurement_type="availability",
                threshold=SLOTarget.AVAILABILITY_99_9.value,
                window=2592000,  # 30 days
                measurement_query="""
                    sum(rate(http_requests_total{service="myservice", status=~"2.."}[5m]))
                    /
                    sum(rate(http_requests_total{service="myservice"}[5m]))
                """
            ),

            "latency_p95": ServiceLevelIndicator(
                name="request_latency_p95",
                description="95th percentile request latency",
                measurement_type="latency",
                threshold=0.95,  # 95% of requests under 100ms
                window=2592000,
                measurement_query="""
                    histogram_quantile(0.95,
                        rate(http_request_duration_seconds_bucket{service="myservice"}[5m])
                    ) < 0.1
                """
            ),

            "error_rate": ServiceLevelIndicator(
                name="error_rate",
                description="Error rate below 0.5%",
                measurement_type="error_rate",
                threshold=0.995,  # Less than 0.5% errors
                window=2592000,
                measurement_query="""
                    (
                        sum(rate(http_requests_total{service="myservice", status=~"2.."}[5m]))
                        /
                        sum(rate(http_requests_total{service="myservice"}[5m]))
                    ) > 0.995
                """
            )
        }

        return slos

    @staticmethod
    def calculate_error_budget(slo_target: float, period_seconds: int) -> dict:
        """Calculate remaining error budget"""

        # Monthly budget (30 days)
        monthly_seconds = 30 * 24 * 60 * 60

        # Allowed downtime for SLO
        allowed_downtime = (1 - slo_target) * monthly_seconds

        return {
            "slo_target": slo_target,
            "allowed_downtime_seconds": allowed_downtime,
            "allowed_downtime_minutes": allowed_downtime / 60,
            "allowed_downtime_hours": allowed_downtime / 3600,
            "percentage": (1 - slo_target) * 100
        }

    @staticmethod
    async def track_error_budget(prometheus_url: str, service_name: str):
        """Track error budget consumption"""

        slos = SLOManager.define_service_slo(service_name)

        # Query current metrics from Prometheus
        async with httpx.AsyncClient() as client:
            results = {}

            for slo_name, sli in slos.items():
                response = await client.get(
                    f"{prometheus_url}/api/v1/query",
                    params={"query": sli.measurement_query}
                )

                data = response.json()
                if data["status"] == "success":
                    value = float(data["data"]["result"][0]["value"][1])

                    # Calculate error budget used
                    budget = SLOManager.calculate_error_budget(
                        sli.threshold,
                        sli.window
                    )

                    error_budget_used = (1 - value) * sli.window

                    results[slo_name] = {
                        "current_value": value,
                        "target": sli.threshold,
                        "error_budget": budget,
                        "budget_used_percentage": (error_budget_used / (budget["allowed_downtime_seconds"])) * 100
                    }

            return results
```

### Error Budget Alerts

```yaml
# Prometheus alert rules for error budget monitoring
groups:
- name: error_budget
  interval: 5m
  rules:

  # Alert when error budget consumption is high
  - alert: HighErrorBudgetConsumption
    expr: |
      (
        1 - (sum(rate(http_requests_total{status=~"2.."}[30d])) / sum(rate(http_requests_total[30d])))
      ) > 0.005  # Using more than 0.5% of budget
    for: 10m
    annotations:
      summary: "{{ $labels.service }} using {{ $value | humanizePercentage }} error budget"
      description: "Error budget consumption is higher than expected"

  # Alert when error budget will be exhausted in 1 week
  - alert: ErrorBudgetExhaustionWarning
    expr: |
      (
        1 - (sum(rate(http_requests_total{status=~"2.."}[7d])) / sum(rate(http_requests_total[7d])))
      ) > 0.003  # Using more than 0.3% in a week
    for: 5m
    annotations:
      summary: "{{ $labels.service }} error budget will be exhausted within 1 week"
      description: "Current burn rate will exhaust monthly error budget in 7 days"

  # Critical alert when error budget is exhausted
  - alert: ErrorBudgetExhausted
    expr: |
      (
        1 - (sum(rate(http_requests_total{status=~"2.."}[30d])) / sum(rate(http_requests_total[30d])))
      ) >= 0.01  # Using entire error budget (1%)
    for: 1m
    annotations:
      summary: "{{ $labels.service }} error budget EXHAUSTED"
      description: "Monthly error budget has been exhausted"
```

---

## Alert Optimization

### Reducing Alert Fatigue

```python
class AlertOptimizer:
    """Optimize alerts to reduce false positives"""

    @staticmethod
    def define_alert_thresholds(metric_baseline: dict) -> dict:
        """Define intelligent alert thresholds based on baselines"""

        # Calculate statistical thresholds
        mean = metric_baseline["mean"]
        std_dev = metric_baseline["std_dev"]
        p95 = metric_baseline["p95"]
        p99 = metric_baseline["p99"]

        return {
            # Warning: 2 sigma above normal
            "warning": mean + (2 * std_dev),

            # Critical: 3 sigma above normal
            "critical": mean + (3 * std_dev),

            # Anomaly: Significantly different from normal
            "anomaly": p99 + (0.5 * std_dev),

            # Surge: Above 95th percentile
            "surge": p95 * 1.5
        }

    @staticmethod
    def configure_alert_suppression() -> str:
        """Configure alert suppression rules"""

        return """
# AlertManager suppression rules
inhibit_rules:

# Suppress warning if critical alert exists
- source_match:
    severity: critical
  target_match:
    severity: warning
  equal: ['alertname', 'dev', 'instance']

# Don't alert during maintenance window
- target_match_re:
    status: ^firing$
  match:
    maintenance_window: 'true'

# Don't alert for expected transient issues
- source_match:
    severity: critical
    alert_type: transient
  target_match:
    severity: warning
"""

    @staticmethod
    def implement_alert_correlation() -> str:
        """Correlate related alerts to reduce noise"""

        return """
# Prometheus recording rules for alert correlation
groups:
- name: alert_correlation
  interval: 1m
  rules:

  # Only alert if multiple metrics indicate problem
  - alert: ServiceDegraded
    expr: |
      (
        rate(http_requests_total{status=~"5.."}[5m]) > 0.05
      )
      and
      (
        histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1.0
      )
    annotations:
      summary: "Service degraded - high error rate AND high latency"

  # Only alert if both CPU and memory are high
  - alert: NodeResourcePressure
    expr: |
      (
        node_cpu_seconds_total > 0.8
      )
      and
      (
        node_memory_usage_bytes / node_memory_limit_bytes > 0.8
      )
    annotations:
      summary: "Node has both CPU and memory pressure"
"""
```

---

## Monitoring Strategy Optimization

### Metric Cardinality Management

```python
class CardinalityManagement:
    """Manage metric cardinality to optimize storage"""

    @staticmethod
    def identify_high_cardinality_metrics(prometheus_url: str) -> list:
        """Find metrics with high cardinality"""

        high_cardinality = []

        # Query Prometheus for cardinality info
        # /api/v1/metadata endpoint
        # /api/v1/cardinality endpoint (Prometheus 2.37+)

        metrics = [
            {
                "name": "http_request_duration_seconds",
                "cardinality": 150000,
                "high_cardinality_labels": ["path", "method"]
            },
            {
                "name": "user_actions",
                "cardinality": 500000,
                "high_cardinality_labels": ["user_id", "action_type"]
            }
        ]

        return metrics

    @staticmethod
    def implement_relabeling_rules() -> str:
        """Reduce cardinality through relabeling"""

        return """
# prometheus.yml - metric_relabel_configs
metric_relabel_configs:

# Drop high-cardinality path labels
- source_labels: [__name__, http_path]
  regex: 'http_request_duration_seconds_bucket;/api/users/[0-9]+'
  action: replace
  target_label: http_path
  replacement: '/api/users/{id}'

# Drop query string parameters from paths
- source_labels: [http_path]
  regex: '(.+)\\?.+'
  action: replace
  target_label: http_path
  replacement: '${1}'

# Drop port information for database connections
- source_labels: [instance]
  regex: '([^:]+):[0-9]+'
  action: replace
  target_label: instance
  replacement: '${1}'

# Limit label value length
- source_labels: [user_id]
  regex: '(.{50}).*'
  action: replace
  target_label: user_id
  replacement: '${1}...'
"""

    @staticmethod
    def implement_downsampling_rules() -> str:
        """Downsample metrics for long-term storage"""

        return """
# Recording rules for downsampling
groups:
- name: downsampling
  interval: 1m
  rules:

  # 5-minute average for 30+ day retention
  - record: http_requests_5m:sum_rate
    expr: sum by (service, method) (rate(http_requests_total[5m]))

  # 15-minute average for 1+ year retention
  - record: http_requests_15m:sum_rate
    expr: sum by (service, method) (rate(http_requests_total[15m]))

  # 1-hour average for 2+ year retention
  - record: http_requests_1h:sum_rate
    expr: sum by (service, method) (rate(http_requests_total[1h]))
"""
```

---

## Performance Baseline Analysis

### Establishing Baselines

```python
class PerformanceBaselines:
    """Establish and track performance baselines"""

    @staticmethod
    def calculate_baseline_metrics(historical_data: list) -> dict:
        """Calculate baseline from historical data"""

        import statistics

        latencies = [x["latency_ms"] for x in historical_data]
        errors = [x["error_count"] for x in historical_data]

        return {
            "latency": {
                "p50": statistics.median(latencies),
                "p95": statistics.quantiles(latencies, n=20)[18],  # 95th percentile
                "p99": statistics.quantiles(latencies, n=100)[98],  # 99th percentile
                "mean": statistics.mean(latencies),
                "stdev": statistics.stdev(latencies)
            },
            "errors": {
                "mean": statistics.mean(errors),
                "max": max(errors),
                "rate": sum(errors) / len(errors)
            }
        }

    @staticmethod
    def detect_anomalies(current_metric: float, baseline: dict, std_devs: float = 3) -> bool:
        """Detect anomalies using statistical analysis"""

        mean = baseline["mean"]
        stdev = baseline["stdev"]

        # Anomaly if > N standard deviations from mean
        threshold = mean + (std_devs * stdev)

        return current_metric > threshold

    @staticmethod
    def forecast_growth(historical_data: list, days_ahead: int = 30) -> dict:
        """Forecast metric growth"""

        import numpy as np
        from scipy.stats import linregress

        # Extract time series data
        x = np.array(range(len(historical_data)))
        y = np.array([d["value"] for d in historical_data])

        # Linear regression
        slope, intercept, r_value, p_value, std_err = linregress(x, y)

        # Forecast
        future_x = len(historical_data) + days_ahead
        forecasted_value = slope * future_x + intercept

        return {
            "current_value": y[-1],
            "forecasted_value": forecasted_value,
            "growth_rate": slope,
            "r_squared": r_value ** 2,
            "will_hit_threshold": forecasted_value > 0.8  # Example threshold
        }
```

---

## Best Practices

- Define clear SLIs and SLOs for each service
- Implement error budget tracking
- Reduce alert fatigue through intelligent tuning
- Monitor metric cardinality and optimize storage
- Establish performance baselines regularly
- Use statistical methods for anomaly detection
- Implement alert correlation to reduce noise
- Document all alert runbooks
- Review alerts monthly for optimization
- Track alert effectiveness and adjust thresholds

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
