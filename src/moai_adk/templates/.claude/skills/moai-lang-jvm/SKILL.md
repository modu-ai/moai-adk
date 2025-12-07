---
name: moai-lang-jvm
description: JVM languages specialist covering Java 21 LTS, Kotlin 2.0, and Scala 3.4 for enterprise applications, microservices, and big data. Use when developing Spring Boot services, Android apps, Akka actors, or Spark data pipelines.
version: 1.0.0
category: language
tags:
  - jvm
  - java
  - kotlin
  - scala
  - enterprise
  - microservices
  - big-data
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

JVM Languages Expert - Java 21 LTS, Kotlin 2.0, Scala 3.4 with enterprise patterns and Context7 integration.

Auto-Triggers: JVM language files (`.java`, `.kt`, `.kts`, `.scala`, `.sc`), build files (`pom.xml`, `build.gradle`, `build.gradle.kts`, `build.sbt`)

Core Capabilities:
- Java 21 LTS: Virtual threads, pattern matching, record patterns, sealed classes
- Kotlin 2.0: K2 compiler, coroutines, Compose Multiplatform, Ktor
- Scala 3.4: Export clauses, extension methods, Akka, Cats Effect, ZIO
- Enterprise: Spring Boot 3.3, Jakarta EE 10, Hibernate 7
- Big Data: Spark 3.5, Kafka Streams, Flink integration
- Testing: JUnit 5, Mockito, TestContainers, ScalaTest

---

## Implementation Guide (5 minutes)

### Java 21 LTS Features

Virtual Threads (Project Loom):
```java
// Lightweight concurrent programming
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    IntStream.range(0, 10_000).forEach(i ->
        executor.submit(() -> {
            Thread.sleep(Duration.ofSeconds(1));
            return i;
        })
    );
}

// Structured concurrency (preview)
try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
    Supplier<User> user = scope.fork(() -> fetchUser(userId));
    Supplier<List<Order>> orders = scope.fork(() -> fetchOrders(userId));
    scope.join().throwIfFailed();
    return new UserWithOrders(user.get(), orders.get());
}
```

Pattern Matching for Switch:
```java
// Exhaustive switch with pattern matching
String describe(Object obj) {
    return switch (obj) {
        case Integer i when i > 0 -> "positive integer: " + i;
        case Integer i -> "non-positive integer: " + i;
        case String s -> "string of length " + s.length();
        case List<?> list -> "list with " + list.size() + " elements";
        case null -> "null value";
        default -> "unknown type";
    };
}
```

Record Patterns:
```java
// Destructuring records in patterns
record Point(int x, int y) {}
record Rectangle(Point topLeft, Point bottomRight) {}

int area(Rectangle rect) {
    return switch (rect) {
        case Rectangle(Point(var x1, var y1), Point(var x2, var y2)) ->
            Math.abs((x2 - x1) * (y2 - y1));
    };
}
```

Sealed Classes:
```java
public sealed interface Shape
    permits Circle, Rectangle, Triangle {
    double area();
}

public record Circle(double radius) implements Shape {
    public double area() { return Math.PI * radius * radius; }
}
```

### Spring Boot 3.3

REST Controller:
```java
@RestController
@RequestMapping("/api/users")
public class UserController {
    private final UserService userService;

    @GetMapping("/{id}")
    public ResponseEntity<UserDto> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<UserDto> createUser(
            @Valid @RequestBody CreateUserRequest request) {
        UserDto user = userService.create(request);
        URI location = URI.create("/api/users/" + user.id());
        return ResponseEntity.created(location).body(user);
    }
}
```

WebFlux Reactive:
```java
@RestController
@RequestMapping("/api/reactive/users")
public class ReactiveUserController {
    private final ReactiveUserService userService;

    @GetMapping(produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    public Flux<UserDto> streamUsers() {
        return userService.findAll()
            .delayElements(Duration.ofMillis(100));
    }

    @GetMapping("/{id}")
    public Mono<ResponseEntity<UserDto>> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .defaultIfEmpty(ResponseEntity.notFound().build());
    }
}
```

Spring Security 6:
```java
@Configuration
@EnableWebSecurity
public class SecurityConfig {
    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        return http
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/api/public/**").permitAll()
                .requestMatchers("/api/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .oauth2ResourceServer(oauth2 -> oauth2.jwt(Customizer.withDefaults()))
            .sessionManagement(session ->
                session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .build();
    }
}
```

### Kotlin 2.0 Features

K2 Compiler and Coroutines:
```kotlin
// Suspend functions with structured concurrency
suspend fun fetchUserWithOrders(userId: Long): UserWithOrders = coroutineScope {
    val userDeferred = async { userRepository.findById(userId) }
    val ordersDeferred = async { orderRepository.findByUserId(userId) }
    UserWithOrders(userDeferred.await(), ordersDeferred.await())
}

// Flow for reactive streams
fun observeUsers(): Flow<User> = flow {
    while (true) {
        emit(userRepository.findLatest())
        delay(1000)
    }
}.flowOn(Dispatchers.IO)
```

Ktor 3 Server:
```kotlin
fun main() {
    embeddedServer(Netty, port = 8080) {
        install(ContentNegotiation) { json() }
        install(Authentication) {
            jwt("auth-jwt") {
                verifier(JwtConfig.verifier)
                validate { credential ->
                    if (credential.payload.audience.contains("api"))
                        JWTPrincipal(credential.payload)
                    else null
                }
            }
        }
        routing {
            route("/api/users") {
                get { call.respond(userService.findAll()) }
                get("/{id}") {
                    val id = call.parameters["id"]?.toLongOrNull()
                        ?: return@get call.respond(HttpStatusCode.BadRequest)
                    userService.findById(id)?.let { call.respond(it) }
                        ?: call.respond(HttpStatusCode.NotFound)
                }
                post {
                    val request = call.receive<CreateUserRequest>()
                    val user = userService.create(request)
                    call.respond(HttpStatusCode.Created, user)
                }
            }
        }
    }.start(wait = true)
}
```

Context Receivers (Experimental):
```kotlin
context(LoggerContext, TransactionContext)
fun processOrder(order: Order): OrderResult {
    logger.info("Processing order: ${order.id}")
    return transaction {
        val validated = validateOrder(order)
        orderRepository.save(validated)
    }
}
```

### Scala 3.4 Features

Extension Methods:
```scala
extension (s: String)
  def words: List[String] = s.split("\\s+").toList
  def truncate(maxLen: Int): String =
    if s.length <= maxLen then s else s.take(maxLen - 3) + "..."

val text = "Hello World Scala"
println(text.words)        // List(Hello, World, Scala)
println(text.truncate(10)) // Hello W...
```

Export Clauses:
```scala
class UserService(repo: UserRepository, validator: UserValidator):
  export repo.{findById, findAll, save}
  export validator.validate

  def createUser(request: CreateUserRequest): User =
    val validated = validate(request)
    save(User.from(validated))
```

Akka Typed Actors:
```scala
object UserActor:
  sealed trait Command
  case class GetUser(id: Long, replyTo: ActorRef[Option[User]]) extends Command
  case class CreateUser(request: CreateUserRequest, replyTo: ActorRef[User]) extends Command

  def apply(repository: UserRepository): Behavior[Command] =
    Behaviors.receiveMessage {
      case GetUser(id, replyTo) =>
        replyTo ! repository.findById(id)
        Behaviors.same
      case CreateUser(request, replyTo) =>
        replyTo ! repository.save(User.from(request))
        Behaviors.same
    }
```

Cats Effect and ZIO:
```scala
// Cats Effect 3
def fetchUser(id: Long): IO[User] =
  IO.fromFuture(IO(userRepository.findById(id)))
    .flatMap {
      case Some(user) => IO.pure(user)
      case None => IO.raiseError(UserNotFound(id))
    }

// ZIO 2
def fetchUserZIO(id: Long): ZIO[UserRepository, UserError, User] =
  for
    repo <- ZIO.service[UserRepository]
    user <- ZIO.fromOption(repo.findById(id))
               .orElseFail(UserNotFound(id))
  yield user
```

---

## Advanced Patterns

### Build Tool Configuration

Maven 3.9:
```xml
<project>
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.3.0</version>
    </parent>
    <properties>
        <java.version>21</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-webflux</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId>
        </dependency>
    </dependencies>
</project>
```

Gradle 8.5 (Kotlin DSL):
```kotlin
plugins {
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.4"
    kotlin("jvm") version "2.0.20"
    kotlin("plugin.spring") version "2.0.20"
}

java { toolchain { languageVersion = JavaLanguageVersion.of(21) } }

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-webflux")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("io.projectreactor:reactor-test")
}
```

SBT 1.10:
```scala
ThisBuild / scalaVersion := "3.4.0"
ThisBuild / organization := "com.example"

lazy val root = (project in file("."))
  .settings(
    name := "scala-service",
    libraryDependencies ++= Seq(
      "com.typesafe.akka" %% "akka-actor-typed" % "2.9.0",
      "com.typesafe.akka" %% "akka-stream" % "2.9.0",
      "org.typelevel" %% "cats-effect" % "3.5.4",
      "dev.zio" %% "zio" % "2.1.0",
      "org.scalatest" %% "scalatest" % "3.2.18" % Test
    )
  )
```

### Testing Quick Reference

JUnit 5 + Mockito (Java):
```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock private UserRepository repo;
    @InjectMocks private UserService service;

    @Test void shouldFindUser() {
        when(repo.findById(1L)).thenReturn(Optional.of(testUser));
        assertThat(service.findById(1L)).contains(testUser);
    }
}
```

For comprehensive testing patterns (TestContainers, Kotlin coroutines testing, ScalaTest with Akka), see [reference.md](reference.md#testing-patterns).

---

## Context7 Integration

Library mappings for latest documentation:
- `/spring-projects/spring-boot` - Spring Boot 3.3 documentation
- `/hibernate/hibernate-orm` - Hibernate 7 ORM patterns
- `/junit-team/junit5` - JUnit 5 testing framework
- `/JetBrains/kotlin` - Kotlin 2.0 language reference
- `/scala/scala3` - Scala 3.4 documentation

Usage:
```python
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/spring-projects/spring-boot",
    topic="virtual-threads webflux",
    page=1
)
```

---

## Language Selection Guide

Use Java 21 When:
- Building enterprise applications with large teams
- Requiring long-term support (LTS until 2031)
- Integrating with existing Java ecosystems
- Maximum IDE and tooling support needed

Use Kotlin 2.0 When:
- Developing Android applications
- Preferring concise, expressive syntax
- Building reactive services with coroutines
- Full Java interoperability required

Use Scala 3.4 When:
- Building big data pipelines with Spark
- Implementing complex functional patterns
- Using actor-based concurrency (Akka)
- Requiring advanced type-level programming

---

## Works Well With

- `moai-domain-backend` - REST API, GraphQL, microservices architecture
- `moai-domain-database` - JPA, Hibernate, R2DBC patterns
- `moai-quality-testing` - JUnit 5, Mockito, TestContainers integration
- `moai-infra-docker` - JVM container optimization
- `moai-infra-kubernetes` - JVM deployment and scaling
- `moai-context7-integration` - Latest framework documentation

---

## Troubleshooting

Java Issues:
- Version mismatch: `java -version`, check `JAVA_HOME`
- Compilation errors: `mvn clean compile -X` or `gradle build --info`
- Virtual thread issues: Ensure Java 21+ with `--enable-preview` if needed

Kotlin Issues:
- K2 compiler: Add `kotlin.experimental.tryK2=true` to gradle.properties
- Coroutine hangs: Check `runBlocking` usage in suspend contexts
- Null safety: Review `!!` operators and null checks

Scala Issues:
- Implicit resolution: `scalac -explain` for detailed errors
- Akka actor issues: Check actor hierarchy and supervision
- SBT slow: Enable parallel compilation, use `~compile` for watch mode

---

## Advanced Documentation

For comprehensive reference materials:
- [reference.md](reference.md) - Complete JVM language coverage, Context7 mappings, performance characteristics
- [examples.md](examples.md) - Production-ready code examples, Spring Boot, Ktor, Akka patterns

---

Last Updated: 2025-12-07
Status: Production Ready (v1.0.0)
