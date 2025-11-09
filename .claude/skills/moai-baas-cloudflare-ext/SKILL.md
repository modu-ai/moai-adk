# Skill: moai-baas-cloudflare-ext

## Metadata

```yaml
skill_id: moai-baas-cloudflare-ext
skill_name: Cloudflare Edge-First Architecture & Workers
version: 1.0.0
created_date: 2025-11-09
language: english
triggers:
  - keywords: ["Cloudflare", "Workers", "D1", "Pages", "Edge", "Wrangler"]
  - contexts: ["cloudflare-detected", "pattern-g", "edge-performance"]
agents:
  - backend-expert
  - devops-expert
  - database-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://developers.cloudflare.com/workers/"
    topic: "Cloudflare Workers Runtime"
  - url: "https://developers.cloudflare.com/d1/"
    topic: "D1 Database (SQLite)"
  - url: "https://developers.cloudflare.com/pages/"
    topic: "Cloudflare Pages Hosting"
  - url: "https://developers.cloudflare.com/analytics/"
    topic: "Analytics Engine"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Cloudflare Edge-First Philosophy (150 words)

**Cloudflare** delivers applications at the edge, eliminating latency through global distribution.

**Core Concept: Edge Computing**:
```
Traditional (Centralized):
  User → Network → Central Datacenter → Response
  └─ Latency: 100-500ms depending on location

Cloudflare Edge (Distributed):
  User → Closest Cloudflare Edge (250+ locations) → Response
  └─ Latency: <50ms globally, cached responses <10ms
```

**Cloudflare Stack**:
```
┌────────────────────────────────────────────┐
│ Cloudflare (Edge-Native Platform)          │
├────────────────────────────────────────────┤
│                                             │
│ 1. Workers (Serverless Functions)          │
│    └─ Run code at edge (80ms cold start)   │
│                                             │
│ 2. Pages (Static + Dynamic Hosting)        │
│    └─ Deploy front-end with edge functions │
│                                             │
│ 3. D1 (Distributed SQLite Database)        │
│    └─ SQL database replicated globally      │
│                                             │
│ 4. KV Store (Key-Value Cache)              │
│    └─ Sub-millisecond response time        │
│                                             │
│ 5. Analytics Engine (Observability)        │
│    └─ Query logs at edge                   │
│                                             │
│ 6. Queues (Message Queuing)                │
│    └─ Publish-subscribe at edge            │
└────────────────────────────────────────────┘
```

**Why Cloudflare**:
- ✅ Fastest globally distributed platform
- ✅ SQLite database (familiar, portable)
- ✅ Low cold start (~80ms)
- ✅ Generous free tier
- ⚠️ Learning curve (different paradigm)
- ⚠️ Limited resources per request (10s CPU time)

---

### 2. Cloudflare Workers Runtime (250 words)

**Workers** are serverless functions that execute on Cloudflare's edge network.

**Worker Basics**:

```typescript
// wrangler.toml - Worker Configuration
name = "my-app"
main = "src/index.ts"
compatibility_date = "2025-01-01"

[env.production]
routes = [
  { pattern = "example.com/*", zone_name = "example.com" }
]

[[d1_databases]]
binding = "DB"
database_name = "mydb"
database_id = "xxx"

[env.development]
d1_databases = [{ binding = "DB", database_name = "mydb-dev" }]
```

**HTTP Worker Example**:

```typescript
// src/index.ts
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    // Route handling
    if (url.pathname === "/api/hello") {
      return new Response(
        JSON.stringify({ message: "Hello from edge!" }),
        { status: 200, headers: { "Content-Type": "application/json" } }
      );
    }

    // Database query
    if (url.pathname === "/api/users") {
      const { results } = await env.DB.prepare(
        "SELECT id, email FROM users LIMIT 10"
      ).all();

      return Response.json(results);
    }

    // File serving
    return new Response("Not found", { status: 404 });
  },
};

export interface Env {
  DB: D1Database;
}
```

**Request/Response Handling**:

```typescript
// Parse JSON body
const body = await request.json();

// Parse URL parameters
const { searchParams } = new URL(request.url);
const limit = searchParams.get("limit") || "10";

// Set response headers
const headers = new Headers({
  "Content-Type": "application/json",
  "Cache-Control": "public, max-age=3600",
});

return new Response(JSON.stringify(data), { status: 200, headers });
```

**Environment Variables**:

```toml
# .env.production
DATABASE_URL = "your-d1-database-id"
API_KEY = "your-secret-key"
```

**Deploying Workers**:

```bash
# Install Wrangler (Cloudflare CLI)
npm install -D wrangler

# Authenticate
wrangler auth

# Deploy
wrangler deploy

# Local testing
wrangler dev
```

---

### 3. D1 Database & SQL Operations (250 words)

**D1** is Cloudflare's distributed SQLite database with global replication.

**Schema Design**:

```sql
-- Schema for D1
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  content TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
```

**Database Operations in Workers**:

```typescript
// Query data
export const getUser = async (
  db: D1Database,
  userId: number
): Promise<User | null> => {
  const { results } = await db
    .prepare("SELECT * FROM users WHERE id = ?")
    .bind(userId)
    .first();

  return results || null;
};

// Insert data
export const createPost = async (
  db: D1Database,
  userId: number,
  title: string,
  content: string
): Promise<number> => {
  const result = await db
    .prepare(
      "INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)"
    )
    .bind(userId, title, content)
    .run();

  return result.meta.last_row_id;
};

// Update data
export const updatePost = async (
  db: D1Database,
  postId: number,
  title: string
): Promise<void> => {
  await db
    .prepare("UPDATE posts SET title = ? WHERE id = ?")
    .bind(title, postId)
    .run();
};

// Batch operations (transactions)
export const migrateData = async (db: D1Database): Promise<void> => {
  const batch = [
    db.prepare("INSERT INTO users (email, name) VALUES (?, ?)").bind(
      "alice@example.com",
      "Alice"
    ),
    db.prepare("INSERT INTO users (email, name) VALUES (?, ?)").bind(
      "bob@example.com",
      "Bob"
    ),
  ];

  await db.batch(batch);
};

// Complex queries
export const getUserWithPosts = async (
  db: D1Database,
  userId: number
): Promise<any> => {
  const { results } = await db
    .prepare(
      `
    SELECT u.*, COUNT(p.id) as post_count
    FROM users u
    LEFT JOIN posts p ON u.id = p.user_id
    WHERE u.id = ?
    GROUP BY u.id
  `
    )
    .bind(userId)
    .first();

  return results;
};
```

**Migration Strategy**:

```bash
# Initialize migrations
wrangler d1 create my-database

# Create migration
wrangler d1 migrations create my-database add_users_table

# Apply migrations
wrangler d1 migrations apply my-database --remote

# Backup/restore
wrangler d1 export my-database --output backup.sql
```

---

### 4. Pages & Full-Stack Deployment (200 words)

**Cloudflare Pages** hosts full-stack applications with Workers integration.

**Pages Project Structure**:

```
my-app/
├─ src/
│  ├─ index.html
│  ├─ pages/           # Static pages
│  └─ _worker.ts       # Pages Functions
├─ functions/
│  ├─ api/
│  │  ├─ users.ts      # GET /api/users
│  │  └─ posts.ts      # GET /api/posts
│  └─ middleware/
│     └─ auth.ts       # Authentication
├─ wrangler.toml
└─ package.json
```

**Pages Functions (Routing)**:

```typescript
// functions/api/users.ts
export const onRequest = async ({ env, request }) => {
  if (request.method === "GET") {
    const users = await env.DB.prepare("SELECT * FROM users").all();
    return Response.json(users.results);
  }

  if (request.method === "POST") {
    const { email, name } = await request.json();
    const result = await env.DB.prepare(
      "INSERT INTO users (email, name) VALUES (?, ?)"
    )
      .bind(email, name)
      .run();

    return Response.json({ id: result.meta.last_row_id });
  }

  return new Response("Method not allowed", { status: 405 });
};
```

**Middleware for Authentication**:

```typescript
// functions/middleware/auth.ts
export const onRequest = async (context) => {
  const token = context.request.headers.get("Authorization");

  if (!token) {
    return new Response("Unauthorized", { status: 401 });
  }

  // Verify token, attach to context
  context.data.userId = verifyToken(token);

  return context.next();
};
```

**Deployment**:

```bash
# Link Pages project to Git
wrangler pages project create my-app

# Deploy from Git (automatic)
# Or manual deploy
wrangler pages deploy dist/
```

---

### 5. Performance Optimization & KV Cache (150 words)

**KV Store** provides sub-millisecond caching at the edge.

**Caching Pattern**:

```typescript
// Cache user data for 5 minutes
export const getUser = async (
  request: Request,
  env: Env,
  userId: number
): Promise<User> => {
  const cacheKey = `user-${userId}`;

  // Try KV cache first
  const cached = await env.KV.get(cacheKey);
  if (cached) return JSON.parse(cached);

  // Fetch from database
  const user = await env.DB.prepare("SELECT * FROM users WHERE id = ?")
    .bind(userId)
    .first();

  // Store in KV for 5 minutes
  await env.KV.put(cacheKey, JSON.stringify(user), {
    expirationTtl: 300,
  });

  return user;
};
```

**Performance Best Practices**:
- ✅ Use KV for frequently accessed data
- ✅ Set appropriate TTL (time-to-live)
- ✅ Batch database queries
- ✅ Use response headers for HTTP caching
- ⚠️ D1 has cold start latency (50-100ms)
- ⚠️ 10-second CPU time limit per request

---

### 6. Common Issues & Solutions (50 words)

| Issue | Solution |
|-------|----------|
| **D1 query slow** | Add indexes, use EXPLAIN QUERY PLAN |
| **Workers timeout** | Break work into smaller requests |
| **KV stale data** | Set appropriate TTL |
| **Large response** | Pagination or streaming |

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-cloudflare-ext")
# Load when Pattern G (Cloudflare Edge-first) detected
```

### Context7 Integration
When Cloudflare platform detected:
- Workers runtime & request handling
- D1 database design & SQL operations
- Pages deployment & Functions routing
- Performance optimization with KV cache

---

## 📚 Reference Materials

- [Cloudflare Workers Docs](https://developers.cloudflare.com/workers/)
- [D1 Database Guide](https://developers.cloudflare.com/d1/)
- [Pages Deployment](https://developers.cloudflare.com/pages/)
- [Analytics Engine](https://developers.cloudflare.com/analytics/)

---

## ✅ Validation Checklist

- [x] Edge-first philosophy & architecture
- [x] Workers runtime & HTTP handling
- [x] D1 database & SQL operations
- [x] Pages deployment & Functions routing
- [x] Performance optimization with KV cache
- [x] Common issues & troubleshooting
- [x] 1000-word target
- [x] English language (policy compliant)
