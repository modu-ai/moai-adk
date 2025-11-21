# Firebase Performance Optimization

**Enterprise-grade performance optimization techniques for Firebase applications.**

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Research Base**: Context7 `/llmstxt/firebase_google-llms.txt`, production optimization patterns

---

## 📚 Table of Contents

1. [Query Plan Optimization](#1-query-plan-optimization)
2. [Index Cardinality](#2-index-cardinality)
3. [Read/Write Optimization](#3-readwrite-optimization)
4. [Cold Start Optimization](#4-cold-start-optimization)
5. [Data Denormalization](#5-data-denormalization)
6. [Caching Patterns](#6-caching-patterns)
7. [Cost Optimization](#7-cost-optimization)
8. [Latency Reduction](#8-latency-reduction)

---

## 1. Query Plan Optimization

### Technique 1.1: Efficient Query Design

**Optimized query patterns**:

```typescript
import { collection, query, where, orderBy, limit, getDocs, startAfter } from 'firebase/firestore'

// ❌ BAD: Fetches all documents, filters client-side
async function badGetActiveTasks() {
  const tasksSnapshot = await getDocs(collection(db, 'tasks'))

  return tasksSnapshot.docs
    .map(doc => ({ id: doc.id, ...doc.data() }))
    .filter(task => task.status === 'active')
    .sort((a, b) => b.priority - a.priority)
    .slice(0, 10)
}

// ✅ GOOD: Server-side filtering and sorting
async function goodGetActiveTasks() {
  const tasksRef = collection(db, 'tasks')
  const q = query(
    tasksRef,
    where('status', '==', 'active'),
    orderBy('priority', 'desc'),
    limit(10)
  )

  const snapshot = await getDocs(q)
  return snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}

// ✅ BETTER: Add tenant scoping for multi-tenant
async function betterGetActiveTasks(tenantId: string) {
  const tasksRef = collection(db, 'tasks')
  const q = query(
    tasksRef,
    where('tenantId', '==', tenantId),
    where('status', '==', 'active'),
    orderBy('priority', 'desc'),
    limit(10)
  )

  const snapshot = await getDocs(q)
  return snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}
```

### Technique 1.2: Cursor-Based Pagination

**Efficient pagination**:

```typescript
class PaginatedQuery<T> {
  private lastDoc: any = null
  private hasMore = true

  constructor(
    private collectionPath: string,
    private filters: Array<{ field: string; operator: any; value: any }>,
    private orderField: string,
    private pageSize: number = 20
  ) {}

  async getNextPage(): Promise<T[]> {
    if (!this.hasMore) {
      return []
    }

    const collectionRef = collection(db, this.collectionPath)
    let q = query(collectionRef)

    // Apply filters
    this.filters.forEach(({ field, operator, value }) => {
      q = query(q, where(field, operator, value))
    })

    // Apply ordering and limit
    q = query(q, orderBy(this.orderField, 'desc'), limit(this.pageSize))

    // Apply cursor
    if (this.lastDoc) {
      q = query(q, startAfter(this.lastDoc))
    }

    const snapshot = await getDocs(q)

    if (snapshot.empty || snapshot.docs.length < this.pageSize) {
      this.hasMore = false
    }

    if (snapshot.docs.length > 0) {
      this.lastDoc = snapshot.docs[snapshot.docs.length - 1]
    }

    return snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() } as T))
  }

  reset() {
    this.lastDoc = null
    this.hasMore = true
  }
}

// Usage
const tasksPagination = new PaginatedQuery<Task>(
  'tasks',
  [
    { field: 'tenantId', operator: '==', value: 'tenant-123' },
    { field: 'status', operator: '==', value: 'active' }
  ],
  'createdAt',
  20
)

const firstPage = await tasksPagination.getNextPage()
const secondPage = await tasksPagination.getNextPage()
```

---

## 2. Index Cardinality

### Technique 2.1: Index Selectivity Analysis

**Optimize index performance**:

```typescript
// ❌ BAD: Low cardinality field first (poor selectivity)
// Index: status (3 values) → tenantId (1000 values) → priority (10 values)
const badQuery = query(
  collection(db, 'tasks'),
  where('status', '==', 'active'),      // Low cardinality
  where('tenantId', '==', 'tenant-123'),  // High cardinality
  orderBy('priority', 'desc')
)

// ✅ GOOD: High cardinality field first (better selectivity)
// Index: tenantId (1000 values) → status (3 values) → priority (10 values)
const goodQuery = query(
  collection(db, 'tasks'),
  where('tenantId', '==', 'tenant-123'),  // High cardinality first
  where('status', '==', 'active'),
  orderBy('priority', 'desc')
)
```

**Cardinality analysis**:

```typescript
async function analyzeFieldCardinality(
  collectionPath: string,
  fieldPath: string
): Promise<number> {
  const snapshot = await getDocs(collection(db, collectionPath))

  const uniqueValues = new Set()
  snapshot.forEach(doc => {
    uniqueValues.add(doc.get(fieldPath))
  })

  const cardinality = uniqueValues.size
  const totalDocs = snapshot.size
  const selectivity = cardinality / totalDocs

  console.log(`Field: ${fieldPath}`)
  console.log(`Cardinality: ${cardinality}`)
  console.log(`Selectivity: ${selectivity.toFixed(2)}`)

  return cardinality
}

// Analyze index effectiveness
await analyzeFieldCardinality('tasks', 'tenantId')    // High cardinality
await analyzeFieldCardinality('tasks', 'status')      // Low cardinality
await analyzeFieldCardinality('tasks', 'priority')    // Medium cardinality
```

---

## 3. Read/Write Optimization

### Technique 3.1: Batch Operations

**Minimize network round trips**:

```typescript
import { writeBatch, doc } from 'firebase/firestore'

// ❌ BAD: Individual writes (N network calls)
async function badBulkCreate(tasks: any[]) {
  for (const task of tasks) {
    await setDoc(doc(collection(db, 'tasks')), task)
  }
}

// ✅ GOOD: Batch write (1 network call, up to 500 ops)
async function goodBulkCreate(tasks: any[]) {
  const batch = writeBatch(db)

  tasks.forEach(task => {
    const taskRef = doc(collection(db, 'tasks'))
    batch.set(taskRef, task)
  })

  await batch.commit()
}

// ✅ BETTER: Chunked batches for large datasets
async function betterBulkCreate(tasks: any[], chunkSize = 500) {
  const chunks = []

  for (let i = 0; i < tasks.length; i += chunkSize) {
    chunks.push(tasks.slice(i, i + chunkSize))
  }

  for (const chunk of chunks) {
    const batch = writeBatch(db)

    chunk.forEach(task => {
      const taskRef = doc(collection(db, 'tasks'))
      batch.set(taskRef, task)
    })

    await batch.commit()
  }
}
```

### Technique 3.2: Read Batching

**Optimize multiple document reads**:

```typescript
import { documentId, getDocs } from 'firebase/firestore'

// ❌ BAD: Individual reads (N network calls)
async function badGetMultipleTasks(taskIds: string[]) {
  const tasks = []

  for (const id of taskIds) {
    const taskDoc = await getDoc(doc(db, 'tasks', id))
    if (taskDoc.exists()) {
      tasks.push({ id: taskDoc.id, ...taskDoc.data() })
    }
  }

  return tasks
}

// ✅ GOOD: Batch read with 'in' query (1 network call, up to 10 IDs)
async function goodGetMultipleTasks(taskIds: string[]) {
  const tasksRef = collection(db, 'tasks')
  const q = query(tasksRef, where(documentId(), 'in', taskIds))

  const snapshot = await getDocs(q)
  return snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}

// ✅ BETTER: Chunked batch reads for >10 IDs
async function betterGetMultipleTasks(taskIds: string[]) {
  const chunks = []

  for (let i = 0; i < taskIds.length; i += 10) {
    chunks.push(taskIds.slice(i, i + 10))
  }

  const allTasks = []

  for (const chunk of chunks) {
    const tasksRef = collection(db, 'tasks')
    const q = query(tasksRef, where(documentId(), 'in', chunk))
    const snapshot = await getDocs(q)

    allTasks.push(...snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() })))
  }

  return allTasks
}
```

---

## 4. Cold Start Optimization

### Technique 4.1: Cloud Functions Warmup

**Reduce cold starts**:

```typescript
// functions/src/index.ts
import { onRequest } from 'firebase-functions/v2/https'
import { onSchedule } from 'firebase-functions/v2/scheduler'

// Keep function warm with scheduled invocations
export const keepWarm = onSchedule('every 5 minutes', async () => {
  console.log('Warmup ping')
  return null
})

// Optimized function initialization
let dbInitialized = false
let db: any

function getDb() {
  if (!dbInitialized) {
    const admin = require('firebase-admin')
    if (!admin.apps.length) {
      admin.initializeApp()
    }
    db = admin.firestore()
    dbInitialized = true
  }
  return db
}

// HTTP function with warm initialization
export const processTask = onRequest(async (req, res) => {
  const firestore = getDb()  // Reuse initialized instance

  // Your logic here
  const result = await firestore.collection('tasks').doc(req.body.taskId).get()

  res.json(result.data())
})
```

### Technique 4.2: Connection Pooling

**Optimize database connections**:

```typescript
// Singleton pattern for Firebase Admin
class FirebaseAdmin {
  private static instance: FirebaseAdmin
  private db: any

  private constructor() {
    const admin = require('firebase-admin')

    if (!admin.apps.length) {
      admin.initializeApp({
        credential: admin.credential.applicationDefault()
      })
    }

    this.db = admin.firestore()
  }

  static getInstance(): FirebaseAdmin {
    if (!FirebaseAdmin.instance) {
      FirebaseAdmin.instance = new FirebaseAdmin()
    }
    return FirebaseAdmin.instance
  }

  getFirestore() {
    return this.db
  }
}

// Usage
const admin = FirebaseAdmin.getInstance()
const db = admin.getFirestore()
```

---

## 5. Data Denormalization

### Technique 5.1: Strategic Denormalization

**Optimize for read performance**:

```typescript
// ❌ BAD: Normalized data (multiple reads)
interface NormalizedTask {
  id: string
  title: string
  assignedToId: string  // Reference only
  projectId: string     // Reference only
}

async function badGetTaskWithDetails(taskId: string) {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))
  const task = taskDoc.data()

  // Additional reads
  const userDoc = await getDoc(doc(db, 'users', task.assignedToId))
  const projectDoc = await getDoc(doc(db, 'projects', task.projectId))

  return {
    ...task,
    assignedTo: userDoc.data(),
    project: projectDoc.data()
  }
}

// ✅ GOOD: Denormalized data (single read)
interface DenormalizedTask {
  id: string
  title: string
  assignedToId: string
  assignedToName: string      // Denormalized
  assignedToEmail: string     // Denormalized
  assignedToAvatar: string    // Denormalized
  projectId: string
  projectName: string         // Denormalized
  projectStatus: string       // Denormalized
}

async function goodGetTaskWithDetails(taskId: string) {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))
  return { id: taskDoc.id, ...taskDoc.data() }  // Single read
}

// Keep denormalized data in sync
async function updateUserProfile(userId: string, updates: any) {
  const batch = writeBatch(db)

  // Update user profile
  batch.update(doc(db, 'users', userId), updates)

  // Update denormalized data in tasks
  const tasksSnapshot = await getDocs(
    query(collection(db, 'tasks'), where('assignedToId', '==', userId))
  )

  tasksSnapshot.forEach(taskDoc => {
    batch.update(taskDoc.ref, {
      assignedToName: updates.name,
      assignedToEmail: updates.email,
      assignedToAvatar: updates.avatar
    })
  })

  await batch.commit()
}
```

---

## 6. Caching Patterns

### Technique 6.1: Client-Side Cache with TTL

**Reduce read costs**:

```typescript
class FirestoreCache<T> {
  private cache = new Map<string, { data: T; expiry: number }>()
  private ttl: number

  constructor(ttlMinutes = 5) {
    this.ttl = ttlMinutes * 60 * 1000
  }

  set(key: string, data: T) {
    this.cache.set(key, {
      data,
      expiry: Date.now() + this.ttl
    })
  }

  get(key: string): T | null {
    const cached = this.cache.get(key)

    if (!cached) {
      return null
    }

    if (Date.now() > cached.expiry) {
      this.cache.delete(key)
      return null
    }

    return cached.data
  }

  clear() {
    this.cache.clear()
  }

  async getOrFetch(
    key: string,
    fetcher: () => Promise<T>
  ): Promise<T> {
    const cached = this.get(key)

    if (cached !== null) {
      return cached
    }

    const data = await fetcher()
    this.set(key, data)
    return data
  }
}

// Usage
const taskCache = new FirestoreCache<Task>(5)  // 5 minute TTL

async function getCachedTask(taskId: string): Promise<Task> {
  return taskCache.getOrFetch(taskId, async () => {
    const taskDoc = await getDoc(doc(db, 'tasks', taskId))
    return { id: taskDoc.id, ...taskDoc.data() } as Task
  })
}
```

### Technique 6.2: Firestore Persistence Cache

**Leverage built-in offline cache**:

```typescript
import { initializeFirestore, persistentLocalCache, persistentMultipleTabManager } from 'firebase/firestore'

// Configure persistent cache
const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentMultipleTabManager(),
    cacheSizeBytes: 50 * 1024 * 1024  // 50 MB
  })
})

// Queries automatically use cache when available
async function getCachedTasks(projectId: string) {
  const tasksRef = collection(db, 'tasks')
  const q = query(tasksRef, where('projectId', '==', projectId))

  const snapshot = await getDocs(q)

  console.log('From cache:', snapshot.metadata.fromCache)

  return snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}
```

---

## 7. Cost Optimization

### Technique 7.1: Read Cost Reduction

**Minimize billable reads**:

```typescript
// ❌ BAD: Reads entire collection (expensive)
async function badCountActiveTasks() {
  const snapshot = await getDocs(
    query(collection(db, 'tasks'), where('status', '==', 'active'))
  )
  return snapshot.size
}

// ✅ GOOD: Use aggregation query (cheaper)
async function goodCountActiveTasks() {
  const snapshot = await getCountFromServer(
    query(collection(db, 'tasks'), where('status', '==', 'active'))
  )
  return snapshot.data().count
}

// ✅ BETTER: Maintain counter (no reads)
async function betterCountActiveTasks() {
  const counterDoc = await getDoc(doc(db, 'counters', 'active_tasks'))
  return counterDoc.data()?.count || 0
}

// Update counter on task status change (Cloud Function)
export const updateTaskCounter = onDocumentUpdated('tasks/{taskId}', async (event) => {
  const before = event.data.before.data()
  const after = event.data.after.data()

  if (before.status === 'active' && after.status !== 'active') {
    await updateDoc(doc(db, 'counters', 'active_tasks'), {
      count: increment(-1)
    })
  } else if (before.status !== 'active' && after.status === 'active') {
    await updateDoc(doc(db, 'counters', 'active_tasks'), {
      count: increment(1)
    })
  }
})
```

### Technique 7.2: Write Cost Reduction

**Optimize write operations**:

```typescript
// ❌ BAD: Unnecessary writes
async function badUpdateTaskStatus(taskId: string, status: string) {
  await updateDoc(doc(db, 'tasks', taskId), {
    status: status,
    updatedAt: new Date(),
    lastModifiedBy: 'user-123'
  })
}

// ✅ GOOD: Conditional update (only if changed)
async function goodUpdateTaskStatus(taskId: string, status: string) {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))

  if (taskDoc.data()?.status === status) {
    console.log('Status unchanged, skipping write')
    return
  }

  await updateDoc(doc(db, 'tasks', taskId), {
    status: status,
    updatedAt: serverTimestamp()
  })
}

// ✅ BETTER: Use transactions for conditional updates
async function betterUpdateTaskStatus(taskId: string, status: string) {
  await runTransaction(db, async (transaction) => {
    const taskRef = doc(db, 'tasks', taskId)
    const taskDoc = await transaction.get(taskRef)

    if (taskDoc.data()?.status === status) {
      console.log('Status unchanged, aborting transaction')
      return
    }

    transaction.update(taskRef, {
      status: status,
      updatedAt: serverTimestamp()
    })
  })
}
```

---

## 8. Latency Reduction

### Technique 8.1: Regional Optimization

**Deploy to user-closest region**:

```typescript
// Configure multi-region deployment
// firebase.json
{
  "functions": [
    {
      "source": "functions",
      "codebase": "default",
      "runtime": "nodejs18",
      "region": [
        "us-central1",
        "europe-west1",
        "asia-northeast1"
      ]
    }
  ]
}
```

### Technique 8.2: Parallel Operations

**Execute concurrent reads**:

```typescript
// ❌ BAD: Sequential reads (slow)
async function badGetTaskDetails(taskId: string) {
  const task = await getDoc(doc(db, 'tasks', taskId))
  const user = await getDoc(doc(db, 'users', task.data().assignedToId))
  const project = await getDoc(doc(db, 'projects', task.data().projectId))

  return { task: task.data(), user: user.data(), project: project.data() }
}

// ✅ GOOD: Parallel reads (fast)
async function goodGetTaskDetails(taskId: string) {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))
  const taskData = taskDoc.data()

  const [userDoc, projectDoc] = await Promise.all([
    getDoc(doc(db, 'users', taskData.assignedToId)),
    getDoc(doc(db, 'projects', taskData.projectId))
  ])

  return {
    task: taskData,
    user: userDoc.data(),
    project: projectDoc.data()
  }
}
```

---

## Performance Benchmarks

**Typical improvements**:

| Optimization | Before | After | Improvement |
|--------------|--------|-------|-------------|
| Batch writes | 5s (500 ops) | 100ms | 98% faster |
| Denormalization | 150ms (3 reads) | 20ms (1 read) | 87% faster |
| Client cache | 50ms | 0ms (cache hit) | 100% faster |
| Aggregation query | 200ms | 10ms | 95% faster |
| Parallel reads | 300ms (3 sequential) | 100ms (parallel) | 67% faster |

---

## Best Practices

**DO**:
- ✅ Use composite indexes for complex queries
- ✅ Batch write operations (up to 500)
- ✅ Denormalize data for read-heavy workloads
- ✅ Implement client-side caching with TTL
- ✅ Use getCountFromServer for counting
- ✅ Deploy functions to multiple regions
- ✅ Execute reads in parallel when possible

**DON'T**:
- ❌ Perform N+1 queries
- ❌ Read full documents when count is enough
- ❌ Write unchanged data
- ❌ Ignore cache opportunities
- ❌ Deploy to single region for global users
- ❌ Use shallow queries when deep queries work

---

**Context7 Reference**: `/llmstxt/firebase_google-llms.txt` (latest API v9+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
