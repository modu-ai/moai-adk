---
name: moai-cloud-gcp-advanced
version: 4.0.0
updated: 2025-11-20
status: stable
tier: specialization
description: Advanced GCP patterns for Cloud Run, Vertex AI, BigQuery, and GKE
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Google Cloud Advanced Architectures

**Enterprise Serverless, ML, and Analytics Patterns**

> **Focus**: Cloud Run Gen2, Vertex AI Pipelines, BigQuery Streaming, GKE  
> **Stack**: Python, Terraform, Docker, Kubeflow

---

## Overview

Production-grade patterns for scalable GCP architectures.

### Core Services

- **Compute**: Cloud Run (Serverless Containers), GKE (Kubernetes)
- **ML**: Vertex AI (Training & Serving), AutoML
- **Data**: BigQuery (Analytics), Pub/Sub (Streaming)
- **Security**: VPC Service Controls, IAM Workload Identity

### Compute Selection Guide

| Service             | Best For              | Scaling            | Cost Model     |
| ------------------- | --------------------- | ------------------ | -------------- |
| **Cloud Run**       | Microservices, APIs   | 0 to N             | Per request    |
| **GKE**             | Complex Stateful Apps | Cluster Autoscaler | Per node       |
| **Cloud Functions** | Event-driven Triggers | Per request        | Per invocation |
| **App Engine**      | Monolithic Web Apps   | Auto               | Per instance   |

---

## Implementation Patterns

### 1. Cloud Run Gen2 Deployment

Traffic splitting and revision management.

```yaml
# service.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: api-service
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/minScale: "1"
        autoscaling.knative.dev/maxScale: "100"
    spec:
      containers:
        - image: gcr.io/project/api:v2
          resources:
            limits:
              cpu: "1000m"
              memory: "512Mi"
          env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: host
  traffic:
    - percent: 10
      revisionName: api-service-v2
    - percent: 90
      revisionName: api-service-v1
```

Deploy command:

```bash
gcloud run services replace service.yaml --region us-central1
```

### 2. Vertex AI Training Pipeline

Custom training job with Python SDK.

```python
from google.cloud import aiplatform

def submit_training_job(
    project_id: str,
    display_name: str,
    container_uri: str,
    model_serving_container_image_uri: str,
):
    aiplatform.init(project=project_id, location="us-central1")

    job = aiplatform.CustomContainerTrainingJob(
        display_name=display_name,
        container_uri=container_uri,
        model_serving_container_image_uri=model_serving_container_image_uri,
    )

    model = job.run(
        dataset=None, # Or provide dataset
        model_display_name="my-model",
        args=["--epochs=10"],
        replica_count=1,
        machine_type="n1-standard-4",
        accelerator_type="NVIDIA_TESLA_T4",
        accelerator_count=1,
    )

    return model
```

### 3. BigQuery Streaming Insert

High-throughput data ingestion.

```python
from google.cloud import bigquery

def stream_data(dataset_id, table_id, rows):
    client = bigquery.Client()
    table_ref = client.dataset(dataset_id).table(table_id)

    errors = client.insert_rows_json(table_ref, rows)

    if errors:
        print(f"Encountered errors: {errors}")
    else:
        print("New rows have been added.")
```

### 4. GKE Workload Identity

Secure access to GCP APIs from Kubernetes.

```bash
# 1. Create Google Service Account (GSA)
gcloud iam service-accounts create gke-sa

# 2. Bind GSA to K8s Service Account (KSA)
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:PROJECT_ID.svc.id.goog[NAMESPACE/KSA_NAME]" \
  gke-sa@PROJECT_ID.iam.gserviceaccount.com

# 3. Annotate KSA
kubectl annotate serviceaccount KSA_NAME \
  --namespace NAMESPACE \
  iam.gke.io/gcp-service-account=gke-sa@PROJECT_ID.iam.gserviceaccount.com
```

---

## Infrastructure as Code (Terraform)

### VPC & Cloud Run Connector

```hcl
resource "google_vpc_access_connector" "connector" {
  name          = "vpc-conn"
  region        = "us-central1"
  ip_cidr_range = "10.8.0.0/28"
  network       = "default"
}

resource "google_cloud_run_service" "default" {
  name     = "my-service"
  location = "us-central1"

  template {
    spec {
      containers {
        image = "gcr.io/google-samples/hello-app:1.0"
      }
    }
    metadata {
      annotations = {
        "run.googleapis.com/vpc-access-connector" = google_vpc_access_connector.connector.name
      }
    }
  }
}
```

---

## Observability

### Cloud Logging (Python)

```python
import google.cloud.logging

client = google.cloud.logging.Client()
client.setup_logging()

import logging
logging.info("Structured log message", extra={
    "json_fields": {"custom_key": "value"}
})
```

---

## Validation Checklist

**Compute**:

- [ ] Cloud Run min instances set (cold start prevention)
- [ ] GKE Workload Identity configured
- [ ] VPC Connector enabled for private access

**Data & ML**:

- [ ] BigQuery partitioning enabled
- [ ] Vertex AI model registry used
- [ ] Pub/Sub dead letter queues configured

**Security**:

- [ ] Least privilege IAM roles assigned
- [ ] Secrets managed via Secret Manager
- [ ] Public access restricted (ingress control)

---

## Related Skills

- `moai-domain-cloud`: Cloud fundamentals
- `moai-domain-ml`: Machine Learning concepts
- `moai-security-devsecops`: CI/CD security

---

**Last Updated**: 2025-11-20
