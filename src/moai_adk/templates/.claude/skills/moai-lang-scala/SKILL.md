---
name: moai-lang-scala
description: Scala 3.4+ development specialist covering Akka, Cats Effect, ZIO, and Spark patterns. Use when building distributed systems, big data pipelines, or functional programming applications.
version: 1.0.0
category: language
tags:
  - scala
  - akka
  - cats-effect
  - zio
  - spark
  - functional-programming
context7-libraries:
  - /akka/akka
  - /typelevel/cats-effect
  - /zio/zio
  - /apache/spark
related-skills:
  - moai-lang-java
  - moai-domain-database
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Scala 3.4+ Development Specialist - Functional programming, effect systems, and big data.

Auto-Triggers: Scala files (`.scala`, `.sc`), build files (`build.sbt`, `project/build.properties`)

Core Capabilities:
- Scala 3.4: Given/using, extension methods, enums, opaque types, match types
- Akka 2.9: Typed actors, streams, clustering, persistence
- Cats Effect 3.5: Pure FP runtime, fibers, concurrent structures
- ZIO 2.1: Effect system, layers, streaming, error handling
- Apache Spark 3.5: DataFrame API, SQL, structured streaming
- Testing: ScalaTest, Specs2, MUnit, Weaver

Key Ecosystem Libraries:
- HTTP: Http4s 0.24, Tapir 1.10
- JSON: Circe 0.15, ZIO JSON 0.6
- Database: Doobie 1.0, Slick 3.5, Quill 4.8
- Streaming: FS2 3.10, ZIO Streams 2.1

---

## Implementation Guide (5 minutes)

### Scala 3.4 Core Features

Extension Methods:
```scala
extension (s: String)
  def words: List[String] = s.split("\\s+").toList
  def truncate(maxLen: Int): String =
    if s.length <= maxLen then s else s.take(maxLen - 3) + "..."
  def isBlank: Boolean = s.trim.isEmpty

extension [A](list: List[A])
  def second: Option[A] = list.drop(1).headOption
  def penultimate: Option[A] = list.dropRight(1).lastOption
```

Given and Using (Context Parameters):
```scala
trait JsonEncoder[A]:
  def encode(value: A): String

given JsonEncoder[String] with
  def encode(value: String): String = s"\"$value\""

given JsonEncoder[Int] with
  def encode(value: Int): String = value.toString

given [A](using encoder: JsonEncoder[A]): JsonEncoder[List[A]] with
  def encode(value: List[A]): String =
    value.map(encoder.encode).mkString("[", ",", "]")

def toJson[A](value: A)(using encoder: JsonEncoder[A]): String =
  encoder.encode(value)

// Usage
val json = toJson(List(1, 2, 3)) // "[1,2,3]"
```

Enum Types and ADTs:
```scala
enum Color(val hex: String):
  case Red extends Color("#FF0000")
  case Green extends Color("#00FF00")
  case Blue extends Color("#0000FF")
  case Custom(override val hex: String) extends Color(hex)

enum Result[+E, +A]:
  case Success(value: A)
  case Failure(error: E)

  def map[B](f: A => B): Result[E, B] = this match
    case Success(a) => Success(f(a))
    case Failure(e) => Failure(e)

  def flatMap[E2 >: E, B](f: A => Result[E2, B]): Result[E2, B] = this match
    case Success(a) => f(a)
    case Failure(e) => Failure(e)
```

Opaque Types:
```scala
object UserId:
  opaque type UserId = Long
  def apply(id: Long): UserId = id
  def fromString(s: String): Option[UserId] = s.toLongOption
  extension (id: UserId)
    def value: Long = id
    def asString: String = id.toString

export UserId.UserId

object Email:
  opaque type Email = String
  def apply(email: String): Either[String, Email] =
    if email.contains("@") && email.contains(".") then Right(email)
    else Left(s"Invalid email: $email")
  extension (email: Email)
    def value: String = email
    def domain: String = email.split("@").last
```

Union and Intersection Types:
```scala
// Union types
type StringOrInt = String | Int

def describe(value: StringOrInt): String = value match
  case s: String => s"String: $s"
  case i: Int => s"Int: $i"

// Intersection types
trait HasName:
  def name: String

trait HasAge:
  def age: Int

type Person = HasName & HasAge

def greet(person: Person): String =
  s"Hello ${person.name}, age ${person.age}"
```

### Cats Effect 3.5

Basic IO Operations:
```scala
import cats.effect.*
import cats.syntax.all.*

def program: IO[Unit] =
  for
    _ <- IO.println("Enter your name:")
    name <- IO.readLine
    _ <- IO.println(s"Hello, $name!")
  yield ()

// Resource management
def withFile[A](path: String)(use: BufferedReader => IO[A]): IO[A] =
  Resource
    .make(IO(new BufferedReader(new FileReader(path))))(r => IO(r.close()))
    .use(use)

// Error handling
def fetchUser(id: Long): IO[User] =
  IO.fromOption(repository.findById(id))(UserNotFound(id))
    .handleErrorWith {
      case _: UserNotFound => IO.raiseError(new Exception(s"User $id not found"))
    }
```

Concurrent Programming:
```scala
import cats.effect.std.*

// Parallel execution
def fetchUserData(userId: Long): IO[UserData] =
  (fetchUser(userId), fetchOrders(userId), fetchPreferences(userId))
    .parMapN(UserData.apply)

// Fibers for background processing
def processInBackground(task: IO[Unit]): IO[Unit] =
  task.start.flatMap(fiber =>
    IO.println("Task started") *> fiber.join.void
  )

// Semaphore for rate limiting
def rateLimitedRequests[A](tasks: List[IO[A]], max: Int): IO[List[A]] =
  Semaphore[IO](max).flatMap { sem =>
    tasks.parTraverse(task => sem.permit.use(_ => task))
  }

// Ref for shared state
def counter: IO[Ref[IO, Int]] = Ref.of[IO, Int](0)
def increment(ref: Ref[IO, Int]): IO[Int] = ref.updateAndGet(_ + 1)
```

Streaming with FS2:
```scala
import fs2.*
import fs2.io.file.*

def processLargeFile(path: Path): Stream[IO, String] =
  Files[IO].readUtf8Lines(path)
    .filter(_.nonEmpty)
    .map(_.toLowerCase)
    .evalTap(line => IO.println(s"Processing: $line"))

def writeResults(path: Path, lines: Stream[IO, String]): IO[Unit] =
  lines.intersperse("\n")
    .through(text.utf8.encode)
    .through(Files[IO].writeAll(path))
    .compile.drain
```

### ZIO 2.1

Basic ZIO Operations:
```scala
import zio.*

val program: ZIO[Any, Nothing, Unit] =
  for
    _ <- Console.printLine("Enter your name:")
    name <- Console.readLine
    _ <- Console.printLine(s"Hello, $name!")
  yield ()

def fetchUser(id: Long): ZIO[UserRepository, UserError, User] =
  for
    repo <- ZIO.service[UserRepository]
    user <- ZIO.fromOption(repo.findById(id)).orElseFail(UserNotFound(id))
  yield user

// Resource management
def withFile[A](path: String): ZIO[Scope, IOException, BufferedReader] =
  ZIO.acquireRelease(
    ZIO.attempt(new BufferedReader(new FileReader(path))).refineToOrDie[IOException]
  )(reader => ZIO.succeed(reader.close()))
```

ZIO Layers (Dependency Injection):
```scala
trait UserRepository:
  def findById(id: Long): Task[Option[User]]
  def save(user: User): Task[User]

trait EmailService:
  def sendEmail(to: String, subject: String, body: String): Task[Unit]

case class UserRepositoryLive(db: Database) extends UserRepository:
  def findById(id: Long): Task[Option[User]] =
    ZIO.attempt(db.query(s"SELECT * FROM users WHERE id = $id")).map(_.headOption)
  def save(user: User): Task[User] =
    ZIO.attempt(db.insert("users", user)).as(user)

object UserRepositoryLive:
  val layer: ZLayer[Database, Nothing, UserRepository] =
    ZLayer.fromFunction(UserRepositoryLive.apply)

// Composing layers
val appLayer = Database.layer >>> UserRepositoryLive.layer ++ EmailServiceLive.layer

object Main extends ZIOAppDefault:
  def run = program.provide(appLayer)
```

ZIO Streaming:
```scala
import zio.stream.*

def processEvents: ZStream[Any, Throwable, ProcessedEvent] =
  ZStream.fromQueue(eventQueue)
    .filter(_.isValid)
    .mapZIO(enrichEvent)
    .grouped(100)
    .mapZIO(batchProcess)
    .flattenIterables
```

### Akka Typed Actors

Actor Definition:
```scala
import akka.actor.typed.*
import akka.actor.typed.scaladsl.*

object UserActor:
  sealed trait Command
  case class GetUser(id: Long, replyTo: ActorRef[Option[User]]) extends Command
  case class CreateUser(request: CreateUserRequest, replyTo: ActorRef[User]) extends Command
  case class UpdateUser(id: Long, name: String, replyTo: ActorRef[Option[User]]) extends Command

  def apply(repository: UserRepository): Behavior[Command] =
    Behaviors.receiveMessage {
      case GetUser(id, replyTo) =>
        replyTo ! repository.findById(id)
        Behaviors.same
      case CreateUser(request, replyTo) =>
        replyTo ! repository.save(User.from(request))
        Behaviors.same
      case UpdateUser(id, name, replyTo) =>
        val updated = repository.findById(id).map(u => repository.save(u.copy(name = name)))
        replyTo ! updated
        Behaviors.same
    }
```

Akka Streams:
```scala
import akka.stream.*
import akka.stream.scaladsl.*

val source: Source[Int, NotUsed] = Source(1 to 1000)
val flow: Flow[Int, String, NotUsed] =
  Flow[Int].filter(_ % 2 == 0).map(_ * 2).map(_.toString)
val sink: Sink[String, Future[Done]] = Sink.foreach(println)

val graph = source.via(flow).toMat(sink)(Keep.right)

// Backpressure handling
val throttledSource = source
  .throttle(100, 1.second)
  .buffer(1000, OverflowStrategy.backpressure)
```

### Apache Spark 3.5

DataFrame Operations:
```scala
import org.apache.spark.sql.{DataFrame, SparkSession}
import org.apache.spark.sql.functions.*

val spark = SparkSession.builder()
  .appName("Data Analysis")
  .config("spark.sql.adaptive.enabled", "true")
  .getOrCreate()

import spark.implicits.*

val userMetrics = orders
  .groupBy("user_id")
  .agg(
    sum("amount").as("total_spent"),
    count("*").as("order_count"),
    avg("amount").as("avg_order_value")
  )
  .join(users, Seq("user_id"), "left")
  .withColumn("customer_tier",
    when(col("total_spent") > 10000, "platinum")
      .when(col("total_spent") > 1000, "gold")
      .otherwise("standard")
  )
```

Structured Streaming:
```scala
val streamingOrders = spark.readStream
  .format("kafka")
  .option("kafka.bootstrap.servers", "localhost:9092")
  .option("subscribe", "orders")
  .load()
  .selectExpr("CAST(value AS STRING)")
  .as[String]
  .map(parseOrder)

val aggregated = streamingOrders
  .withWatermark("timestamp", "10 minutes")
  .groupBy(window($"timestamp", "1 hour"), $"product_category")
  .agg(sum("amount").as("hourly_sales"))

aggregated.writeStream
  .format("delta")
  .outputMode("append")
  .option("checkpointLocation", "/checkpoints/sales")
  .start("/output/hourly-sales")
```

---

## Advanced Patterns

### Build Configuration (SBT 1.10)

```scala
ThisBuild / scalaVersion := "3.4.2"
ThisBuild / organization := "com.example"
ThisBuild / version := "1.0.0"

lazy val root = (project in file("."))
  .settings(
    name := "scala-service",
    libraryDependencies ++= Seq(
      "org.typelevel" %% "cats-effect" % "3.5.4",
      "org.typelevel" %% "cats-core" % "2.10.0",
      "co.fs2" %% "fs2-core" % "3.10.0",
      "dev.zio" %% "zio" % "2.1.0",
      "com.typesafe.akka" %% "akka-actor-typed" % "2.9.0",
      "org.http4s" %% "http4s-ember-server" % "0.24.0",
      "io.circe" %% "circe-generic" % "0.15.0",
      "org.tpolecat" %% "doobie-core" % "1.0.0-RC4",
      "org.scalatest" %% "scalatest" % "3.2.18" % Test,
      "org.typelevel" %% "munit-cats-effect" % "2.0.0" % Test
    ),
    scalacOptions ++= Seq("-deprecation", "-feature", "-Xfatal-warnings")
  )
```

### Testing Quick Reference

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

## Context7 Integration

Library mappings for latest documentation:
- `/scala/scala3` - Scala 3.4 language reference
- `/typelevel/cats-effect` - Cats Effect 3.5 documentation
- `/zio/zio` - ZIO 2.1 documentation
- `/akka/akka` - Akka 2.9 typed actors and streams
- `/http4s/http4s` - Functional HTTP server/client
- `/apache/spark` - Spark 3.5 DataFrame and SQL
- `/circe/circe` - JSON library
- `/slick/slick` - Database access

---

## Troubleshooting

Common Issues:
- Implicit resolution: Use `scalac -explain` for detailed error messages
- Type inference: Add explicit type annotations when inference fails
- SBT slow compilation: Enable `Global / concurrentRestrictions` in build.sbt

Effect System Issues:
- Cats Effect: Check for missing `import cats.effect.*` or `import cats.syntax.all.*`
- ZIO: Verify layer composition with `ZIO.serviceWith` and `ZIO.serviceWithZIO`
- Akka: Review actor hierarchy and supervision strategies

---

## Works Well With

- `moai-lang-java` - JVM interoperability, Spring Boot integration
- `moai-domain-backend` - REST API, GraphQL, microservices patterns
- `moai-domain-database` - Doobie, Slick, database patterns
- `moai-quality-testing` - ScalaTest, MUnit, property-based testing
- `moai-infra-kubernetes` - Scala application deployment

---

## Advanced Documentation

For comprehensive reference materials:
- [reference.md](reference.md) - Complete Scala 3.4 coverage, Context7 mappings, performance
- [examples.md](examples.md) - Production-ready code: Http4s, Akka, Spark patterns

---

Last Updated: 2025-12-07
Status: Production Ready (v1.0.0)
