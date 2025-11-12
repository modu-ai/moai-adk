# Reference: Official Documentation & Resources

## Prometheus & Metrics

**Official Documentation**:
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/)
- [PromQL Query Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Service Discovery](https://prometheus.io/docs/prometheus/latest/configuration/service-discovery/)

**Stable Versions**:
- Prometheus 3.7.x (latest) - [Release Notes](https://github.com/prometheus/prometheus/releases)
- Prometheus 2.55.x (legacy support)
- Alertmanager 0.27.x - [Documentation](https://prometheus.io/docs/alerting/latest/overview/)

**Best Practices**:
- [Best Practices for Naming & Labeling](https://prometheus.io/docs/practices/naming/)
- [Cardinality Explosion Prevention](https://prometheus.io/docs/guides/avoiding-cardinality-explosion/)
- [Recording Rules & Alerts](https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/)

---

## Grafana

**Official Documentation**:
- [Grafana 11.x Documentation](https://grafana.com/docs/grafana/latest/)
- [Grafana Scenes Library](https://grafana.com/docs/grafana/latest/developers-alt/kinds/composable-dashboard-and-ui/)
- [Dashboard Provisioning](https://grafana.com/docs/grafana/latest/dashboards/manage-dashboards/)
- [Alerting Rules](https://grafana.com/docs/grafana/latest/alerting/)

**Version History**:
- Grafana 11.x GA: [Release Announcement (April 2024)](https://grafana.com/blog/2024/04/09/grafana-11-release-all-the-new-features/)
- [Grafana Release Notes](https://grafana.com/docs/grafana/latest/release-notes/)
- [Explore to Drilldown Migration](https://grafana.com/docs/grafana/latest/explore/)

**Key Features**:
- [Scenes-Powered Dashboards](https://grafana.com/docs/grafana/latest/developers-alt/kinds/composable-dashboard-and-ui/)
- [Edit Mode & Dashboard Discovery](https://grafana.com/docs/grafana/latest/dashboards/)
- [PDF Export](https://grafana.com/docs/grafana/latest/dashboards/share-dashboards-panels/)

---

## OpenTelemetry & Distributed Tracing

**Official Documentation**:
- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [OpenTelemetry JavaScript SDK](https://opentelemetry.io/docs/instrumentation/js/)
- [OpenTelemetry Python SDK](https://opentelemetry.io/docs/instrumentation/python/)
- [OpenTelemetry Java SDK](https://opentelemetry.io/docs/instrumentation/java/)

**Stable Status**:
- OpenTelemetry 1.33.x (stable, production-ready)
- [OTel Profiling Support - Generally Available](https://opentelemetry.io/blog/2024/profiling-ga/)
- [OTel Spring Boot Starter - Stable (Sept 2024)](https://opentelemetry.io/docs/instrumentation/java/automatic/spring-boot-starter/)

**Propagation Standards**:
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
- [Jaeger Propagation](https://www.jaegertracing.io/)
- [B3 Multi Header](https://github.com/openzipkin/b3-propagation)

---

## Jaeger

**Official Documentation**:
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Jaeger Architecture](https://www.jaegertracing.io/docs/latest/architecture/)
- [Jaeger Deployment](https://www.jaegertracing.io/docs/latest/deployment/)
- [UI Query](https://www.jaegertracing.io/docs/latest/frontend-ui/)

**Version Information**:
- Jaeger 1.62.x (latest stable)
- [Jaeger v2 Release Notes](https://www.jaegertracing.io/docs/latest/roadmap/)
- [Multi-Tenancy Support](https://www.jaegertracing.io/docs/latest/deployment/#multi-tenancy)

**Performance & Features**:
- OTLP Protocol Support
- Elasticsearch Backend Integration
- Badger Storage Option
- gRPC & HTTP Collectors

---

## Loki

**Official Documentation**:
- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/query/)
- [Loki Architecture](https://grafana.com/docs/loki/latest/design-documents/multi-tenancy/)
- [Deployment Guide](https://grafana.com/docs/loki/latest/setup/install/)

**Version & Features**:
- Loki 3.x (latest) - [v3.0 Release (April 2024)](https://grafana.com/blog/2024/04/09/grafana-loki-3.0-release-all-the-new-features/)
- [Bloom Filters for Query Acceleration](https://grafana.com/docs/loki/latest/query/bloom_filters/)
- [Native OTLP Support](https://grafana.com/docs/loki/latest/send-data/otel/)
- [Loki 3.1 Release Notes](https://grafana.com/docs/loki/latest/release-notes/v3-1/)

**Optimization**:
- [Cost Optimization Guide](https://grafana.com/docs/loki/latest/operations/cost-optimization/)
- [TTL & Retention](https://grafana.com/docs/loki/latest/operations/storage/retention/)

---

## Elasticsearch & Kibana

**Official Documentation**:
- [Elasticsearch 8.x Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Kibana Documentation](https://www.elastic.co/guide/en/kibana/current/index.html)
- [Logstash Documentation](https://www.elastic.co/guide/en/logstash/current/index.html)
- [Elastic Beats Documentation](https://www.elastic.co/guide/en/beats/libbeat/current/index.html)

**Stack Monitoring**:
- [Stack Monitoring in Kibana](https://www.elastic.co/guide/en/kibana/current/monitor-logs.html)
- [Elasticsearch Monitoring](https://www.elastic.co/guide/en/elasticsearch/reference/8.0/monitoring-overview.html)
- [Index Lifecycle Management (ILM)](https://www.elastic.co/guide/en/elasticsearch/reference/8.0/ilm-get-started.html)

**Version Status**:
- Elasticsearch 8.17 (current stable)
- Kibana 8.x (current stable)
- [End of Life Timeline](https://www.elastic.co/support/eol)

---

## Sentry

**Official Documentation**:
- [Sentry Documentation](https://docs.sentry.io/)
- [Sentry SDK Guides](https://docs.sentry.io/platforms/)
- [JavaScript SDK](https://docs.sentry.io/platforms/javascript/configuration/)
- [Python SDK](https://docs.sentry.io/platforms/python/)

**Features & Best Practices**:
- [Performance Monitoring](https://docs.sentry.io/product/performance/)
- [Error Tracking](https://docs.sentry.io/product/error-monitoring/)
- [Sampling Strategy](https://docs.sentry.io/product/performance/sampling/)
- [Alerts & Rules](https://docs.sentry.io/product/alerts-notifications/)

**Version**:
- Sentry 24.x (stable, 2025 November)
- [Release Notes](https://github.com/getsentry/sentry/releases)

---

## New Relic & Datadog

### New Relic

**Documentation**:
- [New Relic APM](https://docs.newrelic.com/docs/apm/)
- [New Relic Node.js Agent](https://docs.newrelic.com/docs/apm/agents/nodejs-agent/)
- [Custom Metrics](https://docs.newrelic.com/docs/data-apis/custom-data/custom-events/)
- [AI-Powered Insights](https://docs.newrelic.com/docs/alerts-applied-intelligence/)

**Features**:
- Code-level diagnostics
- AI-powered anomaly detection
- Full-stack observability
- Unified APM/Infrastructure/Logs

### Datadog

**Documentation**:
- [Datadog APM](https://docs.datadoghq.com/tracing/)
- [Datadog Infrastructure](https://docs.datadoghq.com/infrastructure/)
- [Datadog Logs](https://docs.datadoghq.com/logs/)
- [Datadog Synthetics](https://docs.datadoghq.com/synthetics/)

**Features**:
- Infrastructure-first approach
- Unified dashboarding
- Service maps & dependency tracking
- Continuous profiling

**Comparison 2025**:
- [New Relic vs Datadog - 2025 Guide](https://signoz.io/blog/datadog-vs-newrelic/)
- [Better Stack Comparison](https://betterstack.com/community/comparisons/datadog-vs-newrelic/)

---

## Infrastructure Monitoring

### Prometheus Node Exporter

**Documentation**:
- [Node Exporter GitHub](https://github.com/prometheus/node_exporter)
- [Collectors Documentation](https://github.com/prometheus/node_exporter#collectors)
- [Textfile Collector](https://github.com/prometheus/node_exporter#textfile-collector)

### cAdvisor

**Documentation**:
- [cAdvisor GitHub](https://github.com/google/cadvisor)
- [Container Metrics](https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md)
- [Docker Integration](https://github.com/google/cadvisor/blob/master/docs/running.md)

### Telegraf

**Documentation**:
- [Telegraf Documentation](https://docs.influxdata.com/telegraf/latest/)
- [Input Plugins](https://docs.influxdata.com/telegraf/latest/plugins/#input-plugins)
- [Output Plugins](https://docs.influxdata.com/telegraf/latest/plugins/#output-plugins)

### netdata

**Documentation**:
- [netdata Documentation](https://learn.netdata.cloud/)
- [Real-time Monitoring](https://learn.netdata.cloud/docs/cloud/overview)
- [Parent-Child Architecture](https://learn.netdata.cloud/docs/agent/running-behind-nginx)

---

## Thanos - High Availability

**Documentation**:
- [Thanos Documentation](https://thanos.io/tip/thanos/)
- [Query Component](https://thanos.io/tip/thanos/components/query.md/)
- [Compact Component](https://thanos.io/tip/thanos/components/compact.md/)
- [S3 Backend](https://thanos.io/tip/thanos/storage.md/#s3)

**Version**:
- Thanos 0.36.x (stable)

---

## SLO/SLI Implementation

**Resources**:
- [Site Reliability Engineering - Google SRE Book](https://sre.google/books/)
- [SLO Fundamentals](https://sre.google/sre-book/service-level-objectives/)
- [Error Budget](https://sre.google/sre-book/error-budgets/)

**Tools & Standards**:
- [Service Level Indicators](https://thenewstack.io/monitoring-and-observability-best-practices/)
- [Error Budget Implementation](https://www.thenewstack.io/measuring-error-budgets/)

---

## Observability Best Practices

**Key Resources**:
- [CNCF Observability White Paper](https://www.cncf.io/blog/2021/10/07/observability-whitepaper/)
- [Observability Engineering - O'Reilly Book](https://www.oreilly.com/library/view/observability-engineering/9781492076438/)
- [Three Pillars of Observability: Logs, Metrics, Traces](https://www.splunk.com/en_us/blog/learn/observability-vs-monitoring-logs-metrics-traces.html)

**Industry Standards**:
- [OTEL Conventions](https://opentelemetry.io/docs/specs/semconv/)
- [Cloud Native Computing Foundation (CNCF)](https://www.cncf.io/)

---

## Community & Support

**GitHub Repositories**:
- [Prometheus GitHub](https://github.com/prometheus)
- [Grafana GitHub](https://github.com/grafana)
- [OpenTelemetry GitHub](https://github.com/open-telemetry)
- [Jaeger GitHub](https://github.com/jaegertracing)
- [Loki GitHub](https://github.com/grafana/loki)

**Community Forums**:
- [Prometheus Community Slack](https://prometheus.io/community/)
- [Grafana Community Slack](https://grafana.com/docs/grafana/latest/community/)
- [CNCF Slack](https://www.cncf.io/slack/)

**Conferences & Events**:
- [GrafanaCon](https://grafana.com/about/events/grafanacon/)
- [PromCon](https://promcon.io/)
- [KubeCon](https://www.cncf.io/kubecon-cloudnativecon-events/)

