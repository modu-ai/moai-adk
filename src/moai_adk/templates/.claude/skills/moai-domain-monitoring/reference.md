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

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 monitoring stack updates, and intelligent alerting patterns
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

**End of Enterprise Application Monitoring Expert **
