# Java CLI Reference

Quick reference for Java 24, JUnit 6.0.0, Maven 3.9.9, Gradle 8.10.2, Mockito 5.15.2, and Spring Boot 3.4.2.

---

## Tool Versions (2025-10-22)

| Tool | Version | Release Date | Purpose |
|------|---------|--------------|---------|
| **Java** | 24 | 2025-03-18 | Runtime & Language |
| **JUnit Jupiter** | 6.0.0 | 2025-02-15 | Testing Framework |
| **Maven** | 3.9.9 | 2024-09-19 | Build Tool |
| **Gradle** | 8.10.2 | 2024-09-05 | Build Tool |
| **Mockito** | 5.15.2 | 2025-09-20 | Mocking Framework |
| **Spring Boot** | 3.4.2 | 2025-01-15 | Application Framework |
| **Checkstyle** | 10.20.2 | 2025-09-10 | Code Quality |
| **SpotBugs** | 4.8.6 | 2024-08-15 | Static Analysis |

---

## Java 24

### Installation

```bash
# macOS (Homebrew)
brew install openjdk@24

# Ubuntu/Debian (via SDKMAN)
curl -s "https://get.sdkman.io" | bash
sdk install java 24-open

# Verify installation
java -version
# Expected: openjdk version "24" 2025-03-18

# Set JAVA_HOME
export JAVA_HOME=$(/usr/libexec/java_home -v 24)  # macOS
export JAVA_HOME=/usr/lib/jvm/java-24-openjdk     # Linux
```

### Key Java 24 Features

**1. Preview Features**
```java
// Stream Gatherers (JEP 485 - Preview)
import java.util.stream.Gatherers;

List<Integer> numbers = List.of(1, 2, 3, 4, 5);
List<Integer> windowed = numbers.stream()
    .gather(Gatherers.windowFixed(2))
    .map(window -> window.stream().mapToInt(Integer::intValue).sum())
    .toList();
// Result: [3, 7, 5]
```

**2. Virtual Threads (Stable)**
```java
// Virtual threads are now stable (no preview flag)
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    IntStream.range(0, 10_000).forEach(i -> {
        executor.submit(() -> {
            Thread.sleep(Duration.ofSeconds(1));
            return i;
        });
    });
}
```

**3. Pattern Matching Enhancements**
```java
// Record patterns with switch
record Point(int x, int y) {}

Object obj = new Point(10, 20);
String result = switch (obj) {
    case Point(int x, int y) -> "Point at (%d, %d)".formatted(x, y);
    case String s -> "String: " + s;
    case null -> "null value";
    default -> "Unknown type";
};
```

**4. Sequenced Collections (Stable)**
```java
// Sequenced collections provide consistent ordering operations
List<String> list = new ArrayList<>(List.of("a", "b", "c"));
String first = list.getFirst();  // "a"
String last = list.getLast();    // "c"
list.addFirst("z");              // ["z", "a", "b", "c"]
list.addLast("d");               // ["z", "a", "b", "c", "d"]
List<String> reversed = list.reversed();  // ["d", "c", "b", "a", "z"]
```

### Compiler Options

```bash
# Compile with Java 24 target
javac --release 24 MyClass.java

# Enable preview features (if needed)
javac --enable-preview --release 24 MyClass.java
java --enable-preview MyClass

# Generate verbose output
javac -verbose MyClass.java

# Specify source and target versions
javac -source 24 -target 24 MyClass.java
```

---

## JUnit 6.0.0 (Jupiter)

### Installation

#### Maven (`pom.xml`)
```xml
<dependencies>
    <dependency>
        <groupId>org.junit.jupiter</groupId>
        <artifactId>junit-jupiter</artifactId>
        <version>6.0.0</version>
        <scope>test</scope>
    </dependency>
</dependencies>

<build>
    <plugins>
        <plugin>
            <artifactId>maven-surefire-plugin</artifactId>
            <version>3.5.2</version>
        </plugin>
    </plugins>
</build>
```

#### Gradle (Kotlin DSL - `build.gradle.kts`)
```kotlin
dependencies {
    testImplementation("org.junit.jupiter:junit-jupiter:6.0.0")
}

tasks.test {
    useJUnitPlatform()
}
```

### Core Annotations

```java
import org.junit.jupiter.api.*;

class CalculatorTest {

    @BeforeAll
    static void initAll() {
        // Runs once before all tests (requires static)
    }

    @BeforeEach
    void init() {
        // Runs before each test method
    }

    @Test
    void addition() {
        assertEquals(2, 1 + 1);
    }

    @Test
    @DisplayName("Custom test name for better readability")
    void subtraction() {
        assertEquals(0, 1 - 1);
    }

    @Test
    @Disabled("Not implemented yet")
    void multiplication() {
        // Skipped test
    }

    @AfterEach
    void tearDown() {
        // Runs after each test method
    }

    @AfterAll
    static void tearDownAll() {
        // Runs once after all tests (requires static)
    }
}
```

### Assertions API

```java
import static org.junit.jupiter.api.Assertions.*;

class AssertionsDemo {

    @Test
    void standardAssertions() {
        assertEquals(2, 2);
        assertNotEquals(1, 2);
        assertTrue(true);
        assertFalse(false);
        assertNull(null);
        assertNotNull("value");
    }

    @Test
    void groupedAssertions() {
        // All assertions execute even if some fail
        assertAll("User properties",
            () -> assertEquals("John", user.getFirstName()),
            () -> assertEquals("Doe", user.getLastName()),
            () -> assertEquals(30, user.getAge())
        );
    }

    @Test
    void exceptionTesting() {
        Exception exception = assertThrows(
            IllegalArgumentException.class,
            () -> Integer.parseInt("invalid")
        );
        assertEquals("For input string: \"invalid\"", exception.getMessage());
    }

    @Test
    void timeoutTest() {
        assertTimeout(Duration.ofSeconds(1), () -> {
            // Code that should complete within 1 second
            Thread.sleep(500);
        });
    }

    @Test
    void lazyAssertionMessage() {
        // Message only computed if assertion fails
        assertTrue(isPrime(37), () -> "37 should be prime");
    }
}
```

### Parameterized Tests

```java
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.*;

class ParameterizedTestDemo {

    @ParameterizedTest
    @ValueSource(strings = {"racecar", "radar", "level"})
    void palindromes(String candidate) {
        assertTrue(isPalindrome(candidate));
    }

    @ParameterizedTest
    @CsvSource({
        "apple,  1",
        "banana, 2",
        "cherry, 3"
    })
    void testWithCsvSource(String fruit, int rank) {
        assertNotNull(fruit);
        assertTrue(rank > 0);
    }

    @ParameterizedTest
    @MethodSource("stringProvider")
    void testWithMethodSource(String argument) {
        assertNotNull(argument);
    }

    static Stream<String> stringProvider() {
        return Stream.of("apple", "banana", "cherry");
    }

    @ParameterizedTest
    @EnumSource(Month.class)
    void testWithEnumSource(Month month) {
        assertTrue(month.getValue() >= 1 && month.getValue() <= 12);
    }
}
```

### Nested Tests

```java
@DisplayName("A UserService")
class UserServiceTest {

    UserService service;

    @BeforeEach
    void createService() {
        service = new UserService();
    }

    @Nested
    @DisplayName("when new")
    class WhenNew {

        @Test
        @DisplayName("should be empty")
        void isEmpty() {
            assertTrue(service.isEmpty());
        }

        @Nested
        @DisplayName("after adding a user")
        class AfterAdding {

            @BeforeEach
            void addUser() {
                service.addUser(new User("Alice"));
            }

            @Test
            @DisplayName("should contain one user")
            void hasOneUser() {
                assertEquals(1, service.count());
            }
        }
    }
}
```

### Dynamic Tests

```java
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;

class DynamicTestDemo {

    @TestFactory
    Stream<DynamicTest> dynamicTestsFromStream() {
        return IntStream.iterate(0, n -> n + 2).limit(10)
            .mapToObj(n -> DynamicTest.dynamicTest("test" + n, () -> {
                assertTrue(n % 2 == 0);
            }));
    }
}
```

### Running Tests

```bash
# Maven
mvn test                           # Run all tests
mvn test -Dtest=ClassName          # Run specific test class
mvn test -Dtest=ClassName#method   # Run specific test method
mvn clean test                     # Clean and run tests

# Gradle
gradle test                        # Run all tests
gradle test --tests ClassName      # Run specific test class
gradle test --tests '*method*'     # Pattern matching
gradle clean test                  # Clean and run tests
```

---

## Maven 3.9.9

### Project Setup

#### Basic POM (`pom.xml`)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.example</groupId>
    <artifactId>my-app</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>24</maven.compiler.source>
        <maven.compiler.target>24</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>

    <dependencies>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>6.0.0</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.13.0</version>
            </plugin>
            <plugin>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>3.5.2</version>
            </plugin>
        </plugins>
    </build>
</project>
```

### Common Commands

```bash
# Build lifecycle
mvn clean                          # Clean target directory
mvn compile                        # Compile source code
mvn test                           # Run tests
mvn package                        # Create JAR/WAR
mvn install                        # Install to local repository
mvn deploy                         # Deploy to remote repository

# Dependency management
mvn dependency:tree                # Show dependency tree
mvn dependency:resolve             # Resolve all dependencies
mvn versions:display-dependency-updates  # Check for updates

# Project creation
mvn archetype:generate \
  -DgroupId=com.example \
  -DartifactId=my-app \
  -DarchetypeArtifactId=maven-archetype-quickstart \
  -DinteractiveMode=false

# Skip tests
mvn package -DskipTests            # Compile but skip test execution
mvn package -Dmaven.test.skip=true # Skip test compilation and execution

# Run with specific profile
mvn clean install -Pproduction
```

---

## Gradle 8.10.2 (Kotlin DSL)

### Project Setup

#### Build Script (`build.gradle.kts`)
```kotlin
plugins {
    java
    application
}

group = "com.example"
version = "1.0.0"

repositories {
    mavenCentral()
}

java {
    toolchain {
        languageVersion.set(JavaLanguageVersion.of(24))
    }
}

dependencies {
    implementation("com.google.guava:guava:33.4.0-jre")
    testImplementation("org.junit.jupiter:junit-jupiter:6.0.0")
}

tasks.test {
    useJUnitPlatform()
    testLogging {
        events("passed", "skipped", "failed")
    }
}

application {
    mainClass.set("com.example.App")
}
```

### Common Commands

```bash
# Build tasks
gradle build                       # Full build
gradle clean build                 # Clean and build
gradle assemble                    # Build without tests
gradle check                       # Run all checks including tests

# Testing
gradle test                        # Run tests
gradle test --tests ClassName      # Run specific test
gradle clean test                  # Clean and test

# Running application
gradle run                         # Run main class
gradle run --args="arg1 arg2"      # Run with arguments

# Dependency management
gradle dependencies                # Show dependency tree
gradle dependencyInsight --dependency junit  # Inspect specific dependency

# Project information
gradle projects                    # List all projects
gradle tasks                       # List available tasks
gradle properties                  # Show project properties

# Gradle wrapper
gradle wrapper --gradle-version 8.10.2  # Update wrapper version
./gradlew build                    # Use wrapper (recommended)
```

---

## Mockito 5.15.2

### Setup

```kotlin
// build.gradle.kts
dependencies {
    testImplementation("org.mockito:mockito-core:5.15.2")
    testImplementation("org.mockito:mockito-junit-jupiter:5.15.2")
}
```

### Basic Mocking

```java
import static org.mockito.Mockito.*;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository repository;

    @Test
    void testFindUser() {
        // Arrange
        User expectedUser = new User("Alice");
        when(repository.findById(1L)).thenReturn(Optional.of(expectedUser));

        UserService service = new UserService(repository);

        // Act
        User result = service.findUser(1L);

        // Assert
        assertEquals("Alice", result.getName());
        verify(repository).findById(1L);
    }

    @Test
    void testCreateUser() {
        UserService service = new UserService(repository);
        User newUser = new User("Bob");

        service.createUser(newUser);

        verify(repository).save(newUser);
    }

    @Test
    void testExceptionHandling() {
        when(repository.findById(anyLong()))
            .thenThrow(new RuntimeException("Database error"));

        UserService service = new UserService(repository);

        assertThrows(RuntimeException.class, () -> service.findUser(1L));
    }
}
```

---

## Spring Boot 3.4.2

### Project Creation

```bash
# Using Spring Initializr CLI
spring init \
  --dependencies=web,data-jpa,h2 \
  --build=gradle \
  --language=java \
  --java-version=24 \
  my-spring-app

# Or visit: https://start.spring.io
```

### Application Configuration

```yaml
# src/main/resources/application.yml
spring:
  application:
    name: my-app
  datasource:
    url: jdbc:h2:mem:testdb
    driver-class-name: org.h2.Driver
    username: sa
    password:
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true
```

### REST Controller Example

```java
@RestController
@RequestMapping("/api/users")
public class UserController {

    private final UserService userService;

    @Autowired
    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping
    public List<User> getAllUsers() {
        return userService.findAll();
    }

    @GetMapping("/{id}")
    public ResponseEntity<User> getUserById(@PathVariable Long id) {
        return userService.findById(id)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping
    public ResponseEntity<User> createUser(@Valid @RequestBody User user) {
        User created = userService.save(user);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }
}
```

---

## Code Quality Tools

### Checkstyle 10.20.2

```xml
<!-- Maven plugin -->
<plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-checkstyle-plugin</artifactId>
    <version>3.5.0</version>
    <configuration>
        <configLocation>google_checks.xml</configLocation>
    </configuration>
</plugin>
```

```bash
mvn checkstyle:check               # Run Checkstyle
gradle checkstyleMain              # Gradle Checkstyle
```

### SpotBugs 4.8.6

```kotlin
// build.gradle.kts
plugins {
    id("com.github.spotbugs") version "6.0.26"
}

spotbugs {
    effort.set(com.github.spotbugs.snom.Effort.MAX)
    reportLevel.set(com.github.spotbugs.snom.Confidence.LOW)
}
```

```bash
gradle spotbugsMain                # Run SpotBugs
```

---

## CI/CD with GitHub Actions

```yaml
# .github/workflows/java.yml
name: Java CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up JDK 24
      uses: actions/setup-java@v4
      with:
        java-version: '24'
        distribution: 'temurin'

    - name: Build with Maven
      run: mvn clean verify

    - name: Run tests
      run: mvn test

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
```

---

## TRUST 5 Principles for Java

### T - Test First (Coverage ≥85%)

```bash
# Maven with JaCoCo
mvn clean test jacoco:report
# Report: target/site/jacoco/index.html

# Gradle with JaCoCo
gradle test jacocoTestReport
# Report: build/reports/jacoco/test/html/index.html
```

### R - Readable

```bash
# Use Checkstyle
mvn checkstyle:check

# Format code consistently
# Use IDE formatters or google-java-format
```

### U - Unified (Type Safety)

```java
// Leverage Java 24 type system
record User(String name, int age) {
    // Compiler-enforced immutability and type safety
}
```

### S - Secured

```bash
# Run SpotBugs
gradle spotbugsMain

# Check dependencies for vulnerabilities
mvn dependency-check:check
```

### T - Trackable

```java
// @TAG integration in code
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: AuthServiceTest.java
public class AuthService {
    // Implementation
}
```

---

## Quick Reference Commands

```bash
# Java
java -version                      # Check Java version
javac MyClass.java                 # Compile single file
java MyClass                       # Run compiled class

# Maven
mvn clean install                  # Full build
mvn test                           # Run tests
mvn package                        # Create JAR

# Gradle
gradle build                       # Full build
gradle test                        # Run tests
gradle clean build                 # Clean build

# Testing
mvn test -Dtest=ClassName          # Maven specific test
gradle test --tests ClassName      # Gradle specific test

# Code Quality
mvn checkstyle:check               # Code style
gradle spotbugsMain                # Static analysis
```

---

**Version**: 1.0.0 (2025-10-22)
**Updated**: Latest tool versions verified 2025-10-22
**Framework**: MoAI-ADK Java Language Skill
**Status**: Production-ready
