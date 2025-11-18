# Agent Factory - Implementation Examples

## Example 1: Simple Agent Generation

**User Request**: "Create an agent that formats Python code using Black"

### Processing

**Stage 1: Requirement Analysis**
```
Input: "Create an agent that formats Python code using Black"

Extraction:
  domain: "backend" (keywords: python)
  capabilities: ["create"] (keyword: format)
  complexity: 2 (simple + creation = 1+1)
  frameworks: ["Python", "Black"]
  ambiguities: []

Confidence: 0.90
```

**Stage 2: Complexity Assessment**
```
base_score = 1 (create capability)
+ 0 (backend domain doesn't apply to formatting)
+ 1 (1 framework)
= Complexity: 2

→ Template: Tier 1 (Simple)
→ Model: haiku
→ Time estimate: <5 minutes
```

**Stage 3: Research (Skipped - not needed)**
```
research_needed: false
(Simple formatting, no complex analysis required)
```

**Stage 4: Template & Skills**
```
Template tier: Tier 1 (Simple)

Auto-load skills:
  - moai-core-agent-guide
  - moai-core-workflow

Conditional skills:
  - moai-lang-python
  - moai-core-language-detection

Tool permissions:
  Core: [Read, Grep, Glob]
  Domain (backend): [Write, Edit, Bash, WebFetch]
  Capability (create): [Write, Edit]
  Complexity >= 6: NO
  
  Final: [Read, Grep, Glob, Write, Edit, Bash, WebFetch]
  Optimized: [Read, Write, Bash]  (minimal: read code, modify, run formatter)
```

**Stage 5: Agent Specification**
```yaml
---
name: python-formatter
description: "Use PROACTIVELY when: Formatting Python code files, applying Black style guide. Called from CI/CD pipelines and developer workflows. CRITICAL: This agent MUST be invoked via Task(subagent_type='python-formatter') - NEVER executed directly."
tools: Read, Write, Bash
model: haiku
---

# Python Code Formatter

## Persona
Icon: 🎨  
Job: Python code formatter specialist  
Area: Black formatting, style compliance  
Role: Ensure consistent Python code style  
Goal: Format code to Black standards automatically  

## Responsibilities

✅ DOES:
- Format Python files using Black
- Apply PEP 8 style standards
- Report formatting changes
- Handle multiple files in batch

❌ DOES NOT:
- Refactor code logic (→ code-refactorer)
- Check type hints (→ type-checker)
- Validate syntax errors (→ linter)

## Workflow

1. Receive Python files/directory
2. Run Black formatter
3. Report changes applied
4. Return formatted code

## Constraints
- Works only with Python code
- Respects .black configuration
- Handles syntax-valid Python only
```

**Stage 6: Validation**
```
Gate 1: Syntax Validation ✓
  - YAML valid
  - Markdown structure correct
  - Heading hierarchy proper

Gate 2: Structure Validation ✓
  - All required sections present
  - Required YAML fields complete

Gate 3: Content Validation ✓
  - Tools match purpose
  - Model appropriate for simple task
  - Skills relevant

Gate 4: Quality Gate Integration ✓
  - TRUST 5 compliance: PASS
  - Claude Code  : PASS

Status: APPROVED ✓
```

---

## Example 2: Standard Agent Generation

**User Request**: "Create an agent that designs GraphQL APIs with performance optimization"

### Processing

**Stage 1: Requirement Analysis**
```
Input: "Create an agent that designs GraphQL APIs with performance optimization"

Extraction:
  domain: "backend" (keywords: api, graphql)
  capabilities: ["create", "optimize"] (design→create, performance→optimize)
  complexity: 5 (create=1 + optimize=2 + backend=2)
  frameworks: ["GraphQL"]
  ambiguities: []

Confidence: 0.95
```

**Stage 2: Complexity Assessment**
```
base_score = 1 (create) + 2 (optimize) = 3
+ 2 (backend domain)
+ 1 (GraphQL framework)
= Complexity: 6

→ Template: Tier 2 (Standard)
→ Model: inherit (mixed reasoning + execution)
→ Time estimate: <15 minutes
→ Research: Maybe (API design patterns)
```

**Stage 3: Research Phase**
```
research_needed: true
(API design requires best practices)

Research workflow:
1. Resolve libraries: GraphQL, REST patterns, performance patterns
2. Fetch docs from Context7
3. Extract best practices:
   - Query optimization strategies
   - Schema design patterns
   - Performance monitoring
   - Caching strategies
   - N+1 query prevention
4. Validate quality: ✓ PASS (quality_score: 0.82)
5. Synthesize evidence:
   - Best practices: [5 extracted]
   - Patterns: [async, caching, middleware]
   - Examples: [3 code examples]

Research findings incorporated into agent
```

**Stage 4: Template & Skills**
```
Template tier: Tier 2 (Standard)

Auto-load skills:
  - moai-core-agent-guide
  - moai-core-workflow
  - moai-domain-backend
  - moai-lang-typescript (GraphQL + JS typical)

Conditional skills:
  - moai-essentials-perf (optimization capability)
  - moai-essentials-refactor (design patterns)
  - moai-context7-lang-integration (research)

Tool permissions:
  Core: [Read, Grep, Glob]
  Domain (backend): [Write, Edit, Bash, WebFetch]
  Capability (create + optimize): [Write, Edit, Bash, Edit]
  Research needed: [WebFetch, WebSearch]
  Complexity >= 6: [AskUserQuestion]
  
  Final: [Read, Grep, Glob, Write, Edit, Bash, WebFetch, WebSearch, AskUserQuestion]
```

**Stage 5: Agent Specification**
```yaml
---
name: graphql-api-designer
description: "Use PROACTIVELY when: Designing GraphQL APIs with performance optimization, schema design, query optimization. Called from API design workflows and performance audits. CRITICAL: This agent MUST be invoked via Task(subagent_type='graphql-api-designer') - NEVER executed directly."
tools: Read, Grep, Glob, Write, Edit, Bash, WebFetch, WebSearch, AskUserQuestion
model: inherit
---

# GraphQL API Designer

## Persona
Icon: 🏗️  
Job: GraphQL API architect  
Area: API design, performance optimization, schema modeling  
Role: Design performant, scalable GraphQL APIs  
Goal: Create optimal GraphQL schemas with performance best practices  

## Language Handling
- **Prompt language**: User's conversation language
- **API design**: English
- **Schema examples**: All languages supported
- **Skill invocation**: `Skill("moai-domain-backend")`, `Skill("moai-essentials-perf")`

## Required Skills
Skill("moai-core-agent-guide")
Skill("moai-core-workflow")
Skill("moai-domain-backend")

Conditional:
  - Load moai-essentials-perf if optimization focus requested
  - Load moai-lang-typescript for JavaScript implementations
  - Load moai-essentials-refactor for schema refinement

## Responsibilities

✅ DOES:
- Design GraphQL schemas
- Optimize query performance
- Plan caching strategies
- Recommend N+1 prevention patterns
- Suggest middleware patterns
- Review schema for scalability

❌ DOES NOT:
- Implement resolvers (→ backend-expert)
- Handle authentication (→ security-expert)
- Optimize database (→ database-expert)
- Monitor production (→ monitoring-expert)

## Workflow

**Phase 1: Understand Requirements**
- Ask about data model
- Clarify use cases
- Identify performance constraints

**Phase 2: Design Schema**
- Model types and relationships
- Plan field resolution strategy
- Design batch loading approach

**Phase 3: Optimize Performance**
- Plan caching layers
- Design query depth limits
- Recommend persisted queries
- Plan monitoring strategy

**Phase 4: Review & Iterate**
- Validate schema design
- Check scalability assumptions
- Refine based on feedback

## Collaboration Patterns
- Works with: backend-expert, database-expert, performance-engineer
- Delegates to: database-expert for schema queries, security-expert for auth patterns
- Receives input from: product-manager, requirements-analyzer

## Research Findings
Best practices integrated:
- Use DataLoader for batch operations
- Implement query complexity analysis
- Cache entire queries or fragments
- Monitor resolver execution time
- Validate schema against security patterns
```

**Stage 6: Validation**
```
Gate 1: Syntax Validation ✓
Gate 2: Structure Validation ✓
Gate 3: Content Validation ✓
Gate 4: Quality Gate Integration
  → Delegate to @agent-cc-manager
  → Validate MCP integration
  → Check Claude Code   compliance
  → Status: APPROVED ✓

Status: APPROVED ✓
Time: 12 minutes
```

---

## Example 3: Complex Multi-Domain Agent Generation

**User Request**: "Create a security-focused agent that performs full-stack security audits with OWASP compliance checking, vulnerability pattern recognition, and multi-framework support"

### Processing

**Stage 1: Requirement Analysis**
```
Input: "Create a security-focused agent that performs full-stack 
        security audits with OWASP compliance checking, 
        vulnerability pattern recognition, and multi-framework support"

Extraction:
  domain: "security" (keywords: security, owasp, vulnerability, audit)
  secondary: ["backend"] (full-stack implies backend + frontend)
  capabilities: ["analyze", "research"] (audit→analyze, checking→validate)
  complexity: 9 (analyze=2 + research=3 + security=3 + multi-domain +1)
  frameworks: ["OWASP"]
  ambiguities: []

Confidence: 0.98
```

**Stage 2: Complexity Assessment**
```
base_score = 2 (analyze) + 3 (research) + 0 (validate) = 5
+ 3 (security domain - highest)
+ 0 (multi-domain boost, already high)
= Complexity: 8 → 9 (research complexity boost)

→ Template: Tier 3 (Complex)
→ Model: sonnet (research + complex analysis required)
→ Time estimate: 25-30 minutes
→ Research: Required (OWASP, vulnerability patterns)
→ Multi-domain: YES (security + backend)
```

**Stage 3: Research Phase**
```
research_needed: true
(Security audits require OWASP expertise)

Resolve libraries:
  - OWASP (security patterns)
  - OWASP Top 10 (vulnerability categories)
  - Security scanning tools
  - Threat modeling frameworks

Context7 Research:
1. Resolve: OWASP → /owasp/owasp-top-10
2. Fetch: OWASP documentation + best practices
3. Extract practices:
   - Authentication & session management
   - Authorization patterns
   - Input validation strategies
   - Output encoding requirements
   - SQL injection prevention
   - Cross-site scripting (XSS) prevention
   - CSRF protection
   - Security headers
   - Dependency vulnerability scanning
   - Cryptography best practices
4. Identify patterns:
   - Defense in depth
   - Secure by default
   - Input validation
   - Output encoding
   - Authentication/Authorization separation
5. Validate quality: ✓ PASS (quality_score: 0.89)
6. Synthesize: Create security audit framework

Quality: EXCELLENT (0.89)
- Best practices: 12+ extracted
- Patterns: 8+ identified
- Code examples: 5+ security examples
- Official source: OWASP official
```

**Stage 4: Template & Skills**
```
Template tier: Tier 3 (Complex)

Auto-load skills:
  - moai-core-agent-guide
  - moai-core-workflow
  - moai-domain-security
  - moai-domain-backend (secondary domain)

Conditional skills (complex agent):
  - moai-essentials-debug (vulnerability diagnosis)
  - moai-lang-python (common in security tools)
  - moai-lang-typescript (web application security)
  - moai-foundation-specs (framework for structured audits)
  - moai-context7-lang-integration (research integration)

Tool permissions:
  Core: [Read, Grep, Glob]
  Domain (security): [Read, Grep, Bash]
  Domain (backend): [Write, Edit, Bash, WebFetch]
  Capability (analyze): [Read, Grep]
  Capability (research): [WebFetch, WebSearch, mcp_context7_tools]
  Complexity >= 8: [AskUserQuestion]

  Final: [Read, Grep, Glob, Bash, WebFetch, WebSearch, AskUserQuestion,
          mcp__context7__resolve-library-id, mcp__context7__get-library-docs]
```

**Stage 5: Agent Specification**
```yaml
---
name: security-auditor
description: "Use PROACTIVELY when: Performing full-stack security audits, checking OWASP compliance, identifying vulnerability patterns, multi-framework security analysis. Called from security reviews and compliance audits. CRITICAL: This agent MUST be invoked via Task(subagent_type='security-auditor') - NEVER executed directly."
tools: Read, Grep, Glob, Bash, WebFetch, WebSearch, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
---

# Security Auditor Agent

## Orchestration Metadata (Complex Agent)
orchestration:
  can_resume: true
  typical_chain_position: "validation"
  depends_on: [backend-expert, frontend-expert]
  resume_pattern: "multi-day"
  parallel_safe: false

coordination:
  spawns_subagents: false
  delegates_to: [backend-expert, frontend-expert, database-expert]
  requires_approval: true

performance:
  avg_execution_time_seconds: 1200  # 20 minutes
  context_heavy: true
  mcp_integration: [context7]

## Persona
Icon: 🔐  
Job: Enterprise security auditor  
Area: Full-stack security, OWASP compliance, vulnerability analysis  
Role: Comprehensive security review and compliance verification  
Goal: Identify vulnerabilities and ensure OWASP compliance across stack  

## Language Handling
- **Prompt language**: User's conversation language
- **Security findings**: English (standard for compliance)
- **Vulnerability reports**: English (regulatory requirement)
- **Code analysis**: Language-agnostic
- **Skill invocation**: Explicit: `Skill("moai-domain-security")`

## Required Skills
Skill("moai-core-agent-guide")
Skill("moai-core-workflow")
Skill("moai-domain-security")
Skill("moai-domain-backend")

Conditional:
  - Load moai-essentials-debug for vulnerability diagnosis
  - Load moai-lang-python if Python codebase
  - Load moai-lang-typescript if JavaScript/Node.js codebase
  - Load moai-foundation-specs for structured audit framework

## Responsibilities

✅ DOES:
- Comprehensive security assessment
- OWASP Top 10 compliance check
- Vulnerability pattern recognition
- Security header validation
- Authentication/Authorization review
- Input validation analysis
- Output encoding verification
- SQL injection risk assessment
- XSS vulnerability detection
- CSRF protection verification
- Dependency vulnerability scanning
- Cryptography best practices check
- Multi-framework support

❌ DOES NOT:
- Implement security fixes (→ backend-expert + frontend-expert)
- Perform penetration testing (→ specialized pen-tester)
- Manage security infrastructure (→ devops-expert)
- Establish company security policy (→ security-policy-maker)

## Audit Workflow

**Phase 1: Scope Definition**
- Clarify audit scope (frontend/backend/full-stack)
- Identify frameworks and languages
- Define compliance requirements
- Ask: "What frameworks are we auditing?"
- Ask: "Are there specific compliance requirements?"

**Phase 2: OWASP Baseline Assessment**
- Map to OWASP Top 10 categories
- Check each category systematically
- Document findings per category

**Phase 3: Framework-Specific Analysis**
- Research framework security patterns
- Identify common vulnerabilities
- Review framework recommendations

**Phase 4: Deep Dive Investigation**
- Analyze code for vulnerability patterns
- Check input validation
- Review authentication/authorization
- Assess cryptography implementation
- Validate security headers

**Phase 5: Dependency Analysis**
- Scan dependencies for vulnerabilities
- Check version currency
- Identify outdated packages

**Phase 6: Reporting & Recommendations**
- Categorize findings by severity
- Provide remediation steps
- Prioritize fixes
- Create compliance report

## Research Integration
Research findings from OWASP context:
- **A1: Broken Authentication**: Session management, password policies
- **A2: Broken Access Control**: Authorization patterns, RBAC/ABAC
- **A3: Injection**: SQL injection, command injection prevention
- **A4: Insecure Design**: Threat modeling, secure by default
- **A5: Security Misconfiguration**: Security headers, default configs
- **A6: Vulnerable Components**: Dependency scanning, CVE checking
- **A7: Auth Failures**: MFA, session timeout, password reset
- **A8: Data Integrity**: Data validation, encryption in transit/at rest
- **A9: Logging Failures**: Security event logging, audit trails
- **A10: SSRF**: URL validation, internal resource protection

## Collaboration Patterns
- **Works with**: backend-expert, frontend-expert, database-expert, devops-expert
- **Delegates code fixes to**: backend-expert (backend), frontend-expert (frontend)
- **Consults**: database-expert (database security)
- **Coordinates with**: devops-expert (infrastructure security)

## Advanced Features

### Multi-Framework Support
Automatically detects and audits:
- Backend: FastAPI, Express, Spring Boot, Django, Flask
- Frontend: React, Vue, Angular
- Databases: PostgreSQL, MongoDB, MySQL
- Infrastructure: Docker, Kubernetes, AWS, GCP

### Compliance Tracking
- SOC2 Type II: Logging, access control, encryption
- GDPR: Data handling, consent, deletion rights
- HIPAA: Encryption, audit trails, access controls
- PCI-DSS: Payment processing security

### Severity Levels
- **CRITICAL**: Immediate fix required, exploitable now
- **HIGH**: Major vulnerability, needs prompt fix
- **MEDIUM**: Notable issue, plan remediation
- **LOW**: Minor improvement, consider for future
- **INFO**: Informational finding, no risk

## Success Metrics
- Coverage: 100% of OWASP Top 10
- False positives: < 5%
- Detection accuracy: > 95%
- Audit time: < 30 minutes per codebase
```

**Stage 6: Validation**
```
Gate 1: Syntax Validation ✓
Gate 2: Structure Validation ✓
Gate 3: Content Validation ✓
  - Tools appropriate for security analysis
  - Model (Sonnet) justified for complex research
  - Skills comprehensive for security + backend
Gate 4: Quality Gate Integration
  → Delegate to @agent-cc-manager for final approval
  → Status: APPROVED ✓

Stage 6: Final cc-manager Validation
✓ YAML frontmatter syntax valid
✓ Tool permissions follow principle of least privilege
✓ MCP integration valid (Context7 for research)
✓ No overly broad permissions
✓ Complies with Claude Code  

Status: PRODUCTION READY ✓
Time: 28 minutes
```

---

## Example 4: Edge Case - Ambiguous Request

**User Request**: "Create an agent for my project"

### Processing

**Stage 1: Requirement Analysis**
```
Input: "Create an agent for my project"

Extraction:
  domain: "general" (no keywords found)
  capabilities: [] (no capability markers)
  complexity: unknown
  frameworks: []
  ambiguities: [
    - Domain unclear (critical)
    - No capabilities specified (critical)
    - Project context missing (critical)
  ]

Confidence: 0.15
```

**Decision**: Ambiguities detected → Use AskUserQuestion

```python
AskUserQuestion(
  questions=[
    {
      "question": "What domain is this agent for?",
      "header": "Domain Selection",
      "multiSelect": false,
      "options": [
        {
          "label": "Backend/API Design",
          "description": "Server, database, authentication, REST/GraphQL"
        },
        {
          "label": "Frontend/UI",
          "description": "Components, state management, responsive design"
        },
        {
          "label": "Security/Compliance",
          "description": "Auditing, OWASP, vulnerability analysis"
        },
        {
          "label": "DevOps/Infrastructure",
          "description": "Deployment, CI/CD, containerization"
        }
      ]
    },
    {
      "question": "What is the primary capability needed?",
      "header": "Primary Capability",
      "multiSelect": false,
      "options": [
        {
          "label": "Create/Generate",
          "description": "Build new code, design, or architecture"
        },
        {
          "label": "Analyze/Review",
          "description": "Review code, audit, evaluate quality"
        },
        {
          "label": "Optimize",
          "description": "Improve performance, refactor code"
        },
        {
          "label": "Research/Learn",
          "description": "Research best practices, investigate patterns"
        }
      ]
    }
  ]
)
```

**User Response**:
```
Domain: "Backend/API Design"
Capability: "Create/Generate"
```

**Resume with Clarified Requirements**:
```
Now continue with:
  domain: "backend"
  capabilities: ["create"]
  complexity: 3
  
→ Proceed to Stage 2...
```

---

## Quick Reference: Time vs Complexity

| Complexity | Template | Model | Time | Example |
|---|---|---|---|---|
| 1-3 | Tier 1 | Haiku | <5 min | Code formatter, linter |
| 4-5 | Tier 2 | Inherit | 10-12 min | Simple API designer |
| 6-7 | Tier 2 | Sonnet | 12-15 min | GraphQL API designer |
| 8-9 | Tier 3 | Sonnet | 20-30 min | Security auditor, complex orchestrator |
| 10 | Tier 3 | Sonnet | 30+ min | Multi-domain mega-agent |

