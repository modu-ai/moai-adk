---
name: moai-cloud-aws-advanced
version: 3.0.0
updated: "2025-11-19"
status: stable
description: Advanced AWS container and serverless architectures for ECS/EKS orchestration, Lambda patterns, EventBridge workflows, and production deployments. Use when implementing AWS infrastructure, containers, or serverless.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
---

# Advanced AWS Architecture

Production-grade AWS patterns for containers (ECS/EKS) and serverless (Lambda/EventBridge).

## Quick Start

**Deploy Lambda Function**:

```bash
# Create function
aws lambda create-function \
  --function-name my-function \
  --runtime python3.11 \
  --handler lambda_function.lambda_handler \
  --role arn:aws:iam::ACCOUNT:role/lambda-role \
  --zip-file fileb://function.zip

# Invoke
aws lambda invoke \
  --function-name my-function \
  --payload '{"key":"value"}' \
  response.json
```

**Deploy ECS Service**:

```bash
# Create cluster
aws ecs create-cluster --cluster-name prod-cluster

# Register task definition
aws ecs register-task-definition --cli-input-json file://task-def.json

# Create service
aws ecs create-service \
  --cluster prod-cluster \
  --service-name api-service \
  --task-definition api:1 \
  --desired-count 3 \
  --launch-type FARGATE
```

---

## AWS Services Overview

### Compute Options

| Service         | Best For                   | Pricing       | Cold Start | Scale Time |
| --------------- | -------------------------- | ------------- | ---------- | ---------- |
| **Lambda**      | Event-driven, short tasks  | Per-request   | ~100-500ms | Instant    |
| **ECS Fargate** | Containers, no servers     | Per vCPU-hour | None       | 1-2 min    |
| **EKS**         | Kubernetes, complex apps   | EC2 + EKS fee | None       | 2-5 min    |
| **EC2**         | Full control, long-running | Per-hour      | None       | 1-2 min    |

**Selection Guide**:

- **Lambda**: APIs, data processing, cron jobs
- **ECS Fargate**: Microservices, no infrastructure management
- **EKS**: Multi-cloud, Kubernetes ecosystem
- **EC2**: Custom requirements, specific OS

---

## Lambda Best Practices

### Function Structure

```python
import json
import os
from aws_lambda_powertools import Logger, Tracer, Metrics
from aws_lambda_powertools.utilities.typing import LambdaContext

logger = Logger()
tracer = Tracer()
metrics = Metrics()

@logger.inject_lambda_context
@tracer.capture_lambda_handler
@metrics.log_metrics(capture_cold_start_metric=True)
def lambda_handler(event: dict, context: LambdaContext) -> dict:
    """
    Production Lambda with observability
    """
    # Log incoming event
    logger.info("Processing request", extra={"event": event})

    # Add custom metrics
    metrics.add_metric(name="ProcessedItems", unit="Count", value=1)

    try:
        # Business logic
        result = process_event(event)

        return {
            'statusCode': 200,
            'body': json.dumps(result)
        }
    except Exception as e:
        logger.exception("Processing failed")
        metrics.add_metric(name="Errors", unit="Count", value=1)
        raise
```

### Environment Variables & Secrets

```python
import boto3
from botocore.exceptions import ClientError

def get_secret(secret_name):
    """Get secret from AWS Secrets Manager"""
    client = boto3.client('secretsmanager')

    try:
        response = client.get_secret_value(SecretId=secret_name)
        return json.loads(response['SecretString'])
    except ClientError as e:
        logger.error(f"Failed to get secret: {e}")
        raise

# Use in Lambda
DB_PASSWORD = get_secret('prod/db/password')
```

### Lambda Layers

```bash
# Create layer
mkdir python
pip install requests -t python/
zip -r layer.zip python

# Publish layer
aws lambda publish-layer-version \
  --layer-name my-dependencies \
  --zip-file fileb://layer.zip \
  --compatible-runtimes python3.11

# Attach to function
aws lambda update-function-configuration \
  --function-name my-function \
  --layers arn:aws:lambda:us-east-1:ACCOUNT:layer:my-dependencies:1
```

---

## ECS/Fargate Architecture

### Task Definition

```json
{
  "family": "api-service",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "api",
      "image": "account.dkr.ecr.region.amazonaws.com/api:latest",
      "portMappings": [
        {
          "containerPort": 8000,
          "protocol": "tcp"
        }
      ],
      "environment": [{ "name": "ENV", "value": "production" }],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/api-service",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "api"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:8000/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

### Auto Scaling

```bash
# Register scalable target
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --resource-id service/prod-cluster/api-service \
  --scalable-dimension ecs:service:DesiredCount \
  --min-capacity 2 \
  --max-capacity 10

# Create scaling policy (target tracking)
aws application-autoscaling put-scaling-policy \
  --service-namespace ecs \
  --resource-id service/prod-cluster/api-service \
  --scalable-dimension ecs:service:DesiredCount \
  --policy-name cpu-scaling \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration file://scaling-policy.json
```

**scaling-policy.json**:

```json
{
  "TargetValue": 70.0,
  "PredefinedMetricSpecification": {
    "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
  },
  "ScaleOutCooldown": 60,
  "ScaleInCooldown": 300
}
```

---

## EKS (Kubernetes on AWS)

### Cluster Creation

```bash
# Create cluster with eksctl
eksctl create cluster \
  --name prod-cluster \
  --region us-east-1 \
  --version 1.28 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 5 \
  --managed

# Configure kubectl
aws eks update-kubeconfig --name prod-cluster --region us-east-1
```

### Deploying Applications

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
        - name: api
          image: account.dkr.ecr.region.amazonaws.com/api:latest
          ports:
            - containerPort: 8000
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          env:
            - name: ENV
              value: "production"
---
apiVersion: v1
kind: Service
metadata:
  name: api-service
spec:
  type: LoadBalancer
  selector:
    app: api
  ports:
    - port: 80
      targetPort: 8000
```

```bash
kubectl apply -f deployment.yaml
kubectl get services
```

---

## EventBridge Patterns

### Event-Driven Architecture

```python
import boto3
import json

eventbridge = boto3.client('events')

def publish_event(detail_type, detail):
    """Publish event to EventBridge"""
    response = eventbridge.put_events(
        Entries=[
            {
                'Source': 'com.myapp.orders',
                'DetailType': detail_type,
                'Detail': json.dumps(detail),
                'EventBusName': 'default'
            }
        ]
    )
    return response

# Publish order created event
publish_event(
    'OrderCreated',
    {
        'orderId': 'order-123',
        'customerId': 'customer-456',
        'total': 99.99
    }
)
```

### Event Rule

```bash
# Create rule
aws events put-rule \
  --name order-processing \
  --event-pattern '{
    "source": ["com.myapp.orders"],
    "detail-type": ["OrderCreated"]
  }'

# Add target (Lambda)
aws events put-targets \
  --rule order-processing \
  --targets "Id"="1","Arn"="arn:aws:lambda:region:account:function:process-order"
```

---

## Infrastructure as Code

### CDK Example (Python)

```python
from aws_cdk import (
    Stack,
    aws_lambda as lambda_,
    aws_apigateway as apigw,
    aws_dynamodb as dynamodb,
)

class ApiStack(Stack):
    def __init__(self, scope, id, **kwargs):
        super().__init__(scope, id, **kwargs)

        # DynamoDB table
        table = dynamodb.Table(
            self, "ItemsTable",
            partition_key=dynamodb.Attribute(
                name="id",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST
        )

        # Lambda function
        handler = lambda_.Function(
            self, "ApiHandler",
            runtime=lambda_.Runtime.PYTHON_3_11,
            code=lambda_.Code.from_asset("lambda"),
            handler="index.handler",
            environment={
                "TABLE_NAME": table.table_name
            }
        )

        # Grant permissions
        table.grant_read_write_data(handler)

        # API Gateway
        api = apigw.RestApi(self, "ItemsApi")
        items = api.root.add_resource("items")
        items.add_method("GET", apigw.LambdaIntegration(handler))
```

---

## Observability

### CloudWatch Insights

```bash
# Query Lambda logs
aws logs insights query \
  --log-group-name /aws/lambda/my-function \
  --start-time $(date -u -d '1 hour ago' +%s) \
  --end-time $(date -u +%s) \
  --query-string '
    fields @timestamp, @message
    | filter @message like /ERROR/
    | sort @timestamp desc
    | limit 20
  '
```

### X-Ray Tracing

```python
from aws_xray_sdk.core import xray_recorder
from aws_xray_sdk.core import patch_all

# Patch all supported libraries
patch_all()

@xray_recorder.capture('process_order')
def process_order(order_id):
    # Create subsegment
    subsegment = xray_recorder.begin_subsegment('validate')
    try:
        validate_order(order_id)
    finally:
        xray_recorder.end_subsegment()

    # Another subsegment
    with xray_recorder.capture('charge_payment'):
        charge_payment(order_id)
```

---

## Cost Optimization

### Lambda

- Use ARM64 (Graviton2) for 20% cost savings
- Right-size memory (affects CPU)
- Use Lambda reserved concurrency
- Implement caching

### ECS/EKS

- Use Spot instances for 70% savings
- Fargate Spot for batch workloads
- Right-size containers (CPU/memory)
- Use Savings Plans

---

## Security Best Practices

✅ **DO**:

- Use IAM roles (not access keys)
- Enable VPC for Lambda
- Use Secrets Manager for credentials
- Implement least privilege
- Enable encryption at rest
- Use Security Groups properly

❌ **DON'T**:

- Hardcode secrets
- Use `*` in IAM policies
- Deploy without VPC (for sensitive workloads)
- Skip CloudTrail logging

---

## Advanced Topics

For detailed patterns:

- **[examples.md](examples.md)**: Complete architectures, CI/CD pipelines, multi-region setups
- **[reference.md](reference.md)**: AWS CLI commands, CDK patterns, cost optimization

**Related Skills**:

- `moai-cloud-gcp-advanced`: GCP patterns
- `moai-essentials-perf`: Performance optimization
- `moai-security-devsecops`: DevSecOps practices

---

**Services**: Lambda, ECS, EKS, EventBridge, CloudWatch, X-Ray, CDK

**Version**: 3.0.0  
**Last Updated**: 2025-11-19  
**Status**: Production Ready
