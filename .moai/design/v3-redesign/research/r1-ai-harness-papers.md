# R1 — AI Harness Papers Survey

> Research team: R1
> Papers surveyed: 24
> Date: 2026-04-23
> Scope: Academic + industry literature on LLM agentic harnesses, reasoning loops, multi-agent orchestration, code agents, tool use, evaluation, and safety — with explicit mapping to moai-adk v3 design questions

---

## Executive summary

As of April 2026, the LLM agentic-harness literature has converged on several structural idioms that did not exist three years ago. The canonical loop — **Observe → Reason → Act → Reflect** — first formalized as ReAct (Yao et al. 2022) now appears in some form in every serious coding agent, including SWE-agent, Devin, Agentless, and Anthropic's Claude Code itself. Where 2022-era work focused on whether LLMs could *act at all* (Toolformer, ReAct), 2024-2026 work has shifted to three meta-questions: (a) **How many agents do you actually need?**, (b) **Where does learning/memory live?**, and (c) **How do you evaluate open-ended agentic work without drowning in surface metrics?**

On (a) there is active tension. Multi-agent frameworks (AutoGen, MetaGPT, CrewAI, OpenAI Swarm) advocate role decomposition with explicit handoffs or SOPs. Against this, Agentless (Xia et al. 2024) demonstrates that disciplined three-phase pipelines *without* LLM-driven control flow decisions can be state-of-the-art on SWE-bench Lite. The divergence matters: moai-adk v3 must decide whether "manager + expert + evaluator" is structurally necessary or whether it is 2023-era scaffolding that a 4.7-class model renders redundant. Anthropic's own effort-level guidance (documented for Opus 4.7 in April 2026) explicitly warns against Opus 4.6-era defensive scaffolding — "double-check X before returning", "verify N times" — because 4.7 follows instructions literally and such patterns become counterproductive.

On (b) — memory — the dominant pattern has shifted from undifferentiated "scratchpad + context" toward **typed memory layers**: episodic (Reflexion's memory buffer), semantic/skill (Voyager's skill library), and procedural workflow memory (AWM). Generative Agents added a consolidation step ("reflection") that compresses observations into higher-level beliefs. For moai-adk, the implication is that a single global context store is probably not enough; the harness should expose distinct slots for (i) task-specific observation trace, (ii) reusable skill / code fragments, and (iii) abstracted workflow templates.

On (c) — evaluation — Agent-as-a-Judge (Zhuge et al. 2024) formalized what SWE-bench's pass@1 metric could not: **intermediate-trajectory evaluation**. The empirical finding that LLM-as-a-Judge underperforms human eval but Agent-as-a-Judge matches human reliability, at ~97% of the cost, is directly applicable to moai-adk's `evaluator-active` agent. Combined with Constitutional AI's RLAIF loop, this points toward a harness pattern where evaluation agents operate continuously (not only at PR time) with constitution-anchored rubrics.

Three cross-cutting signals should guide moai v3. First, **execution as action**: CodeAct (Wang et al. 2024) and Voyager both argue that Python code is a more expressive and composable action space than pre-defined tool JSON — a finding reinforced by Chain of Code. Second, **parallelism as a first-class harness concern**: LLMCompiler shows 3.7× latency gains by treating tool orchestration as a classical compiler IR problem. Third, **the harness is increasingly programmable**: DSPy reframes prompts as learned parameters, and ADAS goes further by having a meta-agent discover harness topologies. Moai-adk v3 should decide explicitly whether its harness is handwritten (current) or learned/synthesized.

---

## Paper-by-paper catalog

### 1. ReAct: Synergizing Reasoning and Acting in Language Models
- **Citation**: Yao, Zhao, Yu, Du, Shafran, Narasimhan, Cao (2022) — arXiv:2210.03629, ICLR 2023
- **Contribution**: Introduces the interleaved Thought/Action/Observation prompting paradigm where the LLM alternates verbal reasoning with tool-based actions, grounding reasoning in external observations to reduce hallucination.
- **Key pattern**: Thought→Act→Observe loop where reasoning traces induce/track/update action plans while actions feed back external evidence. A single-model, single-turn-per-step paradigm.
- **Moai applicability**: This is the structural parent of `/moai run` and all moai subagents. The MX tag protocol (`@MX:NOTE`, `@MX:WARN`) can be reinterpreted as serialized ReAct reasoning traces preserved in source code. Hooks like PostToolUse already mirror the "Observation" step; they should feed structured observations back into subsequent Thought steps.
- **Anti-pattern flag**: Pure ReAct can degenerate into very long token traces on long-horizon tasks (confirmed later by ReWOO). Moai's current per-agent conversation style risks accumulating full-trace tokens when tasks exceed ~5 loop iterations.

### 2. Reflexion: Language Agents with Verbal Reinforcement Learning
- **Citation**: Shinn, Cassano, Berman, Gopinath, Narasimhan, Yao (2023) — arXiv:2303.11366, NeurIPS 2023
- **Contribution**: Separates the agent into **Actor / Evaluator / Self-Reflection** modules, where Self-Reflection converts environment signals into verbal feedback stored in an episodic memory buffer — achieving weight-free RL via natural-language gradient proxies.
- **Key pattern**: Verbal self-critique + episodic memory + multi-trial retry. 8% absolute improvement from reflection over memory-only baselines on AlfWorld.
- **Moai applicability**: This is the direct intellectual ancestor of moai's `evaluator-active` agent and the GAN loop (`.claude/rules/moai/design/constitution.md` §11). Moai already has Actor (expert-frontend) and Evaluator (evaluator-active) but lacks an explicit "Self-Reflection" module that converts failing scores into text gradients before the next iteration. The current harness implicitly expects the Actor to self-reflect, which violates Reflexion's separation.
- **Anti-pattern flag**: Reflexion shows that reflection-only with no memory is strictly worse than both retained. Moai's session-scoped `.claude/agent-memory/` is the right shape but is not guaranteed to persist across iterations within a single SPEC.

### 3. Self-Refine: Iterative Refinement with Self-Feedback
- **Citation**: Madaan et al. (2023) — arXiv:2303.17651, NeurIPS 2023
- **Contribution**: A single LLM acts as generator, critic, and refiner in a FEEDBACK→REFINE loop, yielding ~20% absolute improvement across 7 tasks even on GPT-4 without training.
- **Key pattern**: Single-model iterative refinement with explicit prior-iteration history in the prompt. No separate evaluator model needed.
- **Moai applicability**: Directly applicable to `/moai fix` and `/moai loop`. The current `moai-workflow-loop` (Ralph Engine) matches Self-Refine's shape. Paper validates that iterative single-model refinement is sufficient for many fix-type workflows; a separate evaluator agent may be over-engineered for LSP-driven fixes.
- **Anti-pattern flag**: Self-Refine can loop indefinitely without stopping criteria; moai's `max_iterations` in design.yaml correctly guards against this, but needs a stagnation detector (improvement < threshold) — which moai already has in GAN loop (§11) but not in `/moai loop`.

### 4. Tree of Thoughts: Deliberate Problem Solving with Large Language Models
- **Citation**: Yao, Yu, Zhao, Shafran, Griffiths, Cao, Narasimhan (2023) — arXiv:2305.10601, NeurIPS 2023
- **Contribution**: Extends Chain-of-Thought to a tree-search paradigm with BFS/DFS over "thought" units, with self-evaluation driving pruning. GPT-4 on Game-of-24: 4% (CoT) → 74% (ToT).
- **Key pattern**: Explicit deliberation with lookahead and backtracking; separates generation from evaluation in the search.
- **Moai applicability**: Applicable to `/moai plan` when multiple architectural approaches are plausible. Current SPEC workflow is linear (manager-spec → manager-strategy → expert-*); ToT suggests branching at the strategy phase with plan-auditor as the evaluator per branch.
- **Anti-pattern flag**: ToT is token-expensive (5-100× CoT). Use selectively; not for every task. Moai's harness tiering (minimal/standard/thorough) is the right knob.

### 5. Language Agent Tree Search (LATS)
- **Citation**: Zhou, Yan, Shlapentokh-Rothman, Wang, Wang (2023) — arXiv:2310.04406, ICML 2024
- **Contribution**: Unifies ReAct + Reflexion + ToT into a single MCTS-based framework with 6 operations (selection, expansion, evaluation, simulation, backpropagation, reflection). State-of-the-art 92.7% pass@1 on HumanEval with GPT-4.
- **Key pattern**: MCTS with UCT selection over ReAct-style trajectories, with LLM-computed value functions and verbal reflections feeding back into tree updates.
- **Moai applicability**: Too heavy for moai's normal flow, but informs the `thorough` harness level. Suggests that when `/moai run` spawns multiple implementation strategies in parallel (via team mode), a tree search with MCTS-style scoring could beat flat parallel evaluation.
- **Anti-pattern flag**: MCTS requires many simulations; for deterministic code tasks with clear acceptance criteria, simpler Reflexion-style loops match LATS at fraction of cost.

### 6. AutoGen: Multi-Agent Conversation Framework
- **Citation**: Wu, Bansal, Zhang, et al. (2023) — arXiv:2308.08155
- **Contribution**: General-purpose multi-agent framework where AssistantAgent/UserProxyAgent/GroupChatManager communicate via structured messages, with conversation-as-programming. Now in maintenance mode (Microsoft Agent Framework is successor).
- **Key pattern**: Conversable-by-default agents + flexible topology (sequential, group chat, hierarchical). Unified message-based control plane.
- **Moai applicability**: AutoGen's GroupChatManager maps to moai's MoAI orchestrator + manager-strategy. Moai's Team mode (SendMessage, TaskCreate/Update) is effectively an AutoGen-style control plane. One lesson: AutoGen's maintenance-mode transition suggests that static multi-agent topologies age quickly; moai's move to dynamic team generation (general-purpose + role profiles) is the right bet.
- **Anti-pattern flag**: Conversation cascades can hallucinate when messages compound — a direct motivator for MetaGPT's SOPs. Moai's TeammateIdle hook and TaskCompleted hook are defensive measures in the same spirit.

### 7. MetaGPT: Meta Programming for A Multi-Agent Collaborative Framework
- **Citation**: Hong, Zhuge, Chen, et al. (2024) — arXiv:2308.00352, ICLR 2024 Oral
- **Contribution**: Encodes Standardized Operating Procedures (SOPs) — Product Manager, Architect, Engineer, QA — as prompt sequences to constrain LLM hallucinations in multi-agent collaboration. Structured intermediate artifacts (docs, diagrams, APIs) serve as typed handoffs between roles.
- **Key pattern**: Role specialization + structured intermediate outputs + assembly-line task decomposition. Introduces an executive feedback loop that runs code and debugs at runtime (+5.4% MBPP).
- **Moai applicability**: Directly validates moai's role-based agent catalog (manager-spec, manager-strategy, expert-backend, expert-frontend, manager-quality, manager-docs). The typed-artifact handoff is essentially moai's SPEC document as constitutional contract between phases.
- **Anti-pattern flag**: MetaGPT's rigid SOP sequence can over-process simple tasks. Moai's harness levels (minimal/standard/thorough) and conditional agent invocation prevent this, but the default flow should skip phases aggressively when SPEC scope is small.

### 8. Generative Agents: Interactive Simulacra of Human Behavior
- **Citation**: Park, O'Brien, Cai, Morris, Liang, Bernstein (2023) — arXiv:2304.03442, UIST 2023
- **Contribution**: Three-layer memory architecture — observation stream + reflection (compressed beliefs) + plan (future actions) — enables multi-day behavioral consistency across 25 agents.
- **Key pattern**: Memory stream with importance/recency/relevance retrieval. Reflection = periodic synthesis of low-level observations into higher-level beliefs.
- **Moai applicability**: The reflection mechanism is directly applicable to moai's `lessons.md` auto-capture (SPEC-SLQG-001). Moai currently has append-only lessons; Generative Agents suggests a periodic consolidation pass that promotes repeatedly-observed patterns from observations → heuristics → rules (which matches the Learnings Pipeline in design constitution §6).
- **Anti-pattern flag**: Unbounded memory streams become unwieldy. Moai's `max_active_learnings: 50` with archival is correct; should add a consolidation cron similar to Generative Agents' sleep cycle.

### 9. Agent-as-a-Judge: Evaluate Agents with Agents
- **Citation**: Zhuge, Zhao, Ashley, et al. (2024) — arXiv:2410.10934
- **Contribution**: Extends LLM-as-a-Judge with agentic capabilities (tool use, memory, multi-step reasoning) to score intermediate trajectories, not just final outputs. Matches human evaluation reliability at ~97% cost savings on DevAI benchmark (55 dev tasks, 365 hierarchical requirements).
- **Key pattern**: Evaluator is itself an agent that observes the full trace. Enables "flywheel effect" — evaluator feedback becomes reward signal for the evaluated agent.
- **Moai applicability**: This is the intellectual charter for `evaluator-active`. The paper's finding that final-answer evaluation misses failure modes is exactly what moai's Sprint Contract + 4-dimension scoring addresses. Moai should adopt the hierarchical-requirement structure (365 sub-requirements per 55 tasks) as the shape for SPEC acceptance criteria.
- **Anti-pattern flag**: Paper explicitly documents that historical memory in judges can cascade errors — "any errors in previous judgments could lead to a chain of errors". Moai's evaluator should scope memory per iteration, not across iterations (contrary to what GAN loop §11.4 currently suggests).

### 10. OpenAI Swarm / Handoffs Pattern (2024)
- **Citation**: OpenAI Solutions team (Oct 2024) — github.com/openai/swarm. Succeeded by OpenAI Agents SDK (March 2025).
- **Contribution**: Minimal stateless framework with two primitives: Agents (instructions + tools) and Handoffs (function-returning-Agent). Replaces state machines with natural-language routines + transfer_to_X() tool calls.
- **Key pattern**: LLM-decided handoffs (not orchestrator-decided). Agent A calls transfer_to_B() when context warrants switching specialists. Stateless per run() call.
- **Moai applicability**: Currently moai's orchestrator (MoAI) decides all agent invocations explicitly. Swarm suggests an alternative where the active agent itself proposes handoffs to specialists (e.g., expert-backend detects security-sensitive code and hands off to expert-security). This is more fluid but reduces orchestrator observability. The migration to dynamic team generation (general-purpose + role profiles) is moai's analogue.
- **Anti-pattern flag**: Stateless per-call design loses context across handoffs unless explicit context variables are passed. Moai's session persistence + `CLAUDE_PROJECT_DIR` artifacts avoid this.

### 11. SWE-agent: Agent-Computer Interfaces Enable Automated Software Engineering
- **Citation**: Yang, Jimenez, Wettig, Lieret, Yao, Narasimhan, Press (2024) — arXiv:2405.15793, NeurIPS 2024
- **Contribution**: Argues that LM agents are a new end-user class and deserve purpose-built ACIs (Agent-Computer Interfaces). Achieved 12.5% SWE-bench pass@1 and 87.7% on HumanEvalFix (both SOTA at release) via custom file editor, repo navigator, and test runner tailored for LM consumption.
- **Key pattern**: **Interface design matters more than model capability**. Human IDEs are wrong for LM agents; e.g., SWE-agent's edit tool shows syntax errors immediately, filters search noise, and paginates predictably.
- **Moai applicability**: Strongest single lesson for moai-adk v3. Moai's current hook layer (PostToolUse, PreToolUse) is an ACI, but the tool set (Read, Write, Edit, Bash, Grep) is Claude-Code-native, not moai-tailored. Moai should consider domain-specific tools: `moai_spec_read`, `moai_run_tests_for_spec`, `moai_locate_mx_anchor`, etc., that return LM-optimized outputs.
- **Anti-pattern flag**: SWE-agent explicitly rejects "LM sees terminal output as human would" — implies moai hooks that dump raw stderr (e.g., Go compiler errors) should transform them into LM-optimized forms.

### 12. Executable Code Actions Elicit Better LLM Agents (CodeAct)
- **Citation**: Wang, Chen, Yuan, Zhang, Li, Peng, Ji (2024) — arXiv:2402.01030, ICML 2024
- **Contribution**: Replaces JSON tool calls with executable Python code as the action space. +20% success rate over Text/JSON actions; LLMs pre-trained on code handle code actions natively, compose multiple tools through Python control flow, and leverage error messages for self-debugging.
- **Key pattern**: Code-as-action + Python interpreter + multi-turn interaction where code errors feed back as observations. CodeActAgent (finetuned Llama2/Mistral) validates the paradigm on open models.
- **Moai applicability**: Moai currently uses Bash + tool-call JSON (Claude Code native). CodeAct suggests the harness could expose a Python sandbox where agents compose tools via Python rather than sequential tool calls. Alternatively, bash-as-action already serves the same role for unix-native tasks (git, go test). Core insight: **complex tool compositions should be one code block, not N tool calls**.
- **Anti-pattern flag**: Sandboxing is non-trivial. Moai's existing shell-hook permission model and `bypassPermissions` controls are the right layer for code-action safety.

### 13. Voyager: An Open-Ended Embodied Agent with LLMs
- **Citation**: Wang, Xie, Jiang, Mandlekar, Xiao, Zhu, Fan, Anandkumar (2023) — arXiv:2305.16291
- **Contribution**: Lifelong-learning agent in Minecraft with three components: (1) automatic curriculum (GPT-4 proposes next tasks), (2) skill library (executable code stored & retrieved by embeddings), (3) iterative prompting with environment feedback + self-verification.
- **Key pattern**: **Skills as executable code** accumulated across sessions; each task produces a reusable function added to a vector-indexed library.
- **Moai applicability**: Moai's Skill System (.claude/skills/) is already Voyager-shaped — skills are reusable capability units. Voyager suggests skills should accumulate dynamically from successful task trajectories (not only be human-authored). `/moai mx` already tags successful patterns; an auto-skill-distillation pass could convert @MX:ANCHOR functions into skills.
- **Anti-pattern flag**: Voyager's skill library grows unboundedly; practical deployments need GC. Moai should version skills and archive unused ones (currently no such mechanism).

### 14. Toolformer: Language Models Can Teach Themselves to Use Tools
- **Citation**: Schick, Dwivedi-Yu, Dessì, Raileanu, Lomeli, Zettlemoyer, Cancedda, Scialom (2023) — arXiv:2302.04761, NeurIPS 2023
- **Contribution**: Self-supervised training signal where API calls are annotated by the LM, filtered by whether they reduce next-token perplexity, then used for fine-tuning.
- **Key pattern**: LM learns tool use without human demonstrations; selects tools based on expected utility.
- **Moai applicability**: Indirectly — moai relies on prompted tool selection (agents are told which tools to use). Toolformer suggests automatic tool-selection could be learned; in practice, for moai this manifests as the Skill auto-loading system (YAML frontmatter triggers match context automatically).
- **Anti-pattern flag**: Paper shows that naïve tool-use prompting produces spurious tool calls. Moai's permission system + hook audits defend against this.

### 15. An LLM Compiler for Parallel Function Calling (LLMCompiler)
- **Citation**: Kim, Moon, Tabrizi, Lee, Mahoney, Keutzer, Gholami (2023) — arXiv:2312.04511, ICML 2024
- **Contribution**: Three-component system (Planner → Task Fetching Unit → Executor) treats tool orchestration as classical compilation. 3.7× latency improvement, 6.7× cost reduction, +9% accuracy over ReAct.
- **Key pattern**: **Dependency-graph-driven parallel execution**. Planner produces a task DAG; tasks without dependencies run in parallel.
- **Moai applicability**: Directly validates moai's CLAUDE.md §14 Parallel Execution Safeguards and the team mode's file-ownership-based parallelism. Moai's plan phase produces ordered phases; LLMCompiler suggests these should be a DAG, not a linear sequence, with explicit declaration of which phases can run in parallel.
- **Anti-pattern flag**: Naive parallel tool calls create file-write conflicts. Moai's dependency graph + worktree isolation is the correct mitigation.

### 16. Automated Design of Agentic Systems (ADAS)
- **Citation**: Hu, Lu, Clune (2024) — arXiv:2408.08435, ICLR 2025
- **Contribution**: Meta Agent Search — a meta-agent programs new agentic systems in code, iteratively improving on prior discoveries. Invented agents outperform hand-designed ones (CoT, Self-Refine, LLM-Debate, Quality-Diversity) across coding/science/math domains.
- **Key pattern**: **Harness topology itself is learnable**. Turing-complete code defines agents, so any agentic pattern is reachable via program synthesis.
- **Moai applicability**: Forward-looking. Moai's current builder agents (builder-agent, builder-skill, builder-plugin) are the manual version. ADAS suggests a `moai meta-design` mode where the harness redesigns a subsystem (e.g., evaluator scoring profile) based on observed failure modes. This is aligned with moai's design-constitution Section 6 (Learnings Pipeline → graduation).
- **Anti-pattern flag**: Meta-agents can drift unsafe. Moai's FROZEN zone + canary check + human approval (constitution §5) is directly appropriate defense.

### 17. DSPy: Compiling Declarative Language Model Calls into Self-Improving Pipelines
- **Citation**: Khattab, Singhvi, Maheshwari, et al. (2023) — arXiv:2310.03714, ICLR 2024
- **Contribution**: Reframes LM programs as computational graphs of parameterized modules. The DSPy compiler learns few-shot prompts or finetunes small LMs for each module. +25% over standard few-shot, +46% over expert prompts.
- **Key pattern**: **Prompts are learned parameters, not hand-written strings**. Separates module interface from implementation.
- **Moai applicability**: Moai's skills + agents currently use hand-authored prompts. DSPy suggests a `moai compile` step that optimizes agent prompts against project-specific training data (e.g., successful `/moai run` trajectories). The evaluator-profiles directory (`.moai/config/evaluator-profiles/`) is the right shape; DSPy says they should be learned rather than hand-tuned.
- **Anti-pattern flag**: Prompt optimization without eval discipline drifts. DSPy requires explicit validation metrics — aligns with moai's TRUST 5 as the metric source.

### 18. Constitutional AI: Harmlessness from AI Feedback
- **Citation**: Bai, Kadavath, Kundu, et al. (Anthropic, 2022) — arXiv:2212.08073
- **Contribution**: Two-phase method: (1) SL phase with self-critique + revision against a "constitution" of principles, (2) RLAIF phase using model-generated preference labels. Produces harmless-but-non-evasive assistants without human harm labels.
- **Key pattern**: **Explicit declarative constitution** governs agent behavior; self-critique implements the constitution at runtime.
- **Moai applicability**: Moai already uses this pattern — CLAUDE.md + .claude/rules/moai/core/moai-constitution.md + .claude/rules/moai/design/constitution.md form a tiered constitution. HARD rules are the invariant clauses. Moai's FROZEN zone mirrors Constitutional AI's immutable principles. Moai should extend the pattern to agent self-critique: before any agent outputs, it critiques against the relevant constitution subset.
- **Anti-pattern flag**: Constitutions that are too long or contradictory confuse models (Anthropic's internal finding). Moai's constitution is already ~800 lines across multiple files — risk of constitution-sprawl. Need a consolidation pass.

### 19. Claude Opus 4.7 Effort Parameter (Anthropic, 2026)
- **Citation**: Anthropic platform docs — https://platform.claude.com/docs/en/build-with-claude/effort (as of 2026-04)
- **Contribution**: Five-level effort parameter (low/medium/high/xhigh/max) replaces budget_tokens. Opus 4.7 uses Adaptive Thinking exclusively; budget_tokens manual config rejected with HTTP 400. `xhigh` is recommended starting point for coding/agentic work.
- **Key pattern**: **Model self-allocates reasoning tokens** via adaptive thinking. Effort is a behavioral signal, not a strict token budget. At lower effort levels the model scopes work to what was asked rather than over-delivering.
- **Moai applicability**: Direct. Moai's CLAUDE.md already has effort-level guidance (manager-spec/evaluator-active → xhigh; manager-git/Explore → medium). V3 should audit every agent's effort level. Also: the doc explicitly warns that Opus 4.6-era "double-check X" scaffolding is counterproductive on 4.7 — the moai-constitution.md already captures this as "Principle 4" and "Principle 5". V3 should audit all agent bodies for defensive scaffolding.
- **Anti-pattern flag**: Setting `budget_tokens` in Opus 4.7 agents → HTTP 400. Moai-adk's Go binary must never inject this.

### 20. SWE-bench: Can Language Models Resolve Real-World GitHub Issues?
- **Citation**: Jimenez, Yang, Wettig, Yao, Pei, Press, Narasimhan (2023) — arXiv:2310.06770, ICLR 2024 Oral
- **Contribution**: 2,294 real GitHub issues from 12 Python repos; evaluation via fail-to-pass test delta in containerized environments. Baseline Claude 2 = 1.96% → SOTA (Oct 2024) ~50%+.
- **Key pattern**: **Real issues + execution-based grading + containerized reproducibility**. The fail-to-pass criterion avoids LLM-judged fuzziness.
- **Moai applicability**: SWE-bench's fail-to-pass criterion is what moai's TDD workflow (RED-GREEN-REFACTOR) enforces at every SPEC. Moai should publish a moai-bench suite: `moai run SPEC-XXX` against a curated set of historical SPECs with captured pre-state and fail-to-pass tests, to self-measure harness regressions.
- **Anti-pattern flag**: SWE-bench's first-released version had test leakage and solvability issues (→ SWE-bench Verified subset). Any internal benchmark needs the same discipline.

### 21. MLE-bench: Evaluating ML Engineering Agents
- **Citation**: Chan, Chowdhury, Jaffe, et al. (OpenAI 2024) — arXiv:2410.07095
- **Contribution**: 75 Kaggle competitions as end-to-end ML engineering tasks. o1-preview + AIDE scaffolding → 16.9% bronze-medal rate (pass@1), 34.1% at pass@8. Validates scaling-with-attempts: more attempts → more medals.
- **Key pattern**: **Multi-attempt evaluation reveals capacity that pass@1 hides**. Contamination from pretraining is a real concern for agent benchmarks.
- **Moai applicability**: Moai currently evaluates each SPEC once. MLE-bench implies `/moai run` with N-retry (where N > 1) could meaningfully improve success — this is what `/moai loop` does for fixes but not for initial implementation. Consider `/moai run --attempts 3` with best-of-N evaluation. Also: moai's lessons.md auto-capture must guard against contamination (a lesson learned from prior SPEC should not be trivially re-used if it was the only reason the fix worked).
- **Anti-pattern flag**: Pass@1 metrics on code agents under-report capability. Moai's Sprint Contract should report pass@N, not just pass@1.

### 22. Devin / Cognition (2024)
- **Citation**: Cognition AI blog (March 2024) — cognition.ai/blog/introducing-devin
- **Contribution**: End-to-end autonomous SWE agent with shell + editor + browser, long-horizon reasoning (thousands of decisions), human-in-the-loop progress reporting. 13.86% SWE-bench pass@1 (10× previous SOTA).
- **Key pattern**: **Human-equivalent tools + long-horizon planning + human-in-the-loop collaboration**. Real-time progress reporting replaces black-box execution.
- **Moai applicability**: Moai already matches (shell + editor + optional browser via chrome-devtools MCP). Devin's real-time progress reporting maps to moai's TaskList + SendMessage in team mode. The critical Devin lesson for moai is **checkpoint granularity**: Devin commits progress at logical milestones so humans can intervene. Moai should commit after each phase (plan/run/sync) automatically, not only at the end.
- **Anti-pattern flag**: Post-2024 reviews of Devin have documented "theatrical autonomy" — agents that appear to work but produce low-quality output at scale. Moai's evaluator-active + plan-auditor guard against this by keeping an independent skeptical voice.

### 23. AgentBench: Evaluating LLMs as Agents
- **Citation**: Liu, Yu, Zhang, et al. (THUDM 2023) — arXiv:2308.03688, ICLR 2024
- **Contribution**: 8-environment multi-dimensional benchmark spanning code/game/web/database scenarios. Main finding: top commercial LLMs far exceed 70B OSS models on agent tasks; failure modes are long-term reasoning, decision-making, and instruction-following — not core capability.
- **Key pattern**: Multi-environment evaluation uncovers agent failure modes that single-task benchmarks miss.
- **Moai applicability**: Moai should evaluate the harness against multiple workload profiles, not only Go development. The harness-level config (minimal/standard/thorough) needs empirical validation across domains (web UI, backend API, data pipeline, infra).
- **Anti-pattern flag**: Paper explicitly finds that code pretraining is ambivalent for non-code agent tasks. Moai should not assume "better-on-code → better-on-everything".

### 24. Agent Workflow Memory (AWM)
- **Citation**: Wang, Mao, Fried, Neubig (2024) — arXiv:2409.07429, ICML 2024
- **Contribution**: Induces reusable workflows from past trajectories; stores them as agent memory and also wraps them as new high-level actions (AWMAS). +51.1% on WebArena over top autonomous method; outperforms human-expert workflows by 7.9%.
- **Key pattern**: **Workflows are first-class agent memory objects**, distinct from raw episodic memory. Can be induced offline or online, applied as memory context or action-space extension.
- **Moai applicability**: Moai's `.claude/skills/moai/workflows/` directory is the right shape. AWM suggests skills should be *auto-inducted* from successful `/moai run` trajectories, not only human-authored. The Graduation Protocol in design constitution §7 is the correct gatekeeper for promotion from observation → workflow.
- **Anti-pattern flag**: Naively saving all trajectories floods memory. AWM's abstraction step (removing example-specific context) is essential; moai's lesson entries already do some of this but need a proper abstraction pass.

### 25. Agentless: Demystifying LLM-based Software Engineering Agents
- **Citation**: Xia, Deng, Dunn, Zhang (2024) — arXiv:2407.01489
- **Contribution**: Three-phase non-agentic pipeline (localization → repair → validation) that *does not let LLMs decide future actions*. Achieves 27.3% on SWE-bench Lite, outperforming all open-source agentic competitors at lower cost. Later Claude 3.5 Sonnet version reaches 40.7% Lite / 50.8% Verified.
- **Key pattern**: **Fixed-pipeline > autonomous decision flow when the task structure is well-understood**. Complexity is a liability, not a feature.
- **Moai applicability**: Provocative counter-point to moai's agent catalog. For well-structured tasks (fix, lint, typecheck, format), a fixed pipeline may match or beat a multi-agent flow. Moai's `/moai fix` (Ralph Engine) is already Agentless-shaped. The question for v3: should more of the 14 subcommands collapse to fixed pipelines? Likely candidates: `coverage`, `codemaps`, `mx`, `clean`, `e2e`.
- **Anti-pattern flag**: Over-applying the Agentless pattern to genuinely open-ended tasks (plan, design) is wrong. The criterion is: *is the task structure known in advance?*

---

## Cross-cutting synthesis

### Section X: Convergent Design Principles

Seven principles recur across the surveyed literature and should be treated as foundational for moai-adk v3.

1. **Interleaved reasoning-and-acting is the base primitive.** Every surveyed agent harness implements some form of Thought→Action→Observation loop. Supported by: ReAct, Reflexion, Self-Refine, LATS, AutoGen, SWE-agent, Voyager, Devin, CodeAct, Agentless (even the non-agentic baseline has a reasoning-bounded version).

2. **Separate generation from evaluation.** Reflexion's Actor/Evaluator/Self-Reflection split, Tree-of-Thoughts' state-value separation, Agent-as-a-Judge's distinct judge, MetaGPT's QA role, and Constitutional AI's critique-then-revise all converge on this. Moai's manager-* vs evaluator-active separation is structurally correct. Supported by: Reflexion, ToT, LATS, Agent-as-a-Judge, Constitutional AI, MetaGPT, Self-Refine.

3. **Memory is typed, not uniform.** Episodic (trial-by-trial), semantic (facts/skills), procedural (workflows), and reflective (consolidated beliefs) are structurally different. Supported by: Generative Agents, Voyager, AWM, Reflexion. Implication: moai's single `.claude/agent-memory/` directory should be subdivided.

4. **Tools and code are on a spectrum; executable code is the richer end.** Pre-defined JSON tools (ReAct-era) < Python-code actions (CodeAct, Voyager, Chain of Code) < full interpreter sandbox (LATS). The richer action space composes multiple operations in one LLM turn and leverages error messages as observations. Supported by: CodeAct, Voyager, Chain of Code, LLMCompiler.

5. **Interface design is a first-class concern.** The ACI (Agent-Computer Interface) is as important as model choice. Supported by: SWE-agent (primary thesis), Devin's tool-suite design, Voyager's skill-library as interface, AutoGen's message schemas.

6. **Constitutions + self-critique constrain emergent behavior.** Declarative principles + runtime self-critique is cheaper and more predictable than hand-coded guards. Supported by: Constitutional AI, Reflexion (self-reflection as constitution), MetaGPT (SOPs as constitution).

7. **Parallelism requires explicit dependency modeling.** Naive parallel tool calls cause conflicts; compiler-style DAGs prevent them. Supported by: LLMCompiler (primary), ADAS (agent composition), AutoGen (group chat scheduling), moai's own team mode design.

### Section Y: Divergent Design Choices

Where the papers genuinely disagree — moai v3 must pick a side.

| Axis | Option A (papers) | Option B (papers) | Moai implication |
|------|-------------------|-------------------|------------------|
| Agent count | Multi-agent role specialization (MetaGPT, AutoGen, Generative Agents, Devin's MultiDevin) | Single agent with disciplined pipeline (Agentless, Self-Refine, ReAct solo) | Keep multi-agent for open-ended SPEC work (plan/design); collapse to single-agent pipeline for well-structured utility subcommands (fix, coverage, codemaps, mx, clean). |
| Orchestrator vs peer handoffs | Centralized orchestrator decides (MetaGPT, moai current) | Agents self-handoff (OpenAI Swarm, AutoGen group chat) | Keep centralized for auditability; add optional peer-handoff for intra-phase refinement (e.g., expert-backend → expert-security without round-tripping through MoAI). |
| Memory lifetime | Ephemeral per-task (Self-Refine, Agentless) | Persistent across tasks (Voyager, Generative Agents, AWM, moai lessons.md) | Keep persistent for workflow memory / lessons; make episodic memory ephemeral (scope per SPEC). Generative Agents-style periodic consolidation cron. |
| Evaluation frequency | Final output only (SWE-bench original, LLM-as-judge) | Intermediate trajectory (Agent-as-a-Judge, Sprint Contract) | Adopt Agent-as-a-Judge for thorough harness level; allow final-only for minimal harness. |
| Action space | Pre-defined JSON tools (ReAct, Toolformer) | Executable code (CodeAct, Voyager, Chain of Code) | Continue supporting Bash as code-action proxy; consider adding a Python sandbox tool for complex multi-step compositions. |
| Handwritten vs learned harness | Handwritten agents/skills (MetaGPT, moai current) | Meta-learned harness (ADAS, DSPy) | Keep handwritten as primary; add `moai meta-design` experimental mode gated by FROZEN zone + human approval (already compatible with design constitution §5). |
| Search breadth | Deep single trajectory (ReAct, Self-Refine, Agentless) | Tree search with pruning (ToT, LATS) | Default to deep; elevate to tree search only at thorough harness level with explicit user opt-in. |
| Tool-call timing | Sequential (ReAct) | Parallel DAG (LLMCompiler) | Parallel DAG where safe (independent reads), sequential for writes (current moai team-mode file-ownership rule is correct). |

### Section Z: Open research questions for moai v3

1. **Should moai adopt a formal action-DAG for plan-phase output?** Today `/moai plan` produces a sequential SPEC + task list. LLMCompiler and MLE-bench suggest DAG-with-parallelism would reduce total latency and reveal genuine dependencies. Open: does Go's concurrency model plus moai's existing team-mode file-ownership already capture this implicitly? Needs empirical measurement.

2. **Where does skill distillation happen?** Voyager auto-distills skills from successful code; AWM auto-induces workflows. Moai currently relies on human authorship via builder-skill. Open: should there be a `moai skill distill` that promotes high-fan-in @MX:ANCHOR functions into skill bundles? Risk: uncontrolled skill-library growth. Gate: FROZEN zone + graduation protocol (already defined).

3. **Is the evaluator itself an agent or a pipeline?** Agent-as-a-Judge says agent (matches human reliability); Agentless says pipeline (simpler, cheaper). Open: which wins for moai's specific workloads? Likely both — Agent judge for thorough harness, pipeline scorer for minimal/standard.

4. **How persistent is episodic memory?** Reflexion scopes memory per trial; Generative Agents span days; AWM spans workloads. Open: for moai, what's the right unit — per-SPEC (ephemeral), per-project (medium), per-machine (global lessons.md)? Current design is muddled; needs explicit layering.

5. **When does moai invoke "xhigh" vs "max" effort?** Anthropic's guidance: xhigh is the recommended starting point for agentic coding, max only when evals show headroom. Open: moai needs empirical per-agent tuning data. Candidate: instrument all agent invocations with (effort_level, duration, success) tuples, aggregate weekly, auto-recommend.

6. **Can the harness itself be versioned and A/B tested?** ADAS shows harness topologies are learnable. Open: should moai track harness versions (e.g., harness v3.0, v3.1) and run shadow evaluations (design constitution Layer 2 Canary) before promoting changes? This is a v3+ research direction but should be designed in now.

7. **How many agents is too many?** Agent-as-a-Judge hints at a flywheel where judge agents evolve with evaluated agents — but more agents ≠ more quality. Agentless explicitly argues for fewer. Open: what's the right *default* agent count for a given harness level? A data-driven answer requires the benchmarks from Q1 and Q3.

8. **What is the "Sprint Contract" equivalent for plan and sync phases?** GAN loop has Sprint Contracts (design constitution §11.4) in the run phase. Plan and sync don't. Open: should plan produce a "Plan Contract" (explicit acceptance criteria for what a valid plan looks like) and sync a "Release Contract"?

---

## References

- [ReAct — Yao et al. 2022](https://arxiv.org/abs/2210.03629)
- [Reflexion — Shinn et al. 2023](https://arxiv.org/abs/2303.11366)
- [Self-Refine — Madaan et al. 2023](https://arxiv.org/abs/2303.17651)
- [Tree of Thoughts — Yao et al. 2023](https://arxiv.org/abs/2305.10601)
- [LATS — Zhou et al. 2023](https://arxiv.org/abs/2310.04406)
- [AutoGen — Wu et al. 2023](https://arxiv.org/abs/2308.08155)
- [MetaGPT — Hong et al. 2024](https://arxiv.org/abs/2308.00352)
- [Generative Agents — Park et al. 2023](https://arxiv.org/abs/2304.03442)
- [Agent-as-a-Judge — Zhuge et al. 2024](https://arxiv.org/abs/2410.10934)
- [OpenAI Swarm — OpenAI 2024](https://github.com/openai/swarm)
- [SWE-agent — Yang et al. 2024](https://arxiv.org/abs/2405.15793)
- [CodeAct — Wang et al. 2024](https://arxiv.org/abs/2402.01030)
- [Voyager — Wang et al. 2023](https://arxiv.org/abs/2305.16291)
- [Toolformer — Schick et al. 2023](https://arxiv.org/abs/2302.04761)
- [LLMCompiler — Kim et al. 2023](https://arxiv.org/abs/2312.04511)
- [ADAS — Hu et al. 2024](https://arxiv.org/abs/2408.08435)
- [DSPy — Khattab et al. 2023](https://arxiv.org/abs/2310.03714)
- [Constitutional AI — Bai et al. 2022](https://arxiv.org/abs/2212.08073)
- [Claude Opus 4.7 Effort Parameter — Anthropic 2026](https://platform.claude.com/docs/en/build-with-claude/effort)
- [SWE-bench — Jimenez et al. 2024](https://arxiv.org/abs/2310.06770)
- [MLE-bench — Chan et al. 2024](https://arxiv.org/abs/2410.07095)
- [Devin / Cognition — Cognition 2024](https://cognition.ai/blog/introducing-devin)
- [AgentBench — Liu et al. 2023](https://arxiv.org/abs/2308.03688)
- [Agent Workflow Memory — Wang et al. 2024](https://arxiv.org/abs/2409.07429)
- [Agentless — Xia et al. 2024](https://arxiv.org/abs/2407.01489)

Additional related papers consulted but not primary targets:
- [ReWOO — Xu et al. 2023](https://arxiv.org/abs/2305.18323) — decoupling reasoning from observations for token efficiency
- [Chain of Code — Li et al. 2023](https://arxiv.org/abs/2312.04474) — code-as-reasoning with LMulator fallback

## Sources (all URLs verified via WebSearch + direct WebFetch where feasible)

- https://arxiv.org/abs/2210.03629 (ReAct)
- https://arxiv.org/abs/2303.11366 (Reflexion)
- https://arxiv.org/abs/2303.17651 (Self-Refine)
- https://arxiv.org/abs/2305.10601 (Tree of Thoughts)
- https://arxiv.org/abs/2310.04406 (LATS)
- https://arxiv.org/abs/2308.08155 (AutoGen)
- https://arxiv.org/abs/2308.00352 (MetaGPT)
- https://arxiv.org/abs/2304.03442 (Generative Agents)
- https://arxiv.org/abs/2410.10934 (Agent-as-a-Judge)
- https://github.com/openai/swarm (OpenAI Swarm)
- https://arxiv.org/abs/2405.15793 (SWE-agent)
- https://arxiv.org/abs/2402.01030 (CodeAct)
- https://arxiv.org/abs/2305.16291 (Voyager)
- https://arxiv.org/abs/2302.04761 (Toolformer)
- https://arxiv.org/abs/2312.04511 (LLMCompiler)
- https://arxiv.org/abs/2408.08435 (ADAS)
- https://arxiv.org/abs/2310.03714 (DSPy)
- https://arxiv.org/abs/2212.08073 (Constitutional AI)
- https://platform.claude.com/docs/en/build-with-claude/effort (Claude Opus 4.7 Effort)
- https://arxiv.org/abs/2310.06770 (SWE-bench)
- https://arxiv.org/abs/2410.07095 (MLE-bench)
- https://cognition.ai/blog/introducing-devin (Devin)
- https://arxiv.org/abs/2308.03688 (AgentBench)
- https://arxiv.org/abs/2409.07429 (AWM)
- https://arxiv.org/abs/2407.01489 (Agentless)
- https://arxiv.org/abs/2305.18323 (ReWOO, secondary)
- https://arxiv.org/abs/2312.04474 (Chain of Code, secondary)
