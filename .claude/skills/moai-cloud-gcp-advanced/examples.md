# GCP Advanced Patterns - Practical Examples

## Example 1: BigQuery Data Warehouse with Real-Time Streaming

```python
from google.cloud import bigquery
from google.cloud import pubsub_v1
import json
from datetime import datetime

class BigQueryRealtimeWarehouse:
    def __init__(self, project_id: str, dataset_id: str):
        self.client = bigquery.Client(project=project_id)
        self.dataset_id = dataset_id
        self.publisher = pubsub_v1.PublisherClient()

    def create_streaming_table(self):
        """Create table optimized for real-time streaming"""
        schema = [
            bigquery.SchemaField("event_id", "STRING", mode="REQUIRED"),
            bigquery.SchemaField("timestamp", "TIMESTAMP", mode="REQUIRED"),
            bigquery.SchemaField("user_id", "STRING", mode="REQUIRED"),
            bigquery.SchemaField("event_type", "STRING", mode="REQUIRED"),
            bigquery.SchemaField("properties", "JSON", mode="NULLABLE"),
            bigquery.SchemaField("_PARTITIONTIME", "TIMESTAMP", mode="REQUIRED"),
        ]

        table_id = f"{self.dataset_id}.events_stream"
        table = bigquery.Table(table_id, schema=schema)
        table.time_partitioning = bigquery.TimePartitioning(
            type_=bigquery.TimePartitioningType.DAY,
            field="_PARTITIONTIME"
        )
        table.clustering_fields = ["user_id", "event_type"]

        return self.client.create_table(table)

    def stream_insert_rows(self, table_id: str, rows: list) -> list:
        """Insert rows with error handling"""
        errors = self.client.insert_rows_json(table_id, rows)
        if errors:
            print(f"Insert errors: {errors}")
        return errors

    def query_with_cache(self, query: str, cache_hours: int = 24) -> list:
        """Execute query with caching"""
        job_config = bigquery.QueryJobConfig(
            use_query_cache=True,
            maximum_bytes_billed=1000000  # 1MB limit
        )
        query_job = self.client.query(query, job_config=job_config)
        return list(query_job)

    def export_to_gcs(self, table_id: str, gcs_path: str) -> str:
        """Export table to GCS in Parquet format"""
        job_config = bigquery.ExtractJobConfig()
        job_config.destination_format = bigquery.DestinationFormat.PARQUET
        job_config.compression = bigquery.Compression.SNAPPY

        extract_job = self.client.extract_table(
            table_id,
            gcs_path,
            job_config=job_config
        )
        extract_job.result()
        return f"Exported to {gcs_path}"
```

## Example 2: GKE Autopilot with Workload Identity

```yaml
# GKE Autopilot Cluster Configuration
apiVersion: container.cnrm.cloud.google.com/v1beta1
kind: ContainerCluster
metadata:
  name: gke-autopilot-prod
spec:
  location: us-central1
  autopilot:
    enabled: true
  network: default
  subnetwork: default
  addonsConfig:
    httpLoadBalancing:
      disabled: false
    dnsCacheConfig:
      cacheNodeCount: 3

---
# Workload Identity Binding
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMServiceAccount
metadata:
  name: app-service-account
spec:
  displayName: Application Service Account

---
# Kubernetes Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-ksa
  namespace: default

---
# IAM Policy Binding
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicy
metadata:
  name: app-workload-identity
spec:
  resourceRef:
    apiVersion: iam.cnrm.cloud.google.com/v1beta1
    kind: IAMServiceAccount
    name: app-service-account
  bindings:
    - role: roles/iam.workloadIdentityUser
      members:
        - serviceAccount:project-id.svc.id.goog[default/app-ksa]
```

## Example 3: Firestore Real-Time Database with Transactions

```python
from google.cloud import firestore
from typing import Dict, List
import asyncio

class FirestoreRealTimeApp:
    def __init__(self, project_id: str):
        self.db = firestore.Client(project=project_id)
        self.async_db = firestore.AsyncClient(project=project_id)

    def batch_write_optimized(self, documents: List[Dict]) -> None:
        """Batch write for performance"""
        batch = self.db.batch()

        for doc in documents:
            doc_ref = self.db.collection('users').document(doc['id'])
            batch.set(doc_ref, {
                'name': doc['name'],
                'email': doc['email'],
                'created_at': firestore.SERVER_TIMESTAMP
            })

        batch.commit()

    def transactional_transfer(self, from_user: str, to_user: str, amount: float) -> bool:
        """Atomic transaction across documents"""
        @firestore.transactional
        def transfer_money(transaction):
            from_ref = self.db.collection('accounts').document(from_user)
            to_ref = self.db.collection('accounts').document(to_user)

            from_doc = from_ref.get(transaction=transaction)
            to_doc = to_ref.get(transaction=transaction)

            if from_doc.get('balance') < amount:
                return False

            transaction.update(from_ref, {'balance': from_doc.get('balance') - amount})
            transaction.update(to_ref, {'balance': to_doc.get('balance') + amount})
            return True

        transaction = self.db.transaction()
        return transfer_money(transaction)

    def listen_to_realtime_updates(self, user_id: str, callback) -> None:
        """Real-time listener for document changes"""
        def on_snapshot(doc_snapshot, changes, read_time):
            for doc in doc_snapshot:
                callback(doc.to_dict())

        query = self.db.collection('users').document(user_id)
        query.on_snapshot(on_snapshot)

    async def async_query_with_index(self, email: str) -> Dict:
        """Async query using composite index"""
        query = self.async_db.collection('users').where('email', '==', email)
        docs = await query.stream()

        async for doc in docs:
            return doc.to_dict()
```

## Example 4: Vertex AI Model Training & Prediction

```python
from google.cloud import aiplatform
from google.cloud import storage
import pandas as pd

class VertexAIMLPipeline:
    def __init__(self, project_id: str, location: str = 'us-central1'):
        aiplatform.init(project=project_id, location=location)
        self.storage_client = storage.Client()

    def train_automl_model(self, dataset_path: str, target_column: str):
        """Train AutoML model without ML expertise"""
        dataset = aiplatform.TabularDataset.create(
            display_name='training_dataset',
            gcs_source=dataset_path
        )

        model = aiplatform.AutoMLTabularTrainingJob(
            display_name='auto_ml_model',
            optimization_prediction_type='regression',
            optimization_objective='minimize-rmse'
        )

        trained_model = model.run(
            dataset=dataset,
            target_column=target_column,
            training_fraction_split=0.7,
            validation_fraction_split=0.15,
            test_fraction_split=0.15
        )

        return trained_model

    def deploy_model_to_endpoint(self, model, traffic_split=100):
        """Deploy trained model to endpoint"""
        endpoint = aiplatform.Endpoint.create(
            display_name='prediction_endpoint'
        )

        endpoint.deploy(
            model=model,
            deployed_model_display_name='model_v1',
            traffic_split={model.name: traffic_split}
        )

        return endpoint

    def make_predictions(self, endpoint, features_df: pd.DataFrame):
        """Make batch predictions"""
        predictions = endpoint.predict(
            instances=features_df.values.tolist()
        )

        return predictions.predictions
```

## Example 5: Cloud Run Serverless Deployment

```dockerfile
# Dockerfile for Cloud Run
FROM python:3.11-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENV PORT=8080
CMD exec gunicorn --bind :$PORT --workers 4 --max-requests 1000 app:app
```

```yaml
# Cloud Run deployment via gcloud
gcloud run deploy my-service \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 2Gi \
  --cpu 2 \
  --max-instances 100 \
  --min-instances 1 \
  --set-env-vars DATABASE_URL=$DB_URL \
  --service-account my-service-account
```

## Example 6: Pub/Sub Real-Time Event Processing

```python
from google.cloud import pubsub_v1
import json
from datetime import datetime

class PubSubEventProcessor:
    def __init__(self, project_id: str):
        self.publisher = pubsub_v1.PublisherClient()
        self.subscriber = pubsub_v1.SubscriberClient()
        self.project_id = project_id

    def publish_event(self, topic_id: str, message: dict) -> str:
        """Publish event to Pub/Sub topic"""
        topic_path = self.publisher.topic_path(self.project_id, topic_id)

        future = self.publisher.publish(
            topic_path,
            json.dumps(message).encode('utf-8')
        )

        return future.result()

    def subscribe_to_events(self, subscription_id: str, callback):
        """Subscribe to events with callback"""
        subscription_path = self.subscriber.subscription_path(
            self.project_id,
            subscription_id
        )

        flow_control = pubsub_v1.types.FlowControl(
            max_messages=10,
            max_bytes=100 * 1024 * 1024  # 100MB
        )

        streaming_pull_future = self.subscriber.subscribe(
            subscription_path,
            callback=callback,
            flow_control=flow_control
        )

        return streaming_pull_future
```

---

**Learn More**: See advanced-patterns.md for architectural deep-dives.
