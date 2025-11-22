# AWS Advanced Patterns - Practical Examples

## Example 1: Multi-Region Active-Active Deployment

```yaml
# CloudFormation Template: Multi-Region Setup
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Multi-Region Active-Active Architecture'

Parameters:
  PrimaryRegion:
    Type: String
    Default: us-east-1
  SecondaryRegion:
    Type: String
    Default: eu-west-1

Resources:
  # Primary Region Resources
  PrimaryDB:
    Type: AWS::RDS::DBCluster
    Properties:
      Engine: aurora-mysql
      GlobalWriteForwardingEnabled: true
      BackupRetentionPeriod: 30
      AvailabilityZones:
        - us-east-1a
        - us-east-1b

  PrimaryS3Bucket:
    Type: AWS::S3::Bucket
    Properties:
      VersioningConfiguration:
        Status: Enabled
      ReplicationConfiguration:
        Role: !GetAtt S3ReplicationRole.Arn
        Rules:
          - Status: Enabled
            Filter:
              Prefix: ''
            Destination:
              Bucket: !Sub 'arn:aws:s3:::${SecondaryBucket}'
              ReplicationTime:
                Status: Enabled
                Time:
                  Minutes: 15
              Metrics:
                Status: Enabled

  # Route 53 Health Checks
  HealthCheck:
    Type: AWS::Route53::HealthCheck
    Properties:
      Type: HTTPS
      ResourcePath: /health
      FullyQualifiedDomainName: api.example.com
      Port: 443
      RequestInterval: 30
      FailureThreshold: 3

  DNSRecords:
    Type: AWS::Route53::RecordSet
    Properties:
      HostedZoneId: Z123ABC
      Name: api.example.com
      Type: A
      GeolocationLocation:
        ContinentCode: EU
      SetIdentifier: EU-Primary
      AliasTarget:
        HostedZoneId: Z456DEF
        DNSName: primary-alb.eu-west-1.elb.amazonaws.com
        EvaluateTargetHealth: true
```

## Example 2: Lambda + API Gateway Serverless Pattern

```python
# Advanced Serverless Architecture with Caching
import json
import boto3
import asyncio
from aws_lambda_powertools import Logger, Tracer
from aws_lambda_powertools.utilities.data_classes.api_gateway_event import APIGatewayProxyEvent
from aws_lambda_powertools.utilities.data_classes.common_model import CorrelationIdModel

logger = Logger()
tracer = Tracer()
dynamodb = boto3.resource('dynamodb')

@tracer.capture_lambda_handler
@logger.inject_lambda_context
async def handler(event: APIGatewayProxyEvent, context):
    """
    Advanced Lambda handler with:
    - X-Ray tracing
    - Structured logging
    - Error handling
    - DynamoDB integration
    """

    try:
        # Parse request
        path = event.path
        method = event.http_method
        correlation_id = event.get_header_value('x-correlation-id')

        logger.info(f"Processing {method} {path}", extra={"correlation_id": correlation_id})

        # Route handling
        if method == 'GET' and path.startswith('/api/users/'):
            user_id = path.split('/')[-1]
            return await get_user(user_id, correlation_id)

        elif method == 'POST' and path == '/api/users':
            body = json.loads(event.body)
            return await create_user(body, correlation_id)

        else:
            return {
                'statusCode': 404,
                'body': json.dumps({'error': 'Not found'})
            }

    except Exception as e:
        logger.exception(f"Error handling request", extra={"correlation_id": correlation_id})
        return {
            'statusCode': 500,
            'body': json.dumps({'error': 'Internal server error'})
        }

@tracer.capture_asynchronous_lambda_handler
async def get_user(user_id: str, correlation_id: str):
    """Fetch user with caching strategy"""
    table = dynamodb.Table('users-table')

    # Try cache first
    cache_key = f"user#{user_id}"

    response = await asyncio.get_event_loop().run_in_executor(
        None,
        table.get_item,
        {'Key': {'pk': cache_key}}
    )

    if 'Item' in response:
        logger.info(f"Cache hit for user {user_id}")
        return {
            'statusCode': 200,
            'headers': {'x-cache': 'HIT'},
            'body': json.dumps(response['Item'])
        }

    # Fetch from database
    logger.info(f"Cache miss, fetching from database for user {user_id}")

    return {
        'statusCode': 200,
        'headers': {'x-cache': 'MISS', 'cache-control': 'max-age=300'},
        'body': json.dumps({'id': user_id, 'name': 'User Name'})
    }

async def create_user(data: dict, correlation_id: str):
    """Create user with validation and DynamoDB write"""
    table = dynamodb.Table('users-table')

    # Validate
    if not all(k in data for k in ['name', 'email']):
        return {
            'statusCode': 400,
            'body': json.dumps({'error': 'Missing required fields'})
        }

    # Write with auto-generated ID
    user_id = f"user#{correlation_id[:16]}"

    await asyncio.get_event_loop().run_in_executor(
        None,
        table.put_item,
        {
            'Item': {
                'pk': user_id,
                'sk': data['email'],
                'name': data['name'],
                'created_at': int(asyncio.get_event_loop().time())
            }
        }
    )

    return {
        'statusCode': 201,
        'body': json.dumps({'id': user_id, **data})
    }
```

## Example 3: DynamoDB Optimization Patterns

```python
# DynamoDB with Optimized Query Patterns
import boto3
from boto3.dynamodb.conditions import Key, Attr
from decimal import Decimal
from typing import List, Dict, Optional

class DynamoDBOptimized:
    def __init__(self, table_name: str):
        self.dynamodb = boto3.resource('dynamodb')
        self.table = self.dynamodb.Table(table_name)

    def query_by_pk_sk(self, pk: str, sk_begins_with: str) -> List[Dict]:
        """
        Query with partition key and sort key
        Uses Global Secondary Index (GSI) for efficient filtering
        """
        response = self.table.query(
            KeyConditionExpression=Key('pk').eq(pk) & Key('sk').begins_with(sk_begins_with),
            ProjectionExpression='pk, sk, #n, #t',  # Only retrieve needed attributes
            ExpressionAttributeNames={'#n': 'name', '#t': 'timestamp'},
            Limit=100
        )
        return response['Items']

    def scan_with_filter(self, status: str) -> List[Dict]:
        """
        Scan with filter (less efficient but needed for non-key queries)
        Always use limit to prevent timeout
        """
        response = self.table.scan(
            FilterExpression=Attr('status').eq(status),
            Limit=1000,  # Prevent timeout
            ProjectionExpression='pk, sk, #s',
            ExpressionAttributeNames={'#s': 'status'}
        )
        return response['Items']

    def batch_write(self, items: List[Dict]) -> None:
        """
        Batch write for bulk insert (25 items max per batch)
        Handles retries and exponential backoff
        """
        with self.table.batch_writer(
            batch_size=25,
            overwrite_by_pkeys=['pk', 'sk']
        ) as batch:
            for item in items:
                batch.put_item(Item=item)

    def query_with_pagination(self, pk: str, page_size: int = 50):
        """Generator for paginated queries"""
        params = {
            'KeyConditionExpression': Key('pk').eq(pk),
            'Limit': page_size
        }

        while True:
            response = self.table.query(**params)

            for item in response['Items']:
                yield item

            # Check for more pages
            if 'LastEvaluatedKey' not in response:
                break

            params['ExclusiveStartKey'] = response['LastEvaluatedKey']

    def update_with_condition(self, pk: str, sk: str, updates: Dict) -> bool:
        """
        Update with condition expression to prevent overwriting recent changes
        Implements optimistic locking
        """
        try:
            self.table.update_item(
                Key={'pk': pk, 'sk': sk},
                UpdateExpression='SET #v = :val, #t = :now',
                ConditionExpression=Attr('version').eq(updates.get('version', 0)),
                ExpressionAttributeNames={'#v': 'value', '#t': 'updated_at'},
                ExpressionAttributeValues={
                    ':val': updates.get('value'),
                    ':now': int(asyncio.get_event_loop().time())
                },
                ReturnValues='ALL_NEW'
            )
            return True
        except self.dynamodb.meta.client.exceptions.ConditionalCheckFailedException:
            return False  # Version mismatch
```

## Example 4: RDS Aurora Performance Tuning

```sql
-- Aurora MySQL Performance Optimization

-- Enable Query Performance Insights
CALL mysql.rds_show_configuration('binlog retention hours');

-- Create optimized indexes
CREATE INDEX idx_user_email_status ON users(email, status) USING BTREE;
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE FULLTEXT INDEX idx_search ON articles(title, content);

-- Aurora-specific parameter group
CALL mysql.rds_set_configuration('binlog retention hours', 24);

-- Performance monitoring
SELECT *
FROM performance_schema.events_waits_summary_by_instance
WHERE SUM_TIMER_WAIT > 0
ORDER BY SUM_TIMER_WAIT DESC
LIMIT 20;

-- Identify slow queries
SELECT *
FROM performance_schema.events_statements_summary_by_digest
WHERE SUM_TIMER_WAIT > 1000000000000  -- > 1 second
ORDER BY SUM_TIMER_WAIT DESC;
```

## Example 5: Cost Optimization - Reserved Instances vs On-Demand

```python
# Cost Optimization Analysis
import boto3
import json
from datetime import datetime, timedelta

class CostOptimization:
    def __init__(self):
        self.ce = boto3.client('ce')  # Cost Explorer
        self.ec2 = boto3.client('ec2')

    def analyze_reservation_savings(self) -> Dict:
        """Calculate savings from Reserved Instances"""
        end_date = datetime.now().date()
        start_date = end_date - timedelta(days=90)

        # Get On-Demand costs
        response = self.ce.get_cost_and_usage(
            TimePeriod={
                'Start': start_date.isoformat(),
                'End': end_date.isoformat()
            },
            Granularity='MONTHLY',
            Metrics=['UnblendedCost'],
            Filter={
                'Dimensions': {
                    'Key': 'INSTANCE_TYPE',
                    'Values': ['t3.medium', 't3.large']
                }
            }
        )

        # Calculate 1-year and 3-year RI savings
        on_demand_monthly = float(response['ResultsByTime'][0]['Total']['UnblendedCost']['Amount'])

        return {
            'on_demand_monthly': on_demand_monthly,
            '1_year_ri_hourly': on_demand_monthly * 0.4,  # 60% savings
            '3_year_ri_hourly': on_demand_monthly * 0.25,  # 75% savings
            'annual_on_demand': on_demand_monthly * 12,
            'annual_1yr_ri': on_demand_monthly * 12 * 0.4,
            'annual_3yr_ri': on_demand_monthly * 12 * 0.25
        }

    def recommend_instance_mix(self) -> Dict:
        """Recommend optimal instance type mix"""
        instances = self.ec2.describe_instances()

        instance_usage = {}
        for reservation in instances['Reservations']:
            for instance in reservation['Instances']:
                itype = instance['InstanceType']
                instance_usage[itype] = instance_usage.get(itype, 0) + 1

        recommendations = {}
        for itype, count in instance_usage.items():
            # Recommend RI for stable baseline, On-Demand for spikes
            ri_count = max(int(count * 0.7), 1)
            on_demand = count - ri_count
            recommendations[itype] = {
                'total': count,
                'recommend_ri_1yr': ri_count,
                'keep_on_demand': on_demand,
                'estimated_monthly_savings': 500 * ri_count  # Example
            }

        return recommendations
```

## Example 6: EKS Cluster Security & Scaling

```yaml
# EKS Cluster with Security Best Practices
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: production-cluster
  region: us-east-1
  version: '1.28'

nodeGroups:
  - name: system-nodes
    minSize: 3
    maxSize: 10
    desiredCapacity: 5
    instanceType: t3.medium
    labels:
      workload: system
    taints:
      - key: system
        value: 'true'
        effect: NoSchedule
    iam:
      attachPolicy:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Action:
              - 'ecr:GetAuthorizationToken'
              - 'ecr:BatchGetImage'
              - 'ecr:GetDownloadUrlForLayer'
            Resource: '*'

  - name: application-nodes
    minSize: 5
    maxSize: 100
    desiredCapacity: 20
    instanceType: t3.large
    labels:
      workload: application
    tags:
      Environment: production

vpc:
  id: vpc-12345678
  subnets:
    private:
      - id: subnet-private-1
      - id: subnet-private-2
      - id: subnet-private-3

addons:
  - name: vpc-cni
    version: latest
    attachPolicy:
      Version: '2012-10-17'
      Statement:
        - Effect: Allow
          Action:
            - 'ec2:AssignPrivateIpAddresses'
            - 'ec2:AttachNetworkInterface'
          Resource: '*'

  - name: kube-proxy
    version: latest

  - name: coredns
    version: latest

logging:
  clusterLogging:
    - ['api', 'audit', 'authenticator', 'controllerManager', 'scheduler']
```

---

**Learn More**: See advanced-patterns.md for architectural deep-dives.
