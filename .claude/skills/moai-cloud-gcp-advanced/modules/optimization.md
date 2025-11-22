# GCP Cost & Performance Optimization

## 1. BigQuery Cost Optimization

**Query Cost Reduction**:
```sql
-- Expensive: Scans entire table
SELECT * FROM large_table;

-- Optimized: Only needed columns, partition pruning
SELECT user_id, name, email
FROM large_table
WHERE DATE(_PARTITIONTIME) = CURRENT_DATE()
AND region = 'US';

-- Result: 50x-100x cost reduction
```

**Storage Optimization**:
- Standard: $6.25/TB/month
- Archive: $0.025/GB/month (after 365 days)
- Lifecycle policy: Migrate old data to cheaper tiers

**Compression Strategy**:
- Parquet: 5-10x compression
- ORC: 10-15x compression
- Estimated storage savings: 80%

## 2. GKE Cost Optimization

**Workload Optimization**:

```python
# Monitor node utilization
gcloud container clusters describe my-cluster --format='value(nodePools[0].config.machineType)'

# Calculate CPU/Memory requests
requests = {
    'CPU': '500m',      # 0.5 cores
    'Memory': '512Mi'
}

# Resource limits prevent node over-commitment
limits = {
    'CPU': '1000m',     # Max 1 core
    'Memory': '1Gi'
}
```

**Cost Savings**:
- Autopilot: 10-15% better utilization than Standard
- Spot VMs: 60-70% savings
- Committed Use Discounts (CUD): 37% savings

## 3. Cloud Run Optimization

**Concurrency Tuning**:
```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: my-service
spec:
  template:
    spec:
      containerConcurrency: 80  # Requests per container
      containers:
      - resources:
          limits:
            memory: 2Gi
            cpu: 2
```

**Cost Structure**:
- CPU: $0.00002400/second (on-demand)
- Memory: $0.00000250/second/GB
- Requests: $0.40/million

**Optimization**:
- Set appropriate min instances
- Tune max instances per region
- Use Spot pricing (up to 70% discount)

## 4. Firestore Cost Reduction

**Pricing Model**:
- Reads: $0.06/100K reads
- Writes: $0.18/100K writes
- Deletes: $0.02/100K deletes
- Storage: $0.18/GB/month

**Optimization**:
- Batch operations (reduce write count)
- Cache frequently read documents
- Use collection group indexes only when needed
- Archive old data to BigQuery

**Example Cost Comparison**:
```
Naive approach: 1M daily transactions = $5,400/month
Optimized: Batch writes + caching = $180/month
Savings: 97%
```

## 5. Storage Optimization

**Cloud Storage Classes**:
- Standard: $0.020/GB
- Nearline: $0.010/GB (30+ days)
- Coldline: $0.004/GB (90+ days)
- Archive: $0.0025/GB (365+ days)

**Lifecycle Policy**:
```xml
<Rules>
  <Rule>
    <Action type="SetStorageClass">
      <StorageClass>NEARLINE</StorageClass>
    </Action>
    <Condition>
      <AgeInDays>30</AgeInDays>
    </Condition>
  </Rule>
</Rules>
```

**Potential Savings**: 60-80% for logs/archives

## 6. Network Optimization

**Data Transfer Costs**:
- Ingress: Free
- Egress to Internet: $0.12/GB
- Egress between regions: $0.01/GB
- Egress to on-premises: $0.05/GB

**Cost Reduction**:
- Use Cloud CDN ($0.085/GB, 30% savings)
- Colocate resources in same region
- Use VPC Service Controls (no egress)

---

**Combined Optimization Potential**: 40-60% cost reduction across all services.

**Implementation Timeline**: 2-4 weeks for full optimization.
