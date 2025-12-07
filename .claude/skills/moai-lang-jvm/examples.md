# JVM Languages Production Examples

## REST API Implementations

### Java 21 Spring Boot 3.3

Complete User Service:
```java
// UserController.java
@RestController
@RequestMapping("/api/v1/users")
@RequiredArgsConstructor
public class UserController {
    private final UserService userService;

    @GetMapping
    public ResponseEntity<Page<UserDto>> listUsers(
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size,
            @RequestParam(required = false) String search) {
        var pageable = PageRequest.of(page, size, Sort.by("createdAt").descending());
        var users = search != null
            ? userService.searchUsers(search, pageable)
            : userService.findAll(pageable);
        return ResponseEntity.ok(users.map(UserDto::from));
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserDto> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(UserDto::from)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<UserDto> createUser(
            @Valid @RequestBody CreateUserRequest request) {
        var user = userService.create(request);
        var location = URI.create("/api/v1/users/" + user.id());
        return ResponseEntity.created(location).body(UserDto.from(user));
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserDto> updateUser(
            @PathVariable Long id,
            @Valid @RequestBody UpdateUserRequest request) {
        return userService.update(id, request)
            .map(UserDto::from)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        return userService.delete(id)
            ? ResponseEntity.noContent().build()
            : ResponseEntity.notFound().build();
    }
}

// UserService.java
@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class UserService {
    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final EventPublisher eventPublisher;

    public Page<User> findAll(Pageable pageable) {
        return userRepository.findAll(pageable);
    }

    public Page<User> searchUsers(String query, Pageable pageable) {
        return userRepository.findByNameContainingIgnoreCaseOrEmailContainingIgnoreCase(
            query, query, pageable);
    }

    public Optional<User> findById(Long id) {
        return userRepository.findById(id);
    }

    @Transactional
    public User create(CreateUserRequest request) {
        if (userRepository.existsByEmail(request.email())) {
            throw new DuplicateEmailException(request.email());
        }

        var user = User.builder()
            .name(request.name())
            .email(request.email())
            .passwordHash(passwordEncoder.encode(request.password()))
            .status(UserStatus.PENDING)
            .build();

        var saved = userRepository.save(user);
        eventPublisher.publish(new UserCreatedEvent(saved.getId(), saved.getEmail()));
        return saved;
    }

    @Transactional
    public Optional<User> update(Long id, UpdateUserRequest request) {
        return userRepository.findById(id)
            .map(user -> {
                user.setName(request.name());
                if (request.email() != null && !request.email().equals(user.getEmail())) {
                    if (userRepository.existsByEmail(request.email())) {
                        throw new DuplicateEmailException(request.email());
                    }
                    user.setEmail(request.email());
                }
                return userRepository.save(user);
            });
    }

    @Transactional
    public boolean delete(Long id) {
        if (!userRepository.existsById(id)) {
            return false;
        }
        userRepository.deleteById(id);
        eventPublisher.publish(new UserDeletedEvent(id));
        return true;
    }
}

// User.java
@Entity
@Table(name = "users")
@Getter @Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class User extends BaseEntity {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false, unique = true)
    private String email;

    @Column(nullable = false)
    private String passwordHash;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private UserStatus status;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Order> orders = new ArrayList<>();
}

// DTOs as Records
public record UserDto(
    Long id,
    String name,
    String email,
    UserStatus status,
    Instant createdAt
) {
    public static UserDto from(User user) {
        return new UserDto(
            user.getId(),
            user.getName(),
            user.getEmail(),
            user.getStatus(),
            user.getCreatedAt()
        );
    }
}

public record CreateUserRequest(
    @NotBlank @Size(min = 2, max = 100) String name,
    @NotBlank @Email String email,
    @NotBlank @Size(min = 8) String password
) {}

public record UpdateUserRequest(
    @NotBlank @Size(min = 2, max = 100) String name,
    @Email String email
) {}
```

Virtual Threads Integration:
```java
// AsyncUserService.java with Virtual Threads
@Service
@RequiredArgsConstructor
public class AsyncUserService {
    private final UserRepository userRepository;
    private final OrderRepository orderRepository;
    private final NotificationService notificationService;

    public UserWithDetails fetchUserDetails(Long userId) throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
            Supplier<User> userTask = scope.fork(() ->
                userRepository.findById(userId)
                    .orElseThrow(() -> new UserNotFoundException(userId)));
            Supplier<List<Order>> ordersTask = scope.fork(() ->
                orderRepository.findByUserId(userId));
            Supplier<List<Notification>> notificationsTask = scope.fork(() ->
                notificationService.getUnreadNotifications(userId));

            scope.join().throwIfFailed();

            return new UserWithDetails(
                userTask.get(),
                ordersTask.get(),
                notificationsTask.get()
            );
        }
    }

    public void processUsersInParallel(List<Long> userIds) {
        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            var futures = userIds.stream()
                .map(id -> executor.submit(() -> processUser(id)))
                .toList();

            for (var future : futures) {
                try {
                    future.get();
                } catch (Exception e) {
                    log.error("Failed to process user", e);
                }
            }
        }
    }

    private void processUser(Long userId) {
        // CPU-bound or I/O-bound operation
        var user = userRepository.findById(userId).orElseThrow();
        notificationService.sendDailyDigest(user);
    }
}
```

### Kotlin Ktor 3.0

Complete User API:
```kotlin
// Application.kt
fun main() {
    embeddedServer(Netty, port = 8080, host = "0.0.0.0") {
        configureKoin()
        configureSecurity()
        configureRouting()
        configureContentNegotiation()
        configureStatusPages()
    }.start(wait = true)
}

fun Application.configureKoin() {
    install(Koin) {
        modules(appModule)
    }
}

val appModule = module {
    single<Database> { DatabaseFactory.create() }
    single<UserRepository> { UserRepositoryImpl(get()) }
    single<UserService> { UserServiceImpl(get()) }
}

fun Application.configureSecurity() {
    install(Authentication) {
        jwt("auth-jwt") {
            realm = "User API"
            verifier(JwtConfig.verifier)
            validate { credential ->
                if (credential.payload.audience.contains("api"))
                    JWTPrincipal(credential.payload)
                else null
            }
            challenge { _, _ ->
                call.respond(HttpStatusCode.Unauthorized, ErrorResponse("Token invalid"))
            }
        }
    }
}

fun Application.configureRouting() {
    val userService by inject<UserService>()

    routing {
        route("/api/v1") {
            // Public routes
            post("/auth/register") {
                val request = call.receive<CreateUserRequest>()
                val user = userService.create(request)
                call.respond(HttpStatusCode.Created, user.toDto())
            }

            post("/auth/login") {
                val request = call.receive<LoginRequest>()
                val token = userService.authenticate(request)
                call.respond(TokenResponse(token))
            }

            // Protected routes
            authenticate("auth-jwt") {
                route("/users") {
                    get {
                        val page = call.parameters["page"]?.toIntOrNull() ?: 0
                        val size = call.parameters["size"]?.toIntOrNull() ?: 20
                        val users = userService.findAll(page, size)
                        call.respond(users.map { it.toDto() })
                    }

                    get("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: return@get call.respond(HttpStatusCode.BadRequest)
                        userService.findById(id)?.let { call.respond(it.toDto()) }
                            ?: call.respond(HttpStatusCode.NotFound)
                    }

                    get("/me") {
                        val principal = call.principal<JWTPrincipal>()!!
                        val userId = principal.payload.getClaim("userId").asLong()
                        val user = userService.findById(userId)
                            ?: return@get call.respond(HttpStatusCode.NotFound)
                        call.respond(user.toDto())
                    }

                    put("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: return@put call.respond(HttpStatusCode.BadRequest)
                        val request = call.receive<UpdateUserRequest>()
                        userService.update(id, request)?.let { call.respond(it.toDto()) }
                            ?: call.respond(HttpStatusCode.NotFound)
                    }

                    delete("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: return@delete call.respond(HttpStatusCode.BadRequest)
                        if (userService.delete(id)) {
                            call.respond(HttpStatusCode.NoContent)
                        } else {
                            call.respond(HttpStatusCode.NotFound)
                        }
                    }
                }
            }
        }
    }
}

// UserService.kt
interface UserService {
    suspend fun findAll(page: Int, size: Int): List<User>
    suspend fun findById(id: Long): User?
    suspend fun create(request: CreateUserRequest): User
    suspend fun update(id: Long, request: UpdateUserRequest): User?
    suspend fun delete(id: Long): Boolean
    suspend fun authenticate(request: LoginRequest): String
}

class UserServiceImpl(
    private val repository: UserRepository
) : UserService {
    override suspend fun findAll(page: Int, size: Int): List<User> = coroutineScope {
        repository.findAll(page * size, size)
    }

    override suspend fun findById(id: Long): User? = coroutineScope {
        repository.findById(id)
    }

    override suspend fun create(request: CreateUserRequest): User = coroutineScope {
        if (repository.existsByEmail(request.email)) {
            throw DuplicateEmailException(request.email)
        }

        val user = User(
            id = 0,
            name = request.name,
            email = request.email,
            passwordHash = BCrypt.hashpw(request.password, BCrypt.gensalt()),
            status = UserStatus.PENDING,
            createdAt = Instant.now()
        )
        repository.save(user)
    }

    override suspend fun update(id: Long, request: UpdateUserRequest): User? = coroutineScope {
        val existing = repository.findById(id) ?: return@coroutineScope null

        if (request.email != null && request.email != existing.email) {
            if (repository.existsByEmail(request.email)) {
                throw DuplicateEmailException(request.email)
            }
        }

        val updated = existing.copy(
            name = request.name,
            email = request.email ?: existing.email
        )
        repository.update(updated)
    }

    override suspend fun delete(id: Long): Boolean = coroutineScope {
        repository.delete(id)
    }

    override suspend fun authenticate(request: LoginRequest): String = coroutineScope {
        val user = repository.findByEmail(request.email)
            ?: throw AuthenticationException("Invalid credentials")

        if (!BCrypt.checkpw(request.password, user.passwordHash)) {
            throw AuthenticationException("Invalid credentials")
        }

        JwtConfig.generateToken(user)
    }
}

// Models
@Serializable
data class User(
    val id: Long,
    val name: String,
    val email: String,
    val passwordHash: String,
    val status: UserStatus,
    @Serializable(with = InstantSerializer::class)
    val createdAt: Instant
) {
    fun toDto() = UserDto(id, name, email, status, createdAt)
}

@Serializable
data class UserDto(
    val id: Long,
    val name: String,
    val email: String,
    val status: UserStatus,
    @Serializable(with = InstantSerializer::class)
    val createdAt: Instant
)

@Serializable
data class CreateUserRequest(
    val name: String,
    val email: String,
    val password: String
)

@Serializable
data class UpdateUserRequest(
    val name: String,
    val email: String? = null
)

@Serializable
data class LoginRequest(
    val email: String,
    val password: String
)

@Serializable
data class TokenResponse(val token: String)

@Serializable
data class ErrorResponse(val message: String)

@Serializable
enum class UserStatus { PENDING, ACTIVE, SUSPENDED }
```

### Scala 3.4 Http4s

Functional HTTP Service:
```scala
// Main.scala
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
class UserRoutes(service: UserService[IO]) extends Http4sDsl[IO]:
  private val userDecoder = jsonOf[IO, CreateUserRequest]
  private val updateDecoder = jsonOf[IO, UpdateUserRequest]

  val routes: HttpRoutes[IO] = HttpRoutes.of[IO] {
    case GET -> Root / "users" :? PageParam(page) +& SizeParam(size) =>
      for
        users <- service.findAll(page.getOrElse(0), size.getOrElse(20))
        response <- Ok(users.asJson)
      yield response

    case GET -> Root / "users" / LongVar(id) =>
      service.findById(id).flatMap {
        case Some(user) => Ok(user.asJson)
        case None => NotFound()
      }

    case req @ POST -> Root / "users" =>
      for
        request <- req.as[CreateUserRequest]
        result <- service.create(request).attempt
        response <- result match
          case Right(user) => Created(user.asJson)
          case Left(_: DuplicateEmailException) => Conflict()
          case Left(e) => InternalServerError(e.getMessage)
      yield response

    case req @ PUT -> Root / "users" / LongVar(id) =>
      for
        request <- req.as[UpdateUserRequest]
        result <- service.update(id, request)
        response <- result match
          case Some(user) => Ok(user.asJson)
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

// UserRepository.scala
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

### Spark 3.5 with Scala 3

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

### Akka Streams

```scala
// StreamProcessing.scala
import akka.actor.typed.ActorSystem
import akka.stream.scaladsl.*
import akka.stream.alpakka.kafka.scaladsl.*
import akka.kafka.{ConsumerSettings, ProducerSettings}
import org.apache.kafka.common.serialization.*

object StreamProcessing:
  given system: ActorSystem[Nothing] = ActorSystem(Behaviors.empty, "stream-system")

  val consumerSettings = ConsumerSettings(system, new StringDeserializer, new ByteArrayDeserializer)
    .withBootstrapServers("localhost:9092")
    .withGroupId("processor-group")

  val producerSettings = ProducerSettings(system, new StringSerializer, new ByteArraySerializer)
    .withBootstrapServers("localhost:9092")

  def processEvents(): Unit =
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
      geoInfo <- geoService.lookup(event.ipAddress)
    yield EnrichedEvent(event, userInfo, geoInfo)

  def batchProcess(events: Seq[EnrichedEvent]): Future[BatchResult] =
    analyticsService.processBatch(events)
```

---

## Android Development (Kotlin)

### Jetpack Compose with ViewModel

```kotlin
// UserListScreen.kt
@Composable
fun UserListScreen(
    viewModel: UserListViewModel = hiltViewModel(),
    onUserClick: (Long) -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Users") },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { viewModel.onAddUser() }) {
                Icon(Icons.Default.Add, contentDescription = "Add User")
            }
        }
    ) { padding ->
        when (val state = uiState) {
            is UserListUiState.Loading -> {
                Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }
            is UserListUiState.Success -> {
                LazyColumn(
                    modifier = Modifier.padding(padding),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(state.users, key = { it.id }) { user ->
                        UserListItem(
                            user = user,
                            onClick = { onUserClick(user.id) },
                            onDelete = { viewModel.deleteUser(user.id) }
                        )
                    }
                }
            }
            is UserListUiState.Error -> {
                Column(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(state.message, color = MaterialTheme.colorScheme.error)
                    Button(onClick = { viewModel.retry() }) {
                        Text("Retry")
                    }
                }
            }
        }
    }
}

@Composable
fun UserListItem(
    user: User,
    onClick: () -> Unit,
    onDelete: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(4.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            AsyncImage(
                model = user.avatarUrl,
                contentDescription = user.name,
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
            )
            Spacer(Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(user.name, style = MaterialTheme.typography.titleMedium)
                Text(user.email, style = MaterialTheme.typography.bodySmall)
            }
            IconButton(onClick = onDelete) {
                Icon(Icons.Default.Delete, contentDescription = "Delete")
            }
        }
    }
}

// UserListViewModel.kt
@HiltViewModel
class UserListViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel() {
    private val _uiState = MutableStateFlow<UserListUiState>(UserListUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        loadUsers()
    }

    fun loadUsers() {
        viewModelScope.launch {
            _uiState.value = UserListUiState.Loading
            userRepository.getUsers()
                .catch { e -> _uiState.value = UserListUiState.Error(e.message ?: "Unknown error") }
                .collect { users -> _uiState.value = UserListUiState.Success(users) }
        }
    }

    fun refresh() = loadUsers()
    fun retry() = loadUsers()

    fun deleteUser(id: Long) {
        viewModelScope.launch {
            userRepository.deleteUser(id)
                .onSuccess { loadUsers() }
                .onFailure { e ->
                    _uiState.value = UserListUiState.Error("Failed to delete: ${e.message}")
                }
        }
    }

    fun onAddUser() {
        // Navigate to add user screen
    }
}

sealed interface UserListUiState {
    data object Loading : UserListUiState
    data class Success(val users: List<User>) : UserListUiState
    data class Error(val message: String) : UserListUiState
}

// UserRepository.kt
interface UserRepository {
    fun getUsers(): Flow<List<User>>
    suspend fun getUser(id: Long): Result<User>
    suspend fun deleteUser(id: Long): Result<Unit>
    suspend fun createUser(request: CreateUserRequest): Result<User>
}

class UserRepositoryImpl @Inject constructor(
    private val api: UserApi,
    private val dao: UserDao
) : UserRepository {
    override fun getUsers(): Flow<List<User>> =
        dao.getAllUsers()
            .onStart { refreshFromNetwork() }

    private suspend fun refreshFromNetwork() {
        try {
            val users = api.getUsers()
            dao.insertAll(users)
        } catch (e: Exception) {
            // Log error, use cached data
        }
    }

    override suspend fun getUser(id: Long): Result<User> = runCatching {
        dao.getUser(id) ?: api.getUser(id).also { dao.insert(it) }
    }

    override suspend fun deleteUser(id: Long): Result<Unit> = runCatching {
        api.deleteUser(id)
        dao.delete(id)
    }

    override suspend fun createUser(request: CreateUserRequest): Result<User> = runCatching {
        val user = api.createUser(request)
        dao.insert(user)
        user
    }
}
```

---

## Microservices Patterns

### Event Sourcing with Akka

```scala
// UserAggregate.scala
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
      case other =>
        other.replyTo ! Failure("User does not exist")
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
      case other =>
        other.replyTo ! Failure("User is deactivated")
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

## Build Configurations

### Gradle Multi-Project (Kotlin DSL)

```kotlin
// settings.gradle.kts
rootProject.name = "jvm-microservices"

include(
    "common",
    "user-service",
    "order-service",
    "notification-service",
    "api-gateway"
)

// build.gradle.kts (root)
plugins {
    kotlin("jvm") version "2.0.20" apply false
    kotlin("plugin.spring") version "2.0.20" apply false
    id("org.springframework.boot") version "3.3.0" apply false
    id("io.spring.dependency-management") version "1.1.4" apply false
}

allprojects {
    group = "com.example"
    version = "1.0.0"

    repositories {
        mavenCentral()
    }
}

subprojects {
    apply(plugin = "org.jetbrains.kotlin.jvm")
    apply(plugin = "org.jetbrains.kotlin.plugin.spring")

    dependencies {
        implementation(kotlin("stdlib"))
        testImplementation(kotlin("test"))
    }

    tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
        kotlinOptions {
            freeCompilerArgs = listOf("-Xjsr305=strict")
            jvmTarget = "21"
        }
    }

    tasks.withType<Test> {
        useJUnitPlatform()
    }
}

// user-service/build.gradle.kts
plugins {
    id("org.springframework.boot")
    id("io.spring.dependency-management")
}

dependencies {
    implementation(project(":common"))
    implementation("org.springframework.boot:spring-boot-starter-webflux")
    implementation("org.springframework.boot:spring-boot-starter-data-r2dbc")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("io.projectreactor.kotlin:reactor-kotlin-extensions")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")
    implementation("io.r2dbc:r2dbc-postgresql")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("io.projectreactor:reactor-test")
    testImplementation("org.testcontainers:postgresql")
    testImplementation("org.testcontainers:r2dbc")
}
```

---

Last Updated: 2025-12-07
Version: 1.0.0
