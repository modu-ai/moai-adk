---
id: PRODUCT-001
version: 0.2.0
status: active
created: 2025-10-01
updated: 2025-10-27
author: @Alfred
priority: high
---

# MoAI-ADK Product Definition

## HISTORY

### v0.2.0 (2025-10-27)
- **UPDATED**: Auto-generated comprehensive product definition based on codebase analysis
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission (actual team composition), User (research-based personas), Problem (identified challenges), Strategy (competitive differentiators), Success (measurable KPIs)
- **ANALYSIS**: Integrated README.md, pyproject.toml, source code structure analysis

### v0.1.3 (2025-10-17)
- **UPDATED**: Template version synced (v0.3.8)
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission (finalized team of 12 agents: Alfred + 11 specialists)
  - Added implementation-planner, tdd-implementer, quality-gate
  - Split code-builder into implementation-planner + tdd-implementer + quality-gate

### v0.1.2 (2025-10-17)
- **UPDATED**: Agent count adjusted (9 → 11)
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission (updated Alfred SuperAgent roster)

### v0.1.1 (2025-10-17)
- **UPDATED**: Template defaults aligned with the real MoAI-ADK project
- **AUTHOR**: @Alfred
- **SECTIONS**: Mission, User, Problem, Strategy, Success populated with project context

### v0.1.0 (2025-10-01)
- **INITIAL**: Authored the product definition document
- **AUTHOR**: @project-owner
- **SECTIONS**: Mission, User, Problem, Strategy, Success, Legacy

---

## @DOC:MISSION-001 Core Mission

> **"No SPEC, no CODE. No tests without SPEC. No complete documentation without tests and code."**

MoAI-ADK combats Frankenstein code at the root by enforcing a **SPEC-first TDD methodology** with AI-powered automation.

### Core Value Proposition

#### Four Key Values

1. **Consistency**: A four-step workflow (Init → Plan → Run → Sync) safeguards delivery quality through automated checks.
2. **Quality**: TRUST principles (Test First ≥85%, Readable, Unified, Secured, Trackable) apply automatically to every feature.
3. **Traceability**: The @TAG system (`@SPEC → @TEST → @CODE → @DOC`) preserves end-to-end lineage and enables instant impact analysis.
4. **Universality**: Supports 18+ programming languages (Python, TypeScript, Go, Rust, Java, Kotlin, Swift, Dart, C#, C++, C, Scala, Ruby, PHP, SQL, Shell, R, and more).

#### Alfred SuperAgent: Orchestrating 19 Team Members

**Alfred** coordinates a distributed team of 19 AI agents (Alfred + 12 core sub-agents + 6 specialists):

**Core Sub-agents** (Primary workflow agents):
- **spec-builder** 🏗️: Authors SPECs in EARS format (Easy Approach to Requirements Syntax) – Sonnet
- **implementation-planner** 📋: Analyzes SPECs and derives implementation plans with architecture guidance – Sonnet
- **tdd-implementer** 🔬: Executes RED–GREEN–REFACTOR cycles with guaranteed coverage – Sonnet
- **quality-gate** 🛡️: Enforces TRUST 5 principles (Test, Readable, Unified, Secured, Trackable) – Haiku
- **doc-syncer** 📖: Maintains living documentation with automatic sync – Haiku
- **tag-agent** 🏷️: Manages the @TAG system and validates traceability chains – Haiku
- **git-manager** 🚀: Automates Git workflows, branch strategies, and PR management – Haiku
- **debug-helper** 🔍: Diagnoses runtime issues and suggests fixes – Sonnet
- **trust-checker** ✅: Verifies TRUST 5 compliance before deployment – Haiku
- **cc-manager** 🛠️: Configures Claude Code agents, commands, and skills – Sonnet
- **project-manager** 📂: Bootstraps projects with language detection and metadata – Sonnet
- **skill-factory** 🧠: Creates and maintains reusable Skills – Sonnet

**Specialist Sub-agents** (Domain expertise):
- **code-reviewer** 💬: SOLID principles and code quality review
- **test-strategist** 🧪: Test design and coverage optimization
- **doc-writer** 📝: Technical documentation generation
- And more (total 19 members)

**55+ Claude Skills** across 6 tiers:
- Foundation (6): TRUST, TAG, SPEC, Git, EARS principles
- Essentials (4): Debugging, performance, refactoring, code review
- Alfred (7): Workflow automation
- Domain (10): Backend, frontend, security, database, DevOps, etc.
- Language (18): Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin, Swift, Dart, C#, C++, C, Scala, Ruby, PHP, SQL, Shell, R
- Claude Code Ops (1): Session management

### Key Features

1. **SPEC-First Philosophy**: All development starts with clear, EARS-formatted specifications
2. **Automated TDD**: RED → GREEN → REFACTOR cycles executed with guaranteed 85%+ coverage
3. **Living Documentation**: Docs sync automatically with code changes (one command: `/alfred:3-sync`)
4. **@TAG Traceability**: Every SPEC/TEST/CODE/DOC gets a unique @TAG for impact tracking
5. **AI-Powered Agents**: 19 specialized agents handle different development phases
6. **Multi-Language Support**: Works with Python, TypeScript, Go, Rust, Java, and 13+ more languages
7. **Production-Grade Skills**: 55+ reusable knowledge capsules for best practices
8. **Claude Code Integration**: Deep integration with Claude Code hooks and agents

---

## @SPEC:USER-001 Primary Users & Personas

### Primary Audience

**Solo/Small Team Developers (1-5 people)**
- **Who**: Individual developers and small startup teams who want to adopt AI-powered development but struggle with code quality and trust issues
- **Core Needs**:
  - Clear requirements before implementation (eliminate guessing)
  - Guaranteed test coverage (no "testing later" excuses)
  - Auto-generated documentation (reduce manual work)
  - Code that stays maintainable even 6 months later
- **Critical Scenarios**:
  - Writing a new API endpoint and needing tests + docs
  - Refactoring legacy code safely
  - Onboarding a new team member who needs context

**Enterprise/Mid-Market Teams (5-50 people)**
- **Who**: Companies building products with regulated requirements, complex architectures, or distributed teams
- **Core Needs**:
  - Traceability for compliance (FDA, financial regulations)
  - Consistent code standards across the organization
  - Clear audit trail for code changes
  - Reduced context-switching overhead
- **Critical Scenarios**:
  - Multi-team coordination on shared codebases
  - Regulatory compliance audits
  - Knowledge transfer between teams

### Secondary Audience (Optional)

**AI/ML Research Teams**
- **Who**: Data scientists and ML engineers exploring AI-assisted development
- **Needs**:
  - Version control for ML pipelines
  - Reproducible research notebooks
  - Clear experiment documentation
- **When**: Emerging market; not primary focus yet

**Open Source Maintainers**
- **Who**: Project maintainers seeking to standardize contributor workflows
- **Needs**:
  - Template-driven contribution process
  - Automated quality checks
  - Clear onboarding for contributors

---

## @SPEC:PROBLEM-001 Problems to Solve

### High Priority

1. **Unclear AI-Generated Code Trustworthiness**
   - **Problem**: Developers cannot trust code generated by Claude or ChatGPT
   - **Root Cause**: No clear requirements (SPEC) means implementation is based on guessing
   - **MoAI Solution**: SPEC-first forces clarity before code; TDD guarantees coverage

2. **Missing Test Coverage in AI Generation**
   - **Problem**: AI generates happy-path code; edge cases and error handling get missed
   - **Root Cause**: No test-first discipline; testing deferred to "later" phase
   - **MoAI Solution**: TDD loop (RED → GREEN → REFACTOR) guarantees 85%+ coverage

3. **Documentation Drift Over Time**
   - **Problem**: Code changes but documentation stays static; 6 months later, docs are useless
   - **Root Cause**: Manual documentation is hard to keep in sync
   - **MoAI Solution**: `/alfred:3-sync` auto-generates and syncs all documentation

4. **Context Loss Between Sessions**
   - **Problem**: Each session requires explaining the project structure, previous decisions, and constraints from scratch
   - **Root Cause**: No persistent project context; everything lives in chat history
   - **MoAI Solution**: Alfred remembers project structure, standards, and history via `.moai/` files

5. **Unclear Impact of Changes**
   - **Problem**: When requirements change, it's unclear which code/tests/docs are affected
   - **Root Cause**: No traceability system; searching codebase is manual and error-prone
   - **MoAI Solution**: @TAG system enables `rg '@SPEC:USER-001'` to find all related code instantly

### Medium Priority

- Performance optimization challenges in AI-generated code
- Difficulty enforcing coding standards across a team
- Onboarding new developers takes weeks; old code still mysterious
- Supply chain security: unknown dependencies pulled from PyPI/npm

### Current Failure Cases

- "I asked Claude to write a login feature and got code that doesn't validate passwords" → No SPEC meant the requirement wasn't clear
- "Coverage dropped from 87% to 42% after my refactor" → No test-first discipline
- "The README says the API returns 200, but it actually returns 201" → Documentation drifted
- "I need to fix an AUTH bug but can't find all the places it affects" → No traceability

---

## @DOC:STRATEGY-001 Differentiators & Competitive Strengths

### 1. Only Framework with Proven "SPEC → TDD → Sync" Cycle

- **Differentiator**: No other framework enforces SPEC-first development at this scale
- **When It Matters**: Enterprise teams that need audit trails; regulated industries (finance, healthcare)
- **Proof**: 0.5.6 production version with 87.84% coverage in the MoAI-ADK codebase itself

### 2. Complete AI Agent Ecosystem (19 agents + 55 Skills)

- **Differentiator**: Not just a prompt template; a full team of specialized agents with deep knowledge
- **When It Matters**: Complex projects (microservices, multi-team codebases); reducing context loss
- **Proof**: alfred-hooks system, 40+ SPEC documents in project, GitFlow automation

### 3. Language-Agnostic with Deep Best Practices (18+ Languages)

- **Differentiator**: Works equally well with Python, TypeScript, Go, Rust, Java, Kotlin, Swift, Dart, C#, C++, C, Scala, Ruby, PHP, SQL, Shell, R
- **When It Matters**: Polyglot teams; migrating codebases between languages
- **Proof**: moai-lang-* Skills for each language with language-specific TDD practices

### 4. @TAG System for End-to-End Traceability

- **Differentiator**: Unique tagging system that connects SPEC → TEST → CODE → DOC in one searchable index
- **When It Matters**: Compliance audits; impact analysis; refactoring confidence
- **Proof**: `.moai/memory/SPEC-METADATA.md` and `/alfred:3-sync` validation

### 5. Progressive Disclosure of Knowledge (Just-In-Time Skill Loading)

- **Differentiator**: Skills load only when needed; minimizes context overhead
- **When It Matters**: Working with limited Claude context; managing complex projects
- **Proof**: Foundation → Essentials → Alfred → Domain/Language tiers

### 6. Production-Grade Templates with Automatic Optimization

- **Differentiator**: Not sample code; real templates from 0.5.6 production system
- **When It Matters**: Accelerating project bootstrap; enforcing org standards
- **Proof**: 40+ files in `.claude/` directory; automated initialization

---

## @SPEC:SUCCESS-001 Success Metrics & KPIs

### Immediately Measurable KPIs

#### 1. Test Coverage (Guaranteed ≥85%)
- **Metric**: `coverage_percentage`
- **Target**: ≥85% (fail build if <85%)
- **Measurement**: `pytest --cov=src/` (Python), `vitest --coverage` (TypeScript), etc.
- **Baseline**: MoAI-ADK itself: **87.84%** (468/535 lines)
- **Success Cadence**: Per commit; automated in CI/CD

#### 2. SPEC Completion Rate
- **Metric**: `specs_implemented / total_planned_specs × 100`
- **Target**: ≥80% of planned features have corresponding SPECs
- **Measurement**: Count files in `.moai/specs/SPEC-*/spec.md`
- **Baseline**: MoAI-ADK: **28 SPECs** (95% completion estimated)
- **Success Cadence**: Weekly

#### 3. Documentation Sync Rate
- **Metric**: `auto_generated_docs / total_docs × 100`
- **Target**: ≥90% of docs updated by `/alfred:3-sync`
- **Measurement**: Check CHANGELOG, API docs, README for `@DOC` TAGs
- **Baseline**: 100% for `/alfred:3-sync` outputs
- **Success Cadence**: Per sync operation

#### 4. @TAG Chain Integrity
- **Metric**: `orphan_tags_found / total_tags × 100`
- **Target**: 0% orphan TAGs (all SPEC/TEST/CODE/DOC connected)
- **Measurement**: `rg '@(SPEC|TEST|CODE|DOC):' -n` validation
- **Baseline**: 0% orphan rate in current codebase
- **Success Cadence**: Per `/alfred:3-sync` run

#### 5. Development Time Reduction
- **Metric**: `time_to_implement_feature / previous_time`
- **Target**: 40-50% faster feature delivery
- **Measurement**: Time from SPEC written to code deployed
- **Baseline**: Example Todo API: 15 minutes (Plan + Run + Sync)
- **Success Cadence**: Per feature completion

#### 6. Context Retention (AI Memory)
- **Metric**: `questions_repeated_in_same_session / total_questions × 100`
- **Target**: <5% (Alfred remembers context)
- **Measurement**: Monitor chat history for repeated context requests
- **Baseline**: Estimated 100% context loss without MoAI (traditional setup)
- **Success Cadence**: Per session

### Measurement Cadence

- **Per Commit**: Coverage enforcement, SPEC alignment
- **Daily**: Tag integrity, branch status
- **Weekly**: Feature completion rate, documentation sync rate
- **Monthly**: Time-to-delivery trends, team adoption metrics
- **Quarterly**: Cost reduction (fewer bugs, less rework), developer satisfaction

### Long-Term Vision (12+ months)

- **Global Community**: 5,000+ active users across teams
- **Certification Program**: "MoAI-Certified Developer" badge
- **Integration Ecosystem**: Built-in support for major IDEs (VS Code, JetBrains)
- **Open Source Leadership**: Largest open-source SPEC-first TDD framework

---

## Legacy Context

### Existing Assets

- **28 SPEC Documents**: Comprehensive specifications written in EARS format
- **Mature Codebase**: 40+ Python modules with clear separation of concerns
- **Test Infrastructure**: pytest + coverage with 87.84% baseline
- **Template System**: `.moai/` and `.claude/` directories with best practices
- **CI/CD Pipeline**: GitHub Actions with automated testing and quality gates
- **Documentation**: 6 language READMEs (English, Korean, Japanese, Chinese, Thai, Hindi)
- **Skills Library**: 55+ production-grade Claude Skills

### Competitive Positioning

- **vs. GitHub Copilot**: MoAI adds SPEC requirement and guaranteed coverage
- **vs. OpenAI Assistants**: MoAI has traceability (@TAG) and automated documentation
- **vs. Manual Development**: MoAI reduces time 40-50% via automation

---

## TODO:SPEC-BACKLOG-001 Next SPEC Candidates

### High Priority (Next 2 months)

1. **SPEC-ENHANCE-001**: Enhance Alfred's language model selection (Sonnet vs Haiku tuning)
2. **SPEC-SECURITY-001**: Implement supply chain security scanning (bandit + pip-audit integration)
3. **SPEC-PERFORMANCE-001**: Optimize hook execution time (<50ms target)

### Medium Priority (Next 3-6 months)

4. **SPEC-IDE-INTEGRATION-001**: VS Code extension for Alfred commands
5. **SPEC-TEAM-COLLABORATION-001**: Multi-developer conflict resolution in `.moai/specs/`
6. **SPEC-ML-SUPPORT-001**: AutoML for Python data science projects

### Exploration Phase (6+ months)

7. **SPEC-MARKETPLACE-001**: MoAI Skills marketplace for community contributions
8. **SPEC-CERTIFICATION-001**: Official "MoAI-Certified Developer" certification program

---

## EARS Requirement Authoring Guide

### EARS (Easy Approach to Requirements Syntax)

Use these EARS patterns when writing new SPECs:

#### EARS Patterns with Examples

1. **Ubiquitous Requirements** (foundational): The system SHALL provide [capability]
   - Example: "The system SHALL provide SPEC-based requirement validation"

2. **Event-driven Requirements** (conditional): WHEN [event], the system SHALL [behavior]
   - Example: "WHEN a user runs `/alfred:1-plan`, the system SHALL generate EARS-formatted SPEC"

3. **State-driven Requirements** (during state): WHILE [state], the system SHALL [behavior]
   - Example: "WHILE the git branch is feature/*, the system SHALL enforce test coverage ≥85%"

4. **Optional Features**: WHERE [condition], the system MAY [behavior]
   - Example: "WHERE environment is production, the system MAY enable advanced caching"

5. **Constraints**: IF [condition], the system SHALL enforce [constraint]
   - Example: "IF coverage drops below 85%, the system SHALL block merge to main"

#### Sample Product SPEC Application

```markdown
# @SPEC:FEATURE-CLI-001: CLI Enhancement for Quick Onboarding

## Ubiquitous Requirements
- The system SHALL provide a `moai-adk init` command for new projects
- The system SHALL auto-detect the primary language (Python, TypeScript, etc.)

## Event-driven Requirements
- WHEN user runs `moai-adk init my-project`, the system SHALL create `.moai/config.json`
- WHEN `.moai/config.json` is missing, the system SHALL prompt for required fields

## State-driven Requirements
- WHILE in interactive mode, the system SHALL provide TUI menus for language selection

## Optional Features
- WHERE the project has existing code, the system MAY suggest technology stack

## Constraints
- IF Python version < 3.13, the system SHALL display a warning
- Template generation time SHALL NOT exceed 10 seconds
```

---

_This document serves as the baseline when `/alfred:1-plan` runs. Update with actual project context and user research as the project evolves._
