# MoAI-ADK Project Structure

> **Last Updated**: 2026-01-13
> **Version**: 1.1.1

---

## Repository Overview

```
MoAI-ADK/
├── src/moai_adk/             # Python package source
├── tests/                    # Test suite
├── docs/                     # Documentation site (Nextra)
├── assets/                   # Images and resources
├── .claude/                  # Claude Code configuration (synced from templates)
├── .moai/                    # MoAI-ADK configuration (synced from templates)
├── .github/                  # GitHub workflows and scripts
├── CLAUDE.md                 # Alfred execution directives
├── CLAUDE.local.md           # Local development guide (git-ignored)
├── pyproject.toml            # Package configuration (Single Source of Truth for version)
└── README.md                 # Project documentation
```

---

## Package Structure (src/moai_adk/)

```
src/moai_adk/
├── __init__.py               # Package initialization
├── __main__.py               # CLI entry point
├── version.py                # Version management (reads pyproject.toml)
│
├── astgrep/                  # AST-grep integration
│   ├── checker.py            # Syntax checker
│   ├── patterns.py           # Pattern definitions
│   └── scanner.py            # Security scanner
│
├── cli/                      # CLI commands
│   ├── init.py               # Project initialization
│   ├── worktree/             # Git worktree management
│   └── commands/             # CLI subcommands
│
├── core/                     # Core modules (45 files)
│   ├── config.py             # Configuration management
│   ├── spec_manager.py       # SPEC document handling
│   ├── tdd_engine.py         # TDD workflow
│   └── ...                   # Additional core modules
│
├── foundation/               # Foundation components
│   ├── __init__.py           # Package initialization
│   ├── claude.py             # Claude Code integration
│   ├── core.py               # Core principles
│   ├── git/                  # Git operations module (directory structure since v1.1.1)
│   ├── quality.py            # Quality framework
│   └── testing.py            # TDD framework utilities
│
├── loop/                     # Feedback loop system
│   ├── controller.py         # Loop controller
│   ├── processor.py          # Error processor
│   └── hooks.py              # Hook integration
│
├── lsp/                      # LSP (Language Server Protocol) integration
│   ├── client.py             # LSP client
│   ├── diagnostics.py        # Diagnostic handler
│   └── provider.py           # LSP provider
│
├── project/                  # Project management
│   ├── manager.py            # Project manager
│   ├── template.py           # Template handling
│   └── config.py             # Project configuration
│
├── ralph/                    # Ralph Engine (feedback automation)
│   ├── engine.py             # Main engine
│   ├── analyzer.py           # Error analyzer
│   └── fixer.py              # Auto-fixer
│
├── statusline/               # Claude Code statusline
│   ├── display.py            # Status display
│   ├── formatter.py          # Output formatter
│   └── config.py             # Statusline configuration
│
├── templates/                # Distribution templates
│   ├── .claude/              # Claude Code templates
│   ├── .moai/                # MoAI configuration templates
│   └── CLAUDE.md             # Alfred directives template
│
├── utils/                    # Utility modules
│   ├── file_ops.py           # File operations
│   ├── git_ops.py            # Git operations
│   └── yaml_ops.py           # YAML handling
│
├── web/                      # Web UI backend (FastAPI)
│   ├── api/                  # API routes
│   ├── services/             # Business logic
│   └── models/               # Data models
│
└── web-ui/                   # Web UI frontend (Next.js/React)
    ├── src/                  # React components
    ├── public/               # Static assets
    └── package.json          # Node dependencies
```

---

## Configuration Structure (.claude/)

```
.claude/
├── settings.json             # Claude Code settings
├── settings.local.json       # Local settings (git-ignored)
│
├── agents/                   # Sub-agent definitions
│   └── moai/                 # MoAI agents (20 agents)
│       ├── expert-*.md       # Expert agents (8)
│       ├── manager-*.md      # Manager agents (8)
│       └── builder-*.md      # Builder agents (4)
│
├── commands/                 # Slash commands
│   └── moai/                 # MoAI commands
│       ├── 0-project.md      # Project initialization
│       ├── 1-plan.md         # SPEC planning
│       ├── 2-run.md          # TDD execution
│       ├── 3-sync.md         # Documentation sync
│       └── ...               # Additional commands
│
├── hooks/                    # Automation hooks
│   └── moai/                 # MoAI hooks
│       ├── session_start_*.py
│       └── lib/              # Hook libraries
│
├── output-styles/            # Output formatting
│   └── moai/
│       ├── r2d2.md           # R2-D2 style
│       └── yoda.md           # Yoda style
│
└── skills/                   # Domain knowledge
    └── moai/                 # MoAI skills (48 skills)
        ├── moai-foundation-*.md
        ├── moai-domain-*.md
        ├── moai-lang-*.md
        ├── moai-platform-*.md
        ├── moai-workflow-*.md
        └── moai-library-*.md
```

---

## MoAI Configuration Structure (.moai/)

```
.moai/
├── config/                   # Configuration files
│   ├── config.yaml           # Main configuration
│   ├── statusline-config.yaml
│   ├── multilingual-triggers.yaml
│   ├── sections/             # Modular sections
│   │   ├── user.yaml
│   │   ├── language.yaml
│   │   ├── project.yaml
│   │   ├── git-strategy.yaml
│   │   ├── quality.yaml
│   │   ├── system.yaml
│   │   └── ralph.yaml
│   └── questions/            # Configuration UI
│       ├── _schema.yaml
│       ├── tab0-init.yaml
│       ├── tab1-user.yaml
│       ├── tab2-project.yaml
│       ├── tab3-git.yaml
│       ├── tab4-quality.yaml
│       └── tab5-system.yaml
│
├── specs/                    # SPEC documents
│   └── SPEC-*.md             # Individual specifications
│
├── project/                  # Project documentation
│   ├── product.md            # Product description
│   ├── structure.md          # Project structure (this file)
│   └── tech.md               # Technology stack
│
├── memory/                   # Session memory (runtime)
├── cache/                    # Cache data (runtime)
├── logs/                     # Log files (runtime)
├── error_logs/               # Error logs (runtime)
├── rollbacks/                # Rollback data (runtime)
├── analytics/                # Usage analytics (runtime)
└── web/                      # Web UI data (runtime)
```

---

## Test Structure (tests/)

```
tests/
├── unit/                     # Unit tests
│   ├── test_core/            # Core module tests
│   ├── test_cli/             # CLI command tests
│   ├── test_utils/           # Utility tests
│   ├── test_statusline/      # Statusline tests
│   └── test_hooks/           # Hook tests
│
├── astgrep/                  # AST-grep integration tests (v1.1.1)
│   ├── test_analyzer_edge_cases.py      # 52 tests
│   ├── test_analyzer_full_coverage.py   # Coverage tests
│   ├── test_analyzer_exception_coverage.py # Exception handling tests
│   ├── test_models_edge_cases.py         # 32 tests
│   ├── test_rules_advanced.py            # 25 tests
│   └── test_rules_*.py                   # Additional rule tests
│
├── cli/                      # CLI tests (v1.1.1)
│   ├── commands/            # CLI command tests
│   │   ├── test_analyze_enhanced.py
│   │   ├── test_switch.py
│   │   ├── test_rank_comprehensive.py
│   │   ├── test_update_complete_coverage.py
│   │   └── test_*.py
│   ├── prompts/             # Prompt tests
│   │   ├── test_init_prompts_enhanced.py
│   │   └── test_*.py
│   ├── ui/                   # UI tests
│   │   ├── test_progress.py
│   │   └── test_*.py
│   ├── worktree/             # Worktree tests (v1.1.1)
│   │   ├── test_cli.py
│   │   ├── test_manager_enhanced.py
│   │   ├── test_exceptions_enhanced.py
│   │   ├── test_main.py
│   │   ├── test_models.py
│   │   ├── test_registry.py
│   │   └── test_*.py
│   ├── test_main_exception_handling.py # 19 tests
│   └── test_*.py
│
├── foundation/               # Foundation tests (v1.1.1)
│   ├── conftest.py
│   ├── test_backend.py
│   ├── test_commit_templates.py
│   ├── test_database.py
│   ├── test_devops.py
│   ├── test_frontend.py
│   ├── test_git.py
│   ├── test_ears_tdd.py
│   ├── test_langs_tdd.py
│   ├── test_testing_tdd.py
│   └── trust/               # TRUST 5 framework tests
│       └── test_trust_principles.py
│
├── integration/              # Integration tests
│   ├── cli/                  # CLI integration tests (v1.1.1)
│   ├── test_core/            # Core integration tests
│   └── test_workflow/        # Workflow tests
│
├── unit/                     # Additional unit tests (v1.1.1)
│   ├── cli/                  # CLI unit tests
│   │   └── commands/         # Command-specific unit tests
│   ├── core/                 # Core module unit tests
│   │   ├── analysis/
│   │   ├── integration/
│   │   ├── migration/
│   │   ├── project/
│   │   └── statusline/
│   ├── hooks/                # Hook unit tests
│   │   └── moai/
│   ├── statusline/           # Statusline unit tests
│   │   └── test_main_edge_cases.py # 51 tests
│   ├── tag_system/           # TAG System v2.0 tests
│   │   └── test_*.py
│   └── test_entry_points.py
│
├── conftest.py               # Pytest configuration
└── fixtures/                 # Test fixtures
```

**Note**: v1.1.1 added 179 new comprehensive tests covering edge cases, error handling, and type safety across core modules.

---

## Documentation Structure (docs/)

```
docs/
├── app/                      # Nextra documentation
│   ├── agents/               # Agent documentation
│   ├── commands/             # Command documentation
│   ├── skills/               # Skill documentation
│   └── workflows/            # Workflow guides
│
├── public/                   # Static assets
└── next.config.mjs           # Next.js configuration
```

---

## Key Files

| File | Purpose |
|------|---------|
| `pyproject.toml` | Single Source of Truth for package version and dependencies |
| `CLAUDE.md` | Alfred execution directives (Mr. Alfred orchestration rules) |
| `CLAUDE.local.md` | Local development guide (git-ignored) |
| `.moai/config/config.yaml` | Main MoAI configuration |
| `src/moai_adk/version.py` | Version reader (reads from pyproject.toml) |

---

## Sync Mechanism

Templates flow from `src/moai_adk/templates/` to root directories:

```
src/moai_adk/templates/.claude/  →  .claude/
src/moai_adk/templates/.moai/    →  .moai/
src/moai_adk/templates/CLAUDE.md →  ./CLAUDE.md
```

Local-only files (never synced):
- `.claude/settings.local.json`
- `CLAUDE.local.md`
- `.moai/cache/`, `.moai/memory/`, `.moai/logs/`
