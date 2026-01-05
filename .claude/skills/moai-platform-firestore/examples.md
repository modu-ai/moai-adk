# Firestore Code Examples

## Complete Working Examples

### Example 1: Task Management App with Offline Support

```typescript
// lib/firebase.ts
import { initializeApp, getApps } from 'firebase/app'
import {
  initializeFirestore,
  persistentLocalCache,
  persistentMultipleTabManager,
  CACHE_SIZE_UNLIMITED
} from 'firebase/firestore'

const firebaseConfig = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
  storageBucket: process.env.NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID
}

const app = getApps().length === 0 ? initializeApp(firebaseConfig) : getApps()[0]

export const db = initializeFirestore(app, {
  localCache: persistentLocalCache({
    tabManager: persistentMultipleTabManager(),
    cacheSizeBytes: CACHE_SIZE_UNLIMITED
  })
})

export { app }
```

```typescript
// hooks/useTasks.ts
import { useEffect, useState, useCallback } from 'react'
import {
  collection,
  query,
  where,
  orderBy,
  onSnapshot,
  addDoc,
  updateDoc,
  deleteDoc,
  doc,
  serverTimestamp,
  Timestamp
} from 'firebase/firestore'
import { db } from '@/lib/firebase'

interface Task {
  id: string
  title: string
  description: string
  completed: boolean
  priority: 'low' | 'medium' | 'high'
  dueDate: Timestamp | null
  createdAt: Timestamp
  updatedAt: Timestamp
  userId: string
  _pending: boolean
  _fromCache: boolean
}

export function useTasks(userId: string | null) {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    if (!userId) {
      setTasks([])
      setLoading(false)
      return
    }

    const q = query(
      collection(db, 'tasks'),
      where('userId', '==', userId),
      orderBy('createdAt', 'desc')
    )

    const unsubscribe = onSnapshot(
      q,
      { includeMetadataChanges: true },
      (snapshot) => {
        const taskList = snapshot.docs.map((doc) => ({
          id: doc.id,
          ...doc.data(),
          _pending: doc.metadata.hasPendingWrites,
          _fromCache: doc.metadata.fromCache
        })) as Task[]
        setTasks(taskList)
        setLoading(false)
        setError(null)
      },
      (err) => {
        console.error('Tasks subscription error:', err)
        setError(err)
        setLoading(false)
      }
    )

    return () => unsubscribe()
  }, [userId])

  const addTask = useCallback(
    async (task: Omit<Task, 'id' | 'createdAt' | 'updatedAt' | '_pending' | '_fromCache'>) => {
      const docRef = await addDoc(collection(db, 'tasks'), {
        ...task,
        createdAt: serverTimestamp(),
        updatedAt: serverTimestamp()
      })
      return docRef.id
    },
    []
  )

  const updateTask = useCallback(async (taskId: string, updates: Partial<Task>) => {
    await updateDoc(doc(db, 'tasks', taskId), {
      ...updates,
      updatedAt: serverTimestamp()
    })
  }, [])

  const deleteTask = useCallback(async (taskId: string) => {
    await deleteDoc(doc(db, 'tasks', taskId))
  }, [])

  const toggleComplete = useCallback(
    async (taskId: string) => {
      const task = tasks.find((t) => t.id === taskId)
      if (task) {
        await updateTask(taskId, { completed: !task.completed })
      }
    },
    [tasks, updateTask]
  )

  return { tasks, loading, error, addTask, updateTask, deleteTask, toggleComplete }
}
```

---

### Example 2: Organization with Role-Based Access

```typescript
// services/organizationService.ts
import {
  collection,
  doc,
  getDocs,
  query,
  where,
  serverTimestamp,
  writeBatch,
  runTransaction
} from 'firebase/firestore'
import { db } from '@/lib/firebase'

export async function createOrganization(
  userId: string,
  userEmail: string,
  name: string
): Promise<string> {
  const slug = name.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')

  const existingOrg = await getDocs(
    query(collection(db, 'organizations'), where('slug', '==', slug))
  )

  if (!existingOrg.empty) {
    throw new Error('Organization name already taken')
  }

  const batch = writeBatch(db)

  const orgRef = doc(collection(db, 'organizations'))
  batch.set(orgRef, {
    name,
    slug,
    ownerId: userId,
    createdAt: serverTimestamp(),
    settings: {
      allowPublicProjects: false,
      defaultMemberRole: 'viewer'
    }
  })

  const memberRef = doc(db, 'organizations', orgRef.id, 'members', userId)
  batch.set(memberRef, {
    userId,
    email: userEmail,
    displayName: userEmail.split('@')[0],
    role: 'owner',
    joinedAt: serverTimestamp()
  })

  await batch.commit()
  return orgRef.id
}

export async function transferOwnership(
  organizationId: string,
  currentOwnerId: string,
  newOwnerId: string
): Promise<void> {
  await runTransaction(db, async (transaction) => {
    const orgRef = doc(db, 'organizations', organizationId)
    const currentOwnerRef = doc(db, 'organizations', organizationId, 'members', currentOwnerId)
    const newOwnerRef = doc(db, 'organizations', organizationId, 'members', newOwnerId)

    const [orgDoc, currentOwnerDoc, newOwnerDoc] = await Promise.all([
      transaction.get(orgRef),
      transaction.get(currentOwnerRef),
      transaction.get(newOwnerRef)
    ])

    if (!orgDoc.exists()) throw new Error('Organization not found')
    if (!newOwnerDoc.exists()) throw new Error('New owner must be a member')
    if (orgDoc.data().ownerId !== currentOwnerId) throw new Error('Not authorized')

    transaction.update(orgRef, { ownerId: newOwnerId })
    transaction.update(currentOwnerRef, { role: 'admin' })
    transaction.update(newOwnerRef, { role: 'owner' })
  })
}
```

---

### Example 3: Real-time Presence System

```typescript
// hooks/usePresence.ts
import { useEffect, useCallback, useRef, useState } from 'react'
import {
  doc,
  setDoc,
  deleteDoc,
  onSnapshot,
  collection,
  serverTimestamp,
  Timestamp
} from 'firebase/firestore'
import { db } from '@/lib/firebase'

interface PresenceData {
  documentId: string
  userId: string
  displayName: string
  cursorPosition: { line: number; column: number } | null
  lastSeen: Timestamp
  status: 'active' | 'idle' | 'away'
}

export function usePresence(documentId: string, userId: string, displayName: string) {
  const presenceRef = useRef<ReturnType<typeof doc> | null>(null)
  const heartbeatInterval = useRef<NodeJS.Timeout | null>(null)

  useEffect(() => {
    if (!documentId || !userId) return

    presenceRef.current = doc(db, 'documents', documentId, 'presence', userId)

    const updatePresence = async (status: 'active' | 'idle' | 'away' = 'active') => {
      if (presenceRef.current) {
        await setDoc(presenceRef.current, {
          documentId,
          userId,
          displayName,
          cursorPosition: null,
          lastSeen: serverTimestamp(),
          status
        })
      }
    }

    updatePresence('active')

    heartbeatInterval.current = setInterval(() => {
      updatePresence('active')
    }, 30000)

    const handleVisibilityChange = () => {
      if (document.hidden) {
        updatePresence('away')
      } else {
        updatePresence('active')
      }
    }
    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      if (heartbeatInterval.current) {
        clearInterval(heartbeatInterval.current)
      }
      if (presenceRef.current) {
        deleteDoc(presenceRef.current)
      }
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [documentId, userId, displayName])

  const updateCursorPosition = useCallback(
    async (position: { line: number; column: number }) => {
      if (presenceRef.current) {
        await setDoc(
          presenceRef.current,
          {
            cursorPosition: position,
            lastSeen: serverTimestamp(),
            status: 'active'
          },
          { merge: true }
        )
      }
    },
    []
  )

  return { updateCursorPosition }
}

export function useDocumentPresence(documentId: string, currentUserId: string) {
  const [collaborators, setCollaborators] = useState<PresenceData[]>([])

  useEffect(() => {
    if (!documentId) return

    const presenceCollection = collection(db, 'documents', documentId, 'presence')

    const unsubscribe = onSnapshot(presenceCollection, (snapshot) => {
      const presenceList = snapshot.docs
        .filter((doc) => doc.id !== currentUserId)
        .map((doc) => ({
          id: doc.id,
          ...doc.data()
        })) as PresenceData[]

      const now = Date.now()
      const activeCollaborators = presenceList.filter((p) => {
        const lastSeen = p.lastSeen?.toMillis() || 0
        return now - lastSeen < 120000
      })

      setCollaborators(activeCollaborators)
    })

    return () => unsubscribe()
  }, [documentId, currentUserId])

  return collaborators
}
```

---

### Example 4: Cloud Functions V2 Triggers

```typescript
// functions/src/organization.ts
import { onDocumentCreated, onDocumentDeleted } from 'firebase-functions/v2/firestore'
import { onCall, HttpsError } from 'firebase-functions/v2/https'
import { getFirestore, FieldValue } from 'firebase-admin/firestore'

const db = getFirestore()

export const onMemberAdded = onDocumentCreated(
  { document: 'organizations/{orgId}/members/{memberId}', region: 'us-central1' },
  async (event) => {
    const orgId = event.params.orgId
    await db.doc(`organizations/${orgId}`).update({
      memberCount: FieldValue.increment(1),
      updatedAt: FieldValue.serverTimestamp()
    })
  }
)

export const onMemberRemoved = onDocumentDeleted(
  { document: 'organizations/{orgId}/members/{memberId}', region: 'us-central1' },
  async (event) => {
    const orgId = event.params.orgId
    await db.doc(`organizations/${orgId}`).update({
      memberCount: FieldValue.increment(-1),
      updatedAt: FieldValue.serverTimestamp()
    })
  }
)

export const inviteToOrganization = onCall(
  { region: 'us-central1' },
  async (request) => {
    if (!request.auth) {
      throw new HttpsError('unauthenticated', 'Must be signed in')
    }

    const { organizationId, email, role } = request.data

    if (!organizationId || !email || !role) {
      throw new HttpsError('invalid-argument', 'Missing required fields')
    }

    if (!['admin', 'editor', 'viewer'].includes(role)) {
      throw new HttpsError('invalid-argument', 'Invalid role')
    }

    const callerMember = await db
      .doc(`organizations/${organizationId}/members/${request.auth.uid}`)
      .get()

    if (!callerMember.exists) {
      throw new HttpsError('permission-denied', 'Not a member of this organization')
    }

    const callerRole = callerMember.data()?.role
    if (!['owner', 'admin'].includes(callerRole)) {
      throw new HttpsError('permission-denied', 'Must be admin or owner to invite')
    }

    const invitation = await db.collection('invitations').add({
      organizationId,
      email,
      role,
      invitedBy: request.auth.uid,
      createdAt: FieldValue.serverTimestamp(),
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      status: 'pending'
    })

    return { invitationId: invitation.id, success: true }
  }
)
```

---

Status: Code Examples
Last Updated: 2026-01-06
