# JavaScript Production-Ready Examples

## Full-Stack Application Setup

### Project Structure

```
my-api/
├── src/
│   ├── index.js              # Application entry point
│   ├── config/
│   │   ├── index.js          # Configuration loader
│   │   └── database.js       # Database configuration
│   ├── api/
│   │   ├── routes/
│   │   │   ├── index.js      # Route aggregator
│   │   │   ├── users.js      # User routes
│   │   │   └── posts.js      # Post routes
│   │   └── middleware/
│   │       ├── auth.js       # Authentication
│   │       ├── validate.js   # Request validation
│   │       └── errorHandler.js
│   ├── services/
│   │   ├── userService.js    # Business logic
│   │   └── postService.js
│   ├── repositories/
│   │   ├── userRepository.js # Data access
│   │   └── postRepository.js
│   └── utils/
│       ├── logger.js
│       └── helpers.js
├── test/
│   ├── unit/
│   ├── integration/
│   └── fixtures/
├── package.json
├── eslint.config.js
├── vitest.config.js
└── Dockerfile
```

### package.json

```json
{
  "name": "my-api",
  "version": "1.0.0",
  "type": "module",
  "engines": {
    "node": ">=22.0.0"
  },
  "scripts": {
    "dev": "node --watch src/index.js",
    "start": "node src/index.js",
    "test": "vitest",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "vitest run --config vitest.e2e.config.js",
    "lint": "eslint src/",
    "lint:fix": "eslint src/ --fix",
    "format": "biome format --write src/",
    "check": "biome check src/",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "fastify": "^5.0.0",
    "@fastify/cors": "^10.0.0",
    "@fastify/helmet": "^12.0.0",
    "@fastify/rate-limit": "^10.0.0",
    "@fastify/swagger": "^9.0.0",
    "@fastify/swagger-ui": "^5.0.0",
    "zod": "^3.23.0",
    "pg": "^8.13.0",
    "pino": "^9.5.0",
    "pino-pretty": "^13.0.0",
    "dotenv": "^16.4.0"
  },
  "devDependencies": {
    "@biomejs/biome": "^1.9.0",
    "eslint": "^9.15.0",
    "@eslint/js": "^9.15.0",
    "globals": "^15.12.0",
    "vitest": "^2.1.0",
    "@vitest/coverage-v8": "^2.1.0",
    "typescript": "^5.7.0",
    "@types/node": "^22.10.0"
  }
}
```

---

## Fastify Complete API Example

### src/index.js

```javascript
import Fastify from 'fastify';
import cors from '@fastify/cors';
import helmet from '@fastify/helmet';
import rateLimit from '@fastify/rate-limit';
import swagger from '@fastify/swagger';
import swaggerUi from '@fastify/swagger-ui';
import { config } from './config/index.js';
import { routes } from './api/routes/index.js';
import { errorHandler } from './api/middleware/errorHandler.js';
import { logger } from './utils/logger.js';

const app = Fastify({
  logger: logger,
  disableRequestLogging: false,
});

// Security plugins
await app.register(helmet);
await app.register(cors, {
  origin: config.corsOrigins,
  credentials: true,
});
await app.register(rateLimit, {
  max: 100,
  timeWindow: '1 minute',
});

// API documentation
await app.register(swagger, {
  openapi: {
    info: {
      title: 'My API',
      description: 'API documentation',
      version: '1.0.0',
    },
    servers: [{ url: config.apiUrl }],
    components: {
      securitySchemes: {
        bearerAuth: {
          type: 'http',
          scheme: 'bearer',
          bearerFormat: 'JWT',
        },
      },
    },
  },
});
await app.register(swaggerUi, { routePrefix: '/docs' });

// Routes
await app.register(routes, { prefix: '/api/v1' });

// Error handling
app.setErrorHandler(errorHandler);

// Health check
app.get('/health', async () => ({ status: 'ok', timestamp: new Date().toISOString() }));

// Start server
const start = async () => {
  try {
    await app.listen({ port: config.port, host: '0.0.0.0' });
    app.log.info(`Server running on http://localhost:${config.port}`);
  } catch (err) {
    app.log.error(err);
    process.exit(1);
  }
};

start();

export { app };
```

### src/config/index.js

```javascript
import { z } from 'zod';
import 'dotenv/config';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.string().transform(Number).default('3000'),
  DATABASE_URL: z.string().url(),
  JWT_SECRET: z.string().min(32),
  CORS_ORIGINS: z.string().default('http://localhost:3000'),
  API_URL: z.string().url().default('http://localhost:3000'),
  LOG_LEVEL: z.enum(['debug', 'info', 'warn', 'error']).default('info'),
});

const parsed = envSchema.safeParse(process.env);

if (!parsed.success) {
  console.error('Invalid environment variables:');
  console.error(parsed.error.flatten().fieldErrors);
  process.exit(1);
}

export const config = {
  nodeEnv: parsed.data.NODE_ENV,
  port: parsed.data.PORT,
  databaseUrl: parsed.data.DATABASE_URL,
  jwtSecret: parsed.data.JWT_SECRET,
  corsOrigins: parsed.data.CORS_ORIGINS.split(','),
  apiUrl: parsed.data.API_URL,
  logLevel: parsed.data.LOG_LEVEL,
  isDev: parsed.data.NODE_ENV === 'development',
  isProd: parsed.data.NODE_ENV === 'production',
};
```

### src/api/routes/users.js

```javascript
import { z } from 'zod';
import { userService } from '../../services/userService.js';
import { authenticate } from '../middleware/auth.js';

const createUserSchema = z.object({
  name: z.string().min(2).max(100),
  email: z.string().email(),
  password: z.string().min(8).max(100),
});

const updateUserSchema = z.object({
  name: z.string().min(2).max(100).optional(),
  email: z.string().email().optional(),
});

const idParamSchema = z.object({
  id: z.string().uuid(),
});

const querySchema = z.object({
  page: z.string().transform(Number).default('1'),
  limit: z.string().transform(Number).default('10'),
  search: z.string().optional(),
});

export async function userRoutes(fastify) {
  // List users
  fastify.get('/', {
    schema: {
      tags: ['Users'],
      querystring: querySchema,
      response: {
        200: {
          type: 'object',
          properties: {
            users: { type: 'array' },
            pagination: { type: 'object' },
          },
        },
      },
    },
    handler: async (request, reply) => {
      const query = querySchema.parse(request.query);
      const result = await userService.list(query);
      return result;
    },
  });

  // Get user by ID
  fastify.get('/:id', {
    schema: {
      tags: ['Users'],
      params: idParamSchema,
    },
    handler: async (request, reply) => {
      const { id } = idParamSchema.parse(request.params);
      const user = await userService.getById(id);
      if (!user) {
        return reply.code(404).send({ error: 'User not found' });
      }
      return user;
    },
  });

  // Create user
  fastify.post('/', {
    schema: {
      tags: ['Users'],
      body: createUserSchema,
    },
    handler: async (request, reply) => {
      const data = createUserSchema.parse(request.body);
      const user = await userService.create(data);
      return reply.code(201).send(user);
    },
  });

  // Update user (protected)
  fastify.put('/:id', {
    preHandler: [authenticate],
    schema: {
      tags: ['Users'],
      security: [{ bearerAuth: [] }],
      params: idParamSchema,
      body: updateUserSchema,
    },
    handler: async (request, reply) => {
      const { id } = idParamSchema.parse(request.params);
      const data = updateUserSchema.parse(request.body);
      const user = await userService.update(id, data);
      return user;
    },
  });

  // Delete user (protected)
  fastify.delete('/:id', {
    preHandler: [authenticate],
    schema: {
      tags: ['Users'],
      security: [{ bearerAuth: [] }],
      params: idParamSchema,
    },
    handler: async (request, reply) => {
      const { id } = idParamSchema.parse(request.params);
      await userService.delete(id);
      return reply.code(204).send();
    },
  });
}
```

### src/services/userService.js

```javascript
import { userRepository } from '../repositories/userRepository.js';
import { hashPassword, verifyPassword } from '../utils/crypto.js';

class UserService {
  async list({ page, limit, search }) {
    const offset = (page - 1) * limit;
    const [users, total] = await Promise.all([
      userRepository.findMany({ offset, limit, search }),
      userRepository.count({ search }),
    ]);

    return {
      users,
      pagination: {
        page,
        limit,
        total,
        totalPages: Math.ceil(total / limit),
      },
    };
  }

  async getById(id) {
    return userRepository.findById(id);
  }

  async create(data) {
    // Check for existing email
    const existing = await userRepository.findByEmail(data.email);
    if (existing) {
      throw new Error('Email already exists');
    }

    // Hash password
    const hashedPassword = await hashPassword(data.password);

    return userRepository.create({
      ...data,
      password: hashedPassword,
    });
  }

  async update(id, data) {
    const user = await userRepository.findById(id);
    if (!user) {
      throw new Error('User not found');
    }

    if (data.email && data.email !== user.email) {
      const existing = await userRepository.findByEmail(data.email);
      if (existing) {
        throw new Error('Email already exists');
      }
    }

    return userRepository.update(id, data);
  }

  async delete(id) {
    const user = await userRepository.findById(id);
    if (!user) {
      throw new Error('User not found');
    }
    return userRepository.delete(id);
  }

  async authenticate(email, password) {
    const user = await userRepository.findByEmail(email);
    if (!user) {
      return null;
    }

    const valid = await verifyPassword(password, user.password);
    if (!valid) {
      return null;
    }

    return user;
  }
}

export const userService = new UserService();
```

### src/repositories/userRepository.js

```javascript
import { db } from '../config/database.js';

class UserRepository {
  async findMany({ offset, limit, search }) {
    let query = `
      SELECT id, name, email, created_at, updated_at
      FROM users
    `;
    const params = [];

    if (search) {
      query += ` WHERE name ILIKE $1 OR email ILIKE $1`;
      params.push(`%${search}%`);
    }

    query += ` ORDER BY created_at DESC LIMIT $${params.length + 1} OFFSET $${params.length + 2}`;
    params.push(limit, offset);

    const result = await db.query(query, params);
    return result.rows;
  }

  async count({ search }) {
    let query = `SELECT COUNT(*) FROM users`;
    const params = [];

    if (search) {
      query += ` WHERE name ILIKE $1 OR email ILIKE $1`;
      params.push(`%${search}%`);
    }

    const result = await db.query(query, params);
    return parseInt(result.rows[0].count);
  }

  async findById(id) {
    const result = await db.query(
      `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`,
      [id]
    );
    return result.rows[0] || null;
  }

  async findByEmail(email) {
    const result = await db.query(
      `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`,
      [email]
    );
    return result.rows[0] || null;
  }

  async create(data) {
    const result = await db.query(
      `INSERT INTO users (name, email, password)
       VALUES ($1, $2, $3)
       RETURNING id, name, email, created_at, updated_at`,
      [data.name, data.email, data.password]
    );
    return result.rows[0];
  }

  async update(id, data) {
    const fields = [];
    const values = [];
    let paramIndex = 1;

    for (const [key, value] of Object.entries(data)) {
      if (value !== undefined) {
        fields.push(`${key} = $${paramIndex}`);
        values.push(value);
        paramIndex++;
      }
    }

    if (fields.length === 0) {
      return this.findById(id);
    }

    fields.push(`updated_at = NOW()`);
    values.push(id);

    const result = await db.query(
      `UPDATE users SET ${fields.join(', ')} WHERE id = $${paramIndex}
       RETURNING id, name, email, created_at, updated_at`,
      values
    );
    return result.rows[0];
  }

  async delete(id) {
    await db.query(`DELETE FROM users WHERE id = $1`, [id]);
  }
}

export const userRepository = new UserRepository();
```

---

## Testing Examples

### vitest.config.js

```javascript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['test/**/*.test.js'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.js'],
      exclude: ['src/index.js'],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 75,
        statements: 80,
      },
    },
    setupFiles: ['test/setup.js'],
  },
});
```

### test/setup.js

```javascript
import { beforeAll, afterAll, afterEach, vi } from 'vitest';

// Mock environment variables
process.env.NODE_ENV = 'test';
process.env.DATABASE_URL = 'postgres://test:test@localhost:5432/test';
process.env.JWT_SECRET = 'test-secret-key-at-least-32-characters';
process.env.PORT = '3001';

// Global test setup
beforeAll(async () => {
  // Setup test database, start test server, etc.
});

afterAll(async () => {
  // Cleanup
});

afterEach(() => {
  vi.clearAllMocks();
});
```

### test/unit/userService.test.js

```javascript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { userService } from '../../src/services/userService.js';
import { userRepository } from '../../src/repositories/userRepository.js';

// Mock repository
vi.mock('../../src/repositories/userRepository.js', () => ({
  userRepository: {
    findMany: vi.fn(),
    count: vi.fn(),
    findById: vi.fn(),
    findByEmail: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn(),
  },
}));

// Mock crypto utilities
vi.mock('../../src/utils/crypto.js', () => ({
  hashPassword: vi.fn((p) => `hashed_${p}`),
  verifyPassword: vi.fn((plain, hashed) => hashed === `hashed_${plain}`),
}));

describe('UserService', () => {
  const mockUser = {
    id: '123e4567-e89b-12d3-a456-426614174000',
    name: 'John Doe',
    email: 'john@example.com',
    created_at: new Date(),
    updated_at: new Date(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('list', () => {
    it('returns paginated users', async () => {
      userRepository.findMany.mockResolvedValue([mockUser]);
      userRepository.count.mockResolvedValue(1);

      const result = await userService.list({ page: 1, limit: 10 });

      expect(result.users).toHaveLength(1);
      expect(result.pagination).toEqual({
        page: 1,
        limit: 10,
        total: 1,
        totalPages: 1,
      });
    });

    it('calculates offset correctly', async () => {
      userRepository.findMany.mockResolvedValue([]);
      userRepository.count.mockResolvedValue(0);

      await userService.list({ page: 3, limit: 20 });

      expect(userRepository.findMany).toHaveBeenCalledWith({
        offset: 40,
        limit: 20,
        search: undefined,
      });
    });
  });

  describe('create', () => {
    it('creates user with hashed password', async () => {
      userRepository.findByEmail.mockResolvedValue(null);
      userRepository.create.mockResolvedValue(mockUser);

      const result = await userService.create({
        name: 'John Doe',
        email: 'john@example.com',
        password: 'password123',
      });

      expect(userRepository.create).toHaveBeenCalledWith({
        name: 'John Doe',
        email: 'john@example.com',
        password: 'hashed_password123',
      });
      expect(result).toEqual(mockUser);
    });

    it('throws on duplicate email', async () => {
      userRepository.findByEmail.mockResolvedValue(mockUser);

      await expect(
        userService.create({
          name: 'Jane',
          email: 'john@example.com',
          password: 'password',
        })
      ).rejects.toThrow('Email already exists');
    });
  });

  describe('authenticate', () => {
    it('returns user on valid credentials', async () => {
      const userWithPassword = { ...mockUser, password: 'hashed_password123' };
      userRepository.findByEmail.mockResolvedValue(userWithPassword);

      const result = await userService.authenticate('john@example.com', 'password123');

      expect(result).toEqual(userWithPassword);
    });

    it('returns null on invalid password', async () => {
      const userWithPassword = { ...mockUser, password: 'hashed_password123' };
      userRepository.findByEmail.mockResolvedValue(userWithPassword);

      const result = await userService.authenticate('john@example.com', 'wrongpassword');

      expect(result).toBeNull();
    });

    it('returns null on non-existent user', async () => {
      userRepository.findByEmail.mockResolvedValue(null);

      const result = await userService.authenticate('unknown@example.com', 'password');

      expect(result).toBeNull();
    });
  });
});
```

### test/integration/users.test.js

```javascript
import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { app } from '../../src/index.js';
import { db } from '../../src/config/database.js';

describe('Users API', () => {
  beforeAll(async () => {
    await app.ready();
  });

  afterAll(async () => {
    await app.close();
    await db.end();
  });

  beforeEach(async () => {
    // Clean up database
    await db.query('DELETE FROM users');
  });

  describe('POST /api/v1/users', () => {
    it('creates a new user', async () => {
      const response = await app.inject({
        method: 'POST',
        url: '/api/v1/users',
        payload: {
          name: 'John Doe',
          email: 'john@example.com',
          password: 'password123',
        },
      });

      expect(response.statusCode).toBe(201);
      const body = JSON.parse(response.body);
      expect(body).toMatchObject({
        name: 'John Doe',
        email: 'john@example.com',
      });
      expect(body.id).toBeDefined();
      expect(body.password).toBeUndefined();
    });

    it('returns 400 for invalid email', async () => {
      const response = await app.inject({
        method: 'POST',
        url: '/api/v1/users',
        payload: {
          name: 'John Doe',
          email: 'invalid-email',
          password: 'password123',
        },
      });

      expect(response.statusCode).toBe(400);
    });
  });

  describe('GET /api/v1/users', () => {
    it('returns paginated users', async () => {
      // Create test users
      await db.query(
        `INSERT INTO users (name, email, password) VALUES
         ('User 1', 'user1@test.com', 'hashed'),
         ('User 2', 'user2@test.com', 'hashed')`
      );

      const response = await app.inject({
        method: 'GET',
        url: '/api/v1/users?page=1&limit=10',
      });

      expect(response.statusCode).toBe(200);
      const body = JSON.parse(response.body);
      expect(body.users).toHaveLength(2);
      expect(body.pagination.total).toBe(2);
    });
  });
});
```

---

## Hono Edge Example

### src/index.js (Edge/Serverless)

```javascript
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';
import { prettyJSON } from 'hono/pretty-json';
import { compress } from 'hono/compress';
import { cache } from 'hono/cache';
import { jwt } from 'hono/jwt';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';

const app = new Hono();

// Middleware
app.use('*', logger());
app.use('*', secureHeaders());
app.use('*', compress());
app.use('/api/*', cors());
app.use('/api/*', prettyJSON());

// Cache static responses
app.get('/api/health', cache({ cacheName: 'health', cacheControl: 'max-age=60' }));

// Health check
app.get('/api/health', (c) => c.json({ status: 'ok', timestamp: Date.now() }));

// JWT protected routes
app.use('/api/protected/*', jwt({ secret: Bun.env.JWT_SECRET }));

// Validation schemas
const createPostSchema = z.object({
  title: z.string().min(1).max(200),
  content: z.string().min(1),
  tags: z.array(z.string()).optional(),
});

// Routes with validation
app.post('/api/posts',
  jwt({ secret: Bun.env.JWT_SECRET }),
  zValidator('json', createPostSchema),
  async (c) => {
    const payload = c.get('jwtPayload');
    const data = c.req.valid('json');

    // Create post logic
    const post = {
      id: crypto.randomUUID(),
      ...data,
      authorId: payload.sub,
      createdAt: new Date().toISOString(),
    };

    return c.json(post, 201);
  }
);

app.get('/api/posts', async (c) => {
  const page = Number(c.req.query('page') || '1');
  const limit = Number(c.req.query('limit') || '10');

  // Fetch posts logic
  return c.json({
    posts: [],
    pagination: { page, limit, total: 0 },
  });
});

// Error handling
app.onError((err, c) => {
  console.error(`${err}`);
  return c.json({ error: err.message }, 500);
});

app.notFound((c) => c.json({ error: 'Not Found' }, 404));

// Export for different runtimes
export default app;

// Bun
// Bun.serve({ port: 3000, fetch: app.fetch });

// Node.js
// import { serve } from '@hono/node-server';
// serve({ fetch: app.fetch, port: 3000 });
```

---

## ESLint 9 Flat Config

### eslint.config.js

```javascript
import js from '@eslint/js';
import globals from 'globals';

export default [
  {
    ignores: ['dist/', 'node_modules/', 'coverage/'],
  },
  js.configs.recommended,
  {
    files: ['**/*.js'],
    languageOptions: {
      ecmaVersion: 2025,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2025,
      },
    },
    rules: {
      // Best practices
      'no-unused-vars': ['error', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      'no-console': ['warn', { allow: ['warn', 'error', 'info'] }],
      'prefer-const': 'error',
      'no-var': 'error',
      'object-shorthand': 'error',
      'prefer-template': 'error',
      'prefer-arrow-callback': 'error',

      // Code style
      'arrow-body-style': ['error', 'as-needed'],
      'no-multiple-empty-lines': ['error', { max: 1 }],
      'eol-last': ['error', 'always'],

      // Error prevention
      'no-return-await': 'error',
      'require-await': 'error',
      'no-async-promise-executor': 'error',
      'no-promise-executor-return': 'error',
    },
  },
  {
    files: ['test/**/*.js'],
    languageOptions: {
      globals: {
        ...globals.vitest,
      },
    },
    rules: {
      'no-unused-expressions': 'off',
    },
  },
];
```

---

## Dockerfile

```dockerfile
# Build stage
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Production stage
FROM node:22-alpine
WORKDIR /app

# Security: run as non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

COPY --from=builder /app/node_modules ./node_modules
COPY . .

# Set ownership
RUN chown -R nodejs:nodejs /app

USER nodejs

ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["node", "src/index.js"]
```

---

Last Updated: 2025-12-22
Version: 1.0.0
