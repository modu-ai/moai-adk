---
name: moai-lang-java
description: >
  Java 21 LTS development specialist covering Spring Boot 3.3, virtual threads, pattern matching, and enterprise patterns. Use when building enterprise applications, microservices, or Spring projects.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "java, spring-boot, jpa, hibernate, virtual-threads, enterprise"
  context7-libraries: "/spring-projects/spring-boot, /spring-projects/spring-framework, /spring-projects/spring-security"
  related-skills: "moai-lang-kotlin, moai-domain-backend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Java", "Spring Boot", "Spring Framework", "JPA", "Hibernate", "Maven", "Gradle", ".java", "pom.xml", "build.gradle", "virtual thread"]
  languages: ["java"]
---

## Quick Reference (30 seconds)

Java 21 LTS Expert - Enterprise development with Spring Boot 3.3, Virtual Threads, and modern Java features.

Auto-Triggers: Java files with .java extension, build files including pom.xml, build.gradle, or build.gradle.kts

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

Virtual Threads with Project Loom:

Use try-with-resources on Executors.newVirtualThreadPerTaskExecutor. Call IntStream.range from 0 to 10000 and forEach to submit tasks that sleep for one second and return the iteration value.

Structured Concurrency Preview Pattern:

Use try-with-resources on new StructuredTaskScope.ShutdownOnFailure. Fork tasks for fetching user and orders by calling scope.fork with lambda expressions. Call scope.join then throwIfFailed. Return new composite object with results from both task suppliers.

Pattern Matching for Switch:

Create describe method taking Object parameter. Use switch expression with cases for Integer i with guard condition i greater than 0 returning positive integer message, Integer i returning non-positive message, String s returning length message, List with wildcard returning size message, null returning null value, and default returning unknown type.

Record Patterns and Sealed Classes:

Define Point record with int x and int y. Define Rectangle record with Point topLeft and Point bottomRight. Create area method that uses switch with Rectangle pattern destructuring both Point components into variables, returning absolute value of width times height. Define sealed Shape interface permitting Circle and Rectangle. Implement Circle record with area method using PI times radius squared.

### Spring Boot 3.3

REST Controller Pattern:

Create UserController with RestController annotation, RequestMapping for api/users, and RequiredArgsConstructor. Inject UserService. Create getUser method with GetMapping and PathVariable for id, returning ResponseEntity that maps findById result to ok or returns notFound. Create createUser method with PostMapping, Valid annotation, and RequestBody for CreateUserRequest. Create user, build URI location, return created response with body. Create deleteUser method with DeleteMapping that returns noContent or notFound based on service result.

Service Layer Pattern:

Create UserService with Service, RequiredArgsConstructor, and Transactional readOnly true annotations. Inject UserRepository and PasswordEncoder. Create findById method returning Optional. Create transactional create method that checks for duplicate email throwing DuplicateEmailException, builds User with builder pattern encoding password, and saves to repository.

### Spring Security 6

Security Configuration Pattern:

Create SecurityConfig with Configuration and EnableWebSecurity annotations. Define filterChain Bean taking HttpSecurity. Configure authorizeHttpRequests with permitAll for public API paths, hasRole ADMIN for admin paths, and authenticated for all other requests. Configure oauth2ResourceServer with jwt default. Set sessionManagement to STATELESS and disable csrf. Define passwordEncoder Bean returning BCryptPasswordEncoder.

### JPA/Hibernate Patterns

Entity Definition Pattern:

Create User entity with Entity and Table annotations. Add Lombok Getter, Setter, NoArgsConstructor, and Builder annotations. Define id with Id and GeneratedValue IDENTITY. Define name and email with Column nullable false, email also unique. Define status with Enumerated STRING. Define orders with OneToMany mappedBy user, cascade ALL, and orphanRemoval true.

Repository with Custom Queries Pattern:

Create UserRepository extending JpaRepository. Add findByEmail returning Optional. Add existsByEmail returning boolean. Add Query annotation for JPQL with LEFT JOIN FETCH for findByIdWithOrders using Param annotation. Add findByNameContainingIgnoreCase with Pageable for pagination.

DTOs as Records Pattern:

Create UserDto record with id, name, email, and status. Add static from factory method that constructs record from User entity. Create CreateUserRequest record with NotBlank and Size annotations for name, NotBlank and Email for email, NotBlank and Size min 8 for password.

---

## Advanced Patterns

### Virtual Threads Integration

Create AsyncUserService with Service and RequiredArgsConstructor annotations. Create fetchUserDetails method using StructuredTaskScope.ShutdownOnFailure in try-with-resources. Fork tasks for user and orders queries, join and throw if failed, return composite result. Create processUsersInParallel method using newVirtualThreadPerTaskExecutor and streaming user IDs to submit processing tasks.

### Build Configuration

Maven 3.9 Pattern:

Define project with parent for spring-boot-starter-parent version 3.3.0. Set java.version property to 21. Add dependencies for spring-boot-starter-web and spring-boot-starter-data-jpa.

Gradle 8.5 Kotlin DSL Pattern:

Apply plugins for org.springframework.boot, io.spring.dependency-management, and java. Set toolchain languageVersion to 21. Add implementation dependencies for web and data-jpa starters, testImplementation for test starter.

### Testing with JUnit 5

Unit Testing Pattern:

Create test class with ExtendWith MockitoExtension. Add Mock annotation for UserRepository. Add InjectMocks for UserService. Create shouldCreateUser test that stubs existsByEmail to return false and save to return user with id. Call service create and assertThat result id equals 1.

Integration Testing with TestContainers Pattern:

Create test class with Testcontainers and SpringBootTest annotations. Define static Container for PostgreSQL with postgres:16-alpine image. Add DynamicPropertySource to set datasource.url from container. Autowire repository. Create test that saves user and assertThat id isNotNull.

---

## Context7 Integration

Library mappings for latest documentation:

- spring-projects/spring-boot for Spring Boot 3.3 documentation
- spring-projects/spring-framework for Spring Framework core
- spring-projects/spring-security for Spring Security 6
- hibernate/hibernate-orm for Hibernate 7 ORM patterns
- junit-team/junit5 for JUnit 5 testing framework

---

## Works Well With

- moai-lang-kotlin for Kotlin interoperability and Spring Kotlin extensions
- moai-domain-backend for REST API, GraphQL, and microservices architecture
- moai-domain-database for JPA, Hibernate, and R2DBC patterns
- moai-foundation-quality for JUnit 5, Mockito, and TestContainers integration
- moai-infra-docker for JVM container optimization

---

## Troubleshooting

Common Issues:

- Version mismatch: Run java -version and check JAVA_HOME points to Java 21
- Compilation errors: Run mvn clean compile -X or gradle build --info
- Virtual thread issues: Ensure Java 21+ with --enable-preview if needed
- JPA lazy loading: Use Transactional annotation or JOIN FETCH queries

Performance Tips:

- Enable Virtual Threads by setting spring.threads.virtual.enabled to true
- Use GraalVM Native Image for faster startup
- Configure connection pooling with HikariCP

---

## Advanced Documentation

For comprehensive reference materials:

- reference.md for Java 21 features, Context7 mappings, and performance
- examples.md for production-ready Spring Boot examples

---

Last Updated: 2026-01-11
Status: Production Ready (v1.1.0)
