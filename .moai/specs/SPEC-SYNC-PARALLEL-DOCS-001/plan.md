# plan.md — SPEC-SYNC-PARALLEL-DOCS-001

> Implementation plan, milestones, technical approach. Ordered by decision-reversibility (highest-change-likelihood decisions first).

## §A Context

### A.1 Scope envelope

4 axes from the autonomy-workflow redesign §3.5, grounded in current file:line evidence:

| Axis | Current state (file:line) | Target state |
|---|---|---|
| **A5** docs ∥ audit | `doc-execution.md` Phase 11-12 runs serially after `quality-gates-quality.md` Phase 7-10. `manager-docs` spawned at Phase 11 Step 1.4 + Phase 12 Step 2.2. | Docs drafter fan-out (existing `FO-SYNC-4` 5-drafter structure, `doc-execution.md` L126-140) launches CONCURRENTLY with Phase 7 audit. Single-writer applier merges at gate-sync-2. |
| **A7** MX early+parallel | `quality-gates-quality.md` Phase 9 (L139-244) runs serially AFTER Phase 7-8; Phase 10 coverage (L246-303) runs AFTER Phase 9. P1/P2 abort (L143-153) fires AFTER coverage cost is paid. | Phase 9 MX scan (existing `FO-SYNC-2` sharded structure, L186) launches CONCURRENTLY with Phase 7 audit. P1/P2 gate fires BEFORE Phase 10 coverage. |
| **A9** §E + 7-batch | `manager-develop-prompt-template.md` § Section E (E1-E8) is a self-verification deliverable; `agent-common-protocol.md` § Parallel Execution canonical 7-command batch RE-EXECUTES test/lint/vet/cover. | §E promoted to formal attributable artifact (VCI §2: command + output + baseline). Orchestrator batch switches to attributable diff-check against §E + AUDIT-SNAPSHOT-001 A4 snapshot. Fallback to re-execution on any mismatch. |
| **A6** retry Tier ceilings | `harness.yaml` `levels.{minimal,standard,thorough}.plan_audit.max_iterations` = 1/3/3 (flat per-level). `plan-auditor.md` Retry Loop Contract L386-418 hard-codes `max_iterations: 3`. | Tier-aware ceiling: S=1, M=2, L=3. `plan-auditor.md` consults Tier ceiling in place of flat constant. |

### A.2 Dependency

- **SPEC-AUDIT-SNAPSHOT-001** (completed) — A9 builds on the A4 shared diagnostic snapshot (REQ-AUDIT-SNAPSHOT-004). The snapshot infrastructure (`moai verify check --key-current`, keyed by HEAD SHA) is the attribution chain A9's diff-check consults. No new snapshot store is invented.

### A.3 Decision-reversibility ranking (highest-change-likelihood first)

1. **A9 diff-check substitution** (most reversible-semantically, highest binding on verification-claim-integrity invariant) — binds the orchestrator's verification discipline; a bug here silently bypasses verification. Get this reviewed first.
2. **A6 Tier-aware ceiling schema location** (OQ-2 open) — harness.yaml vs. plan-auditor body. Config-schema decision; hard to reverse once consumers depend on the location.
3. **A5/A7 scheduling change** (workflow-skill prose edits, mechanically reversible) — the FO-SYNC-4 / FO-SYNC-2 structures are unchanged; only the scheduling prose changes from "serially after" to "concurrently with".
4. **gate-sync-2 merge sequencing** (mechanical) — single-writer applier pattern already established; A5 adds the audit verdict to the same gate.

## §B Known Issues (auto-injected — relevance-filtered)

- **B3 subagent boundary**: all A5/A7 concurrent agents are read-only drafters/scanners/auditors; the orchestrator is the sole writer at gate-sync-2. No `AskUserQuestion` calls inside any fan-out agent; blocker reports return to the orchestrator per `agent-common-protocol.md` § Blocker Report Format.
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, Test (per OS) can each fail separately. A9 changes orchestrator verification batch behavior (prose-level); no Go test changes required unless plan-auditor resolves OQ-1 toward a mechanical hook.
- **B10 scope discipline**: A5/A7/A9/A6 touch ONLY the sync-phase execution shape + plan-auditor retry contract + §E attribution. Run-phase, plan-phase, and cross-phase flows are NOT modified.
- **B11 AskUserQuestion prohibited in subagents**: A5 drafters, A7 MX shards, and the audit agents return blocker reports; the orchestrator runs `gate-sync-2` AskUserQuestion round.

## §C Pre-flight (manager-develop to run before any code change)

```bash
# 1. Current sync-phase execution shape — confirm serial scheduling
grep -n 'Phase 11\|Phase 12\|Read workflows/sync' .claude/skills/moai/workflows/sync.md
grep -n 'FO-SYNC-4\|FO-SYNC-2\|gate-sync-2' .claude/skills/moai/workflows/sync/doc-execution.md .claude/skills/moai/workflows/sync/quality-gates-quality.md

# 2. Confirm §E + 7-batch re-execution point
grep -n 'Section E\|E1\. AC\|E8\.' .claude/rules/moai/development/manager-develop-prompt-template.md
grep -n 'canonical 7-command\|verification-batch\|trust-but-verify' .claude/rules/moai/core/agent-common-protocol.md

# 3. Confirm plan-auditor Retry Loop Contract + harness.yaml constants
grep -n 'max_iterations\|Retry Loop Contract' .claude/agents/moai/plan-auditor.md
grep -n 'max_iterations' .moai/config/sections/harness.yaml

# 4. Confirm AUDIT-SNAPSHOT-001 A4 snapshot interface
grep -n 'moai verify check\|key-current\|shared.*snapshot' .claude/skills/moai/workflows/sync/quality-gates-quality.md
```

## §D Constraints (DO NOT VIOLATE)

1. **Verification-claim-integrity §1.1 + §2 inviolable** — A9's diff-check substitution binds to the attribution chain; on any mismatch, re-execution is restored (REQ-SPD-009). Silent verification bypass is the named anti-pattern this SPEC exists to avoid.
2. **Concurrency guard `[HARD]`** — no two write-capable agents run concurrently. A5/A7 parallelism is bought entirely by making fan-out read-only.
3. **Input-independence** — docs drafters (REQ-SPD-002) and MX scan (REQ-SPD-006) do NOT read the concurrent audit's output. Violating this creates a hidden serial dependency.
4. **Audit semantics immutable** — PASS thresholds (0.75/0.80/0.85), 4-dim weights (40/25/20/15), severity definitions, AC content are NOT modified.
5. **Backward compatibility** — pre-A6 SPECs (no `tier:` field) keep ceiling 3; minimal-harness users keep current behavior.
6. **PRESERVE list (do NOT modify in this SPEC)**:
   - `manager-develop-prompt-template.md` § Section E item structure (E1-E8) — A9 binds attribution, not item content
   - `FO-SYNC-4` five-drafter set / `FO-SYNC-2` sharded-scan structure — A5/A7 change scheduling, not structure
   - AUDIT-SNAPSHOT-001 snapshot store — A9 consumes, does not modify
   - `gate-sync-2` HUMAN GATE 2 semantics — A5 adds the audit verdict to the same gate, does not redefine it
7. **Forbidden commands**: `--no-verify`, force-push, `--amend` on shared branches.

## §E Self-Verification Deliverables (this SPEC's plan-phase)

- E-plan-1: GEARS compliance — all 12 REQ-SPD-* match one of the five GEARS patterns (Ubiquitous / When / While / Where / shall-not).
- E-plan-2: Out-of-scope section satisfies `OutOfScopeRule` lint — ≥1 `### Out of Scope — <topic>` H3 + `-` bullets.
- E-plan-3: frontmatter 12-field schema validated — no snake_case aliases (`created_at` / `updated_at` / `labels` / `spec_id`).
- E-plan-4: SPEC ID regex PASS — `SPEC-SYNC-PARALLEL-DOCS-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`.
- E-plan-5: `phase: "v3.x target"` — NOT a lifecycle-stage token (`plan`/`run`/`sync`/`mx`).
- E-plan-6: Per-axis current-file grounding — every axis row in §A.1 cites a real file:line that exists in the current tree.
- E-plan-7: AC count (14) ≤ Tier M ceiling (16); REQ count (12) ≤ Tier M ceiling (16).

## §F Milestones

Ordered by decision-reversibility (§A.3), NOT by wall-clock. Priority labels per `agent-common-protocol.md` § Time Estimation.

### M1 — A9 attributable diff-check + fallback (Priority High)

The most review-sensitive milestone: binds the verification-claim-integrity invariant. Reviewers + plan-auditor focus here.

- Promote `manager-develop` §E evidence to formal attributable artifact (REQ-SPD-007): each E1-E8 item names command + verbatim output + baseline-attribution.
- Wire orchestrator trust-but-verify batch to attributable diff-check (REQ-SPD-008): consult AUDIT-SNAPSHOT-001 A4 snapshot key + recorded command + recorded output.
- Implement diff-check fallback (REQ-SPD-009): on snapshot key mismatch / command mismatch / missing §E → re-execution.
- Resolve **OQ-1**: does the orchestrator batch today carry a structured "I am about to re-run command X" preamble? (Determines whether A9 wires a literal hook or applies a doctrinal switch.)

Files touched: `.claude/rules/moai/development/manager-develop-prompt-template.md` § Section E (attribution discipline clause); `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (diff-check doctrinal clause); `.claude/rules/moai/workflow/verification-batch-pattern.md` (attributable diff-check pattern).

### M2 — A6 Tier-aware plan-auditor ceiling (Priority High)

Config-schema decision; hard to reverse once consumers depend on the location.

- Resolve **OQ-2**: harness.yaml per-Tier `max_iterations` map vs. plan-auditor body Tier→ceiling table.
- Add Tier-aware ceiling (S=1, M=2, L=3) to the chosen location (REQ-SPD-010).
- Update `plan-auditor.md` Retry Loop Contract (L386-418) to consult Tier ceiling in place of flat `max_iterations: 3` (REQ-SPD-011).
- Backward-compat: where `tier:` is absent, ceiling remains 3 (Tier L fallback).

Files touched: `.moai/config/sections/harness.yaml` (per-Tier `max_iterations` map — IF OQ-2 resolves to harness.yaml); `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract (Tier-ceiling consultation).

### M3 — A5 docs ∥ audit concurrent scheduling (Priority Medium)

Mechanically reversible (workflow-skill prose edits).

- Update `sync.md` Phase Routing Table (L47-53) + Fan-Out Index (L58-64): mark docs drafter fan-out as concurrent-with-Phase-7 (not serially-after).
- Update `doc-execution.md` Phase 11 Step 1.4 + Phase 12 Step 2.2: docs drafter input derives from SPEC + git diff + divergence report, NOT audit result (REQ-SPD-002).
- Add concurrent-launch clause to `sync.md` FO-SYNC-1 / FO-SYNC-4 scheduling prose (REQ-SPD-001).
- Add gate-sync-2 merge sequencing clause: manager-docs applies drafts as single writer after both fan-outs return (REQ-SPD-003).

Files touched: `.claude/skills/moai/workflows/sync.md`; `.claude/skills/moai/workflows/sync/doc-execution.md`.

### M4 — A7 MX Tag early + parallel (Priority Medium)

Mechanically reversible.

- Update `quality-gates-quality.md` Phase 9 (L139-244): MX scan launches concurrently with Phase 7 audit fan-out (REQ-SPD-004).
- Add P1/P2 pre-coverage gate clause (REQ-SPD-005): on P1/P2 detection, halt BEFORE Phase 10 coverage executes.
- Confirm MX scan input-independence (REQ-SPD-006): scan reads git diff + source, not audit output.
- Add no-false-abort guard (AC-SPD-006): no P1/P2 → Phase 10 coverage proceeds unchanged.

Files touched: `.claude/skills/moai/workflows/sync/quality-gates-quality.md`; `.claude/skills/moai/workflows/sync.md` (Phase Routing Table entry for Phase 9).

### M5 — Cross-cutting: concurrency guard + audit-semantics-unchieved ACs (Priority Medium)

- Codify REQ-SPD-012 (concurrency guard) in `sync.md` parallel-quality-evidence fan-out section: all concurrent agents read-only, single-writer applier at gate-sync-2.
- Codify AC-SPD-014 (audit semantics unchanged) as a spec.md §D constraint (already present) + an acceptance.md invariant.

Files touched: `.claude/skills/moai/workflows/sync.md` (concurrency-guard clause); `.moai/specs/SPEC-SYNC-PARALLEL-DOCS-001/acceptance.md` (invariant ACs).

### M6 — Verification + sync-phase close (Priority Low)

- Run `moai spec lint --strict .moai/specs/SPEC-SYNC-PARALLEL-DOCS-001/` — expect 0 errors.
- Run `moai spec audit --json` — confirm the new SPEC is classified V3R6 era (H-4: §E.2 + §E.4 + sync_commit_sha after sync).
- Implementation Kickoff Approval gate → run-phase → sync-phase.

## §G Anti-Patterns

- **AP-SPD-001 — Silent verification bypass via A9**: substituting re-execution with diff-check WITHOUT the REQ-SPD-009 fallback. The fallback is the safety boundary; omitting it turns A9 into a verification bypass that violates `verification-claim-integrity.md` §1.1.
- **AP-SPD-002 — Concurrent write-capable agents**: spawning `manager-docs` (write-capable) concurrently with the audit on the assumption that "background" makes it safe. The `[HARD]` concurrency guard binds regardless of backgrounding; A5 parallelism is read-only-drafter only.
- **AP-SPD-003 — Input-dependency on the concurrent audit**: a docs drafter that reads "the audit's quality report" to decide CHANGELOG tone, or an MX scan that reads "the audit's functionality score" to gate itself. Both create hidden serial dependencies that defeat the concurrency.
- **AP-SPD-004 — A6 Tier-ceiling override of the plan-auditor's own judgment**: the Tier ceiling bounds the ITERATION COUNT, not the verdict. A Tier S SPEC still gets a full adversarial review on iteration 1; the ceiling only prevents iteration 2+.
- **AP-SPD-005 — Modifying §E item structure under A9**: A9 binds the attribution discipline (command + output + baseline) to E1-E8; it does NOT redefine E1-E8 content. Adding new §E items or rewriting existing ones is out of scope.
- **AP-SPD-006 — Lowering audit thresholds to make A5/A7 "easier"**: PASS thresholds (0.75/0.80/0.85) and 4-dim weights are immutable. A5/A7 change scheduling, not scoring.

## §H Cross-References

- Sibling (completed): `.moai/specs/SPEC-AUDIT-SNAPSHOT-001/` (A1-A4; A4 snapshot reused by A9).
- Sibling (completed): `.moai/specs/SPEC-SYNC-AUDIT-FALSIFICATION-001/` (sync-auditor falsification — distinct from execution-order parallelization here).
- Epic design report: `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §3.5 rows A5/A6/A7/A9.
- GEARS notation: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format.
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields + § Optional Fields (`tier:`).
- Concurrency guard: `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution.
- Verification-claim integrity: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2.
