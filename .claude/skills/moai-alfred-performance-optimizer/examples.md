# Performance Optimization Examples

_Last updated: 2025-10-22_

## Example 1: Python CPU Profiling

### Scenario
Identifying slow function in data processing pipeline.

### Profiling Code
```python
import cProfile
import pstats

def profile_function():
    profiler = cProfile.Profile()
    profiler.enable()

    # Code to profile
    process_large_dataset()

    profiler.disable()
    stats = pstats.Stats(profiler)
    stats.sort_stats('cumulative')
    stats.print_stats(10)

# Output:
#    ncalls  tottime  percall  cumtime  percall filename:lineno(function)
#         1    0.001    0.001    5.234    5.234 processor.py:15(process_large_dataset)
#    100000    3.456    0.000    3.456    0.000 utils.py:42(validate_item)  # ← Bottleneck
#     50000    1.123    0.000    1.123    0.000 parser.py:78(parse_json)
```

### Optimization
```python
# Before (slow): Validation in loop
def process_large_dataset(items):
    for item in items:
        if validate_item(item):  # Called 100k times
            process(item)

# After (fast): Batch validation
def process_large_dataset(items):
    valid_items = bulk_validate(items)  # Single call
    for item in valid_items:
        process(item)

# Result: 3.5s → 0.2s (94% faster)
```

---

## Example 2: Memory Profiling

### Scenario
Identifying memory leaks in long-running service.

### Profiling Code
```python
from memory_profiler import profile

@profile
def load_user_data():
    users = []
    for i in range(10000):
        user = fetch_user(i)
        users.append(user)  # Memory accumulation
    return users

# Output:
# Line    Mem usage    Increment  Occurrences   Line Contents
# ====    =========    =========  ===========   ==============
#    3     50.2 MiB     50.2 MiB           1       users = []
#    4    150.8 MiB     10.1 MiB       10000       user = fetch_user(i)
#    5    250.4 MiB     99.6 MiB       10000       users.append(user)  # ← Memory issue
```

### Optimization
```python
# Before: Load all into memory
def load_user_data():
    return [fetch_user(i) for i in range(10000)]  # 250 MB

# After: Use generator (lazy evaluation)
def load_user_data():
    return (fetch_user(i) for i in range(10000))  # ~50 MB

# Process in batches
for user in load_user_data():
    process(user)  # Memory released after each iteration
```

---

## Example 3: Database Query Optimization

### Scenario
Slow API endpoint due to N+1 query problem.

### Before (N+1 Problem)
```python
# Fetches 1 query for users + N queries for posts
def get_users_with_posts():
    users = User.objects.all()  # 1 query
    result = []
    for user in users:
        posts = user.posts.all()  # N queries (one per user)
        result.append({
            'user': user,
            'posts': list(posts)
        })
    return result

# Result: 101 queries for 100 users (slow!)
```

### After (Eager Loading)
```python
# Single query with JOIN
def get_users_with_posts():
    users = User.objects.prefetch_related('posts').all()  # 2 queries total
    result = []
    for user in users:
        result.append({
            'user': user,
            'posts': list(user.posts.all())  # No additional query
        })
    return result

# Result: 2 queries for 100 users (50x faster!)
```

---

## Example 4: Frontend Performance

### Scenario
Slow React component re-renders.

### Before (Unnecessary Re-renders)
```typescript
function UserList({ users }) {
  return (
    <div>
      {users.map(user => (
        <UserCard key={user.id} user={user} />  // Re-renders on every parent update
      ))}
    </div>
  );
}
```

### After (Memoization)
```typescript
import { memo } from 'react';

const UserCard = memo(({ user }) => {
  return <div>{user.name}</div>;
});

function UserList({ users }) {
  return (
    <div>
      {users.map(user => (
        <UserCard key={user.id} user={user} />  // Only re-renders if user changes
      ))}
    </div>
  );
}

// Result: 90% reduction in re-renders
```

---

_For profiling tools and strategies, see reference.md_
