---
paths: "**/*.kt,**/*.kts,**/build.gradle.kts"
---

# Kotlin Development Guide


---


Kotlin 2.2+ Expert - K2 compiler, coroutines, Ktor, Compose Multiplatform.


Core Capabilities:

- Kotlin 2.2+: K2 compiler, coroutines, Flow, sealed classes, value classes
- Ktor 3.x (current 3.5): Async HTTP server/client, WebSocket, JWT authentication
- Exposed 0.55: Kotlin SQL framework with coroutines support
- Spring Boot (Kotlin): Kotlin-idiomatic Spring with WebFlux
- Compose Multiplatform: Desktop, iOS, Web, Android UI
- Testing: JUnit 5, MockK, Kotest, Turbine for Flow testing

---


### Kotlin 2.2+ Features

Coroutines and Flow: Use coroutineScope with async for parallel operations. Create deferred values with async, then call await on each to get results. Combine results into data classes. For reactive streams, create flow blocks with emit calls inside while loops. Use delay for intervals and flowOn to specify dispatcher.

Sealed Classes and Value Classes: Define sealed interface with generic type parameter. Create data class implementations for success and data object for stateless cases like Loading. Use @JvmInline annotation with value class wrapping a primitive. Add init blocks with require for validation.

### Ktor 3.x Server

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

## Documentation References

Library references for latest documentation (consult the upstream repos / official docs sites):

- `/ktorio/ktor` - Ktor 3.x server/client documentation
- `/jetbrains/exposed` - Exposed SQL framework
- `/JetBrains/kotlin` - Kotlin 2.2+ language reference
- `/Kotlin/kotlinx.coroutines` - Coroutines library
- `/jetbrains/compose-multiplatform` - Compose Multiplatform
- `/arrow-kt/arrow` - Arrow functional programming

Usage: Use WebSearch to find the official documentation site for a library, then WebFetch the relevant page to pull the up-to-date reference.

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


- `.claude/rules/moai/languages/java.md` - Java interoperability and Spring Boot patterns
- `moai-domain-backend` - REST API, GraphQL, microservices architecture
- `moai-domain-database` - JPA, Exposed, R2DBC patterns
- `moai-foundation-quality` + `moai-ref-testing-pyramid` - JUnit 5, MockK, TestContainers integration

---

## Troubleshooting

K2 Compiler: K2 is the default compiler in Kotlin 2.x — no opt-in flag is required (the legacy kotlin.experimental.tryK2 flag is obsolete). If the build still resolves to the old compiler, verify the Kotlin Gradle plugin version aligns with the Kotlin language version (2.2+). Clear the .gradle directory for a full rebuild.

Coroutines: Avoid runBlocking in suspend contexts. Use Dispatchers.IO for blocking operations.

Ktor: Ensure ContentNegotiation is installed. Check JWT verifier configuration. Verify routing hierarchy.

Exposed: Ensure all DB operations run within transaction context. Be aware of lazy entity loading outside transactions.

---
