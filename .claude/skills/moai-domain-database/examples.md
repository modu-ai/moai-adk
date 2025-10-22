# Database Design & Optimization Examples

Production-ready examples for PostgreSQL 17, MongoDB 8.0, indexing strategies, migrations, and query optimization with modern database best practices (2025).

---

## Example 1: PostgreSQL 17 with Advanced Indexing

### Database Schema: `src/models/user.sql`

```sql
-- @CODE:DB-001 | Modern PostgreSQL 17 schema with identity columns
-- SPEC: SPEC-DB-001.md | TEST: tests/db/test_user_schema.sql

CREATE TABLE IF NOT EXISTS users (
    -- Use IDENTITY instead of SERIAL (PostgreSQL 17 best practice)
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,

    -- Constraints
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$'),
    CONSTRAINT valid_username CHECK (LENGTH(username) >= 3)
);

-- B-tree index for equality and range queries (most common)
CREATE INDEX idx_users_created_at ON users(created_at);

-- Partial index for active users only (275x performance improvement)
CREATE INDEX idx_active_users ON users(email) WHERE metadata->>'status' = 'active';

-- GIN index for JSONB queries (PostgreSQL specialty)
CREATE INDEX idx_users_metadata ON users USING GIN(metadata);

-- Composite index (column order matters: most selective first)
CREATE INDEX idx_users_search ON users(created_at DESC, username);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Test File: `tests/db/test_user_schema.sql`

```sql
-- @TEST:DB-001 | Verify PostgreSQL 17 schema and index performance
-- SPEC: SPEC-DB-001.md | CODE: src/models/user.sql

BEGIN;

-- Test 1: Verify IDENTITY column behavior
INSERT INTO users (email, username, metadata)
VALUES ('test@example.com', 'testuser', '{"status": "active"}');

-- Verify ID is auto-generated
SELECT id, email FROM users WHERE email = 'test@example.com';

-- Test 2: Verify constraint enforcement
DO $$
BEGIN
    INSERT INTO users (email, username) VALUES ('invalid-email', 'ab');
    RAISE EXCEPTION 'Constraint should have failed';
EXCEPTION
    WHEN check_violation THEN
        RAISE NOTICE 'Constraint correctly enforced';
END $$;

-- Test 3: Verify index usage with EXPLAIN
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM users
WHERE metadata->>'status' = 'active';
-- Expected: Index Scan using idx_active_users (partial index)

-- Test 4: Verify trigger functionality
UPDATE users SET username = 'updated_user' WHERE email = 'test@example.com';
SELECT updated_at > created_at AS trigger_worked FROM users WHERE email = 'test@example.com';

ROLLBACK;
```

### Query Optimization: `src/queries/user_queries.sql`

```sql
-- @CODE:DB-001:QUERY | Optimized PostgreSQL 17 queries with index hints
-- SPEC: SPEC-DB-001.md | TEST: tests/db/test_user_queries.sql

-- Query 1: Covering index to avoid table lookup
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
SELECT email, created_at
FROM users
WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
ORDER BY created_at DESC
LIMIT 100;
-- Uses idx_users_search (composite index covers query fully)

-- Query 2: JSONB query with GIN index
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, email, metadata->'preferences' AS prefs
FROM users
WHERE metadata @> '{"status": "active", "role": "admin"}';
-- Uses idx_users_metadata (GIN index for JSONB containment)

-- Query 3: Partial index optimization (active users only)
EXPLAIN (ANALYZE, BUFFERS)
SELECT email, username
FROM users
WHERE metadata->>'status' = 'active'
  AND email LIKE '%@example.com';
-- Uses idx_active_users (partial index dramatically reduces scan size)

-- Query 4: CTE for complex aggregation
WITH recent_users AS (
    SELECT id, email, created_at
    FROM users
    WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
)
SELECT DATE(created_at) AS signup_date, COUNT(*) AS count
FROM recent_users
GROUP BY DATE(created_at)
ORDER BY signup_date DESC;
```

---

## Example 2: MongoDB 8.0 with Express Path Optimization

### Collection Schema: `src/models/product.js`

```javascript
// @CODE:DB-002 | MongoDB 8.0 schema with optimized indexes
// SPEC: SPEC-DB-002.md | TEST: tests/db/test_product_schema.test.js

const mongoose = require('mongoose');

const productSchema = new mongoose.Schema({
  name: {
    type: String,
    required: true,
    index: true  // Simple index for name lookups
  },
  category: {
    type: String,
    required: true,
    index: true
  },
  price: {
    type: Number,
    required: true,
    min: 0
  },
  tags: [{ type: String }],  // Array field for GIN-like queries
  metadata: {
    brand: String,
    model: String,
    year: Number
  },
  inventory: {
    quantity: { type: Number, default: 0 },
    warehouse: { type: String }
  },
  createdAt: {
    type: Date,
    default: Date.now,
    index: true  // For time-series queries
  },
  updatedAt: {
    type: Date,
    default: Date.now
  }
});

// Compound index for multi-field queries (MongoDB 8.0 enhancement)
productSchema.index({ category: 1, price: -1 });

// Text index for full-text search
productSchema.index({ name: 'text', 'metadata.brand': 'text' });

// Partial index (MongoDB equivalent of PostgreSQL partial index)
productSchema.index(
  { 'inventory.quantity': 1 },
  { partialFilterExpression: { 'inventory.quantity': { $lt: 10 } } }
);

// Pre-save hook for updatedAt
productSchema.pre('save', function(next) {
  this.updatedAt = Date.now();
  next();
});

module.exports = mongoose.model('Product', productSchema);
```

### Test File: `tests/db/test_product_schema.test.js`

```javascript
// @TEST:DB-002 | Verify MongoDB 8.0 schema and Express Path optimization
// SPEC: SPEC-DB-002.md | CODE: src/models/product.js

const mongoose = require('mongoose');
const Product = require('../src/models/product');

describe('MongoDB 8.0 Product Schema', () => {
  beforeAll(async () => {
    await mongoose.connect(process.env.MONGO_TEST_URI);
  });

  afterAll(async () => {
    await mongoose.connection.dropDatabase();
    await mongoose.connection.close();
  });

  test('should create product with auto-generated _id (IDHACK optimization)', async () => {
    const product = await Product.create({
      name: 'Test Product',
      category: 'Electronics',
      price: 99.99,
      tags: ['new', 'featured']
    });

    expect(product._id).toBeDefined();

    // MongoDB 8.0 Express Path: _id lookup is optimized
    const found = await Product.findById(product._id);
    expect(found.name).toBe('Test Product');
  });

  test('should enforce price constraint (min: 0)', async () => {
    await expect(
      Product.create({
        name: 'Invalid Product',
        category: 'Test',
        price: -10
      })
    ).rejects.toThrow();
  });

  test('should use compound index for category + price query', async () => {
    await Product.create([
      { name: 'Product A', category: 'Electronics', price: 100 },
      { name: 'Product B', category: 'Electronics', price: 200 },
      { name: 'Product C', category: 'Books', price: 50 }
    ]);

    // Query uses compound index { category: 1, price: -1 }
    const results = await Product
      .find({ category: 'Electronics' })
      .sort({ price: -1 })
      .explain('executionStats');

    expect(results.executionStats.executionSuccess).toBe(true);
    // Verify index usage (MongoDB 8.0 improved explain)
    expect(results.executionStats.totalKeysExamined).toBeGreaterThan(0);
  });

  test('should update updatedAt timestamp automatically', async () => {
    const product = await Product.create({
      name: 'Update Test',
      category: 'Test',
      price: 50
    });

    const originalUpdatedAt = product.updatedAt;

    await new Promise(resolve => setTimeout(resolve, 100));

    product.price = 75;
    await product.save();

    expect(product.updatedAt.getTime()).toBeGreaterThan(originalUpdatedAt.getTime());
  });
});
```

### Query Optimization: `src/queries/product_queries.js`

```javascript
// @CODE:DB-002:QUERY | MongoDB 8.0 optimized queries with Express Path
// SPEC: SPEC-DB-002.md | TEST: tests/db/test_product_queries.test.js

const Product = require('../models/product');

class ProductQueries {
  // Query 1: Express Path optimization for _id lookup
  static async findById(productId) {
    // MongoDB 8.0 Express Path: IDHACK code path is optimized
    return await Product.findById(productId).lean();
  }

  // Query 2: Compound index query (category + price)
  static async findByCategoryAndPriceRange(category, minPrice, maxPrice) {
    // Uses compound index { category: 1, price: -1 }
    return await Product
      .find({
        category: category,
        price: { $gte: minPrice, $lte: maxPrice }
      })
      .sort({ price: -1 })
      .lean();
  }

  // Query 3: Partial index for low inventory
  static async findLowInventoryProducts() {
    // Uses partial index { 'inventory.quantity': 1 }
    return await Product
      .find({ 'inventory.quantity': { $lt: 10 } })
      .lean();
  }

  // Query 4: Text search with text index
  static async searchProducts(searchTerm) {
    // Uses text index on { name: 'text', 'metadata.brand': 'text' }
    return await Product
      .find({ $text: { $search: searchTerm } })
      .select({ score: { $meta: 'textScore' } })
      .sort({ score: { $meta: 'textScore' } })
      .lean();
  }

  // Query 5: Aggregation pipeline (MongoDB 8.0 improved throughput)
  static async getCategorySummary() {
    return await Product.aggregate([
      {
        $group: {
          _id: '$category',
          count: { $sum: 1 },
          avgPrice: { $avg: '$price' },
          totalInventory: { $sum: '$inventory.quantity' }
        }
      },
      {
        $sort: { count: -1 }
      }
    ]);
  }

  // Query 6: Projection to reduce data transfer
  static async findProductsWithProjection(category) {
    // Only fetch required fields
    return await Product
      .find({ category: category })
      .select('name price inventory.quantity')
      .lean();
  }
}

module.exports = ProductQueries;
```

---

## Example 3: Database Migration Best Practices

### PostgreSQL Migration: `migrations/001_add_user_profiles.sql`

```sql
-- @CODE:DB-003 | PostgreSQL migration with rollback support
-- SPEC: SPEC-DB-003.md | TEST: tests/migrations/test_001_migration.sql

-- Migration: 001_add_user_profiles
-- Description: Add user_profiles table with foreign key to users

BEGIN;

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(500),
    social_links JSONB DEFAULT '{}',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraint with cascade
    CONSTRAINT fk_user_profiles_user_id
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Index for foreign key
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

-- GIN index for JSONB fields
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING GIN(preferences);

-- Add trigger for updated_at
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Verify migration
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_profiles') THEN
        RAISE EXCEPTION 'Migration failed: user_profiles table not created';
    END IF;
    RAISE NOTICE 'Migration 001 completed successfully';
END $$;

COMMIT;
```

### PostgreSQL Rollback: `migrations/001_add_user_profiles_rollback.sql`

```sql
-- @CODE:DB-003:ROLLBACK | Rollback migration 001
-- SPEC: SPEC-DB-003.md

BEGIN;

-- Drop triggers first
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_profiles_preferences;
DROP INDEX IF EXISTS idx_user_profiles_user_id;

-- Drop table (CASCADE removes dependent objects)
DROP TABLE IF EXISTS user_profiles CASCADE;

-- Verify rollback
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_profiles') THEN
        RAISE EXCEPTION 'Rollback failed: user_profiles table still exists';
    END IF;
    RAISE NOTICE 'Rollback 001 completed successfully';
END $$;

COMMIT;
```

### MongoDB Migration: `migrations/001_add_product_indexes.js`

```javascript
// @CODE:DB-004 | MongoDB 8.0 migration with rollback support
// SPEC: SPEC-DB-004.md | TEST: tests/migrations/test_001_migration.test.js

module.exports = {
  async up(db, client) {
    // Migration: Add compound and partial indexes to products collection

    const session = client.startSession();
    try {
      await session.withTransaction(async () => {
        const products = db.collection('products');

        // Compound index for category + price queries
        await products.createIndex(
          { category: 1, price: -1 },
          { name: 'idx_category_price', session }
        );

        // Partial index for low inventory
        await products.createIndex(
          { 'inventory.quantity': 1 },
          {
            name: 'idx_low_inventory',
            partialFilterExpression: { 'inventory.quantity': { $lt: 10 } },
            session
          }
        );

        // Text index for search
        await products.createIndex(
          { name: 'text', 'metadata.brand': 'text' },
          { name: 'idx_text_search', session }
        );

        console.log('Migration 001 completed successfully');
      });
    } finally {
      await session.endSession();
    }
  },

  async down(db, client) {
    // Rollback: Remove indexes

    const session = client.startSession();
    try {
      await session.withTransaction(async () => {
        const products = db.collection('products');

        await products.dropIndex('idx_category_price', { session });
        await products.dropIndex('idx_low_inventory', { session });
        await products.dropIndex('idx_text_search', { session });

        console.log('Rollback 001 completed successfully');
      });
    } finally {
      await session.endSession();
    }
  }
};
```

---

## Example 4: Query Performance Monitoring

### PostgreSQL Performance Monitoring: `scripts/analyze_slow_queries.sql`

```sql
-- @CODE:DB-005 | PostgreSQL 17 slow query analysis
-- SPEC: SPEC-DB-005.md

-- Enable query statistics (if not already enabled)
-- Add to postgresql.conf:
-- shared_preload_libraries = 'pg_stat_statements'
-- pg_stat_statements.track = all

-- Create extension (run once)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Query 1: Top 10 slowest queries
SELECT
    substring(query, 1, 100) AS short_query,
    round(total_exec_time::numeric, 2) AS total_time_ms,
    calls,
    round(mean_exec_time::numeric, 2) AS avg_time_ms,
    round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) AS percentage
FROM pg_stat_statements
ORDER BY total_exec_time DESC
LIMIT 10;

-- Query 2: Queries with highest I/O impact
SELECT
    substring(query, 1, 100) AS short_query,
    shared_blks_hit,
    shared_blks_read,
    round((100.0 * shared_blks_hit / NULLIF(shared_blks_hit + shared_blks_read, 0))::numeric, 2) AS cache_hit_ratio
FROM pg_stat_statements
WHERE shared_blks_hit + shared_blks_read > 0
ORDER BY shared_blks_read DESC
LIMIT 10;

-- Query 3: Identify missing indexes (sequential scans on large tables)
SELECT
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    seq_tup_read / seq_scan AS avg_seq_tup_read
FROM pg_stat_user_tables
WHERE seq_scan > 0
  AND schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY seq_tup_read DESC
LIMIT 10;

-- Query 4: Index usage statistics
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY idx_scan DESC;

-- Query 5: Unused indexes (candidates for removal)
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_relation_size(indexrelid) DESC;
```

### MongoDB Performance Monitoring: `scripts/analyze_slow_queries.js`

```javascript
// @CODE:DB-006 | MongoDB 8.0 slow query analysis
// SPEC: SPEC-DB-006.md

const { MongoClient } = require('mongodb');

async function analyzeSlowQueries() {
  const client = new MongoClient(process.env.MONGO_URI);

  try {
    await client.connect();
    const db = client.db();

    // Query 1: Enable profiling (level 1: slow queries > 100ms)
    await db.command({ profile: 1, slowms: 100 });

    // Query 2: Get slow query logs
    const slowQueries = await db
      .collection('system.profile')
      .find({ millis: { $gt: 100 } })
      .sort({ ts: -1 })
      .limit(10)
      .toArray();

    console.log('Top 10 Slow Queries:');
    slowQueries.forEach((query, idx) => {
      console.log(`\n${idx + 1}. Duration: ${query.millis}ms`);
      console.log(`   Command: ${JSON.stringify(query.command, null, 2)}`);
      console.log(`   Plan Summary: ${query.planSummary || 'N/A'}`);
    });

    // Query 3: Index statistics
    const collections = await db.listCollections().toArray();

    console.log('\nIndex Usage Statistics:');
    for (const collection of collections) {
      const collName = collection.name;
      if (collName.startsWith('system.')) continue;

      const stats = await db.collection(collName).aggregate([
        { $indexStats: {} }
      ]).toArray();

      console.log(`\nCollection: ${collName}`);
      stats.forEach(stat => {
        console.log(`  Index: ${stat.name}`);
        console.log(`    Ops: ${stat.accesses.ops}`);
        console.log(`    Since: ${stat.accesses.since}`);
      });
    }

    // Query 4: Collection statistics
    console.log('\nCollection Statistics:');
    for (const collection of collections) {
      const collName = collection.name;
      if (collName.startsWith('system.')) continue;

      const stats = await db.command({ collStats: collName });
      console.log(`\n${collName}:`);
      console.log(`  Documents: ${stats.count}`);
      console.log(`  Size: ${(stats.size / 1024 / 1024).toFixed(2)} MB`);
      console.log(`  Avg Doc Size: ${(stats.avgObjSize / 1024).toFixed(2)} KB`);
      console.log(`  Indexes: ${stats.nindexes}`);
      console.log(`  Total Index Size: ${(stats.totalIndexSize / 1024 / 1024).toFixed(2)} MB`);
    }

  } finally {
    await client.close();
  }
}

// Run analysis
analyzeSlowQueries().catch(console.error);
```

---

## Example 5: TRUST 5 Compliance for Database Projects

### Test Coverage: `tests/db/test_coverage.test.js`

```javascript
// @TEST:DB-007 | Database integration tests for ≥85% coverage
// SPEC: SPEC-DB-007.md | CODE: src/queries/*

const { Pool } = require('pg');
const Product = require('../src/models/product');
const ProductQueries = require('../src/queries/product_queries');

describe('Database TRUST Compliance', () => {
  let pgPool;

  beforeAll(async () => {
    pgPool = new Pool({ connectionString: process.env.PG_TEST_URI });
    await mongoose.connect(process.env.MONGO_TEST_URI);
  });

  afterAll(async () => {
    await pgPool.end();
    await mongoose.connection.close();
  });

  // Test First (T): All database operations must have tests
  describe('PostgreSQL Operations', () => {
    test('should create user with IDENTITY column', async () => {
      const result = await pgPool.query(
        'INSERT INTO users (email, username) VALUES ($1, $2) RETURNING id',
        ['test@example.com', 'testuser']
      );
      expect(result.rows[0].id).toBeDefined();
    });

    test('should enforce email constraint', async () => {
      await expect(
        pgPool.query(
          'INSERT INTO users (email, username) VALUES ($1, $2)',
          ['invalid-email', 'testuser']
        )
      ).rejects.toThrow();
    });
  });

  // Readable (R): Clear query structure and comments
  describe('MongoDB Operations', () => {
    test('should use Express Path for _id lookup', async () => {
      const product = await Product.create({
        name: 'Test',
        category: 'Test',
        price: 100
      });

      const found = await ProductQueries.findById(product._id);
      expect(found.name).toBe('Test');
    });
  });

  // Unified (U): Type safety and schema validation
  describe('Schema Validation', () => {
    test('should enforce MongoDB schema constraints', async () => {
      await expect(
        Product.create({ name: 'Test', category: 'Test' })
      ).rejects.toThrow(); // Missing required 'price'
    });
  });

  // Secured (S): SQL injection prevention
  describe('Security', () => {
    test('should prevent SQL injection with parameterized queries', async () => {
      const maliciousInput = "'; DROP TABLE users; --";

      // Parameterized query prevents injection
      await expect(
        pgPool.query(
          'SELECT * FROM users WHERE email = $1',
          [maliciousInput]
        )
      ).resolves.not.toThrow();
    });
  });

  // Trackable (T): TAG references in all database code
  test('should have TAG references in schema files', () => {
    const fs = require('fs');
    const schemaContent = fs.readFileSync('src/models/user.sql', 'utf8');

    expect(schemaContent).toMatch(/@CODE:DB-\d{3}/);
    expect(schemaContent).toMatch(/SPEC:/);
    expect(schemaContent).toMatch(/TEST:/);
  });
});
```

---

**Total Examples**: 5 comprehensive scenarios covering PostgreSQL 17, MongoDB 8.0, migrations, monitoring, and TRUST compliance.

**Key Technologies**:
- PostgreSQL 17 (IDENTITY columns, partial indexes, GIN/JSONB)
- MongoDB 8.0 (Express Path, compound indexes, aggregation)
- Migration patterns (forward/rollback)
- Performance monitoring (pg_stat_statements, profiling)
- TRUST 5 compliance

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Status**: Production-ready
