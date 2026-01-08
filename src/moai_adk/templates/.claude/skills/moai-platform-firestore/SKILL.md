---
name: "moai-platform-firestore"
description: "Firebase Firestore specialist covering NoSQL patterns, real-time sync, offline caching, and Security Rules. Use when building mobile-first apps with offline support, implementing real-time listeners, or configuring Firestore security."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['firestore', 'firebase', 'nosql', 'realtime', 'offline', 'mobile']
updated: 2026-01-08
status: "active"
context7-libraries: ['/firebase/firebase-docs']
related-skills: ['moai-platform-firebase-auth', 'moai-lang-flutter', 'moai-lang-typescript']
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-firestore: Firebase Firestore Specialist

## Quick Reference

Firebase Firestore Expertise: NoSQL document database with real-time synchronization, offline-first architecture, Security Rules, Cloud Functions triggers, and mobile-optimized SDKs.

### Core Capabilities

Real-time Sync: Automatic synchronization across all connected clients
Offline Caching: IndexedDB persistence with automatic sync when online
Security Rules: Declarative field-level access control
Cloud Functions: Document triggers for server-side processing
Composite Indexes: Complex query optimization

### When to Use Firestore

- Mobile-first applications with offline support
- Real-time collaborative features
- Cross-platform apps (iOS, Android, Web, Flutter)
- Projects requiring Google Cloud integration
- Apps with flexible, evolving data structures

### Context7 Documentation Access

Step 1: Resolve the Firebase library ID
- Use mcp__context7__resolve-library-id with libraryName "firebase"
- Returns the Context7-compatible library ID

Step 2: Fetch Firestore documentation
- Use mcp__context7__get-library-docs with the resolved ID
- Set topic to specific area: "firestore security-rules", "firestore offline", "firestore real-time"
- Allocate 6000-8000 tokens for comprehensive coverage

### Module Index

This skill is organized into specialized modules for deep implementation guidance:

Security Rules Module (modules/security-rules.md):
- Basic rule structure and syntax
- Role-based access control patterns
- Custom claims integration
- Field-level validation
- Testing and debugging rules

Offline and Caching Module (modules/offline-cache.md):
- Persistent cache configuration
- Multi-tab manager setup
- Cache size optimization
- Sync status handling
- Network state management

Real-time Listeners Module (modules/realtime-listeners.md):
- Snapshot listener patterns
- Metadata handling for pending writes
- Query subscription optimization
- Listener lifecycle management
- Error handling strategies

Transactions Module (modules/transactions.md):
- Atomic batch operations
- Transaction patterns and constraints
- Distributed counter implementation
- Conflict resolution strategies
- Performance considerations

---

## Implementation Guide

### Firestore Initialization with Offline Persistence

```typescript
import { initializeApp } from 'firebase/app'
import {
  initializeFirestore,
  persistentLocalCache,
  persistentMultipleTabManager,
  CACHE_SIZE_UNLIMITED
} from 'firebase/firestore'

const app = initializeApp({
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID
})

export const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentMultipleTabManager(),
    cacheSizeBytes: CACHE_SIZE_UNLIMITED
  })
})
```

### Basic Security Rules Structure

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    match /users/{userId} {
      allow read, write: if request.auth.uid == userId;
    }

    match /documents/{docId} {
      allow read: if resource.data.isPublic == true
        || request.auth.uid == resource.data.ownerId
        || request.auth.uid in resource.data.collaborators;
      allow create: if request.auth != null
        && request.resource.data.ownerId == request.auth.uid;
      allow update, delete: if request.auth.uid == resource.data.ownerId;
    }
  }
}
```

### Real-time Listener with Metadata

```typescript
import { collection, query, where, orderBy, onSnapshot } from 'firebase/firestore'

export function subscribeToDocuments(userId: string, callback: (docs: any[]) => void) {
  const q = query(
    collection(db, 'documents'),
    where('collaborators', 'array-contains', userId),
    orderBy('createdAt', 'desc')
  )

  return onSnapshot(q, { includeMetadataChanges: true }, (snapshot) => {
    callback(snapshot.docs.map((doc) => ({
      id: doc.id,
      ...doc.data(),
      _pending: doc.metadata.hasPendingWrites,
      _fromCache: doc.metadata.fromCache
    })))
  })
}
```

### Composite Indexes Configuration

```json
{
  "indexes": [
    {
      "collectionGroup": "documents",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "organizationId", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "documents",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tags", "arrayConfig": "CONTAINS" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    }
  ]
}
```

---

## Performance and Pricing

### Performance Characteristics

Read Latency: 50-200ms (varies by region)
Write Latency: 100-300ms
Real-time Propagation: 100-500ms
Offline Sync: Automatic on reconnection

### Free Tier (2024)

Storage: 1GB
Daily Reads: 50,000
Daily Writes: 20,000
Daily Deletes: 20,000

---

## Works Well With

- moai-platform-firebase-auth - Firebase Authentication integration
- moai-lang-flutter - Flutter SDK patterns
- moai-lang-typescript - TypeScript client patterns
- moai-domain-mobile - Mobile architecture patterns
- moai-quality-security - Security Rules best practices

---

## Additional Resources

- reference.md - Extended documentation and best practices
- examples.md - Complete working code examples
- modules/security-rules.md - Security Rules deep dive
- modules/offline-cache.md - Offline persistence patterns
- modules/realtime-listeners.md - Real-time subscription patterns
- modules/transactions.md - Batch and transaction operations

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2026-01-06
Platform: Firebase Firestore
