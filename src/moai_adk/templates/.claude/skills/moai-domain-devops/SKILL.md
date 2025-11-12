---
name: "moai-domain-devops"
version: "4.0.0"
description: Enterprise DevOps expertise with production-grade patterns for Kubernetes 1.31, Docker 27.x, Terraform 1.9, GitHub Actions, and cloud-native architectures; activates for CI/CD pipelines, infrastructure automation, container orchestration, monitoring with Prometheus/Grafana, and DevOps transformation strategies.
allowed-tools: 
  - Read
  - Bash
  - WebSearch
  - WebFetch
status: stable
---

# 🚀 Enterprise DevOps Architect — Production-Grade v4.0

## 🎯 Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **4 (Enterprise)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | DevOps architecture, CI/CD, container orchestration |
| **Trigger cues** | DevOps, Kubernetes, Docker, Terraform, CI/CD, GitOps, infrastructure automation, monitoring, observability |

## 🌟 Technology Stack (2025 Stable Versions)

### Container & Orchestration
- **Kubernetes 1.31.x** (stable, released 2024-08)
- **Docker 27.x** (stable, released 2024)
- **Helm 3.16.x** (stable package management)
- **containerd 1.7.x** (stable container runtime)

### CI/CD Platforms
- **GitHub Actions** (2025 enterprise best practices)
- **GitLab CI/CD 17.x** (stable)
- **Jenkins 2.479.x LTS**
- **ArgoCD 2.13.x** (GitOps stable)

### Infrastructure as Code
- **Terraform 1.9.x** (stable)
- **Pulumi 3.x** (stable)
- **Ansible 2.17.x** (stable automation)

### Monitoring & Observability
- **Prometheus 2.55.x** (stable, pre-3.0 upgrade path)
- **Grafana 11.x** (stable, Scenes-powered dashboards)
- **Loki 3.x** (log aggregation)
- **Jaeger 1.62.x** (distributed tracing)
- **OpenTelemetry 1.33.x** (observability standard)

## 📦 Pattern 1: Multi-Stage Docker Build (Production-Grade)

### Optimized Build with Security Scanning

**Use Case**: Build slim, secure container images with minimal attack surface.

**Dockerfile (Node.js Example)**:
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
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build && npm run test

# Stage 3: Production
FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION}@sha256:1e7902618558e51428d31e6c06c2531e3170417018a45148a1f3d7305302b211
WORKDIR /app

# Security: Non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Copy production deps and build artifacts
COPY --from=deps --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=build --chown=nodejs:nodejs /app/dist ./dist
COPY --chown=nodejs:nodejs package*.json ./

USER nodejs
EXPOSE 3000
ENV NODE_ENV=production

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s \
  CMD node healthcheck.js

CMD ["node", "dist/index.js"]
```

**Key Features**:
- Multi-stage build reduces final image size by 70%
- Pinned base image digest for supply chain integrity
- Non-root user execution (security hardening)
- Health checks for container orchestration
- Layer caching optimization

**Build Command**:
```bash
docker build --no-cache -t myapp:v1.0.0 .
docker scan myapp:v1.0.0  # Security vulnerability scan
```

---

## 🏗️ Pattern 2: Kubernetes Deployment (Production-Ready)

### Complete Deployment with HPA, Probes, and Resource Limits

**Use Case**: Deploy scalable, resilient applications on Kubernetes.

**deployment.yaml**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: production
  labels:
    app: web-app
    version: v1.0.0
    managed-by: argocd
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0  # Zero-downtime deployment
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: web-app-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1001
        fsGroup: 1001
      containers:
      - name: web-app
        image: myapp:v1.0.0@sha256:abc123...
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 8080
          protocol: TCP
        env:
        - name: NODE_ENV
          value: production
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: url
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
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          successThreshold: 1
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: web-app-config
---
apiVersion: v1
kind: Service
metadata:
  name: web-app
  namespace: production
spec:
  type: ClusterIP
  selector:
    app: web-app
  ports:
  - name: http
    port: 80
    targetPort: http
    protocol: TCP
```

**Horizontal Pod Autoscaler (HPA)**:
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
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
```

**Key Features**:
- Rolling update strategy with zero downtime
- Resource requests/limits for QoS
- Liveness and readiness probes
- HPA with CPU/memory metrics
- Security context (non-root user)
- Prometheus scraping annotations

---

## 📊 Pattern 3: GitHub Actions CI/CD (Enterprise Grade)

### Complete Pipeline with Matrix Builds, Caching, and Security

**Use Case**: Automated testing, building, and deployment pipeline.

**.github/workflows/ci-cd.yml**:
```yaml
name: Production CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  workflow_dispatch:

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    name: Test (Node ${{ matrix.node-version }})
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [18, 20, 22]
      fail-fast: false
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run tests with coverage
        run: npm test -- --coverage

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          token: ${{ secrets.CODECOV_TOKEN }}
          files: ./coverage/coverage-final.json
          flags: unittests
          name: node-${{ matrix.node-version }}

  security:
    name: Security Scanning
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'

  build:
    name: Build and Push Docker Image
    needs: [test, security]
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            NODE_VERSION=20
            BUILD_DATE=${{ github.event.head_commit.timestamp }}
            VCS_REF=${{ github.sha }}

      - name: Scan Docker image
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-image-results.sarif'

  deploy:
    name: Deploy to Production
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/web-app \
            web-app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} \
            -n production
          kubectl rollout status deployment/web-app -n production

      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployment to production completed'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()
```

**Key Features**:
- Matrix builds (Node 18, 20, 22)
- Layer caching with GitHub Actions cache
- Security scanning (Trivy)
- Docker image scanning before deployment
- Zero-downtime rolling deployment
- Slack notifications

---

## 🔧 Pattern 4: Terraform Infrastructure (AWS Example)

### Production Infrastructure with Modules and Remote State

**Use Case**: Provision cloud infrastructure as code.

**main.tf**:
```hcl
terraform {
  required_version = ">= 1.9.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "terraform-state-prod"
    key            = "production/terraform.tfstate"
    region         = "us-west-2"
    encrypt        = true
    dynamodb_table = "terraform-locks"
    kms_key_id     = "arn:aws:kms:us-west-2:123456789012:key/abc123"
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "Terraform"
      Project     = var.project_name
    }
  }
}

# VPC Module
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = "${var.project_name}-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = false
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Terraform   = "true"
    Environment = var.environment
  }
}

# EKS Cluster Module
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "19.0.0"

  cluster_name    = "${var.project_name}-eks"
  cluster_version = "1.31"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access = true

  eks_managed_node_groups = {
    general = {
      min_size     = 2
      max_size     = 10
      desired_size = 3

      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"

      update_config = {
        max_unavailable_percentage = 50
      }

      labels = {
        role = "general"
      }

      tags = {
        NodeGroup = "general"
      }
    }
  }

  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
  }

  tags = {
    Environment = var.environment
  }
}
```

**variables.tf**:
```hcl
variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "Environment must be dev, staging, or production."
  }
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}
```

**outputs.tf**:
```hcl
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "eks_cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
  sensitive   = true
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}
```

**Usage**:
```bash
# Initialize
terraform init

# Plan with variable file
terraform plan -var-file="environments/production.tfvars"

# Apply with approval
terraform apply -var-file="environments/production.tfvars"

# State management
terraform state list
terraform state show module.vpc.aws_vpc.this[0]
```

**Key Features**:
- Remote state with S3 + DynamoDB locking
- Encrypted state with KMS
- Modular architecture (reusable VPC, EKS modules)
- Input validation
- Default tags for resource management
- Version pinning for reproducibility

---

## 📈 Pattern 5: Prometheus Monitoring (2.55.x)

### ServiceMonitor and Alert Rules

**Use Case**: Monitor Kubernetes applications with Prometheus.

**servicemonitor.yaml**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: web-app-metrics
  namespace: production
  labels:
    app: web-app
    prometheus: kube-prometheus
spec:
  selector:
    matchLabels:
      app: web-app
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
    relabelings:
    - sourceLabels: [__meta_kubernetes_pod_name]
      targetLabel: pod
    - sourceLabels: [__meta_kubernetes_namespace]
      targetLabel: namespace
```

**prometheusrule.yaml**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: web-app-alerts
  namespace: production
  labels:
    prometheus: kube-prometheus
    role: alert-rules
spec:
  groups:
  - name: web-app.rules
    interval: 30s
    rules:
    # High error rate alert
    - alert: HighErrorRate
      expr: |
        sum(rate(http_requests_total{job="web-app",status=~"5.."}[5m])) 
        / 
        sum(rate(http_requests_total{job="web-app"}[5m])) 
        > 0.05
      for: 5m
      labels:
        severity: critical
        component: web-app
      annotations:
        summary: "High error rate detected"
        description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"

    # High latency alert
    - alert: HighLatency
      expr: |
        histogram_quantile(0.95, 
          sum(rate(http_request_duration_seconds_bucket{job="web-app"}[5m])) by (le)
        ) > 1
      for: 10m
      labels:
        severity: warning
        component: web-app
      annotations:
        summary: "High request latency"
        description: "95th percentile latency is {{ $value }}s (threshold: 1s)"

    # Pod availability alert
    - alert: PodDown
      expr: |
        kube_deployment_status_replicas_available{deployment="web-app"} 
        < 
        kube_deployment_spec_replicas{deployment="web-app"} * 0.7
      for: 5m
      labels:
        severity: critical
        component: web-app
      annotations:
        summary: "Insufficient pod replicas"
        description: "Only {{ $value }} pods available (expected: {{ $labels.spec_replicas }})"

    # Memory pressure alert
    - alert: HighMemoryUsage
      expr: |
        container_memory_working_set_bytes{pod=~"web-app-.*"} 
        / 
        container_spec_memory_limit_bytes{pod=~"web-app-.*"} 
        > 0.9
      for: 5m
      labels:
        severity: warning
        component: web-app
      annotations:
        summary: "High memory usage"
        description: "Pod {{ $labels.pod }} using {{ $value | humanizePercentage }} of memory limit"
```

**Key Features**:
- 15-30 second scrape intervals (best practice)
- Relabeling for consistent metric naming
- Multi-condition alerts (error rate, latency, availability, resources)
- Severity labels for alert routing
- Human-readable annotations

---

## 🎨 Pattern 6: Grafana Dashboard (11.x)

### Production Dashboard with Variables and Templating

**Use Case**: Visualize Prometheus metrics with Grafana.

**Dashboard Configuration Highlights**:
- **Scenes-powered architecture** (Grafana 11.x)
- **Template variables** for namespace/pod filtering
- **Multi-panel layouts** (graph, stat, heatmap)
- **Color-coded thresholds** for quick status identification
- **PromQL queries** with label selectors

**Example Panels**:
1. **Request Rate Graph**: `sum(rate(http_requests_total{namespace="$namespace",pod=~"$pod"}[5m])) by (pod)`
2. **Error Rate Stat**: Threshold-based coloring (green < 1%, yellow < 5%, red ≥ 5%)
3. **Latency Heatmap**: Distribution visualization with histogram buckets

---

## 🔄 Pattern 7: ArgoCD GitOps Deployment

### Declarative Application Management

**Use Case**: Continuous deployment with GitOps workflow.

**application.yaml**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: web-app
  namespace: argocd
  finalizers:
  - resources-finalizer.argocd.argoproj.io
spec:
  project: production
  source:
    repoURL: https://github.com/org/web-app-k8s
    targetRevision: main
    path: overlays/production
    kustomize:
      version: v5.0.0
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
    - PruneLast=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  revisionHistoryLimit: 10
  ignoreDifferences:
  - group: apps
    kind: Deployment
    jsonPointers:
    - /spec/replicas
```

**Key Features**:
- Automated sync with self-healing
- Retry logic with exponential backoff
- Prune orphaned resources
- Ignore HPA-managed replica changes
- Project-level RBAC

---

## 🔐 Pattern 8: Kubernetes Secrets Management

### External Secrets with AWS Secrets Manager

**Use Case**: Manage secrets securely outside Kubernetes.

**externalsecret.yaml**:
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
---
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: production
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
```

**Key Features**:
- External secret provider (AWS Secrets Manager)
- Auto-refresh every hour
- Template-based secret construction
- IRSA authentication

---

## 🌊 Pattern 9: Blue-Green Deployment Strategy

### Zero-Downtime Release with Service Switching

**Use Case**: Deploy new version alongside current, switch traffic instantly.

**Switching Script**:
```bash
#!/bin/bash
# blue-green-switch.sh

set -e

NAMESPACE="production"
SERVICE="web-app"
NEW_VERSION="green"  # or "blue" to rollback

echo "Validating $NEW_VERSION deployment..."
kubectl rollout status deployment/web-app-$NEW_VERSION -n $NAMESPACE

echo "Running smoke tests on $NEW_VERSION..."
kubectl run smoke-test --rm -i --restart=Never \
  --image=curlimages/curl -- \
  http://web-app-$NEW_VERSION.$NAMESPACE.svc.cluster.local/health

echo "Switching traffic to $NEW_VERSION..."
kubectl patch service $SERVICE -n $NAMESPACE \
  -p "{\"spec\":{\"selector\":{\"version\":\"$NEW_VERSION\"}}}"

echo "Traffic switched to $NEW_VERSION. Monitor metrics for 10 minutes."
sleep 600

echo "Deployment successful. Old version can be scaled down."
kubectl scale deployment/web-app-$([ "$NEW_VERSION" = "blue" ] && echo "green" || echo "blue") \
  -n $NAMESPACE --replicas=0
```

**Key Features**:
- Two identical environments (blue/green)
- Instant traffic switch via service selector
- Easy rollback (switch back to previous version)
- No downtime during deployment

---

## 📊 Pattern 10: Helm Chart Structure (Production)

### Reusable Kubernetes Package

**Use Case**: Package, version, and distribute Kubernetes applications.

**Chart.yaml**:
```yaml
apiVersion: v2
name: web-app
description: Production-grade web application
version: 1.0.0
appVersion: "v1.0.0"
type: application
keywords:
  - web
  - production
maintainers:
  - name: DevOps Team
    email: devops@example.com
dependencies:
  - name: postgresql
    version: 12.x.x
    repository: https://charts.bitnami.com/bitnami
    condition: postgresql.enabled
```

**values.yaml**:
```yaml
replicaCount: 3

image:
  repository: myapp
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: app.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: app-tls
      hosts:
        - app.example.com

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

monitoring:
  enabled: true
  serviceMonitor:
    interval: 30s

postgresql:
  enabled: true
  auth:
    username: appuser
    database: appdb
```

**Usage**:
```bash
# Install chart
helm install web-app ./web-app \
  -f web-app/values-production.yaml \
  -n production \
  --create-namespace

# Upgrade with new values
helm upgrade web-app ./web-app \
  -f web-app/values-production.yaml \
  -n production \
  --atomic --timeout 10m

# Rollback to previous version
helm rollback web-app 1 -n production
```

**Key Features**:
- Environment-specific values files
- Dependency management (PostgreSQL subchart)
- Conditional resource creation
- Atomic upgrades (rollback on failure)

---

## 🚀 Best Practices Summary (2025 Edition)

### Container Security
1. **Use minimal base images** (Alpine, distroless)
2. **Pin image digests** for reproducibility
3. **Run as non-root user** (USER directive)
4. **Scan images** with Trivy/Snyk
5. **Multi-stage builds** to reduce attack surface

### Kubernetes Production Readiness
1. **Resource requests/limits** for all containers
2. **Liveness and readiness probes** for health checks
3. **HPA** for automatic scaling
4. **PodDisruptionBudget** for availability
5. **Network policies** for pod-to-pod security

### CI/CD Pipeline Optimization
1. **Matrix builds** for multi-version testing
2. **Caching strategies** (Docker layers, npm/pip cache)
3. **Security scanning** at every stage
4. **Parallel jobs** for faster feedback
5. **Workflow timeouts** (30 minutes recommended)
6. **Secrets management** with GitHub Secrets/Vault

### Infrastructure as Code
1. **Remote state** with locking (S3 + DynamoDB)
2. **State encryption** with KMS
3. **Modular architecture** for reusability
4. **Input validation** with variable constraints
5. **Version pinning** for providers and modules
6. **Plan before apply** (never auto-approve production)

### Monitoring & Observability
1. **Scrape intervals**: 15-30 seconds for most apps
2. **Label management**: Consistent naming, avoid high cardinality
3. **Alert thresholds**: Based on SLIs/SLOs
4. **Dashboard best practices**: Use template variables, organize by user role
5. **Distributed tracing**: OpenTelemetry for end-to-end visibility

### GitOps Workflow
1. **Git as single source of truth**
2. **Automated sync** with self-healing
3. **Declarative configuration** (YAML in Git)
4. **RBAC** for access control
5. **Audit trail** via Git history

---

## 📚 Official Documentation References

### Container & Orchestration
- **Kubernetes 1.31**: https://kubernetes.io/docs/
- **Docker 27.x**: https://docs.docker.com/
- **Helm 3.16**: https://helm.sh/docs/

### CI/CD
- **GitHub Actions**: https://docs.github.com/en/actions
- **ArgoCD 2.13**: https://argo-cd.readthedocs.io/

### Infrastructure as Code
- **Terraform 1.9**: https://developer.hashicorp.com/terraform/docs
- **Ansible 2.17**: https://docs.ansible.com/

### Monitoring
- **Prometheus 2.55**: https://prometheus.io/docs/
- **Grafana 11.x**: https://grafana.com/docs/grafana/latest/

---

## 🎯 When to Use This Skill

**Activate for**:
- Designing CI/CD pipelines for enterprise applications
- Building Kubernetes deployments with production-grade patterns
- Implementing GitOps workflows with ArgoCD
- Setting up monitoring with Prometheus/Grafana
- Writing Terraform modules for cloud infrastructure
- Containerizing applications with multi-stage Docker builds
- Implementing zero-downtime deployment strategies
- Configuring auto-scaling and resource management
- Establishing observability with distributed tracing

**Combines well with**:
- `moai-domain-backend`: Backend deployment optimization
- `moai-domain-frontend`: Frontend build pipelines
- `moai-domain-security`: Security scanning integration
- `moai-domain-database`: Database migration automation
- `moai-domain-testing`: Test automation in CI/CD

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-12  
**Status**: ✅ Production-Ready  
**Technology Stack**: Kubernetes 1.31, Docker 27.x, Terraform 1.9, Prometheus 2.55, Grafana 11.x  
**Patterns**: 10+ production-grade examples  
**Size**: ~30KB | ~950 lines
