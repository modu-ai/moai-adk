    [SemanticResourceAttributes.SERVICE_NAME]: 'your-service-name',
    [SemanticResourceAttributes.SERVICE_VERSION]: '1.0.0',
    [SemanticResourceAttributes.DEPLOYMENT_ENVIRONMENT]: process.env.NODE_ENV,
  }),
  
  // Auto-instrumentation for popular libraries
  instrumentations: [getNodeAutoInstrumentations()],
  
  // Trace exporter for distributed tracing
  traceExporter: new OTLPTraceExporter({
    url: process.env.OTEL_EXPORTER_OTLP_TRACES_ENDPOINT || 'http://jaeger:4317',
  }),
  
  // Metrics exporter
  metricExporter: new OTLPMetricExporter({
    url: process.env.OTEL_EXPORTER_OTLP_METRICS_ENDPOINT || 'http://prometheus:9090',
  }),
  
  // Additional Prometheus endpoint
  metricReader: new PrometheusExporter({
    port: 9464,
    endpoint: '/metrics',
  }),
  
  // Performance optimizations
  spanLimits: {
    attributeCountLimit: 100,
    eventCountLimit: 1000,
    linkCountLimit: 100,
  },
});

// Start the SDK
sdk.start().then(() => {
  console.log('OpenTelemetry initialized successfully');
});

// Graceful shutdown
process.on('SIGTERM', () => {
  sdk.shutdown()
    .then(() => console.log('OpenTelemetry shut down successfully'))
    .catch((error) => console.error('Error shutting down OpenTelemetry', error))
    .finally(() => process.exit(0));
});

// Custom span creation for business logic
import { trace } from '@opentelemetry/api';

export function createBusinessSpan(operationName: string, attributes: Record<string, string>) {
  const tracer = trace.getTracer('business-logic');
  
  return tracer.startSpan(operationName, {
    attributes: {
      'business.operation': operationName,
      'service.name': 'your-service-name',
      ...attributes,
    },
  });
}

// Example usage in business logic
export async function processUserOrder(userId: string, orderId: string) {
  const span = createBusinessSpan('process_user_order', {
    'user.id': userId,
    'order.id': orderId,
  });
  
  try {
    // Business logic here
    const result = await orderService.process(userId, orderId);
    
    span.setAttributes({
      'order.status': result.status,
      'order.amount': result.amount.toString(),
    });
    
    return result;
  } catch (error) {
    span.recordException(error as Error);
    throw error;
  } finally {
    span.end();
  }
}
```

## Prometheus Metrics Implementation

```typescript
// Custom Prometheus metrics for application monitoring
import { Counter, Histogram, Gauge, register } from 'prom-client';

// Business metrics
export const businessMetrics = {
  // Request counters
  httpRequestsTotal: new Counter({
    name: 'http_requests_total',
    help: 'Total number of HTTP requests',
    labelNames: ['method', 'route', 'status_code'],
  }),
  
  // Response time histograms
  httpRequestDuration: new Histogram({
    name: 'http_request_duration_seconds',
    help: 'HTTP request duration in seconds',
    labelNames: ['method', 'route'],
    buckets: [0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10],
  }),
  
  // Active connections gauge
  activeConnections: new Gauge({
    name: 'active_connections',
    help: 'Number of active connections',
  }),
  
  // Business operations
  ordersProcessed: new Counter({
    name: 'orders_processed_total',
    help: 'Total number of orders processed',
    labelNames: ['status', 'payment_method'],
  }),
  
  revenueGenerated: new Counter({
    name: 'revenue_generated_total',
    help: 'Total revenue generated',
    labelNames: ['currency'],
  }),
};

// System metrics
export const systemMetrics = {
  // Memory usage
  memoryUsage: new Gauge({
    name: 'memory_usage_bytes',
    help: 'Memory usage in bytes',
    labelNames: ['type'], // heap, external, array_buffers
  }),
  
  // CPU usage
  cpuUsage: new Gauge({
    name: 'cpu_usage_percent',
    help: 'CPU usage percentage',
  }),
  
  // Event loop lag
  eventLoopLag: new Histogram({
    name: 'event_loop_lag_seconds',
    help: 'Event loop lag in seconds',
    buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5],
  }),
};

// Metrics collection middleware
export function metricsMiddleware() {
  return (req: Request, res: Response, next: NextFunction) => {
    const start = Date.now();
    
    // Increment active connections
    systemMetrics.activeConnections.inc();
    
    res.on('finish', () => {
      const duration = (Date.now() - start) / 1000;
      
      // Record HTTP request metrics
      businessMetrics.httpRequestsTotal
        .labels(req.method, req.route?.path || req.path, res.statusCode.toString())
        .inc();
      
      businessMetrics.httpRequestDuration
        .labels(req.method, req.route?.path || req.path)
        .observe(duration);
      
      // Decrement active connections
      systemMetrics.activeConnections.dec();
    });
    
    next();
  };
}

// Export metrics for Prometheus
export function getMetrics() {
  return register.metrics();
}

// System metrics collection
setInterval(() => {
  const memUsage = process.memoryUsage();
  systemMetrics.memoryUsage.labels('heap').set(memUsage.heapUsed);
  systemMetrics.memoryUsage.labels('external').set(memUsage.external);
  systemMetrics.memoryUsage.labels('array_buffers').set(memUsage.arrayBuffers);
}, 5000);
```


# Advanced Implementation (Level 3)




## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.
