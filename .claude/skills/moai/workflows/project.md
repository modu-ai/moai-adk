---
name: moai-workflow-project
description: >
  Generates project documentation from codebase analysis or user input.
  Creates product.md, structure.md, and tech.md in .moai/project/ directory,
  plus architecture maps in .moai/project/codemaps/ directory.
  Supports new and existing project types with LSP server detection.
  Use when initializing projects or generating project documentation.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "project, documentation, initialization, codebase-analysis, setup"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["project", "init", "documentation", "setup", "initialize"]
  agents: ["manager-project", "manager-docs", "Explore", "expert-devops"]
  phases: ["project"]
---

# Workflow: project - Project Documentation Generation

Purpose: Generate project documentation through smart questions and codebase analysis. Creates product.md, structure.md, and tech.md in .moai/project/ directory, plus architecture documentation in .moai/project/codemaps/ directory.

This workflow is also triggered automatically when project documentation does not exist and the user requests other workflows (plan, run, sync, etc.). See SKILL.md Step 2.5 for the auto-detection mechanism.

---

## Phase 0: Project Type Detection

[HARD] Auto-detect project type by checking for existing source code files FIRST.

Detection Logic:
1. Check if source code files exist in the current directory (using Glob for *.py, *.ts, *.js, *.go, *.java, *.rb, *.rs, src/, lib/, app/)
2. If source code found: Classify as "Existing Project" and present confirmation
3. If no source code found: Classify as "New Project"

[HARD] Present detection result via AskUserQuestion for user confirmation.

Question: Project type detected. Please confirm (in user's conversation_language):

Options (first option is auto-detected recommendation):

If source code found:
- Existing Project (Recommended): Your codebase will be automatically analyzed to generate accurate documentation. MoAI scans your files, architecture, and dependencies to create product.md, structure.md, and tech.md.
- New Project: Choose this if you want to start fresh and define the project from scratch through a guided interview, ignoring existing code.

If no source code found:
- New Project (Recommended): MoAI will guide you through a short interview to understand your project goals, technology choices, and key features. This creates the foundation documents for all future development.
- Existing Project: Choose this if your code exists elsewhere and you want to point MoAI to analyze it.

Routing:

- New Project selected: Proceed to Phase 0.5
- Existing Project selected: Proceed to Phase 1

---

## Phase 0.3: Deep Interview (New Projects Only)

Purpose: Replace the static four-question sequence with a structured deep interview that adapts to user responses. This produces richer project context for documentation generation.

[HARD] All questions MUST use AskUserQuestion in user's conversation_language.
[HARD] During the interview, the agent MUST NOT write implementation code or generate documentation. The sole output is `.moai/project/interview.md`.

**Interview Rounds (3 rounds maximum, configured in `.moai/config/sections/interview.yaml`):**

**Round 1: Vision**

Topic: What does this project do and who is it for?

Present via AskUserQuestion with exactly 4 options tailored to common project patterns. Example:
- Option 1 (Recommended): Web application for end users: A frontend + backend system serving a web-based user interface. Best for dashboards, tools, and customer-facing products.
- Option 2: API service or backend: A REST/GraphQL API or microservice consumed by other clients. Best for mobile backends, integrations, and data platforms.
- Option 3: CLI tool or automation script: A command-line utility run by developers or operators. Best for build tools, deployment scripts, and developer utilities.
- Option 4: Type your own answer: Enter a custom response if none of the above match your vision.

**Round 2: Technology**

Topic: What is the primary technology stack?

Present via AskUserQuestion with exactly 4 options based on Round 1 answer context:
- Option 1 (Recommended): TypeScript/JavaScript: Full-stack or frontend-heavy projects. Largest ecosystem. Works for React frontends, Node.js backends, Bun runtimes.
- Option 2: Python: Backend APIs, AI/ML workloads, scripting. FastAPI, Django, or simple scripts.
- Option 3: Go: High-performance microservices, CLI tools, cloud-native binaries. Simple deployment.
- Option 4: Type your own answer: Enter a custom response to specify Rust, Java, Kotlin, Ruby, Swift, C#, or another stack.

**Round 3: Scope**

Topic: What are the key features and explicit boundaries?

Present via AskUserQuestion with exactly 4 options based on the vision and technology selected. Example for a web app:
- Option 1 (Recommended): Authentication + CRUD data layer + REST API: Core features for most web apps. User login, database persistence, and API endpoints.
- Option 2: Read-only frontend + external API integration: Consumes existing data sources. No database needed.
- Option 3: Real-time collaboration features: WebSocket or SSE for live updates, shared state.
- Option 4: Type your own answer: Describe the exact features and what is explicitly out of scope.

**Output:** Write all answers to `.moai/project/interview.md` with this structure:

```
# Project Interview

## Round 1: Vision
Question: {question asked}
Answer: {user's answer}

## Round 2: Technology
Question: {question asked}
Answer: {user's answer}

## Round 3: Scope
Question: {question asked}
Answer: {user's answer}
```

After the interview, use the gathered information to generate documentation and proceed to Phase 3 (skip Phase 1 and Phase 2 since there is no existing code to analyze). Pass `interview.md` to Phase 3 as the primary input for documentation generation.

---

## Phase 1: Codebase Analysis (Existing Projects Only)

[HARD] Delegate codebase analysis to the Explore subagent.

[SOFT] Apply --deepthink for comprehensive analysis.

Analysis Objectives passed to Explore agent:

- Project Structure: Main directories, entry points, architectural patterns
- Technology Stack: Languages, frameworks, key dependencies
- Core Features: Main functionality and business logic locations
- Build System: Build tools, package managers, scripts

Expected Output from Explore agent:

- Primary Language detected
- Framework identified
- Architecture Pattern (MVC, Clean Architecture, Microservices, etc.)
- Key Directories mapped (source, tests, config, docs)
- Dependencies cataloged with purposes
- Entry Points identified

Execution Modes:

- Fresh Documentation: When .moai/project/ is empty, generate all three files
- Update Documentation: When docs exist, read existing, analyze for changes, ask user which files to regenerate

---

## Phase 1.5: Deep Interview (Existing Projects Only)

Purpose: After codebase analysis, gather user intent and context that cannot be inferred from the code alone. Questions are informed by the analysis results from Phase 1.

[HARD] All questions MUST use AskUserQuestion in user's conversation_language.
[HARD] During the interview, the agent MUST NOT generate documentation or write files. The sole output is `.moai/project/interview.md`.

**Interview Rounds (3 rounds maximum, configured in `.moai/config/sections/interview.yaml`):**

**Round 1: Ownership and Purpose**

Topic: Who maintains this project and what is the primary goal going forward?

Present via AskUserQuestion with exactly 4 options based on Phase 1 detected project type:
- Option 1 (Recommended): Active product being developed further: This codebase is actively developed and the documentation should reflect its current trajectory and roadmap.
- Option 2: Legacy system being maintained: The codebase is stable and the documentation should reflect its current state for maintenance and onboarding.
- Option 3: System being refactored or migrated: Major structural changes are planned and documentation should reflect the target state.
- Option 4: Type your own answer: Enter a custom response to describe the ownership context.

**Round 2: Constraints and Non-Goals**

Topic: What are the known constraints, technical debts, or things this project intentionally does NOT do?

Present via AskUserQuestion with exactly 4 options informed by Phase 1 analysis findings:
- Option 1 (Recommended): No known critical constraints: Document the codebase as-is without constraint annotations.
- Option 2: Performance or scalability constraints exist: There are known bottlenecks or scaling limits that should be documented.
- Option 3: Security or compliance constraints exist: Specific security requirements or compliance rules affect the architecture.
- Option 4: Type your own answer: Describe the specific constraints or non-goals for this project.

**Round 3: Documentation Priority**

Topic: What is the most important aspect to capture accurately in the documentation?

Present via AskUserQuestion with exactly 4 options:
- Option 1 (Recommended): Architecture and module boundaries: Prioritize documenting how the system is structured and how modules interact.
- Option 2: Technology stack and dependencies: Prioritize the frameworks, libraries, and their versions for onboarding.
- Option 3: Core business logic and data flow: Prioritize documenting what the system does and how data moves through it.
- Option 4: Type your own answer: Specify what should be documented with highest fidelity.

**Output:** Write all answers to `.moai/project/interview.md` with this structure:

```
# Project Interview

## Round 1: Ownership and Purpose
Question: {question asked}
Answer: {user's answer}

## Round 2: Constraints and Non-Goals
Question: {question asked}
Answer: {user's answer}

## Round 3: Documentation Priority
Question: {question asked}
Answer: {user's answer}
```

Pass `interview.md` to Phase 2 (User Confirmation) and Phase 3 (Documentation Generation) as additional context. Documentation agents MUST read interview.md before generating files.

---

## Phase 2: User Confirmation

Present analysis summary via AskUserQuestion.

Display in user's conversation_language:

- Detected Language
- Framework
- Architecture
- Key Features list

Options:

- Proceed with documentation generation (Recommended): MoAI will generate product.md, structure.md, and tech.md based on the analysis above. You can review and edit the documents afterwards.
- Review specific analysis details first: See a detailed breakdown of each detected component before generating documents. Useful if you want to correct any misdetected frameworks or features.
- Cancel and adjust project configuration: Stop the process and make changes to your project setup. Choose this if the analysis looks significantly incorrect.

If "Review details": Provide detailed breakdown, allow corrections.
If "Proceed": Continue to Phase 3.
If "Cancel": Exit with guidance.

---

## Phase 3: Documentation Generation

[HARD] Delegate documentation generation to the manager-docs subagent.

Pass to manager-docs:

- Analysis Results from Phase 1 (or user input from Phase 0.5)
- User Confirmation from Phase 2
- Output Directory: .moai/project/
- Language: conversation_language from config

Output Files:

- product.md: Project name, description, target audience, core features, use cases
- structure.md: Directory tree, purpose of each directory, key file locations, module organization
- tech.md: Technology stack overview, framework choices with rationale, dev environment requirements, build and deployment config

---

## Phase 3.1: Independent Document Audit (Conditional)

Purpose: Prevent confirmation bias by running an adversarial audit of the generated project documents before proceeding to codemaps and completion. The auditor sees only the final documents — not the analysis reasoning — and is prompted to find defects, not rationalize acceptance.

Activation: Controlled by harness.yaml `plan_audit.enabled` setting.

- `minimal`: Skip this phase
- `standard`: Run plan-auditor once (default)
- `thorough`: Run plan-auditor + cross-validate with evaluator-active

Skip Conditions:
- harness.yaml `plan_audit.enabled: false`
- Phase 3 produced no output files (documentation generation failed)

#### Step 3.1.1: Invoke plan-auditor

Agent: plan-auditor subagent

Delegation pattern: "Use the plan-auditor subagent to audit project documents at .moai/project/ — document type: project, iteration 1."

Do NOT pass the analysis reasoning or interview context to plan-auditor. The agent enforces context isolation (M1) and will ignore injected reasoning. Pass only the document directory path.

#### Step 3.1.2: Read Verdict

After plan-auditor completes, read the report at `.moai/reports/plan-audit/PROJECT-review-1.md`.

Extract the verdict line: `Verdict: PASS | FAIL`

If PASS: Proceed to Phase 3.3 (Codemaps Generation).

If FAIL: Enter retry loop.

#### Step 3.1.3: Retry Loop (max 3 iterations)

On FAIL:

1. Delegate back to manager-docs: "Use the manager-docs subagent to revise .moai/project/ documents based on the review report at .moai/reports/plan-audit/PROJECT-review-{N}.md. Address all defects listed in the report."

2. After manager-docs revision, re-invoke plan-auditor: "Use the plan-auditor subagent to audit project documents at .moai/project/ — document type: project, iteration {N+1}. Previous review report: .moai/reports/plan-audit/PROJECT-review-{N}.md"

3. Read new verdict from `.moai/reports/plan-audit/PROJECT-review-{N+1}.md`.

4. If PASS: Proceed to Phase 3.3.

5. If FAIL and iteration < 3: Repeat from step 1 with incremented iteration.

6. If FAIL and iteration = 3: Escalate to user via AskUserQuestion with the final review report. Options:
   - Fix manually and retry: User edits documents, then re-run audit
   - Accept as-is: Proceed despite audit failure (user override)
   - Cancel: Stop project documentation generation

---

## Phase 3.3: Codemaps Generation

Purpose: Generate architecture documentation in `.moai/project/codemaps/` directory based on codebase analysis results from Phase 1.

[HARD] This phase runs automatically after Phase 3 documentation generation.

Agent Chain:
- Explore subagent: Analyze codebase architecture (reuse Phase 1 results if available)
- manager-docs subagent: Generate codemaps documentation files

Output Files (in `.moai/project/codemaps/` directory):
- overview.md: High-level architecture summary, design patterns, system boundaries
- modules.md: Module descriptions, responsibilities, public interfaces
- dependencies.md: Dependency graph, external packages, internal module relationships
- entry-points.md: Application entry points, CLI commands, API routes, event handlers
- data-flow.md: Data flow paths, request lifecycle, state management patterns

Skip Conditions:
- New projects with no existing code (Phase 0.5 path): Skip codemaps generation, create placeholder `.moai/project/codemaps/overview.md` with project goals only
- User explicitly requests skip via AskUserQuestion in Phase 2

For detailed codemaps generation process, delegate to codemaps workflow (workflows/codemaps.md).

---

## Phase 3.5: Development Environment Check

Goal: Verify LSP servers are installed for the detected technology stack.

Language-to-LSP Mapping (all 16 MoAI-supported languages, alphabetical):

- C++: clangd (check: which clangd)
- C#: omnisharp or roslyn-ls (check: which omnisharp)
- Elixir: elixir-ls or lexical (check: which elixir-ls)
- Flutter: dart language-server (bundled with Dart SDK, check: which dart)
- Go: gopls (check: which gopls)
- Java: jdtls (Eclipse JDT Language Server)
- JavaScript: typescript-language-server (check: which typescript-language-server)
- Kotlin: kotlin-language-server
- PHP: phpactor or intelephense (check: which phpactor)
- Python: pylsp or pyright-langserver (check: which pylsp)
- R: R with languageserver package (check: which R)
- Ruby: ruby-lsp or solargraph (check: which ruby-lsp)
- Rust: rust-analyzer (check: which rust-analyzer)
- Scala: metals
- Swift: sourcekit-lsp
- TypeScript: typescript-language-server (check: which typescript-language-server)

Note: The canonical language name for Dart/Flutter ecosystem is "Flutter",
matching `.claude/skills/moai/workflows/sync.md` Phase 0.6.1. Per
CLAUDE.local.md Section 22, all 16 languages are treated as equal
first-class citizens; the user's project marker files determine which
server(s) actually spawn at runtime.

If LSP server is NOT installed, present AskUserQuestion:

- Continue without LSP: Proceed to completion
- Show installation instructions: Display setup guide for detected language
- Auto-install now: Use expert-devops subagent to install (requires confirmation)

---

## Phase 3.7: Development Methodology Auto-Configuration

Goal: Automatically set the `development_mode` in `.moai/config/sections/quality.yaml` based on the project analysis results from Phase 0 and Phase 1.

[HARD] This phase runs automatically without user interaction. No AskUserQuestion is needed.

Auto-Detection Logic:

For New Projects (Phase 0 classified as "New Project"):
- Set `development_mode: "tdd"` (test-first development)
- Rationale: New projects benefit from test-first development with clean RED-GREEN-REFACTOR cycles

For Existing Projects (Phase 0 classified as "Existing Project"):
- Step 1: Check for existing test files using Glob patterns (*_test.go, *_test.py, *.test.ts, *.test.js, *.spec.ts, *.spec.js, test_*.py, tests/, __tests__/, spec/)
- Step 2: Estimate test coverage level based on test file count relative to source file count:
  - No test files found (0%): Set `development_mode: "ddd"` (need characterization tests first)
  - Few test files (< 10% ratio): Set `development_mode: "ddd"` (insufficient coverage, characterization tests first)
  - Moderate test files (10-49% ratio): Set `development_mode: "tdd"` (partial tests, expand with test-first development)
  - Good test files (>= 50% ratio): Set `development_mode: "tdd"` (strong test base for test-first development)

Implementation:
- Read current `.moai/config/sections/quality.yaml`
- Update only the `constitution.development_mode` field
- Preserve all other settings in quality.yaml unchanged
- Use the Bash tool with a targeted YAML update (read, modify, write back)

Methodology-to-Mode Mapping Reference:

| Project State | Test Ratio | development_mode | Rationale |
|--------------|-----------|------------------|-----------|
| New (no code) | N/A | tdd | Clean slate, test-first development |
| Existing | >= 50% | tdd | Strong test base for test-first development |
| Existing | 10-49% | tdd | Partial tests, expand with test-first development |
| Existing | < 10% | ddd | No tests, gradual characterization test creation |

---

## Phase 4.1a: DB Detection

Purpose: Detect database technology from generated documentation and dependency
files to conditionally propose `/moai db init` in Next Steps.

[HARD] This phase runs automatically without user interaction. No AskUserQuestion is needed.

Steps:

1. Check `.moai/project/tech.md` exists. If not: set `detected_db=false` and skip to Phase 4.2.
2. Grep `tech.md` for DB engine keywords (case-insensitive). See Detection Keywords Reference → DB Engines section.
3. Glob for dependency manifests across all 16 supported languages (see Detection Keywords Reference → Dependency Files section).
4. For each found manifest file ≤ 1 MB: grep for ORM/ODM keywords relevant to that language (see Detection Keywords Reference → ORMs / ODMs by Language section).
5. Aggregate matches into: `{detected, matched_keywords[], source_files[], scanned_at, tech_md_hash}`.
6. Write state artifact at `.moai/state/db-detection.json`.
7. Proceed to Phase 4.2 with `detected_db` flag.

Guidance message on user selection (REQ-009):
When the user selects the `/moai db init` option from Next Steps, display this message before terminating `/moai project`:

> `/moai db init` will run 4 interview rounds (engine selection, connection config, schema survey, migration strategy) and create `.moai/project/db/` templates. Run it in your next turn.

Then terminate `/moai project` — do NOT auto-execute `/moai db init` (REQ-010). The user invokes it themselves in a subsequent turn.

File size limit: 1 MB. Skip any manifest file larger than 1 MB to avoid scanning generated lockfiles (e.g., `package-lock.json`, `poetry.lock`, `Cargo.lock`).

Tool choice: Grep with `-i` (case-insensitive) for keyword matching; Glob for manifest discovery.

Edge case (REQ-011): If `.moai/project/tech.md` does not exist (e.g., Phase 3 failed or was skipped), Phase 4.1a SHALL skip gracefully without error, set `detected_db=false`, and proceed to Phase 4.2 with the original three options unchanged.

State artifact schema (REQ-013): `.moai/state/db-detection.json` contains:

```json
{
  "detected": true,
  "matched_keywords": ["prisma", "postgresql"],
  "source_files": ["package.json", ".moai/project/tech.md"],
  "scanned_at": "2026-04-21T12:00:00Z",
  "tech_md_hash": "<sha256-of-tech.md-content>"
}
```

The `tech_md_hash` field enables stale-detection: if `tech.md` content changes between runs, Phase 4.2 can detect that the cached detection result is outdated and re-trigger Phase 4.1a.

---

## Phase 4: Completion

### Step 4.1: Content Summary Report

[HARD] Read the generated documents and present a structured summary to the user in conversation_language.

Read these files and extract key information:
- .moai/project/product.md → Project name, description, core features, target audience
- .moai/project/structure.md → Top-level directory structure, architecture pattern
- .moai/project/tech.md → Primary language, framework, key dependencies
- .moai/project/codemaps/ → Number of codemaps files generated (if any)

Display summary using this format:

```
Project Documentation Complete

product.md:
  - Project: [name]
  - Description: [1-2 sentence summary]
  - Core Features: [feature list]

structure.md:
  - Architecture: [pattern detected]
  - Key Directories: [top 3-5 directories with purposes]

tech.md:
  - Language: [primary language]
  - Framework: [framework name]
  - Key Dependencies: [top 3-5 packages]

Codemaps: [N files generated] in .moai/project/codemaps/
Development Mode: [tdd/ddd] (auto-configured in Phase 3.7)
```

### Step 4.2: Next Steps

[HARD] After displaying the summary, read the `detected_db` flag from `.moai/state/db-detection.json` (written by Phase 4.1a), then use AskUserQuestion to present conditional options based on the three-way branch below.

**Branch A — DB detected, `.moai/project/db/` does NOT exist (REQ-006, AC-6):**

When `detected_db` is true AND `.moai/project/db/` is absent, present these options:

- Initialize DB documentation (`/moai db init`) (Recommended): DB technology was detected in your project. Run `/moai db init` to create database schema documentation, connection config, and migration strategy through a 4-round interview. Recommended before creating SPECs that depend on your data model.
- Create SPEC: Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Done: Complete the project setup workflow.

When the user selects "Initialize DB documentation (`/moai db init`)": Display the guidance message from Phase 4.1a and terminate `/moai project`. Do NOT auto-execute `/moai db init`.

**Branch B — DB detected, `.moai/project/db/` already exists (REQ-007, AC-7):**

When `detected_db` is true AND `.moai/project/db/` already exists, present these options (existing order and Recommended flag preserved):

- Create SPEC (Recommended): Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Done: Complete the project setup workflow.
- Refresh DB documentation (`/moai db refresh`): DB documentation already exists. Run `/moai db refresh` to incorporate changes from an updated `tech.md` or schema evolution. This will update `.moai/project/db/` without re-running the full interview.

**Branch C — DB not detected (REQ-008, AC-8):**

When `detected_db` is false, present the original three options unchanged:

- Create SPEC (Recommended): Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Done: Complete the project setup workflow.

---

## Agent Chain Summary

- Phase 0-2: MoAI orchestrator (AskUserQuestion for all user interaction)
- Phase 1: Explore subagent (codebase analysis)
- Phase 3: manager-docs subagent (documentation generation)
- Phase 3.1: plan-auditor subagent (independent document audit, conditional)
- Phase 3.3: Explore + manager-docs subagents (codemaps generation via codemaps workflow)
- Phase 3.5: expert-devops subagent (optional LSP installation)
- Phase 3.7: MoAI orchestrator (automatic development_mode configuration, no user interaction)
- Phase 4.1a: MoAI orchestrator (automatic DB detection via Grep/Glob, no user interaction)

---

## Detection Keywords Reference

Phase 4.1a references the following keyword lists. All matching is case-insensitive. ORM/ODM matches are treated as stronger signals than DB engine name matches alone (mitigates false positives from documentation-only mentions).

### DB Engines

**Relational / SQL:**
- PostgreSQL
- MySQL
- MariaDB
- SQLite
- Oracle
- SQL Server / MSSQL
- CockroachDB
- Supabase
- Neon
- Planetscale

**NoSQL Document:**
- MongoDB
- Firestore
- Firebase
- Couchbase

**NoSQL Key-Value / Wide-column:**
- Redis
- DynamoDB
- Cassandra
- ScyllaDB
- Riak

**Search / Analytics:**
- Elasticsearch
- ClickHouse
- Snowflake
- InfluxDB

### Dependency Files (16 MoAI-supported languages + SQL standalone)

| Language (canonical name) | Dependency manifest files |
|---|---|
| go | `go.mod`, `go.sum` |
| python | `requirements.txt`, `pyproject.toml`, `Pipfile`, `setup.py` |
| typescript | `package.json`, `tsconfig.json` |
| javascript | `package.json` |
| rust | `Cargo.toml`, `Cargo.lock` |
| java | `pom.xml`, `build.gradle` |
| kotlin | `build.gradle.kts`, `build.gradle` |
| csharp | `*.csproj`, `packages.config`, `Directory.Packages.props` |
| ruby | `Gemfile`, `Gemfile.lock`, `*.gemspec` |
| php | `composer.json`, `composer.lock` |
| elixir | `mix.exs`, `mix.lock` |
| cpp | `CMakeLists.txt`, `conanfile.txt`, `conanfile.py`, `vcpkg.json` |
| scala | `build.sbt`, `project/plugins.sbt` |
| r | `DESCRIPTION`, `renv.lock` |
| flutter | `pubspec.yaml`, `pubspec.lock` |
| swift | `Package.swift`, `Podfile`, `Podfile.lock` |
| sql-standalone | `migrations/**/*.sql`, `db/migrate/**/*.sql`, `schema.sql` |

### ORMs / ODMs by Language

**Go:**
- GORM
- SQLc
- Ent
- mongo-go-driver
- sqlx

**Python:**
- SQLAlchemy
- Django ORM (django.db)
- Tortoise ORM
- Peewee
- python-oracledb
- motor (Mongo async)
- pymongo

**TypeScript / JavaScript:**
- Prisma
- TypeORM
- Drizzle
- Sequelize
- Mongoose
- Objection
- Kysely
- MikroORM

**Rust:**
- Diesel
- SQLx
- SeaORM
- mongodb (crate)
- tokio-postgres

**Java:**
- Hibernate
- JPA / jakarta.persistence
- Spring Data
- MyBatis
- jOOQ

**Kotlin:**
- Exposed
- Ktorm
- Hibernate (via JVM)
- JPA (via JVM)

**C#:**
- Entity Framework (EF Core)
- Dapper
- NHibernate
- LINQ to DB

**Ruby:**
- ActiveRecord
- Sequel
- Mongoid
- ROM-rb

**PHP:**
- Eloquent (Laravel)
- Doctrine
- Phinx
- CakePHP ORM

**Elixir:**
- Ecto

**C++:**
- SOCI
- ODB
- SQLite (direct, via CMake/conan)
- mongocxx

**Scala:**
- Slick
- Doobie
- Quill
- ScalikeJDBC

**R:**
- DBI
- dplyr (dbplyr backend)
- RPostgres
- RSQLite
- RMariaDB

**Flutter / Dart:**
- Drift (formerly Moor)
- sqflite
- hive
- isar
- objectbox

**Swift:**
- Core Data
- GRDB
- Realm
- SQLite.swift
- FluentKit (Vapor)

---

## Phase 5: Socratic Interview (Harness Activation)

Purpose: Conduct a 16-question / 4-round Socratic interview using `AskUserQuestion` to gather
project context required by `moai-meta-harness`. Answers are accumulated in an in-memory buffer
(no disk I/O) until Round 4 Q16 final confirmation (REQ-PH-001, REQ-PH-002, REQ-PH-010).

[HARD] Each round is exactly one `AskUserQuestion` call with up to 4 questions (C-PH-003).
[HARD] Each question's first option MUST be marked "(권장)" with a detailed description (C-PH-003).
[HARD] All question text and option labels MUST be in conversation_language (default: ko) (C-PH-004).
[HARD] No disk I/O until Round 4 Q16 "Confirm" answer is received (REQ-PH-010).

In-Memory Buffer Protocol:
- Maintain all 16 answers in memory across the 4 `AskUserQuestion` calls.
- On "Confirm" (Q16): call `Buffer.Commit()`, then proceed to write `.moai/harness/interview-results.md`.
- On "Restart" (Q16): clear the buffer and restart from Round 1.
- On "Abort" (Q16): call `Buffer.Abort()` — clears all answers, writes zero bytes to disk, and exits Phase 5.

---

### Round 1: Q1–Q4 (도메인 / 기술스택 / 규모 / 팀구성)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q1 — 도메인 (Project Domain)**

질문: 이 프로젝트의 주요 도메인은 무엇인가요?

옵션:
- (권장) 웹 (Web Application): 프론트엔드+백엔드 풀스택 또는 API 서비스. 사용자 대면 대시보드, SaaS, 이커머스 등에 최적. React/Vue/Next.js + REST/GraphQL 조합이 일반적.
- 모바일 (iOS): Swift + SwiftUI 또는 UIKit 기반 iOS 네이티브 앱. App Store 배포 대상. FaceID/HealthKit 등 iOS 전용 API 활용 가능.
- 모바일 (Android): Kotlin + Jetpack Compose 또는 XML 기반 Android 앱. Google Play 배포 대상.
- 기타 (Other): CLI 도구, 임베디드 시스템, 데스크톱 앱, 크로스플랫폼 (Flutter/React Native) 등 위 분류에 해당하지 않는 경우.

**Q2 — 기술스택 (Primary Technology Stack)**

질문: 주요 기술 스택은 무엇인가요?

옵션:
- (권장) TypeScript / JavaScript (Node.js + React/Next.js): 풀스택 JS 생태계. 프론트+백 코드 공유, 큰 npm 생태계, Vercel/AWS Lambda 배포 친화적.
- Go: 고성능 마이크로서비스, CLI, 클라우드 네이티브 바이너리. 단순 배포, 정적 컴파일, 낮은 메모리 사용.
- Python: AI/ML 워크로드, 백엔드 API (FastAPI/Django). 데이터 사이언스 라이브러리 풍부.
- 기타 (Swift / Kotlin / Rust / Java / C# 등): 위 3개에 해당하지 않는 언어. 구체적 언어를 직접 입력.

**Q3 — 규모 (Project Scale)**

질문: 프로젝트 규모는 어느 정도인가요?

옵션:
- (권장) MVP (1-3 모듈, 단기): 핵심 기능 1-3개로 빠르게 검증. 1-2주 내 첫 배포 목표. 기술 부채 최소화 우선.
- Small (4-8 모듈, 1-3개월): 안정화된 기능셋, 팀 2-4명, CI/CD 포함 구성.
- Medium (9-20 모듈, 3-12개월): 여러 도메인 레이어, 팀 5-10명, 마이크로서비스 또는 모듈 분리 고려.
- Large (20+ 모듈 또는 멀티팀): 조직 규모 제품, 복수 팀 협업, 플랫폼 엔지니어링 필요.

**Q4 — 팀구성 (Team Composition)**

질문: 팀 구성은 어떻게 되나요?

옵션:
- (권장) 솔로 개발자 (Solo developer): 1인 개발. 모든 역할 담당. 자동화와 AI 보조 도구로 생산성 보완.
- 소규모 팀 (2-4명): 풀스택 개발자 2-4명. 역할 유동적. 코드 리뷰 필수.
- 중간 팀 (5-10명): 프론트/백 분리, QA 포함. 명확한 소유권과 PR 프로세스 필요.
- 대규모 / 멀티팀: 10명 이상 또는 다수 팀. 아키텍처 가이드, API 계약, 플랫폼 레이어 필수.

---

### Round 2: Q5–Q8 (방법론 / 디자인툴 / UI복잡도 / 디자인시스템)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q5 — 방법론 (Development Methodology)**

질문: 주요 개발 방법론은 무엇인가요?

옵션:
- (권장) TDD (테스트 주도 개발): 테스트 먼저 작성 후 구현. RED-GREEN-REFACTOR 사이클. 새 기능 개발에 최적.
- DDD (도메인 주도 개발): 기존 코드베이스 리팩토링. ANALYZE-PRESERVE-IMPROVE 사이클. 레거시 코드에 최적.
- Agile / Scrum: 스프린트 기반 반복 개발. 백로그 관리, 데일리 스탠드업, 스프린트 리뷰.
- 기타 (Kanban / Waterfall / Ad-hoc): 위 방법론에 해당하지 않는 경우 직접 기술.

**Q6 — 디자인툴 (Design Tool)**

질문: UI/UX 디자인에 어떤 도구를 사용하나요?

옵션:
- (권장) Figma: 협업 디자인 도구. 디자인 토큰 추출, 컴포넌트 라이브러리, 개발자 핸드오프 지원.
- Sketch: macOS 전용 디자인 도구. 플러그인 생태계 풍부. Zeplin 핸드오프 많이 사용.
- Adobe XD: Adobe 생태계 통합. 프로토타이핑과 디자인 시스템 관리.
- 없음 / 코드 기반: 별도 디자인 툴 없이 코드로 직접 UI 구현. Storybook 등 컴포넌트 주도.

**Q7 — UI복잡도 (UI Complexity)**

질문: UI 복잡도는 어느 수준인가요?

옵션:
- (권장) 표준 (목록 + 폼 + 네비게이션): 일반적인 CRUD UI. 테이블, 폼, 모달, 내비게이션 바 수준.
- 단순 (정보성 페이지 / 랜딩): 마케팅 페이지, 대시보드 요약, 읽기 전용 뷰.
- 복잡 (데이터 시각화 / 드래그앤드롭): 차트, 그래프, 인터랙티브 에디터, 캔버스 기반 UI.
- 매우 복잡 (실시간 협업 / 3D / 게임): WebRTC, Three.js, 게임 UI 등 고도의 인터랙티비티.

**Q8 — 디자인시스템 (Design System)**

질문: 어떤 디자인 시스템을 사용할 예정인가요?

옵션:
- (권장) 기존 컴포넌트 라이브러리 (MUI / shadcn / Tailwind UI): 검증된 오픈소스 컴포넌트. 빠른 시작, 커스터마이징 가능.
- 커스텀 DTCG 토큰: W3C Design Token Community Group 표준. Figma 토큰 직접 추출, 완전 커스텀.
- 플랫폼 기본 (SwiftUI / Jetpack Compose / WinUI): 플랫폼 네이티브 UI. OS 가이드라인 자동 준수.
- 없음 / 미정: 디자인 시스템 없이 개별 스타일 적용. 추후 도입 예정.

---

### Round 3: Q9–Q12 (보안 / 성능 / 배포 / 외부통합)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q9 — 보안 (Security Requirements)**

질문: 주요 보안 요구사항은 무엇인가요?

옵션:
- (권장) 표준 인증 (JWT + OAuth2): 일반적인 웹/모바일 인증. Access/Refresh 토큰, 소셜 로그인 지원.
- 강화 보안 (OAuth + Keychain / Secure Enclave): iOS Keychain, Android Keystore, HSM 등 하드웨어 보안 요소 활용.
- 엔터프라이즈 (SSO / SAML / MFA): 기업 환경. Azure AD, Okta, LDAP 연동, 다중 인증.
- 최소 보안 (API Key 수준): 내부 도구, 프로토타입. 단순 API Key 또는 Basic Auth.

**Q10 — 성능 (Performance Target)**

질문: 성능 목표는 무엇인가요?

옵션:
- (권장) 일반 UI 반응성 (60fps, <200ms): 표준 앱 성능. 일반적인 CRUD 앱에 적합.
- 고성능 / 실시간 (<50ms): 금융, 게임, 실시간 협업. 최적화된 렌더링, 캐싱, WebSocket.
- 대용량 처리 (배치 / 스트리밍): 대규모 데이터 처리. 비동기 큐, 스트림 처리, 수평 확장.
- 저성능 환경 대응 (제한된 네트워크 / 구형 기기): 모바일 오프라인, IoT, 저사양 디바이스 지원.

**Q11 — 배포 (Deployment Target)**

질문: 어디에 배포할 예정인가요?

옵션:
- (권장) 클라우드 (AWS / GCP / Azure / Vercel): 관리형 클라우드. 오토스케일링, 관리형 DB, CDN.
- 앱 스토어 (App Store / Google Play): 모바일 앱 배포. 앱 심사, 버전 관리, 업데이트 정책 필요.
- 자체 서버 / On-premise: 자체 인프라. Docker + Kubernetes 또는 bare metal.
- 하이브리드 (클라우드 + 앱스토어): 모바일 앱 + 백엔드 API 조합.

**Q12 — 외부통합 (External Integrations)**

질문: 어떤 외부 시스템과 통합이 필요한가요?

옵션:
- (권장) 없음 / 표준 (결제 / 이메일 / SMS): Stripe, SendGrid, Twilio 등 범용 서비스 통합.
- 플랫폼 API (HealthKit / Maps / Push): iOS/Android 플랫폼 전용 API.
- 엔터프라이즈 시스템 (ERP / CRM / SAP): 기업 내부 시스템 연동. REST/SOAP/EDI.
- AI / ML 서비스 (OpenAI / Claude / Vision API): 외부 AI API 호출. 프롬프트 관리, 응답 처리.

---

### Round 4: Q13–Q16 (customization 범위 / 특수제약 / 우선순위 / 최종확인)

Present via `AskUserQuestion` — 4 questions, each with 4 options:

**Q13 — customization 범위 (Harness Customization Scope)**

질문: 프로젝트 전용 harness의 customization 범위는 어떻게 할까요?

옵션:
- (권장) 표준 (Standard): 도메인 특화 에이전트 2개 + 스킬 2개. 대부분의 프로젝트에 충분. moai-meta-harness가 답변 기반으로 최적 구성 자동 생성.
- 경량 (Minimal): 도메인 특화 스킬 1개만. 가장 빠른 setup. MVP 또는 소규모 프로젝트에 적합.
- 심화 (Thorough): 에이전트 3개 이상 + 스킬 3개 이상 + design-extension 포함. 복잡한 도메인에 최적.
- 전체 커스텀 (Advanced / full custom): 모든 요소를 완전 커스텀. design-extension.md 추가 생성 (REQ-PH-012). 고급 사용자용.

**Q14 — 특수제약 (Special Constraints)**

질문: 프로젝트에 특수 제약 사항이 있나요?

옵션:
- (권장) 없음 (No special constraints): 일반적인 제약만 적용. harness가 표준 패턴 사용.
- 최소 OS 버전 (iOS 17+ / Android 12+ 등): 플랫폼 최소 버전 제약. 하위 호환 API 사용 제한.
- 규정 준수 (HIPAA / GDPR / SOC2): 데이터 보호 규정. 암호화, 감사 로그, 데이터 거주지 제약.
- 기타 제약 (오프라인 필수 / 특정 하드웨어 / 정부 규격): 위에 해당하지 않는 특수 제약 사항.

**Q15 — 우선순위 (Harness Quality Level)**

질문: Harness 품질 수준(harness level)을 선택해 주세요.

옵션:
- (권장) standard: 기본 품질 게이트. 대부분의 프로젝트에 적합. 빠른 실행과 충분한 검증의 균형.
- thorough: 전체 evaluator-active + TRUST 5 검증. 복잡한 SPEC 또는 엔터프라이즈 프로젝트에 권장.
- minimal: 빠른 검증만. 단순 변경 또는 프로토타입에 적합. 일부 품질 게이트 생략.
- custom: 직접 구성. `.moai/config/sections/harness.yaml`에서 세부 설정 가능.

**Q16 — 최종확인 (Final Confirmation)**

질문: 위 16개 답변을 바탕으로 프로젝트 전용 harness를 생성할까요?

옵션:
- (권장) Confirm — 생성 진행: 모든 답변을 확인했습니다. `.moai/harness/interview-results.md`에 결과를 기록하고 Phase 6 (meta-harness 호출)으로 진행합니다.
- Restart — 처음부터 다시: Round 1부터 인터뷰를 다시 시작합니다. 이전 답변은 모두 초기화됩니다.
- Abort — 취소: 인터뷰를 중단합니다. 어떠한 파일도 생성되지 않습니다 (REQ-PH-010).

**Q16 Branch Logic:**
- "Confirm" → `Buffer.Commit()` 호출 → `.moai/harness/interview-results.md` 작성 → Phase 6 (meta-harness)으로 진행.
- "Restart" → `Buffer.Abort()` 후 `NewBuffer()` → Round 1부터 재시작.
- "Abort" → `Buffer.Abort()` 호출 → 디스크에 0 파일 작성 → Phase 5 종료 (zero disk writes, REQ-PH-010).

---

## Phase 6: meta-harness Invocation

Purpose: Call `Skill("moai-meta-harness")` with the 16 answers collected in Phase 5,
generating project-specific dynamic harness artifacts in the user area
(REQ-PH-004, T-P2-01).

[HARD] This phase MUST run the FROZEN guard (`EnsureAllowed`) as the **first check**
before any write attempt. Paths in `.claude/agents/moai/`, `.claude/skills/moai-*/`,
or `.claude/rules/moai/` are permanently FROZEN and must be rejected immediately.

[HARD] If meta-harness generation fails mid-way, `CleanupOnFailure` MUST remove all
partial artifacts written so far (REQ-PH-010).

### 6.1 Pre-Condition

- Phase 5 Round 4 Q16 answer is "Confirm" → `Buffer.Commit()` has been called.
- `.moai/harness/interview-results.md` has been written by `WriteResultsToFile`.

### 6.2 Answer-to-Context Schema

Convert the 16 in-memory answers to a structured prompt context before invoking
`Skill("moai-meta-harness")`. Each question maps to a named field:

```yaml
# Answer-to-context schema (YAML form)
context:
  # Round 1 — Domain & Technology
  domain:            # Q01 answer text (e.g., "모바일 (iOS)")
  tech_stack:        # Q02 answer text (e.g., "Swift + SwiftUI")
  project_scale:     # Q03 answer text (e.g., "MVP (1-3 모듈, 단기)")
  team_composition:  # Q04 answer text (e.g., "솔로 개발자")

  # Round 2 — Methodology & Design
  methodology:       # Q05 answer text (e.g., "TDD")
  design_tool:       # Q06 answer text (e.g., "Figma")
  ui_complexity:     # Q07 answer text (e.g., "표준 (목록 + 폼 + 네비게이션)")
  design_system:     # Q08 answer text (e.g., "커스텀 DTCG 토큰")

  # Round 3 — Security, Performance, Deployment
  security:          # Q09 answer text (e.g., "강화 보안 (OAuth + Keychain / Secure Enclave)")
  performance:       # Q10 answer text (e.g., "일반 UI 반응성 (60fps, <200ms)")
  deployment:        # Q11 answer text (e.g., "앱 스토어 (App Store / Google Play)")
  integrations:      # Q12 answer text (e.g., "플랫폼 API (HealthKit / Maps / Push)")

  # Round 4 — Customization & Final Confirmation
  customization_scope: # Q13 answer text (e.g., "표준 (Standard)")
  special_constraints: # Q14 answer text (e.g., "최소 OS 버전 (iOS 17+ / Android 12+ 등)")
  harness_level:       # Q15 answer text (e.g., "standard")
  final_confirmation:  # Q16 answer text — always "Confirm" at this point
```

### 6.3 Invocation Protocol

```
Skill("moai-meta-harness") with:
  - context: <structured answer map above>
  - project_root: <absolute path to project root>
  - spec_id: <SPEC-PROJ-INIT-NNN from interview-results.md>
  - conversation_language: <ko|en|ja|zh>
  - harness_level: <Q15 answer: minimal|standard|thorough>
  - design_extension: <true if Q13 == "전체 커스텀 (Advanced / full custom)", else false>
```

### 6.4 Expected Outputs

After successful meta-harness invocation, the following artifacts must exist
in the **user area** (FROZEN guard pre-verified):

| Artifact | Path | Required |
|----------|------|----------|
| Architect agent | `.claude/agents/my-harness/<domain>-architect.md` | Always |
| Engineer agent | `.claude/agents/my-harness/<domain>-engineer.md` | Always |
| Patterns skill | `.claude/skills/my-harness-<domain>-patterns/SKILL.md` | Always |
| Best-practices skill | `.claude/skills/my-harness-<domain>-best-practices/SKILL.md` | Always |
| Harness directory | `.moai/harness/` | Always |
| Design extension | `.moai/harness/design-extension.md` | Q13 == Advanced only |

All write paths must pass `EnsureAllowed(path)` before the file is created.
Any `FrozenViolationError` causes immediate abort + `CleanupOnFailure`.

### 6.5 Failure Handling

If `Skill("moai-meta-harness")` returns an error or partial output:

1. Call `CleanupOnFailure(tracker, err)` — removes all tracked partial files.
2. Surface the error to the user with a clear message.
3. Do NOT proceed to Phase 7 (5-Layer Activation).

---

Version: 2.4.0
Last Updated: 2026-04-27
SPEC: SPEC-PROJECT-DB-HINT-001, SPEC-V3R3-PROJECT-HARNESS-001
