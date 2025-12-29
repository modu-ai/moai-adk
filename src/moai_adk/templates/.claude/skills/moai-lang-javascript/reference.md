# JavaScript Development Reference

## ES2024/ES2025 Complete Reference

### ES2024 Feature Matrix

| Feature | Description | Use Case |
|---------|-------------|----------|
| Set Methods | intersection, union, difference, etc. | Collection operations |
| Promise.withResolvers | External resolve/reject access | Deferred promises |
| Immutable Arrays | toSorted, toReversed, toSpliced, with | Functional programming |
| Object.groupBy | Group array items by key | Data categorization |
| Unicode String Methods | isWellFormed, toWellFormed | Unicode validation |
| ArrayBuffer Resizing | resize, transfer methods | Memory management |

### ES2025 Feature Matrix

| Feature | Description | Use Case |
|---------|-------------|----------|
| Import Attributes | with { type: 'json' } | JSON/CSS modules |
| RegExp.escape | Escape regex special chars | Safe regex patterns |
| Iterator Helpers | map, filter, take on iterators | Lazy iteration |
| Float16Array | 16-bit floating point arrays | ML/Graphics |
| Duplicate Named Capture Groups | Same name in regex alternation | Pattern matching |

### Complete Set Operations

```javascript
const setA = new Set([1, 2, 3, 4, 5]);
const setB = new Set([4, 5, 6, 7, 8]);

// Union - all elements from both sets
const union = setA.union(setB);
// Set {1, 2, 3, 4, 5, 6, 7, 8}

// Intersection - elements in both sets
const intersection = setA.intersection(setB);
// Set {4, 5}

// Difference - elements in A but not in B
const difference = setA.difference(setB);
// Set {1, 2, 3}

// Symmetric Difference - elements in either but not both
const symmetricDiff = setA.symmetricDifference(setB);
// Set {1, 2, 3, 6, 7, 8}

// Subset check - all elements of A are in B
setA.isSubsetOf(setB); // false
new Set([4, 5]).isSubsetOf(setB); // true

// Superset check - A contains all elements of B
setA.isSupersetOf(new Set([1, 2])); // true

// Disjoint check - no common elements
setA.isDisjointFrom(new Set([10, 11])); // true
```

### Iterator Helpers (ES2025)

```javascript
function* fibonacci() {
  let a = 0, b = 1;
  while (true) {
    yield a;
    [a, b] = [b, a + b];
  }
}

// Take first 10 Fibonacci numbers
const first10 = fibonacci().take(10).toArray();
// [0, 1, 1, 2, 3, 5, 8, 13, 21, 34]

// Filter and map
const evenFib = fibonacci()
  .filter(n => n % 2 === 0)
  .map(n => n * 2)
  .take(5)
  .toArray();
// [0, 4, 16, 68, 288]

// Reduce with iterator
const sum = fibonacci()
  .take(10)
  .reduce((acc, n) => acc + n, 0);
// 88

// forEach on iterator
fibonacci()
  .take(5)
  .forEach(n => console.log(n));

// Find on iterator
const firstOver100 = fibonacci().find(n => n > 100);
// 144

// Some and every
fibonacci().take(10).some(n => n > 10); // true
fibonacci().take(5).every(n => n < 10); // true
```

---

## Node.js Runtime Reference

### Node.js Version Comparison

| Feature | Node.js 20 LTS | Node.js 22 LTS |
|---------|----------------|----------------|
| ES Modules | Full support | Full support |
| Fetch API | Stable | Stable |
| WebSocket | Experimental | Stable (default) |
| Watch Mode | Experimental | Stable |
| TypeScript | Via loaders | Native (strip types) |
| Permission Model | Experimental | Stable |
| Test Runner | Stable | Enhanced |
| Startup Time | Baseline | 30% faster |

### Node.js Built-in Test Runner

```javascript
// test/user.test.js
import { test, describe, before, after, mock } from 'node:test';
import assert from 'node:assert';
import { createUser, getUser } from '../src/user.js';

describe('User Service', () => {
  let mockDb;

  before(() => {
    mockDb = mock.fn(() => ({ id: 1, name: 'Test' }));
  });

  after(() => {
    mock.reset();
  });

  test('creates user successfully', async (t) => {
    const user = await createUser({ name: 'John', email: 'john@test.com' });
    assert.ok(user.id);
    assert.strictEqual(user.name, 'John');
  });

  test('throws on duplicate email', async (t) => {
    await assert.rejects(
      async () => createUser({ name: 'Jane', email: 'existing@test.com' }),
      { code: 'DUPLICATE_EMAIL' }
    );
  });

  test('skipped test', { skip: true }, () => {
    // This test will be skipped
  });

  test('todo test', { todo: 'implement later' }, () => {
    // This test is marked as todo
  });
});
```

Run tests:
```bash
# Run all tests
node --test

# Run specific file
node --test test/user.test.js

# With coverage
node --test --experimental-test-coverage

# Watch mode
node --test --watch

# Parallel execution
node --test --test-concurrency=4
```

### Module System Deep Dive

Package.json Configuration:
```json
{
  "name": "my-package",
  "version": "1.0.0",
  "type": "module",
  "main": "./dist/index.cjs",
  "module": "./dist/index.js",
  "exports": {
    ".": {
      "import": "./dist/index.js",
      "require": "./dist/index.cjs",
      "types": "./dist/index.d.ts"
    },
    "./utils": {
      "import": "./dist/utils.js",
      "require": "./dist/utils.cjs"
    }
  },
  "engines": {
    "node": ">=20.0.0"
  }
}
```

ESM/CommonJS Interoperability:
```javascript
// ESM importing CommonJS
import cjsModule from 'commonjs-package';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const cjsPackage = require('commonjs-only-package');

// Get __dirname and __filename in ESM
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Dynamic import (works in both)
const module = await import('./dynamic-module.js');
```

---

## Package Manager Comparison

| Feature | npm | yarn | pnpm | bun |
|---------|-----|------|------|-----|
| Speed | Baseline | Faster | Fastest Node | Fastest overall |
| Disk Usage | High | High | Low (symlinks) | Low |
| Workspaces | Yes | Yes | Yes | Yes |
| Lockfile | package-lock.json | yarn.lock | pnpm-lock.yaml | bun.lockb |
| Plug'n'Play | No | Yes | No | No |
| Node.js Only | Yes | Yes | Yes | No (own runtime) |

### pnpm Commands

```bash
# Initialize
pnpm init

# Install dependencies
pnpm install
pnpm add express
pnpm add -D vitest

# Workspaces
pnpm -r install      # Install all workspaces
pnpm --filter=api test  # Run in specific workspace

# Performance
pnpm store prune    # Clean unused packages
pnpm dedupe        # Deduplicate dependencies
```

### Bun Commands

```bash
# Initialize
bun init

# Install (incredibly fast)
bun install
bun add express
bun add -d vitest

# Run scripts
bun run dev
bun run test

# Execute files directly
bun run server.js
bun run app.ts  # Native TypeScript support

# Built-in bundler
bun build ./src/index.ts --outdir=./dist
```

---

## Framework Reference

### Express Middleware Patterns

```javascript
import express from 'express';
import helmet from 'helmet';
import compression from 'compression';
import rateLimit from 'express-rate-limit';

const app = express();

// Security middleware
app.use(helmet());
app.use(compression());

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  standardHeaders: true,
  legacyHeaders: false,
});
app.use('/api/', limiter);

// Request logging
app.use((req, res, next) => {
  const start = Date.now();
  res.on('finish', () => {
    console.log(`${req.method} ${req.url} ${res.statusCode} ${Date.now() - start}ms`);
  });
  next();
});

// Error handling middleware
app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(err.status || 500).json({
    error: {
      message: err.message,
      ...(process.env.NODE_ENV === 'development' && { stack: err.stack }),
    },
  });
});
```

### Fastify Plugin Architecture

```javascript
import Fastify from 'fastify';
import fastifySwagger from '@fastify/swagger';
import fastifySwaggerUi from '@fastify/swagger-ui';
import fastifyCors from '@fastify/cors';

const fastify = Fastify({
  logger: {
    level: 'info',
    transport: {
      target: 'pino-pretty',
    },
  },
});

// Register plugins
await fastify.register(fastifyCors, { origin: true });

await fastify.register(fastifySwagger, {
  openapi: {
    info: {
      title: 'My API',
      version: '1.0.0',
    },
  },
});

await fastify.register(fastifySwaggerUi, {
  routePrefix: '/docs',
});

// Custom plugin
const myPlugin = async (fastify, options) => {
  fastify.decorate('db', options.database);

  fastify.addHook('onRequest', async (request) => {
    request.startTime = Date.now();
  });

  fastify.addHook('onResponse', async (request, reply) => {
    const duration = Date.now() - request.startTime;
    fastify.log.info({ duration, url: request.url }, 'request completed');
  });
};

fastify.register(myPlugin, { database: db });
```

### Hono Adapters and Middleware

```javascript
import { Hono } from 'hono';
import { serve } from '@hono/node-server';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';
import { jwt } from 'hono/jwt';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';

const app = new Hono();

// Middleware stack
app.use('*', logger());
app.use('*', secureHeaders());
app.use('/api/*', cors());

// JWT authentication
app.use('/api/protected/*', jwt({ secret: process.env.JWT_SECRET }));

// Zod validation
const createUserSchema = z.object({
  name: z.string().min(2).max(100),
  email: z.string().email(),
});

app.post('/api/users',
  zValidator('json', createUserSchema),
  async (c) => {
    const data = c.req.valid('json');
    const user = await db.users.create(data);
    return c.json(user, 201);
  }
);

// Error handling
app.onError((err, c) => {
  console.error(err);
  return c.json({ error: err.message }, 500);
});

// Not found handler
app.notFound((c) => c.json({ error: 'Not found' }, 404));

// Node.js adapter
serve({ fetch: app.fetch, port: 3000 });

// Or export for Cloudflare Workers, Deno, Bun
export default app;
```

---

## Testing Reference

### Vitest vs Jest Comparison

| Feature | Vitest | Jest |
|---------|--------|------|
| Speed | 4x faster cold, instant HMR | Baseline |
| ESM Support | Native | Requires config |
| TypeScript | Native | Via ts-jest/babel |
| Configuration | vite.config.js | jest.config.js |
| Watch Mode | Instant rerun | Full rerun |
| Snapshot Testing | Yes | Yes |
| Coverage | v8/istanbul | istanbul |
| Concurrent Tests | Per-file default | Optional |

### Vitest Mocking Patterns

```javascript
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { fetchUser, createUser } from './user.js';

// Mock module
vi.mock('./database.js', () => ({
  db: {
    users: {
      findById: vi.fn(),
      create: vi.fn(),
    },
  },
}));

import { db } from './database.js';

describe('User functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('fetches user from database', async () => {
    const mockUser = { id: 1, name: 'John' };
    db.users.findById.mockResolvedValue(mockUser);

    const user = await fetchUser(1);

    expect(db.users.findById).toHaveBeenCalledWith(1);
    expect(user).toEqual(mockUser);
  });

  it('handles fetch errors', async () => {
    db.users.findById.mockRejectedValue(new Error('DB Error'));

    await expect(fetchUser(1)).rejects.toThrow('DB Error');
  });

  // Spy on existing implementation
  it('spies on console.log', () => {
    const spy = vi.spyOn(console, 'log');
    console.log('test');
    expect(spy).toHaveBeenCalledWith('test');
  });

  // Timer mocks
  it('handles timers', async () => {
    vi.useFakeTimers();

    const callback = vi.fn();
    setTimeout(callback, 1000);

    vi.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalled();

    vi.useRealTimers();
  });
});
```

---

## Build Tools Reference

### Vite Configuration

```javascript
// vite.config.js
import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    target: 'es2022',
    outDir: 'dist',
    lib: {
      entry: resolve(__dirname, 'src/index.js'),
      name: 'MyLib',
      formats: ['es', 'cjs'],
      fileName: (format) => `index.${format === 'es' ? 'js' : 'cjs'}`,
    },
    rollupOptions: {
      external: ['express', 'fastify'],
      output: {
        manualChunks: {
          vendor: ['lodash-es'],
        },
      },
    },
    minify: 'esbuild',
    sourcemap: true,
  },
  esbuild: {
    target: 'es2022',
    keepNames: true,
  },
  server: {
    port: 3000,
    hmr: true,
  },
});
```

### esbuild Direct Usage

```javascript
// build.js
import * as esbuild from 'esbuild';

await esbuild.build({
  entryPoints: ['src/index.js'],
  bundle: true,
  minify: true,
  sourcemap: true,
  target: ['es2022'],
  platform: 'node',
  format: 'esm',
  outdir: 'dist',
  external: ['express', 'pg'],
  define: {
    'process.env.NODE_ENV': '"production"',
  },
});

// Watch mode
const ctx = await esbuild.context({
  entryPoints: ['src/index.js'],
  bundle: true,
  outdir: 'dist',
});

await ctx.watch();
console.log('watching...');
```

---

## Context7 Library Mappings

### Primary Libraries

```
/nodejs/node           - Node.js runtime
/expressjs/express     - Express web framework
/fastify/fastify       - Fastify web framework
/honojs/hono           - Hono web framework
/koajs/koa             - Koa web framework
```

### Testing

```
/vitest-dev/vitest     - Vitest testing framework
/jestjs/jest           - Jest testing framework
/testing-library       - Testing Library
```

### Build Tools

```
/vitejs/vite           - Vite build tool
/evanw/esbuild         - esbuild bundler
/rollup/rollup         - Rollup bundler
/biomejs/biome         - Biome linter/formatter
/eslint/eslint         - ESLint linter
```

### Utilities

```
/lodash/lodash         - Lodash utilities
/date-fns/date-fns     - Date utilities
/axios/axios           - HTTP client
/prisma/prisma         - Prisma ORM
```

---

## Security Best Practices

### Input Validation

```javascript
import { z } from 'zod';

const userSchema = z.object({
  name: z.string().min(1).max(100).trim(),
  email: z.string().email().toLowerCase(),
  age: z.number().int().min(0).max(150).optional(),
});

function validateUser(input) {
  const result = userSchema.safeParse(input);
  if (!result.success) {
    throw new Error(result.error.issues[0].message);
  }
  return result.data;
}
```

### Environment Variable Validation

```javascript
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']),
  PORT: z.string().transform(Number).pipe(z.number().min(1).max(65535)),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
});

const env = envSchema.parse(process.env);
export default env;
```

### Secure HTTP Headers

```javascript
import helmet from 'helmet';

app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
  crossOriginEmbedderPolicy: true,
  crossOriginOpenerPolicy: true,
  crossOriginResourcePolicy: { policy: "same-origin" },
  hsts: { maxAge: 31536000, includeSubDomains: true },
}));
```

---

Last Updated: 2025-12-22
Version: 1.0.0
