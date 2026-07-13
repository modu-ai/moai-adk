---
id: SPEC-WORKFLOW-CACHE-OPT-001
title: "Workflow Bottleneck Phase 2 — shared diagnostic snapshot contract + gate merging + delegation relaxation + audit defect-lists + bookkeeping batching"
version: "0.1.0"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P1
phase: "v3.0.x"
module: "internal/verify (new), internal/goal, cmd/moai, .claude/skills/moai/workflows, internal/template/templates"
lifecycle: spec-anchored
tags: "workflow, bottleneck, snapshot, verification, gate-merging, delegation, audit, bookkeeping, prompt-cache"
era: V3R6
tier: L
related_specs: [SPEC-GOAL-ENGINE-001]
---

# SPEC-WORKFLOW-CACHE-OPT-001 — Workflow Bottleneck Phase 2

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Initial plan-phase authoring. Phase 2 of the /moai workflow bottleneck effort (Phase 1 = `cache-aware-execution.md` rule, commit 32b89fb9d). Source analysis: 3-lens parallel audit of SKILL.md + 17 workflow files, 55 bottlenecks (12 HIGH) converging on 5 axes. |

> **Effort lineage**: Phase 1 (shipped) added 5 prompt-cache-aware ordering directives as doctrine (`.claude/rules/moai/workflow/cache-aware-execution.md`). Phase 2 (this SPEC) is the structural work: an evidence-consumption contract between verification layers plus four classes of workflow-body diet.

## §A — Context and Motivation (WHY)

A 3-lens parallel analysis of `.claude/skills/moai/SKILL.md` + 17 `workflows/*.md` files found 55 bottlenecks (12 HIGH), converging on 5 axes:

1. **Duplicate verification stacks** — `go test` (and its per-language equivalents) runs 4-6× per SPEC lifecycle: `/moai gate`, run Phase 2.5/2.75/2.8, sync Phase 0 + 0.5 + 0.7, loop Step 3 + Step 1.5, the `stop-goal` evaluator's Tier-1 mechanical conditions (`internal/goal/evaluate.go` re-executes shell commands each turn-end), and the orchestrator verification batch. Each `/moai loop` iteration re-runs full diagnostics (10-20 min). Root cause: **no evidence-consumption contract between layers** — every layer re-observes because no layer can trust (attributably) what a sibling already observed.
2. **User gate stacking** — 7-10 AskUserQuestion round-trips per default pipeline, 10+ for `/moai project`. Implementation Kickoff Approval (moai.md Step 11.3) and the Execution Mode Selection Gate (Step 11.5) are adjacent blocking rounds; project Stage B asks its 4 extended axes as up to 4 separate calls; `/moai feedback` uses 3 sequential rounds. Each blocking wait > 5 min expires the prompt cache over the full accumulated prefix (cache-aware-execution.md directive 1).
3. **[HARD] delegation forced for mechanical work** — fix/loop spawn agents even for Level-1 mechanical fixes (import sort, formatting); codemaps chains 3 spawns; clean chains 4+ spawns; a typical pipeline spawns 8-15 agents.
4. **Audit repetition excess** — plan-auditor ≤3 full re-audit iterations + annotation cycle ≤6 + run Phase 0.5 re-audit; project doc generation retries up to 3 full regeneration+re-audit loops.
5. **Per-iteration bookkeeping fixed cost** — 3 Task-tool calls per issue in loop/fix, MX_TAG_REPORT emitted every iteration, review's 4 perspectives run serially, the secrets scan re-reads full git history on every review.

## §B — Scope (WHAT this SPEC delivers)

| # | Axis | Deliverable class |
|---|------|-------------------|
| 1 | Shared diagnostic snapshot contract (**M1, biggest lever**) | Go engine (schema + key + freshness + store) + `moai verify` CLI surface + `stop-goal` evaluator integration + doctrine injection into gate / run Phase 2.75 / sync Phase 0 / loop Steps 1·3·1.5 |
| 2 | Gate merging | Doctrine edits: moai.md (Kickoff + 11.5 co-location), project Stage B, harness-build-entry, feedback, full-pipeline completion close |
| 3 | Delegation relaxation | Doctrine edits: fix.md, loop.md, codemaps.md, clean.md, mx.md |
| 4 | Audit improvement | Doctrine edits: Tier S audit default inversion, defect-list return protocol, project doc retry cap |
| 5 | Bookkeeping batching | Doctrine edits: loop bookkeeping batch, MX_TAG_REPORT cadence, review Mode-4 parallel lenses, incremental secrets scan |

Every modified `.claude/**` shipped file lands byte-equal in `internal/template/templates/` (Template-First; all 16 target files verified MIRROR-BYTE-EQ at plan time).

## §C — GEARS Requirements

### §C.1 Shared Diagnostic Snapshot Contract (REQ-SNAP)

- **REQ-SNAP-001** (Ubiquitous): The diagnostic snapshot shall be a structured JSON artifact under `.moai/state/verify/` recording, per executed check: a check identifier, the exact command executed, its exit code, parsed result counts (error count, warning count, test pass/fail, coverage percentage — where applicable), a capture timestamp, execution duration, and the snapshot key. The `conditions` result block shall remain read-compatible with the existing loop-verdict `conditions` shape (`zero_errors` / `error_count` / `tests_pass` / `coverage_threshold` / `coverage_actual` / `zero_warnings`) so existing loop-verdict readers keep working.
- **REQ-SNAP-002** (Ubiquitous): The snapshot key shall bind the snapshot to the exact working-tree state it measured — composed from the HEAD commit SHA plus a working-tree content hash that changes whenever any tracked file's content changes (dirty tree included). A snapshot recorded against one tree state shall be distinguishable from every other tree state.
- **REQ-SNAP-003** (Event-driven): **When** a consumer requests snapshot reuse, the freshness engine shall accept the snapshot only when the stored key equals the key recomputed from the current working tree AND the layer's freshness-acceptance rule holds; on any mismatch the engine shall report the snapshot stale and the consumer shall re-execute the check instead of reusing.
- **REQ-SNAP-004** (Ubiquitous): The snapshot engine shall be implemented in Go with a dedicated package owning the schema, key computation, freshness check, and atomic store; a `moai verify` CLI verb group shall expose record/check operations so doctrine-level (markdown workflow) consumers invoke the engine mechanically rather than hand-writing JSON. The CLI verb shall be registered in the root command tree.
- **REQ-SNAP-005** (Capability gate): **Where** a fresh snapshot covers a check category that `/moai gate` would run, the gate workflow shall reuse the recorded result — citing snapshot path + key + original command + recorded exit code as evidence — instead of re-executing that check, and shall record its own fresh executions back into the snapshot.
- **REQ-SNAP-006** (Capability gate): **Where** a fresh snapshot exists at run Phase 2.75 (Pre-Review Quality Gate), the run workflow shall consume it for covered check categories and shall record Phase 2.75's own executions into the snapshot for downstream consumers.
- **REQ-SNAP-007** (Capability gate): **Where** a fresh snapshot covers the full-test-suite check at sync Phase 0 (`gate-sync-1` pre-sync quality), the sync workflow shall consume it instead of re-running the full suite, citing the snapshot as evidence in the gate report.
- **REQ-SNAP-008** (State-driven): **While** a `/moai loop` sweep is active, Step 3 (Parallel Diagnostics) shall record its parsed diagnostics as a snapshot write in the shared schema, and Step 1's mechanical completion predicate shall re-evaluate from that persisted snapshot — formalizing the existing Step-8 persistence into the shared contract with a mechanical writer.
- **REQ-SNAP-009** (Unwanted): The loop Step 1.5 Independent Final Pass shall NOT consume a snapshot produced by the same loop run — the independence of the final pass is the carve-out that keeps success-exit evidence non-self-referential. Step 1.5's own fresh gate invocation MAY write a new snapshot for downstream consumers (e.g., sync Phase 0).
- **REQ-SNAP-010** (Capability gate): **Where** a Tier-1 mechanical condition's command matches a fresh snapshot entry, the `stop-goal` evaluator shall reuse the recorded exit code instead of re-executing the command, and the reuse (snapshot path + key) shall be recorded in the evaluator's verdict payload; **when** no fresh matching entry exists, the evaluator shall execute the command exactly as today.
- **REQ-SNAP-011** (Unwanted): A stale snapshot shall never be cited as verification evidence. Every reused result shall remain attributable per `verification-claim-integrity.md` §2 — the consumer cites the snapshot path, key, original command, and recorded output/exit code; the freshness rule is what makes the attribution valid (the snapshot IS the observed evidence).

### §C.2 Gate Merging (REQ-GATE)

- **REQ-GATE-001** (Event-driven): **When** the default pipeline reaches the plan→run boundary, the orchestrator shall present the Implementation Kickoff Approval question and the moai.md Step 11.5 execution-shape question within ONE AskUserQuestion call (multi-question, ≤4 questions per call). The Kickoff human gate itself remains mandatory, score-independent, and un-removable — this requirement co-locates questions, never removes the gate.
- **REQ-GATE-002** (Event-driven): **When** project interview Stage B runs, the workflow shall collect all remaining (un-inferred) extended axes in a single AskUserQuestion call carrying up to 4 questions, replacing the per-axis separate calls.
- **REQ-GATE-003** (Event-driven): **When** the harness natural-language entry reaches 100% profile clarity, the final-round harness-generation proposal and the build approval gate shall be presented as one AskUserQuestion round instead of two sequential rounds.
- **REQ-GATE-004** (Event-driven): **When** a full-pipeline contract completes successfully with no genuine pending decision, the completion report shall close with a clean completion statement and NO manufactured next-step question (the askuser-protocol "close with NO question" clause becomes the full-pipeline default). Single-phase contract completions keep their "(Recommended)" next-step chain question unchanged.
- **REQ-GATE-005** (Event-driven): **When** `/moai feedback` collects type, title, and description, it shall do so in one AskUserQuestion round (multi-question) instead of 3 sequential rounds.
- **REQ-GATE-006** (Unwanted): Gate merging shall not reduce the information collected — merged rounds carry the same questions, options, and description quality as the rounds they replace; only the number of blocking round-trips changes.

### §C.3 Delegation Relaxation (REQ-DELEG)

- **REQ-DELEG-001** (Capability gate): **Where** a `/moai fix` issue is classified Level 1 (import sorting, whitespace, formatting), the orchestrator shall execute the fix directly via the language's formatter command without an Agent() spawn; the [HARD] delegation mandate is re-scoped to Level 2 and above.
- **REQ-DELEG-002** (Capability gate): **Where** a `/moai loop` Step-6 fix task is classified Level 1, the same orchestrator-direct formatter execution shall apply.
- **REQ-DELEG-003** (Ubiquitous): The codemaps workflow shall complete with at most 1 Agent() spawn (a single read-only exploration spawn; analysis and map generation performed orchestrator-direct from the exploration output plus deterministic tooling), replacing the current 3-spawn chain.
- **REQ-DELEG-004** (Ubiquitous): The clean workflow shall complete with at most 2 Agent() spawns — one combined analysis spawn (static analysis + usage graph, current Phases 1+2) and one combined removal+verification spawn (current Phases 4+5) — replacing the current 4-spawn chain.
- **REQ-DELEG-005** (Capability gate): **Where** the `/moai mx` pending tag-insertion set is fewer than 5 items, Pass 3 batch edit shall be performed orchestrator-direct without an Agent() spawn.
- **REQ-DELEG-006** (Unwanted): Delegation relaxation shall not alter any approval semantics — Level-3 fix approval, the clean removal-plan AskUserQuestion, and @MX:ANCHOR removal protections remain unchanged, and the Level-classification dispatch table remains a static (non-LLM-decided) mapping.

### §C.4 Audit Improvement (REQ-AUDIT)

- **REQ-AUDIT-001** (Capability gate): **Where** the SPEC tier is S, the plan-audit gate shall run exactly one audit pass by default — the iterative verdict re-execution loop defaults OFF for Tier S (a PASS verdict is final without the score-threshold re-run; FAIL/INCONCLUSIVE still halts and escalates). Tier M/L audit behavior and the Implementation Kickoff Approval gate are unchanged.
- **REQ-AUDIT-002** (Event-driven): **When** an auditor (plan-auditor or sync-auditor) returns a FAIL verdict, the verdict shall carry a structured defect-list (finding id, artifact/file + location, severity, required fix); the orchestrator shall route fixes directly (orchestrator-direct edit or a single re-delegation) and the confirming re-audit shall be scoped to the enumerated defect delta rather than a from-scratch full re-audit — within the existing iteration ceilings.
- **REQ-AUDIT-003** (Ubiquitous): The project doc-generation independent-audit retry loop shall cap at 1 retry (currently up to 3 full regeneration+re-audit iterations before escalation); the escalation AskUserQuestion fires after the single retry fails.
- **REQ-AUDIT-004** (Unwanted): The delta re-check shall not transfer verdict authority — binding PASS/FAIL judgment stays with the auditors; the delta scope reduces re-audit cost, never substitutes an orchestrator self-assessment for an auditor verdict.

### §C.5 Bookkeeping Batching (REQ-BOOK)

- **REQ-BOOK-001** (State-driven): **While** a `/moai loop` iteration is active, Task-tool bookkeeping (TaskCreate for discovered issues, TaskUpdate status transitions) shall consolidate into at most one batched turn per iteration — parallel Task-tool calls in a single turn, with aggregate tasks per file/group — replacing the 3-calls-per-issue cadence.
- **REQ-BOOK-002** (Event-driven): **When** a `/moai loop` sweep exits (success, ceiling, or interruption), the MX_TAG_REPORT shall be emitted once, aggregated across all iterations; per-iteration tag ADD/REMOVE/UPDATE actions still apply — only the report cadence changes.
- **REQ-BOOK-003** (Ubiquitous): The review workflow's 4 perspectives (Security / Performance / Quality / UX) shall execute as a Mode-4 parallel read-only fan-out (≤4 concurrent read-only judges within the 3-5 ceiling) instead of a single sequential pass; the sync-auditor remains the binding synthesis/verdict owner.
- **REQ-BOOK-004** (Capability gate): **Where** a prior secrets-scan checkpoint SHA is recorded under `.moai/state/`, the review secrets scan shall scan incrementally (`<last-sha>..HEAD` plus working tree) and update the checkpoint; **where** no checkpoint exists (first run) or an explicit full-scan flag is passed, the full `--all` history scan shall run.
- **REQ-BOOK-005** (Unwanted): Batching shall not drop records — the same task entries, tag actions, and findings are produced; only call cadence and report timing change.

### §C.6 Invariant Guards (REQ-GUARD)

- **REQ-GUARD-001** (Ubiquitous): The Implementation Kickoff Approval human gate shall remain mandatory and score-independent at the plan→run boundary in every surface this SPEC touches — gate merging co-locates questions into fewer rounds and never removes, weakens, or auto-bypasses the gate.
- **REQ-GUARD-002** (Ubiquitous): The AskUserQuestion channel monopoly shall remain unchanged — every merged round still rides AskUserQuestion; no merged flow degrades to free-form prose questions.
- **REQ-GUARD-003** (Ubiquitous): The verification-claim-integrity invariant shall remain unchanged — snapshot reuse is valid evidence attribution only under the freshness rule (REQ-SNAP-003/011); no surface edited by this SPEC permits an unobserved verification claim.
- **REQ-GUARD-004** (Ubiquitous): Every modified `.claude/**` shipped file shall land byte-equal in `internal/template/templates/` within the same milestone, with `make build` and the template test suite green (Template-First).

## §D — Exclusions (What NOT to Build)

The following are explicitly out of scope for this SPEC.

### Out of Scope — sync-phase-quality-gate.sh Stop hook rewiring

- The `.claude/hooks/moai/sync-phase-quality-gate.sh` Stop hook keeps its own independent lint+test execution; wiring the shell hook into the snapshot contract is deferred (shell-side JSON key computation is a separate risk surface).
- The hook's exit-code semantics and `MOAI_SYNC_GATE_BLOCKING` opt-in are untouched.

### Out of Scope — plan-auditor / annotation-cycle ceiling changes for Tier M/L

- Tier M/L plan-audit iteration ceilings (≤3) and the Phase 1.5 annotation cycle (≤6) keep their current bounds; only Tier S default inversion and the defect-list/delta protocol change.

### Out of Scope — cadence-bridge and CI-watch surfaces

- `.claude/rules/moai/workflow/cadence-bridge.md` recipes and the `moai-workflow-ci-loop` CI watch path are unchanged; scheduled runs keep their read-only invariant and do not consume or produce snapshots in this SPEC.

### Out of Scope — prompt-cache doctrine changes

- Phase 1's `cache-aware-execution.md` directives are consumed as motivation, not re-edited; no new cache-ordering directives are added here.

### Out of Scope — cross-session snapshot sharing

- The snapshot contract is single-checkout scoped. Multi-session/multi-worktree snapshot sharing, distributed invalidation, and remote caches are not built.

### Out of Scope — /moai review verdict semantics

- `/moai review` stays read-only/report-only; the Mode-4 parallel lens change (REQ-BOOK-003) alters execution shape only, never adds fixes or a new verdict class.

## §E — Dependencies and Follow-ups

- **Builds on**: `SPEC-GOAL-ENGINE-001` (the `internal/goal` engine whose `stop-goal` evaluator gains snapshot consumption in REQ-SNAP-010); the loop-verdict JSON schema (`loop.md` § Remaining-Issue Persistence) as the seed shape for the snapshot `conditions` block.
- **Related doctrine (cite, not modify)**: `verification-claim-integrity.md` (attribution invariant), `askuser-protocol.md` (channel + completion-report discipline), `orchestration-mode-selection.md` (Mode 4 ceiling), `cache-aware-execution.md` (Phase 1 motivation).
- **Follow-up candidates (not this SPEC)**: shell-hook snapshot integration; multi-worktree snapshot sharing; Tier M audit-default recalibration informed by post-landing measurements.
