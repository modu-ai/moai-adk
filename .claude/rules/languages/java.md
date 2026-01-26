---
paths:
  - "**/*.java"
  - "**/pom.xml"
  - "**/build.gradle"
  - "**/build.gradle.kts"
---

# Java Rules

Version: Java 21 LTS

## Tooling

- Build: Maven or Gradle
- Testing: JUnit 5 with coverage >= 85%
- Linting: SpotBugs, PMD, Error Prone

## Spring Boot 3.2+ Configuration

```properties
# Enable virtual threads
spring.threads.virtual.enabled=true
```

## Best Practices (2026)

- Use Virtual Threads for I/O-intensive workloads (reduces memory footprint)
- Use records for immutable data classes
- Use pattern matching in switch expressions
- Use Scoped Values instead of ThreadLocal (for virtual thread compatibility)
- Use Structured Concurrency for managing related task groups

## Virtual Threads Caveats

- Avoid `synchronized` blocks with blocking I/O (causes thread pinning)
- Watch for connection pool exhaustion - tune database pools
- Not all libraries are virtual-thread friendly yet
- Test thoroughly when migrating from platform threads

## Concurrency Patterns

```java
// Virtual thread executor
try (var executor = Executors.newVirtualThreadPerTaskExecutor()) {
    executor.submit(() -> processRequest(request));
}

// Structured concurrency (preview)
try (var scope = new StructuredTaskScope.ShutdownOnFailure()) {
    Future<User> user = scope.fork(() -> fetchUser(id));
    Future<Profile> profile = scope.fork(() -> fetchProfile(id));
    scope.join().throwIfFailed();
    return new Response(user.get(), profile.get());
}
```

## MoAI Integration

- Use Skill("moai-lang-java") for detailed patterns
- Follow TRUST 5 quality gates
