---
id: SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001
title: "SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs (audit F1 + F12 resolution)"
version: "0.1.1"
status: completed
created: 2026-05-24
updated: 2026-05-25
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: ".claude/agents/core"
lifecycle: spec-anchored
tags: "agent-ownership, soc, manager-spec, manager-develop, manager-docs, status-transition, schema, audit-tier-2, anthropic-best-practice"
---

# SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 — SPEC artifact ownership realignment across manager-spec / manager-develop / manager-docs (audit F1 + F12 resolution)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-24 | GOOS행님 | Initial creation (plan-phase). Tier 2 SPEC derived from `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F1 + F12. Audits Anthropic Best Practice Category #7 (DRI ownership) at the agent-artifact ownership granularity. Tier S minimal Section A-E variant (precedent: IVB-001 / SARM-001 / TMC-001 / TMD-001 / SIV-001 — 5/5 Tier S minimal 1-pass cohort). |
| 0.1.1 | 2026-05-25 | orchestrator | L60 retroactive Mx-phase backfill — `status: implemented → completed`. progress.md §E.5 Mx-phase Audit-Ready Signal (L162-216) was already authored with SKIP-justified judgment but `mx_commit_sha` frontmatter field never backfilled (L60 chicken-and-egg) and spec.md status drift persisted (L67 manager-docs scope-creep). Cross-file consistency restored. Body 변경 없음. Original sync `11abb9a30`; this commit (post-`d74095e75`)는 atomic close terminator로 작동, mx 본문은 progress.md에서 self-reference. |

## §A. Why this SPEC

### §A.1 Problem statement — SPEC artifact ownership is documented in narrative but not bound in agent frontmatter

The 3-phase MoAI SPEC workflow (`/moai plan` → `/moai run` → `/moai sync`) assigns clear logical ownership to three core manager agents:

- **Phase 1 (plan)** — `manager-spec` creates `spec.md` + `plan.md` + `acceptance.md` + `progress.md` in `.moai/specs/SPEC-{ID}/` and emits `status: draft` in every artifact's frontmatter
- **Phase 2 (run)** — `manager-develop` reads `spec.md` + `plan.md` + `acceptance.md`, executes the M1..Mn milestones, and updates `progress.md` `§Run-phase Evidence` + `§Run-phase Audit-Ready Signal`. The frontmatter status transition `draft → in-progress` (and ultimately `→ implemented` on M-final commit) is documented in CLAUDE.md §5 narrative ONLY — NOT bound in `.claude/agents/core/manager-develop.md` frontmatter or in any hook
- **Phase 3 (sync)** — `manager-docs` reads `spec.md` + `progress.md`, generates the PR with `CHANGELOG.md` `[Unreleased]` entry, updates `progress.md` `§Sync-phase Audit-Ready Signal`, and performs the `in-progress → implemented` frontmatter status transition across all 4 SPEC artifacts

This narrative-only documentation produced a measurable cost: in the TMD-001 sync precedent (`009e68c5d`, 2026-05-24), `manager-docs` modified **5 files** including `spec.md` `§B.1 In-scope` (scope expanded from 5 to 6 files documenting the A3c cascade follow-up), `plan.md` `§A.2 EXTEND` (matching update), `acceptance.md` `§D.4` (cascade follow-on row), and `progress.md` `§E.2..E.5`. The `spec.md` body edit by `manager-docs` is structurally questionable: `spec.md` is the **canonical SSOT** (L48 lesson) authored by `manager-spec`, but `manager-docs`'s agent definition declares its scope as "documentation sync" and uses default `model: haiku`. The audit report (`.moai/research/anthropic-best-practices-2026-05-24.md` §3 F1) classifies this as **P1 Critical**.

### §A.2 Evidence — three converging signals

#### §A.2.1 Audit signal (F1 + F12 P1 Critical)

From `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F1 (verbatim summary):

- manager-spec NOT-FOR clause "documentation sync" is explicit in manager-docs.md
- However sync stage's manager-docs modifies SPEC artifact body (TMD-001 sync `spec.md §B.1` scope expansion) — violates user intuition that SPEC body is manager-spec's domain
- spec-frontmatter-schema.md defines 8 valid statuses but does NOT specify which agent owns each transition

F12 (P3 Improvement) is the downstream consequence: `manager-docs` defaults to `model: haiku`, but sync-phase scope includes SPEC body reasoning (TMD-001 precedent) — a capability mismatch (haiku is appropriate for `CHANGELOG.md` + `README.md` doc generation, not for `spec.md` semantic-content edit). F1 resolution auto-mitigates F12 because manager-docs scope reduces to CHANGELOG-only (no SPEC body modification).

#### §A.2.2 Operational signal (5 commits SPEC artifact frontmatter transitions per TMD-001)

| Commit | Author agent | Files touched | Frontmatter transitions |
|--------|-------------|---------------|------------------------|
| `b2a3a14e1` (plan) | manager-spec | 4 SPEC artifacts created | `(none) → draft` × 4 |
| `9fe1768e8` (run M1) | manager-develop | 6 files (5 SPEC scope + 1 catalog cascade) | `progress.md draft → implemented`; spec.md / plan.md / acceptance.md frontmatter status UNCHANGED at this commit (the audit-report assertion is incorrect here — see §A.2.3) |
| `009e68c5d` (sync) | manager-docs | 5 files (CHANGELOG.md + spec.md + plan.md + acceptance.md + progress.md) | All 4 SPEC artifact frontmatter `draft → implemented` performed atomically; spec.md `§B.1` body expanded; plan.md `§A.2 EXTEND` mirrored; acceptance.md `§D.4` cascade follow-on row added |
| `397875876` (Mx chore) | (orchestrator-direct, no agent) | progress.md `§E.5 Mx-phase Audit-Ready Signal` | None |

#### §A.2.3 Narrative-vs-binding gap (the actionable defect)

`CLAUDE.md` §5 ("SPEC-Based Workflow") describes the agent chain (`manager-spec → manager-strategy → expert-backend → ... → manager-docs`) and the status enum lives in `.claude/rules/moai/development/spec-frontmatter-schema.md` (`draft → planned → in-progress → implemented → completed | superseded | archived | rejected`). However, neither file specifies the **per-transition agent owner**. The schema is enforced at lint-time (any of the 8 enum values is accepted regardless of which agent performed the transition), and the only ownership signal is the manager agent's frontmatter `description:` field — which currently uses prose like "creates SPEC documents" (manager-spec) without explicit artifact ownership boundaries.

### §A.3 Cost of the gap (5 concrete failure modes)

1. **Capability mismatch risk** — `manager-docs` (default `model: haiku`) is invoked for SPEC body reasoning when sync-phase needs scope expansion. Haiku is cost-optimized for doc generation, not for `spec.md` semantic reasoning. This is a latent quality risk that has not yet manifested but will under realistic edge cases (e.g., a SPEC where sync-phase identifies a missed REQ that was implemented).
2. **Separation-of-Concerns (SoC) erosion** — `spec.md` is the canonical SSOT (L48), but lacks an enforced single-author rule. Future maintainers may extend the precedent (manager-docs editing spec.md), eroding the SSOT discipline.
3. **Ownership ambiguity for new maintainers** — A new agent contributor reading `manager-spec.md` + `manager-develop.md` + `manager-docs.md` frontmatter cannot determine "who owns this status transition?" from agent metadata alone; they must read CLAUDE.md §5 narrative and infer.
4. **Status-transition drift detection is impossible** — Without a transition_owner matrix, an automated check ("did manager-docs modify spec.md?") cannot be expressed declaratively; it would require parsing CLAUDE.md prose.
5. **F12 model-allocation issue persists** — Without F1 resolution, the workaround for F12 would be promoting manager-docs to `model: sonnet` (cost increase) instead of reducing its scope (architectural fix).

### §A.4 Why agent-frontmatter binding rather than hook-only enforcement

A PostToolUse hook on `spec.md` writes that validates the calling agent matches the expected owner (per transition) would catch violations at execution time but is **late detection** (after the violating Write succeeds). Frontmatter binding is **shift-left** (declarative ownership in the agent definition itself, so a contributor reading the agent file sees the contract). The two layers are complementary; this SPEC implements frontmatter binding + a schema-level transition matrix in `spec-frontmatter-schema.md` SSOT as the primary intervention. The hook layer is out-of-scope for this SPEC (deferred per §B.2; optional MAY in REQ-ARR-009).

### §A.5 Anthropic Best Practice alignment

Anthropic's "DRI ownership" Best Practice (Category #7 in the audit report §2.2) at the project level translates to **per-artifact ownership** at the agent granularity. Without explicit DRI, the burden of correctness falls on convention and code review — both of which are higher-friction than declarative ownership. This SPEC operationalizes Anthropic Best Practice #7 at the agent-artifact granularity for the 3 core managers.

## §B. Scope

### §B.1 In-scope (5 files, run-phase only)

**Plan-phase artifacts** (created by this SPEC's plan-phase, 4 files):

- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md` (this file)
- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md`
- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/acceptance.md`
- `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md`

**Run-phase artifacts** (modified by future `/moai run`, 5 files exactly):

| # | Path | Operation | Expected delta |
|---|------|-----------|---------------|
| 1 | `.claude/agents/core/manager-spec.md` (operational source) | frontmatter `description:` field + new `## SPEC Artifact Ownership` body section declaring ownership of `spec.md` + `plan.md` + `acceptance.md` body authoring AND emission of initial `status: draft` AND mid-flight artifact body adjustments (e.g., AC re-tightening per the D-NEW-1 inline-fix pattern from SIV-001 run-phase) | +30-50 LOC |
| 2 | `.claude/agents/core/manager-develop.md` (operational source) | frontmatter `description:` field + new `## SPEC Artifact Ownership` body section declaring ownership of `progress.md` `§Run-phase Evidence` + `§Run-phase Audit-Ready Signal` sections AND the `draft → in-progress` transition (on M1 commit start) AND optional cascade follow-ups WITHIN SPEC declared scope envelope (e.g., catalog hash regen per L53) | +30-50 LOC |
| 3 | `.claude/agents/core/manager-docs.md` (operational source) | frontmatter `description:` field + new `## SPEC Artifact Ownership` body section declaring ownership of `CHANGELOG.md` + `README.md` + docs-site sync AND the `in-progress → implemented` transition AND `progress.md` `§Sync-phase Audit-Ready Signal` ONLY. **EXCLUDES** SPEC body (spec.md / plan.md / acceptance.md) modifications — when sync-phase identifies that scope expanded mid-run, manager-docs MUST return a blocker report and the orchestrator re-delegates to manager-spec for the scope-doc update | +30-50 LOC |
| 4 | `internal/template/templates/.claude/agents/core/manager-spec.md` (template mirror) | byte-identical mirror of #1 per CLAUDE.local.md §2 [HARD] Template-First Rule | same as #1 |
| 5 | `internal/template/templates/.claude/agents/core/manager-develop.md` (template mirror) | byte-identical mirror of #2 | same as #2 |
| 6 | `internal/template/templates/.claude/agents/core/manager-docs.md` (template mirror) | byte-identical mirror of #3 | same as #3 |
| 7 | `.claude/rules/moai/development/spec-frontmatter-schema.md` | new `## Status Transition Ownership Matrix` section appended after the Status Enum section, enumerating each transition + owning agent + canonical commit subject pattern | +40-60 LOC |

Total scope: **7 files modified** in run-phase (3 agent operational sources + 3 agent template mirrors + 1 schema doc), **0 files created** in run-phase. Total run-phase LOC delta: ~130-210 LOC across 7 files — within Tier S envelope (≤300 LOC, ≤7 files acceptable for Section A-E minimal mirror-pair pattern per IVB-001/SARM-001/TMC-001/TMD-001/SIV-001 precedent). Net source-vs-mirror byte delta per agent: 0 (byte-identical pairs).

### §B.2 Out of Scope (deferred to follow-up SPECs or explicitly NOT done)

- **Hook-based enforcement** — PostToolUse hook validating "agent that performed Write on spec.md matches the expected owner per transition" is OPTIONAL MAY per REQ-ARR-009 (no AC coverage; would be a follow-up SPEC if desired)
- **Other manager agent scope edits** — `manager-quality`, `manager-strategy`, `manager-brain`, `manager-git`, `manager-project`, `manager-orchestrator` are NOT in scope. Only the 3 core SPEC-lifecycle managers (spec / develop / docs) are affected
- **Expert agent boundaries** — `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring` are invoked during run-phase via manager-develop delegation; their relationship to `progress.md` is NOT in scope (they do not own SPEC artifacts — they contribute implementation files)
- **Meta agent boundaries** — `builder-harness`, `evaluator-active`, `plan-auditor`, `claude-code-guide` are NOT in scope
- **CLAUDE.md §5 narrative update** — narrative may reference the new ownership matrix, but CLAUDE.md edit is OPTIONAL MAY (REQ-ARR-008); the canonical SSOT for ownership lives in the agent frontmatter + schema doc, not in CLAUDE.md
- **TMD-001 retroactive correction** — the TMD-001 sync precedent (`009e68c5d`) is the **motivating example** but is NOT corrected by this SPEC (it remains historical record per Lessons Protocol). The new policy applies forward from merge of this SPEC
- **Status enum extension** — the 8-value status enum (`draft → planned → in-progress → implemented → completed | superseded | archived | rejected`) is NOT extended. Only the transition_owner matrix is added
- **plan-auditor / evaluator-active ownership** — these meta agents READ artifacts (and produce reports) but do NOT modify artifact frontmatter; their scope is outside the transition matrix
- **F11 SIV-001 artifact frontmatter symmetry** — separate Tier 4 follow-up SPEC if desired

## §C. Decision rules

### §C.1 SSOT hierarchy

- `spec.md` is the canonical SSOT for REQ + AC. Every REQ-ARR ID is anchored here.
- `plan.md` is the derived implementation artifact (edit map + tier + milestone breakdown). When the edit map conflicts with spec.md REQ wording, **spec.md wins** (L48 discipline).
- `acceptance.md` is the canonical AC enumeration (PASS/FAIL gates). REQs Covered column links each AC back to its REQ-ARR anchor in spec.md.
- `progress.md` is the runtime evidence log (4-phase Lifecycle Status + Audit-Ready Signal).
- The new `## SPEC Artifact Ownership` sections added to the 3 manager agent bodies are the **operational SSOT** for ownership (consulted by future manager agents at session start).
- The new `## Status Transition Ownership Matrix` section in `spec-frontmatter-schema.md` is the **schema-level SSOT** for transition_owner mappings (consulted by lint enforcement if future SPECs add such enforcement).

### §C.2 Tier S minimal Section A-E justification

This SPEC qualifies for the Tier S minimal Section A-E variant per `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability:

- **Scope**: 7 files modified (3 agent operational sources + 3 mirror pairs + 1 schema doc), ~130-210 LOC total, all changes are declarative ownership-boundary text — no production code change, no API surface change, no behavior change in any runtime path. Per IVB-001 / SARM-001 / TMC-001 / TMD-001 precedent, 5-7 file mirror-pair scope sits within Tier S envelope (envelope override accepted when the additional files are mechanical mirror cp, which is the case here).
- **Risk**: low — all 3 agents continue to function as before; the new sections are advisory ownership boundaries that bind future manager turn behavior but do not change current invocation contracts. The only behavior change is that future manager-docs invocations on sync-phase will return a blocker report (rather than silently modifying spec.md) when the orchestrator routes a SPEC-body-edit task to manager-docs.
- **Verification**: 7 PASS gates (frontmatter mirror parity × 3 + lint clean × 3 + schema doc syntax PASS) plus 0 behavior regressions.

Per Section A-E variant precedent established by IVB-001 (`d3ed4727d`) + SARM-001 (`5e0dc6a9b`) + TMC-001 (`38a638d3c`) + TMD-001 (`397875876`) + SIV-001 (`0e103eacc`), Tier S minimal retains the 4-artifact form (spec/plan/acceptance/progress) for traceability rather than the Tier S strict 2-artifact form. This is a documented pattern (L33 in MEMORY.md), not a deviation.

### §C.3 Mx Step C disposition

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a:

- Mx Step C **SKIP** condition requires `.go` file count = 0 AND @MX tag count delta = 0
- This SPEC modifies **only `.md` files** (3 agent bodies × 2 mirrors + 1 schema doc = 7 `.md` files; 0 `.go` files)
- Mx Step C disposition: **SKIP-eligible** if @MX tag count delta = 0 across all 7 files; orchestrator verifies post-sync via `grep -c '@MX' <each file>` before/after diff
- `progress.md` `§Mx-phase Audit-Ready Signal` records `mx_disposition=SKIP` with rationale OR EVALUATE if @MX tag count drifts

### §C.4 Mirror parity discipline (CLAUDE.local.md §2 [HARD])

The 3 agent operational sources (`/Users/goos/MoAI/moai-adk-go/.claude/agents/core/manager-{spec,develop,docs}.md`) MUST be edited in lockstep with their template mirrors (`internal/template/templates/.claude/agents/core/manager-{spec,develop,docs}.md`). Editing only one half breaks the `TestRuleTemplateMirrorDrift` / `TestLateBranchTemplateMirror` invariant. Run-phase manager-develop edits both files in the same commit, using `diff -q source mirror` post-edit verification to confirm byte-equality.

### §C.5 Frontmatter binding choice (description vs new field)

The ownership declaration could be embedded as a new top-level frontmatter field (e.g., `owns_artifacts: [spec.md, plan.md]`) or as a structured `description:` field extension. This SPEC uses the **body-section approach** (new `## SPEC Artifact Ownership` section in the agent body) for three reasons:

1. The agent frontmatter is consumed by Claude Code's agent loader and adding new top-level fields requires schema extension; the body section is consumed by Claude as natural-language instruction during the agent turn
2. Body-section ownership declarations are more readable for new contributors and can include rationale + edge cases (e.g., "manager-spec MAY adjust spec.md mid-run for AC re-tightening per D-NEW-1 inline-fix pattern" — too long for a frontmatter array entry)
3. The schema-level `## Status Transition Ownership Matrix` in `spec-frontmatter-schema.md` provides the structured machine-readable form for any future lint enforcement; the agent body provides the human-readable contract

The frontmatter `description:` field is updated to a 1-sentence summary referring to the new body section (e.g., "Creates SPEC documents (spec.md, plan.md, acceptance.md); see §SPEC Artifact Ownership for boundaries").

## §D. Lint surface

- **spec.md MUST emit `✓ No findings`** when checked by `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md`. The Out-of-Scope section (§B.2 above) is canonical only in spec.md.
- **plan.md / acceptance.md / progress.md** MAY emit `MissingExclusions` ERROR when lint-checked individually. This is accepted derived-artifact lint surface per IVB-001/SARM-001/TMC-001/TMD-001/SIV-001 precedent (CI does not block on derived-artifact lint per `git-strategy.yaml` pattern). The canonical lint surface is `spec.md` only.
- Frontmatter canonical 12 fields enforced across all 4 plan-phase artifacts per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT. `tags:` MUST be CSV-string form (`tags: "a, b, c"`), not YAML array — array form causes `ParseFailure` in `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding.
- Frontmatter status: ALL 4 plan-phase artifacts use `status: draft` initially; `manager-develop` transitions `draft → in-progress` on M1 commit start (per new ownership matrix this SPEC defines); `manager-docs` transitions `in-progress → implemented` on sync commit (per new ownership matrix).

## §E. Sprint context

### §E.1 Sprint 8 entry position

| Phase | SPEC ID | Tier | Scope | Status |
|-------|---------|------|-------|--------|
| Sprint 7 entry | SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 | S minimal | 4-file mirror cleanup | complete `397875876` |
| Sprint 8 entry | SPEC-V3R6-SPEC-ID-VALIDATION-001 | S minimal | L51 manager-spec body pre-write regex self-check | run-complete `0e103eacc` (sync pending) |
| **Sprint 8 P2** | **SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001** | **S minimal** | **3-agent ownership realignment + schema matrix** | **draft (this SPEC)** |

This SPEC continues the Tier S minimal 1-pass cohort (6/6 if successful) and addresses Audit Tier 2 (F1 + F12 P1+P3 resolution). After this SPEC completes 4-phase lifecycle, the SIV-001 sync-phase that was deferred pending audit can proceed with the new ownership policy in effect.

### §E.2 Sequential dependency

This SPEC has a **soft sequence** with SIV-001 sync-phase (`/moai sync SPEC-V3R6-SPEC-ID-VALIDATION-001`). The audit-report execution sequence (`.moai/research/anthropic-best-practices-2026-05-24.md` §5 Steps 4-7) ordered Tier 2 SPEC plan-phase BEFORE SIV-001 sync-phase so that the new ownership policy is in effect when manager-docs runs the SIV-001 sync. This SPEC's plan-phase is currently being executed; run-phase will follow user decision; sync-phase will complete the lifecycle; then SIV-001 sync-phase can proceed with the new policy applied.

### §E.3 Post-merge follow-up

The audit report's 4-tier plan continues post-Tier-2:

- **Tier 3**: F3 (3 unused skills reconnect) + F9 (subdirectory CLAUDE.md introduction) + F13 (DRI ownership at project level — README.md mention)
- **Tier 4**: F6 (CHANGELOG.md 31 Unreleased headings consolidation) + F7 (~260 retired SPEC refs cleanup) + F8 (sunset.yaml dormant config) + F10 (4 stale `@MX:WARN` without `@MX:REASON`) + F11 (SIV-001 plan/acceptance/progress frontmatter backfill — note: SIV-001 already has 4 artifacts as of this writing, so F11 may already be resolved)
- F12 (manager-docs haiku vs sync-phase scope) is **automatically resolved** by this SPEC (F1 resolution shrinks manager-docs scope to CHANGELOG-only, eliminating the haiku-vs-spec-body-reasoning capability mismatch)

## §F. Requirements (EARS Format)

### REQ-ARR-001 (Ubiquitous, mandatory) — manager-spec body section

The `manager-spec` agent body **shall** include a section titled exactly `## SPEC Artifact Ownership` declaring:
- manager-spec owns the authoring of `spec.md`, `plan.md`, `acceptance.md` body content
- manager-spec emits the initial `status: draft` in all 4 plan-phase artifacts (spec/plan/acceptance/progress)
- manager-spec is the ONLY agent authorized to modify `spec.md` `§A..§F` body content (REQ wording, scope decisions, AC matrix structure)
- manager-spec MAY adjust `spec.md` / `plan.md` / `acceptance.md` mid-run for AC re-tightening when the orchestrator re-delegates per the D-NEW-1 inline-fix pattern (SIV-001 run-phase precedent) — but ONLY upon explicit orchestrator re-delegation, never as a side-effect of another agent's turn

The section MUST be present byte-identically in BOTH `.claude/agents/core/manager-spec.md` (operational source) and `internal/template/templates/.claude/agents/core/manager-spec.md` (template mirror).

### REQ-ARR-002 (Ubiquitous, mandatory) — manager-develop body section

The `manager-develop` agent body **shall** include a section titled exactly `## SPEC Artifact Ownership` declaring:
- manager-develop owns the authoring of `progress.md` `§Run-phase Evidence` table + `§Run-phase Audit-Ready Signal` YAML block
- manager-develop performs the frontmatter status transition `draft → in-progress` on the M1 commit start (all 4 plan-phase artifacts: spec / plan / acceptance / progress) and `in-progress → implemented` (or `→ completed` depending on workflow variant) on the M-final commit for `progress.md` (the other 3 artifacts wait for sync-phase per REQ-ARR-003)
- manager-develop MAY perform cascade follow-ups within the SPEC's declared scope envelope (per L46 attribution discipline) such as the A3c catalog hash regen pattern from TMD-001 (`397875876`)
- manager-develop MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content — when run-phase reveals a scope/REQ/AC inadequacy, manager-develop returns a structured blocker report and the orchestrator re-delegates to manager-spec for the scope-doc update

The section MUST be present byte-identically in BOTH `.claude/agents/core/manager-develop.md` (operational source) and `internal/template/templates/.claude/agents/core/manager-develop.md` (template mirror).

### REQ-ARR-003 (Ubiquitous, mandatory) — manager-docs body section

The `manager-docs` agent body **shall** include a section titled exactly `## SPEC Artifact Ownership` declaring:
- manager-docs owns the authoring of `CHANGELOG.md` `[Unreleased]` entries, `README.md` synchronization, docs-site (`adk.mo.ai.kr`) synchronization
- manager-docs owns the authoring of `progress.md` `§Sync-phase Audit-Ready Signal` YAML block ONLY
- manager-docs performs the frontmatter status transition `in-progress → implemented` on the sync commit for ALL 4 SPEC artifacts (spec / plan / acceptance / progress)
- manager-docs MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content. When sync-phase reveals that scope expanded mid-run (e.g., TMD-001 sync precedent where A3c cascade follow-up was discovered post-run and needed `§B.1` body update), manager-docs **returns a structured blocker report** and the orchestrator re-delegates to manager-spec for the scope-doc update before proceeding to CHANGELOG emission

The section MUST be present byte-identically in BOTH `.claude/agents/core/manager-docs.md` (operational source) and `internal/template/templates/.claude/agents/core/manager-docs.md` (template mirror).

### REQ-ARR-004 (Ubiquitous, mandatory) — schema doc transition matrix

`.claude/rules/moai/development/spec-frontmatter-schema.md` **shall** include a new section titled exactly `## Status Transition Ownership Matrix` enumerating, for each canonical status transition, the owning agent + commit subject pattern:

| Transition | Owning agent | Canonical commit subject pattern |
|------------|-------------|----------------------------------|
| `(none) → draft` | manager-spec | `feat(SPEC-{ID}): plan-phase artifacts ({tier} Section A-E, 4 artifacts)` |
| `draft → in-progress` | manager-develop (on M1 commit start) | `fix(SPEC-{ID})` or `feat(SPEC-{ID})` — first run-phase commit |
| `in-progress → implemented` | manager-docs (on sync commit) | `docs(SPEC-{ID}): sync-phase artifacts` or `chore(SPEC-{ID}): sync-phase artifacts` |
| `implemented → completed` | (terminal — set by orchestrator or manager-docs on Mx chore commit) | `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close` |
| `* → superseded` | manager-spec (when authoring the new SPEC) | `feat(SPEC-{NEW-ID}): supersedes SPEC-{OLD-ID}` |
| `* → archived` | manager-docs (administrative cleanup) | `chore(specs): archive SPEC-{ID}` |
| `* → rejected` | (orchestrator decision, recorded by manager-docs) | `chore(SPEC-{ID}): rejected per <rationale>` |

The matrix MUST be appended to `spec-frontmatter-schema.md` as a new section AFTER the existing `## Status Enum (8 values)` section.

### REQ-ARR-005 (Ubiquitous, mandatory) — mirror parity

The 3 agent operational source files (`/Users/goos/MoAI/moai-adk-go/.claude/agents/core/manager-{spec,develop,docs}.md`) and their template mirrors (`internal/template/templates/.claude/agents/core/manager-{spec,develop,docs}.md`) **shall** remain byte-identical after run-phase edits. Verification: `diff -q` between each source-mirror pair returns empty output (exit code 0) post-edit.

### REQ-ARR-006 (Unwanted Behavior, mandatory) — agent ownership crossing

`manager-docs` **shall not** modify `spec.md`, `plan.md`, or `acceptance.md` body content (frontmatter `updated:` field update on the `in-progress → implemented` transition is allowed and required per REQ-ARR-003; ALL other body modifications are forbidden). When sync-phase reveals a need to modify SPEC body content, manager-docs **shall** return a structured blocker report listing the required edit and the orchestrator re-delegates to manager-spec.

Symmetrically, `manager-develop` **shall not** modify `spec.md`, `plan.md`, or `acceptance.md` body content (frontmatter `updated:` field update on the `draft → in-progress` transition is allowed and required per REQ-ARR-002; ALL other body modifications are forbidden). When run-phase reveals a need to modify SPEC body content, manager-develop **shall** return a structured blocker report and the orchestrator re-delegates to manager-spec.

### REQ-ARR-007 (Event-Driven, mandatory) — frontmatter description update

**When** the agent frontmatter `description:` field is updated in the 3 agent files, the new value **shall** be a 1-2 sentence summary referring readers to the new `## SPEC Artifact Ownership` section. Example for manager-spec: `description: "Creates SPEC documents (spec.md, plan.md, acceptance.md) and emits status: draft. See §SPEC Artifact Ownership for artifact-level boundaries."`. The 1-sentence summary form keeps the frontmatter compact while pointing readers to the authoritative body section.

### REQ-ARR-008 (Optional, MAY) — CLAUDE.md narrative update

`CLAUDE.md` §5 (SPEC-Based Workflow) MAY be updated to reference the new transition matrix in `spec-frontmatter-schema.md`. This update is OPTIONAL; the canonical SSOT for transition ownership is `spec-frontmatter-schema.md` § Status Transition Ownership Matrix + the 3 agent body sections. **No AC coverage** for this REQ per §C.1 decision rule (matches L48 precedent: REQ-SARM-007 / REQ-TMC-007 / REQ-TMD-007 / REQ-SIV-005).

### REQ-ARR-009 (Optional, MAY) — PostToolUse hook enforcement

A future PostToolUse hook MAY be added to validate at execution time that the agent performing a Write on `spec.md` matches the expected owner per the transition matrix. This is a defense-in-depth layer; the primary intervention is the declarative ownership in agent bodies + schema doc. **No AC coverage** for this REQ — this is a deferred follow-up SPEC candidate.

## §G. Cross-References

### §G.1 Audit origin

- `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F1 (P1 Critical — SPEC artifact ownership ambiguity)
- `.moai/research/anthropic-best-practices-2026-05-24.md` §3 F12 (P3 Improvement — manager-docs haiku vs sync-phase scope; auto-resolved by F1)
- `.moai/research/anthropic-best-practices-2026-05-24.md` §4 Tier 2 (this SPEC scope definition)
- `.moai/research/anthropic-best-practices-2026-05-24.md` §2.2 Anthropic Best Practice Category #7 (DRI ownership)

### §G.2 Schema and agent SSOT

- `.claude/rules/moai/development/spec-frontmatter-schema.md` — canonical 12-field schema + 8-value status enum (will receive new `## Status Transition Ownership Matrix` section)
- `.claude/agents/core/manager-spec.md` — operational SSOT (will receive new `## SPEC Artifact Ownership` section)
- `.claude/agents/core/manager-develop.md` — operational SSOT (will receive new section)
- `.claude/agents/core/manager-docs.md` — operational SSOT (will receive new section)
- `internal/template/templates/.claude/agents/core/manager-{spec,develop,docs}.md` — template mirrors per CLAUDE.local.md §2 [HARD] Template-First Rule

### §G.3 Motivating precedent

- TMD-001 sync precedent `009e68c5d` — 5 files including `spec.md §B.1` scope expansion authored by manager-docs (`.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md` + `plan.md` + `acceptance.md` + `progress.md` + `CHANGELOG.md`). This is the **archetype problem** this SPEC addresses.
- SIV-001 run-phase D-NEW-1 inline fix pattern — `acceptance.md` AC re-tightening performed by manager-develop within the SAME SPEC run-phase scope; this is a valid pattern within the new ownership boundary because the re-tightening happened via orchestrator re-delegation to manager-spec (which then re-delegated to manager-develop with the updated AC). The new ownership matrix preserves this pattern explicitly.

### §G.4 Workflow narrative

- `CLAUDE.md` §5 — SPEC-Based Workflow narrative (currently the only documentation of agent-chain ordering; will OPTIONALLY reference the new transition matrix per REQ-ARR-008)
- `.claude/rules/moai/workflow/spec-workflow.md` § Phase Overview — 3-phase workflow with token budget + agent assignment (does NOT specify transition ownership currently)

### §G.5 Tier S minimal precedent (5/5 cohort)

- IVB-001 `d3ed4727d` — Sprint 2 P4.1
- SARM-001 `5e0dc6a9b` — Sprint 2 P4.2
- TMC-001 `38a638d3c` — Sprint 2 P4.3
- TMD-001 `397875876` — Sprint 7 entry
- SIV-001 `0e103eacc` — Sprint 8 entry (run-complete; sync pending the policy this SPEC defines)

### §G.6 Related lessons

- L33 (Tier S minimal sustained pattern — Section A-E variant)
- L46 (attribution discipline — cascade follow-ups within SPEC scope envelope)
- L48 (spec.md SSOT canonical — derived artifacts cannot override)
- L51 (SPEC ID regex pre-write self-check — sister SPEC pattern for shift-left detection)
- Behavior #2 (Manage Confusion Actively) + Behavior #5 (Maintain Scope Discipline) from `.claude/rules/moai/core/moai-constitution.md` — both directly support the ownership-boundary discipline this SPEC encodes
