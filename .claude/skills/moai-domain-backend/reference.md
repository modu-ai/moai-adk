# Backend Architecture Reference

## Architecture Decision Matrix

### When to Choose Each Pattern

| Pattern | Team Size | Complexity | Deployment | Scaling | Use Cases |
|---------|-----------|------------|------------|---------|-----------|
| **Layered Monolith** | 1-5 | Low-Medium | Simple | Vertical | MVPs, startups, internal tools |
| **Modular Monolith** | 5-15 | Medium | Simple | Vertical | Growing products, clear domains |
| **Microservices** | 15+ | High | Complex | Horizontal | Large-scale, multi-team, polyglot |
| **Serverless** | Any | Low-Medium | Managed | Auto | Event-driven, spiky traffic, rapid prototyping |
| **Event-Driven** | 10+ | High | Complex | Horizontal | Real-time analytics, decoupled systems, audit logs |

### Decision Criteria

**Choose Layered Monolith when**:
- Small team (1-5 developers)
- Rapid prototyping phase
- Simple deployment requirements
- Limited operational expertise
- Cost-sensitive project

**Choose Modular Monolith when**:
- Growing team (5-15 developers)
- Clear domain boundaries emerging
- Need independent module development
- Want extraction path to microservices
- Prefer simple deployment

**Choose Microservices when**:
- Large team (15+ developers)
- Multiple independent products/teams
- Polyglot technology requirements
- Need independent scaling per service
- High operational maturity (Kubernetes, Istio)

**Choose Serverless when**:
- Event-driven workloads
- Spiky or unpredictable traffic
- Want zero infrastructure management
- Cost optimization for low usage
- Rapid feature iteration

**Choose Event-Driven when**:
- Real-time data processing
- Audit logging requirements
- Decoupled system integration
- Complex business workflows
- Need event replay capability

## Cloud-Native Tool Comparison

### Container Orchestration

| Tool | Complexity | Ecosystem | Scaling | Multi-cloud | Production Ready |
|------|------------|-----------|---------|-------------|------------------|
| **Kubernetes 1.31.x** | High | Massive | Excellent | Yes | Yes |
| **Docker Swarm** | Low | Limited | Good | Partial | Yes |
| **AWS ECS** | Medium | AWS only | Good | No | Yes |
| **Nomad** | Medium | HashiCorp | Good | Yes | Yes |

**Recommendation**: Kubernetes 1.31.x for production, Docker Compose for local dev.

### Service Mesh

| Tool | Complexity | Features | Performance | Maturity |
|------|------------|----------|-------------|----------|
| **Istio 1.21.x** | High | Comprehensive | Good | Mature |
| **Linkerd 2.14.x** | Low | Core features | Excellent | Mature |
| **Consul Connect** | Medium | Full stack | Good | Mature |
| **AWS App Mesh** | Medium | AWS native | Good | Growing |

**Recommendation**: Istio 1.21.x for full features, Linkerd for simplicity.

### Message Brokers

| Tool | Throughput | Durability | Ordering | Complexity | Use Case |
|------|------------|------------|----------|------------|----------|
| **Apache Kafka 3.7.x** | Very High | Excellent | Partition | High | Event streaming, logs |
| **RabbitMQ 3.13.x** | High | Good | Queue | Medium | Task queues, RPC |
| **Redis Streams 7.2.x** | Very High | Good | Stream | Low | Real-time, lightweight |
| **Amazon SQS** | Medium | Excellent | Best-effort | Low | Serverless, AWS |
| **NATS 2.10.x** | Very High | Good | Stream | Low | Microservices messaging |
| **Apache Pulsar 3.2.x** | Very High | Excellent | Partition | High | Multi-tenancy, geo-replication |

**Recommendation**: Kafka 3.7.x for event streaming, RabbitMQ 3.13.x for task queues, NATS for lightweight microservices messaging.

### Observability Stack Comparison

| Stack | Traces | Metrics | Logs | Complexity | Cost |
|-------|--------|---------|------|------------|------|
| **OpenTelemetry + Prometheus + Jaeger + ELK** | Yes | Yes | Yes | High | OSS |
| **Datadog** | Yes | Yes | Yes | Low | High |
| **New Relic** | Yes | Yes | Yes | Low | High |
| **AWS X-Ray + CloudWatch** | Yes | Yes | Yes | Medium | Medium |
| **Grafana Cloud** | Yes | Yes | Yes | Low | Medium |

**Recommendation**: OpenTelemetry 1.24.0 + Prometheus 2.48.x + Jaeger 1.51.x for vendor-neutral OSS stack.

### Database Selection Matrix

| Database | Type | Scalability | Consistency | Query Flexibility | Use Case |
|----------|------|-------------|-------------|-------------------|----------|
| **PostgreSQL 16.x** | SQL | Vertical + Read Replicas | Strong ACID | SQL + JSONB | Relational, complex queries |
| **MongoDB 8.0.x** | Document | Horizontal (Sharding) | Tunable | Flexible schema | Rapid prototyping, catalogs |
| **Redis 7.2.x** | Key-Value | Horizontal (Cluster) | Eventual | Limited | Caching, sessions, pub/sub |
| **Cassandra 4.1.x** | Wide-Column | Horizontal | Tunable | CQL | Time-series, write-heavy |

**Recommendation**: PostgreSQL 16.x as default, MongoDB 8.0.x for schema flexibility.

## Deployment Strategies

### Blue-Green Deployment

**Pattern**:
- Two identical production environments (Blue = current, Green = new)
- Switch traffic after validation
- Instant rollback capability

**Pros**:
- Zero downtime
- Easy rollback
- Full validation before switch

**Cons**:
- Double infrastructure cost
- Data migration complexity
- Requires load balancer

**Best for**: Critical services with infrequent releases.

### Canary Deployment

**Pattern**:
- Gradually shift traffic to new version (5% → 25% → 50% → 100%)
- Monitor metrics at each stage
- Auto-rollback on errors

**Pros**:
- Risk mitigation
- Real-world testing
- Minimal infrastructure duplication

**Cons**:
- Requires advanced routing (Istio)
- Longer deployment time
- Version compatibility required

**Best for**: User-facing services with high traffic.

**Istio 1.21.x Implementation**:
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: backend-canary
spec:
  hosts:
  - backend.example.com
  http:
  - match:
    - headers:
        x-canary:
          exact: "true"
    route:
    - destination:
        host: backend
        subset: v2
  - route:
    - destination:
        host: backend
        subset: v1
      weight: 90
    - destination:
        host: backend
        subset: v2
      weight: 10
```

### Rolling Update

**Pattern**:
- Incrementally replace instances (pod-by-pod)
- Kubernetes default strategy
- Health checks control rollout

**Pros**:
- Built-in Kubernetes support
- No extra infrastructure
- Automatic rollback on failure

**Cons**:
- Version compatibility required
- Slower than blue-green
- Mixed versions during rollout

**Best for**: Standard microservices deployments.

**Kubernetes Configuration**:
```yaml
spec:
  replicas: 10
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2
      maxUnavailable: 1
```

### A/B Testing (Feature Flags)

**Pattern**:
- Route users to different features based on criteria
- Measure business metrics
- Gradual feature rollout

**Pros**:
- Business metric validation
- User segmentation
- Independent of deployment

**Cons**:
- Code complexity (feature flags)
- Requires analytics
- Technical debt if not cleaned

**Best for**: Product experiments, new features.

**Tools**: LaunchDarkly, Split.io, Unleash, AWS AppConfig.

## Monitoring and Observability Setup

### OpenTelemetry 1.24.0 Integration

**Step 1: Instrument Application**:
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer(__name__)

otlp_exporter = OTLPSpanExporter(endpoint="otel-collector:4317")
span_processor = BatchSpanProcessor(otlp_exporter)
trace.get_tracer_provider().add_span_processor(span_processor)
```

**Step 2: Deploy Collector**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector
spec:
  template:
    spec:
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector:0.92.0
        ports:
        - containerPort: 4317  # OTLP gRPC
        - containerPort: 4318  # OTLP HTTP
```

**Step 3: Configure Backends**:
```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:

processors:
  batch:

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
  jaeger:
    endpoint: "jaeger:14250"
  elasticsearch:
    endpoints: ["http://elasticsearch:9200"]

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [jaeger]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [elasticsearch]
```

### Prometheus 2.48.x Setup

**Step 1: Install Prometheus Operator**:
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi
```

**Step 2: ServiceMonitor for Application**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: backend-metrics
spec:
  selector:
    matchLabels:
      app: backend
  endpoints:
  - port: metrics
    interval: 15s
```

**Step 3: Alerting Rules**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: backend-alerts
spec:
  groups:
  - name: backend
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate detected"
```

### Jaeger 1.51.x Deployment

**All-in-One (Development)**:
```bash
kubectl create deployment jaeger --image=jaegertracing/all-in-one:1.51
kubectl expose deployment jaeger --port=16686 --target-port=16686 --type=LoadBalancer
```

**Production (with Elasticsearch)**:
```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm install jaeger jaegertracing/jaeger \
  --set provisionDataStore.cassandra=false \
  --set provisionDataStore.elasticsearch=true \
  --set storage.type=elasticsearch \
  --set storage.elasticsearch.host=elasticsearch-master \
  --set storage.elasticsearch.port=9200
```

### Grafana 10.x Dashboards

**Key Dashboards**:
- **RED Metrics**: Rate, Errors, Duration per service
- **Kubernetes Cluster**: Node/pod resource usage
- **Database Performance**: Query latency, connection pool
- **Application Logs**: Error rate trends, log volume

**Import Community Dashboards**:
- Kubernetes Cluster Monitoring (ID: 7249)
- Node Exporter Full (ID: 1860)
- PostgreSQL Database (ID: 9628)

## Performance Benchmarking

### Load Testing Tools

| Tool | Protocol | Complexity | Reporting | Use Case |
|------|----------|------------|-----------|----------|
| **k6** | HTTP, gRPC, WebSocket | Low | Good | API load testing |
| **Gatling** | HTTP, WebSocket | Medium | Excellent | Complex scenarios |
| **JMeter** | HTTP, JDBC, JMS, etc. | High | Good | Enterprise testing |
| **Locust** | HTTP, custom | Low | Good | Python-based scripting |

**Recommendation**: k6 for API load testing, Gatling for complex user journeys.

### k6 Example (HTTP API)

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp-up
    { duration: '5m', target: 100 },  // Steady
    { duration: '2m', target: 0 },    // Ramp-down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% under 500ms
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  let res = http.get('https://api.example.com/users');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

**Run**:
```bash
k6 run --vus 100 --duration 10m load-test.js
```

### Performance Targets

| Metric | Target | Good | Excellent |
|--------|--------|------|-----------|
| **API Response Time (p95)** | <500ms | <200ms | <100ms |
| **Error Rate** | <1% | <0.1% | <0.01% |
| **Throughput** | 1000 req/s | 5000 req/s | 10000 req/s |
| **Database Query (p95)** | <100ms | <50ms | <10ms |
| **CPU Usage** | <70% | <50% | <30% |
| **Memory Usage** | <80% | <60% | <40% |

## Security Best Practices

### OWASP API Security Top 10 2023 Checklist

- [ ] **API1: Broken Object Level Authorization**
  - Implement per-object access control checks
  - Validate user owns the resource before access
  - Use UUIDs instead of sequential IDs

- [ ] **API2: Broken Authentication**
  - Use industry-standard auth (OAuth 2.0, OpenID Connect)
  - Implement rate limiting on auth endpoints
  - Rotate secrets regularly

- [ ] **API3: Broken Object Property Level Authorization**
  - Use response DTOs to filter sensitive fields
  - Validate input against allowed fields (no mass assignment)
  - Implement role-based field visibility

- [ ] **API4: Unrestricted Resource Consumption**
  - Enforce pagination (max 100 items per page)
  - Implement rate limiting (Redis-based)
  - Set request timeouts and size limits

- [ ] **API5: Broken Function Level Authorization**
  - Validate roles/permissions on every endpoint
  - Deny by default, allow explicitly
  - Use middleware for consistent checks

- [ ] **API6: Unrestricted Access to Sensitive Business Flows**
  - Implement CAPTCHA for critical flows
  - Detect and block automated abuse
  - Monitor for unusual patterns

- [ ] **API7: Server Side Request Forgery (SSRF)**
  - Validate and whitelist URLs
  - Block private IP ranges
  - Use network segmentation

- [ ] **API8: Security Misconfiguration**
  - Disable debug mode in production
  - Remove unnecessary endpoints
  - Use security headers (HSTS, CSP, etc.)

- [ ] **API9: Improper Inventory Management**
  - Document all API versions
  - Deprecate old versions gracefully
  - Monitor for unauthorized endpoints

- [ ] **API10: Unsafe Consumption of APIs**
  - Validate external API responses
  - Set timeouts for external calls
  - Handle partial failures gracefully

### Rate Limiting Implementation (Redis 7.2.x)

**Token Bucket Algorithm**:
```python
import redis
import time

def is_rate_limited(user_id: str, max_requests: int = 100, window: int = 60) -> bool:
    """Token bucket rate limiting with Redis."""
    r = redis.Redis(host='localhost', port=6379, db=0)
    key = f"rate_limit:{user_id}"

    current = r.get(key)
    if current is None:
        r.setex(key, window, 1)
        return False

    current = int(current)
    if current >= max_requests:
        return True

    r.incr(key)
    return False
```

**Sliding Window (Accurate)**:
```python
def sliding_window_rate_limit(user_id: str, max_requests: int = 100, window: int = 60) -> bool:
    """Sliding window rate limiting with Redis sorted sets."""
    r = redis.Redis(host='localhost', port=6379, db=0)
    key = f"rate_limit:{user_id}"
    now = time.time()

    # Remove old entries
    r.zremrangebyscore(key, 0, now - window)

    # Count requests in window
    count = r.zcard(key)
    if count >= max_requests:
        return True

    # Add current request
    r.zadd(key, {str(now): now})
    r.expire(key, window)
    return False
```

## Cost Optimization

### Resource Right-Sizing

**Kubernetes Resource Requests/Limits**:
```yaml
resources:
  requests:
    cpu: "100m"      # Minimum guaranteed
    memory: "128Mi"
  limits:
    cpu: "500m"      # Maximum allowed
    memory: "512Mi"
```

**Guidelines**:
- Set requests based on p50 usage
- Set limits based on p99 usage
- Enable Vertical Pod Autoscaler for recommendations

### Auto-Scaling Strategies

**Horizontal Pod Autoscaler (HPA)**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Cluster Autoscaler**:
- Automatically adds/removes nodes based on pending pods
- Works with cloud provider APIs (AWS, GCP, Azure)
- Set min/max node count per node group

### Caching Strategies for Cost Reduction

**CDN for Static Assets**:
- CloudFront, CloudFlare: ~$0.085/GB
- Reduces origin server load by 80-90%
- Set long TTLs (1 year) with versioned URLs

**Redis Caching**:
- Cache database queries (TTL: 5-60 minutes)
- Session storage (reduce database connections)
- Precomputed results (expensive calculations)

**Database Connection Pooling**:
- PgBouncer: Reduce database connections by 10x
- Typical pool size: 10-20 connections per app instance
- Significant cost savings on managed databases

### Container Registry Comparison

| Registry | Security Scanning | Geo-Replication | Cost | Integration | Use Case |
|----------|------------------|------------------|------|-------------|----------|
| **Docker Hub** | Free tier limited | No | Free/Paid | Universal | Public images |
| **Amazon ECR** | Yes (Inspector) | Yes | Pay-per-GB | AWS native | AWS workloads |
| **Google Artifact Registry** | Yes (Container Analysis) | Yes | Pay-per-GB | GCP native | GCP workloads |
| **Azure ACR** | Yes (Defender) | Yes | Pay-per-GB | Azure native | Azure workloads |
| **Harbor 2.10.x** | Yes (Trivy) | Yes | Self-hosted | Kubernetes | On-premises, multi-cloud |
| **GitHub Container Registry** | Yes (Dependabot) | Global CDN | Free for public | GitHub Actions | CI/CD integration |

**Recommendation**: Harbor 2.10.x for self-hosted multi-cloud, cloud-native registries for vendor lock-in scenarios.

### API Gateway Comparison

| Gateway | Protocol | Performance | Features | Complexity | Use Case |
|---------|----------|-------------|----------|------------|----------|
| **Kong 3.6.x** | HTTP, gRPC, WebSocket | Excellent | Plugins, mTLS, rate limiting | Medium | Enterprise APIs |
| **Ambassador (Emissary-Ingress) 3.9.x** | HTTP, gRPC | Good | Kubernetes-native, Istio integration | Medium | Kubernetes microservices |
| **Traefik 3.0.x** | HTTP, TCP, gRPC | Good | Auto-discovery, Let's Encrypt | Low | Docker/Kubernetes |
| **NGINX Ingress Controller** | HTTP, TCP, UDP | Excellent | Battle-tested, low overhead | Low | General-purpose |
| **AWS API Gateway** | HTTP, WebSocket | Good | Serverless, AWS Lambda | Low | AWS serverless |
| **Envoy Proxy 1.29.x** | HTTP/1.1, HTTP/2, gRPC | Excellent | L7 proxy, xDS API | High | Service mesh (Istio) |

**Recommendation**: Kong 3.6.x for feature-rich APIs, NGINX Ingress for simplicity, Envoy 1.29.x for service mesh integration.

### CI/CD Platform Comparison

| Platform | Cloud/Self-hosted | Kubernetes Support | Pipeline as Code | Cost | Use Case |
|----------|-------------------|-------------------|------------------|------|----------|
| **GitHub Actions** | Cloud | Yes (Kubernetes runners) | YAML | Free tier + paid | GitHub integration |
| **GitLab CI/CD** | Both | Yes | YAML | Free tier + paid | GitLab ecosystem |
| **Jenkins 2.440.x** | Self-hosted | Yes (Kubernetes plugin) | Groovy/Declarative | Free (self-hosted) | Enterprise, customization |
| **ArgoCD 2.10.x** | Self-hosted | Yes (GitOps) | YAML (Git) | Free | Kubernetes GitOps |
| **Tekton Pipelines 0.56.x** | Self-hosted | Yes (cloud-native) | YAML | Free | Kubernetes-native CI/CD |
| **CircleCI** | Cloud | Yes | YAML | Free tier + paid | Fast feedback loops |

**Recommendation**: ArgoCD 2.10.x for Kubernetes GitOps, GitHub Actions for GitHub-native workflows, Jenkins for complex enterprise pipelines.

## References

- Kubernetes. "Kubernetes Best Practices." https://kubernetes.io/docs/concepts/configuration/overview/ (2025-10-22)
- Istio. "Performance and Scalability." https://istio.io/latest/docs/ops/deployment/performance-and-scalability/ (2025-10-22)
- OpenTelemetry. "Getting Started." https://opentelemetry.io/docs/instrumentation/ (2025-10-22)
- Prometheus. "Best Practices." https://prometheus.io/docs/practices/ (2025-10-22)
- OWASP. "API Security Project." https://owasp.org/www-project-api-security/ (2025-10-22)
- CNCF. "Cloud Native Landscape." https://landscape.cncf.io/ (2025-10-22)
- Kafka. "Apache Kafka Documentation." https://kafka.apache.org/documentation/ (2025-10-22)
- Harbor. "Harbor Documentation." https://goharbor.io/docs/ (2025-10-22)
- ArgoCD. "ArgoCD Documentation." https://argo-cd.readthedocs.io/ (2025-10-22)
