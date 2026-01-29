# Firestore Reference Documentation

## Data Modeling Best Practices

### Document Structure Guidelines

Flat Data Structures:
- Firestore charges per document read, not per field
- Prefer flat structures over deeply nested data
- Use subcollections for related data that grows unbounded

Denormalization Strategies:
- Duplicate frequently accessed data to reduce reads
- Use Cloud Functions to maintain denormalized data consistency
- Balance read efficiency against write complexity

Document Size Limits:
- Maximum document size: 1MB
- Maximum field depth: 20 levels
- Maximum array elements: 40,000 (practical limit for querying)

### Collection Naming Conventions

Root Collections:
- Use plural nouns: users, documents, organizations
- Lowercase with no special characters
- Keep names short but descriptive

Subcollections:
- Represent one-to-many relationships
- Example: /users/{userId}/notifications/{notificationId}
- Enable collection group queries when needed

---

## Query Optimization

### Index Management

Single-Field Indexes:
- Automatically created for all fields
- Support equality and range queries on single fields
- Ascending and descending variants created automatically

Composite Indexes:
- Required for queries with multiple equality or range conditions
- Order matters: equality fields first, then range fields
- Firebase console shows required indexes in error messages

Index Exemptions:
- Exclude fields never used in queries
- Reduce index storage costs
- Configure in firestore.indexes.json

### Query Performance Tips

Limit Result Sets:
- Always use limit() to control result size
- Implement pagination with startAfter() or startAt()
- Avoid fetching more data than displayed

Cursor-Based Pagination:
- More efficient than offset-based pagination
- Use document snapshots as cursors
- Maintain consistent ordering across pages

Query Constraints:
- Maximum 30 disjunction clauses (OR conditions)
- Maximum 10 in/array-contains-any values
- Range filters on single field only

---

## Security Rules Reference

### Request Object Properties

request.auth:
- uid: Authenticated user's ID
- token: JWT claims including custom claims
- token.email: User's email adddess
- token.email_verified: Email verification status

request.resource:
- data: Document data being written
- Validate incoming data before write

request.time:
- Current timestamp for time-based rules

### Resource Object Properties

resource.data:
- Current document data (for reads/updates/deletes)
- null for create operations

resource.id:
- Document ID

resource.__name__:
- Full document path

### Helper Functions

exists(path):
- Check if document exists
- Counts as a read operation

get(path):
- Retrieve document data
- Counts as a read operation
- Maximum 10 get() calls per rule evaluation

getAfter(path):
- Get document state after batch/transaction
- Useful for cross-document validation

---

## Cloud Functions Triggers

### Trigger Types

onDocumentCreated:
- Fires when new document is created
- Receives snapshot of created document

onDocumentUpdated:
- Fires when existing document is modified
- Receives before and after snapshots

onDocumentDeleted:
- Fires when document is deleted
- Receives snapshot of deleted document

onDocumentWritten:
- Fires on any document change (create, update, delete)
- Most flexible but less specific

### Trigger Best Practices

Idempotency:
- Design functions to handle duplicate invocations
- Use document IDs or timestamps for deduplication
- Store processing state to detect reruns

Error Handling:
- Return resolved promises for successful operations
- Throw errors to trigger automatic retry
- Use dead-letter queues for persistent failures

Performance:
- Minimize cold start impact with lightweight dependencies
- Use connection pooling for external services
- Batch related operations when possible

---

## Offline Behavior

### Cache Behavior

Memory Cache:
- Fastest access for recently read data
- Cleared when app closes
- Configured via memoryLocalCache()

Persistent Cache:
- Survives app restarts
- Uses IndexedDB in browsers
- SQLite on mobile platforms

### Sync Strategies

Automatic Sync:
- Writes queued when offline
- Synced automatically on reconnection
- Order preserved within single client

Conflict Resolution:
- Last-write-wins for most operations
- Transactions fail if document changed
- Use server timestamps for consistent ordering

### Network State Detection

```typescript
import { onSnapshotsInSync, enableNetwork, disableNetwork } from 'firebase/firestore'

// Detect when all snapshots are synchronized
onSnapshotsInSync(db, () => {
  console.log('All snapshots synchronized')
})

// Programmatically control network state
await disableNetwork(db)  // Force offline mode
await enableNetwork(db)   // Resume online operations
```

---

## Error Handling

### Common Error Codes

permission-denied:
- Security Rules blocked the operation
- Check authentication state and rules logic

not-found:
- Document or collection does not exist
- Verify path and document ID

already-exists:
- Attempted to create existing document
- Use set() with merge option or update()

resource-exhausted:
- Quota exceeded (reads, writes, or storage)
- Implement rate limiting or request batching

unavailable:
- Service temporarily unavailable
- Implement retry with exponential backoff

### Error Recovery Patterns

```typescript
import { FirestoreError } from 'firebase/firestore'

async function resilientOperation() {
  const maxRetries = 3
  let attempt = 0

  while (attempt < maxRetries) {
    try {
      return await performOperation()
    } catch (error) {
      if (error instanceof FirestoreError) {
        if (error.code === 'unavailable') {
          attempt++
          await new Promise(r => setTimeout(r, Math.pow(2, attempt) * 1000))
          continue
        }
      }
      throw error
    }
  }
}
```

---

## Migration and Backup

### Data Export

Firebase Console:
- Export to Cloud Storage bucket
- Full database or specific collections
- Scheduled exports available

gcloud CLI:
```bash
gcloud firestore export gs://bucket-name/backup-folder
```

### Data Import

```bash
gcloud firestore import gs://bucket-name/backup-folder
```

### Schema Evolution

Adding Fields:
- New fields can be added without migration
- Set default values in application code
- Use Cloud Functions for backfill if needed

Removing Fields:
- Remove from application code first
- Clean up existing data with batch operations
- Update Security Rules to reject old field names

Renaming Fields:
- Create new field with new name
- Copy data from old to new field
- Remove old field after migration complete

---

## Monitoring and Debugging

### Firebase Console Tools

Usage Dashboard:
- Real-time read/write/delete metrics
- Storage consumption tracking
- Quota usage monitoring

Rules Playground:
- Test Security Rules without deployment
- Simulate authenticated requests
- Debug complex rule logic

### Cloud Logging Integration

Enable Logging:
- Firestore operations logged to Cloud Logging
- Filter by collection, operation type, or error code
- Set up alerts for anomalies

Query Debugging:
- Log slow queries exceeding thresholds
- Identify missing indexes
- Track query performance over time

---

Status: Reference Documentation
Last Updated: 2026-01-06
