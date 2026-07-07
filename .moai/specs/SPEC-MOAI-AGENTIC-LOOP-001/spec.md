---
id: SPEC-MOAI-AGENTIC-LOOP-001
title: "/moai agentic completion loop — router requirement analysis, 6-mode wiring, autonomous plan-run-sync pipeline"
version: "0.1.1"
status: completed
created: 2026-07-07
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.x"
module: ".claude/skills/moai"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "orchestration, agentic-loop, mode-catalog, router, taxonomy, skills, template-parity"
depends_on: []
related_specs: [SPEC-AUTONOMY-RUN-GOAL-001]
---

# SPEC-MOAI-AGENTIC-LOOP-001 — /moai Agentic Completion Loop

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-07 | manager-spec | Initial draft — plan-phase authoring from orchestrator-supplied 10-gap survey (anchors re-verified against live tree 2026-07-07) |
| 0.1.1 | 2026-07-07 | manager-spec | plan-auditor iter-1 fixes (D1-D10): autonomous cycle restricted to run→sync→verify, parity-baseline attribution corrected (untracked artifact excluded), AC de-vacuization, baseline-SHA pinning, gate-sync reconciliation, config-key disambiguation |

---

## §A. Context & Problem

### §A.1 User directive (verbatim intent)

> "Like `/goal`, when a `/moai` request arrives, the orchestrator must analyze the requirement, plan a workflow that chains all relevant agents and skills, and drive an agentic loop until the request is fully completed — using sequential / parallel / team / dynamic-workflow orchestration as appropriate, executing plan → run → sync via skills."

### §A.2 Design goal

Transform the `/moai` meta-harness (entry: `.claude/skills/moai/SKILL.md` + `workflows/` tree) into a goal-directed autonomous pipeline with three pillars:

1. **Router-level requirement analysis** — every `/moai` invocation produces a requirement summary + a completion condition (machine-verifiable where possible, reusing `/goal` semantics from `.claude/rules/moai/workflow/goal-directive.md` — no parallel mechanism invented).
2. **Orchestration-shape selection** — the canonical 6-mode Phase 0.95 catalog (`.claude/rules/moai/workflow/orchestration-mode-selection.md` §A: trivial / background / agent-team / parallel / sub-agent / workflow) is wired into the dispatch path, replacing the current team/solo binary.
3. **Autonomous completion loop** — after the ONE mandatory Implementation Kickoff Approval (plan→run HUMAN GATE — [HARD], never bypassed), the pipeline drives run→sync→verify iterations autonomously until the completion condition is met (plan-phase re-entry is never an autonomous step — user-approved escalation only, per REQ-MAL-022/024), with iteration ceiling, no-progress detection, dark-flow guard, blocker escalation via orchestrator AskUserQuestion, and context-threshold handoff integration.

### §A.3 Evidence base — 10 verified gaps

Read-only survey findings, anchor-re-verified against the live tree at authoring time (2026-07-07; line numbers are content-token-anchored and MUST be re-verified at run-phase per line-drift discipline):

| # | Gap | Verified anchor |
|---|-----|-----------------|
| 1 | Requirement-analysis step absent at router; Socratic clarity lives only inside `plan/context-discovery.md` (Phase 0.3) | `SKILL.md` §Intent Router (L36-94) routes by first-word match only; Execution Directive (L229-305) has no analysis step |
| 2 | 6-mode catalog never referenced by SKILL.md / plan.md / sync.md / moai.md; SKILL.md carries only team/solo binary | `SKILL.md:46-52` "Execution Mode Flags" — `--team` / `--solo` / auto-select only; zero `orchestration-mode-selection` references in the four files |
| 3 | Three unreconciled mode taxonomies | (i) 6-mode catalog `orchestration-mode-selection.md` §A; (ii) `--mode` dispatch axis {autopilot, loop, team, pipeline} in `run/context-loading.md:35-67`; (iii) scale table {Fix, Focused, Standard, Full Pipeline, Team} in `run/phase-execution.md:214-235` |
| 4 | Stale "5-mode" label mixing the two axes that `orchestration-mode-selection.md` §G forbids mixing | `run.md:39` — "orchestrator autonomous 5-mode decision (autopilot / loop / team / pipeline / background)" mixes `--mode` axis values with a catalog value |
| 5 | No Mode 4 (parallel fan-out, 3-5 concurrent `Agent()`) operationalization in run bodies | Only `moai.md` Phase 0 spawns parallel read-only agents; `run.md` + `phase-execution.md` Phase 0.95 offer no parallel-subagent instruction |
| 6 | Mode 6 (dynamic workflow / ultracode) has zero operational entry in any `/moai` route | `dynamic-workflows.md` is documentation-only; no route body carries a launch procedure |
| 7 | Implementation Kickoff Approval named only in `run.md:118-127` (+ harness-build-entry.md); the default `moai.md` pipeline body does not surface it | `moai.md` Phase 11.5 is "Execution Mode Selection Gate" — no kickoff-approval mention |
| 8 | plan-auditor / sync-auditor gates not declared in `moai.md` pipeline body | `moai.md` Phases 1/1.5/2/3 name neither auditor as a pipeline step (sync-auditor appears only as a Phase 0 exploration agent) |
| 9 | run→sync chaining exists only in the default `moai.md` route; subcommand routes never auto-chain | `run/mode-orchestration.md` Phase 4: "User presented with next step options" — manual only |
| 10 | Auto-select thresholds (domains≥3 / files≥10 / score≥7) duplicated across 4 skill sites, one semantically drifted | `SKILL.md:50`, `moai.md:47`, `moai.md:190`, `moai.md:200` (+ machine values in `workflow.yaml` `auto_selection`); drift instance: `phase-execution.md:226` requires `score ≥ 7 AND --team flag` for Team Mode, contradicting the flag-free auto-select semantics of `SKILL.md:50` |

### §A.4 Baselines (measured at authoring time)

- Skill-tree mirror parity: markdown files 42/42 identical (git-tracked content). One UNTRACKED runtime artifact is excluded from the parity baseline: `internal/template/templates/.claude/skills/moai/workflows/.moai/` (statusline state leak; `git ls-files` returns empty for it). Parity comparisons therefore use `diff -rq --exclude=.moai` (verified clean, exit 0 at authoring time) or git-tracked file lists; run-phase M1 pre-flight deletes the junk directory (plan.md §F M1).
- `orchestration-mode-selection.md` mirror status: **MIRRORED-DIVERGED** (pre-existing divergence, not caused by this SPEC — run-phase pre-flight must characterize it before editing; REQ-MAL-042).
- `spec-workflow.md`, `dynamic-workflows.md`, `goal-directive.md`, `session-handoff.md` mirrors: identical.

---

## §B. Requirements (GEARS)

### §B.1 Group R — Router requirement analysis

- **REQ-MAL-001** — **When** a `/moai` invocation is routed (any subcommand or the default natural-language route), the router (`SKILL.md` Execution Directive) **shall** produce a requirement-analysis record before loading the workflow body, consisting of: (a) a 1-3 sentence requirement summary, (b) a completion condition, (c) a pipeline contract (`full-pipeline` | `single-phase`), and (d) an orchestration-shape pre-signal (input to Phase 0.95).

- **REQ-MAL-002** — **Where** the derived completion condition is machine-verifiable (test exit code, lint-clean state, grep count, bounded turn count), the router **shall** express it in `/goal`-compatible transcript-measurable form per `.claude/rules/moai/workflow/goal-directive.md` (one measurable end state + a stated check + a bound clause).

- **REQ-MAL-003** — The router **shall not** introduce a completion-condition evaluator mechanism parallel to `/goal`. **Where** `/goal` is unavailable (runtime < v2.1.139, or hooks disabled), the loop **shall** degrade to orchestrator per-turn verification of the same condition text (graceful degradation — no new machinery).

- **REQ-MAL-004** — **Where** the invocation matches a trivial-scope exemption (subcommands `feedback`, `gate`, `codemaps`, `sync` status mode, or a Stage-1-Clarify exception per `askuser-protocol.md` § Ambiguity Triggers and Exceptions), the router **shall** skip requirement-analysis emission and proceed directly.

- **REQ-MAL-005** — **While** requirement analysis leaves intent clarity below 100%, the orchestrator **shall** run the Socratic interview (per `askuser-protocol.md`) BEFORE deriving the completion condition — the completion condition encodes drained intent, never a guess.

### §B.2 Group O — Orchestration-shape selection

- **REQ-MAL-010** — The dispatch surfaces (`SKILL.md` "Execution Mode Flags" section + `workflows/moai.md` mode-selection prose) **shall** reference the Phase 0.95 6-mode catalog (`orchestration-mode-selection.md` §A) as the single mode SSOT, replacing the team/solo binary description. The `--team` / `--solo` flags **shall** remain as forced overrides mapping to Mode 3 / Mode 5 respectively.

- **REQ-MAL-011** — The rule `orchestration-mode-selection.md` **shall** carry a dispatch-axis crosswalk section mapping each `--mode` axis value (autopilot / loop / team / pipeline) and each Phase 0.95 scale-table label (Fix / Focused / Standard / Full Pipeline / Team) to catalog terms. The crosswalk **shall** preserve §G's two-axis separation — it documents correspondence; it does NOT merge the axes and introduces no `--mode workflow` flag value.

- **REQ-MAL-012** — **When** the `run.md` Phase Owners region carrying the stale "5-mode" label is edited, `run.md` **shall** describe Phase 0.95 as the 6-mode catalog decision with a pointer to the rule; the mixed-axis enumeration ("autopilot / loop / team / pipeline / background") **shall not** survive the edit.

- **REQ-MAL-013** — The complexity auto-select thresholds **shall** be stated numerically in exactly one prose SSOT — `orchestration-mode-selection.md` §B.1, which cites `.moai/config/sections/workflow.yaml` `auto_selection` as the machine-readable source. Every other touched surface (`SKILL.md` ×1 site, `moai.md` ×3 sites) **shall** carry a cross-reference instead of inline numbers.

- **REQ-MAL-014** — The run-phase bodies (`run.md` + `run/phase-execution.md` Phase 0.95) **shall** carry an operational Mode 4 entry: **While** pre-implementation work is research-heavy and multi-domain, the orchestrator **shall** spawn 3-5 concurrent read-only `Agent()` calls in a single turn for analysis fan-out; implementation itself remains Mode 5 per the Anthropic coding-task parallelism caveat.

- **REQ-MAL-015** — The dispatch path (`moai.md` + `run.md`) **shall** carry an operational Mode 6 entry documenting the launch procedure and its capability gate per `orchestration-mode-selection.md` §C.3 (Implementation Kickoff Approval passed + all preferences collected + scope ≥ ~30 files mechanical uniform transform + runtime ≥ v2.1.154 + workflows not disabled) and the `progress.md` §F logging obligation before launch.

- **REQ-MAL-016** — **When** `phase-execution.md`'s Phase 0.95 scale table is re-expressed in catalog terms, the semantic drift at the Team-Mode row (`score ≥ 7 AND --team flag`, contradicting the flag-free auto-select semantics) **shall** be reconciled to the catalog's Mode 3 capability gate (§C.1) — auto-select and forced-flag paths both resolve through the same gate.

### §B.3 Group L — Agentic completion loop

- **REQ-MAL-020** [HARD] — The pipeline **shall** present the Implementation Kickoff Approval human gate exactly once per pipeline entry, at the plan→run boundary. The agentic loop **shall** operate ONLY downstream of it, and a derived completion condition **shall never** authorize autonomous run-phase entry (preserves the kickoff-approval mandatory-restoration invariant and `orchestration-mode-selection.md` §H user-veto semantics). The gate binds per plan→run boundary crossing: an escalation-approved re-plan (REQ-MAL-024) opens a NEW plan→run boundary with its own single gate crossing — the loop never re-crosses the boundary autonomously.

- **REQ-MAL-021** — The `moai.md` pipeline body **shall** declare, as named pipeline steps: the plan-audit gate (plan-auditor), Implementation Kickoff Approval, Phase 0.95 mode selection (6-mode catalog), and the sync-audit gate (sync-auditor). These are currently implicit via sub-skill loading; after this SPEC they are explicit in the pipeline body.

- **REQ-MAL-022** — **While** a completion condition is active and unmet, **when** a pipeline phase (run or sync) completes, the orchestrator **shall** start the next loop iteration step (verify → next phase, or re-enter the failing RUN or SYNC phase) instead of returning control — until the condition is met, the iteration ceiling is reached, or an escalation fires. The autonomous iteration cycle is **run→sync→verify ONLY**: plan-phase re-entry is NEVER an autonomous loop step — it occurs solely via the REQ-MAL-024 escalation path with explicit user approval, and the revised plan re-crosses the Implementation Kickoff Approval gate before any run-phase re-entry. No reading of this requirement permits an autonomous mid-loop plan→run re-cross.

- **REQ-MAL-023** — The loop **shall** enforce a max-iteration ceiling on pipeline iterations (default 10), configurable via a documented `workflow.agentic_loop.max_iterations` key in `.moai/config/sections/workflow.yaml` (prose-read by the orchestrator; Go-side key registration is out of scope — see §E). The new `agentic_loop.max_iterations` key (pipeline-level iterations, default 10) is DISTINCT from the pre-existing `loop_prevention.max_iterations: 100` (per-operation / diagnostic fix-loop bound) — separate YAML blocks, separate semantics, no key collision.

- **REQ-MAL-024** — **When** the same failure signature (identical failing check + same error class) is observed in two consecutive loop iterations, the loop **shall** halt and escalate via a structured report; the orchestrator **shall** then run an AskUserQuestion round (continue with manual investigation / revert + re-plan / abort). **Where** the user selects revert + re-plan, the revised plan **shall** re-cross the Implementation Kickoff Approval gate before any run-phase re-entry (see REQ-MAL-020/022).

- **REQ-MAL-025** (dark-flow guard) — The loop **shall** surface a per-iteration visible report in the conversation (iteration #, phase executed, evidence delta, condition-evaluation result). Silent iterations are prohibited — every iteration's evidence rides the transcript. This is also what keeps the `/goal` evaluator functional (transcript-measurability per goal-directive.md).

- **REQ-MAL-026** — **When** a semantic failure (data race, deadlock, panic, test assertion failure) surfaces during any loop iteration, the loop **shall** clear the active completion condition and escalate immediately via AskUserQuestion — the loop **shall never** auto-fix a semantic failure (binds the existing ci-autofix-protocol invariant onto the pipeline-level loop).

- **REQ-MAL-027** — **While** context usage crosses the model-specific handoff threshold (`context-window-management.md` § Context Window Targets), the loop **shall** suspend at the current iteration boundary, persist state to `progress.md`, and emit the paste-ready resume message per `session-handoff.md`. The loop **shall not** start a new iteration past the threshold.

- **REQ-MAL-028** — Subagents and workflow agents operating inside the loop **shall never** prompt the user; all mid-loop user decisions ride structured blocker reports → orchestrator AskUserQuestion (the asymmetric boundary per `agent-common-protocol.md` § User Interaction Boundary, restated as binding on the loop).

- **REQ-MAL-029** — The pipeline-level agentic loop **shall** remain distinct from the Ralph engine (`/moai loop`): the Ralph engine remains the specialized diagnostic fix-loop and MAY be invoked by the agentic loop's verify step for mechanical convergence. `loop.md` **shall** carry a relationship paragraph naming both loops and their granularity difference. Folding the two is rejected (rationale: §D.1).

### §B.4 Group C — run→sync chaining policy

- **REQ-MAL-030** — **When** the pipeline contract is `full-pipeline` (default natural-language route, or the user selected full-pipeline scope at Implementation Kickoff Approval), run-phase completion **shall** auto-chain into sync-phase, announced in the transcript (no additional approval round — sync doc work is non-destructive; PR creation still follows Tier-based PR routing and its own gates). Auto-chain removes ONLY the extra approval at the run→sync phase boundary; the HUMAN GATEs preserved INSIDE the sync workflow (`gate-sync-1` pre-sync quality, `gate-sync-2` documentation scope) still fire unchanged within the chained sync phase.

- **REQ-MAL-031** — **When** the pipeline contract is `single-phase` (explicit `/moai run` or `/moai sync` invocation), phase completion **shall** surface the chain as the "(Recommended)" first option of the existing next-step AskUserQuestion. The chain **shall not** fire silently on single-phase contracts.

- **REQ-MAL-032** — **Where** the sync-audit gate returns FAIL/INCONCLUSIVE or the sync-phase quality gate blocks, the chain **shall** halt and escalate — the loop **shall never** auto-complete past a failing gate (preserves `orchestration-mode-selection.md` §J.2).

### §B.5 Group N — Mirror parity & neutrality

- **REQ-MAL-040** [HARD] — Every touched file under `.claude/skills/moai/**` **shall** land byte-identically in `internal/template/templates/.claude/skills/moai/**` (baseline: full-tree parity at plan time — §A.4).

- **REQ-MAL-041** [HARD] — No skill, rule, or template file touched by this SPEC **shall** carry this SPEC's ID, REQ tokens, internal dates, or commit SHAs (template-neutrality per the template internal-content isolation doctrine). Only generic prose, mechanism descriptions, and permanent rule citations are permitted in shipped surfaces.

- **REQ-MAL-042** — **When** editing `orchestration-mode-selection.md` (pre-existing MIRRORED-DIVERGED state per §A.4), the run-phase **shall** first characterize the divergence (pre-flight diff), land this SPEC's delta in BOTH copies, and either fully re-sync the file or return a structured blocker report documenting the residual divergence — silent divergence growth is prohibited.

---

## §C. Hard Constraints (verbatim from delegation, encoded)

1. [HARD] Implementation Kickoff Approval preserved exactly once per pipeline entry; agentic loop operates ONLY post-kickoff; a `/goal`-style condition never authorizes autonomous run-phase entry (REQ-MAL-020).
2. [HARD] AskUserQuestion boundary: subagents and workflow agents never prompt; only the orchestrator bridges user decisions (REQ-MAL-028).
3. [HARD] Template-First + mirror parity: byte-identical mirror for every touched skill file; shipped files stay template-neutral (REQ-MAL-040/041).
4. [HARD] Taxonomy: the 6-mode catalog is the single mode SSOT; `--mode` axis and scale table get an explicit crosswalk (not a merge); thresholds stated once and cross-referenced (REQ-MAL-011/012/013/016).
5. Loop safety: iteration ceiling, no-progress detection, dark-flow guard, session-handoff integration (REQ-MAL-023..027).
6. Scope: markdown prose surfaces only — zero Go code changes (§E).

---

## §D. Design Decisions (with rationale)

### §D.1 `/moai loop` relationship — KEEP as specialized fix-loop (fold REJECTED)

The Ralph engine (`workflows/loop.md`) remains the specialized diagnostic fix-loop; the new agentic completion loop is a pipeline-level construct that MAY delegate to the Ralph engine during its verify step. Rationale:

1. **Different granularity** — the agentic loop iterates over PHASES (run→sync→verify; plan re-entry only via user-approved escalation per REQ-MAL-022); the Ralph engine iterates over DIAGNOSTICS (LSP/AST-grep/test/coverage) within a fix cycle. `goal-directive.md` already codifies this complementarity ("`/goal` and `/moai loop` are complementary, not competitors").
2. **Contract preservation** — `loop.md` carries a CI-audited alias contract (`/moai loop` ≡ `/moai run --mode loop`) plus snapshot/memory-pressure machinery; a fold would break the audit sentinel and orphan the resume machinery.
3. **Simplicity ladder** — reuse before rebuild: the verify step reuses the existing engine instead of duplicating a diagnostic loop inside the pipeline loop.

### §D.2 Chaining policy — "pipeline contract" derived at requirement analysis

Chaining is governed by the pipeline contract recorded at router requirement analysis (REQ-MAL-001c): `full-pipeline` → auto-chain (announced); `single-phase` → recommended-option chaining. No new CLI flag is introduced. Rationale: the user's route choice already encodes chain intent (natural-language default = "complete my request end-to-end"; explicit `/moai run SPEC-X` = "do this phase"); a flag would duplicate that signal.

### §D.3 Threshold SSOT placement

Prose SSOT = `orchestration-mode-selection.md` §B.1 (already documents the input parameters); machine SSOT = `workflow.yaml` `auto_selection` (`min_complexity_score: 7`, `min_domains_for_team: 3`, `min_files_for_team: 10` — verified live). Skill surfaces become pointers. The prose SSOT gains one sentence citing the machine SSOT so both stay linked. Rationale: the rule file is the only surface that already explains WHAT the thresholds decide; numbers without decision context invite drift (gap #10's observed drift instance).

### §D.4 Completion-condition mechanism — reuse `/goal`, degrade gracefully

The condition is authored in `/goal`-compatible form (transcript-measurable, bounded) and set via `/goal` when the runtime supports it; otherwise the orchestrator evaluates the identical condition text per-turn. No parallel evaluator, no new hook (REQ-MAL-002/003). This mirrors the existing run-phase `ac_converge` pattern in `run.md` § Run-phase Autonomy — the agentic loop generalizes that wiring from run-phase-only to pipeline-level.

### §D.5 Requirement-analysis placement — SKILL.md Execution Directive

A new step between the current Step 2.5 (Project Documentation Check) and Step 3 (Load Workflow Details), so it runs after routing (the route determines the trivial-exemption set, REQ-MAL-004) and before workflow-body token spend. Trivial routes skip it — no ceremony tax on `feedback`/`gate`/status queries.

### §D.6 Tier L artifact-count deviation (disclosed)

Tier L canonically carries 5 plan-phase artifacts. This SPEC ships 4 (spec/plan/acceptance/progress): the research content is the orchestrator-supplied 10-gap survey, re-verified and embedded as §A.3 (a separate research.md would duplicate it verbatim); design decisions are consolidated in §D + plan.md §D. Flagged for plan-auditor review as a conscious deviation.

---

## §E. Exclusions (Out of Scope)

The following are explicitly out of scope for this SPEC. Reports/analysis of existing behavior belong in `.moai/reports/`; the items below are deferred or rejected scope.

### Out of Scope — Go code changes

- No Go source changes of any kind (`internal/`, `cmd/`, `pkg/`). `git diff --stat -- '*.go'` must be empty at close.
- `workflow.agentic_loop.max_iterations` Go-side config-key registration (`internal/config`) — recorded as follow-up; this SPEC documents the key and reads it as prose only.
- Hook-based mechanical enforcement of loop invariants (e.g., a Stop-hook iteration counter) — follow-up SPEC territory.

### Out of Scope — Agent body changes

- `.claude/agents/**/*.md` are untouched. The loop is orchestrator-side prose; agent contracts (manager-spec/develop/docs, auditors) are consumed as-is.

### Out of Scope — Team-mode orchestration bodies

- `team/plan.md`, `team/run.md`, `team/sync.md` internals are untouched — they ARE the Mode 3 execution bodies and already function; only the dispatch prose that selects them changes.

### Out of Scope — CLAUDE.md and always-loaded surface edits

- `CLAUDE.md` §15 carries a 5th inline threshold triple; replacing it with a pointer is deferred (always-loaded-diet governance owns CLAUDE.md edits). Recorded as follow-up.

### Out of Scope — New CLI flags and runtime mechanisms

- No `--chain`, no `--mode workflow`, no new sentinels. Existing `--mode` axis value set {autopilot, loop, team, pipeline} and its sentinels (`MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY`) are preserved verbatim.
- No AskUserQuestion timeout/auto-select mechanism (remains deferred per `orchestration-mode-selection.md` §H.2).
- No replacement or reimplementation of `/goal`, dynamic workflows, or Agent Teams primitives.

---

## §F. Cross-References

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — 6-mode catalog SSOT (§A), thresholds (§B.1), Mode 6 gate (§C.3), two-axis warning (§G), IGGDA (§H-§J)
- `.claude/rules/moai/workflow/goal-directive.md` — `/goal` semantics reused for completion conditions
- `.claude/rules/moai/workflow/dynamic-workflows.md` — Mode 6 primitive (16-concurrent / 1000-total; no mid-run user input)
- `.claude/rules/moai/workflow/session-handoff.md` + `context-window-management.md` — loop suspension triggers
- `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary — asymmetric boundary
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — status transition ownership
- SPEC-AUTONOMY-RUN-GOAL-001 — prior art: run-phase `/goal ac_converge` wiring this SPEC generalizes
