# MoAI-ADK v3 — Pattern Library

> Wave 2 synthesis artifact 3 of 3
> Source: Wave 1 research (R1 papers, R2 opensource, R3 CC rearch, R4 skills, R5 agents, R6 cmds+hooks+rules+config)
> Date: 2026-04-23
> Scope: 37 structural patterns in 7 categories, with per-pattern disposition (ADOPT / CONSIDER / NOT-NOW), trade-offs, and a Top-10 priority list for v3.

---

## Purpose

This library is the vocabulary Wave 3 (architecture design) and Wave 4 (SPEC regeneration) will use when proposing concrete v3 subsystems. Every pattern below is *named*, *sourced* to a specific Wave 1 finding, and *scored* on its fit with moai's existing FROZEN strengths (SPEC system, TRUST 5, 16-language neutrality, @MX TAG protocol).

No pattern is invented here. If a pattern does not appear in Wave 1 evidence, it is out of scope for v3 discussion in this document. Wave 3 is welcome to *compose* patterns (e.g., "ReAct inside Ralph outer loop") but not to *invent* new primitives without fresh Wave 1 evidence.

### Disposition legend

- **ADOPT** — Strong Wave 1 evidence, clear fit with moai FROZEN strengths, v3 should incorporate.
- **CONSIDER** — Evidence is strong but fit is conditional; Wave 3 must design the composition carefully.
- **NOT-NOW** — Valid pattern, but cost exceeds v3 benefit or conflicts with a FROZEN strength; revisit in v4+.

### Per-pattern entry schema

- **ID & name** (e.g., R-1 ReAct Reason-Act-Observe Loop)
- **Source** (Wave 1 doc + section)
- **Description** (2-4 sentences)
- **When to apply** (moai-specific context)
- **Trade-offs** (cost, complexity, conflict with other patterns)
- **v3 disposition** + rationale

---

## Category 1 — Reasoning

### R-1 ReAct Reason-Act-Observe Loop

- **Source**: R1 §1 ReAct (Yao et al. 2022, *arXiv:2210.03629*).
- **Description**: Interleave Thought, Action, and Observation in a single agent turn. Reasoning guides action; observations ground reasoning.
- **When to apply**: Inner loop of every agent turn in `/moai run`, `/moai plan`, `/moai sync`. Baseline primitive.
- **Trade-offs**: Pure ReAct degenerates at >5 iterations due to token accumulation (R1 §1 anti-pattern flag). Needs outer-loop reset (see R-6).
- **v3 disposition**: **ADOPT** — already implicit in moai; v3 makes the Thought step explicit via `@MX:NOTE` tags preserving reasoning traces in source.

### R-2 Self-Refine Iterative FEEDBACK→REFINE

- **Source**: R1 §3 Self-Refine (Madaan et al. 2023, *arXiv:2303.17651*).
- **Description**: Single model acts as generator, critic, and refiner in a bounded loop. ~20% improvement over single-shot even on GPT-4 class models.
- **When to apply**: `/moai fix`, `/moai loop`, and the RED→GREEN cycle within manager-cycle.
- **Trade-offs**: Requires explicit stopping criterion (stagnation detector). Without it, infinite loops.
- **v3 disposition**: **ADOPT** — moai-workflow-loop already implements the shape; v3 adds stagnation detection per GAN loop §11 and wires it to `/moai loop`.

### R-3 Reflexion Actor/Evaluator/Self-Reflection Separation

- **Source**: R1 §2 Reflexion (Shinn et al. 2023, *arXiv:2303.11366*).
- **Description**: Split agent into three modules with an episodic memory buffer. Self-Reflection converts environment signals into verbal gradients. 8% absolute improvement from reflection over memory-only baselines.
- **When to apply**: GAN loop (thorough harness) — expert-frontend as Actor, evaluator-active as Evaluator, explicit Reflector module converts failing scores into text feedback for next iteration.
- **Trade-offs**: Three modules = three prompts per iteration. Memory-only without reflection is strictly worse than both retained.
- **v3 disposition**: **ADOPT** — moai already has Actor and Evaluator; missing piece is an explicit Reflector module that produces structured text gradients before Actor's next turn.

### R-4 Tree-of-Thoughts Branching + Self-Evaluation

- **Source**: R1 §4 Tree of Thoughts (Yao et al. 2023, *arXiv:2305.10601*).
- **Description**: Extend Chain-of-Thought to a tree with BFS/DFS over "thought" units; self-evaluation drives pruning. GPT-4 on Game-of-24: 4% (CoT) → 74% (ToT).
- **When to apply**: `/moai plan` at thorough harness when multiple architectural approaches are plausible. Branch at strategy phase; plan-auditor scores each branch.
- **Trade-offs**: 5-100× token cost vs CoT. Only justified for genuinely open-ended tasks.
- **v3 disposition**: **CONSIDER** — introduce only at `harness: thorough`; gate behind user opt-in at plan time. Do not default for `/moai run`.

### R-5 CodeAct Executable Code as Action Space

- **Source**: R1 §12 CodeAct (Wang et al. 2024, *arXiv:2402.01030*).
- **Description**: Replace JSON tool calls with executable Python/Bash code. +20% success rate over Text/JSON actions. Compose multiple tools in one LLM turn; leverage error messages as observations.
- **When to apply**: Complex multi-step tool compositions (e.g., "find all files matching X, run Y on each, collect results"). Today moai splits into N Bash calls.
- **Trade-offs**: Sandboxing is non-trivial (R1 §12 anti-pattern flag). Needs process isolation (see S-3).
- **v3 disposition**: **CONSIDER** — valuable for ACI efficiency, but only if S-3 (sandboxed execution) ships simultaneously. Do not expose raw Python eval to unsandboxed agents.

### R-6 Ralph Outer-Loop with Fresh Context

- **Source**: R2 §1-2 (iannuttall/ralph, snarktank/ralph), R2 §3 OMC `/ralph` mode.
- **Description**: Outer loop with fresh LM context per iteration; state persists only on disk (git + `.moai/state/`). Beats elaborate context compression by being bounded by disk, not by context window.
- **When to apply**: Multi-iteration workflows — `/moai loop`, `/moai fix --iterate`, GAN loop when `/moai run` exceeds 5 ReAct iterations.
- **Trade-offs**: Fresh context forgets within-task reasoning — mitigated by persistence in SPEC, `@MX` tags, and `.moai/state/`.
- **v3 disposition**: **ADOPT** — elevate `/moai loop` to first-class execution mode per R2 §A top-5 #2 and design-principles P7. Adopt Ralph's `STALE_SECONDS` primitive as crash-resume semantics.

---

## Category 2 — Orchestration

### O-1 LLM-Compiler Parallel Task DAG

- **Source**: R1 §15 LLMCompiler (Kim et al. 2023, *arXiv:2312.04511*, ICML 2024).
- **Description**: Treat tool orchestration as classical compilation. Planner emits a task DAG; Task Fetching Unit schedules; Executor runs independents in parallel. 3.7× latency, 6.7× cost reduction vs ReAct.
- **When to apply**: `/moai plan` output becomes a DAG (not a linear task list); team mode file-ownership declarations become DAG edges.
- **Trade-offs**: Naive parallel writes cause conflicts. Requires worktree isolation for write agents (S-1 and existing moai rule).
- **v3 disposition**: **ADOPT** — partially in place (team mode + worktree isolation); v3 formalizes plan-phase DAG emission and per-agent `reads: [paths]` / `writes: [paths]` declarations.

### O-2 Publish-Subscribe Shared Message Pool

- **Source**: R1 §7 MetaGPT (Hong et al. 2024), R2 §13 MetaGPT.
- **Description**: Agents communicate via shared message pool rather than direct agent-to-agent calls. Avoids N² coupling; scales better past 3-4 agents.
- **When to apply**: Team mode with ≥5 teammates; cross-agent context propagation (plan → strategy → backend → frontend → quality).
- **Trade-offs**: Message pool can become noisy; needs filter semantics per-agent.
- **v3 disposition**: **CONSIDER** — moai team mode currently uses direct SendMessage; adding a shared pool is valuable but not critical for current team sizes. Evaluate after team mode stabilizes past 10 teammates.

### O-3 Magentic Dynamic Task Ledger

- **Source**: R2 §12 Microsoft Agent Framework 1.0 (GA April 2026), Magentic pattern.
- **Description**: Manager agent builds a dynamic task ledger, coordinates specialists + humans. Differs from static-roster multi-agent by spawning teammates on-demand.
- **When to apply**: manager-strategy extending its current role to dynamic teammate spawning via workflow.yaml role_profiles + general-purpose Agent().
- **Trade-offs**: Dynamic spawn = less predictable cost envelope; requires task-ledger observability.
- **v3 disposition**: **ADOPT** — moai already adopted dynamic team generation (workflow.yaml role_profiles + general-purpose spawn). v3 formalizes the "task ledger" as `.moai/state/task-ledger.md` with append-only semantics per R2 ralph file-first pattern.

### O-4 Multi-Mode Router (team/autopilot/ultrawork/ralph/ralplan/pipeline)

- **Source**: R2 §3 oh-my-claudecode (OMC) 6 operating modes.
- **Description**: Explicit top-level mode surface beyond subcommands: `/team` (staged pipeline), `/autopilot` (single lead), `/ultrawork` (burst parallelism), `/ralph` (loop), `/ralplan` (iterative planning consensus), `omc team` (tmux workers).
- **When to apply**: Clarify moai's /moai surface — currently 15 subcommands without an explicit "execution style" axis.
- **Trade-offs**: Adds a second axis (subcommand × mode) to user mental model. Needs clear defaults.
- **v3 disposition**: **CONSIDER** — don't adopt all 6 OMC modes; instead add 2-3 execution styles as explicit `--mode` flags on `/moai run` (e.g., `--mode loop` = Ralph, `--mode team` = team orchestration, `--mode autopilot` = single-lead). Default mode auto-selected by harness.

### O-5 Plan/Act Mode Separation

- **Source**: R2 §5 Cline, §18 opencode (Tab-key mode switch).
- **Description**: Two modes with distinct permission envelopes: `plan` (read-only; writes denied) and `act` (writes permitted). Approval gate between them.
- **When to apply**: Every agent can declare `permissionMode: plan | acceptEdits | bypassPermissions | bubble` (see S-1, S-2).
- **Trade-offs**: Mode transitions add user friction. Mitigated by auto-transition on explicit user approval.
- **v3 disposition**: **ADOPT** — moai already has permissionMode on agents; v3 formalizes bubble mode for fork-agent escalation (see S-2) and documents the plan→act approval gate.

### O-6 Agentless Fixed Pipeline for Well-Structured Tasks

- **Source**: R1 §25 Agentless (Xia et al. 2024, *arXiv:2407.01489*).
- **Description**: Three-phase non-agentic pipeline (localize → repair → validate) without LLM-driven control flow. 27.3% SWE-bench Lite, outperforming open-source agentic competitors at lower cost.
- **When to apply**: Utility subcommands where task structure is known: `/moai fix`, `/moai coverage`, `/moai codemaps`, `/moai mx`, `/moai clean`. Currently some over-use multi-agent.
- **Trade-offs**: Wrong pattern for genuinely open-ended tasks (plan, design).
- **v3 disposition**: **ADOPT** — classify utility subcommands as fixed-pipeline; keep multi-agent for plan/run/design/sync. Per R1 §Y divergence matrix.

---

## Category 3 — Memory

### M-1 Typed Memory Taxonomy (user/feedback/project/reference)

- **Source**: R3 §1.1 memdir, §4 Adoption Candidate 7.
- **Description**: Four typed memory kinds with distinct lifetimes and prompts — `user` (about the developer), `feedback` (correction patterns), `project` (state of current work), `reference` (external system pointers).
- **When to apply**: `.claude/agent-memory/{manager-*|expert-*|builder-*}/` directories subdivide by type; MEMORY.md is the index.
- **Trade-offs**: More types = more discovery cost; mitigated by LLM-selected retrieval.
- **v3 disposition**: **ADOPT** — moai's existing MEMORY.md already uses this schema informally. Formalize as rule in `.claude/rules/moai/workflow/moai-memory.md`. Low-cost, high-value normalization.

### M-2 Three-Layer Memory (observation / reflection / plan)

- **Source**: R1 §8 Generative Agents (Park et al. 2023, *arXiv:2304.03442*).
- **Description**: Observation stream + periodic Reflection consolidation + forward Plan. Reflection compresses low-level observations into higher-level beliefs via a "sleep cycle."
- **When to apply**: lessons.md auto-capture (SPEC-SLQG-001) gains a periodic consolidation cron that promotes repeatedly-observed patterns from observations → heuristics → rules.
- **Trade-offs**: Consolidation is an LLM call — costs tokens. Schedule per-project weekly or per-10-SPEC completions.
- **v3 disposition**: **ADOPT** — aligns with existing graduation protocol (design-constitution §7). Add the consolidation cron as a `/moai project memory consolidate` subcommand.

### M-3 Skill Library with Embeddings-Indexed Retrieval

- **Source**: R1 §13 Voyager (Wang et al. 2023, *arXiv:2305.16291*).
- **Description**: Skills as executable code, accumulated across sessions, retrieved by embedding similarity on task description.
- **When to apply**: `.claude/skills/` already Voyager-shaped; Voyager suggests skills should auto-distill from successful `/moai run` trajectories, not only be human-authored.
- **Trade-offs**: Skill library grows unboundedly without GC (R1 §13 anti-pattern). Needs versioning + archival.
- **v3 disposition**: **CONSIDER** — auto-distillation is promising but premature; first ship versioning + archival for existing 48 skills (R4 recommendation), then evaluate auto-distillation in v4+.

### M-4 Workflow Memory Induction from Trajectories (AWM)

- **Source**: R1 §24 Agent Workflow Memory (Wang et al. 2024, *arXiv:2409.07429*).
- **Description**: Induce reusable workflows from past trajectories; store as first-class memory objects distinct from raw episodic memory. +51.1% on WebArena; outperforms human-expert workflows by 7.9%.
- **When to apply**: Procedural memory layer for moai — successful `/moai run SPEC-XXX` trajectories become abstracted workflow templates for future SPECs.
- **Trade-offs**: Naive trajectory-saving floods memory; abstraction step (removing example-specific context) is essential.
- **v3 disposition**: **CONSIDER** — gate behind graduation protocol; pilot with 1-2 workflow categories (e.g., "add REST endpoint" or "add test for existing function") before broad rollout.

### M-5 LLM-Selected Relevance with Staleness Caveat

- **Source**: R3 §2 Decision 13 & 14.
- **Description**: Memory retrieval uses an LLM sideQuery (not regex/grep) to select relevant memories. Stale memories (>1 day) wrapped in `<system-reminder>` with explicit caveat to prevent over-trust.
- **When to apply**: Every session start, moai runs a relevance query against `.claude/agent-memory/` to inject top-5 memories into context with freshness annotation.
- **Trade-offs**: Each retrieval costs a sideQuery LLM call. Mitigated by caching per-SPEC relevance query results.
- **v3 disposition**: **ADOPT** — moai lessons.md currently has no staleness awareness; adding this aligns moai with CC's proven pattern at low cost.

---

## Category 4 — Tool-use Interface

### T-1 Agent-Computer Interface (ACI) with LM-Centric Commands

- **Source**: R1 §11 SWE-agent (*arXiv:2405.15793*, NeurIPS 2024), R2 §8 SWE-agent deep analysis, R2 §18 opencode LSP-auto-load.
- **Description**: Curated command set (open, scroll, edit, search, linter-gated-write) with structured, paginated, LM-optimized responses. Empirically 8× pass@1 vs raw shell.
- **When to apply**: Common moai operations — SPEC read, MX locate, test run, LSP symbol lookup — become named commands. Raw `Bash` remains the escape hatch.
- **Trade-offs**: Wrapping maintenance; one more thing to keep current with underlying tools.
- **v3 disposition**: **ADOPT (priority 1)** — strongest single leverage pattern per R1/R2 consensus. Start with 6 commands: `moai_spec_read`, `moai_locate_mx_anchor`, `moai_run_tests_for_spec`, `moai_lsp_find_references`, `moai_lsp_workspace_symbols`, `moai_linter_gated_write`.

### T-2 Tree-sitter Repo-Map as Context Primer

- **Source**: R2 §6 Aider ("Best git integration among terminal coding agents").
- **Description**: Tree-sitter-derived repository map across 100+ languages. Primes LM context with structural skeleton; active files tracked via `/add` and `/read-only`.
- **When to apply**: Agent spawn time, inject a compact Tree-sitter skeleton of relevant code regions for the target SPEC.
- **Trade-offs**: Tree-sitter setup per language; existing codemaps already provide some coverage.
- **v3 disposition**: **CONSIDER** — moai codemaps (`/moai codemaps`) already partially covers this; v3 should deepen Tree-sitter integration specifically for plan-phase repo understanding, not reinvent wholesale.

### T-3 LSP Auto-Load per Language

- **Source**: R2 §18 opencode (LSP auto-load per language) + existing moai SPEC-LSP-CORE-002 / powernap integration.
- **Description**: Language-server runs automatically for the current project language; LSP-derived structural queries are first-class tools (find references, workspace symbols, rename).
- **When to apply**: Every `/moai run` session; LSP-backed ACI commands (see T-1).
- **Trade-offs**: LSP startup cost; 16-language neutrality requires per-language adapters (powernap handles this).
- **v3 disposition**: **ADOPT** — already in place via powernap. v3 elevates LSP commands to the ACI tier (see T-1) and standardizes output transformation for LM consumption.

### T-4 4 Hook Types with Cost Ladder (command/prompt/http/agent)

- **Source**: R3 §2 Decision 6.
- **Description**: Hooks can be `command` (cheap subprocess), `prompt` (cheap LLM yes/no), `agent` (expensive LLM-with-tools), or `http` (SSRF-guarded webhook). Cost-appropriate per problem.
- **When to apply**: Current moai hooks are all `command` type. Upgrade specific hooks where richer reasoning justifies cost (e.g., PostToolUse MX tag validation could be `prompt` type).
- **Trade-offs**: 4-way type choice adds config complexity.
- **v3 disposition**: **CONSIDER** — document the cost ladder; introduce `prompt`-type hooks for MX validation and pre-toollike negotiation. Do not rush all hooks to higher tiers.

### T-5 Hook JSON-OR-ExitCode Dual Protocol

- **Source**: R3 §2 Decision 5, R3 §4 Adoption Candidate 4.
- **Description**: Hooks can reply with `{additionalContext, permissionDecision, updatedInput, systemMessage, continue}` JSON *or* exit code. Backward compatible; unlocks programmable mid-turn rewrites.
- **When to apply**: Every moai hook handler gains a JSON output option. Sprint Contract injection, MX tag injection, config-reload triggers all become programmable.
- **Trade-offs**: Handler authors must learn the JSON schema.
- **v3 disposition**: **ADOPT (priority 2)** — per R3 §4 and design-principles P10. Migrate the 5 critical handlers first (subagent-stop, config-change, setup, instructions-loaded, file-changed per R6 recommendations).

---

## Category 5 — Safety / Permission

### S-1 Multi-Source Permission Resolution

- **Source**: R3 §1.3 hooks settings precedence, §4 Adoption Candidate 2.
- **Description**: Permissions resolve via ordered stack — `policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision` → `allow | ask | deny + updatedInput?`.
- **When to apply**: Every tool invocation passes through the stack. Enables pre-allowlist for common ops, bubble-to-user for risky ops, programmatic rule rewriting via hook responses.
- **Trade-offs**: Stack resolution adds per-call overhead (microseconds). Debugging permission denials needs provenance (see X-2).
- **v3 disposition**: **ADOPT (priority 3)** — moai lacks this entirely. Foundational for S-2, S-3, and design-principles P9.

### S-2 Permission Bubble/Escalation Mode

- **Source**: R3 §2 Decision 15.
- **Description**: `bubble` is a first-class permission mode (not a boolean). Fork agents inheriting parent context use bubble to push permission decisions to the parent terminal, not their own mailbox.
- **When to apply**: Fork agents spawned by implementer agents (e.g., manager-ddd spawning expert-security for a sensitive check).
- **Trade-offs**: Bubble adds user prompts — approval fatigue risk.
- **v3 disposition**: **ADOPT** — pair with S-1 as a single v3 SPEC. Bubble mode only for genuinely novel decisions; pre-allowlist and session rules handle the 80% case.

### S-3 Ephemeral Sandboxed Execution

- **Source**: R2 §A Top-5 Pattern 5; OWASP Top 10 for Agentic Apps (Dec 2025); R2 §B Anti-pattern 1.
- **Description**: Implementer agents execute in ephemeral isolated sandbox (Bubblewrap/Seatbelt/Docker/E2B/Modal). Network egress denylist; file-write scope enforcement. Approval-only safety is empirically exploitable.
- **When to apply**: Every implementation agent that writes files. Opt-in via agent frontmatter `sandbox: bubblewrap|seatbelt|docker|none`.
- **Trade-offs**: Sandbox startup tax; cross-platform support needs per-OS backends.
- **v3 disposition**: **ADOPT** — the only surveyed tool (other than snarktank/ralph's Docker-by-default) that ships sandboxing. moai can lead by correcting the ecosystem gap. Align with security.yaml already-loaded config (R6 §5.1).

### S-4 FROZEN Zone + Graduation Protocol

- **Source**: R1 §18 Constitutional AI (*arXiv:2212.08073*), design-constitution.md §2 (FROZEN/EVOLVABLE), §5-7 (safety + graduation).
- **Description**: Declarative constitution with explicit FROZEN zone (immutable by automation) and EVOLVABLE zone (amendable by graduation protocol). Observations progress through tiers: 1× observation → 3× heuristic → 5× rule → 10× auto-propose-evolution.
- **When to apply**: Extend from design subsystem (already FROZEN in v3.3.0) to core constitution. FROZEN: AskUserQuestion monopoly, SPEC format, TRUST 5 baseline, 16-language neutrality. EVOLVABLE: skill bodies, effort levels, harness thresholds within bounds.
- **Trade-offs**: Extra process layer for rule changes; mitigated by EVOLVABLE fast lane.
- **v3 disposition**: **ADOPT** — core pattern is already partially implemented in design-constitution; v3 generalizes to core constitution.

### S-5 5-Layer Safety Architecture (FrozenGuard / Canary / Contradiction / RateLimit / Human)

- **Source**: design-constitution.md §5 (existing FROZEN layer).
- **Description**: Five concentric safety layers gating all evolution — FrozenGuard blocks FROZEN-zone writes; Canary shadow-evaluates on last 3 projects; Contradiction Detector flags rule conflicts; Rate Limiter bounds evolution velocity (3/week, 24h cooldown, 50 active learnings); Human Oversight requires explicit approval.
- **When to apply**: Any automated constitution amendment, skill evolution, or learner-proposed rule change.
- **Trade-offs**: High process cost per change; mitigated by EVOLVABLE zone scoping.
- **v3 disposition**: **ADOPT** — already FROZEN in design subsystem; extend to core constitution evolution in v3.

### S-6 3-Failure Circuit Breaker

- **Source**: R3 §2 Decision 4 (autocompact 3-failure breaker), empirical BigQuery data (1,279 sessions × 250K wasted API calls/day).
- **Description**: Bound consecutive failures to 3 before hard-stop. Applied to autocompact in CC; the general principle: failure modes must be capped, not trusted.
- **When to apply**: Selectively — GAN loop escalation (already has max_iterations); `/moai loop` stagnation detection; external API retries in moai binary.
- **Trade-offs**: Wrong for local deterministic operations that should fail fast instead.
- **v3 disposition**: **CONSIDER** — apply only where operations call paid APIs or loop over LLM turns. Not a universal pattern.

---

## Category 6 — Evaluation

### E-1 Agent-as-a-Judge Intermediate Trajectory Scoring

- **Source**: R1 §9 Agent-as-a-Judge (Zhuge et al. 2024, *arXiv:2410.10934*).
- **Description**: Extend LLM-as-a-Judge with agentic capabilities (tool use, memory, multi-step reasoning). Scores intermediate trajectories, not only final outputs. Matches human evaluation reliability at 97% cost savings.
- **When to apply**: evaluator-active at thorough harness — scores trajectory, not just final artifact. Hierarchical requirements (Zhuge's 365 sub-requirements per 55 tasks) map to moai's acceptance criteria per SPEC.
- **Trade-offs**: Judge memory across iterations cascades errors (R1 §9 anti-pattern flag) — must scope memory per-iteration.
- **v3 disposition**: **ADOPT (priority 9)** — evaluator-active already in this shape; v3 formalizes hierarchical acceptance criteria and enforces per-iteration memory scope.

### E-2 Sprint Contract Negotiation

- **Source**: design-constitution.md §11.4 (existing moai pattern), aligns with R1 §9 Agent-as-a-Judge.
- **Description**: Before each GAN loop iteration, evaluator-active and Actor negotiate a Sprint Contract: acceptance checklist, priority dimension, test scenarios, pass conditions. Passed criteria carry forward; failed criteria get refined. Evaluator cannot score on non-contracted criteria.
- **When to apply**: GAN loop at thorough harness (already). v3 extends the pattern to plan phase ("Plan Contract" — what defines a valid plan) and sync phase ("Release Contract" — what defines a ready-to-merge artifact).
- **Trade-offs**: Contract negotiation adds an LLM call per iteration.
- **v3 disposition**: **ADOPT (priority 6)** — existing moai strength; generalize from run-phase to plan-phase and sync-phase per R1 Open Research Question 8.

### E-3 Rubric-Anchored Scoring with Independent Re-evaluation

- **Source**: design-constitution.md §12 (Mechanism 1 & 4), aligns with R1 Reflexion + Constitutional AI critique-then-revise.
- **Description**: Every evaluation criterion has concrete rubric with examples at 0.25, 0.50, 0.75, 1.0. Every 5th project undergoes independent re-evaluation: scores must be within 0.10 of each other; divergence triggers calibration review.
- **When to apply**: evaluator-active and plan-auditor both reference rubrics in `.moai/config/evaluator-profiles/`. Independent re-evaluation automatic every 5th SPEC.
- **Trade-offs**: Rubric authoring is expensive one-time cost.
- **v3 disposition**: **ADOPT** — already FROZEN in design subsystem; v3 populates evaluator-profiles per harness level (default, strict, lenient, frontend).

### E-4 Pass@N Multi-Attempt Evaluation

- **Source**: R1 §21 MLE-bench (Chan et al. 2024, *arXiv:2410.07095*).
- **Description**: Multi-attempt evaluation reveals capacity that pass@1 hides. o1-preview + AIDE: 16.9% bronze (pass@1), 34.1% at pass@8.
- **When to apply**: `/moai run --attempts N` for SPECs where the first attempt is likely suboptimal (e.g., novel architecture). Best-of-N evaluator selects the winning attempt.
- **Trade-offs**: N× cost per SPEC; only justified when first-attempt success rate is measurably low.
- **v3 disposition**: **CONSIDER** — introduce as `--attempts` flag gated by `harness: thorough` + user opt-in. Do not default.

---

## Category 7 — Extension

### X-1 Markdown-Authored Agents/Skills with YAML Frontmatter

- **Source**: R2 §A Top-5 Pattern 3 (universal: Continue `.continue/checks/*.md`, Claude Code skills, crewAI `agents.yaml`, moai existing).
- **Description**: One file = one agent/skill/check. Markdown body, YAML frontmatter for typed config. Cleanest discoverability pattern across the surveyed ecosystem.
- **When to apply**: Already moai's native pattern. v3 should formalize a single shared schema across agents, skills, commands, and rules.
- **Trade-offs**: Schema drift between four artifact types.
- **v3 disposition**: **ADOPT** — build a shared JSON schema + validator (Go-side), fail CI on schema violation. Foundation for S-1 (provenance) and the agent CI lint from R5.

### X-2 Multi-Layer Settings with Provenance Tags

- **Source**: R3 §2 Decision 11, §4 Adoption Candidate 1.
- **Description**: Every setting value carries a `source` tag (policy / user / project / local / plugin / skill / session / builtin). Merge is deterministic; `/moai doctor` shows "this rule came from which file."
- **When to apply**: All 5 missing yaml loaders (see Problem Catalog P-H06) gain `source` tagging at merge time. Hook registrations, MCP servers, permission rules, skill activations all tagged.
- **Trade-offs**: Metadata bloat in merged representation; mitigated by tagging at merge, not in source YAML.
- **v3 disposition**: **ADOPT (priority 4)** — prerequisite for S-1, S-2, S-3 and auditability. Low per-file cost, high leverage.

### X-3 Output Style as Override Contract

- **Source**: R3 §4 Adoption Candidate 5; `.claude/output-styles/*.md` with frontmatter `{name, description, keep-coding-instructions, force-for-plugin}`.
- **Description**: Versioned, layered, hot-reloadable system prompt modification mechanism. Project-level overrides user-level; plugins can declare `forceForPlugin: true`.
- **When to apply**: moai already ships 2 output styles (MoAI, Einstein). v3 formalizes the frontmatter schema to match CC's exactly, ensuring serialization identity.
- **Trade-offs**: Small — output styles are already working.
- **v3 disposition**: **ADOPT** — cheap alignment with CC canonical format. Schema audit + documentation.

### X-4 Three-Origin Plugin System (builtin / installed / inline)

- **Source**: R3 §2 Decision 12.
- **Description**: Plugins come from three origins with different trust/lifetime — `builtin` (bundled, always trusted), `installed` (marketplace, persistent), `inline` (`--plugin-dir`, session-only, dev loop).
- **When to apply**: moai plugin distribution currently via git clone / template copy; no native plugin abstraction beyond CC's own.
- **Trade-offs**: Adding a plugin system is substantial new surface area.
- **v3 disposition**: **NOT-NOW** — moai's extensibility via skills + agents is already sufficient; a second plugin layer adds confusion. Revisit if v4 adds a marketplace.

### X-5 Versioned Migration Auto-Apply at preAction

- **Source**: R3 §2 Decision 10 (CC `CURRENT_MIGRATION_VERSION = 11`).
- **Description**: Migrations run silently in a commander preAction hook; each migration is idempotent with a version guard. Users never need to know migrations exist.
- **When to apply**: `moai init`, `moai hook session-start`, `moai update` — silently apply pending migrations.
- **Trade-offs**: Silent changes can surprise users; mitigated by logging migration results.
- **v3 disposition**: **ADOPT** — moai's current `moai migrate agency` is an explicit command users must remember; R6 Problem P-C06 flags this. Silent preAction migrations align with CC and reduce support burden.

---

## Top-10 Priority Patterns for v3

Selection rubric: **Impact** (Wave 1 evidence strength × moai gap size) × **Feasibility** (implementation cost, Claude Code compatibility) × **Alignment** (reinforces FROZEN strengths — SPEC, TRUST 5, 16-lang neutrality, @MX tags).

Ranked 1-10:

### 1. T-1 Agent-Computer Interface (ACI)

Biggest single leverage per R1/R2 consensus. Moai today exposes raw `Bash` to agents; SWE-agent empirically 8× better with purpose-built ACI. Starts paying off immediately once 6 core commands ship. Aligns with existing @MX tag protocol and LSP integration.

### 2. T-5 Hook JSON-OR-ExitCode Dual Protocol

Unlocks programmable Sprint Contract injection, MX tag additions, mid-turn permission rewriting. Backward compatible (exit codes still work). Migration path: 5 critical handlers first (per R6), then broader. Foundation for every future hook enhancement.

### 3. S-1 Multi-Source Permission Resolution

Moai lacks this entirely. Foundation for bubble mode (S-2), sandbox routing (S-3), and auditable configuration. CC validates at scale. Without it, moai's security posture matches the 2026 anti-pattern (approval-fatigue-only).

### 4. X-2 Multi-Layer Settings with Provenance Tags

Prerequisite for S-1 and for /moai doctor observability. Addresses Problem P-H06 (5 yaml sections with no Go loader) and P-C04 (two-tier config with no provenance). Low per-file cost, high downstream leverage.

### 5. M-1 Typed Memory Taxonomy

Formalizes the schema MEMORY.md already uses informally. Aligns moai with CC's proven typed memory approach. Low-risk, high-clarity win. Supports M-5 (staleness caveats) and lessons.md periodic consolidation.

### 6. E-2 Sprint Contract Negotiation (generalized)

Existing moai strength (design-constitution §11.4) generalized from GAN loop to plan phase (Plan Contract) and sync phase (Release Contract). Answers R1 Open Research Question 8 about what the SPEC equivalent is for non-run phases.

### 7. R-6 Ralph Outer-Loop Fresh Context

Already on roadmap (moai-workflow-loop, `/moai loop` command). Formalize `.moai/state/` as the authoritative file-first state substrate. Adopt `STALE_SECONDS` primitive. Elevate `/moai loop` to first-class execution mode alongside `/moai run`.

### 8. S-3 Ephemeral Sandboxed Execution

Critical for 2026 security posture (OWASP Top 10 Agentic Apps mandatory). moai differentiates in the ecosystem where every surveyed tool except snarktank/ralph has "none" for sandboxing. Opt-in per agent role (implementer = sandboxed; reviewer = read-only).

### 9. E-1 Agent-as-a-Judge Intermediate Trajectory Scoring

Elevates evaluator-active to the Agent-as-a-Judge canonical pattern. Hierarchical acceptance criteria (R1 §9 style) enhance SPEC acceptance.md format. Per-iteration memory scope closes the cascade-error gap identified in current design-constitution §11.4.

### 10. O-4 Multi-Mode Router (reduced to 2-3 modes)

Not all 6 OMC modes, but 2-3 explicit execution styles on `/moai run` — `--mode loop` (Ralph), `--mode team` (orchestrated), `--mode autopilot` (single-lead). Resolves the current subcommand-only surface; surfaces Ralph and team execution as first-class options. Auto-selected by harness when flag omitted.

---

## Pattern composition guidance for Wave 3

Wave 3 architecture design should *compose* these patterns, not cherry-pick. Three canonical compositions emerge from the Top-10:

**Composition A — Autonomous Run:**
```
O-4 (Multi-Mode Router: --mode autopilot)
  ├── R-6 (Ralph outer loop, fresh context per iteration)
  │   └── R-1 (ReAct inner loop)
  │       ├── T-1 (ACI commands, not raw Bash)
  │       ├── M-1 (Typed memory retrieval at turn start)
  │       └── T-5 (Hook JSON injection mid-turn)
  └── E-2 (Sprint Contract bounds the iterations)
```

**Composition B — Team Orchestration:**
```
O-4 (--mode team)
  ├── O-1 (LLM-Compiler parallel DAG from plan phase)
  ├── O-3 (Magentic dynamic task ledger)
  ├── S-1 (Permission stack per teammate)
  │   └── S-2 (Bubble mode for fork agents)
  │   └── S-3 (Sandbox for implementer roles)
  └── E-1 (Agent-as-a-Judge evaluator-active on artifacts)
```

**Composition C — Fixed-Pipeline Utility:**
```
O-6 (Agentless fixed pipeline)
  ├── /moai fix | /moai coverage | /moai mx | /moai clean
  ├── R-2 (Self-Refine bounded loop)
  │   └── S-6 (3-failure circuit breaker on LLM retries)
  └── T-1 (Minimal ACI surface: locate, edit, verify)
```

These compositions are templates, not prescriptions. Wave 3 SPECs should declare which composition they extend or diverge from.

---

## Source mapping (Wave 1 citations per pattern)

| Pattern | Primary source | Secondary sources |
|---------|----------------|-------------------|
| R-1 ReAct | R1 §1 | Universal across all surveyed coding agents |
| R-2 Self-Refine | R1 §3 | R2 §1 ralph (bounded iteration) |
| R-3 Reflexion | R1 §2 | design-constitution §11 (GAN loop) |
| R-4 Tree-of-Thoughts | R1 §4 | R1 §5 LATS |
| R-5 CodeAct | R1 §12 | R1 §13 Voyager, R1 Chain of Code |
| R-6 Ralph Fresh-Context | R2 §1-2, §3 | R1 §25 Agentless |
| O-1 LLM-Compiler DAG | R1 §15 | CLAUDE.md §14 Parallel Safeguards |
| O-2 Pub-Sub Message Pool | R1 §7 MetaGPT, R2 §13 | — |
| O-3 Magentic Ledger | R2 §12 MS Agent Framework | Existing moai workflow.yaml role_profiles |
| O-4 Multi-Mode Router | R2 §3 OMC | — |
| O-5 Plan/Act Mode | R2 §5 Cline, §18 opencode | moai existing permissionMode |
| O-6 Agentless Pipeline | R1 §25 | R2 §1 ralph |
| M-1 Typed Memory Taxonomy | R3 §1.1, §4 Adopt 7 | moai MEMORY.md informal |
| M-2 3-Layer Memory | R1 §8 Generative Agents | moai lessons.md |
| M-3 Skill Library Embeddings | R1 §13 Voyager | R4 skill audit (needs versioning) |
| M-4 Workflow Memory Induction | R1 §24 AWM | design-constitution §7 graduation |
| M-5 LLM-Selected + Staleness | R3 §2 Dec 13-14 | R3 §1.1 memdir |
| T-1 ACI | R1 §11, R2 §8 | R2 §18 opencode LSP |
| T-2 Tree-sitter Repo-Map | R2 §6 Aider | moai codemaps |
| T-3 LSP Auto-Load | R2 §18, moai SPEC-LSP-CORE-002 | powernap dependency |
| T-4 4 Hook Types | R3 §2 Dec 6 | moai current command-only |
| T-5 Hook JSON-OR-ExitCode | R3 §2 Dec 5, §4 Adopt 4 | R6 recommendations |
| S-1 Multi-Source Permission | R3 §1.3, §4 Adopt 2 | design-principles P9 |
| S-2 Permission Bubble | R3 §2 Dec 15 | CC source: agent-team.md §4.3 |
| S-3 Sandboxed Execution | R2 §A Top-5 #5, OWASP 2025 | R1 §12 anti-pattern flag |
| S-4 FROZEN + Graduation | R1 §18, design-constitution §2, §5-7 | moai existing |
| S-5 5-Layer Safety | design-constitution §5 (FROZEN) | R1 §16 ADAS canary |
| S-6 3-Failure Circuit | R3 §2 Dec 4 | BigQuery data cited in R3 |
| E-1 Agent-as-a-Judge | R1 §9 | design-constitution §11.4 Sprint |
| E-2 Sprint Contract | design-constitution §11.4 | R1 §9 alignment |
| E-3 Rubric + Independent Re-eval | design-constitution §12 | R1 Reflexion + Constitutional AI |
| E-4 Pass@N | R1 §21 MLE-bench | — |
| X-1 Markdown+YAML Frontmatter | R2 §A Top-5 #3 (universal) | moai existing |
| X-2 Multi-Layer Settings Provenance | R3 §2 Dec 11, §4 Adopt 1 | R6 §5 unused sections |
| X-3 Output Style as Override | R3 §4 Adopt 5 | moai 2 existing styles |
| X-4 3-Origin Plugin | R3 §2 Dec 12 | — |
| X-5 Versioned Migration preAction | R3 §2 Dec 10 | moai current explicit `moai migrate` |

---

## Patterns deliberately NOT adopted (divergence from Wave 1)

For traceability, Wave 3 should know which Wave 1 patterns are explicitly rejected or deferred:

| Pattern / source | Why not now | Revisit trigger |
|------------------|-------------|-----------------|
| CC Bridge architecture (R3 §3.9, §5 div 3) | Cloud-tied, 33 files, OAuth dependency. moai uses local tmux for multi-machine. | Anthropic partnership or dual-hosted infra |
| CC Ink TUI fork (R3 §3.10, §5 div 1) | 750KB vendored TUI duplicates work; moai rides CC's renderer. | Never — moai stays headless |
| CC GrowthBook feature flags (R3 §5 div 4) | OSS binary cannot hide commands behind runtime flags. | Never — moai uses explicit config keys |
| CC Centralized MCP registry (R3 §5 div 5) | moai treats MCP as per-project; no marketplace. | v4 marketplace if needed |
| CC 52-subcommand CLI (R3 §5 div 6) | Scope reflects CC-internal enterprise tooling. | Never — moai's ~11 subcommands are right-sized |
| X-4 3-Origin Plugin System | Adds second extension axis alongside skills; confusing surface. | v4+ if skill system proves insufficient |
| O-2 Publish-Subscribe Message Pool | Current moai team sizes don't warrant; direct SendMessage works. | Team mode past ~10 teammates |
| M-4 Workflow Memory Induction (AWM) | Auto-induction premature; needs trajectory corpus. | After 100+ SPEC completions with telemetry |
| R-4 Tree-of-Thoughts | 5-100× token cost; thorough-harness opt-in only. | Evidence of plan-phase failure modes |
| R-5 CodeAct Python sandbox | Needs S-3 to ship first. | Post-S-3, targeted ACI extension |
| E-4 Pass@N default | N× cost; thorough-harness opt-in only. | Thorough-harness metrics show pass@1 regression |
| LangGraph graph framework | R2 anti-pattern 9 (framework-over-primitives). | Never — file-first primitive preference |
| DSPy compiler / prompt optimization | Research framework; production effort high. | v5+ if prompt quality plateaus |
| ADAS Meta-Agent harness synthesis | Far-horizon; needs S-4/S-5 maturity. | v5+ as experimental mode |

---

**End of Artifact 3.**
