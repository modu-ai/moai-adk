---
title: "Firebase 완벽 가이드"
description: "Google Cloud Native BaaS - Firestore NoSQL, Cloud Functions, Firebase Auth로 모바일 최적화 백엔드 구축"
---

# Firebase 완벽 가이드

> **Google Cloud Native BaaS**: NoSQL (Firestore), Serverless Functions, ML Kit로 모바일 우선 애플리케이션을 빠르게 구축

## 1. Firebase란?

### 개요
```yaml
firebase:
  tagline: "Google's Mobile and Web App Development Platform"
  category: "NoSQL-based BaaS"
  managed_by: "Google Cloud"

  core_stack:
    database:
      - "Firestore (NoSQL Document DB)"
      - "Realtime Database (JSON tree)"
    functions: "Cloud Functions (Node.js/Python)"
    auth: "Firebase Authentication"
    storage: "Cloud Storage"
    hosting: "Firebase Hosting"
    ml: "ML Kit (On-device ML)"

  pricing: "Spark (Free) → Blaze (Pay-as-you-go)"
```

### 왜 Firebase인가?

**장점**:
- ✅ **모바일 최적화**: Android/iOS SDK 우수
- ✅ **빠른 개발**: NoSQL로 스키마 변경 자유
- ✅ **Google 통합**: Analytics, AdMob, Cloud 연계
- ✅ **실시간 동기화**: 자동 오프라인 지원

**단점**:
- ❌ **SQL 불가**: 복잡한 쿼리 제한
- ❌ **벤더 종속**: Google Cloud 의존
- ❌ **비용**: 대규모 시 Firestore 읽기/쓰기 비용 증가

## 2. 핵심 기능

### 2.1 Firestore (NoSQL Database)

#### Document 구조
```typescript
// Collection → Document → Subcollection
firebase/
  users/                   # Collection
    user123/               # Document
      name: "John Doe"
      email: "john@example.com"
      posts/               # Subcollection
        post1/             # Document
          title: "Hello"
          content: "..."
```

#### CRUD 작업
```typescript
import {
  getFirestore,
  collection,
  doc,
  getDoc,
  setDoc,
  updateDoc,
  deleteDoc,
  query,
  where,
  orderBy,
  limit,
  getDocs
} from 'firebase/firestore'

const db = getFirestore()

// Create
await setDoc(doc(db, 'posts', 'post123'), {
  title: 'Hello Firebase',
  author: 'John',
  published: true,
  createdAt: new Date()
})

// Read
const docSnap = await getDoc(doc(db, 'posts', 'post123'))
if (docSnap.exists()) {
  console.log(docSnap.data())
}

// Update
await updateDoc(doc(db, 'posts', 'post123'), {
  title: 'Updated Title',
  updatedAt: new Date()
})

// Delete
await deleteDoc(doc(db, 'posts', 'post123'))
```

#### Queries
```typescript
// 간단한 쿼리
const q = query(
  collection(db, 'posts'),
  where('published', '==', true),
  where('author', '==', 'John'),
  orderBy('createdAt', 'desc'),
  limit(10)
)

const querySnapshot = await getDocs(q)
querySnapshot.forEach((doc) => {
  console.log(doc.id, doc.data())
})

// Composite queries (복합 인덱스 필요)
const q2 = query(
  collection(db, 'posts'),
  where('category', '==', 'tech'),
  where('views', '>', 100),
  orderBy('views', 'desc')
)
// → Firebase Console에서 자동 인덱스 생성 링크 제공
```

#### Realtime Listeners
```typescript
import { onSnapshot } from 'firebase/firestore'

// 실시간 구독
const unsubscribe = onSnapshot(
  doc(db, 'posts', 'post123'),
  (doc) => {
    console.log('Current data:', doc.data())
    // 변경 즉시 콜백 호출
  }
)

// Collection 구독
const unsubscribeCollection = onSnapshot(
  query(collection(db, 'posts'), where('published', '==', true)),
  (snapshot) => {
    snapshot.docChanges().forEach((change) => {
      if (change.type === 'added') {
        console.log('New post:', change.doc.data())
      }
      if (change.type === 'modified') {
        console.log('Modified post:', change.doc.data())
      }
      if (change.type === 'removed') {
        console.log('Removed post:', change.doc.data())
      }
    })
  }
)

// 정리
unsubscribe()
```

#### Security Rules
```javascript
// firestore.rules
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    // 공개 읽기, 인증된 사용자만 쓰기
    match /posts/{postId} {
      allow read: if true;
      allow write: if request.auth != null;
    }

    // 소유자만 수정 가능
    match /users/{userId} {
      allow read: if request.auth != null;
      allow write: if request.auth.uid == userId;
    }

    // Role-based 접근
    match /admin/{document=**} {
      allow read, write: if get(/databases/$(database)/documents/users/$(request.auth.uid)).data.role == 'admin';
    }

    // Custom claims 사용
    match /premium/{document=**} {
      allow read: if request.auth.token.premium == true;
    }
  }
}
```

### 2.2 Cloud Functions

#### HTTP Functions
```typescript
import { onRequest } from 'firebase-functions/v2/https'
import { initializeApp } from 'firebase-admin/app'
import { getFirestore } from 'firebase-admin/firestore'

initializeApp()

// HTTP 트리거
export const api = onRequest(async (request, response) => {
  const db = getFirestore()

  // CORS 처리
  response.set('Access-Control-Allow-Origin', '*')

  if (request.method === 'OPTIONS') {
    response.set('Access-Control-Allow-Methods', 'GET, POST')
    response.status(204).send('')
    return
  }

  // POST /api
  if (request.method === 'POST') {
    const data = request.body

    const docRef = await db.collection('posts').add({
      ...data,
      createdAt: new Date()
    })

    response.json({ id: docRef.id, success: true })
  }
})
```

#### Firestore Triggers
```typescript
import { onDocumentCreated, onDocumentUpdated } from 'firebase-functions/v2/firestore'

// Document 생성 시 트리거
export const onPostCreated = onDocumentCreated(
  'posts/{postId}',
  async (event) => {
    const snapshot = event.data
    const data = snapshot?.data()

    // 새 포스트 알림 전송
    await sendNotification({
      title: 'New Post',
      body: data?.title
    })

    // 통계 업데이트
    const db = getFirestore()
    await db.collection('stats').doc('posts').update({
      totalPosts: FieldValue.increment(1)
    })
  }
)

// Document 업데이트 시 트리거
export const onPostUpdated = onDocumentUpdated(
  'posts/{postId}',
  async (event) => {
    const before = event.data?.before.data()
    const after = event.data?.after.data()

    if (before?.published === false && after?.published === true) {
      // 게시 시 알림
      await notifySubscribers(after)
    }
  }
)
```

#### Scheduled Functions
```typescript
import { onSchedule } from 'firebase-functions/v2/scheduler'
import { getFirestore } from 'firebase-admin/firestore'

// 매일 자정 실행
export const dailyCleanup = onSchedule('0 0 * * *', async () => {
  const db = getFirestore()

  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() - 30)  // 30일 전

  const snapshot = await db
    .collection('logs')
    .where('createdAt', '<', cutoff)
    .get()

  const batch = db.batch()
  snapshot.docs.forEach((doc) => {
    batch.delete(doc.ref)
  })

  await batch.commit()
  console.log(`Deleted ${snapshot.size} old logs`)
})
```

### 2.3 Firebase Authentication

#### Email/Password
```typescript
import {
  getAuth,
  createUserWithEmailAndPassword,
  signInWithEmailAndPassword,
  signOut,
  onAuthStateChanged
} from 'firebase/auth'

const auth = getAuth()

// 회원가입
const userCredential = await createUserWithEmailAndPassword(
  auth,
  email,
  password
)
const user = userCredential.user

// 로그인
await signInWithEmailAndPassword(auth, email, password)

// 로그아웃
await signOut(auth)

// 인증 상태 감지
onAuthStateChanged(auth, (user) => {
  if (user) {
    console.log('User signed in:', user.uid)
  } else {
    console.log('User signed out')
  }
})
```

#### Social Login
```typescript
import {
  GoogleAuthProvider,
  signInWithPopup,
  signInWithRedirect
} from 'firebase/auth'

// Google 로그인
const provider = new GoogleAuthProvider()
provider.addScope('profile')
provider.addScope('email')

const result = await signInWithPopup(auth, provider)
const user = result.user

// Custom claims 설정 (Admin SDK)
import { getAuth } from 'firebase-admin/auth'

await getAuth().setCustomUserClaims(user.uid, {
  premium: true,
  role: 'editor'
})
```

### 2.4 Cloud Storage

#### 파일 업로드
```typescript
import {
  getStorage,
  ref,
  uploadBytes,
  getDownloadURL,
  deleteObject
} from 'firebase/storage'

const storage = getStorage()

// 파일 업로드
const storageRef = ref(storage, `images/${userId}/${fileName}`)
const snapshot = await uploadBytes(storageRef, file)

// Download URL 획득
const url = await getDownloadURL(snapshot.ref)
console.log('File available at:', url)

// 메타데이터 설정
await uploadBytes(storageRef, file, {
  contentType: 'image/jpeg',
  customMetadata: {
    'uploaded-by': userId
  }
})

// 파일 삭제
await deleteObject(ref(storage, 'images/old-file.jpg'))
```

#### Security Rules
```javascript
// storage.rules
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {

    // 사용자별 폴더 (소유자만 접근)
    match /users/{userId}/{allPaths=**} {
      allow read, write: if request.auth.uid == userId;
    }

    // 공개 읽기, 인증된 사용자만 쓰기
    match /public/{allPaths=**} {
      allow read: if true;
      allow write: if request.auth != null;
    }

    // 파일 크기 제한
    match /uploads/{fileName} {
      allow write: if request.resource.size < 5 * 1024 * 1024  // 5MB
                   && request.resource.contentType.matches('image/.*');
    }
  }
}
```

### 2.5 Firebase Hosting

#### 배포
```bash
# Firebase CLI 설치
npm install -g firebase-tools

# 로그인
firebase login

# 프로젝트 초기화
firebase init hosting

# 배포
firebase deploy --only hosting

# Preview URL 생성
firebase hosting:channel:deploy preview-branch
# → https://project-preview-branch.web.app
```

#### 설정
```json
// firebase.json
{
  "hosting": {
    "public": "out",
    "ignore": ["firebase.json", "**/.*", "**/node_modules/**"],
    "rewrites": [
      {
        "source": "/api/**",
        "function": "api"
      },
      {
        "source": "**",
        "destination": "/index.html"
      }
    ],
    "headers": [
      {
        "source": "/assets/**",
        "headers": [
          {
            "key": "Cache-Control",
            "value": "public, max-age=31536000, immutable"
          }
        ]
      }
    ]
  }
}
```

## 3. Architecture Pattern E: Event-driven

### 시나리오: 이미지 업로드 → 썸네일 생성 → 메타데이터 저장

```typescript
// 1. 클라이언트: 이미지 업로드
import { getStorage, ref, uploadBytes } from 'firebase/storage'

async function uploadImage(file: File) {
  const storage = getStorage()
  const imageRef = ref(storage, `images/${userId}/${file.name}`)

  await uploadBytes(imageRef, file)
  console.log('Uploaded image')
}

// 2. Cloud Function: Storage 트리거
import { onObjectFinalized } from 'firebase-functions/v2/storage'
import sharp from 'sharp'

export const generateThumbnail = onObjectFinalized(async (event) => {
  const filePath = event.data.name
  const contentType = event.data.contentType

  // 이미지 파일만 처리
  if (!contentType?.startsWith('image/')) {
    return
  }

  // 원본 이미지 다운로드
  const bucket = getStorage().bucket()
  const file = bucket.file(filePath)
  const [buffer] = await file.download()

  // Sharp로 썸네일 생성
  const thumbnail = await sharp(buffer)
    .resize(200, 200, { fit: 'cover' })
    .jpeg({ quality: 80 })
    .toBuffer()

  // 썸네일 업로드
  const thumbnailPath = filePath.replace(/(\.[^.]+)$/, '_thumb$1')
  await bucket.file(thumbnailPath).save(thumbnail, {
    contentType: 'image/jpeg'
  })

  // Firestore에 메타데이터 저장
  const db = getFirestore()
  await db.collection('images').add({
    originalPath: filePath,
    thumbnailPath: thumbnailPath,
    uploadedBy: event.data.metadata?.uploadedBy,
    createdAt: new Date()
  })

  console.log('Thumbnail generated:', thumbnailPath)
})

// 3. Firestore 트리거: 알림 전송
import { onDocumentCreated } from 'firebase-functions/v2/firestore'

export const onImageUploaded = onDocumentCreated(
  'images/{imageId}',
  async (event) => {
    const data = event.data?.data()

    // FCM 푸시 알림 전송
    await getMessaging().send({
      topic: 'new-images',
      notification: {
        title: 'New Image Uploaded',
        body: `Image uploaded at ${data?.createdAt}`
      }
    })
  }
)
```

## 4. Performance Optimization

### Firestore 최적화

#### Batch Writes
```typescript
import { writeBatch } from 'firebase/firestore'

const batch = writeBatch(db)

// 여러 작업을 하나의 트랜잭션으로
posts.forEach((post) => {
  const docRef = doc(collection(db, 'posts'))
  batch.set(docRef, post)
})

// 한 번에 커밋 (최대 500개)
await batch.commit()
```

#### Pagination
```typescript
import { query, limit, startAfter, getDocs } from 'firebase/firestore'

let lastVisible: any = null

async function loadMorePosts() {
  let q = query(
    collection(db, 'posts'),
    orderBy('createdAt', 'desc'),
    limit(10)
  )

  if (lastVisible) {
    q = query(q, startAfter(lastVisible))
  }

  const snapshot = await getDocs(q)
  lastVisible = snapshot.docs[snapshot.docs.length - 1]

  return snapshot.docs.map(doc => doc.data())
}
```

#### Offline Persistence
```typescript
import { enableIndexedDbPersistence } from 'firebase/firestore'

// 오프라인 캐싱 활성화
await enableIndexedDbPersistence(db)

// 쿼리 결과 자동 캐싱
const posts = await getDocs(collection(db, 'posts'))
// → 오프라인에서도 캐시된 데이터 반환
```

### Cloud Functions 최적화

#### Cold Start 최소화
```typescript
import { onRequest } from 'firebase-functions/v2/https'
import { setGlobalOptions } from 'firebase-functions/v2'

// Gen 2 함수 설정
setGlobalOptions({
  region: 'asia-northeast3',  // Seoul
  memory: '512MiB',
  timeoutSeconds: 60,
  minInstances: 1  // Cold start 방지
})

export const api = onRequest(
  {
    concurrency: 80  // 인스턴스당 동시 요청 수
  },
  async (req, res) => {
    // ...
  }
)
```

## 5. 비용 최적화

### Spark (Free) vs Blaze (Pay-as-you-go)

```yaml
spark_plan:
  firestore:
    stored: "1 GiB"
    reads: "50,000/day"
    writes: "20,000/day"
    deletes: "20,000/day"

  functions:
    invocations: "125,000/month"
    compute: "40,000 GB-seconds/month"

  storage: "5 GB"
  hosting: "10 GB/month"

blaze_plan:
  firestore:
    stored: "$0.18/GiB"
    reads: "$0.06 per 100,000"
    writes: "$0.18 per 100,000"
    deletes: "$0.02 per 100,000"

  functions:
    invocations: "$0.40 per million"
    compute: "$0.0000025 per GB-second"
```

### 비용 절감 팁

```typescript
// 1. Selective fields (not entire document)
const docRef = doc(db, 'posts', 'post123')
const docSnap = await getDoc(docRef)
const { title, author } = docSnap.data()  // 필요한 필드만 사용

// 2. Aggregation queries (더 적은 읽기)
import { getCountFromServer } from 'firebase/firestore'

const snapshot = await getCountFromServer(
  query(collection(db, 'posts'), where('published', '==', true))
)
console.log('Count:', snapshot.data().count)
// → 모든 문서 읽기 대신 count만 조회

// 3. Cache 활용
const { docs, fromCache } = await getDocs(
  query(collection(db, 'static-content'))
)
if (fromCache) {
  console.log('Loaded from cache - no read cost')
}
```

## 6. 다음 단계

- [BaaS Ecosystem 개요](../baas-ecosystem) - 9개 플랫폼 비교
- [Supabase 가이드](./supabase) - PostgreSQL 기반 BaaS
- [Vercel 가이드](./vercel) - Edge Platform
- [Advanced Skills](../advanced-skills) - Context7, MCP Builder
