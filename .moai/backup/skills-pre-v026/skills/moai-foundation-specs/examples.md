# moai-foundation-specs - Examples

_Last updated: 2025-11-12 | Version: 4.0.0_

## Complete SPEC Examples

### Example 1: Simple Feature SPEC (Draft → Active)

```markdown
---
name: User Profile Picture Upload
spec_id: SPEC-050
version: 1.0.0
status: active
created: 2025-11-01
approved_date: 2025-11-08
approved_by: sarah-jones
---

# SPEC-050: User Profile Picture Upload

## Problem Statement
Users cannot customize their profile with pictures. Currently shows placeholder avatar only.
This limits user personalization and account identity.

## Functional Requirements

### REQ-001: Image Upload
When user_uploads_image the system eventually satisfies image_persisted_and_displayed

### REQ-002: File Size Validation
The system shall always satisfy uploaded_image_size <= 5MB

### REQ-003: Format Validation
The system shall never satisfy (invalid_format AND image_stored)

### REQ-004: Thumbnail Generation
When image_uploaded the system eventually satisfies thumbnail_generated

### REQ-005: Cache Invalidation
When profile_picture_changed the system eventually satisfies cached_image_invalidated

## Non-Functional Requirements

- Performance: Upload < 3 seconds on 4G network
- Security: Scan images for malware
- Scalability: Handle 10k concurrent uploads
- Availability: Image always accessible when profile viewed

## Acceptance Criteria

### Upload Functionality
- [x] Accept PNG, JPEG, WebP formats
- [x] Reject files > 5MB with clear error
- [x] Image appears in profile immediately after upload
- [x] Works on mobile and desktop browsers

### Image Handling
- [x] Generate 3 thumbnails (50x50, 150x150, 300x300)
- [x] Delete previous image when new one uploaded
- [x] Handle corrupted files gracefully
- [x] Support drag-and-drop upload

### Performance & Security
- [x] Upload completes within 3 seconds on 4G
- [x] Scan images with ClamAV for malware
- [x] Store images in S3 with encryption
- [x] CDN caching with 1-hour TTL

### Error Handling
- [x] Clear error message if file too large
- [x] Clear error if unsupported format
- [x] Graceful failure if upload interrupted
- [x] User can retry failed uploads

## Technical Design

### Architecture
```
User Browser
    ↓ Upload
Web Server (validation)
    ↓
S3 (store original + thumbnails)
    ↓
CloudFront CDN (serve cached)
    ↓
User Profile View
```

### Technology Stack
- Frontend: React with Dropzone.js
- Backend: Python/FastAPI
- Storage: AWS S3
- Processing: ImageMagick for thumbnails
- Scanning: ClamAV
- CDN: CloudFront

### Implementation Notes
- Use multipart upload for files > 100MB
- Generate thumbnails asynchronously
- Implement exponential backoff for retries
- Cache profile data in Redis

## Dependencies

### External Services
- AWS S3 (storage)
- AWS CloudFront (CDN)
- ClamAV (malware scanning)

### Internal Dependencies
- User authentication system
- Profile service
- Image processing service

### Infrastructure Requirements
- S3 bucket creation
- CloudFront distribution setup
- ClamAV daemon deployment

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Malware upload | Medium | High | ClamAV scanning |
| S3 outage | Low | High | Backup S3 bucket region |
| CDN cache issues | Low | Medium | Cache versioning |
| Large file crashes | Medium | High | Size limit + multipart upload |

## Testing Strategy

### Unit Tests
- Image format validation
- File size validation
- Thumbnail generation

### Integration Tests
- Upload workflow end-to-end
- S3 storage and retrieval
- CDN cache invalidation

### E2E Tests
- User uploads image
- Image appears in profile
- Image appears on other user's view
- Thumbnail generation completes

### Load Tests
- 100 concurrent uploads
- 1000 req/sec read throughput

### Security Tests
- Upload malware test file
- Bypass size limits
- Unauthorized access to images

## Timeline & Resources

| Phase | Duration | Resources |
|-------|----------|-----------|
| Development | 2 weeks | 2 backend engineers |
| Frontend | 1 week | 1 frontend engineer |
| Testing | 1 week | 2 QA engineers |
| Deployment | 1 day | 1 devops engineer |
| **Total** | **5 weeks** | **6 people** |

## Success Metrics

- [ ] Zero image upload failures
- [ ] Upload completes < 3 seconds (P99)
- [ ] 100% test coverage
- [ ] Zero malware uploads detected
- [ ] User satisfaction > 4.5/5 stars
- [ ] No CDN cache issues

## Version History

### v1.0.0 (2025-11-08)
- Initial specification approved
- Author: john-smith | Approver: sarah-jones

### v0.3.0 (2025-11-06)
- Added security scanning requirement
- Changed max file size to 5MB
- Added multipart upload note

### v0.2.0 (2025-11-04)
- QA feedback: Added error handling AC
- Added performance testing criteria

### v0.1.0 (2025-11-01)
- Initial draft
```

---

### Example 2: Complex System SPEC (Active → Deprecated)

```markdown
---
name: Payment Processing Refactor
spec_id: SPEC-051
version: 2.1.0
status: active
created: 2025-10-15
approved_date: 2025-11-01
approved_by: tech-lead, product-owner
deprecated: false
---

# SPEC-051: Payment Processing Refactor

## Executive Summary
Migrate from monolithic payment system to provider-agnostic architecture supporting
Stripe, PayPal, and Square. Enable rapid provider additions without code changes.

## Problem Statement
Current payment system is tightly coupled to Stripe. Adding new providers requires
extensive refactoring. Need flexible, pluggable architecture.

## Current State
- Only Stripe supported
- Tight coupling to Stripe API
- 10,000+ lines in single payment module
- Difficult to test
- No abstraction layer

## Desired State
- Support Stripe, PayPal, Square
- Provider-agnostic abstraction
- Easy to add new providers
- Comprehensive test coverage
- Clear provider contracts

## System Architecture

### Component Structure
```
PaymentService (abstract interface)
├── StripeProvider (implements PaymentService)
├── PayPalProvider (implements PaymentService)
├── SquareProvider (implements PaymentService)
└── MockProvider (for testing)

PaymentProcessor
├── Validate transaction
├── Route to provider
├── Handle response
└── Log transaction
```

### Integration Points
```
Frontend (Payment UI)
    ↓
Payment Service API
    ↓
PaymentProcessor
    ├→ Provider Router
    ├→ Transaction Logger
    ├→ Webhook Handler
    └→ Notification Service
    ↓
External APIs (Stripe, PayPal, Square)
```

## Functional Requirements

### REQ-001-010: Provider Support
- Process transactions with Stripe
- Process transactions with PayPal
- Process transactions with Square
- Detect provider status
- Fallback to alternative provider

### REQ-011-020: Transaction Handling
- Create new transaction
- Capture authorized transaction
- Refund transaction
- Cancel pending transaction
- Handle transaction failures

### REQ-021-030: Webhook Handling
- Receive provider webhooks
- Verify webhook authenticity
- Process webhook events
- Retry failed webhook processing
- Log all webhook events

## Non-Functional Requirements

### Performance
- Throughput: ≥ 1000 transactions/sec
- Latency: P99 < 500ms
- Availability: 99.95% uptime
- Provider timeout: 30 seconds

### Scalability
- Handle peak load: 10,000 transactions/sec
- Auto-scale providers horizontally
- Queue-based processing for high volume

### Security
- PCI-DSS Level 1 compliance
- Encrypt sensitive data (PAN, CVV)
- No plaintext credentials in logs
- TLS 1.2+ for all connections
- Rotate provider API keys monthly

### Reliability
- Automatic retry on transient failures
- Circuit breaker for failing providers
- Comprehensive transaction journaling
- Zero transaction loss

## Implementation Phases

### Phase 1: Foundation (Weeks 1-4)
- Define PaymentService interface
- Create StripeProvider implementation
- Implement transaction logging
- Set up testing framework

### Phase 2: Provider Support (Weeks 5-8)
- Implement PayPalProvider
- Implement SquareProvider
- Create provider router
- Implement provider failover

### Phase 3: Advanced Features (Weeks 9-12)
- Webhook handling
- Reconciliation system
- Analytics & reporting
- Performance optimization

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Provider API changes | High | Medium | Version API contracts, versioning |
| Provider outage | Low | High | Multi-provider support, failover |
| Data loss | Low | Critical | Transaction journaling, backup |
| Compliance violation | Low | Critical | Regular security audits |
| Performance degradation | Medium | High | Load testing, optimization |

## Success Metrics

- [ ] All tests passing (≥ 90% coverage)
- [ ] Payment success rate > 99.5%
- [ ] Transaction latency P99 < 500ms
- [ ] Zero PCI-DSS violations
- [ ] Zero data loss incidents
- [ ] New provider addition < 2 weeks

## Migration Path

### Phase 1: Parallel Operation (Week 1)
- New system operational alongside old
- 10% traffic to new system
- Monitor for issues

### Phase 2: Gradual Rollout (Weeks 2-3)
- Increase to 50% traffic
- Test failover scenarios
- Verify accuracy

### Phase 3: Complete Cutover (Week 4)
- 100% traffic to new system
- Maintain old system for 1 week (rollback)
- Final monitoring & verification

## Version History

### v2.1.0 (2025-11-01)
- Updated security requirements (PCI-DSS Level 1)
- Added Circuit Breaker pattern
- Extended timeout to 30 seconds
- Author: john | Approver: sarah

### v2.0.0 (2025-10-28)
- Added PayPal and Square support
- Major architecture revision

### v1.2.0 (2025-10-22)
- Webhook handling requirements

### v1.0.0 (2025-10-15)
- Initial specification
```

---

### Example 3: Deprecated SPEC (Transition Plan)

```markdown
---
name: Old Authentication System (DEPRECATED)
spec_id: SPEC-042
version: 1.5.0
status: deprecated
created: 2025-05-15
deprecated: true
eol_date: 2025-12-31
---

# SPEC-042: Old Authentication System (DEPRECATED)

**Status**: DEPRECATED
**Successor**: SPEC-045 (New MFA Authentication)
**End-of-Life**: 2025-12-31 (6 months from deprecation date 2025-06-30)

## Deprecation Notice

This specification is being phased out in favor of SPEC-045 which provides
modern MFA support and improved security.

## Migration Timeline

```
Now (2025-07-01): New system available, parallel operation
September 2025: Default switch to new system
December 2025: Old system shutdown
```

## Migration Path for Users

### Step 1: Enable New Authentication (Immediate)
1. Visit Account Settings → Security
2. Click "Enable New Authentication"
3. Enroll in MFA (Google Authenticator or Authy)
4. Keep account in both systems until September

### Step 2: Switch Default (August 2025)
1. Systems will auto-switch default in August
2. You can switch back if issues
3. Contact support if problems: migration-team@example.com

### Step 3: Phase-Out (December 2025)
1. Old system disabled December 31, 2025
2. Must be on new system by then
3. Contact support for urgent migration help

## Support & Resources

- Migration Guide: https://example.com/migration-guide
- Video Tutorial: https://example.com/video
- FAQ: https://example.com/faq
- Support: migration-team@example.com
- Slack Channel: #auth-migration

## Known Issues with New System

### Minor Issues
- Recovery codes can't be regenerated (planned v2.1)
- SMS delivery occasionally delayed (uses backup)

### Workarounds
- Save recovery codes immediately after setup
- Enable both Google Auth and SMS if available

## Statistics

- Users migrated: 45% (15,000 of 33,000)
- Planned completion: 95% by December 31
- Support tickets: 127 (mostly questions, resolved quickly)

## Changes in New System

### Benefits
✅ Multi-factor authentication (MFA)
✅ Biometric support (fingerprint, face)
✅ Passwordless sign-in option
✅ Better recovery mechanisms
✅ Improved security audit trail

### Feature Parity
- Email verification: Still supported
- Login history: Enhanced logging
- Session management: Improved
- API tokens: Better management

### Notable Differences
- SMS codes now 6-digit (was 4-digit)
- Authenticator now required for admin
- Recovery codes changed (new system)
```

---

## SPEC Organization Examples

### Startup Project (3-5 Specs)

```
.moai/specs/
├── README.md
├── SPEC-001-user-auth/
│   ├── spec.md
│   ├── acceptance-criteria.md
│   └── CHANGELOG.md
├── SPEC-002-product-catalog/
│   ├── spec.md
│   └── acceptance-criteria.md
└── SPEC-003-checkout/
    ├── spec.md
    └── acceptance-criteria.md
```

### Growing Product (15-30 Specs)

```
.moai/specs/
├── index.md (SPEC registry)
├── core/
│   ├── SPEC-001-auth/
│   ├── SPEC-002-api/
│   └── SPEC-003-database/
├── features/
│   ├── SPEC-010-user-profile/
│   ├── SPEC-011-search/
│   └── SPEC-012-analytics/
├── payments/
│   ├── SPEC-020-payment-processing/
│   ├── SPEC-021-invoicing/
│   └── SPEC-022-refunds/
├── deprecated/
│   └── SPEC-000-old-auth/
└── CHANGELOG-all.md
```

### Enterprise System (50+ Specs)

```
.moai/specs/
├── index.md (SPEC registry with search)
├── platform/
│   ├── auth/ (4 specs)
│   ├── api/ (3 specs)
│   ├── user-mgmt/ (3 specs)
│   └── notifications/ (2 specs)
├── features/
│   ├── payments/ (5 specs)
│   ├── analytics/ (4 specs)
│   ├── reporting/ (3 specs)
│   └── integrations/ (4 specs)
├── infrastructure/
│   ├── backend/ (5 specs)
│   ├── frontend/ (3 specs)
│   ├── devops/ (4 specs)
│   ├── security/ (3 specs)
│   └── monitoring/ (2 specs)
├── deprecated/ (5 old specs)
├── archive/ (historical specs)
├── CHANGELOG-all.md
└── SPEC-registry.md (with links)
```

---

## SPEC Change Examples

### Minor Change (Patch Version)

```
Title: Fix typo in requirement
Old version: 1.0.0
New version: 1.0.1

Change Log Entry:
## v1.0.1 (2025-11-12) - Typo Fix
- Fixed typo in REQ-003: "persistant" → "persistent"
- Author: jane-smith
```

### Feature Addition (Minor Version)

```
Title: Add support for batch uploads
Old version: 1.0.0
New version: 1.1.0

Changes:
- New requirement: Batch upload up to 100 files
- New acceptance criteria: 5 items
- No breaking changes

Change Log Entry:
## v1.1.0 (2025-11-12) - Batch Upload Support
- Added batch upload requirement (REQ-004)
- Added 5 new acceptance criteria for batch operations
- Timeline: +1 week for implementation
- Author: john-smith | Approver: sarah-jones
```

### Major Change (Major Version)

```
Title: Switch from image format to video support
Old version: 1.0.0
New version: 2.0.0

Changes:
- Remove image upload requirements (REQ-001-003 deprecated)
- Add video upload requirements (REQ-101-103)
- Different format support (MP4, WebM instead of PNG, JPEG)
- Different performance targets

Change Log Entry:
## v2.0.0 (2025-11-12) - Video Support
- BREAKING: Changed from image to video format support
- Removed image-specific requirements
- Added video encoding requirements
- Updated performance targets for video
- Timeline: +4 weeks for implementation
- Migration: Users can still upload images until SPEC-001 deprecated
- Author: john-smith | Approver: sarah-jones
```

---

## Quick SPEC Template

```markdown
---
name: Feature Name
spec_id: SPEC-XXX
version: 1.0.0
status: draft
created: 2025-11-12
approved_by: 
approval_date: 
---

# SPEC-XXX: Feature Name

## Problem Statement
[What problem does this solve?]

## Functional Requirements
[List requirements using EARS patterns]

## Non-Functional Requirements
[Performance, security, scalability, etc.]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
...

## Technical Design
[Architecture, technology choices]

## Timeline
[Development time, resources]

## Success Metrics
[How to measure success]
```

---

**Use these examples as templates for your own SPECs. Keep them clear, specific, and actionable.**
