# SPEC-V1-001 Acceptance Criteria

## ✅ Definition of Done (DoD)

### Each Plugin Component

**Command Implementation**:
- ✅ Slash command registered in `.claude-plugin/plugin.json`
- ✅ Command markdown file (`.md`) documents arguments, allowed-tools, examples
- ✅ Help text displays when user runs `/command-name --help`
- ✅ Tool permissions (allowed-tools) are minimally scoped (least privilege)
- ✅ No hardcoded secrets in command code
- ✅ Error handling with user-friendly messages

**Agent Implementation**:
- ✅ Agent markdown (`.md`) describes persona and decision-making rules
- ✅ Agent responds in user's configured conversation_language
- ✅ Agent output includes structured summaries (task lists, tables, code blocks)
- ✅ Agent error handling escalates blockers clearly

**Skill Implementation**:
- ✅ SKILL-*.md file with YAML frontmatter (metadata, templates, examples)
- ✅ Skill provides reusable patterns/checklists for model to invoke
- ✅ Skill content is under 500 words (progressive disclosure)
- ✅ Skill examples are runnable/testable

**Hook Implementation**:
- ✅ hooks.json defines event triggers (PreToolUse, PostToolUse, SessionStart, etc.)
- ✅ Hook actions are lightweight (<100 ms execution)
- ✅ Hook prevents destructive actions without confirmation
- ✅ Hook logs executed actions (for audit trail)

**MCP Server Integration**:
- ✅ .mcp.json defines remote MCP servers (Figma, DevTools, Vercel, Supabase, etc.)
- ✅ MCP connection requires OAuth or token-based auth (no hardcoded creds)
- ✅ MCP credentials stored in OS Keychain, not in .json files
- ✅ MCP server capability discovery works on first connection

---

### Testing Standards

**Code Coverage**:
- ✅ Python plugins: ≥85% pytest coverage (commands, agents, models)
- ✅ JavaScript/TypeScript plugins: ≥80% Jest/Vitest coverage (components, utils)
- ✅ Test files: `tests/test_*.py` or `tests/test_*.ts`
- ✅ CI/CD pipeline runs tests on every commit

**Type Safety**:
- ✅ Python: 100% type hints (mypy strict mode, no `Any` without explanation)
- ✅ TypeScript: `strict: true` in tsconfig.json
- ✅ Type errors: 0 in pre-commit checks

**Linting & Formatting**:
- ✅ Python: ruff with MoAI-ADK rules (0 errors)
- ✅ TypeScript/JavaScript: Biome with MoAI-ADK config (0 errors)
- ✅ Pre-commit hooks enforce linting before git push

**Documentation**:
- ✅ All public functions/classes have docstrings (Google-style)
- ✅ README.md covers: installation, configuration, command reference
- ✅ USAGE.md provides step-by-step examples for each command
- ✅ Inline comments explain non-obvious logic
- ✅ All documentation in English (code standard) + user language guides

---

### Security Acceptance

**Secret Management**:
- ✅ No API keys/tokens in .json, .md, .py, .ts files
- ✅ Credentials loaded from OS Keychain / .env.local (never committed)
- ✅ .env.local is in .gitignore
- ✅ CI/CD uses GitHub Actions Secrets for sensitive data
- ✅ OAuth flows use redirect_uri validation

**Tool Permissions**:
- ✅ Each command lists only required allowed-tools
- ✅ No Bash(rm -rf /), git reset --hard, or destructive commands without confirmation
- ✅ Tool permissions are validated on plugin load
- ✅ Denied-tools explicitly list prohibited commands

**Vulnerability Scanning**:
- ✅ Python dependencies: pip-audit reports 0 critical/high CVEs
- ✅ npm dependencies: npm audit reports 0 critical/high vulnerabilities
- ✅ No known supply-chain compromises

**Data Privacy**:
- ✅ No telemetry/analytics without user opt-in
- ✅ User project files only accessed within project scope
- ✅ MCP connections logged (but not sensitive payloads)

---

### Version Locking & Reproducibility

**Dependency Pinning**:
- ✅ `plugin.json`: version field (e.g., "0.1.0")
- ✅ `marketplace.json`: all plugins pinned (e.g., "moai-cc-frontend@0.1.0")
- ✅ `pyproject.toml`: exact versions with `==` (not `~=` or `^`)
- ✅ `package.json`: exact versions or narrowed ranges (e.g., "16.0.0")
- ✅ Lock files committed: `uv.lock`, `bun.lock`, `package-lock.json`, `pnpm-lock.yaml`

**Platform Support**:
- ✅ Windows / macOS / Linux compatibility verified
- ✅ Python 3.11+ supported (tested on 3.11, 3.12, 3.13, 3.14)
- ✅ Node 24 LTS verified for Frontend plugin
- ✅ Shell scripts POSIX-compliant (or bash with shebang)

---

## 📋 Individual Plugin Acceptance Checklists

### PM Plugin Acceptance

**Commands**:
- ✅ `/init-pm [project-name] [domain?]` creates SPEC structure
- ✅ SPEC template includes EARS requirements (ubiquitous, event-driven, state-driven, optional, unwanted)
- ✅ Project charter includes vision, scope, success criteria, risks
- ✅ Risk matrix uses probability × impact grid (3×3 or 5×5)

**Outputs**:
- ✅ `.moai/specs/SPEC-{PROJECT}-001/spec.md` (EARS format)
- ✅ `.moai/specs/SPEC-{PROJECT}-001/plan.md` (WBS, timeline, milestones)
- ✅ `.moai/specs/SPEC-{PROJECT}-001/acceptance.md` (this format)
- ✅ `.moai/docs/project-charter.md` (vision/scope/constraints)
- ✅ `.moai/analysis/risk-matrix.md` (risk grid with mitigations)

**Tests**:
- ✅ pytest test_pm_spec_template.py (validates EARS structure)
- ✅ pytest test_pm_charter_generation.py (checks required sections)
- ✅ pytest test_pm_risk_analysis.py (validates risk matrix format)

**Documentation**:
- ✅ README.md (installation, `/init-pm --help`)
- ✅ USAGE.md (step-by-step PM workflow)
- ✅ FAQ.md (common PM issues)

---

### UI/UX Plugin Acceptance

**Commands**:
- ✅ `/setup-shadcn-ui [components?] [pm=bun|npm|pnpm]` initializes design system
- ✅ Tailwind v4 configured with custom theme
- ✅ shadcn/ui components installable via `/setup-shadcn-ui button input card`

**Outputs**:
- ✅ `tailwind.config.ts` with design tokens (colors, spacing, fonts)
- ✅ `globals.css` with Tailwind base + shadcn CSS variables
- ✅ `lib/cn.ts` (className utility)
- ✅ `components/ui/` (shadcn components scaffold)
- ✅ `.moai/docs/design-system.md` (component guidelines)

**Tests**:
- ✅ Component rendering tests (React Testing Library or Vitest snapshot)
- ✅ CSS variable validation (all colors exist in theme)
- ✅ Tailwind purge check (no unused classes)

**Documentation**:
- ✅ README.md (shadcn/ui setup guide)
- ✅ Component inventory (list all available components + usage)
- ✅ Design token guide (color palette, spacing scale)

---

### Frontend Plugin Acceptance

**Commands**:
- ✅ `/init-next [app-name] [pm=bun|npm|pnpm]` scaffolds Next.js 16 + React 19.2 app
- ✅ `/biome-setup` configures Biome formatter/linter + pre-commit hook
- ✅ `/setup-shadcn-ui [components]` installs shadcn/ui components
- ✅ `/connect-mcp-devtools` registers Next.js DevTools MCP
- ✅ `/routes-diagnose` shows routing structure and build errors

**Outputs**:
- ✅ `app/layout.tsx`, `app/page.tsx` (App Router structure)
- ✅ `app/api/` directory (API routes example)
- ✅ `components/` directory (UI component scaffold)
- ✅ `.biomerc.json` (formatter + linter config)
- ✅ `.husky/pre-commit` (Biome hook)
- ✅ `bun.lock` / `package-lock.json` / `pnpm-lock.yaml` (lockfile per PM choice)
- ✅ `next.config.js` (optimization settings)
- ✅ `.env.local` template (dev secrets)

**Tests**:
- ✅ Next.js build test (npm run build succeeds)
- ✅ Component snapshot tests (Vitest)
- ✅ Route validation (all routes are registered)
- ✅ Biome linting test (0 errors)

**Package Manager Support**:
- ✅ Bun 1.3.x: `bun install`, `bun run dev`, generates `bun.lock`
- ✅ npm: `npm install`, `npm run dev`, generates `package-lock.json`
- ✅ pnpm: `pnpm install`, `pnpm dev`, generates `pnpm-lock.yaml`

**Documentation**:
- ✅ README.md (Next.js 16 setup + package manager choices)
- ✅ USAGE.md (common tasks: add page, component, API route)
- ✅ DevTools MCP integration guide

---

### Backend Plugin Acceptance

**Commands**:
- ✅ `/init-fastapi [app-name]` scaffolds FastAPI 0.120.2 project with uv
- ✅ `/db-setup [driver=postgres|mysql]` initializes Alembic + DB connection
- ✅ `/resource-crud [resource-name]` generates models/schemas/routes from SPEC
- ✅ `/run-dev` starts uvicorn with auto-reload + .env loading
- ✅ `/api-profile` measures API latency, throughput, memory usage

**Outputs**:
- ✅ `app/main.py` (FastAPI entry point)
- ✅ `app/api/` (route modules)
- ✅ `app/models/` (SQLAlchemy ORM models)
- ✅ `app/schemas/` (Pydantic request/response schemas)
- ✅ `app/db/` (database session, config)
- ✅ `app/core/config.py` (Settings with Pydantic)
- ✅ `alembic/versions/` (migration files)
- ✅ `tests/` (pytest suite)
- ✅ `.env.example` (template for .env)
- ✅ `pyproject.toml` (with Python 3.14 support + tool.uv.index for custom registries)
- ✅ `uv.lock` (frozen dependencies)

**Database Support**:
- ✅ PostgreSQL 18 (asyncpg for async, psycopg for sync)
- ✅ MySQL 8.4 LTS (aiomysql for async, mysqlclient for sync)
- ✅ Connection string in `.env.local` (never hardcoded)

**Tests**:
- ✅ Model tests (SQLAlchemy ORM validation)
- ✅ Schema tests (Pydantic validation)
- ✅ API route tests (FastAPI TestClient)
- ✅ Migration tests (Alembic upgrade/downgrade)
- ✅ Coverage: ≥85% (pytest cov)

**uv Integration**:
- ✅ `uv venv .venv` creates virtual environment
- ✅ `uv pip install -e ".[dev]"` installs dev dependencies
- ✅ `uv run uvicorn app.main:app --reload` runs server with auto-reload
- ✅ `uv lock` generates frozen lock file
- ✅ `UV_INDEX_URL` or `[tool.uv.index]` for custom PyPI mirror/registry

**Documentation**:
- ✅ README.md (FastAPI + uv + database setup)
- ✅ USAGE.md (common tasks: add model, create migration, write test)
- ✅ Database migration guide
- ✅ Async/await patterns documentation
- ✅ API documentation (auto-generated via Swagger/OpenAPI)

---

### DevOps Plugin Acceptance

**Commands**:
- ✅ `/deploy-config [platform=vercel|render] [backend=supabase|custom-api]` generates deployment files
- ✅ `/connect-vercel` OAuth login to Vercel account
- ✅ `/connect-supabase` OAuth/token setup for Supabase project
- ✅ `/generate-github-actions` creates .github/workflows/ (test, build, deploy)
- ✅ `/secrets-setup` registers CI/CD secrets in GitHub Actions

**Outputs**:
- ✅ `vercel.json` (Vercel deployment config)
- ✅ `supabase/` project structure (if using Supabase)
- ✅ `.github/workflows/test.yml` (pytest/vitest on PR)
- ✅ `.github/workflows/build.yml` (build & artifacts)
- ✅ `.github/workflows/deploy.yml` (to Vercel/Render)
- ✅ `.env.production` template (secrets structure)
- ✅ `.moai/docs/deployment-guide.md` (3-path: Vercel+Supabase, Render+Postgres, custom)

**Multi-Cloud Support**:
- ✅ **Path 1**: Vercel (frontend) + Supabase (backend + DB)
- ✅ **Path 2**: Render (frontend + backend) + PostgreSQL
- ✅ **Path 3**: Custom cloud (AWS/GCP/Azure) + DB

**Tests**:
- ✅ YAML schema validation (.github/workflows/*.yml)
- ✅ Secrets management test (no hardcoded creds in workflows)
- ✅ Deployment simulation (dry-run mode)

**Documentation**:
- ✅ README.md (3 deployment paths)
- ✅ USAGE.md (step-by-step deployment)
- ✅ Troubleshooting guide (common CI/CD errors)
- ✅ Backup & recovery guide
- ✅ Monitoring & alerting setup

---

## 📦 Marketplace Acceptance

**marketplace.json**:
- ✅ Valid JSON schema (metadata, owner, plugins array)
- ✅ All 5 plugins registered with correct metadata
- ✅ Version locking (each plugin pinned to specific version)
- ✅ Plugin categories (frontend, backend, devops, pm, uiux)

**SECURITY.md**:
- ✅ Org policies documented (permissions, secrets, registries)
- ✅ Key rotation strategy (quarterly recommendation)
- ✅ Registry configuration (NPM_CONFIG_REGISTRY, UV_INDEX_URL examples)
- ✅ OAuth/token management guidance

**README.md**:
- ✅ Installation instructions (`/plugin marketplace add ...`)
- ✅ Plugin list with descriptions
- ✅ Quick-start guide (install 1-2 plugins, run example command)
- ✅ Support & feedback channels

**CI/CD Pipeline**:
- ✅ .github/workflows/plugin-validate.yml (schema + security checks on PR)
- ✅ .github/workflows/plugin-release.yml (tag → GitHub Release)
- ✅ Automated version bump (semver: major.minor.patch)

---

## 📚 Book Chapter Acceptance

### ch08: Claude Code Plugins & Migration

**Sections**:
- ✅ 8-1: Why Plugin Ecosystem (extensibility, governance, team collab)
- ✅ 8-2: Plugin Components Deep-Dive (commands, agents, skills, hooks, MCP)
- ✅ 8-3: plugin.json Schema Reference (with examples)
- ✅ 8-4: Slash Command Authoring (FM, arguments, tool permissions)
- ✅ 8-5: Hook Event System (PreToolUse, PostToolUse, SessionStart)
- ✅ 8-6: Skill Auto-Loading & Context Invocation
- ✅ 8-7: MCP Server Integration (.mcp.json, OAuth, token management)
- ✅ 8-8: Marketplace Publishing (CI/CD, version locking, org policies)
- ✅ 8-9: Migration Guide (from Output Styles to plugins/hooks/skills)

**Hands-On Labs**:
- ✅ Lab 8A: Write custom slash command (`/my-linter-check`)
- ✅ Lab 8B: Create plugin-scoped skill (SKILL-MY-PLUGIN-FORMATTER.md)
- ✅ Lab 8C: Connect local MCP server (manifest.json, stdio/http)
- ✅ Lab 8D: Publish plugin to marketplace.json

**Quality**:
- ✅ All code examples are tested & runnable
- ✅ Screenshots/diagrams for complex concepts
- ✅ FAQs section (common issues & solutions)

---

### ch09: 5-Plugin Development Workflow

**9-1: PM Plugin**:
- ✅ Why project management plugin (kickoff automation)
- ✅ Step-by-step: `/init-pm myapp`
- ✅ Output walkthrough (SPEC, charter, risk matrix)
- ✅ Integration with `/alfred:1-plan` SPEC workflow
- ✅ FAQ & troubleshooting

**9-2: UI/UX Plugin**:
- ✅ Why design system plugin (consistency, scalability)
- ✅ Step-by-step: `/setup-shadcn-ui button input card`
- ✅ Design token customization guide
- ✅ shadcn/ui component inventory
- ✅ Figma MCP integration (optional)
- ✅ FAQ & troubleshooting

**9-3: Frontend Plugin**:
- ✅ Why Next.js 16 + React 19.2 (latest stable)
- ✅ Step-by-step: `/init-next myapp bun`
- ✅ Package manager choices (bun vs npm vs pnpm)
- ✅ Biome setup & pre-commit hooks
- ✅ DevTools MCP integration
- ✅ Common patterns (pages, API routes, middleware)
- ✅ FAQ & troubleshooting

**9-4: Backend Plugin**:
- ✅ Why FastAPI + uv (performance, simplicity, dependency management)
- ✅ Step-by-step: `/init-fastapi myapi`
- ✅ Database setup (PostgreSQL vs MySQL)
- ✅ CRUD generation from SPEC (`/resource-crud users`)
- ✅ Migration workflow with Alembic
- ✅ Custom PyPI index setup (UV_INDEX_URL)
- ✅ Testing patterns (pytest, fixtures)
- ✅ FAQ & troubleshooting

**9-5: DevOps Plugin**:
- ✅ Why 3-cloud deployment paths
- ✅ Step-by-step: `/deploy-config vercel`
- ✅ Secrets management (OS Keychain, GitHub Actions)
- ✅ GitHub Actions workflow walkthrough
- ✅ Vercel + Supabase path (1)
- ✅ Render path (2)
- ✅ Custom cloud path (3)
- ✅ Monitoring & alerting setup
- ✅ FAQ & troubleshooting

---

### ch10: Full Project - Personal Blog Platform

**10-1: Requirements & Design**:
- ✅ EARS SPEC (user stories, features)
- ✅ Data model diagram (ERD: posts, users, comments, tags)
- ✅ API routes list (OpenAPI/Swagger)
- ✅ Architecture diagram (frontend, backend, database)
- ✅ Deployment architecture (3 paths)

**10-2: Frontend Implementation**:
- ✅ Project scaffolding (`/init-next blog-app bun`)
- ✅ App Router structure (app/ vs pages/)
- ✅ shadcn/ui layout (header, sidebar, main content)
- ✅ Post list page (pagination, filtering)
- ✅ Post detail page (markdown rendering)
- ✅ Admin dashboard (RBAC: admin, editor, viewer)
- ✅ Markdown editor integration (Tiptap 3 or similar)
- ✅ Image upload (Supabase Storage or Vercel Blob)
- ✅ Comment section (moderation queue)
- ✅ State management (React Context or Zustand)
- ✅ Testing (component snapshot tests)
- ✅ Build & deploy to Vercel

**10-3: Backend Implementation**:
- ✅ Project scaffolding (`/init-fastapi blog-api`)
- ✅ Database schema (Alembic migrations)
- ✅ Models (Post, User, Comment, Tag)
- ✅ Schemas (Pydantic request/response)
- ✅ Routes (CRUD for posts, users, comments, auth)
- ✅ Authentication (JWT with FastAPI-Users or similar)
- ✅ RBAC (role-based access control)
- ✅ Pagination & filtering
- ✅ Error handling (custom exception handlers)
- ✅ Testing (pytest with fixtures)
- ✅ Database migrations (Alembic workflow)
- ✅ Build & deploy (Render or custom cloud)

**10-4: Deployment & Operations**:
- ✅ Environment setup (.env.local, .env.production)
- ✅ GitHub Actions CI/CD setup
- ✅ Database backup strategy (automated daily backups)
- ✅ Secrets rotation (quarterly keys)
- ✅ Monitoring dashboard (logs, metrics, alerts)
- ✅ Health checks (endpoint monitoring)
- ✅ Disaster recovery plan (restore from backup)
- ✅ Post-deployment checklist (DNS, SSL, monitoring)

---

## 🎯 Success Metrics

### Quantitative

| Metric                    | Target |
| ------------------------- | ------ |
| Plugin Code Coverage      | ≥85%   |
| Type Safety (Python/TS)   | 100%   |
| Linting Errors            | 0      |
| Security Vulnerabilities  | 0      |
| Documentation Coverage    | 95%    |
| Book Chapter Completeness | 100%   |

### Qualitative

- ✅ Marketplace is usable by teams (install/configure in <10 min)
- ✅ Each plugin README is clear and beginner-friendly
- ✅ ch08-ch10 are professionally edited (grammar, formatting, examples)
- ✅ Users can build blog platform following ch10 without errors
- ✅ Community feedback (GitHub issues, discussions) is positive

---

## 📋 Sign-Off Checklist

**Before v1.0.0 Release**:
- ✅ All plugins tested locally
- ✅ All commands & agents working
- ✅ All skills loading without errors
- ✅ All hooks triggering correctly
- ✅ MCP servers connecting (Figma, DevTools, Vercel, Supabase, Render)
- ✅ marketplace.json valid JSON + schema compliant
- ✅ All 5 plugins registered & versioned
- ✅ SECURITY.md & README.md complete
- ✅ ch08, ch09, ch10 chapters peer-reviewed
- ✅ Blog platform project runs end-to-end
- ✅ Git tags created (v1.0.0)
- ✅ GitHub Releases published
- ✅ Marketing announcement (optional: blog post, newsletter)

---

**Status**: Ready for Phase 1 execution
**Next Step**: Execute `/alfred:2-run SPEC-V1-001` to begin TDD implementation cycle
