---
name: moai-library-mermaid/advanced-patterns
description: Enterprise Mermaid diagram patterns, architectures, and advanced techniques
---

# Advanced Mermaid Diagram Patterns (v6.0.0)

## Enterprise Architecture Patterns

### 1. Microservices Architecture with C4 Diagram

```mermaid
C4Context
 title System Context - Microservices Platform

 Person(user, "End User", "Uses the platform")
 System_Boundary(system, "Platform") {
 System(api_gateway, "API Gateway", "Routes requests")
 System(auth_service, "Auth Service", "JWT tokens")
 System(user_service, "User Service", "User management")
 System(order_service, "Order Service", "Order processing")
 System(payment_service, "Payment Service", "Payment handling")
 System(notification_service, "Notification Service", "Sends emails/SMS")
 }
 System_Ext(email_provider, "Email Provider", "SendGrid/AWS SES")
 System_Ext(payment_gateway, "Payment Gateway", "Stripe/PayPal")
 System_Ext(analytics_service, "Analytics", "Mixpanel/Segment")

 Rel(user, api_gateway, "Makes requests")
 Rel(api_gateway, auth_service, "Authenticates")
 Rel(api_gateway, user_service, "Manages users")
 Rel(api_gateway, order_service, "Places orders")
 Rel(order_service, payment_service, "Processes payment")
 Rel(payment_service, payment_gateway, "Payment processing")
 Rel(notification_service, email_provider, "Sends emails")
 Rel(order_service, analytics_service, "Tracks events")
```

### 2. Complex State Machine (Ecommerce Order Flow)

```mermaid
stateDiagram-v2
 [*] --> Created

 Created --> Validating: Validate items
 Created --> Cancelled: User cancels

 Validating --> PaymentPending: Items valid
 Validating --> Cancelled: Validation failed

 PaymentPending --> Processing: Payment successful
 PaymentPending --> PaymentFailed: Payment failed
 PaymentFailed --> Cancelled: User gives up
 PaymentFailed --> PaymentPending: Retry payment

 Processing --> Shipped: Order prepared
 Processing --> Cancelled: System error

 Shipped --> InTransit: Carrier picked up
 InTransit --> Delivered: Delivered to customer
 InTransit --> ReturnRequested: Customer requests return

 Delivered --> ReturnRequested: Within 30 days
 ReturnRequested --> ReturnInProgress: Return authorized
 ReturnInProgress --> Refunded: Return received

 Cancelled --> [*]
 Refunded --> [*]
 Delivered --> [*]
```

### 3. Multi-Tenant Architecture

```mermaid
graph TB
 LB["Load Balancer"]

 subgraph "Shared Infrastructure"
 Gateway["API Gateway<br/>(Tenant Router)"]
 Cache["Redis Cache<br/>(Shared)"]
 Queue["Message Queue<br/>(RabbitMQ/Kafka)"]
 end

 subgraph "Microservices"
 AuthSvc["Auth Service<br/>(Tenant Context)"]
 UserSvc["User Service<br/>(Isolated)"]
 DataSvc["Data Service<br/>(Tenant DB)"]
 end

 subgraph "Data Layer"
 SharedDB["Shared Database<br/>(System Tables)"]
 TenantDB1["Tenant-1 DB<br/>(Isolated)"]
 TenantDB2["Tenant-2 DB<br/>(Isolated)"]
 TenantDBN["Tenant-N DB<br/>(Isolated)"]
 end

 subgraph "Storage"
 S3Shared["S3 Bucket<br/>(Shared Prefix)"]
 FileStore["File Store<br/>(Tenant Partitioned)"]
 end

 LB --> Gateway
 Gateway --> Cache
 Gateway --> Queue
 Gateway --> AuthSvc
 Gateway --> UserSvc
 Gateway --> DataSvc

 AuthSvc --> SharedDB
 UserSvc --> SharedDB
 DataSvc --> TenantDB1
 DataSvc --> TenantDB2
 DataSvc --> TenantDBN

 DataSvc --> S3Shared
 DataSvc --> FileStore

 style SharedDB fill:#e1f5ff
 style TenantDB1 fill:#fff3e0
 style TenantDB2 fill:#fff3e0
 style TenantDBN fill:#fff3e0
```

## Integration Pattern Examples

### 1. Event-Driven Architecture

```mermaid
graph LR
 OrderSvc["Order Service<br/>(Publishes)"]
 PaymentSvc["Payment Service<br/>(Publishes)"]
 NotificationSvc["Notification Service<br/>(Consumes)"]
 AnalyticsSvc["Analytics Service<br/>(Consumes)"]
 InventorySvc["Inventory Service<br/>(Consumes)"]

 EventBus["Event Bus<br/>(Kafka/RabbitMQ)"]

 OrderSvc -->|order.created<br/>order.paid<br/>order.shipped| EventBus
 PaymentSvc -->|payment.succeeded<br/>payment.failed| EventBus
 EventBus -->|Subscribe| NotificationSvc
 EventBus -->|Subscribe| AnalyticsSvc
 EventBus -->|Subscribe| InventorySvc

 NotificationSvc -->|Email/SMS| Users["Users"]
 AnalyticsSvc -->|Store| DataWarehouse["Data Warehouse"]
 InventorySvc -->|Update| Inventory["Inventory DB"]

 style EventBus fill:#ffe082
 style OrderSvc fill:#b3e5fc
 style PaymentSvc fill:#b3e5fc
 style NotificationSvc fill:#c8e6c9
 style AnalyticsSvc fill:#c8e6c9
```

### 2. Data Pipeline Architecture

```mermaid
graph TB
 DataSources["Data Sources"]

 subgraph "Ingestion"
 Kafka["Kafka Cluster"]
 S3Raw["S3 Raw Layer<br/>(Bronze)"]
 end

 subgraph "Processing"
 Spark["Spark Jobs<br/>(ETL)"]
 DBT["DBT Models<br/>(Transformations)"]
 S3Processed["S3 Processed Layer<br/>(Silver)"]
 end

 subgraph "Analytics"
 Warehouse["Data Warehouse<br/>(Redshift/BigQuery)"]
 S3Analytics["S3 Analytics Layer<br/>(Gold)"]
 Catalog["Data Catalog<br/>(Collibra/Alation)"]
 end

 subgraph "Consumption"
 BI["BI Tools<br/>(Tableau/Looker)"]
 ML["ML Platform<br/>(MLflow)"]
 Reports["Reporting<br/>(Custom)"]
 end

 DataSources -->|Stream| Kafka
 Kafka -->|Batch| S3Raw
 S3Raw -->|Process| Spark
 Spark -->|Transform| DBT
 DBT -->|Load| S3Processed
 S3Processed -->|Aggregate| Warehouse
 Warehouse -->|Export| S3Analytics
 Warehouse --> Catalog

 S3Analytics -->|Dashboard| BI
 Warehouse -->|Training| ML
 Warehouse -->|Query| Reports

 style Kafka fill:#ffb3ba
 style S3Raw fill:#fff9c4
 style S3Processed fill:#fff9c4
 style S3Analytics fill:#fff9c4
 style Warehouse fill:#b3e5fc
```

## Advanced Sequence Patterns

### 1. API Request with Error Handling and Retry

```mermaid
sequenceDiagram
 participant Client as Client App
 participant LB as Load Balancer
 participant API as API Server
 participant Cache as Redis Cache
 participant DB as Database
 participant Logger as Log Service

 Client->>+LB: GET /api/users/123
 LB->>+API: Forward request
 API->>+Cache: Check cache

 alt Cache Hit
 Cache-->>-API: Return cached data
 API-->>-LB: 200 OK (cached)
 LB-->>-Client: Return data
 else Cache Miss
 Cache-->>API: Not found
 API->>+DB: Query database

 alt Database Success
 DB-->>-API: Return data
 API->>+Cache: Store in cache
 Cache-->>-API: OK
 API->>Logger: Log success
 API-->>-LB: 200 OK
 LB-->>-Client: Return data
 else Database Timeout (Retry)
 DB-->>API: Timeout
 API->>Logger: Log timeout
 API->>API: Wait 1s
 API->>+DB: Retry query
 DB-->>-API: Success (retry)
 API-->>-LB: 200 OK
 LB-->>-Client: Return data
 else Database Failure
 DB-->>API: Error
 API->>Logger: Log error
 API-->>-LB: 500 Error
 LB-->>-Client: Return error
 end
 end
```

### 2. OAuth2 Authorization Code Flow

```mermaid
sequenceDiagram
 participant User as User
 participant App as App<br/>(your app)
 participant AuthSvr as Auth Server<br/>(Google/GitHub)
 participant API as API Server

 User->>+App: Click "Login with Google"
 App->>+AuthSvr: Redirect to /authorize<br/>client_id=xxx&redirect_uri=callback
 AuthSvr->>User: Show login form
 User->>AuthSvr: Enter credentials
 AuthSvr-->>-App: Redirect with code=xyz

 App->>+AuthSvr: POST /token<br/>code=xyz&client_secret=*
 AuthSvr-->>-App: Return access_token

 App->>+API: GET /user<br/>Authorization: Bearer token
 API-->>-App: Return user profile

 App-->>-User: Logged in! Redirect home
```

## Performance Optimization Patterns

### 1. Database Query Optimization Flow

```mermaid
graph TB
 Query["Slow Query<br/>(detected)"]

 Analyze{Analysis}

 Query --> Analyze

 Analyze -->|No Index| AddIndex["Add Index<br/>(btree/hash)"]
 Analyze -->|N+1 Queries| BatchFetch["Implement Batch<br/>Fetching/Joins"]
 Analyze -->|Large Result| Paginate["Add Pagination<br/>(offset/cursor)"]
 Analyze -->|Complex Logic| Denormalize["Denormalize<br/>Columns/Views"]
 Analyze -->|Caching Miss| CacheAdd["Add Redis Cache<br/>(TTL-based)"]

 AddIndex --> Test["Run Benchmark"]
 BatchFetch --> Test
 Paginate --> Test
 Denormalize --> Test
 CacheAdd --> Test

 Test -->|Performance | Monitor["Monitor in<br/>Production"]
 Test -->|Performance | Rollback["Rollback &<br/>Try Different"]
 Rollback --> Analyze

 style Query fill:#ffcdd2
 style Test fill:#c8e6c9
 style Monitor fill:#b3e5fc
```

### 2. Frontend Performance Optimization

```mermaid
graph TB
 PerfIssue["Performance Issue<br/>(Lighthouse audit)"]

 Issue{Issue Type}

 PerfIssue --> Issue

 Issue -->|Large Bundle| CodeSplit["Code Splitting<br/>(Route-based)"]
 Issue -->|Large Images| ImageOpt["Image Optimization<br/>(WebP/AVIF)"]
 Issue -->|Render Blocking| Defer["Defer JS Loading<br/>(async/defer)"]
 Issue -->|DOM Heavy| Virtual["Virtual Scrolling<br/>(react-window)"]
 Issue -->|Unused CSS| PurgeCSS["PurgeCSS<br/>(remove unused)"]
 Issue -->|No Caching| ServiceWorker["Service Worker<br/>(offline support)"]

 CodeSplit --> Deploy
 ImageOpt --> Deploy
 Defer --> Deploy
 Virtual --> Deploy
 PurgeCSS --> Deploy
 ServiceWorker --> Deploy

 Deploy["Deploy &<br/>Measure"] --> LighthouseCheck{"Lighthouse<br/>Score ≥ 90?"}

 LighthouseCheck -->|Yes| Success["Performance<br/> Improved"]
 LighthouseCheck -->|No| Issue

 style PerfIssue fill:#ffcdd2
 style Deploy fill:#fff9c4
 style Success fill:#c8e6c9
```

## Testing and QA Patterns

### 1. Test Coverage and Quality Gates

```mermaid
graph TB
 Code["Code Committed"]

 Code --> UnitTest["Unit Tests<br/>(Jest, pytest)"]
 Code --> IntegrationTest["Integration Tests<br/>(Supertest, TestContainers)"]
 Code --> E2ETest["E2E Tests<br/>(Playwright, Cypress)"]
 Code --> SecurityScan["Security Scan<br/>(SonarQube, Snyk)"]
 Code --> Linting["Code Quality<br/>(ESLint, Pylint)"]

 UnitTest --> Coverage{"Coverage<br/>≥ 80%?"}
 IntegrationTest --> Performance{"Performance<br/>?"}
 E2ETest --> Functional{"Critical<br/>Flows ?"}
 SecurityScan --> Security{"No HIGH<br/>Vulns?"}
 Linting --> Quality{"Lint<br/>Pass?"}

 Coverage -->|No| Reject["Reject<br/>Pull Request"]
 Performance -->|No| Reject
 Functional -->|No| Reject
 Security -->|No| Reject
 Quality -->|No| Reject

 Coverage -->|Yes| Merge[" Merge to<br/>Main Branch"]
 Performance -->|Yes| Merge
 Functional -->|Yes| Merge
 Security -->|Yes| Merge
 Quality -->|Yes| Merge

 Merge --> Deploy["Deploy to<br/>Production"]
 Reject --> Feedback["Notify Developer<br/>for Fixes"]

 style Reject fill:#ffcdd2
 style Merge fill:#c8e6c9
 style Deploy fill:#b3e5fc
```

### 2. Incident Response Flow

```mermaid
graph TB
 Alert["Production Alert<br/>(High Error Rate)"]

 Alert --> Page["PagerDuty<br/>Page On-Call"]
 Page --> Investigation["Start Investigation<br/>(Logs/Metrics)"]

 Investigation --> Issue{Issue<br/>Identified?}

 Issue -->|Database| DBMitigation["Failover to<br/>Replica DB"]
 Issue -->|Service Down| Restart["Restart Service<br/>Instances"]
 Issue -->|Memory Leak| Deploy["Deploy Fix<br/>from main"]
 Issue -->|DDoS| Rate["Enable Rate<br/>Limiting"]

 DBMitigation --> Monitor["Monitor System"]
 Restart --> Monitor
 Deploy --> Monitor
 Rate --> Monitor

 Monitor --> Stable{"System<br/>Stable?"}

 Stable -->|No| EscalateIssue{"Multiple<br/>Attempts?"}
 EscalateIssue -->|Yes| Escalate["Escalate to<br/>Principal Engineer"]
 Escalate --> Investigation
 EscalateIssue -->|No| Investigation

 Stable -->|Yes| Postmortem["Start Postmortem<br/>(within 24h)"]

 Postmortem --> Document["Document Timeline<br/>& Root Cause"]
 Document --> ActionItems["Create Action Items<br/>for Prevention"]
 ActionItems --> Close["Close Incident"]

 style Alert fill:#ffcdd2
 style Stable fill:#fff9c4
 style Postmortem fill:#c8e6c9
 style Close fill:#b3e5fc
```

## Context7 Integration Examples

### 1. Mermaid Documentation Workflow

```mermaid
graph LR
 context7["Context7<br/>(Latest Docs)"]
 requirements["Business Requirements"]

 context7 -->|Latest Patterns| DesignReview["Design Review<br/>(Team)"]
 requirements -->|User Stories| DesignReview

 DesignReview -->|Approved| Architecture["Create Architecture<br/>Diagram (Mermaid)"]
 DesignReview -->|Feedback| Iterate["Iterate Design"]
 Iterate --> DesignReview

 Architecture -->|Visualize| Documentation["Update Documentation"]
 Documentation -->|Publish| KnowledgeBase["Knowledge Base<br/>(Wiki/Notion)"]

 KnowledgeBase -->|Reference| TeamOnboarding["Team Onboarding"]
 TeamOnboarding -->|Questions| context7

 style context7 fill:#e1f5ff
 style DesignReview fill:#fff9c4
 style Architecture fill:#c8e6c9
 style Documentation fill:#b3e5fc
 style KnowledgeBase fill:#f3e5f5
```

## Best Practices for Complex Diagrams

1. Diagram Size: Keep to <50 nodes for readability
2. Color Coding: Use consistent colors for system components (blue=services, green=databases, orange=external)
3. Labeling: Include technology names (Kafka, PostgreSQL, Redis)
4. Hierarchy: Use subgraphs/boundaries to show ownership/boundaries
5. Arrows: Use descriptive labels on arrows for protocols/data types
6. Version Control: Store diagrams in git with markdown files
7. Documentation: Add context and assumptions in comments above diagrams

---

Version: 6.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
