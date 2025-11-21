# Cloud Architecture Examples

## Example 1: AWS Lambda with Lambda Powertools

**Production Lambda handler with full observability**:

```python
from aws_lambda_powertools import Logger, Tracer, Metrics
from aws_lambda_powertools.utilities.data_classes.s3_event import S3Event
from aws_lambda_powertools.utilities.batch import BatchProcessor, EventType
from aws_lambda_powertools.utilities.batch.exceptions import BatchProcessingError
import json
import boto3

logger = Logger()
tracer = Tracer()
metrics = Metrics()
batch_processor = BatchProcessor(event_type=EventType.SQSDataClass)
s3_client = boto3.client('s3')

@tracer.capture_lambda_handler
@logger.inject_lambda_context
@metrics.log_cold_start_metric
def lambda_handler(event: S3Event, context):
    """Production Lambda handler with full observability."""
    for record in event.records:
        batch_processor.add_task(process_s3_object, record=record)

    try:
        results = batch_processor.run()
    except BatchProcessingError as e:
        logger.exception("Batch processing failed")
        metrics.add_metric(name="ProcessingErrors", unit="Count", value=len(e.failed_messages))

    metrics.publish_stored_metrics()
    return {"batchItemFailures": batch_processor.fail_messages}

@tracer.capture_function_handler
def process_s3_object(record):
    """Process individual S3 object with tracing."""
    bucket = record.s3.bucket.name
    key = record.s3.object.key
    logger.info(f"Processing {bucket}/{key}")
    return {"statusCode": 200, "key": key}
```

---

## Example 2: AWS CDK Infrastructure as Code

**Multi-layer cloud infrastructure definition**:

```python
from aws_cdk import (
    aws_ec2 as ec2,
    aws_rds as rds,
    aws_s3 as s3,
    aws_cloudfront as cloudfront,
    core
)

class MultiLayerStack(core.Stack):
    """Complete cloud infrastructure with VPC, RDS, and CDN."""

    def __init__(self, scope: core.Construct, id: str, **kwargs):
        super().__init__(scope, id, **kwargs)

        # VPC with public and private subnets
        vpc = ec2.Vpc(self, "VPC",
            max_azs=3,
            nat_gateways=1,
            subnet_configuration=[
                ec2.SubnetConfiguration(
                    subnet_type=ec2.SubnetType.PUBLIC,
                    name="Public"
                ),
                ec2.SubnetConfiguration(
                    subnet_type=ec2.SubnetType.PRIVATE,
                    name="Private"
                )
            ]
        )

        # RDS PostgreSQL database
        db = rds.DatabaseInstance(self, "Database",
            engine=rds.DatabaseInstanceEngine.postgres(
                version=rds.PostgresEngineVersion.VER_15
            ),
            instance_type=ec2.InstanceType("t4g.micro"),
            vpc=vpc,
            allocated_storage=100,
            backup_retention=core.Duration.days(7),
            multi_az=True
        )

        # S3 bucket for static assets
        bucket = s3.Bucket(self, "Assets",
            block_public_access=s3.BlockPublicAccess.BLOCK_ALL,
            versioned=True,
            lifecycle_rules=[
                s3.LifecycleRule(
                    transitions=[
                        s3.Transition(
                            storage_class=s3.StorageClass.GLACIER,
                            transition_after=core.Duration.days(90)
                        )
                    ]
                )
            ]
        )

        # CloudFront distribution
        distribution = cloudfront.CloudFrontWebDistribution(
            self, "Distribution",
            origin_configs=[
                cloudfront.SourceConfiguration(
                    s3_origin_source=cloudfront.S3OriginConfig(
                        s3_bucket_source=bucket
                    ),
                    behaviors=[
                        cloudfront.Behavior(
                            is_default_behavior=True,
                            compress=True
                        )
                    ]
                )
            ]
        )
```

---

## Example 3: Kubernetes Deployment with Helm Charts

**Production-ready Kubernetes deployment**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-deployment
  labels:
    app: api
    version: v1
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
        version: v1
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - api
              topologyKey: kubernetes.io/hostname

      containers:
      - name: api
        image: myregistry.azurecr.io/api:v1.2.3
        imagePullPolicy: IfNotPresent

        ports:
        - name: http
          containerPort: 8080
          protocol: TCP

        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi

        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10

        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5

        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-url

        volumeMounts:
        - name: config
          mountPath: /etc/config
          readOnly: true

      volumes:
      - name: config
        configMap:
          name: app-config

---
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
  selector:
    app: api

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-deployment
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

---

## Example 4: Docker Multi-Stage Build

**Optimized container image for production**:

```dockerfile
# Stage 1: Build
FROM python:3.12-slim as builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install dependencies
COPY requirements.txt .
RUN pip install --user --no-cache-dir -r requirements.txt

# Stage 2: Runtime
FROM python:3.12-slim

WORKDIR /app

# Create non-root user
RUN useradd -m -u 1000 appuser

# Copy only necessary files from builder
COPY --from=builder /root/.local /home/appuser/.local
COPY --chown=appuser:appuser app/ /app/

# Set environment variables
ENV PATH=/home/appuser/.local/bin:$PATH \
    PYTHONUNBUFFERED=1 \
    PYTHONHASHSEED=random

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:8000/health')"

# Expose port
EXPOSE 8000

# Run application
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

---

## Example 5: Terraform Multi-Environment Setup

**Manage multiple cloud environments with Terraform**:

```hcl
# main.tf
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-lock"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Terraform   = "true"
      CreatedAt   = timestamp()
    }
  }
}

# VPC and networking
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true

  tags = {
    Name = "${var.environment}-vpc"
  }
}

# Public subnets
resource "aws_subnet" "public" {
  count                   = length(var.availability_zones)
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.environment}-public-${count.index + 1}"
  }
}

# Private subnets
resource "aws_subnet" "private" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 10)
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "${var.environment}-private-${count.index + 1}"
  }
}

# Auto Scaling Group
resource "aws_autoscaling_group" "main" {
  name                = "${var.environment}-asg"
  vpc_zone_identifier = aws_subnet.private[*].id
  min_size            = var.min_size
  max_size            = var.max_size
  desired_capacity    = var.desired_capacity

  launch_template {
    id      = aws_launch_template.main.id
    version = "$Latest"
  }

  tag {
    key                 = "Name"
    value               = "${var.environment}-asg-instance"
    propagate_at_launch = true
  }
}

# Launch template
resource "aws_launch_template" "main" {
  name_prefix   = "${var.environment}-"
  image_id      = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  user_data = base64encode(templatefile("${path.module}/user-data.sh", {
    environment = var.environment
  }))

  iam_instance_profile {
    name = aws_iam_instance_profile.main.name
  }

  monitoring {
    enabled = true
  }

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = "${var.environment}-instance"
    }
  }
}

# IAM role
resource "aws_iam_role" "main" {
  name = "${var.environment}-instance-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_instance_profile" "main" {
  name = "${var.environment}-instance-profile"
  role = aws_iam_role.main.name
}

# Security group
resource "aws_security_group" "main" {
  name        = "${var.environment}-sg"
  description = "Security group for ${var.environment}"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-sg"
  }
}
```

---

## Example 6: Cloud Monitoring with Prometheus

**Complete observability stack setup**:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: production
    environment: prod

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093

rule_files:
  - '/etc/prometheus/rules/*.yml'

scrape_configs:
  - job_name: 'kubernetes-apiservers'
    kubernetes_sd_configs:
    - role: endpoints
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

  - job_name: 'kubernetes-nodes'
    kubernetes_sd_configs:
    - role: node
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

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

---
# alerting-rules.yml
groups:
  - name: kubernetes-cluster
    interval: 30s
    rules:
    - alert: HighCPUUsage
      expr: 'node_cpu_seconds_total > 0.8'
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High CPU usage detected"
        description: "CPU usage is {{ $value }}%"

    - alert: PodCrashLooping
      expr: 'rate(kube_pod_container_status_restarts_total[1h]) > 5'
      for: 10m
      labels:
        severity: critical
      annotations:
        summary: "Pod {{ $labels.pod }} is crash looping"
        description: "Pod {{ $labels.namespace }}/{{ $labels.pod }} has restarted {{ $value }} times in the last hour"

    - alert: KubernetesNodeNotReady
      expr: 'kube_node_status_condition{condition="Ready",status="true"} == 0'
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Kubernetes Node not ready"
        description: "Node {{ $labels.node }} is not in Ready state"
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
