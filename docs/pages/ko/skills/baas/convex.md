# Convex Real-time Backend 완전 가이드

## 개요

Convex는 실시간 동기화에 특화된 백엔드 플랫폼입니다. TypeScript로 서버 함수를 작성하면 자동으로 클라이언트 훅이 생성되며, 모든 데이터는 실시간으로 동기화됩니다.

**핵심 장점**:
- **Real-time by Default**: 모든 쿼리가 자동으로 실시간 구독
- **TypeScript Native**: 완벽한 타입 안전성
- **Zero Config**: 데이터베이스, 서버, 배포 자동 설정
- **Reactive Queries**: 데이터 변경 시 자동 리렌더링
- **Built-in File Storage**: 파일 업로드/다운로드 내장

## 왜 Convex인가?

### 1. Real-time 동기화의 단순함 (Pattern F)

```typescript
// convex/messages.ts - 서버 함수
import { query, mutation } from './_generated/server'
import { v } from 'convex/values'

// 메시지 목록 조회 (실시간)
export const list = query({
  args: {},
  handler: async (ctx) => {
    return await ctx.db.query('messages').order('desc').take(100)
  },
})

// 메시지 생성
export const send = mutation({
  args: {
    author: v.string(),
    body: v.string(),
  },
  handler: async (ctx, args) => {
    await ctx.db.insert('messages', {
      author: args.author,
      body: args.body,
      timestamp: Date.now(),
    })
  },
})

// app/chat/page.tsx - 클라이언트
'use client'

import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'

export default function Chat() {
  // 실시간 구독 - 데이터 변경 시 자동 업데이트
  const messages = useQuery(api.messages.list)

  // Mutation
  const sendMessage = useMutation(api.messages.send)

  const handleSend = async (body: string) => {
    await sendMessage({ author: 'John', body })
  }

  return (
    <div>
      {messages?.map((msg) => (
        <div key={msg._id}>
          <strong>{msg.author}</strong>: {msg.body}
        </div>
      ))}
      {/* 입력 폼 */}
    </div>
  )
}
```

### 2. TypeScript 완전 통합

```typescript
// convex/schema.ts - 스키마 정의
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  messages: defineTable({
    author: v.string(),
    body: v.string(),
    timestamp: v.number(),
  }).index('by_timestamp', ['timestamp']),

  users: defineTable({
    name: v.string(),
    email: v.string(),
    avatar: v.optional(v.string()),
  }).index('by_email', ['email']),

  rooms: defineTable({
    name: v.string(),
    ownerId: v.id('users'),
    members: v.array(v.id('users')),
  }),
})

// TypeScript가 자동으로 타입 추론
const message = await ctx.db.query('messages').first()
// message.author - string
// message.body - string
// message.timestamp - number
```

## 주요 기능

### 1. Queries (실시간 조회)

```typescript
// convex/posts.ts
import { query } from './_generated/server'
import { v } from 'convex/values'

export const get = query({
  args: { postId: v.id('posts') },
  handler: async (ctx, args) => {
    const post = await ctx.db.get(args.postId)

    if (!post) {
      throw new Error('Post not found')
    }

    // 관계 데이터 로드
    const author = await ctx.db.get(post.authorId)

    return {
      ...post,
      author,
    }
  },
})

export const list = query({
  args: {
    limit: v.optional(v.number()),
    cursor: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const posts = await ctx.db
      .query('posts')
      .order('desc')
      .paginate(args)

    return posts
  },
})

// 클라이언트
const post = useQuery(api.posts.get, { postId: '123' })
const { results, continueCursor } = useQuery(api.posts.list, { limit: 20 })
```

### 2. Mutations (데이터 수정)

```typescript
// convex/posts.ts
import { mutation } from './_generated/server'
import { v } from 'convex/values'

export const create = mutation({
  args: {
    title: v.string(),
    content: v.string(),
    authorId: v.id('users'),
  },
  handler: async (ctx, args) => {
    // 트랜잭션으로 자동 처리
    const postId = await ctx.db.insert('posts', {
      title: args.title,
      content: args.content,
      authorId: args.authorId,
      createdAt: Date.now(),
      likes: 0,
    })

    // 알림 생성
    await ctx.db.insert('notifications', {
      type: 'new_post',
      postId,
      timestamp: Date.now(),
    })

    return postId
  },
})

export const update = mutation({
  args: {
    postId: v.id('posts'),
    title: v.optional(v.string()),
    content: v.optional(v.string()),
  },
  handler: async (ctx, args) => {
    const { postId, ...updates } = args

    await ctx.db.patch(postId, {
      ...updates,
      updatedAt: Date.now(),
    })
  },
})

export const deletePost = mutation({
  args: { postId: v.id('posts') },
  handler: async (ctx, args) => {
    // Soft delete
    await ctx.db.patch(args.postId, {
      deletedAt: Date.now(),
    })
  },
})

// 클라이언트
const createPost = useMutation(api.posts.create)
const updatePost = useMutation(api.posts.update)
const deletePost = useMutation(api.posts.deletePost)

await createPost({ title: 'Hello', content: 'World', authorId: userId })
```

### 3. Actions (외부 API 호출)

```typescript
// convex/ai.ts
import { action } from './_generated/server'
import { v } from 'convex/values'
import { api } from './_generated/api'

export const generateSummary = action({
  args: {
    postId: v.id('posts'),
  },
  handler: async (ctx, args) => {
    // 1. 포스트 가져오기
    const post = await ctx.runQuery(api.posts.get, { postId: args.postId })

    // 2. OpenAI API 호출
    const response = await fetch('https://api.openai.com/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${process.env.OPENAI_API_KEY}`,
      },
      body: JSON.stringify({
        model: 'gpt-4',
        messages: [
          {
            role: 'user',
            content: `Summarize this post: ${post.content}`,
          },
        ],
      }),
    })

    const data = await response.json()
    const summary = data.choices[0].message.content

    // 3. 데이터베이스 업데이트
    await ctx.runMutation(api.posts.update, {
      postId: args.postId,
      summary,
    })

    return summary
  },
})

// 클라이언트
const generateSummary = useAction(api.ai.generateSummary)
await generateSummary({ postId: '123' })
```

### 4. Scheduled Functions (Cron)

```typescript
// convex/cron.ts
import { cronJobs } from 'convex/server'
import { internal } from './_generated/api'

const crons = cronJobs()

// 매시간 실행
crons.hourly(
  'cleanup expired sessions',
  { hourUTC: 0 }, // 매시 0분
  internal.sessions.cleanup
)

// 매일 실행
crons.daily(
  'send daily digest',
  { hourUTC: 9, minuteUTC: 0 }, // 매일 오전 9시
  internal.emails.sendDailyDigest
)

// 커스텀 스케줄
crons.cron(
  'backup database',
  '0 2 * * *', // 매일 오전 2시
  internal.backup.run
)

export default crons

// convex/sessions.ts
import { internalMutation } from './_generated/server'

export const cleanup = internalMutation({
  handler: async (ctx) => {
    const expired = await ctx.db
      .query('sessions')
      .filter((q) => q.lt(q.field('expiresAt'), Date.now()))
      .collect()

    for (const session of expired) {
      await ctx.db.delete(session._id)
    }

    console.log(`Cleaned up ${expired.length} expired sessions`)
  },
})
```

### 5. File Storage

```typescript
// convex/files.ts
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const generateUploadUrl = mutation({
  handler: async (ctx) => {
    // 업로드 URL 생성 (10분 유효)
    return await ctx.storage.generateUploadUrl()
  },
})

export const saveFile = mutation({
  args: {
    storageId: v.id('_storage'),
    name: v.string(),
    type: v.string(),
  },
  handler: async (ctx, args) => {
    const fileId = await ctx.db.insert('files', {
      storageId: args.storageId,
      name: args.name,
      type: args.type,
      uploadedAt: Date.now(),
    })

    return fileId
  },
})

export const getFileUrl = query({
  args: { fileId: v.id('files') },
  handler: async (ctx, args) => {
    const file = await ctx.db.get(args.fileId)

    if (!file) return null

    // 다운로드 URL 생성
    const url = await ctx.storage.getUrl(file.storageId)

    return url
  },
})

// 클라이언트 - 파일 업로드
'use client'

import { useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'

export function FileUpload() {
  const generateUploadUrl = useMutation(api.files.generateUploadUrl)
  const saveFile = useMutation(api.files.saveFile)

  const handleUpload = async (file: File) => {
    // 1. 업로드 URL 생성
    const uploadUrl = await generateUploadUrl()

    // 2. 파일 업로드
    const response = await fetch(uploadUrl, {
      method: 'POST',
      body: file,
    })

    const { storageId } = await response.json()

    // 3. 메타데이터 저장
    await saveFile({
      storageId,
      name: file.name,
      type: file.type,
    })
  }

  return <input type="file" onChange={(e) => handleUpload(e.target.files[0])} />
}
```

### 6. Authentication

```typescript
// convex/auth.ts
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const signUp = mutation({
  args: {
    email: v.string(),
    password: v.string(),
    name: v.string(),
  },
  handler: async (ctx, args) => {
    // 기존 사용자 확인
    const existing = await ctx.db
      .query('users')
      .withIndex('by_email', (q) => q.eq('email', args.email))
      .first()

    if (existing) {
      throw new Error('User already exists')
    }

    // 비밀번호 해시 (실제로는 bcrypt 등 사용)
    const passwordHash = await hashPassword(args.password)

    // 사용자 생성
    const userId = await ctx.db.insert('users', {
      email: args.email,
      passwordHash,
      name: args.name,
      createdAt: Date.now(),
    })

    return userId
  },
})

export const getCurrentUser = query({
  args: {},
  handler: async (ctx) => {
    // Convex Auth integration
    const identity = await ctx.auth.getUserIdentity()

    if (!identity) {
      return null
    }

    // 사용자 정보 조회
    const user = await ctx.db
      .query('users')
      .withIndex('by_token', (q) => q.eq('tokenIdentifier', identity.tokenIdentifier))
      .unique()

    return user
  },
})

// app/providers.tsx
'use client'

import { ConvexProvider, ConvexReactClient } from 'convex/react'
import { ClerkProvider, useAuth } from '@clerk/nextjs'
import { ConvexProviderWithClerk } from 'convex/react-clerk'

const convex = new ConvexReactClient(process.env.NEXT_PUBLIC_CONVEX_URL!)

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider>
      <ConvexProviderWithClerk client={convex} useAuth={useAuth}>
        {children}
      </ConvexProviderWithClerk>
    </ClerkProvider>
  )
}
```

## 시작하기

### 1. 설치 및 초기화

```bash
# Convex CLI 설치
npm install -g convex

# 프로젝트 초기화
npx create-next-app@latest my-convex-app
cd my-convex-app

# Convex 설치
npm install convex

# Convex 초기화
npx convex dev
```

### 2. 프로젝트 구조

```
my-convex-app/
├── convex/
│   ├── _generated/         # 자동 생성된 타입
│   ├── schema.ts           # 데이터베이스 스키마
│   ├── messages.ts         # 서버 함수
│   └── auth.ts
├── app/
│   ├── layout.tsx
│   └── page.tsx
└── convex.json             # Convex 설정
```

### 3. 스키마 정의

```typescript
// convex/schema.ts
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  users: defineTable({
    name: v.string(),
    email: v.string(),
    tokenIdentifier: v.string(),
  })
    .index('by_email', ['email'])
    .index('by_token', ['tokenIdentifier']),

  posts: defineTable({
    title: v.string(),
    content: v.string(),
    authorId: v.id('users'),
    likes: v.number(),
    createdAt: v.number(),
  }).index('by_author', ['authorId']),

  comments: defineTable({
    postId: v.id('posts'),
    authorId: v.id('users'),
    body: v.string(),
    createdAt: v.number(),
  }).index('by_post', ['postId']),
})
```

### 4. Provider 설정

```typescript
// app/providers.tsx
'use client'

import { ConvexProvider, ConvexReactClient } from 'convex/react'

const convex = new ConvexReactClient(process.env.NEXT_PUBLIC_CONVEX_URL!)

export function Providers({ children }: { children: React.ReactNode }) {
  return <ConvexProvider client={convex}>{children}</ConvexProvider>
}

// app/layout.tsx
import { Providers } from './providers'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ko">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
```

## 사용 가이드

### Pattern F: Real-time Collaboration

```typescript
// convex/documents.ts - 실시간 문서 편집
import { mutation, query } from './_generated/server'
import { v } from 'convex/values'

export const get = query({
  args: { documentId: v.id('documents') },
  handler: async (ctx, args) => {
    const doc = await ctx.db.get(args.documentId)

    // 현재 편집 중인 사용자 목록
    const activeUsers = await ctx.db
      .query('presence')
      .withIndex('by_document', (q) => q.eq('documentId', args.documentId))
      .filter((q) => q.gt(q.field('lastSeen'), Date.now() - 30000)) // 30초 이내
      .collect()

    return {
      ...doc,
      activeUsers,
    }
  },
})

export const updateContent = mutation({
  args: {
    documentId: v.id('documents'),
    content: v.string(),
    userId: v.id('users'),
  },
  handler: async (ctx, args) => {
    // Optimistic locking
    const doc = await ctx.db.get(args.documentId)

    await ctx.db.patch(args.documentId, {
      content: args.content,
      lastEditedBy: args.userId,
      lastEditedAt: Date.now(),
      version: (doc?.version || 0) + 1,
    })

    // 편집 히스토리 저장
    await ctx.db.insert('edits', {
      documentId: args.documentId,
      userId: args.userId,
      content: args.content,
      timestamp: Date.now(),
    })
  },
})

export const updatePresence = mutation({
  args: {
    documentId: v.id('documents'),
    userId: v.id('users'),
    cursor: v.optional(v.object({ x: v.number(), y: v.number() })),
  },
  handler: async (ctx, args) => {
    // Presence 업데이트 (현재 위치, 커서 등)
    const existing = await ctx.db
      .query('presence')
      .withIndex('by_user_document', (q) =>
        q.eq('userId', args.userId).eq('documentId', args.documentId)
      )
      .first()

    if (existing) {
      await ctx.db.patch(existing._id, {
        cursor: args.cursor,
        lastSeen: Date.now(),
      })
    } else {
      await ctx.db.insert('presence', {
        documentId: args.documentId,
        userId: args.userId,
        cursor: args.cursor,
        lastSeen: Date.now(),
      })
    }
  },
})

// 클라이언트
'use client'

import { useQuery, useMutation } from 'convex/react'
import { api } from '@/convex/_generated/api'
import { useEffect } from 'react'

export function CollaborativeEditor({ documentId }: { documentId: string }) {
  const document = useQuery(api.documents.get, { documentId })
  const updateContent = useMutation(api.documents.updateContent)
  const updatePresence = useMutation(api.documents.updatePresence)

  // Presence 업데이트 (5초마다)
  useEffect(() => {
    const interval = setInterval(() => {
      updatePresence({ documentId, userId: 'current-user-id' })
    }, 5000)

    return () => clearInterval(interval)
  }, [documentId])

  return (
    <div>
      <div>
        {document?.activeUsers?.map((user) => (
          <div key={user._id}>{user.name} is editing</div>
        ))}
      </div>

      <textarea
        value={document?.content}
        onChange={(e) => {
          updateContent({
            documentId,
            content: e.target.value,
            userId: 'current-user-id',
          })
        }}
      />
    </div>
  )
}
```

## 코드 예제

### 1. Pagination

```typescript
// convex/posts.ts
export const list = query({
  args: {
    paginationOpts: paginationOptsValidator,
  },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('posts')
      .order('desc')
      .paginate(args.paginationOpts)
  },
})

// 클라이언트
const { results, status, loadMore } = usePaginatedQuery(
  api.posts.list,
  {},
  { initialNumItems: 20 }
)
```

### 2. Full-text Search

```typescript
// convex/search.ts
export const searchPosts = query({
  args: {
    query: v.string(),
  },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('posts')
      .withSearchIndex('search_title_content', (q) => q.search('title', args.query))
      .collect()
  },
})
```

### 3. Optimistic Updates

```typescript
// 클라이언트
const likePost = useMutation(api.posts.like).withOptimisticUpdate((localStore, args) => {
  // 낙관적 업데이트 - 서버 응답 전에 UI 즉시 업데이트
  const post = localStore.getQuery(api.posts.get, { postId: args.postId })
  if (post) {
    localStore.setQuery(api.posts.get, { postId: args.postId }, {
      ...post,
      likes: post.likes + 1,
    })
  }
})
```

## Best Practices

### 1. 인덱스 최적화

```typescript
// 자주 조회하는 필드에 인덱스 생성
export default defineSchema({
  posts: defineTable({
    authorId: v.id('users'),
    createdAt: v.number(),
    status: v.string(),
  })
    .index('by_author', ['authorId'])
    .index('by_status_created', ['status', 'createdAt']),
})
```

### 2. 에러 처리

```typescript
// 클라이언트
const createPost = useMutation(api.posts.create)

try {
  await createPost({ title, content })
} catch (error) {
  console.error('Failed to create post:', error)
  // 사용자에게 에러 메시지 표시
}
```

### 3. TypeScript 타입 활용

```typescript
// convex/_generated/dataModel.d.ts 자동 생성
import { Doc, Id } from './_generated/dataModel'

// 타입 안전한 함수
function processPost(post: Doc<'posts'>) {
  // TypeScript가 post의 모든 필드 타입 알고 있음
}
```

## 문제 해결

### 1. Cold Start 최적화

```typescript
// 자주 사용하는 데이터 미리 로드
export const preload = query({
  handler: async (ctx) => {
    await Promise.all([
      ctx.db.query('users').collect(),
      ctx.db.query('posts').take(100),
    ])
  },
})
```

### 2. Large Payloads

```typescript
// 페이지네이션 사용
export const list = query({
  args: { paginationOpts: paginationOptsValidator },
  handler: async (ctx, args) => {
    return await ctx.db.query('posts').paginate(args.paginationOpts)
  },
})
```

## 다음 단계

- [Vercel 가이드](/ko/skills/baas/vercel) - 배포 플랫폼
- [Pattern F: Real-time Backend](/ko/skills/patterns/pattern-f) - 아키텍처 가이드
- [BaaS 개요](/ko/skills/baas) - 플랫폼 비교

## 참고 자료

- [Convex 공식 문서](https://docs.convex.dev/)
- [Next.js 통합 가이드](https://docs.convex.dev/quickstart/nextjs)
- [Convex Discord](https://discord.com/invite/convex)
