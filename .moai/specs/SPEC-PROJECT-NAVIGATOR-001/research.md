# Research — SPEC-PROJECT-NAVIGATOR-001

> Web grounding (with URLs) + prior-work review. Used at plan-phase; cited by spec.md §B and plan.md §C/§E.

## §1. Web grounding (external research)

### Aider — repo map via tree-sitter + PageRank

- **Source**: https://aider.chat/2023/10/22/repomap.html , https://aider.chat/docs/repomap.html
- **Claim**: Aider builds a repository map by parsing the codebase with tree-sitter, constructing a graph of symbol definitions and references, and running PageRank to rank the most important symbols. Only the highest-ranked symbols are delivered to the LLM as context for the current edit.
- **Implication for Navigator**: confirms that a MAP reduces context rather than inflating it; the map should answer task-shaped questions (what is done, what is next), not be an exhaustive dump. Navigator's rollup-reference design (REQ-PN-013 non-duplication) aligns. The PageRank symbol-ranking idea is deferred to SPEC-003 (tree-sitter auto-derivation), NOT used by 001.

### Codebase-Memory (arXiv 2603.27277)

- **Source**: https://arxiv.org/html/2603.27277v1
- **Claim**: tree-sitter + MCP knowledge graph persisted in a single SQLite file; survives sessions; reported ~10× token reduction and ~2.1× fewer tool calls vs file exploration.
- **Implication for Navigator**: confirms the value of PERSISTENCE across sessions — the central gap the Navigator closes. SQLite-per-project is a possible future evolution of `.moai/state/navigator/` if markdown rollups become a bottleneck; for 001 we keep markdown (human-readable, diff-friendly, no new binary dependency).

### "Coding Agents Need Codebase Maps, Not Bigger Prompts"

- **Source**: https://www.developersdigest.tech/blog/codebase-knowledge-graphs-ai-coding-agents
- **Claim**: a good map REDUCES context rather than inflating it; the map must answer task-shaped questions; every claim needs provenance (commit hash + timestamp).
- **Implication for Navigator**: directly shapes REQ-PN-002 (provenance) — commit-sha + captured-at on every row — and REQ-PN-013 (non-duplication: rollup references, not content copies). Also shapes the "answer task-shaped questions" framing of `navigator.md` (current frontier + next task, not an exhaustive inventory dump).

### AGENTS.md standard

- **Source**: https://agents.md/ , https://github.com/agentsmd/agents.md , impact measured at https://arxiv.org/abs/2601.20404
- **Claim**: a single predictable place for agent instructions; nested/hierarchical; measurable impact on agent task completion.
- **Implication for Navigator**: `navigator.md` is the project-state analogue of `AGENTS.md`'s instruction-state role — a single predictable entry point. The Navigator does NOT replace `AGENTS.md` (which is instruction-shaped); it complements it with state-shaped rollup.

### MemGPT / Letta, Mem0 — persistent cross-session memory

- **Source**: https://mem0.ai/blog/ai-agent-platforms-with-persistent-memory
- **Claim**: persistent cross-session memory; agents explicitly decide what to store/retrieve.
- **Implication for Navigator**: confirms the cross-session persistence gap. The key difference: MemGPT/Mem0 are agent-mediated (the agent decides what to remember); the Navigator is project-mediated (it rolls up project state regardless of agent memory). The two operate at different layers — Navigator is the project-level substrate, agent-memory is the session-level learner.

### Claude compaction

- **Source**: https://platform.claude.com/docs/en/build-with-claude/compaction
- **Claim**: incremental summarization preserves CONVERSATION, not PROJECT-LEVEL intent.
- **Implication for Navigator**: motivates a separate navigation surface. Compaction keeps a single session's thread alive; it does not roll up the project's feature inventory. A returning session post-`/clear` has no conversation to compact into existence; it needs an external artifact — the Navigator.

## §2. Prior-work review (this repo)

### SPEC-LSEL-LOCAL-EVOLUTION-001 — boundary analysis

- **Status**: completed (per frontmatter read 2026-08-05).
- **Owns**: closing the PROPOSE→APPLY seam in user-owned surfaces (lessons-inbox drain → candidate `feedback_*.md` topics → `hns-lsel-*` harness edit proposals). Inputs: `.moai/lessons-inbox.jsonl` + usage-log. Outputs: `memory/feedback_*.md` + `hns-lsel-*` proposals + `CLAUDE.local.md §28`.
- **Boundary vs Navigator** (full table in plan.md §E): Navigator is project-scoped + present-tense (SPEC registry → capability-map); LSEL is harness-scoped + past→future (lessons → proposals). They share NO input and NO output surface. REQ-PN-016 codifies the non-overlap mechanically.
- **Touchpoint (future, not in scope)**: Navigator's `--audit` (SPEC-002) could surface "missing SPEC" findings that LSEL might consume as proposal seeds. This is a one-way future feed, not a bidirectional coupling.

### SPEC-WORKFLOW-LIFECYCLE-001 — sync-phase chain

- **Status**: exists in `.moai/specs/`; reviewed for sync-phase integration.
- **Owns**: the SPEC lifecycle (plan/run/sync) and its frontmatter transitions.
- **Boundary vs Navigator**: Navigator's regeneration hooks INTO the sync phase that this SPEC codifies — manager-docs invokes the Navigator regeneration step before the sync commit lands (REQ-PN-003). Navigator does NOT modify the lifecycle; it consumes SPEC statuses as inputs.

### SPEC-WF-AUDIT-GATE-001 — plan-audit gate

- **Status**: exists in `.moai/specs/`; reviewed for audit-gate integration.
- **Owns**: the plan-audit gate semantics (skip-eligible 0.90 autonomous bypass policy etc.).
- **Boundary vs Navigator**: this SPEC's plan-phase passes through the audit gate owned by SPEC-WF-AUDIT-GATE-001. Navigator is downstream of the audit gate, not a modifier of it.

### `/moai project` skill — extension target

- **Path**: `.claude/skills/moai-workflow-project/SKILL.md`
- **Current scope**: documentation scaffolding (product/structure/tech.md), language initialization, template optimization, docs generation, JIT document loading.
- **Gap**: documentation is generated ONCE at init; no living rollup.
- **Navigator extension**: adds a regeneration phase (M2) + a `--brief` mode (M3). The skill body is the single point of extension; no new CLI subcommand is introduced.

### `/moai codemaps` workflow — sibling, NOT extended by 001

- **Path**: `.claude/skills/moai/workflows/codemaps.md`
- **Current scope**: scan codebase → architecture documentation in `.moai/project/codemaps/`. Agentless fixed-pipeline (localize → repair → validate).
- **Boundary vs Navigator**: codemaps is the CODE structural map (modules, dependencies, entry points); Navigator is the FEATURE/status map (which SPECs exist, what phase, what frontier). SIBLING surfaces, not overlapping. SPEC-003 (tree-sitter) will BRIDGE them by auto-deriving capability rows from the codemaps output; 001 does not touch codemaps.

### SessionStart hook layer — extension target

- **Path**: `.claude/hooks/moai/handle-session-start.sh` (existing) + Go handler `internal/hook/session_start.go`
- **Existing pattern**: emits `hookSpecificOutput.additionalContext` (used for the moai-binary SessionStart probe).
- **Navigator extension**: adds a sibling hook `handle-session-start-navigator.sh` that emits a bounded Navigator brief into additionalContext. Fail-open + time-boxed (aligns with `.claude/rules/moai/development/coding-standards.md` § Advisory-Check Discipline). Does NOT modify the existing hook.

## §3. Summary — what the research changes in this SPEC

| Research finding | Where it shaped the SPEC |
|------------------|--------------------------|
| Map reduces context, not inflates it | REQ-PN-013 non-duplication; `navigator.md` answers task-shaped questions (frontier + next task) |
| Every claim needs provenance | REQ-PN-002 commit-sha + captured-at on every row |
| Persistence across sessions closes the gap | REQ-PN-003 sync-phase regeneration (staleness ≤ 1 sync cycle) |
| Compaction preserves conversation, not project intent | motivates Navigator as a separate project-scoped surface |
| AGENTS.md = single predictable entry point | `navigator.md` as the single project-state entry point |
| tree-sitter + PageRank (Aider, Codebase-Memory) | DEFERRED to SPEC-003 (auto-derivation); 001 uses SPEC registry + git log only |
