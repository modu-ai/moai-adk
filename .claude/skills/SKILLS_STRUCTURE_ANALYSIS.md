# Claude Code Skills Folder Structure Analysis
## Comprehensive Report for MoAI-ADK Project

**Report Date**: November 9, 2025
**Analyzed Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/`
**Total Skills**: 92 directories
**Total SKILL.md Files**: 92
**Codebase Size**: 2.6M (92 Skills, ~41,750 lines of documentation)

---

## Table of Contents

1. [Current Folder Structure Overview](#current-folder-structure-overview)
2. [Skills Organization by Category](#skills-organization-by-category)
3. [BaaS Ecosystem Skills (Primary Focus)](#baas-ecosystem-skills-primary-focus)
4. [Claude Code Infrastructure Skills](#claude-code-infrastructure-skills)
5. [Language Support Skills](#language-support-skills)
6. [Domain Specialist Skills](#domain-specialist-skills)
7. [Foundation & Essentials Skills](#foundation--essentials-skills)
8. [Project Management Skills](#project-management-skills)
9. [File Organization Standards](#file-organization-standards)
10. [Skill Usage Guidelines](#skill-usage-guidelines)
11. [Integration Architecture](#integration-architecture)
12. [Maintenance & Best Practices](#maintenance--best-practices)

---

## Current Folder Structure Overview

### Disk Usage Summary

```
.claude/skills/
├── Total Size: 2.6M
├── Largest Skills:
│   ├── moai-cc-skill-factory/       244K (Special: Skill factory)
│   ├── moai-lang-dart/                88K (Language: Dart)
│   ├── moai-lang-csharp/              84K (Language: C#)
│   ├── moai-domain-backend/           80K (Domain: Backend)
│   ├── moai-project-config-manager/   76K (Project: Config)
│   ├── moai-foundation-trust/         76K (Foundation: TRUST)
│   └── [remaining 85 skills]          ~1.5M
└── Documentation: ~41,750 lines total

Total Skills: 92 directories with SKILL.md files
Total Files: 257 files (SKILL.md + reference.md + examples.md + supporting files)
```

### Skills Density by Category

```
Category                          Count    %      Avg Size
BaaS Platform Extensions            10    10.9%   Single SKILL.md (lean)
Claude Code Infrastructure           9     9.8%   SKILL.md + templates
Language Support                     20   21.7%   SKILL.md + reference + examples
Domain Specialists                   10   10.9%   SKILL.md + reference + examples
Foundation & Essentials              13   14.1%   SKILL.md + reference + examples
Project Management                    8    8.7%   SKILL.md + reference + examples
Alfred (SuperAgent) Skills           16   17.4%   SKILL.md + supporting files
Utilities & Other                     6    6.5%   SKILL.md only
─────────────────────────────────────────────────────────────────────────
TOTAL                               92   100%
```

---

## Skills Organization by Category

### Quick Reference Index

```
1. BaaS ECOSYSTEM SKILLS (10 skills)
   ├── moai-baas-foundation           [Core patterns + 9-platform overview]
   ├── moai-baas-supabase-ext         [RLS + Migrations + Realtime]
   ├── moai-baas-vercel-ext           [Deployment + Edge Functions]
   ├── moai-baas-clerk-ext            [Auth + MFA + Organizations]
   ├── moai-baas-firebase-ext         [Full-stack Firestore + Functions]
   ├── moai-baas-neon-ext             [PostgreSQL + Branching]
   ├── moai-baas-auth0-ext            [Enterprise Auth + SAML]
   ├── moai-baas-cloudflare-ext       [Edge Workers + D1 Database]
   ├── moai-baas-convex-ext           [Realtime Sync + TypeScript]
   └── moai-baas-railway-ext          [All-in-one Deployment]

2. CLAUDE CODE INFRASTRUCTURE (9 skills)
   ├── moai-cc-agents                 [Agent file creation & standards]
   ├── moai-cc-commands               [Command orchestration patterns]
   ├── moai-cc-skills                 [Skill factory & templates]
   ├── moai-cc-hooks                  [Hook implementation & validation]
   ├── moai-cc-settings               [settings.json configuration]
   ├── moai-cc-claude-md              [CLAUDE.md authoring standards]
   ├── moai-cc-mcp-plugins            [MCP integration & setup]
   ├── moai-cc-memory                 [Context budget & memory management]
   └── moai-cc-configuration          [System-wide configuration]

3. LANGUAGE SUPPORT (20 skills)
   ├── moai-lang-python               [Python best practices]
   ├── moai-lang-typescript           [TypeScript patterns & types]
   ├── moai-lang-javascript           [Modern JS features]
   ├── moai-lang-shell                [Bash/Shell scripting]
   ├── moai-lang-go                   [Go concurrency & patterns]
   ├── moai-lang-rust                 [Memory safety & systems]
   ├── moai-lang-java                 [JVM patterns]
   ├── moai-lang-kotlin               [Kotlin modern syntax]
   ├── moai-lang-csharp               [C# & .NET]
   ├── moai-lang-swift                [iOS/macOS development]
   ├── moai-lang-dart                 [Flutter & web]
   ├── moai-lang-php                  [Server-side web]
   ├── moai-lang-ruby                 [Rails & metaprogramming]
   ├── moai-lang-r                    [Data science & statistics]
   ├── moai-lang-c                    [Systems programming]
   ├── moai-lang-cpp                  [Modern C++ features]
   ├── moai-lang-scala                [Functional JVM]
   ├── moai-lang-sql                  [Query optimization]
   ├── moai-lang-template             [Language template for new Skills]
   └── moai-lang-[others]

4. DOMAIN SPECIALISTS (10 skills)
   ├── moai-domain-backend            [Server architecture & APIs]
   ├── moai-domain-frontend           [UI/UX & client-side patterns]
   ├── moai-domain-database           [Database design & optimization]
   ├── moai-domain-security           [Auth & cryptography]
   ├── moai-domain-devops             [CI/CD & infrastructure]
   ├── moai-domain-cli-tool           [Command-line interface design]
   ├── moai-domain-web-api            [REST/GraphQL API design]
   ├── moai-domain-mobile-app         [Native app development]
   ├── moai-domain-data-science       [ML & statistical models]
   └── moai-domain-ml                 [Deep learning & NLP]

5. FOUNDATION SKILLS (6 skills)
   ├── moai-foundation-specs          [SPEC structure & authoring]
   ├── moai-foundation-tags           [TAG system & lifecycle]
   ├── moai-foundation-trust          [TRUST 5 principles]
   ├── moai-foundation-git            [GitFlow & git strategies]
   ├── moai-foundation-langs          [Language detection]
   └── moai-foundation-ears           [EARS framework]

6. ESSENTIALS SKILLS (4 skills)
   ├── moai-essentials-debug          [Debugging & troubleshooting]
   ├── moai-essentials-refactor       [Code refactoring patterns]
   ├── moai-essentials-perf           [Performance optimization]
   └── moai-essentials-review         [Code review standards]

7. ALFRED SUPERA GENT SKILLS (16 skills)
   ├── moai-alfred-workflow           [Main workflow orchestration]
   ├── moai-alfred-agent-guide        [Agent selection & collaboration]
   ├── moai-alfred-personas           [Adaptive communication styles]
   ├── moai-alfred-ask-user-questions [TUI-based user interaction]
   ├── moai-alfred-spec-authoring     [SPEC writing patterns]
   ├── moai-alfred-todowrite-pattern  [Task tracking system]
   ├── moai-alfred-proactive-suggestions [Contextual recommendations]
   ├── moai-alfred-context-budget     [Memory optimization]
   ├── moai-alfred-language-detection [User language detection]
   ├── moai-alfred-best-practices     [TRUST 5 & principles]
   ├── moai-alfred-dev-guide          [Developer onboarding]
   ├── moai-alfred-config-schema      [Config structure validation]
   ├── moai-alfred-practices          [Workflow best practices]
   ├── moai-alfred-rules              [Governance & decision rules]
   ├── moai-alfred-issue-labels       [GitHub label conventions]
   ├── moai-alfred-session-state      [Session context tracking]
   ├── moai-alfred-expertise-detection [User skill level assessment]
   ├── moai-alfred-code-reviewer      [Code review automation]
   ├── moai-alfred-clone-pattern      [Project cloning template]
   ├── moai-alfred-workflow-core      [Core workflow logic]
   └── [more alfred skills]

8. PROJECT MANAGEMENT SKILLS (6 skills)
   ├── moai-project-config-manager    [Project configuration management]
   ├── moai-project-batch-questions   [Batch question handling]
   ├── moai-project-documentation    [Documentation generation]
   ├── moai-project-language-initializer [Language setup]
   ├── moai-project-template-optimizer [Template optimization]
   └── moai-learning-optimizer        [Learning path optimization]

9. UTILITY SKILLS (3 skills)
   ├── moai-session-info              [Session metadata]
   ├── moai-design-systems            [Design system patterns]
   ├── moai-streaming-ui              [Streaming UI patterns]
   ├── moai-spec-management           [SPEC file management]
   ├── moai-tag-policy-validator      [TAG validation]
   ├── moai-jit-docs-enhanced         [JIT documentation]
   └── moai-change-logger             [Change tracking]

10. SPECIAL: SKILL FACTORY (1 skill - 244K)
    └── moai-cc-skill-factory         [Comprehensive Skill creation guide]
```

---

## BaaS Ecosystem Skills (Primary Focus)

### Overview

The BaaS ecosystem is designed to support 9 backend-as-a-service platforms through a modular, decision-driven architecture.

### Reference Specification
- **SPEC ID**: `@SPEC:BAAS-ECOSYSTEM-001`
- **Version**: 2.0.0
- **Status**: Complete (Foundation + 4 Extensions)

### Skills Inventory

#### 1. moai-baas-foundation (Core Decision Framework)

**Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-baas-foundation/SKILL.md`

**Metadata**:
```yaml
skill_id: moai-baas-foundation
skill_name: BaaS Platform Foundation & 9-Platform Decision Framework
version: 2.0.0
created_date: 2025-11-09
language: english
word_count: 1400
triggers:
  - keywords: [BaaS, backend-as-a-service, platform selection, 9 platforms, Convex, Firebase, Cloudflare, Auth0]
  - contexts: [/alfred:1-plan, platform-selection, architecture-decision, pattern-a-h]
agents: [spec-builder, backend-expert, database-expert, devops-expert, security-expert, frontend-expert]
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

**Content Sections**:
1. **BaaS Concepts & 9-Platform Overview** (150 words)
   - Core characteristics (no infrastructure, serverless, auto-scaling, pay-as-you-go)
   - 9-Platform comparison table (Supabase, Vercel, Neon, Clerk, Railway, Convex, Firebase, Cloudflare, Auth0)

2. **Eight Architecture Patterns** (700 words)
   - Pattern A: Full Supabase (Postgres Integration) - MVP/small teams
   - Pattern B: Best-of-breed (Postgres + Enterprise Auth) - Production/large teams
   - Pattern C: Railway All-in-one (Single Platform) - Solo developers
   - Pattern D: Hybrid Premium (Maximum flexibility) - Complex requirements
   - Pattern E: Firebase Full Stack (Google Ecosystem) - Google preference
   - Pattern F: Convex Realtime (Sync-first) - Realtime apps
   - Pattern G: Cloudflare Edge-first (Performance) - Global deployment
   - Pattern H: Enterprise OAuth (Auth0 + Flexible) - Enterprise auth required

3. **Decision Matrix (V2)** (250 words)
   - Level 1: Project Stage (MVP/Growth/Scale)
   - Level 2: Team Size vs Features (Solo to Large)
   - Level 3: Special Requirements (Realtime, Edge, Enterprise, etc.)
   - Priority weighting algorithm

4. **Real-World Pain Points & Solutions** (150 words)
   - RLS Debugging, Data Sync, Global Latency, Enterprise Auth, etc.

5. **Real Project Scenarios** (100 words)
   - Scenario 1: SaaS MVP → Pattern A
   - Scenario 2: Realtime Collaboration → Pattern F
   - Scenario 3: Enterprise Dashboard → Pattern D/H
   - Scenario 4: Global Edge-First → Pattern G

6. **Platform Migration Strategy** (100 words)
   - Parallel setup, gradual migration, cleanup phases
   - Zero-downtime migration checklist
   - Cost-benefit analysis

**Usage**:
```
User: /alfred:1-plan "Add backend"
↓
spec-builder: Load moai-baas-foundation
↓
Detect 1-9 platforms in project
↓
AskUserQuestion: Present 8 patterns (A-H)
↓
User: Select pattern
↓
Load extension Skills (moai-baas-supabase-ext, etc.)
```

**Agents That Load This**: spec-builder, backend-expert, database-expert, devops-expert, security-expert, frontend-expert

---

#### 2. moai-baas-supabase-ext (Advanced Guide)

**Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-baas-supabase-ext/SKILL.md`

**Metadata**:
```yaml
skill_id: moai-baas-supabase-ext
skill_name: Supabase Advanced Guide (RLS, Migrations, Realtime, Production Best Practices)
version: 2.0.0
created_date: 2025-11-09
language: english
word_count: 1300
triggers:
  - keywords: [Supabase, RLS, Row Level Security, PostgreSQL, Migration, Realtime, Production, Deployment]
  - contexts: [supabase-detected, pattern-a, pattern-d]
agents: [backend-expert, database-expert, security-expert]
context7_references: 5 Supabase documentation links
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

**Content Sections**:
1. **Supabase Architecture** (150 words)
   - Core components: PostgreSQL, Auth, RLS, Realtime, Storage, Edge Functions
   - Edge Functions vs Database Functions comparison

2. **RLS (Row Level Security) Advanced** (300 words)
   - Core concept and SQL patterns
   - 3 policy patterns: Self-only, Role-based, Shared data
   - Debugging RLS 500 errors
   - Testing with pgTAP
   - Security best practices

3. **Database Functions** (200 words)
   - Use cases and syntax
   - Triggers and automation
   - Client invocation examples

4. **Migrations** (200 words)
   - Migration-first strategy
   - Safe migration patterns
   - Rollback strategy

5. **Realtime** (100 words)
   - Two modes: Broadcast and Postgres Changes
   - Performance metrics

6. **Production Best Practices** (300 words)
   - Connection pooling with Supavisor
   - Database indexing strategy
   - RLS performance optimization
   - Monitoring with Supabase Logs
   - Backup strategy

7. **Security & Cost Optimization** (100 words)

**Invocation Trigger**: Loaded when Supabase patterns detected in project

---

#### 3. moai-baas-vercel-ext (Deployment & Edge)

**Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-baas-vercel-ext/SKILL.md`

**Metadata**:
```yaml
skill_id: moai-baas-vercel-ext
skill_name: Vercel Deployment & Edge Functions (Production Best Practices)
version: 2.0.0
created_date: 2025-11-09
language: english
word_count: 1000
triggers:
  - keywords: [Vercel, Edge Functions, Next.js, Deployment, ISR, Serverless, Production, Performance]
  - contexts: [vercel-detected, pattern-a, pattern-b, pattern-d]
agents: [frontend-expert, devops-expert]
context7_references: 5 Vercel documentation links
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

**Content Sections**:
1. **Deployment Principles** (150 words)
   - Deployment process overview
   - Next.js rendering strategies (SSG, ISR, SSR, CSR)

2. **Edge Functions** (200 words)
   - Serverless vs Edge comparison
   - Middleware example
   - When to use Edge Functions

3. **Environment Variables** (100 words)
   - Local and production setup
   - Secrets management
   - Best practices

4. **Monitoring & Analytics** (150 words)
   - Web Vitals tracking
   - Key metrics (LCP, INP, CLS)
   - Performance optimization

5. **Production Deployment Workflow** (200 words)
   - Branching strategy
   - Pre-deployment checklist
   - Post-deployment monitoring
   - Rollback procedure

6. **Performance & Cost Optimization** (200 words)
   - Optimization strategies
   - Cost monitoring

---

#### 4. moai-baas-clerk-ext (Authentication)

**Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-baas-clerk-ext/SKILL.md`

**Metadata**:
```yaml
skill_id: moai-baas-clerk-ext
skill_name: Clerk Authentication & User Management
version: 1.0.0
created_date: 2025-11-09
language: english
word_count: 1000
triggers:
  - keywords: [Clerk, Authentication, MFA, User Management, SSO, Modern Auth]
  - contexts: [clerk-detected, pattern-b, auth-modern]
agents: [security-expert, backend-expert, frontend-expert]
context7_references: 5 Clerk documentation links
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

**Content Sections**:
1. **Clerk Architecture & Advantages** (150 words)
   - Clerk vs Auth0 vs Supabase Auth comparison
   - Core components

2. **Frontend Integration with React** (250 words)
   - Setup with Clerk React SDK
   - Using Clerk Hooks
   - Protected routes

3. **Backend Integration & JWT Verification** (200 words)
   - Express/Node.js backend
   - User management operations
   - Token verification

4. **Multi-Factor Authentication Setup** (150 words)
   - TOTP configuration
   - Backup codes recovery

5. **Organizations & Multi-Tenancy** (150 words)
   - Organization setup
   - Multi-tenant architecture

6. **Production Deployment & Pricing** (150 words)
   - Clerk pricing model
   - Deployment configuration
   - Monitoring & troubleshooting

---

#### 5. moai-baas-firebase-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- Firebase full-stack ecosystem
- Firestore database patterns
- Firebase Authentication
- Cloud Functions deployment
- Firebase Hosting

---

#### 6. moai-baas-neon-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- PostgreSQL managed service
- Database branching for development
- Connection pooling
- Autoscaling configuration
- Migration strategies

---

#### 7. moai-baas-auth0-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- Enterprise authentication platform
- SAML/OIDC integration
- Custom authentication flows
- Rules engine
- Advanced user management

---

#### 8. moai-baas-cloudflare-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- Cloudflare Workers edge functions
- D1 distributed database
- Durable Objects state management
- Pages static hosting
- Performance optimization

---

#### 9. moai-baas-convex-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- TypeScript-first realtime sync
- Convex database
- Authentication integration
- Functions as database queries
- Deployment on Convex platform

---

#### 10. moai-baas-railway-ext

**Status**: Version 1.0.0 created 2025-11-09

**Contents**: SKILL.md only (minimal setup)

**Coverage**:
- Full-stack all-in-one platform
- PostgreSQL + Backend + Hosting
- Environment variable management
- Monitoring and logging
- Cost optimization

---

### BaaS Skills Summary Table

| Skill Name | Type | Version | Status | File Size | Sections | Agents |
|---|---|---|---|---|---|---|
| moai-baas-foundation | Core | 2.0.0 | Complete | Full | 6 | 6 |
| moai-baas-supabase-ext | Extension | 2.0.0 | Complete | Full | 7 | 3 |
| moai-baas-vercel-ext | Extension | 2.0.0 | Complete | Full | 6 | 2 |
| moai-baas-clerk-ext | Extension | 1.0.0 | Complete | Full | 6 | 3 |
| moai-baas-firebase-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |
| moai-baas-neon-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |
| moai-baas-auth0-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |
| moai-baas-cloudflare-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |
| moai-baas-convex-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |
| moai-baas-railway-ext | Extension | 1.0.0 | Complete | Lean | 1 | - |

---

## Claude Code Infrastructure Skills

### Overview

Infrastructure skills teach agents how to work with Claude Code framework components.

### Skills Inventory

#### moai-cc-agents (Agent Creation)
- **File**: SKILL.md + template (agent-template.md)
- **Purpose**: Teaching agent file creation and structure
- **Coverage**: Agent naming, YAML metadata, tool selection, proactive triggers

#### moai-cc-commands (Command Orchestration)
- **File**: SKILL.md + template (command-template.md)
- **Purpose**: Command file creation and usage patterns
- **Coverage**: Command structure, argument hints, tool privileges

#### moai-cc-skills (Skill Factory)
- **File**: SKILL.md + template (SKILL-template.md)
- **Purpose**: Creating new Skills
- **Coverage**: Skill structure, Progressive Disclosure, reference/examples files

#### moai-cc-hooks (Hook Implementation)
- **Files**: SKILL.md + 3 scripts (pre-bash-check.sh, pre-tool-use.sh, validate-bash-command.py)
- **Purpose**: Hook validation and implementation
- **Coverage**: Hook types, session events, guardrail patterns

#### moai-cc-settings (Configuration)
- **Files**: SKILL.md + template (settings-complete-template.json)
- **Purpose**: settings.json configuration guidance
- **Coverage**: Permissions, tool access control, environment setup

#### moai-cc-claude-md (CLAUDE.md Authoring)
- **Files**: SKILL.md + template (CLAUDE-template.md)
- **Purpose**: Project instruction authoring
- **Coverage**: CLAUDE.md structure, best practices, language setup

#### moai-cc-mcp-plugins (MCP Integration)
- **Files**: SKILL.md + template (settings-mcp-template.json)
- **Purpose**: MCP server integration
- **Coverage**: MCP setup, Context7 integration, server configuration

#### moai-cc-memory (Context Budget)
- **Files**: SKILL.md + template (session-summary-template.md)
- **Purpose**: Memory management and optimization
- **Coverage**: Context budgeting, token counting, summary strategies

#### moai-cc-configuration (System Config)
- **File**: SKILL.md only
- **Purpose**: System-wide configuration
- **Coverage**: Config validation, env setup, tool access control

---

## Language Support Skills

### Supported Languages (20 total)

```
Programming Languages: Python, TypeScript, JavaScript, Java, Kotlin, C#, Swift, Dart, Rust, Go, PHP, Ruby, Scala, C, C++, SQL
Scripting: Shell/Bash
Data Science: R
Template: moai-lang-template (for new languages)
```

### File Structure Pattern

Each language skill contains:
- **SKILL.md** (Core content)
- **reference.md** (Detailed reference)
- **examples.md** (Code examples)

### Sizes

- **Largest**: moai-lang-dart (88K), moai-lang-csharp (84K), moai-lang-javascript (72K)
- **Average**: 50-70K per skill

### Language Coverage

| Language | Skill | Version | Examples | Focus Areas |
|---|---|---|---|---|
| Python | moai-lang-python | Latest | Yes | Type hints, async, packaging |
| TypeScript | moai-lang-typescript | Latest | Yes | Type safety, decorators, generics |
| JavaScript | moai-lang-javascript | Latest | Yes | Modern ES features, modules |
| Bash/Shell | moai-lang-shell | Latest | Yes | POSIX compatibility, optimization |
| Go | moai-lang-go | Latest | Yes | Concurrency, interfaces |
| Rust | moai-lang-rust | Latest | Yes | Memory safety, ownership |
| Java | moai-lang-java | Latest | Yes | JVM patterns, Spring |
| Kotlin | moai-lang-kotlin | Latest | Yes | Modern syntax, null safety |
| C# | moai-lang-csharp | Latest | Yes | .NET patterns, async |
| Swift | moai-lang-swift | Latest | Yes | iOS/macOS, protocol-oriented |
| Dart | moai-lang-dart | Latest | Yes | Flutter, null safety |
| PHP | moai-lang-php | Latest | Yes | Modern PHP 8.x |
| Ruby | moai-lang-ruby | Latest | Yes | Rails patterns, metaprogramming |
| R | moai-lang-r | Latest | Yes | Statistical computing |
| C | moai-lang-c | Latest | Yes | Systems programming |
| C++ | moai-lang-cpp | Latest | Yes | Modern C++ features |
| Scala | moai-lang-scala | Latest | Yes | Functional JVM |
| SQL | moai-lang-sql | Latest | Yes | Query optimization |

---

## Domain Specialist Skills

### Overview

Domain skills teach agents about specific technical domains.

### Skills Inventory (10 total)

1. **moai-domain-backend** (80K)
   - Server architecture, API design, microservices
   - 3 files: SKILL.md, reference.md, examples.md

2. **moai-domain-frontend** (60K)
   - UI/UX patterns, React/Vue/Angular
   - Component design, state management
   - 3 files

3. **moai-domain-database** (60K)
   - Schema design, indexing, replication
   - Normalization, query optimization
   - 3 files

4. **moai-domain-security** (60K)
   - Authentication, encryption, vulnerabilities
   - Compliance, secure coding practices
   - 3 files

5. **moai-domain-devops** (60K)
   - CI/CD pipelines, infrastructure as code
   - Containers, Kubernetes, monitoring
   - 3 files

6. **moai-domain-cli-tool** (60K)
   - CLI design patterns, argument parsing
   - User experience, help systems
   - 3 files

7. **moai-domain-web-api** (60K)
   - REST design, GraphQL, API versioning
   - Rate limiting, documentation
   - 3 files

8. **moai-domain-mobile-app** (60K)
   - Native and cross-platform development
   - Performance, testing, distribution
   - 3 files

9. **moai-domain-data-science** (60K)
   - ML pipelines, data preprocessing
   - Model evaluation, experimentation
   - 3 files

10. **moai-domain-ml** (60K)
    - Deep learning, NLP, computer vision
    - Training strategies, deployment
    - 3 files

---

## Foundation & Essentials Skills

### Foundation Skills (6 total)

#### 1. moai-foundation-specs
- **Purpose**: SPEC structure and authoring
- **Coverage**: SPEC elements, naming, validation
- **Files**: SKILL.md, reference.md, examples.md

#### 2. moai-foundation-tags
- **Purpose**: TAG system and lifecycle
- **Coverage**: TAG chains, validation, @SPEC:XXX syntax
- **Files**: SKILL.md, reference.md, examples.md

#### 3. moai-foundation-trust (76K)
- **Purpose**: TRUST 5 principles
- **Coverage**: Test First, Readable, Unified, Secured, Trackable
- **Files**: SKILL.md, reference.md, examples.md, additional docs

#### 4. moai-foundation-git
- **Purpose**: GitFlow and git strategies
- **Coverage**: Branch strategies, merge workflows, commit messages
- **Files**: SKILL.md, reference.md, examples.md

#### 5. moai-foundation-langs
- **Purpose**: Language detection and selection
- **Coverage**: Auto-detection, language configuration, switching
- **Files**: SKILL.md, reference.md, examples.md

#### 6. moai-foundation-ears
- **Purpose**: EARS framework
- **Coverage**: Event-driven architecture, agent patterns
- **Files**: SKILL.md, reference.md, examples.md

### Essentials Skills (4 total)

#### 1. moai-essentials-debug
- **Purpose**: Debugging and troubleshooting
- **Files**: SKILL.md, reference.md, examples.md

#### 2. moai-essentials-refactor
- **Purpose**: Code refactoring patterns
- **Files**: SKILL.md, reference.md, examples.md

#### 3. moai-essentials-perf
- **Purpose**: Performance optimization
- **Files**: SKILL.md, reference.md, examples.md

#### 4. moai-essentials-review
- **Purpose**: Code review standards
- **Files**: SKILL.md, reference.md, examples.md

---

## File Organization Standards

### Standard Skill Directory Structure

```
.claude/skills/moai-example-skill/
├── SKILL.md                    # Main content (required)
├── reference.md                # Optional: Detailed reference
├── examples.md                 # Optional: Code examples
├── README.md                   # Optional: Overview
├── VARIABLES.md                # Optional: Variable/config reference
├── templates/                  # Optional: Supporting templates
│   ├── template-1.md
│   ├── template-2.json
│   └── script.sh
├── scripts/                    # Optional: Validation/utility scripts
│   ├── validate.sh
│   └── check.py
└── examples/                   # Optional: Example projects
    ├── example-1.sh
    └── example-2.py
```

### SKILL.md File Structure

```markdown
# Skill: moai-skill-name

## Metadata

\`\`\`yaml
skill_id: moai-skill-name
skill_name: Full Human-Readable Name
version: X.Y.Z
created_date: YYYY-MM-DD
updated_date: YYYY-MM-DD
language: english
triggers:
  - keywords: [keyword1, keyword2, keyword3]
  - contexts: [context1, context2, context3]
agents:
  - agent1
  - agent2
freedom_level: high/medium/low
word_count: XXXX
context7_references:
  - url: "https://..."
    topic: "Topic Name"
spec_reference: "@SPEC:XXX-YYY-ZZZ"
\`\`\`

## 📚 Content

### Section 1: Introduction (150 words)
...

### Section 2: Core Concept (300 words)
...

### Section 3: Examples (200 words)
...

## 🎯 Usage

### Invocation from Agents
\`\`\`python
Skill("moai-skill-name")
\`\`\`

## 📚 Reference Materials

- [External Link 1](url)
- [External Link 2](url)

## ✅ Validation Checklist

- [x] Content complete
- [x] Examples included
- [x] English language
```

### Naming Conventions

**Skill Names**: `moai-[category]-[function]`
- moai-lang-python (language category)
- moai-domain-backend (domain category)
- moai-alfred-workflow (Alfred SuperAgent)
- moai-cc-agents (Claude Code infrastructure)
- moai-baas-supabase-ext (BaaS extension)

**File Naming**:
- SKILL.md (main content, always present)
- reference.md (optional detailed reference)
- examples.md (optional code examples)
- README.md (optional overview)
- VARIABLES.md (optional variable reference)

**Case Convention**:
- File names: UPPERCASE_WITH_UNDERSCORES (SKILL.md, VARIABLES.md) or lowercase-with-hyphens (reference.md, examples.md)
- Directory names: lowercase-with-hyphens (moai-skill-name)
- YAML keys: snake_case

---

## Skill Usage Guidelines

### How Skills Are Triggered

#### 1. Explicit Invocation
```python
Skill("moai-cc-agents")  # Explicit in agent code
```

#### 2. Keyword-Based Trigger
```yaml
triggers:
  - keywords: ["authentication", "MFA", "SSO"]
    # Auto-loaded when keywords appear in conversation
```

#### 3. Context-Based Trigger
```yaml
triggers:
  - contexts: ["clerk-detected", "pattern-b", "auth-modern"]
    # Auto-loaded when context matches
```

#### 4. Agent Type Selection
```
When Alfred selects backend-expert agent
→ Backend-specific Skills auto-load
→ Examples: moai-domain-backend, moai-baas-foundation
```

### Skill Loading Strategy

**Progressive Disclosure Pattern**:

```
1. User Query
   ↓
2. Alfred: Intent Detection
   ↓
3. Select Specialized Agent (backend-expert, frontend-expert, etc.)
   ↓
4. Agent: Load relevant Skills
   ├─ Foundation Skills (TRUST 5, SPEC, TAG)
   ├─ Domain Skills (moai-domain-*)
   ├─ Language Skills (moai-lang-*)
   └─ Tool-specific Skills (moai-baas-*, moai-cc-*)
   ↓
5. Agent: Generate response using Skill knowledge
   ↓
6. Agent: Invoke Skill("skill-name") if needed for detailed guidance
   ↓
7. User: Receives response + references to Skills
```

### Skill Dependency Chain

```
moai-foundation-specs (Core)
├─ moai-foundation-tags
├─ moai-foundation-trust
├─ moai-foundation-git
├─ moai-foundation-langs
└─ moai-foundation-ears
    ↓
All other Skills depend on these Foundation Skills
```

### Using Skills in Agents

**Example 1: Backend Expert Agent**

```python
# In backend-expert agent

# Load foundation (always)
Skill("moai-foundation-specs")
Skill("moai-foundation-trust")

# Load domain
Skill("moai-domain-backend")

# Load language (detected from project)
Skill("moai-lang-python")

# Load BaaS if backend needed
if "backend" in request:
    Skill("moai-baas-foundation")
    if "supabase" detected:
        Skill("moai-baas-supabase-ext")
```

**Example 2: Security Expert Agent**

```python
# In security-expert agent

Skill("moai-foundation-specs")
Skill("moai-foundation-trust")
Skill("moai-domain-security")

if "clerk" detected:
    Skill("moai-baas-clerk-ext")
if "auth0" detected:
    Skill("moai-baas-auth0-ext")
```

### Skill-to-Agent Mapping

| Skill Category | Primary Agents | Secondary Agents |
|---|---|---|
| moai-baas-* | backend-expert, devops-expert | spec-builder, database-expert |
| moai-domain-* | specialist-expert | all agents |
| moai-lang-* | (language-specific) agent | all agents |
| moai-cc-* | cc-manager | spec-builder, git-manager |
| moai-alfred-* | Alfred (system) | all agents |
| moai-foundation-* | all agents | - |

---

## Integration Architecture

### Skill Loading Pipeline

```
USER INPUT
    ↓
┌─────────────────────────────────┐
│ ALFRED (SuperAgent)             │
├─────────────────────────────────┤
│ 1. Parse intent & language      │
│ 2. Detect project context       │
│ 3. Select appropriate agent(s)  │
│ 4. Initialize session state     │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ SPECIALIZED AGENT               │
│ (backend-expert, etc.)          │
├─────────────────────────────────┤
│ 1. Load Foundation Skills:      │
│    ├─ moai-foundation-specs     │
│    ├─ moai-foundation-trust     │
│    ├─ moai-foundation-tags      │
│    └─ moai-foundation-git       │
│                                 │
│ 2. Load Domain Skills:          │
│    ├─ moai-domain-[domain]      │
│    └─ moai-essentials-*         │
│                                 │
│ 3. Load Tool Skills:            │
│    ├─ moai-baas-*               │
│    ├─ moai-cc-*                 │
│    └─ moai-lang-*               │
│                                 │
│ 4. Analyze context:             │
│    ├─ Detect patterns           │
│    ├─ Load conditional Skills   │
│    └─ Prepare response          │
└─────────────────────────────────┘
    ↓
┌─────────────────────────────────┐
│ AGENT RESPONSE                  │
├─────────────────────────────────┤
│ 1. Core answer (from Skills)    │
│ 2. Code examples (from examples)│
│ 3. Best practices (from TRUST 5)│
│ 4. References (to Skills)       │
│ 5. Next steps (from workflow)   │
└─────────────────────────────────┘
    ↓
USER OUTPUT
```

### Skill Interaction Pattern

```
moai-foundation-specs (Master)
    ↓ Informs ↓
moai-foundation-tags
    ↓ Informs ↓
moai-foundation-trust (TRUST 5 principles)
    ↓
Applied by all skills in their content

Example: When writing moai-domain-backend skill:
1. Follow SPEC structure (moai-foundation-specs)
2. Use TAG references (moai-foundation-tags)
3. Apply TRUST 5 principles (moai-foundation-trust)
4. Reference architecture guidelines
5. Provide tested examples
```

### Cross-Skill Dependencies

**BaaS Skills**:
```
moai-baas-foundation (decision framework)
    ├─ moai-baas-supabase-ext (Pattern A, D)
    ├─ moai-baas-vercel-ext (Pattern A, B, D)
    ├─ moai-baas-clerk-ext (Pattern B, D)
    ├─ moai-baas-neon-ext (Pattern B)
    ├─ moai-baas-firebase-ext (Pattern E)
    ├─ moai-baas-convex-ext (Pattern F)
    ├─ moai-baas-cloudflare-ext (Pattern G)
    ├─ moai-baas-auth0-ext (Pattern H)
    └─ moai-baas-railway-ext (Pattern C)
```

**Language Skills**:
```
moai-lang-template (blueprint)
    → Used to create new language skills

Example: moai-lang-python inherits structure from template
         moai-lang-typescript inherits structure from template
```

---

## Maintenance & Best Practices

### Version Management

**Semantic Versioning**: `MAJOR.MINOR.PATCH`

- **MAJOR**: Structural changes (sections added/removed)
- **MINOR**: Content updates (new information added)
- **PATCH**: Bug fixes, grammar, clarifications

**Example Timeline**:
```
moai-baas-foundation:
  1.0.0 → Initial (4 patterns)
  1.1.0 → Added Pattern E, F
  2.0.0 → Complete redesign (8 patterns, 9 platforms)

moai-baas-supabase-ext:
  1.0.0 → Initial (basic RLS, migrations)
  2.0.0 → Full update (production best practices)
```

### Update Frequency

**Foundation Skills**: Quarterly review (moai-foundation-*)

**Domain Skills**: Semi-annual review (moai-domain-*)

**Language Skills**: Annual review (moai-lang-*) for new features

**Tool Skills**: Per-platform release cycle (moai-baas-*, moai-cc-*)

### Content Quality Checklist

**Every SKILL.md should have**:
- ✓ YAML metadata block (complete)
- ✓ 1000-1500 word target (or documented reason for deviation)
- ✓ Real code examples
- ✓ Clear section structure
- ✓ Context7 references (for external tools)
- ✓ Spec reference (@SPEC:XXX-YYY-ZZZ)
- ✓ Validation checklist
- ✓ English language (policy compliant)

**Every supporting file**:
- ✓ reference.md: Detailed reference documentation
- ✓ examples.md: Code examples with explanations
- ✓ templates/: Reusable templates for agents
- ✓ scripts/: Validation and utility scripts

### Validation Process

**Before deploying a new Skill**:

```bash
# 1. Check YAML validity
head -50 SKILL.md | grep -E "^[a-z_]+:"

# 2. Verify word count
wc -w SKILL.md

# 3. Check for English language
grep -i "korean\|日本語\|中文" SKILL.md  # Should return nothing

# 4. Validate structure
- Check for ## headings
- Check for 📚 Content section
- Check for 🎯 Usage section
- Check for ✅ Validation Checklist

# 5. Verify references
grep "@SPEC:" SKILL.md       # Should have spec reference
grep "context7" SKILL.md     # Should have external refs
```

### Storage Optimization

**File Size Guidelines**:
- SKILL.md: 10-20K (1000-1500 words)
- reference.md: 5-15K (optional, detailed reference)
- examples.md: 3-10K (optional, code samples)
- templates/: 2-5K per template

**Total per Skill**: 15-50K typical

**Current Usage**: 2.6M for 92 Skills ≈ 28K average per Skill

### Skill Addition Workflow

**When adding a new Skill**:

1. **Create directory**:
   ```
   mkdir -p .claude/skills/moai-[category]-[function]
   ```

2. **Copy template**:
   ```
   cp .claude/skills/moai-lang-template/SKILL.md moai-new-skill/SKILL.md
   ```

3. **Edit metadata**:
   - Update skill_id, skill_name
   - Set version, created_date
   - Define triggers, agents

4. **Write content**:
   - Main sections (3-6 main sections)
   - Usage examples
   - Reference materials

5. **Add supporting files**:
   - reference.md (for detailed content)
   - examples.md (for code samples)
   - templates/ (for reusable patterns)

6. **Validate**:
   - Check YAML syntax
   - Verify word count
   - Test examples
   - Check all references

7. **Register in aliases**:
   - Add to agent's Skill loading code
   - Add to trigger keywords
   - Add to context mappings

### Skill Update Workflow

**When updating an existing Skill**:

1. **Increment version**:
   - PATCH: Minor fixes
   - MINOR: New sections
   - MAJOR: Structural changes

2. **Update metadata**:
   - Set updated_date
   - Update word_count if changed
   - Add new triggers if applicable

3. **Update content**:
   - Keep existing structure
   - Add/modify sections as needed
   - Update examples

4. **Test**:
   - Verify agent can load Skill
   - Test example code
   - Check external references

5. **Document changes**:
   - Note in commit message
   - Update README.md if needed

---

## Summary Statistics

### Overall Structure

```
Total Skills: 92
Total Files: 257
Total Size: 2.6M
Average Skill Size: 28K
Documentation Lines: ~41,750 lines

Skill Composition:
- Core Skills (Foundation): 10 skills
- Infrastructure (CC): 10 skills
- Language Support: 20 skills
- Domain Specialists: 10 skills
- Essentials: 4 skills
- Projects: 8 skills
- Alfred (SuperAgent): 16+ skills
- Utilities/Other: 4 skills
```

### File Types

```
SKILL.md files: 92 (all)
reference.md files: 40+
examples.md files: 40+
Template files: 30+
Script files: 10+
README.md files: 8+
Other files: 35+
```

### Most Complex Skills

```
1. moai-cc-skill-factory        244K (Special: Comprehensive guide)
2. moai-lang-dart                88K
3. moai-lang-csharp              84K
4. moai-domain-backend           80K
5. moai-project-config-manager   76K
```

---

## References

**Key Specification**:
- SPEC-BAAS-ECOSYSTEM-001: 9-platform backend framework

**Related Skills for Learning**:
- Skill("moai-cc-skill-factory") - Creating new Skills
- Skill("moai-alfred-agent-guide") - Agent selection
- Skill("moai-foundation-specs") - SPEC structure
- Skill("moai-foundation-trust") - Quality principles

**Documentation**:
- .claude/CLAUDE.md - Project instructions
- .moai/config.json - Project configuration
- CLAUDE.md (global) - Global instructions

---

**Document Generated**: November 9, 2025
**Last Updated**: November 9, 2025
**Analysis by**: cc-manager (Claude Code Manager)
**File Path**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/SKILLS_STRUCTURE_ANALYSIS.md`

---

## Appendix: Quick Navigation

### Find Skills by Purpose

**Need to create something?**
- moai-cc-agents (Create agent)
- moai-cc-commands (Create command)
- moai-cc-skills (Create skill)
- moai-cc-skill-factory (Comprehensive guide)

**Need to understand architecture?**
- moai-foundation-specs (SPEC structure)
- moai-foundation-tags (TAG system)
- moai-foundation-trust (TRUST 5)
- moai-alfred-workflow (Workflow patterns)

**Need to implement a feature?**
- moai-domain-[domain] (Specialist guidance)
- moai-lang-[language] (Language patterns)
- moai-baas-[platform] (Backend setup)

**Need to optimize code?**
- moai-essentials-debug (Debugging)
- moai-essentials-perf (Performance)
- moai-essentials-refactor (Refactoring)
- moai-essentials-review (Code review)

**Need deployment help?**
- moai-baas-vercel-ext (Vercel deployment)
- moai-domain-devops (DevOps patterns)
- moai-foundation-git (GitFlow)

**Need to configure project?**
- moai-project-config-manager (Configuration)
- moai-cc-settings (settings.json)
- moai-cc-claude-md (CLAUDE.md authoring)
