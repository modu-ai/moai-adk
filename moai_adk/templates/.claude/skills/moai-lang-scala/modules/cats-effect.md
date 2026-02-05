# Cats Effect 3.5

Comprehensive coverage of Cats Effect 3.5 including IO monad, Resources, Fibers, and FS2 Streaming.

## Context7 Library Mappings

```
/typelevel/cats-effect - Cats Effect 3.5 documentation
/typelevel/cats - Cats 2.10 functional abstractions
/co.fs2/fs2 - FS2 3.10 streaming
/http4s/http4s - Http4s 0.24 functional HTTP
/tpolecat/doobie - Doobie 1.0 functional JDBC
```

---

## IO Monad Basics

### Creating IO Values

```scala
import cats.effect.*
import cats.syntax.all.*

// Pure values
val pureIO: IO[Int] = IO.pure(42)
val delayedIO: IO[Int] = IO.delay(expensiveComputation())
val suspendedIO: IO[Int] = IO(sideEffectingComputation())

// Error handling
val failedIO: IO[Int] = IO.raiseError(new RuntimeException("Oops"))
val fromOption: IO[Int] = IO.fromOption(maybeInt)(new NoSuchElementException)
val fromEither: IO[Int] = IO.fromEither(eitherValue)
val fromTry: IO[Int] = IO.fromTry(tryValue)

// Console operations
val readLine: IO[String] = IO.readLine
val printLine: IO[Unit] = IO.println("Hello")

// Async operations
val asyncIO: IO[Int] = IO.async_ { callback =>
  someAsyncApi(result => callback(Right(result)))
}
```

### Composing IO

```scala
// For comprehension
def program: IO[Unit] =
  for
    _ <- IO.println("Enter your name:")
    name <- IO.readLine
    _ <- IO.println(s"Hello, $name!")
  yield ()

// Combinators
val combined: IO[(Int, String)] = (io1, io2).tupled
val sequenced: IO[String] = io1 *> io2  // Discard first result
val sequencedKeepFirst: IO[Int] = io1 <* io2  // Discard second result
val mapped: IO[String] = io1.map(_.toString)
val flatMapped: IO[String] = io1.flatMap(n => IO.pure(n.toString))

// Error handling
val recovered: IO[Int] = failedIO.handleErrorWith {
  case _: RuntimeException => IO.pure(0)
}
val attempted: IO[Either[Throwable, Int]] = io1.attempt
val redeemed: IO[String] = io1.redeem(
  err => s"Error: ${err.getMessage}",
  value => s"Success: $value"
)
```

---

## Resource Management

### Basic Resource Usage

```scala
import cats.effect.Resource
import java.io.*

// Creating resources
def fileResource(path: String): Resource[IO, BufferedReader] =
  Resource.make(
    IO(new BufferedReader(new FileReader(path)))
  )(reader => IO(reader.close()).handleError(_ => ()))

// Using resources
def readFirstLine(path: String): IO[String] =
  fileResource(path).use(reader => IO(reader.readLine()))

// Auto-closeable resources
def autoCloseableResource[A <: AutoCloseable](acquire: IO[A]): Resource[IO, A] =
  Resource.fromAutoCloseable(acquire)
```

### Composing Resources

```scala
// Sequential composition
def openFiles(path1: String, path2: String): Resource[IO, (BufferedReader, BufferedReader)] =
  for
    reader1 <- fileResource(path1)
    reader2 <- fileResource(path2)
  yield (reader1, reader2)

// Parallel acquisition
def parallelResources: Resource[IO, (Connection, Connection)] =
  (databaseConnection, cacheConnection).tupled

// Resource with finalizer
def connectionPool(size: Int): Resource[IO, ConnectionPool] =
  Resource.make(
    IO(ConnectionPool.create(size))
  )(pool => 
    pool.closeAll.handleError(e => 
      IO.println(s"Error closing pool: ${e.getMessage}")
    )
  )
```

### Resource Patterns

```scala
// Nested resources with proper cleanup
def applicationResources: Resource[IO, AppContext] =
  for
    config <- Resource.eval(loadConfig)
    pool <- connectionPool(config.poolSize)
    cache <- cacheResource(config.cacheSize)
    httpClient <- httpClientResource
  yield AppContext(config, pool, cache, httpClient)

// Background resource
def backgroundTask: Resource[IO, Fiber[IO, Throwable, Unit]] =
  Resource.make(
    periodicTask.start
  )(fiber => fiber.cancel)

// Finalizer ordering
def orderedCleanup: Resource[IO, Unit] =
  Resource.make(IO.println("Acquire A"))(_ => IO.println("Release A")) *>
  Resource.make(IO.println("Acquire B"))(_ => IO.println("Release B"))
  // Releases in reverse order: B then A
```

---

## Concurrency with Fibers

### Basic Fiber Operations

```scala
import cats.effect.std.*

// Starting fibers
val fiber: IO[Fiber[IO, Throwable, Int]] = expensiveComputation.start
val result: IO[Int] = fiber.flatMap(_.join).flatMap {
  case Outcome.Succeeded(fa) => fa
  case Outcome.Errored(e) => IO.raiseError(e)
  case Outcome.Canceled() => IO.raiseError(new Exception("Canceled"))
}

// Fire and forget
val background: IO[Unit] = longRunningTask.start.void

// Racing
val winner: IO[Either[String, Int]] = IO.race(io1, io2)
val faster: IO[Int] = IO.race(io1, io2).map(_.merge)
val raceOutcome: IO[Int] = IO.raceOutcome(io1, io2).flatMap {
  case Left(outcome) => outcome match
    case Outcome.Succeeded(fa) => fa
    case Outcome.Errored(e) => IO.raiseError(e)
    case Outcome.Canceled() => IO.pure(-1)
  case Right(outcome) => ???
}
```

### Parallel Execution

```scala
// Parallel map
def fetchUserData(userId: Long): IO[UserData] =
  (fetchUser(userId), fetchOrders(userId), fetchPreferences(userId))
    .parMapN(UserData.apply)

// Parallel traverse
def fetchAllUsers(ids: List[Long]): IO[List[User]] =
  ids.parTraverse(fetchUser)

// Parallel sequence
val allResults: IO[List[Int]] = listOfIOs.parSequence

// With parallelism limit
def rateLimited[A](tasks: List[IO[A]], maxConcurrent: Int): IO[List[A]] =
  tasks.parTraverseN(maxConcurrent)(identity)
```

### Cancellation

```scala
// Cancellable operations
val cancellable: IO[Unit] = IO.uncancelable { poll =>
  // Critical section - cannot be cancelled
  acquire.flatMap { resource =>
    poll(longOperation).guarantee(release(resource))
  }
}

// Timeout with cancellation
val withTimeout: IO[Int] = slowOperation.timeout(5.seconds)

// Bracket pattern
val bracketed: IO[String] = IO.bracket(acquire)(use)(release)
```

---

## Concurrent Data Structures

### Ref (Atomic Reference)

```scala
import cats.effect.Ref

// Creating and using Ref
def counter: IO[Unit] =
  for
    ref <- Ref.of[IO, Int](0)
    _ <- ref.update(_ + 1)
    _ <- ref.updateAndGet(_ + 1).flatMap(n => IO.println(s"Count: $n"))
    current <- ref.get
    _ <- IO.println(s"Final: $current")
  yield ()

// Atomic update
def atomicTransfer(from: Ref[IO, Int], to: Ref[IO, Int], amount: Int): IO[Unit] =
  for
    _ <- from.update(_ - amount)
    _ <- to.update(_ + amount)
  yield ()

// Modify with result
def dequeue[A](ref: Ref[IO, List[A]]): IO[Option[A]] =
  ref.modify {
    case Nil => (Nil, None)
    case head :: tail => (tail, Some(head))
  }
```

### Deferred (One-Shot Promise)

```scala
import cats.effect.Deferred

// Producer-consumer with Deferred
def producerConsumer: IO[Unit] =
  for
    deferred <- Deferred[IO, Int]
    producer = IO.sleep(1.second) *> deferred.complete(42)
    consumer = deferred.get.flatMap(n => IO.println(s"Got: $n"))
    _ <- (producer, consumer).parTupled
  yield ()
```

### Semaphore

```scala
import cats.effect.std.Semaphore

// Rate limiting with Semaphore
def rateLimitedRequests[A](tasks: List[IO[A]], maxConcurrent: Int): IO[List[A]] =
  Semaphore[IO](maxConcurrent).flatMap { sem =>
    tasks.parTraverse(task => sem.permit.use(_ => task))
  }

// Connection pool pattern
def withConnection[A](pool: Semaphore[IO])(use: Connection => IO[A]): IO[A] =
  pool.permit.use { _ =>
    acquireConnection.flatMap(use)
  }
```

### Queue

```scala
import cats.effect.std.Queue

// Bounded queue
def boundedProducerConsumer: IO[Unit] =
  for
    queue <- Queue.bounded[IO, Int](100)
    producer = (1 to 1000).toList.traverse_(n => queue.offer(n))
    consumer = queue.take.flatMap(n => IO.println(s"Consumed: $n")).foreverM
    _ <- (producer, consumer).parTupled
  yield ()

// With backpressure
def processWithBackpressure(queue: Queue[IO, Task]): IO[Unit] =
  queue.take.flatMap(processTask).foreverM
```

---

## FS2 Streaming

### Basic Streams

```scala
import fs2.*
import fs2.io.file.*

// Creating streams
val numbers: Stream[IO, Int] = Stream.range(1, 100)
val fromList: Stream[IO, String] = Stream.emits(List("a", "b", "c"))
val infinite: Stream[IO, Int] = Stream.iterate(0)(_ + 1)
val periodic: Stream[IO, Unit] = Stream.awakeEvery[IO](1.second)

// Transformations
val processed: Stream[IO, String] = numbers
  .filter(_ % 2 == 0)
  .map(_ * 2)
  .take(10)
  .evalMap(n => IO.println(n).as(n.toString))
```

### File Operations

```scala
// Reading files
def readLines(path: Path): Stream[IO, String] =
  Files[IO].readUtf8Lines(path)

def processLargeFile(path: Path): Stream[IO, ProcessedLine] =
  Files[IO].readUtf8Lines(path)
    .filter(_.nonEmpty)
    .map(_.toLowerCase)
    .evalTap(line => IO.println(s"Processing: $line"))
    .map(ProcessedLine.apply)

// Writing files
def writeResults(path: Path, lines: Stream[IO, String]): IO[Unit] =
  lines.intersperse("\n")
    .through(text.utf8.encode)
    .through(Files[IO].writeAll(path))
    .compile.drain
```

### Stream Concurrency

```scala
// Parallel processing
def parallelProcess[A, B](s: Stream[IO, A], parallelism: Int)(f: A => IO[B]): Stream[IO, B] =
  s.parEvalMap(parallelism)(f)

// Merge streams
val merged: Stream[IO, Int] = stream1.merge(stream2)
val interleaved: Stream[IO, Int] = stream1.interleave(stream2)

// Concurrent with limit
def processWithConcurrency(items: Stream[IO, Item]): Stream[IO, Result] =
  items.parEvalMapUnordered(maxConcurrent = 10)(processItem)
```

### Stream Resources

```scala
// Resource streams
def databaseRecords(pool: ConnectionPool): Stream[IO, Record] =
  Stream.resource(pool.connection).flatMap { conn =>
    Stream.evalSeq(conn.query("SELECT * FROM records"))
  }

// Bracket in streams
def processWithCleanup: Stream[IO, String] =
  Stream.bracket(acquire)(_ => release).flatMap { resource =>
    Stream.evalSeq(resource.getData)
  }
```

---

## HTTP with Http4s

### Basic Server

```scala
import org.http4s.*
import org.http4s.dsl.io.*
import org.http4s.ember.server.*
import org.http4s.circe.*
import io.circe.generic.auto.*

val routes: HttpRoutes[IO] = HttpRoutes.of[IO] {
  case GET -> Root / "users" =>
    Ok(userService.findAll)
  
  case GET -> Root / "users" / LongVar(id) =>
    userService.findById(id).flatMap {
      case Some(user) => Ok(user)
      case None => NotFound()
    }
  
  case req @ POST -> Root / "users" =>
    for
      user <- req.as[CreateUserRequest]
      created <- userService.create(user)
      response <- Created(created)
    yield response
}

def server: IO[Unit] =
  EmberServerBuilder.default[IO]
    .withHost(host"0.0.0.0")
    .withPort(port"8080")
    .withHttpApp(routes.orNotFound)
    .build
    .useForever
```

---

## Testing with MUnit

```scala
import munit.CatsEffectSuite

class UserServiceSpec extends CatsEffectSuite:
  test("should fetch user successfully") {
    val testUser = User(1L, "John", "john@example.com")
    val service = UserService.make(mockRepository)
    
    service.findById(1L).map { result =>
      assertEquals(result, Some(testUser))
    }
  }
  
  test("should handle concurrent operations") {
    val users = (1 to 10).map(i => User(i.toLong, s"User$i", s"user$i@example.com")).toList
    val service = UserService.make(mockRepository)
    
    users.parTraverse(u => service.findById(u.id)).map { results =>
      assertEquals(results.flatten.size, 10)
    }
  }
  
  test("should timeout slow operations") {
    val slowService = UserService.make(slowRepository)
    
    slowService.findById(1L)
      .timeout(100.millis)
      .attempt
      .map { result =>
        assert(result.isLeft)
      }
  }
```

---

## Best Practices

IO Creation:
- Use IO.pure for already computed values
- Use IO.delay or IO() for side effects
- Prefer IO.fromOption/Either/Try over manual conversion

Resource Management:
- Always use Resource for acquisitions that need cleanup
- Compose resources with for-comprehensions
- Resources release in reverse acquisition order

Concurrency:
- Use parMapN for independent parallel operations
- Use Semaphore for rate limiting
- Always handle Outcome.Canceled in fiber joins

Streaming:
- Use FS2 for large data processing
- Prefer parEvalMap over manual fiber management
- Use .compile.drain for side-effectful streams

---

Last Updated: 2026-01-06
Version: 2.0.0
