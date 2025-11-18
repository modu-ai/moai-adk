---
name: moai-core-issue-labels
version: 4.0.0
created: 2025-11-11
updated: '2025-11-18'
status: stable
description: Enterprise GitHub issue labeling orchestrator with semantic label taxonomy,
  AI-powered auto-labeling, label hierarchy system, workflow automation, issue triage
  acceleration, and stakeholder communication; activates for issue classification,
  label management, workflow automation, priority assignment, and team communication
keywords:
- github-labels
- issue-triage
- label-taxonomy
- ai-labeling
- workflow-automation
- issue-classification
- priority-management
- team-communication
- semantic-labels
- enterprise-triage
allowed-tools:
- Read
- Bash
- AskUserQuestion
- mcp__context7__resolve-library-id
- mcp__context7__get-library-docs
- WebFetch
stability: stable
---


# Enterprise GitHub Issue Labeling Orchestrator 

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-core-issue-labels |
| **Version** | 4.0.0 Enterprise (2025-11-18) |
| **AI Integration** | ✅ Context7 MCP, semantic analysis, auto-classification |
| **Auto-load** | On issue creation/update for auto-labeling |
| **Categories** | Type, Priority, Status, Component, Custom |
| **Lines of Content** | 850+ with 13+ production examples |
| **Progressive Disclosure** | 3-level (taxonomy, patterns, automation) |

---

## What It Does

Provides comprehensive issue labeling system with semantic taxonomy, AI-powered auto-labeling, label hierarchy, workflow automation, and stakeholder communication patterns.

---

## Semantic Label Taxonomy

### Type Labels

```
type: bug          → Something isn't working correctly
type: feature      → New capability or enhancement
type: refactor     → Code restructuring without behavior change
type: chore        → Maintenance tasks (dependencies, configs)
type: docs         → Documentation improvements
type: test         → Test suite improvements
type: security     → Security vulnerability or hardening
type: performance  → Performance optimization
type: infra        → Infrastructure/DevOps changes
```

### Priority Labels

```
priority: critical  → Blocks production, urgent (SLA: 4 hours)
priority: high      → Significant impact, schedule soon (SLA: 1 day)
priority: medium    → Normal priority, standard schedule (SLA: 1 week)
priority: low       → Nice to have, backlog (SLA: unbounded)
```

### Status Labels

```
status: stable      → Waiting for team analysis
status: stable → Team actively investigating
status: stable     → Waiting for external dependency
status: stable       → Ready for implementation
status: stable-progress → Currently being worked on
status: stable      → In code review
status: stable     → In QA/testing
status: stable        → Completed and verified
status: stable     → Intentionally not fixing
status: stable   → Duplicate of another issue
```

### Component Labels

```
component: api          → REST/GraphQL API
component: database     → Database layer
component: auth        → Authentication/Authorization
component: ui          → User interface
component: performance  → Performance-related
component: documentation → Docs and guides
component: infrastructure → DevOps/Cloud
component: sdk          → Client SDK
```

### Special Labels

```
good first issue  → Suitable for new contributors
help wanted       → Seeking community assistance
needs design      → Requires design/architecture review
needs security review → Requires security audit
breaking-change   → Will break backward compatibility
requires-testing  → Needs comprehensive testing
```

---

## AI-Powered Auto-Labeling

### Detection Heuristics

```
Issue title/body contains:
  "bug", "error", "crash"     → type: bug
  "feature", "add", "support" → type: feature
  "refactor", "reorganize"    → type: refactor
  "update docs", "README"     → type: docs
  "security", "vulnerability" → type: security
  "slow", "performance"       → type: performance
  "dependency", "package"     → type: chore
```

### Severity Assessment

```
Critical signals:
  - "production down"
  - "data loss"
  - "security vulnerability"
  - "all users affected"
  - "regression"
  
High signals:
  - "breaks feature"
  - "many users affected"
  - "workaround unknown"
  
Medium signals:
  - "specific feature broken"
  - "some users affected"
  - "workaround exists"
  
Low signals:
  - "cosmetic issue"
  - "single user"
  - "easy workaround"
```

---

## Label Workflow Automation

### Triage Workflow

```
New Issue
    ↓
Auto-labeled (AI classification)
    ↓
[Label confirmed?]
    ├─ Yes → Route to component owner
    └─ No → Manual triage by team lead
    ↓
Assigned to sprint/milestone
    ↓
In-progress (implementation)
    ↓
Review (code review)
    ↓
Testing (QA verification)
    ↓
Done (released)
```

### Label Transition Rules

```
triage → investigating → [blocked|ready]
  ↓
ready → in-progress → review → testing → done

Blocked → ready (dependency resolved)
WontFix → closed (decision made)
Duplicate → linked to original
```

---

## Best Practices

### DO
- ✅ Use exactly 5-8 labels per issue (minimal, curated)
- ✅ Always include: type + priority + status
- ✅ Use component labels for multi-repo tracking
- ✅ Update status as work progresses
- ✅ Use "blocking" relationships for dependencies
- ✅ Review and prune unused labels monthly
- ✅ Link duplicate issues
- ✅ Add assignee before "in-progress"

### DON'T
- ❌ Use 20+ labels per issue (too much metadata)
- ❌ Create labels for single issues (not scalable)
- ❌ Leave issues in "triage" indefinitely
- ❌ Use labels instead of milestones
- ❌ Change priority without discussion
- ❌ Add "working on it" without in-progress label
- ❌ Forget to update status as issue progresses

---

## Related Skills

- `moai-core-practices` (Workflow patterns)
- `moai-foundation-specs` (Issue specification)

---

**For detailed label reference**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-18  
**Status**: Production Ready (Enterprise )
