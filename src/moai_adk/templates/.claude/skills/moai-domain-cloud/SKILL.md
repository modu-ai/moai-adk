---
name: "moai-domain-cloud"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: domain
description: "Enterprise-grade cloud architecture expertise with production-ready patterns for AWS (Lambda 3.13, ECS/Fargate 1.4.0, RDS, CDK 2.223.0), GCP (Cloud Run Gen2, Cloud Functions 2nd gen, Cloud SQL), Azure (Functions v4, Container Apps, AKS), and multi-cloud orchestration (Terraform 1.9.8, Pulumi 3.x, Kubernetes 1.34). Covers serverless architectures, container orchestration, multi-cloud deployments, cloud-native databases, infrastructure automation, cost optimization, security patterns, and disaster recovery for 2025 stable versions."
allowed-tools: "Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "cloud-expert"
secondary-agents: [qa-validator, alfred, doc-syncer]
keywords: [cloud, AWS, GCP, Azure, Lambda, serverless, ECS, Kubernetes, Terraform, multi-cloud, IaC, cloud-native, database, DevOps]
tags: [domain-expert, 2025-stable]
orchestration: 
can_resume: true
typical_chain_position: "middle"
depends_on: []
---

# moai-domain-cloud — Enterprise Cloud Architecture (v4.0)

**Enterprise-Grade Cloud Architecture Expertise**

> **Primary Agent**: cloud-expert
> **Secondary Agents**: qa-validator, alfred, doc-syncer
> **Version**: 4.0.0 (2025 Stable)
> **Keywords**: AWS, GCP, Azure, Lambda, serverless, Kubernetes, Terraform, multi-cloud, IaC

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Enterprise-grade cloud architecture expertise with production-ready patterns for multi-cloud deployments, serverless computing, container orchestration, and infrastructure automation using 2025 stable versions.

**When to Use:**
- ✅ Deploying serverless applications (Lambda, Cloud Run, Azure Functions)
- ✅ Building multi-cloud architectures with unified tooling
- ✅ Orchestrating containers with Kubernetes across clouds
- ✅ Implementing infrastructure-as-code with Terraform/Pulumi
- ✅ Designing cloud-native database architectures
- ✅ Optimizing cloud costs and implementing cost controls
- ✅ Establishing cloud security, compliance, and disaster recovery
- ✅ Managing multi-cloud networking and service mesh
- ✅ Implementing cloud monitoring and observability
- ✅ Migrating workloads to cloud platforms

**Quick Start Pattern:**

```python
# AWS Lambda with Python 3.13 — Serverless Compute
import json
import boto3
from aws_lambda_powertools import Logger, Tracer
from aws_lambda_powertools.utilities.data_classes.api_gateway_event import APIGatewayProxyEvent
from aws_lambda_powertools.utilities.data_classes.common_http_response import Response

logger = Logger()
tracer = Tracer()
s3_client = boto3.client('s3')

@tracer.capture_lambda_handler
@logger.inject_lambda_context
def lambda_handler(event: APIGatewayProxyEvent, context) -> Response:
    """Production-ready Lambda handler with structured logging and tracing."""
    try:
        # Lambda Powertools automatically extracts data from event
        body = json.loads(event.body) if event.body else {}
        user_id = body.get('user_id')
        
        # Structured logging with context
        logger.info("Processing request", extra={"user_id": user_id})
        
        # S3 operation with tracing
        response = s3_client.get_object(Bucket='my-bucket', Key=f'user/{user_id}')
        data = json.load(response['Body'])
        
        return Response(
            status_code=200,
            body=json.dumps({"message": "Success", "data": data})
        )
    except Exception as e:
        logger.exception("Error processing request")
        return Response(
            status_code=500,
            body=json.dumps({"error": str(e)})
        )
```

**Core Technology Stack (2025 Stable):**
- **AWS**: Lambda (Python 3.13), ECS/Fargate (v1.4.0), RDS (PostgreSQL 17), CDK (2.223.0)
- **GCP**: Cloud Run (Gen2), Cloud Functions 2nd gen, Cloud SQL (PostgreSQL 17)
- **Azure**: Functions (v4), Container Apps, SQL Database, AKS (1.34.x)
- **Multi-Cloud IaC**: Terraform (1.9.8), Pulumi (3.205.0), Kubernetes (1.34), Docker (27.5.1)
- **Observability**: CloudWatch, Stackdriver, Application Insights, Prometheus, Grafana

---

### Level 2: Practical Implementation (Production Patterns)

#### Pattern 1: AWS Lambda with Python 3.13 & Lambda Powertools

**Problem**: Lambda functions need structured logging, distributed tracing, and environment-based configuration without boilerplate.

**Solution**: Use AWS Lambda Powertools for production-ready patterns.

```python
# requirements.txt
aws-lambda-powertools[all]==2.41.0

# handler.py
from aws_lambda_powertools import Logger, Tracer, Metrics
from aws_lambda_powertools.utilities.data_classes.s3_event import S3Event
from aws_lambda_powertools.utilities.batch import BatchProcessor, EventType
from aws_lambda_powertools.utilities.batch.exceptions import BatchProcessingError
import json

logger = Logger()
tracer = Tracer()
metrics = Metrics()
batch_processor = BatchProcessor(event_type=EventType.SQSDataClass)

@tracer.capture_lambda_handler
@logger.inject_lambda_context
@metrics.log_cold_start_metric
def s3_event_handler(event: S3Event, context):
    """Process S3 events with batch error handling."""
    for record in event.records:
        batch_processor.add_task(process_s3_object, record=record)
    
    try:
        results = batch_processor.run()
    except BatchProcessingError as e:
        logger.exception("Batch processing failed", extra={"failed": e.failed_messages})
        metrics.add_metric(name="ProcessingErrors", unit="Count", value=len(e.failed_messages))
    
    metrics.publish_stored_metrics()
    return {"batchItemFailures": batch_processor.fail_messages}

@tracer.capture_function_handler
def process_s3_object(record):
    """Process individual S3 object."""
    bucket = record.s3.bucket.name
    key = record.s3.object.key
    logger.info(f"Processing {bucket}/{key}")
    # Custom processing logic
    return {"statusCode": 200, "key": key}
```

**Key Benefits:**
- Structured JSON logging with context injection
- X-Ray distributed tracing integration
- Custom metrics with EMF (Embedded Metric Format)
- Batch processing with partial failure handling
- Cold start detection and optimization
- Dependency injection patterns

**When to Use:**
- Event-driven architectures (S3, DynamoDB, SQS)
- Real-time data processing pipelines
- API backends with Lambda integration
- Asynchronous job processing
- Cost-sensitive, serverless-first designs

**Production Checklist:**
- ✅ Structured logging (JSON with correlation IDs)
- ✅ Distributed tracing (X-Ray integration)
- ✅ Custom metrics (CloudWatch EMF)
- ✅ Error handling with partial failure
- ✅ Environment-based configuration
- ✅ Secrets management (AWS Secrets Manager)
- ✅ VPC configuration for database access
- ✅ Memory allocation tuning (128 MB - 10 GB)
- ✅ Timeout configuration (max 15 minutes)
- ✅ Concurrency limits and reserved concurrency

---

#### Pattern 2: AWS CDK 2.x Infrastructure as Code

**Problem**: CloudFormation templates are verbose; infrastructure code needs type safety and reusability.

**Solution**: Use AWS CDK 2.x for high-level infrastructure definitions.

```python
# requirements.txt
aws-cdk-lib==2.223.0
constructs>=10.0.0,<11.0.0

# stacks/api_stack.py
from aws_cdk import (
    Stack,
    aws_lambda as lambda_,
    aws_apigateway as apigw,
    aws_dynamodb as dynamodb,
    aws_logs as logs,
    Duration,
    RemovalPolicy,
)
from constructs import Construct

class APIStack(Stack):
    def __init__(self, scope: Construct, id: str, **kwargs):
        super().__init__(scope, id, **kwargs)
        
        # DynamoDB table
        table = dynamodb.Table(
            self, "UsersTable",
            partition_key=dynamodb.Attribute(
                name="user_id",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.DESTROY,  # ⚠️ Only for dev
            point_in_time_recovery=True,
        )
        
        # Lambda function
        handler = lambda_.Function(
            self, "APIHandler",
            runtime=lambda_.Runtime.PYTHON_3_13,
            code=lambda_.Code.from_asset("src/handlers"),
            handler="api_handler.handler",
            timeout=Duration.seconds(30),
            memory_size=512,
            environment={"USERS_TABLE": table.table_name},
            log_retention=logs.RetentionDays.ONE_WEEK,
        )
        
        # Grant Lambda read/write to DynamoDB
        table.grant_read_write_data(handler)
        
        # API Gateway
        api = apigw.RestApi(
            self, "UsersAPI",
            default_cors_preflight_options=apigw.CorsOptions(
                allow_origins=apigw.Cors.ALL_ORIGINS,
                allow_methods=apigw.Cors.ALL_METHODS,
            )
        )
        
        users_resource = api.root.add_resource("users")
        users_resource.add_method(
            "POST",
            apigw.LambdaIntegration(handler),
        )
        
        # Outputs
        self.api_url = api.url

# app.py
from aws_cdk import App
from stacks.api_stack import APIStack

app = App()
APIStack(app, "api-stack", env={"region": "us-east-1"})
app.synth()
```

**Deploy:**
```bash
cdk bootstrap aws://ACCOUNT/REGION
cdk deploy
```

**Key Benefits:**
- Type-safe infrastructure definitions
- Code reusability with constructs
- Automatic permission management (least privilege)
- Built-in best practices (encryption, logging)
- CloudFormation generation and validation

**Patterns:**
- Custom constructs for common architectures
- Cross-stack references for microservices
- Environment-specific configurations
- Automated testing with assertions CDK

---

#### Pattern 3: ECS/Fargate Container Deployment

**Problem**: Managing containers at scale requires orchestration, load balancing, and auto-scaling.

**Solution**: AWS ECS with Fargate for serverless container management.

```python
# AWS CDK for ECS Fargate cluster
from aws_cdk import (
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_ecs_patterns as ecs_patterns,
    aws_logs as logs,
)

class FargateStack(Stack):
    def __init__(self, scope: Construct, id: str, **kwargs):
        super().__init__(scope, id, **kwargs)
        
        # VPC
        vpc = ec2.Vpc(self, "EcsVpc", max_azs=3)
        
        # ECS Cluster
        cluster = ecs.Cluster(self, "EcsCluster", vpc=vpc)
        
        # Fargate Service with ALB
        service = ecs_patterns.ApplicationLoadBalancedFargateService(
            self, "Service",
            cluster=cluster,
            memory_limit_mib=512,
            cpu=256,  # 0.25 vCPU
            desired_count=3,
            public_load_balancer=True,
            task_image_options=ecs_patterns.ApplicationLoadBalancedTaskImageOptions(
                image=ecs.ContainerImage.from_registry(
                    "123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:latest"
                ),
                container_port=8000,
                environment={"LOG_LEVEL": "INFO"},
                log_driver=ecs.LogDriver.aws_logs(
                    log_retention=logs.RetentionDays.ONE_WEEK
                ),
            ),
        )
        
        # Auto-scaling
        service.service.auto_scale_task_count(
            min_capacity=3,
            max_capacity=10,
        ).scale_on_cpu_utilization(
            "cpu-scaling",
            target_utilization_percent=70,
        )
```

**CloudFormation Alternative:**
```yaml
Resources:
  ECSCluster:
    Type: AWS::ECS::Cluster
    Properties:
      ClusterName: production-cluster
      ClusterSettings:
        - Name: containerInsights
          Value: enabled

  FargateService:
    Type: AWS::ECS::Service
    Properties:
      Cluster: !Ref ECSCluster
      TaskDefinition: !Ref TaskDefinition
      LaunchType: FARGATE
      DesiredCount: 3
      PlatformVersion: LATEST
      NetworkConfiguration:
        AwsvpcConfiguration:
          Subnets:
            - !Ref SubnetA
            - !Ref SubnetB
          SecurityGroups:
            - !Ref SecurityGroup
          AssignPublicIp: DISABLED

  TaskDefinition:
    Type: AWS::ECS::TaskDefinition
    Properties:
      Family: my-app
      NetworkMode: awsvpc
      RequiresCompatibilities:
        - FARGATE
      Cpu: "256"
      Memory: "512"
      ExecutionRoleArn: !GetAtt ECSTaskExecutionRole.Arn
      ContainerDefinitions:
        - Name: app
          Image: 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:latest
          PortMappings:
            - ContainerPort: 8000
          LogConfiguration:
            LogDriver: awslogs
            Options:
              awslogs-group: /ecs/my-app
              awslogs-region: us-east-1
              awslogs-stream-prefix: ecs
```

**Features (Fargate Platform 1.4.0):**
- EFS volume support for persistent storage
- 20 GB ephemeral storage (default)
- Network performance improvements
- Cost optimization (pay-per-second billing)

---

#### Pattern 4: GCP Cloud Run Gen2 Deployment

**Problem**: Deploying containerized applications to GCP requires simplified deployment and scaling.

**Solution**: Use Cloud Run Gen2 for event-driven, auto-scaling containers.

```python
# main.py (Flask application)
from flask import Flask, request, jsonify
import functions_framework
from google.cloud import firestore

app = Flask(__name__)
db = firestore.Client()

@functions_framework.http
def hello_world(request):
    """HTTP Cloud Function."""
    request_json = request.get_json(silent=True)
    user_id = request_json.get('user_id') if request_json else None
    
    if user_id:
        doc = db.collection('users').document(user_id).get()
        return jsonify({"data": doc.to_dict()})
    
    return jsonify({"error": "user_id required"}), 400

# Deploy with gcloud
# gcloud run deploy hello-world \
#     --source . \
#     --runtime python312 \
#     --region us-central1 \
#     --platform managed \
#     --allow-unauthenticated

# Terraform deployment
# terraform {
#   required_providers {
#     google = {
#       source  = "hashicorp/google"
#       version = "~> 5.0"
#     }
#   }
# }
#
# resource "google_cloud_run_v2_service" "default" {
#   name     = "hello-world"
#   location = "us-central1"
#
#   template {
#     spec {
#       service_account = google_service_account.default.email
#       timeout_seconds = 30
#
#       containers {
#         image = "us-central1-docker.pkg.dev/PROJECT/REPO/hello-world:latest"
#
#         env {
#           name  = "LOG_LEVEL"
#           value = "INFO"
#         }
#
#         resources {
#           limits = {
#             cpu    = "2"
#             memory = "512Mi"
#           }
#         }
#       }
#     }
#
#     scaling {
#       min_instance_count = 0
#       max_instance_count = 100
#     }
#   }
# }
```

**Key Features:**
- Full Linux compatibility (not emulated)
- Automatic scaling (0 to 100+ instances)
- Concurrency controls (50-1000 per instance)
- Network file system mounting (NFS)
- Startup CPU boost for faster cold starts
- Multiple regions and auto-failover

---

#### Pattern 5: Azure Functions v4 Deployment

**Problem**: Building serverless functions on Azure requires Python 4.x runtime for modern async support.

**Solution**: Use Azure Functions v4 with Python 3.11+.

```python
# requirements.txt
azure-functions
azure-storage-blob
azure-identity

# function_app.py
import azure.functions as func
from azure.storage.blob import BlobClient
from azure.identity import DefaultAzureCredential
import logging

app = func.FunctionApp()

@app.function_name("ProcessBlob")
@app.blob_trigger(arg_name="myblob", path="input/{name}", connection="AzureWebJobsStorage")
def blob_processor(myblob: func.InputStream):
    """Process blob with managed identity."""
    try:
        logging.info(f"Processing blob: {myblob.name}")
        
        # Process blob content
        content = myblob.read()
        
        # Use managed identity for authentication
        credential = DefaultAzureCredential()
        blob_client = BlobClient(
            account_url="https://myaccount.blob.core.windows.net",
            container_name="output",
            blob_name=myblob.name,
            credential=credential
        )
        
        blob_client.upload_blob(content, overwrite=True)
        logging.info(f"Completed processing: {myblob.name}")
    except Exception as e:
        logging.error(f"Error: {str(e)}")
        raise

@app.route(route="http_trigger")
def http_trigger(req: func.HttpRequest) -> func.HttpResponse:
    """HTTP-triggered function."""
    try:
        req_body = req.get_json()
        user_id = req_body.get('user_id')
        return func.HttpResponse(
            f"User {user_id} processed",
            status_code=200
        )
    except ValueError:
        return func.HttpResponse("Invalid request body", status_code=400)

# Deploy with Azure CLI
# func azure functionapp publish myFunctionApp
```

**Bicep Infrastructure:**
```bicep
param location string = resourceGroup().location
param functionAppName string = 'myFunctionApp'

resource storageAccount 'Microsoft.Storage/storageAccounts@2023-01-01' = {
  name: 'storage${uniqueString(resourceGroup().id)}'
  location: location
  kind: 'StorageV2'
  sku: {
    name: 'Standard_LRS'
  }
  properties: {
    accessTier: 'Hot'
  }
}

resource appServicePlan 'Microsoft.Web/serverfarms@2023-01-01' = {
  name: '${functionAppName}-plan'
  location: location
  sku: {
    name: 'Y1'
    tier: 'Dynamic'
  }
}

resource functionApp 'Microsoft.Web/sites@2023-01-01' = {
  name: functionAppName
  location: location
  kind: 'functionapp'
  identity: {
    type: 'SystemAssigned'
  }
  properties: {
    serverFarmId: appServicePlan.id
    siteConfig: {
      appSettings: [
        {
          name: 'AzureWebJobsStorage'
          value: 'DefaultEndpointsProtocol=https;AccountName=${storageAccount.name};AccountKey=${storageAccount.listKeys().keys[0].value}'
        }
        {
          name: 'FUNCTIONS_WORKER_RUNTIME'
          value: 'python'
        }
      ]
    }
  }
}
```

**Key Features:**
- Python 3.11+ support (v4 runtime)
- Managed identities for authentication
- Durable Functions for stateful workflows
- Bindings for 200+ services
- Auto-scaling based on demand
- Premium plan for VNet integration

---

#### Pattern 6: Terraform for Multi-Cloud Infrastructure

**Problem**: Managing infrastructure across AWS, GCP, and Azure requires consistent configuration.

**Solution**: Use Terraform 1.9.8 for unified infrastructure management.

```hcl
# terraform 1.9.8

terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
  
  backend "s3" {
    bucket         = "terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}

# AWS Provider
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

# GCP Provider
provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# Azure Provider
provider "azurerm" {
  features {}
  
  subscription_id = var.azure_subscription_id
}

# Variables
variable "environment" {
  type    = string
  default = "prod"
}

variable "project_name" {
  type = string
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# AWS: Lambda function
resource "aws_lambda_function" "api" {
  filename         = "lambda.zip"
  function_name    = "${var.project_name}-api"
  role            = aws_iam_role.lambda_role.arn
  handler         = "index.handler"
  source_code_hash = filebase64sha256("lambda.zip")
  runtime         = "python3.13"
  timeout         = 30
  memory_size     = 512
  
  environment {
    variables = {
      ENVIRONMENT = var.environment
      PROJECT     = var.project_name
    }
  }
  
  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic
  ]
}

# AWS: IAM Role
resource "aws_iam_role" "lambda_role" {
  name = "${var.project_name}-lambda-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.lambda_role.name
}

# GCP: Cloud Run Service
resource "google_cloud_run_v2_service" "api" {
  name     = "${var.project_name}-api"
  location = var.gcp_region
  
  template {
    spec {
      service_account = google_service_account.cloud_run.email
      timeout_seconds = 30
      
      containers {
        image = "gcr.io/${var.gcp_project_id}/${var.project_name}:latest"
        
        env {
          name  = "ENVIRONMENT"
          value = var.environment
        }
        
        resources {
          limits = {
            cpu    = "2"
            memory = "512Mi"
          }
        }
      }
    }
    
    scaling {
      min_instance_count = 0
      max_instance_count = 100
    }
  }
}

# Azure: Function App
resource "azurerm_linux_function_app" "api" {
  name                = "${var.project_name}-func"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  
  service_plan_id = azurerm_service_plan.main.id
  
  storage_account_name       = azurerm_storage_account.main.name
  storage_account_access_key = azurerm_storage_account.main.primary_access_key
  
  site_config {
    application_stack {
      python_version = "3.11"
    }
  }
  
  app_settings = {
    ENVIRONMENT = var.environment
    PROJECT     = var.project_name
  }
}

# Outputs
output "aws_lambda_arn" {
  value = aws_lambda_function.api.arn
}

output "gcp_cloud_run_url" {
  value = google_cloud_run_v2_service.api.uri
}

output "azure_function_url" {
  value = "https://${azurerm_linux_function_app.api.default_hostname}"
}
```

**Best Practices:**
- Use remote state (S3 with locking, Terraform Cloud)
- Separate dev/staging/prod configurations
- Use modules for code reusability
- Implement pre-commit hooks for validation
- Version your provider versions explicitly
- Use workspaces for environment isolation

---

#### Pattern 7: Kubernetes Deployment with Helm

**Problem**: Deploying applications to Kubernetes requires complex YAML; upgrades and rollbacks need automation.

**Solution**: Use Helm 3.16 for package management and templating.

```yaml
# Chart.yaml
apiVersion: v2
name: my-app
description: Production-ready Kubernetes application
type: application
version: 1.0.0
appVersion: "1.0"

dependencies:
  - name: postgresql
    version: "12.5.0"
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled

# values.yaml
replicaCount: 3

image:
  repository: my-registry/my-app
  tag: "1.0.0"
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 80
  targetPort: 8000

ingress:
  enabled: true
  className: nginx
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
  targetMemoryUtilizationPercentage: 80

postgresql:
  enabled: true
  auth:
    username: app
    password: changeme
    database: app

# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "my-app.fullname" . }}
  labels:
    {{- include "my-app.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "my-app.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
      labels:
        {{- include "my-app.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "my-app.serviceAccountName" . }}
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: http
          containerPort: 8000
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          {{- toYaml .Values.resources | nindent 12 }}
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-url
        - name: ENVIRONMENT
          value: {{ .Values.environment }}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app.kubernetes.io/name
                  operator: In
                  values:
                  - {{ include "my-app.name" . }}
              topologyKey: kubernetes.io/hostname

# templates/hpa.yaml
{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "my-app.fullname" . }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "my-app.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
{{- end }}
```

**Deploy:**
```bash
helm repo add my-repo https://charts.example.com
helm repo update
helm install my-app my-repo/my-app -f values.yaml

# Upgrade
helm upgrade my-app my-repo/my-app -f values.yaml

# Rollback
helm rollback my-app 1
```

**Best Practices:**
- Use Helm charts for all deployments
- Version your charts with semantic versioning
- Implement pre-install hooks for migrations
- Use init containers for setup tasks
- Implement resource quotas and limits
- Use NetworkPolicies for security

---

#### Pattern 8: Cloud Database Architecture (Multi-Cloud)

**Problem**: Choosing and configuring cloud databases requires understanding trade-offs across AWS, GCP, Azure.

**Solution**: Implement cloud-native database patterns with replication and backups.

```python
# AWS RDS with PostgreSQL 17
# Terraform configuration
resource "aws_db_instance" "postgres" {
  allocated_storage    = 100
  db_name              = "appdb"
  engine               = "postgres"
  engine_version       = "17.1"  # Latest 17.x
  instance_class       = "db.t4g.medium"
  username             = "postgres"
  password             = random_password.db_password.result
  
  # High availability
  multi_az             = true
  publicly_accessible  = false
  
  # Backup strategy
  backup_retention_period = 30
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"
  
  # Security
  skip_final_snapshot       = false
  final_snapshot_identifier = "appdb-final-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"
  
  # Monitoring
  enabled_cloudwatch_logs_exports = ["postgresql"]
  monitoring_interval              = 60
  monitoring_role_arn             = aws_iam_role.rds_monitoring.arn
  
  # Performance Insights
  performance_insights_enabled = true
  
  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.default.name
  
  deletion_protection = true
  
  tags = {
    Name = "production-postgres"
  }
}

# GCP Cloud SQL with PostgreSQL 17
resource "google_sql_database_instance" "postgres" {
  name             = "production-postgres"
  database_version = "POSTGRES_17"
  region           = "us-central1"
  
  settings {
    tier = "db-custom-2-7680"  # 2vCPU, 7.5GB RAM
    
    # High availability
    availability_type = "REGIONAL"
    
    # Backup strategy
    backup_configuration {
      enabled                        = true
      binary_log_enabled             = false
      location                       = "us"
      backup_retention_settings {
        retained_backups = 30
        retention_unit   = "COUNT"
      }
      transaction_log_retention_days = 7
    }
    
    # Security
    database_flags {
      name  = "cloudsql_iam_authentication"
      value = "on"
    }
    
    # SSL enforcement
    require_ssl = true
    
    # Monitoring
    insights_config {
      query_insights_enabled  = true
      query_string_length     = 1024
      record_application_tags = true
    }
    
    # Maintenance
    maintenance_window {
      day          = 7  # Sunday
      hour         = 3
      update_track = "stable"
    }
  }
  
  deletion_protection = true
  
  depends_on = [google_service_networking_connection.private_vpc_connection]
}

# Azure SQL Database
resource "azurerm_mssql_server" "main" {
  name                         = "production-sql"
  resource_group_name          = azurerm_resource_group.main.name
  location                     = azurerm_resource_group.main.location
  version                      = "12.0"
  administrator_login          = "sqladmin"
  administrator_login_password = random_password.sql_password.result
  
  # Security
  minimum_tls_version            = "1.2"
  public_network_access_enabled  = false
  azuread_administrator {
    login_username = "AzureAD Admin"
    object_id      = data.azuread_client_config.current.object_id
  }
}

resource "azurerm_mssql_database" "main" {
  name           = "appdb"
  server_id      = azurerm_mssql_server.main.id
  collation      = "SQL_Latin1_General_CP1_CI_AS"
  license_type   = "LicenseIncluded"
  max_size_gb    = 250
  
  # Backup strategy
  long_term_retention_policy {
    weekly_retention  = "P4W"    # 4 weeks
    monthly_retention = "P12M"   # 12 months
  }
  
  # Short-term retention (automatic)
  short_term_retention_policy {
    retention_days = 30
  }
}
```

**Key Patterns:**
- Multi-AZ/region replication for HA
- Automated backup with point-in-time recovery
- Encryption at rest and in transit
- VPC/Private Link for network isolation
- Database activity monitoring (CloudTrail, audit logs)
- Read replicas for scaling read workloads
- Connection pooling for cost optimization

---

#### Pattern 9: Serverless Cost Optimization

**Problem**: Serverless architectures can incur unexpectedly high costs without proper optimization.

**Solution**: Implement cost controls and monitoring.

```python
# AWS Lambda Cost Optimization
import boto3
from datetime import datetime, timedelta

ce_client = boto3.client('ce')  # Cost Explorer

def analyze_lambda_costs():
    """Analyze Lambda costs and identify optimization opportunities."""
    
    # Get last 30 days of Lambda costs
    response = ce_client.get_cost_and_usage(
        TimePeriod={
            'Start': (datetime.now() - timedelta(days=30)).strftime('%Y-%m-%d'),
            'End': datetime.now().strftime('%Y-%m-%d')
        },
        Granularity='DAILY',
        Metrics=['UnblendedCost'],
        Filter={
            'Dimensions': {
                'Key': 'SERVICE',
                'Values': ['AWS Lambda']
            }
        },
        GroupBy=[
            {
                'Type': 'DIMENSION',
                'Key': 'FUNCTION_NAME'
            }
        ]
    )
    
    # Identify expensive functions
    for result in response['ResultsByTime']:
        for group in result['Groups']:
            function_name = group['Keys'][0]
            cost = float(group['Metrics']['UnblendedCost']['Amount'])
            
            if cost > 10:  # Alert on functions costing >$10/day
                print(f"High-cost function: {function_name} = ${cost:.2f}/day")
                
                # Get metrics
                cloudwatch = boto3.client('cloudwatch')
                metrics = cloudwatch.get_metric_statistics(
                    Namespace='AWS/Lambda',
                    MetricName='Duration',
                    Dimensions=[
                        {
                            'Name': 'FunctionName',
                            'Value': function_name
                        }
                    ],
                    StartTime=datetime.now() - timedelta(days=30),
                    EndTime=datetime.now(),
                    Period=86400,
                    Statistics=['Average', 'Maximum']
                )
                
                avg_duration = metrics['Datapoints'][0]['Average'] if metrics['Datapoints'] else 0
                print(f"  Average duration: {avg_duration:.0f}ms")
                print(f"  Optimization: Reduce memory → faster execution → lower cost")

# CloudWatch Dashboard for cost tracking
cf_template = {
    "AWSTemplateFormatVersion": "2010-09-09",
    "Resources": {
        "CostDashboard": {
            "Type": "AWS::CloudWatch::Dashboard",
            "Properties": {
                "DashboardName": "Serverless-Cost-Monitoring",
                "DashboardBody": {
                    "widgets": [
                        {
                            "type": "metric",
                            "properties": {
                                "metrics": [
                                    ["AWS/Lambda", "Duration", {"stat": "Average"}],
                                    [".", "Invocations", {"stat": "Sum"}],
                                    [".", "Errors", {"stat": "Sum"}]
                                ],
                                "period": 300,
                                "stat": "Average",
                                "region": "us-east-1",
                                "title": "Lambda Performance"
                            }
                        }
                    ]
                }
            }
        }
    }
}

# Cost optimization checklist
"""
✅ Lambda Memory Optimization
   - Start with 512 MB, measure duration
   - Increase memory to reduce execution time (CPU improves)
   - Formula: Cost = (GB × Duration × $0.0000166667)
   - Example: 512MB for 500ms = (0.5 × 0.5 × $0.0000166667) ≈ $0.000004

✅ Reserved Concurrency
   - Set reserved concurrency to prevent scaling surprises
   - Reduces per-invocation costs by ~40%

✅ Ephemeral Storage
   - Default 512 MB, can increase to 10 GB
   - Cache dependencies, avoid re-downloading (saves Lambda duration)

✅ Lambda@Edge for CloudFront
   - Reduce origin requests by caching/filtering
   - Pay per request, often cheaper than origin Lambda

✅ SQS/SNS for Decoupling
   - Process messages asynchronously
   - Distribute load evenly
   - Reduce idle Lambda runs

✅ ARM-based Graviton2
   - Use `arm64` architecture (half the price)
   - Same functionality, 20% better performance
"""
```

**Cost Monitoring:**
- Use AWS Cost Explorer for trend analysis
- Set up AWS Budgets alerts
- Implement X-Ray tracing to identify bottlenecks
- Use Lambda Insights for performance/cost correlation
- Regular cost reviews (weekly/monthly)

---

#### Pattern 10: Multi-Cloud Disaster Recovery

**Problem**: Failing over between cloud providers requires infrastructure, data sync, and validation.

**Solution**: Implement active-passive DR strategy.

```python
# Multi-Cloud DR Architecture
# Active: AWS
# Passive: GCP (standby)

# AWS Primary Database
# - PostgreSQL 17 with automated backups
# - Replication to GCP Cloud SQL via AWS DMS

resource "aws_dms_replication_instance" "main" {
  replication_instance_id   = "primary-to-gcp"
  replication_instance_class = "dms.c5.large"
  
  allocated_storage          = 100
  vpc_security_group_ids     = [aws_security_group.dms.id]
  publicly_accessible        = false
  multi_az                   = true
  
  engine_version             = "3.4.7"
  
  tags = {
    Name = "DR-Replication"
  }
}

resource "aws_dms_replication_task" "aws_to_gcp" {
  replication_instance_arn   = aws_dms_replication_instance.main.arn
  replication_task_id        = "aws-rds-to-gcp-cloudsql"
  migration_type             = "cdc"  # Change Data Capture
  source_engine_name         = "postgres"
  target_engine_name         = "postgres"
  
  source_endpoint_arn        = aws_dms_endpoint.source.arn
  target_endpoint_arn        = aws_dms_endpoint.target.arn
  
  table_mappings = jsonencode({
    rules = [
      {
        rule-type = "selection"
        rule-id   = "1"
        rule-name = "include-all-tables"
        object-locator = {
          schema-name = "%"
          table-name  = "%"
        }
        rule-action = "include"
      }
    ]
  })
  
  replication_task_settings = jsonencode({
    TargetMetadata = {
      ParallelLoadThreads       = 8
      BatchApplyEnabled         = true
      ParallelApplyThreads      = 4
      MaxFullLoadSubTasks       = 8
      TransactionConsistencyTimeout = 1
    }
    FullLoadSettings = {
      MaxFullLoadSubTasks       = 8
      TransactionConsistencyTimeout = 1
      CommitRate                = 50000
    }
  })
  
  start_replication_task      = true
  migration_type              = "cdc"
}

# Health Check and Failover Logic
import boto3
from typing import Dict, Tuple

class MultiCloudDR:
    def __init__(self):
        self.aws_client = boto3.client('rds', region_name='us-east-1')
        self.gcp_client = None  # GCP Cloud SQL client
    
    def check_primary_health(self) -> Dict:
        """Check AWS RDS primary database health."""
        try:
            response = self.aws_client.describe_db_instances(
                DBInstanceIdentifier='production-postgres'
            )
            
            db_instance = response['DBInstances'][0]
            
            return {
                'status': db_instance['DBInstanceStatus'],
                'healthy': db_instance['DBInstanceStatus'] == 'available',
                'endpoint': db_instance['Endpoint']['Address'],
                'cpu': db_instance.get('CPUUtilization', 0),
                'storage': db_instance.get('FreeStorageSpace', 0)
            }
        except Exception as e:
            return {'healthy': False, 'error': str(e)}
    
    def check_secondary_health(self) -> Dict:
        """Check GCP Cloud SQL secondary database health."""
        # Similar logic for GCP Cloud SQL
        pass
    
    def should_failover(self) -> bool:
        """Determine if failover is needed."""
        primary_health = self.check_primary_health()
        
        # Failover if primary is down or degraded
        return not primary_health.get('healthy', False)
    
    def execute_failover(self):
        """Switch traffic to secondary region."""
        # 1. Promote GCP replica to standalone
        # 2. Update Route53/Anycast DNS to GCP endpoint
        # 3. Update application configs
        # 4. Test connectivity
        # 5. Monitor for issues
        
        print("Failover initiated:")
        print("1. Promoting GCP Cloud SQL replica...")
        
        # Update DNS (Route53)
        route53 = boto3.client('route53')
        route53.change_resource_record_sets(
            HostedZoneId='Z123456789',
            ChangeBatch={
                'Changes': [
                    {
                        'Action': 'UPSERT',
                        'ResourceRecordSet': {
                            'Name': 'db.example.com',
                            'Type': 'CNAME',
                            'TTL': 60,  # Short TTL for fast failover
                            'ResourceRecords': [
                                {'Value': 'cloud-sql.googleapis.com'}
                            ]
                        }
                    }
                ]
            }
        )
        
        print("2. DNS updated to GCP endpoint")
        print("3. Application connections will re-route within 60 seconds")

# CloudWatch Alarm for automatic failover
{
    "AlarmName": "DR-Primary-Down",
    "MetricName": "DBInstanceHealthStatus",
    "Namespace": "AWS/RDS",
    "Statistic": "Average",
    "Period": 300,
    "EvaluationPeriods": 2,
    "Threshold": 1,
    "ComparisonOperator": "LessThanThreshold",
    "AlarmActions": [
        "arn:aws:sns:us-east-1:ACCOUNT:failover-topic"
    ]
}
```

**DR Strategy:**
- **RTO (Recovery Time Objective)**: < 5 minutes
- **RPO (Recovery Point Objective)**: < 1 minute
- Active-passive replication (AWS primary, GCP standby)
- Continuous data sync with AWS DMS
- Automated failover via CloudWatch alarms + SNS
- Regular failover drills (quarterly)
- Test restore procedures (monthly)

---

### Level 3: Deep Dives & Advanced Topics

#### Cold Start Optimization

**Lambda Cold Start Problem:**
- Container initialization: 0-50ms
- Runtime startup: 20-100ms
- Custom code initialization: Variable
- Total typical cold start: 100-500ms

**Solutions:**
1. **Provisioned Concurrency**: Keep 1-10 warm instances (~$0.03/hour per concurrency)
2. **Graviton2 Processors** (arm64): 20% faster cold starts
3. **Lambda Layers** for shared code: Faster initialization
4. **Container Image Optimization**: Minimal base image, pre-compiled dependencies
5. **Startup CPU Boost**: GCP Cloud Run feature for 80-90% cold start reduction

---

#### Network Design for Multi-Cloud

**VPN Connectivity:**
- AWS Site-to-Site VPN ↔ GCP Cloud VPN
- Azure ExpressRoute for high-bandwidth, low-latency
- Establish mesh networking with service discovery

**Multi-Cloud Load Balancing:**
- Global Load Balancer (GCP) + Route 53 (AWS)
- Health checks across regions
- Geo-routing for latency optimization

---

#### Compliance & Security Patterns

**SOC 2 Compliance:**
- Encryption at rest (AES-256) and in transit (TLS 1.2+)
- MFA for all administrative access
- Audit logging (CloudTrail, Cloud Logging, Azure Audit)
- Regular security assessments (annual)
- Incident response plan with documented procedures

**PCI-DSS for Payment Processing:**
- No storing of full credit card numbers
- Use tokenization services (AWS Payment Cryptography, Stripe, Square)
- VPC/subnet isolation for payment services
- Network segmentation (NACLs, security groups)

---

## 📚 Reference Links

### AWS Services
- **Lambda**: https://docs.aws.amazon.com/lambda/
- **ECS**: https://docs.aws.amazon.com/ecs/
- **RDS**: https://docs.aws.amazon.com/rds/
- **CDK**: https://docs.aws.amazon.com/cdk/
- **Lambda Powertools**: https://github.com/aws-powertools/powertools-lambda-python

### GCP Services
- **Cloud Run**: https://cloud.google.com/run/docs
- **Cloud Functions**: https://cloud.google.com/functions/docs
- **Cloud SQL**: https://cloud.google.com/sql/docs

### Azure Services
- **Azure Functions**: https://learn.microsoft.com/en-us/azure/azure-functions/
- **Azure Container Apps**: https://learn.microsoft.com/en-us/azure/container-apps/
- **Bicep**: https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/

### IaC Tools
- **Terraform**: https://www.terraform.io/docs
- **Pulumi**: https://www.pulumi.com/docs/
- **Kubernetes**: https://kubernetes.io/docs/
- **Helm**: https://helm.sh/docs/

### Community Resources
- **Cloud Architecture Patterns**: https://www.cloudarchitecturepatterns.com/
- **Well-Architected Frameworks**: AWS/GCP/Azure official guides
- **CNCF Landscape**: https://landscape.cncf.io/

---

## ✅ Summary

**moai-domain-cloud v4.0** provides production-ready patterns for:
1. Building serverless applications (Lambda, Cloud Run, Azure Functions)
2. Containerized deployments (ECS, GKE, AKS, Kubernetes)
3. Multi-cloud infrastructure (Terraform, Pulumi)
4. Cloud-native databases with HA/DR
5. Cost optimization and monitoring
6. Security, compliance, and disaster recovery

**Key Principles:**
- Use managed services (reduce operational burden)
- Implement auto-scaling (optimize costs)
- Monitor everything (observability)
- Plan for failure (disaster recovery)
- Automate deployments (GitOps, IaC)
- Version your infrastructure (Terraform/Pulumi state)

