# Firebase Platform Examples & Patterns

**Complete production-ready examples for Firestore, Authentication, Cloud Functions, and Storage.**

**Research Base**: Official Firebase documentation, Context7 `/llmstxt/firebase_google-llms.txt` patterns
**Last Updated**: 2025-11-22
**Version**: 1.0.0

---

## 📚 Table of Contents

1. [Firestore Initialization & CRUD](#1-firestore-initialization--crud)
2. [Collection Queries & Filtering](#2-collection-queries--filtering)
3. [Composite Indexes](#3-composite-indexes)
4. [Real-time Listeners](#4-real-time-listeners)
5. [Transactions & Batch Writes](#5-transactions--batch-writes)
6. [Cloud Functions Triggers](#6-cloud-functions-triggers)
7. [Firebase Authentication](#7-firebase-authentication)
8. [Security Rules](#8-security-rules)
9. [Offline Persistence](#9-offline-persistence)
10. [Performance Monitoring](#10-performance-monitoring)

---

## 1. Firestore Initialization & CRUD

### Example 1.1: Initialize Firestore

**Complete Firestore setup**:

```javascript
// JavaScript/Node.js
import { initializeApp } from 'firebase/app'
import { getFirestore, collection, doc, setDoc, getDoc, updateDoc, deleteDoc } from 'firebase/firestore'

const firebaseConfig = {
  apiKey: process.env.FIREBASE_API_KEY,
  authDomain: process.env.FIREBASE_AUTH_DOMAIN,
  projectId: process.env.FIREBASE_PROJECT_ID,
  storageBucket: process.env.FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.FIREBASE_APP_ID
}

const app = initializeApp(firebaseConfig)
const db = getFirestore(app)

// Create document
async function createUser(userId, userData) {
  try {
    const userRef = doc(db, 'users', userId)
    await setDoc(userRef, {
      ...userData,
      createdAt: new Date(),
      updatedAt: new Date()
    })
    console.log('User created:', userId)
  } catch (error) {
    console.error('Error creating user:', error)
    throw error
  }
}

// Read document
async function getUser(userId) {
  try {
    const userRef = doc(db, 'users', userId)
    const userSnap = await getDoc(userRef)

    if (userSnap.exists()) {
      return { id: userSnap.id, ...userSnap.data() }
    } else {
      throw new Error('User not found')
    }
  } catch (error) {
    console.error('Error getting user:', error)
    throw error
  }
}

// Update document
async function updateUser(userId, updates) {
  try {
    const userRef = doc(db, 'users', userId)
    await updateDoc(userRef, {
      ...updates,
      updatedAt: new Date()
    })
    console.log('User updated:', userId)
  } catch (error) {
    console.error('Error updating user:', error)
    throw error
  }
}

// Delete document
async function deleteUser(userId) {
  try {
    const userRef = doc(db, 'users', userId)
    await deleteDoc(userRef)
    console.log('User deleted:', userId)
  } catch (error) {
    console.error('Error deleting user:', error)
    throw error
  }
}
```

**Python Admin SDK**:

```python
import firebase_admin
from firebase_admin import credentials, firestore
from datetime import datetime

# Initialize Firebase Admin
cred = credentials.Certificate('path/to/serviceAccountKey.json')
firebase_admin.initialize_app(cred)

db = firestore.client()

# Create document
def create_user(user_id, user_data):
    try:
        user_ref = db.collection('users').document(user_id)
        user_ref.set({
            **user_data,
            'created_at': datetime.now(),
            'updated_at': datetime.now()
        })
        print(f'User created: {user_id}')
    except Exception as e:
        print(f'Error creating user: {e}')
        raise

# Read document
def get_user(user_id):
    try:
        user_ref = db.collection('users').document(user_id)
        user_doc = user_ref.get()

        if user_doc.exists:
            return {'id': user_doc.id, **user_doc.to_dict()}
        else:
            raise ValueError('User not found')
    except Exception as e:
        print(f'Error getting user: {e}')
        raise

# Update document
def update_user(user_id, updates):
    try:
        user_ref = db.collection('users').document(user_id)
        user_ref.update({
            **updates,
            'updated_at': datetime.now()
        })
        print(f'User updated: {user_id}')
    except Exception as e:
        print(f'Error updating user: {e}')
        raise

# Delete document
def delete_user(user_id):
    try:
        user_ref = db.collection('users').document(user_id)
        user_ref.delete()
        print(f'User deleted: {user_id}')
    except Exception as e:
        print(f'Error deleting user: {e}')
        raise
```

---

## 2. Collection Queries & Filtering

### Example 2.1: Advanced Firestore Queries

**Complex query patterns**:

```javascript
import { query, collection, where, orderBy, limit, getDocs, startAfter } from 'firebase/firestore'

// Simple query
async function getActiveUsers() {
  const usersRef = collection(db, 'users')
  const q = query(
    usersRef,
    where('status', '==', 'active'),
    orderBy('createdAt', 'desc'),
    limit(10)
  )

  const querySnapshot = await getDocs(q)
  const users = []

  querySnapshot.forEach((doc) => {
    users.push({ id: doc.id, ...doc.data() })
  })

  return users
}

// Compound query
async function getUsersByRoleAndStatus(role, status) {
  const usersRef = collection(db, 'users')
  const q = query(
    usersRef,
    where('role', '==', role),
    where('status', '==', status),
    orderBy('createdAt', 'desc')
  )

  const querySnapshot = await getDocs(q)
  return querySnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}

// Range query
async function getUsersCreatedBetween(startDate, endDate) {
  const usersRef = collection(db, 'users')
  const q = query(
    usersRef,
    where('createdAt', '>=', startDate),
    where('createdAt', '<=', endDate),
    orderBy('createdAt', 'asc')
  )

  const querySnapshot = await getDocs(q)
  return querySnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}

// Pagination
async function getUsersPaginated(pageSize = 10, lastDoc = null) {
  const usersRef = collection(db, 'users')
  let q = query(
    usersRef,
    orderBy('createdAt', 'desc'),
    limit(pageSize)
  )

  if (lastDoc) {
    q = query(q, startAfter(lastDoc))
  }

  const querySnapshot = await getDocs(q)
  const users = querySnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
  const lastVisible = querySnapshot.docs[querySnapshot.docs.length - 1]

  return {
    users,
    lastDoc: lastVisible,
    hasMore: querySnapshot.docs.length === pageSize
  }
}

// Array contains query
async function getUsersWithSkill(skill) {
  const usersRef = collection(db, 'users')
  const q = query(
    usersRef,
    where('skills', 'array-contains', skill)
  )

  const querySnapshot = await getDocs(q)
  return querySnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}

// In query
async function getUsersByIds(userIds) {
  const usersRef = collection(db, 'users')
  const q = query(
    usersRef,
    where('__name__', 'in', userIds)
  )

  const querySnapshot = await getDocs(q)
  return querySnapshot.docs.map(doc => ({ id: doc.id, ...doc.data() }))
}
```

**Python queries**:

```python
# Simple query
def get_active_users():
    users_ref = db.collection('users')
    query = users_ref.where('status', '==', 'active').order_by('created_at', direction=firestore.Query.DESCENDING).limit(10)

    docs = query.stream()
    return [{'id': doc.id, **doc.to_dict()} for doc in docs]

# Compound query
def get_users_by_role_and_status(role, status):
    users_ref = db.collection('users')
    query = users_ref.where('role', '==', role).where('status', '==', status)

    docs = query.stream()
    return [{'id': doc.id, **doc.to_dict()} for doc in docs]
```

---

## 3. Composite Indexes

### Example 3.1: Index Configuration

**firestore.indexes.json**:

```json
{
  "indexes": [
    {
      "collectionGroup": "users",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "role", "order": "ASCENDING" },
        { "fieldPath": "status", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "tasks",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "projectId", "order": "ASCENDING" },
        { "fieldPath": "status", "order": "ASCENDING" },
        { "fieldPath": "priority", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "tasks",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "assignedTo", "order": "ASCENDING" },
        { "fieldPath": "dueDate", "order": "ASCENDING" }
      ]
    }
  ],
  "fieldOverrides": [
    {
      "collectionGroup": "users",
      "fieldPath": "tags",
      "indexes": [
        { "order": "ASCENDING", "queryScope": "COLLECTION" },
        { "arrayConfig": "CONTAINS", "queryScope": "COLLECTION" }
      ]
    }
  ]
}
```

**Deploy indexes**:

```bash
firebase deploy --only firestore:indexes
```

---

## 4. Real-time Listeners

### Example 4.1: Document and Collection Listeners

**Real-time data sync**:

```javascript
import { doc, onSnapshot, collection, query, where } from 'firebase/firestore'

// Document listener
function listenToUser(userId, callback) {
  const userRef = doc(db, 'users', userId)

  const unsubscribe = onSnapshot(
    userRef,
    (doc) => {
      if (doc.exists()) {
        callback({ id: doc.id, ...doc.data() })
      } else {
        callback(null)
      }
    },
    (error) => {
      console.error('Listener error:', error)
    }
  )

  return unsubscribe
}

// Collection listener
function listenToActiveTasks(projectId, callback) {
  const tasksRef = collection(db, 'tasks')
  const q = query(
    tasksRef,
    where('projectId', '==', projectId),
    where('status', 'in', ['pending', 'in_progress'])
  )

  const unsubscribe = onSnapshot(
    q,
    (snapshot) => {
      const tasks = []
      snapshot.forEach((doc) => {
        tasks.push({ id: doc.id, ...doc.data() })
      })
      callback(tasks)
    },
    (error) => {
      console.error('Listener error:', error)
    }
  )

  return unsubscribe
}

// Listener with changes detection
function listenToTaskChanges(projectId, onAdded, onModified, onRemoved) {
  const tasksRef = collection(db, 'tasks')
  const q = query(tasksRef, where('projectId', '==', projectId))

  const unsubscribe = onSnapshot(q, (snapshot) => {
    snapshot.docChanges().forEach((change) => {
      const task = { id: change.doc.id, ...change.doc.data() }

      if (change.type === 'added') {
        onAdded(task)
      } else if (change.type === 'modified') {
        onModified(task)
      } else if (change.type === 'removed') {
        onRemoved(task)
      }
    })
  })

  return unsubscribe
}
```

**React hook example**:

```javascript
import { useEffect, useState } from 'react'

function useRealtimeTasks(projectId) {
  const [tasks, setTasks] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const tasksRef = collection(db, 'tasks')
    const q = query(tasksRef, where('projectId', '==', projectId))

    const unsubscribe = onSnapshot(
      q,
      (snapshot) => {
        const tasksData = snapshot.docs.map(doc => ({
          id: doc.id,
          ...doc.data()
        }))
        setTasks(tasksData)
        setLoading(false)
      },
      (err) => {
        setError(err)
        setLoading(false)
      }
    )

    return () => unsubscribe()
  }, [projectId])

  return { tasks, loading, error }
}
```

---

## 5. Transactions & Batch Writes

### Example 5.1: Firestore Transactions

**Atomic operations**:

```javascript
import { runTransaction, writeBatch } from 'firebase/firestore'

// Transaction example: Transfer credits
async function transferCredits(fromUserId, toUserId, amount) {
  try {
    await runTransaction(db, async (transaction) => {
      const fromRef = doc(db, 'users', fromUserId)
      const toRef = doc(db, 'users', toUserId)

      const fromDoc = await transaction.get(fromRef)
      const toDoc = await transaction.get(toRef)

      if (!fromDoc.exists() || !toDoc.exists()) {
        throw new Error('User not found')
      }

      const fromCredits = fromDoc.data().credits
      const toCredits = toDoc.data().credits

      if (fromCredits < amount) {
        throw new Error('Insufficient credits')
      }

      transaction.update(fromRef, { credits: fromCredits - amount })
      transaction.update(toRef, { credits: toCredits + amount })
    })

    console.log('Transfer successful')
  } catch (error) {
    console.error('Transfer failed:', error)
    throw error
  }
}

// Batch write example
async function batchCreateTasks(tasks) {
  const batch = writeBatch(db)

  tasks.forEach((taskData) => {
    const taskRef = doc(collection(db, 'tasks'))
    batch.set(taskRef, {
      ...taskData,
      createdAt: new Date(),
      updatedAt: new Date()
    })
  })

  try {
    await batch.commit()
    console.log('Batch write successful')
  } catch (error) {
    console.error('Batch write failed:', error)
    throw error
  }
}
```

**Python transactions**:

```python
from google.cloud import firestore

@firestore.transactional
def transfer_credits(transaction, from_user_id, to_user_id, amount):
    from_ref = db.collection('users').document(from_user_id)
    to_ref = db.collection('users').document(to_user_id)

    from_doc = from_ref.get(transaction=transaction)
    to_doc = to_ref.get(transaction=transaction)

    if not from_doc.exists or not to_doc.exists:
        raise ValueError('User not found')

    from_credits = from_doc.get('credits')
    to_credits = to_doc.get('credits')

    if from_credits < amount:
        raise ValueError('Insufficient credits')

    transaction.update(from_ref, {'credits': from_credits - amount})
    transaction.update(to_ref, {'credits': to_credits + amount})

# Execute transaction
transaction = db.transaction()
transfer_credits(transaction, 'user1', 'user2', 100)
```

---

## 6. Cloud Functions Triggers

### Example 6.1: Firestore Triggers

**Cloud Functions v2**:

```javascript
// functions/src/index.js
import { onDocumentCreated, onDocumentUpdated, onDocumentDeleted } from 'firebase-functions/v2/firestore'
import { initializeApp } from 'firebase-admin/app'
import { getFirestore } from 'firebase-admin/firestore'

initializeApp()
const db = getFirestore()

// Trigger on document creation
export const onUserCreated = onDocumentCreated('users/{userId}', async (event) => {
  const userId = event.params.userId
  const userData = event.data.data()

  console.log('New user created:', userId, userData)

  // Create welcome notification
  await db.collection('notifications').add({
    userId: userId,
    message: `Welcome ${userData.name}!`,
    createdAt: new Date()
  })
})

// Trigger on document update
export const onUserUpdated = onDocumentUpdated('users/{userId}', async (event) => {
  const userId = event.params.userId
  const beforeData = event.data.before.data()
  const afterData = event.data.after.data()

  // Detect role change
  if (beforeData.role !== afterData.role) {
    console.log(`User ${userId} role changed: ${beforeData.role} -> ${afterData.role}`)

    // Log role change
    await db.collection('audit_logs').add({
      userId: userId,
      action: 'role_change',
      before: beforeData.role,
      after: afterData.role,
      timestamp: new Date()
    })
  }
})

// Trigger on document deletion
export const onUserDeleted = onDocumentDeleted('users/{userId}', async (event) => {
  const userId = event.params.userId
  const userData = event.data.data()

  console.log('User deleted:', userId)

  // Clean up user data
  const batch = db.batch()

  // Delete user tasks
  const tasksSnapshot = await db.collection('tasks').where('assignedTo', '==', userId).get()
  tasksSnapshot.docs.forEach((doc) => {
    batch.delete(doc.ref)
  })

  // Delete user notifications
  const notificationsSnapshot = await db.collection('notifications').where('userId', '==', userId).get()
  notificationsSnapshot.docs.forEach((doc) => {
    batch.delete(doc.ref)
  })

  await batch.commit()
})
```

---

## 7. Firebase Authentication

### Example 7.1: Auth Integration

**Authentication with Firestore**:

```javascript
import { getAuth, createUserWithEmailAndPassword, signInWithEmailAndPassword, onAuthStateChanged } from 'firebase/auth'
import { doc, setDoc, getDoc } from 'firebase/firestore'

const auth = getAuth()

// Sign up
async function signUp(email, password, userData) {
  try {
    const userCredential = await createUserWithEmailAndPassword(auth, email, password)
    const user = userCredential.user

    // Create user profile in Firestore
    await setDoc(doc(db, 'users', user.uid), {
      email: email,
      ...userData,
      createdAt: new Date(),
      updatedAt: new Date()
    })

    return user
  } catch (error) {
    console.error('Sign up error:', error)
    throw error
  }
}

// Sign in
async function signIn(email, password) {
  try {
    const userCredential = await signInWithEmailAndPassword(auth, email, password)
    return userCredential.user
  } catch (error) {
    console.error('Sign in error:', error)
    throw error
  }
}

// Auth state listener
function listenToAuthState(callback) {
  return onAuthStateChanged(auth, async (user) => {
    if (user) {
      // Fetch user profile
      const userDoc = await getDoc(doc(db, 'users', user.uid))
      callback({
        uid: user.uid,
        email: user.email,
        ...userDoc.data()
      })
    } else {
      callback(null)
    }
  })
}
```

---

## 8. Security Rules

### Example 8.1: Firestore Security Rules

**firestore.rules**:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Helper functions
    function isAuthenticated() {
      return request.auth != null;
    }

    function isOwner(userId) {
      return isAuthenticated() && request.auth.uid == userId;
    }

    function hasRole(role) {
      return isAuthenticated() &&
             get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == role;
    }

    // Users collection
    match /users/{userId} {
      allow read: if isAuthenticated();
      allow create: if isOwner(userId);
      allow update: if isOwner(userId) || hasRole('admin');
      allow delete: if hasRole('admin');
    }

    // Tasks collection
    match /tasks/{taskId} {
      allow read: if isAuthenticated() &&
                     (resource.data.assignedTo == request.auth.uid ||
                      hasRole('admin'));
      allow create: if isAuthenticated();
      allow update: if isAuthenticated() &&
                       (resource.data.assignedTo == request.auth.uid ||
                        hasRole('admin'));
      allow delete: if hasRole('admin');
    }

    // Projects collection with team access
    match /projects/{projectId} {
      function isMember() {
        return isAuthenticated() &&
               request.auth.uid in resource.data.members;
      }

      allow read: if isMember() || hasRole('admin');
      allow create: if isAuthenticated();
      allow update: if isMember() || hasRole('admin');
      allow delete: if hasRole('admin');
    }
  }
}
```

---

## 9. Offline Persistence

### Example 9.1: Enable Offline Support

**Offline data sync**:

```javascript
import { initializeFirestore, enableIndexedDbPersistence } from 'firebase/firestore'

// Initialize with offline persistence
const db = initializeFirestore(app, {
  cacheSizeBytes: 50 * 1024 * 1024  // 50 MB cache
})

// Enable offline persistence
enableIndexedDbPersistence(db)
  .catch((err) => {
    if (err.code === 'failed-precondition') {
      console.warn('Multiple tabs open, persistence can only be enabled in one tab')
    } else if (err.code === 'unimplemented') {
      console.warn('Browser does not support persistence')
    }
  })

// Offline-aware query
async function getTasksOfflineAware(projectId) {
  try {
    const tasksRef = collection(db, 'tasks')
    const q = query(tasksRef, where('projectId', '==', projectId))

    const querySnapshot = await getDocs(q)

    const tasks = querySnapshot.docs.map(doc => ({
      id: doc.id,
      ...doc.data(),
      fromCache: doc.metadata.fromCache  // Check if from cache
    }))

    return tasks
  } catch (error) {
    console.error('Error fetching tasks:', error)
    throw error
  }
}
```

---

## 10. Performance Monitoring

### Example 10.1: Firebase Performance

**Track performance**:

```javascript
import { getPerformance, trace } from 'firebase/performance'

const perf = getPerformance(app)

// Custom trace
async function fetchTasksWithMonitoring(projectId) {
  const customTrace = trace(perf, 'fetch_tasks')
  customTrace.start()

  try {
    const tasks = await getTasks(projectId)
    customTrace.putAttribute('project_id', projectId)
    customTrace.putMetric('task_count', tasks.length)
    customTrace.stop()
    return tasks
  } catch (error) {
    customTrace.stop()
    throw error
  }
}
```

---

## Best Practices

**DO**:
- ✅ Use composite indexes for complex queries
- ✅ Implement offline persistence
- ✅ Use security rules for all collections
- ✅ Batch writes for bulk operations
- ✅ Monitor performance metrics
- ✅ Clean up listeners on unmount

**DON'T**:
- ❌ Query without indexes
- ❌ Skip security rules
- ❌ Over-fetch data
- ❌ Forget to unsubscribe listeners
- ❌ Store large binary data in Firestore

---

**Context7 Reference**: `/llmstxt/firebase_google-llms.txt` (latest API v9+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
