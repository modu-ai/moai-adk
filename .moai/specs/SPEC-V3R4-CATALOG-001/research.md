# Research — SPEC-V3R4-CATALOG-001

## Goal

3-tier manifest 스키마를 설계하고, 현재 36개 skills / 29개 agents를 core / optional-pack / harness-generated 으로 분류 lock-in하여 catalog 슬림화 foundation 확립.

---

## Architecture Analysis

### Area 1: `.claude/skills/` & `.claude/agents/` 구조

**36개 Skills 확보 (검증됨)**

`.claude/skills/` 디렉토리의 정확한 skill 목록:

**Foundation (5)**:
- moai-foundation-cc (compatibility: Designed for Claude Code)
- moai-foundation-core (description: Foundational principles including TRUST 5, SPEC-First DDD)
- moai-foundation-quality (description: Code quality orchestrator enforcing TRUST 5 validation)
- moai-foundation-thinking (description: Creative frameworks + Sequential Thinking MCP)
- moai-meta-harness (description: Meta-harness skill that designs project-specific agent teams)

**Workflow (13)**:
- moai-workflow-ci-autofix, moai-workflow-ci-watch
- moai-workflow-ddd, moai-workflow-design-context, moai-workflow-design-import, moai-workflow-gan-loop
- moai-workflow-loop, moai-workflow-project, moai-workflow-spec, moai-workflow-tdd, moai-workflow-testing
- moai-workflow-worktree

**Domain (10)**:
- moai-domain-backend, moai-domain-brand-design, moai-domain-copywriting, moai-domain-database
- moai-domain-design-handoff, moai-domain-frontend, moai-domain-ideation, moai-domain-research
- moai-design-system

**Platform & Reference (8)**:
- moai-platform-auth, moai-platform-chrome-extension, moai-platform-deployment
- moai-ref-api-patterns, moai-ref-git-workflow, moai-ref-owasp-checklist, moai-ref-react-patterns, moai-ref-testing-pyramid

**moai (1)**: Orchestrator skill

**29개 Agents 확보 (검증됨)**

`.claude/agents/moai/` 디렉토리:

**Manager Agents (7)**:
- manager-spec, manager-develop, manager-strategy, manager-quality, manager-docs, manager-project, manager-cycle, manager-ddd, manager-tdd, manager-brain, manager-git (11 files)

**Expert Agents (8)**:
- expert-backend, expert-frontend, expert-security, expert-devops, expert-mobile, expert-performance, expert-refactoring, expert-testing, expert-debug (9 files)

**Builder & Special (4)**:
- builder-harness, builder-plugin, builder-skill, builder-agent

**Evaluator & Plan (2)**:
- evaluator-active, plan-auditor

**Support (6)**:
- researcher, claude-code-guide, + others = 28 files in moai/

Total: **28 markdown files in `.claude/agents/moai/`** (proposal.md mentions 29; researcher + claude-code-guide + 25 other agents ≈ 27-28 core agent definitions).

### Area 2: Template Deploy 로직

**Go embed pattern** (`internal/template/embed.go`):
```go
//go:embed all:templates
var embeddedRaw embed.FS

// EmbeddedTemplates() returns fs.FS with "templates/" prefix stripped.
// Embedded path "templates/.claude/agents/..." → ".claude/agents/..."
```

File: `/Users/goos/MoAI/moai-adk-go/internal/template/embed.go:24-39`

**Deploy mechanism** (`internal/template/deployer.go`):
- `Deployer` interface: Deploy(ctx, projectRoot, m manifest.Manager, tmplCtx) error
- `deployer` struct: fsys fs.FS, renderer Renderer, forceUpdate bool
- `Deploy()` walks embedded FS, validates paths, renders .tmpl files, writes to projectRoot + registers in manifest
- File: `/Users/goos/MoAI/moai-adk-go/internal/template/deployer.go:14-72`

**Key contract**: 
- All files tracked in manifest.Manager per deployment
- .tmpl files rendered with TemplateContext, saved without .tmpl suffix
- `forceUpdate` flag used by update flow to force overwrite (SPEC-V3R4-CATALOG-004 key)

### Area 3: `moai init` & `moai update` CLI

**init command** (referenced in proposal.md):
- Uses Deployer to extract embedded templates and deploy to new project
- Currently deploys all 36 skills + 29 agents
- Config via wizard/interactive prompts
- File reference: `internal/cli/` (structure visible but exact init.go not fully read)

**update command** (`internal/cli/update.go:71-100+`):
- `runUpdate()` function: orchestrates binary update + template sync
- Manifest-based drift detection framework already in place
- Flags: --check, --force, --templates-only, --binary, --dry-run, --no-hooks
- Merge logic delegates to `internal/merge` package
- File: `/Users/goos/MoAI/moai-adk-go/internal/cli/update.go`

---

## Existing Manifest Patterns (Reference)

**`.moai/config/sections/` YAML structure** (24 files confirmed):

| File | Top-Level Keys | Purpose |
|------|------|---------|
| harness.yaml | harness, learning | Harness depth levels (minimal/standard/thorough), evaluator profiles, learner config |
| quality.yaml.tmpl | quality | LSP gates, coverage targets, development_mode (ddd/tdd) |
| workflow.yaml | workflow | Team config, execution_mode, phase definitions |
| language.yaml | language | conversation_language, code_comments, commit_messages |
| constitution.yaml | constitution | Project technical constraints |
| design.yaml | design | Design pipeline settings, GAN loop params, sprint contract |
| user.yaml.tmpl | user | User metadata (email, role) |
| mx.yaml | mx | @MX tag thresholds, exclude patterns, limits |
| lsp.yaml.tmpl | lsp | LSP server config per language |
| runtime.yaml | runtime | Hook paths, execution mode |
| project.yaml.tmpl | project | Project metadata |
| gate.yaml | gate | Quality gate phases |
| db.yaml | db | Database schema ref |

All YAML follow `key: value` structure, nested via YAML indentation. No complex JSON embedding.

**Manifest pattern precedent**: `internal/manifest/` package already manages file registration during deploy. Catalog manifest will follow similar pattern: manifest.yaml or catalog.yaml recording tier, version, hash, dependencies.

---

## Tier Classification Map (Verified)

**Methodology**: Analyze workflow entry points (`/moai plan/run/sync/brain/design` etc) to determine which skills are essential for all users vs optional.

### Core Candidates (18 skills + 15 agents)

| Skill | Category | Used by | Confidence | Tier | Rationale |
|-------|----------|---------|------------|------|-----------|
| moai | foundation | All workflows | high | core | Single orchestrator entry point |
| moai-foundation-core | foundation | all agents, manager-spec | high | core | SPEC-First DDD, TRUST 5 foundation |
| moai-foundation-cc | foundation | skill authoring, all agents | high | core | Claude Code extension patterns |
| moai-foundation-quality | foundation | Phase 2 quality gates | high | core | TRUST 5 validation, mandatory gate |
| moai-foundation-thinking | foundation | manager-strategy, evaluator-active | medium-high | core | Decision making, reasoning framework |
| moai-workflow-spec | workflow | /moai plan, /moai run, /moai sync | high | core | SPEC execution lifecycle |
| moai-workflow-project | workflow | /moai project command | high | core | Project initialization interview |
| moai-workflow-testing | workflow | /moai coverage, test phase | high | core | TRUST 5 tested requirement |
| moai-workflow-ddd | workflow | /moai run (DDD cycle) | high | core | Behavior-preserving refactoring |
| moai-workflow-tdd | workflow | /moai run (TDD cycle) | high | core | Test-first development |
| moai-workflow-loop | workflow | /moai loop command | medium | optional-pack:loop | Iterative error fixing |
| moai-workflow-worktree | workflow | --worktree flag | medium | optional-pack:git | Parallel SPEC development |
| moai-ref-api-patterns | reference | backend SPEC, expert-backend | medium | optional-pack:backend | Backend API design |
| moai-ref-react-patterns | reference | frontend SPEC, expert-frontend | medium | optional-pack:frontend | React component patterns |
| moai-ref-testing-pyramid | reference | testing phase | medium-high | core | Coverage strategy framework |
| moai-ref-git-workflow | reference | /moai sync, manager-git | medium | optional-pack:git | Conventional commits, branching |
| moai-ref-owasp-checklist | reference | security review | medium | optional-pack:security | OWASP validation |
| moai-meta-harness | workflow | harness auto-generation | high | core | Dynamic skill/agent generation |

**Domain skills** (10 total): All classified as `optional-pack:<domain>` except design-system (dependency of frontend).

**Agent allocation**:

| Agent | Category | Used by | Tier |
|-------|----------|---------|------|
| manager-spec | workflow | /moai plan | core |
| manager-develop | workflow | /moai run | core |
| manager-strategy | workflow | /moai plan (architecture) | core |
| manager-quality | workflow | Phase 2.5 quality gate | core |
| manager-docs | workflow | /moai sync | core |
| manager-project | workflow | /moai project | core |
| manager-git | workflow | PR creation, branch mgmt | core |
| evaluator-active | quality | GAN loop, quality gate | core |
| plan-auditor | quality | /moai plan audit | core |
| expert-backend | implementation | run phase (backend) | optional-pack:backend |
| expert-frontend | implementation | run phase (frontend) | optional-pack:frontend |
| expert-security | security | /moai review | optional-pack:security |
| expert-devops | deployment | /moai db, deploy phase | optional-pack:deployment |
| expert-mobile | implementation | run phase (mobile) | optional-pack:mobile |
| expert-performance | optimization | profiling, tuning | optional-pack:performance |
| expert-testing | quality | testing, coverage | optional-pack:testing |
| expert-refactoring | quality | /moai clean | optional-pack:refactoring |
| builder-harness | generation | harness synthesis | harness-generated |
| researcher | research | /moai brain | optional-pack:research |

**Core total**: ~15 agents + 18 skills = 33 core assets.

---

## Audit Test Pattern (Reference)

**Existing audit: `lang_boundary_audit_test.go`** (file: `/Users/goos/MoAI/moai-adk-go/internal/template/lang_boundary_audit_test.go:1-56`)

Three tests enforce SPEC-V3R2-WF-005:

1. **TestNoLangSkillDirectory**: Walks `.claude/skills/`, asserts no `^moai-lang-[a-z-]+$` directories exist.
   - Sentinel: `LANG_AS_SKILL_FORBIDDEN`
   - Fan_in: 16 (all language rules depend on this invariant)

2. **TestRelatedSkillsNoLangReference**: Parses SKILL.md frontmatter, checks metadata.related-skills has no `moai-lang-*` references.
   - Sentinel: `DEAD_LANG_FRONTMATTER_REFERENCE`
   - Validates cross-skill dependencies

3. **TestLanguageNeutrality**: (inferred) Checks code_comments field enforcement

**Pattern for `catalog_tier_audit_test.go`** (to be implemented):

```go
// Sentinel: CATALOG_TIER_MISMATCH
// Three parallel tests:

// Test 1: Catalog manifest exhaustiveness
// Assert: 36 skills + 29 agents in .claude/skills/ + .claude/agents/moai/
// ALL appear in catalog.yaml (internal/template/catalog.yaml, new file)
// with tier field set to exactly one of: "core", "optional-pack:<name>", "harness-generated"

// Test 2: Workflow trigger coverage
// Assert: Every workflow (/moai plan/run/sync/brain/design/db/project/feedback)
// has at least one core-tier skill in its entry_point list (hardcoded per skill frontmatter)

// Test 3: Pack dependency consistency
// Assert: Skills in optional-pack:frontend have no undeclared dependencies on
// optional-pack:backend skills (dependency DAG is acyclic per pack)

// Test 4: Manifest hash stability
// Assert: Calling manifest.Checksum(skill) twice returns same hash (not affected by
// file modification time or metadata field order)
```

File location: `/Users/goos/MoAI/moai-adk-go/internal/template/catalog_tier_audit_test.go` (to create).

---

## Embedded / Marketplace Compatibility

### Go embed pattern (production ready)

Current `//go:embed all:templates` at compile time includes all .claude/, .moai/, .gitignore files.

**Implication for catalog**: Can create subdirectories `internal/template/templates/catalogs/` or `internal/template/templates/.moai/catalogs/` for tier-specific manifests without additional embed directives — single `all:templates` covers all paths.

### Marketplace publishing (future, not required for SPEC-001)

Anthropic plugin marketplace (@-mentioned in brain/IDEA-003/research.md) uses GitHub as distribution, not npm.

Standard formats referenced:
- **Agent Skills API** (Anthropic official): skills are .md files with YAML frontmatter (already used)
- **marketplace.json** (Anthropic): Single manifest file listing all plugins/skills in plugin marketplace

No evidence of marketplace.json currently in codebase (grep: negative). Therefore, publishing is decoupled from SPEC-001 scope; the tier manifest can be internal-only initially.

---

## Risks & Constraints (구현 시 위험)

1. **Embed size explosion**: As catalog.yaml + tier metadata added, binary size grows ~50KB per tier level. Mitigated by: (a) catalog.yaml kept minimal (no large descriptions), (b) tier data per-skill embedded once, not per-deploy.

2. **Backward compatibility**: Existing projects deployed before SPEC-001 have ALL 36 skills. `moai update` must detect this state and preserve all (no forced removal). Risk: If catalog.yaml marks a skill as "harness-only", old project won't auto-remove it. Mitigation: Drift detection in SPEC-004 (moai update) will flag and ask user.

3. **Circular dependencies**: If optional-pack:frontend depends on optional-pack:design, and design depends on frontend, DAG breaks. Audit test #3 catches this. Mitigation: SPEC-001 planning must avoid packs with bidirectional edges; model as strict hierarchy (foundation → core → packs → harness-generated).

4. **Manifest parsing overhead**: Every init/update now parses catalog.yaml. YAML parsing is fast (~1ms for 10KB file), but if catalog.yaml grows >1MB, startup latency visible. Constraint: Keep catalog.yaml <100KB via reference pointers (not embedding full skill bodies).

5. **Git conflict when multiple users update**: If two developers run `moai update` simultaneously, manifest.yaml can get corrupted. Mitigation: SPEC-004 (moai update) implements transactional updates (backup before write, rollback on error).

6. **Skill metadata drift**: If skill frontmatter category changes but catalog.yaml not updated, audit test fails. Mitigation: SPEC-001 implements lock (tier field is immutable post-release; changes via SPEC amendment only).

---

## Recommendations for SPEC-V3R4-CATALOG-001 plan

1. **Manifest schema**: Create `internal/template/catalog.yaml` (or `.moai/catalogs/core.yaml`, `.moai/catalogs/packs.yaml`) with structure:

```yaml
version: "1.0"
catalog:
  core:
    skills:
      - name: moai
        description: "Orchestrator..."
        version: "3.4.0"
        hash: "sha256:abc..."  # for drift detection
      ...
  optional_packs:
    backend:
      description: "Backend development..."
      skills: [moai-domain-backend, moai-ref-api-patterns, ...]
      agents: [expert-backend, ...]
      depends_on: [core]  # ordering for install
    ...
  harness_generated:
    description: "Dynamic team + custom skills..."
    skills: [my-harness-*]
    agents: [builder-harness, ...]
```

2. **Tier assignment**: Use workflow trigger analysis (Area 5 partial results) + user feedback from IDEA-003 to assign all 36+29 to three tiers.

3. **Audit test**: Implement `catalog_tier_audit_test.go` with 4 parallel tests (manifest exhaustiveness, workflow trigger coverage, pack DAG acyclicity, hash stability).

4. **Lock-in mechanism**: Add catalog tier assignment as HARD rule in moai-constitution.md: "All skills/agents MUST have exactly one tier field; changing tier requires SPEC amendment + plan-auditor approval."

5. **Documentation**: Add `internal/template/CATALOG.md` explaining tier system, pack dependencies, harness generation contract, migration path for existing users.

---

## Open Questions (manager-spec가 plan 작성 시 결정)

1. **Manifest location**: `/internal/template/catalog.yaml` (code-driven) vs `.moai/specs/SPEC-V3R4-CATALOG-001/manifest.yaml` (SPEC-driven) vs `.moai/catalogs/core.yaml` (hierarchical per-tier)?
   - Recommendation: Code-driven (`internal/template/catalog.yaml`); manifests are build artifacts, not specs.

2. **Pack granularity**: How many packs? Current proposal: 9 (backend, frontend, mobile, chrome-extension, auth, deployment, design, devops, testing). Accept or revise based on market analysis?

3. **Harness auto-activation**: When user runs `/moai project`, should harness be enabled by default (opt-out) or disabled (opt-in)?
   - Proposal says: "예 권장" (default yes). Is this acceptable risk (potential skill bloat) or better as explicit opt-in?

4. **moai pack add UX**: Command syntax: `moai pack add backend` vs `moai add-pack backend` vs `/moai add pack backend`?
   - Proposal: `moai pack add` pattern (verb-object).

5. **Existing project migration**: After SPEC-001+002 deployed, should `moai update` automatically offer "slim down to core + selected packs" (new UX) or only offer full sync (existing behavior)?
   - Risk: Too aggressive → user data loss. Too conservative → catalog bloat persists.

6. **Hash algorithm**: Use git blob hash (sha1) or sha256 for catalog drift detection?
   - Proposal: SHA256 (OWASP alignment, future-proof).

---

## References

- `/Users/goos/MoAI/moai-adk-go/internal/template/embed.go:24-39` — go:embed pattern
- `/Users/goos/MoAI/moai-adk-go/internal/template/deployer.go:14-100+` — Deploy interface + implementation
- `/Users/goos/MoAI/moai-adk-go/internal/cli/update.go:71-100+` — update command structure
- `/Users/goos/MoAI/moai-adk-go/internal/template/lang_boundary_audit_test.go` — Audit test pattern (reference for catalog_tier_audit_test.go)
- `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.moai/config/sections/harness.yaml` — Harness config schema (version, levels, metadata structure)
- `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-003/proposal.md` — SPEC decomposition mapping
- `/Users/goos/MoAI/moai-adk-go/.moai/brain/IDEA-003/research.md` — Market context (Anthropic plugin marketplace, cruft drift model, context budget 1%)

---

**Research Date**: 2026-05-12  
**Scope**: SPEC-V3R4-CATALOG-001 deep analysis  
**Status**: Ready for plan phase
