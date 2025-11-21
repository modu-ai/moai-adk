# Advanced Monitoring Patterns

> **Version**: 4.0.0
> **Last Updated**: 2025-11-22
> **Focus**: Distributed tracing, OpenTelemetry, correlation IDs, error tracking, performance profiling

---

## Distributed Tracing with OpenTelemetry

### OpenTelemetry Instrumentation

```python
from opentelemetry import trace, metrics, baggage
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.exporter.prometheus import PrometheusMetricReader
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from fastapi import FastAPI, Request

# Configure Jaeger exporter
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)

# Setup tracing
trace_provider = TracerProvider()
trace_provider.add_span_processor(BatchSpanProcessor(jaeger_exporter))
trace.set_tracer_provider(trace_provider)

# Setup metrics
metrics_reader = PrometheusMetricReader()
meter_provider = MeterProvider(metric_readers=[metrics_reader])
metrics.set_meter_provider(meter_provider)

# Create FastAPI app
app = FastAPI()

# Auto-instrument libraries
FastAPIInstrumentor.instrument_app(app)
SQLAlchemyInstrumentor().instrument()

# Manual instrumentation
tracer = trace.get_tracer(__name__)
meter = metrics.get_meter(__name__)

# Business metrics
order_counter = meter.create_counter(
    "orders.created",
    description="Number of orders created",
    unit="1"
)

payment_duration = meter.create_histogram(
    "payment.duration",
    description="Payment processing duration",
    unit="ms"
)

@app.post("/orders/")
async def create_order(request: Request, order_data: dict):
    """Create order with full tracing"""

    with tracer.start_as_current_span("create_order") as span:
        # Add attributes to span
        span.set_attribute("order.amount", order_data["amount"])
        span.set_attribute("order.currency", order_data["currency"])
        span.set_attribute("user.id", order_data["user_id"])

        # Nested span for database
        with tracer.start_as_current_span("db.save_order"):
            # Save order to database
            saved_order = await save_order(order_data)

        # Nested span for payment processing
        with tracer.start_as_current_span("payment.process"):
            # Process payment with metrics
            start_time = time.time()

            payment_result = await process_payment(
                order_id=saved_order.id,
                amount=order_data["amount"]
            )

            duration_ms = (time.time() - start_time) * 1000
            payment_duration.record(int(duration_ms))

        # Record metric
        order_counter.add(1, {"currency": order_data["currency"]})

        return {"order_id": saved_order.id, "status": "created"}
```

### Correlation ID Propagation

```python
import contextvars
from opentelemetry import trace, baggage, context as otel_context
from opentelemetry.propagators.b3 import B3MultiFormat
from opentelemetry.propagators.jaeger import JaegerPropagator
from opentelemetry.propagators.composite import CompositePropagator

# Setup context propagators
CompositePropagator([JaegerPropagator(), B3MultiFormat()])

# Context variable for correlation ID
correlation_id: contextvars.ContextVar[str] = contextvars.ContextVar(
    'correlation_id', default=None
)

class CorrelationIDMiddleware:
    """Manage correlation IDs across request lifecycle"""

    async def __call__(self, request: Request, call_next):
        # Extract correlation ID from headers (or generate new)
        corr_id = request.headers.get(
            'X-Correlation-ID',
            str(uuid.uuid4())
        )

        # Set in context
        correlation_id.set(corr_id)

        # Add to baggage for distribution
        ctx = baggage.set_baggage("correlation_id", corr_id)

        # Add to response header
        response = await call_next(request)
        response.headers["X-Correlation-ID"] = corr_id

        return response

app.add_middleware(CorrelationIDMiddleware)

async def call_external_service(service_url: str, data: dict):
    """Call external service with correlation ID"""

    # Get current correlation ID
    corr_id = correlation_id.get()

    headers = {
        "X-Correlation-ID": corr_id,
        "Traceparent": trace.get_current_span().to_trace_id(),  # W3C format
    }

    async with httpx.AsyncClient() as client:
        return await client.post(
            service_url,
            json=data,
            headers=headers
        )
```

---

## Advanced Error Tracking

### Exception Context and Metadata

```python
class ErrorTracker:
    """Comprehensive error tracking with context"""

    @staticmethod
    def track_error_with_context(
        error: Exception,
        error_context: dict = None,
        user_id: str = None,
        request_id: str = None
    ):
        """Track error with full context"""

        span = trace.get_current_span()

        # Record exception in span
        span.record_exception(
            error,
            attributes={
                "error.type": type(error).__name__,
                "error.severity": get_severity(error),
                "user.id": user_id,
                "request.id": request_id,
                **error_context
            }
        )

        # Send to error tracking service
        error_event = {
            "exception": {
                "type": type(error).__name__,
                "message": str(error),
                "stacktrace": traceback.format_exc(),
                "severity": get_severity(error)
            },
            "context": {
                "user_id": user_id,
                "request_id": request_id,
                "correlation_id": correlation_id.get(),
                "timestamp": datetime.utcnow().isoformat(),
                **error_context
            }
        }

        # Send to monitoring service (Sentry, Datadog, etc.)
        send_to_error_service(error_event)

        return error_event

    @staticmethod
    @app.exception_handler(Exception)
    async def global_exception_handler(request, exc):
        """Global exception handler with error tracking"""

        error_event = ErrorTracker.track_error_with_context(
            error=exc,
            error_context={
                "endpoint": request.url.path,
                "method": request.method,
            },
            user_id=get_user_id(request),
            request_id=request.headers.get("X-Request-ID")
        )

        return JSONResponse(
            status_code=500,
            content={
                "error": "Internal server error",
                "error_id": error_event["context"]["request_id"]
            }
        )
```

---

## Performance Profiling Patterns

### CPU and Memory Profiling

```python
import cProfile
import memory_profiler
from py_spy import Spy

class PerformanceProfiler:
    """Profile CPU and memory performance"""

    @staticmethod
    def profile_cpu_intensive_operation(func):
        """CPU profiling decorator"""

        def wrapper(*args, **kwargs):
            profiler = cProfile.Profile()
            profiler.enable()

            result = func(*args, **kwargs)

            profiler.disable()

            # Analyze results
            import io
            import pstats

            s = io.StringIO()
            ps = pstats.Stats(profiler, stream=s).sort_stats('cumulative')
            ps.print_stats(10)

            # Log results
            print(s.getvalue())

            return result

        return wrapper

    @staticmethod
    def profile_memory(func):
        """Memory profiling decorator"""

        @memory_profiler.profile
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    @staticmethod
    async def continuous_profiling():
        """Continuous profiling with py_spy"""

        # Profile running process
        spy = Spy(pid=os.getpid())

        async with asyncio.timeout(60):  # Profile for 60 seconds
            while True:
                frame = spy.sample()
                if frame:
                    trace.get_current_span().set_attribute(
                        "cpu.profile",
                        frame
                    )
                await asyncio.sleep(0.1)
```

### Custom Metrics Collection

```python
from prometheus_client import Histogram, Counter, Gauge, Summary

class BusinessMetrics:
    """Define and collect business metrics"""

    # Request metrics
    request_duration = Histogram(
        'request_duration_seconds',
        'Request duration',
        buckets=(0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0),
        labelnames=['method', 'endpoint', 'status']
    )

    # Business metrics
    order_value = Histogram(
        'order_value_usd',
        'Order value in USD',
        buckets=(10, 50, 100, 500, 1000, 5000),
        labelnames=['currency', 'country']
    )

    # System metrics
    database_query_duration = Summary(
        'db_query_duration_seconds',
        'Database query duration',
        labelnames=['operation', 'table']
    )

    cache_hit_ratio = Gauge(
        'cache_hit_ratio',
        'Cache hit ratio',
        labelnames=['cache_name']
    )

    # Usage metrics
    active_users = Gauge(
        'active_users',
        'Number of active users'
    )

    @staticmethod
    def record_request(method: str, endpoint: str, status: int, duration: float):
        """Record HTTP request"""
        BusinessMetrics.request_duration.labels(
            method=method,
            endpoint=endpoint,
            status=status
        ).observe(duration)

    @staticmethod
    def record_order(amount: float, currency: str, country: str):
        """Record order"""
        BusinessMetrics.order_value.labels(
            currency=currency,
            country=country
        ).observe(amount)
```

---

## Alerting Patterns

### Intelligent Alert Routing

```yaml
# AlertManager configuration with intelligent routing
global:
  resolve_timeout: 5m

route:
  # Default route
  receiver: default
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 4h

  # Critical routing
  routes:
  - match:
      severity: critical
    receiver: critical
    group_wait: 0s
    group_interval: 1m
    repeat_interval: 1h

    # Auto-escalation if not resolved
    routes:
    - match:
        escalate: 'true'
      receiver: escalation
      group_wait: 5m
      repeat_interval: 30m

  # Warning routing
  - match:
      severity: warning
    receiver: warnings
    group_wait: 30s
    group_interval: 5m
    repeat_interval: 12h

  # Info routing
  - match:
      severity: info
    receiver: info
    group_wait: 1m
    group_interval: 30m
    repeat_interval: 24h

receivers:
- name: default
  slack_configs:
  - channel: '#alerts'
    title: 'Alert: {{ .GroupLabels.alertname }}'
    text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

- name: critical
  slack_configs:
  - channel: '#critical-incidents'
    title: 'CRITICAL: {{ .GroupLabels.alertname }}'
  pagerduty_configs:
  - service_key: '{{ .GroupLabels.pagerduty_key }}'
    description: '{{ .GroupLabels.alertname }}'

- name: escalation
  pagerduty_configs:
  - service_key: '{{ .GroupLabels.escalation_key }}'
  opsgenie_configs:
  - api_key: '{{ .GroupLabels.opsgenie_key }}'
    responders:
    - name: '{{ .GroupLabels.team }}'
      type: team

- name: warnings
  slack_configs:
  - channel: '#warnings'

- name: info
  slack_configs:
  - channel: '#info'
```

---

## Best Practices

- Use correlation IDs for request tracing
- Implement structured logging with context
- Profile regularly in production
- Set up comprehensive error tracking
- Define clear alerting policies
- Monitor both technical and business metrics
- Implement distributed tracing end-to-end
- Use sampling for high-volume traces
- Maintain detailed runbooks
- Practice alert escalation

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
