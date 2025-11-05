---
name: moai-alfred-best-practices
version: 1.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Quality gates, compliance patterns, and mandatory rules for Alfred workflow execution. Enforces TRUST 5 principles, TAG validation, Skill invocation rules, and AskUserQuestion scenarios. Use when validating workflow compliance, checking quality gates, enforcing MoAI-ADK standards, or verifying rule adherence.
keywords: ['quality-gates', 'compliance', 'trust', 'validation', 'rules', 'standards', 'alfred']
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
---

# Alfred Best Practices & Quality Gates

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-best-practices |
| **Version** | 1.0.0 (2025-11-06) |
| **Status** | Active |
| **Tier** | Alfred Core |
| **Purpose** | Quality gates and compliance enforcement |

---

## What It Does

Enforces mandatory rules, quality standards, and compliance patterns for Alfred workflow execution.

**Core capabilities**:
- ✅ TRUST 5 quality gate validation
- ✅ TAG chain integrity checking
- ✅ Mandatory Skill invocation rules
- ✅ AskUserQuestion scenario guidelines
- ✅ Context engineering best practices
- ✅ Workflow compliance validation

---

## When to Use

**Quality Gates**:
- Before marking tasks complete
- During code review phases
- When validating implementation quality
- Before git commits and PRs

**Compliance Validation**:
- Checking if mandatory Skills were invoked
- Validating TAG chain integrity
- Ensuring proper workflow execution
- Verifying rule adherence

**Pattern Recognition**:
- Identifying workflow violations
- Detecting missing quality gates
- Finding non-compliant patterns
- Recommending corrective actions

---

## TRUST 5 Quality Gates

### The 5 Pillars

| Pillar | Requirement | Validation Method |
|--------|-------------|-------------------|
| **T** – Test-Driven | 85%+ test coverage required | Run test coverage analysis |
| **R** – Readable | No code smells, SOLID principles | Linting + static analysis |
| **U** – Unified | Consistent patterns, no duplicates | Pattern matching + duplication detection |
| **S** – Secured | OWASP Top 10 compliance, no secrets | Security scanning + secret detection |
| **E** – Evaluated | @TAG chain intact (SPEC→TEST→CODE→DOC) | TAG validation tools |

### Gate Validation Checklist

**Before Task Completion**:
- [ ] Tests pass with ≥85% coverage
- [ ] Code follows SOLID principles
- [ ] No duplicate code patterns detected
- [ ] Security scan passes (OWASP Top 10)
- [ ] All @TAG references resolved
- [ ] Documentation matches implementation

**Git Commit Requirements**:
- [ ] Commit message follows format
- [ ] Alfred co-authorship included
- [ ] @TAG references included in commit
- [ ] Quality gate status documented

---

## Mandatory Skill Invocations

### 10 Required Skills (Always Invoke)

| Context | Skill | Invocation Pattern | Validation |
|---------|-------|-------------------|------------|
| Quality validation | `moai-foundation-trust` | `Skill("moai-foundation-trust")` | Before completion |
| TAG system | `moai-foundation-tags` | `Skill("moai-foundation-tags")` | TAG creation/updates |
| SPEC documents | `moai-foundation-specs` | `Skill("moai-foundation-specs")` | SPEC authoring/validation |
| Requirements | `moai-foundation-ears` | `Skill("moai-foundation-ears")` | EARS syntax/formatting |
| Git workflow | `moai-foundation-git` | `Skill("moai-foundation-git")` | Branch/commit operations |
| Language detection | `moai-foundation-langs` | `Skill("moai-foundation-langs")` | Stack detection/analysis |
| Debugging | `moai-essentials-debug` | `Skill("moai-essentials-debug")` | Error analysis/triage |
| Refactoring | `moai-essentials-refactor` | `Skill("moai-essentials-refactor")` | Code improvements |
| Performance | `moai-essentials-perf` | `Skill("moai-essentials-perf")` | Optimization tasks |
| Code review | `moai-essentials-review` | `Skill("moai-essentials-review")` | Quality reviews |

### Skill Invocation Rules

**Mandatory (Must Call)**:
- When specific context applies (see table above)
- Before task completion in relevant domain
- When quality gates require validation

**Optional (Consider Calling)**:
- Context-specific domain skills
- Language/framework-specific skills
- Project-specific custom skills

**Forbidden (Never Call Directly)**:
- Internal utility skills
- Deprecated skills (check migration guide)
- Skills merged into consolidations

---

## AskUserQuestion Scenarios

### 5 Mandatory Question Scenarios

Use `AskUserQuestion` when:

1. **Tech Stack Unclear**
   - Multiple frameworks/languages possible
   - Technology choice affects implementation
   - Example: "Should we use React or Vue for frontend?"

2. **Architecture Decision Needed**
   - Monolith vs microservices
   - Database design choices
   - Integration patterns
   - Example: "Should we use REST API or GraphQL?"

3. **User Intent Ambiguous**
   - Multiple valid interpretations
   - Business logic unclear
   - Feature scope undefined
   - Example: "Improve user experience" → Which specific aspect?

4. **Existing Component Impacts Unknown**
   - Potential breaking changes
   - Integration points unclear
   - Migration requirements
   - Example: "Will this change affect existing APIs?"

5. **Resource Constraints Unclear**
   - Budget limitations
   - Timeline constraints
   - Team capacity
   - Example: "Is this feature needed for MVP or can it wait?"

### Question Design Rules

**Format Requirements**:
- Present exactly 3-5 options (never open-ended)
- Use structured format with header and description
- Include clear action-oriented choices
- Avoid technical jargon when possible

**Example Structure**:
```json
{
  "header": "Architecture Approach Selection",
  "question": "Which architecture pattern should we use for this feature?",
  "options": [
    {
      "label": "Microservices",
      "description": "Scalable, independent services with eventual consistency"
    },
    {
      "label": "Modular Monolith",
      "description": "Single deployment with well-defined module boundaries"
    },
    {
      "label": "Serverless Functions",
      "description": "Event-driven functions with automatic scaling"
    }
  ]
}
```

---

## TAG Chain Management Rules

### TAG Assignment Standards

**Format**: `<DOMAIN>-<###>` (e.g., `AUTH-003`, `API-042`)

**Domain Prefixes**:
- `AUTH`: Authentication/authorization
- `API`: REST API endpoints
- `UI`: User interface components
- `DB`: Database schema/queries
- `SEC`: Security implementations
- `PERF`: Performance optimizations
- `TEST`: Test implementations
- `DOC`: Documentation updates

### TAG Chain Integrity

**Required Sequence**: SPEC → TEST → CODE → DOC

1. **@SPEC:ID**: Created during planning phase
2. **@TEST:ID**: Created before implementation (TDD RED)
3. **@CODE:ID**: Created during implementation (TDD GREEN)
4. **@DOC:ID**: Created during documentation (SYNC phase)

**Validation Rules**:
- No orphan TAGs (TAG without corresponding code)
- Sequential order must be maintained
- HISTORY section must document TAG lifecycle
- Cross-references must be bidirectional

### TAG Lifecycle Management

**Creation**:
```
Step 1: Assign @SPEC:ID during feature planning
Step 2: Create @TEST:ID before writing implementation
Step 3: Implement @CODE:ID with minimal working code
Step 4: Add @DOC:ID during documentation updates
```

**Validation**:
```
Check 1: All TAGs have corresponding files
Check 2: TAG sequence follows SPEC→TEST→CODE→DOC
Check 3: HISTORY section updated
Check 4: No broken cross-references
```

---

## Context Engineering Best Practices

### JIT (Just-in-Time) Retrieval Strategy

**Core Principle**: Load only what's needed now

**Implementation**:
- Use Explore agent for large searches
- Cache results in task thread for reuse
- Layer context from high-level to specific
- Remove irrelevant context immediately

**Context Layering**:
```
Layer 1: High-level brief
├─ Purpose, stakeholders, success criteria
         ↓
Layer 2: Technical core
├─ Entry points, domain models, utilities
         ↓
Layer 3: Edge cases
├─ Known bugs, constraints, SLAs
```

### Efficient Use of Explore Agent

**When to Use Explore**:
- Call graph analysis for core module changes
- Similar feature searches for implementation reference
- Large-scale dependency mapping
- Complex pattern recognition

**Explore Patterns**:
- Dependency mapping: "Show me all files that import X"
- Feature similarity: "Find similar implementations to Y"
- Impact analysis: "What will break if I change Z"
- Reference finding: "Where is @TAG-123 used"

---

## Workflow Compliance Validation

### 4-Step Workflow Validation

**Step 1 (Intent)**: 
- [ ] Clarity assessed
- [ ] AskUserQuestion used if needed
- [ ] User responses documented

**Step 2 (Planning)**:
- [ ] Tasks decomposed with dependencies
- [ ] Plan Agent consulted for complex tasks
- [ ] TodoWrite initialized

**Step 3 (Execution)**:
- [ ] Tasks tracked with TodoWrite
- [ ] Quality gates applied
- [ ] Blockers handled properly

**Step 4 (Reporting)**:
- [ ] Reports generated only if requested
- [ ] Git commits created always
- [ ] Documentation updated

### Quality Gate Validation

**Before Task Completion**:
```bash
# Validate test coverage
pytest --cov=src --cov-report=term-missing

# Check code quality
ruff check src/
mypy src/

# Security scan
bandit -r src/

# TAG validation
Skill("moai-foundation-tags")
```

**Before Git Commit**:
```bash
# Validate all quality gates
Skill("moai-foundation-trust")

# Check TAG chain integrity
Skill("moai-foundation-tags")

# Ensure proper commit format
# (Built into git commit hooks)
```

---

## Common Violations & Fixes

### Violation 1: Skipping Quality Gates
```
❌ "Task done, moving on"
✅ Complete all TRUST 5 validations
✅ Run mandatory Skill checks
✅ Document gate status
```

### Violation 2: Orphan TAGs
```
❌ @CODE-123 without corresponding @TEST-123
✅ Create @TEST-123 first (TDD RED)
✅ Implement @CODE-123 (TDD GREEN)
✅ Add @DOC-123 (SYNC phase)
```

### Violation 3: Ambiguous Intent
```
❌ "I think they mean X, let's proceed"
✅ Use AskUserQuestion to clarify
✅ Document user's explicit choice
✅ Proceed with clear requirements
```

### Violation 4: Missing Mandatory Skills
```
❌ Direct implementation without validation
✅ Call required Skills before completion
✅ Skill("moai-foundation-trust") for quality
✅ Skill("moai-foundation-tags") for TAGs
```

---

## Integration Patterns

### With moai-alfred-workflow-core
- Quality gates integrated into Step 4
- Mandatory Skills called during execution
- Compliance validation before completion

### With SPEC-First TDD
- TAG chain validation during SYNC phase
- TRUST 5 applied during REFACTOR
- Test coverage validated in GREEN phase

### With Context Optimization
- JIT retrieval for large projects
- Explore agent for complex searches
- Layered context management

---

**End of Skill** | Consolidated from moai-alfred-practices + moai-alfred-rules
