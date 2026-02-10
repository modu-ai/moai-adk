---
name: moai-workflow-project
description: >
  Generates project documentation from codebase analysis or user input.
  Creates product.md, structure.md, and tech.md in .moai/project/ directory.
  Supports new and existing project types with LSP server detection.
  Use when initializing projects or generating project documentation.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-07"
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

Purpose: Generate project documentation through smart questions and codebase analysis. Creates product.md, structure.md, and tech.md in .moai/project/ directory.

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

## Phase 0.5: New Project Requirements Collection (New Projects Only)

Goal: Understand user requirements through smart questions to generate accurate project documentation.

[HARD] All questions MUST use AskUserQuestion in user's conversation_language.

Question 1 - Project Purpose (AskUserQuestion):

Header: "Project Type"

- Web Application (Recommended): Build a frontend, backend, or full-stack web application. Includes HTML/CSS/JS frontend with a server-side backend. Best for websites, dashboards, and web-based tools.
- API Service: Build a REST API, GraphQL endpoint, or microservices backend. Best for mobile app backends, third-party integrations, and data services.
- CLI Tool: Build a command-line utility or automation script. Best for developer tools, system utilities, and build automation.
- Library/Package: Build a reusable code library, SDK, or framework. Best for shared utilities, open-source packages, and internal toolkits.

Question 2 - Primary Language (AskUserQuestion):

Header: "Language"

- TypeScript/JavaScript (Recommended): Most versatile choice for web development. Works for frontend (React, Vue), backend (Node.js, Bun), and full-stack applications. Largest ecosystem of packages and tools.
- Python: Excellent for backend APIs (FastAPI, Django), data science, AI/ML, and automation scripts. Easy to learn with extensive library support.
- Go: Best for high-performance microservices, CLI tools, and cloud-native applications. Fast compilation, strong concurrency support, and simple deployment.
- Other: Choose this for Rust, Java, Kotlin, Ruby, Swift, C#, or other languages. You will be asked to specify.

Question 3 - Project Description (AskUserQuestion with free text via "Other"):

Header: "Description"

Present a question asking the user to describe their project. The user provides free text including:
- Project name
- Main features or goals
- Target users or audience

Question 4 - Key Features (AskUserQuestion, multiSelect: true):

Header: "Features"

Based on the selected project type and language, present relevant feature options:

For Web Applications:
- Authentication: User login, registration, session management
- Database: Data persistence with ORM and migrations
- API Integration: External API calls and webhooks
- Real-time: WebSocket or SSE for live updates

For API Services:
- REST Endpoints: CRUD operations with validation
- Authentication: JWT, OAuth, API keys
- Database: SQL or NoSQL data layer
- Documentation: OpenAPI/Swagger auto-generation

For CLI Tools:
- Interactive prompts: User input collection with TUI
- Configuration: Config file management (YAML, JSON, TOML)
- Output formatting: Tables, colors, progress bars
- Plugin system: Extensible architecture

For Library/Package:
- Type safety: Full type annotations
- Documentation: Auto-generated API docs
- Testing: Unit and integration test suite
- CI/CD: Automated publishing pipeline

After collection, use the gathered information to generate documentation and proceed to Phase 3 (skip Phase 1 and 2 since there is no existing code to analyze).

---

## Phase 1: Codebase Analysis (Existing Projects Only)

[HARD] Delegate codebase analysis to the Explore subagent.

[SOFT] Apply --ultrathink for comprehensive analysis.

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

## Phase 3.5: Development Environment Check

Goal: Verify LSP servers are installed for the detected technology stack.

Language-to-LSP Mapping (16 languages):

- Python: pyright or pylsp (check: which pyright)
- TypeScript/JavaScript: typescript-language-server (check: which typescript-language-server)
- Go: gopls (check: which gopls)
- Rust: rust-analyzer (check: which rust-analyzer)
- Java: jdtls (Eclipse JDT Language Server)
- Ruby: solargraph (check: which solargraph)
- PHP: intelephense (check via npm)
- C/C++: clangd (check: which clangd)
- Kotlin: kotlin-language-server
- Scala: metals
- Swift: sourcekit-lsp
- Elixir: elixir-ls
- Dart/Flutter: dart language-server (bundled with Dart SDK)
- C#: OmniSharp or csharp-ls
- R: languageserver (R package)
- Lua: lua-language-server

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
- Set `development_mode: "hybrid"` (TDD for new features, DDD for structure)
- Rationale: New projects benefit from test-first development for new features while maintaining DDD structure for overall architecture

For Existing Projects (Phase 0 classified as "Existing Project"):
- Step 1: Check for existing test files using Glob patterns (*_test.go, *_test.py, *.test.ts, *.test.js, *.spec.ts, *.spec.js, test_*.py, tests/, __tests__/, spec/)
- Step 2: Estimate test coverage level based on test file count relative to source file count:
  - No test files found (0%): Set `development_mode: "ddd"` (need characterization tests first)
  - Few test files (< 10% ratio): Set `development_mode: "ddd"` (insufficient coverage for TDD)
  - Moderate test files (10-49% ratio): Set `development_mode: "hybrid"` (expand with DDD, new code with TDD)
  - Good test files (>= 50% ratio): Set `development_mode: "hybrid"` (sufficient base, hybrid approach)

Implementation:
- Read current `.moai/config/sections/quality.yaml`
- Update only the `constitution.development_mode` field
- Preserve all other settings in quality.yaml unchanged
- Use the Bash tool with a targeted YAML update (read, modify, write back)

Methodology-to-Mode Mapping Reference:

| Project State | Test Ratio | development_mode | Rationale |
|--------------|-----------|------------------|-----------|
| New (no code) | N/A | hybrid | Clean slate, TDD for features + DDD structure |
| Existing | >= 50% | hybrid | Sufficient test base for hybrid development |
| Existing | 10-49% | hybrid | Partial tests, expand with DDD then TDD for new |
| Existing | < 10% | ddd | No tests, gradual characterization test creation |

---

## Phase 4: Completion

Display completion message in user's conversation_language:

- Files created: List generated files
- Location: .moai/project/
- Status: Success or partial completion

Next Steps (AskUserQuestion):

- Write SPEC (Recommended): Execute /moai plan to define your first feature specification. This is the natural next step after project setup - it creates a detailed plan for what you want to build.
- Review Documentation: Open the generated product.md, structure.md, and tech.md files for review and manual editing. Choose this if you want to verify or customize the generated content.
- Start New Session: Clear the current context and start fresh. Choose this if you want to work on something completely different.

---

## Agent Chain Summary

- Phase 0-2: MoAI orchestrator (AskUserQuestion for all user interaction)
- Phase 1: Explore subagent (codebase analysis)
- Phase 3: manager-docs subagent (documentation generation)
- Phase 3.5: expert-devops subagent (optional LSP installation)
- Phase 3.7: MoAI orchestrator (automatic development_mode configuration, no user interaction)

---

Version: 2.1.0
Last Updated: 2026-02-10
