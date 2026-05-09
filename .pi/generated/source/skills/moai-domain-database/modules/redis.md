# Redis Advanced Patterns

## Overview
Advanced Redis patterns covering caching strategies, data structures, real-time analytics, distributed locking, and performance optimization for high-throughput applications.

## Quick Implementation

### Caching Strategies

```javascript
class CacheManager {
 constructor(redisClient) {
 this.redis = redisClient;
 }

 // Multi-layer caching with fallback
 async getWithFallback(key, fetchFunction, ttl = 3600) {
 // Try memory cache first
 const memoryCache = this.getMemoryCache(key);
 if (memoryCache) return memoryCache;

 // Try Redis cache
 const redisCache = await this.redis.get(key);
 if (redisCache) {
 const data = JSON.parse(redisCache);
 this.setMemoryCache(key, data, ttl / 10); // Shorter memory TTL
 return data;
 }

 // Fetch from source
 const data = await fetchFunction();

 // Set both caches
 await this.redis.setex(key, ttl, JSON.stringify(data));
 this.setMemoryCache(key, data, ttl / 10);

 return data;
 }

 // Write-through caching
 async setWithWriteThrough(key, data, fetchFunction, ttl = 3600) {
 // Update source first
 await fetchFunction(data);

 // Update caches
 const pipeline = this.redis.pipeline();
 pipeline.setex(key, ttl, JSON.stringify(data));

 // Invalidate related cache keys
 const relatedKeys = await this.getRelatedKeys(key);
 relatedKeys.forEach(relatedKey => {
 pipeline.del(relatedKey);
 });

 await pipeline.exec();
 this.setMemoryCache(key, data, ttl / 10);
 }

 // Rate limiting with sliding window
 async checkRateLimit(key, limit, windowMs) {
 const now = Date.now();
 const pipeline = this.redis.pipeline();

 // Remove old entries
 pipeline.zremrangebyscore(key, 0, now - windowMs);

 // Add current request
 pipeline.zadd(key, now, now);

 // Count current window requests
 pipeline.zcard(key);

 // Set expiration
 pipeline.expire(key, Math.ceil(windowMs / 1000));

 const results = await pipeline.exec();
 const currentCount = results[2][1];

 return {
 allowed: currentCount <= limit,
 count: currentCount,
 remaining: Math.max(0, limit - currentCount),
 resetTime: now + windowMs
 };
 }
}
```

### Advanced Data Structures

```javascript
class RedisDataManager {
 constructor(redisClient) {
 this.redis = redisClient;
 }

 // Leaderboard with time decay
 async addToTimeDecayLeaderboard(key, memberId, score, decayRate = 0.95) {
 const timestamp = Date.now();
 const decayedScore = score * Math.pow(decayRate, (Date.now() - timestamp) / (1000 * 60 * 60));

 await this.redis.zadd(key, decayedScore, memberId);

 // Remove old entries
 const weekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000);
 await this.redis.zremrangebyscore(key, 0, weekAgo);
 }

 // Real-time analytics with HyperLogLog
 async trackUniqueVisitors(pageKey, userId) {
 // Track unique users with HyperLogLog
 await this.redis.pfadd(`${pageKey}:unique`, userId);

 // Track total visits with regular counter
 await this.redis.incr(`${pageKey}:total`);

 // Track user activity set for recent activity
 const activityKey = `${pageKey}:activity:${Math.floor(Date.now() / (60 * 1000))}`;
 await this.redis.sadd(activityKey, userId);
 await this.redis.expire(activityKey, 300); // 5 minutes
 }

 async getAnalytics(pageKey) {
 const pipeline = this.redis.pipeline();

 // Unique visitors estimate
 pipeline.pfcount(`${pageKey}:unique`);

 // Total page views
 pipeline.get(`${pageKey}:total`);

 // Recent active users (last 5 minutes)
 const now = Math.floor(Date.now() / (60 * 1000));
 for (let i = 0; i < 5; i++) {
 pipeline.scard(`${pageKey}:activity:${now - i}`);
 }

 const results = await pipeline.exec();

 const uniqueVisitors = results[0][1];
 const totalViews = parseInt(results[1][1]) || 0;
 const recentActivity = results.slice(2).map(r => r[1]).reduce((a, b) => a + b, 0);

 return {
 uniqueVisitors,
 totalViews,
 recentActiveUsers: recentActivity
 };
 }

 // Distributed locking with timeout and retry
 async acquireLock(key, ttl = 30000, retryCount = 3, retryDelay = 100) {
 const lockKey = `lock:${key}`;
 const lockValue = `${Date.now()}-${Math.random()}`;

 for (let attempt = 0; attempt < retryCount; attempt++) {
 const result = await this.redis.set(
 lockKey,
 lockValue,
 'PX', ttl,
 'NX'
 );

 if (result === 'OK') {
 return {
 acquired: true,
 lockValue,
 release: async () => await this.releaseLock(lockKey, lockValue)
 };
 }

 // Wait before retry
 if (attempt < retryCount - 1) {
 await new Promise(resolve => setTimeout(resolve, retryDelay));
 }
 }

 return { acquired: false };
 }

 async releaseLock(lockKey, lockValue) {
 const script = `
 if redis.call("get", KEYS[1]) == ARGV[1] then
 return redis.call("del", KEYS[1])
 else
 return 0
 end
 `;

 return await this.redis.eval(script, 1, lockKey, lockValue);
 }
}
```

## Key Features

### 1. Caching Patterns
- Multi-layer caching (memory + Redis)
- Cache-aside and write-through strategies
- Cache invalidation and warming
- Rate limiting and throttling

### 2. Data Structures
- Strings for simple key-value storage
- Hashes for object storage
- Lists for queues and stacks
- Sets for unique collections
- Sorted sets for leaderboards
- HyperLogLog for cardinality estimation

### 3. Real-time Features
- Pub/sub messaging
- Streams for event processing
- Geospatial indexing
- Bitmap operations

### 4. Performance Optimization
- Pipeline operations for batch processing
- Lua scripting for atomic operations
- Connection pooling strategies
- Memory optimization techniques

## Advanced Patterns

### Distributed Locking
- Redlock algorithm for safety
- Timeout and retry mechanisms
- Lock release verification
- Deadlock prevention

### Rate Limiting
- Sliding window implementation
- Fixed window counters
- Token bucket algorithm
- Distributed rate limiting

### Analytics and Counting
- Real-time metrics collection
- Time-series data storage
- Approximate counting with HyperLogLog
- Bitmap analytics for user behavior

### Session Management
- User session storage
- Shopping cart management
- Real-time collaboration
- Presence detection

## Best Practices
- Use appropriate data structures
- Implement proper key naming conventions
- Set appropriate TTLs for cache entries
- Monitor memory usage and performance
- Use pipelining for batch operations
- Implement proper error handling

## Integration Points
- Connection pooling libraries (ioredis, redis-py)
- Cache middleware for web frameworks
- Message queue systems
- Real-time analytics platforms
- Session management systems
