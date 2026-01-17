# Scala Production Examples

## REST API Implementations

### Http4s Functional HTTP Service

```scala
// Main.scala
import cats.effect.*
import org.http4s.*
import org.http4s.dsl.io.*
import org.http4s.ember.server.*
import org.http4s.server.Router
import com.comcast.ip4s.*

object Main extends IOApp.Simple:
  def run: IO[Unit] =
    for
      config <- Config.load
      xa <- Database.transactor(config.database)
      repository = UserRepository.make(xa)
      service = UserService.make(repository)
      httpApp = Router(
        "/api/v1" -> UserRoutes(service).routes
      ).orNotFound
      _ <- EmberServerBuilder
        .default[IO]
        .withHost(config.server.host)
        .withPort(config.server.port)
        .withHttpApp(httpApp)
        .build
        .useForever
    yield ()

// UserRoutes.scala
import cats.effect.*
import org.http4s.*
import org.http4s.dsl.io.*
import org.http4s.circe.*
import io.circe.generic.auto.*

class UserRoutes(service: UserService[IO]) extends Http4sDsl[IO]:
  given EntityDecoder[IO, CreateUserRequest] = jsonOf[IO, CreateUserRequest]
  given EntityDecoder[IO, UpdateUserRequest] = jsonOf[IO, UpdateUserRequest]
  given EntityEncoder[IO, User] = jsonEncoderOf[IO, User]
  given EntityEncoder[IO, List[User]] = jsonEncoderOf[IO, List[User]]

  object PageParam extends OptionalQueryParamDecoderMatcher[Int]("page")
  object SizeParam extends OptionalQueryParamDecoderMatcher[Int]("size")

  val routes: HttpRoutes[IO] = HttpRoutes.of[IO] {
    case GET -> Root / "users" :? PageParam(page) +& SizeParam(size) =>
      for
        users <- service.findAll(page.getOrElse(0), size.getOrElse(20))
        response <- Ok(users)
      yield response

    case GET -> Root / "users" / LongVar(id) =>
      service.findById(id).flatMap {
        case Some(user) => Ok(user)
        case None => NotFound()
      }

    case req @ POST -> Root / "users" =>
      for
        request <- req.as[CreateUserRequest]
        result <- service.create(request).attempt
        response <- result match
          case Right(user) => Created(user)
          case Left(_: DuplicateEmailException) => Conflict()
          case Left(e) => InternalServerError(e.getMessage)
      yield response

    case req @ PUT -> Root / "users" / LongVar(id) =>
      for
        request <- req.as[UpdateUserRequest]
        result <- service.update(id, request)
        response <- result match
          case Some(user) => Ok(user)
          case None => NotFound()
      yield response

    case DELETE -> Root / "users" / LongVar(id) =>
      service.delete(id).flatMap {
        case true => NoContent()
        case false => NotFound()
      }
  }

// UserService.scala
trait UserService[F[_]]:
  def findAll(page: Int, size: Int): F[List[User]]
  def findById(id: Long): F[Option[User]]
  def create(request: CreateUserRequest): F[User]
  def update(id: Long, request: UpdateUserRequest): F[Option[User]]
  def delete(id: Long): F[Boolean]

object UserService:
  def make(repository: UserRepository[IO]): UserService[IO] = new UserService[IO]:
    def findAll(page: Int, size: Int): IO[List[User]] =
      repository.findAll(page * size, size)

    def findById(id: Long): IO[Option[User]] =
      repository.findById(id)

    def create(request: CreateUserRequest): IO[User] =
      for
        exists <- repository.existsByEmail(request.email)
        _ <- IO.raiseWhen(exists)(DuplicateEmailException(request.email))
        passwordHash = BCrypt.hashpw(request.password, BCrypt.gensalt())
        user = User(0, request.name, request.email, passwordHash, UserStatus.Pending, Instant.now)
        saved <- repository.save(user)
      yield saved

    def update(id: Long, request: UpdateUserRequest): IO[Option[User]] =
      repository.findById(id).flatMap {
        case Some(existing) =>
          val checkEmail = request.email.filter(_ != existing.email).traverse_ { email =>
            repository.existsByEmail(email).flatMap { exists =>
              IO.raiseWhen(exists)(DuplicateEmailException(email))
            }
          }
          val updated = existing.copy(
            name = request.name,
            email = request.email.getOrElse(existing.email)
          )
          checkEmail *> repository.update(updated).map(Some(_))
        case None => IO.pure(None)
      }

    def delete(id: Long): IO[Boolean] =
      repository.delete(id)

// UserRepository.scala (Doobie)
trait UserRepository[F[_]]:
  def findAll(offset: Int, limit: Int): F[List[User]]
  def findById(id: Long): F[Option[User]]
  def findByEmail(email: String): F[Option[User]]
  def existsByEmail(email: String): F[Boolean]
  def save(user: User): F[User]
  def update(user: User): F[User]
  def delete(id: Long): F[Boolean]

object UserRepository:
  def make(xa: Transactor[IO]): UserRepository[IO] = new UserRepository[IO]:
    import doobie.*
    import doobie.implicits.*
    import doobie.postgres.implicits.*

    def findAll(offset: Int, limit: Int): IO[List[User]] =
      sql"""
        SELECT id, name, email, password_hash, status, created_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $limit OFFSET $offset
      """.query[User].to[List].transact(xa)

    def findById(id: Long): IO[Option[User]] =
      sql"""
        SELECT id, name, email, password_hash, status, created_at
        FROM users WHERE id = $id
      """.query[User].option.transact(xa)

    def findByEmail(email: String): IO[Option[User]] =
      sql"""
        SELECT id, name, email, password_hash, status, created_at
        FROM users WHERE email = $email
      """.query[User].option.transact(xa)

    def existsByEmail(email: String): IO[Boolean] =
      sql"SELECT EXISTS(SELECT 1 FROM users WHERE email = $email)"
        .query[Boolean].unique.transact(xa)

    def save(user: User): IO[User] =
      sql"""
        INSERT INTO users (name, email, password_hash, status, created_at)
        VALUES (${user.name}, ${user.email}, ${user.passwordHash}, ${user.status}, ${user.createdAt})
      """.update.withUniqueGeneratedKeys[Long]("id")
        .map(id => user.copy(id = id))
        .transact(xa)

    def update(user: User): IO[User] =
      sql"""
        UPDATE users SET name = ${user.name}, email = ${user.email}
        WHERE id = ${user.id}
      """.update.run.transact(xa).as(user)

    def delete(id: Long): IO[Boolean] =
      sql"DELETE FROM users WHERE id = $id".update.run.transact(xa).map(_ > 0)

// Models.scala
import io.circe.*
import java.time.Instant

case class User(
  id: Long,
  name: String,
  email: String,
  passwordHash: String,
  status: UserStatus,
  createdAt: Instant
) derives Encoder.AsObject, Decoder

enum UserStatus derives Encoder, Decoder:
  case Pending, Active, Suspended

case class CreateUserRequest(
  name: String,
  email: String,
  password: String
) derives Decoder

case class UpdateUserRequest(
  name: String,
  email: Option[String] = None
) derives Decoder

class DuplicateEmailException(email: String)
  extends RuntimeException(s"Email already exists: $email")
```

---

## Big Data Examples

### Spark 3.5 Analytics

```scala
// UserAnalytics.scala
import org.apache.spark.sql.{DataFrame, SparkSession}
import org.apache.spark.sql.functions.*

object UserAnalytics:
  def main(args: Array[String]): Unit =
    val spark = SparkSession.builder()
      .appName("User Analytics")
      .config("spark.sql.adaptive.enabled", "true")
      .config("spark.sql.shuffle.partitions", "200")
      .getOrCreate()

    import spark.implicits.*

    val users = spark.read.parquet("s3://data/users")
    val orders = spark.read.parquet("s3://data/orders")
    val events = spark.read.parquet("s3://data/events")

    // User lifetime value analysis
    val userLtv = calculateUserLtv(users, orders)

    // User engagement metrics
    val engagement = calculateEngagement(users, events)

    // Cohort analysis
    val cohorts = performCohortAnalysis(users, orders)

    userLtv.write.parquet("s3://output/user-ltv")
    engagement.write.parquet("s3://output/user-engagement")
    cohorts.write.parquet("s3://output/cohorts")

    spark.stop()

  def calculateUserLtv(users: DataFrame, orders: DataFrame): DataFrame =
    orders
      .groupBy("user_id")
      .agg(
        sum("amount").as("total_spent"),
        count("*").as("order_count"),
        avg("amount").as("avg_order_value"),
        min("created_at").as("first_order"),
        max("created_at").as("last_order")
      )
      .join(users, Seq("user_id"), "left")
      .withColumn("days_as_customer",
        datediff(col("last_order"), col("first_order")))
      .withColumn("ltv_score",
        col("total_spent") * (col("order_count") / (col("days_as_customer") + 1)))

  def calculateEngagement(users: DataFrame, events: DataFrame): DataFrame =
    events
      .filter(col("event_date") >= date_sub(current_date(), 30))
      .groupBy("user_id")
      .agg(
        countDistinct("session_id").as("sessions"),
        count("*").as("total_events"),
        sum(when(col("event_type") === "page_view", 1).otherwise(0)).as("page_views"),
        sum(when(col("event_type") === "click", 1).otherwise(0)).as("clicks")
      )
      .join(users, Seq("user_id"), "left")
      .withColumn("engagement_score",
        (col("sessions") * 0.3) + (col("page_views") * 0.2) + (col("clicks") * 0.5))

  def performCohortAnalysis(users: DataFrame, orders: DataFrame): DataFrame =
    val usersWithCohort = users
      .withColumn("cohort_month", date_trunc("month", col("created_at")))

    val ordersWithPeriod = orders
      .withColumn("order_month", date_trunc("month", col("created_at")))

    usersWithCohort
      .join(ordersWithPeriod, "user_id")
      .withColumn("period_number",
        months_between(col("order_month"), col("cohort_month")).cast("int"))
      .groupBy("cohort_month", "period_number")
      .agg(
        countDistinct("user_id").as("users"),
        sum("amount").as("revenue")
      )
      .orderBy("cohort_month", "period_number")
```

### Akka Streams Processing

```scala
// StreamProcessing.scala
import akka.actor.typed.ActorSystem
import akka.actor.typed.scaladsl.Behaviors
import akka.stream.scaladsl.*
import akka.stream.alpakka.kafka.scaladsl.*
import akka.kafka.{ConsumerSettings, ProducerSettings}
import org.apache.kafka.common.serialization.*
import scala.concurrent.duration.*

object StreamProcessing:
  given system: ActorSystem[Nothing] = ActorSystem(Behaviors.empty, "stream-system")
  given ec: ExecutionContext = system.executionContext

  val consumerSettings = ConsumerSettings(system, new StringDeserializer, new ByteArrayDeserializer)
    .withBootstrapServers("localhost:9092")
    .withGroupId("processor-group")

  val producerSettings = ProducerSettings(system, new StringSerializer, new ByteArraySerializer)
    .withBootstrapServers("localhost:9092")

  def processEvents(): Future[Done] =
    Consumer
      .plainSource(consumerSettings, Subscriptions.topics("user-events"))
      .map(record => parseEvent(record.value()))
      .filter(_.isValid)
      .mapAsync(4)(enrichEvent)
      .groupedWithin(100, 5.seconds)
      .mapAsync(2)(batchProcess)
      .map(result => new ProducerRecord[String, Array[Byte]](
        "processed-events", result.key, result.toByteArray))
      .runWith(Producer.plainSink(producerSettings))

  def parseEvent(bytes: Array[Byte]): Event =
    Event.parseFrom(bytes)

  def enrichEvent(event: Event): Future[EnrichedEvent] =
    for
      userInfo <- userService.getUser(event.userId)
      geoInfo <- geoService.lookup(event.ipAdddess)
    yield EnrichedEvent(event, userInfo, geoInfo)

  def batchProcess(events: Seq[EnrichedEvent]): Future[BatchResult] =
    analyticsService.processBatch(events)
```

---

## Event Sourcing with Akka

```scala
// UserAggregate.scala
import akka.actor.typed.*
import akka.actor.typed.scaladsl.*
import akka.persistence.typed.*
import akka.persistence.typed.scaladsl.*
import java.time.Instant

object UserAggregate:
  sealed trait Command
  case class CreateUser(name: String, email: String, replyTo: ActorRef[Response]) extends Command
  case class UpdateEmail(email: String, replyTo: ActorRef[Response]) extends Command
  case class Deactivate(replyTo: ActorRef[Response]) extends Command
  case class GetState(replyTo: ActorRef[Option[User]]) extends Command

  sealed trait Event
  case class UserCreated(id: String, name: String, email: String, at: Instant) extends Event
  case class EmailUpdated(email: String, at: Instant) extends Event
  case class UserDeactivated(at: Instant) extends Event

  sealed trait Response
  case class Success(user: User) extends Response
  case class Failure(reason: String) extends Response

  case class User(
    id: String,
    name: String,
    email: String,
    status: UserStatus,
    createdAt: Instant,
    updatedAt: Instant
  )

  enum UserStatus:
    case Active, Deactivated

  def apply(id: String): Behavior[Command] =
    EventSourcedBehavior[Command, Event, Option[User]](
      persistenceId = PersistenceId("User", id),
      emptyState = None,
      commandHandler = commandHandler(id),
      eventHandler = eventHandler
    ).withRetention(RetentionCriteria.snapshotEvery(100, 2))

  private def commandHandler(id: String)(state: Option[User], cmd: Command): Effect[Event, Option[User]] =
    state match
      case None => handleNew(id, cmd)
      case Some(user) if user.status == UserStatus.Active => handleActive(user, cmd)
      case Some(_) => handleDeactivated(cmd)

  private def handleNew(id: String, cmd: Command): Effect[Event, Option[User]] =
    cmd match
      case CreateUser(name, email, replyTo) =>
        val event = UserCreated(id, name, email, Instant.now)
        Effect
          .persist(event)
          .thenRun(state => replyTo ! Success(state.get))
      case other: Command =>
        other match
          case GetState(replyTo) => replyTo ! None
          case CreateUser(_, _, replyTo) => replyTo ! Failure("Unexpected")
          case UpdateEmail(_, replyTo) => replyTo ! Failure("User does not exist")
          case Deactivate(replyTo) => replyTo ! Failure("User does not exist")
        Effect.none

  private def handleActive(user: User, cmd: Command): Effect[Event, Option[User]] =
    cmd match
      case UpdateEmail(email, replyTo) =>
        Effect
          .persist(EmailUpdated(email, Instant.now))
          .thenRun(state => replyTo ! Success(state.get))
      case Deactivate(replyTo) =>
        Effect
          .persist(UserDeactivated(Instant.now))
          .thenRun(state => replyTo ! Success(state.get))
      case GetState(replyTo) =>
        replyTo ! Some(user)
        Effect.none
      case CreateUser(_, _, replyTo) =>
        replyTo ! Failure("User already exists")
        Effect.none

  private def handleDeactivated(cmd: Command): Effect[Event, Option[User]] =
    cmd match
      case GetState(replyTo) =>
        Effect.none.thenRun(state => replyTo ! state)
      case CreateUser(_, _, replyTo) =>
        replyTo ! Failure("User is deactivated")
        Effect.none
      case UpdateEmail(_, replyTo) =>
        replyTo ! Failure("User is deactivated")
        Effect.none
      case Deactivate(replyTo) =>
        replyTo ! Failure("User is deactivated")
        Effect.none

  private val eventHandler: (Option[User], Event) => Option[User] = (state, event) =>
    event match
      case UserCreated(id, name, email, at) =>
        Some(User(id, name, email, UserStatus.Active, at, at))
      case EmailUpdated(email, at) =>
        state.map(_.copy(email = email, updatedAt = at))
      case UserDeactivated(at) =>
        state.map(_.copy(status = UserStatus.Deactivated, updatedAt = at))
```

---

## ZIO Application

### Complete ZIO Service

```scala
// Main.scala
import zio.*
import zio.http.*
import zio.json.*

object Main extends ZIOAppDefault:
  def run =
    Server.serve(UserApp.routes)
      .provide(
        Server.default,
        UserServiceLive.layer,
        UserRepositoryLive.layer,
        Database.layer
      )

// UserApp.scala
object UserApp:
  val routes: Routes[UserService, Nothing] = Routes(
    Method.GET / "users" -> handler { (req: Request) =>
      for
        service <- ZIO.service[UserService]
        users <- service.findAll(0, 20)
      yield Response.json(users.toJson)
    },

    Method.GET / "users" / long("id") -> handler { (id: Long, req: Request) =>
      for
        service <- ZIO.service[UserService]
        user <- service.findById(id)
        response <- user match
          case Some(u) => ZIO.succeed(Response.json(u.toJson))
          case None => ZIO.succeed(Response.status(Status.NotFound))
      yield response
    },

    Method.POST / "users" -> handler { (req: Request) =>
      for
        body <- req.body.asString
        request <- ZIO.fromEither(body.fromJson[CreateUserRequest])
          .mapError(e => Response.text(e).status(Status.BadRequest))
        service <- ZIO.service[UserService]
        user <- service.create(request)
      yield Response.json(user.toJson).status(Status.Created)
    }
  )

// UserService.scala
trait UserService:
  def findAll(page: Int, size: Int): Task[List[User]]
  def findById(id: Long): Task[Option[User]]
  def create(request: CreateUserRequest): Task[User]

case class UserServiceLive(repository: UserRepository) extends UserService:
  def findAll(page: Int, size: Int): Task[List[User]] =
    repository.findAll(page * size, size)

  def findById(id: Long): Task[Option[User]] =
    repository.findById(id)

  def create(request: CreateUserRequest): Task[User] =
    for
      exists <- repository.existsByEmail(request.email)
      _ <- ZIO.fail(new Exception("Email exists")).when(exists)
      user = User(0, request.name, request.email, UserStatus.Pending)
      saved <- repository.save(user)
    yield saved

object UserServiceLive:
  val layer: ZLayer[UserRepository, Nothing, UserService] =
    ZLayer.fromFunction(UserServiceLive.apply)

// Models.scala
import zio.json.*

case class User(
  id: Long,
  name: String,
  email: String,
  status: UserStatus
) derives JsonEncoder, JsonDecoder

enum UserStatus derives JsonEncoder, JsonDecoder:
  case Pending, Active, Suspended

case class CreateUserRequest(
  name: String,
  email: String
) derives JsonDecoder
```

---

## Build Configuration

### Multi-Project SBT

```scala
// build.sbt
ThisBuild / scalaVersion := "3.4.2"
ThisBuild / organization := "com.example"
ThisBuild / version := "1.0.0"

lazy val commonSettings = Seq(
  scalacOptions ++= Seq(
    "-deprecation",
    "-feature",
    "-unchecked",
    "-Xfatal-warnings"
  )
)

lazy val root = (project in file("."))
  .aggregate(core, api, analytics)
  .settings(
    name := "scala-microservices"
  )

lazy val core = (project in file("core"))
  .settings(commonSettings)
  .settings(
    name := "core",
    libraryDependencies ++= Seq(
      "org.typelevel" %% "cats-effect" % "3.5.4",
      "io.circe" %% "circe-generic" % "0.15.0",
      "org.scalatest" %% "scalatest" % "3.2.18" % Test
    )
  )

lazy val api = (project in file("api"))
  .dependsOn(core)
  .settings(commonSettings)
  .settings(
    name := "api",
    libraryDependencies ++= Seq(
      "org.http4s" %% "http4s-ember-server" % "0.24.0",
      "org.http4s" %% "http4s-circe" % "0.24.0",
      "org.http4s" %% "http4s-dsl" % "0.24.0",
      "org.tpolecat" %% "doobie-core" % "1.0.0-RC4",
      "org.tpolecat" %% "doobie-postgres" % "1.0.0-RC4"
    )
  )

lazy val analytics = (project in file("analytics"))
  .dependsOn(core)
  .settings(commonSettings)
  .settings(
    name := "analytics",
    libraryDependencies ++= Seq(
      "org.apache.spark" %% "spark-sql" % "3.5.0" % Provided,
      "io.delta" %% "delta-spark" % "3.0.0"
    )
  )
```

---

Last Updated: 2025-12-07
Version: 1.0.0
