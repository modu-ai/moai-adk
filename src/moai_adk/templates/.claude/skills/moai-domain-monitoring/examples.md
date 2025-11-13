# Practical Examples & Code Patterns

## Example 1: Prometheus Scrape Configuration with Kubernetes Service Discovery

**Use Case**: Auto-discover and scrape metrics from Kubernetes pods

**prometheus.yml**:
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'production'
    environment: 'prod'
    region: 'us-east-1'

scrape_configs:
  # Kubernetes pod discovery
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - monitoring
            - production
            - staging
    
    relabel_configs:
      # Keep only pods with prometheus scrape annotation
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: 'true'
      
      # Use custom path from annotation
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: '(.+)'
        replacement: '${1}'
      
      # Use custom port from annotation
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: '([^:]+)(?::\d+)?;(\d+)'
        replacement: '${1}:${2}'
        target_label: __address__
      
      # Add pod name as label
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod
      
      # Add namespace as label
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      
      # Add pod label as service tag
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: replace
        target_label: service
    
    metric_relabel_configs:
      # Drop high-cardinality metrics
      - source_labels: [__name__]
        regex: 'http_request_duration_seconds_bucket'
        action: drop
      
      # Reduce cardinality by relabeling paths
      - source_labels: [__name__, http_path]
        regex: 'request_duration.*;/api/users/[0-9]+'
        target_label: http_path
        replacement: '/api/users/{id}'
```

---

## Example 2: Grafana 11.x Scenes-Powered Dashboard (TypeScript)

**Use Case**: Create modular, reusable dashboard components

**dashboard.ts**:
```typescript
import {
  SceneApp,
  SceneFlexItem,
  SceneFlexLayout,
  SceneTimeRangeState,
  PanelBuilders,
  VizPanel
} from '@grafana/scenes';
import { PrometheusDatasource } from '@grafana/scenes';

export function createProductionObservabilityDashboard() {
  return new SceneApp({
    title: 'Production Observability - v4.0',
    timeRange: new SceneTimeRangeState({
      from: 'now-6h',
      to: 'now',
    }),
    body: new SceneFlexLayout({
      direction: 'column',
      children: [
        // Top row: Key metrics
        new SceneFlexItem({
          minHeight: 200,
          body: new SceneFlexLayout({
            children: [
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Request Rate (5m avg)')
                  .setUnit('reqps')
                  .setDataSource('Prometheus')
                  .setTargets([
                    {
                      expr: 'rate(http_requests_total[5m])',
                      refId: 'A',
                    }
                  ])
                  .setOptions({
                    colorMode: 'gradient',
                    graphMode: 'area',
                    orientation: 'auto',
                  })
                  .setFieldConfig({
                    defaults: {
                      color: {
                        mode: 'gradient-gauge',
                        seriesBy: 'last',
                      },
                      thresholds: {
                        mode: 'percentage',
                        steps: [
                          { color: 'green', value: 0 },
                          { color: 'yellow', value: 70 },
                          { color: 'red', value: 90 },
                        ],
                      },
                    },
                  })
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Error Rate (5m avg)')
                  .setUnit('percent')
                  .setTargets([
                    {
                      expr: 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Response Time p99')
                  .setUnit('ms')
                  .setTargets([
                    {
                      expr: 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) * 1000',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
              new SceneFlexItem({
                width: '25%',
                body: PanelBuilders.stat()
                  .setTitle('Active Traces')
                  .setUnit('short')
                  .setTargets([
                    {
                      expr: 'active_traces',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
            ],
          })
        }),
        
        // Second row: Timeseries charts
        new SceneFlexItem({
          minHeight: 300,
          body: new SceneFlexLayout({
            children: [
              new SceneFlexItem({
                width: '50%',
                body: PanelBuilders.timeseries()
                  .setTitle('Request Rate by Service')
                  .setUnit('reqps')
                  .setTargets([
                    {
                      expr: 'rate(http_requests_total[5m])',
                      legendFormat: '{{service}}',
                      refId: 'A',
                    }
                  ])
                  .setOptions({
                    showLegend: true,
                    legend: {
                      showLegend: true,
                      placement: 'right',
                    },
                  })
                  .build()
              }),
              new SceneFlexItem({
                width: '50%',
                body: PanelBuilders.timeseries()
                  .setTitle('Error Rate by Service')
                  .setUnit('percent')
                  .setTargets([
                    {
                      expr: 'rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) * 100',
                      legendFormat: '{{service}}',
                      refId: 'A',
                    }
                  ])
                  .build()
              }),
            ],
          })
        }),
        
        // Third row: Heatmap and histogram
        new SceneFlexItem({
          minHeight: 300,
          body: PanelBuilders.heatmap()
            .setTitle('Response Time Distribution')
            .setUnit('ms')
            .setTargets([
              {
                expr: 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) * 1000',
                refId: 'A',
              }
            ])
            .build()
        }),
      ],
    }),
  });
}
```

**Benefits of Scenes API**:
- Type-safe dashboard code
- Reusable components
- Programmatic dashboard generation
- Version control friendly (JSON → TypeScript)

---

## Example 3: OpenTelemetry Node.js Instrumentation with Tracing

**Use Case**: Auto-instrument a Node.js Express application with OpenTelemetry

**instrumentation.js**:
```javascript
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { JaegerExporter } = require('@opentelemetry/exporter-trace-jaeger');
const { BatchSpanProcessor } = require('@opentelemetry/sdk-trace-node');
const { PeriodicExportingMetricReader } = require('@opentelemetry/sdk-metrics-node');
const { PrometheusExporter } = require('@opentelemetry/exporter-prometheus');
const { ConsoleSpanExporter, SimpleSpanProcessor } = require('@opentelemetry/sdk-trace-node');

// Initialize Jaeger exporter for traces
const jaegerExporter = new JaegerExporter({
  endpoint: process.env.JAEGER_ENDPOINT || 'http://localhost:14268/api/traces',
});

// Initialize Prometheus exporter for metrics
const prometheusExporter = new PrometheusExporter({
  port: 8888,
  endpoint: '/metrics',
});

// Create and start Node SDK
const sdk = new NodeSDK({
  traceExporter: jaegerExporter,
  instrumentations: [getNodeAutoInstrumentations()],
  serviceName: 'order-service',
  serviceVersion: '4.0.0',
  metricReader: new PeriodicExportingMetricReader({
    exporter: prometheusExporter,
    intervalMillis: 10000,
  }),
});

sdk.start();
console.log('OpenTelemetry SDK initialized');

process.on('SIGTERM', () => {
  sdk.shutdown()
    .then(() => console.log('SDK shutdown successfully'))
    .catch(error => console.log('Error shutting down SDK', error))
    .finally(() => process.exit(0));
});
```

**express-middleware.js** (Custom span creation):
```javascript
const { trace, context, SpanStatusCode } = require('@opentelemetry/api');

const tracer = trace.getTracer('order-service', '4.0.0');

// Express middleware for automatic span creation
function traceMiddleware(req, res, next) {
  const span = tracer.startSpan('http.request', {
    attributes: {
      'http.method': req.method,
      'http.url': req.url,
      'http.target': req.path,
      'http.host': req.hostname,
      'http.scheme': req.protocol,
      'http.client_ip': req.ip,
    },
  });

  // Ensure span is active during request processing
  context.with(trace.setSpan(context.active(), span), () => {
    res.on('finish', () => {
      span.setAttributes({
        'http.status_code': res.statusCode,
      });
      
      if (res.statusCode >= 400) {
        span.setStatus({ code: SpanStatusCode.ERROR });
      } else {
        span.setStatus({ code: SpanStatusCode.OK });
      }
      
      span.end();
    });
    
    next();
  });
}

// Example route with custom spans
app.post('/api/orders', traceMiddleware, async (req, res) => {
  const activeSpan = trace.getActiveSpan();
  
  // Create child span for order validation
  const validationSpan = tracer.startSpan('order.validation', {
    parent: activeSpan,
  });
  
  try {
    validationSpan.addEvent('validation_started');
    const errors = validateOrder(req.body);
    
    if (errors.length > 0) {
      validationSpan.setStatus({ code: SpanStatusCode.ERROR, message: 'Validation failed' });
      return res.status(400).json({ errors });
    }
    
    validationSpan.setStatus({ code: SpanStatusCode.OK });
    validationSpan.end();
    
    // Create child span for database operation
    const dbSpan = tracer.startSpan('db.order.create', {
      parent: activeSpan,
    });
    
    const order = await db.orders.create(req.body);
    dbSpan.end();
    
    activeSpan.addEvent('order_created', { 'order.id': order.id });
    res.json(order);
  } catch (error) {
    activeSpan.recordException(error);
    activeSpan.setStatus({ code: SpanStatusCode.ERROR, message: error.message });
    res.status(500).json({ error: error.message });
  }
});

app.use(traceMiddleware);
```

---

## Example 4: Jaeger Python Distributed Tracing

**Use Case**: Implement distributed tracing in Python microservices

**tracing-setup.py**:
```python
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from opentelemetry.propagators.jaeger import JaegerPropagator
from opentelemetry.propagate import set_global_textmap

import os

# Create Jaeger exporter
jaeger_exporter = JaegerExporter(
    agent_host_name=os.getenv('JAEGER_HOST', 'localhost'),
    agent_port=int(os.getenv('JAEGER_PORT', 6831)),
)

# Create and set TracerProvider
trace_provider = TracerProvider()
trace_provider.add_span_processor(BatchSpanProcessor(jaeger_exporter))

# Set global TracerProvider
from opentelemetry import trace as otel_trace
otel_trace.set_tracer_provider(trace_provider)

# Set propagator for W3C Trace Context
set_global_textmap(JaegerPropagator())

# Auto-instrument libraries
FlaskInstrumentor().instrument()
RequestsInstrumentor().instrument()
SQLAlchemyInstrumentor().instrument_connection_string(
    'postgresql://user:password@localhost/db'
)
```

**flask-app.py** (Custom spans):
```python
from opentelemetry import trace
from flask import Flask, jsonify, request
import requests

app = Flask(__name__)
tracer = trace.get_tracer(__name__)

@app.route('/api/users/<user_id>', methods=['GET'])
def get_user(user_id):
    with tracer.start_as_current_span('user.lookup') as span:
        span.set_attribute('user.id', user_id)
        
        # Simulate database lookup
        user_data = {
            'id': user_id,
            'name': f'User {user_id}',
            'email': f'user{user_id}@example.com',
        }
        
        span.set_attribute('user.found', user_data is not None)
        
        # Call downstream service with context propagation
        with tracer.start_as_current_span('enrichment.service') as enrich_span:
            headers = {}
            # Propagator automatically injects trace context
            from opentelemetry.propagate import inject
            inject(headers)
            
            response = requests.get(
                'http://enrichment-service/enrich',
                json=user_data,
                headers=headers
            )
            enriched_user = response.json()
        
        return jsonify(enriched_user)

if __name__ == '__main__':
    app.run(debug=False)
```

---

## Example 5: Loki LogQL Query with Bloom Filter Optimization

**Use Case**: Query logs efficiently with Bloom filter acceleration

**logql-queries.txt**:
```logql
# Basic error log query (Bloom accelerated)
{job="api-server", level="ERROR"} | json

# Count errors by service
sum(count_over_time({level="ERROR"}[5m])) by (service)

# Error rate calculation
sum(rate({level="ERROR"}[5m])) / sum(rate({level!=""}[5m])) * 100

# Pattern matching with Bloom filters (highly optimized)
{job="api-server"} 
  |= "ERROR" 
  |= "database" 
  |= "connection timeout"

# Complex query with multiple filters
{job="payment-service", env="prod"}
  | json response_time=response_time_ms, status=http_status
  | status >= "400"
  | response_time > 1000
  | stats avg(response_time), count() by (endpoint)

# Trace correlation (using trace_id)
{job="microservice"} 
  | json trace_id=trace_id 
  | trace_id="abc123def456"
```

**Bloom Filter Impact**:
- Without Bloom: Process 100% of chunks
- With Bloom: Process only 10-30% of chunks
- Result: 70-90% query acceleration for specific text searches

---

## Example 6: Elasticsearch Query DSL for Log Analysis

**Use Case**: Advanced log analysis with Elasticsearch aggregations

**elasticsearch-queries.json**:
```json
{
  "query": {
    "bool": {
      "must": [
        {
          "range": {
            "@timestamp": {
              "gte": "now-1h",
              "lte": "now"
            }
          }
        },
        {
          "terms": {
            "level": ["ERROR", "WARN"]
          }
        }
      ],
      "filter": [
        {
          "term": {
            "environment": "production"
          }
        }
      ]
    }
  },
  "aggs": {
    "errors_by_service": {
      "terms": {
        "field": "service.keyword",
        "size": 20
      },
      "aggs": {
        "error_trends": {
          "date_histogram": {
            "field": "@timestamp",
            "calendar_interval": "1m"
          },
          "aggs": {
            "error_count": {
              "value_count": {
                "field": "trace_id"
              }
            }
          }
        },
        "top_errors": {
          "top_hits": {
            "size": 3,
            "_source": ["message", "stack_trace", "@timestamp"],
            "sort": [{"@timestamp": {"order": "desc"}}]
          }
        }
      }
    }
  }
}
```

---

## Example 7: Sentry Error Tracking with Event Sampling

**Use Case**: Implement Sentry SDK with intelligent sampling

**sentry-setup.js**:
```javascript
const Sentry = require("@sentry/node");
const { ProfilingIntegration } = require("@sentry/profiling-node");

Sentry.init({
  dsn: process.env.SENTRY_DSN,
  environment: process.env.NODE_ENV,
  release: process.env.APP_VERSION,
  
  // Transaction sampling
  tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
  
  // Profiling sampling
  profilesSampleRate: 0.1,
  
  integrations: [
    new Sentry.Integrations.OnUncaughtException({
      onFatalError: async (error) => {
        console.error('Fatal error:', error);
      },
    }),
    new Sentry.Integrations.OnUnhandledRejection(),
    new Sentry.Integrations.Http({ breadcrumbs: true, tracing: true }),
    new Sentry.Integrations.Express({
      request: true,
      serverName: false,
      transaction: true,
    }),
    new ProfilingIntegration(),
  ],
  
  // Error filtering
  beforeSend(event, hint) {
    // Filter 404 errors
    if (event.request && event.request.url?.includes('/health')) {
      return null;
    }
    
    // Filter known third-party errors
    if (event.exception && event.exception[0].value?.includes('external-library')) {
      event.level = 'info';
    }
    
    return event;
  },
  
  // Breadcrumb filtering
  beforeBreadcrumb(breadcrumb, hint) {
    if (breadcrumb.category === 'console' && breadcrumb.level === 'debug') {
      return null;
    }
    return breadcrumb;
  },
});

export default Sentry;
```

**Error handling with context**:
```javascript
app.post('/api/payment', (req, res) => {
  Sentry.captureMessage('Payment processing started', 'info');
  
  try {
    const result = processPayment(req.body);
    res.json(result);
  } catch (error) {
    // Capture with context
    Sentry.captureException(error, {
      level: 'error',
      tags: {
        'payment.processor': 'stripe',
        'payment.type': req.body.type,
      },
      contexts: {
        payment: {
          amount: req.body.amount,
          currency: req.body.currency,
          customer_id: req.user.id,
        },
      },
    });
    
    res.status(500).json({ error: 'Payment processing failed' });
  }
});
```

---

## Example 8: Custom Prometheus Exporter (Python)

**Use Case**: Export custom business metrics to Prometheus

**custom-exporter.py**:
```python
from prometheus_client import Counter, Gauge, Histogram, start_http_server
import time

# Define custom metrics
order_total = Counter(
    'orders_total',
    'Total number of orders processed',
    ['status', 'payment_method']
)

order_value = Gauge(
    'order_value_usd',
    'Current order value in USD',
    ['currency']
)

order_processing_time = Histogram(
    'order_processing_seconds',
    'Order processing time in seconds',
    ['processor'],
    buckets=(0.1, 0.5, 1.0, 5.0, 10.0)
)

# Start metrics server
start_http_server(8000)

# Example usage
def process_order(order_data):
    start_time = time.time()
    
    try:
        # Process order
        result = perform_payment(order_data)
        
        # Record metrics
        order_total.labels(
            status='completed',
            payment_method=order_data['payment_method']
        ).inc()
        
        order_value.labels(currency=order_data['currency']).set(order_data['amount'])
        
        duration = time.time() - start_time
        order_processing_time.labels(processor='payment-gateway').observe(duration)
        
        return result
    except Exception as e:
        order_total.labels(
            status='failed',
            payment_method=order_data['payment_method']
        ).inc()
        raise
```

---

## Example 9: Prometheus Alert Rule with Composite Conditions

**Use Case**: Define complex alert rules combining multiple conditions

**alert-rules.yaml**:
```yaml
groups:
  - name: application_slo
    interval: 30s
    rules:
      # Alert when error rate exceeds threshold AND latency degrades
      - alert: ServiceDegradation
        expr: |
          (rate(http_requests_total{status=~"5.."}[5m]) > 0.01)
          AND
          (histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5)
        for: 5m
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "Service {{ $labels.service }} experiencing degradation"
          description: |
            Error rate: {{ $value | humanizePercentage }}
            P95 latency: {{ query "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))" | humanizeDuration }}
      
      # Memory leak detection
      - alert: MemoryLeakDetected
        expr: |
          rate(process_resident_memory_bytes[30m]) > 0
          AND
          rate(process_resident_memory_bytes[30m]) > 
          rate(process_resident_memory_bytes offset 6h [30m]) * 1.2
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Possible memory leak in {{ $labels.pod }}"
```

---

## Example 10: SLO/SLI with Error Budget Implementation

**Use Case**: Define and monitor SLOs with error budgets

**slo-definition.yaml**:
```yaml
groups:
  - name: slo_monitoring
    interval: 1m
    rules:
      # SLO: 99.95% availability (0.05% error budget)
      - record: slo:http_requests:ratio_5m
        expr: |
          sum(rate(http_requests_total{status!~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))
      
      - record: slo:http_requests:ratio_30d
        expr: |
          sum(increase(http_requests_total{status!~"5.."}[30d]))
          /
          sum(increase(http_requests_total[30d]))
      
      # Error budget burn rate
      - record: slo:error_budget_burn_rate_5m
        expr: |
          (1 - slo:http_requests:ratio_5m) / 0.0005  # 0.05% budget
      
      # Alert if consuming error budget too fast
      - alert: SLOErrorBudgetBurnTooFast
        expr: |
          (
            slo:error_budget_burn_rate_5m > 1  # Consuming faster than expected
          )
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Error budget burn rate is {{ $value | humanize }}x"
          description: "SLO is on track to exhaust error budget in {{ $value }} days"
```

---

## Quick Reference: Common Patterns

| Pattern | Use Case | Example |
|---------|----------|---------|
| **Cardinality Reduction** | Prevent metric explosion | Use regex relabeling for path patterns |
| **Sampling** | Control data volume | Set tracesSampleRate in SDKs |
| **Aggregation** | Reduce query load | Use recording rules for frequent queries |
| **Retention** | Cost optimization | Configure ILM in Elasticsearch |
| **Multi-tenancy** | Multi-customer support | Use Jaeger multi-tenancy mode |
| **HA Architecture** | High availability | Deploy Thanos with remote storage |

