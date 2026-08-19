---
description: Canonical reference for AskUserQuestion-only interaction protocol, ToolSearch deferred-tool preload procedure, and Socratic interview standards
---

# AskUserQuestion Protocol — Canonical Reference

> This file is the **single source of truth** for AskUserQuestion interaction rules.
> Cross-referenced by: CLAUDE.md §8, moai-constitution.md §MoAI Orchestrator, agent-common-protocol.md §User Interaction Boundary, output-styles/moai/moai.md §3/§10.
>
> **Loading scope**: Intentionally always-loaded (no `paths:` restriction). The orchestrator may compose an `AskUserQuestion` on any non-trivial turn, so the channel-monopoly rule and the ToolSearch deferred-tool preload procedure must be available every session.
>
> **Detail companion**: `askuser-protocol-reference.md` (paths-scoped to this file) — recommendation-placement evidence base, preview-field usage catalogue, and the Non-ASCII encoding root-cause mechanism / pollution-loop detail. Read it when those details are needed.

---

## Channel Monopoly

**AskUserQuestion is the only user-facing question channel.** The MoAI orchestrator MUST route every user-facing question through an `AskUserQuestion` tool invocation. Free-form interrogative prose in the response body is **prohibited** as a question channel.

Applies to all orchestrator turns involving: clarification questions (Stage 1 Clarify), preference/decision questions ("Which approach?", "Continue or abort?"), Socratic interview rounds during Context-First Discovery (CLAUDE.md §7 Rule 5), branch and workflow selection, and conflict resolution (merge strategy, rollback confirmation, etc.).

**Exceptions** (free-form prose questions permitted ONLY when):
- `AskUserQuestion` is technically unavailable — should not occur in normal orchestrator operation
- The expression is a statement of status that happens to end with a question mark, not a genuine request for a decision

**Anti-pattern (NEVER repeat)**: a free-form prose question with a `- A: / - B:` option list in the response body. **Correct pattern**: always use `AskUserQuestion` (see §Free-form Circumvention Prohibition for the "Other" mechanism).

---

## ToolSearch Preload Procedure

`AskUserQuestion` is a **deferred tool** in Claude Code. Its JSON schema is NOT loaded into the active context at agent initialization time. Attempting to invoke it without first selecting it results in `InputValidationError: tool not in schema`.

### Mandatory Preload Step

Immediately before **every** `AskUserQuestion` call, the orchestrator MUST invoke:

```
ToolSearch(query: "select:AskUserQuestion")
```

### General Rule for Deferred Tools

Any deferred tool requires a `ToolSearch` select preload before invocation. The pattern generalizes: `ToolSearch(query: "select:<tool>[,<tool>...]")` — single (`select:AskUserQuestion`) or multiple (`select:AskUserQuestion,TaskCreate`).

**Preload sequence** (per turn — if a new turn begins and `AskUserQuestion` will be called again, preload again; never reverse or omit Step 1):

```
[Turn N]
Step 1: ToolSearch(query: "select:AskUserQuestion")   ← preload deferred schema
Step 2: AskUserQuestion({ questions: [...] })           ← now valid to invoke
```

---

## Socratic Interview Structure

When a Stage 1 Clarify trigger is satisfied (see §Ambiguity Triggers and Exceptions), the orchestrator conducts a **Socratic interview** through sequential `AskUserQuestion` rounds (each round: ToolSearch preload → AskUserQuestion; later rounds build on earlier answers; final round is the confirmation — "Proceed with this plan?").

### Structural Constraints (all mandatory)

1. **Round limit**: Maximum 4 questions per `AskUserQuestion` call (Claude Code hard limit)
2. **Option limit**: Maximum 4 options per question (Claude Code hard limit)
3. **First option label**: MUST carry the `(권장)` (Korean) or `(Recommended)` (English) suffix to signal the recommended choice
4. **Language**: All question text, option labels, and option descriptions MUST be in the user's `conversation_language` (read from `.moai/config/sections/language.yaml`)
5. **Round progression**: Each subsequent round MUST narrow ambiguity by building on previous answers — repeating the same question is prohibited
6. **Termination condition**: Rounds continue until intent clarity reaches 100%; the interview MUST NOT end prematurely
7. **Pre-execution confirmation**: After clarity is achieved, consolidate findings into a brief report and obtain **explicit final confirmation** via `AskUserQuestion` before irreversible actions

> **Note**: "Interview round" denotes a turn of Socratic questioning (generic English usage), NOT the retired SPEC taxonomy term `Round` (folded into `Milestone` per `.claude/rules/moai/development/sprint-round-naming.md`).

---

## Option Description Standards

Every option in an `AskUserQuestion` call MUST have a `description` field populated with sufficient detail for the user to evaluate implications and trade-offs **without consulting external context**.

Each option description MUST include:
1. **Immediate result**: What happens immediately if this option is selected
2. **Side effects and risks**: Any follow-on consequences, risks, or irreversibility
3. **Quantitative information** (where applicable): Token cost, latency, file count, etc. (e.g., "saves ~30K tokens", "modifies 5 files")

**Bias prevention**: Option descriptions MUST use neutral, factual language — no persuasive or deprecating tone. The recommendation signal is conveyed **exclusively** through the `(권장)` / `(Recommended)` label suffix on the first option; descriptions must not phrase the recommended option more favorably or the non-recommended options more negatively than the facts justify. ("This is the best approach because..." is bias — state facts only.)

---

## Recommendation Placement Principles

The `(Recommended)` / `(권장)` label MUST be grounded in the statistically-majority rational default the user has actually been observed to select — never a policy default the system wants to push. Five principles bind its placement; the reasoning, the evidence base, and the worked detail live in `askuser-protocol-reference.md` § Recommendation Placement Principles.

1. **Emission timing.** Ask when the decision is genuinely uncertain (estimated p ≈ 0.5, where information gain peaks). When the outcome is nearly certain, auto-resolve to the majority option and omit the question.
2. **Question ordering.** Within one call, order questions by descending information gain — the highest-gain question first.
3. **Recommended option = observed majority.** Where observations are insufficient (cold start), fall back to the static default AND disclose that in the option description ("based on static default, N observations needed for personalization", or its `conversation_language` equivalent). An undisclosed cold-start recommendation is an unobserved-recommendation claim (`verification-claim-integrity.md` §1.1 surface 3).
4. **Precondition statement.** The recommended option's `description` MUST state the condition under which the recommendation holds ("Recommended when <precondition>"), so the user can reject it immediately when the precondition does not apply.
5. **Adaptive strength.** High estimated proficiency → weak recommendation (disclose the inferred preference, omit the label). Low proficiency → strong recommendation (label + transparent rationale). Proficiency unknown (early sessions) → neutral: place no inferred-preference label at all.

---

## Preview Field Standards

The `preview` field renders a monospace block beside the option list, switching the TUI to a side-by-side layout. It **complements** `description` and never replaces it: `description` is always required. Full usage catalogue, format guidance, and a worked example: `askuser-protocol-reference.md` § Preview Field Standards.

- **Use it** when options differ structurally or quantitatively and benefit from visual comparison (SPEC selection, migration strategies, tier envelopes, architecture variants); prefer a consistent key set across all options so deltas scan vertically
- **Skip it** when labels and descriptions already suffice — yes/no confirmations, permission grants, continue/abort gates
- [HARD] **Single-select only.** `preview` is silently dropped when `multiSelect: true`. If multi-select is required, put the content in richer `description` text instead
- **Keep it short.** The preview pane does not scroll; content past roughly 12 lines is truncated with no way to reach it
- **Bias prevention is inherited.** The recommendation signal is carried *only* by the `(Recommended)` / `(권장)` label; preview content stays neutral and factual

---

## Report-Before-Ask Gate

[ZONE:Evolvable] [HARD] A decision-type `AskUserQuestion` whose options derive from investigation results MUST be preceded — in the same turn's response body — by a substantive findings report. Investigation results include: `Agent()` fan-out returns (multi-lens analysis, audits, scans), verification batches, and any multi-source evidence gathering the orchestrator performed before composing the question. Asking the user to choose among options they were never given the evidence to evaluate is a gate violation, even when the AskUserQuestion call itself is structurally compliant (labels, descriptions, previews, `(권장)` placement).

### Requested-Deliverable Primacy (user requirement analysis first)

[ZONE:Evolvable] [HARD] When the user's latest message explicitly requests a report, analysis, or explanation ("report on X", "explain why", "analyze this first"), that requested deliverable IS the turn's terminal output: the orchestrator MUST complete the report as a standalone response and end the turn WITHOUT appending a decision-type `AskUserQuestion` to the same turn. Pipeline-stage needs (clarification resolution, scope selection, audit-gate unblocking, next-step routing) NEVER override or preempt the user's stated information request.

- **Requirement analysis before question composition**: re-read the user's latest message; if it asks for information, deliver the information and stop — ask only when it asks for (or clearly requires) a decision
- **No question-as-epilogue**: a scope/next-step question appended to a requested report demotes the report to a preamble — deliver the report; let the user respond
- **Deferred pipeline questions**: pending workflow questions surface in a LATER turn — after the user reacts to the report, or explicitly says to proceed

### Report Completeness Criteria (all mandatory)

1. **Per-source coverage**: the report names each investigation source (agent, lens, audit dimension) and states its key findings with quantification (N findings, severity/classification breakdown). A single-line completion claim ("investigation complete", in any locale) is NOT a report.
2. **Option-to-report traceability**: every codename, identifier, or finding referenced in the question's option labels / descriptions / previews MUST have been introduced and explained in the preceding report body — the user cannot evaluate what was never explained.
3. **Structured rendering**: render the report via the Discovery banner (`.claude/output-styles/moai/moai.md` §8 Discovery Report) or equivalent structured markdown with per-source subsections, scaled to the investigation's size.

### Preview-as-Report Substitution (named anti-pattern)

[HARD] Option `preview` / `description` fields MUST NOT be the sole carrier of investigation findings. The preview compresses a comparison; the report explains the evidence. Compressing all findings into an option preview table while the response body carries only a one-line completion claim is the named anti-pattern **preview-as-report substitution**.

### Report-Promise Fulfillment

[HARD] When prior narration in the same task promised a consolidated report ("I will consolidate and report", or its equivalent in any locale), the report MUST be rendered before any subsequent decision AskUserQuestion. Claiming the report was delivered when none was rendered is an unobserved completion claim (`verification-claim-integrity.md` §1.1 surface 1).

### Exceptions (gate does not apply)

1. Pure clarify rounds during Context-First Discovery — questions asked BEFORE any investigation exists
2. Confirmation gates on already-reported context (e.g., Implementation Kickoff Approval after plan artifacts were presented in prose)
3. Blocker re-delegation rounds where the subagent's blocker report was already surfaced
4. Preference questions with no investigative basis (naming, formatting choices)

### Pre-emit self-check (report-before-ask) — 5 items

- [ ] Does the user's latest message request a report / analysis / explanation rather than a decision? If yes, this turn ends with the report — defer this AskUserQuestion to a later turn.
- [ ] Do this question's options derive from investigation results? If yes, does a substantive report precede this call in the same turn?
- [ ] Is every codename / identifier appearing in the options explained in the preceding report?
- [ ] Do the findings live in the response body (not only inside option previews)?
- [ ] If a report was promised earlier in the task, has it actually been rendered?

---

## Orchestrator–Subagent Boundary

The `AskUserQuestion` interaction channel is **asymmetric** by design.

### Orchestrator Obligations

The MoAI orchestrator (main session) MUST:
- Use `AskUserQuestion` as the exclusive channel for all user-facing questions
- Preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")` before each call
- Collect all necessary user preferences **before** delegating to subagents
- On receiving a blocker report from a subagent: run an `AskUserQuestion` round with the user, inject the user's responses into a fresh subagent prompt, and re-delegate

### Subagent Prohibitions

Subagents invoked via `Agent()` operate in isolated, stateless contexts and CANNOT interact with users directly:
- [ZONE:Frozen] [HARD] Subagents MUST NOT invoke `AskUserQuestion`
- [ZONE:Frozen] [HARD] Subagents MUST NOT output free-form prose questions directed at the user
- [ZONE:Frozen] [HARD] Subagents MUST NOT embed AskUserQuestion call syntax in their response body

### Blocker Report Format / Re-delegation Procedure

Owned by `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format and § Re-delegation Procedure — see there.

---

## Ambiguity Triggers and Exceptions

This section is the **single source of truth** for Stage 1 Clarify trigger conditions. Both `CLAUDE.md §7 Rule 5` and `CLAUDE.md §8 Ambiguity Triggers` cross-reference this definition.

### The Four Triggers (any one activates Stage 1)

1. **Pronoun or demonstrative without clear referent**: "this", "that", "it", "the previous one" — the referent cannot be unambiguously determined from context
2. **Multi-interpretable action verb without specified scope**: "clean up", "process", "improve", "fix" — the action could apply to multiple different implementations
3. **Unclear boundaries**: How far to go, how much to change, which files are in scope, where to stop
4. **Potential conflict with existing state**: Uncommitted changes, in-progress branches, overlapping work that the request might conflict with

### The Five Exceptions (Stage 1 is skipped)

1. Single-line typo or formatting fix — scope is self-evident
2. Bug fix with explicit reproduction provided — the reproducer defines scope
3. Direct file read when the path is explicitly specified — no interpretation needed
4. Command invocation with all required arguments provided — no ambiguity
5. Continuation of previously confirmed work in the same session — intent already established

### The Unknowns 4-Quadrant Lens

Classify the ambiguity by **user blind spot** (Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns):

- **Known-Knowns** — stated + confirmed facts. No clarification needed
- **Known-Unknowns** — gaps the user is aware of. Resolve via a Socratic interview round (§ Socratic Interview Structure)
- **Unknown-Knowns** — constraints implicit in the codebase the user has not surfaced. Resolve via `Agent(Explore)` read-only reconnaissance, then confirm with the user
- **Unknown-Unknowns** — risks neither side has articulated. When suspected (unfamiliar domain/subsystem/design territory), run a Blind Spot Pass (§ Blind Spot Pass) before plan-phase entry

### First-Action Sequence After Trigger

```
Trigger detected
  → Step 1: ToolSearch(query: "select:AskUserQuestion")   [deferred tool preload]
  → Step 2: Compose AskUserQuestion round (≤4 Q, ≤4 options, (권장) first, conversation_language)
  → Step 3: Send AskUserQuestion, collect responses
  → Step 4: Assess intent clarity (100% required)
  → Step 5: If <100%: go to Step 1 with narrowed questions
             If 100%: consolidate report → final confirmation → execute
```

---

## Blind Spot Pass

The **Blind Spot Pass** is an OPTIONAL pre-plan Discovery technique for surfacing the user's **unknown-unknowns**: read-only reconnaissance by `Agent(Explore)`, with findings surfaced to the user through the orchestrator's `AskUserQuestion` channel.

- **When**: the user is working in an **unfamiliar** domain (new subsystem, unfamiliar design/library territory) AND the orchestrator suspects unknown-unknowns — SHOULD run **before plan-phase entry**, before authoring the SPEC. The trigger is a judgment call, NOT an automatic gate; in a familiar domain with no suspected unknown-unknowns, the pass is skipped with no forced overhead.
- **Mechanism**: (1) spawn `Agent(Explore)` in **read-only** mode to scan the relevant domain (subsystem, library surface, integration points); (2) surface the likely unknown-unknowns through a single `AskUserQuestion` round so the user can react before the plan is authored.
- **Subagent boundary (preserved)**: `Agent(Explore)` — and any subagent — **does not prompt the user** directly; findings surface only through the orchestrator's channel. A subagent that lacks input returns a blocker report; it never asks the user.

---

## Free-form Circumvention Prohibition

Free-form interrogative prose in the response body MUST NOT be used as a substitute for `AskUserQuestion` — always use AskUserQuestion.

`AskUserQuestion` automatically appends an **"Other"** option to every question set: users preferring free-form answers select "Other" and type their response, so the orchestrator does NOT need free-form questions to support free-form answers. The "Other" mechanism covers edge cases not anticipated in the option list, preferences that do not fit the options, and free-form elaboration on a structured choice.

**Prohibited patterns** (all are Channel Monopoly violations):
- A free-form question in prose ("Which direction would you like to proceed?")
- A markdown option list in prose (`- **A**: … / - **B**: … / - **C**: …`)
- An inline question at the end of a response paragraph ("I've completed the changes. Should I create a PR now?")

**Correct pattern**: `ToolSearch(query: "select:AskUserQuestion")` → `AskUserQuestion({ questions: [{ question, header, options: [{ label: "... (Recommended)", description: "..." }, ...] }] })`.

### Completion-Report Next-Step Discipline

[ZONE:Evolvable] [HARD] A completion report (a "done" / "All Done" summary) MUST NOT end with a free-form prose next-step question — "What would you like to do next?", "무엇을 도와드릴까요? (예: A / B / C)", or the same idea in any `conversation_language`, optionally trailed by parenthetical or dashed option examples. This is a Channel Monopoly violation even when the report body itself is correct.

A completion report has exactly TWO valid closes:

1. **Route a genuine next-step decision through `AskUserQuestion`** — preload, then ask, so the user selects-and-enters instead of typing. The recommended option carries the `(Recommended)` / `(권장)` label.
2. **Close with NO question** — a clean completion statement (what was done, the evidence, the current state). When no decision is actually required, do NOT manufacture a next-step question; an unneeded prompt is noise.

"Ask through `AskUserQuestion`, or do not ask" — there is no third "ask in prose" option. The convenience rationalization "a short trailing next-step question on a finished report can be plain prose" is the exact failure mode this clause forbids.

**Pre-emit self-check (completion report)** — before sending any "done" report:
- [ ] Does the report end with a `?`-bearing prose next-step prompt? If yes → convert to `AskUserQuestion`, or drop the prompt entirely.
- [ ] If a next-step decision is genuinely needed, is it routed through `AskUserQuestion` (not prose, not a markdown option list)?
- [ ] If no decision is needed, does the report close cleanly with no manufactured question?

## Non-ASCII Tool-Call Encoding

The `AskUserQuestion` payload — `question`, `header`, and every option `label` / `description` / `preview` — routinely carries text in the user's `conversation_language`. For Korean, Japanese, Chinese, and other multi-byte scripts, this text MUST be written as **native UTF-8 directly** in the tool-call JSON. Hand-authored `\uXXXX` escape sequences are **PROHIBITED**.

**Failure Mode**: a malformed escape (stray space, truncated code point, half-written `\u`) corrupts the JSON so the `questions` array parses as a bare string — the call is rejected with `Invalid tool parameters` / `InputValidationError`, and the clarification round silently fails on its first attempt. (Root-cause mechanism, the self-reinforcing pollution loop, and the scope note: `askuser-protocol-reference.md` § Non-ASCII Tool-Call Encoding detail.)

### Directive and Recovery

- **Preventive (always):** write all `conversation_language` text as native UTF-8 in the tool-call JSON — this binds **every** tool call carrying multi-byte text, not only `AskUserQuestion` but Bash commands, Write / Edit content arguments, and any other tool-call payload. Never hand-escape a non-ASCII character.
- **Recovery (on failure):** if a call is rejected with `Invalid tool parameters` and the payload contained non-ASCII text, re-issue the identical call with the text rewritten as native UTF-8 — do not try to "repair" the escape sequence. Do not carry the corrupted form forward; re-author the next non-ASCII payload from the intended source text, not by transcribing the `\uXXXX` run visible in context. Persistent recurrence within a session → escalate to `/clear` with a paste-ready resume (last-resort loop-break).

### Pre-Emit Self-Check (before any tool call carrying non-ASCII text) — 3 items

- [ ] Is every `conversation_language` string in this payload written as native UTF-8 characters (한글 / 日本語 / 中文), with **zero** hand-authored `\uXXXX` sequences?
- [ ] Am I authoring this text from the intended source meaning, not transcribing an escaped `\uXXXX` run visible in my own context?
- [ ] If a prior call in this turn already failed with `Invalid tool parameters` on non-ASCII text, have I re-authored — not repaired — this payload, and am I watching for a saturated context that warrants `/clear`?

---

Version: 1.3.0
Classification: Canonical Reference — do not duplicate content; cross-reference this file instead.
