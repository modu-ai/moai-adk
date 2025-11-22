---
name: moai-lang-python
description: Python 3.12+ with FastAPI, Pydantic, async/await
allowed-tools: [Read, Bash, WebFetch]
---

## Quick Reference
**Primary Focus**: Backend, data science, automation  
**Best For**: APIs, ML, scripting  
**Key Libraries**: FastAPI, Pydantic, Pandas

**Versions**:
- Python: 3.12.x
- FastAPI: 0.109.x
- Pydantic: 2.x

## What It Does
Modern Python development for backends and data science.

## When to Use
- REST APIs with FastAPI
- Data processing with Pandas
- Machine learning projects
- Automation scripts

## Three-Level Learning
1. **Fundamentals**: See examples.md (10 examples)
2. **Advanced**: See modules/advanced-patterns.md
3. **Performance**: See modules/optimization.md

## Best Practices
### DO ✅
```python
from typing import Optional

def get_user(user_id: int) -> Optional[User]:
    return repository.find_by_id(user_id)
```

### DON'T ❌
```python
# DON'T use mutable defaults
def append_to(item, list=[]):  # Bad
    list.append(item)
    return list

# DO use None
def append_to(item, list=None):  # Good
    if list is None:
        list = []
    list.append(item)
    return list
```

## Works Well With
- `moai-domain-backend`
- `moai-essentials-perf`

## Learn More
- **Examples**: [examples.md](examples.md)
- **Advanced**: [modules/advanced-patterns.md](modules/advanced-patterns.md)
- **Performance**: [modules/optimization.md](modules/optimization.md)

---
**Version**: 3.0.0 | **Last Updated**: 2025-11-22
