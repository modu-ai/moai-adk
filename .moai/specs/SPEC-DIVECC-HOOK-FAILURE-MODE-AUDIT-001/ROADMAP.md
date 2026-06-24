# Epic Dive-into-CC — Roadmap

> **Epic identifier**: `Epic Dive-into-CC`
> **SPEC-ID domain token**: `DIVECC`
> **Entry SPEC**: `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001` (N1)
> **Status**: plan-phase (draft)
> **Taxonomy**: per `.claude/rules/moai/development/sprint-round-naming.md`, an Epic is a thematic multi-SPEC container. This Epic uses a thematic name (release-cut / dependency-batch grouping), not a sequence number.

---

## Epic Origin

This Epic applies findings from an academic reverse-engineering analysis of Claude Code to moai-adk's own harness — a **dogfooding** (self-improvement) exercise. moai-adk IS a Claude Code harness, so the analysis both validates moai-adk's architectural direction and surfaces concrete improvement candidates.

Source body of work (one publication, two surfaces):

- **arXiv:2604.14228** — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (Liu, Zhao, Shang, Shen, 2026, cs.SE). A reverse-engineering of Claude Code v2.1.88 internals.
- **github.com/VILA-Lab/Dive-into-Claude-Code** — companion repository + "Build Your Own AI Agent: A Design Guide".

Central thesis: **"98.4% infrastructure, 1.6% AI"** — the agent loop is a trivial `while`-loop; the harness is the differentiator. The paper's open-direction findings map onto moai-adk's harness surfaces (hooks, extension mechanisms, delegation cost model, observability, runtime inventory).

> **Note on the central-thesis figure.** The "98.4% / 1.6%" split and the paper's open-direction enumeration are reproduced here as the paper's own claims (the Epic's external grounding). They are NOT moai-adk verification claims and require no re-grounding by moai-adk tooling. Only the per-SPEC premises that assert facts about moai-adk's own tree (e.g. N1's shared-failure-mode evidence) are subject to the verification-claim-integrity invariant and must be grounded against the moai-adk repo before a SPEC's run-phase.

---

## Candidate Roadmap (N1–N7)

The table is the at-a-glance index; per-candidate detail follows below it.

| ID | Priority | Intent (one line) | Target artifact | Tier | Premise | Ordering |
|----|----------|-------------------|-----------------|------|---------|----------|
| N1 | HIGH | Hook defense-in-depth shared-failure-mode audit + independence rule | new rule `.claude/rules/moai/development/hook-independence.md` + audit of `.claude/hooks/moai/` | **M** | **VERIFIED** (evidence in N1 research.md) | Entry SPEC — no deps |
| N2 | HIGH | Graduated context-cost decision criterion for extension authoring | `.claude/rules/moai/development/agent-authoring.md` (+ builder-harness policy) | **S** | needs light grounding (paper claim, no moai-tree assertion) | independent of N1 |
| N3 | HIGH | Add token-cost (~7×) signal to the delegation decision | `.claude/output-styles/moai/moai.md` §4 + CLAUDE.md §16 self-check | **S** | needs light grounding (paper claim) | independent; pairs with N2 |
| N4 | MED | Observability-closes-the-loop (trace→eval→cluster→policy→repair) | `moai-harness-learner` subsystem | **M/L** | needs research-phase grounding | after N1 (hook/observability context) |
| N5 | MED | 5-layer compaction naming cross-reference (DOC-ALIGNMENT only) | `context-window-management.md` + `runtime-recovery-doctrine.md` | **S** | VERIFIED-by-citation (paper names the layers) | independent |
| N6 | MED | Unified agent/harness inventory view (runtime-is-first-class) | new CLI surface (`moai session list` / worktrees / `/harness:devkit list` unification) | **M** | needs research-phase grounding | after N1/N4 |
| N7 | LOW | Archive the paper as a CC-internals reference | `.moai/research/` + rule cross-ref | **S** | trivial (archival) | independent; can land anytime |

Tier legend: **S** = spec.md + plan.md + acceptance.md (3 files); **M/L** = adds design.md and/or research.md.

Priority legend: HIGH / MED / LOW as supplied by the orchestrator's mapping.

---

### N1 — Hook defense-in-depth shared-failure-mode audit `[HIGH]` `[Tier M]`

- **Intent**: Audit every shared failure mode across moai-adk's hook layer, classify each as acceptable-by-design vs genuine-risk, and author an independence rule documenting the shared-failure-mode catalogue plus a "does this new hook introduce an independent failure mode?" authoring checklist.
- **Target artifact**: new rule `.claude/rules/moai/development/hook-independence.md` (+ its template mirror under `internal/template/templates/.claude/rules/...`) + the audit recorded in the SPEC's research.md.
- **Tier rationale (M)**: genuine audit/research component (read all 34 hook scripts, classify N shared dependencies); produces a cross-cutting new doctrine (not a single-line edit); the mitigation-recommendation work requires design reasoning. More than Tier S (single-doc, no research), but does not touch Go code and does not require multi-module architecture design, so not Tier L. → 6-file set.
- **Premise**: **VERIFIED** by orchestrator read-only grounding and independently reproduced during this plan-phase (see N1 `research.md` §A for the verbatim grep/read evidence). The paper insight "Defense-in-depth fails when layers share failure modes" is made concrete by 31/34 wrappers sharing one binary-resolution chain and 3/3 governance gates sharing one bypass flag.
- **Dependencies / ordering**: none — entry SPEC. N1 establishes the hook/observability vocabulary that N4 and N6 build on.

### N2 — Graduated context-cost decision criterion for extension authoring `[HIGH]` `[Tier S]`

- **Intent**: Document the extension-mechanism context-cost ladder — Hooks (zero) → Skills (low) → Plugins (medium) → MCP (high) — as a decision criterion: choosing an extension mechanism is a context-cost decision.
- **Target artifact**: `.claude/rules/moai/development/agent-authoring.md` and/or a builder-harness policy section.
- **Tier rationale (S)**: doc/rule-only addition; parallels the existing skill 3-level progressive-disclosure taxonomy and the dynamic-workflows purpose→effort taxonomy — adds a third parallel "mechanism→context-cost" axis. No code, no audit.
- **Premise**: paper claim about extension-mechanism cost. No assertion about the moai-adk tree, so only light grounding (confirm the ladder against moai-adk's actual extension surfaces) is needed at authoring time — not a research-phase blocker.
- **Dependencies / ordering**: independent of N1; thematically pairs with N3 (both refine the cost model of MoAI's own extension/delegation surfaces). May be co-authored with N3.

### N3 — Token-cost (~7×) signal in the delegation decision `[HIGH]` `[Tier S]`

- **Intent**: Add a token-cost signal to the delegation decision. Paper insight: `SkillTool` injects into the current context (cheap) vs `AgentTool` spawns an isolated context (~7× tokens). Add the directive "prefer Skill injection when shared context is acceptable; spawn an Agent only when isolation is genuinely needed."
- **Target artifact**: `.claude/output-styles/moai/moai.md` §4 (Delegation Decision) + CLAUDE.md §16 self-check (four self-check questions).
- **Tier rationale (S)**: doc/rule-only. The current §4 weighs quality/independence/bias but carries no token-cost signal; this adds one axis.
- **Premise**: paper claim about Skill-vs-Agent token cost. The ~7× figure is the paper's measurement of Claude Code internals, reproduced as the paper's claim (not a moai-adk verification claim). Light grounding only — confirm the §4 / §16 surfaces still read as described before editing.
- **Dependencies / ordering**: independent of N1; pairs with N2.

### N4 — Observability-closes-the-loop `[MED]` `[Tier M/L]`

- **Intent**: Close the harness-learning loop: traces → evaluation → failure-clustering → policy → prompt/tool repair (paper open-direction 6). moai-adk currently does manual failure clustering via `feedback_*.md` lessons; there is no structured trace→harness-learning auto loop.
- **Target artifact**: the `moai-harness-learner` subsystem (touches Go code).
- **Tier rationale (M/L)**: touches Go code (the learner subsystem) and introduces a new data flow (trace ingestion → clustering → policy emission). Tier to be finalized at its own plan-phase after research-phase grounding determines blast radius; M if it layers on existing learner hooks, L if it introduces a new pipeline.
- **Premise**: needs research-phase grounding — the current learner subsystem's trace surface and the feasibility of an auto-loop must be read and confirmed before scoping.
- **Dependencies / ordering**: after N1 (N1 establishes the hook/observability vocabulary; the learner consumes hook-emitted observability taps). Lower priority / longer horizon.

### N5 — 5-layer compaction naming cross-reference `[MED]` `[Tier S]`

- **Intent**: Add the paper's exact compaction-layer names — Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact — as a cross-reference to enrich provenance. moai-adk **consumes** (does not implement) Claude Code compaction, so this is **DOC-ALIGNMENT only**.
- **Target artifact**: `.claude/rules/moai/workflow/context-window-management.md` + `.claude/rules/moai/workflow/runtime-recovery-doctrine.md`.
- **Tier rationale (S)**: cross-reference enrichment only; no code, no audit. `runtime-recovery-doctrine.md` already cites `wquguru/harness-books`; the VILA-Lab paper is a convergent second source for the same graduated-compaction concept.
- **Premise**: VERIFIED-by-citation — the paper names the five layers explicitly; the work is to record the names, not to assert behavior about the moai-adk tree.
- **Dependencies / ordering**: independent. Note the boundary: N5 must NOT imply moai-adk implements these layers (it consumes CC's compaction).

### N6 — Unified agent/harness inventory view `[MED]` `[Tier M]`

- **Intent**: Make durable execution, checkpoints, and agent inventory user-visible (paper open-direction 1: runtime-is-first-class). moai-adk has `moai session list`, worktrees, and `/harness:devkit list` partially; there is no unified inventory.
- **Target artifact**: a new/unified CLI surface composing `moai session list` + worktree state + harness inventory.
- **Tier rationale (M)**: CLI work (Go code) that composes existing inventory surfaces into one view; bounded but non-trivial.
- **Premise**: needs research-phase grounding — the three existing inventory surfaces (`moai session list`, worktree registry, `/harness:devkit list`) must be read to confirm what a unified view can compose.
- **Dependencies / ordering**: after N1 and N4 (inventory should surface hook/observability state that N1 and N4 define). Lower priority.

### N7 — Archive the paper as a CC-internals reference `[LOW]` `[Tier S]`

- **Intent**: Archive the paper as an authoritative Claude Code v2.1.88 architecture reference. moai-adk's release-update tracker tracks CC versions; this paper is a durable internals reference worth cross-linking.
- **Target artifact**: `.moai/research/` archive entry + a rule cross-reference.
- **Tier rationale (S)**: trivial archival + one cross-reference line.
- **Premise**: trivial (archival). No grounding needed beyond confirming the citation.
- **Dependencies / ordering**: independent; can land at any point in the Epic.

---

## Suggested Execution Order

The Epic has no hard cross-SPEC blocking dependencies except N4/N6 benefiting from N1's hook/observability vocabulary. A defensible order by priority then dependency:

1. **N1** (HIGH, premise verified) — entry SPEC, establishes hook/observability vocabulary.
2. **N2 + N3** (HIGH, doc-only, pair) — extension/delegation cost model; co-authorable.
3. **N5** (MED, doc-alignment) — compaction naming cross-reference; independent, cheap.
4. **N7** (LOW, archival) — can land opportunistically alongside any of the above.
5. **N4** (MED, M/L, Go) — observability loop; after N1.
6. **N6** (MED, M, Go/CLI) — unified inventory; after N1/N4.

> Per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation, this order uses priority + dependency ordering only — no time predictions.

---

## Premise-Verification Status Summary

| Status | Candidates | Meaning |
|--------|-----------|---------|
| VERIFIED (moai-tree evidence) | N1 | Shared-failure-mode evidence reproduced against the moai-adk repo this plan-phase. |
| VERIFIED-by-citation | N5 | The paper names the artifact (compaction layers) explicitly; no moai-tree assertion. |
| Light grounding at authoring | N2, N3 | Paper claims about CC internals; confirm the target moai-adk surface still reads as described before editing. |
| Needs research-phase grounding | N4, N6 | Touch Go code / new data flow; blast radius must be read and confirmed at each SPEC's own plan-phase. |
| Trivial | N7 | Archival; no grounding beyond citation confirmation. |

---

## Cross-References

- `.claude/rules/moai/development/sprint-round-naming.md` — Epic taxonomy (thematic multi-SPEC container).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 — defect/premise claims must be grounded against the domain's tooling before being asserted as fact (binds N1, N4, N6 moai-tree premises).
- `SPEC-DIVECC-HOOK-FAILURE-MODE-AUDIT-001/` — N1 entry SPEC full artifact set.

---

_Roadmap authored at plan-phase. Each candidate beyond N1 is authored as its own SPEC when the orchestrator promotes it; this roadmap is the Epic's planning index, not an implementation artifact._
