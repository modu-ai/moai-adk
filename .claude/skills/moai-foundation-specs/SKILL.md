---
name: moai-foundation-specs
description: SPEC document management - lifecycle, versioning, approval workflows, SPEC-first TDD integration
---

## Quick Reference

SPEC (Specification) is the formal requirements document that drives SPEC-first, TDD development.

**Quick Facts**:
- **4 SPEC Lifecycle States**: Draft, Active, Deprecated, Archived
- **Version Management**: Semantic versioning (major.minor.patch)
- **Approval Workflow**: Author → Review → Approval → Deployment
- **Integration**: Core of `/alfred:1-plan` workflow in MoAI-ADK

**When to Use**:
- Creating formal specifications before development
- Managing specification versions and evolution
- Setting up approval workflows for requirements
- Tracing requirements through code and tests
- Organizing multiple specifications in complex projects

---

## Implementation Guide

### Draft State - Specification Creation

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

## Risks
- User adoption of MFA might be low
- SMS delivery reliability (use backup)
```

### Review State - Formal Evaluation

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
[ ] Timeline realistic
```

**Review Duration**: 3-7 business days (parallel review)

### Active State - Implementation Period

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

**Completion Criteria**:
- ✅ All requirements implemented
- ✅ All acceptance criteria passed
- ✅ Code review approved
- ✅ Tests passing (≥85% coverage)
- ✅ Documentation updated
- ✅ Ready for deployment

### Deprecated State - Phase-Out Period

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

---

## Advanced Patterns

### Organization Patterns

**Small Project (1-3 specs)**:
```
.moai/specs/
├── SPEC-001/
│   ├── spec.md
│   └── acceptance-criteria.md
├── SPEC-002/
└── SPEC-003/
```

**Medium Project (5-20 specs)**:
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

**Large Project (50+ specs)**:
```
.moai/specs/
├── index.md (SPEC registry)
├── platform/
│   ├── auth/ (4 specs)
│   ├── api/ (3 specs)
│   └── payments/ (3 specs)
├── features/
│   ├── analytics/ (3 specs)
│   └── mobile/ (4 specs)
├── infrastructure/
│   ├── backend/ (5 specs)
│   └── devops/ (4 specs)
└── deprecated/ (archived specs)
```

### SPEC Integration with MoAI-ADK

**With `/alfred:1-plan` Command**:
```bash
/alfred:1-plan "user profile enhancement feature"
  ↓
Creates SPEC-XXX structure
  ├── spec.md (specification)
  ├── acceptance-criteria.md
  └── CHANGELOG.md
  ↓
Status: ACTIVE
```

**With `/alfred:2-run` Command**:
```bash
/alfred:2-run SPEC-050
  ↓
TDD cycle:
  RED: Tests from acceptance criteria
  GREEN: Implementation
  REFACTOR: Code quality
  ↓
Tests link to requirements
```

**With `/alfred:3-sync` Command**:
```bash
/alfred:3-sync auto SPEC-050
  ↓
Validates all acceptance criteria met
  ↓
Updates documentation
  ↓
Creates PR to develop
```

**Best Practices**:
- ✅ Use EARS patterns for all requirements
- ✅ Define acceptance criteria before development
- ✅ Include rationale for non-obvious requirements
- ✅ Use semantic versioning consistently
- ✅ Link tests to requirements
