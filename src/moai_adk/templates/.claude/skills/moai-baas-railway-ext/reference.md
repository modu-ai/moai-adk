## API Reference

### Core Railway Operations
- `deploy_service(project_id, service_config)` - Deploy service
- `create_database(project_id, db_type)` - Provision database
- `scale_service(service_id, replicas, resources)` - Scale service
- `rollback_deployment(service_id, deployment_id)` - Rollback deployment
- `set_environment_variables(project_id, variables)` - Set environment variables

### Context7 Integration
- `get_latest_railway_documentation()` - Official Railway docs via Context7
- `analyze_container_optimization()` - Container best practices via Context7
- `optimize_deployment_strategy()` - Deployment patterns via Context7

## Best Practices (November 2025)

### DO
- Use separate environments for development, staging, and production
- Implement comprehensive health checks for all services
- Configure proper logging and monitoring for observability
- Use connection pooling for database connections
- Set up automated testing before deployments
- Monitor costs and implement spending limits
- Use volume mounts for persistent data storage
- Implement proper error handling and retry logic

### DON'T
- Hardcode environment variables in application code
- Skip health checks and monitoring setup
- Use production database for development testing
- Ignore scaling limits and cost controls
- Deploy without proper testing
- Forget to implement backup strategies
- Overprovision resources without optimization
- Skip security configuration for production

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture patterns)
- `moai-domain-backend` (Backend deployment patterns)
- `moai-domain-devops` (DevOps and CI/CD workflows)
- `moai-essentials-perf` (Performance optimization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-vercel-ext` (Frontend deployment comparison)
- `moai-baas-neon-ext` (PostgreSQL database integration)
- `moai-domain-database` (Database optimization)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Railway platform updates, and advanced deployment automation
- **v2.0.0** (2025-11-11): Complete metadata structure, deployment patterns, CI/CD integration
- **v1.0.0** (2025-11-11): Initial Railway full-stack platform

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Container Security
- Secure base images and vulnerability scanning
- Runtime security monitoring and threat detection
- Network isolation and firewall configuration
- Secret management with encrypted environment variables

### Compliance Management
- GDPR compliance with data protection measures
- SOC2 Type II security controls
- Automated security scanning and patching
- Comprehensive audit logging and monitoring

---

**End of Enterprise Railway Full-Stack Platform Expert **
