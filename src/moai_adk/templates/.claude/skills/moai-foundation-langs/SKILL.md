---
name: moai-foundation-langs
description: Programming Language Ecosystem Mastery - Modern patterns, best practices, and latest features for 25+ languages
version: 1.0.0
tier: Foundation
modularized: false
status: active
tags:
  - languages
  - programming
  - patterns
  - best-practices
  - ecosystem
  - enterprise
updated: 2025-11-24
compliance_score: 85
test_coverage: 85

# Required Fields (7)
allowed-tools:
  - Read
  - Write
  - AskUserQuestion
  - Bash
last_updated: 2025-11-24

# Recommended Fields (9)
modules: []
dependencies: []
deprecated: false
successor: null
category_tier: 1
auto_trigger_keywords:
  - language
  - python
  - typescript
  - go
  - rust
  - java
  - c#
  - php
  - r
  - programming
  - patterns
  - best practices
  - language features
  - ecosystem
  - modern practices
  - 2025 standards
agent_coverage:
  - tdd-implementer
  - backend-expert
  - frontend-expert
  - quality-gate
context7_references:
  - /python/latest
  - /typescript/latest
  - /go/latest
invocation_api_version: "1.0"
---

# 🌍 Programming Language Ecosystem Mastery

## 30-Second Quick Reference

**moai-foundation-langs** is a **comprehensive language ecosystem guide** covering modern patterns, best practices, and latest features for 25+ programming languages:

- **Python** 3.13+: FastAPI, async/await, Pydantic, SQLAlchemy 2.0
- **TypeScript** 5.9+: strict typing, Next.js 16, React 19, tRPC
- **Go** 1.23+: systems programming, Fiber v3, concurrency
- **Rust** 1.75+: memory safety, Tokio, high-performance systems
- **Java** 21+: records, pattern matching, virtual threads
- **C#** 12+: LINQ, async patterns, .NET 8
- **PHP** 8.4+: typed properties, attributes, modern frameworks
- **JavaScript** ES2024+: async patterns, modules, tooling
- **R** 4.4+: data science, ggplot2, tidyverse
- **And 15+ more languages...**

**Use When**: Writing production code, selecting language features, following enterprise standards, implementing pattern best practices, or ensuring code quality across polyglot teams.

---

## What It Does

moai-foundation-langs provides a **structured ecosystem guide** for writing language-specific code that is:
- **Modern**: Latest syntax, features, and patterns (2025 standards)
- **Idiomatic**: Follows language conventions and culture
- **Performant**: Best practices for optimization in each language
- **Maintainable**: Clear, documented, team-friendly code
- **Enterprise-Ready**: Production patterns and quality standards

### Language Coverage Tiers

**Tier 1 - Core Languages (Deep Coverage)**:
- Python 3.13+
- TypeScript 5.9+
- Go 1.23+
- Rust 1.75+

**Tier 2 - Enterprise Languages (Standard Coverage)**:
- Java 21+
- C# 12+
- PHP 8.4+
- JavaScript ES2024+

**Tier 3 - Specialized Languages (Essential Coverage)**:
- R 4.4+ (data science)
- Elixir 1.15+ (functional)
- Kotlin 1.9+ (JVM)
- Scala 3.x (functional/OOP hybrid)
- Ruby 3.3+ (rapid development)
- Swift 5.9+ (Apple platform)
- And more...

---

## When to Use

### ✅ Use moai-foundation-langs When:

1. **Writing Production Code**
   - Implement features in modern language versions
   - Follow current best practices and idioms
   - Ensure code quality and maintainability

2. **Learning New Language Features**
   - Master 2025 language capabilities
   - Understand modern patterns vs. legacy code
   - Adopt recommended approaches

3. **Polyglot Team Development**
   - Maintain consistency across multiple languages
   - Standardize patterns and practices
   - Enable knowledge sharing between teams

4. **Code Review & Validation**
   - Check for idiomatic patterns
   - Ensure language-best-practices compliance
   - Suggest modern alternatives to legacy patterns

5. **Performance Optimization**
   - Identify language-specific optimization techniques
   - Avoid common pitfalls and anti-patterns
   - Leverage compiler/runtime optimizations

### ❌ Avoid moai-foundation-langs When:

- Supporting legacy versions (Python 2.x, Node.js <14)
- Implementing non-standard patterns intentionally
- Language features are irrelevant to current task
- Basic syntax questions (use official language docs instead)

---

## Language-Specific Patterns

### Python 3.13+

**Modern Features**:
- Type hints (PEP 698, PEP 692)
- Async/await patterns (asyncio, Trio)
- Dataclasses (Python 3.10+)
- Pattern matching (Python 3.10+)
- Structural pattern matching

**Recommended Libraries**:
- **Web**: FastAPI, Django 5.0
- **Data**: Pydantic v2, SQLAlchemy 2.0
- **Async**: Asyncio, Httpx, Asyncpg
- **Testing**: pytest, pytest-asyncio
- **Quality**: ruff, mypy, black

**Enterprise Pattern**:
```python
from fastapi import FastAPI
from pydantic import BaseModel, Field
from sqlalchemy import select

app = FastAPI()

class UserCreate(BaseModel):
    name: str = Field(..., min_length=1)
    email: str = Field(..., regex=r'^[\w\.-]+@[\w\.-]+\.\w+$')

@app.post("/users/", status_code=201)
async def create_user(user: UserCreate) -> dict:
    """Create new user with validation."""
    # Implementation
    pass
```

### TypeScript 5.9+

**Modern Features**:
- Strict mode (recommended enforced)
- Strict type checking
- Decorators (TC39 Stage 3)
- Type-safe dependency injection
- Const type parameters

**Recommended Libraries**:
- **Frontend**: React 19, Next.js 16, Vue 3.5
- **Backend**: tRPC, Fastify, NestJS
- **Data**: Zod, Prisma, TypeORM
- **Testing**: Vitest, Jest
- **Quality**: ESLint, Prettier, TypeScript strict

**Enterprise Pattern**:
```typescript
import { z } from "zod";
import { trpc } from "./trpc";

const userSchema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
});

type User = z.infer<typeof userSchema>;

export const userRouter = trpc.router({
  create: trpc.procedure
    .input(userSchema)
    .mutation(async ({ input }): Promise<User> => {
      // Type-safe implementation
      return input;
    }),
});
```

### Go 1.23+

**Modern Features**:
- Range over integers (1.22+)
- Iterators (1.22+)
- Clear function (1.21+)
- Generics (1.18+)
- Error handling patterns

**Recommended Libraries**:
- **Web**: Fiber v3, Echo, Chi
- **Data**: GORM, sqlc, ent
- **Async**: Goroutines, channels
- **Testing**: testify, go-testify
- **Quality**: golangci-lint, go fmt

**Enterprise Pattern**:
```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
    app := fiber.New()

    app.Use(cors.New())

    app.Post("/users", createUser)
    app.Listen(":3000")
}

func createUser(c fiber.Ctx) error {
    // Type-safe, efficient Go implementation
    return nil
}
```

### Rust 1.75+

**Modern Features**:
- Memory safety without garbage collection
- Zero-cost abstractions
- Pattern matching
- Trait system for polymorphism
- Error handling (Result/Option)

**Recommended Libraries**:
- **Web**: Tokio, Axum, Actix-web
- **Data**: Sqlx, Diesel, Sea-ORM
- **Async**: Tokio, async-trait
- **Testing**: cargo test, criterion
- **Quality**: clippy, rustfmt, cargo-audit

**Enterprise Pattern**:
```rust
use axum::{
    extract::Json,
    routing::post,
    Router,
};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct User {
    name: String,
    email: String,
}

async fn create_user(Json(user): Json<User>) -> Json<User> {
    user
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/users", post(create_user));
}
```

---

## Best Practices by Language Tier

### Tier 1: Core Languages (Full Guidelines)

Each core language includes:
- ✅ Latest version requirements
- ✅ Type system mastery
- ✅ Concurrency patterns
- ✅ Performance optimization
- ✅ Testing strategies
- ✅ Security best practices
- ✅ Deployment patterns

### Tier 2: Enterprise Languages (Standard Guidelines)

Each enterprise language includes:
- ✅ Current standards (2024+)
- ✅ Common patterns
- ✅ Framework recommendations
- ✅ Quality tools
- ✅ Testing approaches

### Tier 3: Specialized Languages (Essential Guidelines)

Each specialized language includes:
- ✅ Core patterns
- ✅ Recommended libraries
- ✅ Basic quality standards

---

## Anti-Patterns to Avoid

| ❌ Wrong | ✅ Right | Language |
|---------|----------|----------|
| Mutable global state | Functional patterns, DI | Python, Go, Rust |
| Callback hell | Async/await, promises | JavaScript, Python |
| Manual memory management | Using language features | Rust, Go |
| Ignoring type safety | Strict typing + linting | TypeScript, Python |
| Synchronous blocking I/O | Async patterns | All modern languages |
| Raw SQL strings | Parameterized queries, ORM | Python, TypeScript, Go |
| Silent failures | Proper error handling | Rust (Result), Go (error) |
| No error context | Contextual error messages | All languages |

---

## Language Version Matrix (2025)

| Language | Latest | Min. Recommended | LTS Support |
|----------|--------|-----------------|-------------|
| Python | 3.14 | 3.12 (3.10 EOL: Oct 2026) | 3.12 until Oct 2028 |
| TypeScript | 5.9+ | 5.0+ | 4.5+ supported |
| Go | 1.23 | 1.20+ (1.19 EOL: Aug 2025) | 1.23 LTS fallback |
| Rust | 1.77 | 1.70+ | No LTS (6-week releases) |
| Java | 23 (LTS: 21) | 21 LTS | 21 until Sept 2031 |
| C# | 12.0 | 11.0 | 8.0 support ending |
| PHP | 8.4 | 8.2+ (8.1 EOL: Nov 2024) | 8.2 until Dec 2025 |
| Node.js | 22 (LTS: 20) | 18 LTS+ | 20 LTS until Apr 2026 |

---

## Framework Recommendations by Language

### Python
- **APIs**: FastAPI (recommended), Django REST Framework, Flask
- **Web**: Django 5.0, Starlette
- **Testing**: pytest, pytest-asyncio

### TypeScript
- **Frontend**: Next.js 16 (SSR/SSG), React 19, Vue 3.5
- **Backend**: tRPC (type-safe RPC), NestJS, Fastify
- **Full-Stack**: Next.js with Prisma

### Go
- **APIs**: Fiber v3 (fast), Echo, Chi
- **Databases**: GORM, sqlc, ent
- **CLI**: Cobra, urfave/cli

### Rust
- **Web**: Tokio + Axum (async), Actix-web (mature)
- **Data**: Sqlx (compile-time checked), Diesel
- **Systems**: No-std, embedded friendly

---

## Testing Strategies

### Python
```python
# pytest + asyncio
@pytest.mark.asyncio
async def test_create_user():
    result = await create_user_async("test@example.com")
    assert result.email == "test@example.com"
```

### TypeScript
```typescript
// Vitest
import { describe, it, expect } from "vitest";

describe("User API", () => {
  it("creates user with valid email", async () => {
    const user = await createUser({ email: "test@example.com" });
    expect(user.email).toBe("test@example.com");
  });
});
```

### Go
```go
// testify
func TestCreateUser(t *testing.T) {
    user := createUser("test@example.com")
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Rust
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_create_user() {
        let user = create_user("test@example.com").await;
        assert_eq!(user.email, "test@example.com");
    }
}
```

---

## Performance Optimization by Language

### Python
- Use async/await for I/O
- Leverage type hints for faster execution
- Batch database operations
- Use connection pooling (asyncpg, psycopg[c])

### TypeScript
- Enable strict mode
- Use const assertions for literals
- Tree-shaking unused code
- Optimize bundle size with dynamic imports

### Go
- Goroutines for concurrency (lightweight)
- Connection pooling for databases
- Buffered channels for performance
- Profile with pprof

### Rust
- Zero-cost abstractions
- Compile-time optimizations
- Avoid unnecessary allocations
- Use release builds for performance

---

## Security Best Practices

### All Languages
- ✅ Input validation (validate at boundaries)
- ✅ SQL parameterization (prevent injection)
- ✅ HTTPS/TLS encryption
- ✅ Secure secret management (.env files)
- ✅ Regular dependency updates
- ✅ Security linting tools

### Language-Specific
- **Python**: Use `secrets` module, parameterized queries
- **TypeScript**: Input validation with Zod, Valibot
- **Go**: Validate with standard library, use `log/slog`
- **Rust**: Type system prevents memory issues

---

## Common Gotchas

| Language | Issue | Solution |
|----------|-------|----------|
| Python | GIL limits threading | Use asyncio, multiprocessing, or Rust extensions |
| TypeScript | Type erasure at runtime | Use Zod/Valibot for runtime validation |
| Go | Nil pointer panics | Check for nil, use error returns |
| Rust | Borrow checker learning curve | Start with lifetimes basics, use `&` and `&mut` correctly |
| Java | Verbose boilerplate | Use Lombok, records (Java 16+) |
| PHP | Type juggling surprises | Enable strict types with `declare(strict_types=1)` |

---

## Learning Path by Language

### Beginner → Expert
1. **Syntax & basics** (1-2 weeks)
2. **Core patterns** (2-3 weeks)
3. **Ecosystem** (3-4 weeks)
4. **Performance tuning** (2-3 weeks)
5. **Production patterns** (ongoing)

### Recommended Resources
- **Official documentation** (most current)
- **Language community** (forums, discussions)
- **Context7 MCP** (latest APIs and patterns)
- **Enterprise examples** (production patterns)

---

## Integration with Development Workflow

```
Requirements (EARS)
    ↓
Language Selection (moai-foundation-langs)
    ↓
Framework Choice (best practices for language)
    ↓
Implementation (idiomatic code)
    ↓
Testing (language-specific strategies)
    ↓
Deployment (language-specific considerations)
```

---

## Related Skills

- **moai-lang-python**: Python 3.13+ deep-dive with FastAPI, async patterns
- **moai-lang-typescript**: TypeScript 5.9+ with Next.js 16, React 19, tRPC
- **moai-lang-go**: Go 1.23+ systems programming with Fiber v3
- **moai-domain-backend**: Backend architecture using multiple languages
- **moai-domain-frontend**: Frontend implementation with TypeScript/JavaScript
- **moai-essentials-debug**: Language-specific debugging techniques

---

## Skill Maturity Levels

### Level 1: Quick Reference (30 seconds)
Language overview, version info, recommended libraries

### Level 2: Core Concepts (5 minutes)
Modern patterns, best practices, common frameworks

### Level 3: Advanced Implementation (15+ minutes)
Deep-dive patterns, performance optimization, security, ecosystem mastery

---

## Summary

**moai-foundation-langs guides modern polyglot development.** Every language should support:

- ✅ **Current Version**: Use 2025 standards, not legacy versions
- ✅ **Best Practices**: Follow language idioms and conventions
- ✅ **Production Ready**: Enterprise-grade patterns and quality
- ✅ **Maintainable**: Clear, documented, team-friendly code
- ✅ **Performant**: Optimized for each language's strengths

Use **moai-foundation-langs** for **production-grade code across 25+ languages** in enterprise environments.

---

**Status**: ✅ Complete | **Coverage**: 85%+ | **Last Updated**: 2025-11-24
