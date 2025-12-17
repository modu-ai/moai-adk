---
name: moai-platform-firestore
description: Firebase Firestore specialist covering NoSQL patterns, real-time sync, offline caching, and Security Rules. Use when building mobile-first apps with offline support, implementing real-time listeners, or configuring Firestore security.
version: 1.0.0
category: platform
updated: 2025-12-07
status: active
tags:
  - firestore
  - firebase
  - nosql
  - realtime
  - offline
  - mobile
context7-libraries:
  - /firebase/firebase-docs
related-skills:
  - moai-platform-firebase-auth
  - moai-lang-flutter
  - moai-lang-typescript
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-firestore: Firebase Firestore Specialist

## Quick Reference (30 seconds)

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

### Context7 Library Access

```python
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/firebase/firebase-docs",
    topic="firestore security-rules offline-persistence cloud-functions indexes",
    tokens=6000
)
```

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

### Real-time Listeners with Metadata

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

### Security Rules

Basic Structure:
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

Role-Based Access with Custom Claims:
```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    function isSignedIn() { return request.auth != null; }
    function isAdmin() { return request.auth.token.admin == true; }

    match /organizations/{orgId} {
      function isMember() {
        return exists(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid));
      }
      function getMemberRole() {
        return get(/databases/$(database)/documents/organizations/$(orgId)/members/$(request.auth.uid)).data.role;
      }
      function isOrgAdmin() { return isMember() && getMemberRole() in ['admin', 'owner']; }

      allow read: if isSignedIn() && isMember();
      allow update: if isOrgAdmin();
      allow delete: if getMemberRole() == 'owner';

      match /members/{memberId} {
        allow read: if isMember();
        allow write: if isOrgAdmin();
      }

      match /projects/{projectId} {
        allow read: if isMember();
        allow create: if isMember() && getMemberRole() in ['admin', 'owner', 'editor'];
        allow update, delete: if isOrgAdmin() || resource.data.createdBy == request.auth.uid;
      }
    }
  }
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

### Cloud Functions V2 Triggers

```typescript
import { onDocumentUpdated } from 'firebase-functions/v2/firestore'
import { onCall, HttpsError } from 'firebase-functions/v2/https'
import { onSchedule } from 'firebase-functions/v2/scheduler'
import { getFirestore, FieldValue } from 'firebase-admin/firestore'

const db = getFirestore()

export const onDocumentUpdate = onDocumentUpdated(
  { document: 'documents/{docId}', region: 'us-central1' },
  async (event) => {
    const before = event.data?.before.data()
    const after = event.data?.after.data()
    if (!before || !after) return

    const batch = db.batch()
    batch.set(db.collection('changes').doc(), {
      documentId: event.params.docId,
      before, after,
      changedAt: FieldValue.serverTimestamp()
    })
    batch.update(db.doc('stats/documents'), {
      totalModifications: FieldValue.increment(1)
    })
    await batch.commit()
  }
)

export const inviteToOrganization = onCall({ region: 'us-central1' }, async (request) => {
  if (!request.auth) throw new HttpsError('unauthenticated', 'Must be signed in')

  const { organizationId, email, role } = request.data
  const memberDoc = await db.doc(`organizations/${organizationId}/members/${request.auth.uid}`).get()

  if (!memberDoc.exists || !['admin', 'owner'].includes(memberDoc.data()?.role)) {
    throw new HttpsError('permission-denied', 'Must be organization admin')
  }

  const invitation = await db.collection('invitations').add({
    organizationId, email, role,
    invitedBy: request.auth.uid,
    createdAt: FieldValue.serverTimestamp(),
    status: 'pending'
  })

  return { invitationId: invitation.id }
})

export const dailyCleanup = onSchedule(
  { schedule: '0 0 * * *', timeZone: 'UTC', region: 'us-central1' },
  async () => {
    const thirtyDaysAgo = new Date()
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

    const oldDocs = await db.collection('tempFiles')
      .where('createdAt', '<', thirtyDaysAgo).limit(500).get()

    const batch = db.batch()
    oldDocs.docs.forEach((doc) => batch.delete(doc.ref))
    await batch.commit()
  }
)
```

---

## Advanced Patterns

### Offline-First React Hook

```typescript
import { useEffect, useState } from 'react'
import { collection, query, where, orderBy, onSnapshot, addDoc, updateDoc, doc, serverTimestamp } from 'firebase/firestore'

export function useTasks(userId: string) {
  const [tasks, setTasks] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!userId) return
    const q = query(collection(db, 'tasks'), where('userId', '==', userId), orderBy('createdAt', 'desc'))

    return onSnapshot(q, { includeMetadataChanges: true }, (snapshot) => {
      setTasks(snapshot.docs.map((doc) => ({
        id: doc.id, ...doc.data(),
        _pending: doc.metadata.hasPendingWrites,
        _fromCache: doc.metadata.fromCache
      })))
      setLoading(false)
    })
  }, [userId])

  const addTask = (title: string) => addDoc(collection(db, 'tasks'), {
    title, completed: false, userId, createdAt: serverTimestamp()
  })

  const toggleTask = (taskId: string, completed: boolean) =>
    updateDoc(doc(db, 'tasks', taskId), { completed, updatedAt: serverTimestamp() })

  return { tasks, loading, addTask, toggleTask }
}
```

### Batch Operations and Transactions

```typescript
import { writeBatch, runTransaction, doc, increment } from 'firebase/firestore'

async function batchUpdate(updates: Array<{ id: string; data: any }>) {
  const batch = writeBatch(db)
  updates.forEach(({ id, data }) => {
    batch.update(doc(db, 'documents', id), { ...data, updatedAt: serverTimestamp() })
  })
  await batch.commit()
}

async function transferCredits(fromUserId: string, toUserId: string, amount: number) {
  await runTransaction(db, async (transaction) => {
    const fromRef = doc(db, 'users', fromUserId)
    const toRef = doc(db, 'users', toUserId)
    const fromDoc = await transaction.get(fromRef)

    if (!fromDoc.exists()) throw new Error('Sender not found')
    if (fromDoc.data().credits < amount) throw new Error('Insufficient credits')

    transaction.update(fromRef, { credits: increment(-amount) })
    transaction.update(toRef, { credits: increment(amount) })
  })
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

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2025-12-07
Platform: Firebase Firestore
