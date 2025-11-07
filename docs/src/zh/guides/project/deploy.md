# 完整部署指南

一份全面的指南，涵盖 MoAI-ADK 项目的部署流程。本文档详细介绍了从 Docker 容器部署到云服务和 CI/CD 流水线构建的所有内容，解释了各种部署环境和现代 DevOps 实践。

## 目录

01. [概述](#概述)
02. [部署架构](#部署架构)
03. [Docker 容器部署](#docker-容器部署)
04. [云平台部署](#云平台部署)
05. [CI/CD 流水线构建](#cicd-流水线构建)
06. [环境特定部署策略](#环境特定部署策略)
07. [发布策略](#发布策略)
08. [监控和日志](#监控和日志)
09. [安全与合规](#安全与合规)
10. [部署自动化](#部署自动化)
11. [性能优化](#性能优化)
12. [故障排除和恢复](#故障排除和恢复)
13. [最佳实践](#最佳实践)

## 概述

部署 MoAI-ADK 项目不仅仅是将代码上传到服务器——它是构建一个稳定、可扩展和安全的系统。Alfred 的 SPEC-first 开发方法确保了整个部署过程的一致性和可追溯性。

### 部署系统的核心原则

1. **自动化优先**：所有部署流程必须自动化
2. **零停机时间**：用户不会遇到服务中断
3. **随时可回滚**：必须能够随时回滚到以前的版本
4. **可观测性**：实时查看部署状态和系统健康状况
5. **安全性**：整个部署过程必须符合安全规定

### MoAI-ADK 部署特性

- **SPEC 驱动部署**：基于 SPEC 文档的部署规划
- **基于 TAG 的跟踪**：所有已部署组件的 TAG 跟踪
- **自动化质量门禁**：部署前自动验证质量
- **渐进式部署**：支持逐步部署
- **环境一致性**：确保所有环境之间的一致性

## 部署架构

### 现代部署架构

```mermaid
graph TD
    A[开发人员] --> B[Git Push]
    B --> C[CI/CD 流水线]
    C --> D[构建与测试]
    D --> E[安全扫描]
    E --> F[容器构建]
    F --> G[制品仓库]
    G --> H[部署策略]
    H --> I[预发布环境]
    I --> J[自动化测试]
    J --> K{批准？}
    K -->|是| L[生产部署]
    K -->|否| M[回滚]
    L --> N[健康检查]
    N --> O[监控]
    O --> P[告警]
```

### MoAI-ADK 部署工作流

```mermaid
sequenceDiagram
    participant Dev as 开发人员
    participant Git as Git 仓库
    participant CI as CI/CD 流水线
    participant QA as 质量门禁
    participant Deploy as 部署系统
    participant Prod as 生产环境

    Dev->>Git: /alfred:3-sync
    Git->>CI: 触发流水线
    CI->>CI: 构建应用
    CI->>CI: 运行测试（TRUST 验证）
    CI->>QA: 质量门禁检查
    QA->>CI: 通过/失败
    CI->>Deploy: 创建 Docker 镜像
    Deploy->>Deploy: 部署到预发布
    Deploy->>QA: 自动化 E2E 测试
    QA->>Deploy: 批准
    Deploy->>Prod: 蓝绿部署
    Prod->>Deploy: 健康检查
    Deploy->>Dev: 部署报告
```

## Docker 容器部署

### 基本 Dockerfile 配置

为 MoAI-ADK 项目优化的 Dockerfile：

```dockerfile
# 多阶段构建以优化生产
FROM python:3.13-slim as builder

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 Python 依赖
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv && \
    uv pip install --system --no-cache -r <(uv pip compile pyproject.toml)

# 生产阶段
FROM python:3.13-slim

WORKDIR /app

# 仅安装运行时依赖
RUN apt-get update && apt-get install -y \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

# 从构建器复制已安装的包
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# 复制应用代码
COPY . .

# 创建非 root 用户
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# 暴露端口
EXPOSE 8000

# 启动应用
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose 配置

#### 开发环境

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - /app/.venv  # 排除虚拟环境
    ports:
      - "8000:8000"
    environment:
      - ENVIRONMENT=development
      - DEBUG=true
      - DATABASE_URL=postgresql://dev:devpass@db:5432/moai_dev
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - db
      - redis
    command: uvicorn main:app --reload --host 0.0.0.0 --port 8000

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=dev
      - POSTGRES_PASSWORD=devpass
      - POSTGRES_DB=moai_dev
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  mailhog:
    image: mailhog/mailhog:latest
    ports:
      - "1025:1025"  # SMTP
      - "8025:8025"  # Web UI

volumes:
  postgres_data:
  redis_data:
```

#### 预发布环境

```yaml
# docker-compose.staging.yml
version: '3.8'

services:
  app:
    image: ${DOCKER_REGISTRY}/moai-adk:${VERSION}
    ports:
      - "8000:8000"
    environment:
      - ENVIRONMENT=staging
      - DEBUG=false
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SECRET_KEY=${SECRET_KEY}
    depends_on:
      - db
      - redis
    deploy:
      replicas: 2
      restart_policy:
        condition: on-failure
        max_attempts: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME}
    volumes:
      - postgres_staging:/var/lib/postgresql/data
    deploy:
      placement:
        constraints:
          - node.role == manager

  redis:
    image: redis:7-alpine
    volumes:
      - redis_staging:/data
    command: redis-server --appendonly yes

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - app

volumes:
  postgres_staging:
  redis_staging:
```

#### 生产环境

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  app:
    image: ${DOCKER_REGISTRY}/moai-adk:${VERSION}
    environment:
      - ENVIRONMENT=production
      - DEBUG=false
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - SECRET_KEY=${SECRET_KEY}
      - SENTRY_DSN=${SENTRY_DSN}
    deploy:
      replicas: 5
      update_config:
        parallelism: 2
        delay: 10s
        failure_action: rollback
        monitor: 30s
      restart_policy:
        condition: on-failure
        max_attempts: 5
        window: 120s
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME}
    volumes:
      - postgres_prod:/var/lib/postgresql/data
    deploy:
      placement:
        constraints:
          - node.labels.type == database
      resources:
        limits:
          cpus: '4.0'
          memory: 4G

  redis:
    image: redis:7-alpine
    volumes:
      - redis_prod:/data
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    deploy:
      placement:
        constraints:
          - node.labels.type == cache

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.prod.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - nginx_cache:/var/cache/nginx
    depends_on:
      - app
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

volumes:
  postgres_prod:
  redis_prod:
  nginx_cache:
```

### Docker 构建和推送

```bash
#!/bin/bash
# scripts/docker-build.sh

set -euo pipefail

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-docker.io/myorg}
IMAGE_NAME="moai-adk"

echo "构建 Docker 镜像：${REGISTRY}/${IMAGE_NAME}:${VERSION}"

# 构建多平台镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
  --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
  --push \
  .

echo "镜像推送成功"

# 扫描漏洞
docker scout cves "${REGISTRY}/${IMAGE_NAME}:${VERSION}"
```

## Kubernetes 部署

### 基本 Kubernetes 配置

#### 命名空间

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: moai-adk
  labels:
    name: moai-adk
    environment: production
```

#### ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: moai-adk-config
  namespace: moai-adk
data:
  ENVIRONMENT: "production"
  DEBUG: "false"
  LOG_LEVEL: "INFO"
  DATABASE_HOST: "postgres-service"
  DATABASE_PORT: "5432"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
```

#### Secret

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: moai-adk-secret
  namespace: moai-adk
type: Opaque
stringData:
  DATABASE_URL: "postgresql://user:password@postgres-service:5432/moai_prod"
  SECRET_KEY: "your-secret-key-here"
  REDIS_PASSWORD: "your-redis-password"
  SENTRY_DSN: "https://your-sentry-dsn"
```

#### Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: moai-adk
  namespace: moai-adk
  labels:
    app: moai-adk
    version: v1.0.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: moai-adk
  template:
    metadata:
      labels:
        app: moai-adk
        version: v1.0.0
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8000"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: moai-adk
      containers:
      - name: moai-adk
        image: docker.io/myorg/moai-adk:v1.0.0
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
          name: http
          protocol: TCP
        envFrom:
        - configMapRef:
            name: moai-adk-config
        - secretRef:
            name: moai-adk-secret
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
      terminationGracePeriodSeconds: 30
```

#### Service

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: moai-adk-service
  namespace: moai-adk
  labels:
    app: moai-adk
spec:
  type: ClusterIP
  selector:
    app: moai-adk
  ports:
  - port: 80
    targetPort: 8000
    protocol: TCP
    name: http
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
```

#### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: moai-adk-ingress
  namespace: moai-adk
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/limit-connections: "10"
spec:
  tls:
  - hosts:
    - api.moai-adk.example.com
    secretName: moai-adk-tls
  rules:
  - host: api.moai-adk.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: moai-adk-service
            port:
              number: 80
```

#### HorizontalPodAutoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: moai-adk-hpa
  namespace: moai-adk
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: moai-adk
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
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Max
```

## 云平台部署

### AWS ECS Fargate

#### 任务定义

```json
{
  "family": "moai-adk",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "moai-adk",
      "image": "ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/moai-adk:VERSION",
      "portMappings": [
        {
          "containerPort": 8000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENVIRONMENT",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:moai-adk/database-url"
        },
        {
          "name": "SECRET_KEY",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:moai-adk/secret-key"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/moai-adk",
          "awslogs-region": "us-east-1",
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

#### CloudFormation 模板

```yaml
# cloudformation/moai-adk.yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'MoAI-ADK ECS Fargate 部署'

Parameters:
  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC ID
  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: ECS 任务的子网 ID
  ImageVersion:
    Type: String
    Default: latest
    Description: Docker 镜像版本

Resources:
  ECSCluster:
    Type: AWS::ECS::Cluster
    Properties:
      ClusterName: moai-adk-cluster
      CapacityProviders:
        - FARGATE
        - FARGATE_SPOT
      DefaultCapacityProviderStrategy:
        - CapacityProvider: FARGATE
          Weight: 1
        - CapacityProvider: FARGATE_SPOT
          Weight: 3

  TaskDefinition:
    Type: AWS::ECS::TaskDefinition
    Properties:
      Family: moai-adk
      NetworkMode: awsvpc
      RequiresCompatibilities:
        - FARGATE
      Cpu: '1024'
      Memory: '2048'
      ExecutionRoleArn: !GetAtt ECSTaskExecutionRole.Arn
      TaskRoleArn: !GetAtt ECSTaskRole.Arn
      ContainerDefinitions:
        - Name: moai-adk
          Image: !Sub '${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/moai-adk:${ImageVersion}'
          PortMappings:
            - ContainerPort: 8000
              Protocol: tcp
          LogConfiguration:
            LogDriver: awslogs
            Options:
              awslogs-group: !Ref LogGroup
              awslogs-region: !Ref AWS::Region
              awslogs-stream-prefix: ecs

  Service:
    Type: AWS::ECS::Service
    DependsOn: LoadBalancerListener
    Properties:
      ServiceName: moai-adk-service
      Cluster: !Ref ECSCluster
      TaskDefinition: !Ref TaskDefinition
      DesiredCount: 3
      LaunchType: FARGATE
      NetworkConfiguration:
        AwsvpcConfiguration:
          AssignPublicIp: DISABLED
          Subnets: !Ref SubnetIds
          SecurityGroups:
            - !Ref ServiceSecurityGroup
      LoadBalancers:
        - ContainerName: moai-adk
          ContainerPort: 8000
          TargetGroupArn: !Ref TargetGroup
      HealthCheckGracePeriodSeconds: 60

  LoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: moai-adk-alb
      Type: application
      Scheme: internet-facing
      Subnets: !Ref SubnetIds
      SecurityGroups:
        - !Ref LoadBalancerSecurityGroup

  TargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: moai-adk-tg
      Port: 8000
      Protocol: HTTP
      VpcId: !Ref VpcId
      TargetType: ip
      HealthCheckPath: /health
      HealthCheckIntervalSeconds: 30
      HealthCheckTimeoutSeconds: 5
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3

  LoadBalancerListener:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      LoadBalancerArn: !Ref LoadBalancer
      Port: 443
      Protocol: HTTPS
      Certificates:
        - CertificateArn: !Ref Certificate
      DefaultActions:
        - Type: forward
          TargetGroupArn: !Ref TargetGroup

  AutoScalingTarget:
    Type: AWS::ApplicationAutoScaling::ScalableTarget
    Properties:
      MaxCapacity: 10
      MinCapacity: 3
      ResourceId: !Sub 'service/${ECSCluster}/${Service.Name}'
      RoleARN: !Sub 'arn:aws:iam::${AWS::AccountId}:role/aws-service-role/ecs.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_ECSService'
      ScalableDimension: ecs:service:DesiredCount
      ServiceNamespace: ecs

  AutoScalingPolicy:
    Type: AWS::ApplicationAutoScaling::ScalingPolicy
    Properties:
      PolicyName: moai-adk-scaling-policy
      PolicyType: TargetTrackingScaling
      ScalingTargetId: !Ref AutoScalingTarget
      TargetTrackingScalingPolicyConfiguration:
        PredefinedMetricSpecification:
          PredefinedMetricType: ECSServiceAverageCPUUtilization
        TargetValue: 70.0
        ScaleInCooldown: 300
        ScaleOutCooldown: 60

Outputs:
  LoadBalancerURL:
    Description: 负载均衡器的 URL
    Value: !GetAtt LoadBalancer.DNSName
```

### Google Cloud Run

#### 服务配置

```yaml
# gcloud/service.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: moai-adk
  namespace: default
  annotations:
    run.googleapis.com/ingress: all
    run.googleapis.com/execution-environment: gen2
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: '1'
        autoscaling.knative.dev/maxScale: '100'
        run.googleapis.com/cpu-throttling: 'false'
        run.googleapis.com/startup-cpu-boost: 'true'
    spec:
      containerConcurrency: 80
      timeoutSeconds: 300
      containers:
      - image: gcr.io/PROJECT_ID/moai-adk:VERSION
        ports:
        - name: http1
          containerPort: 8000
        env:
        - name: ENVIRONMENT
          value: production
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-url
              key: latest
        resources:
          limits:
            cpu: '2000m'
            memory: '2Gi'
        startupProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 0
          periodSeconds: 10
```

#### Terraform 配置

```hcl
# terraform/gcp/main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_cloud_run_service" "moai_adk" {
  name     = "moai-adk"
  location = var.region

  template {
    spec {
      containers {
        image = "gcr.io/${var.project_id}/moai-adk:${var.image_version}"

        ports {
          container_port = 8000
        }

        env {
          name  = "ENVIRONMENT"
          value = "production"
        }

        env {
          name = "DATABASE_URL"
          value_from {
            secret_key_ref {
              name = google_secret_manager_secret.database_url.secret_id
              key  = "latest"
            }
          }
        }

        resources {
          limits = {
            cpu    = "2000m"
            memory = "2Gi"
          }
        }
      }

      container_concurrency = 80
      timeout_seconds      = 300
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/minScale"         = "1"
        "autoscaling.knative.dev/maxScale"         = "100"
        "run.googleapis.com/cpu-throttling"        = "false"
        "run.googleapis.com/startup-cpu-boost"     = "true"
        "run.googleapis.com/execution-environment" = "gen2"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  autogenerate_revision_name = true
}

resource "google_cloud_run_service_iam_member" "public" {
  service  = google_cloud_run_service.moai_adk.name
  location = google_cloud_run_service.moai_adk.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_secret_manager_secret" "database_url" {
  secret_id = "database-url"

  replication {
    automatic = true
  }
}

output "service_url" {
  value = google_cloud_run_service.moai_adk.status[0].url
}
```

### Azure 容器实例

#### ARM 模板

```json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "containerGroupName": {
      "type": "string",
      "defaultValue": "moai-adk-group",
      "metadata": {
        "description": "容器组的名称"
      }
    },
    "imageVersion": {
      "type": "string",
      "defaultValue": "latest",
      "metadata": {
        "description": "容器镜像版本"
      }
    }
  },
  "variables": {
    "containerName": "moai-adk",
    "registryServer": "myregistry.azurecr.io",
    "imageName": "[concat(variables('registryServer'), '/moai-adk:', parameters('imageVersion'))]"
  },
  "resources": [
    {
      "type": "Microsoft.ContainerInstance/containerGroups",
      "apiVersion": "2023-05-01",
      "name": "[parameters('containerGroupName')]",
      "location": "[resourceGroup().location]",
      "properties": {
        "containers": [
          {
            "name": "[variables('containerName')]",
            "properties": {
              "image": "[variables('imageName')]",
              "ports": [
                {
                  "port": 8000,
                  "protocol": "TCP"
                }
              ],
              "environmentVariables": [
                {
                  "name": "ENVIRONMENT",
                  "value": "production"
                }
              ],
              "resources": {
                "requests": {
                  "cpu": 2,
                  "memoryInGb": 4
                }
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
          ],
          "dnsNameLabel": "[parameters('containerGroupName')]"
        },
        "imageRegistryCredentials": [
          {
            "server": "[variables('registryServer')]",
            "username": "[parameters('registryUsername')]",
            "password": "[parameters('registryPassword')]"
          }
        ]
      }
    }
  ],
  "outputs": {
    "containerFQDN": {
      "type": "string",
      "value": "[reference(resourceId('Microsoft.ContainerInstance/containerGroups', parameters('containerGroupName'))).ipAddress.fqdn]"
    }
  }
}
```

## CI/CD 流水线构建

### GitHub Actions

#### 质量检查工作流

```yaml
# .github/workflows/quality-check.yml
name: Quality Check

on:
  pull_request:
    branches: [develop, main]
  push:
    branches: [develop]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.11', '3.12', '3.13']

    steps:
    - uses: actions/checkout@v4

    - name: 安装 uv
      uses: astral-sh/setup-uv@v3
      with:
        version: "latest"

    - name: 设置 Python ${{ matrix.python-version }}
      run: uv python install ${{ matrix.python-version }}

    - name: 安装依赖
      run: |
        uv sync --all-extras --dev

    - name: 运行代码检查
      run: |
        uv run ruff check .
        uv run ruff format --check .

    - name: 运行类型检查
      run: |
        uv run mypy src/

    - name: 运行测试
      run: |
        uv run pytest --cov=src --cov-report=xml --cov-report=term

    - name: 上传覆盖率
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.xml
        flags: unittests
        name: codecov-umbrella

    - name: 安全扫描
      run: |
        uv run bandit -r src/ -f json -o bandit-report.json

    - name: TRUST 验证
      run: |
        uv run python scripts/validate_trust.py
```

#### 构建和推送镜像

```yaml
# .github/workflows/build-image.yml
name: Build and Push Docker Image

on:
  push:
    branches: [main]
    tags:
      - 'v*.*.*'

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - uses: actions/checkout@v4

    - name: 设置 Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: 登录容器注册表
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: 提取元数据
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=sha

    - name: 构建并推送
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: 扫描镜像
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.meta.outputs.version }}
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: 上传扫描结果
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: 'trivy-results.sarif'
```

#### 部署到预发布环境

```yaml
# .github/workflows/deploy-staging.yml
name: Deploy to Staging

on:
  workflow_run:
    workflows: ["Build and Push Docker Image"]
    types:
      - completed
    branches: [develop]

jobs:
  deploy-staging:
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}
    environment:
      name: staging
      url: https://staging.moai-adk.example.com

    steps:
    - uses: actions/checkout@v4

    - name: 配置 AWS 凭证
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: 更新 ECS 服务
      run: |
        aws ecs update-service \
          --cluster moai-adk-staging \
          --service moai-adk-service \
          --force-new-deployment

    - name: 等待部署完成
      run: |
        aws ecs wait services-stable \
          --cluster moai-adk-staging \
          --services moai-adk-service

    - name: 运行冒烟测试
      run: |
        curl -f https://staging.moai-adk.example.com/health || exit 1

    - name: 通知部署
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'Staging deployment completed'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

#### 部署到生产环境

```yaml
# .github/workflows/deploy-production.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.moai-adk.example.com

    steps:
    - uses: actions/checkout@v4

    - name: 配置 AWS 凭证
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: 获取当前任务定义
      run: |
        aws ecs describe-task-definition \
          --task-definition moai-adk \
          --query taskDefinition > task-definition.json

    - name: 更新任务定义中的镜像
      run: |
        NEW_IMAGE="${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.ref_name }}"
        jq --arg IMAGE "$NEW_IMAGE" \
           '.containerDefinitions[0].image = $IMAGE' \
           task-definition.json > new-task-definition.json

    - name: 注册新任务定义
      run: |
        aws ecs register-task-definition \
          --cli-input-json file://new-task-definition.json

    - name: 蓝绿部署
      run: |
        # 获取当前任务数量
        DESIRED_COUNT=$(aws ecs describe-services \
          --cluster moai-adk-prod \
          --services moai-adk-service \
          --query 'services[0].desiredCount' \
          --output text)

        # 部署新版本（绿色）
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-green \
          --task-definition moai-adk:$REVISION \
          --desired-count $DESIRED_COUNT

        # 等待绿色环境稳定
        aws ecs wait services-stable \
          --cluster moai-adk-prod \
          --services moai-adk-service-green

        # 将流量切换到绿色
        aws elbv2 modify-rule \
          --rule-arn ${{ secrets.ALB_RULE_ARN }} \
          --actions Type=forward,TargetGroupArn=${{ secrets.GREEN_TG_ARN }}

        # 缩减蓝色环境
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-blue \
          --desired-count 0

    - name: 运行生产测试
      run: |
        npm install -g newman
        newman run tests/production-tests.json \
          --environment production.json \
          --reporters cli,json \
          --reporter-json-export test-results.json

    - name: 创建 GitHub 发布
      uses: softprops/action-gh-release@v1
      with:
        body_path: CHANGELOG.md
        generate_release_notes: true

    - name: 通知部署
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'Production deployment completed: ${{ github.ref_name }}'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

### GitLab CI/CD

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
  IMAGE_TAG: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA

.python-base:
  image: python:3.13-slim
  before_script:
    - pip install uv
    - uv sync --all-extras --dev

validate:
  extends: .python-base
  stage: validate
  script:
    - uv run ruff check .
    - uv run ruff format --check .
    - uv run mypy src/
  only:
    - merge_requests
    - develop
    - main

test:
  extends: .python-base
  stage: test
  script:
    - uv run pytest --cov=src --cov-report=xml --cov-report=term
    - uv run bandit -r src/
  coverage: '/^TOTAL.+?(\d+\%)$/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
  only:
    - merge_requests
    - develop
    - main

build:
  stage: build
  image: docker:24-dind
  services:
    - docker:24-dind
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker build -t $IMAGE_TAG .
    - docker push $IMAGE_TAG
    - |
      if [ "$CI_COMMIT_BRANCH" == "main" ]; then
        docker tag $IMAGE_TAG $CI_REGISTRY_IMAGE:latest
        docker push $CI_REGISTRY_IMAGE:latest
      fi
  only:
    - develop
    - main
    - tags

deploy-staging:
  stage: deploy-staging
  image: alpine/k8s:1.28.3
  script:
    - kubectl config use-context staging
    - kubectl set image deployment/moai-adk moai-adk=$IMAGE_TAG -n moai-adk
    - kubectl rollout status deployment/moai-adk -n moai-adk
    - kubectl get pods -n moai-adk
  environment:
    name: staging
    url: https://staging.moai-adk.example.com
  only:
    - develop

deploy-production:
  stage: deploy-production
  image: alpine/k8s:1.28.3
  script:
    - kubectl config use-context production
    - |
      # 蓝绿部署
      kubectl apply -f k8s/deployment-green.yaml
      kubectl wait --for=condition=available --timeout=300s deployment/moai-adk-green -n moai-adk
      kubectl patch service moai-adk-service -n moai-adk -p '{"spec":{"selector":{"version":"green"}}}'
      kubectl scale deployment/moai-adk-blue --replicas=0 -n moai-adk
  environment:
    name: production
    url: https://api.moai-adk.example.com
  when: manual
  only:
    - main
    - tags
```

## 环境特定部署策略

### 开发环境

**目的**：本地开发和调试

**特点**：
- 启用热重载
- 开启调试模式
- 详细日志记录
- 本地数据库
- 某些服务无需认证

**配置**：

```yaml
# .env.development
ENVIRONMENT=development
DEBUG=true
LOG_LEVEL=DEBUG
DATABASE_URL=postgresql://dev:devpass@localhost:5432/moai_dev
REDIS_URL=redis://localhost:6379/0
ALLOWED_HOSTS=["localhost", "127.0.0.1"]
CORS_ORIGINS=["http://localhost:3000", "http://localhost:8000"]
```

**部署命令**：

```bash
# 使用 Docker Compose
docker-compose -f docker-compose.dev.yml up

# 或直接使用 Python
uv run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### 预发布环境

**目的**：使用类生产数据进行预生产测试

**特点**：
- 类生产配置
- 合成或匿名化数据
- 启用性能测试
- 完整认证
- 启用监控

**配置**：

```yaml
# .env.staging
ENVIRONMENT=staging
DEBUG=false
LOG_LEVEL=INFO
DATABASE_URL=${STAGING_DATABASE_URL}
REDIS_URL=${STAGING_REDIS_URL}
ALLOWED_HOSTS=["staging.moai-adk.example.com"]
CORS_ORIGINS=["https://staging-app.moai-adk.example.com"]
SENTRY_DSN=${SENTRY_DSN}
SENTRY_ENVIRONMENT=staging
```

**部署策略**：

```bash
# 部署到预发布
kubectl apply -f k8s/staging/

# 运行集成测试
npm run test:integration -- --env=staging

# 冒烟测试
curl -f https://staging.moai-adk.example.com/health
```

### 生产环境

**目的**：为真实用户提供实时生产服务

**特点**：
- 高可用性（3+ 副本）
- 启用自动扩展
- 高级监控和告警
- 严格的安全策略
- 蓝绿或金丝雀部署

**配置**：

```yaml
# .env.production（存储在密钥管理中）
ENVIRONMENT=production
DEBUG=false
LOG_LEVEL=WARNING
DATABASE_URL=${PROD_DATABASE_URL}
REDIS_URL=${PROD_REDIS_URL}
ALLOWED_HOSTS=["api.moai-adk.example.com"]
CORS_ORIGINS=["https://app.moai-adk.example.com"]
SENTRY_DSN=${SENTRY_DSN}
SENTRY_ENVIRONMENT=production
SECRET_KEY=${SECRET_KEY}
```

**部署策略**：

```bash
# 蓝绿部署
./scripts/deploy-blue-green.sh production v1.2.3

# 验证健康状态
./scripts/health-check.sh production

# 监控指标
./scripts/monitor-deployment.sh production v1.2.3
```

## 发布策略

### 蓝绿部署

在两个相同环境之间完成流量切换。

**脚本**：

```bash
#!/bin/bash
# scripts/deploy-blue-green.sh

set -euo pipefail

ENVIRONMENT=$1
VERSION=$2
NAMESPACE="moai-adk"

echo "为 ${ENVIRONMENT} 启动蓝绿部署（版本：${VERSION}）"

# 识别当前活动部署（蓝色或绿色）
CURRENT_COLOR=$(kubectl get service moai-adk-service -n $NAMESPACE \
  -o jsonpath='{.spec.selector.version}')

if [ "$CURRENT_COLOR" == "blue" ]; then
  NEW_COLOR="green"
else
  NEW_COLOR="blue"
fi

echo "当前：$CURRENT_COLOR，部署：$NEW_COLOR"

# 更新新部署
kubectl set image deployment/moai-adk-$NEW_COLOR \
  moai-adk=myregistry/moai-adk:$VERSION \
  -n $NAMESPACE

# 等待新部署就绪
kubectl rollout status deployment/moai-adk-$NEW_COLOR -n $NAMESPACE

# 对新部署运行冒烟测试
NEW_POD=$(kubectl get pod -l app=moai-adk,version=$NEW_COLOR -n $NAMESPACE \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward $NEW_POD 8080:8000 -n $NAMESPACE &
PF_PID=$!
sleep 5

if ! curl -f http://localhost:8080/health; then
  echo "新部署健康检查失败"
  kill $PF_PID
  exit 1
fi
kill $PF_PID

# 将流量切换到新部署
kubectl patch service moai-adk-service -n $NAMESPACE \
  -p "{\"spec\":{\"selector\":{\"version\":\"$NEW_COLOR\"}}}"

echo "流量已切换到 $NEW_COLOR"

# 监控 5 分钟
echo "监控新部署 5 分钟..."
sleep 300

# 缩减旧部署
kubectl scale deployment/moai-adk-$CURRENT_COLOR --replicas=0 -n $NAMESPACE

echo "蓝绿部署成功完成"
```

### 金丝雀部署

使用 Argo Rollouts 逐步将流量转移到新版本。

**发布配置**：

```yaml
# k8s/rollout.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: moai-adk
  namespace: moai-adk
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 5m}
      - setWeight: 25
      - pause: {duration: 5m}
      - setWeight: 50
      - pause: {duration: 10m}
      - setWeight: 75
      - pause: {duration: 5m}
      canaryService: moai-adk-canary
      stableService: moai-adk-stable
      trafficRouting:
        nginx:
          stableIngress: moai-adk-stable
          annotationPrefix: nginx.ingress.kubernetes.io
      analysis:
        templates:
        - templateName: success-rate
        - templateName: latency
        args:
        - name: service-name
          value: moai-adk-canary
  selector:
    matchLabels:
      app: moai-adk
  template:
    metadata:
      labels:
        app: moai-adk
    spec:
      containers:
      - name: moai-adk
        image: myregistry/moai-adk:VERSION
        ports:
        - containerPort: 8000
```

**分析模板**：

```yaml
# k8s/analysis-template.yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
  namespace: moai-adk
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 1m
    successCondition: result >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus:9090
        query: |
          sum(rate(
            http_requests_total{
              service="{{args.service-name}}",
              status=~"2.."
            }[5m]
          )) /
          sum(rate(
            http_requests_total{
              service="{{args.service-name}}"
            }[5m]
          ))
---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: latency
  namespace: moai-adk
spec:
  args:
  - name: service-name
  metrics:
  - name: p95-latency
    interval: 1m
    successCondition: result <= 500
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus:9090
        query: |
          histogram_quantile(0.95,
            sum(rate(
              http_request_duration_seconds_bucket{
                service="{{args.service-name}}"
              }[5m]
            )) by (le)
          ) * 1000
```

### 滚动更新

Kubernetes 原生滚动更新策略。

```yaml
# k8s/deployment-rolling.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: moai-adk
  namespace: moai-adk
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 2        # 更新期间最多额外 2 个 Pod
      maxUnavailable: 0  # 确保零停机时间
  selector:
    matchLabels:
      app: moai-adk
  template:
    metadata:
      labels:
        app: moai-adk
        version: v1.2.3
    spec:
      containers:
      - name: moai-adk
        image: myregistry/moai-adk:v1.2.3
        ports:
        - containerPort: 8000
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8000
          initialDelaySeconds: 30
          periodSeconds: 10
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 15"]
      terminationGracePeriodSeconds: 30
```

**部署命令**：

```bash
# 应用新版本
kubectl set image deployment/moai-adk \
  moai-adk=myregistry/moai-adk:v1.2.3 \
  -n moai-adk

# 观察发布进度
kubectl rollout status deployment/moai-adk -n moai-adk

# 必要时回滚
kubectl rollout undo deployment/moai-adk -n moai-adk
```

## 监控和日志

### Prometheus 指标

**应用指标**：

```python
# src/monitoring/metrics.py
from prometheus_client import Counter, Histogram, Gauge
import time
from functools import wraps

# 定义指标
http_requests_total = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)

http_request_duration_seconds = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration in seconds',
    ['method', 'endpoint']
)

active_connections = Gauge(
    'active_connections',
    'Number of active connections'
)

database_query_duration_seconds = Histogram(
    'database_query_duration_seconds',
    'Database query duration in seconds',
    ['query_type']
)

def track_request_metrics(func):
    """跟踪 HTTP 请求指标的装饰器"""
    @wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = time.time()

        try:
            response = await func(*args, **kwargs)
            status = response.status_code

            # 记录指标
            http_requests_total.labels(
                method=request.method,
                endpoint=request.path,
                status=status
            ).inc()

            duration = time.time() - start_time
            http_request_duration_seconds.labels(
                method=request.method,
                endpoint=request.path
            ).observe(duration)

            return response

        except Exception as e:
            http_requests_total.labels(
                method=request.method,
                endpoint=request.path,
                status=500
            ).inc()
            raise

    return wrapper
```

**Prometheus 配置**：

```yaml
# prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'moai-adk-prod'
    environment: 'production'

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093

rule_files:
  - '/etc/prometheus/rules/*.yml'

scrape_configs:
  - job_name: 'moai-adk'
    kubernetes_sd_configs:
    - role: pod
      namespaces:
        names:
        - moai-adk
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

**告警规则**：

```yaml
# prometheus/rules/alerts.yml
groups:
  - name: moai-adk-alerts
    interval: 30s
    rules:
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) /
          sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
          component: application
        annotations:
          summary: "检测到高错误率"
          description: "错误率为 {{ $value | humanizePercentage }}（阈值：5%）"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le)
          ) > 1
        for: 10m
        labels:
          severity: warning
          component: application
        annotations:
          summary: "检测到高延迟"
          description: "P95 延迟为 {{ $value }}秒（阈值：1秒）"

      - alert: PodCrashLooping
        expr: |
          rate(kube_pod_container_status_restarts_total{namespace="moai-adk"}[15m]) > 0
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "Pod {{ $labels.pod }} 正在崩溃循环"
          description: "Pod 在过去 15 分钟内重启了 {{ $value }} 次"

      - alert: HighMemoryUsage
        expr: |
          container_memory_usage_bytes{namespace="moai-adk"} /
          container_spec_memory_limit_bytes{namespace="moai-adk"} > 0.9
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "{{ $labels.pod }} 内存使用率高"
          description: "内存使用率为 {{ $value | humanizePercentage }}"
```

### Grafana 仪表板

**仪表板 JSON**（摘录）：

```json
{
  "dashboard": {
    "title": "MoAI-ADK 生产仪表板",
    "panels": [
      {
        "title": "请求速率",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"moai-adk\"}[5m])) by (status)",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "P95 延迟",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace=\"moai-adk\"}[5m])) by (le))",
            "legendFormat": "P95"
          }
        ]
      },
      {
        "title": "错误率",
        "type": "singlestat",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"moai-adk\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{namespace=\"moai-adk\"}[5m]))"
          }
        ]
      }
    ]
  }
}
```

### 结构化日志

```python
# src/logging/structured.py
import json
import logging
import sys
from datetime import datetime
from typing import Any, Dict

class StructuredLogger:
    def __init__(self, name: str, level: str = "INFO"):
        self.logger = logging.getLogger(name)
        self.logger.setLevel(getattr(logging, level.upper()))

        # JSON 格式化器
        handler = logging.StreamHandler(sys.stdout)
        handler.setFormatter(self.JsonFormatter())
        self.logger.addHandler(handler)

    class JsonFormatter(logging.Formatter):
        def format(self, record: logging.LogRecord) -> str:
            log_data: Dict[str, Any] = {
                "timestamp": datetime.utcnow().isoformat(),
                "level": record.levelname,
                "logger": record.name,
                "message": record.getMessage(),
                "module": record.module,
                "function": record.funcName,
                "line": record.lineno,
            }

            # 添加额外字段
            if hasattr(record, "extra"):
                log_data.update(record.extra)

            # 添加异常信息（如果存在）
            if record.exc_info:
                log_data["exception"] = self.formatException(record.exc_info)

            return json.dumps(log_data)

    def info(self, message: str, **kwargs):
        self.logger.info(message, extra=kwargs)

    def warning(self, message: str, **kwargs):
        self.logger.warning(message, extra=kwargs)

    def error(self, message: str, **kwargs):
        self.logger.error(message, extra=kwargs)

    def debug(self, message: str, **kwargs):
        self.logger.debug(message, extra=kwargs)

# 使用方式
logger = StructuredLogger("moai-adk")

logger.info(
    "用户登录成功",
    user_id="12345",
    ip_address="192.168.1.1",
    duration_ms=245
)
```

## 安全与合规

### 容器安全

**扫描 Dockerfile**：

```bash
# 使用 Trivy 扫描
trivy image --severity HIGH,CRITICAL myregistry/moai-adk:v1.2.3

# 使用 Snyk 扫描
snyk container test myregistry/moai-adk:v1.2.3

# 使用 Docker Scout 扫描
docker scout cves myregistry/moai-adk:v1.2.3
```

**安全最佳实践**：

```dockerfile
# 使用最小基础镜像
FROM python:3.13-slim

# 以非 root 用户运行
RUN useradd -m -u 1000 appuser
USER appuser

# 设置安全头
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# 限制资源使用
LABEL com.docker.swarm.cpu-quota="1000000"
LABEL com.docker.swarm.memory-limit="1073741824"
```

### 网络策略

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: moai-adk-netpol
  namespace: moai-adk
spec:
  podSelector:
    matchLabels:
      app: moai-adk
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8000
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: moai-adk
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443   # HTTPS external
    - protocol: TCP
      port: 53    # DNS
    - protocol: UDP
      port: 53
```

### OWASP 安全扫描

```bash
#!/bin/bash
# scripts/security-scan.sh

set -euo pipefail

echo "运行 OWASP 安全扫描..."

# 依赖检查
dependency-check \
  --project "MoAI-ADK" \
  --scan . \
  --format HTML \
  --format JSON \
  --out reports/dependency-check

# ZAP 扫描
docker run -v $(pwd):/zap/wrk/:rw \
  -t owasp/zap2docker-stable zap-baseline.py \
  -t https://staging.moai-adk.example.com \
  -r zap-report.html

# Bandit Python 安全
bandit -r src/ -f json -o reports/bandit-report.json

echo "安全扫描完成"
```

## 部署自动化

### 流水线编排

```python
# scripts/deploy_pipeline.py
import asyncio
from dataclasses import dataclass
from enum import Enum
from typing import List, Optional

class DeploymentStage(Enum):
    VALIDATE = "validate"
    BUILD = "build"
    TEST = "test"
    DEPLOY_STAGING = "deploy_staging"
    INTEGRATION_TEST = "integration_test"
    DEPLOY_PRODUCTION = "deploy_production"
    SMOKE_TEST = "smoke_test"

@dataclass
class StageResult:
    stage: DeploymentStage
    success: bool
    duration: float
    message: str

class DeploymentPipeline:
    def __init__(self, version: str, environment: str):
        self.version = version
        self.environment = environment
        self.results: List[StageResult] = []

    async def run(self) -> bool:
        """执行完整部署流水线"""
        stages = [
            self.validate_code,
            self.build_image,
            self.run_tests,
            self.deploy_to_staging,
            self.run_integration_tests,
        ]

        if self.environment == "production":
            stages.extend([
                self.deploy_to_production,
                self.run_smoke_tests,
            ])

        for stage_func in stages:
            result = await stage_func()
            self.results.append(result)

            if not result.success:
                await self.rollback()
                return False

        return True

    async def validate_code(self) -> StageResult:
        """验证代码质量"""
        start = time.time()

        # 运行代码检查
        lint_result = await asyncio.create_subprocess_exec(
            "uv", "run", "ruff", "check", ".",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await lint_result.wait()

        # 运行类型检查
        mypy_result = await asyncio.create_subprocess_exec(
            "uv", "run", "mypy", "src/",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await mypy_result.wait()

        success = lint_result.returncode == 0 and mypy_result.returncode == 0
        duration = time.time() - start

        return StageResult(
            stage=DeploymentStage.VALIDATE,
            success=success,
            duration=duration,
            message="代码验证完成" if success else "代码验证失败"
        )

    async def build_image(self) -> StageResult:
        """构建 Docker 镜像"""
        start = time.time()

        process = await asyncio.create_subprocess_exec(
            "docker", "build",
            "-t", f"myregistry/moai-adk:{self.version}",
            ".",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await process.wait()

        success = process.returncode == 0
        duration = time.time() - start

        return StageResult(
            stage=DeploymentStage.BUILD,
            success=success,
            duration=duration,
            message=f"镜像已构建：{self.version}" if success else "构建失败"
        )

    async def rollback(self):
        """失败时回滚部署"""
        logger.error("部署失败，启动回滚")

        # 获取上一个版本
        previous_version = await self.get_previous_version()

        # 回滚到上一个版本
        await asyncio.create_subprocess_exec(
            "kubectl", "rollout", "undo",
            "deployment/moai-adk",
            "-n", "moai-adk"
        )

        logger.info(f"已回滚到版本：{previous_version}")

# 使用方式
async def main():
    pipeline = DeploymentPipeline(
        version="v1.2.3",
        environment="production"
    )

    success = await pipeline.run()

    if success:
        print("部署成功完成")
    else:
        print("部署失败")

if __name__ == "__main__":
    asyncio.run(main())
```

### 多环境管理

```python
# scripts/environment_manager.py
from dataclasses import dataclass
from typing import Dict, List
import yaml

@dataclass
class Environment:
    name: str
    namespace: str
    replicas: int
    resources: Dict[str, str]
    database_url: str
    redis_url: str

class EnvironmentManager:
    def __init__(self, config_path: str):
        with open(config_path) as f:
            self.config = yaml.safe_load(f)

    def get_environment(self, name: str) -> Environment:
        """获取环境配置"""
        env_config = self.config['environments'][name]

        return Environment(
            name=name,
            namespace=env_config['namespace'],
            replicas=env_config['replicas'],
            resources=env_config['resources'],
            database_url=env_config['database_url'],
            redis_url=env_config['redis_url']
        )

    def deploy(self, environment: str, version: str):
        """部署到特定环境"""
        env = self.get_environment(environment)

        # 生成 Kubernetes 清单
        manifests = self.generate_manifests(env, version)

        # 应用清单
        for manifest in manifests:
            self.apply_manifest(manifest, env.namespace)

        # 等待发布
        self.wait_for_rollout(env.namespace)

    def generate_manifests(self, env: Environment, version: str) -> List[str]:
        """为环境生成 Kubernetes 清单"""
        # 实现细节...
        pass

# 使用方式
manager = EnvironmentManager('config/environments.yaml')
manager.deploy('production', 'v1.2.3')
```

### 部署跟踪

```python
# src/deployment/tracker.py
from datetime import datetime
from typing import Optional
from sqlalchemy import Column, String, DateTime, Boolean, Integer
from database import Base

class Deployment(Base):
    __tablename__ = "deployments"

    id = Column(Integer, primary_key=True)
    version = Column(String, nullable=False)
    environment = Column(String, nullable=False)
    status = Column(String, nullable=False)  # pending, in_progress, succeeded, failed, rolled_back
    started_at = Column(DateTime, default=datetime.utcnow)
    completed_at = Column(DateTime, nullable=True)
    deployed_by = Column(String, nullable=False)
    commit_sha = Column(String, nullable=False)
    rollback_version = Column(String, nullable=True)

class DeploymentTracker:
    def __init__(self, db_session):
        self.db = db_session

    def start_deployment(
        self,
        version: str,
        environment: str,
        deployed_by: str,
        commit_sha: str
    ) -> Deployment:
        """记录部署开始"""
        deployment = Deployment(
            version=version,
            environment=environment,
            status="in_progress",
            deployed_by=deployed_by,
            commit_sha=commit_sha
        )

        self.db.add(deployment)
        self.db.commit()

        return deployment

    def complete_deployment(self, deployment_id: int, success: bool):
        """标记部署为已完成"""
        deployment = self.db.query(Deployment).get(deployment_id)
        deployment.status = "succeeded" if success else "failed"
        deployment.completed_at = datetime.utcnow()

        self.db.commit()

    def get_current_version(self, environment: str) -> Optional[str]:
        """获取当前部署的版本"""
        deployment = (
            self.db.query(Deployment)
            .filter_by(environment=environment, status="succeeded")
            .order_by(Deployment.completed_at.desc())
            .first()
        )

        return deployment.version if deployment else None
```

## 性能优化

### 数据库优化

```sql
-- 为频繁查询创建索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status) WHERE status != 'completed';

-- 按日期分区大表
CREATE TABLE orders_2024_01 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 分析查询性能
EXPLAIN ANALYZE
SELECT * FROM orders
WHERE user_id = 123 AND status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- 使用物化视图优化
CREATE MATERIALIZED VIEW user_order_stats AS
SELECT
    user_id,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_spent,
    MAX(created_at) as last_order_date
FROM orders
GROUP BY user_id;

CREATE INDEX idx_user_order_stats_user_id ON user_order_stats(user_id);

-- 刷新物化视图
REFRESH MATERIALIZED VIEW CONCURRENTLY user_order_stats;
```

```python
# src/database/optimization.py
from sqlalchemy import event
from sqlalchemy.orm import Session
from sqlalchemy.pool import Pool

# 连接池配置
engine = create_engine(
    DATABASE_URL,
    pool_size=20,              # 维护的连接数
    max_overflow=40,           # 最大溢出连接数
    pool_timeout=30,           # 获取连接的超时时间
    pool_recycle=3600,         # 1 小时后回收连接
    pool_pre_ping=True,        # 验证连接健康状况
    echo_pool=True,            # 记录连接池事件
)

# 查询优化
class QueryOptimizer:
    @staticmethod
    def optimize_select(query):
        """添加查询优化"""
        return (
            query
            .with_hint(User, 'USE INDEX (idx_users_email)')
            .execution_options(compiled_cache={})
        )

    @staticmethod
    def batch_insert(session: Session, model, data: List[dict]):
        """优化的批量插入"""
        session.bulk_insert_mappings(model, data)
        session.commit()

    @staticmethod
    def batch_update(session: Session, model, data: List[dict]):
        """优化的批量更新"""
        session.bulk_update_mappings(model, data)
        session.commit()
```

### 缓存策略

```python
# src/cache/redis_cache.py
from typing import Optional, Any
import redis
import pickle
from functools import wraps

class RedisCache:
    def __init__(self, redis_url: str):
        self.redis = redis.from_url(redis_url)

    def get(self, key: str) -> Optional[Any]:
        """从缓存获取值"""
        value = self.redis.get(key)
        return pickle.loads(value) if value else None

    def set(self, key: str, value: Any, ttl: int = 3600):
        """设置缓存值并设置 TTL"""
        self.redis.setex(key, ttl, pickle.dumps(value))

    def delete(self, key: str):
        """从缓存删除键"""
        self.redis.delete(key)

    def cached(self, ttl: int = 3600):
        """缓存函数结果的装饰器"""
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # 生成缓存键
                cache_key = f"{func.__name__}:{str(args)}:{str(kwargs)}"

                # 检查缓存
                cached_value = self.get(cache_key)
                if cached_value is not None:
                    return cached_value

                # 执行函数
                result = await func(*args, **kwargs)

                # 存储在缓存中
                self.set(cache_key, result, ttl)

                return result
            return wrapper
        return decorator

# 使用方式
cache = RedisCache("redis://localhost:6379/0")

@cache.cached(ttl=600)
async def get_user_profile(user_id: int):
    # 昂贵的数据库查询
    return await db.fetch_one(
        "SELECT * FROM users WHERE id = :user_id",
        {"user_id": user_id}
    )
```

### 应用性能监控

```python
# src/monitoring/performance.py
import time
from functools import wraps
from typing import Callable

def performance_monitor(threshold_ms: float = 1000):
    """监控函数性能并记录慢操作"""
    def decorator(func: Callable):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            start_time = time.time()

            try:
                result = await func(*args, **kwargs)
                return result
            finally:
                duration_ms = (time.time() - start_time) * 1000

                if duration_ms > threshold_ms:
                    logger.warning(
                        f"检测到慢操作：{func.__name__}",
                        duration_ms=duration_ms,
                        threshold_ms=threshold_ms,
                        args=str(args),
                        kwargs=str(kwargs)
                    )

        return wrapper
    return decorator

# 使用方式
@performance_monitor(threshold_ms=500)
async def process_large_dataset(data: List[dict]):
    # 处理逻辑
    pass
```

## 故障排除和恢复

### 回滚脚本

```bash
#!/bin/bash
# scripts/rollback.sh

set -euo pipefail

ENVIRONMENT=$1
TARGET_VERSION=${2:-"previous"}
NAMESPACE="moai-adk"

echo "将 ${ENVIRONMENT} 回滚到 ${TARGET_VERSION}"

if [ "$TARGET_VERSION" == "previous" ]; then
  # 回滚到上一个修订版本
  kubectl rollout undo deployment/moai-adk -n $NAMESPACE
else
  # 回滚到特定版本
  kubectl set image deployment/moai-adk \
    moai-adk=myregistry/moai-adk:$TARGET_VERSION \
    -n $NAMESPACE
fi

# 等待发布完成
kubectl rollout status deployment/moai-adk -n $NAMESPACE

# 验证健康状态
echo "验证健康状态..."
sleep 30

HEALTH_URL="https://${ENVIRONMENT}.moai-adk.example.com/health"
if curl -f $HEALTH_URL; then
  echo "回滚成功"
else
  echo "回滚后健康检查失败"
  exit 1
fi
```

### 事件响应手册

```yaml
# docs/incident-response.yaml
playbooks:
  - name: 高错误率
    trigger: 错误率 > 5% 持续 5 分钟
    severity: critical
    steps:
      - 检查最近的部署（过去 24 小时）
      - 检查 Sentry/CloudWatch 中的错误日志
      - 检查数据库连接池状态
      - 验证外部服务依赖
      - 如果与部署相关，回滚到上一个版本
      - 如果是流量激增，扩展资源
      - 创建事件报告

  - name: 高延迟
    trigger: P95 延迟 > 1秒 持续 10 分钟
    severity: warning
    steps:
      - 检查数据库查询性能（慢查询日志）
      - 检查缓存命中率（Redis）
      - 检查外部 API 延迟
      - 检查最近的代码更改
      - 分析 APM 跟踪（Sentry/DataDog）
      - 如果与流量相关，水平扩展
      - 如果与数据库相关，优化慢查询

  - name: 服务宕机
    trigger: 健康检查失败
    severity: critical
    steps:
      - 检查 Pod 状态（kubectl get pods）
      - 查看最近的 Pod 日志（kubectl logs）
      - 检查资源使用情况（CPU/内存）
      - 验证网络连接
      - 检查数据库连接
      - 必要时重启 Pod
      - 如果与部署相关，回滚
      - 如果是基础设施问题，升级
```

### 诊断命令

```bash
#!/bin/bash
# scripts/diagnose.sh

NAMESPACE="moai-adk"

echo "=== Pod 状态 ==="
kubectl get pods -n $NAMESPACE

echo -e "\n=== 最近事件 ==="
kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' | tail -20

echo -e "\n=== 资源使用情况 ==="
kubectl top pods -n $NAMESPACE

echo -e "\n=== Pod 日志（最后 100 行）==="
POD=$(kubectl get pod -n $NAMESPACE -l app=moai-adk -o jsonpath='{.items[0].metadata.name}')
kubectl logs $POD -n $NAMESPACE --tail=100

echo -e "\n=== 服务端点 ==="
kubectl get endpoints -n $NAMESPACE

echo -e "\n=== Ingress 状态 ==="
kubectl get ingress -n $NAMESPACE

echo -e "\n=== 最近部署 ==="
kubectl rollout history deployment/moai-adk -n $NAMESPACE
```

## 最佳实践

### 部署前检查清单

- [ ] **代码质量**
  - [ ] 所有测试通过（单元、集成、E2E）
  - [ ] 代码覆盖率 > 80%
  - [ ] 代码检查和类型检查通过
  - [ ] 安全扫描完成（无 HIGH/CRITICAL 漏洞）
  - [ ] TRUST 5 验证通过

- [ ] **配置**
  - [ ] 环境变量已配置
  - [ ] 密钥正确管理（不在代码中）
  - [ ] 数据库迁移已准备
  - [ ] 功能标志已配置
  - [ ] 监控告警已配置

- [ ] **文档**
  - [ ] SPEC 文档已更新
  - [ ] API 文档已生成
  - [ ] 部署运行手册已更新
  - [ ] 回滚计划已记录

- [ ] **基础设施**
  - [ ] 资源限制已配置
  - [ ] 自动扩展策略已设置
  - [ ] 备份策略已验证
  - [ ] 灾难恢复计划已审查

- [ ] **沟通**
  - [ ] 利益相关者已通知部署窗口
  - [ ] 维护页面已准备（如需要）
  - [ ] 回滚团队待命
  - [ ] 事件响应联系人已验证

### 部署后检查清单

- [ ] **健康验证**
  - [ ] 健康检查端点响应
  - [ ] 所有 Pod 运行并就绪
  - [ ] 日志中无错误激增
  - [ ] 指标在正常范围内

- [ ] **冒烟测试**
  - [ ] 关键用户流程已测试
  - [ ] API 端点响应
  - [ ] 数据库连接正常
  - [ ] 外部集成正常运行

- [ ] **监控**
  - [ ] 指标仪表板已审查
  - [ ] 错误率正常
  - [ ] 延迟在 SLA 内
  - [ ] 资源使用可接受

- [ ] **文档**
  - [ ] 部署已记录在跟踪系统中
  - [ ] 发布说明已发布
  - [ ] 已知问题已记录
  - [ ] 回滚计划已确认

- [ ] **沟通**
  - [ ] 利益相关者已通知完成
  - [ ] 支持团队已简要了解变更
  - [ ] 状态页面已更新

### 部署成功指标

| 指标 | 目标 | 关键阈值 |
|--------|--------|--------------------|
| 部署频率 | > 10/周 | < 1/周 |
| 变更交付时间 | < 1 小时 | > 1 天 |
| 平均恢复时间 (MTTR) | < 15 分钟 | > 1 小时 |
| 变更失败率 | < 5% | > 15% |
| 部署成功率 | > 95% | < 85% |
| 回滚率 | < 5% | > 15% |
| P95 部署持续时间 | < 30 分钟 | > 2 小时 |

---

## 总结

本部署指南涵盖了部署 MoAI-ADK 项目的完整生命周期：

1. **容器化**：具有安全最佳实践的 Docker 多阶段构建
2. **编排**：Kubernetes、Docker Compose 和云原生解决方案
3. **云平台**：AWS ECS、Google Cloud Run、Azure 容器实例
4. **CI/CD**：GitHub Actions 和 GitLab CI/CD 流水线
5. **环境管理**：开发、预发布和生产配置
6. **发布策略**：蓝绿、金丝雀和滚动部署
7. **监控**：Prometheus 指标、Grafana 仪表板、结构化日志
8. **安全**：容器扫描、网络策略、OWASP 合规
9. **自动化**：流水线编排和部署跟踪
10. **性能**：数据库优化、缓存和 APM
11. **恢复**：回滚程序和事件响应
12. **最佳实践**：全面的检查清单和成功指标

通过遵循这些模式和实践，团队可以实现符合 MoAI-ADK 的 SPEC-first 理念和 TRUST 5 原则的可靠、安全和高效部署。
