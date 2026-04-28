---
description: Canonical reference for AskUserQuestion-only interaction protocol, ToolSearch deferred-tool preload procedure, and Socratic interview standards
globs:
---

# AskUserQuestion Protocol — Canonical Reference

> Source: SPEC-ASKUSER-ENFORCE-001 v1.0.0 (2026-04-25)
> This file is the **single source of truth** for AskUserQuestion interaction rules.
> Cross-referenced by: CLAUDE.md §8, moai-constitution.md §MoAI Orchestrator, agent-common-protocol.md §User Interaction Boundary, output-styles/moai/moai.md §3/§10.

---

## Channel Monopoly

**AskUserQuestion is the only user-facing question channel.** The MoAI orchestrator MUST route every user-facing question through an `AskUserQuestion` tool invocation. Free-form interrogative prose in the response body is **prohibited** as a question channel.

Applies to all orchestrator turns involving:
- Clarification questions when user intent is ambiguous (Stage 1 Clarify)
- Preference and decision questions ("Which approach?", "Continue or abort?")
- Socratic interview rounds during Context-First Discovery (CLAUDE.md §7 Rule 5)
- Branch and workflow selection
- Conflict resolution (merge strategy, rollback confirmation, etc.)

**Exceptions** (free-form prose questions permitted ONLY when):
- `AskUserQuestion` is technically unavailable — this should not occur in normal orchestrator operation
- The expression is a statement of status that happens to end with a question mark, not a genuine request for a decision

**Anti-pattern (NEVER repeat)**:
```
# Wrong — free-form prose question in response body
다음 진행 방향을 알려주세요:
- A: 지금 즉시 시작
- B: PR ready + 세션 종료
```

**Correct pattern**: Always use `AskUserQuestion`. See §Free-form Circumvention Prohibition for the "Other" option mechanism.

---

## ToolSearch Preload Procedure

`AskUserQuestion` is a **deferred tool** in Claude Code. Its JSON schema is NOT loaded into the active context at agent initialization time. Attempting to invoke `AskUserQuestion` without first selecting it results in `InputValidationError: tool not in schema`.

### Mandatory Preload Step

Immediately before **every** `AskUserQuestion` call, the orchestrator MUST invoke:

```
ToolSearch(query: "select:AskUserQuestion")
```

This loads the tool schema into the active context and makes the subsequent `AskUserQuestion` call valid.

### General Rule for Deferred Tools

Any deferred tool requires a `ToolSearch` select preload before invocation. The pattern generalizes:

```
ToolSearch(query: "select:<tool>[,<tool>...]")
```

- Single tool: `ToolSearch(query: "select:AskUserQuestion")`
- Multiple tools: `ToolSearch(query: "select:AskUserQuestion,TaskCreate")`

### Preload Sequence

```
[Turn N]
Step 1: ToolSearch(query: "select:AskUserQuestion")   ← preload deferred schema
Step 2: AskUserQuestion({ questions: [...] })           ← now valid to invoke
```

Never reverse or omit Step 1. The preload applies per-turn — if a new turn begins and `AskUserQuestion` will be called again, preload again.

---

## Socratic Interview Structure

When a Stage 1 Clarify trigger is satisfied (see §Ambiguity Triggers and Exceptions), the orchestrator conducts a **Socratic interview** through sequential `AskUserQuestion` rounds.

### Structural Constraints (all mandatory)

1. **Round limit**: Maximum 4 questions per `AskUserQuestion` call (Claude Code hard limit)
2. **Option limit**: Maximum 4 options per question (Claude Code hard limit)
3. **First option label**: MUST carry the `(권장)` (Korean) or `(Recommended)` (English) suffix to signal the recommended choice
4. **Language**: All question text, option labels, and option descriptions MUST be in the user's `conversation_language` (read from `.moai/config/sections/language.yaml`)
5. **Round progression**: Each subsequent round MUST narrow ambiguity by building on previous answers — repeating the same question is prohibited
6. **Termination condition**: Rounds continue until intent clarity reaches 100%; the interview MUST NOT end prematurely
7. **Pre-execution confirmation**: After clarity is achieved, consolidate findings into a brief report and obtain **explicit final confirmation** via `AskUserQuestion` before irreversible actions

### Round Structure Example

```
Round 1: ToolSearch → AskUserQuestion (scope questions)
Round 2: ToolSearch → AskUserQuestion (approach questions, built on Round 1 answers)
...
Final:   ToolSearch → AskUserQuestion (confirmation: "Proceed with this plan?")
```

---

## Option Description Standards

Every option in an `AskUserQuestion` call MUST have a `description` field populated with sufficient detail for the user to evaluate implications and trade-offs **without consulting external context**.

### Required description content

Each option description MUST include:

1. **Immediate result**: What happens immediately if this option is selected
2. **Side effects and risks**: Any follow-on consequences, risks, or irreversibility
3. **Quantitative information** (where applicable): Token cost, latency, file count, etc. (e.g., "saves ~30K tokens", "modifies 5 files")

### Bias Prevention

**bias prevention rule**: Option descriptions MUST use neutral, factual language — no persuasive or deprecating tone
- The recommendation signal is conveyed **exclusively** through the `(권장)` / `(Recommended)` label suffix on the first option
- Descriptions must not phrase the recommended option more favorably or the non-recommended options more negatively than the facts justify

**Anti-pattern**: Writing a description that says "This is the best approach because..." — that is bias. State facts only.

---

## Orchestrator–Subagent Boundary

The `AskUserQuestion` interaction channel is **asymmetric** by design.

### Orchestrator Obligations

The MoAI orchestrator (main session) MUST:
- Use `AskUserQuestion` as the exclusive channel for all user-facing questions
- Preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")` before each call
- Collect all necessary user preferences **before** delegating to subagents
- On receiving a blocker report from a subagent: run an `AskUserQuestion` round with the user, inject the user's responses into a fresh subagent prompt, and re-delegate

See `.claude/rules/moai/core/askuser-protocol.md` (this file) for the complete preload sequence.

### Subagent Prohibitions

Subagents invoked via `Agent()` operate in isolated, stateless contexts and CANNOT interact with users directly:
- [HARD] Subagents MUST NOT invoke `AskUserQuestion`
- [HARD] Subagents MUST NOT output free-form prose questions directed at the user
- [HARD] Subagents MUST NOT embed AskUserQuestion call syntax in their response body

### Blocker Report Format

When a subagent requires user input that was not provided in the spawn prompt, it MUST return a structured blocker report instead of attempting to interact with the user:

```markdown
## Missing Inputs

The following parameters are required but were not provided:

| Parameter | Type | Expected Values | Rationale |
|-----------|------|-----------------|-----------|
| [name]    | [type] | [values]      | [why needed] |

**Blocker**: Cannot proceed without the above inputs. Please re-delegate with these values injected into the prompt.
```

### Re-delegation Procedure

```
[Orchestrator receives blocker report]
Step 1: ToolSearch(query: "select:AskUserQuestion")
Step 2: AskUserQuestion — ask user for the missing inputs
Step 3: Construct fresh subagent prompt with user's answers injected
Step 4: Re-delegate to subagent
```

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

### First-Action Sequence After Trigger

```
Trigger detected
  → Step 1: ToolSearch(query: "select:AskUserQuestion")   [REQ-AUE-002]
  → Step 2: Compose AskUserQuestion round (≤4 Q, ≤4 options, (권장) first, conversation_language)
  → Step 3: Send AskUserQuestion, collect responses
  → Step 4: Assess intent clarity (100% required)
  → Step 5: If <100%: go to Step 1 with narrowed questions
             If 100%: consolidate report → final confirmation → execute
```

---

## Free-form Circumvention Prohibition

Free-form interrogative prose in the response body is **prohibited** as a substitute for `AskUserQuestion`. Using free-form prose questions is prohibited and circumvents the required AskUserQuestion channel. This prohibition is absolute and cannot be overridden by the appearance of a "simple" or "quick" question.

### Why this matters

`AskUserQuestion` automatically appends an **"Other"** option to every question set. This means:
- Users who prefer free-form answers can select "Other" and type their response
- The orchestrator does NOT need to produce free-form questions to support free-form answers
- Structured options via `AskUserQuestion` are faster and less error-prone than prose for most users

### The "Other" Mechanism

When the orchestrator constructs an `AskUserQuestion` round that does not exhaustively cover all possibilities, the "Other" option is automatically available. This covers:
- Edge cases not anticipated in the option list
- User preferences that do not fit the provided options
- Free-form elaboration on a structured choice

### Prohibited Patterns

```
# Prohibited — free-form question in prose
"Which direction would you like to proceed?"

# Prohibited — markdown list as options in prose
- **A**: Run SPEC immediately
- **B**: Review first
- **C**: Abort

# Prohibited — inline question at end of response paragraph
"I've completed the changes. Should I create a PR now?"
```

### Correct Pattern

```
# Correct — always through AskUserQuestion
ToolSearch(query: "select:AskUserQuestion")
AskUserQuestion({
  questions: [{
    question: "다음 단계를 선택하세요.",
    header: "진행 방향",
    options: [
      { label: "PR 즉시 생성 (권장)", description: "현재 변경사항으로 PR을 생성합니다. CI가 자동 실행됩니다." },
      { label: "검토 후 PR", description: "변경사항을 먼저 검토하고 PR을 생성합니다." },
      { label: "중단", description: "현재 작업을 중단하고 상태를 보존합니다." }
    ]
  }]
})
```

---

Version: 1.0.0
Source: SPEC-ASKUSER-ENFORCE-001 (2026-04-25)
Classification: Canonical Reference — do not duplicate content; cross-reference this file instead.
