---
name: "moai-lang-scala"
description: "Scala 3.4+ development specialist covering Akka, Cats Effect, ZIO, and Spark patterns. Use when building distributed systems, big data pipelines, or functional programming applications."
version: 2.0.0
category: "language"
modularized: true
user-invocable: false
updated: 2026-01-08
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
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

```scala
ThisBuild / scalaVersion := "3.4.2"
ThisBuild / organization := "com.example"

lazy val root = (project in file("."))
  .settings(
    name := "scala-service",
    libraryDependencies ++= Seq(
      "org.typelevel" %% "cats-effect" % "3.5.4",
      "dev.zio" %% "zio" % "2.1.0",
      "com.typesafe.akka" %% "akka-actor-typed" % "2.9.0",
      "org.http4s" %% "http4s-ember-server" % "0.24.0",
      "io.circe" %% "circe-generic" % "0.15.0",
      "org.scalatest" %% "scalatest" % "3.2.18" % Test
    ),
    scalacOptions ++= Seq("-deprecation", "-feature", "-Xfatal-warnings")
  )
```

### Quick Examples

Extension Methods:
```scala
extension (s: String)
  def words: List[String] = s.split("\\s+").toList
  def truncate(maxLen: Int): String =
    if s.length <= maxLen then s else s.take(maxLen - 3) + "..."
```

Given and Using:
```scala
trait JsonEncoder[A]:
  def encode(value: A): String

given JsonEncoder[String] with
  def encode(value: String): String = s"\"$value\""

def toJson[A](value: A)(using encoder: JsonEncoder[A]): String =
  encoder.encode(value)
```

Enum Types:
```scala
enum Result[+E, +A]:
  case Success(value: A)
  case Failure(error: E)

  def map[B](f: A => B): Result[E, B] = this match
    case Success(a) => Success(f(a))
    case Failure(e) => Failure(e)
```

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

ScalaTest:
```scala
class UserServiceSpec extends AnyFlatSpec with Matchers:
  "UserService" should "create user successfully" in {
    val result = service.createUser(CreateUserRequest("John", "john@example.com"))
    result.name shouldBe "John"
  }
```

MUnit with Cats Effect:
```scala
class UserServiceSuite extends CatsEffectSuite:
  test("should fetch user") {
    UserService.findById(1L).map { result =>
      assertEquals(result.name, "John")
    }
  }
```

ZIO Test:
```scala
object UserServiceSpec extends ZIOSpecDefault:
  def spec = suite("UserService")(
    test("should find user") {
      for result <- UserService.findById(1L)
      yield assertTrue(result.name == "John")
    }
  )
```

---

## Troubleshooting

Common Issues:
- Implicit resolution: Use scalac -explain for detailed error messages
- Type inference: Add explicit type annotations when inference fails
- SBT slow compilation: Enable Global / concurrentRestrictions in build.sbt

Effect System Issues:
- Cats Effect: Check for missing import cats.effect.* or import cats.syntax.all.*
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

Last Updated: 2026-01-06
Status: Production Ready (v2.0.0)
