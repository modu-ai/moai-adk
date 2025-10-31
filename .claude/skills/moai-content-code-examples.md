# moai-content-code-examples

Curating and writing code examples for documentation, tutorials, and blog posts.

## Quick Start

Good code examples make documentation useful and tutorials effective. Use this skill when writing sample code, creating code snippets, or documenting APIs.

## Core Patterns

### Pattern 1: Example Progression

**Pattern**: Show examples from simple to complex.

```typescript
// Example 1: Simplest use case
const user = await getUser('123');
console.log(user.name);

// Example 2: With error handling
try {
  const user = await getUser('123');
  console.log(user.name);
} catch (error) {
  console.error('Failed to load user:', error);
}

// Example 3: With validation and retries
async function getWithRetry(userId: string, maxRetries = 3) {
  if (!userId) throw new Error('User ID required');

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await getUser(userId);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await delay(1000 * (i + 1)); // Exponential backoff
    }
  }
}

// Example 4: Production patterns with logging
async function getUserProduction(userId: string) {
  logger.info('Fetching user', { userId });

  try {
    const user = await getWithRetry(userId);
    metrics.increment('user.fetched');
    return user;
  } catch (error) {
    logger.error('Failed to fetch user', { userId, error });
    metrics.increment('user.fetch_failed');
    throw error;
  }
}
```

**When to use**:
- Teaching new concepts
- Building from basics to advanced
- Progressive learning

**Key benefits**:
- Gradual complexity increase
- Clear learning path
- Practical application

### Pattern 2: Common Use Cases

```python
# Use Case 1: Simple data processing
data = load_data('file.csv')
processed = [clean(row) for row in data]

# Use Case 2: Batch processing
results = []
for batch in batches(data, size=100):
    processed = parallel_process(batch)
    results.extend(processed)

# Use Case 3: Streaming with error handling
def process_stream(stream, handler):
    for item in stream:
        try:
            result = transform(item)
            handler(result)
        except ProcessingError as e:
            log_error(e, item)
            continue

# Use Case 4: Async processing with monitoring
async def process_async(items):
    tasks = [process_item(item) for item in items]
    results = await asyncio.gather(*tasks, return_exceptions=True)

    for result in results:
        if isinstance(result, Exception):
            handle_error(result)
```

**When to use**:
- Showing practical applications
- Demonstrating different scenarios
- Real-world patterns

**Key benefits**:
- Applicable to actual projects
- Shows when to use what
- Practical guidance

## Progressive Disclosure

### Level 1: Simple Examples
- One-liners
- Basic usage
- Happy path only

### Level 2: Practical Examples
- Error handling
- Edge cases
- Common patterns

### Level 3: Production Examples
- Logging and monitoring
- Performance optimization
- Advanced patterns

## Works Well With

- **Syntax highlighting**: For code visibility
- **Runnable code**: Verify examples work
- **Code comments**: Explain key lines
- **Diagrams**: Show flow with code
- **Tests**: Validate examples

## References

- **GitHub Examples**: https://github.com/
- **Official Documentation**: Platform-specific docs
- **Stack Overflow**: Real-world use cases
- **Dev.to**: Community examples
