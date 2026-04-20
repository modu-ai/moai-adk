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

# Install (30x faster than npm)
bun install
bun add express
bun add -d vitest

# Run scripts
bun run dev
bun run test

# Execute files directly (native TypeScript/JSX support)
bun run server.js
bun run app.ts
bun run app.tsx

# Built-in bundler
bun build ./src/index.ts --outdir=./dist
bun build ./index.html --outdir=./dist    # HTML bundling
bun build --splitting                      # Code splitting
bun build --minify                         # Minification

# Compile to standalone executable
bun build --compile ./src/index.ts --outfile=myapp

# Run tests
bun test
bun test --watch
bun test --coverage
bun test --bail                            # Stop on first failure

# Hot reloading (state preserved)
bun --hot server.ts
```

---

## Bun Complete API Reference

### Bun.serve() - HTTP Server

Basic Server:
```typescript
const server = Bun.serve({
  port: 3000,
  hostname: "0.0.0.0",
  fetch(req) {
    const url = new URL(req.url);
    if (url.pathname === "/") return new Response("Hello!");
    if (url.pathname === "/json") {
      return Response.json({ message: "Hello JSON" });
    }
    return new Response("Not Found", { status: 404 });
  },
  error(error) {
    return new Response(`Error: ${error.message}`, { status: 500 });
  },
});
console.log(`Server running at ${server.url}`);
```

Routes Configuration (Bun 1.2.3+):
```typescript
Bun.serve({
  routes: {
    // Static routes
    "/api/status": new Response("OK"),

    // Dynamic routes with parameters
    "/users/:id": req => new Response(`User ${req.params.id}`),

    // Per-method handlers
    "/api/posts": {
      GET: () => Response.json({ posts: [] }),
      POST: async req => {
        const body = await req.json();
        return Response.json({ created: true, ...body }, { status: 201 });
      },
    },

    // Wildcard routes
    "/api/*": Response.json({ error: "Not found" }, { status: 404 }),

    // Static file serving
    "/favicon.ico": Bun.file("./favicon.ico"),
  },

  // Fallback handler
  fetch(req) {
    return new Response("Not Found", { status: 404 });
  },
});
```

TLS/HTTPS Configuration:
```typescript
Bun.serve({
  port: 443,
  tls: {
    cert: Bun.file("./cert.pem"),
    key: Bun.file("./key.pem"),
    ca: Bun.file("./ca.pem"),         // Optional CA certificate
    passphrase: "secret",              // Optional key passphrase
  },
  fetch(req) {
    return new Response("Secure!");
  },
});
```

WebSocket Server with Full Options:
```typescript
Bun.serve({
  port: 3001,
  fetch(req, server) {
    const success = server.upgrade(req, {
      data: { userId: crypto.randomUUID() }, // Per-connection data
    });
    if (success) return undefined;
    return new Response("Upgrade failed", { status: 400 });
  },
  websocket: {
    open(ws) {
      ws.subscribe("chat");            // Subscribe to topic
      console.log("Connected:", ws.data.userId);
    },
    message(ws, message) {
      ws.publish("chat", message);     // Publish to all subscribers
      ws.send(`Echo: ${message}`);
    },
    close(ws, code, reason) {
      ws.unsubscribe("chat");
      console.log("Disconnected:", code, reason);
    },
    // Advanced options
    maxPayloadLength: 16 * 1024 * 1024,  // 16MB max message size
    backpressureLimit: 1024 * 1024,      // 1MB backpressure limit
    idleTimeout: 120,                     // 2 minutes idle timeout
    perMessageDeflate: true,              // Enable compression
    sendPings: true,                      // Send ping frames
  },
});
```

Server Metrics and Lifecycle:
```typescript
const server = Bun.serve({
  fetch(req, server) {
    // Get client IP
    const ip = server.requestIP(req);

    // Set custom timeout for this request
    server.timeout(req, 60); // 60 seconds

    return Response.json({
      activeRequests: server.pendingRequests,
      activeWebSockets: server.pendingWebSockets,
      chatUsers: server.subscriberCount("chat"),
      clientIP: ip?.adddess,
    });
  },
});

// Hot reload routes without restart
server.reload({
  routes: {
    "/api/version": () => Response.json({ version: "2.0.0" }),
  },
});

// Graceful shutdown
await server.stop();      // Wait for in-flight requests
await server.stop(true);  // Force close all connections
```

### Bun.file() - File Operations

Reading Files:
```typescript
const file = Bun.file("./data.txt");

// File metadata
file.size;              // number of bytes
file.type;              // MIME type
await file.exists();    // boolean

// Reading methods
const text = await file.text();           // string
const json = await file.json();           // parsed JSON
const bytes = await file.bytes();         // Uint8Array
const buffer = await file.arrayBuffer();  // ArrayBuffer
const stream = file.stream();             // ReadableStream

// Partial reads (HTTP Range header)
const first1KB = await file.slice(0, 1024).text();
const last500 = await file.slice(-500).text();

// File references
Bun.file(1234);                          // file descriptor
Bun.file(new URL(import.meta.url));      // file:// URL
Bun.file("data.json", { type: "application/json" }); // custom MIME
```

Writing Files:
```typescript
// Simple writes (returns bytes written)
await Bun.write("./output.txt", "Hello, Bun!");
await Bun.write("./data.json", JSON.stringify({ key: "value" }));

// Copy file
await Bun.write(Bun.file("output.txt"), Bun.file("input.txt"));

// Write from Response
const response = await fetch("https://example.com");
await Bun.write("index.html", response);

// Write to stdout
await Bun.write(Bun.stdout, "Hello stdout!\n");

// Delete file
await Bun.file("temp.txt").delete();
```

Incremental Writing (FileSink):
```typescript
const file = Bun.file("large-output.txt");
const writer = file.writer({ highWaterMark: 1024 * 1024 }); // 1MB buffer

writer.write("First chunk\n");
writer.write("Second chunk\n");
writer.flush();  // Flush buffer to disk
writer.end();    // Flush and close

// Control process lifecycle
writer.unref();  // Allow process to exit
writer.ref();    // Re-ref later
```

### Bun Shell

Basic Usage:
```typescript
import { $ } from "bun";

// Execute commands
await $`echo "Hello World!"`;

// Get output
const text = await $`ls -la`.text();
const json = await $`cat config.json`.json();
const blob = await $`cat image.png`.blob();

// Iterate lines
for await (const line of $`cat file.txt`.lines()) {
  console.log(line);
}
```

Piping and Redirection:
```typescript
// Pipe commands
const wordCount = await $`echo "Hello World!" | wc -w`.text();

// Redirect to file
await $`echo "Hello" > greeting.txt`;
await $`echo "More" >> greeting.txt`;  // Append

// Redirect stderr
await $`command 2> error.log`;
await $`command &> all.log`;  // Both stdout and stderr

// JavaScript object as stdin
const response = new Response("input data");
await $`cat < ${response} | wc -c`;
```

Environment and Working Directory:
```typescript
// Inline environment
await $`FOO=bar bun -e 'console.log(process.env.FOO)'`;

// Per-command environment
await $`echo $API_KEY`.env({ ...process.env, API_KEY: "secret" });

// Global environment
$.env({ NODE_ENV: "production" });

// Working directory
await $`pwd`.cwd("/tmp");
$.cwd("/home/user");
```

Error Handling:
```typescript
// Default: throws on non-zero exit code
try {
  await $`failing-command`.text();
} catch (err) {
  console.log(`Exit code: ${err.exitCode}`);
  console.log(err.stdout.toString());
  console.log(err.stderr.toString());
}

// Disable throwing
const { stdout, stderr, exitCode } = await $`command`.nothrow().quiet();

// Global configuration
$.nothrow();       // Don't throw by default
$.throws(true);    // Restore default
```

### Bun SQLite

Database Connection:
```typescript
import { Database } from "bun:sqlite";

// File-based
const db = new Database("mydb.sqlite");

// In-memory
const memoryDb = new Database(":memory:");

// Options
const db = new Database("mydb.sqlite", {
  readonly: true,      // Read-only mode
  create: true,        // Create if not exists
  strict: true,        // Enable strict mode
  safeIntegers: true,  // Return bigint for large integers
});

// Import via attribute
import db from "./mydb.sqlite" with { type: "sqlite" };

// WAL mode (recommended for performance)
db.run("PRAGMA journal_mode = WAL;");
```

Prepared Statements:
```typescript
// Create table
db.run(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE
  )
`);

// Prepare and execute
const insertUser = db.prepare(
  "INSERT INTO users (name, email) VALUES ($name, $email)"
);
insertUser.run({ $name: "John", $email: "john@example.com" });

// Query methods
const getUser = db.prepare("SELECT * FROM users WHERE id = ?");
const user = getUser.get(1);           // Single row as object
const users = getUser.all();           // All rows as array
const values = getUser.values();       // All rows as arrays
const result = getUser.run();          // { lastInsertRowid, changes }

// Iterate lazily
for (const row of getUser.iterate()) {
  console.log(row);
}

// Map to class
class User {
  id: number;
  name: string;
  get displayName() { return `User: ${this.name}`; }
}
const users = db.query("SELECT * FROM users").as(User).all();
```

Transactions:
```typescript
const insertUser = db.prepare("INSERT INTO users (name) VALUES (?)");

// Transaction function
const insertMany = db.transaction((names: string[]) => {
  for (const name of names) {
    insertUser.run(name);
  }
  return names.length;
});

const count = insertMany(["Alice", "Bob", "Charlie"]);

// Transaction types
insertMany(data);              // BEGIN
insertMany.deferred(data);     // BEGIN DEFERRED
insertMany.immediate(data);    // BEGIN IMMEDIATE
insertMany.exclusive(data);    // BEGIN EXCLUSIVE
```

### Bun Test

Test Structure:
```typescript
import { describe, it, test, expect, beforeAll, beforeEach, afterEach, afterAll, mock } from "bun:test";

describe("User Service", () => {
  beforeAll(() => { /* Setup once */ });
  beforeEach(() => { /* Setup each */ });
  afterEach(() => { /* Cleanup each */ });
  afterAll(() => { /* Cleanup once */ });

  it("should create a user", () => {
    expect({ name: "John" }).toEqual({ name: "John" });
  });

  test("async operations", async () => {
    const data = await fetchData();
    expect(data).toBeDefined();
  });

  // Concurrent tests
  test.concurrent("parallel test 1", async () => { /* ... */ });
  test.concurrent("parallel test 2", async () => { /* ... */ });

  // Skip and todo
  test.skip("skipped test", () => { /* ... */ });
  test.todo("implement later");
});
```

Mocking:
```typescript
import { mock, spyOn } from "bun:test";

// Mock function
const fn = mock(() => 42);
fn();
expect(fn).toHaveBeenCalled();
expect(fn).toHaveBeenCalledTimes(1);

// Spy on existing
const spy = spyOn(console, "log");
console.log("test");
expect(spy).toHaveBeenCalledWith("test");
```

Snapshots:
```typescript
test("snapshot", () => {
  expect({ a: 1, b: 2 }).toMatchSnapshot();
});

// Update: bun test --update-snapshots
```

Test Commands:
```bash
bun test                          # Run all tests
bun test --watch                  # Watch mode
bun test --coverage               # Code coverage
bun test --bail                   # Stop on first failure
bun test --timeout 10000          # 10s timeout
bun test -t "pattern"             # Filter by name
bun test --concurrent             # Parallel execution
bun test --reporter=junit         # JUnit XML output
```

### Bun S3 Client

Basic Operations:
```typescript
import { s3, S3Client } from "bun";

// Read from S3
const file = s3.file("data/config.json");
const data = await file.json();
const text = await file.text();
const bytes = await file.bytes();
const stream = file.stream();

// Partial read
const first1KB = await file.slice(0, 1024).text();

// Write to S3
await file.write("Hello World!");
await file.write(JSON.stringify(data), { type: "application/json" });

// Delete
await file.delete();

// Check existence
const exists = await file.exists();
const size = await file.size();
```

Presigned URLs:
```typescript
// Download URL
const downloadUrl = s3.presign("my-file.txt", {
  expiresIn: 3600,  // 1 hour
});

// Upload URL
const uploadUrl = s3.presign("my-file.txt", {
  method: "PUT",
  expiresIn: 3600,
  type: "application/json",
  acl: "public-read",
});
```

S3 Client Configuration:
```typescript
const client = new S3Client({
  accessKeyId: process.env.S3_ACCESS_KEY_ID,
  secretAccessKey: process.env.S3_SECRET_ACCESS_KEY,
  bucket: "my-bucket",
  region: "us-east-1",
  // Or use endpoint for S3-compatible services
  endpoint: "https://s3.us-east-1.amazonaws.com",
});

// Works with: AWS S3, Cloudflare R2, DigitalOcean Spaces,
// MinIO, Google Cloud Storage, Supabase Storage
```

Large File Streaming:
```typescript
const writer = s3.file("large-file.bin").writer({
  retry: 3,
  queueSize: 10,
  partSize: 5 * 1024 * 1024,  // 5MB chunks
});

for (const chunk of largeData) {
  writer.write(chunk);
  await writer.flush();
}
await writer.end();
```

### Bun Glob

Pattern Matching:
```typescript
import { Glob } from "bun";

// Create glob instance
const glob = new Glob("**/*.ts");

// Async iteration
for await (const file of glob.scan(".")) {
  console.log(file);
}

// Sync iteration
for (const file of glob.scanSync(".")) {
  console.log(file);
}

// String matching
glob.match("src/index.ts");  // true
glob.match("src/index.js");  // false
```

Scan Options:
```typescript
const glob = new Glob("**/*.ts");

for await (const file of glob.scan({
  cwd: "./src",
  dot: true,              // Include dotfiles
  absolute: true,         // Return absolute paths
  followSymlinks: true,   // Follow symlinks
  onlyFiles: true,        // Files only (default)
})) {
  console.log(file);
}
```

Pattern Syntax:
```typescript
// Single character: ?
new Glob("???.ts").match("foo.ts");         // true

// Zero or more chars (no path sep): *
new Glob("*.ts").match("index.ts");         // true
new Glob("*.ts").match("src/index.ts");     // false

// Any chars including path sep: **
new Glob("**/*.ts").match("src/index.ts");  // true

// Character sets: [abc], [a-z], [^abc]
new Glob("ba[rz].ts").match("bar.ts");      // true
new Glob("ba[!a-z].ts").match("ba1.ts");    // true

// Alternation: {a,b,c}
new Glob("{src,lib}/**/*.ts").match("src/index.ts");  // true

// Negation: !
new Glob("!node_modules/**").match("src/index.ts");   // true
```

### Bun Semver

Version Comparison:
```typescript
import { semver } from "bun";

// Check if version satisfies range
semver.satisfies("1.0.0", "^1.0.0");   // true
semver.satisfies("2.0.0", "^1.0.0");   // false
semver.satisfies("1.0.0", "~1.0.0");   // true
semver.satisfies("1.0.0", "1.0.x");    // true
semver.satisfies("1.0.0", "1.0.0 - 2.0.0");  // true

// Compare versions
semver.order("1.0.0", "1.0.0");   // 0
semver.order("1.0.0", "1.0.1");   // -1
semver.order("1.0.1", "1.0.0");   // 1

// Sort versions
const versions = ["1.0.0", "1.0.1", "1.0.0-alpha", "1.0.0-beta"];
versions.sort(semver.order);
// ["1.0.0-alpha", "1.0.0-beta", "1.0.0", "1.0.1"]
```

### Bun DNS

DNS Resolution:
```typescript
import { dns } from "bun";
import * as nodeDns from "node:dns";

// Prefetch DNS (optimization)
dns.prefetch("api.example.com", 443);

// Get cache stats
const stats = dns.getCacheStats();
console.log(stats);
// { cacheHitsCompleted, cacheHitsInflight, cacheMisses, size, errors, totalCount }

// Node.js compatible API
const addds = await nodeDns.promises.resolve4("bun.sh", { ttl: true });
// [{ adddess: "172.67.161.226", family: 4, ttl: 0 }, ...]
```

### Bun Bundler

Build API:
```typescript
const result = await Bun.build({
  entrypoints: ["./src/index.ts"],
  outdir: "./dist",
  target: "browser",         // browser, bun, node
  format: "esm",             // esm, cjs, iife
  splitting: true,           // Code splitting
  minify: true,              // Full minification
  sourcemap: "linked",       // none, linked, external, inline

  // Environment variables
  env: "PUBLIC_*",           // Inline PUBLIC_* vars

  // External modules
  external: ["lodash"],

  // Custom loaders
  loader: {
    ".png": "dataurl",
    ".txt": "file",
  },

  // Naming patterns
  naming: {
    entry: "[dir]/[name].[ext]",
    chunk: "[name]-[hash].[ext]",
    asset: "[name]-[hash].[ext]",
  },

  // Public path for CDN
  publicPath: "https://cdn.example.com/",

  // Define replacements
  define: {
    "process.env.VERSION": JSON.stringify("1.0.0"),
  },

  // Drop function calls
  drop: ["console", "debugger"],

  // Plugins
  plugins: [myPlugin],
});

if (!result.success) {
  console.error(result.logs);
}
```

HTML Bundling:
```typescript
// Build static site
await Bun.build({
  entrypoints: ["./index.html", "./about.html"],
  outdir: "./dist",
  minify: true,
});

// Or run dev server
// bun ./index.html
```

### Bun Plugins

Plugin Structure:
```typescript
import type { BunPlugin } from "bun";

const myPlugin: BunPlugin = {
  name: "my-plugin",
  setup(build) {
    // Runs when bundle starts
    build.onStart(() => {
      console.log("Bundle started!");
    });

    // Custom module resolution
    build.onResolve({ filter: /^virtual:/ }, args => {
      return { path: args.path, namespace: "virtual" };
    });

    // Custom module loading
    build.onLoad({ filter: /.*/, namespace: "virtual" }, args => {
      return {
        contents: `export default "Virtual module: ${args.path}"`,
        loader: "js",
      };
    });
  },
};

await Bun.build({
  entrypoints: ["./index.ts"],
  plugins: [myPlugin],
});
```

### Bun Macros

Bundle-Time Execution:
```typescript
// getVersion.ts
export function getVersion() {
  const { stdout } = Bun.spawnSync({
    cmd: ["git", "rev-parse", "HEAD"],
    stdout: "pipe",
  });
  return stdout.toString().trim();
}

// app.ts
import { getVersion } from "./getVersion.ts" with { type: "macro" };

// Executed at bundle-time, result inlined
console.log(`Version: ${getVersion()}`);
```

Fetch at Bundle-Time:
```typescript
// fetchData.ts
export async function fetchConfig(url: string) {
  const response = await fetch(url);
  return response.json();
}

// app.ts
import { fetchConfig } from "./fetchData.ts" with { type: "macro" };

// Config fetched during build, not runtime
const config = fetchConfig("https://api.example.com/config");
```

### Bun Hot Reloading

State-Preserving Reload:
```typescript
// server.ts
declare global {
  var count: number;
}

globalThis.count ??= 0;
globalThis.count++;

Bun.serve({
  fetch(req) {
    return new Response(`Reloaded ${globalThis.count} times`);
  },
  port: 3000,
});

// Run with: bun --hot server.ts
// Changes reload without restarting process
// Global state (globalThis) preserved
```

### Bun Workers

Web Workers:
```typescript
// worker.ts
declare var self: Worker;

self.onmessage = (event: MessageEvent) => {
  const result = heavyComputation(event.data);
  self.postMessage(result);
};

function heavyComputation(input: number): number {
  return input * 2;
}

// main.ts
const worker = new Worker(new URL("./worker.ts", import.meta.url));

worker.onmessage = (event) => {
  console.log("Result:", event.data);
};

worker.postMessage(42);
```

### Bun Utilities

Password Hashing:
```typescript
const hash = await Bun.password.hash("mypassword", {
  algorithm: "argon2id",  // argon2id, argon2i, argon2d, bcrypt
  memoryCost: 65536,
  timeCost: 3,
});

const isValid = await Bun.password.verify("mypassword", hash);
```

Other Utilities:
```typescript
// Sleep
await Bun.sleep(1000);  // 1 second

// UUID v7 (time-ordered)
const uuid = Bun.randomUUIDv7();

// Environment variables (auto-loads .env)
const apiKey = Bun.env.API_KEY;

// Peek at promises without awaiting
const status = Bun.peek(promise);  // pending, fulfilled, rejected

// Deep equals
Bun.deepEquals({ a: 1 }, { a: 1 });  // true

// Escape HTML
Bun.escapeHTML("<script>alert('xss')</script>");

// Resolve import path
const path = Bun.resolveSync("./module", import.meta.dir);
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

Last Updated: 2026-01-05
Version: 1.1.0
