---
name: moai-lang-kotlin
description: >
  Kotlin 2.0+ development specialist covering Ktor, coroutines, Compose
  Multiplatform, and Kotlin-idiomatic patterns. Use when building Kotlin
  server apps, Android apps, or multiplatform projects.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "kotlin, ktor, coroutines, compose, android, multiplatform"
  context7-libraries: "/ktorio/ktor, /jetbrains/compose-multiplatform, /jetbrains/exposed"
  related-skills: "moai-lang-java, moai-lang-swift"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Kotlin", "Ktor", "coroutine", "Flow", "Compose", "Android", ".kt", ".kts", "build.gradle.kts"]
  languages: ["kotlin"]
---

## Quick Reference (30 seconds)

Kotlin 2.0+ Expert - K2 compiler, coroutines, Ktor, Compose Multiplatform with Context7 integration.

Auto-Triggers: Kotlin files (`.kt`, `.kts`), Gradle Kotlin DSL (`build.gradle.kts`, `settings.gradle.kts`)

Core Capabilities:

- Kotlin 2.0: K2 compiler, coroutines, Flow, sealed classes, value classes
- Ktor 3.0: Async HTTP server/client, WebSocket, JWT authentication
- Exposed 0.55: Kotlin SQL framework with coroutines support
- Spring Boot (Kotlin): Kotlin-idiomatic Spring with WebFlux
- Compose Multiplatform: Desktop, iOS, Web, Android UI
- Testing: JUnit 5, MockK, Kotest, Turbine for Flow testing

---

## Implementation Guide (5 minutes)

### Kotlin 2.0 Features

Coroutines and Flow: Use coroutineScope with async for parallel operations. Create deferred values with async, then call await on each to get results. Combine results into data classes. For reactive streams, create flow blocks with emit calls inside while loops. Use delay for intervals and flowOn to specify dispatcher.

Sealed Classes and Value Classes: Define sealed interface with generic type parameter. Create data class implementations for success and data object for stateless cases like Loading. Use @JvmInline annotation with value class wrapping a primitive. Add init blocks with require for validation.

### Ktor 3.0 Server

Application Setup: Call embeddedServer with Netty, port, and host parameters. Inside the lambda, call configuration functions for Koin, security, routing, and content negotiation. Call start with wait equals true.

For Koin configuration, install Koin plugin and define modules with single declarations for singletons. For security, install Authentication plugin and configure JWT with realm, verifier, and validate callback. For content negotiation, install ContentNegotiation with json configuration.

Routing with Authentication: Define routing function on Application. Inside routing block, use route for path prefixes. Create unauthenticated endpoints with post and call.receive for request body. Use authenticate block with verifier name for protected routes. Inside route blocks, define get endpoints with call.parameters for path/query params. Use call.respond with status code and response body.

### Exposed SQL Framework

Table and Entity: Define object extending LongIdTable with table name. Declare columns with varchar, enumerationByName, and timestamp functions. Use uniqueIndex() and defaultExpression for defaults.

Create entity class extending LongEntity with companion object extending LongEntityClass. Declare properties with by syntax using table column references. Create toModel function to map entity to domain model.

Repository with Coroutines: Create repository implementation taking Database parameter. Implement suspend functions wrapping Exposed operations in dbQuery helper. Use findById for single entity lookup. Use Entity.new for inserts. Define private dbQuery function using newSuspendedTransaction with IO dispatcher.

### Spring Boot with Kotlin

WebFlux Controller: Annotate class with @RestController and @RequestMapping. Create suspend functions for endpoints with @GetMapping and @PostMapping. Return Flow for collections using map to convert entities. Return ResponseEntity with status codes. Use @Valid for request validation.

---

## Advanced Patterns

### Compose Multiplatform

Shared UI Component: Create @Composable function taking ViewModel and callback parameters. Collect uiState as state with collectAsState. Use when expression on sealed state to show different composables for Loading, Success, and Error.

For list items, create Card composables with Modifier.fillMaxWidth and clickable. Use Row with padding, AsyncImage for avatars with CircleShape clip, and Column for text content with MaterialTheme.typography.

### Testing with MockK

Create test class with mockk for dependencies. Initialize service with mock in declaration. Use @Test with runTest for coroutine tests. Use coEvery with coAnswers for async mocking with delay. Use assertThat for assertions. For Flow testing, use toList to collect emissions and assert on size and content.

### Gradle Build Configuration

Use plugins block with kotlin("jvm") and kotlin("plugin.serialization") with version strings. Add id for ktor.plugin. Configure kotlin block with jvmToolchain. In dependencies block, add ktor server modules, kotlinx coroutines, exposed modules, and postgresql driver. Add test dependencies for mockk, coroutines-test, and turbine.

---

## Context7 Integration

Library mappings for latest documentation:

- `/ktorio/ktor` - Ktor 3.0 server/client documentation
- `/jetbrains/exposed` - Exposed SQL framework
- `/JetBrains/kotlin` - Kotlin 2.0 language reference
- `/Kotlin/kotlinx.coroutines` - Coroutines library
- `/jetbrains/compose-multiplatform` - Compose Multiplatform
- `/arrow-kt/arrow` - Arrow functional programming

Usage: Call mcp__context7__get_library_docs with context7CompatibleLibraryID, topic string for specific areas, and tokens parameter for response size.

---

## When to Use Kotlin

Use Kotlin When:

- Developing Android applications (official language)
- Building modern server applications with Ktor
- Preferring concise, expressive syntax
- Building reactive services with coroutines and Flow
- Creating multiplatform applications (iOS, Desktop, Web)
- Full Java interoperability required

Consider Alternatives When:

- Legacy Java codebase requiring minimal changes
- Big data pipelines (prefer Scala with Spark)

---

## Works Well With

- `moai-lang-java` - Java interoperability and Spring Boot patterns
- `moai-domain-backend` - REST API, GraphQL, microservices architecture
- `moai-domain-database` - JPA, Exposed, R2DBC patterns
- `moai-quality-testing` - JUnit 5, MockK, TestContainers integration
- `moai-infra-docker` - JVM container optimization

---

## Troubleshooting

K2 Compiler: Add kotlin.experimental.tryK2=true to gradle.properties. Clear .gradle directory for full rebuild.

Coroutines: Avoid runBlocking in suspend contexts. Use Dispatchers.IO for blocking operations.

Ktor: Ensure ContentNegotiation is installed. Check JWT verifier configuration. Verify routing hierarchy.

Exposed: Ensure all DB operations run within transaction context. Be aware of lazy entity loading outside transactions.

---

## Advanced Documentation

For comprehensive reference materials:

- [reference.md](reference.md) - Complete ecosystem, Context7 mappings, testing patterns, performance
- [examples.md](examples.md) - Production-ready code examples, Ktor, Compose, Android patterns

---

Last Updated: 2026-01-11
Status: Production Ready (v1.1.0)
