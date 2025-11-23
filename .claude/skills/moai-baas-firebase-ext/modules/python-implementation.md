# Python Cloud Functions Implementation

Complete Python examples for Firebase Cloud Functions.

## Real-Time Data Synchronization

```python
from firebase_functions import https_fn, firestore_fn, auth_fn, storage_fn
from firebase_admin import firestore, auth, storage
from google.cloud import pubsub_v1
from datetime import datetime, timedelta
import json
import os

# Real-time data synchronization function
@https_fn.on_request()
def sync_realtime_data(request: https_fn.Request) -> https_fn.Response:
    """Handle real-time data synchronization requests."""

    try:
        # Parse request data
        data = request.get_json()

        # Validate request
        if not data or 'collection' not in data or 'document' not in data:
            return https_fn.Response(
                json.dumps({"error": "Missing required fields"}),
                status=400,
                mimetype="application/json"
            )

        # Get Firestore client
        db = firestore.client()

        # Get document reference
        doc_ref = db.collection(data['collection']).document(data['document'])

        # Update document with timestamp
        doc_ref.set({
            'data': data.get('data', {}),
            'updated_at': datetime.utcnow(),
            'sync_source': data.get('source', 'unknown'),
        }, merge=True)

        # Trigger real-time update notification
        pubsub_client = pubsub_v1.PublisherClient()
        topic_path = pubsub_client.topic_path(
            os.environ.get('GCP_PROJECT', 'default-project'),
            'realtime-updates'
        )

        pubsub_client.publish(
            topic_path,
            data=json.dumps({
                'collection': data['collection'],
                'document': data['document'],
                'timestamp': datetime.utcnow().isoformat(),
            }).encode('utf-8')
        )

        return https_fn.Response(
            json.dumps({"success": True, "message": "Data synchronized successfully"}),
            status=200,
            mimetype="application/json"
        )

    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )
```

## User Management Function

```python
@auth_fn.on_user_created
def new_user_created(user: auth_fn.AuthEvent) -> None:
    """Handle new user creation."""

    try:
        # Get Firestore client
        db = firestore.client()

        # Create user profile
        db.collection('users').document(user.uid).set({
            'email': user.email,
            'display_name': user.display_name,
            'photo_url': user.photo_url,
            'email_verified': user.email_verified,
            'created_at': datetime.utcnow(),
            'last_login': datetime.utcnow(),
            'preferences': {
                'notifications': True,
                'theme': 'light',
                'language': 'en',
            },
            'subscription_tier': 'free',
        })

        # Create initial user statistics
        db.collection('user_stats').document(user.uid).set({
            'documents_created': 0,
            'collaborations': 0,
            'last_activity': datetime.utcnow(),
        })

    except Exception as e:
        print(f"Error creating user profile: {e}")
```

## Storage File Processing

```python
@storage_fn.on_object_finalized()
def process_uploaded_file(event: storage_fn.CloudEvent) -> None:
    """Process uploaded files and extract metadata."""

    try:
        # Get file information
        file_path = event.data.name
        bucket_name = event.data.bucket

        # Get storage client
        storage_client = storage.Client()
        bucket = storage_client.bucket(bucket_name)
        blob = bucket.blob(file_path)

        # Extract metadata
        metadata = blob.metadata or {}

        # Get Firestore client
        db = firestore.client()

        # Create file record
        db.collection('files').document(blob.name).set({
            'name': blob.name,
            'content_type': blob.content_type,
            'size': blob.size,
            'created': blob.time_created,
            'updated': blob.updated,
            'metadata': metadata,
            'public_url': blob.public_url,
            'processed': True,
        })

        # If image, generate thumbnail
        if blob.content_type.startswith('image/'):
            generate_thumbnail(blob.name)

    except Exception as e:
        print(f"Error processing file {event.data.name}: {e}")


def generate_thumbnail(file_path: str):
    """Generate thumbnail for uploaded images."""
    try:
        db = firestore.client()

        db.collection('files').document(file_path).update({
            'thumbnail_generated': True,
            'thumbnail_url': f"https://storage.googleapis.com/thumbnails/{file_path}",
        })

    except Exception as e:
        print(f"Error generating thumbnail for {file_path}: {e}")
```

## Automated Backup Function

```python
@https_fn.on_request(schedule="0 2 * * *")  # Daily at 2 AM
def automated_backup(request: https_fn.Request) -> https_fn.Response:
    """Perform automated database backup."""

    try:
        # Get Firestore client
        db = firestore.client()

        # Get backup configuration
        backup_config = db.collection('config').document('backup').get().to_dict()

        if not backup_config or not backup_config.get('enabled', False):
            return https_fn.Response("Backup disabled", status=200)

        # Create backup record
        backup_ref = db.collection('backups').document()
        backup_ref.set({
            'created_at': datetime.utcnow(),
            'status': 'in_progress',
            'type': 'automated',
            'config': backup_config,
        })

        # Perform backup
        collections_to_backup = backup_config.get('collections', [])
        backup_data = {}

        for collection_name in collections_to_backup:
            collection_ref = db.collection(collection_name)
            docs = collection_ref.stream()

            backup_data[collection_name] = [
                {**doc.to_dict(), 'id': doc.id} for doc in docs
            ]

        # Store backup in Cloud Storage
        storage_client = storage.Client()
        bucket = storage_client.bucket(backup_config['storage_bucket'])

        backup_blob = bucket.blob(f"backups/{backup_ref.id}.json")
        backup_blob.upload_from_string(
            json.dumps(backup_data, default=str),
            content_type='application/json'
        )

        # Update backup record
        backup_ref.update({
            'status': 'completed',
            'completed_at': datetime.utcnow(),
            'storage_path': backup_blob.name,
            'document_count': sum(len(docs) for docs in backup_data.values()),
        })

        # Clean old backups
        clean_old_backups(backup_config['retention_days'])

        return https_fn.Response(
            json.dumps({
                "success": True,
                "backup_id": backup_ref.id,
                "document_count": sum(len(docs) for docs in backup_data.values())
            }),
            status=200,
            mimetype="application/json"
        )

    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )


def clean_old_backups(retention_days: int):
    """Clean old backup files."""
    try:
        cutoff_date = datetime.utcnow() - timedelta(days=retention_days)

        db = firestore.client()
        old_backups = db.collection('backups').where(
            'created_at', '<', cutoff_date
        ).stream()

        storage_client = storage.Client()
        bucket = storage_client.bucket(os.environ.get('BACKUP_BUCKET'))

        for backup in old_backups:
            # Delete from Cloud Storage
            backup_path = backup.to_dict().get('storage_path')
            if backup_path:
                blob = bucket.blob(backup_path)
                blob.delete()

            # Delete from Firestore
            db.collection('backups').document(backup.id).delete()

    except Exception as e:
        print(f"Error cleaning old backups: {e}")
```

## Scheduled Aggregation Function

```python
@https_fn.on_request(schedule="0 */6 * * *")  # Every 6 hours
def aggregate_user_statistics(request: https_fn.Request) -> https_fn.Response:
    """Aggregate user statistics and activity metrics."""

    try:
        db = firestore.client()

        # Get all user statistics
        users_ref = db.collection('user_stats').stream()

        aggregated_stats = {
            'total_users': 0,
            'total_documents': 0,
            'total_collaborations': 0,
            'active_users': 0,
            'timestamp': datetime.utcnow(),
        }

        for user_doc in users_ref:
            user_data = user_doc.to_dict()
            aggregated_stats['total_users'] += 1
            aggregated_stats['total_documents'] += user_data.get('documents_created', 0)
            aggregated_stats['total_collaborations'] += user_data.get('collaborations', 0)

            # Check if user was active in last 24 hours
            last_activity = user_data.get('last_activity')
            if last_activity:
                time_diff = datetime.utcnow() - last_activity
                if time_diff.total_seconds() < 86400:  # 24 hours
                    aggregated_stats['active_users'] += 1

        # Store aggregated statistics
        db.collection('analytics').document('daily_stats').set({
            **aggregated_stats,
            'period': 'daily',
        }, merge=True)

        return https_fn.Response(
            json.dumps({
                "success": True,
                "stats": aggregated_stats
            }),
            status=200,
            mimetype="application/json"
        )

    except Exception as e:
        return https_fn.Response(
            json.dumps({"error": str(e)}),
            status=500,
            mimetype="application/json"
        )
```

## Batch Data Migration

```python
def migrate_collection_data(
    source_collection: str,
    target_collection: str,
    transform_fn: callable = None
) -> dict:
    """Migrate data from one collection to another with optional transformation."""

    try:
        db = firestore.client()

        # Get source documents
        source_docs = db.collection(source_collection).stream()

        migrated_count = 0
        errors = []

        for doc in source_docs:
            try:
                data = doc.to_dict()

                # Apply transformation if provided
                if transform_fn:
                    data = transform_fn(data)

                # Write to target collection
                db.collection(target_collection).document(doc.id).set(data)
                migrated_count += 1

            except Exception as doc_error:
                errors.append({
                    'doc_id': doc.id,
                    'error': str(doc_error)
                })

        return {
            'success': len(errors) == 0,
            'migrated_count': migrated_count,
            'errors': errors,
        }

    except Exception as e:
        return {
            'success': False,
            'error': str(e),
        }
```

---

## Best Practices

### Error Handling Pattern

```python
def safe_firebase_operation(operation_name: str, operation_fn: callable):
    """Wrap Firebase operation with error handling and logging."""
    try:
        result = operation_fn()
        print(f"✓ {operation_name} succeeded")
        return result
    except Exception as e:
        print(f"✗ {operation_name} failed: {e}")
        # Log to Cloud Logging
        import logging
        logging.error(f"Firebase operation {operation_name} failed: {e}")
        raise
```

### Timeout Management

```python
from google.cloud.firestore import Client
from concurrent.futures import TimeoutError

db = Client()

# Set timeout for operations
try:
    result = db.collection('users').document('user123').get(timeout=5)
except TimeoutError:
    print("Operation timed out after 5 seconds")
```

### Batch Size Optimization

```python
def batch_write_with_chunking(
    collection: str,
    documents: list,
    batch_size: int = 500
) -> int:
    """Write documents in batches to respect Firestore limits."""

    db = firestore.client()
    total_written = 0

    for i in range(0, len(documents), batch_size):
        batch = db.batch()
        chunk = documents[i:i+batch_size]

        for doc in chunk:
            doc_ref = db.collection(collection).document(doc['id'])
            batch.set(doc_ref, doc['data'])

        batch.commit()
        total_written += len(chunk)

    return total_written
```
