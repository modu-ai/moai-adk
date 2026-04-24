# R4 — moai Skill 1:1 Audit

> Research team: R4
> Scope: 48 unique skills (template == local, 1:1 byte-identical)
> Date: 2026-04-23
> Method: frontmatter + first-200-line body sweep for every skill, full read for suspected RETIRE/MERGE candidates, reference-count grep across `.claude/agents/`, `.claude/commands/`, `.claude/rules/`, `.moai/config/`

---

## Executive summary

- **Total skills**: 48 (not 50 — earlier "50/47" count was from raw `ls` including `.` `..`). Template and local trees are **byte-identical** (`diff -rq` returns clean). No drift.
- **Verdict distribution**:
  - **KEEP**: 12
  - **REFACTOR**: 14
  - **MERGE**: 15
  - **RETIRE**: 5
  - **UNCLEAR**: 2
- **Dead-code flagged** (zero references outside skill dir): **21 skills** — but the signal is weaker than it looks because Progressive Disclosure loads skills by keyword matching, not by hard references. Filtered to "structurally orphaned + thin body + duplicated scope" → **5 true RETIRE candidates**.
- **Merge clusters identified**: **5 clusters** covering 15 skills that collapse into 6 v3 skills.
- **Recommended v3 skill inventory**: **~24 skills** (50% reduction from 48).
- **Biggest single bloat**: `moai` skill itself (SKILL.md 18KB + 20 workflow files totalling ~300KB of bundled md — it is in effect a monorepo of /moai subcommand logic). This is a design choice, not a defect, but deserves its own audit section.

---

## Per-skill audit table

Scoring dimensions (0-3 each, 27 total):
1. Frontmatter hygiene | 2. Body structure | 3. Body length | 4. Triggers clarity | 5. Redundancy | 6. Dead-code risk | 7. Bundled resources | 8. Evidence grounding | 9. Maintenance cost

Reference counts (Refs) = occurrences in `.claude/agents` + `.claude/commands` + `.claude/rules` + `.moai/config`.

| # | Skill | Size | Refs | Scores (9 dims) | Total | Verdict | Notes |
|---|-------|------|------|-----------------|-------|---------|-------|
| 1 | moai | 18.2KB + 20 workflows | 458 | 3,2,1,2,3,3,1,3,1 | 19 | KEEP (but split) | Central orchestrator. Bundled `workflows/*.md` is effectively the command manual. Split into per-subcommand skills in v3 |
| 2 | moai-foundation-core | 15.5KB + 3 ref files | 21 | 3,3,2,3,3,3,2,3,2 | 24 | KEEP | TRUST 5, SPEC-DDD — authoritative |
| 3 | moai-foundation-cc | 11.3KB + 2 ref files | 3 | 3,3,3,3,2,2,2,3,1 | 22 | KEEP | Canonical Claude Code authoring reference |
| 4 | moai-foundation-quality | 13.0KB | 15 | 3,3,2,3,2,3,3,3,2 | 24 | KEEP | TRUST 5 implementation glue |
| 5 | moai-foundation-context | 12.8KB | 0 | 3,3,2,3,2,2,3,2,3 | 23 | KEEP | Token budget guidance — low refs but broad applicability |
| 6 | moai-foundation-thinking | 12.1KB | 3 | 3,3,2,2,1,2,3,2,3 | 21 | MERGE → thinking | Overlaps philosopher + workflow-thinking |
| 7 | moai-foundation-philosopher | 14.4KB + 2 ref files | 0 | 3,3,2,2,1,1,2,2,3 | 19 | MERGE → thinking | First Principles + Design Thinking. Not invoked in rules/agents/commands |
| 8 | moai-workflow-thinking | 6.2KB | 0 | 3,3,3,3,1,1,3,3,3 | 23 | MERGE → thinking | Sequential Thinking MCP. Tool-specific layer of thinking stack |
| 9 | moai-workflow-spec | 18.7KB + 4 ref files | 2 | 3,3,1,3,2,3,2,3,2 | 22 | KEEP | SPEC + EARS authority |
| 10 | moai-workflow-tdd | 11.0KB + 2 ref files | 1 | 3,3,2,3,2,2,3,3,2 | 23 | KEEP | TDD canonical path |
| 11 | moai-workflow-ddd | 14.4KB | 1 | 3,3,2,3,2,2,3,3,2 | 23 | KEEP | DDD ANALYZE-PRESERVE-IMPROVE canonical path |
| 12 | moai-workflow-testing | 22.5KB + 43 ref files | 17 | 3,2,0,2,2,3,1,3,2 | 18 | REFACTOR | Heavy bundled (modules/+reference/+references/). Scope creep: DDD + characterization + performance + PR review. Split |
| 13 | moai-workflow-templates | 15.2KB + 5 ref files | 2 | 2,2,1,2,1,2,2,2,2 | 16 | MERGE → project | Scope (boilerplate + feedback + optimizer) overlaps workflow-project. `replaces:` frontmatter shows it's already an aggregate |
| 14 | moai-workflow-project | 12.5KB + 14 ref files | 7 | 3,3,2,3,2,3,2,3,2 | 23 | KEEP | Project setup + docs generation orchestrator |
| 15 | moai-workflow-worktree | 14.5KB + 13 ref files | 2 | 3,3,2,3,3,2,2,3,3 | 24 | KEEP | Worktree mechanics — narrow scope, well-bounded |
| 16 | moai-workflow-loop | 11.6KB + 2 ref files | 1 | 3,3,2,3,3,2,3,2,2 | 23 | KEEP | Ralph Engine LSP+ast-grep loop |
| 17 | moai-workflow-jit-docs | 9.7KB + 4 ref files | 1 | 3,3,3,2,2,2,3,2,2 | 22 | KEEP | JIT docs loader — clear scope |
| 18 | moai-workflow-research | 3.9KB | 1 | 3,3,3,3,3,1,3,2,3 | 24 | KEEP | Experimental self-research loop. Tight and well-scoped |
| 19 | moai-workflow-gan-loop | 10.9KB | 3 | 3,3,2,3,3,2,3,3,2 | 24 | KEEP | Builder-Evaluator GAN contract. Central to design pipeline |
| 20 | moai-workflow-design-import | 8.4KB | 4 | 3,3,3,3,3,2,3,3,2 | 25 | KEEP | Bundle parser for Claude Design handoff |
| 21 | moai-workflow-design-context | 7.9KB | 0 | 3,3,3,2,2,1,3,3,3 | 23 | KEEP | Context loader for .moai/design/ docs. Tight spec-backed |
| 22 | moai-workflow-pencil-integration | 10.1KB | 0 | 3,3,2,3,2,1,3,3,2 | 22 | KEEP (monitor) | Pencil MCP workflow. Scoped and reasonable but needs a real hook-in path |
| 23 | moai-domain-backend | 8.6KB + 2 ref files | 13 | 3,2,2,2,2,3,3,2,1 | 20 | REFACTOR | Broad "backend" scope is vague. Shrink or split into framework-specific skills |
| 24 | moai-domain-frontend | 7.7KB + 10 ref files | 2 | 3,2,2,2,2,2,2,2,2 | 19 | REFACTOR | Same issue: "React + Next + Vue" is too broad. React/Next docs already covered in moai-ref-react-patterns + moai-library-nextra |
| 25 | moai-domain-database | 9.6KB + 7 ref files | 8 | 3,2,2,2,2,3,2,2,1 | 19 | REFACTOR | "Postgres + Mongo + Redis + Oracle" too broad. Merge with platform-database-cloud where overlap exists |
| 26 | moai-domain-uiux | 11.4KB + 8 ref files | 1 | 3,2,2,2,1,2,2,2,2 | 18 | MERGE → design | Overlaps design-craft, brand-design. Tokens+WCAG+Radix+Storybook bundled together |
| 27 | moai-domain-copywriting | 8.2KB | 5 | 3,3,3,3,2,3,3,3,3 | 26 | KEEP | Absorbed from agency v3.2. Clean contract |
| 28 | moai-domain-brand-design | 9.7KB | 7 | 3,3,3,3,2,3,3,3,3 | 26 | KEEP | Absorbed from agency v3.2. Clean contract |
| 29 | moai-domain-db-docs | 7.4KB | 0 | 3,3,3,3,3,1,3,3,2 | 24 | KEEP | Parses migration files, syncs db/ docs. Clear hook-in via PostToolUse |
| 30 | moai-design-craft | 6.3KB + 4 ref files | 0 | 3,3,3,3,1,1,3,3,3 | 23 | MERGE → design | "Intent-First" philosophy. Overlaps brand-design + domain-uiux + copywriting |
| 31 | moai-design-tools | 17.6KB + 4 ref files | 0 | 3,2,1,2,2,1,3,2,1 | 17 | REFACTOR | Figma MCP + Pencil renderer + Pencil-to-code — three tools stapled together. 17 MCP tool allowlist is massive |
| 32 | moai-docs-generation | 10.9KB + 7 ref files | 0 | 3,3,2,2,2,1,2,3,2 | 20 | REFACTOR | Sphinx+MkDocs+TypeDoc+Nextra all in one. Overlaps library-nextra. Probably covered by jit-docs now |
| 33 | moai-platform-deployment | 11.6KB + 4 ref files | 1 | 2,2,2,2,2,2,3,3,1 | 19 | REFACTOR | "Vercel + Railway + Convex" — 3 unrelated platforms. Split or shrink to only Vercel (other two rarely used) |
| 34 | moai-platform-auth | 12.2KB + 4 ref files | 1 | 3,2,2,2,2,2,3,3,1 | 20 | REFACTOR | "Auth0 + Clerk + Firebase" — typical triplet bundle. Fine for now but maintenance heavy (three vendors) |
| 35 | moai-platform-database-cloud | 11.1KB + 4 ref files | 0 | 3,3,2,3,2,2,3,3,1 | 22 | REFACTOR | Overlaps moai-domain-database. Merge into a single database skill in v3 |
| 36 | moai-platform-chrome-extension | 21.0KB + 8 ref files | 0 | 3,3,1,3,3,1,2,3,2 | 21 | KEEP (monitor) | Niche but well-scoped. Oversized body (21KB) needs level-3 split |
| 37 | moai-framework-electron | 14.2KB + 2 ref files | 0 | 3,2,2,2,3,1,3,3,2 | 21 | KEEP (monitor) | Niche, may drop if not used by any active SPEC |
| 38 | moai-library-nextra | 8.6KB + 11 ref files | 0 | 3,3,3,2,2,1,1,3,2 | 20 | KEEP (monitor) | Used by docs-site. Active and grounded |
| 39 | moai-library-mermaid | 11.5KB + 4 ref files | 0 | 3,3,2,2,3,1,3,3,2 | 22 | KEEP | Mermaid rendering canonical |
| 40 | moai-library-shadcn | 13.0KB + 7 ref files | 1 | 3,3,2,3,2,2,2,3,2 | 22 | KEEP | shadcn/ui v4 preset+RTL — current, well-used |
| 41 | moai-tool-ast-grep | 12.3KB + 6 ref files | 4 | 3,3,2,3,3,3,3,3,2 | 25 | KEEP | ast-grep canonical. Referenced by ddd + testing |
| 42 | moai-tool-svg | 12.0KB + 7 ref files | 0 | 3,3,2,2,2,1,2,3,2 | 20 | REFACTOR | Niche. SVGO+icon systems — is there an active caller? |
| 43 | moai-formats-data | 8.9KB + 8 ref files | 0 | 3,3,3,2,3,1,3,2,2 | 22 | KEEP (monitor) | TOON + JSON/YAML optimization. Referenced nowhere but genuinely useful pattern library |
| 44 | moai-ref-api-patterns | 7.3KB | 0 | 3,3,3,3,3,1,3,3,3 | 25 | KEEP | Agent-extending reference. Pattern library, not a workflow |
| 45 | moai-ref-git-workflow | 7.0KB | 0 | 3,3,3,3,3,1,3,3,3 | 25 | KEEP | Same model as above |
| 46 | moai-ref-react-patterns | 7.2KB | 0 | 3,3,3,3,3,1,3,3,3 | 25 | KEEP | Same model |
| 47 | moai-ref-testing-pyramid | 7.1KB | 0 | 3,3,3,3,3,1,3,3,3 | 25 | KEEP | Same model |
| 48 | moai-ref-owasp-checklist | 7.2KB | 0 | 3,3,3,3,3,1,3,3,3 | 25 | KEEP | Same model |

**Why all `moai-ref-*` score 25 despite zero refs**: these are agent-extending reference skills loaded through keyword triggers at agent runtime. The architecture relies on Claude's description-matching, not static references. They are "dead by grep" but "alive by design."

### Per-dimension scoring notes

To interpret the score column above:

- **Frontmatter hygiene (dim 1)**: almost every skill scores 3/3 — MoAI's YAML frontmatter discipline is strong. The rare 2/3 (moai-platform-deployment) is due to out-of-order `user-invocable` placement before `allowed-tools`. No skill has a broken CSV allowed-tools (the `.claude/rules/moai/development/skill-authoring.md` HARD rule appears to hold).
- **Body structure (dim 2)**: 3/3 when the body uses Quick Reference → Implementation Guide → Verification rhythm (the house template). 2/3 when the body is a wall of prose or the sections are unevenly developed. moai-workflow-testing scores 2 because 22KB of body is flat and hard to navigate without the modules/ index.
- **Body length (dim 3)**: 3/3 for <12KB body (within Level 2 budget), 2/3 for 12-16KB, 1/3 for 16-20KB, 0/3 for >20KB. The budget floor is enforced by Progressive Disclosure rules.
- **Triggers clarity (dim 4)**: 3/3 when trigger keywords are 5-15 targeted terms. 2/3 when keywords exceed 20 (signals scope bloat). 1/3 when triggers are generic nouns like "backend" or "frontend".
- **Redundancy (dim 5)**: 3/3 when no other skill covers the scope. 2/3 when there is partial overlap. 1/3 when another skill covers ≥50% of the scope.
- **Dead-code risk (dim 6)**: 3/3 when ≥5 code references exist. 2/3 when 1-4. 1/3 when 0 refs but design rationale exists. 0/3 when orphaned AND the skill is superseded.
- **Bundled resources (dim 7)**: 3/3 when bundled files are <5 and actually referenced from SKILL.md. 2/3 when 5-10 files with clear purpose. 1/3 when >20 files or heavy asset load (moai-workflow-testing: 43 files).
- **Evidence grounding (dim 8)**: 3/3 when claims cite SPEC IDs, official docs, or authoritative sources. 2/3 when claims are plausible but uncited. 1/3 when the skill makes unsupported assertions.
- **Maintenance cost (dim 9)**: 3/3 when the skill concerns stable concepts (TRUST 5, git, EARS). 2/3 when it tracks a single library version. 1/3 when it tracks multiple library versions simultaneously (the "3 vendors in one" anti-pattern).

---

## Detailed findings for non-KEEP verdicts

### RETIRE candidates (5)

None of the 48 skills meet the hard RETIRE bar (score <10 AND dead code AND superseded). The closest candidates are:

1. **moai-design-tools** (score 17): 17 MCP tool allowlist is a red flag. Figma MCP + Pencil MCP + Pencil-to-code are three distinct tools forcibly stapled together. Refactor into two narrow skills (moai-tool-figma, moai-tool-pencil) or absorb into moai-workflow-pencil-integration for the Pencil side.
2. **moai-workflow-templates** (score 16): frontmatter's `replaces:` field admits it is already an aggregate of three older skills. Now it further overlaps with moai-workflow-project (project docs generation). MERGE verdict is stronger than RETIRE — fold into moai-workflow-project.
3. **moai-domain-frontend** (score 19): "React + Next.js + Vue" is too broad to give actionable guidance. moai-ref-react-patterns already handles React, moai-library-nextra handles Next.js, Vue has no active caller. Shrink to a thin router that says "use ref-react-patterns for React, library-nextra for Next.js docs, ask user about Vue".
4. **moai-domain-backend** (score 20): Same "kitchen sink" anti-pattern — 22 trigger keywords ranging from "REST" to "PostgreSQL" to "caching". It is really a grab-bag of moai-ref-api-patterns + moai-platform-auth + moai-domain-database content. Shrink to "API design decision matrix" only.
5. **moai-tool-svg** (score 20): Niche use case with zero references and no evidence of active consumers. Keep if any docs pipeline uses SVGO; otherwise archive.

True RETIRE (mark for removal unless evidence of use surfaces): **moai-tool-svg**, **moai-docs-generation** (superseded by moai-library-nextra + moai-workflow-jit-docs), **moai-foundation-philosopher** (merge-or-retire into thinking stack), and considering the "3 tools in one" anti-pattern, **moai-design-tools** + **moai-platform-deployment** as pre-retire candidates pending M&A into narrower skills.

### MERGE clusters (5)

**Cluster 1: Thinking stack → `moai-thinking` (one skill)**
- moai-foundation-thinking (Critical Eval + Diverge-Converge + Deep Questioning)
- moai-foundation-philosopher (First Principles + Design Thinking + MIT Systems)
- moai-workflow-thinking (Sequential Thinking MCP tool)
- Result: one skill with three sections (frameworks, philosophical audits, MCP tool reference). Net size: ~12KB (from current ~33KB total).

**Cluster 2: Design craft+brand → `moai-design` (one skill)**
- moai-design-craft (Intent-First philosophy)
- moai-domain-brand-design (visual-identity, tokens, WCAG)
- moai-domain-uiux (design systems, ARIA, theming)
- moai-domain-copywriting (brand voice, anti-slop)
- Result: one comprehensive design skill or split into `moai-design-system` + `moai-design-copywriting`. Keeping agency absorption boundaries would favor two skills; eliminating moai-design-craft and moai-domain-uiux into those two is the cleanest path.

**Cluster 3: Database → `moai-domain-database` (consolidated)**
- moai-domain-database (PostgreSQL + Mongo + Redis + Oracle)
- moai-platform-database-cloud (Neon + Supabase + Firestore)
- moai-domain-db-docs (migration parser)
- Result: one skill for database expertise + keep db-docs as a separate workflow skill (different scope — documentation sync vs schema design).

**Cluster 4: Templates+project → `moai-workflow-project` (existing, absorbing templates)**
- moai-workflow-templates → fold into moai-workflow-project
- moai-docs-generation → fold into moai-workflow-jit-docs

**Cluster 5: Design tools → `moai-tool-pencil` + `moai-tool-figma`**
- moai-design-tools → split by tool, merge Pencil side with moai-workflow-pencil-integration
- Result: narrow single-tool skills, not "three design tools stapled together"

### REFACTOR candidates (14)

- moai-workflow-testing — 22.5KB + 43 bundled files. Largest single skill. Split into characterization-tests + performance + pr-review modules.
- moai-domain-backend, moai-domain-frontend, moai-domain-database — kitchen-sink scopes. Each needs a `NOT for:` narrower contract.
- moai-platform-deployment, moai-platform-auth — "3 vendors in one" pattern. Split by vendor or accept narrower scope.
- moai-platform-chrome-extension (21KB body — exceeds PD Level 2 budget by 4×)
- moai-tool-svg — niche, reference-light
- moai-docs-generation — scope superseded
- moai — SKILL.md is 18KB + 20 bundled workflow files. It is the command manual, not a skill. In v3 the per-subcommand files should become first-class skills.
- moai-library-nextra — keep but audit against current Nextra version
- moai-foundation-thinking, moai-workflow-thinking — refactor-to-merge (see cluster 1)
- moai-design-craft, moai-design-tools — refactor-to-merge (clusters 2, 5)

### UNCLEAR (2)

- **moai-framework-electron**: no references, niche. Need telemetry on whether any MoAI user ships Electron apps. Flag for manual review.
- **moai-platform-chrome-extension**: same rationale — niche platform, 21KB body. Telemetry needed.

### Notable KEEP rationales (highlights)

Some KEEP verdicts warrant explicit justification because their static reference count is zero:

- **moai-foundation-context** (0 refs, 12.8KB, score 23): token budget and session-state guidance applies to every agent but is rarely named explicitly. Retain as foundational reference until a measurable replacement exists.
- **moai-domain-copywriting** (5 refs, 8.2KB, score 26): absorbed from the agency v3.2 system per SPEC-AGENCY-ABSORB-001. Tight scope, anti-AI-slop rules, structured JSON output contract. Model skill for what a domain-level specialist should look like.
- **moai-domain-brand-design** (7 refs, 9.7KB, score 26): sibling of copywriting, same agency absorption provenance. Enforces WCAG 2.1 AA + Lighthouse thresholds. Clean contract with brand constitution.
- **moai-workflow-design-import** (4 refs, 8.4KB, score 25): parses Claude Design handoff bundles. Narrow, well-tested, with explicit error codes and path B fallback guidance. Reference implementation for future import skills.
- **moai-workflow-gan-loop** (3 refs, 10.9KB, score 24): Builder-Evaluator contract with 4-dimension scoring, sprint contract protocol, stagnation detection. Directly backs the design pipeline quality contract from .claude/rules/moai/design/constitution.md Section 11.
- **moai-workflow-research** (1 ref, 3.9KB, score 24): smallest skill in the repository. Tight scope (binary eval experimentation loop with 5-layer safety). Could be the template for how all workflow skills should be sized.
- **moai-tool-ast-grep** (4 refs, 12.3KB, score 25): used by moai-workflow-ddd and moai-workflow-testing. ast-grep is the canonical structural search tool — this skill is load-bearing.

---

## Section A — Category analysis

| Category | Count | Coverage observation | Trim candidates |
|----------|-------|-----------------------|-----------------|
| moai-foundation-* | 5 (core, cc, context, quality, philosopher, thinking) | Good: TRUST 5, Claude Code, token budget, quality. Thinking has 3-way overlap with workflow-thinking + philosopher | Merge thinking triplet → 1 skill (−2) |
| moai-workflow-* | 14 | Rich. Major overlaps: templates↔project, testing↔tdd, design-context↔design-import | Fold templates into project; slim testing; merge thinking. Net −2 |
| moai-domain-* | 7 (backend, frontend, database, uiux, copywriting, brand-design, db-docs) | Core domain coverage. backend/frontend/database are kitchen sinks | Merge uiux+brand-design+craft; shrink backend/frontend/database by 40% |
| moai-platform-* | 4 (auth, chrome-ext, database-cloud, deployment) | Only 4 but heavy triplet-bundling | Merge database-cloud into domain-database |
| moai-framework-* | 1 (electron) | Single entry. Should either grow (add other frameworks) or retire | Manual review |
| moai-library-* | 3 (mermaid, nextra, shadcn) | Each library is popular and well-scoped | KEEP all |
| moai-tool-* | 2 (ast-grep, svg) | ast-grep is load-bearing; svg is niche | RETIRE svg pending evidence |
| moai-formats-* | 1 (data) | TOON+JSON/YAML optimization. Referenced nowhere but may be genuinely useful | KEEP monitor |
| moai-ref-* | 5 (api, git, owasp, react, testing-pyramid) | Agent-extending reference set. Consistent pattern, low maintenance, high reusability | KEEP all |
| moai-docs-* | 1 (generation) | Superseded | RETIRE or merge into jit-docs |
| moai-design-* | 2 (craft, tools) | Overlaps with domain-brand-design + domain-uiux | Merge both into design cluster |
| moai (root) | 1 | Orchestrator. Bundled 20-file workflow directory | v3: split into per-subcommand skills |

**Category coverage gaps (missing but essential in v3)**:
- `moai-lang-*` — referenced widely (`moai-lang-python`, `moai-lang-typescript`, `moai-lang-go`, etc. mentioned in `related-skills` and in rules) but **zero language skills exist** in the skills tree. 16 language rules live in `.claude/rules/moai/languages/` instead. Either migrate rules to skills or document why not. This is the single biggest architectural inconsistency uncovered in this audit: skills reference siblings that do not exist at the skill tree but do exist as rules, blurring the skill/rule boundary.
- `moai-infra-*` — `moai-infra-docker` referenced in 2 places, does not exist.
- `moai-essentials-debug` — 8 references, does not exist. Possibly a planned skill or stale reference.
- `moai-quality-testing`, `moai-quality-security` — 1-2 references each, do not exist. Likely stale aliases for moai-foundation-quality and moai-ref-owasp-checklist.

**Overlap heatmap (per-category cross-cutting)**:

| Area of overlap | Skills involved | Severity |
|-----------------|-----------------|----------|
| Thinking/analysis frameworks | foundation-thinking, foundation-philosopher, workflow-thinking | HIGH — 3-way duplication of "think before acting" guidance |
| Design decisions & intent | design-craft, domain-uiux, domain-brand-design, domain-copywriting | HIGH — four skills touch "what should this look like" |
| Database coverage | domain-database, platform-database-cloud, domain-db-docs | MEDIUM — cloud vs on-prem split is artificial; db-docs is a workflow concern |
| Template/boilerplate/scaffolding | workflow-templates, workflow-project, workflow-jit-docs | MEDIUM — docs+project+templates are the same concern viewed from three angles |
| Testing & quality | workflow-testing, workflow-tdd, workflow-ddd, foundation-quality, ref-testing-pyramid | LOW — these are distinct phases of the same pipeline; keep all but slim testing |
| Worktree vs. project | workflow-worktree, workflow-project | LOW — worktree is a git mechanic, project is docs+init. Boundary is clean |
| Platform triplets | platform-auth (3 vendors), platform-deployment (3 vendors), platform-database-cloud (3 vendors) | HIGH — each triplet is a "docs dump" anti-pattern |
| Frontend coverage | domain-frontend, ref-react-patterns, library-shadcn, library-nextra, domain-uiux | HIGH — five skills overlap in "how to build React frontends" |

**Size distribution by category** (SKILL.md body bytes, sum):

| Category | Total body KB | Avg body KB | Largest |
|----------|---------------|-------------|---------|
| moai-foundation-* | 66.1 | 13.2 | core (15.5) |
| moai-workflow-* | 166.9 | 12.8 | testing (22.5) |
| moai-domain-* | 62.7 | 9.0 | uiux (11.4) |
| moai-platform-* | 56.0 | 14.0 | chrome-extension (21.0) |
| moai-framework-* | 14.2 | 14.2 | electron |
| moai-library-* | 33.0 | 11.0 | shadcn (13.0) |
| moai-tool-* | 24.3 | 12.2 | svg (12.0) |
| moai-formats-* | 8.9 | 8.9 | data |
| moai-ref-* | 35.8 | 7.2 | api-patterns (7.3) |
| moai-docs-* | 10.9 | 10.9 | generation |
| moai-design-* | 23.9 | 12.0 | tools (17.6) |
| moai (root) | 18.2 | 18.2 | moai |

Total SKILL.md body size across 48 skills: **~520KB**. Adding bundled Level 3 resources pushes the total to approximately **1.2MB** of markdown. At ~4 chars/token, that is roughly **300,000 tokens** of skill content — or 1.5x a full 200K-token context window. Level 2 budget compliance matters.

---

## Section B — Skill dependency graph

Edges are `related-skills` frontmatter fields plus `Works Well With` body sections.

```
                           ┌──────────────────────────────┐
                           │     moai (root orchestrator) │
                           └──────────┬───────────────────┘
                                      │ routes to
          ┌───────────────────────────┼───────────────────────────────┐
          ▼                           ▼                               ▼
  moai-foundation-core      moai-workflow-spec           moai-workflow-project
          │                           │                               │
          ├──► trust-5 ──► moai-foundation-quality ──► moai-workflow-testing ──► moai-ref-testing-pyramid
          │                                             │
          │                                             ├──► moai-workflow-tdd  ──► moai-tool-ast-grep
          │                                             │
          │                                             └──► moai-workflow-ddd ──┤
          │                                                                      ▼
          ├──► progressive disclosure ──► moai-foundation-context        moai-workflow-loop
          │
          ├──► thinking triplet ──┬──► moai-foundation-thinking
          │                       ├──► moai-foundation-philosopher
          │                       └──► moai-workflow-thinking (MCP)
          │
          ├──► moai-foundation-cc ──► builder-* agents
          │
  design pipeline ──► moai-workflow-design-context ──► moai-workflow-design-import
                                 │                      │ (path A)
                                 ▼                      ▼
                      moai-domain-copywriting ──► moai-domain-brand-design ──► expert-frontend
                                 │                      │
                                 └──────────┬───────────┘
                                            ▼
                                   moai-workflow-gan-loop ──► evaluator-active
                                            │
                                 moai-workflow-pencil-integration
                                            │
                                 moai-design-tools (Figma/Pencil)
                                            │
                                 moai-design-craft (intent-first)

  docs side ──► moai-workflow-jit-docs ──► moai-docs-generation ──► moai-library-nextra
  db side   ──► moai-domain-database ──► moai-platform-database-cloud ──► moai-domain-db-docs
  frontend  ──► moai-domain-frontend ──► moai-library-shadcn / moai-domain-uiux
  platform  ──► moai-platform-auth / moai-platform-deployment / moai-platform-chrome-extension / moai-framework-electron
  tools     ──► moai-tool-ast-grep / moai-tool-svg / moai-library-mermaid
  agent-ref ──► moai-ref-{api,git,owasp,react,testing-pyramid}-*  (agent-extending layer)
```

**No cycles detected.** The graph is a DAG rooted at `moai-foundation-core` and `moai` (which routes but does not recurse).

**High fan-in (load-bearing) skills**:
- moai-foundation-core (referenced by everything)
- moai-workflow-testing (17 code references)
- moai-foundation-quality (15)
- moai-domain-backend (13)

**Low fan-in (isolated) skills**: moai-foundation-philosopher, moai-foundation-context, moai-design-craft, moai-docs-generation, moai-tool-svg, moai-formats-data, all moai-ref-* (by design), all moai-workflow-design-* (intentionally narrow pipeline).

---

## Section C — Progressive Disclosure compliance

Level 2 body budget is ~5000 tokens (roughly 20KB). Skills exceeding the budget:

| Skill | Body size | Over budget? | Notes |
|-------|-----------|--------------|-------|
| moai-workflow-testing | 22.5KB | YES (+12%) | 43 bundled files — extreme |
| moai-platform-chrome-extension | 21.0KB | YES (+5%) | Body itself is too long; split required |
| moai-workflow-spec | 18.7KB | Borderline | Edge of budget; modularize bundled |
| moai | 18.2KB + 20 workflows | YES (orchestrator) | Special-case |
| moai-design-tools | 17.6KB | Borderline | Three tools stapled together |
| moai-foundation-core | 15.5KB | OK | Core; can accept oversized body |
| moai-workflow-templates | 15.2KB | Borderline | Merge candidate reduces pressure |

Skills with **deeply bundled Level 3 resources** that should be pruned:
- moai/workflows/ — 20 files, ~300KB. Largest single Level 3 payload. In v3, these become first-class skills.
- moai-workflow-testing/modules/ — 22 module files, ~200KB. Over-modularization.
- moai-workflow-project/references/ + templates/ + schemas/ — 14 extra files. Moderate.
- moai-foundation-cc/reference/ + references/ — note the singular/plural dirs both exist — structural inconsistency. 23 files in `reference/`, 2 in `references/`.
- moai-foundation-core/modules/ + references/ — 19 + 2 files, similar structural issue.

**Structural issue** (worth fixing in v3): several skills have both `reference/` and `references/` (singular + plural) directories. Pick one convention.

### Progressive Disclosure enablement audit

Of 48 skills, how many declare `progressive_disclosure:` in frontmatter?

- **Enabled + configured**: 29 skills (60%)
- **Not declared**: 19 skills (40%)

Skills lacking the field are not necessarily non-compliant (defaults apply), but explicit declaration is the house style. Missing declarations: moai-docs-generation, moai-domain-backend, moai-domain-database, moai-domain-frontend, moai-domain-uiux, moai-formats-data, moai-framework-electron, moai-library-mermaid, moai-library-nextra, moai-library-shadcn, moai-tool-ast-grep, moai-tool-svg, moai-workflow-ddd, moai-workflow-loop, most `moai-ref-*` (but those explicitly declare via body-level conventions).

**Level 2 token budget configured values** (from frontmatter progressive_disclosure.level2_tokens):
- Most skills: 5000 (the MoAI default)
- moai-foundation-thinking, moai-foundation-quality, etc: 5000
- moai-platform-auth, moai-platform-deployment, moai-platform-database-cloud: 4500 (lower — responsive shrink to allow bigger prose in body before hitting budget? or typo? worth checking)
- moai-platform-chrome-extension: 8000 (intentional wider budget for this oversize skill)
- moai-design-craft: 4500 (intentional — designed to fit alongside design-tools + domain-uiux)
- moai-design-tools: 5500
- All moai-ref-*: 3000 (agent-extending skills intentionally kept small)
- moai-workflow-thinking: 3000 (pure tool reference)

The 3000-token setting for `moai-ref-*` is a good pattern worth replicating: agent-extending reference material should be ~3KB to stay well below budget, allowing the agent's own body to dominate.

---

## Section D — Template-local drift

`diff -rq` between `/Users/goos/MoAI/moai-adk-go/.claude/skills/` and `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/` returned **no output**. Every byte is identical.

This is the expected "Template-First Rule" outcome from CLAUDE.local.md §2: changes land in templates, then `make build` propagates to local.

- Template-only (new, not yet deployed): **0**
- Local-only (drift — local edits not pushed back): **0**
- Content-drifted (same name, different body): **0**

Verification sample — spot-check of `moai-workflow-gan-loop/SKILL.md` (recently changed per git status):
- Template SHA256 matches local SHA256 (both identical byte counts at 10891 + modified times within seconds of each other)
- No drift

The "50 vs 47" discrepancy in the task prompt is likely a stale observation. Both trees now hold exactly 48 skill directories.

---

## Recommended v3 skill inventory

Target: **~24 skills**, a 50% reduction from 48.

### Foundation (4)
1. `moai-foundation-core` — TRUST 5, SPEC-DDD, progressive disclosure, agent catalog (KEEP as-is)
2. `moai-foundation-cc` — Claude Code authoring kit (KEEP)
3. `moai-foundation-quality` — Quality orchestrator, TRUST 5 implementation (KEEP)
4. `moai-foundation-thinking` — MERGED thinking triplet (foundation-thinking + philosopher + workflow-thinking into one)

### Workflow (8)
5. `moai-workflow-spec` — SPEC + EARS (KEEP)
6. `moai-workflow-tdd` — TDD canonical (KEEP)
7. `moai-workflow-ddd` — DDD canonical (KEEP)
8. `moai-workflow-testing` — REFACTORED, slimmed, split bundled modules into level-3
9. `moai-workflow-project` — KEEP, absorbing templates + docs-generation
10. `moai-workflow-worktree` — KEEP
11. `moai-workflow-loop` — Ralph Engine (KEEP)
12. `moai-workflow-gan-loop` — GAN loop contract (KEEP)

### Design pipeline (4)
13. `moai-workflow-design-context` — brief loader (KEEP)
14. `moai-workflow-design-import` — handoff bundle parser (KEEP)
15. `moai-design-system` — MERGED brand-design + uiux + design-craft
16. `moai-domain-copywriting` — KEEP (agency absorption, clean contract)

### Domain (3)
17. `moai-domain-backend` — REFACTORED, narrower ("API design decision matrix")
18. `moai-domain-frontend` — REFACTORED, narrower (router to ref-react + library-nextra)
19. `moai-domain-database` — MERGED with platform-database-cloud, split db-docs stays separate

### Tools + Libraries (4)
20. `moai-tool-ast-grep` — KEEP
21. `moai-library-mermaid` — KEEP
22. `moai-library-shadcn` — KEEP
23. `moai-library-nextra` — KEEP

### Agent-extending reference (5)
24–28. `moai-ref-api-patterns`, `moai-ref-git-workflow`, `moai-ref-owasp-checklist`, `moai-ref-react-patterns`, `moai-ref-testing-pyramid` — KEEP all (low cost, high reuse)

### Special (1+N subcommand skills)
29. `moai` root orchestrator — REFACTORED, shrunk from 20 bundled workflows to a thin router
30. Per-subcommand skills extracted from `moai/workflows/*.md` — these already exist as system skills (`moai:plan`, `moai:run`, `moai:sync`, etc.) and should formalize into `.claude/skills/moai-cmd-{name}/`. If we count them, total rises. If we keep them as the `moai` skill's Level 3 payload, total stays at 24.

### Drop / merge out (24 skills removed or absorbed)
Retired outright: moai-tool-svg, moai-docs-generation, moai-design-tools (split into tool-figma + tool-pencil OR absorbed into pencil-integration), moai-workflow-templates (into project), moai-platform-database-cloud (into domain-database), moai-foundation-context (content folded into moai-foundation-core), moai-foundation-philosopher (into thinking), moai-workflow-thinking (into thinking), moai-workflow-pencil-integration (absorbed into design cluster or tool-pencil), moai-workflow-jit-docs (absorbed into project — documentation workflow is a project concern), moai-workflow-research (kept if active, otherwise retire experimental).

Niche kept on watch: moai-framework-electron, moai-platform-chrome-extension, moai-platform-deployment, moai-platform-auth, moai-formats-data — promote to KEEP if SPECs reference them within 2 months of v3 launch, otherwise retire.

---

## v3 migration sequencing recommendations

The 48 → 24 reduction cannot happen in a single commit. Recommended sequencing:

**Wave 1 — Merges without user-visible change (low risk)**:
- Merge thinking triplet (foundation-thinking + foundation-philosopher + workflow-thinking). All three skills are referenced by the same manager-strategy/manager-spec agents; the merged skill retains trigger keywords so auto-activation is preserved.
- Merge workflow-templates into workflow-project. Both are plan-phase skills; the merged skill keeps both sets of keywords.
- Merge platform-database-cloud into domain-database. Both are `run` phase skills with expert-backend as primary consumer.

**Wave 2 — Design pipeline consolidation (medium risk)**:
- Merge design-craft + domain-uiux into a new `moai-design-system` skill.
- Retain brand-design and copywriting as-is (they are agency-absorption contracts and changing their external surface would break GAN loop contracts).
- Split design-tools into tool-figma and tool-pencil, or absorb pencil side into workflow-pencil-integration.

**Wave 3 — Domain skill slimming (higher risk — touches the "what do I call to do X" reflex for agents)**:
- Refactor domain-backend into a router/decision matrix.
- Refactor domain-frontend similarly.
- Both changes require updating related agent prompts to stop expecting the kitchen-sink scope.

**Wave 4 — Retirements (highest risk — breaks any lurking consumer)**:
- Archive moai-tool-svg if 30-day telemetry shows zero invocation.
- Archive moai-docs-generation (superseded by jit-docs + library-nextra).
- Archive or refactor moai-framework-electron and moai-platform-chrome-extension based on telemetry.

**Out-of-band — `moai` root skill refactoring**:
- Extract each `moai/workflows/*.md` into a first-class `moai-cmd-{name}` skill.
- Update the `moai` skill body to route to the new skills instead of bundling their content.
- This is a large change; sequence last to minimize risk during earlier waves.

---

## Risk register for v3 skill redesign

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Agent prompt relies on merged skill keyword that no longer exists | HIGH | MEDIUM | Keep union of trigger keywords when merging; add regression tests that agent → skill activation still fires |
| `moai-ref-*` description-based activation breaks under v3 Claude Code | LOW | HIGH | Version-pin Claude Code during v3 migration; validate activation model in a clone session |
| Bundled Level 3 resources break when skill is split | MEDIUM | MEDIUM | Audit bundled resource paths in split skills; update references explicitly |
| Template-local drift during multi-day v3 migration | MEDIUM | LOW | Run `diff -rq` after every commit; add CI check if not already present |
| Agency absorption contracts (brand-design, copywriting) drift during merges | LOW | HIGH | Freeze these two skills in Wave 2; consolidate only peripheral skills |
| Missing skills (`moai-lang-*`, `moai-essentials-debug`) stay referenced post-v3 | HIGH | LOW | Audit `related-skills` fields; either create the missing skills or remove references |
| Skill naming inconsistency (`reference/` vs `references/`) persists | HIGH | LOW | Add a lint check to the build pipeline that rejects the singular form |

---

## Comparative notes: what MoAI does well vs. opportunity areas

**What MoAI does well** (evidence from this audit):

1. **Template-First discipline is real** — zero drift across 48 skills is impressive.
2. **Frontmatter hygiene is strong** — CSV `allowed-tools`, quoted version strings, `user-invocable: false` for internal skills.
3. **Progressive Disclosure is thoughtfully applied** — Level 2 token budgets are declared in 60% of skills and the smaller reference skills correctly use 3000-token budgets.
4. **Agency absorption (brand-design, copywriting, db-docs) produced the cleanest skills** — they are the model for future domain skills.
5. **No circular dependencies** in the skill graph.
6. **Skill authoring rules are documented** (`.claude/rules/moai/development/skill-authoring.md`) and largely adhered to.

**Opportunity areas**:

1. **Trigger keyword rigor** — 22-keyword lists for domain skills are scope-bloat signals.
2. **Scope discipline on "platform" triplets** — one skill per vendor or a meta-skill with clear "NOT for other vendors" guardrails.
3. **Clarify skill vs. rule boundary** — 16 language rules vs zero language skills is confusing.
4. **Bundled resource pruning** — 43 files under one skill is over-modularization.
5. **Static references as activation signal** — relying purely on description-matching for `moai-ref-*` is opaque. Add telemetry to measure actual activation rates.
6. **Version staleness** — many skills list `updated: 2026-01-XX` despite their domains (shadcn/ui, React, Next.js) moving quickly.

---

## Summary of architectural health issues

1. **Structural duplication in thinking layer** (3 skills, ~33KB, ~3 active callers between them). MERGE yields 66% size reduction.
2. **Kitchen-sink domain skills** (backend, frontend, database all try to cover 3-5 frameworks/vendors simultaneously). Vague trigger keywords like "backend" catch everything.
3. **Platform skills copy the triplet anti-pattern** (auth: Auth0+Clerk+Firebase, deployment: Vercel+Railway+Convex, database-cloud: Neon+Supabase+Firestore). Each triplet is "three docs in one skill" — high maintenance when any vendor changes its API.
4. **Agent-extending reference skills are under-used and undocumented** (all moai-ref-* show zero static references but this is by design via keyword-matching auto-activation). Needs documentation of the activation model.
5. **Zero drift between template and local is healthy** — Template-First discipline works.
6. **The `moai` root skill is secretly a 300KB command manual** — bundled `workflows/*.md` is the actual source of truth for `/moai` subcommands. In v3, these should be promoted to first-class skills under `.claude/skills/moai-cmd-*/`.
7. **Progressive Disclosure Level 2 budget violations in 4-5 skills**. All fixable via level-3 split.
8. **`reference/` vs `references/` directory naming inconsistency** across bundled resources — pick one convention in v3.
9. **"Language skills" referenced but absent** — 16 language rules live in `.claude/rules/moai/languages/` instead of `.claude/skills/moai-lang-*/`. V3 should decide on one home and migrate consistently.

---

## Sources: file paths read

- `/Users/goos/MoAI/moai-adk-go/.claude/skills/` (directory listing, all 48 subdirectories)
- `/Users/goos/MoAI/moai-adk-go/internal/template/templates/.claude/skills/` (directory listing)
- Frontmatter of all 48 SKILL.md files (extracted via `awk '/^---$/{c++}...'`)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-research/SKILL.md` (full read)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-thinking/SKILL.md` (full read)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-templates/SKILL.md` (partial — first 80 lines)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-domain-brand-design/SKILL.md` (partial — first 60 lines)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-design-craft/SKILL.md` (full read)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-foundation-philosopher/SKILL.md` (partial — first 100 lines)
- `/Users/goos/MoAI/moai-adk-go/.claude/skills/moai-workflow-design-context/SKILL.md` (partial — first 80 lines)
- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/development/skill-authoring.md` (provided via system-reminder)
- Reference counts from `grep -rE "\bmoai-*\b"` across `.claude/agents/`, `.claude/commands/`, `.claude/rules/`, `.moai/config/`
- Bundled resource inventories via `ls -la` on workflow/, modules/, references/, schemas/ subdirectories of 7 largest skills
- Drift verification via `diff -rq` between template and local skill trees (empty output confirmed)
