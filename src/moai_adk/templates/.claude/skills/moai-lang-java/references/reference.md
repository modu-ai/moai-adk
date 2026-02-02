# Java 21 LTS Reference Guide

## Language Features Overview

### Java 21 LTS (Long-Term Support)

Version Information:
- Latest: 21.0.5 (LTS, support until September 2031)
- Key JEPs: 444 (Virtual Threads), 441 (Pattern Matching), 440 (Record Patterns)
- JVM: HotSpot, GraalVM Native Image support

Core Features Summary:

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

---

## Enterprise Ecosystem

### Spring Boot 3.3

Key Features:
- Native GraalVM compilation support
- Virtual Threads integration
- Observability with Micrometer
- Problem Details for HTTP APIs (RFC 7807)
- SSL Bundle support

Configuration:
```properties
# application.properties
spring.threads.virtual.enabled=true
spring.datasource.url=jdbc:postgresql://localhost:5432/db
spring.datasource.hikari.maximum-pool-size=10
spring.jpa.hibernate.ddl-auto=validate
spring.jpa.show-sql=false
```

### Spring Security 6

Key Features:
- Lambda DSL configuration
- OAuth2 Resource Server
- JWT token validation
- Method security with annotations

### Hibernate 7 / Jakarta Persistence

Key Features:
- Java 21 record support
- Improved batch processing
- StatelessSession for bulk operations
- Native SQL result mapping

---

## Context7 Library Mappings

### Spring Ecosystem
```
/spring-projects/spring-boot - Spring Boot framework
/spring-projects/spring-framework - Spring Core framework
/spring-projects/spring-security - Security framework
/spring-projects/spring-data-jpa - JPA repositories
/spring-projects/spring-data-r2dbc - Reactive database access
```

### Persistence
```
/hibernate/hibernate-orm - Hibernate ORM
/querydsl/querydsl - Type-safe queries
/flyway/flyway - Database migrations
/liquibase/liquibase - Database change management
```

### Testing
```
/junit-team/junit5 - JUnit 5 testing
/mockito/mockito - Mocking framework
/testcontainers/testcontainers-java - Container testing
/assertj/assertj - Fluent assertions
```

### Build Tools
```
/gradle/gradle - Build automation
/apache/maven - Project management
```

### Utilities
```
/resilience4j/resilience4j - Fault tolerance
/open-telemetry/opentelemetry-java - Observability
/micrometer-metrics/micrometer - Application metrics
/mapstruct/mapstruct - Object mapping
```

---

## Performance Characteristics

### JVM Startup and Memory

| Runtime | Cold Start | Warm Start | Base Memory | With GraalVM Native |
|---------|------------|------------|-------------|---------------------|
| Java 21 (HotSpot) | 2-5s | <100ms | 256MB+ | 50-100ms, 64MB |
| Spring Boot 3.3 | 3-6s | <100ms | 512MB+ | 100-200ms, 128MB |

### Throughput Benchmarks

| Framework | Requests/sec | Latency P99 | Memory Usage |
|-----------|-------------|-------------|--------------|
| Spring Boot 3.3 (Virtual Threads) | 150K | 2ms | 512MB |
| Spring WebFlux | 180K | 1.5ms | 384MB |
| Spring MVC (Thread Pool) | 80K | 5ms | 768MB |

### Compilation Times

| Build Tool | Clean Build | Incremental | With Cache |
|------------|------------|-------------|------------|
| Maven 3.9 | 30-60s | 5-10s | 15-30s |
| Gradle 8.5 | 20-40s | 3-8s | 10-20s |

---

## Development Environment

### IDE Support

| IDE | Java Support | Spring Support | Best For |
|-----|--------------|----------------|----------|
| IntelliJ IDEA | Excellent | Excellent | Enterprise development |
| VS Code | Good | Good | Lightweight editing |
| Eclipse | Good | Good | Legacy projects |

### Recommended IntelliJ Plugins
- Spring Boot Assistant
- JPA Buddy
- TestContainers
- Key Promoter X
- Lombok

### Linters and Formatters

| Tool | Purpose | Config File |
|------|---------|-------------|
| Checkstyle | Code style | checkstyle.xml |
| SpotBugs | Bug detection | spotbugs-exclude.xml |
| PMD | Code analysis | ruleset.xml |
| google-java-format | Formatting | N/A (convention) |

---

## Container Optimization

### Docker Multi-Stage Build

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

### GraalVM Native Image

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
# Kubernetes deployment
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
```

---

## Migration Guide: Java 17 to 21

### Key Changes

1. Virtual Threads:
   - Replace `Executors.newFixedThreadPool()` with `Executors.newVirtualThreadPerTaskExecutor()`
   - Consider structured concurrency for complex concurrent tasks

2. Pattern Matching:
   - Refactor `instanceof` checks to pattern matching
   - Use guarded patterns with `when` clause

3. Record Patterns:
   - Use destructuring in switch expressions
   - Combine with sealed classes for exhaustive matching

4. Sequenced Collections:
   - Use `SequencedCollection.getFirst()` and `getLast()`
   - Replace iteration with sequence methods

### Gradle Configuration Update

```kotlin
// Before (Java 17)
java { toolchain { languageVersion = JavaLanguageVersion.of(17) } }

// After (Java 21)
java { toolchain { languageVersion = JavaLanguageVersion.of(21) } }
```

### Maven Configuration Update

```xml
<!-- Before -->
<java.version>17</java.version>

<!-- After -->
<java.version>21</java.version>
```

---

## Best Practices

### Code Style
- Use records for DTOs and value objects
- Prefer sealed interfaces for type hierarchies
- Use pattern matching in switch expressions
- Apply virtual threads for I/O-bound operations

### Spring Boot
- Use constructor injection (no @Autowired on constructors)
- Apply @Transactional at service layer
- Use records for request/response DTOs
- Configure proper connection pooling

### Testing
- Use JUnit 5 nested tests for organization
- Apply @DisplayName for readable test names
- Use TestContainers for integration tests
- Mock external dependencies with Mockito

### Security
- Never store passwords in plain text
- Use BCrypt for password hashing
- Validate all input data
- Apply principle of least privilege

---

Last Updated: 2025-12-07
Version: 1.0.0
