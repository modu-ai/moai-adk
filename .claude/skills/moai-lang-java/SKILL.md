---
name: moai-lang-java
version: 3.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Java 21 enterprise development with Spring Boot 3.5.3, virtual threads, structured concurrency, GraalVM AOT, and Context7 MCP integration.
keywords: ['java', 'java21', 'spring-boot', 'enterprise', 'virtual-threads', 'graalvm', 'aot', 'structured-concurrency']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# 🚀 Java 21 Enterprise Development Premium Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-java |
| **Version** | 3.0.0 (2025-11-06) - Premium Edition |
| **Allowed tools** | Read, Bash, Context7 MCP Integration |
| **Auto-load** | Java projects, Spring Boot applications, enterprise development |
| **Tier** | Premium Language |
| **Context7 Integration** | Java SE 21 + Spring Boot 3.5.3 Official Docs |

---

## 🎯 What It Does

**Enterprise-grade Java 21 development** with Spring Boot 3.5.3 ecosystem, virtual threads, structured concurrency, GraalVM AOT compilation, and production-ready patterns.

### Core Capabilities

**🏗️ Modern Java 21 Features**:
- **Virtual Threads**: Lightweight concurrency with `Thread.ofVirtual()`
- **Structured Concurrency**: `StructuredTaskScope` for managed concurrent operations
- **Pattern Matching**: Enhanced `instanceof` and `switch` expressions
- **Record Patterns**: Destructuring for complex data structures
- **String Templates**: Compile-time safe string interpolation

**⚡ Spring Boot 3.5.3 Enterprise**:
- **Production Configuration**: AOT compilation, GraalVM native images
- **Observability**: Built-in metrics, tracing, and health checks
- **Security**: OAuth2, JWT, CSRF protection with latest Spring Security
- **Data Access**: Spring Data JPA, R2DBC, reactive programming
- **Testing**: @SpringBootTest integration with virtual threads

**🔧 Enterprise Development Tools**:
- **Build Systems**: Maven 3.9+, Gradle 8.5+ with Kotlin DSL
- **Testing**: JUnit 5.10+, Testcontainers, AssertJ
- **Code Quality**: SpotBugs, Checkstyle, PMD, SonarQube integration
- **CI/CD**: GitHub Actions, Docker, Kubernetes deployment

---

## 🌟 Enterprise Patterns & Best Practices

### Virtual Threading Architecture

```java
// High-throughput web service with virtual threads
@RestController
public class OrderService {

    @Autowired
    private OrderRepository orderRepository;

    @PostMapping("/orders")
    public CompletableFuture<OrderResponse> createOrder(@RequestBody OrderRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            // Virtual thread automatically used by Spring Boot 3.5+
            return processOrder(request);
        });
    }

    // Custom virtual thread pool for specific workloads
    @Bean
    public ExecutorService virtualThreadExecutor() {
        return Executors.newVirtualThreadPerTaskExecutor();
    }
}
```

### Structured Concurrency for Complex Workflows

```java
public class OrderProcessingService {

    public OrderResult processComplexOrder(Order order) {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            // Parallel inventory, payment, and shipping checks
            Subtask<InventoryCheck> inventoryTask = scope.fork(() ->
                inventoryService.checkAvailability(order));

            Subtask<PaymentResult> paymentTask = scope.fork(() ->
                paymentService.processPayment(order));

            Subtask<ShippingQuote> shippingTask = scope.fork(() ->
                shippingService.calculateShipping(order));

            scope.join();
            scope.throwIfFailed();

            return OrderResult.builder()
                .inventory(inventoryTask.get())
                .payment(paymentTask.get())
                .shipping(shippingTask.get())
                .build();
        }
    }
}
```

### Spring Boot 3.5.3 Production Configuration

```yaml
# application.yml for production
server:
  port: 8080
  tomcat:
    threads:
      max: 200
      min-spare: 10

spring:
  application:
    name: enterprise-service

  # Database configuration
  datasource:
    url: jdbc:postgresql://${DB_HOST:localhost}:5432/${DB_NAME:enterprise}
    username: ${DB_USER:app}
    password: ${DB_PASSWORD:secret}
    hikari:
      maximum-pool-size: 20
      minimum-idle: 5

  # JPA configuration
  jpa:
    hibernate:
      ddl-auto: validate
    show-sql: false
    properties:
      hibernate:
        format_sql: true

  # Virtual threads configuration
  threads:
    virtual:
      enabled: true

# Management endpoints
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics,prometheus
  endpoint:
    health:
      show-details: always
  metrics:
    export:
      prometheus:
        enabled: true
```

### GraalVM Native Image Configuration

```java
@NativeImageHint(
    name = "enterprise-service",
    mainClass = "com.enterprise.EnterpriseApplication",
    args = {"--enable-url-protocols=http,https", "--enable-all-security-services"}
)
public class NativeImageConfiguration {

    @RegisterReflectionForBinding({
        Order.class,
        Customer.class,
        Product.class
    })
    public void registerReflectionHints() {
        // Reflection hints for native compilation
    }

    @RegisterForReflection(
        classes = {
            org.hibernate.internal.SessionImpl.class,
            org.postgresql.jdbc.PgConnection.class
        }
    )
    public void registerJdbcHints() {
        // JDBC driver reflection hints
    }
}
```

---

## 🔧 Modern Development Workflow

### Build Configuration (Gradle 8.5+ with Kotlin DSL)

```kotlin
// build.gradle.kts
plugins {
    id("org.springframework.boot") version "3.5.3"
    id("io.spring.dependency-management") version "1.1.6"
    kotlin("jvm") version "1.9.24"
    kotlin("plugin.spring") version "1.9.24"
    kotlin("plugin.jpa") version "1.9.24"
    id("org.graalvm.buildtools.native") version "0.10.2"
}

group = "com.enterprise"
version = "1.0.0"

java {
    sourceCompatibility = JavaVersion.VERSION_21
}

configurations {
    compileOnly {
        extendsFrom(configurations.annotationProcessor.get())
    }
}

dependencies {
    // Spring Boot starters
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.springframework.boot:spring-boot-starter-data-jpa")
    implementation("org.springframework.boot:spring-boot-starter-security")
    implementation("org.springframework.boot:spring-boot-starter-actuator")
    implementation("org.springframework.boot:spring-boot-starter-validation")

    // Kotlin
    implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
    implementation("org.jetbrains.kotlin:kotlin-reflect")

    // Database
    runtimeOnly("org.postgresql:postgresql")
    implementation("org.liquibase:liquibase-core")

    // Virtual threads support
    implementation("org.springframework:spring-context")

    // Testing
    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation("org.springframework.security:spring-security-test")
    testImplementation("org.testcontainers:junit-jupiter")
    testImplementation("org.testcontainers:postgresql")

    // Code quality
    annotationProcessor("org.springframework.boot:spring-boot-configuration-processor")

    // Observability
    implementation("io.micrometer:micrometer-registry-prometheus")
    implementation("io.opentelemetry:opentelemetry-api")
}

// Virtual threads testing configuration
tasks.withType<Test> {
    jvmArgs = listOf("-XX:+UseVirtualThreads")
}

// Native image build
graalvmNative {
    binaries {
        getByName("main") {
            imageName.set("enterprise-service")
            buildArgs.add("--enable-url-protocols=http,https")
        }
    }
}
```

### Advanced Testing with Virtual Threads

```java
@SpringBootTest
@TestMethodOrder(OrderAnnotation.class)
class OrderServiceIntegrationTest {

    @Autowired
    private OrderService orderService;

    @Test
    @Order(1)
    void testVirtualThreadPerformance() {
        // Test with virtual threads enabled
        var startTime = System.currentTimeMillis();

        var futures = IntStream.range(0, 1000)
            .mapToObj(i -> CompletableFuture.runAsync(() ->
                orderService.processOrder(createTestOrder(i))))
            .toList();

        CompletableFuture.allOf(futures.toArray(new CompletableFuture[0])).join();

        var duration = System.currentTimeMillis() - startTime;
        assertTrue("Virtual threads should process 1000 orders quickly",
                  duration < 5000);
    }

    @Test
    @Order(2)
    void testStructuredConcurrency() {
        var order = createComplexOrder();

        var result = orderService.processComplexOrder(order);

        assertNotNull(result);
        assertTrue(result.inventory().available());
        assertEquals(PaymentStatus.APPROVED, result.payment().status());
        assertNotNull(result.shipping().quote());
    }
}
```

---

## 📊 Performance Optimization Strategies

### Virtual Threads Best Practices

```java
@Component
public class OptimizedOrderProcessor {

    private final ExecutorService virtualThreadExecutor;

    public OptimizedOrderProcessor() {
        // Custom virtual thread executor with tuned parameters
        this.virtualThreadExecutor = Executors.newVirtualThreadPerTaskExecutor();
    }

    @Async("virtualThreadExecutor")
    public CompletableFuture<ProcessingResult> processOrders(List<Order> orders) {
        try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {

            var futures = orders.stream()
                .map(order -> scope.fork(() -> processSingleOrder(order)))
                .toList();

            scope.join();
            scope.throwIfFailed();

            var results = futures.stream()
                .map(Subtask::get)
                .toList();

            return CompletableFuture.completedFuture(
                ProcessingResult.builder()
                    .processed(results.size())
                    .results(results)
                    .build()
            );
        }
    }

    private OrderResult processSingleOrder(Order order) {
        // Business logic with blocking I/O operations
        return orderProcessor.process(order);
    }
}
```

### Connection Pooling for Virtual Threads

```java
@Configuration
public class DatabaseConfiguration {

    @Bean
    @Primary
    public DataSource primaryDataSource() {
        HikariConfig config = new HikariConfig();
        config.setJdbcUrl("jdbc:postgresql://localhost:5432/primary");
        config.setMaximumPoolSize(50); // Increase for virtual thread workloads
        config.setMinimumIdle(10);
        config.setIdleTimeout(300000);
        config.setConnectionTimeout(20000);
        config.setLeakDetectionThreshold(60000);
        return new HikariDataSource(config);
    }

    @Bean
    public PlatformTransactionManager transactionManager(DataSource dataSource) {
        return new DataSourceTransactionManager(dataSource);
    }
}
```

---

## 🔒 Security Best Practices

### Spring Security 6.x Configuration

```java
@Configuration
@EnableWebSecurity
@EnableMethodSecurity(prePostEnabled = true)
public class SecurityConfiguration {

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .csrf(csrf -> csrf
                .csrfTokenRepository(CookieCsrfTokenRepository.withHttpOnlyFalse())
                .ignoringRequestMatchers("/api/public/**"))
            .sessionManagement(session -> session
                .sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/api/public/**").permitAll()
                .requestMatchers("/api/admin/**").hasRole("ADMIN")
                .requestMatchers("/actuator/health").permitAll()
                .anyRequest().authenticated())
            .oauth2ResourceServer(oauth2 -> oauth2
                .jwt(jwt -> jwt.jwtDecoder(jwtDecoder())))
            .headers(headers -> headers
                .frameOptions().deny()
                .contentTypeOptions().and()
                .httpStrictTransportSecurity());

        return http.build();
    }

    @Bean
    public JwtDecoder jwtDecoder() {
        // JWT configuration for OAuth2/OIDC
        return NimbusJwtDecoder.withJwkSetUri("https://auth.example.com/.well-known/jwks.json")
            .build();
    }
}
```

---

## 📈 Monitoring & Observability

### Custom Metrics with Micrometer

```java
@Component
public class OrderMetrics {

    private final Counter orderCounter;
    private final Timer orderProcessingTimer;
    private final Gauge activeOrdersGauge;
    private final AtomicLong activeOrders = new AtomicLong(0);

    public OrderMetrics(MeterRegistry meterRegistry) {
        this.orderCounter = Counter.builder("orders.created")
            .description("Total number of orders created")
            .register(meterRegistry);

        this.orderProcessingTimer = Timer.builder("orders.processing.time")
            .description("Time taken to process orders")
            .register(meterRegistry);

        this.activeOrdersGauge = Gauge.builder("orders.active")
            .description("Number of currently processing orders")
            .register(meterRegistry, this, OrderMetrics::getActiveOrders);
    }

    public void recordOrderCreated(OrderType type) {
        orderCounter.increment(Tags.of("type", type.name()));
    }

    public <T> T recordProcessingTime(Supplier<T> operation) {
        return Timer.Sample.start(Metrics.globalRegistry)
            .stop(orderProcessingTimer, operation::get);
    }

    public void incrementActiveOrders() {
        activeOrders.incrementAndGet();
    }

    public void decrementActiveOrders() {
        activeOrders.decrementAndGet();
    }

    private double getActiveOrders() {
        return activeOrders.get();
    }
}
```

---

## 🔄 Context7 MCP Integration

### Real-time Documentation Access

```java
@Component
public class JavaDocumentationService {

    private static final String JAVA_21_API = "/websites/oracle-en-java-javase-21-api";
    private static final String SPRING_BOOT_API = "/websites/spring_io-spring-boot-api-java";

    @Autowired
    private Context7Client context7Client;

    public String getVirtualThreadDocumentation() {
        return context7Client.getLibraryDocs(
            JAVA_21_API,
            "virtual threads Thread.ofVirtual()"
        );
    }

    public String getSpringBootAotConfiguration() {
        return context7Client.getLibraryDocs(
            SPRING_BOOT_API,
            "SpringApplicationAotProcessor GraalVM AOT"
        );
    }

    public String getLatestBestPractice(String feature) {
        // Always get the most current documentation
        return context7Client.getLibraryDocs(JAVA_21_API, feature);
    }
}
```

---

## 📚 Progressive Disclosure Examples

### High Freedom (Quick Answer - 15 tokens)
"Use Spring Boot 3.5.3 with virtual threads for high-concurrency applications."

### Medium Freedom (Detailed Guidance - 35 tokens)
"Configure `Executors.newVirtualThreadPerTaskExecutor()` for I/O-bound tasks, use `@Async` with virtual threads, and enable `spring.threads.virtual.enabled=true` in application.yml."

### Low Freedom (Comprehensive Implementation - 80 tokens)
"Implement structured concurrency with `StructuredTaskScope.ShutdownOnFailure()`, configure HikariCP pool size 50+ for virtual threads, use GraalVM native image compilation with `@RegisterReflectionForBinding`, and monitor with Micrometer metrics for thread utilization."

---

## 🎯 Works Well With

### Core MoAI Skills
- `Skill("moai-domain-backend")` - Backend architecture patterns
- `Skill("moai-domain-security")` - Enterprise security patterns
- `Skill("moai-foundation-trust")` - TRUST 5 compliance

### Enterprise Technologies
- **Databases**: PostgreSQL, MySQL, MongoDB, Redis
- **Message Queues**: RabbitMQ, Apache Kafka, ActiveMQ
- **Caching**: Redis, Hazelcast, Caffeine
- **API Gateway**: Spring Cloud Gateway, Kong
- **Service Mesh**: Istio, Linkerd

---

## 🚀 Production Deployment

### Docker Configuration

```dockerfile
# Dockerfile for GraalVM native image
FROM ghcr.io/graalvm/native-image-community:21
COPY target/enterprise-service executable
RUN chmod +x executable

EXPOSE 8080
ENTRYPOINT ["./executable"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: enterprise-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: enterprise-service
  template:
    metadata:
      labels:
        app: enterprise-service
    spec:
      containers:
      - name: enterprise-service
        image: enterprise-service:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: SPRING_PROFILES_ACTIVE
          value: "production"
        - name: JAVA_OPTS
          value: "-XX:+UseVirtualThreads"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

## ✅ Quality Assurance Checklist

- [ ] **Java 21 LTS Features**: Virtual threads, structured concurrency, pattern matching
- [ ] **Spring Boot 3.5.3**: Latest security patches and performance improvements
- [ ] **Context7 Integration**: Real-time documentation from official sources
- [ ] **Virtual Thread Optimization**: Proper configuration and monitoring
- [ ] **GraalVM Native Image**: Reflection hints and production optimization
- [ ] **Security Configuration**: OAuth2, JWT, CSRF protection
- [ ] **Testing Coverage**: Unit, integration, and performance tests
- [ ] **Observability**: Metrics, tracing, and health checks
- [ ] **Production Readiness**: Docker, Kubernetes, CI/CD pipeline

---

**Last Updated**: 2025-11-06
**Version**: 3.0.0 (Premium Edition - Java 21 + Spring Boot 3.5.3)
**Context7 Integration**: Fully integrated with Java SE 21 and Spring Boot 3.5.3 official APIs
**Status**: Production Ready - Enterprise Grade