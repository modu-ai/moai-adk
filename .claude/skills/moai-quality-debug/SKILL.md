---
name: moai-quality-debug
description: "Debugging consolidated: systematic strategies, error analysis, RCA, distributed systems"
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules:
  - debugging-strategies
  - error-analysis
  - distributed-debugging
dependencies:
  - moai-foundation-trust
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - debug
  - error
  - exception
  - troubleshoot
  - diagnostics
  - failure
  - crash
  - rca
  - root-cause
  - analyze
  - trace
  - logs
agent_coverage:
  - debug-helper
  - quality-gate
context7_references: []
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Debugging Consolidated**

Unified debugging framework consolidating essentials-debug and jit-docs with systematic debugging strategies, error analysis, root cause analysis (RCA), and distributed system debugging patterns.

**Core Capabilities**:
- ✅ Systematic debugging (5 strategies)
- ✅ Root cause analysis (5 Whys, Fishbone)
- ✅ Error pattern recognition
- ✅ Distributed system debugging
- ✅ Memory leak detection
- ✅ Performance profiling
- ✅ Automated debugging workflows

**When to Use**:
- Unhandled exceptions and crashes
- Performance degradation
- Intermittent/race condition bugs
- Distributed system failures
- Memory leaks and resource issues
- Complex stack trace analysis

**Core Framework**: DETECT → REPRODUCE → ANALYZE → FIX
```
1. Error Detection & Classification
   ↓
2. Context Collection
   ↓
3. Root Cause Analysis
   ↓
4. Solution Generation
   ↓
5. Prevention Implementation
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: Binary Search Debugging

**Concept**: Narrow down bug location by eliminating half of code each iteration.

```python
# Problem: Function returns wrong value somewhere
def process_data(items):
    result = []
    for item in items:
        # Step 1: Add items to result
        result.append(item)

        # Step 2: Transform items
        result = [x * 2 for x in result]

        # Step 3: Filter items
        result = [x for x in result if x > 10]

        # Step 4: Sort items
        result.sort()

    return result

# Debug: Binary search
def test_with_debug_points():
    items = [1, 2, 3, 4, 5]

    # Add debug points to narrow down issue
    result_after_append = [item for item in items]
    print(f"After append: {result_after_append}")

    result_after_transform = [x * 2 for x in result_after_append]
    print(f"After transform: {result_after_transform}")

    # Found the issue! Filtering removes all values ≤10
    result_after_filter = [x for x in result_after_transform if x > 10]
    print(f"After filter (BUG HERE): {result_after_filter}")
```

**Use Case**: Find bug location in <5 minutes.

---

### Pattern 2: Five Whys Analysis

**Concept**: Ask "Why?" 5 times to find root cause.

```
Problem: User registration fails 50% of the time

Why 1? Database insert fails intermittently
Why 2? Connection pool exhausted under load
Why 3? Connection timeout too short (10 seconds)
Why 4? No connection retry logic in pool
Why 5? Pool configuration inherited from template, never tested under load

Root Cause: Connection pool timeout too short for slow database queries
Solution: Increase timeout to 30 seconds, add health checks
```

**Use Case**: Understand system failure root causes.

---

### Pattern 3: Rubber Duck Debugging

**Concept**: Explain code line-by-line to find bugs.

```python
# Problem: Function doesn't return expected result
def calculate_average(items):
    """Calculate average of items."""
    total = 0
    for item in items:
        total += item
    return total / len(items)  # Bug: Division by zero if len(items) == 0

# Rubber duck explanation:
# "This function takes items, initializes total to 0, then adds each item...
# Oh wait, what if items is empty? len(items) would be 0, causing division by zero!"

# Fixed version:
def calculate_average(items):
    """Calculate average of items."""
    if not items:
        return 0  # Handle empty list
    total = sum(items)
    return total / len(items)
```

**Use Case**: Find obvious bugs through explanation.

---

### Pattern 4: Fishbone Diagram Analysis

**Concept**: Visualize all potential causes of failure.

```
                       Database
                        |
                   Connection pool
                   Config timeout
                        |
                   [FAILURE]
                    /   |   \
                   /    |    \
              Code   System  External
               |      |        |
            Logic   Memory   API
            Error    Leak     Timeout
            Off-by-1 Free     Rate-limit
```

**Use Case**: Systematically identify all potential causes.

---

### Pattern 5: Automated Debugging with Logging

**Concept**: Add strategic logging to diagnose issues in production.

```python
import logging
from functools import wraps

logger = logging.getLogger(__name__)

def debug_trace(func):
    """Decorator to trace function execution."""
    @wraps(func)
    def wrapper(*args, **kwargs):
        logger.debug(f"Entering {func.__name__} with args={args}, kwargs={kwargs}")
        try:
            result = func(*args, **kwargs)
            logger.debug(f"Exiting {func.__name__} with result={result}")
            return result
        except Exception as e:
            logger.error(f"Exception in {func.__name__}: {e}", exc_info=True)
            raise

    return wrapper

@debug_trace
def process_payment(amount, user_id):
    """Process payment with automatic debugging."""
    logger.info(f"Processing payment: ${amount} for user {user_id}")
    # ... implementation ...
    logger.info(f"Payment processed successfully")

# Logs show entire flow:
# DEBUG: Entering process_payment with args=(100.00, 123)
# INFO: Processing payment: $100.00 for user 123
# INFO: Payment processed successfully
# DEBUG: Exiting process_payment with result=True
```

**Use Case**: Diagnose production issues without redeployment.

---

## Advanced Documentation

For detailed debugging patterns:

- **[modules/debugging-strategies.md](modules/debugging-strategies.md)** - 5 systematic debugging approaches
- **[modules/error-analysis.md](modules/error-analysis.md)** - Error analysis framework
- **[modules/distributed-debugging.md](modules/distributed-debugging.md)** - Multi-service debugging

---

## Best Practices

### ✅ DO
- Reproduce the bug first (non-reproducible = hard to fix)
- Start with simplest explanation
- Use systematic debugging (binary search, 5 Whys)
- Add logging for diagnosis
- Check assumptions (off-by-one, null, empty)
- Test fixes before deploying
- Document root cause
- Implement prevention strategies

### ❌ DON'T
- Guess at fixes (use scientific method)
- Change multiple things simultaneously
- Ignore error messages
- Skip error context (logs, stack traces)
- Debug in production without logging
- Assume infrastructure is correct
- Deploy untested fixes

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
