# SPEC Examples: Real-World Scenarios

Explore comprehensive examples of SPEC documents across different domains and complexity levels. These examples demonstrate best practices for writing clear, testable, and complete specifications.

## Example Categories

1. **Authentication & Security** - User authentication, authorization, and security features
2. **API Development** - REST API endpoints, data models, and integration patterns
3. **User Management** - User profiles, preferences, and account management
4. **Data Processing** - Background jobs, data pipelines, and analytics
5. **E-commerce** - Shopping carts, payments, and order management
6. **Communication** - Notifications, messaging, and email systems
7. **File Management** - Upload, storage, and processing of files
8. **Reporting & Analytics** - Dashboards, reports, and data visualization

---

## Example 1: Authentication System (Medium Complexity)

**Domain**: AUTH-001
**Estimated Hours**: 12
**Priority**: High

```yaml
---
id: AUTH-001
version: 0.2.0
status: in_review
priority: high
created: 2025-01-15
updated: 2025-01-16
author: @security-team
domain: authentication
complexity: medium
estimated_hours: 12
dependencies: [USER-001, EMAIL-001, RATE-001]
tags: [security, jwt, oauth, mfa]
reviewers: [@tech-lead, @security-expert]
milestone: "Sprint 23 - Q1 2025"
---

# @SPEC:EX-AUTH-001: Comprehensive Authentication System

## Overview
Provide secure, scalable user authentication with multiple authentication methods, session management, and security features including multi-factor authentication and social login integration.

## EARS Requirements

### Ubiquitous Requirements
- The system SHALL provide user authentication via email and password
- The system SHALL support JWT token-based session management
- The system SHALL maintain secure user session state
- The system SHALL log all authentication attempts with timestamps
- The system SHALL enforce secure password policies
- The system SHALL protect against common authentication attacks

### Event-driven Requirements
- WHEN valid credentials are provided, the system SHALL issue JWT access and refresh tokens
- WHEN invalid credentials are provided, the system SHALL return 401 error with generic message
- WHEN multiple failed login attempts occur (5+ within 15 minutes), the system SHALL implement rate limiting
- WHEN password reset is requested, the system SHALL send secure reset link via email
- WHEN multi-factor authentication is enabled, the system SHALL require additional verification
- WHEN social login callback is received, the system SHALL create or link user account
- WHEN session expires, the system SHALL require re-authentication
- WHEN suspicious activity is detected, the system SHALL trigger security alerts

### State-driven Requirements
- WHILE a user is authenticated, the system SHALL allow access to protected resources
- WHILE a user session is active, the system SHALL automatically refresh tokens before expiry
- WHILE multi-factor authentication setup is pending, the system SHALL restrict sensitive operations
- WHILE account is locked due to security concerns, the system SHALL deny all authentication attempts
- WHILE rate limiting is active, the system SHALL reject excess authentication requests

### Optional Requirements
- WHERE social login providers are configured, the system SHALL support OAuth 2.0 authentication
- WHERE biometric authentication is available, the system SHALL support fingerprint and face recognition
- WHERE hardware security keys are supported, the system SHALL implement WebAuthn protocol
- WHERE risk-based authentication is enabled, the system SHALL adjust security requirements based on context
- WHERE single sign-on (SSO) is configured, the system SHALL integrate with identity providers

### Unwanted Behaviors
- The system SHALL NOT store passwords in plain text or reversible encryption
- The system SHALL NOT reveal whether an email address is registered in lockout scenarios
- Passwords SHALL NOT be less than 8 characters or contain common dictionary words
- The system SHALL NOT allow concurrent sessions with the same credentials (configurable)
- JWT tokens SHALL NOT contain sensitive user information in payload
- The system SHALL NOT accept authentication requests over unencrypted connections
- Login attempts SHALL NOT exceed 10 per minute per IP address
- The system SHALL NOT allow password reuse within last 5 passwords

## Technical Requirements

### Security Requirements
- Passwords SHALL be hashed using bcrypt with minimum 12 rounds
- JWT tokens SHALL use RS256 signing algorithm with 2048-bit keys
- All authentication endpoints SHALL use HTTPS exclusively
- Input validation SHALL prevent SQL injection, XSS, and CSRF attacks
- Session tokens SHALL have configurable expiration (default: 15 minutes access, 7 days refresh)
- Failed authentication attempts SHALL be logged with IP, user agent, and timestamp
- Account lockout SHALL occur after 10 failed attempts with 30-minute lockout duration

### Performance Requirements
- Authentication response time SHALL be under 500ms (95th percentile)
- Token validation SHALL be under 100ms
- System SHALL support 10,000 concurrent authentication requests
- Database queries SHALL be optimized for authentication lookups
- Password verification SHALL use constant-time comparison

### Integration Requirements
- Email service integration for password reset and verification
- Rate limiting service integration for brute force protection
- Audit logging integration for compliance requirements
- Social login provider integration (Google, GitHub, Microsoft)
- Multi-factor authentication service integration (SMS, authenticator apps)

### Data Requirements
- Users SHALL have unique email addresses
- Passwords SHALL meet complexity requirements (8+ chars, mixed case, numbers, symbols)
- User sessions SHALL be tracked with device fingerprinting
- Authentication events SHALL be retained for 90 days for audit purposes
- PII SHALL be encrypted at rest and in transit

## Acceptance Criteria

### Functional Requirements
- [ ] Users can register with email and password
- [ ] Users can authenticate with valid credentials
- [ ] System issues JWT tokens with proper expiration
- [ ] Users can refresh tokens before expiry
- [ ] Password reset functionality works end-to-end
- [ ] Multi-factor authentication can be enabled and used
- [ ] Social login integration works with configured providers
- [ ] Rate limiting prevents brute force attacks
- [ ] Account lockout protects against repeated failures
- [ ] Session management works across multiple devices

### Security Requirements
- [ ] Passwords are properly hashed and salted
- [ ] JWT tokens use secure signing algorithm
- [ ] All endpoints are protected by HTTPS
- [ ] Input validation prevents injection attacks
- [ ] Sensitive data is encrypted at rest
- [ ] Audit logs capture all authentication events
- [ ] Error messages don't reveal sensitive information
- [ ] Rate limiting is effective and configurable
- [ ] Session hijacking protection is implemented
- [ ] CSRF protection is enabled for web forms

### Performance Requirements
- [ ] Authentication response time < 500ms under load
- [ ] System handles 10,000 concurrent requests
- [ ] Database queries are optimized and indexed
- [ ] Token validation is cached and fast
- [ ] Memory usage remains stable under load

### Integration Requirements
- [ ] Email service integration works reliably
- [ ] Social login providers authenticate correctly
- [ ] MFA services generate and verify codes
- [ ] Audit logging captures all required events
- [ ] Rate limiting service is properly integrated

## Dependencies

### Internal Dependencies
- **USER-001**: User management system (required for authentication)
- **EMAIL-001**: Email notification service (required for password reset)
- **RATE-001**: Rate limiting service (required for security)
- **AUDIT-001**: Audit logging system (required for compliance)

### External Dependencies
- **OAuth Providers**: Google, GitHub, Microsoft (optional social login)
- **MFA Services**: Twilio (SMS), Authenticator apps (TOTP)
- **Email Service**: SendGrid, AWS SES (transactional emails)
- **Monitoring**: Datadog, New Relic (performance monitoring)

## Risk Assessment

### High Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Token security breach | Critical | Medium | Use RS256, rotate keys, implement token blacklisting |
| Password database compromise | Critical | Low | Strong hashing, encryption at rest, limited access |
| Social login provider outage | High | Medium | Fallback to email authentication, monitor providers |

### Medium Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Performance under high load | Medium | Medium | Load testing, caching, horizontal scaling |
| MFA delivery failures | Medium | Low | Multiple MFA options, fallback methods |
| User enumeration attacks | Medium | High | Generic error messages, rate limiting |

### Low Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Session management edge cases | Low | Medium | Comprehensive testing, clear session policies |
| Password policy user friction | Low | High | User-friendly password requirements, strength meter |

## Implementation Notes

### Architecture Recommendations
@EXPERT:BACKEND - Use microservices architecture with separate authentication service
@EXPERT:SECURITY - Implement zero-trust architecture with principle of least privilege
@EXPERT:PERF - Use Redis for session storage and caching
@EXPERT:DEVOPS - Deploy with blue-green deployment for zero downtime

### Security Considerations
- Implement account takeover detection
- Use device fingerprinting for anomaly detection
- Implement progressive authentication based on risk
- Regular security audits and penetration testing
- Compliance with GDPR, CCPA, and other regulations

### Testing Strategy
- Unit tests for all authentication logic (90%+ coverage)
- Integration tests for external service dependencies
- Security tests for common attack vectors
- Performance tests under various load conditions
- Chaos engineering for failure scenarios

### Monitoring and Alerting
- Authentication success/failure rates
- Response time percentiles (p50, p95, p99)
- Security event alerts and dashboards
- Error rates and patterns
- Resource utilization (CPU, memory, database)

## Success Metrics

### Technical Metrics
- Authentication success rate > 99.5%
- Average response time < 500ms
- System availability > 99.9%
- Zero security breaches
- Test coverage > 90%

### Business Metrics
- User registration completion rate > 80%
- Login success rate > 95%
- Support tickets related to authentication < 5%
- User satisfaction score > 4.5/5
- MFA adoption rate > 60%

## History

### Version 0.2.0 (2025-01-16)
- Added social login integration requirements
- Enhanced MFA support with hardware security keys
- Improved performance requirements
- Added risk-based authentication options

### Version 0.1.0 (2025-01-15)
- Initial specification creation
- Core authentication requirements
- Security and performance baseline
```

---

## Example 2: File Upload API (Simple Complexity)

**Domain**: API-002
**Estimated Hours**: 6
**Priority**: Medium

```yaml
---
id: API-002
version: 0.1.0
status: draft
priority: medium
created: 2025-01-16
author: @backend-team
domain: api
complexity: simple
estimated_hours: 6
dependencies: [STORAGE-001, VIRUS-001]
tags: [api, file-upload, storage, validation]
reviewers: [@tech-lead]
milestone: "Sprint 23 - Q1 2025"
---

# @SPEC:EX-API-002: File Upload API

## Overview
Provide secure file upload functionality with validation, virus scanning, and cloud storage integration.

## EARS Requirements

### Ubiquitous Requirements
- The system SHALL provide file upload endpoint
- The system SHALL validate file types and sizes
- The system SHALL scan uploaded files for viruses
- The system SHALL store files securely in cloud storage
- The system SHALL return file metadata and access URLs

### Event-driven Requirements
- WHEN file is uploaded, the system SHALL validate format and size limits
- WHEN validation passes, the system SHALL initiate virus scan
- WHEN virus scan completes successfully, the system SHALL store file permanently
- WHEN virus scan fails, the system SHALL delete file and return error
- WHEN file upload exceeds size limit, the system SHALL return 413 error

### Unwanted Behaviors
- The system SHALL NOT accept executable files
- File uploads SHALL NOT exceed 100MB
- The system SHALL NOT store files without virus scanning
- File names SHALL NOT contain special characters or path traversal

## Technical Requirements

### API Endpoint
- **Method**: POST
- **Path**: /api/v2/files/upload
- **Content-Type**: multipart/form-data
- **Authentication**: Required (JWT token)

### Validation Rules
- Allowed formats: pdf, doc, docx, jpg, jpeg, png, gif
- Maximum file size: 100MB
- Maximum filename length: 255 characters
- Required metadata: file description, category

### Security Requirements
- All uploads require authenticated user
- Files scanned before permanent storage
- File access URLs expire after 1 hour
- Upload rate limiting: 10 files per minute per user

## Acceptance Criteria

- [ ] Users can upload files through API endpoint
- [ ] File validation prevents invalid formats
- [ ] Virus scanning protects against malicious files
- [ ] Cloud storage integration works reliably
- [ ] File access URLs are secure and temporary
- [ ] Error handling covers all failure scenarios
- [ ] Performance requirements are met under load

## Dependencies

- **STORAGE-001**: Cloud storage service (AWS S3 integration)
- **VIRUS-001**: Virus scanning service (ClamAV integration)

## Risk Assessment

### Medium Risk
- Cloud storage service outage
- Virus scanning service performance
- Large file upload performance

### Low Risk
- File format validation complexity
- User permission management
```

---

## Example 3: User Profile Management (High Complexity)

**Domain**: USER-002
**Estimated Hours**: 20
**Priority**: High

```yaml
---
id: USER-002
version: 0.1.0
status: draft
priority: high
created: 2025-01-16
author: @frontend-team
domain: user
complexity: high
estimated_hours: 20
dependencies: [AUTH-001, NOTIF-001, VALID-001]
tags: [user-profile, preferences, privacy, gdpr]
reviewers: [@tech-lead, @privacy-expert, @ux-lead]
milestone: "Sprint 24 - Q1 2025"
---

# @SPEC:EX-USER-002: Comprehensive User Profile Management

## Overview
Provide complete user profile management system with personal information, preferences, privacy controls, and GDPR compliance features.

## EARS Requirements

### Ubiquitous Requirements
- The system SHALL provide user profile management interface
- The system SHALL maintain user personal information securely
- The system SHALL support user preference management
- The system SHALL provide privacy control options
- The system SHALL comply with GDPR requirements

### Event-driven Requirements
- WHEN user updates profile information, the system SHALL validate and save changes
- WHEN user changes privacy settings, the system SHALL apply changes immediately
- WHEN user requests data export, the system SHALL generate comprehensive data package
- WHEN user requests account deletion, the system SHALL initiate data removal process
- WHEN user profile is viewed by others, the system SHALL apply privacy settings

### State-driven Requirements
- WHILE user profile is being edited, the system SHALL prevent concurrent modifications
- WHILE data export is being processed, the system SHALL provide progress updates
- WHILE account deletion is in progress, the system SHALL restrict access to non-essential features

### Optional Requirements
- WHERE social features are enabled, the system SHALL provide public profile options
- WHERE premium features are available, the system SHALL offer enhanced profile customization
- WHERE analytics are enabled, the system SHALL provide profile completion insights

### Unwanted Behaviors
- The system SHALL NOT expose private information without explicit consent
- Profile changes SHALL NOT require email verification for non-sensitive fields
- The system SHALL NOT retain data longer than legally required after deletion
- Profile data SHALL NOT be accessible to unauthorized users

## Technical Requirements

### Data Model
- **Profile Information**: Name, bio, avatar, location, website
- **Preferences**: Language, timezone, notification settings, privacy options
- **Privacy Controls**: Profile visibility, data sharing preferences, searchability
- **GDPR Features**: Data portability, right to be forgotten, consent management

### API Endpoints
- GET /api/v2/users/profile - Retrieve current user profile
- PUT /api/v2/users/profile - Update profile information
- GET /api/v2/users/preferences - Retrieve user preferences
- PUT /api/v2/users/preferences - Update preferences
- POST /api/v2/users/export-data - Request data export
- DELETE /api/v2/users/account - Request account deletion

### Security Requirements
- All profile operations require authentication
- Sensitive data encrypted at rest
- Profile changes logged for audit
- Rate limiting on profile updates
- Input validation and sanitization

### GDPR Compliance
- Clear consent for data processing
- Easy data export functionality
- Complete account deletion process
- Data retention policies
- Right to access and correction

## Acceptance Criteria

### Functional Requirements
- [ ] Users can view and edit profile information
- [ ] Users can manage privacy settings
- [ ] System validates all profile inputs
- [ ] Profile changes are saved immediately
- [ ] Public profile options work correctly
- [ ] Data export functionality provides complete user data
- [ ] Account deletion process removes all user data
- [ ] GDPR compliance features work as required

### Security Requirements
- [ ] Profile data is properly encrypted
- [ ] Privacy controls are enforced correctly
- [ ] Audit logging captures all profile changes
- [ ] Rate limiting prevents abuse
- [ ] Input validation prevents injection attacks

### UX Requirements
- [ ] Profile interface is intuitive and accessible
- [ ] Form validation provides clear feedback
- [ ] Privacy settings are easy to understand
- [ ] Profile completion is encouraged but not required
- [ ] Mobile responsive design works correctly

## Dependencies

- **AUTH-001**: Authentication system (required for profile access)
- **NOTIF-001**: Notification service (for profile change confirmations)
- **VALID-001**: Input validation service (for profile data validation)

## Risk Assessment

### High Risk
- GDPR compliance violations
- Data privacy breaches
- Profile data loss or corruption

### Medium Risk
- Performance with large profile data
- Privacy setting complexity
- Third-party integration issues

## Implementation Notes

### UX Recommendations
@EXPERT:UI-UX - Use progressive disclosure for complex privacy settings
@EXPERT:UX - Provide clear feedback for profile update status
@EXPERT:ACCESSIBILITY - Ensure WCAG 2.2 AA compliance for all profile interfaces

### Technical Considerations
- Use database encryption for sensitive fields
- Implement caching for frequently accessed profile data
- Consider CDN for avatar and media files
- Plan for data migration from legacy systems
```

---

## Example 4: Background Job Processing (Medium Complexity)

**Domain**: JOB-001
**Estimated Hours**: 10
**Priority**: Medium

```yaml
---
id: JOB-001
version: 0.1.0
status: draft
priority: medium
created: 2025-01-16
author: @backend-team
domain: job
complexity: medium
estimated_hours: 10
dependencies: [QUEUE-001, MONITOR-001]
tags: [background-jobs, queue, processing, monitoring]
reviewers: [@tech-lead, @devops-lead]
milestone: "Sprint 23 - Q1 2025"
---

# @SPEC:EX-JOB-001: Background Job Processing System

## Overview
Provide reliable background job processing system with queue management, job monitoring, retry logic, and failure handling.

## EARS Requirements

### Ubiquitous Requirements
- The system SHALL provide job queue management
- The system SHALL support job retry with exponential backoff
- The system SHALL maintain job execution history
- The system SHALL provide job monitoring and alerting
- The system SHALL handle job failures gracefully

### Event-driven Requirements
- WHEN job is queued, the system SHALL assign priority and schedule execution
- WHEN job fails, the system SHALL retry according to configured policy
- WHEN job succeeds, the system SHALL record completion and trigger follow-up jobs
- WHEN job exceeds maximum retries, the system SHALL move to dead letter queue
- WHEN queue capacity is exceeded, the system SHALL alert administrators

### State-driven Requirements
- WHILE job is processing, the system SHALL update status and progress
- WHILE queue is under high load, the system SHALL scale worker processes
- WHILE dead letter queue has jobs, the system SHALL alert for manual intervention

### Unwanted Behaviors
- The system SHALL NOT lose jobs during processing
- Jobs SHALL NOT execute indefinitely without timeout
- The system SHALL NOT allow duplicate job execution
- Job payload SHALL NOT exceed maximum size limits

## Technical Requirements

### Job Types
- **Email Jobs**: Send notifications, marketing emails
- **Report Jobs**: Generate reports, analytics data
- **Cleanup Jobs**: Data cleanup, archive old records
- **Integration Jobs**: Process webhooks, external API calls

### Queue Configuration
- Redis queue backend for performance
- Priority queues (high, medium, low)
- Maximum job size: 1MB
- Job timeout: 30 minutes default
- Retry policy: 3 retries with exponential backoff

### Monitoring Requirements
- Job execution metrics
- Queue depth monitoring
- Worker process health
- Failure rate alerting
- Performance dashboards

## Acceptance Criteria

- [ ] Jobs can be queued and processed reliably
- [ ] Retry logic handles transient failures
- [ ] Dead letter queue captures failed jobs
- [ ] Monitoring provides visibility into job status
- [ ] Performance requirements are met under load
- [ ] System handles queue overflow gracefully
- [ ] Admin interface provides job management

## Dependencies

- **QUEUE-001**: Redis queue service
- **MONITOR-001**: Monitoring and alerting service

## Risk Assessment

### Medium Risk
- Queue service reliability
- Job processing performance
- Retry logic complexity

### Low Risk
- Worker process management
- Job priority handling
```

---

## SPEC Quality Checklist

For each SPEC, ensure it meets these quality standards:

### Content Quality
- [ ] Requirements are clear and unambiguous
- [ ] All use cases are covered
- [ ] Error conditions are specified
- [ ] Acceptance criteria are testable
- [ ] Dependencies are identified

### Format Compliance
- [ ] YAML frontmatter is complete and valid
- [ ] EARS patterns are used correctly
- [ ] Language is consistent throughout
- [ ] Structure follows standard template
- [ ] TAG references are correct

### Technical Feasibility
- [ ] Requirements are technically achievable
- [ ] Security considerations are addressed
- [ ] Performance requirements are realistic
- [ ] Integration points are defined
- [ ] Risk assessment is comprehensive

### Review Process
- [ ] Technical review completed
- [ ] Domain expert validation
- [ ] Security review (if applicable)
- [ ] Stakeholder approval obtained
- [ ] Version control properly managed

## Creating Your Own SPEC

When writing your own SPEC, follow this process:

### 1. Requirements Gathering
- Talk to stakeholders
- Document all use cases
- Identify constraints and dependencies
- Consider edge cases and error conditions

### 2. Draft the SPEC
- Start with YAML frontmatter
- Write EARS requirements systematically
- Include technical requirements
- Define acceptance criteria

### 3. Review and Refine
- Get feedback from technical team
- Validate with domain experts
- Check for completeness and clarity
- Ensure testability of all requirements

### 4. Finalize and Approve
- Incorporate review feedback
- Obtain stakeholder sign-off
- Version control the SPEC
- Share with implementation team

Remember: A good SPEC is an investment in project success. Take the time to get it right, and the implementation phase will be much smoother! 🎯