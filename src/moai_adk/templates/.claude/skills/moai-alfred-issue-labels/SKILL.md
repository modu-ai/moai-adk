---
name: moai-alfred-issue-labels
version: 1.1.0
created: 2025-11-05
updated: 2025-11-05
status: active
description: GitHub issue label configuration, label taxonomy, and workflow automation mapping for MoAI-ADK projects
keywords: ['github', 'issues', 'labels', 'workflow', 'taxonomy', 'automation']
allowed-tools:
  - Read
  - Bash
---

# Alfred Issue Labels Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-issue-labels |
| **Version** | 1.1.0 (2025-11-05) |
| **Status** | Active |
| **Tier** | Alfred |
| **Purpose** | Standardize GitHub issue labeling and workflow automation |

---

## What It Does

Provides comprehensive GitHub issue label configuration, taxonomy management, and workflow automation mapping for MoAI-ADK projects. Ensures consistent issue categorization and priority management across all repositories.

**Key capabilities**:
- ✅ Standardized label taxonomy for all issue types
- ✅ Priority-based labeling system (critical/high/medium/low)
- ✅ Workflow automation mappings
- ✅ Consistent issue management across projects
- ✅ Integration with Alfred command workflows

---

## When to Use

**Automatic triggers**:
- Creating new issues via `/alfred:9-feedback`
- Issue triage and classification
- Project setup and configuration
- Label management and cleanup

**Manual reference**:
- Understanding label taxonomy
- Setting up new repositories
- Training team members on issue management
- Customizing workflow automation

---

## Label Taxonomy System

### Bug Issues (`--bug`)

**Primary Labels**: `bug`, `reported`

**Priority Labels**:
- `priority-critical` - System down, data loss risk
- `priority-high` - Major feature broken
- `priority-medium` - Normal bug (default)
- `priority-low` - Minor issue, cosmetic

**Usage Pattern**:
```bash
/alfred:9-feedback --bug --priority-high "Login system broken"
```

### Feature Request Issues (`--feature`)

**Primary Labels**: `feature-request`, `enhancement`

**Priority Labels**:
- `priority-critical` - Blocking, must implement immediately
- `priority-high` - Important feature
- `priority-medium` - Normal priority (default)
- `priority-low` - Nice to have

**Usage Pattern**:
```bash
/alfred:9-feedback --feature --priority-medium "Add dark mode support"
```

### Improvement Issues (`--improvement`)

**Primary Labels**: `improvement`, `enhancement`

**Priority Labels**:
- `priority-critical` - Critical refactoring needed
- `priority-high` - Important improvement
- `priority-medium` - Normal priority (default)
- `priority-low` - Technical debt, can wait

### Question/Discussion Issues (`--question`)

**Primary Labels**: `question`, `help-wanted`, `discussion`

**Usage Patterns**:
- Technical questions and clarifications
- Architecture discussions
- Feature feedback and suggestions
- Community engagement

---

## Workflow Automation Mappings

### Label-Based Automation

**Critical Priority**:
- Immediate notification to maintainers
- Auto-assignment to lead developers
- SLA: 24-hour response time

**High Priority**:
- Enhanced visibility in project boards
- Auto-add to milestone planning
- SLA: 72-hour response time

**Medium Priority**:
- Standard workflow processing
- Regular backlog grooming
- Normal development cycle

**Low Priority**:
- Future consideration queue
- Community contribution opportunities
- Best-effort timeline

---

## Integration with Alfred Commands

### `/alfred:9-feedback` Command Integration

```bash
# Create bug report with automatic labeling
/alfred:9-feedback --bug --priority-high "Critical issue description"

# Create feature request
/alfred:9-feedback --feature --priority-medium "Feature description"

# Create improvement suggestion
/alfred:9-feedback --improvement --priority-low "Improvement idea"
```

### Auto-Assignment Rules

- **Bug issues** → Assign to bug-fix team
- **Feature requests** → Assign to feature team
- **Documentation** → Assign to docs team
- **Infrastructure** → Assign to DevOps team

---

## Best Practices

✅ **DO**:
- Always select appropriate primary label type
- Use priority labels consistently
- Provide detailed descriptions
- Link related issues with dependencies
- Update labels when issue scope changes
- Follow team response SLAs

❌ **DON'T**:
- Mix primary label types (bug + feature)
- Use priority labels without justification
- Leave issues unlabeled
- Ignore label-based automation
- Create custom labels without team consensus

---

## Repository Setup

### Required Labels (Minimum Set)

```bash
# Primary type labels
bug, feature-request, improvement, question, documentation

# Priority labels  
priority-critical, priority-high, priority-medium, priority-low

# Status labels
needs-triage, in-progress, blocked, ready-for-review, done
```

### Optional Enhanced Labels

```bash
# Team labels
team-frontend, team-backend, team-devops, team-docs

# Component labels
component-ui, component-api, component-database, component-auth

# Type-specific labels
performance, security, accessibility, testing, refactoring
```

---

## Quality Metrics

**Label Compliance Rate**: Target >95%
- Issues with proper primary label
- Issues with appropriate priority
- Response time compliance by priority

**Automation Success Rate**: Target >90%
- Correct auto-assignments
- Proper workflow routing
- SLA compliance tracking

---

Learn more in `reference.md` for complete label definitions, automation workflows, and team configuration details.

**Related Skills**: 
- `moai-alfred-workflow` - Issue workflow management
- `moai-alfred-agent-guide` - Team coordination
- `moai-foundation-specs` - Requirements management
