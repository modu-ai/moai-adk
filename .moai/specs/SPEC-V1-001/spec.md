---
spec_id: SPEC-V1-001
title: Enterprise Plugin Ecosystem (v1.0) - 5 Official Plugins Development & Deployment
version: 1.0.0-rc1
status: Completed
owner: 🪿GOOS엉아 / MoAI-ADK Team
tags: ["plugin", "cli", "ui/ux", "fastapi", "devops", "v1.0"]
created: 2025-10-30
modified: 2025-10-31
language: en
---

# SPEC-V1-001: Enterprise Plugin Ecosystem (v1.0) - 5 Official Plugins Development & Deployment

## 📋 Overview

This SPEC defines the development and deployment of **5 official MoAI-ADK plugins** to support enterprise users implementing Claude Code plugins within their organizations.

### Strategic Goals

1. **Marketplace Establishment** — Publish moai-adk/moai-alfred-marketplace with plugin governance policies
2. **Plugin Reference Architecture** — Deliver production-grade plugin examples (Frontend/Backend/DevOps)
3. **Ecosystem Documentation** — Complete ch08-ch10 of "Claude Code & MoAI-ADK" book
4. **Team Onboarding** — Enable seamless plugin installation/configuration across teams

---

## 🎯 EARS Requirements

### Ubiquitous Behaviors (Base Features)

**1. PM Plugin (Project Management Kickoff)**

- GIVEN a user initiates `/init-pm` command
- WHEN project kickoff workflow starts
- THEN system generates EARS SPEC template, project charter, and risk matrix
- AND stores output in `.moai/specs/` directory

**2. UI/UX Plugin (Design System Foundation)**

- GIVEN a user runs `/setup-shadcn-ui` command
- WHEN Tailwind CSS + shadcn/ui initialization starts
- THEN system creates design tokens, component library, and Figma integration hooks
- AND outputs component checklist and usage guidelines

**3. Frontend Plugin (Next.js 16 + React 19.2)**

- GIVEN a user executes `/init-next` with package manager selection (bun|npm|pnpm)
- WHEN frontend project scaffolding begins
- THEN system creates Next.js 16 app with React 19.2, Biome, DevTools MCP, shadcn/ui
- AND generates routing structure, layout components, and API integration examples

**4. Backend Plugin (FastAPI + uv)**

- GIVEN a user runs `/init-fastapi` command
- WHEN backend project scaffolding starts
- THEN system creates FastAPI 0.120.2 project with uv, Pydantic 2.12, SQLAlchemy 2.0.44, Alembic 1.17
- AND initializes database schema, migration structure, and CRUD templates

**5. DevOps Plugin (Vercel/Supabase/Render MCP)**

- GIVEN a user calls `/deploy-config` command
- WHEN deployment configuration workflow starts
- THEN system generates environment files, secrets management setup, and CI/CD pipeline templates
- AND connects Vercel MCP (frontend), Supabase MCP (backend), Render MCP (alternative)

### Event-Driven Behaviors

**Plugin Installation Events**:

- WHEN `/plugin marketplace add moai-adk/moai-alfred-marketplace` is executed
- THEN system validates marketplace JSON schema, loads plugin index, and registers available plugins

**Plugin Configuration Events**:

- WHEN plugin `.claude-plugin/plugin.json` is loaded
- THEN system parses commands, agents, hooks, MCP servers
- AND validates tool permissions (allowed-tools, denied-tools)
- AND initializes plugin context in `.claude/settings.local.json`

**MCP Server Connection Events**:

- WHEN `/connect-mcp-devtools` or `/connect-mcp-figma` is invoked
- THEN system establishes secure OAuth/token-based connection
- AND stores credentials in OS Keychain (no VCS tracking)
- AND validates MCP server accessibility

### State-Driven Behaviors

**Plugin State Machine**:

- State: `installed` → `configured` → `ready` → `active`
- WHEN plugin moves from `installed` to `configured`
- THEN system validates all commands, agents, skills are loaded
- AND runs schema validation (hooks.json, plugin.json, commands/\*.md)

**Development Progress Tracking**:

- Plugin Development Status: `skeleton` → `dev` → `rc` → `stable` → `deprecated`
- WHEN plugin version changes
- THEN system updates marketplace JSON, generates release notes, and triggers CHANGELOG sync

### Optional Behaviors

**Package Manager Selection (Frontend Plugin)**:

- GIVEN user preference for bun|npm|pnpm
- WHEN `/init-next` command receives pm argument
- THEN EITHER install with selected PM OR default to bun if unspecified
- AND update lock file (bun.lock|package-lock.json|pnpm-lock.yaml)

**Database Selection (Backend Plugin)**:

- GIVEN user choice of PostgreSQL 18 or MySQL 8.4 LTS
- WHEN `/db-setup` receives driver argument
- THEN EITHER configure asyncpg (Postgres) OR aiomysql (MySQL)
- AND generate appropriate connection string in `.env`

**Custom MCP Registrations**:

- GIVEN user has proprietary MCP servers (internal tools, APIs)
- WHEN user manually adds entries to `.mcp.json`
- THEN EITHER use local MCP OR register remote OAuth/token-based MCP
- AND validate MCP server capability discovery

### Unwanted Behaviors (Prohibitions)

**❌ Security Prohibitions**:

- NEVER store API keys, tokens, or credentials in version-controlled files (`.env`, `.claude/settings.json`)
- NEVER execute destructive bash commands (rm -rf, git reset --hard) without explicit user confirmation
- NEVER bypass tool permission checks (allowed-tools validation must pass)
- NEVER expose secrets in git commit messages or PR descriptions

**❌ Configuration Errors**:

- NEVER allow plugin installation without marketplace schema validation
- NEVER permit conflicting MCP server registrations (duplicate names)
- NEVER override system-level settings without user approval

**❌ Data Loss Risks**:

- NEVER auto-delete existing `.moai/` directories without backup confirmation
- NEVER overwrite user project files during scaffolding (add .new suffix if collision detected)

---

## 🏗️ Implementation Architecture

### Marketplace Structure (moai-adk/moai-alfred-marketplace)

```
moai-alfred-marketplace/
├── .claude-plugin/
│   └── marketplace.json                    ← Plugin index + org policies
├── plugins/
│   ├── moai-alfred-pm/
│   │   ├── .claude-plugin/
│   │   │   ├── plugin.json
│   │   │   └── hooks.json
│   │   ├── commands/
│   │   │   ├── init-pm.md
│   │   │   ├── spec-template.md
│   │   │   └── risk-matrix.md
│   │   ├── agents/
│   │   │   └── pm-agent.md
│   │   ├── skills/
│   │   │   └── SKILL-PM-*.md
│   │   ├── .mcp.json                      ← Figma MCP (optional)
│   │   └── README.md
│   ├── moai-alfred-uiux/                      ← Tailwind + shadcn/ui
│   ├── moai-alfred-frontend/                  ← Next.js 16 + React 19.2
│   ├── moai-alfred-backend/                   ← FastAPI + uv
│   └── moai-alfred-devops/                    ← Vercel/Supabase/Render
├── README.md                              ← Installation instructions
├── SECURITY.md                            ← Org policies (keys, registries)
└── CHANGELOG.md                           ← Version history
```

### Plugin Components (Each Plugin Delivers)

| Component      | Files                   | Purpose                              |
| -------------- | ----------------------- | ------------------------------------ |
| **Command**    | `commands/*.md`         | Slash commands with FM/tools/args    |
| **Agent**      | `agents/*.md`           | Sub-agents for complex workflows     |
| **Skill**      | `skills/SKILL-*.md`     | Reusable knowledge (YAML frontmatter |
| **Hook**       | `hooks/hooks.json`      | Event-triggered automation           |
| **MCP Server** | `.mcp.json`             | External tool integrations (OAuth)   |
| **Tests**      | `tests/test_*.py`       | Plugin validation (pytest)           |
| **Docs**       | `README.md`, `USAGE.md` | Installation, configuration, FAQ     |

---

## 📦 5 Official Plugins Specification

### Plugin 1: PM Plugin (Project Management Kickoff)

**Purpose**: Automate project initiation (EARS SPEC template, project charter, risk assessment)

**Stack**:

- Command: `/init-pm [project-name] [domain?=software|hardware|service]`
- Agent: PM-Agent (project manager super agent)
- Skills: SPEC template, Risk analysis, Stakeholder mapping
- MCP: Figma MCP (optional: design mockup integration)

**Deliverables**:

- `.moai/specs/SPEC-{PROJECT}-001/spec.md` (EARS-formatted)
- `.moai/specs/SPEC-{PROJECT}-001/plan.md` (breakdown, milestones)
- `.moai/specs/SPEC-{PROJECT}-001/acceptance.md` (QA criteria)
- `.moai/docs/project-charter.md` (vision, scope, constraints)
- `.moai/analysis/risk-matrix.md` (risk probability × impact grid)

---

### Plugin 2: UI/UX Plugin (Design System)

**Purpose**: Establish Tailwind CSS + shadcn/ui design system with Figma integration

**Stack**:

- Command: `/setup-shadcn-ui [components?=button|input|card|...] [pm=bun|npm|pnpm]`
- Agent: Design-Agent (component library curator)
- Skills: Tailwind config, shadcn component registry, design token management
- MCP: Figma MCP (optional: Figma → shadcn/ui sync)

**Deliverables**:

- `tailwind.config.ts` (custom theme, design tokens)
- `globals.css` (Tailwind base + shadcn CSS variables)
- `lib/cn.ts` (classname utility)
- `components/ui/` (scaffolded shadcn components)
- `.moai/docs/design-system.md` (component guidelines, usage examples)

---

### Plugin 3: Frontend Plugin (Next.js 16 + React 19.2)

**Purpose**: Full-stack Next.js 16 frontend with React 19.2, Biome, DevTools MCP, shadcn/ui

**Stack**:

- Node.js 24 LTS, Next.js 16, React 19.2
- Package manager: bun 1.3.x (default) | npm | pnpm
- Formatter/Linter: Biome 2.x
- Component library: shadcn/ui (Tailwind-based)
- DevTools MCP for debugging, routing diagnosis

**Commands**:

- `/init-next [app-name] [pm=bun|npm|pnpm]` — Create Next.js 16 scaffolding
- `/biome-setup` — Configure Biome, add pre-commit hook
- `/setup-shadcn-ui [components]` — Initialize Tailwind + shadcn/ui
- `/connect-mcp-devtools` — Register Next DevTools MCP
- `/routes-diagnose` — Analyze routing, build errors, performance metrics

**Deliverables**:

- `app/layout.tsx`, `app/page.tsx` (App Router structure)
- `app/api/` (API routes example)
- `components/` (shadcn/ui scaffold)
- `.biomerc.json` (formatter/linter config)
- `bun.lock` / `package-lock.json` / `pnpm-lock.yaml` (lockfile)
- `.next/` (build cache)
- `.env.local` (dev secrets, .gitignore'd)

---

### Plugin 4: Backend Plugin (FastAPI + uv)

**Purpose**: Enterprise Python backend with FastAPI, SQLAlchemy, Alembic, uv dependency management

**Stack**:

- Python 3.14 (via uv)
- FastAPI 0.120.2, Uvicorn
- Pydantic 2.12 (data validation)
- SQLAlchemy 2.0.44 (ORM)
- Alembic 1.17 (database migrations)
- pytest (testing framework)
- ruff (linting), mypy (type checking)
- uv (dependency management: venv, pip, lock, run)

**Commands**:

- `/init-fastapi [app-name]` — Create FastAPI 0.120.2 project with uv
- `/db-setup [driver=postgres|mysql]` — Initialize database connection + Alembic
- `/resource-crud [resource-name]` — Generate CRUD routes from EARS SPEC
- `/run-dev` — Start uvicorn with auto-reload + .env loading
- `/api-profile` — Performance metrics (latency p95, throughput, memory)

**Deliverables**:

- `app/main.py` (FastAPI application entry point)
- `app/api/` (API route modules)
- `app/models/` (SQLAlchemy ORM models)
- `app/schemas/` (Pydantic request/response schemas)
- `app/db/` (database session, config)
- `alembic/versions/` (migrations)
- `tests/` (pytest test suite)
- `.env.example` (template for .env)
- `uv.lock` (locked dependencies)
- `pyproject.toml` (project config with tool.uv.index for custom registries)

**Database Support**:

- PostgreSQL 18 (asyncpg driver, async-first)
- MySQL 8.4 LTS (aiomysql driver, async option)

---

### Plugin 5: DevOps Plugin (Vercel/Supabase/Render MCP)

**Purpose**: Multi-cloud deployment configuration (Vercel frontend, Supabase backend, Render alternative)

**Stack**:

- MCP: Vercel MCP (frontend deployment)
- MCP: Supabase MCP (Postgres + Auth + Storage)
- MCP: Render MCP (alternative backend host)
- CI/CD: GitHub Actions (example workflows)

**Commands**:

- `/deploy-config [platform=vercel|render] [backend=supabase|custom-api]` — Generate deployment files
- `/connect-vercel` — OAuth authentication with Vercel account
- `/connect-supabase` — Register Supabase project, generate .env
- `/generate-github-actions` — Create workflows (test, build, deploy)
- `/secrets-setup` — OS Keychain / CI environment variable registration

**Deliverables**:

- `vercel.json` (Vercel deployment config)
- `supabase/` (Supabase project setup)
- `.github/workflows/` (CI/CD pipeline)
- `.env.production` template (secrets management strategy)
- `.moai/docs/deployment-guide.md` (step-by-step for 3 platforms)
- `scripts/backup-db.sh` (database backup automation)

---

## 📚 Documentation Deliverables (Book Chapters)

### ch08: Claude Code Plugins Introduction & Migration (Output Styles → Plugins/Hooks/Skills)

**Topics**:

1. Why plugin ecosystem (extensibility, governance, team collaboration)
2. Plugin components (commands, agents, skills, hooks, MCP)
3. Plugin.json schema (metadata, tool permissions, MCP servers)
4. Slash command authoring (FM, arguments, allowed-tools)
5. Hook event system (PreToolUse, PostToolUse, SessionStart)
6. Skill auto-loading and context invocation
7. MCP server integration (.mcp.json, OAuth, token management)
8. Marketplace publishing (CI/CD, version locking, org policies)
9. Migration from Output Styles (deprecated EOL) to plugins

**Hands-on Examples**:

- Write custom slash command (`/my-linter-check`)
- Create plugin-scoped skill (`SKILL-MY-PLUGIN-FORMATTER.md`)
- Connect local MCP server
- Publish plugin to marketplace.json

---

### ch09: MoAI-ADK 5-Plugin Development & Deployment Workflow

**Structure**:

- 9-1: PM Plugin hands-on walkthrough
- 9-2: UI/UX Plugin (shadcn/ui setup, design tokens)
- 9-3: Frontend Plugin (Next.js 16, Biome, DevTools MCP)
- 9-4: Backend Plugin (FastAPI, uv, multi-database)
- 9-5: DevOps Plugin (Vercel/Supabase/Render 3-path deployment)

**Each Section Includes**:

- Plugin architecture diagram
- Command reference table
- Step-by-step installation (worktree-based development)
- Configuration checklist (secrets, MCP registration)
- Troubleshooting FAQ

---

### ch10: Full Project — Personal Blog Platform (Next.js + FastAPI + Supabase)

**Project Scope**:

- **Frontend**: Next.js 16, React 19.2, shadcn/ui, Markdown editor (Tiptap 3)
- **Backend**: FastAPI 0.120.2, SQLAlchemy 2.0.44, Pydantic 2.12, RBAC (role-based access)
- **Database**: Supabase (PostgreSQL 18)
- **Deployment**: Vercel (frontend) + Render (backend API) + Supabase (database)

**Features**:

1. Admin panel (RBAC: admin, editor, viewer)
2. Post management (draft, published, scheduled, deleted)
3. Markdown editor with preview (Tiptap 3)
4. Image upload (Supabase Storage)
5. RSS feed generation
6. Comment system with moderation
7. Analytics dashboard
8. Multi-tenant support (future)

**Chapters**:

- 10-1: Requirements & Design (EARS SPEC, data model, API routes)
- 10-2: Frontend implementation (App Router, components, state management)
- 10-3: Backend implementation (CRUD APIs, authentication, migrations)
- 10-4: Deployment & Operations (CI/CD, secrets, monitoring, backups)

---

## 🔧 Version Lock Matrix (Stable 2025-10-30)

| Component    | Version       | Notes                                    |
| ------------ | ------------- | ---------------------------------------- |
| Node.js      | 24 LTS        | Latest LTS from nodejs.org               |
| npm/pnpm     | Latest        | Auto-sync with Node 24                   |
| bun          | 1.3.x         | Stable, full npm compat                  |
| Next.js      | 16            | GA release, React 19.2 compatible        |
| React        | 19.2          | Latest stable                            |
| Tailwind CSS | 4.x           | Latest version                           |
| shadcn/ui    | Latest (main) | CLI-based component system               |
| Biome        | 2.x           | Formatter + linter                       |
| Python       | 3.14          | Latest via uv --default                  |
| FastAPI      | 0.120.2       | PyPI release                             |
| Pydantic     | 2.12.0        | Full v2 stable                           |
| SQLAlchemy   | 2.0.44        | 2.0 LTS track                            |
| Alembic      | 1.17.0        | Latest stable                            |
| uv           | Latest        | Automatic Python/package/lock management |
| PostgreSQL   | 18            | Latest GA                                |
| MySQL        | 8.4 LTS       | Long-term support release                |

---

## 🎓 Success Criteria

### Deliverables Checklist

- ✅ moai-alfred-marketplace created (public GitHub repo)
- ✅ 5 plugins fully implemented & tested (each in separate directory)
- ✅ marketplace.json published with org policies (SECURITY.md)
- ✅ ch08 chapter complete (plugin ecosystem guide)
- ✅ ch09 chapter complete (5-plugin workflow)
- ✅ ch10 chapter complete (blog platform project)
- ✅ Plugin testing framework (pytest for each plugin)
- ✅ CI/CD pipeline (GitHub Actions for plugin release)
- ✅ Documentation (README, USAGE, troubleshooting for each plugin)

### Quality Metrics

- **Plugin Code Coverage**: ≥85% (pytest)
- **Documentation**: ≥95% inline code comments (English)
- **Security**: No hardcoded secrets, all OAuth/tokens in OS Keychain
- **Linting**: 0 errors (ruff/Biome)
- **Type Safety**: 100% type hints (mypy strict mode)
- **Release**: Semantic versioning (v1.0.0-rc1 → v1.0.0)

---

## 📅 Development Timeline

| Phase                    | Duration  | Deliverables                                     |
| ------------------------ | --------- | ------------------------------------------------ |
| **Phase 1: Setup**       | Week 1    | Worktree, SPEC docs, plugin skeleton             |
| **Phase 2: Plugins 1-3** | Weeks 2-3 | PM, UI/UX, Frontend plugins (tests + docs)       |
| **Phase 3: Plugins 4-5** | Week 4    | Backend, DevOps plugins (tests + docs)           |
| **Phase 4: Integration** | Week 5    | Marketplace release, ch08-ch10 finalization      |
| **Phase 5: Release**     | Week 6    | RC testing, v1.0.0 release, docs synchronization |

---

## 🔐 Governance & Security

### Marketplace Policies (SECURITY.md)

**Permissions Model**:

- `allowed-tools`: Bash(bunx|npx|pnpm|npm|git|uv|uvicorn), Read, Write, Bash(python|python3)
- `denied-tools`: Destructive commands (rm -rf, git reset --hard), file system outside project

**Secret Management**:

- API Keys: OS Keychain (macOS) / Secrets Manager (Windows) / `pass` (Linux)
- Database credentials: `.env.local` (NOT in git)
- CI/CD tokens: GitHub Actions encrypted secrets

**Registry Configuration**:

- npm: NPM_CONFIG_REGISTRY (custom registry URL)
- pip: UV_INDEX_URL or [tool.uv.index] in pyproject.toml

**Version Pinning**:

- All plugins lock versions in plugin.json (version field)
- Marketplace.json pins plugin versions (moai-alfred-frontend@0.1.0)
- pyproject.toml uses exact versions for reproducibility (==, not ~=)

---

## 📞 HISTORY & References

**Related SPECs**:

- SPEC-ALFRED-001: Alfred SuperAgent Workflow (foundation)
- SPEC-TRUST-001: TRUST 5 Principles (quality gates)

**External References**:

- Claude Code Plugins: https://docs.claude.com/en/docs/claude-code/plugins
- shadcn/ui CLI: https://github.com/shadcn-ui/ui (init/add commands)
- uv Documentation: https://docs.astral.sh/uv/
- FastAPI: https://fastapi.tiangolo.com/
- Next.js 16: https://nextjs.org/docs
- React 19: https://react.dev/

---

**Status**: This SPEC is Ready for Development (Phase 1 initiated)
**Last Updated**: 2025-10-30 · **Next Review**: Weekly sprint sync
