---
paths: "**/*.scala,**/*.sc,**/build.sbt"
---

# Scala Development Guide

## Scala 3.5+ Development Specialist

Functional programming, effect systems, and big data processing for JVM applications.


Core Capabilities:

- Scala 3.5: Given/using, extension methods, enums, opaque types, match types
- Akka 2.9: Typed actors, streams, clustering, persistence (note: Akka went BSL-licensed in 2022; Apache Pekko is the Apache-licensed fork and drop-in replacement for new projects)
- Cats Effect 3.5: Pure FP runtime, fibers, concurrent structures
- ZIO 2.1: Effect system, layers, streaming, error handling
- Apache Spark 4.0: DataFrame API, SQL, structured streaming

Key Ecosystem Libraries:

- HTTP: Http4s 0.24, Tapir 1.10
- JSON: Circe 0.15, ZIO JSON 0.6
- Database: Doobie 1.0, Slick 3.5, Quill 4.8
- Streaming: FS2 3.10, ZIO Streams 2.1
- Testing: ScalaTest, Specs2, MUnit, Weaver

---

## Coverage Areas

This guide is self-contained. Use the sections below as the primary reference for:

- Scala 3.5 functional programming: Given/Using, type classes, enums, opaque types, extension methods
- Cats Effect 3.5 effect system: IO monad, resources, fibers, FS2 streaming
- ZIO 2.1: effects, layers, ZIO streams, error handling
- Akka Typed Actors 2.9 (or Apache Pekko 1.x as the Apache-licensed fork): actors, streams, clustering patterns
- Apache Spark 4.0: DataFrame API, SQL, structured streaming

---


### Project Setup (SBT 1.10)

In build.sbt, set ThisBuild / scalaVersion to "3.5.0" and organization. Define lazy val root project with settings including name and libraryDependencies. Add dependencies for cats-effect, zio, akka-actor-typed (or org.apache.pekko for the Apache-licensed fork), http4s-ember-server, circe-generic, and scalatest for test scope. Include scalacOptions for deprecation, feature warnings, and Xfatal-warnings.

### Quick Examples

Extension Methods: Use extension keyword with parameter in parentheses. Define methods like words splitting on whitespace and truncate checking length before taking characters and appending ellipsis.

Given and Using: Define trait with abstract method signature. Create given instance with with keyword and implement the method. Create functions with using parameter clause for implicit resolution.

Enum Types: Define enum with generic type parameters and plus variance annotations. Create case entries with parameters. Define methods on enum using match expression to handle each case, returning appropriate results.

---

## Documentation References

Library references for latest documentation:

Core Scala:

- /scala/scala3 - Scala 3.5 language reference
- /scala/scala-library - Standard library

Effect Systems:

- /typelevel/cats-effect - Cats Effect 3.5 documentation
- /typelevel/cats - Cats 2.10 functional abstractions
- /zio/zio - ZIO 2.1 documentation
- /zio/zio-streams - ZIO Streams 2.1

Akka Ecosystem (and Apache Pekko fork):

- /akka/akka - Akka 2.9 typed actors and streams (BSL-licensed since 2022)
- /apache/incubator-pekko - Apache Pekko 1.x, the Apache-licensed drop-in fork
- /akka/akka-http - Akka HTTP REST APIs
- /akka/alpakka - Akka connectors

HTTP and Web:

- /http4s/http4s - Functional HTTP server/client
- /softwaremill/tapir - API-first design

Big Data:

- /apache/spark - Spark 4.0 DataFrame and SQL
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
- Akka (or Apache Pekko): Review actor hierarchy and supervision strategies

## Related Resources

- `.claude/rules/moai/languages/java.md` - JVM interoperability, Spring Boot integration
- moai-domain-backend - REST API, GraphQL, microservices patterns
- moai-domain-database - Doobie, Slick, database patterns
- moai-workflow-testing - ScalaTest, MUnit, property-based testing

---


For comprehensive reference materials, use the Coverage Areas and Documentation References sections above together with the testing reference; this guide is self-contained for Http4s, Akka, and Spark patterns.

---

