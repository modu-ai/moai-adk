# MongoDB Advanced Patterns

## Overview
Advanced MongoDB patterns covering document modeling, aggregation pipelines, indexing strategies, and performance optimization for scalable NoSQL applications.

## Quick Implementation

### Document Modeling

```javascript
// User schema with embedded and referenced patterns
const userSchema = new Schema({
 _id: ObjectId,
 username: { type: String, required: true, unique: true },
 email: { type: String, required: true, unique: true },
 profile: {
 displayName: String,
 bio: String,
 avatar: String,
 preferences: {
 theme: { type: String, enum: ['light', 'dark'], default: 'light' },
 language: { type: String, default: 'en' },
 notifications: {
 email: { type: Boolean, default: true },
 push: { type: Boolean, default: false }
 }
 }
 },
 security: {
 passwordHash: String,
 lastLoginAt: Date,
 failedLoginAttempts: { type: Number, default: 0 },
 lockedUntil: Date
 },
 activity: {
 lastSeenAt: Date,
 loginCount: { type: Number, default: 0 }
 }
}, {
 timestamps: true,
 // Optimized indexes
 index: [
 { username: 1 },
 { email: 1 },
 { 'security.lastLoginAt': -1 },
 { 'activity.lastSeenAt': -1 }
 ]
});

// Post schema with comments embedded for performance
const postSchema = new Schema({
 _id: ObjectId,
 authorId: { type: ObjectId, ref: 'User', required: true },
 title: { type: String, required: true },
 content: { type: String, required: true },
 tags: [String],
 metadata: {
 viewCount: { type: Number, default: 0 },
 likeCount: { type: Number, default: 0 },
 commentCount: { type: Number, default: 0 }
 },
 // Embed recent comments for performance
 recentComments: [{
 _id: ObjectId,
 authorId: ObjectId,
 authorName: String,
 content: String,
 createdAt: { type: Date, default: Date.now }
 }],
 status: { type: String, enum: ['draft', 'published', 'archived'], default: 'draft' }
}, {
 timestamps: true,
 index: [
 { authorId: 1, createdAt: -1 },
 { status: 1, createdAt: -1 },
 { tags: 1 },
 { 'metadata.viewCount': -1 }
 ]
});
```

### Advanced Aggregation Pipelines

```javascript
// User analytics with complex aggregation
const getUserAnalytics = async (userId, timeRange) => {
 return await User.aggregate([
 // Match specific user
 { $match: { _id: userId } },

 // Lookup posts and engagement
 {
 $lookup: {
 from: 'posts',
 localField: '_id',
 foreignField: 'authorId',
 as: 'posts'
 }
 },

 // Unwind posts for processing
 { $unwind: '$posts' },

 // Filter by time range
 {
 $match: {
 'posts.createdAt': {
 $gte: timeRange.start,
 $lte: timeRange.end
 }
 }
 },

 // Group by user and calculate metrics
 {
 $group: {
 _id: '$_id',
 username: { $first: '$username' },
 totalPosts: { $sum: 1 },
 totalViews: { $sum: '$posts.metadata.viewCount' },
 totalLikes: { $sum: '$posts.metadata.likeCount' },
 avgViewsPerPost: { $avg: '$posts.metadata.viewCount' },
 tags: { $push: '$posts.tags' }
 }
 },

 // Flatten and count tags
 {
 $addFields: {
 allTags: { $reduce: {
 input: '$tags',
 initialValue: [],
 in: { $concatArrays: ['$$value', '$$this'] }
 }}
 }
 },

 {
 $addFields: {
 uniqueTags: { $setUnion: ['$allTags', []] },
 topTags: { $slice: [
 { $sortArray: { input: { $objectToArray: { $size: '$allTags' } }, sortBy: { count: -1 } } },
 5
 ]}
 }
 }
 ]);
};

// Time series aggregation for activity tracking
const getActivityTrends = async (userId, granularity = 'daily') => {
 const groupFormat = granularity === 'daily' ?
 { $dateToString: { format: '%Y-%m-%d', date: '$createdAt' } } :
 { $dateToString: { format: '%Y-%U', date: '$createdAt' } };

 return await Activity.aggregate([
 { $match: { userId } },
 {
 $group: {
 _id: groupFormat,
 totalEvents: { $sum: 1 },
 eventTypes: { $addToSet: '$eventType' },
 uniqueSessions: { $addToSet: '$sessionId' }
 }
 },
 {
 $addFields: {
 eventCount: { $size: '$eventTypes' },
 sessionCount: { $size: '$uniqueSessions' }
 }
 },
 { $sort: { _id: 1 } }
 ]);
};
```

## Key Features

### 1. Document Design Patterns
- Embedding vs referencing strategies
- Denormalization for read performance
- Schema validation and constraints
- Array and object modeling

### 2. Indexing Strategies
- Compound indexes for multi-field queries
- Text indexes for full-text search
- Geospatial indexes for location data
- TTL indexes for automatic expiration

### 3. Aggregation Framework
- Multi-stage data processing pipelines
- Lookup operations for joins
- Group and accumulate operations
- Window functions for analytics

### 4. Performance Optimization
- Query pattern analysis
- Index usage optimization
- Connection pooling strategies
- Sharding for horizontal scaling

## Design Patterns

### Embedding Patterns
- One-to-One: Embed related data
- One-to-Many: Embed for small arrays
- Many-to-Many: Use references with lookups

### Reference Patterns
- Population for related data
- Manual joins with aggregation
- Caching frequently accessed data

### Schema Evolution
- Version-controlled schema changes
- Backward compatibility strategies
- Data migration approaches

## Best Practices
- Design for query patterns
- Use appropriate data types
- Implement proper indexing
- Monitor performance metrics
- Plan for schema evolution

## Integration Points
- ORMs (Mongoose, TypeGOOSE)
- ODMs for Node.js applications
- Connection pooling libraries
- Monitoring and analytics tools
