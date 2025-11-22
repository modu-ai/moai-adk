# AWS Cost & Performance Optimization Strategies

## Cost Optimization Framework

### 1. Compute Optimization

**EC2 Right-Sizing**:
```python
import boto3
from datetime import datetime, timedelta

class EC2Optimizer:
    def __init__(self):
        self.cloudwatch = boto3.client('cloudwatch')
        self.ec2 = boto3.client('ec2')

    def analyze_cpu_utilization(self, instance_id, days=7):
        """Analyze CPU usage to right-size instance"""
        end_time = datetime.now()
        start_time = end_time - timedelta(days=days)

        response = self.cloudwatch.get_metric_statistics(
            Namespace='AWS/EC2',
            MetricName='CPUUtilization',
            Dimensions=[{'Name': 'InstanceId', 'Value': instance_id}],
            StartTime=start_time,
            EndTime=end_time,
            Period=3600,  # 1 hour
            Statistics=['Average', 'Maximum']
        )

        datapoints = response['Datapoints']
        avg_cpu = sum(dp['Average'] for dp in datapoints) / len(datapoints)
        max_cpu = max(dp['Maximum'] for dp in datapoints)

        return {
            'instance_id': instance_id,
            'avg_utilization': avg_cpu,
            'peak_utilization': max_cpu,
            'recommendation': self._recommend_instance_type(avg_cpu, max_cpu)
        }

    def _recommend_instance_type(self, avg, peak):
        """Recommend smaller instance if utilization is low"""
        if avg < 10 and peak < 30:
            return 'Consider t3.nano or t3.micro'
        elif avg < 20 and peak < 50:
            return 'Consider t3.small'
        else:
            return 'Current instance type is appropriate'
```

**Savings Plans**:
- **Compute Savings Plans**: 24-40% savings, flexible across instances
- **EC2 Instance Savings Plans**: 30-50% savings, specific instance family
- **3-Year Commitment**: 40% more savings than 1-year

### 2. Storage Optimization

**S3 Storage Classes**:
```
Standard:        $0.023/GB       (< 1ms access)
Intelligent-Tier: $0.0125/GB     (auto-optimized)
Standard-IA:     $0.0125/GB      (infrequent access)
Glacier:         $0.004/GB       (archive, 1-12hr retrieval)
Deep Archive:    $0.00099/GB     (long-term, 12-48hr retrieval)
```

**Lifecycle Policy**:
```json
{
  "Rules": [
    {
      "Filter": {"Prefix": "logs/"},
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "STANDARD_IA"
        },
        {
          "Days": 90,
          "StorageClass": "GLACIER"
        },
        {
          "Days": 365,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ],
      "Expiration": {
        "Days": 2555  // 7 years
      }
    }
  ]
}
```

**Estimated Savings**: 70-80% for log data.

### 3. Database Optimization

**RDS Cost Reduction**:

| Strategy | Savings | Implementation |
|----------|---------|-----------------|
| Reserved Instances | 30-40% | Commit 1-3 years |
| Graviton2 CPU | 20% | Switch to t4g/r6g |
| Storage Optimization | 15-25% | Delete unused snapshots |
| Automated Backups | 10% | Reduce retention to 7 days |

**DynamoDB On-Demand vs Provisioned**:
```
Provisioned: 100 WCU × $1.25/month = $125/month
On-Demand: Pay per request (good for variable)

Break-even: ~4M requests/month
```

### 4. Network Optimization

**Data Transfer Costs**:
```
Internet → AWS: Free
AWS → Internet: $0.09/GB (first 1GB free)
Inter-Region: $0.01-0.02/GB
CloudFront: $0.085/GB (50% less than direct)
```

**Optimization**:
- Use CloudFront for static assets
- VPC endpoints for AWS service access (free)
- AWS Global Accelerator for global optimization

### Performance Optimization

## 1. Lambda Performance

**Cold Start Reduction**:
```python
# Provisioned Concurrency
import boto3

lambda_client = boto3.client('lambda')

# Pre-warm 100 concurrent executions
lambda_client.put_provisioned_concurrency_config(
    FunctionName='my-function',
    ProvisionedConcurrentExecutions=100
)

# Result: 0.3-0.4s cold start → 50-100ms
```

**Memory Tuning**:
```
512 MB: ~$0.0000083/s, slower
1024 MB: ~0.0000167/s, 2x faster but 2x cost
1769 MB: $0.0000294/s, best cost/performance for CPU-bound
```

## 2. RDS Performance

**Connection Pooling**:
```python
# Use RDS Proxy
import psycopg2

# Without proxy: 1000 connections = 4GB RAM
# With proxy: 1000 connections = 20MB RAM

connection_pool = {
    'host': 'proxy-endpoint.rds-proxy.us-east-1.amazonaws.com',
    'max_connections': 100,
    'idle_client_timeout': 900
}
```

**Query Optimization**:
- Add indexes on frequently queried columns
- Use EXPLAIN ANALYZE for slow queries
- Enable Performance Insights ($0.02/hour)

## 3. DynamoDB Performance

**Partition Key Distribution**:
```python
# Bad: All users write to same partition
partition_key = 'GLOBAL#METRICS'  # Hot partition!

# Good: Distribute across multiple partitions
partition_key = f'METRICS#{datetime.now().strftime("%H")}'  # Per-hour bucket
```

**Global Secondary Index Strategy**:
```
Main Query: user_id + created_at
GSI: status + user_id
GSI: region + timestamp

Monitor: Query throttling alerts
```

---

## Cost vs Performance Trade-offs

### Decision Matrix

```
┌─────────────────┬──────────────┬──────────────┐
│ Requirement     │ Cost Focus   │ Performance  │
├─────────────────┼──────────────┼──────────────┤
│ 99.99% uptime   │ Multi-AZ RDS │ Multi-region │
│ 100M requests   │ DynamoDB On- │ Provisioned  │
│ Read-heavy      │ Read replicas│ Aurora Proxy │
│ Write-heavy     │ DynamoDB     │ Sharding     │
│ Batch/ETL       │ Athena       │ Spark/Glue  │
└─────────────────┴──────────────┴──────────────┘
```

### Implementation Checklist

- [ ] Enable CloudTrail for compliance
- [ ] Set up Cost Anomaly Detection
- [ ] Implement tagging for cost allocation
- [ ] Review Reserved Instance recommendations (monthly)
- [ ] Archive unused resources
- [ ] Monitor data transfer costs
- [ ] Use AWS Compute Optimizer (free)

---

**Monthly Cost Savings Potential**: 30-50% with proper optimization.

**Estimated Annual Savings**: $50K-500K depending on current architecture.
