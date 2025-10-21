---
id: STRUCTURE-001
version: 0.1.2
status: active
created: 2025-10-01
updated: 2025-10-22
author: @architect
priority: medium
---

# MoAI-ADK Structure Design

## HISTORY

### v0.1.2 (2025-10-22)
- **UPDATED**: Template optimization complete (v0.4.1)
- **AUTHOR**: @Alfred (@project-manager)
- **SECTIONS**: Expanded architecture with 4-layer stack, module details, integration points, and TAG traceability
- **CHANGES**: Added real MoAI-ADK architecture (Commands → Agents → Skills → Hooks)

### v0.1.1 (2025-10-17)
- **UPDATED**: Template version synced (v0.3.8)
- **AUTHOR**: @Alfred
- **SECTIONS**: Metadata standardization (single `author` field, added `priority`)

### v0.1.0 (2025-10-01)
- **INITIAL**: Authored the structure design document
- **AUTHOR**: @architect
- **SECTIONS**: Architecture, Modules, Integration, Traceability

---

## @DOC:ARCHITECTURE-001 System Architecture

### Architectural Strategy: 4-Layer Agentic Stack

MoAI-ADK follows a layered architecture where each layer has a single responsibility, enabling Progressive Disclosure of context and knowledge on demand.

```
MoAI-ADK 4-Layer Architecture (v0.4.1)
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: Commands (User ↔ Alfred Interface)                │
│ - /alfred:0-project, 1-plan, 2-run, 3-sync                 │
│ - Workflow orchestration with approval gates               │
│ - Entry points for SPEC → TDD → Sync cadence               │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: Sub-agents (Deep Reasoning & Decision Making)     │
│ - 12 specialist agents (Sonnet/Haiku)                      │
│ - spec-builder, code-builder pipeline, doc-syncer, etc.    │
│ - Task delegation, status reporting, blocker escalation    │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: Skills (Reusable Knowledge Capsules)              │
│ - 44 Claude Skills across 5 tiers                          │
│ - Foundation (TRUST/TAG/Git), Essentials, Domain, Language │
│ - Just-in-time loading via Progressive Disclosure          │
├─────────────────────────────────────────────────────────────┤
│ Layer 4: Hooks (Runtime Guardrails & JIT Context)          │
│ - SessionStart (project status card)                       │
│ - PreToolUse (destructive command blocker)                 │
│ - <100ms validation and context hints                      │
└─────────────────────────────────────────────────────────────┘
```

**Rationale**:
1. **Separation of Concerns**: Commands handle orchestration, agents handle reasoning, skills handle knowledge, hooks handle safety
2. **Progressive Disclosure**: Load only the context/knowledge needed for the current step (reduces token usage by 60%+)
3. **Agent Specialization**: Each agent is an expert in one domain (follows single responsibility principle)
4. **Fail-Safe Design**: Hooks provide pre-execution guardrails; agents can escalate to debug-helper on failure

## @DOC:MODULES-001 Module Responsibilities

### 1. Alfred Command Layer (`/alfred:*`)

- **Responsibilities**: Workflow orchestration, phase management, user interaction, approval gates
- **Inputs**: User commands (`/alfred:0-project`, `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`)
- **Processing**:
  1. Parse command and extract parameters
  2. Validate project state (config.json, existing specs, git status)
  3. Delegate to appropriate sub-agents
  4. Track progress via TodoWrite
  5. Enforce quality gates (TRUST principles)
- **Outputs**: Updated project files, git commits, status reports

| Command              | Phase | Key Capabilities                                                |
| -------------------- | ----- | --------------------------------------------------------------- |
| `/alfred:0-project`  | Init  | Project bootstrap, metadata interview, language detection       |
| `/alfred:1-plan`     | Plan  | SPEC authoring (EARS), plan board consolidation, TAG assignment |
| `/alfred:2-run`      | Run   | TDD implementation (RED→GREEN→REFACTOR), automated testing      |
| `/alfred:3-sync`     | Sync  | Living docs update, TAG chain validation, PR readiness check    |

### 2. Agent Orchestration Layer (`.claude/agents/alfred/`)

- **Responsibilities**: Task execution, domain expertise, reasoning, decision making
- **Inputs**: Delegated tasks from commands, context from hooks, knowledge from skills
- **Processing**:
  1. Analyze task requirements and constraints
  2. Load relevant skills (Progressive Disclosure)
  3. Execute specialized logic (SPEC authoring, TDD cycles, Git automation)
  4. Report status, confidence, and blockers
  5. Escalate failures to debug-helper
- **Outputs**: SPEC files, source code, tests, documentation, git operations

| Agent                      | Model  | Specialty                               |
| -------------------------- | ------ | --------------------------------------- |
| project-manager 📋         | Sonnet | Project initialization, metadata setup  |
| spec-builder 🏗️            | Sonnet | EARS-based SPEC authoring               |
| implementation-planner 📋  | Sonnet | Implementation strategy, library choice |
| tdd-implementer 🔬         | Sonnet | RED-GREEN-REFACTOR execution            |
| doc-syncer 📖              | Haiku  | Living documentation sync               |
| tag-agent 🏷️               | Haiku  | TAG inventory, orphan detection         |
| git-manager 🚀             | Haiku  | GitFlow automation, PR management       |
| debug-helper 🔍            | Sonnet | Failure diagnosis, fix-forward guidance |
| trust-checker ✅           | Haiku  | TRUST 5 principle enforcement           |
| quality-gate 🛡️            | Haiku  | Coverage delta, release validation      |
| cc-manager 🛠️              | Sonnet | Claude Code session tuning              |

### 3. Skills Repository Layer (`.claude/skills/`)

- **Responsibilities**: Reusable knowledge encapsulation, best practices, templates
- **Inputs**: Skill load requests from agents (e.g., `Skill("moai-foundation-trust")`)
- **Processing**:
  1. Progressive Disclosure: Load metadata first, full content on demand
  2. Provide templates (SPEC, test, commit message formats)
  3. Offer checklists and decision trees
- **Outputs**: Contextual knowledge for agents (EARS syntax, TRUST principles, language-specific TDD patterns)

| Skill Tier         | Count | Examples                                                           |
| ------------------ | ----- | ------------------------------------------------------------------ |
| Foundation         | 6     | trust, tags, specs, ears, git, langs                               |
| Essentials         | 4     | debug, perf, refactor, review                                      |
| Domain             | 10    | backend, frontend, web-api, mobile-app, security, devops, etc.     |
| Language           | 23    | Python, TypeScript, Go, Rust, Java, Kotlin, Swift, etc.            |
| Claude Code Ops    | 1     | claude-code (session settings, output styles, Skill lifecycle)     |

### 4. Hook System Layer (`.claude/hooks/alfred/`)

- **Responsibilities**: Runtime safety, pre-execution validation, just-in-time context hints
- **Inputs**: Session events (SessionStart, PreToolUse), tool invocations (Bash, Edit, Write)
- **Processing**:
  1. SessionStart: Load project config, display status card
  2. PreToolUse: Block destructive commands (rm -rf, git reset --hard without confirmation)
  3. Context injection: Surface relevant SPEC/TAG pointers
- **Outputs**: Guardrail warnings, context hints, execution blocks (when unsafe)

## @DOC:INTEGRATION-001 External Integrations

### Claude API Integration (Anthropic)

- **Purpose**: Core reasoning engine for Alfred and all sub-agents
- **Authentication**: API key via environment variable (`ANTHROPIC_API_KEY`)
- **Models Used**:
  - **Claude 4.5 Sonnet**: Planning, implementation, troubleshooting (Alfred, spec-builder, code-builder pipeline, debug-helper, cc-manager, project-manager)
  - **Claude 4.5 Haiku**: Documentation, TAG management, Git automation, rule-based checks (doc-syncer, tag-agent, git-manager, trust-checker, quality-gate)
- **Data Exchange**: JSON via Messages API (streaming enabled for real-time feedback)
- **Failure Handling**:
  - Retry with exponential backoff (3 attempts)
  - Fallback to cached context when API unavailable
  - Graceful degradation: Manual mode prompts if API fails
- **Risk Level**: HIGH (critical dependency)
  - **Mitigation**: Local caching of frequently used Skill content, offline mode for read-only operations

### Git/GitHub Integration

- **Purpose**: Version control, GitFlow automation, PR management, CI/CD triggers
- **Authentication**: SSH key or GitHub token (GITHUB_TOKEN)
- **Operations**:
  - Branch creation (feature/SPEC-XXX)
  - Commit generation (RED→GREEN→REFACTOR)
  - Draft PR creation and Ready PR promotion
  - Tag-based releases
- **Failure Handling**:
  - Detect merge conflicts and surface resolution guidance
  - Block force-push without explicit confirmation
  - Validate remote connectivity before push operations
- **Performance Requirements**: <2s for local git operations, <5s for remote push/pull
- **Risk Level**: MEDIUM (degraded workflow if unavailable)
  - **Mitigation**: Local-first architecture (all operations work offline, sync when online)

### ripgrep (rg) Integration

- **Purpose**: Fast code scanning for @TAG traceability, SPEC validation, duplicate detection
- **Dependency Level**: CRITICAL (TAG system relies on code-first scanning)
- **Operations**:
  - `rg '@(SPEC|TEST|CODE|DOC):' -n` for TAG chain validation
  - `rg '@SPEC:AUTH' -n` for duplicate detection
  - Pattern matching for EARS syntax validation
- **Fallback**: grep (slower, universal availability)
- **Performance Requirements**: <1s for full codebase scan (up to 100K LOC)
- **Risk Level**: LOW (fallback available)

### Language-Specific Toolchains

- **Purpose**: Linting, testing, type checking, building per language
- **Examples**:
  - Python: pytest, ruff, mypy, uv
  - TypeScript: vitest, biome, tsc, pnpm
  - Go: go test, golangci-lint, go build
  - Rust: cargo test, clippy, cargo build
- **Dependency Level**: HIGH (TRUST principles require these tools)
- **Failure Handling**: Skip optional tools (e.g., linter) if unavailable, block required tools (e.g., test runner)
- **Risk Level**: MEDIUM (quality gates degraded without proper tools)

## @DOC:TRACEABILITY-001 Traceability Strategy

### Applying the TAG Framework

**Full TDD Alignment**: SPEC → Tests → Implementation → Documentation
- `@SPEC:ID` (`.moai/specs/`) → `@TEST:ID` (`tests/`) → `@CODE:ID` (`src/`) → `@DOC:ID` (`docs/`)

**Implementation Detail Levels**: Annotation within `@CODE:ID`
- `@CODE:ID:API` – REST APIs, GraphQL endpoints
- `@CODE:ID:UI` – Components, views, screens
- `@CODE:ID:DATA` – Data models, schemas, types
- `@CODE:ID:DOMAIN` – Business logic, domain rules
- `@CODE:ID:INFRA` – Infrastructure, databases, integrations

### Managing TAG Traceability (Code-Scan Approach)

- **Verification**: Run `/alfred:3-sync`, which scans with `rg '@(SPEC|TEST|CODE|DOC):' -n`
- **Coverage**: Full project source (`.moai/specs/`, `tests/`, `src/`, `docs/`)
- **Cadence**: Validate whenever the code changes
- **Code-First Principle**: TAG truth lives in the source itself

## Legacy Context

### Current System Snapshot (MoAI-ADK v0.4.1)

**Production-ready 4-layer architecture with 44 Skills and 12 agents**

```
MoAI-ADK/
├── .claude/                    # Claude Code configuration layer
│   ├── agents/alfred/          # 12 sub-agent definitions (Sonnet/Haiku)
│   ├── commands/alfred/        # 4 workflow commands (0-project, 1-plan, 2-run, 3-sync)
│   ├── skills/                 # 44 Claude Skills (Foundation, Essentials, Domain, Language, Ops)
│   ├── hooks/alfred/           # Runtime guardrails (SessionStart, PreToolUse)
│   └── settings.json           # Session configuration
├── .moai/                      # Project metadata and documentation
│   ├── config.json             # Project settings (mode, language, optimized flag)
│   ├── project/                # Product/structure/tech.md (this file)
│   ├── specs/                  # SPEC repository (EARS-based requirements)
│   ├── memory/                 # Knowledge corpus (TRUST, GitFlow, SPEC metadata policies)
│   └── reports/                # Living documentation (sync reports, TAG chain validation)
├── src/moai_adk/               # Python CLI implementation
│   ├── cli/                    # moai-adk init/update commands
│   ├── core/                   # Project detection, template management
│   └── templates/              # Bootstrap templates for new projects
└── tests/                      # pytest test suite (85%+ coverage)
```

### Evolution History

1. **v0.1.0–v0.2.x**: Single-agent prototype (Alfred only)
2. **v0.3.0–v0.3.8**: 9-agent ecosystem, initial Skill system
3. **v0.4.0–v0.4.1**: 12-agent roster (code-builder pipeline split), 44 Skills, Progressive Disclosure optimization

### Migration Considerations for Future Adopters

1. **Multi-project workspace support** – Priority: MEDIUM
   - Current: One project per directory
   - Planned: Monorepo detection, multi-root workspaces
2. **Agent performance telemetry** – Priority: MEDIUM
   - Current: Manual status reporting
   - Planned: Automated task duration, token usage, error rate tracking
3. **Cross-repository SPEC references** – Priority: LOW
   - Current: Single-repo TAG chains
   - Planned: Inter-repo @SPEC linking for microservices

## TODO:STRUCTURE-001 Structural Improvements

1. **Agent communication protocol** – Standardize message format for agent-to-agent handoffs (Priority: HIGH)
2. **Skill dependency graph** – Auto-detect required Skills based on project stack (Priority: MEDIUM)
3. **Hook extensibility API** – Allow custom hooks without modifying core alfred_hooks.py (Priority: LOW)

## EARS for Architectural Requirements

### Applying EARS to Architecture

Use EARS patterns to write clear architectural requirements:

#### Architectural EARS Example
```markdown
### Ubiquitous Requirements (Baseline Architecture)
- The system shall adopt a layered architecture.
- The system shall maintain loose coupling across modules.

### Event-driven Requirements
- WHEN an external API call fails, the system shall execute fallback logic.
- WHEN a data change event occurs, the system shall notify dependent modules.

### State-driven Requirements
- WHILE the system operates in scale-out mode, it shall load new modules dynamically.
- WHILE in development mode, the system shall provide verbose debug information.

### Optional Features
- WHERE the deployment runs in the cloud, the system may use distributed caching.
- WHERE high performance is required, the system may apply in-memory caching.

### Constraints
- IF the security level is elevated, the system shall encrypt all inter-module communication.
- Each module shall keep cyclomatic complexity under 15.
```

---

_This structure informs the TDD implementation when `/alfred:2-run` runs._
