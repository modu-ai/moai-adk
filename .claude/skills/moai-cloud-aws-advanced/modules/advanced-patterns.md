# AWS Advanced Architectural Patterns

## Multi-Region Active-Active Architecture

### Pattern Overview

**Use Case**: Global applications requiring zero-downtime failover and sub-second latency.

**Architecture Diagram**:
```
┌─────────────────────────────────────────────────────┐
│         Route 53 (Geolocation Routing)              │
└────────┬──────────────────────────┬─────────────────┘
         │                          │
    ┌────▼────┐              ┌──────▼────┐
    │US-EAST-1│              │EU-WEST-1  │
    │ Region  │◄─────────────►│  Region   │
    └────┬────┘              └──────┬────┘
         │ ALB + ASG         ALB + ASG│
         │ RDS Aurora ──────► RDS Aurora
         │ S3 + CRR ────────► S3 + CRR
         │ DynamoDB GT ─────► DynamoDB GT
         │ CloudFront CDN covers both regions
```

### Implementation Steps

1. **Primary Region Setup** (us-east-1)
   - VPC with public/private subnets across 3 AZs
   - Application Load Balancer (ALB) with auto-scaling group
   - RDS Aurora MySQL with read replicas
   - S3 bucket with versioning and CRR enabled

2. **Secondary Region Setup** (eu-west-1)
   - Identical VPC configuration
   - ALB + ASG mirroring primary
   - Aurora with Global Database (replication)
   - S3 destination bucket for CRR

3. **Route 53 Configuration**
   - Geolocation routing policies
   - Health checks every 30 seconds
   - Failover based on health status
   - TTL: 60 seconds for fast failover

### Cost Considerations

- **Data Transfer**: €0.02/GB for CRR
- **Global Database**: +30% RDS cost for replication
- **Route 53**: €0.50/million queries
- **CloudFront**: €0.085/GB outbound

**Estimated Monthly Cost**: $5,000-10,000 depending on traffic.

---

## Serverless Data Lake Architecture

### Pattern Overview

**Use Case**: Scalable analytics on massive datasets with cost-efficient storage.

**Components**:
```
S3 Data Lake ──► AWS Glue (ETL) ──► Athena (Query) ──► QuickSight (BI)
                       │
                  Crawler (Auto-schema)
                       │
                   Glue Catalog
```

### Lake Structure

```
s3://data-lake-prod/
├── raw/
│   ├── customers/
│   │   ├── year=2024/month=11/day=22/
│   │   │   └── customers_20241122.parquet
│   ├── orders/
│   │   ├── year=2024/month=11/
│   │   │   └── orders_20241122.parquet
│
├── processed/
│   ├── customer_metrics/
│   │   └── year=2024/month=11/
│   │       └── customer_metrics_v2.parquet
│
├── aggregated/
│   ├── daily_sales/
│   │   └── year=2024/month=11/
│   │       └── sales_daily.parquet
```

### Athena Query Patterns

```sql
-- Cost-optimized queries using partitions
SELECT
    customer_id,
    COUNT(*) as order_count,
    SUM(total_amount) as revenue
FROM orders
WHERE year = 2024 AND month = 11  -- Partition pruning!
GROUP BY customer_id
HAVING COUNT(*) > 5
ORDER BY revenue DESC;

-- External table for real-time data
CREATE EXTERNAL TABLE IF NOT EXISTS logs (
    timestamp STRING,
    level STRING,
    message STRING
)
PARTITIONED BY (date STRING)
STORED AS JSON
LOCATION 's3://logs-bucket/application-logs/';

-- Query with automatic decompression
SELECT * FROM logs
WHERE date = '2024-11-22'
AND level = 'ERROR'
LIMIT 100;
```

### Cost Optimization

- **Columnar Format**: Use Parquet (5-10x compression)
- **Partitioning**: Partition by date, region, environment
- **Bucketing**: For frequently joined tables
- **Approximate Queries**: Use COUNT DISTINCT APPROX for large datasets

**Expected Cost**: $0.005 per GB scanned (Athena). Parquet reduces scan by 80%.

---

## Lambda-Driven Microservices Architecture

### Pattern Components

```yaml
API Gateway
  ├─ POST /users ──► Lambda (User Service)
  │                   │
  │                   ├─► DynamoDB (Users Table)
  │                   ├─► SNS Topic
  │
  ├─ GET /orders ──► Lambda (Order Service)
  │                   │
  │                   ├─► RDS Aurora
  │                   ├─► DynamoDB Cache
  │
  └─ DELETE /orders/{id} ──► Lambda (Order Deletion)
                              │
                              ├─► SQS Queue
                              ├─► EventBridge
```

### Event-Driven Pattern

**Service Communication**:
1. Order Service → SNS Topic
2. SNS triggers multiple Lambda functions:
   - Inventory Service (reduce stock)
   - Notification Service (send email)
   - Analytics Service (track metrics)
   - Fulfillment Service (start packing)

**Benefits**: Decoupling, scalability, fault isolation.

### Cost Optimization Strategies

1. **Concurrency Limits**: Set per function (prevent runaway)
2. **Memory Optimization**: Find optimal memory size (higher = faster, lower = cheaper)
3. **Timeout Settings**: Minimal timeout reduces cost
4. **Compute Savings Plans**: 3-year commitment = 39% savings

**AWS Lambda Cost Calculator**:
```
Invocations: 10M/month
Memory: 512 MB
Duration: 500ms

Cost = (10M × 0.0000002) + (10M × 500ms × 512MB × 0.0000166667)
     = $2 + $42
     = $44/month
```

---

## Database Performance Patterns

### RDS Aurora Optimization

**Read Replica Distribution**:
```
Writer Node (us-east-1a)
  ├─► Read Replica (us-east-1b)
  ├─► Read Replica (us-east-1c)
  ├─► Read Replica (eu-west-1)
  └─► Read Replica (ap-southeast-1)

Writer: 100% write traffic
Readers: All read queries distributed
```

**Query Performance**:
- Enable query cache (if compatible with your pattern)
- Create appropriate indexes
- Use Aurora Serverless for variable workloads

### DynamoDB Design Patterns

**Single-Table Design**:
```
Primary Key: PK=ENTITY#123, SK=TYPE#ORDER#456
GSI-1: PK=USER#789, SK=CREATED_AT#2024-11-22
GSI-2: PK=STATUS#PENDING, SK=CREATED_AT#2024-11-22
```

**Benefits**:
- Single API call for related data
- Better cost efficiency
- Easier scaling

**Partition Key Distribution**:
- Avoid hot partitions
- Use write-sharding for high-frequency updates
- Monitor throttling with CloudWatch

---

**Related**: See optimization.md for specific cost reduction strategies.
