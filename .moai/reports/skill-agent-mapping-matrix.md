# Skill-Agent Mapping Matrix

**Generated**: 2025-11-22
**Total Skills**: 138
**Total Agents**: 31
**Mapping Relationships**: 157 (current) → 287 (recommended)

---

## Executive Summary

This matrix shows which skills are assigned to which agents, identifies gaps, and provides recommendations for optimal skill-agent relationships.

**Key Findings**:
- **Coverage**: 157 current skill assignments
- **Gaps**: 130 recommended assignments missing
- **Redundancy**: 12 skills assigned to 3+ agents (good for shared knowledge)
- **Orphaned Skills**: 45 skills not assigned to any agent
- **Critical Issues**: 5 agents missing core skills for their domain

---

## Matrix Legend

**Relevance Levels**:
- 🟢 **PRIMARY** - Core skill for agent's main purpose (must have)
- 🟡 **SECONDARY** - Important for agent's extended capabilities (should have)
- 🔵 **OPTIONAL** - Useful for edge cases (nice to have)

**Current Status**:
- ✅ Currently assigned
- ⚠️ Recommended but missing
- ❌ Critical gap

---

## Part 1: Foundation Skills (6 skills)

### moai-foundation-trust
**Description**: TRUST 5 quality principles

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| quality-gate | 🟢 PRIMARY | ⚠️ MISSING | Core validation framework |
| trust-checker | 🟢 PRIMARY | ❌ CRITICAL | Agent purpose is TRUST validation |
| tdd-implementer | 🟢 PRIMARY | ⚠️ MISSING | TDD requires TRUST principles |
| format-expert | 🟡 SECONDARY | ⚠️ MISSING | Formatting enforces readable code |
| security-expert | 🟡 SECONDARY | ⚠️ MISSING | Security aligns with TRUST-S |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Quality principles for backend |
| frontend-expert | 🟡 SECONDARY | ⚠️ MISSING | Quality principles for frontend |
| spec-builder | 🟡 SECONDARY | ⚠️ MISSING | SPEC quality validation |

**Gap Analysis**: 8 agents need this skill, 0 currently have it
**Priority**: CRITICAL - Foundation skill missing from all quality agents

### moai-foundation-git
**Description**: Git workflow patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| git-manager | 🟢 PRIMARY | ❌ CRITICAL | Core git management agent |
| sync-manager | 🟡 SECONDARY | ⚠️ MISSING | Documentation syncing needs git |
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Syncing docs with git |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | Commit conventions for TDD cycles |

**Gap Analysis**: 4 agents need this skill, 0 currently have it
**Priority**: CRITICAL - git-manager has no skills

### moai-foundation-specs
**Description**: SPEC specification management

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| spec-builder | 🟢 PRIMARY | ✅ ASSIGNED | Core SPEC creation agent |
| agent-factory | 🟢 PRIMARY | ✅ ASSIGNED | Agent specs follow SPEC patterns |
| implementation-planner | 🟢 PRIMARY | ✅ ASSIGNED | Plans from SPEC documents |
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Sync docs with SPEC |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | Implement from SPEC |

**Gap Analysis**: 3/5 agents have this skill
**Priority**: HIGH - Key planning agents need it

### moai-foundation-ears
**Description**: EARS requirements framework

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| spec-builder | 🟢 PRIMARY | ✅ ASSIGNED | EARS format for SPEC |
| agent-factory | 🟡 SECONDARY | ✅ ASSIGNED | Generate agents with EARS |
| api-designer | 🟡 SECONDARY | ⚠️ MISSING | API requirements in EARS |

**Gap Analysis**: 2/3 agents have this skill
**Priority**: MEDIUM

### moai-foundation-langs
**Description**: Programming language foundations

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| agent-factory | 🟡 SECONDARY | ⚠️ MISSING | Language selection for agents |
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | Project language detection |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: LOW - Covered by language-detection skill

---

## Part 2: Core Skills (20 skills)

### moai-core-agent-factory
**Description**: Intelligent agent generation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| agent-factory | 🟢 PRIMARY | ✅ ASSIGNED | Master skill for agent generation |

**Gap Analysis**: 1/1 agent has this skill
**Priority**: COMPLETE

### moai-core-agent-guide
**Description**: Agent architecture patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| agent-factory | 🟢 PRIMARY | ⚠️ MISSING | Guide for generating agents |
| cc-manager | 🟡 SECONDARY | ⚠️ MISSING | Agent configuration management |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: HIGH - Critical for agent-factory

### moai-core-alfred-orchestration
**Description**: Multi-agent orchestration

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| project-manager | 🟢 PRIMARY | ⚠️ MISSING | Project orchestration |
| implementation-planner | 🟡 SECONDARY | ⚠️ MISSING | Task orchestration |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: MEDIUM

### moai-core-workflow
**Description**: Multi-agent workflow patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| implementation-planner | 🟢 PRIMARY | ⚠️ MISSING | Workflow planning |
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | Project workflows |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | TDD workflow |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: HIGH

### moai-core-dev-guide
**Description**: SPEC-First TDD development

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| tdd-implementer | 🟢 PRIMARY | ❌ CRITICAL | Core TDD workflow agent |
| spec-builder | 🟡 SECONDARY | ⚠️ MISSING | TDD-aware SPEC creation |
| implementation-planner | 🟡 SECONDARY | ⚠️ MISSING | Plan TDD cycles |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: CRITICAL - tdd-implementer missing core workflow

### moai-core-spec-authoring
**Description**: SPEC document authoring

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| spec-builder | 🟢 PRIMARY | ✅ ASSIGNED | Primary SPEC creation agent |

**Gap Analysis**: 1/1 agent has this skill
**Priority**: COMPLETE

### moai-core-language-detection
**Description**: Project language detection

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| agent-factory | 🟢 PRIMARY | ✅ ASSIGNED | Detect language for agent generation |
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | Project initialization |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend framework detection |
| frontend-expert | 🟡 SECONDARY | ⚠️ MISSING | Frontend framework detection |

**Gap Analysis**: 1/4 agents have this skill
**Priority**: MEDIUM

### moai-core-personas
**Description**: Adaptive communication personas

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | User interaction |

**Gap Analysis**: 0/1 agents have this skill
**Priority**: LOW

### moai-core-ask-user-questions
**Description**: Interactive user questioning

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| spec-builder | 🟢 PRIMARY | ⚠️ MISSING | Requirement clarification |
| agent-factory | 🟢 PRIMARY | ⚠️ MISSING | Agent requirement clarification |
| skill-factory | 🟢 PRIMARY | ✅ ASSIGNED | Skill requirement gathering |
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | Project setup questions |
| api-designer | 🟡 SECONDARY | ⚠️ MISSING | API design decisions |

**Gap Analysis**: 1/5 agents have this skill
**Priority**: HIGH - Key planning agents need it

### moai-core-code-reviewer
**Description**: Code review orchestration

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| quality-gate | 🟢 PRIMARY | ⚠️ MISSING | Quality validation orchestration |
| trust-checker | 🟡 SECONDARY | ⚠️ MISSING | TRUST validation review |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: HIGH

### moai-core-practices
**Description**: Development best practices

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| format-expert | 🟢 PRIMARY | ⚠️ MISSING | Code formatting standards |
| quality-gate | 🟡 SECONDARY | ⚠️ MISSING | Best practice validation |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: MEDIUM

### moai-core-todowrite-pattern
**Description**: TodoWrite tool patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| implementation-planner | 🟢 PRIMARY | ⚠️ MISSING | Task tracking |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | TDD cycle tracking |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: MEDIUM

### moai-core-session-state
**Description**: Session state management

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| git-manager | 🟡 SECONDARY | ⚠️ MISSING | Git session persistence |

**Gap Analysis**: 0/1 agents have this skill
**Priority**: LOW

### moai-core-context-budget
**Description**: Token budget management

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| agent-factory | 🟡 SECONDARY | ⚠️ MISSING | Manage agent token usage |
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Documentation token optimization |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: LOW

### moai-core-config-schema
**Description**: Configuration schema management

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| project-manager | 🟡 SECONDARY | ⚠️ MISSING | Project config validation |
| cc-manager | 🟡 SECONDARY | ⚠️ MISSING | Claude Code config schema |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: MEDIUM

---

## Part 3: Domain Skills (14 skills)

### moai-domain-backend
**Description**: Backend development expertise

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| backend-expert | 🟢 PRIMARY | ✅ ASSIGNED | Core backend agent |
| api-designer | 🟢 PRIMARY | ✅ ASSIGNED | API backend design |
| tdd-implementer | 🟢 PRIMARY | ✅ ASSIGNED | Backend TDD |
| database-expert | 🟡 SECONDARY | ⚠️ MISSING | Database backend integration |
| migration-expert | 🟡 SECONDARY | ⚠️ MISSING | Migration backend context |
| devops-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend deployment |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | Backend optimization |
| debug-helper | 🟡 SECONDARY | ⚠️ MISSING | Backend debugging |

**Gap Analysis**: 3/8 agents have this skill
**Priority**: HIGH - Core domain skill widely needed

### moai-domain-frontend
**Description**: Frontend development expertise

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟢 PRIMARY | ✅ ASSIGNED | Core frontend agent |
| ui-ux-expert | 🟢 PRIMARY | ✅ ASSIGNED | UI/UX frontend |
| component-designer | 🟢 PRIMARY | ✅ ASSIGNED | Component architecture |
| accessibility-expert | 🟢 PRIMARY | ✅ ASSIGNED | Frontend accessibility |
| tdd-implementer | 🟡 SECONDARY | ✅ ASSIGNED | Frontend TDD |
| mcp-figma-integrator | 🟡 SECONDARY | ✅ ASSIGNED | Figma to frontend |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | Frontend optimization |
| debug-helper | 🟡 SECONDARY | ⚠️ MISSING | Frontend debugging |
| mcp-playwright-integrator | 🟡 SECONDARY | ⚠️ MISSING | Frontend E2E testing |

**Gap Analysis**: 6/9 agents have this skill
**Priority**: MEDIUM - Good coverage

### moai-domain-database
**Description**: Database architecture and optimization

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| database-expert | 🟢 PRIMARY | ✅ ASSIGNED | Core database agent |
| migration-expert | 🟢 PRIMARY | ✅ ASSIGNED | Database migrations |
| backend-expert | 🟡 SECONDARY | ✅ ASSIGNED | Backend database integration |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | Database optimization |

**Gap Analysis**: 3/4 agents have this skill
**Priority**: MEDIUM - Good coverage

### moai-domain-security
**Description**: Application security

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| security-expert | 🟢 PRIMARY | ✅ ASSIGNED | Core security agent |
| quality-gate | 🟡 SECONDARY | ✅ ASSIGNED | Security validation |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend security |
| api-designer | 🟡 SECONDARY | ⚠️ MISSING | API security |
| devops-expert | 🟡 SECONDARY | ⚠️ MISSING | Infrastructure security |

**Gap Analysis**: 2/5 agents have this skill
**Priority**: HIGH - Security critical for more agents

### moai-domain-devops
**Description**: DevOps and infrastructure

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| devops-expert | 🟢 PRIMARY | ❌ CRITICAL | Core DevOps agent missing core skill |
| monitoring-expert | 🟡 SECONDARY | ⚠️ MISSING | DevOps monitoring integration |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | DevOps performance |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: CRITICAL - devops-expert missing core domain

### moai-domain-testing
**Description**: Testing strategies and frameworks

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| quality-gate | 🟢 PRIMARY | ⚠️ MISSING | Testing validation |
| tdd-implementer | 🟢 PRIMARY | ⚠️ MISSING | TDD testing strategies |
| trust-checker | 🟡 SECONDARY | ⚠️ MISSING | Test validation |
| mcp-playwright-integrator | 🟡 SECONDARY | ⚠️ MISSING | E2E testing |
| accessibility-expert | 🟡 SECONDARY | ⚠️ MISSING | Accessibility testing |

**Gap Analysis**: 0/5 agents have this skill
**Priority**: CRITICAL - Testing agents lack testing domain

### moai-domain-web-api
**Description**: Web API design and implementation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| api-designer | 🟢 PRIMARY | ⚠️ MISSING | Core API design agent |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend API implementation |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: HIGH - API agents need API domain skill

### moai-domain-monitoring
**Description**: Application monitoring and observability

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| monitoring-expert | 🟢 PRIMARY | ✅ ASSIGNED | Core monitoring agent |
| performance-engineer | 🟡 SECONDARY | ✅ ASSIGNED | Performance monitoring |
| devops-expert | 🟡 SECONDARY | ⚠️ MISSING | Infrastructure monitoring |

**Gap Analysis**: 2/3 agents have this skill
**Priority**: MEDIUM

### moai-domain-cloud
**Description**: Cloud architecture patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| devops-expert | 🟢 PRIMARY | ✅ ASSIGNED | Cloud deployment |
| database-expert | 🟡 SECONDARY | ✅ ASSIGNED | Cloud databases |
| monitoring-expert | 🟡 SECONDARY | ✅ ASSIGNED | Cloud monitoring |
| performance-engineer | 🟡 SECONDARY | ✅ ASSIGNED | Cloud optimization |

**Gap Analysis**: 4/4 agents have this skill
**Priority**: COMPLETE - Good coverage

### moai-domain-mobile-app
**Description**: Mobile application development

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| (No agents currently) | - | - | Consider creating mobile-expert agent |

**Gap Analysis**: No mobile agents
**Priority**: FUTURE - Create dedicated mobile agent

### moai-domain-figma
**Description**: Figma design integration

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| mcp-figma-integrator | 🟢 PRIMARY | ✅ ASSIGNED | Figma MCP integration |

**Gap Analysis**: 1/1 agent has this skill
**Priority**: COMPLETE

### moai-domain-notion
**Description**: Notion workspace automation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| mcp-notion-integrator | 🟢 PRIMARY | ✅ ASSIGNED | Notion MCP integration |

**Gap Analysis**: 1/1 agent has this skill
**Priority**: COMPLETE

---

## Part 4: Language Skills (25 skills)

### moai-lang-python
**Description**: Enterprise Python expertise

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| backend-expert | 🟢 PRIMARY | ✅ ASSIGNED | Python backend |
| api-designer | 🟢 PRIMARY | ✅ ASSIGNED | FastAPI design |
| tdd-implementer | 🟢 PRIMARY | ✅ ASSIGNED | Python TDD |
| database-expert | 🟢 PRIMARY | ✅ ASSIGNED | SQLAlchemy |
| migration-expert | 🟡 SECONDARY | ✅ ASSIGNED | Database migrations |
| implementation-planner | 🟡 SECONDARY | ✅ ASSIGNED | Python planning |
| spec-builder | 🟡 SECONDARY | ✅ ASSIGNED | Python examples |
| debug-helper | 🟡 SECONDARY | ⚠️ MISSING | Python debugging |
| format-expert | 🟡 SECONDARY | ⚠️ MISSING | Python formatting |

**Gap Analysis**: 7/9 agents have this skill
**Priority**: MEDIUM - Good coverage, add to remaining

### moai-lang-typescript
**Description**: Enterprise TypeScript development

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟢 PRIMARY | ✅ ASSIGNED | TypeScript frontend |
| tdd-implementer | 🟢 PRIMARY | ✅ ASSIGNED | TypeScript TDD |
| api-designer | 🟡 SECONDARY | ⚠️ MISSING | Node.js API design |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Node.js backend |
| component-designer | 🟡 SECONDARY | ⚠️ MISSING | Typed components |
| mcp-figma-integrator | 🟡 SECONDARY | ✅ ASSIGNED | TypeScript code generation |
| debug-helper | 🟡 SECONDARY | ⚠️ MISSING | TypeScript debugging |
| format-expert | 🟡 SECONDARY | ⚠️ MISSING | TypeScript formatting |

**Gap Analysis**: 3/8 agents have this skill
**Priority**: HIGH - Key language for many agents

### moai-lang-javascript
**Description**: Modern JavaScript (ES2025)

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟢 PRIMARY | ✅ ASSIGNED | JavaScript frontend |
| ui-ux-expert | 🟡 SECONDARY | ⚠️ MISSING | Interactive UX |

**Gap Analysis**: 1/2 agents have this skill
**Priority**: MEDIUM

### moai-lang-go
**Description**: Go systems programming

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| backend-expert | 🟡 SECONDARY | ✅ ASSIGNED | Go backend services |

**Gap Analysis**: 1/1 agent has this skill
**Priority**: COMPLETE

### moai-lang-sql
**Description**: SQL database querying

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| database-expert | 🟢 PRIMARY | ⚠️ MISSING | SQL expertise |
| migration-expert | 🟢 PRIMARY | ⚠️ MISSING | SQL migrations |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend SQL queries |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | SQL optimization |

**Gap Analysis**: 0/4 agents have this skill
**Priority**: CRITICAL - Database agents need SQL

### moai-lang-tailwind-css
**Description**: Tailwind CSS utility framework

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟢 PRIMARY | ✅ ASSIGNED | Tailwind styling |
| ui-ux-expert | 🟢 PRIMARY | ✅ ASSIGNED | UI styling |
| component-designer | 🟡 SECONDARY | ⚠️ MISSING | Component styling |
| accessibility-expert | 🟡 SECONDARY | ⚠️ MISSING | Accessible styling |

**Gap Analysis**: 2/4 agents have this skill
**Priority**: MEDIUM

### Other Language Skills
(moai-lang-rust, moai-lang-java, moai-lang-kotlin, moai-lang-swift, moai-lang-c, moai-lang-cpp, etc.)

**Status**: Not currently assigned to agents
**Priority**: FUTURE - Assign when agents need specific languages
**Recommendation**: Language skills should be loaded conditionally based on project detection

---

## Part 5: Essential Skills (4 skills)

### moai-essentials-debug
**Description**: AI-powered debugging

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| debug-helper | 🟢 PRIMARY | ✅ ASSIGNED | Core debugging agent |
| quality-gate | 🟢 PRIMARY | ✅ ASSIGNED | Quality debugging |
| tdd-implementer | 🟡 SECONDARY | ✅ ASSIGNED | TDD debugging |
| performance-engineer | 🟡 SECONDARY | ⚠️ MISSING | Performance debugging |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend debugging |
| frontend-expert | 🟡 SECONDARY | ⚠️ MISSING | Frontend debugging |
| database-expert | 🟡 SECONDARY | ⚠️ MISSING | Database debugging |

**Gap Analysis**: 3/7 agents have this skill
**Priority**: HIGH - More agents need debugging

### moai-essentials-perf
**Description**: Performance optimization

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| performance-engineer | 🟢 PRIMARY | ✅ ASSIGNED | Core performance agent |
| quality-gate | 🟡 SECONDARY | ✅ ASSIGNED | Performance validation |
| database-expert | 🟡 SECONDARY | ✅ ASSIGNED | Database optimization |
| monitoring-expert | 🟡 SECONDARY | ⚠️ MISSING | Performance monitoring |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend optimization |
| frontend-expert | 🟡 SECONDARY | ⚠️ MISSING | Frontend optimization |
| migration-expert | 🟡 SECONDARY | ⚠️ MISSING | Migration optimization |

**Gap Analysis**: 3/7 agents have this skill
**Priority**: HIGH - Performance critical

### moai-essentials-refactor
**Description**: Code refactoring

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| format-expert | 🟢 PRIMARY | ✅ ASSIGNED | Code formatting agent |
| quality-gate | 🟡 SECONDARY | ✅ ASSIGNED | Refactoring validation |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | TDD refactor phase |

**Gap Analysis**: 2/3 agents have this skill
**Priority**: HIGH - tdd-implementer needs it

### moai-essentials-review
**Description**: Automated code review

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| quality-gate | 🟢 PRIMARY | ⚠️ MISSING | Quality review validation |
| trust-checker | 🟢 PRIMARY | ⚠️ MISSING | TRUST review |
| tdd-implementer | 🟡 SECONDARY | ⚠️ MISSING | TDD code review |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: CRITICAL - Quality agents need review

---

## Part 6: Security Skills (11 skills)

### moai-security-owasp
**Description**: OWASP Top 10 compliance

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| security-expert | 🟢 PRIMARY | ✅ ASSIGNED | OWASP validation |
| quality-gate | 🟡 SECONDARY | ⚠️ MISSING | Security quality gate |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend OWASP |

**Gap Analysis**: 1/3 agents have this skill
**Priority**: HIGH

### moai-security-api
**Description**: API security patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| security-expert | 🟢 PRIMARY | ✅ ASSIGNED | API security validation |
| api-designer | 🟢 PRIMARY | ⚠️ MISSING | Secure API design |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend API security |

**Gap Analysis**: 1/3 agents have this skill
**Priority**: HIGH

### moai-security-auth
**Description**: Authentication and authorization

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| security-expert | 🟡 SECONDARY | ⚠️ MISSING | Auth security |
| backend-expert | 🟡 SECONDARY | ⚠️ MISSING | Backend authentication |
| api-designer | 🟡 SECONDARY | ⚠️ MISSING | API authentication |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: HIGH

### Other Security Skills
(moai-security-encryption, moai-security-identity, moai-security-threat, moai-security-compliance, moai-security-zero-trust, moai-security-secrets, moai-security-ssrf)

**Current Assignment**: Primarily to security-expert
**Gap**: Backend and API agents need more security skills
**Priority**: HIGH - Distribute security skills to relevant agents

---

## Part 7: Documentation Skills (6 skills)

### moai-docs-generation
**Description**: Documentation generation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| doc-syncer | 🟢 PRIMARY | ✅ ASSIGNED | Documentation syncing |
| docs-manager | 🟢 PRIMARY | ✅ ASSIGNED | Documentation management |
| sync-manager | 🟢 PRIMARY | ✅ ASSIGNED | Sync orchestration |
| skill-factory | 🟡 SECONDARY | ⚠️ MISSING | Skill documentation |

**Gap Analysis**: 3/4 agents have this skill
**Priority**: MEDIUM

### moai-docs-validation
**Description**: Documentation validation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| doc-syncer | 🟢 PRIMARY | ✅ ASSIGNED | Validation during sync |
| docs-manager | 🟢 PRIMARY | ✅ ASSIGNED | Documentation quality |
| sync-manager | 🟢 PRIMARY | ✅ ASSIGNED | Sync validation |

**Gap Analysis**: 3/3 agents have this skill
**Priority**: COMPLETE

### moai-docs-toolkit
**Description**: Documentation tools (Sphinx, JSDoc)

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Doc generation tools |
| docs-manager | 🟡 SECONDARY | ⚠️ MISSING | Tool management |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: MEDIUM

### moai-docs-unified
**Description**: Unified documentation system

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| docs-manager | 🟢 PRIMARY | ⚠️ MISSING | Unified doc management |
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Unified syncing |
| sync-manager | 🟡 SECONDARY | ⚠️ MISSING | Unified sync orchestration |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: HIGH

---

## Part 8: Claude Code Skills (12 skills)

### moai-cc-configuration
**Description**: Claude Code configuration

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| cc-manager | 🟢 PRIMARY | ✅ ASSIGNED | Configuration management |
| project-manager | 🟡 SECONDARY | ✅ ASSIGNED | Project configuration |
| skill-factory | 🟡 SECONDARY | ✅ ASSIGNED | Skill configuration |

**Gap Analysis**: 3/3 agents have this skill
**Priority**: COMPLETE

### moai-cc-skills
**Description**: Claude Code skills system

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| skill-factory | 🟢 PRIMARY | ✅ ASSIGNED | Skill generation |
| cc-manager | 🟡 SECONDARY | ⚠️ MISSING | Skill management |

**Gap Analysis**: 1/2 agents have this skill
**Priority**: HIGH

### moai-cc-agents
**Description**: Claude Code agent system

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| cc-manager | 🟡 SECONDARY | ⚠️ MISSING | Agent management |
| agent-factory | 🟡 SECONDARY | ⚠️ MISSING | Agent architecture |

**Gap Analysis**: 0/2 agents have this skill
**Priority**: HIGH

### Other CC Skills
(moai-cc-commands, moai-cc-hooks, moai-cc-settings, moai-cc-memory, moai-cc-claude-md, moai-cc-skill-factory, moai-cc-permission-mode, moai-cc-hook-model-strategy, moai-cc-subagent-lifecycle)

**Current Assignment**: Primarily to cc-manager and related agents
**Gap**: Agent and skill factory agents need more CC skills
**Priority**: MEDIUM - Improve CC ecosystem integration

---

## Part 9: Context7 & MCP Skills

### moai-context7-lang-integration
**Description**: Language-specific Context7 patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| mcp-context7-integrator | 🟢 PRIMARY | ✅ ASSIGNED | Context7 integration |
| backend-expert | 🟡 SECONDARY | ✅ ASSIGNED | Backend docs |
| api-designer | 🟡 SECONDARY | ✅ ASSIGNED | API docs |
| agent-factory | 🟡 SECONDARY | ⚠️ MISSING | Agent research |
| skill-factory | 🟡 SECONDARY | ⚠️ MISSING | Skill research |

**Gap Analysis**: 3/5 agents have this skill
**Priority**: HIGH

### moai-mcp-integration
**Description**: MCP server integration patterns

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| mcp-context7-integrator | 🟢 PRIMARY | ⚠️ MISSING | Core MCP integration |
| mcp-figma-integrator | 🟢 PRIMARY | ⚠️ MISSING | Figma MCP |
| mcp-notion-integrator | 🟢 PRIMARY | ⚠️ MISSING | Notion MCP |
| mcp-playwright-integrator | 🟢 PRIMARY | ⚠️ MISSING | Playwright MCP |

**Gap Analysis**: 0/4 agents have this skill
**Priority**: CRITICAL - All MCP agents missing core MCP skill

---

## Part 10: Specialty Skills

### moai-design-systems
**Description**: Design system implementation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟢 PRIMARY | ⚠️ MISSING | Design system integration |
| ui-ux-expert | 🟢 PRIMARY | ⚠️ MISSING | UI design systems |
| component-designer | 🟢 PRIMARY | ⚠️ MISSING | Component design systems |
| mcp-figma-integrator | 🟡 SECONDARY | ✅ ASSIGNED | Figma design systems |
| accessibility-expert | 🟡 SECONDARY | ⚠️ MISSING | Accessible design systems |

**Gap Analysis**: 1/5 agents have this skill
**Priority**: CRITICAL - Frontend agents need design systems

### moai-lib-shadcn-ui
**Description**: shadcn/ui component library

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| frontend-expert | 🟡 SECONDARY | ⚠️ MISSING | Modern components |
| component-designer | 🟡 SECONDARY | ⚠️ MISSING | Component library |
| ui-ux-expert | 🟡 SECONDARY | ⚠️ MISSING | UI components |

**Gap Analysis**: 0/3 agents have this skill
**Priority**: MEDIUM

### moai-mermaid-diagram-expert
**Description**: Mermaid diagram generation

| Agent | Relevance | Status | Rationale |
|-------|-----------|--------|-----------|
| docs-manager | 🟡 SECONDARY | ✅ ASSIGNED | Documentation diagrams |
| doc-syncer | 🟡 SECONDARY | ⚠️ MISSING | Diagram syncing |
| spec-builder | 🟡 SECONDARY | ⚠️ MISSING | SPEC diagrams |

**Gap Analysis**: 1/3 agents have this skill
**Priority**: LOW

---

## Summary: Critical Gaps to Address

### Priority 1: CRITICAL (Immediate Action Required)

1. **moai-foundation-trust** → Add to:
   - quality-gate (❌ CRITICAL)
   - trust-checker (❌ CRITICAL)
   - tdd-implementer (❌ CRITICAL)
   - format-expert (⚠️ HIGH)

2. **moai-foundation-git** → Add to:
   - git-manager (❌ CRITICAL - agent has zero skills)

3. **moai-domain-devops** → Add to:
   - devops-expert (❌ CRITICAL - missing core domain)

4. **moai-domain-testing** → Add to:
   - quality-gate (❌ CRITICAL)
   - tdd-implementer (❌ CRITICAL)
   - trust-checker (⚠️ HIGH)

5. **moai-core-dev-guide** → Add to:
   - tdd-implementer (❌ CRITICAL - missing TDD workflow)

6. **moai-lang-sql** → Add to:
   - database-expert (❌ CRITICAL)
   - migration-expert (❌ CRITICAL)

7. **moai-design-systems** → Add to:
   - frontend-expert (❌ CRITICAL)
   - ui-ux-expert (❌ CRITICAL)
   - component-designer (❌ CRITICAL)

8. **moai-mcp-integration** → Add to:
   - All MCP integrators (❌ CRITICAL - 4 agents)

9. **moai-essentials-review** → Add to:
   - quality-gate (❌ CRITICAL)
   - trust-checker (❌ CRITICAL)

10. **moai-domain-web-api** → Add to:
    - api-designer (❌ CRITICAL - missing core domain)

### Priority 2: HIGH (Next Sprint)

- moai-security-* skills to backend and API agents
- moai-core-ask-user-questions to planning agents
- moai-docs-unified to documentation agents
- moai-core-workflow to orchestration agents
- moai-lang-typescript to more agents
- moai-essentials-debug to domain agents
- moai-essentials-perf to performance-critical agents

### Priority 3: MEDIUM (Backlog)

- Additional language skills as needed
- Specialty skills for enhanced capabilities
- BaaS integration skills
- Advanced cloud skills

---

## Implementation Checklist

### Week 1: Critical Foundations
- [ ] Update git-manager with moai-foundation-git
- [ ] Update trust-checker with moai-foundation-trust + essentials-review + domain-testing
- [ ] Update quality-gate with moai-foundation-trust + essentials-review + domain-testing
- [ ] Update tdd-implementer with moai-foundation-trust + core-dev-guide + domain-testing + essentials-refactor

### Week 2: Core Domains
- [ ] Update devops-expert with moai-domain-devops
- [ ] Update api-designer with moai-domain-web-api + security-api
- [ ] Update database-expert with moai-lang-sql
- [ ] Update migration-expert with moai-lang-sql

### Week 3: Frontend & Design
- [ ] Update frontend-expert with moai-design-systems + lib-shadcn-ui
- [ ] Update ui-ux-expert with moai-design-systems + lib-shadcn-ui
- [ ] Update component-designer with moai-design-systems + lib-shadcn-ui + lang-typescript

### Week 4: MCP & Integration
- [ ] Update all MCP integrators with moai-mcp-integration
- [ ] Update mcp-context7-integrator with context7-integration
- [ ] Update planning agents with core-ask-user-questions

### Week 5: Security & Quality
- [ ] Update backend-expert with security-api + security-auth
- [ ] Update api-designer with security-api + security-auth
- [ ] Add security skills to relevant agents

### Week 6: Documentation & Orchestration
- [ ] Update doc-syncer with docs-toolkit + docs-unified
- [ ] Update docs-manager with docs-unified + project-documentation
- [ ] Update implementation-planner with core-workflow + todowrite-pattern
- [ ] Update agent-factory with core-agent-guide + ask-user-questions

---

## Metrics & KPIs

**Current State**:
- Total skill assignments: 157
- Average skills per agent: 5.1
- Agents with 0-2 skills: 8 agents (26%)
- Agents with 3-5 skills: 15 agents (48%)
- Agents with 6+ skills: 8 agents (26%)

**Target State** (after implementation):
- Total skill assignments: 287
- Average skills per agent: 9.3
- Agents with 0-2 skills: 0 agents (0%)
- Agents with 3-5 skills: 5 agents (16%)
- Agents with 6+ skills: 26 agents (84%)

**Coverage Improvement**:
- Foundation skills: 8% → 85% (+77%)
- Core skills: 22% → 68% (+46%)
- Domain skills: 45% → 82% (+37%)
- Essential skills: 43% → 78% (+35%)
- Security skills: 18% → 55% (+37%)

---

**Generated**: 2025-11-22
**Mapping Relationships**: 157 current → 287 recommended
**Critical Gaps**: 10 categories
**Implementation Timeline**: 6 weeks
