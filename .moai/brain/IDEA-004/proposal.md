# Proposal: Self-Evolving Harness v2 for MoAI-ADK
*Generated: 2026-05-14 | Idea: IDEA-004*

## Vision Statement

MoAI-ADK ships a project-tailored harness system that observes its own usage, learns from patterns across multiple Claude Code event types, self-critiques proposals against an explicit design constitution, and applies improvements autonomously within a five-layer safety stack — all without any binary CLI dependency. The harness becomes a living, principle-grounded improvement engine that compounds in capability over time while preserving the orchestrator's human-oversight contract.

## Product Summary

A unified, CLI-free, self-evolving harness system that consolidates three prior harness SPECs into a single V3R4 architecture. The harness lifecycle (generation → observation → learning → evolution → application) operates entirely through `/moai:harness` skill invocation plus dedicated subagents plus Claude Code hooks, eliminating dependence on Go binary subcommands. Learning combines verbal-reinforcement self-critique with constitution-based principle scoring, gated by a five-layer safety architecture and a human-oversight checkpoint at the final application step.

## Target User

MoAI-ADK + Claude Code development teams running long-lived projects with active SPEC workflows. Specifically: solo power-users operating multiple projects simultaneously, project tech leads owning 2-5 MoAI projects, and MoAI-ADK maintainers themselves who need the harness to self-improve so that recurring tooling work no longer requires manual specialist authoring. The product also serves MoAI projects themselves as customers — each project receives a project-instance harness, and the meta-harness orchestrator treats the project as client.

## Core Problems Solved

1. **CLI dependency creates a brittle invocation path.** The current Go binary harness subcommand is not wired into the v2.14.0 root command, causing slash-command failures and creating a single point of failure for the entire harness lifecycle.

2. **Static promotion thresholds cannot adapt to project-specific evolution rhythms.** Frequency-count detection over surface tool calls misses semantic equivalences and produces both false positives (noise patterns) and false negatives (real patterns masked by surface variation).

3. **Learning is intra-project siloed and one-directional.** Observations from project A never benefit project B; successful evolutions never feed back into harness generation; every project starts from the same seed with no compounding effect.

## Proposed Solution

Top capabilities, described at the capability level without implementation prescription:

- Multi-event observation collection covering tool-use, session boundaries, subagent completion, and user-prompt events for full lifecycle signal coverage.
- Embedding-cluster pattern detection replacing frequency-count thresholds, enabling semantic-equivalence recognition across surface variations.
- Verbal-reinforcement self-critique loop with hard iteration cap, producing natural-language lessons that persist in episodic memory.
- Principle-based self-scoring against the project's explicit design constitution, pre-screening proposals before human oversight to reduce reviewer fatigue.
- Auto-organizing skill library with embedding-indexed retrieval for novel scenario coverage and compositional skill reuse.
- Multi-objective effectiveness measurement across quality, token cost, latency, and iteration count, with automatic rollback on any-axis regression.
- Effectiveness-decay pruning so stale patterns deprecate automatically and the skill library does not bloat monotonically.
- Optional, opt-in, anonymized cross-project lesson federation with strict default-off and namespace isolation.

## Recommended Execution Order

1. SPEC-V3R4-HARNESS-EVOLVE-001 first (foundation: unified architecture + CLI retirement)
2. SPEC-V3R4-HARNESS-OBSERVER-DIVERSITY-001 (expand signal sources before downstream consumers depend on them)
3. SPEC-V3R4-HARNESS-CLUSTER-DETECT-001 (semantic detection replaces frequency-count in the promotion pipeline)
4. SPEC-V3R4-HARNESS-REFLECTION-001 (self-critique loop layered on top of detection)
5. SPEC-V3R4-HARNESS-MULTI-OBJECTIVE-001 (multi-axis scoring + auto-rollback)
6. SPEC-V3R4-HARNESS-SKILL-LIBRARY-001 (Voyager-style skill organization with embedding retrieval)
7. SPEC-V3R4-HARNESS-CROSS-PROJECT-001 last (privacy-sensitive, deferred until earlier SPECs validate the core loop)

Rationale: each downstream SPEC depends on the foundation laid by the previous one. EVOLVE-001 establishes the architectural baseline (CLI retirement, unified skill structure, 5-Layer Safety preservation). OBSERVER-DIVERSITY widens the input. CLUSTER-DETECT and REFLECTION upgrade the reasoning core. MULTI-OBJECTIVE adds the safety net. SKILL-LIBRARY compounds the learning. CROSS-PROJECT is deferred because privacy concerns require validated single-project track record first.

## Out of Scope (v0.1)

- Migration tooling to import historical `usage-log.jsonl` data into the new embedding-cluster format (deferred; users start fresh with v2)
- GUI/dashboard for evolution-history inspection (text artifacts in `.moai/harness/` remain the interface)
- Real-time evolution streaming (batch evolution cycles continue, latency requirements remain measured in days not seconds)
- Mobile or web client for AskUserQuestion approval flow (terminal-based Claude Code interaction only)
- Multi-tenant SaaS hosting of cross-project federation (lesson sharing is peer-to-peer via auto-memory, not a hosted service)
- Cross-organization sharing (federation is bounded to a user's own opt-in projects in v0.1)

### SPEC Decomposition Candidates

- SPEC-HARNESS-001: Unified self-evolving harness foundation — supersedes V3R3 harness SPECs, retires CLI subcommand path, consolidates lifecycle into skill + subagent + hooks
- SPEC-HARNESS-002: Multi-event observer pipeline — capture signals across tool-use, session boundary, subagent completion, and user-prompt events with unified log format
- SPEC-HARNESS-003: Embedding-cluster pattern detector — semantic-equivalence detection replacing frequency-count thresholds in the promotion pipeline
- SPEC-HARNESS-004: Verbal-reinforcement self-critique loop — Reflexion-style iterative reflection capped at three iterations, with episodic memory of natural-language lessons
- SPEC-HARNESS-005: Principle-based self-scoring against design constitution — pre-screen evolution proposals against project's explicit principles before human oversight
- SPEC-HARNESS-006: Multi-objective effectiveness measurement with auto-rollback — quality + cost + latency + iteration tuple scoring; regression on any axis triggers reversal
- SPEC-HARNESS-007: Embedding-indexed skill library with retrieval-augmented generation — auto-organize specialist skills, retrieve top-K relevant skills for novel scenarios
- SPEC-HARNESS-008: Opt-in cross-project lesson federation — anonymized lesson sharing via auto-memory namespace, strict default-off, namespace isolation per organization

## Notes

- The +60% effectiveness target reported by the reference framework should be treated as aspirational rather than baseline. V0.1 success is defined as positive evolution-application count per week per project combined with rollback rate below 10%.
- Conflict between Reflexion self-critique and evaluator-active scoring resolves with evaluator-active retaining absolute veto power. Self-critique runs first as a pre-screen; evaluator-active runs second as the binding gate. This ordering must be locked into SPEC-HARNESS-004 and SPEC-HARNESS-005.
- AskUserQuestion fatigue mitigation starts with a conservative rate limit (one Tier-4 application per week per project) that widens adaptively based on user acceptance rate.
- Embedding model versioning is identified as a risk; clusters must persist embedding-model identifiers and emit warnings on mismatch after model upgrades.
- Cross-project federation has open privacy concerns even with anonymization. SPEC-HARNESS-008 is intentionally deferred until SPEC-HARNESS-001 through SPEC-HARNESS-007 demonstrate single-project track record over multiple release cycles.
- All FROZEN zones in the design constitution remain immutable. Evolution operates strictly within EVOLVABLE zones. This is not negotiable and must be re-asserted in every SPEC's [HARD] rule section.
