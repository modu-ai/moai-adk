---
name: moai-lang-javascript
description: JavaScript ES2024+ with async/await, modules, classes
version: 1.0.0
modularized: true
allowed-tools:
  - Read
  - Bash
  - WebFetch
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - javascript
  - lang
  - python
category_tier: 1
---

## Quick Reference
**Primary Focus**: Web development, Node.js, frontend  
**Best For**: SPAs, APIs, full-stack apps  
**Key Libraries**: React, Express, Axios

**Versions**:
- JavaScript: ES2024
- Node.js: 20 LTS

## What It Does
Modern JavaScript development for web and Node.js.

## When to Use
- Web app development
- Node.js backends
- Frontend frameworks

## Three-Level Learning
1. **Fundamentals**: See examples.md (10 examples)
2. **Advanced**: See modules/advanced-patterns.md
3. **Performance**: See modules/optimization.md

## Best Practices
### DO ✅
```javascript
async function fetchData() {
    const res = await fetch('/api');
    return res.json();
}
```

### DON'T ❌
```javascript
// DON'T use var
var x = 1; // Use let/const

// DON'T ignore promises
fetch('/api'); // Always handle with await or .then()
```

## Works Well With
- `moai-lang-typescript`
- `moai-domain-frontend`

## Learn More
- **Examples**: [examples.md](examples.md)
- **Advanced**: [modules/advanced-patterns.md](modules/advanced-patterns.md)
- **Performance**: [modules/optimization.md](modules/optimization.md)

---
**Version**: 3.0.0 | **Last Updated**: 2025-11-22