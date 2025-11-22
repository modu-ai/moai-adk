---
name: moai-lang-typescript
description: TypeScript 5.x with static typing, generics, advanced types
version: 1.0.0
modularized: true
allowed-tools:
  - Read
  - Bash
  - WebFetch
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - lang
  - typescript
category_tier: 1
---

## Quick Reference
**Primary Focus**: Type-safe JavaScript development  
**Best For**: Large-scale apps, enterprise projects  
**Key Features**: Static types, generics, interfaces

**Versions**:
- TypeScript: 5.3.x
- Node.js: 20 LTS

## What It Does
Static typing for JavaScript with advanced type system.

## When to Use
- Enterprise applications
- Large codebases
- Team projects requiring type safety

## Three-Level Learning
1. **Fundamentals**: See examples.md (10 examples)
2. **Advanced**: See modules/advanced-patterns.md
3. **Performance**: See modules/optimization.md

## Best Practices
### DO ✅
```typescript
interface User {
    id: number;
    name: string;
}

function getUser(id: number): User {
    return { id, name: "John" };
}
```

### DON'T ❌
```typescript
// DON'T use any
function process(data: any) {} // Use specific types

// DON'T ignore null checks
function getName(user: User) {
    return user.name.toUpperCase(); // Check for null first
}
```

## Works Well With
- `moai-lang-javascript`
- `moai-domain-frontend`

## Learn More
- **Examples**: [examples.md](examples.md)
- **Advanced**: [modules/advanced-patterns.md](modules/advanced-patterns.md)
- **Performance**: [modules/optimization.md](modules/optimization.md)

---
**Version**: 3.0.0 | **Last Updated**: 2025-11-22