# JVM Languages Reference Guide

## Complete Language Coverage

### Java 21 LTS (Long-Term Support)

Version Information:
- Latest: 21.0.5 (LTS, support until September 2031)
- Key JEPs: 444 (Virtual Threads), 441 (Pattern Matching for switch), 440 (Record Patterns)
- JVM: HotSpot, GraalVM Native Image support

Core Features:

| Feature | Status | JEP | Description |
|---------|--------|-----|-------------|
| Virtual Threads | Final | 444 | Lightweight threads for high-concurrency |
| Pattern Matching for switch | Final | 441 | Type patterns with guards |
| Record Patterns | Final | 440 | Destructuring in pattern matching |
| Sealed Classes | Final | 409 | Restrict class hierarchies |
| Sequenced Collections | Final | 431 | Ordered collection interfaces |
| String Templates | Preview | 430 | Embedded expressions in strings |
| Structured Concurrency | Preview | 453 | Scope-based concurrency |
| Scoped Values | Preview | 446 | Immutable inherited values |

Enterprise Ecosystem:
- Spring Boot 3.3: Native GraalVM, Virtual Threads, Observability
- Jakarta EE 10: Cloud-native APIs, MicroProfile integration
- Hibernate 7: Java 21 features, improved performance
- Quarkus 3.8: Supersonic Subatomic Java, Kubernetes-native

### Kotlin 2.0

Version Information:
- Latest: 2.0.20 (November 2025)
- K2 Compiler: 2x faster compilation, better type inference
- Multiplatform: JVM, JS, Native, WASM targets

Core Features:

| Feature | Status | Description |
|---------|--------|-------------|
| K2 Compiler | Stable | New compiler frontend with improved performance |
| Context Receivers | Experimental | Multiple implicit receivers |
| Data Objects | Stable | Singleton data classes |
| Value Classes | Stable | Inline wrapper types |
| Sealed Interfaces | Stable | Restricted interface implementations |
| Explicit API Mode | Stable | Enforce visibility modifiers |

Coroutines Ecosystem:
- kotlinx.coroutines 1.9: Virtual thread integration, improved Flow
- Ktor 3.0: Async HTTP client/server, WebSocket support
- Exposed 0.55: Lightweight SQL framework with coroutines
- Arrow 2.0: Functional programming utilities

Android Ecosystem:
- Jetpack Compose 1.7: Declarative UI, Material 3
- Compose Multiplatform 1.7: Desktop, iOS, Web targets
- ViewModels + StateFlow: Reactive state management
- Room 2.7: Kotlin-first database, KSP support

### Scala 3.4

Version Information:
- Latest: 3.4.2 (November 2025)
- Dotty: New compiler with improved type system
- TASTy: Portable intermediate representation

Core Features:

| Feature | Status | Description |
|---------|--------|-------------|
| Export Clauses | Stable | Selective member export |
| Extension Methods | Stable | Type-safe extensions |
| Enum Types | Stable | Algebraic data types |
| Opaque Types | Stable | Zero-cost type abstractions |
| Union Types | Stable | A or B type unions |
| Intersection Types | Stable | A and B type combinations |
| Match Types | Stable | Type-level computation |
| Inline Methods | Stable | Compile-time evaluation |
| Given/Using | Stable | Context parameters |

Functional Ecosystem:
- Cats Effect 3.5: Pure FP runtime, fibers
- ZIO 2.1: Effect system, layers, streaming
- FS2 3.10: Functional streams
- Circe 0.15: JSON parsing
- Http4s 0.24: Functional HTTP

Big Data Ecosystem:
- Apache Spark 3.5: Scala 3 support, Delta Lake
- Apache Flink 1.19: Stream processing
- Apache Kafka 3.7: Event streaming
- Akka 2.9: Typed actors, streams, clustering

---

## Context7 Library Mappings

### Java Libraries

```
/spring-projects/spring-boot - Spring Boot framework
/spring-projects/spring-framework - Spring Core framework
/spring-projects/spring-security - Security framework
/spring-projects/spring-data-jpa - JPA repositories
/hibernate/hibernate-orm - Hibernate ORM
/junit-team/junit5 - JUnit 5 testing
/mockito/mockito - Mocking framework
/testcontainers/testcontainers-java - Container testing
/gradle/gradle - Build automation
/apache/maven - Project management
/resilience4j/resilience4j - Fault tolerance
/open-telemetry/opentelemetry-java - Observability
```

### Kotlin Libraries

```
/JetBrains/kotlin - Kotlin language
/Kotlin/kotlinx.coroutines - Coroutines library
/ktorio/ktor - Async framework
/JetBrains/Exposed - SQL framework
/arrow-kt/arrow - Functional programming
/google/ksp - Symbol processing
/Kotlin/kotlinx.serialization - Serialization
/cashapp/sqldelight - Type-safe SQL
```

### Scala Libraries

```
/scala/scala3 - Scala 3 language
/typelevel/cats-effect - Pure FP runtime
/zio/zio - Effect system
/akka/akka - Actor system
/http4s/http4s - Functional HTTP
/circe/circe - JSON library
/softwaremill/tapir - API first
/apache/spark - Big data
/slick/slick - Database access
```

---

## Testing Patterns

### JUnit 5 with Mockito (Complete)

```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock private UserRepository userRepository;
    @Mock private EmailService emailService;
    @InjectMocks private UserService userService;

    @Nested
    @DisplayName("createUser tests")
    class CreateUserTests {
        @Test
        @DisplayName("Should create user successfully")
        void shouldCreateUser() {
            // Arrange
            var request = new CreateUserRequest("John", "john@example.com");
            var savedUser = new User(1L, "John", "john@example.com");
            when(userRepository.save(any(User.class))).thenReturn(savedUser);
            doNothing().when(emailService).sendWelcomeEmail(anyString());

            // Act
            var result = userService.createUser(request);

            // Assert
            assertThat(result).isNotNull();
            assertThat(result.getId()).isEqualTo(1L);
            assertThat(result.getName()).isEqualTo("John");

            verify(userRepository).save(any(User.class));
            verify(emailService).sendWelcomeEmail("john@example.com");
        }

        @Test
        @DisplayName("Should throw exception for duplicate email")
        void shouldThrowForDuplicateEmail() {
            var request = new CreateUserRequest("John", "existing@example.com");
            when(userRepository.existsByEmail("existing@example.com")).thenReturn(true);

            assertThatThrownBy(() -> userService.createUser(request))
                .isInstanceOf(DuplicateEmailException.class)
                .hasMessageContaining("existing@example.com");

            verify(userRepository, never()).save(any());
        }
    }

    @ParameterizedTest
    @ValueSource(strings = {"", " ", "ab"})
    @DisplayName("Should reject invalid names")
    void shouldRejectInvalidNames(String name) {
        var request = new CreateUserRequest(name, "valid@example.com");

        assertThatThrownBy(() -> userService.createUser(request))
            .isInstanceOf(ValidationException.class);
    }
}
```

### TestContainers Integration

```java
@Testcontainers
@SpringBootTest
@Transactional
class UserRepositoryIntegrationTest {
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
        .withDatabaseName("testdb")
        .withUsername("test")
        .withPassword("test");

    @Container
    static GenericContainer<?> redis = new GenericContainer<>("redis:7-alpine")
        .withExposedPorts(6379);

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
        registry.add("spring.data.redis.host", redis::getHost);
        registry.add("spring.data.redis.port", () -> redis.getMappedPort(6379));
    }

    @Autowired
    private UserRepository userRepository;

    @Test
    void shouldPerformCRUDOperations() {
        // Create
        var user = new User(null, "John", "john@example.com");
        var saved = userRepository.save(user);
        assertThat(saved.getId()).isNotNull();

        // Read
        var found = userRepository.findById(saved.getId());
        assertThat(found).isPresent();
        assertThat(found.get().getName()).isEqualTo("John");

        // Update
        saved.setName("Jane");
        userRepository.save(saved);
        var updated = userRepository.findById(saved.getId());
        assertThat(updated.get().getName()).isEqualTo("Jane");

        // Delete
        userRepository.delete(saved);
        assertThat(userRepository.findById(saved.getId())).isEmpty();
    }

    @Test
    void shouldFindByEmailIgnoreCase() {
        userRepository.save(new User(null, "Test", "Test@Example.COM"));

        var result = userRepository.findByEmailIgnoreCase("test@example.com");

        assertThat(result).isPresent();
    }
}
```

### Kotlin Coroutines Testing

```kotlin
class UserServiceTest {
    private val userRepository = mockk<UserRepository>()
    private val orderRepository = mockk<OrderRepository>()
    private val userService = UserService(userRepository, orderRepository)

    @Test
    fun `should fetch user with orders concurrently`() = runTest {
        // Arrange
        val testUser = User(1L, "John", "john@example.com")
        val testOrders = listOf(
            Order(1L, 1L, BigDecimal("99.99")),
            Order(2L, 1L, BigDecimal("149.99"))
        )
        coEvery { userRepository.findById(1L) } coAnswers {
            delay(100) // Simulate network delay
            testUser
        }
        coEvery { orderRepository.findByUserId(1L) } coAnswers {
            delay(100)
            testOrders
        }

        // Act
        val startTime = testScheduler.currentTime
        val result = userService.fetchUserWithOrders(1L)
        val duration = testScheduler.currentTime - startTime

        // Assert
        assertThat(result.user).isEqualTo(testUser)
        assertThat(result.orders).containsExactlyElementsOf(testOrders)
        assertThat(duration).isLessThan(150) // Proves concurrent execution
    }

    @Test
    fun `should handle Flow emissions`() = runTest {
        // Arrange
        val users = listOf(
            User(1L, "John", "john@example.com"),
            User(2L, "Jane", "jane@example.com")
        )
        coEvery { userRepository.findAllAsFlow() } returns users.asFlow()

        // Act & Assert
        userService.streamUsers()
            .toList()
            .also { result ->
                assertThat(result).hasSize(2)
                assertThat(result.map { it.name }).containsExactly("John", "Jane")
            }
    }

    @Test
    fun `should cancel coroutine on timeout`() = runTest {
        coEvery { userRepository.findById(any()) } coAnswers {
            delay(5000) // Long operation
            mockk()
        }

        assertThrows<TimeoutCancellationException> {
            withTimeout(100) {
                userService.findById(1L)
            }
        }
    }
}
```

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

    "create user successfully" in {
      val request = CreateUserRequest("Jane", "jane@example.com")
      val createdUser = User(2L, "Jane", "jane@example.com")
      when(mockRepository.save(any[User])).thenReturn(createdUser)

      val actor = spawn(UserActor(mockRepository))
      val probe = createTestProbe[User]()

      actor ! CreateUser(request, probe.ref)

      probe.expectMessage(createdUser)
    }

    "handle multiple requests concurrently" in {
      val users = (1 to 100).map(i => User(i.toLong, s"User$i", s"user$i@example.com"))
      users.foreach(u => when(mockRepository.findById(u.id)).thenReturn(Some(u)))

      val actor = spawn(UserActor(mockRepository))
      val probes = users.map(_ => createTestProbe[Option[User]]())

      // Send all requests
      users.zip(probes).foreach { case (user, probe) =>
        actor ! GetUser(user.id, probe.ref)
      }

      // Verify all responses
      users.zip(probes).foreach { case (user, probe) =>
        probe.expectMessage(Some(user))
      }
    }
  }
```

### Cats Effect Testing

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

---

## Performance Characteristics

### JVM Startup and Memory

| Language/Runtime | Cold Start | Warm Start | Base Memory | With GraalVM Native |
|------------------|------------|------------|-------------|---------------------|
| Java 21 (HotSpot) | 2-5s | <100ms | 256MB+ | 50-100ms, 64MB |
| Kotlin (JVM) | 2-5s | <100ms | 256MB+ | 50-100ms, 64MB |
| Scala 3 (JVM) | 3-6s | <100ms | 512MB+ | Not recommended |

### Throughput Benchmarks

| Framework | Requests/sec | Latency P99 | Memory Usage |
|-----------|-------------|-------------|--------------|
| Spring Boot 3.3 (Virtual Threads) | 150K | 2ms | 512MB |
| Spring WebFlux | 180K | 1.5ms | 384MB |
| Ktor 3 (Netty) | 200K | 1ms | 256MB |
| Akka HTTP | 180K | 1.2ms | 384MB |
| Http4s (Blaze) | 160K | 1.5ms | 320MB |

### Compilation Times

| Build Tool | Clean Build | Incremental | With Cache |
|------------|------------|-------------|------------|
| Maven 3.9 (Java) | 30-60s | 5-10s | 15-30s |
| Gradle 8.5 (Java) | 20-40s | 3-8s | 10-20s |
| Gradle 8.5 (Kotlin) | 40-80s | 5-15s | 20-40s |
| SBT 1.10 (Scala) | 60-120s | 10-30s | 30-60s |

---

## Development Environment

### IDE Support

| IDE | Java | Kotlin | Scala | Best For |
|-----|------|--------|-------|----------|
| IntelliJ IDEA | Excellent | Excellent | Good | All JVM languages |
| VS Code | Good | Good | Fair | Lightweight development |
| Eclipse | Good | Fair | Fair | Enterprise Java |
| Neovim | Good | Good | Good | Terminal-based workflow |

### Recommended Plugins

IntelliJ IDEA:
- Kotlin (bundled with IntelliJ)
- Scala (by JetBrains)
- Spring Boot Assistant
- TestContainers (for integration tests)
- Key Promoter X (keyboard shortcuts)

VS Code:
- Extension Pack for Java
- Kotlin Language
- Scala (Metals)
- Spring Boot Extension Pack
- Test Runner for Java

### Linters and Formatters

| Language | Linter | Formatter | Config File |
|----------|--------|-----------|-------------|
| Java | Checkstyle, SpotBugs | google-java-format | checkstyle.xml |
| Kotlin | Detekt | ktlint | detekt.yml, .editorconfig |
| Scala | Scalastyle, WartRemover | Scalafmt | .scalafmt.conf |

---

## Container Optimization

### Docker Multi-Stage Builds

Java 21 Optimized:
```dockerfile
FROM eclipse-temurin:21-jdk-alpine AS builder
WORKDIR /app
COPY . .
RUN ./gradlew bootJar --no-daemon

FROM eclipse-temurin:21-jre-alpine
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app
WORKDIR /app
COPY --from=builder /app/build/libs/*.jar app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

GraalVM Native Image:
```dockerfile
FROM ghcr.io/graalvm/native-image-community:21 AS builder
WORKDIR /app
COPY . .
RUN ./gradlew nativeCompile

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/build/native/nativeCompile/app /app
ENTRYPOINT ["/app"]
```

### JVM Tuning for Containers

```yaml
# Kubernetes deployment with JVM tuning
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
          -XX:+OptimizeStringConcat
```

---

## Migration Guides

### Java 17 to 21

Key Changes:
1. Virtual Threads: Replace `Executors.newFixedThreadPool()` with `Executors.newVirtualThreadPerTaskExecutor()`
2. Pattern Matching: Refactor `instanceof` checks to pattern matching
3. Record Patterns: Use destructuring in switch expressions
4. Sequenced Collections: Use `SequencedCollection.getFirst()` instead of iterating

### Kotlin 1.9 to 2.0

Key Changes:
1. K2 Compiler: Add `kotlin.experimental.tryK2=true` to gradle.properties
2. Context Receivers: Migrate from extension functions where appropriate
3. Data Objects: Convert singleton data classes
4. Explicit API: Enable for library projects

### Scala 2.13 to 3.4

Key Changes:
1. Braceless Syntax: Optional significant indentation
2. Given/Using: Replace `implicit` with `given` and `using`
3. Extension Methods: Replace implicit classes with `extension`
4. Enums: Replace sealed traits with `enum`
5. Export Clauses: Replace trait mixing with exports

---

Last Updated: 2025-12-07
Version: 1.0.0
