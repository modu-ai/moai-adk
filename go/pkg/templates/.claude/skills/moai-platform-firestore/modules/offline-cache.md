# Firestore Offline Cache Module

## Overview

Firestore provides robust offline support through local caching, enabling apps to work seamlessly without network connectivity. Data is cached locally and synchronized automatically when the connection is restored.

---

## Cache Configuration

### Web SDK (Firebase v9+)

Memory Cache (Default):
```typescript
import { initializeFirestore, memoryLocalCache } from 'firebase/firestore'

const db = initializeFirestore(app, {
  localCache: memoryLocalCache()
})
```

Persistent Cache with Multi-Tab Support:
```typescript
import {
  initializeFirestore,
  persistentLocalCache,
  persistentMultipleTabManager,
  CACHE_SIZE_UNLIMITED
} from 'firebase/firestore'

const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentMultipleTabManager(),
    cacheSizeBytes: CACHE_SIZE_UNLIMITED
  })
})
```

Persistent Cache with Single Tab:
```typescript
import {
  initializeFirestore,
  persistentLocalCache,
  persistentSingleTabManager,
} from 'firebase/firestore'

const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentSingleTabManager({
      forceOwnership: true
    }),
    cacheSizeBytes: 100 * 1024 * 1024  // 100MB
  })
})
```

### React Native / Mobile

```typescript
import firestore from '@react-native-firebase/firestore'

firestore().settings({
  persistence: true,
  cacheSizeBytes: firestore.CACHE_SIZE_UNLIMITED
})
```

### Flutter

```dart
import 'package:cloud_firestore/cloud_firestore.dart';

FirebaseFirestore.instance.settings = const Settings(
  persistenceEnabled: true,
  cacheSizeBytes: Settings.CACHE_SIZE_UNLIMITED,
);
```

---

## Cache Size Management

### Automatic Garbage Collection

When cache exceeds the configured size, Firestore automatically removes the least recently used documents. The default cache size is 40MB.

Recommended Cache Sizes:
- Small apps: 10-40MB (default)
- Medium apps: 100MB
- Data-heavy apps: CACHE_SIZE_UNLIMITED (use with caution)

### Manual Cache Clearing

```typescript
import { clearIndexedDbPersistence, terminate } from 'firebase/firestore'

async function clearCache() {
  await terminate(db)
  await clearIndexedDbPersistence(db)
  window.location.reload()
}
```

---

## Sync Status Handling

### Document Metadata

```typescript
import { onSnapshot, doc } from 'firebase/firestore'

onSnapshot(doc(db, 'documents', docId), { includeMetadataChanges: true }, (snapshot) => {
  const data = snapshot.data()
  const metadata = snapshot.metadata

  console.log({
    data,
    hasPendingWrites: metadata.hasPendingWrites,
    fromCache: metadata.fromCache
  })
})
```

### Visual Indicators for Sync Status

```typescript
interface SyncStatus {
  synced: boolean
  pending: boolean
  offline: boolean
}

function getSyncStatus(metadata: SnapshotMetadata): SyncStatus {
  return {
    synced: !metadata.hasPendingWrites && !metadata.fromCache,
    pending: metadata.hasPendingWrites,
    offline: metadata.fromCache && !metadata.hasPendingWrites
  }
}

function DocumentStatus({ metadata }: { metadata: SnapshotMetadata }) {
  const status = getSyncStatus(metadata)

  if (status.pending) {
    return <span className="text-yellow-500">Saving...</span>
  }
  if (status.offline) {
    return <span className="text-gray-500">Offline</span>
  }
  return <span className="text-green-500">Synced</span>
}
```

### Snapshots In Sync Detection

```typescript
import { onSnapshotsInSync } from 'firebase/firestore'

const unsubscribe = onSnapshotsInSync(db, () => {
  console.log('All local changes have been synced to server')
})
```

---

## Network State Control

### Programmatic Network Control

```typescript
import { enableNetwork, disableNetwork } from 'firebase/firestore'

async function goOffline() {
  await disableNetwork(db)
  console.log('Network disabled - using cache only')
}

async function goOnline() {
  await enableNetwork(db)
  console.log('Network enabled - syncing changes')
}
```

### Firestore-Aware Network Hook

```typescript
import { useState, useEffect, useCallback } from 'react'
import { enableNetwork, disableNetwork, waitForPendingWrites } from 'firebase/firestore'
import { db } from '@/lib/firebase'

interface FirestoreNetworkState {
  isOnline: boolean
  hasPendingWrites: boolean
  forceOffline: () => Promise<void>
  forceOnline: () => Promise<void>
  waitForSync: () => Promise<void>
}

function useFirestoreNetwork(): FirestoreNetworkState {
  const [isOnline, setIsOnline] = useState(navigator.onLine)
  const [hasPendingWrites, setHasPendingWrites] = useState(false)

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true)
      enableNetwork(db)
    }
    const handleOffline = () => {
      setIsOnline(false)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  const forceOffline = useCallback(async () => {
    await disableNetwork(db)
    setIsOnline(false)
  }, [])

  const forceOnline = useCallback(async () => {
    await enableNetwork(db)
    setIsOnline(true)
  }, [])

  const waitForSync = useCallback(async () => {
    await waitForPendingWrites(db)
    setHasPendingWrites(false)
  }, [])

  return { isOnline, hasPendingWrites, forceOffline, forceOnline, waitForSync }
}
```

---

## Offline Write Handling

### Optimistic Updates

Firestore provides optimistic updates by default:

```typescript
import { addDoc, collection, serverTimestamp } from 'firebase/firestore'

async function createDocument(title: string) {
  const docRef = await addDoc(collection(db, 'documents'), {
    title,
    createdAt: serverTimestamp(),
    status: 'pending'
  })

  console.log('Created document:', docRef.id)
  return docRef.id
}
```

### Wait for Pending Writes

```typescript
import { waitForPendingWrites } from 'firebase/firestore'

async function ensureDataSynced() {
  await waitForPendingWrites(db)
  console.log('All data synced to server')
}

async function beforeLogout() {
  await waitForPendingWrites(db)
  await auth.signOut()
}
```

---

## Cache-First Queries

### Source Options

```typescript
import { getDoc, doc, getDocFromCache, getDocFromServer } from 'firebase/firestore'

// Default: Try cache first, fall back to server
const docSnap = await getDoc(doc(db, 'documents', docId))

// Cache only (throws if not in cache)
const cachedSnap = await getDocFromCache(doc(db, 'documents', docId))

// Server only (ignores cache)
const serverSnap = await getDocFromServer(doc(db, 'documents', docId))
```

### Hybrid Strategy Pattern

```typescript
async function fetchWithFallback<T>(
  ref: DocumentReference,
  transform: (data: DocumentData) => T
): Promise<{ data: T; source: 'cache' | 'server' }> {
  try {
    const cached = await getDocFromCache(ref)
    if (cached.exists()) {
      getDocFromServer(ref).catch(() => {})
      return { data: transform(cached.data()!), source: 'cache' }
    }
  } catch {
    // Cache miss
  }

  const server = await getDocFromServer(ref)
  if (!server.exists()) {
    throw new Error('Document not found')
  }
  return { data: transform(server.data()!), source: 'server' }
}
```

---

## Multi-Tab Synchronization

### Understanding Tab Managers

persistentSingleTabManager:
- Only one tab can access persistence at a time
- Other tabs fall back to memory-only cache
- Use when multi-tab sync is not needed

persistentMultipleTabManager:
- All tabs share the same persistent cache
- Changes in one tab appear in all tabs
- Recommended for most applications

---

## Best Practices

### Cache Warming

```typescript
async function warmCache(userId: string) {
  await getDoc(doc(db, 'users', userId))

  await getDocs(
    query(
      collection(db, 'documents'),
      where('userId', '==', userId),
      orderBy('updatedAt', 'desc'),
      limit(20)
    )
  )

  console.log('Cache warmed for user:', userId)
}
```

### Offline-Aware UI

```typescript
function OfflineIndicator() {
  const { isOnline, hasPendingWrites } = useFirestoreNetwork()

  if (!isOnline) {
    return (
      <div className="bg-yellow-100 p-2 text-center">
        You are offline. Changes will sync when reconnected.
      </div>
    )
  }

  if (hasPendingWrites) {
    return (
      <div className="bg-blue-100 p-2 text-center">
        Syncing changes...
      </div>
    )
  }

  return null
}
```

---

Status: Module Documentation
Last Updated: 2026-01-06
