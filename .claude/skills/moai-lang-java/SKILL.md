---
name: moai-lang-java
description: Java 21 LTS enterprise development with Spring Boot, JPA, reactive patterns
version: 1.0.0
modularized: true
allowed-tools:
  - Read
  - Bash
  - WebFetch
last_updated: 2025-11-22
compliance_score: 75
auto_trigger_keywords:
  - java
  - lang
  - typescript
category_tier: 1
---

## Quick Reference (30 seconds)

**Primary Focus**: Enterprise applications, Spring Boot, Android development  
**Best For**: Microservices, REST APIs, Android apps, large-scale systems  
**Key Libraries**: Spring Boot 3.x, Hibernate, Project Reactor  
**Auto-triggers**: Java development, Spring Boot, enterprise patterns

**Version Matrix (2025-11-22)**:
- Java: 21 LTS (stable), 23 (latest)
- Spring Boot: 3.2.x
- Hibernate: 6.4.x
- JUnit: 5.10.x

---

## What It Does

Enterprise Java development with Spring framework, JPA/Hibernate, and reactive programming for building scalable applications.

**Core Capabilities**:
- ✅ Spring Boot enterprise development
- ✅ JPA/Hibernate database access
- ✅ Reactive programming with Project Reactor
- ✅ RESTful API design
- ✅ Testing with JUnit 5

---

## When to Use

**Automatic Triggers**:
- Spring Boot development
- Enterprise application architecture
- Android development
- Microservices design

**Manual Invocation**:
```
Skill("moai-lang-java")
```

---

## Three-Level Learning Path

### Level 1: Fundamentals
**See examples.md for 10 practical examples**:
- Spring Boot REST API
- Streams and functional programming
- CompletableFuture async
- JPA and Hibernate
- Exception handling
- JUnit 5 testing

### Level 2: Advanced Patterns
**See modules/advanced-patterns.md**:
- Gang of Four design patterns
- Generics and bounded types
- Custom annotations
- Concurrency with ExecutorService
- Reactive streams with Reactor

### Level 3: Production Optimization
**See modules/optimization.md**:
- JVM tuning and GC
- Collections optimization
- Stream parallelization
- Database batch operations
- Caching strategies

---

## Best Practices

### DO ✅
```java
// Use Spring Boot for enterprise apps
@SpringBootApplication
public class App {
    public static void main(String[] args) {
        SpringApplication.run(App.class, args);
    }
}

// Use Optional for null safety
public Optional<User> findById(Long id) {
    return repository.findById(id);
}

// Use CompletableFuture for async
CompletableFuture<String> future = fetchDataAsync();
```

### DON'T ❌
```java
// DON'T use null returns
public User findById(Long id) {
    return null; // Use Optional instead
}

// DON'T ignore exceptions
try {
    riskyOperation();
} catch (Exception e) {
    // Empty catch block
}
```

---

## Tool Versions (2025-11-22)

**Core**:
- Java: 21 LTS (September 2023)
- Maven: 3.9.x / Gradle: 8.x

**Spring Ecosystem**:
- Spring Boot: 3.2.x
- Spring Framework: 6.1.x
- Spring Data JPA: 3.2.x

**Database**:
- Hibernate: 6.4.x
- Liquibase: 4.25.x

**Testing**:
- JUnit: 5.10.x
- Mockito: 5.x

---

## Installation & Setup

```bash
# Install Java
brew install openjdk@21

# Create Spring Boot project
curl https://start.spring.io/starter.zip \
  -d dependencies=web,data-jpa,h2 \
  -d bootVersion=3.2.0 \
  -o demo.zip

# Run
./mvnw spring-boot:run
```

---

## Works Well With

- `moai-domain-backend` - Backend patterns
- `moai-domain-database` - Database design
- `moai-essentials-perf` - Performance optimization

---

## Learn More

- **Examples**: [examples.md](examples.md) - 10 examples
- **Advanced**: [modules/advanced-patterns.md](modules/advanced-patterns.md)
- **Performance**: [modules/optimization.md](modules/optimization.md)

---

## Changelog

- **v3.0.0** (2025-11-22): Complete modularization
- **v2.0.0** (2025-11-11): Added metadata

---

## Context7 Integration

### Related Libraries
- [Spring Boot](/spring-projects/spring-boot)
- [Hibernate](/hibernate/hibernate-orm)
- [Jackson](/FasterXML/jackson)

### Official Documentation
- [Java Documentation](https://docs.oracle.com/en/java/)
- [Spring Boot](https://spring.io/projects/spring-boot)

---

**Version**: 3.0.0  
**Status**: Production Ready  
**Last Updated**: 2025-11-22