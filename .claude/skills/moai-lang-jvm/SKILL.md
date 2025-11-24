---
name: moai-lang-jvm
description: "JVM languages consolidated: Java 21 LTS, Kotlin, Scala for backend services and enterprise"
version: 1.0.0
modularized: false
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules: []
dependencies:
  - moai-foundation-trust
  - moai-domain-backend
deprecated: false
successor: null
category_tier: 4
auto_trigger_keywords:
  - java
  - kotlin
  - scala
  - jvm
  - enterprise
  - spring
  - backend
  - gradle
  - maven
  - loom
  - coroutines
agent_coverage:
  - backend-expert
  - tdd-implementer
  - quality-gate
context7_references:
  - java
  - kotlin
  - scala
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**JVM Languages Consolidated**

Unified JVM platform with Java 21 LTS, Kotlin, and Scala for enterprise backend services. Supports Spring Framework, virtual threads, reactive programming, and functional patterns.

**Core Languages**:
- **Java 21 LTS**: Classic enterprise with virtual threads (Project Loom), pattern matching, records
- **Kotlin**: Concise, null-safe, 100% interoperable with Java, coroutines for async
- **Scala**: Functional + OOP, pattern matching, lazy evaluation, Akka for distributed systems

**When to Use**:
- Enterprise backend services requiring stability and performance
- Microservices with Spring Boot or Quarkus
- Event-driven systems with Akka or Kafka
- Functional programming with Scala
- Building on existing JVM infrastructure

---

## Core Languages

### Java 21 LTS

Modern Java with virtual threads, pattern matching, and sealed classes.

```java
// Virtual threads (Project Loom) - lightweight concurrency
ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();
executor.submit(() -> System.out.println("Running on virtual thread"));

// Records (immutable data classes)
record Person(String name, int age) {}

// Pattern matching
Object obj = "Hello";
String message = switch(obj) {
    case String s -> "String: " + s;
    case Integer i -> "Integer: " + i;
    default -> "Unknown";
};
```

---

### Kotlin

Concise, null-safe language with first-class coroutines.

```kotlin
// Coroutines for async/await
suspend fun fetchUser(id: Int): User = withContext(Dispatchers.IO) {
    // Lightweight suspension, no threads
    val response = httpClient.get("/users/$id")
    response.body<User>()
}

// Data classes (auto equals, hashCode, toString)
data class User(val id: Int, val name: String, val email: String)

// Extension functions
fun String.isValidEmail(): Boolean = this.contains("@")
```

---

### Scala

Functional + OOP with pattern matching and lazy evaluation.

```scala
// Pattern matching
def describe(obj: Any): String = obj match {
  case s: String => s"String: $s"
  case i: Int => s"Integer: $i"
  case _ => "Unknown"
}

// Functional collections
val numbers = List(1, 2, 3, 4, 5)
  .filter(_ % 2 == 0)
  .map(_ * 2)
  .sum  // 12
```

---

## Best Practices

### ✅ DO
- Use Java 21 LTS for new projects
- Use virtual threads for high concurrency
- Leverage Kotlin for new code (100% Java compatible)
- Use Spring Boot for microservices
- Implement functional patterns where appropriate

### ❌ DON'T
- Use older Java versions (<17) for new projects
- Mix too many paradigms in same project
- Ignore null safety (use Optional, Kotlin's nullability)
- Create large god classes

---

## Related Skills

- `moai-domain-backend` (Spring Boot patterns)
- `moai-quality-testing` (JUnit, Spock testing)
- `moai-lang-python` (Alternative backend language)

---

## Consolidation

This skill consolidates:
- moai-lang-java (Java language patterns)
- moai-lang-kotlin (Kotlin language)
- moai-lang-scala (Scala language)

All functionality preserved in unified JVM skill.

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
