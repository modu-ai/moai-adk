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
