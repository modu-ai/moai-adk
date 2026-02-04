---
name: moai-lang-scala
description: >
  Scala 3.4+ development specialist covering Akka, Cats Effect, ZIO, and
  Spark patterns. Use when building distributed systems, big data pipelines,
  or functional programming applications.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "language, scala, akka, cats-effect, zio, spark, sbt"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Scala", "Akka", "Cats Effect", "ZIO", "Spark", ".scala", ".sc", "build.sbt", "sbt"]
  languages: ["scala"]
---

# Scala 3.4+ Development Specialist

Functional programming, effect systems, and big data processing for JVM applications.

## Quick Reference

Auto-Triggers: Scala files (.scala, .sc), build files (build.sbt, project/build.properties)

Core Capabilities:

- Scala 3.4: Given/using, extension methods, enums, opaque types, match types
- Akka 2.9: Typed actors, streams, clustering, persistence
- Cats Effect 3.5: Pure FP runtime, fibers, concurrent structures
- ZIO 2.1: Effect system, layers, streaming, error handling
- Apache Spark 3.5: DataFrame API, SQL, structured streaming

Key Ecosystem Libraries:

- HTTP: Http4s 0.24, Tapir 1.10
- JSON: Circe 0.15, ZIO JSON 0.6
- Database: Doobie 1.0, Slick 3.5, Quill 4.8
- Streaming: FS2 3.10, ZIO Streams 2.1
- Testing: ScalaTest, Specs2, MUnit, Weaver

---

## Module Index

This skill uses progressive disclosure with specialized modules:

### Core Language

- [functional-programming.md](modules/functional-programming.md) - Scala 3.4 features: Given/Using, Type Classes, Enums, Opaque Types, Extension Methods

### Effect Systems

- [cats-effect.md](modules/cats-effect.md) - Cats Effect 3.5: IO monad, Resources, Fibers, FS2 Streaming
- [zio-patterns.md](modules/zio-patterns.md) - ZIO 2.1: Effects, Layers, ZIO Streams, Error handling

### Frameworks

- [akka-actors.md](modules/akka-actors.md) - Akka Typed Actors 2.9: Actors, Streams, Clustering patterns
- [spark-data.md](modules/spark-data.md) - Apache Spark 3.5: DataFrame API, SQL, Structured Streaming

---

## Implementation Guide

### Project Setup (SBT 1.10)

In build.sbt, set ThisBuild / scalaVersion to "3.4.2" and organization. Define lazy val root project with settings including name and libraryDependencies. Add dependencies for cats-effect, zio, akka-actor-typed, http4s-ember-server, circe-generic, and scalatest for test scope. Include scalacOptions for deprecation, feature warnings, and Xfatal-warnings.

### Quick Examples

Extension Methods: Use extension keyword with parameter in parentheses. Define methods like words splitting on whitespace and truncate checking length before taking characters and appending ellipsis.

Given and Using: Define trait with abstract method signature. Create given instance with with keyword and implement the method. Create functions with using parameter clause for implicit resolution.

Enum Types: Define enum with generic type parameters and plus variance annotations. Create case entries with parameters. Define methods on enum using match expression to handle each case, returning appropriate results.

---

## Context7 Integration

Library mappings for latest documentation:

Core Scala:

- /scala/scala3 - Scala 3.4 language reference
- /scala/scala-library - Standard library

Effect Systems:

- /typelevel/cats-effect - Cats Effect 3.5 documentation
- /typelevel/cats - Cats 2.10 functional abstractions
- /zio/zio - ZIO 2.1 documentation
- /zio/zio-streams - ZIO Streams 2.1

Akka Ecosystem:

- /akka/akka - Akka 2.9 typed actors and streams
- /akka/akka-http - Akka HTTP REST APIs
- /akka/alpakka - Akka connectors

HTTP and Web:

- /http4s/http4s - Functional HTTP server/client
- /softwaremill/tapir - API-first design

Big Data:

- /apache/spark - Spark 3.5 DataFrame and SQL
- /apache/flink - Flink 1.19 streaming
- /apache/kafka - Kafka clients 3.7

---

## Testing Quick Reference

ScalaTest: Extend AnyFlatSpec with Matchers. Use string description with should in for behavior. Make assertions with shouldBe for equality checks.

MUnit with Cats Effect: Extend CatsEffectSuite. Define test with string name. Return IO containing assertEquals assertions.

ZIO Test: Extend ZIOSpecDefault. Define spec as suite with test entries. Use for-comprehension to run effects and yield assertTrue assertions.

---

## Troubleshooting

Common Issues:

- Implicit resolution: Use scalac -explain for detailed error messages
- Type inference: Add explicit type annotations when inference fails
- SBT slow compilation: Enable Global / concurrentRestrictions in build.sbt

Effect System Issues:

- Cats Effect: Check for missing import cats.effect._ or import cats.syntax.all._
- ZIO: Verify layer composition with ZIO.serviceWith and ZIO.serviceWithZIO
- Akka: Review actor hierarchy and supervision strategies

---

## Works Well With

- moai-lang-java - JVM interoperability, Spring Boot integration
- moai-domain-backend - REST API, GraphQL, microservices patterns
- moai-domain-database - Doobie, Slick, database patterns
- moai-workflow-testing - ScalaTest, MUnit, property-based testing

---

## Additional Resources

For comprehensive reference materials:

- [reference.md](reference.md) - Complete Scala 3.4 coverage, Context7 mappings, performance
- [examples.md](examples.md) - Production-ready code: Http4s, Akka, Spark patterns

---

Last Updated: 2026-01-11
Status: Production Ready (v2.1.0)
