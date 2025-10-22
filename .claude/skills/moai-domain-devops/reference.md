# DevOps - CI/CD and Infrastructure Reference

> Official documentation for Docker, Kubernetes, CI/CD pipelines, and infrastructure as code

---

## Official Documentation Links

### Container Technologies

| Tool | Version | Documentation | Status |
|------|---------|--------------|--------|
| **Docker** | 27.4.0 | https://docs.docker.com/ | ✅ Current (2025) |
| **Docker Compose** | 2.30 | https://docs.docker.com/compose/ | ✅ Current (2025) |
| **Kubernetes** | 1.32.0 | https://kubernetes.io/docs/ | ✅ Current (2025) |
| **Helm** | 3.16 | https://helm.sh/docs/ | ✅ Current (2025) |

### CI/CD Platforms

| Platform | Documentation | Status |
|----------|--------------|--------|
| **GitHub Actions** | https://docs.github.com/en/actions | ✅ Current (2025) |
| **GitLab CI** | https://docs.gitlab.com/ee/ci/ | ✅ Current (2025) |
| **Jenkins** | https://www.jenkins.io/doc/ | ✅ Current (2025) |
| **CircleCI** | https://circleci.com/docs/ | ✅ Current (2025) |
| **Azure DevOps** | https://learn.microsoft.com/en-us/azure/devops/ | ✅ Current (2025) |

### Infrastructure as Code

| Tool | Version | Documentation | Status |
|------|---------|--------------|--------|
| **Terraform** | 1.10.0 | https://developer.hashicorp.com/terraform/docs | ✅ Current (2025) |
| **Ansible** | 2.18 | https://docs.ansible.com/ | ✅ Current (2025) |
| **Pulumi** | 3.140 | https://www.pulumi.com/docs/ | ✅ Current (2025) |

---

## Docker Best Practices

### Dockerfile Optimization

**Multi-stage builds** (reduce image size):
```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

# Production stage
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

**Layer caching** (faster builds):
```dockerfile
# ❌ BAD (invalidates cache on any file change)
COPY . .
RUN npm install

# ✅ GOOD (cache dependencies separately)
COPY package*.json ./
RUN npm ci --only=production
COPY . .
```

**Security best practices**:
```dockerfile
# Use specific versions (not latest)
FROM node:20.11.0-alpine

# Run as non-root user
RUN addgroup -g 1001 appgroup && adduser -D -u 1001 -G appgroup appuser
USER appuser

# Use COPY instead of ADD
COPY --chown=appuser:appgroup . .

# Remove unnecessary packages
RUN apk add --no-cache --virtual .build-deps gcc musl-dev \
    && pip install -r requirements.txt \
    && apk del .build-deps
```

### Docker Compose

**Development environment**:
```yaml
version: '3.9'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    volumes:
      - .:/app
      - /app/node_modules
    environment:
      - NODE_ENV=development
    depends_on:
      - db
      - redis

  db:
    image: postgres:17.2-alpine
    environment:
      - POSTGRES_USER=app
      - POSTGRES_PASSWORD=secret
      - POSTGRES_DB=appdb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7.4.0-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

---

## Kubernetes Essentials

### Core Resources

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  labels:
    app: web
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: app
        image: myapp:1.0.0
        ports:
        - containerPort: 3000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
```

**Service**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

**ConfigMap**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  app.properties: |
    environment=production
    log.level=info
    max.connections=100
```

**Secret** (base64 encoded):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  url: cG9zdGdyZXNxbDovL3VzZXI6cGFzc0BkYjo1NDMyL2Ri
```

### kubectl Commands

```bash
# Apply resources
kubectl apply -f deployment.yaml

# Get resources
kubectl get pods
kubectl get deployments
kubectl get services

# Describe resource (detailed info)
kubectl describe pod <pod-name>

# View logs
kubectl logs <pod-name>
kubectl logs -f <pod-name>  # Follow logs

# Execute command in pod
kubectl exec -it <pod-name> -- /bin/sh

# Port forwarding (local testing)
kubectl port-forward <pod-name> 3000:3000

# Scale deployment
kubectl scale deployment web-app --replicas=5

# Rollback deployment
kubectl rollout undo deployment/web-app

# View rollout history
kubectl rollout history deployment/web-app
```

---

## CI/CD Pipeline Patterns

### GitHub Actions Workflow

**Complete CI/CD pipeline**:
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run tests
        run: npm test -- --coverage

      - name: Check coverage threshold
        run: |
          COVERAGE=$(jq '.total.lines.pct' coverage/coverage-summary.json)
          if (( $(echo "$COVERAGE < 85" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 85%"
            exit 1
          fi

  build:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

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
            type=semver,pattern={{version}}
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

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig.yaml
          export KUBECONFIG=kubeconfig.yaml

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/web-app \
            app=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main-${{ github.sha }}
          kubectl rollout status deployment/web-app
```

### GitLab CI/CD

```yaml
stages:
  - test
  - build
  - deploy

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

test:
  stage: test
  image: node:20-alpine
  cache:
    paths:
      - node_modules/
  script:
    - npm ci
    - npm run lint
    - npm test -- --coverage
    - |
      COVERAGE=$(jq '.total.lines.pct' coverage/coverage-summary.json)
      if (( $(echo "$COVERAGE < 85" | bc -l) )); then
        echo "Coverage $COVERAGE% is below 85%"
        exit 1
      fi
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  only:
    - main
    - develop

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - echo "$KUBE_CONFIG" | base64 -d > kubeconfig.yaml
    - export KUBECONFIG=kubeconfig.yaml
  script:
    - kubectl set image deployment/web-app app=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - kubectl rollout status deployment/web-app
  only:
    - main
  environment:
    name: production
    url: https://app.example.com
```

---

## Infrastructure as Code (Terraform)

### AWS Infrastructure Example

**main.tf**:
```hcl
terraform {
  required_version = ">= 1.10"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket = "terraform-state-bucket"
    key    = "production/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "main-vpc"
    Environment = var.environment
  }
}

# Subnets
resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "public-subnet-${count.index + 1}"
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "app-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "app-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = jsonencode([
    {
      name  = "app"
      image = var.app_image
      portMappings = [
        {
          containerPort = 3000
          protocol      = "tcp"
        }
      ]
      environment = [
        {
          name  = "NODE_ENV"
          value = "production"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = "/ecs/app"
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "app-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = var.app_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.public[*].id
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "app"
    container_port   = 3000
  }
}
```

**variables.tf**:
```hcl
variable "aws_region" {
  description = "AWS region"
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  default     = "production"
}

variable "app_image" {
  description = "Docker image for application"
  type        = string
}

variable "app_count" {
  description = "Number of app instances"
  default     = 2
}
```

**outputs.tf**:
```hcl
output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}
```

### Terraform Commands

```bash
# Initialize (download providers)
terraform init

# Format code
terraform fmt

# Validate configuration
terraform validate

# Plan changes (dry-run)
terraform plan

# Apply changes
terraform apply

# Destroy infrastructure
terraform destroy

# Show current state
terraform show

# Import existing resource
terraform import aws_instance.example i-1234567890abcdef0
```

---

## Monitoring and Logging

### Prometheus + Grafana

**prometheus.yml**:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

### ELK Stack (Elasticsearch, Logstash, Kibana)

**Filebeat configuration**:
```yaml
filebeat.inputs:
  - type: container
    paths:
      - '/var/lib/docker/containers/*/*.log'

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  indices:
    - index: "app-logs-%{+yyyy.MM.dd}"

processors:
  - add_docker_metadata: ~
  - decode_json_fields:
      fields: ["message"]
      target: ""
      overwrite_keys: true
```

---

## Security Best Practices

### Container Security

**Scan for vulnerabilities**:
```bash
# Trivy (vulnerability scanner)
trivy image myapp:1.0.0

# Docker Scout
docker scout cves myapp:1.0.0
```

**Sign images** (Docker Content Trust):
```bash
export DOCKER_CONTENT_TRUST=1
docker push myapp:1.0.0
```

### Secrets Management

**HashiCorp Vault**:
```bash
# Store secret
vault kv put secret/myapp/db password=secretpass

# Read secret
vault kv get secret/myapp/db
```

**Kubernetes Secrets** (use external secrets operator):
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: db-secret
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: db-secret
  data:
  - secretKey: password
    remoteRef:
      key: secret/myapp/db
      property: password
```

---

## Additional Resources

### Learning Platforms

- **Docker Documentation**: https://docs.docker.com/get-started/
- **Kubernetes Official Tutorial**: https://kubernetes.io/docs/tutorials/
- **Terraform Learn**: https://developer.hashicorp.com/terraform/tutorials
- **GitHub Actions Docs**: https://docs.github.com/en/actions/learn-github-actions

### Books

- "The Phoenix Project" by Gene Kim
- "The DevOps Handbook" by Gene Kim
- "Site Reliability Engineering" by Google

---

**Last Updated**: 2025-10-22
**Tool Versions**: Docker 27.4.0, Kubernetes 1.32.0, Terraform 1.10.0
**Standards**: 12-Factor App, GitOps, Infrastructure as Code
