# Firestore Real-time Listeners Module

## Overview

Firestore real-time listeners provide automatic synchronization of data changes across all connected clients. When data changes on the server, all subscribed clients receive updates within milliseconds.

---

## Basic Listener Patterns

### Single Document Listener

```typescript
import { doc, onSnapshot, DocumentSnapshot } from 'firebase/firestore'

function subscribeToDocument(docId: string, callback: (data: any) => void) {
  const docRef = doc(db, 'documents', docId)

  return onSnapshot(docRef, (snapshot: DocumentSnapshot) => {
    if (snapshot.exists()) {
      callback({ id: snapshot.id, ...snapshot.data() })
    } else {
      callback(null)
    }
  })
}

const unsubscribe = subscribeToDocument('doc-123', (data) => {
  console.log('Document updated:', data)
})

unsubscribe()
```

### Collection Listener

```typescript
import { collection, onSnapshot, QuerySnapshot } from 'firebase/firestore'

function subscribeToCollection(collectionPath: string, callback: (docs: any[]) => void) {
  const collectionRef = collection(db, collectionPath)

  return onSnapshot(collectionRef, (snapshot: QuerySnapshot) => {
    const docs = snapshot.docs.map((doc) => ({
      id: doc.id,
      ...doc.data()
    }))
    callback(docs)
  })
}
```

### Query Listener

```typescript
import { collection, query, where, orderBy, limit, onSnapshot } from 'firebase/firestore'

function subscribeToActiveDocuments(userId: string, callback: (docs: any[]) => void) {
  const q = query(
    collection(db, 'documents'),
    where('userId', '==', userId),
    where('status', '==', 'active'),
    orderBy('updatedAt', 'desc'),
    limit(50)
  )

  return onSnapshot(q, (snapshot) => {
    callback(snapshot.docs.map((doc) => ({
      id: doc.id,
      ...doc.data()
    })))
  })
}
```

---

## Metadata and Pending Writes

### Understanding Metadata

```typescript
import { onSnapshot, doc } from 'firebase/firestore'

interface DocumentWithMeta {
  id: string
  data: any
  metadata: {
    hasPendingWrites: boolean
    fromCache: boolean
  }
}

function subscribeWithMetadata(docId: string, callback: (doc: DocumentWithMeta) => void) {
  return onSnapshot(
    doc(db, 'documents', docId),
    { includeMetadataChanges: true },
    (snapshot) => {
      callback({
        id: snapshot.id,
        data: snapshot.data(),
        metadata: {
          hasPendingWrites: snapshot.metadata.hasPendingWrites,
          fromCache: snapshot.metadata.fromCache
        }
      })
    }
  )
}
```

### Metadata Change Events

With `includeMetadataChanges: true`, the listener fires when:
- Document data changes
- Pending write is confirmed (hasPendingWrites: true -> false)
- Data source changes (fromCache: true -> false when synced)

---

## Change Type Detection

### Document Changes in Collections

```typescript
import { collection, onSnapshot, DocumentChange } from 'firebase/firestore'

function subscribeWithChanges(
  collectionPath: string,
  handlers: {
    onAdded: (doc: any) => void
    onModified: (doc: any) => void
    onRemoved: (docId: string) => void
  }
) {
  return onSnapshot(collection(db, collectionPath), (snapshot) => {
    snapshot.docChanges().forEach((change: DocumentChange) => {
      const doc = { id: change.doc.id, ...change.doc.data() }

      switch (change.type) {
        case 'added':
          handlers.onAdded(doc)
          break
        case 'modified':
          handlers.onModified(doc)
          break
        case 'removed':
          handlers.onRemoved(change.doc.id)
          break
      }
    })
  })
}
```

---

## Error Handling

### Basic Error Handler

```typescript
import { onSnapshot, FirestoreError } from 'firebase/firestore'

function subscribeWithErrorHandling(
  ref: any,
  onData: (data: any) => void,
  onError: (error: FirestoreError) => void
) {
  return onSnapshot(
    ref,
    (snapshot) => {
      onData(snapshot.data())
    },
    (error: FirestoreError) => {
      console.error('Listener error:', error.code, error.message)
      onError(error)
    }
  )
}
```

### Common Error Codes

permission-denied:
- User lacks access to the document/collection
- Security Rules blocking the read

unavailable:
- Network issues or service temporarily down
- Firestore will automatically reconnect

cancelled:
- Listener was unsubscribed
- Normal cleanup, not an error

---

## React Hooks for Listeners

### Basic Document Hook

```typescript
import { useState, useEffect } from 'react'
import { doc, onSnapshot, DocumentData } from 'firebase/firestore'
import { db } from '@/lib/firebase'

interface UseDocumentResult<T> {
  data: T | null
  loading: boolean
  error: Error | null
}

function useDocument<T = DocumentData>(
  collectionPath: string,
  docId: string | null
): UseDocumentResult<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!docId) {
      setData(null)
      setLoading(false)
      return
    }

    const docRef = doc(db, collectionPath, docId)

    const unsubscribe = onSnapshot(
      docRef,
      (snapshot) => {
        if (snapshot.exists()) {
          setData({ id: snapshot.id, ...snapshot.data() } as T)
        } else {
          setData(null)
        }
        setLoading(false)
        setError(null)
      },
      (err) => {
        setError(err)
        setLoading(false)
      }
    )

    return () => unsubscribe()
  }, [collectionPath, docId])

  return { data, loading, error }
}
```

### Hook with Metadata

```typescript
interface DocumentWithSync<T> {
  data: T | null
  loading: boolean
  error: Error | null
  isPending: boolean
  isFromCache: boolean
}

function useDocumentWithSync<T>(
  collectionPath: string,
  docId: string | null
): DocumentWithSync<T> {
  const [state, setState] = useState<DocumentWithSync<T>>({
    data: null,
    loading: true,
    error: null,
    isPending: false,
    isFromCache: false
  })

  useEffect(() => {
    if (!docId) {
      setState((s) => ({ ...s, data: null, loading: false }))
      return
    }

    const docRef = doc(db, collectionPath, docId)

    const unsubscribe = onSnapshot(
      docRef,
      { includeMetadataChanges: true },
      (snapshot) => {
        setState({
          data: snapshot.exists() ? ({ id: snapshot.id, ...snapshot.data() } as T) : null,
          loading: false,
          error: null,
          isPending: snapshot.metadata.hasPendingWrites,
          isFromCache: snapshot.metadata.fromCache
        })
      },
      (error) => {
        setState((s) => ({ ...s, loading: false, error }))
      }
    )

    return () => unsubscribe()
  }, [collectionPath, docId])

  return state
}
```

---

## Performance Optimization

### Limiting Query Scope

```typescript
// BAD: Listening to entire collection
onSnapshot(collection(db, 'messages'), callback)

// GOOD: Limit to relevant subset
onSnapshot(
  query(
    collection(db, 'messages'),
    where('roomId', '==', currentRoomId),
    orderBy('createdAt', 'desc'),
    limit(100)
  ),
  callback
)
```

### Batched Updates

```typescript
import { unstable_batchedUpdates } from 'react-dom'

function subscribeWithBatching(collectionPath: string, setItems: (items: any[]) => void) {
  return onSnapshot(collection(db, collectionPath), (snapshot) => {
    unstable_batchedUpdates(() => {
      const items = snapshot.docs.map((doc) => ({
        id: doc.id,
        ...doc.data()
      }))
      setItems(items)
    })
  })
}
```

---

## Advanced Patterns

### Composite Subscriptions

```typescript
function useUserWithOrganizations(userId: string) {
  const [user, setUser] = useState(null)
  const [organizations, setOrganizations] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!userId) return

    let loadedCount = 0
    const checkLoaded = () => {
      loadedCount++
      if (loadedCount === 2) setLoading(false)
    }

    const unsubUser = onSnapshot(doc(db, 'users', userId), (snapshot) => {
      setUser(snapshot.data())
      checkLoaded()
    })

    const unsubOrgs = onSnapshot(
      query(
        collection(db, 'organizations'),
        where('members', 'array-contains', userId)
      ),
      (snapshot) => {
        setOrganizations(snapshot.docs.map((d) => ({ id: d.id, ...d.data() })))
        checkLoaded()
      }
    )

    return () => {
      unsubUser()
      unsubOrgs()
    }
  }, [userId])

  return { user, organizations, loading }
}
```

---

## Best Practices

### Always Unsubscribe

```typescript
useEffect(() => {
  const unsubscribe = onSnapshot(ref, callback)
  return () => unsubscribe()
}, [dependencies])
```

### Handle Loading and Error States

```typescript
function DataComponent() {
  const { data, loading, error } = useDocument('users', userId)

  if (loading) return <Spinner />
  if (error) return <ErrorMessage error={error} />
  if (!data) return <NotFound />

  return <UserProfile user={data} />
}
```

### Debounce Rapid Updates

```typescript
import { debounce } from 'lodash'

function subscribeDebounced(ref: any, callback: (data: any) => void, wait = 100) {
  const debouncedCallback = debounce(callback, wait)

  return onSnapshot(ref, (snapshot) => {
    debouncedCallback(snapshot.data())
  })
}
```

---

Status: Module Documentation
Last Updated: 2026-01-06
