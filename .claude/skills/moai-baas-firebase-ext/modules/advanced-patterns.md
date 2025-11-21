# Firebase Advanced Patterns

**Enterprise-grade patterns for complex Firebase architectures.**

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Research Base**: Context7 `/llmstxt/firebase_google-llms.txt`, production patterns

---

## 📚 Table of Contents

1. [Multi-Tenant Security Rules](#1-multi-tenant-security-rules)
2. [Composite Index Strategy](#2-composite-index-strategy)
3. [Eventual Consistency Patterns](#3-eventual-consistency-patterns)
4. [Offline-First Architecture](#4-offline-first-architecture)
5. [Real-time Multiplayer](#5-real-time-multiplayer)
6. [Trigger-Based Workflows](#6-trigger-based-workflows)
7. [Distributed Tracing](#7-distributed-tracing)
8. [Data Modeling](#8-data-modeling)

---

## 1. Multi-Tenant Security Rules

### Pattern 1.1: Hierarchical Tenant Isolation

**Advanced security rules for SaaS**:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Helper functions
    function isAuthenticated() {
      return request.auth != null;
    }

    function getUserData() {
      return get(/databases/$(database)/documents/users/$(request.auth.uid)).data;
    }

    function isTenantMember(tenantId) {
      return isAuthenticated() &&
             tenantId in getUserData().tenants;
    }

    function getTenantRole(tenantId) {
      return getUserData().tenantRoles[tenantId];
    }

    function hasMinRole(tenantId, minRole) {
      let role = getTenantRole(tenantId);
      let roles = ['viewer', 'member', 'admin', 'owner'];
      let roleIndex = roles.indexOf(role);
      let minRoleIndex = roles.indexOf(minRole);
      return roleIndex >= minRoleIndex;
    }

    // Tenants collection
    match /tenants/{tenantId} {
      allow read: if isTenantMember(tenantId);
      allow create: if isAuthenticated();
      allow update: if hasMinRole(tenantId, 'admin');
      allow delete: if hasMinRole(tenantId, 'owner');

      // Tenant-scoped subcollections
      match /projects/{projectId} {
        allow read: if isTenantMember(tenantId);
        allow create: if hasMinRole(tenantId, 'member');
        allow update: if hasMinRole(tenantId, 'member');
        allow delete: if hasMinRole(tenantId, 'admin');

        // Project tasks
        match /tasks/{taskId} {
          allow read: if isTenantMember(tenantId);
          allow write: if hasMinRole(tenantId, 'member') &&
                          (request.resource.data.assignedTo == request.auth.uid ||
                           hasMinRole(tenantId, 'admin'));
        }
      }

      // Tenant members
      match /members/{userId} {
        allow read: if isTenantMember(tenantId);
        allow create: if hasMinRole(tenantId, 'admin');
        allow update: if hasMinRole(tenantId, 'admin') ||
                         (isAuthenticated() && request.auth.uid == userId);
        allow delete: if hasMinRole(tenantId, 'owner');
      }
    }

    // User profiles
    match /users/{userId} {
      allow read: if isAuthenticated();
      allow create: if isAuthenticated() && request.auth.uid == userId;
      allow update: if isAuthenticated() && request.auth.uid == userId;
      allow delete: if false;  // Users cannot delete themselves
    }
  }
}
```

**TypeScript client for tenant context**:

```typescript
import { collection, query, where, getDocs, doc, setDoc } from 'firebase/firestore'

class TenantContext {
  private currentTenantId: string | null = null

  async setTenant(tenantId: string, userId: string) {
    // Verify user has access to tenant
    const userDoc = await getDocs(doc(db, 'users', userId))
    const userData = userDoc.data()

    if (!userData?.tenants.includes(tenantId)) {
      throw new Error('User does not have access to this tenant')
    }

    this.currentTenantId = tenantId
  }

  getCurrentTenantId(): string {
    if (!this.currentTenantId) {
      throw new Error('No tenant selected')
    }
    return this.currentTenantId
  }

  async getTenantProjects() {
    const tenantId = this.getCurrentTenantId()
    const projectsRef = collection(db, `tenants/${tenantId}/projects`)
    const snapshot = await getDocs(projectsRef)

    return snapshot.docs.map(doc => ({
      id: doc.id,
      ...doc.data()
    }))
  }

  async createTenantProject(projectData: any) {
    const tenantId = this.getCurrentTenantId()
    const projectsRef = collection(db, `tenants/${tenantId}/projects`)

    await setDoc(doc(projectsRef), {
      ...projectData,
      tenantId,
      createdAt: new Date(),
      updatedAt: new Date()
    })
  }
}

export const tenantContext = new TenantContext()
```

---

## 2. Composite Index Strategy

### Pattern 2.1: Optimal Index Design

**Strategic index planning**:

```json
{
  "indexes": [
    {
      "collectionGroup": "tasks",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tenantId", "order": "ASCENDING" },
        { "fieldPath": "projectId", "order": "ASCENDING" },
        { "fieldPath": "status", "order": "ASCENDING" },
        { "fieldPath": "priority", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "tasks",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tenantId", "order": "ASCENDING" },
        { "fieldPath": "assignedTo", "order": "ASCENDING" },
        { "fieldPath": "dueDate", "order": "ASCENDING" }
      ]
    },
    {
      "collectionGroup": "tasks",
      "queryScope": "COLLECTION_GROUP",
      "fields": [
        { "fieldPath": "assignedTo", "order": "ASCENDING" },
        { "fieldPath": "status", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "analytics",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tenantId", "order": "ASCENDING" },
        { "fieldPath": "eventType", "order": "ASCENDING" },
        { "fieldPath": "timestamp", "order": "DESCENDING" }
      ]
    }
  ],
  "fieldOverrides": [
    {
      "collectionGroup": "tasks",
      "fieldPath": "tags",
      "indexes": [
        { "arrayConfig": "CONTAINS", "queryScope": "COLLECTION" }
      ]
    },
    {
      "collectionGroup": "users",
      "fieldPath": "skills",
      "indexes": [
        { "arrayConfig": "CONTAINS", "queryScope": "COLLECTION" }
      ]
    }
  ]
}
```

**Query optimization analysis**:

```typescript
// ❌ BAD: No index, will fail or be slow
async function badQuery() {
  const tasksRef = collection(db, 'tasks')
  const q = query(
    tasksRef,
    where('status', '==', 'active'),
    where('priority', '>', 5),
    orderBy('createdAt', 'desc')
  )

  return await getDocs(q)
}

// ✅ GOOD: Uses composite index
async function goodQuery(tenantId: string, projectId: string) {
  const tasksRef = collection(db, 'tasks')
  const q = query(
    tasksRef,
    where('tenantId', '==', tenantId),
    where('projectId', '==', projectId),
    where('status', '==', 'active'),
    orderBy('priority', 'desc')
  )

  return await getDocs(q)
}

// ✅ BETTER: Collection group query with index
async function collectionGroupQuery(userId: string) {
  const tasksRef = collectionGroup(db, 'tasks')
  const q = query(
    tasksRef,
    where('assignedTo', '==', userId),
    where('status', 'in', ['pending', 'in_progress']),
    orderBy('createdAt', 'desc'),
    limit(50)
  )

  return await getDocs(q)
}
```

---

## 3. Eventual Consistency Patterns

### Pattern 3.1: Aggregate Counters

**Distributed counter pattern**:

```typescript
import { runTransaction, doc, increment } from 'firebase/firestore'

class DistributedCounter {
  private numShards: number

  constructor(numShards = 10) {
    this.numShards = numShards
  }

  private getShardId(): string {
    return `shard_${Math.floor(Math.random() * this.numShards)}`
  }

  async incrementCounter(counterId: string, amount = 1) {
    const shardId = this.getShardId()
    const shardRef = doc(db, `counters/${counterId}/shards/${shardId}`)

    await runTransaction(db, async (transaction) => {
      transaction.set(shardRef, {
        count: increment(amount)
      }, { merge: true })
    })
  }

  async getCount(counterId: string): Promise<number> {
    const shardsRef = collection(db, `counters/${counterId}/shards`)
    const snapshot = await getDocs(shardsRef)

    let totalCount = 0
    snapshot.forEach((doc) => {
      totalCount += doc.data().count || 0
    })

    return totalCount
  }
}

const counter = new DistributedCounter()

// Usage
await counter.incrementCounter('page_views')
const views = await counter.getCount('page_views')
```

### Pattern 3.2: Denormalization

**Data denormalization for read optimization**:

```typescript
// ❌ BAD: Requires multiple reads
async function getTaskWithUserInfo(taskId: string) {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))
  const task = taskDoc.data()

  // Extra read for user info
  const userDoc = await getDoc(doc(db, 'users', task.assignedTo))
  const user = userDoc.data()

  return {
    ...task,
    assignee: {
      name: user.name,
      email: user.email
    }
  }
}

// ✅ GOOD: Denormalized data (single read)
interface Task {
  id: string
  title: string
  assignedTo: string
  assignee: {
    name: string
    email: string
    avatarUrl: string
  }
  status: string
}

async function getTaskOptimized(taskId: string): Promise<Task> {
  const taskDoc = await getDoc(doc(db, 'tasks', taskId))
  return { id: taskDoc.id, ...taskDoc.data() } as Task
}

// Update denormalized data when user changes
async function updateUserProfile(userId: string, updates: any) {
  const batch = writeBatch(db)

  // Update user profile
  batch.update(doc(db, 'users', userId), updates)

  // Update denormalized data in tasks
  const tasksSnapshot = await getDocs(
    query(collection(db, 'tasks'), where('assignedTo', '==', userId))
  )

  tasksSnapshot.forEach((taskDoc) => {
    batch.update(taskDoc.ref, {
      'assignee.name': updates.name,
      'assignee.email': updates.email,
      'assignee.avatarUrl': updates.avatarUrl
    })
  })

  await batch.commit()
}
```

---

## 4. Offline-First Architecture

### Pattern 4.1: Optimistic UI Updates

**Instant feedback with eventual sync**:

```typescript
import { enableIndexedDbPersistence, writeBatch } from 'firebase/firestore'

class OfflineFirstTaskManager {
  constructor() {
    // Enable offline persistence
    enableIndexedDbPersistence(db).catch(console.error)
  }

  async createTaskOptimistic(taskData: any) {
    const taskRef = doc(collection(db, 'tasks'))
    const optimisticTask = {
      id: taskRef.id,
      ...taskData,
      status: 'pending',
      createdAt: new Date(),
      _pending: true  // Mark as pending sync
    }

    // 1. Update UI immediately
    this.onTaskCreated(optimisticTask)

    try {
      // 2. Write to Firestore (will queue if offline)
      await setDoc(taskRef, {
        ...taskData,
        createdAt: serverTimestamp(),
        status: 'pending'
      })

      // 3. Update UI with server data
      this.onTaskSynced(taskRef.id)
    } catch (error) {
      // 4. Handle sync error
      this.onTaskSyncFailed(taskRef.id, error)
      throw error
    }
  }

  async updateTaskOptimistic(taskId: string, updates: any) {
    // 1. Update UI immediately
    this.onTaskUpdated(taskId, { ...updates, _pending: true })

    try {
      // 2. Write to Firestore
      await updateDoc(doc(db, 'tasks', taskId), {
        ...updates,
        updatedAt: serverTimestamp()
      })

      // 3. Clear pending state
      this.onTaskSynced(taskId)
    } catch (error) {
      this.onTaskSyncFailed(taskId, error)
      throw error
    }
  }

  // Override these methods
  protected onTaskCreated(task: any) {}
  protected onTaskUpdated(taskId: string, updates: any) {}
  protected onTaskSynced(taskId: string) {}
  protected onTaskSyncFailed(taskId: string, error: any) {}
}
```

---

## 5. Real-time Multiplayer

### Pattern 5.1: Presence System

**Track online users**:

```typescript
import { ref, onValue, onDisconnect, serverTimestamp, set } from 'firebase/database'
import { getDatabase } from 'firebase/database'

const rtdb = getDatabase()

class PresenceSystem {
  async trackUserPresence(userId: string, metadata: any) {
    const userStatusRef = ref(rtdb, `status/${userId}`)

    // Online status
    const isOnlineData = {
      state: 'online',
      lastChanged: serverTimestamp(),
      ...metadata
    }

    // Offline status
    const isOfflineData = {
      state: 'offline',
      lastChanged: serverTimestamp(),
      ...metadata
    }

    // Create a reference to the special '.info/connected' path
    const connectedRef = ref(rtdb, '.info/connected')

    onValue(connectedRef, (snapshot) => {
      if (snapshot.val() === false) {
        return
      }

      // Set offline status on disconnect
      onDisconnect(userStatusRef).set(isOfflineData).then(() => {
        // Set online status
        set(userStatusRef, isOnlineData)
      })
    })
  }

  listenToUserPresence(userId: string, callback: (presence: any) => void) {
    const userStatusRef = ref(rtdb, `status/${userId}`)
    return onValue(userStatusRef, (snapshot) => {
      callback(snapshot.val())
    })
  }

  listenToRoomPresence(roomId: string, callback: (users: any[]) => void) {
    const roomRef = ref(rtdb, `rooms/${roomId}/users`)
    return onValue(roomRef, (snapshot) => {
      const users: any[] = []
      snapshot.forEach((childSnapshot) => {
        users.push({
          id: childSnapshot.key,
          ...childSnapshot.val()
        })
      })
      callback(users)
    })
  }
}
```

---

## 6. Trigger-Based Workflows

### Pattern 6.1: Event-Driven Architecture

**Complex workflows with Cloud Functions**:

```typescript
// functions/src/workflows.ts
import { onDocumentCreated, onDocumentUpdated } from 'firebase-functions/v2/firestore'
import { onCall } from 'firebase-functions/v2/https'
import { getFirestore } from 'firebase-admin/firestore'

const db = getFirestore()

// Workflow: Task assignment notification
export const onTaskAssigned = onDocumentUpdated('tasks/{taskId}', async (event) => {
  const before = event.data.before.data()
  const after = event.data.after.data()

  // Check if assignedTo changed
  if (before.assignedTo !== after.assignedTo) {
    const taskId = event.params.taskId

    // Create notification
    await db.collection('notifications').add({
      userId: after.assignedTo,
      type: 'task_assigned',
      taskId: taskId,
      message: `You have been assigned to task: ${after.title}`,
      createdAt: new Date(),
      read: false
    })

    // Send email (via another function)
    await db.collection('mail_queue').add({
      to: after.assignedTo,
      template: 'task_assigned',
      data: {
        taskId: taskId,
        taskTitle: after.title
      }
    })
  }
})

// Workflow: Project completion
export const onProjectCompleted = onDocumentUpdated('projects/{projectId}', async (event) => {
  const before = event.data.before.data()
  const after = event.data.after.data()

  if (before.status !== 'completed' && after.status === 'completed') {
    const projectId = event.params.projectId

    // Update all tasks to completed
    const tasksSnapshot = await db.collection('tasks')
      .where('projectId', '==', projectId)
      .where('status', '!=', 'completed')
      .get()

    const batch = db.batch()
    tasksSnapshot.docs.forEach((doc) => {
      batch.update(doc.ref, {
        status: 'completed',
        completedAt: new Date()
      })
    })

    await batch.commit()

    // Create analytics event
    await db.collection('analytics').add({
      eventType: 'project_completed',
      projectId: projectId,
      timestamp: new Date(),
      metadata: {
        taskCount: tasksSnapshot.size,
        duration: after.completedAt - after.createdAt
      }
    })
  }
})

// Callable function for complex operations
export const assignTasksToUser = onCall(async (request) => {
  const { taskIds, userId } = request.data

  if (!request.auth) {
    throw new Error('Unauthorized')
  }

  const batch = db.batch()

  for (const taskId of taskIds) {
    const taskRef = db.collection('tasks').doc(taskId)
    batch.update(taskRef, {
      assignedTo: userId,
      updatedAt: new Date()
    })
  }

  await batch.commit()

  return { success: true, count: taskIds.length }
})
```

---

## 7. Distributed Tracing

### Pattern 7.1: Cloud Trace Integration

**Performance monitoring**:

```typescript
import { trace } from '@opentelemetry/api'
import { getFirestore } from 'firebase/firestore'

class TracedFirestoreClient {
  private tracer = trace.getTracer('firestore-client')

  async getDocument(collectionPath: string, docId: string) {
    return this.tracer.startActiveSpan('getDocument', async (span) => {
      span.setAttribute('collection', collectionPath)
      span.setAttribute('docId', docId)

      try {
        const docRef = doc(db, collectionPath, docId)
        const snapshot = await getDoc(docRef)

        span.setAttribute('exists', snapshot.exists())
        span.setStatus({ code: 0 })  // Success

        return snapshot
      } catch (error) {
        span.setStatus({ code: 2, message: error.message })  // Error
        throw error
      } finally {
        span.end()
      }
    })
  }

  async queryCollection(collectionPath: string, filters: any[]) {
    return this.tracer.startActiveSpan('queryCollection', async (span) => {
      span.setAttribute('collection', collectionPath)
      span.setAttribute('filterCount', filters.length)

      try {
        const collectionRef = collection(db, collectionPath)
        let q = query(collectionRef)

        filters.forEach((filter) => {
          q = query(q, where(filter.field, filter.operator, filter.value))
        })

        const snapshot = await getDocs(q)

        span.setAttribute('resultCount', snapshot.size)
        span.setStatus({ code: 0 })

        return snapshot
      } catch (error) {
        span.setStatus({ code: 2, message: error.message })
        throw error
      } finally {
        span.end()
      }
    })
  }
}
```

---

## 8. Data Modeling

### Pattern 8.1: Hierarchical Data

**Effective data modeling**:

```typescript
// ❌ BAD: Deep nesting (limited to 20 levels)
interface BadStructure {
  tenants: {
    [tenantId: string]: {
      projects: {
        [projectId: string]: {
          tasks: {
            [taskId: string]: {
              comments: {
                [commentId: string]: any
              }
            }
          }
        }
      }
    }
  }
}

// ✅ GOOD: Flat structure with references
interface Tenant {
  id: string
  name: string
  createdAt: Date
}

interface Project {
  id: string
  tenantId: string  // Reference to tenant
  name: string
  createdAt: Date
}

interface Task {
  id: string
  projectId: string  // Reference to project
  tenantId: string   // Denormalized for queries
  title: string
  status: string
}

interface Comment {
  id: string
  taskId: string  // Reference to task
  userId: string
  content: string
  createdAt: Date
}

// Querying with references
async function getProjectTasks(projectId: string) {
  const tasksSnapshot = await getDocs(
    query(
      collection(db, 'tasks'),
      where('projectId', '==', projectId)
    )
  )

  return tasksSnapshot.docs.map(doc => ({
    id: doc.id,
    ...doc.data()
  }))
}
```

---

## Best Practices

**DO**:
- ✅ Use hierarchical security rules for multi-tenancy
- ✅ Create composite indexes strategically
- ✅ Implement denormalization for read optimization
- ✅ Use offline persistence for mobile apps
- ✅ Track presence with Realtime Database
- ✅ Monitor performance with tracing

**DON'T**:
- ❌ Nest data more than 2-3 levels deep
- ❌ Skip index planning
- ❌ Ignore eventual consistency
- ❌ Over-normalize data
- ❌ Forget offline scenarios

---

**Context7 Reference**: `/llmstxt/firebase_google-llms.txt` (latest API v9+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
