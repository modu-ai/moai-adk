# Neon Serverless PostgreSQL 완전 가이드

## 개요

Neon은 완전히 서버리스화된 PostgreSQL 데이터베이스 플랫폼입니다. 자동 스케일링, 즉시 생성되는 브랜치, 그리고 사용한 만큼만 지불하는 가격 정책으로 현대적인 애플리케이션 개발에 최적화되어 있습니다.

**핵심 장점**:
- **Instant Branching**: Git처럼 데이터베이스 브랜치 생성 (1초 이내)
- **Autoscaling**: 사용량에 따라 자동으로 리소스 조정
- **Scale to Zero**: 미사용 시 자동으로 리소스 해제
- **Time Travel**: 과거 시점으로 데이터 복원
- **Connection Pooling**: 서버리스 환경에 최적화된 연결 관리

## 왜 Neon인가?

### 1. 개발 워크플로우 혁신

```bash
# 기존 방식: 개발 DB 설정 (수십 분 소요)
docker run -d postgres
pg_dump production > backup.sql
psql development < backup.sql

# Neon: 1초 안에 프로덕션 복사본 생성
neon branches create --parent main dev-john
# 완료! 즉시 사용 가능
```

### 2. Pattern C & D에 최적화

**Pattern C (Full-stack Monolith)**:
```typescript
// app/lib/db.ts
import { Pool, neonConfig } from '@neondatabase/serverless'
import { drizzle } from 'drizzle-orm/neon-serverless'
import ws from 'ws'

// WebSocket polyfill (Next.js Edge에서 필요)
neonConfig.webSocketConstructor = ws

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
})

export const db = drizzle(pool)

// app/api/users/route.ts
import { db } from '@/lib/db'
import { users } from '@/lib/schema'
import { eq } from 'drizzle-orm'

export async function GET(request: Request) {
  const allUsers = await db.select().from(users)
  return Response.json(allUsers)
}

export async function POST(request: Request) {
  const { email, name } = await request.json()

  const [user] = await db
    .insert(users)
    .values({ email, name })
    .returning()

  return Response.json(user, { status: 201 })
}
```

**Pattern D (Microservices)**:
```typescript
// services/auth/db.ts - Auth 서비스 전용 DB
const authDb = drizzle(new Pool({
  connectionString: process.env.AUTH_DATABASE_URL,
}))

// services/orders/db.ts - Orders 서비스 전용 DB
const ordersDb = drizzle(new Pool({
  connectionString: process.env.ORDERS_DATABASE_URL,
}))

// 각 서비스가 독립적인 Neon 브랜치 사용
// - main: 프로덕션
// - auth-dev: Auth 서비스 개발
// - orders-dev: Orders 서비스 개발
```

## 주요 기능

### 1. Database Branching

**Git처럼 데이터베이스 관리**:

```bash
# 1. Main 브랜치에서 새 브랜치 생성
neonctl branches create --name feature-users --parent main

# 2. 브랜치별 연결 문자열 확인
neonctl connection-string feature-users
# postgresql://user:pass@ep-feature-users.neon.tech/neondb

# 3. 개발 완료 후 삭제
neonctl branches delete feature-users
```

**CI/CD 통합**:
```yaml
# .github/workflows/preview.yml
name: Preview Environment

on:
  pull_request:
    branches: [main]

jobs:
  deploy-preview:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Create Neon Branch
        id: create-branch
        run: |
          BRANCH_NAME="preview-${{ github.event.pull_request.number }}"
          neonctl branches create --name $BRANCH_NAME --parent main
          echo "branch_name=$BRANCH_NAME" >> $GITHUB_OUTPUT

      - name: Deploy to Vercel
        env:
          DATABASE_URL: ${{ steps.create-branch.outputs.connection_string }}
        run: |
          vercel deploy --env DATABASE_URL=$DATABASE_URL

      - name: Comment PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `✅ Preview deployed with database branch: ${process.env.BRANCH_NAME}`
            })
```

### 2. Connection Pooling

**서버리스 환경을 위한 최적화**:

```typescript
// lib/db.ts - Pooled Connection
import { Pool } from '@neondatabase/serverless'

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  // Neon Pooler 사용 (권장)
  // postgresql://user:pass@ep-xxx.pooler.neon.tech/db
  max: 1, // Serverless에서는 1개 연결만 사용
})

export async function query<T>(text: string, params?: any[]): Promise<T[]> {
  const client = await pool.connect()
  try {
    const result = await client.query(text, params)
    return result.rows
  } finally {
    client.release()
  }
}
```

**Direct Connection (Prisma 권장)**:
```typescript
// lib/prisma.ts
import { PrismaClient } from '@prisma/client'
import { PrismaNeon } from '@prisma/adapter-neon'
import { Pool } from '@neondatabase/serverless'

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
})

const adapter = new PrismaNeon(pool)

export const prisma = new PrismaClient({
  adapter,
  log: process.env.NODE_ENV === 'development' ? ['query', 'error', 'warn'] : ['error'],
})
```

### 3. Read Replicas

**읽기 성능 최적화**:

```typescript
// lib/db.ts
import { Pool } from '@neondatabase/serverless'

// Primary (쓰기 + 읽기)
const primary = new Pool({
  connectionString: process.env.DATABASE_URL,
})

// Read Replica (읽기 전용)
const replica = new Pool({
  connectionString: process.env.DATABASE_REPLICA_URL,
})

export async function write<T>(query: string, params?: any[]): Promise<T> {
  const client = await primary.connect()
  try {
    const result = await client.query(query, params)
    return result.rows[0]
  } finally {
    client.release()
  }
}

export async function read<T>(query: string, params?: any[]): Promise<T[]> {
  const client = await replica.connect()
  try {
    const result = await client.query(query, params)
    return result.rows
  } finally {
    client.release()
  }
}

// 사용 예제
const users = await read('SELECT * FROM users WHERE active = $1', [true])
const user = await write(
  'INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *',
  ['user@example.com', 'John Doe']
)
```

### 4. Point-in-Time Recovery

**과거 시점으로 데이터 복원**:

```bash
# 24시간 전 상태로 새 브랜치 생성
neonctl branches create \
  --name recovery-20240101 \
  --parent main \
  --timestamp "2024-01-01T00:00:00Z"

# 복원된 데이터 확인
psql $DATABASE_URL_RECOVERY -c "SELECT * FROM orders WHERE deleted_at IS NOT NULL"

# 필요한 데이터만 추출하여 메인으로 복구
pg_dump $DATABASE_URL_RECOVERY --table=orders --data-only > recovery.sql
psql $DATABASE_URL < recovery.sql
```

### 5. Autoscaling

**자동 리소스 조정**:

```sql
-- 현재 리소스 사용량 확인
SELECT
  datname as database,
  numbackends as connections,
  pg_size_pretty(pg_database_size(datname)) as size
FROM pg_stat_database
WHERE datname NOT IN ('template0', 'template1', 'postgres');

-- Neon이 자동으로:
-- 1. CPU/메모리 사용량 모니터링
-- 2. 트래픽 증가 시 리소스 자동 증가
-- 3. 유휴 시 Scale to Zero (무료 플랜 5분, 유료 플랜 설정 가능)
```

## 시작하기

### 1. Neon 프로젝트 생성

```bash
# Neon CLI 설치
npm install -g neonctl

# 로그인
neonctl auth

# 프로젝트 생성
neonctl projects create --name my-app --region aws-us-east-2

# 연결 문자열 확인
neonctl connection-string main
```

### 2. ORM 통합 (Prisma)

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
  previewFeatures = ["driverAdapters"]
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        Int      @id @default(autoincrement())
  email     String   @unique
  name      String
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  posts Post[]
}

model Post {
  id        Int      @id @default(autoincrement())
  title     String
  content   String?
  published Boolean  @default(false)
  authorId  Int
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  author User @relation(fields: [authorId], references: [id])
}
```

```bash
# 마이그레이션 생성
npx prisma migrate dev --name init

# Prisma Client 생성
npx prisma generate
```

### 3. ORM 통합 (Drizzle)

```typescript
// lib/schema.ts
import { pgTable, serial, text, timestamp, boolean, integer } from 'drizzle-orm/pg-core'

export const users = pgTable('users', {
  id: serial('id').primaryKey(),
  email: text('email').notNull().unique(),
  name: text('name').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})

export const posts = pgTable('posts', {
  id: serial('id').primaryKey(),
  title: text('title').notNull(),
  content: text('content'),
  published: boolean('published').default(false).notNull(),
  authorId: integer('author_id').references(() => users.id).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
})
```

```typescript
// lib/db.ts
import { drizzle } from 'drizzle-orm/neon-http'
import { neon } from '@neondatabase/serverless'

const sql = neon(process.env.DATABASE_URL!)
export const db = drizzle(sql)
```

```bash
# 마이그레이션 생성
npx drizzle-kit generate:pg

# 마이그레이션 실행
npx drizzle-kit push:pg
```

## 사용 가이드

### 1. Next.js App Router 통합

```typescript
// app/lib/db.ts
import { Pool } from '@neondatabase/serverless'
import { PrismaNeon } from '@prisma/adapter-neon'
import { PrismaClient } from '@prisma/client'

const pool = new Pool({ connectionString: process.env.DATABASE_URL })
const adapter = new PrismaNeon(pool)
const prisma = new PrismaClient({ adapter })

export default prisma

// app/api/posts/route.ts
import prisma from '@/lib/db'

export async function GET() {
  const posts = await prisma.post.findMany({
    where: { published: true },
    include: { author: true },
    orderBy: { createdAt: 'desc' },
  })

  return Response.json(posts)
}

export const runtime = 'edge' // Edge Runtime에서도 작동!
```

### 2. 환경별 브랜치 전략

```bash
# 1. Production 브랜치 (Main)
main → postgresql://user:pass@ep-main.neon.tech/db

# 2. Staging 브랜치
neonctl branches create --name staging --parent main
staging → postgresql://user:pass@ep-staging.neon.tech/db

# 3. 개발자별 브랜치
neonctl branches create --name dev-john --parent staging
neonctl branches create --name dev-jane --parent staging

# 4. PR별 임시 브랜치 (CI/CD에서 자동 생성/삭제)
neonctl branches create --name pr-123 --parent main
```

**환경 변수 설정**:
```bash
# Production
DATABASE_URL=postgresql://user:pass@ep-main.neon.tech/db

# Staging
DATABASE_URL=postgresql://user:pass@ep-staging.neon.tech/db

# Development (로컬)
DATABASE_URL=postgresql://user:pass@ep-dev-john.neon.tech/db
```

### 3. 트랜잭션 관리

```typescript
// lib/transactions.ts
import prisma from './db'

export async function createOrderWithItems(
  userId: number,
  items: { productId: number; quantity: number }[]
) {
  return await prisma.$transaction(async (tx) => {
    // 1. 주문 생성
    const order = await tx.order.create({
      data: {
        userId,
        status: 'pending',
        total: 0,
      },
    })

    // 2. 주문 상품 생성
    let total = 0
    for (const item of items) {
      const product = await tx.product.findUnique({
        where: { id: item.productId },
      })

      if (!product || product.stock < item.quantity) {
        throw new Error(`Product ${item.productId} is out of stock`)
      }

      await tx.orderItem.create({
        data: {
          orderId: order.id,
          productId: item.productId,
          quantity: item.quantity,
          price: product.price,
        },
      })

      // 3. 재고 감소
      await tx.product.update({
        where: { id: item.productId },
        data: { stock: { decrement: item.quantity } },
      })

      total += product.price * item.quantity
    }

    // 4. 주문 총액 업데이트
    await tx.order.update({
      where: { id: order.id },
      data: { total },
    })

    return order
  })
}
```

### 4. 성능 최적화

```typescript
// lib/queries.ts
import prisma from './db'

// ❌ N+1 문제
export async function getPostsWithAuthorsSlow() {
  const posts = await prisma.post.findMany()

  for (const post of posts) {
    post.author = await prisma.user.findUnique({
      where: { id: post.authorId },
    })
  }

  return posts
}

// ✅ Eager Loading
export async function getPostsWithAuthorsFast() {
  return await prisma.post.findMany({
    include: { author: true },
  })
}

// ✅ Select 최적화
export async function getPostsSummary() {
  return await prisma.post.findMany({
    select: {
      id: true,
      title: true,
      createdAt: true,
      author: {
        select: {
          name: true,
        },
      },
    },
  })
}

// ✅ 인덱스 활용
// prisma/schema.prisma
// model Post {
//   ...
//   @@index([published, createdAt])
// }

export async function getPublishedPosts() {
  return await prisma.post.findMany({
    where: { published: true },
    orderBy: { createdAt: 'desc' },
  })
}
```

## 코드 예제

### 1. Full-text Search

```sql
-- 마이그레이션: Full-text search 인덱스 생성
CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE posts
ADD COLUMN search_vector tsvector
GENERATED ALWAYS AS (
  to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, ''))
) STORED;

CREATE INDEX idx_posts_search ON posts USING GIN (search_vector);
```

```typescript
// lib/search.ts
import { sql } from '@vercel/postgres'

export async function searchPosts(query: string) {
  const result = await sql`
    SELECT
      id,
      title,
      content,
      ts_rank(search_vector, plainto_tsquery('english', ${query})) as rank
    FROM posts
    WHERE search_vector @@ plainto_tsquery('english', ${query})
    ORDER BY rank DESC
    LIMIT 20
  `

  return result.rows
}
```

### 2. Soft Delete 패턴

```prisma
// prisma/schema.prisma
model Post {
  id        Int       @id @default(autoincrement())
  title     String
  content   String?
  deletedAt DateTime?

  @@index([deletedAt])
}
```

```typescript
// lib/soft-delete.ts
import prisma from './db'

export async function softDelete(postId: number) {
  return await prisma.post.update({
    where: { id: postId },
    data: { deletedAt: new Date() },
  })
}

export async function restore(postId: number) {
  return await prisma.post.update({
    where: { id: postId },
    data: { deletedAt: null },
  })
}

export async function hardDelete(postId: number) {
  return await prisma.post.delete({
    where: { id: postId },
  })
}

// 기본 쿼리에서 삭제된 항목 제외
export async function findActivePost(postId: number) {
  return await prisma.post.findFirst({
    where: {
      id: postId,
      deletedAt: null,
    },
  })
}
```

### 3. Pagination

```typescript
// lib/pagination.ts
import prisma from './db'

interface PaginationOptions {
  page: number
  pageSize: number
  orderBy?: any
  where?: any
}

export async function paginate<T>(
  model: any,
  options: PaginationOptions
) {
  const { page, pageSize, orderBy, where } = options
  const skip = (page - 1) * pageSize

  const [data, total] = await Promise.all([
    model.findMany({
      where,
      orderBy,
      skip,
      take: pageSize,
    }),
    model.count({ where }),
  ])

  return {
    data,
    pagination: {
      page,
      pageSize,
      total,
      totalPages: Math.ceil(total / pageSize),
      hasNext: page * pageSize < total,
      hasPrev: page > 1,
    },
  }
}

// 사용 예제
const result = await paginate(prisma.post, {
  page: 1,
  pageSize: 10,
  where: { published: true },
  orderBy: { createdAt: 'desc' },
})
```

### 4. Database Seeding

```typescript
// prisma/seed.ts
import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()

async function main() {
  // 기존 데이터 삭제
  await prisma.post.deleteMany()
  await prisma.user.deleteMany()

  // 사용자 생성
  const users = await Promise.all([
    prisma.user.create({
      data: {
        email: 'john@example.com',
        name: 'John Doe',
      },
    }),
    prisma.user.create({
      data: {
        email: 'jane@example.com',
        name: 'Jane Smith',
      },
    }),
  ])

  // 게시글 생성
  await Promise.all([
    prisma.post.create({
      data: {
        title: 'First Post',
        content: 'This is the first post',
        published: true,
        authorId: users[0].id,
      },
    }),
    prisma.post.create({
      data: {
        title: 'Second Post',
        content: 'This is the second post',
        published: true,
        authorId: users[1].id,
      },
    }),
  ])

  console.log('Seeding completed')
}

main()
  .catch((e) => {
    console.error(e)
    process.exit(1)
  })
  .finally(async () => {
    await prisma.$disconnect()
  })
```

```json
// package.json
{
  "prisma": {
    "seed": "tsx prisma/seed.ts"
  }
}
```

```bash
# Seeding 실행
npx prisma db seed
```

## Best Practices

### 1. 연결 관리

```typescript
// ✅ 연결 풀 사용
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 1, // Serverless는 1개면 충분
})

// ❌ 매번 새 연결 생성
const client = new Client({ connectionString: process.env.DATABASE_URL })
await client.connect()
```

### 2. 브랜치 정리

```bash
# 오래된 브랜치 목록
neonctl branches list --project-id $PROJECT_ID

# 사용하지 않는 브랜치 삭제
neonctl branches delete pr-123

# 자동 정리 스크립트
for branch in $(neonctl branches list --json | jq -r '.[] | select(.name | startswith("pr-")) | .name'); do
  neonctl branches delete $branch
done
```

### 3. 마이그레이션 전략

```bash
# 1. 개발 브랜치에서 마이그레이션 생성
DATABASE_URL=$DEV_DATABASE_URL npx prisma migrate dev --name add-user-role

# 2. 테스트
npm test

# 3. 프로덕션 배포 전 마이그레이션 확인
DATABASE_URL=$PROD_DATABASE_URL npx prisma migrate deploy --dry-run

# 4. 프로덕션 배포
DATABASE_URL=$PROD_DATABASE_URL npx prisma migrate deploy
```

### 4. 모니터링

```typescript
// lib/monitoring.ts
import { Pool } from '@neondatabase/serverless'

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
})

// 연결 풀 상태 모니터링
setInterval(() => {
  console.log('Pool stats:', {
    total: pool.totalCount,
    idle: pool.idleCount,
    waiting: pool.waitingCount,
  })
}, 60000)

// 쿼리 성능 로깅
prisma.$use(async (params, next) => {
  const before = Date.now()
  const result = await next(params)
  const after = Date.now()

  console.log(`Query ${params.model}.${params.action} took ${after - before}ms`)

  return result
})
```

## 문제 해결

### 1. Connection Timeout

```typescript
// Timeout 설정 증가
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  connectionTimeoutMillis: 5000, // 기본: 0 (무제한)
  idleTimeoutMillis: 30000,
})
```

### 2. Too Many Connections

```typescript
// Connection pooling 사용 (권장)
// DATABASE_URL을 pooler URL로 변경
// postgresql://user:pass@ep-xxx.pooler.neon.tech/db

// 또는 연결 수 제한
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 1, // Serverless 환경에서는 1개만 사용
})
```

### 3. Slow Queries

```sql
-- 느린 쿼리 확인
SELECT
  query,
  calls,
  mean_exec_time,
  max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- 인덱스 추가
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
```

## 다음 단계

- [Vercel 가이드](/ko/skills/baas/vercel) - Neon + Vercel 통합
- [Railway 가이드](/ko/skills/baas/railway) - 전체 스택 플랫폼
- [Pattern C: Full-stack Monolith](/ko/skills/patterns/pattern-c) - 아키텍처 가이드
- [Pattern D: Microservices](/ko/skills/patterns/pattern-d) - 마이크로서비스 DB 전략

## 참고 자료

- [Neon 공식 문서](https://neon.tech/docs)
- [Neon CLI 가이드](https://neon.tech/docs/reference/cli-reference)
- [Prisma + Neon 가이드](https://www.prisma.io/docs/guides/database/neon)
- [Drizzle + Neon 가이드](https://orm.drizzle.team/docs/get-started-postgresql#neon)
