# PM Plugin - Usage Guide

Practical examples and workflows for using PM Plugin effectively in your projects.

## Table of Contents

1. [Basic Workflow](#basic-workflow)
2. [Use Case Examples](#use-case-examples)
3. [Best Practices](#best-practices)
4. [Integration Patterns](#integration-patterns)
5. [Troubleshooting Guide](#troubleshooting-guide)

---

## Basic Workflow

### Step 1: Initialize Project with PM Plugin

```bash
/init-pm my-new-project
```

**What happens**:
- Validates project name format
- Creates `.moai/specs/SPEC-MY-NEW-PROJECT-001/` directory
- Generates 5 documents (spec.md, plan.md, acceptance.md, charter.md, risk-matrix.json)
- Displays completion summary

**Output**:
```
✅ Project 'my-new-project' initialized successfully
📁 Location: .moai/specs/SPEC-MY-NEW-PROJECT-001/
📊 Risk Level: medium
📋 Template: moai-spec
📝 Files created: 5
```

### Step 2: Review Generated Specification

```bash
# View the EARS specification
cat .moai/specs/SPEC-MY-NEW-PROJECT-001/spec.md

# Review the implementation plan
cat .moai/specs/SPEC-MY-NEW-PROJECT-001/plan.md

# Check acceptance criteria
cat .moai/specs/SPEC-MY-NEW-PROJECT-001/acceptance.md
```

### Step 3: Customize for Your Project

Edit the generated files to customize for your specific project:

```bash
# Edit SPEC requirements
code .moai/specs/SPEC-MY-NEW-PROJECT-001/spec.md

# Adjust implementation plan
code .moai/specs/SPEC-MY-NEW-PROJECT-001/plan.md

# Refine acceptance criteria
code .moai/specs/SPEC-MY-NEW-PROJECT-001/acceptance.md

# Update stakeholder information
code .moai/specs/SPEC-MY-NEW-PROJECT-001/charter.md
```

### Step 4: Implement According to Plan

```bash
# Run Alfred's implementation workflow
/alfred:2-run SPEC-MY-NEW-PROJECT-001
```

### Step 5: Synchronize Documentation

```bash
# Sync implementation with specifications
/alfred:3-sync auto SPEC-MY-NEW-PROJECT-001
```

---

## Use Case Examples

### Example 1: Simple Internal Tool

**Scenario**: Building a quick internal utility with minimal process.

```bash
/init-pm inventory-tracker --risk-level=low --skip-charter
```

**Result**:
- Creates minimal SPEC (no charter.md)
- Risk matrix with only 3 identified risks
- Faster setup for small scope projects
- Perfect for internal tools and prototypes

**Why this approach**:
- No governance overhead needed
- Reduces template bloat
- Still maintains EARS requirements specification
- Risk management still in place

---

### Example 2: Enterprise Payment System

**Scenario**: Building a critical financial system requiring strict governance.

```bash
/init-pm payment-gateway --template=enterprise --risk-level=high
```

**Result**:
- Creates enterprise-grade specification
- Comprehensive charter with governance structure
- 10+ identified risks in risk-matrix.json
- Detailed stakeholder matrix
- Full 5-phase implementation plan

**Customization steps**:

1. **Fill in Stakeholder Matrix** (charter.md):
   ```markdown
   | Stakeholder | Role | Responsibility | Contact |
   |-----------|------|-----------------|---------|
   | John Smith | Executive Sponsor | Strategic direction | john@company.com |
   | Jane Doe | Project Manager | Execution | jane@company.com |
   | Mike Chen | Tech Lead | Architecture | mike@company.com |
   | Sarah Johnson | QA Lead | Testing | sarah@company.com |
   ```

2. **Identify Critical Risks** (risk-matrix.json):
   - Payment processing failures
   - Data security breaches
   - Regulatory compliance violations
   - System downtime during peak load
   - Integration with legacy systems

3. **Set Budget & Schedule** (charter.md):
   - Total Budget: $500K
   - Duration: 4 months
   - Contingency: 20%

4. **Define EARS Requirements** (spec.md):
   - Ubiquitous: Payment processing always available
   - Event-driven: When payment received, notification sent
   - State-driven: Cannot refund if payment not received
   - Optional: Multi-currency support
   - Unwanted: No unauthorized fund transfers

---

### Example 3: Agile Mobile App

**Scenario**: Building a consumer mobile app with iterative approach.

```bash
/init-pm social-media-app --template=agile --risk-level=medium
```

**Result**:
- Creates agile-focused specification
- Lighter governance (still charter.md but simplified)
- Medium risk assessment (6 identified risks)
- Implementation plan structured for sprints
- Regular milestone-based deliverables

**Sprint-based customization**:

```yaml
# Modified Phase 3: Implementation (from plan.md)

## Sprint 1 (Week 4-5): User Authentication
- GIVEN user launches app
- WHEN entering credentials
- THEN user is authenticated

## Sprint 2 (Week 6-7): Feed Display
- GIVEN user is authenticated
- WHEN accessing home page
- THEN feed displays in real-time

## Sprint 3 (Week 8-9): Social Interactions
- GIVEN user views post
- WHEN clicking like/comment
- THEN interaction is recorded and broadcast

## Sprint 4 (Week 10): Polish & Release
- GIVEN all features implemented
- WHEN running full test suite
- THEN 0 critical bugs remain
```

---

### Example 4: API Service Development

**Scenario**: Building a REST API with clear documentation requirements.

```bash
/init-pm user-service-api --template=moai-spec --risk-level=medium
```

**EARS specification customization for APIs**:

```markdown
## Ubiquitous Behaviors

**Feature 1: User CRUD Operations**
- GIVEN a valid API key
- WHEN making authenticated requests
- THEN user data can be created, read, updated, deleted

## Event-Driven Behaviors

**Event 1: User Created**
- WHEN a new user is created via POST /users
- THEN notification service is triggered

**Event 2: User Updated**
- WHEN user profile is updated via PUT /users/{id}
- THEN audit log entry is created

## State-Driven Behaviors

**State 1: User Status Flow**
- GIVEN user is in "active" state
- WHEN user calls DELETE /users/{id}
- THEN user transitions to "inactive" state

## Optional Behaviors

**Optional 1: Rate Limiting**
- GIVEN rate limiting is enabled
- WHEN API threshold exceeded (optional feature)
- THEN requests are throttled with 429 response

## Unwanted Behaviors

**Unwanted 1: SQL Injection Prevention**
- GIVEN an API endpoint
- WHEN user provides malicious SQL input
- THEN query is parameterized and injection prevented

**Unwanted 2: Authentication Failure**
- GIVEN invalid API key
- WHEN requesting protected endpoint
- THEN 401 Unauthorized response returned
```

---

### Example 5: Data Migration Project

**Scenario**: Migrating legacy system to modern platform.

```bash
/init-pm legacy-to-modern-migration --template=enterprise --risk-level=high
```

**Risk matrix customization for migration**:

```json
{
  "risks": [
    {
      "id": "RISK-001",
      "description": "Data loss during migration",
      "category": "Technical",
      "probability": "High",
      "impact": "High",
      "mitigation": "Backup all data before migration, run parallel systems for 2 weeks",
      "owner": "Database Administrator",
      "status": "Identified"
    },
    {
      "id": "RISK-002",
      "description": "System downtime during cutover",
      "category": "Technical",
      "probability": "Medium",
      "impact": "High",
      "mitigation": "Plan cutover during low-traffic window, prepare rollback plan",
      "owner": "DevOps Lead",
      "status": "Identified"
    },
    {
      "id": "RISK-003",
      "description": "User adoption resistance",
      "category": "Process",
      "probability": "High",
      "impact": "Medium",
      "mitigation": "Comprehensive training, help desk support, documentation",
      "owner": "Change Management",
      "status": "Identified"
    }
  ]
}
```

---

## Best Practices

### 1. Project Naming Convention

**✅ Good names**:
- `user-authentication-service` - Clear purpose
- `payment-gateway` - Specific domain
- `mobile-app-v2` - Version indicator
- `data-pipeline-etl` - Technology + purpose

**❌ Avoid**:
- `MyProject` - Uses uppercase (not allowed)
- `project` - Too generic
- `a_project_name` - Uses underscores (use hyphens)
- `project name` - Uses spaces (not allowed)
- `p` - Too short (minimum 3 chars)

### 2. Risk Level Selection

**Choose LOW risk level when**:
- Building internal tools
- Small scope projects
- Prototyping new features
- Low business impact if delayed

**Choose MEDIUM risk level when**:
- Standard feature development
- Moderate scope projects
- Team familiar with similar projects
- Business impact manageable

**Choose HIGH risk level when**:
- Critical business systems
- Large-scale architecture changes
- New technology adoption
- Significant compliance requirements
- High financial impact

### 3. Template Selection

**Choose moai-spec (default) when**:
- Standard development project
- Balanced governance needs
- Most common use case

**Choose enterprise when**:
- Large organizations
- Regulatory compliance required
- Multiple stakeholder approvals needed
- Budget/resource tracking critical

**Choose agile when**:
- Fast-moving teams
- Iterative development
- Sprint-based planning
- Regular release cycles

### 4. Charter vs. Skip-Charter

**Create charter (default) when**:
- Multi-team projects
- External stakeholders
- Formal approval required
- Budget tracking needed

**Skip charter when**:
- Solo developer projects
- Quick prototypes
- Internal tools
- Rapid iteration focus

### 5. EARS Requirement Writing

**✅ Good EARS format**:
```
GIVEN a user is logged in
WHEN they click the "checkout" button
THEN the shopping cart is submitted for payment
```

**❌ Poor format**:
```
System should checkout when user clicks button
The checkout feature must work
Enable users to pay
```

**Key guidelines**:
- Use GIVEN/WHEN/THEN explicitly
- One requirement per line
- Focus on behavior, not implementation
- Include both success and failure scenarios

### 6. Risk Matrix Maintenance

**Keep risk matrix updated**:
- Review weekly during project execution
- Update status as mitigation strategies deployed
- Add new risks as they emerge
- Close risks once fully mitigated

**Risk status lifecycle**:
```
Identified → Mitigated → Closed
```

### 7. Documentation Synchronization

**Sync frequently**:
```bash
# After major implementation phase
/alfred:3-sync auto SPEC-MY-PROJECT-001

# Before stakeholder reviews
/alfred:3-sync force SPEC-MY-PROJECT-001

# Check sync status
/alfred:3-sync status SPEC-MY-PROJECT-001
```

---

## Integration Patterns

### Pattern 1: SPEC-First Development

1. Initialize project with PM Plugin
2. Customize EARS requirements
3. Get stakeholder approval on spec.md
4. Run `/alfred:2-run` for implementation
5. Sync documentation with `/alfred:3-sync`

### Pattern 2: Phased Rollout

```bash
# Create base SPEC
/init-pm ecommerce-platform --risk-level=high

# Phase 1 implementation
/alfred:2-run SPEC-ECOMMERCE-PLATFORM-001/phase-1

# Phase 2 implementation
/alfred:2-run SPEC-ECOMMERCE-PLATFORM-001/phase-2

# Final sync
/alfred:3-sync auto SPEC-ECOMMERCE-PLATFORM-001
```

### Pattern 3: Multiple Parallel Features

```bash
# Create specs for parallel work streams
/init-pm payment-feature --template=enterprise
/init-pm reporting-feature --template=enterprise
/init-pm notification-feature --template=enterprise

# Parallel implementation
/alfred:2-run SPEC-PAYMENT-FEATURE-001 &
/alfred:2-run SPEC-REPORTING-FEATURE-001 &
/alfred:2-run SPEC-NOTIFICATION-FEATURE-001 &

# Sequential sync
/alfred:3-sync auto SPEC-PAYMENT-FEATURE-001
/alfred:3-sync auto SPEC-REPORTING-FEATURE-001
/alfred:3-sync auto SPEC-NOTIFICATION-FEATURE-001
```

### Pattern 4: Risk-Driven Implementation

1. Generate high-risk SPEC
2. Identify 10+ risks in risk-matrix.json
3. Assign risk owners for each mitigation strategy
4. Execute implementation with risk tracking
5. Update risk status weekly
6. Close risks upon mitigation completion

---

## Troubleshooting Guide

### Issue: "Project name must contain only lowercase letters"

**Problem**: Attempted to create project with uppercase or special characters.

**Solution**:
```bash
# ❌ Wrong
/init-pm MyAwesomeProject

# ✅ Correct
/init-pm my-awesome-project
```

### Issue: "SPEC already exists"

**Problem**: Project with same name already initialized.

**Solutions**:

Option 1 - Use version suffix:
```bash
/init-pm my-project-v2
```

Option 2 - Remove existing SPEC:
```bash
rm -rf .moai/specs/SPEC-MY-PROJECT-001/
/init-pm my-project
```

Option 3 - Keep existing, create feature branch:
```bash
/init-pm my-project-feature-1
/init-pm my-project-feature-2
```

### Issue: Charter file too long or not customized

**Problem**: Generated charter.md has placeholders but needs customization.

**Solution**: Edit the file directly and fill in your project details:

```bash
code .moai/specs/SPEC-MY-PROJECT-001/charter.md

# Edit these sections:
# - Stakeholder Matrix (add actual names/roles)
# - Budget & Schedule (add realistic numbers)
# - Business Case (define actual objectives)
```

### Issue: Risk matrix doesn't match our risks

**Problem**: Generated risks are generic and don't match your project.

**Solution**: Edit risk-matrix.json with your specific risks:

```bash
code .moai/specs/SPEC-MY-PROJECT-001/risk-matrix.json

# Replace generic risks with:
# - Your identified project risks
# - Your actual mitigation strategies
# - Your team members' assignments
```

### Issue: EARS requirements seem vague

**Problem**: Generated EARS patterns are too generic.

**Solution**: Customize them to be specific to your project:

```bash
# From generic:
GIVEN the project is initialized
WHEN project charter is created
THEN stakeholder roles are defined

# To specific:
GIVEN payment processing system is initialized
WHEN payment request received from e-commerce platform
THEN transaction is validated against fraud detection rules
```

### Issue: Implementation plan timeline doesn't match our schedule

**Problem**: 12-week default plan doesn't fit your project duration.

**Solution**: Edit plan.md with your actual timeline:

```markdown
# Adjust phases based on your schedule

## Phase 1: Kickoff (2 weeks)
## Phase 2: Design (1 week)
## Phase 3: Implementation (3 weeks)
## Phase 4: Validation (1 week)
## Phase 5: Release (1 week)

# Total: 8 weeks (instead of 12)
```

---

## Command Reference Quick Lookup

| Task | Command |
|------|---------|
| Create basic project | `/init-pm my-project` |
| High-risk enterprise | `/init-pm project --template=enterprise --risk-level=high` |
| Low-risk internal tool | `/init-pm project --risk-level=low --skip-charter` |
| Agile mobile app | `/init-pm project --template=agile` |
| View SPEC | `cat .moai/specs/SPEC-{ID}/spec.md` |
| Edit SPEC | `code .moai/specs/SPEC-{ID}/spec.md` |
| Run implementation | `/alfred:2-run SPEC-{ID}-001` |
| Sync documentation | `/alfred:3-sync auto SPEC-{ID}-001` |

---

## Next Steps

After initializing a project:

1. **Customize EARS Requirements** → Edit spec.md with your specific requirements
2. **Define Stakeholders** → Add team members to charter.md
3. **Assess Risks** → Update risk-matrix.json with your identified risks
4. **Plan Phases** → Adjust plan.md timeline to your schedule
5. **Implement** → Run `/alfred:2-run SPEC-{ID}-001`
6. **Synchronize** → Run `/alfred:3-sync auto SPEC-{ID}-001`

---

**Last Updated**: 2025-10-30
**Generated with [Claude Code](https://claude.com/claude-code)**
**Co-Authored-By**: 🎩 Alfred <alfred@mo.ai.kr>
