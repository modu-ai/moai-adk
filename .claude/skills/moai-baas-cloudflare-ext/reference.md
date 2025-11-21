## API Reference

### Core Cloudflare Operations
- `create_worker(script, bindings)` - Deploy Workers script
- `create_pages_project(project_config)` - Create Pages project
- `create_durable_object(class_name, script_name)` - Deploy Durable Object
- `create_kv_namespace(name)` - Create KV storage namespace
- `create_d1_database(name)` - Create D1 database
- `configure_waf(rules)` - Configure WAF rules

### Context7 Integration
- `get_latest_cloudflare_documentation()` - Official Cloudflare docs via Context7
- `analyze_edge_performance_patterns()` - Edge optimization via Context7
- `optimize_workers_configuration()` - Workers best practices via Context7

## Best Practices (November 2025)

### DO
- Use Workers for compute-intensive tasks at the edge
- Implement comprehensive caching strategies
- Configure DDoS protection and WAF rules
- Use Durable Objects for real-time collaboration features
- Optimize global routing for minimum latency
- Monitor edge performance and costs
- Use appropriate storage (KV, D1, R2) based on use case
- Implement security headers and policies

### DON'T
- Deploy large applications to Workers (size limits apply)
- Skip security configuration for production
- Ignore edge computing costs and limits
- Use D1 for highly transactional workloads
- Forget to implement proper error handling
- Neglect monitoring and observability
- Overuse edge resources when origin is sufficient
- Skip compliance requirements for data residency

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture patterns)
- `moai-domain-frontend` (Frontend edge optimization)
- `moai-security-api` (API security implementation)
- `moai-essentials-perf` (Performance optimization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-vercel-ext` (Edge deployment comparison)
- `moai-baas-railway-ext` (Full-stack deployment alternative)
- `moai-domain-backend` (Backend edge optimization)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Cloudflare platform updates, and advanced edge security patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, edge patterns, Workers optimization
- **v1.0.0** (2025-11-11): Initial Cloudflare edge platform

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Edge Security Framework
- Zero-trust network architecture with Cloudflare Access
- Advanced DDoS protection with automatic mitigation
- Web Application Firewall with custom rules and ML protection
- Bot management and CAPTCHA integration

### Data Protection
- End-to-end encryption for all edge communications
- GDPR compliance with data processing at the edge
- Regional data residency with smart routing
- Comprehensive audit logging and monitoring

---

**End of Enterprise Cloudflare Edge Platform Expert **
