---
name: moai-lang-javascript
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Enterprise JavaScript for Node.js and browser: Node.js 22.11.0 LTS (Jod),
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# JavaScript Runtime & Ecosystem — Enterprise  

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Lines** | ~950 lines |
| **Size** | ~30KB |
| **Tier** | **3 (Professional)** |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch |
| **Auto-load** | Node.js runtime, async operations, npm ecosystem, ES2025 features |
| **Trigger cues** | JavaScript, Node.js, npm, async, module, ES2025, runtime, npm scripts |

## Technology Stack (November 2025 Stable)

### Runtime
- **Node.js 22.11.0 LTS** (Jod, long-term support until 2027)
  - ES2025 features
  - Performance improvements
  - Security updates
  - Stability guarantee

- **Bun 1.1.x** (Alternative runtime)
- **Deno 2.x** (Modern runtime)

### Package Manager
- **npm 11.x** (Default package manager)
  - Workspace support
  - Provenance attestation
  - Dependency auditing

- **yarn 4.x** (Alternative, v4 faster)
- **pnpm 9.x** (Efficient package management)

### Web Frameworks
- **Express 4.21.x** (Mature framework)
  - Middleware system
  - Routing
  - Template engines

- **Fastify 5.x** (Performance-focused)
  - Low overhead
  - Validation
  - Hooks system

- **Hapi 21.x** (Enterprise framework)
  - Plugin ecosystem
  - Request validation

### Async/Concurrency
- **Promises** (Native async handling)
- **async/await** (Modern syntax)
- **EventEmitter** (Event-driven architecture)

### Module Systems
- **ES Modules** (Standard, import/export)
- **CommonJS** (Traditional, require/module.exports)
- **Dual module packages** (Supporting both)

### Testing & Quality
- **Vitest 2.x** (Unit testing)
- **Jest 30.x** (Comprehensive testing)
- **Mocha 10.x** (Simple test runner)

### Build Tools
- **Webpack 6.x** (Module bundler)
- **Turbopack** (Rust-based, fast)
- **esbuild 0.23.x** (Fast transpiler)

## Level 1: Fundamentals (High Freedom)

### 1. Node.js 22 Runtime

Node.js 22 brings ES2025 features and improvements:

**Basic HTTP Server**:
```javascript
const http = require('http');

const server = http.createServer((req, res) => {
    if (req.url === '/hello') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ message: 'Hello, World!' }));
    } else {
        res.writeHead(404);
        res.end('Not Found');
    }
});

server.listen(3000, () => {
    console.log('Server running on :3000');
});
```

**ES Modules**:
```javascript
// math.mjs
export function add(a, b) {
    return a + b;
}

export function multiply(a, b) {
    return a * b;
}

// app.mjs
import { add, multiply } from './math.mjs';

console.log(add(2, 3)); // 5
console.log(multiply(2, 3)); // 6
```

**Async/Await Patterns**:
```javascript
async function fetchUserData(userId) {
    try {
        const response = await fetch(\`https://api.example.com/users/\${userId}\`);
        
        if (!response.ok) {
            throw new Error(\`API error: \${response.statusText}\`);
        }
        
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Fetch failed:', error);
        throw error;
    }
}

async function main() {
    const user = await fetchUserData(1);
    console.log(user);
}

main();
```

### 2. npm Ecosystem Management

npm 11 provides modern package management:

**Package.json Best Practices**:
```json
{
    "name": "@org/myapp",
    "version": "1.0.0",
    "description": "Application description",
    "type": "module",
    "main": "dist/index.js",
    "exports": {
        ".": "./dist/index.js",
        "./utils": "./dist/utils.js"
    },
    "engines": {
        "node": ">=22.0.0",
        "npm": ">=11.0.0"
    },
    "scripts": {
        "dev": "node --watch src/index.js",
        "build": "tsc",
        "test": "vitest",
        "lint": "eslint ."
    },
    "dependencies": {
        "express": "^4.21.0",
        "dotenv": "^16.0.0"
    },
    "devDependencies": {
        "@types/node": "^22.0.0",
        "vitest": "^2.0.0"
    }
}
```

**npm Scripts**:
```bash
# Development
npm run dev

# Build
npm run build

# Test
npm test

# Publish
npm publish

# Audit dependencies
npm audit
npm audit fix
```

### 3. Express Framework (4.21.x)

Express provides a lightweight web framework:

**Basic Express App**:
```javascript
import express from 'express';
import { json } from 'express';

const app = express();
app.use(json());

// Routes
app.get('/users', (req, res) => {
    res.json([
        { id: 1, name: 'John' },
        { id: 2, name: 'Jane' }
    ]);
});

app.post('/users', (req, res) => {
    const { name } = req.body;
    res.status(201).json({ id: 3, name });
});

app.get('/users/:id', (req, res) => {
    const { id } = req.params;
    res.json({ id: parseInt(id), name: 'User' });
});

// Error handling
app.use((err, req, res, next) => {
    console.error(err.stack);
    res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(3000, () => {
    console.log('Server running on :3000');
});
```

**Middleware**:
```javascript
// Logging middleware
const logger = (req, res, next) => {
    console.log(\`\${req.method} \${req.path}\`);
    next();
};

// Authentication middleware
const authenticateToken = (req, res, next) => {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];
    
    if (!token) {
        return res.status(401).json({ error: 'Missing token' });
    }
    
    // Verify token...
    next();
};

app.use(logger);
app.use('/api/protected', authenticateToken);

app.get('/api/protected/data', (req, res) => {
    res.json({ data: 'sensitive information' });
});
```

## Level 2: Advanced Patterns (Medium Freedom)

### 1. Fastify Framework (5.x)

Fastify is performance-focused alternative:

**Fastify Setup**:
```javascript
import Fastify from 'fastify';

const fastify = Fastify({ logger: true });

fastify.post<{ Body: { name: string } }>('/users', async (request, reply) => {
    const { name } = request.body;
    return { id: 1, name };
});

fastify.get<{ Params: { id: string } }>('/users/:id', async (request, reply) => {
    const { id } = request.params;
    return { id: parseInt(id), name: 'User' };
});

fastify.listen({ port: 3000 }, (err, address) => {
    if (err) throw err;
    console.log(\`Server listening at \${address}\`);
});
```

### 2. Promise Patterns

Advanced Promise usage:

**Promise Composition**:
```javascript
// Promise.all - parallel execution
async function fetchMultipleUsers() {
    const [user1, user2, user3] = await Promise.all([
        fetch('/api/users/1').then(r => r.json()),
        fetch('/api/users/2').then(r => r.json()),
        fetch('/api/users/3').then(r => r.json()),
    ]);
    
    return [user1, user2, user3];
}

// Promise.race - first to complete
async function fetchWithTimeout(promise, timeout) {
    return Promise.race([
        promise,
        new Promise((_, reject) =>
            setTimeout(() => reject(new Error('Timeout')), timeout)
        ),
    ]);
}

// Promise.allSettled - all results
async function fetchAll(urls) {
    const results = await Promise.allSettled(
        urls.map(url => fetch(url).then(r => r.json()))
    );
    return results;
}
```

### 3. Testing with Vitest

Vitest for modern testing:

**Unit Tests**:
```javascript
import { describe, it, expect } from 'vitest';
import { add, multiply } from './math.js';

describe('Math utils', () => {
    it('should add numbers', () => {
        expect(add(2, 3)).toBe(5);
    });
    
    it('should multiply numbers', () => {
        expect(multiply(2, 3)).toBe(6);
    });
});
```

**Integration Tests**:
```javascript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { createServer } from './server.js';

describe('API Endpoints', () => {
    let server;
    
    beforeEach(() => {
        server = createServer();
    });
    
    afterEach(() => {
        server.close();
    });
    
    it('GET /users returns user list', async () => {
        const response = await fetch('http://localhost:3000/users');
        const data = await response.json();
        
        expect(response.ok).toBe(true);
        expect(Array.isArray(data)).toBe(true);
    });
});
```

## Level 3: Production Deployment (Low Freedom, Expert Only)

### 1. Process Management

**PM2 for Process Management**:
```bash
npm install -g pm2
pm2 start app.js --name "myapp"
pm2 save
pm2 startup
```

### 2. Docker Deployment

```dockerfile
FROM node:22-alpine

WORKDIR /app
COPY package*.json ./

RUN npm ci --only=production

COPY . .
EXPOSE 3000

CMD ["node", "src/index.js"]
```

### 3. Environment Configuration

```javascript
// config.js
const config = {
    development: {
        port: 3000,
        database: 'mongodb://localhost:27017/mydb',
        logLevel: 'debug',
    },
    production: {
        port: process.env.PORT || 8080,
        database: process.env.DATABASE_URL,
        logLevel: 'info',
    },
};

const env = process.env.NODE_ENV || 'development';
export default config[env];
```

## Auto-Load Triggers

This Skill activates when you:
- Work with Node.js projects and JavaScript
- Need async/await pattern guidance
- Implement npm package management
- Use Express or Fastify frameworks
- Debug JavaScript runtime issues
- Handle module systems
- Optimize performance

## Best Practices Summary

1. Always use async/await for async operations
2. Handle promise rejections properly
3. Use strict type checking with TypeScript
4. Validate input in API endpoints
5. Implement proper error handling
6. Use npm workspaces for monorepos
7. Test with Vitest or Jest
8. Monitor with observability tools
9. Use environment variables for config
10. Implement graceful shutdown

