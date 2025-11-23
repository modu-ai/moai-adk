---
name: moai-cloud-aws-advanced
description: Advanced AWS architecture patterns, cost optimization, and enterprise deployment strategies
version: 1.0.0
modularized: false
tags:
  - aws
  - enterprise
  - development
  - advanced
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: cloud, aws, advanced, moai  


# Advanced AWS Cloud Architecture

## Quick Reference

Enterprise-grade AWS patterns covering multi-region deployments, cost optimization, security hardening, and high-availability architectures for production workloads at scale.

**Core Services** (November 2025):
- **Compute**: EC2, ECS, EKS, Lambda, Fargate
- **Storage**: S3, EBS, EFS, FSx
- **Database**: RDS, Aurora, DynamoDB, DocumentDB
- **Network**: VPC, CloudFront, Route 53, Transit Gateway
- **Security**: IAM, KMS, Secrets Manager, WAF, Shield

**Architecture Patterns**:
- Multi-region active-active (99.99% uptime)
- Serverless-first (Lambda + API Gateway + DynamoDB)
- Microservices on EKS (Kubernetes orchestration)
- Data lakes (S3 + Athena + Glue)


## Implementation Guide

### Multi-Region Architecture

**Active-Active Pattern**:
```yaml
Regions: us-east-1, eu-west-1, ap-southeast-1

Load Balancing:
  - Route 53 geolocation routing
  - CloudFront global distribution
  - Health checks every 30s

Data Replication:
  - Aurora Global Database (< 1s lag)
  - S3 Cross-Region Replication (CRR)
  - DynamoDB Global Tables (multi-master)

Failover Strategy:
  - Automatic health-based failover
  - < 60s RTO (Recovery Time Objective)
  - Zero RPO (Recovery Point Objective) for databases
```

### Cost Optimization

**Reserved Instances**:
```python
# 72% savings for steady-state workloads
reserved_instances = {
    "EC2": {
        "term": "3-year All Upfront",
        "savings": "72%",
        "use_case": "Baseline compute capacity"
    },
    "RDS": {
        "term": "1-year Partial Upfront",
        "savings": "38%",
        "use_case": "Production databases"
    }
}
```

**Spot Instances**:
```python
# 90% savings for fault-tolerant workloads
spot_strategies = {
    "Batch Processing": "100% Spot (accept interruptions)",
    "Web Servers": "Mix of 50% On-Demand + 50% Spot",
    "ML Training": "100% Spot with checkpointing"
}
```

### Security Hardening

**IAM Least Privilege**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::my-bucket/*",
      "Condition": {
        "IpAddress": {
          "aws:SourceIp": "10.0.0.0/16"
        }
      }
    }
  ]
}
```

**KMS Encryption**:
```python
# Encrypt all data at rest
encryption_strategy = {
    "S3": "AES-256 with KMS CMK",
    "RDS": "TDE with KMS",
    "EBS": "KMS encryption by default",
    "Secrets Manager": "KMS CMK per environment"
}
```


## Advanced Patterns

### Serverless Architecture

**Lambda + API Gateway + DynamoDB**:
```yaml
API Gateway:
  - REST API with request validation
  - Throttling: 10,000 requests/second
  - Caching: 5-minute TTL

Lambda:
  - Runtime: Python 3.12
  - Memory: 512MB-3GB (auto-scaling)
  - Concurrency: 1000 reserved
  - Cold start optimization: < 200ms

DynamoDB:
  - On-Demand billing (auto-scaling)
  - Point-in-time recovery enabled
  - Global tables for multi-region
```

### Microservices on EKS

**Cluster Configuration**:
```yaml
apiVersion: eks.amazonaws.com/v1
kind: Cluster
spec:
  version: "1.28"
  nodeGroups:
    - name: general-purpose
      instanceTypes: [t3.large, t3.xlarge]
      desiredSize: 3
      minSize: 2
      maxSize: 10
      spotInstances: true
  addons:
    - kube-proxy
    - vpc-cni
    - coredns
    - aws-load-balancer-controller
```


## Best Practices

### ✅ DO
- Use multi-AZ deployments (99.99% SLA)
- Enable CloudTrail for all regions (audit compliance)
- Implement auto-scaling (horizontal + vertical)
- Use VPC endpoints (reduce NAT gateway costs)
- Tag all resources (cost allocation, automation)
- Enable AWS Config (compliance monitoring)
- Use Systems Manager for patch management

### ❌ DON'T
- Use root account for daily operations
- Store credentials in code (use Secrets Manager)
- Ignore cost anomaly alerts (budget overruns)
- Deploy without disaster recovery plan
- Skip VPC flow logs (security blind spot)
- Use default security groups (too permissive)


## Works Well With

- `moai-cloud-gcp-advanced` (Multi-cloud strategies)
- `moai-domain-devops` (CI/CD integration)
- `moai-security-api` (API security patterns)


**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready  
**Official Reference**: https://aws.amazon.com/architecture/well-architected/
