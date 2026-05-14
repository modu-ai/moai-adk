# Extended Context: Self-Evolving Harness v2

> This file supplements prompt.md with additional context for your design session.
> It is NOT meant to be pasted into Claude Design — use prompt.md for that.

## Product Background

The product is a self-evolving harness system for developer tooling. A harness is a project-tailored configuration that defines specialized agents (each with a narrow domain of expertise) plus the skills those agents use. In the v2 release, the harness becomes self-evolving: it observes its own usage, detects patterns, proposes improvements, scores those improvements against an explicit design constitution, and applies approved changes within a five-layer safety architecture.

### Lean Canvas Summary

- **Problem (top 3)**: (1) CLI dependency creates a brittle single-point-of-failure invocation path; (2) static promotion thresholds cannot adapt to project-specific evolution rhythms and miss semantic equivalences across surface variations; (3) learning is intra-project siloed and one-directional, with no compounding effect across projects.
- **Customer Segments**: solo power-users, project tech leads, tooling maintainers, plus each MoAI project itself as a meta-customer.
- **Unique Value Proposition**: the only self-evolving harness combining verbal self-critique with principle-grounded scoring against an explicit design constitution, gated by a five-layer safety stack that no competing framework offers.
- **Solution capabilities (top 8)**: multi-event observation; embedding-cluster pattern detection; verbal-reinforcement self-critique loop capped at 3 iterations; principle-based self-scoring; auto-organizing embedding-indexed skill library; multi-objective effectiveness measurement with auto-rollback; effectiveness-decay pruning; opt-in cross-project lesson federation.
- **Channels**: bundled with MoAI-ADK; `/moai:harness` slash command; auto-load via meta-harness skill on project init; hook-driven observation; AskUserQuestion notification at the Tier-4 application gate.
- **Revenue Streams** (this is an open-source project; adoption metrics substitute): weekly Tier-4 application count, Tier-4 reach rate, cross-project lesson adoption count, contributor retention.
- **Cost Structure**: Anthropic API spend per evolution cycle (bounded by 3-iteration cap and rejection criteria); local compute for embedding indexing; maintainer time on safety stack maintenance; user attention budget for AskUserQuestion approvals (rate-limited to roughly 1-2/week per project initially).
- **Key Metrics**: autonomous evolution application count (primary), Tier-4 reach rate, evaluator score delta, pattern diversity index, rollback rate (target below 10%).
- **Unfair Advantage**: five-layer safety architecture (Frozen Guard, Canary Check, Contradiction Detector, Rate Limiter, Human Oversight); pre-existing design constitution that no competitor has; orchestrator-only AskUserQuestion contract preventing subagent user-prompting; 16-language neutrality discipline.

## Roadmap Context

The product launches across eight scoped specifications, executed in this dependency order:

1. Unified harness foundation (CLI retirement, skill+subagent+hooks lifecycle, 5-layer safety preservation)
2. Multi-event observer pipeline (signal coverage expansion)
3. Embedding-cluster pattern detector (semantic-equivalence detection)
4. Verbal-reinforcement self-critique loop (Reflexion-style with hard iteration cap)
5. Principle-based self-scoring against design constitution
6. Multi-objective effectiveness measurement with auto-rollback
7. Embedding-indexed skill library with retrieval-augmented generation
8. Opt-in cross-project lesson federation (privacy-sensitive, deferred until single-project validation)

Each specification represents an independently shippable improvement on top of the previous. Specification 8 is intentionally deferred because privacy concerns require demonstrated single-project track record first.

## Research Findings Summary

Three peer-reviewed academic papers form the theoretical scaffolding:

- **Reflexion (Shinn et al., NeurIPS 2023, arXiv:2303.11366)**: verbal reinforcement learning for language agents. Three-model architecture (Actor + Evaluator + Self-Reflection). Empirical +8% absolute improvement over episodic-memory baseline. Foundation for the self-critique loop in specification 4.
- **Voyager (Wang et al., NVIDIA + Caltech, arXiv:2305.16291)**: embodied lifelong learning with an auto-curriculum, embedding-indexed skill library, and iterative prompting with self-verification. 3.3× more unique skill acquisition than baseline agents. Foundation for the skill library in specification 7.
- **Constitutional AI / RLAIF (Bai et al., Anthropic 2022, arXiv:2212.08073)**: principle-based AI training using a natural-language constitution. Two-phase: SL self-critique-revise, then RL via AI-feedback preference scoring. Foundation for the principle-based self-scoring in specification 5.

Industrial frameworks surveyed: LangGraph reflection pattern (production standard for self-correcting agents, iteration cap 2-3 is empirical sweet spot), Microsoft AutoGen now in maintenance mode replaced by Microsoft Agent Framework (self-improvement remains a stated research direction without shipped implementation, validating the opportunity), revfactory/harness (the competitor whose 6-phase workflow and 6 architectural patterns serve as the baseline this product improves on, with empirical +60% quality A/B evidence from the sister repository).

Anthropic's 2026 Claude Code Skills and Agents documentation confirms that commands and skills have merged into a unified namespace, that progressive disclosure is the core design principle (frontmatter → body → references), and that skills are folders not just markdown — validating the CLI retirement decision as industry-aligned movement.

## Brand Context

The project's brand voice files exist as templates at `.moai/project/brand/` but are not yet populated. All three files (brand-voice.md, visual-identity.md, target-audience.md) contain `_TBD_` placeholders. As a result, the prompt.md uses the brand-absent fallback template with default professional-technical voice guidance.

Before using the prompt.md in Claude Design, the brand voice should be either populated through a brand interview process or edited directly in the prompt with the actual brand voice. The current default suggests technical confidence, principle-grounded direct prose, neutral grayscale palette with one restrained accent, and clean sans-serif plus monospace typography.

## What This Design Will Enable

After the design is completed and translated into a working site, it will become the public face of a developer-tooling release that:

- Replaces three superseded specifications with a single unified architecture
- Retires a binary subcommand path in favor of skill-based invocation
- Introduces academically-grounded self-improvement mechanisms that competitors do not offer
- Demonstrates that autonomous capability and human oversight are not in tension when the safety architecture is explicit

The design therefore needs to do two jobs: convince technical readers that the mechanisms are sound (depth, principle visibility), and reassure them that autonomy does not equal recklessness (the five-layer safety stack must be the most prominent visual element on the page).
