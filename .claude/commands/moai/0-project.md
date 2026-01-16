---
description: "Generate project documentation from codebase analysis"
argument-hint: ""
type: workflow
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git branch --show-current

## Essential Files

@.moai/config/sections/language.yaml
@.moai/project/product.md
@.moai/project/structure.md
@.moai/project/tech.md

---

# MoAI-ADK Step 0: Generate Project Documentation

User Interaction Architecture: AskUserQuestion tool must be used at COMMAND level only, not within subagents. Subagents invoked via Task() operate in isolated, stateless contexts and cannot interact with users directly.

Architecture: Commands delegate to Agents, which coordinate Skills. This command orchestrates exclusively through Task() tool.

Workflow Integration: This command implements Step 0 of Alfred's three-step execution model (Understand-Plan-Execute). See CLAUDE.md for complete workflow details.

---

## Command Purpose

Generate project documentation by analyzing the existing codebase. This command creates three documentation files:

- `.moai/project/product.md` - Product overview, features, and user value
- `.moai/project/structure.md` - Project architecture and directory organization
- `.moai/project/tech.md` - Technology stack, dependencies, and technical decisions

WHY: Documentation generation is separated from project initialization to allow CLI-based setup.

IMPACT: Users can generate accurate project documentation after codebase development.

Note: Project configuration (language, service, pricing, git mode) is handled by `moai init` CLI command.

---

## Execution Modes

### Mode 1: Fresh Documentation (Default)

When `.moai/project/` directory is empty or missing:

1. Analyze codebase using Explore agent
2. Generate all three documentation files
3. Confirm completion to user

### Mode 2: Update Documentation

When documentation files already exist:

1. Read existing documentation
2. Analyze codebase for changes
3. Ask user which files to regenerate
4. Update selected files while preserving manual edits

---

## PHASE 1: Codebase Analysis

Goal: Understand the project structure and technology stack.

### Step 1: Invoke Explore Agent

[HARD] Delegate codebase analysis to the Explore subagent.

Use the Explore subagent to analyze the codebase with the following objectives:

Analysis Objectives:

1. Project Structure: Identify main directories, entry points, and architectural patterns
2. Technology Stack: Detect programming languages, frameworks, and key dependencies
3. Core Features: Identify main functionality and business logic locations
4. Build System: Detect build tools, package managers, and scripts

Output Format:

The Explore agent should return a structured analysis including:

- Primary Language: Main programming language used
- Framework: Web framework or application framework
- Architecture Pattern: MVC, Clean Architecture, Microservices, etc.
- Key Directories: Source, tests, configuration, documentation
- Dependencies: Major libraries and their purposes
- Entry Points: Main files or startup scripts

WHY: Accurate codebase analysis ensures documentation reflects actual project state.

---

## PHASE 2: User Confirmation

Goal: Confirm analysis results and get user approval before generating documentation.

### Step 1: Present Analysis Results

Use AskUserQuestion to present the analysis summary and get approval:

Question: Based on codebase analysis, here is the project summary. Should I proceed with documentation generation?

Present in user's conversation_language:

- Detected Language: [language]
- Framework: [framework]
- Architecture: [pattern]
- Key Features: [features list]

Options:

- Proceed with documentation generation
- Review specific analysis details first
- Cancel and adjust project configuration

### Step 2: Handle User Response

If user selects "Review details":

- Provide detailed analysis breakdown
- Allow corrections via follow-up AskUserQuestion

If user selects "Proceed":

- Continue to Phase 3

If user selects "Cancel":

- Exit with guidance to modify configuration

---

## PHASE 3: Documentation Generation

Goal: Generate project documentation files.

### Step 1: Invoke Manager-Docs Agent

[HARD] Delegate documentation generation to the manager-docs subagent.

Use the manager-docs subagent to generate documentation with the following parameters:

- Analysis Results: Pass the Explore agent analysis
- User Confirmation: Pass approved project summary
- Output Directory: `.moai/project/`
- Language: User's conversation_language from config

Documentation Files:

1. product.md:
   - Project name and description
   - Target audience
   - Core features and benefits
   - Use cases and examples

2. structure.md:
   - Directory tree visualization
   - Purpose of each major directory
   - Key file locations
   - Module organization

3. tech.md:
   - Technology stack overview
   - Framework and library choices with rationale
   - Development environment requirements
   - Build and deployment configuration

WHY: Manager-docs agent has specialized expertise in technical writing and documentation standards.

---

## PHASE 3.5: Development Environment Check

Goal: Verify LSP servers are installed for detected tech stack and guide installation if needed.

### Step 1: Check LSP Server Status

Based on the detected primary language from Phase 1, check if corresponding LSP server is available.

Language-to-LSP Mapping:

- Python: pyright or pylsp (check with `which pyright` or `which pylsp`)
- TypeScript/JavaScript: typescript-language-server (check with `which typescript-language-server`)
- Go: gopls (check with `which gopls`)
- Rust: rust-analyzer (check with `which rust-analyzer`)
- Java: jdtls (Eclipse JDT Language Server)
- Ruby: solargraph (check with `which solargraph`)
- PHP: intelephense (check via npm)
- C/C++: clangd (check with `which clangd`)

### Step 2: Present LSP Status to User

If LSP server is NOT installed, inform user with installation guidance:

Use conversation_language to display:

LSP Server Status Message Format:

```
Development Environment Check:

Detected Language: [language]
LSP Server: [server_name]
Status: Not Installed

For better code intelligence (autocomplete, diagnostics, go-to-definition),
install the recommended LSP server:

[Installation command based on language]
```

Installation Commands by Language:

Python:

- pyright: `npm install -g pyright` or `pip install pyright`
- pylsp: `pip install python-lsp-server`

TypeScript/JavaScript:

- `npm install -g typescript typescript-language-server`

Go:

- `go install golang.org/x/tools/gopls@latest`

Rust:

- Via rustup: `rustup component add rust-analyzer`
- Or standalone: Download from GitHub releases

Ruby:

- `gem install solargraph`

PHP:

- `npm install -g intelephense`

C/C++:

- macOS: `brew install llvm` (includes clangd)
- Ubuntu: `apt install clangd`

### Step 3: Continue or Assist Installation

Use AskUserQuestion:

Question: LSP server for [language] is not detected. Would you like assistance?

Options:

- Continue without LSP - Proceed to completion
- Show installation instructions - Display detailed setup guide
- Auto-install now - Attempt automatic installation (requires confirmation)

If user selects "Auto-install now":

- Use expert-devops subagent to execute installation
- Verify installation success
- Report result to user

WHY: LSP servers significantly improve development experience with Claude Code by providing accurate diagnostics.

IMPACT: Missing LSP reduces code intelligence quality and may cause incomplete error detection.

---

## PHASE 4: Completion

Goal: Confirm documentation generation and guide user to next steps.

### Step 1: Display Results

Show completion message in user's language:

- Files created: List of generated files
- Location: `.moai/project/`
- Status: Success or partial completion

### Step 2: Offer Next Steps

Use AskUserQuestion:

Question: Project documentation generated. What would you like to do next?

Options:

- Write SPEC - Execute /moai:1-plan to define feature specifications
- Review Documentation - Open generated files for review
- Start New Session - Clear context and start fresh

---

## Critical Rules

### Tool Usage Guidelines

This command has access to all tools for flexibility:

- Task() for agent orchestration (recommended for complex tasks)
- AskUserQuestion() for user interaction at command level
- Read, Write, Edit, Bash, Glob, Grep for direct operations when needed

Agent delegation is recommended for complex tasks that benefit from specialized expertise. Direct tool usage is permitted when appropriate for simpler operations.

### Language Handling

[HARD] Always use user's conversation_language for all output and prompts.

Read language from `.moai/config/sections/language.yaml` before starting.

WHY: User comprehension requires language consistency.

### User Interaction

[HARD] Use AskUserQuestion for ALL user interaction at COMMAND level only.

[HARD] No emoji characters in AskUserQuestion fields.

[HARD] Maximum 4 options per AskUserQuestion question.

WHY: Subagents cannot interact with users directly; emoji parsing varies across platforms.

### Agent Delegation

[HARD] Delegate ALL execution to specialized agents.

Agent Chain:

1. Explore agent: Codebase analysis
2. Manager-docs agent: Documentation generation

WHY: Agent provides specialized expertise and validation.

---

## Output Format

### Output Format Rules

[HARD] User-Facing Reports: Always use Markdown formatting for user communication. Never display XML tags to users.
WHY: Users expect readable formatted text, not markup
IMPACT: XML tags in user output create confusion and reduce comprehension

[HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only.
WHY: XML structure enables automated parsing for downstream agent coordination
IMPACT: Using XML for user output degrades user experience

### User-Facing Output (Markdown)

Progress reports and completion summaries must use Markdown:

- Headers for phase identification
- Lists for itemized findings
- Bold for emphasis on key results
- Code blocks for file paths and technical details

### Internal Agent Communication (XML)

For agent-to-agent data transfer only (never displayed to users):

```xml
<analysis>Codebase analysis results including detected technologies and architecture</analysis>
<approach>Documentation generation strategy based on analysis</approach>
<phase>Current execution phase with progress status</phase>
<completion>Summary of generated files and next recommended action</completion>
```

---

## Quick Reference

Entry Point: /moai:0-project

Mode Detection:

- No existing docs: Fresh documentation generation
- Existing docs: Update documentation flow

Agent Chain:

1. Explore subagent: Analyze codebase
2. Manager-docs subagent: Generate documentation

Output Files:

- `.moai/project/product.md`
- `.moai/project/structure.md`
- `.moai/project/tech.md`

Associated Skills:

- moai-workflow-project: Project management and configuration

---

Version: 3.1.0 (LSP Server Check Integration)
Last Updated: 2026-01-10
Architecture: Commands to Agents to Skills (Complete delegation)

Changes from v3.0.0:

- Added: PHASE 3.5 Development Environment Check
- Added: LSP server detection for 8 major languages
- Added: Installation guidance with language-specific commands
- Added: Auto-install option via expert-devops agent

Changes from v2.0.0:

- Removed: All configuration and settings functionality (moved to moai-adk init CLI)
- Removed: Tab-based question system (handled by CLI)
- Removed: Mode detection for SETTINGS, UPDATE, GLM_CONFIGURATION
- Simplified: Focus on documentation generation only
- Simplified: Two-agent chain (Explore + Manager-docs)
- Reduced: From 1400+ lines to approximately 250 lines

---

## EXECUTION DIRECTIVE

You must NOW execute the command following the phases described above.

1. Invoke Explore subagent to analyze codebase
2. Present analysis results via AskUserQuestion
3. Invoke Manager-docs subagent to generate documentation
4. Confirm completion and offer next steps

Do NOT just describe what you will do. DO IT.
