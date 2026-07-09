---
paths: "**/runtime-recovery-doctrine.md"
---

# Runtime Recovery Doctrine

> **Single source of truth** for runtime-recovery policy when the loop itself fails mid-SPEC.
> Cross-referenced by: `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface (render surface for the anti-death-spiral carve-out), the constitution registry (anti-death-spiral entry), `.claude/rules/moai/workflow/session-handoff.md` (recovery ladder rungs 2-3), `.claude/rules/moai/core/verification-claim-integrity.md` (narrative-consistency invariant 5).
>
> **Policy-layer only.** This doctrine is normative guidance for moai-adk agents and FUTURE hook authors. It does NOT reimplement any Claude Code TypeScript internal (`queryLoop`, `recoverFromError`, `truncateHeadForPTLRetry`, `hasAttemptedReactiveCompact`). moai-adk is a harness ON TOP of Claude Code and cannot modify the native query loop. Grounding: book1 (internal reference: harness-books book1) ch03 "Query Loop is the heartbeat" + ch06 "Errors and recovery".

---

## §1. The Withheld-Recoverable Error Set

[ZONE:Evolvable] [HARD] The runtime recovery doctrine names the **withheld-recoverable-error set** as:

```
{ prompt_too_long (PTL), max_output_tokens, media_size, compact-failure }
```

These four error types are **WITHHELD** and routed to layered recovery (the cheapest-first ladder in §2) BEFORE they are surfaced to the user. An agent observing one of these errors mid-turn MUST NOT treat the error as a terminal failure; it MUST consult this doctrine and apply the ladder (§6 agent obligation).

This set is grounded in book1 ch03's framing of the `queryLoop()`'s first duty — input governance (memory prefetch → snip → microcompact → context-collapse → autocompact) BEFORE the model call — during which recoverable errors are withheld and routed to layered recovery rather than surfaced raw. The four named error types are the moai-adk policy projection of that input-governance duty onto observable mid-SPEC signals.

The withheld property is normative, not mechanical: the moai-adk agent's obligation is to recognize these signals and apply the ladder, not to intercept them at a platform layer (that interception lives inside Claude Code and is out of moai-adk's reach per the policy-layer constraint above).

**Convergent second source (graduated-compaction naming).** The book1 input-governance sequence above (`memory prefetch → snip → microcompact → context-collapse → autocompact`) is independently corroborated by a **convergent second source**: the public paper "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (arXiv:2604.14228; companion repository github.com/VILA-Lab/Dive-into-Claude-Code) names the same graduated-compaction concept as five escalating layers — `Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact`. The two sources map onto each other: book1's `snip → microcompact → context-collapse → autocompact` align with the paper's `Snip → Microcompact → Context Collapse → Auto-Compact`, and the paper's leading `Budget Reduction` layer corresponds to the budget-reduction step book1 folds into its `memory prefetch` input-governance preamble. moai-adk CONSUMES this Claude Code graduated compaction; it does NOT implement it (the interception lives inside Claude Code's `queryLoop()`, out of moai-adk's reach per the policy-layer constraint above). Recording the paper's exact layer names is provenance enrichment — it adds a second citation for the same concept and changes no behavior.

---

## §2. The 4-Rung Cheapest-First Recovery Ladder

[ZONE:Evolvable] [HARD] The runtime recovery doctrine defines a **4-rung cheapest-first recovery ladder**. Each rung is mapped to a concrete moai-adk artifact so that no rung is introduced without a recovery path. **Cheapest-first** means: try the lowest-cost rung before escalating to a higher-cost rung. No rung may be skipped while a lower rung remains unattempted for the current failure (the ordering rule in §2.1 below).

| Rung | Recovery action | moai-adk artifact cross-reference |
|------|-----------------|-----------------------------------|
| 1 | **In-turn self-correction** — continue working without recap or apology. The agent reasons over what it just did, corrects the approach, and resumes. (book1 ch06 meta-message: recovery's goal is to *keep working*, not to apologize.) | (agent behavior; no separate artifact) |
| 2 | **Paste-ready resume + `/clear`** — emit a paste-ready resume message, then `/clear` to reset the context window. | `.claude/rules/moai/workflow/session-handoff.md` § Canonical Format (the 6-block structure) |
| 3 | **Session-handoff Block-0 worktree restart** — when the failure is structural (stale worktree base, L1 worktree divergence, accumulated compact residue), restart inside a fresh worktree using the Block-0 anchored resume. | `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored Resume Pattern |
| 4 | **Abort + preserve** — the last-resort escape. Persist all in-flight state to `progress.md` (PRESERVE-list discipline), close the in-flight ledger (invariant 4 below), and end the session so the next session resumes rather than restarts. | `.moai/specs/<SPEC-ID>/progress.md` + §4 abort-closes-ledger below |

### §2.1 Cheapest-first ordering rule

[ZONE:Evolvable] [HARD] **While** a lower rung has not been attempted for the current failure, the agent SHALL NOT jump to a higher-cost rung. Concretely:

- No `/clear` (rung 2) before in-turn self-correction (rung 1) has been attempted.
- No worktree restart (rung 3) before `/clear` (rung 2) has been attempted.
- No abort + preserve (rung 4) before the lower rungs have been attempted — EXCEPT when the compact-can-PTL last-resort escape (invariant 3) fires: when reactive-compact itself PTLs, the agent MUST fall back directly to rung 4 rather than loop.

The ordering is normative (it binds the agent's choice), not mechanical. No runtime check enforces it at the policy layer, and per the policy-layer constraint, none is wanted here — the native query loop owns the mechanical ladder inside Claude Code.

---

## §3. The 5 Circuit-Breaker Invariants

[ZONE:Evolvable] [HARD] The runtime recovery doctrine codifies book1 ch06 §5's five **circuit-breaker invariants** as moai-adk-level policy. Each invariant is stated as a rule the agent (or, for invariant 1's anti-death-spiral projection, a FUTURE runtime-layer hook) must respect. These invariants prevent the recovery procedure itself from becoming a death-spiral.

1. **Max-consecutive-autocompact-failure analogue** — after **3 consecutive recovery failures at the same rung** (e.g., three autocompact/PTL recoveries that each re-fail), the agent MUST escalate to the next rung rather than retry the same rung. book1 ch06 §5 names this as `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES = 3`; the moai-adk projection generalizes from autocompact-specific to same-rung recovery failures of any kind.

2. **`hasAttemptedReactiveCompact` analogue (no self-loop)** — a recovery action already attempted for the current failure MUST NOT be re-attempted in the same turn. The agent MUST advance to the next rung. book1 ch06 §5 names this as the `hasAttemptedReactiveCompact` guard; the moai-adk projection generalizes from reactive-compact-specific to any recovery action already attempted this turn. Re-attempting the same recovery in the same turn is the canonical shape of a death-spiral.

3. **Compact-can-PTL last-resort escape** — when reactive-compact itself PTLs (the recovery action triggers the very error it is trying to recover from), the agent MUST fall back to the last-resort truncation rung (rung 4 — abort + preserve), NOT loop on another compact. book1 ch06 §5 names this as `truncateHeadForPTLRetry()` as the last-resort escape hatch; the moai-adk projection is the rung-4 abort + preserve action, since moai-adk cannot call the platform's `truncateHeadForPTLRetry` directly.

4. **Abort-closes-ledger** — when recovery aborts (rung 4), the agent MUST close the in-flight ledger by persisting state to `progress.md` before the session ends, so the next session **resumes** rather than **restarts**. book1 ch06 §5 names this as the abort-closes-ledger invariant; the moai-adk projection maps the ledger to `progress.md` (§E run-phase evidence + PRESERVE-list discipline). An abort that leaves `progress.md` stale forces the next session to rediscover the in-flight state — a silent restart disguised as a resume.

5. **Narrative-consistency requirement** — across the compact/recovery boundary, the agent MUST preserve narrative consistency via the 5-Section Evidence-Bearing Report format. book1 ch06 §5 names this as the narrative-consistency requirement; the moai-adk projection cross-references `.claude/rules/moai/core/verification-claim-integrity.md` § The 5-Section Evidence-Bearing Report Format. The report MUST state: what was attempted, why it failed, what recovery was used, and what remains a gap. Without this, a compact boundary silently drops the agent's account of what it tried, and the next turn (or next session) re-attempts a recovery already known to fail — invariant 2's death-spiral shape, caused by a missing narrative rather than a missing guard.

---

## §4. Anti-Death-Spiral Hook Carve-Out (documentation-only policy)

> **Scope binding (LOAD-BEARING)**: This subsection binds **AGENT BEHAVIOR** and **FUTURE hook evolution**. It does NOT bind the current hook scripts (`sync-phase-quality-gate.sh`, `status-transition-ownership.sh`), which receive PostToolUse/Stop JSON but do not parse a recovery signal; no mechanical enforcement is possible without a runtime-layer hook that parses `stopReason`, which is deferred to a follow-up runtime-layer SPEC. This doctrine does NOT rewrite the hook scripts.

### Recovery-Signal Carve-Out

[ZONE:Evolvable] [HARD] **Recovery-Signal Carve-Out.** **While** a turn's `stopReason` or surrounding context indicates the turn is itself a **recovery signal** — i.e., the turn is recovering from a sync failure, a compact, a `prompt_too_long` (PTL), a `max_output_tokens` exhaustion, or a `media_size`/`compact-failure` — Stop/PostToolUse hooks SHOULD exit 0 (allow the turn to end / the tool call to proceed) rather than exit 2 (block), so that recovery turns are NOT placed into the `error → stop-hook-blocks → retry → error` loop that book1 ch06 names the **death-spiral** and explicitly warns recovery MUST skip stop-hooks to avoid.

This carve-out is stated as **policy guidance** (a SHOULD recommendation) to agents and FUTURE hook authors, NOT as a mechanically-enforced gate:

- **Current hooks do NOT mechanically enforce this.** The current `sync-phase-quality-gate.sh` (Stop hook) and `status-transition-ownership.sh` (PostToolUse hook) receive PostToolUse/Stop JSON but do not parse a recovery signal from `stopReason` or turn context; they therefore cannot mechanically distinguish a recovery turn from a normal turn and cannot mechanically exit 0 on recovery turns.
- **Runtime-layer enforcement is deferred.** A future runtime-layer SPEC MAY propose a hook that parses `stopReason` to mechanically enforce this carve-out. This doctrine deliberately defers mechanical enforcement to that follow-up; the acceptance criterion for the carve-out is SHOULD and documentation-only precisely because mechanical enforcement is not possible at this layer.

### Both hooks named

The two current hooks the carve-out addresses are:

- `sync-phase-quality-gate.sh` — Stop hook, fires on every turn-end; per the Stop self-gate caveat in `agent-common-protocol.md` § Hook Invocation Surface, it inspects conversation/working-tree state to decide whether the turn is a genuine completion point. On a recovery turn, it SHOULD self-gate to exit 0.
- `status-transition-ownership.sh` — PostToolUse hook on Write/Edit of SPEC body content; on a recovery turn that writes a SPEC artifact as part of recovery (e.g., persisting state to `progress.md` per invariant 4), it SHOULD exit 0 rather than block the recovery write.

The carve-out's heading literal **"Recovery-Signal Carve-Out"** is pinned to appear in BOTH this SSOT rule (§4) and the render surface (`.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface) so that a future runtime-layer hook SPEC can locate the obligation mechanically (via grep) without re-deriving it.

### Why this is a carve-out and not a gate removal

The carve-out does NOT weaken the hooks' gate function on non-recovery turns. The Stop/PostToolUse gates still fire exit 2 (block) on genuine gate failures (lint, test, ownership mismatch) during normal turns. The carve-out only says: when the turn is a recovery turn (the turn's own purpose is to recover from a withheld-recoverable error), the gate SHOULD defer to the recovery rather than block it — because blocking a recovery turn is the textbook shape of the death-spiral. Determining "is this a recovery turn?" is the mechanical step the current hooks cannot take; that determination is what the deferred runtime-layer SPEC would add.

---

## §5. Cross-References (lineage traceability)

[ZONE:Evolvable] This doctrine's lineage is traceable through the following cross-references:

- `.claude/rules/moai/workflow/session-handoff.md` — recovery ladder rungs 2 and 3 (paste-ready resume + `/clear`; worktree-anchored Block-0 restart). The session-handoff rule is the artifact that makes rungs 2-3 actionable.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the 5-Section Evidence-Bearing Report Format = circuit-breaker invariant 5 (narrative consistency across the compact/recovery boundary).
- `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface — render surface for the §4 anti-death-spiral carve-out (the carve-out appears in BOTH this SSOT rule and that render surface).
- the constitution registry — the anti-death-spiral entry exposes the anti-death-spiral invariant to `moai constitution list`.
- book1 (internal reference: harness-books book1):
  - **ch03 "Query Loop is the heartbeat"** — input governance, withheld-recoverable errors (`{PTL, max_output_tokens, media_size}`), layered recovery BEFORE the model call. Source for §1.
  - **ch06 "Errors and recovery"** — "错误路径就是主路径" (the error path IS the main path); "恢复的目标是继续工作" (recovery's goal is to keep working, not to apologize); cheapest-first layered recovery; `hasAttemptedReactiveCompact` self-loop guard; autocompact circuit breaker `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES = 3`; the death-spiral hazard; narrative consistency. Source for §2, §3, §4.
  - **§9.3 / §9.6 / §9.7** — the named principles (recoverable error withholding, cheapest-first ladder, narrative consistency).
  - **ch06 §5** — the 5-line circuit-breaker invariants (source for §3's five invariants).
- The sibling interrupt-ledger SPEC owns the `§Ledger Closure` subsection in `agent-common-protocol.md`; this doctrine deliberately excludes that subsection. The two carve-outs are sibling subsections under § Hook Invocation Surface, not nested.
- Forward-link: a future runtime-layer hook SPEC would parse `stopReason` to mechanically enforce the §4 carve-out. This doctrine defers mechanical enforcement to that follow-up.

---

## §6. Agent Consult-the-Doctrine Obligation

[ZONE:Evolvable] [HARD] A moai-adk agent observing a withheld-recoverable error (§1) mid-turn MUST consult this doctrine and apply the cheapest-first ladder (§2) BEFORE concluding the turn failed. The recovery obligation is normatively attached to the **agent**, not only to this doctrine document: a turn that hits a PTL, a `max_output_tokens` exhaustion, a `media_size` failure, or a `compact-failure` is not a failed turn — it is a turn that the agent is obliged to recover via the ladder before it may report failure. Concluding the turn failed without consulting the doctrine + applying the ladder violates this obligation.

This obligation is the moai-adk projection of book1 ch06's "错误路径就是主路径" (the error path IS the main path): the error is not an exception to the workflow, it is part of the workflow, and the agent's job is to route it through the ladder rather than surface it raw.

---

## §7. Anti-Patterns

- **AP-RR-001 — Go runtime creep**: introducing `internal/recovery/` or similar. OUT OF SCOPE per the policy-layer constraint. If a runtime-layer need emerges, return a blocker and spawn a separate SPEC.
- **AP-RR-002 — Re-attempting the same recovery in the same turn**: violates invariant 2 (`hasAttemptedReactiveCompact` no-self-loop). Advance to the next rung instead.
- **AP-RR-003 — Aborting without closing the ledger**: violates invariant 4 (abort-closes-ledger). An abort MUST persist state to `progress.md` before the session ends.
- **AP-RR-004 — Paraphrasing away book1 named principles**: "withheld-recoverable", "cheapest-first", "death-spiral", "narrative consistency", `hasAttemptedReactiveCompact`, `truncateHeadForPTLRetry`, `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES` are load-bearing terms from book1 ch03/ch06. Preserve them verbatim.
- **AP-RR-005 — Treating the carve-out as a gate removal**: the §4 carve-out does NOT weaken the hooks' gate function on non-recovery turns; it only says recovery turns SHOULD defer to the recovery. Determining "is this a recovery turn?" is the mechanical step deferred to the runtime-layer follow-up SPEC.
- **AP-RR-006 — Claiming mechanical enforcement this layer cannot provide**: the current hooks do not parse `stopReason`; do NOT claim or imply that the §4 carve-out is mechanically enforced today. State honestly that enforcement is documentation-only and deferred.

---

Status: Active — single source of truth for runtime-recovery policy; binds agent behavior and future hook authors.
