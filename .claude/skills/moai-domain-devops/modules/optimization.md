# DevOps Performance Optimization

> **Version**: 4.0.0
> **Last Updated**: 2025-11-22
> **Focus**: CI/CD efficiency, pipeline optimization, infrastructure scaling, cost optimization

---

## CI/CD Pipeline Optimization

### Build Speed Optimization

```yaml
name: Optimized Build Pipeline

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest

    # Use larger runners for CPU-intensive tasks
    strategy:
      matrix:
        build-type: [quick, full]

    steps:
      - uses: actions/checkout@v4
        with:
          # Shallow clone for faster checkout
          fetch-depth: 1

      # Parallel build stages
      - name: Setup build cache
        uses: actions/cache@v3
        with:
          path: |
            ~/.npm
            .next/cache
            dist/
          key: ${{ runner.os }}-build-${{ github.sha }}
          restore-keys: |
            ${{ runner.os }}-build-

      # Quick tests (fail fast)
      - name: Quick lint and unit tests
        run: |
          npm ci --legacy-peer-deps
          npm run lint --max-warnings=0
          npm run test:unit -- --coverage --bail

      # Full tests (only on main branch)
      - name: Full test suite
        if: github.ref == 'refs/heads/main'
        run: npm run test:integration

      # Parallel build jobs
      - name: Build application
        run: npm run build

      - name: Upload build artifacts
        uses: actions/upload-artifact@v3
        with:
          name: build-artifacts-${{ matrix.build-type }}
          path: dist/
          retention-days: 1
```

### Docker Build Optimization

```dockerfile
# Multi-stage build with layer caching optimization
# syntax=docker/dockerfile:1

ARG NODE_VERSION=20-alpine3.21

# Stage 1: Dependencies (stable, cached)
FROM node:${NODE_VERSION} AS deps
WORKDIR /app

# Copy only package files (cheap)
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Stage 2: Build (may change frequently)
FROM node:${NODE_VERSION} AS builder
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build && npm run test

# Stage 3: Production (minimal)
FROM node:${NODE_VERSION}
WORKDIR /app

# Use BuildKit for better layer caching
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/package*.json ./

# Security: Non-root user
RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001
USER nodejs

EXPOSE 3000
ENV NODE_ENV=production

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s \
  CMD node -e "require('http').get('http://localhost:3000/health', (r) => {if (r.statusCode !== 200) throw new Error(r.statusCode)})"

CMD ["node", "dist/index.js"]
```

### Dependency Caching Strategy

```yaml
name: Smart Dependency Caching

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      # Cache npm
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: '**/package-lock.json'

      # Cache pip (Python)
      - name: Cache pip
        uses: actions/cache@v3
        with:
          path: ~/.cache/pip
          key: ${{ runner.os }}-pip-${{ hashFiles('**/requirements.txt') }}
          restore-keys: |
            ${{ runner.os }}-pip-

      # Docker layer caching
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        with:
          # Use local cache backend for faster builds
          driver-options: |
            image=moby/buildkit:latest

      - name: Build with cache
        uses: docker/build-push-action@v5
        with:
          context: .
          cache-from: type=local,src=/tmp/.buildx-cache
          cache-to: type=local,dest=/tmp/.buildx-cache-new
```

---

## Infrastructure Optimization

### Kubernetes Cluster Optimization

```python
class KubernetesOptimizer:
    """Optimize Kubernetes cluster resource usage"""

    @staticmethod
    async def analyze_resource_utilization(api_client):
        """Analyze and report resource optimization opportunities"""

        # Get all pods and their resource requests/limits
        pods = api_client.list_pod_for_all_namespaces()

        optimization_report = {
            "over_provisioned": [],
            "under_provisioned": [],
            "cost_savings": 0
        }

        for pod in pods.items:
            metrics = get_pod_metrics(pod)

            # Check for over-provisioned pods
            if metrics.actual_cpu < pod.resources.requests.cpu * 0.3:
                optimization_report["over_provisioned"].append({
                    "pod": pod.metadata.name,
                    "requested_cpu": pod.resources.requests.cpu,
                    "actual_cpu": metrics.actual_cpu,
                    "savings": calculate_cost_savings(pod.resources.requests.cpu, metrics.actual_cpu)
                })

        return optimization_report

    @staticmethod
    async def implement_vertical_pod_autoscaling(api_client):
        """Implement VPA for automatic resource optimization"""

        vpa_config = {
            "apiVersion": "autoscaling.k8s.io/v1",
            "kind": "VerticalPodAutoscaler",
            "metadata": {
                "name": "myapp-vpa"
            },
            "spec": {
                "targetRef": {
                    "apiVersion": "apps/v1",
                    "kind": "Deployment",
                    "name": "myapp"
                },
                "updatePolicy": {
                    "updateMode": "Auto"  # Auto-update resources
                },
                "resourcePolicy": {
                    "containerPolicies": [{
                        "containerName": "app",
                        "minAllowed": {
                            "cpu": "100m",
                            "memory": "128Mi"
                        },
                        "maxAllowed": {
                            "cpu": "2",
                            "memory": "2Gi"
                        }
                    }]
                }
            }
        }

        return vpa_config
```

### Cost Optimization

```yaml
name: Cost Optimization Report

on:
  schedule:
    - cron: '0 9 * * 1'  # Weekly Monday

jobs:
  cost-analysis:
    runs-on: ubuntu-latest

    steps:
      - name: Get EKS cluster metrics
        run: |
          # Analyze cluster costs
          aws ce get-cost-and-usage \
            --time-period Start=2025-11-15,End=2025-11-22 \
            --granularity DAILY \
            --metrics "BlendedCost" \
            --group-by Type=DIMENSION,Key=SERVICE \
            --filter file://filter.json \
            > cost-report.json

      - name: Analyze unused resources
        run: |
          # Find idle load balancers
          aws elb describe-load-balancers \
            --query 'LoadBalancerDescriptions[?length(Instances) == `0`]' \
            > unused-lb.json

          # Find unattached volumes
          aws ec2 describe-volumes \
            --filters Name=status,Values=available \
            > unused-volumes.json

      - name: Generate optimization recommendations
        run: |
          python3 scripts/analyze-costs.py \
            cost-report.json \
            unused-lb.json \
            unused-volumes.json \
            > optimization-report.md

      - name: Post recommendations
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('optimization-report.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: report
            });
```

---

## Monitoring and Observability Optimization

### Prometheus Query Optimization

```python
class PrometheusOptimization:
    """Optimize Prometheus queries and storage"""

    @staticmethod
    def optimize_long_range_queries():
        """Use recording rules to optimize expensive queries"""

        recording_rules = """
groups:
- name: optimization
  interval: 15s
  rules:
  # Pre-compute expensive aggregations
  - record: job:http_requests_total:rate5m
    expr: rate(http_requests_total[5m])

  - record: job:http_request_duration_seconds:p95
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

  - record: instance:node_memory_available_ratio
    expr: node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes

  # Alert rule optimization
  - alert: HighErrorRate
    expr: job:http_requests_total:rate5m{status=~"5.."}> 0.05
    for: 5m
    annotations:
      summary: "High error rate detected"
"""
        return recording_rules

    @staticmethod
    def implement_retention_policies():
        """Configure Prometheus retention for cost optimization"""

        prometheus_config = {
            "global": {
                "scrape_interval": "15s",
                "evaluation_interval": "15s"
            },
            "storage": {
                "tsdb": {
                    "retention": "30d",  # Keep 30 days
                    "max_block_duration": "31d",
                    "retention_size": "100GB"  # Stop at 100GB
                }
            }
        }

        return prometheus_config
```

### Log Aggregation Optimization

```python
class LogOptimization:
    """Optimize log collection and storage"""

    @staticmethod
    def configure_log_sampling():
        """Reduce log volume through intelligent sampling"""

        fluent_config = """
# Sample info logs (10%)
<filter **>
  @type sampling
  sample_percentage 10
  sample_bufsample_percentageize 100
</filter>

# Keep all error and warning logs
<filter **>
  @type detect_exceptions
  message ^stacktrace
  stream STDERR
</filter>

# Index optimization for ELK
<match **>
  @type elasticsearch_dynamic
  host elasticsearch.example.com
  logstash_format true
  logstash_prefix "logs"

  # Optimize index patterns
  <buffer tag, time>
    timekey 3600  # Hourly indices
    timekey_wait 10m
  </buffer>
</match>
"""
        return fluent_config
```

---

## Performance Benchmarking

### Load Test Optimization

```python
from k6 import http, check, sleep
from k6.thresholds import Threshold

# Performance thresholds
thresholds = {
    'http_req_duration': [
        Threshold('p(95)<500'),  # 95% under 500ms
        Threshold('p(99)<1000')  # 99% under 1000ms
    ],
    'http_req_failed': [
        Threshold('rate<0.1')  # Error rate < 10%
    ]
}

def run_optimized_load_test():
    """Run k6 load test with optimal parameters"""

    stages = [
        {"duration": "30s", "target": 10},   # Ramp up
        {"duration": "1m", "target": 50},    # Increase load
        {"duration": "2m", "target": 50},    # Sustained load
        {"duration": "30s", "target": 0}     # Ramp down
    ]

    options = {
        "stages": stages,
        "thresholds": thresholds,
        "vus": 1,
        "duration": "4m",
        "ext": {
            "loadimpact": {
                "projectID": 3356643,
                "name": "Production Load Test"
            }
        }
    }

    return options
```

---

**Best Practices**:
- Use caching aggressively at every level
- Implement fail-fast in CI/CD pipelines
- Monitor and optimize resource utilization
- Use sampling for high-volume logs
- Implement cost monitoring and alerts
- Regularly analyze and optimize queries
- Use multi-stage Docker builds effectively

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
