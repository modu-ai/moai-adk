# Native Invocation-Model Alignment

Policy-layer doctrine naming the axis on which a MoAI `/moai` subcommand's existence is justified: the **invocation model** of its nearest Claude Code native counterpart, NOT a naive "does it duplicate a native command" test.

> **Loading scope**: orchestration-level guidance. Read when deciding whether a MoAI `/moai` subcommand is justified against a native Claude Code command, when adding a new subcommand that overlaps a native capability, or when auditing the subcommand surface. This is **doctrine (codification) only** — no hook, lint rule, or runtime mechanism enforces the classification below.

## Why the naive axis is wrong

The naive question — "does a MoAI `/moai` subcommand duplicate a Claude Code native command?" — is the wrong axis. Two native commands with the same user-facing verb can differ fundamentally in **who can trigger them**. The correct axis is the **invocation model**: can an orchestrator/agent trigger the native command programmatically inside an automated pipeline, or can only a human typing the slash command trigger it?

## The Two-Category Taxonomy

The invocation-model axis sorts native Claude Code commands into two categories, discriminated by a single observable signal in the official docs.

### PROGRAMMATIC — orchestrator-auto-invocable

A **bundled skill** or **bundled workflow** is **prompt-based**: an orchestrator/agent CAN auto-invoke it programmatically via the `Skill()` / `Workflow()` tool inside an automated pipeline (unless auto-invocation is disabled — see the caveat below). In the commands reference each such command carries a `[Skill]` or `[Workflow]` marker.

The official `skills.md` states the defining property verbatim:

> "Unlike most built-in commands, which execute fixed logic directly, bundled skills are prompt-based: they give Claude detailed instructions and let it orchestrate the work using its tools."

**Classification heuristic**: a `[Skill]` / `[Workflow]` marker in the commands reference is **present** for bundled skills/workflows (PROGRAMMATIC) and **absent** for built-in commands. A narrow bridge exists — a few built-in commands are ALSO exposed through the `Skill` tool even though they carry no marker; those are a PROGRAMMATIC sub-case (see the matrix note on this).

### HUMAN-ONLY — human-typed only

A **built-in command** executes fixed CLI logic directly and carries NO `[Skill]` / `[Workflow]` marker. Only a human typing the slash command triggers it; an orchestrator CANNOT trigger it via any tool call, and it is not exposed through the `Skill` tool. The official `skills.md` names `/compact` as one such command that is NOT available through the `Skill` tool.

## Classification Matrix

Each row carries an inline citation anchor — an observed `[Skill]` / `[Workflow]` marker OR an official-docs quote. This classification was verified at run-phase by fetching the official `commands.md` (commands reference) and `skills.md`; the classification is not asserted from memory.

| Native command | Category | Citation anchor (observed) |
|----------------|----------|-----------------------------|
| `/code-review` | PROGRAMMATIC | `[Skill]` marker — commands reference: `/code-review [low\|medium\|high\|xhigh\|max\|ultra] [--fix] [--comment] [target]` **Skill** |
| `/simplify` | PROGRAMMATIC | `[Skill]` marker — commands reference: `/simplify [target]` **Skill** |
| `/loop` | PROGRAMMATIC | `[Skill]` marker — commands reference: `/loop [interval] [prompt]` **Skill**. "Omit the interval and Claude self-paces between iterations" — it self-paces when the interval is omitted, so it is NOT merely a fixed time-interval scheduler |
| `/deep-research` | PROGRAMMATIC | `[Workflow]` marker — commands reference: `/deep-research <question>` **Workflow** |
| `/security-review` | PROGRAMMATIC (built-in exposed via the `Skill` tool) | `skills.md`: "A few built-in commands are also available through the Skill tool, including `/init`, `/review`, and `/security-review`." No `[Skill]` marker in the commands reference, but the `Skill`-tool bridge makes it orchestrator-invocable |
| `/goal` | HUMAN-ONLY | No `[Skill]`/`[Workflow]` marker — commands reference: `/goal [condition\|clear] Set a goal`. Not exposed through the `Skill` tool |
| `/review` | PROGRAMMATIC (built-in exposed via the `Skill` tool) | `skills.md`: "A few built-in commands are also available through the Skill tool, including `/init`, `/review`, and `/security-review`." No `[Skill]` marker in the commands reference, but the `Skill`-tool bridge makes it orchestrator-invocable |
| `/clear` | HUMAN-ONLY | No `[Skill]`/`[Workflow]` marker — commands reference: `/clear [name] Start a new conversation with empty context`. Not exposed through the `Skill` tool |
| `/compact` | HUMAN-ONLY | No `[Skill]`/`[Workflow]` marker — commands reference: `/compact [instructions]`. `skills.md`: "Other built-in commands such as `/compact` are not [available through the Skill tool]" |

### Classification-divergence note (observation over provisional expectation)

The plan-phase provisional expectation placed `/security-review` and `/review` in the HUMAN-ONLY category (assuming a built-in command with no programmatic bridge). Run-phase verification of the official `skills.md` **overrode** that expectation: both are built-in commands that are ALSO available through the `Skill` tool, so an orchestrator CAN trigger them programmatically. The live observation wins; they are classified PROGRAMMATIC (built-in-exposed-via-`Skill`-tool sub-case). This divergence is recorded so the classification rests on observed evidence, not on a memory-asserted assumption.

### `/loop` framing correction (does NOT edit the source it corrects)

The sibling rule `.claude/rules/moai/workflow/goal-directive.md` describes native `/loop` as "a fixed time interval elapses (re-runs the prompt/command on a schedule)". That framing is INCOMPLETE. The official commands reference shows native `/loop` is a bundled `[Skill]` that **self-paces when the interval is omitted** ("Omit the interval and Claude self-paces between iterations") — not merely a wall-clock scheduler. This doctrine acknowledges and corrects the `goal-directive.md` framing; it does NOT edit `goal-directive.md` itself (that would be out of scope). Native `/loop` (a bundled `[Skill]`) remains distinct from MoAI's `/moai loop` (the diagnostic-driven Ralph Engine).

## Automation-Justification Thesis

A MoAI `/moai` subcommand's existence is justified by **AUTOMATION** — integrating a capability into the orchestrator's plan→run→sync pipeline WITHOUT a human typing the slash command. The invocation-model axis yields two justification axes.

- **Axis A — reuse the native PROGRAMMATIC command.** Where the nearest native equivalent is PROGRAMMATIC (a bundled skill/workflow, or a built-in exposed through the `Skill` tool), MoAI SHOULD prefer invoking the native command via `Skill()` / `Workflow()` over reimplementing it. Reuse over reinvention: the native skill already carries Claude Code's maintained logic.
- **Axis B — reimplement the native HUMAN-ONLY command.** Where the nearest native equivalent is HUMAN-ONLY (a built-in command with no `Skill`-tool bridge — e.g. `/goal`, `/clear`, `/compact`), a MoAI subcommand that automates that capability inside the pipeline is NOT redundant reinvention — it is the ONLY pipeline path, because no tool call can trigger a non-exposed built-in. A human would otherwise have to type the command by hand, breaking automation.

## Per-Subcommand Justification Notes

- **`feedback` — no native equivalent (unique justification).** `/moai feedback` has NO native counterpart. It targets the remote `modu-ai/moai-adk` tool repository (bug reports and feature requests about the MoAI-ADK tool itself), NOT the user's own repository. The native `gh issue create` targets the user's own repo, so it cannot serve this purpose. `/moai feedback` is therefore a legitimate first-class MoAI workflow rather than a reimplementation of anything.
- **`review` (including `review --security`).** The nearest native counterparts are `/review` and `/security-review`. Live verification (above) shows both are built-in commands ALSO exposed through the `Skill` tool — PROGRAMMATIC, NOT HUMAN-ONLY. So `/moai review`'s justification rests on **Axis A** (prefer invoking the native `/security-review` / `/review` via `Skill()` where available) PLUS a broader-orchestration argument: `/moai review` composes a security pass with `@MX` tag-compliance checking and code-review synthesis into one pipeline step, a composition the native single-purpose commands do not provide. Its justification does NOT rest on a HUMAN-ONLY premise (the provisional plan-phase expectation that `/security-review` was HUMAN-ONLY was corrected by observation).
- **`clean`.** The nearest native counterpart is `/simplify`, a bundled `[Skill]` (PROGRAMMATIC). By **Axis A**, `/moai clean` is a candidate for future alignment onto the native `/simplify` via `Skill()`. That reuse refactor is deferred to a follow-up effort; it is recorded here as an axis observation, not performed by this doctrine.
- **`loop`.** MoAI's `/moai loop` (Ralph Engine — diagnostic-driven fix loop over LSP / AST-grep / test / coverage findings) is distinct from the native `/loop` bundled `[Skill]` (which self-paces prompt re-runs). They are complementary, not duplicates; `/moai loop` knows the project's quality tooling and the SPEC lifecycle.

### Axis B worked illustration (HUMAN-ONLY automation)

The genuinely HUMAN-ONLY native commands are `/goal`, `/clear`, and `/compact` (built-in, no `Skill`-tool bridge). If a MoAI subcommand needed one of these capabilities *inside* an automated pipeline step, Axis B would justify a MoAI reimplementation, because no tool call can trigger a non-exposed built-in — a human would otherwise have to type it. MoAI does not currently reimplement any of the three; the Axis B principle is stated here for completeness so a future subcommand facing a genuinely HUMAN-ONLY native counterpart has a recorded justification path.

## Conditional-PROGRAMMATIC Caveat (verify before relying on Axis A)

PROGRAMMATIC status is **not unconditional**. Auto-invocability can be revoked at the skill or the settings level:

- A bundled skill carrying `disable-model-invocation: true` in its frontmatter loses auto-invocability. The official `skills.md`: "Add `disable-model-invocation: true` to prevent Claude from triggering it automatically." Such a skill can then only be invoked by a human typing `/skill-name`.
- The settings-level `disableBundledSkills` disables the bundled-skill set entirely. The official `skills.md`: bundled skills "are available in every session unless disabled with the `disableBundledSkills` setting".
- The `Skill` tool itself can be denied in `/permissions`, which also removes the `Skill`-tool bridge that makes `/security-review` and `/review` orchestrator-invocable. The official `skills.md`: "Disable all skills by denying the `Skill` tool in `/permissions`."

**Consequence**: before relying on Axis A (invoking a native PROGRAMMATIC command via `Skill()`), an orchestrator/agent MUST verify at runtime that the target is actually auto-invocable in the current session — a bundled skill with `disable-model-invocation: true`, a session with `disableBundledSkills`, or a denied `Skill` tool all remove auto-invocability. This runtime verification is an agent obligation, not a mechanically-enforced gate (this doctrine is policy-layer only).

## Scope boundary

This doctrine (1) names the invocation-model axis, (2) classifies the overlapping native commands, and (3) applies the axis to `/moai feedback` as the worked example that proves it. Refactoring existing MoAI subcommands onto native bundled skills via Axis A (for example, pointing `/moai clean` at native `/simplify`, or `/moai review` at native `/code-review`) is a separate follow-up effort — the axis observation is recorded above, but no existing subcommand is retired or re-pointed by this doctrine. The classification is codification only; no runtime mechanism enforces the PROGRAMMATIC / HUMAN-ONLY tags.

## Cross-references

- Official Claude Code `commands.md` (commands reference) — the `[Skill]` / `[Workflow]` marker source for the classification matrix.
- Official Claude Code `skills.md` — the "bundled skills are prompt-based" defining quote, the `Skill`-tool bridge for `/init` / `/review` / `/security-review`, and the `disable-model-invocation` / `disableBundledSkills` caveats.
- `.claude/rules/moai/workflow/goal-directive.md` — the `/loop` "time-interval scheduler" framing acknowledged and corrected above (not edited by this doctrine).
- `.claude/rules/moai/workflow/dynamic-workflows.md` — the bundled `/deep-research` `[Workflow]` and the Workflow orchestration primitive.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the no-unobserved-claim invariant that governs the run-phase docs verification recorded for this matrix.

---

Version: 1.0.0
Classification: Canonical Reference (policy-layer codification) — the invocation-model axis for justifying MoAI `/moai` subcommand existence. No runtime enforcement.
