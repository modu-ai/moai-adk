---
name: moai-baas-firebase-firestore
description: Firestore database operations, real-time subscriptions, and data modeling patterns
---

## Firestore Operations & Real-time Features

### Advanced Firestore Implementation

```typescript
// Enterprise Firestore operations with TypeScript
import { getFirestore, collection, doc, setDoc, getDoc, updateDoc, deleteDoc, query, where, orderBy, limit, onSnapshot } from 'firebase-admin/firestore';

export class FirestoreManager {
  private firestore: Firestore;

  constructor() {
    this.firestore = getFirestore();
  }

  // Batch operations for performance
  async batchUpdateDocuments(
    updates: Array<{ collection: string; docId: string; data: any }>
  ): Promise<void> {
    const batch = this.firestore.batch();
    
    for (const update of updates) {
      const docRef = doc(this.firestore, update.collection, update.docId);
      batch.set(docRef, {
        ...update.data,
        updatedAt: new Date(),
        updatedBy: 'system',
      }, { merge: true });
    }

    await batch.commit();
  }

  // Real-time subscription with advanced filtering
  subscribeToRealtimeUpdates<T>(
    collectionPath: string,
    filters: QueryFilter[] = [],
    callback: (data: T[]) => void
  ): () => void {
    let queryRef = collection(this.firestore, collectionPath);
    
    // Apply filters
    for (const filter of filters) {
      if (filter.type === 'where') {
        queryRef = query(queryRef, where(filter.field, filter.operator, filter.value));
      } else if (filter.type === 'orderBy') {
        queryRef = query(queryRef, orderBy(filter.field, filter.direction));
      } else if (filter.type === 'limit') {
        queryRef = query(queryRef, limit(filter.value));
      }
    }

    const unsubscribe = onSnapshot(
      queryRef,
      (snapshot) => {
        const data: T[] = [];
        snapshot.forEach((doc) => {
          data.push({ id: doc.id, ...doc.data() } as T);
        });
        callback(data);
      },
      (error) => {
        console.error('Real-time subscription error:', error);
      }
    );

    return unsubscribe;
  }

  // Transaction for consistency
  async transferCredits(fromUserId: string, toUserId: string, amount: number): Promise<void> {
    const fromRef = doc(this.firestore, 'users', fromUserId);
    const toRef = doc(this.firestore, 'users', toUserId);

    await this.firestore.runTransaction(async (transaction) => {
      const fromDoc = await transaction.get(fromRef);
      const toDoc = await transaction.get(toRef);

      if (!fromDoc.exists() || !toDoc.exists()) {
        throw new Error('User not found');
      }

      const fromBalance = fromDoc.data().balance;
      if (fromBalance < amount) {
        throw new Error('Insufficient balance');
      }

      transaction.update(fromRef, { balance: fromBalance - amount });
      transaction.update(toRef, { balance: toDoc.data().balance + amount });
    });
  }
}
```

---

### Data Modeling Best Practices

**Denormalization Pattern**:
```typescript
// Denormalize for read performance
interface UserProfile {
  id: string;
  name: string;
  email: string;
  recentOrders: Array<{  // Denormalized
    orderId: string;
    orderDate: Date;
    totalAmount: number;
  }>;
}

// Update denormalized data on write
async function createOrder(userId: string, order: Order) {
  const orderRef = doc(firestore, 'orders', order.id);
  const userRef = doc(firestore, 'users', userId);

  const batch = firestore.batch();
  batch.set(orderRef, order);
  batch.update(userRef, {
    recentOrders: arrayUnion({
      orderId: order.id,
      orderDate: order.date,
      totalAmount: order.total,
    }),
  });

  await batch.commit();
}
```

---

### Query Optimization

**Indexing Strategy**:
```javascript
// firestore.indexes.json
{
  "indexes": [
    {
      "collectionGroup": "orders",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "userId", "order": "ASCENDING" },
        { "fieldPath": "createdAt", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "products",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "category", "order": "ASCENDING" },
        { "fieldPath": "price", "order": "ASCENDING" },
        { "fieldPath": "rating", "order": "DESCENDING" }
      ]
    }
  ]
}
```

---

### Security Rules

**Firestore Security Rules**:
```javascript
// firestore.rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // User data - owner only
    match /users/{userId} {
      allow read: if request.auth != null && request.auth.uid == userId;
      allow write: if request.auth != null && request.auth.uid == userId;
    }

    // Public read, authenticated write
    match /posts/{postId} {
      allow read: if true;
      allow create: if request.auth != null;
      allow update, delete: if request.auth != null && 
                              resource.data.authorId == request.auth.uid;
    }

    // Admin-only access
    match /admin/{document=**} {
      allow read, write: if request.auth != null && 
                          request.auth.token.admin == true;
    }
  }
}
```

---

**End of Module** | moai-baas-firebase-firestore
