# Auth0 Security Specialist - Reference Documentation

## Official Documentation

### Core Documentation
- **Auth0 Secure Hub**: https://auth0.com/docs/secure
  - Central security documentation
  - Best practices overview
  - Implementation guides

- **Attack Protection**: https://auth0.com/docs/secure/attack-protection
  - Bot detection configuration
  - Brute force protection
  - Breached password detection
  - Suspicious IP throttling

- **Multi-Factor Authentication**: https://auth0.com/docs/secure/multi-factor-authentication
  - Factor setup and configuration
  - Guardian push notifications
  - WebAuthn and biometrics
  - Adaptive MFA

- **Token Management**: https://auth0.com/docs/secure/tokens
  - JWT fundamentals
  - Access token configuration
  - Refresh token strategies
  - Token revocation

- **Sender Constraining**: https://auth0.com/docs/secure/sender-constraining
  - DPoP implementation
  - mTLS configuration
  - Certificate binding

- **Compliance**: https://auth0.com/docs/secure/data-privacy-and-compliance
  - GDPR compliance
  - HIPAA guidelines
  - FAPI implementation
  - SOC 2 and ISO 27001

### API Reference

#### Management API
- **Management API v2**: https://auth0.com/docs/api/management/v2
  - Endpoints for tenant configuration
  - Attack protection management
  - MFA factor configuration
  - User and role management

- **Authentication API**: https://auth0.com/docs/api/authentication
  - OAuth 2.0 flows
  - Token endpoints
  - MFA challenge flows

#### SDKs & Libraries
- **Auth0 Node.js SDK**: https://github.com/auth0/node-auth0
  - Management API client
  - Authentication SDK
  - Configuration examples

- **Auth0 Python SDK**: https://github.com/auth0/auth0-python
  - Python management API
  - Async support
  - Type hints

- **Auth0 Go SDK**: https://github.com/auth0/auth0
  - Go management API
  - Context support
  - Error handling

### Dashboard & Configuration

- **Auth0 Dashboard**: https://manage.auth0.com/
  - Tenant configuration
  - Application settings
  - Security rules
  - Monitoring and logs

- **Attack Protection Settings**: Dashboard > Security > Attack Protection
  - Bot detection configuration
  - Brute force thresholds
  - IP throttling rules

- **MFA Configuration**: Dashboard > Security > Multi-factor Auth
  - Factor selection
  - Guardian setup
  - Policy configuration

- **Security Center**: Dashboard > Security > Security Center
  - Threat monitoring
  - Attack analytics
  - Security metrics

### Security Standards

#### FAPI (Financial API)
- **FAPI 1.0 Specification**: https://openid.net/specs/openid-financial-api-part-1-1_0.html
  - Security requirements
  - Implementation guidelines
  - Compliance testing

- **Auth0 FAPI Guide**: https://auth0.com/docs/secure/attack-protection#highly-regulated-industries
  - JAR and PAR configuration
  - Private key JWT
  - mTLS authentication

#### GDPR Compliance
- **Auth0 GDPR Guide**: https://auth0.com/docs/secure/data-privacy-and-compliance/gdpr
  - Data processing agreements
  - User rights implementation
  - Consent management
  - Breach notification

- **Data Export**: https://auth0.com/docs/api/management/v2#!/Users/export_users
  - User data portability
  - JSON export format
  - Batch operations

#### HIPAA Compliance
- **HIPAA Guide**: https://auth0.com/docs/secure/data-privacy-and-compliance/hipaa
  - BAA availability
  - PHI handling
  - Audit logging

### Certifications & Compliance

- **Compliance Matrix**: https://www.auth0.com/compliance
  - ISO 27001/27017/27018
  - SOC 2 Type 2
  - CSA STAR Level 2
  - PCI DSS
  - HIPAA/HITECH

### Module Organization

This skill contains 39 modules organized into 6 categories:

#### Attack Protection (8 modules)
- `modules/attack-protection-overview.md` - Complete attack protection framework
- `modules/bot-detection.md` - CAPTCHA and bot detection
- `modules/breached-password-detection.md` - Compromised credential screening
- `modules/brute-force-protection.md` - Login attempt limits
- `modules/suspicious-ip-throttling.md` - Velocity-based throttling
- `modules/akamai-integration.md` - Third-party integration
- `modules/attack-protection-log-events.md` - Event monitoring
- `modules/state-parameters.md` - State management

#### Multi-Factor Authentication (9 modules)
- `modules/mfa-overview.md` - MFA fundamentals
- `modules/mfa-factors.md` - Available factors
- `modules/webauthn-fido.md` - Passwordless authentication
- `modules/adaptive-mfa.md` - Risk-based MFA
- `modules/guardian-configuration.md` - Push notifications
- `modules/step-up-authentication.md` - Enhanced verification
- `modules/mfa-api-management.md` - API-based MFA
- `modules/customize-mfa.md` - Custom MFA flows
- `modules/ropg-flow-mfa.md` - Resource owner password grant

#### Token Security (8 modules)
- `modules/tokens-overview.md` - Token types and usage
- `modules/jwt-fundamentals.md` - JWT structure and validation
- `modules/id-tokens.md` - User identity tokens
- `modules/access-tokens.md` - API authorization
- `modules/delegation-tokens.md` - Token delegation
- `modules/refresh-tokens.md` - Session management
- `modules/token-revocation.md` - Token invalidation
- `modules/token-best-practices.md` - Security guidelines

#### Sender Constraining (2 modules)
- `modules/dpop-implementation.md` - DPoP configuration
- `modules/mtls-sender-constraining.md` - mTLS binding

#### Compliance (8 modules)
- `modules/compliance-overview.md` - Compliance framework
- `modules/fapi-implementation.md` - Financial API security
- `modules/highly-regulated-identity.md` - HRI features
- `modules/gdpr-compliance.md` - GDPR requirements
- `modules/certifications.md` - Official certifications
- `modules/tenant-access-control.md` - Access governance
- `modules/customer-managed-keys.md` - Key management

#### Security Operations (4 modules)
- `modules/security-center.md` - Monitoring dashboard
- `modules/application-credentials.md` - Client secrets and keys
- `modules/continuous-session-protection.md` - Session monitoring
- `modules/security-guidance.md` - Best practices
- `modules/mdl-verification.md` - Multi-device login

### Community & Support

- **Auth0 Community**: https://community.auth0.com/
  - Forums and discussions
  - Q&A threads
  - Community solutions

- **Auth0 GitHub**: https://github.com/auth0
  - Open-source SDKs
  - Sample implementations
  - Integration examples

- **Support Portal**: https://support.auth0.com/
  - Ticket submission
  - Knowledge base
  - Status page

### Training & Certification

- **Auth0 University**: https://auth0.com/university
  - Free training courses
  - Certification programs
  - Video tutorials

### Related Skills

- **moai-platform-clerk**: Modern authentication alternative
- **moai-platform-firebase-auth**: Firebase authentication
- **moai-domain-backend**: Backend security patterns
- **moai-platform-supabase**: Supabase authentication

---

**Last Updated**: 2025-12-24
**Skill Version**: 1.0.0
**Total Modules**: 39
**Categories**: 6
