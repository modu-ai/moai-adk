---
name: moai-lang-java
description: Java 21 LTS development specialist covering Spring Boot 3.3, virtual threads, pattern matching, and enterprise patterns. Use when building enterprise applications, microservices, or Spring projects.
version: 1.0.0
category: language
tags: [java, spring-boot, jpa, hibernate, virtual-threads, enterprise]
context7-libraries: [/spring-projects/spring-boot, /spring-projects/spring-framework, /spring-projects/spring-security]
related-skills: [moai-lang-kotlin, moai-domain-backend]
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Java 21 LTS Expert - Enterprise development with Spring Boot 3.3, Virtual Threads, and modern Java features.

Auto-Triggers: Java files (`.java`), build files (`pom.xml`, `build.gradle`, `build.gradle.kts`)

Core Capabilities:
- Java 21 LTS: Virtual threads, pattern matching, record patterns, sealed classes
- Spring Boot 3.3: REST controllers, services, repositories, WebFlux reactive
- Spring Security 6: JWT authentication, OAuth2, role-based access control
- JPA/Hibernate 7: Entity mapping, relationships, queries, transactions
- JUnit 5: Unit testing, mocking, TestContainers integration
- Build Tools: Maven 3.9, Gradle 8.5 with Kotlin DSL

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

Record Patterns and Sealed Classes:
```java
record Point(int x, int y) {}
record Rectangle(Point topLeft, Point bottomRight) {}

int area(Rectangle rect) {
    return switch (rect) {
        case Rectangle(Point(var x1, var y1), Point(var x2, var y2)) ->
            Math.abs((x2 - x1) * (y2 - y1));
    };
}

public sealed interface Shape permits Circle, Rectangle {
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
@RequiredArgsConstructor
public class UserController {
    private final UserService userService;

    @GetMapping("/{id}")
    public ResponseEntity<UserDto> getUser(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<UserDto> createUser(@Valid @RequestBody CreateUserRequest request) {
        UserDto user = userService.create(request);
        URI location = URI.create("/api/users/" + user.id());
        return ResponseEntity.created(location).body(user);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        return userService.delete(id)
            ? ResponseEntity.noContent().build()
            : ResponseEntity.notFound().build();
    }
}
```

Service Layer:
```java
@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class UserService {
    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;

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
        return userRepository.save(user);
    }
}
```

### Spring Security 6

Security Configuration:
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
            .csrf(csrf -> csrf.disable())
            .build();
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }
}
```

### JPA/Hibernate Patterns

Entity Definition:
```java
@Entity
@Table(name = "users")
@Getter @Setter
@NoArgsConstructor @Builder
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String name;

    @Column(nullable = false, unique = true)
    private String email;

    @Enumerated(EnumType.STRING)
    private UserStatus status;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Order> orders = new ArrayList<>();
}
```

Repository with Custom Queries:
```java
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);
    boolean existsByEmail(String email);

    @Query("SELECT u FROM User u LEFT JOIN FETCH u.orders WHERE u.id = :id")
    Optional<User> findByIdWithOrders(@Param("id") Long id);

    Page<User> findByNameContainingIgnoreCase(String name, Pageable pageable);
}
```

DTOs as Records:
```java
public record UserDto(Long id, String name, String email, UserStatus status) {
    public static UserDto from(User user) {
        return new UserDto(user.getId(), user.getName(), user.getEmail(), user.getStatus());
    }
}

public record CreateUserRequest(
    @NotBlank @Size(min = 2, max = 100) String name,
    @NotBlank @Email String email,
    @NotBlank @Size(min = 8) String password
) {}
```

---

## Advanced Patterns

### Virtual Threads Integration

```java
@Service
@RequiredArgsConstructor
public class AsyncUserService {
    public UserWithDetails fetchUserDetails(Long userId) throws Exception {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
            Supplier<User> userTask = scope.fork(() -> userRepo.findById(userId).orElseThrow());
            Supplier<List<Order>> ordersTask = scope.fork(() -> orderRepo.findByUserId(userId));
            scope.join().throwIfFailed();
            return new UserWithDetails(userTask.get(), ordersTask.get());
        }
    }

    public void processUsersInParallel(List<Long> userIds) {
        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            userIds.stream().map(id -> executor.submit(() -> processUser(id))).toList();
        }
    }
}
```

### Build Configuration

Maven 3.9:
```xml
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.3.0</version>
    </parent>
    <properties><java.version>21</java.version></properties>
    <dependencies>
        <dependency><groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId></dependency>
        <dependency><groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId></dependency>
    </dependencies>
</project>
```

Gradle 8.5 (Kotlin DSL):
```kotlin
plugins {
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.4"
    java
}
java { toolchain { languageVersion = JavaLanguageVersion.of(21) } }
dependencies {
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    testImplementation("org.springframework.boot:spring-boot-starter-test")
}
```

### Testing with JUnit 5

Unit Testing:
```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock private UserRepository userRepository;
    @InjectMocks private UserService userService;

    @Test
    void shouldCreateUser() {
        when(userRepository.existsByEmail(anyString())).thenReturn(false);
        when(userRepository.save(any())).thenReturn(User.builder().id(1L).build());
        var result = userService.create(new CreateUserRequest("John", "john@test.com", "pass"));
        assertThat(result.getId()).isEqualTo(1L);
    }
}
```

Integration Testing with TestContainers:
```java
@Testcontainers
@SpringBootTest
class UserRepositoryTest {
    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine");

    @DynamicPropertySource
    static void props(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
    }

    @Autowired UserRepository repo;

    @Test
    void shouldSaveUser() {
        var user = repo.save(User.builder().name("John").email("john@test.com").build());
        assertThat(user.getId()).isNotNull();
    }
}
```

---

## Context7 Integration

Library mappings for latest documentation:
- `/spring-projects/spring-boot` - Spring Boot 3.3 documentation
- `/spring-projects/spring-framework` - Spring Framework core
- `/spring-projects/spring-security` - Spring Security 6
- `/hibernate/hibernate-orm` - Hibernate 7 ORM patterns
- `/junit-team/junit5` - JUnit 5 testing framework

---

## Works Well With

- `moai-lang-kotlin` - Kotlin interoperability and Spring Kotlin extensions
- `moai-domain-backend` - REST API, GraphQL, microservices architecture
- `moai-domain-database` - JPA, Hibernate, R2DBC patterns
- `moai-foundation-quality` - JUnit 5, Mockito, TestContainers integration
- `moai-infra-docker` - JVM container optimization

---

## Troubleshooting

Common Issues:
- Version mismatch: `java -version`, check `JAVA_HOME` points to Java 21
- Compilation errors: `mvn clean compile -X` or `gradle build --info`
- Virtual thread issues: Ensure Java 21+ with `--enable-preview` if needed
- JPA lazy loading: Use `@Transactional` or `JOIN FETCH` queries

Performance Tips:
- Enable Virtual Threads: `spring.threads.virtual.enabled=true`
- Use GraalVM Native Image for faster startup
- Configure connection pooling with HikariCP

---

## Advanced Documentation

For comprehensive reference materials:
- [reference.md](reference.md) - Java 21 features, Context7 mappings, performance
- [examples.md](examples.md) - Production-ready Spring Boot examples

---

Last Updated: 2025-12-07
Status: Production Ready (v1.0.0)
