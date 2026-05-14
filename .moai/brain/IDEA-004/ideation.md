# Idea: CLI-free self-evolving harness v2 for MoAI-ADK
*Session: 2026-05-14 | Idea: IDEA-004*

## Converged Concept

A unified, CLI-free, self-evolving harness system for MoAI-ADK that subsumes three prior SPECs (V3R3-HARNESS-001, V3R3-HARNESS-LEARNING-001, V3R3-PROJECT-HARNESS-001) into one V3R4 architecture. The harness lifecycle (generation, observation, learning, evolution, application) runs entirely inside the `/moai:harness` skill + dedicated subagents + Claude Code hooks — no `moai harness <verb>` Go CLI dependency. Learning is multi-layer: signal observation across PostToolUse/Stop/SubagentStop/UserPromptSubmit events feeds embedding-cluster pattern detection; promotion uses Reflexion-style verbal reinforcement plus Constitutional-AI principle self-critique against the existing design constitution; Tier-4 application remains gated by 5-Layer Safety + AskUserQuestion approval. Skills self-organize Voyager-style into an embedding-indexed library that supports retrieval-augmented prompting for novel scenarios.

## Lean Canvas

### Problem

The top 3 problems this v2 harness solves:

1. **CLI dependency creates a brittle invocation path.** Current `moai harness <verb>` Go subcommand is not registered in the v2.14.0 binary (newHarnessCmd defined but never wired into root.go), causing `/moai:harness` slash command failures. The CLI layer is a single point of failure that breaks the entire harness lifecycle.
2. **Static 4-Tier ladder cannot adapt to project-specific evolution rhythms.** The Hwang (2026) borrowed promotion thresholds (3→5→10 observations) assume frequency-count detection over surface tool calls. It misses semantic equivalences across natural-language variations and produces both false positives (noise patterns) and false negatives (real patterns disguised by surface variation).
3. **Learning is intra-project siloed and one-directional.** Observations from project A never benefit project B. Successful evolutions never feed back into the harness generation step. There is no skill-library compounding effect — every project starts from the same seed.

### Customer Segments

Primary segment: **MoAI-ADK + Claude Code development teams** (1-10 engineers) running long-lived projects with active SPEC workflows, who already operate within the MoAI orchestrator and have project-specific recurring patterns that justify harness specialization.

Personas:
- **The Solo Power-User**: GOOS-style developer running multiple projects, observes recurring inefficiencies, wants the harness to learn without manual maintenance.
- **The Project Tech Lead**: Owns 2-5 MoAI projects, wants project-tailored specialists without authoring each one by hand, requires the safety gates to prevent autonomous regressions.
- **The MoAI Maintainer**: Authors MoAI-ADK itself, needs the harness to self-improve so common tooling tasks (CI fixes, dependency updates, test isolation) no longer require manual specialist authoring.

Secondary segment: **MoAI projects themselves** as a customer (each project gets its own harness instance, the meta-harness orchestrator serves the project as client).

### Unique Value Proposition

The only self-evolving Claude Code harness with principle-grounded autonomous improvement: combines Reflexion-style verbal self-critique with Constitutional-AI principle scoring against an explicit design constitution, gated by a 5-Layer Safety stack that no competing framework offers — delivering autonomous evolution that the user can trust without micro-managing.

### Solution

Top capabilities (5-8 bullets, capability-only):

- **Multi-event observation collection**: signals captured across tool-use, session boundaries, subagent completion, and user-prompt events — not just single-event sampling.
- **Embedding-cluster pattern detection**: semantic equivalence detection over observation log entries, replacing frequency-count thresholds with cluster density.
- **Verbal-reinforcement self-critique loop**: harness writes natural-language reflections on its own evolution proposals before promotion; capped at 3 iterations per cycle.
- **Principle-based self-scoring against design constitution**: proposals pre-screened against the project's explicit constitution principles before reaching human oversight.
- **Voyager-style skill library with embedding retrieval**: specialist skills indexed by description embedding; top-K retrieval at generation time for novel scenario coverage.
- **Multi-objective effectiveness measurement**: each evolution scored across quality + token cost + latency + iteration count; regressions on any axis trigger auto-rollback.
- **Effectiveness-decay pruning**: patterns lose confidence over time without re-observation; stale evolutions deprecate automatically.
- **Optional cross-project lesson federation**: opt-in anonymized lesson sharing via auto-memory namespace; default off for privacy.

### Channels

How users discover and adopt this harness:

- **Bundled with MoAI-ADK**: ships in every `moai init` scaffold; immediately available in every new project.
- **`/moai:harness` slash command invocation**: single entry point for all harness lifecycle operations (generate, observe, learn, apply).
- **Auto-load via meta-harness skill on `/moai project`**: project initialization auto-detects the need for harness setup and loads the meta-harness skill.
- **Hook-driven observation**: PostToolUse and other Claude Code hooks transparently record signals without user action.
- **AskUserQuestion notification on Tier 4 application**: user discovers evolution proposals through their existing review flow.

### Revenue Streams

For this FOSS project, "revenue" maps to **adoption and autonomous-evolution metrics**:

- **Weekly Tier-4 application count** (user-approved evolutions per week per project) — primary leading indicator of value delivered.
- **Tier-4 reach rate** (% of observations promoted to Tier 4) — efficiency metric for the promotion pipeline.
- **Cross-project lesson adoption count** (opt-in users referencing federated lessons) — network effect indicator.
- **MoAI-ADK contributor retention** (active monthly contributors) — sustainability indicator for the FOSS commons.

### Cost Structure

Main cost drivers:

- **Anthropic API spend per evolution cycle**: tokens consumed by Reflexion self-critique + evaluator-active + constitution self-scoring. Hard iteration cap (3) and multi-objective rejection criteria bound this.
- **Local compute for embedding indexing**: skill-description embeddings and observation-cluster embeddings; mitigated by using lightweight local models (bge-small class) rather than API-based embeddings.
- **MoAI maintainer time on safety stack maintenance**: 5-Layer Safety + Frozen Guard must be code-reviewed for every change; the FROZEN zone immutability discipline pays this cost upfront in exchange for runtime safety.
- **User attention budget for AskUserQuestion at Tier 4**: rate-limited to ≤2/week per project via multi-objective scoring to prevent oversight fatigue.

### Key Metrics

- **Autonomous evolution application count**: user-approved Tier-4 applications per week (primary success metric, user-defined).
- **Tier-4 reach rate**: ratio of observations promoted through Tier 1→2→3→4 ladder.
- **Evaluator score delta**: average quality improvement of evolved skill bodies vs pre-evolution baseline (measured by evaluator-active).
- **Pattern diversity index**: number of distinct embedding clusters with active patterns (proxy for harness coverage breadth).
- **Rollback rate**: % of Tier-4 applications auto-rolled-back within the 3-7 day effectiveness window (target <10%).

### Unfair Advantage

What is genuinely hard for competitors (revfactory/harness, AutoGen/MAF, LangGraph, CrewAI) to replicate:

- **Existing 5-Layer Safety architecture**: Frozen Guard + Canary Check + Contradiction Detector + Rate Limiter + Human Oversight. None of the surveyed competitors implements this stack. Replicating it requires not just code but the discipline of maintaining a FROZEN zone across many releases.
- **Working AskUserQuestion contract for orchestrator-only user interaction**: deferred-tool preload + Socratic interview structure + structured-only-questions rule. Industry frameworks rely on prose questions or web-UI confirmations; MoAI's structured channel is meaningfully harder for users to misinterpret.
- **Pre-existing design constitution**: `.claude/rules/moai/design/constitution.md` is already in place with FROZEN/EVOLVABLE zone definitions, principle list, and machine-readable adjacency files. Constitutional-AI self-scoring works on day one without authoring effort.
- **MoAI orchestrator + subagent contract isolation**: subagents in isolated contexts cannot prompt users; orchestrator owns the AskUserQuestion channel. This isolation discipline makes evolution proposals safer by construction.
- **16-language neutrality discipline**: project_markers-based language detection + per-language guideline files. Cross-project lesson federation can scope by language without baking language-specific assumptions into the harness core.

---

## Evaluation Report

### Strengths

**Evidence-backed strengths from critical evaluation:**

1. **Architecture is grounded in three peer-reviewed papers** (Reflexion NeurIPS 2023, Voyager NVIDIA/Caltech 2023, Constitutional AI Anthropic 2022). This is not speculative — every core mechanism has empirical validation in adjacent domains. Reflexion reports +8% absolute improvement from verbal self-critique; Voyager reports 3.3× more unique skill acquisition than baselines; CAI reports successful harmlessness training without human labels.

2. **CLI retirement aligns with Anthropic's command-skill merger trajectory (2026)**. Retiring `moai harness <verb>` is industry-aligned movement, not idiosyncratic. The skill namespace consolidation reduces user-facing surface area while increasing capability via subagent execution and progressive disclosure.

3. **5-Layer Safety + Frozen Guard is a genuine moat**. Survey of AutoGen, LangGraph, CrewAI, MAF reveals no equivalent stack. Replicating it requires sustained discipline (FROZEN zone, AskUserQuestion contract, immutable principles) — not just code.

4. **Constitutional AI rubric exists today**. `.claude/rules/moai/design/constitution.md` § FROZEN + § 5-Layer Safety + § Section 3.1-3.3 [HARD] principles can serve as the natural-language constitution that the harness scores itself against. No new authoring needed for v2 launch.

5. **Multi-objective scoring prevents single-axis runaway**. LangGraph production guidance (iteration cap 2-3, token budget tracking) is mature. MoAI v2 adopts this with quality + cost + latency + iteration-count tuple, eliminating the failure mode where pure quality-optimization triggers user-fatigue or token-cost blowout.

### Weaknesses

**Identified gaps, assumptions, and risks:**

1. **Hwang (2026) +60% effectiveness target may be unreproducible in MoAI's evaluator-active context.** The empirical evidence came from revfactory/harness's specific A/B setup with 15 SE tasks. MoAI's evaluator-active has different scoring axes (4-dimension scoring: Functionality/Security/Craft/Consistency) and different task distributions. Treat +60% as aspirational rather than baseline.

2. **Embedding model versioning fragility.** Pattern clusters formed under embedding model A may not align with clusters formed under model B after upgrade. Mitigation requires persisting the embedding model version with each cluster and emitting warnings on mismatch.

3. **Constitution self-critique vs evaluator-active veto resolution is non-trivial.** What if internal self-critique passes (constitution-aligned) but evaluator-active rejects (quality-deficient)? Or vice versa? Design constraint must be encoded: evaluator-active retains absolute veto power; constitution scoring is necessary-but-not-sufficient. This is a real design decision that the SPEC must lock in.

4. **AskUserQuestion fatigue risk is hard to quantify pre-deployment.** Tier-4 rate-limiting to ≤2/week is a guess. Real fatigue threshold depends on user role (Solo Power-User vs Tech Lead) and project state (active vs maintenance). Initial conservative cap (1/week) with adaptive widening based on user acceptance rate would be safer.

5. **Cross-project lesson federation has open privacy concerns even with anonymization.** Hashing file paths and function names doesn't eliminate re-identification through code-pattern fingerprinting. Strict opt-in is mandatory; default must be off; namespace isolation per organization. Even then, regulated industries (healthcare, finance) cannot opt in without legal review.

6. **Reflexion's reliance on LLM self-evaluation is an inherent limitation.** When self-critique converges on wrong conclusions, no formal mechanism corrects it. The 5-Layer Safety + AskUserQuestion + evaluator-active veto stack is the only safeguard. Must verify in practice that the safeguards catch the cases self-critique misses.

7. **Skill-library bloat is asymmetric.** Adding skills is automatic; removing requires effectiveness-decay pruning that has to be tuned. Initial pruning windows (30 days no re-observation → deprecate) are guesses. Real tuning requires usage data.

### First Principles Validation

**Per moai-foundation-thinking/modules/first-principles.md decomposition:**

**Minimum viable evolution loop** (what must be true for self-evolution to work at all):
- observation must be recorded reliably (Claude Code hook system, already exists)
- patterns must be detectable from observations (algorithm choice: frequency-count OR embedding-cluster — embedding is strictly more powerful)
- safety check must precede application (5-Layer Safety, already exists in MoAI)
- application must be reversible (snapshot + auto-rollback on effectiveness regression)
- effectiveness must be measurable (evaluator-active score delta, already exists)

All five primitives exist or have clear implementation paths. The architecture is feasible.

**Unchangeably FROZEN regardless of evolution algorithm sophistication** (per design constitution § Frozen vs Evolvable Zones):
- The constitution file itself
- Safety architecture (5 Layers)
- Pipeline phase ordering (manager-spec first, evaluator-active last in GAN loop)
- Pass threshold floor (≥0.60)
- Human approval requirement at Tier 4
- AskUserQuestion-Only Interaction rule

These cannot be evolved by any v2 mechanism. The evolution algorithm operates strictly within the EVOLVABLE zone.

**Absolute lower bound for human oversight** (per design constitution § 2 frozen + § 11.4 5-Layer Safety):
- All Tier 4 applications require AskUserQuestion approval (cannot be lowered)
- All evolution proposals must pass Canary Check (≤0.10 score drop) before user sees them
- All contradictions with existing rules must be surfaced explicitly, never silently overridden
- Maximum 3 evolutions per week (rate limit, may be tightened but not loosened)
- 24-hour cooldown between evolutions (may be tightened)

The lower bound is structurally enforced. Even with maximum algorithmic sophistication, the harness cannot deploy autonomous changes that bypass these gates.

### Adversarial Scenarios

**What if Hwang's +60% effectiveness target is unreproducible?**

Acceptable. The target was aspirational, derived from a different framework's A/B setup. MoAI v2 success doesn't depend on matching +60%; it depends on positive evolution-application count per week without negative rollback rate. Even +10% effectiveness with low rollback would justify deployment.

**What if cross-project lesson sharing leaks proprietary patterns?**

Cross-project federation is opt-in and default-off. The anonymization layer is necessary-but-not-sufficient. Mitigation: namespace isolation per organization, strict opt-in via AskUserQuestion at project init, sample-and-review workflow before patterns enter shared namespace. This feature can be deferred to a later SPEC if v1 launch needs to prioritize safety.

**What if observer log gets corrupted mid-evolution?**

Observer log is append-only JSONL. Corruption can only affect future entries, not past clusters. Pattern detection should be resilient to single-line parse failures (skip-and-log). The snapshot mechanism preserves a recoverable state.

**What if Reflexion-style self-critique conflicts with evaluator-active's adversarial role?**

This is the design constraint identified in Weakness #3. Resolution: evaluator-active retains absolute veto power. Self-critique runs first as a pre-screen (saves evaluator-active tokens by rejecting obviously bad proposals); evaluator-active runs second as the binding gate. Document this ordering explicitly in the SPEC.

**What if Tier 4 application latency exceeds user attention threshold?**

Multi-objective scoring includes latency as an axis. Evolutions that increase latency above a threshold get rejected before reaching Tier 4. The user-facing AskUserQuestion presentation must be concise (≤4 options, ≤30 second comprehension target).

### Verdict

**Proceed.** The concept is grounded in three peer-reviewed papers with empirical validation, exploits MoAI's existing assets (design constitution, 5-Layer Safety, AskUserQuestion contract, evaluator-active), and has a clear architectural path. The identified weaknesses are manageable through SPEC-level design decisions (evaluator-active veto, opt-in federation, conservative rate limits). The +60% effectiveness target should be downgraded to "positive evolution count with low rollback rate" as the v1 success criterion.

Recommended progression: write SPEC-V3R4-HARNESS-EVOLVE-001 as the unified replacement, with cross-project federation deferred to a separate later SPEC (SPEC-V3R4-HARNESS-CROSS-PROJECT-001) due to privacy complexity.
