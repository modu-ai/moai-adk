# AWS Advanced Services - API Reference

## EC2 Instance Types Quick Reference

```yaml
General Purpose (T3, M6, M5):
  - t3.nano: 0.5 GB RAM, 2 vCPU - Burstable, free tier
  - m6i.large: 8 GB RAM, 2 vCPU - Baseline performance
  - m6i.2xlarge: 32 GB RAM, 8 vCPU - Production workloads
  Cost: $0.0116 - $0.69/hour

Compute Optimized (C5, C6):
  - c6i.large: 4 GB RAM, 2 vCPU - Web servers, batch processing
  - c6i.4xlarge: 32 GB RAM, 16 vCPU - High-performance computing
  Cost: $0.085 - $2.66/hour

Memory Optimized (R5, R6, X2):
  - r6i.large: 16 GB RAM, 2 vCPU - Databases, in-memory caches
  - r6i.4xlarge: 128 GB RAM, 16 vCPU - SAP HANA, real-time analytics
  Cost: $0.126 - $4.03/hour

Storage Optimized (I3, D2):
  - i3.large: NVMe SSD, 160 GB - NoSQL databases
  - d2.8xlarge: HDD, 336 GB - Data warehousing
  Cost: $0.591 - $6.048/hour
```

## RDS Database Engine Comparison

```yaml
MySQL/MariaDB:
  - Compatibility: High with standard SQL
  - Performance: Good for general workloads
  - Cost: $0.20-3.50/hour
  - Best for: Web applications, CMSs

PostgreSQL:
  - Compatibility: Enterprise-grade features
  - Performance: Excellent for complex queries
  - Cost: $0.20-3.50/hour
  - Best for: Data warehousing, analytics

Oracle:
  - Compatibility: Enterprise applications
  - Performance: Optimized for large databases
  - Cost: $2-30+/hour
  - Best for: Legacy migration, enterprise apps

Aurora MySQL/PostgreSQL:
  - Compatibility: Drop-in replacement
  - Performance: 3x MySQL, 2x PostgreSQL
  - Cost: $1.25-5.00/hour
  - Best for: High-performance, scalability
```

## Lambda Pricing Calculator

```
Request Charges:
  1M requests/month = $0.20
  10M requests/month = $2.00
  1B requests/month = $200.00

Duration Charges:
  512 MB, 1 second = 512 × 0.0000166667 = $0.0000085
  1024 MB, 1 second = 1024 × 0.0000166667 = $0.000017

Example: 10M requests, 512MB, 500ms average
  Monthly cost = $2 (requests) + (10M × 0.5s × 512MB × 0.0000166667) = $2 + $42 = $44

Always: First 1M requests free, first 400K GB-seconds free per month
```

## DynamoDB Capacity Planning

```yaml
Provisioned Mode:
  Write Capacity Unit (WCU): 1KB/sec
  Read Capacity Unit (RCU): 4KB/sec (eventually consistent)

  Example: 100 WCU, 100 RCU
  Cost: (100 × $1.25) + (100 × $0.25) = $150/month

On-Demand Mode:
  Write: $1.25 per 1M writes
  Read: $0.25 per 1M reads

  Example: 1B writes, 1B reads/month
  Cost: $1,250 + $250 = $1,500/month

  Break-even: ~830M requests/month
```

## S3 Storage Pricing

```yaml
Standard:
  Storage: $0.023/GB (first 50TB/month)
  PUT: $0.005 per 1000 requests
  GET: $0.0004 per 1000 requests

Standard-IA:
  Storage: $0.0125/GB
  Retrieval: $0.01 per 1000 requests
  Minimum storage: 30 days

Glacier:
  Storage: $0.004/GB
  Retrieval: Expedited ($0.03), Standard ($0.01), Bulk ($0.0025)
  Minimum storage: 90 days

Cost Example: 100TB, 10M requests/month
  Standard: $2,300 + $50 + $4 = $2,354
  Glacier (8M archived): $320 + $80 = $400
  Savings with tiering: ~83%
```

## VPC & Networking

### VPC Endpoints

```yaml
Gateway Endpoints (Free):
  - S3
  - DynamoDB
  - EC2
  - SNS
  - SQS

Interface Endpoints:
  Cost: $0.01/hour + $0.01/GB processed
  For: Hundreds of services (SNS, RDS, etc.)

NAT Gateway:
  Cost: $0.045/hour + $0.045/GB
  Alternative: NAT instance (cheaper but manual)
```

### Route 53

```yaml
Hosted Zones: $0.50 per zone/month
Standard Queries: $0.40 per million
Health Checks: $0.50 per health check/month
Geolocation Routing: No additional cost
Traffic Policy: $50 per policy/month
```

## CloudFront & Content Delivery

```yaml
Data Out Costs (varies by region):
  US/EU: $0.085/GB
  Asia: $0.099/GB
  South America: $0.140/GB

HTTP/HTTPS Requests:
  HTTP: $0.0075 per 10K requests
  HTTPS: $0.01 per 10K requests

Lambda@Edge: $0.60 per 1M invocations
```

## RDS Backup & Snapshots

```yaml
Storage Costs:
  Automated backups: Included (up to DB storage size)
  Manual snapshots: $0.023/GB
  Cross-region backup: $0.023/GB + data transfer

Example: 500GB database
  Monthly backup cost: $500 × $0.023 = $11.50
  Annual cost: $138
```

## AWS Lambda Layers

```python
# Layer Structure
/python/lib/python3.11/site-packages/
/python/bin/
/nodejs/node_modules/
/java/lib/

# Usage in Lambda
{
  "Layers": [
    "arn:aws:lambda:us-east-1:123456789012:layer:MyLayer:1"
  ]
}

# Size Limits
Uncompressed: 262 MB
Zipped: 50 MB
```

## Service Quotas & Limits

```yaml
EC2:
  Instances per region: 20 (on-demand)
  Security groups per region: 500
  VPCs per region: 5

Lambda:
  Concurrent executions: 1000 (default)
  Function timeout: 900 seconds (15 min)
  Deployment package: 50MB (zipped)
  Ephemeral storage: 512MB (or up to 10GB)

RDS:
  Database instances: 40 per region
  Database backups: 40 per region
  Max storage: 65.5 TB

DynamoDB:
  Partitions: Unlimited
  Per-partition throughput: 40GB storage
  Item size: 400KB
```

## Environment Variables Reference

```python
import os

# Lambda Environment
aws_request_id = os.environ['AWS_REQUEST_ID']
lambda_function_name = os.environ['AWS_LAMBDA_FUNCTION_NAME']
lambda_function_version = os.environ['AWS_LAMBDA_FUNCTION_VERSION']
lambda_log_group_name = os.environ['AWS_LAMBDA_LOG_GROUP_NAME']
lambda_memory_limit_mb = os.environ['AWS_LAMBDA_FUNCTION_MEMORY_IN_MB']

# AWS SDK Specific
aws_region = os.environ.get('AWS_REGION', 'us-east-1')
aws_access_key_id = os.environ.get('AWS_ACCESS_KEY_ID')
aws_secret_access_key = os.environ.get('AWS_SECRET_ACCESS_KEY')
aws_session_token = os.environ.get('AWS_SESSION_TOKEN')
```

## Context7 Integration

### Related AWS Services
- [AWS EC2](/aws/ec2): Elastic Compute Cloud
- [AWS RDS](/aws/rds): Relational Database Service
- [AWS Lambda](/aws/lambda): Serverless Functions
- [AWS DynamoDB](/aws/dynamodb): NoSQL Database
- [AWS S3](/aws/s3): Object Storage

### Official Documentation
- [AWS Services Overview](https://aws.amazon.com/products/)
- [EC2 Instance Types](https://aws.amazon.com/ec2/instance-types/)
- [RDS Features](https://aws.amazon.com/rds/features/)
- [Lambda Pricing](https://aws.amazon.com/lambda/pricing/)

### Configuration Examples
- [Terraform AWS Provider](/hashicorp/terraform-provider-aws)
- [AWS CloudFormation](https://aws.amazon.com/cloudformation/)
- [SAM (Serverless Application Model)](/aws/sam)

---

**Last Updated**: November 2025
**AWS Service Count**: 200+
**Regional Availability**: 33 regions globally
