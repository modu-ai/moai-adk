# Firebase Firestore Specialist

## Core Capabilities

Real-time Sync provides automatic synchronization across all connected clients. Changes propagate to all listeners within 100-500ms.

Offline Caching uses IndexedDB persistence with automatic sync when online. Supports multi-tab windows with shared cache.

Security Rules offer declarative field-level access control based on authentication, document data, and request context.

Cloud Functions enable document triggers for server-side processing. Triggers fire on document create, update, delete operations.

Composite Indexes support complex query optimization for multi-field queries with range filters and ordering.

Mobile SDKs provide first-party support for iOS, Android, Web, and Flutter with unified API.

## Initialization and Setup

Package Installation requires firebase/app and firebase/firestore for Web SDK. For mobile, use firebase-ios-sdk, firebase-android-sdk, or cloud_firestore for Flutter.

Firebase Configuration requires importing initializeApp from firebase/app and initializeFirestore from firebase/firestore. Create app with initializeApp passing config object containing apiKey, authDomain, and projectId from environment variables.

Offline Persistence requires importing persistentLocalCache, persistentMultipleTabManager, and CACHE_SIZE_UNLIMITED from firebase/firestore. Export db instance created with initializeFirestore, passing app and localCache configuration using persistentLocalCache with tabManager set to persistentMultipleTabManager() and cacheSizeBytes set to CACHE_SIZE_UNLIMITED.

## Security Rules

Basic Rule Structure uses rules_version 2 and service cloud.firestore. In databases match block, create collection match with documentId wildcard.

Role-based Access Control allows read/write if request.auth.uid equals userId field. Use request.auth != null to check authentication.

Custom Claims Integration uses request.auth.token.claimName to access custom claims from Firebase Auth. Useful for role-based access.

Field-level Validation uses request.resource.data.fieldName to validate incoming data. Example: allow create if request.resource.data.price is number.

Condition Evaluation supports &&, ||, and ! operators. Use exists(), has(), and is() functions for complex conditions.

Public/Private Data pattern allows read if document.isPublic is true or if request.auth.uid equals document.ownerId.

Testing Rules uses Firebase Rules Simulator in console or local testing with @firebase/rules-unit-testing library.

## Real-time Listeners

Snapshot Listener uses onSnapshot from firebase/firestore. Query with collection, query, where, and orderBy functions.

Metadata Handling includes includeMetadataChanges parameter set to true to detect pending writes and cached results. Access doc.metadata.hasPendingWrites and doc.metadata.fromCache.

Subscription Optimization involves limiting listener scope with specific queries. Use where clauses to reduce data transfer.

Listener Lifecycle requires storing unsubscribe function returned by onSnapshot. Call unsubscribe on component unmount to prevent memory leaks.

Error Handling includes handling permission errors, network errors, and retry logic. Implement exponential backoff for failed listeners.

Real-time Updates provide snapshot.docChanges() array with change type (added, modified, removed). Use for incremental UI updates.

## Offline and Caching

Persistent Cache Configuration uses persistentLocalCache with persistentMultipleTabManager for multi-tab support. Set cacheSizeBytes to CACHE_SIZE_UNLIMITED or custom size.

Cache Sync Status uses fromCache metadata to determine if data is from cache or server. Handle pending writes with hasPendingWrites flag.

Multi-tab Manager shares IndexedDB cache across browser tabs. Prevents duplicate network requests and race conditions.

Network State Handling uses onSnapshotsInSync() to detect when all listeners are in sync. Useful for loading indicators.

Cache Throttling implements custom throttling for offline writes. Queue mutations and sync when connection restored.

## Transactions and Batches

Atomic Operations use runTransaction from firebase/firestore. Transaction provides access to transaction.get(), transaction.set(), transaction.update(), transaction.delete().

Batched Writes use writeBatch() for multiple operations without automatic retry. Supports up to 500 operations per batch.

Transaction Constraints require all reads to happen before writes. Cannot read documents after write operations in same transaction.

Distributed Counter uses shard pattern with counter documents. Increment random shard in transaction for conflict-free counting.

Conflict Resolution relies on automatic retry for failed transactions. Implement exponential backoff and max retry limits.

Performance considerations include keeping transactions short, minimizing read operations, and avoiding hot documents.

## Query Optimization

Composite Indexes require creating indexes in firestore.indexes.json for queries with multiple range filters or ordering with range filter.

Index Ordering uses orderBy() for sorting. Multiple orderBy calls create composite index requirement.

Pagination uses startAfter(), startAt(), endBefore(), and endAt() with document snapshots for cursor-based pagination.

Query Limits use limit() and limitToLast() to constrain result set. Always apply reasonable limits for performance.

Collection Group Queries use collectionGroup() to query across all collections with same ID. Requires index for large datasets.

In Operator uses where('field', 'in', array) for up to 10 values. For larger sets, use array-contains-any with separate array field.

## Data Modeling

Document Structure favors denormalized data over normalized for Firestore. Embed related data in documents for better read performance.

Subcollections organize related documents under parent document. Querying subcollection requires parent document reference.

Array Fields use array-contains and array-contains-any queries. Avoid arrays that grow unboundedly.

Map Fields use dot notation for nested field queries. Avoid deeply nested maps for query performance.

Reference Fields use doc() references for relationships. Require additional queries to load referenced documents.

## Performance and Pricing

Read Latency ranges from 50 to 200ms depending on region and query complexity.

Write Latency ranges from 100 to 300ms depending on document size and indexes.

Real-time Propagation takes 100 to 500ms for changes to reach all connected clients.

Offline Sync occurs automatically on reconnection. Includes conflict resolution for concurrent edits.

Free Tier (2024) provides 1GB storage, 50K daily reads, 20K daily writes, 20K daily deletes. Real-time listeners included.

Pricing Scale charges per document read, write, and delete. Storage per GB-month. Network egress applies.

## Mobile SDK Patterns

Flutter SDK uses cloud_firestore package with FirebaseFirestore.instance. Enable persistence with FirebaseFirestore.instance.settings = Settings(persistenceEnabled: true).

React Native uses @react-native-firebase/firestore with automatic persistence. Use onSnapshot for real-time listeners.

Swift SDK uses Firestore.firestore() with settings for persistence. Use addSnapshotListener for real-time updates.

Kotlin SDK uses FirebaseFirestore.getInstance() with setFirestoreSettings. Use addSnapshotListener for real-time data.

## When to Use Firestore

Mobile-First Applications benefit from offline-first architecture and mobile-optimized SDKs.

Real-time Collaboration features are built-in with listeners and presence tracking.

Cross-Platform Apps use unified SDKs across iOS, Android, Web, and Flutter.

Flexible Schemas benefit from NoSQL document model with optional field validation via Security Rules.

Google Cloud Integration provides seamless integration with other Firebase services and Google Cloud.

## When to Consider Alternatives

Need SQL Features: Consider PostgreSQL options (Neon, Supabase) for complex joins, transactions, and relational queries.

Need Full-Text Search: Consider dedicated search service or PostgreSQL with full-text search.

Need Strong Consistency: Firestore provides eventual consistency for queries. Consider SQL database for strong consistency.

Need Complex Aggregations: Firestore aggregation capabilities are limited. Consider SQL database for complex analytics.

---

Status: Production Ready
Platform: Firebase Firestore
Version: 2.1.0
Coverage: NoSQL, Real-time, Offline, Mobile SDKs
