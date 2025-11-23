# Advanced Firebase Patterns

Advanced patterns for complex Firebase implementations.

## Query Optimization with Composite Indexes

Firebase Firestore requires composite indexes for complex queries with multiple filters and ordering:

```python
# Query requiring composite index
db.collection('posts')
  .where('status', '==', 'published')
  .where('category', '==', 'tech')
  .orderBy('createdAt', 'desc')
  .limit(10)
```

**Index Strategy**:
- Let Firestore auto-create indexes (usually works)
- For large collections (>50K docs), pre-create indexes
- Use compound indexes wisely (they have write overhead)
- Review unused indexes monthly

## Cursor-Based Pagination for Large Datasets

```typescript
interface PaginationState {
  pageSize: number;
  currentCursor: DocumentSnapshot | null;
  hasMore: boolean;
}

async function getPage<T>(
  collectionRef: CollectionReference,
  state: PaginationState
): Promise<{ data: T[]; nextCursor: DocumentSnapshot | null }> {
  let q = query(
    collectionRef,
    orderBy('createdAt', 'desc'),
    limit(state.pageSize + 1)
  );

  if (state.currentCursor) {
    q = query(q, startAfter(state.currentCursor));
  }

  const snapshot = await getDocs(q);
  const docs = snapshot.docs;

  const hasMore = docs.length > state.pageSize;
  const data = hasMore ? docs.slice(0, -1) : docs;

  return {
    data: data.map(doc => ({ id: doc.id, ...doc.data() } as T)),
    nextCursor: hasMore ? docs[docs.length - 2] : null,
  };
}
```

## Real-Time Collaboration with Operational Transformation

For collaborative editing, implement operational transformation:

```typescript
interface CollaborativeDocument {
  id: string;
  content: string;
  version: number;
  collaborators: Map<string, ClientState>;
  operations: Operation[];
}

interface Operation {
  userId: string;
  timestamp: number;
  position: number;
  content: string;
  type: 'insert' | 'delete';
  version: number;
}

class CollaborativeEditorManager {
  async applyRemoteOperation(
    docId: string,
    operation: Operation
  ): Promise<void> {
    const db = getFirestore();
    const docRef = doc(db, 'collaborative_docs', docId);

    await runTransaction(db, async (transaction) => {
      const docSnapshot = await transaction.get(docRef);
      const currentVersion = docSnapshot.get('version');

      if (operation.version !== currentVersion) {
        const transformedOp = this.transformOperation(
          operation,
          currentVersion - operation.version
        );

        transaction.update(docRef, {
          content: applyOperation(docSnapshot.get('content'), transformedOp),
          version: currentVersion + 1,
        });
      } else {
        transaction.update(docRef, {
          content: applyOperation(docSnapshot.get('content'), operation),
          version: currentVersion + 1,
        });
      }
    });
  }

  private transformOperation(op: Operation, versionDiff: number): Operation {
    return op;
  }
}
```

## Security Rules Best Practices

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // User profile - only user can read/write own profile
    match /users/{userId} {
      allow read, write: if request.auth.uid == userId;
    }

    // Public posts - anyone can read, authors can write
    match /posts/{postId} {
      allow read: if true;
      allow write: if request.auth.uid == resource.data.authorId;
    }

    // Team documents - only team members
    match /teams/{teamId}/documents/{docId} {
      allow read, write: if request.auth.uid in resource.data.memberIds;
    }

    // Admin-only
    match /admin/{adminId} {
      allow read, write: if 'admin' in request.auth.token.customClaims;
    }
  }
}
```

## Cost Optimization Strategies

1. **Avoid unnecessary reads**: Use collection group queries carefully
2. **Denormalize data**: Store computed values to avoid reads
3. **Batch operations**: Use batch writes (500 max per batch)
4. **Index cleanup**: Delete unused indexes monthly
5. **Query monitoring**: Use Cloud Logging to find expensive queries

```python
# Monitor query costs
def log_query_metrics(query_name: str, docs_read: int):
    cost = docs_read * 0.06 / 1_000_000  # $0.06 per 1M reads
    print(f"Query: {query_name}")
    print(f"Documents read: {docs_read}")
    print(f"Estimated cost: ${cost:.6f}")
```

## Performance Monitoring

```typescript
class FirebasePerformanceMonitor {
  measureQuery<T>(
    name: string,
    queryFn: () => Promise<T>
  ): Promise<T> {
    const start = performance.now();

    return queryFn().then((result) => {
      const duration = performance.now() - start;
      console.log(`Query ${name} took ${duration.toFixed(2)}ms`);

      if (duration > 1000) {
        console.warn(`Query ${name} is slow!`);
      }

      return result;
    });
  }
}

Move advanced content here:
- Complex implementation patterns
- Edge cases and error handling
- Performance optimization techniques
- Advanced configuration options
- Troubleshooting guides

## How to organize

Organize advanced patterns in a way that makes sense for this skill's domain.
Each subsection should be self-contained and referenceable from SKILL.md.

