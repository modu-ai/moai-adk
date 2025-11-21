---
name: moai-cloud-gcp-advanced
description: Advanced GCP architecture patterns, BigQuery optimization, and Kubernetes Engine best practices
allowed-tools: [Read, Bash, WebFetch]
---

# Advanced GCP Cloud Architecture

## Quick Reference

Enterprise GCP patterns for Kubernetes Engine (GKE), BigQuery analytics, multi-region deployments, and cost-optimized data pipelines.

**Core Services** (November 2025):
- **Compute**: Compute Engine, GKE, Cloud Run, Cloud Functions
- **Storage**: Cloud Storage, Persistent Disk, Filestore
- **Database**: Cloud SQL, Spanner, Firestore, Bigtable
- **Analytics**: BigQuery, Dataflow, Pub/Sub, Dataproc
- **AI/ML**: Vertex AI, Cloud AI Platform

---

## Implementation Guide

### GKE Autopilot

**Managed Kubernetes**:
```yaml
apiVersion: container.googleapis.com/v1
kind: Cluster
spec:
  autopilot:
    enabled: true
  releaseChannel: REGULAR
  workloadIdentity: enabled
  binaryAuthorization: enabled
  podSecurity: BASELINE
```

### BigQuery Optimization

**Partitioning Strategy**:
```sql
CREATE TABLE `project.dataset.events`
PARTITION BY DATE(timestamp)
CLUSTER BY user_id, event_type
OPTIONS(
  partition_expiration_days=90,
  require_partition_filter=true
);
```

---

## Best Practices

### ✅ DO
- Use managed services (Cloud Run, Cloud SQL)
- Enable workload identity (no service account keys)
- Implement budget alerts (prevent cost overruns)
- Use Cloud CDN (global content delivery)

### ❌ DON'T
- Store service account keys in code
- Skip VPC Service Controls (data exfiltration risk)
- Ignore security command center alerts

---

## Works Well With

- `moai-cloud-aws-advanced` (Multi-cloud comparison)
- `moai-domain-devops` (GCP CI/CD)

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Official Reference**: https://cloud.google.com/architecture/framework
