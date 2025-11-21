# Distributed Tracing Strategy — OpenTelemetry & Observability

_Last updated: 2025-11-22_

## OpenTelemetry Setup

### Instrumentation

```python
# Python FastAPI
from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)

trace.set_tracer_provider(TracerProvider())
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

tracer = trace.get_tracer(__name__)

@app.get("/users/{user_id}")
async def get_user(user_id: str):
    with tracer.start_as_current_span("get_user") as span:
        span.set_attribute("user.id", user_id)
        
        with tracer.start_as_current_span("fetch_from_database"):
            user = await db.get_user(user_id)
        
        return user
```

### Correlation IDs

```typescript
// TypeScript Express
import { Request, Response, NextFunction } from 'express';
import { v4 as uuidv4 } from 'uuid';

export function correlationIdMiddleware(
    req: Request,
    res: Response,
    next: NextFunction
) {
    const correlationId = req.headers['x-correlation-id'] || uuidv4();
    req.correlationId = correlationId;
    res.setHeader('x-correlation-id', correlationId);

    res.on('finish', () => {
        console.log(`[${correlationId}] ${req.method} ${req.path} ${res.statusCode}`);
    });

    next();
}

app.use(correlationIdMiddleware);
```

## Sampling Strategies

### Adaptive Sampling
- Low volume services: 100% sampling
- High volume services: Proportional sampling
- Error traces: Always sample
- Slow traces: Tail-based sampling

### Decision-Based Sampling
```
if error or latency > threshold:
    sample = true
elif service_type == "critical":
    sample = true
else:
    sample = random(0, 1) < 0.1  # 10% sampling
```

## Common Patterns

### Request Tracing
- Generate trace ID on entry point
- Propagate through request context
- Log trace ID in all services
- Query traces by trace ID in UI

### Error Context
- Capture exception in span
- Include stack trace
- Log error metadata
- Alert on anomalies

### Performance Analysis
- Measure service latency
- Identify bottlenecks
- Compare trace durations
- Analyze tail latencies (p99)

---

**Last Updated**: 2025-11-22

