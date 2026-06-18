---
id: SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001
title: "Interrupt Ledger Closure — synthetic result for aborted Agent() delegations, team-ac-verify exit-2 ledger_note, and Error Recovery banner annotation"
version: "0.2.0"
status: implemented
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tags: "orchestration, ledger, interrupt, harness"
era: V3R6
depends_on:
  - SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001
---

# SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001

> **Tier M** (standard) — 3 core artifacts (spec.md + plan.md + acceptance.md) + progress.md §E skeleton. No research.md / design.md: this is a policy-layer codification of an externally grounded book doctrine (harness-books book1 ch04 "账本闭环" + ch07 parent-abort/orphan-task cleanup), not a codebase-analysis-driven design.

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-06-18 | manager-spec | Initial draft — Sprint 15 harness-books application cohort (P1a). Introduces `ledger` / `synthetic result` / `dangling tool_use` vocabulary into active doctrine (closes grep-zero-hit gap). Depends on P0 sibling SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001 (§Hook Invocation Surface Recovery-Signal Carve-Out); the §Ledger Closure subsection is in a distinct section to avoid collision. |

## §A. Context and Problem

### §A.1 The Gap (verified at plan-time)

A grep for `interruptBehavior|ledger.?clos|synthetic.*tool_result|dangling.*tool_use` and a broader grep for the bare words `ledger` / `synthetic` / `dangling` across `.claude/rules/moai/core/` and `.claude/output-styles/moai/` returns **zero hits** as of 2026-06-18 (run during plan-phase; see acceptance.md AC-LEDGER-005 for the reproducible command and §A.3 for the honest-baseline note). The ledger-closure invariant — well-known in the harness-books source material — is absent from active MoAI-ADK orchestration doctrine. (The scope is narrowed to `.claude/rules/moai/core/` because the bare word `ledger` does appear elsewhere under `.claude/rules/moai/workflow/` — e.g. `team-protocol.md` "Magentic ledger" / "Task Ledger" — but in a distinct sense: an append-only tasklist data structure, not the orchestration-interrupt ledger-closure invariant this SPEC introduces. §A.3 documents this scoping explicitly.)

### §A.2 Why this matters

MoAI-ADK's orchestration model has exactly the hazard the ledger-closure invariant addresses:

1. **Orchestrator → Agent() delegation**: orchestrator spawns a sub-agent; the sub-agent runs, times out, or is aborted (user interrupt, parent-abort propagation). If the orchestrator proceeds as if the delegation "returned cleanly", it leaves a **dangling tool_use** in its own context — the next turn sees an open promise, not a closed result. This is the orchestration-layer analogue of a dangling `tool_use` with no matching `tool_result` at the model-API layer.
2. **Team-mode hooks**: `team-ac-verify.sh` (TaskCompleted) and the TeammateIdle hook may reject a task (exit 2) or signal "keep working". A rejected task without a closing artifact in the orchestrator's context is an open promise.
3. **Persistence layer**: `session-handoff.md` paste-ready resume Block 3-4 preconditions already implement the analogue of ledger closure across `/clear` (the resume message re-establishes verifiable preconditions before continuing). But the in-session interrupt case — no `/clear` — has no codified closure step.

The current "Missing Inputs" blocker-report + re-delegation pattern (`.claude/rules/moai/core/agent-common-protocol.md` §Missing Inputs + §Re-delegation Procedure) partially addresses the *recoverable* case (sub-agent returned a blocker). It does **not** address the *interrupted* case (no return at all) and does **not** require a synthetic closing artifact.

### §A.3 Honest baseline

The plan-time grep claim is "zero hits in `.claude/rules/moai/core/` and `.claude/output-styles/moai/` for the orchestration-interrupt sense of ledger closure". The plan-time verification command was:

```
grep -rniE 'interruptBehavior|ledger.?clos|synthetic.*tool_result|dangling.*tool_use' .claude/rules/moai/ .claude/output-styles/
grep -rniE '\bledger\b|\bsynthetic\b|\bdangling\b' .claude/rules/moai/core/ .claude/output-styles/moai/
```

The phrase-targeted grep returned zero hits across `.claude/rules/moai/` and `.claude/output-styles/`. The bare-word grep is scoped to `.claude/rules/moai/core/` deliberately: the bare word `ledger` DOES appear under `.claude/rules/moai/workflow/` (`team-protocol.md` "Magentic ledger" / "Task Ledger" — an append-only tasklist data structure), but in a distinct sense that is NOT the orchestration-interrupt ledger-closure invariant this SPEC introduces. The §A.1 claim is therefore honestly narrowed to `.claude/rules/moai/core/` + `.claude/output-styles/moai/` (the orchestration-doctrine surfaces), not the broader rule tree. If at run-time the bare word `ledger` appears in unrelated prose elsewhere, the claim must remain narrowed to "no hits for the orchestration-interrupt sense of ledger closure in the two target surfaces before this SPEC" — the SPEC author must not overclaim. The AC (AC-LEDGER-005) re-runs the narrower phrase-targeted grep at run-time, which is the more defensible signal than the bare-word grep.

### §A.4 Source material (externally grounded)

| Source | Chapter | Invariant | Application here |
|--------|---------|-----------|------------------|
| github.com/wquguru/harness-books | book1 ch04 | "账本闭环" (ledger closure): "只要系统向外承诺了一段执行，就要在中断时把账补平" — whenever the system has promised an execution externally, it must close the ledger on interrupt. book1 ch04 describes a streaming tool executor that emits synthetic results for interrupted/sibling-error cases so every `tool_use` receives a `tool_result` (paraphrase of the chapter's mechanism — the specific branch taxonomy is not quoted verbatim). | REQ-LEDGER-001 (orchestrator emits synthetic result on aborted Agent()) |
| github.com/wquguru/harness-books | book1 ch07 | Parent-abort propagates to forked children; cleanup handlers registered to avoid orphan tasks; agents are observable lifecycle objects (SubagentStart/SubagentStop hooks, exit-code-2 stderr feedback). | REQ-LEDGER-002 / REQ-LEDGER-003 (team-mode reject path emits closing artifact; rejected task not left open without owner) |

## §B. Requirements (GEARS)

### REQ-LEDGER-001 — Synthetic result on aborted Agent() delegation

[ZONE:Evolvable] [HARD] **When** an `Agent()` delegation is aborted (user interrupt, parent-abort propagation, or timeout), the orchestrator **shall** emit a synthetic result summary into its own context before the next delegation. This artifact is named the **ledger-closing artifact**.

GEARS form: **When** [an Agent() delegation is aborted by user interrupt, parent-abort propagation, or timeout], the [orchestrator] **shall** [emit a synthetic result summary — the ledger-closing artifact — into its own context before issuing the next delegation].

Notes:
- The ledger-closing artifact is a short prose summary, not a structured data record. Its purpose is to close the open promise in the orchestrator's context so the next turn does not see a dangling delegation.
- "Aborted" includes: user Ctrl+C / interrupt during the delegation; parent-abort propagation (the orchestrator's own turn was aborted and the sub-agent was killed); sub-agent timeout (no return before wall-clock or token-budget ceiling).
- This requirement does NOT change the "Missing Inputs" blocker-report pattern: a blocker report is a *return*, not an *abort*. The blocker-report path already produces a structured return; REQ-LEDGER-001 covers the case where no return is produced at all.

### REQ-LEDGER-002 — team-ac-verify.sh exit-2 ledger_note field

[ZONE:Evolvable] [HARD] **When** the `.claude/hooks/moai/team-ac-verify.sh` hook rejects a `TaskCompleted` (exit 2), the hook's structured JSON output **shall** include a `ledger_note` field carrying a short human-readable note about why the task was rejected. The orchestrator injects this `ledger_note` as the ledger-closing artifact for that task.

GEARS form: **When** [team-ac-verify.sh rejects a TaskCompleted with exit code 2], the [hook] **shall** [emit a `ledger_note` field in its structured JSON output carrying a short human-readable rejection note].

Notes:
- The exit-code semantics are unchanged: exit 2 still = reject completion. This REQ only ADDS the `ledger_note` field to the rejection JSON.
- The orchestrator's consumption of `ledger_note` (injecting it as the closing artifact) is binding on the orchestrator; the orchestrator's obligation to read and inject is normative, not the hook's obligation.
- The hook today does NOT emit exit 2 in any path (every active-team-mode branch currently emits `decision: allow` and exits 0). REQ-LEDGER-002 introduces the `ledger_note` field onto the reject path that this SPEC's M2 milestone adds minimally. Full AC verification logic (parsing acceptance.md and running the AC's evidence command) is explicitly OUT OF SCOPE — see §X.

### REQ-LEDGER-003 — TeammateIdle exit-2 task closure

[ZONE:Evolvable] [HARD] **When** the TeammateIdle hook fires exit-2 "keep working" and rejects a task's completion, the rejected task's TaskList entry **shall not** be left in an open state without a reassignment owner. The orchestrator re-assigns the task (to a new or existing teammate) or closes it with a ledger-closing artifact.

GEARS form: **When** [the TeammateIdle hook rejects a task completion via exit code 2], the [orchestrator] **shall** [re-assign the rejected task to a definite owner or close it with a ledger-closing artifact, eliminating the open-without-owner state].

Notes:
- This REQ binds the orchestrator's TaskList hygiene, not the hook's behavior. The hook already emits exit 2 = reject; REQ-LEDGER-003 closes the resulting open task.
- "Re-assign" includes: spawning a new teammate, re-delegating to the same teammate with a refined prompt, or closing the task as obsolete with a synthetic closing note.

### REQ-LEDGER-004 — Error Recovery banner annotation

[ZONE:Evolvable] [HARD] **When** an `Agent()` delegation was interrupted or aborted, the orchestrator's next response **shall** reference the synthetic ledger-closing artifact rather than proceeding as if the delegation returned cleanly. The `.claude/output-styles/moai/moai.md` §8 Error Recovery banner carries a small "Interrupt Closure" annotation to this effect.

GEARS form: **When** [an Agent() delegation was interrupted or aborted], the [orchestrator's next response] **shall** [reference the synthetic ledger-closing artifact rather than proceeding as if the delegation returned cleanly].

Notes:
- The change to `moai.md` §8 is a small annotation on the existing Error Recovery banner, NOT a structural rewrite of §8.
- The annotation is a render-surface reminder; the binding obligation on the orchestrator is REQ-LEDGER-001 (emit the artifact). REQ-LEDGER-004 just surfaces the reminder in the visible banner template.

### REQ-LEDGER-005 — Section placement (collision-free)

[ZONE:Evolvable] [HARD] **Where** the §Ledger Closure subsection is added to `.claude/rules/moai/core/agent-common-protocol.md`, it **shall** occupy a new subsection distinct from the §User Interaction Boundary's existing prose AND distinct from the §Hook Invocation Surface subsection (the latter owned by the P0 sibling SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001 Recovery-Signal Carve-Out).

GEARS form: **Where** [the §Ledger Closure subsection is added to agent-common-protocol.md], the [subsection] **shall** [occupy a new location distinct from the existing §User Interaction Boundary prose and from the §Hook Invocation Surface subsection — collision-free with the P0 sibling SPEC's Recovery-Signal Carve-Out].

Notes:
- This REQ enforces the scope boundary with the P0 SPEC. AC-LEDGER-006 verifies the two additions live in distinct `## `-level or `### `-level sections via grep.
- The intended placement is: a new `### Ledger Closure` subsection under the existing `## User Interaction Boundary` H2, as a sibling to `### Subagent Prohibitions`, `### Orchestrator Obligations`, `### Hook Invocation Surface`, and `### Blocker Report Format`. This keeps it under the user-interaction surface (where the AskUserQuestion monopoly already lives) without colliding with the §Hook Invocation Surface's Recovery-Signal Carve-Out.

### REQ-LEDGER-006 — Cross-references

[ZONE:Evolvable] [HARD] The §Ledger Closure subsection **shall** cross-reference (a) the harness-books book1 ch04 ledger-closure invariant, (b) book1 ch07 parent-abort/orphan-task cleanup, and (c) `.claude/rules/moai/workflow/session-handoff.md` paste-ready resume Block 3-4 preconditions as the persistence-layer analogue of ledger closure across `/clear`.

GEARS form: **Where** [the §Ledger Closure subsection is authored], the [subsection] **shall** [cross-reference book1 ch04 ledger closure + book1 ch07 parent-abort/orphan-task cleanup + session-handoff.md Block 3-4 preconditions as the persistence-layer analogue].

## §C. Constraints

| ID | Constraint | Source |
|----|------------|--------|
| CON-1 | The MEANING of exit 2 in `team-ac-verify.sh` remains "reject completion" (unchanged). This SPEC ADDS a minimal reject path (currently ABSENT from the hook — verified at plan-time the hook has zero exit-2 branches) that emits exit 2 with a `ledger_note` field. The trigger is a minimal stub per §X.1, not full AC-verification logic (full AC-verification is out of scope, forward-linked to a follow-up SPEC). This is a behavioral ADDITION (a new code path), honestly described — not a no-op field-only addition onto an existing reject path. | User constraint (Sprint 15 prompt, honestly scoped per verification-claim-integrity §1.1 surface 3) |
| CON-2 | moai.md §8 Error Recovery banner is a small annotation only — NOT a structural rewrite of §8. | User constraint (Sprint 15 prompt) |
| CON-3 | The §Ledger Closure subsection goes in a NEW subsection of agent-common-protocol.md, distinct from §User Interaction Boundary's existing prose AND from §Hook Invocation Surface. | User constraint (Sprint 15 prompt) |
| CON-4 | Do NOT proceed to run-phase — plan-phase only. | User constraint (Sprint 15 prompt) |
| CON-5 | Use `## §X. Out of Scope (...)` + h3 sub-section pattern for the exclusions section (avoid the MissingExclusions lint ERROR that P0/P3 hit). | User constraint (Sprint 15 prompt) |
| CON-6 | book1 ch04 + ch07 cited. | User constraint (Sprint 15 prompt) |

## §D. Non-Functional Requirements

- **Readability**: the §Ledger Closure subsection is ≤ 80 lines (prose + REQ list + cross-references). It does not duplicate session-handoff.md content; it cross-references.
- **Backward compatibility**: no existing rule's semantics change. The P0 sibling SPEC's §Hook Invocation Surface Recovery-Signal Carve-Out is untouched (verified by AC-LEDGER-006).
- **Lint compliance**: spec-lint clean (`moai spec lint` exit 0). The exclusions section uses the `## §X. Out of Scope` pattern per CON-5.

## §E. AC Matrix (summary — full enumeration in acceptance.md)

| REQ | Primary AC | Severity |
|-----|-----------|----------|
| REQ-LEDGER-001 | AC-LEDGER-001 (§Ledger Closure has clause (a)) | MUST-PASS |
| REQ-LEDGER-002 | AC-LEDGER-002 (team-ac-verify.sh exit-2 path has `ledger_note`) | MUST-PASS |
| REQ-LEDGER-003 | AC-LEDGER-001 (clause c) | MUST-PASS |
| REQ-LEDGER-004 | AC-LEDGER-004 (moai.md §8 Error Recovery banner has Interrupt Closure annotation) | MUST-PASS |
| REQ-LEDGER-005 | AC-LEDGER-006 (collision-free section placement) | MUST-PASS |
| REQ-LEDGER-006 | AC-LEDGER-001 (cross-references present) + AC-LEDGER-007 (book1 ch04/ch07 cited) | MUST-PASS |
| (gap-closure) | AC-LEDGER-005 (grep reproducibility — vocabulary introduced) | MUST-PASS |

> **Note on REQ-LEDGER-003 traceability (no standalone AC-LEDGER-003):** REQ-LEDGER-003 (TeammateIdle exit-2 task closure) is verified by **AC-LEDGER-001 clause (c)** — it appears as one clause inside the §Ledger Closure subsection's four-clause enumeration, alongside REQ-LEDGER-001 (clause a), REQ-LEDGER-002 (clause b), and REQ-LEDGER-006 (clause d). There is intentionally **no standalone AC-LEDGER-003**; the acceptance.md AC ID sequence jumps `001 → 002 → 004` by design. The §E matrix row for REQ-LEDGER-003 therefore cites `AC-LEDGER-001 (clause c)`, and acceptance.md §D.2 carries the matching traceability row. This keeps spec.md §E and acceptance.md §D.2 in agreement (no dangling `AC-LEDGER-003` reference in either file).

## §F. Dependencies

- **depends_on**: `SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001` (P0 sibling — owns §Hook Invocation Surface Recovery-Signal Carve-Out in the same target file `agent-common-protocol.md`). The dependency is a **collision-avoidance** dependency, not a content dependency: this SPEC must not modify the §Hook Invocation Surface area the P0 SPEC owns. If the P0 SPEC has not merged when this SPEC runs, the run-phase still proceeds (this SPEC touches a different section); AC-LEDGER-006 verifies the boundary regardless of merge order.

## §G. Risks

| Risk | Mitigation |
|------|-----------|
| Overclaim on grep-zero-hit if `ledger` appears in unrelated prose elsewhere | AC-LEDGER-005 uses the phrase-targeted grep (`interruptBehavior|ledger.?clos|synthetic.*tool_result|dangling.*tool_use`), not the bare-word grep. §A.3 documents the honest baseline. |
| Scope creep into full AC verification logic (parsing acceptance.md, running evidence commands) | §X.1 explicitly excludes full AC verification. REQ-LEDGER-002 only adds the `ledger_note` field to the reject path; the reject-path trigger condition itself is a minimal stub. |
| Collision with P0 SPEC's §Hook Invocation Surface edit | REQ-LEDGER-005 + AC-LEDGER-006 enforce distinct-section placement. Plan.md §C pre-flight verifies the section anchors before edit. |
| moai.md §8 structural drift if the annotation is too large | CON-2 bounds the annotation to a small note; M3 milestone enforces the bound. |

## §H. Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — target file for §Ledger Closure subsection (REQ-LEDGER-001/003/005/006).
- `.claude/rules/moai/core/agent-common-protocol.md` §Hook Invocation Surface — owned by P0 SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001 (Recovery-Signal Carve-Out). Collision-avoidance boundary (REQ-LEDGER-005).
- `.claude/hooks/moai/team-ac-verify.sh` — target file for `ledger_note` field (REQ-LEDGER-002).
- `.claude/output-styles/moai/moai.md` §8 Response Templates → Error Recovery banner — target for Interrupt Closure annotation (REQ-LEDGER-004).
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume Block 3-4 preconditions are the persistence-layer analogue (REQ-LEDGER-006 cross-reference).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the §1.1 orchestrator self-report surface binds the ledger-closing artifact's truthfulness (the artifact must be a real summary, not a fabricated "success").
- github.com/wquguru/harness-books book1 ch04 (账本闭环) + ch07 (parent-abort/orphan-task cleanup) — externally grounded source (REQ-LEDGER-006).

## §X. Out of Scope (What NOT to Build)

This section is present to satisfy the MissingExclusions lint rule and to prevent scope creep. Per CON-5 and the `internal/spec/CLAUDE.md` Heading convention, it uses the `## §X. Out of Scope` H2 with `### Out of Scope —` infix h3 sub-headings, each carrying at least one `-` bullet item.

### §X.1 Out of Scope — Full AC verification logic in team-ac-verify.sh

- No full AC verification logic (parsing `acceptance.md`, running the AC's evidence command, emitting exit 2 based on AC failure) is added to `team-ac-verify.sh`. The hook today does not emit exit 2 in any path; this SPEC adds the `ledger_note` field onto a **minimal reject path** (M2) whose trigger is a stub (e.g., an explicit `--reject` test flag or a documented TODO).
- No change to the reject-path trigger condition beyond the minimal stub. Full AC verification is deferred to a follow-up SPEC (forward-link: future `SPEC-V3R6-TEAM-AC-VERIFY-FULL-001`).

### §X.2 Out of Scope — TeammateIdle hook modification

- No modification to the TeammateIdle hook script itself. REQ-LEDGER-003 binds the **orchestrator's** TaskList hygiene (re-assign or close the rejected task), not the hook's exit-2 emission.
- No new hook script files under `.claude/hooks/moai/`. The TeammateIdle exit-2 "keep working" signal is consumed by the orchestrator; no hook-side change is in scope.

### §X.3 Out of Scope — Structured ledger artifact (data record)

- No JSON schema for the ledger-closing artifact. The artifact (REQ-LEDGER-001) is a short **prose summary**, not a structured data record.
- No `.moai/state/ledger.json` state file or equivalent. A structured artifact is a possible follow-up but is out of scope here (forward-link: future `SPEC-V3R6-LEDGER-ARTIFACT-STRUCT-001`).

### §X.4 Out of Scope — Run-phase execution

- No run-phase commit, no `moai spec lint` run on the modified target files, no M1-M4 file edits. Per CON-4, this delivery is **plan-phase only** — it delivers spec.md + plan.md + acceptance.md + progress.md §E skeleton.
- No frontmatter status transition beyond the initial `(none) → draft` set at plan-phase. The `draft → in-progress` transition is owned by manager-develop at run-phase entry (per the Status Transition Ownership Matrix).

### §X.5 Out of Scope — Subagent-side ledger closure

- No obligation on sub-agents to emit their own ledger-closing artifacts. This SPEC binds the **orchestrator's** obligation to close the ledger when a delegation it spawned is aborted; sub-agents already return blocker reports via the existing "Missing Inputs" pattern.
- No change to the `agent-common-protocol.md` §Missing Inputs or §Re-delegation Procedure subsections. The abort case (no return at all) is the orchestrator's responsibility and is distinct from the blocker-report case (structured return).
