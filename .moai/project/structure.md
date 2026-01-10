# MoAI-ADK Project Structure

> **Last Updated**: 2026-01-10
> **Version**: 0.41.2

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
│   ├── claude.py             # Claude Code integration
│   ├── core.py               # Core principles
│   └── quality.py            # Quality framework
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
    └── moai/                 # MoAI skills (47 skills)
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
│   ├── test_core/
│   ├── test_cli/
│   └── test_utils/
│
├── integration/              # Integration tests
│   ├── test_workflow/
│   └── test_agents/
│
├── e2e/                      # End-to-end tests
│   └── test_full_cycle/
│
├── web/                      # Web API tests
│   └── test_*_router.py
│
├── conftest.py               # Pytest configuration
└── fixtures/                 # Test fixtures
```

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
