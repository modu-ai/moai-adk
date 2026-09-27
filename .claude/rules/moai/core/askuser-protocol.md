---
description: Canonical reference for AskUserQuestion-only interaction protocol, ToolSearch deferred-tool preload procedure, and Socratic interview standards
---

# AskUserQuestion Protocol — Canonical Reference

> This file is the **single source of truth** for AskUserQuestion interaction rules.
> Cross-referenced by: CLAUDE.md §8, moai-constitution.md §MoAI Orchestrator, agent-common-protocol.md §User Interaction Boundary, output-styles/moai/moai.md §3/§10.
>
> **Loading scope**: Intentionally always-loaded (no `paths:` restriction). The orchestrator may compose an `AskUserQuestion` on any non-trivial turn, so the channel-monopoly rule and the ToolSearch deferred-tool preload procedure must be available every session.
>
> **Detail companion**: `askuser-protocol-reference.md` — recommendation-placement evidence base, preview-field usage catalogue, the Non-ASCII encoding root-cause mechanism / pollution-loop detail, § Blind Spot Pass, and the bodies relocated from here: § General Rule for Deferred Tools · § The Four Triggers · § The Five Exceptions · § The Unknowns 4-Quadrant Lens · § First-Action Sequence After Trigger · § Directive and Recovery · § Pre-Emit Self-Check (non-ASCII) — 3 items · § Pre-emit self-check (report-before-ask) — 5 items. Load it when classifying an ambiguity trigger, preloading a deferred tool, recovering a rejected non-ASCII payload, or running a pre-emit self-check.

---

## Channel Monopoly

**AskUserQuestion is the only user-facing question channel.** The MoAI orchestrator MUST route every user-facing question through an `AskUserQuestion` tool invocation. Free-form interrogative prose in the response body is **prohibited** as a question channel.

Applies to every orchestrator turn involving clarification (Stage 1 Clarify), a preference or decision ("Which approach?", "Continue or abort?"), a Socratic interview round during Context-First Discovery (CLAUDE.md §7 Rule 5), branch and workflow selection, or conflict resolution.

**Exceptions** (free-form prose questions permitted ONLY when):
- `AskUserQuestion` is technically unavailable — should not occur in normal orchestrator operation
- The expression is a statement of status that happens to end with a question mark, not a genuine request for a decision

**Anti-pattern (NEVER repeat)**: a free-form prose question with a `- A: / - B:` option list in the response body. **Correct pattern**: always use `AskUserQuestion` (see §Free-form Circumvention Prohibition for the "Other" mechanism).

---

## ToolSearch Preload Procedure

`AskUserQuestion` is a **deferred tool**: its JSON schema is not loaded at agent initialization, so invoking it without selecting it first yields `InputValidationError: tool not in schema`.

### Mandatory Preload Step

Immediately before **every** `AskUserQuestion` call, the orchestrator MUST invoke:

```
ToolSearch(query: "select:AskUserQuestion")
```

## Socratic Interview Structure

When a Stage 1 Clarify trigger is satisfied (see §Ambiguity Triggers and Exceptions), the orchestrator conducts a **Socratic interview** through sequential `AskUserQuestion` rounds (each round: ToolSearch preload → AskUserQuestion; later rounds build on earlier answers; final round is the confirmation — "Proceed with this plan?").

### Structural Constraints (all mandatory)

1. **Round limit**: Maximum 4 questions per `AskUserQuestion` call (Claude Code hard limit)
2. **Option limit**: Maximum 4 options per question (Claude Code hard limit)
3. **First option label**: MUST carry the `(권장)` (Korean) or `(Recommended)` (English) suffix to signal the recommended choice — this is the `push`-mode branch; while `interview.recommendation_mode` is `pull` the suffix is withheld from every option (§ Recommendation Placement Principles → Recommendation mode)
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

The `(Recommended)` / `(권장)` label is grounded in the statistically-majority rational default the
user has actually been observed to select — never a policy default the system wants to push. Five
principles bind its placement; reasoning and evidence base:
`askuser-protocol-reference.md` § Recommendation Placement Principles.

1. **Emission timing.** Ask when the decision is genuinely uncertain (p ≈ 0.5). When the outcome is
   nearly certain, auto-resolve to the majority option and omit the question.
2. **Question ordering.** Within one call, order questions by descending information gain.
3. **Recommended option = observed majority.** On cold start, fall back to the static default AND
   disclose that in the description — an undisclosed cold-start recommendation is an
   unobserved-recommendation claim (`verification-claim-integrity.md` §1.1 surface 3).
4. **Precondition statement.** The recommended option's `description` states the condition under
   which the recommendation holds, so the user can reject it when it does not apply.
5. **Adaptive strength.** High estimated proficiency → weak recommendation (disclose the inferred
   preference, omit the label); low → label plus rationale; unknown → no inferred-preference label
   at all. This is the label-suppressing condition the mode axis below generalizes, from an
   inferred estimate to an explicit operator setting.

### Recommendation mode

The five principles above state the `push` branch — the distributed default, and the behavior
whenever the key below is absent, empty, or unrecognized. `pull` is the judgment-first branch,
generalizing principle 5 rather than adding a parallel mechanism.

```yaml
interview:
  recommendation_mode: push   # push (default) | pull (judgment-first)
```

[ZONE:Evolvable] [HARD] While `recommendation_mode` is `pull`, the orchestrator MUST omit the
`(권장)` / `(Recommended)` suffix from **every** option label on **every** `AskUserQuestion` call,
and MUST NOT re-encode the same preference through option ordering, description wording, or
`preview` content. The obligation is deliberately unscoped by question class: the tool payload
carries no question-type field, so an obligation scoped to a class the runtime cannot distinguish
could not be measured.

The mode resolves at output-composition time, per surface — never latched at session start, so a
change takes effect on the next composed output and never rewrites or re-renders a round already
emitted.

`pull` withholds a recommendation and nothing else. The observation, the evidence, the enumerated
options, and every gate and evidence obligation elsewhere in this file stay binding in both modes —
including the mandatory, score-independent Implementation Kickoff Approval gate, which under `pull`
asks the same question with an unlabeled first option.

### On-request emission

[ZONE:Evolvable] [HARD] When the user explicitly asks for a recommendation, a preference, or an
analysis, the orchestrator MUST emit the withheld recommendation on the requested surface, in the
same form it would carry under `push` — the `(권장)` / `(Recommended)` label included where the
request concerns an `AskUserQuestion` round. `pull` defers a recommendation until it is asked for;
it does not abolish it.

### The three adopted conditions

Binding on every `pull`-mode composition:

1. **Detect → Explain → Ask, but never decide.** Withholding a recommendation never withholds the
   observation, the evidence, or the enumerated options.
2. **An LLM 'best practice' is not a policy.** A model-inferred default MUST NOT be presented as an
   established project rule.
3. **When uncertain, escalate. Never downgrade.** Where it is unclear which option holds, the
   decision goes to the user — never to a silently-applied default.

## Preview Field Standards

The `preview` field renders a monospace block beside the option list; it **complements**
`description` and never replaces it. [HARD] Single-select only — it is silently dropped when
`multiSelect: true`. Keep it under roughly 12 lines: the pane does not scroll. The recommendation
signal stays on the `(Recommended)` / `(권장)` label; preview content stays neutral and factual.
When to use it, when to skip it, and a worked example: `askuser-protocol-reference.md`
§ Preview Field Standards.

## Report-Before-Ask Gate

[ZONE:Evolvable] [HARD] A decision-type `AskUserQuestion` whose options derive from investigation results MUST be preceded — in the same turn's response body — by a substantive findings report. Investigation results include: `Agent()` fan-out returns (multi-lens analysis, audits, scans), verification batches, and any multi-source evidence gathering the orchestrator performed before composing the question. Asking the user to choose among options they were never given the evidence to evaluate is a gate violation, even when the AskUserQuestion call itself is structurally compliant (labels, descriptions, previews, `(권장)` placement).

### Requested-Deliverable Primacy (user requirement analysis first)

[ZONE:Evolvable] [HARD] When the user's latest message explicitly requests a report, analysis, or explanation ("report on X", "explain why", "analyze this first"), that requested deliverable IS the turn's terminal output: the orchestrator MUST complete the report as a standalone response and end the turn WITHOUT appending a decision-type `AskUserQuestion` to the same turn. Pipeline-stage needs (clarification resolution, scope selection, audit-gate unblocking, next-step routing) NEVER override or preempt the user's stated information request.

- **Requirement analysis first**: re-read the latest message; if it asks for information, deliver it and stop — ask only when a decision is asked for or clearly required
- **No question-as-epilogue**: a next-step question appended to a requested report demotes the report to a preamble
- **Deferred pipeline questions**: they surface in a LATER turn, after the user reacts or says to proceed

### Report Completeness Criteria (all mandatory)

1. **Per-source coverage**: name each investigation source (agent, lens, audit dimension) and state its key findings with quantification (N findings, severity breakdown). A single-line completion claim, in any locale, is NOT a report.
2. **Option-to-report traceability**: every codename, identifier, or finding referenced in the question's option labels / descriptions / previews MUST have been introduced and explained in the preceding report body — the user cannot evaluate what was never explained.
3. **Structured rendering**: the Discovery banner (`.claude/output-styles/moai/moai.md` §8 Discovery Report) or equivalent structured markdown with per-source subsections, scaled to the investigation.

### Preview-as-Report Substitution (named anti-pattern)

[HARD] Option `preview` / `description` fields MUST NOT be the sole carrier of investigation findings. The preview compresses a comparison; the report explains the evidence. Compressing all findings into an option preview table while the response body carries only a one-line completion claim is the named anti-pattern **preview-as-report substitution**.

### Report-Promise Fulfillment

[HARD] When prior narration in the same task promised a consolidated report ("I will consolidate and report", or its equivalent in any locale), the report MUST be rendered before any subsequent decision AskUserQuestion. Claiming the report was delivered when none was rendered is an unobserved completion claim (`verification-claim-integrity.md` §1.1 surface 1).

### Exceptions (gate does not apply)

1. Pure clarify rounds during Context-First Discovery — questions asked BEFORE any investigation exists
2. Confirmation gates on already-reported context (e.g., Implementation Kickoff Approval after plan artifacts were presented in prose)
3. Blocker re-delegation rounds where the subagent's blocker report was already surfaced
4. Preference questions with no investigative basis (naming, formatting choices)

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

Owned by `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format and `.claude/rules/moai/core/agent-common-protocol-reference.md` § Re-delegation Procedure — see there.

---

## Ambiguity Triggers and Exceptions

This section is the **single source of truth** for Stage 1 Clarify trigger conditions. Both `CLAUDE.md §7 Rule 5` and `CLAUDE.md §8 Ambiguity Triggers` cross-reference this definition.

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

1. **Route the decision through `AskUserQuestion`** — preload, then ask, so the user selects instead of typing. The recommended option carries the `(Recommended)` / `(권장)` label.
2. **Close with NO question** — what was done, the evidence, the current state. Where no decision is required, do NOT manufacture one; an unneeded prompt is noise.

"Ask through `AskUserQuestion`, or do not ask" — there is no third "ask in prose" option. The rationalization that a short trailing next-step question on a finished report can be plain prose is the exact failure mode this clause forbids.

**Pre-emit self-check (completion report)** — before sending any "done" report:
- [ ] Does the report end with a `?`-bearing prose next-step prompt? If yes → convert to `AskUserQuestion`, or drop the prompt entirely.
- [ ] If a next-step decision is genuinely needed, is it routed through `AskUserQuestion` (not prose, not a markdown option list)?
- [ ] If no decision is needed, does the report close cleanly with no manufactured question?

## Non-ASCII Tool-Call Encoding

The `AskUserQuestion` payload — `question`, `header`, and every option `label` / `description` / `preview` — routinely carries text in the user's `conversation_language`. For Korean, Japanese, Chinese, and other multi-byte scripts, this text MUST be written as **native UTF-8 directly** in the tool-call JSON. Hand-authored `\uXXXX` escape sequences are **PROHIBITED**.

**Failure Mode**: a malformed escape (stray space, truncated code point, half-written `\u`) corrupts the JSON so the `questions` array parses as a bare string — the call is rejected with `Invalid tool parameters` / `InputValidationError`, and the clarification round silently fails on its first attempt. (Root-cause mechanism, the self-reinforcing pollution loop, and the scope note: `askuser-protocol-reference.md` § Non-ASCII Tool-Call Encoding detail.)
