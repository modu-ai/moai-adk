---
name: moai-lang-javascript
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "JavaScript best practices with Node.js, modern frameworks, and ecosystem patterns for 2025"
keywords: [javascript, nodejs, frontend, backend, legacy, maintenance, modernization, browser]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# JavaScript Development Mastery

**Modern JavaScript Development with 2025 Best Practices**

> Comprehensive JavaScript development guidance covering Node.js backend development, legacy browser support, modern framework integration, and progressive enhancement patterns using the latest tools and methodologies.

## What It Does

- **Node.js Backend Services**: Express, Koa, and modern server-side JavaScript
- **Legacy Browser Support**: Progressive enhancement and polyfill strategies
- **API Development**: REST, GraphQL, and real-time communication
- **CLI Tools & Automation**: Node.js-based developer utilities
- **Modern Build Systems**: Webpack, Vite, and bundling optimization
- **Testing & Quality**: Unit testing, integration testing, and code quality
- **Migration Strategies**: Legacy code modernization and TypeScript migration

## When to Use

### Perfect Scenarios
- **Legacy JavaScript application maintenance and modernization**
- **Node.js backend services and APIs**
- **CLI tools and developer utilities**
- **Progressive web applications requiring broad browser support**
- **Projects needing gradual TypeScript migration**
- **Serverless functions and edge computing**
- **Real-time applications with WebSockets**

### Common Triggers
- "Create Node.js API"
- "Modernize JavaScript legacy code"
- "Set up Express server"
- "JavaScript best practices"
- "Migrate to TypeScript"
- "Browser compatibility issues"

## Tool Version Matrix (2025-11-06)

### Core JavaScript
- **Node.js**: 22.x (LTS) / 20.x (Active LTS)
- **npm**: 10.x - Node package manager
- **Yarn**: 4.x - Alternative package manager
- **pnpm**: 9.x - Fast, disk space efficient package manager

### Backend Frameworks
- **Express**: 4.21.x / 5.1.x (new default) - Web framework
- **Koa**: 2.15.x - Next generation web framework
- **Fastify**: 5.x - Fast and low overhead web framework
- **NestJS**: 10.x - Progressive Node.js framework
- **Hapi**: 21.x - Rich framework for building applications

### Frontend Integration
- **React**: 19.x - UI library (for integration patterns)
- **Vue**: 3.5.x - Progressive framework
- **Webpack**: 5.x - Module bundler
- **Vite**: 6.x - Fast build tool
- **Rollup**: 4.x - Module bundler for libraries

### Database & Storage
- **Prisma**: 5.22.x - Next-generation ORM
- **Sequelize**: 6.37.x - SQL ORM
- **Mongoose**: 8.8.x - MongoDB ODM
- **Redis**: 4.7.x - Redis client
- **LowDB**: 7.x - Small JSON database

### Testing Tools
- **Jest**: 30.x - JavaScript testing framework
- **Mocha**: 10.x - Feature-rich test framework
- **Chai**: 5.x - BDD/TDD assertion library
- **Supertest**: 7.x - HTTP assertion testing
- **Playwright**: 1.48.x - End-to-end testing

### Development Tools
- **ESLint**: 9.x - Pluggable linter
- **Prettier**: 3.3.x - Code formatter
- **Babel**: 7.26.x - JavaScript compiler
- **Nodemon**: 3.1.x - Monitor for changes and restart
- **PM2**: 5.4.x - Production process manager

## Ecosystem Overview

### Project Setup (2025 Best Practice)

```bash
# Modern Node.js project with package.json
npm init -y
npm install express helmet cors compression morgan dotenv
npm install -D nodemon jest supertest eslint prettier concurrently

# Project structure
mkdir -p {src/{routes,middleware,services,utils,models},test/{unit,integration},config,scripts}

# TypeScript support for gradual migration
npm install -D typescript @types/node @types/express ts-node

# Modern build setup
npm install -D vite @vitejs/plugin-node
```

### Modern Project Structure

```
my-javascript-project/
├── package.json              # Package configuration
├── package-lock.json         # Lock file
├── .eslintrc.js              # ESLint configuration
├── .prettierrc               # Prettier configuration
├── .gitignore
├── README.md
├── jest.config.js            # Jest testing configuration
├── nodemon.json              # Nodemon configuration
├── src/
│   ├── app.js                # Application entry point
│   ├── routes/               # API routes
│   │   ├── index.js
│   │   ├── users.js
│   │   └── auth.js
│   ├── middleware/           # Express middleware
│   │   ├── auth.js
│   │   ├── validation.js
│   │   └── errorHandler.js
│   ├── services/             # Business logic
│   │   ├── userService.js
│   │   ├── emailService.js
│   │   └── cacheService.js
│   ├── models/               # Data models
│   │   ├── User.js
│   │   └── index.js
│   ├── utils/                # Utility functions
│   │   ├── logger.js
│   │   ├── validator.js
│   │   └── helpers.js
│   └── config/               # Configuration
│       ├── database.js
│       └── index.js
├── test/
│   ├── unit/                 # Unit tests
│   ├── integration/          # Integration tests
│   └── fixtures/             # Test data
├── scripts/                  # Build and utility scripts
├── docs/                     # Documentation
└── dist/                     # Build output
```

## Modern JavaScript Patterns

### Async Programming with Modern Patterns

```javascript
// src/services/userService.js
const { promisify } = require('util');
const crypto = require('crypto');
const EventEmitter = require('events');

class UserService extends EventEmitter {
  constructor(database) {
    super();
    this.db = database;
    this.cache = new Map();
  }

  // Async/await pattern for database operations
  async createUser(userData) {
    try {
      // Validate input
      this.validateUserData(userData);

      // Hash password
      const hashedPassword = await this.hashPassword(userData.password);
      
      // Create user object
      const user = {
        id: crypto.randomUUID(),
        email: userData.email,
        name: userData.name,
        password: hashedPassword,
        createdAt: new Date(),
        updatedAt: new Date()
      };

      // Save to database
      const savedUser = await this.db.collection('users').insertOne(user);
      
      // Emit event
      this.emit('userCreated', savedUser);
      
      // Remove password before returning
      const { password, ...userResponse } = savedUser;
      
      return userResponse;
    } catch (error) {
      this.emit('error', error);
      throw new Error(`Failed to create user: ${error.message}`);
    }
  }

  async getUserById(userId) {
    // Check cache first
    if (this.cache.has(userId)) {
      return this.cache.get(userId);
    }

    try {
      const user = await this.db.collection('users').findOne({ id: userId });
      
      if (!user) {
        throw new Error('User not found');
      }

      // Cache result for 5 minutes
      this.cache.set(userId, user);
      setTimeout(() => this.cache.delete(userId), 5 * 60 * 1000);

      const { password, ...userResponse } = user;
      return userResponse;
    } catch (error) {
      this.emit('error', error);
      throw error;
    }
  }

  // Batch processing with Promise.allSettled
  async processUsersBatch(userIds) {
    const userPromises = userIds.map(id => 
      this.getUserById(id).catch(error => ({ id, error: error.message }))
    );

    const results = await Promise.allSettled(userPromises);
    
    return results.map((result, index) => ({
      userId: userIds[index],
      status: result.status,
      value: result.status === 'fulfilled' ? result.value : null,
      reason: result.status === 'rejected' ? result.reason : null
    }));
  }

  // Async iterator pattern for streaming large datasets
  async *getAllUsersStream() {
    let cursor;
    try {
      cursor = await this.db.collection('users').find();
      
      while (await cursor.hasNext()) {
        const user = await cursor.next();
        const { password, ...userResponse } = user;
        yield userResponse;
      }
    } finally {
      if (cursor) {
        await cursor.close();
      }
    }
  }

  // Promise-based helper methods
  async hashPassword(password) {
    const scrypt = promisify(crypto.scrypt);
    const salt = crypto.randomBytes(16).toString('hex');
    const derivedKey = await scrypt(password, salt, 64);
    return `${salt}:${derivedKey.toString('hex')}`;
  }

  validateUserData(userData) {
    const errors = [];
    
    if (!userData.email || !this.isValidEmail(userData.email)) {
      errors.push('Valid email is required');
    }
    
    if (!userData.name || userData.name.length < 2) {
      errors.push('Name must be at least 2 characters');
    }
    
    if (!userData.password || userData.password.length < 8) {
      errors.push('Password must be at least 8 characters');
    }
    
    if (errors.length > 0) {
      throw new Error(errors.join(', '));
    }
  }

  isValidEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }
}

module.exports = UserService;
```

### Express.js Application with Modern Middleware

```javascript
// src/app.js
const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const compression = require('compression');
const morgan = require('morgan');
const rateLimit = require('express-rate-limit');
const { createProxyMiddleware } = require('http-proxy-middleware');

const userRoutes = require('./routes/users');
const authRoutes = require('./routes/auth');
const errorHandler = require('./middleware/errorHandler');
const logger = require('./utils/logger');
const config = require('./config');

const app = express();

// Security middleware
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
}));

// CORS configuration
app.use(cors({
  origin: config.cors.origins,
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization']
}));

// Rate limiting
const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: config.rateLimit.max, // limit each IP to 100 requests per windowMs
  message: {
    error: 'Too many requests from this IP, please try again later.'
  },
  standardHeaders: true,
  legacyHeaders: false,
});

app.use('/api/', limiter);

// Body parsing middleware
app.use(express.json({ 
  limit: '10mb',
  verify: (req, res, buf) => {
    req.rawBody = buf;
  }
}));

app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// Compression middleware
app.use(compression());

// Logging middleware
app.use(morgan('combined', {
  stream: {
    write: (message) => logger.info(message.trim())
  }
}));

// Health check endpoint
app.get('/health', (req, res) => {
  res.status(200).json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    version: process.env.npm_package_version
  });
});

// API routes
app.use('/api/users', userRoutes);
app.use('/api/auth', authRoutes);

// API documentation proxy (if using Swagger)
if (config.env === 'development') {
  app.use('/api-docs', createProxyMiddleware({
    target: 'http://localhost:3001',
    changeOrigin: true,
    pathRewrite: {
      '^/api-docs': ''
    }
  }));
}

// 404 handler
app.use('*', (req, res) => {
  res.status(404).json({
    error: 'Route not found',
    path: req.originalUrl,
    method: req.method
  });
});

// Error handling middleware (must be last)
app.use(errorHandler);

module.exports = app;
```

### Modern Route Handlers

```javascript
// src/routes/users.js
const express = require('express');
const { body, param, query, validationResult } = require('express-validator');
const UserService = require('../services/userService');
const auth = require('../middleware/auth');
const logger = require('../utils/logger');

const router = express.Router();
const userService = new UserService(database); // Assuming database is available

// Validation middleware
const handleValidationErrors = (req, res, next) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    return res.status(400).json({
      error: 'Validation failed',
      details: errors.array()
    });
  }
  next();
};

// POST /api/users - Create user
router.post('/',
  [
    body('email').isEmail().normalizeEmail(),
    body('name').isLength({ min: 2, max: 100 }).trim().escape(),
    body('password').isLength({ min: 8 }).matches(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/),
  ],
  handleValidationErrors,
  async (req, res, next) => {
    try {
      const user = await userService.createUser(req.body);
      
      res.status(201).json({
        success: true,
        data: user,
        message: 'User created successfully'
      });
      
      logger.info(`User created: ${user.id}`);
    } catch (error) {
      next(error);
    }
  }
);

// GET /api/users - Get users with pagination and filtering
router.get('/',
  [
    query('page').optional().isInt({ min: 1 }).toInt(),
    query('limit').optional().isInt({ min: 1, max: 100 }).toInt(),
    query('search').optional().isLength({ min: 1, max: 100 }).trim().escape(),
  ],
  handleValidationErrors,
  async (req, res, next) => {
    try {
      const { page = 1, limit = 10, search } = req.query;
      const skip = (page - 1) * limit;
      
      // Build filter
      const filter = {};
      if (search) {
        filter.$or = [
          { name: { $regex: search, $options: 'i' } },
          { email: { $regex: search, $options: 'i' } }
        ];
      }
      
      // Get users
      const users = await userService.getUsers(filter, { skip, limit });
      const total = await userService.countUsers(filter);
      
      res.json({
        success: true,
        data: {
          users,
          pagination: {
            page,
            limit,
            total,
            pages: Math.ceil(total / limit)
          }
        }
      });
    } catch (error) {
      next(error);
    }
  }
);

// GET /api/users/:id - Get user by ID
router.get('/:id',
  [
    param('id').isUUID().withMessage('Invalid user ID format')
  ],
  handleValidationErrors,
  async (req, res, next) => {
    try {
      const user = await userService.getUserById(req.params.id);
      
      if (!user) {
        return res.status(404).json({
          error: 'User not found',
          userId: req.params.id
        });
      }
      
      res.json({
        success: true,
        data: user
      });
    } catch (error) {
      next(error);
    }
  }
);

// PUT /api/users/:id - Update user
router.put('/:id',
  auth.authenticate,
  [
    param('id').isUUID().withMessage('Invalid user ID format'),
    body('name').optional().isLength({ min: 2, max: 100 }).trim().escape(),
    body('email').optional().isEmail().normalizeEmail(),
  ],
  handleValidationErrors,
  async (req, res, next) => {
    try {
      // Check if user is updating their own profile or is admin
      if (req.user.id !== req.params.id && req.user.role !== 'admin') {
        return res.status(403).json({
          error: 'Forbidden: You can only update your own profile'
        });
      }
      
      const updateData = {};
      if (req.body.name) updateData.name = req.body.name;
      if (req.body.email) updateData.email = req.body.email;
      
      const user = await userService.updateUser(req.params.id, updateData);
      
      res.json({
        success: true,
        data: user,
        message: 'User updated successfully'
      });
    } catch (error) {
      next(error);
    }
  }
);

// DELETE /api/users/:id - Delete user
router.delete('/:id',
  auth.authenticate,
  auth.requireRole('admin'),
  [
    param('id').isUUID().withMessage('Invalid user ID format')
  ],
  handleValidationErrors,
  async (req, res, next) => {
    try {
      await userService.deleteUser(req.params.id);
      
      res.json({
        success: true,
        message: 'User deleted successfully'
      });
    } catch (error) {
      next(error);
    }
  }
);

module.exports = router;
```

## Performance Optimization

### Caching and Memory Management

```javascript
// src/services/cacheService.js
const NodeCache = require('node-cache');
const redis = require('redis');
const crypto = require('crypto');

class CacheService {
  constructor(options = {}) {
    // In-memory cache for frequently accessed data
    this.memoryCache = new NodeCache({
      stdTTL: options.memoryTTL || 300, // 5 minutes
      checkperiod: options.checkPeriod || 60, // 1 minute
      useClones: false // Improve performance for objects
    });

    // Redis cache for distributed caching
    if (options.redis) {
      this.redisClient = redis.createClient(options.redis);
      this.redisClient.on('error', (err) => {
        console.error('Redis client error:', err);
      });
    }
  }

  // Multi-level caching
  async get(key) {
    // Check memory cache first
    let value = this.memoryCache.get(key);
    if (value !== undefined) {
      return value;
    }

    // Check Redis cache
    if (this.redisClient) {
      try {
        value = await this.redisClient.get(key);
        if (value) {
          const parsed = JSON.parse(value);
          // Store in memory cache for faster access
          this.memoryCache.set(key, parsed);
          return parsed;
        }
      } catch (error) {
        console.error('Redis get error:', error);
      }
    }

    return null;
  }

  async set(key, value, options = {}) {
    const serialized = JSON.stringify(value);

    // Set in memory cache
    this.memoryCache.set(key, value, options.ttl);

    // Set in Redis cache
    if (this.redisClient) {
      try {
        await this.redisClient.setEx(key, options.ttl || 300, serialized);
      } catch (error) {
        console.error('Redis set error:', error);
      }
    }
  }

  // Cache warming
  async warmup(dataLoader, keys) {
    const promises = keys.map(async (key) => {
      const value = await dataLoader(key);
      await this.set(key, value);
      return { key, value };
    });

    return Promise.all(promises);
  }

  // Cache invalidation with tags
  async invalidateByPattern(pattern) {
    // Invalidate memory cache
    const keys = this.memoryCache.keys();
    const regex = new RegExp(pattern);
    
    keys.forEach(key => {
      if (regex.test(key)) {
        this.memoryCache.del(key);
      }
    });

    // Invalidate Redis cache
    if (this.redisClient) {
      try {
        const redisKeys = await this.redisClient.keys(pattern);
        if (redisKeys.length > 0) {
          await this.redisClient.del(redisKeys);
        }
      } catch (error) {
        console.error('Redis pattern deletion error:', error);
      }
    }
  }

  generateCacheKey(prefix, params) {
    const sortedParams = Object.keys(params)
      .sort()
      .reduce((result, key) => {
        result[key] = params[key];
        return result;
      }, {});

    const paramString = JSON.stringify(sortedParams);
    const hash = crypto.createHash('md5').update(paramString).digest('hex');
    
    return `${prefix}:${hash}`;
  }

  // Memory leak prevention
  cleanup() {
    this.memoryCache.flushAll();
    
    if (this.redisClient) {
      this.redisClient.quit();
    }
  }
}

module.exports = CacheService;
```

### Memory-Efficient Data Processing

```javascript
// src/utils/streamProcessor.js
const { Transform, pipeline } = require('stream');
const fs = require('fs');
const readline = require('readline');
const { promisify } = require('util');

const pipelineAsync = promisify(pipeline);

class StreamProcessor {
  // Process large files without loading into memory
  static async processLargeFile(filePath, processor) {
    const fileStream = fs.createReadStream(filePath);
    const rl = readline.createInterface({
      input: fileStream,
      crlfDelay: Infinity
    });

    let lineNumber = 0;
    const results = [];

    for await (const line of rl) {
      try {
        const result = await processor(line, lineNumber);
        if (result !== null) {
          results.push(result);
        }
      } catch (error) {
        console.error(`Error processing line ${lineNumber}:`, error);
      }
      
      lineNumber++;
      
      // Yield control periodically to prevent blocking
      if (lineNumber % 1000 === 0) {
        await new Promise(resolve => setImmediate(resolve));
      }
    }

    return results;
  }

  // Transform stream for data processing
  static createTransformStream(processor) {
    return new Transform({
      objectMode: true,
      transform(chunk, encoding, callback) {
        try {
          const result = processor(chunk);
          callback(null, result);
        } catch (error) {
          callback(error);
        }
      }
    });
  }

  // Batch processing with memory management
  static async* batchProcessor(data, batchSize = 100) {
    for (let i = 0; i < data.length; i += batchSize) {
      const batch = data.slice(i, i + batchSize);
      yield batch;
      
      // Force garbage collection every 10 batches
      if (i % (batchSize * 10) === 0) {
        if (global.gc) {
          global.gc();
        }
      }
    }
  }

  // Memory-efficient CSV processing
  static async processCSV(filePath, options = {}) {
    const { transform, filter, batchSize = 1000 } = options;
    const results = [];
    let currentBatch = [];

    return new Promise((resolve, reject) => {
      const fileStream = fs.createReadStream(filePath);
      const rl = readline.createInterface({
        input: fileStream,
        crlfDelay: Infinity
      });

      let lineNumber = 0;
      let headers = [];

      rl.on('line', async (line) => {
        try {
          const values = line.split(',');
          
          if (lineNumber === 0) {
            headers = values;
          } else {
            const record = {};
            headers.forEach((header, index) => {
              record[header.trim()] = values[index]?.trim() || '';
            });

            // Apply filter if provided
            if (filter && !filter(record)) {
              return;
            }

            // Apply transform if provided
            const processedRecord = transform ? transform(record) : record;
            
            currentBatch.push(processedRecord);

            // Process batch when full
            if (currentBatch.length >= batchSize) {
              results.push(...currentBatch);
              currentBatch = [];
              
              // Yield control
              await new Promise(resolve => setImmediate(resolve));
            }
          }
        } catch (error) {
          console.error(`Error processing line ${lineNumber}:`, error);
        }
        
        lineNumber++;
      });

      rl.on('close', () => {
        // Process remaining records
        results.push(...currentBatch);
        resolve(results);
      });

      rl.on('error', reject);
    });
  }
}

module.exports = StreamProcessor;
```

## Testing Strategies

### Comprehensive Unit Testing

```javascript
// test/unit/userService.test.js
const UserService = require('../../src/services/userService');
const { EventEmitter } = require('events');

// Mock database
const mockDatabase = {
  collection: jest.fn(() => ({
    insertOne: jest.fn(),
    findOne: jest.fn(),
    find: jest.fn(),
    updateOne: jest.fn(),
    deleteOne: jest.fn(),
  }))
};

describe('UserService', () => {
  let userService;

  beforeEach(() => {
    userService = new UserService(mockDatabase);
    jest.clearAllMocks();
  });

  describe('createUser', () => {
    it('should create a user with valid data', async () => {
      const userData = {
        email: 'test@example.com',
        name: 'Test User',
        password: 'Password123'
      };

      const mockUser = {
        id: 'user-123',
        email: userData.email,
        name: userData.name,
        password: 'hashed-password',
        createdAt: new Date(),
        updatedAt: new Date()
      };

      mockDatabase.collection().insertOne.mockResolvedValue(mockUser);

      const result = await userService.createUser(userData);

      expect(result).toEqual({
        id: mockUser.id,
        email: mockUser.email,
        name: mockUser.name,
        createdAt: mockUser.createdAt,
        updatedAt: mockUser.updatedAt
      });

      expect(mockDatabase.collection).toHaveBeenCalledWith('users');
      expect(mockDatabase.collection().insertOne).toHaveBeenCalledWith(
        expect.objectContaining({
          email: userData.email,
          name: userData.name,
          password: expect.stringMatching(/^[a-f0-9]+:.+$/)
        })
      );
    });

    it('should reject invalid email', async () => {
      const userData = {
        email: 'invalid-email',
        name: 'Test User',
        password: 'Password123'
      };

      await expect(userService.createUser(userData))
        .rejects.toThrow('Valid email is required');
    });

    it('should reject short password', async () => {
      const userData = {
        email: 'test@example.com',
        name: 'Test User',
        password: 'short'
      };

      await expect(userService.createUser(userData))
        .rejects.toThrow('Password must be at least 8 characters');
    });

    it('should emit userCreated event', async () => {
      const userData = {
        email: 'test@example.com',
        name: 'Test User',
        password: 'Password123'
      };

      const mockUser = {
        id: 'user-123',
        email: userData.email,
        name: userData.name,
        password: 'hashed-password',
        createdAt: new Date(),
        updatedAt: new Date()
      };

      mockDatabase.collection().insertOne.mockResolvedValue(mockUser);

      const eventSpy = jest.fn();
      userService.on('userCreated', eventSpy);

      await userService.createUser(userData);

      expect(eventSpy).toHaveBeenCalledWith(mockUser);
    });
  });

  describe('getUserById', () => {
    it('should return user when found', async () => {
      const userId = 'user-123';
      const mockUser = {
        id: userId,
        email: 'test@example.com',
        name: 'Test User',
        password: 'hashed-password'
      };

      mockDatabase.collection().findOne.mockResolvedValue(mockUser);

      const result = await userService.getUserById(userId);

      expect(result).toEqual({
        id: mockUser.id,
        email: mockUser.email,
        name: mockUser.name
      });
    });

    it('should return null when user not found', async () => {
      const userId = 'nonexistent-user';

      mockDatabase.collection().findOne.mockResolvedValue(null);

      const result = await userService.getUserById(userId);

      expect(result).toBeNull();
    });

    it('should use cache for subsequent calls', async () => {
      const userId = 'user-123';
      const mockUser = {
        id: userId,
        email: 'test@example.com',
        name: 'Test User',
        password: 'hashed-password'
      };

      mockDatabase.collection().findOne.mockResolvedValue(mockUser);

      // First call
      await userService.getUserById(userId);
      // Second call (should use cache)
      await userService.getUserById(userId);

      // Database should be called only once
      expect(mockDatabase.collection().findOne).toHaveBeenCalledTimes(1);
    });
  });

  describe('processUsersBatch', () => {
    it('should process users successfully', async () => {
      const userIds = ['user-1', 'user-2', 'user-3'];

      mockDatabase.collection().findOne.mockImplementation((query) => {
        const user = {
          id: query.id,
          email: `${query.id}@example.com`,
          name: `User ${query.id}`,
          password: 'hashed-password'
        };
        return Promise.resolve(user);
      });

      const results = await userService.processUsersBatch(userIds);

      expect(results).toHaveLength(3);
      expect(results[0]).toEqual({
        userId: 'user-1',
        status: 'fulfilled',
        value: expect.objectContaining({ id: 'user-1' }),
        reason: null
      });
    });

    it('should handle errors gracefully', async () => {
      const userIds = ['user-1', 'invalid-user'];

      mockDatabase.collection().findOne.mockImplementation((query) => {
        if (query.id === 'invalid-user') {
          return Promise.reject(new Error('Database error'));
        }
        return Promise.resolve({
          id: query.id,
          email: `${query.id}@example.com`,
          name: `User ${query.id}`,
          password: 'hashed-password'
        });
      });

      const results = await userService.processUsersBatch(userIds);

      expect(results).toHaveLength(2);
      expect(results[0].status).toBe('fulfilled');
      expect(results[1].status).toBe('rejected');
      expect(results[1].reason).toContain('Database error');
    });
  });

  describe('getAllUsersStream', () => {
    it('should stream users one by one', async () => {
      const mockUsers = [
        { id: 'user-1', email: 'user1@example.com', name: 'User 1', password: 'hash1' },
        { id: 'user-2', email: 'user2@example.com', name: 'User 2', password: 'hash2' }
      ];

      const mockCursor = {
        hasNext: jest.fn()
          .mockResolvedValueOnce(true)
          .mockResolvedValueOnce(true)
          .mockResolvedValueOnce(false),
        next: jest.fn()
          .mockResolvedValueOnce(mockUsers[0])
          .mockResolvedValueOnce(mockUsers[1]),
        close: jest.fn().mockResolvedValue()
      };

      mockDatabase.collection().find.mockResolvedValue(mockCursor);

      const users = [];
      for await (const user of userService.getAllUsersStream()) {
        users.push(user);
      }

      expect(users).toHaveLength(2);
      expect(users[0]).toEqual({
        id: 'user-1',
        email: 'user1@example.com',
        name: 'User 1'
      });
      expect(mockCursor.close).toHaveBeenCalled();
    });
  });
});
```

### Integration Testing with Test Databases

```javascript
// test/integration/api.test.js
const request = require('supertest');
const app = require('../../src/app');
const { MongoMemoryServer } = require('mongodb-memory-server');
const { MongoClient } = require('mongodb');

describe('User API Integration Tests', () => {
  let mongoServer;
  let client;
  let database;

  beforeAll(async () => {
    // Start in-memory MongoDB server
    mongoServer = await MongoMemoryServer.create();
    const mongoUri = mongoServer.getUri();
    
    // Connect to test database
    client = new MongoClient(mongoUri);
    await client.connect();
    database = client.db('testdb');
    
    // Mock database in app
    require('../../src/config/database').setDatabase(database);
  });

  afterAll(async () => {
    await client.close();
    await mongoServer.stop();
  });

  beforeEach(async () => {
    // Clean up database before each test
    await database.collection('users').deleteMany({});
  });

  describe('POST /api/users', () => {
    it('should create a new user', async () => {
      const userData = {
        email: 'integration@example.com',
        name: 'Integration User',
        password: 'Password123'
      };

      const response = await request(app)
        .post('/api/users')
        .send(userData)
        .expect(201);

      expect(response.body.success).toBe(true);
      expect(response.body.data.email).toBe(userData.email);
      expect(response.body.data.name).toBe(userData.name);
      expect(response.body.data.password).toBeUndefined(); // Password should not be returned

      // Verify user was actually created in database
      const userInDb = await database.collection('users').findOne({ email: userData.email });
      expect(userInDb).toBeTruthy();
      expect(userInDb.name).toBe(userData.name);
    });

    it('should return validation error for invalid data', async () => {
      const invalidUserData = {
        email: 'invalid-email',
        name: 'A', // Too short
        password: '123' // Too short
      };

      const response = await request(app)
        .post('/api/users')
        .send(invalidUserData)
        .expect(400);

      expect(response.body.error).toBe('Validation failed');
      expect(response.body.details).toHaveLength(3);
    });

    it('should return error for duplicate email', async () => {
      const userData = {
        email: 'duplicate@example.com',
        name: 'User One',
        password: 'Password123'
      };

      // Create first user
      await request(app)
        .post('/api/users')
        .send(userData)
        .expect(201);

      // Try to create user with same email
      const response = await request(app)
        .post('/api/users')
        .send({
          ...userData,
          name: 'User Two'
        })
        .expect(500);

      expect(response.body.error).toContain('duplicate');
    });
  });

  describe('GET /api/users', () => {
    beforeEach(async () => {
      // Insert test data
      await database.collection('users').insertMany([
        {
          id: 'user-1',
          email: 'user1@example.com',
          name: 'User One',
          password: 'hashed-pass',
          createdAt: new Date()
        },
        {
          id: 'user-2',
          email: 'user2@example.com',
          name: 'User Two',
          password: 'hashed-pass',
          createdAt: new Date()
        }
      ]);
    });

    it('should return paginated users', async () => {
      const response = await request(app)
        .get('/api/users?page=1&limit=10')
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data.users).toHaveLength(2);
      expect(response.body.data.pagination).toEqual({
        page: 1,
        limit: 10,
        total: 2,
        pages: 1
      });
    });

    it('should search users by name', async () => {
      const response = await request(app)
        .get('/api/users?search=One')
        .expect(200);

      expect(response.body.data.users).toHaveLength(1);
      expect(response.body.data.users[0].name).toBe('User One');
    });
  });

  describe('GET /api/users/:id', () => {
    let userId;

    beforeEach(async () => {
      // Create a test user
      const result = await database.collection('users').insertOne({
        id: 'test-user-123',
        email: 'getuser@example.com',
        name: 'Get User',
        password: 'hashed-pass',
        createdAt: new Date()
      });
      userId = 'test-user-123';
    });

    it('should return user when valid ID is provided', async () => {
      const response = await request(app)
        .get(`/api/users/${userId}`)
        .expect(200);

      expect(response.body.success).toBe(true);
      expect(response.body.data.email).toBe('getuser@example.com');
      expect(response.body.data.password).toBeUndefined();
    });

    it('should return 404 for non-existent user', async () => {
      const response = await request(app)
        .get('/api/users/non-existent-id')
        .expect(404);

      expect(response.body.error).toBe('User not found');
    });

    it('should return 400 for invalid ID format', async () => {
      const response = await request(app)
        .get('/api/users/invalid-id')
        .expect(400);

      expect(response.body.error).toBe('Validation failed');
    });
  });
});
```

## Security Best Practices

### Input Validation and Sanitization

```javascript
// src/middleware/validation.js
const { body, param, query, validationResult } = require('express-validator');
const createDOMPurify = require('dompurify');
const { JSDOM } = require('jsdom');

// Create DOMPurify instance
const window = new JSDOM('').window;
const dompurify = createDOMPurify(window);

class SecurityValidator {
  // XSS prevention
  static sanitizeHtml(input) {
    if (typeof input !== 'string') {
      return input;
    }
    
    return dompurify.sanitize(input, {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a'],
      ALLOWED_ATTR: ['href', 'title'],
      ALLOW_DATA_ATTR: false
    });
  }

  // SQL injection prevention
  static sanitizeSql(input) {
    if (typeof input !== 'string') {
      return input;
    }
    
    // Remove dangerous SQL characters and patterns
    return input
      .replace(/['"\;]/g, '')
      .replace(/--/g, '')
      .replace(/\/\*/g, '')
      .replace(/\*\//g, '')
      .replace(/drop\s+table/i, '')
      .replace(/delete\s+from/i, '')
      .replace(/insert\s+into/i, '')
      .replace(/update\s+\w+\s+set/i, '');
  }

  // Email validation with advanced checks
  static isValidEmail(email) {
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    
    if (!emailRegex.test(email)) {
      return false;
    }

    // Additional checks
    const [localPart, domain] = email.split('@');
    
    // Local part validation
    if (localPart.length > 64) return false;
    if (localPart.startsWith('.') || localPart.endsWith('.')) return false;
    if (localPart.includes('..')) return false;
    
    // Domain validation
    if (domain.length > 253) return false;
    if (!domain.includes('.')) return false;
    
    // Check for disposable email domains
    const disposableDomains = ['10minutemail.com', 'tempmail.org', 'guerrillamail.com'];
    if (disposableDomains.some(d => domain.toLowerCase().includes(d))) {
      return false;
    }

    return true;
  }

  // Password strength validation
  static validatePasswordStrength(password) {
    const issues = [];

    if (password.length < 8) {
      issues.push('Password must be at least 8 characters long');
    }

    if (password.length > 128) {
      issues.push('Password must be less than 128 characters');
    }

    if (!/[a-z]/.test(password)) {
      issues.push('Password must contain at least one lowercase letter');
    }

    if (!/[A-Z]/.test(password)) {
      issues.push('Password must contain at least one uppercase letter');
    }

    if (!/\d/.test(password)) {
      issues.push('Password must contain at least one number');
    }

    if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
      issues.push('Password must contain at least one special character');
    }

    // Check for common patterns
    const commonPatterns = [
      /123456/,
      /password/i,
      /qwerty/i,
      /admin/i,
      /letmein/i
    ];

    if (commonPatterns.some(pattern => pattern.test(password))) {
      issues.push('Password contains common patterns that make it weak');
    }

    // Check for sequential characters
    const isSequential = (str) => {
      for (let i = 0; i < str.length - 2; i++) {
        const char1 = str.charCodeAt(i);
        const char2 = str.charCodeAt(i + 1);
        const char3 = str.charCodeAt(i + 2);
        
        if (char2 === char1 + 1 && char3 === char2 + 1) {
          return true;
        }
      }
      return false;
    };

    if (isSequential(password.toLowerCase())) {
      issues.push('Password contains sequential characters');
    }

    return {
      isValid: issues.length === 0,
      issues,
      score: Math.max(0, 100 - (issues.length * 20))
    };
  }

  // Rate limiting by IP and user
  static createRateLimiter(options = {}) {
    const rateLimit = require('express-rate-limit');
    
    return rateLimit({
      windowMs: options.windowMs || 15 * 60 * 1000, // 15 minutes
      max: options.max || 100, // limit each IP
      message: {
        error: 'Too many requests, please try again later',
        retryAfter: options.windowMs / 1000
      },
      standardHeaders: true,
      legacyHeaders: false,
      keyGenerator: options.keyGenerator || ((req) => {
        return req.ip + (req.user ? `:${req.user.id}` : '');
      }),
      skip: options.skip || ((req) => {
        // Skip rate limiting for certain routes or trusted IPs
        return req.path.startsWith('/health') || req.path.startsWith('/metrics');
      })
    });
  }

  // Content Security Policy
  static getCSPHeaders() {
    return {
      'Content-Security-Policy': [
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline' https://trusted-cdn.com",
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' data: https:",
        "font-src 'self'",
        "connect-src 'self' https://api.example.com",
        "frame-ancestors 'none'",
        "base-uri 'self'",
        "form-action 'self'"
      ].join('; '),
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      'X-XSS-Protection': '1; mode=block',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
    };
  }
}

module.exports = SecurityValidator;
```

### Authentication and Authorization

```javascript
// src/middleware/auth.js
const jwt = require('jsonwebtoken');
const bcrypt = require('bcrypt');
const crypto = require('crypto');
const UserService = require('../services/userService');

class AuthMiddleware {
  constructor(jwtSecret = process.env.JWT_SECRET) {
    this.jwtSecret = jwtSecret;
    this.userService = new UserService(database); // Assuming database is available
  }

  // JWT token generation
  generateToken(payload, options = {}) {
    const defaultOptions = {
      expiresIn: '24h',
      issuer: 'your-app',
      audience: 'your-app-users'
    };

    return jwt.sign(payload, this.jwtSecret, { ...defaultOptions, ...options });
  }

  // JWT token verification
  verifyToken(token) {
    try {
      return jwt.verify(token, this.jwtSecret);
    } catch (error) {
      if (error.name === 'TokenExpiredError') {
        throw new Error('Token expired');
      } else if (error.name === 'JsonWebTokenError') {
        throw new Error('Invalid token');
      }
      throw new Error('Token verification failed');
    }
  }

  // Authentication middleware
  authenticate() {
    return async (req, res, next) => {
      try {
        const authHeader = req.headers.authorization;
        
        if (!authHeader || !authHeader.startsWith('Bearer ')) {
          return res.status(401).json({
            error: 'Authorization header required'
          });
        }

        const token = authHeader.substring(7);
        const decoded = this.verifyToken(token);

        // Get user from database to ensure they still exist
        const user = await this.userService.getUserById(decoded.sub);
        if (!user) {
          return res.status(401).json({
            error: 'User not found'
          });
        }

        // Add user to request object
        req.user = user;
        req.token = token;

        next();
      } catch (error) {
        return res.status(401).json({
          error: error.message
        });
      }
    };
  }

  // Role-based authorization
  requireRole(roles) {
    return (req, res, next) => {
      if (!req.user) {
        return res.status(401).json({
          error: 'Authentication required'
        });
      }

      const userRoles = Array.isArray(req.user.roles) ? req.user.roles : [req.user.role];
      const requiredRoles = Array.isArray(roles) ? roles : [roles];

      const hasRequiredRole = requiredRoles.some(role => userRoles.includes(role));

      if (!hasRequiredRole) {
        return res.status(403).json({
          error: 'Insufficient permissions',
          required: requiredRoles,
          current: userRoles
        });
      }

      next();
    };
  }

  // Permission-based authorization
  requirePermission(permission) {
    return (req, res, next) => {
      if (!req.user) {
        return res.status(401).json({
          error: 'Authentication required'
        });
      }

      const userPermissions = req.user.permissions || [];
      
      if (!userPermissions.includes(permission)) {
        return res.status(403).json({
          error: 'Insufficient permissions',
          required: permission,
          current: userPermissions
        });
      }

      next();
    };
  }

  // Resource ownership check
  requireOwnership(resourceIdParam = 'id') {
    return async (req, res, next) => {
      try {
        if (!req.user) {
          return res.status(401).json({
            error: 'Authentication required'
          });
        }

        const resourceId = req.params[resourceIdParam];
        const resource = await this.getResourceById(resourceId, req.path);

        if (!resource) {
          return res.status(404).json({
            error: 'Resource not found'
          });
        }

        // Check if user owns the resource or is admin
        const isOwner = resource.userId === req.user.id;
        const isAdmin = req.user.role === 'admin';

        if (!isOwner && !isAdmin) {
          return res.status(403).json({
            error: 'Access denied: You can only access your own resources'
          });
        }

        req.resource = resource;
        next();
      } catch (error) {
        return res.status(500).json({
          error: 'Authorization check failed'
        });
      }
    };
  }

  // Password hashing and verification
  async hashPassword(password) {
    const saltRounds = 12;
    return bcrypt.hash(password, saltRounds);
  }

  async verifyPassword(password, hashedPassword) {
    return bcrypt.compare(password, hashedPassword);
  }

  // Password reset token generation
  generateResetToken() {
    return crypto.randomBytes(32).toString('hex');
  }

  // Session management
  async createSession(user, req) {
    const sessionToken = crypto.randomBytes(32).toString('hex');
    
    // Store session in database or Redis
    await this.userService.createSession({
      token: sessionToken,
      userId: user.id,
      userAgent: req.get('User-Agent'),
      ipAddress: req.ip,
      createdAt: new Date(),
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000) // 24 hours
    });

    return sessionToken;
  }

  async invalidateSession(token) {
    await this.userService.deleteSession(token);
  }

  // Helper method to get resource by ID
  async getResourceById(id, path) {
    // This is a simplified example
    // In practice, you'd have different strategies for different resource types
    
    if (path.includes('/users')) {
      return await this.userService.getUserById(id);
    }
    
    if (path.includes('/posts')) {
      return await this.userService.getPostById(id);
    }
    
    return null;
  }
}

module.exports = AuthMiddleware;
```

## Integration Patterns

### REST API with GraphQL Integration

```javascript
// src/graphql/resolver.js
const { gql } = require('apollo-server-express');
const UserService = require('../services/userService');

const userService = new UserService(database);

// GraphQL type definitions
const typeDefs = gql`
  type User {
    id: ID!
    email: String!
    name: String!
    createdAt: String!
    updatedAt: String!
  }

  type Query {
    users(limit: Int, offset: Int, search: String): [User!]!
    user(id: ID!): User
  }

  type Mutation {
    createUser(input: CreateUserInput!): User!
    updateUser(id: ID!, input: UpdateUserInput!): User!
    deleteUser(id: ID!): Boolean!
  }

  input CreateUserInput {
    email: String!
    name: String!
    password: String!
  }

  input UpdateUserInput {
    name: String
    email: String
  }
`;

// GraphQL resolvers
const resolvers = {
  Query: {
    users: async (_, { limit = 10, offset = 0, search }) => {
      try {
        const filter = {};
        if (search) {
          filter.$or = [
            { name: { $regex: search, $options: 'i' } },
            { email: { $regex: search, $options: 'i' } }
          ];
        }

        const users = await userService.getUsers(filter, { limit, offset });
        return users;
      } catch (error) {
        throw new Error(`Failed to fetch users: ${error.message}`);
      }
    },

    user: async (_, { id }) => {
      try {
        const user = await userService.getUserById(id);
        if (!user) {
          throw new Error('User not found');
        }
        return user;
      } catch (error) {
        throw new Error(`Failed to fetch user: ${error.message}`);
      }
    }
  },

  Mutation: {
    createUser: async (_, { input }) => {
      try {
        const user = await userService.createUser(input);
        return user;
      } catch (error) {
        throw new Error(`Failed to create user: ${error.message}`);
      }
    },

    updateUser: async (_, { id, input }) => {
      try {
        const user = await userService.updateUser(id, input);
        if (!user) {
          throw new Error('User not found');
        }
        return user;
      } catch (error) {
        throw new Error(`Failed to update user: ${error.message}`);
      }
    },

    deleteUser: async (_, { id }) => {
      try {
        await userService.deleteUser(id);
        return true;
      } catch (error) {
        throw new Error(`Failed to delete user: ${error.message}`);
      }
    }
  }
};

module.exports = { typeDefs, resolvers };
```

### WebSocket Integration

```javascript
// src/websocket/socketHandler.js
const { Server } = require('ws');
const jwt = require('jsonwebtoken');
const UserService = require('../services/userService');

class SocketHandler {
  constructor(server, options = {}) {
    this.wss = new Server({ 
      server,
      path: options.path || '/ws',
      verifyClient: this.verifyClient.bind(this)
    });
    
    this.clients = new Map(); // userId -> WebSocket connection
    this.rooms = new Map(); // room -> Set of userIds
    
    this.setupEventHandlers();
  }

  // Client verification
  async verifyClient(info) {
    try {
      const token = this.extractToken(info.req);
      if (!token) {
        return false;
      }

      const decoded = jwt.verify(token, process.env.JWT_SECRET);
      const user = await new UserService(database).getUserById(decoded.sub);
      
      if (!user) {
        return false;
      }

      info.req.user = user;
      return true;
    } catch (error) {
      return false;
    }
  }

  // Extract JWT token from request
  extractToken(req) {
    const authHeader = req.headers.authorization;
    if (authHeader && authHeader.startsWith('Bearer ')) {
      return authHeader.substring(7);
    }
    
    // Also check query parameters for fallback
    const urlParams = new URL(req.url, `http://${req.headers.host}`).searchParams;
    return urlParams.get('token');
  }

  // Setup WebSocket event handlers
  setupEventHandlers() {
    this.wss.on('connection', (ws, req) => {
      this.handleConnection(ws, req);
    });

    this.wss.on('error', (error) => {
      console.error('WebSocket error:', error);
    });
  }

  // Handle new connection
  handleConnection(ws, req) {
    const user = req.user;
    
    // Store client connection
    this.clients.set(user.id, {
      ws,
      user,
      joinedAt: new Date(),
      lastActivity: new Date()
    });

    // Send welcome message
    this.sendToUser(user.id, {
      type: 'connection',
      message: 'Connected successfully',
      timestamp: new Date().toISOString()
    });

    // Setup message handler
    ws.on('message', (message) => {
      this.handleMessage(user.id, message);
    });

    // Setup close handler
    ws.on('close', () => {
      this.handleDisconnection(user.id);
    });

    // Setup error handler
    ws.on('error', (error) => {
      console.error(`WebSocket error for user ${user.id}:`, error);
    });

    // Update last activity
    ws.on('pong', () => {
      const client = this.clients.get(user.id);
      if (client) {
        client.lastActivity = new Date();
      }
    });

    console.log(`User ${user.id} connected via WebSocket`);
  }

  // Handle incoming messages
  async handleMessage(userId, message) {
    try {
      const client = this.clients.get(userId);
      if (!client) return;

      client.lastActivity = new Date();

      const data = JSON.parse(message);
      
      switch (data.type) {
        case 'join_room':
          await this.handleJoinRoom(userId, data.room);
          break;
          
        case 'leave_room':
          await this.handleLeaveRoom(userId, data.room);
          break;
          
        case 'send_message':
          await this.handleSendMessage(userId, data);
          break;
          
        case 'typing':
          await this.handleTyping(userId, data);
          break;
          
        default:
          this.sendToUser(userId, {
            type: 'error',
            message: 'Unknown message type',
            timestamp: new Date().toISOString()
          });
      }
    } catch (error) {
      console.error(`Error handling message from user ${userId}:`, error);
      this.sendToUser(userId, {
        type: 'error',
        message: 'Failed to process message',
        timestamp: new Date().toISOString()
      });
    }
  }

  // Handle room joining
  async handleJoinRoom(userId, roomName) {
    const room = this.rooms.get(roomName) || new Set();
    room.add(userId);
    this.rooms.set(roomName, room);

    // Notify room members
    this.broadcastToRoom(roomName, {
      type: 'user_joined',
      userId,
      room: roomName,
      timestamp: new Date().toISOString()
    }, userId);

    // Send confirmation to user
    this.sendToUser(userId, {
      type: 'joined_room',
      room: roomName,
      timestamp: new Date().toISOString()
    });
  }

  // Handle room leaving
  async handleLeaveRoom(userId, roomName) {
    const room = this.rooms.get(roomName);
    if (room) {
      room.delete(userId);
      
      if (room.size === 0) {
        this.rooms.delete(roomName);
      }

      // Notify room members
      this.broadcastToRoom(roomName, {
        type: 'user_left',
        userId,
        room: roomName,
        timestamp: new Date().toISOString()
      }, userId);
    }

    // Send confirmation to user
    this.sendToUser(userId, {
      type: 'left_room',
      room: roomName,
      timestamp: new Date().toISOString()
    });
  }

  // Handle message sending
  async handleSendMessage(userId, data) {
    const { room, message, type = 'text' } = data;

    // Save message to database (implement this)
    const savedMessage = await this.saveMessage({
      userId,
      room,
      message,
      type,
      timestamp: new Date()
    });

    // Broadcast to room members
    this.broadcastToRoom(room, {
      type: 'new_message',
      message: savedMessage,
      timestamp: new Date().toISOString()
    });
  }

  // Handle typing indicators
  async handleTyping(userId, data) {
    const { room, isTyping } = data;

    this.broadcastToRoom(room, {
      type: 'typing_indicator',
      userId,
      isTyping,
      timestamp: new Date().toISOString()
    }, userId);
  }

  // Send message to specific user
  sendToUser(userId, message) {
    const client = this.clients.get(userId);
    if (client && client.ws.readyState === client.ws.OPEN) {
      client.ws.send(JSON.stringify(message));
    }
  }

  // Broadcast message to room
  broadcastToRoom(roomName, message, excludeUserId = null) {
    const room = this.rooms.get(roomName);
    if (!room) return;

    room.forEach(userId => {
      if (userId !== excludeUserId) {
        this.sendToUser(userId, message);
      }
    });
  }

  // Handle client disconnection
  handleDisconnection(userId) {
    // Remove from all rooms
    for (const [roomName, room] of this.rooms.entries()) {
      room.delete(userId);
      
      if (room.size === 0) {
        this.rooms.delete(roomName);
      } else {
        // Notify remaining room members
        this.broadcastToRoom(roomName, {
          type: 'user_disconnected',
          userId,
          timestamp: new Date().toISOString()
        });
      }
    }

    // Remove client
    this.clients.delete(userId);
    console.log(`User ${userId} disconnected`);
  }

  // Get connection statistics
  getStats() {
    return {
      connectedClients: this.clients.size,
      activeRooms: this.rooms.size,
      rooms: Array.from(this.rooms.entries()).map(([name, users]) => ({
        name,
        userCount: users.size
      }))
    };
  }

  // Cleanup inactive connections
  cleanup() {
    const now = new Date();
    const timeoutMs = 5 * 60 * 1000; // 5 minutes

    for (const [userId, client] of this.clients.entries()) {
      if (now - client.lastActivity > timeoutMs) {
        client.ws.terminate();
        this.handleDisconnection(userId);
      }
    }
  }

  // Save message to database (implement this)
  async saveMessage(messageData) {
    // This would save to your database
    return {
      id: crypto.randomUUID(),
      ...messageData
    };
  }
}

module.exports = SocketHandler;
```

## Modern Development Workflow

### Configuration Management

```javascript
// src/config/index.js
const path = require('path');
require('dotenv').config();

class Config {
  constructor() {
    this.env = process.env.NODE_ENV || 'development';
    this.isDevelopment = this.env === 'development';
    this.isProduction = this.env === 'production';
    this.isTest = this.env === 'test';
    
    this.loadConfiguration();
  }

  loadConfiguration() {
    this.server = {
      host: process.env.HOST || '0.0.0.0',
      port: parseInt(process.env.PORT) || 3000,
      cors: {
        origins: this.parseArray(process.env.CORS_ORIGINS) || ['http://localhost:3000']
      }
    };

    this.database = {
      url: process.env.DATABASE_URL || 'mongodb://localhost:27017/myapp',
      options: {
        useNewUrlParser: true,
        useUnifiedTopology: true,
        maxPoolSize: parseInt(process.env.DB_MAX_POOL_SIZE) || 10,
        serverSelectionTimeoutMS: parseInt(process.env.DB_TIMEOUT) || 5000,
      }
    };

    this.redis = {
      url: process.env.REDIS_URL || 'redis://localhost:6379',
      options: {
        retryDelayOnFailover: 100,
        enableReadyCheck: false,
        maxRetriesPerRequest: null
      }
    };

    this.auth = {
      jwtSecret: process.env.JWT_SECRET || this.generateSecret(),
      jwtExpiration: process.env.JWT_EXPIRATION || '24h',
      bcryptRounds: parseInt(process.env.BCRYPT_ROUNDS) || 12
    };

    this.rateLimit = {
      max: parseInt(process.env.RATE_LIMIT_MAX) || 100,
      windowMs: parseInt(process.env.RATE_LIMIT_WINDOW) || 15 * 60 * 1000
    };

    this.logging = {
      level: process.env.LOG_LEVEL || 'info',
      format: process.env.LOG_FORMAT || 'combined',
      file: process.env.LOG_FILE
    };

    this.upload = {
      maxFileSize: parseInt(process.env.MAX_FILE_SIZE) || 5 * 1024 * 1024, // 5MB
      allowedTypes: this.parseArray(process.env.ALLOWED_FILE_TYPES) || ['image/jpeg', 'image/png'],
      destination: process.env.UPLOAD_DESTINATION || './uploads'
    };

    this.email = {
      provider: process.env.EMAIL_PROVIDER || 'sendgrid',
      from: process.env.EMAIL_FROM,
      apiKey: process.env.EMAIL_API_KEY
    };

    this.monitoring = {
      enableMetrics: process.env.ENABLE_METRICS === 'true',
      metricsPort: parseInt(process.env.METRICS_PORT) || 9090,
      healthCheckInterval: parseInt(process.env.HEALTH_CHECK_INTERVAL) || 30000
    };
  }

  parseArray(value) {
    if (!value) return null;
    if (Array.isArray(value)) return value;
    return value.split(',').map(item => item.trim());
  }

  generateSecret() {
    if (this.isProduction) {
      throw new Error('JWT_SECRET must be set in production');
    }
    return require('crypto').randomBytes(64).toString('hex');
  }

  validate() {
    const errors = [];

    if (this.isProduction && !process.env.JWT_SECRET) {
      errors.push('JWT_SECRET is required in production');
    }

    if (!this.database.url) {
      errors.push('DATABASE_URL is required');
    }

    if (errors.length > 0) {
      throw new Error(`Configuration validation failed: ${errors.join(', ')}`);
    }
  }

  getDatabaseConfig() {
    return {
      url: this.database.url,
      options: this.database.options
    };
  }

  getRedisConfig() {
    return {
      url: this.redis.url,
      options: this.redis.options
    };
  }
}

// Create singleton instance
const config = new Config();

// Validate configuration on startup
config.validate();

module.exports = config;
```

### Package.json with Modern Scripts

```json
{
  "name": "my-javascript-project",
  "version": "1.0.0",
  "description": "Modern Node.js application",
  "main": "src/app.js",
  "type": "commonjs",
  "scripts": {
    "start": "node src/app.js",
    "dev": "nodemon src/app.js",
    "dev:debug": "nodemon --inspect src/app.js",
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:integration": "jest --testPathPattern=test/integration",
    "lint": "eslint src/ test/",
    "lint:fix": "eslint src/ test/ --fix",
    "format": "prettier --write src/ test/",
    "validate": "npm run lint && npm run test",
    "build": "echo 'No build step required for JavaScript'",
    "clean": "rm -rf node_modules package-lock.json",
    "fresh": "npm run clean && npm install",
    "security:audit": "npm audit",
    "security:fix": "npm audit fix",
    "db:migrate": "node scripts/migrate.js",
    "db:seed": "node scripts/seed.js",
    "logs": "tail -f logs/app.log",
    "docker:build": "docker build -t my-app .",
    "docker:run": "docker run -p 3000:3000 my-app"
  },
  "dependencies": {
    "express": "^4.21.0",
    "helmet": "^7.1.0",
    "cors": "^2.8.5",
    "compression": "^1.7.5",
    "morgan": "^1.10.0",
    "express-rate-limit": "^7.4.0",
    "express-validator": "^7.1.0",
    "bcrypt": "^5.1.1",
    "jsonwebtoken": "^9.0.2",
    "mongoose": "^8.8.1",
    "redis": "^4.7.0",
    "dotenv": "^16.4.5",
    "winston": "^3.15.0",
    "joi": "^17.13.3",
    "multer": "^1.4.5-lts.1",
    "nodemailer": "^6.9.8",
    "apollo-server-express": "^3.12.1",
    "graphql": "^16.9.0",
    "ws": "^8.18.0",
    "uuid": "^10.0.0"
  },
  "devDependencies": {
    "nodemon": "^3.1.4",
    "jest": "^30.0.4",
    "supertest": "^7.0.0",
    "eslint": "^9.0.0",
    "eslint-config-standard": "^17.1.0",
    "eslint-plugin-import": "^2.31.0",
    "eslint-plugin-node": "^11.1.0",
    "eslint-plugin-promise": "^7.1.0",
    "prettier": "^3.3.3",
    "@types/jest": "^30.0.4",
    "mongodb-memory-server": "^9.16.0"
  },
  "engines": {
    "node": ">=18.0.0",
    "npm": ">=8.0.0"
  },
  "keywords": [
    "nodejs",
    "express",
    "javascript",
    "api",
    "backend"
  ],
  "author": "Your Name",
  "license": "MIT"
}
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**JavaScript Target**: Node.js 22.x + Modern ES2025 Features  

This skill provides comprehensive JavaScript development guidance with 2025 best practices, covering everything from legacy code maintenance to modern Node.js backend development and browser compatibility strategies.
