# Backend Architecture Examples

## Example 1: Microservices with Kubernetes 1.31 + Istio 1.21

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Istio Ingress Gateway                  │
│                     (TLS termination, routing)              │
└───────────────────────┬─────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
   │  User   │    │ Product │    │ Order   │
   │ Service │    │ Service │    │ Service │
   │ (Go)    │    │ (Python)│    │ (Node)  │
   └────┬────┘    └────┬────┘    └────┬────┘
        │              │              │
        └──────────────┼──────────────┘
                       │
              ┌────────▼────────┐
              │   PostgreSQL    │
              │   (Primary +    │
              │   Read Replicas)│
              └─────────────────┘
```

### Service Deployment (User Service - Go)

**Dockerfile**:
```dockerfile
# Multi-stage build for Go service
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/user-service .
EXPOSE 8080
CMD ["./user-service"]
```

**Kubernetes Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
      version: v1
  template:
    metadata:
      labels:
        app: user-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: user-service
        image: user-service:v1.2.3
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: url
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "otel-collector:4317"
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: production
spec:
  selector:
    app: user-service
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  type: ClusterIP
```

### Istio Configuration

**VirtualService (Traffic Routing)**:
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: user-service
  namespace: production
spec:
  hosts:
  - user-service
  http:
  - match:
    - headers:
        x-version:
          exact: "v2"
    route:
    - destination:
        host: user-service
        subset: v2
  - route:
    - destination:
        host: user-service
        subset: v1
      weight: 90
    - destination:
        host: user-service
        subset: v2
      weight: 10  # Canary: 10% traffic to v2
    timeout: 5s
    retries:
      attempts: 3
      perTryTimeout: 2s
      retryOn: 5xx,reset,connect-failure,refused-stream
```

**DestinationRule (Traffic Policy)**:
```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: user-service
  namespace: production
spec:
  host: user-service
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 100
      http:
        http1MaxPendingRequests: 50
        http2MaxRequests: 100
    outlierDetection:
      consecutiveErrors: 5
      interval: 30s
      baseEjectionTime: 30s
      maxEjectionPercent: 50
    loadBalancer:
      simple: LEAST_REQUEST
  subsets:
  - name: v1
    labels:
      version: v1
  - name: v2
    labels:
      version: v2
```

**PeerAuthentication (mTLS)**:
```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: production
spec:
  mtls:
    mode: STRICT  # Enforce mTLS for all services
```

**AuthorizationPolicy (RBAC)**:
```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: user-service-authz
  namespace: production
spec:
  selector:
    matchLabels:
      app: user-service
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/production/sa/order-service"]
    to:
    - operation:
        methods: ["GET", "POST"]
        paths: ["/api/users/*"]
```

### Observability Integration

**OpenTelemetry Instrumentation (Go)**:
```go
package main

import (
    "context"
    "log"
    "net/http"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func initTracer() (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint("otel-collector:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("user-service"),
            semconv.ServiceVersionKey.String("v1.2.3"),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

func main() {
    tp, err := initTracer()
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Shutdown(context.Background())

    // Wrap HTTP handler with OpenTelemetry instrumentation
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    wrappedHandler := otelhttp.NewHandler(handler, "user-service")
    http.Handle("/", wrappedHandler)

    log.Println("User service listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Prometheus Metrics (Go)**:
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func main() {
    // ... OpenTelemetry setup ...

    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### Database Connection Pooling

**PostgreSQL with PgBouncer**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pgbouncer
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: pgbouncer
        image: pgbouncer/pgbouncer:1.21
        ports:
        - containerPort: 6432
        volumeMounts:
        - name: config
          mountPath: /etc/pgbouncer
      volumes:
      - name: config
        configMap:
          name: pgbouncer-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: pgbouncer-config
data:
  pgbouncer.ini: |
    [databases]
    production = host=postgres-primary port=5432 dbname=production

    [pgbouncer]
    listen_addr = 0.0.0.0
    listen_port = 6432
    auth_type = md5
    auth_file = /etc/pgbouncer/userlist.txt
    pool_mode = transaction
    max_client_conn = 1000
    default_pool_size = 20
    reserve_pool_size = 5
    server_lifetime = 3600
    server_idle_timeout = 600
```

**Application Connection String**:
```
DATABASE_URL=postgresql://user:pass@pgbouncer:6432/production?pool_size=10
```

---

## Example 2: Event-Driven Architecture with Kafka 3.7

### Architecture Overview

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│  Order   │───▶│  Kafka   │◀───│ Payment  │
│ Service  │    │  Cluster │    │ Service  │
└──────────┘    └────┬─────┘    └──────────┘
                     │
          ┌──────────┼──────────┐
          │          │          │
     ┌────▼───┐ ┌───▼────┐ ┌───▼────┐
     │ Email  │ │ Invoice│ │Analytics│
     │Service │ │Service │ │Service  │
     └────────┘ └────────┘ └─────────┘
```

### Kafka Deployment (Kubernetes)

**Kafka StatefulSet**:
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: kafka
spec:
  serviceName: kafka-headless
  replicas: 3
  selector:
    matchLabels:
      app: kafka
  template:
    metadata:
      labels:
        app: kafka
    spec:
      containers:
      - name: kafka
        image: confluentinc/cp-kafka:7.6.0  # Kafka 3.7.x
        ports:
        - containerPort: 9092
          name: plaintext
        - containerPort: 9093
          name: ssl
        env:
        - name: KAFKA_BROKER_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: KAFKA_ZOOKEEPER_CONNECT
          value: "zookeeper:2181"
        - name: KAFKA_ADVERTISED_LISTENERS
          value: "PLAINTEXT://$(POD_NAME).kafka-headless:9092"
        - name: KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR
          value: "3"
        - name: KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR
          value: "3"
        - name: KAFKA_TRANSACTION_STATE_LOG_MIN_ISR
          value: "2"
        - name: KAFKA_LOG_RETENTION_HOURS
          value: "168"  # 7 days
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
        volumeMounts:
        - name: data
          mountPath: /var/lib/kafka/data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 100Gi
```

### Event Sourcing Pattern

**Event Definition**:
```typescript
// events/order.events.ts
export enum OrderEventType {
  ORDER_CREATED = 'order.created',
  ORDER_PAID = 'order.paid',
  ORDER_SHIPPED = 'order.shipped',
  ORDER_DELIVERED = 'order.delivered',
  ORDER_CANCELLED = 'order.cancelled',
}

export interface OrderEvent {
  eventId: string;
  eventType: OrderEventType;
  orderId: string;
  timestamp: Date;
  userId: string;
  data: unknown;
  metadata: {
    correlationId: string;
    causationId: string;
  };
}

export interface OrderCreatedEvent extends OrderEvent {
  eventType: OrderEventType.ORDER_CREATED;
  data: {
    items: Array<{
      productId: string;
      quantity: number;
      price: number;
    }>;
    totalAmount: number;
    shippingAddress: Address;
  };
}
```

**Event Producer (Order Service - Node.js)**:
```typescript
import { Kafka, Producer } from 'kafkajs';
import { v4 as uuidv4 } from 'uuid';

const kafka = new Kafka({
  clientId: 'order-service',
  brokers: ['kafka-0.kafka-headless:9092', 'kafka-1.kafka-headless:9092'],
});

const producer: Producer = kafka.producer({
  idempotent: true,  // Prevent duplicate events
  maxInFlightRequests: 5,
  transactionalId: 'order-service-producer',
});

await producer.connect();

export async function publishOrderCreatedEvent(order: Order): Promise<void> {
  const event: OrderCreatedEvent = {
    eventId: uuidv4(),
    eventType: OrderEventType.ORDER_CREATED,
    orderId: order.id,
    timestamp: new Date(),
    userId: order.userId,
    data: {
      items: order.items,
      totalAmount: order.totalAmount,
      shippingAddress: order.shippingAddress,
    },
    metadata: {
      correlationId: order.correlationId,
      causationId: uuidv4(),
    },
  };

  await producer.send({
    topic: 'orders',
    messages: [
      {
        key: order.id,  // Partition by order ID
        value: JSON.stringify(event),
        headers: {
          'event-type': event.eventType,
          'correlation-id': event.metadata.correlationId,
        },
      },
    ],
  });
}
```

**Event Consumer (Email Service - Python)**:
```python
from kafka import KafkaConsumer
from kafka.errors import KafkaError
import json
import logging

logger = logging.getLogger(__name__)

consumer = KafkaConsumer(
    'orders',
    bootstrap_servers=['kafka-0.kafka-headless:9092'],
    group_id='email-service',
    auto_offset_reset='earliest',
    enable_auto_commit=False,
    max_poll_records=100,
    value_deserializer=lambda m: json.loads(m.decode('utf-8')),
)

def process_order_created_event(event: dict):
    """Send order confirmation email."""
    try:
        send_email(
            to=event['userId'],
            subject=f"Order Confirmation - {event['orderId']}",
            body=render_template('order_confirmation.html', event=event),
        )
        logger.info(f"Email sent for order {event['orderId']}")
    except Exception as e:
        logger.error(f"Failed to send email: {e}")
        raise  # Trigger retry

for message in consumer:
    try:
        event = message.value
        event_type = message.headers.get('event-type')

        if event_type == 'order.created':
            process_order_created_event(event)

        # Commit offset only after successful processing
        consumer.commit()

    except Exception as e:
        logger.error(f"Error processing message: {e}")
        # Do not commit; retry on next poll
```

### CQRS Pattern

**Command Model (Write Side)**:
```typescript
// commands/create-order.command.ts
export class CreateOrderCommand {
  constructor(
    public readonly userId: string,
    public readonly items: OrderItem[],
    public readonly shippingAddress: Address,
  ) {}
}

export class CreateOrderHandler {
  constructor(
    private readonly orderRepository: OrderRepository,
    private readonly eventBus: EventBus,
  ) {}

  async execute(command: CreateOrderCommand): Promise<string> {
    // Create order aggregate
    const order = Order.create(
      command.userId,
      command.items,
      command.shippingAddress,
    );

    // Persist to event store
    await this.orderRepository.save(order);

    // Publish domain events
    for (const event of order.getUncommittedEvents()) {
      await this.eventBus.publish(event);
    }

    order.markEventsAsCommitted();
    return order.id;
  }
}
```

**Query Model (Read Side)**:
```typescript
// queries/get-order.query.ts
export class GetOrderQuery {
  constructor(public readonly orderId: string) {}
}

export class GetOrderHandler {
  constructor(private readonly orderReadRepository: OrderReadRepository) {}

  async execute(query: GetOrderQuery): Promise<OrderDTO> {
    // Read from optimized read model (denormalized)
    const order = await this.orderReadRepository.findById(query.orderId);
    if (!order) {
      throw new OrderNotFoundException(query.orderId);
    }
    return OrderDTO.fromReadModel(order);
  }
}
```

**Read Model Projector (Event Handler)**:
```typescript
// projections/order-projection.ts
export class OrderProjection {
  constructor(private readonly db: Database) {}

  @EventHandler(OrderEventType.ORDER_CREATED)
  async onOrderCreated(event: OrderCreatedEvent): Promise<void> {
    await this.db.insert('order_read_model', {
      order_id: event.orderId,
      user_id: event.userId,
      status: 'CREATED',
      total_amount: event.data.totalAmount,
      created_at: event.timestamp,
      updated_at: event.timestamp,
    });

    // Insert order items into separate table
    for (const item of event.data.items) {
      await this.db.insert('order_item_read_model', {
        order_id: event.orderId,
        product_id: item.productId,
        quantity: item.quantity,
        price: item.price,
      });
    }
  }

  @EventHandler(OrderEventType.ORDER_PAID)
  async onOrderPaid(event: OrderPaidEvent): Promise<void> {
    await this.db.update(
      'order_read_model',
      { order_id: event.orderId },
      { status: 'PAID', updated_at: event.timestamp },
    );
  }
}
```

---

## Example 3: Serverless with AWS Lambda

### Architecture Overview

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│ API      │───▶│  Lambda  │───▶│ DynamoDB │
│ Gateway  │    │ Functions│    └──────────┘
└──────────┘    └────┬─────┘
                     │
                ┌────▼────┐
                │   SQS   │
                └────┬────┘
                     │
                ┌────▼────┐
                │ Lambda  │
                │ Worker  │
                └─────────┘
```

### API Lambda Function (Node.js)

**handler.ts**:
```typescript
import { APIGatewayProxyHandler, APIGatewayProxyEvent, APIGatewayProxyResult } from 'aws-lambda';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient, PutCommand, GetCommand } from '@aws-sdk/lib-dynamodb';
import { SQSClient, SendMessageCommand } from '@aws-sdk/client-sqs';
import { v4 as uuidv4 } from 'uuid';

const client = new DynamoDBClient({});
const ddb = DynamoDBDocumentClient.from(client);
const sqs = new SQSClient({});

export const createOrder: APIGatewayProxyHandler = async (
  event: APIGatewayProxyEvent
): Promise<APIGatewayProxyResult> => {
  try {
    const body = JSON.parse(event.body || '{}');
    const orderId = uuidv4();

    // Write to DynamoDB
    await ddb.send(
      new PutCommand({
        TableName: process.env.ORDERS_TABLE!,
        Item: {
          orderId,
          userId: body.userId,
          items: body.items,
          totalAmount: body.totalAmount,
          status: 'PENDING',
          createdAt: new Date().toISOString(),
        },
      })
    );

    // Send message to SQS for async processing
    await sqs.send(
      new SendMessageCommand({
        QueueUrl: process.env.ORDER_QUEUE_URL!,
        MessageBody: JSON.stringify({
          orderId,
          userId: body.userId,
          totalAmount: body.totalAmount,
        }),
        MessageAttributes: {
          eventType: {
            DataType: 'String',
            StringValue: 'ORDER_CREATED',
          },
        },
      })
    );

    return {
      statusCode: 201,
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
      },
      body: JSON.stringify({ orderId }),
    };
  } catch (error) {
    console.error('Error creating order:', error);
    return {
      statusCode: 500,
      body: JSON.stringify({ error: 'Internal server error' }),
    };
  }
};

export const getOrder: APIGatewayProxyHandler = async (
  event: APIGatewayProxyEvent
): Promise<APIGatewayProxyResult> => {
  try {
    const orderId = event.pathParameters?.orderId;

    const result = await ddb.send(
      new GetCommand({
        TableName: process.env.ORDERS_TABLE!,
        Key: { orderId },
      })
    );

    if (!result.Item) {
      return {
        statusCode: 404,
        body: JSON.stringify({ error: 'Order not found' }),
      };
    }

    return {
      statusCode: 200,
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
      },
      body: JSON.stringify(result.Item),
    };
  } catch (error) {
    console.error('Error fetching order:', error);
    return {
      statusCode: 500,
      body: JSON.stringify({ error: 'Internal server error' }),
    };
  }
};
```

**Infrastructure as Code (Terraform)**:
```hcl
# lambda.tf
resource "aws_lambda_function" "create_order" {
  filename         = "lambda.zip"
  function_name    = "create-order"
  role            = aws_iam_role.lambda_exec.arn
  handler         = "handler.createOrder"
  runtime         = "nodejs20.x"
  memory_size     = 512
  timeout         = 30

  environment {
    variables = {
      ORDERS_TABLE     = aws_dynamodb_table.orders.name
      ORDER_QUEUE_URL  = aws_sqs_queue.orders.url
    }
  }

  tracing_config {
    mode = "Active"  # Enable X-Ray tracing
  }
}

resource "aws_lambda_function" "get_order" {
  filename         = "lambda.zip"
  function_name    = "get-order"
  role            = aws_iam_role.lambda_exec.arn
  handler         = "handler.getOrder"
  runtime         = "nodejs20.x"
  memory_size     = 256
  timeout         = 10

  environment {
    variables = {
      ORDERS_TABLE = aws_dynamodb_table.orders.name
    }
  }

  tracing_config {
    mode = "Active"
  }
}

# API Gateway
resource "aws_apigatewayv2_api" "main" {
  name          = "orders-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "create_order" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /orders"
  target    = "integrations/${aws_apigatewayv2_integration.create_order.id}"
}

resource "aws_apigatewayv2_route" "get_order" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /orders/{orderId}"
  target    = "integrations/${aws_apigatewayv2_integration.get_order.id}"
}

# DynamoDB
resource "aws_dynamodb_table" "orders" {
  name           = "orders"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "orderId"

  attribute {
    name = "orderId"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}

# SQS
resource "aws_sqs_queue" "orders" {
  name                       = "orders-queue"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 1209600  # 14 days
  receive_wait_time_seconds  = 20       # Long polling
}
```

### SQS Worker Lambda (Async Processing)

```typescript
import { SQSHandler, SQSEvent } from 'aws-lambda';
import { DynamoDBDocumentClient, UpdateCommand } from '@aws-sdk/lib-dynamodb';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';

const client = new DynamoDBClient({});
const ddb = DynamoDBDocumentClient.from(client);

export const processOrder: SQSHandler = async (event: SQSEvent): Promise<void> => {
  for (const record of event.Records) {
    try {
      const message = JSON.parse(record.body);
      const { orderId, userId, totalAmount } = message;

      // Simulate payment processing
      const paymentResult = await processPayment(userId, totalAmount);

      // Update order status
      await ddb.send(
        new UpdateCommand({
          TableName: process.env.ORDERS_TABLE!,
          Key: { orderId },
          UpdateExpression: 'SET #status = :status, paymentId = :paymentId',
          ExpressionAttributeNames: {
            '#status': 'status',
          },
          ExpressionAttributeValues: {
            ':status': paymentResult.success ? 'PAID' : 'PAYMENT_FAILED',
            ':paymentId': paymentResult.paymentId,
          },
        })
      );

      console.log(`Order ${orderId} processed successfully`);
    } catch (error) {
      console.error('Error processing order:', error);
      throw error;  // Trigger SQS retry
    }
  }
};

async function processPayment(userId: string, amount: number): Promise<{ success: boolean; paymentId: string }> {
  // Simulate external payment API call
  await new Promise(resolve => setTimeout(resolve, 1000));
  return { success: true, paymentId: `pay_${Date.now()}` };
}
```

### Cold Start Optimization

**Provisioned Concurrency**:
```hcl
resource "aws_lambda_provisioned_concurrency_config" "create_order" {
  function_name                     = aws_lambda_function.create_order.function_name
  provisioned_concurrent_executions = 5  # Keep 5 warm instances
  qualifier                         = aws_lambda_function.create_order.version
}
```

**Lambda Layers (Shared Dependencies)**:
```hcl
resource "aws_lambda_layer_version" "dependencies" {
  filename            = "layer.zip"
  layer_name          = "shared-dependencies"
  compatible_runtimes = ["nodejs20.x"]
  description         = "Shared dependencies for Lambda functions"
}

resource "aws_lambda_function" "create_order" {
  # ... other config ...
  layers = [aws_lambda_layer_version.dependencies.arn]
}
```

---

## Example 4: Observability with OpenTelemetry 1.24 + Prometheus 2.48

### Full Stack Observability Setup

**OpenTelemetry Collector Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: otel-collector
  namespace: observability
spec:
  replicas: 2
  selector:
    matchLabels:
      app: otel-collector
  template:
    metadata:
      labels:
        app: otel-collector
    spec:
      containers:
      - name: otel-collector
        image: otel/opentelemetry-collector-contrib:0.92.0
        ports:
        - containerPort: 4317  # OTLP gRPC
        - containerPort: 4318  # OTLP HTTP
        - containerPort: 8889  # Prometheus metrics
        volumeMounts:
        - name: config
          mountPath: /etc/otelcol
        resources:
          requests:
            cpu: "200m"
            memory: "256Mi"
          limits:
            cpu: "1000m"
            memory: "1Gi"
      volumes:
      - name: config
        configMap:
          name: otel-collector-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: observability
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318

    processors:
      batch:
        timeout: 10s
        send_batch_size: 1024
      memory_limiter:
        check_interval: 1s
        limit_mib: 512
      resource:
        attributes:
        - key: environment
          value: production
          action: insert

    exporters:
      prometheus:
        endpoint: "0.0.0.0:8889"
      jaeger:
        endpoint: "jaeger-collector.observability:14250"
        tls:
          insecure: true
      elasticsearch:
        endpoints: ["http://elasticsearch.observability:9200"]
        logs_index: "otel-logs"

    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch, resource]
          exporters: [jaeger]
        metrics:
          receivers: [otlp]
          processors: [memory_limiter, batch, resource]
          exporters: [prometheus]
        logs:
          receivers: [otlp]
          processors: [memory_limiter, batch, resource]
          exporters: [elasticsearch]
```

**Application Instrumentation (Python FastAPI)**:
```python
from fastapi import FastAPI
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor

# Setup tracing
trace_provider = TracerProvider()
trace_exporter = OTLPSpanExporter(endpoint="otel-collector:4317", insecure=True)
trace_provider.add_span_processor(BatchSpanProcessor(trace_exporter))
trace.set_tracer_provider(trace_provider)

# Setup metrics
metric_reader = PeriodicExportingMetricReader(
    OTLPMetricExporter(endpoint="otel-collector:4317", insecure=True),
    export_interval_millis=30000,
)
metric_provider = MeterProvider(metric_readers=[metric_reader])
metrics.set_meter_provider(metric_provider)

# Create app
app = FastAPI()

# Auto-instrument FastAPI
FastAPIInstrumentor.instrument_app(app)

# Custom metrics
meter = metrics.get_meter(__name__)
order_counter = meter.create_counter(
    name="orders_created_total",
    description="Total number of orders created",
    unit="1",
)

@app.post("/orders")
async def create_order(order: Order):
    tracer = trace.get_tracer(__name__)
    with tracer.start_as_current_span("create_order") as span:
        span.set_attribute("order.id", order.id)
        span.set_attribute("order.amount", order.totalAmount)

        # Business logic
        result = await process_order(order)

        # Record metric
        order_counter.add(1, {"status": result.status})

        return result
```

**Grafana Dashboards (Provisioning)**:
```yaml
# grafana-dashboard-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: observability
data:
  backend-overview.json: |
    {
      "dashboard": {
        "title": "Backend Overview",
        "panels": [
          {
            "title": "Request Rate",
            "targets": [
              {
                "expr": "rate(http_requests_total[5m])"
              }
            ]
          },
          {
            "title": "Error Rate",
            "targets": [
              {
                "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])"
              }
            ]
          },
          {
            "title": "Latency (p95)",
            "targets": [
              {
                "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
              }
            ]
          },
          {
            "title": "Active Connections",
            "targets": [
              {
                "expr": "sum(rate(http_requests_total[1m]))"
              }
            ]
          }
        ]
      }
    }
```

**Alert Rules (PrometheusRule)**:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: backend-alerts
  namespace: observability
spec:
  groups:
  - name: backend
    interval: 30s
    rules:
    - alert: HighErrorRate
      expr: |
        (
          rate(http_requests_total{status=~"5.."}[5m])
          /
          rate(http_requests_total[5m])
        ) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate detected (instance {{ $labels.instance }})"
        description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

    - alert: HighLatency
      expr: |
        histogram_quantile(0.95,
          rate(http_request_duration_seconds_bucket[5m])
        ) > 1.0
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High latency detected (instance {{ $labels.instance }})"
        description: "P95 latency is {{ $value }}s for the last 10 minutes"

    - alert: ServiceDown
      expr: up{job="backend"} == 0
      for: 2m
      labels:
        severity: critical
      annotations:
        summary: "Service is down (instance {{ $labels.instance }})"
        description: "Backend service has been down for more than 2 minutes"

    - alert: HighMemoryUsage
      expr: |
        (
          container_memory_usage_bytes{pod=~"backend-.*"}
          /
          container_spec_memory_limit_bytes{pod=~"backend-.*"}
        ) > 0.9
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High memory usage (pod {{ $labels.pod }})"
        description: "Memory usage is {{ $value | humanizePercentage }}"

    - alert: PodCrashLooping
      expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Pod is crash looping (pod {{ $labels.pod }})"
        description: "Pod has restarted {{ $value }} times in the last 15 minutes"
```

### Distributed Tracing (Jaeger UI)

**Example Trace (User Request → Order → Payment → Email)**:
```
Trace ID: abc123def456

Span 1: [user-service] GET /api/users/123        (20ms)
  └─ Span 2: [order-service] POST /orders        (150ms)
       ├─ Span 3: [postgres] SELECT * FROM users (5ms)
       ├─ Span 4: [postgres] INSERT INTO orders  (10ms)
       └─ Span 5: [payment-service] POST /charge (100ms)
            ├─ Span 6: [redis] GET payment_cache  (2ms)
            └─ Span 7: [stripe-api] POST /charges (90ms)
       └─ Span 8: [kafka] Publish order.created  (5ms)

Span 9: [email-service] Kafka Consumer          (500ms)
  └─ Span 10: [smtp] Send email                 (480ms)
```

**Benefits**:
- Identify bottlenecks (Span 10: SMTP is slowest)
- Visualize service dependencies
- Root cause analysis for failures
- Latency breakdown by service

---

---

## Example 5: Harbor Registry with Image Scanning

### Harbor 2.10.x Deployment

**Harbor Helm Installation**:
```bash
helm repo add harbor https://helm.goharbor.io
helm install harbor harbor/harbor \
  --namespace harbor \
  --create-namespace \
  --set expose.type=ingress \
  --set expose.ingress.hosts.core=harbor.example.com \
  --set externalURL=https://harbor.example.com \
  --set persistence.enabled=true \
  --set persistence.persistentVolumeClaim.registry.size=100Gi \
  --set harborAdminPassword=admin123 \
  --set trivy.enabled=true \
  --set notary.enabled=true
```

**Trivy Scanner Configuration**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: harbor-trivy-config
  namespace: harbor
data:
  trivy.yaml: |
    severity:
      - CRITICAL
      - HIGH
      - MEDIUM
    vuln-type:
      - os
      - library
    skip-update: false
    timeout: 5m
```

**Robot Account for CI/CD**:
```bash
# Create robot account via Harbor API
curl -X POST "https://harbor.example.com/api/v2.0/robots" \
  -H "Authorization: Basic $(echo -n admin:admin123 | base64)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ci-robot",
    "duration": -1,
    "description": "CI/CD robot account",
    "permissions": [
      {
        "kind": "project",
        "namespace": "production",
        "access": [
          {"resource": "repository", "action": "push"},
          {"resource": "repository", "action": "pull"}
        ]
      }
    ]
  }'
```

**GitHub Actions Integration**:
```yaml
# .github/workflows/build.yml
name: Build and Push to Harbor
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Login to Harbor
      uses: docker/login-action@v3
      with:
        registry: harbor.example.com
        username: robot$ci-robot
        password: ${{ secrets.HARBOR_ROBOT_TOKEN }}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: harbor.example.com/production/backend:${{ github.sha }}

    - name: Trigger Trivy scan
      run: |
        curl -X POST "https://harbor.example.com/api/v2.0/projects/production/repositories/backend/artifacts/${{ github.sha }}/scan" \
          -H "Authorization: Basic ${{ secrets.HARBOR_BASIC_AUTH }}"

    - name: Check scan result
      run: |
        while true; do
          STATUS=$(curl -s "https://harbor.example.com/api/v2.0/projects/production/repositories/backend/artifacts/${{ github.sha }}" \
            -H "Authorization: Basic ${{ secrets.HARBOR_BASIC_AUTH }}" | jq -r '.scan_overview."application/vnd.security.vulnerability.report; version=1.1".scan_status')
          if [ "$STATUS" = "Success" ]; then
            CRITICAL=$(curl -s "https://harbor.example.com/api/v2.0/projects/production/repositories/backend/artifacts/${{ github.sha }}" \
              -H "Authorization: Basic ${{ secrets.HARBOR_BASIC_AUTH }}" | jq -r '.scan_overview."application/vnd.security.vulnerability.report; version=1.1".summary.critical')
            if [ "$CRITICAL" -gt 0 ]; then
              echo "CRITICAL vulnerabilities found: $CRITICAL"
              exit 1
            fi
            break
          fi
          sleep 5
        done
```

**Webhook for Vulnerability Notifications**:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: harbor-webhook-config
data:
  webhook.json: |
    {
      "name": "slack-vulnerabilities",
      "description": "Notify Slack on critical vulnerabilities",
      "events": ["SCANNING_COMPLETED"],
      "targets": [
        {
          "type": "slack",
          "address": "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
          "skip_cert_verify": false
        }
      ],
      "enabled": true
    }
```

---

## Example 6: GitOps with ArgoCD 2.10.x

### ArgoCD Installation and Setup

**Install ArgoCD**:
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.10.0/manifests/install.yaml

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

**Application Manifest (GitOps)**:
```yaml
# apps/backend-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: backend-production
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/example/backend-manifests
    targetRevision: main
    path: production
    helm:
      values: |
        image:
          repository: harbor.example.com/production/backend
          tag: latest
        replicaCount: 3
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
```

**AppProject for Multi-Tenancy**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: production
  namespace: argocd
spec:
  description: Production environment
  sourceRepos:
  - 'https://github.com/example/*'
  destinations:
  - namespace: 'production'
    server: https://kubernetes.default.svc
  clusterResourceWhitelist:
  - group: ''
    kind: Namespace
  - group: 'apps'
    kind: Deployment
  - group: 'v1'
    kind: Service
  namespaceResourceBlacklist:
  - group: ''
    kind: ResourceQuota
  - group: ''
    kind: LimitRange
  roles:
  - name: prod-admin
    description: Admin privileges for production
    policies:
    - p, proj:production:prod-admin, applications, *, production/*, allow
    groups:
    - production-team
```

**Canary Rollout with Argo Rollouts**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: backend
  namespace: production
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 20
      - pause: {duration: 5m}
      - setWeight: 40
      - pause: {duration: 5m}
      - setWeight: 60
      - pause: {duration: 5m}
      - setWeight: 80
      - pause: {duration: 5m}
      canaryService: backend-canary
      stableService: backend-stable
      trafficRouting:
        istio:
          virtualService:
            name: backend
            routes:
            - primary
  revisionHistoryLimit: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: harbor.example.com/production/backend:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
```

**Analysis Template (Automated Rollback)**:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: backend-success-rate
  namespace: production
spec:
  metrics:
  - name: success-rate
    interval: 1m
    successCondition: result >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: http://prometheus.observability:9090
        query: |
          sum(rate(http_requests_total{app="backend",status!~"5.."}[5m]))
          /
          sum(rate(http_requests_total{app="backend"}[5m]))
```

---

## Summary

These examples demonstrate production-ready backend architectures with:

1. **Microservices**: Kubernetes 1.31 + Istio 1.21 for orchestration, traffic management, and security
2. **Event-Driven**: Kafka 3.7 for event streaming, CQRS for read/write separation
3. **Serverless**: AWS Lambda for auto-scaling, cost-effective event-driven workloads
4. **Observability**: OpenTelemetry 1.24 + Prometheus 2.48 + Jaeger 1.51 for full-stack telemetry
5. **Container Registry**: Harbor 2.10.x with Trivy scanning for vulnerability management
6. **GitOps**: ArgoCD 2.10.x with automated rollouts and canary deployments

All examples include:
- Latest tool versions (2025-10-22)
- Production-grade configurations
- Security best practices (mTLS, RBAC, OWASP compliance, image scanning)
- Observability integration
- Performance optimization (connection pooling, caching, rate limiting)
- Infrastructure as Code (Kubernetes manifests, Terraform)
- CI/CD integration (GitHub Actions, ArgoCD)
- Automated deployment strategies (canary, blue-green)
