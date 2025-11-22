---
name: moai-lang-scala
description: Scala 3.6+ with functional programming, Play Framework, ZIO, and big data processing.
version: 1.0.0
modularized: true
last_updated: 2025-11-22
compliance_score: 70
auto_trigger_keywords:
  - lang
  - scala
category_tier: 1
---

## Quick Reference (30 seconds)

# Scala Functional Programming — Enterprise

**Primary Focus**: Scala 3.6+ with ZIO, Cats, Play Framework, and functional patterns
**Best For**: Functional programming, JVM interop, distributed systems, data processing
**Key Libraries**: Scala 3.6, ZIO 2, Cats 2, Play Framework 3, ScalaTest 3.2
**Auto-triggers**: Scala, Play, ZIO, Cats, Spark, Flink, functional programming

| Version | Release | Support |
|---------|---------|---------|
| Scala 3.6.0 | 2025-11 | Active |
| ZIO 2.1.0 | 2024-11 | Active |
| Play Framework 3.0 | 2024-10 | 2026-10 |
| ScalaTest 3.2 | 2024-06 | Active |

---

## Three-Level Learning Path

### Level 1: Fundamentals (Read examples.md)

Core Scala 3 concepts with practical examples:
- **Scala 3 Core**: Given/Using, Enums, opaque types, extension methods
- **Functional Programming**: Pattern matching, higher-order functions, monads
- **Play Framework**: Controllers, routes, HTTP handling
- **Testing**: ScalaTest assertions, property-based testing
- **Examples**: See `examples.md` for 15+ working examples

### Level 2: Advanced Patterns (See modules/)

Production-ready enterprise patterns:
- **ZIO Effect System**: Async operations, error handling, resource management
- **Cats & FP Abstractions**: Monads, Functors, Applicatives, Composable effects
- **Type System**: Higher-kinded types, dependent types, compiler reasoning
- **Pattern Matching**: Complex patterns, GADTs, exhaustiveness checking
- **References**: See `reference.md` for API and advanced FP concepts

### Level 3: Production Deployment (Consult domain skills)

Enterprise deployment:
- **Distributed Systems**: Akka, Play clustering, service discovery
- **Data Processing**: Apache Spark, Apache Flink, streaming
- **Performance**: JVM tuning, GC optimization, profiling
- **Details**: Skill("moai-domain-backend"), Skill("moai-essentials-perf")

---

## Learn More

- **Examples**: See `examples.md` for Scala 3, ZIO, Play, and FP patterns
- **Advanced Patterns**: See `modules/advanced-patterns.md` for ZIO, Cats, and functional abstractions
- **Optimization**: See `modules/optimization.md` for performance tuning
- **Scala 3 Reference**: https://docs.scala-lang.org/scala3/book/
- **ZIO Documentation**: https://zio.dev/
- **Play Framework**: https://www.playframework.com/documentation

---

## Installation Commands

```bash
# Scala setup with sbt
curl https://raw.githubusercontent.com/coursier/install/master/install.sh | bash
cs install scala
cs install sbt

# Create project
sbt new scala/scala3.g8

# sbt dependencies (in build.sbt)
val scala3Version = "3.6.0"

libraryDependencies ++= Seq(
  "dev.zio" %% "zio" % "2.1.0",
  "dev.zio" %% "zio-test" % "2.1.0" % Test,
  "org.typelevel" %% "cats-core" % "2.12.0",
  "com.typesafe.play" %% "play-json" % "2.10.0",
  "org.scalatest" %% "scalatest" % "3.2.19" % Test
)
```

---

## Best Practices

1. **Leverage type system**: Use Scala 3's modern syntax effectively
2. **Use given/using**: Replace implicits with given instances
3. **Embrace FP**: Prefer immutability and pure functions
4. **ZIO for async**: Use ZIO for concurrent and async operations
5. **Cats for abstractions**: Use Cats for composable FP abstractions
6. **Test with property-based testing**: Use ScalaCheck for exhaustive testing
7. **Handle errors explicitly**: Use ZIO/Cats error types instead of exceptions
8. **Optimize JVM**: Tune GC and JVM settings for production

---

## What It Does

Scala 3.6+ with ZIO, Cats, Play Framework, and comprehensive functional programming support.

**Key capabilities**:
- Modern Scala 3 syntax (given/using, enums, opaque types)
- Functional programming with Cats and ZIO
- Type-safe concurrency and async operations
- Web development with Play Framework
- Property-based testing with ScalaTest
- JVM interoperability and optimization

---

## Works Well With

- `moai-domain-backend` (backend architecture)
- `moai-essentials-perf` (JVM performance tuning)
- `moai-essentials-debug` (debugging support)
- `moai-foundation-trust` (quality gates)

---

## Changelog

- **v3.1.0** (2025-11-22): Modularized structure with advanced patterns and optimization modules
- **v3.0.0** (2025-11-21): Scala 3.6, ZIO 2, Cats 2 update
- **v2.0.0** (2025-10-15): Latest tool versions, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial release

---

## Advanced Topics

See `modules/advanced-patterns.md` for:
- ZIO ecosystem (ZIO Streams, ZIO HTTP, ZIO Config)
- Cats Effect and typeclasses
- GADT and advanced type patterns
- Macro programming in Scala 3
- Domain-specific languages (DSLs)

See `modules/optimization.md` for:
- JVM tuning for Scala
- GC optimization strategies
- Profiling with JFR and async-profiler
- Scala compilation performance
- Memory optimization patterns

---

## Context7 Integration

### Related Libraries & Tools
- [Scala](/lampepfl/dotty): Scala 3 compiler and language
- [ZIO](/zio/zio): Type-safe, composable asynchronous and concurrent programming
- [Cats](/typelevel/cats): Lightweight, modular functional programming library
- [Play Framework](/playframework/playframework): The High Velocity Web Framework
- [ScalaTest](/scalatest/scalatest): Testing framework for Scala

### Official Documentation
- [Scala 3 Book](https://docs.scala-lang.org/scala3/book/)
- [ZIO Documentation](https://zio.dev/)
- [Cats Documentation](https://typelevel.org/cats/)
- [Play Framework Guide](https://www.playframework.com/documentation)
- [ScalaTest Guide](http://www.scalatest.org/)

### Version-Specific Guides
Latest stable version: Scala 3.6.0, ZIO 2.1.0, Play Framework 3.0
- [Scala 3.6 Release Notes](https://github.com/scala/scala3/releases)
- [ZIO 2.0 Migration](https://zio.dev/migration-guide/)
- [Cats 2.x Guide](https://typelevel.org/cats/typeclasses.html)
- [Play 3 Migration](https://www.playframework.com/documentation/3.0.x/Migration30)