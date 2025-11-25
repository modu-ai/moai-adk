## API Reference

### Core Clerk Operations
- `create_user(email, password)` - Create new user account
- `create_organization(name, slug)` - Create organization
- `invite_to_organization(org_id, email, role)` - Invite user to organization
- `add_webauthn_credential(user_id, credential)` - Add security key
- `generate_m2m_token(template)` - Generate machine-to-machine token

### Context7 Integration
- `get_latest_clerk_documentation()` - Official Clerk docs via Context7
- `analyze_modern_auth_patterns()` - Modern authentication via Context7
- `optimize_user_experience()` - UX best practices via Context7

## Best Practices (November 2025)

### DO
- Use Clerk components for consistent user experience
- Implement proper session management and security
- Configure organization features for multi-tenant applications
- Enable WebAuthn for enhanced security
- Customize appearance to match your brand
- Monitor authentication events and user activity
- Implement proper error handling for auth flows
- Use M2M tokens for service-to-service authentication

### DON'T
- Skip security configuration for production
- Ignore user experience optimization opportunities
- Forget to configure organization permissions properly
- Use hardcoded secrets or API keys
- Neglect monitoring and analytics
- Skip accessibility considerations in auth UI
- Forget to implement proper logout and session cleanup
- Ignore compliance requirements for user data

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture patterns)
- `moai-security-api` (API security and authorization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-auth0-ext` (Enterprise authentication comparison)
- `moai-domain-frontend` (Frontend auth integration)
- `moai-essentials-perf` (Authentication performance optimization)
- `moai-domain-backend` (Backend auth integration)
- `moai-security-encryption` (Data protection and encryption)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Clerk platform updates, and advanced WebAuthn implementation
- **v2.0.0** (2025-11-11): Complete metadata structure, auth patterns, organization management
- **v1.0.0** (2025-11-11): Initial Clerk authentication platform

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Modern Security Framework
- Multi-factor authentication with WebAuthn support
- Advanced session management with device fingerprinting
- Real-time threat detection and anomaly analysis
- Comprehensive audit logging and compliance reporting

### Data Protection
- GDPR compliance with data portability and deletion
- SOC2 Type II security controls
- Advanced encryption for sensitive authentication data
- Regional data residency with smart routing

---

**End of Enterprise Clerk Authentication Platform Expert **
