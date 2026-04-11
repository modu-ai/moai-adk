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

[HARD] After displaying the summary, use AskUserQuestion to ask about next steps.

Next Steps (AskUserQuestion):

- Create SPEC (Recommended): Run /moai plan to define your first feature specification. This is the natural next step after project setup.
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

---

Version: 2.1.0
Last Updated: 2026-02-10
