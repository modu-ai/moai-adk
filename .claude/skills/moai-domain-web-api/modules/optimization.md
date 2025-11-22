# API Performance Optimization

## Response Compression & Streaming

### GZIP Compression

```typescript
import compression from 'compression';
import express from 'express';

const app = express();

// Enable GZIP compression for responses > 1KB
app.use(compression({
    threshold: 1024,      // Only compress responses > 1KB
    level: 6,             // Compression level (0-9)
    filter: (req, res) => {
        // Don't compress if client doesn't support
        if (req.headers['x-no-compression']) {
            return false;
        }
        return compression.filter(req, res);
    }
}));

// Example response
app.get('/large-dataset', (req, res) => {
    const data = Array(10000).fill({
        id: Math.random(),
        name: 'Item',
        value: Math.random() * 100
    });
    // Response automatically compressed due to middleware
    res.json(data);
});
```

**Performance Impact**:
- JSON response: 500KB → 50KB (90% reduction)
- Transfer time: Reduced by same ratio
- CPU: Minimal overhead (6-7% overhead for compression)

### Server-Sent Events (Streaming Responses)

```javascript
// Avoid buffering entire response
app.get('/stream-events', (req, res) => {
    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    // Stream data progressively
    let eventId = 0;
    const interval = setInterval(() => {
        res.write(`id: ${eventId}\n`);
        res.write(`data: ${JSON.stringify({ value: Math.random() })}\n\n`);
        eventId++;

        if (eventId > 100) {
            clearInterval(interval);
            res.end();
        }
    }, 100);

    // Cleanup on disconnect
    req.on('close', () => clearInterval(interval));
});
```

---

## Pagination Optimization

### Cursor-Based Pagination (vs Offset)

```typescript
// Bad: Offset pagination (O(n) database operation)
// SELECT * FROM users ORDER BY id LIMIT 10 OFFSET 1000000
// Database must scan 1,000,000+ rows!

// Good: Cursor-based pagination (O(1) database operation with index)
app.get('/users', (req, res) => {
    const cursor = req.query.cursor || null;
    const limit = Math.min(parseInt(req.query.limit) || 10, 100);

    let query = db.users.orderBy('id', 'asc');

    if (cursor) {
        // Decode cursor and fetch from that point
        const cursorId = Buffer.from(cursor, 'base64').toString();
        query = query.where('id', '>', cursorId);
    }

    const users = query.limit(limit + 1).get();

    // Check if more results available
    const hasMore = users.length > limit;
    const result = users.slice(0, limit);

    // Encode next cursor
    const nextCursor = hasMore
        ? Buffer.from(result[result.length - 1].id.toString()).toString('base64')
        : null;

    res.json({
        data: result,
        pagination: {
            cursor: nextCursor,
            hasMore: hasMore
        }
    });
});
```

**Performance Comparison**:
- Offset pagination on page 10,000: ~500ms (scans 10M rows)
- Cursor pagination on page 10,000: ~5ms (direct lookup via index)

---

## Caching Strategy

### HTTP Cache Headers

```typescript
app.get('/api/public-data', (req, res) => {
    res.setHeader('Cache-Control', 'public, max-age=3600');
    res.setHeader('ETag', '"abc123"');
    res.json({ data: 'public-immutable-data' });
});

app.get('/api/user-data', (req, res) => {
    // Private data (not cacheable by proxies)
    res.setHeader('Cache-Control', 'private, max-age=60');
    res.json({ userId: 123, email: 'user@example.com' });
});

app.get('/api/frequently-changing', (req, res) => {
    // Must revalidate with server
    res.setHeader('Cache-Control', 'max-age=0, must-revalidate');
    res.setHeader('Last-Modified', new Date().toUTCString());
    res.json({ data: 'fresh' });
});

// Handle conditional requests (304 Not Modified)
app.get('/api/data', (req, res) => {
    const data = { content: 'important' };
    const etag = generateETag(data);

    if (req.get('If-None-Match') === etag) {
        // Client has cached version
        return res.status(304).send();
    }

    res.setHeader('ETag', etag);
    res.json(data);
});
```

### Redis Caching

```typescript
import Redis from 'ioredis';

const redis = new Redis({
    host: 'localhost',
    port: 6379,
    retryStrategy: (times) => Math.min(times * 50, 2000)
});

async function getCachedUser(userId) {
    const cacheKey = `user:${userId}`;

    // Try cache first
    const cached = await redis.get(cacheKey);
    if (cached) {
        return JSON.parse(cached);
    }

    // Cache miss: fetch from database
    const user = await db.users.findById(userId);

    // Cache for 1 hour
    await redis.setex(cacheKey, 3600, JSON.stringify(user));

    return user;
}

app.get('/users/:id', async (req, res) => {
    const user = await getCachedUser(req.params.id);
    res.json(user);
});

// Cache invalidation on update
app.put('/users/:id', async (req, res) => {
    const user = await db.users.updateById(req.params.id, req.body);

    // Invalidate cache
    await redis.del(`user:${req.params.id}`);

    res.json(user);
});
```

---

## Request/Response Optimization

### Request Timeout Configuration

```typescript
import http from 'http';

const server = http.createServer((req, res) => {
    // Set socket timeout (detects hung connections)
    req.socket.setTimeout(30000); // 30 seconds

    // Set header timeout
    server.headersTimeout = 60000; // 60 seconds

    // Set keep-alive timeout
    server.keepAliveTimeout = 65000; // 65 seconds
});

// Handle timeout
server.on('clientError', (err, socket) => {
    if (err.code === 'ECONNRESET' || !socket.writable) {
        return;
    }
    socket.end('HTTP/1.1 408 Request Timeout\r\n\r\n');
});
```

### Field Selection / Sparse Fieldsets

```typescript
// Client can request only needed fields
// GET /api/users/123?fields=id,name,email

app.get('/api/users/:id', (req, res) => {
    let query = db.users.findById(req.params.id);

    // Apply field filtering
    if (req.query.fields) {
        const fields = req.query.fields.split(',');
        query = query.select(...fields);
    }

    const user = query.get();
    res.json(user);
});
```

**Benefits**:
- Smaller response size
- Faster transfer
- Less data parsing on client

---

## Load Testing & Monitoring

### Performance Benchmarking

```typescript
import autocannon from 'autocannon';

async function benchmarkApi() {
    const result = await autocannon({
        url: 'http://localhost:3000',
        connections: 100,      // Concurrent connections
        duration: 30,           // Test duration (seconds)
        requests: [
            {
                path: '/',
                method: 'GET'
            },
            {
                path: '/api/users',
                method: 'GET'
            }
        ]
    });

    console.log('Throughput:', result.throughput);
    console.log('Latency:', result.latency);
}

// Output:
// Requests/sec: 5,234
// Throughput: 15.2 MB/s
// P99 latency: 145ms
```

### Metrics Collection

```typescript
import express from 'express';
import prometheus from 'prom-client';

const app = express();

// Create metrics
const httpRequestDuration = new prometheus.Histogram({
    name: 'http_request_duration_seconds',
    help: 'Duration of HTTP requests in seconds',
    labelNames: ['method', 'route', 'status_code'],
    buckets: [0.01, 0.05, 0.1, 0.5, 1, 2, 5]
});

const httpRequestTotal = new prometheus.Counter({
    name: 'http_requests_total',
    help: 'Total HTTP requests',
    labelNames: ['method', 'route', 'status_code']
});

// Middleware to track metrics
app.use((req, res, next) => {
    const start = Date.now();

    res.on('finish', () => {
        const duration = (Date.now() - start) / 1000;
        httpRequestDuration
            .labels(req.method, req.route?.path || req.path, res.statusCode)
            .observe(duration);

        httpRequestTotal
            .labels(req.method, req.route?.path || req.path, res.statusCode)
            .inc();
    });

    next();
});

// Expose metrics endpoint
app.get('/metrics', (req, res) => {
    res.set('Content-Type', prometheus.register.contentType);
    res.end(prometheus.register.metrics());
});
```

---

## GraphQL Optimization

### Query Depth Limiting

```typescript
function getQueryDepth(query) {
    let maxDepth = 0;

    function traverse(field, depth = 0) {
        if (!field.selectionSet) return;
        maxDepth = Math.max(maxDepth, depth);

        field.selectionSet.selections.forEach(selection => {
            traverse(selection, depth + 1);
        });
    }

    query.definitions.forEach(definition => {
        if (definition.selectionSet) {
            definition.selectionSet.selections.forEach(field => {
                traverse(field);
            });
        }
    });

    return maxDepth;
}

// Reject queries deeper than 10 levels
const validationRules = [
    ...GraphQLValidationRules,
    (context) => {
        return {
            Document(node) {
                const depth = getQueryDepth(node);
                if (depth > 10) {
                    context.reportError(
                        new GraphQLError('Query too complex - max depth 10')
                    );
                }
            }
        };
    }
];
```

### Query Cost Analysis

```typescript
const queryComplexity = new QueryComplexityPlugin({
    maximumComplexity: 1000,
    variables: {
        pageSize: 100
    },
    onComplete: (complexity) => {
        console.log(`Query complexity: ${complexity}`);
    },
    createError: (max, actual) => {
        return new GraphQLError(
            `Query complexity exceeds limit. Max: ${max}, Actual: ${actual}`
        );
    }
});

const server = new ApolloServer({
    schema,
    plugins: [queryComplexity]
});
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
