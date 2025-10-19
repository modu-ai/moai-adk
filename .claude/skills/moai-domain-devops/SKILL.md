---
name: moai-domain-devops
description: CI/CD pipelines, Docker containerization, Kubernetes orchestration, and infrastructure as code
allowed-tools:
  - Read
  - Bash
tier: 4
auto-load: "false"
---

# DevOps Expert

## What it does

Provides expertise in continuous integration/deployment (CI/CD), Docker containerization, Kubernetes orchestration, and infrastructure as code (IaC) for automated deployment workflows.

## When to use

- "CI/CD 파이프라인", "Docker 컨테이너화", "Kubernetes 배포", "인프라 코드"
- Automatically invoked when working with DevOps projects
- DevOps SPEC implementation (`/alfred:2-build`)

## How it works

**CI/CD Pipelines**:
- **GitHub Actions**: Workflow automation (.github/workflows)
- **GitLab CI**: .gitlab-ci.yml configuration
- **Jenkins**: Pipeline as code (Jenkinsfile)
- **CircleCI**: .circleci/config.yml
- **Pipeline stages**: Build → Test → Deploy

**Docker Containerization**:
- **Dockerfile**: Multi-stage builds for optimization
- **docker-compose**: Local development environments
- **Image optimization**: Layer caching, alpine base images
- **Container registries**: Docker Hub, GitHub Container Registry

**Kubernetes Orchestration**:
- **Deployments**: Rolling updates, rollbacks
- **Services**: LoadBalancer, ClusterIP, NodePort
- **ConfigMaps/Secrets**: Configuration management
- **Helm charts**: Package management
- **Ingress**: Traffic routing

**Infrastructure as Code (IaC)**:
- **Terraform**: Cloud-agnostic provisioning
- **Ansible**: Configuration management
- **CloudFormation**: AWS-specific IaC
- **Pulumi**: Programmatic infrastructure

**Monitoring & Logging**:
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **ELK Stack**: Logging (Elasticsearch, Logstash, Kibana)

## Examples

### Example 1: GitHub Actions CI/CD Pipeline

**Workflow Structure**:
```yaml
# @CODE:CICD-001 | .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
      # 1단계: 코드 체크아웃
      - uses: actions/checkout@v3

      # 2단계: 의존성 설치
      - name: Install dependencies
        run: npm install

      # 3단계: 테스트
      - name: Run tests
        run: npm run test:coverage

      # 4단계: 빌드
      - name: Build
        run: npm run build

      # 5단계: 배포 (main만)
      - name: Deploy to production
        if: github.ref == 'refs/heads/main'
        run: |
          npm run build
          npm run deploy

# 효과:
# - 자동 테스트 (모든 PR)
# - 자동 배포 (main push)
# - 0 다운타임 ✅
```

### Example 2: Docker Containerization

**❌ Before (Large Image)**:
```dockerfile
# @CODE:DOCKER-001: 비효율
FROM ubuntu:20.04

RUN apt-get update && apt-get install -y \
    nodejs npm python3 gcc

COPY . /app
WORKDIR /app

RUN npm install && npm run build

EXPOSE 3000
CMD ["npm", "start"]

# 문제:
# - 이미지 크기: 1GB 이상
# - 레이어 많음 (빌드 시간 오래)
```

**✅ After (Multi-stage Build)**:
```dockerfile
# @CODE:DOCKER-001: 최적화
# Stage 1: Builder
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Stage 2: Runtime
FROM node:18-alpine

WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000
CMD ["node", "dist/index.js"]

# 개선:
# - 이미지 크기: 150MB (87% 감소!)
# - 빌드 시간: 30% 단축
# - 보안: 빌드 도구 제외
```

### Example 3: Kubernetes Deployment

**Manifest Structure**:
```yaml
# @CODE:K8S-001 | deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server

spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-server

  template:
    metadata:
      labels:
        app: api-server

    spec:
      containers:
      - name: api-server
        image: myregistry.azurecr.io/api:latest
        ports:
        - containerPort: 3000

        # 환경 변수
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: db-config
              key: host

        # 헬스 체크
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10

        # 리소스 제한
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"

---
apiVersion: v1
kind: Service
metadata:
  name: api-server

spec:
  selector:
    app: api-server
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

### Example 4: Infrastructure as Code (Terraform)

**Configuration**:
```hcl
# @CODE:IaC-001 | main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# EC2 인스턴스
resource "aws_instance" "api_server" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.medium"

  tags = {
    Name = "api-server"
  }
}

# RDS 데이터베이스
resource "aws_db_instance" "main" {
  identifier     = "main-db"
  engine         = "postgres"
  engine_version = "13.7"
  instance_class = "db.t3.micro"

  allocated_storage = 20
  username          = var.db_username
  password          = var.db_password

  skip_final_snapshot = true
}

# 출력
output "api_server_ip" {
  value = aws_instance.api_server.public_ip
}

output "db_endpoint" {
  value = aws_db_instance.main.endpoint
}
```

## Keywords

"CI/CD", "Docker", "Kubernetes", "Terraform", "IaC", "자동화", "배포", "파이프라인", "모니터링", "infrastructure as code", "container orchestration"

## Reference

- CI/CD pipelines: `.moai/memory/development-guide.md#CI-CD-파이프라인`
- Kubernetes guide: CLAUDE.md#Kubernetes-배포
- Infrastructure as Code: `.moai/memory/development-guide.md#Terraform-IaC`

## Works well with

- moai-domain-backend (서버 배포)
- moai-domain-security (보안 배포)
- moai-foundation-trust (배포 검증)
