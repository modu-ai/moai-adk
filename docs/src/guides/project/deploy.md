# 배포 완전 가이드

MoAI-ADK 프로젝트의 배포 과정을 포괄적으로 다루는 가이드입니다. Docker 컨테이너 배포부터 클라우드 서비스, CI/CD 파이프라인 구축까지 다양한 배포 환경과 현대적인 DevOps 관행을 상세하게 설명합니다.

## 목차

1. [개요](#개요)
2. [배포 아키텍처](#배포-아키텍처)
3. [Docker 컨테이너 배포](#docker-컨테이너-배포)
4. [클라우드 플랫폼 배포](#클라우드-플랫폼-배포)
5. [CI/CD 파이프라인 구축](#cicd-파이프라인-구축)
6. [환경별 배포 전략](#환경별-배포-전략)
7. [롤아웃 전략](#롤아웃-전략)
8. [모니터링과 로깅](#모니터링과-로깅)
9. [보안 및 규제 준수](#보안-및-규제-준수)
10. [배포 자동화](#배포-자동화)
11. [성능 최적화](#성능-최적화)
12. [문제 해결 및 복구](#문제-해결-및-복구)
13. [모범 사례](#모범-사례)

## 개요

MoAI-ADK 프로젝트의 배포는 단순히 코드를 서버에 올리는 것을 넘어, 안정적이고 확장 가능하며 안전한 시스템을 구축하는 과정입니다. Alfred의 SPEC-first 개발 방식은 배포 과정에서도 일관성과 추적성을 보장합니다.

### 배포 시스템의 핵심 원칙

1. **Automation First**: 모든 배포 과정은 자동화되어야 함
2. **Zero Downtime**: 사용자에게 서비스 중단이 없어야 함
3. **Rollback Ready**: 언제든지 이전 버전으로 롤백 가능해야 함
4. **Observable**: 배포 상태와 시스템 건강 상태를 실시간으로 파악 가능해야 함
5. **Secure**: 배포 과정 전체가 보안 규정을 준수해야 함

### MoAI-ADK 배포의 특징

- **SPEC-Driven Deployment**: SPEC 문서를 기반으로 한 배포 계획
- **TAG-Based Tracking**: 배포된 모든 구성요소의 TAG 추적
- **Automated Quality Gates**: 배포 전 자동 품질 검증
- **Progressive Deployment**: 점진적 배포 지원
- **Environment Consistency**: 모든 환경의 일관성 보장

## 배포 아키텍처

### 현대적인 배포 아키텍처

```mermaid
graph TD
    A[개발자] --> B[Git Push]
    B --> C[CI/CD Pipeline]
    C --> D[Build & Test]
    D --> E[Security Scan]
    E --> F[Container Build]
    F --> G[Artifact Registry]
    G --> H[Deployment Strategy]
    H --> I[Staging]
    I --> J[Automated Tests]
    J --> K{Approval?}
    K -->|Yes| L[Production Deployment]
    K -->|No| M[Rollback]
    L --> N[Health Check]
    N --> O[Monitoring]
    O --> P[Alerting]
```

### MoAI-ADK 배포 워크플로우

```mermaid
sequenceDiagram
    participant Dev as 개발자
    participant Git as Git Repository
    participant CI as CI/CD Pipeline
    participant QA as Quality Gate
    participant Deploy as Deployment System
    participant Prod as Production

    Dev->>Git: /alfred:3-sync
    Git->>CI: Trigger Pipeline
    CI->>CI: Build Application
    CI->>CI: Run Tests (TRUST 검증)
    CI->>QA: Quality Gate Check
    QA->>CI: Pass/Fail
    CI->>Deploy: Create Docker Image
    Deploy->>Deploy: Deploy to Staging
    Deploy->>QA: Automated E2E Tests
    QA->>Deploy: Approval
    Deploy->>Prod: Blue-Green Deployment
    Prod->>Deploy: Health Check
    Deploy->>Dev: Deployment Report
```

## Docker 컨테이너 배포

### 기본 Dockerfile 설정

MoAI-ADK 프로젝트를 위한 최적화된 Dockerfile:

```dockerfile
# Multi-stage build for production optimization
FROM python:3.13-slim as builder

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PIP_NO_CACHE_DIR=1
ENV PIP_DISABLE_PIP_VERSION_CHECK=1

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install uv for fast package management
RUN pip install uv

# Create virtual environment
RUN uv venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy dependency files
COPY pyproject.toml uv.lock ./

# Install dependencies
RUN uv pip install --system -e .

# Production stage
FROM python:3.13-slim as production

# Create non-root user
RUN groupadd -r appuser && useradd -r -g appuser appuser

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PATH="/opt/venv/bin:$PATH"

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy virtual environment from builder stage
COPY --from=builder /opt/venv /opt/venv

# Copy application code
WORKDIR /app
COPY src/ ./src/
COPY .moai/ ./.moai/
COPY CLAUDE.md ./

# Set ownership
RUN chown -R appuser:appuser /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Expose port
EXPOSE 8000

# Run application
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose 설정

#### 개발 환경용 Docker Compose

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: builder
    ports:
      - "8000:8000"
    environment:
      - DEBUG=true
      - DATABASE_URL=postgresql://postgres:password@db:5432/myapp_dev
      - REDIS_URL=redis://redis:6379/0
    volumes:
      - ./src:/app/src
      - ./.moai:/app/.moai
      - ./CLAUDE.md:/app/CLAUDE.md
    depends_on:
      - db
      - redis
    command: uvicorn src.main:app --host 0.0.0.0 --port 8000 --reload

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=myapp_dev
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx/dev.conf:/etc/nginx/nginx.conf
    depends_on:
      - app

volumes:
  postgres_data:
  redis_data:
```

#### 프로덕션 환경용 Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    image: ${REGISTRY}/myapp:${VERSION}
    restart: unless-stopped
    environment:
      - DEBUG=false
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SECRET_KEY=${SECRET_KEY}
    depends_on:
      - db
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  db:
    image: postgres:15
    restart: unless-stopped
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/prod.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes 배포 설정

#### 네임스페이스 및 ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: myapp-prod
  labels:
    name: myapp-prod
    environment: production

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
  namespace: myapp-prod
data:
  DEBUG: "false"
  LOG_LEVEL: "INFO"
  REDIS_URL: "redis://redis-service:6379/0"
  SPEC_VERSION: "v1.0.0"
  TRUST_LEVEL: "strict"
```

#### Secret 설정

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
  namespace: myapp-prod
type: Opaque
data:
  DATABASE_URL: <base64-encoded-database-url>
  SECRET_KEY: <base64-encoded-secret-key>
  JWT_SECRET: <base64-encoded-jwt-secret>
```

#### Deployment 설정

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deployment
  namespace: myapp-prod
  labels:
    app: myapp
    version: v1.0.0
    spec-id: SPEC-DEPLOY-001
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1.0.0
        spec-id: SPEC-DEPLOY-001
    spec:
      containers:
      - name: myapp
        image: myregistry.com/myapp:v1.0.0
        ports:
        - containerPort: 8000
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: DATABASE_URL
        - name: SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: SECRET_KEY
        envFrom:
        - configMapRef:
            name: myapp-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
```

#### Service 및 Ingress 설정

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
  namespace: myapp-prod
  labels:
    app: myapp
spec:
  selector:
    app: myapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp-ingress
  namespace: myapp-prod
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  tls:
  - hosts:
    - api.myapp.com
    secretName: myapp-tls
  rules:
  - host: api.myapp.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myapp-service
            port:
              number: 80
```

## 클라우드 플랫폼 배포

### AWS 배포

#### ECS Fargate 설정

```json
{
  "family": "myapp-task-definition",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "myapp",
      "image": "ACCOUNT.dkr.ecr.REGION.amazonaws.com/myapp:v1.0.0",
      "portMappings": [
        {
          "containerPort": 8000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DEBUG",
          "value": "false"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT:secret:myapp/database-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/myapp",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8000/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

#### CloudFormation 템플릿

```yaml
# cloudformation.yml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'MoAI-ADK Application Infrastructure'

Parameters:
  Environment:
    Type: String
    Default: production
    AllowedValues: [development, staging, production]
  Version:
    Type: String
    Default: v1.0.0

Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true
      EnableDnsSupport: true
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-myapp-vpc'
        - Key: Project
          Value: myapp
        - Key: SPEC
          Value: SPEC-INFRA-001

  PublicSubnet1:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.1.0/24
      AvailabilityZone: !Select [0, !GetAZs '']
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-public-subnet-1'

  PublicSubnet2:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.2.0/24
      AvailabilityZone: !Select [1, !GetAZs '']
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: !Sub '${Environment}-public-subnet-2'

  ECSCluster:
    Type: AWS::ECS::Cluster
    Properties:
      ClusterName: !Sub '${Environment}-myapp-cluster'
      CapacityProviders:
        - FARGATE
        - FARGATE_SPOT
      DefaultCapacityProviderStrategy:
        - CapacityProvider: FARGATE
          Weight: 1

  TaskDefinition:
    Type: AWS::ECS::TaskDefinition
    Properties:
      Family: !Sub '${Environment}-myapp'
      Cpu: 512
      Memory: 1024
      NetworkMode: awsvpc
      RequiresCompatibilities:
        - FARGATE
      ExecutionRoleArn: !Ref ECSExecutionRole
      TaskRoleArn: !Ref ECSTaskRole
      ContainerDefinitions:
        - Name: myapp
          Image: !Sub '${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/myapp:${Version}'
          PortMappings:
            - ContainerPort: 8000
          Environment:
            - Name: ENVIRONMENT
              Value: !Ref Environment
          LogConfiguration:
            LogDriver: awslogs
            Options:
              awslogs-group: !Ref CloudWatchLogsGroup
              awslogs-region: !Ref AWS::Region
              awslogs-stream-prefix: ecs

  Service:
    Type: AWS::ECS::Service
    Properties:
      ServiceName: !Sub '${Environment}-myapp-service'
      Cluster: !Ref ECSCluster
      TaskDefinition: !Ref TaskDefinition
      DesiredCount: 2
      LaunchType: FARGATE
      NetworkConfiguration:
        AwsvpcConfiguration:
          SecurityGroups:
            - !Ref SecurityGroup
          Subnets:
            - !Ref PublicSubnet1
            - !Ref PublicSubnet2
          AssignPublicIp: ENABLED
      LoadBalancers:
        - ContainerName: myapp
          ContainerPort: 8000
          TargetGroupArn: !Ref TargetGroup

  LoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: !Sub '${Environment}-myapp-alb'
      Scheme: internet-facing
      Type: application
      Subnets:
        - !Ref PublicSubnet1
        - !Ref PublicSubnet2
      SecurityGroups:
        - !Ref LoadBalancerSecurityGroup

  TargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: !Sub '${Environment}-myapp-tg'
      Port: 80
      Protocol: HTTP
      VpcId: !Ref VPC
      HealthCheckProtocol: HTTP
      HealthCheckPort: traffic-port
      HealthCheckPath: /health
      Matcher:
        HttpCode: '200'
      TargetType: ip

Outputs:
  LoadBalancerDNS:
    Description: Load Balancer DNS Name
    Value: !GetAtt LoadBalancer.DNSName
    Export:
      Name: !Sub '${Environment}-LoadBalancerDNS'
```

### Google Cloud 배포

#### Cloud Run 설정

```yaml
# cloudrun.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: myapp
  namespace: production
  annotations:
    run.googleapis.com/ingress: all
    run.googleapis.com/execution-environment: gen2
spec:
  template:
    metadata:
      annotations:
        run.googleapis.com/cpu-throttling: "false"
        run.googleapis.com/memory: "512Mi"
        run.googleapis.com/cpu: "1"
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "10"
    spec:
      containerConcurrency: 100
      timeoutSeconds: 300
      containers:
      - image: gcr.io/PROJECT_ID/myapp:v1.0.0
        ports:
        - containerPort: 8000
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: database-url
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### Terraform 구성

```hcl
# main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# VPC 네트워크
resource "google_compute_network" "vpc" {
  name                    = "${var.environment}-myapp-vpc"
  auto_create_subnetworks = false
}

# 서브넷
resource "google_compute_subnetwork" "subnet" {
  name          = "${var.environment}-myapp-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.vpc.id
}

# Cloud Run 서비스
resource "google_cloud_run_service" "myapp" {
  name     = "${var.environment}-myapp"
  location = var.region

  template {
    metadata {
      annotations = {
        "run.googleapis.com/ingress"                    = "all"
        "run.googleapis.com/execution-environment"     = "gen2"
        "autoscaling.knative.dev/minScale"             = "1"
        "autoscaling.knative.dev/maxScale"             = "10"
      }
    }

    spec {
      container_concurrency = 100
      timeout_seconds        = 300

      containers {
        image = "gcr.io/${var.project_id}/myapp:${var.version}"

        ports {
          container_port = 8000
        }

        env {
          name  = "ENVIRONMENT"
          value = var.environment
        }

        env {
          name = "DATABASE_URL"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret_version.db_url.secret_id
              key  = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "1"
            memory = "512Mi"
          }
        }

        liveness_probe {
          http_get {
            path = "/health"
            port = 8000
          }
          initial_delay_seconds = 10
          period_seconds        = 10
        }

        readiness_probe {
          http_get {
            path = "/ready"
            port = 8000
          }
          initial_delay_seconds = 5
          period_seconds        = 5
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

# IAM 정책
resource "google_cloud_run_service_iam_member" "public" {
  location = google_cloud_run_service.myapp.location
  project  = google_cloud_run_service.myapp.project
  service  = google_cloud_run_service.myapp.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# Secret Manager
resource "google_secret_manager_secret" "db_url" {
  provider = google-beta
  secret_id = "${var.environment}-myapp-db-url"

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "db_url" {
  secret      = google_secret_manager_secret.db_url.id
  secret_data = var.database_url
}
```

### Azure 배포

#### ARM 템플릿

```json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "environment": {
      "type": "string",
      "defaultValue": "production",
      "allowedValues": ["development", "staging", "production"]
    },
    "version": {
      "type": "string",
      "defaultValue": "v1.0.0"
    }
  },
  "resources": [
    {
      "type": "Microsoft.ContainerInstance/containerGroups",
      "apiVersion": "2021-09-01",
      "name": "[concat(parameters('environment'), '-myapp-container-group')]",
      "location": "[resourceGroup().location]",
      "properties": {
        "containers": [
          {
            "name": "myapp",
            "properties": {
              "image": "[concat('.acr.io/myapp:', parameters('version'))]",
              "ports": [
                {
                  "port": 8000,
                  "protocol": "TCP"
                }
              ],
              "environmentVariables": [
                {
                  "name": "ENVIRONMENT",
                  "value": "[parameters('environment')]"
                },
                {
                  "name": "DATABASE_URL",
                  "secureValue": "[reference(resourceId('Microsoft.KeyVault/vaults/secrets', 'myapp-kv', 'database-url'), '2021-06-01-preview').value]"
                }
              ],
              "resources": {
                "requests": {
                  "cpu": 1.0,
                  "memoryInGB": 1.0
                }
              },
              "livenessProbe": {
                "exec": {
                  "command": [
                    "/bin/sh",
                    "-c",
                    "curl -f http://localhost:8000/health || exit 1"
                  ]
                },
                "periodSeconds": 30,
                "timeoutSeconds": 10,
                "failureThreshold": 3
              }
            }
          }
        ],
        "osType": "Linux",
        "restartPolicy": "Always",
        "ipAddress": {
          "type": "Public",
          "ports": [
            {
              "port": 8000,
              "protocol": "TCP"
            }
          ]
        }
      }
    }
  ],
  "outputs": {
    "containerGroupIP": {
      "type": "string",
      "value": "[reference(resourceId('Microsoft.ContainerInstance/containerGroups', concat(parameters('environment'), '-myapp-container-group')), '2021-09-01').ipAddress.ip]"
    }
  }
}
```

## CI/CD 파이프라인 구축

### GitHub Actions 워크플로우

#### 메인 CI/CD 파이프라인

```yaml
# .github/workflows/deploy.yml
name: Deploy Application

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # 코드 품질 및 보안 검사
  quality-check:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.13'

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
        pip install -r requirements-dev.txt

    - name: Run linting
      run: |
        ruff check src/
        ruff format --check src/

    - name: Run type checking
      run: mypy src/

    - name: Run security scan
      run: |
        bandit -r src/
        safety check

    - name: Run tests
      run: |
        pytest tests/ --cov=src/ --cov-report=xml

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.xml

  # Docker 이미지 빌드
  build-image:
    needs: quality-check
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

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
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  # Staging 환경 배포
  deploy-staging:
    needs: build-image
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2

    - name: Deploy to ECS Staging
      run: |
        aws ecs update-service \
          --cluster myapp-staging-cluster \
          --service myapp-staging-service \
          --force-new-deployment

    - name: Wait for deployment
      run: |
        aws ecs wait services-stable \
          --cluster myapp-staging-cluster \
          --services myapp-staging-service

    - name: Run smoke tests
      run: |
        chmod +x ./scripts/smoke-tests.sh
        ./scripts/smoke-tests.sh staging

  # 프로덕션 환경 배포
  deploy-production:
    needs: build-image
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2

    - name: Deploy to ECS Production
      run: |
        aws ecs update-service \
          --cluster myapp-production-cluster \
          --service myapp-production-service \
          --force-new-deployment

    - name: Wait for deployment
      run: |
        aws ecs wait services-stable \
          --cluster myapp-production-cluster \
          --services myapp-production-service

    - name: Run health checks
      run: |
        chmod +x ./scripts/health-checks.sh
        ./scripts/health-checks.sh production

    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

#### Blue-Green 배포 워크플로우

```yaml
# .github/workflows/blue-green-deploy.yml
name: Blue-Green Deployment

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Target environment'
        required: true
        default: 'staging'
        type: choice
        options:
        - staging
        - production
      version:
        description: 'Application version'
        required: true
        default: 'latest'

env:
  CLUSTER_NAME: myapp-cluster
  SERVICE_NAME: myapp-service

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.environment }}
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2

    - name: Get current task definition
      id: current-task
      run: |
        TASK_DEF=$(aws ecs describe-services \
          --cluster ${{ env.CLUSTER_NAME }} \
          --services ${{ env.SERVICE_NAME }} \
          --query 'services[0].taskDefinition' \
          --output text)
        echo "task-definition=$TASK_DEF" >> $GITHUB_OUTPUT

    - name: Create new task definition
      id: new-task
      run: |
        # Create new task definition with new image version
        NEW_TASK_DEF=$(aws ecs register-task-definition \
          --cli-input-json "$(cat task-definition.json | \
          jq '.containerDefinitions[0].image = "${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.event.inputs.version }}"')" \
          --query 'taskDefinition.taskDefinitionArn' \
          --output text)
        echo "task-definition=$NEW_TASK_DEF" >> $GITHUB_OUTPUT

    - name: Update service with new task definition
      run: |
        aws ecs update-service \
          --cluster ${{ env.CLUSTER_NAME }} \
          --service ${{ env.SERVICE_NAME }} \
          --task-definition ${{ steps.new-task.outputs.task-definition }} \
          --force-new-deployment

    - name: Wait for new tasks to start
      run: |
        aws ecs wait services-stable \
          --cluster ${{ env.CLUSTER_NAME }} \
          --services ${{ env.SERVICE_NAME }}

    - name: Run smoke tests
      run: |
        chmod +x ./scripts/smoke-tests.sh
        ./scripts/smoke-tests.sh ${{ github.event.inputs.environment }}

    - name: Update DNS to point to new version
      if: success()
      run: |
        # Update Route53 or load balancer target
        ./scripts/update-dns.sh ${{ github.event.inputs.environment }} new

    - name: Rollback on failure
      if: failure()
      run: |
        aws ecs update-service \
          --cluster ${{ env.CLUSTER_NAME }} \
          --service ${{ env.SERVICE_NAME }} \
          --task-definition ${{ steps.current-task.outputs.task-definition }} \
          --force-new-deployment

        aws ecs wait services-stable \
          --cluster ${{ env.CLUSTER_NAME }} \
          --services ${{ env.SERVICE_NAME }}

        ./scripts/update-dns.sh ${{ github.event.inputs.environment }} old
```

### GitLab CI/CD 설정

```yaml
# .gitlab-ci.yml
stages:
  - validate
  - test
  - build
  - deploy-staging
  - deploy-production

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"

# 코드 품질 검증
validate:
  stage: validate
  image: python:3.13
  before_script:
    - pip install -r requirements-dev.txt
  script:
    - ruff check src/
    - ruff format --check src/
    - mypy src/
  artifacts:
    reports:
      junit: reports/junit.xml
    paths:
      - reports/
    expire_in: 1 week

# 테스트 실행
test:
  stage: test
  image: python:3.13
  services:
    - postgres:15
    - redis:7
  variables:
    POSTGRES_DB: test_db
    POSTGRES_USER: test_user
    POSTGRES_PASSWORD: test_pass
    DATABASE_URL: postgresql://test_user:test_pass@postgres:5432/test_db
    REDIS_URL: redis://redis:6379/0
  before_script:
    - pip install -r requirements.txt
    - pip install -r requirements-dev.txt
  script:
    - pytest tests/ --cov=src/ --cov-report=xml --cov-report=html --junitxml=reports/junit.xml
  coverage: '/TOTAL.+?(\d+\%)$/'
  artifacts:
    reports:
      junit: reports/junit.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - htmlcov/
      - coverage.xml
    expire_in: 1 week

# Docker 이미지 빌드
build:
  stage: build
  image: docker:24.0.5
  services:
    - docker:24.0.5-dind
  before_script:
    - echo $CI_REGISTRY_PASSWORD | docker login -u $CI_REGISTRY_USER --password-stdin $CI_REGISTRY
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker build -t $CI_REGISTRY_IMAGE:latest .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
  only:
    - main
    - develop

# Staging 환경 배포
deploy-staging:
  stage: deploy-staging
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -L https://github.com/aws/aws-cli/releases/download/v2.13.25/awscliv2.zip -o awscliv2.zip
    - unzip awscliv2.zip
    - ./aws/install
  script:
    - aws ecs update-service --cluster myapp-staging --service myapp-service --force-new-deployment
    - aws ecs wait services-stable --cluster myapp-staging --services myapp-service
  environment:
    name: staging
    url: https://staging-api.myapp.com
  only:
    - develop

# 프로덕션 환경 배포
deploy-production:
  stage: deploy-production
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -L https://github.com/aws/aws-cli/releases/download/v2.13.25/awscliv2.zip -o awscliv2.zip
    - unzip awscliv2.zip
    - ./aws/install
  script:
    - aws ecs update-service --cluster myapp-production --service myapp-service --force-new-deployment
    - aws ecs wait services-stable --cluster myapp-production --services myapp-service
  environment:
    name: production
    url: https://api.myapp.com
  when: manual
  only:
    - main
```

## 환경별 배포 전략

### 개발 환경 (Development)

#### 목적 및 특징
- **목적**: 개발자 빠른 피드백 루프
- **특징**: 빠른 배포, 자동 재시작, 상세 로깅

#### 설정 예시

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8000:8000"
    environment:
      - DEBUG=true
      - LOG_LEVEL=DEBUG
      - RELOAD=true
    volumes:
      - ./src:/app/src
      - ./.moai:/app/.moai
    depends_on:
      - db
      - redis
    command: uvicorn src.main:app --host 0.0.0.0 --port 8000 --reload

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=myapp_dev
      - POSTGRES_USER=dev
      - POSTGRES_PASSWORD=dev123
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # 개발 도구들
  adminer:
    image: adminer
    ports:
      - "8080:8080"
    depends_on:
      - db

  redis-commander:
    image: rediscommander/redis-commander:latest
    environment:
      - REDIS_HOSTS=local:redis:6379
    ports:
      - "8081:8081"
    depends_on:
      - redis

volumes:
  postgres_dev_data:
```

#### 개발 환경 배포 스크립트

```bash
#!/bin/bash
# scripts/deploy-dev.sh

set -e

echo "<span class="material-icons">rocket_launch</span> Deploying to development environment..."

# 환경 변수 설정
export COMPOSE_PROJECT_NAME="myapp-dev"
export COMPOSE_FILE="docker-compose.dev.yml"

# 기존 컨테이너 정리
echo "🧹 Cleaning up existing containers..."
docker-compose down --remove-orphans

# 최신 이미지 빌드
echo "🔨 Building development images..."
docker-compose build

# 데이터베이스 마이그레이션
echo "<span class="material-icons">analytics</span> Running database migrations..."
docker-compose run --rm app alembic upgrade head

# 서비스 시작
echo "<span class="material-icons">rocket_launch</span> Starting development services..."
docker-compose up -d

# 상태 확인
echo "📋 Checking service status..."
sleep 10
docker-compose ps

# 상태 검사
echo "<span class="material-icons">search</span> Running health checks..."
./scripts/health-checks.sh dev

echo "<span class="material-icons">check_circle</span> Development environment deployed successfully!"
echo "<span class="material-icons">analytics</span> App available at: http://localhost:8000"
echo "<span class="material-icons">storage</span>  Database available at: localhost:5432"
echo "<span class="material-icons">settings</span> Redis available at: localhost:6379"
echo "🎛️  Adminer available at: http://localhost:8080"
echo "📱 Redis Commander available at: http://localhost:8081"
```

### 스테이징 환경 (Staging)

#### 목적 및 특징
- **목적**: 프로덕션 전 최종 검증
- **특징**: 프로덕션과 유사한 환경, 자동화된 테스트

#### 설정 예시

```yaml
# docker-compose.staging.yml
version: '3.8'

services:
  app:
    image: ${REGISTRY}/myapp:${VERSION}
    restart: unless-stopped
    environment:
      - DEBUG=false
      - LOG_LEVEL=INFO
      - ENVIRONMENT=staging
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SECRET_KEY=${SECRET_KEY}
    depends_on:
      - db
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  db:
    image: postgres:15
    restart: unless-stopped
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_staging_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_staging_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/staging.conf:/etc/nginx/nginx.conf
      - ./ssl/staging:/etc/nginx/ssl
    depends_on:
      - app
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # 모니터링
  prometheus:
    image: prom/prometheus:latest
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus-staging.yml:/etc/prometheus/prometheus.yml
      - prometheus_staging_data:/prometheus

  grafana:
    image: grafana/grafana:latest
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana_staging_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards

volumes:
  postgres_staging_data:
  redis_staging_data:
  prometheus_staging_data:
  grafana_staging_data:
```

#### 스테이징 배포 스크립트

```bash
#!/bin/bash
# scripts/deploy-staging.sh

set -e

VERSION=${1:-latest}
ENVIRONMENT="staging"

echo "<span class="material-icons">rocket_launch</span> Deploying version $VERSION to staging environment..."

# 환경 변수 로드
source .env.staging

# 이미지 풀
echo "📦 Pulling application image..."
docker pull ${REGISTRY}/myapp:${VERSION}

# 데이터베이스 백업
echo "💾 Creating database backup..."
./scripts/backup-db.sh $ENVIRONMENT

# 블루-그린 배포
echo "🔄 Starting blue-green deployment..."

# 현재 실행 중인 서비스 확인
CURRENT_SERVICE=$(docker-compose -f docker-compose.staging.yml ps -q app)

if [ -n "$CURRENT_SERVICE" ]; then
    echo "📋 Current service found: $CURRENT_SERVICE"

    # 새 서비스 시작 (green)
    echo "🌱 Starting new service (green)..."
    VERSION=$VERSION docker-compose -f docker-compose.staging.yml up -d app

    # 상태 검사
    echo "<span class="material-icons">search</span> Running health checks on new service..."
    ./scripts/wait-for-health.sh http://staging-api.myapp.com/health

    # 스모크 테스트
    echo "🧪 Running smoke tests..."
    ./scripts/smoke-tests.sh $ENVIRONMENT

    # 트래픽 전환
    echo "🔄 Switching traffic to new service..."
    ./scripts/switch-traffic.sh $ENVIRONMENT new

    # 이전 서비스 정리
    echo "🧹 Cleaning up old service (blue)..."
    docker-compose -f docker-compose.staging.yml stop app
else
    echo "🌱 Starting initial service..."
    VERSION=$VERSION docker-compose -f docker-compose.staging.yml up -d
fi

# 배포 확인
echo "<span class="material-icons">check_circle</span> Deployment completed successfully!"
echo "<span class="material-icons">language</span> Application available at: https://staging-api.myapp.com"
echo "<span class="material-icons">analytics</span> Monitoring available at: https://staging-grafana.myapp.com"

# 알림 발송
./scripts/notify-deployment.sh $ENVIRONMENT $VERSION success
```

### 프로덕션 환경 (Production)

#### 목적 및 특징
- **목적**: 실제 서비스 운영
- **특징**: 고가용성, 모니터링, 보안, 롤백 준비

#### 설정 예시

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    image: ${REGISTRY}/myapp:${VERSION}
    restart: unless-stopped
    environment:
      - DEBUG=false
      - LOG_LEVEL=WARNING
      - ENVIRONMENT=production
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SECRET_KEY=${SECRET_KEY}
      - SENTRY_DSN=${SENTRY_DSN}
    depends_on:
      - db
      - redis
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s

  # 로드 밸런서 및 리버스 프록시
  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/prod.conf:/etc/nginx/nginx.conf
      - ./ssl/prod:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - app
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M

  # 데이터베이스 (PostgreSQL)
  db:
    image: postgres:15
    restart: unless-stopped
    environment:
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_prod_data:/var/lib/postgresql/data
      - ./backups:/backups
      - ./postgresql/postgresql.conf:/etc/postgresql/postgresql.conf
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G

  # Redis (캐시 및 세션)
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server /usr/local/etc/redis/redis.conf
    volumes:
      - redis_prod_data:/data
      - ./redis/redis.conf:/usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

  # 모니터링 스택
  prometheus:
    image: prom/prometheus:latest
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus-prod.yml:/etc/prometheus/prometheus.yml
      - prometheus_prod_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_INSTALL_PLUGINS=grafana-piechart-panel
    volumes:
      - grafana_prod_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

  # 로그 수집
  loki:
    image: grafana/loki:latest
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - ./monitoring/loki-config.yml:/etc/loki/local-config.yaml
      - loki_prod_data:/loki
    command: -config.file=/etc/loki/local-config.yaml

  promtail:
    image: grafana/promtail:latest
    restart: unless-stopped
    volumes:
      - ./monitoring/promtail-config.yml:/etc/promtail/config.yml
      - ./logs:/var/log/app
    command: -config.file=/etc/promtail/config.yml

volumes:
  postgres_prod_data:
  redis_prod_data:
  prometheus_prod_data:
  grafana_prod_data:
  loki_prod_data:
```

## 롤아웃 전략

### Blue-Green 배포

Blue-Green 배포는 두 개의 동일한 환경을 유지하며 트래픽을 전환하는 방식입니다.

#### Blue-Green 배포 흐름

```mermaid
graph TD
    A[현재 Blue 환경] --> B[Green 환경 구축]
    B --> C[Green 환경 테스트]
    C --> D{테스트 통과?}
    D -->|Yes| E[트래픽을 Green으로 전환]
    D -->|No| F[Green 환경 폐기]
    E --> G[Blue 환경 대기]
    G --> H[문제 발생 시]
    H --> I[트래픽을 Blue로 롤백]
    E --> J[Green을 새로운 Blue로]
    F --> A
    I --> A
    J --> A
```

#### Blue-Green 배포 스크립트

```bash
#!/bin/bash
# scripts/blue-green-deploy.sh

set -e

VERSION=${1:-latest}
ENVIRONMENT=${2:-staging}

echo "🔄 Starting Blue-Green deployment of version $VERSION to $ENVIRONMENT"

# 설정 로드
source .env.$ENVIRONMENT

# 현재 활성 환경 확인
CURRENT_ENV=$(curl -s https://api.myapp.com/info | jq -r '.environment')
echo "📋 Current active environment: $CURRENT_ENV"

if [ "$CURRENT_ENV" = "blue" ]; then
    NEW_ENV="green"
    OLD_ENV="blue"
else
    NEW_ENV="blue"
    OLD_ENV="green"
fi

echo "🌱 Deploying to $NEW_ENV environment"

# 새 환경 설정 업데이트
sed -i "s/ENVIRONMENT=.*/ENVIRONMENT=$NEW_ENV/" .env.$ENVIRONMENT
sed -i "s/VERSION=.*/VERSION=$VERSION/" .env.$ENVIRONMENT

# 새 환경 배포
echo "<span class="material-icons">rocket_launch</span> Starting $NEW_ENV environment..."
docker-compose -f docker-compose.$ENVIRONMENT.yml up -d

# 상태 검사 대기
echo "⏳ Waiting for $NEW_ENV to be healthy..."
./scripts/wait-for-health.sh https://$NEW_ENV-api.myapp.com/health

# 스모크 테스트
echo "🧪 Running smoke tests on $NEW_ENV..."
./scripts/smoke-tests.sh $NEW_ENV

# 트래픽 전환
echo "🔄 Switching traffic to $NEW_ENV..."
./scripts/switch-traffic.sh $ENVIRONMENT $NEW_ENV

# 트래픽 전환 확인
echo "<span class="material-icons">check_circle</span> Traffic switched to $NEW_ENV"
sleep 30

# 최종 상태 검사
echo "<span class="material-icons">search</span> Final health check..."
curl -f https://api.myapp.com/health || {
    echo "<span class="material-icons">cancel</span> Health check failed, rolling back..."
    ./scripts/switch-traffic.sh $ENVIRONMENT $OLD_ENV
    exit 1
}

# 이전 환경 정리 (선택적)
echo "🧹 Cleaning up $OLD_ENV environment..."
docker-compose -f docker-compose.$ENVIRONMENT.yml stop

echo "<span class="material-icons">check_circle</span> Blue-Green deployment completed successfully!"
echo "<span class="material-icons">language</span> Application running on: https://api.myapp.com"
echo "<span class="material-icons">analytics</span> Environment: $NEW_ENV"
```

### Canary 배포

Canary 배포는 점진적으로 새 버전을 일부 사용자에게만 노출하는 방식입니다.

#### Canary 배포 흐름

```mermaid
graph TD
    A[현재 버전 100%] --> B[Canary 5% 배포]
    B --> C[모니터링]
    C --> D{상태 양호?}
    D -->|Yes| E[Canary 25%]
    D -->|No| F[롤백]
    E --> G[모니터링]
    G --> H{상태 양호?}
    H -->|Yes| I[Canary 50%]
    H -->|No| F
    I --> J[모니터링]
    J --> K{상태 양호?}
    K -->|Yes| L[전체 배포 100%]
    K -->|No| F
    F --> A
    L --> M[Canary 제거]
```

#### Canary 배포 설정

```yaml
# k8s/canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp-canary
  namespace: production
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 5
      - pause: {duration: 5m}
      - setWeight: 25
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 15m}
      - setWeight: 75
      - pause: {duration: 20m}
      canaryService: myapp-canary
      stableService: myapp-stable
      trafficRouting:
        istio:
          virtualService:
            name: myapp-vsvc
            routes:
            - primary
  selector:
    matchLabels:
      app: myapp
      version: canary
  template:
    metadata:
      labels:
        app: myapp
        version: canary
    spec:
      containers:
      - name: myapp
        image: myregistry.com/myapp:${CANARY_VERSION}
        ports:
        - containerPort: 8000
        env:
        - name: VERSION
          value: "${CANARY_VERSION}"
        - name: ENVIRONMENT
          value: "production"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### Rolling 배포

Rolling 배포는 점진적으로 인스턴스를 교체하는 가장 일반적인 배포 방식입니다.

#### Rolling 배포 설정

```yaml
# k8s/rolling-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-rolling
  namespace: production
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # 한 번에 추가되는 최대 Pod 수
      maxUnavailable: 0  # 한 번에 중지되는 최대 Pod 수
  selector:
    matchLabels:
      app: myapp
      deployment: rolling
  template:
    metadata:
      labels:
        app: myapp
        deployment: rolling
        version: v1.0.0
    spec:
      containers:
      - name: myapp
        image: myregistry.com/myapp:v1.0.0
        ports:
        - containerPort: 8000
        env:
        - name: ENVIRONMENT
          value: "production"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          successThreshold: 1
          failureThreshold: 3
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      terminationGracePeriodSeconds: 30
```

## 모니터링과 로깅

### Prometheus 모니터링 설정

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'myapp'
    static_configs:
      - targets: ['app:8000']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx:9113']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### 알림 규칙 설정

```yaml
# monitoring/alert_rules.yml
groups:
  - name: myapp_alerts
    rules:
      - alert: ApplicationDown
        expr: up{job="myapp"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Application is down"
          description: "Application {{ $labels.instance }} has been down for more than 1 minute."

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }} errors per second."

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }} seconds."

      - alert: DatabaseConnectionsHigh
        expr: pg_stat_activity_count > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High database connections"
          description: "Database has {{ $value }} active connections."
```

### Grafana 대시보드 설정

```json
{
  "dashboard": {
    "title": "MyApp Application Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      },
      {
        "title": "Application Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"myapp\"}",
            "legendFormat": "Health Status"
          }
        ]
      }
    ]
  }
}
```

### 구조화된 로깅

#### Python 애플리케이션 로깅 설정

```python
# src/logging_config.py
import logging
import logging.config
import json
from datetime import datetime

# 구조화된 로그 포매터
class StructuredFormatter(logging.Formatter):
    def format(self, record):
        log_entry = {
            "timestamp": datetime.utcnow().isoformat(),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno
        }

        # 추가 컨텍스트 정보
        if hasattr(record, 'request_id'):
            log_entry['request_id'] = record.request_id
        if hasattr(record, 'user_id'):
            log_entry['user_id'] = record.user_id
        if hasattr(record, 'spec_id'):
            log_entry['spec_id'] = record.spec_id

        # 예외 정보
        if record.exc_info:
            log_entry['exception'] = self.formatException(record.exc_info)

        return json.dumps(log_entry)

# 로깅 설정
LOGGING_CONFIG = {
    "version": 1,
    "disable_existing_loggers": False,
    "formatters": {
        "structured": {
            "()": StructuredFormatter
        },
        "detailed": {
            "format": "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        }
    },
    "handlers": {
        "console": {
            "class": "logging.StreamHandler",
            "level": "INFO",
            "formatter": "structured",
            "stream": "ext://sys.stdout"
        },
        "file": {
            "class": "logging.handlers.RotatingFileHandler",
            "level": "DEBUG",
            "formatter": "structured",
            "filename": "/var/log/app/app.log",
            "maxBytes": 10485760,  # 10MB
            "backupCount": 5
        }
    },
    "loggers": {
        "": {
            "level": "INFO",
            "handlers": ["console", "file"],
            "propagate": False
        },
        "uvicorn": {
            "level": "INFO",
            "handlers": ["console", "file"],
            "propagate": False
        },
        "sqlalchemy.engine": {
            "level": "WARNING",
            "handlers": ["file"],
            "propagate": False
        }
    }
}
```

## 보안 및 규제 준수

### 컨테이너 보안 설정

```dockerfile
# Dockerfile.security
FROM python:3.13-slim as security-base

# 비root 사용자 생성
RUN groupadd -r appuser && useradd -r -g appuser appuser

# 시스템 패키지 설치 (최소한)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Python 패키지 설치
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 애플리케이션 복사
COPY src/ /app/src/
COPY --chown=appuser:appuser . /app/

# 작업 디렉토리
WORKDIR /app

# 비root 사용자로 전환
USER appuser

# 보안 관련 환경 변수
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/app/.venv/bin:$PATH"

# 포트 (변경 가능)
EXPOSE 8000

# 헬스 체크
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# 애플리케이션 시작
CMD ["python", "-m", "uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### 네트워크 보안 설정

```yaml
# docker-compose.security.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.security
    networks:
      - frontend
      - backend
    user: "1000:1000"
    read_only: true
    tmpfs:
      - /tmp
      - /var/log/app
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - SETUID

  nginx:
    image: nginx:alpine
    networks:
      - frontend
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - ./static:/var/www/static:ro
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - NET_BIND_SERVICE

  db:
    image: postgres:15
    networks:
      - backend
    environment:
      - POSTGRES_DB_FILE=/run/secrets/postgres_db
      - POSTGRES_USER_FILE=/run/secrets/postgres_user
      - POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
    secrets:
      - postgres_db
      - postgres_user
      - postgres_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    security_opt:
      - no-new-privileges:true

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true

secrets:
  postgres_db:
    file: ./secrets/postgres_db.txt
  postgres_user:
    file: ./secrets/postgres_user.txt
  postgres_password:
    file: ./secrets/postgres_password.txt

volumes:
  postgres_data:
    driver: local
```

### OWASP 보안 검사

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * 1'  # 매주 월요일 새벽 2시

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.13'

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install bandit safety safety-db

    - name: Run Bandit security scan
      run: |
        bandit -r src/ -f json -o bandit-report.json
        bandit -r src/

    - name: Check dependencies for known vulnerabilities
      run: |
        safety check --json --output safety-report.json
        safety check

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: OWASP ZAP Baseline Scan
      uses: zaproxy/action-baseline@v0.7.0
      with:
        target: 'http://localhost:8000'
        rules_file_name: '.zap/rules.tsv'
        cmd_options: '-a'

    - name: Upload security scan results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'

    - name: Comment security findings
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');

          if (fs.existsSync('bandit-report.json')) {
            const bandit = JSON.parse(fs.readFileSync('bandit-report.json', 'utf8'));
            if (bandit.results.length > 0) {
              const comment = `## <span class="material-icons">lock</span> Security Scan Results\n\n**Bandit found ${bandit.results.length} issues:**\n\n${bandit.results.map(issue => `- **${issue.test_name}**: ${issue.issue_text} (${issue.filename}:${issue.line_number})`).join('\n')}`;
              github.rest.issues.createComment({
                issue_number: context.issue.number,
                owner: context.repo.owner,
                repo: context.repo.repo,
                body: comment
              });
            }
          }
```

## 배포 자동화

### 배포 파이프라인 오케스트레이션

```bash
#!/bin/bash
# scripts/deploy-pipeline.sh

set -e

# 인자 처리
ENVIRONMENT=${1:-staging}
VERSION=${2:-latest}
STRATEGY=${3:-rolling}

echo "<span class="material-icons">rocket_launch</span> Starting deployment pipeline"
echo "📋 Environment: $ENVIRONMENT"
echo "<span class="material-icons">label</span>  Version: $VERSION"
echo "🔄 Strategy: $STRATEGY"

# 1. 사전 검사
echo "<span class="material-icons">search</span> Running pre-deployment checks..."
./scripts/pre-deployment-checks.sh $ENVIRONMENT

# 2. 백업 생성
echo "💾 Creating backups..."
./scripts/backup-system.sh $ENVIRONMENT

# 3. 테스트 환경 배포
if [ "$ENVIRONMENT" != "production" ]; then
    echo "🧪 Deploying to test environment first..."
    ./scripts/deploy-to-test.sh $VERSION

    echo "🧪 Running integration tests..."
    ./scripts/integration-tests.sh test
fi

# 4. 대상 환경 배포
echo "<span class="material-icons">rocket_launch</span> Deploying to $ENVIRONMENT environment..."

case $STRATEGY in
    "rolling")
        ./scripts/rolling-deploy.sh $ENVIRONMENT $VERSION
        ;;
    "blue-green")
        ./scripts/blue-green-deploy.sh $ENVIRONMENT $VERSION
        ;;
    "canary")
        ./scripts/canary-deploy.sh $ENVIRONMENT $VERSION
        ;;
    *)
        echo "<span class="material-icons">cancel</span> Unknown deployment strategy: $STRATEGY"
        exit 1
        ;;
esac

# 5. 배포 후 검증
echo "<span class="material-icons">search</span> Running post-deployment validation..."
./scripts/post-deployment-checks.sh $ENVIRONMENT

# 6. 성능 테스트
if [ "$ENVIRONMENT" = "staging" ]; then
    echo "<span class="material-icons">analytics</span> Running performance tests..."
    ./scripts/performance-tests.sh $ENVIRONMENT
fi

# 7. 보안 검사
echo "<span class="material-icons">lock</span> Running security validation..."
./scripts/security-validation.sh $ENVIRONMENT

# 8. 알림 발송
echo "📧 Sending deployment notifications..."
./scripts/notify-deployment.sh $ENVIRONMENT $VERSION success

echo "<span class="material-icons">check_circle</span> Deployment pipeline completed successfully!"
```

### 멀티 환경 배포 관리

```yaml
# deploy-config.yml
environments:
  development:
    type: "compose"
    config_file: "docker-compose.dev.yml"
    auto_deploy: true
    health_check_timeout: 60
    rollback_on_failure: false

  staging:
    type: "kubernetes"
    namespace: "staging"
    config_file: "k8s/staging/"
    auto_deploy: true
    health_check_timeout: 300
    rollback_on_failure: true
    tests:
      - smoke
      - integration
      - performance

  production:
    type: "kubernetes"
    namespace: "production"
    config_file: "k8s/production/"
    auto_deploy: false
    health_check_timeout: 600
    rollback_on_failure: true
    approvals_required: 2
    tests:
      - smoke
      - integration
      - security
      - performance

strategies:
  rolling:
    default: true
    max_unavailable: 0
    max_surge: 1

  blue_green:
    default: false
    switch_traffic_automatically: true
    health_check_duration: 300

  canary:
    default: false
    steps:
      - weight: 5
        duration: 300
      - weight: 25
        duration: 600
      - weight: 50
        duration: 900
      - weight: 100
        duration: 0

notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK}"
    channels:
      development: "#dev-deployments"
      staging: "#staging-deployments"
      production: "#prod-deployments"

  email:
    smtp_server: "${SMTP_SERVER}"
    recipients:
      development: ["dev-team@company.com"]
      staging: ["qa-team@company.com"]
      production: ["ops-team@company.com", "management@company.com"]
```

### 배포 상태 추적

```python
# scripts/deployment_tracker.py
import json
import time
import requests
from datetime import datetime
from typing import Dict, List, Optional

class DeploymentTracker:
    def __init__(self, config_file: str = "deploy-config.yml"):
        self.config = self.load_config(config_file)

    def load_config(self, config_file: str) -> Dict:
        """배포 설정 로드"""
        with open(config_file, 'r') as f:
            return yaml.safe_load(f)

    def start_deployment(self, environment: str, version: str, strategy: str) -> str:
        """배포 시작 추적"""
        deployment_id = f"{environment}-{version}-{int(time.time())}"

        deployment = {
            "id": deployment_id,
            "environment": environment,
            "version": version,
            "strategy": strategy,
            "status": "in_progress",
            "started_at": datetime.utcnow().isoformat(),
            "steps": []
        }

        self.save_deployment(deployment)
        return deployment_id

    def add_step(self, deployment_id: str, step_name: str, status: str,
                  details: Optional[Dict] = None):
        """배포 단계 추가"""
        deployment = self.get_deployment(deployment_id)

        step = {
            "name": step_name,
            "status": status,
            "timestamp": datetime.utcnow().isoformat(),
            "details": details or {}
        }

        deployment["steps"].append(step)
        self.save_deployment(deployment)

    def complete_deployment(self, deployment_id: str, status: str):
        """배포 완료"""
        deployment = self.get_deployment(deployment_id)
        deployment["status"] = status
        deployment["completed_at"] = datetime.utcnow().isoformat()

        # 성공 메트릭 계산
        if status == "success":
            deployment["duration"] = self.calculate_duration(deployment)
            deployment["success_rate"] = self.calculate_success_rate(deployment)

        self.save_deployment(deployment)

        # 알림 발송
        self.send_notification(deployment)

    def get_deployment(self, deployment_id: str) -> Dict:
        """배포 정보 조회"""
        deployments_file = "deployments.json"

        if not os.path.exists(deployments_file):
            return {}

        with open(deployments_file, 'r') as f:
            deployments = json.load(f)

        return deployments.get(deployment_id, {})

    def save_deployment(self, deployment: Dict):
        """배포 정보 저장"""
        deployments_file = "deployments.json"

        deployments = {}
        if os.path.exists(deployments_file):
            with open(deployments_file, 'r') as f:
                deployments = json.load(f)

        deployments[deployment["id"]] = deployment

        with open(deployments_file, 'w') as f:
            json.dump(deployments, f, indent=2)

    def send_notification(self, deployment: Dict):
        """배포 알림 발송"""
        message = self.format_notification_message(deployment)

        # Slack 알림
        if "slack" in self.config.get("notifications", {}):
            self.send_slack_notification(message)

        # Email 알림
        if "email" in self.config.get("notifications", {}):
            self.send_email_notification(deployment, message)

    def format_notification_message(self, deployment: Dict) -> str:
        """알림 메시지 포맷팅"""
        status_emoji = "<span class="material-icons">check_circle</span>" if deployment["status"] == "success" else "<span class="material-icons">cancel</span>"

        message = f"""
{status_emoji} Deployment {deployment['status'].upper()}

📋 Environment: {deployment['environment']}
<span class="material-icons">label</span>  Version: {deployment['version']}
🔄 Strategy: {deployment['strategy']}
⏰ Duration: {deployment.get('duration', 'N/A')}
<span class="material-icons">analytics</span> Success Rate: {deployment.get('success_rate', 'N/A')}%

Steps:
"""

        for step in deployment["steps"]:
            step_emoji = "<span class="material-icons">check_circle</span>" if step["status"] == "success" else "<span class="material-icons">cancel</span>"
            message += f"{step_emoji} {step['name']}\n"

        return message
```

## 성능 최적화

### 애플리케이션 성능 최적화

```python
# src/performance.py
import asyncio
import time
from functools import wraps
from typing import Callable, Any

def performance_monitor(func: Callable) -> Callable:
    """성능 모니터링 데코레이터"""
    @wraps(func)
    async def async_wrapper(*args, **kwargs):
        start_time = time.time()
        try:
            result = await func(*args, **kwargs)
            execution_time = time.time() - start_time

            # 성능 메트릭 기록
            record_performance_metric(func.__name__, execution_time, "success")

            return result
        except Exception as e:
            execution_time = time.time() - start_time
            record_performance_metric(func.__name__, execution_time, "error")
            raise

    @wraps(func)
    def sync_wrapper(*args, **kwargs):
        start_time = time.time()
        try:
            result = func(*args, **kwargs)
            execution_time = time.time() - start_time

            # 성능 메트릭 기록
            record_performance_metric(func.__name__, execution_time, "success")

            return result
        except Exception as e:
            execution_time = time.time() - start_time
            record_performance_metric(func.__name__, execution_time, "error")
            raise

    return async_wrapper if asyncio.iscoroutinefunction(func) else sync_wrapper

def record_performance_metric(function_name: str, execution_time: float, status: str):
    """성능 메트릭 기록"""
    # Prometheus 메트릭 기록
    from prometheus_client import Histogram, Counter

    REQUEST_DURATION = Histogram(
        'request_duration_seconds',
        'Request duration',
        ['function_name', 'status']
    )

    REQUEST_COUNT = Counter(
        'request_count_total',
        'Total requests',
        ['function_name', 'status']
    )

    REQUEST_DURATION.labels(function_name=function_name, status=status).observe(execution_time)
    REQUEST_COUNT.labels(function_name=function_name, status=status).inc()
```

### 데이터베이스 성능 최적화

```sql
-- database/performance-optimizations.sql

-- 인덱스 최적화
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);
CREATE INDEX CONCURRENTLY idx_users_created_at ON users(created_at);
CREATE INDEX CONCURRENTLY idx_posts_author_id ON posts(author_id);

-- 파티셔닝 (대용량 테이블)
CREATE TABLE logs (
    id SERIAL,
    message TEXT,
    level VARCHAR(10),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 월별 파티션
CREATE TABLE logs_2024_01 PARTITION OF logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 쿼리 최적화
EXPLAIN (ANALYZE, BUFFERS)
SELECT u.id, u.name, COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.author_id
WHERE u.created_at >= '2024-01-01'
GROUP BY u.id, u.name
ORDER BY post_count DESC;

-- 커넥션 풀링 설정
-- postgresql.conf
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
```

### 캐싱 전략

```python
# src/cache.py
import redis
import json
import pickle
from typing import Any, Optional, Union
from functools import wraps

class CacheManager:
    def __init__(self, redis_url: str):
        self.redis_client = redis.from_url(redis_url)

    def cache_result(self, key_prefix: str, ttl: int = 3600):
        """결과 캐싱 데코레이터"""
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # 캐시 키 생성
                cache_key = f"{key_prefix}:{hash(str(args) + str(kwargs))}"

                # 캐시 조회
                cached_result = self.get(cache_key)
                if cached_result is not None:
                    return cached_result

                # 함수 실행
                result = await func(*args, **kwargs)

                # 결과 캐싱
                self.set(cache_key, result, ttl)

                return result
            return wrapper
        return decorator

    def get(self, key: str) -> Optional[Any]:
        """캐시에서 값 조회"""
        try:
            value = self.redis_client.get(key)
            if value:
                return pickle.loads(value)
        except Exception as e:
            print(f"Cache get error: {e}")
        return None

    def set(self, key: str, value: Any, ttl: int = 3600):
        """캐시에 값 저장"""
        try:
            serialized_value = pickle.dumps(value)
            self.redis_client.setex(key, ttl, serialized_value)
        except Exception as e:
            print(f"Cache set error: {e}")

    def delete(self, key: str):
        """캐시에서 값 삭제"""
        try:
            self.redis_client.delete(key)
        except Exception as e:
            print(f"Cache delete error: {e}")

    def clear_pattern(self, pattern: str):
        """패턴으로 캐시 정리"""
        try:
            keys = self.redis_client.keys(pattern)
            if keys:
                self.redis_client.delete(*keys)
        except Exception as e:
            print(f"Cache clear error: {e}")

# 사용 예시
cache_manager = CacheManager("redis://localhost:6379/0")

@cache_manager.cache_result("user_profile", ttl=1800)
async def get_user_profile(user_id: int):
    """사용자 프로필 조회 (30분 캐시)"""
    # 데이터베이스 조회 로직
    pass

@cache_manager.cache_result("popular_posts", ttl=3600)
async def get_popular_posts(limit: int = 10):
    """인기 포스트 조회 (1시간 캐시)"""
    # 복잡한 쿼리 실행
    pass
```

## 문제 해결 및 복구

### 롤백 전략

```bash
#!/bin/bash
# scripts/rollback.sh

set -e

ENVIRONMENT=${1:-staging}
TARGET_VERSION=${2:-previous}

echo "🔄 Starting rollback for $ENVIRONMENT environment"
echo "<span class="material-icons">target</span> Target version: $TARGET_VERSION"

# 현재 상태 백업
echo "💾 Backing up current state..."
./scripts/backup-current-state.sh $ENVIRONMENT

# 롤백 방식 선택
case $TARGET_VERSION in
    "previous")
        VERSION=$(git describe --tags --abbrev=0 HEAD~1)
        ;;
    "stable")
        VERSION=$(git tag --sort=-version:refname | head -n 1)
        ;;
    *)
        VERSION=$TARGET_VERSION
        ;;
esac

echo "<span class="material-icons">label</span>  Rolling back to version: $VERSION"

# 롤백 실행
echo "🔄 Executing rollback..."

# 1. 배포 중지
echo "⏸️  Stopping current deployment..."
./scripts/stop-deployment.sh $ENVIRONMENT

# 2. 이전 버전 배포
echo "<span class="material-icons">rocket_launch</span> Deploying previous version..."
./scripts/deploy-version.sh $ENVIRONMENT $VERSION

# 3. 상태 검증
echo "<span class="material-icons">search</span> Validating rollback..."
./scripts/validate-rollback.sh $ENVIRONMENT $VERSION

# 4. 데이터베이스 마이그레이션 (필요시)
echo "<span class="material-icons">analytics</span> Running database migrations..."
./scripts/migrate-database.sh $ENVIRONMENT rollback

# 5. 롤백 완료 확인
echo "<span class="material-icons">check_circle</span> Rollback completed successfully!"
echo "<span class="material-icons">language</span> Application running on: https://api.myapp.com"
echo "<span class="material-icons">label</span>  Current version: $VERSION"

# 알림 발송
./scripts/notify-rollback.sh $ENVIRONMENT $VERSION
```

### 장애 대응 절차

```yaml
# incident-response.yml
incident_response:
  severity_levels:
    critical:
      response_time: 15  # 분
      escalation_time: 30  # 분
      notification_channels: ["slack", "email", "sms", "call"]

    high:
      response_time: 30
      escalation_time: 60
      notification_channels: ["slack", "email"]

    medium:
      response_time: 60
      escalation_time: 120
      notification_channels: ["slack"]

    low:
      response_time: 240
      escalation_time: 480
      notification_channels: ["slack"]

  playbooks:
    application_down:
      detection:
        - health_check_failure
        - error_rate_threshold
        - user_reports

      immediate_actions:
        - check_logs
        - verify_system_resources
        - check_external_dependencies

      recovery_actions:
        - restart_services
        - rollback_deployment
        - scale_up_resources

      post_incident:
        - root_cause_analysis
        - prevention_measures
        - documentation_update

    performance_degradation:
      detection:
        - response_time_threshold
        - cpu_memory_threshold
        - database_slow_queries

      immediate_actions:
        - check_system_metrics
        - identify_bottlenecks
        - scale_resources

      recovery_actions:
        - optimize_queries
        - add_caching
        - scale_horizontally

    security_incident:
      detection:
        - unusual_access_patterns
        - security_scan_alerts
        - data_breach_indicators

      immediate_actions:
        - isolate_affected_systems
        - preserve_evidence
        - notify_security_team

      recovery_actions:
        - patch_vulnerabilities
        - reset_credentials
        - restore_from_backup

  communication_templates:
    initial_alert: |
      🚨 **Incident Alert**

      **Severity**: {severity}
      **Service**: {service}
      **Impact**: {impact}
      **Started**: {timestamp}

      **Next Update**: {next_update_time}

      **Response Team**: {team}

    progress_update: |
      📋 **Incident Update**

      **Incident ID**: {incident_id}
      **Status**: {status}
      **Actions Taken**: {actions}
      **ETA**: {eta}

      **Next Update**: {next_update_time}

    resolution: |
      <span class="material-icons">check_circle</span> **Incident Resolved**

      **Incident ID**: {incident_id}
      **Resolution Time**: {resolution_time}
      **Root Cause**: {root_cause}
      **Prevention Measures**: {prevention}

      **Post-Mortem**: {post_mortem_link}
```

## 모범 사례

### 배포 체크리스트

#### 사전 배포 확인사항

```markdown
## 📋 Pre-Deployment Checklist

### <span class="material-icons">check_circle</span> 코드 준비
- [ ] 모든 테스트 통과 (단위, 통합, E2E)
- [ ] 코드 리뷰 완료 및 승인
- [ ] 보안 스캔 통과
- [ ] 성능 테스트 통과
- [ ] 문서 업데이트 완료

### <span class="material-icons">check_circle</span> 환경 준비
- [ ] 타겟 환경 상태 정상
- [ ] 데이터베이스 백업 완료
- [ ] 설정 파일 검증 완료
- [ ] 리소스 용량 확인
- [ ] 네트워크 연결 확인

### <span class="material-icons">check_circle</span> 배포 계획
- [ ] 롤백 계획 수립
- [ ] 배포 시간 창 확보
- [ ] 관련팀 통지 완료
- [ ] 모니터링 시스템 준비
- [ ] 알림 채널 설정

### <span class="material-icons">check_circle</span> 보안 및 규제
- [ ] 접근 권한 확인
- [ ] 보안 정책 준수
- [ ] 규제 요구사항 검증
- [ ] 감사 로그 활성화
- [ ] 민감정보 처리 확인
```

#### 배포 후 확인사항

```markdown
## <span class="material-icons">check_circle</span> Post-Deployment Checklist

### <span class="material-icons">rocket_launch</span> 배포 상태 확인
- [ ] 애플리케이션 정상 시작
- [ ] 헬스 체크 통과
- [ ] 로그 에러 없음
- [ ] 데이터베이스 연결 정상
- [ ] 외부 API 연동 정상

### <span class="material-icons">analytics</span> 기능 검증
- [ ] 주요 기능 동작 확인
- [ ] API 엔드포인트 응답 정상
- [ ] 사용자 인증 작동
- [ ] 데이터 CRUD 작업 정상
- [ ] 파일 업로드/다운로드 정상

### <span class="material-icons">search</span> 성능 모니터링
- [ ] 응답 시간 정상 범위
- [ ] CPU/메모리 사용량 정상
- [ ] 데이터베이스 쿼리 성능 양호
- [ ] 캐시 적중률 정상
- [ ] 에러률 임계치 이하

### <span class="material-icons">trending_up</span> 모니터링 알림
- [ ] Prometheus 메트릭 수집
- [ ] Grafana 대시보드 정상
- [ ] 로그 수집 및 분석
- [ ] 알림 규칙 동작 확인
- [ ] SLA 지표 충족

### 🔄 롤백 준비
- [ ] 이전 버전 이미지 보관
- [ ] 롤백 스크립트 검증
- [ ] 데이터베이스 마이그레이션 스크립트 준비
- [ ] 롤백 시나리오 테스트 완료
- [ ] 긴급 연락망 확인
```

### 배포 성공 지표

```yaml
# deployment-metrics.yml
success_metrics:
  deployment_success_rate:
    target: 95%
    measurement: "Successful deployments / Total deployments"

  deployment_time:
    target: 15  # minutes
    measurement: "Time from commit to production"

  rollback_rate:
    target: 5%
    measurement: "Rollbacks / Total deployments"

  downtime_per_deployment:
    target: 30  # seconds
    measurement: "Service unavailability during deployment"

  post_deployment_issues:
    target: 2
    measurement: "Critical issues in first 24 hours"

  performance_regression:
    target: 5%
    measurement: "Performance degradation after deployment"

monitoring:
  real_time_alerts:
    - deployment_failure
    - health_check_failure
    - error_rate_spike
    - response_time_degradation

  daily_reports:
    - deployment_summary
    - performance_metrics
    - error_analysis
    - user_impact_assessment

  weekly_reviews:
    - deployment_process_review
    - incident_postmortem
    - optimization_opportunities
    - team_feedback
```

---

이 가이드를 통해 MoAI-ADK 프로젝트를 안전하고 효율적으로 배포할 수 있습니다. 현대적인 DevOps 관행과 자동화된 워크플로우를 통해 안정적인 서비스 운영을 달성할 수 있을 것입니다.