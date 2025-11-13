# moai-foundation-specs - Reference Guide

_Last updated: 2025-11-12 | Version: 4.0.0_

## Quick Reference - SPEC Lifecycle States

| State | Purpose | Duration | Key Activities |
|-------|---------|----------|-----------------|
| **Draft** | Creation & refinement | 2-5 days to 2-4 weeks | Write spec, gather input, refine |
| **Review** | Peer evaluation | 3-7 business days | Review, feedback, revisions |
| **Active** | Implementation | 1-8 weeks | Develop, test, deploy |
| **Deprecated** | Phase-out period | 6-12 months | Maintain, migrate users |
| **Archived** | Historical record | Indefinite | Read-only, audit trail |

## SPEC Directory Structure

### Minimal (1-5 specs)
```
.moai/specs/
├── SPEC-001/
│   ├── spec.md
│   └── acceptance-criteria.md
└── SPEC-002/
```

### Organized (5-20 specs)
```
.moai/specs/
├── core/
│   ├── SPEC-001-auth/
│   └── SPEC-002-api/
├── features/
│   ├── SPEC-010-profile/
│   └── SPEC-011-payments/
└── deprecated/
    └── SPEC-000-old-feature/
```

### Structured (50+ specs)
```
.moai/specs/
├── index.md (SPEC registry)
├── platform/ (system specs)
├── features/ (feature specs)
├── infrastructure/ (infra specs)
├── deprecated/ (phase-out)
└── archive/ (historical)
```

## Semantic Versioning Quick Reference

```
Format: major.minor.patch (e.g., 1.2.3)

PATCH Version (1.0.X):
  - Bug fixes
  - Minor clarifications
  - No scope change
  Example: 1.0.0 → 1.0.1

MINOR Version (1.X.0):
  - New acceptance criteria
  - Feature refinements
  - Non-breaking additions
  Example: 1.0.0 → 1.1.0

MAJOR Version (X.0.0):
  - Scope changes
  - Architecture redesign
  - Incompatible changes
  Example: 1.0.0 → 2.0.0

Pre-release:
  - 0.1.0 (draft versions)
  - 1.0.0-rc1 (release candidate)
  - 1.0.0-beta (beta)
```

## SPEC Metadata Template

```yaml
---
name: Feature Name
spec_id: SPEC-XXX
version: 1.0.0
status: active  # draft, review, active, deprecated, archived
created: 2025-11-01
updated: 2025-11-12
approved_by: tech-lead-name
approval_date: 2025-11-08
deprecated: false
eol_date: null  # Set when deprecated
---
```

## SPEC Frontmatter Checklist

```
[ ] Name: Clear, descriptive title
[ ] spec_id: SPEC-XXX format
[ ] version: Semantic version (major.minor.patch)
[ ] status: One of 5 lifecycle states
[ ] created: ISO 8601 date
[ ] updated: ISO 8601 date
[ ] approved_by: Tech lead name (if approved)
[ ] approval_date: Date approved (if approved)
[ ] deprecated: true/false
[ ] eol_date: End-of-life date (if deprecated)
```

## Review Checklist

- [ ] Requirements clear and unambiguous
- [ ] Requirements use EARS patterns
- [ ] Acceptance criteria measurable
- [ ] No conflicting requirements
- [ ] Architecture feasible
- [ ] Risk assessment complete
- [ ] Timeline realistic
- [ ] Resources adequate
- [ ] Dependencies identified
- [ ] No external blockers

## Change Log Entry Format

```markdown
## v1.2.3 (2025-11-12) - Brief Description
- Change 1 description
- Change 2 description
- Author: name | Reviewer: name (if applicable)
```

## Traceability Workflow

```
SPEC-050 (Specification)
  ↓
  Defines requirements (REQ-001, REQ-002, ...)
  ↓
  ↓
  Tests verify requirements
    - test_001_implements_req_001
    - test_002_implements_req_002
  ↓
  Documentation references SPEC-050
  ↓
  Complete traceability: SPEC → Code → Tests → Docs
```

## MoAI-ADK Integration

### Command: `/alfred:1-plan`
```bash
/alfred:1-plan "feature description"
  Creates: .moai/specs/SPEC-XXX/spec.md
  Status: draft
  Next: Author edits, marks ready
```

### Command: `/alfred:2-run`
```bash
/alfred:2-run SPEC-XXX
  Reads: SPEC-XXX/spec.md
  Creates: feature/SPEC-XXX branch
  TDD cycle using acceptance criteria
```

### Command: `/alfred:3-sync`
```bash
/alfred:3-sync auto SPEC-XXX
  Validates: Acceptance criteria met
  Updates: Documentation
  Creates: PR to develop
```

## Key Metrics

| Metric | Target | Notes |
|--------|--------|-------|
| Review time | 3-7 days | Parallel review speeds this up |
| Acceptance criteria | ≥ 5 per feature | Prevent incomplete specs |
| Code coverage | ≥ 85% | Link tests to requirements |
| Test pass rate | 100% | All tests passing before merge |
| Spec version < major | ≥ 2 minor versions | Only 1-2 major versions typical |

## Anti-Patterns to Avoid

| Anti-Pattern | Issue | Fix |
|--------------|-------|-----|
| Vague requirements | "system shall be fast" | Add threshold: "≤ 100ms" |
| No acceptance criteria | Can't verify completion | Add 5+ measurable criteria |
| Mixed patterns | Confusing requirements | Use EARS patterns clearly |
| No version history | Can't track changes | Log every update with rationale |
| Orphaned specs | Outdated, unused | Archive or deprecate clearly |

## Common SPEC Mistakes & Fixes

| Mistake | Example | Fix |
|---------|---------|-----|
| Too vague | "User can upload files" | "User can upload PNG/JPEG/WebP ≤ 5MB" |
| Unmeasurable | "System is reliable" | "Uptime ≥ 99.9%" |
| Missing rationale | Just a list of requirements | Explain why each requirement exists |
| No timeline | Doesn't say how long | "Development 2 weeks, testing 1 week" |
| No risk assessment | Surprised by issues | Identify risks, mitigations upfront |

## Official References

| Category | Link |
|----------|------|
| **IEEE 830** | https://standards.ieee.org/standard/830-1998.html |
| **ISO/IEC/IEEE 29148** | https://standards.ieee.org/standard/29148-2018.html |
| **Semantic Versioning** | https://semver.org/ |
| **Git Flow** | https://nvie.com/posts/a-successful-git-branching-model/ |
| **Conventional Commits** | https://www.conventionalcommits.org/ |
| **ReqIF Standard** | https://www.omg.org/spec/ReqIF/ |

---

**Use this reference for quick lookups on SPEC structure, versioning, and lifecycle management.**
