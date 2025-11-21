---
name: moai-baas-firebase-functions
description: Cloud Functions for Firebase with triggers and integration patterns
---

## Cloud Functions & Serverless Integration

### Advanced Cloud Functions with Python

```python
# Advanced Cloud Functions with Python
from firebase_functions import https_fn, firestore_fn, scheduler_fn, storage_fn
from firebase_admin import initialize_app, firestore
from google.cloud.firestore_v1.base_query import FieldFilter

initialize_app()

# Real-time data synchronization function
@firestore_fn.on_document_written(document="users/{userId}")
def sync_user_data(event: firestore_fn.Event[firestore_fn.Change]) -> None:
    """Synchronize user data across collections on write."""
    
    user_id = event.params["userId"]
    
    # Get new data
    new_data = event.data.after.to_dict() if event.data.after else None
    old_data = event.data.before.to_dict() if event.data.before else None
    
    if new_data is None:
        # Document deleted
        return
    
    # Update denormalized user info in related collections
    db = firestore.client()
    
    # Update orders
    orders_ref = db.collection("orders")
    orders_query = orders_ref.where(filter=FieldFilter("userId", "==", user_id))
    
    batch = db.batch()
    for order_doc in orders_query.stream():
        batch.update(order_doc.reference, {
            "userName": new_data.get("name"),
            "userEmail": new_data.get("email"),
        })
    
    batch.commit()

# HTTP callable function with authentication
@https_fn.on_call()
def process_payment(req: https_fn.CallableRequest) -> dict:
    """Process payment with Stripe integration."""
    
    # Verify authentication
    if not req.auth:
        raise https_fn.HttpsError(
            code=https_fn.FunctionsErrorCode.UNAUTHENTICATED,
            message="User must be authenticated"
        )
    
    # Extract payment data
    payment_data = req.data
    amount = payment_data.get("amount")
    currency = payment_data.get("currency", "usd")
    
    # Validate input
    if not amount or amount <= 0:
        raise https_fn.HttpsError(
            code=https_fn.FunctionsErrorCode.INVALID_ARGUMENT,
            message="Invalid payment amount"
        )
    
    # Process payment (Stripe integration)
    try:
        # Stripe payment processing here
        payment_intent = {
            "id": "pi_123456",
            "amount": amount,
            "currency": currency,
            "status": "succeeded",
        }
        
        # Log payment
        db = firestore.client()
        db.collection("payments").add({
            "userId": req.auth.uid,
            "amount": amount,
            "currency": currency,
            "status": "completed",
            "timestamp": firestore.SERVER_TIMESTAMP,
        })
        
        return {
            "success": True,
            "paymentIntent": payment_intent,
        }
    except Exception as e:
        raise https_fn.HttpsError(
            code=https_fn.FunctionsErrorCode.INTERNAL,
            message=f"Payment processing failed: {str(e)}"
        )

# Storage function for file processing
@storage_fn.on_object_finalized()
def process_uploaded_image(event: storage_fn.CloudEvent[storage_fn.StorageObjectData]) -> None:
    """Process uploaded images (resize, thumbnail generation)."""
    
    file_path = event.data.name
    content_type = event.data.content_type
    
    # Only process images
    if not content_type or not content_type.startswith("image/"):
        return
    
    # Extract file info
    bucket_name = event.data.bucket
    
    # Generate thumbnail (placeholder)
    # In production: use PIL or Cloud Vision API
    thumbnail_path = f"thumbnails/{file_path}"
    
    print(f"Processing image: {file_path}")
    print(f"Thumbnail would be created at: {thumbnail_path}")

# Scheduled function (cron job)
@scheduler_fn.on_schedule(schedule="every day 00:00")
def cleanup_old_data(event: scheduler_fn.ScheduledEvent) -> None:
    """Clean up old data daily."""
    
    from datetime import datetime, timedelta
    
    db = firestore.client()
    
    # Delete old logs (older than 30 days)
    cutoff_date = datetime.now() - timedelta(days=30)
    old_logs_ref = db.collection("logs").where(
        filter=FieldFilter("timestamp", "<", cutoff_date)
    )
    
    # Batch delete
    batch = db.batch()
    deleted_count = 0
    
    for log_doc in old_logs_ref.limit(500).stream():
        batch.delete(log_doc.reference)
        deleted_count += 1
    
    batch.commit()
    print(f"Deleted {deleted_count} old log entries")
```

---

### Trigger Types

**Firestore Triggers**:
- `on_document_created`: New document created
- `on_document_updated`: Document modified
- `on_document_deleted`: Document removed
- `on_document_written`: Any write operation

**Storage Triggers**:
- `on_object_finalized`: File upload complete
- `on_object_deleted`: File removed
- `on_object_archived`: File archived
- `on_object_metadata_updated`: Metadata changed

**Authentication Triggers**:
- `on_user_created`: New user registered
- `on_user_deleted`: User account deleted

**Scheduler Triggers**:
- `on_schedule`: Cron-based execution

---

### Error Handling Best Practices

```python
@https_fn.on_call()
def robust_function(req: https_fn.CallableRequest) -> dict:
    """Function with comprehensive error handling."""
    
    try:
        # Validate auth
        if not req.auth:
            raise https_fn.HttpsError(
                code=https_fn.FunctionsErrorCode.UNAUTHENTICATED,
                message="Authentication required"
            )
        
        # Validate input
        data = req.data
        if not data or "field" not in data:
            raise https_fn.HttpsError(
                code=https_fn.FunctionsErrorCode.INVALID_ARGUMENT,
                message="Missing required field"
            )
        
        # Business logic
        result = process_data(data["field"])
        
        return {"success": True, "result": result}
        
    except https_fn.HttpsError:
        # Re-raise Firebase errors
        raise
    except ValueError as e:
        # Validation errors
        raise https_fn.HttpsError(
            code=https_fn.FunctionsErrorCode.INVALID_ARGUMENT,
            message=str(e)
        )
    except Exception as e:
        # Unexpected errors
        print(f"Unexpected error: {e}")
        raise https_fn.HttpsError(
            code=https_fn.FunctionsErrorCode.INTERNAL,
            message="Internal server error"
        )
```

---

**End of Module** | moai-baas-firebase-functions
