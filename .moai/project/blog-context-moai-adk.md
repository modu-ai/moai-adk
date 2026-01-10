# MoAI-ADK Blog Context Document

> **Purpose**: LLM context for writing an introductory blog post about MoAI-ADK
> **Target Audience**: Developers interested in AI-powered development tools
> **Key Message**: MoAI-ADK transforms Claude Code into a full-fledged agentic development platform

---

## Executive Summary

**MoAI-ADK (Modu-AI Agentic Development Kit)** is an open-source framework that extends Claude Code into a comprehensive agentic development platform. It provides:

- **20 specialized agents** for domain-specific tasks
- **90+ skills** covering languages, platforms, and workflows
- **Ralph Engine** for autonomous code quality assurance
- **SPEC-First TDD methodology** for structured development

**Version**: 0.42.3
**License**: Apache 2.0
**Repository**: https://github.com/moduai/moai-adk

---

## Why MoAI-ADK Matters

### The Problem with Current AI Coding Tools

1. **Single-Agent Limitations**: Most AI coding assistants operate as a single general-purpose agent, lacking specialized domain expertise
2. **Hallucination Risk**: LLMs can generate plausible but incorrect code patterns
3. **Context Loss**: Long conversations lose important context
4. **No Quality Assurance**: Lack of automated verification for generated code
5. **Fragmented Workflows**: Development, testing, and documentation are disconnected

### The MoAI-ADK Solution

MoAI-ADK addresses these challenges by transforming Claude Code into an orchestrated system of specialized agents, each with domain-specific knowledge and tools.

---

## Core Architecture

### 1. Alfred: The Strategic Orchestrator

Alfred is the central intelligence that coordinates all development tasks:

```
User Request → Alfred → Agent Selection → Specialized Agent → Quality Verification → Response
```

**Key Principles**:
- **Full Delegation**: Every implementation task is delegated to specialized agents
- **Complexity Analysis**: Matches task difficulty to appropriate agent capabilities
- **Multilingual Routing**: Automatically routes requests in Korean, Japanese, Chinese, or English to the right agent

**Three-Step Execution Model**:
1. **Understand**: Analyze request, load required skills, clarify ambiguities
2. **Plan**: Select agents, decompose tasks, get user approval
3. **Execute**: Invoke agents, monitor progress, integrate results

### 2. Agent Ecosystem (20 Specialized Agents)

**Expert Agents** (Domain Specialists):
- `expert-backend`: API design, authentication, database integration
- `expert-frontend`: React, Vue, Next.js, component architecture
- `expert-security`: OWASP, vulnerability assessment, secure coding
- `expert-debug`: Error diagnosis, troubleshooting, root cause analysis
- `expert-performance`: Profiling, optimization, benchmarking
- `expert-testing`: E2E, integration testing, QA automation
- `expert-refactoring`: AST-based transformations, large-scale changes
- `expert-devops`: CI/CD, Docker, Kubernetes, deployment

**Manager Agents** (Workflow Coordinators):
- `manager-tdd`: RED-GREEN-REFACTOR cycle implementation
- `manager-quality`: TRUST 5 validation, code review
- `manager-spec`: EARS-format requirements, acceptance criteria
- `manager-docs`: README, API documentation, technical writing
- `manager-git`: Commits, branches, PR management
- `manager-project`: Project setup, configuration
- `manager-strategy`: Architecture decisions, technology evaluation
- `manager-claude-code`: Claude Code configuration, MCP integration

**Builder Agents** (Creation Specialists):
- `builder-agent`: Custom agent creation
- `builder-command`: Slash command development
- `builder-skill`: Knowledge skill authoring
- `builder-plugin`: Plugin development for marketplace

### 3. Skill System (90+ Skills)

Skills provide domain-specific knowledge that prevents hallucination:

**Foundation Skills**:
- `moai-foundation-core`: SPEC system, TRUST 5, execution rules
- `moai-foundation-claude`: Agent orchestration, CLI reference
- `moai-foundation-quality`: Proactive analysis, best practices
- `moai-foundation-context`: Token budget optimization

**Language Skills** (16 Languages):
- Python, TypeScript, JavaScript, Go, Rust, Java, Kotlin
- Ruby, PHP, Elixir, R, C#, Swift, C++, Scala, Flutter/Dart

**Platform Skills**:
- Auth0, Clerk, Firebase Auth, Firestore
- Vercel, Neon, Railway, Convex, Supabase

**Domain Skills**:
- Frontend (React 19, Next.js 16, Vue 3.5)
- Backend (API design, microservices)
- Database (PostgreSQL, MongoDB, Redis)

**Skill Loading Features**:
- LRU cache with TTL for performance
- Dependency management (auto-load related skills)
- Effort-based filtering (1=basic to 5=comprehensive)
- Auto-detection from user prompts

### 4. /moai:alfred - One-Click Development Automation

**The flagship command that embodies MoAI-ADK's philosophy: complete development automation with a single command.**

```
/moai:alfred "User authentication with JWT tokens"
```

This single command orchestrates the entire development lifecycle:

```
┌─────────────────────────────────────────────────────────────────┐
│                    /moai:alfred Workflow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   User Input: "feature description"                              │
│         │                                                        │
│         ▼                                                        │
│   ┌─────────────────────────────────────────┐                   │
│   │  Intelligent Routing Analysis           │                   │
│   │  - Detect domain keywords               │                   │
│   │  - Count domains involved               │                   │
│   │  - Determine optimal path               │                   │
│   └─────────────────────────────────────────┘                   │
│         │                                                        │
│         ├─── Single Domain ──▶ Expert Agent Direct              │
│         │                                                        │
│         ▼                                                        │
│   ┌─────────────────────────────────────────┐                   │
│   │  Phase 1: SPEC Generation               │                   │
│   │  - EARS format requirements             │ (/moai:1-plan)    │
│   │  - User approval checkpoint             │                   │
│   └─────────────────────────────────────────┘                   │
│         │                                                        │
│         ▼                                                        │
│   ┌─────────────────────────────────────────┐                   │
│   │  Phase 2: TDD Implementation            │                   │
│   │  - RED-GREEN-REFACTOR cycle             │ (/moai:2-run)     │
│   │  - 85% test coverage target             │                   │
│   │  - Quality gate validation              │                   │
│   └─────────────────────────────────────────┘                   │
│         │                                                        │
│         ▼                                                        │
│   ┌─────────────────────────────────────────┐                   │
│   │  Phase 3: Documentation Sync            │                   │
│   │  - Update README/docs                   │ (/moai:3-sync)    │
│   │  - Git commit                           │                   │
│   │  - PR creation (optional)               │                   │
│   └─────────────────────────────────────────┘                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Key Features**:

1. **Intelligent Routing**: Analyzes request to determine optimal execution path
   - Multi-domain tasks → Full SPEC workflow (Plan → Run → Sync)
   - Single-domain tasks → Direct expert agent delegation

2. **Multi-LLM Mode Support**:
   - `opus-only`: All phases in Claude Opus (current terminal)
   - `hybrid`: Plan in Opus, Run/Sync in GLM (worktree)
   - `glm-only`: Full workflow in GLM via worktree

3. **Quality Gates at Each Phase**:
   - Phase 1 → 2: User must approve SPEC
   - Phase 2 → 3: TRUST 5 validation, 85%+ coverage
   - Phase 3 → Complete: Documentation sync verified

4. **Resumable Workflows**:
   ```
   /moai:alfred resume SPEC-AUTH-001
   ```
   - State persisted in `.moai/cache/alfred-{spec-id}.json`
   - Resume from last successful checkpoint

5. **Git Strategy Integration**:
   - `--branch`: Force feature branch creation
   - `--pr`: Force PR creation after sync
   - Respects `git-strategy.yaml` configuration

6. **Ralph Engine Integration**:
   - LSP diagnostics after each file change
   - AST-grep security scan before tests
   - Automatic fix-verify-fix cycles (max 3 iterations)

**Usage Examples**:
```bash
# Basic workflow
/moai:alfred "Shopping cart with discount codes"

# With branch and PR
/moai:alfred "OAuth2 authentication" --branch --pr

# Resume interrupted workflow
/moai:alfred resume SPEC-OAUTH-001
```

**Why This Matters**:
- **One command replaces 10+ manual steps**
- **Consistent quality through enforced gates**
- **Parallel development with worktree support**
- **Complete traceability from requirement to PR**

### 5. Ralph Engine: Autonomous Quality Assurance

Ralph Engine is the unified code quality system integrating three technologies:

```
┌─────────────────────────────────────────────────────────┐
│                    Ralph Engine                          │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │ LSP Client  │  │  AST-grep   │  │ Loop Controller │  │
│  │             │  │             │  │                 │  │
│  │ Real-time   │  │ Structural  │  │ Autonomous      │  │
│  │ Diagnostics │  │ Patterns    │  │ Feedback        │  │
│  └─────────────┘  └─────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**LSP Client** (`moai_adk/lsp/`):
- Full JSON-RPC 2.0 implementation
- Real-time type errors, syntax errors, warnings
- Multi-language server management
- 394 lines of production code

**AST-grep Integration** (`moai_adk/astgrep/`):
- Structural pattern matching (not regex)
- Security vulnerability detection
- Code smell identification
- Configurable rule sets

**Loop Controller** (`moai_adk/loop/`):
- Promise-based completion conditions
- Configurable max iterations
- State persistence and resumption
- Autonomous fix-verify-fix cycles

**Key Method**: `diagnose_file()` combines LSP diagnostics + AST-grep matches into unified `DiagnosisResult`

### 5. SPEC-First TDD Methodology

SPEC (Specification-First Development) enforces structured requirements before coding:

**EARS Format** (Easy Approach to Requirements Syntax):
- Ubiquitous: "The system shall..."
- Event-driven: "When [event], the system shall..."
- State-driven: "While [state], the system shall..."
- Unwanted: "If [condition], the system shall not..."

**Workflow Commands**:
1. `/moai:0-project` - Project configuration
2. `/moai:1-plan` - Create SPEC with requirements
3. `/moai:2-run` - TDD implementation cycle
4. `/moai:3-sync` - Documentation synchronization

**TRUST 5 Quality Principles**:
1. **T**ype Safety - Strong typing, no `any`
2. **R**eadability - Clean code, meaningful names
3. **U**nit Testing - 85% coverage target
4. **S**ecurity - OWASP compliance
5. **T**raceability - Requirement tracking

---

## Technical Highlights

### Performance Optimization

**CLI Startup Time**: 75% reduction (400ms → 100ms)
- Lazy-loaded commands and heavy libraries
- Console instance created only when needed
- Pyfiglet logo rendering deferred

```python
# Example: Lazy-loaded CLI structure
@cli.command()
def init(...):
    from moai_adk.cli.commands.init import init as _init  # Lazy import
    ctx.invoke(_init, ...)
```

### Claude Code Integration

**Template Variable Processing**:
- `{{PROJECT_DIR}}` → Absolute project path
- `{{CONVERSATION_LANGUAGE}}` → User's preferred language
- `$CLAUDE_PROJECT_DIR` → Runtime environment variable

**Headless Automation**:
```python
class ClaudeCLIIntegration:
    async def execute_prompt(self, prompt: str) -> AsyncIterator[str]:
        # JSON streaming for real-time feedback
        async for line in self._run_claude_headless(prompt):
            yield self._parse_streaming_response(line)
```

**Multilingual Support**:
- UI responses: User's conversation_language (ko, en, ja, zh)
- Internal instructions: Always English
- Code comments: Configurable

### Modular Configuration

```yaml
# .moai/config/sections/language.yaml
language:
  conversation_language: "ko"
  agent_prompt_language: en
  git_commit_messages: en
  code_comments: en
  documentation: en
```

---

## Competitive Advantages

### vs. Vanilla Claude Code

| Feature | Claude Code | MoAI-ADK |
|---------|-------------|----------|
| Agent Specialization | Single general agent | 20 specialized agents |
| Domain Knowledge | General LLM knowledge | 90+ curated skills |
| Quality Assurance | Manual review | Ralph Engine automation |
| Methodology | Ad-hoc | SPEC-First TDD |
| Multi-language | English-centric | 4-language routing |

### vs. Other AI Coding Tools

1. **Deeper Integration**: Native Claude Code extension, not a separate tool
2. **Orchestration**: Multi-agent collaboration, not single-agent limitations
3. **Anti-Hallucination**: Skill-based verified patterns
4. **Open Source**: Full customization, Apache 2.0 license
5. **Enterprise Ready**: Token optimization, context management

---

## Use Cases

### 1. Enterprise Development Teams
- Standardized workflows with SPEC-First TDD
- Consistent code quality via TRUST 5
- Multi-language team support

### 2. Full-Stack Projects
- Parallel frontend + backend development
- Integrated testing and documentation
- Unified quality gates

### 3. Security-Critical Applications
- OWASP compliance built-in
- Vulnerability scanning with AST-grep
- Security expert agent reviews

### 4. Open Source Maintainers
- Automated PR reviews
- Documentation generation
- Consistent contribution guidelines

---

## Getting Started

### Installation
```bash
# Using uv (recommended)
uv tool install moai-adk

# Initialize a project
moai-adk init my-project
```

### First Workflow (One Command)
```bash
# The simplest way: One-click automation
/moai:alfred "user authentication with JWT and refresh tokens"
```

This single command:
1. Generates SPEC with EARS-format requirements
2. Implements with TDD (85% coverage target)
3. Syncs documentation and creates PR

### Manual Workflow (Step by Step)
```bash
# 1. Generate SPEC
/moai:1-plan "user authentication system"

# 2. Implement with TDD
/moai:2-run SPEC-001

# 3. Sync documentation
/moai:3-sync SPEC-001
```

### Quick Commands
- `/moai:alfred` - **One-click automation (flagship command)**
- `/moai:fix` - Auto-fix current LSP errors
- `/moai:loop` - Start Ralph feedback loop
- `/moai:cancel-loop` - Cancel active feedback loop

---

## Architecture Summary

```
┌──────────────────────────────────────────────────────────────┐
│                         User Request                          │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                    Alfred (Orchestrator)                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │ Understand  │→ │    Plan     │→ │      Execute        │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
     ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
     │   Expert    │ │   Manager   │ │   Builder   │
     │   Agents    │ │   Agents    │ │   Agents    │
     │  (8 types)  │ │  (8 types)  │ │  (4 types)  │
     └─────────────┘ └─────────────┘ └─────────────┘
              │               │               │
              └───────────────┼───────────────┘
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                       Skill System                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ Foundation  │ │  Language   │ │  Platform   │  (90+)     │
│  └─────────────┘ └─────────────┘ └─────────────┘            │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                      Ralph Engine                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐     │
│  │     LSP     │ │  AST-grep   │ │   Loop Controller   │     │
│  └─────────────┘ └─────────────┘ └─────────────────────┘     │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                    Quality-Verified Output                    │
└──────────────────────────────────────────────────────────────┘
```

---

## Key Metrics

- **1** command to rule them all (`/moai:alfred`)
- **3** workflow phases (Plan → Run → Sync)
- **20** specialized agents
- **90+** domain skills
- **16** programming languages supported
- **4** UI languages (EN, KO, JA, ZH)
- **3** LLM modes (opus-only, hybrid, glm-only)
- **75%** CLI startup time reduction
- **85%** target test coverage
- **759** lines in `/moai:alfred` command definition
- **308** lines in Ralph Engine core
- **580** lines in Skill Loading System
- **394** lines in Claude Integration

---

## Blog Post Suggestions

### Possible Titles
1. "One Command, Complete Development: Introducing /moai:alfred"
2. "From AI Assistant to AI Team: How MoAI-ADK Transforms Claude Code"
3. "The Future of Agentic Coding: Plan → Run → Sync with MoAI-ADK"
4. "Why Single-Agent AI Coding Tools Aren't Enough"
5. "Building Quality Software with AI: The SPEC-First TDD Approach"

### Key Talking Points
1. **One-Click Revolution**: `/moai:alfred` automates the entire development lifecycle
2. **The Agent Revolution**: Moving from single AI assistants to coordinated agent teams
3. **Quality at Scale**: How Ralph Engine ensures code quality automatically
4. **Anti-Hallucination Strategy**: Skill-based knowledge prevents LLM mistakes
5. **Developer Experience**: Familiar workflows enhanced with AI capabilities
6. **Resumable Workflows**: Never lose progress with checkpoint-based execution
7. **Multi-LLM Flexibility**: Use different models for different phases
8. **Open Source Philosophy**: Customizable, extensible, community-driven

### Target Keywords
- Agentic development
- Claude Code extension
- AI coding automation
- One-click development
- SPEC-First TDD
- Multi-agent orchestration
- Code quality AI
- /moai:alfred

---

## Source Code References

### Core Modules
| Module | Path | Lines | Purpose |
|--------|------|-------|---------|
| CLI Entry | `src/moai_adk/__main__.py` | 200 | Lazy-loaded CLI |
| Claude Integration | `src/moai_adk/core/claude_integration.py` | 394 | Headless automation |
| Skill System | `src/moai_adk/core/skill_loading_system.py` | 580 | Skill management |
| Ralph Engine | `src/moai_adk/ralph/engine.py` | 308 | Quality assurance |
| LSP Client | `src/moai_adk/lsp/` | ~500 | Language diagnostics |
| Loop Controller | `src/moai_adk/loop/` | ~400 | Feedback loops |
| Foundation | `src/moai_adk/foundation/` | ~600 | EARS, Git, Languages |

### Flagship Command
| Command | Path | Lines | Purpose |
|---------|------|-------|---------|
| /moai:alfred | `.claude/commands/moai/alfred.md` | 759 | One-click automation |

### Templates
| Component | Count | Path |
|-----------|-------|------|
| Agents | 20 | `.claude/agents/moai/` |
| Skills | 90+ | `.claude/skills/` |
| Commands | 11 | `.claude/commands/moai/` |
| Output Styles | 2 | `.claude/output-styles/moai/` |

---

## Document Metadata

- **Created**: 2026-01-10
- **Based on**: Source code analysis of v0.42.3
- **Files Analyzed**: 15+ core modules, 20 agents, 90+ skills
- **Purpose**: LLM context for blog post generation

---

*This document provides comprehensive context for an LLM to write an engaging, accurate, and technically detailed blog post about MoAI-ADK.*
