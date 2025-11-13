---
name: "moai-domain-monitoring"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Application Monitoring with AI-powered observability architecture, Context7 integration, and intelligent performance orchestration for scalable modern applications
keywords: ['monitoring', 'observability', 'performance', 'analytics', 'metrics', 'logging', 'tracing', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Application Monitoring Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-domain-monitoring |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Monitoring Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when monitoring keywords detected |

---

## What It Does

Enterprise Application Monitoring expert with AI-powered observability architecture, Context7 integration, and intelligent performance orchestration for scalable modern applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Monitoring Architecture** using Context7 MCP for latest observability patterns
- 📊 **Intelligent Performance Analytics** with automated anomaly detection and optimization
- 🚀 **Advanced Observability Integration** with AI-driven distributed tracing and correlation
- 🔗 **Enterprise Alerting Systems** with zero-configuration intelligent incident management
- 📈 **Predictive Performance Insights** with usage forecasting and capacity planning

---

## When to Use

**Automatic triggers**:
- Application monitoring architecture and observability strategy discussions
- Performance optimization and bottleneck analysis planning
- Alerting and incident management system implementation
- Distributed tracing and system correlation analysis

**Manual invocation**:
- Designing enterprise monitoring architectures with optimal observability
- Implementing comprehensive performance monitoring and analytics
- Planning incident response and alerting strategies
- Optimizing system performance and capacity planning

---

# Quick Reference (Level 1)

## Modern Monitoring Stack (November 2025)

### Core Monitoring Components
- **Metrics Collection**: Prometheus, Grafana, DataDog, New Relic
- **Logging**: ELK Stack, Grafana Loki, Fluentd, Logstash
- **Tracing**: Jaeger, OpenTelemetry, Zipkin, AWS X-Ray
- **APM**: Application Performance Monitoring with real-time insights
- **Synthetic Monitoring**: Active user experience simulation

### Key Observability Pillars
- **Logs**: Structured event logging with correlation IDs
- **Metrics**: Time-series data for system performance
- **Traces**: Distributed request flow across services
- **Events**: Business and system event correlation
- **Profiles**: Application performance profiling

### Popular Integration Patterns
- **OpenTelemetry**: Vendor-neutral observability data collection
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Visualization and dashboarding
- **DataDog**: Full-stack monitoring and APM
- **New Relic**: Application performance and infrastructure monitoring

### Alerting Strategy
- **SLI/SLO Monitoring**: Service level objectives and indicators
- **Threshold-based Alerts**: Performance and availability thresholds
- **Anomaly Detection**: AI-powered anomaly identification
- **Escalation Policies**: Multi-level alerting and notification

---

# Core Implementation (Level 2)

## Monitoring Architecture Intelligence

```python
# AI-powered monitoring architecture optimization with Context7
class MonitoringArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.observability_analyzer = ObservabilityAnalyzer()
        self.performance_optimizer = PerformanceOptimizer()
    
    async def design_optimal_monitoring_architecture(self, 
                                                   requirements: MonitoringRequirements) -> MonitoringArchitecture:
        """Design optimal monitoring architecture using AI analysis."""
        
        # Get latest monitoring and observability documentation via Context7
        monitoring_docs = await self.context7_client.get_library_docs(
            context7_library_id='/monitoring/docs',
            topic="observability metrics tracing logging alerting 2025",
            tokens=3000
        )
        
        observability_docs = await self.context7_client.get_library_docs(
            context7_library_id='/observability/docs',
            topic="opentelemetry prometheus grafana performance 2025",
            tokens=2000
        )
        
        # Optimize observability stack
        observability_design = self.observability_analyzer.optimize_stack(
            requirements.application_complexity,
            requirements.scale_requirements,
            monitoring_docs
        )
        
        # Design alerting strategy
        alerting_strategy = self.performance_optimizer.design_alerting(
            requirements.service_level_objectives,
            requirements.notification_preferences,
            observability_docs
        )
        
        return MonitoringArchitecture(
            metrics_collection=self._configure_metrics(requirements),
            logging_system=self._configure_logging(requirements),
            tracing_setup=self._configure_tracing(requirements),
            alerting_framework=alerting_strategy,
            observability_stack=observability_design,
            dashboard_configuration=self._design_dashboards(requirements),
            performance_predictions=observability_design.predictions
        )
```

## OpenTelemetry Integration

```typescript
// Comprehensive OpenTelemetry setup for Node.js applications
import { NodeSDK } from '@opentelemetry/sdk-node';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { Resource } from '@opentelemetry/resources';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';
import { OTLPTraceExporter } from '@opentelemetry/exporter-otlp-grpc';
import { OTLPMetricExporter } from '@opentelemetry/exporter-otlp-grpc';
import { PrometheusExporter } from '@opentelemetry/exporter-prometheus';

// Initialize OpenTelemetry SDK
const sdk = new NodeSDK({
  resource: new Resource({
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

---

# Advanced Implementation (Level 3)

## Advanced Alerting and Incident Management

```python
class IntelligentAlertingSystem:
    def __init__(self):
        self.anomaly_detector = AnomalyDetector()
        self.escalation_manager = EscalationManager()
        self.correlation_engine = AlertCorrelationEngine()
    
    async def setup_intelligent_alerting(self, 
                                       monitoring_config: MonitoringConfiguration) -> AlertingSetup:
        """Configure intelligent alerting with anomaly detection."""
        
        # Set up anomaly detection
        anomaly_config = self.anomaly_detector.configure_detection(
            monitoring_config.metrics,
            sensitivity_level=monitoring_config.sensitivity,
            learning_period=monitoring_config.learning_period
        )
        
        # Configure escalation policies
        escalation_policies = self.escalation_manager.create_policies(
            monitoring_config.severity_levels,
            monitoring_config.notification_channels
        )
        
        # Set up alert correlation
        correlation_rules = self.correlation_engine.define_correlation_rules(
            monitoring_config.service_dependencies,
            monitoring_config.infrastructure_topology
        )
        
        return AlertingSetup(
            anomaly_detection=anomaly_config,
            escalation_policies=escalation_policies,
            correlation_rules=correlation_rules,
            suppression_rules=self._configure_suppression_rules(),
            enrichment_rules=self._configure_enrichment_rules()
        )
```

### Performance Optimization with Machine Learning

```typescript
// AI-powered performance optimization
export class PerformanceOptimizer {
  private performanceData: PerformanceMetrics[] = [];
  private model: PerformanceModel;

  constructor() {
    this.model = new PerformanceModel();
  }

  async collectPerformanceMetrics(): Promise<void> {
    // Collect comprehensive performance metrics
    const metrics = await this.gatherMetrics();
    this.performanceData.push(metrics);
    
    // Keep only recent data for training
    if (this.performanceData.length > 1000) {
      this.performanceData = this.performanceData.slice(-1000);
    }
  }

  async predictPerformanceIssues(): Promise<PerformancePrediction[]> {
    // Use trained model to predict potential issues
    const features = this.extractFeatures(this.performanceData);
    const predictions = await this.model.predict(features);
    
    return predictions.map((prediction, index) => ({
      timestamp: Date.now() + (index * 60000), // Next hour predictions
      issue_type: prediction.type,
      confidence: prediction.confidence,
      severity: prediction.severity,
      recommended_actions: this.getRecommendedActions(prediction),
    }));
  }

  async optimizeResourceAllocation(): Promise<ResourceOptimization> {
    // Analyze current resource usage patterns
    const usagePatterns = this.analyzeUsagePatterns();
    
    // Generate optimization recommendations
    return {
      cpu_scaling: this.optimizeCPUAllocation(usagePatterns.cpu),
      memory_scaling: this.optimizeMemoryAllocation(usagePatterns.memory),
      database_scaling: this.optimizeDatabaseAllocation(usagePatterns.database),
      cache_optimization: this.optimizeCacheConfiguration(usagePatterns.cache),
    };
  }

  private getRecommendedActions(prediction: any): string[] {
    const actions: string[] = [];
    
    switch (prediction.type) {
      case 'high_cpu':
        actions.push('Scale up CPU resources');
        actions.push('Optimize CPU-intensive operations');
        break;
      case 'memory_leak':
        actions.push('Investigate memory usage patterns');
        actions.push('Consider memory profiling');
        break;
      case 'slow_database':
        actions.push('Check database query performance');
        actions.push('Optimize database indexes');
        break;
      case 'high_response_time':
        actions.push('Analyze request handling bottlenecks');
        actions.push('Implement request batching');
        break;
    }
    
    return actions;
  }
}
```

### Distributed Tracing Implementation

```typescript
// Advanced distributed tracing with correlation
export class DistributedTracing {
  private tracer: Tracer;

  constructor() {
    this.tracer = trace.getTracer('distributed-tracing');
  }

  async traceWorkflow(workflowName: string, steps: WorkflowStep[]): Promise<void> {
    const mainSpan = this.tracer.startSpan(`workflow.${workflowName}`, {
      attributes: {
        'workflow.name': workflowName,
        'workflow.steps_count': steps.length.toString(),
      },
    });

    try {
      for (const step of steps) {
        const stepSpan = this.tracer.startSpan(`step.${step.name}`, {
          parent: mainSpan,
          attributes: {
            'step.name': step.name,
            'step.type': step.type,
            'step.service': step.service,
          },
        });

        try {
          await this.executeStep(step);
          
          stepSpan.setAttributes({
            'step.status': 'success',
            'step.duration': stepSpan.duration[0].toString(),
          });
        } catch (error) {
          stepSpan.recordException(error as Error);
          stepSpan.setAttributes({
            'step.status': 'error',
            'step.error': (error as Error).message,
          });
          throw error;
        } finally {
          stepSpan.end();
        }
      }
    } finally {
      mainSpan.end();
    }
  }

  private async executeStep(step: WorkflowStep): Promise<void> {
    // Add custom baggage for context propagation
    const baggage = propagate.getActiveBaggage();
    if (!baggage) {
      propagate.setBaggage(
        Baggage.fromEntries([
          ['workflow.id', crypto.randomUUID()],
          ['correlation.id', crypto.randomUUID()],
          ['user.id', step.context?.userId || 'anonymous'],
        ])
      );
    }

    // Execute the step with proper context
    await step.execute();
  }

  // Correlation analysis for distributed systems
  async analyzeCorrelations(traceData: TraceData[]): Promise<CorrelationAnalysis> {
    const correlations = new Map<string, CorrelationResult>();
    
    // Analyze trace patterns
    for (const trace of traceData) {
      const correlationId = trace.attributes['correlation.id'];
      
      if (correlationId) {
        const existing = correlations.get(correlationId) || {
          correlationId,
          spans: [],
          services: new Set(),
          errors: [],
          totalDuration: 0,
        };
        
        existing.spans.push(trace);
        existing.services.add(trace.attributes['service.name']);
        
        if (trace.attributes['error']) {
          existing.errors.push(trace);
        }
        
        existing.totalDuration += trace.duration || 0;
        correlations.set(correlationId, existing);
      }
    }
    
    return {
      totalCorrelations: correlations.size,
      correlationResults: Array.from(correlations.values()),
      errorRate: this.calculateErrorRate(correlations),
      averageDuration: this.calculateAverageDuration(correlations),
    };
  }
}
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Monitoring Operations
- `create_metric(name, type, labels)` - Create custom metric
- `record_event(event_name, attributes)` - Record business event
- `create_span(name, parent_span)` - Create tracing span
- `set_alert(condition, severity, channels)` - Configure alert
- `create_dashboard(metrics, visualization)` - Create monitoring dashboard

### Context7 Integration
- `get_latest_monitoring_documentation()` - Official monitoring docs via Context7
- `analyze_observability_patterns()` - Observability best practices via Context7
- `optimize_monitoring_stack()` - Monitoring optimization via Context7

## Best Practices (November 2025)

### DO
- Use OpenTelemetry for vendor-neutral observability
- Implement structured logging with correlation IDs
- Set up comprehensive alerting with proper escalation
- Monitor business metrics alongside technical metrics
- Use dashboards for real-time system visibility
- Implement anomaly detection for proactive monitoring
- Set up SLI/SLO monitoring for service reliability
- Use distributed tracing for microservice debugging

### DON'T
- Skip monitoring for development environments
- Create too many alerts without proper prioritization
- Ignore business metrics and user experience
- Forget to monitor infrastructure costs
- Use alerting as a replacement for proper monitoring
- Skip performance testing and benchmarking
- Ignore monitoring data retention policies
- Forget to secure monitoring endpoints and data

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS monitoring)
- `moai-essentials-perf` (Performance optimization)
- `moai-security-api` (Security monitoring)
- `moai-foundation-trust` (Compliance monitoring)
- `moai-domain-backend` (Backend application monitoring)
- `moai-domain-frontend` (Frontend performance monitoring)
- `moai-domain-devops` (DevOps and infrastructure monitoring)
- `moai-security-owasp` (Security threat monitoring)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 monitoring stack updates, and intelligent alerting patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, monitoring patterns, alerting configuration
- **v1.0.0** (2025-11-11): Initial application monitoring

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Monitoring Security
- Secure transmission of monitoring data with encryption
- Access controls for sensitive metrics and logs
- Data anonymization for user privacy protection
- Secure API endpoints for monitoring data collection

### Compliance Management
- GDPR compliance with data minimization in monitoring
- SOC2 monitoring controls and audit trails
- Industry-specific compliance monitoring (HIPAA, PCI-DSS)
- Automated compliance reporting and alerting

---

**End of Enterprise Application Monitoring Expert v4.0.0**
