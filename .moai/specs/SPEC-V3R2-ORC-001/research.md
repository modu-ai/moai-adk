---
spec_id: SPEC-V3R2-ORC-001
phase: "0.5 — Codebase Research"
created_at: 2026-05-09
author: manager-spec
status: research-complete
---

# Research: Agent roster consolidation (22 → 17)

Phase 0.5 codebase research for SPEC-V3R2-ORC-001. Audits the current agent
roster present in this branch (`feature/SPEC-V3R2-ORC-001-roster`, base
`origin/main` at `464366583`), maps each file to its v3r2 destiny, and cites
file:line evidence for every consolidation decision.

---

## 1. Current Roster (as-found at base `464366583`)

Files under `internal/template/templates/.claude/agents/moai/` (template tree,
canonical source per CLAUDE.local.md §2 Template-First Rule):

| # | Agent file | Lines | Frontmatter status | Body status |
|---|------------|------:|--------------------|-------------|
|  1 | `builder-agent.md` | 116 | active (sonnet, bypassPermissions, memory:user) | full body, 5-phase workflow |
|  2 | `builder-plugin.md` | 149 | active | full body, 5-phase workflow |
|  3 | `builder-skill.md` | 105 | active | full body, 5-phase workflow |
|  4 | `claude-code-guide.md` | 105 | active (sonnet, plan mode) | upstream-investigation specialist |
|  5 | `evaluator-active.md` | 113 | active | full body |
|  6 | `expert-backend.md` | 141 | active | full body, references retired `manager-ddd` (L62, L119) |
|  7 | `expert-debug.md` | 213 | active | full body, references retired `manager-ddd` (L59, L90+) |
|  8 | `expert-devops.md` | 123 | active | full body |
|  9 | `expert-frontend.md` | 202 | active | full body, mixes code + Pencil MCP |
| 10 | `expert-mobile.md` | 185 | active | full body (post-R5 addition; not in original roster) |
| 11 | `expert-performance.md` | 110 | active | full body |
| 12 | `expert-refactoring.md` | 102 | active | full body |
| 13 | `expert-security.md` | 132 | active | full body |
| 14 | `expert-testing.md` | 116 | active | full body, declares OUT-OF-SCOPE for unit/load tests |
| 15 | `manager-brain.md` | 151 | active (opus, effort: xhigh) | full body (post-R5 addition; not in original roster) |
| 16 | `manager-cycle.md` | 237 | **active (already created)** | full body, accepts `cycle_type: ddd\|tdd` |
| 17 | `manager-ddd.md` | 39 | **already retired stub** (`retired: true`, `retired_replacement: manager-cycle`) | one-line redirect |
| 18 | `manager-docs.md` | 107 | active | full body |
| 19 | `manager-git.md` | 156 | active | full body |
| 20 | `manager-project.md` | 126 | active | full body, retains `settings_modification`, `glm_configuration`, `template_update_optimization` modes (L58-61) |
| 21 | `manager-quality.md` | 115 | active | full body, NO Diagnostic Sub-Mode section, declares Context7 in tools list (L13) |
| 22 | `manager-spec.md` | 186 | active | full body |
| 23 | `manager-strategy.md` | 141 | active | full body |
| 24 | `manager-tdd.md` | 39 | **already retired stub** (`retired: true`, `retired_replacement: manager-cycle`) | one-line redirect |
| 25 | `plan-auditor.md` | 270 | active, NO `memory:` field (L1-17) | full body |
| 26 | `researcher.md` | 59 | active | minimal body |

**Total: 26 files in template tree** (not 22 as originally framed in
spec.md §1). The discrepancy is explained in §2 below.

Local tree (`.claude/agents/moai/`) currently lacks `manager-cycle.md` and
diverges from the template on 5 files (`evaluator-active`, `expert-backend`,
`expert-debug`, plus two more — captured by `diff -r` baseline below). The
`make build` step in M2 will resolve the divergence.

### 1.1 Roster discrepancy explanation

The spec.md baseline of "22 agents" matches R5 audit
`.moai/design/v3-redesign/research/r5-agent-audit.md` L14, which excludes
Claude Code built-in surfaces (`Explore`, `Plan`, `general-purpose`,
`statusline-setup`, `claude-code-guide`). It also predates two post-R5
additions:

- `expert-mobile.md` (post-R5; mobile native + cross-platform specialist)
- `manager-brain.md` (added by SPEC-IDEA-002 / brain workflow line)
- `claude-code-guide.md` (treated by R5 as a Claude built-in but exists as a
  project agent file)

These 3 agents are scope-additive to the R5-baseline 22. They are NOT in scope
for ORC-001 retirement and remain KEEP per their post-R5 introductions. The
final roster after ORC-001 is therefore **20 active agents**, not 17, when
counted against the template tree (17 from the R5-baseline 22 + 3 post-R5
additions).

The plan reconciles this by:
1. Treating REQ-ORC-001-001's "exactly 17" as the R5-baseline subset.
2. Adding §10.1 amendment to spec.md (during M5 documentation step) noting
   that the 3 post-R5 agents are scope-additive and not subject to the −5
   delta.

This is captured as **OQ-1** below; manager-spec will resolve at audit.

---

## 2. Already-Completed Work (carry-over from prior SPECs)

The following retirements were performed by **SPEC-V3R3-RETIRED-AGENT-001**
(merged 2026-05-04, PR #776 / commit `20d77d931`) and
**SPEC-V3R3-RETIRED-DDD-001** (merged 2026-05-04, PR #781 / commit
`f0d34cef6`):

| Already-completed change | Evidence |
|--------------------------|----------|
| `manager-cycle.md` created with `cycle_type: ddd\|tdd` parameter | `internal/template/templates/.claude/agents/moai/manager-cycle.md:1-44` |
| `manager-ddd.md` retired to stub | `internal/template/templates/.claude/agents/moai/manager-ddd.md:1-12` |
| `manager-tdd.md` retired to stub | `internal/template/templates/.claude/agents/moai/manager-tdd.md:1-12` |
| Migration table (DDD → cycle_type=ddd; TDD → cycle_type=tdd) documented in `manager-cycle.md` | `internal/template/templates/.claude/agents/moai/manager-cycle.md:60-66` |

This SPEC therefore inherits a **partially-complete roster** and must:

1. Verify the already-done work (M1) — do not regress.
2. Apply the **5 outstanding consolidations** (M2):
   - Retire `builder-agent.md` (full body) → `builder-platform.md` (new)
   - Retire `builder-skill.md` (full body) → `builder-platform.md` (new)
   - Retire `builder-plugin.md` (full body) → `builder-platform.md` (new)
   - Retire `expert-debug.md` (full body) → `manager-quality.md` Diagnostic Sub-Mode
   - Retire `expert-testing.md` (full body) → `manager-cycle.md` strategy + `expert-performance.md` load mode
3. Apply the **6 refactor passes** (M3):
   - manager-quality: add Diagnostic Sub-Mode section; drop Context7 preload
   - manager-project: scope-shrink to `.moai/project/` only
   - expert-backend: trigger dedup + cap to 12-15 EN tokens
   - expert-performance: optional `Write` scope addition
   - manager-git: trim trigger list per P-A16
   - plan-auditor: add `memory: project` field
4. Sync downstream references (M4): rule files, skill bodies, root CLAUDE.md.

---

## 3. Consolidation Evidence Catalogue (file:line citations)

Each consolidation decision is backed by concrete file:line evidence found
during this Phase 0.5 sweep. Exit-code-zero `grep -n` invocations are listed
inline so the auditor can replay them.

### 3.1 manager-cycle (already done — verification only)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/manager-cycle.md:17` | tools list: `Read, Write, Edit, MultiEdit, Bash, Grep, Glob, TodoWrite, Skill, mcp__sequential-thinking__sequentialthinking, mcp__context7__resolve-library-id, mcp__context7__get-library-docs` (write-capable, no advisor-only violation) |
| `internal/template/templates/.claude/agents/moai/manager-cycle.md:54` | `**cycle_type**: Must be specified as 'ddd' or 'tdd' in the spawn prompt.` |
| `internal/template/templates/.claude/agents/moai/manager-cycle.md:60-66` | Migration Notes table documents `manager-ddd` → `cycle_type=ddd` and `manager-tdd` → `cycle_type=tdd` |
| `internal/template/templates/.claude/agents/moai/manager-cycle.md:8-15` | Trigger keywords: union of DDD (EN/KO/JA/ZH) + TDD (EN/KO/JA/ZH) preserved per REQ-ORC-001-009 |

### 3.2 builder-platform (NEW — to be authored in M2)

Three full-body builders to merge:

| Source agent | Lines | Notable rows |
|--------------|------:|--------------|
| `builder-agent.md` | 116 | tools L13: `Read, Write, Edit, ..., Agent, Skill, mcp__sequential-thinking, mcp__context7__*`; skills L17-19: `moai-foundation-cc, moai-foundation-core, moai-workflow-project` |
| `builder-skill.md` | 105 | tools L13 same; skills L17-19: `moai-foundation-core, moai-foundation-cc, moai-workflow-project` |
| `builder-plugin.md` | 149 | tools L13 same; skills L17-19: `moai-foundation-cc, moai-foundation-core, moai-workflow-project` |

Frontmatter overlap is essentially 100% (model: sonnet, permissionMode:
bypassPermissions, memory: user, identical tool list). Body workflow follows
the same 5-phase template (Requirements → Research → Architecture →
Implementation → Validation). The merge produces ONE file
`builder-platform.md` with `artifact_type:
agent|skill|plugin|command|hook|mcp-server|lsp-server` parameter. Per
REQ-ORC-001-009 the trigger keyword UNION must be preserved.

Trigger keyword union estimation:
- builder-agent EN: 6 keywords (`create agent, new agent, agent blueprint, sub-agent, agent definition, custom agent`)
- builder-skill EN: 5 keywords (`create skill, new skill, skill optimization, knowledge domain, YAML frontmatter`)
- builder-plugin EN: 9 keywords (`create plugin, plugin, plugin validation, plugin structure, marketplace, new plugin, marketplace creation, marketplace.json, plugin distribution`)
- Union (deduped): ~20 EN keywords; KO/JA/ZH same proportions; total ≤ 80.

### 3.3 expert-debug retirement (M2)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/expert-debug.md:13` | tools list: `Read, Grep, Glob, Bash, Skill, mcp__sequential-thinking, mcp__context7__*` — **NO Write/Edit** (P-A06 advisor-only finding) |
| `internal/template/templates/.claude/agents/moai/expert-debug.md:59` | body declares: `Code implementation (delegate to manager-ddd)` (=delegation router) |
| `internal/template/templates/.claude/agents/moai/expert-debug.md:198-204` | body declares: `expert-debug ... is a subagent so MUST NOT call AskUserQuestion` (already common-protocol-aware) |
| `internal/template/templates/.claude/agents/moai/manager-quality.md:13` | manager-quality tools list: `Read, Grep, Glob, Bash, Skill, mcp__sequential-thinking, mcp__context7__*` — same shape, can absorb |
| `internal/template/templates/.claude/agents/moai/manager-quality.md:14` | manager-quality `permissionMode: plan` — preserves diagnostic-only stance |

Decision: expert-debug body becomes stub; its diagnostic-routing table is
inlined into manager-quality.md as a new "Diagnostic Sub-Mode" section.
Common-Protocol scrub (the AskUserQuestion HARD rule) is preserved verbatim
into the new section.

### 3.4 expert-testing retirement (M2)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/expert-testing.md:11` | NOT-FOR list explicitly excludes: `production code implementation, architecture design, DevOps, security audits, performance optimization` |
| `internal/template/templates/.claude/agents/moai/expert-testing.md:13` | tools list includes Write/Edit but body declares strategy-only role (P-A07 finding) |
| `internal/template/templates/.claude/agents/moai/expert-performance.md:11-12` | scope: `performance, profiling, optimization, slow, bottleneck, latency` — accepts load-test responsibility |
| `internal/template/templates/.claude/agents/moai/manager-cycle.md:54-58` | accepts cycle_type=tdd which includes RED-phase test strategy |

Decision: expert-testing body becomes stub; strategy responsibility folds
into manager-cycle (cycle_type=tdd RED phase); load-test execution folds into
expert-performance via a new `--deepthink load-test` mode.

### 3.5 manager-project scope shrink (M3)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/manager-project.md:32` | declares: `CANNOT use AskUserQuestion — all user choices must be pre-collected by the command` (consistent with subagent rules) |
| `internal/template/templates/.claude/agents/moai/manager-project.md:58` | `settings_modification → Configuration update` (over-scoped — belongs in `moai update`) |
| `internal/template/templates/.claude/agents/moai/manager-project.md:60` | `template_update_optimization → Template enhancement` (over-scoped — belongs in `moai update -t`) |
| `internal/template/templates/.claude/agents/moai/manager-project.md:61` | `glm_configuration → GLM API integration setup` (over-scoped — belongs in `moai glm` / `moai cc`) |

Decision: M3 strips L57-65 routing block, leaving only `.moai/project/`
document generation responsibility per REQ-ORC-001-008.

### 3.6 expert-backend trigger dedup (M3)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/expert-backend.md:6` | EN trigger row contains 22 tokens: `backend, API, server, authentication, database, REST, GraphQL, microservices, JWT, OAuth, SQL, NoSQL, PostgreSQL, MongoDB, Redis, Oracle, PL/SQL, schema, query, index, data modeling` (over the 12-15 cap from P-A17) |
| same file L7 | KO row repeats `Oracle` standalone in addition to `오라클` (near-duplicate per P-A17) |
| `grep -c "Oracle" expert-backend.md` | returns 4 (EN, KO standalone+localized, JA, ZH) — confirms duplication pattern |

Decision: M3 reduces EN row to 12-15 high-precision tokens; localized rows
keep one form per language.

### 3.7 manager-git trigger and Context7 cleanup (M3)

Per P-A16 (R5 audit) and REQ-ORC-001-010 / 011:

| Citation | Evidence |
|----------|----------|
| `grep -c "context7" internal/template/templates/.claude/agents/moai/manager-git.md` | returns 0 — already clean (no action needed for REQ-011 part) |
| `internal/template/templates/.claude/agents/moai/manager-git.md:6-9` | trigger row currently 12+ EN tokens; reducer to ~8 high-precision per P-A16 |

Decision: M3 trims trigger list only; Context7 already absent.

### 3.8 manager-quality Context7 drop (M3)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/manager-quality.md:13` | tools list currently includes `mcp__context7__resolve-library-id, mcp__context7__get-library-docs` |
| `internal/template/templates/.claude/agents/moai/manager-quality.md:1-30` | body never references Context7 lookups — preload is unused |

Decision: M3 drops Context7 preload from manager-quality tools list per
REQ-ORC-001-011.

### 3.9 plan-auditor memory field (M3)

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/plan-auditor.md:1-17` | frontmatter contains `name, description, tools, model, effort, permissionMode` — **NO `memory:` field** |
| spec.md REQ-ORC-001-013 | requires `memory: project` |

Decision: M3 adds `memory: project` line after `permissionMode: default` in
plan-auditor frontmatter.

### 3.10 expert-performance optional Write scope (M3)

Optional per REQ-ORC-001-014:

| Citation | Evidence |
|----------|----------|
| `internal/template/templates/.claude/agents/moai/expert-performance.md:1-30` | tools list excludes Write/Edit; body recommends "Create analysis file" — P-A08 contradiction |

Decision: M3 adds `Write` to tools list scoped via PreToolUse hook check
limiting writes to `.moai/docs/performance-analysis-*.md` only. (Hook rule
itself remains future work; this SPEC adds the tool entry only.)

### 3.11 Downstream reference sweep (M4)

Skill files referencing retired agents (full sweep `grep -rln
"manager-ddd|manager-tdd|builder-agent|builder-skill|builder-plugin|expert-debug|expert-testing"
.claude/skills/ internal/template/templates/.claude/skills/`):

- 57 skill files in template + local trees
- 28 rule files in template + local trees
- 4 references in root `CLAUDE.md` (L62, L136, L147, L378)
- 2 references in template `CLAUDE.md` (L136, L378)

Sweep targets for M4:
- Replace `Use the manager-ddd subagent` → `Use the manager-cycle subagent
  with cycle_type=ddd`
- Replace `manager-tdd` → `manager-cycle with cycle_type=tdd`
- Replace `builder-agent / builder-skill / builder-plugin` → `builder-platform
  with artifact_type=agent|skill|plugin`
- Replace `expert-debug subagent` → `manager-quality (diagnostic mode)`
- Replace `expert-testing` → context-dependent (manager-cycle or
  expert-performance)
- Update Agent Catalog L106-126 in `CLAUDE.md` (manager 8→7, expert 8→6,
  builder 3→1 [builder-platform]; net 17 active R5-baseline + 3 post-R5)

---

## 4. Open Questions for Audit

OQ-1 — Roster delta arithmetic (post-R5 additions):
- **Question**: Does REQ-ORC-001-001 "exactly 17" treat `expert-mobile`,
  `manager-brain`, `claude-code-guide` as additive (final = 20) or
  in-scope-but-also-retired (final = 17)?
- **Recommendation**: Additive. They are post-R5 introductions with active
  user-facing roles and are not in the deprecation list per Master §8 BC
  catalog. spec.md §10.1 destiny table is authoritative and excludes them
  from the consolidation target.
- **Resolution path**: M1 audit step adds a clarifying note to spec.md or a
  HISTORY entry; manager-spec confirms with user via AskUserQuestion if the
  text-level interpretation is contested.

OQ-2 — Stub frontmatter shape consistency:
- **Question**: Should the new stubs (builder-agent, builder-skill,
  builder-plugin, expert-debug, expert-testing) match the existing
  manager-ddd/manager-tdd stub schema (`retired: true`,
  `retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`)?
- **Recommendation**: Yes, match exactly. This is the established
  SPEC-V3R3-RETIRED-DDD-001 pattern.
- **Resolution path**: M2 step uses the same stub template across all 5 new
  retirees.

OQ-3 — Migration of legacy SPEC bodies (deferred to MIG-001):
- **Question**: Should this SPEC rewrite `subagent_type="ddd-implementer"`
  inside `.claude/skills/moai-foundation-core/references/examples.md` lines
  118, 181 (and similar)?
- **Recommendation**: NO. Scope is agent-file only; example skill content is
  pedagogical and is rewritten by MIG-001.
- **Resolution path**: M4 sweep excludes `examples.md` files explicitly.

OQ-4 — Concurrent SPEC interaction with ORC-002:
- **Question**: Will ORC-002 CI lint reject the new bodies (manager-quality
  diagnostic section + builder-platform body) for any
  `AskUserQuestion`-text presence?
- **Recommendation**: M2 implementations adhere to the ORC-002 lint contract
  preemptively (no literal `AskUserQuestion` strings in body, only
  meta-references inside fenced quoted-rule blocks).
- **Resolution path**: M5 verification step runs the proposed ORC-002 lint
  rule locally before PR creation.

OQ-5 — `claude-code-guide.md` inclusion in active count:
- **Question**: R5 audit explicitly lists `claude-code-guide` as a Claude
  built-in (NOT a project agent). The file exists in this repo. Should it be
  retired or kept?
- **Recommendation**: KEEP. It is an active project-specific upstream
  investigator added post-R5. It is not in the consolidation list.
- **Resolution path**: spec.md §10.1 destiny table updated in M5 to add a
  row for `claude-code-guide` with destiny KEEP.

---

## 5. References

### Citation summary (file:line anchors used in this research)

Total citations: 30+ unique file:line references across the following
locations:

- 4× `internal/template/templates/.claude/agents/moai/manager-cycle.md` (L17, L54, L60-66, L8-15)
- 3× `internal/template/templates/.claude/agents/moai/manager-ddd.md` (L1-12)
- 3× `internal/template/templates/.claude/agents/moai/manager-tdd.md` (L1-12)
- 4× `internal/template/templates/.claude/agents/moai/expert-debug.md` (L13, L59, L198-204)
- 3× `internal/template/templates/.claude/agents/moai/expert-testing.md` (L11, L13)
- 5× `internal/template/templates/.claude/agents/moai/manager-project.md` (L32, L58, L60, L61, L57-65)
- 4× `internal/template/templates/.claude/agents/moai/expert-backend.md` (L6, L7, grep counts)
- 3× `internal/template/templates/.claude/agents/moai/manager-quality.md` (L13, L14, L1-30)
- 2× `internal/template/templates/.claude/agents/moai/manager-git.md` (L6-9, grep)
- 2× `internal/template/templates/.claude/agents/moai/plan-auditor.md` (L1-17)
- 1× `internal/template/templates/.claude/agents/moai/expert-performance.md` (L1-30)
- 4× root and template `CLAUDE.md` (L62, L136, L147, L378)
- 1× R5 audit `.moai/design/v3-redesign/research/r5-agent-audit.md` (L14)
- 1× CLAUDE.md Agent Catalog block (L106-126)

### Master design pointers

- `docs/design/major-v3-master.md` §4.4 (Layer 4 Orchestration)
- `docs/design/major-v3-master.md` §7.2 (Agent inventory table)
- `docs/design/major-v3-master.md` §8 (BC catalog: BC-V3R2-005, 009, 016)
- `docs/design/major-v3-master.md` §11.4 (ORC-001..005 definitions, L1050-1056)
- `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 6 (P-A05..P-A20)
- `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6
- `.moai/design/v3-redesign/research/r5-agent-audit.md` §Recommended v3 agent inventory

### Prior-art SPECs (carry-over context)

- SPEC-V3R3-RETIRED-AGENT-001 — established the retired-stub pattern
- SPEC-V3R3-RETIRED-DDD-001 — applied the pattern to manager-ddd/tdd, created manager-cycle

---

End of research.
