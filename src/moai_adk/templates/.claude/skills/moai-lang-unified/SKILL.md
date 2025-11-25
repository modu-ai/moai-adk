---
name: moai-lang-unified
description: Unified enterprise programming language skill covering 25+ languages including Python 3.13, TypeScript 5.9, Go 1.23, Rust 1.91, Java 21, JavaScript ES2025, C++, C#, PHP, Swift, Kotlin, Scala, Elixir, Ruby, R, SQL, Shell, and more with patterns, best practices, and Context7 integration
version: 1.0.0
modularized: false
tags:
  - programming-language
  - enterprise
  - development
  - unified
  - multi-language
updated: 2025-11-25
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-25  
**compliance_score**: 92%  
**auto_trigger_keywords**: lang, programming, code, development, python, typescript, go, rust, java, javascript, cpp, csharp, php, swift, kotlin, scala, elixir, ruby, r, sql, shell  

## Quick Reference (30 seconds)

**Unified Enterprise Programming Language Expert** - 25+ languages with patterns, best practices, and Context7 integration.

**Auto-Triggers**: Any language-specific code, files, or discussions (`.py`, `.ts`, `.go`, `.rs`, `.java`, `.js`, `.cpp`, `.cs`, `.php`, `.swift`, `.kt`, `.scala`, `.ex`, `.rb`, `.r`, `.sql`, `.sh`)

**Core Capabilities**:
- ✅ 25+ programming languages with latest versions
- ✅ Language-specific patterns and best practices
- ✅ Cross-language concepts and comparisons
- ✅ Context7 integration for latest documentation
- ✅ Progressive disclosure from basics to advanced
- ✅ Enterprise-ready patterns and deployment

## Language Coverage

### Scripting & Dynamic Languages
- **Python 3.13** - FastAPI, Django, async patterns, data science
- **JavaScript ES2025** - Node.js 22 LTS, Express, browser development
- **TypeScript 5.9** - React 19, Next.js 16, type safety
- **PHP 8.4** - Laravel, Symfony, composer patterns
- **Ruby 3.3** - Rails, Sinatra, metaprogramming
- **Shell/Bash** - DevOps, scripting, system automation

### Systems & Performance Languages
- **Go 1.23** - Concurrency, Fiber, systems programming
- **Rust 1.91** - Memory safety, Tokio, async systems
- **C++** - Systems programming, performance optimization
- **C** - Low-level programming, embedded systems

### Enterprise & JVM Languages
- **Java 21 LTS** - Spring Boot, enterprise patterns
- **Kotlin** - Android, server-side, coroutines
- **Scala** - Functional programming, big data

### Mobile & Platform Languages
- **Swift** - iOS development, server-side Swift
- **C# (.NET 8)** - Enterprise applications, game development
- **Dart** - Flutter, cross-platform development

### Data & Domain Languages
- **R** - Statistical computing, data analysis
- **SQL** - Database queries across PostgreSQL, MySQL, etc.
- **Elixir** - Functional programming, Phoenix, BEAM

---

## Quick Reference Examples

### Python 3.13+ FastAPI Pattern
```python
from fastapi import FastAPI, Depends
from sqlalchemy.ext.asyncio import AsyncSession

app = FastAPI(title="API", version="1.0.0")

@app.get("/users/{user_id}")
async def get_user(user_id: int, db: AsyncSession = Depends(get_db)) -> UserResponse:
    user = await get_user_by_id(db, user_id)
    return UserResponse.model_validate(user)
```

### TypeScript 5.9+ React Pattern
```typescript
// React 19 Server Component
export default async function UserProfile({ userId }: { userId: string }) {
  const user = await getUser(userId)
  return <div><h1>{user.name}</h1><p>{user.email}</p></div>
}
```

### Go 1.23+ Fiber Pattern
```go
func main() {
    app := fiber.New()
    app.Get("/users", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"users": []string{"John", "Jane"}})
    })
    app.Listen(":3000")
}
```

### Rust 1.91+ Axum Pattern
```rust
#[tokio::main]
async fn main() {
    let app = Router::new().route("/users/:id", get(get_user));
    axum::serve(listener, app).await.unwrap()
}
```

### Java 21 LTS Spring Boot Pattern
```java
@RestController
@RequestMapping("/api/users")
public class UserController {
    @GetMapping("/{id}")
    public ResponseEntity<UserDto> getUser(@PathVariable Long id) {
        return ResponseEntity.ok(userService.findById(id));
    }
}
```

---

## Cross-Language Patterns

### Async Programming Comparison

| Language | Syntax | Runtime | Best For |
|----------|--------|---------|----------|
| **Python** | `async/await` | asyncio | I/O-bound, data science |
| **TypeScript** | `async/await` | Node.js | Full-stack, real-time |
| **Go** | `goroutines` | Go runtime | Concurrent systems |
| **Rust** | `async/await` | Tokio | Performance-critical |
| **Java** | `virtual threads` | JVM | Enterprise systems |

### Error Handling Patterns

**Python**: `try/except` with exceptions
**TypeScript**: `try/catch` with async/await
**Go**: Multiple return values `(result, err)`
**Rust**: `Result/Option` types with `match` or `?`
**Java**: Checked exceptions with try/catch/finally

### Package Management

| Language | Package Manager | Lock File | Registry |
|----------|----------------|-----------|----------|
| **Python** | pip/poetry | requirements.txt/poetry.lock | PyPI |
| **TypeScript** | npm/yarn | package-lock.json/yarn.lock | npm |
| **Go** | go modules | go.sum | Go Modules |
| **Rust** | cargo | Cargo.lock | crates.io |
| **Java** | Maven/Gradle | pom.xml/build.gradle | Maven Central |

---

## When to Use Each Language

### Python 3.13
✅ **Use for**: Data science, ML/AI, web APIs, automation
❌ **Avoid**: Real-time performance <1ms, embedded systems

### TypeScript 5.9
✅ **Use for**: Enterprise applications, full-stack, type safety
❌ **Avoid**: Rapid prototyping, simple scripts

### Go 1.23
✅ **Use for**: Microservices, CLI tools, cloud-native apps
❌ **Avoid**: Complex GUI, rapid prototyping

### Rust 1.91
✅ **Use for**: Performance-critical, systems programming
❌ **Avoid**: Rapid prototyping, simple CRUD

### Java 21 LTS
✅ **Use for**: Enterprise apps, large systems, big data
❌ **Avoid**: Lightweight CLI, rapid prototyping

---

## Context7 Integration

Automatically fetches latest documentation for:

**Python**: FastAPI, Django, Pydantic, SQLAlchemy, pytest  
**TypeScript**: React, Next.js, Node.js, tRPC, Zod  
**Go**: Gin, Echo, GORM  
**Rust**: Tokio, Axum, Serde  
**Java**: Spring Boot, Hibernate  

Usage example:
```python
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/tiangolo/fastapi",
    topic="async dependency-injection",
    page=1
)
```

---

## Works Well With

- `moai-domain-backend` — REST API, GraphQL, microservices
- `moai-domain-frontend` — React, Vue, Angular, UI components
- `moai-domain-database` — SQL, NoSQL, ORM patterns
- `moai-foundation-trust` — TRUST 5 quality principles
- `moai-essentials-debug` — AI-powered debugging
- `moai-context7-integration` — Latest documentation access

---

## Troubleshooting

**Python**: Check venv, `pip list`, `python -c "import sys"`  
**TypeScript**: `npx tsc --noEmit`, `npm ls typescript`  
**Go**: `go mod tidy`, `go mod verify`  
**Rust**: `rustc --version`, `cargo check`, `cargo tree`  
**Java**: `java -version`, `mvn/gradle build`

---

## Advanced Documentation

For comprehensive reference materials:

- **[reference.md](reference.md)** - Complete language coverage, Context7 library mappings, performance characteristics
- **[examples.md](examples.md)** - Multi-language code examples, REST API implementations, testing patterns, deployment configurations

---

## Conclusion

This unified language skill replaces all individual moai-lang* skills while maintaining their expertise and functionality. It automatically detects programming context and provides relevant expertise for 25+ languages with Context7 integration and cross-language capabilities.

**Last Updated**: 2025-11-25  
**Status**: Production Ready (Enterprise v1.0.0)  
**Replaces**: All moai-lang-* individual skills
