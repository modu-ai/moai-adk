# Code Examples — moai-domain-cloud

## AWS Lambda with Lambda Powertools

```python
# requirements.txt
aws-lambda-powertools[all]==2.41.0
boto3==1.35.0

# handler.py
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

## AWS CDK Infrastructure

```python
# requirements.txt
aws-cdk-lib==2.223.0
constructs>=10.0.0

# app.py
from aws_cdk import App, Stack
from aws_cdk import aws_lambda as lambda_
from aws_cdk import aws_apigateway as apigw
from aws_cdk import aws_dynamodb as dynamodb
from aws_cdk import Duration, RemovalPolicy
from constructs import Construct

class APIStack(Stack):
    def __init__(self, scope: Construct, id: str, **kwargs):
        super().__init__(scope, id, **kwargs)
        
        # DynamoDB
        table = dynamodb.Table(
            self, "UsersTable",
            partition_key=dynamodb.Attribute(
                name="user_id",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.DESTROY,
            point_in_time_recovery=True,
        )
        
        # Lambda
        handler = lambda_.Function(
            self, "APIHandler",
            runtime=lambda_.Runtime.PYTHON_3_13,
            code=lambda_.Code.from_asset("src/handlers"),
            handler="api_handler.handler",
            timeout=Duration.seconds(30),
            memory_size=512,
            environment={"USERS_TABLE": table.table_name},
        )
        table.grant_read_write_data(handler)
        
        # API Gateway
        api = apigw.RestApi(self, "UsersAPI")
        users = api.root.add_resource("users")
        users.add_method("POST", apigw.LambdaIntegration(handler))

app = App()
APIStack(app, "api-stack")
app.synth()
```

## GCP Cloud Run Deployment

```python
# main.py
import functions_framework
from google.cloud import firestore
from flask import jsonify
import logging

db = firestore.Client()

@functions_framework.http
def hello_world(request):
    """HTTP Cloud Function."""
    try:
        request_json = request.get_json(silent=True)
        user_id = request_json.get('user_id') if request_json else None
        
        if user_id:
            doc = db.collection('users').document(user_id).get()
            return jsonify({"data": doc.to_dict()})
        
        return jsonify({"error": "user_id required"}), 400
    except Exception as e:
        logging.error(str(e))
        return jsonify({"error": "Internal error"}), 500
```

## Terraform Multi-Cloud

```hcl
terraform {
  required_version = ">= 1.9.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_lambda_function" "api" {
  filename         = "lambda.zip"
  function_name    = "my-api"
  role            = aws_iam_role.lambda_role.arn
  handler         = "index.handler"
  runtime         = "python3.13"
  timeout         = 30
  memory_size     = 512
}

resource "aws_iam_role" "lambda_role" {
  name = "lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}
```

## Azure Functions v4

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
@app.blob_trigger(arg_name="myblob", path="input/{name}", 
                  connection="AzureWebJobsStorage")
def blob_processor(myblob: func.InputStream):
    """Process blob with managed identity."""
    try:
        logging.info(f"Processing: {myblob.name}")
        content = myblob.read()
        
        credential = DefaultAzureCredential()
        blob_client = BlobClient(
            account_url="https://myaccount.blob.core.windows.net",
            container_name="output",
            blob_name=myblob.name,
            credential=credential
        )
        blob_client.upload_blob(content, overwrite=True)
    except Exception as e:
        logging.error(str(e))
        raise
```

## Kubernetes Helm Chart

```yaml
# Chart.yaml
apiVersion: v2
name: my-app
version: 1.0.0
appVersion: "1.0"

# values.yaml
replicaCount: 3
image:
  repository: my-registry/my-app
  tag: "1.0.0"

# templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "my-app.fullname" . }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ include "my-app.name" . }}
  template:
    metadata:
      labels:
        app: {{ include "my-app.name" . }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        ports:
        - containerPort: 8000
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

## PostgreSQL RDS Setup (Terraform)

```hcl
resource "aws_db_instance" "postgres" {
  allocated_storage    = 100
  db_name              = "appdb"
  engine               = "postgres"
  engine_version       = "17.1"
  instance_class       = "db.t4g.medium"
  username             = "postgres"
  password             = random_password.db.result
  
  multi_az             = true
  backup_retention_period = 30
  skip_final_snapshot  = false
  
  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.default.name
  
  deletion_protection = true
}
```

## Multi-Cloud Failover

```python
import boto3
from datetime import datetime, timedelta

class MultiCloudDR:
    def __init__(self):
        self.aws_rds = boto3.client('rds')
        self.route53 = boto3.client('route53')
    
    def check_primary_health(self):
        """Check AWS RDS health."""
        try:
            response = self.aws_rds.describe_db_instances(
                DBInstanceIdentifier='primary-db'
            )
            db = response['DBInstances'][0]
            return {'healthy': db['DBInstanceStatus'] == 'available'}
        except:
            return {'healthy': False}
    
    def failover_to_secondary(self):
        """Switch to GCP Cloud SQL."""
        self.route53.change_resource_record_sets(
            HostedZoneId='Z123456789',
            ChangeBatch={
                'Changes': [{
                    'Action': 'UPSERT',
                    'ResourceRecordSet': {
                        'Name': 'db.example.com',
                        'Type': 'CNAME',
                        'TTL': 60,
                        'ResourceRecords': [{
                            'Value': 'cloud-sql.googleapis.com'
                        }]
                    }
                }]
            }
        )
```

## Cost Analysis (AWS)

```python
import boto3

def analyze_lambda_costs():
    """Analyze Lambda costs."""
    ce = boto3.client('ce')
    response = ce.get_cost_and_usage(
        TimePeriod={
            'Start': (datetime.now() - timedelta(days=30)).strftime('%Y-%m-%d'),
            'End': datetime.now().strftime('%Y-%m-%d')
        },
        Granularity='DAILY',
        Metrics=['UnblendedCost'],
        Filter={'Dimensions': {'Key': 'SERVICE', 'Values': ['AWS Lambda']}},
        GroupBy=[{'Type': 'DIMENSION', 'Key': 'FUNCTION_NAME'}]
    )
    
    for result in response['ResultsByTime']:
        for group in result['Groups']:
            name = group['Keys'][0]
            cost = float(group['Metrics']['UnblendedCost']['Amount'])
            if cost > 10:
                print(f"High-cost function: {name} = ${cost:.2f}/day")
```

## Container Image Optimization (Dockerfile)

```dockerfile
# Multi-stage build for minimal image
FROM python:3.13-slim as builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.13-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY . .

RUN addgroup --system app && adduser --system app --ingroup app
USER app

EXPOSE 8000
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0"]
```

## Environment-Based Configuration

```python
# config.py
import os
from dataclasses import dataclass

@dataclass
class CloudConfig:
    environment: str = os.getenv("ENVIRONMENT", "dev")
    database_url: str = os.getenv("DATABASE_URL", "postgresql://localhost/app")
    log_level: str = os.getenv("LOG_LEVEL", "INFO")
    debug: bool = os.getenv("DEBUG", "false").lower() == "true"

config = CloudConfig()

# Usage
if config.debug:
    logging.basicConfig(level=logging.DEBUG)
```

