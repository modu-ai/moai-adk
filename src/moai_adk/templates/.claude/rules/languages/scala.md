---
paths:
  - "**/*.scala"
  - "**/*.sc"
  - "**/build.sbt"
---

# Scala Rules

Version: Scala 3.4+

## Tooling

- Build: sbt or Mill
- Testing: ScalaTest or munit
- Formatting: scalafmt
- Linting: scalafix, WartRemover

## Best Practices (2026)

- Prefer immutability and pure functions
- Use case classes for data modeling
- Use ZIO 2.x or Cats Effect 3.x for effect systems
- Use Iron for compile-time type refinement
- Use Tapir for type-safe HTTP APIs

## ZIO Patterns

```scala
// ZIO service pattern
trait UserService:
  def getUser(id: UserId): Task[User]

object UserService:
  def getUser(id: UserId): ZIO[UserService, Throwable, User] =
    ZIO.serviceWithZIO[UserService](_.getUser(id))

// ZLayer for dependency injection
val live: ZLayer[Database, Nothing, UserService] =
  ZLayer.fromFunction(UserServiceLive(_))
```

## Iron Type Refinement

```scala
import io.github.iltotore.iron.*
import io.github.iltotore.iron.constraint.numeric.*

// Compile-time validated types
type Age = Int :| Positive
type Email = String :| Match["^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"]

def createUser(name: String, age: Age): User = ???
```

## Spark Patterns

- Use Dataset over RDD for type safety
- Use Catalyst optimizer hints
- Partition data appropriately
- Cache intermediate results

## MoAI Integration

- Use Skill("moai-lang-scala") for detailed patterns
- Follow TRUST 5 quality gates
