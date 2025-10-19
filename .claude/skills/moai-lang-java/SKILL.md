---
name: moai-lang-java
description: Java best practices with JUnit, Maven/Gradle, Checkstyle, and Spring Boot patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Java Expert

## What it does

Provides Java-specific expertise for TDD development, including JUnit testing, Maven/Gradle build tools, Checkstyle linting, and Spring Boot patterns.

## When to use

- "Java 테스트 작성", "JUnit 사용법", "Spring Boot 패턴", "엔터프라이즈 애플리케이션", "마이크로서비스", "데이터 처리"
- "Spring", "Spring Data", "Hibernate", "Jakarta EE"
- "Maven", "Gradle", "Quarkus", "Vert.x", "Micronaut"
- Automatically invoked when working with Java projects
- Java SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **JUnit 5**: Unit testing with annotations (@Test, @BeforeEach)
- **Mockito**: Mocking framework for dependencies
- **AssertJ**: Fluent assertion library
- Test coverage ≥85% with JaCoCo

**Build Tools**:
- **Maven**: pom.xml, dependency management
- **Gradle**: build.gradle, Kotlin DSL support
- Multi-module project structures

**Code Quality**:
- **Checkstyle**: Java style checker (Google/Sun conventions)
- **PMD**: Static code analysis
- **SpotBugs**: Bug detection

**Spring Boot Patterns**:
- Dependency Injection (@Autowired, @Component)
- REST controllers (@RestController, @RequestMapping)
- Service layer separation (@Service, @Repository)
- Configuration properties (@ConfigurationProperties)

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Interfaces for abstraction
- Builder pattern for complex objects
- Exception handling with custom exceptions

## Modern Java (21+)

**Recommended Version**: Java 21+ (LTS), 17+ (previous LTS), 11+ (legacy support)

**Modern Features**:
- **Virtual Threads** (21): Lightweight threading (Project Loom)
- **Records** (16+): Immutable data carriers with pattern matching
- **Text blocks** (15+): Multi-line strings (""")
- **Sealed classes** (17+): Controlled inheritance
- **Pattern matching** (21): Enhanced switch expressions
- **String templates** (preview): Formatted string interpolation

**Version Check**:
```bash
java --version  # Check Java version
javac --version
```

## Package Management Commands

### Using Maven
```bash
# Initialize project
mvn archetype:generate -DgroupId=com.example -DartifactId=my-app

# Install dependencies
mvn dependency:resolve
mvn clean install

# Add dependency to pom.xml manually or use plugin
mvn dependency:copy-dependencies

# Build
mvn clean build
mvn clean package
mvn clean compile

# Test
mvn test
mvn verify
mvn test -Dtest=MyTestClass

# Check dependencies
mvn dependency:tree
mvn dependency:analyze

# Update dependencies
mvn versions:use-latest-releases
```

### Using Gradle
```bash
# Initialize project
gradle init --type java-application
gradle init --type java-library

# Build
gradle build
gradle buildDependents

# Test
gradle test
gradle test --tests=MyTestClass
gradle test --rerun-tasks

# Add dependencies (in build.gradle)
dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web:3.0.0'
    testImplementation 'junit:junit:4.13.2'
}

# Run
gradle run
gradle bootRun  # Spring Boot

# Check dependencies
gradle dependencies
gradle dependencyTree

# Format and lint
gradle spotlessApply  # Format
gradle checkstyleMain  # Lint
```

### Common Commands
```bash
# Compile
javac src/Main.java

# Run
java -cp target/classes com.example.Main
java -jar target/my-app.jar

# Generate documentation
javadoc -d docs src/**/*.java

# Package as JAR
jar -cfv my-app.jar -C target/classes .

# IDE reload
mvn eclipse:eclipse  # Eclipse
mvn idea:idea  # IntelliJ
```

## Examples

### Example 1: TDD with JUnit
User: "/alfred:2-build SERVICE-001"
Claude: (creates RED test with JUnit 5, GREEN implementation, REFACTOR with interfaces)

### Example 2: Build execution
User: "Maven 빌드 실행"
Claude: (runs mvn clean test and reports results)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Java-specific review)
- database-expert (JPA/Hibernate patterns)
