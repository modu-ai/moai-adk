---
id: SPEC-V3R6-AGENT-TEAM-REBUILD-001
artifact: plan
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
sync_commit_sha: "<pending>"
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L plan-phase artifact authored from spec.md §C.1 scope inventory + Audit 3 findings; 8 milestones (M1-M8); ~50-60 files in run-phase scope (17 agent files + 3 workflow skills + ~10 rule files + 3 hook scripts + ~13 template mirrors + predecessor SPEC supersedence + NOTICE.md). Supersedes SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 plan structure. |

---

## §A — Lifecycle Table

| Step | Phase | Owner | Status | Commit SHA | Notes |
|------|-------|-------|--------|------------|-------|
| 0 | Pre-flight (research.md authored — Audit 3 synthesis) | orchestrator + manager-spec | IN-PROGRESS | `<pending>` | This SPEC's research.md appends §H Audit 3 synthesis to base research |
| 0 | Plan-phase M0 (5 SPEC artifacts) | manager-spec | IN-PROGRESS | `<pending>` | spec.md + plan.md + acceptance.md + design.md + research.md (5 files Tier L) + predecessor SPEC supersedence (frontmatter-only) |
| 0.5 | plan-auditor (Tier L PASS ≥ 0.85, skip-eligible ≥ 0.90) | plan-auditor | NOT-STARTED | n/a | Single-pass + max-3 retry contract; design.md mandatory at this tier |
| 0.95 | Mode Selection (NEW per REQ-ATR-008) | orchestrator | NOT-STARTED | n/a | 5-mode autonomous selection; expected `sequential single-spawn manager-develop` per Finding A4 coding-task parallelism caveat (markdown-heavy, no real parallel implementation benefit) |
| 1 | Run-phase M1 — Agent retention frontmatter refinement (8 retained agents) | manager-develop | NOT-STARTED | n/a | REQ-ATR-001 + REQ-ATR-002 + REQ-ATR-003 + REQ-ATR-004; frontmatter + minor body refinement of 7 MoAI-custom retained agent files |
| 1 | Run-phase M2 — Workflow router skill phase-owner declarations | manager-develop | NOT-STARTED | n/a | REQ-ATR-001 (manager-strategy chain removal) + REQ-ATR-007 (Skill router invocation discipline preserved) + REQ-ATR-008 (Phase 0.95 Mode Selection logging); 3 router skills |
| 1 | Run-phase M3 — Agent archive to .moai/backups/ | manager-develop | NOT-STARTED | n/a | REQ-ATR-005; 12 phantom + domain-expert agents archived to `.moai/backups/agent-archive-2026-05-25/` (preserving structure) BEFORE deletion from `.claude/agents/` and template mirror |
| 1 | Run-phase M4 — Hook scripts authoring (3 NEW files) | manager-develop | NOT-STARTED | n/a | REQ-ATR-009 (Stop sync-phase quality gate) + REQ-ATR-014 (dependency manifest audit) + PostToolUse Status Transition Ownership + TaskCompleted team-mode AC verify |
| 1 | Run-phase M5 — Rule files updates (~10 files) | manager-develop | NOT-STARTED | n/a | REQ-ATR-007 + REQ-ATR-008 + REQ-ATR-012 + REQ-ATR-016; orchestration-mode-selection.md NEW + archived-agent-rejection.md NEW + 8 existing rule file updates |
| 1 | Run-phase M6 — Predecessor SPEC supersedence | manager-develop OR manager-spec | NOT-STARTED | n/a | REQ-ATR-006; `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` frontmatter `status: superseded` + HISTORY v0.1.2 entry (Status Transition Ownership Matrix `* → superseded` ownership → manager-spec; orchestrator may invoke manager-spec via dedicated spawn for this milestone) |
| 1 | Run-phase M7 — CLAUDE.md + CLAUDE.local.md doctrine updates | manager-develop | NOT-STARTED | n/a | REQ-ATR-001 (CLAUDE.md §4 Agent Catalog 17→8) + REQ-ATR-019 (NOTICE.md attribution) + REQ-ATR-020 (CLAUDE.local.md §23 manager-git PR doctrine reconciliation) + REQ-ATR-015 (AskUserQuestion mandatory restoration on GATE-2 detection) |
| 1 | Run-phase M8 — Template-First parity check + verification batch | manager-develop | NOT-STARTED | n/a | REQ-ATR-018; byte-for-byte mirror of all changes to `internal/template/templates/` + 7-item Trust-but-verify verification batch |
| 2 | Sync-phase | manager-docs | NOT-STARTED | n/a | CHANGELOG.md + frontmatter `status: in-progress → implemented` for all 5 artifacts + §E.4 Sync-phase Audit-Ready Signal in progress.md |
| 3 | Mx-phase | orchestrator | NOT-STARTED | n/a | Step C judgement per mx-tag-protocol.md §a — expected EVALUATE-SKIP (mostly markdown + 3 NEW shell scripts, 0 .go files, 0 goroutines, 0 fan_in delta in Go code); §E.5 Mx-phase Audit-Ready Signal in progress.md |
| 4 | 4-phase close | orchestrator | NOT-STARTED | n/a | Status `implemented → completed` + L60 atomic backfill (4-artifact frontmatter `sync_commit_sha:` consistency) |

Key dates: created 2026-05-25, target plan-phase complete 2026-05-25, run-phase target 2026-05-26 to 2026-05-28 (3-day Tier L window per Sprint 10 cadence).

---

## §B — Run-phase Strategy

### §B.1 Target: Single-pass (Tier L 1-pass success)

Tier L SPECs target a 1-pass run-phase completion per the plan-auditor max-3 retry contract. Given that this SPEC is **markdown + shell-script-only** (no Go source modifications), the single-pass probability is HIGH provided:

1. plan-auditor verdict on this plan.md achieves PASS ≥ 0.85 (Tier L threshold) — self-estimate ~0.88-0.91
2. Skip-eligible threshold 0.90 preferred (Tier L MARGINAL with 5-artifact set + 20 REQs + ≥20 ACs + 8 risks-with-mitigation pairing)
3. M1-M8 milestone scope respects the §C in-scope file inventory exactly (path-specific staging discipline L46)

### §B.2 Known Issues Section B — Pre-flight injection

Per `.claude/rules/moai/development/manager-develop-prompt-template.md` §1.B, the run-phase delegation prompt MUST inject the B1-B12 known-issues catalog. For this Tier L markdown-and-shell SPEC, the following B-categories are particularly relevant:

- **B1 Cross-platform Build Tags** — hook scripts MUST use POSIX-compatible bash (no `#!/bin/zsh`); Windows hook fallback per platform compatibility doctrine
- **B3 Subagent Boundary Discipline** — REQ-ATR-015 makes this central; hook scripts MUST NOT invoke AskUserQuestion (orchestrator-only contract)
- **B4 Frontmatter Canonical Schema** — all NEW rule file frontmatter must use canonical fields
- **B6 spec-lint Heading Regulation** — `## Out of Scope` h2 alone is insufficient; use h3 sub-section (already applied in spec.md §C.3)
- **B8 Working Tree Hygiene** — do NOT touch runtime files (`.moai/state/`, `.moai/harness/`, `.moai/cache/`)
- **B9 Git Commit + Push (Hybrid Trunk)** — manager-develop performs main-direct push for Tier S/M; Tier L OR `--pr` flag routes through manager-git per REQ-ATR-020
- **B10 PRESERVE List Invariant** — only the §C.1 scope inventory may be modified; no drive-by edits to predecessor SPECs (except SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 frontmatter for supersedence per REQ-ATR-006)
- **B11 AskUserQuestion 금지 (Subagent Boundary)** — manager-develop must return blocker reports, not call AskUserQuestion
- **B12 Sync-phase CHANGELOG discipline** (manager-docs only) — duplicate-entry detection via `grep -c '<SPEC-ID>' CHANGELOG.md` before append

### §B.3 Pre-flight Self-verification (orchestrator before M1 spawn)

```bash
# 1. Working tree clean baseline
git rev-parse HEAD && git status --porcelain | wc -l
# Expected: <plan-phase commit SHA>, 0 lines

# 2. Divergence check (parallel session race detection)
git fetch origin main && git rev-list --count --left-right origin/main...HEAD
# Expected: "0 0" (clean)

# 3. SPEC artifact inventory verification (this SPEC)
ls -la .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/
# Expected: 5 files (spec.md, plan.md, acceptance.md, design.md, research.md)

# 4. Predecessor SPEC current state verification
grep '^status:' .moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md
# Expected: status: draft (at plan-phase iter-2; will be transitioned to superseded by M6)

# 5. Agent file inventory baseline
ls .claude/agents/core/ .claude/agents/expert/ .claude/agents/meta/ .claude/agents/agency/ 2>&1 | wc -l
# Expected: ≥17 (current 17-agent catalog before M3 archive)

# 6. Workflow skill inventory baseline
ls .claude/skills/moai/workflows/*.md
# Expected: plan.md, run.md, sync.md, fix.md, loop.md, coverage.md, mx.md, codemaps.md, clean.md, design.md, feedback.md, review.md, e2e.md

# 7. Template mirror baseline
diff -q .claude/agents/ internal/template/templates/.claude/agents/ 2>&1 | head -20
# Expected: list of any current drift (will be reconciled in M8)
```

---

## §C — Scope

### §C.1 In-scope file inventory (run-phase scope)

This subsection enumerates the run-phase modification scope. See spec.md §C.1 for the corresponding domain narrative.

**Tier 1 — Retained agent files (7 MoAI-custom files, frontmatter + minor body refinement)**:
- `.claude/agents/core/manager-spec.md` (this file is the agent currently executing; refinement to add manager-strategy planning role absorption)
- `.claude/agents/core/manager-develop.md` (cycle_type=autofix mode addition)
- `.claude/agents/core/manager-docs.md` (manager-project absorption notes)
- `.claude/agents/core/manager-git.md` (Tier L OR `--pr` flag usage clarification)
- `.claude/agents/meta/plan-auditor.md` (frontmatter + NOT-for)
- `.claude/agents/meta/evaluator-active.md` (frontmatter + harness-level invocation surface)
- `.claude/agents/meta/builder-harness.md` (frontmatter + NOT-for)

(Anthropic built-in `Explore` is not a MoAI file; documented via cross-reference in `.claude/rules/moai/development/agent-patterns.md`)

**Tier 2 — Archived agent files (12 files to be moved to .moai/backups/ then deleted from active paths)**:
- `.claude/agents/core/manager-strategy.md`
- `.claude/agents/core/manager-quality.md`
- `.claude/agents/core/manager-brain.md`
- `.claude/agents/core/manager-project.md`
- `.claude/agents/meta/claude-code-guide.md`
- `.claude/agents/agency/researcher.md` (if present — verify in M3 pre-flight)
- `.claude/agents/expert/expert-backend.md`
- `.claude/agents/expert/expert-frontend.md`
- `.claude/agents/expert/expert-security.md`
- `.claude/agents/expert/expert-devops.md`
- `.claude/agents/expert/expert-performance.md`
- `.claude/agents/expert/expert-refactoring.md`

**Tier 3 — Workflow router skills (3 files, owner-declaration + hook-reference replacement)**:
- `.claude/skills/moai/workflows/plan.md`
- `.claude/skills/moai/workflows/run.md`
- `.claude/skills/moai/workflows/sync.md`

**Tier 4 — Rule files (~10 files: 2 NEW + 8 existing-modify)**:
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` (NEW)
- `.claude/rules/moai/workflow/spec-workflow.md`
- `.claude/rules/moai/development/agent-patterns.md`
- `.claude/rules/moai/development/agent-authoring.md`
- `.claude/rules/moai/development/manager-develop-prompt-template.md`
- `.claude/rules/moai/development/spec-frontmatter-schema.md`
- `.claude/rules/moai/core/agent-common-protocol.md`
- `CLAUDE.md` (§4 Agent Catalog)
- `CLAUDE.local.md` (§23 Hybrid Trunk + manager-git PR doctrine reconciliation)

**Tier 5 — Hook scripts (3 NEW files)**:
- `.claude/hooks/moai/status-transition-ownership.sh` (PostToolUse)
- `.claude/hooks/moai/sync-phase-quality-gate.sh` (Stop)
- `.claude/hooks/moai/team-ac-verify.sh` (TaskCompleted, dormant by default per REQ-ATR-013 capability gate)

**Tier 6 — Predecessor SPEC supersedence (1 file frontmatter modification ONLY)**:
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` (frontmatter `status` + `updated` + HISTORY v0.1.2 entry; body content untouched per L48 SSOT)

**Tier 7 — Documentation (1 file)**:
- `NOTICE.md` (REQ-ATR-019 Anthropic 2026 attribution citation)

**Tier 8 — Template mirror parity (~13 files byte-for-byte sync)**:
- `internal/template/templates/.claude/agents/{core,meta,expert,agency}/*.md` — 7 retained agents synced + 12 archived agents REMOVED from template
- `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync}.md`
- `internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW mirror)
- `internal/template/templates/.claude/rules/moai/workflow/archived-agent-rejection.md` (NEW mirror)
- `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
- `internal/template/templates/.claude/rules/moai/development/agent-patterns.md`
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md`
- `internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md`
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
- `internal/template/templates/.claude/hooks/moai/{status-transition-ownership,sync-phase-quality-gate,team-ac-verify}.sh`
- `internal/template/templates/CLAUDE.md` (the template's own CLAUDE.md — separate from project-root CLAUDE.md)
- (CLAUDE.local.md is NOT mirrored — local-only per §22 [HARD] settings.local.json Separation analog)

**Tier 9 — SPEC artifacts (this SPEC, 5 files — plan-phase scope, already authored at M0)**:
- `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/{spec,plan,acceptance,design,research}.md`

Estimated total run-phase scope: **~50-60 files** (7 retain + 12 archive + 3 workflow skills + 10 rule files + 3 hook scripts + 1 SPEC frontmatter + NOTICE.md + ~13 template mirrors + 5 SPEC artifacts). LOC equivalent: ~1500-2000 lines markdown edits + ~300-500 lines new shell hook scripts.

### §C.2 PRESERVE list (DO NOT MODIFY)

Per `manager-develop-prompt-template.md` §1.D Constraints, the following are explicit PRESERVE-list entries (out-of-scope):

- All files in `.moai/specs/SPEC-V3R6-*` predecessor directories EXCEPT `SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` frontmatter-only per REQ-ATR-006
- All files in `internal/spec/` (Go-side spec-lint engine — out-of-scope; no Go code changes)
- All files in `.claude/rules/moai/languages/` (16-language neutrality contract)
- All retained agent **body content** beyond frontmatter + minor scoped refinement per §F.1 (out-of-scope)
- All workflow skill **body content** beyond phase-owner-declaration + hook-reference-replacement scope per §F.2 (out-of-scope)
- Sub-skill modules under `.claude/skills/moai/workflows/{plan,run}/*.md` (out-of-scope)
- All files in `.moai/state/`, `.moai/cache/`, `.moai/logs/`, `.moai/harness/usage-log.jsonl` (runtime-managed)
- All files in `.moai/research/` outside this SPEC directory
- `CHANGELOG.md` body modifications outside this SPEC's sync-phase entry
- 88 pre-v3 SPEC bodies under `.moai/specs/SPEC-*` (excluding V3R6 family) — strict L48

### §C.3 Out of Scope

#### §C.3.1 Out of Scope — Go source code modifications
No Go source file modifications (`internal/*.go`, `cmd/*.go`, `pkg/*.go`) are within scope. ARCHIVED_AGENT_REJECTED runtime enforcement (REQ-ATR-016) is implemented at the orchestrator-discipline layer + rule documentation; deferred Go-side enforcement is a follow-up SPEC.

#### §C.3.2 Out of Scope — Retained agent body re-authoring
Beyond frontmatter + minor scoped refinement, full body re-authoring of the 7 MoAI-custom retained agents is deferred to `SPEC-V3R6-AGENT-BODY-COMPRESSION-001` (future).

#### §C.3.3 Out of Scope — Workflow skill body re-authoring
Beyond phase-owner-declaration + hook-reference-replacement, full body re-authoring of the 23 workflow + 10 sub-skill bodies (13K LOC Audit 2 finding) is deferred to `SPEC-V3R6-WORKFLOW-SKILL-REBUILD-001` (future).

---

## §D — Milestone Decomposition

### §D.1 Plan-phase M0 — SPEC artifact authoring (CURRENT)

**Owner**: manager-spec
**Scope**: 5 files (spec.md + plan.md + acceptance.md + design.md + research.md) + 1 frontmatter-only modification of predecessor SPEC (REQ-ATR-006 deferred to M6 run-phase per Status Transition Ownership Matrix)
**REQs covered**: all 20 REQ-ATR-XXX (declarative)
**Self-verification**:
- 12 canonical frontmatter fields present in spec.md
- GEARS notation ≥80% in §D Requirements
- ≥20 AC-ATR-XXX in acceptance.md
- Traceability 100% (every REQ-ATR has ≥1 AC-ATR)
- design.md §B Target Architecture + §D Hook Architecture + §E Karpathy mapping present
- research.md §H Audit 3 synthesis appended
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): plan-phase artifacts (Tier L Section A-G, 5 artifacts; supersedes SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001)`

### §D.2 Phase 0.5 — plan-auditor verdict

**Owner**: plan-auditor (subagent invocation by orchestrator)
**Threshold**: Tier L PASS ≥ 0.85 (mandatory), skip-eligible ≥ 0.90 (preferred)
**MP-1 Scope clarity target**: ≥ 0.85 (5-artifact set covering Audit 3 architectural pivot)
**MP-2 GEARS/EARS rigor target**: ≥ 0.85 (20 REQs with mandated GEARS pattern distribution)
**MP-3 Traceability target**: ≥ 0.85 (REQ-ATR↔AC-ATR mapping at 100%)
**MP-4 Risk-mitigation pairing target**: ≥ 0.85 (8 risks in spec.md §G each paired with mitigation strategy)
**Self-audit estimate**: ~0.88-0.91 PASS (Tier L MARGINAL to skip-eligible — explicit citations of Anthropic verbatim audit findings strengthen MP-1 and MP-2)
**Retry contract**: max 3 iterations; iter(N+1) < iter(N) triggers STOP signal per plan-auditor.md

### §D.3 Phase 0.95 — Mode Selection (NEW)

**Owner**: orchestrator (autonomous decision)
**Inputs**: §C.1 scope (~50-60 files), 5 domain categories (agents, workflow skills, rules, hook scripts, template mirrors)
**Expected mode**: **sequential single-spawn manager-develop with Tier L Section A-E template + 8 milestone breakdown** per Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"). Markdown + shell-script-only scope; no genuine concurrent implementation benefit from parallel multi-spawn.
**Decision logged at**: `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/progress.md` § Mode Selection
**Rationale to record**: "Tier L + multi-domain (5 categories) + scope ~50-60 files satisfies REQ-ATR-017 Compound preconditions. However, the task is markdown + shell-script-only (no Go code), and Anthropic Finding A4 advises coding tasks remain single-agent. Final decision: sequential single-spawn manager-develop with Tier L Section A-E template + 8 milestone breakdown. Agent Teams mode NOT selected because (a) workflow.team.enabled likely false per default project config, (b) CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS env var typically unset, (c) coding-task parallelism caveat applies."

### §D.4 Run-phase M1 — Agent retention frontmatter refinement

**Owner**: manager-develop (single-spawn, cycle_type=ddd per Audit 1 SRP — preserving existing agent body content while refining frontmatter)
**Scope**: 7 retained agent files (frontmatter + minor scoped body refinement)
- `.claude/agents/core/{manager-spec, manager-develop, manager-docs, manager-git}.md`
- `.claude/agents/meta/{plan-auditor, evaluator-active, builder-harness}.md`
**REQs covered**: REQ-ATR-001, REQ-ATR-002, REQ-ATR-003, REQ-ATR-004
**Refinements**:
- manager-spec: description absorbs "manager-strategy planning role" mention + NOT-for clause "not for run-phase code implementation"
- manager-develop: description adds `cycle_type=autofix` mode + NOT-for clause "not for SPEC body authoring"
- manager-docs: description absorbs "manager-project absorption" notes + NOT-for clause "not for SPEC body authoring"
- manager-git: description clarifies "Tier L OR --pr flag" usage + NOT-for clause "not invoked for Tier S/M default Hybrid Trunk"
- plan-auditor / evaluator-active / builder-harness: NOT-for clauses for boundary discipline
**Body refinement constraint**: each retained agent's body line count ≤ 500 (REQ-ATR-002) — if any exceeds, defer compression to future SPEC and document the overage
**Self-verification**: `wc -l .claude/agents/core/*.md .claude/agents/meta/*.md` ≤ 500 per file; `grep -l "NOT-for:" .claude/agents/core/*.md .claude/agents/meta/*.md` = 7 files
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M1 — 7 retained agent frontmatter refinement (Anthropic 2026 alignment)`

### §D.5 Run-phase M2 — Workflow router skill phase-owner declarations

**Owner**: manager-develop
**Scope**: 3 workflow router skill files
- `.claude/skills/moai/workflows/plan.md` — phase owner declarations (Explore + manager-spec); remove manager-strategy chain references; preserve existing Skill router invocation discipline
- `.claude/skills/moai/workflows/run.md` — phase owner declarations (manager-develop single-agent per Finding A4); remove manager-strategy chain; add Phase 0.95 Mode Selection logging hook reference per REQ-ATR-008; add cycle_type=autofix mode reference per REQ-ATR-012
- `.claude/skills/moai/workflows/sync.md` — phase owner declarations (manager-docs single-agent); replace manager-quality/expert-security/manager-develop-coverage spawn references with Stop hook references per REQ-ATR-009
**REQs covered**: REQ-ATR-001 (manager-strategy chain removal), REQ-ATR-007 (Skill router invocation), REQ-ATR-008 (Phase 0.95 Mode Selection logging), REQ-ATR-012 (cycle_type=autofix mode reference)
**Body modification constraint**: changes restricted to phase-owner declarations + hook-reference replacement. Sub-skill modules and 13K LOC body content out-of-scope per §F.2.
**Self-verification**: `grep -c "manager-strategy" .claude/skills/moai/workflows/*.md` = 0 in plan/run/sync (allowing references in archive documentation may exist in other files)
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M2 — workflow router skill phase-owner declarations (manager-strategy chain removed)`

### §D.6 Run-phase M3 — Agent archive to .moai/backups/

**Owner**: manager-develop
**Scope**: 12 phantom and domain-expert agent files
**Operation sequence (REQ-ATR-005)**:
1. Create archive directory: `mkdir -p .moai/backups/agent-archive-2026-05-25/{core,meta,expert,agency}`
2. Move (preserve git history via `git mv`):
   - `git mv .claude/agents/core/manager-{strategy,quality,brain,project}.md .moai/backups/agent-archive-2026-05-25/core/`
   - `git mv .claude/agents/meta/claude-code-guide.md .moai/backups/agent-archive-2026-05-25/meta/`
   - `git mv .claude/agents/agency/researcher.md .moai/backups/agent-archive-2026-05-25/agency/` (verify existence first; some projects may have already removed)
   - `git mv .claude/agents/expert/expert-{backend,frontend,security,devops,performance,refactoring}.md .moai/backups/agent-archive-2026-05-25/expert/`
3. Mirror archive in template: `git mv internal/template/templates/.claude/agents/core/manager-{strategy,quality,brain,project}.md ` ... (analogous moves for template mirror; template archive may go to a sibling `internal/template/templates/.moai/backups/` or be omitted from template entirely per REQ-ATR-001 "zero phantom agents" — manager-develop decides per template-deployment semantics in M8)
4. Create `.moai/backups/agent-archive-2026-05-25/README.md` documenting archive date + per-agent rationale + replacement pattern
**REQs covered**: REQ-ATR-005
**Self-verification**: `ls .moai/backups/agent-archive-2026-05-25/**/*.md | wc -l` = 12 (if researcher.md exists; else 11) + README.md
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M3 — archive 12 phantom and domain-expert agents`

### §D.7 Run-phase M4 — Hook scripts authoring (3 NEW files)

**Owner**: manager-develop
**Scope**: 3 NEW shell scripts
- `.claude/hooks/moai/status-transition-ownership.sh` (PostToolUse) — verify Write/Edit invoker matches Status Transition Ownership Matrix
- `.claude/hooks/moai/sync-phase-quality-gate.sh` (Stop) — verify lint + test + coverage delta on sync-phase commits per REQ-ATR-009 + REQ-ATR-014 dependency manifest audit
- `.claude/hooks/moai/team-ac-verify.sh` (TaskCompleted, team mode only — dormant by default)
**REQs covered**: REQ-ATR-009, REQ-ATR-014
**Each hook script structure**:
- Shebang: `#!/bin/bash`
- Read stdin JSON from Claude Code per hook contract
- Parse relevant fields (tool_name, tool_input, etc.)
- Execute verification commands; capture exit codes
- Output structured JSON to stdout (per Anthropic hook contract); exit 0 (continue) or 2 (block)
- Support `--skip-hook` flag for opt-out logging to `.moai/logs/hook-skip.log`
**Self-verification**:
- `bash -n .claude/hooks/moai/*.sh` (syntax check) exit 0 for all 3
- `shellcheck .claude/hooks/moai/*.sh` exit 0 (or skip if shellcheck unavailable)
- Per-hook golden-output test under `.moai/tests/hooks/` (if test framework available; else manual smoke test documented)
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M4 — 3 NEW hook scripts (PostToolUse + Stop + TaskCompleted)`

### §D.8 Run-phase M5 — Rule files updates

**Owner**: manager-develop
**Scope**: ~10 rule files (2 NEW + 8 existing-modify)
**NEW files**:
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — 5-mode autonomous selection rule per REQ-ATR-008. Documents decision tree (trivial / background / agent-team / parallel / sub-agent) with explicit thresholds from Audit 3 Finding A4.
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — ARCHIVED_AGENT_REJECTED error specification + migration table per REQ-ATR-016. Documents per-archived-agent replacement pattern (12 entries).
**Existing modifications**:
- `.claude/rules/moai/workflow/spec-workflow.md` — Phase Overview table agent column updates (8 retained); archive table cross-reference
- `.claude/rules/moai/development/agent-patterns.md` — domain-expert spawn-prompt pattern documentation (Finding A5); Explore canonical read-only-investigation agent reference; deprecate manager-strategy chain pattern
- `.claude/rules/moai/development/agent-authoring.md` — per-spawn specialization vs static agent file decision tree
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — cycle_type=autofix reference
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Status Transition Ownership Matrix archived-agent reference purge (12 archived agents removed from any owner rows; only retained 8 in matrix)
- `.claude/rules/moai/core/agent-common-protocol.md` — orchestrator obligations hook invocation surface
- `.claude/rules/moai/workflow/git-workflow-doctrine.md` — Hybrid Trunk vs PR-based policy explicit consolidation per REQ-ATR-020
- `.claude/skills/moai-foundation-core/SKILL.md` — Agent Catalog 17→8 reference (if any agent count cited; likely 1-2 references to purge)
**REQs covered**: REQ-ATR-007, REQ-ATR-008, REQ-ATR-012, REQ-ATR-016, REQ-ATR-020
**Self-verification**: `ls .claude/rules/moai/workflow/orchestration-mode-selection.md .claude/rules/moai/workflow/archived-agent-rejection.md` (both NEW files exist); `grep -c "manager-strategy" .claude/rules/moai/development/agent-patterns.md` = 0 (or "deprecated" annotation only)
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M5 — rule files updates (2 NEW + 8 modified)`

### §D.9 Run-phase M6 — Predecessor SPEC supersedence

**Owner**: orchestrator-mediated re-delegation to manager-spec (Status Transition Ownership Matrix `* → superseded` owner) — OR manager-develop with mid-run authority per the D-NEW-1 inline-fix pattern (orchestrator decision)
**Scope**: 1 file frontmatter-only modification
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` — frontmatter `status: <current> → superseded`, `updated: 2026-05-25`, HISTORY v0.1.2 entry citing architectural-pivot rationale
**REQs covered**: REQ-ATR-006
**Constraint**: body content untouched per L48 SSOT discipline. Only YAML frontmatter + HISTORY table.
**HISTORY v0.1.2 entry text**: `Superseded by SPEC-V3R6-AGENT-TEAM-REBUILD-001 — Audit 3 findings (catalog inflation 17→5-8 + hierarchical fiction "subagents cannot spawn subagents" + 12 phantom agents 0-invoc) require fundamental architecture pivot beyond original 17-phase-restoration scope.`
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M6 — supersedes SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (frontmatter status: superseded)`

### §D.10 Run-phase M7 — CLAUDE.md + CLAUDE.local.md doctrine updates

**Owner**: manager-develop
**Scope**: 3 doctrine files + NOTICE.md
- `CLAUDE.md` §4 (Agent Catalog) — 17→8 retention list update; archive table cross-reference; remove `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, 6 `expert-*` from the catalog enumeration; add 8 retained list + Anthropic built-in `Explore`
- `CLAUDE.local.md` §23 — manager-git PR doctrine reconciliation per REQ-ATR-020 (Tier S/M = main-direct, Tier L OR `--pr` = PR via manager-git)
- `CLAUDE.local.md` §19 (AskUserQuestion Enforcement) — cross-reference REQ-ATR-015 (GATE-2 mandatory restoration on autonomous-flow skip detection)
- `NOTICE.md` — REQ-ATR-019 Anthropic 2026 attribution citation (verbatim Findings A1-A6 quotes + source URLs)
**REQs covered**: REQ-ATR-001 (catalog list), REQ-ATR-015 (AskUserQuestion mandatory restoration), REQ-ATR-019 (NOTICE.md attribution), REQ-ATR-020 (manager-git PR doctrine reconciliation)
**Self-verification**: `grep -A 20 "## 4. Agent Catalog" CLAUDE.md | grep -c "manager-strategy"` = 0
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M7 — CLAUDE.md catalog + CLAUDE.local.md doctrine + NOTICE.md attribution`

### §D.11 Run-phase M8 — Template-First parity check + verification batch

**Owner**: manager-develop
**Scope**: ~13 template mirror files
**Operations**:
1. For each modified local file under `.claude/`, `cp` (or `git mv` for archives) to the corresponding `internal/template/templates/.claude/` path
2. For each archived agent, ensure the template mirror also has the archive applied (template mirror archive — if template mirror should not include archived files, simply delete from template; if archive directory should also be mirrored, copy `.moai/backups/agent-archive-2026-05-25/` to `internal/template/templates/.moai/backups/` — decision per `moai update` namespace protection contract in `CLAUDE.local.md` §24)
3. Run `make build` to regenerate `internal/template/embedded.go`
4. Verify `diff -r .claude/agents/ internal/template/templates/.claude/agents/` = empty (or only intended drift documented)
5. Verify `diff -r .claude/skills/moai/workflows/ internal/template/templates/.claude/skills/moai/workflows/` = empty
6. Verify `diff -r .claude/rules/moai/ internal/template/templates/.claude/rules/moai/` = empty (within scope of this SPEC's rule file modifications)
7. Verify `diff -r .claude/hooks/moai/ internal/template/templates/.claude/hooks/moai/` = empty (3 NEW hook scripts present in both)
**REQs covered**: REQ-ATR-018
**7-item Trust-but-verify verification batch** (orchestrator-side, parallel):
```bash
# 1. Full test suite (Go) — expected pass since no Go changes
go test ./...
# 2. Coverage report (baseline)
go test -coverprofile=cover.out ./internal/template/...
# 3. Subagent-boundary grep (C-HRA-008)
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ | grep -v "^[^:]*:[0-9]*:[ \t]*#"
# 4. Archived-agent reference scan
grep -rn 'subagent_type.*manager-strategy\|subagent_type.*manager-quality\|subagent_type.*expert-' .claude/ internal/template/ 2>&1 | grep -v ".moai/backups/" | grep -v "archived"
# 5. CLI smoke check
go run ./cmd/moai --version
# 6. Template embedded regeneration
make build && git diff --stat internal/template/embedded.go
# 7. Lint baseline
golangci-lint run --timeout=2m 2>&1 | tail -10
```
**Commit**: `feat(SPEC-V3R6-AGENT-TEAM-REBUILD-001): M8 — Template-First parity + verification batch (all 8 milestones complete)`

---

## §E — Verification Strategy

### §E.1 Per-milestone 7-item Trust-but-verify

Each of M1-M8 includes a self-verification command block in the manager-develop completion report. The orchestrator independently verifies via the canonical 7-item parallel Bash batch (test suite + coverage + subagent boundary grep + sentinel scan + CLI smoke + template parity + lint) per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution.

### §E.2 Run-phase audit-ready signal

After M8 completion, write to `progress.md` § E.2 Run-phase Audit-Ready Signal:
- All 8 milestones completed with commit SHAs
- All 20 REQ-ATR mapped to passing ACs (acceptance.md §B Traceability Matrix)
- 7-item Trust-but-verify batch passed
- Template parity diff empty
- Archived 12 agents present in `.moai/backups/agent-archive-2026-05-25/`

### §E.3 Sync-phase audit-ready signal (manager-docs scope)

manager-docs writes to `progress.md` § E.4 Sync-phase Audit-Ready Signal:
- CHANGELOG.md entry added (REQ-ATR-019 attribution + Audit 3 alignment summary)
- All 4 frontmatter `status: in-progress → implemented` (spec/plan/acceptance/design + research)
- Sync hooks (Stop sync-phase quality gate) verified passing
- No PRESERVE-list violations

### §E.4 Mx-phase audit-ready signal (orchestrator scope, expected EVALUATE-SKIP)

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a (markdown-heavy SPEC, 0 .go files, 0 goroutines, 0 fan_in delta), Step C judgement expected EVALUATE-SKIP. Document skip rationale in `progress.md` § E.5 Mx-phase Audit-Ready Signal.

---

## §F — Cross-References

- spec.md §C.1 file inventory (run-phase scope basis)
- spec.md §D 20 REQ-ATR-XXX (per-milestone coverage)
- acceptance.md §A ≥20 AC-ATR-XXX (per-AC evidence commands)
- design.md §B Target Architecture (8-agent delegation graph)
- design.md §D Hook Architecture (M4 hook scripts implementation guidance)
- research.md §H Audit 3 synthesis (manager-develop run-phase context)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section A-E (M1-M8 delegation prompt template)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase 0.5 Plan Audit Gate + § SPEC Complexity Tier
- `.claude/rules/moai/workflow/session-handoff.md` § paste-ready resume (multi-session continuity for M1-M8 across Tier L wall-time)
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` (REQ-ATR-012 cycle_type=autofix mode reference)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (REQ-ATR-006 M6 supersedence ownership)
- `CLAUDE.local.md` §24 (`moai update` namespace protection contract — M8 template parity considerations)

---

Version: 0.1.0
Status: draft (plan-phase initial authoring)
Tier: L
