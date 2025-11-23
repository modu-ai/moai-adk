# GCP Services - Quick Reference

## Compute Services

```yaml
Compute Engine:
  - vCPU: $0.033-0.054/hour (on-demand)
  - Memory: $0.0044/GB/hour
  - CUD: 37% discount (annual commitment)

GKE:
  - Autopilot: $0.10/hour per cluster + node costs
  - Standard: Free control plane + node costs
  - Per-node: $0.048/hour (e2-medium)

Cloud Run:
  - CPU: $0.00002400/second
  - Memory: $0.00000250/second/GB
  - Requests: $0.40/million

Cloud Functions:
  - Invocations: $0.40/million
  - Compute: $0.000008333/GB-second
  - Free: 2M invocations/month
```

## Database Services

```yaml
Cloud SQL:
  - MySQL/PostgreSQL: $0.36-0.65/hour (shared)
  - High Availability: 2x cost
  - Storage: $0.30/GB/month
  - Backup: $0.10/GB/month

Firestore:
  - Reads: $0.06/100K
  - Writes: $0.18/100K
  - Deletes: $0.02/100K
  - Storage: $0.18/GB/month

Bigtable:
  - Node: $1.50/hour
  - Storage: $0.01/GB/month
  - Backup: $0.05/GB/month

Spanner:
  - Node: $7.80/hour
  - Storage: $0.30/GB/month
  - Backup: $0.15/GB/month
```

## Analytics Services

```yaml
BigQuery:
  - Queries: $7.50/TB scanned
  - Storage: Standard $6.25/TB, Archive $0.025/GB
  - BI Engine: $0.02/hour (cache)
  - Streaming: $0.50/100K rows

Dataflow:
  - Worker: $0.00/hour (free tier: 100 vCPU/month)
  - CPU: $0.064/vCPU/hour
  - Memory: $0.008/GB/hour
  - Storage: $0.25/GB

Pub/Sub:
  - Publish: $0.50/million messages
  - Subscribe: $0.50/million messages
  - Subscription age: $0.10/GB/hour
```

## Machine Learning

```yaml
Vertex AI:
  - AutoML Training: $1.95-9.87/hour
  - Training Job: Google Cloud resources + $0.3/hour
  - Batch Prediction: $0.00003125/1K instances
  - Online Prediction: $10/month + compute

Vision API:
  - Image recognition: $1.50/1K images
  - Document AI: $1.50/page
  - Optical character recognition: $1.50/1K

Natural Language API:
  - Text analysis: $0.50/1K requests
  - Entity recognition: $1.50/1K requests
  - Sentiment analysis: $1.00/1K requests
```

## Common gcloud Commands

```bash
# Create GKE Autopilot cluster
gcloud container clusters create my-cluster \
  --enable-autopilot \
  --region=us-central1

# Deploy to Cloud Run
gcloud run deploy my-service \
  --source . \
  --region us-central1 \
  --allow-unauthenticated

# Deploy to Compute Engine
gcloud compute instances create my-vm \
  --machine-type=e2-medium \
  --zone=us-central1-a

# Create BigQuery dataset
gcloud bq datasets create --project_id=my-project my_dataset

# Deploy Cloud Function
gcloud functions deploy my-function \
  --runtime python311 \
  --trigger-topic my-topic
```

## Environment Variables

```python
import os

# Default from gcloud
GOOGLE_APPLICATION_CREDENTIALS = os.getenv('GOOGLE_APPLICATION_CREDENTIALS')
GOOGLE_CLOUD_PROJECT = os.getenv('GOOGLE_CLOUD_PROJECT')

# Cloud Functions
FUNCTION_SIGNATURE_TYPE = 'http'
FUNCTION_NAME = os.getenv('K_SERVICE')

# Cloud Run
PORT = os.getenv('PORT', '8080')
K_SERVICE = os.getenv('K_SERVICE')  # Service name
K_REVISION = os.getenv('K_REVISION')  # Revision number
```

## Service Quotas

```yaml
Compute Engine:
  CPUs: 100 (default per region)
  GPUs: 0 (request quota)
  Persistent Disks: 500GB (per region)

Cloud SQL:
  Instances: 10 (per region)
  Backups: 30 (per instance)
  Storage: 15TB (default)

Cloud Pub/Sub:
  Topics: 10,000 (per region)
  Subscriptions: 10,000 (per region)
  Message size: 10MB maximum

BigQuery:
  Tables: 10,000 (per dataset)
  Datasets: 1,000 (per project)
  Query size: 12.5MB (compressed)
```

## Context7 Integration

### Related GCP Services
- [Google Cloud Compute](/gcp/compute): Virtual machines and containers
- [Google Cloud SQL](/gcp/cloudsql): Managed relational databases
- [Google Cloud Pub/Sub](/gcp/pubsub): Real-time messaging
- [Google Kubernetes Engine](/gcp/gke): Managed Kubernetes

### Official Documentation
- [GCP Documentation](https://cloud.google.com/docs)
- [Cloud Architecture Center](https://cloud.google.com/architecture)
- [GCP Pricing Calculator](https://cloud.google.com/products/calculator)

---

**Last Updated**: November 2025
**GCP Service Count**: 100+
**Regional Availability**: 40+ regions globally
