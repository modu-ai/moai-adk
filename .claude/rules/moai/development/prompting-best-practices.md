---
paths: "**/.claude/agents/**,**/.claude/skills/**"
---

# Prompting Best Practices (Claude Latest Models)

Condensed reference of Anthropic's official prompt-engineering guidance for Claude's latest models (Opus 4.8 / 4.7, Sonnet (current generation), Haiku 4.5), applied to MoAI agent prompts, skill bodies, and orchestrator output. Complements the Karpathy quick-reference (`.claude/rules/moai/development/karpathy-quickref.md`) and skill-writing craft (`.claude/rules/moai/development/skill-writing-craft.md`); cross-referenced from `.claude/rules/moai/development/agent-authoring.md`.

> **Loading scope**: read when authoring or tuning an agent prompt / skill body / system prompt. Reference: https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices

## Foundational Techniques (apply in order)

1. **Be clear and direct.** Treat the model as a brilliant new colleague with no context on your norms. State the desired output format and constraints explicitly. If you want "above and beyond" behavior, ask for it — do not rely on inference. Golden rule: if a colleague with minimal context would be confused by the prompt, so will the model.
2. **Add context / motivation.** Explaining *why* an instruction matters lets the model generalize correctly (e.g. "this will be read aloud, so never use ellipses" beats a bare "NEVER use ellipses").
3. **Use examples (multishot).** A few well-crafted examples are the most reliable way to steer format, tone, and structure. Make them relevant, diverse (cover edge cases), and structured — wrap each in `<example>` tags. 3–5 examples is the sweet spot. Positive examples beat negative instructions.
4. **Structure with XML tags.** Wrap distinct content types in their own tags (`<instructions>`, `<context>`, `<input>`) so the model parses mixed prompts unambiguously. Use consistent, descriptive tag names; nest when content is hierarchical.
5. **Give a role** in the system prompt. Even one sentence focuses behavior and tone.
6. **Long-context layout.** For 20k+ token inputs, put longform data at the TOP, above the query and instructions (improves quality by up to ~30% on multi-document tasks). Wrap documents in `<document>` tags with `<source>` metadata. Ask the model to quote relevant passages first to cut through noise.

## Output Control

- **Tell the model what to do, not what to avoid.** "Write in flowing prose paragraphs" beats "don't use markdown". Use XML format indicators (`<smoothly_flowing_prose_paragraphs>`) for strong steering. Matching your prompt's own style to the desired output style helps.
- **Verbosity calibrates to perceived complexity** on the latest models — shorter on simple lookups, longer on open-ended analysis. If you need a fixed style, prompt for it; prefer positive examples of the desired concision over negative "don't over-explain" instructions.
- **No prefill.** Prefilled assistant messages on the last turn are unsupported on Claude 4.6+ (HTTP 400). Migrate format-forcing prefills to Structured Outputs or a direct "respond without preamble" instruction; migrate continuations into the user turn.

## Tool Use & Parallelism

- **Be explicit to trigger action.** "Change this function" makes the edit; "can you suggest changes" only suggests. State the action verb directly when you want a tool call.
- **Don't over-prompt tool use.** On Opus 4.5/4.6+, aggressive "CRITICAL: you MUST use this tool" language causes *over*triggering. Use normal phrasing ("Use this tool when…"). Tools that undertriggered on older models now trigger appropriately.
- **Parallel tool calls** are a strong default on the latest models. The canonical instruction (already a MoAI HARD rule) is: make all independent tool calls in parallel; call sequentially only when one call's output feeds another's parameters; never guess missing parameters.

## Thinking & Reasoning

- **Adaptive thinking** (`thinking: {type: "adaptive"}`) is the mode for Opus 4.7+/4.8 and Sonnet (current) — the model self-allocates reasoning by effort + query complexity. Do NOT set `budget_tokens` (deprecated; rejected on Opus 4.7+). Control depth with the **effort** parameter, not a token budget. On Opus 4.8 thinking is OFF unless explicitly enabled; the `ultrathink` keyword is MoAI's canonical trigger for `effort: xhigh`.
- **Effort calibration**: `xhigh` for coding/agentic work, minimum `high` for intelligence-sensitive work, `medium`/`low` only for speed-critical or simple tasks. At `max`/`xhigh`, set a large max output budget (start ~64k) so the model has room to think and act across tool calls. Raising effort is the first lever for shallow reasoning — prefer it over prompt scaffolding.
- **Prefer general thinking instructions** ("think thoroughly", "reason through the tradeoffs") over hand-written step-by-step plans. Ask the model to self-check before finishing ("verify your answer against the test criteria").
- **Curb overthinking** when needed: "choose an approach and commit to it; avoid revisiting decisions unless new information contradicts your reasoning" — or simply lower effort.

## Agentic & Long-Horizon Patterns

- **Literal instruction following (Opus 4.8).** The model interprets prompts literally and does not silently generalize one instruction to other items. State scope explicitly: "apply to every section, not just the first." This precision is an asset for pipelines and structured extraction.
- **Persistence across context windows.** For long autonomous work, tell the model the context will be compacted and it should continue indefinitely — never stop early for token-budget reasons; save state to memory before the window refreshes. Use structured formats (JSON) for state data, freeform notes for progress, and git for checkpoints.
- **Subagent spawning is steerable.** Opus 4.8 spawns *fewer* subagents by default; earlier 4.x spawned *more*. Give explicit guidance: spawn multiple subagents in one turn when fanning out across items/files; work directly (no subagent) for tasks completable in a single response or that need shared context across steps.
- **Balance autonomy and safety.** Take local, reversible actions (edit files, run tests) freely; confirm before hard-to-reverse / shared-system / destructive actions (force-push, `rm -rf`, dropping tables, pushing, posting). Never use destructive shortcuts (`--no-verify`) to get past obstacles.
- **Minimize hallucination**: never claim facts about code not yet opened; read referenced files before answering.
- **Reduce stray file creation**: instruct cleanup of temporary scratch files at task end (aligns with MoAI temp-file hygiene).

## Anti-Overengineering (canonical snippet)

The latest models tend to overengineer — extra files, unnecessary abstractions, unrequested flexibility. The canonical counter-prompt (already encoded in MoAI's "Enforce Simplicity" core behavior):

> Avoid over-engineering. Only make changes directly requested or clearly necessary. A bug fix doesn't need surrounding cleanup; a simple feature doesn't need extra configurability. Don't add docstrings/comments/types to code you didn't change. Don't add error handling for scenarios that can't happen — validate only at system boundaries. Don't build abstractions for one-time operations or hypothetical future requirements. The right amount of complexity is the minimum needed for the current task.

## Code-Review Harness Note

On the latest models, a review prompt that says "only report high-severity issues" / "be conservative" is followed faithfully — the model investigates just as deeply but reports fewer low-severity findings (precision up, measured recall down). For coverage, separate finding from filtering: "Report every issue including uncertain/low-severity ones; a later step will rank them. For each, include confidence and severity." This applies to MoAI's `sync-auditor` and review skills.

## MoAI Alignment Summary

Most of this guidance is already encoded in MoAI doctrine — this file is the consolidated external reference:

- Literal instruction following + no `budget_tokens` + effort routing → `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ / 4.8 Prompt Philosophy
- Anti-overengineering + scope discipline + verify-don't-assume → `moai-constitution.md` § Agent Core Behaviors
- Persistence / never-stop-early → `.claude/output-styles/moai/moai.md` § Persistence & Context Awareness
- Parallel tool calls → `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution
- Subagent spawn steering → `moai-constitution.md` § Parallel Execution + `.claude/rules/moai/workflow/dynamic-workflows.md`

## Cross-references

- https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices — canonical source
- `.claude/rules/moai/development/karpathy-quickref.md` — 4 coding principles (Think/Simplicity/Surgical/Goal-driven)
- `.claude/rules/moai/development/skill-writing-craft.md` — skill-body craft
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter + prompt structure

---

Version: 1.0.0
Classification: Evolvable craft reference — applies when authoring agent prompts, skill bodies, and system prompts
