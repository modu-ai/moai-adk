---
name: moai-foundation-specs
description: SPEC document management - lifecycle, versioning, approval workflows, 50+ references, SPEC-first TDD integration
---

## Quick Reference (30 seconds)

# SPEC Foundation Skill - Expert  

## Skill Overview

**SPEC** (Specification) is the formal requirements document that drives SPEC-first, TDD development. This Skill provides comprehensive guidance on SPEC lifecycle management, version control, approval workflows, traceability, and integration with MoAI-ADK development pipeline.

### Quick Facts
- **4 SPEC Lifecycle States**: Draft, Active, Deprecated, Archived
- **Version Management**: Semantic versioning (major.minor.patch)
- **Approval Workflow**: Author → Review → Approval → Deployment
- **Integration**: Core of `/alfred:1-plan` workflow in MoAI-ADK

### When to Use This Skill
- Creating formal specifications before development
- Managing specification versions and evolution
- Setting up approval workflows for requirements
- Tracing requirements through code and tests
- Organizing multiple specifications in complex projects

---

## Risks
- User adoption of MFA might be low
- SMS delivery reliability (use backup)
```

**Draft Anti-Patterns - Avoid**:
- ❌ Vague requirements ("system shall be secure")
- ❌ Mixing implementation details with requirements
- ❌ Incomplete acceptance criteria
- ❌ No identified risks or constraints
- ❌ Unsourceable or unmeasurable requirements

---

### 2. Review State - Formal Evaluation

**Purpose**: Peer review and stakeholder feedback

**Review Participants**:
- **Author**: Specification creator
- **Technical Lead**: Architecture and feasibility review
- **QA Lead**: Test coverage and acceptance criteria review
- **Product Owner**: Business requirement alignment
- **Domain Experts**: Subject matter expert review (if applicable)

**Review Checklist**:
```
[ ] Requirements are clear and unambiguous
[ ] All requirements are EARS-format
[ ] Acceptance criteria are measurable
[ ] No conflicting requirements
[ ] Architecture feasible
[ ] Risk assessment complete
[ ] Traceability clear
[ ] No external dependencies missing
[ ] Timeline realistic
[ ] Budget/resources adequate
```

**Review Process**:
1. **Initial Submission**: Author marks SPEC ready for review
2. **Reviewer Comments**: Team adds comments/questions
3. **Author Responses**: Author clarifies or updates spec
4. **Revisions**: 2-3 rounds typical for complex specs
5. **Consensus**: Team agrees specification is complete
6. **Approval Gate**: Technical lead + Product owner sign-off

**Review Duration**: 3-7 business days (parallel review)

**Version Bumping Rules**:
- Each revision during review: increment patch (0.1.0 → 0.1.1)
- Major revisions (scope change): increment minor (0.1.0 → 0.2.0)

---

### 3. Active State - Implementation Period

**Purpose**: Specification is approved and development proceeds

**Activation Steps**:
1. Technical lead approves and signs SPEC
2. Create feature branch: `feature/SPEC-XXX`
3. Implement per SPEC requirements
4. Tests validate against acceptance criteria

**During Active Phase**:
- ✅ Spec is reference for development
- ✅ Any change discussion references spec
- ✅ Code reviews verify against spec
- ✅ Tests trace to spec requirements
- ✅ Track deviations and change requests

**Change Management**:
```
If requirement change needed during development:
1. Assess impact on timeline/scope
2. Document change request
3. Get approval from technical lead + product owner
4. Update SPEC with new version
5. Notify implementation team
6. Update code and tests accordingly
```

**Typical Duration**: Development time + testing (1-8 weeks)

**Version Bumping**:
- Minor feature additions: increment minor (1.0.0 → 1.1.0)
- Bug fixes to spec: increment patch (1.0.0 → 1.0.1)
- Scope changes: increment major (1.0.0 → 2.0.0)

**Completion Criteria**:
- ✅ All requirements implemented
- ✅ All acceptance criteria passed
- ✅ Code review approved
- ✅ Tests passing (≥85% coverage)
- ✅ Documentation updated
- ✅ Ready for deployment

---

### 4. Deprecated State - Phase-Out Period

**Purpose**: Feature is being replaced or removed

**Triggering Events**:
- New feature replaces old functionality
- System architecture change
- Technology upgrade required
- Business decision to sunset feature

**Deprecation Process**:
1. Mark SPEC as DEPRECATED in metadata
2. Create successor SPEC (if applicable)
3. Document migration path for users
4. Set end-of-life date (typically 6-12 months)
5. Notify stakeholders of timeline

**During Deprecation**:
- ✅ Maintain feature (bug fixes)
- ✅ No new feature development
- ✅ Gradual user migration
- ✅ Support successor feature
- ✅ Plan removal

**Deprecation Notice Format**:
```markdown
# SPEC-042: Old Authentication (DEPRECATED)

**Status**: DEPRECATED
**Successor**: SPEC-045 (New MFA Authentication)
**Migration Guide**: See migration-guide.md
**EOL Date**: 2025-12-31 (6 months from deprecation)

## Support
- Questions: Ask in #migration channel
- Migration assistance: migration-team@example.com
```

**Version Marking**:
- Mark SPEC with version tag: `v1.5.0 (DEPRECATED)`
- Update status in all references
- Create deprecation notice in documentation

---

### 5. Archived State - Historical Reference

**Purpose**: SPEC is no longer active, kept for historical record

**Archival Process**:
1. Feature removed from production
2. Mark SPEC as ARCHIVED
3. Move to archive directory: `.moai/specs/archived/`
4. Maintain for audit/compliance purposes
5. Tag with final version and EOL date

**Archive Retention**:
- Keep indefinitely for compliance requirements
- Compress old versions
- Index for historical search

**Archive Access**:
- Readable by all (reference)
- No modifications allowed
- Available for audit trails

---

## Integration Points

- Payment Gateway APIs (Stripe, PayPal)
- Accounting system (QuickBooks API)
- Fraud detection (Third-party service)
- Notification system (Email, SMS, in-app)

## Organization Patterns

### Small Project (1-3 specs)
```
.moai/specs/
├── SPEC-001/
│   ├── spec.md
│   └── acceptance-criteria.md
├── SPEC-002/
└── SPEC-003/
```

### Medium Project (5-20 specs)
```
.moai/specs/
├── core/
│   ├── SPEC-001/ (auth)
│   └── SPEC-002/ (api)
├── features/
│   ├── SPEC-010/ (profile)
│   └── SPEC-011/ (payments)
├── infrastructure/
│   ├── SPEC-020/ (database)
│   └── SPEC-021/ (monitoring)
└── deprecated/
    └── SPEC-000/ (old feature)
```

### Large Project (50+ specs)
```
.moai/specs/
├── index.md (SPEC registry)
├── platform/
│   ├── auth/ (4 specs)
│   ├── api/ (3 specs)
│   ├── user/ (5 specs)
│   └── payments/ (3 specs)
├── features/
│   ├── analytics/ (3 specs)
│   ├── reporting/ (2 specs)
│   └── mobile/ (4 specs)
├── infrastructure/
│   ├── backend/ (5 specs)
│   ├── devops/ (4 specs)
│   └── security/ (3 specs)
├── deprecated/ (archived specs)
└── archive/ (historical reference)
```

---

---

## Core Implementation

## Level 1: Foundation - SPEC Lifecycle

### 1. Draft State - Specification Creation

**Purpose**: Initial specification authoring and refinement

**Activities**:
```
1. Specification Author creates SPEC-XXX/spec.md
2. Define requirements using EARS patterns
3. Gather stakeholder input
4. Refine until ready for review
5. Create acceptance criteria
6. Document known risks and constraints
```

**Typical Duration**: 2-5 days (simple features) to 2-4 weeks (complex systems)

**Key Artifacts**:
- `spec.md` - Main specification document
- `acceptance-criteria.md` - Acceptance tests (if separate)
- `technical-notes.md` - Implementation guidance (optional)

**Deliverables for Review**:
- ✅ Clear problem statement
- ✅ EARS-format requirements (functional & non-functional)
- ✅ Acceptance criteria for all requirements
- ✅ Architecture/design notes
- ✅ Risk assessment
- ✅ Dependencies identified

**Example Draft Structure**:
```markdown
# SPEC-045: User Authentication System

## Problem Statement
Current system lacks multi-factor authentication. Need MFA for security compliance.

## Requirements
REQ-001 (Event-Driven): When login_attempted the system eventually satisfies 
        mfa_challenge_presented
REQ-002 (Ubiquitous): The system shall always satisfy mfa_enabled_for_admin = true
REQ-003 (Optional): When mfa_timeout_exceeded the system immediately satisfies 
        session_terminated

## Acceptance Criteria
- [ ] MFA works with authenticator apps (Google, Microsoft)
- [ ] Fallback SMS when app unavailable
- [ ] Session timeout after 10 minutes inactivity
- [ ] Audit log all MFA events

## Technical Notes
- Use TOTP (RFC 6238) for time-based codes
- Backup codes for emergency access
- Consider integration with existing identity system

## Migration Timeline
- 2025-07: New system available, parallel operation
- 2025-09: Default switch to new system
- 2025-12: Old system shutdown

## v1.2.3 (2025-11-12) - Clarifications
- Clarified requirement REQ-002 acceptance criteria
- Added risk assessment for database migration
- Updated timeline from 4 weeks to 6 weeks
- Author: john-smith | Tech Lead: sarah-jones

## v1.2.2 (2025-11-10) - Bug Fix
- Fixed requirement numbering consistency
- Author: john-smith

## v1.2.1 (2025-11-08) - Review Updates
- Addressed QA concerns about test coverage
- Added backup procedures to recovery plan
- Author: john-smith | Reviewer: qa-lead
```

---

## Level 3: Practical Application

### Complete SPEC Examples

#### Example 1: Simple Feature SPEC (SPEC-050)

```markdown
---
name: User Profile Enhancement
spec_id: SPEC-050
version: 1.0.0
status: stable
created: 2025-11-01
approved_date: 2025-11-08
approved_by: tech-lead
---

# SPEC-050: User Profile Enhancement

## Problem Statement
Users cannot upload profile pictures. Current profile view shows placeholder only.

## Functional Requirements

REQ-001 (Event-Driven):
  When user_uploads_profile_image the system eventually satisfies 
  image_stored_in_profile_and_cache_updated
  Acceptance: Image appears immediately after upload
  
REQ-002 (Ubiquitous):
  The system shall always satisfy profile_image_size <= 5MB
  Acceptance: Upload fails with error if exceeds 5MB

REQ-003 (Unwanted):
  The system shall never satisfy (invalid_image_format AND stored)
  Acceptance: Only PNG, JPEG, WebP accepted

## Non-Functional Requirements

- Performance: Image upload completes within 3 seconds
- Security: Images scanned for malware
- Compliance: GDPR-compliant data storage

## Acceptance Criteria
- [ ] Upload works for PNG, JPEG, WebP
- [ ] File size limited to 5MB
- [ ] Image appears in profile immediately
- [ ] Old images automatically deleted
- [ ] Performance < 3 seconds on 4G
- [ ] Mobile and desktop tested

## Testing Strategy
- Unit tests: Image validation, storage
- Integration tests: Upload workflow
- E2E tests: User upload → profile view
- Manual: Test on various devices

## Technical Notes
- Use S3 for image storage
- CloudFront CDN for distribution
- ImgProxy for optimization

## Risks
- Malware in images (mitigate with scanning)
- Storage costs (monitor usage)
- CDN cache invalidation (use versioning)

## Dependencies
- S3 bucket provisioning
- ImgProxy service deployment
- Malware scanning service

## Timeline
- Development: 2 weeks
- Testing: 1 week
- Deployment: 1 day
```

#### Example 2: Complex System SPEC (SPEC-051)

```markdown
---
name: Payment Processing Refactor
spec_id: SPEC-051
version: 2.1.0
status: stable
created: 2025-10-15
approved_date: 2025-11-01
approved_by: tech-lead, product-owner
---

# SPEC-051: Payment Processing Refactor

## Problem Statement
Current payment system doesn't support multiple payment providers. Need flexibility 
to add Stripe, PayPal, Square without major refactoring.

## Architecture

### System Components
```
Payment Service (core abstraction)
├── Stripe Provider (implementation)
├── PayPal Provider (implementation)
└── Square Provider (future)
```

## Functional Requirements

REQ-001-005: [5 event-driven payment flow requirements]
REQ-006-010: [5 ubiquitous invariants for payment safety]
REQ-011-015: [5 state-driven mode requirements]

## Non-Functional Requirements

- Throughput: ≥ 1000 transactions/sec
- Latency: P99 < 500ms
- Availability: 99.95% uptime
- Security: PCI-DSS Level 1 compliance
- Scalability: Auto-scale to 10k transactions/sec

## Acceptance Criteria

- [ ] All payment methods work end-to-end
- [ ] Transactions persist through failures
- [ ] Receipts generated automatically
- [ ] Refunds processed within 2 hours
- [ ] All error cases handled gracefully
- [ ] Performance targets met
- [ ] Security audit passed

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| API rate limits | Medium | High | Implement queue, caching |
| Data loss | Low | Critical | Transaction journaling |
| Fraud | Medium | High | Third-party fraud detection |
| Compliance violation | Low | Critical | Regular audits |

## Timeline & Resources

- Backend Development: 4 weeks (2 engineers)
- Frontend Integration: 2 weeks (1 engineer)
- QA Testing: 2 weeks (2 QA engineers)
- Deployment & monitoring: 1 week (devops)
- Total: 9 weeks, 6 people

## Success Metrics

- Zero payment failures
- Payment latency < 500ms (P99)
- User-reported issues < 0.1%
- All tests passing (≥90% coverage)
```

---

## Best Practices

### 1. Specification Clarity
- ✅ Use EARS patterns for all requirements
- ✅ Define acceptance criteria before development
- ✅ Include rationale for non-obvious requirements
- ✅ Document constraints and assumptions
- ✅ Keep specifications concise but complete

### 2. Approval Process
- ✅ Define clear reviewers (technical, product, domain)
- ✅ Use structured review checklist
- ✅ Set review timeline (3-7 days)
- ✅ Document approval decision
- ✅ Require sign-off from decision makers

### 3. Version Management
- ✅ Use semantic versioning consistently
- ✅ Document every change with rationale
- ✅ Keep complete version history
- ✅ Mark breaking changes clearly
- ✅ Create migration guides for major versions

### 4. Traceability
- ✅ Link tests to requirements
- ✅ Link documentation to spec
- ✅ Create traceability matrix
- ✅ Verify no orphaned requirements

### 5. Organization
- ✅ Use consistent directory structure: `.moai/specs/SPEC-XXX/`
- ✅ Keep related specs together
- ✅ Link dependent specs
- ✅ Archive deprecated specs
- ✅ Index active specs

---

## SPEC Integration with MoAI-ADK

### With `/alfred:1-plan` Command
```bash
/alfred:1-plan "user profile enhancement feature"
  ↓
Creates SPEC-XXX structure
  ├── spec.md (specification)
  ├── acceptance-criteria.md
  ├── technical-notes.md (optional)
  └── CHANGELOG.md
  ↓
Author reviews and marks ready
  ↓
Tech lead approves
  ↓
Status: ACTIVE
```

### With `/alfred:2-run` Command
```bash
/alfred:2-run SPEC-050
  ↓
Reads SPEC-050 specification
  ↓
TDD cycle:
  RED: Tests from acceptance criteria
  GREEN: Implementation
  REFACTOR: Code quality
  ↓
  ↓
Tests link to requirements
```

### With `/alfred:3-sync` Command
```bash
/alfred:3-sync auto SPEC-050
  ↓
Validates all acceptance criteria met
  ↓
Updates documentation
  ↓
Verifies test coverage
  ↓
Creates PR to develop
```

### With moai-foundation-tags
- Documentation includes spec rationale
- Complete traceability: SPEC → Code → Tests → Docs

---



---

## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.
