---
id: STRUCTURE-001
version: 0.2.0
status: active
created: 2025-10-01
updated: 2025-10-27
author: @Alfred
priority: high
---

# MoAI-ADK Structure Design

## HISTORY

### v0.2.0 (2025-10-27)
- **UPDATED**: Auto-generated comprehensive structure design based on actual codebase analysis
- **AUTHOR**: @Alfred
- **SECTIONS**: Architecture (verified from 40+ modules), Modules (detailed responsibility matrix), Integration (GitPython, Claude Code, PyPI), Traceability (@TAG strategy implementation)
- **ANALYSIS**: Analyzed `/src/moai_adk/` directory structure; scanned 40+ Python modules; identified 4 core layers

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

### Architectural Strategy: Layered Modular Monolith with Event-Driven Git Integration

**MoAI-ADK** uses a **four-layer architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│  CLI LAYER (User Interface)                                     │
│  ├─ Click commands (init, doctor, update, status, backup)      │
│  ├─ Rich terminal UI (banners, progress indicators)             │
│  ├─ questionary TUI (interactive menus, language selection)    │
│  └─ Prompt management (init_prompts.py)                         │
└────────────────────────┬────────────────────────────────────────┘
                         │ Command routing
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  CORE LAYER (Business Logic)                                    │
│  ├─ git/     → GitFlow management, checkpoints, commits        │
│  ├─ project/ → Initialization, detection, validation           │
│  ├─ quality/ → TRUST 5 validation, coverage checking           │
│  └─ template/→ Template merging, backup, processor             │
└────────────────────────┬────────────────────────────────────────┘
                         │ Service orchestration
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  TEMPLATES LAYER (Configuration & Scaffolding)                  │
│  ├─ .claude/     → Agents, commands, skills, hooks             │
│  ├─ .moai/       → Config, project metadata, specs, reports    │
│  └─ .github/     → CI/CD workflows (GitHub Actions)            │
└────────────────────────┬────────────────────────────────────────┘
                         │ Template instantiation
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  UTILS LAYER (Cross-cutting Concerns)                           │
│  ├─ logger.py  → Structured logging                            │
│  ├─ banner.py  → Terminal formatting & ASCII art               │
│  └─ decorators → Function wrapping (error handling, timing)    │
└─────────────────────────────────────────────────────────────────┘
```

### Architectural Rationale

**Why Layered Modular Monolith?**
1. **Simplicity**: Single codebase, easy to reason about for small-to-medium teams
2. **Performance**: No network overhead; all operations in-process
3. **Testability**: Clear module boundaries; can test each layer independently
4. **Scalability**: Can be split into microservices later if needed (e.g., spec-builder as separate service)

**Why Event-Driven Git?**
1. **Automatic checkpoints**: Detects critical operations (delete, merge, refactor) and creates safety checkpoints
2. **TDD commits**: Enforces RED → GREEN → REFACTOR pattern in git history
3. **Context preservation**: Each commit message includes @TAG references for traceability

**Why Progressive Disclosure of Knowledge?**
1. **Context efficiency**: Foundation skills load at session start; Domain/Language skills load just-in-time
2. **User experience**: Faster initial session boot; customized skill packs per language
3. **Maintainability**: 55+ skills organized in 6 tiers (Foundation, Essentials, Alfred, Domain, Language, Ops)

---

## @DOC:MODULES-001 Module Responsibilities

### 1. **CLI Module** (`moai_adk/cli/`)

**Role**: User-facing command interface with rich terminal UI

| Component          | Responsibility                                          | Key Code Files              |
| ------------------ | ------------------------------------------------------- | --------------------------- |
| **main.py**        | Entry point; routes commands (init, doctor, etc.)      | cli/main.py                 |
| **commands/**      | Implementations: init, doctor, update, status, backup  | cli/commands/*.py           |
| **prompts/**       | Interactive questionary TUI and interview flows        | cli/prompts/init_prompts.py |
| **Output Format**  | Rich formatting, banners, progress indicators          | utils/banner.py             |

**Inputs**: User commands + command-line arguments
**Processing**:
1. Parse command + arguments
2. Call appropriate Core module
3. Format output with Rich
**Outputs**:
- Terminal output (status, errors, reports)
- File system changes (delegated to Core)

**Quality Attributes**:
- ✅ Responsive: <1s command execution for status queries
- ✅ User-friendly: Clear error messages, progress indicators
- ✅ Accessible: Works with both interactive and non-interactive terminals

---

### 2. **Git Module** (`moai_adk/core/git/`)

**Role**: GitFlow management, automatic checkpoints, TDD commit enforcement

| Component           | Responsibility                                          | Key Code Files          |
| ------------------- | ------------------------------------------------------- | ----------------------- |
| **manager.py**      | Orchestrates git operations (branch, commit, rebase)  | core/git/manager.py     |
| **branch_manager.py**| Feature branch creation, checkout, tracking            | core/git/branch.py      |
| **commit.py**       | Commit message formatting with TDD pattern             | core/git/commit.py      |
| **checkpoint.py**   | Auto-checkpoint before critical operations             | core/git/checkpoint.py  |
| **event_detector.py**| Detects critical events (delete, merge, refactor)     | core/git/event_detector.py |

**Inputs**:
- Git repository state
- TDD phase (RED/GREEN/REFACTOR)
- Critical event triggers
**Processing**:
1. Detect git state (branch, uncommitted changes)
2. Create checkpoint if critical operation detected
3. Format commit message with @TAG and TDD phase
4. Execute git operation (commit, push, PR creation)
**Outputs**:
- Git commits with TDD phase labels (🔴 test, 🟢 feat, ♻️ refactor)
- Feature branches (feature/SPEC-*)
- Draft PRs (in team mode)

**Quality Attributes**:
- ✅ Safety-first: Never loses data; checkpoints created automatically
- ✅ TDD-enforced: Commit structure ensures RED→GREEN→REFACTOR cadence
- ✅ Traceable: All commits tagged with @SPEC/@TEST/@CODE IDs

---

### 3. **Project Module** (`moai_adk/core/project/`)

**Role**: Project initialization, language detection, validation, metadata management

| Component          | Responsibility                                          | Key Code Files           |
| ------------------ | ------------------------------------------------------- | ------------------------ |
| **initializer.py** | Bootstrap `.moai/`, `.claude/`, `.github/` directories | core/project/initializer.py |
| **detector.py**    | Auto-detect language (Python, TypeScript, Go, etc.)   | core/project/detector.py |
| **validator.py**   | Validate `.moai/config.json`, SPEC YAML frontmatter   | core/project/validator.py |
| **checker.py**     | `moai doctor` diagnostics (versions, dependencies)    | core/project/checker.py  |
| **backup_utils.py**| Backup/restore `.moai/project/` files                 | core/project/backup_utils.py |
| **phase_executor.py**| Execute `/alfred:*` command phases                   | core/project/phase_executor.py |

**Inputs**:
- Project directory structure
- pyproject.toml, package.json, go.mod, etc.
- Existing `.moai/config.json`
**Processing**:
1. Scan directory for language hints
2. Auto-detect primary language and framework
3. Create/validate config.json
4. Bootstrap template structure
5. Execute multi-phase initialization
**Outputs**:
- `.moai/config.json` (project metadata)
- `.moai/project/*.md` (product, structure, tech documents)
- `.moai/memory/` (Alfred's knowledge base)
- `.moai/specs/` (SPEC directory structure)
- Git commits with metadata

**Quality Attributes**:
- ✅ Non-destructive: Never overwrites existing user code
- ✅ Intelligent: Auto-detects language and recommends tech stack
- ✅ Resumable: Can run multiple times safely

---

### 4. **Quality Module** (`moai_adk/core/quality/`)

**Role**: TRUST 5 principles enforcement, test coverage validation, code quality gates

| Component              | Responsibility                                          | Key Code Files          |
| ---------------------- | ------------------------------------------------------- | ----------------------- |
| **trust_checker.py**   | Validates TRUST 5 (Test, Readable, Unified, Secured, Trackable) | core/quality/trust_checker.py |
| **validators/*.py**    | Specific validators (coverage, linter, type-checker)  | core/quality/validators/ |
| **base_validator.py**  | Base class for all validators                          | core/quality/validators/base_validator.py |

**Inputs**:
- Python source code
- Test coverage data
- Linter/formatter output
- Security audit results
**Processing**:
1. Run pytest + coverage analysis
2. Execute ruff linter
3. Run mypy type checking
4. Scan for security issues (bandit, pip-audit)
5. Aggregate results; fail if any threshold violated
**Outputs**:
- TRUST report (pass/fail/warning)
- Coverage percentage
- Linting violations
- Type errors
- Security vulnerabilities

**Quality Attributes**:
- ✅ Strict: Fails build if coverage < 85% (non-negotiable)
- ✅ Multi-faceted: Checks test, readability, type safety, security
- ✅ Actionable: Clear error messages with fix suggestions

---

### 5. **Template Module** (`moai_adk/core/template/`)

**Role**: Template scaffolding, merging, configuration management

| Component        | Responsibility                                          | Key Code Files           |
| ---------------- | ------------------------------------------------------- | ------------------------ |
| **processor.py** | Template variable substitution ({{VARIABLES}})         | core/template/processor.py |
| **merger.py**    | Intelligent merge of templates with existing files     | core/template/merger.py  |
| **backup.py**    | Backup old templates before overwriting                | core/template/backup.py  |
| **config.py**    | Configuration file management                          | core/template/config.py  |
| **languages.py** | Language-specific template paths and configurations   | core/template/languages.py |

**Inputs**:
- Template source files (in src/moai_adk/templates/)
- User configuration (language, mode, locale)
- Existing project files
**Processing**:
1. Backup existing `.moai/` and `.claude/` directories
2. Copy template structure
3. Substitute variables ({{LANGUAGE}}, {{VERSION}}, etc.)
4. Merge user customizations with new templates
5. Update `.moai/config.json` with optimized flag
**Outputs**:
- Scaffolded `.moai/` directory
- Scaffolded `.claude/` directory
- Updated `.moai/config.json`
- `.moai-backups/` (timestamped backups)

**Quality Attributes**:
- ✅ Non-destructive: Never loses user customizations
- ✅ Idempotent: Can run multiple times safely
- ✅ Intelligent merge: Uses content analysis to preserve intent

---

### 6. **Diagnostics Module** (`moai_adk/core/diagnostics/`)

**Role**: System health checks, slash command validation, project status reporting

| Component              | Responsibility                                          | Key Code Files           |
| ---------------------- | ------------------------------------------------------- | ------------------------ |
| **slash_commands.py**  | Validate `/alfred:*` command availability             | core/diagnostics/slash_commands.py |

**Inputs**:
- Project configuration
- Installed Python packages
- `.claude/commands/` directory
**Processing**:
1. Check Python version compatibility
2. Verify uv, pytest, ruff installations
3. Scan `.claude/` for available commands
4. Test Claude Code integration
**Outputs**:
- Health report (checklist of green/red statuses)
- Command availability report
- Dependency status

---

## @DOC:INTEGRATION-001 External Integrations

### 1. **Git Integration** (GitPython)

- **Authentication**: Git credentials from system keychain/SSH
- **Data Exchange**: Repository metadata, commit messages, branch info
- **Failure Handling**: Graceful fallback; warn if git not initialized
- **Risk Level**: Low (read-heavy; writes only for checkpoints/commits)

### 2. **Claude Code Integration**

- **Purpose**: Activation of Alfred agents and Claude Skills
- **Dependency Level**: Critical (moai-adk requires Claude Code for `/alfred:*` commands)
- **Performance Requirements**: Command routing latency <100ms
- **Integration Points**:
  - `.claude/agents/alfred/` (12 sub-agents)
  - `.claude/commands/alfred/` (4 commands: 0-project, 1-plan, 2-run, 3-sync)
  - `.claude/hooks/alfred/` (5 event-driven hooks)
  - `.claude/skills/` (55+ Claude Skills)

### 3. **PyPI Integration** (packaging)

- **Purpose**: Distribution of moai-adk as installable package
- **Dependency Level**: Optional (nice-to-have; not required for local usage)
- **Performance Requirements**: Upload latency (one-time per release)
- **Integration Points**:
  - `pyproject.toml` (package metadata)
  - GitHub Actions workflow (automated release)

### 4. **GitHub Integration** (GitHub Actions)

- **Purpose**: Automated CI/CD pipeline (test, coverage, release)
- **Dependency Level**: High (development workflow depends on it)
- **Failure Handling**: Graceful degradation (manual testing still possible)
- **Integration Points**:
  - `.github/workflows/moai-gitflow.yml` (automated tests + coverage reporting)

### 5. **File System Operations**

- **Purpose**: Read/write `.moai/`, `.claude/`, templates, backups
- **Dependency Level**: Critical
- **Failure Handling**: Atomic operations; rollback on failure
- **Risk Level**: Medium (writes to user filesystem; mitigated by backups)

---

## @DOC:TRACEABILITY-001 Traceability Strategy

### Applying the @TAG Framework

MoAI-ADK implements end-to-end traceability using the **@TAG system**. Every requirement, test, implementation, and documentation piece gets a unique identifier.

#### Full TDD Alignment: SPEC → TEST → CODE → DOC

```
@SPEC:FEATURE-001 (.moai/specs/SPEC-FEATURE-001/spec.md)
    ↓ [requirement drives test]
@TEST:FEATURE-001 (tests/test_feature.py)
    ↓ [test drives implementation]
@CODE:FEATURE-001 (src/moai_adk/*/feature.py)
    ↓ [code is documented]
@DOC:FEATURE-001 (docs/FEATURE.md)
```

#### Implementation Detail Levels: Annotation within @CODE

When a SPEC involves multiple components, use sub-annotations:

```python
# @CODE:FEATURE-001:CLI
# Handles: `/moai-adk feature-command`
def feature_command():
    pass

# @CODE:FEATURE-001:VALIDATOR
# Validates: YAML schema, config format
def validate_feature():
    pass

# @CODE:FEATURE-001:MERGER
# Merges: Old + new configuration
def merge_configs():
    pass
```

**Standard sub-types**:
- `:CLI` – Command-line interface, argument parsing
- `:API` – REST endpoints, function interfaces
- `:VALIDATOR` – Input validation, schema checking
- `:MERGER` – Merging, template processing
- `:BACKUP` – Backup, restore operations
- `:ERROR` – Error handling, exception types

#### Managing TAG Traceability (Code-Scan Approach)

**Verification Workflow**:
1. Run `/alfred:3-sync` (explicitly syncs and validates)
2. Scans entire project with `rg '@(SPEC|TEST|CODE|DOC):' -n`
3. Validates TAG chains (no orphan TAGs)
4. Generates TAG inventory report

**Coverage**:
- SPEC documents: `.moai/specs/SPEC-*/`
- Test files: `tests/**/*.py` (Python), `tests/**/*.ts` (TypeScript), etc.
- Implementation: `src/moai_adk/**/*.py`
- Documentation: `docs/` + README files

**Cadence**:
- **Per commit**: Git hook validates @TAG presence in modified files
- **Per PR**: GitHub Actions runs tag validation
- **Per release**: Comprehensive audit of tag chains

**Code-First Principle**:
- TAGs live in source code (`.py` files, `.ts` files, `.md` files)
- SPEC documents reference code TAGs
- Single source of truth: The code itself

---

## Legacy Context

### Current System Snapshot

MoAI-ADK v0.5.6 is a **mature, production-ready framework** with:

```
moai-adk/
├── src/moai_adk/              # Core implementation (40+ Python modules)
│   ├── cli/                   # User interface layer
│   ├── core/
│   │   ├── git/              # Git integration (5 modules)
│   │   ├── project/          # Project initialization (6 modules)
│   │   ├── quality/          # TRUST validation (4 modules)
│   │   └── template/         # Template scaffolding (5 modules)
│   ├── utils/                # Cross-cutting concerns (2 modules)
│   └── templates/            # Bootstrap templates (100+ files)
│
├── tests/                     # Test suite (87.84% coverage)
├── .moai/                    # Project metadata
│   ├── project/              # product.md, structure.md, tech.md
│   ├── specs/                # 28 SPEC documents
│   ├── memory/               # Alfred's knowledge base (14 files)
│   └── reports/              # Analysis and phase reports
│
├── .claude/                  # Claude Code integration
│   ├── agents/               # 12 sub-agents
│   ├── commands/             # 4 workflow commands
│   ├── skills/               # 55+ Claude Skills
│   ├── hooks/                # 5 event-driven hooks
│   └── output-styles/        # 3 response styles
│
├── .github/                  # GitHub integration
│   └── workflows/            # CI/CD pipeline
│
└── docs/                     # Generated documentation
```

### Migration Considerations

1. **Language Support Expansion** (in-progress)
   - Currently: Python 3.13+ core
   - Future: Add native support for JavaScript/TypeScript, Go, Rust

2. **IDE Integration** (planned)
   - Current: CLI-only + Claude Code integration
   - Future: VS Code extension for Alfred commands

3. **Team Collaboration** (planned)
   - Current: Single-developer workflow
   - Future: Multi-developer SPEC conflict resolution, PR automation

4. **Skills Marketplace** (exploration phase)
   - Current: 55+ built-in Skills
   - Future: Community-contributed Skills marketplace

---

## TODO:STRUCTURE-001 Structural Improvements

### Immediate (Next sprint)

1. **API Stability** – Lock down Claude Code integration points (agents, hooks, skills schema)
2. **Error Handling** – Standardize error codes across all modules
3. **Logging** – Implement structured JSON logging for better observability

### Short-term (Next quarter)

4. **Module Documentation** – Auto-generate API docs for each core module
5. **Dependency Injection** – Reduce coupling between CLI and Core layers
6. **Cache Layer** – Add in-memory cache for template processing (optimize multi-run scenarios)

### Long-term (Next 6 months)

7. **Microservices Split** – Extract spec-builder into standalone service
8. **Plugin System** – Allow third-party integrations (custom validators, custom template processors)
9. **Observability** – Add metrics collection (hook execution time, command latency, etc.)

---

## EARS for Architectural Requirements

### Applying EARS to Architecture

Use EARS patterns when designing new components or refactoring existing ones:

#### Architectural EARS Example

```markdown
### Ubiquitous Requirements (Baseline Architecture)
- The system SHALL maintain a four-layer architecture (CLI → Core → Templates → Utils).
- The system SHALL ensure no circular dependencies between modules.
- The system SHALL expose clear module boundaries via Python `__init__.py` files.

### Event-driven Requirements
- WHEN a critical git operation is detected (delete, merge), the system SHALL automatically create a checkpoint.
- WHEN `/alfred:1-plan` is invoked, the system SHALL trigger spec-builder agent via Claude Code.
- WHEN a SPEC document is written, the system SHALL validate YAML frontmatter compliance.

### State-driven Requirements
- WHILE the git repository is in a feature branch, the system SHALL enforce commit message TDD pattern (RED/GREEN/REFACTOR).
- WHILE `.moai/config.json` is missing, the system SHALL guide users through initialization prompts.

### Optional Features
- WHERE environment is development, the system MAY enable verbose logging.
- WHERE GitHub Actions is configured, the system MAY auto-deploy releases to PyPI.

### Constraints
- IF a module exceeds 500 lines of code, the system SHALL refactor it into sub-modules.
- Test execution time SHALL NOT exceed 30 seconds (on standard CI runners).
- CLI command latency SHALL NOT exceed 5 seconds (for interactive commands).
```

---

_This structure document guides the design and refactoring of MoAI-ADK when `/alfred:2-run` executes implementation phases. Update when major architectural decisions are made._
