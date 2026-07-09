---
id: SPEC-FIX-LOOPMAP-001
title: "Fix Phase-4 Mechanical Verification, Regression Guard, and Loop Taxonomy Mapping"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: S
depends_on: [SPEC-LOOP-VERDICT-CONTRACT-001]
tags: "fix, verification-independence, regression-guard, loop-taxonomy, escalation, one-shot-residue"
---

# SPEC-FIX-LOOPMAP-001 — Fix Phase-4 Mechanical Verification, Regression Guard, and Loop Taxonomy Mapping

## Relation to Epic Workflow-Reflex

This SPEC is **not** one of the 6 Epic Workflow-Reflex SPECs; it is the named follow-up that SPEC-LOOP-VERDICT-CONTRACT-001's `§Out of Scope — /moai fix Phase 4 scanner independence (L5)` explicitly deferred:

> "Audit finding L5 (fix.md Phase 4 re-runs the same scanners as the builder — partial mechanical independence only) is recorded in Traceability as provenance but deliberately NOT remediated here... L5 remains open audit debt for a follow-up SPEC."

- **Dependency**: `depends_on: [SPEC-LOOP-VERDICT-CONTRACT-001]` — this SPEC's residual-persistence requirement (REQ-FLM-003) consumes the `.moai/state/loop-verdict-<id>.json` schema that SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005 defines (plan.md §D D2). loop.md is shared surface between the two SPECs (see Constraints).
- **Tier**: S (minimal) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability

| Finding ID | Source | Summary |
|------------|--------|---------|
| L5 | SPEC-LOOP-VERDICT-CONTRACT-001 Traceability (MED) | `/moai fix` Phase 4 re-runs the same scanners as the builder — partial mechanical independence only. No regression guard against the Phase 1 baseline. No residue persistence on unresolved exit (fix.md is one-shot; residue evaporates in the transcript, mirroring loop.md's L8). No mapping of `/moai fix` (turn-based) vs `/moai loop` (goal-based) vs cadence recipes (time-based) vs CI watch (proactive) in either workflow's own documentation. |

---

## User Story

**As a** user running `/moai fix` (the one-shot Agentless diagnostic pipeline),
**I want** Phase 4 to prove its PASS claims with re-executed scanner evidence, guard against regressions introduced by its own fixes, persist any leftover residue instead of letting it evaporate in the transcript, and tell me clearly where `/moai fix` sits relative to `/moai loop` and the other cadence surfaces,
**so that** I can trust a fix report's claims, never silently accept a fix that broke something else, resume unresolved work via `/moai loop` without re-discovering it by hand, and pick the right entry point (fix vs loop vs cadence vs CI watch) without re-deriving the taxonomy myself.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read. Line numbers indicative; content anchors are authoritative.

### GAP-1 — Non-evidence-bearing Phase 4 verification (L5 core)

- **Measured source**: `.claude/skills/moai/workflows/fix.md` lines 189-193 — `## Phase 4: Verification` reads in full: *"- Re-run affected diagnostics on modified files\n- Confirm fixes resolved the targeted issues\n- Detect any regressions introduced by fixes"*. No claim/evidence pairing, no command citation, no output capture instruction.
- **Observed pattern**: "Confirm" and "Detect" are prose verbs with no mechanical binding — a builder following this text could self-assess without re-running anything, in direct tension with `verification-claim-integrity.md` §1.1 (no unobserved verification claim).

### GAP-2 — No regression guard against the Phase 1 baseline

- **Measured source**: fix.md line 191 scopes Phase 4 to *"affected diagnostics on modified files"* only — narrower than Phase 1's full parallel scan (lines 61-138, three scanners across the whole target).
- **Observed pattern**: A fix that resolves the targeted issue but introduces a NEW issue outside the modified-files scope, or in a file already scanned but not re-diffed against the Phase 1 baseline list, has no detection path. "Detect any regressions" (line 193) names the goal but not the mechanism (full re-scan + baseline diff).

### GAP-3 — Residue evaporates on one-shot exit

- **Measured source**: fix.md Phase 3 (lines 181-186) — Level 4 items are *"listed in report as manual action items"* only; § Snapshot Save/Resume (lines 257-271) persists scan state for `--resume`, not unresolved-issue residue as a re-enterable artifact; no cross-reference anywhere in fix.md to `/moai loop` as the escalation path for what Level 4/unresolved-error residue should become.
- **Observed pattern**: Same evaporation pattern SPEC-LOOP-VERDICT-CONTRACT-001 GAP-4 (L8) names for loop.md ceiling exits, but on the one-shot pipeline: nothing outlives the transcript.

### GAP-4 — No loop-taxonomy self-placement in either workflow

- **Measured source**: fix.md § Pipeline Contract (lines 45-59) states the Agentless localize→repair→validate contract and the one-shot Repeatability clause, but never names WHERE `/moai fix` sits relative to `/moai loop`, cadence recipes, or CI watch. loop.md § Relationship to the Pipeline-Level Agentic Completion Loop (lines 48-50) contrasts loop.md against the pipeline-level agentic loop only — not against fix.md, cadence, or CI watch.
- **Observed pattern**: A user choosing between `/moai fix` and `/moai loop` has no in-document decision aid; the taxonomy exists only in this SPEC's problem statement, not in the workflows themselves.

### Aggregate defect claim

**`/moai fix` Phase 4 asserts success it does not mechanically prove, cannot detect a regression outside its narrow re-scan, drops unresolved work on exit, and documents no position in the loop taxonomy it belongs to.** This SPEC makes Phase 4 evidence-bearing, widens its re-scan to a baseline-comparable regression guard, wires one-shot residue into the SPEC-LOOP-VERDICT-CONTRACT-001 persistence schema with a non-auto-invoking escalation recommendation, and adds a compact Loop Taxonomy Position section to both fix.md and loop.md.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the fix workflow", "the fix report", "the fix workflow's Phase 4"). No legacy `IF/THEN` modality.

### REQ-FLM-001 — Ubiquitous — mechanical verification evidence

The fix workflow's Phase 4 (Verification) SHALL derive every PASS claim exclusively from re-executed scanner exit codes and parsed issue counts — never from prose self-assessment — per `verification-claim-integrity.md` §1.1, and the fix report SHALL include claim/evidence rows pairing each claim with the command invoked and its verbatim-or-bounded-tail output.

### REQ-FLM-002 — Ubiquitous — regression guard against the Phase 1 baseline

The fix workflow's Phase 4 SHALL re-run the full Phase-1-equivalent scan (not only diagnostics scoped to modified files) and SHALL compare the resulting issue list against the Phase 1 baseline issue list; any issue present in the post-fix scan but absent from the baseline SHALL be treated as a regression — the fix that introduced it SHALL be reverted or explicitly reported as failed, and SHALL NOT be silently accepted.

### REQ-FLM-003 — Event-driven (When) — residue persistence to the loop-verdict schema

**When** the fix workflow exits with residual issues (Level 4 manual items, unresolved errors, or a Phase 4 regression-guard failure), the fix workflow SHALL persist the residue to `.moai/state/loop-verdict-<id>.json` using the schema SPEC-LOOP-VERDICT-CONTRACT-001 defines (plan.md §D D2: `spec_or_scope`, `exit_kind`, `iterations_used`, `ceiling_applied` + source, `conditions` final state, `remaining_issues[]`, `vci_report_ref`, `created_at`), setting `exit_kind: "one-shot-residue"` (extending the schema's `ceiling|manual-residue` enum with a third value naming the one-shot-pipeline exit path) and `iterations_used: 1`.

### REQ-FLM-004 — Event-driven (When) — fix-to-loop escalation recommendation without auto-entry

**When** the fix report is generated with non-empty residue, the fix workflow SHALL recommend `/moai loop` entry for re-fixable residue (or manual action for Level 4 items) as a suggestion inside the report — and SHALL NOT auto-invoke `/moai loop` or any other subcommand. The Pipeline Contract's Repeatability clause — *"Even when the parent invocation supplies `--mode loop`, the pipeline runs once per command invocation. Re-entry requires explicit user re-invocation."* — is PRESERVED verbatim and unweakened by this requirement.

### REQ-FLM-005 — Ubiquitous — Loop Taxonomy Position sections

fix.md and loop.md SHALL each contain a compact (≤15 line) "Loop Taxonomy Position" section placing the workflow within a 4-quadrant taxonomy (turn-based / goal-based / time-based / proactive) and stating entry guidance on three axes — how it starts, how it ends, when it fits. The section SHALL cross-reference the sibling quadrants by file path — `.claude/rules/moai/workflow/cadence-bridge.md` for the time-based quadrant (recommended placement per SPEC-CADENCE-BRIDGE-001, not yet authored as of this SPEC's plan-phase) and the `moai-workflow-ci-loop` skill for the proactive quadrant — rather than by SPEC ID.

### REQ-FLM-006 — Capability gate (Where) — template-first boundary

**Where** fix.md and loop.md carry verified template mirrors under `internal/template/templates/`, the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees, and SHALL NOT introduce this SPEC's ID or any SPEC ID into the template-tree copies (template-neutrality guard).

---

## Acceptance Criteria (§3 — inline AC, Tier S)

Minimum 2 required; 10 provided, mapped to the 6 REQs above plus 2 regression/guard checks.

| AC | REQ | Given / When / Then |
|----|-----|----------------------|
| AC-FLM-001 | REQ-FLM-001 | **Given** a fix run completes Phase 3, **When** Phase 4 executes per the rewritten fix.md text, **Then** the fix report's verification section is a claim/evidence table where every PASS row cites the exact re-run command and its parsed exit code/count — verify via `grep -A 15 '^## Phase 4' .claude/skills/moai/workflows/fix.md` showing claim/evidence table language and a `verification-claim-integrity.md` cross-reference. |
| AC-FLM-002 | REQ-FLM-001 | **Given** the rewritten Phase 4 text, **When** inspected for the pre-existing bare-prose bullets, **Then** "Confirm fixes resolved the targeted issues" and "Detect any regressions introduced by fixes" no longer stand as unqualified prose — each is replaced or qualified with an evidence-bearing instruction (re-run command + parsed output). |
| AC-FLM-003 | REQ-FLM-002 | **Given** Phase 1 produced a baseline issue list, **When** Phase 4 runs per the rewritten text, **Then** fix.md instructs a FULL re-scan (not scoped only to modified files) and an explicit diff of the resulting list against the Phase 1 baseline — verify via grep for "baseline" and "regression" co-occurring in the Phase 4 section. |
| AC-FLM-004 | REQ-FLM-002 | **Given** the post-fix full scan surfaces an issue absent from the Phase 1 baseline, **When** Phase 4 evaluates it per the rewritten text, **Then** the instructed action is revert-or-report-failed — never silent acceptance (grep for an explicit "regression" handling clause naming both outcomes). |
| AC-FLM-005 | REQ-FLM-003 | **Given** fix exits with residual Level 4 items or unresolved errors, **When** the fix report is generated, **Then** fix.md instructs persistence to `.moai/state/loop-verdict-<id>.json` with `exit_kind: "one-shot-residue"` and `iterations_used: 1` — verify via `grep -n 'loop-verdict\|one-shot-residue' .claude/skills/moai/workflows/fix.md` (≥1 match each). |
| AC-FLM-006 | REQ-FLM-004 | **Given** non-empty residue, **When** the report is generated, **Then** fix.md recommends `/moai loop` entry as a suggestion only, and the Pipeline Contract's Repeatability clause text remains present verbatim — verify via `grep -c "Re-entry requires explicit user re-invocation" .claude/skills/moai/workflows/fix.md` returning exactly 1 (unchanged, not duplicated). |
| AC-FLM-007 | REQ-FLM-005 | **Given** fix.md, **When** grepped for `"Loop Taxonomy Position"`, **Then** exactly one H2 section exists, its line span is ≤15 lines, and it names all four quadrants (turn-based / goal-based / time-based / proactive). |
| AC-FLM-008 | REQ-FLM-005 | **Given** loop.md, **When** the same check runs AFTER SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite has landed (verified via that SPEC's progress.md §E.4 `sync_commit_sha` populated, or `git log` showing its commits), **Then** loop.md also carries exactly one ≤15-line "Loop Taxonomy Position" section naming all four quadrants, added as a pure addition (no lines inside SPEC-LOOP-VERDICT-CONTRACT-001's scope — Steps 1/4/9, § Completion Conditions — are touched). |
| AC-FLM-009 | REQ-FLM-006 | **Given** the edited fix.md and loop.md, **When** `diff` runs between `.claude/skills/moai/workflows/{fix,loop}.md` and their `internal/template/templates/` mirrors, **Then** the two trees are byte-identical, AND `grep -rn "SPEC-FIX-LOOPMAP" internal/template/templates/` returns 0 matches. |
| AC-FLM-010 | Regression guard (Agentless contract) | **Given** the edited fix.md, **When** `go test -run TestAgentlessUtilityNoLLMControlFlow ./internal/template/...` runs, **Then** it PASSES — no forbidden LLM-dispatch phrase (per `agentless_audit_test.go` `forbiddenControlFlowPatterns`) was introduced by this SPEC's edits. |

---

## Constraints

1. **Agentless fixed-pipeline preserved (HARD)** — fix.md stays classified Agentless: no LLM-driven control flow is introduced; the Phase 3 static Level→agent dispatch table (lines 176-179) is untouched; `TestAgentlessUtilityNoLLMControlFlow` (`internal/template/agentless_audit_test.go`) MUST continue to pass (AC-FLM-010).
2. **3-phase contract preserved (HARD)** — the localize→repair→validate mapping (Phase 1+2+2.5 / Phase 3 / Phase 4), the no-op exit-code-0 semantics, and the fail-fast repair semantics are unchanged; this SPEC only deepens Phase 4 (validate) and adds a residue-persistence tail, it does not restructure the phase mapping.
3. **loop.md is shared, additive-only surface (HARD)** — SPEC-LOOP-VERDICT-CONTRACT-001 declares fix.md fully out of its own scope (its plan.md §G: *"Touching fix.md Phase 4 (L5 is explicitly out of scope — record, don't remediate)"*), so fix.md edits (M1, M2) have NO landing-order dependency. loop.md, however, IS touched by both SPECs: SPEC-LOOP-VERDICT-CONTRACT-001 rewrites Steps 1/4/9 and § Completion Conditions; this SPEC ADDS one new "Loop Taxonomy Position" section only. The addition MUST be sequenced AFTER SPEC-LOOP-VERDICT-CONTRACT-001's loop.md rewrite lands, to avoid diff/merge collision on a file mid-rewrite by a sibling SPEC.
4. **No Go code changes (HARD)** — doctrine/skill-doc edits only; no new config keys, no new ceilings, no Go loader for the residue file (consistent with SPEC-LOOP-VERDICT-CONTRACT-001's own no-Go-loader constraint).
5. **Scope discipline (HARD)** — `moai.md`, `ralph.yaml`, `workflow.yaml`, `gate.md`, `review.md` are NOT touched by this SPEC.
6. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — loop-verdict JSON schema definition

- The `.moai/state/loop-verdict-<id>.json` schema itself (field set, `ceiling`/`manual-residue` base enum values, the 5-section vci verdict report format) is owned and defined by SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005 / plan.md §D D2. This SPEC only consumes that schema and extends its `exit_kind` enum with one additional value (`one-shot-residue`).

### Out of Scope — cadence recipe catalog

- The sanctioned cadence recipe catalog (drift watcher, lean review, backlog re-discovery) and the `.claude/rules/moai/workflow/cadence-bridge.md` rule file itself are owned by SPEC-CADENCE-BRIDGE-001. This SPEC only cross-references that file path from the fix.md/loop.md Loop Taxonomy Position sections.

### Out of Scope — `/moai loop` exit semantics

- Any change to `/moai loop`'s success-exit predicate, ceiling-exit verdict contract, independent final pass, or max-iterations precedence is owned by SPEC-LOOP-VERDICT-CONTRACT-001 (REQ-LVC-001..008). This SPEC's only loop.md touch is the additive Loop Taxonomy Position section.

### Out of Scope — independent-agent verification spawn for `/moai fix`

- Spawning a separate verifier `Agent()` (fresh-context re-run, distinct from the builder) for fix's Phase 4 — the independent-evaluator pattern SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-003 introduces for the iterative loop — is deliberately NOT applied to `/moai fix`. Rationale: `/moai fix` is a one-shot Agentless pipeline (constraint 1); its mechanical independence is achieved by re-running scanners against exit codes/counts (REQ-FLM-001/002), not by agent independence. Introducing an agent-spawned verifier into a one-shot pipeline would blur the Agentless classification and is a different SPEC's scope if ever pursued.

---

## Cross-References

- **EXTEND base (doc)**: `.claude/skills/moai/workflows/fix.md` (Phase 4 rewrite, new Loop Taxonomy Position section) + `.claude/skills/moai/workflows/loop.md` (additive Loop Taxonomy Position section only). Both have verified template mirrors under `internal/template/templates/`.
- **Producer consumed**: SPEC-LOOP-VERDICT-CONTRACT-001 REQ-LVC-005 (`.moai/state/loop-verdict-<id>.json` schema, plan.md §D D2) — extended, not redefined, by REQ-FLM-003.
- **Sibling scope boundary cited**: SPEC-LOOP-VERDICT-CONTRACT-001 plan.md §G (*"Touching fix.md Phase 4 ... is explicitly out of scope"*) — the L5 origin of this SPEC.
- **Forward reference (not yet authored)**: `.claude/rules/moai/workflow/cadence-bridge.md` — recommended placement per SPEC-CADENCE-BRIDGE-001 (Epic Workflow-Reflex 5 of 6); cross-referenced by file path only, per REQ-FLM-005.
- **Cited, unmodified**: `moai-workflow-ci-loop` skill (proactive quadrant); `verification-claim-integrity.md` §1.1 + §3 (evidence-bearing format REQ-FLM-001 reuses); `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` (Agentless pipeline contract fix.md already documents).
- **Guard preserved**: `internal/template/agentless_audit_test.go` `TestAgentlessUtilityNoLLMControlFlow` (AC-FLM-010).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + progress; Tier S, AC inline in §3). Follow-up to SPEC-LOOP-VERDICT-CONTRACT-001 §Out of Scope L5 deferral. Evidence-bearing Phase 4 + regression guard + one-shot residue persistence/escalation + Loop Taxonomy Position sections + template-first boundary. |
