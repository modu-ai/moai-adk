---
name: moai-lang-javascript
description: JavaScript ES2024+ development specialist covering Node.js 22 LTS, modern runtimes (Deno, Bun), testing (Vitest, Jest), linting (ESLint 9, Biome), and backend frameworks (Express, Fastify, Hono). Use when developing JavaScript APIs, web applications, or Node.js projects.
version: 1.0.0
updated: 2025-12-22
status: active
allowed-tools: Read, Grep, Glob, Bash, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

JavaScript ES2024+ Development Specialist - Modern JavaScript with Node.js 22 LTS, multiple runtimes, and contemporary tooling.

Auto-Triggers: `.js`, `.mjs`, `.cjs` files, `package.json`, Node.js projects, JavaScript discussions

Core Stack:
- ES2024+: Set methods, Promise.withResolvers, immutable arrays, import attributes
- Node.js 22 LTS: Native TypeScript, built-in WebSocket, stable watch mode
- Runtimes: Node.js 20/22 LTS, Deno 2.x, Bun 1.x
- Testing: Vitest, Jest, Node.js test runner
- Linting: ESLint 9 flat config, Biome
- Bundlers: Vite, esbuild, Rollup
- Frameworks: Express, Fastify, Hono, Koa

Quick Commands:
```bash
# Create Vite project
npm create vite@latest my-app -- --template vanilla

# Initialize with modern tooling
npm init -y && npm install -D vitest eslint @eslint/js

# Run with Node.js watch mode
node --watch server.js

# Run TypeScript directly in Node.js 22+
node --experimental-strip-types app.ts
```

---

## Implementation Guide (5 minutes)

### ES2024 Key Features

Set Operations:
```javascript
const setA = new Set([1, 2, 3, 4]);
const setB = new Set([3, 4, 5, 6]);

setA.intersection(setB);      // Set {3, 4}
setA.union(setB);             // Set {1, 2, 3, 4, 5, 6}
setA.difference(setB);        // Set {1, 2}
setA.symmetricDifference(setB); // Set {1, 2, 5, 6}
setA.isSubsetOf(setB);        // false
setA.isSupersetOf(setB);      // false
setA.isDisjointFrom(setB);    // false
```

Promise.withResolvers():
```javascript
function createDeferred() {
  const { promise, resolve, reject } = Promise.withResolvers();
  return { promise, resolve, reject };
}

const deferred = createDeferred();
setTimeout(() => deferred.resolve('done'), 1000);
const result = await deferred.promise;
```

Immutable Array Methods:
```javascript
const original = [3, 1, 4, 1, 5];

// New methods return new arrays (don't mutate)
const sorted = original.toSorted();           // [1, 1, 3, 4, 5]
const reversed = original.toReversed();       // [5, 1, 4, 1, 3]
const spliced = original.toSpliced(1, 2, 9);  // [3, 9, 1, 5]
const changed = original.with(2, 99);         // [3, 1, 99, 1, 5]

console.log(original); // [3, 1, 4, 1, 5] - unchanged
```

Object.groupBy and Map.groupBy:
```javascript
const items = [
  { type: 'fruit', name: 'apple' },
  { type: 'vegetable', name: 'carrot' },
  { type: 'fruit', name: 'banana' },
];

const grouped = Object.groupBy(items, item => item.type);
// { fruit: [{...}, {...}], vegetable: [{...}] }

const mapGrouped = Map.groupBy(items, item => item.type);
// Map { 'fruit' => [...], 'vegetable' => [...] }
```

### ES2025 Features

Import Attributes (JSON Modules):
```javascript
import config from './config.json' with { type: 'json' };
import styles from './styles.css' with { type: 'css' };

console.log(config.apiUrl);
```

RegExp.escape:
```javascript
const userInput = 'hello (world)';
const safePattern = RegExp.escape(userInput);
// "hello\\ \\(world\\)"
const regex = new RegExp(safePattern);
```

### Node.js 22 LTS Features

Built-in WebSocket Client:
```javascript
const ws = new WebSocket('wss://example.com/socket');

ws.addEventListener('open', () => {
  ws.send(JSON.stringify({ type: 'hello' }));
});

ws.addEventListener('message', (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
});
```

Native TypeScript Support (Experimental):
```bash
# Run .ts files directly in Node.js 22.6+
node --experimental-strip-types app.ts

# In Node.js 22.18+, type stripping is enabled by default
node app.ts
```

Watch Mode (Stable):
```bash
# Auto-restart on file changes
node --watch server.js

# Watch specific files
node --watch-path=./src --watch-path=./config server.js
```

Permission Model:
```bash
# Restrict file system access
node --permission --allow-fs-read=/app/data server.js

# Restrict network access
node --permission --allow-net=api.example.com server.js
```

### Backend Frameworks

Express (Traditional):
```javascript
import express from 'express';

const app = express();
app.use(express.json());

app.get('/api/users', async (req, res) => {
  const users = await db.users.findAll();
  res.json(users);
});

app.post('/api/users', async (req, res) => {
  const user = await db.users.create(req.body);
  res.status(201).json(user);
});

app.listen(3000, () => console.log('Server running on port 3000'));
```

Fastify (High Performance):
```javascript
import Fastify from 'fastify';

const fastify = Fastify({ logger: true });

const userSchema = {
  body: {
    type: 'object',
    required: ['name', 'email'],
    properties: {
      name: { type: 'string', minLength: 2 },
      email: { type: 'string', format: 'email' },
    },
  },
};

fastify.post('/api/users', { schema: userSchema }, async (request, reply) => {
  const user = await db.users.create(request.body);
  return reply.code(201).send(user);
});

await fastify.listen({ port: 3000 });
```

Hono (Edge-First):
```javascript
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { validator } from 'hono/validator';

const app = new Hono();

app.use('*', logger());
app.use('/api/*', cors());

app.get('/api/users', async (c) => {
  const users = await db.users.findAll();
  return c.json(users);
});

app.post('/api/users',
  validator('json', (value, c) => {
    if (!value.name || !value.email) {
      return c.json({ error: 'Invalid input' }, 400);
    }
    return value;
  }),
  async (c) => {
    const user = await db.users.create(c.req.valid('json'));
    return c.json(user, 201);
  }
);

export default app;
```

### Testing with Vitest

Configuration:
```javascript
// vitest.config.js
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
});
```

Test Example:
```javascript
// user.test.js
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createUser, getUser } from './user.js';

describe('User Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should create a user', async () => {
    const user = await createUser({ name: 'John', email: 'john@example.com' });
    expect(user).toMatchObject({ name: 'John', email: 'john@example.com' });
    expect(user.id).toBeDefined();
  });

  it('should throw on invalid email', async () => {
    await expect(createUser({ name: 'John', email: 'invalid' }))
      .rejects.toThrow('Invalid email');
  });
});
```

### ESLint 9 Flat Config

```javascript
// eslint.config.js
import js from '@eslint/js';
import globals from 'globals';

export default [
  js.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 2025,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2025,
      },
    },
    rules: {
      'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'prefer-const': 'error',
      'no-var': 'error',
    },
  },
];
```

### Biome (All-in-One)

```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.0/schema.json",
  "organizeImports": { "enabled": true },
  "linter": {
    "enabled": true,
    "rules": { "recommended": true }
  },
  "formatter": {
    "enabled": true,
    "indentStyle": "space",
    "indentWidth": 2
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "semicolons": "always"
    }
  }
}
```

---

## Advanced Patterns

For comprehensive documentation including advanced async patterns, module system details, performance optimization, and production deployment configurations, see:

- [reference.md](reference.md) - Complete API reference, Context7 library mappings, package manager comparison
- [examples.md](examples.md) - Production-ready code examples, full-stack patterns, testing templates

### Context7 Integration

```javascript
// Node.js - mcp__context7__get_library_docs("/nodejs/node", "esm modules async", 1)
// Express - mcp__context7__get_library_docs("/expressjs/express", "middleware routing", 1)
// Fastify - mcp__context7__get_library_docs("/fastify/fastify", "plugins hooks", 1)
// Hono - mcp__context7__get_library_docs("/honojs/hono", "middleware validators", 1)
// Vitest - mcp__context7__get_library_docs("/vitest-dev/vitest", "mocking coverage", 1)
```

---

## Works Well With

- `moai-lang-typescript` - TypeScript integration, type checking with JSDoc
- `moai-domain-backend` - API design, microservices architecture
- `moai-domain-database` - Database integration, ORM patterns
- `moai-workflow-testing` - TDD workflows, testing strategies
- `moai-foundation-quality` - Code quality standards
- `moai-essentials-debug` - Debugging JavaScript applications

---

## Quick Troubleshooting

Module System Issues:
```bash
# Check package.json type
cat package.json | grep '"type"'

# ESM: "type": "module" - use import/export
# CommonJS: "type": "commonjs" or omitted - use require/module.exports
```

Node.js Version Check:
```bash
node --version  # Should be 20.x or 22.x LTS
npm --version   # Should be 10.x+
```

Common Fixes:
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
rm -rf node_modules package-lock.json && npm install

# Fix permission issues
npm config set prefix ~/.npm-global
```

ESM/CommonJS Interop:
```javascript
// Import CommonJS from ESM
import pkg from 'commonjs-package';
const { namedExport } = pkg;

// Dynamic import in CommonJS
const { default: esmModule } = await import('esm-package');
```

---

Last Updated: 2025-12-22
Status: Active (v1.0.0)
