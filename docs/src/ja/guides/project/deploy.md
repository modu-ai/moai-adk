# 完全デプロイメントガイド

MoAI-ADKプロジェクトのデプロイメントプロセスを網羅的に解説するガイドです。Dockerコンテナデプロイメントからクラウドサービス、CI/CDパイプライン構築まで、様々なデプロイメント環境と最新のDevOpsプラクティスを詳しく説明します。

## 目次

01. [概要](#概要)
02. [デプロイメントアーキテクチャ](#デプロイメントアーキテクチャ)
03. [Dockerコンテナデプロイメント](#dockerコンテナデプロイメント)
04. [クラウドプラットフォームデプロイメント](#クラウドプラットフォームデプロイメント)
05. [CI/CDパイプライン構築](#cicdパイプライン構築)
06. [環境別デプロイメント戦略](#環境別デプロイメント戦略)
07. [ロールアウト戦略](#ロールアウト戦略)
08. [モニタリングとログ](#モニタリングとログ)
09. [セキュリティとコンプライアンス](#セキュリティとコンプライアンス)
10. [デプロイメント自動化](#デプロイメント自動化)
11. [パフォーマンス最適化](#パフォーマンス最適化)
12. [トラブルシューティングと復旧](#トラブルシューティングと復旧)
13. [ベストプラクティス](#ベストプラクティス)

## 概要

MoAI-ADKプロジェクトのデプロイメントは、単にコードをサーバーにアップロードすること以上のものです。安定的でスケーラブル、かつセキュアなシステムを構築することが目的です。AlfredのSPECファーストな開発アプローチにより、デプロイメントプロセス全体で一貫性とトレーサビリティが保証されます。

### デプロイメントシステムの核心原則

1. **自動化優先**: すべてのデプロイメントプロセスは自動化されている必要があります
2. **ゼロダウンタイム**: ユーザーにサービス中断を発生させません
3. **ロールバック準備**: いつでも以前のバージョンにロールバックできる必要があります
4. **可観測性**: デプロイメント状態とシステムヘルスをリアルタイムで可視化
5. **セキュア**: デプロイメントプロセス全体がセキュリティ規制に準拠

### MoAI-ADKのデプロイメント機能

- **SPEC駆動デプロイメント**: SPEC文書に基づいたデプロイメント計画
- **TAGベーストラッキング**: デプロイされるすべてのコンポーネントのTAG追跡
- **自動化された品質ゲート**: デプロイメント前の自動品質検証
- **段階的デプロイメント**: 段階的なデプロイメントのサポート
- **環境の一貫性**: すべての環境間での一貫性の保証

## デプロイメントアーキテクチャ

### モダンデプロイメントアーキテクチャ

```mermaid
graph TD
    A[開発者] --> B[Git Push]
    B --> C[CI/CDパイプライン]
    C --> D[ビルド & テスト]
    D --> E[セキュリティスキャン]
    E --> F[コンテナビルド]
    F --> G[アーティファクトレジストリ]
    G --> H[デプロイメント戦略]
    H --> I[ステージング]
    I --> J[自動テスト]
    J --> K{承認?}
    K -->|はい| L[本番デプロイメント]
    K -->|いいえ| M[ロールバック]
    L --> N[ヘルスチェック]
    N --> O[モニタリング]
    O --> P[アラート]
```

### MoAI-ADKデプロイメントワークフロー

```mermaid
sequenceDiagram
    participant Dev as 開発者
    participant Git as Gitリポジトリ
    participant CI as CI/CDパイプライン
    participant QA as 品質ゲート
    participant Deploy as デプロイメントシステム
    participant Prod as 本番環境

    Dev->>Git: /alfred:3-sync
    Git->>CI: パイプライントリガー
    CI->>CI: アプリケーションビルド
    CI->>CI: テスト実行 (TRUST検証)
    CI->>QA: 品質ゲートチェック
    QA->>CI: 合格/不合格
    CI->>Deploy: Dockerイメージ作成
    Deploy->>Deploy: ステージングにデプロイ
    Deploy->>QA: 自動E2Eテスト
    QA->>Deploy: 承認
    Deploy->>Prod: Blue-Greenデプロイメント
    Prod->>Deploy: ヘルスチェック
    Deploy->>Dev: デプロイメントレポート
```

## Dockerコンテナデプロイメント

### 基本Dockerfile設定

MoAI-ADKプロジェクト用に最適化されたDockerfile:

```dockerfile
# 本番環境最適化のためのマルチステージビルド
FROM python:3.13-slim as builder

# 作業ディレクトリ設定
WORKDIR /app

# システム依存関係インストール
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Python依存関係インストール
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv && \
    uv pip install --system --no-cache -r <(uv pip compile pyproject.toml)

# 本番ステージ
FROM python:3.13-slim

WORKDIR /app

# ランタイム依存関係のみインストール
RUN apt-get update && apt-get install -y \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

# ビルダーからインストール済みパッケージをコピー
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# アプリケーションコードコピー
COPY . .

# 非rootユーザー作成
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# ポート公開
EXPOSE 8000

# アプリケーション起動
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose設定

#### 開発環境

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
      - /app/.venv  # 仮想環境を除外
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

#### ステージング環境

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

#### 本番環境

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

### Dockerビルドとプッシュ

```bash
#!/bin/bash
# scripts/docker-build.sh

set -euo pipefail

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-docker.io/myorg}
IMAGE_NAME="moai-adk"

echo "Dockerイメージビルド中: ${REGISTRY}/${IMAGE_NAME}:${VERSION}"

# マルチプラットフォームイメージビルド
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
  --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
  --push \
  .

echo "イメージプッシュ成功"

# 脆弱性スキャン
docker scout cves "${REGISTRY}/${IMAGE_NAME}:${VERSION}"
```

## Kubernetesデプロイメント

### 基本Kubernetes設定

#### Namespace

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

## クラウドプラットフォームデプロイメント

### AWS ECS Fargate

#### タスク定義

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

#### CloudFormationテンプレート

```yaml
# cloudformation/moai-adk.yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'MoAI-ADK ECS Fargateデプロイメント'

Parameters:
  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC ID
  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: ECSタスク用サブネットID
  ImageVersion:
    Type: String
    Default: latest
    Description: Dockerイメージバージョン

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
    Description: ロードバランサーのURL
    Value: !GetAtt LoadBalancer.DNSName
```

### Google Cloud Run

#### サービス設定

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

#### Terraform設定

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

### Azure Container Instances

#### ARMテンプレート

```json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "containerGroupName": {
      "type": "string",
      "defaultValue": "moai-adk-group",
      "metadata": {
        "description": "コンテナグループ名"
      }
    },
    "imageVersion": {
      "type": "string",
      "defaultValue": "latest",
      "metadata": {
        "description": "コンテナイメージバージョン"
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

## CI/CDパイプライン構築

### GitHub Actions

#### 品質チェックワークフロー

```yaml
# .github/workflows/quality-check.yml
name: 品質チェック

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

    - name: uvインストール
      uses: astral-sh/setup-uv@v3
      with:
        version: "latest"

    - name: Python ${{ matrix.python-version }}セットアップ
      run: uv python install ${{ matrix.python-version }}

    - name: 依存関係インストール
      run: |
        uv sync --all-extras --dev

    - name: リンティング実行
      run: |
        uv run ruff check .
        uv run ruff format --check .

    - name: 型チェック実行
      run: |
        uv run mypy src/

    - name: テスト実行
      run: |
        uv run pytest --cov=src --cov-report=xml --cov-report=term

    - name: カバレッジアップロード
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.xml
        flags: unittests
        name: codecov-umbrella

    - name: セキュリティスキャン
      run: |
        uv run bandit -r src/ -f json -o bandit-report.json

    - name: TRUST検証
      run: |
        uv run python scripts/validate_trust.py
```

#### イメージビルドとプッシュ

```yaml
# .github/workflows/build-image.yml
name: Dockerイメージビルドとプッシュ

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

    - name: Docker Buildxセットアップ
      uses: docker/setup-buildx-action@v3

    - name: コンテナレジストリにログイン
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: メタデータ抽出
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

    - name: ビルドとプッシュ
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: イメージスキャン
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.meta.outputs.version }}
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: スキャン結果アップロード
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: 'trivy-results.sarif'
```

#### ステージングにデプロイ

```yaml
# .github/workflows/deploy-staging.yml
name: ステージングへデプロイ

on:
  workflow_run:
    workflows: ["Dockerイメージビルドとプッシュ"]
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

    - name: AWS認証情報設定
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: ECSサービス更新
      run: |
        aws ecs update-service \
          --cluster moai-adk-staging \
          --service moai-adk-service \
          --force-new-deployment

    - name: デプロイメント待機
      run: |
        aws ecs wait services-stable \
          --cluster moai-adk-staging \
          --services moai-adk-service

    - name: スモークテスト実行
      run: |
        curl -f https://staging.moai-adk.example.com/health || exit 1

    - name: デプロイメント通知
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'ステージングデプロイメント完了'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

#### 本番環境へデプロイ

```yaml
# .github/workflows/deploy-production.yml
name: 本番環境へデプロイ

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

    - name: AWS認証情報設定
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: 現在のタスク定義取得
      run: |
        aws ecs describe-task-definition \
          --task-definition moai-adk \
          --query taskDefinition > task-definition.json

    - name: タスク定義のイメージ更新
      run: |
        NEW_IMAGE="${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.ref_name }}"
        jq --arg IMAGE "$NEW_IMAGE" \
           '.containerDefinitions[0].image = $IMAGE' \
           task-definition.json > new-task-definition.json

    - name: 新しいタスク定義登録
      run: |
        aws ecs register-task-definition \
          --cli-input-json file://new-task-definition.json

    - name: Blue-Greenデプロイメント
      run: |
        # 現在のタスク数取得
        DESIRED_COUNT=$(aws ecs describe-services \
          --cluster moai-adk-prod \
          --services moai-adk-service \
          --query 'services[0].desiredCount' \
          --output text)

        # 新バージョンデプロイ (green)
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-green \
          --task-definition moai-adk:$REVISION \
          --desired-count $DESIRED_COUNT

        # greenの安定待機
        aws ecs wait services-stable \
          --cluster moai-adk-prod \
          --services moai-adk-service-green

        # トラフィックをgreenに切り替え
        aws elbv2 modify-rule \
          --rule-arn ${{ secrets.ALB_RULE_ARN }} \
          --actions Type=forward,TargetGroupArn=${{ secrets.GREEN_TG_ARN }}

        # blueをスケールダウン
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-blue \
          --desired-count 0

    - name: 本番テスト実行
      run: |
        npm install -g newman
        newman run tests/production-tests.json \
          --environment production.json \
          --reporters cli,json \
          --reporter-json-export test-results.json

    - name: GitHubリリース作成
      uses: softprops/action-gh-release@v1
      with:
        body_path: CHANGELOG.md
        generate_release_notes: true

    - name: デプロイメント通知
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: '本番デプロイメント完了: ${{ github.ref_name }}'
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
      # Blue-Greenデプロイメント
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

## 環境別デプロイメント戦略

### 開発環境

**目的**: ローカル開発とデバッグ

**特徴**:
- ホットリロード有効
- デバッグモードオン
- 詳細なログ出力
- ローカルデータベース
- 一部サービスの認証なし

**設定**:

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

**デプロイコマンド**:

```bash
# Docker Composeを使用
docker-compose -f docker-compose.dev.yml up

# または直接Python
uv run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### ステージング環境

**目的**: 本番に近いデータでの本番前テスト

**特徴**:
- 本番に近い設定
- 合成データまたは匿名化データ
- パフォーマンステスト有効
- 完全な認証
- モニタリング有効

**設定**:

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

**デプロイ戦略**:

```bash
# ステージングにデプロイ
kubectl apply -f k8s/staging/

# 統合テスト実行
npm run test:integration -- --env=staging

# スモークテスト
curl -f https://staging.moai-adk.example.com/health
```

### 本番環境

**目的**: 実際のユーザーにサービスを提供する本番環境

**特徴**:
- 高可用性 (3+レプリカ)
- オートスケーリング有効
- 高度なモニタリングとアラート
- 厳格なセキュリティポリシー
- Blue-greenまたはカナリアデプロイメント

**設定**:

```yaml
# .env.production (シークレット管理に保存)
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

**デプロイ戦略**:

```bash
# Blue-greenデプロイメント
./scripts/deploy-blue-green.sh production v1.2.3

# ヘルス確認
./scripts/health-check.sh production

# メトリクスモニタリング
./scripts/monitor-deployment.sh production v1.2.3
```

## ロールアウト戦略

### Blue-Greenデプロイメント

2つの同一環境間での完全なトラフィック切り替え。

**スクリプト**:

```bash
#!/bin/bash
# scripts/deploy-blue-green.sh

set -euo pipefail

ENVIRONMENT=$1
VERSION=$2
NAMESPACE="moai-adk"

echo "${ENVIRONMENT}のBlue-Greenデプロイメント開始 (バージョン: ${VERSION})"

# 現在アクティブなデプロイメント(blueまたはgreen)を特定
CURRENT_COLOR=$(kubectl get service moai-adk-service -n $NAMESPACE \
  -o jsonpath='{.spec.selector.version}')

if [ "$CURRENT_COLOR" == "blue" ]; then
  NEW_COLOR="green"
else
  NEW_COLOR="blue"
fi

echo "現在: $CURRENT_COLOR、デプロイ: $NEW_COLOR"

# 新しいデプロイメントを更新
kubectl set image deployment/moai-adk-$NEW_COLOR \
  moai-adk=myregistry/moai-adk:$VERSION \
  -n $NAMESPACE

# 新しいデプロイメントの準備完了を待機
kubectl rollout status deployment/moai-adk-$NEW_COLOR -n $NAMESPACE

# 新しいデプロイメントに対してスモークテスト実行
NEW_POD=$(kubectl get pod -l app=moai-adk,version=$NEW_COLOR -n $NAMESPACE \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward $NEW_POD 8080:8000 -n $NAMESPACE &
PF_PID=$!
sleep 5

if ! curl -f http://localhost:8080/health; then
  echo "新しいデプロイメントのヘルスチェック失敗"
  kill $PF_PID
  exit 1
fi
kill $PF_PID

# 新しいデプロイメントにトラフィックを切り替え
kubectl patch service moai-adk-service -n $NAMESPACE \
  -p "{\"spec\":{\"selector\":{\"version\":\"$NEW_COLOR\"}}}"

echo "トラフィックを$NEW_COLORに切り替え完了"

# 5分間モニタリング
echo "5分間新しいデプロイメントをモニタリング中..."
sleep 300

# 古いデプロイメントをスケールダウン
kubectl scale deployment/moai-adk-$CURRENT_COLOR --replicas=0 -n $NAMESPACE

echo "Blue-Greenデプロイメント成功"
```

### カナリアデプロイメント

Argo Rolloutsを使用した段階的なトラフィック切り替え。

**ロールアウト設定**:

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

**分析テンプレート**:

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

### ローリングアップデート

Kubernetesネイティブのローリングアップデート戦略。

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
      maxSurge: 2        # アップデート中の最大追加Pod数
      maxUnavailable: 0  # ゼロダウンタイムの保証
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

**デプロイコマンド**:

```bash
# 新バージョン適用
kubectl set image deployment/moai-adk \
  moai-adk=myregistry/moai-adk:v1.2.3 \
  -n moai-adk

# ロールアウト進捗監視
kubectl rollout status deployment/moai-adk -n moai-adk

# 必要に応じてロールバック
kubectl rollout undo deployment/moai-adk -n moai-adk
```

## モニタリングとログ

### Prometheusメトリクス

**アプリケーションメトリクス**:

```python
# src/monitoring/metrics.py
from prometheus_client import Counter, Histogram, Gauge
import time
from functools import wraps

# メトリクス定義
http_requests_total = Counter(
    'http_requests_total',
    'HTTPリクエスト総数',
    ['method', 'endpoint', 'status']
)

http_request_duration_seconds = Histogram(
    'http_request_duration_seconds',
    'HTTPリクエスト時間(秒)',
    ['method', 'endpoint']
)

active_connections = Gauge(
    'active_connections',
    'アクティブ接続数'
)

database_query_duration_seconds = Histogram(
    'database_query_duration_seconds',
    'データベースクエリ時間(秒)',
    ['query_type']
)

def track_request_metrics(func):
    """HTTPリクエストメトリクスを追跡するデコレーター"""
    @wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = time.time()

        try:
            response = await func(*args, **kwargs)
            status = response.status_code

            # メトリクス記録
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

**Prometheus設定**:

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

**アラートルール**:

```yaml
# prometheus/rules/alerts.yml
groups:
  - name: moai-adk-alerts
    interval: 30s
    rules:
      - alert: 高エラー率
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m])) /
          sum(rate(http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
          component: application
        annotations:
          summary: "高エラー率検出"
          description: "エラー率は{{ $value | humanizePercentage }}です (閾値: 5%)"

      - alert: 高レイテンシ
        expr: |
          histogram_quantile(0.95,
            sum(rate(http_request_duration_seconds_bucket[5m])) by (le)
          ) > 1
        for: 10m
        labels:
          severity: warning
          component: application
        annotations:
          summary: "高レイテンシ検出"
          description: "P95レイテンシは{{ $value }}秒です (閾値: 1秒)"

      - alert: Podクラッシュループ
        expr: |
          rate(kube_pod_container_status_restarts_total{namespace="moai-adk"}[15m]) > 0
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "Pod {{ $labels.pod }}がクラッシュループ中"
          description: "Podは過去15分で{{ $value }}回再起動しました"

      - alert: 高メモリ使用率
        expr: |
          container_memory_usage_bytes{namespace="moai-adk"} /
          container_spec_memory_limit_bytes{namespace="moai-adk"} > 0.9
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "{{ $labels.pod }}で高メモリ使用率"
          description: "メモリ使用率は{{ $value | humanizePercentage }}です"
```

### Grafanaダッシュボード

**ダッシュボードJSON** (抜粋):

```json
{
  "dashboard": {
    "title": "MoAI-ADK本番ダッシュボード",
    "panels": [
      {
        "title": "リクエストレート",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"moai-adk\"}[5m])) by (status)",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "P95レイテンシ",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace=\"moai-adk\"}[5m])) by (le))",
            "legendFormat": "P95"
          }
        ]
      },
      {
        "title": "エラー率",
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

### 構造化ログ

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

        # JSONフォーマッター
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

            # 追加フィールドを追加
            if hasattr(record, "extra"):
                log_data.update(record.extra)

            # 例外情報があれば追加
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

# 使用例
logger = StructuredLogger("moai-adk")

logger.info(
    "ユーザーログイン成功",
    user_id="12345",
    ip_address="192.168.1.1",
    duration_ms=245
)
```

## セキュリティとコンプライアンス

### コンテナセキュリティ

**Dockerfileスキャン**:

```bash
# Trivyでスキャン
trivy image --severity HIGH,CRITICAL myregistry/moai-adk:v1.2.3

# Snykでスキャン
snyk container test myregistry/moai-adk:v1.2.3

# Docker Scoutでスキャン
docker scout cves myregistry/moai-adk:v1.2.3
```

**セキュリティベストプラクティス**:

```dockerfile
# 最小限のベースイメージを使用
FROM python:3.13-slim

# 非rootユーザーとして実行
RUN useradd -m -u 1000 appuser
USER appuser

# セキュリティヘッダー設定
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# リソース使用量制限
LABEL com.docker.swarm.cpu-quota="1000000"
LABEL com.docker.swarm.memory-limit="1073741824"
```

### ネットワークポリシー

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

### OWASPセキュリティスキャン

```bash
#!/bin/bash
# scripts/security-scan.sh

set -euo pipefail

echo "OWASPセキュリティスキャン実行中..."

# 依存関係チェック
dependency-check \
  --project "MoAI-ADK" \
  --scan . \
  --format HTML \
  --format JSON \
  --out reports/dependency-check

# ZAPスキャン
docker run -v $(pwd):/zap/wrk/:rw \
  -t owasp/zap2docker-stable zap-baseline.py \
  -t https://staging.moai-adk.example.com \
  -r zap-report.html

# Bandit Pythonセキュリティ
bandit -r src/ -f json -o reports/bandit-report.json

echo "セキュリティスキャン完了"
```

## デプロイメント自動化

### パイプラインオーケストレーション

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
        """完全なデプロイメントパイプライン実行"""
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
        """コード品質検証"""
        start = time.time()

        # リンティング実行
        lint_result = await asyncio.create_subprocess_exec(
            "uv", "run", "ruff", "check", ".",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await lint_result.wait()

        # 型チェック実行
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
            message="コード検証完了" if success else "コード検証失敗"
        )

    async def build_image(self) -> StageResult:
        """Dockerイメージビルド"""
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
            message=f"イメージビルド完了: {self.version}" if success else "ビルド失敗"
        )

    async def rollback(self):
        """失敗時のデプロイメントロールバック"""
        logger.error("デプロイメント失敗、ロールバック開始")

        # 以前のバージョン取得
        previous_version = await self.get_previous_version()

        # 以前のバージョンにロールバック
        await asyncio.create_subprocess_exec(
            "kubectl", "rollout", "undo",
            "deployment/moai-adk",
            "-n", "moai-adk"
        )

        logger.info(f"バージョン{previous_version}にロールバック完了")

# 使用例
async def main():
    pipeline = DeploymentPipeline(
        version="v1.2.3",
        environment="production"
    )

    success = await pipeline.run()

    if success:
        print("デプロイメント成功")
    else:
        print("デプロイメント失敗")

if __name__ == "__main__":
    asyncio.run(main())
```

### マルチ環境管理

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
        """環境設定取得"""
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
        """特定の環境へデプロイ"""
        env = self.get_environment(environment)

        # Kubernetesマニフェスト生成
        manifests = self.generate_manifests(env, version)

        # マニフェスト適用
        for manifest in manifests:
            self.apply_manifest(manifest, env.namespace)

        # ロールアウト待機
        self.wait_for_rollout(env.namespace)

    def generate_manifests(self, env: Environment, version: str) -> List[str]:
        """環境用Kubernetesマニフェスト生成"""
        # 実装詳細...
        pass

# 使用例
manager = EnvironmentManager('config/environments.yaml')
manager.deploy('production', 'v1.2.3')
```

### デプロイメント追跡

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
        """デプロイメント開始記録"""
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
        """デプロイメント完了マーク"""
        deployment = self.db.query(Deployment).get(deployment_id)
        deployment.status = "succeeded" if success else "failed"
        deployment.completed_at = datetime.utcnow()

        self.db.commit()

    def get_current_version(self, environment: str) -> Optional[str]:
        """現在デプロイされているバージョン取得"""
        deployment = (
            self.db.query(Deployment)
            .filter_by(environment=environment, status="succeeded")
            .order_by(Deployment.completed_at.desc())
            .first()
        )

        return deployment.version if deployment else None
```

## パフォーマンス最適化

### データベース最適化

```sql
-- 頻繁なクエリ用のインデックス作成
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status) WHERE status != 'completed';

-- 日付による大きなテーブルのパーティション化
CREATE TABLE orders_2024_01 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- クエリパフォーマンス分析
EXPLAIN ANALYZE
SELECT * FROM orders
WHERE user_id = 123 AND status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- マテリアライズドビューで最適化
CREATE MATERIALIZED VIEW user_order_stats AS
SELECT
    user_id,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_spent,
    MAX(created_at) as last_order_date
FROM orders
GROUP BY user_id;

CREATE INDEX idx_user_order_stats_user_id ON user_order_stats(user_id);

-- マテリアライズドビュー更新
REFRESH MATERIALIZED VIEW CONCURRENTLY user_order_stats;
```

```python
# src/database/optimization.py
from sqlalchemy import event
from sqlalchemy.orm import Session
from sqlalchemy.pool import Pool

# 接続プーリング設定
engine = create_engine(
    DATABASE_URL,
    pool_size=20,              # 維持する接続数
    max_overflow=40,           # 最大オーバーフロー接続数
    pool_timeout=30,           # 接続取得タイムアウト
    pool_recycle=3600,         # 1時間後に接続をリサイクル
    pool_pre_ping=True,        # 接続ヘルス検証
    echo_pool=True,            # 接続プールイベントログ
)

# クエリ最適化
class QueryOptimizer:
    @staticmethod
    def optimize_select(query):
        """クエリ最適化追加"""
        return (
            query
            .with_hint(User, 'USE INDEX (idx_users_email)')
            .execution_options(compiled_cache={})
        )

    @staticmethod
    def batch_insert(session: Session, model, data: List[dict]):
        """最適化されたバッチインサート"""
        session.bulk_insert_mappings(model, data)
        session.commit()

    @staticmethod
    def batch_update(session: Session, model, data: List[dict]):
        """最適化されたバッチアップデート"""
        session.bulk_update_mappings(model, data)
        session.commit()
```

### キャッシング戦略

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
        """キャッシュから値取得"""
        value = self.redis.get(key)
        return pickle.loads(value) if value else None

    def set(self, key: str, value: Any, ttl: int = 3600):
        """TTL付きでキャッシュに値設定"""
        self.redis.setex(key, ttl, pickle.dumps(value))

    def delete(self, key: str):
        """キャッシュからキー削除"""
        self.redis.delete(key)

    def cached(self, ttl: int = 3600):
        """関数結果をキャッシュするデコレーター"""
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # キャッシュキー生成
                cache_key = f"{func.__name__}:{str(args)}:{str(kwargs)}"

                # キャッシュチェック
                cached_value = self.get(cache_key)
                if cached_value is not None:
                    return cached_value

                # 関数実行
                result = await func(*args, **kwargs)

                # キャッシュに保存
                self.set(cache_key, result, ttl)

                return result
            return wrapper
        return decorator

# 使用例
cache = RedisCache("redis://localhost:6379/0")

@cache.cached(ttl=600)
async def get_user_profile(user_id: int):
    # 高コストなデータベースクエリ
    return await db.fetch_one(
        "SELECT * FROM users WHERE id = :user_id",
        {"user_id": user_id}
    )
```

### アプリケーションパフォーマンスモニタリング

```python
# src/monitoring/performance.py
import time
from functools import wraps
from typing import Callable

def performance_monitor(threshold_ms: float = 1000):
    """関数パフォーマンスを監視し、遅い操作をログ"""
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
                        f"遅い操作検出: {func.__name__}",
                        duration_ms=duration_ms,
                        threshold_ms=threshold_ms,
                        args=str(args),
                        kwargs=str(kwargs)
                    )

        return wrapper
    return decorator

# 使用例
@performance_monitor(threshold_ms=500)
async def process_large_dataset(data: List[dict]):
    # 処理ロジック
    pass
```

## トラブルシューティングと復旧

### ロールバックスクリプト

```bash
#!/bin/bash
# scripts/rollback.sh

set -euo pipefail

ENVIRONMENT=$1
TARGET_VERSION=${2:-"previous"}
NAMESPACE="moai-adk"

echo "${ENVIRONMENT}を${TARGET_VERSION}にロールバック中"

if [ "$TARGET_VERSION" == "previous" ]; then
  # 以前のリビジョンにロールバック
  kubectl rollout undo deployment/moai-adk -n $NAMESPACE
else
  # 特定のバージョンにロールバック
  kubectl set image deployment/moai-adk \
    moai-adk=myregistry/moai-adk:$TARGET_VERSION \
    -n $NAMESPACE
fi

# ロールアウト待機
kubectl rollout status deployment/moai-adk -n $NAMESPACE

# ヘルス確認
echo "ヘルス確認中..."
sleep 30

HEALTH_URL="https://${ENVIRONMENT}.moai-adk.example.com/health"
if curl -f $HEALTH_URL; then
  echo "ロールバック成功"
else
  echo "ロールバック後のヘルスチェック失敗"
  exit 1
fi
```

### インシデント対応プレイブック

```yaml
# docs/incident-response.yaml
playbooks:
  - name: 高エラー率
    trigger: エラー率が5分間5%超過
    severity: critical
    steps:
      - 最近のデプロイメント確認 (過去24時間)
      - Sentry/CloudWatchでエラーログレビュー
      - データベース接続プールステータス確認
      - 外部サービス依存関係確認
      - デプロイメント関連の場合、以前のバージョンにロールバック
      - トラフィックスパイクの場合、リソースをスケールアップ
      - インシデントレポート作成

  - name: 高レイテンシ
    trigger: P95レイテンシが10分間1秒超過
    severity: warning
    steps:
      - データベースクエリパフォーマンス確認 (スロークエリログ)
      - キャッシュヒット率レビュー (Redis)
      - 外部APIレイテンシ確認
      - 最近のコード変更レビュー
      - APMトレース分析 (Sentry/DataDog)
      - トラフィック関連の場合、水平スケール
      - データベース関連の場合、遅いクエリを最適化

  - name: サービスダウン
    trigger: ヘルスチェック失敗
    severity: critical
    steps:
      - Podステータス確認 (kubectl get pods)
      - 最近のPodログレビュー (kubectl logs)
      - リソース使用状況確認 (CPU/メモリ)
      - ネットワーク接続確認
      - データベース接続確認
      - 必要に応じてPod再起動
      - デプロイメント関連の場合ロールバック
      - インフラ問題の場合エスカレーション
```

### 診断コマンド

```bash
#!/bin/bash
# scripts/diagnose.sh

NAMESPACE="moai-adk"

echo "=== Podステータス ==="
kubectl get pods -n $NAMESPACE

echo -e "\n=== 最近のイベント ==="
kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' | tail -20

echo -e "\n=== リソース使用状況 ==="
kubectl top pods -n $NAMESPACE

echo -e "\n=== Podログ (最新100行) ==="
POD=$(kubectl get pod -n $NAMESPACE -l app=moai-adk -o jsonpath='{.items[0].metadata.name}')
kubectl logs $POD -n $NAMESPACE --tail=100

echo -e "\n=== サービスエンドポイント ==="
kubectl get endpoints -n $NAMESPACE

echo -e "\n=== Ingressステータス ==="
kubectl get ingress -n $NAMESPACE

echo -e "\n=== 最近のデプロイメント ==="
kubectl rollout history deployment/moai-adk -n $NAMESPACE
```

## ベストプラクティス

### デプロイメント前チェックリスト

- [ ] **コード品質**
  - [ ] すべてのテストが合格 (ユニット、統合、E2E)
  - [ ] コードカバレッジ > 80%
  - [ ] リンティングと型チェックが合格
  - [ ] セキュリティスキャン完了 (HIGH/CRITICAL脆弱性なし)
  - [ ] TRUST 5検証が合格

- [ ] **設定**
  - [ ] 環境変数設定済み
  - [ ] シークレットが適切に管理されている (コード内にない)
  - [ ] データベースマイグレーション準備済み
  - [ ] フィーチャーフラグ設定済み
  - [ ] モニタリングアラート設定済み

- [ ] **ドキュメント**
  - [ ] SPEC文書更新済み
  - [ ] APIドキュメント生成済み
  - [ ] デプロイメントランブック更新済み
  - [ ] ロールバック計画文書化済み

- [ ] **インフラ**
  - [ ] リソース制限設定済み
  - [ ] オートスケーリングポリシー設定済み
  - [ ] バックアップ戦略確認済み
  - [ ] 災害復旧計画レビュー済み

- [ ] **コミュニケーション**
  - [ ] ステークホルダーにデプロイメント時間帯通知済み
  - [ ] メンテナンスページ準備済み (必要な場合)
  - [ ] ロールバックチーム待機中
  - [ ] インシデント対応連絡先確認済み

### デプロイメント後チェックリスト

- [ ] **ヘルス確認**
  - [ ] ヘルスチェックエンドポイント応答中
  - [ ] すべてのPodが実行中で準備完了
  - [ ] ログにエラースパイクなし
  - [ ] メトリクスが正常範囲内

- [ ] **スモークテスト**
  - [ ] 重要なユーザーフローテスト済み
  - [ ] APIエンドポイント応答中
  - [ ] データベース接続動作中
  - [ ] 外部統合が機能的

- [ ] **モニタリング**
  - [ ] メトリクスダッシュボードレビュー済み
  - [ ] エラー率が正常
  - [ ] レイテンシがSLA内
  - [ ] リソース使用量が許容範囲

- [ ] **ドキュメント**
  - [ ] デプロイメントが追跡システムに記録済み
  - [ ] リリースノート公開済み
  - [ ] 既知の問題文書化済み
  - [ ] ロールバック計画確認済み

- [ ] **コミュニケーション**
  - [ ] ステークホルダーに完了通知済み
  - [ ] サポートチームに変更内容説明済み
  - [ ] ステータスページ更新済み

### デプロイメント成功メトリクス

| メトリクス | 目標 | 重要閾値 |
|--------|--------|--------------------|
| デプロイメント頻度 | > 10回/週 | < 1回/週 |
| 変更のリードタイム | < 1時間 | > 1日 |
| 平均復旧時間 (MTTR) | < 15分 | > 1時間 |
| 変更失敗率 | < 5% | > 15% |
| デプロイメント成功率 | > 95% | < 85% |
| ロールバック率 | < 5% | > 15% |
| P95デプロイメント時間 | < 30分 | > 2時間 |

---

## まとめ

このデプロイメントガイドは、MoAI-ADKプロジェクトのデプロイメントライフサイクル全体をカバーしています:

1. **コンテナ化**: セキュリティベストプラクティスを含むDockerマルチステージビルド
2. **オーケストレーション**: Kubernetes、Docker Compose、クラウドネイティブソリューション
3. **クラウドプラットフォーム**: AWS ECS、Google Cloud Run、Azure Container Instances
4. **CI/CD**: GitHub ActionsとGitLab CI/CDパイプライン
5. **環境管理**: 開発、ステージング、本番環境設定
6. **ロールアウト戦略**: Blue-green、カナリア、ローリングデプロイメント
7. **モニタリング**: Prometheusメトリクス、Grafanaダッシュボード、構造化ログ
8. **セキュリティ**: コンテナスキャン、ネットワークポリシー、OWASPコンプライアンス
9. **自動化**: パイプラインオーケストレーションとデプロイメント追跡
10. **パフォーマンス**: データベース最適化、キャッシング、APM
11. **復旧**: ロールバック手順とインシデント対応
12. **ベストプラクティス**: 包括的なチェックリストと成功メトリクス

これらのパターンとプラクティスに従うことで、チームはMoAI-ADKのSPECファーストな哲学とTRUST 5原則に合致した、信頼性が高く、安全で、効率的なデプロイメントを実現できます。
