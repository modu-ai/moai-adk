# Claude Code Skills - Quick Reference Guide
## Fast lookup for common questions

**Last Updated**: November 9, 2025
**Total Skills**: 92
**Categories**: 9 main categories

---

## Quick Answers

### "I need to create a new [Thing]. What Skill helps?"

**Create a new Agent**
```
→ Skill("moai-cc-agents")
→ File: .claude/agents/my-agent.md
→ Template: moai-cc-agents/templates/agent-template.md
→ Agents that use this: cc-manager
```

**Create a new Command**
```
→ Skill("moai-cc-commands")
→ File: .claude/commands/my-command.md
→ Template: moai-cc-commands/templates/command-template.md
→ Agents that use this: spec-builder
```

**Create a new Skill**
```
→ Skill("moai-cc-skill-factory")  [comprehensive guide]
→ OR Skill("moai-cc-skills")      [quick guide]
→ File: .claude/skills/moai-my-skill/SKILL.md
→ Template: moai-lang-template/ or moai-cc-skills/SKILL-template.md
```

**Create a new Language Skill**
```
→ Skill("moai-cc-skill-factory")
→ Copy: .claude/skills/moai-lang-template/
→ File: .claude/skills/moai-lang-[name]/SKILL.md
→ Pattern: moai-lang-python, moai-lang-typescript, etc.
```

**Create a SPEC document**
```
→ Skill("moai-foundation-specs")
→ File: .moai/specs/SPEC-[CATEGORY]-[NUMBER].md
→ Structure: Define → Architecture → Implementation → Testing
```

---

### "I'm working with [Technology]. Which Skill?"

**Backend-as-a-Service (BaaS)**
```
→ moai-baas-foundation            (Start here - all 9 platforms)
   ├─ moai-baas-supabase-ext      (PostgreSQL-based)
   ├─ moai-baas-vercel-ext        (Frontend deployment)
   ├─ moai-baas-clerk-ext         (Authentication)
   ├─ moai-baas-firebase-ext      (Google ecosystem)
   ├─ moai-baas-neon-ext          (Database branching)
   ├─ moai-baas-auth0-ext         (Enterprise auth)
   ├─ moai-baas-cloudflare-ext    (Edge computing)
   ├─ moai-baas-convex-ext        (Realtime sync)
   └─ moai-baas-railway-ext       (All-in-one)
```

**Python Development**
```
→ moai-lang-python                (Python patterns)
→ moai-domain-backend             (If building backend)
→ moai-domain-data-science        (If doing ML)
→ moai-essentials-debug           (Debugging)
→ moai-essentials-perf            (Optimization)
```

**TypeScript/JavaScript**
```
→ moai-lang-typescript            (Type safety)
→ moai-lang-javascript            (Modern features)
→ moai-domain-frontend            (UI development)
→ moai-domain-backend             (Node.js backend)
→ moai-baas-vercel-ext            (Deployment)
```

**Database Work**
```
→ moai-domain-database            (Design & optimization)
→ moai-baas-supabase-ext          (If using Supabase)
→ moai-baas-neon-ext              (If using Neon)
→ moai-lang-sql                   (Query patterns)
```

**Authentication**
```
→ moai-domain-security            (Auth concepts)
→ moai-baas-clerk-ext             (Clerk platform)
→ moai-baas-auth0-ext             (Auth0 platform)
→ moai-baas-supabase-ext          (Supabase auth)
```

**DevOps/Deployment**
```
→ moai-domain-devops              (CI/CD, containers)
→ moai-baas-vercel-ext            (Vercel deployment)
→ moai-foundation-git             (GitFlow, branching)
→ moai-project-config-manager     (Configuration)
```

**Security**
```
→ moai-domain-security            (Auth, encryption, vulnerabilities)
→ moai-foundation-trust           (TRUST 5 principles)
→ moai-essentials-review          (Code review standards)
```

**Mobile Development**
```
→ moai-domain-mobile-app          (Native apps)
→ moai-lang-swift                 (iOS)
→ moai-lang-kotlin                (Android)
→ moai-lang-dart                  (Flutter)
```

**API Design**
```
→ moai-domain-web-api             (REST, GraphQL)
→ moai-domain-backend             (Server architecture)
→ moai-essentials-perf            (Performance)
```

---

### "I need to understand [Concept]. What Skill?"

**Architecture Patterns**
```
→ moai-baas-foundation            (8 BaaS patterns)
→ moai-domain-backend             (Backend patterns)
→ moai-domain-frontend            (Frontend patterns)
→ moai-domain-database            (DB patterns)
```

**SPEC System**
```
→ moai-foundation-specs           (SPEC structure)
→ moai-alfred-spec-authoring      (Writing SPECs)
→ moai-spec-management            (Managing SPECs)
```

**TAG System**
```
→ moai-foundation-tags            (TAG lifecycle)
→ moai-tag-policy-validator       (TAG validation)
```

**TRUST 5 Principles**
```
→ moai-foundation-trust           (Core: Test, Readable, Unified, Secured, Trackable)
```

**GitFlow & Branching**
```
→ moai-foundation-git             (Branch strategies)
→ moai-alfred-code-reviewer       (Code review workflow)
→ moai-essentials-review          (Review standards)
```

**Agent Architecture**
```
→ moai-alfred-workflow            (Main workflow)
→ moai-alfred-agent-guide         (Agent selection & patterns)
→ moai-cc-agents                  (Creating agents)
```

**Task Management**
```
→ moai-alfred-todowrite-pattern   (TodoWrite system)
→ moai-project-batch-questions    (Batch Q&A)
```

**User Interaction**
```
→ moai-alfred-ask-user-questions  (TUI-based questions)
→ moai-alfred-personas            (Communication styles)
```

**Configuration**
```
→ moai-cc-settings                (settings.json)
→ moai-cc-claude-md               (CLAUDE.md)
→ moai-project-config-manager     (Project config)
```

**Claude Code Framework**
```
→ moai-cc-agents                  (Agent files)
→ moai-cc-commands                (Command files)
→ moai-cc-skills                  (Skill files)
→ moai-cc-hooks                   (Hook files)
```

**Memory & Context**
```
→ moai-alfred-context-budget      (Context optimization)
→ moai-cc-memory                  (Context management)
```

---

### "I'm getting [Error]. What Skill?"

**Agent Won't Load**
```
→ moai-cc-agents                  (Check YAML structure)
→ moai-cc-settings                (Check permissions)
→ moai-essentials-debug           (Debugging)
```

**Command Not Recognized**
```
→ moai-cc-commands                (Check naming)
→ moai-cc-settings                (Check tool access)
```

**Skill Can't Be Found**
```
→ moai-cc-skills                  (Check file location)
→ moai-cc-settings                (Check imports)
```

**Hook Not Running**
```
→ moai-cc-hooks                   (Check hook setup)
→ moai-essentials-debug           (Debugging)
```

**SPEC Validation Failed**
```
→ moai-foundation-specs           (Check SPEC structure)
→ moai-alfred-spec-authoring      (SPEC patterns)
```

**TAG Chain Broken**
```
→ moai-foundation-tags            (TAG system)
→ moai-tag-policy-validator       (TAG validation)
```

**Code Not Passing Tests**
```
→ moai-foundation-trust           (TRUST 5 principles)
→ moai-essentials-debug           (Debugging)
→ moai-essentials-review          (Code review)
```

**Performance Issues**
```
→ moai-essentials-perf            (Optimization)
→ moai-domain-[domain]            (Domain-specific)
→ moai-lang-[language]            (Language-specific)
```

**Deployment Failed**
```
→ moai-baas-vercel-ext            (Vercel issues)
→ moai-domain-devops              (DevOps issues)
→ moai-foundation-git             (Git issues)
```

---

## Skill Categories Overview

### BaaS (Backend-as-a-Service)
**When**: Choosing or implementing backend platforms
**Skills**: moai-baas-foundation + 9 extensions
**File Size**: 10 skills total

```
moai-baas-foundation (2.0.0)
├─ Covers 9 platforms
├─ Describes 8 patterns (A-H)
└─ 1400 words

moai-baas-supabase-ext (2.0.0)
├─ RLS, Migrations, Realtime
└─ 1300 words

moai-baas-vercel-ext (2.0.0)
├─ Deployment, Edge Functions
└─ 1000 words

moai-baas-clerk-ext (1.0.0)
├─ Authentication, MFA, Organizations
└─ 1000 words

[6 more extensions: Firebase, Neon, Auth0, Cloudflare, Convex, Railway]
```

### Claude Code Infrastructure
**When**: Working with Claude Code files (agents, commands, skills, hooks)
**Skills**: 9 infrastructure skills
**File Size**: 4-20K per skill

```
moai-cc-agents         → Create/structure agent files
moai-cc-commands       → Create/structure command files
moai-cc-skills         → Create/structure skill files
moai-cc-hooks          → Implement hooks
moai-cc-settings       → Configure settings.json
moai-cc-claude-md      → Author CLAUDE.md
moai-cc-mcp-plugins    → Setup MCP servers
moai-cc-memory         → Optimize context
moai-cc-configuration  → System configuration
```

### Language Support
**When**: Working with specific programming languages
**Skills**: 20 language skills
**File Size**: 50-88K per skill

```
moai-lang-python       → Python patterns
moai-lang-typescript   → TypeScript/type safety
moai-lang-javascript   → JavaScript/ES features
moai-lang-shell        → Bash/Shell scripts
[+ 16 more languages]
```

### Domain Specialists
**When**: Working in specific technical domains
**Skills**: 10 domain skills
**File Size**: 60-80K per skill

```
moai-domain-backend         → Server architecture
moai-domain-frontend        → UI/UX patterns
moai-domain-database        → DB design
moai-domain-security        → Auth, encryption
moai-domain-devops          → CI/CD, containers
moai-domain-cli-tool        → CLI design
moai-domain-web-api         → REST/GraphQL
moai-domain-mobile-app      → Native apps
moai-domain-data-science    → ML, preprocessing
moai-domain-ml              → Deep learning, NLP
```

### Foundation Skills
**When**: Understanding core MoAI-ADK concepts
**Skills**: 6 foundation skills
**File Size**: 50-76K per skill

```
moai-foundation-specs       → SPEC structure
moai-foundation-tags        → TAG system
moai-foundation-trust       → TRUST 5 principles
moai-foundation-git         → GitFlow
moai-foundation-langs       → Language detection
moai-foundation-ears        → EARS framework
```

### Essentials
**When**: Improving code quality
**Skills**: 4 essentials skills
**File Size**: 50-56K per skill

```
moai-essentials-debug       → Debugging
moai-essentials-refactor    → Refactoring
moai-essentials-perf        → Performance
moai-essentials-review      → Code review
```

### Alfred SuperAgent
**When**: Working with Alfred and workflow orchestration
**Skills**: 16+ Alfred skills
**File Size**: Varies

```
moai-alfred-workflow        → Main workflow
moai-alfred-agent-guide     → Agent selection
moai-alfred-personas        → Communication styles
moai-alfred-ask-user-questions → User interaction
moai-alfred-spec-authoring  → SPEC writing
[+ 11 more Alfred skills]
```

### Project Management
**When**: Setting up and managing projects
**Skills**: 6+ project skills
**File Size**: 40-76K per skill

```
moai-project-config-manager → Configuration
moai-project-batch-questions → Batch Q&A
moai-project-documentation → Documentation
moai-project-language-initializer → Language setup
moai-project-template-optimizer → Template optimization
moai-learning-optimizer     → Learning paths
```

### Utility Skills
**When**: Specialized utilities
**Skills**: 4+ utility skills
**File Size**: Variable

```
moai-session-info           → Session metadata
moai-design-systems         → Design patterns
moai-streaming-ui           → Streaming UI
moai-spec-management        → SPEC management
moai-tag-policy-validator   → TAG validation
moai-jit-docs-enhanced      → JIT documentation
moai-change-logger          → Change tracking
```

### Special: Skill Factory
**When**: Creating new skills comprehensively
**Skills**: 1 (moai-cc-skill-factory)
**File Size**: 244K (largest, most comprehensive)

```
moai-cc-skill-factory
├─ SKILL.md                      [Main guide]
├─ STRUCTURE.md                  [File structure]
├─ STEP-BY-STEP-GUIDE.md        [Detailed steps]
├─ CHECKLIST.md                  [Validation]
├─ EXAMPLES.md                   [Example skills]
├─ WEB-RESEARCH.md              [Research guide]
├─ SKILL-FACTORY-WORKFLOW.md    [Workflow]
├─ SKILL-UPDATE-ADVISOR.md      [Updating skills]
├─ METADATA.md                   [Metadata guide]
├─ PARALLEL-ANALYSIS-REPORT.md [Analysis]
├─ PYTHON-VERSION-MATRIX.md    [Version info]
├─ reference.md                  [Reference]
├─ templates/                    [Templates]
│  ├─ SKILL-template.md
│  ├─ reference-template.md
│  └─ examples-template.md
└─ scripts/                      [Scripts]
   ├─ generate-structure.sh
   └─ validate-skill.sh
```

---

## Navigation by File Type

### I want to read about [TOPIC]

**SKILL.md** (Main content, 1000-1500 words)
```
Each skill has a SKILL.md with:
- Metadata (YAML frontmatter)
- Content sections (main material)
- Usage examples
- Validation checklist
```

**reference.md** (Detailed reference)
```
For skills that have more detailed material:
- API references
- Configuration options
- Advanced patterns
- Edge cases
```

**examples.md** (Code examples)
```
For practical examples:
- Complete working code
- Step-by-step tutorials
- Before/after comparisons
- Testing examples
```

**templates/** (Reusable templates)
```
For templates that agents use:
- agent-template.md
- command-template.md
- SKILL-template.md
- settings-template.json
```

**scripts/** (Utility scripts)
```
For validation and setup:
- validate-skill.sh
- pre-bash-check.sh
- generate-structure.sh
```

---

## Find Skills by Size (Largest to Smallest)

```
244K  moai-cc-skill-factory       (Special: Comprehensive guide)
 88K  moai-lang-dart
 84K  moai-lang-csharp
 80K  moai-domain-backend
 76K  moai-project-config-manager
 76K  moai-foundation-trust
 76K  moai-design-systems
 72K  moai-lang-javascript
 68K  moai-lang-swift
 60K  moai-lang-rust
 60K  moai-lang-go
 56K  moai-lang-kotlin
 56K  moai-lang-java
 56K  moai-essentials-debug
 56K  moai-alfred-personas
 52K  moai-project-batch-questions
 52K  moai-lang-python
 52K  moai-project-language-initializer
 52K  moai-alfred-spec-authoring
 52K  moai-alfred-ask-user-questions
 ...
```

---

## Find Skills by Update Date

**Recently Updated (Nov 6-9, 2025)**:
- All moai-baas-* skills
- moai-cc-skills, moai-cc-hooks, moai-cc-commands
- moai-lang-* skills (most)
- moai-domain-* skills
- moai-foundation-* skills
- moai-alfred-* skills
- moai-project-* skills
- moai-essentials-* skills

**Previously Updated (Oct-Nov)**:
- Foundation skills
- Domain skills
- Language skills

**Older Updates (Before October)**:
- moai-lang-template
- Some utility skills

---

## Common Workflows

### Workflow: Setup New Project
```
1. moai-project-language-initializer    → Select language
2. moai-cc-settings                     → Configure settings.json
3. moai-cc-claude-md                    → Create CLAUDE.md
4. moai-foundation-specs                → Create initial SPEC
5. moai-project-config-manager          → Configure project
6. moai-foundation-git                  → Setup GitFlow
```

### Workflow: Add Backend
```
1. moai-baas-foundation                 → Choose pattern (A-H)
2. moai-baas-[platform]-ext             → Setup chosen platform
3. moai-domain-database                 → Design schema
4. moai-domain-security                 → Setup auth
5. moai-domain-backend                  → Build API
6. moai-foundation-git                  → Commit structure
```

### Workflow: Build Frontend
```
1. moai-domain-frontend                 → UI patterns
2. moai-lang-typescript                 → Language patterns
3. moai-baas-vercel-ext                 → Deployment
4. moai-essentials-perf                 → Optimize
5. moai-essentials-review               → Code review
6. moai-foundation-git                  → Merge to main
```

### Workflow: Improve Code Quality
```
1. moai-foundation-trust                → Review TRUST 5
2. moai-essentials-debug                → Debug issues
3. moai-essentials-refactor             → Refactor
4. moai-essentials-perf                 → Optimize
5. moai-essentials-review               → Review
6. moai-foundation-specs                → Document
```

### Workflow: Troubleshoot
```
1. moai-essentials-debug                → Debugging tools
2. moai-cc-settings                     → Check configuration
3. Domain/Language skill                → Domain-specific help
4. moai-foundation-tags                 → Check TAG chain
5. moai-alfred-context-budget           → Check context
```

---

## Keyword Index

### A
- Agent creation → moai-cc-agents
- Authentication → moai-domain-security, moai-baas-clerk-ext
- Architecture → moai-baas-foundation, moai-domain-backend

### B
- Backend → moai-domain-backend, moai-baas-*
- BaaS → moai-baas-foundation
- Best practices → moai-alfred-best-practices, moai-foundation-trust

### C
- CLI → moai-domain-cli-tool
- Clerk → moai-baas-clerk-ext
- Claude Code → moai-cc-*
- Cloudflare → moai-baas-cloudflare-ext
- Code review → moai-essentials-review
- Command creation → moai-cc-commands
- Configuration → moai-project-config-manager, moai-cc-settings
- Convex → moai-baas-convex-ext
- Context budget → moai-alfred-context-budget

### D
- Database → moai-domain-database
- Data science → moai-domain-data-science
- Dart → moai-lang-dart
- Debug → moai-essentials-debug
- DevOps → moai-domain-devops
- Documentation → moai-project-documentation

### E
- Edge functions → moai-baas-vercel-ext
- EARS → moai-foundation-ears

### F
- Firebase → moai-baas-firebase-ext
- Frontend → moai-domain-frontend

### G
- Git → moai-foundation-git
- Go → moai-lang-go
- GitHub → moai-alfred-issue-labels

### H
- Hooks → moai-cc-hooks

### J
- Java → moai-lang-java
- JavaScript → moai-lang-javascript
- JIT → moai-jit-docs-enhanced

### K
- Kotlin → moai-lang-kotlin

### L
- Language detection → moai-alfred-language-detection
- Languages → moai-foundation-langs, moai-lang-*

### M
- Memory → moai-cc-memory
- Mobile → moai-domain-mobile-app
- MCP → moai-cc-mcp-plugins

### N
- Neon → moai-baas-neon-ext

### P
- Performance → moai-essentials-perf
- Personas → moai-alfred-personas
- PHP → moai-lang-php
- Python → moai-lang-python
- Project setup → moai-project-*

### R
- Railway → moai-baas-railway-ext
- React → moai-domain-frontend
- Refactor → moai-essentials-refactor
- Ruby → moai-lang-ruby
- Rust → moai-lang-rust

### S
- SPEC → moai-foundation-specs
- Security → moai-domain-security
- Shell → moai-lang-shell
- Skill creation → moai-cc-skill-factory, moai-cc-skills
- SQL → moai-lang-sql
- Supabase → moai-baas-supabase-ext
- Swift → moai-lang-swift

### T
- TAG → moai-foundation-tags
- Template → moai-lang-template
- TodoWrite → moai-alfred-todowrite-pattern
- Trust → moai-foundation-trust
- TypeScript → moai-lang-typescript

### V
- Vercel → moai-baas-vercel-ext

### W
- Web API → moai-domain-web-api
- Workflow → moai-alfred-workflow

---

**Need more help?** Check the comprehensive analysis at `/Users/goos/MoAI/MoAI-ADK/.claude/skills/SKILLS_STRUCTURE_ANALYSIS.md`

Last updated: November 9, 2025
Total skills indexed: 92
