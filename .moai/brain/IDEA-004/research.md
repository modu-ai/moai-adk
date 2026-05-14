# Research: CLI-free self-evolving harness v2 for MoAI-ADK
*Phase 3 — Brain Workflow | Date: 2026-05-14 | Idea: IDEA-004*

## Executive Summary

- Industry reference `revfactory/harness` ships a **6-phase** workflow (Domain Analysis → Team Architecture Design → Agent Generation → Skill Generation → Integration → Validation) — not 7-phase as initial framing suggested. Empirical A/B study reports +60% quality improvement on 15 SE tasks, scaling with complexity (Basic +23.8 → Expert +36.2).
- Reflexion (Shinn et al., NeurIPS 2023, arXiv:2303.11366) supplies the canonical **verbal reinforcement** pattern: Actor + Evaluator + Self-Reflection trio with short-term trajectory memory and long-term reflective text. Achieves +8% absolute boost over episodic memory alone; remains foundational reference in 2026 (e.g., StraTA hierarchical extension).
- Voyager (Wang et al., arXiv:2305.16291) demonstrates **automatic skill library** with three components: (1) auto-curriculum, (2) embedding-indexed executable skill repository, (3) iterative prompting with self-verification. 3.3× more unique items, 15.3× faster milestone unlock vs ReAct/Reflexion/AutoGPT baselines.
- Constitutional AI / RLAIF (Bai et al., Anthropic, arXiv:2212.08073) provides the **principle-based self-critique** template: SL phase (self-critique + revise) → RL phase (AI-feedback preference scoring). Direct analog to MoAI's design constitution acting as the immutable rubric for harness evolution proposals.
- LangGraph Reflection Pattern (2026 production standard) formalizes the cycle as Generator → Evaluator/Critic → Router → Output, with the empirical finding that **diminishing returns kick in after 2-3 iterations** for code/summarization tasks. Token budget tracking and hard iteration caps are mandatory production practice.
- Anthropic official Claude Code Skills/Agents docs (2026) confirm: skills are **folders not just markdown**, with **progressive disclosure** (frontmatter metadata → SKILL.md body → reference files) as the core design principle, and **"Gotchas" sections** as the highest-value content. Commands and skills have merged into a unified `.claude/skills/` namespace.
- AutoGen has entered **maintenance mode** in 2026; Microsoft Agent Framework (MAF) is the successor combining AutoGen orchestration + Semantic Kernel enterprise features. Self-improvement remains a research vision rather than a shipped feature, validating MoAI's opportunity to lead on this dimension.

## Market Landscape

The agent-harness space in 2026 has consolidated around three architectural families:

- **Team-architecture factories** (revfactory/harness, ai-boost/awesome-harness-engineering): Generate `.claude/agents/` + `.claude/skills/` directories from domain analysis. 6 patterns: Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical Delegation. MoAI's current Pipeline pattern aligns with this family.
- **Conversational multi-agent** (AutoGen → MAF, CrewAI): Model interaction as natural-language dialog. Self-improvement is a research direction, not a shipped capability.
- **Graph orchestration** (LangGraph): Explicit state machines with cyclic edges enable reflection loops as a first-class primitive. The `langgraph-reflection` prebuilt graph (main agent + critique subagent + conditional router) is the canonical implementation.

Sources:
- [revfactory/harness](https://github.com/revfactory/harness): Pioneer 6-phase team-architecture factory; provides the cookbook MoAI's harness skills already absorbed.
- [revfactory/claude-code-harness](https://github.com/revfactory/claude-code-harness): Sister repo with empirical +60% quality A/B evidence.
- [Harness website](https://revfactory.github.io/harness/): Public-facing documentation of the 6 architectural patterns.
- [ai-boost/awesome-harness-engineering](https://github.com/ai-boost/awesome-harness-engineering): Community curation of harness patterns, evals, memory, MCP, permissions, observability.

## Academic Foundations (Self-Improvement)

Three foundational papers form the theoretical scaffolding for harness self-evolution:

**Reflexion (Shinn et al., NeurIPS 2023)**: Three-model architecture (Actor Ma, Evaluator Me, Self-Reflection Mr). Verbal reinforcement via episodic reflective text replaces weight updates. Critical for MoAI v2: maps directly onto existing trio (specialist agent = Actor, evaluator-active = Evaluator, missing third role = Self-Reflection writer).

**Voyager (Wang et al., NVIDIA + Caltech)**: Three-component skill library architecture. Skills stored as executable code indexed by natural-language description embeddings. Top-K retrieval at task time. Self-verification gate before library admission. Maps directly onto MoAI v2: `my-harness-*` skills become the equivalent of Voyager's skill library, with embedding-based retrieval replacing manual seed.

**Constitutional AI / RLAIF (Bai et al., Anthropic 2022)**: Two-phase training. SL self-critique-revise against principles, then RL via AI-feedback preference scoring. The "constitution" is a natural-language principle list, not weights. Maps directly onto MoAI v2: `.claude/rules/moai/design/constitution.md` is already the explicit principle list — harness evolution proposals can be self-scored against it before AskUserQuestion approval.

Sources:
- [Reflexion arXiv:2303.11366](https://arxiv.org/abs/2303.11366): Original paper with full Actor/Evaluator/Self-Reflection formalism.
- [Reflexion GitHub](https://github.com/noahshinn/reflexion): Reference implementation for HumanEval/MBPP/LeetcodeHardGym benchmarks.
- [NeurIPS 2023 Poster](https://neurips.cc/virtual/2023/poster/70114): Conference artifact with results tables.
- [Voyager arXiv:2305.16291](https://arxiv.org/abs/2305.16291): Original Voyager paper.
- [Voyager project page](https://voyager.minedojo.org/): Demos and ablation data.
- [Voyager GitHub MineDojo](https://github.com/minedojo/voyager): MIT-licensed reference implementation.
- [Constitutional AI arXiv:2212.08073](https://arxiv.org/abs/2212.08073): Original CAI/RLAIF paper.
- [Anthropic CAI research page](https://www.anthropic.com/research/constitutional-ai-harmlessness-from-ai-feedback): Official Anthropic positioning.
- [Claude's Constitution](https://www.anthropic.com/news/claudes-constitution): The actual principle list used in production Claude training.

## Industrial Frameworks (Reflection in Production)

**LangGraph Reflection Pattern** is the 2026 production standard for self-correcting agents:

- Cyclic graph with Generator + Evaluator/Critic + Router nodes
- Hard iteration cap (2-3 iterations sweet spot, 4+ only for analytical tasks)
- Token budget tracking per cycle (compute multiplies linearly)
- Cognitive separation: Evaluator can be a smaller/specialized model
- Prebuilt `langgraph-reflection` package: main agent + critique subagent + conditional return-to-main

**AutoGen → Microsoft Agent Framework**: AutoGen in maintenance mode since 2026. MAF combines AutoGen's conversational orchestration with Semantic Kernel's enterprise features (session state, type safety, telemetry). Workflow API exposes sequential/parallel/Magentic orchestration patterns. Self-improvement still a stated research direction without shipped implementation.

Sources:
- [LangChain Blog: Reflection Agents](https://blog.langchain.com/reflection-agents/): Official LangChain treatment.
- [langgraph-reflection GitHub](https://github.com/langchain-ai/langgraph-reflection): Reference implementation.
- [SitePoint Agentic Design Patterns 2026 Guide](https://www.sitepoint.com/the-definitive-guide-to-agentic-design-patterns-in-2026/): Industry survey including iteration limits and trade-offs.
- [LearnOpenCV LangGraph Self-Correcting Agent](https://learnopencv.com/langgraph-self-correcting-agent-code-generation/): Code-generation case study showing Generate/CheckCode/Reflect three-node loop.
- [DEV Reflection Pattern Guide](https://dev.to/programmingcentral/stop-llms-from-lying-build-self-correcting-agents-with-the-reflection-pattern-1df): Practitioner perspective on the generate-critique-refine cycle.
- [Microsoft AutoGen GitHub](https://github.com/microsoft/autogen): Maintenance-mode v0.7.x line.
- [Microsoft Agent Framework Overview](https://learn.microsoft.com/en-us/agent-framework/overview/): Successor framework with workflow API.
- [AutoGen Research project](https://www.microsoft.com/en-us/research/project/autogen/): Self-improvement stated as research direction.

## Anthropic Claude Code Ecosystem (Skills/Agents Authority)

Definitive guidance for harness v2 implementation:

- Skills are **folders containing SKILL.md + scripts + assets + references**, not just markdown. Progressive disclosure: frontmatter (always loaded) → body (loaded on trigger) → references (on-demand).
- Commands have been merged into the skills namespace. `.claude/commands/X.md` and `.claude/skills/X/SKILL.md` both register `/X`.
- Agent field in skill frontmatter allows skill-to-subagent delegation; built-in agents include Explore, Plan, general-purpose. Custom agents under `.claude/agents/`.
- Anthropic's first-party design guidance: start with evaluation (find capability gaps), structure for scale (split when SKILL.md gets unwieldy), code is both tool and documentation, push beyond defaults (skill content should differ from base Claude knowledge), include Gotchas section (highest-ROI content), iterate incrementally.
- Security: skills can direct Claude to execute arbitrary code; treat as supply-chain risk. Only install from trusted sources.

Sources:
- [Claude Code Skills docs](https://code.claude.com/docs/en/skills): Authoritative Claude Code skills documentation.
- [Anthropic Agent Skills API docs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview): Cross-product (claude.ai, API, AWS, Microsoft Foundry) Skills standard.
- [Anthropic Engineering: Equipping Agents](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills): Best-practices essay from Anthropic engineering team.
- [anthropics/skills GitHub](https://github.com/anthropics/skills): Official Anthropic skills repository.

## Synthesis — 7 Key Insights for Harness v2

1. **Reflexion-style internal self-critique is additive to evaluator-active, not redundant.** Reflexion's Self-Reflection model writes verbal lessons that persist in episodic memory; evaluator-active scores against rubric. Combining them means the harness produces both a numeric score AND a natural-language explanation of what to change next iteration. Iteration cap should default to 3 (LangGraph empirical sweet spot).

2. **Voyager's three components map cleanly onto MoAI v2 architecture.** Auto-curriculum = observation-driven pattern detection. Skill library = `.claude/skills/my-harness-*/` with embedding-based retrieval replacing manual seed. Iterative prompting + self-verification = Tier-4 application + 3-7 day effectiveness measurement window.

3. **Constitutional AI gives MoAI a head-start.** `.claude/rules/moai/design/constitution.md` § FROZEN zone + § 5-Layer Safety is already an explicit principle list — equivalent to CAI's natural-language constitution. Harness evolution proposals can be self-scored against it BEFORE reaching the AskUserQuestion gate, reducing user fatigue.

4. **Embedding-cluster pattern detection scales where frequency-count cannot.** Current frequency-count requires ≥3 observations on identical tool/argument combos to promote to Tier 2. Embedding clusters detect semantic equivalence across surface variations (e.g., "fix this lint error" vs "linter failed, resolve") and promote earlier with less noise. Voyager's top-K retrieval over NL-description embeddings is the proven pattern.

5. **Multi-objective evaluation prevents single-axis over-optimization.** Pure quality optimization risks token-cost blowout and user-fatigue from excessive AskUserQuestion prompts. Production guidance from LangGraph community: track quality + cost + latency + iteration count as a tuple; reject evolutions that degrade any axis below baseline.

6. **Cross-project lesson federation requires anonymization layer.** Sharing `usage-log.jsonl` raw entries leaks file paths, function names, and commit SHAs. Need a hash/abstraction layer that strips identifiers while preserving pattern structure (verb + intent + outcome). Opt-in only; default off per privacy/IP concerns.

7. **CLI retirement is consistent with Anthropic's command-skill merger.** The 2026 Skills standard merges `.claude/commands/X.md` into `.claude/skills/X/SKILL.md`. Retiring `moai harness <verb>` Go CLI in favor of `/moai:harness` skill-only invocation follows the same trajectory: lifecycle commands move from binary subcommands to skill-orchestrated subagent execution. This is industry-aligned, not just MoAI-specific.

## Risk Signals

- **Reflexion iteration explosion**: If self-critique loop has no hard cap, token costs scale linearly per failed evolution. Mandatory hard cap of 3 iterations per cycle, plus token budget tracking.
- **Embedding model drift**: Pattern clusters formed under one embedding model may shift if model is upgraded. Persist clusters with embedding model version; warn on mismatch.
- **Constitutional rubric conflict with evaluator-active**: If self-critique against constitution and evaluator-active adversarial scoring disagree, which wins? Design constraint: constitution-based pre-screen is necessary-but-not-sufficient; evaluator-active retains veto power.
- **AskUserQuestion fatigue**: If Tier 4 application rate exceeds ~2/week, user fatigue compromises oversight quality. Multi-objective scoring should rate-limit Tier 4 proposals.
- **Skill library bloat**: Voyager's library grows monotonically; MoAI v2 needs effectiveness-decay pruning to avoid `my-harness-*/` directory accumulating dead skills.
- **Cross-project leak**: Even with anonymization, common code patterns can be re-identifying. Strict opt-in + namespace isolation.

## Opportunities

- **MoAI's 5-Layer Safety is genuinely differentiated.** No surveyed framework (AutoGen, LangGraph, CrewAI, MAF) implements Frozen Guard + Canary Check + Contradiction Detector + Rate Limiter + Human Oversight as a unified safety stack. This is unfair advantage material.
- **`.claude/rules/moai/design/constitution.md` is already a working constitution.** MoAI can deploy Constitutional AI self-critique with zero new authoring work — the constitution exists, the principles are explicit, the FROZEN zones are machine-verifiable.
- **Embedding infrastructure is local/lightweight.** Voyager's NL-description embedding indexing can run on small local models (e.g., bge-small) without API spend. No Anthropic API call for retrieval — only for evolution proposal generation.
- **Anthropic's command-skill merger validates CLI retirement.** Retiring `moai harness` Go subcommand in favor of `/moai:harness` skill aligns with Anthropic's 2026 trajectory.

## Sources Summary

| Source | Type | Relevance |
|--------|------|-----------|
| [revfactory/harness](https://github.com/revfactory/harness) | competitor | Baseline reference; 6-phase workflow; 6 architectural patterns |
| [revfactory/claude-code-harness](https://github.com/revfactory/claude-code-harness) | case_study | +60% quality A/B empirical evidence |
| [Harness public site](https://revfactory.github.io/harness/) | competitor | Public docs for 6 patterns |
| [ai-boost/awesome-harness-engineering](https://github.com/ai-boost/awesome-harness-engineering) | technical_ecosystem | Community curation across harness engineering dimensions |
| [Reflexion arXiv:2303.11366](https://arxiv.org/abs/2303.11366) | user_research | Canonical verbal-reinforcement self-improvement paper |
| [Reflexion GitHub](https://github.com/noahshinn/reflexion) | technical_ecosystem | Reference implementation |
| [Reflexion NeurIPS 2023](https://neurips.cc/virtual/2023/poster/70114) | user_research | Conference artifact |
| [Voyager arXiv:2305.16291](https://arxiv.org/abs/2305.16291) | user_research | Canonical skill-library lifelong-learning paper |
| [Voyager project page](https://voyager.minedojo.org/) | case_study | Empirical ablation data |
| [Voyager GitHub MineDojo](https://github.com/minedojo/voyager) | technical_ecosystem | Reference implementation (MIT) |
| [Constitutional AI arXiv:2212.08073](https://arxiv.org/abs/2212.08073) | user_research | Canonical principle-based self-critique paper |
| [Anthropic CAI research](https://www.anthropic.com/research/constitutional-ai-harmlessness-from-ai-feedback) | user_research | Official Anthropic positioning |
| [Claude's Constitution](https://www.anthropic.com/news/claudes-constitution) | user_research | Production principle list reference |
| [LangChain Blog: Reflection Agents](https://blog.langchain.com/reflection-agents/) | technical_ecosystem | Industry-standard reflection-loop documentation |
| [langgraph-reflection](https://github.com/langchain-ai/langgraph-reflection) | technical_ecosystem | Prebuilt reflection graph |
| [SitePoint Agentic Design Patterns 2026](https://www.sitepoint.com/the-definitive-guide-to-agentic-design-patterns-in-2026/) | market_data | 2026 industry survey including iteration trade-offs |
| [LearnOpenCV LangGraph Self-Correcting Agent](https://learnopencv.com/langgraph-self-correcting-agent-code-generation/) | case_study | Code-generation reflection pattern walkthrough |
| [Microsoft AutoGen GitHub](https://github.com/microsoft/autogen) | competitor | Maintenance-mode predecessor framework |
| [Microsoft Agent Framework](https://learn.microsoft.com/en-us/agent-framework/overview/) | competitor | AutoGen successor; self-improvement still vision-only |
| [AutoGen Research project](https://www.microsoft.com/en-us/research/project/autogen/) | competitor | Microsoft Research positioning |
| [Claude Code Skills docs](https://code.claude.com/docs/en/skills) | technical_ecosystem | Authoritative skills authoring guide |
| [Anthropic Agent Skills API docs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview) | technical_ecosystem | Cross-product Agent Skills standard |
| [Anthropic Engineering: Equipping Agents](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills) | technical_ecosystem | Best-practices essay |
| [anthropics/skills GitHub](https://github.com/anthropics/skills) | technical_ecosystem | Official Anthropic skills examples |

Total sources: 24 (target was 10-15; depth exceeds target due to high-quality WebSearch + Context7 yield).

Context7 library IDs resolved (for downstream documentation lookup if needed):
- `/anthropics/claude-code` (Claude Code, 761 snippets, score 79.95)
- `/websites/code_claude` (Claude Code official docs, 7393 snippets, score 81.52)
- `/websites/langchain_oss_python_langgraph` (LangGraph Python, 1121 snippets, score 89.5)
- `/langchain-ai/langgraph` (LangGraph core, 228 snippets, score 79.66)
- `/microsoft/autogen` (AutoGen, 1174 snippets, score 85.45)
