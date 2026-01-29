# Firestore Transactions Module

## Overview

Firestore transactions and batch writes enable atomic operations across multiple documents. Transactions ensure data consistency when operations depend on current values, while batch writes provide efficient bulk updates.

---

## Batch Writes

### Basic Batch Operations

Batch writes allow up to 500 operations in a single atomic commit:

```typescript
import { writeBatch, doc, serverTimestamp } from 'firebase/firestore'
import { db } from '@/lib/firebase'

async function batchCreateDocuments(items: Array<{ id: string; data: any }>) {
  const batch = writeBatch(db)

  items.forEach(({ id, data }) => {
    const docRef = doc(db, 'documents', id)
    batch.set(docRef, {
      ...data,
      createdAt: serverTimestamp()
    })
  })

  await batch.commit()
}
```

### Batch Operations Types

```typescript
import { writeBatch, doc, deleteField } from 'firebase/firestore'

async function demonstrateBatchOperations() {
  const batch = writeBatch(db)

  // Set (create or overwrite)
  batch.set(doc(db, 'users', 'user-1'), {
    name: 'Alice',
    email: 'alice@example.com'
  })

  // Set with merge
  batch.set(
    doc(db, 'users', 'user-2'),
    { lastLogin: serverTimestamp() },
    { merge: true }
  )

  // Update specific fields
  batch.update(doc(db, 'users', 'user-3'), {
    score: 100,
    updatedAt: serverTimestamp()
  })

  // Delete a field
  batch.update(doc(db, 'users', 'user-4'), {
    temporaryData: deleteField()
  })

  // Delete document
  batch.delete(doc(db, 'users', 'user-5'))

  await batch.commit()
}
```

---

## Transactions

### Basic Transaction Pattern

```typescript
import { runTransaction, doc } from 'firebase/firestore'

async function incrementCounter(counterId: string) {
  await runTransaction(db, async (transaction) => {
    const counterRef = doc(db, 'counters', counterId)
    const counterDoc = await transaction.get(counterRef)

    if (!counterDoc.exists()) {
      throw new Error('Counter not found')
    }

    const currentValue = counterDoc.data().value
    transaction.update(counterRef, { value: currentValue + 1 })
  })
}
```

### Transaction with Multiple Documents

```typescript
async function transferCredits(
  fromUserId: string,
  toUserId: string,
  amount: number
) {
  await runTransaction(db, async (transaction) => {
    const fromRef = doc(db, 'users', fromUserId)
    const toRef = doc(db, 'users', toUserId)

    const [fromDoc, toDoc] = await Promise.all([
      transaction.get(fromRef),
      transaction.get(toRef)
    ])

    if (!fromDoc.exists() || !toDoc.exists()) {
      throw new Error('User not found')
    }

    const fromCredits = fromDoc.data().credits
    if (fromCredits < amount) {
      throw new Error('Insufficient credits')
    }

    transaction.update(fromRef, { credits: fromCredits - amount })
    transaction.update(toRef, { credits: toDoc.data().credits + amount })
  })
}
```

### Transaction Return Values

```typescript
async function claimUniqueUsername(userId: string, username: string): Promise<boolean> {
  try {
    const result = await runTransaction(db, async (transaction) => {
      const usernameRef = doc(db, 'usernames', username.toLowerCase())
      const usernameDoc = await transaction.get(usernameRef)

      if (usernameDoc.exists()) {
        return { success: false, reason: 'Username taken' }
      }

      transaction.set(usernameRef, {
        userId,
        createdAt: serverTimestamp()
      })

      transaction.update(doc(db, 'users', userId), {
        username,
        usernameClaimedAt: serverTimestamp()
      })

      return { success: true }
    })

    return result.success
  } catch (error) {
    console.error('Failed to claim username:', error)
    return false
  }
}
```

---

## Transaction Rules

### Read Before Write Rule

All reads must happen before any writes in a transaction:

```typescript
// CORRECT: All reads first, then writes
await runTransaction(db, async (transaction) => {
  const doc1 = await transaction.get(ref1)
  const doc2 = await transaction.get(ref2)

  transaction.update(ref1, { ... })
  transaction.update(ref2, { ... })
})

// INCORRECT: Read after write causes error
await runTransaction(db, async (transaction) => {
  const doc1 = await transaction.get(ref1)
  transaction.update(ref1, { ... })
  const doc2 = await transaction.get(ref2)  // ERROR
})
```

### Maximum Document Limit

A single transaction can read or write up to 500 documents.

### Automatic Retries

Transactions automatically retry if a document changes during execution.

---

## Distributed Counters

For high-contention counters, use sharded counters:

### Creating Sharded Counter

```typescript
const NUM_SHARDS = 10

async function createCounter(counterId: string) {
  const batch = writeBatch(db)

  batch.set(doc(db, 'counters', counterId), {
    numShards: NUM_SHARDS,
    createdAt: serverTimestamp()
  })

  for (let i = 0; i < NUM_SHARDS; i++) {
    batch.set(doc(db, 'counters', counterId, 'shards', i.toString()), {
      count: 0
    })
  }

  await batch.commit()
}
```

### Incrementing Sharded Counter

```typescript
import { increment } from 'firebase/firestore'

async function incrementShardedCounter(counterId: string, value = 1) {
  const shardId = Math.floor(Math.random() * NUM_SHARDS).toString()
  const shardRef = doc(db, 'counters', counterId, 'shards', shardId)

  await updateDoc(shardRef, {
    count: increment(value)
  })
}
```

### Reading Sharded Counter Total

```typescript
async function getCounterTotal(counterId: string): Promise<number> {
  const shardsSnapshot = await getDocs(
    collection(db, 'counters', counterId, 'shards')
  )

  let total = 0
  shardsSnapshot.forEach((shard) => {
    total += shard.data().count
  })

  return total
}
```

---

## Field Value Operations

### Atomic Increment/Decrement

```typescript
import { updateDoc, doc, increment } from 'firebase/firestore'

await updateDoc(doc(db, 'users', userId), {
  loginCount: increment(1),
  score: increment(10)
})

await updateDoc(doc(db, 'users', userId), {
  credits: increment(-5)
})
```

### Array Operations

```typescript
import { updateDoc, arrayUnion, arrayRemove } from 'firebase/firestore'

await updateDoc(doc(db, 'documents', docId), {
  tags: arrayUnion('new-tag', 'another-tag'),
  collaborators: arrayUnion(userId)
})

await updateDoc(doc(db, 'documents', docId), {
  tags: arrayRemove('old-tag'),
  collaborators: arrayRemove(userId)
})
```

### Server Timestamp

```typescript
import { updateDoc, serverTimestamp } from 'firebase/firestore'

await updateDoc(doc(db, 'documents', docId), {
  updatedAt: serverTimestamp(),
  lastModifiedBy: userId
})
```

### Delete Field

```typescript
import { updateDoc, deleteField } from 'firebase/firestore'

await updateDoc(doc(db, 'users', userId), {
  temporaryToken: deleteField(),
  legacyField: deleteField()
})
```

---

## Error Handling

### Transaction Failures

```typescript
import { FirestoreError } from 'firebase/firestore'

async function safeTransaction() {
  try {
    await runTransaction(db, async (transaction) => {
      // Transaction logic
    })
  } catch (error) {
    if (error instanceof FirestoreError) {
      switch (error.code) {
        case 'aborted':
          console.error('Transaction aborted due to conflict')
          break
        case 'failed-precondition':
          console.error('Transaction precondition failed')
          break
        case 'cancelled':
          console.error('Transaction cancelled')
          break
        default:
          console.error('Transaction error:', error.message)
      }
    }
    throw error
  }
}
```

---

## Advanced Patterns

### Idempotent Transactions

```typescript
async function idempotentOperation(
  operationId: string,
  userId: string,
  amount: number
) {
  await runTransaction(db, async (transaction) => {
    const operationRef = doc(db, 'operations', operationId)
    const operationDoc = await transaction.get(operationRef)

    if (operationDoc.exists()) {
      console.log('Operation already completed')
      return
    }

    const userRef = doc(db, 'users', userId)
    const userDoc = await transaction.get(userRef)

    transaction.update(userRef, {
      balance: userDoc.data().balance + amount
    })

    transaction.set(operationRef, {
      userId,
      amount,
      completedAt: serverTimestamp()
    })
  })
}
```

### Chunked Batch Operations

For operations exceeding 500 documents:

```typescript
async function bulkUpdate(
  updates: Array<{ id: string; data: any }>,
  collectionPath: string
) {
  const BATCH_SIZE = 500
  const chunks: Array<Array<{ id: string; data: any }>> = []

  for (let i = 0; i < updates.length; i += BATCH_SIZE) {
    chunks.push(updates.slice(i, i + BATCH_SIZE))
  }

  for (const chunk of chunks) {
    const batch = writeBatch(db)

    chunk.forEach(({ id, data }) => {
      batch.update(doc(db, collectionPath, id), {
        ...data,
        updatedAt: serverTimestamp()
      })
    })

    await batch.commit()
  }
}
```

---

## Best Practices

### Keep Transactions Small

```typescript
// GOOD: Minimal operations
await runTransaction(db, async (transaction) => {
  const doc = await transaction.get(ref)
  transaction.update(ref, { count: doc.data().count + 1 })
})
```

### Use increment() for Counters

```typescript
// GOOD: Atomic increment, no transaction needed
await updateDoc(ref, { count: increment(1) })
```

### Handle Contention

```typescript
async function incrementWithJitter(ref: DocumentReference) {
  const delay = Math.random() * 100
  await new Promise((r) => setTimeout(r, delay))
  await updateDoc(ref, { count: increment(1) })
}
```

---

Status: Module Documentation
Last Updated: 2026-01-06
