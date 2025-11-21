---
name: moai-lang-scala
description: Scala 3.6+ best practices with ScalaTest 3.2, sbt 1.10, functional programming
  patterns, and Play Framework.
---

## Quick Reference (30 seconds)

# Scala Functional Programming — Enterprise

## References (Latest Documentation)

_Documentation links updated 2025-11-22_

---

---

## Implementation Guide

## What It Does

Scala 3.6+ best practices with ScalaTest 3.2, sbt 1.10, functional programming patterns, and Play Framework.

**Key capabilities**:
- ✅ Best practices enforcement for language domain
- ✅ TRUST 5 principles integration
- ✅ Latest tool versions (2025-11-22)
- ✅ TDD workflow support
- ✅ Play Framework web application patterns
- ✅ Scala 3 modern syntax (Given/Using, Enums, Opaque Types)
- ✅ Functional programming with ZIO & Cats Effect
- ✅ Big data processing with Spark & Flink

---

## When to Use

**Automatic triggers**:
- Related code discussions and file patterns
- SPEC implementation (`/alfred:2-run`)
- Code review requests

**Manual invocation**:
- Review code for TRUST 5 compliance
- Design new features
- Troubleshoot issues

---

## Tool Version Matrix (2025-11-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Scala** | 3.6.0 | Runtime | ✅ Current |
| **ScalaTest** | 3.2.19 | Testing | ✅ Current |
| **sbt** | 1.10.0 | Build tool | ✅ Current |
| **Play Framework** | 3.0.9 | Web framework | ✅ Current |
| **ZIO** | 2.1.0 | Effect system | ✅ Current |
| **Cats** | 2.12.0 | FP library | ✅ Current |
| **Spark** | 3.5.3 | Big data | ✅ Current |

---

## Scala 3 Modern Syntax

### Given/Using for Context Abstraction

```scala
// Scala 3 given/using syntax replaces implicits
trait Ord[T]:
  def compare(x: T, y: T): Int
  extension (x: T)
    def < (y: T): Boolean = compare(x, y) < 0
    def > (y: T): Boolean = compare(x, y) > 0

given Ord[Int] with
  def compare(x: Int, y: Int): Int = 
    if x < y then -1 else if x > y then 1 else 0

given Ord[String] with
  def compare(x: String, y: String): Int = 
    x.compareTo(y)

// Using context parameters
def max[T](x: T, y: T)(using ord: Ord[T]): T =
  if ord.compare(x, y) > 0 then x else y

// Automatic resolution
max(10, 20)  // Uses given Ord[Int]
max("hello", "world")  // Uses given Ord[String]

// Context bounds
def sort[T: Ord](xs: List[T]): List[T] =
  xs.sortWith(summon[Ord[T]].compare(_, _) < 0)
```

### Enums with Type Safety

```scala
// Scala 3 enums are powerful algebraic data types
enum Status:
  case Pending, Approved, Rejected

enum HttpResponse:
  case Success(data: String)
  case Error(code: Int, message: String)
  case Redirect(url: String)

// Pattern matching
def handleResponse(response: HttpResponse): String = response match
  case HttpResponse.Success(data) => s"Data: $data"
  case HttpResponse.Error(code, msg) => s"Error $code: $msg"
  case HttpResponse.Redirect(url) => s"Redirect to $url"

// Enums with methods
enum Color(val rgb: Int):
  case Red extends Color(0xFF0000)
  case Green extends Color(0x00FF00)
  case Blue extends Color(0x0000FF)
  
  def hex: String = f"#$rgb%06X"

Color.Red.hex  // "#FF0000"
```

### Opaque Types for Type Safety

```scala
// Opaque types provide zero-cost abstractions
object UserTypes:
  opaque type UserId = Long
  opaque type Email = String
  
  object UserId:
    def apply(value: Long): UserId = value
    extension (id: UserId)
      def toLong: Long = id
  
  object Email:
    def apply(value: String): Either[String, Email] =
      if value.contains("@") then Right(value)
      else Left("Invalid email format")
    
    extension (email: Email)
      def value: String = email

import UserTypes._

// Type-safe usage
val userId: UserId = UserId(12345)
val email: Email = Email("user@example.com").getOrElse(???)

// Compile error: userId and email are different types
// val x: UserId = email  // Error!

// Runtime: no wrapper overhead
userId.toLong  // Direct access, no boxing
```

### Union Types & Intersection Types

```scala
// Union types (Scala 3)
type StringOrInt = String | Int

def process(value: StringOrInt): String = value match
  case s: String => s"String: $s"
  case i: Int => s"Int: $i"

// Intersection types
trait Loggable:
  def log(): Unit

trait Serializable:
  def serialize(): String

// Both traits required
def saveAndLog(obj: Loggable & Serializable): Unit =
  obj.log()
  val data = obj.serialize()
  // Save data...

// Type refinement
type PositiveInt = Int & Positive
// where Positive is a refinement type
```

### Extension Methods

```scala
// Scala 3 extension methods
extension (s: String)
  def truncate(maxLength: Int): String =
    if s.length <= maxLength then s
    else s"${s.take(maxLength)}..."
  
  def toTitleCase: String =
    s.split(" ").map(_.capitalize).mkString(" ")

// Usage
"hello world".toTitleCase  // "Hello World"
"long string".truncate(4)  // "long..."

// Generic extensions
extension [T](xs: List[T])
  def second: Option[T] = xs.drop(1).headOption
  
  def groupByCount[K](f: T => K): Map[K, Int] =
    xs.groupBy(f).view.mapValues(_.size).toMap

List(1, 2, 3).second  // Some(2)
List("a", "ab", "abc").groupByCount(_.length)  // Map(1 -> 1, 2 -> 1, 3 -> 1)
```

---

## Advanced Functional Programming

### Monads & Functors with Cats

```scala
import cats._
import cats.implicits._

// Functor: map over context
val listFunctor = Functor[List]
listFunctor.map(List(1, 2, 3))(_ * 2)  // List(2, 4, 6)

// Applicative: combine independent effects
import cats.data.Validated

case class User(name: String, age: Int)

def validateName(name: String): Validated[String, String] =
  if name.nonEmpty then name.valid
  else "Name cannot be empty".invalid

def validateAge(age: Int): Validated[String, Int] =
  if age > 0 then age.valid
  else "Age must be positive".invalid

// Parallel validation
(validateName("Alice"), validateAge(25)).mapN(User.apply)
// Valid(User("Alice", 25))

(validateName(""), validateAge(-5)).mapN(User.apply)
// Invalid("Name cannot be empty, Age must be positive")

// Monad: flatMap for sequential operations
import cats.effect.IO

def readFile(path: String): IO[String] = IO(???)
def parseJson(content: String): IO[Json] = IO(???)
def validateJson(json: Json): IO[ValidJson] = IO(???)

// Monadic composition
val program: IO[ValidJson] = for
  content <- readFile("data.json")
  json <- parseJson(content)
  valid <- validateJson(json)
yield valid
```

### Monad Transformers for Nested Effects

```scala
import cats.data.{EitherT, OptionT}
import cats.effect.IO

// EitherT for IO[Either[Error, A]]
type Result[A] = EitherT[IO, String, A]

def findUser(id: Long): Result[User] =
  EitherT(IO {
    // Database query
    if id > 0 then Right(User(id, "Alice"))
    else Left("Invalid user ID")
  })

def getUserPosts(user: User): Result[List[Post]] =
  EitherT(IO {
    Right(List(Post(1, "First post")))
  })

// Monadic composition without nesting
val program: Result[List[Post]] = for
  user <- findUser(123)
  posts <- getUserPosts(user)
yield posts

program.value  // IO[Either[String, List[Post]]]

// OptionT for IO[Option[A]]
def findUserOpt(id: Long): OptionT[IO, User] =
  OptionT(IO(Some(User(id, "Bob"))))

def getProfile(user: User): OptionT[IO, Profile] =
  OptionT(IO(Some(Profile(user.id, "Developer"))))

val profileProgram: OptionT[IO, Profile] = for
  user <- findUserOpt(456)
  profile <- getProfile(user)
yield profile
```

### Free Monads for Embedded DSLs

```scala
import cats.free.Free
import cats.free.Free.liftF

// DSL operations
sealed trait KVStoreOp[A]
case class Get(key: String) extends KVStoreOp[Option[String]]
case class Put(key: String, value: String) extends KVStoreOp[Unit]
case class Delete(key: String) extends KVStoreOp[Unit]

type KVStore[A] = Free[KVStoreOp, A]

// Smart constructors
def get(key: String): KVStore[Option[String]] = liftF(Get(key))
def put(key: String, value: String): KVStore[Unit] = liftF(Put(key, value))
def delete(key: String): KVStore[Unit] = liftF(Delete(key))

// Program in DSL
def program: KVStore[Option[String]] = for
  _ <- put("username", "alice")
  _ <- put("email", "alice@example.com")
  username <- get("username")
  _ <- delete("email")
yield username

// Interpreter
def interpreter: KVStoreOp ~> IO = new (KVStoreOp ~> IO):
  val kvs = scala.collection.mutable.Map.empty[String, String]
  
  def apply[A](op: KVStoreOp[A]): IO[A] = op match
    case Get(key) => IO(kvs.get(key))
    case Put(key, value) => IO { kvs(key) = value }
    case Delete(key) => IO { kvs.remove(key) }

// Run program
program.foldMap(interpreter)
```

---

## ZIO Effect System

### ZIO Basics

```scala
import zio._

// Pure values
val meaningOfLife: UIO[Int] = ZIO.succeed(42)

// Effectful computations
def readFile(path: String): Task[String] =
  ZIO.attempt(scala.io.Source.fromFile(path).mkString)

// Error handling
def parseJson(content: String): IO[ParseError, Json] =
  ZIO.attempt(parse(content))
     .refineToOrDie[ParseError]

// Parallel execution
val parallel: UIO[Int] = for
  fiber1 <- ZIO.succeed(1).fork
  fiber2 <- ZIO.succeed(2).fork
  v1 <- fiber1.join
  v2 <- fiber2.join
yield v1 + v2

// Racing
val raced: UIO[Int] =
  ZIO.succeed(1).delay(1.second)
     .race(ZIO.succeed(2).delay(2.seconds))
```

### ZIO Layers for Dependency Injection

```scala
import zio._

// Service definitions
trait Database:
  def query(sql: String): Task[List[Row]]

trait Cache:
  def get(key: String): Task[Option[String]]
  def set(key: String, value: String): Task[Unit]

// Service implementations
case class PostgresDatabase(config: DbConfig) extends Database:
  def query(sql: String): Task[List[Row]] = ???

case class RedisCache(client: RedisClient) extends Cache:
  def get(key: String): Task[Option[String]] = ???
  def set(key: String, value: String): Task[Unit] = ???

// Layers
object Database:
  val live: ZLayer[DbConfig, Throwable, Database] =
    ZLayer.fromFunction(PostgresDatabase.apply)

object Cache:
  val live: ZLayer[RedisClient, Throwable, Cache] =
    ZLayer.fromFunction(RedisCache.apply)

// Application layer
val appLayer: ZLayer[Any, Throwable, Database & Cache] =
  ZLayer.make[Database & Cache](
    Database.live,
    Cache.live,
    DbConfig.live,
    RedisClient.live
  )

// Using services
def getUserData(userId: Long): ZIO[Database & Cache, Throwable, UserData] =
  for
    cache <- ZIO.service[Cache]
    db <- ZIO.service[Database]
    cached <- cache.get(s"user:$userId")
    result <- cached match
      case Some(data) => ZIO.succeed(deserialize(data))
      case None =>
        for
          rows <- db.query(s"SELECT * FROM users WHERE id = $userId")
          userData = toUserData(rows)
          _ <- cache.set(s"user:$userId", serialize(userData))
        yield userData
  yield result

// Run with dependencies
val program = getUserData(123).provide(appLayer)
```

---

## Akka & Actor Model

### Actor Definition & Message Handling

```scala
import akka.actor.typed._
import akka.actor.typed.scaladsl._

// Message protocol
sealed trait UserCommand
case class CreateUser(name: String, email: String, replyTo: ActorRef[UserCreated]) extends UserCommand
case class GetUser(id: Long, replyTo: ActorRef[UserResponse]) extends UserCommand
case class DeleteUser(id: Long) extends UserCommand

sealed trait UserEvent
case class UserCreated(id: Long) extends UserEvent
case class UserResponse(user: Option[User]) extends UserEvent

// Actor behavior
object UserManager:
  def apply(): Behavior[UserCommand] =
    Behaviors.setup { context =>
      var users = Map.empty[Long, User]
      var nextId = 1L
      
      Behaviors.receiveMessage {
        case CreateUser(name, email, replyTo) =>
          val user = User(nextId, name, email)
          users = users + (nextId -> user)
          context.log.info(s"Created user: $user")
          replyTo ! UserCreated(nextId)
          nextId += 1
          Behaviors.same
        
        case GetUser(id, replyTo) =>
          replyTo ! UserResponse(users.get(id))
          Behaviors.same
        
        case DeleteUser(id) =>
          users = users - id
          context.log.info(s"Deleted user: $id")
          Behaviors.same
      }
    }

// Usage
val system: ActorSystem[UserCommand] = 
  ActorSystem(UserManager(), "user-system")

implicit val timeout: Timeout = 3.seconds
implicit val scheduler: Scheduler = system.scheduler

val createUser: Future[UserCreated] = 
  system.ask(replyTo => CreateUser("Alice", "alice@example.com", replyTo))
```

### Akka Persistence for Event Sourcing

```scala
import akka.persistence.typed.PersistenceId
import akka.persistence.typed.scaladsl._

// Commands
sealed trait AccountCommand
case class Deposit(amount: BigDecimal) extends AccountCommand
case class Withdraw(amount: BigDecimal, replyTo: ActorRef[WithdrawResult]) extends AccountCommand
case class GetBalance(replyTo: ActorRef[BigDecimal]) extends AccountCommand

// Events
sealed trait AccountEvent
case class Deposited(amount: BigDecimal) extends AccountEvent
case class Withdrawn(amount: BigDecimal) extends AccountEvent

// Result
sealed trait WithdrawResult
case object Successful extends WithdrawResult
case object InsufficientFunds extends WithdrawResult

// State
case class AccountState(balance: BigDecimal)

// Persistent actor
object AccountActor:
  def apply(accountId: String): EventSourcedBehavior[AccountCommand, AccountEvent, AccountState] =
    EventSourcedBehavior(
      persistenceId = PersistenceId.ofUniqueId(accountId),
      emptyState = AccountState(0),
      commandHandler = commandHandler,
      eventHandler = eventHandler
    )
  
  private def commandHandler: (AccountState, AccountCommand) => Effect[AccountEvent, AccountState] = {
    (state, command) => command match
      case Deposit(amount) =>
        Effect.persist(Deposited(amount))
      
      case Withdraw(amount, replyTo) =>
        if state.balance >= amount then
          Effect
            .persist(Withdrawn(amount))
            .thenReply(replyTo)(_ => Successful)
        else
          Effect.reply(replyTo)(InsufficientFunds)
      
      case GetBalance(replyTo) =>
        Effect.reply(replyTo)(state.balance)
  }
  
  private def eventHandler: (AccountState, AccountEvent) => AccountState = {
    (state, event) => event match
      case Deposited(amount) => state.copy(balance = state.balance + amount)
      case Withdrawn(amount) => state.copy(balance = state.balance - amount)
  }
```

---

## Big Data Processing

### Apache Spark Integration

```scala
import org.apache.spark.sql._
import org.apache.spark.sql.functions._

// Spark session
val spark = SparkSession.builder()
  .appName("DataProcessing")
  .master("local[*]")
  .getOrCreate()

import spark.implicits._

// Load data
val usersDF = spark.read
  .option("header", "true")
  .option("inferSchema", "true")
  .csv("users.csv")

// Transformations
val result = usersDF
  .filter($"age" > 18)
  .groupBy($"country")
  .agg(
    count("*").as("user_count"),
    avg("age").as("avg_age"),
    max("registration_date").as("latest_registration")
  )
  .orderBy($"user_count".desc)

// Window functions
val windowSpec = Window.partitionBy($"country").orderBy($"revenue".desc)

val rankedUsers = usersDF
  .withColumn("rank", row_number().over(windowSpec))
  .filter($"rank" <= 10)

// User-defined functions
val categorizeAge = udf((age: Int) => age match
  case a if a < 18 => "Minor"
  case a if a < 65 => "Adult"
  case _ => "Senior"
)

val categorized = usersDF
  .withColumn("age_category", categorizeAge($"age"))

// Save result
result.write
  .mode("overwrite")
  .parquet("output/user_stats")
```

### Flink Streaming

```scala
import org.apache.flink.streaming.api.scala._
import org.apache.flink.streaming.api.windowing.time.Time

case class Event(userId: Long, eventType: String, timestamp: Long)

val env = StreamExecutionEnvironment.getExecutionEnvironment

// Kafka source
val events: DataStream[Event] = env
  .addSource(new FlinkKafkaConsumer[String]("events", schema, properties))
  .map(parseEvent)

// Windowed aggregation
val eventCounts = events
  .keyBy(_.userId)
  .timeWindow(Time.minutes(5))
  .aggregate(new CountAggregator)

// Complex event processing
val alerts = events
  .keyBy(_.userId)
  .flatMap(new SuspiciousActivityDetector)
  .filter(_.severity == "high")

// Sink
alerts.addSink(new AlertingSink)

env.execute("Event Processing")
```

---

## Play Framework Web Development

### Controller with Dependency Injection

```scala
import play.api.mvc._
import play.api.libs.json._
import javax.inject._

@Singleton
class UserController @Inject()(
  cc: ControllerComponents,
  userService: UserService
) extends AbstractController(cc):
  
  implicit val userWrites: Writes[User] = Json.writes[User]
  implicit val userReads: Reads[User] = Json.reads[User]
  
  def list(): Action[AnyContent] = Action.async { implicit request =>
    userService.findAll().map { users =>
      Ok(Json.toJson(users))
    }
  }
  
  def show(id: Long): Action[AnyContent] = Action.async {
    userService.findById(id).map {
      case Some(user) => Ok(Json.toJson(user))
      case None => NotFound
    }
  }
  
  def create(): Action[JsValue] = Action.async(parse.json) { request =>
    request.body.validate[User].fold(
      errors => Future.successful(BadRequest(Json.obj("errors" -> JsError.toJson(errors)))),
      user => userService.create(user).map { created =>
        Created(Json.toJson(created))
      }
    )
  }
  
  def update(id: Long): Action[JsValue] = Action.async(parse.json) { request =>
    request.body.validate[User].fold(
      errors => Future.successful(BadRequest(Json.obj("errors" -> JsError.toJson(errors)))),
      user => userService.update(id, user).map {
        case Some(updated) => Ok(Json.toJson(updated))
        case None => NotFound
      }
    )
  }
  
  def delete(id: Long): Action[AnyContent] = Action.async {
    userService.delete(id).map { _ =>
      NoContent
    }
  }
```

### Routes Configuration

```routes
# conf/routes

# Users
GET     /api/users              controllers.UserController.list
GET     /api/users/:id          controllers.UserController.show(id: Long)
POST    /api/users              controllers.UserController.create
PUT     /api/users/:id          controllers.UserController.update(id: Long)
DELETE  /api/users/:id          controllers.UserController.delete(id: Long)
```

---

## Testing with ScalaTest

### BDD Style Tests

```scala
import org.scalatest.flatspec.AnyFlatSpec
import org.scalatest.matchers.should.Matchers

class UserServiceSpec extends AnyFlatSpec with Matchers:
  
  "UserService" should "create a user with valid data" in {
    val service = new UserService
    val user = service.create("Alice", "alice@example.com")
    
    user.name shouldBe "Alice"
    user.email shouldBe "alice@example.com"
    user.id should be > 0L
  }
  
  it should "reject invalid email addresses" in {
    val service = new UserService
    
    an [ValidationException] should be thrownBy {
      service.create("Bob", "invalid-email")
    }
  }
  
  it should "find user by ID" in {
    val service = new UserService
    val created = service.create("Charlie", "charlie@example.com")
    
    val found = service.findById(created.id)
    
    found shouldBe Some(created)
  }
  
  "UserRepository" should "return None for non-existent ID" in {
    val repo = new UserRepository
    
    repo.findById(999) shouldBe None
  }
```

### Property-Based Testing with ScalaCheck

```scala
import org.scalatest.propspec.AnyPropSpec
import org.scalatestplus.scalacheck.ScalaCheckPropertyChecks

class MathPropertiesSpec extends AnyPropSpec with ScalaCheckPropertyChecks:
  
  property("addition is commutative") {
    forAll { (a: Int, b: Int) =>
      whenever(willNotOverflow(a, b)) {
        add(a, b) should equal(add(b, a))
      }
    }
  }
  
  property("string reversal is involutive") {
    forAll { (s: String) =>
      s.reverse.reverse should equal(s)
    }
  }
  
  property("list concatenation associativity") {
    forAll { (xs: List[Int], ys: List[Int], zs: List[Int]) =>
      (xs ++ ys) ++ zs should equal(xs ++ (ys ++ zs))
    }
  }
```

---

## Inputs

- Language-specific source directories
- Configuration files
- Test suites and sample data

## Outputs

- Test/lint execution plan
- TRUST 5 review checkpoints
- Migration guidance

## Failure Modes

- When required tools are not installed
- When dependencies are missing
- When test coverage falls below 85%

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-langs` for language detection
- Integration with `moai-foundation-trust` for quality gates

---

## Changelog

- **v3.0.0** (2025-11-22): Massive expansion with Scala 3 features, functional programming patterns, ZIO, Akka, big data processing
- **v2.0.0** (2025-10-22): Major update with latest tool versions, comprehensive best practices, TRUST 5 integration
- **v1.0.0** (2025-03-29): Initial Skill release

---

## Works Well With

- `moai-foundation-trust` (quality gates)
- `moai-alfred-code-reviewer` (code review)
- `moai-essentials-debug` (debugging support)
- `moai-domain-backend` (backend architecture)

---

## Best Practices

✅ **DO**:
- Use Scala 3 syntax (given/using, enums, opaque types)
- Leverage effect systems (ZIO or Cats Effect)
- Apply functional programming patterns
- Use immutable data structures
- Implement type classes for abstraction
- Write property-based tests with ScalaCheck
- Use Akka for distributed systems
- Profile JVM performance with JFR

❌ **DON'T**:
- Use null (use Option instead)
- Mutate variables unnecessarily
- Use var when val suffices
- Ignore compiler warnings
- Mix imperative and functional styles inconsistently
- Skip type annotations on public APIs
- Use deprecated Scala 2 implicit syntax
- Optimize without profiling

---

## Advanced Patterns




---

## Context7 Integration

### Related Libraries & Tools
- [Scala](/scala/scala): Functional programming language
- [ZIO](/zio/zio): Effect system
- [Cats](/typelevel/cats): Functional programming library
- [Akka](/akka/akka): Actor model toolkit
- [Spark](/apache/spark): Big data processing

### Official Documentation
- [Documentation](https://docs.scala-lang.org/)
- [API Reference](https://scala-lang.org/api/)
- [ZIO Documentation](https://zio.dev/)
- [Cats Documentation](https://typelevel.org/cats/)
- [Akka Documentation](https://doc.akka.io/)

### Version-Specific Guides
Latest stable version: Scala 3.6
- [Scala 3 Migration](https://docs.scala-lang.org/scala3/guides/migration/compatibility-intro.html)
- [New in Scala 3](https://docs.scala-lang.org/scala3/new-in-scala3.html)
- [Opaque Types](https://docs.scala-lang.org/scala3/reference/other-new-features/opaques.html)
