# ZIO 2.1 Patterns

Comprehensive coverage of ZIO 2.1 including Effects, Layers, ZIO Streams, and Error handling.

## Context7 Library Mappings

```
/zio/zio - ZIO 2.1 documentation
/zio/zio-streams - ZIO Streams 2.1
/zio/zio-http - ZIO HTTP
/zio/zio-json - ZIO JSON 0.6
/zio/zio-config - ZIO Config
/zio/zio-logging - ZIO Logging
```

---

## ZIO Effect Basics

### ZIO Type Signature

```scala
// ZIO[R, E, A]
// R = Environment (dependencies)
// E = Error type
// A = Success type

import zio.*

// Common type aliases
type Task[A] = ZIO[Any, Throwable, A]      // No deps, any error
type UIO[A] = ZIO[Any, Nothing, A]          // No deps, no error
type RIO[R, A] = ZIO[R, Throwable, A]       // With deps, any error
type IO[E, A] = ZIO[Any, E, A]              // No deps, typed error
type URIO[R, A] = ZIO[R, Nothing, A]        // With deps, no error
```

### Creating Effects

```scala
// Pure values
val succeed: UIO[Int] = ZIO.succeed(42)
val fail: IO[String, Nothing] = ZIO.fail("error")

// From effects
val attempt: Task[Int] = ZIO.attempt(dangerousOperation())
val fromOption: IO[None.type, Int] = ZIO.fromOption(maybeInt)
val fromEither: IO[String, Int] = ZIO.fromEither(eitherValue)
val fromTry: Task[Int] = ZIO.fromTry(tryValue)

// Async operations
val async: Task[Int] = ZIO.async { callback =>
  someAsyncApi { result =>
    callback(ZIO.succeed(result))
  }
}

// Console operations
val readLine: Task[String] = Console.readLine
val printLine: UIO[Unit] = Console.printLine("Hello").orDie
```

### Composing Effects

```scala
// For comprehension
val program: ZIO[Any, Nothing, Unit] =
  for
    _ <- Console.printLine("Enter your name:").orDie
    name <- Console.readLine.orDie
    _ <- Console.printLine(s"Hello, $name!").orDie
  yield ()

// Combinators
val combined: UIO[(Int, String)] = (zio1 <*> zio2)
val sequenced: UIO[String] = zio1 *> zio2
val mapped: UIO[String] = zio1.map(_.toString)
val flatMapped: UIO[String] = zio1.flatMap(n => ZIO.succeed(n.toString))

// Parallel composition
val parallel: UIO[(Int, String)] = zio1 <&> zio2
val racing: UIO[Int] = zio1.race(zio2)
```

---

## Error Handling

### Typed Errors

```scala
// Define domain errors
enum AppError:
  case NotFound(id: Long)
  case ValidationError(message: String)
  case DatabaseError(cause: Throwable)

// Effect with typed error
def findUser(id: Long): ZIO[UserRepository, AppError, User] =
  for
    repo <- ZIO.service[UserRepository]
    user <- ZIO.fromOption(repo.findById(id))
      .orElseFail(AppError.NotFound(id))
  yield user

// Error recovery
val recovered: UIO[User] = findUser(1L)
  .catchAll {
    case AppError.NotFound(_) => ZIO.succeed(User.default)
    case AppError.ValidationError(msg) => ZIO.die(new Exception(msg))
    case AppError.DatabaseError(cause) => ZIO.die(cause)
  }

// Partial recovery
val partialRecover: ZIO[UserRepository, AppError.DatabaseError, User] =
  findUser(1L).catchSome {
    case AppError.NotFound(_) => ZIO.succeed(User.default)
    case AppError.ValidationError(_) => ZIO.succeed(User.default)
  }
```

### Error Transformation

```scala
// Map errors
val mappedError: ZIO[UserRepository, String, User] =
  findUser(1L).mapError(_.toString)

// Refine errors
val refined: ZIO[UserRepository, AppError.NotFound, User] =
  findUser(1L).refineToOrDie[AppError.NotFound]

// Either conversion
val asEither: URIO[UserRepository, Either[AppError, User]] =
  findUser(1L).either

// Absorb errors
val absorbed: RIO[UserRepository, User] =
  findUser(1L).absorb(e => new Exception(e.toString))
```

### Ensuring and Finalizing

```scala
// Ensure cleanup runs
val withCleanup: Task[Unit] =
  acquire
    .flatMap(use)
    .ensuring(release)

// On success/failure
val withHooks: Task[Result] =
  operation
    .onExit {
      case Exit.Success(value) => logSuccess(value)
      case Exit.Failure(cause) => logFailure(cause)
    }

// Bracket pattern
val bracketed: Task[Result] =
  ZIO.acquireReleaseWith(acquire)(release)(use)
```

---

## ZIO Layers (Dependency Injection)

### Defining Services

```scala
// Service trait
trait UserRepository:
  def findById(id: Long): Task[Option[User]]
  def save(user: User): Task[User]
  def delete(id: Long): Task[Boolean]

// Accessor object
object UserRepository:
  def findById(id: Long): ZIO[UserRepository, Throwable, Option[User]] =
    ZIO.serviceWithZIO(_.findById(id))
  
  def save(user: User): ZIO[UserRepository, Throwable, User] =
    ZIO.serviceWithZIO(_.save(user))

// Implementation
case class UserRepositoryLive(db: Database) extends UserRepository:
  def findById(id: Long): Task[Option[User]] =
    ZIO.attempt(db.query(s"SELECT * FROM users WHERE id = $id")).map(_.headOption)
  
  def save(user: User): Task[User] =
    ZIO.attempt(db.insert("users", user)).as(user)
  
  def delete(id: Long): Task[Boolean] =
    ZIO.attempt(db.delete("users", id)).map(_ > 0)

// Layer definition
object UserRepositoryLive:
  val layer: ZLayer[Database, Nothing, UserRepository] =
    ZLayer.fromFunction(UserRepositoryLive.apply)
```

### Composing Layers

```scala
// Horizontal composition (both)
val combinedLayer: ZLayer[Database, Nothing, UserRepository & EmailService] =
  UserRepositoryLive.layer ++ EmailServiceLive.layer

// Vertical composition (dependency)
val fullLayer: ZLayer[Any, Nothing, UserRepository] =
  Database.layer >>> UserRepositoryLive.layer

// Complex layer graph
val appLayer: ZLayer[Any, Throwable, AppEnv] =
  ZLayer.make[AppEnv](
    Database.layer,
    UserRepositoryLive.layer,
    EmailServiceLive.layer,
    UserService.layer,
    Config.layer
  )
```

### Using Layers

```scala
// Provide layers to effects
object Main extends ZIOAppDefault:
  def run: ZIO[Any, Throwable, Unit] =
    program.provide(
      Database.layer,
      UserRepositoryLive.layer,
      EmailServiceLive.layer,
      UserService.layer
    )

// Provide specific dependencies
val provided: Task[User] = findUser(1L).provide(
  UserRepositoryLive.layer,
  Database.layer
)

// Partial provision
val partial: ZIO[Database, Throwable, User] =
  findUser(1L).provideSome[Database](UserRepositoryLive.layer)
```

### Testing with Layers

```scala
// Test layer
val testUserRepository: ULayer[UserRepository] =
  ZLayer.succeed {
    new UserRepository:
      private var users = Map(1L -> User(1L, "Test", "test@example.com"))
      def findById(id: Long) = ZIO.succeed(users.get(id))
      def save(user: User) = ZIO.succeed { users += (user.id -> user); user }
      def delete(id: Long) = ZIO.succeed { users -= id; true }
  }

// Test with layers
object UserServiceSpec extends ZIOSpecDefault:
  def spec = suite("UserService")(
    test("should find user") {
      for
        result <- UserService.findById(1L)
      yield assertTrue(result.isDefined)
    }
  ).provide(testUserRepository, UserService.layer)
```

---

## ZIO Streams

### Creating Streams

```scala
import zio.stream.*

// Basic streams
val numbers: ZStream[Any, Nothing, Int] = ZStream.range(1, 100)
val fromList: ZStream[Any, Nothing, String] = ZStream.fromIterable(List("a", "b", "c"))
val infinite: ZStream[Any, Nothing, Int] = ZStream.iterate(0)(_ + 1)
val ticks: ZStream[Any, Nothing, Long] = ZStream.repeatZIOWithSchedule(
  Clock.currentTime(TimeUnit.MILLISECONDS),
  Schedule.spaced(1.second)
)

// From effects
val fromEffect: ZStream[Any, Throwable, User] =
  ZStream.fromZIO(fetchUser(1L))

val fromQueue: ZStream[Any, Nothing, Event] =
  ZStream.fromQueue(eventQueue)
```

### Stream Transformations

```scala
// Basic transformations
val processed: ZStream[Any, Nothing, String] = numbers
  .filter(_ % 2 == 0)
  .map(_ * 2)
  .take(10)
  .mapZIO(n => Console.printLine(n).as(n.toString).orDie)

// Batching
val batched: ZStream[Any, Nothing, Chunk[Int]] =
  numbers.grouped(100)

val windowed: ZStream[Any, Nothing, Chunk[Int]] =
  numbers.groupedWithin(100, 5.seconds)

// Flatmap variants
val flatMapped: ZStream[Any, Nothing, Int] =
  numbers.flatMap(n => ZStream.range(0, n))

val concatted: ZStream[Any, Nothing, Int] =
  stream1 ++ stream2
```

### Stream Concurrency

```scala
// Parallel processing
val parallel: ZStream[Any, Throwable, Result] =
  events.mapZIOPar(10)(processEvent)

// Merge streams
val merged: ZStream[Any, Nothing, Int] =
  stream1.merge(stream2)

// Interleave
val interleaved: ZStream[Any, Nothing, Int] =
  stream1.interleaveWith(stream2)(Schedule.fixed(100.millis), Schedule.fixed(200.millis))

// Broadcast
val broadcasted: ZIO[Any, Nothing, (ZStream[Any, Nothing, Int], ZStream[Any, Nothing, Int])] =
  numbers.broadcast(2, 16)
```

### Stream Sinks

```scala
import zio.stream.ZSink

// Collecting results
val collected: ZIO[Any, Nothing, List[Int]] =
  numbers.run(ZSink.collectAll[Int]).map(_.toList)

// Folding
val sum: ZIO[Any, Nothing, Int] =
  numbers.run(ZSink.foldLeft(0)(_ + _))

// Head/Last
val first: ZIO[Any, Nothing, Option[Int]] =
  numbers.run(ZSink.head[Int])

// To file
val toFile: ZIO[Any, Throwable, Unit] =
  lines.run(ZSink.fromFile(Path.of("output.txt")))
```

---

## Resource Management

### Scoped Resources

```scala
// Define scoped resource
def connection: ZIO[Scope, Throwable, Connection] =
  ZIO.acquireRelease(
    ZIO.attempt(createConnection())
  )(conn => ZIO.succeed(conn.close()))

// Use scoped resource
val result: Task[QueryResult] =
  ZIO.scoped {
    for
      conn <- connection
      result <- conn.query("SELECT * FROM users")
    yield result
  }

// Multiple scoped resources
val multiResource: Task[Result] =
  ZIO.scoped {
    for
      db <- databaseConnection
      cache <- cacheConnection
      result <- processWithResources(db, cache)
    yield result
  }
```

### Scoped Streams

```scala
// Stream with resource lifecycle
def databaseRecords: ZStream[Scope, Throwable, Record] =
  ZStream.unwrapScoped {
    for
      conn <- connection
      stream = ZStream.fromIterator(conn.streamAll())
    yield stream
  }

// Finalization in streams
val withFinalization: ZStream[Any, Nothing, Int] =
  ZStream.acquireReleaseWith(acquire)(release).flatMap { resource =>
    ZStream.fromIterator(resource.iterator)
  }
```

---

## ZIO HTTP

### Basic Server

```scala
import zio.http.*

object Main extends ZIOAppDefault:
  val routes = Routes(
    Method.GET / "users" -> handler { (req: Request) =>
      for
        users <- UserService.findAll
      yield Response.json(users.toJson)
    },

    Method.GET / "users" / long("id") -> handler { (id: Long, req: Request) =>
      for
        user <- UserService.findById(id)
        response <- user match
          case Some(u) => ZIO.succeed(Response.json(u.toJson))
          case None => ZIO.succeed(Response.status(Status.NotFound))
      yield response
    },

    Method.POST / "users" -> handler { (req: Request) =>
      for
        body <- req.body.asString
        request <- ZIO.fromEither(body.fromJson[CreateUserRequest])
        user <- UserService.create(request)
      yield Response.json(user.toJson).status(Status.Created)
    }
  )

  def run = Server.serve(routes).provide(
    Server.default,
    UserServiceLive.layer,
    UserRepositoryLive.layer,
    Database.layer
  )
```

---

## ZIO Test

### Writing Tests

```scala
import zio.test.*
import zio.test.Assertion.*

object UserServiceSpec extends ZIOSpecDefault:
  def spec = suite("UserService")(
    test("should find existing user") {
      for
        result <- UserService.findById(1L)
      yield assertTrue(result.isDefined)
    },

    test("should return None for non-existent user") {
      for
        result <- UserService.findById(999L)
      yield assertTrue(result.isEmpty)
    },

    test("should handle parallel requests") {
      for
        results <- ZIO.foreachPar(1 to 100)(id => UserService.findById(id.toLong))
      yield assertTrue(results.flatten.nonEmpty)
    }
  ).provide(testUserRepositoryLayer, UserService.layer)
```

### Test Aspects

```scala
def spec = suite("UserService")(
  test("timeout test") {
    slowOperation
  } @@ TestAspect.timeout(5.seconds),

  test("flaky test") {
    flakyOperation
  } @@ TestAspect.flaky(3),

  test("ignored test") {
    ???
  } @@ TestAspect.ignore,

  test("sequential test") {
    sequentialOperation
  }
) @@ TestAspect.sequential
```

---

## Best Practices

Effect Types:
- Use typed errors for domain-specific failures
- Use Task for interop with external libraries
- Prefer ZIO.serviceWithZIO for service access

Layers:
- Define one layer per service implementation
- Use ZLayer.make for automatic wiring
- Create test layers that mirror production structure

Streams:
- Use groupedWithin for time-based batching
- Prefer mapZIOPar for CPU-bound parallel work
- Use broadcast for multiple consumers

Error Handling:
- Model errors as ADTs for exhaustive handling
- Use catchSome for partial recovery
- Use refineToOrDie for unexpected errors

---

Last Updated: 2026-01-06
Version: 2.0.0
