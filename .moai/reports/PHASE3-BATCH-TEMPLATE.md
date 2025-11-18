# PHASE 3: Domain & Core Skills Batch Processing Template

**Batch Update Guide for remaining 33 Domain & Core Skills**

Updated: 2025-11-19 | Target: All 37 Domain + Core Skills to v4.0.0

---

## Overview

This template provides the pattern for updating all remaining 33 Domain and Core Skills to enterprise v4.0.0 standard. The 4 representative Skills are already updated:

✅ **Updated (Representative Sample - Domain)**:
- moai-domain-backend (FastAPI 0.121.0, async patterns, PostgreSQL 17)
- moai-domain-frontend (React 19.2, Next.js 16, component architecture)

✅ **Updated (Representative Sample - Core)**:
- moai-core-workflow (agent delegation, orchestration, Context7 integration)
- moai-core-context-budget (token optimization, memory management)

⏳ **Remaining (9 Domain Skills)**:
- moai-domain-cli-tool, moai-domain-cloud, moai-domain-data-science
- moai-domain-database, moai-domain-devops, moai-domain-ml
- moai-domain-ml-ops, moai-domain-mobile-app, moai-domain-monitoring

⏳ **Remaining (24 Core Skills)**:
- moai-core-agent-factory, moai-core-agent-guide, moai-core-ask-user-questions
- moai-core-clone-pattern, moai-core-code-reviewer, moai-core-config-schema
- moai-core-dev-guide, moai-core-env-security, moai-core-expertise-detection
- moai-core-feedback-templates, moai-core-issue-labels, moai-core-language-detection
- moai-core-personas, moai-core-practices, moai-core-proactive-suggestions
- moai-core-rules, moai-core-session-state, moai-core-spec-authoring
- moai-core-todowrite-pattern, moai-core-testing, moai-core-web-api

---

## Domain Skills Template Structure (v4.0.0)

### Frontmatter
```yaml
---
name: "moai-domain-{domain}"
version: "4.0.0"
tier: "Domain"
created: 2025-11-11
updated: 2025-11-19
status: "stable"
description: |
  Enterprise {Domain} architecture with {key_tech_1}, {key_tech_2}.
  Activates for {use_case_1}, {use_case_2}, {use_case_3}.
primary-agent: "{domain}-expert"
secondary-agents: ["backend-expert", "security-expert"]
keywords: ["{domain}", "{key_tech_1}", "{pattern_1}"]
allowed-tools:
  - Read
  - Bash
  - WebFetch
  - mcp__context7__get-library-docs
---
```

### Content: Progressive Disclosure

#### Level 1: Technology Stack Reference
- Current versions matrix (November 2025)
- Primary frameworks and tools
- Auto-triggers and activation keywords

#### Level 2: Core Patterns (5-7 production patterns)
- Architecture pattern with code
- Security pattern with best practices
- Performance pattern with benchmarks
- Testing pattern with coverage
- Deployment pattern with YAML examples

#### Level 3: Advanced Topics (3-5 deep dives)
- Advanced optimization techniques
- Enterprise considerations
- Multi-team coordination
- Disaster recovery
- Cost optimization

#### Case Studies (2-3 real-world examples)
- Architecture diagram description
- Key decisions and trade-offs
- Results and metrics
- Lessons learned

---

## Domain Skills: Specific Guidance

### moai-domain-backend
✅ Already updated with:
- FastAPI 0.121.0 async patterns
- SQLAlchemy 2.0 ORM patterns
- PostgreSQL 17 optimization
- Kubernetes deployment
- gRPC microservices

### moai-domain-frontend
✅ Already updated with:
- React 19.2 server/client components
- Next.js 16 routing and optimization
- Component architecture patterns
- State management with Context7
- Performance metrics and optimization

### moai-domain-cli-tool
Update guidance:
- Modern CLI frameworks (Click, Cobra, oclif)
- Async CLI patterns
- Progress indicators and spinners
- Configuration management
- Testing CLI applications
- Distribution and packaging

### moai-domain-cloud
Update guidance:
- Cloud providers: AWS, GCP, Azure (latest SDKs)
- Infrastructure as Code (Terraform, CDK)
- Serverless patterns
- Container orchestration
- Cost optimization strategies
- Multi-cloud strategies

### moai-domain-database
Update guidance:
- PostgreSQL 17, MySQL 8.4, MongoDB 8
- Schema design patterns
- Query optimization and indexing
- Connection pooling
- Backup and recovery
- Replication and clustering

### moai-domain-devops
Update guidance:
- CI/CD with GitHub Actions, GitLab CI
- Container security scanning
- Infrastructure monitoring
- Disaster recovery automation
- Infrastructure as Code best practices
- Kubernetes deployment patterns

### moai-domain-data-science
Update guidance:
- Python 3.13, pandas, scikit-learn
- Jupyter notebooks best practices
- Data validation and cleaning
- ML pipeline automation
- Visualization (matplotlib, plotly, seaborn)
- Reproducibility and versioning

### moai-domain-ml, moai-domain-ml-ops
Update guidance:
- PyTorch 2.x, TensorFlow 2.18, JAX
- Training pipelines and optimization
- Model evaluation and validation
- A/B testing strategies
- Model versioning (MLflow)
- Production deployment of ML models

### moai-domain-mobile-app
Update guidance:
- iOS: Swift 6.0, SwiftUI
- Android: Kotlin 2.1, Jetpack
- Cross-platform: Flutter 3.27, React Native
- Performance optimization
- Offline-first architecture
- Push notifications and background tasks

### moai-domain-monitoring
Update guidance:
- Prometheus 3.0 time-series data
- Grafana 11 dashboards
- OpenTelemetry 1.28 instrumentation
- Distributed tracing with Jaeger
- Alert management strategies
- SLO/SLI definition and tracking

### moai-domain-testing
Update guidance:
- Unit testing frameworks per language
- Integration testing patterns
- End-to-end testing with Playwright/Cypress
- Performance testing and load testing
- Security testing (OWASP, penetration)
- Test automation strategies

### moai-domain-web-api
Update guidance:
- REST API design (OpenAPI 3.1)
- GraphQL patterns (Apollo Server)
- gRPC services and streaming
- API versioning strategies
- Rate limiting and throttling
- API security and authentication

---

## Core Skills Template Structure (v4.0.0)

Core Skills tend to be more Alfred-specific and orchestration-focused.

### Frontmatter
```yaml
---
name: "moai-core-{capability}"
version: "4.0.0"
tier: "Alfred"
status: "production"
description: |
  Enterprise {capability} specialist for MoAI-ADK.
  Handles {responsibility_1}, {responsibility_2}, {responsibility_3}.
primary-agent: "alfred"
secondary-agents: ["plan-agent", "session-manager"]
keywords: ["{capability}", "orchestration", "automation"]
allowed-tools: ["Read", "Write", "Edit", "Bash", "WebFetch"]
---
```

### Content Structure

#### Overview
- Clear statement of responsibility
- Key capabilities matrix
- When to use (automatic vs. manual)

#### Core Patterns (3-5 essential patterns)
- Code examples showing integration with Alfred
- Context7 MCP usage where applicable
- Performance considerations
- Error handling

#### Best Practices
- 10+ specific practices for this capability
- Trade-offs and when to deviate
- Common pitfalls

#### Advanced Topics
- Deep integration scenarios
- Multi-agent orchestration
- Performance optimization
- Monitoring and debugging

#### References
- Cross-links to other Core Skills
- Links to official MoAI-ADK documentation
- External resources

---

## Core Skills Specific Guidance

### Alfred Core Skills (Top Priority)

**moai-core-workflow**
✅ Already updated with:
- Multi-agent orchestration patterns
- Context7 integration examples
- Dependency management
- Error recovery strategies
- Performance monitoring

**moai-core-context-budget**
✅ Already updated with:
- 200K token optimization strategies
- Aggressive clearing patterns
- Memory file management (<500 lines)
- MCP server efficiency
- Phase-based token planning

**moai-core-agent-factory**
Provides:
- Agent creation patterns
- Domain-specific agent design
- Agent communication protocols
- Agent lifecycle management
- Scaling considerations

**moai-core-agent-guide**
Provides:
- Overview of all 19+ Alfred agents
- When to use each agent
- Agent specialization matrix
- Agent collaboration patterns
- Agent error handling

**moai-core-ask-user-questions**
Provides:
- Interactive question patterns
- Response validation
- Multi-select vs. single-select
- Custom text input handling
- Question workflow design

### MoAI Foundation Skills

**moai-core-spec-authoring**
Provides:
- EARS format mastery
- Ubiquitous patterns
- Event-driven specifications
- Unwanted/error patterns
- State-driven specifications
- Optional/variant patterns

**moai-core-personas**
Provides:
- Alfred persona system
- Yoda deep-learning mode
- R2-D2 production-focused
- Keating personalized learning
- Persona switching strategies

**moai-core-practices**
Provides:
- TRUST 5 principles
- Quality gates
- Code review patterns
- Refactoring strategies
- Technical debt management

**moai-core-rules**
Provides:
- MoAI-ADK rules and constraints
- Configuration requirements
- Environment setup rules
- Security rules
- Git workflow rules

---

## Checklist: Per-Skill Domain/Core Update

### Frontmatter (YAML)
- [ ] `name`: `moai-domain-{domain}` or `moai-core-{capability}`
- [ ] `version`: `4.0.0`
- [ ] `tier`: `Domain` or `Alfred`
- [ ] `created`: 2025-11-11
- [ ] `updated`: 2025-11-19
- [ ] `status`: `stable` or `production`
- [ ] `description`: ~150 chars with primary use cases
- [ ] `primary-agent`: Matches domain (e.g., backend-expert)
- [ ] `secondary-agents`: 2-3 complementary agents
- [ ] `keywords`: 5-7 searchable terms
- [ ] `allowed-tools`: Read, Bash, WebFetch minimum

### Content Structure (Domain)
- [ ] **Tech Stack**: 2025 versions for primary frameworks
- [ ] **5-7 Patterns**: Production-ready code examples
- [ ] **Security Pattern**: OWASP/best practice
- [ ] **Performance Pattern**: Benchmarks included
- [ ] **Testing Pattern**: Real test code
- [ ] **Deployment Pattern**: YAML or IaC example
- [ ] **2-3 Case Studies**: Real-world decisions
- [ ] **10+ Best Practices**: Domain-specific guidance
- [ ] **Cross-references**: Links to complementary Skills

### Content Structure (Core)
- [ ] **Overview**: Clear responsibility statement
- [ ] **Capability Matrix**: What this Skill handles
- [ ] **3-5 Patterns**: Integration examples with code
- [ ] **Error Handling**: How to manage failures
- [ ] **10+ Best Practices**: Specific to this capability
- [ ] **Advanced Topics**: Deep integration scenarios
- [ ] **Monitoring**: How to observe this component
- [ ] **Cross-references**: Links to other Core Skills

### Version Accuracy
- [ ] Framework versions from November 2025
- [ ] No deprecated patterns recommended
- [ ] Security advisories addressed
- [ ] Support timelines verified

### Cross-references
- [ ] All internal references use `Skill("name")` format
- [ ] External links HTTPS and valid
- [ ] References to examples.md and reference.md
- [ ] Links to official documentation

---

## Domain Skill Categories

| Domain | Count | Primary Frameworks | Priority |
|--------|-------|-------------------|----------|
| **Backend** | 1 | FastAPI, Django, Go Fiber | ✅ Done |
| **Frontend** | 1 | React, Next.js, Vue | ✅ Done |
| **CLI Tools** | 1 | Click, Cobra, oclif | Medium |
| **Cloud** | 1 | AWS, GCP, Azure | High |
| **Database** | 1 | PostgreSQL, MongoDB | High |
| **DevOps** | 1 | K8s, Docker, Terraform | High |
| **Data Science** | 1 | pandas, scikit-learn | Medium |
| **ML** | 1 | PyTorch, TensorFlow | Medium |
| **ML Ops** | 1 | MLflow, Kubernetes | Medium |
| **Mobile** | 1 | Swift, Kotlin, Flutter | Medium |
| **Monitoring** | 1 | Prometheus, Grafana | High |
| **Testing** | 1 | pytest, Jest, Vitest | High |
| **Web API** | 1 | REST, GraphQL, gRPC | High |

---

## Core Skill Categories

| Category | Count | Responsibility | Priority |
|----------|-------|-----------------|----------|
| **Agent** | 2 | Agent factory + guide | High |
| **Orchestration** | 2 | Workflow + session management | High |
| **Questioning** | 1 | User interaction | High |
| **Foundation** | 5 | SPEC, personas, practices, rules, config | High |
| **Context** | 2 | Memory management, context budgeting | High |
| **Detection** | 2 | Language + expertise detection | Medium |
| **Development** | 4 | Dev guide, code review, patterns, testing | Medium |
| **Automation** | 2 | Feedback templates, issue labels | Low |

---

## Time Estimates

| Category | Count | Per-Skill | Total |
|----------|-------|-----------|-------|
| **Domain Skills (9)** | 9 | 20-25 min | 3-4 hours |
| **Core Skills (24)** | 24 | 15-20 min | 6-8 hours |
| **Total (33)** | 33 | 17 min avg | 9-12 hours |
| **With Validation** | 33 | 22 min avg | 12-15 hours |
| **Parallel (3-4 devs)** | 33 | 20 min avg | 2.5-3.5 hours |

---

## Quality Gates

Each updated Skill must pass:

### ✅ MUST PASS
- [ ] YAML frontmatter valid and complete
- [ ] Version is exactly "4.0.0"
- [ ] Status is "stable" or "production"
- [ ] Minimum 3 allowed-tools
- [ ] Markdown structure valid
- [ ] No broken internal references

### ✅ SHOULD PASS
- [ ] Framework versions current (Nov 2025)
- [ ] All external links 200 status
- [ ] 5+ code examples with syntax highlighting
- [ ] 10+ best practices or patterns
- [ ] Cross-references use `Skill()` format
- [ ] No spelling or grammar errors

### ✅ NICE TO HAVE
- [ ] Case studies included (Domain only)
- [ ] Performance benchmarks mentioned
- [ ] Security considerations explicit
- [ ] Deployment examples included
- [ ] Architecture diagrams described
- [ ] Consistent tone and style

---

## Next Steps

1. **Categorize remaining skills** by domain/core type
2. **Assign to team members** for parallel processing
3. **Use this template** for each skill update
4. **Validate with TRUST 5** (see PHASE4 guide)
5. **Cross-reference** with other Skills
6. **Update timestamps** to 2025-11-19
7. **Run batch validation** when complete

---

**Scope**: Domain & Core Skills batch update strategy
**Version**: 1.0 (2025-11-19)
**Purpose**: Standardize all 37 Domain + Core Skills to v4.0.0 Enterprise
**Estimated Completion**: 9-12 hours (sequential) or 2.5-3.5 hours (parallel)

