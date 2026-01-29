# Scala 3.4+ Reference Guide

## Complete Language Coverage

### Scala 3.4 (November 2025)

Version Information:
- Latest: 3.4.2
- Dotty: New compiler with improved type system
- TASTy: Portable intermediate representation
- JVM Target: 11, 17, 21 (recommended: 21)

Core Features:

- Export Clauses: Selective member export from composed objects
- Extension Methods: Type-safe extensions without implicit classes
- Enum Types: Algebraic data types with exhaustive pattern matching
- Opaque Types: Zero-cost type abstractions
- Union Types: A or B type unions for flexible APIs
- Intersection Types: A and B type combinations
- Match Types: Type-level computation and pattern matching
- Inline Methods: Compile-time evaluation and metaprogramming
- Given/Using: Context parameters replacing implicits
- Braceless Syntax: Optional significant indentation

---

## Context7 Library Mappings

### Core Scala

```
/scala/scala3 - Scala 3.4 language reference
/scala/scala-library - Standard library
```

### Effect Systems

```
/typelevel/cats-effect - Cats Effect 3.5 (Pure FP runtime)
/typelevel/cats - Cats 2.10 (Functional abstractions)
/zio/zio - ZIO 2.1 (Effect system)
/zio/zio-streams - ZIO Streams (Streaming)
```

### Akka Ecosystem

```
/akka/akka - Akka 2.9 (Typed actors, streams)
/akka/akka-http - Akka HTTP (REST APIs)
/akka/alpakka - Akka Alpakka (Connectors)
```

### HTTP and Web

```
/http4s/http4s - Http4s 0.24 (Functional HTTP)
/softwaremill/tapir - Tapir 1.10 (API-first design)
```

### JSON

```
/circe/circe - Circe 0.15 (JSON parsing)
/zio/zio-json - ZIO JSON 0.6 (Fast JSON)
```

### Database

```
/tpolecat/doobie - Doobie 1.0 (Functional JDBC)
/slick/slick - Slick 3.5 (FRM)
/getquill/quill - Quill 4.8 (Compile-time SQL)
```

### Big Data

```
/apache/spark - Apache Spark 3.5
/apache/flink - Apache Flink 1.19
/apache/kafka - Kafka Clients 3.7
```

### Testing

```
/scalatest/scalatest - ScalaTest 3.2
/typelevel/munit-cats-effect - MUnit Cats Effect 2.0
/zio/zio-test - ZIO Test 2.1
```

---

## Testing Patterns

### ScalaTest with Akka TestKit

```scala
class UserActorSpec extends ScalaTestWithActorTestKit with AnyWordSpecLike with Matchers:
  import UserActor.*

  val mockRepository: UserRepository = mock[UserRepository]

  "UserActor" should {
    "return user when found" in {
      val testUser = User(1L, "John", "john@example.com")
      when(mockRepository.findById(1L)).thenReturn(Some(testUser))

      val actor = spawn(UserActor(mockRepository))
      val probe = createTestProbe[Option[User]]()

      actor ! GetUser(1L, probe.ref)

      probe.expectMessage(Some(testUser))
      verify(mockRepository).findById(1L)
    }

    "return None when user not found" in {
      when(mockRepository.findById(999L)).thenReturn(None)

      val actor = spawn(UserActor(mockRepository))
      val probe = createTestProbe[Option[User]]()

      actor ! GetUser(999L, probe.ref)

      probe.expectMessage(None)
    }

    "handle multiple requests concurrently" in {
      val users = (1 to 100).map(i => User(i.toLong, s"User$i", s"user$i@example.com"))
      users.foreach(u => when(mockRepository.findById(u.id)).thenReturn(Some(u)))

      val actor = spawn(UserActor(mockRepository))
      val probes = users.map(_ => createTestProbe[Option[User]]())

      users.zip(probes).foreach { case (user, probe) =>
        actor ! GetUser(user.id, probe.ref)
      }

      users.zip(probes).foreach { case (user, probe) =>
        probe.expectMessage(Some(user))
      }
    }
  }
```

### Cats Effect Testing (MUnit)

```scala
class UserServiceSpec extends CatsEffectSuite:
  val mockRepository = mock[UserRepository[IO]]

  test("should fetch user successfully") {
    val testUser = User(1L, "John", "john@example.com")
    when(mockRepository.findById(1L)).thenReturn(IO.pure(Some(testUser)))

    val service = UserService(mockRepository)

    service.findById(1L).map { result =>
      assertEquals(result, Some(testUser))
    }
  }

  test("should handle concurrent operations") {
    val users = (1 to 10).map(i => User(i.toLong, s"User$i", s"user$i@example.com")).toList
    users.foreach(u => when(mockRepository.findById(u.id)).thenReturn(IO.pure(Some(u))))

    val service = UserService(mockRepository)

    val results = users.parTraverse(u => service.findById(u.id))

    results.map { list =>
      assertEquals(list.flatten.size, 10)
    }
  }

  test("should timeout slow operations") {
    when(mockRepository.findById(any[Long])).thenReturn(IO.sleep(5.seconds) *> IO.none)

    val service = UserService(mockRepository)

    service.findById(1L)
      .timeout(100.millis)
      .attempt
      .map { result =>
        assert(result.isLeft)
        assert(result.left.exists(_.isInstanceOf[TimeoutException]))
      }
  }
```

### ZIO Testing

```scala
object UserServiceSpec extends ZIOSpecDefault:
  val testUser = User(1L, "John", "john@example.com")

  val mockRepositoryLayer: ULayer[UserRepository] = ZLayer.succeed {
    new UserRepository:
      def findById(id: Long): UIO[Option[User]] =
        if id == 1L then ZIO.some(testUser) else ZIO.none
      def save(user: User): UIO[User] = ZIO.succeed(user)
  }

  def spec = suite("UserService")(
    test("should find existing user") {
      for
        service <- ZIO.service[UserService]
        result <- service.findById(1L)
      yield assertTrue(result == Some(testUser))
    }.provide(mockRepositoryLayer, UserService.layer),

    test("should return None for non-existent user") {
      for
        service <- ZIO.service[UserService]
        result <- service.findById(999L)
      yield assertTrue(result.isEmpty)
    }.provide(mockRepositoryLayer, UserService.layer),

    test("should handle parallel requests") {
      for
        service <- ZIO.service[UserService]
        results <- ZIO.foreachPar(1 to 100)(id => service.findById(id.toLong))
      yield assertTrue(results.flatten.size == 1)
    }.provide(mockRepositoryLayer, UserService.layer)
  )
```

### Property-Based Testing (ScalaCheck)

```scala
class UserValidationSpec extends AnyFlatSpec with Matchers with ScalaCheckPropertyChecks:
  "Email validation" should "accept valid emails" in {
    forAll(Gen.alphaNumStr, Gen.alphaNumStr) { (local, domain) =>
      whenever(local.nonEmpty && domain.nonEmpty) {
        val email = s"$local@$domain.com"
        Email(email) shouldBe a[Right[_, _]]
      }
    }
  }

  "UserId" should "roundtrip through string conversion" in {
    forAll(Gen.posNum[Long]) { id =>
      UserId.fromString(UserId(id).asString) shouldBe Some(UserId(id))
    }
  }
```

---

## Performance Characteristics

### JVM Startup and Memory

- Cold Start: 3-6s (JVM warmup)
- Warm Start: Less than 100ms
- Base Memory: 512MB or more
- GraalVM Native: Not recommended for Scala 3

### Compilation Times

- Clean Build: 60-120s
- Incremental: 10-30s
- With Cache: 30-60s
- Note: Scala 3 compiler is faster than Scala 2

### Framework Throughput

- Http4s (Blaze): 160K requests per second, P99 latency 1.5ms
- Http4s (Ember): 140K requests per second, P99 latency 2ms
- Akka HTTP: 180K requests per second, P99 latency 1.2ms
- ZIO HTTP: 170K requests per second, P99 latency 1.3ms

---

## Development Environment

### IDE Support

- IntelliJ IDEA: Good (Scala plugin required)
- VS Code: Good (Metals extension)
- Neovim: Good (Metals LSP)

### Recommended Plugins

IntelliJ IDEA:
- Scala (by JetBrains)
- ZIO for IntelliJ
- Cats Support

VS Code:
- Scala (Metals)
- Scala Syntax (official)

### Linters and Formatters

- Scalafmt: Code formatter (.scalafmt.conf)
- Scalafix: Linting and refactoring
- WartRemover: Code quality checks

Example .scalafmt.conf:
```hocon
version = 3.7.17
runner.dialect = scala3
maxColumn = 100
indent.main = 2
indent.callSite = 2
align.preset = more
rewrite.rules = [SortImports, RedundantBraces, PreferCurlyFors]
```

---

## Container Optimization

### Docker Multi-Stage Build

```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app
COPY . .
RUN sbt assembly

FROM eclipse-temurin:21-jre-alpine
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app
WORKDIR /app
COPY --from=builder /app/target/scala-3.4.2/*.jar app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### JVM Tuning for Containers

```yaml
containers:
  - name: app
    image: myapp:latest
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
    env:
      - name: JAVA_OPTS
        value: >-
          -XX:+UseContainerSupport
          -XX:MaxRAMPercentage=75.0
          -XX:+UseG1GC
          -XX:+UseStringDeduplication
```

---

## Migration Guide: Scala 2.13 to 3.4

Key Changes:

1. Braceless Syntax: Optional significant indentation
2. Given/Using: Replace `implicit` with `given` and `using`
3. Extension Methods: Replace implicit classes with `extension`
4. Enums: Replace sealed traits with `enum`
5. Export Clauses: Replace trait mixing with exports
6. Opaque Types: Replace value classes with opaque types
7. Union Types: Replace Either with union types where appropriate
8. Match Types: Replace type-level programming patterns

Example Migration:

Scala 2.13:
```scala
implicit class StringOps(s: String) {
  def words: List[String] = s.split("\\s+").toList
}

implicit def jsonEncoder: JsonEncoder[String] = ???
```

Scala 3.4:
```scala
extension (s: String)
  def words: List[String] = s.split("\\s+").toList

given JsonEncoder[String] = ???
```

---

## Effect System Comparison

### Cats Effect vs ZIO

Cats Effect:
- Pure FP approach, minimal runtime
- Better interop with Typelevel ecosystem
- Smaller learning curve from cats-core
- Resource safety via Resource type

ZIO:
- Rich built-in functionality (layers, config, logging)
- Better error handling with typed errors
- Comprehensive testing utilities
- Larger standard library

### When to Use Which

Use Cats Effect When:
- Already using Typelevel libraries (http4s, doobie, fs2)
- Prefer minimal runtime overhead
- Team familiar with tagless final pattern

Use ZIO When:
- Building complex applications with many dependencies
- Need comprehensive error handling
- Prefer opinionated framework with batteries included
- Building applications from scratch

---

Last Updated: 2025-12-07
Version: 1.0.0
