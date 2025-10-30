# SPEC-V1-001 Development Plan

## 📊 Work Breakdown Structure (WBS)

### Phase 1: Foundation & Setup (Week 1)
**Goal**: Establish v1.0 development environment and create plugin skeleton

**Tasks**:
1. **P1-1**: Git worktree setup for v1.0 (release/v1.0 branch)
   - Create release/v1.0 branch from develop
   - Set up MoAI-ADK-v1.0 worktree directory
   - Initialize Python 3.14 environment (uv)
   - Status: ✅ COMPLETED

2. **P1-2**: SPEC documentation & planning
   - Write SPEC-V1-001 (spec.md, plan.md, acceptance.md)
   - Create 5-plugin architecture diagram
   - Define version lock matrix (Node/Python/deps)
   - Status: IN PROGRESS

3. **P1-3**: Marketplace repository setup (moai-adk/moai-cc-marketplace)
   - Create GitHub repository
   - Initialize directory structure
   - Write SECURITY.md (org policies)
   - Create README.md (installation guide)
   - Status: PENDING

4. **P1-4**: Plugin skeleton generation
   - Generate 5 plugin directory structures
   - Create plugin.json templates
   - Add hooks.json template
   - Status: PENDING

---

### Phase 2: Frontend Stack (Weeks 2-3)
**Goal**: Deliver PM Plugin + UI/UX Plugin + Frontend Plugin (Next.js 16)

**P2-1: PM Plugin (Project Management)**
- Deliverables:
  - `/init-pm [project-name]` command
  - EARS SPEC template generator (spec.md)
  - Project charter template (charter.md)
  - Risk assessment matrix (risk-matrix.md)
  - SPEC validation skill (SKILL-PM-SPEC-VALIDATOR.md)
- Tests: pytest with >85% coverage
- Status: PENDING

**P2-2: UI/UX Plugin (Design System)**
- Deliverables:
  - `/setup-shadcn-ui` command (Tailwind + shadcn/ui)
  - Design token configuration (tailwind.config.ts)
  - Component inventory skill
  - Figma MCP integration (.mcp.json)
  - Design guidelines (design-system.md)
- Tests: shadcn/ui component rendering tests
- Status: PENDING

**P2-3: Frontend Plugin (Next.js 16 + React 19.2)**
- Deliverables:
  - `/init-next [app-name] [pm=bun|npm|pnpm]` command
  - `/biome-setup` command (formatting + linting)
  - `/connect-mcp-devtools` command (DevTools MCP)
  - Next.js 16 app scaffolding (App Router)
  - shadcn/ui integration
  - Biome configuration (.biomerc.json)
  - Pre-commit hooks
- Tests: Next.js build tests, component snapshot tests
- Status: PENDING

---

### Phase 3: Backend Stack (Week 4)
**Goal**: Deliver Backend Plugin (FastAPI + uv) + DevOps Plugin

**P3-1: Backend Plugin (FastAPI + uv)**
- Deliverables:
  - `/init-fastapi [app-name]` command
  - `/db-setup [driver=postgres|mysql]` command
  - `/resource-crud [resource-name]` command (generates models/schemas/routes from SPEC)
  - `/run-dev` command (uvicorn with auto-reload + .env)
  - `/api-profile` command (performance metrics)
  - FastAPI 0.120.2 scaffolding
  - SQLAlchemy 2.0.44 model templates
  - Alembic 1.17 migration setup
  - Pydantic 2.12 schema validators
  - uv configuration (custom index/registry support)
  - pytest fixtures and example tests
- Tests: pytest with >85% coverage (models, schemas, routes, migrations)
- Status: PENDING

**P3-2: DevOps Plugin (Vercel/Supabase/Render)**
- Deliverables:
  - `/deploy-config [platform=vercel|render]` command
  - `/connect-vercel` OAuth command
  - `/connect-supabase` OAuth command
  - `/generate-github-actions` command (CI/CD workflows)
  - `/secrets-setup` command (OS Keychain registration)
  - vercel.json template
  - .github/workflows/ (test, build, deploy)
  - Supabase project setup guide
  - Deployment troubleshooting guide
- Tests: YAML schema validation, workflow syntax checks
- Status: PENDING

---

### Phase 4: Integration & Documentation (Week 5)
**Goal**: Finalize plugins, write chapters 8-10, publish marketplace

**P4-1: Plugin Integration**
- Marketplace creation: moai-adk/moai-cc-marketplace
- Plugin publication: Add 5 plugins to marketplace.json
- CI/CD setup: GitHub Actions for plugin validation
- Version locking: Frozen deps for all 5 plugins
- Status: PENDING

**P4-2: Book Chapter Development**
- **ch08**: Claude Code Plugins & Migration (Output Styles → Plugins/Hooks/Skills)
  - Why plugins ecosystem
  - Plugin architecture & components
  - plugin.json schema breakdown
  - Slash command authoring guide
  - Hook event system tutorial
  - Skill auto-loading & context invocation
  - MCP server integration patterns
  - Marketplace publishing workflow
  - Migration checklist (from Output Styles)
- **ch09**: 5-Plugin Development Workflow
  - 9-1: PM Plugin walkthrough
  - 9-2: UI/UX Plugin (shadcn/ui)
  - 9-3: Frontend Plugin (Next.js 16)
  - 9-4: Backend Plugin (FastAPI + uv)
  - 9-5: DevOps Plugin (Vercel/Supabase/Render)
  - Each section: architecture, commands, config, FAQ
- **ch10**: Full Project - Personal Blog Platform
  - 10-1: Requirements & Design (EARS SPEC, data model)
  - 10-2: Frontend (Next.js, shadcn/ui, Markdown editor)
  - 10-3: Backend (FastAPI, RBAC, migrations)
  - 10-4: Deployment (3 platforms, CI/CD, monitoring)
- Status: PENDING

---

### Phase 5: Release & QA (Week 6)
**Goal**: Final testing, documentation sync, v1.0.0 release

**P5-1: Testing & Quality Gates**
- Pytest coverage: ≥85% for each plugin
- Linting: 0 errors (ruff, biome)
- Type checking: 100% type hints (mypy strict)
- Security scan: No hardcoded secrets
- Documentation: All commands/agents/skills documented
- Status: PENDING

**P5-2: Release Preparation**
- CHANGELOG update (v1.0.0 final)
- Documentation sync (.moai/docs/, .moai/specs/)
- Release notes (features, breaking changes, migration guide)
- Git tag: v1.0.0
- GitHub Releases: Publish to main branch
- Status: PENDING

**P5-3: Post-Release**
- Marketplace marketplace.json publication
- Plugin availability announcement
- Book chapters ch08-ch10 publication
- Community feedback collection
- v1.1.0-dev roadmap initialization
- Status: PENDING

---

## 🔄 Dependency Map

```
P1-3 (Marketplace setup) ← P1-1 (Worktree)
         ↓
P2-1 (PM Plugin) ← P1-2 (SPEC)
P2-2 (UI/UX Plugin) ← P1-2 (SPEC)
P2-3 (Frontend Plugin) ← P1-2 (SPEC) + P2-2
         ↓
P3-1 (Backend Plugin) ← P1-2 (SPEC)
P3-2 (DevOps Plugin) ← P1-2 (SPEC)
         ↓
P4-1 (Integration) ← P2-*, P3-*
P4-2 (Book ch08-ch10) ← P2-*, P3-*, P4-1
         ↓
P5-* (Release & QA) ← P4-*
```

---

## 📦 Deliverables Checklist

### By Plugin

**PM Plugin**:
- [ ] `/init-pm` command
- [ ] SPEC template generator
- [ ] PM-Agent (sub-agent)
- [ ] SKILL-PM-*.md files
- [ ] README.md + USAGE.md
- [ ] pytest tests (>85% coverage)

**UI/UX Plugin**:
- [ ] `/setup-shadcn-ui` command
- [ ] Tailwind config template
- [ ] Design token skill
- [ ] Figma MCP integration
- [ ] Component inventory documentation
- [ ] README.md + USAGE.md
- [ ] pytest tests

**Frontend Plugin**:
- [ ] `/init-next [app-name] [pm]` command
- [ ] `/biome-setup` command
- [ ] `/connect-mcp-devtools` command
- [ ] `/routes-diagnose` command
- [ ] Next.js 16 scaffold
- [ ] Biome config (.biomerc.json)
- [ ] Pre-commit hooks
- [ ] README.md + USAGE.md
- [ ] Build & component tests

**Backend Plugin**:
- [ ] `/init-fastapi [app-name]` command
- [ ] `/db-setup [driver]` command
- [ ] `/resource-crud [resource]` command
- [ ] `/run-dev` command
- [ ] `/api-profile` command
- [ ] FastAPI scaffold
- [ ] Alembic migration setup
- [ ] uv configuration (custom index)
- [ ] pytest tests (>85% coverage)
- [ ] README.md + USAGE.md

**DevOps Plugin**:
- [ ] `/deploy-config [platform]` command
- [ ] `/connect-vercel` OAuth command
- [ ] `/connect-supabase` OAuth command
- [ ] `/generate-github-actions` command
- [ ] `/secrets-setup` command
- [ ] vercel.json template
- [ ] .github/workflows/ templates
- [ ] Deployment guide documentation
- [ ] README.md + USAGE.md
- [ ] YAML schema validation tests

### Marketplace

- [ ] moai-adk/moai-cc-marketplace repository
- [ ] .claude-plugin/marketplace.json (all 5 plugins)
- [ ] SECURITY.md (org policies)
- [ ] README.md (installation guide)
- [ ] CHANGELOG.md (version history)
- [ ] CI/CD pipeline (.github/workflows/)

### Book Chapters

- [ ] ch08: Claude Code Plugins Introduction & Migration
  - [ ] Plugin ecosystem why/what/how
  - [ ] Components breakdown (commands, agents, skills, hooks, MCP)
  - [ ] plugin.json schema reference
  - [ ] Hands-on: Write custom command
  - [ ] Hands-on: Create plugin skill
  - [ ] Hands-on: Connect MCP server
  - [ ] Migration guide (Output Styles → plugins)

- [ ] ch09: 5-Plugin Development Workflow
  - [ ] 9-1: PM Plugin
  - [ ] 9-2: UI/UX Plugin
  - [ ] 9-3: Frontend Plugin (Next.js 16)
  - [ ] 9-4: Backend Plugin (FastAPI)
  - [ ] 9-5: DevOps Plugin

- [ ] ch10: Full Project - Personal Blog Platform
  - [ ] 10-1: Requirements & Design
  - [ ] 10-2: Frontend Implementation
  - [ ] 10-3: Backend Implementation
  - [ ] 10-4: Deployment & Operations

---

## 🎯 Daily Standup Template

**Format**: For each task, track:
- Status (To Do / In Progress / Blocked / Done)
- % Complete
- Blocker (if any)
- Next action

**Example**:
```
P2-3: Frontend Plugin
- Status: In Progress
- % Complete: 35% (init-next command done, biome-setup in progress)
- Blocker: Need Biome v2.x release confirmation
- Next: Finish biome-setup, start connect-mcp-devtools
```

---

## 📅 Timeline Visualization

```
Week 1 (Setup)        Week 2-3 (Frontend)    Week 4 (Backend)       Week 5-6 (Integration)
[P1-1 P1-2 P1-3 P1-4]  [P2-1 P2-2 P2-3]       [P3-1 P3-2]            [P4-1 P4-2 P5-*]
  ↓                       ↓                      ↓                       ↓
Worktree ready         Frontend ready       Backend ready          Release v1.0.0
```

---

## 🚀 Ready to Start?

To begin development:
1. You are now in the v1.0 worktree: `/Users/goos/MoAI/MoAI-ADK-v1.0`
2. Start with Phase 1, Task P1-3: Create moai-cc-marketplace repository
3. Then proceed to P2-1: PM Plugin implementation
4. Use `/alfred:2-run SPEC-V1-001` to track progress with TDD
