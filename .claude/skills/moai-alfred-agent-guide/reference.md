# moai-alfred-agent-guide: Reference

**Complete 19-Agent Team Reference Documentation**

## Team Structure Overview

### Orchestration Layer (1 Agent)

#### alfred (SuperAgent Orchestrator)
- **Model**: Sonnet (complex orchestration decisions)
- **Role**: Workflow coordinator and team orchestrator
- **Core Responsibilities**:
  - Interpret user intent and clarify ambiguities
  - Route tasks to appropriate specialist agents
  - Manage multi-agent handoffs and dependencies
  - Enforce SPEC-first TDD workflow
  - Track progress with TodoWrite
  - Report outcomes and coordinate completion
- **When to Use**: User-facing interactions, workflow coordination, strategic decision-making
- **Delegation Pattern**: Never executes directly, always delegates to specialists
- **Skills Used**: All 55 Claude Skills (full knowledge base access)

---

### Core Agents (10 Agents)

#### plan-agent (Strategic Planning)
- **Model**: Sonnet (complex reasoning required)
- **Role**: Strategic analysis and execution planning
- **Core Responsibilities**:
  - Decompose user requests into actionable tasks
  - Identify dependencies and execution order
  - Estimate effort and resource requirements
  - Assess risks and propose mitigation strategies
  - Create approval-ready execution plans
  - Initialize TodoWrite task tracking
- **When to Use**: New feature planning, refactoring strategies, migration planning, architecture decisions, risk assessment
- **Typical Output**: Structured plan with task DAG, agent assignments, milestones
- **Skills Used**: moai-alfred-workflow, moai-alfred-todowrite-pattern, moai-foundation-tags

#### tdd-implementer (TDD Development)
- **Model**: Haiku (pattern-based execution)
- **Role**: Test-driven development execution
- **Core Responsibilities**:
  - Execute RED-GREEN-REFACTOR cycle
  - Write failing tests first (RED)
  - Implement minimal passing code (GREEN)
  - Refactor for quality and maintainability
  - Follow TRUST 5 principles
- **When to Use**: Feature implementation, bug fixes with regression tests, code refactoring, API development
- **Process Flow**: Tests → Implementation → Refactoring
- **Skills Used**: moai-lang-python, moai-domain-backend, moai-foundation-tags

#### test-engineer (Testing & QA)
- **Model**: Haiku (rule-based validation)
- **Role**: Comprehensive testing and quality validation
- **Core Responsibilities**:
  - Design test strategies (unit, integration, e2e)
  - Validate test coverage (90%+ target)
  - Create regression test suites
  - Perform edge case testing
  - Generate test reports
  - Maintain test infrastructure
- **When to Use**: Test coverage validation, regression testing, QA gate enforcement, test infrastructure setup
- **Quality Targets**: Unit tests 95%, Integration tests 90%, Security tests 100%
- **Skills Used**: moai-lang-python, moai-domain-testing, moai-alfred-best-practices

#### doc-syncer (Documentation)
- **Model**: Haiku (template-driven content)
- **Role**: Documentation generation and synchronization
- **Core Responsibilities**:
  - Generate API documentation from code
  - Synchronize SPEC → CODE → DOC
  - Update README, CHANGELOG, reference.md
  - Create usage examples
  - Validate documentation accuracy
- **When to Use**: Documentation updates, API reference generation, changelog maintenance, README updates
- **Output Formats**: Markdown, OpenAPI, code examples
- **Skills Used**: moai-alfred-document-management, moai-foundation-tags, moai-alfred-reporting

#### git-manager (Git Operations)
- **Model**: Haiku (deterministic workflows)
- **Role**: Git workflow and version control
- **Core Responsibilities**:
  - Create feature branches (feature/SPEC-XXX)
  - Commit with TDD cycle (RED, GREEN, REFACTOR)
  - Generate conventional commit messages
  - Create pull requests targeting develop
  - Validate GitFlow compliance
  - Tag releases
- **When to Use**: Branch creation, commit operations, pull request creation, release tagging
- **Branch Strategy**: GitFlow with develop branch integration
- **Skills Used**: moai-alfred-agent-guide, moai-foundation-tags

#### qa-validator (Quality Assurance)
- **Model**: Haiku (rule-based validation)
- **Role**: Quality gate enforcement
- **Core Responsibilities**:
  - Enforce TRUST 5 principles (Test First, Readable, Unified, Secured, Trackable)
  - Validate code quality metrics
  - Check test coverage thresholds
  - Verify security best practices
  - Generate quality reports
- **When to Use**: Quality gate checks, pre-merge validation, TRUST 5 enforcement, code review automation
- **Validation Criteria**: Coverage ≥90%, Style guide compliance, Security validations
- **Skills Used**: moai-alfred-best-practices, moai-foundation-tags, moai-alfred-rules

#### project-manager (Project Coordination)
- **Model**: Sonnet (strategic coordination)
- **Role**: Project planning and resource coordination
- **Core Responsibilities**:
  - Coordinate multi-feature roadmaps
  - Manage sprint planning
  - Track milestones and deliverables
  - Allocate agent resources
  - Generate project reports
  - Facilitate stakeholder communication
- **When to Use**: Sprint planning, roadmap coordination, resource allocation, project reporting
- **Planning Horizon**: 2-4 week sprints, quarterly roadmaps
- **Skills Used**: moai-alfred-workflow, moai-alfred-todowrite-pattern, moai-foundation-tags

#### debug-helper (Troubleshooting)
- **Model**: Sonnet (complex investigation)
- **Role**: Debugging and root cause analysis
- **Core Responsibilities**:
  - Investigate bugs and errors
  - Perform root cause analysis
  - Profile application performance
  - Analyze logs and stack traces
  - Propose fix strategies
  - Coordinate with domain experts
- **When to Use**: Bug investigation, performance issues, production errors, unexpected behavior
- **Debug Process**: Investigation → Hypothesis → Validation → Fix Strategy
- **Skills Used**: moai-lang-python, moai-domain-backend, moai-alfred-dev-guide

#### trust-checker (TRUST 5 Validation)
- **Model**: Haiku (rule-based validation)
- **Role**: TRUST 5 principle enforcement
- **Core Responsibilities**:
  - Validate Test First principle (code has tests)
  - Check Readability (style guides, documentation)
  - Verify Unified patterns (consistency)
  - Audit Security posture (OWASP compliance)
  - Generate TRUST 5 compliance reports
- **When to Use**: Pre-merge validation, quality gate enforcement, TRUST 5 audits, compliance reporting
- **Validation Framework**: Test First, Readable, Unified, Secured, Trackable
- **Skills Used**: moai-alfred-best-practices, moai-alfred-rules, moai-foundation-tags

#### tag-agent (TAG System Management)
- **Model**: Haiku (deterministic identifier management)
- **Role**: TAG system integrity and traceability management
- **Core Responsibilities**:
  - Maintain TAG chain integrity (SPEC → TEST → CODE → DOC)
  - Detect orphaned TAGs and broken chains
  - Generate traceability reports
  - Repair broken TAG chains
  - Validate TAG format compliance
- **When to Use**: Traceability validation, TAG chain verification, orphan detection and repair, TAG system health checks
- **Chain Sequence**: SPEC → TEST → CODE → DOC (bidirectional links)
- **Skills Used**: moai-foundation-tags, moai-alfred-workflow

---

### Domain Specialists (6 Agents)

#### backend-expert (Backend Development)
- **Model**: Sonnet (architecture design) or Haiku (implementation)
- **Role**: Backend architecture and implementation specialist
- **Core Responsibilities**:
  - Design backend APIs and services
  - Implement business logic
  - Database schema design
  - Performance optimization
  - Integration with external services
  - Backend security patterns
- **When to Use**: API design, backend architecture, service implementation, database schema design
- **Architecture Patterns**: REST APIs, GraphQL, microservices, event-driven systems
- **Skills Used**: moai-domain-backend, moai-domain-web-api, moai-lang-python

#### frontend-expert (Frontend Development)
- **Model**: Sonnet (architecture) or Haiku (component implementation)
- **Role**: Frontend architecture and UI implementation specialist
- **Core Responsibilities**:
  - Design frontend architecture
  - Implement UI components
  - State management patterns
  - Frontend performance optimization
  - Accessibility compliance
  - Frontend testing strategies
- **When to Use**: UI component design, frontend architecture, state management, client-side routing
- **Technology Stack**: React, Vue, Angular, TypeScript, Webpack, Vite
- **Skills Used**: moai-domain-frontend, moai-lang-javascript, moai-domain-web-api

#### database-expert (Database Design)
- **Model**: Sonnet (schema design) or Haiku (query optimization)
- **Role**: Database architecture and optimization specialist
- **Core Responsibilities**:
  - Design database schemas
  - Optimize queries and indexes
  - Data migration strategies
  - Database performance tuning
  - Data integrity and constraints
  - Backup and recovery planning
- **When to Use**: Database schema design, query optimization, data migrations, performance tuning
- **Database Systems**: PostgreSQL, MySQL, MongoDB, Redis, Elasticsearch
- **Skills Used**: moai-domain-database, moai-lang-python, moai-domain-backend

#### security-expert (Security & Authentication)
- **Model**: Sonnet (complex security analysis)
- **Role**: Security architecture and threat mitigation specialist
- **Core Responsibilities**:
  - Security architecture design
  - Threat modeling and risk assessment
  - Authentication/authorization strategies
  - Vulnerability assessment
  - Compliance validation (OWASP, GDPR)
  - Security testing and auditing
- **When to Use**: Security design, threat modeling, vulnerability assessment, compliance validation
- **Security Frameworks**: OWASP Top 10, NIST, ISO 27001, SOC 2
- **Skills Used**: moai-domain-security, moai-alfred-best-practices, moai-domain-backend

#### devops-expert (Infrastructure & DevOps)
- **Model**: Sonnet (infrastructure design) or Haiku (configuration)
- **Role**: Infrastructure architecture and deployment specialist
- **Core Responsibilities**:
  - Infrastructure as Code (IaC) design
  - CI/CD pipeline configuration
  - Container orchestration (Docker, Kubernetes)
  - Cloud platform management (AWS, Azure, GCP)
  - Monitoring and observability
  - Deployment strategies (blue-green, canary)
- **When to Use**: Infrastructure design, CI/CD setup, deployment automation, monitoring configuration
- **Technologies**: Docker, Kubernetes, Terraform, GitHub Actions, AWS CloudFormation
- **Skills Used**: moai-domain-devops, moai-lang-shell, moai-alfred-workflow

#### api-designer (API Architecture)
- **Model**: Sonnet (API architecture) or Haiku (OpenAPI generation)
- **Role**: API design and contract specialist
- **Core Responsibilities**:
  - RESTful API design
  - GraphQL schema design
  - API versioning strategies
  - OpenAPI/Swagger documentation
  - API contract testing
  - Backward compatibility management
- **When to Use**: API design, contract definition, API versioning, OpenAPI documentation
- **API Standards**: OpenAPI 3.0, REST best practices, GraphQL specifications
- **Skills Used**: moai-domain-web-api, moai-domain-backend, moai-alfred-document-management

---

### Built-in Agents (2 Agents)

#### Plan (Built-in Strategic Analysis Agent)
- **Model**: Sonnet (Claude Code default)
- **Role**: Deep strategic analysis and planning (Claude Code built-in)
- **Core Responsibilities**:
  - Complex strategic analysis
  - Multi-constraint optimization
  - Comparative evaluation
  - Recommendation generation
  - Risk assessment
- **When to Use**: Strategic decisions, technology selection, architecture comparisons, risk analysis
- **Access**: All Claude Code built-in capabilities
- **Integration**: Used directly via Task() with subagent_type="Plan"

#### WebSearch (Built-in Research Agent)
- **Model**: Sonnet (Claude Code default)
- **Role**: Web research and information validation (Claude Code built-in)
- **Core Responsibilities**:
  - Latest documentation lookup
  - Best practice research
  - Technology comparison
  - Version compatibility checks
  - Security advisory lookup
- **When to Use**: Latest framework versions, security advisories, best practice research, documentation lookup
- **Access**: WebSearch tool directly (not via Task())
- **Integration**: Results inform other agent decisions

---

## Agent Capabilities Matrix

| Agent | Primary Model | Complex Reasoning | Pattern Execution | Security Analysis | Architecture Design | Testing | Documentation |
|-------|---------------|------------------|-------------------|------------------|--------------------|---------|---------------|
| alfred | Sonnet | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| plan-agent | Sonnet | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| tdd-implementer | Haiku | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| test-engineer | Haiku | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| doc-syncer | Haiku | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ |
| git-manager | Haiku | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| qa-validator | Haiku | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| project-manager | Sonnet | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| debug-helper | Sonnet | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ |
| trust-checker | Haiku | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ |
| tag-agent | Haiku | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| backend-expert | Sonnet/Haiku | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| frontend-expert | Sonnet/Haiku | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ |
| database-expert | Sonnet/Haiku | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ |
| security-expert | Sonnet | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| devops-expert | Sonnet/Haiku | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| api-designer | Sonnet/Haiku | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Plan | Sonnet | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| WebSearch | Sonnet | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

## Performance Characteristics

### Model Selection Guidelines

**Sonnet Advantages**:
- Deep reasoning and creative solutions
- Context synthesis across multiple domains
- Complex decision-making with trade-offs
- Strategic planning capabilities
- Security threat modeling

**Haiku Advantages**:
- Fast execution (0.5-1 second vs 2-5 seconds)
- Lower cost (80% reduction vs Sonnet)
- Pattern recognition and consistency
- Rule-based validation
- Standardized workflows

### Cost Optimization

**Typical Feature Breakdown**:
- Planning: Sonnet (1 agent, 5 min, ~$3)
- Implementation: Haiku (3 agents, 10 min, ~$1.50)
- Testing: Haiku (2 agents, 5 min, ~$0.50)
- Documentation: Haiku (1 agent, 2 min, ~$0.25)
- Validation: Haiku (2 agents, 3 min, ~$0.25)

**Total**: ~$5.50 per feature (vs ~$15+ for all-Sonnet approach)

### Performance Optimization

**Parallel Execution**: Independent agents can run simultaneously
**Context Budgeting**: Load only relevant skills per agent type
**Model Overrides**: Force specific models when needed
**Handoff Validation**: Ensure complete context transfer

---

## Error Handling and Recovery

### Common Failure Modes

1. **Agent Timeouts**: Increase timeout for complex tasks
2. **Context Overflow**: Reduce context scope, use Skill() loading
3. **Handoff Failures**: Validate required fields before delegation
4. **Model Mismatches**: Override model when task complexity differs from defaults

### Recovery Strategies

1. **Retry with Different Model**: Haiku → Sonnet for complex failures
2. **Debug-helper Escalation**: Complex failures to debug-helper
3. **Plan Agent Re-routing**: Strategic issues back to plan-agent
4. **Manual Intervention**: Alfred coordinates human intervention

---

*Generated with MoAI-ADK Skill Factory v5.0*  
*Complete reference documentation extracted from original 2226-line guide*  
*Last Updated: 2025-11-13*
