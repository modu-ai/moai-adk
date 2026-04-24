# MoAI-ADK v3 Design Principles

> Wave 2 synthesis artifact 1 of 3
> Source: Convergence of R1 (AI harness papers) ∩ R2 (opensource tools) ∩ R3 (CC architecture re-read)
> Date: 2026-04-23
> Scope: 12 foundational principles that should govern every v3 architectural decision

---

## Preamble

Principles differ from rules. A rule says "do X." A principle says "when forced to choose between A and B, A wins — for this reason — unless a higher principle overrides." Rules are verifiable at commit time; principles are adjudicated at design time.

The 12 principles below are not arbitrary preferences. Each is grounded in at least two Wave 1 sources (an academic paper, a battle-tested opensource tool, or CC's own architectural decisions). They form a loose partial order: principles higher in the hierarchy constrain lower ones when they conflict.

A v3 SPEC decision that contradicts a principle MUST either (a) cite a more senior principle it is defending, or (b) escalate to the constitution FROZEN zone for human amendment. Principles are not infallible, but deviations must be argued explicitly — never silently.

---

## Principle 1: SPEC as Constitutional Contract

**Statement**: The SPEC document — with EARS requirements, acceptance criteria, and typed artifacts — is the single source of truth across phases. LLM context, agent memory, and generated code all derive from the SPEC; none of them override it.

**Rationale**:
- R1 §7 MetaGPT (Hong et al. 2024, ICLR Oral): "Encode Standard Operating Procedures as prompt sequences ... Structured intermediate artifacts (docs, diagrams, APIs) serve as typed handoffs between roles." MetaGPT empirically validates that typed artifacts between roles reduce hallucination in multi-agent collaboration.
- R1 §18 Constitutional AI (Bai et al. 2022): Declarative constitution + self-critique is cheaper and more predictable than hand-coded guards.
- R3 §6 "moai's big idea": "moai-adk is a SPEC-governed development workflow orchestrator ... CC asks 'how do I safely run one more model turn?', moai asks 'how do I ship the SPEC correctly?' ... the product bet is that the SPEC is the source of truth."
- R2 §7 Continue: `.continue/checks/*.md` markdown-authored agents validate the pattern of markdown + YAML frontmatter as a constitutional configuration surface.

**Application**:
- Every phase (plan, run, sync) reads from the same SPEC document; no agent reinvents requirements mid-flow.
- Acceptance criteria are Given/When/Then testable statements, not subjective judgments.
- When LLM context and SPEC disagree, the SPEC wins and the disagreement must surface as a blocker report, not a silent interpretation.
- SPEC itself is version-controlled; every SPEC edit is a git commit with a conventional message.

**Conflicts with**: Principle 3 (fresh context) when a SPEC grows large enough that reloading it per iteration is expensive — resolved by truncation priority in `.moai/design/` token budget (currently `spec.md > system.md > research.md > pencil-plan.md` per design constitution §3.2).

---

## Principle 2: Interface Design Over Tool Count (ACI)

**Statement**: The Agent-Computer Interface — the shape, naming, response structure, and error format of the tools an agent sees — matters more than the number of tools or the model's raw capability. Design LM-optimized interfaces, not human-optimized ones.

**Rationale**:
- R1 §11 SWE-agent (Yang et al. 2024, NeurIPS): "LM agents are a new end-user class and deserve purpose-built ACIs." 12.5% SWE-bench pass@1 and 87.7% HumanEvalFix (both SOTA at release) from interface design, not model upgrades.
- R2 §8 SWE-agent deep analysis: "ACI is the single most-important architectural insight — don't expose raw shell; design a compact command set optimized for LM cognition." Empirically 8× improvement over shell-only agents.
- R2 §18 OpenCode: "LSP auto-loaded per language" — validates the ACI principle by making language-server commands a first-class tool family.
- R3 §1.1: CC's Tool abstraction bundles `(input schema, permission definition, permission prompt UI)` — every tool is a purpose-built interface, not a shell passthrough.

**Application**:
- The raw `Bash` tool is permitted but discouraged for common operations. Common moves (find a symbol, read a function body, apply a structural edit, check an LSP diagnostic, add an @MX tag) should have dedicated named operations with structured, concise feedback.
- moai's existing hook layer (PostToolUse, PreToolUse), LSP integration (SPEC-LSP-CORE-002), @MX tag protocol, and ast-grep tool are pieces of an emerging ACI. v3 should audit them as a single interface and smooth rough edges.
- Tool output must be LM-readable first, human-readable second. Transform compiler stderr, test output, and git porcelain into structured observations before returning.

**Conflicts with**: Principle 11 (file-first primitives). Too many named tools recreate a framework. Bounded by "~15-25 named operations covering 95% of moves" per Aider's 9-command observation (R2 §6).

---

## Principle 3: Fresh-Context Iteration Over Session Accumulation

**Statement**: Long-horizon tasks should be decomposed into iterations, each starting from a clean LM context with state persisted to disk (SPEC docs, MX tags, checkpoint files, git history) — not carried across in the prompt. Session accumulation is a context-rot tax that LM cost and quality both pay.

**Rationale**:
- R1 §1 ReAct, §31 anti-pattern flag: "Pure ReAct can degenerate into very long token traces on long-horizon tasks."
- R1 §2 Reflexion: Explicitly scopes memory per trial; cross-trial state is the episodic buffer file.
- R2 §1-2 Ralph loops (iannuttall + snarktank): File-as-memory pattern beats elaborate context compression; fresh agent context per iteration is a proven primitive with ~30k GitHub stars in aggregate adoption.
- R2 §3 oh-my-claudecode: `/ralph` is a first-class operating mode, not an add-on.
- R3 §3.2 (CC technical debt): BigQuery data showed 1,279 sessions with 50+ consecutive autocompact failures wasting 250k API calls/day — the circuit breaker exists because unbounded accumulation is empirically harmful.

**Application**:
- `/moai loop` and `/moai run` with retry semantics should explicitly reset LM context between iterations and rebuild from the SPEC, not chain transcripts.
- Checkpoint files at `.moai/state/` are the cross-iteration state; the prompt builds from them, not from accumulated turn traces.
- Ralph's `STALE_SECONDS` crash-recovery primitive maps to moai's session-state TTL — explicit stale detection should trigger fresh restart, not resume.

**Conflicts with**: Principle 5 (typed state + durable checkpoint). Resolution: the checkpoint is durable; the LM context that consumes the checkpoint is ephemeral. Both compatible.

---

## Principle 4: Evaluator Judgments Fresh, Contract State Durable

**Statement**: Agent-as-a-Judge evaluators must start each iteration with no memory of previous judgments; what persists across iterations is the typed Sprint Contract (pass/fail per criterion, active checklist), not the evaluator's reasoning about why something failed last time.

**Rationale**:
- R1 §9 Agent-as-a-Judge (Zhuge et al. 2024) anti-pattern flag (direct quote): "any errors in previous judgments could lead to a chain of errors." Historical judgment memory cascades errors in judges.
- R1 §2 Reflexion: Actor/Evaluator/Self-Reflection separation validates that the evaluator is a distinct module, not a continuation of prior reasoning.
- R2 §16 LangGraph: Checkpointer saves typed state, not conversation memory — the separation is already a pattern at the graph level.
- R1 §8 Generative Agents: Observations and reflections are separate layers; reflections compress but do not contaminate raw observations.

**Application**:
- The Sprint Contract (moai design constitution §11.4) carries forward CRITERIA across iterations (passed criteria must not regress) but must NOT carry forward judgment transcripts.
- `evaluator-active` must be spawned with a fresh context that sees: (a) the BRIEF, (b) the Sprint Contract state, (c) the artifact to evaluate. It must NOT see prior scoring rationale.
- Judgment outputs feed back into Sprint Contract state (passed/failed/refined criteria), not into evaluator's next context.
- Add to design.yaml: `evaluator.memory_scope: per_iteration` with `contract.state: durable`.

**Conflicts with**: Current moai design/constitution.md §11.4 implicitly retains memory. This principle forces an amendment.

---

## Principle 5: Typed State + Durable Checkpoint at Phase Boundaries

**Statement**: Cross-phase and cross-iteration state is a typed schema with immutable updates, checkpointed at every phase boundary. "Whatever disk looks like at phase boundary = whatever the next phase reads" — no implicit context, no pattern-matching a conversation history.

**Rationale**:
- R1 §17 DSPy: "Signatures define I/O; optimizers produce optimised prompts + few-shot demos." Typed I/O is the unit of program-level reasoning, not free-form text.
- R1 §24 AWM: "Workflows are first-class agent memory objects, distinct from raw episodic memory." Typed workflow memory enables reuse across workloads.
- R2 §16 LangGraph: "Typed state schema with immutable updates — prevents race conditions in parallel agent work, solves a problem moai currently papers over with file-ownership conventions." Checkpointer-per-step for durable execution; `interrupt()` primitive for clean HITL.
- R2 §12 Microsoft Agent Framework 1.0: "Checkpointing + hydration for long-running" — GA April 2026 after absorbing AutoGen.
- R3 §1.1: CC's Session/QueryEngine/Turn are typed, bounded lifetimes with explicit state shape.

**Application**:
- Define a typed `SessionState` / `PhaseState` schema in `internal/config/types.go` (today 5 yaml sections lack loaders per R6 §5.2 — constitution, context, interview, design, harness).
- At each phase boundary (plan→run→sync), write a checkpoint file that fully captures the inputs the next phase needs. Never rely on "the conversation that just happened."
- `interrupt()`-equivalent: when the orchestrator needs a user decision, produce a typed `BlockerReport` that surfaces to AskUserQuestion; the resume point reads the resolved answer from state, not from a dialog transcript.
- Resumable agents (moai claims this) need formal `checkpoint()` + `hydrate()` semantics, not implicit session replay.

**Conflicts with**: None structural. Tension with Principle 11 (file-first primitives) if typed state becomes heavy framework. Bounded: checkpoints are markdown + yaml + json files on disk, not a graph framework.

---

## Principle 6: Permission Bubble Over Bypass

**Statement**: Tool permissions resolve through a multi-source stack with provenance (policy > user > project > local > plugin > skill > session > builtin). When risk is ambiguous, escalate to the parent terminal (bubble) rather than default-allow or default-deny. Permission modes are first-class values, not boolean flags.

**Rationale**:
- R3 §2 Decision 15: CC's `bubble` permission mode for fork agents. "A fork that inherits context SHOULD ask the parent-terminal's user for permission (bubble), not the teammate's mailbox. `bubble` is a permission mode value, which means it's a first-class primitive, not a boolean special case."
- R3 §4 Adoption Candidate 2: "moai today has implicit trust ... adding a permission envelope would enable: pre-allowlist for common dev ops, bubble-to-user for risky cases, and programmatic rule rewriting via hook responses."
- R2 §5 Cline 2026 npm-token exfiltration incident: "approval-only safety — the 2026 npm-token exfil proved approval fatigue is exploitable; moai must not repeat."
- R2 §14 Open Interpreter: Cited in 2026 security reviews as "do not run on host" — the anti-pattern moai must diverge from.

**Application**:
- Introduce explicit permission modes: `default`, `acceptEdits`, `plan`, `bubble`. No silent `bypassPermissions` in agent frontmatter without SPEC-documented justification.
- Every config value carries a `source` tag (R3 Adoption 1). v3 `.moai/config/sections/` should expose provenance: "this rule came from which file."
- Hook response JSON (Principle 8) becomes the programmable rule-rewriting channel: a PreToolUse hook returning `{permissionDecision: "ask"}` bubbles to user.
- Read-only agents default to `plan` mode; implementer agents default to `acceptEdits` scoped to worktree.

**Conflicts with**: Implementation velocity. Bubble prompts fatigue users. Resolution: bubble only when policy says risk is non-trivial; use pre-approved allowlists for common dev ops (`go test`, `ruff check`, etc.).

---

## Principle 7: Sandboxed Execution as Default Safety Layer

**Statement**: Agent-initiated code execution runs in an ephemeral, isolated sandbox by default — not the host machine. Approval prompts alone are insufficient; defense-in-depth requires ephemeral isolated execution, network egress control, and file-write scope enforcement.

**Rationale**:
- R2 Executive summary: "Safety architecture has become a first-class concern in 2026. Post-incidents (Claude Code's `rm -rf ~/`, Cline prompt-injection → npm token exfiltration, Alibaba LLM spontaneous cryptomining), the consensus is that approval prompts alone are insufficient."
- R2 §A Top-5 Pattern 5: "Ephemeral sandboxed execution as the default safety layer ... OWASP Top 10 for Agentic Apps (Dec 2025) codifies this as mandatory. Defense-in-depth requires ephemeral isolated execution (E2B / Modal / Cloudflare V8 / Bubblewrap / Seatbelt / Landlock), network egress blocking, file-write scope enforcement."
- R2 §B Anti-pattern 1: "Approval-fatigue-only safety ... moai currently has `permissions.allow` lists but no sandbox — matches the anti-pattern."
- R2 Sandboxing column in Design Space Taxonomy: **every surveyed tool except snarktank/ralph has "none" for sandboxing**. moai can lead by correcting this.
- R1 §12 CodeAct anti-pattern flag: "Sandboxing is non-trivial. Moai's existing shell-hook permission model and `bypassPermissions` controls are the right LAYER for code-action safety" — but sandboxing itself still missing.

**Application**:
- v3 adds a `sandbox` config section. Opt-in per agent role: implementer-role agents (manager-ddd/tdd, expert-backend/frontend, builder-platform) default to sandboxed execution; research/reviewer agents read-only (no sandbox needed).
- Sandbox transport per platform: Bubblewrap (Linux), Seatbelt (macOS native `sandbox-exec`), Docker (CI). 16-language neutrality preserved because sandbox operates on the shell layer, not the language.
- Default network egress denylist for implementation teammates; allowlist for `github.com`, `registry.npmjs.org`, `pypi.org`, `proxy.golang.org` etc., extensible per project.
- Document the exact sandbox invocation in `.moai/config/sections/security.yaml` (R6 §5.1: security.yaml already exists and is loaded).

**Conflicts with**: Principle 11 (file-first primitives). Sandbox adds infrastructure. Resolution: sandbox is a thin process-level shell wrapper; the "framework" it adds is ~1 config block + per-agent flag, not a full runtime.

---

## Principle 8: Hook Output = JSON Protocol (Exit Codes Remain as Fallback)

**Statement**: Hook handlers communicate via structured JSON output capable of carrying `additionalContext`, `permissionDecision`, `updatedInput`, `systemMessage`, and `continue` fields. Exit codes remain as a backward-compatible fallback for bash-script authors, but all first-party moai hooks return JSON.

**Rationale**:
- R3 §2 Decision 5: "Hook output protocol is JSON-OR-exitcode, not just exitcode ... rich hook capabilities (`additionalContext`, `permissionDecision`, `updatedInput`, `watchPaths`) require structured output. But exit codes remain for scripts without JSON producers."
- R3 §4 Adoption Candidate 4: "moai hooks today signal only via exit code. A JSON protocol would make Sprint Contract negotiation and MX tag injection actually programmable."
- R2 §3 oh-my-claudecode: OpenClaw hook bridge with 6 events + template variables (`{{sessionId}}`, `{{projectName}}`, `{{prompt}}`) — adjacent-tool recognition that hook data needs structure.
- R6 §2.4 R6 §A: 10 of 27 moai handlers are "logging-only" no-ops because they have nothing useful to emit via exit code alone — the protocol limits the handler's ambition.

**Application**:
- `internal/hook/*.go` handlers migrate to return typed JSON structs (`HookResponse`) serialized to stdout. Backward-compat: empty JSON or missing JSON falls back to exit code parsing.
- `PreToolUse` JSON can return `{updatedInput: {...}}` to mutate tool input mid-turn.
- `PostToolUse` JSON can return `{additionalContext: "@MX:WARN at line 42..."}` to inject observations.
- `SubagentStop` JSON returns `{continue: false}` only when teammate should be blocked from idling — enables quality gate enforcement programmatically.
- Document the schema in `.claude/rules/moai/core/hooks-system.md`; migrate handlers in v3 sequencing (5 stubs upgraded first per R6 recommendations, then broader migration).

**Conflicts with**: Existing 26 shell wrappers that currently parse exit codes. Resolution: shell wrappers unchanged; Go handlers emit JSON on stdout; Claude Code reads JSON first, falls back to exit code.

---

## Principle 9: Parallelism via Explicit Dependency DAG

**Statement**: When multiple tasks or agents can execute concurrently, represent their dependencies as an explicit DAG — not an implicit ordering in a task list. File-ownership per agent, read vs write declaration, and independent-task parallelization are the compiler-style primitives; the harness schedules, the agents execute.

**Rationale**:
- R1 §15 LLMCompiler (Kim et al. 2023, ICML): "Dependency-graph-driven parallel execution ... 3.7× latency improvement, 6.7× cost reduction, +9% accuracy over ReAct."
- R1 §16 ADAS: "Harness topology itself is learnable" — validates that execution graphs are a first-class object.
- R2 §16 LangGraph: "Parallel agent work without typed immutable state is a race-condition farm. Checkpoint-per-step lets long-running workflows survive failures."
- R3 §2 Decision 9: "Worktree isolation as an orthogonal axis to agent isolation. ... read-only agent needs context isolation but NOT worktree; a fork subagent needs fast shared context AND shared worktree. Decoupling the axes yields 4 deliberate combinations."
- R5 P-A11: 6 cross-file-write moai agents missing `isolation: worktree` — current moai relies on implicit ordering, not explicit dependency modeling.

**Application**:
- `/moai plan` output should produce a task DAG with explicit `depends_on` fields per task, not only a sequential task list.
- Team mode file-ownership (current moai) is a partial DAG — formalize as declared `reads: [paths]` + `writes: [paths]` per agent spawn.
- LLMCompiler's Planner → Task Fetching Unit → Executor maps to: moai-spec (plan) → manager-strategy (DAG synthesis) → per-agent execution via `Agent(isolation: worktree)`.
- Sequential-only invocations are a degenerate DAG; the tooling supports both but documents the explicit-DAG path as preferred for team mode.

**Conflicts with**: Simplicity for single-agent flows. Resolution: DAG is optional for N=1 agent invocations; required for N≥2 with write overlap.

---

## Principle 10: Agent Count Matches Task Structure

**Statement**: Well-structured tasks (known inputs, known outputs, deterministic fix criteria — e.g., linting, formatting, LSP error resolution, coverage gap filling) use fixed Agentless-style pipelines with at most one LLM agent. Open-ended tasks (requirements analysis, architecture design, ambiguous feature requests) use multi-agent role specialization.

**Rationale**:
- R1 §25 Agentless (Xia et al. 2024) — primary thesis: "Fixed-pipeline > autonomous decision flow when the task structure is well-understood. Complexity is a liability, not a feature." 27.3% SWE-bench Lite outperforming all open-source agentic competitors at lower cost.
- R1 §7 MetaGPT: Role specialization validates for open-ended collaboration.
- R1 Cross-cutting §Y Divergence: "Keep multi-agent for open-ended SPEC work (plan/design); collapse to single-agent pipeline for well-structured utility subcommands (fix, coverage, codemaps, mx, clean)."
- R2 Executive summary: Two architectural camps — persistent file-first loops (Agentless-adjacent) vs graph/state orchestration (multi-agent). Neither wins globally.

**Application**:
- v3 subcommand classification:
  - **Multi-agent**: `/moai plan`, `/moai run`, `/moai design`, `/moai sync` (phase-typed pipelines with multiple specialist roles).
  - **Single-agent fixed pipeline**: `/moai fix`, `/moai coverage`, `/moai codemaps`, `/moai mx`, `/moai clean` (deterministic input → deterministic transform → verify).
  - **Loop primitive** (fresh-context iteration): `/moai loop` wraps a fixed pipeline with retry semantics.
- Agentless-shape subcommands should not spawn subagents when the task is well-structured; they delegate to CC tools directly (Grep, ast-grep, LSP).
- When in doubt, default to fewer agents; the classification decision itself is a Philosopher Framework exercise (moai-foundation-thinking) at /moai plan time.

**Conflicts with**: Historical moai habit of "always delegate to specialist." Resolution: delegate is still the default for open-ended tasks; for utility subcommands, measure empirically (pass@1, cost) before multi-agent promotion.

---

## Principle 11: File-First Primitives Over Framework Abstractions

**Statement**: Prefer git + disk + markdown + YAML as the state and configuration substrate over bespoke frameworks (LangGraph, AutoGen topologies, DSPy modules). State that fits in a file, readable by humans and machines alike, auditable via `git log`, beats any in-memory graph or typed DAG library.

**Rationale**:
- R2 §1 Ralph: "File-as-memory pattern beats elaborate context compression ... persistence purely via git + on-disk files. Context window is thrown away every iteration." Validated by 895 stars on iannuttall/ralph and 30.9k stars on oh-my-claudecode's `/ralph` mode.
- R2 §B Anti-pattern 9: "Framework-over-primitives complexity — LangGraph, AutoGen, Agent Framework 1.0 — powerful but steep. Users reach for Ralph instead because 200 LOC > 200 KLOC. moai risk: v3 could over-engineer; Ralph's minimalism is a competitive benchmark."
- R2 §11 smol-ai/developer: Dormant since 2024 validates that without persistence, prompt-only approach plateaus; BUT the markdown-all-you-need insight is universal — "`prompt.md` approach works if combined with state persistence."
- R3 §1.1: CC's Memory layer is markdown files with YAML frontmatter. The plugin system uses file-scoped origin (`builtin`/`installed`/`inline`). Provenance flows through files.
- moai CLAUDE.local.md §14: "[HARD] Go 코드 하드코딩 금지 ... URL, 모델명 → const로 추출" — moai's own rules align with file-first externalization.

**Application**:
- New v3 state MUST materialize as a file in `.moai/state/`, `.moai/specs/`, `.moai/sprints/`, or similar — NOT as an in-memory struct that doesn't survive process restart.
- When choosing between "typed Go struct" and "YAML section with loader," prefer both (typed struct reads YAML), but the yaml file is the authoritative artifact.
- Reject frameworks that require > ~500 LOC of scaffolding. Measure marginal complexity vs primitive.
- Ralph loop's `progress.md`, `activity.log`, `errors.log`, `runs/*` shape is the canonical file-first layout.

**Conflicts with**: Principle 5 (typed state). Resolution: the type is the schema that the file conforms to; both are coherent if the type is thin and the file is primary.

---

## Principle 12: Constitutional Governance with FROZEN/EVOLVABLE Zones

**Statement**: Architectural rules and invariants live in declarative constitution files with explicit FROZEN (immutable by any automation, human-only changes) and EVOLVABLE (may be amended by graduation protocol) zones. No rule silently changes; every amendment is auditable, reversible, and canary-tested.

**Rationale**:
- R1 §18 Constitutional AI (Bai et al. 2022): "Explicit declarative constitution governs agent behavior; self-critique implements the constitution at runtime."
- R1 §16 ADAS anti-pattern flag: "Meta-agents can drift unsafe. Moai's FROZEN zone + canary check + human approval (constitution §5) is directly appropriate defense."
- R3 §4 Adoption Candidate 7: "Typed memory taxonomy (user/feedback/project/reference) — moai MEMORY.md informally already uses this schema; formalizing it as a constitutional rule prevents drift."
- moai's existing `.claude/rules/moai/core/moai-constitution.md` + `.claude/rules/moai/design/constitution.md` v3.3.0 already embody this pattern. Principle is "codify and expand, do not regress."

**Application**:
- FROZEN zone: this principles document itself, TRUST 5 framework, SPEC-First requirement, 16-language neutrality, EARS format invariance, AskUserQuestion-only-for-orchestrator rule, sandbox default, @MX tag protocol.
- EVOLVABLE zone: skill bodies, evaluator rubrics, harness level thresholds, pipeline adaptations, brand tokens.
- Every amendment requires: (a) canary check (R1 §5 Layer 2 — shadow evaluation), (b) contradiction detector (R1 §5 Layer 3 — no silent override), (c) human approval gate (R1 §5 Layer 5), (d) rollback protocol (design constitution §14).
- Constitutional sprawl is a real risk (R1 §18 anti-pattern flag: "Constitutions that are too long or contradictory confuse models"); v3 should consolidate `.claude/rules/moai/` from 34 files to ~28 per R6 recommendations.

**Conflicts with**: Speed of iteration. Resolution: EVOLVABLE zone is the fast lane; FROZEN changes go through SPEC workflow like any other invariant change.

---

## Principle interaction matrix

Read row × column as "how does row P_i interact with column P_j?" R = reinforces; C = conflicts, resolution required; I = independent.

| | P1 SPEC | P2 ACI | P3 Fresh | P4 Eval | P5 State | P6 Perm | P7 Sandbox | P8 Hook | P9 DAG | P10 Count | P11 File | P12 Const |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| **P1 SPEC** | — | R | R | R | R | I | I | I | R | R | R | R |
| **P2 ACI** | R | — | I | I | I | R | R | R | I | I | R | I |
| **P3 Fresh** | R | I | — | R | C→split | I | I | I | R | R | R | I |
| **P4 Eval** | R | I | R | — | R | I | I | I | I | I | R | R |
| **P5 State** | R | I | C→split | R | — | I | I | R | R | I | C→thin | R |
| **P6 Perm** | I | R | I | I | I | — | R | R | I | I | I | R |
| **P7 Sandbox** | I | R | I | I | I | R | — | I | I | I | C→thin | R |
| **P8 Hook** | I | R | I | I | R | R | I | — | R | I | R | I |
| **P9 DAG** | R | I | R | I | R | I | I | R | — | R | I | I |
| **P10 Count** | R | I | R | I | I | I | I | I | R | — | R | I |
| **P11 File** | R | R | R | R | C→thin | I | C→thin | R | I | R | — | R |
| **P12 Const** | R | I | I | R | R | R | R | I | I | I | R | — |

Three conflicts to note:
- **P3 ↔ P5**: Fresh-context iteration vs durable checkpoint. Resolution: checkpoint is durable on disk; LM context that consumes it is fresh. Both compatible.
- **P5 ↔ P11** and **P7 ↔ P11**: Typed state and sandbox add framework surface. Resolution: keep both thin — typed state is ~1 yaml loader per section, sandbox is ~1 shell wrapper per agent. Under ~500 LOC framework budget total.

## Principle hierarchy (when principles compete)

**Tier 1 — Never compromise (FROZEN invariants)**:
1. P1 SPEC as Constitutional Contract
2. P12 Constitutional Governance with FROZEN/EVOLVABLE Zones
3. P7 Sandboxed Execution as Default Safety Layer

**Tier 2 — Structural invariants (change requires constitutional amendment)**:
4. P2 Interface Design Over Tool Count (ACI)
5. P4 Evaluator Judgments Fresh, Contract State Durable
6. P5 Typed State + Durable Checkpoint at Phase Boundaries
7. P6 Permission Bubble Over Bypass
8. P8 Hook Output = JSON Protocol

**Tier 3 — Optimization principles (refine based on empirical data)**:
9. P3 Fresh-Context Iteration Over Session Accumulation
10. P9 Parallelism via Explicit Dependency DAG
11. P10 Agent Count Matches Task Structure
12. P11 File-First Primitives Over Framework Abstractions

When Tier 1 conflicts with any lower tier, Tier 1 wins. When same-tier principles conflict, resolution requires a SPEC-documented trade-off decision with Philosopher Framework analysis.

---

## v3 North Star (single-paragraph mission)

MoAI-ADK v3 is a **SPEC-governed, language-neutral, harness-routed development orchestrator that rides on Claude Code's turn runtime** — treating the EARS-formatted SPEC document, its Given/When/Then acceptance criteria, and its associated @MX tags as the constitutional contract that every phase (plan, run, sync, design, review) reads from and writes to. It combines (a) Ralph-style file-first persistence with fresh-context iteration for the run and loop phases, (b) MetaGPT-style role-specialized multi-agent orchestration for open-ended plan and design work, (c) Agent-as-a-Judge independent adversarial evaluation with **fresh judgments per iteration** and durable Sprint Contract state, (d) Constitutional AI-style declarative governance with explicit FROZEN and EVOLVABLE zones, (e) SWE-agent-style Agent-Computer Interface (@MX tags + LSP integration + ast-grep + structured hook JSON) as the preferred tool layer, (f) LangGraph-style typed state checkpointed at phase boundaries with `interrupt()`-equivalent blocker reports surfacing to AskUserQuestion, and (g) TRUST 5 quality gates routed by a 3-level harness system (minimal/standard/thorough) that scales evaluation depth with SPEC complexity. It differs from **Claude Code** (a safety-enveloped turn runtime — moai is the workflow on top, never a replacement), from **Ralph** (which has no intent contract — moai adds SPEC/EARS as the contract that Ralph's loop iterates against), from **oh-my-claudecode** (which has a rich mode surface but no intent governance, no TRUST 5, no EARS, no 16-language neutrality), from **Aider** (which is single-agent and git-commit-native but lacks SPEC contracts and multi-agent roles), and from **SWE-agent / Agentless** (which are fixed pipelines — moai uses Agentless patterns for utility subcommands like `/moai fix` while retaining multi-agent for open-ended phases). Moai v3's **unique position**: the only agentic development orchestrator that combines a constitutional SPEC contract with adversarial skeptical evaluation, typed phase transitions, sandboxed execution by default, and language-neutral harness routing — betting that explicit, externalized intent plus adversarial verification produces more correct software than any autonomy-maximizing agent, at a complexity cost bounded by harness routing and file-first primitives.
