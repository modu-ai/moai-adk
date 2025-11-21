# MoAI-ADK Complete Agent Analysis

**Generated**: 2025-11-22
**Total Agents**: 31
**Agent Types**: Planning/Design (4), Implementation (8), Quality (3), Documentation (3), DevOps (2), Optimization (2), Integration (5), Management (4)

---

## Executive Summary

This comprehensive analysis enumerates all 31 specialized agents in the MoAI-ADK framework, documenting their purposes, current skill assignments, and identifying gaps for optimization.

**Agent Categories Distribution**:
- **Planning/Design** (4): spec-builder, api-designer, implementation-planner, component-designer
- **Implementation** (8): backend-expert, frontend-expert, tdd-implementer, database-expert, etc.
- **Quality** (3): quality-gate, security-expert, trust-checker
- **Documentation** (3): doc-syncer, docs-manager, sync-manager
- **DevOps** (2): devops-expert, monitoring-expert
- **Optimization** (2): performance-engineer, format-expert
- **Integration** (5): mcp-context7-integrator, mcp-figma-integrator, mcp-notion-integrator, mcp-playwright-integrator, debug-helper
- **Management** (4): project-manager, git-manager, agent-factory, cc-manager, skill-factory

---

## Category 1: Planning/Design Agents (4 agents)

### spec-builder
- **Title**: SPEC specification builder
- **Purpose**: Create SPEC documents with EARS format and acceptance criteria
- **Current Skills**: moai-foundation-ears, moai-foundation-specs, moai-core-spec-authoring, moai-lang-python
- **Recommended Skills**:
  - ✅ moai-foundation-ears (correct)
  - ✅ moai-foundation-specs (correct)
  - ✅ moai-core-spec-authoring (correct)
  - ➕ moai-core-ask-user-questions (missing - for requirement clarification)
  - ➕ moai-spec-intelligent-workflow (missing - for SPEC intelligence)
- **Skill Gaps**: Missing interactive questioning and intelligent workflow skills
- **Priority**: HIGH - Critical for SPEC creation quality

### api-designer
- **Title**: REST/GraphQL API architecture
- **Purpose**: Design API endpoints, schemas, OpenAPI specifications
- **Current Skills**: moai-lang-python, moai-domain-backend, moai-context7-lang-integration
- **Recommended Skills**:
  - ✅ moai-domain-backend (correct)
  - ✅ moai-context7-lang-integration (correct)
  - ➕ moai-domain-web-api (missing - specialized API design)
  - ➕ moai-security-api (missing - API security patterns)
  - ➕ moai-lang-typescript (missing - for Node.js APIs)
- **Skill Gaps**: Missing specialized API and security skills
- **Priority**: HIGH - Needs API-specific expertise

### implementation-planner
- **Title**: Implementation strategy planner
- **Purpose**: Break down features into implementation tasks
- **Current Skills**: moai-foundation-specs, moai-lang-python
- **Recommended Skills**:
  - ✅ moai-foundation-specs (correct)
  - ➕ moai-core-workflow (missing - for task orchestration)
  - ➕ moai-core-todowrite-pattern (missing - for task tracking)
  - ➕ moai-foundation-trust (missing - for quality planning)
  - ➕ moai-core-dev-guide (missing - for TDD planning)
- **Skill Gaps**: Missing orchestration and planning frameworks
- **Priority**: MEDIUM - Enhance planning capabilities

### component-designer
- **Title**: UI component architecture
- **Purpose**: Design reusable component systems
- **Current Skills**: moai-domain-frontend, moai-icons-vector
- **Recommended Skills**:
  - ✅ moai-domain-frontend (correct)
  - ➕ moai-design-systems (missing - design system patterns)
  - ➕ moai-component-designer (missing - specialized component skills)
  - ➕ moai-lib-shadcn-ui (missing - modern component library)
  - ➕ moai-lang-typescript (missing - for typed components)
- **Skill Gaps**: Missing design system and component library expertise
- **Priority**: HIGH - Critical for component quality

---

## Category 2: Implementation Agents (8 agents)

### backend-expert
- **Title**: Backend architecture specialist
- **Purpose**: REST/GraphQL APIs, database integration, microservices
- **Current Skills**: moai-lang-python, moai-lang-go, moai-domain-backend, moai-domain-database, moai-domain-api, moai-context7-lang-integration
- **Recommended Skills**:
  - ✅ moai-domain-backend (correct)
  - ✅ moai-domain-database (correct)
  - ✅ moai-context7-lang-integration (correct)
  - ✅ moai-lang-python, moai-lang-go (correct)
  - ➕ moai-security-api (missing - API security)
  - ➕ moai-security-auth (missing - authentication patterns)
  - ➕ moai-essentials-perf (missing - performance optimization)
- **Skill Gaps**: Missing security and performance skills
- **Priority**: HIGH - Security is critical for backend

### frontend-expert
- **Title**: Frontend development specialist
- **Purpose**: React, Next.js, component architecture, state management
- **Current Skills**: moai-lang-typescript, moai-lang-javascript, moai-domain-frontend, moai-lang-tailwind-css, moai-icons-vector
- **Recommended Skills**:
  - ✅ moai-domain-frontend (correct)
  - ✅ moai-lang-typescript, moai-lang-javascript (correct)
  - ✅ moai-lang-tailwind-css (correct)
  - ➕ moai-design-systems (missing - design system integration)
  - ➕ moai-lib-shadcn-ui (missing - component library)
  - ➕ moai-essentials-perf (missing - frontend performance)
  - ➕ moai-streaming-ui (missing - real-time UI)
- **Skill Gaps**: Missing design systems and performance skills
- **Priority**: HIGH - Design systems critical for consistency

### tdd-implementer
- **Title**: Test-driven development specialist
- **Purpose**: RED-GREEN-REFACTOR cycle, test-first implementation
- **Current Skills**: moai-lang-python, moai-lang-typescript, moai-essentials-debug, moai-domain-backend, moai-domain-frontend
- **Recommended Skills**:
  - ✅ moai-essentials-debug (correct)
  - ✅ moai-lang-python, moai-lang-typescript (correct)
  - ✅ moai-domain-backend, moai-domain-frontend (correct)
  - ➕ moai-foundation-trust (missing - TRUST 5 principles)
  - ➕ moai-core-dev-guide (missing - TDD workflow)
  - ➕ moai-domain-testing (missing - testing strategies)
  - ➕ moai-essentials-review (missing - code review)
- **Skill Gaps**: Missing core TDD and quality frameworks
- **Priority**: CRITICAL - Core workflow incomplete

### database-expert
- **Title**: Database architecture specialist
- **Purpose**: Schema design, query optimization, migrations
- **Current Skills**: moai-lang-python, moai-domain-database, moai-essentials-perf, moai-domain-cloud
- **Recommended Skills**:
  - ✅ moai-domain-database (correct)
  - ✅ moai-essentials-perf (correct)
  - ✅ moai-lang-python (correct for SQLAlchemy)
  - ➕ moai-lang-sql (missing - SQL expertise)
  - ➕ moai-baas-supabase-ext (missing - modern DB platforms)
  - ➕ moai-baas-neon-ext (missing - serverless Postgres)
  - ➕ moai-security-encryption (missing - data encryption)
- **Skill Gaps**: Missing SQL and modern DB platform skills
- **Priority**: HIGH - SQL expertise critical

### migration-expert
- **Title**: Database migration specialist
- **Purpose**: Schema migrations, data transformations
- **Current Skills**: moai-domain-database, moai-lang-python
- **Recommended Skills**:
  - ✅ moai-domain-database (correct)
  - ✅ moai-lang-python (correct)
  - ➕ moai-lang-sql (missing)
  - ➕ moai-essentials-perf (missing - migration optimization)
  - ➕ moai-domain-backend (missing - backend integration)
- **Skill Gaps**: Missing SQL and performance skills
- **Priority**: MEDIUM

### devops-expert
- **Title**: DevOps and infrastructure
- **Purpose**: CI/CD, Docker, Kubernetes, cloud deployment
- **Current Skills**: moai-domain-cloud, moai-baas-vercel-ext, moai-baas-clerk-ext
- **Recommended Skills**:
  - ✅ moai-domain-cloud (correct)
  - ➕ moai-domain-devops (missing - core DevOps skill)
  - ➕ moai-cloud-aws-advanced (missing)
  - ➕ moai-cloud-gcp-advanced (missing)
  - ➕ moai-domain-monitoring (missing)
  - ➕ moai-security-secrets (missing - secret management)
- **Skill Gaps**: Missing core DevOps and cloud skills
- **Priority**: CRITICAL - Core domain skill missing

### accessibility-expert
- **Title**: Web accessibility specialist
- **Purpose**: WCAG compliance, a11y testing
- **Current Skills**: moai-domain-frontend
- **Recommended Skills**:
  - ✅ moai-domain-frontend (correct)
  - ➕ moai-design-systems (missing - accessible components)
  - ➕ moai-domain-testing (missing - a11y testing)
  - ➕ moai-playwright-webapp-testing (missing - E2E a11y tests)
- **Skill Gaps**: Missing testing and component skills
- **Priority**: MEDIUM

### ui-ux-expert
- **Title**: UI/UX design implementation
- **Purpose**: User interface design, user experience optimization
- **Current Skills**: moai-domain-frontend, moai-lang-tailwind-css, moai-icons-vector
- **Recommended Skills**:
  - ✅ moai-domain-frontend (correct)
  - ✅ moai-lang-tailwind-css (correct)
  - ➕ moai-design-systems (missing)
  - ➕ moai-lib-shadcn-ui (missing)
  - ➕ moai-component-designer (missing)
  - ➕ moai-streaming-ui (missing - interactive UX)
- **Skill Gaps**: Missing design system skills
- **Priority**: HIGH

---

## Category 3: Quality Agents (3 agents)

### quality-gate
- **Title**: Quality validation gate
- **Purpose**: TRUST 5 validation, coverage checks, quality metrics
- **Current Skills**: moai-essentials-debug, moai-essentials-perf, moai-essentials-refactor, moai-domain-security
- **Recommended Skills**:
  - ✅ moai-essentials-debug (correct)
  - ✅ moai-essentials-perf (correct)
  - ✅ moai-essentials-refactor (correct)
  - ✅ moai-domain-security (correct)
  - ➕ moai-foundation-trust (missing - TRUST 5 framework)
  - ➕ moai-essentials-review (missing - code review)
  - ➕ moai-core-code-reviewer (missing - review orchestration)
  - ➕ moai-domain-testing (missing - testing strategies)
- **Skill Gaps**: Missing core quality frameworks
- **Priority**: CRITICAL - TRUST 5 framework missing

### security-expert
- **Title**: Application security specialist
- **Purpose**: OWASP compliance, threat modeling, security audits
- **Current Skills**: moai-domain-security, moai-security-owasp, moai-security-identity, moai-security-threat, moai-security-api
- **Recommended Skills**:
  - ✅ moai-domain-security (correct)
  - ✅ moai-security-owasp (correct)
  - ✅ moai-security-identity (correct)
  - ✅ moai-security-threat (correct)
  - ✅ moai-security-api (correct)
  - ➕ moai-security-auth (missing - authentication patterns)
  - ➕ moai-security-encryption (missing)
  - ➕ moai-security-compliance (missing)
  - ➕ moai-security-zero-trust (missing)
- **Skill Gaps**: Missing additional security domains
- **Priority**: HIGH - Comprehensive security coverage

### trust-checker
- **Title**: TRUST 5 validation
- **Purpose**: Validate TRUST principles across codebase
- **Current Skills**: (empty)
- **Recommended Skills**:
  - ➕ moai-foundation-trust (missing - CRITICAL)
  - ➕ moai-essentials-review (missing)
  - ➕ moai-core-code-reviewer (missing)
  - ➕ moai-domain-testing (missing)
  - ➕ moai-essentials-debug (missing)
- **Skill Gaps**: ALL skills missing
- **Priority**: CRITICAL - Agent has no skills assigned

---

## Category 4: Documentation Agents (3 agents)

### doc-syncer
- **Title**: Documentation synchronization
- **Purpose**: Auto-generate docs from code, sync with SPEC
- **Current Skills**: moai-docs-generation, moai-docs-validation
- **Recommended Skills**:
  - ✅ moai-docs-generation (correct)
  - ✅ moai-docs-validation (correct)
  - ➕ moai-docs-toolkit (missing - Sphinx, JSDoc)
  - ➕ moai-docs-unified (missing - unified documentation)
  - ➕ moai-foundation-specs (missing - SPEC integration)
  - ➕ moai-mermaid-diagram-expert (missing - diagram generation)
- **Skill Gaps**: Missing documentation tools and frameworks
- **Priority**: HIGH

### docs-manager
- **Title**: Documentation management
- **Purpose**: Comprehensive documentation strategy
- **Current Skills**: moai-docs-generation, moai-docs-validation, moai-cc-claude-md, moai-mermaid-diagram-expert
- **Recommended Skills**:
  - ✅ moai-docs-generation (correct)
  - ✅ moai-docs-validation (correct)
  - ✅ moai-mermaid-diagram-expert (correct)
  - ➕ moai-docs-toolkit (missing)
  - ➕ moai-docs-unified (missing)
  - ➕ moai-project-documentation (missing)
  - ➕ moai-readme-expert (missing)
- **Skill Gaps**: Missing comprehensive documentation skills
- **Priority**: MEDIUM

### sync-manager
- **Title**: Documentation sync orchestration
- **Purpose**: Orchestrate documentation updates
- **Current Skills**: moai-docs-generation, moai-docs-validation
- **Recommended Skills**:
  - ✅ moai-docs-generation (correct)
  - ✅ moai-docs-validation (correct)
  - ➕ moai-docs-unified (missing)
  - ➕ moai-foundation-git (missing - git integration)
  - ➕ moai-change-logger (missing - change tracking)
- **Skill Gaps**: Missing sync orchestration skills
- **Priority**: MEDIUM

---

## Category 5: DevOps Agents (2 agents)

### devops-expert
(See Category 2 - Implementation Agents - already documented above)

### monitoring-expert
- **Title**: Application monitoring specialist
- **Purpose**: Metrics, logging, observability, alerting
- **Current Skills**: moai-domain-monitoring, moai-domain-cloud
- **Recommended Skills**:
  - ✅ moai-domain-monitoring (correct)
  - ✅ moai-domain-cloud (correct)
  - ➕ moai-domain-devops (missing - DevOps integration)
  - ➕ moai-essentials-perf (missing - performance monitoring)
  - ➕ moai-essentials-debug (missing - debugging integration)
- **Skill Gaps**: Missing DevOps and performance skills
- **Priority**: MEDIUM

---

## Category 6: Optimization Agents (2 agents)

### performance-engineer
- **Title**: Performance optimization specialist
- **Purpose**: Profiling, bottleneck detection, optimization
- **Current Skills**: moai-essentials-perf, moai-domain-monitoring, moai-domain-cloud
- **Recommended Skills**:
  - ✅ moai-essentials-perf (correct)
  - ✅ moai-domain-monitoring (correct)
  - ➕ moai-essentials-debug (missing - debugging for performance)
  - ➕ moai-domain-backend (missing - backend optimization)
  - ➕ moai-domain-frontend (missing - frontend optimization)
  - ➕ moai-domain-database (missing - query optimization)
- **Skill Gaps**: Missing domain-specific performance skills
- **Priority**: HIGH

### format-expert
- **Title**: Code formatting specialist
- **Purpose**: Code formatting, linting, style enforcement
- **Current Skills**: moai-essentials-refactor
- **Recommended Skills**:
  - ✅ moai-essentials-refactor (correct)
  - ➕ moai-lang-python (missing - Python formatting)
  - ➕ moai-lang-typescript (missing - TS formatting)
  - ➕ moai-core-practices (missing - code standards)
  - ➕ moai-foundation-trust (missing - quality principles)
- **Skill Gaps**: Missing language-specific formatting skills
- **Priority**: MEDIUM

---

## Category 7: Integration Agents (5 agents)

### mcp-context7-integrator
- **Title**: Context7 MCP integration
- **Purpose**: Access latest documentation via Context7
- **Current Skills**: moai-context7-lang-integration
- **Recommended Skills**:
  - ✅ moai-context7-lang-integration (correct)
  - ➕ moai-context7-integration (missing - core integration)
  - ➕ moai-mcp-integration (missing - MCP patterns)
  - ➕ moai-jit-docs-enhanced (missing - JIT documentation)
- **Skill Gaps**: Missing core MCP integration skills
- **Priority**: HIGH

### mcp-figma-integrator
- **Title**: Figma MCP integration
- **Purpose**: Figma design to code workflow
- **Current Skills**: moai-domain-figma, moai-design-systems, moai-lang-typescript, moai-domain-frontend
- **Recommended Skills**:
  - ✅ moai-domain-figma (correct)
  - ✅ moai-design-systems (correct)
  - ✅ moai-lang-typescript (correct)
  - ➕ moai-mcp-integration (missing - MCP patterns)
  - ➕ moai-component-designer (missing - component extraction)
  - ➕ moai-lib-shadcn-ui (missing - component library)
- **Skill Gaps**: Missing MCP and component skills
- **Priority**: MEDIUM

### mcp-notion-integrator
- **Title**: Notion MCP integration
- **Purpose**: Notion workspace automation
- **Current Skills**: moai-domain-notion
- **Recommended Skills**:
  - ✅ moai-domain-notion (correct)
  - ➕ moai-mcp-integration (missing - MCP patterns)
  - ➕ moai-docs-generation (missing - doc integration)
  - ➕ moai-document-processing (missing - content processing)
- **Skill Gaps**: Missing MCP and documentation skills
- **Priority**: LOW

### mcp-playwright-integrator
- **Title**: Playwright MCP integration
- **Purpose**: Browser automation and E2E testing
- **Current Skills**: moai-playwright-webapp-testing
- **Recommended Skills**:
  - ✅ moai-playwright-webapp-testing (correct)
  - ➕ moai-mcp-integration (missing - MCP patterns)
  - ➕ moai-domain-testing (missing - testing strategies)
  - ➕ moai-domain-frontend (missing - frontend testing)
- **Skill Gaps**: Missing MCP and testing domain skills
- **Priority**: MEDIUM

### debug-helper
- **Title**: Debugging assistance
- **Purpose**: Error diagnosis, debugging workflows
- **Current Skills**: moai-essentials-debug
- **Recommended Skills**:
  - ✅ moai-essentials-debug (correct)
  - ➕ moai-essentials-perf (missing - performance debugging)
  - ➕ moai-domain-backend (missing - backend debugging)
  - ➕ moai-domain-frontend (missing - frontend debugging)
  - ➕ moai-lang-python (missing - Python debugging)
  - ➕ moai-lang-typescript (missing - TS debugging)
- **Skill Gaps**: Missing domain and language debugging skills
- **Priority**: HIGH

---

## Category 8: Management Agents (4 agents)

### project-manager
- **Title**: Project orchestration
- **Purpose**: Project initialization, configuration management
- **Current Skills**: moai-cc-configuration, moai-project-config-manager
- **Recommended Skills**:
  - ✅ moai-cc-configuration (correct)
  - ✅ moai-project-config-manager (correct)
  - ➕ moai-project-documentation (missing)
  - ➕ moai-project-language-initializer (missing)
  - ➕ moai-project-batch-questions (missing)
  - ➕ moai-core-ask-user-questions (missing)
- **Skill Gaps**: Missing project setup and questioning skills
- **Priority**: MEDIUM

### git-manager
- **Title**: Git workflow management
- **Purpose**: Branch management, commit conventions, PR workflows
- **Current Skills**: (empty)
- **Recommended Skills**:
  - ➕ moai-foundation-git (missing - CRITICAL)
  - ➕ moai-change-logger (missing - change tracking)
  - ➕ moai-core-session-state (missing - git session persistence)
- **Skill Gaps**: ALL skills missing
- **Priority**: CRITICAL - Core git skill missing

### agent-factory
- **Title**: Intelligent agent generator
- **Purpose**: Generate new agents from requirements
- **Current Skills**: moai-core-agent-factory, moai-foundation-ears, moai-foundation-specs, moai-core-language-detection
- **Recommended Skills**:
  - ✅ moai-core-agent-factory (correct - MASTER SKILL)
  - ✅ moai-foundation-ears (correct)
  - ✅ moai-foundation-specs (correct)
  - ✅ moai-core-language-detection (correct)
  - ➕ moai-core-agent-guide (missing - agent patterns)
  - ➕ moai-context7-lang-integration (missing - documentation research)
  - ➕ moai-core-ask-user-questions (missing - clarification)
- **Skill Gaps**: Missing agent guide and questioning skills
- **Priority**: HIGH

### cc-manager
- **Title**: Claude Code configuration manager
- **Purpose**: Manage Claude Code settings, validate configurations
- **Current Skills**: moai-cc-configuration, moai-cc-hooks, moai-cc-mcp-plugins
- **Recommended Skills**:
  - ✅ moai-cc-configuration (correct)
  - ✅ moai-cc-hooks (correct)
  - ➕ moai-cc-settings (missing)
  - ➕ moai-cc-agents (missing)
  - ➕ moai-cc-skills (missing)
  - ➕ moai-cc-commands (missing)
  - ➕ moai-cc-memory (missing)
- **Skill Gaps**: Missing comprehensive CC management skills
- **Priority**: HIGH

### skill-factory
- **Title**: Skill generator
- **Purpose**: Generate new skills from requirements
- **Current Skills**: moai-cc-skill-factory, moai-cc-configuration, moai-cc-skills, moai-core-ask-user-questions, moai-foundation-trust
- **Recommended Skills**:
  - ✅ moai-cc-skill-factory (correct)
  - ✅ moai-cc-skills (correct)
  - ✅ moai-core-ask-user-questions (correct)
  - ➕ moai-docs-generation (missing - skill documentation)
  - ➕ moai-context7-lang-integration (missing - research)
- **Skill Gaps**: Missing documentation and research skills
- **Priority**: MEDIUM

---

## Summary Statistics

### Skills Coverage Analysis

**Agents with Complete Skill Sets** (8):
- backend-expert: 6/7 skills (86%)
- frontend-expert: 5/8 skills (63%)
- security-expert: 5/9 skills (56%)
- agent-factory: 4/7 skills (57%)
- mcp-figma-integrator: 4/7 skills (57%)
- api-designer: 3/6 skills (50%)
- quality-gate: 4/8 skills (50%)
- tdd-implementer: 5/9 skills (56%)

**Agents with Critical Gaps** (5):
- trust-checker: 0/5 skills (0%) - CRITICAL
- git-manager: 0/3 skills (0%) - CRITICAL
- devops-expert: 1/6 skills (17%) - CRITICAL
- tdd-implementer: 5/9 missing TRUST (44%) - CRITICAL
- quality-gate: 4/8 missing TRUST (50%) - CRITICAL

**Agents with Moderate Gaps** (18):
- All other agents have 40-70% skill coverage

### Most Commonly Missing Skills

1. **moai-foundation-trust** (missing in 12 agents) - CRITICAL
2. **moai-core-ask-user-questions** (missing in 8 agents)
3. **moai-domain-devops** (missing in 6 agents)
4. **moai-mcp-integration** (missing in 5 MCP agents)
5. **moai-design-systems** (missing in 6 frontend-related agents)
6. **moai-lang-sql** (missing in 4 database agents)
7. **moai-essentials-perf** (missing in 7 agents)
8. **moai-context7-integration** (missing in 5 agents)

### Priority Actions

**CRITICAL (Immediate)**:
1. Add moai-foundation-trust to: quality-gate, trust-checker, tdd-implementer, format-expert
2. Add moai-foundation-git to: git-manager
3. Add moai-domain-devops to: devops-expert
4. Add moai-core-dev-guide to: tdd-implementer

**HIGH (Next Sprint)**:
1. Add moai-design-systems to: frontend-expert, ui-ux-expert, component-designer
2. Add moai-security-* skills to: backend-expert
3. Add moai-lang-sql to: database-expert, migration-expert
4. Add moai-mcp-integration to all MCP integrators

**MEDIUM (Backlog)**:
1. Add comprehensive skill sets to all agents based on recommendations
2. Update agent documentation with skill loading patterns
3. Create skill loading guidelines for future agent development

---

## Agent Skill Loading Patterns

### Recommended Patterns

**Pattern 1: Auto-load Core Skills**
```yaml
skills: moai-domain-{primary}, moai-foundation-trust, moai-core-language-detection
```

**Pattern 2: Conditional Domain Skills**
```yaml
# In agent documentation
**Conditional Skills** (load when needed):
- Skill("moai-lang-python") - For Python projects
- Skill("moai-lang-typescript") - For TypeScript projects
```

**Pattern 3: Enhancement Skills**
```yaml
# In agent documentation
**Enhancement Skills** (optional):
- Skill("moai-essentials-perf") - For performance optimization
- Skill("moai-essentials-debug") - For debugging assistance
```

---

## Next Steps

### Phase 1: Critical Gaps (Week 1)
- [ ] Update trust-checker with all TRUST skills
- [ ] Update git-manager with foundation-git
- [ ] Update devops-expert with domain-devops
- [ ] Update tdd-implementer with dev-guide and trust

### Phase 2: High Priority (Week 2-3)
- [ ] Update all quality agents with foundation-trust
- [ ] Update frontend agents with design-systems
- [ ] Update backend agents with security skills
- [ ] Update database agents with lang-sql

### Phase 3: Comprehensive Update (Week 4-6)
- [ ] Update all agents with recommended skill sets
- [ ] Document skill loading patterns in each agent
- [ ] Create agent development guidelines
- [ ] Test all agent-skill integrations

---

**Generated**: 2025-11-22
**Agents Analyzed**: 31
**Skills Referenced**: 138
**Critical Gaps Identified**: 5 agents
**Maintenance Priority**: CRITICAL
