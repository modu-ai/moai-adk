---
name: "moai-domain-devops"
version: "4.0.0"
description: "Enterprise DevOps with Kubernetes 1.31, Docker 27.x, Terraform 1.9, GitHub Actions, monitoring with Prometheus/Grafana, and cloud-native architectures"
allowed_tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# Enterprise DevOps Architect - Production-Grade v4.0

## Technology Stack (2025 Stable)

- **Kubernetes 1.31.x** (container orchestration)
- **Docker 27.x** (container runtime)
- **GitHub Actions** (CI/CD automation)
- **Terraform 1.9.x** (infrastructure as code)
- **Prometheus 2.55.x** (monitoring & observability)
- **Grafana 11.x** (visualization dashboards)
- **ArgoCD 2.13.x** (GitOps deployments)

---

## Level 1: Quick Reference

### Core DevOps Patterns

**Multi-Stage Docker Build**:
```dockerfile
# syntax=docker/dockerfile:1
ARG NODE_VERSION=20
ARG ALPINE_VERSION=3.21

# Stage 1: Dependencies
FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION} AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Stage 2: Build
FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION} AS build
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build && npm run test

# Stage 3: Production
FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION}@sha256:1e7902618558e51428d31e6c06c2531e3170417018a45148a1f3d7305302b211
WORKDIR /app

# Security: Non-root user
RUN addgroup -g 1001 -S nodejs && adduser -S nodejs -u 1001
COPY --from=deps --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=build --chown=nodejs:nodejs /app/dist ./dist
COPY --chown=nodejs:nodejs package*.json ./

USER nodejs
EXPOSE 3000
ENV NODE_ENV=production
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s CMD node healthcheck.js
CMD ["node", "dist/index.js"]
```

**Kubernetes Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
      containers:
      - name: web-app
        image: myapp:v1.0.0@sha256:abc123...
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
```

**GitHub Actions CI/CD**:
```yaml
name: Production CI/CD

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 22]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'npm'
      - run: npm ci
      - run: npm test -- --coverage

  build:
    name: Build and Push
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**Terraform Infrastructure**:
```hcl
terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
  
  backend "s3" {
    bucket = "terraform-state-prod"
    key = "production/terraform.tfstate"
    region = "us-west-2"
    encrypt = true
    dynamodb_table = "terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy = "Terraform"
      Project = var.project_name
    }
  }
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"
  name = "${var.project_name}-vpc"
  cidr = "10.0.0.0/16"
  azs = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
  enable_nat_gateway = true
  enable_dns_hostnames = true
  enable_dns_support = true
}
```

---

## Level 2: Core Implementation

### Advanced Patterns

**Horizontal Pod Autoscaler**:
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-app-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  minReplicas: 3
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

**Prometheus Alert Rules**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: web-app-alerts
  namespace: production
spec:
  groups:
  - name: web-app.rules
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: |
        sum(rate(http_requests_total{job="web-app",status=~"5.."}[5m])) 
        / sum(rate(http_requests_total{job="web-app"}[5m])) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate detected"
    
    - alert: HighLatency
      expr: |
        histogram_quantile(0.95, 
          sum(rate(http_request_duration_seconds_bucket{job="web-app"}[5m])) by (le)
        ) > 1
      for: 10m
      labels:
        severity: warning
```

**ArgoCD GitOps**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: web-app
  namespace: argocd
spec:
  project: production
  source:
    repoURL: https://github.com/org/web-app-k8s
    targetRevision: main
    path: overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
      allowEmpty: false
    syncOptions:
    - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

---

## Level 3: Advanced Integration

### Enterprise Production Patterns

**External Secrets**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: db-credentials
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: db-credentials
    creationPolicy: Owner
    template:
      engineVersion: v2
      data:
        url: "postgresql://{{ .username }}:{{ .password }}@{{ .host }}:{{ .port }}/{{ .database }}"
  dataFrom:
  - extract:
      key: production/database
```

**Helm Chart**:
```yaml
# Chart.yaml
apiVersion: v2
name: web-app
description: Production-grade web application
version: 1.0.0
dependencies:
  - name: postgresql
    version: 12.x.x
    repository: https://charts.bitnami.com/bitnami

# values.yaml
replicaCount: 3
image:
  repository: myapp
  tag: "v1.0.0"
service:
  type: ClusterIP
  port: 80
  targetPort: 8080
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits: { cpu: 500m, memory: 512Mi }
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

**Blue-Green Deployment**:
```bash
#!/bin/bash
set -e
NAMESPACE="production"
SERVICE="web-app"
NEW_VERSION="green"

echo "Validating $NEW_VERSION deployment..."
kubectl rollout status deployment/web-app-$NEW_VERSION -n $NAMESPACE

echo "Running smoke tests..."
kubectl run smoke-test --rm -i --restart=Never \
  --image=curlimages/curl -- \
  http://web-app-$NEW_VERSION.$NAMESPACE.svc.cluster.local/health

echo "Switching traffic to $NEW_VERSION..."
kubectl patch service $SERVICE -n $NAMESPACE \
  -p "{\"spec\":{\"selector\":{\"version\":\"$NEW_VERSION\"}}}"

echo "Monitor for 10 minutes..."
sleep 600

echo "Scale down old version..."
kubectl scale deployment/web-app-$([ "$NEW_VERSION" = "blue" ] && echo "green" || echo "blue") \
  -n $NAMESPACE --replicas=0
```

---

## Level 4: Reference & Integration

### Best Practices Summary

**Container Security**:
- ✅ Use minimal base images (Alpine, distroless)
- ✅ Pin image digests for reproducibility
- ✅ Run as non-root user
- ✅ Scan images with Trivy/Snyk
- ✅ Multi-stage builds reduce attack surface

**Kubernetes Production**:
- ✅ Resource requests/limits for all containers
- ✅ Liveness and readiness probes
- ✅ HPA for automatic scaling
- ✅ PodDisruptionBudget for availability
- ✅ Network policies for security

**CI/CD Optimization**:
- ✅ Matrix builds for multi-version testing
- ✅ Caching strategies (Docker layers, npm/pip)
- ✅ Security scanning at every stage
- ✅ Parallel jobs for faster feedback
- ✅ Workflow timeouts (30 minutes recommended)

**Infrastructure as Code**:
- ✅ Remote state with locking (S3 + DynamoDB)
- ✅ State encryption with KMS
- ✅ Modular architecture for reusability
- ✅ Input validation with constraints
- ✅ Version pinning for providers

**Monitoring & Observability**:
- ✅ Scrape intervals: 15-30 seconds for most apps
- ✅ Consistent label naming, avoid high cardinality
- ✅ Alert thresholds based on SLIs/SLOs
- ✅ Template variables in dashboards
- ✅ OpenTelemetry for distributed tracing

**GitOps Workflow**:
- ✅ Git as single source of truth
- ✅ Automated sync with self-healing
- ✅ Declarative configuration (YAML in Git)
- ✅ RBAC for access control
- ✅ Audit trail via Git history

### Related Skills
- `Skill("moai-security-backend")` for security patterns
- `Skill("moai-essentials-perf")` for performance optimization
- `Skill("moai-domain-cloud")` for cloud architecture

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-13  
**Status**: Production Ready  
**Tech Stack**: Kubernetes 1.31, Docker 27.x, Terraform 1.9, Prometheus 2.55, Grafana 11.x
