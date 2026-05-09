# Java 21 Production Examples

## Complete REST API Implementation

### Spring Boot 3.3 User Service

UserController.java:
```java
@RestController
@RequestMapping("/api/v1/users")
@RequiredArgsConstructor
@Validated
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
        var location = URI.create("/api/v1/users/" + user.getId());
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

    @ExceptionHandler(DuplicateEmailException.class)
    public ResponseEntity<ProblemDetail> handleDuplicateEmail(DuplicateEmailException ex) {
        var problem = ProblemDetail.forStatusAndDetail(
            HttpStatus.CONFLICT, ex.getMessage());
        problem.setTitle("Duplicate Email");
        return ResponseEntity.status(HttpStatus.CONFLICT).body(problem);
    }
}
```

UserService.java:
```java
@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
@Slf4j
public class UserService {
    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final ApplicationEventPublisher eventPublisher;

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
        log.info("Creating user with email: {}", request.email());

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
        eventPublisher.publishEvent(new UserCreatedEvent(saved.getId(), saved.getEmail()));

        log.info("User created with id: {}", saved.getId());
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
        eventPublisher.publishEvent(new UserDeletedEvent(id));
        return true;
    }
}
```

User.java (Entity):
```java
@Entity
@Table(name = "users", indexes = {
    @Index(name = "idx_users_email", columnList = "email"),
    @Index(name = "idx_users_status", columnList = "status")
})
@Getter @Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class User {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, length = 100)
    private String name;

    @Column(nullable = false, unique = true, length = 255)
    private String email;

    @Column(nullable = false)
    private String passwordHash;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private UserStatus status;

    @CreatedDate
    @Column(nullable = false, updatable = false)
    private Instant createdAt;

    @LastModifiedDate
    private Instant updatedAt;

    @Version
    private Long version;

    @OneToMany(mappedBy = "user", cascade = CascadeType.ALL, orphanRemoval = true)
    @Builder.Default
    private List<Order> orders = new ArrayList<>();

    public void addOrder(Order order) {
        orders.add(order);
        order.setUser(this);
    }

    public void removeOrder(Order order) {
        orders.remove(order);
        order.setUser(null);
    }

    @PrePersist
    protected void onCreate() {
        createdAt = Instant.now();
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = Instant.now();
    }
}

public enum UserStatus {
    PENDING, ACTIVE, SUSPENDED, DELETED
}
```

UserRepository.java:
```java
public interface UserRepository extends JpaRepository<User, Long> {
    Optional<User> findByEmail(String email);

    boolean existsByEmail(String email);

    @Query("SELECT u FROM User u WHERE u.status = :status")
    List<User> findByStatus(@Param("status") UserStatus status);

    @Query("SELECT u FROM User u LEFT JOIN FETCH u.orders WHERE u.id = :id")
    Optional<User> findByIdWithOrders(@Param("id") Long id);

    Page<User> findByNameContainingIgnoreCaseOrEmailContainingIgnoreCase(
        String name, String email, Pageable pageable);

    @Modifying
    @Query("UPDATE User u SET u.status = :status WHERE u.id = :id")
    int updateStatus(@Param("id") Long id, @Param("status") UserStatus status);
}
```

DTOs (Records):
```java
public record UserDto(
    Long id,
    String name,
    String email,
    UserStatus status,
    Instant createdAt,
    Instant updatedAt
) {
    public static UserDto from(User user) {
        return new UserDto(
            user.getId(),
            user.getName(),
            user.getEmail(),
            user.getStatus(),
            user.getCreatedAt(),
            user.getUpdatedAt()
        );
    }
}

public record CreateUserRequest(
    @NotBlank(message = "Name is required")
    @Size(min = 2, max = 100, message = "Name must be between 2 and 100 characters")
    String name,

    @NotBlank(message = "Email is required")
    @Email(message = "Invalid email format")
    String email,

    @NotBlank(message = "Password is required")
    @Size(min = 8, message = "Password must be at least 8 characters")
    String password
) {}

public record UpdateUserRequest(
    @NotBlank(message = "Name is required")
    @Size(min = 2, max = 100, message = "Name must be between 2 and 100 characters")
    String name,

    @Email(message = "Invalid email format")
    String email
) {}
```

---

## Virtual Threads Examples

### Async Service with Structured Concurrency

```java
@Service
@RequiredArgsConstructor
@Slf4j
public class AsyncUserService {
    private final UserRepository userRepository;
    private final OrderRepository orderRepository;
    private final NotificationService notificationService;
    private final ExternalApiClient externalApiClient;

    public UserWithDetails fetchUserDetails(Long userId) throws Exception {
        log.info("Fetching user details for userId: {}", userId);

        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
            Supplier<User> userTask = scope.fork(() ->
                userRepository.findById(userId)
                    .orElseThrow(() -> new UserNotFoundException(userId)));

            Supplier<List<Order>> ordersTask = scope.fork(() ->
                orderRepository.findByUserId(userId));

            Supplier<List<Notification>> notificationsTask = scope.fork(() ->
                notificationService.getUnreadNotifications(userId));

            Supplier<UserProfile> profileTask = scope.fork(() ->
                externalApiClient.fetchUserProfile(userId));

            scope.join().throwIfFailed();

            return new UserWithDetails(
                userTask.get(),
                ordersTask.get(),
                notificationsTask.get(),
                profileTask.get()
            );
        }
    }

    public List<UserSummary> processUsersInParallel(List<Long> userIds) {
        log.info("Processing {} users in parallel", userIds.size());

        try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
            var futures = userIds.stream()
                .map(id -> executor.submit(() -> processUser(id)))
                .toList();

            return futures.stream()
                .map(future -> {
                    try {
                        return future.get();
                    } catch (Exception e) {
                        log.error("Failed to process user", e);
                        return null;
                    }
                })
                .filter(Objects::nonNull)
                .toList();
        }
    }

    private UserSummary processUser(Long userId) {
        var user = userRepository.findById(userId).orElseThrow();
        var orderCount = orderRepository.countByUserId(userId);
        return new UserSummary(user.getId(), user.getName(), orderCount);
    }
}

public record UserWithDetails(
    User user,
    List<Order> orders,
    List<Notification> notifications,
    UserProfile profile
) {}

public record UserSummary(Long id, String name, int orderCount) {}
```

### Virtual Thread Configuration

```java
@Configuration
public class VirtualThreadConfig {

    @Bean
    public TomcatProtocolHandlerCustomizer<?> protocolHandlerVirtualThreadExecutorCustomizer() {
        return protocolHandler -> {
            protocolHandler.setExecutor(Executors.newVirtualThreadPerTaskExecutor());
        };
    }

    @Bean
    public AsyncTaskExecutor applicationTaskExecutor() {
        return new TaskExecutorAdapter(Executors.newVirtualThreadPerTaskExecutor());
    }
}
```

---

## Testing Examples

### Unit Tests with JUnit 5 and Mockito

```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {
    @Mock private UserRepository userRepository;
    @Mock private PasswordEncoder passwordEncoder;
    @Mock private ApplicationEventPublisher eventPublisher;
    @InjectMocks private UserService userService;

    @Nested
    @DisplayName("Create User Tests")
    class CreateUserTests {
        @Test
        @DisplayName("Should create user successfully when email is unique")
        void shouldCreateUserSuccessfully() {
            // Arrange
            var request = new CreateUserRequest("John Doe", "john@example.com", "password123");
            var savedUser = User.builder()
                .id(1L)
                .name("John Doe")
                .email("john@example.com")
                .status(UserStatus.PENDING)
                .build();

            when(userRepository.existsByEmail("john@example.com")).thenReturn(false);
            when(passwordEncoder.encode("password123")).thenReturn("hashedPassword");
            when(userRepository.save(any(User.class))).thenReturn(savedUser);

            // Act
            var result = userService.create(request);

            // Assert
            assertThat(result).isNotNull();
            assertThat(result.getId()).isEqualTo(1L);
            assertThat(result.getName()).isEqualTo("John Doe");
            assertThat(result.getEmail()).isEqualTo("john@example.com");
            assertThat(result.getStatus()).isEqualTo(UserStatus.PENDING);

            verify(userRepository).save(argThat(user ->
                user.getName().equals("John Doe") &&
                user.getEmail().equals("john@example.com") &&
                user.getPasswordHash().equals("hashedPassword")
            ));
            verify(eventPublisher).publishEvent(any(UserCreatedEvent.class));
        }

        @Test
        @DisplayName("Should throw exception when email already exists")
        void shouldThrowExceptionForDuplicateEmail() {
            // Arrange
            var request = new CreateUserRequest("John", "existing@example.com", "password");
            when(userRepository.existsByEmail("existing@example.com")).thenReturn(true);

            // Act & Assert
            assertThatThrownBy(() -> userService.create(request))
                .isInstanceOf(DuplicateEmailException.class)
                .hasMessageContaining("existing@example.com");

            verify(userRepository, never()).save(any());
            verify(eventPublisher, never()).publishEvent(any());
        }
    }

    @Nested
    @DisplayName("Find User Tests")
    class FindUserTests {
        @Test
        @DisplayName("Should return user when found")
        void shouldReturnUserWhenFound() {
            var user = User.builder().id(1L).name("John").build();
            when(userRepository.findById(1L)).thenReturn(Optional.of(user));

            var result = userService.findById(1L);

            assertThat(result).isPresent();
            assertThat(result.get().getName()).isEqualTo("John");
        }

        @Test
        @DisplayName("Should return empty when user not found")
        void shouldReturnEmptyWhenNotFound() {
            when(userRepository.findById(999L)).thenReturn(Optional.empty());

            var result = userService.findById(999L);

            assertThat(result).isEmpty();
        }
    }

    @ParameterizedTest
    @ValueSource(strings = {"john", "doe", "example.com"})
    @DisplayName("Should search users by query")
    void shouldSearchUsersByQuery(String query) {
        var pageable = PageRequest.of(0, 10);
        var users = new PageImpl<>(List.of(
            User.builder().id(1L).name("John Doe").email("john@example.com").build()
        ));
        when(userRepository.findByNameContainingIgnoreCaseOrEmailContainingIgnoreCase(
            query, query, pageable)).thenReturn(users);

        var result = userService.searchUsers(query, pageable);

        assertThat(result.getContent()).hasSize(1);
    }
}
```

### Integration Tests with TestContainers

```java
@Testcontainers
@SpringBootTest
@AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)
@Transactional
class UserRepositoryIntegrationTest {

    @Container
    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>("postgres:16-alpine")
        .withDatabaseName("testdb")
        .withUsername("test")
        .withPassword("test");

    @DynamicPropertySource
    static void configureProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", postgres::getJdbcUrl);
        registry.add("spring.datasource.username", postgres::getUsername);
        registry.add("spring.datasource.password", postgres::getPassword);
    }

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private TestEntityManager entityManager;

    @BeforeEach
    void setUp() {
        userRepository.deleteAll();
    }

    @Test
    @DisplayName("Should save and find user by email")
    void shouldSaveAndFindUserByEmail() {
        // Arrange
        var user = User.builder()
            .name("John Doe")
            .email("john@example.com")
            .passwordHash("hashedPassword")
            .status(UserStatus.ACTIVE)
            .build();

        // Act
        var saved = userRepository.save(user);
        entityManager.flush();
        entityManager.clear();

        // Assert
        var found = userRepository.findByEmail("john@example.com");
        assertThat(found).isPresent();
        assertThat(found.get().getName()).isEqualTo("John Doe");
        assertThat(found.get().getId()).isEqualTo(saved.getId());
    }

    @Test
    @DisplayName("Should check email existence correctly")
    void shouldCheckEmailExistence() {
        var user = User.builder()
            .name("Jane")
            .email("jane@example.com")
            .passwordHash("hash")
            .status(UserStatus.ACTIVE)
            .build();
        userRepository.save(user);

        assertThat(userRepository.existsByEmail("jane@example.com")).isTrue();
        assertThat(userRepository.existsByEmail("nonexistent@example.com")).isFalse();
    }

    @Test
    @DisplayName("Should find users by status")
    void shouldFindUsersByStatus() {
        userRepository.saveAll(List.of(
            User.builder().name("Active1").email("a1@test.com").passwordHash("h").status(UserStatus.ACTIVE).build(),
            User.builder().name("Active2").email("a2@test.com").passwordHash("h").status(UserStatus.ACTIVE).build(),
            User.builder().name("Pending").email("p@test.com").passwordHash("h").status(UserStatus.PENDING).build()
        ));

        var activeUsers = userRepository.findByStatus(UserStatus.ACTIVE);

        assertThat(activeUsers).hasSize(2);
        assertThat(activeUsers).extracting(User::getStatus)
            .containsOnly(UserStatus.ACTIVE);
    }

    @Test
    @DisplayName("Should search users with pagination")
    void shouldSearchUsersWithPagination() {
        userRepository.saveAll(List.of(
            User.builder().name("John Smith").email("john@test.com").passwordHash("h").status(UserStatus.ACTIVE).build(),
            User.builder().name("Jane Doe").email("jane@test.com").passwordHash("h").status(UserStatus.ACTIVE).build(),
            User.builder().name("Bob Johnson").email("bob@test.com").passwordHash("h").status(UserStatus.ACTIVE).build()
        ));

        var pageable = PageRequest.of(0, 10);
        var result = userRepository.findByNameContainingIgnoreCaseOrEmailContainingIgnoreCase(
            "john", "john", pageable);

        assertThat(result.getContent()).hasSize(2);
        assertThat(result.getContent()).extracting(User::getName)
            .containsExactlyInAnyOrder("John Smith", "Bob Johnson");
    }
}
```

---

## Spring Security Examples

### JWT Authentication

```java
@RestController
@RequestMapping("/api/auth")
@RequiredArgsConstructor
public class AuthController {
    private final AuthenticationManager authenticationManager;
    private final JwtTokenProvider tokenProvider;
    private final UserService userService;

    @PostMapping("/login")
    public ResponseEntity<AuthResponse> login(@Valid @RequestBody LoginRequest request) {
        var authentication = authenticationManager.authenticate(
            new UsernamePasswordAuthenticationToken(request.email(), request.password())
        );

        SecurityContextHolder.getContext().setAuthentication(authentication);
        var token = tokenProvider.generateToken(authentication);

        return ResponseEntity.ok(new AuthResponse(token, "Bearer"));
    }

    @PostMapping("/register")
    public ResponseEntity<UserDto> register(@Valid @RequestBody CreateUserRequest request) {
        var user = userService.create(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(UserDto.from(user));
    }

    @GetMapping("/me")
    public ResponseEntity<UserDto> getCurrentUser(@AuthenticationPrincipal UserDetails userDetails) {
        return userService.findByEmail(userDetails.getUsername())
            .map(UserDto::from)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }
}

public record LoginRequest(
    @NotBlank @Email String email,
    @NotBlank String password
) {}

public record AuthResponse(String token, String type) {}
```

### Security Configuration

```java
@Configuration
@EnableWebSecurity
@EnableMethodSecurity
@RequiredArgsConstructor
public class SecurityConfig {
    private final JwtTokenProvider tokenProvider;
    private final UserDetailsService userDetailsService;

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        return http
            .csrf(AbstractHttpConfigurer::disable)
            .cors(cors -> cors.configurationSource(corsConfigurationSource()))
            .sessionManagement(session ->
                session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/api/auth/**").permitAll()
                .requestMatchers("/api/public/**").permitAll()
                .requestMatchers("/actuator/health").permitAll()
                .requestMatchers("/api/admin/**").hasRole("ADMIN")
                .anyRequest().authenticated()
            )
            .addFilterBefore(jwtAuthenticationFilter(), UsernamePasswordAuthenticationFilter.class)
            .exceptionHandling(ex -> ex
                .authenticationEntryPoint(new JwtAuthenticationEntryPoint())
                .accessDeniedHandler(new JwtAccessDeniedHandler())
            )
            .build();
    }

    @Bean
    public JwtAuthenticationFilter jwtAuthenticationFilter() {
        return new JwtAuthenticationFilter(tokenProvider, userDetailsService);
    }

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }

    @Bean
    public AuthenticationManager authenticationManager(AuthenticationConfiguration config) throws Exception {
        return config.getAuthenticationManager();
    }

    @Bean
    public CorsConfigurationSource corsConfigurationSource() {
        var configuration = new CorsConfiguration();
        configuration.setAllowedOrigins(List.of("http://localhost:3000"));
        configuration.setAllowedMethods(List.of("GET", "POST", "PUT", "DELETE", "OPTIONS"));
        configuration.setAllowedHeaders(List.of("*"));
        configuration.setAllowCredentials(true);

        var source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", configuration);
        return source;
    }
}
```

---

## Build Configuration Examples

### Maven pom.xml

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         https://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.3.0</version>
    </parent>

    <groupId>com.example</groupId>
    <artifactId>user-service</artifactId>
    <version>1.0.0</version>

    <properties>
        <java.version>21</java.version>
        <testcontainers.version>1.19.7</testcontainers.version>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-security</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-validation</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-actuator</artifactId>
        </dependency>

        <dependency>
            <groupId>org.postgresql</groupId>
            <artifactId>postgresql</artifactId>
            <scope>runtime</scope>
        </dependency>
        <dependency>
            <groupId>org.projectlombok</groupId>
            <artifactId>lombok</artifactId>
            <optional>true</optional>
        </dependency>
        <dependency>
            <groupId>io.jsonwebtoken</groupId>
            <artifactId>jjwt-api</artifactId>
            <version>0.12.5</version>
        </dependency>

        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-test</artifactId>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>org.testcontainers</groupId>
            <artifactId>postgresql</artifactId>
            <version>${testcontainers.version}</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
```

### Gradle build.gradle.kts

```kotlin
plugins {
    java
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.4"
}

group = "com.example"
version = "1.0.0"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(21)
    }
}

configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}

repositories {
    mavenCentral()
}

dependencies {
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-security")
    implementation("org.springframework.boot:spring-boot-starter-validation")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("io.jsonwebtoken:jjwt-api:0.12.5")

    runtimeOnly("org.postgresql:postgresql")
    runtimeOnly("io.jsonwebtoken:jjwt-impl:0.12.5")
    runtimeOnly("io.jsonwebtoken:jjwt-jackson:0.12.5")

    compileOnly("org.projectlombok:lombok")
    annotationProcessor("org.projectlombok:lombok")

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testImplementation("org.testcontainers:postgresql")
    testImplementation("org.testcontainers:junit-jupiter")
}

tasks.withType<Test> {
    useJUnitPlatform()
}
```

---

Last Updated: 2025-12-07
Version: 1.0.0
