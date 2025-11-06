---
name: moai-lang-java
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Java best practices with Spring ecosystem, enterprise backend development, and JVM optimization for 2025"
keywords: [java, programming, backend, spring, enterprise, maven, gradle, testing]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Java Development Mastery

**Modern Java Development with 2025 Best Practices**

> Comprehensive Java development guidance covering enterprise backend applications, Spring ecosystem integration, and production-ready development using the latest JVM features and frameworks.

## What It Does

- **Backend API Development**: Spring Boot, Spring Framework applications with modern reactive patterns
- **Database Integration**: Spring Data JPA, Hibernate, JDBC with query optimization
- **API Development**: REST, GraphQL, and gRPC services with Spring Security
- **Testing & Quality**: JUnit 5, Mockito, TestContainers with integration tests
- **Performance Optimization**: JVM tuning, reactive programming, caching patterns
- **Production Deployment**: Docker, Kubernetes, CI/CD, monitoring with Spring Actuator
- **Security Best Practices**: Spring Security, input validation, dependency scanning
- **Code Quality**: Checkstyle, SpotBugs, pre-commit hooks, static analysis

## When to Use

### Perfect Scenarios
- **Building REST APIs and microservices**
- **Developing enterprise applications and services**
- **Creating database-driven applications with JPA/Hibernate**
- **Integrating with enterprise systems and messaging**
- **Automating backend workflows with Spring Batch**
- **Deploying scalable server applications on JVM**
- **Enterprise application development with Spring**

### Common Triggers
- "Create Java backend API"
- "Set up Spring Boot project"
- "Optimize Java performance"
- "Test Java application"
- "Deploy Spring Boot application"
- "Java best practices"
- "Spring Boot microservice"

## Tool Version Matrix (2025-11-06)

### Core Java
- **Java**: 23 (current) / 21 (LTS) / 17 (LTS)
- **Package Managers**: Maven 3.9.x, Gradle 8.8
- **Build Tools**: Maven Wrapper, Gradle Wrapper
- **JVM**: OpenJDK, Oracle JDK, Amazon Corretto

### Web Frameworks
- **Spring Boot**: 3.3.x - Production-ready applications
- **Spring Framework**: 6.2.x - Core application framework
- **Spring Security**: 6.4.x - Authentication and authorization
- **Spring Data**: 3.3.x - Database access layer
- **Micronaut**: 4.7.x - Reactive microservices framework
- **Quarkus**: 3.17.x - Cloud-native Java framework

### Testing Tools
- **JUnit**: 5.11.x - Primary testing framework
- **Mockito**: 5.15.x - Mocking framework
- **TestContainers**: 1.20.x - Integration testing with containers
- **AssertJ**: 3.26.x - Fluent assertions library
- **Jacoco**: 0.8.x - Code coverage analysis

### Code Quality
- **Checkstyle**: 10.20.x - Code style checking
- **SpotBugs**: 4.8.x - Static analysis
- **PMD**: 7.7.x - Code quality analysis
- **Error Prone**: 2.36.x - Google's error detection

### Database Tools
- **Hibernate**: 6.6.x - ORM framework
- **Flyway**: 10.20.x - Database migrations
- **Liquibase**: 4.29.x - Database schema management
- **H2**: 2.3.x - In-memory database for testing

## Ecosystem Overview

### Package Management

```bash
# Maven project creation
mvn archetype:generate -DgroupId=com.example -DartifactId=my-app -DarchetypeArtifactId=maven-archetype-quickstart -DinteractiveMode=false

# Gradle project creation (modern approach)
gradle init --type java-application --dsl kotlin --test-framework junit-jupiter --package com.example

# Dependency management
# Maven (pom.xml)
# Gradle (build.gradle.kts)
./mvnw spring-boot:run
./gradlew bootRun
```

### Project Structure (2025 Best Practice)

```
my-java-project/
├── build.gradle.kts          # Gradle build configuration (Kotlin DSL)
├── settings.gradle.kts       # Gradle settings
├── gradlew                   # Gradle wrapper
├── gradlew.bat               # Windows Gradle wrapper
├── src/
│   ├── main/
│   │   ├── java/com/example/
│   │   │   ├── Application.java      # Main application class
│   │   │   ├── controller/           # REST controllers
│   │   │   ├── service/              # Business logic
│   │   │   ├── repository/           # Data access layer
│   │   │   ├── model/                # Domain models
│   │   │   ├── config/               # Configuration classes
│   │   │   └── exception/            # Custom exceptions
│   │   └── resources/
│   │       ├── application.yml       # Application configuration
│   │       ├── application-dev.yml   # Development profile
│   │       └── application-prod.yml  # Production profile
│   └── test/
│       └── java/com/example/
│           ├── unit/                 # Unit tests
│           ├── integration/          # Integration tests
│           └── testutil/             # Test utilities
├── .github/
│   └── workflows/                    # CI/CD pipelines
├── Dockerfile
├── docker-compose.yml                # Development environment
├── README.md
└── .gitignore
```

## Modern Development Patterns

### Modern Java Features (Java 21+)

```java
import java.util.List;
import java.util.Optional;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Stream;
import java.util.function.Predicate;
import java.util.record.Record;

// Records (immutable data carriers)
public record User(Long id, String username, String email) {
    // Compact constructor for validation
    public User {
        if (username == null || username.isBlank()) {
            throw new IllegalArgumentException("Username cannot be null or blank");
        }
    }
}

// Pattern Matching for instanceof
public void processUser(Object obj) {
    if (obj instanceof User(String username, String email)) {
        System.out.println("User: " + username + ", Email: " + email);
    }
}

// Sealed classes for domain modeling
public sealed interface Payment permits CreditCardPayment, PayPalPayment {
    BigDecimal getAmount();
}

public final record CreditCardPayment(BigDecimal amount, String cardNumber) implements Payment {}
public final record PayPalPayment(BigDecimal amount, String email) implements Payment {}

// Switch expressions with pattern matching
public String processPayment(Payment payment) {
    return switch (payment) {
        case CreditCardPayment(var amount, var card) -> 
            "Processing credit card payment of " + amount;
        case PayPalPayment(var amount, var email) -> 
            "Processing PayPal payment of " + amount + " for " + email;
    };
}

// Virtual Threads (Project Loom)
import java.util.concurrent.Executors;

public class VirtualThreadExample {
    public static void main(String[] args) {
        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            // Create thousands of virtual threads
            List<CompletableFuture<String>> futures = IntStream.range(0, 10_000)
                .mapToObj(i -> CompletableFuture.supplyAsync(() -> {
                    // Simulate I/O-bound operation
                    try {
                        Thread.sleep(100);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                    return "Task " + i + " completed";
                }, executor))
                .toList();
            
            // Wait for all to complete
            CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).join();
        }
    }
}

// Scoped Values (Structured concurrency)
import java.util.concurrent.ScopedValue;

public final class Context {
    private static final ScopedValue<String> USER_ID = ScopedValue.newInstance();
    
    public static void main(String[] args) {
        // Bind value in a structured scope
        ScopedValue.where(USER_ID, "user123").run(() -> {
            // All code in this scope can access USER_ID.get()
            processRequest();
            handleData();
        });
    }
}
```

### Spring Boot Reactive Patterns

```java
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.reactive.function.server.RouterFunction;
import org.springframework.web.reactive.function.server.ServerResponse;
import org.springframework.web.reactive.function.server.RouterFunctions;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import org.springframework.web.reactive.function.server.ServerRequest;

@SpringBootApplication
public class ReactiveApplication {
    
    @Bean
    public RouterFunction<ServerResponse> userRoutes(UserHandler userHandler) {
        return RouterFunctions.route()
            .GET("/api/users", userHandler::getAllUsers)
            .GET("/api/users/{id}", userHandler::getUserById)
            .POST("/api/users", userHandler::createUser)
            .PUT("/api/users/{id}", userHandler::updateUser)
            .DELETE("/api/users/{id}", userHandler::deleteUser)
            .build();
    }
    
    public static void main(String[] args) {
        SpringApplication.run(ReactiveApplication.class, args);
    }
}

// Functional endpoint handler
@Component
public class UserHandler {
    private final UserRepository userRepository;
    
    public UserHandler(UserRepository userRepository) {
        this.userRepository = userRepository;
    }
    
    public Mono<ServerResponse> getAllUsers(ServerRequest request) {
        Flux<User> users = userRepository.findAll();
        return ServerResponse.ok().body(users, User.class);
    }
    
    public Mono<ServerResponse> getUserById(ServerRequest request) {
        String id = request.pathVariable("id");
        return userRepository.findById(Long.valueOf(id))
            .flatMap(user -> ServerResponse.ok().bodyValue(user))
            .switchIfEmpty(ServerResponse.notFound().build());
    }
    
    public Mono<ServerResponse> createUser(ServerRequest request) {
        Mono<User> userMono = request.bodyToMono(User.class);
        return userMono.flatMap(user -> 
            userRepository.save(user)
                .flatMap(savedUser -> ServerResponse.created(URI.create("/api/users/" + savedUser.id()))
                    .bodyValue(savedUser))
        );
    }
}

// Reactive repository
@Repository
public interface ReactiveUserRepository extends ReactiveCrudRepository<User, Long> {
    Flux<User> findByUsername(String username);
    Mono<User> findByEmail(String email);
    @Query("SELECT * FROM users WHERE active = true")
    Flux<User> findActiveUsers();
}
```

### Modern Testing Patterns

```java
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;
import org.junit.jupiter.params.provider.CsvSource;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.testcontainers.containers.PostgreSQLContainer;
import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import static org.assertj.core.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
@DisplayName("UserService Tests")
class UserServiceTest {
    
    @Mock
    private UserRepository userRepository;
    
    @Mock
    private EmailService emailService;
    
    private UserService userService;
    
    @BeforeEach
    void setUp() {
        userService = new UserService(userRepository, emailService);
    }
    
    @Nested
    @DisplayName("Create User Tests")
    class CreateUserTests {
        
        @Test
        @DisplayName("Should create user successfully with valid data")
        void shouldCreateUserSuccessfully() {
            // Given
            UserCreateDto dto = new UserCreateDto("testuser", "test@example.com");
            User expectedUser = new User(1L, "testuser", "test@example.com");
            
            when(userRepository.save(any(User.class))).thenReturn(Mono.just(expectedUser));
            when(userRepository.existsByUsername("testuser")).thenReturn(Mono.just(false));
            when(userRepository.existsByEmail("test@example.com")).thenReturn(Mono.just(false));
            
            // When
            Mono<User> result = userService.createUser(dto);
            
            // Then
            StepVerifier.create(result)
                .assertNext(user -> {
                    assertThat(user.id()).isEqualTo(1L);
                    assertThat(user.username()).isEqualTo("testuser");
                    assertThat(user.email()).isEqualTo("test@example.com");
                })
                .verifyComplete();
            
            verify(emailService).sendWelcomeEmail(expectedUser);
        }
        
        @ParameterizedTest
        @ValueSource(strings = {"", "ab", "a".repeat(51)})
        @DisplayName("Should reject invalid usernames")
        void shouldRejectInvalidUsernames(String invalidUsername) {
            // Given
            UserCreateDto dto = new UserCreateDto(invalidUsername, "test@example.com");
            
            // When & Then
            assertThatThrownBy(() -> userService.createUser(dto))
                .isInstanceOf(ValidationException.class)
                .hasMessageContaining("Invalid username");
        }
    }
}

// Integration testing with TestContainers
@Testcontainers
@SpringBootTest
class UserRepositoryIntegrationTest {
    
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
            .withDatabaseName("testdb")
            .withUsername("test")
            .withPassword("test");
    
    @DynamicPropertySource
    static void postgresProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }
    
    @Autowired
    private UserRepository userRepository;
    
    @Test
    void shouldSaveAndRetrieveUser() {
        // Given
        User user = new User(null, "testuser", "test@example.com");
        
        // When
        User saved = userRepository.save(user);
        
        // Then
        assertThat(saved.id()).isNotNull();
        
        Optional<User> retrieved = userRepository.findById(saved.id());
        assertThat(retrieved).isPresent();
        assertThat(retrieved.get().username()).isEqualTo("testuser");
    }
}
```

## Performance Considerations

### JVM Performance Tuning

```java
// Efficient collections usage
public class CollectionOptimization {
    
    // Use appropriate collection types
    private final Map<String, User> userCache = new ConcurrentHashMap<>();  // Thread-safe
    private final List<String> recentItems = new CopyOnWriteArrayList<>();   // Read-heavy
    private final Queue<Task> taskQueue = new ConcurrentLinkedQueue<>();     // Lock-free
    
    // Use primitive collections for performance-critical code
    private final Int2ObjectHashMap<User> userById = new Int2ObjectHashMap<>();
    
    // Lazy initialization
    private volatile ExpensiveResource resource;
    
    public ExpensiveResource getResource() {
        ExpensiveResource result = resource;
        if (result == null) {
            synchronized (this) {
                result = resource;
                if (result == null) {
                    resource = result = new ExpensiveResource();
                }
            }
        }
        return result;
    }
}

// Memory-efficient data processing
public class StreamOptimization {
    
    public List<UserDto> processUsersLargeDataset() {
        try (Stream<User> userStream = Files.lines(Paths.get("users.csv"))
                .skip(1) // Skip header
                .parallel() // Parallel processing for large files
                .map(this::parseUser)
                .filter(Objects::nonNull)
                .filter(user -> user.isActive())) {
            
            return userStream
                    .map(this::convertToDto)
                    .collect(Collectors.toList());
        } catch (IOException e) {
            throw new RuntimeException("Failed to process users", e);
        }
    }
    
    // Use primitive streams for numeric operations
    public IntSummaryStatistics getAgeStatistics(List<User> users) {
        return users.stream()
                .mapToInt(User::age)
                .summaryStatistics();
    }
}

// Connection pooling and resource management
@Configuration
public class DatabaseConfig {
    
    @Bean
    @Primary
    public DataSource dataSource() {
        HikariConfig config = new HikariConfig();
        config.setJdbcUrl("jdbc:postgresql://localhost:5432/mydb");
        config.setUsername("user");
        config.setPassword("password");
        config.setMaximumPoolSize(20);
        config.setMinimumIdle(5);
        config.setConnectionTimeout(30000);
        config.setIdleTimeout(600000);
        config.setMaxLifetime(1800000);
        config.setLeakDetectionThreshold(60000);
        return new HikariDataSource(config);
    }
}
```

### Reactive Programming Performance

```java
// Backpressure handling
@Service
public class DataProcessingService {
    
    public Flux<ProcessedData> processDataWithBackpressure(Flux<RawData> rawData) {
        return rawData
                .onBackpressureBuffer(1000, // Buffer up to 1000 items
                    rawData::log, // Log dropped items
                    BufferOverflowStrategy.DROP_OLDEST)
                .publishOn(Schedulers.boundedElastic(), 256) // Limit concurrency
                .flatMap(this::processItem, 10) // Process 10 items in parallel
                .doOnError(error -> log.error("Processing error", error))
                .retryWhen(Retry.backoff(3, Duration.ofSeconds(1))
                    .maxBackoff(Duration.ofSeconds(10))
                    .jitter(0.5))
                .timeout(Duration.ofMinutes(5));
    }
    
    // Efficient batching
    public Flux<BatchResult> processInBatches(Flux<DataItem> items, int batchSize) {
        return items
                .window(batchSize)
                .flatMap(window -> window.collectList())
                .flatMap(this::processBatch);
    }
}

// Caching strategies
@Service
public class CacheService {
    
    private final Cache<String, User> userCache = Caffeine.newBuilder()
            .maximumSize(10_000)
            .expireAfterWrite(Duration.ofMinutes(30))
            .refreshAfterWrite(Duration.ofMinutes(10))
            .recordStats()
            .build();
    
    @Cacheable(value = "users", key = "#id")
    public User getUserById(Long id) {
        return userCache.get(id.toString(), key -> userRepository.findById(id)
                .orElseThrow(() -> new UserNotFoundException(id)));
    }
    
    // Reactive cache with Spring Boot
    @Bean
    public ReactiveCacheManager cacheManager() {
        return new ReactiveCaffeineCacheManager(
            Caffeine.newBuilder()
                    .maximumSize(1000)
                    .expireAfterWrite(Duration.ofMinutes(10))
                    .build()
        );
    }
}
```

### Profiling and Monitoring

```bash
# JVM monitoring with jcmd
jcmd <pid> VM.flags
jcmd <pid> GC.heap_info
jcmd <pid> Thread.print
jcmd <pid> GC.class_histogram

# Java Flight Recorder
java -XX:StartFlightRecording=duration=60s,filename=myrecording.jfr -jar app.jar

# JDK Mission Control (visual monitoring)
jmc

# Memory analysis with jmap and jhat
jmap -dump:format=b,file=heapdump.hprof <pid>
jhat heapdump.hprof

# Performance profiling with async-profiler
./profiler.sh -d 30 -f profile.svg <pid>

# Thread dump analysis
jstack <pid> > thread_dump.txt
# or with jcmd
jcmd <pid> Thread.print > thread_dump.txt
```

## Testing Strategy

### Gradle Configuration (build.gradle.kts)

```kotlin
plugins {
    application
    jacoco
    checkstyle
    pmd
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.5"
}

group = "com.example"
version = "0.1.0"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

repositories {
    mavenCentral()
}

dependencies {
    // Spring Boot starters
    implementation("org.springframework.boot:spring-boot-starter-webflux")
    implementation("org.springframework.boot:spring-boot-starter-data-r2dbc")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    
    // Database
    implementation("org.postgresql:r2dbc-postgresql")
    implementation("org.flywaydb:flyway-core")
    implementation("org.springframework:spring-jdbc")
    
    // Utility libraries
    implementation("com.fasterxml.jackson.core:jackson-databind")
    implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310")
    implementation("org.apache.commons:commons-lang3")
    
    // Development tools
    compileOnly("org.projectlombok:lombok")
    annotationProcessor("org.projectlombok:lombok")
    
    // Testing
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testImplementation("io.projectreactor:reactor-test")
    testImplementation("org.testcontainers:junit-jupiter")
    testImplementation("org.testcontainers:postgresql")
    testImplementation("org.testcontainers:r2dbc")
    
    // Code quality
    checkstyle("com.github.sevntu-checkstyle:sevntu-checks:1.44.1")
}

tasks.test {
    useJUnitPlatform()
    
    testLogging {
        events("passed", "skipped", "failed")
        showExceptions = true
        showCauses = true
        showStackTraces = true
    }
    
    // Parallel test execution
    maxParallelForks = (Runtime.getRuntime().availableProcessors() / 2).takeIf { it > 0 } ?: 1
}

// JaCoCo configuration for code coverage
jacoco {
    toolVersion = "0.8.12"
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required.set(true)
        html.required.set(true)
        csv.required.set(false)
    }
}

tasks.jacocoTestCoverageVerification {
    violationRules {
        rule {
            limit {
                minimum = "0.80".toBigDecimal() // 80% coverage minimum
            }
        }
    }
}

// Checkstyle configuration
tasks.checkstyleMain {
    source = fileTree("src/main/java")
}

tasks.checkstyleTest {
    source = fileTree("src/test/java")
}

// PMD configuration
pmd {
    ruleSets = listOf(
        "category/java/bestpractices.xml",
        "category/java/codesize.xml",
        "category/java/design.xml",
        "category/java/performance.xml",
        "category/java/security.xml"
    )
}
```

### Modern Testing Patterns

```java
// Test slices for focused testing
@WebFluxTest(UserController.class)
@Import({UserService.class, SecurityConfig.class})
class UserControllerTest {
    
    @Autowired
    private WebTestClient webTestClient;
    
    @MockBean
    private UserService userService;
    
    @Test
    void shouldReturnUserWhenFound() {
        // Given
        User user = new User(1L, "testuser", "test@example.com");
        when(userService.getUserById(1L)).thenReturn(Mono.just(user));
        
        // When & Then
        webTestClient.get()
                .uri("/api/users/1")
                .accept(MediaType.APPLICATION_JSON)
                .exchange()
                .expectStatus().isOk()
                .expectBody()
                .jsonPath("$.id").isEqualTo(1)
                .jsonPath("$.username").isEqualTo("testuser");
    }
}

// Property-based testing with jqwik
@Property
void stringReversalIsInverse(@ForAll("validStrings") String str) {
    String reversed = new StringBuilder(str).reverse().toString();
    String doubleReversed = new StringBuilder(reversed).reverse().toString();
    assertThat(doubleReversed).isEqualTo(str);
}

@Provide
Arbitrary<String> validStrings() {
    return Arbitraries.strings()
            .withChars("abc")
            .ofMinLength(1)
            .ofMaxLength(100);
}
```

## Security Best Practices

### Input Validation and Sanitization

```java
import jakarta.validation.constraints.*;
import org.hibernate.validator.constraints.URL;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

// DTO with validation annotations
public record UserCreateDto(
    
    @NotBlank(message = "Username is required")
    @Size(min = 3, max = 50, message = "Username must be between 3 and 50 characters")
    @Pattern(regexp = "^[a-zA-Z0-9_]+$", message = "Username can only contain letters, numbers, and underscores")
    String username,
    
    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    @Size(max = 100, message = "Email must be less than 100 characters")
    String email,
    
    @NotBlank(message = "Password is required")
    @Size(min = 8, max = 128, message = "Password must be between 8 and 128 characters")
    @Pattern(regexp = "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]", 
             message = "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character")
    String password
) {
    
    // Custom validation method
    public UserCreateDto {
        if (email != null && email.toLowerCase().contains("@admin.com")) {
            throw new IllegalArgumentException("Admin email domains are not allowed");
        }
    }
}

// Password encryption service
@Service
public class PasswordService {
    
    private final PasswordEncoder passwordEncoder = new BCryptPasswordEncoder(12);
    
    public String encodePassword(String rawPassword) {
        return passwordEncoder.encode(rawPassword);
    }
    
    public boolean matches(String rawPassword, String encodedPassword) {
        return passwordEncoder.matches(rawPassword, encodedPassword);
    }
}

// Input sanitization for XSS prevention
@Component
public class InputSanitizer {
    
    private static final org.owasp.encoder.Encoder encoder = org.owasp.encoder.Encoder.getInstance();
    
    public String sanitizeForHtml(String input) {
        return input == null ? null : encoder.encodeForHTML(input);
    }
    
    public String sanitizeForJavaScript(String input) {
        return input == null ? null : encoder.encodeForJavaScript(input);
    }
}
```

### Authentication and Authorization

```java
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.reactive.EnableWebFluxSecurity;
import org.springframework.security.config.web.server.ServerHttpSecurity;
import org.springframework.security.web.server.SecurityWebFilterChain;
import org.springframework.security.web.server.authentication.AuthenticationWebFilter;

@Configuration
@EnableWebFluxSecurity
public class SecurityConfig {
    
    @Bean
    public SecurityWebFilterChain securityWebFilterChain(ServerHttpSecurity http) {
        return http
                .csrf(ServerHttpSecurity.CsrfSpec::disable)
                .authorizeExchange(exchanges -> exchanges
                        .pathMatchers("/api/auth/**").permitAll()
                        .pathMatchers("/api/public/**").permitAll()
                        .pathMatchers("/actuator/health").permitAll()
                        .pathMatchers("/api/admin/**").hasRole("ADMIN")
                        .pathMatchers("/api/users/**").hasAnyRole("USER", "ADMIN")
                        .anyExchange().authenticated()
                )
                .oauth2ResourceServer(oauth2 -> oauth2
                        .jwt(jwt -> jwt.jwtDecoder(jwtDecoder()))
                )
                .addFilterAt(jwtAuthenticationFilter(), AuthenticationWebFilter.class)
                .build();
    }
    
    @Bean
    public JwtDecoder jwtDecoder() {
        return NimbusJwtDecoder.withJwkSetUri("https://your-auth-server/.well-known/jwks.json")
                .build();
    }
    
    @Bean
    public ReactiveAuthenticationManager authenticationManager() {
        return new UserRepositoryReactiveAuthenticationManager(userRepository, passwordService);
    }
}

// Method-level security
@Service
public class DocumentService {
    
    @PreAuthorize("hasRole('USER') and @documentSecurity.canRead(#documentId, authentication.name)")
    public Mono<Document> getDocument(Long documentId) {
        return documentRepository.findById(documentId);
    }
    
    @PreAuthorize("hasRole('ADMIN') or @documentSecurity.isOwner(#documentId, authentication.name)")
    public Mono<Document> updateDocument(Long documentId, DocumentUpdateDto updateDto) {
        return documentRepository.findById(documentId)
                .filter(document -> canUpdate(document, AuthenticationHolder.getCurrentUsername()))
                .flatMap(document -> {
                    document.update(updateDto);
                    return documentRepository.save(document);
                });
    }
    
    @PostAuthorize("returnObject.isPresent() && returnObject.get().isPublic()")
    public Optional<Document> getPublicDocument(Long documentId) {
        return documentRepository.findById(documentId);
    }
}
```

### Dependency Security Scanning

```bash
# Add to build.gradle.kts
plugins {
    id("org.owasp.dependencycheck") version "10.0.3"
    id("com.github.ben-manes.versions") version "0.51.0"
}

// OWASP Dependency Check
dependencyCheck {
    failBuildOnCVSS = 7.0f
    suppressionFile = "dependency-check-suppressions.xml"
    analyzers {
        experimentalEnabled = false
    }
}

// Dependency version updates
tasks.named("dependencyUpdates") {
    checkForGradleUpdate = true
    outputFormatter = "json"
    outputDir = "build/dependencyUpdates"
    reportfileName = "report"
}
```

## Integration Patterns

### Database Integration with R2DBC

```java
import org.springframework.data.annotation.Id;
import org.springframework.data.relational.core.mapping.Table;
import org.springframework.data.repository.reactive.ReactiveCrudRepository;
import org.springframework.r2dbc.core.DatabaseClient;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

@Table("users")
public record User(
    @Id Long id,
    String username,
    String email,
    Instant createdAt,
    boolean active
) {
    public User {
        if (createdAt == null) {
            createdAt = Instant.now();
        }
    }
}

@Repository
public interface UserRepository extends ReactiveCrudRepository<User, Long> {
    Flux<User> findByUsername(String username);
    Flux<User> findByEmail(String email);
    Flux<User> findByActiveTrue();
    
    @Query("SELECT * FROM users WHERE username LIKE :pattern")
    Flux<User> findByUsernamePattern(String pattern);
}

// Custom repository with complex queries
@Repository
public class CustomUserRepository {
    
    private final DatabaseClient databaseClient;
    
    public CustomUserRepository(DatabaseClient databaseClient) {
        this.databaseClient = databaseClient;
    }
    
    public Flux<UserSummary> findUserSummaries() {
        return databaseClient.sql("SELECT id, username, email FROM users WHERE active = true")
                .map((row, metadata) -> new UserSummary(
                    row.get("id", Long.class),
                    row.get("username", String.class),
                    row.get("email", String.class)
                ))
                .all();
    }
    
    public Mono<Integer> bulkUpdateStatus(List<Long> userIds, boolean active) {
        String placeholders = userIds.stream()
                .map(id -> "?")
                .collect(Collectors.joining(","));
        
        return databaseClient.sql(
                "UPDATE users SET active = :active WHERE id IN (" + placeholders + ")"
            )
            .bind("active", active)
            .bind(0, userIds.get(0))
            // ... bind remaining parameters
            .fetch()
            .rowsUpdated();
    }
}
```

### Message Queue Integration with RabbitMQ

```java
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.amqp.support.converter.Jackson2JsonMessageConverter;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.amqp.core.*;

@Configuration
public class RabbitMQConfig {
    
    public static final String USER_QUEUE = "user.queue";
    public static final String USER_EXCHANGE = "user.exchange";
    public static final String USER_ROUTING_KEY = "user.created";
    
    @Bean
    public Queue userQueue() {
        return QueueBuilder.durable(USER_QUEUE).build();
    }
    
    @Bean
    public TopicExchange userExchange() {
        return new TopicExchange(USER_EXCHANGE);
    }
    
    @Bean
    public Binding userBinding() {
        return BindingBuilder.bind(userQueue())
                .to(userExchange())
                .with(USER_ROUTING_KEY);
    }
    
    @Bean
    public Jackson2JsonMessageConverter messageConverter() {
        return new Jackson2JsonMessageConverter();
    }
    
    @Bean
    public RabbitTemplate rabbitTemplate(ConnectionFactory connectionFactory) {
        RabbitTemplate rabbitTemplate = new RabbitTemplate(connectionFactory);
        rabbitTemplate.setMessageConverter(messageConverter());
        return rabbitTemplate;
    }
}

@Service
public class UserEventPublisher {
    
    private final RabbitTemplate rabbitTemplate;
    
    public UserEventPublisher(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }
    
    public void publishUserCreated(User user) {
        UserCreatedEvent event = new UserCreatedEvent(user.id(), user.username(), user.email());
        rabbitTemplate.convertAndSend(
            RabbitMQConfig.USER_EXCHANGE,
            RabbitMQConfig.USER_ROUTING_KEY,
            event
        );
    }
}

@Component
public class UserEventConsumer {
    
    private final EmailService emailService;
    private final AuditService auditService;
    
    public UserEventConsumer(EmailService emailService, AuditService auditService) {
        this.emailService = emailService;
        this.auditService = auditService;
    }
    
    @RabbitListener(queues = RabbitMQConfig.USER_QUEUE)
    public void handleUserCreated(UserCreatedEvent event) {
        log.info("Received user created event: {}", event);
        
        // Send welcome email
        emailService.sendWelcomeEmail(event.userId(), event.email());
        
        // Create audit record
        auditService.recordUserCreation(event.userId(), event.username());
    }
}

// Event record
public record UserCreatedEvent(
    Long userId,
    String username,
    String email,
    Instant timestamp
) {
    public UserCreatedEvent {
        if (timestamp == null) {
            timestamp = Instant.now();
        }
    }
}
```

### Redis Integration for Caching

```java
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.data.redis.core.ReactiveValueOperations;
import reactor.core.publisher.Mono;

@Service
public class CacheService {
    
    private final ReactiveRedisTemplate<String, Object> redisTemplate;
    private final ReactiveValueOperations<String, Object> valueOps;
    
    public CacheService(ReactiveRedisTemplate<String, Object> redisTemplate) {
        this.redisTemplate = redisTemplate;
        this.valueOps = redisTemplate.opsForValue();
    }
    
    public <T> Mono<T> get(String key, Class<T> type) {
        return valueOps.get(key)
                .cast(type);
    }
    
    public <T> Mono<T> getOrCompute(String key, Supplier<Mono<T>> supplier, Duration ttl, Class<T> type) {
        return get(key, type)
                .switchIfEmpty(
                    supplier.get()
                        .flatMap(result -> set(key, result, ttl).thenReturn(result))
                );
    }
    
    public <T> Mono<Boolean> set(String key, T value, Duration ttl) {
        return valueOps.set(key, value, ttl);
    }
    
    public Mono<Boolean> delete(String key) {
        return redisTemplate.delete(key).map(count -> count > 0);
    }
    
    // Rate limiting implementation
    public Mono<Boolean> isRateLimited(String identifier, int maxRequests, Duration window) {
        String key = "rate_limit:" + identifier;
        
        return redisTemplate.opsForValue()
                .increment(key)
                .flatMap(count -> {
                    if (count == 1) {
                        // Set expiration for first request in window
                        return redisTemplate.expire(key, window).thenReturn(count);
                    }
                    return Mono.just(count);
                })
                .map(count -> count > maxRequests);
    }
}
```

## Modern Development Workflow

### Gradle Configuration (build.gradle.kts)

```kotlin
plugins {
    application
    jacoco
    checkstyle
    pmd
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.5"
    id("com.google.cloud.tools.jib") version "3.4.1"
}

group = "com.example"
version = "0.1.0"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

repositories {
    mavenCentral()
}

dependencies {
    // Spring Boot starters
    implementation("org.springframework.boot:spring-boot-starter-webflux")
    implementation("org.springframework.boot:spring-boot-starter-data-r2dbc")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    
    // Database
    implementation("org.postgresql:r2dbc-postgresql")
    implementation("org.flywaydb:flyway-core")
    implementation("org.springframework:spring-jdbc")
    
    // Message queue
    implementation("org.springframework.boot:spring-boot-starter-amqp")
    
    // Caching
    implementation("org.springframework.boot:spring-boot-starter-data-redis-reactive")
    
    // Security
    implementation("org.springframework.boot:spring-boot-starter-security")
    implementation("org.springframework.security:spring-security-oauth2-resource-server")
    
    // Utility libraries
    implementation("com.fasterxml.jackson.core:jackson-databind")
    implementation("com.fasterxml.jackson.datatype:jackson-datatype-jsr310")
    implementation("org.apache.commons:commons-lang3")
    implementation("org.apache.commons:commons-collections4")
    
    // Development tools
    compileOnly("org.projectlombok:lombok")
    annotationProcessor("org.projectlombok:lombok")
    developmentOnly("org.springframework.boot:spring-boot-devtools")
    
    // Testing
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testImplementation("io.projectreactor:reactor-test")
    testImplementation("org.testcontainers:junit-jupiter")
    testImplementation("org.testcontainers:postgresql")
    testImplementation("org.testcontainers:r2dbc")
    testImplementation("org.testcontainers:rabbitmq")
    
    // Performance testing
    testImplementation("org.junit.jupiter:junit-jupiter-params")
    testImplementation("net.jqwik:jqwik:1.8.3")
}

application {
    mainClass.set("com.example.Application")
}

tasks.test {
    useJUnitPlatform()
    testLogging {
        events("passed", "skipped", "failed")
    }
}

// Jib configuration for Docker image generation
jib {
    from {
        image = "eclipse-temurin:21-jre-alpine"
    }
    to {
        image = "my-registry.com/my-app:latest"
    }
    container {
        jvmFlags = listOf("-Xms512m", "-Xmx2g")
        ports = listOf("8080")
        environment = mapOf(
            "SPRING_PROFILES_ACTIVE" to "prod"
        )
    }
}
```

### Pre-commit Configuration

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict
      - id: check-executables-have-shebangs

  - repo: https://github.com/codacy/codacy-checkstyle.git
    rev: 5.1.0
    hooks:
      - id: codacy-checkstyle-java
        args: ["-c", "/path/to/checkstyle.xml"]

  - repo: https://github.com/ulises-jeremias/pre-commit-pmd
    rev: v2.3.1
    hooks:
      - id: pmd
        args: ["-d", "src/main/java", "-r", "ruleset.xml"]

  - repo: local
    hooks:
      - id: gradle-spotless
        name: Code formatting with Spotless
        entry: ./gradlew spotlessApply
        language: system
        files: '\.(java|kt|kts)$'

      - id: gradle-test
        name: Run tests
        entry: ./gradlew test
        language: system
        pass_filenames: false
        always_run: true

      - id: gradle-check
        name: Run code quality checks
        entry: ./gradlew check
        language: system
        pass_filenames: false
        always_run: true
```

### Docker Best Practices

```dockerfile
# Multi-stage Docker build with Jib integration
# This Dockerfile is for local development; use Jib for production

# Build stage
FROM eclipse-temurin:21-jdk-alpine AS builder

WORKDIR /app

# Copy gradle wrapper and cache dependencies
COPY gradlew gradlew.bat gradle/ ./
COPY build.gradle.kts settings.gradle.kts ./

# Download dependencies
RUN ./gradlew --no-daemon dependencies

# Copy source code
COPY src/ ./src/

# Build application
RUN ./gradlew --no-daemon bootJar -x test

# Runtime stage
FROM eclipse-temurin:21-jre-alpine

# Security best practices
RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser && \
    mkdir -p /app && \
    chown -R appuser:appgroup /app

WORKDIR /app

# Copy application jar
COPY --from=builder /app/build/libs/*.jar app.jar

# Set appropriate permissions
RUN chmod +x app.jar

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/actuator/health || exit 1

EXPOSE 8080

# JVM optimizations
ENV JAVA_OPTS="-Xms512m -Xmx2g -XX:+UseG1GC -XX:+UseStringDeduplication"

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

## Backend Development Patterns

### Spring Boot Application Structure

```java
@SpringBootApplication
@EnableReactiveSecurity
@EnableR2dbcAuditing
public class Application {
    
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
    
    @Bean
    public WebFluxConfigurer corsConfigurer() {
        return new WebFluxConfigurer() {
            @Override
            public void addCorsMappings(CorsRegistry registry) {
                registry.addMapping("/api/**")
                        .allowedOrigins("http://localhost:3000")
                        .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
                        .allowedHeaders("*")
                        .maxAge(3600);
            }
        };
    }
}

// Global error handling
@ControllerAdvice
public class GlobalExceptionHandler {
    
    private static final Logger log = LoggerFactory.getLogger(GlobalExceptionHandler.class);
    
    @ExceptionHandler(ValidationException.class)
    public ResponseEntity<ErrorResponse> handleValidationException(ValidationException ex) {
        log.warn("Validation error: {}", ex.getMessage());
        return ResponseEntity.badRequest()
                .body(new ErrorResponse("VALIDATION_ERROR", ex.getMessage()));
    }
    
    @ExceptionHandler(ResourceNotFoundException.class)
    public ResponseEntity<ErrorResponse> handleResourceNotFoundException(ResourceNotFoundException ex) {
        log.warn("Resource not found: {}", ex.getMessage());
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(new ErrorResponse("RESOURCE_NOT_FOUND", ex.getMessage()));
    }
    
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ErrorResponse> handleGenericException(Exception ex) {
        log.error("Unexpected error occurred", ex);
        return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(new ErrorResponse("INTERNAL_ERROR", "An unexpected error occurred"));
    }
}

// Audit configuration
@Configuration
@EnableJpaAuditing(auditorAwareRef = "auditorProvider")
public class AuditConfig {
    
    @Bean
    public AuditorAware<String> auditorProvider() {
        return () -> {
            Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
            if (authentication == null || !authentication.isAuthenticated()) {
                return Optional.of("system");
            }
            return Optional.of(authentication.getName());
        };
    }
}

// Custom metrics with Micrometer
@Component
public class CustomMetrics {
    
    private final Counter userRegistrationCounter;
    private final Timer userLoginTimer;
    private final Gauge activeUserGauge;
    
    public CustomMetrics(MeterRegistry meterRegistry) {
        this.userRegistrationCounter = Counter.builder("users.registered")
                .description("Number of user registrations")
                .register(meterRegistry);
        
        this.userLoginTimer = Timer.builder("users.login.duration")
                .description("User login duration")
                .register(meterRegistry);
        
        this.activeUserGauge = Gauge.builder("users.active")
                .description("Number of active users")
                .register(meterRegistry, this, CustomMetrics::getActiveUserCount);
    }
    
    public void incrementUserRegistration() {
        userRegistrationCounter.increment();
    }
    
    public Timer.Sample startLoginTimer() {
        return Timer.start();
    }
    
    public void recordLoginTime(Timer.Sample sample) {
        sample.stop(userLoginTimer);
    }
    
    private double getActiveUserCount() {
        // Implementation to count active users
        return 0.0;
    }
}
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Java Target**: 21 (LTS) with modern Spring Boot and reactive patterns  

This skill provides comprehensive Java development guidance with 2025 best practices, covering everything from basic project setup to advanced enterprise integration and production deployment patterns.
