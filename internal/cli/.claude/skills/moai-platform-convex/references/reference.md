# Convex Platform Reference

Extended documentation, API reference, and deployment guidance for Convex applications.

---

## Convex CLI Reference

### Development Commands

```bash
# Start development server
npx convex dev

# Deploy to production
npx convex deploy

# Run specific function
npx convex run functions/documents:list

# Generate types
npx convex codegen
```

### Environment Management

```bash
# Set environment variable
npx convex env set OPENAI_API_KEY sk-...

# List environment variables
npx convex env list

# Remove environment variable
npx convex env unset OPENAI_API_KEY
```

### Database Operations

```bash
# Export data
npx convex export --path ./backup

# Import data
npx convex import --path ./backup

# Clear all data (development only)
npx convex dev --clear
```

---

## Schema Reference

### Validator Types

Primitive Validators:
- v.string() - String values
- v.number() - Floating point numbers
- v.boolean() - Boolean values
- v.null() - Null value
- v.int64() - 64-bit integer
- v.bytes() - Binary data

Complex Validators:
- v.array(validator) - Array of specified type
- v.object({...}) - Object with specified fields
- v.union(...validators) - One of multiple types
- v.optional(validator) - Optional field
- v.literal(value) - Exact value match
- v.any() - Any value (use sparingly)

Reference Validators:
- v.id('tableName') - Document ID reference

### Index Types

Standard Index:
```typescript
defineTable({...})
  .index('index_name', ['field1', 'field2'])
```

Search Index:
```typescript
defineTable({...})
  .searchIndex('search_name', {
    searchField: 'content',
    filterFields: ['category', 'status']
  })
```

---

## Query Builder Reference

### Query Methods

```typescript
// Basic query
ctx.db.query('tableName')

// With index
.withIndex('index_name', (q) => q.eq('field', value))

// Ordering
.order('asc')  // or 'desc'

// Collect results
.collect()  // Returns array

// Get first result
.first()  // Returns single document or null

// Pagination
.paginate(paginationOpts)

// Take limited results
.take(10)
```

### Index Query Operators

```typescript
// Equality
q.eq('field', value)

// Greater than / less than
q.gt('field', value)
q.gte('field', value)
q.lt('field', value)
q.lte('field', value)

// Chaining (compound index)
q.eq('field1', value1).eq('field2', value2)
```

### Search Query Operators

```typescript
// Full-text search
q.search('fieldName', 'search terms')

// With filters
q.search('content', 'query').eq('isPublic', true)
```

---

## React Hooks Reference

### useQuery

```typescript
// Basic usage
const data = useQuery(api.module.queryName, { arg1: 'value' })

// Returns undefined while loading
// Returns null if query returns null
// Returns data when loaded

// Conditional query (skip)
const data = useQuery(
  api.module.queryName,
  condition ? { arg: value } : 'skip'
)
```

### useMutation

```typescript
// Basic usage
const mutation = useMutation(api.module.mutationName)
await mutation({ arg1: 'value' })

// With optimistic update
const mutation = useMutation(api.module.mutationName)
  .withOptimisticUpdate((localStore, args) => {
    // Update local cache
  })
```

### usePaginatedQuery

```typescript
const { results, status, loadMore } = usePaginatedQuery(
  api.module.paginatedQuery,
  { arg: value },
  { initialNumItems: 10 }
)

// status: 'LoadingFirstPage' | 'CanLoadMore' | 'LoadingMore' | 'Exhausted'
// loadMore: (numItems: number) => void
```

### useAction

```typescript
const action = useAction(api.module.actionName)
const result = await action({ arg: value })
```

---

## Function Context Reference

### Query Context (ctx)

```typescript
ctx.db          // Database access
ctx.auth        // Authentication
ctx.storage     // File storage (read-only in queries)
```

### Mutation Context (ctx)

```typescript
ctx.db          // Database access (read/write)
ctx.auth        // Authentication
ctx.storage     // File storage (read/write)
ctx.scheduler   // Schedule functions
```

### Action Context (ctx)

```typescript
ctx.auth        // Authentication
ctx.storage     // File storage (read/write)
ctx.runQuery    // Run queries
ctx.runMutation // Run mutations
ctx.runAction   // Run other actions
ctx.scheduler   // Schedule functions
```

---

## Database Operations Reference

### Read Operations

```typescript
// Get by ID
const doc = await ctx.db.get(documentId)

// Query with collect
const docs = await ctx.db.query('table').collect()

// Query with first
const doc = await ctx.db.query('table').first()
```

### Write Operations

```typescript
// Insert
const id = await ctx.db.insert('table', { field: value })

// Patch (partial update)
await ctx.db.patch(documentId, { field: newValue })

// Replace (full update)
await ctx.db.replace(documentId, { ...allFields })

// Delete
await ctx.db.delete(documentId)
```

---

## Storage Operations Reference

### Upload Flow

```typescript
// 1. Generate upload URL (mutation)
const uploadUrl = await ctx.storage.generateUploadUrl()

// 2. Upload file (client)
const response = await fetch(uploadUrl, {
  method: 'POST',
  headers: { 'Content-Type': file.type },
  body: file
})
const { storageId } = await response.json()

// 3. Get download URL (query)
const url = await ctx.storage.getUrl(storageId)

// 4. Delete file (mutation)
await ctx.storage.delete(storageId)
```

---

## Scheduler Reference

### Schedule Functions

```typescript
// Schedule after delay
await ctx.scheduler.runAfter(
  60000, // milliseconds
  internal.module.function,
  { arg: value }
)

// Schedule at specific time
await ctx.scheduler.runAt(
  timestamp, // milliseconds since epoch
  internal.module.function,
  { arg: value }
)
```

### Cancel Scheduled Function

```typescript
await ctx.scheduler.cancel(scheduledFunctionId)
```

---

## Deployment Reference

### Production Deployment

```bash
# Deploy all functions
npx convex deploy

# Deploy with production flag
npx convex deploy --prod
```

### Environment Configuration

Development (.env.local):
```
CONVEX_DEPLOYMENT=dev:your-deployment
VITE_CONVEX_URL=https://your-deployment.convex.cloud
```

Production:
```
CONVEX_DEPLOYMENT=prod:your-deployment
VITE_CONVEX_URL=https://your-deployment.convex.cloud
```

---

## Error Codes Reference

Common Convex Errors:
- UNAUTHORIZED: User not authenticated
- FORBIDDEN: User lacks permission
- NOT_FOUND: Resource not found
- VALIDATION_ERROR: Invalid input
- RATE_LIMITED: Too many requests
- INTERNAL_ERROR: Server error

---

## Performance Considerations

Query Performance:
- Always use indexes for filtered queries
- Avoid full table scans
- Use pagination for large result sets
- Limit search results with take()

Mutation Performance:
- Keep transactions small
- Batch related operations
- Use internal mutations for complex workflows

Action Performance:
- Cache external API responses when possible
- Use appropriate timeouts
- Handle rate limits gracefully

---

Version: 2.0.0
Type: Reference Documentation
Parent Skill: moai-platform-convex
