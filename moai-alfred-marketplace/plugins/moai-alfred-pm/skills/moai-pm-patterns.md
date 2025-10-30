# Skill: moai-pm-patterns

**Purpose**: Provides project management patterns and best practices for the PM Plugin, including SPEC generation patterns, charter structures, and risk assessment strategies.

**Status**: Stable
**Version**: 1.0.0
**Applies to**: PM Plugin (/init-pm), SPEC generation, project governance

---

## Quick Reference

### The Three Core PM Patterns

| Pattern | Use Case | Risk Level | Files Created | When to Use |
|---------|----------|-----------|--------|-----------|
| **Minimal** | Quick projects | Low | spec.md, plan.md, acceptance.md | Internal tools, prototypes |
| **Standard** | Regular features | Medium | + charter.md | Production features |
| **Enterprise** | Critical systems | High | + governance, +10 risks | Mission-critical systems |

### Command Cheat Sheet

```bash
# Minimal pattern
/init-pm tool-name --skip-charter --risk-level=low

# Standard pattern (default)
/init-pm feature-name

# Enterprise pattern
/init-pm system-name --template=enterprise --risk-level=high
```

---

## Pattern 1: Minimal Project Pattern

**Best for**: Internal tools, quick prototypes, solo projects

**Characteristics**:
- 3 core documents (spec.md, plan.md, acceptance.md)
- No charter.md (governance skipped)
- Low risk assessment (3 risks)
- Faster setup, less documentation overhead

**Command**:
```bash
/init-pm my-tool --skip-charter --risk-level=low
```

**Generated Structure**:
```
.moai/specs/SPEC-MY-TOOL-001/
├── spec.md (EARS: Ubiquitous, Event-driven, State-driven)
├── plan.md (5-phase, but brief)
├── acceptance.md (basic metrics)
└── risk-matrix.json (3 risks)
```

**Example Use Cases**:
- CLI utilities
- Internal scripts
- Development tools
- Testing utilities
- Quick POCs

**Key EARS Requirements for Tools**:

```yaml
Ubiquitous:
  - GIVEN user runs the tool
  - WHEN valid arguments provided
  - THEN tool executes successfully

Event-Driven:
  - WHEN tool receives input
  - THEN tool processes and outputs result

State-Driven:
  - GIVEN tool is initialized
  - WHEN tool completes execution
  - THEN tool exits cleanly

Optional:
  - GIVEN logging is enabled (optional)
  - WHEN tool executes
  - THEN detailed logs are written

Unwanted:
  - GIVEN invalid arguments
  - WHEN user runs tool
  - THEN helpful error message shown
```

**Risk Profile**:
1. User input validation errors
2. System dependency unavailable
3. Performance/timeout issues

---

## Pattern 2: Standard Project Pattern

**Best for**: Regular feature development, API services, web applications

**Characteristics**:
- 5 documents including charter.md
- Medium risk assessment (6 risks)
- Balanced governance and implementation focus
- Most common pattern for team projects

**Command** (default):
```bash
/init-pm my-api
# OR explicitly:
/init-pm my-api --template=moai-spec --risk-level=medium
```

**Generated Structure**:
```
.moai/specs/SPEC-MY-API-001/
├── spec.md (complete EARS specification)
├── plan.md (detailed 5-phase plan)
├── acceptance.md (quality metrics, sign-off)
├── charter.md (governance, stakeholders, budget)
└── risk-matrix.json (6 identified risks)
```

**Example Use Cases**:
- REST/GraphQL APIs
- Web features
- Microservices
- Library packages
- SaaS features

**Key EARS Requirements for APIs**:

```yaml
Ubiquitous:
  - GIVEN valid API key
  - WHEN authenticated request received
  - THEN user data returned or created

Event-Driven:
  - WHEN user creates resource via POST
  - THEN notification service is triggered

Event-Driven (Alternative):
  - WHEN API rate limit exceeded
  - THEN 429 Too Many Requests returned

State-Driven:
  - GIVEN user is authenticated
  - WHEN user is inactive for 1 hour
  - THEN session token expires

State-Driven (Alternative):
  - GIVEN order in "pending" state
  - WHEN payment received
  - THEN order transitions to "confirmed"

Optional:
  - GIVEN caching is enabled (optional)
  - WHEN GET request received
  - THEN cached response returned if available

Optional (Alternative):
  - GIVEN webhook notifications configured (optional)
  - WHEN resource changed
  - THEN webhook payload sent to registered URL

Unwanted:
  - GIVEN unauthorized API key
  - WHEN attempting protected endpoint
  - THEN 401 Unauthorized response

Unwanted (Alternative):
  - GIVEN malicious SQL in input
  - WHEN query executed
  - THEN SQL injection prevented by parameterization
```

**Stakeholder Charter Template**:

```yaml
Project Manager: [Name]
  - Execution responsibility
  - Timeline & resource management
  - Stakeholder communication

Tech Lead: [Name]
  - Architecture decisions
  - Technical implementation oversight
  - Code quality standards

QA Lead: [Name]
  - Test planning & execution
  - Quality metrics tracking
  - Bug prioritization

Product Owner: [Name]
  - Feature requirements
  - Business priorities
  - Acceptance criteria

Sponsor/Executive: [Name]
  - Strategic alignment
  - Budget approval
  - Escalation authority
```

**Risk Pattern**:
1. Technical implementation risks (3-4)
   - Architecture complexity
   - Technology selection
   - Integration challenges
2. Process risks (1-2)
   - Timeline slippage
   - Resource constraints
3. External risks (1)
   - Dependency delays
   - Scope creep

**5-Phase Timeline** (Standard 12 weeks):
```
Week 1: Kickoff + Planning
Week 2-3: Design & Architecture
Week 4-8: Development & Testing (5 weeks)
Week 9-10: Validation & UAT
Week 11-12: Release & Documentation
```

---

## Pattern 3: Enterprise Project Pattern

**Best for**: Mission-critical systems, large organizations, regulated industries

**Characteristics**:
- Full 5-document specification
- Enterprise template (full governance)
- High risk assessment (10+ risks)
- Comprehensive stakeholder management
- Detailed budget and schedule tracking

**Command**:
```bash
/init-pm critical-system --template=enterprise --risk-level=high
```

**Generated Structure**:
```
.moai/specs/SPEC-CRITICAL-SYSTEM-001/
├── spec.md (detailed EARS with all 5 patterns)
├── plan.md (detailed milestones, dependencies)
├── acceptance.md (comprehensive sign-off)
├── charter.md (full governance structure)
└── risk-matrix.json (10+ risks with mitigation)
```

**Example Use Cases**:
- Payment systems
- Healthcare platforms
- Financial services
- Regulatory compliance systems
- Mission-critical infrastructure

**Key EARS Requirements for Enterprise Systems**:

```yaml
Ubiquitous:
  - GIVEN system is operational
  - WHEN legitimate transaction submitted
  - THEN transaction processed within SLA

Event-Driven (Primary):
  - WHEN transaction received
  - THEN fraud detection rules applied

Event-Driven (Secondary):
  - WHEN threshold exceeded
  - THEN escalation alert sent to operations

State-Driven (Primary):
  - GIVEN transaction in "pending" state
  - WHEN risk assessment complete
  - THEN transaction moves to "processing" state

State-Driven (Secondary):
  - GIVEN system in "maintenance" mode
  - WHEN scheduled maintenance window ends
  - THEN system returns to "operational" state

Optional (Compliance):
  - GIVEN audit logging enabled (required for compliance)
  - WHEN transaction processed
  - THEN audit record created and retained per regulations

Optional (Enhancement):
  - GIVEN multi-currency support enabled (optional feature)
  - WHEN transaction in foreign currency
  - THEN automatic conversion applied

Unwanted (Security):
  - GIVEN unauthorized user
  - WHEN attempting to access customer data
  - THEN access denied, security alert logged

Unwanted (Data Integrity):
  - GIVEN incomplete transaction record
  - WHEN database query executed
  - THEN partial record prevented by constraints

Unwanted (Performance):
  - GIVEN system under high load
  - WHEN additional request received
  - THEN graceful degradation occurs, not crash
```

**Comprehensive Governance Structure**:

```yaml
Executive Sponsor: [C-level executive]
  - Strategic oversight
  - Budget approval authority
  - Escalation point for critical issues
  - Board reporting

Project Manager: [Senior PM]
  - Day-to-day execution
  - Stakeholder coordination
  - Timeline management
  - Risk tracking & reporting

Architect/Tech Lead: [Senior engineer]
  - Technical design
  - Architecture decisions
  - Technology selection
  - Technical risk management

QA Lead: [Senior QA]
  - Test strategy & planning
  - Quality standards
  - Compliance validation
  - Sign-off authority

Compliance Officer: [Regulatory]
  - Regulatory requirements
  - Compliance verification
  - Audit preparation
  - Documentation

Change Manager: [Change specialist]
  - User impact analysis
  - Training strategy
  - Adoption planning
  - Support preparation

Operations Lead: [Operations]
  - Deployment readiness
  - Production support
  - Disaster recovery
  - Monitoring & alerting
```

**Comprehensive Risk Management**:

```json
{
  "risks": [
    {
      "id": "RISK-001",
      "description": "Data loss or corruption",
      "category": "Technical",
      "probability": "Low",
      "impact": "Critical",
      "owner": "Database Administrator",
      "mitigation": "Daily backups, master-slave replication, disaster recovery testing quarterly",
      "status": "Identified"
    },
    {
      "id": "RISK-002",
      "description": "System outage during peak business hours",
      "category": "Technical",
      "probability": "Medium",
      "impact": "Critical",
      "owner": "DevOps Lead",
      "mitigation": "Multi-region failover, load balancing, chaos engineering testing",
      "status": "Identified"
    },
    {
      "id": "RISK-003",
      "description": "Security breach or unauthorized access",
      "category": "Security",
      "probability": "Medium",
      "impact": "Critical",
      "owner": "Security Lead",
      "mitigation": "Penetration testing, encryption, access controls, audit logging",
      "status": "Identified"
    },
    {
      "id": "RISK-004",
      "description": "Regulatory compliance violation",
      "category": "Compliance",
      "probability": "Low",
      "impact": "Critical",
      "owner": "Compliance Officer",
      "mitigation": "Legal review, compliance audit, documentation, training",
      "status": "Identified"
    },
    {
      "id": "RISK-005",
      "description": "Integration with legacy systems fails",
      "category": "Technical",
      "probability": "High",
      "impact": "High",
      "owner": "Integration Lead",
      "mitigation": "Early integration testing, API contracts, fallback mechanisms",
      "status": "Identified"
    }
  ]
}
```

**Extended 16-week Timeline**:
```
Week 1-2: Kickoff + Governance Setup
Week 3-4: Design + Compliance Review
Week 5-10: Development + Integration (6 weeks)
Week 11-13: Validation + UAT + Compliance Testing
Week 14-16: Release Prep + Cutover + Stabilization
```

**Budget & Resource Planning**:

```yaml
Team Composition:
  - Project Manager: 1 FTE
  - Tech Lead: 1 FTE
  - Developers: 4-6 FTE
  - QA Engineers: 2-3 FTE
  - DevOps: 1-2 FTE
  - Business Analyst: 1 FTE
  - Compliance Officer: 0.5 FTE
  - Total: 11-15 FTE

Budget Allocation:
  - Development: 40%
  - Infrastructure/Tools: 20%
  - Testing/QA: 20%
  - Operations/Support: 15%
  - Contingency (20%): Reserved

Timeline Dependency:
  - Kickoff → Design (2 weeks)
  - Design → Development (concurrent with compliance)
  - Development → Integration Testing (1-2 weeks overlap)
  - Testing → UAT (sequential)
  - UAT → Release (cutover planning, 1-2 weeks)
```

---

## Best Practices by Pattern

### Minimal Pattern Best Practices

✅ **DO**:
- Start with minimal, expand if needed
- Use for internal/private projects
- Document critical flows even if brief
- Test thoroughly despite minimal overhead

❌ **DON'T**:
- Skip requirements documentation entirely
- Use for external/customer-facing systems
- Skip risk assessment completely
- Ignore acceptance criteria

### Standard Pattern Best Practices

✅ **DO**:
- Define clear stakeholder roles
- Estimate realistic timelines (12 weeks typical)
- Identify 5-8 key risks
- Get stakeholder approval on charter before implementation
- Track progress through phases

❌ **DON'T**:
- Underestimate implementation time
- Skip stakeholder communication
- Ignore emerging risks
- Change scope mid-project without approval

### Enterprise Pattern Best Practices

✅ **DO**:
- Establish governance structure upfront
- Assign risk owners before project starts
- Plan for contingencies (20% buffer)
- Do compliance review during design phase
- Establish clear escalation paths
- Regular executive steering committee updates
- Comprehensive testing strategy

❌ **DON'T**:
- Skip executive alignment
- Underestimate regulatory requirements
- Skip penetration testing/security review
- Begin development without design approval
- Ignore legacy system compatibility
- Rush the release/cutover process

---

## Pattern Selection Decision Tree

```
Starting new project?
├─ Is it an internal tool or quick prototype?
│  └─ YES → Use MINIMAL PATTERN
│          /init-pm name --skip-charter --risk-level=low
│
├─ Is it a standard feature or API?
│  └─ YES → Use STANDARD PATTERN
│          /init-pm name
│
└─ Is it mission-critical or highly regulated?
   └─ YES → Use ENTERPRISE PATTERN
           /init-pm name --template=enterprise --risk-level=high
```

---

## Integration with Alfred Workflow

After generating SPEC with PM Plugin:

```
1. /init-pm my-project
   ↓
2. Customize SPEC documents
   ├─ Edit spec.md (EARS requirements)
   ├─ Edit charter.md (stakeholders, budget)
   └─ Edit risk-matrix.json (identified risks)
   ↓
3. Get stakeholder approval
   ├─ Share spec.md for review
   ├─ Get charter sign-off
   └─ Confirm risk assessment
   ↓
4. /alfred:2-run SPEC-MY-PROJECT-001
   (Implement according to SPEC)
   ↓
5. /alfred:3-sync auto SPEC-MY-PROJECT-001
   (Synchronize documentation with implementation)
```

---

## Common Pattern Mistakes

### Mistake 1: Using Minimal for Complex Projects

❌ **WRONG**: `/init-pm payment-system --skip-charter --risk-level=low`
✅ **CORRECT**: `/init-pm payment-system --template=enterprise --risk-level=high`

**Why**: Payment systems need governance, stakeholder alignment, comprehensive risk management.

### Mistake 2: Over-engineering Simple Projects

❌ **WRONG**: `/init-pm helper-script --template=enterprise --risk-level=high`
✅ **CORRECT**: `/init-pm helper-script --skip-charter --risk-level=low`

**Why**: Simple tools don't need enterprise governance overhead.

### Mistake 3: Incomplete EARS Specification

❌ **WRONG**:
```yaml
Ubiquitous:
  - Feature should work
  - System should be fast
```

✅ **CORRECT**:
```yaml
Ubiquitous:
  - GIVEN user is authenticated
  - WHEN user requests report
  - THEN report is generated within 10 seconds
```

### Mistake 4: Vague Risk Assessment

❌ **WRONG**:
```json
{
  "id": "RISK-001",
  "description": "Things might go wrong"
}
```

✅ **CORRECT**:
```json
{
  "id": "RISK-001",
  "description": "Database migration fails, causing data loss",
  "probability": "Medium",
  "impact": "Critical",
  "mitigation": "Backup all data, test migration on staging first"
}
```

### Mistake 5: Ignoring Timeline Reality

❌ **WRONG**: Keep default 12-week plan for 4-week project
✅ **CORRECT**: Update plan.md to reflect actual timeline

---

## Quick Lookup Table

| Need | Pattern | Command |
|------|---------|---------|
| Quick tool | Minimal | `/init-pm tool --skip-charter --risk-level=low` |
| Web API | Standard | `/init-pm api` |
| Feature | Standard | `/init-pm feature` |
| Payment system | Enterprise | `/init-pm payments --template=enterprise --risk-level=high` |
| Healthcare | Enterprise | `/init-pm health --template=enterprise --risk-level=high` |
| Internal script | Minimal | `/init-pm script --skip-charter` |
| Microservice | Standard | `/init-pm service` |
| Data pipeline | Standard | `/init-pm pipeline` |

---

**Skill Version**: 1.0.0
**Last Updated**: 2025-10-30
**Generated with [Claude Code](https://claude.com/claude-code)**
**Co-Authored-By**: 🎩 Alfred <alfred@mo.ai.kr>
