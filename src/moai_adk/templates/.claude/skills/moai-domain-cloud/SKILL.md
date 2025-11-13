---
name: "moai-domain-cloud"
version: "4.0.0"
created: "2025-11-12"
updated: "2025-11-12"
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

**Infrastructure as Code (AWS CDK v2.223.0):**

```python
# lib/serverless_stack.py
from aws_cdk import (
    Stack,
    aws_lambda as _lambda,
    aws_iam as iam,
    aws_s3 as s3,
    aws_s3_notifications as s3_notifications,
    Duration
)
from constructs import Construct

class ServerlessStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        
        # S3 bucket for data storage
        bucket = s3.Bucket(
            self, "DataBucket",
            versioned=True,
            encryption=s3.BucketEncryption.S3_MANAGED,
            block_public_access=s3.BlockPublicAccess.BLOCK_ALL,
            removal_policy=RemovalPolicy.DESTROY
        )
        
        # Lambda function with Python 3.13
        lambda_function = _lambda.Function(
            self, "DataProcessor",
            runtime=_lambda.Runtime.PYTHON_3_13,
            handler="handler.lambda_handler",
            code=_lambda.Code.from_asset("lambda"),
            timeout=Duration.minutes(5),
            memory_size=256,
            environment={
                "LOG_LEVEL": "INFO",
                "POWERTOOLS_SERVICE_NAME": "data-processor"
            }
        )
        
        # Grant permissions
        bucket.grant_read(lambda_function)
        lambda_function.add_to_role_policy(
            iam.PolicyStatement(
                effect=iam.Effect.ALLOW,
                actions=[
                    "logs:CreateLogGroup",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                ],
                resources=["arn:aws:logs:*:*:*"]
            )
        )
        
        # S3 event notification
        bucket.add_event_notification(
            s3.EventType.OBJECT_CREATED,
            s3_notifications.LambdaDestination(lambda_function)
        )
```

---

#### Pattern 2: Multi-Cloud Kubernetes with Terraform

**Problem**: Deploy consistent Kubernetes clusters across AWS, GCP, and Azure with unified networking and observability.

**Solution**: Use Terraform modules with cloud-specific implementations.

```hcl
# terraform/modules/kubernetes-cluster/main.tf
variable "cloud_provider" {
  description = "Cloud provider: aws, gcp, or azure"
  type        = string
}

variable "cluster_name" {
  description = "Name of the Kubernetes cluster"
  type        = string
}

variable "region" {
  description = "Cloud region"
  type        = string
}

# AWS EKS Cluster
resource "aws_eks_cluster" "main" {
  count = var.cloud_provider == "aws" ? 1 : 0
  
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster[0].arn
  version  = "1.34"
  
  vpc_config {
    subnet_ids = var.subnet_ids
  }
  
  depends_on = [
    aws_iam_role_policy_attachment.cluster_policy[0]
  ]
}

# GKE Cluster
resource "google_container_cluster" "main" {
  count = var.cloud_provider == "gcp" ? 1 : 0
  
  name               = var.cluster_name
  location           = var.region
  initial_node_count = 1
  
  remove_default_node_pool = true
  min_master_version      = "1.34"
  
  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {
    cluster_secondary_range_name = "pods"
    services_secondary_range_name = "services"
  }
}

# Azure AKS Cluster
resource "azurerm_kubernetes_cluster" "main" {
  count = var.cloud_provider == "azure" ? 1 : 0
  
  name                = var.cluster_name
  location            = var.region
  resource_group_name = var.resource_group_name
  dns_prefix          = "${var.cluster_name}-dns"
  
  kubernetes_version = "1.34.0"
  
  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_D2s_v3"
  }
  
  identity {
    type = "SystemAssigned"
  }
}

# Output cluster connection details
output "cluster_endpoint" {
  value = var.cloud_provider == "aws" ? aws_eks_cluster.main[0].endpoint :
         var.cloud_provider == "gcp" ? google_container_cluster.main[0].endpoint :
         azurerm_kubernetes_cluster.main[0].fqdn
}

output "cluster_ca_certificate" {
  value = var.cloud_provider == "aws" ? aws_eks_cluster.main[0].certificate_authority[0].data :
         var.cloud_provider == "gcp" ? google_container_cluster.main[0].master_auth[0].cluster_ca_certificate :
         azurerm_kubernetes_cluster.main[0].kube_config[0].cluster_ca_certificate
}
```

**Kubernetes Deployment for Multi-Cloud:**

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  labels:
    app: webapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:1.27
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: webapp-service
spec:
  selector:
    app: webapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: LoadBalancer
```

---

#### Pattern 3: Cloud-Native Database with AWS RDS PostgreSQL 17

**Problem**: Need scalable, highly available database with automated backups, monitoring, and security.

**Solution**: AWS RDS with PostgreSQL 17 and enhanced monitoring.

```python
# lib/database_stack.py
from aws_cdk import (
    Stack,
    aws_rds as rds,
    aws_ec2 as ec2,
    aws_secretsmanager as secretsmanager,
    RemovalPolicy
)
from constructs import Construct

class DatabaseStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, vpc, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)
        
        # Database security group
        db_security_group = ec2.SecurityGroup(
            self, "DatabaseSecurityGroup",
            vpc=vpc,
            description="Security group for RDS database",
            allow_all_outbound=False
        )
        
        # Database credentials secret
        db_secret = secretsmanager.Secret(
            self, "DatabaseSecret",
            secret_name="database-credentials",
            description="Database credentials for application"
        )
        
        # RDS PostgreSQL 17 instance
        database = rds.DatabaseInstance(
            self, "ApplicationDatabase",
            engine=rds.DatabaseInstanceEngine.postgres(
                version=rds.PostgresEngineVersion.VER_17
            ),
            instance_type=ec2.InstanceType("db.t3.micro"),
            vpc=vpc,
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            security_groups=[db_security_group],
            database_name="appdb",
            credentials=rds.Credentials.from_secret(db_secret),
            backup_retention=Duration.days(7),
            deletion_protection=False,
            removal_policy=RemovalPolicy.DESTROY,
            monitoring_interval=Duration.seconds(60),
            enable_performance_insights=True,
            performance_insight_retention=rds.PerformanceInsightRetention.DEFAULT
        )
        
        # Export database connection details
        self.database_secret = db_secret
        self.database_instance = database
```

---

### Level 3: Advanced Integration

#### Multi-Cloud Cost Optimization Strategy

```python
# cost_optimizer.py
import boto3
import google.cloud
from azure.mgmt.cost_management import CostManagementClient
from datetime import datetime, timedelta

class MultiCloudCostOptimizer:
    """Optimize costs across AWS, GCP, and Azure."""
    
    def __init__(self):
        self.aws_client = boto3.client('ce')
        self.gcp_client = google.cloud.billing.BudgetServiceClient()
        self.azure_client = CostManagementClient()
    
    def analyze_aws_costs(self, start_date, end_date):
        """Analyze AWS costs by service and region."""
        response = self.aws_client.get_cost_and_usage(
            TimePeriod={
                'Start': start_date,
                'End': end_date
            },
            Granularity='MONTHLY',
            Metrics=['BlendedCost'],
            GroupBy=[
                {'Type': 'DIMENSION', 'Key': 'SERVICE'},
                {'Type': 'DIMENSION', 'Key': 'REGION'}
            ]
        )
        
        return self._process_cost_data(response['ResultsByTime'])
    
    def optimize_aws_resources(self):
        """Provide AWS-specific cost optimization recommendations."""
        recommendations = []
        
        # Lambda optimization
        recommendations.append({
            'service': 'Lambda',
            'suggestion': 'Use provisioned concurrency for predictable workloads',
            'potential_savings': '20-30%'
        })
        
        # RDS optimization
        recommendations.append({
            'service': 'RDS',
            'suggestion': 'Enable serverless for bursty workloads',
            'potential_savings': '40-60%'
        })
        
        # EC2 optimization
        recommendations.append({
            'service': 'EC2',
            'suggestion': 'Use Spot instances for fault-tolerant workloads',
            'potential_savings': '70-90%'
        })
        
        return recommendations
```
