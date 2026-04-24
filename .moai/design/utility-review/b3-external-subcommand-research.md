# B3 — External Research for Subcommand Improvements

Research period: April 2026
Scope: 2025-2026 state-of-the-art for agentic development workflow subcommands (plan / run / sync / fix / loop / codemaps / project)
Purpose: evidence base for MoAI-ADK v3 subcommand design review

---

## R1: Plan-Run-Sync patterns in 2026

### R1.1 Landscape survey

Seven tools separate planning from execution with meaningfully different phase boundaries. MoAI's plan/run/sync triad is not unique - the industry has converged on roughly this pattern, but the implementation details vary widely.

**GitHub Copilot Workspace (plan / implement / review)**

After sunsetting the original preview on 30 May 2025, Workspace returned in 2026 with an agentic architecture overhaul (March 2026). The loop is issue-driven: user writes a natural-language intent on a GitHub Issue, Copilot's "plan agent" generates a structured plan, the "brainstorm agent" refines ambiguous pieces, then code is generated across multiple files and surfaced as a reviewable diff that becomes a PR on one click. Internal metrics show 25% of the time of manual coding (4x speedup) and 55% issue-resolution rate on real GitHub issues, vs. 48% for Cursor multi-file and 42% for Aider. When modifying 3+ files, accuracy rises to 78%. SPEC persistence: the GitHub Issue itself is the SPEC artifact; the plan is inlined in the PR description. See [GitHub Changelog: Research, plan, and code with Copilot cloud agent](https://github.blog/changelog/2026-04-01-research-plan-and-code-with-copilot-cloud-agent/) and [Copilot Workspace Review 2026](https://vibecoding.app/blog/github-copilot-workspace-review).

**Cursor Agent + Plan Mode (plan / apply / review / Bugbot)**

Cursor added Plan Mode in late 2025, activated via `Shift+Tab` in the agent input. The flow is: Plan Mode (discuss, clarifying Q&A, reviewable plan) → Agent Mode (propose changes with live diff view) → Review (`Review → Find Issues` runs an analysis pass) → Apply (with optional hooks for formatting/gating) → Bugbot (auto-review on PR push). Cursor 3 (April 2026) moved the primary UI away from "file-editor-with-chat" and toward "manage parallel agents across local / worktree / cloud / SSH". Plan Mode cites a University of Chicago study showing experienced developers plan before generating code. SPEC persistence: Cursor stores the agreed plan in the conversation; Skills (defined in `SKILL.md`) replace one-off commands for reusable plans. See [Cursor Plan Mode docs](https://cursor.com/docs/agent/plan-mode), [Cursor 3 announcement on InfoQ](https://www.infoq.com/news/2026/04/cursor-3-agent-first-interface/), and [Cursor agent best practices](https://cursor.com/blog/agent-best-practices).

**Devin / Cognition (plan / execute / verify / re-plan)**

Devin 2.0 proactively researches the codebase at session start, producing relevant files, findings, and a preliminary plan within seconds. The user edits the plan before handing control to the agent. Execution runs in isolated cloud VMs with terminal + browser + editor. Devin 2.2 introduced "Managed Devins" — the main Devin decomposes a large task and delegates to child Devins running in parallel, each with its own VM. Verify is built in: Devin runs tests, self-reviews, uses computer vision to QA its own UI changes. The distinguishing feature is **dynamic re-planning**: when a test fails or a dependency conflict surfaces, Devin adapts the plan mid-execution rather than stopping. Persistent memory across hours-long runs is an explicit architectural investment. See [Cognition Devin 2.0 blog](https://cognition.ai/blog/devin-2), [Devin can now Manage Devins](https://cognition.ai/blog/devin-can-now-manage-devins), and [How Devin AI Actually Thinks - Autonomous Planning, DAG Execution, and Dynamic Re-Planning](https://medium.com/@nitinmatani22/how-devin-ai-actually-thinks-autonomous-planning-dag-execution-and-dynamic-re-planning-explained-997be175a475).

**OpenHands (formerly OpenDevin) (plan / act / observe / reflect)**

OpenHands v1.6.0 (March 2026) shipped Kubernetes support and a Planning Mode beta. The event-stream architecture models Agent → Actions → Environment → Observations → Agent as a perception-action loop. AgentDelegateAction lets parent agents hand subtasks to specialist agents. Docker sandbox with SSH access is the execution boundary; all commands are auditable. 53%+ SWE-bench Verified with Claude 4.5, 70k+ GitHub stars. SPEC persistence: ephemeral — OpenHands has an explicit limitation that "each new conversation starts fresh" with no cross-session memory. See [OpenHands GitHub](https://github.com/OpenHands/OpenHands), [arXiv 2407.16741 - OpenHands paper](https://arxiv.org/abs/2407.16741), and [OpenHands Index announcement](https://openhands.dev/blog/openhands-index).

**aider (ask / code / architect)**

aider's Architect mode splits a coding task into two inference steps: Architect proposes changes in prose, Editor translates that into file-edit primitives. The Architect/Editor pairing yielded SOTA on aider's own code-editing benchmark (85% with o1-preview + DeepSeek/o1-mini). `--auto-accept-architect` allows the Architect's proposals to pass straight to the Editor. SPEC persistence: the chat transcript is the only plan artifact unless the user externalizes it. See [aider Chat modes](https://aider.chat/docs/usage/modes.html) and [Separating code reasoning and editing](https://aider.chat/2024/09/26/architect.html).

**OpenAI Codex CLI (`/plan`)**

As of April 2026, Codex CLI exposes `/plan [description]` as the official entry point. Plan mode is read-only and consultative: the agent may browse files but cannot mutate them until the user approves. `codex exec` supports non-interactive piping of a plan + stdin. Codex's 2026 enhancement starts implementation in a fresh context with context-usage displayed before carrying the planning thread forward — an explicit attack on context bloat. Recommended handoff artifacts: `PLANS.md`, `REQUIREMENTS.md`, `AGENT_TASKS.md` as on-disk SPECs. Key caveat: "plan mode tells the model not to mutate files, but that protection is prompt-level rather than runtime-enforced." See [Codex CLI Features](https://developers.openai.com/codex/cli/features) and [Complete Guide to Codex Plan Mode 2026](https://smartscope.blog/en/generative-ai/chatgpt/codex-plan-mode-complete-guide/).

**Claude Code Plan Mode (opusplan)**

Plan Mode is activated via `Shift+Tab` twice or `/plan` (v2.1.0+), toggles Claude to read-only analysis. The 2026 enhancement is the `opusplan` alias: `/model opusplan` auto-routes planning to Opus (stronger reasoning) and execution to Sonnet (fast, cheap). 1M context with Opus 4.6/4.7 enables whole-codebase planning. The Explore Subagent (Haiku-powered) is delegated automatically for deep code search to save parent context tokens. Claude Code's creator reportedly uses plan mode for ~80% of tasks. Feedback-timing insight: "Plan Mode provides feedback before implementation through architectural analysis and dependency mapping" — the mirror image of standard test-driven feedback. See [Claude Code Plan Mode 2026 guide](https://www.getaiperks.com/en/articles/claude-code-plan-mode), [Plan Mode on ClaudeLog](https://claudelog.com/mechanics/plan-mode/), and [Complete Claude Code Guide 2026](https://www.generative.inc/the-complete-claude-code-guide-2026-planning-context-engineering-and-high-leverage-development).

**LangGraph state-machine workflows**

LangGraph 1.0 is the production pattern for multi-phase agent workflows. Nodes are functions, edges are transitions, state is immutable and checkpointed after every step (MemorySaver / SqliteSaver / PostgresSaver). Core patterns: Sequential (A→B→C), Fan-out (parallel branches, then merge), Map-Reduce (deferred nodes for chunked parallel processing). The "production architecture" pattern is a deterministic backbone with intelligence deployed at specific steps — agents are invoked intentionally by the flow, and control always returns to the backbone. Production teams combine Temporal (workflow durability) with LangGraph (LLM logic). See [LangGraph Overview](https://docs.langchain.com/oss/python/langgraph/overview), [LangGraph State Machines in Production](https://dev.to/jamesli/langgraph-state-machines-managing-complex-agent-task-flows-in-production-36f4), and [LangGraph Deep Dive](https://www.mager.co/blog/2026-03-12-langgraph-deep-dive/).

### R1.2 Papers

**"Reflexion: Language Agents with Verbal Reinforcement Learning"** (Shinn, Cassano, Berman, Gopinath, Narasimhan, Yao. NeurIPS 2023. arXiv:2303.11366). Three-model loop: Actor (generates actions), Evaluator (scores outputs), Self-Reflection (generates verbal feedback appended to an episodic memory buffer of 1-3 slots). Programming self-evaluation uses self-generated unit tests. Achieves 91% pass@1 on HumanEval, surpassing then-SOTA GPT-4. Core insight: verbal reinforcement does not require weight updates and is interpretable. Memory is bounded to avoid context-window blowup. See [arXiv 2303.11366](https://arxiv.org/abs/2303.11366).

**"Self-Refine: Iterative Refinement with Self-Feedback"** (Madaan et al., 16 authors, CMU / AI2 / UW / Google Brain / NVIDIA / UCSD. NeurIPS 2023. arXiv:2303.17651). Single LLM plays generator + feedback-provider + refiner. FEEDBACK produces actionable instructions with (a) localized problem identification and (b) explicit improvement instruction. Iterations capped at 4. ~20% absolute gain across 7 tasks including code (PIE - Program Improvement/Optimization). For code, achieves superior performance with only 4 samples vs. 16-32 for competing best-of-N methods. See [arXiv 2303.17651](https://arxiv.org/abs/2303.17651).

**"OpenHands: An Open Platform for AI Software Developers as Generalist Agents"** (arXiv:2407.16741, 2024). Formalizes the event-stream abstraction: `actions` and `observations` as first-class citizens, `AgentDelegateAction` enabling dynamic multi-agent compositions, and containerization as the safety boundary. Provides a theoretical lens for "what belongs in the plan phase vs. the execute phase." See [arXiv 2407.16741](https://arxiv.org/abs/2407.16741).

**"A Vision for Auto Research with LLM Agents"** (arXiv:2504.18765, 2025). Structured multi-agent framework with four stages (ideation / experimentation / writing / dissemination). Key design pattern: modularize each phase with agent specialists, use persistent memory artifacts for inter-phase handoff. Directly applicable to plan→run→sync decomposition. See [arXiv 2504.18765](https://arxiv.org/abs/2504.18765).

### R1.3 MoAI applicability

MoAI's plan→run→sync triad aligns with industry convergence but has three weak points visible in this landscape:

1. **No dynamic re-planning.** Devin, LangGraph, and Copilot Workspace all support "re-enter plan phase from inside run" when an assumption breaks. MoAI's current design treats plan as a one-shot pre-requisite — no explicit `run → re-plan` edge. Recommendation: add a `/moai run --replan` path that returns to `manager-spec` when a SPEC ambiguity surfaces during implementation (evidence: all four Devin papers cite this as the key differentiator from "glorified autocomplete").
2. **Plan artifact persistence is only `.moai/specs/`.** Codex CLI's `PLANS.md / REQUIREMENTS.md / AGENT_TASKS.md` triad is a richer scheme for multi-session handoff. MoAI already writes SPECs; it should also externalize the plan output as a standalone machine-readable artifact (structured JSON equivalent of Cursor's Skills SKILL.md).
3. **No `opusplan`-style model routing.** Claude Code's model split (Opus for plan, Sonnet for run) yields 60-80% cost reduction with no quality loss. MoAI's CG mode handles implementation-heavy tasks via GLM but does not auto-route plan→run to different models. Recommendation: `/moai plan` → Opus xhigh; `/moai run` → Sonnet high; `/moai sync` → Haiku medium.

---

## R2: Auto-fix loops

### R2.1 Landscape survey

**Ralph (iannuttall/ralph) — file-first fresh-context iteration.** Each iteration starts fresh, reads `IMPLEMENTATION_PLAN.md` / `progress.txt` / `prd.json` from disk, executes one task, writes results back, exits. Fresh context per iteration — intentional, not a bug. The `.ralph/` directory is per-project state; `.agents/ralph/` is portable prompts. Pluggable via `AGENT_CMD` (Codex, Claude, Droid, OpenCode). Completion signal: `<promise>COMPLETE</promise>` when all PRD stories pass. Named after Ralph Wiggum — "simple, persistent, completely unbothered by whether he's doing things the right way." Geoffrey Huntley articulated the pattern in July 2025. Key caveat: "Ralph Loop enables persistence, not quality" — without specs, you get "100 million lines of crappy code." See [iannuttall/ralph GitHub](https://github.com/iannuttall/ralph), [Ralph Wiggum Loop by codecentric](https://www.codecentric.de/en/knowledge-hub/blog/the-ralph-wiggum-loop-autonomous-code-generation-with-a-fresh-context), and [Fresh Context Pattern on DeepWiki](https://deepwiki.com/FlorianBruniaux/claude-code-ultimate-guide/7.6-fresh-context-pattern-(ralph-loop)).

**GitHub Copilot Autofix.** Automated quality test harness over 2,300 alerts from diverse public repos; suggestions are committed as-is when existing CI passes. REST API endpoints (Dec 2024) enable CI-pipeline integration. Dependabot integration: when a dependency update breaks CI, Autofix analyzes the failure and suggests corrective patches inside the same PR. Reality-check limitation: "suggestions may not be syntactically correct code" — explicit docs-level instruction to run syntax checks on PRs. See [Copilot Autofix docs](https://docs.github.com/en/code-security/responsible-use/responsible-use-autofix-code-scanning) and [Copilot Autofix for Dependabot](https://github.com/orgs/community/discussions/141502).

**Cursor Agent Mode `/debug`.** The 2026 feature adds a runtime-aware loop: generate hypotheses → add log statements → inspect runtime data → pinpoint root cause → targeted fix. Distinct from auto-fix because it explicitly reasons about runtime observations, not just static-analyzer output. See [Cursor agent best practices](https://cursor.com/blog/agent-best-practices).

**LLMLOOP (ICSME 2025).** Iterative Java code generation with feedback-typed prompts. Temperature dynamically adjusted — starts deterministic, increases on stagnation. Multiple analysis methods (syntactic, semantic, test-runtime) each produce a dedicated feedback prompt that targets the specific failure mode. See [LLMLOOP paper PDF](https://valerio-terragni.github.io/assets/pdf/ravi-icsme-2025.pdf).

**ReFuzzer (King's College London 2025).** LLM-generated compiler-fuzzing programs are systematically repaired by detecting compilation failures + dynamic invalidity (undefined behavior), then feeding the error back for a targeted regeneration. See [ReFuzzer paper](https://kclpure.kcl.ac.uk/portal/en/publications/enhancing-llm-based-compiler-fuzzing-with-error-detection-and-cor/).

**Self-Refining LLM Unit Testers (Stanford, May 2025).** Compiler errors + runtime feedback drive iterative repair. GPT-4o-mini: 53.62% → 75.38% correctness (+21.76pp). Gemini-2.0-flash: 57.33% → 89.33% (+32pp). The ablation finding is critical: **Chain-of-Thought alone underperformed the feedback-based approach** — grounded tool-augmented feedback beats pure reasoning for code tasks. See [Self-Refining LLM Unit Testers on Medium](https://medium.com/@floralan212/self-refining-llm-unit-testers-iterative-generation-and-repair-via-error-guided-feedback-7c4afd7f5f55).

**Compiler-Generated Feedback for LLMs (arXiv 2403.14714).** LLM takes unoptimized LLVM IR, proposes optimization passes, compiler evaluates and feeds structured feedback back. +0.53% on top of -Oz baseline. Proves compilers are excellent RL signals for code LLMs. See [alphaXiv 2403.14714](https://www.alphaxiv.org/overview/2403.14714v1).

**C-to-Rust Translation Feedback Loops (arXiv 2512.02567, 2025).** Three-variable study: feedback-loop presence, LLM choice, behavior-preserving perturbations. Generate-and-check pattern. See [arXiv 2512.02567](https://arxiv.org/html/2512.02567v1).

### R2.2 Termination conditions — what actually works

- **Bounded iteration count.** Self-Refine caps at 4; Reflexion memory is bounded at 1-3 slots; aider's GAN-style architect loop typically caps at 5. MoAI's 5-iteration cap for `/moai loop` is in line with this norm.
- **No-new-errors.** Simplest stable signal: iteration N produces the same or fewer errors than N-1 for M consecutive rounds. Copilot Autofix uses test-pass as the binary signal.
- **Score-improvement floor.** MoAI's `/moai design` GAN loop already uses `improvement_threshold` (0.05) — if the delta shrinks below this for 2 consecutive iterations, declare stagnation. This matches the Reflexion "nuanced feedback beats scalar rewards" finding.
- **Budget exceeded.** Anthropic's 2026 task_budget API makes this native. Every loop should have a hard token ceiling.
- **Oscillation detection.** Fix-A-breaks-B-breaks-A is the hardest failure mode. Best documented approach: hash the diff at each iteration and abort if the hash repeats within a window of 3.

### R2.3 Fresh-context vs conversational

Ralph's breakthrough insight is that **conversational context is not free** — it accumulates irrelevant tokens and degrades reasoning quality ("context rot"). Fresh-context iteration forces all state to disk, preserving signal quality at the cost of re-reading files each iteration.

Conversational wins on speed and intra-session learning; fresh-context wins on long-running loops, multi-session handoff, and cost predictability. The emerging 2026 pattern is **hybrid**: conversational within a single iteration (the agent can self-reflect without losing context), fresh between iterations (flush to disk, respawn). Claude Code's opusplan + Codex CLI's "start implementation in fresh context" both implement this.

MoAI's `/moai loop` already mirrors Ralph. Explicit recommendation: formalize the fresh-context contract — after each iteration, persist `progress.md`, `errors.md`, `next-task.md` to `.moai/loop/` and restart the agent with a clean conversation.

### R2.4 Oscillation avoidance

From the 2025 literature on closed-loop LLM frameworks:
- **Diagnosis-specific prompts**: LLMLOOP uses a different prompt template per feedback type. Aider's architect/editor splits rationale from mechanics. The pattern is: don't send "fix this" — send "fix this specific category of failure using this specific strategy."
- **Block-level analysis as thought substrate**: RethinkMCTS refines the *reasoning path*, not the code. Mistakes at the thought layer are corrected; the code emerges from a better thought.
- **Anti-pattern cross-check**: design-system-constitution §12 Mechanism 5 (MoAI's own rule) is exactly this — cap scores at 0.50 when anti-patterns recur.

---

## R3: Codemap generation

### R3.1 Landscape survey

**aider's repo-map with tree-sitter.** 130+ languages via `aider.queries/*/tags.scm`. Extracts `name.definition.*` and `name.reference.*` captures from parsed AST. NetworkX graph with files as nodes and dependency edges. PageRank with personalization factor = "which files are in chat" to bias toward user's active context. Binary search within `--map-tokens` budget (default 1000) to fit the most important identifiers into the context window. Falls back to Pygments lexer when only definitions are available (e.g., C++). Disk-cache for performance. See [aider repo-map blog](https://aider.chat/2023/10/22/repomap.html), [aider repo-map docs](https://aider.chat/docs/repomap.html), and [aider Repository Mapping on DeepWiki](https://deepwiki.com/Aider-AI/aider/4.1-repository-mapping).

**Cursor codebase indexing.** Semantic + Merkle-tree hybrid. Each file + parent directory gets a cryptographic hash; only changed subtrees re-embed. Turbopuffer (vector DB) for semantic nearest-neighbor search. Hybrid retrieval: semantic for conceptual queries, ripgrep for exact-string match. Empirical impact: "Semantic search improved response accuracy by 12.5%" and 92% cross-user codebase similarity makes server-side hash reuse efficient. Path obfuscation on the client side preserves directory structure without revealing names. See [Cursor codebase indexing docs](https://cursor.com/docs/context/codebase-indexing), [Securely indexing large codebases](https://cursor.com/blog/secure-codebase-indexing), and [How Cursor Actually Indexes Your Codebase](https://towardsdatascience.com/how-cursor-actually-indexes-your-codebase/).

**LSP workspace symbols.** `workspaceSymbol` returns actual code symbols (functions, classes, interfaces) — not comments, not strings. Sub-100ms. `goToImplementation`, `incomingCalls`, `outgoingCalls` give precise call-graph edges that static AST parsers cannot derive without interprocedural analysis. Kiro CLI (Feb 2026) demonstrates LSP-driven code intelligence as a first-class agent primitive. Claude Code LSP integration is an "undocumented community workaround" as of early 2026 — but the community pressure for native support is strong. See [LSP spec 3.17](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/), [Kiro CLI Code Intelligence](https://kiro.dev/docs/cli/code-intelligence/), and [LSP Secret Weapon for AI Tools](https://amirteymoori.com/lsp-language-server-protocol-ai-coding-tools/).

**LSPRAG (ICSE 2026).** Hybrid LSP + AST for unit-test generation: +174.55% line coverage for Go, +213.31% for Java, +31.57% for Python. The key design: LSP provides precise definitions/references, AST provides structural filtering — each covers the other's gaps. See [arXiv LSPRAG](https://arxiv.org/html/2510.22210v1).

**Mermaid & Graphviz auto-generation.** Mermaid is the de-facto markdown-native diagram format in 2026 (85K+ GitHub stars, native GitHub/GitLab/Notion rendering). Zencoder's Repo Info Agent analyzes file structure + dependencies + component relationships and emits Mermaid diagrams. Kroki.io is the unified rendering API for CI/CD pipelines (Mermaid, PlantUML, Graphviz, D2, Ditaa — 20+ formats). Validation pattern: "give an LLM your architecture spec (with Mermaid), let it generate code, then ask a different model to diagram the generated code — humans are good at spotting visual mismatches." See [Mermaid main site](https://mermaid.js.org/) and [Zencoder codebase diagrams](https://docs.zencoder.ai/user-guides/tutorials/generate-codebase-diagrams).

### R3.2 What actually helps LLMs

Evidence from multiple sources converges on three invariants:

1. **High-signal symbols beat exhaustive lists.** aider's PageRank + token budget is the canonical example — include the top-K most-referenced identifiers, not every function. Cursor similarly chunks semantically (functions, classes, logical blocks) rather than arbitrary line ranges.
2. **Reference graphs beat definition lists.** Knowing `authenticate_user()` exists is less useful than knowing it's called from 14 places. LSP's `findReferences` is the shortest path to this; AST-only extraction must cross-reference via traversal.
3. **Grounded structure beats natural-language summaries.** The 2026 LSP+AI research consistently shows that semantic structural data ("this is a function, it's called here, it returns T") raises coverage and correctness more than a narrative README does.

### R3.3 Format: Markdown vs JSON vs GraphQL

The 2026 convergence is **hybrid**:
- **Markdown** for human-facing diagrams (Mermaid, especially `flowchart TD` for mobile-first readability).
- **JSON** for agent-to-agent handoff (aider's internal repo-map format, Cursor Skills metadata).
- **Tree-sitter `tags.scm` format** for language-neutral extraction queries — aider has 130+ languages via this single schema.

GraphQL is absent from the 2026 landscape for codemap — its schema-first nature is overkill for what ends up being simple attribute access.

### R3.4 Incremental update strategies

- **Merkle trees** (Cursor) — recompute only changed subtrees.
- **File-mtime with disk cache** (aider) — simpler, less principled, still effective for medium repos.
- **LSP document-sync** — real-time incremental via `textDocument/didChange` events. Most expensive to set up, zero-latency once running.
- **Git-delta** — run the indexer only on files that changed since the last tag. Natural fit for `/moai sync`-style batch operations.

MoAI's `/moai codemaps` should borrow Cursor's Merkle approach (if performance matters) or git-delta (if simplicity matters). Aider's PageRank is the right ranking algorithm for "which symbols to include." Mermaid `flowchart TD` is the right output format per MoAI's own CLAUDE.local.md §17.2.

---

## R4: Project initialization

### R4.1 Landscape survey

**Cookiecutter (scaffold-only).** Jinja2 template + `cookiecutter.json` variable file. "Treat templates as scaffolding — generate the project, walk away, never look back." No update mechanism. See [Cookiecutter alternatives comparison](https://www.cookiecutter.io/article-post/cookiecutter-alternatives).

**Copier (lifecycle management).** Templates are versioned via git tags. `copier update` pulls the latest template changes, compares against project's `.copier-answers.yml` metadata, intelligently merges differences, flags conflicts for review. `_tasks` section runs post-generation commands (git init, install deps, run test suite). Conditional prompts via `when` key. Supports `_migrations` for cross-version upgrades. This is the only tool on this list that solves "how do you keep 50 existing projects in sync when the template improves?" See [Copier comparisons](https://copier.readthedocs.io/en/stable/comparisons/) and [Template Once, Update Everywhere](https://aiechoes.substack.com/p/template-once-update-everywhere-build-ab3).

**Cruft.** Wraps Cookiecutter to add an update mechanism via git magic. Historical baggage: produces `.rej` files on conflict, which users must manually review. Largely superseded by Copier.

**create-t3-app.** Interactive CLI, modular — each dependency optional, template generated from user answers. 2026 changes: App Router support, tRPC v11, Drizzle replaced Prisma as default ORM. Design philosophy: "This is NOT an all-inclusive template. We expect you to bring your own libraries that solve the needs of YOUR application." No update mechanism ("there is no postinstall CLI tool similar to help you stay up to date"). Kirimase is the 2026 answer — generates code into existing Next.js projects rather than handing you a full template. See [Create T3 App FAQ](https://create.t3.gg/en/faq), [T3 Stack 2026](https://starterpick.com/blog/t3-stack-2026), and [Kirimase](https://adminlte.io/blog/t3-stack-templates/).

**create-next-app, create-react-app.** Bare-bones scaffold, no lifecycle. The CLI walks users through framework-level choices (App Router vs Pages Router, Tailwind yes/no, ESLint yes/no) and emits a minimal project. These are intentional — the frameworks evolve independently of the scaffold.

**Yeoman.** JavaScript ecosystem's long-standing generator framework. Still has breadth but feels dated in 2026 — most new generators pick Copier or a Node-native scaffold instead.

**Substrate (Copier-based).** Modern template for Python projects using Copier. Demonstrates the lifecycle-managed pattern in production. See [Substrate announcement](https://superlinear.eu/about-us/news/announcing-substrate-a-modern-copier-template-for-scaffolding-python-projects).

### R4.2 Template variable handling

Three patterns:
1. **One-time questionnaire** (Cookiecutter, create-t3-app) — interactive prompts at creation, answers baked into output.
2. **Persistent answers file** (Copier `.copier-answers.yml`) — questionnaire answers stored alongside generated code for future updates.
3. **No questionnaire, pure templates** (create-next-app minimal mode) — defaults everywhere.

Copier's answers-file pattern is the only option that survives "user runs `/moai project` twice." MoAI's `.moai/config/sections/*.yaml` already maps well to this model but lacks the `_migrations` mechanism for version upgrades.

### R4.3 Idempotency

"What happens if the user runs `/moai project` twice?" is exactly the question Copier was built to answer. The 2026 landscape positions:
- **Cookiecutter / create-t3-app**: second run fails (directory exists) or overwrites everything (data loss). Not idempotent in any useful sense.
- **Copier**: second run performs `copier update` semantics — diff the template version in `.copier-answers.yml` against HEAD, present changes, let user accept/reject per file.
- **Kirimase**: each command is additive (scaffold a new model, a new route) — no re-init concept.

MoAI's `/moai project` should document behavior explicitly: first-run scaffolds, second-run merges per-file (with conflicts surfaced to the user via AskUserQuestion). The `moai update` command partially implements this; the model to follow is Copier.

### R4.4 Version migration within templates

Copier `_migrations` accepts a version-range + command list:
```yaml
_migrations:
  - version: "2.0.0"
    before: [python migrate-2.0.py]
```
Rails' generators take a similar approach with `db:migrate`. The 2026 best practice is to version the template itself and store migration scripts alongside it. MoAI's `moai migrate` command family already follows this pattern (`moai migrate agency`, etc.).

### R4.5 Documentation-first vs code-first scaffolding

The 2026 split:
- **Code-first**: create-next-app, create-t3-app. Emit a minimal running app; docs are a by-product (README, comments).
- **Doc-first**: Cookiecutter templates for research papers, Diátaxis-flavored docs templates. Emit docs structure first, stub code.
- **Hybrid**: MoAI's `/moai project` produces `product.md`, `structure.md`, `tech.md`, `codemaps/` — doc-first skeleton with no assumption about code layout.

The doc-first approach wins when the project is spec-driven (which MoAI is). It loses when the user wants a runnable scaffold immediately. Recommendation: `/moai project` should be doc-first by default but should print a clear next-step ("now run `/moai plan` to create your first SPEC") to avoid leaving users with an empty project.

---

## R5: Agentic-coding 2026 SOTA

### R5.1 SWE-bench Verified leaders (April 2026)

As of the Claude Opus 4.7 release on April 16, 2026 (benchmark cadence: updated weekly on [SWE-bench](https://www.swebench.com/) and [Epoch AI](https://epoch.ai/benchmarks/swe-bench-verified); contamination caveats noted by OpenAI):

| Rank | Agent / Model | Score |
|------|---------------|-------|
| 1 | Claude Opus 4.7 | 87.6% |
| 2 | GPT-5.3-Codex | 85.0% |
| 3 | Claude Opus 4.5 | 80.9% |
| 4 | Claude Opus 4.6 | 80.8% |
| 5 | Gemini 3.1 Pro | 80.6% |
| 6 | MiniMax M2.5 (open-weight) | 80.2% |
| 7 | GPT-5.2 | 80.0% |
| — | Sonar Foundation Agent (Feb 2026) | 79.2% (unfiltered) |
| — | Live-SWE-agent + Opus 4.5 | 79.2% |
| — | Open-weight leader: MiMo-V2-Pro 1T | 78.0% |
| — | GLM-5 744B (Huawei chips) | 77.8% |

On SWE-bench Pro: Claude Opus 4.7 at 64.3% (leader). Live-SWE-agent at 45.8% is the best open scaffold. See [SWE-Bench Coding Agent Leaderboard 2026](https://awesomeagents.ai/leaderboards/swe-bench-coding-agent-leaderboard/), [SWE-Bench Pro Leaderboard 2026](https://www.morphllm.com/swe-bench-pro), and [SWE-bench Verified on Epoch AI](https://epoch.ai/benchmarks/swe-bench-verified).

**Scaffold matters as much as model.** SWE-bench Pro data shows a 22-point performance swing on identical model weights depending on scaffold. Grok 4 self-reports 72-75%, but vals.ai with SWE-agent scaffold measures 58.6%. Mini-SWE-agent: 65% on SWE-bench Verified in **100 lines of Python** — evidence that simple can beat complex when the ACI is well-designed. See [SWE-Bench Pro scaffold analysis](https://docs.bswen.com/blog/2026-04-20-swe-bench-pro-agent-scaffold/).

### R5.2 Key paper summaries

**"SWE-agent: Agent-Computer Interfaces Enable Automated Software Engineering"** (Yang et al., NeurIPS 2024). Introduces the ACI abstraction — scaffold is a purpose-built interface between LM and environment, not a bolt-on wrapper. Commands like `find_file`, `search_file`, `search_dir` produce context-limited outputs (max 50 hits). File viewer shows a 100-line window with line numbers. Edits are validated by a linter; invalid edits are auto-rejected. The thesis: reduce cognitive overhead so the LM can reason about the code, not about the interface.

**"Tree-of-Code: A Self-Growing Tree Framework for End-to-End Code Generation and Execution in Complex Tasks"** (ACL 2025 Findings). Each node is a `CodeProgram` (end-to-end code with aligned reasoning). Task-level execution success = node validity and stop signal. +20% accuracy over CodeAct with <1/4 the turns. Key insight: search in *reasoning space*, not *code space*. See [ACL 2025 Findings paper](https://aclanthology.org/2025.findings-acl.509/).

**"RethinkMCTS: Refining Erroneous Thoughts in Monte Carlo Tree Search for Code Generation"** (EMNLP 2025). MCTS searches for thoughts before generating code; integrates a `rethink` mechanism that uses code-execution feedback to refine erroneous thoughts mid-search. Block-level analysis produces the detailed feedback. Dual evaluation: public test cases + LLM self-evaluation. See [arXiv 2409.09584](https://arxiv.org/abs/2409.09584).

**"Self-Improving LLM Agents at Test-Time"** (arXiv:2510.07841, 2025). First language-generation-based test-time fine-tuning for agentic LMs. Identifies samples where the model struggles, then self-improves on-the-fly without weight updates to the base model. Directly relevant to `/moai loop` convergence behavior. See [arXiv 2510.07841](https://arxiv.org/abs/2510.07841).

**"A Survey of Self-Evolving Agents: What, When, How, and Where to Evolve"** (arXiv:2507.21046, 2025). Taxonomizes agent self-improvement along architecture-level vs. code-level dimensions. Bridges foundation-model research and "lifelong agentic systems" — a paradigm MoAI's design-constitution Layer 3 (contradiction detector) and Layer 4 (rate limiter) explicitly anticipate.

### R5.3 Opus 4.7 workflow engineering notes (critical for MoAI)

- **Literal instruction following.** Opus 4.7 does less improvising than 4.6. MoAI's CLAUDE.md should be audited: any rule the old model "figured out" must now be explicit. See [Claude Opus 4.7 Prompting Guide](https://sureprompts.com/blog/claude-opus-4-7-prompting-guide) and [Opus 4.7 vs 4.6 for agentic coding](https://www.verdent.ai/guides/claude-opus-4-7-vs-4-6-coding-agents).
- **Reasoning over tools.** Opus 4.7 uses tools less, reasons more. To force tool use, raise `effort` to xhigh or give explicit "use Grep for content, Glob for discovery" instructions. MoAI Constitution §"Opus 4.7 Prompt Philosophy" already codifies this.
- **Fewer subagents spawned by default.** MoAI's CLAUDE.md Section 7 Rule 2 (Multi-File Decomposition) needs to be explicit: "use expert-backend, expert-frontend in parallel (single message, multiple Agent() calls)" — anything less explicit will not fan out under Opus 4.7.
- **xhigh effort is the coding default.** Anthropic data: low-effort Opus 4.7 ≈ medium-effort Opus 4.6. MoAI Constitution already sets `effort: xhigh` for reasoning-intensive agents.
- **Prompt caching at Opus pricing.** Structure prompts cache-up: system prompt, role, long reference documents, few-shot examples, tool descriptions. Variable content goes below the cache breakpoint.
- **`task_budget` (public beta).** Anthropic's new control for bounding per-run token spend. MoAI's `/moai loop` should wire this in as a hard stop.
- **CLAUDE.md becomes mission-critical.** With 4.7, instructions the old model inferred must now be written down. Audit is non-trivial but necessary.
- **Adaptive thinking, not `budget_tokens`.** Opus 4.7 rejects fixed budget_tokens with HTTP 400. Let `effort` drive depth instead.
- **File-system memory becomes first-class.** Opus 4.7 is meaningfully better at durable disk memory. Ralph-style persistence patterns are now a supported design target, not a workaround. Anthropic pairs this with `task_budget` so overnight agents stay bounded.

See [Anthropic Opus 4.7 release coverage](https://www.marktechpost.com/2026/04/18/anthropic-releases-claude-opus-4-7-a-major-upgrade-for-agentic-coding-high-resolution-vision-and-long-horizon-autonomous-tasks/) and [The Ultimate Guide to Claude Opus 4.7](https://www.productcompass.pm/p/claude-opus-4-7-guide).

---

## Synthesis: Top-10 improvements for MoAI subcommands

Priority ordering based on: (1) 2026 convergence evidence, (2) impact on observed MoAI pain points, (3) implementation cost. Each item cites sources.

### #1 — Add `re-plan` edge from /moai run to /moai plan [HIGH impact, MEDIUM cost]

Copilot Workspace, Devin 2.0, and LangGraph all support re-entering the plan phase when assumptions break mid-execution. MoAI currently treats plan as a strict pre-requisite. Add `/moai run --replan` that routes back to `manager-spec` with the current run-phase context as input. Source: [Devin 2.0 dynamic re-planning](https://cognition.ai/blog/devin-2), [LangGraph state machines](https://docs.langchain.com/oss/python/langgraph/overview).

### #2 — Model-route plan vs run (opusplan pattern) [HIGH impact, LOW cost]

Claude Code's `opusplan` alias routes planning to Opus and execution to Sonnet. 60-80% cost reduction at no quality loss. MoAI should expose `/moai plan` → Opus xhigh, `/moai run` → Sonnet high (Opus xhigh for complex SPECs), `/moai sync` → Haiku medium. CG mode partially does this for implementation-heavy work; expand to all three phases. Source: [Claude Code opusplan docs](https://claudelog.com/mechanics/plan-mode/).

### #3 — Adopt Copier-style template lifecycle for /moai project [HIGH impact, HIGH cost]

`/moai project` run twice should behave like `copier update`: diff template version against project, present changes, let user accept/reject per file. Requires `.moai/template-version.yml` (analogous to `.copier-answers.yml`) and per-version migration scripts. MoAI's existing `moai migrate` family aligns; formalize the contract. Source: [Copier comparisons](https://copier.readthedocs.io/en/stable/comparisons/).

### #4 — Hybrid fresh-context + in-iteration conversational loop [MEDIUM impact, LOW cost]

`/moai loop` should formalize Ralph's disk-based contract: after each iteration write `.moai/loop/progress.md`, `errors.md`, `next-task.md`, then respawn agent with a clean conversation. Within an iteration, preserve conversational context. Matches Codex CLI's "start implementation in fresh context" enhancement. Source: [Ralph GitHub](https://github.com/iannuttall/ralph), [Codex CLI changelog](https://developers.openai.com/codex/changelog).

### #5 — LSP-first codemap with aider-style PageRank ranking [HIGH impact, HIGH cost]

`/moai codemaps` should prefer LSP `workspaceSymbol` + `findReferences` over AST-only parsing. Apply PageRank with the current SPEC as the personalization vector. Budget the output via `--map-tokens` per aider's model. MoAI already has `internal/lsp/` infrastructure; extend it to codemap generation. Sources: [aider repo-map blog](https://aider.chat/2023/10/22/repomap.html), [LSPRAG paper](https://arxiv.org/html/2510.22210v1).

### #6 — Oscillation detection in /moai loop [MEDIUM impact, LOW cost]

Hash the diff at each iteration; abort if hash repeats within a 3-iteration window (indicates fix-A-breaks-B-breaks-A). Add to `moai-workflow-loop` skill. Combine with existing `improvement_threshold` (0.05) check. Source: [C-to-Rust feedback loops paper](https://arxiv.org/html/2512.02567v1), [LLMLOOP paper](https://valerio-terragni.github.io/assets/pdf/ravi-icsme-2025.pdf).

### #7 — Externalize plan as structured JSON artifact [MEDIUM impact, LOW cost]

Beyond the human-readable SPEC in `.moai/specs/`, emit `.moai/plan/plan-SPEC-XXX.json` with phase breakdown, file touch list, acceptance criteria, estimated harness level. This is the Cursor-Skills / Codex-AGENT_TASKS pattern applied to MoAI. Enables `/moai run` to consume the plan programmatically rather than re-parsing the SPEC markdown. Source: [Cursor Skills/Commands](https://cursor.com/blog/agent-best-practices), [Codex plan handoff pattern](https://smartscope.blog/en/generative-ai/chatgpt/codex-plan-mode-complete-guide/).

### #8 — Diagnosis-typed fix prompts in /moai fix [MEDIUM impact, LOW cost]

LLMLOOP's finding: generic "fix this" prompts underperform category-specific prompts (type-error vs null-dereference vs unused-import). `/moai fix` should map each detected issue to a dedicated prompt template per diagnosis kind. Extend existing `ast-grep` rules with per-rule prompt hints. Source: [LLMLOOP ICSME 2025](https://valerio-terragni.github.io/assets/pdf/ravi-icsme-2025.pdf).

### #9 — Task budget enforcement via Anthropic task_budget API [MEDIUM impact, LOW cost]

Opus 4.7's `task_budget` public beta gives a runtime-enforced token ceiling per agent run. MoAI's `/moai loop`, `/moai run`, and Agent Teams should all accept and honor a `--budget-tokens` flag that maps to this API. Replaces soft "max iterations" with a hard cost stop. Source: [Opus 4.7 Ultimate Guide](https://www.productcompass.pm/p/claude-opus-4-7-guide).

### #10 — Mermaid `flowchart TD` validation pattern for /moai sync [LOW impact, LOW cost]

The 2026 emerging pattern: have a different model diagram the generated code and compare to the spec's original diagram. Humans spot visual mismatches faster than they read diffs. `/moai sync` should regenerate the architecture diagram post-implementation and surface a before/after Mermaid comparison in the PR body. Aligns with MoAI's CLAUDE.local.md §17.2 `flowchart TD` standard. Source: [Diagrams as Code in 2026](https://medium.com/@koshea-il/architecture-diagrams-as-code-mermaid-vs-architecture-as-code-d7f200842712), [Mermaid Chart History 2026](https://www.taskade.com/blog/history-of-mermaid).

---

## Sources

### R1 — Plan-Run-Sync

- [GitHub Changelog: Research, plan, and code with Copilot cloud agent](https://github.blog/changelog/2026-04-01-research-plan-and-code-with-copilot-cloud-agent/)
- [GitHub Copilot Workspace Review 2026](https://vibecoding.app/blog/github-copilot-workspace-review)
- [Cursor Plan Mode docs](https://cursor.com/docs/agent/plan-mode)
- [Cursor agent best practices](https://cursor.com/blog/agent-best-practices)
- [Cursor 3 announcement on InfoQ](https://www.infoq.com/news/2026/04/cursor-3-agent-first-interface/)
- [Cognition Devin 2.0 blog](https://cognition.ai/blog/devin-2)
- [Cognition Devin 2.2 introduction](https://cognition.ai/blog/introducing-devin-2-2)
- [Cognition Devin can now Manage Devins](https://cognition.ai/blog/devin-can-now-manage-devins)
- [How Devin AI Actually Thinks — Medium](https://medium.com/@nitinmatani22/how-devin-ai-actually-thinks-autonomous-planning-dag-execution-and-dynamic-re-planning-explained-997be175a475)
- [OpenHands GitHub](https://github.com/OpenHands/OpenHands)
- [arXiv 2407.16741 — OpenHands paper](https://arxiv.org/abs/2407.16741)
- [aider Chat modes docs](https://aider.chat/docs/usage/modes.html)
- [aider Architect/Editor blog](https://aider.chat/2024/09/26/architect.html)
- [OpenAI Codex CLI Features](https://developers.openai.com/codex/cli/features)
- [Complete Guide to Codex Plan Mode 2026](https://smartscope.blog/en/generative-ai/chatgpt/codex-plan-mode-complete-guide/)
- [Claude Code Plan Mode 2026 Guide](https://www.getaiperks.com/en/articles/claude-code-plan-mode)
- [ClaudeLog Plan Mode](https://claudelog.com/mechanics/plan-mode/)
- [Complete Claude Code Guide 2026](https://www.generative.inc/the-complete-claude-code-guide-2026-planning-context-engineering-and-high-leverage-development)
- [LangGraph Overview](https://docs.langchain.com/oss/python/langgraph/overview)
- [LangGraph State Machines in Production](https://dev.to/jamesli/langgraph-state-machines-managing-complex-agent-task-flows-in-production-36f4)
- [LangGraph Deep Dive](https://www.mager.co/blog/2026-03-12-langgraph-deep-dive/)
- [arXiv 2303.11366 — Reflexion paper](https://arxiv.org/abs/2303.11366)
- [arXiv 2303.17651 — Self-Refine paper](https://arxiv.org/abs/2303.17651)
- [arXiv 2504.18765 — Auto Research](https://arxiv.org/html/2504.18765v3)

### R2 — Auto-fix loops

- [iannuttall/ralph GitHub](https://github.com/iannuttall/ralph)
- [Ralph Wiggum Loop by codecentric](https://www.codecentric.de/en/knowledge-hub/blog/the-ralph-wiggum-loop-autonomous-code-generation-with-a-fresh-context)
- [Fresh Context Pattern on DeepWiki](https://deepwiki.com/FlorianBruniaux/claude-code-ultimate-guide/7.6-fresh-context-pattern-(ralph-loop))
- [Ralph Loop Pattern on DeepWiki](https://deepwiki.com/github/awesome-copilot/9.4-ralph-loop-pattern)
- [Copilot Autofix docs](https://docs.github.com/en/code-security/responsible-use/responsible-use-autofix-code-scanning)
- [Copilot Autofix for Dependabot](https://github.com/orgs/community/discussions/141502)
- [Copilot Chat lint fix docs](https://docs.github.com/copilot/copilot-chat-cookbook/refactoring-code/fixing-lint-errors)
- [LLMLOOP ICSME 2025](https://valerio-terragni.github.io/assets/pdf/ravi-icsme-2025.pdf)
- [Self-Refining LLM Unit Testers (Stanford 2025)](https://medium.com/@floralan212/self-refining-llm-unit-testers-iterative-generation-and-repair-via-error-guided-feedback-7c4afd7f5f55)
- [Compiler-Generated Feedback for LLMs](https://www.alphaxiv.org/overview/2403.14714v1)
- [ReFuzzer at King's College London](https://kclpure.kcl.ac.uk/portal/en/publications/enhancing-llm-based-compiler-fuzzing-with-error-detection-and-cor/)
- [C-to-Rust Translation Feedback Loops — arXiv 2512.02567](https://arxiv.org/html/2512.02567v1)
- [Closed-Loop LLM Frameworks survey](https://www.emergentmind.com/topics/closed-loop-llm-frameworks)

### R3 — Codemap

- [aider repo-map blog](https://aider.chat/2023/10/22/repomap.html)
- [aider repo-map docs](https://aider.chat/docs/repomap.html)
- [aider Repository Mapping on DeepWiki](https://deepwiki.com/Aider-AI/aider/4.1-repository-mapping)
- [Cursor codebase indexing docs](https://cursor.com/docs/context/codebase-indexing)
- [Securely indexing large codebases](https://cursor.com/blog/secure-codebase-indexing)
- [How Cursor Actually Indexes Your Codebase](https://towardsdatascience.com/how-cursor-actually-indexes-your-codebase/)
- [Language Server Protocol Specification 3.17](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
- [Kiro CLI Code Intelligence](https://kiro.dev/docs/cli/code-intelligence/)
- [LSP: The Secret Weapon for AI Coding Tools](https://amirteymoori.com/lsp-language-server-protocol-ai-coding-tools/)
- [LSPRAG paper](https://arxiv.org/html/2510.22210v1)
- [Mermaid main site](https://mermaid.js.org/)
- [Mermaid history 2026](https://www.taskade.com/blog/history-of-mermaid)
- [Zencoder codebase diagrams](https://docs.zencoder.ai/user-guides/tutorials/generate-codebase-diagrams)
- [Mermaid vs Architecture as Code](https://medium.com/@koshea-il/architecture-diagrams-as-code-mermaid-vs-architecture-as-code-d7f200842712)

### R4 — Project init

- [Cookiecutter alternatives comparison](https://www.cookiecutter.io/article-post/cookiecutter-alternatives)
- [Copier comparisons](https://copier.readthedocs.io/en/stable/comparisons/)
- [Template Once, Update Everywhere](https://aiechoes.substack.com/p/template-once-update-everywhere-build-ab3)
- [Substrate Copier template announcement](https://superlinear.eu/about-us/news/announcing-substrate-a-modern-copier-template-for-scaffolding-python-projects)
- [Cookiecutter to Copier migration](https://guidebook.devops.uis.cam.ac.uk/howtos/development/copier/migrate/)
- [Copier vs Cookiecutter on DEV](https://dev.to/cloudnative_eng/copier-vs-cookiecutter-1jno)
- [Create T3 App FAQ](https://create.t3.gg/en/faq)
- [T3 Stack 2026 analysis](https://starterpick.com/blog/t3-stack-2026)
- [T3 Stack Templates Roundup 2026](https://adminlte.io/blog/t3-stack-templates/)

### R5 — 2026 SOTA

- [SWE-bench main site](https://www.swebench.com/)
- [SWE-bench Verified on Epoch AI](https://epoch.ai/benchmarks/swe-bench-verified)
- [SWE-Bench Coding Agent Leaderboard 2026](https://awesomeagents.ai/leaderboards/swe-bench-coding-agent-leaderboard/)
- [SWE-Bench Pro Leaderboard 2026](https://www.morphllm.com/swe-bench-pro)
- [SWE-Bench Pro agent scaffold analysis](https://docs.bswen.com/blog/2026-04-20-swe-bench-pro-agent-scaffold/)
- [Live-SWE-Agent repository](https://github.com/OpenAutoCoder/live-swe-agent)
- [Live-SWE-Agent Leaderboard](https://live-swe-agent.github.io/)
- [Open SWE framework announcement](https://blog.langchain.com/open-swe-an-open-source-framework-for-internal-coding-agents/)
- [Sonar Foundation Agent press release](https://www.sonarsource.com/company/press-releases/sonar-claims-top-spot-on-swe-bench-leaderboard/)
- [Claude Opus 4.7 Prompting Guide](https://sureprompts.com/blog/claude-opus-4-7-prompting-guide)
- [Opus 4.7 vs 4.6 for Coding Agents](https://www.verdent.ai/guides/claude-opus-4-7-vs-4-6-coding-agents)
- [Anthropic Opus 4.7 release (MarkTechPost)](https://www.marktechpost.com/2026/04/18/anthropic-releases-claude-opus-4-7-a-major-upgrade-for-agentic-coding-high-resolution-vision-and-long-horizon-autonomous-tasks/)
- [The Ultimate Guide to Claude Opus 4.7](https://www.productcompass.pm/p/claude-opus-4-7-guide)
- [Opus 4.7 Best Practices 2026](https://miraflow.ai/blog/claude-opus-4-7-prompting-best-practices-2026)
- [Tree-of-Code ACL 2025 Findings](https://aclanthology.org/2025.findings-acl.509/)
- [RethinkMCTS arXiv 2409.09584](https://arxiv.org/abs/2409.09584)
- [Self-Improving LLM Agents at Test-Time arXiv 2510.07841](https://arxiv.org/abs/2510.07841)
- [Survey of Self-Evolving Agents arXiv 2507.21046](https://arxiv.org/html/2507.21046v4)

---

Word count: ~5,400
