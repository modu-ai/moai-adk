---
name: moai-lang-kotlin
version: 3.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Kotlin 2.1+ enterprise development with KMP, coroutines, Flow, Spring Boot, Compose, and Context7 MCP integration.
keywords: ['kotlin', 'kotlin21', 'kmp', 'coroutines', 'flow', 'spring-boot', 'compose', 'enterprise', 'context7']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# 🚀 Kotlin 2.1+ Enterprise Development Premium Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-kotlin |
| **Version** | 3.0.0 (2025-11-06) - Premium Edition |
| **Allowed tools** | Read, Bash, Context7 MCP Integration |
| **Auto-load** | Kotlin projects, KMP applications, enterprise development |
| **Tier** | Premium Language |
| **Context7 Integration** | Kotlin 2.1 + KMP Official Docs |

---

## 🎯 What It Does

**Enterprise-grade Kotlin 2.1+ development** with Kotlin Multiplatform (KMP), advanced coroutines, Flow, Spring Boot, Compose, and production-ready cross-platform applications.

### Core Capabilities

**🏗️ Modern Kotlin 2.1+ Features**:
- **Kotlin Multiplatform (KMP)**: Shared code across JVM, Android, iOS, Web, Native
- **Advanced Coroutines**: Structured concurrency, SupervisorScope, Flow integration
- **Context Receivers**: Dependency injection without frameworks
- **Sealed Classes**: Type-safe state management and error handling
- **Value Classes**: Zero-overhead memory optimization

**⚡ Enterprise Ecosystem**:
- **Spring Boot 3.x**: Modern reactive web applications with Kotlin DSL
- **Jetpack Compose**: Declarative UI for Android, Desktop, Web
- **Ktor**: Asynchronous HTTP client and server framework
- **Exposing**: Type-safe SQL with compile-time verification
- **Arrow**: Functional programming with Arrow Core and Effects

**🔧 Cross-Platform Development**:
- **KMP Libraries**: SQLDelight, Koin, Ktor, Kermit
- **Build Systems**: Gradle 8.5+ with Kotlin DSL
- **Testing**: Kotest, MockK, Turbine for async testing
- **CI/CD**: GitHub Actions, TeamCity, GitLab CI

---

## 🌟 Enterprise Patterns & Best Practices

### Modern KMP Architecture with Shared Business Logic

```kotlin
// shared/src/commonMain/kotlin/com/example/data/OrderRepository.kt
interface OrderRepository {
    suspend fun getOrderById(id: String): Order?
    suspend fun createOrder(order: Order): Result<Order>
    suspend fun getOrdersByCustomerId(customerId: String): Flow<List<Order>>
}

// Android Implementation
class AndroidOrderRepository(
    private val database: AppDatabase
) : OrderRepository {

    override suspend fun getOrderById(id: String): Order? {
        return database.orderDao().getById(id)
    }

    override suspend fun createOrder(order: Order): Result<Order> {
        return try {
            val id = database.orderDao().insert(order.toEntity())
            order.copy(id = id.toString()).let { Result.success(it) }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getOrdersByCustomerId(customerId: String): Flow<List<Order>> {
        return database.orderDao().getByCustomerId(customerId).map { entities ->
            entities.map { it.toOrder() }
        }
    }
}

// JVM Backend Implementation
class JvmOrderRepository(
    private val dataSource: DataSource
) : OrderRepository {

    override suspend fun getOrderById(id: String): Order? {
        return dataSource.connection.use { connection ->
            connection.prepareStatement(
                "SELECT * FROM orders WHERE id = ?"
            ).use { stmt ->
                stmt.setString(1, id)
                stmt.executeQuery().use { rs ->
                    if (rs.next()) rs.toOrder() else null
                }
            }
        }
    }

    override suspend fun createOrder(order: Order): Result<Order> {
        return try {
            val generatedId = dataSource.connection.use { connection ->
                connection.prepareStatement(
                    "INSERT INTO orders (customer_id, total_amount, status, created_at) VALUES (?, ?, ?, ?) RETURNING id"
                ).use { stmt ->
                    stmt.setString(1, order.customerId)
                    stmt.setBigDecimal(2, order.totalAmount.toBigDecimal())
                    stmt.setString(3, order.status.name)
                    stmt.setTimestamp(4, Timestamp.from(order.createdAt.toInstant()))
                    stmt.executeQuery().use { rs ->
                        if (rs.next()) rs.getString("id") else throw SQLException("Failed to generate ID")
                    }
                }
            }
            order.copy(id = generatedId).let { Result.success(it) }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun getOrdersByCustomerId(customerId: String): Flow<List<Order>> = flow {
        val orders = mutableListOf<Order>()
        dataSource.connection.use { connection ->
            connection.prepareStatement(
                "SELECT * FROM orders WHERE customer_id = ? ORDER BY created_at DESC"
            ).use { stmt ->
                stmt.setString(1, customerId)
                stmt.executeQuery().use { rs ->
                    while (rs.next()) {
                        orders.add(rs.toOrder())
                    }
                }
            }
        }
        emit(orders)
    }
}
```

### Advanced Coroutines with Structured Concurrency

```kotlin
@OptIn(ExperimentalCoroutinesApi::class)
class OrderProcessingService(
    private val orderRepository: OrderRepository,
    private val paymentService: PaymentService,
    private val inventoryService: InventoryService,
    private val notificationService: NotificationService
) {

    suspend fun processOrderWithStructuredConcurrency(order: Order): ProcessedOrder {
        return coroutineScope {
            // Create supervisors for different failure domains
            val paymentSupervisor = SupervisorJob()
            val inventorySupervisor = SupervisorJob()
            val notificationSupervisor = SupervisorJob()

            try {
                val paymentResult = async(paymentSupervisor) {
                    paymentService.processPayment(order.id, order.totalAmount)
                }

                val inventoryResult = async(inventorySupervisor) {
                    inventoryService.reserveItems(order.items)
                }

                // Wait for both operations with timeout
                val payment = withTimeout(30.seconds) { paymentResult.await() }
                val inventory = withTimeout(30.seconds) { inventoryResult.await() }

                // Validate results
                if (payment.isSuccess && inventory.isSuccess) {
                    val processedOrder = order.copy(
                        status = OrderStatus.COMPLETED,
                        paymentId = payment.getOrThrow().transactionId,
                        processedAt = Clock.System.now()
                    )

                    // Fire-and-forget notification
                    launch(notificationSupervisor) {
                        notificationService.sendOrderConfirmation(processedOrder)
                    }

                    processedOrder
                } else {
                    // Handle partial failures
                    handlePartialFailure(order, payment, inventory)
                }
            } finally {
                // Clean up supervisors
                paymentSupervisor.cancel()
                inventorySupervisor.cancel()
                notificationSupervisor.cancel()
            }
        }
    }

    private suspend fun handlePartialFailure(
        order: Order,
        payment: Result<PaymentResult>,
        inventory: Result<InventoryResult>
    ): ProcessedOrder {
        // Compensation logic for partial failures
        payment.onSuccess { paymentResult ->
            if (paymentResult.status == PaymentStatus.PROCESSING) {
                paymentService.refundPayment(paymentResult.transactionId)
            }
        }

        inventory.onSuccess { inventoryResult ->
            inventoryResult.reservedItems.forEach { (productId, quantity) ->
                inventoryService.releaseReservation(productId, quantity)
            }
        }

        return order.copy(
            status = OrderStatus.FAILED,
            failureReason = "Partial processing failure",
            processedAt = Clock.System.now()
        )
    }
}
```

### Modern Flow-Based Reactive Data Streams

```kotlin
class OrderStreamProcessor(
    private val orderRepository: OrderRepository,
    private val analyticsService: AnalyticsService
) {

    // Hot Flow that emits order events continuously
    private val _orderEvents = MutableSharedFlow<OrderEvent>(
        replay = 0,
        extraBufferCapacity = Channel.CONFLATED
    )
    val orderEvents: SharedFlow<OrderEvent> = _orderEvents

    // Cold Flow that transforms order data
    fun getProcessedOrdersStream(): Flow<ProcessedOrderData> =
        orderRepository.getOrders()
            .mapLatest { orders ->
                orders.map { order ->
                    ProcessedOrderData(
                        order = order,
                        analytics = analyticsService.getOrderAnalytics(order.id),
                        recommendations = generateRecommendations(order)
                    )
                }
            }
            .distinctUntilChangedBy { it.order.id }
            .debounce(100.milliseconds)

    // Compose multiple flows with advanced operators
    fun getRealtimeOrderInsights(): Flow<OrderInsights> =
        combine(
            getProcessedOrdersStream(),
            orderEvents,
            clock.flow { it.instant() }
        ) { orders, events, currentTime ->
            OrderInsights(
                totalOrders = orders.size,
                averageOrderValue = orders.map { it.order.totalAmount }.average(),
                activeOrders = events.count { it.eventType == EventType.ORDER_CREATED },
                currentTime = currentTime
            )
        }
        .sample(1.seconds)
        .catch { exception ->
            logger.error("Error in order insights stream", exception)
            emit(OrderInsights.empty)
        }

    fun emitOrderEvent(event: OrderEvent) {
        _orderEvents.tryEmit(event)
    }

    private suspend fun generateRecommendations(order: Order): List<String> {
        return buildList {
            if (order.totalAmount > 1000.0) {
                add("Consider premium shipping")
            }
            if (order.items.any { it.category == "electronics" }) {
                add("Add warranty options")
            }
            if (order.customerLoyaltyTier == LoyaltyTier.GOLD) {
                add("Apply discount code GOLD2025")
            }
        }
    }
}
```

### Spring Boot 3.x with Kotlin Configuration

```kotlin
// Application.kt
@SpringBootApplication
@EnableConfigurationProperties
class OrderApplication

fun main(args: Array<String>) {
    runApplication<OrderApplication>(*args)
}

// Configuration classes with type safety
@ConfigurationProperties(prefix = "app.orders")
data class OrderProperties(
    val batchSize: Int = 100,
    val processingTimeout: Duration = Duration.ofSeconds(30),
    val retryAttempts: Int = 3,
    val maxConcurrentOrders: Int = 50
)

// Primary constructor for dependency injection
@Service
class OrderService(
    private val orderRepository: OrderRepository,
    private val paymentService: PaymentService,
    private val inventoryService: InventoryService,
    private val orderProperties: OrderProperties,
    private val meterRegistry: MeterRegistry,
    private val applicationEventPublisher: ApplicationEventPublisher
) {

    private val orderProcessingTimer = Timer.Sample.start(meterRegistry, "orders.processing.time")

    @Transactional(isolation = IsolationLevel.READ_COMMITTED)
    suspend fun createOrder(request: CreateOrderRequest): OrderResult {
        return orderProcessingTimer.recordCallable {
            try {
                // Validate request
                val validation = validateOrderRequest(request)
                if (!validation.isValid) {
                    return OrderResult.failure(validation.errors)
                }

                // Create order entity
                val order = request.toOrder()
                val createdOrder = orderRepository.save(order)

                // Publish event for async processing
                applicationEventPublisher.publishEvent(OrderCreatedEvent(createdOrder.id))

                // Start async processing
                asyncProcessOrder(createdOrder.id)

                OrderResult.success(createdOrder)
            } catch (e: Exception) {
                meterRegistry.counter("orders.creation.error").increment()
                throw OrderProcessingException("Failed to create order", e)
            }
        }
    }

    private fun asyncProcessOrder(orderId: UUID) {
        GlobalScope.launch {
            try {
                processOrderAsync(orderId)
            } catch (e: Exception) {
                logger.error("Async order processing failed for order: $orderId", e)
                applicationEventPublisher.publishEvent(OrderProcessingFailedEvent(orderId, e.message))
            }
        }
    }

    private suspend fun processOrderAsync(orderId: UUID) {
        val order = orderRepository.findById(orderId)
            ?: throw OrderNotFoundException("Order not found: $orderId")

        // Payment processing with retry logic
        repeat(orderProperties.retryAttempts) { attempt ->
            try {
                val paymentResult = paymentService.processPayment(
                    orderId = orderId,
                    amount = order.totalAmount,
                    customerId = order.customerId
                )

                if (paymentResult.isSuccess) {
                    // Update order status
                    orderRepository.save(order.copy(
                        status = OrderStatus.PAID,
                        paymentId = paymentResult.getOrThrow().transactionId
                    ))

                    // Trigger inventory allocation
                    inventoryService.allocateItems(order.items)
                    return
                }
            } catch (e: PaymentException) {
                if (attempt == orderProperties.retryAttempts - 1) {
                    throw e
                }
                delay(Duration.ofSeconds(2).multipliedBy(attempt + 1L))
            }
        }
    }

    private fun validateOrderRequest(request: CreateOrderRequest): ValidationResult {
        val errors = mutableListOf<String>()

        if (request.customerId.isBlank()) {
            errors.add("Customer ID is required")
        }

        if (request.items.isEmpty()) {
            errors.add("Order must contain at least one item")
        }

        if (request.items.any { it.quantity <= 0 }) {
            errors.add("Item quantities must be positive")
        }

        return if (errors.isEmpty()) {
            ValidationResult.success
        } else {
            ValidationResult.failure(errors)
        }
    }
}
```

### Jetpack Compose Modern UI with State Management

```kotlin
@Composable
fun OrderScreen(
    viewModel: OrderViewModel = viewModel(),
    onOrderClick: (String) -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    val scaffoldState = rememberScaffoldState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Orders") },
                actions = {
                    IconButton(onClick = { viewModel.refreshOrders() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.createNewOrder() }
            ) {
                Icon(Icons.Default.Add, contentDescription = "Create Order")
            }
        }
    ) { paddingValues ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            verticalArrangement = Arrangement.spacedBy(8.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            when (val state = uiState) {
                is OrderUiState.Loading -> {
                    items(3) {
                        OrderCardSkeleton(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp)
                        )
                    }
                }
                is OrderUiState.Success -> {
                    items(state.orders) { order ->
                        OrderCard(
                            order = order,
                            onClick = { onOrderClick(order.id) },
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 16.dp)
                        )
                    }
                }
                is OrderUiState.Error -> {
                    item {
                        ErrorMessage(
                            message = state.message,
                            onRetry = { viewModel.refreshOrders() },
                            modifier = Modifier.fillMaxWidth()
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun OrderCard(
    order: Order,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.clickable { onClick() },
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Order #${order.id.take(8)}",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
                StatusChip(
                    status = order.status,
                    modifier = Modifier.height(32.dp)
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = formatCurrency(order.totalAmount),
                style = MaterialTheme.typography.headlineSmall,
                color = MaterialTheme.colorScheme.primary
            )

            Text(
                text = "Created: ${order.createdAt.format(DateTimeFormatter.ofLocalizedDate(FormatStyle.MEDIUM))}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            if (order.items.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))

                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(order.items.take(3)) { item ->
                        ItemChip(item = item)
                    }
                    if (order.items.size > 3) {
                        Text(
                            text = "+${order.items.size - 3} more",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.primary
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun StatusChip(
    status: OrderStatus,
    modifier: Modifier = Modifier
) {
    val (color, label) = when (status) {
        OrderStatus.PENDING -> MaterialTheme.colorScheme.tertiary to "Pending"
        OrderStatus.PROCESSING -> MaterialTheme.colorScheme.primary to "Processing"
        OrderStatus.COMPLETED -> MaterialTheme.colorScheme.primary to "Completed"
        OrderStatus.FAILED -> MaterialTheme.colorScheme.error to "Failed"
        OrderStatus.CANCELLED -> MaterialTheme.colorScheme.secondary to "Cancelled"
    }

    AssistChip(
        onClick = { /* Handle status click if needed */ },
        label = { Text(label) },
        colors = AssistChipDefaults.assistChipColors(
            containerColor = color.copy(alpha = 0.12f),
            labelColor = color
        ),
        modifier = modifier
    )
}
```

---

## 🔧 Modern Development Workflow

### Gradle 8.5+ with Kotlin DSL and KMP Configuration

```kotlin
// build.gradle.kts
plugins {
    kotlin("multiplatform") version "2.1.0"
    kotlin("plugin.spring") version "1.9.24"
    id("org.jetbrains.kotlin.plugin.jpa") version "1.9.24"
    id("org.springframework.boot") version "3.2.0"
    id("com.google.devtools.ksp") version "1.9.24-1.0.0"
    id("com.github.ben-manes.versions") version "0.51.0"
}

kotlin {
    jvmToolchain(21)

    sourceSets {
        val commonMain by getting
        val jvmMain by getting
        val androidMain by getting
    }

    androidTarget {
        compilations.all {
            kotlinOptions {
                jvmTarget = "17"
            }
        }
    }

    jvm {
        withJava()
        testRuns["test"] {
            executionTask.configure {
                useJUnitPlatform()
            }
        }
    }

    sourceSets {
        commonMain.dependencies {
            implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
            implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.3")
            implementation("io.arrow-kt:arrow-core:1.2.1")
        }

        jvmMain.dependencies {
            implementation("org.springframework.boot:spring-boot-starter-web")
            implementation("org.springframework.boot:spring-boot-starter-data-jpa")
            implementation("org.springframework.boot:spring-boot-starter-actuator")
            implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
            implementation("org.jetbrains.kotlinx:kotlinx-coroutines-reactor:1.8.0")
            implementation("com.fasterxml.jackson.module:jackson-module-kotlin")
        }

        androidMain.dependencies {
            implementation("androidx.core:core-ktx:1.12.0")
            implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.7.0")
            implementation("androidx.activity:activity-compose:1.8.2")
            implementation(platform("androidx.compose:compose-bom:2024.02.02"))
            implementation("androidx.compose.ui:ui")
            implementation("androidx.compose.ui:ui-tooling-preview")
            implementation("androidx.compose.material3:material3")
        }
    }
}

// Testing dependencies
dependencies {
    commonTestImplementation(kotlin("test"))
    commonTestImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.8.0")

    jvmTestImplementation("org.springframework.boot:spring-boot-starter-test")
    jvmTestImplementation("org.testcontainers:postgresql:1.19.7")
    jvmTestImplementation("io.kotest:kotest-runner-junit5:5.8.0")
    jvmTestImplementation("io.mockk:mockk:1.13.10")

    androidTestImplementation("androidx.compose.ui:ui-test-junit4")
    androidTestImplementation(platform("androidx.compose:compose-bom:2024.02.02"))
    androidTestImplementation("androidx.compose.ui:ui-test-manifest")
}
```

### Advanced Testing with Kotest and MockK

```kotlin
@SpringBootTest
@TestConstructorOrder(OrderService::class, OrderRepository::class)
class OrderServiceTest(
    private val orderService: OrderService,
    private val orderRepository: OrderRepository
) {

    @Test
    fun `should create order successfully when valid request provided`() {
        // Given
        val request = CreateOrderRequest(
            customerId = "customer-123",
            items = listOf(
                CreateOrderItemRequest(
                    productId = "product-456",
                    quantity = 2,
                    price = 29.99
                )
            )
        )

        every { orderRepository.save(any()) } returnsArgument { it }

        // When
        val result = orderService.createOrder(request)

        // Then
        result.isSuccess shouldBe true
        val order = result.getOrNull()
        order shouldNotBe null
        order!!.customerId shouldBe "customer-123"
        order.totalAmount shouldBe 59.98

        verify { orderRepository.save(any()) }
    }

    @Test
    fun `should throw OrderProcessingException when repository throws`() {
        // Given
        val request = CreateOrderRequest(
            customerId = "customer-123",
            items = emptyList()
        )

        every { orderRepository.save(any()) } throws RuntimeException("Database error")

        // When & Then
        shouldThrow<OrderProcessingException> {
            orderService.createOrder(request)
        }
    }

    @Test
    fun `should validate order request and return failure for invalid data`() {
        // Given
        val invalidRequest = CreateOrderRequest(
            customerId = "", // Invalid
            items = emptyList()
        )

        // When
        val result = orderService.createOrder(invalidRequest)

        // Then
        result.isFailure shouldBe true
        result.exceptionOrNull() should beInstanceOf(OrderValidationException::class)
    }
}

// Property-based testing with Kotest
class OrderValidationProperties : StringSpec({
    "Order ID validation" {
        listOf(
            "", "   ", "\t", "   \t   ", null
        ).forEach { invalidId ->
            shouldThrow<ValidationException> {
                validateOrderId(invalidId)
            }
        }

        listOf(
            "order-123", "ORDER-456", "123abc", "valid-order-id-789"
        ).forEach { validId ->
            shouldNotThrowAny {
                validateOrderId(validId)
            }
        }
    }

    "Order amount validation" {
        checkAll(
            { amount: Double -> amount > 0 },
            -1000.0, -0.01, 0.0
        )
    }

    "Email validation with regex" {
        val emailRegex = Regex("^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$")

        forAll(
            row("test@example.com", "user.name@domain.org", "test123@sub.domain.co.uk"),
            row("invalid-email", "@domain.com", "test@.com", "test@domain.")
        ) { email: String, valid: Boolean ->
            if (valid) {
                emailRegex shouldMatch email
            } else {
                emailRegex shouldNotMatch email
            }
        }
    }
})
```

---

## 📊 Performance Optimization Strategies

### Kotlin Coroutines Performance Tuning

```kotlin
@OptIn(ExperimentalCoroutinesApi::class)
class HighPerformanceOrderProcessor(
    private val dispatcher: CoroutineDispatcher = Dispatchers.Default.limitedParallelism(100)
) {

    // Use limited parallelism dispatcher
    private val processingScope = CoroutineScope(SupervisorJob() + dispatcher)

    suspend fun processBatchOrders(orders: List<Order>): BatchProcessingResult {
        return processingScope.async {
            val startTime = System.nanoTime()

            // Channel for backpressure control
            val processingChannel = Channel<Order>(Channel.UNLIMITED)

            // Launch consumers
            val consumers = (1..10).map {
                launch {
                    for (order in processingChannel) {
                        try {
                            processSingleOrder(order)
                        } catch (e: Exception) {
                            // Log error but continue processing
                            logger.error("Error processing order ${order.id}", e)
                        }
                    }
                }
            }

            // Feed orders to channel
            launch {
                orders.forEach { order ->
                    processingChannel.sendBlocking(order)
                }
                processingChannel.close()
            }

            // Wait for all consumers to complete
            consumers.joinAll()

            val endTime = System.nanoTime()
            val processingTime = (endTime - startTime) / 1_000_000

            BatchProcessingResult(
                processedCount = orders.size,
                processingTimeMs = processingTime,
                throughput = orders.size.toDouble() / (processingTime / 1000.0)
            )
        }.await()
    }

    // Optimized single order processing
    private suspend fun processSingleOrder(order: Order) {
        withContext(Dispatchers.IO) {
            // Use value classes for memory optimization
            val orderData = OrderData(
                id = order.id,
                customerId = order.customerId,
                totalAmount = order.totalAmount,
                itemCount = order.items.size
            )

            // Batch database operations
            databaseService.updateOrderStatistics(orderData)

            // Avoid unnecessary object creation
            if (order.totalAmount > 1000.0) {
                priorityQueue.add(order.id)
            }
        }
    }
}

// Memory-optimized data structures
@JvmInlineValue
value class OrderData(
    val id: String,
    val customerId: String,
    val totalAmount: Double,
    val itemCount: Int
)

// Efficient queue for priority processing
class PriorityOrderQueue {
    private val highPriorityQueue = Channel<String>(Channel.UNLIMITED)
    private val normalPriorityQueue = Channel<String>(Channel.UNLIMITED)

    suspend fun add(orderId: String, isHighPriority: Boolean = false) {
        if (isHighPriority) {
            highPriorityQueue.send(orderId)
        } else {
            normalPriorityQueue.send(orderId)
        }
    }

    suspend fun next(): String? {
        return highPriorityQueue.tryReceive().getOrNull()
            ?: normalPriorityQueue.tryReceive().getOrNull()
    }
}
```

---

## 🔒 Security Best Practices

### Type-Safe Configuration with Sealed Classes

```kotlin
// Type-safe permission system
sealed class Permission {
    abstract val name: String

    data class ReadOrders(override val name: String = "orders:read") : Permission()
    data class CreateOrders(override val name: String = "orders:create") : Permission()
    data class DeleteOrders(override val name: String = "orders:delete") : Permission()
    data class AdminAccess(override val name: String = "admin:access") : Permission()
}

@Serializable
data class UserContext(
    val userId: String,
    val permissions: Set<Permission>,
    val roles: Set<UserRole>,
    val sessionId: String
) {
    fun hasPermission(permission: Permission): Boolean {
        return permission in permissions || UserRole.ADMIN in roles
    }
}

// Context receiver for dependency injection without frameworks
context(UserContext)
suspend fun createOrder(request: CreateOrderRequest): OrderResult {
    if (!hasPermission(Permission.CreateOrders)) {
        throw AuthorizationException("Insufficient permissions to create orders")
    }

    // User context is available through context receiver
    return orderService.createOrder(request, userId)
}
```

### Secure API Development with Ktor

```kotlin
object SecurityConfig {
    fun configureSecurity(application: Application) {
        application.install(ContentNegotiation) {
            register(ContentType.Application.Json)
            register(ContentType.Text.Plain)
            register(ContentType.Application.Xml)
        }

        application.install(CallLogging) {
            level = LogLevel.INFO
            filter { call ->
                call.request.path().startsWith("/api/")
            }
        }

        application.install(Authentication) {
            jwt {
                realm = "order-api"
                verifier = makeJwtVerifier()
                validate { credentials ->
                    UserIdPrincipal(credentials.payload.getClaim("sub")?.toString())
                }
            }
        }

        application.install(Authorization) {
            roleBasedAuthorization {
                role("admin") {
                    protect { method, path ->
                        path.startsWith("/api/admin/")
                    }
                }
                role("user") {
                    protect { method, path ->
                        path.startsWith("/api/orders/")
                    }
                }
            }
        }

        application.install(RateLimit) {
            register(RateLimitPath("/api/orders")) {
                rateLimiter = RateLimiter.global(100, 1.minutes)
            }
        }
    }

    private fun makeJwtVerifier(): JWTVerifier {
        return JWTVerifier(Algorithm.HS256, "your-secret-key")
    }
}

@Serializable
data class ApiError(
    val code: String,
    val message: String,
    val details: Map<String, Any> = emptyMap()
)
```

---

## 📈 Monitoring & Observability

### Micrometer Metrics Integration

```kotlin
@Component
class OrderMetrics(
    private val meterRegistry: MeterRegistry,
    private val applicationEventPublisher: ApplicationEventPublisher
) {

    private val orderCounter = Counter
        .builder("orders.created")
        .description("Total number of orders created")
        .register(meterRegistry)

    private val orderAmountSummary = DistributionSummary
        .builder("orders.amount")
        .description("Distribution of order amounts")
        .register(meterRegistry)

    private val orderProcessingTimer = Timer
        .builder("orders.processing.time")
        .description("Time taken to process orders")
        .register(meterRegistry)

    fun recordOrderCreated(order: Order) {
        orderCounter.increment()
        orderAmountSummary.record(order.totalAmount)

        applicationEventPublisher.publishEvent(
            OrderMetricsEvent(
                orderId = order.id,
                amount = order.totalAmount,
                itemCount = order.items.size
            )
        )
    }

    fun <T> recordProcessingTime(operation: String, block: suspend () -> T): T {
        val sample = Timer.start(meterRegistry, "orders.processing.time")
        val tags = Tags.of("operation", operation)

        return try {
            val result = block()
            sample.stop(Timer.Builder().tag(tags))
            result
        } catch (e: Exception) {
            sample.stop(Timer.Builder().tags(tags.and("status", "error")))
            throw e
        }
    }
}

// Structured logging with correlation IDs
class CorrelationLogger {
    companion object {
        private val logger = LoggerFactory.getLogger(CorrelationLogger::class.java)
    }

    data class LogContext(
        val correlationId: String,
        val userId: String?,
        val requestPath: String?
    )

    fun logInfo(message: String, context: LogContext, vararg keyValues: Any?) {
        logger.info(
            "[{}] {} {} {}",
            context.correlationId,
            context.userId?.let { "[user:$it]" } ?: "",
            context.requestPath?.let { "[path:$it]" } ?: "",
            message,
            *keyValues
        )
    }

    fun logError(message: String, context: LogContext, exception: Throwable? = null) {
        if (exception != null) {
            logger.error(
                "[{}] {} {} {}",
                context.correlationId,
                context.userId?.let { "[user:$it]" } ?: "",
                context.requestPath?.let { "[path:$it]" } ?: "",
                message,
                exception
            )
        } else {
            logger.error(
                "[{}] {} {} {}",
                context.correlationId,
                context.userId?.let { "[user:$it]" } ?: "",
                context.requestPath?.let { "[path:$it]" } ?: "",
                message
            )
        }
    }
}
```

---

## 🔄 Context7 MCP Integration

### Real-time Documentation Access

```kotlin
class KotlinDocumentationService(
    private val context7Client: Context7Client
) {
    companion object {
        private const val KOTLIN_API = "/jetbrains/kotlin"
        private const val KOTLIN_MULTIPATFORM = "/websites/jetbrains-help-kotlin-multiplatform-dev"
    }

    suspend fun getCoroutinesDocumentation(): String {
        return context7Client.getLibraryDocs(
            KOTLIN_API,
            "kotlinx.coroutines structured concurrency SupervisorScope Flow"
        )
    }

    suspend fun getKmpPatterns(): String {
        return context7Client.getLibraryDocs(
            KOTLIN_MULTIPPLATFORM,
            "Kotlin Multiplatform shared code android ios web"
        )
    }

    suspend fun getLatestBestPractice(feature: String): String {
        return context7Client.getLibraryDocs(
            KOTLIN_API,
            "$feature best practices examples"
        )
    }

    suspend fun getSpringBootKotlinIntegration(): String {
        return context7Client.getLibraryDocs(
            KOTLIN_API,
            "Spring Boot Kotlin DSL configuration primary constructor"
        )
    }
}
```

---

## 📚 Progressive Disclosure Examples

### High Freedom (Quick Answer - 15 tokens)
"Use Kotlin 2.1+ with coroutines and Flow for reactive cross-platform applications."

### Medium Freedom (Detailed Guidance - 35 tokens)
"Implement KMP with shared business logic, use structured concurrency with SupervisorScope, apply Spring Boot with Kotlin DSL, and use Compose for declarative UIs."

### Low Freedom (Comprehensive Implementation - 80 tokens)
"Configure KMP project with shared common module, implement reactive data streams with Flow, use structured concurrency with SupervisorScope for error isolation, integrate Spring Boot 3.x with Kotlin DSL, apply Compose Desktop for cross-platform UI, and add Micrometer for observability."

---

## 🎯 Works Well With

### Core MoAI Skills
- `Skill("moai-domain-backend")` - Backend architecture patterns
- `Skill("moai-domain-security")` - Enterprise security patterns
- `Skill("moai-foundation-trust")` - TRUST 5 compliance

### Kotlin Ecosystem
- **Build Tools**: Gradle 8.5+, Maven, Kotlin DSL
- **Frameworks**: Spring Boot, Ktor, Arrow, Exposed
- **UI**: Jetpack Compose, Compose Multiplatform, Swing
- **Testing**: Kotest, MockK, Turbine, Testcontainers

---

## 🚀 Production Deployment

### Docker Multi-stage Build for KMP

```dockerfile
# Build stage
FROM gradle:8.5-jdk21 AS build
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts ./
COPY gradlew gradlew.bat ./
COPY --chown=gradle:gradle /home/gradle/.gradle .
COPY . .

# Build JVM backend
RUN ./gradlew :backend:jvmJar

# Build Android
RUN ./gradlew :android:assembleRelease

# Production runtime
FROM openjdk:21-jre-slim
WORKDIR /app
COPY --from=build /app/backend/build/libs/*.jar ./
COPY --from=build /app/android/build/outputs/apk/release/*.apk ./

EXPOSE ["8080"]
CMD ["java", "-jar", "backend-1.0.0.jar"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kotlin-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kotlin-api
  template:
    metadata:
      labels:
        app: kotlin-api
    spec:
      containers:
      - name: kotlin-api
        image: kotlin-api:1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: SPRING_PROFILES_ACTIVE
          value: "production"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /actuator/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

---

## ✅ Quality Assurance Checklist

- [ ] **Kotlin 2.1+ Modern Features**: KMP, context receivers, sealed classes, value classes
- [ ] **Coroutines & Flow**: Structured concurrency, SupervisorScope, reactive streams
- [ ] **Context7 Integration**: Real-time documentation from JetBrains official sources
- [ ] **Performance Optimization**: Limited parallelism dispatcher, memory-efficient data structures
- [ ] **Type Safety**: Sealed classes, value classes, compile-time validation
- [ ] **Testing Coverage**: Kotest, MockK, Turbine, property-based testing
- [ ] **Cross-Platform**: Android, JVM, iOS, Web, Native targets
- [ ] **Production Readiness**: Docker, Kubernetes, observability
- [ ] **Security**: Type-safe permissions, JWT authentication, secure API design

---

**Last Updated**: 2025-11-06
**Version**: 3.0.0 (Premium Edition - Kotlin 2.1+ + KMP)
**Context7 Integration**: Fully integrated with Kotlin official APIs
**Status**: Production Ready - Enterprise Grade