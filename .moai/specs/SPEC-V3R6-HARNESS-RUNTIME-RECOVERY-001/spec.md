---
id: SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001
title: "Runtime Recovery Doctrine — withheld-recoverable errors, cheapest-first ladder, and anti-death-spiral hook carve-out"
version: "0.2.0"
status: draft
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "harness, recovery, runtime, doctrine"
era: V3R6
depends_on: []
---

# SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001

> **Tier M** (standard) — 3 core artifacts (spec.md + plan.md + acceptance.md) + progress.md §E skeleton. No research.md / design.md: this is a policy-layer codification of an externally grounded book doctrine, not a codebase-analysis-driven design.

## §A. Problem Statement

moai-adk has a sophisticated **deliverable-recovery** doctrine (CI autofix loop, paste-ready resume messages, CLAUDE.md §11 error handling) but it has **no doctrine for when the loop itself fails mid-SPEC**. When a compact PTLs (`prompt_too_long`), when a stop-hook blocks a turn that is itself a recovery signal, when a turn exhausts `max_output_tokens` mid-write — no moai-adk agent or rule says what to do. CLAUDE.md §11 explicitly defers to the platform ("token limit errors: execute `/clear`"), and that single sentence is the entire moai-adk surface for native runtime failure.

The sharpest manifestation of the gap: **moai-adk's own hooks can cause the death-spiral**. `sync-phase-quality-gate.sh` (Stop hook) and `status-transition-ownership.sh` (PostToolUse hook) are designed to exit-2 (block) on gate failure. When the turn that triggers them is itself a *recovery* turn — recovering from a compact, a PTL, a failed sync — a block puts the agent into the `error → stop-hook-blocks → retry → error` loop that book1 ch06 names the **death-spiral** and explicitly warns recovery MUST skip stop-hooks to avoid.

This gap was independently verified before authoring. Two distinct vocabularies were checked separately, because the error-type names and the recovery-ladder names are different surfaces:

```bash
# (1) Recovery-LADDER vocabulary across all rules files → ZERO hits (the real gap)
grep -rEl 'reactive.?compact|death.?spiral|circuit.?break|withheld.*recover' \
  .claude/rules/moai/   # → (no output)

# (2) Error-TYPE vocabulary — already has PRE-EXISTING coverage elsewhere
grep -rEl 'max_output_tokens' .claude/rules/moai/   # → hooks-system.md (3 hits, as a StopFailure error type)
```

The recovery-ladder vocabulary (reactive-compact, death-spiral, circuit-breaker, withheld-recoverable) is genuinely absent — that is the gap this SPEC closes (see REQ-RR-010). The error-type vocabulary (`max_output_tokens`) is already covered in `hooks-system.md` as a `StopFailure` error type and is NOT part of the gap this SPEC closes; the gap is the *recovery-ladder* vocabulary specifically.

## §B. Background (book grounding)

This SPEC codifies the single biggest gap found when diffing `github.com/wquguru/harness-books` book1 ("Claude Code harness engineering") against moai-adk. The doctrine is grounded in two chapters:

- **ch03 — "Query Loop is the heartbeat"**: the agent's core capability is maintaining a RECOVERABLE execution loop. The `queryLoop()`'s first duty is **input governance** (memory prefetch → snip → microcompact → context-collapse → autocompact) BEFORE the model call. Recoverable errors `{prompt_too_long (PTL), max_output_tokens, media_size}` are **WITHHELD** and routed to layered recovery BEFORE surfacing to the user.

- **ch06 — "Errors and recovery"**: "错误路径就是主路径" (the error path IS the main path); "恢复的目标是继续工作" (recovery's goal is to *keep working*, not to apologize). Recovery is layered **CHEAPEST-FIRST**: PTL → drain staged collapse → reactive-compact → surface. `hasAttemptedReactiveCompact` prevents self-loops. Autocompact has a circuit breaker `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3`. The **DEATH-SPIRAL** hazard: `error → stop-hook-blocks → retry → error` — so recovery MUST skip stop-hooks. Compact itself can PTL → `truncateHeadForPTLRetry()` last-resort. The deep goal is **NARRATIVE CONSISTENCY** — the system must explain what it tried, why it failed, what recovery it used.

- **§9.3 / §9.6 / §9.7** — the named principles (recoverable error withholding, cheapest-first ladder, narrative consistency).
- **ch06 §5** — the 5-line circuit-breaker invariants (max-consecutive-autocompact-failure, hasAttemptedReactiveCompact, compact-can-PTL escape, abort-closes-ledger, narrative-consistency requirement).

## §C. Scope

**IN SCOPE** (policy-layer deliverables):

1. NEW rule `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` defining the withheld-recoverable-error set, the 4-rung cheapest-first recovery ladder, the 5 circuit-breaker invariants, and the anti-death-spiral hook carve-out.
2. EDIT `.claude/rules/moai/core/agent-common-protocol.md` §Hook Invocation Surface ONLY — add a "Recovery-Signal Carve-Out" subsection.
3. EDIT `.claude/rules/moai/core/zone-registry.md` — add one new `CONST-V3R6-0XX` entry for the anti-death-spiral invariant.

## §D. Non-Goals (Policy-Layer NOT Runtime)

[HARD] This is a **POLICY-LAYER doctrine**, not a runtime reimplementation.

- This SPEC **does NOT** reimplement `queryLoop`, `recoverFromError`, `truncateHeadForPTLRetry`, `hasAttemptedReactiveCompact`, or any Claude Code TypeScript internal. moai-adk is a harness ON TOP of Claude Code and **cannot** modify the native query loop.
- This SPEC **does NOT** introduce Go code under `internal/` to detect or handle PTL / max_output_tokens / media_size. Those are platform-level signals already handled inside Claude Code.
- This SPEC **does NOT** touch `.claude/hooks/moai/*.sh` hook script bodies — it documents the *policy* that the EXISTING hooks already self-gate against death-spiral; no hook rewrite is in scope.
- This SPEC is explicitly **NOT** book1 appendix A.10's "build an agent runtime from scratch" pseudocode. The deliverable is a moai-OWNED POLICY DOCTRINE that tells moai-adk agents (and binds moai-adk's own hooks) what to do when native mechanisms fire mid-SPEC.

## §E. Out of Scope (What NOT to Build)

### §E.1 Out of Scope — excluded runtime rewrites

- No Go runtime implementation of recovery primitives (out of scope per §D).
- No modification to `agent-common-protocol.md` §User Interaction Boundary (separate concern).
- No new `§Ledger Closure` subsection in `agent-common-protocol.md` — that is **P1a SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001** (file-edit collision avoidance).
- No new hook script files (`.claude/hooks/moai/*.sh`) — the carve-out is documented as policy guidance binding agent behavior and future hook authors; no current hook script is rewritten by this SPEC (see REQ-RR-006/007).
- No change to CLAUDE.md §11 (the platform-deference sentence stays; the new doctrine layers ON TOP of it, it does not replace the `/clear` recommendation).
- No mechanical enforcement of the recovery-signal carve-out by the current hook scripts — `sync-phase-quality-gate.sh` and `status-transition-ownership.sh` receive PostToolUse/Stop JSON but do not parse a recovery signal; runtime-layer enforcement requires a stopReason-parsing hook deferred to a follow-up SPEC (forward-link: future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`).

## §F. Requirements (GEARS)

### §F.1 The withheld-recoverable-error set

**REQ-RR-001** (Ubiquitous): The runtime recovery doctrine shall name the withheld-recoverable-error set as `{prompt_too_long (PTL), max_output_tokens, media_size, compact-failure}` and shall state that these errors are routed to layered recovery rather than surfaced raw to the user.

**REQ-RR-002** (Event-detected): **When** a moai-adk agent observes a turn whose stopReason or context indicates one of the withheld-recoverable errors fired, the agent shall consult the runtime recovery doctrine and apply the cheapest-first ladder before concluding the turn failed.

### §F.2 The 4-rung cheapest-first recovery ladder

**REQ-RR-003** (Ubiquitous): The runtime recovery doctrine shall define a 4-rung cheapest-first recovery ladder, with each rung mapped to an existing moai-adk artifact, so that no rung is introduced without a concrete recovery path:

| Rung | Recovery action | moai-adk artifact cross-reference |
|------|-----------------|-----------------------------------|
| 1 | In-turn self-correction (continue without recap/apology per ch06 meta-message) | (agent behavior; no artifact) |
| 2 | Paste-ready resume + `/clear` | `.claude/rules/moai/workflow/session-handoff.md` |
| 3 | Session-handoff Block-0 worktree restart | `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern |
| 4 | Abort + preserve (PRESERVE-list discipline; progress.md state) | `.moai/specs/<SPEC-ID>/progress.md` + §F.4 below |

**REQ-RR-004** (State-driven): **While** a lower rung has not been attempted, the agent shall not jump to a higher-cost rung (no `/clear` before in-turn self-correction; no worktree restart before `/clear`).

### §F.3 The 5 circuit-breaker invariants

**REQ-RR-005** (Ubiquitous): The runtime recovery doctrine shall codify book1 ch06 §5's five circuit-breaker invariants as policy, each stated as a moai-adk-level rule the agent or hook must respect:

1. **Max-consecutive-autocompact-failure analogue** — after 3 consecutive autocompact/recovery failures at the same rung, the agent MUST escalate to the next rung rather than retry.
2. **hasAttemptedReactiveCompact analogue (no self-loop)** — a recovery action already attempted for the current failure MUST NOT be re-attempted in the same turn; the agent MUST advance to the next rung.
3. **Compact-can-PTL last-resort escape** — when reactive-compact itself PTLs, the agent MUST fall back to the last-resort truncation rung (rung 4 abort + preserve), not loop.
4. **Abort-closes-ledger** — when recovery aborts (rung 4), the agent MUST close the in-flight ledger by persisting state to `progress.md` before the session ends, so the next session resumes rather than restarts.
5. **Narrative-consistency requirement** — across the compact/recovery boundary, the agent MUST preserve narrative consistency via the 5-Section Evidence-Bearing Report format (cross-reference `.claude/rules/moai/core/verification-claim-integrity.md`): what was attempted, why it failed, what recovery was used, what remains a gap.

### §F.4 Anti-death-spiral hook carve-out (documentation-only policy)

> **Scope binding (LOAD-BEARING)**: This doctrine binds **AGENT BEHAVIOR** and **FUTURE hook evolution**. It does NOT bind the current hook scripts (`sync-phase-quality-gate.sh`, `status-transition-ownership.sh`), which receive PostToolUse/Stop JSON but do not parse a recovery signal; no mechanical enforcement is possible without the runtime-layer hook that parses `stopReason`, which is deferred to a follow-up runtime-layer SPEC (forward-link: future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`). This SPEC does NOT rewrite the hook scripts.

**REQ-RR-006** (State-driven, **documentation-only policy recommendation**): **While** a turn's stopReason or context indicates the turn is a recovery signal (recovering from a sync failure, a compact, a PTL, or a max_output_tokens exhaustion), the runtime recovery doctrine shall RECOMMEND (as policy guidance to agents and future hook authors) that Stop/PostToolUse hooks exit 0 (allow) rather than exit 2 (block), so that recovery turns are not placed into the `error → stop-hook-blocks → retry → error` death-spiral. This is a policy/doctrine recommendation — the current `sync-phase-quality-gate.sh` (Stop) and `status-transition-ownership.sh` (PostToolUse) hooks do not parse a recovery signal and therefore cannot mechanically enforce this carve-out; runtime-layer enforcement is deferred to a follow-up SPEC (future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001`). The carve-out is nonetheless documented in `agent-common-protocol.md` §Hook Invocation Surface as guidance, not as a mechanically-enforced gate.

**REQ-RR-007** (Ubiquitous, **documentation-only policy recommendation**): The runtime recovery doctrine shall document the anti-death-spiral carve-out as a policy recommendation (not a self-gating obligation on the current hooks, and not a mechanical guarantee). The doctrine rule is the SSOT; `agent-common-protocol.md` §Hook Invocation Surface is the render surface where the carve-out appears as guidance to agents and future hook authors. The heading literal "Recovery-Signal Carve-Out" MUST appear in both the SSOT rule and the render surface so that a future runtime-layer hook SPEC can locate and mechanically enforce the obligation without re-deriving it.

**REQ-RR-008** (Capability gate): **Where** the anti-death-spiral invariant applies, `zone-registry.md` shall expose a queryable `CONST-V3R6-0XX` entry so that `moai constitution list` returns the invariant alongside the other Frozen-zone safety clauses.

### §F.5 Cross-references and grep reproducibility

**REQ-RR-009** (Ubiquitous): The runtime recovery doctrine shall cross-reference `.claude/rules/moai/workflow/session-handoff.md` (rungs 2-3), `.claude/rules/moai/core/verification-claim-integrity.md` (narrative-consistency / invariant 5), book1 ch03 (query loop / input governance / withheld-recoverable errors), and book1 ch06 (cheapest-first ladder / circuit breakers / death-spiral) so the doctrine's lineage is traceable.

**REQ-RR-010** (Event-detected): **When** a maintainer greps `.claude/rules/moai/` for the recovery-ladder vocabulary that is genuinely absent at plan-phase baseline — `reactive-compact`, `death-spiral`, `withheld-recoverable`, `circuit-breaker` — the rule set under `.claude/rules/moai/` shall contain at least one hit per term, closing the recovery-ladder grep-zero-hit gap that motivated this SPEC. The error-type vocabulary (`max_output_tokens`, `media_size`, `compact-failure`) is explicitly EXCLUDED from this gap-closing claim: `max_output_tokens` already has pre-existing coverage in `hooks-system.md` as a `StopFailure` error type (L30/L105/L309), and the gap this SPEC closes is the *recovery-ladder* vocabulary specifically, not the error-type vocabulary.

### §F.6 agent-common-protocol.md scope boundary

**REQ-RR-011** (State-driven): **While** the run-phase is editing `.claude/rules/moai/core/agent-common-protocol.md`, the run-phase shall not add a `§Ledger Closure` heading (or any `Ledger`-containing heading) outside §Hook Invocation Surface. The `§Ledger Closure` reservation is owned by the P1a sibling SPEC `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001`; this SPEC's edit surface in `agent-common-protocol.md` is confined to §Hook Invocation Surface only.

**REQ-RR-012** (Ubiquitous): The runtime recovery doctrine rule shall state that a moai-adk agent observing a withheld-recoverable error mid-turn MUST consult the doctrine + apply the cheapest-first ladder (§F.2) BEFORE concluding the turn failed — so that the recovery obligation is normatively attached to the agent, not only to the doctrine document.

## §G. Constraints

- **Policy layer only**: no Go code under `internal/`, no hook script bodies modified.
- **SCOPE LIMIT on agent-common-protocol.md**: edit ONLY the §Hook Invocation Surface area. Do NOT touch §User Interaction Boundary; do NOT add a §Ledger Closure subsection (that is P1a SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001).
- **Book-grounding fidelity**: REQ/AC wording MUST stay faithful to book1 ch03/ch06 framing (withheld-recoverable, cheapest-first, death-spiral, narrative consistency). Do not paraphrase away the named principles.
- **Single CONST entry**: zone-registry.md gets exactly ONE new entry (`CONST-V3R6-0XX`). Do not batch multiple invariants into one entry; each invariant stays in the doctrine rule body.
- **Template neutrality**: the new rule and the edits introduce recovery vocabulary only; no moai-adk-internal SPEC IDs, no commit SHAs, no macOS-bias paths leak into `internal/template/templates/` (per CLAUDE.local.md §25).

## §H. Cross-References

- `.claude/rules/moai/workflow/session-handoff.md` — rungs 2 and 3 of the recovery ladder (paste-ready resume, worktree Block-0 restart).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the 5-Section Evidence-Bearing Report format = invariant 5 (narrative consistency across compact boundary).
- `.claude/rules/moai/core/agent-common-protocol.md` §Hook Invocation Surface — render surface for the anti-death-spiral carve-out (REQ-RR-006/007).
- `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R6-0XX` entry (REQ-RR-008).
- book1 (`github.com/wquguru/harness-books`): ch03 (query loop / input governance / withheld-recoverable), ch06 (cheapest-first / circuit breakers / death-spiral / narrative consistency), §9.3 / §9.6 / §9.7, ch06 §5.
- P1a sibling SPEC `SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001` is queued for Sprint 15 plan-phase (cohort task #2, plan-phase pending). The `§Ledger Closure` reservation in `agent-common-protocol.md` is owned by that SPEC; this SPEC deliberately excludes it to avoid file-edit collision (see REQ-RR-011). The §H prose cross-reference is load-bearing — there is intentionally no `related_specs:` frontmatter field (non-canonical; the prose link is the canonical cross-reference).
- Forward-link: future `SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001` (runtime-layer hook that parses `stopReason` to mechanically enforce the recovery-signal carve-out) — this SPEC defers mechanical enforcement to that follow-up (see REQ-RR-006 scope binding).

## §I. History

| Date | Version | Author | Change |
|------|---------|--------|--------|
| 2026-06-18 | 0.1.0 | manager-spec | Initial plan-phase draft — 10 REQs across 5 sub-sections; policy-layer doctrine grounded in book1 ch03/ch06; anti-death-spiral carve-out binds existing hooks. |
| 2026-06-18 | 0.2.0 | manager-spec | iter-2 plan-auditor defect fixes: D1 §E→Out of Scope + h3 sub-section (MissingExclusions ERROR); D2 `era: V3R6` frontmatter; D3 REQ-RR-010 reframed to recovery-ladder vocabulary only (max_output_tokens pre-existing in hooks-system.md); D4 REQ-RR-006/007 reframed as documentation-only policy recommendation (OPTION a, forward-link future SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001); D5 §H Sprint 15 P1a queue annotation; D6 removed non-canonical `related_specs:` field; D7 added REQ-RR-011 (scope boundary) + REQ-RR-012 (agent consult obligation); D8 REQ-RR-007 pins "Recovery-Signal Carve-Out" heading literal; D9 REQ-RR-010 system-actor subject; D10 plan.md §C M3 in-flight SPEC check. REQ count 10→12, AC count 10→11. |
