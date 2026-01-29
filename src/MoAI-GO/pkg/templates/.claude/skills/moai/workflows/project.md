# Workflow: project - Project Documentation Generation

Purpose: Generate project documentation from codebase analysis. Creates product.md, structure.md, and tech.md in .moai/project/ directory.

---

## Phase 0: Project Type Detection

[HARD] Ask project type FIRST before any analysis using AskUserQuestion.

Question: What type of project are you working on?

Options (in user's conversation_language):

- New Project: Starting from scratch, will collect project information interactively
- Existing Project: Documenting an existing codebase, will analyze code automatically

Routing:

- New Project selected: Proceed to Phase 0.5
- Existing Project selected: Proceed to Phase 1

---

## Phase 0.5: New Project Information Collection (New Projects Only)

Goal: Collect project details when no existing code is available to analyze.

Question 1 - Project Purpose (AskUserQuestion):

- Web Application: Frontend, backend, or full-stack web app
- API Service: REST API, GraphQL, or microservices
- CLI Tool: Command-line utility or automation tool
- Library/Package: Reusable code library or SDK

Question 2 - Primary Language (AskUserQuestion):

- Python: Backend, data science, automation
- TypeScript/JavaScript: Web, Node.js, frontend
- Go: High-performance services, CLI tools
- Other: Rust, Java, Ruby, etc. (will ask for details)

Question 3 - Project Description (free text input):

- Project name
- Main features or goals
- Target users

After collection, generate starter documentation from user input and proceed to Phase 4.

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

- Proceed with documentation generation
- Review specific analysis details first
- Cancel and adjust project configuration

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

## Phase 4: Completion

Display completion message in user's conversation_language:

- Files created: List generated files
- Location: .moai/project/
- Status: Success or partial completion

Next Steps (AskUserQuestion):

- Write SPEC: Execute /moai plan to define feature specifications
- Review Documentation: Open generated files for review
- Start New Session: Clear context and start fresh

---

## Agent Chain Summary

- Phase 0-2: MoAI orchestrator (AskUserQuestion for all user interaction)
- Phase 1: Explore subagent (codebase analysis)
- Phase 3: manager-docs subagent (documentation generation)
- Phase 3.5: expert-devops subagent (optional LSP installation)

---

Version: 1.0.0
Last Updated: 2026-01-28
