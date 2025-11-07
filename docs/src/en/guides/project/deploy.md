# Complete Deployment Guide

A comprehensive guide covering the deployment process for MoAI-ADK projects. This document details everything from Docker container deployment to cloud services and CI/CD pipeline construction, explaining various deployment environments and modern DevOps practices.

## Table of Contents

01. [Overview](#overview)
02. [Deployment Architecture](#deployment-architecture)
03. [Docker Container Deployment](#docker-container-deployment)
04. [Cloud Platform Deployment](#cloud-platform-deployment)
05. [CI/CD Pipeline Construction](#cicd-pipeline-construction)
06. [Environment-Specific Deployment Strategies](#environment-specific-deployment-strategies)
07. [Rollout Strategies](#rollout-strategies)
08. [Monitoring and Logging](#monitoring-and-logging)
09. [Security and Compliance](#security-and-compliance)
10. [Deployment Automation](#deployment-automation)
11. [Performance Optimization](#performance-optimization)
12. [Troubleshooting and Recovery](#troubleshooting-and-recovery)
13. [Best Practices](#best-practices)

## Overview

Deploying a MoAI-ADK project goes beyond simply uploading code to a server—it's about building a stable, scalable, and secure system. Alfred's SPEC-first development approach ensures consistency and traceability throughout the deployment process.

### Core Principles of Deployment Systems

1. **Automation First**: All deployment processes must be automated
2. **Zero Downtime**: No service interruptions for users
3. **Rollback Ready**: Must be able to rollback to previous versions at any time
4. **Observable**: Real-time visibility into deployment status and system health
5. **Secure**: The entire deployment process must comply with security regulations

### MoAI-ADK Deployment Features

- **SPEC-Driven Deployment**: Deployment planning based on SPEC documents
- **TAG-Based Tracking**: TAG tracking for all deployed components
- **Automated Quality Gates**: Automated quality verification before deployment
- **Progressive Deployment**: Support for gradual deployment
- **Environment Consistency**: Ensuring consistency across all environments

## Deployment Architecture

### Modern Deployment Architecture

```mermaid
graph TD
    A[Developer] --> B[Git Push]
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

### MoAI-ADK Deployment Workflow

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Git as Git Repository
    participant CI as CI/CD Pipeline
    participant QA as Quality Gate
    participant Deploy as Deployment System
    participant Prod as Production

    Dev->>Git: /alfred:3-sync
    Git->>CI: Trigger Pipeline
    CI->>CI: Build Application
    CI->>CI: Run Tests (TRUST verification)
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

## Docker Container Deployment

### Basic Dockerfile Configuration

Optimized Dockerfile for MoAI-ADK projects:

```dockerfile
# Multi-stage build for production optimization
FROM python:3.13-slim as builder

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    make \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY pyproject.toml uv.lock ./
RUN pip install --no-cache-dir uv && \
    uv pip install --system --no-cache -r <(uv pip compile pyproject.toml)

# Production stage
FROM python:3.13-slim

WORKDIR /app

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    libpq5 \
    && rm -rf /var/lib/apt/lists/*

# Copy installed packages from builder
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy application code
COPY . .

# Create non-root user
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Expose port
EXPOSE 8000

# Start application
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Docker Compose Configuration

#### Development Environment

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
      - /app/.venv  # Exclude virtual environment
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

#### Staging Environment

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

#### Production Environment

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

### Docker Build and Push

```bash
#!/bin/bash
# scripts/docker-build.sh

set -euo pipefail

VERSION=${1:-latest}
REGISTRY=${DOCKER_REGISTRY:-docker.io/myorg}
IMAGE_NAME="moai-adk"

echo "Building Docker image: ${REGISTRY}/${IMAGE_NAME}:${VERSION}"

# Build multi-platform image
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
  --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
  --push \
  .

echo "Image pushed successfully"

# Scan for vulnerabilities
docker scout cves "${REGISTRY}/${IMAGE_NAME}:${VERSION}"
```

## Kubernetes Deployment

### Basic Kubernetes Configuration

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

## Cloud Platform Deployment

### AWS ECS Fargate

#### Task Definition

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

#### CloudFormation Template

```yaml
# cloudformation/moai-adk.yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'MoAI-ADK ECS Fargate Deployment'

Parameters:
  VpcId:
    Type: AWS::EC2::VPC::Id
    Description: VPC ID
  SubnetIds:
    Type: List<AWS::EC2::Subnet::Id>
    Description: Subnet IDs for ECS tasks
  ImageVersion:
    Type: String
    Default: latest
    Description: Docker image version

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
    Description: URL of the load balancer
    Value: !GetAtt LoadBalancer.DNSName
```

### Google Cloud Run

#### Service Configuration

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

#### Terraform Configuration

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

#### ARM Template

```json
{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "containerGroupName": {
      "type": "string",
      "defaultValue": "moai-adk-group",
      "metadata": {
        "description": "Name for the container group"
      }
    },
    "imageVersion": {
      "type": "string",
      "defaultValue": "latest",
      "metadata": {
        "description": "Container image version"
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

## CI/CD Pipeline Construction

### GitHub Actions

#### Quality Check Workflow

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

    - name: Install uv
      uses: astral-sh/setup-uv@v3
      with:
        version: "latest"

    - name: Set up Python ${{ matrix.python-version }}
      run: uv python install ${{ matrix.python-version }}

    - name: Install dependencies
      run: |
        uv sync --all-extras --dev

    - name: Run linting
      run: |
        uv run ruff check .
        uv run ruff format --check .

    - name: Run type checking
      run: |
        uv run mypy src/

    - name: Run tests
      run: |
        uv run pytest --cov=src --cov-report=xml --cov-report=term

    - name: Upload coverage
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.xml
        flags: unittests
        name: codecov-umbrella

    - name: Security scan
      run: |
        uv run bandit -r src/ -f json -o bandit-report.json

    - name: TRUST validation
      run: |
        uv run python scripts/validate_trust.py
```

#### Build and Push Image

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
          type=sha

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

    - name: Scan image
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.meta.outputs.version }}
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload scan results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: 'trivy-results.sarif'
```

#### Deploy to Staging

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

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: Update ECS service
      run: |
        aws ecs update-service \
          --cluster moai-adk-staging \
          --service moai-adk-service \
          --force-new-deployment

    - name: Wait for deployment
      run: |
        aws ecs wait services-stable \
          --cluster moai-adk-staging \
          --services moai-adk-service

    - name: Run smoke tests
      run: |
        curl -f https://staging.moai-adk.example.com/health || exit 1

    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        text: 'Staging deployment completed'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

#### Deploy to Production

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

    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v4
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-east-1

    - name: Get current task definition
      run: |
        aws ecs describe-task-definition \
          --task-definition moai-adk \
          --query taskDefinition > task-definition.json

    - name: Update image in task definition
      run: |
        NEW_IMAGE="${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.ref_name }}"
        jq --arg IMAGE "$NEW_IMAGE" \
           '.containerDefinitions[0].image = $IMAGE' \
           task-definition.json > new-task-definition.json

    - name: Register new task definition
      run: |
        aws ecs register-task-definition \
          --cli-input-json file://new-task-definition.json

    - name: Blue-Green deployment
      run: |
        # Get current task count
        DESIRED_COUNT=$(aws ecs describe-services \
          --cluster moai-adk-prod \
          --services moai-adk-service \
          --query 'services[0].desiredCount' \
          --output text)

        # Deploy new version (green)
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-green \
          --task-definition moai-adk:$REVISION \
          --desired-count $DESIRED_COUNT

        # Wait for green to be stable
        aws ecs wait services-stable \
          --cluster moai-adk-prod \
          --services moai-adk-service-green

        # Switch traffic to green
        aws elbv2 modify-rule \
          --rule-arn ${{ secrets.ALB_RULE_ARN }} \
          --actions Type=forward,TargetGroupArn=${{ secrets.GREEN_TG_ARN }}

        # Scale down blue
        aws ecs update-service \
          --cluster moai-adk-prod \
          --service moai-adk-service-blue \
          --desired-count 0

    - name: Run production tests
      run: |
        npm install -g newman
        newman run tests/production-tests.json \
          --environment production.json \
          --reporters cli,json \
          --reporter-json-export test-results.json

    - name: Create GitHub release
      uses: softprops/action-gh-release@v1
      with:
        body_path: CHANGELOG.md
        generate_release_notes: true

    - name: Notify deployment
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
      # Blue-Green deployment
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

## Environment-Specific Deployment Strategies

### Development Environment

**Purpose**: Local development and debugging

**Characteristics**:
- Hot reload enabled
- Debug mode on
- Detailed logging
- Local database
- No authentication for some services

**Configuration**:

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

**Deployment Command**:

```bash
# Using Docker Compose
docker-compose -f docker-compose.dev.yml up

# Or direct Python
uv run uvicorn main:app --reload --host 0.0.0.0 --port 8000
```

### Staging Environment

**Purpose**: Pre-production testing with production-like data

**Characteristics**:
- Production-like configuration
- Synthetic or anonymized data
- Performance testing enabled
- Full authentication
- Monitoring enabled

**Configuration**:

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

**Deployment Strategy**:

```bash
# Deploy to staging
kubectl apply -f k8s/staging/

# Run integration tests
npm run test:integration -- --env=staging

# Smoke tests
curl -f https://staging.moai-adk.example.com/health
```

### Production Environment

**Purpose**: Live production serving real users

**Characteristics**:
- High availability (3+ replicas)
- Auto-scaling enabled
- Advanced monitoring and alerting
- Strict security policies
- Blue-green or canary deployments

**Configuration**:

```yaml
# .env.production (stored in secrets management)
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

**Deployment Strategy**:

```bash
# Blue-green deployment
./scripts/deploy-blue-green.sh production v1.2.3

# Verify health
./scripts/health-check.sh production

# Monitor metrics
./scripts/monitor-deployment.sh production v1.2.3
```

## Rollout Strategies

### Blue-Green Deployment

Complete traffic switch between two identical environments.

**Script**:

```bash
#!/bin/bash
# scripts/deploy-blue-green.sh

set -euo pipefail

ENVIRONMENT=$1
VERSION=$2
NAMESPACE="moai-adk"

echo "Starting blue-green deployment for ${ENVIRONMENT} (version: ${VERSION})"

# Identify current active deployment (blue or green)
CURRENT_COLOR=$(kubectl get service moai-adk-service -n $NAMESPACE \
  -o jsonpath='{.spec.selector.version}')

if [ "$CURRENT_COLOR" == "blue" ]; then
  NEW_COLOR="green"
else
  NEW_COLOR="blue"
fi

echo "Current: $CURRENT_COLOR, Deploying: $NEW_COLOR"

# Update the new deployment
kubectl set image deployment/moai-adk-$NEW_COLOR \
  moai-adk=myregistry/moai-adk:$VERSION \
  -n $NAMESPACE

# Wait for new deployment to be ready
kubectl rollout status deployment/moai-adk-$NEW_COLOR -n $NAMESPACE

# Run smoke tests against new deployment
NEW_POD=$(kubectl get pod -l app=moai-adk,version=$NEW_COLOR -n $NAMESPACE \
  -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward $NEW_POD 8080:8000 -n $NAMESPACE &
PF_PID=$!
sleep 5

if ! curl -f http://localhost:8080/health; then
  echo "Health check failed on new deployment"
  kill $PF_PID
  exit 1
fi
kill $PF_PID

# Switch traffic to new deployment
kubectl patch service moai-adk-service -n $NAMESPACE \
  -p "{\"spec\":{\"selector\":{\"version\":\"$NEW_COLOR\"}}}"

echo "Traffic switched to $NEW_COLOR"

# Monitor for 5 minutes
echo "Monitoring new deployment for 5 minutes..."
sleep 300

# Scale down old deployment
kubectl scale deployment/moai-adk-$CURRENT_COLOR --replicas=0 -n $NAMESPACE

echo "Blue-green deployment completed successfully"
```

### Canary Deployment

Gradual traffic shift to new version using Argo Rollouts.

**Rollout Configuration**:

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

**Analysis Template**:

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

### Rolling Update

Kubernetes native rolling update strategy.

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
      maxSurge: 2        # Maximum 2 extra pods during update
      maxUnavailable: 0  # Ensure zero downtime
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

**Deployment Command**:

```bash
# Apply new version
kubectl set image deployment/moai-adk \
  moai-adk=myregistry/moai-adk:v1.2.3 \
  -n moai-adk

# Watch rollout progress
kubectl rollout status deployment/moai-adk -n moai-adk

# Rollback if needed
kubectl rollout undo deployment/moai-adk -n moai-adk
```

## Monitoring and Logging

### Prometheus Metrics

**Application Metrics**:

```python
# src/monitoring/metrics.py
from prometheus_client import Counter, Histogram, Gauge
import time
from functools import wraps

# Define metrics
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
    """Decorator to track HTTP request metrics"""
    @wraps(func)
    async def wrapper(*args, **kwargs):
        start_time = time.time()

        try:
            response = await func(*args, **kwargs)
            status = response.status_code

            # Record metrics
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

**Prometheus Configuration**:

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

**Alert Rules**:

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
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"

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
          summary: "High latency detected"
          description: "P95 latency is {{ $value }}s (threshold: 1s)"

      - alert: PodCrashLooping
        expr: |
          rate(kube_pod_container_status_restarts_total{namespace="moai-adk"}[15m]) > 0
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "Pod {{ $labels.pod }} is crash looping"
          description: "Pod has restarted {{ $value }} times in the last 15 minutes"

      - alert: HighMemoryUsage
        expr: |
          container_memory_usage_bytes{namespace="moai-adk"} /
          container_spec_memory_limit_bytes{namespace="moai-adk"} > 0.9
        for: 5m
        labels:
          severity: warning
          component: infrastructure
        annotations:
          summary: "High memory usage in {{ $labels.pod }}"
          description: "Memory usage is {{ $value | humanizePercentage }}"
```

### Grafana Dashboards

**Dashboard JSON** (excerpt):

```json
{
  "dashboard": {
    "title": "MoAI-ADK Production Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"moai-adk\"}[5m])) by (status)",
            "legendFormat": "{{status}}"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace=\"moai-adk\"}[5m])) by (le))",
            "legendFormat": "P95"
          }
        ]
      },
      {
        "title": "Error Rate",
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

### Structured Logging

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

        # JSON formatter
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

            # Add extra fields
            if hasattr(record, "extra"):
                log_data.update(record.extra)

            # Add exception info if present
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

# Usage
logger = StructuredLogger("moai-adk")

logger.info(
    "User login successful",
    user_id="12345",
    ip_address="192.168.1.1",
    duration_ms=245
)
```

## Security and Compliance

### Container Security

**Scan Dockerfile**:

```bash
# Scan with Trivy
trivy image --severity HIGH,CRITICAL myregistry/moai-adk:v1.2.3

# Scan with Snyk
snyk container test myregistry/moai-adk:v1.2.3

# Scan with Docker Scout
docker scout cves myregistry/moai-adk:v1.2.3
```

**Security Best Practices**:

```dockerfile
# Use minimal base images
FROM python:3.13-slim

# Run as non-root user
RUN useradd -m -u 1000 appuser
USER appuser

# Set security headers
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Limit resource usage
LABEL com.docker.swarm.cpu-quota="1000000"
LABEL com.docker.swarm.memory-limit="1073741824"
```

### Network Policies

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

### OWASP Security Scanning

```bash
#!/bin/bash
# scripts/security-scan.sh

set -euo pipefail

echo "Running OWASP security scans..."

# Dependency check
dependency-check \
  --project "MoAI-ADK" \
  --scan . \
  --format HTML \
  --format JSON \
  --out reports/dependency-check

# ZAP scan
docker run -v $(pwd):/zap/wrk/:rw \
  -t owasp/zap2docker-stable zap-baseline.py \
  -t https://staging.moai-adk.example.com \
  -r zap-report.html

# Bandit Python security
bandit -r src/ -f json -o reports/bandit-report.json

echo "Security scan completed"
```

## Deployment Automation

### Pipeline Orchestration

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
        """Execute full deployment pipeline"""
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
        """Validate code quality"""
        start = time.time()

        # Run linting
        lint_result = await asyncio.create_subprocess_exec(
            "uv", "run", "ruff", "check", ".",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        await lint_result.wait()

        # Run type checking
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
            message="Code validation completed" if success else "Code validation failed"
        )

    async def build_image(self) -> StageResult:
        """Build Docker image"""
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
            message=f"Image built: {self.version}" if success else "Build failed"
        )

    async def rollback(self):
        """Rollback deployment on failure"""
        logger.error("Deployment failed, initiating rollback")

        # Get previous version
        previous_version = await self.get_previous_version()

        # Rollback to previous version
        await asyncio.create_subprocess_exec(
            "kubectl", "rollout", "undo",
            "deployment/moai-adk",
            "-n", "moai-adk"
        )

        logger.info(f"Rolled back to version: {previous_version}")

# Usage
async def main():
    pipeline = DeploymentPipeline(
        version="v1.2.3",
        environment="production"
    )

    success = await pipeline.run()

    if success:
        print("Deployment completed successfully")
    else:
        print("Deployment failed")

if __name__ == "__main__":
    asyncio.run(main())
```

### Multi-Environment Management

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
        """Get environment configuration"""
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
        """Deploy to specific environment"""
        env = self.get_environment(environment)

        # Generate Kubernetes manifests
        manifests = self.generate_manifests(env, version)

        # Apply manifests
        for manifest in manifests:
            self.apply_manifest(manifest, env.namespace)

        # Wait for rollout
        self.wait_for_rollout(env.namespace)

    def generate_manifests(self, env: Environment, version: str) -> List[str]:
        """Generate Kubernetes manifests for environment"""
        # Implementation details...
        pass

# Usage
manager = EnvironmentManager('config/environments.yaml')
manager.deploy('production', 'v1.2.3')
```

### Deployment Tracking

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
        """Record deployment start"""
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
        """Mark deployment as completed"""
        deployment = self.db.query(Deployment).get(deployment_id)
        deployment.status = "succeeded" if success else "failed"
        deployment.completed_at = datetime.utcnow()

        self.db.commit()

    def get_current_version(self, environment: str) -> Optional[str]:
        """Get currently deployed version"""
        deployment = (
            self.db.query(Deployment)
            .filter_by(environment=environment, status="succeeded")
            .order_by(Deployment.completed_at.desc())
            .first()
        )

        return deployment.version if deployment else None
```

## Performance Optimization

### Database Optimization

```sql
-- Create indexes for frequent queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status) WHERE status != 'completed';

-- Partition large tables by date
CREATE TABLE orders_2024_01 PARTITION OF orders
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Analyze query performance
EXPLAIN ANALYZE
SELECT * FROM orders
WHERE user_id = 123 AND status = 'pending'
ORDER BY created_at DESC
LIMIT 10;

-- Optimize with materialized view
CREATE MATERIALIZED VIEW user_order_stats AS
SELECT
    user_id,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_spent,
    MAX(created_at) as last_order_date
FROM orders
GROUP BY user_id;

CREATE INDEX idx_user_order_stats_user_id ON user_order_stats(user_id);

-- Refresh materialized view
REFRESH MATERIALIZED VIEW CONCURRENTLY user_order_stats;
```

```python
# src/database/optimization.py
from sqlalchemy import event
from sqlalchemy.orm import Session
from sqlalchemy.pool import Pool

# Connection pooling configuration
engine = create_engine(
    DATABASE_URL,
    pool_size=20,              # Number of connections to maintain
    max_overflow=40,           # Maximum overflow connections
    pool_timeout=30,           # Timeout for getting connection
    pool_recycle=3600,         # Recycle connections after 1 hour
    pool_pre_ping=True,        # Verify connection health
    echo_pool=True,            # Log connection pool events
)

# Query optimization
class QueryOptimizer:
    @staticmethod
    def optimize_select(query):
        """Add query optimizations"""
        return (
            query
            .with_hint(User, 'USE INDEX (idx_users_email)')
            .execution_options(compiled_cache={})
        )

    @staticmethod
    def batch_insert(session: Session, model, data: List[dict]):
        """Optimized batch insert"""
        session.bulk_insert_mappings(model, data)
        session.commit()

    @staticmethod
    def batch_update(session: Session, model, data: List[dict]):
        """Optimized batch update"""
        session.bulk_update_mappings(model, data)
        session.commit()
```

### Caching Strategy

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
        """Get value from cache"""
        value = self.redis.get(key)
        return pickle.loads(value) if value else None

    def set(self, key: str, value: Any, ttl: int = 3600):
        """Set value in cache with TTL"""
        self.redis.setex(key, ttl, pickle.dumps(value))

    def delete(self, key: str):
        """Delete key from cache"""
        self.redis.delete(key)

    def cached(self, ttl: int = 3600):
        """Decorator for caching function results"""
        def decorator(func):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # Generate cache key
                cache_key = f"{func.__name__}:{str(args)}:{str(kwargs)}"

                # Check cache
                cached_value = self.get(cache_key)
                if cached_value is not None:
                    return cached_value

                # Execute function
                result = await func(*args, **kwargs)

                # Store in cache
                self.set(cache_key, result, ttl)

                return result
            return wrapper
        return decorator

# Usage
cache = RedisCache("redis://localhost:6379/0")

@cache.cached(ttl=600)
async def get_user_profile(user_id: int):
    # Expensive database query
    return await db.fetch_one(
        "SELECT * FROM users WHERE id = :user_id",
        {"user_id": user_id}
    )
```

### Application Performance Monitoring

```python
# src/monitoring/performance.py
import time
from functools import wraps
from typing import Callable

def performance_monitor(threshold_ms: float = 1000):
    """Monitor function performance and log slow operations"""
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
                        f"Slow operation detected: {func.__name__}",
                        duration_ms=duration_ms,
                        threshold_ms=threshold_ms,
                        args=str(args),
                        kwargs=str(kwargs)
                    )

        return wrapper
    return decorator

# Usage
@performance_monitor(threshold_ms=500)
async def process_large_dataset(data: List[dict]):
    # Processing logic
    pass
```

## Troubleshooting and Recovery

### Rollback Scripts

```bash
#!/bin/bash
# scripts/rollback.sh

set -euo pipefail

ENVIRONMENT=$1
TARGET_VERSION=${2:-"previous"}
NAMESPACE="moai-adk"

echo "Rolling back ${ENVIRONMENT} to ${TARGET_VERSION}"

if [ "$TARGET_VERSION" == "previous" ]; then
  # Rollback to previous revision
  kubectl rollout undo deployment/moai-adk -n $NAMESPACE
else
  # Rollback to specific version
  kubectl set image deployment/moai-adk \
    moai-adk=myregistry/moai-adk:$TARGET_VERSION \
    -n $NAMESPACE
fi

# Wait for rollout
kubectl rollout status deployment/moai-adk -n $NAMESPACE

# Verify health
echo "Verifying health..."
sleep 30

HEALTH_URL="https://${ENVIRONMENT}.moai-adk.example.com/health"
if curl -f $HEALTH_URL; then
  echo "Rollback successful"
else
  echo "Health check failed after rollback"
  exit 1
fi
```

### Incident Response Playbook

```yaml
# docs/incident-response.yaml
playbooks:
  - name: High Error Rate
    trigger: Error rate > 5% for 5 minutes
    severity: critical
    steps:
      - Check recent deployments (last 24 hours)
      - Review error logs in Sentry/CloudWatch
      - Check database connection pool status
      - Verify external service dependencies
      - If deployment-related, rollback to previous version
      - If traffic spike, scale up resources
      - Create incident report

  - name: High Latency
    trigger: P95 latency > 1s for 10 minutes
    severity: warning
    steps:
      - Check database query performance (slow query log)
      - Review cache hit rate (Redis)
      - Check external API latency
      - Review recent code changes
      - Analyze APM traces (Sentry/DataDog)
      - Scale horizontally if traffic-related
      - Optimize slow queries if database-related

  - name: Service Down
    trigger: Health check failing
    severity: critical
    steps:
      - Check pod status (kubectl get pods)
      - Review recent pod logs (kubectl logs)
      - Check resource usage (CPU/memory)
      - Verify network connectivity
      - Check database connectivity
      - Restart pods if necessary
      - Rollback if deployment-related
      - Escalate if infrastructure issue
```

### Diagnostic Commands

```bash
#!/bin/bash
# scripts/diagnose.sh

NAMESPACE="moai-adk"

echo "=== Pod Status ==="
kubectl get pods -n $NAMESPACE

echo -e "\n=== Recent Events ==="
kubectl get events -n $NAMESPACE --sort-by='.lastTimestamp' | tail -20

echo -e "\n=== Resource Usage ==="
kubectl top pods -n $NAMESPACE

echo -e "\n=== Pod Logs (last 100 lines) ==="
POD=$(kubectl get pod -n $NAMESPACE -l app=moai-adk -o jsonpath='{.items[0].metadata.name}')
kubectl logs $POD -n $NAMESPACE --tail=100

echo -e "\n=== Service Endpoints ==="
kubectl get endpoints -n $NAMESPACE

echo -e "\n=== Ingress Status ==="
kubectl get ingress -n $NAMESPACE

echo -e "\n=== Recent Deployments ==="
kubectl rollout history deployment/moai-adk -n $NAMESPACE
```

## Best Practices

### Pre-Deployment Checklist

- [ ] **Code Quality**
  - [ ] All tests passing (unit, integration, e2e)
  - [ ] Code coverage > 80%
  - [ ] Linting and type checking passed
  - [ ] Security scan completed (no HIGH/CRITICAL vulnerabilities)
  - [ ] TRUST 5 validation passed

- [ ] **Configuration**
  - [ ] Environment variables configured
  - [ ] Secrets properly managed (not in code)
  - [ ] Database migrations prepared
  - [ ] Feature flags configured
  - [ ] Monitoring alerts configured

- [ ] **Documentation**
  - [ ] SPEC documents updated
  - [ ] API documentation generated
  - [ ] Deployment runbook updated
  - [ ] Rollback plan documented

- [ ] **Infrastructure**
  - [ ] Resource limits configured
  - [ ] Auto-scaling policies set
  - [ ] Backup strategy verified
  - [ ] Disaster recovery plan reviewed

- [ ] **Communication**
  - [ ] Stakeholders notified of deployment window
  - [ ] Maintenance page ready (if needed)
  - [ ] Rollback team on standby
  - [ ] Incident response contacts verified

### Post-Deployment Checklist

- [ ] **Health Verification**
  - [ ] Health check endpoints responding
  - [ ] All pods running and ready
  - [ ] No error spikes in logs
  - [ ] Metrics within normal range

- [ ] **Smoke Tests**
  - [ ] Critical user flows tested
  - [ ] API endpoints responding
  - [ ] Database connections working
  - [ ] External integrations functional

- [ ] **Monitoring**
  - [ ] Metrics dashboard reviewed
  - [ ] Error rates normal
  - [ ] Latency within SLA
  - [ ] Resource usage acceptable

- [ ] **Documentation**
  - [ ] Deployment recorded in tracking system
  - [ ] Release notes published
  - [ ] Known issues documented
  - [ ] Rollback plan confirmed

- [ ] **Communication**
  - [ ] Stakeholders notified of completion
  - [ ] Support team briefed on changes
  - [ ] Status page updated

### Deployment Success Metrics

| Metric | Target | Critical Threshold |
|--------|--------|--------------------|
| Deployment Frequency | > 10/week | < 1/week |
| Lead Time for Changes | < 1 hour | > 1 day |
| Mean Time to Recovery (MTTR) | < 15 minutes | > 1 hour |
| Change Failure Rate | < 5% | > 15% |
| Deployment Success Rate | > 95% | < 85% |
| Rollback Rate | < 5% | > 15% |
| P95 Deployment Duration | < 30 minutes | > 2 hours |

---

## Summary

This deployment guide covers the complete lifecycle of deploying MoAI-ADK projects:

1. **Containerization**: Docker multi-stage builds with security best practices
2. **Orchestration**: Kubernetes, Docker Compose, and cloud-native solutions
3. **Cloud Platforms**: AWS ECS, Google Cloud Run, Azure Container Instances
4. **CI/CD**: GitHub Actions and GitLab CI/CD pipelines
5. **Environment Management**: Development, staging, and production configurations
6. **Rollout Strategies**: Blue-green, canary, and rolling deployments
7. **Monitoring**: Prometheus metrics, Grafana dashboards, structured logging
8. **Security**: Container scanning, network policies, OWASP compliance
9. **Automation**: Pipeline orchestration and deployment tracking
10. **Performance**: Database optimization, caching, and APM
11. **Recovery**: Rollback procedures and incident response
12. **Best Practices**: Comprehensive checklists and success metrics

By following these patterns and practices, teams can achieve reliable, secure, and efficient deployments that align with MoAI-ADK's SPEC-first philosophy and TRUST 5 principles.
