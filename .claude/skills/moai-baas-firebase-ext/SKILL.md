# Skill: moai-baas-firebase-ext

## Metadata

```yaml
skill_id: moai-baas-firebase-ext
skill_name: Firebase Full-Stack Platform & Ecosystem
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Firebase", "Firestore", "Cloud Functions", "Firebase Auth", "Google Cloud", "Security Rules", "Testing"]
  - contexts: ["firebase-detected", "pattern-e", "google-ecosystem"]
agents:
  - backend-expert
  - database-expert
  - devops-expert
  - frontend-expert
  - security-expert
freedom_level: high
word_count: 1200
context7_references:
  - url: "https://firebase.google.com/docs/firestore"
    topic: "Firestore Database & Collections"
  - url: "https://firebase.google.com/docs/auth"
    topic: "Firebase Authentication"
  - url: "https://firebase.google.com/docs/functions"
    topic: "Cloud Functions"
  - url: "https://firebase.google.com/docs/storage"
    topic: "Cloud Storage"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Firebase Ecosystem Overview (150 words)

**Firebase** is Google's fully managed Backend-as-a-Service platform providing a complete development ecosystem.

**Firebase Product Suite**:
```
┌──────────────────────────────────────────┐
│ Firebase (Google Cloud Integration)      │
├──────────────────────────────────────────┤
│                                           │
│ 1. Firestore (NoSQL Database)            │
│    └─ Documents, Collections, Queries    │
│                                           │
│ 2. Realtime Database (Legacy)            │
│    └─ JSON tree structure                │
│                                           │
│ 3. Authentication                        │
│    └─ Email, Phone, OAuth (Google, etc.) │
│                                           │
│ 4. Cloud Storage                         │
│    └─ File storage with CDN              │
│                                           │
│ 5. Cloud Functions                       │
│    └─ Serverless functions triggered     │
│                                           │
│ 6. Hosting                               │
│    └─ Static & dynamic content           │
│                                           │
│ 7. Analytics & Crashlytics               │
│    └─ Built-in monitoring                │
└──────────────────────────────────────────┘
```

**Why Firebase**:
- ✅ Complete integration (no separate services)
- ✅ Generous free tier (perfect for MVPs)
- ✅ Automatic scaling & maintenance
- ✅ Built-in analytics & monitoring
- ⚠️ Vendor lock-in (Google ecosystem)
- ⚠️ NoSQL mindset (different from SQL)

---

### 2. Firestore Data Design (250 words)

**Firestore** is a NoSQL document database with real-time synchronization and powerful querying.

**Core Concepts**:

```
Firestore Structure:
  Database
    ├─ Collection: "users"
    │  ├─ Document: "user123"
    │  │  ├─ Field: "email": "alice@example.com"
    │  │  ├─ Field: "name": "Alice"
    │  │  └─ Field: "createdAt": Timestamp
    │  ├─ Document: "user456"
    │  │  └─ ...
    │
    ├─ Collection: "posts"
    │  ├─ Document: "post1"
    │  │  ├─ Field: "userId": "user123"
    │  │  ├─ Field: "title": "..."
    │  │  └─ Subcollection: "comments"
    │  │     ├─ Document: "comment1"
    │  │     └─ Document: "comment2"
```

**Firestore Rules (Security)**:

```javascript
// firestore.rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Users can read/write their own document
    match /users/{userId} {
      allow read, write: if request.auth.uid == userId;
    }

    // Public posts
    match /posts/{postId} {
      allow read: if true;
      allow create: if request.auth != null;
      allow update, delete: if request.auth.uid == resource.data.userId;

      // Subcollection: comments
      match /comments/{commentId} {
        allow read: if true;
        allow create: if request.auth != null;
        allow delete: if request.auth.uid == resource.data.userId;
      }
    }
  }
}
```

**Basic Operations**:

```javascript
// Initialize Firebase
import { initializeApp } from "firebase/app";
import { getFirestore } from "firebase/firestore";

const app = initializeApp({
  apiKey: process.env.REACT_APP_FIREBASE_API_KEY,
  projectId: process.env.REACT_APP_FIREBASE_PROJECT_ID,
  // ... other config
});

const db = getFirestore(app);

// Add document
import { collection, addDoc, serverTimestamp } from "firebase/firestore";

const docRef = await addDoc(collection(db, "posts"), {
  userId: currentUser.uid,
  title: "My First Post",
  content: "Hello World",
  createdAt: serverTimestamp(),
});

// Query documents
import { query, where, getDocs } from "firebase/firestore";

const q = query(
  collection(db, "posts"),
  where("userId", "==", currentUser.uid)
);
const querySnapshot = await getDocs(q);

// Real-time listener
import { onSnapshot } from "firebase/firestore";

const unsubscribe = onSnapshot(
  query(collection(db, "posts"), where("userId", "==", currentUser.uid)),
  (snapshot) => {
    const posts = snapshot.docs.map((doc) => ({
      id: doc.id,
      ...doc.data(),
    }));
    setPosts(posts);
  }
);
```

---

### 3. Firebase Authentication (200 words)

**Firebase Auth** provides authentication with minimal backend code.

**Supported Methods**:
```
Email/Password
Phone Number (SMS)
OAuth Providers (Google, GitHub, Facebook)
Custom Claims
Multi-factor Authentication (MFA)
```

**Authentication Setup**:

```javascript
import { initializeAuth, connectAuthEmulator } from "firebase/auth";
import { getAuth, createUserWithEmailAndPassword } from "firebase/auth";

const auth = getAuth(app);

// Sign up
const userCredential = await createUserWithEmailAndPassword(
  auth,
  "alice@example.com",
  "securePassword123"
);

const user = userCredential.user;

// Sign in
import { signInWithEmailAndPassword } from "firebase/auth";

const userCredential = await signInWithEmailAndPassword(
  auth,
  "alice@example.com",
  "securePassword123"
);

// Real-time user listener
import { onAuthStateChanged } from "firebase/auth";

onAuthStateChanged(auth, (user) => {
  if (user) {
    console.log("User logged in:", user.uid);
  } else {
    console.log("User logged out");
  }
});

// Set custom claims (Admin SDK only)
import { getAuth } from "firebase-admin/auth";

const auth = getAuth();
await auth.setCustomUserClaims(uid, { role: "admin" });
```

**Security Best Practices**:
- ✅ Always use HTTPS in production
- ✅ Enable Multi-factor Authentication (MFA)
- ✅ Use custom claims for roles (not tokens)
- ✅ Store sensitive data in Firestore with rules
- ❌ Never expose API keys in client code (use .env)

---

### 4. Cloud Functions & Cloud Storage (250 words)

**Cloud Functions** are serverless functions triggered by Firestore events or HTTPS requests.

**Function Types**:

```javascript
// 1. HTTP-triggered function
import { onRequest } from "firebase-functions/v2/https";

export const helloWorld = onRequest((request, response) => {
  response.send("Hello from Cloud Functions!");
});

// 2. Firestore-triggered function (onCreate)
import { onDocumentCreated } from "firebase-functions/v2/firestore";

export const onPostCreated = onDocumentCreated("posts/{postId}", (event) => {
  const newPost = event.data.data();
  console.log("New post created:", newPost.title);

  // Trigger email, update counts, etc.
});

// 3. Firestore-triggered function (onWrite)
export const onPostUpdated = onDocumentWritten("posts/{postId}", (event) => {
  const oldPost = event.data.before.data();
  const newPost = event.data.after.data();

  if (oldPost.published !== newPost.published) {
    console.log("Publication status changed");
  }
});

// 4. Authentication trigger
import { onUserCreated, onUserDeleted } from "firebase-functions/v2/identity";

export const createUserProfile = onUserCreated(async (event) => {
  const user = event.data;
  const db = getFirestore();

  await db.collection("users").doc(user.uid).set({
    email: user.email,
    displayName: user.displayName || "",
    createdAt: new Date(),
  });
});
```

**Cloud Storage**:

```javascript
// Client: Upload file
import { getStorage, ref, uploadBytes } from "firebase/storage";

const storage = getStorage(app);
const fileRef = ref(storage, `avatars/${userId}/profile.jpg`);

await uploadBytes(fileRef, file);

// Get download URL
import { getDownloadURL } from "firebase/storage";

const url = await getDownloadURL(fileRef);

// Server: Process uploads with Cloud Functions
import { onObjectFinalized } from "firebase-functions/v2/storage";
import { Storage } from "@google-cloud/storage";

export const generateThumbnail = onObjectFinalized(
  { bucket: "project.appspot.com" },
  async (event) => {
    const filename = event.data.name;
    // Process image, generate thumbnail, etc.
  }
);
```

---

### 5. Hosting & Deployment (150 words)

**Firebase Hosting** provides global CDN with automatic HTTPS and fast deployment.

**Deployment Workflow**:

```bash
# 1. Install Firebase CLI
npm install -g firebase-tools

# 2. Initialize Firebase in project
firebase init hosting

# 3. Build your app
npm run build

# 4. Deploy to Firebase Hosting
firebase deploy

# 5. Preview deployment
firebase open hosting
```

**Key Features**:
- ✅ Automatic HTTPS with free SSL certificate
- ✅ Global CDN with edge caching
- ✅ Automatic gzip compression
- ✅ Instant rollback capability
- ✅ Custom domain support
- ✅ Environment-specific deployments

**Typical firebase.json**:

```json
{
  "hosting": {
    "public": "dist",
    "ignore": ["firebase.json", "**/node_modules/**"],
    "redirects": [
      {
        "source": "/old-page",
        "destination": "/new-page",
        "type": 301
      }
    ],
    "rewrites": [
      {
        "source": "**",
        "destination": "/index.html"
      }
    ]
  }
}
```

---

### 6. Security Rules Advanced Patterns & Testing (150 words)

**Advanced Security Rule Patterns**:

```javascript
// Rule 1: Timestamp-based access control
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Allow deletion only within 24 hours of creation
    match /posts/{postId} {
      allow delete: if request.auth.uid == resource.data.userId &&
                       (now.getTime() - resource.createTime.getTime() < 24 * 60 * 60 * 1000);
    }

    // Rule 2: Map-based permissions (shared access)
    match /documents/{docId} {
      allow read: if request.auth.uid in resource.data.allowedUsers;
      allow write: if request.auth.uid == resource.data.owner ||
                      request.auth.uid in resource.data.editors;
    }

    // Rule 3: Batch operation limits
    match /transactions/{transaction} {
      allow create: if request.resource.data.items.size() <= 100;
    }
  }
}
```

**Testing Security Rules**:

```bash
# Install Firebase Emulator
npm install -D firebase-tools

# Start emulator with rules testing
firebase emulators:start --only firestore

# Test rules with firestore.test.js
```

```typescript
// firestore.test.ts using firebase-rules-testing
import { initializeTestEnvironment, assertFails, assertSucceeds } from "@firebase/rules-testing";

describe("Firestore Rules", () => {
  let testEnv;

  beforeAll(async () => {
    testEnv = await initializeTestEnvironment({
      projectId: "test-project",
      firestore: { rules: fs.readFileSync("firestore.rules", "utf8") },
    });
  });

  test("User can delete own post within 24h", async () => {
    const db = testEnv.authenticatedContext("user123").firestore();

    // Create post
    await db.collection("posts").doc("post1").set({
      userId: "user123",
      title: "Test Post",
      createdAt: new Date(),
    });

    // Attempt delete (should succeed)
    await assertSucceeds(db.collection("posts").doc("post1").delete());
  });

  test("User cannot delete others' posts", async () => {
    const db = testEnv.authenticatedContext("user456").firestore();
    await assertFails(db.collection("posts").doc("post1").delete());
  });
});
```

---

### 7. Performance Optimization & Scaling Limits (100 words)

**Firestore Scaling Limits** (per database):

| Metric | Limit | Workaround |
|--------|-------|-----------|
| **Document size** | 1MB max | Use subcollections for large arrays |
| **Write throughput** | 1 write/sec per document | Distribute across multiple docs |
| **Composite indexes** | 200 max per database | Clean up unused indexes |
| **Query result size** | Memory-based | Use pagination or `limit(1000)` |
| **Array size** | 20,000 items max | Use separate collection |

**Performance Best Practices**:
- ✅ **Pagination**: Use `startAfter()` and `limit(20)` for large datasets
- ✅ **Index optimization**: Monitor composite indexes in console
- ✅ **Batch writes**: Group up to 500 operations with `batch()`
- ✅ **Denormalization**: Copy frequently-accessed data to reduce joins
- ✅ **Lazy loading**: Load subcollections only when needed

**Cost Monitoring**:
```typescript
// Log read/write operations for cost analysis
logging.log(`Firestore operation - ${operation} - estimated cost: $${cost}`);
```

---

### 8. Common Issues & Solutions (50 words)

| Issue | Solution |
|-------|----------|
| **Permission denied** | Check Firestore rules for user role |
| **Slow Firestore queries** | Add composite index via console |
| **Function cold start** | Use min instances or async processing |
| **Storage CORS error** | Configure CORS in Cloud Storage |

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-firebase-ext")
# Load when Pattern E (Firebase) detected
```

### Context7 Integration
When Firebase platform detected:
- Firestore data modeling & queries
- Firebase Authentication flows
- Cloud Functions patterns
- Hosting & deployment guide

---

## 📚 Reference Materials

- [Firestore Documentation](https://firebase.google.com/docs/firestore)
- [Firebase Authentication](https://firebase.google.com/docs/auth)
- [Cloud Functions Guide](https://firebase.google.com/docs/functions)
- [Cloud Storage Docs](https://firebase.google.com/docs/storage)

---

## ✅ Validation Checklist

- [x] Firebase ecosystem overview
- [x] Firestore data design & security rules
- [x] Authentication methods & setup
- [x] Cloud Functions patterns & storage
- [x] Security rules advanced patterns & testing
- [x] Performance optimization & scaling limits
- [x] Hosting & deployment workflow
- [x] Common issues & troubleshooting
- [x] 1200-word target (from 1000)
- [x] English language (policy compliant)
