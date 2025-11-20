---
name: moai-core-proactive-suggestions
version: 4.0.0
updated: 2025-11-20
status: stable
tier: core
description: Proactive suggestions for code quality, security, and best practices
allowed-tools: [Read, Bash]
---

# Proactive Suggestions Expert

**Intelligent Code Improvement Recommendations**

> **Focus**: Code Quality, Security, Performance, Best Practices  
> **Approach**: Analyze → Suggest → Prioritize

---

## Overview

Automatically detect improvement opportunities and provide actionable suggestions.

### Suggestion Categories

1.  **Code Quality**: Readability, maintainability
2.  **Security**: Vulnerabilities, best practices
3.  **Performance**: Bottlenecks, optimization opportunities
4.  **Architecture**: Design patterns, structure

---

## Suggestion Patterns

### 1. Code Quality Suggestions

**Pattern: Overly Complex Function**

```python
# Detected Issue
def process_data(data):
    if len(data) > 0:
        for item in data:
            if item['status'] == 'active':
                if item['type'] == 'user':
                    # 20+ more lines...

# Suggestion
✅ **Simplify Logic**:
- Extract nested loops into separate functions
- Use early returns to reduce nesting
- Current complexity: 15 → Target: <10

# Improved
def process_data(data):
    active_users = filter_active_users(data)
    return [process_user(user) for user in active_users]
```

### 2. Security Suggestions

**Pattern: Hardcoded Secrets**

```python
# Detected Issue
API_KEY = "sk-1234567890abcdef"

# Suggestion
❌ **Security Risk: Hardcoded Secret**
- Move to environment variable
- Use secrets management (e.g., .env, AWS Secrets Manager)
- Add .env to .gitignore

# Fixed
import os
API_KEY = os.getenv("API_KEY")
```

### 3. Performance Suggestions

**Pattern: N+1 Query**

```python
# Detected Issue
users = User.query.all()
for user in users:
    posts = Post.query.filter_by(user_id=user.id).all()

# Suggestion
⚡ **Performance: N+1 Query Detected**
- Use eager loading (joinedload)
- Reduces queries from N+1 to 1

# Optimized
users = User.query.options(joinedload(User.posts)).all()
```

### 4. Best Practice Suggestions

**Pattern: Missing Type Hints**

```python
# Detected Issue
def calculate_total(items):
    return sum(item.price for item in items)

# Suggestion
📝 **Best Practice: Add Type Hints**
- Improves IDE autocomplete
- Catches type errors early
- Self-documenting code

# Improved
from typing import List

def calculate_total(items: List[Item]) -> float:
    return sum(item.price for item in items)
```

---

## Suggestion Priority Levels

### Critical (🔴 Fix Immediately)

- Security vulnerabilities
- Data loss risks
- Production bugs

### High (🟡 Fix Soon)

- Performance bottlenecks (>2x improvement possible)
- Code smells affecting maintainability
- Missing error handling

### Medium (🟢 Consider)

- Code style inconsistencies
- Missing documentation
- Potential refactoring

### Low (ℹ️ Optional)

- Minor optimizations
- Cosmetic improvements

---

## Suggestion Workflow

1.  **Analyze**: Scan code for patterns
2.  **Detect**: Identify improvement opportunities
3.  **Prioritize**: Rank by impact and effort
4.  **Suggest**: Provide actionable recommendations
5.  **Track**: Monitor adoption rate

---

## Example Suggestions

### Missing Error Handling

```python
# Current
def fetch_data(url):
    response = requests.get(url)
    return response.json()

# Suggestion
🛡️ **Add Error Handling**
- Network failures not handled
- Invalid JSON not handled

# Improved
def fetch_data(url):
    try:
        response = requests.get(url, timeout=5)
        response.raise_for_status()
        return response.json()
    except requests.Timeout:
        logger.error(f"Timeout fetching {url}")
        raise
    except requests.HTTPError as e:
        logger.error(f"HTTP error: {e}")
        raise
```

---

## Validation Checklist

- [ ] **Actionable**: Suggestions include concrete fixes?
- [ ] **Prioritized**: Critical issues flagged first?
- [ ] **Justified**: Explanation provided for each suggestion?
- [ ] **Measurable**: Impact quantified (e.g., "2x faster")?

---

## Related Skills

- `moai-essentials-review`: Code review
- `moai-essentials-refactor`: Refactoring patterns
- `moai-security-devsecops`: Security scanning

---

**Last Updated**: 2025-11-20
