---
description: "Project Phase 3/3.1/3.3/3.5/3.7/4.1a/4 — Documentation generation, audit, codemaps, LSP check, dev mode config, DB detection, and completion"
user-invocable: false
metadata:
  parent: moai-workflow-project
  phase: "Phase 3/3.1/3.3/3.5/3.7/4.1a/4: Documentation Generation and Completion"
---

<!-- TRACE PROBE: workflow-split baseline trace mechanism -->
<!-- Activated by MOAI_TRACE_PHASES=1 environment variable -->

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
- `thorough`: Run plan-auditor + cross-validate with sync-auditor

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
`.claude/rules/moai/development/coding-standards.md` § Language Policy
(16-language neutrality contract), all 16 languages are treated as equal
first-class citizens; the user's project marker files determine which
server(s) actually spawn at runtime.

If LSP server is NOT installed, present AskUserQuestion:

- Continue without LSP: Proceed to completion
- Show installation instructions: Display setup guide for detected language
- Auto-install now: Use a per-spawn `Agent(general-purpose)` devops specialist to install (requires confirmation; devops whitelist per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C row 10)

---

## Phase 3.7: Development Methodology Auto-Configuration

Goal: Automatically set the `development_mode` in `.moai/config/sections/quality.yaml` based on the project analysis results from Phase 0 and Phase 1.

[HARD] This phase runs automatically without user interaction. No AskUserQuestion is needed.

Auto-Detection Logic:

For New Projects (Phase 0 classified as "New Project"):
- Set `development_mode: "tdd"` (test-first development)
- Rationale: New projects benefit from test-first development with clean RED-GREEN-REFACTOR cycles

For Existing Projects (Phase 0 classified as "Existing Project"):
- Step 1: Check for existing test files using Glob patterns across all 16 MoAI-supported languages (alphabetical): C++ (*_test.cpp, *_test.cc), C# (*Test.cs, *Tests.cs), Elixir (*_test.exs), Flutter (*_test.dart), Go (*_test.go), Java (*Test.java, *Tests.java), JavaScript (*.test.js, *.spec.js), Kotlin (*Test.kt), PHP (*Test.php), Python (*_test.py, test_*.py), R (test-*.R), Ruby (*_spec.rb, *_test.rb), Rust (tests/*.rs), Scala (*Test.scala, *Spec.scala), Swift (*Tests.swift), TypeScript (*.test.ts, *.spec.ts) — plus common test directories (tests/, __tests__/, spec/, test/)
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

## Detection Keywords Reference

Full DB engine keywords, dependency manifest files (all 16 MoAI-supported languages), and ORM/ODM lists used by Phase 4.1a below.

**DB engine keywords** (grepped against `tech.md`, case-insensitive): Relational/SQL (PostgreSQL, MySQL, MariaDB, SQLite, Oracle, SQL Server, CockroachDB, Supabase, Neon, Planetscale), NoSQL Document (MongoDB, Firestore, Firebase, Couchbase), NoSQL Key-Value (Redis, DynamoDB, Cassandra, ScyllaDB, Riak), Search/Analytics (Elasticsearch, ClickHouse, Snowflake, InfluxDB).

**Dependency manifest + ORM/ODM keywords per language** (alphabetical):

| Language | Dependency manifest(s) | Common ORM/ODM keywords |
|----------|------------------------|--------------------------|
| C++ | `conanfile.txt`, `vcpkg.json`, `CMakeLists.txt` | sqlite3, soci, odb |
| C# | `*.csproj`, `packages.config` | entityframework, dapper |
| Elixir | `mix.exs` | ecto |
| Flutter | `pubspec.yaml` | sqflite, drift, isar |
| Go | `go.mod` | gorm, ent, sqlx, sqlc |
| Java | `pom.xml`, `build.gradle` | hibernate, jpa, mybatis |
| JavaScript | `package.json` | sequelize, mongoose, knex |
| Kotlin | `build.gradle.kts` | exposed, ktorm, hibernate |
| PHP | `composer.json` | doctrine, eloquent |
| Python | `requirements.txt`, `pyproject.toml`, `Pipfile` | sqlalchemy, django.db, peewee, tortoise |
| R | `DESCRIPTION` | dbplyr, dbi |
| Ruby | `Gemfile` | activerecord, sequel, mongoid |
| Rust | `Cargo.toml` | diesel, sea-orm, sqlx |
| Scala | `build.sbt` | slick, doobie, quill |
| Swift | `Package.swift`, `Podfile` | coredata, grdb, realm |
| TypeScript | `package.json` | prisma, typeorm, sequelize, mongoose, drizzle |

---

## Phase 4.1a: DB Detection

Purpose: Detect database technology from generated documentation and dependency
files. Detected metadata is consumed by sync workflow Phase 0.08 (DB Schema Doc
Check) to drive automatic refresh via `moai hook db-schema-sync` when
`db.auto_sync.enabled: true` is set in `.moai/config/sections/db.yaml` (the
`auto_sync:` key is a nested object — `enabled:` is its toggle sub-key,
alongside `debounce_seconds`, `require_user_approval`, and `excluded_patterns`).

[HARD] This phase runs automatically without user interaction. No AskUserQuestion is needed.

Steps:

1. Check `.moai/project/tech.md` exists. If not: set `detected_db=false` and skip to Phase 4.2.
2. Grep `tech.md` for DB engine keywords (case-insensitive). See Detection Keywords Reference above.
3. Glob for dependency manifests across all 16 supported languages (see Detection Keywords Reference above).
4. For each found manifest file ≤ 1 MB: grep for ORM/ODM keywords relevant to that language.
5. Aggregate matches into: `{detected, matched_keywords[], source_files[], scanned_at, tech_md_hash}`.
6. Write state artifact at `.moai/state/db-detection.json`.
7. Proceed to Phase 4.2 with `detected_db` flag.

When `detected_db=true`, Phase 4.2 (Next Steps) emits a guidance note to enable
`db.auto_sync.enabled: true` in `.moai/config/sections/db.yaml`. The user opts in once,
and subsequent `/moai sync` runs automatically refresh `.moai/project/db/` derived
docs (schema.md, erd.mmd, migrations.md) via Phase 0.08 → `moai hook db-schema-sync`.

The `/moai db` slash command was retired (Bundle A, 2026-05-16). Initial DB
documentation scaffolding is now handled by `.moai/project/db/` templates created
on first sync when `db.enabled: true`.

File size limit: 1 MB. Skip any manifest file larger than 1 MB to avoid scanning generated lockfiles (e.g., `package-lock.json`, `poetry.lock`, `Cargo.lock`).

Tool choice: Grep with `-i` (case-insensitive) for keyword matching; Glob for manifest discovery.

Edge case: If `.moai/project/tech.md` does not exist (e.g., Phase 3 failed or was skipped), Phase 4.1a SHALL skip gracefully without error, set `detected_db=false`, and proceed to Phase 4.2 with the original three options unchanged.

State artifact schema: `.moai/state/db-detection.json` contains:

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

**Branch A — DB detected, `.moai/project/db/` does NOT exist:**

When `detected_db` is true AND `.moai/project/db/` is absent, present these options:

- Enable automatic DB doc sync (Recommended): DB technology was detected in your project. Set `db.enabled: true` and `db.auto_sync.enabled: true` in `.moai/config/sections/db.yaml` (create the file if absent). Subsequent `/moai sync` runs will automatically generate and refresh `.moai/project/db/` via Phase 0.08 (`moai hook db-schema-sync`). Recommended before creating SPECs that depend on your data model.
- Create SPEC: Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Generate project-specific harness: Proceed to Phase 5/6 (`project/meta-harness.md`) to build a domain-specific harness (agents + skills) tailored to this project via the v4 harness Builder.
- Done: Complete the project setup workflow.

When the user selects "Enable automatic DB doc sync": Display guidance to edit `.moai/config/sections/db.yaml` and then run `/moai sync` on the next milestone. Do NOT auto-modify the config file.

**Branch B — DB detected, `.moai/project/db/` already exists:**

When `detected_db` is true AND `.moai/project/db/` already exists, present these options (existing order and Recommended flag preserved):

- Create SPEC (Recommended): Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Generate project-specific harness: Proceed to Phase 5/6 (`project/meta-harness.md`) to build a domain-specific harness (agents + skills) tailored to this project via the v4 harness Builder.
- Done: Complete the project setup workflow.
- Verify auto-sync enabled: DB documentation already exists. Confirm `db.auto_sync.enabled: true` in `.moai/config/sections/db.yaml`. When set, subsequent `/moai sync` runs automatically refresh `.moai/project/db/` via Phase 0.08 (`moai hook db-schema-sync`) on detected migration changes.

**Branch C — DB not detected:**

When `detected_db` is false, present the original three options plus the harness-generation option:

- Create SPEC (Recommended): Run `/moai plan` to define your first feature specification. This is the natural next step after project setup.
- Review and Edit Documentation: Open the generated files for review and manual editing before proceeding.
- Generate project-specific harness: Proceed to Phase 5/6 (`project/meta-harness.md`) to build a domain-specific harness (agents + skills) tailored to this project via the v4 harness Builder.
- Done: Complete the project setup workflow.

---

## Agent Chain Summary

- Phase 0-2: MoAI orchestrator (AskUserQuestion for all user interaction)
- Phase 1: Explore subagent (codebase analysis)
- Phase 3: manager-docs subagent (documentation generation)
- Phase 3.1: plan-auditor subagent (independent document audit, conditional)
- Phase 3.3: Explore + manager-docs subagents (codemaps generation via codemaps workflow)
- Phase 3.5: per-spawn `Agent(general-purpose)` devops specialist (optional LSP installation)
- Phase 3.7: MoAI orchestrator (automatic development_mode configuration, no user interaction)
- Phase 4.1a: MoAI orchestrator (automatic DB detection via Grep/Glob, no user interaction)
